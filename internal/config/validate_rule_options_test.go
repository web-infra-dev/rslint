package config

import (
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule"
)

type ruleCatalogBuilder struct {
	rules []rule.Rule
}

func (b *ruleCatalogBuilder) Add(name string, ruleImpl rule.Rule) {
	ruleImpl.Name = name
	b.rules = append(b.rules, ruleImpl)
}

func (b *ruleCatalogBuilder) Catalog() *rule.Catalog {
	return rule.NewCatalog(b.rules...)
}

// newOptionsTestCatalog builds a catalog with a representative mix of
// schema situations: a rule with a real options schema, a no-options rule on
// the shared EmptyArraySchema, a not-yet-migrated rule without a schema, and
// a rule whose schema JSON is broken (an authoring bug).
func newOptionsTestCatalog() *rule.Catalog {
	catalog := &ruleCatalogBuilder{}
	noopRun := func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{}
	}
	catalog.Add("with-schema", rule.Rule{
		Name: "with-schema",
		Schema: rule.NewSchema([]byte(`{
			"type": "array",
			"items": [
				{
					"type": "object",
					"properties": {
						"allow": { "type": "array", "items": { "type": "string" } }
					},
					"additionalProperties": false
				}
			],
			"minItems": 0,
			"maxItems": 1
		}`)),
		Run: noopRun,
	})
	catalog.Add("no-options", rule.Rule{
		Name:   "no-options",
		Schema: rule.EmptyArraySchema,
		Run:    noopRun,
	})
	catalog.Add("unmigrated", rule.Rule{
		Name: "unmigrated",
		Run:  noopRun,
	})
	catalog.Add("broken-schema", rule.Rule{
		Name:   "broken-schema",
		Schema: rule.NewSchema([]byte(`not json`)),
		Run:    noopRun,
	})
	return catalog.Catalog()
}

func newDefaultsOptionsTestCatalog() *rule.Catalog {
	catalog := &ruleCatalogBuilder{}
	noopRun := func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{}
	}
	catalog.Add("with-defaults", rule.Rule{
		Name: "with-defaults",
		Schema: rule.NewSchema([]byte(`{
			"type": "array",
			"items": [
				{
					"type": "object",
					"properties": {
						"allow": { "type": "array", "items": { "type": "string" }, "default": [] },
						"strict": { "type": "boolean", "default": true }
					},
					"additionalProperties": false
				}
			],
			"minItems": 0,
			"maxItems": 1
		}`)),
		Run: noopRun,
	})
	catalog.Add("with-nested-defaults", rule.Rule{
		Name: "with-nested-defaults",
		Schema: rule.NewSchema([]byte(`{
			"type": "array",
			"items": [
				{
					"type": "object",
					"properties": {
						"nested": {
							"type": "object",
							"properties": {
								"enabled": { "type": "boolean", "default": true }
							},
							"additionalProperties": false
						}
					},
					"additionalProperties": false
				}
			],
			"minItems": 0,
			"maxItems": 1
		}`)),
		Run: noopRun,
	})
	catalog.Add("with-tuple-defaults", rule.Rule{
		Name: "with-tuple-defaults",
		Schema: rule.NewSchema([]byte(`{
			"type": "array",
			"items": [
				{
					"type": "object",
					"default": { "values": [] },
					"properties": {
						"values": {
							"type": "array",
							"items": [
								{ "type": "string", "default": "filled" }
							],
							"maxItems": 1
						}
					},
					"additionalProperties": false
				}
			],
			"maxItems": 1
		}`)),
		Run: noopRun,
	})
	return catalog.Catalog()
}

func configWithRules(rules Rules) RslintConfig {
	return RslintConfig{ConfigEntry{Rules: rules}}
}

func TestValidateResolvedRuleOptionsPreservesSingleConfigMode(t *testing.T) {
	inputOptions := map[string]any{"values": []any{"original"}}
	input := configWithRules(Rules{
		"unknown-rule": []any{"error", inputOptions},
	})

	normalizedMap, normalized, errs := ValidateResolvedRuleOptions(
		nil,
		input,
		newOptionsTestCatalog(),
	)
	if normalizedMap != nil {
		t.Fatalf("single-config mode returned a non-nil config map: %#v", normalizedMap)
	}
	if len(errs) != 0 {
		t.Fatalf("unexpected validation errors: %v", errs)
	}

	normalizedOptions := normalized[0].Rules["unknown-rule"].([]any)[1].(map[string]any)
	normalizedOptions["values"].([]any)[0] = "changed"
	if got := inputOptions["values"].([]any)[0]; got != "original" {
		t.Fatalf("normalized config aliases its input: %#v", got)
	}
}

func TestValidateResolvedRuleOptionsPreservesMultiConfigMode(t *testing.T) {
	singleConfig := configWithRules(Rules{"unused": "error"})
	normalizedMap, normalizedSingle, errs := ValidateResolvedRuleOptions(
		map[string]RslintConfig{},
		singleConfig,
		newOptionsTestCatalog(),
	)
	if normalizedMap == nil || len(normalizedMap) != 0 {
		t.Fatalf("non-nil empty config map changed mode: %#v", normalizedMap)
	}
	if !reflect.DeepEqual(normalizedSingle, singleConfig) {
		t.Fatalf("inactive single config changed in multi-config mode: %#v", normalizedSingle)
	}
	if len(errs) != 0 {
		t.Fatalf("unexpected validation errors: %v", errs)
	}

	inputOptions := map[string]any{"values": []any{"original"}}
	inputMap := map[string]RslintConfig{
		"/workspace/a": configWithRules(Rules{
			"unknown-rule": []any{"error", inputOptions},
		}),
	}
	normalizedMap, _, errs = ValidateResolvedRuleOptions(
		inputMap,
		nil,
		newOptionsTestCatalog(),
	)
	if len(errs) != 0 {
		t.Fatalf("unexpected validation errors: %v", errs)
	}
	normalizedOptions := normalizedMap["/workspace/a"][0].Rules["unknown-rule"].([]any)[1].(map[string]any)
	normalizedOptions["values"].([]any)[0] = "changed"
	if got := inputOptions["values"].([]any)[0]; got != "original" {
		t.Fatalf("normalized config map aliases its input: %#v", got)
	}
}

func TestValidateResolvedRuleOptionsReportsOwnersInOrder(t *testing.T) {
	invalid := func(value any) RslintConfig {
		return configWithRules(Rules{
			"with-schema": []any{"error", map[string]any{"allow": value}},
		})
	}
	_, _, errs := ValidateResolvedRuleOptions(
		map[string]RslintConfig{
			"/workspace/z": invalid("not-an-array"),
			"/workspace/a": invalid(42),
		},
		nil,
		newOptionsTestCatalog(),
	)
	if len(errs) != 2 {
		t.Fatalf("validation errors = %d, want 2: %v", len(errs), errs)
	}
	if errs[0].ConfigDirectory != "/workspace/a" || errs[1].ConfigDirectory != "/workspace/z" {
		t.Fatalf("owner order = %q, %q", errs[0].ConfigDirectory, errs[1].ConfigDirectory)
	}
	for _, err := range errs {
		if !strings.Contains(err.Error(), "(config at "+err.ConfigDirectory+")") {
			t.Errorf("error does not identify owner: %q", err.Error())
		}
	}
}

func TestValidateRuleOptionsAcceptsValidConfig(t *testing.T) {
	catalog := newOptionsTestCatalog()
	config := configWithRules(Rules{
		"with-schema": []any{"error", map[string]any{"allow": []any{"warn"}}},
		"no-options":  "error",
		"unmigrated":  []any{"warn", map[string]any{"whatever": true}},
	})
	if _, errs := ValidateRuleOptions(config, catalog); len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestValidateRuleOptionsReportsEveryFailure(t *testing.T) {
	catalog := newOptionsTestCatalog()
	config := configWithRules(Rules{
		// `allow` must be an array of strings.
		"with-schema": []any{"error", map[string]any{"allow": "warn"}},
		// A no-options rule given an option.
		"no-options": []any{"error", map[string]any{"unexpected": true}},
	})
	_, errs := ValidateRuleOptions(config, catalog)
	if len(errs) != 2 {
		t.Fatalf("expected 2 errors, got %d: %v", len(errs), errs)
	}
	// Deterministic rule-name order.
	if errs[0].RuleName != "no-options" || errs[1].RuleName != "with-schema" {
		t.Errorf("expected errors sorted by rule name, got %q, %q", errs[0].RuleName, errs[1].RuleName)
	}
	if !strings.Contains(errs[1].Error(), `"with-schema"`) {
		t.Errorf("expected the message to name the rule, got: %v", errs[1])
	}
}

func TestValidateRuleOptionsSkipsDisabledUnknownAndUnmigrated(t *testing.T) {
	catalog := newOptionsTestCatalog()
	config := configWithRules(Rules{
		// Disabled: options would be invalid, but the rule never runs.
		"with-schema": []any{"off", map[string]any{"allow": "warn"}},
		// Unknown rule names are not this step's concern (planned separately).
		"no-such-rule": []any{"error", map[string]any{"allow": "warn"}},
		// No schema declared yet: passes through unvalidated.
		"unmigrated": []any{"error", map[string]any{"whatever": true}},
	})
	if _, errs := ValidateRuleOptions(config, catalog); len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestValidateRuleOptionsDetachesSkippedRuleValues(t *testing.T) {
	catalog := newOptionsTestCatalog()
	sharedOptions := map[string]any{"nested": []any{"original"}}
	config := configWithRules(Rules{
		"with-schema": []any{"off", sharedOptions},
		"no-such-rule": []any{
			"error",
			sharedOptions,
		},
		"unmigrated": []any{"error", sharedOptions},
	})

	normalized, errs := ValidateRuleOptions(config, catalog)
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got: %v", errs)
	}

	disabled := normalized[0].Rules["with-schema"].([]any)[1].(map[string]any)
	unknown := normalized[0].Rules["no-such-rule"].([]any)[1].(map[string]any)
	unmigrated := normalized[0].Rules["unmigrated"].([]any)[1].(map[string]any)
	disabled["nested"].([]any)[0] = "changed"
	normalized[0].Rules["new-rule"] = "error"

	if got := sharedOptions["nested"].([]any)[0]; got != "original" {
		t.Fatalf("normalized skipped value still aliases input: got %#v", got)
	}
	if got := unknown["nested"].([]any)[0]; got != "original" {
		t.Fatalf("normalized skipped values still alias each other: got %#v", got)
	}
	if got := unmigrated["nested"].([]any)[0]; got != "original" {
		t.Fatalf("normalized skipped values still alias each other: got %#v", got)
	}
	if _, exists := config[0].Rules["new-rule"]; exists {
		t.Fatal("normalized Rules map still aliases input")
	}
}

func TestValidateRuleOptionsSurfacesSchemaCompileError(t *testing.T) {
	catalog := newOptionsTestCatalog()
	config := configWithRules(Rules{"broken-schema": "error"})
	_, errs := ValidateRuleOptions(config, catalog)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if errs[0].RuleName != "broken-schema" {
		t.Errorf("expected the compile failure to name the rule, got %q", errs[0].RuleName)
	}
}

func TestValidateRuleOptionsReportsSharedSchemaCompileErrorPerRule(t *testing.T) {
	catalog := &ruleCatalogBuilder{}
	sharedBrokenSchema := rule.NewSchema([]byte(`not json`))
	noopRun := func(rule.RuleContext, []any) rule.RuleListeners {
		return rule.RuleListeners{}
	}
	for _, ruleName := range []string{"alpha", "beta"} {
		catalog.Add(ruleName, rule.Rule{
			Name:   ruleName,
			Schema: sharedBrokenSchema,
			Run:    noopRun,
		})
	}

	_, errs := ValidateRuleOptions(configWithRules(Rules{
		"beta":  "error",
		"alpha": "error",
	}), catalog.Catalog())
	if len(errs) != 2 {
		t.Fatalf("expected one compile error per rule, got %d: %v", len(errs), errs)
	}
	if errs[0].RuleName != "alpha" || errs[1].RuleName != "beta" {
		t.Fatalf("compile errors are not in rule-name order: %v", errs)
	}
}

func TestValidateRuleOptionsReportsSharedEmptyOptionsErrorPerRule(t *testing.T) {
	catalog := &ruleCatalogBuilder{}
	sharedSchema := rule.NewSchema([]byte(`{
		"type": "array",
		"items": [{ "type": "string" }],
		"minItems": 1,
		"maxItems": 1
	}`))
	noopRun := func(rule.RuleContext, []any) rule.RuleListeners {
		return rule.RuleListeners{}
	}
	for _, ruleName := range []string{"alpha", "beta"} {
		catalog.Add(ruleName, rule.Rule{
			Name:   ruleName,
			Schema: sharedSchema,
			Run:    noopRun,
		})
	}

	_, errs := ValidateRuleOptions(RslintConfig{
		{Rules: Rules{"alpha": "error", "beta": []any{"error"}}},
		{Rules: Rules{"alpha": "error", "beta": []any{"error"}}},
	}, catalog.Catalog())
	if len(errs) != 2 {
		t.Fatalf("expected one empty-options error per rule, got %d: %v", len(errs), errs)
	}
	if errs[0].RuleName != "alpha" || errs[1].RuleName != "beta" {
		t.Fatalf("empty-options errors are not in rule-name order: %v", errs)
	}
}

func TestValidateRuleOptionsValidatesEachEntryAndDeduplicates(t *testing.T) {
	catalog := newOptionsTestCatalog()
	badOptions := []any{"error", map[string]any{"allow": "warn"}}
	config := RslintConfig{
		// The identical bad entry twice: reported once.
		ConfigEntry{Rules: Rules{"with-schema": badOptions}},
		ConfigEntry{Rules: Rules{"with-schema": badOptions}},
		// A different bad entry for the same rule: reported separately.
		ConfigEntry{Rules: Rules{"with-schema": []any{"error", map[string]any{"allow": 42}}}},
		// A valid entry for the same rule doesn't mask the bad ones.
		ConfigEntry{Rules: Rules{"with-schema": []any{"error", map[string]any{"allow": []any{"warn"}}}}},
	}
	_, errs := ValidateRuleOptions(config, catalog)
	if len(errs) != 2 {
		t.Fatalf("expected 2 errors (duplicate collapsed, distinct kept), got %d: %v", len(errs), errs)
	}
	for _, err := range errs {
		if err.RuleName != "with-schema" {
			t.Errorf("expected all errors to name with-schema, got %q", err.RuleName)
		}
	}
}

func TestValidateRuleOptionsFillsSchemaDefaultsIntoConfig(t *testing.T) {
	catalog := newDefaultsOptionsTestCatalog()

	options := map[string]any{"strict": false}
	config := configWithRules(Rules{"with-defaults": []any{"error", options}})
	normalized, errs := ValidateRuleOptions(config, catalog)
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got: %v", errs)
	}

	want := map[string]any{"allow": []any{}, "strict": false}
	normalizedRule, _, err := parseRuleConfigValue(normalized[0].Rules["with-defaults"])
	if err != nil {
		t.Fatalf("parse normalized rule: %v", err)
	}
	got := normalizedRule.Options[0].(map[string]any)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected defaults filled into normalized options, got %#v, want %#v", got, want)
	}
	if !reflect.DeepEqual(options, map[string]any{"strict": false}) {
		t.Errorf("expected input options to remain unchanged, got %#v", options)
	}
}

func TestValidateRuleOptionsPreservesTopLevelShapeAndNestedTupleDefaults(t *testing.T) {
	catalog := newDefaultsOptionsTestCatalog()
	nestedValues := []any{}
	config := RslintConfig{
		{Rules: Rules{"with-tuple-defaults": "error"}},
		{Rules: Rules{"with-tuple-defaults": []any{"error"}}},
		{Rules: Rules{
			"with-tuple-defaults": []any{
				"error",
				map[string]any{"values": nestedValues},
			},
		}},
	}

	normalized, errs := ValidateRuleOptions(config, catalog)
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got: %v", errs)
	}
	if normalized[0].Rules["with-tuple-defaults"] != "error" {
		t.Fatalf("severity-only scalar changed shape: %#v", normalized[0].Rules["with-tuple-defaults"])
	}
	severityOnly := normalized[1].Rules["with-tuple-defaults"].([]any)
	if !reflect.DeepEqual(severityOnly, []any{"error"}) {
		t.Fatalf("severity-only array changed shape: %#v", severityOnly)
	}

	ruleConfig, _, err := parseRuleConfigValue(normalized[2].Rules["with-tuple-defaults"])
	if err != nil {
		t.Fatalf("parse normalized rule: %v", err)
	}
	options := ruleConfig.Options[0].(map[string]any)
	if got := options["values"]; !reflect.DeepEqual(got, []any{"filled"}) {
		t.Fatalf("nested tuple default was not published: %#v", got)
	}
	if len(nestedValues) != 0 {
		t.Fatalf("nested tuple default mutated input: %#v", nestedValues)
	}
}

func TestValidateRuleOptionsIsolatesNestedAliases(t *testing.T) {
	catalog := newDefaultsOptionsTestCatalog()
	sharedNested := map[string]any{}
	config := RslintConfig{
		{Rules: Rules{
			"with-nested-defaults": []any{"error", map[string]any{"nested": sharedNested}},
		}},
		{Rules: Rules{
			"with-nested-defaults": []any{"error", map[string]any{"nested": sharedNested}},
		}},
	}

	normalized, errs := ValidateRuleOptions(config, catalog)
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got: %v", errs)
	}
	if len(sharedNested) != 0 {
		t.Fatalf("expected shared input to remain unchanged, got %#v", sharedNested)
	}

	nested := make([]map[string]any, len(normalized))
	for index, entry := range normalized {
		ruleConfig, _, err := parseRuleConfigValue(entry.Rules["with-nested-defaults"])
		if err != nil {
			t.Fatalf("parse normalized rule at entry %d: %v", index, err)
		}
		options := ruleConfig.Options[0].(map[string]any)
		nested[index] = options["nested"].(map[string]any)
		if got := nested[index]["enabled"]; got != true {
			t.Fatalf("entry %d: enabled = %#v, want true", index, got)
		}
	}
	nested[0]["enabled"] = false
	if got := nested[1]["enabled"]; got != true {
		t.Fatalf("normalized entries still share nested options: second enabled = %#v", got)
	}
}

func TestValidateRuleOptionsDoesNotPublishPartialDefaultsOnError(t *testing.T) {
	catalog := newDefaultsOptionsTestCatalog()
	options := map[string]any{"strict": "invalid"}
	config := configWithRules(Rules{
		"with-defaults": []any{"error", options},
	})

	normalized, errs := ValidateRuleOptions(config, catalog)
	if len(errs) != 1 {
		t.Fatalf("expected one validation error, got %d: %v", len(errs), errs)
	}
	ruleConfig, _, err := parseRuleConfigValue(normalized[0].Rules["with-defaults"])
	if err != nil {
		t.Fatalf("parse normalized rule: %v", err)
	}
	got := ruleConfig.Options[0].(map[string]any)
	if !reflect.DeepEqual(got, options) {
		t.Fatalf("failed options contain partial defaults: got %#v, want %#v", got, options)
	}
	got["strict"] = "changed"
	if options["strict"] != "invalid" {
		t.Fatalf("failed normalized options still alias input: input = %#v", options)
	}
}

func TestValidateRuleOptionsConcurrentCallsShareInputSafely(t *testing.T) {
	catalog := newDefaultsOptionsTestCatalog()
	shared := map[string]any{"strict": false}
	ruleValue := []any{"error", shared}
	config := RslintConfig{
		{Rules: Rules{"with-defaults": ruleValue}},
		{Rules: Rules{"with-defaults": ruleValue}},
	}

	var waitGroup sync.WaitGroup
	for call := range 16 {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			normalized, errs := ValidateRuleOptions(config, catalog)
			if len(errs) != 0 {
				t.Errorf("call %d: expected no errors, got: %v", call, errs)
				return
			}
			for entryIndex, entry := range normalized {
				ruleConfig, _, err := parseRuleConfigValue(entry.Rules["with-defaults"])
				if err != nil {
					t.Errorf("call %d entry %d: parse normalized rule: %v", call, entryIndex, err)
					continue
				}
				options := ruleConfig.Options[0].(map[string]any)
				if _, ok := options["allow"]; !ok {
					t.Errorf("call %d entry %d: default was not applied: %#v", call, entryIndex, options)
				}
			}
		}()
	}
	waitGroup.Wait()

	if !reflect.DeepEqual(shared, map[string]any{"strict": false}) {
		t.Fatalf("concurrent validation mutated shared input: %#v", shared)
	}
}

func TestValidateRuleOptionsAcceptsSeverityOnlyAndEmptyOptions(t *testing.T) {
	catalog := newOptionsTestCatalog()
	config := configWithRules(Rules{
		// Bare severity string: options are nil → validated as [].
		"with-schema": "error",
		// Array form with severity only.
		"no-options": []any{"warn"},
	})
	if _, errs := ValidateRuleOptions(config, catalog); len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

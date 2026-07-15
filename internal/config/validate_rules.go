package config

import (
	"fmt"
	"sort"
	"sync"

	"github.com/web-infra-dev/rslint/internal/rule"
)

// RuleValidationError describes one configured rule that failed config-time
// validation: either the rule name resolves to no known rule implementation
// (a typo, an unregistered plugin, or a rule missing from a mounted ESLint
// plugin), or the rule's options failed validation against its declared
// schema (or the schema itself failed to compile — an authoring bug surfaced
// the same way).
type RuleValidationError struct {
	// RuleName is the configured rule, e.g. "no-console".
	RuleName string
	// Message is the complete human-readable failure.
	Message string
}

func (e RuleValidationError) Error() string { return e.Message }

// ValidateRules validates every enabled rule referenced in config — both
// that the rule name resolves to a known rule and that the rule's options
// satisfy its declared schema — and returns every failure (not just the
// first) sorted by rule name for deterministic output.
//
// It is meant to run as a separate step right after configuration is
// resolved and before linting starts, so a bad config fails fast instead of
// being silently ignored (unknown names) or surfacing mid-lint (bad
// options). Rules configured as "off" are exempt from both checks, matching
// ESLint, which never validates a disabled rule. The CLI and API treat every
// returned failure as fatal; the LSP downgrades them to warnings and
// disables the offending rules instead of failing the config transaction.
//
// Name validation resolves against the registry's native rules (core rules
// and first-class plugins) plus the rules of the ESLint plugins mounted via
// eslintPlugins. Mounted rules are resolved against eslintPlugins rather
// than the registry's placeholder registrations: the placeholders may not be
// registered yet when validation runs (the CLI and API validate before
// RegisterEslintPluginRules), and in a long-lived process they may be stale
// leftovers from an earlier config. Declaring a prefix under `plugins`
// grants no exemption: a caller that enables a plugin's rules must also
// supply that plugin's rule names.
//
// Options validation is skipped for rules that declare no schema yet
// (rule.Rule.Schema == nil — the pre-framework status quo) and for disabled
// ("off") entries. ESLint-plugin rules mounted via the config's object-form
// `plugins` never carry a Go schema; the Node worker's own ESLint validates
// their options.
//
// Each entry's options are validated independently, mirroring ESLint, which
// validates every config array element's options rather than only the final
// merged value. The parallel loop leans on [rule.Schema]'s internal
// sync.Once: racing first uses compile each schema at most once, and a
// schema shared by many rules (EmptyArraySchema) compiles a single time.
//
// Validation is also where schema-declared `default` values are filled in,
// exactly like ajv's `useDefaults` in ESLint: [rule.Schema.Validate] mutates
// the options' own maps and slices in place, and because each work item's
// options slice aliases the raw config entry value it was parsed from
// (parseRuleConfigValue sub-slices, it never copies), the defaults land in
// the very options the per-file config merge later hands to rules — no
// write-back needed. The parallel loop stays race-free because every entry's
// rule value is its own decoded JSON value, never shared with another
// entry's.
func ValidateRules(config RslintConfig, registry *RuleRegistry, eslintPlugins []EslintPluginEntry) []RuleValidationError {
	mountedRules := make(map[string]struct{})
	mountedPrefixes := make(map[string]struct{}, len(eslintPlugins))
	for _, plugin := range eslintPlugins {
		if plugin.Prefix == "" {
			continue
		}
		mountedPrefixes[plugin.Prefix] = struct{}{}
		for _, ruleName := range plugin.RuleNames {
			mountedRules[plugin.Prefix+"/"+ruleName] = struct{}{}
		}
	}

	type workItem struct {
		entryIndex int
		ruleName   string
		schema     *rule.Schema
		options    []any
	}

	var errs []RuleValidationError
	unknownSeen := make(map[string]struct{})
	var items []workItem
	for entryIndex, entry := range config {
		for ruleName, ruleValue := range entry.Rules {
			ruleConfig, _, err := parseRuleConfigValue(ruleValue)
			// Malformed rule values (bad severity etc.) are rejected at config
			// ingress by ValidateConfig; they are not this step's concern.
			if err != nil || !ruleConfig.IsEnabled() {
				continue
			}
			ruleImpl, known := registry.GetRule(ruleName)
			if known && ruleImpl.IsEslintPluginRule {
				// A registry placeholder doesn't make a mounted rule known:
				// whether it resolves is decided by this config's own
				// eslintPlugins entries.
				known = false
			}
			if _, mounted := mountedRules[ruleName]; !known && !mounted {
				if _, dup := unknownSeen[ruleName]; dup {
					continue
				}
				unknownSeen[ruleName] = struct{}{}
				errs = append(errs, RuleValidationError{
					RuleName: ruleName,
					Message:  unknownRuleMessage(ruleName, mountedPrefixes),
				})
				continue
			}
			if !known || ruleImpl.Schema == nil {
				continue
			}
			items = append(items, workItem{
				entryIndex: entryIndex,
				ruleName:   ruleName,
				schema:     ruleImpl.Schema,
				options:    ruleConfig.Options,
			})
		}
	}

	// entry.Rules is a map, so collection order is random per process; fix an
	// order up front so the returned errors are deterministic.
	sort.Slice(items, func(i, j int) bool {
		if items[i].ruleName != items[j].ruleName {
			return items[i].ruleName < items[j].ruleName
		}
		return items[i].entryIndex < items[j].entryIndex
	})

	results := make([]error, len(items))
	var wg sync.WaitGroup
	for i, item := range items {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i] = item.schema.Validate(item.options)
		}()
	}
	wg.Wait()

	seen := map[string]bool{}
	for i, err := range results {
		if err == nil {
			continue
		}
		// The same rule configured identically in several entries (common in
		// multi-entry configs) would repeat one message verbatim; report it once.
		key := items[i].ruleName + "\x00" + err.Error()
		if seen[key] {
			continue
		}
		seen[key] = true
		errs = append(errs, RuleValidationError{
			RuleName: items[i].ruleName,
			Message:  fmt.Sprintf("invalid options for rule %q: %v", items[i].ruleName, err),
		})
	}

	// Unknown-name and options failures were collected in separate passes; a
	// final sort interleaves them deterministically.
	sort.Slice(errs, func(i, j int) bool {
		if errs[i].RuleName != errs[j].RuleName {
			return errs[i].RuleName < errs[j].RuleName
		}
		return errs[i].Message < errs[j].Message
	})
	return errs
}

// unknownRuleMessage formats the failure for an enabled rule name that
// resolves to no known rule, distinguishing a core-rule typo, a typo within
// a known or mounted plugin, and a reference to a plugin that is not
// registered at all.
func unknownRuleMessage(ruleName string, mountedPrefixes map[string]struct{}) string {
	prefix := RulePluginPrefix(ruleName)
	if prefix == "" {
		return fmt.Sprintf("unknown rule %q", ruleName)
	}
	if _, mounted := mountedPrefixes[prefix]; mounted || isKnownPluginPrefix(prefix) {
		return fmt.Sprintf("unknown rule %q: plugin %q has no such rule", ruleName, prefix)
	}
	return fmt.Sprintf("unknown rule %q: plugin %q is not registered", ruleName, prefix)
}

// isKnownPluginPrefix reports whether prefix belongs to a first-class plugin
// (one whose rules RegisterAllRules registers natively).
func isKnownPluginPrefix(prefix string) bool {
	for _, plugin := range KnownPlugins {
		if plugin.RulePrefix == prefix {
			return true
		}
	}
	return false
}

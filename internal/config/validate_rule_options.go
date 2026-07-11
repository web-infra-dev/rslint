package config

import (
	"fmt"
	"sort"
	"sync"

	"github.com/web-infra-dev/rslint/internal/rule"
)

// RuleOptionsError describes one configured rule whose options failed
// validation against the rule's declared schema (or whose schema itself
// failed to compile — an authoring bug surfaced the same way).
type RuleOptionsError struct {
	// RuleName is the configured rule, e.g. "no-console".
	RuleName string
	// Err is the schema compile or validation failure.
	Err error
}

func (e RuleOptionsError) Error() string {
	return fmt.Sprintf("invalid options for rule %q: %v", e.RuleName, e.Err)
}

// ValidateRuleOptions validates every enabled rule's options in config
// against the rule's declared schema, in parallel, and returns every failure
// (not just the first) sorted by rule name for deterministic output.
//
// It is meant to run as a separate step right after configuration is
// resolved and before linting starts, so a bad config fails fast instead of
// surfacing mid-lint. Validation is skipped for rules that declare no schema
// yet (rule.Rule.Schema == nil — the pre-framework status quo), for rules
// not present in the registry (unknown names are not an error here — making
// them fatal is planned separately), and for disabled ("off") entries.
// ESLint-plugin rules mounted via the config's object-form `plugins` never
// carry a Go schema; the Node worker's own ESLint validates their options.
//
// Each entry's options are validated independently, mirroring ESLint, which
// validates every config array element's options rather than only the final
// merged value. The parallel loop leans on [rule.Schema]'s internal
// sync.Once: racing first uses compile each schema at most once, and a
// schema shared by many rules (EmptyArraySchema) compiles a single time.
func ValidateRuleOptions(config RslintConfig, registry *RuleRegistry) []RuleOptionsError {
	type workItem struct {
		entryIndex int
		ruleName   string
		schema     *rule.Schema
		options    []any
	}

	var items []workItem
	for entryIndex, entry := range config {
		for ruleName, ruleValue := range entry.Rules {
			ruleConfig, _, err := parseRuleConfigValue(ruleValue)
			// Malformed rule values (bad severity etc.) are rejected at config
			// ingress by ValidateConfig; they are not this step's concern.
			if err != nil || !ruleConfig.IsEnabled() {
				continue
			}
			ruleImpl, exists := registry.GetRule(ruleName)
			if !exists || ruleImpl.Schema == nil {
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

	var errs []RuleOptionsError
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
		errs = append(errs, RuleOptionsError{RuleName: items[i].ruleName, Err: err})
	}
	return errs
}

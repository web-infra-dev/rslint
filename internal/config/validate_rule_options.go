package config

import (
	"fmt"
	"runtime"
	"slices"
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

// ValidateRuleOptions validates every enabled rule's options in config against
// the rule's declared schema and returns a normalized config plus every failure
// (not just the first) sorted by rule name for deterministic output. The input
// config is never mutated. The returned config owns each entry's Rules map and
// its JSON-shaped rule values; successful options contain schema defaults,
// while failed options remain an unmodified copy of their input value.
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
// merged value. Unique schemas are compiled first with bounded parallelism;
// options are then normalized and validated by a second bounded worker pool.
// Separating the phases keeps repeated entries from occupying every worker
// while they wait on the same [rule.Schema] compile.
//
// Validation is also where schema-declared `default` values are filled in,
// exactly like ajv's `useDefaults` in ESLint. [rule.Schema.Validate] mutates
// maps and slices in place, so every options-bearing work item receives a deep
// copy of its complete raw rule value first. This makes arbitrary aliases
// between entries safe without serializing independent schema validation.
func ValidateRuleOptions(config RslintConfig, registry *RuleRegistry) (RslintConfig, []RuleOptionsError) {
	type workItem struct {
		entryIndex  int
		ruleName    string
		schema      *rule.Schema
		schemaIndex int
		ruleValue   any
		hasOptions  bool
	}
	type workResult struct {
		ruleValue any
		err       error
	}

	normalized := slices.Clone(config)
	schemaIndexes := make(map[*rule.Schema]int)
	var schemas []*rule.Schema
	var schemaNeedsEmptyOptionsValidation []bool
	var items []workItem
	for entryIndex, entry := range config {
		if entry.Rules == nil {
			continue
		}
		normalizedRules := make(Rules, len(entry.Rules))
		normalized[entryIndex].Rules = normalizedRules
		for ruleName, ruleValue := range entry.Rules {
			ruleConfig, hasOptions, err := parseRuleConfigValue(ruleValue)
			// Malformed rule values (bad severity etc.) are rejected at config
			// ingress by ValidateConfig; they are not this step's concern.
			if err != nil || !ruleConfig.IsEnabled() {
				normalizedRules[ruleName] = cloneConfigValue(ruleValue)
				continue
			}
			ruleImpl, exists := registry.GetRule(ruleName)
			if !exists || ruleImpl.Schema == nil {
				normalizedRules[ruleName] = cloneConfigValue(ruleValue)
				continue
			}
			schemaIndex, exists := schemaIndexes[ruleImpl.Schema]
			if !exists {
				schemaIndex = len(schemas)
				schemaIndexes[ruleImpl.Schema] = schemaIndex
				schemas = append(schemas, ruleImpl.Schema)
				schemaNeedsEmptyOptionsValidation = append(schemaNeedsEmptyOptionsValidation, false)
			}
			if !hasOptions {
				schemaNeedsEmptyOptionsValidation[schemaIndex] = true
			}
			items = append(items, workItem{
				entryIndex:  entryIndex,
				ruleName:    ruleName,
				schema:      ruleImpl.Schema,
				schemaIndex: schemaIndex,
				ruleValue:   ruleValue,
				hasOptions:  hasOptions,
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

	compileErrors := make([]error, len(schemas))
	emptyOptionsErrors := make([]error, len(schemas))
	parallelRuleOptionsWork(len(schemas), func(index int) {
		_, compileErrors[index] = schemas[index].Compile()
		if compileErrors[index] == nil && schemaNeedsEmptyOptionsValidation[index] {
			emptyOptionsErrors[index] = schemas[index].Validate(nil)
		}
	})

	results := make([]workResult, len(items))
	var optionItemIndexes []int
	for index, item := range items {
		if err := compileErrors[item.schemaIndex]; err != nil {
			results[index].err = err
			continue
		}
		if !item.hasOptions {
			results[index].err = emptyOptionsErrors[item.schemaIndex]
			continue
		}
		optionItemIndexes = append(optionItemIndexes, index)
	}
	parallelRuleOptionsWork(len(optionItemIndexes), func(workIndex int) {
		index := optionItemIndexes[workIndex]
		item := items[index]
		// hasOptions is only true for array-form rule values with at least one
		// element after the severity, so the cloned options start at index 1.
		clonedValue, ok := cloneConfigValue(item.ruleValue).([]any)
		if !ok {
			results[index].err = fmt.Errorf("internal error: cloned rule value has unexpected type %T", item.ruleValue)
			return
		}
		results[index].err = item.schema.Validate(clonedValue[1:])
		if results[index].err == nil {
			results[index].ruleValue = clonedValue
		}
	})

	for index, item := range items {
		ruleValue := results[index].ruleValue
		if results[index].err != nil || !item.hasOptions {
			ruleValue = cloneConfigValue(item.ruleValue)
		}
		normalized[item.entryIndex].Rules[item.ruleName] = ruleValue
	}

	var errs []RuleOptionsError
	seen := map[string]bool{}
	for i, result := range results {
		err := result.err
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
	return normalized, errs
}

func parallelRuleOptionsWork(taskCount int, work func(int)) {
	if taskCount == 0 {
		return
	}
	workerCount := min(runtime.GOMAXPROCS(0), taskCount)
	if workerCount == 1 {
		for index := range taskCount {
			work(index)
		}
		return
	}

	jobs := make(chan int, workerCount)
	var workers sync.WaitGroup
	workers.Add(workerCount)
	for range workerCount {
		go func() {
			defer workers.Done()
			for index := range jobs {
				work(index)
			}
		}()
	}
	for index := range taskCount {
		jobs <- index
	}
	close(jobs)
	workers.Wait()
}

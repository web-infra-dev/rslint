package linter

import (
	"sync"

	"github.com/web-infra-dev/rslint/internal/rule"
)

// preparedRuleSet is the immutable rule selection shared by file plans with
// the same configured rules and checker capability. Execution consumes these
// ordered projections without selecting or copying rules again. File contexts,
// checker instances, language defaults and plugin routing remain per-file.
type preparedRuleSet struct {
	all     []rule.ConfiguredRule
	native  []rule.ConfiguredRule
	plugins []rule.ConfiguredRule

	environment           *rule.RuleEnvironment
	pluginLanguageOptions rule.LanguageOptions
}

// ruleSetBuilder belongs only to plan preparation. Its identity table is
// discarded after workers join; published plans retain only prepared sets.
type ruleSetBuilder struct {
	sets sync.Map // ruleSetKey -> *preparedRuleSet
}

type ruleSetKey struct {
	// A typed pointer keeps the input alive. Exact start and length distinguish
	// immutable slice ranges; equal names do not imply equal configurations.
	first          *rule.ConfiguredRule
	length         int
	hasTypeChecker bool
}

func (builder *ruleSetBuilder) prepare(rules []rule.ConfiguredRule, hasTypeChecker bool) *preparedRuleSet {
	if len(rules) == 0 {
		return nil
	}
	key := ruleSetKey{first: &rules[0], length: len(rules), hasTypeChecker: hasTypeChecker}
	if cached, ok := builder.sets.Load(key); ok {
		return cached.(*preparedRuleSet) //nolint:forcetypeassert // Private table stores only prepared sets, including nil.
	}
	// Preparation is pure. Concurrent misses may compute duplicate candidates,
	// but every reader receives the same complete winning set.
	prepared := prepareRuleSet(rules, hasTypeChecker)
	actual, _ := builder.sets.LoadOrStore(key, prepared)
	return actual.(*preparedRuleSet) //nolint:forcetypeassert // See the private table invariant above.
}

func prepareRuleSet(rules []rule.ConfiguredRule, hasTypeChecker bool) *preparedRuleSet {
	if !hasTypeChecker {
		rules = rule.FilterNonTypeAwareRules(rules)
	}
	if len(rules) == 0 {
		return nil
	}
	prepared := &preparedRuleSet{all: rules}
	nativeCount := 0
	for _, configured := range rules {
		// Preserve the plugin projection's last configured environment, while
		// native execution uses the first native rule's non-nil environment.
		if configured.Environment != nil {
			prepared.pluginLanguageOptions = configured.Environment.LanguageOptions
		}
		if !configured.IsEslintPluginRule {
			nativeCount++
			if prepared.environment == nil {
				prepared.environment = configured.Environment
			}
		}
	}
	switch nativeCount {
	case len(rules):
		prepared.native = rules
	case 0:
		prepared.plugins = rules
	default:
		// Partition once, preserving order within each executor's projection.
		// Uniform sets reuse the eligible input without another allocation.
		partition := make([]rule.ConfiguredRule, len(rules))
		prepared.native = partition[:nativeCount:nativeCount]
		prepared.plugins = partition[nativeCount:]
		nativeIndex, pluginIndex := 0, 0
		for _, configured := range rules {
			if configured.IsEslintPluginRule {
				prepared.plugins[pluginIndex] = configured
				pluginIndex++
			} else {
				prepared.native[nativeIndex] = configured
				nativeIndex++
			}
		}
	}
	return prepared
}

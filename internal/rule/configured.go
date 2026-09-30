package rule

import (
	"sync"

	"github.com/web-infra-dev/rslint/internal/utils"
)

// ConfiguredRule is one enabled rule after configuration resolution. It is a
// rule-framework value: config produces it and linter consumes it.
type ConfiguredRule struct {
	Name             string
	Environment      *RuleEnvironment
	Severity         DiagnosticSeverity
	RequiresTypeInfo bool
	// IsEslintPluginRule marks a rule that executes in the Node plugin-lint
	// worker rather than natively in Go. Run remains a no-op placeholder for
	// those entries.
	IsEslintPluginRule bool
	// Options is the raw user-configured rule options after severity. Native
	// rules capture it in Run; the Node worker consumes it directly.
	Options []any
	Run     func(ctx RuleContext) RuleListeners

	prepared *preparedRule
}

// preparedRule is the identity and idle-runner pool of one resolved rule
// configuration. Copies of ConfiguredRule share it; separately configured rules
// never do, even when their names or option values are equal. The pool belongs
// to the configuration, not the catalog or a process-wide cache. Idle runners
// may be discarded by the runtime; active runners belong to their Executor.
type preparedRule struct {
	runners sync.Pool
}

// Configure binds immutable, normalized options to this rule. Config resolution
// supplies Environment and Severity on the returned descriptor. Preparation is
// lazy, so disabled, filtered, and plugin-only work need not initialize a rule.
func (r Rule) Configure(options []any) ConfiguredRule {
	configured := ConfiguredRule{
		Name:               r.Name,
		RequiresTypeInfo:   r.RequiresTypeInfo,
		IsEslintPluginRule: r.IsEslintPluginRule,
		Options:            options,
		Run: func(ctx RuleContext) RuleListeners {
			return r.Run(ctx, options)
		},
	}
	if r.Prepare != nil {
		configured.prepared = &preparedRule{runners: sync.Pool{
			New: func() any { return r.Prepare(options) },
		}}
	}
	return configured
}

// RuleEnvironment is the immutable file-level configuration shared by every
// configured rule in one resolved config shape.
type RuleEnvironment struct {
	Settings map[string]interface{}
	// LanguageOptions is normalized once and exposed through ctx.LanguageOptions;
	// it also selects the language-global catalog. Its zero value means latest
	// ECMAScript with module source semantics.
	LanguageOptions LanguageOptions
	// Globals contains config-declared language globals. Inline declarations
	// are merged once per source file during execution.
	Globals map[string]utils.GlobalAccess
}

// FilterNonTypeAwareRules returns the entries that do not require a checker.
// The input is immutable shared config state and is never compacted in place.
func FilterNonTypeAwareRules(rules []ConfiguredRule) []ConfiguredRule {
	filtered := make([]ConfiguredRule, 0, len(rules))
	for _, configured := range rules {
		if !configured.RequiresTypeInfo {
			filtered = append(filtered, configured)
		}
	}
	return filtered
}

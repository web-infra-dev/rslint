package rule

import "github.com/web-infra-dev/rslint/internal/utils"

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
}

// RuleEnvironment is the immutable file-level configuration shared by every
// configured rule in one resolved config shape.
type RuleEnvironment struct {
	Settings map[string]interface{}
	// LanguageOptions is normalized once and used to construct file-level
	// globals and scope facts.
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

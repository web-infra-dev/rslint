package max_nested_calls

import (
	_ "embed"

	"github.com/web-infra-dev/rslint/internal/rule"
)

//go:embed schema.json
var schemaJSON []byte

// MaxNestedCallsRule is intentionally inert for the tests-first red checkpoint.
// Port target: eslint-plugin-unicorn v75.0.0.
var MaxNestedCallsRule = rule.Rule{
	Name:   "unicorn/max-nested-calls",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{}
	},
}

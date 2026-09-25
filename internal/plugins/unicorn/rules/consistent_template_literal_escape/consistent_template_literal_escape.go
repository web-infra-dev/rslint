package consistent_template_literal_escape

import "github.com/web-infra-dev/rslint/internal/rule"

// ConsistentTemplateLiteralEscapeRule is intentionally inert for the tests-first red checkpoint.
// Port target: eslint-plugin-unicorn v75.0.0.
var ConsistentTemplateLiteralEscapeRule = rule.Rule{
	Name:   "unicorn/consistent-template-literal-escape",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{}
	},
}

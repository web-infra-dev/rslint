package no_global_object_property_assignment

import "github.com/web-infra-dev/rslint/internal/rule"

// NoGlobalObjectPropertyAssignmentRule is intentionally inert for the tests-first red checkpoint.
// Port target: eslint-plugin-unicorn v75.0.0.
var NoGlobalObjectPropertyAssignmentRule = rule.Rule{
	Name:   "unicorn/no-global-object-property-assignment",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{}
	},
}

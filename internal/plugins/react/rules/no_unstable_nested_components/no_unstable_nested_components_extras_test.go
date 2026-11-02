package no_unstable_nested_components

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Configured patterns preserve minimatch negation and expansion semantics.
func TestNoUnstableNestedComponentsRuleMinimatchPatterns(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnstableNestedComponentsRule, []rule_tester.ValidTestCase{
		{
			Code:    `function ParentComponent() { return <Component renderBody={() => <div />} />; }`,
			Tsx:     true,
			Options: map[string]interface{}{"propNamePattern": "render!(Header|Footer)"},
		},
		{
			Code:    `function ParentComponent() { return <Component render2={() => <div />} />; }`,
			Tsx:     true,
			Options: map[string]interface{}{"propNamePattern": "render{1..3}"},
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    `function ParentComponent() { return <Component renderHeader={() => <div />} />; }`,
			Tsx:     true,
			Options: map[string]interface{}{"propNamePattern": "render!(Header|Footer)"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: errorMessageComponentAsProps}},
		},
	})
}

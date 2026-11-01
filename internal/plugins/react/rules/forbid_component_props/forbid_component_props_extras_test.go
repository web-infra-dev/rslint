package forbid_component_props

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Configured patterns preserve minimatch negation and expansion semantics.
func TestForbidComponentPropsRuleMinimatchPatterns(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ForbidComponentPropsRule, []rule_tester.ValidTestCase{
		{
			Code:    `<Widget data-safe="x" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{map[string]interface{}{"propNamePattern": "data-!(safe|test)"}}},
		},
		{
			Code:    `<WidgetModern className="x" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{map[string]interface{}{"propName": "className", "allowedForPatterns": []interface{}{"Widget!(Legacy)"}}}},
		},
		{
			Code:    `<WidgetLegacy className="x" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{map[string]interface{}{"propName": "className", "disallowedForPatterns": []interface{}{"Widget!(Legacy)"}}}},
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    `<Widget data-secret="x" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{map[string]interface{}{"propNamePattern": "data-!(safe|test)"}}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "propIsForbidden"}},
		},
		{
			Code:    `<WidgetLegacy className="x" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{map[string]interface{}{"propName": "className", "allowedForPatterns": []interface{}{"Widget!(Legacy)"}}}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "propIsForbidden"}},
		},
		{
			Code:    `<WidgetModern className="x" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{map[string]interface{}{"propName": "className", "disallowedForPatterns": []interface{}{"Widget!(Legacy)"}}}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "propIsForbidden"}},
		},
		{
			Code:    `<Widget data-2="x" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{map[string]interface{}{"propNamePattern": "data-{1..3}"}}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "propIsForbidden"}},
		},
	})
}

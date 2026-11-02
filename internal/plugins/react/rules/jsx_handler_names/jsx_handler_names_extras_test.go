package jsx_handler_names

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Configured patterns preserve minimatch negation and expansion semantics.
func TestJsxHandlerNamesRuleMinimatchPatterns(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxHandlerNamesRule, []rule_tester.ValidTestCase{
		{
			Code:    `<WidgetModern onClick={bad} />;`,
			Tsx:     true,
			Options: map[string]interface{}{"checkLocalVariables": true, "ignoreComponentNames": []interface{}{"Widget!(Legacy)"}},
		},
		{
			Code:    `<Widget2 onClick={bad} />;`,
			Tsx:     true,
			Options: map[string]interface{}{"checkLocalVariables": true, "ignoreComponentNames": []interface{}{"Widget{1..3}"}},
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    `<WidgetLegacy onClick={bad} />;`,
			Tsx:     true,
			Options: map[string]interface{}{"checkLocalVariables": true, "ignoreComponentNames": []interface{}{"Widget!(Legacy)"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "badHandlerName"}},
		},
	})
}

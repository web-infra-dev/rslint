package jsx_uses_react

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxUsesReactRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxUsesReactRule, []rule_tester.ValidTestCase{
		{
			Code: `var x = <div />`,
			Tsx:  true,
		},
	}, []rule_tester.InvalidTestCase{})
}

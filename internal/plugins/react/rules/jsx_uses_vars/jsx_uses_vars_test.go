package jsx_uses_vars

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxUsesVarsRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxUsesVarsRule, []rule_tester.ValidTestCase{
		{
			Code: `var x = <div />`,
			Tsx:  true,
		},
	}, []rule_tester.InvalidTestCase{})
}

package explicit_function_return_type

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestVoidParenthesized(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ExplicitFunctionReturnTypeRule, []rule_tester.ValidTestCase{
		// Parenthesized void expression — should still be valid
		{
			Code:    `const log = (message: string) => (void console.log(message));`,
			Options: map[string]interface{}{"allowConciseArrowFunctionExpressionsStartingWithVoid": true},
		},
	}, []rule_tester.InvalidTestCase{})
}

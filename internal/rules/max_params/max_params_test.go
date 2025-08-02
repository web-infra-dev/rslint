package max_params

import (
	"testing"

	"github.com/typescript-eslint/rslint/internal/rule_tester"
	"github.com/typescript-eslint/rslint/internal/rules/fixtures"
)

func TestMaxParamsRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &MaxParamsRule, 
		[]rule_tester.ValidTestCase{
			{Code: "function foo(a: number) {}"},
			{Code: "function foo(a: number, b: string) {}"},
		}, 
		[]rule_tester.InvalidTestCase{
			// TODO: Add invalid test cases with too many parameters
		})
}
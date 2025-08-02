package explicit_module_boundary_types

import (
	"testing"

	"github.com/typescript-eslint/rslint/internal/rule_tester"
	"github.com/typescript-eslint/rslint/internal/rules/fixtures"
)

func TestExplicitModuleBoundaryTypesRule(t *testing.T) {
	// TODO: Add test cases
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ExplicitModuleBoundaryTypesRule, 
		[]rule_tester.ValidTestCase{
			{Code: "function foo(): number { return 1; }"},
		}, 
		[]rule_tester.InvalidTestCase{
			// TODO: Add invalid test cases
		})
}
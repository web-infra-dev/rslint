package explicit_function_return_type

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestParenthesizedExpressionEdgeCases tests that tsgo's preserved ParenthesizedExpression
// nodes don't break the rule. ESLint strips parens from AST; tsgo keeps them.
func TestParenthesizedExpressionEdgeCases(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ExplicitFunctionReturnTypeRule, []rule_tester.ValidTestCase{
		// HOF: arrow with parenthesized function body — should recognize inner as function
		{
			Code:    `() => ((): void => {});`,
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
		},
		// HOF: return parenthesized function — should recognize as HOF
		{
			Code: `
function foo() {
  return ((): void => {});
}
			`,
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
		},
		// allowDirectConstAssertionInArrowFunctions: parenthesized `as const`
		{
			Code:    `const fn = (value: number) => (({ type: 'X', value }) as const);`,
			Options: map[string]interface{}{"allowDirectConstAssertionInArrowFunctions": true},
		},
		// ancestorHasReturnType: property value wrapped in parens
		{
			Code: `
interface Foo { bar: () => string; }
function foo(): Foo {
  return { bar: (() => 'test') };
}
			`,
		},
	}, []rule_tester.InvalidTestCase{
		// Parenthesized non-function body — NOT HOF
		{
			Code:   `() => (1);`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType"}},
		},
	})
}

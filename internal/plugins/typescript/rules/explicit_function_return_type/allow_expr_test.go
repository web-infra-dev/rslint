package explicit_function_return_type

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestAllowExpressionsObjectMethods tests that allowExpressions correctly handles
// object method shorthand and object getters.
// In ESLint's AST: { foo() {} } → Property > FunctionExpression.
// Property is NOT in allowExpressions exclusion list, so it should be valid.
// In tsgo: { foo() {} } → ObjectLiteralExpression > MethodDeclaration.
func TestAllowExpressionsObjectMethods(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ExplicitFunctionReturnTypeRule, []rule_tester.ValidTestCase{
		// Object method shorthand with allowExpressions: true
		{
			Code:    `const obj = { foo() { return 1; } };`,
			Options: map[string]interface{}{"allowExpressions": true},
		},
		// Object getter with allowExpressions: true
		{
			Code:    `const obj = { get foo() { return 1; } };`,
			Options: map[string]interface{}{"allowExpressions": true},
		},
		// Object method with function value — already works (PropertyAssignment parent)
		{
			Code:    `const obj = { foo: function() { return 1; } };`,
			Options: map[string]interface{}{"allowExpressions": true},
		},
	}, []rule_tester.InvalidTestCase{
		// Class method — NOT valid with allowExpressions (MethodDefinition is excluded in ESLint)
		{
			Code: `
class Foo {
  method() { return 1; }
}
			`,
			Options: map[string]interface{}{"allowExpressions": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 3, Column: 3}},
		},
		// Class getter — NOT valid with allowExpressions
		{
			Code: `
class Foo {
  get prop() { return 1; }
}
			`,
			Options: map[string]interface{}{"allowExpressions": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 3, Column: 3}},
		},
	})
}

// TestValidParamsExtras locks in branches and edge shapes that the upstream test suite
// doesn't exercise. The rule intentionally follows eslint-plugin-promise's syntax-only
// isPromise helper instead of using type information.
package valid_params_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/promise/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/valid_params"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestValidParamsExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&valid_params.ValidParamsRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: parenthesized Promise receiver ----
			{Code: `(Promise).resolve(1)`},
			{Code: `((Promise)).reject(error)`},

			// ---- Dimension 4: parenthesized chain receiver ----
			{Code: `(somePromise()).then(success)`},
			{Code: `((promiseReference)).catch(callback)`},

			// ---- Dimension 4: TS non-null / type-expression wrappers are not recognised as Promise static receiver ----
			// Upstream only accepts callee.object.type === "Identifier" for Promise statics.
			{Code: `Promise!.resolve(1, 2)`},
			{Code: `(Promise as any).resolve(1, 2)`},
			{Code: `(Promise satisfies any).resolve(1, 2)`},

			// ---- Dimension 4: element access is not a MemberExpression property.name match ----
			{Code: `Promise["resolve"](1, 2)`},
			{Code: "Promise[`resolve`](1, 2)"},
			{Code: `somePromise()["then"]()`},
			{Code: "somePromise()[`catch`]()"},

			// ---- Dimension 4: optional-call shape with valid arity ----
			{Code: `promiseReference.then?.(success)`},
			{Code: `promiseReference.catch?.(callback)`},

			// ---- Dimension 4: optional chain with element access remains unrecognised ----
			{Code: `promiseReference?.["then"]()`},

			// ---- Dimension 4: empty arguments list on non-checked method ----
			{Code: `Promise.withResolvers()`},
			{Code: `Promise.withResolvers(1, 2)`},
			{Code: `somePromise().done()`},

			// ---- Dimension 4: nested traversal boundaries still visit inner calls independently ----
			{Code: `function outer() { function inner() { return Promise.resolve(1) } return inner }`},
			{Code: `class C { method() { return Promise.all(items) } }`},

			// ---- Dimension 4: TS declaration/body-absent forms do not crash ----
			{Code: `declare function f(): Promise<void>`},
			{Code: `abstract class C { abstract method(): Promise<void> }`},

			// ---- N/A: access/key forms on object/class property declarations ----
			// This rule only inspects CallExpression callees, not declaration keys.
			// ---- N/A: autofix boundaries ----
			// This rule does not provide fixes.

			// ---- Branch lock-in: exclude parses []string in Go tests as well as JSON []interface{} ----
			{Code: `somePromise().catch(TypeError, handler)`, Options: map[string]interface{}{"exclude": []string{"catch"}}},

			// ---- Branch lock-in: unknown methods are ignored even when nested after a promise call ----
			{Code: `Promise.resolve(1).unknown(1, 2, 3)`},

			// ---- Real-user: Bluebird filtered catch can be excluded ----
			{Code: `fetchData().catch(TypeError, recover).catch(logError)`, Options: map[string]interface{}{"exclude": []interface{}{"catch"}}},

			// ---- Real-user: type checker would reject this, upstream valid-params does not ----
			{Code: `Promise.all(123)`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized Promise receiver with invalid arity ----
			{
				Code:   `(Promise).resolve(1, 2)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireOneOptionalArgument", Message: "Promise.resolve() requires 0 or 1 arguments, but received 2", Line: 1, Column: 1}},
			},

			// ---- Dimension 4: parenthesized chain receiver with invalid arity ----
			{
				Code:   `(somePromise()).then()`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireTwoOptionalArguments", Message: "Promise.then() requires 1 or 2 arguments, but received 0", Line: 1, Column: 1}},
			},

			// ---- Dimension 4: optional property access / optional call with invalid arity ----
			{
				Code:   `promiseReference?.then()`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireTwoOptionalArguments", Message: "Promise.then() requires 1 or 2 arguments, but received 0", Line: 1, Column: 1}},
			},
			{
				Code:   `promiseReference.then?.()`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireTwoOptionalArguments", Message: "Promise.then() requires 1 or 2 arguments, but received 0", Line: 1, Column: 1}},
			},

			// ---- Branch lock-in: resolve/reject allow zero args but not two ----
			{
				Code:   `Promise.reject((error), extra)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireOneOptionalArgument", Message: "Promise.reject() requires 0 or 1 arguments, but received 2", Line: 1, Column: 1}},
			},

			// ---- Branch lock-in: then upper bound ----
			{
				Code:   `Promise.resolve(1).then(a, b, c, d)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireTwoOptionalArguments", Message: "Promise.then() requires 1 or 2 arguments, but received 4", Line: 1, Column: 1}},
			},

			// ---- Branch lock-in: one-argument methods lower bound ----
			{
				Code:   `Promise.all()`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireOneArgument", Message: "Promise.all() requires 1 argument, but received 0", Line: 1, Column: 1}},
			},

			// ---- Branch lock-in: exclude only suppresses listed method ----
			{
				Code:    `somePromise().then().catch(TypeError, handler)`,
				Options: map[string]interface{}{"exclude": []interface{}{"catch"}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "requireTwoOptionalArguments", Message: "Promise.then() requires 1 or 2 arguments, but received 0", Line: 1, Column: 1}},
			},

			// ---- Branch lock-in: chained call recursion detects methods after Promise statics ----
			{
				Code:   `Promise.all(items).finally()`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireOneArgument", Message: "Promise.finally() requires 1 argument, but received 0", Line: 1, Column: 1}},
			},

			// ---- Real-user: dangling catch/finally calls are reported even though TS allows optional callbacks ----
			{
				Code:   `fetchData().catch().finally()`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requireOneArgument", Message: "Promise.finally() requires 1 argument, but received 0", Line: 1, Column: 1}, {MessageId: "requireOneArgument", Message: "Promise.catch() requires 1 argument, but received 0", Line: 1, Column: 1}},
			},
		},
	)
}

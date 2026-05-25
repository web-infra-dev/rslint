package no_useless_call

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUselessCall(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUselessCallRule,
		[]rule_tester.ValidTestCase{
			// `this` binding is different.
			{Code: `foo.apply(obj, 1, 2);`},
			{Code: `obj.foo.apply(null, 1, 2);`},
			{Code: `obj.foo.apply(otherObj, 1, 2);`},
			{Code: `a.b(x, y).c.foo.apply(a.b(x, z).c, 1, 2);`},
			{Code: `foo.apply(obj, [1, 2]);`},
			{Code: `obj.foo.apply(null, [1, 2]);`},
			{Code: `obj.foo.apply(otherObj, [1, 2]);`},
			{Code: `a.b(x, y).c.foo.apply(a.b(x, z).c, [1, 2]);`},
			{Code: `a.b.foo.apply(a.b.c, [1, 2]);`},

			// ignores variadic.
			{Code: `foo.apply(null, args);`},
			{Code: `obj.foo.apply(obj, args);`},

			// ignores computed property.
			{Code: `var call; foo[call](null, 1, 2);`},
			{Code: `var apply; foo[apply](null, [1, 2]);`},

			// ignores incomplete things.
			{Code: `foo.call();`},
			{Code: `obj.foo.call();`},
			{Code: `foo.apply();`},
			{Code: `obj.foo.apply();`},

			// Optional chaining: receiver shape differs from thisArg.
			{Code: `obj?.foo.bar.call(obj.foo, 1, 2);`},

			// Private member: tsgo PropertyAccessExpression.Name() may be a
			// PrivateIdentifier, which ESLint's `property.type === "Identifier"`
			// check excludes.
			{Code: `class C { #call: any; wrap(foo: any) { foo.#call(undefined, 1, 2); } }`},

			// ---- tsgo-/TS-specific receiver shapes ----
			// AsExpression in the receiver but not in thisArg → token streams
			// differ (`obj as any` vs `obj`), so the rule must NOT report.
			{Code: `(obj as any).foo.call(obj, 1, 2);`},
			// NonNullAssertion in the receiver but not in thisArg.
			{Code: `obj!.foo.call(obj, 1, 2);`},
			// SatisfiesExpression in the receiver but not in thisArg.
			{Code: `(obj satisfies any).foo.call(obj, 1, 2);`},
			// Generic call: `.call<T>(thisArg, ...)` — type args don't change
			// the call/apply detection but receiver still differs.
			{Code: `obj.foo.call<number>(other, 1, 2);`},
			// Spread thisArg → not null/undefined → ignored.
			{Code: `foo.call(...args);`},
			// Numeric literal as thisArg with no MemberExpression receiver
			// → not null/undefined.
			{Code: `foo.call(0, 1, 2);`},
			// Nested .call.call chain with mismatched receiver.
			{Code: `foo.call.call(other, 1, 2);`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- call ----
			{
				Code: `foo.call(undefined, 1, 2);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCall", Message: "Unnecessary '.call()'.", Line: 1, Column: 1},
				},
			},
			{
				Code: `foo.call(void 0, 1, 2);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCall", Message: "Unnecessary '.call()'.", Line: 1, Column: 1},
				},
			},
			{
				Code: `foo.call(null, 1, 2);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCall", Message: "Unnecessary '.call()'.", Line: 1, Column: 1},
				},
			},
			{
				Code: `obj.foo.call(obj, 1, 2);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCall", Message: "Unnecessary '.call()'.", Line: 1, Column: 1},
				},
			},
			{
				Code: `a.b.c.foo.call(a.b.c, 1, 2);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCall", Message: "Unnecessary '.call()'.", Line: 1, Column: 1},
				},
			},
			{
				Code: `a.b(x, y).c.foo.call(a.b(x, y).c, 1, 2);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCall", Message: "Unnecessary '.call()'.", Line: 1, Column: 1},
				},
			},

			// ---- apply ----
			{
				Code: `foo.apply(undefined, [1, 2]);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCall", Message: "Unnecessary '.apply()'.", Line: 1, Column: 1},
				},
			},
			{
				Code: `foo.apply(void 0, [1, 2]);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCall", Message: "Unnecessary '.apply()'.", Line: 1, Column: 1},
				},
			},
			{
				Code: `foo.apply(null, [1, 2]);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCall", Message: "Unnecessary '.apply()'.", Line: 1, Column: 1},
				},
			},
			{
				Code: `obj.foo.apply(obj, [1, 2]);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCall", Message: "Unnecessary '.apply()'.", Line: 1, Column: 1},
				},
			},
			{
				Code: `a.b.c.foo.apply(a.b.c, [1, 2]);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCall", Message: "Unnecessary '.apply()'.", Line: 1, Column: 1},
				},
			},
			{
				Code: `a.b(x, y).c.foo.apply(a.b(x, y).c, [1, 2]);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCall", Message: "Unnecessary '.apply()'.", Line: 1, Column: 1},
				},
			},
			{
				Code: `[].concat.apply([ ], [1, 2]);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCall", Message: "Unnecessary '.apply()'.", Line: 1, Column: 1},
				},
			},
			{
				Code: "[].concat.apply([\n/*empty*/\n], [1, 2]);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCall", Message: "Unnecessary '.apply()'.", Line: 1, Column: 1},
				},
			},
			{
				Code: `abc.get("foo", 0).concat.apply(abc . get("foo",  0 ), [1, 2]);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCall", Message: "Unnecessary '.apply()'.", Line: 1, Column: 1},
				},
			},

			// ---- Optional chaining ----
			{
				Code: `foo.call?.(undefined, 1, 2);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCall", Message: "Unnecessary '.call()'.", Line: 1, Column: 1},
				},
			},
			{
				Code: `foo?.call(undefined, 1, 2);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCall", Message: "Unnecessary '.call()'.", Line: 1, Column: 1},
				},
			},
			{
				Code: `(foo?.call)(undefined, 1, 2);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCall", Message: "Unnecessary '.call()'.", Line: 1, Column: 1},
				},
			},
			{
				Code: `obj.foo.call?.(obj, 1, 2);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCall", Message: "Unnecessary '.call()'.", Line: 1, Column: 1},
				},
			},
			{
				Code: `obj?.foo.call(obj, 1, 2);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCall", Message: "Unnecessary '.call()'.", Line: 1, Column: 1},
				},
			},
			{
				Code: `(obj?.foo).call(obj, 1, 2);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCall", Message: "Unnecessary '.call()'.", Line: 1, Column: 1},
				},
			},
			{
				Code: `(obj?.foo.call)(obj, 1, 2);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCall", Message: "Unnecessary '.call()'.", Line: 1, Column: 1},
				},
			},
			{
				Code: `obj?.foo.bar.call(obj?.foo, 1, 2);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCall", Message: "Unnecessary '.call()'.", Line: 1, Column: 1},
				},
			},
			{
				Code: `(obj?.foo).bar.call(obj?.foo, 1, 2);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCall", Message: "Unnecessary '.call()'.", Line: 1, Column: 1},
				},
			},
			{
				Code: `obj.foo?.bar.call(obj.foo, 1, 2);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCall", Message: "Unnecessary '.call()'.", Line: 1, Column: 1},
				},
			},

			// ---- ElementAccessExpression as the applied function ----
			// Exercises the `KindElementAccessExpression` branch of the
			// receiver-extraction switch.
			{
				Code: `obj['foo'].call(obj, 1, 2);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCall", Message: "Unnecessary '.call()'.", Line: 1, Column: 1},
				},
			},
			{
				Code: `obj['foo'].apply(obj, [1, 2]);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCall", Message: "Unnecessary '.apply()'.", Line: 1, Column: 1},
				},
			},

			// ---- `this` keyword as receiver/thisArg ----
			{
				Code: `class C { run() { this.foo.call(this, 1, 2); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCall", Message: "Unnecessary '.call()'.", Line: 1, Column: 19},
				},
			},
			{
				Code: `class C { run() { this.foo.apply(this, [1, 2]); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCall", Message: "Unnecessary '.apply()'.", Line: 1, Column: 19},
				},
			},

			// ---- Direct function/arrow call via .call/.apply ----
			// `applied` is not a member access, so expectedThis is nil and
			// the rule falls back to IsNullOrUndefined(thisArg).
			{
				Code: `(() => 1).call(undefined);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCall", Message: "Unnecessary '.call()'.", Line: 1, Column: 1},
				},
			},
			{
				Code: `(function () {}).apply(null, []);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCall", Message: "Unnecessary '.apply()'.", Line: 1, Column: 1},
				},
			},

			// ---- Extra edge cases (token-level equivalence) ----
			// thisArg wrapped in parens is unwrapped by IsNullOrUndefined.
			{
				Code: `foo.call((null), 1, 2);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCall", Message: "Unnecessary '.call()'.", Line: 1, Column: 1},
				},
			},
			{
				Code: `foo.apply((undefined), [1, 2]);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCall", Message: "Unnecessary '.apply()'.", Line: 1, Column: 1},
				},
			},
			// Empty array receiver vs empty array thisArg — different
			// internal whitespace must still token-match.
			{
				Code: `[].concat.apply([/*c*/], [1, 2]);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCall", Message: "Unnecessary '.apply()'.", Line: 1, Column: 1},
				},
			},
			// Empty object receiver token-stream-equal to whitespace-padded form.
			{
				Code: `({}).valueOf.call({ }, );`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCall", Message: "Unnecessary '.call()'.", Line: 1, Column: 1},
				},
			},
			// `.apply(thisArg, [...args])` is a non-variadic apply (args[1] is
			// ArrayLiteral); ESLint flags it the same way as a literal-array
			// apply. The replacement would be `<applied>(...args)`.
			{
				Code: `foo.apply(null, [...args]);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCall", Message: "Unnecessary '.apply()'.", Line: 1, Column: 1},
				},
			},
		},
	)
}

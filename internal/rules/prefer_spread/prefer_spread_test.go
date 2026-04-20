package prefer_spread

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferSpreadRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferSpreadRule,
		// Valid cases - ported from ESLint
		[]rule_tester.ValidTestCase{
			{Code: `foo.apply(obj, args);`},
			{Code: `obj.foo.apply(null, args);`},
			{Code: `obj.foo.apply(otherObj, args);`},
			{Code: `a.b(x, y).c.foo.apply(a.b(x, z).c, args);`},
			{Code: `a.b.foo.apply(a.b.c, args);`},
			{Code: `foo.apply(undefined, [1, 2]);`},
			{Code: `foo.apply(null, [1, 2]);`},
			{Code: `obj.foo.apply(obj, [1, 2]);`},
			{Code: `var apply; foo[apply](null, args);`},
			{Code: `foo.apply();`},
			{Code: `obj.foo.apply();`},
			{Code: `obj.foo.apply(obj, ...args);`},
			// `(a?.b).c` has extra parens around `a?.b`, `a?.b.c` does not — tokens differ
			{Code: `(a?.b).c.foo.apply(a?.b.c, args);`},
			{Code: `a?.b.c.foo.apply((a?.b).c, args);`},
			// Private identifier named `#apply` — `getStaticPropertyName` does not
			// resolve it to "apply", so the member access does not match
			{Code: `class C { #apply; foo() { foo.#apply(undefined, args); } }`},

			// ---- Real-world patterns beyond the ESLint suite ----
			// Identifier thisArg `this` vs identifier `obj` — tokens differ (Kind differs)
			{Code: `obj.foo.apply(this, args);`},
			// `this` receiver vs non-this thisArg — tokens differ
			{Code: `class C { m(args: any) { this.foo.apply(that, args); } }`},
			// `super.foo.apply(this, args)` — expectedThis is `super`, thisArg is `this`
			{Code: `class C extends B { m(args: any) { super.foo.apply(this, args); } }`},
			// Type assertion on receiver only — tokens differ
			{Code: `(obj as any).foo.apply(obj, args);`},
			// Non-empty array receivers with different contents
			{Code: `[1, 2].concat.apply([1, 3], args);`},
			// Hex vs decimal inside receiver — ESLint equalTokens keeps them distinct
			{Code: `[0x1].concat.apply([1], args);`},
			// Trailing comma present on one side only
			{Code: `[a,].concat.apply([a], args);`},
			// Deeply-nested call receiver where arguments differ
			{Code: `outer(inner(x)).m.apply(outer(inner(y)).m, args);`},
			// Bracket access with non-static key — static-value resolution fails
			{Code: `foo[getKey()](null, args);`},
			// Cross-class: string key "#apply" does NOT match the method
			// "apply" (different names), AND never collides with the private
			// identifier class.
			{Code: `foo["#apply"](null, args);`},
			// Different property name entirely
			{Code: `foo.bind(null, args);`},
			// Uppercase method name — not "apply"
			{Code: `foo.APPLY(null, args);`},

			// ---- Parenthesized operands (ESTree paren-transparency) ----
			// `([1, 2])` must be treated as an ArrayExpression (parens are
			// transparent in ESTree), so the rule skips the call.
			{Code: `foo.apply(null, ([1, 2]));`},
			// Nested parens + internal trivia — still an ArrayLiteral
			{Code: "foo.apply(null, (([\n/* x */\n])));"},
		},
		// Invalid cases - ported from ESLint
		[]rule_tester.InvalidTestCase{
			// Lock in exact message text + full reported range (the rule emits
			// a single messageId with no modifier combinations).
			{
				Code: `foo.apply(undefined, args);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferSpread",
						Message:   "Use the spread operator instead of '.apply()'.",
						Line:      1, Column: 1, EndLine: 1, EndColumn: 27,
					},
				},
			},
			{
				Code: `foo.apply(void 0, args);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferSpread", Line: 1, Column: 1},
				},
			},
			{
				Code: `foo.apply(null, args);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferSpread", Line: 1, Column: 1},
				},
			},
			{
				Code: `obj.foo.apply(obj, args);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferSpread", Line: 1, Column: 1},
				},
			},
			{
				Code: `a.b.c.foo.apply(a.b.c, args);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferSpread", Line: 1, Column: 1},
				},
			},
			{
				Code: `a.b(x, y).c.foo.apply(a.b(x, y).c, args);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferSpread", Line: 1, Column: 1},
				},
			},
			// Empty array literal thisArg matched against `[]`
			{
				Code: `[].concat.apply([ ], args);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferSpread", Line: 1, Column: 1},
				},
			},
			{
				Code: "[].concat.apply([\n/*empty*/\n], args);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferSpread", Line: 1, Column: 1},
				},
			},
			// ---- Optional chaining variants ----
			{
				Code: `foo.apply?.(undefined, args);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferSpread", Line: 1, Column: 1},
				},
			},
			{
				Code: `foo?.apply(undefined, args);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferSpread", Line: 1, Column: 1},
				},
			},
			{
				Code: `foo?.apply?.(undefined, args);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferSpread", Line: 1, Column: 1},
				},
			},
			{
				Code: `(foo?.apply)(undefined, args);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferSpread", Line: 1, Column: 1},
				},
			},
			{
				Code: `(foo?.apply)?.(undefined, args);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferSpread", Line: 1, Column: 1},
				},
			},
			{
				Code: `(obj?.foo).apply(obj, args);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferSpread", Line: 1, Column: 1},
				},
			},
			{
				Code: `a?.b.c.foo.apply(a?.b.c, args);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferSpread", Line: 1, Column: 1},
				},
			},
			{
				Code: `(a?.b.c).foo.apply(a?.b.c, args);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferSpread", Line: 1, Column: 1},
				},
			},
			{
				Code: `(a?.b).c.foo.apply((a?.b).c, args);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferSpread", Line: 1, Column: 1},
				},
			},
			// Private identifier `#foo` — the outer `.apply` is still a regular
			// member access, so the rule reports.
			{
				Code: `class C { #foo; foo() { obj.#foo.apply(obj, args); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferSpread", Line: 1, Column: 25},
				},
			},

			// ---- Real-world patterns beyond the ESLint suite ----
			// `this.foo.apply(this, args)` — common class-method pattern
			{
				Code: `class C { m(args: any) { this.foo.apply(this, args); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferSpread", Line: 1, Column: 26},
				},
			},
			// Call-expression receiver: `a.b()` as expected `this`
			{
				Code: `a.b().c.foo.apply(a.b().c, args);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferSpread", Line: 1, Column: 1},
				},
			},
			// Top-level call-expression receiver (no member)
			{
				Code: `getFn().apply(undefined, args);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferSpread", Line: 1, Column: 1},
				},
			},
			// Bracket access with static string as the `.apply` callee
			{
				Code: `foo["apply"](null, args);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferSpread", Line: 1, Column: 1},
				},
			},
			{
				Code: `obj["foo"].apply(obj, args);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferSpread", Line: 1, Column: 1},
				},
			},
			// Non-substitution template literal as bracket key
			{
				Code: "foo[`apply`](null, args);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferSpread", Line: 1, Column: 1},
				},
			},
			// Multiline receiver — newlines / indentation are trivia. Also locks
			// in EndLine / EndColumn to prove the range spans the whole call.
			{
				Code: "obj\n  .foo\n  .apply(obj, args);",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferSpread",
						Line:      1, Column: 1, EndLine: 3, EndColumn: 20,
					},
				},
			},
			// Comments between tokens inside the receiver
			{
				Code: `obj /* x */ . foo . apply(obj, args);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferSpread", Line: 1, Column: 1},
				},
			},
			// Type arguments on the call — should still report
			{
				Code: `foo.apply<any>(undefined, args);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferSpread", Line: 1, Column: 1},
				},
			},
			// Non-empty array receivers with identical contents
			{
				Code: `[1, 2].concat.apply([1, 2], args);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferSpread", Line: 1, Column: 1},
				},
			},
			// Trailing commas on BOTH sides — tokens match
			{
				Code: `[a,].concat.apply([a,], args);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferSpread", Line: 1, Column: 1},
				},
			},
			// Whitespace inside empty array on one side — trivia, tokens match
			{
				Code: "[].concat.apply([\n/* comment */\n], args);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferSpread", Line: 1, Column: 1},
				},
			},
			// `new Foo()` on both sides — ESLint reports (tokens match); migrating
			// to spread changes runtime semantics but the rule follows tokens
			{
				Code: `(new Foo()).bar.apply(new Foo(), args);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferSpread", Line: 1, Column: 1},
				},
			},
			// Call inside arguments: only the inner `.apply` matters
			{
				Code: `wrap(foo.apply(null, args));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferSpread", Line: 1, Column: 6},
				},
			},
			// Deeply-nested call-expression receiver — `expectedThis` is
			// `outer(inner(x))` (the call before `.m`), matched by an
			// identical-token thisArg
			{
				Code: `outer(inner(x)).m.apply(outer(inner(x)), args);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferSpread", Line: 1, Column: 1},
				},
			},
			// ---- Parenthesized operands (ESTree paren-transparency) ----
			// Paren-wrapped `null` as thisArg — IsNullOrUndefined sees through
			{
				Code: `foo.apply((null), args);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferSpread", Line: 1, Column: 1},
				},
			},
			// Paren-wrapped thisArg matching member-access receiver — HasSameTokens sees through outer parens
			{
				Code: `obj.foo.apply((obj), args);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferSpread", Line: 1, Column: 1},
				},
			},
			// Paren-wrapped identifier as args[1] — still NOT an array/spread
			// after stripping parens, so the rule proceeds normally
			{
				Code: `foo.apply(null, (args));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferSpread", Line: 1, Column: 1},
				},
			},
		},
	)
}

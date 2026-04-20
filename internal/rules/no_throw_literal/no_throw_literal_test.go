package no_throw_literal

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoThrowLiteral(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoThrowLiteralRule,
		[]rule_tester.ValidTestCase{
			// ---- ESLint upstream valid cases ----
			{Code: `throw new Error();`},
			{Code: `throw new Error('error');`},
			{Code: `throw Error('error');`},
			{Code: `var e = new Error(); throw e;`},
			{Code: `try {throw new Error();} catch (e) {throw e;};`},
			{Code: `throw a;`},          // Identifier
			{Code: `throw foo();`},      // CallExpression
			{Code: `throw new foo();`},  // NewExpression
			{Code: `throw foo.bar;`},    // PropertyAccessExpression (ESTree MemberExpression)
			{Code: `throw foo[bar];`},   // ElementAccessExpression (ESTree MemberExpression)
			{Code: `class C { #field: any; foo() { throw foo.#field; } }`}, // private field member access
			{Code: `throw foo = new Error();`},                              // AssignmentExpression `=`
			{Code: `throw foo.bar ||= 'literal'`},                           // logical-assign `||=` (left could be Error)
			{Code: `throw foo[bar] ??= 'literal'`},                          // logical-assign `??=` (left could be Error)
			{Code: `throw 1, 2, new Error();`},                              // SequenceExpression (last is Error)
			{Code: `throw 'literal' && new Error();`},                       // LogicalExpression `&&` (right is Error)
			{Code: `throw new Error() || 'literal';`},                       // LogicalExpression `||` (left is Error)
			{Code: `throw foo ? new Error() : 'literal';`},                  // ConditionalExpression (consequent)
			{Code: `throw foo ? 'literal' : new Error();`},                  // ConditionalExpression (alternate)
			{Code: "throw tag `${foo}`;"},                                   // TaggedTemplateExpression
			{Code: `function* foo() { var index = 0; throw yield index++; }`}, // YieldExpression
			{Code: `async function foo() { throw await bar; }`},               // AwaitExpression
			{Code: `throw obj?.foo`},                                          // optional chain (PropertyAccess)
			{Code: `throw obj?.foo()`},                                        // optional chain (CallExpression)

			// ---- tsgo-/TS-specific edge cases ----
			// Parenthesized argument is unwrapped before the Identifier/`undefined` check.
			{Code: `throw (foo);`},
			// Nested parens around an Error-producing call.
			{Code: `throw ((new Error()));`},
			// Conditional with both branches Error-shaped.
			{Code: `throw cond ? new Error() : foo();`},
			// Sequence in parens — couldBeError sees the rightmost; the
			// outer node is BinaryExpression, so IsUndefinedIdentifier returns
			// false even when the rightmost is `undefined`.
			{Code: `throw (1, new Error());`},
			{Code: `throw (1, undefined);`},
			// Compound logical assignment where the RHS is Error-shaped.
			{Code: `throw foo &&= new Error();`},
			// Nested logical: any operand could be Error.
			{Code: `throw a || b || new Error();`},
			{Code: `throw a ?? new Error();`},
			{Code: `throw (cond ? new Error() : foo()) || 'literal';`},
			// Optional element access.
			{Code: `throw obj?.[key];`},
			// Triple-wrapped Identifier `foo` is plainly valid.
			{Code: `throw (((foo)));`},
			// Yield inside a generator with await.
			{Code: `async function* foo() { throw await (yield 1); }`},
			// TS assertion wrappers are NOT transparent (matching upstream
			// ESLint when run on a `.ts` file via @typescript-eslint/parser),
			// but the OUTER access shape still drives the classification:
			// `foo!.bar` is a PropertyAccessExpression at the top level →
			// could-be-Error → valid. Same for `(foo as Error).bar`.
			{Code: `throw foo!.bar;`},
			{Code: `throw (foo as Error).bar;`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- ESLint upstream invalid cases ----
			{
				Code: `throw 'error';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "object", Message: "Expected an error object to be thrown.", Line: 1, Column: 1},
				},
			},
			{
				Code: `throw 0;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "object", Message: "Expected an error object to be thrown.", Line: 1, Column: 1},
				},
			},
			{
				Code: `throw false;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "object", Message: "Expected an error object to be thrown.", Line: 1, Column: 1},
				},
			},
			{
				Code: `throw null;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "object", Message: "Expected an error object to be thrown.", Line: 1, Column: 1},
				},
			},
			{
				Code: `throw {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "object", Message: "Expected an error object to be thrown.", Line: 1, Column: 1},
				},
			},
			{
				Code: `throw undefined;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Message: "Do not throw undefined.", Line: 1, Column: 1},
				},
			},

			// ---- String concatenation ----
			{
				Code: `throw 'a' + 'b';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "object", Line: 1, Column: 1},
				},
			},
			{
				Code: `var b = new Error(); throw 'a' + b;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "object", Line: 1, Column: 22},
				},
			},

			// ---- AssignmentExpression ----
			{
				Code: `throw foo = 'error';`, // RHS is a literal
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "object", Line: 1, Column: 1},
				},
			},
			{
				Code: `throw foo += new Error();`, // arithmetic compound-assign returns a primitive
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "object", Line: 1, Column: 1},
				},
			},
			{
				Code: `throw foo &= new Error();`, // bitwise compound-assign returns a primitive
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "object", Line: 1, Column: 1},
				},
			},
			{
				Code: `throw foo &&= 'literal'`, // either falsy `foo` or 'literal'; neither can be an Error
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "object", Line: 1, Column: 1},
				},
			},

			// ---- SequenceExpression ----
			{
				Code: `throw new Error(), 1, 2, 3;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "object", Line: 1, Column: 1},
				},
			},

			// ---- LogicalExpression ----
			{
				Code: `throw 'literal' && 'not an Error';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "object", Line: 1, Column: 1},
				},
			},
			{
				Code: `throw foo && 'literal'`, // either falsy `foo` (not Error) or 'literal'
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "object", Line: 1, Column: 1},
				},
			},

			// ---- ConditionalExpression ----
			{
				Code: `throw foo ? 'not an Error' : 'literal';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "object", Line: 1, Column: 1},
				},
			},

			// ---- TemplateLiteral ----
			{
				Code: "throw `${err}`;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "object", Line: 1, Column: 1},
				},
			},

			// ---- Extra tsgo/TS edge cases ----
			// Parenthesized literal still reported.
			{
				Code: `throw ('error');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "object", Line: 1, Column: 1},
				},
			},
			// Parenthesized `undefined` still maps to undef.
			{
				Code: `throw (undefined);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 1},
				},
			},
			// Sequence whose last element is a literal.
			{
				Code: `throw (foo(), 'error');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "object", Line: 1, Column: 1},
				},
			},
			// Conditional with both branches non-Error.
			{
				Code: `throw cond ? 'a' : 'b';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "object", Line: 1, Column: 1},
				},
			},
			// BigInt literal.
			{
				Code: `throw 1n;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "object", Line: 1, Column: 1},
				},
			},
			// Regex literal.
			{
				Code: `throw /foo/;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "object", Line: 1, Column: 1},
				},
			},
			// `void 0` evaluates to undefined at runtime, but is a UnaryExpression,
			// not the Identifier `undefined`. ESLint reports "object" (not "undef")
			// because it falls through couldBeError's default case.
			{
				Code: `throw void 0;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "object", Line: 1, Column: 1},
				},
			},
			// ClassExpression / FunctionExpression / ArrowFunction / ArrayLiteral
			// / ObjectLiteral / `this` / `new.target`: none are in couldBeError's
			// node-type list, so they all report "object".
			{
				Code: `throw class {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "object", Line: 1, Column: 1},
				},
			},
			{
				Code: `throw function() {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "object", Line: 1, Column: 1},
				},
			},
			{
				Code: `throw () => {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "object", Line: 1, Column: 1},
				},
			},
			{
				Code: `throw [];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "object", Line: 1, Column: 1},
				},
			},
			{
				Code: `class C { foo() { throw this; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "object", Line: 1, Column: 19},
				},
			},
			{
				Code: `function foo() { throw new.target; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "object", Line: 1, Column: 18},
				},
			},
			// Nested logical short-circuit chain that cannot resolve to Error.
			{
				Code: `throw a && b && 'literal';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "object", Line: 1, Column: 1},
				},
			},
			// ---- TS assertion wrappers: NOT transparent in upstream ESLint ----
			// Verified empirically: ESLint core run on `.ts` via
			// `@typescript-eslint/parser` reports "object" for all of these,
			// because TSAsExpression / TSNonNullExpression / TSSatisfiesExpression
			// / TSTypeAssertion are absent from `astUtils.couldBeError` and
			// fall through its default branch.
			{
				Code: `throw foo as Error;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "object", Line: 1, Column: 1},
				},
			},
			{
				Code: `throw foo!;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "object", Line: 1, Column: 1},
				},
			},
			{
				Code: `throw foo satisfies unknown;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "object", Line: 1, Column: 1},
				},
			},
			{
				Code: `throw <Error>foo;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "object", Line: 1, Column: 1},
				},
			},
			// Non-null on optional chain: outer node is NonNullExpression, not
			// PropertyAccessExpression → falls through → "object".
			{
				Code: `throw obj?.foo!;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "object", Line: 1, Column: 1},
				},
			},
			// Underlying literal still reported.
			{
				Code: `throw 'foo' as Error;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "object", Line: 1, Column: 1},
				},
			},
			{
				Code: `throw (5 as any);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "object", Line: 1, Column: 1},
				},
			},
			// `undefined` after parens IS detected as undef (parens transparent).
			{
				Code: `throw (((undefined)));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef", Line: 1, Column: 1},
				},
			},
			// `undefined as any` / `undefined!` are TS assertion wrappers — not
			// transparent — so they report "object", not "undef".
			{
				Code: `throw undefined as any;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "object", Line: 1, Column: 1},
				},
			},
			{
				Code: `throw undefined!;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "object", Line: 1, Column: 1},
				},
			},

			// ---- JSX (TSX) ----
			// JSXElement / JSXFragment / JsxSelfClosingElement are not in
			// couldBeError → "object".
			{
				Code:   `const _ = () => { throw <div />; };`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "object", Line: 1, Column: 19}},
			},
			{
				Code:   `const _ = () => { throw <></>; };`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "object", Line: 1, Column: 19}},
			},
			// Both branches of a conditional are JSX → couldBeError false on both.
			{
				Code:   `const _ = (c: boolean) => { throw c ? <div /> : <span />; };`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "object", Line: 1, Column: 29}},
			},

			// ---- Full Line/Column/EndLine/EndColumn assertions ----
			// Single-line: report range spans the entire ThrowStatement
			// (`throw 'error';` is 14 columns, so end is column 15).
			{
				Code: `throw 'error';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "object", Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
				},
			},
			// Multi-line throw expression — ASI does not fire because the
			// next token after `throw` is `'a'` (no line break between them).
			// Range must span both lines.
			{
				Code: "throw 'a' +\n  'b';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "object", Line: 1, Column: 1, EndLine: 2, EndColumn: 7},
				},
			},
		},
	)
}

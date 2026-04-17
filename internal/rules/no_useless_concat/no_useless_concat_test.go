package no_useless_concat

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUselessConcat(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUselessConcatRule,
		[]rule_tester.ValidTestCase{
			// Non-`+` operators must not trigger.
			{Code: `var a = 1 + 1;`},
			{Code: `var a = 1 - 2;`},
			{Code: `var a = 1 * '2';`},
			{Code: `var a = 'a' * 'b';`},
			{Code: `var a = 'a' - 'b';`},
			{Code: `var a = 'a' < 'b';`},

			// At least one neighbor of every `+` is not a string literal.
			{Code: `var a = foo + bar;`},
			{Code: `var a = 'foo' + bar;`},
			{Code: `var a = 1 + '1';`},
			{Code: "var a = 1 + `1`;"},
			{Code: "var a = `1` + 1;"},
			{Code: `var a = 'a' + 1;`},
			{Code: `var a = foo + 'a' + bar;`},
			{Code: `var a = 'a' + 1 + 'b';`},
			{Code: `var a = 1 + 'a' + 2;`},
			{Code: `var a = a + 'b' + c + 'd';`},
			{Code: `var a = foo() + 'b';`},
			{Code: `var a = 'a' + foo();`},
			{Code: `var a = 'a' + (b ? 'c' : 'd');`},
			{Code: `var a = 'a' + (b as string);`},
			{Code: `var a = ('a' as const) + 'b';`},

			// Unary `+`/`-` produces a PrefixUnaryExpression, not a string literal.
			{Code: "var a = (1 + +2) + `b`;"},
			{Code: `var a = +'1' + 100;`},
			{Code: `var a = -'a' + 'b';`},

			// Multi-line concatenation is explicitly allowed by ESLint.
			{Code: "var foo = 'foo' +\n 'bar';"},
			{Code: "var a = ('a') +\n ('b');"},
			{Code: "var a = 'a' +\n// comment\n'b';"},
			{Code: "var a = 'a' + /*\n*/ 'b';"},

			// Compound assignment is not a `+` BinaryExpression.
			{Code: `x += 'y';`},
			{Code: "x += `y`;"},

			// TaggedTemplateExpression is not a TemplateLiteral.
			{Code: "var a = tag`a` + 'b';"},
			{Code: "var a = 'a' + tag`b`;"},

			// `as` / `satisfies` are postfix; their output is not a string literal node.
			{Code: `var a = ('a' as const) + ('b' as const);`},
			{Code: `var a = 'a' + ('b' satisfies string);`},

			// Logical / bitwise / comparison operators — not `+`.
			{Code: `var a = 'a' || 'b';`},
			{Code: `var a = 'a' && 'b';`},
			{Code: `var a = 'a' ?? 'b';`},

			// Comma expression with non-string neighbor of `+`.
			{Code: `var a = (1, 'b') + foo;`},

			// Multiple prefix unary operators on the left operand.
			{Code: `var a = + + 'a' + 'b';`},

			// Call expression on either side.
			{Code: `var a = foo() + 'b';`},
			{Code: `var a = 'a' + foo();`},

			// `this`, optional chain, non-null assertion — all non-literal.
			{Code: `class C { m() { return this + 'a'; } }`},
			{Code: `var a = x?.y + 'z';`},
			{Code: `var a = x! + 'z';`},
		},
		[]rule_tester.InvalidTestCase{
			// Basic two-literal concatenation.
			{
				Code: `'a' + 'b'`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 5},
				},
			},
			{
				Code: "`a` + 'b'",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 5},
				},
			},
			{
				Code: "`a` + `b`",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 5},
				},
			},
			// Empty literals still count.
			{
				Code: `'' + ''`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 4},
				},
			},
			{
				Code: "`` + ``",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 4},
				},
			},

			// Templates with substitutions (TemplateExpression).
			{
				Code: "`a${x}` + 'b'",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 9},
				},
			},
			{
				Code: "'a' + `b${x}`",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 5},
				},
			},
			{
				Code: "`${x}a` + `b${y}`",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 9},
				},
			},

			// Left-associative chains: each `+` joining two literals is reported.
			{
				Code: `foo + 'a' + 'b'`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 11},
				},
			},
			{
				Code: `'a' + 'b' + 'c'`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 11},
					{MessageId: "unexpectedConcat", Line: 1, Column: 5},
				},
			},
			{
				Code: `'a' + 'b' + 'c' + 'd'`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 17},
					{MessageId: "unexpectedConcat", Line: 1, Column: 11},
					{MessageId: "unexpectedConcat", Line: 1, Column: 5},
				},
			},
			{
				Code: `foo + 'a' + 'b' + 'c'`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 17},
					{MessageId: "unexpectedConcat", Line: 1, Column: 11},
				},
			},
			{
				Code: `'a' + 'b' + foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 5},
				},
			},

			// Parentheses (including multi-layer) are transparent.
			{
				Code: `(foo + 'a') + ('b' + 'c')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 13},
					{MessageId: "unexpectedConcat", Line: 1, Column: 20},
				},
			},
			{
				Code: `('a') + ('b')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 7},
				},
			},
			{
				Code: `(('a')) + (('b'))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 9},
				},
			},
			// Right-associative parenthesized chain.
			{
				Code: `'a' + ('b' + 'c')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 5},
					{MessageId: "unexpectedConcat", Line: 1, Column: 12},
				},
			},
			{
				Code: `('a' + 'b') + ('c' + 'd')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 13},
					{MessageId: "unexpectedConcat", Line: 1, Column: 6},
					{MessageId: "unexpectedConcat", Line: 1, Column: 20},
				},
			},
			{
				Code: `'a' + ('b' + 'c') + 'd'`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 19},
					{MessageId: "unexpectedConcat", Line: 1, Column: 5},
					{MessageId: "unexpectedConcat", Line: 1, Column: 12},
				},
			},
			{
				Code: `(foo + 'a' + 'b') + 'c'`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 19},
					{MessageId: "unexpectedConcat", Line: 1, Column: 12},
				},
			},

			// Multi-line mixed: only same-line `+` fires.
			{
				Code: "'a' +\n'b' + 'c'",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 2, Column: 5},
				},
			},
			{
				Code: "'a' + 'b' +\n'c'",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 5},
				},
			},
			// Comments between operator and operand should not hide same-line cases.
			{
				Code: `'a' + /* x */ 'b'`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 5},
				},
			},

			// Template-literal chains, including `foo + `a` + `b``.
			{
				Code: "foo + `a` + `b`",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 11},
				},
			},

			// Nested inside other expressions.
			{
				Code: `foo('a' + 'b')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 9},
				},
			},
			{
				Code: `var x = {a: 'a' + 'b'};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 17},
				},
			},
			{
				Code: `var x = ['a' + 'b'];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 14},
				},
			},
			{
				Code: "`${'a' + 'b'}`",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 8},
				},
			},
			{
				Code: `var x = 'a' + 'b';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 13},
				},
			},
			{
				Code: `x = 'a' + 'b'`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 9},
				},
			},
			{
				Code: `x += 'a' + 'b'`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 10},
				},
			},
			{
				Code: `function f() { return 'a' + 'b'; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 27},
				},
			},
			{
				Code: `('a' + 'b') === 'c'`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 6},
				},
			},

			// Special characters inside literals.
			{
				Code: `'\n' + '\t'`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 6},
				},
			},
			{
				Code: `'🎉' + '🎊'`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 6},
				},
			},

			// Deep right-associative nesting via parentheses.
			{
				Code: `'a' + ('b' + ('c' + 'd'))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 5},
					{MessageId: "unexpectedConcat", Line: 1, Column: 12},
					{MessageId: "unexpectedConcat", Line: 1, Column: 19},
				},
			},
			// Asymmetric tree mixing left-/right-associative chunks.
			{
				Code: `(('a' + 'b') + ('c' + ('d' + 'e'))) + 'f'`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 37},
					{MessageId: "unexpectedConcat", Line: 1, Column: 14},
					{MessageId: "unexpectedConcat", Line: 1, Column: 7},
					{MessageId: "unexpectedConcat", Line: 1, Column: 21},
					{MessageId: "unexpectedConcat", Line: 1, Column: 28},
				},
			},
			// Concatenation inside a template-literal placeholder.
			{
				Code: "`a${'x' + 'y'}b` + 'c'",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 18},
					{MessageId: "unexpectedConcat", Line: 1, Column: 9},
				},
			},
			// Line comment forces the right operand onto a new line.
			{
				Code: "'a' + 'b' // same-line comment\n+ 'c'",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 5},
				},
			},

			// Conditional expression — both branches concatenate.
			{
				Code: `x ? 'a' + 'b' : 'c' + 'd'`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 9},
					{MessageId: "unexpectedConcat", Line: 1, Column: 21},
				},
			},
			// Comma expression separates two chains.
			{
				Code: `('a' + 'b', 'c' + 'd')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 6},
					{MessageId: "unexpectedConcat", Line: 1, Column: 17},
				},
			},
			// Logical operators are not `+`, but nested `+` still fires.
			{
				Code: `'a' + 'b' && 'c'`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 5},
				},
			},
			{
				Code: `'a' + 'b' ?? 'c'`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 5},
				},
			},
			// Default parameter value.
			{
				Code: `function f(x = 'a' + 'b') { return x; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 20},
				},
			},
			// Arrow function body.
			{
				Code: `const f = () => 'a' + 'b';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 21},
				},
			},
			// Computed property name.
			{
				Code: `const o = {['a' + 'b']: 1};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 17},
				},
			},
			// Destructuring default.
			{
				Code: `const {x = 'a' + 'b'} = o;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 16},
				},
			},
			// Class field initializer.
			{
				Code: `class C { x = 'a' + 'b'; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 19},
				},
			},
			// Physically multi-line template literal — end/start both on the final line.
			{
				Code: "`a\nb` + 'c'",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 2, Column: 4},
				},
			},

			// JSX expression container.
			{
				Code: `const el = <div>{'a' + 'b'}</div>;`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 22},
				},
			},
			// JSX attribute value.
			{
				Code: `const el = <div title={'a' + 'b'} />;`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedConcat", Line: 1, Column: 28},
				},
			},
		},
	)
}

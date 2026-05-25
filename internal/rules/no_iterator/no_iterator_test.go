package no_iterator

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoIteratorRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoIteratorRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// Computed property with identifier (not a string literal)
			{Code: `var a = test[__iterator__];`},
			// Variable declaration, not member access
			{Code: `var __iterator__ = null;`},
			// Template literal missing trailing __
			{Code: "foo[`__iterator`] = null;"},
			// Template literal with newline (not exact match)
			{Code: "foo[`__iterator__\n`] = null;"},
			// Template literal with expression (TemplateExpression, not NoSubstitutionTemplateLiteral)
			{Code: "foo[`__iterator__${x}`] = null;"},
			// Object property key (not a member access)
			{Code: `var obj = { __iterator__: 1 };`},
			// Object shorthand
			{Code: `var obj = { __iterator__ };`},
			// Destructuring (not a member access)
			{Code: `const { __iterator__ } = obj;`},
			{Code: `const { __iterator__: alias } = obj;`},
			// Function declaration
			{Code: `function __iterator__() {}`},
			// Class method declaration
			{Code: `class Foo { __iterator__() {} }`},
			// TypeScript interface property
			{Code: `interface Foo { __iterator__: any; }`},
			// TypeScript type alias
			{Code: `type Foo = { __iterator__: any };`},
			// Parameter name
			{Code: `function foo(__iterator__: any) {}`},
			// Enum member
			{Code: `enum Foo { __iterator__ }`},
			// Private identifier (KindPrivateIdentifier, not KindIdentifier)
			{Code: `class Foo { #__iterator__ = 1; m() { this.#__iterator__; } }`},
			// TypeScript QualifiedName in type position (not PropertyAccessExpression)
			{Code: `declare namespace A { var __iterator__: number; } type X = typeof A.__iterator__;`},
			// Import/export specifier
			{Code: `export { __iterator__ } from 'mod';`},
			// Label (not a member access)
			{Code: `__iterator__: for (;;) { break __iterator__; }`},
			// Computed with string concatenation (not a static literal)
			{Code: `obj['__iterator' + '__'];`},
			// Computed with variable
			{Code: `const key = '__iterator__'; obj[key];`},
			// Getter/setter declaration in object
			{Code: `var obj = { get __iterator__() { return 1; } };`},
			// JSX attribute (not a member access)
			{Code: `var x = <Component __iterator__={value} />;`, Tsx: true},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// ========================================
			// Basic access patterns
			// ========================================
			// Dot notation
			{
				Code: `var a = test.__iterator__;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 9},
				},
			},
			// Bracket with single-quoted string
			{
				Code: `var a = test['__iterator__'];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 9},
				},
			},
			// Bracket with double-quoted string
			{
				Code: `var a = test["__iterator__"];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 9},
				},
			},
			// Bracket with template literal (no substitution)
			{
				Code: "var a = test[`__iterator__`];",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 9},
				},
			},

			// ========================================
			// Assignment targets
			// ========================================
			{
				Code: `Foo.prototype.__iterator__ = function() {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 1},
				},
			},
			{
				Code: "test[`__iterator__`] = function () {};",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 1},
				},
			},
			{
				Code: `obj.__iterator__ = 42;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 1},
				},
			},

			// ========================================
			// Chained member access
			// ========================================
			{
				Code: `a.b.__iterator__;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 1},
				},
			},
			{
				Code: `a.b.c.__iterator__;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 1},
				},
			},
			// Access on result of __iterator__ (2 errors: a.__iterator__ and a.__iterator__.__iterator__)
			{
				Code: `a.__iterator__.__iterator__;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 1},
					{MessageId: "noIterator", Line: 1, Column: 1},
				},
			},

			// ========================================
			// Optional chaining
			// ========================================
			{
				Code: `obj?.__iterator__;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 1},
				},
			},
			{
				Code: `obj?.['__iterator__'];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 1},
				},
			},
			{
				Code: `a?.b?.__iterator__;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 1},
				},
			},

			// ========================================
			// Parenthesized / TypeScript outer expressions
			// ========================================
			{
				Code: `(obj).__iterator__;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 1},
				},
			},
			{
				Code: `((obj)).__iterator__;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 1},
				},
			},
			// Non-null assertion on object
			{
				Code: `obj!.__iterator__;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 1},
				},
			},
			// Type assertion on object
			{
				Code: `(obj as any).__iterator__;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 1},
				},
			},
			// Angle-bracket type assertion
			{
				Code: `(<any>obj).__iterator__;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 1},
				},
			},

			// ========================================
			// this / class contexts
			// ========================================
			{
				Code: `this.__iterator__;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 1},
				},
			},
			{
				Code: `class A { m() { this.__iterator__; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 17},
				},
			},
			{
				Code: `class A { x = this.__iterator__; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 15},
				},
			},
			{
				Code: `class A { static { this.__iterator__; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 20},
				},
			},

			// ========================================
			// Nested in functions / arrows
			// ========================================
			{
				Code: `function foo() { obj.__iterator__; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 18},
				},
			},
			{
				Code: `const f = () => obj.__iterator__;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 17},
				},
			},
			{
				Code: `const f = () => { return obj.__iterator__; };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 26},
				},
			},

			// ========================================
			// In expressions
			// ========================================
			// Conditional expression
			{
				Code: `var x = cond ? obj.__iterator__ : null;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 16},
				},
			},
			// As function argument
			{
				Code: `foo(obj.__iterator__);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 5},
				},
			},
			// In template literal expression
			{
				Code: "var x = `${obj.__iterator__}`;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 12},
				},
			},

			// ========================================
			// Multiple violations in one file
			// ========================================
			{
				Code: "a.__iterator__;\nb['__iterator__'];",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 1},
					{MessageId: "noIterator", Line: 2, Column: 1},
				},
			},

			// ========================================
			// In loops
			// ========================================
			{
				Code: `for (var x in obj.__iterator__) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 15},
				},
			},
			{
				Code: `while (obj.__iterator__) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 8},
				},
			},

			// ========================================
			// Unary / delete / typeof / void
			// ========================================
			{
				Code: `delete obj.__iterator__;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 8},
				},
			},
			{
				Code: `typeof obj.__iterator__;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 8},
				},
			},
			{
				Code: `void obj.__iterator__;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 6},
				},
			},

			// ========================================
			// Call / new / tagged template
			// ========================================
			{
				Code: `obj.__iterator__();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 1},
				},
			},
			{
				Code: `new obj.__iterator__();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 5},
				},
			},

			// ========================================
			// super
			// ========================================
			{
				Code: `class A extends B { m() { super.__iterator__; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 27},
				},
			},
			{
				Code: `class A extends B { m() { super['__iterator__']; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 27},
				},
			},

			// ========================================
			// Spread / array / object value
			// ========================================
			{
				Code: `var a = [obj.__iterator__];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 10},
				},
			},
			{
				Code: `var a = { x: obj.__iterator__ };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 14},
				},
			},
			{
				Code: `var a = { ...obj.__iterator__ };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 14},
				},
			},

			// ========================================
			// Logical / nullish / assignment operators
			// ========================================
			{
				Code: `obj.__iterator__ || fallback;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 1},
				},
			},
			{
				Code: `obj.__iterator__ ?? fallback;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 1},
				},
			},
			{
				Code: `obj.__iterator__ ??= fallback;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 1},
				},
			},

			// ========================================
			// Async / generator
			// ========================================
			{
				Code: `async function f() { await obj.__iterator__; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 28},
				},
			},
			{
				Code: `function* g() { yield obj.__iterator__; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 23},
				},
			},

			// ========================================
			// Mixed bracket + dot chaining
			// ========================================
			{
				Code: `obj['__iterator__'].__iterator__;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 1},
					{MessageId: "noIterator", Line: 1, Column: 1},
				},
			},

			// ========================================
			// Satisfies (TS 4.9+)
			// ========================================
			{
				Code: `(obj satisfies any).__iterator__;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 1},
				},
			},

			// ========================================
			// Escape sequences in string/template literals
			// ========================================
			// Hex escape in string literal → cooked value is __iterator__
			{
				Code: `obj['\x5F\x5Fiterator\x5F\x5F'];`, // cspell:disable-line
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 1},
				},
			},
			// Unicode escape in string literal
			{
				Code: `obj['\u005F\u005Fiterator\u005F\u005F'];`, // cspell:disable-line
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 1},
				},
			},
			// Hex escape in template literal
			{
				Code: "obj[`\\x5F\\x5Fiterator\\x5F\\x5F`];", // cspell:disable-line
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 1},
				},
			},
			// Unicode escape in dot notation identifier
			{
				Code: `obj.\u005F\u005Fiterator\u005F\u005F;`, // cspell:disable-line
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 1},
				},
			},

			// ========================================
			// Multiline
			// ========================================
			{
				Code: "obj\n  .__iterator__;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 1},
				},
			},
			{
				Code: "obj\n  ['__iterator__'];",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noIterator", Line: 1, Column: 1},
				},
			},
		},
	)
}

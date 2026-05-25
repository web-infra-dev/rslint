package no_multi_str

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoMultiStrRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoMultiStrRule,
		// ================================================================
		// Valid cases — should NOT trigger
		// ================================================================
		[]rule_tester.ValidTestCase{
			// ---- Single-line strings ----
			{Code: `var a = 'Line 1 Line 2';`},
			{Code: `var a = "double quoted";`},
			{Code: `var a = '';`},
			{Code: `var a = "";`},

			// ---- Escape sequences that look like newlines but are NOT real ----
			// \n escape → produces newline in value, but raw source stays single-line
			{Code: `var a = 'hello\nworld';`},
			{Code: `var a = "hello\nworld";`},
			// \r escape
			{Code: `var a = 'hello\rworld';`},
			// \u2028 / \u2029 escape
			{Code: `var a = 'hello\u2028world';`},
			{Code: `var a = 'hello\u2029world';`},

			// ---- Template literals — allowed to span multiple lines ----
			{Code: "var a = `Line 1\nLine 2`;"},
			{Code: "var a = `\n`;"},

			// ---- String concatenation across lines (each string is single-line) ----
			{Code: "var a = 'Line 1' +\n'Line 2';"},

			// ---- JSX contexts — string literals in JSX are exempt ----
			// JSX attribute (direct string): parent = JsxAttribute
			{Code: `var a = <div attr="hello"></div>;`, Tsx: true},
			// JSX attribute (expression): parent = JsxExpression
			{Code: `var a = <div attr={'hello'}></div>;`, Tsx: true},
			// JSX child expression: parent = JsxExpression
			{Code: `var a = <div>{'hello'}</div>;`, Tsx: true},
			// Multiline string inside JSX expression — exempt because parent is JsxExpression
			{Code: "<div attr={'foo\\\nbar'}></div>;", Tsx: true},
			{Code: "<div>{'foo\\\nbar'}</div>;", Tsx: true},

			// ---- Non-string expressions ----
			{Code: `var a = 42;`},
			{Code: `var a = true;`},
		},
		// ================================================================
		// Invalid cases — should trigger
		// ================================================================
		[]rule_tester.InvalidTestCase{
			// ================================================================
			// Dimension 1: Line break types
			// ================================================================

			// LF (\n)
			{
				Code: "var x = 'Line 1 \\\n Line 2'",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multilineString", Line: 1, Column: 9},
				},
			},
			// CR (\r)
			{
				Code: "'foo\\\rbar';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multilineString", Line: 1, Column: 1},
				},
			},
			// CRLF (\r\n)
			{
				Code: "'foo\\\r\nbar';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multilineString", Line: 1, Column: 1},
				},
			},
			// LS (U+2028)
			{
				Code: "'foo\\\u2028bar';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multilineString", Line: 1, Column: 1},
				},
			},
			// PS (U+2029)
			{
				Code: "'foo\\\u2029ar';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multilineString", Line: 1, Column: 1},
				},
			},

			// ================================================================
			// Dimension 2: Quote types
			// ================================================================

			// Single quote
			{
				Code: "'foo\\\nbar';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multilineString", Line: 1, Column: 1},
				},
			},
			// Double quote
			{
				Code: "\"foo\\\nbar\";",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multilineString", Line: 1, Column: 1},
				},
			},

			// ================================================================
			// Dimension 3: Nesting contexts
			// ================================================================

			// Function call argument
			{
				Code: "test('Line 1 \\\n Line 2');",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multilineString", Line: 1, Column: 6},
				},
			},
			// Object property value
			{
				Code: "var obj = { key: 'foo\\\nbar' };",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multilineString", Line: 1, Column: 18},
				},
			},
			// Computed property key
			{
				Code: "var obj = { ['foo\\\nbar']: 1 };",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multilineString", Line: 1, Column: 14},
				},
			},
			// Array element
			{
				Code: "var arr = ['foo\\\nbar'];",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multilineString", Line: 1, Column: 12},
				},
			},
			// Return statement
			{
				Code: "function f() { return 'foo\\\nbar'; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multilineString", Line: 1, Column: 23},
				},
			},
			// Conditional / ternary
			{
				Code: "var a = x ? 'foo\\\nbar' : 'safe';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multilineString", Line: 1, Column: 13},
				},
			},
			// Default parameter
			{
				Code: "function f(x = 'foo\\\nbar') {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multilineString", Line: 1, Column: 16},
				},
			},
			// Class property
			{
				Code: "class A { prop = 'foo\\\nbar' }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multilineString", Line: 1, Column: 18},
				},
			},
			// Arrow function implicit return
			{
				Code: "const f = () => 'foo\\\nbar';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multilineString", Line: 1, Column: 17},
				},
			},
			// Binary expression (+ concatenation)
			{
				Code: "var a = 'foo\\\nbar' + x;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multilineString", Line: 1, Column: 9},
				},
			},
			// Logical OR
			{
				Code: "var a = x || 'foo\\\nbar';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multilineString", Line: 1, Column: 14},
				},
			},
			// Nullish coalescing
			{
				Code: "var a = x ?? 'foo\\\nbar';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multilineString", Line: 1, Column: 14},
				},
			},
			// Assignment
			{
				Code: "x = 'foo\\\nbar';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multilineString", Line: 1, Column: 5},
				},
			},
			// Throw statement
			{
				Code: "throw 'foo\\\nbar';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multilineString", Line: 1, Column: 7},
				},
			},
			// Switch case
			{
				Code: "switch (x) { case 'foo\\\nbar': break; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multilineString", Line: 1, Column: 19},
				},
			},
			// Computed member access
			{
				Code: "var a = obj['foo\\\nbar'];",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multilineString", Line: 1, Column: 13},
				},
			},
			// Parenthesized expression
			{
				Code: "var a = ('foo\\\nbar');",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multilineString", Line: 1, Column: 10},
				},
			},
			// Nested function call
			{
				Code: "fn1(fn2('foo\\\nbar'));",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multilineString", Line: 1, Column: 9},
				},
			},
			// Nested object
			{
				Code: "var obj = { a: { b: 'foo\\\nbar' } };",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multilineString", Line: 1, Column: 21},
				},
			},
			// Template literal expression (string inside template)
			{
				Code: "var a = `${'foo\\\nbar'}`;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multilineString", Line: 1, Column: 12},
				},
			},

			// ================================================================
			// Dimension 4: TypeScript-specific contexts
			// ================================================================

			// Enum initializer
			{
				Code: "enum E { A = 'foo\\\nbar' }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multilineString", Line: 1, Column: 14},
				},
			},
			// Type position (string literal type)
			{
				Code: "type T = 'foo\\\nbar';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multilineString", Line: 1, Column: 10},
				},
			},

			// ================================================================
			// Dimension 5: Edge cases
			// ================================================================

			// String spanning 3+ lines
			{
				Code: "'line1\\\nline2\\\nline3';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multilineString", Line: 1, Column: 1},
				},
			},
			// Empty continuation (backslash + newline + closing quote)
			{
				Code: "'\\\n';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multilineString", Line: 1, Column: 1},
				},
			},
			// String not at beginning of file (later line)
			{
				Code: "var x;\nvar y = 'foo\\\nbar';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multilineString", Line: 2, Column: 9},
				},
			},

			// ================================================================
			// Dimension 6: Multiple errors in one file
			// ================================================================

			// Two multiline strings in separate statements
			{
				Code: "'foo\\\nbar';\n'baz\\\nqux';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multilineString", Line: 1, Column: 1},
					{MessageId: "multilineString", Line: 3, Column: 1},
				},
			},
			// Two multiline strings as function arguments
			{
				Code: "fn('one\\\ntwo', 'three\\\nfour');",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multilineString", Line: 1, Column: 4},
					{MessageId: "multilineString", Line: 2, Column: 7},
				},
			},
		},
	)
}

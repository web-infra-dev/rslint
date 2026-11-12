// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-string-slice.js
package prefer_string_slice_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_string_slice"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferStringSliceUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t,
		&prefer_string_slice.PreferStringSliceRule,
		[]rule_tester.ValidTestCase{
			// Basic cases.
			{Code: "const substr = foo.substr", FileName: "file.js"},
			{Code: "const substring = foo.substring", FileName: "file.js"},
			{Code: "foo.slice()", FileName: "file.js"},
			{Code: "foo.slice(0)", FileName: "file.js"},
			{Code: "foo.slice(1, 2)", FileName: "file.js"},
			{Code: "foo?.slice(1, 2)", FileName: "file.js"},
			{Code: "foo?.slice?.(1, 2)", FileName: "file.js"},
			{Code: "foo?.bar.baz.slice(1, 2)", FileName: "file.js"},
			{Code: "foo.slice(-3, -2)", FileName: "file.js"},
			{Code: "foo.substring(0, 1)", FileName: "file.js"},
			{Code: "foo.substring(1, 2)", FileName: "file.js"},
			{Code: "foo.substring(2, 1)", FileName: "file.js"},
			{Code: "\"foo\".substring(1, 2)", FileName: "file.js"},
			{Code: "\"foo\".substring(2, 1)", FileName: "file.js"},
			{Code: "foo.substring(index, index + 1)", FileName: "file.js"},
			{Code: "foo.substring(index, 1 + index)", FileName: "file.js"},
			{Code: "foo.substring(index - 1, index)", FileName: "file.js"},
			{Code: "foo.substring(index + 1, index)", FileName: "file.js"},
			{Code: "foo.substring(1 + index, index)", FileName: "file.js"},
			{Code: "foo.substring(index, /* comment */ index + 1)", FileName: "file.js"},
			{Code: "foo.substring(index, index + /* comment */ 1)", FileName: "file.js"},
			{Code: "foo?.substring(index, index + 1)", FileName: "file.js"},
			{Code: "foo.substring?.(index, index + 1)", FileName: "file.js"},
			{Code: "foo.substring((( index )), (( index )) + 1)", FileName: "file.js"},
			{Code: "foo.substring((( index )) - 1, (( index )))", FileName: "file.js"},
			{Code: "function foo(collection: {substring(start: number): string}) { collection.substring(1); }", FileName: "file.ts"},
		}, []rule_tester.InvalidTestCase{
			// Basic cases.
			{Code: "const modes = new Set([\"foo\"]); modes.clear(); foo.substr(modes.size ? 0 : 1, length);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 48, EndLine: 1, EndColumn: 86},
			}},
			{Code: "const modes = new Set([\"foo\"]); modes.clear(); foo.substr((modes.size ? 0 : 1) as number, length);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 48, EndLine: 1, EndColumn: 98},
			}},
			{Code: "const modes = new Set([\"foo\"]); modes.clear(); foo.substr(+(modes.size ? 0 : 1), length);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 48, EndLine: 1, EndColumn: 89},
			}},
			{Code: "const modes = new Set([\"foo\"]); modes.clear(); foo.substr((modes.size && 0) || index, length);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 48, EndLine: 1, EndColumn: 94},
			}},
			{Code: "const object = {value: true}; Object.defineProperty(object, \"value\", {get() { return 0; }}); foo.substr(object.value ? 0 : index, length);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 94, EndLine: 1, EndColumn: 138},
			}},
			{Code: "foo.substr()", FileName: "file.js", Output: []string{"foo.slice()"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
			}},
			{Code: "foo?.substr()", FileName: "file.js", Output: []string{"foo?.slice()"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
			}},
			{Code: "foo.bar?.substring()", FileName: "file.js", Output: []string{"foo.bar?.slice()"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
			}},
			{Code: "foo?.[0]?.substring()", FileName: "file.js", Output: []string{"foo?.[0]?.slice()"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
			}},
			{Code: "foo.bar.substr?.()", FileName: "file.js", Output: []string{"foo.bar.slice?.()"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
			}},
			{Code: "foo.bar?.substring?.()", FileName: "file.js", Output: []string{"foo.bar?.slice?.()"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 23},
			}},
			{Code: "foo.bar?.baz?.substr()", FileName: "file.js", Output: []string{"foo.bar?.baz?.slice()"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 23},
			}},
			{Code: "foo.bar?.baz.substring()", FileName: "file.js", Output: []string{"foo.bar?.baz.slice()"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
			}},
			{Code: "foo.bar.baz?.substr()", FileName: "file.js", Output: []string{"foo.bar.baz?.slice()"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
			}},
			{Code: "\"foo\".substr()", FileName: "file.js", Output: []string{"\"foo\".slice()"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
			}},
			{Code: "\"foo\".substr(bar.length, Math.min(baz, 100))", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 45},
			}},
			{Code: "\"foo\".substr(1, \"abc\".length)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 30},
			}},
			{Code: "\"foo\".substr(\"1\", 2)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
			}},
			{Code: "const length = 123;\n\"foo\".substr(1, length)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 2, Column: 1, EndLine: 2, EndColumn: 24},
			}},
			{Code: "const length = 123;\n\"foo\".substr(0, length)", FileName: "file.js", Output: []string{"const length = 123;\n\"foo\".slice(0, Math.max(0, length === undefined ? Infinity : length))"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 2, Column: 1, EndLine: 2, EndColumn: 24},
			}},
			{Code: "const length = 123;\n\"foo\".substr('0', length)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 2, Column: 1, EndLine: 2, EndColumn: 26},
			}},
			{Code: "\"foo\".substr(0, -1)", FileName: "file.js", Output: []string{"\"foo\".slice(0, 0)"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
			}},
			{Code: "\"foo\".substr(0, \"foo\".length)", FileName: "file.js", Output: []string{"\"foo\".slice(0, \"foo\".length)"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 30},
			}},
			{Code: "const length = 123;\n\"foo\".substr(1, length - 4)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 2, Column: 1, EndLine: 2, EndColumn: 28},
			}},
			{Code: "\"foo\".substr(1, length)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
			}},
			{Code: "const uri = 'foo';\n((uri || '')).substr(1)", FileName: "file.js", Output: []string{"const uri = 'foo';\n((uri || '')).slice(1)"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 2, Column: 1, EndLine: 2, EndColumn: 24},
			}},
			{Code: "foo.substr(start)", FileName: "file.js", Output: []string{"foo.slice(start)"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
			}},
			{Code: "\"foo\".substr(1)", FileName: "file.js", Output: []string{"\"foo\".slice(1)"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
			}},
			{Code: "foo.substr(start, length)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
			}},
			{Code: "\"foo\".substr(1, 2)", FileName: "file.js", Output: []string{"\"foo\".slice(1, 3)"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
			}},
			{Code: "foo.substr(1, 0)", FileName: "file.js", Output: []string{"foo.slice(1, 1)"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 17},
			}},
			{Code: "foo.substr(1, 2, 3)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
			}},
			{Code: "\"Sample\".substr(0, \"Sample\".lastIndexOf(\"/\"))", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 46},
			}},
			{Code: "foo.substr(0, foo.length - 2)", FileName: "file.js", Output: []string{"foo.slice(0, -2)"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 30},
			}},
			{Code: "foo?.substr(0, foo.length - 2)", FileName: "file.js", Output: []string{"foo?.slice(0, -2)"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
			}},
			{Code: "foo.substr(0, (foo.length - 2))", FileName: "file.js", Output: []string{"foo.slice(0, -2)"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
			}},
			{Code: "\"foo\".substr(0, \"foo\".length - 2)", FileName: "file.js", Output: []string{"\"foo\".slice(0, -2)"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 34},
			}},
			// Re-reading a getter or coercion is not safe; upstream still offers a fix.
			{Code: "foo.bar.substr(0, foo.bar.length - 2)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 38},
			}},
			// Re-reading a getter or coercion is not safe; upstream still offers a fix.
			{Code: "foo.substr(0, foo.length - count)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 34},
			}},
			// Re-reading a getter or coercion is not safe; upstream still offers a fix.
			{Code: "foo.substr(0, foo.length - 0)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 30},
			}},
			// Re-reading a getter or coercion is not safe; upstream still offers a fix.
			{Code: "foo.substr(0, foo.length - 2.5)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
			}},
			// Re-reading a getter or coercion is not safe; upstream still offers a fix.
			{Code: "foo.substr(0, foo.length - -1)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
			}},
			// Re-reading a getter or coercion is not safe; upstream still offers a fix.
			{Code: "foo.substr(0, bar.length - 2)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 30},
			}},
			// Re-reading a getter or coercion is not safe; upstream still offers a fix.
			{Code: "foo.substr(0, foo.length - /* comment */ 2)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 44},
			}},
			// Re-reading a getter or coercion is not safe; upstream still offers a fix.
			{Code: "foo.substr(0, /* comment */ foo.length - 2)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 44},
			}},
			// Re-reading a getter or coercion is not safe; upstream still offers a fix.
			{Code: "foo.substr(0, (/* comment */ foo.length - 2))", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 46},
			}},
			// Re-reading a getter or coercion is not safe; upstream still offers a fix.
			{Code: "foo.substr(0, (/* comment */ (foo.length - 2)))", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 48},
			}},
			// Re-reading a getter or coercion is not safe; upstream still offers a fix.
			{Code: "foo.substr(0, /* comment */ (foo.length - 2))", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 46},
			}},
			{Code: "foo.substring()", FileName: "file.js", Output: []string{"foo.slice()"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
			}},
			{Code: "\"foo\".substring()", FileName: "file.js", Output: []string{"\"foo\".slice()"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
			}},
			{Code: "\"foo\".substring(1)", FileName: "file.js", Output: []string{"\"foo\".slice(1)"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
			}},
			{Code: "\"foo\".substring(2, 4)", FileName: "file.js", Output: []string{"\"foo\".slice(2, 4)"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
			}},
			{Code: "\"foo\".substring(4, 2)", FileName: "file.js", Output: []string{"\"foo\".slice(2, 4)"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
			}},
			{Code: "\"foo\".substring(-1, -5)", FileName: "file.js", Output: []string{"\"foo\".slice(0, 0)"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
			}},
			{Code: "\"foo\".substring(-1, 2)", FileName: "file.js", Output: []string{"\"foo\".slice(0, 2)"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 23},
			}},
			{Code: "\"foo\".substring(length)", FileName: "file.js", Output: []string{"\"foo\".slice(Math.max(0, length))"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
			}},
			{Code: "\"foo\".substring(\"fo\".length)", FileName: "file.js", Output: []string{"\"foo\".slice(\"fo\".length)"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 29},
			}},
			{Code: "\"foo\".substring(0, length)", FileName: "file.js", Output: []string{"\"foo\".slice(0, Math.max(0, length === undefined ? Infinity : length))"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 27},
			}},
			{Code: "\"foo\".substring(length, 0)", FileName: "file.js", Output: []string{"\"foo\".slice(0, Math.max(0, length))"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 27},
			}},
			{Code: "foo.substring(start)", FileName: "file.js", Output: []string{"foo.slice(Math.max(0, start))"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
			}},
			{Code: "foo.substring(start, end)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
			}},
			{Code: "\"foo\".substring(1, 3)", FileName: "file.js", Output: []string{"\"foo\".slice(1, 3)"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
			}},
			{Code: "foo.substring(1, 2, 3)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 23},
			}},
			// TypeScript assertions.
			{Code: "function foo() {\n\treturn (bar as string).substr(3);\n}", FileName: "file.ts", Output: []string{"function foo() {\n\treturn (bar as string).slice(3);\n}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 2, Column: 9, EndLine: 2, EndColumn: 34},
			}},
			{Code: "function foo() {\n\treturn ((bar as string)).substring(3);\n}", FileName: "file.ts", Output: []string{"function foo() {\n\treturn ((bar as string)).slice(3);\n}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 2, Column: 9, EndLine: 2, EndColumn: 39},
			}},
			{Code: "const value: 'foo' = 'foo';\nvalue.substring(1);", FileName: "file.ts", Output: []string{"const value: 'foo' = 'foo';\nvalue.slice(1);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 2, Column: 1, EndLine: 2, EndColumn: 19},
			}},
			// Snapshot cases.
			{Code: "/* 1 */ (( /* 2 */ 0 /* 3 */, /* 4 */ foo /* 5 */ )) /* 6 */\n\t. /* 7 */ substring /* 8 */ (\n\t\t/* 9 */ (( /* 10 */ bar /* 11 */ )) /* 12 */,\n\t\t/* 13 */ (( /* 14 */ 0 /* 15 */ )) /* 16 */,\n\t\t/* 17 */\n\t)\n/* 18 */", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 9, EndLine: 6, EndColumn: 3},
			}},
			{Code: "foo.substr(0, ...bar)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
			}},
			{Code: "foo.substr(...bar)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
			}},
			{Code: "foo.substr(0, (100, 1))", FileName: "file.js", Output: []string{"foo.slice(0, Math.max(0, (100, 1) === undefined ? Infinity : (100, 1)))"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
			}},
			{Code: "foo.substr(0, 1, extraArgument)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
			}},
			{Code: "foo.substr((0, bar.length), (0, baz.length))", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 45},
			}},
			{Code: "foo.substring((10, 1), 0)", FileName: "file.js", Output: []string{"foo.slice(0, Math.max(0, (10, 1)))"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
			}},
			{Code: "foo.substring(0, (10, 1))", FileName: "file.js", Output: []string{"foo.slice(0, Math.max(0, (10, 1) === undefined ? Infinity : (10, 1)))"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
			}},
			{Code: "foo.substring(0, await 1)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
			}},
			{Code: "foo.substring((10, bar))", FileName: "file.js", Output: []string{"foo.slice(Math.max(0, (10, bar)))"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
			}},
			{Code: "const string = \"::\";\nconst output = string.substr(-2, 2);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 2, Column: 16, EndLine: 2, EndColumn: 36},
			}},
			{Code: "foo.substr(0.5, 0.5)", FileName: "file.js", Output: []string{"foo.slice(0.5, 0)"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
			}},
			{Code: "foo.substr(1.5, 2.5)", FileName: "file.js", Output: []string{"foo.slice(1.5, 3)"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
			}},
			{Code: "foo.substr(-1.5, 2.5)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
			}},
			{Code: "foo.substring(1, -n)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
			}},
			{Code: "foo.substring(1, -Infinity)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
			}},
			{Code: "foo.substring(1, -foo.length)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 30},
			}},
			{Code: "foo.substring((/* keep */ 5), 2)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 33},
			}},
			{Code: "foo.substring((/* keep */ -1))", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
			}},
			{Code: "foo.substring((/* keep */ -1), 3)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 34},
			}},
			{Code: "foo.substr(1, (/* keep */ 2))", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 30},
			}},
			{Code: "foo.substr(0, (/* keep */ -1))", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
			}},
			{Code: "foo.substring(0, (/* keep */ 2))", FileName: "file.js", Output: []string{"foo.slice(0, (/* keep */ 2))"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 33},
			}},
			{Code: "foo.substr(0, undefined)", FileName: "file.js", Output: []string{"foo.slice(0, Math.max(0, undefined === undefined ? Infinity : undefined))"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
			}},
			{Code: "foo.substring(0, undefined)", FileName: "file.js", Output: []string{"foo.slice(0, Math.max(0, undefined === undefined ? Infinity : undefined))"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
			}},
			{Code: "foo.substring(0, null)", FileName: "file.js", Output: []string{"foo.slice(0, Math.max(0, null === undefined ? Infinity : null))"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 23},
			}},
			{Code: "foo.substring(0, NaN)", FileName: "file.js", Output: []string{"foo.slice(0, Math.max(0, NaN === undefined ? Infinity : NaN))"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
			}},
			// Operator precedence and side effects.
			{Code: "foo.substring(0, a || b)", FileName: "file.js", Output: []string{"foo.slice(0, Math.max(0, (a || b) === undefined ? Infinity : (a || b)))"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
			}},
			{Code: "foo.substr(0, a && b)", FileName: "file.js", Output: []string{"foo.slice(0, Math.max(0, (a && b) === undefined ? Infinity : (a && b)))"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
			}},
			{Code: "foo.substring(0, () => a || b)", FileName: "file.js", Output: []string{"foo.slice(0, Math.max(0, (() => a || b) === undefined ? Infinity : (() => a || b)))"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
			}},
			{Code: "foo.substr(0, () => {})", FileName: "file.js", Output: []string{"foo.slice(0, Math.max(0, (() => {}) === undefined ? Infinity : (() => {})))"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
			}},
			{Code: "foo.substring(0, a ? b : c)", FileName: "file.js", Output: []string{"foo.slice(0, Math.max(0, (a ? b : c) === undefined ? Infinity : (a ? b : c)))"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
			}},
			{Code: "foo.substring(0, a = b)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
			}},
			{Code: "async function f() { foo.substring(0, await a || b); }", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 52},
			}},
			{Code: "function * f() { foo.substring(0, yield a || b); }", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 18, EndLine: 1, EndColumn: 48},
			}},
			{Code: "foo.substring(0, (a || b))", FileName: "file.js", Output: []string{"foo.slice(0, Math.max(0, (a || b) === undefined ? Infinity : (a || b)))"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 27},
			}},
			{Code: "foo.substring(0, a ?? b)", FileName: "file.js", Output: []string{"foo.slice(0, Math.max(0, (a ?? b) === undefined ? Infinity : (a ?? b)))"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
			}},
			// Re-reading a getter or coercion is not safe; upstream still offers a fix.
			{Code: "foo.substring(0, a + b)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
			}},
			// Re-reading a getter or coercion is not safe; upstream still offers a fix.
			{Code: "foo.substring(0, a.b)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
			}},
			// Re-reading a getter or coercion is not safe; upstream still offers a fix.
			{Code: "foo.substring(0, -a)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
			}},
			// Swapped arguments.
			{Code: "foo.substring(a, 0)", FileName: "file.js", Output: []string{"foo.slice(0, Math.max(0, a))"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
			}},
			{Code: "foo.substring(a || b, 0)", FileName: "file.js", Output: []string{"foo.slice(0, Math.max(0, a || b))"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
			}},
			{Code: "foo.substring(a, 0.0)", FileName: "file.js", Output: []string{"foo.slice(0, Math.max(0, a))"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
			}},
			{Code: "foo.substring(bar(), 0)", FileName: "file.js", Output: []string{"foo.slice(0, Math.max(0, bar()))"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
			}},
			{Code: "foo.substring((a, b), 0)", FileName: "file.js", Output: []string{"foo.slice(0, Math.max(0, (a, b)))"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
			}},
			{Code: "foo.substring((a /* comment */), 0)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 36},
			}},
			{Code: "foo.substring(0, a)", FileName: "file.js", Output: []string{"foo.slice(0, Math.max(0, a === undefined ? Infinity : a))"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
			}},
			// a `null` end is not rewritten to the whole string.
			{Code: "foo.substring(0, null);", FileName: "file.js", Output: []string{"foo.slice(0, Math.max(0, null === undefined ? Infinity : null));"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 23},
			}},
			{Code: "foo.substr(0, null);", FileName: "file.js", Output: []string{"foo.slice(0, Math.max(0, null === undefined ? Infinity : null));"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
			}},
			{Code: "const n = null; foo.substring(0, n);", FileName: "file.js", Output: []string{"const n = null; foo.slice(0, Math.max(0, n === undefined ? Infinity : n));"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 17, EndLine: 1, EndColumn: 36},
			}},
			{Code: "foo.substring(0, undefined);", FileName: "file.js", Output: []string{"foo.slice(0, Math.max(0, undefined === undefined ? Infinity : undefined));"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
			}},
			{Code: "foo.substring(0, n);", FileName: "file.js", Output: []string{"foo.slice(0, Math.max(0, n === undefined ? Infinity : n));"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
			}},
			// an end with a side effect is not rewritten.
			{Code: "foo.substring(0, getEnd());", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 27},
			}},
			{Code: "foo.substring(0, end());", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
			}},
			{Code: "foo.substring(0, a++);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
			}},
			{Code: "foo.substr(0, getEnd());", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
			}},
			{Code: "foo.substr(0, tag`end`);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
			}},
			{Code: "foo.substring(0, tag`end`);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 27},
			}},
			{Code: "foo.substr(0, [...ends][0]);", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
			}},
			// Static template receivers.
			{Code: "`abc`.substr(0, `abc`.length - 1);", FileName: "file.js", Output: []string{"`abc`.slice(0, -1);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 34},
			}},
			{Code: "`abc`.substr(0, \"abc\".length - 1);", FileName: "file.js", Output: []string{"`abc`.slice(0, -1);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 34},
			}},
			{Code: "`ab\\u{63}`.substr(0, `abc`.length - 1);", FileName: "file.js", Output: []string{"`ab\\u{63}`.slice(0, -1);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 39},
			}},
			// Documentation examples.
			{Code: "const str = 'hello world';\n\n// ❌\nstr.substr(0, 5);\n\n// ✅\nstr.slice(0, 5);\n// → 'hello'\n", FileName: "file.js", Output: []string{"const str = 'hello world';\n\n// ❌\nstr.slice(0, 5);\n\n// ✅\nstr.slice(0, 5);\n// → 'hello'\n"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substr", Message: "Prefer `String#slice()` over `String#substr()`.", Line: 4, Column: 1, EndLine: 4, EndColumn: 17},
			}},
			{Code: "const str = 'hello world';\n\n// ❌\nstr.substring(6, 11);\n\n// ✅\nstr.slice(6, 11);\n// → 'world'\n", FileName: "file.js", Output: []string{"const str = 'hello world';\n\n// ❌\nstr.slice(6, 11);\n\n// ✅\nstr.slice(6, 11);\n// → 'world'\n"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 4, Column: 1, EndLine: 4, EndColumn: 21},
			}},
			{Code: "const str = 'hello world';\n\n// ❌ substring swaps arguments when start > end\nstr.substring(5, 0);\n// → 'hello' (confusing!)\n\n// ✅ slice handles it intuitively\nstr.slice(5, 0);\n// → '' (empty, as expected)\n\n// ✅ slice handles negative indices\nstr.slice(-5);\n// → 'world' (last 5 characters)\n", FileName: "file.js", Output: []string{"const str = 'hello world';\n\n// ❌ substring swaps arguments when start > end\nstr.slice(0, 5);\n// → 'hello' (confusing!)\n\n// ✅ slice handles it intuitively\nstr.slice(5, 0);\n// → '' (empty, as expected)\n\n// ✅ slice handles negative indices\nstr.slice(-5);\n// → 'world' (last 5 characters)\n"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "substring", Message: "Prefer `String#slice()` over `String#substring()`.", Line: 4, Column: 1, EndLine: 4, EndColumn: 20},
			}},
		})
}

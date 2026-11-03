// Additional AST and edit-boundary coverage compared with Unicorn v79.0.0.
package no_double_comparison_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_double_comparison"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDoubleComparisonExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_double_comparison.NoDoubleComparisonRule, []rule_tester.ValidTestCase{
		// TypeScript wrappers, comment boundaries and JavaScript text values
		{Code: "x === \"\\uD800\" || x < \"�\"", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const re = /a/ as RegExp; re === x || re < x", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const re = /a/ satisfies RegExp; re === x || re < x", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const a = Object.freeze(/a/); const b = a; b === x || b < x", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},

		// Matching and reversed comparisons
		{Code: "ready || x < y", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "(a + b) === y || (a + b) < y", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "x === -1 || x < -1", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "x === `ab` || x < `cd`", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Member references and nested optional chains
		{Code: "a[key] === y || a[other] < y", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class C { #value; f(y) { return this.#value === y || this[\"#value\"] < y; } }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "a[key?.name] === y || a[key?.name] < y", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "(x?.a).b === y || (x?.a).b < y", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "x === y?.z || x < y?.z", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "x[fn?.()] === y || x[fn?.()] < y", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Static regular expressions and binding boundaries
		{Code: "const re = /a/; const alias = re; alias === x || alias < x", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const re = /a/; re <= x && re >= x", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const re = true ? /a/ : /b/; re === x || re < x", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
	}, []rule_tester.InvalidTestCase{
		// TypeScript wrappers, comment boundaries and JavaScript text values
		{Code: "(x === y)<boolean> || x < y", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `x <= y`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "x <= y"}}},
		}},
		{Code: "x === y || (x < y)<boolean>", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `x <= y`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "x <= y"}}},
		}},
		{Code: "(x === y)< /* type */ boolean> || x < y", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `x <= y`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 40, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "x === y || ((x < y) as boolean)<boolean>", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `x <= y`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 41, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "x <= y"}}},
		}},
		{Code: "(/* before left */ x === y) || x < y", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `x <= y`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 37, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "x === /* before right */ y || x < y", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `x <= y`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "x === (y) /* outside */ || x < y", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `x <= (y)`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 33, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(x /* left */) === (y /* right */) || x < y", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `(x /* left */) <= (y /* right */)`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 44, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "(x /* left */) <= (y /* right */)"}}},
		}},
		{Code: "\"😀\";\u2028x === y || x < y;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `x <= y`.", Line: 2, Column: 1, EndLine: 2, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "\"😀\";\u2028x <= y;"}}},
		}},
		{Code: "x ===\r\n y || x < y;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `x <= y`.", Line: 1, Column: 1, EndLine: 2, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "x <= y;"}}},
		}},
		{Code: "x === \"/* text */\" || x < \"/* text */\"", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `x <= \"/* text */\"`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 39, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "x <= \"/* text */\""}}},
		}},
		{Code: "x === \"\\uD800\" || x < `\\uD800`", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `x <= \"\\uD800\"`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "x <= \"\\uD800\""}}},
		}},
		{Code: "\\u0078 === y || x < y", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `\\u0078 <= y`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "\\u0078 <= y"}}},
		}},
		{Code: "const a = [/a/][0]; a === x || a < x", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `a <= x`.", Line: 1, Column: 21, EndLine: 1, EndColumn: 37, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "const a = [/a/][0]; a <= x"}}},
		}},
		{Code: "const Object = { freeze: x => 0 }; const a = Object.freeze(/a/); a === x || a < x", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `a <= x`.", Line: 1, Column: 66, EndLine: 1, EndColumn: 82, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "const Object = { freeze: x => 0 }; const a = Object.freeze(/a/); a <= x"}}},
		}},

		// Matching and reversed comparisons
		{Code: "x >= y && y !== x", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `x > y`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "x > y"}}}}},
		{Code: "x !== y && y >= x", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `x < y`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "x < y"}}}}},
		{Code: "x >= y && y >= x", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `x === y`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "x === y"}}}}},
		{Code: "x < y || y < x", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `x !== y`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "x !== y"}}}}},
		{Code: "x === 0x10 || x < 16", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `x <= 0x10`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "x <= 0x10"}}}}},
		{Code: "x === 1n || x < 0x1n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `x <= 1n`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "x <= 1n"}}}}},
		{Code: "x === `ab` || x < \"a\\u0062\"", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `x <= `ab``.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "x <= `ab`"}}}}},
		// Member references and nested optional chains
		{Code: "a[\"b\" + \"c\"] === y || a.bc < y", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `a[\"b\" + \"c\"] <= y`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "a[\"b\" + \"c\"] <= y"}}}}},
		{Code: "class C { #value; f(y) { return this.#value === y || this.#value < y; } }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `this.#value <= y`.", Line: 1, Column: 33, EndLine: 1, EndColumn: 69, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "class C { #value; f(y) { return this.#value <= y; } }"}}}}},
		{Code: "class C extends B { f(y) { return super.x === y || super.x < y; } }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `super.x <= y`.", Line: 1, Column: 35, EndLine: 1, EndColumn: 63, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "class C extends B { f(y) { return super.x <= y; } }"}}}}},
		{Code: "this === y || this < y", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `this <= y`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "this <= y"}}}}},
		// Static regular expressions and binding boundaries
		{Code: "let re = /a/; re === x || re < x", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `re <= x`.", Line: 1, Column: 15, EndLine: 1, EndColumn: 33, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "let re = /a/; re <= x"}}}}},
		{Code: "const re = /a/; function f(re) { return re === x || re < x; }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `re <= x`.", Line: 1, Column: 41, EndLine: 1, EndColumn: 59, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "const re = /a/; function f(re) { return re <= x; }"}}}}},
		{Code: "re === x || re < x; const re = /a/", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `re <= x`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "re <= x; const re = /a/"}}}}},
		{Code: "const re = new RegExp('a'); re === x || re < x", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `re <= x`.", Line: 1, Column: 29, EndLine: 1, EndColumn: 47, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "const re = new RegExp('a'); re <= x"}}}}},
		{Code: "const re = ({value: /a/}).value; re === x || re < x", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `re <= x`.", Line: 1, Column: 34, EndLine: 1, EndColumn: 52, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "const re = ({value: /a/}).value; re <= x"}}}}},
		{Code: "let re = /a/; re = 1; re === x || re < x", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `re <= x`.", Line: 1, Column: 23, EndLine: 1, EndColumn: 41, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "let re = /a/; re = 1; re <= x"}}}}},
		// Parentheses, comments and UTF-16 ranges
		{Code: "((x)) === ((y)) || x < y", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `((x)) <= ((y))`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "((x)) <= ((y))"}}}}},
		{Code: "x /* dropped */ === y || x < y", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `x <= y`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}}}},
		{Code: "x === (y /* kept */) || x < y", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `x <= (y /* kept */)`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "x <= (y /* kept */)"}}}}},
		{Code: "x === y || x < (y /* dropped */)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `x <= y`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 33, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}}}},
		{Code: "// before\n((x === y || x < y)) // after", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `x <= y`.", Line: 2, Column: 3, EndLine: 2, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "// before\n((x <= y)) // after"}}}}},
		{Code: "\"😀\";\n(变量) ===\n  值 || 变量 < 值;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `(变量) <= 值`.", Line: 2, Column: 1, EndLine: 3, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "\"😀\";\n(变量) <= 值;"}}}}},
		{Code: "\"😀\"; x === y || x < y;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `x <= y`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "\"😀\"; x <= y;"}}}}},
		{Code: "(x // kept\n) === y || x < y", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `(x // kept\n) <= y`.", Line: 1, Column: 1, EndLine: 2, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "(x // kept\n) <= y"}}}}},
		{Code: "x === y || // dropped\n x < y", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `x <= y`.", Line: 1, Column: 1, EndLine: 2, EndColumn: 7, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}}}},
		// TypeScript and JSX representation
		{Code: "((x === y) as boolean) || ((x < y) satisfies boolean)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `x <= y`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 54, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "x <= y"}}}}},
		{Code: "(<number>x) === y || (x as number) < y", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `(<number>x) <= y`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 39, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "(<number>x) <= y"}}}}},
		{Code: "x === y! || x < y!", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `x <= y!`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "x <= y!"}}}}},
		{Code: "(a as Box)[\"b\" as string] === y || a.b < y", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `(a as Box)[\"b\" as string] <= y`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "(a as Box)[\"b\" as string] <= y"}}}}},
		{Code: "(fn<number>) === y || (fn<string>) < y", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `(fn<number>) <= y`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 39, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "(fn<number>) <= y"}}}}},
		{Code: "(fn<number>).value === y || fn.value < y", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `(fn<number>).value <= y`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 41, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "(fn<number>).value <= y"}}}}},
		{Code: "a[key<string>] === y || a[key] < y", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `a[key<string>] <= y`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "a[key<string>] <= y"}}}}},
		{Code: "x === y || x < (y as number)", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `x <= y`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "x <= y"}}}}},
		{Code: "const view = <div>{x === y || x < y}</div>;", FileName: "case.tsx", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `x <= y`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "const view = <div>{x <= y}</div>;"}}}}},
		{Code: "/** @type {number} */ (x) === y || x < y", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `(x) <= y`.", Line: 1, Column: 23, EndLine: 1, EndColumn: 41, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "/** @type {number} */ (x) <= y"}}}}},
		{Code: "(x === y || x < y) as boolean", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-double-comparison", Message: "These comparisons can be simplified to `x <= y`.", Line: 1, Column: 2, EndLine: 1, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-double-comparison/suggestion", Output: "(x <= y) as boolean"}}}}},
	})
}

func TestNoDoubleComparisonSuggestionsFollowDemand(t *testing.T) {
	for _, test := range []struct {
		code, output, replacement string
	}{
		{"x === y || x < y", "x <= y", "x <= y"},
		{"x === y || /* keep */ x < y", "", "x <= y"},
		{"x === (y as number) || x < y", "x <= (y as number)", "x <= (y as number)"},
	} {
		t.Run(test.code, func(t *testing.T) {
			helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
			program, sourceFile, err := helper.CreateTestProgram(test.code, "edit-demand.ts", "tsconfig.json")
			if err != nil {
				t.Fatal(err)
			}
			diagnostics := make(map[rule.EditDemand]rule.RuleDiagnostic, 4)
			for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
				var got []rule.RuleDiagnostic
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program: lintprogram.NewFromCompiler(program), File: sourceFile.FileName(),
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{Name: no_double_comparison.NoDoubleComparisonRule.Name, Severity: rule.SeverityError,
							Run: func(ctx rule.RuleContext) rule.RuleListeners {
								return no_double_comparison.NoDoubleComparisonRule.Run(ctx, nil)
							},
						}}
					},
					Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { got = append(got, d) }},
				})
				if len(got) != 1 {
					t.Fatalf("demand %d: got %d diagnostics", demand, len(got))
				}
				diagnostics[demand] = got[0]
			}
			baseline, all := diagnostics[rule.EditDemandNone], diagnostics[rule.EditDemandAll]
			for demand, diagnostic := range diagnostics {
				if diagnostic.Range != baseline.Range || !reflect.DeepEqual(diagnostic.Message, baseline.Message) || diagnostic.Severity != baseline.Severity {
					t.Errorf("demand %d changed diagnostic identity", demand)
				}
				if diagnostic.FixesPtr != nil {
					t.Errorf("demand %d produced an autofix", demand)
				}
				wantSuggestion := test.output != "" && (demand == rule.EditDemandSuggestion || demand == rule.EditDemandAll)
				if !wantSuggestion {
					if diagnostic.Suggestions != nil {
						t.Errorf("demand %d produced unexpected suggestions", demand)
					}
					continue
				}
				if diagnostic.Suggestions == nil || len(*diagnostic.Suggestions) != 1 || !reflect.DeepEqual(diagnostic.Suggestions, all.Suggestions) {
					t.Fatalf("demand %d: inconsistent suggestions", demand)
				}
				suggestion := (*diagnostic.Suggestions)[0]
				if suggestion.Message.Id != "no-double-comparison/suggestion" || suggestion.Message.Description != "Replace with `"+test.replacement+"`." {
					t.Errorf("demand %d: wrong suggestion message", demand)
				}
				output, _, fixed := linter.ApplyRuleFixes(test.code, (*diagnostic.Suggestions)[:1])
				if !fixed || output != test.output {
					t.Errorf("demand %d: unexpected suggestion %q", demand, output)
				}
			}
		})
	}
}

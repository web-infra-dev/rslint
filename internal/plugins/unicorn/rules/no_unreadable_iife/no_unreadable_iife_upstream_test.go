// Ported from eslint-plugin-unicorn v74.0.0; see LICENSE.
package no_unreadable_iife_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_unreadable_iife"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnreadableIifeUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_unreadable_iife.NoUnreadableIifeRule, []rule_tester.ValidTestCase{
		{Code: "const foo = (bar => bar)();", FileName: "case.js"},
		{Code: "const foo = (() => {\n\treturn a ? b : c\n})();", FileName: "case.js"},
		{Code: "const bar = getBar();\nconst foo = bar ? bar.baz : baz;", FileName: "case.js"},
		{Code: "const getBaz = bar => (bar ? bar.baz : baz);\nconst foo = getBaz(getBar());", FileName: "case.js"},
		{Code: "const foo = {bar, baz};", FileName: "case.js"},
		{Code: "const foo = (bar => {\n\treturn bar ? bar.baz : baz;\n})(getBar());", FileName: "case.js"},
		{Code: "const foo = (() => { return a ? b : c; })();", FileName: "suggested.js"},
		{Code: "const foo = (() => { return a ? b : c; })();", FileName: "suggested.js"},
		{Code: "const foo = (\n\t() => { return a ? b : c; }\n)();", FileName: "suggested.js"},
		{Code: "const foo = (() => { return a, b; })();", FileName: "suggested.js"},
		{Code: "const foo = (() => { return {\n\ta: b,\n}; })();", FileName: "suggested.js"},
		{Code: "const foo = (bar => { return bar; })();", FileName: "suggested.js"},
		{Code: "(async () => { return {\n\tbar,\n}; })();", FileName: "suggested.js"},
		{Code: "const foo = (async (bar) => { return {\n\tbar: await baz(),\n}; })();", FileName: "suggested.js"},
		{Code: "(async () => { return {bar}; })();", FileName: "suggested.js"},
		{Code: "const foo = (bar => { return bar ? bar.baz : baz; })(getBar());", FileName: "suggested.js"},
		{Code: "const foo = ((bar, baz) => { return {bar, baz}; })(bar, baz);", FileName: "suggested.js"},
	}, []rule_tester.InvalidTestCase{
		{Code: "const foo = (() => (a ? b : c))();", FileName: "case.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unreadable-iife", Message: "IIFE with parenthesized arrow function body is considered unreadable.", Line: 1, Column: 20, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const foo = (() => { return a ? b : c; })();"}}},
		}},
		{Code: "const foo = (() => (\n\ta ? b : c\n))();", FileName: "case.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unreadable-iife", Message: "IIFE with parenthesized arrow function body is considered unreadable.", Line: 1, Column: 20, EndLine: 3, EndColumn: 2, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const foo = (() => { return a ? b : c; })();"}}},
		}},
		{Code: "const foo = (\n\t() => (\n\t\ta ? b : c\n\t)\n)();", FileName: "case.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unreadable-iife", Message: "IIFE with parenthesized arrow function body is considered unreadable.", Line: 2, Column: 8, EndLine: 4, EndColumn: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const foo = (\n\t() => { return a ? b : c; }\n)();"}}},
		}},
		{Code: "const foo = (() => (/* comment */ a ? b : c))();", FileName: "case.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unreadable-iife", Message: "IIFE with parenthesized arrow function body is considered unreadable.", Line: 1, Column: 20, EndLine: 1, EndColumn: 45, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = (() => (\n\ta, b\n))();", FileName: "case.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unreadable-iife", Message: "IIFE with parenthesized arrow function body is considered unreadable.", Line: 1, Column: 20, EndLine: 3, EndColumn: 2, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const foo = (() => { return a, b; })();"}}},
		}},
		{Code: "const foo = (() => ({\n\ta: b,\n}))();", FileName: "case.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unreadable-iife", Message: "IIFE with parenthesized arrow function body is considered unreadable.", Line: 1, Column: 20, EndLine: 3, EndColumn: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const foo = (() => { return {\n\ta: b,\n}; })();"}}},
		}},
		{Code: "const foo = (bar => (bar))();", FileName: "case.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unreadable-iife", Message: "IIFE with parenthesized arrow function body is considered unreadable.", Line: 1, Column: 21, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const foo = (bar => { return bar; })();"}}},
		}},
		{Code: "(async () => ({\n\tbar,\n}))();", FileName: "case.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unreadable-iife", Message: "IIFE with parenthesized arrow function body is considered unreadable.", Line: 1, Column: 14, EndLine: 3, EndColumn: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "(async () => { return {\n\tbar,\n}; })();"}}},
		}},
		{Code: "const foo = (async (bar) => ({\n\tbar: await baz(),\n}))();", FileName: "case.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unreadable-iife", Message: "IIFE with parenthesized arrow function body is considered unreadable.", Line: 1, Column: 29, EndLine: 3, EndColumn: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const foo = (async (bar) => { return {\n\tbar: await baz(),\n}; })();"}}},
		}},
		{Code: "(async () => (( {bar} )))();", FileName: "case.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unreadable-iife", Message: "IIFE with parenthesized arrow function body is considered unreadable.", Line: 1, Column: 14, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "(async () => { return {bar}; })();"}}},
		}},
		{Code: "const foo = (bar => (bar ? bar.baz : baz))(getBar());", FileName: "case.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unreadable-iife", Message: "IIFE with parenthesized arrow function body is considered unreadable.", Line: 1, Column: 21, EndLine: 1, EndColumn: 42, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const foo = (bar => { return bar ? bar.baz : baz; })(getBar());"}}},
		}},
		{Code: "const foo = ((bar, baz) => ({bar, baz}))(bar, baz);", FileName: "case.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unreadable-iife", Message: "IIFE with parenthesized arrow function body is considered unreadable.", Line: 1, Column: 28, EndLine: 1, EndColumn: 40, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const foo = ((bar, baz) => { return {bar, baz}; })(bar, baz);"}}},
		}},
	})
}

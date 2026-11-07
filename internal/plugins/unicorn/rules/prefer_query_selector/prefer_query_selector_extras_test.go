package prefer_query_selector_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_query_selector"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Expectations include the documented safety differences from v79.0.0, with
// cases covering AST wrappers, read/write accesses, comments, and whitespace.
func TestPreferQuerySelectorExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_query_selector.PreferQuerySelectorRule, []rule_tester.ValidTestCase{
		{Code: "document?.getElementById(\"foo\");", FileName: "extras.js"},
		{Code: "document.getElementById?.(\"foo\");", FileName: "extras.js"},
		{Code: "(document?.getElementById)(\"foo\");", FileName: "extras.js"},
		{Code: "document[\"getElementsByTagName\"](\"form\");", FileName: "extras.js"},
		{Code: "class Node { #getElementById() {} f() { this.#getElementById(\"id\"); } }", FileName: "extras.js"},
		{Code: "(document.getElementById as Function)(\"foo\");", FileName: "extras.ts"},
		{Code: "document.getElementsByClassName(1);", FileName: "extras.js", Options: map[string]any{"allowWithVariables": true}},
		{Code: "document.getElementById(true);", FileName: "extras.js", Options: map[string]any{"allowWithVariables": true}},
		{Code: "document.getElementById(/foo/);", FileName: "extras.js", Options: map[string]any{"allowWithVariables": true}},
		{Code: "document.getElementById(foo || \"bar\");", FileName: "extras.js", Options: map[string]any{"allowWithVariables": true}},
		{Code: "document.getElementById(foo ?? \"bar\");", FileName: "extras.js", Options: map[string]any{"allowWithVariables": true}},
		{Code: "document.getElementById((foo, \"bar\"));", FileName: "extras.js", Options: map[string]any{"allowWithVariables": true}},
		{Code: "document.getElementById(foo = \"bar\");", FileName: "extras.js", Options: map[string]any{"allowWithVariables": true}},
		{Code: "document.getElementById(condition ? \"a\" : \"b\");", FileName: "extras.js", Options: map[string]any{"allowWithVariables": true}},
		{Code: "document.getElementById(` \t${foo}\n`);", FileName: "extras.js", Options: map[string]any{"allowWithVariables": true}},
		{Code: "document.getElementById(value as string);", FileName: "extras.ts", Options: map[string]any{"allowWithVariables": true}},
		{Code: "document.getElementById(value!);", FileName: "extras.ts", Options: map[string]any{"allowWithVariables": true}},
		{Code: "document.getElementById(value satisfies string);", FileName: "extras.ts", Options: map[string]any{"allowWithVariables": true}},
		{Code: "document.getElementById(<string>value);", FileName: "extras.ts", Options: map[string]any{"allowWithVariables": true}},
		{Code: "document.getElementById(fn<string>);", FileName: "extras.ts", Options: map[string]any{"allowWithVariables": true}},
		{Code: "(undefined).getElementsByName(\"foo\");", FileName: "extras.js"},
		{Code: "function f(undefined) { undefined.getElementsByName(\"foo\"); }", FileName: "extras.js"},
		{Code: "(/** @type {Element} */ (\"value\")).getElementById(\"foo\");", FileName: "extras.js"},
	}, []rule_tester.InvalidTestCase{
		{Code: "(document.getElementById)((\"foo\"));", FileName: "extras.js", Output: []string{"(document.querySelector)((\"#foo\"));"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 25},
		}},
		{Code: "document.getElementById(/* id */ (\"foo\" /* keep */));", FileName: "extras.js", Output: []string{"document.querySelector(/* id */ (\"#foo\" /* keep */));"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 24},
		}},
		{Code: "document.getElementById(\"foo\" satisfies string);", FileName: "extras.ts", Output: []string{"document.querySelector(\"#foo\" satisfies string);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 24},
		}},
		{Code: "document.getElementById(<string>\"foo\");", FileName: "extras.ts", Output: []string{"document.querySelector(<string>\"#foo\");"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 24},
		}},
		{Code: "document.getElementById((\"foo\" as const)!);", FileName: "extras.ts", Output: []string{"document.querySelector((\"#foo\" as const)!);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 24},
		}},
		{Code: "document.getElementById<string>(\"foo\");", FileName: "extras.ts", Output: []string{"document.querySelector<string>(\"#foo\");"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 24},
		}},
		{Code: "(<div />).getElementsByName(\"foo\");", FileName: "extras.tsx", Output: []string{"(<div />).querySelectorAll(\"[name='foo']\");"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByName()`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 28},
		}},
		{Code: "(void node).getElementById(\"foo\");", FileName: "extras.js", Output: []string{"(void node).querySelector(\"#foo\");"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 13, EndLine: 1, EndColumn: 27},
		}},
		{Code: "([], []).getElementById(\"foo\");", FileName: "extras.js", Output: []string{"([], []).querySelector(\"#foo\");"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 24},
		}},
		{Code: "([] as any).getElementById(\"foo\");", FileName: "extras.ts", Output: []string{"([] as any).querySelector(\"#foo\");"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 13, EndLine: 1, EndColumn: 27},
		}},
		{Code: "(document).getElementsByTagName(\"form\");", FileName: "extras.js", Output: []string{"(document).querySelectorAll(\"form\");"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 12, EndLine: 1, EndColumn: 32},
		}},
		{Code: "(window).document.getElementsByName(\"foo\");", FileName: "extras.js", Output: []string{"(window).document.querySelectorAll(\"[name='foo']\");"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByName()`.", Line: 1, Column: 19, EndLine: 1, EndColumn: 36},
		}},
		{Code: "globalThis.document.getElementById(\"foo\");", FileName: "extras.js", Output: []string{"globalThis.document.querySelector(\"#foo\");"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 21, EndLine: 1, EndColumn: 35},
		}},
		{Code: "window[\"document\"].getElementsByTagName(\"*\");", FileName: "extras.js", Output: []string{"window[\"document\"].querySelectorAll(\"*\");"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 40},
		}},
		{Code: "self.document.getElementById(\"foo\");", FileName: "extras.js", Output: []string{"self.document.querySelector(\"#foo\");"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 15, EndLine: 1, EndColumn: 29},
		}},
		{Code: "window?.document.getElementById(\"foo\");", FileName: "extras.js", Output: []string{"window?.document.querySelector(\"#foo\");"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 18, EndLine: 1, EndColumn: 32},
		}},
		{Code: "(node?.child).getElementById(\"foo\");", FileName: "extras.js", Output: []string{"(node?.child).querySelector(\"#foo\");"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 15, EndLine: 1, EndColumn: 29},
		}},
		{Code: "function f(document) { document.getElementById(\"foo\"); }", FileName: "extras.js", Output: []string{"function f(document) { document.querySelector(\"#foo\"); }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 33, EndLine: 1, EndColumn: 47},
		}},
		{Code: "(document as Document).getElementById(\"foo\");", FileName: "extras.ts", Output: []string{"(document as Document).querySelector(\"#foo\");"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 24, EndLine: 1, EndColumn: 38},
		}},
		{Code: "(document.getElementsByTagName(\"form\"))[0];", FileName: "extras.js", Output: []string{"(document.querySelectorAll(\"form\"))[0];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 31},
		}},
		{Code: "((document.getElementsByClassName(\"foo\")).at)(0);", FileName: "extras.js", Output: []string{"((document.querySelectorAll(\".foo\")).at)(0);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 12, EndLine: 1, EndColumn: 34},
		}},
		{Code: "(document.getElementsByName(\"foo\")).item((0));", FileName: "extras.js", Output: []string{"(document.querySelector(\"[name='foo']\"));"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementsByName()`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 28},
		}},
		{Code: "document.getElementsByTagName(\"form\")[0x0];", FileName: "extras.js", Output: []string{"document.querySelectorAll(\"form\")[0x0];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: "document.getElementsByTagName(\"form\").at(0.0);", FileName: "extras.js", Output: []string{"document.querySelectorAll(\"form\").at(0.0);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: "document.getElementsByTagName(\"form\").item(0e0);", FileName: "extras.js", Output: []string{"document.querySelector(\"form\");"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: `document.getElementsByTagName("form")[1e-999];`, FileName: "extras.js", Output: []string{"document.querySelectorAll(\"form\")[1e-999];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: "document.getElementsByTagName(\"form\")[-0];", FileName: "extras.js", Output: []string{"document.querySelectorAll(\"form\")[-0];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: "document.getElementsByTagName(\"form\")[\"0\"];", FileName: "extras.js", Output: []string{"document.querySelectorAll(\"form\")[\"0\"];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: "document.getElementsByTagName(\"form\")[0n];", FileName: "extras.js", Output: []string{"document.querySelectorAll(\"form\")[0n];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: "document.getElementsByTagName(\"form\")[0 as number];", FileName: "extras.ts", Output: []string{"document.querySelectorAll(\"form\")[0 as number];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: "document.getElementsByTagName(\"form\")[1];", FileName: "extras.js", Output: []string{"document.querySelectorAll(\"form\")[1];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: "document.getElementsByTagName(\"form\")?.[0];", FileName: "extras.js", Output: []string{"document.querySelectorAll(\"form\")?.[0];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: "document.getElementsByTagName(\"form\")?.at(0);", FileName: "extras.js", Output: []string{"document.querySelectorAll(\"form\")?.at(0);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: "document.getElementsByTagName(\"form\").at?.(0);", FileName: "extras.js", Output: []string{"document.querySelectorAll(\"form\").at?.(0);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: "document.getElementsByTagName(\"form\")[0]?.submit();", FileName: "extras.js", Output: []string{"document.querySelectorAll(\"form\")[0]?.submit();"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: "document.getElementsByTagName(\"form\").at(0, other);", FileName: "extras.js", Output: []string{"document.querySelectorAll(\"form\").at(0, other);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: "document.getElementsByTagName(\"form\").item(...[0]);", FileName: "extras.js", Output: []string{"document.querySelectorAll(\"form\").item(...[0]);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: "document.getElementsByTagName(\"form\")[/* keep */ 0];", FileName: "extras.js", Output: []string{"document.querySelectorAll(\"form\")[/* keep */ 0];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: "document.getElementsByTagName(\"form\")./* keep */ at(0);", FileName: "extras.js", Output: []string{"document.querySelectorAll(\"form\")./* keep */ at(0);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: "((document.getElementsByTagName(\"form\") /* keep */))[0];", FileName: "extras.js", Output: []string{"((document.querySelectorAll(\"form\") /* keep */))[0];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 12, EndLine: 1, EndColumn: 32},
		}},
		{Code: "document.getElementsByTagName(/* keep */ \"form\")[0];", FileName: "extras.js", Output: []string{"document.querySelectorAll(/* keep */ \"form\")[0];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: "document.getElementsByTagName(\"form\")[0] = value;", FileName: "extras.js", Output: []string{"document.querySelectorAll(\"form\")[0] = value;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: "document.getElementsByTagName(\"form\")[0] ||= value;", FileName: "extras.js", Output: []string{"document.querySelectorAll(\"form\")[0] ||= value;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: "++document.getElementsByTagName(\"form\")[0];", FileName: "extras.js", Output: []string{"++document.querySelectorAll(\"form\")[0];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 12, EndLine: 1, EndColumn: 32},
		}},
		{Code: "document.getElementsByTagName(\"form\")[0]--;", FileName: "extras.js", Output: []string{"document.querySelectorAll(\"form\")[0]--;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: "[document.getElementsByTagName(\"form\")[0]] = values;", FileName: "extras.js", Output: []string{"[document.querySelectorAll(\"form\")[0]] = values;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 31},
		}},
		{Code: "({value: document.getElementsByTagName(\"form\")[0]} = object);", FileName: "extras.js", Output: []string{"({value: document.querySelectorAll(\"form\")[0]} = object);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 19, EndLine: 1, EndColumn: 39},
		}},
		{Code: `({[document.getElementsByTagName("form")[0]]: other} = object);`, FileName: "extras.js", Output: []string{"({[document.querySelectorAll(\"form\")[0]]: other} = object);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 13, EndLine: 1, EndColumn: 33},
		}},
		{Code: "[...document.getElementsByTagName(\"form\")[0]] = values;", FileName: "extras.js", Output: []string{"[...document.querySelectorAll(\"form\")[0]] = values;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 34},
		}},
		{Code: "delete (document.getElementsByTagName(\"form\")[0]);", FileName: "extras.js", Output: []string{"delete (document.querySelectorAll(\"form\")[0]);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 18, EndLine: 1, EndColumn: 38},
		}},
		{Code: "(document.getElementsByTagName(\"form\")[0] as Element) = value;", FileName: "extras.ts", Output: []string{"(document.querySelectorAll(\"form\")[0] as Element) = value;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 31},
		}},
		{Code: "(document.getElementsByTagName(\"form\")[0]!)++;", FileName: "extras.ts", Output: []string{"(document.querySelectorAll(\"form\")[0]!)++;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 31},
		}},
		{Code: "for ((document.getElementsByTagName(\"form\")[0] as Element) of values) {}", FileName: "extras.ts", Output: []string{"for ((document.querySelectorAll(\"form\")[0] as Element) of values) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 16, EndLine: 1, EndColumn: 36},
		}},
		{Code: "document.getElementById(\"-foo\");", FileName: "extras.js", Output: []string{"document.querySelector(\"#-foo\");"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 24},
		}},
		{Code: "document.getElementById(\"_foo-2\");", FileName: "extras.js", Output: []string{"document.querySelector(\"#_foo-2\");"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 24},
		}},
		{Code: "document.getElementById(\"--foo\");", FileName: "extras.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 24},
		}},
		{Code: "document.getElementById(\"123\");", FileName: "extras.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 24},
		}},
		{Code: "document.getElementById(\"é\");", FileName: "extras.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 24},
		}},
		{Code: "document.getElementById(\"\\u0061\");", FileName: "extras.js", Output: []string{"document.querySelector(\"#a\");"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 24},
		}},
		{Code: "element.getElementsByTagName(\"\\u0061\");", FileName: "extras.js", Output: []string{"element.querySelectorAll(\"\\u0061\");"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 29},
		}},
		{Code: "document.getElementsByTagName(\"\\u0061\");", FileName: "extras.js", Output: []string{"document.querySelectorAll(\"\\u0061\");"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: "element.getElementsByClassName(\"foo\uFEFFbar baz\");", FileName: "extras.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 31},
		}},
		{Code: "document.getElementsByClassName(\"foo\u0085bar\");", FileName: "extras.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 32},
		}},
		{Code: "document.getElementsByClassName(\"\uFEFF\t\\n\");", FileName: "extras.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 32},
		}},
		{Code: "document.getElementsByName(\"\uFEFF\");", FileName: "extras.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 27},
		}},
		{Code: "element.getElementsByTagName(`*`);", FileName: "extras.js", Output: []string{"element.querySelectorAll(`*`);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 29},
		}},
		{Code: "element.getElementsByName(`foo`);", FileName: "extras.js", Output: []string{"element.querySelectorAll(`[name='foo']`);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByName()`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 26},
		}},
		{Code: "element.getElementsByName(\"foo\");", FileName: "extras.js", Output: []string{"element.querySelectorAll(\"[name='foo']\");"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByName()`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 26},
		}},
		{Code: "element.getElementById(null);", FileName: "extras.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 23},
		}},
		{Code: "element.getElementsByName(null)[0];", FileName: "extras.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByName()`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 26},
		}},
		{Code: "document.getElementById(null);", FileName: "extras.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 24},
		}},
		{Code: "document.getElementsByName(\"f\\u006fo\");", FileName: "extras.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 27},
		}},
		{Code: "document.getElementsByName(`foo\nbar`);", FileName: "extras.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 27},
		}},
		{Code: "document.getElementsByName(`foo\u2028bar`);", FileName: "extras.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 27},
		}},
		{Code: "document.getElementsByName(\"foo`bar\");", FileName: "extras.js", Output: []string{"document.querySelectorAll(\"[name='foo`bar']\");"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 27},
		}},
		{Code: "document.getElementsByName(\"foo]bar\");", FileName: "extras.js", Output: []string{"document.querySelectorAll(\"[name='foo]bar']\");"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 27},
		}},
		{Code: "document.getElementsByName((\"foo\" as string)!);", FileName: "extras.ts", Output: []string{"document.querySelectorAll((\"[name='foo']\" as string)!);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 27},
		}},
		{Code: "document.getElementById(value);", FileName: "extras.js", Options: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 24},
		}},
		{Code: "document.getElementById(value);", FileName: "extras.js", Options: map[string]any{"allowWithVariables": false}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 24},
		}},
		{Code: "document.getElementById(a === b);", FileName: "extras.js", Options: map[string]any{"allowWithVariables": true}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 24},
		}},
		{Code: "document.getElementById(\"x\" + \"y\");", FileName: "extras.js", Options: map[string]any{"allowWithVariables": true}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 24},
		}},
		{Code: "document.getElementById(`\u0085${value}`);", FileName: "extras.js", Options: map[string]any{"allowWithVariables": true}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 24},
		}},
		{Code: "document.getElementById(`${value}tail`);", FileName: "extras.js", Options: map[string]any{"allowWithVariables": true}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 24},
		}},
		{Code: "document.getElementsByName(`${value}`);", FileName: "extras.js", Options: map[string]any{"allowWithVariables": true}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 27},
		}},
		{Code: "document.getElementsByTagName(value || fallback);", FileName: "extras.js", Options: map[string]any{"allowWithVariables": true}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: "\"😀\"; document.getElementById(\"foo\");", FileName: "extras.js", Output: []string{"\"😀\"; document.querySelector(\"#foo\");"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 16, EndLine: 1, EndColumn: 30},
		}},
		{Code: "document\n  .getElementsByName(\"foo\");", FileName: "extras.js", Output: []string{"document\n  .querySelectorAll(\"[name='foo']\");"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByName()`.", Line: 2, Column: 4, EndLine: 2, EndColumn: 21},
		}},
		// Parentheses ending an optional chain must retain the subsequent access.
		{Code: `(window?.document.getElementsByTagName("form"))[0];`, FileName: "extras.js", Output: []string{`(window?.document.querySelectorAll("form"))[0];`}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 19, EndLine: 1, EndColumn: 39},
		}},
		{Code: `(window?.document.getElementsByTagName("form")).item(0);`, FileName: "extras.js", Output: []string{`(window?.document.querySelectorAll("form")).item(0);`}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 19, EndLine: 1, EndColumn: 39},
		}},
		{Code: `window?.document.getElementsByTagName("form")[0];`, FileName: "extras.js", Output: []string{"window?.document.querySelectorAll(\"form\")[0];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 18, EndLine: 1, EndColumn: 38},
		}},
		{Code: `(window?.document).getElementsByTagName("form")[0];`, FileName: "extras.js", Output: []string{"(window?.document).querySelectorAll(\"form\")[0];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 40},
		}},
		{Code: `(/** @type {NodeList} */ (document.getElementsByTagName("form")))[0];`, FileName: "extras.js", Output: []string{"(/** @type {NodeList} */ (document.querySelectorAll(\"form\")))[0];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 36, EndLine: 1, EndColumn: 56},
		}},
		{Code: `(/** @type {NodeList} */ (window?.document.getElementsByTagName("form")))[0];`, FileName: "extras.js", Output: []string{`(/** @type {NodeList} */ (window?.document.querySelectorAll("form")))[0];`}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 44, EndLine: 1, EndColumn: 64},
		}},
		// Simple selectors work on fragments and shadow roots without :scope.
		{Code: `fragment.getElementById("foo");`, FileName: "extras.js", Output: []string{`fragment.querySelector("#foo");`}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 24},
		}},
		{Code: `shadowRoot.getElementById("foo");`, FileName: "extras.js", Output: []string{`shadowRoot.querySelector("#foo");`}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 12, EndLine: 1, EndColumn: 26},
		}},
		{Code: `document.getElementById((null as any)!);`, FileName: "extras.ts", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 24},
		}},
		{Code: `document.getElementsByName(null).item(0);`, FileName: "extras.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementsByName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 27},
		}},
		// Indexing an empty collection returns undefined; item(0) returns null.
		{Code: `document.getElementsByTagName("missing")[0] === undefined;`, FileName: "extras.js", Output: []string{`document.querySelectorAll("missing")[0] === undefined;`}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: `document.getElementsByName("missing").at(0) === undefined;`, FileName: "extras.js", Output: []string{`document.querySelectorAll("[name='missing']").at(0) === undefined;`}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 27},
		}},
		{Code: `document.getElementsByTagName("missing").item(0) === null;`, FileName: "extras.js", Output: []string{`document.querySelector("missing") === null;`}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		// Only DOM ASCII whitespace separates class tokens; VT and NBSP do not.
		{Code: `document.getElementsByClassName("foo\tbar\n\f\r baz");`, FileName: "extras.js", Output: []string{`document.querySelectorAll(".foo.bar.baz");`}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 32},
		}},
		{Code: `document.getElementsByClassName("foo\u00a0bar");`, FileName: "extras.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 32},
		}},
		{Code: `document.getElementsByClassName("foo\vbar");`, FileName: "extras.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 32},
		}},
		// Actual source characters, rather than backslash escape spellings.
		{Code: "document.getElementsByName(\"foo\fbar\");", FileName: "extras.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 27},
		}},
		{Code: "document.getElementsByName(\"foo\x00bar\");", FileName: "extras.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 27},
		}},
		{Code: "document.getElementsByName(`foo\fbar`);", FileName: "extras.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 27},
		}},
		{Code: "document.getElementsByName(`foo\x00bar`);", FileName: "extras.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 27},
		}},
		// NodeList has no HTMLCollection.namedItem, including computed accesses.
		{Code: `document.getElementsByTagName("div").namedItem("named");`, FileName: "extras.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: `document.getElementsByClassName("foo")["namedItem"]("named");`, FileName: "extras.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 32},
		}},
		{Code: "document.getElementsByTagName(\"div\")[`namedItem`](\"named\");", FileName: "extras.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: `(document.getElementsByTagName("div") as HTMLCollection).namedItem("named");`, FileName: "extras.ts", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 31},
		}},
		{Code: `document.getElementsByTagName("div")?.namedItem?.("named");`, FileName: "extras.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: `document.getElementsByTagName("div").length;`, FileName: "extras.js", Output: []string{`document.querySelectorAll("div").length;`}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
	})
}

func TestPreferQuerySelectorArtifactsFollowDemand(t *testing.T) {
	for _, test := range []struct{ code, output string }{
		{`element.getElementsByClassName("foo bar").at(0);`, `element.querySelectorAll(".foo.bar").at(0);`},
		{`document.getElementsByTagName(null)[0];`, `document.querySelectorAll(null)[0];`},
		{`document.getElementById("#foo");`, `document.getElementById("#foo");`},
		{`element.getElementsByClassName("foo bar").item(0);`, `element.querySelector(".foo.bar");`},
		{`document.getElementById(null);`, `document.getElementById(null);`},
		{`document.getElementsByTagName("div").namedItem("named");`, `document.getElementsByTagName("div").namedItem("named");`},
	} {
		t.Run(test.code, func(t *testing.T) {
			program, sourceFile, err := rule_tester.NewProgramHelper(fixtures.GetRootDir()).CreateTestProgram(test.code, "edit-demand.js", "tsconfig.json")
			if err != nil {
				t.Fatal(err)
			}
			demands := []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll}
			diagnostics := make([]rule.RuleDiagnostic, len(demands))
			for index, demand := range demands {
				var found []rule.RuleDiagnostic
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program: lintprogram.NewFromCompiler(program), File: sourceFile.FileName(),
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{Name: prefer_query_selector.PreferQuerySelectorRule.Name, Severity: rule.SeverityError, Run: func(ctx rule.RuleContext) rule.RuleListeners {
							return prefer_query_selector.PreferQuerySelectorRule.Run(ctx, nil)
						}}}
					},
					Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { found = append(found, d) }},
				})
				if len(found) != 1 {
					t.Fatalf("demand %d: got %d diagnostics, want 1", demand, len(found))
				}
				diagnostics[index] = found[0]
			}
			all := diagnostics[len(diagnostics)-1]
			for index, demand := range demands {
				diagnostic := diagnostics[index]
				if diagnostic.Range != all.Range || !reflect.DeepEqual(diagnostic.Message, all.Message) {
					t.Errorf("demand %d changed diagnostic identity", demand)
				}
				wantFix := test.code != test.output && (demand == rule.EditDemandAutofix || demand == rule.EditDemandAll)
				if (diagnostic.FixesPtr != nil) != wantFix {
					t.Errorf("demand %d: unexpected fix artifacts", demand)
				}
				if wantFix && !reflect.DeepEqual(diagnostic.FixesPtr, all.FixesPtr) {
					t.Errorf("demand %d changed fix artifacts", demand)
				}
				if diagnostic.Suggestions != nil {
					t.Errorf("demand %d: unexpected suggestions", demand)
				}
			}
			// Applying fixes can sort slices; finish artifact comparisons first.
			for index, demand := range demands {
				wantFix := test.code != test.output && (demand == rule.EditDemandAutofix || demand == rule.EditDemandAll)
				expected := test.code
				if wantFix {
					expected = test.output
				}
				output, _, fixed := linter.ApplyRuleFixes(test.code, diagnostics[index:index+1])
				if output != expected || fixed != wantFix {
					t.Errorf("demand %d: output = %q, want %q", demand, output, expected)
				}
			}
		})
	}
}

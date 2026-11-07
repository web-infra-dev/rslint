// Upstream source, tests, and documentation pinned to eslint-plugin-unicorn v79.0.0.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-query-selector.js
package prefer_query_selector_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_query_selector"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferQuerySelectorUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_query_selector.PreferQuerySelectorRule, []rule_tester.ValidTestCase{
		// test/prefer-query-selector.js, including all not-dom-node-types fixtures.
		{Code: "new document.getElementById(foo);", FileName: "upstream.js"},
		{Code: "getElementById(foo);", FileName: "upstream.js"},
		{Code: "document['getElementById'](bar);", FileName: "upstream.js"},
		{Code: "document[getElementById](bar);", FileName: "upstream.js"},
		{Code: "document.foo(bar);", FileName: "upstream.js"},
		{Code: "document.getElementById();", FileName: "upstream.js"},
		{Code: "document.getElementsByClassName(\"foo\", \"bar\");", FileName: "upstream.js"},
		{Code: "document.getElementById(...[\"id\"]);", FileName: "upstream.js"},
		{Code: "([]).getElementById(foo)", FileName: "upstream.js"},
		{Code: "([element]).getElementById(foo)", FileName: "upstream.js"},
		{Code: "([...elements]).getElementById(foo)", FileName: "upstream.js"},
		{Code: "(() => {}).getElementById(foo)", FileName: "upstream.js"},
		{Code: "(class Node {}).getElementById(foo)", FileName: "upstream.js"},
		{Code: "(function() {}).getElementById(foo)", FileName: "upstream.js"},
		{Code: "(0).getElementById(foo)", FileName: "upstream.js"},
		{Code: "(1).getElementById(foo)", FileName: "upstream.js"},
		{Code: "(0.1).getElementById(foo)", FileName: "upstream.js"},
		{Code: "(\"\").getElementById(foo)", FileName: "upstream.js"},
		{Code: "(\"string\").getElementById(foo)", FileName: "upstream.js"},
		{Code: "(/regex/).getElementById(foo)", FileName: "upstream.js"},
		{Code: "(null).getElementById(foo)", FileName: "upstream.js"},
		{Code: "(0n).getElementById(foo)", FileName: "upstream.js"},
		{Code: "(1n).getElementById(foo)", FileName: "upstream.js"},
		{Code: "(true).getElementById(foo)", FileName: "upstream.js"},
		{Code: "(false).getElementById(foo)", FileName: "upstream.js"},
		{Code: "({}).getElementById(foo)", FileName: "upstream.js"},
		{Code: "(`templateLiteral`).getElementById(foo)", FileName: "upstream.js"},
		{Code: "(undefined).getElementById(foo)", FileName: "upstream.js"},
		{Code: "document.querySelector(\"#foo\");", FileName: "upstream.js"},
		{Code: "document.querySelector(\".bar\");", FileName: "upstream.js"},
		{Code: "document.querySelector(\"main #foo .bar\");", FileName: "upstream.js"},
		{Code: "document.querySelectorAll(\".foo .bar\");", FileName: "upstream.js"},
		{Code: "document.querySelectorAll(\"li a\");", FileName: "upstream.js"},
		{Code: "document.querySelector(\"li\").querySelectorAll(\"a\");", FileName: "upstream.js"},
		{Code: "document.getElementsByName();", FileName: "upstream.js"},
		{Code: "document.getElementById(someId);", FileName: "upstream.js", Options: map[string]any{"allowWithVariables": true}},
		{Code: "document.getElementsByClassName(someClass);", FileName: "upstream.js", Options: map[string]any{"allowWithVariables": true}},
		{Code: "document.getElementsByClassName(fn());", FileName: "upstream.js", Options: map[string]any{"allowWithVariables": true}},
		{Code: "document.getElementsByClassName(`${someClass}`);", FileName: "upstream.js", Options: map[string]any{"allowWithVariables": true}},
		{Code: "document.getElementById(`${someId}`);", FileName: "upstream.js", Options: map[string]any{"allowWithVariables": true}},
		// docs/rules/prefer-query-selector.md.
		{Code: "document.querySelector('#foo');", FileName: "upstream.js"},
		{Code: "document.querySelectorAll('.foo');", FileName: "upstream.js"},
		{Code: "document.querySelectorAll('.foo.bar');", FileName: "upstream.js"},
		{Code: "document.querySelectorAll('main');", FileName: "upstream.js"},
		{Code: "const singleClassName = 'foo';\ndocument.querySelectorAll(`.${CSS.escape(singleClassName)}`);", FileName: "upstream.js"},
		{Code: "document.querySelector('form');", FileName: "upstream.js"},
		{Code: "document.getElementById(someId);\ndocument.getElementsByClassName(someClass);", FileName: "upstream.js", Options: map[string]any{"allowWithVariables": true}},
	}, []rule_tester.InvalidTestCase{
		// test/prefer-query-selector.js.
		{Code: "document.getElementById(\"foo\");", FileName: "upstream.js", Output: []string{"document.querySelector(\"#foo\");"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 24},
		}},
		{Code: "document.getElementById(\"#foo\");", FileName: "upstream.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 24},
		}},
		{Code: "document.getElementById(\".foo\");", FileName: "upstream.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 24},
		}},
		{Code: "document.getElementById(\"foo bar\");", FileName: "upstream.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 24},
		}},
		{Code: "document.getElementById(\"foo:bar\");", FileName: "upstream.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 24},
		}},
		{Code: "document.getElementById(\"foo\" as string);", FileName: "upstream.ts", Output: []string{"document.querySelector(\"#foo\" as string);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 24},
		}},
		{Code: "document.getElementById(\"#foo\" as string);", FileName: "upstream.ts", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 24},
		}},
		{Code: "document.getElementsByClassName(\"foo\");", FileName: "upstream.js", Output: []string{"document.querySelectorAll(\".foo\");"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 32},
		}},
		{Code: "document.getElementsByClassName(\".foo\");", FileName: "upstream.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 32},
		}},
		{Code: "document.getElementsByClassName(\"#foo\");", FileName: "upstream.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 32},
		}},
		{Code: "document.getElementsByClassName(\"foo:bar\");", FileName: "upstream.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 32},
		}},
		{Code: "document.getElementsByClassName(\"foo\" as string);", FileName: "upstream.ts", Output: []string{"document.querySelectorAll(\".foo\" as string);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 32},
		}},
		{Code: "document.getElementsByClassName(\".foo\" as string);", FileName: "upstream.ts", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 32},
		}},
		{Code: "element.getElementsByClassName(\"foo\");", FileName: "upstream.js", Output: []string{"element.querySelectorAll(\":scope .foo\");"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 31},
		}},
		{Code: "element.getElementsByClassName(\".foo\");", FileName: "upstream.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 31},
		}},
		{Code: "document.getElementsByClassName(\"foo bar\");", FileName: "upstream.js", Output: []string{"document.querySelectorAll(\".foo.bar\");"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 32},
		}},
		{Code: "document.getElementsByTagName(\"foo\");", FileName: "upstream.js", Output: []string{"document.querySelectorAll(\"foo\");"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: "document.getElementsByTagName(\"foo:bar\");", FileName: "upstream.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: "document.getElementsByTagName(\"*\");", FileName: "upstream.js", Output: []string{"document.querySelectorAll(\"*\");"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: "element.getElementsByTagName(\"foo\");", FileName: "upstream.js", Output: []string{"element.querySelectorAll(\":scope foo\");"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 29},
		}},
		{Code: "element.getElementsByTagName(\"foo bar\");", FileName: "upstream.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 29},
		}},
		{Code: "document.getElementById(\"\");", FileName: "upstream.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 24},
		}},
		{Code: "document.getElementById('foo');", FileName: "upstream.js", Output: []string{"document.querySelector('#foo');"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 24},
		}},
		{Code: "document.getElementsByClassName('foo');", FileName: "upstream.js", Output: []string{"document.querySelectorAll('.foo');"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 32},
		}},
		{Code: "document.getElementsByClassName('foo bar');", FileName: "upstream.js", Output: []string{"document.querySelectorAll('.foo.bar');"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 32},
		}},
		{Code: "document.getElementsByTagName('foo');", FileName: "upstream.js", Output: []string{"document.querySelectorAll('foo');"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: "document.getElementsByClassName('');", FileName: "upstream.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 32},
		}},
		{Code: "document.getElementById(`foo`);", FileName: "upstream.js", Output: []string{"document.querySelector(`#foo`);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 24},
		}},
		{Code: "document.getElementById(`#foo`);", FileName: "upstream.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 24},
		}},
		{Code: "document.getElementsByClassName(`foo`);", FileName: "upstream.js", Output: []string{"document.querySelectorAll(`.foo`);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 32},
		}},
		{Code: "document.getElementsByClassName(`.foo`);", FileName: "upstream.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 32},
		}},
		{Code: "element.getElementsByClassName(`foo`);", FileName: "upstream.js", Output: []string{"element.querySelectorAll(`:scope .foo`);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 31},
		}},
		{Code: "document.getElementsByClassName(`foo bar`);", FileName: "upstream.js", Output: []string{"document.querySelectorAll(`.foo.bar`);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 32},
		}},
		{Code: "document.getElementsByTagName(`foo`);", FileName: "upstream.js", Output: []string{"document.querySelectorAll(`foo`);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: "element.getElementsByTagName(`foo`);", FileName: "upstream.js", Output: []string{"element.querySelectorAll(`:scope foo`);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 29},
		}},
		{Code: "document.getElementsByTagName(``);", FileName: "upstream.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: "document.getElementsByClassName(`${fn()}`);", FileName: "upstream.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 32},
		}},
		{Code: "document.getElementsByClassName(`foo ${undefined}`);", FileName: "upstream.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 32},
		}},
		{Code: "document.getElementsByClassName(null);", FileName: "upstream.js", Output: []string{"document.querySelectorAll(null);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 32},
		}},
		{Code: "document.getElementsByTagName(null);", FileName: "upstream.js", Output: []string{"document.querySelectorAll(null);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: "document.getElementsByClassName(fn());", FileName: "upstream.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 32},
		}},
		{Code: "document.getElementsByClassName(\"foo\" + fn());", FileName: "upstream.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 32},
		}},
		{Code: "document.getElementsByClassName(foo + \"bar\");", FileName: "upstream.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 32},
		}},
		{Code: "for (const div of document.body.getElementById(\"id\").getElementsByClassName(\"class\")) {\n\tconsole.log(div.getElementsByTagName(\"div\"));\n}", FileName: "upstream.js", Output: []string{"for (const div of document.body.querySelector(\":scope #id\").querySelectorAll(\":scope .class\")) {\n\tconsole.log(div.querySelectorAll(\":scope div\"));\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			// The Go harness observes outer-call-first traversal; public diagnostics are source-sorted.
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 54, EndLine: 1, EndColumn: 76},
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 33, EndLine: 1, EndColumn: 47},
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 2, Column: 18, EndLine: 2, EndColumn: 38},
		}},
		{Code: "e.getElementById(3)", FileName: "upstream.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 3, EndLine: 1, EndColumn: 17},
		}},
		{Code: "document.getElementsByName(\"foo\");", FileName: "upstream.js", Output: []string{"document.querySelectorAll(\"[name='foo']\");"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 27},
		}},
		{Code: "document.getElementsByName('foo');", FileName: "upstream.js", Output: []string{"document.querySelectorAll('[name=\"foo\"]');"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 27},
		}},
		{Code: "document.getElementsByName(`foo`);", FileName: "upstream.js", Output: []string{"document.querySelectorAll(`[name='foo']`);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 27},
		}},
		{Code: "document.getElementsByName(`${'foo'}`);", FileName: "upstream.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 27},
		}},
		{Code: "document.getElementsByName(null);", FileName: "upstream.js", Output: []string{"document.querySelectorAll(null);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 27},
		}},
		{Code: "document.getElementsByName(\"\");", FileName: "upstream.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 27},
		}},
		{Code: "document.getElementsByName(foo + \"bar\");", FileName: "upstream.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 27},
		}},
		{Code: "document.getElementsByName(\"multiple name should be fixable\");", FileName: "upstream.js", Output: []string{"document.querySelectorAll(\"[name='multiple name should be fixable']\");"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 27},
		}},
		{Code: "document.getElementsByName(\"foo'bar\");", FileName: "upstream.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 27},
		}},
		{Code: "document.getElementsByName('foo\"bar');", FileName: "upstream.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 27},
		}},
		{Code: "document.getElementsByName(`foo'bar`);", FileName: "upstream.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 27},
		}},
		{Code: "document.getElementsByName(\"foo\\\\bar\");", FileName: "upstream.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 27},
		}},
		{Code: "document.getElementsByName(\"foo\\nbar\");", FileName: "upstream.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 27},
		}},
		{Code: "document.getElementsByName(\"foo\\\\bar\" as string);", FileName: "upstream.ts", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 27},
		}},
		{Code: "document.getElementsByTagName(\"form\")[0].addEventListener(\"submit\", submitFunction);", FileName: "upstream.js", Output: []string{"document.querySelector(\"form\").addEventListener(\"submit\", submitFunction);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: "document.getElementsByTagName(\"form\").item(0).submit();", FileName: "upstream.js", Output: []string{"document.querySelector(\"form\").submit();"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: "document.getElementsByClassName(\"submit-button\").at(0).click();", FileName: "upstream.js", Output: []string{"document.querySelector(\".submit-button\").click();"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementsByClassName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 32},
		}},
		{Code: "document.getElementsByName(\"email\")[0].focus();", FileName: "upstream.js", Output: []string{"document.querySelector(\"[name='email']\").focus();"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementsByName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 27},
		}},
		{Code: "document.getElementsByName(\"email\").item(0).focus();", FileName: "upstream.js", Output: []string{"document.querySelector(\"[name='email']\").focus();"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementsByName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 27},
		}},
		{Code: "delete document.getElementsByTagName(\"form\").at(0);", FileName: "upstream.js", Output: []string{"delete document.querySelectorAll(\"form\").at(0);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 17, EndLine: 1, EndColumn: 37},
		}},
		{Code: "for (document.getElementsByTagName(\"form\")[0] in object) {}", FileName: "upstream.js", Output: []string{"for (document.querySelectorAll(\"form\")[0] in object) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 15, EndLine: 1, EndColumn: 35},
		}},
		{Code: "for (document.getElementsByTagName(\"form\")[0] of iterable) {}", FileName: "upstream.js", Output: []string{"for (document.querySelectorAll(\"form\")[0] of iterable) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 15, EndLine: 1, EndColumn: 35},
		}},
		{Code: "document.getElementsByTagName(\"form\") /* keep */ [0].addEventListener(\"submit\", submitFunction);", FileName: "upstream.js", Output: []string{"document.querySelectorAll(\"form\") /* keep */ [0].addEventListener(\"submit\", submitFunction);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: "document.getElementsByTagName(\"form\").item(/* keep */ 0).addEventListener(\"submit\", submitFunction);", FileName: "upstream.js", Output: []string{"document.querySelectorAll(\"form\").item(/* keep */ 0).addEventListener(\"submit\", submitFunction);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: "document.getElementById(\"foo\");", FileName: "upstream.js", Options: map[string]any{"allowWithVariables": true}, Output: []string{"document.querySelector(\"#foo\");"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 24},
		}},
		{Code: "document.getElementById(\"foo\" as string);", FileName: "upstream.ts", Options: map[string]any{"allowWithVariables": true}, Output: []string{"document.querySelector(\"#foo\" as string);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 24},
		}},
		{Code: "document.getElementsByClassName(\"foo\");", FileName: "upstream.js", Options: map[string]any{"allowWithVariables": true}, Output: []string{"document.querySelectorAll(\".foo\");"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 32},
		}},
		{Code: "document.getElementsByClassName(\"foo\"!);", FileName: "upstream.ts", Options: map[string]any{"allowWithVariables": true}, Output: []string{"document.querySelectorAll(\".foo\"!);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 32},
		}},
		{Code: "document.getElementsByTagName(\"foo\");", FileName: "upstream.js", Options: map[string]any{"allowWithVariables": true}, Output: []string{"document.querySelectorAll(\"foo\");"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: "document.getElementsByName(\"foo\");", FileName: "upstream.js", Options: map[string]any{"allowWithVariables": true}, Output: []string{"document.querySelectorAll(\"[name='foo']\");"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 27},
		}},
		{Code: "document.getElementsByTagName(someTag);", FileName: "upstream.js", Options: map[string]any{"allowWithVariables": true}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: "document.getElementsByName(someName);", FileName: "upstream.js", Options: map[string]any{"allowWithVariables": true}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 27},
		}},
		{Code: "document.getElementsByClassName(null);", FileName: "upstream.js", Options: map[string]any{"allowWithVariables": true}, Output: []string{"document.querySelectorAll(null);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 32},
		}},
		{Code: "document.getElementsByClassName(`foo`);", FileName: "upstream.js", Options: map[string]any{"allowWithVariables": true}, Output: []string{"document.querySelectorAll(`.foo`);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 32},
		}},
		{Code: "document.getElementsByClassName(variable + \"x\");", FileName: "upstream.js", Options: map[string]any{"allowWithVariables": true}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 32},
		}},
		{Code: "document.getElementsByClassName(\"foo\" + fn());", FileName: "upstream.js", Options: map[string]any{"allowWithVariables": true}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 32},
		}},
		{Code: "document.getElementsByClassName(foo + \"bar\");", FileName: "upstream.js", Options: map[string]any{"allowWithVariables": true}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 32},
		}},
		{Code: "document.getElementById(`x${someId}`);", FileName: "upstream.js", Options: map[string]any{"allowWithVariables": true}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 24},
		}},
		{Code: "document.getElementsByClassName(`foo ${someClass}`);", FileName: "upstream.js", Options: map[string]any{"allowWithVariables": true}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 32},
		}},
		{Code: "window.frame.getElementById(\"bar\");", FileName: "upstream.js", Output: []string{"window.frame.querySelector(\":scope #bar\");"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 28},
		}},
		{Code: "class Foo { #document; bar() { return window.#document.getElementById(\"bar\"); } }", FileName: "upstream.js", Output: []string{"class Foo { #document; bar() { return window.#document.querySelector(\":scope #bar\"); } }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 56, EndLine: 1, EndColumn: 70},
		}},
		{Code: "for (const element of document.getElementsByClassName(\"foo\")[0]) {}", FileName: "upstream.js", Output: []string{"for (const element of document.querySelector(\".foo\")) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementsByClassName()`.", Line: 1, Column: 32, EndLine: 1, EndColumn: 54},
		}},
		// docs/rules/prefer-query-selector.md.
		{Code: "document.getElementById('foo');", FileName: "upstream.js", Output: []string{"document.querySelector('#foo');"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 24},
		}},
		{Code: "document.getElementsByClassName('foo');", FileName: "upstream.js", Output: []string{"document.querySelectorAll('.foo');"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 32},
		}},
		{Code: "document.getElementsByClassName('foo bar');", FileName: "upstream.js", Output: []string{"document.querySelectorAll('.foo.bar');"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 32},
		}},
		{Code: "document.getElementsByTagName('main');", FileName: "upstream.js", Output: []string{"document.querySelectorAll('main');"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: "const singleClassName = 'foo';\ndocument.getElementsByClassName(singleClassName);", FileName: "upstream.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 2, Column: 10, EndLine: 2, EndColumn: 32},
		}},
		{Code: "document.getElementsByTagName('form')[0];", FileName: "upstream.js", Output: []string{"document.querySelector('form');"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
		}},
		{Code: "document.getElementById('foo');\ndocument.getElementsByClassName('foo');", FileName: "upstream.js", Options: map[string]any{"allowWithVariables": true}, Output: []string{"document.querySelector('#foo');\ndocument.querySelectorAll('.foo');"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelector()` over `.getElementById()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 24},
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByClassName()`.", Line: 2, Column: 10, EndLine: 2, EndColumn: 32},
		}},
		{Code: "document.getElementsByTagName(someTag);\ndocument.getElementsByName(someName);", FileName: "upstream.js", Options: map[string]any{"allowWithVariables": true}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByTagName()`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 30},
			{MessageId: "prefer-query-selector", Message: "Prefer `.querySelectorAll()` over `.getElementsByName()`.", Line: 2, Column: 10, EndLine: 2, EndColumn: 27},
		}},
	})
}

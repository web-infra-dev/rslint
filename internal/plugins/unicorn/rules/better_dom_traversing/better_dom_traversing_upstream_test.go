// Ported from eslint-plugin-unicorn v79.0.0. See LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/better-dom-traversing.js
package better_dom_traversing_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/better_dom_traversing"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestBetterDomTraversingUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: "", FileName: "upstream.js"},
		// Writes to collection indexes.
		{Code: "element.childNodes[0] = node;", FileName: "upstream.js"},
		{Code: "element.children[0] = node;", FileName: "upstream.js"},
		{Code: "element.childNodes[0] += node;", FileName: "upstream.js"},
		{Code: "element.childNodes[0] ||= node;", FileName: "upstream.js"},
		{Code: "element.childNodes[0]++;", FileName: "upstream.js"},
		{Code: "delete element.childNodes[0];", FileName: "upstream.js"},
		{Code: "for (element.childNodes[0] of nodes) {}", FileName: "upstream.js"},
		{Code: "for (element.childNodes[0] in nodes) {}", FileName: "upstream.js"},
		{Code: "[element.childNodes[0]] = nodes;", FileName: "upstream.js"},
		{Code: "({node: element.childNodes[0]} = object);", FileName: "upstream.js"},
		{Code: "element.children[1] = node;", FileName: "upstream.js"},
		// Preferred APIs.
		{Code: "element.firstChild;", FileName: "upstream.js"},
		{Code: "element.firstElementChild;", FileName: "upstream.js"},
		{Code: "element.querySelector(\"li\");", FileName: "upstream.js"},
		{Code: "element.closest(\"form\");", FileName: "upstream.js"},
		// Traversal on a bare identifier.
		{Code: "children[0];", FileName: "upstream.js"},
		{Code: "parentElement.parentElement;", FileName: "upstream.js"},
		// Non-literal and unsupported indexes.
		{Code: "element.children[index];", FileName: "upstream.js"},
		{Code: "element.children[\"0\"];", FileName: "upstream.js"},
		{Code: "element.children[-1];", FileName: "upstream.js"},
		{Code: "element.children[1.5];", FileName: "upstream.js"},
		{Code: "element.childNodes[index];", FileName: "upstream.js"},
		{Code: "element.childNodes[1];", FileName: "upstream.js"},
		// Computed collection names.
		{Code: "element[\"children\"][0];", FileName: "upstream.js"},
		{Code: "element[children][0];", FileName: "upstream.js"},
		{Code: "element[\"childNodes\"][0];", FileName: "upstream.js"},
		{Code: "element[childNodes][0];", FileName: "upstream.js"},
		{Code: "element[\"parentElement\"].parentElement;", FileName: "upstream.js"},
		{Code: "element.parentElement[\"parentElement\"];", FileName: "upstream.js"},
		// Optional chaining.
		{Code: "element?.children[0];", FileName: "upstream.js"},
		{Code: "element.children?.[0];", FileName: "upstream.js"},
		{Code: "element?.parentElement.parentElement;", FileName: "upstream.js"},
		{Code: "element.parentElement?.parentElement;", FileName: "upstream.js"},
		{Code: "element.querySelector?.(\"a\").querySelector(\"b\");", FileName: "upstream.js"},
		{Code: "element.querySelector(\"a\")?.querySelector(\"b\");", FileName: "upstream.js"},
		{Code: "element.querySelector?.(\"a\").querySelector(\"b\").querySelector(\"c\");", FileName: "upstream.js"},
		{Code: "element.querySelector(\"a\")?.querySelector(\"b\").querySelector(\"c\");", FileName: "upstream.js"},
		{Code: "element.querySelector(\"a\").querySelector(\"b\")?.querySelector(\"c\");", FileName: "upstream.js"},
		{Code: "element.querySelector(\"a\").querySelector(\"b\")?.foo;", FileName: "upstream.js"},
		{Code: "element.querySelector(\"a\").querySelector(\"b\").foo?.querySelector(\"c\");", FileName: "upstream.js"},
		{Code: "(element?.querySelector(\"a\")).querySelector(\"b\").querySelector(\"c\");", FileName: "upstream.js"},
		// Other methods and argument counts.
		{Code: "element.querySelector();", FileName: "upstream.js"},
		{Code: "element.querySelector(\"a\", root).querySelector(\"b\");", FileName: "upstream.js"},
		{Code: "element.querySelector(\"a\").querySelector();", FileName: "upstream.js"},
		{Code: "element.querySelector(\"a\").querySelector(\"b\", root);", FileName: "upstream.js"},
		{Code: "element.querySelectorAll(\"a\").querySelector(\"b\");", FileName: "upstream.js"},
		{Code: "element.querySelector(\"a\").querySelectorAll(\"b\");", FileName: "upstream.js"},
		// Non-static selectors.
		{Code: "element.querySelector(selector).querySelector(\"b\");", FileName: "upstream.js"},
		{Code: "element.querySelector(\"a\").querySelector(selector);", FileName: "upstream.js"},
		{Code: "element.querySelector(`${selector}`).querySelector(\"b\");", FileName: "upstream.js"},
		{Code: "element.querySelector(\"a\").querySelector(`${selector}`);", FileName: "upstream.js"},
		{Code: "element.querySelector(tag`a`).querySelector(\"b\");", FileName: "upstream.js"},
		// Component props.children.
		{Code: "class Component { render() { return <div>{this.props.children[0]}</div>; } }", FileName: "upstream.jsx"},
		{Code: "const Component = props => <div>{props.children[0]}</div>;", FileName: "upstream.jsx"},
		{Code: "const child = props.children[1];", FileName: "upstream.js"},
		{Code: "const child = component.props.children[0];", FileName: "upstream.js"},
		// Remaining documentation examples.
		{Code: "element.querySelector(':scope selector');", FileName: "upstream.js"},
		{Code: "element.closest('selector');", FileName: "upstream.js"},
		{Code: "element.querySelector(':scope a b');", FileName: "upstream.js"},
	}
	// The upstream not-dom-node-types fixtures are tested against all four traversal forms.
	for _, receiver := range []string{
		"[]",
		"[element]",
		"[...elements]",
		"() => {}",
		"class Node {}",
		"function() {}",
		"0",
		"1",
		"0.1",
		"\"\"",
		"\"string\"",
		"/regex/",
		"null",
		"0n",
		"1n",
		"true",
		"false",
		"{}",
		"`templateLiteral`",
		"undefined",
	} {
		for _, access := range []string{
			".children[0]", ".childNodes[0]", ".parentElement.parentElement",
			".querySelector(\"a\").querySelector(\"b\")",
		} {
			valid = append(valid, rule_tester.ValidTestCase{Code: "(" + receiver + ")" + access + ";", FileName: "upstream.js"})
		}
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &better_dom_traversing.BetterDomTraversingRule, valid, []rule_tester.InvalidTestCase{
		invalidTraversal("element.childNodes[0];", "first-child", 1, 1, 1, 22, "element.firstChild;"),
		invalidTraversal("(element.childNodes)[0];", "first-child", 1, 1, 1, 24, "(element.firstChild);"),
		invalidTraversal("(element.children)[0];", "first-element-child", 1, 1, 1, 22, "(element.firstElementChild);"),
		invalidTraversal("element.children[0];", "first-element-child", 1, 1, 1, 20, "element.firstElementChild;"),
		invalidTraversal("element.children[1];", "query-selector", 1, 1, 1, 20, ""),
		invalidTraversal("element.children[10];", "query-selector", 1, 1, 1, 21, ""),
		invalidTraversal("element.children[1].children[2];", "query-selector", 1, 1, 1, 32, ""),
		invalidTraversal("element.parentElement.parentElement;", "closest", 1, 1, 1, 36, ""),
		invalidTraversal("element.parentElement.parentElement.parentElement;", "closest", 1, 1, 1, 50, ""),
		invalidTraversal("element.querySelector(\"a\").querySelector(\"b\");", "merge-query-selector", 1, 1, 1, 46, "element.querySelector(\":scope a b\");"),
		invalidTraversal("document.querySelector(\"a\").querySelector(\"b\");", "merge-query-selector", 1, 1, 1, 47, "document.querySelector(\"a b\");"),
		invalidTraversal("document.body.querySelector(\"a\").querySelector(\"b\");", "merge-query-selector", 1, 1, 1, 52, "document.body.querySelector(\":scope a b\");"),
		invalidTraversal("element.querySelector('a').querySelector('b');", "merge-query-selector", 1, 1, 1, 46, "element.querySelector(':scope a b');"),
		invalidTraversal("element.querySelector(`a`).querySelector(`b`);", "merge-query-selector", 1, 1, 1, 46, "element.querySelector(':scope a b');"),
		invalidTraversal("element.querySelector(\"a\").querySelector(`b`);", "merge-query-selector", 1, 1, 1, 46, "element.querySelector(\":scope a b\");"),
		invalidTraversal("element.querySelector(\"a > b\").querySelector(\".c\");", "merge-query-selector", 1, 1, 1, 51, "element.querySelector(\":scope a > b .c\");"),
		invalidTraversal("element.querySelector(\".a, .b\").querySelector(\".c\");", "merge-query-selector", 1, 1, 1, 52, ""),
		invalidTraversal("element.querySelector(\".a\").querySelector(\".b, .c\");", "merge-query-selector", 1, 1, 1, 52, ""),
		invalidTraversal("element.querySelector(\":scope a\").querySelector(\"b\");", "merge-query-selector", 1, 1, 1, 53, ""),
		invalidTraversal("element.querySelector(\"a\").querySelector(\"b\").querySelector(\"c\");", "merge-query-selector", 1, 1, 1, 65, "element.querySelector(\":scope a b c\");"),
		invalidTraversal("element.querySelector(\"a\").querySelector(\"b\").querySelector(selector);", "merge-query-selector", 1, 1, 1, 46, "element.querySelector(\":scope a b\").querySelector(selector);"),
		invalidTraversal("element.querySelector(selector).querySelector(\"b\").querySelector(\"c\");", "merge-query-selector", 1, 1, 1, 70, "element.querySelector(selector).querySelector(\":scope b c\");"),
		invalidTraversal("(getElement()).querySelector(\"a\").querySelector(\"b\");", "merge-query-selector", 1, 1, 1, 53, "(getElement()).querySelector(\":scope a b\");"),
		invalidTraversal("(foo || bar).querySelector(\"a\").querySelector(\"b\");", "merge-query-selector", 1, 1, 1, 51, "(foo || bar).querySelector(\":scope a b\");"),
		invalidTraversal("element.childNodes[/* comment */ 0];", "first-child", 1, 1, 1, 36, ""),
		invalidTraversal("element.children[/* comment */ 0];", "first-element-child", 1, 1, 1, 34, ""),
		invalidTraversal("element.querySelector(/* comment */ \"a\").querySelector(\"b\");", "merge-query-selector", 1, 1, 1, 60, ""),
		invalidTraversal("element.querySelector(\"a\").querySelector(/* comment */ \"b\");", "merge-query-selector", 1, 1, 1, 60, ""),
		invalidTraversal("const item = element\n\t.children[0];", "first-element-child", 1, 14, 2, 14, "const item = element\n\t.firstElementChild;"),
		invalidTraversal("const item = element\n\t.querySelector(\"a\")\n\t.querySelector(\"b\");", "merge-query-selector", 1, 14, 3, 21, "const item = element.querySelector(\":scope a b\");"),
		invalidTraversal("element.childNodes[0].textContent = text;", "first-child", 1, 1, 1, 22, "element.firstChild.textContent = text;"),
		invalidTraversal("window.document.querySelector(\"a\").querySelector(\"b\");", "merge-query-selector", 1, 1, 1, 54, "window.document.querySelector(\"a b\");"),
		invalidTraversal("globalThis.document.querySelector(\"a\").querySelector(\"b\");", "merge-query-selector", 1, 1, 1, 58, "globalThis.document.querySelector(\"a b\");"),
		invalidTraversal("window.frame.querySelector(\"a\").querySelector(\"b\");", "merge-query-selector", 1, 1, 1, 51, "window.frame.querySelector(\":scope a b\");"),
		invalidTraversal("window[document].querySelector(\"a\").querySelector(\"b\");", "merge-query-selector", 1, 1, 1, 55, "window[document].querySelector(\":scope a b\");"),
		// Upstream regression cases for a receiver starting with an object literal.
		invalidTraversal("({element}.element.querySelector(\"a\")).querySelector(\"b\");", "merge-query-selector", 1, 1, 1, 58, "({element}.element).querySelector(\":scope a b\");"),
		invalidTraversal("const result = ({element}.element.querySelector(\"a\")).querySelector(\"b\");", "merge-query-selector", 1, 16, 1, 73, "const result = {element}.element.querySelector(\":scope a b\");"),
		invalidTraversal("const read = () => ({element}.element.querySelector(\"a\")).querySelector(\"b\");", "merge-query-selector", 1, 20, 1, 77, "const read = () => ({element}.element).querySelector(\":scope a b\");"),
		invalidTraversal("({element}.element).querySelector(\"a\").querySelector(\"b\");", "merge-query-selector", 1, 1, 1, 58, "({element}.element).querySelector(\":scope a b\");"),
		// Documentation examples. The identical first-child and parent examples appear above.
		invalidTraversal("element.children[2];", "query-selector", 1, 1, 1, 20, ""),
	})
}

var traversalMessages = map[string]string{
	"first-child":          "Prefer `.firstChild` over `.childNodes[0]`.",
	"first-element-child":  "Prefer `.firstElementChild` over `.children[0]`.",
	"query-selector":       "Consider `.querySelector()` over positional child traversal.",
	"closest":              "Consider `.closest()` over chaining `.parentElement`.",
	"merge-query-selector": "Consider merging chained `.querySelector()` calls.",
}

func invalidTraversal(code, id string, line, column, endLine, endColumn int, suggested string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code: code, FileName: "upstream.js",
		Errors: []rule_tester.InvalidTestCaseError{traversalError(id, line, column, endLine, endColumn, suggested)},
	}
}

func traversalError(id string, line, column, endLine, endColumn int, suggested string) rule_tester.InvalidTestCaseError {
	expected := rule_tester.InvalidTestCaseError{
		MessageId: id, Message: traversalMessages[id],
		Line: line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}
	if suggested != "" {
		expected.Suggestions = []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-" + id, Output: suggested}}
	}
	return expected
}

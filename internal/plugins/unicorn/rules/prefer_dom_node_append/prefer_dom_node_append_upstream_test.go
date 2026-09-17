// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/test/prefer-dom-node-append.js
package prefer_dom_node_append_test

import (
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_dom_node_append"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestPreferDomNodeAppendUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_dom_node_append.PreferDomNodeAppendRule, []rule_tester.ValidTestCase{
		{Code: "parent.append(child);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "new parent.appendChild(child);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "appendChild(child);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "parent['appendChild'](child);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "parent[appendChild](child);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "parent.foo(child);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "parent.appendChild(one, two);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "parent.appendChild();", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "parent.appendChild(...argumentsArray)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "parent.appendChild?.(child)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "([]).appendChild(foo)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "([element]).appendChild(foo)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "([...elements]).appendChild(foo)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "(() => {}).appendChild(foo)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "(class Node {}).appendChild(foo)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "(function() {}).appendChild(foo)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "(0).appendChild(foo)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "(1).appendChild(foo)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "(0.1).appendChild(foo)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "(\"\").appendChild(foo)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "(\"string\").appendChild(foo)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "(/regex/).appendChild(foo)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "(null).appendChild(foo)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "(0n).appendChild(foo)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "(1n).appendChild(foo)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "(true).appendChild(foo)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "(false).appendChild(foo)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "({}).appendChild(foo)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "(`templateLiteral`).appendChild(foo)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "(undefined).appendChild(foo)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.appendChild([])", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.appendChild([element])", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.appendChild([...elements])", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.appendChild(() => {})", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.appendChild(class Node {})", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.appendChild(function() {})", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.appendChild(0)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.appendChild(1)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.appendChild(0.1)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.appendChild(\"\")", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.appendChild(\"string\")", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.appendChild(/regex/)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.appendChild(null)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.appendChild(0n)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.appendChild(1n)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.appendChild(true)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.appendChild(false)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.appendChild({})", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.appendChild(`templateLiteral`)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "foo.appendChild(undefined)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "// ✅\nelement.append(child);\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "// ✅\n// append() can handle multiple nodes in one call\nparent.append(child1, child2, child3);\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "// ✅\n// append() can mix nodes and strings\ncontainer.append(divElement, 'Some text content');\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
	}, []rule_tester.InvalidTestCase{
		{Code: "node.appendChild(child);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"node.append(child);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "document.body.appendChild(child);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"document.body.append(child);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 33, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "node.appendChild(foo)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"node.append(foo)"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "function foo() {\n\tnode.appendChild(bar);\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"function foo() {\n\tnode.append(bar);\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 2, Column: 2, EndLine: 2, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = node.appendChild(child);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 1, Column: 13, EndLine: 1, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "console.log(node.appendChild(child));", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 1, Column: 13, EndLine: 1, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "node.appendChild(child).appendChild(grandchild);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"node.appendChild(child).append(grandchild);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "node.appendChild(child) || \"foo\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "node.appendChild(child) + 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "+node.appendChild(child);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 1, Column: 2, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "node.appendChild(child) ? \"foo\" : \"bar\";", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (node.appendChild(child)) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = [node.appendChild(child)]", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 37, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = { bar: node.appendChild(child) }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "function foo() { return node.appendChild(child); }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 1, Column: 25, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = () => { return node.appendChild(child); }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 1, Column: 28, EndLine: 1, EndColumn: 51, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo(bar = node.appendChild(child))", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 34, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "node?.appendChild(child);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"node?.append(child);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "() => node?.appendChild(child)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "// ❌\nelement.appendChild(child);\n\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"// ❌\nelement.append(child);\n\n"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 2, Column: 1, EndLine: 2, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "// ❌\n// Multiple nodes require chaining\nparent.appendChild(child1);\nparent.appendChild(child2);\nparent.appendChild(child3);\n\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"// ❌\n// Multiple nodes require chaining\nparent.append(child1);\nparent.append(child2);\nparent.append(child3);\n\n"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 3, Column: 1, EndLine: 3, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 4, Column: 1, EndLine: 4, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 5, Column: 1, EndLine: 5, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "// ❌\n// appendChild only works with Node objects\ncontainer.appendChild(divElement);\n\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"// ❌\n// appendChild only works with Node objects\ncontainer.append(divElement);\n\n"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-dom-node-append", Message: "Prefer `Element#append()` over `Node#appendChild()`.", Line: 3, Column: 1, EndLine: 3, EndColumn: 34, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
	})
}

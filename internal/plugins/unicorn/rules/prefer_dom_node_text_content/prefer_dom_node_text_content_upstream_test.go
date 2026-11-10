// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation; see LICENSE.
package prefer_dom_node_text_content_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_dom_node_text_content"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferDomNodeTextContentUpstream(t *testing.T) {
	// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-dom-node-text-content.js
	valid := []rule_tester.ValidTestCase{
		{FileName: "case.js", Code: `innerText;`},
		{FileName: "case.js", Code: `node.textContent;`},
		{FileName: "case.js", Code: `node[innerText];`},
		{FileName: "case.js", Code: `innerText = true;`},
		{FileName: "case.js", Code: `node['innerText'];`},
		{FileName: "case.js", Code: `innerText.textContent`},
		{FileName: "case.js", Code: `const [innerText] = node;`},
		{FileName: "case.js", Code: `[innerText] = node;`},
		{FileName: "case.js", Code: `const {[innerText]: text} = node;`},
		{FileName: "case.js", Code: `({[innerText]: text} = node);`},
		{FileName: "case.js", Code: `const foo = {innerText}`},
		{FileName: "case.js", Code: `const foo = {innerText: text}`},
		{Code: `interface Value {innerText: string} declare const value: Value; value.innerText;`},
		{Code: `interface Value {innerText: string} declare const value: Value; const {innerText} = value;`},
		{Code: `interface Value {innerText: string} declare const value: Value; let innerText; ({innerText} = value);`},
		{Code: `interface Value {innerText: string} function foo({innerText}: Value) { return innerText; }`},
		// Documentation examples.
		{FileName: "case.js", Code: `const text = foo.textContent;`},
		{FileName: "case.js", Code: `const {textContent} = foo;`},
		{FileName: "case.js", Code: `foo.textContent = '🦄';`},
	}
	var invalid []rule_tester.InvalidTestCase
	for _, testCase := range []struct {
		code, suggestion string
		column           int
	}{
		{`node.innerText;`, `node.textContent;`, 6},
		{`node?.innerText;`, `node?.textContent;`, 7},
		{`node.innerText = 'foo';`, `node.textContent = 'foo';`, 6},
		{`innerText.innerText;`, `innerText.textContent;`, 11},
		{`const {innerText} = node;`, `const {textContent: innerText} = node;`, 8},
		{`const {innerText,} = node;`, `const {textContent: innerText,} = node;`, 8},
		{`const {innerText: text} = node;`, `const {textContent: text} = node;`, 8},
		{`const {innerText = "default text"} = node;`, `const {textContent: innerText = "default text"} = node;`, 8},
		{`const {innerText: text = "default text"} = node;`, `const {textContent: text = "default text"} = node;`, 8},
		{`({innerText} = node);`, `({textContent: innerText} = node);`, 3},
		{`({innerText: text} = node);`, `({textContent: text} = node);`, 3},
		{`({innerText = "default text"} = node);`, `({textContent: innerText = "default text"} = node);`, 3},
		{`({innerText: text = "default text"} = node);`, `({textContent: text = "default text"} = node);`, 3},
		{`function foo({innerText}) {return innerText}`, `function foo({textContent: innerText}) {return innerText}`, 15},
		{`for (const [{innerText}] of elements);`, `for (const [{textContent: innerText}] of elements);`, 14},
		// Documentation examples.
		{`const text = foo.innerText;`, `const text = foo.textContent;`, 18},
		{`const {innerText} = foo;`, `const {textContent: innerText} = foo;`, 8},
		{`foo.innerText = '🦄';`, `foo.textContent = '🦄';`, 5},
	} {
		invalid = append(invalid, rule_tester.InvalidTestCase{
			FileName: "case.js", Code: testCase.code,
			Errors: []rule_tester.InvalidTestCaseError{textContentError(1, testCase.column, testCase.column+len("innerText"), testCase.suggestion)},
		})
	}
	testedRule := preferDomNodeTextContentWithoutTypes()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &testedRule, valid, invalid)
}

func preferDomNodeTextContentWithoutTypes() rule.Rule {
	testedRule := prefer_dom_node_text_content.PreferDomNodeTextContentRule
	// The upstream JavaScript and TypeScript cases both run without parser type services.
	run := testedRule.Run
	testedRule.Run = func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		ctx.TypeChecker = nil
		return run(ctx, options)
	}
	return testedRule
}

func textContentError(line, column, endColumn int, suggestion string) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "error", Message: "Prefer `.textContent` over `.innerText`.",
		Line: line, Column: column, EndLine: line, EndColumn: endColumn,
		Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: suggestion}},
	}
}

// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/test/no-typeof-undefined.js
package no_typeof_undefined_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_typeof_undefined"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const errorMessage = "Compare with `undefined` directly instead of using `typeof`."

func valid(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}}
}

func invalidFixed(code, output string) rule_tester.InvalidTestCase {
	start := strings.Index(code, "typeof")
	line := strings.Count(code[:start], "\n") + 1
	lastNewline := strings.LastIndex(code[:start], "\n")
	column := start + 1
	if lastNewline >= 0 {
		column = start - lastNewline
	}
	return rule_tester.InvalidTestCase{
		Code: code, FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"},
		Output: []string{output},
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "no-typeof-undefined/error",
			Message:   errorMessage,
			Line:      line, Column: column, EndLine: line, EndColumn: column + len("typeof"),
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
		}},
	}
}

func invalidGlobal(code, output string) rule_tester.InvalidTestCase {
	start := strings.Index(code, "typeof")
	line := strings.Count(code[:start], "\n") + 1
	lastNewline := strings.LastIndex(code[:start], "\n")
	column := start + 1
	if lastNewline >= 0 {
		column = start - lastNewline
	}
	return rule_tester.InvalidTestCase{
		Code: code, FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"},
		Options: []any{map[string]any{"checkGlobalVariables": true}},
		Output:  []string{},
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "no-typeof-undefined/error",
			Message:   errorMessage,
			Line:      line, Column: column, EndLine: line, EndColumn: column + len("typeof"),
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
				MessageId: "no-typeof-undefined/suggestion",
				Output:    output,
			}},
		}},
	}
}

func TestNoTypeofUndefinedUpstream(t *testing.T) {
	validCases := []rule_tester.ValidTestCase{
		valid("typeof a.b"),
		valid("typeof a.b > \"undefined\""),
		valid("a.b === \"undefined\""),
		valid("void a.b === \"undefined\""),
		valid("+a.b === \"undefined\""),
		valid("++a.b === \"undefined\""),
		valid("a.b++ === \"undefined\""),
		valid("foo === undefined"),
		valid("typeof a.b === \"string\""),
		valid("typeof foo === \"undefined\""),
		valid("foo = 2; typeof foo === \"undefined\""),
		valid("/* globals foo: readonly */ typeof foo === \"undefined\""),
		valid("/* globals globalThis: readonly */ typeof globalThis === \"undefined\""),
		valid("function parse() {\n\tswitch (typeof value === 'undefined') {}\n}"),
		valid("/* globals value: readonly */\nfunction parse() {\n\tswitch (typeof value === 'undefined') {}\n}"),
		valid("\"undefined\" === typeof a.b"),
		valid("const UNDEFINED = \"undefined\"; typeof a.b === UNDEFINED"),
		valid("typeof a.b === `undefined`"),
	}

	invalidCases := []rule_tester.InvalidTestCase{
		invalidFixed("typeof a.b === \"undefined\"", "a.b === undefined"),
		invalidFixed("typeof a.b !== \"undefined\"", "a.b !== undefined"),
		invalidFixed("typeof a.b == \"undefined\"", "a.b === undefined"),
		invalidFixed("typeof a.b != \"undefined\"", "a.b !== undefined"),
		invalidFixed("typeof a.b == 'undefined'", "a.b === undefined"),
		invalidFixed("let foo; typeof foo === \"undefined\"", "let foo; foo === undefined"),
		invalidFixed("const foo = 1; typeof foo === \"undefined\"", "const foo = 1; foo === undefined"),
		invalidFixed("var foo; typeof foo === \"undefined\"", "var foo; foo === undefined"),
		invalidFixed("var foo; var foo; typeof foo === \"undefined\"", "var foo; var foo; foo === undefined"),
		invalidFixed("for (const foo of bar) typeof foo === \"undefined\";", "for (const foo of bar) foo === undefined;"),
		invalidFixed("let foo;\nfunction bar() {\n\ttypeof foo === \"undefined\";\n}", "let foo;\nfunction bar() {\n\tfoo === undefined;\n}"),
		invalidFixed("function foo() {typeof foo === \"undefined\"}", "function foo() {foo === undefined}"),
		invalidFixed("function foo(bar) {typeof bar === \"undefined\"}", "function foo(bar) {bar === undefined}"),
		invalidFixed("function foo({bar}) {typeof bar === \"undefined\"}", "function foo({bar}) {bar === undefined}"),
		invalidFixed("function foo([bar]) {typeof bar === \"undefined\"}", "function foo([bar]) {bar === undefined}"),
		invalidFixed("typeof foo.bar === \"undefined\"", "foo.bar === undefined"),
		invalidFixed("import foo from 'foo';\ntypeof foo.bar === \"undefined\"", "import foo from 'foo';\nfoo.bar === undefined"),
		invalidFixed("foo\ntypeof [] === \"undefined\";", "foo\n;[] === undefined;"),
		invalidFixed("foo\ntypeof (a ? b : c) === \"undefined\";", "foo\n;(a ? b : c) === undefined;"),
		invalidFixed("function a() {\n\treturn typeof // comment\n\t\ta.b === 'undefined';\n}", "function a() {\n\treturn ( // comment\n\t\ta.b === undefined);\n}"),
		invalidFixed("function a() {\n\treturn (typeof // ReturnStatement argument is parenthesized\n\t\ta.b === 'undefined');\n}", "function a() {\n\treturn (// ReturnStatement argument is parenthesized\n\t\ta.b === undefined);\n}"),
		invalidFixed("function a() {\n\treturn (typeof // UnaryExpression is parenthesized\n\t\ta.b) === 'undefined';\n}", "function a() {\n\treturn (// UnaryExpression is parenthesized\n\t\ta.b) === undefined;\n}"),
		invalidFixed("function parse(value) {\n\tswitch (typeof value === 'undefined') {}\n}", "function parse(value) {\n\tswitch (value === undefined) {}\n}"),

		invalidGlobal("typeof undefinedVariableIdentifier === \"undefined\"", "undefinedVariableIdentifier === undefined"),
		invalidGlobal("typeof Array !== \"undefined\"", "Array !== undefined"),
		invalidGlobal("function parse() {\n\tswitch (typeof value === 'undefined') {}\n}", "function parse() {\n\tswitch (value === undefined) {}\n}"),
		invalidGlobal("/* globals value: readonly */\nfunction parse() {\n\tswitch (typeof value === 'undefined') {}\n}", "/* globals value: readonly */\nfunction parse() {\n\tswitch (value === undefined) {}\n}"),
	}

	if len(validCases) != 18 || len(invalidCases) != 27 {
		t.Fatalf("upstream coverage accounting changed: valid=%d invalid=%d", len(validCases), len(invalidCases))
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_typeof_undefined.NoTypeofUndefinedRule, validCases, invalidCases)
}

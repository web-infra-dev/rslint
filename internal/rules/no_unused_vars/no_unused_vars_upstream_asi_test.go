package no_unused_vars

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func upstreamASIError(name string, assigned bool, suggestionOutput *string) rule_tester.InvalidTestCaseError {
	action := "defined"
	if assigned {
		action = "assigned a value"
	}
	expected := rule_tester.InvalidTestCaseError{
		MessageId: "unusedVar",
		Message:   "'" + name + "' is " + action + " but never used.",
	}
	if suggestionOutput != nil {
		expected.Suggestions = []rule_tester.InvalidTestCaseSuggestion{{
			MessageId: "removeVar",
			Output:    *suggestionOutput,
		}}
	}
	return expected
}

func upstreamASISuggestion(output string) *string {
	return &output
}

// TestNoUnusedVarsUpstreamASISuggestions covers the no-unused-vars cases added
// upstream after the full v10.7.0 fixture, through ESLint v10.9.1. They pin
// both sides of the ASI safety boundary for removal suggestions.
func TestNoUnusedVarsUpstreamASISuggestions(t *testing.T) {
	module := rule.LanguageOptions{SourceType: "module"}
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnusedVarsRule,
		nil,
		[]rule_tester.InvalidTestCase{
			{
				Code: "if (true) {}\nfunction unused() {}\n(console.log)()",
				Errors: []rule_tester.InvalidTestCaseError{
					upstreamASIError("unused", false, upstreamASISuggestion("if (true) {}\n\n(console.log)()")),
				},
			},
			{
				Code:            "export class A {}\nfunction unused() {}\n(console.log)()",
				LanguageOptions: module,
				Errors: []rule_tester.InvalidTestCaseError{
					upstreamASIError("unused", false, upstreamASISuggestion("export class A {}\n\n(console.log)()")),
				},
			},
			{
				Code: "switch (1) {}\nfunction unused() {}\n(console.log)()",
				Errors: []rule_tester.InvalidTestCaseError{
					upstreamASIError("unused", false, upstreamASISuggestion("switch (1) {}\n\n(console.log)()")),
				},
			},
			{
				Code:            "export const f = function() {}\nfunction unused() {}\n(console.log)()",
				LanguageOptions: module,
				Errors:          []rule_tester.InvalidTestCaseError{upstreamASIError("unused", false, nil)},
			},
			{
				Code:            "export const obj = { a: 1 }\nfunction unused() {}\n(console.log)()",
				LanguageOptions: module,
				Errors:          []rule_tester.InvalidTestCaseError{upstreamASIError("unused", false, nil)},
			},
			{
				Code:            "const x = 1\nfunction unused() {}\n(console.log)()",
				LanguageOptions: module,
				Errors: []rule_tester.InvalidTestCaseError{
					upstreamASIError("x", true, upstreamASISuggestion("\nfunction unused() {}\n(console.log)()")),
					upstreamASIError("unused", false, nil),
				},
			},
			{
				Code:            "const x = 1\nfunction unused() {}\n/regex/.test('a')",
				LanguageOptions: module,
				Errors: []rule_tester.InvalidTestCaseError{
					upstreamASIError("x", true, upstreamASISuggestion("\nfunction unused() {}\n/regex/.test('a')")),
					upstreamASIError("unused", false, nil),
				},
			},
			{
				Code:            "const x = 1\nclass Unused {}\n(console.log)()",
				LanguageOptions: module,
				Errors: []rule_tester.InvalidTestCaseError{
					upstreamASIError("x", true, upstreamASISuggestion("\nclass Unused {}\n(console.log)()")),
					upstreamASIError("Unused", false, nil),
				},
			},
			{
				Code:            "const x = 1,\n      y = () => {}\n(console.log)()",
				LanguageOptions: module,
				Errors: []rule_tester.InvalidTestCaseError{
					upstreamASIError("x", true, upstreamASISuggestion("const \n      y = () => {}\n(console.log)()")),
					upstreamASIError("y", true, nil),
				},
			},
			{
				Code:            "const { x } = obj,\n      y = () => {}\n(console.log)()",
				LanguageOptions: module,
				Errors: []rule_tester.InvalidTestCaseError{
					upstreamASIError("x", true, upstreamASISuggestion("const \n      y = () => {}\n(console.log)()")),
					upstreamASIError("y", true, nil),
				},
			},
			{
				Code: "var x = 1;\nfunction unused() {}\nvar z = 3",
				Errors: []rule_tester.InvalidTestCaseError{
					upstreamASIError("x", true, upstreamASISuggestion("\nfunction unused() {}\nvar z = 3")),
					upstreamASIError("unused", false, upstreamASISuggestion("var x = 1;\n\nvar z = 3")),
					upstreamASIError("z", true, upstreamASISuggestion("var x = 1;\nfunction unused() {}\n")),
				},
			},
			{
				Code: "{ console.log(); class Unused {} }",
				Errors: []rule_tester.InvalidTestCaseError{
					upstreamASIError("Unused", false, upstreamASISuggestion("{ console.log();  }")),
				},
			},
			{
				Code: "console.log();\nfunction unused() {}",
				Errors: []rule_tester.InvalidTestCaseError{
					upstreamASIError("unused", false, upstreamASISuggestion("console.log();\n")),
				},
			},
			{
				Code: "console.log();\nclass Unused {}",
				Errors: []rule_tester.InvalidTestCaseError{
					upstreamASIError("Unused", false, upstreamASISuggestion("console.log();\n")),
				},
			},
			{
				Code:   "const before = true\nfunction unused() {}\nif (before) {}",
				Errors: []rule_tester.InvalidTestCaseError{upstreamASIError("unused", false, nil)},
			},
			{
				Code:            "/* comment */\nfunction unused() {}\n(a) => {}",
				LanguageOptions: module,
				Errors: []rule_tester.InvalidTestCaseError{
					upstreamASIError("unused", false, upstreamASISuggestion("/* comment */\n\n(a) => {}")),
					upstreamASIError("a", false, upstreamASISuggestion("/* comment */\nfunction unused() {}\n() => {}")),
				},
			},
			{
				Code:            "const x = 1;\nfunction unused() {}\n/* comment */\n(a) => {}",
				LanguageOptions: module,
				Errors: []rule_tester.InvalidTestCaseError{
					upstreamASIError("x", true, upstreamASISuggestion("\nfunction unused() {}\n/* comment */\n(a) => {}")),
					upstreamASIError("unused", false, upstreamASISuggestion("const x = 1;\n\n/* comment */\n(a) => {}")),
					upstreamASIError("a", false, upstreamASISuggestion("const x = 1;\nfunction unused() {}\n/* comment */\n() => {}")),
				},
			},
			{
				Code:            "const x = 1,\n      y = () => {}\n() => {};",
				LanguageOptions: module,
				Errors: []rule_tester.InvalidTestCaseError{
					upstreamASIError("x", true, upstreamASISuggestion("const \n      y = () => {}\n() => {};")),
					upstreamASIError("y", true, nil),
				},
			},
			{
				Code: "var a = 1, b = 2; console.log(a)",
				Errors: []rule_tester.InvalidTestCaseError{
					upstreamASIError("b", true, upstreamASISuggestion("var a = 1; console.log(a)")),
				},
			},
		},
	)
}

// TestNoUnusedVarsASIAdversarial attacks boundaries adjacent to the upstream
// cases: the documented variable-after-function shape, expression braces,
// string directives, EOF, and a retained semicolon in a multi-declaration.
func TestNoUnusedVarsASIAdversarial(t *testing.T) {
	module := rule.LanguageOptions{SourceType: "module"}
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnusedVarsRule,
		nil,
		[]rule_tester.InvalidTestCase{
			{
				Code: "function used() {}\nvar Type2Parser = function type2Parser() {};\nused();",
				Errors: []rule_tester.InvalidTestCaseError{
					upstreamASIError("Type2Parser", true, upstreamASISuggestion("function used() {}\n\nused();")),
				},
			},
			{
				Code: "function used() {}\nfunction unused() {}\n(console.log)();\nused();",
				Errors: []rule_tester.InvalidTestCaseError{
					upstreamASIError("unused", false, upstreamASISuggestion("function used() {}\n\n(console.log)();\nused();")),
				},
			},
			{
				Code:            "export const arrow = () => {}\nfunction unused() {}\n(console.log)()",
				LanguageOptions: module,
				Errors:          []rule_tester.InvalidTestCaseError{upstreamASIError("unused", false, nil)},
			},
			{
				Code:            "export const Expression = class {}\nfunction unused() {}\n(console.log)()",
				LanguageOptions: module,
				Errors:          []rule_tester.InvalidTestCaseError{upstreamASIError("unused", false, nil)},
			},
			{
				Code: "function used() {}\nvar unused = 1;\n'not a directive';\nused();",
				Errors: []rule_tester.InvalidTestCaseError{
					upstreamASIError("unused", true, nil),
				},
			},
			{
				Code:            "export const expression = function() {}\nvar unused = 1",
				LanguageOptions: module,
				Errors: []rule_tester.InvalidTestCaseError{
					upstreamASIError("unused", true, upstreamASISuggestion("export const expression = function() {}\n")),
				},
			},
			{
				Code:            "export const arrow = () => {}\nvar used = 1, unused = 2; consume(used)",
				LanguageOptions: module,
				Errors: []rule_tester.InvalidTestCaseError{
					upstreamASIError("unused", true, upstreamASISuggestion("export const arrow = () => {}\nvar used = 1; consume(used)")),
				},
			},
		},
	)
}

package no_empty_function

import (
	"reflect"
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoEmptyFunctionReportedPayloads(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoEmptyFunctionRule,
		nil,
		[]rule_tester.InvalidTestCase{
			noEmptyPayloadCase(`class C { registerHooks = () => {}; }`, "method 'registerHooks'", nil),
			noEmptyPayloadCase(`const view = { closeMessagePanel: () => {} };`, "method 'closeMessagePanel'", nil),
			noEmptyPayloadCase(`const runAfterHideLoading: any = () => {};`, "arrow function", nil),
			noEmptyPayloadCase(`class C { static onCheckResultChange = () => {}; }`, "static method 'onCheckResultChange'", nil),
			noEmptyPayloadCase(`const foo = function() {};`, "function", nil),
			noEmptyPayloadCase(`const foo = function bar() {};`, "function 'bar'", nil),
			noEmptyPayloadCase(`const foo = async () => {};`, "async arrow function", nil),
			noEmptyPayloadCase(`class C { static constructor() {} }`, "static method 'constructor'", nil),
			noEmptyPayloadCase(`class C { accessor field = () => {}; }`, "arrow function", nil),
			noEmptyPayloadCase(`class C { static accessor field = () => {}; }`, "arrow function", nil),
			noEmptyPayloadCase(`class C { accessor field = function() {}; }`, "function", nil),
			noEmptyPayloadCase(`class C { accessor field = function named() {}; }`, "function 'named'", nil),
		},
	)
}

func TestNoEmptyFunctionAllowKindsStayIsolated(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoEmptyFunctionRule,
		[]rule_tester.ValidTestCase{
			{Code: `function f() {}`, Options: noEmptyAllow("functions")},
			{Code: `const f = async () => {};`, Options: noEmptyAllow("arrowFunctions")},
			{Code: `class C { async f() {} }`, Options: noEmptyAllow("asyncMethods")},
			{Code: `class C { *f() {} }`, Options: noEmptyAllow("generatorMethods")},
			{Code: `class C { private constructor() {} }`, Options: noEmptyAllow("private-constructors")},
			{Code: `class C { protected constructor() {} }`, Options: noEmptyAllow("protected-constructors")},
			{Code: `class C { static constructor() {} }`, Options: noEmptyAllow("methods")},
			{Code: `class C { accessor field = () => {}; }`, Options: noEmptyAllow("arrowFunctions")},
			{Code: `class C { accessor field = function() {}; }`, Options: noEmptyAllow("functions")},
		},
		[]rule_tester.InvalidTestCase{
			noEmptyPayloadCase(`const f = () => {};`, "arrow function", noEmptyAllow("functions")),
			noEmptyPayloadCase(`const f = async () => {};`, "async arrow function", noEmptyAllow("asyncFunctions")),
			noEmptyPayloadCase(`class C { async f() {} }`, "async method 'f'", noEmptyAllow("asyncFunctions")),
			noEmptyPayloadCase(`class C { *f() {} }`, "generator method 'f'", noEmptyAllow("generatorFunctions")),
			noEmptyPayloadCase(`class C { f() {} }`, "method 'f'", noEmptyAllow("functions")),
			noEmptyPayloadCase(`class C { get f() {} }`, "getter 'f'", noEmptyAllow("functions")),
			noEmptyPayloadCase(`class C { constructor() {} }`, "constructor", noEmptyAllow("functions")),
			noEmptyPayloadCase(`class C { field = () => {} }`, "method 'field'", noEmptyAllow("methods")),
			noEmptyPayloadCase(`class C { static constructor() {} }`, "static method 'constructor'", noEmptyAllow("constructors")),
		},
	)
}

func TestNoEmptyFunctionIgnoresHostedJSDocSyntax(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoEmptyFunctionRule,
		nil,
		[]rule_tester.InvalidTestCase{
			hostedJSDocNoEmptyPayloadCase(
				"class C {\n  /** @override */\n  method() {}\n}",
				"method 'method'",
				noEmptyAllow("overrideMethods"),
			),
			hostedJSDocNoEmptyPayloadCase(
				"class C {\n  /** @private */\n  constructor() {}\n}",
				"constructor",
				noEmptyAllow("private-constructors"),
			),
			hostedJSDocNoEmptyPayloadCase(
				"class C {\n  /** @protected */\n  constructor() {}\n}",
				"constructor",
				noEmptyAllow("protected-constructors"),
			),
			hostedJSDocNoEmptyPayloadCase(
				"const value = { callback: /** @type {Function} */ (() => {}) };",
				"method 'callback'",
				nil,
			),
			hostedJSDocNoEmptyPayloadCase(
				"class C { static #callback = /** @type {Function} */ (() => {}); }",
				"static private method #callback",
				nil,
			),
		},
	)
}

func hostedJSDocNoEmptyPayloadCase(code, name string, options any) rule_tester.InvalidTestCase {
	testCase := noEmptyPayloadCase(code, name, options)
	testCase.FileName = "file.mjs"
	testCase.TSConfig = "tsconfig.allow-js.json"
	return testCase
}

func TestNormalizeConstructorOptions(t *testing.T) {
	options := []any{map[string]any{
		"allow": []any{"functions", "private-constructors", "protected-constructors"},
	}}
	wantOriginal := []any{map[string]any{
		"allow": []any{"functions", "private-constructors", "protected-constructors"},
	}}
	normalized := normalizeConstructorOptions(options)

	if !reflect.DeepEqual(options, wantOriginal) {
		t.Fatalf("normalizeConstructorOptions mutated its input: %#v", options)
	}
	wantNormalized := []any{map[string]any{
		"allow": []any{"functions", "privateConstructors", "protectedConstructors"},
	}}
	if !reflect.DeepEqual(normalized, wantNormalized) {
		t.Fatalf("normalizeConstructorOptions() = %#v, want %#v", normalized, wantNormalized)
	}

	unchanged := []any{map[string]any{"allow": []any{"functions"}}}
	if got := normalizeConstructorOptions(unchanged); &got[0] != &unchanged[0] {
		t.Fatal("normalization copied options that needed no translation")
	}
}

func TestNoEmptyFunctionConstructorOptionSchema(t *testing.T) {
	for _, option := range []string{"private-constructors", "protected-constructors"} {
		if err := NoEmptyFunctionRule.Schema.Validate(noEmptyAllow(option)); err != nil {
			t.Errorf("schema rejected typescript-eslint option %q: %v", option, err)
		}
	}
	for _, option := range []string{"privateConstructors", "protectedConstructors"} {
		if err := NoEmptyFunctionRule.Schema.Validate(noEmptyAllow(option)); err == nil {
			t.Errorf("schema accepted ESLint-only option %q", option)
		}
	}
}

func noEmptyAllow(kinds ...string) []any {
	allow := make([]any, len(kinds))
	for i, kind := range kinds {
		allow[i] = kind
	}
	return []any{map[string]any{"allow": allow}}
}

func noEmptyPayloadCase(code string, name string, options any) rule_tester.InvalidTestCase {
	bodyStart := strings.Index(code, "{}")
	if bodyStart < 0 {
		panic("no-empty-function payload case has no empty body")
	}
	line, column := noEmptyLineColumn(code, bodyStart)
	endLine, endColumn := noEmptyLineColumn(code, bodyStart+2)
	return rule_tester.InvalidTestCase{
		Code:    code,
		Options: options,
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "unexpected",
			Message:   "Unexpected empty " + name + ".",
			Line:      line,
			Column:    column,
			EndLine:   endLine,
			EndColumn: endColumn,
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
				MessageId: "suggestComment",
				Output:    code[:bodyStart] + "{ /* empty */ }" + code[bodyStart+2:],
			}},
		}},
	}
}

func withNoEmptySuggestions(cases []rule_tester.InvalidTestCase) []rule_tester.InvalidTestCase {
	for caseIndex := range cases {
		searchStart := 0
		for errorIndex := range cases[caseIndex].Errors {
			relativeStart := strings.Index(cases[caseIndex].Code[searchStart:], "{}")
			if relativeStart < 0 {
				panic("no-empty-function test case has fewer empty bodies than diagnostics")
			}
			bodyStart := searchStart + relativeStart
			expectedError := &cases[caseIndex].Errors[errorIndex]
			expectedError.EndLine, expectedError.EndColumn = noEmptyLineColumn(cases[caseIndex].Code, bodyStart+2)
			expectedError.Suggestions = []rule_tester.InvalidTestCaseSuggestion{{
				MessageId: "suggestComment",
				Output: cases[caseIndex].Code[:bodyStart] +
					"{ /* empty */ }" +
					cases[caseIndex].Code[bodyStart+2:],
			}}
			searchStart = bodyStart + 2
		}
	}
	return cases
}

func noEmptyLineColumn(code string, offset int) (int, int) {
	line, column := 1, 1
	for index := range offset {
		if code[index] == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}
	return line, column
}

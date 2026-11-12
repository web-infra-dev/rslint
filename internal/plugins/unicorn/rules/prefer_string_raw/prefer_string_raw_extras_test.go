package prefer_string_raw_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_string_raw"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferStringRawExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_string_raw.PreferStringRawRule,
		[]rule_tester.ValidTestCase{
			{Code: "const x = (String?.raw)`a`;", FileName: "case.js"},
			{Code: "class C { #raw; use() { return this.#raw`a`; } }", FileName: "case.js"},
			{Code: `expect(x).toMatchInlineSnapshot(('a\\b'));`, FileName: "case.js"},
			{
				Code:            "with (scope) String.raw`plain`;",
				FileName:        "case.js",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:     `const π = "🙂\\path";`,
				FileName: "case.js",
				Output:   []string{"const π = String.raw`🙂\\path`;"},
				Errors:   []rule_tester.InvalidTestCaseError{preferRawError(1, 11, 1, 21)},
			},
			{
				Code:     "const s = `${a}first\\\\part${b}`;",
				FileName: "case.js",
				Output:   []string{"const s = String.raw`${a}first\\part${b}`;"},
				Errors:   []rule_tester.InvalidTestCaseError{preferRawError(1, 11, 1, 32)},
			},
			{
				// Only the changed part normalizes CRLF to LF.
				Code:     "const s = `line\\\\path\r\n${x}keep\r\n`;",
				FileName: "case.js",
				Output:   []string{"const s = String.raw`line\\path\n${x}keep\r\n`;"},
				Errors:   []rule_tester.InvalidTestCaseError{preferRawError(1, 11, 3, 2)},
			},
			{
				Code:     "function* scan() { yield* String.raw\n`abc`; }",
				FileName: "case.js",
				Output:   []string{"function* scan() { yield* \n'abc'; }"},
				Errors:   []rule_tester.InvalidTestCaseError{unnecessaryRawError(1, 27, 1, 37)},
			},
			{
				Code:     "const a = ((/* keep */ String.raw)) `a`;",
				FileName: "case.js",
				Output:   []string{"const a = /* keep */  'a';"},
				Errors:   []rule_tester.InvalidTestCaseError{unnecessaryRawError(1, 24, 1, 34)},
			},
			{
				Code:     "if (enabled) String.raw`prefix ${name}`;",
				FileName: "case.js",
				Output:   []string{"if (enabled) `prefix ${name}`;"},
				Errors:   []rule_tester.InvalidTestCaseError{unnecessaryRawError(1, 14, 1, 24)},
			},
			{
				Code:     "function f(a=1) {String.raw`use strict`;}",
				FileName: "case.js",
				Output:   []string{"function f(a=1) {`use strict`;}"},
				Errors:   []rule_tester.InvalidTestCaseError{unnecessaryRawError(1, 18, 1, 28)},
			},
			{
				Code:     "function f(a=1) {String.raw`hello`; 'use strict';}",
				FileName: "case.js",
				Output:   []string{"function f(a=1) {`hello`; 'use strict';}"},
				Errors:   []rule_tester.InvalidTestCaseError{unnecessaryRawError(1, 18, 1, 28)},
			},
			{
				Code:            "String.raw`use strict`;",
				FileName:        "case.js",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Output:          []string{"`use strict`;"},
				Errors:          []rule_tester.InvalidTestCaseError{unnecessaryRawError(1, 1, 1, 11)},
			},
			{
				Code:     `function f(String) { return 'a\\b'; }`,
				FileName: "case.js",
				Errors:   []rule_tester.InvalidTestCaseError{preferRawError(1, 29, 1, 35)},
			},
			{
				Code:            `var path = 'C:\\temp';`,
				FileName:        "case.js",
				LanguageOptions: rule.LanguageOptions{SourceType: "script", ECMAVersion: 5},
				Errors:          []rule_tester.InvalidTestCaseError{preferRawError(1, 12, 1, 22)},
			},
			{
				Code:            `with (scope) path = 'a\\b';`,
				FileName:        "case.js",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Errors:          []rule_tester.InvalidTestCaseError{preferRawError(1, 21, 1, 27)},
			},
			{
				Code:     "const path = 'a\\\\b'; String.raw`abc`;",
				FileName: "case.js",
				Globals:  map[string]any{"String": "off"},
				Errors:   []rule_tester.InvalidTestCaseError{preferRawError(1, 14, 1, 20)},
			},
			{
				Code:     "const x = String.raw<string>`abc`;",
				FileName: "case.ts",
				Errors:   []rule_tester.InvalidTestCaseError{unnecessaryRawError(1, 11, 1, 21)},
			},
			{
				Code:     `const view = <div path={'a\\b'} />;`,
				FileName: "case.jsx",
				Output:   []string{"const view = <div path={String.raw`a\\b`} />;"},
				Errors:   []rule_tester.InvalidTestCaseError{preferRawError(1, 25, 1, 31)},
			},
			{
				// Optional and computed calls do not match the snapshot exemption.
				Code: `expect(obj)?.toMatchInlineSnapshot('a\\b');
expect(obj).toMatchInlineSnapshot?.('a\\b');
expect?.(obj).toMatchInlineSnapshot('a\\b');
expect(obj)['toMatchInlineSnapshot']('a\\b');`,
				FileName: "case.js",
				Output: []string{"expect(obj)?.toMatchInlineSnapshot(String.raw`a\\b`);\n" +
					"expect(obj).toMatchInlineSnapshot?.(String.raw`a\\b`);\n" +
					"expect?.(obj).toMatchInlineSnapshot(String.raw`a\\b`);\n" +
					"expect(obj)['toMatchInlineSnapshot'](String.raw`a\\b`);"},
				Errors: []rule_tester.InvalidTestCaseError{
					preferRawError(1, 36, 1, 42), preferRawError(2, 37, 2, 43),
					preferRawError(3, 37, 3, 43), preferRawError(4, 38, 4, 44),
				},
			},
			{
				Code:     `const {'a\\b': name, field = 'c\\d'} = input;`,
				FileName: "case.js",
				Output:   []string{"const {'a\\\\b': name, field = String.raw`c\\d`} = input;"},
				Errors:   []rule_tester.InvalidTestCaseError{preferRawError(1, 30, 1, 36)},
			},
			{
				Code:     "const path = `a\\\\b``suffix`;",
				FileName: "case.js",
				Output:   []string{"const path = String.raw`a\\b``suffix`;"},
				Errors:   []rule_tester.InvalidTestCaseError{preferRawError(1, 14, 1, 20)},
			},
			{
				Code:     `class C { static { 'a\\b'; } }`,
				FileName: "case.js",
				Output:   []string{"class C { static { String.raw`a\\b`; } }"},
				Errors:   []rule_tester.InvalidTestCaseError{preferRawError(1, 20, 1, 26)},
			},
		},
	)
}

func TestPreferStringRawSyntaxExtras(t *testing.T) {
	r := preferStringRawWithoutTypes()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &r,
		[]rule_tester.ValidTestCase{
			{Code: `const path = 'C:\\temp' as (const);`, FileName: "case.ts"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:     `const path: ((string)) = 'C:\\temp';`,
				FileName: "case.ts",
				Output:   []string{"const path: ((string)) = String.raw`C:\\temp`;"},
				Errors:   []rule_tester.InvalidTestCaseError{preferRawError(1, 26, 1, 36)},
			},
		},
	)
}

func TestPreferStringRawEditDemand(t *testing.T) {
	const source = `let mutable = 'a\\b'; const literal = 'c\\d';`
	program, file, err := rule_tester.NewProgramHelper(fixtures.GetRootDir()).CreateTestProgram(source, "edit-demand.ts", "tsconfig.json")
	if err != nil {
		t.Fatal(err)
	}
	r := preferStringRawWithoutTypes()
	run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: lintprogram.NewFromCompiler(program),
			File:    file.FileName(),
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{
					Name:     r.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return r.Run(ctx, nil)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, d)
			}},
		})
		if len(diagnostics) != 2 {
			t.Fatalf("demand %d: got %d diagnostics, want 2", demand, len(diagnostics))
		}
		return diagnostics
	}

	all := run(rule.EditDemandAll)
	if len(all[0].Fixes()) == 0 || all[0].Suggestions != nil || all[1].FixesPtr != nil ||
		all[1].Suggestions == nil || len(*all[1].Suggestions) != 1 {
		t.Fatalf("expected one autofix and one suggestion: %+v", all)
	}
	suggestion := (*all[1].Suggestions)[0]
	if suggestion.Message.Description != "Use `String.raw`. This may widen the string literal type to `string`." {
		t.Errorf("unexpected suggestion message %q", suggestion.Message.Description)
	}
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion} {
		for i, diagnostic := range run(demand) {
			expected := all[i]
			if demand != rule.EditDemandAutofix {
				expected.FixesPtr = nil
			}
			if demand != rule.EditDemandSuggestion {
				expected.Suggestions = nil
			}
			if diagnostic.Range != expected.Range || !reflect.DeepEqual(diagnostic.Message, expected.Message) ||
				!reflect.DeepEqual(diagnostic.FixesPtr, expected.FixesPtr) || !reflect.DeepEqual(diagnostic.Suggestions, expected.Suggestions) {
				t.Errorf("demand %d changed diagnostic %d: %+v, want %+v", demand, i, diagnostic, expected)
			}
		}
	}
}

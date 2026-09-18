// Additional AST, scope, and edit-demand coverage checked against Unicorn v75.0.0.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/test/no-useless-error-capture-stack-trace.js
package no_useless_error_capture_stack_trace_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_useless_error_capture_stack_trace"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUselessErrorCaptureStackTraceExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_useless_error_capture_stack_trace.NoUselessErrorCaptureStackTraceRule, []rule_tester.ValidTestCase{
		{Code: "class C extends Error {constructor(){async function nested(){Error.captureStackTrace(this,C);}}}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class C extends Error {constructor(){const f = function(){Error.captureStackTrace(this,C);};}}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class C extends Error {constructor(){({method(){Error.captureStackTrace(this,C);}});}}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class C extends Error {constructor(){({get value(){Error.captureStackTrace(this,C);}});}}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class C extends Error {static constructor(){Error.captureStackTrace(this,C);}}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class C extends Error {[\"constructor\"](){Error.captureStackTrace(this,C);}}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class C extends Error {constructor(){class D extends Error {constructor(){Error.captureStackTrace(this,C);}}}}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const C = class D extends Error {constructor(){Error.captureStackTrace(this,C);}};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const C = class extends Error {constructor(){Error.captureStackTrace(this,C);}};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class C extends globalThis.Error {constructor(){Error.captureStackTrace(this,C);}}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class C extends Error {constructor(Error){Error.captureStackTrace(this,C);}}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class C extends Error {constructor(){let C; Error.captureStackTrace(this,C);}}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class C extends Error {constructor(){globalThis.Error.captureStackTrace(this,C);}}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class C extends Error {constructor(){Error[\"captureStackTrace\"](this,C);}}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class C extends Error {constructor(){(Error?.captureStackTrace)(this,C);}}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class C extends Error {constructor(){Error.captureStackTrace(this,this[\"constructor\"]);}}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class C extends Error {constructor(){Error.captureStackTrace(this,this?.constructor);}}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class C extends (Error as ErrorConstructor) {constructor(){Error.captureStackTrace(this,C);}}", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class C extends Error {constructor(){Error.captureStackTrace(this,C as typeof C);}}", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class C extends RangeError {constructor(){Error.captureStackTrace(this,C);}}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"Error": "off"}},
		{Code: "class C extends RangeError {constructor(){Error.captureStackTrace(this,C);}}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"RangeError": "off"}},
		{Code: "type Error = unknown; class C extends RangeError {constructor(){Error.captureStackTrace(this,C);}}", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "script"}},
	}, []rule_tester.InvalidTestCase{
		{Code: "class C extends Error {constructor(){(() => {Error.captureStackTrace(this,C);})();}}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"class C extends Error {constructor(){(() => {})();}}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 1, Column: 46, EndLine: 1, EndColumn: 77, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "class C extends Error {constructor(){const f = async () => {Error.captureStackTrace(this,C);};}}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"class C extends Error {constructor(){const f = async () => {};}}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 1, Column: 61, EndLine: 1, EndColumn: 92, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "class C extends Error {constructor(){({[Error.captureStackTrace(this,C)](){}});}}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 1, Column: 41, EndLine: 1, EndColumn: 72, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "class C extends Error {constructor(x=Error.captureStackTrace(this,new.target)) {}}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 1, Column: 38, EndLine: 1, EndColumn: 78, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "class C extends Error {constructor(){class D extends Error {constructor(){Error.captureStackTrace(this,D);}}}}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"class C extends Error {constructor(){class D extends Error {constructor(){}}}}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 1, Column: 75, EndLine: 1, EndColumn: 106, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const C = class D extends Error {constructor(){Error.captureStackTrace(this,D);}};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const C = class D extends Error {constructor(){}};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 1, Column: 48, EndLine: 1, EndColumn: 79, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "class C extends (Error) {constructor(){Error.captureStackTrace(this,C);}}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"class C extends (Error) {constructor(){}}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 1, Column: 40, EndLine: 1, EndColumn: 71, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "class C extends Error {constructor(){(Error.captureStackTrace)?.(this,C);}}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"class C extends Error {constructor(){}}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 1, Column: 38, EndLine: 1, EndColumn: 73, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "class C extends Error {constructor(){Error.captureStackTrace((this),(this).constructor);}}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"class C extends Error {constructor(){}}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 1, Column: 38, EndLine: 1, EndColumn: 88, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "class C extends Error {constructor(){Error.captureStackTrace(this,(new.target));}}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"class C extends Error {constructor(){}}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 1, Column: 38, EndLine: 1, EndColumn: 80, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "class C extends Error {constructor(){ /* before */ Error.captureStackTrace(this,C); /* after */ }}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"class C extends Error {constructor(){ /* before */  /* after */ }}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 1, Column: 52, EndLine: 1, EndColumn: 83, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "class C extends Error {constructor(){Error.captureStackTrace(/* target */ this,C);}}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"class C extends Error {constructor(){}}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 1, Column: 38, EndLine: 1, EndColumn: 82, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "class C extends Error {constructor(){switch(x){case 0: Error.captureStackTrace(this,C);}}}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 1, Column: 56, EndLine: 1, EndColumn: 87, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "class C extends Error {constructor(){for(;;) Error.captureStackTrace(this,C);}}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 1, Column: 46, EndLine: 1, EndColumn: 77, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "class C extends Error {constructor(){return Error.captureStackTrace(this,C);}}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 1, Column: 45, EndLine: 1, EndColumn: 76, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "\"😀\"; class C extends TypeError {constructor(){Error.captureStackTrace(this,C);}}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"\"😀\"; class C extends TypeError {constructor(){}}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 1, Column: 48, EndLine: 1, EndColumn: 79, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "class C extends Error<string> {constructor(){Error.captureStackTrace(this,C);}}", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"class C extends Error<string> {constructor(){}}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 1, Column: 46, EndLine: 1, EndColumn: 77, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "class C extends Error {constructor(){Error.captureStackTrace(this,C);}} interface C {}", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"class C extends Error {constructor(){}} interface C {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 1, Column: 38, EndLine: 1, EndColumn: 69, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "class C extends Error {constructor(){Error.captureStackTrace(this,C);}} namespace C {}", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"class C extends Error {constructor(){}} namespace C {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 1, Column: 38, EndLine: 1, EndColumn: 69, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "type Error = unknown; class C extends RangeError {constructor(){Error.captureStackTrace(this,C);}}", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"type Error = unknown; class C extends RangeError {constructor(){}}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-useless-error-capture-stack-trace/error", Message: "Unnecessary `Error.captureStackTrace(…)` call.", Line: 1, Column: 65, EndLine: 1, EndColumn: 96, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
	})
}

func TestNoUselessErrorCaptureStackTraceArtifactsFollowDemand(t *testing.T) {
	for _, testCase := range []struct {
		source, output string
		options        []any
		suggestions    []struct{ messageID, description, output string }
	}{
		{source: "class MyError extends Error {\n\tconstructor() {\n\t\tError.captureStackTrace(this, MyError);\n\t}\n}", output: "class MyError extends Error {\n\tconstructor() {\n\t\t\n\t}\n}", options: []any{}},
		{source: "class MyError extends Error {\n\tconstructor() {\n\t\tif (a) Error.captureStackTrace(this, MyError)\n\t}\n}", output: "class MyError extends Error {\n\tconstructor() {\n\t\tif (a) Error.captureStackTrace(this, MyError)\n\t}\n}", options: []any{}},
	} {
		t.Run(testCase.source, func(t *testing.T) {
			helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
			program, sourceFile, err := helper.CreateTestProgram(testCase.source, "edit-demand.js", "tsconfig.json")
			if err != nil {
				t.Fatal(err)
			}
			diagnostics := make(map[rule.EditDemand]rule.RuleDiagnostic, 4)
			for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
				var got []rule.RuleDiagnostic
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program: lintprogram.NewFromCompiler(program), File: sourceFile.FileName(),
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{Name: no_useless_error_capture_stack_trace.NoUselessErrorCaptureStackTraceRule.Name, Severity: rule.SeverityError,
							Run: func(ctx rule.RuleContext) rule.RuleListeners {
								return no_useless_error_capture_stack_trace.NoUselessErrorCaptureStackTraceRule.Run(ctx, testCase.options)
							},
						}}
					},
					Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { got = append(got, d) }},
				})
				if len(got) != 1 {
					t.Fatalf("demand %d: got %d diagnostics", demand, len(got))
				}
				diagnostics[demand] = got[0]
			}
			baseline := diagnostics[rule.EditDemandNone]
			all := diagnostics[rule.EditDemandAll]
			for demand, diagnostic := range diagnostics {
				if diagnostic.Range != baseline.Range || !reflect.DeepEqual(diagnostic.Message, baseline.Message) || diagnostic.Severity != baseline.Severity {
					t.Errorf("demand %d changed diagnostic identity", demand)
				}
				wantFix := testCase.output != testCase.source && (demand == rule.EditDemandAutofix || demand == rule.EditDemandAll)
				if (diagnostic.FixesPtr != nil) != wantFix {
					t.Errorf("demand %d: unexpected autofix artifacts", demand)
				}
				if wantFix && !reflect.DeepEqual(diagnostic.FixesPtr, all.FixesPtr) {
					t.Errorf("demand %d changed autofix artifacts", demand)
				}
				expectedOutput := testCase.source
				if wantFix {
					expectedOutput = testCase.output
				}
				output, _, fixed := linter.ApplyRuleFixes(testCase.source, []rule.RuleDiagnostic{diagnostic})
				if output != expectedOutput || fixed != wantFix {
					t.Errorf("demand %d: unexpected autofix %q", demand, output)
				}
				wantSuggestions := len(testCase.suggestions) > 0 && (demand == rule.EditDemandSuggestion || demand == rule.EditDemandAll)
				if !wantSuggestions {
					if diagnostic.Suggestions != nil {
						t.Errorf("demand %d produced suggestions without demand", demand)
					}
					continue
				}
				if diagnostic.Suggestions == nil || len(*diagnostic.Suggestions) != len(testCase.suggestions) || !reflect.DeepEqual(diagnostic.Suggestions, all.Suggestions) {
					t.Fatalf("demand %d: inconsistent suggestions", demand)
				}
				for index, expected := range testCase.suggestions {
					suggestion := (*diagnostic.Suggestions)[index]
					if suggestion.Message.Id != expected.messageID || suggestion.Message.Description != expected.description {
						t.Errorf("demand %d: wrong suggestion message", demand)
					}
					output, _, fixed := linter.ApplyRuleFixes(testCase.source, (*diagnostic.Suggestions)[index:index+1])
					if !fixed || output != expected.output {
						t.Errorf("demand %d: unexpected suggestion %q", demand, output)
					}
				}
			}
		})
	}
}

// Additional AST and scope regressions verified against Unicorn v75.0.0.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/test/prefer-optional-catch-binding.js
package prefer_optional_catch_binding_test

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_optional_catch_binding"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"reflect"
	"testing"
)

func TestPreferOptionalCatchBindingExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_optional_catch_binding.PreferOptionalCatchBindingRule, []rule_tester.ValidTestCase{
		{Code: "try {} catch(e) { e = replacement; }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "try {} catch(e) { (() => e)(); }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "try {} catch(e) { var e = 1; }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "try {} catch(e) { typeof e; }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "try {} catch({message: value}) { consume(value); }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "try {} catch([value = sideEffect()]) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "try {} catch({message: value = sideEffect()}) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "try {} catch({...rest}) { consume(rest); }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "try {} catch(e: any) { console.log(e); }", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "try {} catch(e) { type E = typeof e; }", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "try {} catch(e) { var {e} = value; }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "try {} catch(e) { for (var e of values) {} }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "try {} catch(e) { for (var e in values) {} }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "try {} catch(e) { (() => {e = 1;})(); }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "try {} catch(e) { for (var e in object) {} }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "try {} catch(e) { var {x:e} = source; }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "try {} catch(e) { try {} catch(other) { var e = 1; consume(other); } }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
	}, []rule_tester.InvalidTestCase{
		{Code: "try {} catch(e) { const f = (e) => e; }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"try {} catch{ const f = (e) => e; }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "with-name", Message: "Remove unused catch binding `e`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "try {} catch(e) { { let e; consume(e); } }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"try {} catch{ { let e; consume(e); } }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "with-name", Message: "Remove unused catch binding `e`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "try {} catch([value]) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"try {} catch{}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "without-name", Message: "Remove unused catch binding.", Line: 1, Column: 14, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "try {} catch([,,]) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"try {} catch{}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "without-name", Message: "Remove unused catch binding.", Line: 1, Column: 14, EndLine: 1, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "try {} catch({cause: {message}}) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"try {} catch{}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "without-name", Message: "Remove unused catch binding.", Line: 1, Column: 14, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "try {} catch({ /* retain */ cause: {message}}) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "without-name", Message: "Remove unused catch binding.", Line: 1, Column: 14, EndLine: 1, EndColumn: 46, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "try {} catch(/* before */ e /* after */) /* body */ {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"try {} catch/* before */  /* after *//* body */ {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "with-name", Message: "Remove unused catch binding `e`.", Line: 1, Column: 27, EndLine: 1, EndColumn: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "try {} catch(e) \ufeff {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"try {} catch{}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "with-name", Message: "Remove unused catch binding `e`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "try {} catch(e) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"try {} catch{}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "with-name", Message: "Remove unused catch binding `e`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "try {} catch(e) // retain\n{}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"try {} catch// retain\n{}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "with-name", Message: "Remove unused catch binding `e`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "\"😀\"; try {} catch(e) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"\"😀\"; try {} catch{}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "with-name", Message: "Remove unused catch binding `e`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "try {} catch(e) { try {} catch(e) { consume(e); } }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"try {} catch{ try {} catch(e) { consume(e); } }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "with-name", Message: "Remove unused catch binding `e`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "try {} catch(e: unknown) {}", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"try {} catch{}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "with-name", Message: "Remove unused catch binding `e`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "try {} catch(e) { var e; }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"try {} catch{ var e; }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "with-name", Message: "Remove unused catch binding `e`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "try {} catch(e) { function f(){var e = 1;} }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"try {} catch{ function f(){var e = 1;} }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "with-name", Message: "Remove unused catch binding `e`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "try {} catch(e) { try {} catch(e) {var e = 1;} }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"try {} catch{ try {} catch(e) {var e = 1;} }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "with-name", Message: "Remove unused catch binding `e`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "try {} catch(e) { {let e = 1;} }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"try {} catch{ {let e = 1;} }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "with-name", Message: "Remove unused catch binding `e`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "try {} catch(e) { (() => {var e = 1;})(); }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"try {} catch{ (() => {var e = 1;})(); }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "with-name", Message: "Remove unused catch binding `e`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "try {} catch(e) { function f() { var e = 1; } }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"try {} catch{ function f() { var e = 1; } }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "with-name", Message: "Remove unused catch binding `e`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "try {} catch(e) { (() => { var e = 1; })(); }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"try {} catch{ (() => { var e = 1; })(); }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "with-name", Message: "Remove unused catch binding `e`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "try {} catch(e) { try {} catch(e) { var e = 1; } }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"try {} catch{ try {} catch(e) { var e = 1; } }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "with-name", Message: "Remove unused catch binding `e`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "try {} catch(e) { { let e; e = 1; } }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"try {} catch{ { let e; e = 1; } }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "with-name", Message: "Remove unused catch binding `e`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "try {} catch(e) { class C { f() { var e = 1; } } }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"try {} catch{ class C { f() { var e = 1; } } }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "with-name", Message: "Remove unused catch binding `e`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
	})
}

func TestPreferOptionalCatchBindingArtifactsFollowDemand(t *testing.T) {
	for _, testCase := range []struct{ source, output string }{{source: "try {} catch (_) {}", output: "try {} catch {}"},
		{source: "try {} catch ({/* inner comment */ message}) {}", output: "try {} catch ({/* inner comment */ message}) {}"}} {
		t.Run(testCase.source, func(t *testing.T) {
			program, sourceFile, err := rule_tester.NewProgramHelper(fixtures.GetRootDir()).CreateTestProgram(testCase.source, "edit-demand.js", "tsconfig.json")
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
						return []rule.ConfiguredRule{{Name: prefer_optional_catch_binding.PreferOptionalCatchBindingRule.Name, Severity: rule.SeverityError, Run: func(ctx rule.RuleContext) rule.RuleListeners {
							return prefer_optional_catch_binding.PreferOptionalCatchBindingRule.Run(ctx, nil)
						}}}
					},
					Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { found = append(found, d) }},
				})
				if len(found) != 1 {
					t.Fatalf("demand %d: expected one diagnostic, got %d", demand, len(found))
				}
				diagnostics[index] = found[0]
			}
			all := diagnostics[len(diagnostics)-1]
			for index, demand := range demands {
				diagnostic := diagnostics[index]
				if diagnostic.Range != all.Range || !reflect.DeepEqual(diagnostic.Message, all.Message) {
					t.Errorf("demand %d changed diagnostic identity", demand)
				}
				wantFix := testCase.source != testCase.output && (demand == rule.EditDemandAutofix || demand == rule.EditDemandAll)
				if (diagnostic.FixesPtr != nil) != wantFix {
					t.Errorf("demand %d: unexpected fix artifacts", demand)
				}
				if wantFix && !reflect.DeepEqual(diagnostic.FixesPtr, all.FixesPtr) {
					t.Errorf("demand %d changed fix artifacts", demand)
				}
				if diagnostic.Suggestions != nil {
					t.Errorf("demand %d: autofix-only rule produced suggestions", demand)
				}
			}
			// Applying fixes can sort slices in place; identity checks are finished.
			for index, demand := range demands {
				wantFix := testCase.source != testCase.output && (demand == rule.EditDemandAutofix || demand == rule.EditDemandAll)
				expected := testCase.source
				if wantFix {
					expected = testCase.output
				}
				output, _, fixed := linter.ApplyRuleFixes(testCase.source, diagnostics[index:index+1])
				if output != expected || fixed != wantFix {
					t.Errorf("demand %d: got %q, want %q", demand, output, expected)
				}
			}
		})
	}
}

// Verified against eslint-plugin-unicorn v75.0.0; see LICENSE.
package no_array_reverse_test

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_array_reverse"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"reflect"
	"testing"
)

// Covers wrapper and receiver shapes, option defaults, and UTF-16 locations.
func TestNoArrayReverseExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_array_reverse.NoArrayReverseRule, []rule_tester.ValidTestCase{
		{Code: "result = array[\"reverse\"]()", FileName: "case.js", Options: []any{}},
		{Code: "result = array.reverse?.()", FileName: "case.js", Options: []any{}},
		{Code: "result = array?.reverse?.()", FileName: "case.js", Options: []any{}},
		{Code: "(array.reverse())", FileName: "case.js", Options: []any{}},
		{Code: "result = (array?.reverse)()", FileName: "case.js", Options: []any{}},
		{Code: "result = new Set().reverse()", FileName: "case.js", Options: []any{}},
		{Code: "class C { #reverse() {} f() { return this.#reverse(); } }", FileName: "case.js", Options: []any{}},
		{Code: "array.reverse()", FileName: "case.js", Options: []any{map[string]any{}}},
		{Code: "array.reverse()", FileName: "case.js", Options: []any{map[string]any{"allowExpressionStatement": true}}},
		{Code: "function f(array: Set<number>) { return array.reverse(); }", FileName: "case.ts", Options: []any{}},
		{Code: "result = (array.reverse as Function)()", FileName: "case.ts", Options: []any{}},
	}, []rule_tester.InvalidTestCase{
		{Code: "result = (array.reverse)()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toReversed()` instead of `Array#reverse()`.", Line: 1, Column: 17, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "result = (array.toReversed)()"}}},
		}},
		{Code: "result = ([...array]).reverse()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toReversed()` instead of `Array#reverse()`.", Line: 1, Column: 23, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-spreading-array", Output: "result = (array).toReversed()"}, {MessageId: "suggestion-not-spreading-array", Output: "result = ([...array]).toReversed()"}}},
		}},
		{Code: "result = [...array,].reverse()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toReversed()` instead of `Array#reverse()`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-spreading-array", Output: "result = array.toReversed()"}, {MessageId: "suggestion-not-spreading-array", Output: "result = [...array,].toReversed()"}}},
		}},
		{Code: "result = [...array, other].reverse()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toReversed()` instead of `Array#reverse()`.", Line: 1, Column: 28, EndLine: 1, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "result = [...array, other].toReversed()"}}},
		}},
		{Code: "result = [...(a, b)].reverse()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toReversed()` instead of `Array#reverse()`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-spreading-array", Output: "result = (a, b).toReversed()"}, {MessageId: "suggestion-not-spreading-array", Output: "result = [...(a, b)].toReversed()"}}},
		}},
		{Code: "result = [...(a ? b : c)].reverse()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toReversed()` instead of `Array#reverse()`.", Line: 1, Column: 27, EndLine: 1, EndColumn: 34, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-spreading-array", Output: "result = (a ? b : c).toReversed()"}, {MessageId: "suggestion-not-spreading-array", Output: "result = [...(a ? b : c)].toReversed()"}}},
		}},
		{Code: "result = [...new Set].reverse()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toReversed()` instead of `Array#reverse()`.", Line: 1, Column: 23, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-spreading-array", Output: "result = (new Set).toReversed()"}, {MessageId: "suggestion-not-spreading-array", Output: "result = [...new Set].toReversed()"}}},
		}},
		{Code: "result = [...(value => value)].reverse()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toReversed()` instead of `Array#reverse()`.", Line: 1, Column: 32, EndLine: 1, EndColumn: 39, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-spreading-array", Output: "result = (value => value).toReversed()"}, {MessageId: "suggestion-not-spreading-array", Output: "result = [...(value => value)].toReversed()"}}},
		}},
		{Code: "result = [... /* keep */ (array)].reverse()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toReversed()` instead of `Array#reverse()`.", Line: 1, Column: 35, EndLine: 1, EndColumn: 42, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-spreading-array", Output: "result = (array).toReversed()"}, {MessageId: "suggestion-not-spreading-array", Output: "result = [... /* keep */ (array)].toReversed()"}}},
		}},
		{Code: "result = [/* copy */ ...array].reverse()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toReversed()` instead of `Array#reverse()`.", Line: 1, Column: 32, EndLine: 1, EndColumn: 39, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-spreading-array", Output: "result = array.toReversed()"}, {MessageId: "suggestion-not-spreading-array", Output: "result = [/* copy */ ...array].toReversed()"}}},
		}},
		{Code: "result = /** @type {Array<number>} */ ([...array]).reverse()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toReversed()` instead of `Array#reverse()`.", Line: 1, Column: 52, EndLine: 1, EndColumn: 59, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-spreading-array", Output: "result = /** @type {Array<number>} */ (array).toReversed()"}, {MessageId: "suggestion-not-spreading-array", Output: "result = /** @type {Array<number>} */ ([...array]).toReversed()"}}},
		}},
		{Code: "result = array.\nreverse()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toReversed()` instead of `Array#reverse()`.", Line: 2, Column: 1, EndLine: 2, EndColumn: 8, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "result = array.\ntoReversed()"}}},
		}},
		{Code: "\"😀\"; result = array.reverse()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toReversed()` instead of `Array#reverse()`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "\"😀\"; result = array.toReversed()"}}},
		}},
		{Code: "function f() { return array.reverse(); }", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toReversed()` instead of `Array#reverse()`.", Line: 1, Column: 29, EndLine: 1, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "function f() { return array.toReversed(); }"}}},
		}},
		{Code: "for (array.reverse(); ready; next()) {}", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toReversed()` instead of `Array#reverse()`.", Line: 1, Column: 12, EndLine: 1, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "for (array.toReversed(); ready; next()) {}"}}},
		}},
		{Code: "result = new Uint8Array().reverse()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toReversed()` instead of `Array#reverse()`.", Line: 1, Column: 27, EndLine: 1, EndColumn: 34, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "result = new Uint8Array().toReversed()"}}},
		}},
		{Code: "result = (function() {}).reverse()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toReversed()` instead of `Array#reverse()`.", Line: 1, Column: 26, EndLine: 1, EndColumn: 33, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "result = (function() {}).toReversed()"}}},
		}},
		{Code: "(array.reverse())", FileName: "case.js", Options: []any{map[string]any{"allowExpressionStatement": false}}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toReversed()` instead of `Array#reverse()`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "(array.toReversed())"}}},
		}},
		{Code: "if (ok) array.reverse()", FileName: "case.js", Options: []any{map[string]any{"allowExpressionStatement": false}}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toReversed()` instead of `Array#reverse()`.", Line: 1, Column: 15, EndLine: 1, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "if (ok) array.toReversed()"}}},
		}},
		{Code: "function f(array: number[]) { return array.reverse(); }", FileName: "case.ts", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toReversed()` instead of `Array#reverse()`.", Line: 1, Column: 44, EndLine: 1, EndColumn: 51, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "function f(array: number[]) { return array.toReversed(); }"}}},
		}},
		{Code: "function f(array: Uint8Array) { return array.reverse(); }", FileName: "case.ts", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toReversed()` instead of `Array#reverse()`.", Line: 1, Column: 46, EndLine: 1, EndColumn: 53, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "function f(array: Uint8Array) { return array.toReversed(); }"}}},
		}},
		{Code: "result = (array as number[]).reverse()", FileName: "case.ts", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toReversed()` instead of `Array#reverse()`.", Line: 1, Column: 30, EndLine: 1, EndColumn: 37, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "result = (array as number[]).toReversed()"}}},
		}},
		{Code: "result = ([...array] as number[]).reverse()", FileName: "case.ts", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toReversed()` instead of `Array#reverse()`.", Line: 1, Column: 35, EndLine: 1, EndColumn: 42, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "result = ([...array] as number[]).toReversed()"}}},
		}},
		{Code: "result = array.reverse<number>()", FileName: "case.ts", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toReversed()` instead of `Array#reverse()`.", Line: 1, Column: 16, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "result = array.toReversed<number>()"}}},
		}},
	})
}

func TestNoArrayReverseEditDemand(t *testing.T) {
	for _, testCase := range []struct {
		source, output string
		suggestions    []struct{ messageID, description, output string }
	}{
		{source: "reversed = [...array].reverse()", output: "reversed = [...array].reverse()", suggestions: []struct{ messageID, description, output string }{{"suggestion-spreading-array", "The spreading object is an array.", "reversed = array.toReversed()"}, {"suggestion-not-spreading-array", "The spreading object is NOT an array.", "reversed = [...array].toReversed()"}}},
		{source: "reversed = array.reverse()", output: "reversed = array.reverse()", suggestions: []struct{ messageID, description, output string }{{"suggestion-apply-replacement", "Switch to `.toReversed()`.", "reversed = array.toReversed()"}}},
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
						return []rule.ConfiguredRule{{Name: no_array_reverse.NoArrayReverseRule.Name, Severity: rule.SeverityError,
							Run: func(ctx rule.RuleContext) rule.RuleListeners {
								return no_array_reverse.NoArrayReverseRule.Run(ctx, nil)
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
				if diagnostic.Range != baseline.Range || !reflect.DeepEqual(diagnostic.Message, baseline.Message) {
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

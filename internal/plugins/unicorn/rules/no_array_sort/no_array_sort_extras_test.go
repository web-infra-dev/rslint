// Verified against eslint-plugin-unicorn v75.0.0; see LICENSE.
package no_array_sort_test

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_array_sort"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"reflect"
	"testing"
)

// Covers wrapper and receiver shapes, option defaults, and UTF-16 locations.
func TestNoArraySortExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_array_sort.NoArraySortRule, []rule_tester.ValidTestCase{
		{Code: "result = array[\"sort\"]()", FileName: "case.js", Options: []any{}},
		{Code: "result = array.sort?.()", FileName: "case.js", Options: []any{}},
		{Code: "result = array?.sort?.()", FileName: "case.js", Options: []any{}},
		{Code: "(array.sort())", FileName: "case.js", Options: []any{}},
		{Code: "result = (array?.sort)()", FileName: "case.js", Options: []any{}},
		{Code: "result = new Set().sort()", FileName: "case.js", Options: []any{}},
		{Code: "class C { #sort() {} f() { return this.#sort(); } }", FileName: "case.js", Options: []any{}},
		{Code: "array.sort()", FileName: "case.js", Options: []any{map[string]any{}}},
		{Code: "array.sort()", FileName: "case.js", Options: []any{map[string]any{"allowExpressionStatement": true}}},
		{Code: "function f(array: Set<number>) { return array.sort(); }", FileName: "case.ts", Options: []any{}},
		{Code: "result = (array.sort as Function)()", FileName: "case.ts", Options: []any{}},
		{Code: "result = array.sort((1))", FileName: "case.js", Options: []any{}},
		{Code: "result = array.sort(getComparator())", FileName: "case.js", Options: []any{}},
		{Code: "result = array.sort(undefined)", FileName: "case.js", Options: []any{}},
		{Code: "result = array.sort(null)", FileName: "case.js", Options: []any{}},
		{Code: "result = array.sort(false)", FileName: "case.js", Options: []any{}},
		{Code: "result = array.sort(/x/)", FileName: "case.js", Options: []any{}},
		{Code: "result = array.sort(1n)", FileName: "case.js", Options: []any{}},
		{Code: "result = array.sort(!compare)", FileName: "case.js", Options: []any{}},
		{Code: "result = array.sort(typeof compare)", FileName: "case.js", Options: []any{}},
		{Code: "result = array.sort(void compare)", FileName: "case.js", Options: []any{}},
		{Code: "result = array.sort(delete obj.key)", FileName: "case.js", Options: []any{}},
		{Code: "result = array.sort(i++)", FileName: "case.js", Options: []any{}},
		{Code: "result = array.sort(class {})", FileName: "case.js", Options: []any{}},
		{Code: "result = array.sort(new Comparator)", FileName: "case.js", Options: []any{}},
		{Code: "result = array.sort(tag`x`)", FileName: "case.js", Options: []any{}},
		{Code: "result = array.sort(this)", FileName: "case.js", Options: []any{}},
		{Code: "result = array.sort(a + b)", FileName: "case.js", Options: []any{}},
		{Code: "result = array.sort(a = b)", FileName: "case.js", Options: []any{}},
		{Code: "async function f() { return array.sort(await compare); }", FileName: "case.js", Options: []any{}},
	}, []rule_tester.InvalidTestCase{
		{Code: "result = (array.sort)()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 17, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "result = (array.toSorted)()"}}},
		}},
		{Code: "result = ([...array]).sort()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 23, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-spreading-array", Output: "result = (array).toSorted()"}, {MessageId: "suggestion-not-spreading-array", Output: "result = ([...array]).toSorted()"}}},
		}},
		{Code: "result = [...array,].sort()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-spreading-array", Output: "result = array.toSorted()"}, {MessageId: "suggestion-not-spreading-array", Output: "result = [...array,].toSorted()"}}},
		}},
		{Code: "result = [...array, other].sort()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 28, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "result = [...array, other].toSorted()"}}},
		}},
		{Code: "result = [...(a, b)].sort()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-spreading-array", Output: "result = (a, b).toSorted()"}, {MessageId: "suggestion-not-spreading-array", Output: "result = [...(a, b)].toSorted()"}}},
		}},
		{Code: "result = [...(a ? b : c)].sort()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 27, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-spreading-array", Output: "result = (a ? b : c).toSorted()"}, {MessageId: "suggestion-not-spreading-array", Output: "result = [...(a ? b : c)].toSorted()"}}},
		}},
		{Code: "result = [...new Set].sort()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 23, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-spreading-array", Output: "result = (new Set).toSorted()"}, {MessageId: "suggestion-not-spreading-array", Output: "result = [...new Set].toSorted()"}}},
		}},
		{Code: "result = [...(value => value)].sort()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 32, EndLine: 1, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-spreading-array", Output: "result = (value => value).toSorted()"}, {MessageId: "suggestion-not-spreading-array", Output: "result = [...(value => value)].toSorted()"}}},
		}},
		{Code: "result = [... /* keep */ (array)].sort()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 35, EndLine: 1, EndColumn: 39, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-spreading-array", Output: "result = (array).toSorted()"}, {MessageId: "suggestion-not-spreading-array", Output: "result = [... /* keep */ (array)].toSorted()"}}},
		}},
		{Code: "result = [/* copy */ ...array].sort()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 32, EndLine: 1, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-spreading-array", Output: "result = array.toSorted()"}, {MessageId: "suggestion-not-spreading-array", Output: "result = [/* copy */ ...array].toSorted()"}}},
		}},
		{Code: "result = /** @type {Array<number>} */ ([...array]).sort()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 52, EndLine: 1, EndColumn: 56, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-spreading-array", Output: "result = /** @type {Array<number>} */ (array).toSorted()"}, {MessageId: "suggestion-not-spreading-array", Output: "result = /** @type {Array<number>} */ ([...array]).toSorted()"}}},
		}},
		{Code: "result = array.\nsort()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 2, Column: 1, EndLine: 2, EndColumn: 5, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "result = array.\ntoSorted()"}}},
		}},
		{Code: "\"😀\"; result = array.sort()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "\"😀\"; result = array.toSorted()"}}},
		}},
		{Code: "function f() { return array.sort(); }", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 29, EndLine: 1, EndColumn: 33, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "function f() { return array.toSorted(); }"}}},
		}},
		{Code: "for (array.sort(); ready; next()) {}", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 12, EndLine: 1, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "for (array.toSorted(); ready; next()) {}"}}},
		}},
		{Code: "result = new Uint8Array().sort()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 27, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "result = new Uint8Array().toSorted()"}}},
		}},
		{Code: "result = (function() {}).sort()", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 26, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "result = (function() {}).toSorted()"}}},
		}},
		{Code: "(array.sort())", FileName: "case.js", Options: []any{map[string]any{"allowExpressionStatement": false}}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "(array.toSorted())"}}},
		}},
		{Code: "if (ok) array.sort()", FileName: "case.js", Options: []any{map[string]any{"allowExpressionStatement": false}}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 15, EndLine: 1, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "if (ok) array.toSorted()"}}},
		}},
		{Code: "function f(array: number[]) { return array.sort(); }", FileName: "case.ts", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 44, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "function f(array: number[]) { return array.toSorted(); }"}}},
		}},
		{Code: "function f(array: Uint8Array) { return array.sort(); }", FileName: "case.ts", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 46, EndLine: 1, EndColumn: 50, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "function f(array: Uint8Array) { return array.toSorted(); }"}}},
		}},
		{Code: "result = (array as number[]).sort()", FileName: "case.ts", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 30, EndLine: 1, EndColumn: 34, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "result = (array as number[]).toSorted()"}}},
		}},
		{Code: "result = ([...array] as number[]).sort()", FileName: "case.ts", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 35, EndLine: 1, EndColumn: 39, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "result = ([...array] as number[]).toSorted()"}}},
		}},
		{Code: "result = array.sort<number>()", FileName: "case.ts", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 16, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "result = array.toSorted<number>()"}}},
		}},
		{Code: "result = array.sort((compare))", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 16, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "result = array.toSorted((compare))"}}},
		}},
		{Code: "result = array.sort(() => 0)", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 16, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "result = array.toSorted(() => 0)"}}},
		}},
		{Code: "result = array.sort(function(a, b) { return a - b; })", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 16, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "result = array.toSorted(function(a, b) { return a - b; })"}}},
		}},
		{Code: "result = array.sort(compare.bind(null))", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 16, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "result = array.toSorted(compare.bind(null))"}}},
		}},
		{Code: "result = array.sort(compare?.bind(null))", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 16, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "result = array.toSorted(compare?.bind(null))"}}},
		}},
		{Code: "result = array.sort(compare.bind?.(null))", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 16, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "result = array.toSorted(compare.bind?.(null))"}}},
		}},
		{Code: "result = array.sort(a && b)", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 16, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "result = array.toSorted(a && b)"}}},
		}},
		{Code: "result = array.sort(a || b)", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 16, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "result = array.toSorted(a || b)"}}},
		}},
		{Code: "result = array.sort(a ?? b)", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 16, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "result = array.toSorted(a ?? b)"}}},
		}},
		{Code: "result = array.sort((a, b))", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 16, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "result = array.toSorted((a, b))"}}},
		}},
		{Code: "result = array.sort(a ? b : c)", FileName: "case.js", Options: []any{}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 16, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "result = array.toSorted(a ? b : c)"}}},
		}},
	})
}

func TestNoArraySortEditDemand(t *testing.T) {
	for _, testCase := range []struct {
		source, output string
		suggestions    []struct{ messageID, description, output string }
	}{
		{source: "sorted = [...array].sort()", output: "sorted = [...array].sort()", suggestions: []struct{ messageID, description, output string }{{"suggestion-spreading-array", "The spreading object is an array.", "sorted = array.toSorted()"}, {"suggestion-not-spreading-array", "The spreading object is NOT an array.", "sorted = [...array].toSorted()"}}},
		{source: "sorted = array.sort()", output: "sorted = array.sort()", suggestions: []struct{ messageID, description, output string }{{"suggestion-apply-replacement", "Switch to `.toSorted()`.", "sorted = array.toSorted()"}}},
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
						return []rule.ConfiguredRule{{Name: no_array_sort.NoArraySortRule.Name, Severity: rule.SeverityError,
							Run: func(ctx rule.RuleContext) rule.RuleListeners { return no_array_sort.NoArraySortRule.Run(ctx, nil) },
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

// Dynamic imports are ESTree ImportExpressions, not ordinary callback calls.
func TestNoArraySortImportComparator(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_array_sort.NoArraySortRule, nil, []rule_tester.InvalidTestCase{
		{Code: "const sorted = array.sort(import(\"x\"));", FileName: "probe.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "const sorted = array.toSorted(import(\"x\"));"}}},
		}},
		{Code: "const sorted = array.sort(import(\"x\"));", FileName: "probe.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error", Message: "Use `Array#toSorted()` instead of `Array#sort()`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion-apply-replacement", Output: "const sorted = array.toSorted(import(\"x\"));"}}},
		}},
	})
}

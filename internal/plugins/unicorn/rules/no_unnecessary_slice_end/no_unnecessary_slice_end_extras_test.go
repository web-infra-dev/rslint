// Additional AST and scope regressions verified against Unicorn v75.0.0.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/test/no-unnecessary-slice-end.js
package no_unnecessary_slice_end_test

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_unnecessary_slice_end"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"reflect"
	"testing"
)

func TestNoUnnecessarySliceEndExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_unnecessary_slice_end.NoUnnecessarySliceEndRule, []rule_tester.ValidTestCase{
		{Code: "const x = (a?.slice)(1, a.length);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const x = a?.slice?.(1, a.length);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const x = a[\"slice\"](1, a.length);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const x = a.slice(1, a[\"length\"]);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const x = getArray().slice(1, getArray().length);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const x = a.slice(...start, Infinity);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const x = a.slice(1, ...end);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const x = a.slice(1, +Infinity);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const x = a.slice(1, Number[\"POSITIVE_INFINITY\"]);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "const x = a.slice(1, Number?.POSITIVE_INFINITY);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "class C { #slice(start, end) {} f() { this.#slice(1, Infinity); } }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "function f(Infinity) { return a.slice(1, Infinity); }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "function f(Number) { return a.slice(1, Number.POSITIVE_INFINITY); }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "a.slice(1, Infinity);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly", "Infinity": "off"}},
		{Code: "function f(a: Set<number>) { return a.slice(1, Infinity); }", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
	}, []rule_tester.InvalidTestCase{
		{Code: "const x = a.slice(1, (a.length));", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const x = a.slice(1);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-slice-end", Message: "Passing `a.length` as the `end` argument is unnecessary.", Line: 1, Column: 23, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const x = a.slice((1 /* keep */), /* remove */ ((a.length)), /* keep */);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const x = a.slice((1 /* keep */), /* keep */);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-slice-end", Message: "Passing `a.length` as the `end` argument is unnecessary.", Line: 1, Column: 50, EndLine: 1, EndColumn: 58, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const x = a.slice(1 /* keep */, a.length /* keep */);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const x = a.slice(1 /* keep */ /* keep */);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-slice-end", Message: "Passing `a.length` as the `end` argument is unnecessary.", Line: 1, Column: 33, EndLine: 1, EndColumn: 41, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const x = (a.slice)(1, a.length);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const x = (a.slice)(1);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-slice-end", Message: "Passing `a.length` as the `end` argument is unnecessary.", Line: 1, Column: 24, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const x = a.slice(1, a?.length);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const x = a.slice(1);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-slice-end", Message: "Passing `a?.length` as the `end` argument is unnecessary.", Line: 1, Column: 22, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const x = obj.a.slice(1, obj.a.length);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const x = obj.a.slice(1);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-slice-end", Message: "Passing `….length` as the `end` argument is unnecessary.", Line: 1, Column: 26, EndLine: 1, EndColumn: 38, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const x = obj[\"a\"].slice(1, obj.a.length);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const x = obj[\"a\"].slice(1);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-slice-end", Message: "Passing `….length` as the `end` argument is unnecessary.", Line: 1, Column: 29, EndLine: 1, EndColumn: 41, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "\"😀\"; a.slice(1, Infinity);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"\"😀\"; a.slice(1);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-slice-end", Message: "Passing `Infinity` as the `end` argument is unnecessary.", Line: 1, Column: 18, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const x = /** @type {Array<number>} */ (a).slice(1, a.length);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"const x = /** @type {Array<number>} */ (a).slice(1);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-slice-end", Message: "Passing `a.length` as the `end` argument is unnecessary.", Line: 1, Column: 53, EndLine: 1, EndColumn: 61, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "function f(a: readonly number[]) { return a.slice(1, a.length); }", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"function f(a: readonly number[]) { return a.slice(1); }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-slice-end", Message: "Passing `a.length` as the `end` argument is unnecessary.", Line: 1, Column: 54, EndLine: 1, EndColumn: 62, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "function f(a: string) { return a.slice(1, a.length); }", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"function f(a: string) { return a.slice(1); }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-slice-end", Message: "Passing `a.length` as the `end` argument is unnecessary.", Line: 1, Column: 43, EndLine: 1, EndColumn: 51, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(a satisfies number[]).slice(1, (a satisfies number[]).length);", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"(a satisfies number[]).slice(1);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-slice-end", Message: "Passing `….length` as the `end` argument is unnecessary.", Line: 1, Column: 33, EndLine: 1, EndColumn: 62, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "a!.slice(1, a!.length);", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"a!.slice(1);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unnecessary-slice-end", Message: "Passing `….length` as the `end` argument is unnecessary.", Line: 1, Column: 13, EndLine: 1, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
	})
}

func TestNoUnnecessarySliceEndArtifactsFollowDemand(t *testing.T) {
	for _, testCase := range []struct{ source, output string }{{source: "foo.slice(1, foo.length)", output: "foo.slice(1)"},
		{source: "// ❌\nconst foo = string.slice(1, Number.POSITIVE_INFINITY);\n\n", output: "// ❌\nconst foo = string.slice(1);\n\n"}} {
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
						return []rule.ConfiguredRule{{Name: no_unnecessary_slice_end.NoUnnecessarySliceEndRule.Name, Severity: rule.SeverityError, Run: func(ctx rule.RuleContext) rule.RuleListeners {
							return no_unnecessary_slice_end.NoUnnecessarySliceEndRule.Run(ctx, nil)
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

func TestNoUnnecessarySliceEndReviewRegressions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_unnecessary_slice_end.NoUnnecessarySliceEndRule, []rule_tester.ValidTestCase{
		{
			Code:     "class Box { length = 2; slice(start: number, end = 99) { return end; } } const value = new Box(); value.slice(1, value.length);",
			FileName: "review.ts",
		},
		{
			Code:     "Number = {POSITIVE_INFINITY: 2}; [0, 1, 2].slice(1, Number.POSITIVE_INFINITY);",
			FileName: "review.js",
		},
		{
			Code:     "let i = 0; const first = [0,1,2], second = [0]; const obj = { get a() { return i++ ? second : first; } }; obj.a.slice(1, obj.a.length);",
			FileName: "review.js",
		},
		{
			Code:     "let a = [0,1,2], b = [0]; a.slice((a = b, 1), a.length);",
			FileName: "review.js",
		},
		{
			Code:     "array.slice(getStart(), array.length);",
			FileName: "review.js",
		},
		{
			Code:     "array.slice(start + getOffset(), array.length);",
			FileName: "review.js",
		},
		{
			Code:     "array.slice(options.start, array.length);",
			FileName: "review.js",
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code:     "foo[-0].slice(1, foo[0].length)",
			FileName: "review.js",
			Output:   []string{"foo[-0].slice(1)"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "no-unnecessary-slice-end",
			}},
		},
		{
			Code:     `foo["a" + "b"].slice(1, foo.ab.length)`,
			FileName: "review.js",
			Output:   []string{`foo["a" + "b"].slice(1)`},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "no-unnecessary-slice-end",
			}},
		},
		{
			Code:     "array.slice(start + 1, array.length);",
			FileName: "review.js",
			Output:   []string{"array.slice(start + 1);"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "no-unnecessary-slice-end",
			}},
		},
		{
			Code:     "array.slice(index ?? 0, array.length);",
			FileName: "review.js",
			Output:   []string{"array.slice(index ?? 0);"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "no-unnecessary-slice-end",
			}},
		},
		{
			Code:     "array.slice(flag ? 1 : 2, array.length);",
			FileName: "review.js",
			Output:   []string{"array.slice(flag ? 1 : 2);"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "no-unnecessary-slice-end",
			}},
		},
	})
}

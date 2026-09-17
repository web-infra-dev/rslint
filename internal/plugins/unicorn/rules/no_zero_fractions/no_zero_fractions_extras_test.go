// Verified against eslint-plugin-unicorn v75.0.0; see LICENSE.
package no_zero_fractions_test

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_zero_fractions"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"reflect"
	"testing"
)

// Covers keyword spacing, ASI-sensitive member receivers, numeric spellings,
// UTF-16 locations, and TypeScript literal types.
func TestNoZeroFractionsExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_zero_fractions.NoZeroFractionsRule, []rule_tester.ValidTestCase{
		{Code: "const n = 0xFF;", FileName: "case.js", Options: []any{}},
		{Code: "const n = 1_000n;", FileName: "case.js", Options: []any{}},
		{Code: "const n = .5;", FileName: "case.js", Options: []any{}},
		{Code: "const n = 1.0_1;", FileName: "case.js", Options: []any{}},
	}, []rule_tester.InvalidTestCase{
		{Code: "const n = .50;", FileName: "case.js", Options: []any{}, Output: []string{"const n = .5;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 13, EndLine: 1, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const n = .000_0;", FileName: "case.js", Options: []any{}, Output: []string{"const n = 0;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 12, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const n = 0.000;", FileName: "case.js", Options: []any{}, Output: []string{"const n = 0;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 12, EndLine: 1, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const n = 1.020_000e+003;", FileName: "case.js", Options: []any{}, Output: []string{"const n = 1.02e+003;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 15, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const n = 1.010;", FileName: "case.js", Options: []any{}, Output: []string{"const n = 1.01;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 15, EndLine: 1, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const café = 1.00;", FileName: "case.js", Options: []any{}, Output: []string{"const café = 1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 15, EndLine: 1, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "\"😀\"; 1.00;", FileName: "case.js", Options: []any{}, Output: []string{"\"😀\"; 1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 8, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const n =\n  1.00;", FileName: "case.js", Options: []any{}, Output: []string{"const n =\n  1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 2, Column: 4, EndLine: 2, EndColumn: 7, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const obj = {1.0: true};", FileName: "case.js", Options: []any{}, Output: []string{"const obj = {1: true};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 15, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const obj = {[1.0]: true};", FileName: "case.js", Options: []any{}, Output: []string{"const obj = {[1]: true};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 16, EndLine: 1, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const value = `${1.0}`;", FileName: "case.js", Options: []any{}, Output: []string{"const value = `${1}`;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 19, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "function* f() { yield.0; }", FileName: "case.js", Options: []any{}, Output: []string{"function* f() { yield 0; }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 23, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "async function f() { return await.0; }", FileName: "case.js", Options: []any{}, Output: []string{"async function f() { return await 0; }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 35, EndLine: 1, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo()\n1.00[key]", FileName: "case.js", Options: []any{}, Output: []string{"foo()\n;(1)[key]"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 2, Column: 2, EndLine: 2, EndColumn: 5, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo()\n1.00?.[key]", FileName: "case.js", Options: []any{}, Output: []string{"foo()\n;(1)?.[key]"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 2, Column: 2, EndLine: 2, EndColumn: 5, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo()\n1.00.toString().trim()", FileName: "case.js", Options: []any{}, Output: []string{"foo()\n;(1).toString().trim()"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 2, Column: 2, EndLine: 2, EndColumn: 5, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "value = {}\n1.00.toString()", FileName: "case.js", Options: []any{}, Output: []string{"value = {}\n;(1).toString()"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 2, Column: 2, EndLine: 2, EndColumn: 5, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "function f() {}\n1.00.toString()", FileName: "case.js", Options: []any{}, Output: []string{"function f() {}\n(1).toString()"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 2, Column: 2, EndLine: 2, EndColumn: 5, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "if (ready) 1.00.toString()", FileName: "case.js", Options: []any{}, Output: []string{"if (ready) (1).toString()"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 13, EndLine: 1, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "while (ready) 1.00.toString()", FileName: "case.js", Options: []any{}, Output: []string{"while (ready) (1).toString()"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 16, EndLine: 1, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "for (;;) 1.00.toString()", FileName: "case.js", Options: []any{}, Output: []string{"for (;;) (1).toString()"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 11, EndLine: 1, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(1.00).toString()", FileName: "case.js", Options: []any{}, Output: []string{"(1).toString()"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 3, EndLine: 1, EndColumn: 6, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "((1.00)).toString()", FileName: "case.js", Options: []any{}, Output: []string{"((1)).toString()"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 4, EndLine: 1, EndColumn: 7, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(/** @type {number} */ (1.00)).toString()", FileName: "case.js", Options: []any{}, Output: []string{"(/** @type {number} */ (1)).toString()"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 26, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "type Value = 1.00;", FileName: "case.ts", Options: []any{}, Output: []string{"type Value = 1;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 15, EndLine: 1, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const n = 1.00 as const;", FileName: "case.ts", Options: []any{}, Output: []string{"const n = 1 as const;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 12, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const n = 1.00 satisfies number;", FileName: "case.ts", Options: []any{}, Output: []string{"const n = 1 satisfies number;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 1, Column: 12, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
	})
}

func TestNoZeroFractionsEditDemand(t *testing.T) {
	for _, testCase := range []struct {
		source, output string
	}{
		{source: "const foo = 1.0", output: "const foo = 1"},
		{source: "function foo(){return.0}", output: "function foo(){return 0}"},
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
						return []rule.ConfiguredRule{{Name: no_zero_fractions.NoZeroFractionsRule.Name, Severity: rule.SeverityError,
							Run: func(ctx rule.RuleContext) rule.RuleListeners {
								return no_zero_fractions.NoZeroFractionsRule.Run(ctx, nil)
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
				if diagnostic.Suggestions != nil {
					t.Errorf("demand %d produced suggestions for an autofix-only rule", demand)
				}
			}

			// ApplyRuleFixes sorts fix slices in place. Finish every artifact
			// comparison first so map iteration cannot change the baseline.
			for demand, diagnostic := range diagnostics {
				wantFix := demand == rule.EditDemandAutofix || demand == rule.EditDemandAll
				expectedOutput := testCase.source
				if wantFix {
					expectedOutput = testCase.output
				}
				output, _, fixed := linter.ApplyRuleFixes(testCase.source, []rule.RuleDiagnostic{diagnostic})
				if output != expectedOutput || fixed != wantFix {
					t.Errorf("demand %d: unexpected autofix %q", demand, output)
				}
			}
		})
	}
}

// A newly parenthesized receiver must not call the preceding function or class.
func TestNoZeroFractionsStatementBoundaries(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_zero_fractions.NoZeroFractionsRule, nil, []rule_tester.InvalidTestCase{
		{Code: "const foo = function() {}\n1.0.toString()", FileName: "probe.js", Output: []string{"const foo = function() {}\n;(1).toString()"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 2, Column: 2, EndLine: 2, EndColumn: 4, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = () => {}\n1.0[0]", FileName: "probe.js", Output: []string{"const foo = () => {}\n;(1)[0]"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 2, Column: 2, EndLine: 2, EndColumn: 4, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "const foo = class {}\n1.0?.[0]", FileName: "probe.js", Output: []string{"const foo = class {}\n;(1)?.[0]"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 2, Column: 2, EndLine: 2, EndColumn: 4, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo = function() {}\n.0.toString()", FileName: "probe.js", Output: []string{"foo = function() {}\n;(0).toString()"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 2, Column: 2, EndLine: 2, EndColumn: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo = () => {}\n1.0.toString()", FileName: "probe.js", Output: []string{"foo = () => {}\n;(1).toString()"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 2, Column: 2, EndLine: 2, EndColumn: 4, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo = class {}\n1.0[0]", FileName: "probe.js", Output: []string{"foo = class {}\n;(1)[0]"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 2, Column: 2, EndLine: 2, EndColumn: 4, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "class Foo {}\n1.0.toString()", FileName: "probe.js", Output: []string{"class Foo {}\n(1).toString()"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 2, Column: 2, EndLine: 2, EndColumn: 4, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "foo = {bar(){}}\n1.0.toString()", FileName: "probe.js", Output: []string{"foo = {bar(){}}\n;(1).toString()"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "zero-fraction", Message: "Don't use a zero fraction in the number.", Line: 2, Column: 2, EndLine: 2, EndColumn: 4, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
	})
}

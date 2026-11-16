package require_array_sort_compare_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/require_array_sort_compare"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestRequireArraySortCompareExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t,
		&require_array_sort_compare.RequireArraySortCompareRule,
		[]rule_tester.ValidTestCase{
			valid(`new Uint8Array().toSorted()`, "file.js"),
			valid(`array["toSorted"]()`, "file.js"),
		},
		[]rule_tester.InvalidTestCase{
			invalid(`((array)).sort()`, "sort", "file.js", true),
			invalid(`((array)).toSorted()`, "toSorted", "file.js", true),
			// Comments outside the replaced argument range are preserved.
			commentCase(`array.sort(/* before */ undefined /* after */)`, "sort", "file.js", 1, 7, `array.sort(/* before */ COMPARATOR /* after */)`),
			commentCase("array.sort(\n\t// keep\n)", "sort", "file.js", 1, 7, "array.sort(\n\t// keep\nCOMPARATOR)"),
			// The JSDoc cast comment sits outside the parenthesized argument.
			commentCase(`array.sort(/** @type {any} */ (undefined))`, "sort", "file.js", 1, 7, `array.sort(/** @type {any} */ COMPARATOR)`),
			commentCase(`array.sort((undefined /* keep */))`, "sort", "file.js", 1, 7, ""),
			{
				// Upstream replaceArgument replaces the parenthesized argument range.
				Code:     `array.sort(((undefined)))`,
				FileName: "file.js",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Message:   message,
					Line:      1, Column: 7, EndLine: 1, EndColumn: 11,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: numericSuggestionID, Output: "array.sort(" + numericCompare + ")"},
						{MessageId: stringSuggestionID, Output: "array.sort(" + stringCompare + ")"},
					},
				}},
			},
		},
	)
}

// TestRequireArraySortCompareEditDemand checks that the diagnostic is identical
// under every edit demand and that the comparator suggestions appear only when
// suggestions are requested. The rule offers no autofix.
func TestRequireArraySortCompareEditDemand(t *testing.T) {
	for _, test := range []struct {
		code    string
		outputs []string // applied suggestion outputs; nil when suggestions are withheld
	}{
		{"array.sort();", []string{"array.sort(" + numericCompare + ");", "array.sort(" + stringCompare + ");"}},
		{"array.sort((undefined /* keep */));", nil},
	} {
		t.Run(test.code, func(t *testing.T) {
			helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
			program, source, err := helper.CreateTestProgram(test.code, "edit-demand.ts", "tsconfig.json")
			if err != nil {
				t.Fatal(err)
			}
			diagnostics := make(map[rule.EditDemand]rule.RuleDiagnostic, 4)
			for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
				var got []rule.RuleDiagnostic
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program: lintprogram.NewFromCompiler(program), File: source.FileName(),
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{
							Name: require_array_sort_compare.RequireArraySortCompareRule.Name, Severity: rule.SeverityError,
							Run: func(ctx rule.RuleContext) rule.RuleListeners {
								return require_array_sort_compare.RequireArraySortCompareRule.Run(ctx, nil)
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

			baseline, all := diagnostics[rule.EditDemandNone], diagnostics[rule.EditDemandAll]
			for demand, diagnostic := range diagnostics {
				if diagnostic.Range != baseline.Range || !reflect.DeepEqual(diagnostic.Message, baseline.Message) || diagnostic.Severity != baseline.Severity {
					t.Errorf("demand %d changed diagnostic identity", demand)
				}
				if diagnostic.FixesPtr != nil && len(*diagnostic.FixesPtr) != 0 {
					t.Errorf("demand %d produced an unexpected autofix", demand)
				}
				wantSuggestions := test.outputs != nil && (demand == rule.EditDemandSuggestion || demand == rule.EditDemandAll)
				if !wantSuggestions {
					if diagnostic.Suggestions != nil && len(*diagnostic.Suggestions) != 0 {
						t.Errorf("demand %d produced unexpected suggestions", demand)
					}
					continue
				}
				if diagnostic.Suggestions == nil || len(*diagnostic.Suggestions) != len(test.outputs) || !reflect.DeepEqual(diagnostic.Suggestions, all.Suggestions) {
					t.Fatalf("demand %d: inconsistent suggestions", demand)
				}
				for i, suggestion := range *diagnostic.Suggestions {
					output, _, fixed := linter.ApplyRuleFixes(test.code, []rule.RuleSuggestion{suggestion})
					if !fixed || output != test.outputs[i] {
						t.Errorf("demand %d: unexpected suggestion %d output %q", demand, i, output)
					}
				}
			}
		})
	}
}

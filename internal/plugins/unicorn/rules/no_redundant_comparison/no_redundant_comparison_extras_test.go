package no_redundant_comparison_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_redundant_comparison"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoRedundantComparisonTemplateBoundsProgramModes(t *testing.T) {
	code := "function check(a, b) {\n" +
		"a === b && a > Math[`PI`] && b > 2;\n" +
		"a === b && a > `abc`.length && b > 2;\n" +
		"const text = `abc`; a === b && a > text.length && b > 2;\n}"
	program, source, err := rule_tester.NewProgramHelper(fixtures.GetRootDir()).CreateTestProgram(code, "template-bounds.ts", "tsconfig.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []struct {
		name    string
		checker bool
	}{{"source-only", false}, {"project", true}} {
		t.Run(mode.name, func(t *testing.T) {
			sourceProgram := lintprogram.NewFromCompiler(program)
			if !mode.checker {
				sourceProgram, err = lintprogram.NewFromBoundSources(program, []*ast.SourceFile{source})
				if err != nil {
					t.Fatal(err)
				}
			}
			var diagnostics []rule.RuleDiagnostic
			var hasChecker bool
			linter.LintSingleFile(linter.LintSingleFileOptions{
				Program: sourceProgram, File: source.FileName(),
				GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
					return []rule.ConfiguredRule{{Name: no_redundant_comparison.NoRedundantComparisonRule.Name, Severity: rule.SeverityError,
						Run: func(ctx rule.RuleContext) rule.RuleListeners {
							hasChecker = ctx.TypeChecker != nil
							return no_redundant_comparison.NoRedundantComparisonRule.Run(ctx, nil)
						},
					}}
				},
				Consumer: rule.DiagnosticConsumer{Demand: rule.EditDemandAll, Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }},
			})
			if hasChecker != mode.checker || len(diagnostics) != 3 {
				t.Fatalf("checker = %t, diagnostics = %d; want %t and 3", hasChecker, len(diagnostics), mode.checker)
			}
			cursor := 0
			for _, diagnostic := range diagnostics {
				start := cursor + strings.Index(code[cursor:], "&& b > 2") + len("&& ")
				if diagnostic.Range.Pos() != start || diagnostic.Range.End() != start+len("b > 2") ||
					diagnostic.Message.Id != "no-redundant-comparison" ||
					diagnostic.Message.Description != "This comparison is redundant because it is already implied by the other conditions." {
					t.Fatalf("unexpected diagnostic: %+v", diagnostic)
				}
				if diagnostic.FixesPtr != nil && len(*diagnostic.FixesPtr) != 0 || diagnostic.Suggestions != nil {
					t.Fatal("unproven ordering comparisons must not offer edits")
				}
				cursor = start + len("b > 2")
			}
		})
	}
}

func TestNoRedundantComparisonEditsFollowDemand(t *testing.T) {
	for _, test := range []struct {
		code, output    string
		fix, suggestion bool
	}{
		{"function check(a, b) { return a === b && a === 1 && b === 1; }", "function check(a, b) { return a === b && a === 1; }", true, false},
		{"const a = 1, b = 1; a === b && a > 0 && b > 0", "const a = 1, b = 1; a === b && a > 0", false, true},
		{"a.x === b.x && a.x === 1 && b.x === 1", "", false, false},
		{"a === b && a === 1 && /* keep */ b === 1", "", false, false},
		{"function check(a, b, label) { return a === b && a === 1 && label && b === 1; }", "", false, false},
		{"function check(a, b, guard) { return a === b && a === 1 && guard.ready && b === 1; }", "", false, false},
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
						return []rule.ConfiguredRule{{Name: no_redundant_comparison.NoRedundantComparisonRule.Name, Severity: rule.SeverityError,
							Run: func(ctx rule.RuleContext) rule.RuleListeners {
								return no_redundant_comparison.NoRedundantComparisonRule.Run(ctx, nil)
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
				wantFix := test.fix && (demand == rule.EditDemandAutofix || demand == rule.EditDemandAll)
				if wantFix {
					if diagnostic.FixesPtr == nil || len(*diagnostic.FixesPtr) != 1 || !reflect.DeepEqual(diagnostic.FixesPtr, all.FixesPtr) {
						t.Fatalf("demand %d: inconsistent autofixes", demand)
					}
					output, _, fixed := linter.ApplyRuleFixes(test.code, []rule.RuleDiagnostic{diagnostic})
					if !fixed || output != test.output {
						t.Errorf("demand %d: unexpected autofix %q", demand, output)
					}
				} else if diagnostic.FixesPtr != nil && len(*diagnostic.FixesPtr) != 0 {
					t.Errorf("demand %d produced an unexpected autofix", demand)
				}
				wantSuggestion := test.suggestion && (demand == rule.EditDemandSuggestion || demand == rule.EditDemandAll)
				if !wantSuggestion {
					if diagnostic.Suggestions != nil {
						t.Errorf("demand %d produced unexpected suggestions", demand)
					}
					continue
				}
				if diagnostic.Suggestions == nil || len(*diagnostic.Suggestions) != 1 || !reflect.DeepEqual(diagnostic.Suggestions, all.Suggestions) {
					t.Fatalf("demand %d: inconsistent suggestions", demand)
				}
				suggestion := (*diagnostic.Suggestions)[0]
				if suggestion.Message.Id != "no-redundant-comparison/suggestion" || suggestion.Message.Description != "Remove this redundant comparison." {
					t.Errorf("demand %d: wrong suggestion message", demand)
				}
				output, _, fixed := linter.ApplyRuleFixes(test.code, []rule.RuleSuggestion{suggestion})
				if !fixed || output != test.output {
					t.Errorf("demand %d: unexpected suggestion %q", demand, output)
				}
			}
		})
	}
}

// Additional diagnostics mirror Unicorn; edits obey the stricter safety contract.
func TestNoRedundantComparisonExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// Numeric bounds, static values and inclusive endpoints
		{Code: "a > 0 && b <= 0 && a === b", FileName: "case.js"},
		{Code: "a < \"2\" && b < \"3\" && a === b", FileName: "case.js"},
		{Code: "a > true && b > false && a === b", FileName: "case.js"},
		{Code: "a > 2n && b > 1n && a === b", FileName: "case.js"},
		{Code: "a > (0 / 0) && b > 1 && a === b", FileName: "case.js"},
		{Code: "a > 1 && b > (0 / 0) && a === b", FileName: "case.js"},
		{Code: "let limit = 5; limit = unknown; a > limit && b > 3 && a === b", FileName: "case.js"},
		{Code: "a === 1 && b !== 1 && a === b", FileName: "case.js"},
		{Code: "a === 1 && a === 2 && b !== 2 && a !== b", FileName: "case.js"},
		// Reference classes and AST boundaries
		{Code: "a === b && a > b && b > a", FileName: "case.js"},
		{Code: "a === b && c === d && a === 1 && d === 1", FileName: "case.js"},
		{Code: "a === b && a !== 1 && b !== 2", FileName: "case.js"},
		{Code: "a[key] === b[key] && a[key] === 1 && b[other] === 1", FileName: "case.js"},
		{Code: "a === b && a === \"\\uD800\" && b === \"�\"", FileName: "case.js"},
		{Code: "class C { #value; f(b) { return this.#value === b && this.#value === 1 && this[\"#value\"] === 1; } }", FileName: "case.js"},
		{Code: "a === b && a?.value === 1 && b.value === 1", FileName: "case.js"},
		{Code: "a[key?.name] === b[key?.name] && a[key?.name] === 1 && b[key?.name] === 1", FileName: "case.js"},
		{Code: "a === b && a[fn?.()] === 1 && b[fn?.()] === 1", FileName: "case.js"},
		{Code: "a === b && ((a === 1 && b === 1) as boolean)", FileName: "case.ts"},
		// Side effects, comments and removal spans
		{Code: "a === b && a === 1 && String.raw`value` && b === 1", FileName: "case.js"},
		{Code: "a === b && a === 1 && (() => tag`value`) && b === 1", FileName: "case.js"},
		{Code: "async function f() { return a === b && a === 1 && await ready && b === 1; }", FileName: "case.js"},
		{Code: "function* f() { return a === b && a === 1 && (yield ready) && b === 1; }", FileName: "case.js"},
		{Code: "a === b && a === 1 && (a += 1) && b === 1", FileName: "case.js"},
	}
	invalid := []rule_tester.InvalidTestCase{
		// Numeric bounds, static values and inclusive endpoints
		{
			Code: "a >= 3 && b >= 3 && a === b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 11, EndLine: 1, EndColumn: 17,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a >= 3 && b > 3 && a === b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a <= 3 && b <= 3 && a === b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 11, EndLine: 1, EndColumn: 17,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a < 3 && b <= 3 && a === b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 10, EndLine: 1, EndColumn: 16,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a >= 0 && b > -0 && a === b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a < -1 && b < -2 && a === b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a < -2 && b < -1 && a === b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 11, EndLine: 1, EndColumn: 17,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a > Infinity && b > 1 && a === b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 17, EndLine: 1, EndColumn: 22,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "const limit = 2 + 3; a > limit && b > 3 && a === b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 35, EndLine: 1, EndColumn: 40,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "let limit = 5; a > limit && b > 3 && a === b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 29, EndLine: 1, EndColumn: 34,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a >= 1 && 0 <= b && a === b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 11, EndLine: 1, EndColumn: 17,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a === true && b === true && a === b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 15, EndLine: 1, EndColumn: 25,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a === null && b !== null && a !== b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 15, EndLine: 1, EndColumn: 25,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a === 1n && b === 1n && a === b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 13, EndLine: 1, EndColumn: 21,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a === /x/g && b === /x/g && a === b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 15, EndLine: 1, EndColumn: 25,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a === `value` && b === \"value\" && a === b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 18, EndLine: 1, EndColumn: 31,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		// Reference classes and AST boundaries
		{
			Code: "a === b && c === d && c === a && a === 1 && d === 1", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 45, EndLine: 1, EndColumn: 52,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "b !== a && a === 1 && b !== 1", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 23, EndLine: 1, EndColumn: 30,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a === b && a === 1 && b === 1 && b === 1", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 23, EndLine: 1, EndColumn: 30,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "(a === b && (a === 1 && b === 1)) || ready", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 25, EndLine: 1, EndColumn: 32,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a === b && a === 1 && b === 1; c === d && c === 2 && d === 2;", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 23, EndLine: 1, EndColumn: 30,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 54, EndLine: 1, EndColumn: 61,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "this === other && this === 1 && other === 1", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 33, EndLine: 1, EndColumn: 44,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "class C extends B { f(b) { return super.x === b && super.x === 1 && b === 1; } }", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 69, EndLine: 1, EndColumn: 76,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "class C { #x; #y; f() { return this.#x === this.#y && this.#x === 1 && this.#y === 1; } }", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 72, EndLine: 1, EndColumn: 85,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "left[\"a\" + \"b\"] === right.ab && left.ab !== \"x\" && right[\"ab\"] !== `x`", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 52, EndLine: 1, EndColumn: 71,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "(<number>a) === b && a === 1 && b! === 1", FileName: "case.ts",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 33, EndLine: 1, EndColumn: 41,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "fn<number> === other && fn === 1 && other === 1", FileName: "case.ts",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 37, EndLine: 1, EndColumn: 48,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a === b && a === 1 && ((b === 1) as boolean)", FileName: "case.ts",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 24, EndLine: 1, EndColumn: 44,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a === b && a === 1 && ((b === 1) satisfies boolean)", FileName: "case.ts",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 24, EndLine: 1, EndColumn: 51,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a === b && a === 1 && (b === 1)!", FileName: "case.ts",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 23, EndLine: 1, EndColumn: 33,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "(a as Box)[\"v\" as string] === b.v && a.v === 1 && b.v === 1", FileName: "case.ts",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 51, EndLine: 1, EndColumn: 60,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "const view = <div>{a === b && a === 1 && b === 1}</div>;", FileName: "case.tsx",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 42, EndLine: 1, EndColumn: 49,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "\"😀\"; 变量 === 值 && 变量 === 1 && 值 === 1;", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 31, EndLine: 1, EndColumn: 38,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "if (\r\n a === b &&\r\n a === 1 &&\r\n b === 1\r\n) {}", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 4, Column: 2, EndLine: 4, EndColumn: 9,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		// Side effects, comments and removal spans
		{
			Code: "a === b && a === \"/* text */\" && b === \"/* text */\"", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 34, EndLine: 1, EndColumn: 52,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "/* keep */ b !== 1 && (a === 1) && a !== b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 12, EndLine: 1, EndColumn: 19,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "(b !== 1 /* keep */) && (a === 1) && a !== b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 2, EndLine: 1, EndColumn: 9,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "b !== 1 && /* keep */ (a === 1) && a !== b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 1, EndLine: 1, EndColumn: 8,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a === b && a === 1 && (/* keep */ b === 1)", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 35, EndLine: 1, EndColumn: 42,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a === b && a === 1 && b === 1 // keep\n", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 23, EndLine: 1, EndColumn: 30,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "(a === b && a === 1) /* keep */ && b === 1", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 36, EndLine: 1, EndColumn: 43,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a === b && a === 1 && (b === 1 /* keep */)", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 24, EndLine: 1, EndColumn: 31,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "foo()\nb !== 1 && ((a === 1) && a !== b)", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 2, Column: 1, EndLine: 2, EndColumn: 8,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "if (ready) b !== 1 && ((a === 1) && a !== b)", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 12, EndLine: 1, EndColumn: 19,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "tag`value` && a === b && a === 1 && b === 1 && (a = 2)", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 37, EndLine: 1, EndColumn: 44,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
	}
	for i := range valid {
		valid[i].LanguageOptions = rule.LanguageOptions{SourceType: "module"}
	}
	for i := range invalid {
		invalid[i].LanguageOptions = rule.LanguageOptions{SourceType: "module"}
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_redundant_comparison.NoRedundantComparisonRule, valid, invalid)
}

func TestNoRedundantComparisonRuntimeBoundaries(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// JSDoc casts and evaluated expression boundaries
		{Code: "a === b && a === 1 && class { [mutate()]() {} } && b === 1", FileName: "case.js"},
		{Code: "a === b && a === 1 && class { static { mutate(); } } && b === 1", FileName: "case.js"},
		{Code: "a in object && b in object && a === b", FileName: "case.js"},
		{Code: "a instanceof Type && b instanceof Type && a === b", FileName: "case.js"},
		{Code: "a === b && a > lower && b > upper && a > other", FileName: "case.js"},
		{Code: "foo()\n/** @type {boolean} */ (b !== 1) && (a === 1) && a !== b", FileName: "case.js"},
		{Code: "foo()\n/** @satisfies {boolean} */ (b !== 1) && (a === 1) && a !== b", FileName: "case.js"},
	}
	invalid := []rule_tester.InvalidTestCase{
		// JSDoc casts and evaluated expression boundaries
		{
			Code: "/** @type {boolean} */ (a === b && a === 1) && b === 1", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 48, EndLine: 1, EndColumn: 55,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a === b && a === 1 && /** @type {boolean} */ (b === 1)", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 47, EndLine: 1, EndColumn: 54,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "/** @type {boolean} */ (b !== 1) && a === 1 && a !== b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 25, EndLine: 1, EndColumn: 32,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a === b && /** @type {boolean} */ (a === 1 && b === 1)", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 47, EndLine: 1, EndColumn: 54,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "/** @satisfies {boolean} */ ((a === b && a === 1)) && b === 1", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 55, EndLine: 1, EndColumn: 62,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a === b && /** @satisfies {boolean} */ (a === 1 && b === 1)", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 52, EndLine: 1, EndColumn: 59,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a > /** @type {number} */ (1) && b > 2 && a === b", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 1, EndLine: 1, EndColumn: 30,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a === b && a === 1 && (() => mutate()) && b === 1", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 43, EndLine: 1, EndColumn: 50,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a === b && a === 1 && class { method() { mutate(); } } && b === 1", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 59, EndLine: 1, EndColumn: 66,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a === b && a === 1 && ({ get value() { mutate(); } }) && b === 1", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 58, EndLine: 1, EndColumn: 65,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a === b && a === 1 && (b === 1)<boolean>", FileName: "case.ts",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 23, EndLine: 1, EndColumn: 41,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a === b && a === 1 && ((b === 1)! as boolean)<boolean>", FileName: "case.ts",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 23, EndLine: 1, EndColumn: 55,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a === b && a > unknown && a > 1 && b > 2", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 27, EndLine: 1, EndColumn: 32,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a === b && a > 10 && b > unknown && a > 0", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 37, EndLine: 1, EndColumn: 42,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
	}
	for i := range valid {
		valid[i].LanguageOptions = rule.LanguageOptions{SourceType: "module"}
	}
	for i := range invalid {
		invalid[i].LanguageOptions = rule.LanguageOptions{SourceType: "module"}
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_redundant_comparison.NoRedundantComparisonRule, valid, invalid)
}

func TestNoRedundantComparisonStaticTemplateBounds(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// PR review: static template string bounds
		{Code: "const Math = { PI: 4 }; a === b && a > Math[`PI`] && b > 2", FileName: "case.js"},
		{Code: "let text = `abc`; text = `x`; a === b && a > text.length && b > 2", FileName: "case.js"},
		{Code: "const object = {text: `abc`}; a === b && a > object.text.length && b > 2", FileName: "case.js"},
		{Code: "a === b && a > Math?.[`PI`] && b > 2", FileName: "case.js"},
	}
	invalid := []rule_tester.InvalidTestCase{
		// PR review: static template string bounds
		{
			Code: "a === b && a > Math[`PI`] && b > 2", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 30, EndLine: 1, EndColumn: 35,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a === b && a > `abc`.length && b > 2", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 32, EndLine: 1, EndColumn: 37,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "const text = `abc`; a === b && a > text.length && b > 2", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 51, EndLine: 1, EndColumn: 56,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a === b && a > `😀a`.length && b > 2", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 32, EndLine: 1, EndColumn: 37,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "const text = `abc`; a === b && a > text[`length`] && b > 2", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 54, EndLine: 1, EndColumn: 59,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a === b && a > Number[`MAX_SAFE_INTEGER`] && b > 2", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 46, EndLine: 1, EndColumn: 51,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
		{
			Code: "a === b && a > Math[(`PI`)] && b > 2", FileName: "case.js",
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-redundant-comparison", Message: "This comparison is redundant because it is already implied by the other conditions.", Line: 1, Column: 32, EndLine: 1, EndColumn: 37,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
				},
			},
		},
	}
	for i := range valid {
		valid[i].LanguageOptions = rule.LanguageOptions{SourceType: "module"}
	}
	for i := range invalid {
		invalid[i].LanguageOptions = rule.LanguageOptions{SourceType: "module"}
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_redundant_comparison.NoRedundantComparisonRule, valid, invalid)
}

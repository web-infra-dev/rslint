package prefer_flat_math_min_max_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_flat_math_min_max"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferFlatMathMinMaxExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_flat_math_min_max.PreferFlatMathMinMaxRule,
		[]rule_tester.ValidTestCase{
			{Code: "Math.max(Math.min(a, b), Math.max?.(c, d));"},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "Math.max((Math.max(a, b)), Math.min(c, d));", Output: []string{"Math.max(a, b, Math.min(c, d));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-flat-math-min-max", Message: "Prefer a flat `Math.max()` call instead of nested calls.", Line: 1, Column: 1, EndLine: 1, EndColumn: 43}}},
		},
	)
}

func TestPreferFlatMathMinMaxEditDemand(t *testing.T) {
	const source = "Math.max(Math.max(a, b), c);"
	const fixedSource = "Math.max(a, b, c);"
	program, sourceFile, err := rule_tester.NewProgramHelper(fixtures.GetRootDir()).CreateTestProgram(source, "edit-demand.js", "tsconfig.json")
	if err != nil {
		t.Fatal(err)
	}

	diagnostics := make(map[rule.EditDemand]rule.RuleDiagnostic, 4)
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
		var got []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: lintprogram.NewFromCompiler(program),
			File:    sourceFile.FileName(),
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{
					Name:     prefer_flat_math_min_max.PreferFlatMathMinMaxRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return prefer_flat_math_min_max.PreferFlatMathMinMaxRule.Run(ctx, nil)
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
		wantFix := demand == rule.EditDemandAutofix || demand == rule.EditDemandAll
		if (diagnostic.FixesPtr != nil) != wantFix {
			t.Errorf("demand %d: unexpected autofix artifacts", demand)
		}
		if wantFix && !reflect.DeepEqual(diagnostic.FixesPtr, all.FixesPtr) {
			t.Errorf("demand %d changed autofix artifacts", demand)
		}
		if diagnostic.Suggestions != nil {
			t.Errorf("demand %d produced suggestions", demand)
		}
		output, _, fixed := linter.ApplyRuleFixes(source, []rule.RuleDiagnostic{diagnostic})
		expected := source
		if wantFix {
			expected = fixedSource
		}
		if fixed != wantFix || output != expected {
			t.Errorf("demand %d: unexpected autofix %q", demand, output)
		}
	}
}

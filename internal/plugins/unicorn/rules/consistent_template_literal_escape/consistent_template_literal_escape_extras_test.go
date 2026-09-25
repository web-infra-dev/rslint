package consistent_template_literal_escape_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/consistent_template_literal_escape"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestConsistentTemplateLiteralEscapeExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&consistent_template_literal_escape.ConsistentTemplateLiteralEscapeRule,
		[]rule_tester.ValidTestCase{
			{Code: "const tagged = tag`$\\{value}`;"},
			{Code: "const tagged = tag`${expr}$\\{value}`;"},
			{Code: "const tagged = tag`${a}$\\{x}${b}tail`;"},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "const value = `prefix $\\{x} suffix`;", Output: []string{"const value = `prefix \\${x} suffix`;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistent-template-literal-escape", Message: "Use `\\${` instead of `$\\{` to escape in template literals.", Line: 1, Column: 15, EndLine: 1, EndColumn: 36}}},
			{Code: "const value = `${a}$\\{x}${b}`;", Output: []string{"const value = `${a}\\${x}${b}`;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistent-template-literal-escape", Message: "Use `\\${` instead of `$\\{` to escape in template literals.", Line: 1, Column: 19, EndLine: 1, EndColumn: 27}}},
		},
	)
}

func TestConsistentTemplateLiteralEscapeEditDemand(t *testing.T) {
	const source = "const value = `$\\{x}`;"
	const fixedSource = "const value = `\\${x}`;"
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
					Name:     consistent_template_literal_escape.ConsistentTemplateLiteralEscapeRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return consistent_template_literal_escape.ConsistentTemplateLiteralEscapeRule.Run(ctx, nil)
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

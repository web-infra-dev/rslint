package prefer_array_last_methods_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/lintprogram"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_array_last_methods"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferArrayLastMethodsExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_array_last_methods.PreferArrayLastMethodsRule,
		[]rule_tester.ValidTestCase{
			{Code: "array.reverse?.().find(logic);"},
			{Code: "array.toReversed?.().reduce(logic);"},
			{Code: "array.reverse().find;"},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "array.reverse().find();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-array-last-methods", Message: "Prefer `Array#findLast()` over `Array#reverse().find()`.", Line: 1, Column: 17, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replace", Output: "array.findLast();"}}}}},
			{Code: "array.toReversed().indexOf(value, fromIndex);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-array-last-methods", Message: "Prefer `Array#lastIndexOf()` over `Array#toReversed().indexOf()`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replace", Output: "array.lastIndexOf(value, fromIndex);"}}}}},
		},
	)
}

func TestPreferArrayLastMethodsEditDemand(t *testing.T) {
	const source = "array.reverse().find(logic);"
	const suggestedSource = "array.findLast(logic);"
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
					Name:     prefer_array_last_methods.PreferArrayLastMethodsRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return prefer_array_last_methods.PreferArrayLastMethodsRule.Run(ctx, nil)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { got = append(got, d) }},
		})
		if len(got) != 1 {
			t.Fatalf("demand %d: got %d diagnostics", demand, len(got))
		}
		diagnostics[demand] = got[0]
		output, _, fixed := linter.ApplyRuleFixes(source, got)
		if fixed || output != source {
			t.Fatalf("demand %d: unexpected autofix %q", demand, output)
		}
	}

	baseline := diagnostics[rule.EditDemandNone]
	for demand, diagnostic := range diagnostics {
		if diagnostic.Range != baseline.Range || !reflect.DeepEqual(diagnostic.Message, baseline.Message) {
			t.Errorf("demand %d changed diagnostic identity", demand)
		}
		if diagnostic.FixesPtr != nil {
			t.Errorf("demand %d produced an autofix", demand)
		}
	}

	if diagnostics[rule.EditDemandNone].Suggestions != nil || diagnostics[rule.EditDemandAutofix].Suggestions != nil {
		t.Fatal("suggestions produced without demand")
	}
	all := diagnostics[rule.EditDemandAll].Suggestions
	if all == nil || len(*all) != 1 || !reflect.DeepEqual(all, diagnostics[rule.EditDemandSuggestion].Suggestions) {
		t.Fatal("inconsistent suggestion artifacts")
	}
	if (*all)[0].Message.Id != "replace" {
		t.Fatalf("wrong suggestion message id %q", (*all)[0].Message.Id)
	}
	output, _, fixed := linter.ApplyRuleFixes(source, *all)
	if !fixed || output != suggestedSource {
		t.Fatalf("unexpected suggestion output %q", output)
	}
}

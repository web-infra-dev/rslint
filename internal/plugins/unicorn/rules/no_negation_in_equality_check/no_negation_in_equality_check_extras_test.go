// Additional AST, scope, and edit-demand coverage checked against Unicorn v75.0.0.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/test/no-negation-in-equality-check.js
package no_negation_in_equality_check_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_negation_in_equality_check"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoNegationInEqualityCheckExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_negation_in_equality_check.NoNegationInEqualityCheckRule, []rule_tester.ValidTestCase{
		{Code: "!(!foo) === bar", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "!((!foo)) === bar", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "foo == !bar", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "!foo < bar", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "!foo + bar", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "!foo instanceof bar", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "(!foo as boolean) === bar", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
	}, []rule_tester.InvalidTestCase{
		{Code: "(!foo) === bar", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-negation-in-equality-check/error", Message: "Negated expression is not allowed in equality check.", Line: 1, Column: 2, EndLine: 1, EndColumn: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-negation-in-equality-check/suggestion", Output: "(foo) !== bar"}}},
		}},
		{Code: "!foo === !bar", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-negation-in-equality-check/error", Message: "Negated expression is not allowed in equality check.", Line: 1, Column: 1, EndLine: 1, EndColumn: 2, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-negation-in-equality-check/suggestion", Output: "foo !== !bar"}}},
		}},
		{Code: "\"😀\"; !foo === bar;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-negation-in-equality-check/error", Message: "Negated expression is not allowed in equality check.", Line: 1, Column: 7, EndLine: 1, EndColumn: 8, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-negation-in-equality-check/suggestion", Output: "\"😀\"; foo !== bar;"}}},
		}},
		{Code: "!foo /* lhs */ === /* rhs */ bar", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-negation-in-equality-check/error", Message: "Negated expression is not allowed in equality check.", Line: 1, Column: 1, EndLine: 1, EndColumn: 2, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-negation-in-equality-check/suggestion", Output: "foo /* lhs */ !== /* rhs */ bar"}}},
		}},
		{Code: "foo\n![] === bar && ready", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-negation-in-equality-check/error", Message: "Negated expression is not allowed in equality check.", Line: 2, Column: 1, EndLine: 2, EndColumn: 2, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-negation-in-equality-check/suggestion", Output: "foo\n;[] !== bar && ready"}}},
		}},
		{Code: "foo\n![] === bar ? a : b", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-negation-in-equality-check/error", Message: "Negated expression is not allowed in equality check.", Line: 2, Column: 1, EndLine: 2, EndColumn: 2, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-negation-in-equality-check/suggestion", Output: "foo\n;[] !== bar ? a : b"}}},
		}},
		{Code: "foo\n!+bar === baz", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-negation-in-equality-check/error", Message: "Negated expression is not allowed in equality check.", Line: 2, Column: 1, EndLine: 2, EndColumn: 2, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-negation-in-equality-check/suggestion", Output: "foo\n;+bar !== baz"}}},
		}},
		{Code: "foo\n!--bar === baz", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-negation-in-equality-check/error", Message: "Negated expression is not allowed in equality check.", Line: 2, Column: 1, EndLine: 2, EndColumn: 2, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-negation-in-equality-check/suggestion", Output: "foo\n;--bar !== baz"}}},
		}},
		{Code: "foo\n!/x/.test(value) === yes", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-negation-in-equality-check/error", Message: "Negated expression is not allowed in equality check.", Line: 2, Column: 1, EndLine: 2, EndColumn: 2, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-negation-in-equality-check/suggestion", Output: "foo\n;/x/.test(value) !== yes"}}},
		}},
		{Code: "foo;\n!(a) === b", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-negation-in-equality-check/error", Message: "Negated expression is not allowed in equality check.", Line: 2, Column: 1, EndLine: 2, EndColumn: 2, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-negation-in-equality-check/suggestion", Output: "foo;\n(a) !== b"}}},
		}},
		{Code: "if (ready) ![a] === b", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-negation-in-equality-check/error", Message: "Negated expression is not allowed in equality check.", Line: 1, Column: 12, EndLine: 1, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-negation-in-equality-check/suggestion", Output: "if (ready) [a] !== b"}}},
		}},
		{Code: "while (ready) ![a] === b", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-negation-in-equality-check/error", Message: "Negated expression is not allowed in equality check.", Line: 1, Column: 15, EndLine: 1, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-negation-in-equality-check/suggestion", Output: "while (ready) [a] !== b"}}},
		}},
		{Code: "const f = () => ![a] === b", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-negation-in-equality-check/error", Message: "Negated expression is not allowed in equality check.", Line: 1, Column: 17, EndLine: 1, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-negation-in-equality-check/suggestion", Output: "const f = () => [a] !== b"}}},
		}},
		{Code: "const x = (!foo === bar)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-negation-in-equality-check/error", Message: "Negated expression is not allowed in equality check.", Line: 1, Column: 12, EndLine: 1, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-negation-in-equality-check/suggestion", Output: "const x = (foo !== bar)"}}},
		}},
		{Code: "function f(){return!/* keep\n */foo === bar;}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-negation-in-equality-check/error", Message: "Negated expression is not allowed in equality check.", Line: 1, Column: 20, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-negation-in-equality-check/suggestion", Output: "function f(){return  (/* keep\n */foo !== bar);}"}}},
		}},
		{Code: "function f(){return !\n foo === bar /* tail */;}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-negation-in-equality-check/error", Message: "Negated expression is not allowed in equality check.", Line: 1, Column: 21, EndLine: 1, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-negation-in-equality-check/suggestion", Output: "function f(){return ( \n foo !== bar /* tail */);}"}}},
		}},
		{Code: "function f(){return (!\n foo === bar);}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-negation-in-equality-check/error", Message: "Negated expression is not allowed in equality check.", Line: 1, Column: 22, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-negation-in-equality-check/suggestion", Output: "function f(){return (\n foo !== bar);}"}}},
		}},
		{Code: "function f(){throw!\r\n foo !== bar;}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-negation-in-equality-check/error", Message: "Negated expression is not allowed in equality check.", Line: 1, Column: 19, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-negation-in-equality-check/suggestion", Output: "function f(){throw  (\r\n foo === bar);}"}}},
		}},
		{Code: "function* f(){yield!foo===bar;}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-negation-in-equality-check/error", Message: "Negated expression is not allowed in equality check.", Line: 1, Column: 20, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-negation-in-equality-check/suggestion", Output: "function* f(){yield foo!==bar;}"}}},
		}},
		{Code: "function f(){return !\nfoo===bar}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-negation-in-equality-check/error", Message: "Negated expression is not allowed in equality check.", Line: 1, Column: 21, EndLine: 1, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-negation-in-equality-check/suggestion", Output: "function f(){return ( \nfoo!==bar)}"}}},
		}},
		{Code: "class C {[!a===b](){}}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-negation-in-equality-check/error", Message: "Negated expression is not allowed in equality check.", Line: 1, Column: 11, EndLine: 1, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-negation-in-equality-check/suggestion", Output: "class C {[a!==b](){}}"}}},
		}},
		{Code: "!/** @type {boolean} */ (foo) === bar", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-negation-in-equality-check/error", Message: "Negated expression is not allowed in equality check.", Line: 1, Column: 1, EndLine: 1, EndColumn: 2, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-negation-in-equality-check/suggestion", Output: "/** @type {boolean} */ (foo) !== bar"}}},
		}},
		{Code: "/** @type {boolean} */ (!foo) === bar", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-negation-in-equality-check/error", Message: "Negated expression is not allowed in equality check.", Line: 1, Column: 25, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-negation-in-equality-check/suggestion", Output: "/** @type {boolean} */ (foo) !== bar"}}},
		}},
		{Code: "!(foo as boolean) === bar", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-negation-in-equality-check/error", Message: "Negated expression is not allowed in equality check.", Line: 1, Column: 1, EndLine: 1, EndColumn: 2, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-negation-in-equality-check/suggestion", Output: "(foo as boolean) !== bar"}}},
		}},
		{Code: "!foo! === bar", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-negation-in-equality-check/error", Message: "Negated expression is not allowed in equality check.", Line: 1, Column: 1, EndLine: 1, EndColumn: 2, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-negation-in-equality-check/suggestion", Output: "foo! !== bar"}}},
		}},
	})
}

func TestNoNegationInEqualityCheckArtifactsFollowDemand(t *testing.T) {
	for _, testCase := range []struct {
		source, output string
		options        []any
		suggestions    []struct{ messageID, description, output string }
	}{
		{source: "!foo === bar", output: "!foo === bar", options: []any{}, suggestions: []struct{ messageID, description, output string }{{"no-negation-in-equality-check/suggestion", "Switch to '!==' check.", "foo !== bar"}}},
		{source: "foo\n!(a) === b", output: "foo\n!(a) === b", options: []any{}, suggestions: []struct{ messageID, description, output string }{{"no-negation-in-equality-check/suggestion", "Switch to '!==' check.", "foo\n;(a) !== b"}}},
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
						return []rule.ConfiguredRule{{Name: no_negation_in_equality_check.NoNegationInEqualityCheckRule.Name, Severity: rule.SeverityError,
							Run: func(ctx rule.RuleContext) rule.RuleListeners {
								return no_negation_in_equality_check.NoNegationInEqualityCheckRule.Run(ctx, testCase.options)
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
				if diagnostic.Range != baseline.Range || !reflect.DeepEqual(diagnostic.Message, baseline.Message) || diagnostic.Severity != baseline.Severity {
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

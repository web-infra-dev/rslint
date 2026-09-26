package no_useless_continue_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_useless_continue"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUselessContinueExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t,
		&no_useless_continue.NoUselessContinueRule,
		[]rule_tester.ValidTestCase{
			{Code: "for (const value of values as number[]) { if (value < 0) { continue; } consume(value); }", FileName: "case.ts"},
			{Code: "for (const value of values) { try { continue; } finally { cleanup(); } }", FileName: "case.js"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: "for (const value of values as number[]) { consume(value); continue; }", FileName: "case.ts",
				Output: []string{"for (const value of values as number[]) { consume(value);  }"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-useless-continue", Message: "Unnecessary `continue` statement.", Line: 1, Column: 59, EndLine: 1, EndColumn: 68}},
			},
			{
				Code: "for (const value of values) {\r\n\tconsume(value);\r\n\tcontinue;\r\n}\r\n", FileName: "case.js",
				Output: []string{"for (const value of values) {\r\n\tconsume(value);\r\n}\r\n"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-useless-continue", Message: "Unnecessary `continue` statement.", Line: 3, Column: 2, EndLine: 3, EndColumn: 11}},
			},
		},
	)
}

func TestNoUselessContinueRemovalBoundaries(t *testing.T) {
	var invalid []rule_tester.InvalidTestCase
	add := func(source, output string, count int) {
		errors := make([]rule_tester.InvalidTestCaseError, count)
		for i := range errors {
			errors[i].MessageId = "no-useless-continue"
		}
		invalid = append(invalid, rule_tester.InvalidTestCase{
			Code: source, FileName: "case.js", Output: []string{output}, Errors: errors,
		})
	}
	for _, newline := range []string{"\n", "\r\n", "\r"} {
		for _, whitespace := range []string{"", " \t\v\f", "\u00a0\u1680\u2000\u200a\u202f\u205f\u3000\ufeff", "\u2028\u2029"} {
			for _, statement := range []string{"continue;", "continue"} {
				add("while (ready) {"+newline+whitespace+statement+whitespace+newline+"}", "while (ready) {"+newline+"}", 1)
			}
		}
	}
	for _, testCase := range []struct{ source, output string }{
		{"while (ready) {\n\t/* before 😀 */ continue; \n}", "while (ready) {\n\t/* before 😀 */  \n}"},
		{"while (ready) {\n\tcontinue; /* after 😀 */\n}", "while (ready) {\n\t /* after 😀 */\n}"},
		{"while (ready) { \u00a0continue;\u00a0 }", "while (ready) { \u00a0\u00a0 }"},
		{"while (ready) {\n\tcontinue /* keep */\n}", "while (ready) {\n\t /* keep */\n}"},
		{"while (ready) {\n\tcontinue\n;\n}", "while (ready) {\n}"},
		{"while (ready) {\r\n\tcontinue\r\n;\r\n}", "while (ready) {\r\n}"},
		{"while (ready) {\n\tcontinue; }", "while (ready) {\n\t }"},
		// Preserve the existing treatment of Unicode separators around a statement.
		{"\ufeff" + "while (ready) {\u2028\tcontinue;\u2029}", "\ufeff" + "while (ready) {\u2028\t\u2029}"},
		{"while (ready) {\n\tcontinue;\u2028// after\n}", "while (ready) {\n\t\u2028// after\n}"},
		{"while (ready) {\n// before\u2028\tcontinue;\n}", "while (ready) {\n// before\u2028\t\n}"},
	} {
		add(testCase.source, testCase.output, 1)
	}
	add(strings.Repeat("while (ready) { continue; }", 64), strings.Repeat("while (ready) {  }", 64), 64)
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_useless_continue.NoUselessContinueRule, nil, invalid)
}

func TestNoUselessContinueEditDemand(t *testing.T) {
	const ruleName = "unicorn/no-useless-continue"
	for _, testCase := range []struct {
		name, source, output string
		count                int
	}{
		{"inline", "while (ready) { continue; }", "while (ready) {  }", 1},
		{"whole CRLF line", "while (ready) {\r\n\tcontinue;\r\n}", "while (ready) {\r\n}", 1},
		{"leading trivia", "while (ready) {\r\n\t/* 😀 */continue;\r\n}", "while (ready) {\r\n\t/* 😀 */\r\n}", 1},
		{"block disabled", "/* eslint-disable unicorn/no-useless-continue */ while (ready) { continue; }", "", 0},
		{"line disabled", "while (ready) {\ncontinue; // eslint-disable-line unicorn/no-useless-continue\n}", "", 0},
		{"next line disabled", "while (ready) {\n// rslint-disable-next-line unicorn/no-useless-continue\ncontinue;\n}", "", 0},
		{"enabled again", "/* eslint-disable unicorn/no-useless-continue */\nwhile (ready) { continue; }\n/* eslint-enable unicorn/no-useless-continue */\nwhile (ready) { continue; }", "/* eslint-disable unicorn/no-useless-continue */\nwhile (ready) { continue; }\n/* eslint-enable unicorn/no-useless-continue */\nwhile (ready) {  }", 1},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
			program, sourceFile, err := helper.CreateTestProgram(testCase.source, "edit-demand.js", "tsconfig.json")
			if err != nil {
				t.Fatal(err)
			}
			var autofix *rule.RuleDiagnostic
			for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
				var diagnostics []rule.RuleDiagnostic
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program: lintprogram.NewFromCompiler(program), File: sourceFile.FileName(),
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{Name: ruleName, Severity: rule.SeverityError,
							Run: func(ctx rule.RuleContext) rule.RuleListeners {
								return no_useless_continue.NoUselessContinueRule.Run(ctx, nil)
							},
						}}
					},
					Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }},
				})
				if len(diagnostics) != testCase.count {
					t.Fatalf("demand %d: got %d diagnostics, want %d", demand, len(diagnostics), testCase.count)
				}
				wantFix := testCase.count > 0 && demand&rule.EditDemandAutofix != 0
				if len(diagnostics) > 0 {
					diagnostic := diagnostics[0]
					start := strings.LastIndex(testCase.source, "continue;")
					if diagnostic.Range.Pos() != start || diagnostic.Range.End() != start+len("continue;") ||
						diagnostic.Message.Id != "no-useless-continue" || diagnostic.Message.Description != "Unnecessary `continue` statement." ||
						diagnostic.RuleName != ruleName || diagnostic.Severity != rule.SeverityError {
						t.Errorf("demand %d: unexpected diagnostic: %+v", demand, diagnostic)
					}
					if (diagnostic.FixesPtr != nil) != wantFix || diagnostic.Suggestions != nil {
						t.Fatalf("demand %d: unexpected edit artifacts", demand)
					}
					if wantFix {
						if autofix != nil && !reflect.DeepEqual(diagnostic.FixesPtr, autofix.FixesPtr) {
							t.Errorf("demand %d changed fix ranges", demand)
						}
						autofix = &diagnostic
						if testCase.name == "whole CRLF line" {
							fixes := *diagnostic.FixesPtr
							if len(fixes) != 1 || fixes[0].Range.Pos() != len("while (ready) {") || fixes[0].Range.End() != start+len("continue;") {
								t.Errorf("demand %d: wrong standalone line removal range: %+v", demand, fixes)
							}
						}
					}
				}
				wantOutput := testCase.source
				if wantFix {
					wantOutput = testCase.output
				}
				output, _, fixed := linter.ApplyRuleFixes(testCase.source, diagnostics)
				if output != wantOutput || fixed != wantFix {
					t.Errorf("demand %d: got fix %q (%t), want %q (%t)", demand, output, fixed, wantOutput, wantFix)
				}
			}
		})
	}
}

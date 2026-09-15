// TestNoEmptyComputedAssignmentKey locks in rule behavior inside an evaluated
// computed property name of a destructuring assignment target. The shared
// traversal contract is covered by internal/linter; this test protects the
// rule's listener-to-diagnostic integration.
package no_empty

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoEmptyComputedAssignmentKey(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoEmptyRule,
		nil,
		[]rule_tester.InvalidTestCase{
			{
				Code: `({ [(() => { if (condition) {} })()]: target } = source);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    29,
						EndLine:   1,
						EndColumn: 31,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
							MessageId: "suggestComment",
							Output:    `({ [(() => { if (condition) { /* empty */ } })()]: target } = source);`,
						}},
					},
				},
			},
		},
	)
}

// ESLint represents class static blocks as StaticBlock rather than
// BlockStatement, so the core no-empty rule must leave them to
// no-empty-static-block.
func TestNoEmptyIgnoresClassStaticBlocks(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoEmptyRule,
		[]rule_tester.ValidTestCase{
			{Code: `class C { static {} }`},
			{Code: `class C { static { } }`},
		},
		nil,
	)
}

func TestNoEmptyIgnoresOnlyCommentsInsideBraces(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoEmptyRule,
		nil,
		[]rule_tester.InvalidTestCase{
			invalidNoEmptyCase(
				`if (foo) /* { // outside */ {}`,
				nil,
				noEmptyErrorSpec{occurrence: 0, statementType: "block", output: `if (foo) /* { // outside */ { /* empty */ }`},
			),
			invalidNoEmptyCase(
				"if (foo) // { // outside\n{}",
				nil,
				noEmptyErrorSpec{occurrence: 0, statementType: "block", output: "if (foo) // { // outside\n{ /* empty */ }"},
			),
			invalidNoEmptyCase(
				`if (foo) {} /* { // outside */`,
				nil,
				noEmptyErrorSpec{occurrence: 0, statementType: "block", output: `if (foo) { /* empty */ } /* { // outside */`},
			),
			invalidNoEmptyCase(
				`switch (foo) /* { // outside */ {}`,
				nil,
				noEmptyErrorSpec{occurrence: 0, statementType: "switch", output: `switch (foo) /* { // outside */ { /* empty */ }`},
			),
		},
	)
}

func TestNoEmptyEditDemand(t *testing.T) {
	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(
		`if (condition) {} switch (value) {}`,
		"edit-demand.ts",
		"tsconfig.json",
	)
	if err != nil {
		t.Fatal(err)
	}
	sourceProgram := lintprogram.NewFromCompiler(program)
	run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
		t.Helper()
		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: sourceProgram,
			File:    sourceFile.FileName(),
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{
					Name:     NoEmptyRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return NoEmptyRule.Run(ctx, nil)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{
				Demand: demand,
				Report: func(diagnostic rule.RuleDiagnostic) {
					diagnostics = append(diagnostics, diagnostic)
				},
			},
		})
		return diagnostics
	}

	for _, demand := range []rule.EditDemand{
		rule.EditDemandNone,
		rule.EditDemandAutofix,
		rule.EditDemandSuggestion,
		rule.EditDemandAll,
	} {
		diagnostics := run(demand)
		if len(diagnostics) != 2 {
			t.Fatalf("demand %d diagnostics = %d, want 2", demand, len(diagnostics))
		}
		for index, diagnostic := range diagnostics {
			statementType := []string{"block", "switch"}[index]
			description := "Empty " + statementType + " statement."
			if diagnostic.Message.Id != "unexpected" || diagnostic.Message.Description != description {
				t.Errorf("demand %d diagnostic %d message = %#v, want unexpected/%q", demand, index, diagnostic.Message, description)
			}
			if diagnostic.Message.Data["type"] != statementType {
				t.Errorf("demand %d diagnostic %d data = %#v, want type %q", demand, index, diagnostic.Message.Data, statementType)
			}
			if got := sourceFile.Text()[diagnostic.Range.Pos():diagnostic.Range.End()]; got != "{}" {
				t.Errorf("demand %d diagnostic %d range text = %q, want %q", demand, index, got, "{}")
			}
			if diagnostic.FixesPtr != nil {
				t.Errorf("demand %d diagnostic %d unexpectedly has autofixes", demand, index)
			}
			wantSuggestions := demand&rule.EditDemandSuggestion != 0
			if (diagnostic.Suggestions != nil) != wantSuggestions {
				t.Errorf(
					"demand %d diagnostic %d suggestion presence = %v, want %v",
					demand,
					index,
					diagnostic.Suggestions != nil,
					wantSuggestions,
				)
			} else if wantSuggestions {
				if len(*diagnostic.Suggestions) != 1 {
					t.Errorf("demand %d diagnostic %d suggestions = %d, want 1", demand, index, len(*diagnostic.Suggestions))
					continue
				}
				suggestion := (*diagnostic.Suggestions)[0]
				suggestionDescription := "Add comment inside empty " + statementType + " statement."
				if suggestion.Message.Id != "suggestComment" || suggestion.Message.Description != suggestionDescription {
					t.Errorf("demand %d diagnostic %d suggestion message = %#v, want suggestComment/%q", demand, index, suggestion.Message, suggestionDescription)
				}
				if suggestion.Message.Data["type"] != statementType {
					t.Errorf("demand %d diagnostic %d suggestion data = %#v, want type %q", demand, index, suggestion.Message.Data, statementType)
				}
				if len(suggestion.FixesArr) != 1 {
					t.Errorf("demand %d diagnostic %d suggestion fixes = %d, want 1", demand, index, len(suggestion.FixesArr))
					continue
				}
				fix := suggestion.FixesArr[0]
				if fix.Text != " /* empty */ " || fix.Range.Pos() != diagnostic.Range.Pos()+1 || fix.Range.End() != diagnostic.Range.End()-1 {
					t.Errorf("demand %d diagnostic %d suggestion fix = %#v, want inner range and comment", demand, index, fix)
				}
			}
		}
	}
}

package name_replacements_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/name_replacements"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNameReplacementsExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&name_replacements.NameReplacementsRule,
		[]rule_tester.ValidTestCase{
			{Code: "let error"},
			{Code: "const i18nData = {}"},
			{Code: "({err: 1})"},
			{Code: "const err = 1", Options: map[string]any{"checkVariables": false}},
			{Code: "const err = 1", Options: map[string]any{"allowList": map[string]any{"err": true}}},
			{Code: "const {err} = value"},
			{Code: "object.err", Options: map[string]any{"checkProperties": true}},
			{Code: "+object.err", Options: map[string]any{"checkProperties": true}},
			{Code: "!object.err", Options: map[string]any{"checkProperties": true}},
			{Code: `object["err"] = value`, Options: map[string]any{"checkProperties": true}},
			{Code: "class Example { #err = 1 }", Options: map[string]any{"checkProperties": true}},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   "const err = 1; use(err)",
				Output: []string{"const error = 1; use(error)"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "replace",
					Message:   "The variable `err` should be named `error`. A more descriptive name will do too.",
					Line:      1, Column: 7, EndLine: 1, EndColumn: 10,
				}},
			},
			{
				Code: "let e = value",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "suggestion",
					Message:   "Please rename the variable `e`. Suggested names are: `error`, `event`. A more descriptive name will do too.",
					Line:      1, Column: 5, EndLine: 1, EndColumn: 6,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "rename", Output: "let error = value"},
						{MessageId: "rename", Output: "let event = value"},
					},
				}},
			},
			{
				Code:    "object.err = 1",
				Options: map[string]any{"checkProperties": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "replace",
					Message:   "The property `err` should be named `error`. A more descriptive name will do too.",
					Line:      1, Column: 8, EndLine: 1, EndColumn: 11,
				}},
			},
			{
				Code:    "++object.err",
				Options: map[string]any{"checkProperties": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "replace", Line: 1, Column: 10, EndLine: 1, EndColumn: 13,
				}},
			},
			{
				Code:     "use(value)",
				FileName: "err.ts",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "replace",
					Message:   "The filename `err.ts` should be named `error.ts`. A more descriptive name will do too.",
					Line:      1, Column: 1, EndLine: 1, EndColumn: 11,
				}},
			},
			{
				Code:   "const err = 1; use((err))",
				Output: []string{"const error = 1; use((error))"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "replace", Line: 1, Column: 7, EndLine: 1, EndColumn: 10,
				}},
			},
			{
				Code: "export const err = 1",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "replace", Line: 1, Column: 14, EndLine: 1, EndColumn: 17,
				}},
			},
			{
				Code: "const err = () => null; const element = <err />",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "replace", Line: 1, Column: 7, EndLine: 1, EndColumn: 10,
				}},
			},
		},
	)
}

func TestNameReplacementsRejectsInvalidIgnorePattern(t *testing.T) {
	options := []any{map[string]any{"ignore": []any{"["}}}
	err := name_replacements.NameReplacementsRule.Schema.Validate(options)
	if err == nil || !strings.Contains(err.Error(), "ignore") {
		t.Fatalf("expected invalid ignore regexp to be rejected, got %v", err)
	}
}

func TestNameReplacementsEditDemand(t *testing.T) {
	for _, testCase := range []struct {
		code        string
		wantFix     bool
		wantSuggest bool
	}{
		{code: "const err = 1", wantFix: true},
		{code: "let e = 1", wantSuggest: true},
	} {
		t.Run(testCase.code, func(t *testing.T) {
			program, sourceFile, err := rule_tester.NewProgramHelper(fixtures.GetRootDir()).CreateTestProgram(testCase.code, "edit-demand.js", "tsconfig.json")
			if err != nil {
				t.Fatal(err)
			}
			diagnostics := map[rule.EditDemand]rule.RuleDiagnostic{}
			for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
				var found []rule.RuleDiagnostic
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program: lintprogram.NewFromCompiler(program), File: sourceFile.FileName(),
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{Name: name_replacements.NameReplacementsRule.Name, Severity: rule.SeverityError, Run: func(ctx rule.RuleContext) rule.RuleListeners {
							return name_replacements.NameReplacementsRule.Run(ctx, nil)
						}}}
					},
					Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(diagnostic rule.RuleDiagnostic) { found = append(found, diagnostic) }},
				})
				if len(found) != 1 {
					t.Fatalf("demand %d: got %d diagnostics", demand, len(found))
				}
				diagnostics[demand] = found[0]
			}
			all := diagnostics[rule.EditDemandAll]
			for demand, diagnostic := range diagnostics {
				if diagnostic.Range != all.Range || !reflect.DeepEqual(diagnostic.Message, all.Message) {
					t.Fatalf("demand %d changed diagnostic identity", demand)
				}
				hasFix := testCase.wantFix && demand&rule.EditDemandAutofix != 0
				hasSuggestion := testCase.wantSuggest && demand&rule.EditDemandSuggestion != 0
				if (diagnostic.FixesPtr != nil) != hasFix || (diagnostic.Suggestions != nil) != hasSuggestion {
					t.Fatalf("demand %d: unexpected edit artifacts", demand)
				}
				if hasFix && !reflect.DeepEqual(diagnostic.FixesPtr, all.FixesPtr) {
					t.Fatalf("demand %d changed fixes", demand)
				}
				if hasSuggestion && !reflect.DeepEqual(diagnostic.Suggestions, all.Suggestions) {
					t.Fatalf("demand %d changed suggestions", demand)
				}
			}
		})
	}
}

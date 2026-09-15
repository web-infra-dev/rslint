package prefer_strict_equal_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/prefer_strict_equal"
)

func sharedRule() rule.Rule {
	return shared.NewRule(shared.Config{
		Name: "test/prefer-strict-equal",
		Prepare: func(rule.RuleContext) shared.Runtime {
			return shared.Runtime{Parse: func(node *ast.Node) *shared.ExpectCall {
				entries := testFramework.GetMemberEntries(node)
				if len(entries) == 0 {
					return nil
				}
				matcher := entries[len(entries)-1]
				return &shared.ExpectCall{Matcher: matcher.Name, MatcherEntry: matcher}
			}}
		},
	})
}

func TestPreferStrictEqualSharedAccessorContract(t *testing.T) {
	r := sharedRule()
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&r,
		[]rule_tester.ValidTestCase{
			{Code: `expect(value).toStrictEqual(other)`},
			{Code: `expect(value)[matcher](other)`},
			{Code: `expect(value)[0](other)`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `expect(something).toEqual(somethingElse);`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "useToStrictEqual", Message: "Use `toStrictEqual()` instead", Line: 1, Column: 19,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestReplaceWithStrictEqual", Output: `expect(something).toStrictEqual(somethingElse);`}},
				}},
			},
			{
				Code: `expect(something)['toEqual'](somethingElse);`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "useToStrictEqual", Line: 1, Column: 19,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestReplaceWithStrictEqual", Output: `expect(something)['toStrictEqual'](somethingElse);`}},
				}},
			},
			{
				Code: "expect(something)[`toEqual`](somethingElse);",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "useToStrictEqual", Line: 1, Column: 19,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestReplaceWithStrictEqual", Output: "expect(something)[`toStrictEqual`](somethingElse);"}},
				}},
			},
			{
				Code:   `const toEqual = 'toEqual'; expect(something)[toEqual](somethingElse);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToStrictEqual", Line: 1, Column: 46}},
			},
		},
	)
}

func TestPreferStrictEqualSharedEditDemand(t *testing.T) {
	r := sharedRule()
	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(
		`expect(something)["toEqual"](somethingElse);`,
		"testdata/edit-demand.ts",
		"tsconfig.json",
	)
	if err != nil {
		t.Fatal(err)
	}

	run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
		t.Helper()
		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program:     lintprogram.NewFromCompiler(program),
			File:        sourceFile.FileName(),
			HasTypeInfo: true,
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{
					Name: r.Name, Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners { return r.Run(ctx, nil) },
				}}
			},
			Consumer: rule.DiagnosticConsumer{
				Demand: demand,
				Report: func(diagnostic rule.RuleDiagnostic) { diagnostics = append(diagnostics, diagnostic) },
			},
		})
		if len(diagnostics) != 1 {
			t.Fatalf("demand %d: diagnostics = %d, want 1", demand, len(diagnostics))
		}
		return diagnostics
	}

	diagnosticsOnly := run(rule.EditDemandNone)
	autofixOnly := run(rule.EditDemandAutofix)
	suggestionOnly := run(rule.EditDemandSuggestion)
	allEdits := run(rule.EditDemandAll)

	withoutEdits := func(diagnostic rule.RuleDiagnostic) rule.RuleDiagnostic {
		diagnostic.FixesPtr = nil
		diagnostic.Suggestions = nil
		return diagnostic
	}
	want := withoutEdits(allEdits[0])
	for demand, diagnostics := range map[rule.EditDemand][]rule.RuleDiagnostic{
		rule.EditDemandNone: diagnosticsOnly, rule.EditDemandAutofix: autofixOnly, rule.EditDemandSuggestion: suggestionOnly,
	} {
		if got := withoutEdits(diagnostics[0]); !reflect.DeepEqual(got, want) {
			t.Errorf("demand %d changed diagnostic:\ngot  %#v\nwant %#v", demand, got, want)
		}
	}
	if diagnosticsOnly[0].Suggestions != nil || autofixOnly[0].Suggestions != nil {
		t.Fatal("suggestions were materialized without suggestion demand")
	}
	if diagnosticsOnly[0].FixesPtr != nil || autofixOnly[0].FixesPtr != nil || suggestionOnly[0].FixesPtr != nil || allEdits[0].FixesPtr != nil {
		t.Fatal("suggestion-only rule unexpectedly materialized autofixes")
	}
	if suggestionOnly[0].Suggestions == nil || !reflect.DeepEqual(suggestionOnly[0].Suggestions, allEdits[0].Suggestions) {
		t.Fatal("suggestions differ between suggestion and all edit demands")
	}
}

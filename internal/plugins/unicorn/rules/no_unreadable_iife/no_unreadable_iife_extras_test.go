// Ported from eslint-plugin-unicorn v74.0.0; see LICENSE.
package no_unreadable_iife_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_unreadable_iife"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnreadableIifeExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_unreadable_iife.NoUnreadableIifeRule, []rule_tester.ValidTestCase{
		{Code: "const f = () => (value);", FileName: "case.js"},
		{Code: "(function () { return value; })();", FileName: "case.js"},
		{Code: "(() => value)();", FileName: "case.js"},
		{Code: "(() => {})();", FileName: "case.js"},
		{Code: "new (() => (value))();", FileName: "case.js"},
		{Code: "(() => (value)).call(null);", FileName: "case.js"},
		{Code: "((() => (value)) as Function)();", FileName: "case.ts"},
		{Code: "(() => { return value; })();", FileName: "suggested.js"},
		{Code: "(() => { return a /* inside */ + b; })();", FileName: "suggested.js"},
		{Code: "(() => { return {/* inside */ value}; })();", FileName: "suggested.js"},
		{Code: "(() => { return \"/* not a comment */\"; })();", FileName: "suggested.js"},
		{Code: "const emoji = \"😀\"; (() => { return \"é\"; })();", FileName: "suggested.js"},
		{Code: "(() => { return a,\r\n  b; })();", FileName: "suggested.js"},
		{Code: "(() => { return value; })?.();", FileName: "suggested.js"},
		{Code: "(/** @type {Function} */ (() => { return value; }))();", FileName: "suggested.js"},
		{Code: "(() => { return left() + right(); })();", FileName: "suggested.js"},
	}, []rule_tester.InvalidTestCase{
		{Code: "(() => (((value))))();", FileName: "case.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unreadable-iife", Message: "IIFE with parenthesized arrow function body is considered unreadable.", Line: 1, Column: 8, EndLine: 1, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "(() => { return value; })();"}}},
		}},
		{Code: "(() => (value /* keep */))();", FileName: "case.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unreadable-iife", Message: "IIFE with parenthesized arrow function body is considered unreadable.", Line: 1, Column: 8, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(() => (/* outer */ (value)))();", FileName: "case.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unreadable-iife", Message: "IIFE with parenthesized arrow function body is considered unreadable.", Line: 1, Column: 8, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(() => ((value) /* outer */))();", FileName: "case.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unreadable-iife", Message: "IIFE with parenthesized arrow function body is considered unreadable.", Line: 1, Column: 8, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(() => (\n// keep\nvalue\n))();", FileName: "case.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unreadable-iife", Message: "IIFE with parenthesized arrow function body is considered unreadable.", Line: 1, Column: 8, EndLine: 4, EndColumn: 2, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(() => (value // keep\n))();", FileName: "case.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unreadable-iife", Message: "IIFE with parenthesized arrow function body is considered unreadable.", Line: 1, Column: 8, EndLine: 2, EndColumn: 2, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(() => (a /* inside */ + b))();", FileName: "case.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unreadable-iife", Message: "IIFE with parenthesized arrow function body is considered unreadable.", Line: 1, Column: 8, EndLine: 1, EndColumn: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "(() => { return a /* inside */ + b; })();"}}},
		}},
		{Code: "(() => ({/* inside */ value}))();", FileName: "case.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unreadable-iife", Message: "IIFE with parenthesized arrow function body is considered unreadable.", Line: 1, Column: 8, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "(() => { return {/* inside */ value}; })();"}}},
		}},
		{Code: "(() => (\"/* not a comment */\"))();", FileName: "case.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unreadable-iife", Message: "IIFE with parenthesized arrow function body is considered unreadable.", Line: 1, Column: 8, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "(() => { return \"/* not a comment */\"; })();"}}},
		}},
		{Code: "const emoji = \"😀\"; (() => (\"é\"))();", FileName: "case.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unreadable-iife", Message: "IIFE with parenthesized arrow function body is considered unreadable.", Line: 1, Column: 28, EndLine: 1, EndColumn: 33, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const emoji = \"😀\"; (() => { return \"é\"; })();"}}},
		}},
		{Code: "(() => (\r\n  a,\r\n  b\r\n))();", FileName: "case.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unreadable-iife", Message: "IIFE with parenthesized arrow function body is considered unreadable.", Line: 1, Column: 8, EndLine: 4, EndColumn: 2, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "(() => { return a,\r\n  b; })();"}}},
		}},
		{Code: "(() => (value))?.();", FileName: "case.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unreadable-iife", Message: "IIFE with parenthesized arrow function body is considered unreadable.", Line: 1, Column: 8, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "(() => { return value; })?.();"}}},
		}},
		{Code: "(/** @type {Function} */ (() => (value)))();", FileName: "case.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unreadable-iife", Message: "IIFE with parenthesized arrow function body is considered unreadable.", Line: 1, Column: 33, EndLine: 1, EndColumn: 40, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "(/** @type {Function} */ (() => { return value; }))();"}}},
		}},
		{Code: "(() => (/** @type {number} */ (value)))();", FileName: "case.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unreadable-iife", Message: "IIFE with parenthesized arrow function body is considered unreadable.", Line: 1, Column: 8, EndLine: 1, EndColumn: 39, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "(() => (left() + right()))();", FileName: "case.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unreadable-iife", Message: "IIFE with parenthesized arrow function body is considered unreadable.", Line: 1, Column: 8, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "(() => { return left() + right(); })();"}}},
		}},
		// The compiler synthesizes assertion nodes for these JavaScript comments.
		{Code: "(() => /** @type {number} */ (value))();", FileName: "case.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unreadable-iife", Message: "IIFE with parenthesized arrow function body is considered unreadable.", Line: 1, Column: 30, EndLine: 1, EndColumn: 37, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "(() => /** @type {number} */ { return value; })();"}}},
		}},
		{Code: "(() => /** @satisfies {number} */ (value))();", FileName: "case.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "no-unreadable-iife", Message: "IIFE with parenthesized arrow function body is considered unreadable.", Line: 1, Column: 35, EndLine: 1, EndColumn: 42, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "(() => /** @satisfies {number} */ { return value; })();"}}},
		}},
	})
}

func TestNoUnreadableIifeEditDemand(t *testing.T) {
	t.Parallel()
	const source = "(() => (value))();"
	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(source, "edit-demand.js", "tsconfig.json")
	if err != nil {
		t.Fatal(err)
	}
	diagnostics := make(map[rule.EditDemand]rule.RuleDiagnostic, 4)
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
		var got []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: lintprogram.NewFromCompiler(program), File: sourceFile.FileName(),
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{Name: no_unreadable_iife.NoUnreadableIifeRule.Name, Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return no_unreadable_iife.NoUnreadableIifeRule.Run(ctx, nil)
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
	base := diagnostics[rule.EditDemandNone]
	for demand, d := range diagnostics {
		if d.Range != base.Range || !reflect.DeepEqual(d.Message, base.Message) {
			t.Errorf("demand %d changed diagnostic identity", demand)
		}
		if d.FixesPtr != nil {
			t.Errorf("demand %d produced an autofix", demand)
		}
	}
	if diagnostics[rule.EditDemandNone].Suggestions != nil || diagnostics[rule.EditDemandAutofix].Suggestions != nil {
		t.Fatal("suggestions produced without demand")
	}
	suggestions := diagnostics[rule.EditDemandAll].Suggestions
	if suggestions == nil || len(*suggestions) != 1 || !reflect.DeepEqual(suggestions, diagnostics[rule.EditDemandSuggestion].Suggestions) {
		t.Fatal("inconsistent suggestion artifacts")
	}
	if (*suggestions)[0].Message.Description != "Use a block statement body." {
		t.Fatal("wrong suggestion message")
	}
	output, _, fixed := linter.ApplyRuleFixes(source, *suggestions)
	if !fixed || output != "(() => { return value; })();" {
		t.Fatalf("unexpected suggestion output %q", output)
	}
}

// TestNoDivRegexExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise. Each case carries an inline comment pointing
// at the specific branch / Dimension 4 row / tsgo AST quirk it covers, so
// future refactors can't silently regress them without breaking a named
// lock-in.
//
// Dimension walk notes for no-div-regex:
//   - Dimension 2 (scoping & nesting): N/A — the rule inspects only a
//     RegularExpressionLiteral's own token text; it performs no scope lookup
//     or ancestor walk, so function/class/block nesting cannot affect
//     detection.
//   - Dimension 4 (access/key forms): N/A — the rule never inspects a
//     property or key; it fires directly on the RegularExpressionLiteral
//     node.
//   - Dimension 4 (declaration/container forms): N/A — the rule does not
//     target function or class declarations.
//   - Dimension 4 (graceful degradation: SpreadAssignment/RestElement,
//     overload signatures/abstract/declare members): N/A — none of these
//     shapes can contain or affect a regular expression literal's own token
//     text.
package no_div_regex

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDivRegexExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoDivRegexRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 1: flags don't affect the leading-character check ----
			{Code: `var re = /foo=bar/gimsuy;`},
			// ---- Dimension 4: receiver wrapper — parenthesized, non-`=` pattern stays valid ----
			{Code: `var re = (/foo/);`},
			// ---- Dimension 1: other literal kinds coexisting with a valid regex must not
			// be mistaken for regex literals or otherwise trigger the rule ----
			{Code: `var s = "=foo"; var n = 0; var b = true; var re = /foo/;`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: graceful degradation — minimal one-character pattern ("=") ----
			{
				Code:   `var re = /=/;`,
				Output: []string{`var re = /[=]/;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 10, EndLine: 1, EndColumn: 13},
				},
			},
			// ---- Dimension 1: flags after an offending pattern don't suppress the fix ----
			{
				Code:   `var re = /=foo/gi;`,
				Output: []string{`var re = /[=]foo/gi;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 10, EndLine: 1, EndColumn: 18},
				},
			},
			// ---- Dimension 4: receiver wrapper — parenthesized regex literal is still detected ----
			{
				Code:   `var re = (/=foo/);`,
				Output: []string{`var re = (/[=]foo/);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 11, EndLine: 1, EndColumn: 17},
				},
			},
			// ---- Dimension 4: TS non-null assertion suffix on the literal itself is still detected ----
			{
				Code:   `var ok = /=foo/!.test("x");`,
				Output: []string{`var ok = /[=]foo/!.test("x");`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 10, EndLine: 1, EndColumn: 16},
				},
			},
			// ---- Dimension 4: TS `as` type-expression wrapper is still detected ----
			{
				Code:   `var ok = (/=foo/ as RegExp).test("x");`,
				Output: []string{`var ok = (/[=]foo/ as RegExp).test("x");`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 11, EndLine: 1, EndColumn: 17},
				},
			},
			// ---- Position assertion: multi-line source, literal on its own line ----
			{
				Code:   "var re =\n  /=foo/;",
				Output: []string{"var re =\n  /[=]foo/;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 2, Column: 3, EndLine: 2, EndColumn: 9},
				},
			},
			// ---- Real-user: default-parameter validator pattern ----
			{
				Code:   `function validate(input, pattern = /=x/) { return pattern.test(input); }`,
				Output: []string{`function validate(input, pattern = /[=]x/) { return pattern.test(input); }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 36, EndLine: 1, EndColumn: 40},
				},
			},
			// ---- Real-user: class static field validator pattern ----
			{
				Code:   `class Validator { static pattern = /=email/; }`,
				Output: []string{`class Validator { static pattern = /[=]email/; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 36, EndLine: 1, EndColumn: 44},
				},
			},
			// ---- Locks in upstream Literal() sole branch: the listener reports
			// unconditionally per-node, so multiple offending literals in the same file
			// are each reported independently (no dedupe / early-exit after the first) ----
			{
				Code:   `var a = /=foo/; var b = /=bar/;`,
				Output: []string{`var a = /[=]foo/; var b = /[=]bar/;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 9, EndLine: 1, EndColumn: 15},
					{MessageId: "unexpected", Line: 1, Column: 25, EndLine: 1, EndColumn: 31},
				},
			},
		},
	)
}

// TestNoDivRegexEditDemand exercises Dimension 3 (autofix boundaries):
// diagnostic count, message, and range must stay identical across every edit
// demand, and the fix must materialize only when autofix is requested.
func TestNoDivRegexEditDemand(t *testing.T) {
	t.Parallel()

	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(
		"var a = /=foo/;\nvar b = /=bar/;",
		"edit-demand.ts",
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
					Name:     NoDivRegexRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return NoDivRegexRule.Run(ctx, nil)
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
		if len(diagnostics) != 2 {
			t.Fatalf("demand %d: diagnostics = %d, want 2", demand, len(diagnostics))
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
	for index := range allEdits {
		for demand, diagnostics := range map[rule.EditDemand][]rule.RuleDiagnostic{
			rule.EditDemandNone:       diagnosticsOnly,
			rule.EditDemandAutofix:    autofixOnly,
			rule.EditDemandSuggestion: suggestionOnly,
		} {
			if got, want := withoutEdits(diagnostics[index]), withoutEdits(allEdits[index]); !reflect.DeepEqual(got, want) {
				t.Errorf("demand %d changed diagnostic %d:\ngot  %#v\nwant %#v", demand, index, got, want)
			}
		}
		if diagnosticsOnly[index].FixesPtr != nil || suggestionOnly[index].FixesPtr != nil {
			t.Fatalf("diagnostic %d: non-autofix demand materialized fixes", index)
		}
		if autofixOnly[index].FixesPtr == nil ||
			!reflect.DeepEqual(autofixOnly[index].FixesPtr, allEdits[index].FixesPtr) {
			t.Fatalf("diagnostic %d: autofix and all-edits demands produced different fixes", index)
		}
		if fixes := *allEdits[index].FixesPtr; len(fixes) == 0 {
			t.Fatalf("diagnostic %d: all-edits demand produced no fixes", index)
		}
	}
	for _, diagnostics := range [][]rule.RuleDiagnostic{
		diagnosticsOnly,
		autofixOnly,
		suggestionOnly,
		allEdits,
	} {
		for index, diagnostic := range diagnostics {
			if diagnostic.Suggestions != nil {
				t.Fatalf("diagnostic %d: autofix-only rule materialized suggestions", index)
			}
		}
	}
}

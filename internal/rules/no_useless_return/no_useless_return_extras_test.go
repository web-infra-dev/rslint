package no_useless_return

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"

	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoUselessReturnExtras locks in the edge shapes the upstream test suite
// doesn't exercise: every code path root tsgo builds, the containers the walk
// has to degrade gracefully on, and code shapes real projects write. Lock-ins
// for the arms of the upstream source live in
// no_useless_return_extras_branches_test.go. Every verdict below was taken from
// ESLint itself.
func TestNoUselessReturnExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUselessReturnRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: nesting / traversal boundaries — a nested code path keeps its own returns ----
			{Code: `function outer() { if (c) { return; } const inner = () => { post(); }; }`},
			// ---- Dimension 4: graceful degradation ----
			{Code: `function f() {}`},
			// ---- Dimension 4: graceful degradation ----
			{Code: `const f = () => 5;`},
			// ---- Dimension 4: graceful degradation ----
			{Code: `function f() { return; ; }`},
			// ---- Dimension 4: graceful degradation ----
			{Code: `function f() { switch (s) {} }`},
			// ---- Dimension 4: graceful degradation ----
			{Code: `function f() { if (c) { return; } switch (s) {} }`},
			// ---- Dimension 4: graceful degradation ----
			{Code: `class K { static {} }`},
			// ---- Real-user: Express-style middleware guard ----
			{Code: `function middleware(req, res, next) { if (!req.user) { res.sendStatus(401); return; } next(); }`},
			// ---- Real-user: guard clause that is the whole body of the branch ----
			{Code: `function render(props) { if (!props.visible) return; draw(props); }`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: declaration / container forms — a code path root of every shape the rule can report in ----
			{
				Code:   `function f() { return; }`,
				Output: []string{`function f() {  }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 16, EndLine: 1, EndColumn: 23},
				},
			},
			// ---- Dimension 4: declaration / container forms — a code path root of every shape the rule can report in ----
			{
				Code:   `const f = function () { return; };`,
				Output: []string{`const f = function () {  };`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 25, EndLine: 1, EndColumn: 32},
				},
			},
			// ---- Dimension 4: declaration / container forms — a code path root of every shape the rule can report in ----
			{
				Code:   `const f = () => { return; };`,
				Output: []string{`const f = () => {  };`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 19, EndLine: 1, EndColumn: 26},
				},
			},
			// ---- Dimension 4: declaration / container forms — a code path root of every shape the rule can report in ----
			{
				Code:   `class K { m() { return; } }`,
				Output: []string{`class K { m() {  } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 17, EndLine: 1, EndColumn: 24},
				},
			},
			// ---- Dimension 4: declaration / container forms — a code path root of every shape the rule can report in ----
			{
				Code:   `class K { constructor() { return; } }`,
				Output: []string{`class K { constructor() {  } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 27, EndLine: 1, EndColumn: 34},
				},
			},
			// ---- Dimension 4: declaration / container forms — a code path root of every shape the rule can report in ----
			{
				Code:   `class K { get p() { return; } }`,
				Output: []string{`class K { get p() {  } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 21, EndLine: 1, EndColumn: 28},
				},
			},
			// ---- Dimension 4: declaration / container forms — a code path root of every shape the rule can report in ----
			{
				Code:   `class K { set p(v) { return; } }`,
				Output: []string{`class K { set p(v) {  } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 22, EndLine: 1, EndColumn: 29},
				},
			},
			// ---- Dimension 4: declaration / container forms — a code path root of every shape the rule can report in ----
			{
				Code:   `class K { static m() { return; } }`,
				Output: []string{`class K { static m() {  } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 24, EndLine: 1, EndColumn: 31},
				},
			},
			// ---- Dimension 4: declaration / container forms — a code path root of every shape the rule can report in ----
			{
				Code:   `class K { static { return; } }`,
				Output: []string{`class K { static {  } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 20, EndLine: 1, EndColumn: 27},
				},
			},
			// ---- Dimension 4: declaration / container forms — a code path root of every shape the rule can report in ----
			{
				Code:   `class K { p = () => { return; }; }`,
				Output: []string{`class K { p = () => {  }; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 23, EndLine: 1, EndColumn: 30},
				},
			},
			// ---- Dimension 4: declaration / container forms — a code path root of every shape the rule can report in ----
			{
				Code:   `const o = { m() { return; } };`,
				Output: []string{`const o = { m() {  } };`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 19, EndLine: 1, EndColumn: 26},
				},
			},
			// ---- Dimension 4: declaration / container forms — a code path root of every shape the rule can report in ----
			{
				Code:   `async function f() { return; }`,
				Output: []string{`async function f() {  }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 22, EndLine: 1, EndColumn: 29},
				},
			},
			// ---- Dimension 4: declaration / container forms — a code path root of every shape the rule can report in ----
			{
				Code:   `function* f() { return; }`,
				Output: []string{`function* f() {  }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 17, EndLine: 1, EndColumn: 24},
				},
			},
			// ---- Dimension 4: declaration / container forms — a code path root of every shape the rule can report in ----
			{
				Code:   `async function* f() { return; }`,
				Output: []string{`async function* f() {  }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 23, EndLine: 1, EndColumn: 30},
				},
			},
			// ---- Dimension 4: declaration / container forms — a code path root of every shape the rule can report in ----
			{
				Code:   `const f = async () => { return; };`,
				Output: []string{`const f = async () => {  };`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 25, EndLine: 1, EndColumn: 32},
				},
			},
			// ---- Dimension 4: nesting / traversal boundaries — a nested code path keeps its own returns ----
			{
				Code:   `function outer() { pre(); function inner() { return; } }`,
				Output: []string{`function outer() { pre(); function inner() {  } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 46, EndLine: 1, EndColumn: 53},
				},
			},
			// ---- Dimension 4: nesting / traversal boundaries — a nested code path keeps its own returns ----
			{
				Code:   `function outer() { return; function inner() { post(); } }`,
				Output: []string{`function outer() {  function inner() { post(); } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 20, EndLine: 1, EndColumn: 27},
				},
			},
			// ---- Dimension 4: graceful degradation ----
			{
				Code:   `function f() { return; {} }`,
				Output: []string{`function f() {  {} }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 16, EndLine: 1, EndColumn: 23},
				},
			},
			// ---- Dimension 4: graceful degradation ----
			{
				Code:   `function f() { return; { { } } }`,
				Output: []string{`function f() {  { { } } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 16, EndLine: 1, EndColumn: 23},
				},
			},
			// ---- Real-user: Express-style middleware guard ----
			{
				Code:   `function middleware(req, res, next) { if (!req.user) { res.sendStatus(401); return; } }`,
				Output: []string{`function middleware(req, res, next) { if (!req.user) { res.sendStatus(401);  } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 77, EndLine: 1, EndColumn: 84},
				},
			},
			// ---- Real-user: event handler that bails out of a switch ----
			{
				Code:   `function onKey(e) { switch (e.key) { case 'Escape': close(); return; case 'Enter': submit(); return; } }`,
				Output: []string{`function onKey(e) { switch (e.key) { case 'Escape': close(); return; case 'Enter': submit();  } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 94, EndLine: 1, EndColumn: 101},
				},
			},
			// ---- Real-user: cleanup wrapper — the finally block still runs, so the return in the try stays reported ----
			{
				Code:   `async function load() { try { setBusy(true); return; } finally { setBusy(false); } }`,
				Output: []string{`async function load() { try { setBusy(true);  } finally { setBusy(false); } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 46, EndLine: 1, EndColumn: 53},
				},
			},
			// ---- Real-user: eslint#8026 — the fix must not clash with a no-else-return fix on the same if ----
			{
				Code:   `function foo() { if (c) { bar(); return; } else { baz(); } }`,
				Output: []string{`function foo() { if (c) { bar();  } else { baz(); } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 34, EndLine: 1, EndColumn: 41},
				},
			},
		},
	)
}

// TestNoUselessReturnEditDemand checks that what the diagnostic says never
// depends on whether its fix was asked for, and that the fix is built only when
// it was.
func TestNoUselessReturnEditDemand(t *testing.T) {
	t.Parallel()

	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(
		`function first() { return; }
function second() { if (c) return; }
function third() { g(); return /* keep */; }`,
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
			Program:      program,
			File:         sourceFile.FileName(),
			ExcludePaths: []string{},
			GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
				return []linter.ConfiguredRule{{
					Name:     NoUselessReturnRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return NoUselessReturnRule.Run(ctx, nil)
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
		if len(diagnostics) != 3 {
			t.Fatalf("demand %d: diagnostics = %d, want 3", demand, len(diagnostics))
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
		want := withoutEdits(allEdits[index])
		for demand, diagnostics := range map[rule.EditDemand][]rule.RuleDiagnostic{
			rule.EditDemandNone:       diagnosticsOnly,
			rule.EditDemandAutofix:    autofixOnly,
			rule.EditDemandSuggestion: suggestionOnly,
		} {
			if got := withoutEdits(diagnostics[index]); !reflect.DeepEqual(got, want) {
				t.Errorf("demand %d diagnostic %d changed:\ngot  %#v\nwant %#v", demand, index, got, want)
			}
		}
	}

	// Only the first return stands in a statement list and carries no comment.
	wantFixes := []bool{true, false, false}
	for index, wantFix := range wantFixes {
		if got := autofixOnly[index].FixesPtr != nil; got != wantFix {
			t.Errorf("autofix diagnostic %d fix presence = %t, want %t", index, got, wantFix)
		}
		if !reflect.DeepEqual(autofixOnly[index].FixesPtr, allEdits[index].FixesPtr) {
			t.Errorf("autofix and all-edits diagnostic %d produced different fixes", index)
		}
	}
	for _, diagnostics := range [][]rule.RuleDiagnostic{diagnosticsOnly, suggestionOnly} {
		for index, diagnostic := range diagnostics {
			if diagnostic.FixesPtr != nil {
				t.Errorf("non-autofix diagnostic %d materialized fixes", index)
			}
		}
	}
	for _, diagnostics := range [][]rule.RuleDiagnostic{diagnosticsOnly, autofixOnly, suggestionOnly, allEdits} {
		for index, diagnostic := range diagnostics {
			if diagnostic.Suggestions != nil {
				t.Errorf("diagnostic %d materialized suggestions", index)
			}
		}
	}
}

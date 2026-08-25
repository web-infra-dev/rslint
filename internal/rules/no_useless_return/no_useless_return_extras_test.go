package no_useless_return

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"

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
			// ---- A `for` whose head has no incrementor leaves the `try` block
			// with no throwable node, so nothing reaches the `catch` clause ----
			{Code: `function f() { try { for (var i = 0;;) { break; } return 1; } catch (e) { return; } }`},
			// ---- The incrementor is laid out where the walk stands, so inside
			// code nothing reaches it forks nothing either ----
			{Code: `function f() { throw e; try { for (var i = 0;; i++) { break; } return 1; } catch (e) { return; } }`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- JSDoc traversal boundary: the annotation is comment metadata,
			// while the authored ReturnStatement is still visited exactly once. ----
			// ESLint 10.9.0 with @typescript-eslint/parser 8.67.0 reports this
			// single runtime return at the same range.
			{
				Code:     `/** @returns {void} */ function f(){ return; } f();`,
				FileName: "jsdoc-return.js",
				TSConfig: "tsconfig.allow-js.json",
				Output:   []string{`/** @returns {void} */ function f(){  } f();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 38, EndLine: 1, EndColumn: 45},
				},
			},
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
			// ---- A `for` incrementor inside a `try` block is laid out before
			// the body, so it is one of the throwable nodes the block forks on
			// even when the body always leaves the loop abruptly ----
			{
				Code:   `function f() { try { for (var i = 0;; i++) { break; } return 1; } catch (e) { return; } }`,
				Output: []string{`function f() { try { for (var i = 0;; i++) { break; } return 1; } catch (e) {  } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 79, EndLine: 1, EndColumn: 86},
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
			// ---- Dimension 4: fix range — a computed key sits outside the method's retained range, so both returns go in one pass ----
			{
				Code:   `class K { [(() => { return; })()]() { return; } }`,
				Output: []string{`class K { [(() => {  })()]() {  } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 21, EndLine: 1, EndColumn: 28},
					{MessageId: "unnecessaryReturn", Line: 1, Column: 39, EndLine: 1, EndColumn: 46},
				},
			},
			// ---- Dimension 4: fix range — a computed key sits outside the method's retained range, so both returns go in one pass ----
			{
				Code:   `const o = { [(() => { return; })()]() { return; } };`,
				Output: []string{`const o = { [(() => {  })()]() {  } };`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 23, EndLine: 1, EndColumn: 30},
					{MessageId: "unnecessaryReturn", Line: 1, Column: 41, EndLine: 1, EndColumn: 48},
				},
			},
			// ---- Dimension 4: fix range — a computed key sits outside the method's retained range, so both returns go in one pass ----
			{
				Code:   `class K { get [(() => { return; })()]() { return; } }`,
				Output: []string{`class K { get [(() => {  })()]() {  } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 25, EndLine: 1, EndColumn: 32},
					{MessageId: "unnecessaryReturn", Line: 1, Column: 43, EndLine: 1, EndColumn: 50},
				},
			},
			// ---- Dimension 4: fix range — a computed key sits outside the method's retained range, so both returns go in one pass ----
			{
				Code:   `class K { set [(() => { return; })()](v) { return; } }`,
				Output: []string{`class K { set [(() => {  })()](v) {  } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 25, EndLine: 1, EndColumn: 32},
					{MessageId: "unnecessaryReturn", Line: 1, Column: 44, EndLine: 1, EndColumn: 51},
				},
			},
			// ---- Dimension 4: fix range — a decorator sits outside the method's retained range, so both returns go in one pass ----
			{
				Code:   `class K { @dec(() => { return; }) m() { return; } }`,
				Output: []string{`class K { @dec(() => {  }) m() {  } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 24, EndLine: 1, EndColumn: 31},
					{MessageId: "unnecessaryReturn", Line: 1, Column: 41, EndLine: 1, EndColumn: 48},
				},
			},
			// ---- Dimension 4: fix range — a parameter decorator is inside the retained range, so its return waits for the next pass, as upstream does ----
			{
				Code: `class K { constructor(@dec(() => { return; }) x) { return; } }`,
				Output: []string{
					`class K { constructor(@dec(() => { return; }) x) {  } }`,
					`class K { constructor(@dec(() => {  }) x) {  } }`,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 36, EndLine: 1, EndColumn: 43},
					{MessageId: "unnecessaryReturn", Line: 1, Column: 52, EndLine: 1, EndColumn: 59},
				},
			},
			// ---- Dimension 4: fix range — a property's arrow keeps its own range, which never covered the computed key ----
			{
				Code:   `class K { [(() => { return; })()] = () => { return; }; }`,
				Output: []string{`class K { [(() => {  })()] = () => {  }; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 21, EndLine: 1, EndColumn: 28},
					{MessageId: "unnecessaryReturn", Line: 1, Column: 45, EndLine: 1, EndColumn: 52},
				},
			},
			// ---- Dimension 4: fix range — a chain of computed keys stays one pass however deep it goes ----
			{
				Code:   `class A { [(() => { class B { [(() => { class C { [(() => { return; })()]() { return; } } })()]() { return; } } })()]() { return; } }`,
				Output: []string{`class A { [(() => { class B { [(() => { class C { [(() => {  })()]() {  } } })()]() {  } } })()]() {  } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 61, EndLine: 1, EndColumn: 68},
					{MessageId: "unnecessaryReturn", Line: 1, Column: 79, EndLine: 1, EndColumn: 86},
					{MessageId: "unnecessaryReturn", Line: 1, Column: 101, EndLine: 1, EndColumn: 108},
					{MessageId: "unnecessaryReturn", Line: 1, Column: 123, EndLine: 1, EndColumn: 130},
				},
			},
			// ---- Dimension 4: fix range — the key's own return retains the whole key expression, so its nested method waits a pass, as upstream does ----
			{
				Code: `class K { [(() => { class I { [(() => { return; })()]() { return; } } return; })()]() { return; } }`,
				Output: []string{
					`class K { [(() => { class I { [(() => { return; })()]() { return; } }  })()]() {  } }`,
					`class K { [(() => { class I { [(() => {  })()]() {  } }  })()]() {  } }`,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 41, EndLine: 1, EndColumn: 48},
					{MessageId: "unnecessaryReturn", Line: 1, Column: 59, EndLine: 1, EndColumn: 66},
					{MessageId: "unnecessaryReturn", Line: 1, Column: 71, EndLine: 1, EndColumn: 78},
					{MessageId: "unnecessaryReturn", Line: 1, Column: 89, EndLine: 1, EndColumn: 96},
				},
			},
			// ---- Dimension 4: fix range — a key holding a function expression ----
			{
				Code:   `class K { [(function () { return; })()]() { return; } }`,
				Output: []string{`class K { [(function () {  })()]() {  } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 27, EndLine: 1, EndColumn: 34},
					{MessageId: "unnecessaryReturn", Line: 1, Column: 45, EndLine: 1, EndColumn: 52},
				},
			},
			// ---- Dimension 4: fix range — a key holding a method of its own ----
			{
				Code:   `class K { [({ m() { return; } }).m()]() { return; } }`,
				Output: []string{`class K { [({ m() {  } }).m()]() {  } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 21, EndLine: 1, EndColumn: 28},
					{MessageId: "unnecessaryReturn", Line: 1, Column: 43, EndLine: 1, EndColumn: 50},
				},
			},
			// ---- Dimension 4: fix range — type parameters sit between the key and the parameter list ----
			{
				Code:   `class K { [(() => { return; })()]<T>() { return; } }`,
				Output: []string{`class K { [(() => {  })()]<T>() {  } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 21, EndLine: 1, EndColumn: 28},
					{MessageId: "unnecessaryReturn", Line: 1, Column: 42, EndLine: 1, EndColumn: 49},
				},
			},
			// ---- Dimension 4: fix range — modifiers ahead of the key don't move the retained range ----
			{
				Code:   `class K { async [(() => { return; })()]() { return; } }`,
				Output: []string{`class K { async [(() => {  })()]() {  } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 27, EndLine: 1, EndColumn: 34},
					{MessageId: "unnecessaryReturn", Line: 1, Column: 45, EndLine: 1, EndColumn: 52},
				},
			},
			// ---- Dimension 4: fix range — modifiers ahead of the key don't move the retained range ----
			{
				Code:   `class K { *[(() => { return; })()]() { return; } }`,
				Output: []string{`class K { *[(() => {  })()]() {  } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 22, EndLine: 1, EndColumn: 29},
					{MessageId: "unnecessaryReturn", Line: 1, Column: 40, EndLine: 1, EndColumn: 47},
				},
			},
			// ---- Dimension 4: fix range — an object literal accessor with a computed key ----
			{
				Code:   `const o = { get [(() => { return; })()]() { return; } };`,
				Output: []string{`const o = { get [(() => {  })()]() {  } };`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 27, EndLine: 1, EndColumn: 34},
					{MessageId: "unnecessaryReturn", Line: 1, Column: 45, EndLine: 1, EndColumn: 52},
				},
			},
			// ---- Dimension 4: fix range — a decorator and a computed key on the same accessor ----
			{
				Code:   `class K { @dec(() => { return; }) get [(() => { return; })()]() { return; } }`,
				Output: []string{`class K { @dec(() => {  }) get [(() => {  })()]() {  } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 24, EndLine: 1, EndColumn: 31},
					{MessageId: "unnecessaryReturn", Line: 1, Column: 49, EndLine: 1, EndColumn: 56},
					{MessageId: "unnecessaryReturn", Line: 1, Column: 67, EndLine: 1, EndColumn: 74},
				},
			},
			// ---- Dimension 4: fix range — a default parameter value is inside the retained range, so its return waits for the next pass, as upstream does ----
			{
				Code: `class K { m(x = () => { return; }) { return; } }`,
				Output: []string{
					`class K { m(x = () => { return; }) {  } }`,
					`class K { m(x = () => {  }) {  } }`,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 25, EndLine: 1, EndColumn: 32},
					{MessageId: "unnecessaryReturn", Line: 1, Column: 38, EndLine: 1, EndColumn: 45},
				},
			},
			// ---- Dimension 4: fix range — a return nested in the method body still retains the whole body ----
			{
				Code:   `class K { [(() => { return; })()]() { if (c) { return; } } }`,
				Output: []string{`class K { [(() => {  })()]() { if (c) {  } } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 21, EndLine: 1, EndColumn: 28},
					{MessageId: "unnecessaryReturn", Line: 1, Column: 48, EndLine: 1, EndColumn: 55},
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
			Program:      lintprogram.NewFromCompiler(program),
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

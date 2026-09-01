// TestPreferAddEventListenerOptionsExtras covers tsgo-specific edge shapes,
// real-user examples, and every reachable upstream branch. The complete
// upstream suite remains isolated in the sibling upstream test file.
package prefer_add_event_listener_options_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	prefer_add_event_listener_options "github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_add_event_listener_options"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferAddEventListenerOptionsExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_add_event_listener_options.PreferAddEventListenerOptionsRule,
		[]rule_tester.ValidTestCase{
			// Locks in upstream isMethodCall() arm 1: a bare call is not a method call.
			{Code: `addEventListener("click", listener, true)`, FileName: "file.js"},
			// Locks in upstream isMethodCall() name constraint.
			{Code: `window.notAddEventListener("click", listener, true)`, FileName: "file.js"},
			// Locks in upstream argumentsLength === 3.
			{Code: `window.addEventListener()`, FileName: "file.js"},
			{Code: `window.addEventListener("click", listener, true, extra)`, FileName: "file.js"},

			// ---- Dimension 4: element and private access keys are excluded ----
			{Code: `window['addEventListener']("click", listener, true)`, FileName: "file.js"},
			{Code: "window[`addEventListener`](\"click\", listener, true)", FileName: "file.js"},
			{Code: `window[0]("click", listener, true)`, FileName: "file.js"},
			{Code: `window[Symbol.addEventListener]("click", listener, true)`, FileName: "file.js"},
			{Code: `class C { #addEventListener() {} method() { this.#addEventListener("click", listener, true) } }`, FileName: "file.js"},

			// ---- Dimension 4: optional member and optional call forms are excluded ----
			{Code: `((window))?.addEventListener("click", listener, true)`, FileName: "file.ts", Tsx: false},
			{Code: `(window.addEventListener)?.("click", listener, true)`, FileName: "file.ts", Tsx: false},

			// ---- Dimension 4: TypeScript wrappers around the inspected boolean remain visible upstream ----
			{Code: `window.addEventListener("click", listener, true as boolean)`, FileName: "file.ts", Tsx: false},
			{Code: `window.addEventListener("click", listener, true satisfies boolean)`, FileName: "file.ts", Tsx: false},
			{Code: `window.addEventListener("click", listener, true!)`, FileName: "file.ts", Tsx: false},

			// ---- Graceful degradation: spread arguments disqualify the call ----
			{Code: `window.addEventListener(...args, listener, true)`, FileName: "file.js"},
			{Code: `window.addEventListener("click", listener, ...options)`, FileName: "file.js"},

			// Locks in upstream boolean-literal gate: other literal kinds are ignored.
			{Code: `window.addEventListener("click", listener, 1)`, FileName: "file.js"},
			{Code: `window.addEventListener("click", listener, "true")`, FileName: "file.js"},

			// ---- Real-user: #2067 already uses the preferred extensible object form ----
			{Code: `el.addEventListener('click', () => {}, {capture: true})`, FileName: "file.js"},

			// N/A: declaration/container variants; the rule only inspects calls.
			// N/A: object-literal spread/rest and body-less declarations; neither is traversed specially.
			// N/A: ancestor walks; the rule does not inspect parents or enclosing scopes.
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: single- and multi-level parenthesized receivers are transparent ----
			{
				Code:     `((window)).addEventListener("click", listener, true)`,
				FileName: "file.js",
				Output:   []string{`((window)).addEventListener("click", listener, {capture: true})`},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID}},
			},

			// ---- Dimension 4: TypeScript receiver wrappers do not affect the method match ----
			{
				Code:     `window!.addEventListener("click", listener, false)`,
				FileName: "file.ts",
				Tsx:      false,
				Output:   []string{`window!.addEventListener("click", listener, {capture: false})`},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID}},
			},
			{
				Code:     `(window as Window).addEventListener("click", listener, true)`,
				FileName: "file.ts",
				Tsx:      false,
				Output:   []string{`(window as Window).addEventListener("click", listener, {capture: true})`},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID}},
			},
			{
				Code:     `(window satisfies Window).addEventListener("click", listener, true)`,
				FileName: "file.ts",
				Tsx:      false,
				Output:   []string{`(window satisfies Window).addEventListener("click", listener, {capture: true})`},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID}},
			},

			// ---- Dimension 4: parenthesized callee is transparent in ESTree ----
			{
				Code:     `(window.addEventListener)("click", listener, true)`,
				FileName: "file.js",
				Output:   []string{`(window.addEventListener)("click", listener, {capture: true})`},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID}},
			},

			// ---- Dimension 4: JavaScript JSDoc casts are invisible to ESTree ----
			{
				Code:     `window.addEventListener("click", listener, /** @type {boolean} */ (true))`,
				FileName: "file.js",
				Output:   []string{`window.addEventListener("click", listener, /** @type {boolean} */ ({capture: true}))`},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID}},
			},
			{
				Code:     `/** @type {any} */ (window.addEventListener)("click", listener, true)`,
				FileName: "file.js",
				Output:   []string{`/** @type {any} */ (window.addEventListener)("click", listener, {capture: true})`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Message:   expectedMessage("true"),
					Line:      1,
					Column:    65,
					EndLine:   1,
					EndColumn: 69,
				}},
			},
			{
				Code:     `/** @satisfies {any} */ (window.addEventListener)("click", listener, false)`,
				FileName: "file.js",
				Output:   []string{`/** @satisfies {any} */ (window.addEventListener)("click", listener, {capture: false})`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Message:   expectedMessage("false"),
					Line:      1,
					Column:    70,
					EndLine:   1,
					EndColumn: 75,
				}},
			},

			// Locks in upstream CallExpression handling with TypeScript type arguments.
			{
				Code:     `window.addEventListener<Event>("click", listener, true)`,
				FileName: "file.ts",
				Tsx:      false,
				Output:   []string{`window.addEventListener<Event>("click", listener, {capture: true})`},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID}},
			},

			// Locks in upstream replacement branch 1: true maps to capture: true.
			{
				Code:     `target.addEventListener("x", handler, true)`,
				FileName: "file.js",
				Output:   []string{`target.addEventListener("x", handler, {capture: true})`},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID, Message: expectedMessage("true")}},
			},
			// Locks in upstream replacement branch 2: false maps to capture: false.
			{
				Code:     `target.addEventListener("x", handler, false)`,
				FileName: "file.js",
				Output:   []string{`target.addEventListener("x", handler, {capture: false})`},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID, Message: expectedMessage("false")}},
			},

			// ---- Dimension 4: same-kind nesting reports each matching call independently ----
			{
				Code:     `window.addEventListener("outer", () => window.addEventListener("inner", listener, false), true)`,
				FileName: "file.js",
				Output:   []string{`window.addEventListener("outer", () => window.addEventListener("inner", listener, {capture: false}), {capture: true})`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: messageID},
					{MessageId: messageID},
				},
			},

			// ---- Real-user: #2067 legacy boolean signature with an inline listener ----
			{
				Code:     `el.addEventListener('click', () => {}, true)`,
				FileName: "file.js",
				Output:   []string{`el.addEventListener('click', () => {}, {capture: true})`},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID}},
			},
		},
	)
}

// TestPreferAddEventListenerOptionsEditDemand verifies that the deferred
// autofix never changes the diagnostic identity and is only materialized for
// consumers that request autofixes.
func TestPreferAddEventListenerOptionsEditDemand(t *testing.T) {
	t.Parallel()

	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/edit-demand.ts",
		Path:     "/edit-demand.ts",
	}, `window.addEventListener("click", listener, true);`, core.ScriptKindTS)

	run := func(demand rule.EditDemand) rule.RuleDiagnostic {
		t.Helper()

		comments := rule.NewCommentStore(sourceFile)
		diagnostics := make([]rule.RuleDiagnostic, 0, 1)
		ctx := rule.RuleContext{
			SourceFile:     sourceFile,
			Comments:       comments,
			DisableManager: rule.NewDisableManager(sourceFile, comments),
		}.WithDiagnosticConsumer(
			prefer_add_event_listener_options.PreferAddEventListenerOptionsRule.Name,
			rule.SeverityError,
			rule.DiagnosticConsumer{
				Demand: demand,
				Report: func(diagnostic rule.RuleDiagnostic) {
					diagnostics = append(diagnostics, diagnostic)
				},
			},
		)

		listeners := prefer_add_event_listener_options.PreferAddEventListenerOptionsRule.Run(ctx, nil)
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if listener := listeners[node.Kind]; listener != nil {
				listener(node)
			}
			return node.ForEachChild(visit)
		}
		sourceFile.AsNode().ForEachChild(visit)
		if len(diagnostics) != 1 {
			t.Fatalf("demand %d: diagnostics = %d, want 1", demand, len(diagnostics))
		}
		return diagnostics[0]
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
	for demand, diagnostic := range map[rule.EditDemand]rule.RuleDiagnostic{
		rule.EditDemandNone:       diagnosticsOnly,
		rule.EditDemandAutofix:    autofixOnly,
		rule.EditDemandSuggestion: suggestionOnly,
	} {
		if got, want := withoutEdits(diagnostic), withoutEdits(allEdits); !reflect.DeepEqual(got, want) {
			t.Errorf("demand %d changed diagnostic identity:\ngot  %#v\nwant %#v", demand, got, want)
		}
	}
	if diagnosticsOnly.FixesPtr != nil || suggestionOnly.FixesPtr != nil {
		t.Fatal("a non-autofix demand materialized fixes")
	}
	if autofixOnly.FixesPtr == nil || allEdits.FixesPtr == nil ||
		!reflect.DeepEqual(*autofixOnly.FixesPtr, *allEdits.FixesPtr) {
		t.Fatal("autofix and all-edits demands produced different fixes")
	}
	if diagnosticsOnly.Suggestions != nil || autofixOnly.Suggestions != nil ||
		suggestionOnly.Suggestions != nil || allEdits.Suggestions != nil {
		t.Fatal("autofix-only rule materialized suggestions")
	}
}

package no_nested_ternary_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_nested_ternary"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func linesE(parts ...string) string {
	return strings.Join(parts, "\n")
}

// TestNoNestedTernaryExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise. Each case carries an inline comment pointing at
// the specific branch / Dimension 4 row / tsgo AST quirk it covers, so future
// refactors can't silently regress them without breaking a named lock-in.
func TestNoNestedTernaryExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_nested_ternary.NoNestedTernaryRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: paren wrappers on the outer ----
			// Multi-level paren wrappers on the outer test position.
			// `((a?b:c))` is parsed as ParenExpr(ParenExpr(CondExpr)).
			// The inner's effective parent is the outer CondExpr; the inner is
			// "source-parenthesized" because its direct parent is a ParenExpr.
			rule_tester.ValidTestCase{Code: `const foo = ((a ? b : c)) ? d : e;`},

			// ---- Real-user: ternary in template literal (issue #663 family) ----
			rule_tester.ValidTestCase{Code: "const s = `${a ? b : c}`;"},

			// ---- Real-user: ternary in object property value ----
			rule_tester.ValidTestCase{Code: `const o = { x: a ? b : c };`},

			// ---- Real-user: ternary as function argument (no nesting) ----
			rule_tester.ValidTestCase{Code: `f(a ? b : c);`},

			// ---- Real-user: nested ternary in JSX expression container ----
			rule_tester.ValidTestCase{Code: `const el = <div>{a ? (b ? c : d) : e}</div>;`, Tsx: true},

			// ---- Branch lock-in: 2-level nesting with paren-wrapped inner is valid ----
			// `a ? (b ? c : d) : e` — the inner is parenthesized and only one
			// level deep, so the rule accepts it. Locks in the early-return
			// (outer consequent is CondExpr → outer skipped) and the
			// nestLevel === 1 + parenthesized → no report branch.
			rule_tester.ValidTestCase{Code: `const foo = a ? (b ? c : d) : e;`},

			// ---- Branch lock-in: nestLevel === 0 (no CondExpr ancestors) ----
			// A bare top-level ternary in a paren must NOT report, even with
			// arbitrary paren wrappers around it.
			rule_tester.ValidTestCase{Code: `const x = (((a ? b : c)));`},

			// ---- Branch lock-in: nestLevel === 1 with paren, multi-line ----
			// The inner ternary is parenthesized even when split across lines.
			rule_tester.ValidTestCase{Code: linesE(
				`const x = a ?`,
				`	(`,
				`		b ? c : d`,
				`	) : e;`,
			)},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: optional chain in the test position ----
			// `a?.b ? c ? d : e : f` — the consequent is a CondExpr so the outer
			// early-returns; the inner has nestLevel=1, is NOT parenthesized, and
			// the autofix wraps it.
			rule_tester.InvalidTestCase{
				Code:   `const foo = a?.b ? c ? d : e : f;`,
				Output: []string{`const foo = a?.b ? (c ? d : e) : f;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: shouldParenMessageID}},
			},

			// ---- Dimension 4: non-null assertion in the test position ----
			// `a! ? b ? c : d : e` — same shape, asserts that the non-null
			// assertion does not break the early-return / nestLevel logic.
			rule_tester.InvalidTestCase{
				Code:   `const foo = a! ? b ? c : d : e;`,
				Output: []string{`const foo = a! ? (b ? c : d) : e;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: shouldParenMessageID}},
			},

			// ---- Dimension 4: parenthesized type assertion in the test ----
			// `(a as any) ? b ? c : d : e` — paren-wrapped type assertion is a
			// ParenExpr(AsExpr). The inner CondExpr is not parenthesized, so
			// should-parenthesized fires.
			rule_tester.InvalidTestCase{
				Code:   `const foo = (a as any) ? b ? c : d : e;`,
				Output: []string{`const foo = (a as any) ? (b ? c : d) : e;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: shouldParenMessageID}},
			},

			// ---- Branch lock-in: 3-level nesting with paren-wrapped inner reports too-deep ----
			// `a ? (b ? c : (d ? e : f)) : g` — for the innermost, walking up
			// through parens gives nestLevel=2, isParenthesized=true.
			// nestLevel > 1 fires, target = node (the innermost). Locks in
			// the "parenthesized doesn't suppress too-deep" branch.
			rule_tester.InvalidTestCase{
				Code:   `const foo = a ? (b ? c : (d ? e : f)) : g;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: tooDeepMessageID, Line: 1, Column: 27, EndLine: 1, EndColumn: 36}},
			},

			// ---- Branch lock-in: nestLevel === 3 reports on ancestors[0] ----
			// For 3 ternary ancestors, `ancestors[nestLevel-3] === ancestors[0]`
			// is the innermost CondExpr parent. The innermost itself is wrapped
			// in parens, so it's the 1st CondExpr parent that gets the report.
			// Locks in upstream arm: `nestLevel > 2 ? ancestors[nestLevel-3] : node`
			// with nestLevel === 3.
			rule_tester.InvalidTestCase{
				Code: linesE(
					`const foo = a ? b : (`,
					`	c ? d : (`,
					`		e ? f : (g ? h : i)`,
					`	)`,
					`);`,
				),
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: tooDeepMessageID, Line: 3, Column: 3}},
			},

			// ---- Branch lock-in: nestLevel === 4 reports on ancestors[1] ----
			// For 4 ternary ancestors, ancestors[1] is the 2nd CondExpr parent
			// (one level up from the innermost). The innermost is in a paren
			// inside another paren, so the inner CondExpr parent is 1 level up.
			rule_tester.InvalidTestCase{
				Code: linesE(
					`const foo = a ? b : (`,
					`	c ? d : (`,
					`		e ? f : (`,
					`			g ? h : (i ? j : k)`,
					`		)`,
					`	)`,
					`);`,
				),
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: tooDeepMessageID, Line: 3, Column: 3}},
			},

			// ---- Branch lock-in: early-return when consequent is CondExpr ----
			// `a ? b ? c : d : e` — outer's consequent is a CondExpr, so the
			// outer early-returns. Only the inner reports (1 error, not 2).
			// Locks in upstream arm: `if ([test, consequent, alternate].some(...)) return`.
			rule_tester.InvalidTestCase{
				Code:   `const foo = a ? b ? c : d : e;`,
				Output: []string{`const foo = a ? (b ? c : d) : e;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: shouldParenMessageID, Line: 1, Column: 17}},
			},

			// ---- Real-user: nested ternary in template literal expression ----
			// Common pattern in production: a `?` inside a template, where the
			// alternate of an outer ternary is itself a ternary.
			rule_tester.InvalidTestCase{
				Code:   "const s = `${a ? b : c ? d : e}`;",
				Output: []string{"const s = `${a ? b : (c ? d : e)}`;"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: shouldParenMessageID}},
			},

			// ---- Real-user: nested ternary in object property ----
			rule_tester.InvalidTestCase{
				Code:   `const o = { x: a ? b : c ? d : e };`,
				Output: []string{`const o = { x: a ? b : (c ? d : e) };`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: shouldParenMessageID}},
			},

			// ---- Real-user: nested ternary as function argument ----
			rule_tester.InvalidTestCase{
				Code:   `f(a ? b : c ? d : e);`,
				Output: []string{`f(a ? b : (c ? d : e));`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: shouldParenMessageID}},
			},
		},
	)
}

func TestNoNestedTernaryEditDemand(t *testing.T) {
	t.Parallel()

	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/edit-demand.ts",
		Path:     "/edit-demand.ts",
	}, "const value = outer ? inner ? yes : no : fallback;", core.ScriptKindTS)

	run := func(demand rule.EditDemand) rule.RuleDiagnostic {
		t.Helper()

		comments := rule.NewCommentStore(sourceFile)
		diagnostics := make([]rule.RuleDiagnostic, 0, 1)
		ctx := rule.RuleContext{
			SourceFile:     sourceFile,
			Comments:       comments,
			DisableManager: rule.NewDisableManager(sourceFile, comments),
		}.WithDiagnosticConsumer(
			no_nested_ternary.NoNestedTernaryRule.Name,
			rule.SeverityError,
			rule.DiagnosticConsumer{
				Demand: demand,
				Report: func(diagnostic rule.RuleDiagnostic) {
					diagnostics = append(diagnostics, diagnostic)
				},
			},
		)

		listeners := no_nested_ternary.NoNestedTernaryRule.Run(ctx, nil)
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

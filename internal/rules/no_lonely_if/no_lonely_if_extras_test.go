// TestNoLonelyIfExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise. Each case carries an inline comment pointing
// at the specific branch / Dimension row / real-user issue it covers, so
// future refactors can't silently regress them without breaking a named
// lock-in.
//
// Dimension 4 (universal edge shapes) note: no-lonely-if is purely
// structural — it never reads a receiver expression, a property/key form, or
// a declaration/container shape. The "Receiver / expression wrappers",
// "Access / key forms", and "Declaration / container forms" rows are N/A for
// this rule. The applicable rows are "Nesting / traversal boundaries" and
// "Graceful degradation" (empty bodies), both covered below.
package no_lonely_if

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoLonelyIfExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoLonelyIfRule,
		[]rule_tester.ValidTestCase{
			// Locks in upstream arm: `parent === grandparent.alternate` is false
			// because parent is the grandparent's *consequent* block, not its
			// alternate.
			{Code: "if (a) { if (b) {} } else { baz(); }"},

			// Locks in upstream arm: `grandparent.type === "IfStatement"` is
			// false (grandparent is a WhileStatement).
			{Code: "while (a) { if (b) {} }"},

			// Locks in upstream arm: `parent.body.length === 1` is false (the
			// else block has two statements).
			{Code: "if (a) {} else { if (b) {} foo(); }"},

			// Locks in upstream arm: `parent && parent.type === "BlockStatement"`
			// is false — the if has no parent block at all (top-level statement).
			{Code: "if (b) {}"},

			// Locks in upstream arm: `parent && parent.type === "BlockStatement"`
			// is false — parent is another IfStatement (an `else if` chain, no
			// block wrapper).
			{Code: "if (a) {} else if (b) {} else if (c) {}"},

			// ---- Real-user: eslint/eslint#19033 ----
			// "no-lonely-if has incorrect fix when nested inside parent if
			// without BlockStatement". Reporting here would allow an unsafe
			// autofix that makes the outer `else` unreachable, so
			// areBracesNecessary's hasUnsafeIf/isFollowedByElseKeyword gate must
			// suppress the report. (This is the exact scenario the upstream fix
			// in eslint/eslint#19087 added regression coverage for.)
			{Code: "if (true)\n" +
				"\tif (false) {}\n" +
				"\telse { if (false) {} }\n" +
				"else throw Error('unreachable');"},

			// Locks in HasUnsafeIf's ForStatement recursion arm: the dangling
			// `if` is reached through a `for (;;)` loop in the else-chain.
			{Code: "if (x)\n" +
				"  if (a) {} else { if (b) {} else for (;;) if (c) foo(); }\n" +
				"else\n" +
				"  bar();"},

			// Locks in HasUnsafeIf's WhileStatement recursion arm: same shape,
			// through a `while` loop.
			{Code: "if (x)\n" +
				"  if (a) {} else { if (b) {} else while (c) if (d) foo(); }\n" +
				"else\n" +
				"  bar();"},

			// Locks in HasUnsafeIf's LabeledStatement recursion arm: same
			// shape, through a labeled statement.
			{Code: "if (x)\n" +
				"  if (a) {} else { if (b) {} else lbl: if (c) foo(); }\n" +
				"else\n" +
				"  bar();"},

			// ---- Dimension 2: deeply nested (4-level) lonely-if chain, only the ----
			// outermost is a lonely-if relative to its own immediate else block;
			// each level reports independently (verified in the invalid chain
			// below). This valid case checks a chain that bottoms out in an
			// `else if` (no trailing lonely block), so nothing is reported.
			{Code: "if (a) {} else if (b) {} else if (c) {} else if (d) {} else {}"},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 2: multiple independent lonely-ifs must not bleed ----
			// into each other's boundary — both the outer and the inner
			// lonely-if are reported, each fixed independently.
			{
				Code: "if (a) {\n" +
					"} else {\n" +
					"  if (b) {\n" +
					"  } else {\n" +
					"    if (c) {}\n" +
					"  }\n" +
					"}",
				Output: []string{
					"if (a) {\n" +
						"} else if (b) {\n" +
						"  } else {\n" +
						"    if (c) {}\n" +
						"  }",
					"if (a) {\n" +
						"} else if (b) {\n" +
						"  } else if (c) {}",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLonelyIf", Line: 3, Column: 3},
					{MessageId: "unexpectedLonelyIf", Line: 5, Column: 5},
				},
			},

			// ---- Dimension 3: consequent is an EmptyStatement; its own last ----
			// token is the semicolon, so the ASI-hazard check's
			// `lastIfToken.Kind != KindSemicolonToken` guard is false and the fix
			// applies normally.
			{
				Code:   "if (a) {} else { if (b); }",
				Output: []string{"if (a) {} else if (b);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLonelyIf", Line: 1, Column: 18},
				},
			},

			// ---- Dimension 3: no token follows the else block at all (end of ----
			// file); `hasTokenAfter` is false so the ASI-hazard check is skipped
			// entirely and the fix applies.
			{
				Code:   "if (foo) {} else { if (bar) baz() }",
				Output: []string{"if (foo) {} else if (bar) baz()"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLonelyIf", Line: 1, Column: 20},
				},
			},

			// ---- Dimension 3: token after the else block starts with `(` on a ----
			// different line than the consequent (IIFE-call ASI hazard); not
			// fixed.
			{
				Code: "if (foo) {\n" +
					"} else {\n" +
					"  if (bar) baz()\n" +
					"}\n" +
					"(qux)();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLonelyIf", Line: 3, Column: 3},
				},
			},

			// ---- Dimension 3: token after the else block starts with `/` on a ----
			// different line (regex-literal ASI hazard); not fixed.
			{
				Code: "if (foo) {\n" +
					"} else {\n" +
					"  if (bar) baz()\n" +
					"}\n" +
					"/regex/.test(qux);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLonelyIf", Line: 3, Column: 3},
				},
			},

			// ---- Dimension 4 (graceful degradation): the whitespace the fixer ----
			// scans for interference is ECMAScript's, not Go's. U+FEFF is
			// whitespace to JavaScript, so the braces come off…
			{
				Code:   "if (a) {} else {\uFEFFif (b) c();\uFEFF}",
				Output: []string{"if (a) {} else if (b) c();"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLonelyIf", Line: 1, Column: 18},
				},
			},

			// …and U+0085 is not, so the interference check sees a character
			// between the brace and the `if` and the diagnostic carries no fix.
			{
				Code: "if (a) {} else {\u0085if (b) c();\u0085}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLonelyIf", Line: 1, Column: 18},
				},
			},

			// ---- Dimension 3: `else{` with no space before the opening brace; ----
			// the fixer must insert a leading space so `else` and the replacement
			// `if` text don't fuse into `elseif`.
			{
				Code:   "if (foo) {} else{ if (bar) baz(); }",
				Output: []string{"if (foo) {} else if (bar) baz();"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedLonelyIf", Line: 1, Column: 19},
				},
			},
		},
	)
}

// TestNoLonelyIfEditDemand verifies that the diagnostic (range, message) is
// identical across all four edit demands, that the autofix is only built
// under EditDemandAutofix/EditDemandAll (no-lonely-if never emits
// suggestions), and that EditDemandAll's fix matches EditDemandAutofix's.
func TestNoLonelyIfEditDemand(t *testing.T) {
	code := "if (a) {} else { if (b) baz(); }"
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/no-lonely-if-edit-demand.ts",
		Path:     tspath.Path("/no-lonely-if-edit-demand.ts"),
	}, code, core.ScriptKindTS)

	outerIf := sourceFile.Statements.Nodes[0]
	elseBlock := outerIf.AsIfStatement().ElseStatement
	innerIf := elseBlock.AsBlock().Statements.Nodes[0]

	run := func(demand rule.EditDemand) rule.RuleDiagnostic {
		comments := rule.NewCommentStore(sourceFile)
		var diagnostics []rule.RuleDiagnostic
		ctx := rule.RuleContext{
			SourceFile:     sourceFile,
			Comments:       comments,
			DisableManager: rule.NewDisableManager(sourceFile, comments),
		}.WithDiagnosticConsumer(NoLonelyIfRule.Name, rule.SeverityError, rule.DiagnosticConsumer{
			Demand: demand,
			Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			},
		})

		NoLonelyIfRule.Run(ctx, nil)[ast.KindIfStatement](innerIf)
		if len(diagnostics) != 1 {
			t.Fatalf("diagnostics = %d, want 1", len(diagnostics))
		}
		return diagnostics[0]
	}

	none := run(rule.EditDemandNone)
	autofix := run(rule.EditDemandAutofix)
	suggestion := run(rule.EditDemandSuggestion)
	all := run(rule.EditDemandAll)

	for name, d := range map[string]rule.RuleDiagnostic{"autofix": autofix, "suggestion": suggestion, "all": all} {
		if d.Range != none.Range || !reflect.DeepEqual(d.Message, none.Message) {
			t.Fatalf("%s: diagnostic changed with edit demand: none=%#v %s=%#v", name, none, name, d)
		}
	}

	if none.FixesPtr != nil {
		t.Fatalf("EditDemandNone consumer received fixes: %#v", none.Fixes())
	}
	if len(suggestion.Fixes()) != 0 {
		t.Fatalf("EditDemandSuggestion consumer received fixes: %#v", suggestion.Fixes())
	}
	if len(autofix.Fixes()) != 1 {
		t.Fatalf("EditDemandAutofix fix count = %d, want 1", len(autofix.Fixes()))
	}
	if !reflect.DeepEqual(autofix.Fixes(), all.Fixes()) {
		t.Fatalf("EditDemandAll fixes = %#v, want %#v (same as EditDemandAutofix)", all.Fixes(), autofix.Fixes())
	}
}

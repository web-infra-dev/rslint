// TestRequireUnicodeRegexpExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it
// covers, so future refactors can't silently regress them without breaking a
// named lock-in.
//
// Dimension walk notes for require-unicode-regexp:
//   - Dimension 4 (declaration/container forms): N/A — the rule targets
//     regex literals and RegExp() calls, never a function or class
//     declaration.
//   - Dimension 4 (graceful degradation: SpreadAssignment/RestElement inside
//     an object literal or binding pattern, empty destructuring pattern,
//     overload signatures/abstract/declare members): N/A — none of these
//     shapes can appear as a regex literal, a RegExp call, or one of its two
//     arguments.
package require_unicode_regexp

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

func TestRequireUnicodeRegexpExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&RequireUnicodeRegexpRule,
		[]rule_tester.ValidTestCase{
			// ---- Locks in upstream trackMap arm: patternNode SpreadElement skips
			// the whole call, even with further arguments present ----
			{Code: "RegExp(...args, 'gi')"},
			// ---- Dimension 2: shadowing a global-object alias suppresses the
			// member-access form, same as shadowing RegExp itself ----
			{Code: "function f(globalThis) { return new globalThis.RegExp('foo') }"},
			// ---- Locks in the isDeclaredGlobal callback's negative branch: an
			// explicitly un-declared RegExp global is not a builtin callee ----
			{Code: "RegExp('foo')", Globals: map[string]any{"RegExp": "off"}},
			// ---- Dimension 3: a non-static (side-effecting) flags argument is
			// never reported, matching a non-literal flags identifier ----
			{Code: "new RegExp('foo', getFlags())"},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: optional call — tsgo keeps `?.` as a flag on the
			// CallExpression rather than an ESTree ChainExpression wrapper ----
			{
				Code: "RegExp?.('foo')",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 16,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: `RegExp?.('foo', "u")`}},
					},
				},
			},
			// ---- Dimension 4: TS non-null assertion on the callee ----
			{
				Code: "RegExp!('foo')",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 15,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: `RegExp!('foo', "u")`}},
					},
				},
			},
			// ---- Dimension 4: TS `as` type-expression wrapper on the callee ----
			{
				Code: "(RegExp as any)('foo')",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 23,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: `(RegExp as any)('foo', "u")`}},
					},
				},
			},
			// ---- Dimension 4: single and multi-level parenthesized callee ----
			{
				Code: "(RegExp)('foo')",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 16,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: `(RegExp)('foo', "u")`}},
					},
				},
			},
			{
				Code: "((RegExp))('foo')",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 18,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: `((RegExp))('foo', "u")`}},
					},
				},
			},
			// ---- Dimension 4: element access on a global-object alias with a
			// static string-literal key ----
			{
				Code: "globalThis['RegExp']('foo')",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 28,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: `globalThis['RegExp']('foo', "u")`}},
					},
				},
			},
			// ---- Dimension 4: element access on a global-object alias with a
			// static no-substitution template key ----
			{
				Code: "globalThis[`RegExp`]('foo')",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 28,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: "globalThis[`RegExp`]('foo', \"u\")"}},
					},
				},
			},
			// ---- Dimension 2: detection reaches a RegExp call 3+ scopes deep
			// (class method -> nested function) without bleeding into or out of
			// the intervening scopes ----
			{
				Code: "class C { m() { function f() { return new RegExp('foo') } return f } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 39, EndLine: 1, EndColumn: 56,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: `class C { m() { function f() { return new RegExp('foo', "u") } return f } }`}},
					},
				},
			},
			// ---- Dimension 3: a comment between the pattern and flags arguments
			// is preserved by the fix, which only replaces the flags node's own
			// range ----
			{
				Code: "new RegExp('foo', /* flags */ 'gimy')",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 38,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: "new RegExp('foo', /* flags */ 'gimyu')"}},
					},
				},
			},
			// ---- Dimension 3: multi-line call — the diagnostic spans the whole
			// call across lines and the fix targets only the flags node's own line ----
			{
				Code: "new RegExp(\n  'foo',\n  'gimy'\n)",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 4, EndColumn: 2,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: "new RegExp(\n  'foo',\n  'gimyu'\n)"}},
					},
				},
			},
			// Locks in upstream Program() listener arm: the comma operator's
			// value is its right operand — the left operand (an unresolvable
			// identifier here) does not gate evaluation, matching
			// getStringIfConstant's SequenceExpression handling.
			{
				Code: "new RegExp('foo', (unknownVar, 'gi'))",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 38},
				},
			},
			// Locks in isValidWithUnicodeFlag's ecmaVersion <= 5 arm for the
			// default (unset) requireFlag: pre-ES2015 environments never get a
			// `u`-flag suggestion.
			{
				Code:            "/foo/",
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 5},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
				},
			},
			// Locks in isValidWithUnicodeFlag's ecmaVersion <= 2023 arm at its
			// exact boundary (2023, one edition below the `v` flag's ES2024
			// introduction) — upstream only tests the boundary at ecmaVersion 6.
			{
				Code:            "/foo/",
				Options:         []any{map[string]any{"requireFlag": "v"}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2023},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "requireVFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 6},
				},
			},
			// Locks in the fixer's "we intentionally don't suggest concatenating
			// to non-literals" arm for a PropertyAccessExpression flags node that
			// evaluates successfully (so the report itself fires) but is
			// syntactically ineligible for a fix — upstream's own tests only
			// exercise a bare identifier there.
			{
				Code: "const obj = { flags: 'gi' }; new RegExp('foo', obj.flags)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "requireUFlag", Line: 1, Column: 30, EndLine: 1, EndColumn: 58},
				},
			},
			// ---- Dimension 2: nested RegExp calls are each evaluated
			// independently — the outer call's non-static flags node (a nested
			// `new RegExp(...)`) suppresses only the outer report ----
			{
				Code: "new RegExp('foo', new RegExp('bar'))",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 19, EndLine: 1, EndColumn: 36,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: `new RegExp('foo', new RegExp('bar', "u"))`}},
					},
				},
			},
			// Locks in the fixer's append branch (flags does not contain the
			// opposite flag) staying unguarded by the raw-backslash check that
			// only applies to the includes(flag) branch: a `g`-escaped
			// template still gets a plain `u` appended.
			{
				Code: "new RegExp('foo', `\\u0067`)",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 28,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: "new RegExp('foo', `\\u0067u`)"}},
					},
				},
			},
			// ---- A lone `]` or `}` outside a class is a literal without the flag
			// but a SyntaxError with it, so no suggestion may be offered ----
			{
				Code: "/]/",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 4},
				},
			},
			{
				Code: "/}/",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 4},
				},
			},
			// ---- Under `u` the inner `[` is literal, so the trailing `]` is
			// stray and the pattern can't take the flag ----
			{
				Code: "/[[a][b]]/",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 11},
				},
			},
			// ---- The same pattern nests cleanly under `v`, where the suggestion
			// stays available ----
			{
				Code:            "/[[a][b]]/",
				Options:         []any{map[string]any{"requireFlag": "v"}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireVFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 11,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addVFlag", Output: "/[[a][b]]/v"}},
					},
				},
			},
			// ---- Braces closing a quantifier are structural, not stray ----
			{
				Code: "/a{1,2}/",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: "/a{1,2}/u"}},
					},
				},
			},
			// ---- Dimension 4: parenthesized flags argument — ESTree has no
			// parenthesis node, so the fix rewrites the literal inside ----
			{
				Code: "new RegExp('foo', ('gi'))",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 26,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: "new RegExp('foo', ('giu'))"}},
					},
				},
			},
			// ---- Real-user: a Unicode property escape only becomes meaningful
			// under the `u` flag (Annex B reads `\p{L}` as literal `p{L}` without
			// it) — a common pattern for validating letters/digits ----
			{
				Code: "const isLetter = /^\\p{L}+$/;",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 18, EndLine: 1, EndColumn: 28,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: "const isLetter = /^\\p{L}+$/u;"}},
					},
				},
			},
		},
	)
}

// TestRequireUnicodeRegexpEditDemand exercises Dimension 3 (autofix
// boundaries) for a suggestion-only rule: diagnostic identity (count,
// message, range) must stay identical across every edit demand, and the
// suggestion must materialize only when the suggestion or all-edits demand is
// requested — this rule never produces an autofix.
func TestRequireUnicodeRegexpEditDemand(t *testing.T) {
	t.Parallel()

	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(
		"var a = /foo/;\n",
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
			HasTypeInfo:  true,
			ExcludePaths: []string{},
			GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
				return []linter.ConfiguredRule{{
					Name:     RequireUnicodeRegexpRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return RequireUnicodeRegexpRule.Run(ctx, nil)
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
	for _, diagnostics := range [][]rule.RuleDiagnostic{diagnosticsOnly, autofixOnly, suggestionOnly} {
		if got, want := withoutEdits(diagnostics[0]), withoutEdits(allEdits[0]); !reflect.DeepEqual(got, want) {
			t.Errorf("changed diagnostic identity:\ngot  %#v\nwant %#v", got, want)
		}
	}

	if diagnosticsOnly[0].FixesPtr != nil || autofixOnly[0].FixesPtr != nil || suggestionOnly[0].FixesPtr != nil || allEdits[0].FixesPtr != nil {
		t.Fatalf("require-unicode-regexp never emits autofixes, but a fix was materialized")
	}

	if diagnosticsOnly[0].Suggestions != nil || autofixOnly[0].Suggestions != nil {
		t.Fatalf("non-suggestion demand unexpectedly materialized suggestions")
	}
	if suggestionOnly[0].Suggestions == nil || allEdits[0].Suggestions == nil {
		t.Fatalf("suggestion/all demand did not materialize suggestions")
	}
	if !reflect.DeepEqual(*suggestionOnly[0].Suggestions, *allEdits[0].Suggestions) {
		t.Fatalf("suggestion artifacts differ between suggestion-only and all demand")
	}
	if len(*suggestionOnly[0].Suggestions) != 1 {
		t.Fatalf("suggestions = %#v, want 1", *suggestionOnly[0].Suggestions)
	}
}

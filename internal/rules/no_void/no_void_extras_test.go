// TestNoVoidExtras locks in branches and edge shapes that the upstream test
// suite doesn't exercise. Each case carries an inline comment pointing at the
// specific branch / Dimension 4 row / tsgo AST quirk it covers, so future
// refactors can't silently regress them without breaking a named lock-in.
//
// Dimension 4 walk (rows that don't apply to no-void, with reasons):
//   - N/A access / key forms (identifier, string, numeric, private, computed,
//     element access): the rule inspects a UnaryExpression's own parent, not a
//     property key.
//   - N/A optional chain (X?.y, X?.()): void has no member-access receiver.
//   - N/A declaration/container forms (class/function declaration vs
//     expression, async/generator variants): the rule fires the same way in
//     every enclosing container; no listener targets a declaration.
//   - N/A autofix boundaries: the rule has neither an autofix nor a
//     suggestion.
//   - N/A graceful degradation around SpreadAssignment / RestElement / empty
//     bodies / overload signatures: the rule never inspects object, array, or
//     class-member shapes.
package no_void

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

var allowAsStatement = []any{map[string]any{"allowAsStatement": true}}

func TestNoVoidExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoVoidRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: parenthesized wrapper transparent to the statement check ----
			// ESTree has no ParenthesizedExpression node, so `(void 0);` sees
			// ExpressionStatement as the UnaryExpression's direct parent.
			// tsgo interposes a ParenthesizedExpression node here instead; the
			// rule must walk past it to match upstream.
			{Code: `(void 0);`, Options: allowAsStatement},
			{Code: `((void 0));`, Options: allowAsStatement},

			// ---- Dimension 2: scoping & nesting — void statement stays allowed across every enclosing container ----
			{Code: `if (a) void 0; else void 1;`, Options: allowAsStatement},
			{Code: `while (a) void 0;`, Options: allowAsStatement},
			{Code: `do void 0; while (a);`, Options: allowAsStatement},
			{Code: `for (let i = 0; i < 1; i++) void 0;`, Options: allowAsStatement},
			{Code: `switch (a) { case 1: void 0; }`, Options: allowAsStatement},
			{Code: `label: void 0;`, Options: allowAsStatement},
			{Code: `{ void 0; }`, Options: allowAsStatement},
			{Code: `class Foo { m() { void 0; } }`, Options: allowAsStatement},
			{Code: `const f = async () => { void 0; };`, Options: allowAsStatement},

			// ---- Dimension 4: void 0! is still a bare statement — the `!` binds to the argument, not the void expression ----
			{Code: `void 0!;`, Options: allowAsStatement, FileName: "file.ts"},

			// ---- Real-user: eslint/eslint#19876 — void used to explicitly discard a floating promise ----
			{Code: `void fetchData();`, Options: allowAsStatement},
			{Code: `elements.forEach(el => { void process(el); });`, Options: allowAsStatement},

			// ---- Real-user: eslint/eslint#12688 — void async IIFE used as a top-level statement ----
			{Code: `void (async () => { await doStuff(); })();`, Options: allowAsStatement},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: nesting where only the outer void counts as the statement ----
			// Locks in upstream UnaryExpression listener: allowAsStatement is
			// irrelevant once the void expression is nested inside another
			// expression, even another void expression. The outer void's
			// direct parent is the ExpressionStatement (allowed); the inner
			// void's direct parent is the outer UnaryExpression (reported).
			{
				Code:    `void void 0;`,
				Options: allowAsStatement,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noVoid", Line: 1, Column: 6, EndLine: 1, EndColumn: 12},
				},
			},

			// Locks in upstream UnaryExpression listener arm: node.parent.type
			// must be exactly "ExpressionStatement" — a ConditionalExpression,
			// SequenceExpression, or ForStatement init clause does not count,
			// even though the surrounding statement is expression-only.
			{
				Code:    `a ? void 0 : void 1;`,
				Options: allowAsStatement,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noVoid", Line: 1, Column: 5, EndLine: 1, EndColumn: 11},
					{MessageId: "noVoid", Line: 1, Column: 14, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    `void 0, void 1;`,
				Options: allowAsStatement,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noVoid", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
					{MessageId: "noVoid", Line: 1, Column: 9, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:    `for (void 0; ; ) {}`,
				Options: allowAsStatement,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noVoid", Line: 1, Column: 6, EndLine: 1, EndColumn: 12},
				},
			},

			// ---- Dimension 3-adjacent: multi-line report range ----
			// The reported range spans from the `void` keyword through the end
			// of its argument even when they sit on different lines.
			{
				Code: "var foo =\n  void\n  0;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noVoid", Line: 2, Column: 3, EndLine: 3, EndColumn: 4},
				},
			},

			// ---- Real-user: eslint/eslint#12688 — void used to make an arrow function's non-value-returning body explicit ----
			// Locks in upstream UnaryExpression listener arm: an arrow
			// function's concise (expression) body is not an ExpressionStatement,
			// so allowAsStatement does not cover it even though the intent is
			// statement-like.
			{
				Code:    `const log = x => void console.log(x);`,
				Options: allowAsStatement,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noVoid", Line: 1, Column: 18, EndLine: 1, EndColumn: 37},
				},
			},

			// Locks in upstream UnaryExpression listener arm: await unwraps to
			// an AwaitExpression parent, not ExpressionStatement.
			{
				Code:    `async function f() { await void 0; }`,
				Options: allowAsStatement,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noVoid", Line: 1, Column: 28, EndLine: 1, EndColumn: 34},
				},
			},

			// ---- Dimension 4: TS-only wrappers around the void expression itself keep reporting ----
			// A NonNullExpression, AsExpression, or SatisfiesExpression that
			// wraps the void expression (through parentheses or, for `as`/
			// `satisfies`, directly by TS precedence) is a genuine expression
			// use, not a bare statement; ESLint has no equivalent syntax to
			// compare against, so this locks in rslint's own behavior.
			{
				Code:     `(void 0)!;`,
				Options:  allowAsStatement,
				FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noVoid", Line: 1, Column: 2, EndLine: 1, EndColumn: 8},
				},
			},
			{
				Code:     `void 0 as any;`,
				Options:  allowAsStatement,
				FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noVoid", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
				},
			},
			{
				Code:     `void 0 satisfies unknown;`,
				Options:  allowAsStatement,
				FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noVoid", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
				},
			},
		},
	)
}

// TestArrowBodyStyleExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise. Each case carries an inline comment pointing at
// the specific branch / Dimension 4 row / tsgo AST quirk it covers, so future
// refactors can't silently regress them without breaking a named lock-in.
package arrow_body_style

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestArrowBodyStyleExtras(t *testing.T) {
	asNeeded := []any{"as-needed"}
	always := []any{"always"}
	never := []any{"never"}
	requireReturn := []any{"as-needed", map[string]any{"requireReturnForObjectLiteral": true}}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ArrowBodyStyleRule,
		[]rule_tester.ValidTestCase{
			// Locks in upstream validate() early-return: as-needed +
			// requireReturnForObjectLiteral keeps braces around a parenthesized
			// object return too (tsgo wraps it in a ParenthesizedExpression; we
			// SkipParentheses before the ObjectExpression check, matching ESTree).
			{Code: "var foo = () => { return ({ bar: 0 }); };", Options: requireReturn},
			// Locks in: requireReturnForObjectLiteral does not flag a non-object
			// arrow returning a parenthesized non-object.
			{Code: "var foo = () => (bar);", Options: requireReturn},
			// ---- Dimension 4: async arrow container form (valid under as-needed) ----
			{Code: "var foo = async () => bar();"},
			// ---- N/A: access/key forms — the rule inspects the arrow body, not property keys ----
			// ---- N/A: class/function declaration vs expression — the rule only matches arrow functions ----
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized-receiver body — tsgo preserves the
			// ParenthesizedExpression nodes; the report location must still be the
			// fully unwrapped inner expression (ESTree flattens parens). ----
			{
				Code:    "x => ((y))",
				Output:  []string{"x => {return ((y))}"},
				Options: always,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 8, EndColumn: 9, MessageId: "expectedBlock"}},
			},
			// ---- Dimension 4: TS non-null assertion body ----
			{
				Code:    "var foo = () => x!;",
				Output:  []string{"var foo = () => {return x!};"},
				Options: always,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 17, MessageId: "expectedBlock"}},
			},
			{
				Code:    "var foo = () => { return x!; };",
				Output:  []string{"var foo = () => x!;"},
				Options: asNeeded,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 17, MessageId: "unexpectedSingleBlock"}},
			},
			// ---- Dimension 4: TS `as` type-expression body ----
			{
				Code:    "var foo = () => x as string;",
				Output:  []string{"var foo = () => {return x as string};"},
				Options: always,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 17, MessageId: "expectedBlock"}},
			},
			{
				Code:    "var foo = () => { return x as string; };",
				Output:  []string{"var foo = () => x as string;"},
				Options: asNeeded,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 17, MessageId: "unexpectedSingleBlock"}},
			},
			// ---- Dimension 4: TS `satisfies` type-expression body ----
			{
				Code:    "var foo = () => x satisfies string;",
				Output:  []string{"var foo = () => {return x satisfies string};"},
				Options: always,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 17, MessageId: "expectedBlock"}},
			},
			// ---- Dimension 4: optional-chain body (tsgo flag, no ChainExpression wrapper) ----
			{
				Code:    "var foo = () => a?.b;",
				Output:  []string{"var foo = () => {return a?.b};"},
				Options: always,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 17, MessageId: "expectedBlock"}},
			},
			{
				Code:    "var foo = () => { return a?.b; };",
				Output:  []string{"var foo = () => a?.b;"},
				Options: asNeeded,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 17, MessageId: "unexpectedSingleBlock"}},
			},
			// ---- Dimension 4: TS return-type annotation on the arrow ----
			{
				Code:    "var f = (): number => { return 0; };",
				Output:  []string{"var f = (): number => 0;"},
				Options: asNeeded,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 23, MessageId: "unexpectedSingleBlock"}},
			},
			// ---- Dimension 4: class-field arrow container form ----
			{
				Code:    "class C { f = () => { return 0; }; }",
				Output:  []string{"class C { f = () => 0; }"},
				Options: asNeeded,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 21, MessageId: "unexpectedSingleBlock"}},
			},
			// ---- Dimension 4: async arrow container form ----
			{
				Code:    "var f = async () => { return 0; };",
				Output:  []string{"var f = async () => 0;"},
				Options: asNeeded,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 21, MessageId: "unexpectedSingleBlock"}},
			},
			{
				Code:    "var f = async () => 0;",
				Output:  []string{"var f = async () => {return 0};"},
				Options: always,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 21, MessageId: "expectedBlock"}},
			},
			// ---- Dimension 4: same-kind nesting — only the inner arrow's block
			// matches under as-needed; the listener must not bleed to the outer
			// arrow (whose body is the inner arrow expression). ----
			{
				Code:    "var f = () => () => { return 0; };",
				Output:  []string{"var f = () => () => 0;"},
				Options: asNeeded,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 21, MessageId: "unexpectedSingleBlock"}},
			},
			// ---- Dimension 4: graceful degradation — object spread in the returned literal ----
			{
				Code:    "var f = () => { return {...a}; };",
				Output:  []string{"var f = () => ({...a});"},
				Options: asNeeded,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 15, MessageId: "unexpectedObjectBlock"}},
			},
			// ---- Dimension 4: graceful degradation — destructuring parameter ----
			{
				Code:    "var f = ({a}) => { return a; };",
				Output:  []string{"var f = ({a}) => a;"},
				Options: asNeeded,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 18, MessageId: "unexpectedSingleBlock"}},
			},
			// ---- Real-user: React useCallback body collapse (eslint/eslint arrow-body-style usage) ----
			{
				Code:    "useCallback(() => { return x; }, [])",
				Output:  []string{"useCallback(() => x, [])"},
				Options: asNeeded,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 19, MessageId: "unexpectedSingleBlock"}},
			},
			// ---- Real-user: arrow returning a parenthesized object that is the
			// callee of a member call — must keep exactly one paren layer. ----
			{
				Code:    "const f = () => { return ({ a: 1 }).a; };",
				Output:  []string{"const f = () => ({ a: 1 }).a;"},
				Options: asNeeded,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 17, MessageId: "unexpectedSingleBlock"}},
			},
			// Locks in upstream validate() arm: argument === null under `never`
			// reports unexpectedSingleBlock and emits no fix (no return value).
			{
				Code:    "var foo = () => { return };",
				Options: never,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 17, MessageId: "unexpectedSingleBlock"}},
			},
			// Locks in upstream fix() arm: argument already parenthesized — the
			// sequence wrap must NOT add a second paren layer.
			{
				Code:    "var foo = () => { return (a, b); };",
				Output:  []string{"var foo = () => (a, b);"},
				Options: asNeeded,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 17, MessageId: "unexpectedSingleBlock"}},
			},
			// Locks in upstream hasASIProblem() arm: `+` punctuator after the
			// block suppresses the fix (output unchanged).
			{
				Code:    "var foo = () => { return bar }\n+x",
				Options: never,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 17, MessageId: "unexpectedSingleBlock"}},
			},
			// Locks in upstream hasASIProblem() arm: `-` punctuator after the block.
			{
				Code:    "var foo = () => { return bar }\n-x",
				Options: never,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 17, MessageId: "unexpectedSingleBlock"}},
			},
			// ---- Dimension 4: 3-level arrow nesting — only the innermost block
			// matches; the listener must not bleed across the two outer arrows. ----
			{
				Code:    "var f = () => () => () => { return 0; };",
				Output:  []string{"var f = () => () => () => 0;"},
				Options: asNeeded,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 27, MessageId: "unexpectedSingleBlock"}},
			},
			// Locks in upstream BinaryExpression[operator='in'] funcInfo propagation
			// across a nested-arrow boundary: an `in` buried in a nested arrow still
			// flags the outer for-init arrow, so its body is parenthesized on collapse.
			{
				Code:    "for (var f = () => { return g(() => a in b) } ;;);",
				Output:  []string{"for (var f = () => (g(() => a in b)) ;;);"},
				Options: asNeeded,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 20, MessageId: "unexpectedSingleBlock"}},
			},
			// Locks in upstream fix() comment arm: a comment sitting AFTER the
			// `return` keyword (not adjacent to `{`) must still take the
			// token-removal path and be preserved (golden output from ESLint).
			{
				Code:    "var f = () => { return /* c */ x; };",
				Output:  []string{"var f = () =>   /* c */ x ;"},
				Options: asNeeded,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 15, MessageId: "unexpectedSingleBlock"}},
			},
			// Comment between the value and the inner `;` lives inside the kept
			// value span (non-comment branch); it is preserved verbatim.
			{
				Code:    "var f = () => { return x /* t */; };",
				Output:  []string{"var f = () => x /* t */;"},
				Options: asNeeded,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 15, MessageId: "unexpectedSingleBlock"}},
			},
			// Comment in the (lastValueToken, closingBrace) gap triggers the
			// comment arm via the second range.
			{
				Code:    "var f = () => { return x; /* t */ };",
				Output:  []string{"var f = () =>   x /* t */ ;"},
				Options: asNeeded,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 15, MessageId: "unexpectedSingleBlock"}},
			},
			// ---- tsgo↔ESTree: destructuring-assignment target ----
			// `({b} = a)` — tsgo parses the LHS as an ObjectLiteralExpression, but
			// ESTree models it as an ObjectPattern, so the forced parens must be
			// kept (else branch), not unwrapped as a parenthesized object literal.
			{
				Code:    "var f = () => ({ b } = a);",
				Output:  []string{"var f = () => {return ({ b } = a)};"},
				Options: always,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 16, MessageId: "expectedBlock"}},
			},
			// Assignment target followed by a member access — still a pattern, parens kept.
			{
				Code:    "var f = () => ({a} = b).c;",
				Output:  []string{"var f = () => {return ({a} = b).c};"},
				Options: always,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 15, MessageId: "expectedBlock"}},
			},
			// `return {a} = b`: value starts with `{` (unexpectedObjectBlock), and the
			// assignment is parenthesized on collapse.
			{
				Code:    "var f = () => { return {a} = b; };",
				Output:  []string{"var f = () => ({a} = b);"},
				Options: asNeeded,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 15, MessageId: "unexpectedObjectBlock"}},
			},
			// Array-destructuring assignment: value starts with `[`, not `{`
			// (unexpectedSingleBlock), and needs no extra parens.
			{
				Code:    "var f = () => { return [a] = b; };",
				Output:  []string{"var f = () => [a] = b;"},
				Options: asNeeded,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 15, MessageId: "unexpectedSingleBlock"}},
			},
			// ---- tsgo↔ESTree divergence (documented): `/` after an arrow block `}`
			// is division under tsgo (regex under espree). De-bracing would yield
			// `() => x\n/y/.test(z)` = `() => x / y / .test(z)`, a syntax error — so
			// the fix is suppressed (no Output), unlike ESLint which de-braces. ----
			{
				Code:    "var f = () => { return x }\n/y/.test(z)",
				Options: never,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 15, MessageId: "unexpectedSingleBlock"}},
			},
			// ---- Dimension 4: outer arrow block returns an inner arrow (no bleed) ----
			{
				Code:    "var f = () => { return () => ({ a: 1 }); };",
				Output:  []string{"var f = () => () => ({ a: 1 });"},
				Options: asNeeded,
				Errors:  []rule_tester.InvalidTestCaseError{{Line: 1, Column: 15, MessageId: "unexpectedSingleBlock"}},
			},
		},
	)
}

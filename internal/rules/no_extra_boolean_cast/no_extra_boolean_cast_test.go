package no_extra_boolean_cast

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoExtraBooleanCastRule(t *testing.T) {
	enforceLogicalOperands := map[string]any{"enforceForLogicalOperands": true}
	enforceInnerExpressions := map[string]any{"enforceForInnerExpressions": true}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoExtraBooleanCastRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// !! not in boolean context
			{Code: `var foo = !!bar;`},
			{Code: `function foo() { return !!bar; }`},
			// Boolean() not in boolean context
			{Code: `var foo = Boolean(bar);`},
			// No extra cast
			{Code: `if (foo) {}`},
			// !! in for initializer (not condition)
			{Code: `for(!!foo;;) {}`},
			// new Boolean() outside boolean context is fine
			{Code: `var x = new Boolean(foo);`},
			// new Boolean() is never flagged — it produces a truthy
			// object, so it is not equivalent to a plain value in a
			// boolean context (matches ESLint).
			{Code: `if (new Boolean(foo)) {}`},
			{Code: `while (new Boolean(foo)) {}`},
			{Code: `!new Boolean(foo)`},
			// Boolean identifier alone is not a call
			{Code: `var x = Boolean;`},
			// Single negation is not redundant
			{Code: `if (!foo) {}`},
			// Non-Boolean named function should not trigger
			{Code: `if (Foo(bar)) {}`},
			// Not flagged when enforceForLogicalOperands is off
			{Code: `if (x || !!y) {}`},
			{Code: `if (x && Boolean(y)) {}`},
			// Not flagged when enforceForInnerExpressions is off
			{Code: `if (x ? !!y : z) {}`},
			{Code: `if (x ?? !!y) {}`},
			{Code: `if ((x, !!y, z)) {}`}, // not the last expression, Boolean() form
			// enforceForLogicalOperands: ?? is not included (only with enforceForInnerExpressions)
			{Code: `if (x ?? !!y) {}`, Options: []any{enforceLogicalOperands}},
			// enforceForInnerExpressions: ?? but !! is on the LEFT (not reached)
			{Code: `if (!!x ?? y) {}`, Options: []any{enforceInnerExpressions}},
			// enforceForInnerExpressions: not the last expression in a sequence
			{Code: `if ((Boolean(a), b)) {}`, Options: []any{enforceInnerExpressions}},
			// Inner parentheses wrapping !! or Boolean() outside boolean context
			{Code: `var x = (!!foo);`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// !! in if test
			{
				Code:   `if (!!foo) {}`,
				Output: []string{`if (foo) {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegation", Line: 1, Column: 5},
				},
			},
			// !! in while test
			{
				Code:   `while (!!foo) {}`,
				Output: []string{`while (foo) {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegation", Line: 1, Column: 8},
				},
			},
			// !! in do-while test
			{
				Code:   `do {} while (!!foo)`,
				Output: []string{`do {} while (foo)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegation", Line: 1, Column: 14},
				},
			},
			// !! in for condition
			{
				Code:   `for (;!!foo;) {}`,
				Output: []string{`for (;foo;) {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegation", Line: 1, Column: 7},
				},
			},
			// Boolean() in if test
			{
				Code:   `if (Boolean(foo)) {}`,
				Output: []string{`if (foo) {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 5},
				},
			},
			// Boolean() in while test
			{
				Code:   `while (Boolean(foo)) {}`,
				Output: []string{`while (foo) {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 8},
				},
			},
			// Boolean() as operand of !
			{
				Code:   `!Boolean(foo)`,
				Output: []string{`!foo`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 2},
				},
			},
			// !! in ternary condition
			{
				Code:   `!!foo ? bar : baz`,
				Output: []string{`foo ? bar : baz`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegation", Line: 1, Column: 1},
				},
			},
			// !!! - inner !! is operand of outer !
			{
				Code:   `!!!foo`,
				Output: []string{`!foo`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegation", Line: 1, Column: 2},
				},
			},
			// !! nested inside Boolean() call
			{
				Code:   `Boolean(!!foo)`,
				Output: []string{`Boolean(foo)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegation", Line: 1, Column: 9},
				},
			},
			// Boolean() nested inside new Boolean() — the argument
			// position of new Boolean() is treated as a boolean context
			// (for recursion), so the inner Boolean() is flagged.
			{
				Code:   `new Boolean(Boolean(foo))`,
				Output: []string{`new Boolean(foo)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 13},
				},
			},
			// !! nested inside new Boolean() — inner !! is flagged, outer
			// new Boolean() is not.
			{
				Code:   `new Boolean(!!foo)`,
				Output: []string{`new Boolean(foo)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegation", Line: 1, Column: 13},
				},
			},
			// Parenthesized !! in boolean context — wrapping parens stay.
			{
				Code:   `if ((!!foo)) {}`,
				Output: []string{`if ((foo)) {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegation", Line: 1, Column: 6},
				},
			},
			{
				Code:   `if ((Boolean(foo))) {}`,
				Output: []string{`if ((foo)) {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 6},
				},
			},
			// Double parens
			{
				Code:   `if (((!!foo))) {}`,
				Output: []string{`if (((foo))) {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegation", Line: 1, Column: 7},
				},
			},
			// Boolean() with no arguments in boolean context (via `!`).
			{
				Code:   `!Boolean()`,
				Output: []string{`true`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 2},
				},
			},
			// Boolean() with no args as argument of Boolean() — collapses
			// to `false` (the inner call is replaced, the outer then goes
			// away in a second pass because Boolean(false) is still flagged
			// until fixed).
			{
				Code:   `if (Boolean()) {}`,
				Output: []string{`if (false) {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 5},
				},
			},
			// Boolean(expr) with spread: fix is skipped, still reported.
			{
				Code:   `if (Boolean(...args)) {}`,
				Output: nil,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 5},
				},
			},
			// Boolean(a, b): two arguments, fix is skipped.
			{
				Code:   `if (Boolean(a, b)) {}`,
				Output: nil,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 5},
				},
			},
			// Comments inside — skip fix.
			{
				Code:   `if (!!/* keep */foo) {}`,
				Output: nil,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegation", Line: 1, Column: 5},
				},
			},
			{
				Code:   `if (Boolean(/* keep */foo)) {}`,
				Output: nil,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 5},
				},
			},

			// enforceForLogicalOperands: !! inside ||
			{
				Code:    `if (x || !!y) {}`,
				Output:  []string{`if (x || y) {}`},
				Options: []any{enforceLogicalOperands},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegation", Line: 1, Column: 10},
				},
			},
			// enforceForLogicalOperands: Boolean() inside &&
			{
				Code:    `while (x && Boolean(y)) {}`,
				Output:  []string{`while (x && y) {}`},
				Options: []any{enforceLogicalOperands},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 13},
				},
			},
			// enforceForLogicalOperands: deeply nested ||
			{
				Code:    `if (a || (b && !!c)) {}`,
				Output:  []string{`if (a || (b && c)) {}`},
				Options: []any{enforceLogicalOperands},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegation", Line: 1, Column: 16},
				},
			},
			// enforceForLogicalOperands: Boolean() as left of ||
			{
				Code:    `if (Boolean(a) || b) {}`,
				Output:  []string{`if (a || b) {}`},
				Options: []any{enforceLogicalOperands},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 5},
				},
			},

			// enforceForInnerExpressions: !! inside ?? right operand
			{
				Code:    `if (x ?? !!y) {}`,
				Output:  []string{`if (x ?? y) {}`},
				Options: []any{enforceInnerExpressions},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegation", Line: 1, Column: 10},
				},
			},
			// enforceForInnerExpressions: Boolean() inside ternary consequent
			{
				Code:    `if (cond ? Boolean(a) : b) {}`,
				Output:  []string{`if (cond ? a : b) {}`},
				Options: []any{enforceInnerExpressions},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 12},
				},
			},
			// enforceForInnerExpressions: !! inside ternary alternate
			{
				Code:    `if (cond ? a : !!b) {}`,
				Output:  []string{`if (cond ? a : b) {}`},
				Options: []any{enforceInnerExpressions},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegation", Line: 1, Column: 16},
				},
			},
			// enforceForInnerExpressions: last expression of a sequence
			{
				Code:    `if ((a, b, Boolean(c))) {}`,
				Output:  []string{`if ((a, b, c)) {}`},
				Options: []any{enforceInnerExpressions},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 12},
				},
			},
			// enforceForInnerExpressions also handles ||/&&
			{
				Code:    `if (x || !!y) {}`,
				Output:  []string{`if (x || y) {}`},
				Options: []any{enforceInnerExpressions},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegation", Line: 1, Column: 10},
				},
			},
			// enforceForInnerExpressions: composed logical + ternary chain
			{
				Code:    `if (a ? (b || !!c) : d) {}`,
				Output:  []string{`if (a ? (b || c) : d) {}`},
				Options: []any{enforceInnerExpressions},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNegation", Line: 1, Column: 15},
				},
			},
			// Mixed logical + coalesce: replacement keeps its own parens
			// to avoid a syntax error (`a && b ?? c` is not valid).
			{
				Code:    `if (a && Boolean(b ?? c)) {}`,
				Output:  []string{`if (a && (b ?? c)) {}`},
				Options: []any{enforceLogicalOperands},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 10},
				},
			},
		},
	)
}

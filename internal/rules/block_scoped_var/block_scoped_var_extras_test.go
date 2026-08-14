// Edge-shape augmentation (Dimensions 1-4) and upstream-branch lock-ins for
// block-scoped-var, beyond what tests/lib/rules/block-scoped-var.js covers.
// See block_scoped_var_upstream_test.go for the migrated upstream suite.
package block_scoped_var

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestBlockScopedVarExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&BlockScopedVarRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 1: TS-specific wrappers on an in-scope read ----
			{Code: "function f() { if (true) { var a = 1; (a).toString(); } }"},
			{Code: "function f() { if (true) { var a = 1; a!.toString(); } }"},
			{Code: "function f() { if (true) { var a: any = 1; (a as string).length; } }"},
			{Code: "function f() { if (true) { var a = 1; a?.toString(); } }"},

			// ---- Dimension 2: 3+ levels of nesting, only the boundary that matters flags ----
			{Code: "function f() { if (true) { if (true) { if (true) { var a = 1; a; } } } }"},
			{Code: "function f() { var a = 1; if (true) { if (true) { if (true) { a; } } } }"},

			// ---- Dimension 4: graceful degradation on binding-pattern shapes ----
			{Code: "function f() { if (true) { var { a, ...rest } = { a: 1 }; a; rest; } }"},
			{Code: "function f() { if (true) { var [a, ...rest] = [1]; a; rest; } }"},
			{Code: "function f() { if (true) { var {} = {}; } }"},
			{Code: "function f() { if (true) { var [] = []; } }"},

			// ---- Dimension 4: TS type annotation on an in-scope var ----
			{Code: "function f() { if (true) { var a: number = 1; a; } }"},

			// ---- Dimension 4: ambient `declare var` does not crash or false-positive ----
			{Code: "declare var a: number; a;"},
			{Code: "function f() { declare var a: number; a; }"},

			// ---- Real-user: catch-clause param shadows an unrelated same-named outer var ----
			{Code: "function f() { var e = 1; try {} catch (e) { e; } e; }"},

			// ---- Locks in: a bare `var` declarator (no initializer) never
			// ---- creates a self-reference, so two sibling bare declarations
			// ---- of the same name never cross-flag each other. ----
			{Code: "if (true) { var a; } else { var a; }"},
			{Code: "function f() { for (var a;;) {} for (var a;;) {} }"},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: TS wrappers around an out-of-scope read still flag the identifier ----
			outOfScope("function f() { if (true) { var a = 1; } (a).toString(); }",
				scopeErr("a", 1, 32, 1, 42)),
			outOfScope("function f() { if (true) { var a = 1; } ((a)).toString(); }",
				scopeErr("a", 1, 32, 1, 43)),
			outOfScope("function f() { if (true) { var a = 1; } a!.toString(); }",
				scopeErr("a", 1, 32, 1, 41)),
			outOfScope("function f() { if (true) { var a: any = 1; } (a as string).length; }",
				scopeErr("a", 1, 32, 1, 47)),
			outOfScope("function f() { if (true) { var a = 1; } a?.toString(); }",
				scopeErr("a", 1, 32, 1, 41)),
			outOfScope("function f() { if (true) { var a = function(){}; } a?.(); }",
				scopeErr("a", 1, 32, 1, 52)),

			// ---- Dimension 4: computed key and template-literal reads of an out-of-scope var ----
			outOfScope("function f() { if (true) { var a = 1; } var o = { [a]: 1 }; }",
				scopeErr("a", 1, 32, 1, 52)),
			outOfScope("function f() { if (true) { var a = 1; } `${a}`; }",
				scopeErr("a", 1, 32, 1, 44)),

			// ---- Dimension 4: destructured rest binding used outside its block ----
			outOfScope("function f() { if (true) { var { a, ...rest } = { a: 1 }; } rest; }",
				scopeErr("rest", 1, 40, 1, 61)),
			outOfScope("function f() { if (true) { var [a, ...rest] = [1]; } rest; }",
				scopeErr("rest", 1, 39, 1, 54)),

			// ---- Dimension 2: 3-level nesting, innermost declaration leaks past its own block only ----
			outOfScope("function f() { if (true) { if (true) { if (true) { var a = 1; } a; } } }",
				scopeErr("a", 1, 56, 1, 65)),

			// ---- Real-user: `var` assigned only inside a conditional branch, read after the block ----
			outOfScope("function connect(cb) { if (cb) { var onDone = function() { cb(); }; } return onDone; }",
				scopeErr("onDone", 1, 38, 1, 78)),

			// ---- Real-user: try/catch fallback assignment read after both blocks (classic gotcha) ----
			// Locks in that sibling `var` declarations of the same name cross-flag
			// each other's own identifier, not just later reads (see the if/else
			// and duplicate-for-loop cases in the upstream suite).
			outOfScope("function parse(input) { try { var result = input.a; } catch (e) { var result = null; } return result; }",
				scopeErr("result", 1, 71, 1, 35),
				scopeErr("result", 1, 35, 1, 71),
				scopeErr("result", 1, 35, 1, 95),
				scopeErr("result", 1, 71, 1, 95)),

			// ---- Locks in: an initialized `var` declarator's self-reference
			// ---- only fires against a sibling declarator that itself has an
			// ---- initializer (or is a for-in/for-of binding) — a bare
			// ---- sibling declarator is never flagged and never flags back.
			outOfScope("function f() { try { var a = compute(); } catch (e) { var a; } return a; }",
				scopeErr("a", 1, 59, 1, 26),
				scopeErr("a", 1, 26, 1, 71),
				scopeErr("a", 1, 59, 1, 71)),
			outOfScope("function f() { for (var a in {}) {} for (var a in {}) {} }",
				scopeErr("a", 1, 46, 1, 25),
				scopeErr("a", 1, 25, 1, 46)),
		},
	)
}

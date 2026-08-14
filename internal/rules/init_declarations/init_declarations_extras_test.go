package init_declarations

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestInitDeclarationsExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension row / real-user scenario it
// covers, so future refactors can't silently regress them without breaking a
// named lock-in.
func TestInitDeclarationsExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&InitDeclarationsRule,
		[]rule_tester.ValidTestCase{
			// ---- Locks in upstream: `id.type === "Identifier"` gate ----
			// Destructuring declarators are silently skipped regardless of
			// mode/kind; none of these upstream-migrated cases exercise a
			// destructuring pattern at all.
			{Code: `var {a} = obj;`, Options: []any{"never"}},
			{Code: `let [a] = arr;`, Options: []any{"never"}},
			{Code: `var {} = obj;`, Options: []any{"never"}}, // N/A: Dimension 4 graceful degradation — empty binding pattern
			{Code: `for (let [a] of items) {}`, Options: []any{"never"}},
			{Code: `for (const {a} of items) {}`},

			// ---- Locks in upstream: `node.declare` on a non-const kind ----
			// Only "declare const" is migrated from upstream; a bare
			// `declare` on let/var must exempt them too, even though they
			// aren't otherwise exempt via CONSTANT_BINDINGS.
			{Code: `declare let x: number;`, Options: []any{"always"}},
			{Code: `declare var y: number = 1;`, Options: []any{"never"}},

			// ---- Dimension 4: nesting/traversal boundary — `declare global` ----
			{Code: `declare global { var x: number; }`, Options: []any{"always"}},

			// ---- Real-user: typescript-eslint#4392 eg.2 ----
			// A `declare const` sibling inside a declared namespace must not
			// disturb ambient status for a later, non-declared nested
			// namespace. The pre-ESLint-core-dialect typescript-eslint wrapper
			// tracked ambient nesting with a boolean toggled by
			// TSModuleDeclaration enter/exit; this port instead derives
			// ambient status from the binder's NodeFlagsAmbient
			// (utils.IsInAmbientContext), which propagates structurally and
			// isn't sensitive to sibling declare statements.
			{
				Code: `
declare namespace obj1 {
  declare const a: number;
  namespace obj1_1 {
    const b: string;
  }
}
`,
				Options: []any{"always"},
			},

			// ---- Dimension 4: nesting boundary — redundant nested "declare namespace" ----
			// A `declare namespace` nested inside another `declare namespace`
			// must not prematurely end ambient status for its later siblings.
			{
				Code: `
declare namespace A {
  declare namespace B {
    let x: number;
  }
  let y: number;
}
`,
				Options: []any{"always"},
			},

			// ---- Dimension 2: scoping — options JSON-shape coverage for ignoreForLoopInit ----
			{Code: `for (var foo in []) {}`, Options: []any{"never", map[string]any{"ignoreForLoopInit": true}}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Locks in parseOptions' `mode` default fallback (upstream `meta.defaultOptions: ["always"]`) ----
			{
				Code:   `var foo;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "initialized", Line: 1, Column: 5, EndLine: 1, EndColumn: 8}},
			},

			// ---- Dimension 4: options JSON-shape coverage — ignoreForLoopInit: false still reports ----
			{
				Code:    `for (var foo in []) {}`,
				Options: []any{"never", map[string]any{"ignoreForLoopInit": false}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "notInitialized", Line: 1, Column: 10, EndLine: 1, EndColumn: 13}},
			},

			// ---- Dimension 2: scoping — arrow function body ----
			{
				Code:    `const f = () => { var a; };`,
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "initialized", Line: 1, Column: 23, EndLine: 1, EndColumn: 24}},
			},

			// ---- Dimension 2: scoping — getter body ----
			{
				Code:    `class C { get x() { var a; } }`,
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "initialized", Line: 1, Column: 25, EndLine: 1, EndColumn: 26}},
			},

			// ---- Dimension 2: scoping — class static block (explicit Dimension-4 checklist row) ----
			{
				Code:    `class C { static { var a; } }`,
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "initialized", Line: 1, Column: 24, EndLine: 1, EndColumn: 25}},
			},

			// ---- Diagnostic contract: exact message text per messageId ----
			{
				Code:    `var count;`,
				Options: []any{"always"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "initialized",
					Message:   "Variable 'count' should be initialized on declaration.",
					Line:      1, Column: 5, EndLine: 1, EndColumn: 10,
				}},
			},
			{
				Code:    `var count = 1;`,
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "notInitialized",
					Message:   "Variable 'count' should not be initialized on declaration.",
					Line:      1, Column: 5, EndLine: 1, EndColumn: 14,
				}},
			},
		},
	)
}

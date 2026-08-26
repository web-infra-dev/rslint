// TestNoInnerDeclarationsExtras locks in rslint-specific AST shapes, option
// boundaries, and real-world regressions that are not present in the upstream
// ESLint v10.8.1 RuleTester suite. The 1:1 migration lives in the sibling
// no_inner_declarations_upstream_test.go file.
package no_inner_declarations

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoInnerDeclarationsExtras(t *testing.T) {
	// N/A: the rule has no fixer or suggestions, so edit-range and suggestion
	// contracts do not apply.
	// N/A: receiver and expression wrappers do not apply because the rule only
	// visits declarations.
	// N/A: access and key forms do not apply because the rule never compares or
	// classifies names, member access, or property keys.
	//
	// Applicable syntax dimensions below cover declaration/container forms,
	// same-kind nesting, comments, destructuring/rest, empty bodies, and
	// body-less TypeScript declarations.
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoInnerDeclarationsRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: declaration-list placement ----
			// Locks in upstream create() listener arm 1: the default "functions"
			// mode does not register a VariableDeclaration listener, including for
			// a bare loop initializer in tsgo.
			noInnerValid(`if (test) { var value; } for (var i = 0; i < 1; i++) {}`, nil, 2015),
			noInnerValid(`if (test) { var value; } for (var i = 0; i < 1; i++) {}`, []any{"functions", map[string]any{}}, 2015),

			// Locks in upstream VariableDeclaration() arm 1: only var participates
			// in "both"; all other declaration keywords return without checking.
			noInnerValid(`for (let i = 0; i < 1; i++) {}`, []any{"both"}, 2015),
			noInnerValid(`for (const key in object) {}`, []any{"both"}, 2015),
			noInnerValid(`for (const value of values) {}`, []any{"both"}, 2015),
			noInnerValid(`if (test) { using resource = acquire(); }`, []any{"both"}, 2026),
			noInnerValid(`async function f() { if (test) { await using resource = acquire(); } }`, []any{"both"}, 2026),

			// ---- Dimension 4: declaration and container forms ----
			// Locks in upstream isInAllowedPosition() arm 1: direct program and
			// static-block children are roots. Locks in arm 2: function-like block
			// bodies are roots across their tsgo container forms.
			// N/A: ESTree's ExportNamedDeclaration and ExportDefaultDeclaration
			// parent arms are flattened by tsgo; the Layer 1 export cases lock in
			// the same public behavior.
			noInnerValid(`function declaration() { var x; }`, []any{"both"}, 2015),
			noInnerValid(`const expression = function () { var x; };`, []any{"both"}, 2015),
			noInnerValid(`const arrow = () => { var x; };`, []any{"both"}, 2015),
			noInnerValid(`class C { constructor() { var x; } method() { var y; } get value() { var z; return z; } set value(v) { var z; } static { var q; } }`, []any{"both"}, 2022),

			// Locks in upstream FunctionDeclaration() arm 1: strict ES2015+ scopes
			// suppress the report, including implicit class and module strictness.
			noInnerValid(`class C { method() { if (test) { function nested() {} } } }`, []any{"functions", map[string]any{"blockScopedFunctions": "allow"}}, 2015),
			{
				Code:            `export {}; function outer() { if (test) { function nested() {} } }`,
				Options:         []any{"functions", map[string]any{"blockScopedFunctions": "allow"}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015, SourceType: "module"},
			},

			// ---- Real-user: eslint/eslint#14821 TypeScript namespace behavior ----
			// @typescript-eslint/scope-manager marks namespace/module scopes as
			// strict, so the default allow policy accepts block functions there.
			noInnerValid(`namespace N { if (test) { function nested() {} } }`, []any{"functions", map[string]any{"blockScopedFunctions": "allow"}}, 2015),

			// ---- Dimension 4: graceful body-less and empty declarations ----
			// tsgo exposes these body-less signatures as FunctionDeclaration,
			// while ESTree exposes TSDeclareFunction and the upstream listener
			// never receives them.
			noInnerValid(`namespace N { function overload(value: string): string; }`, []any{"functions", map[string]any{"blockScopedFunctions": "disallow"}}, 2022),
			noInnerValid(`declare module "pkg" { function exported(): void; }`, []any{"functions", map[string]any{"blockScopedFunctions": "disallow"}}, 2022),
			noInnerValid(`function empty() {}`, []any{"both"}, 2015),

			// ---- Disable-directive boundaries ----
			// RuleTester registers the rule as "test". Lock in both line forms;
			// the diagnostic starts at the nested declaration, not at the outer
			// control-flow statement that owns the disabled line.
			noInnerValid(`// eslint-disable-next-line test
if (test) { function nested() {} }`, []any{"functions", map[string]any{"blockScopedFunctions": "disallow"}}, 2015),
			noInnerValid(`if (test) { var value; } // eslint-disable-line test`, []any{"both"}, 2015),
		},
		[]rule_tester.InvalidTestCase{
			// ---- Real-user: eslint/eslint#976 nested var loop headers ----
			// Locks in upstream VariableDeclaration() arm 2 and check() arm 1:
			// nested var declarations report against the program root. These are
			// bare VariableDeclarationList nodes in tsgo rather than statements.
			noInnerInvalid(
				`for (var i = 0; i < 1; i++) {}`,
				[]any{"both"},
				2015,
				noInnerError("Move variable declaration to program root.", 1, 6, 1, 15),
			),
			noInnerInvalid(
				`function f() {
  for (var i = 0, j = 1; i < j; i++) {}
}`,
				[]any{"both"},
				2015,
				noInnerError("Move variable declaration to function body root.", 2, 8, 2, 24),
			),
			noInnerInvalid(
				`for (var key in object) {}`,
				[]any{"both"},
				2015,
				noInnerError("Move variable declaration to program root.", 1, 6, 1, 13),
			),
			noInnerInvalid(
				`for (var value of values) {}`,
				[]any{"both"},
				2015,
				noInnerError("Move variable declaration to program root.", 1, 6, 1, 15),
			),
			noInnerInvalid(
				`async function f() { for await (var value of values) {} }`,
				[]any{"both"},
				2022,
				noInnerError("Move variable declaration to function body root.", 1, 33, 1, 42),
			),

			// ---- Dimension 4: comments, destructuring, and rest bindings ----
			// Report ranges exclude the for keyword, delimiters, and leading
			// comments, while retaining comments inside the declaration node.
			noInnerInvalid(
				`for (/* before */ var /* binding */ i = 0; i < 1; i++) {}`,
				[]any{"both"},
				2015,
				noInnerError("Move variable declaration to program root.", 1, 19, 1, 42),
			),
			noInnerInvalid(
				`for (var {a, ...rest} of values) {}`,
				[]any{"both"},
				2018,
				noInnerError("Move variable declaration to program root.", 1, 6, 1, 22),
			),

			// Locks in upstream check() arm 3: static-block diagnostics name the
			// class static block body.
			noInnerInvalid(
				`class C { static { for (var i = 0; i < 1; i++) {} } }`,
				[]any{"both"},
				2022,
				noInnerError("Move variable declaration to class static block body root.", 1, 25, 1, 34),
			),
			noInnerInvalid(
				`class C {
  static {
    if (test) {
      var foo;
    }
  }
}`,
				[]any{"both"},
				2022,
				noInnerError("Move variable declaration to class static block body root.", 4, 7, 4, 15),
			),

			// ---- Dimension 4: same-kind nesting and traversal ----
			noInnerInvalid(
				`for (var i = 0; i < 1; i++) { for (var j = 0; j < 1; j++) {} }`,
				[]any{"both"},
				2015,
				noInnerError("Move variable declaration to program root.", 1, 6, 1, 15),
				noInnerError("Move variable declaration to program root.", 1, 36, 1, 45),
			),
			noInnerInvalid(
				`if (a) { function outer() { if (b) function inner() {} } }`,
				[]any{"functions", map[string]any{"blockScopedFunctions": "disallow"}},
				2015,
				noInnerError("Move function declaration to program root.", 1, 10, 1, 57),
				noInnerError("Move function declaration to function body root.", 1, 36, 1, 55),
			),

			// Locks in upstream check() arm 2: nested declarations in every
			// function-like container name the function body.
			noInnerInvalid(
				`async function outer() { if (x) function inner() {} }`,
				[]any{"functions", map[string]any{"blockScopedFunctions": "disallow"}},
				2017,
				noInnerError("Move function declaration to function body root.", 1, 33, 1, 52),
			),
			noInnerInvalid(
				`function* outer() { if (x) function inner() {} }`,
				[]any{"functions", map[string]any{"blockScopedFunctions": "disallow"}},
				2015,
				noInnerError("Move function declaration to function body root.", 1, 28, 1, 47),
			),
			noInnerInvalid(
				`async function* outer() { if (x) function inner() {} }`,
				[]any{"functions", map[string]any{"blockScopedFunctions": "disallow"}},
				2018,
				noInnerError("Move function declaration to function body root.", 1, 34, 1, 53),
			),

			// ---- Real-user: eslint/eslint#19955 ES5 strict-mode gate ----
			// Locks in upstream FunctionDeclaration() arm 2: strictness alone is
			// insufficient. Annex B block functions are only allowed when the
			// configured ECMAScript edition is ES2015 or newer.
			noInnerInvalid(
				`"use strict"; if (test) { function nested() {} }`,
				[]any{"functions", map[string]any{"blockScopedFunctions": "allow"}},
				5,
				noInnerError("Move function declaration to program root.", 1, 27, 1, 47),
			),

			// Default and explicitly empty option objects both resolve to allow,
			// but non-strict script code must still report.
			noInnerInvalid(
				`if (test) { function nested() {} }`,
				nil,
				2015,
				noInnerError("Move function declaration to program root.", 1, 13, 1, 33),
			),
			noInnerInvalid(
				`if (test) { function nested() {} }`,
				[]any{"functions", map[string]any{}},
				2015,
				noInnerError("Move function declaration to program root.", 1, 13, 1, 33),
			),

			// ---- Real-user: eslint/eslint#14821 TypeScript namespace behavior ----
			// Namespace/module blocks are not declaration roots in the core rule.
			noInnerInvalid(
				`namespace N { if (test) { var value; } }`,
				[]any{"both"},
				2022,
				noInnerError("Move variable declaration to program root.", 1, 27, 1, 37),
			),
			noInnerInvalid(
				`namespace N { if (test) { function nested() {} } }`,
				[]any{"functions", map[string]any{"blockScopedFunctions": "disallow"}},
				2022,
				noInnerError("Move function declaration to program root.", 1, 27, 1, 47),
			),

			// A scoped disable must stop at eslint-enable, while a subsequent
			// next-line directive suppresses only that one source line.
			noInnerInvalid(
				`/* eslint-disable test */
if (hidden) { function hidden() {} }
/* eslint-enable test */
// eslint-disable-next-line test
if (lineHidden) { var hidden; }
if (shown) { var shown; }`,
				[]any{"both", map[string]any{"blockScopedFunctions": "disallow"}},
				2015,
				noInnerError("Move variable declaration to program root.", 6, 14, 6, 24),
			),
		},
	)
}

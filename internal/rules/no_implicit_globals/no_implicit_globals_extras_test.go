package no_implicit_globals

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoImplicitGlobalsExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it
// covers, so future refactors can't silently regress them without breaking a
// named lock-in.
func TestNoImplicitGlobalsExtras(t *testing.T) {
	lexical := []any{map[string]any{"lexicalBindings": true}}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoImplicitGlobalsRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: receiver wrappers, opaque TS wrappers ----

			// `as`/`satisfies` writes are not eslint-scope-recognized patterns
			// (see findPureAssignmentRoot doc comment); mirrors
			// no_global_assign's isWriteThroughTypeAssertion exclusion.
			{Code: `(foo as any) = 1;`},
			{Code: `(foo satisfies any) = 1;`},
			{Code: `(Array as any) = 1;`},
			{Code: `(Array satisfies any) = 1;`},

			// ---- Dimension 4: declaration vs expression forms ----

			// Class expressions never bind a name at the declaration site.
			{Code: `const C = class Foo {};`},
			{Code: `const C = class Array {};`},

			// ---- Dimension 4: nesting / traversal boundaries ----

			// Locks in: `let`/`const` in a for-loop head get their own
			// per-iteration scope, not the global scope.
			{Code: `for (let i = 0; i < 1; i++) {}`, Options: lexical},
			{Code: `for (const i of []) {}`, Options: lexical},

			// Locks in: double-nested block does not leak `let` to global scope.
			{Code: `{ { let foo = 1; } }`, Options: lexical},

			// ---- Dimension 4: graceful degradation ----

			// Empty destructuring patterns bind no names — nothing to report,
			// and CollectBindingNames must not panic on them.
			{Code: `const {} = {};`, Options: lexical},
			{Code: `const [] = [];`, Options: lexical},

			// ---- Dimension 4: `declare` / ambient forms have no runtime binding ----

			{Code: `declare var foo: number;`},
			{Code: `declare function foo(): void;`},
			{Code: `declare class Foo {}`, Options: lexical},
			{Code: `declare global { var foo: number; }`},
			{Code: `declare global { function foo(): void; }`},

			// ---- Branch lock-in: `using` has no ESLint analog ----
			//
			// (`await using` can't be tested at true global scope: top-level
			// await requires module context, which disables the whole
			// declaration-checking path via hasNonGlobalTopLevelScope — so
			// `using`'s shared code path is the only reachable case.)
			{Code: `using foo = getResource();`, Options: lexical},

			// ---- Real-user: readonly builtin shadowed by a for-in/for-of loop variable is a common false-positive concern ----
			{Code: `for (const Array of [[1], [2]]) { Array.length; }`, Options: lexical},

			// ---- Options contract: an empty options object fills the schema
			// default (lexicalBindings: false), same as omitting options entirely ----
			{Code: `const foo = 1;`, Options: []any{map[string]any{}}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: receiver wrappers on the leak/readonly identifier ----

			// Locks in: parenthesized assignment target is still a pure write —
			// findPureAssignmentRoot passes transparently through parens.
			{
				Code:   `(foo) = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalVariableLeak"}},
			},
			{
				Code:   `((foo)) = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalVariableLeak"}},
			},
			// Non-null assertion is transparent too (unlike `as`/`satisfies`).
			{
				Code:   `foo! = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalVariableLeak"}},
			},
			{
				Code:   `Array! = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "assignmentToReadonlyGlobal"}},
			},
			{
				Code:   `(Array) = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "assignmentToReadonlyGlobal"}},
			},

			// ---- Dimension 4: access/key forms in destructuring assignment ----

			// Numeric-literal key.
			{
				Code:   `({0: foo} = obj);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalVariableLeak"}},
			},
			// Computed key — the key expression itself is a read, not a write;
			// only the bound `foo` leaks.
			{
				Code:   `const k = "x"; ({[k]: foo} = obj);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalVariableLeak"}},
			},

			// ---- Dimension 4: nesting / traversal boundaries ----

			// Locks in: nested function declarations don't bleed to the
			// global-scope check — exactly one error, for `outer`; `inner`
			// must not also be reported.
			{
				Code:   `function outer() { function inner() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalNonLexicalBinding"}},
			},

			// `var` hoists past a bare block to the true global scope — this
			// is untested upstream (their block-scoping tests only cover
			// `const`/`let`/`class`).
			{
				Code:   `{ var foo = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalNonLexicalBinding"}},
			},
			// `var` in a for-loop head also hoists to the global scope.
			{
				Code:   `for (var i = 0; i < 1; i++) {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalNonLexicalBinding"}},
			},
			{
				Code:   `for (var x in {}) {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalNonLexicalBinding"}},
			},

			// ---- Dimension 4: graceful degradation ----

			// Rest element in an object destructuring declaration: both `a`
			// and `rest` are separate bindings from the same declarator, and
			// (matching ESLint's def.node quirk) share the same reported
			// position.
			{
				Code:    `const {a, ...rest} = obj;`,
				Options: lexical,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalLexicalBinding"},
					{MessageId: "globalLexicalBinding"},
				},
			},
			// Object rest in a destructuring *assignment* (not a declaration)
			// is still a write target for leak purposes.
			{
				Code:   `({...foo} = obj);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalVariableLeak"}},
			},

			// ---- Branch lock-in: default value inside a destructuring assignment ----

			// Locks in findPureAssignmentRoot's IsDefaultValueInDestructuringAssignment
			// branch: `foo`'s own `= 1` is a pattern default, not a real
			// assignment, so the walk must continue past it to the enclosing
			// `[foo = 1] = arr` to find the true leak root.
			{
				Code:   `[foo = 1] = arr;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalVariableLeak"}},
			},
			{
				Code:   `({x: foo = 1} = obj);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalVariableLeak"}},
			},

			// ---- Diagnostic contract: full Line/Column/EndLine/EndColumn per container, including a multi-line case ----
			//
			// Upstream's own suite mostly asserts message text only; these
			// lock in exact ranges rslint reports, one per message category,
			// each over two declarations/writes on separate lines.
			{
				Code: "var foo = 1;\nvar bar = 2;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalNonLexicalBinding", Line: 1, Column: 5, EndLine: 1, EndColumn: 12},
					{MessageId: "globalNonLexicalBinding", Line: 2, Column: 5, EndLine: 2, EndColumn: 12},
				},
			},
			{
				Code: "function foo() {\n  return 1;\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalNonLexicalBinding", Line: 1, Column: 1, EndLine: 3, EndColumn: 2},
				},
			},
			{
				Code:    "class Foo {\n  method() {}\n}",
				Options: lexical,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalLexicalBinding", Line: 1, Column: 1, EndLine: 3, EndColumn: 2},
				},
			},
			{
				Code:    "const foo = 1;\nlet bar = 2;",
				Options: lexical,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalLexicalBinding", Line: 1, Column: 7, EndLine: 1, EndColumn: 14},
					{MessageId: "globalLexicalBinding", Line: 2, Column: 5, EndLine: 2, EndColumn: 12},
				},
			},
			{
				Code: "foo = 1;\nbar = 2;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalVariableLeak", Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
					{MessageId: "globalVariableLeak", Line: 2, Column: 1, EndLine: 2, EndColumn: 8},
				},
			},
			{
				Code: "Array = 1;\nObject = 2;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "assignmentToReadonlyGlobal", Line: 1, Column: 1, EndLine: 1, EndColumn: 10},
					{MessageId: "assignmentToReadonlyGlobal", Line: 2, Column: 1, EndLine: 2, EndColumn: 11},
				},
			},
			{
				Code: "var Array = 1;\nvar Object = 2;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "redeclarationOfReadonlyGlobal", Line: 1, Column: 5, EndLine: 1, EndColumn: 14},
					{MessageId: "redeclarationOfReadonlyGlobal", Line: 2, Column: 5, EndLine: 2, EndColumn: 15},
				},
			},

			// ---- Real-user: a locally-scoped-looking accumulator forgets `var`/`let` inside a function ----
			//
			// The classic bug this rule exists to catch: a script author means
			// `total` to be local to the IIFE, forgets the declaration, and
			// silently creates a global. Wrapped in an IIFE (a function
			// expression, not a declaration) so the only diagnostic is the leak.
			{
				Code:   `(function calc(items) { total = 0; for (const item of items) { total += item; } return total; })();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "globalVariableLeak"}},
			},
		},
	)
}

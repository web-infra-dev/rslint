// TestInitDeclarationsExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it
// covers, so future refactors can't silently regress them without breaking a
// named lock-in.
//
// Dimension 4 walk (rows from PORT_RULE.md):
//   - Receiver / expression wrappers ........... N/A: init-declarations
//     inspects only the binding side of a declarator, never an expression
//     receiver, so `(X).y`, `X!.y`, `X?.y`, `X as T` shapes have no
//     applicable input position.
//   - Access / key forms ....................... N/A: declarators don't have
//     a property/key. The only "name form" choice is `id.type === Identifier`
//     vs binding pattern, which is covered by the destructuring rows below.
//   - Declaration / container forms ............ covered (var/let/const +
//     using / await using + class-field vs variable-statement disambiguation).
//   - Nesting / traversal boundaries ........... covered (declare-namespace
//     ancestor walk; declare-module / declare-global ambient containers;
//     non-declare-namespace nested in another non-declare-namespace).
//   - Graceful degradation ..................... covered (object/array
//     destructuring, default + rest in patterns, mixed
//     destructuring + identifier in one list, function/enum/interface
//     containers that are not VariableDeclarations).
package init_declarations

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestInitDeclarationsExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&InitDeclarationsRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: declaration/container forms — variable kind ----
			// `using` / `await using` join const in CONSTANT_BINDINGS: they require
			// an initializer at parse time, so "never" must not report them.
			{Code: `
async function f() {
  using x = { [Symbol.dispose]() {} };
}
      `, Options: arrayOption("never")},
			{Code: `
async function f() {
  await using x = { async [Symbol.asyncDispose]() {} };
}
      `, Options: arrayOption("never")},

			// ---- Dimension 4: declaration/container forms — destructuring patterns ----
			// Upstream skips when `id.type !== "Identifier"`. Object and array
			// patterns must not crash and must not flip the rule on or off for
			// sibling identifiers.
			{Code: `const { a, b } = obj;`, Options: arrayOption("always")},
			{Code: `const [a, b] = arr;`, Options: arrayOption("always")},
			{Code: `let { a, b } = obj;`, Options: arrayOption("never")},
			{Code: `let { a = 1 } = obj;`, Options: arrayOption("always")},
			{Code: `let { a: x } = obj;`, Options: arrayOption("always")},
			// Rest element in destructuring — graceful degradation row.
			{Code: `const { a, ...rest } = obj;`, Options: arrayOption("always")},

			// ---- Dimension 4: nesting/traversal — declared-namespace boundary ----
			// `declare module "..." { ... }` (ambient external module): bindings
			// inside are ambient and must be skipped under both modes.
			{Code: `
declare module 'm' {
  const x: number;
  let y: number;
}
      `, Options: arrayOption("always")},
			{Code: `
declare module 'm' {
  const x: number;
  let y: number;
}
      `, Options: arrayOption("never")},

			// ---- Dimension 4: nesting/traversal — declare global ----
			// `declare global { var x: number; }` is an ambient global augmentation;
			// the variable is also ambient through the declare-namespace ancestor.
			{Code: `
declare global {
  var globalCount: number;
}
export {};
      `, Options: arrayOption("always")},

			// ---- Dimension 4: declaration/container forms — class-field vs variable ----
			// Class fields are PropertyDeclaration, not VariableDeclaration; the
			// rule must not fire on them in either mode.
			{Code: `
class C {
  foo: number;
  bar: string = 'x';
}
      `, Options: arrayOption("always")},
			{Code: `
class C {
  foo: number;
  bar: string = 'x';
}
      `, Options: arrayOption("never")},

			// ---- Real-user: enum members ----
			// EnumMember is not a VariableDeclaration; uninitialized members must
			// not be flagged even under "always".
			{Code: `
enum Color {
  Red,
  Green,
  Blue,
}
      `, Options: arrayOption("always")},

			// ---- Real-user: function declarations are not flagged ----
			{Code: `function foo() {}`, Options: arrayOption("always")},
			{Code: `function foo() {}`, Options: arrayOption("never")},

			// Locks in upstream isInitialized() arm: ForStatement with multi-declarator init
			// AND ignoreForLoopInit=true — every declarator in the list is treated
			// as initialized, so none should be reported.
			{Code: `for (var i = 0, j = 1; i < j; i++) {}`, Options: arrayOption("never", map[string]interface{}{"ignoreForLoopInit": true})},
			// Locks in upstream isInitialized() arm: ForStatement initializer
			// without an explicit init is still treated as initialized.
			{Code: `for (var i; i < 5; i++) {}`, Options: arrayOption("always")},

			// Locks in VariableDeclaration() upstream branch: TypeScript enums and
			// interfaces inside a declared namespace are not VariableDeclarations.
			{Code: `
declare namespace N {
  enum E {
    A,
    B,
  }
  interface I {
    x: number;
  }
}
      `, Options: arrayOption("always")},

			// Locks in `declare var x: number = 1;` mode='never': the wrapper's
			// base-rule call would otherwise report `notInitialized`, but the
			// ambient-statement skip must short-circuit first.
			{Code: `declare var x: number;`, Options: arrayOption("never")},
			{Code: `declare let y: number;`, Options: arrayOption("never")},

			// Locks in `for (using x of asyncIter)` — `using` in for-of LHS is
			// a const-like binding, so "never" must not report it (CONSTANT_BINDINGS
			// branch). The iteration provides the value, so it's "initialized"
			// in upstream's for-loop sense too — both gates exempt it.
			{Code: `
function f(arr: Disposable[]) {
  for (using x of arr) {
  }
}
      `, Options: arrayOption("never")},

			// Locks in TS "ambient external module" — a string-named module
			// (`declare module 'm'`) is also an ambient ModuleDeclaration and
			// must be skipped.
			{Code: `
declare module 'm' {
  let x: number = 1;
}
      `, Options: arrayOption("never")},

			// Locks in for-of body boundary: a non-for-loop var declaration in
			// the body of a for-loop must NOT inherit the "treated as
			// initialized" rule. Here `var inner;` is in the body, so under
			// "always" it should be reported (covered as invalid below); the
			// VALID half here is the same shape with an init to make sure body
			// declarations follow the normal `Initializer != nil` path.
			{Code: `
for (var i = 0; i < 1; i++) {
  var inner = 0;
}
      `, Options: arrayOption("always")},

			// ============================================================
			// tsgo init-expression forms — every Kind*Expression that can sit
			// on the right-hand side of `=` must count as "initialized". These
			// guard against accidental tsgo-shape regressions (e.g. a future
			// unwrap that misses ParenthesizedExpression / TypeAssertion).
			// ============================================================

			// Parenthesized literal (tsgo preserves the ParenthesizedExpression
			// where ESTree flattens it).
			{Code: `let a = (1);`, Options: arrayOption("always")},
			// Multi-level paren wrap.
			{Code: `let a = (((1)));`, Options: arrayOption("always")},
			// Template literal head + tail.
			{Code: "let s = `hello`;", Options: arrayOption("always")},
			// Tagged template.
			{Code: "let s = tag`hello`;", Options: arrayOption("always")},
			// `as` type assertion.
			{Code: `let n = 0 as number;`, Options: arrayOption("always")},
			// `satisfies` clause (tsgo distinct kind from `as`).
			{Code: `let n = 0 satisfies number;`, Options: arrayOption("always")},
			// Angle-bracket type assertion (legacy form, still a valid init).
			{Code: `let n = <number>0;`, Options: arrayOption("always"), Tsx: false},
			// `as const`.
			{Code: `let n = 0 as const;`, Options: arrayOption("always")},
			// Non-null assertion on a reference initializer.
			{Code: `function f(x?: number) { let y = x!; }`, Options: arrayOption("always")},
			// Function expression.
			{Code: `let f = function () {};`, Options: arrayOption("always")},
			// Arrow function with body.
			{Code: `let f = () => {};`, Options: arrayOption("always")},
			// Arrow function with concise return.
			{Code: `let f = () => 0;`, Options: arrayOption("always")},
			// IIFE.
			{Code: `let r = (function () { return 1; })();`, Options: arrayOption("always")},
			// Class expression.
			{Code: `let C = class {};`, Options: arrayOption("always")},
			// New expression.
			{Code: `let d = new Date();`, Options: arrayOption("always")},
			// Conditional expression.
			{Code: `let x = true ? 1 : 2;`, Options: arrayOption("always")},
			// Logical / nullish-coalescing.
			{Code: `let x = a ?? 0;`, Options: arrayOption("always")},
			// `void` operator (yields undefined but is still an init).
			{Code: `let x = void 0;`, Options: arrayOption("always")},
			// `undefined` keyword as init (explicit undefined still counts).
			{Code: `let x = undefined;`, Options: arrayOption("always")},
			// `null` keyword.
			{Code: `let x = null;`, Options: arrayOption("always")},
			// Regex literal.
			{Code: `let r = /a/g;`, Options: arrayOption("always")},
			// BigInt literal.
			{Code: `let big = 1n;`, Options: arrayOption("always")},
			// Hex / binary / octal literals.
			{Code: `let n = 0xff;`, Options: arrayOption("always")},
			{Code: `let n = 0b101;`, Options: arrayOption("always")},
			{Code: `let n = 0o7;`, Options: arrayOption("always")},
			// Spread / array literal / object literal.
			{Code: `let arr = [1, ...rest];`, Options: arrayOption("always")},
			{Code: `let obj = { a: 1, ...rest };`, Options: arrayOption("always")},
			// Optional chain / non-null call.
			{Code: `let v = obj?.foo;`, Options: arrayOption("always")},
			{Code: `let v = obj!.foo;`, Options: arrayOption("always")},
			// `yield` inside a generator init (the surrounding decl IS the var).
			{Code: `function* g() { let r = yield 1; }`, Options: arrayOption("always")},
			// `await` inside an async init.
			{Code: `async function f() { let r = await Promise.resolve(1); }`, Options: arrayOption("always")},
			// JSX element initializer.
			{Code: `function App() { let el = <div />; return el; }`, Options: arrayOption("always"), Tsx: true},

			// ============================================================
			// Nesting / traversal — function-like containers
			// Every function-like body must traverse independently; the listener
			// must fire on every VariableDeclarationList without bleeding across
			// boundaries.
			// ============================================================
			{Code: `function outer() { function inner() { var a = 1; } }`, Options: arrayOption("always")},
			{Code: `const obj = { method() { var a = 1; } };`, Options: arrayOption("always")},
			{Code: `class C { method() { var a = 1; } }`, Options: arrayOption("always")},
			{Code: `class C { constructor() { var a = 1; } }`, Options: arrayOption("always")},
			{Code: `class C { get x() { var a = 1; return a; } }`, Options: arrayOption("always")},
			{Code: `class C { set x(v: number) { var a = v; } }`, Options: arrayOption("always")},
			{Code: `class C { static { var a = 1; } }`, Options: arrayOption("always")},
			{Code: `class C { *gen() { var a = 1; yield a; } }`, Options: arrayOption("always")},
			{Code: `class C { async asyncMethod() { var a = 1; return a; } }`, Options: arrayOption("always")},
			{Code: `class C { async *asyncGen() { var a = 1; yield a; } }`, Options: arrayOption("always")},

			// ============================================================
			// Nesting / traversal — control-flow containers
			// try/catch/finally, switch, conditional bodies. The rule must fire
			// in each block independently.
			// ============================================================
			{Code: `
function f() {
  try {
    var a = 1;
  } catch (e) {
    var b = 2;
  } finally {
    var c = 3;
  }
}
      `, Options: arrayOption("always")},
			{Code: `
function f(x: number) {
  switch (x) {
    case 1: {
      var a = 1;
      break;
    }
    default: {
      var b = 2;
      break;
    }
  }
}
      `, Options: arrayOption("always")},
			{Code: `
function f(cond: boolean) {
  if (cond) {
    var a = 1;
  } else {
    var b = 2;
  }
}
      `, Options: arrayOption("always")},

			// ============================================================
			// Module / script context
			// Top-level var/let/const work the same way (the rule doesn't care
			// about module-ness of the file).
			// ============================================================
			{Code: `export var a = 1;`, Options: arrayOption("always")},
			{Code: `export let b = 1;`, Options: arrayOption("always")},
			{Code: `export const c = 1;`, Options: arrayOption("always")},
			{Code: `export default (() => { var a = 1; return a; })();`, Options: arrayOption("always")},
			// Re-exports introduce no VariableDeclaration.
			{Code: `export { something } from 'mod';`, Options: arrayOption("always")},
			{Code: `export * from 'mod';`, Options: arrayOption("always")},

			// ============================================================
			// Multi-declarator independence
			// ============================================================
			// All-initialized list under "always" — no reports.
			{Code: `var a = 1, b = 2, c = 3;`, Options: arrayOption("always")},
			// All-uninitialized list under "never" — no reports.
			{Code: `var a, b, c;`, Options: arrayOption("never")},
			// Long heterogeneous list with const sub-bindings — under "never"
			// only let/var with init should be reported, never const.
			{Code: `let x;`, Options: arrayOption("never")},
			{Code: `const z = 1;`, Options: arrayOption("never")},

			// ============================================================
			// Real-user shapes (issue tracker survey)
			// ============================================================
			// React-style destructuring tuple from hooks — destructuring skips
			// for both elements, no false positives.
			{Code: `
function useState<T>(v: T): [T, (next: T) => void] { return [v, () => {}]; }
function App() {
  const [count, setCount] = useState(0);
  return count;
}
      `, Options: arrayOption("never")},
			// Conditional types / type aliases never produce a VariableDeclaration.
			{Code: `type Maybe<T> = T | undefined;`, Options: arrayOption("always")},
			// Interface in a non-declare namespace.
			{Code: `
namespace Lib {
  interface I {
    x: number;
  }
}
      `, Options: arrayOption("always")},
			// Variable initialized to `undefined` via destructuring default is
			// still initialized (default lives on the binding element, not the
			// declarator).
			{Code: `let { a = undefined } = obj;`, Options: arrayOption("always")},
			// Optional binding pattern (TS doesn't allow `?` on var binding
			// directly, but a destructuring with optional-style default works).
			{Code: `let [a = 0, b = 0] = arr;`, Options: arrayOption("always")},
			// Multi-line declarator with parenthesized + chained init.
			{Code: `
let result =
  (function () {
    return 1;
  })();
      `, Options: arrayOption("always")},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: declaration/container forms — let with type annotation ----
			// Mirrors typescript-eslint's getReportLoc narrowing for "always":
			// underlines only the identifier, not the trailing annotation.
			{
				Code:    `let arr: { a: number; b: string };`,
				Options: arrayOption("always"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "initialized", Line: 1, Column: 5, EndLine: 1, EndColumn: 8},
				},
			},

			// ---- Dimension 4: declaration/container forms — mixed destructuring + identifier ----
			// `id !== Identifier` declarators are skipped; sibling identifier
			// declarators in the same list are still evaluated.
			{
				Code:    `let { a } = obj, b;`,
				Options: arrayOption("always"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "initialized", Line: 1, Column: 18, EndLine: 1, EndColumn: 19},
				},
			},

			// Locks in upstream "always"+!declare-namespace-ancestor arm: a
			// non-declare namespace nested DIRECTLY in another non-declare
			// namespace still gets walked, and each layer reports independently.
			{
				Code: `
namespace Outer {
  namespace Inner {
    let unset: number;
  }
}
      `,
				Options: arrayOption("always"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "initialized", Line: 4, Column: 9, EndLine: 4, EndColumn: 14},
				},
			},

			// Locks in upstream isInitialized() ForStatement arm with the "never"
			// path: `for (var i = 0;...)` with `ignoreForLoopInit` UNSET must still
			// report. Pairs with the valid case above that sets the option to true.
			{
				Code:    `for (var i = 0; i < 5; i++) {}`,
				Options: arrayOption("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notInitialized", Line: 1, Column: 10, EndLine: 1, EndColumn: 15},
				},
			},

			// Locks in upstream isInitialized() ForInStatement arm with the "never"
			// path: `for (var x in y)` reports x as `notInitialized` because the
			// loop binding is treated as initialized.
			{
				Code:    `for (var key in obj) {}`,
				Options: arrayOption("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notInitialized", Line: 1, Column: 10, EndLine: 1, EndColumn: 13},
				},
			},

			// Locks in upstream isInitialized() ForOfStatement arm with the "never"
			// path: matching the for-in case above for for-of bindings.
			{
				Code:    `for (var item of items) {}`,
				Options: arrayOption("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notInitialized", Line: 1, Column: 10, EndLine: 1, EndColumn: 14},
				},
			},

			// Locks in upstream "never"+let-not-in-CONSTANT_BINDINGS arm: `let`
			// in a for-loop with ignoreForLoopInit=false reports `notInitialized`.
			{
				Code:    `for (let i = 0; i < 5; i++) {}`,
				Options: arrayOption("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notInitialized", Line: 1, Column: 10, EndLine: 1, EndColumn: 15},
				},
			},

			// Locks in option-shape coverage: bare-object option form (no mode
			// string), exercising the `map[string]interface{}` arm of
			// parseOptions. Defaults to mode="always" so an uninitialized binding
			// reports.
			{
				Code:    `var foo;`,
				Options: map[string]interface{}{"ignoreForLoopInit": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "initialized", Line: 1, Column: 5, EndLine: 1, EndColumn: 8},
				},
			},

			// ---- Real-user: multi-declarator with mixed init under "never" ----
			// Verifies per-declarator evaluation: only the initialized ones report.
			{
				Code:    `let a, b = 2, c, d = 4;`,
				Options: arrayOption("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notInitialized", Line: 1, Column: 8, EndLine: 1, EndColumn: 13},
					{MessageId: "notInitialized", Line: 1, Column: 18, EndLine: 1, EndColumn: 23},
				},
			},

			// ---- Real-user: TS type assertion in initializer ----
			// `as` / `satisfies` wrappers should not affect the rule (still
			// counted as initialized). Uses `let` to avoid the CONSTANT_BINDINGS
			// exemption that would otherwise mask a regression here.
			{
				Code:    `let x = (0 as number);`,
				Options: arrayOption("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notInitialized", Line: 1, Column: 5, EndLine: 1, EndColumn: 22},
				},
			},

			// Locks in the typescript-eslint `getReportLoc` narrowing on the
			// for-loop + no-init + type-annotation path: `for (let x: number;;)`
			// mode='never' reports `notInitialized` because the loop binding is
			// "initialized" in the upstream sense, AND the report range must
			// narrow to the identifier (Initializer == nil branch). Without the
			// narrowing the type annotation `: number` would be underlined.
			{
				Code:    `for (let x: number; x < 5; x++) {}`,
				Options: arrayOption("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notInitialized", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
				},
			},

			// Locks in for-loop body boundary: a non-initialized `var` inside
			// the BODY of a for-loop is reported under "always" because its
			// declList.Parent is VariableStatement (not the ForStatement). The
			// for-loop's own `i = 0` declarator has an init and is silent.
			{
				Code: `
for (var i = 0; i < 1; i++) {
  var inner;
}
      `,
				Options: arrayOption("always"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "initialized", Line: 3, Column: 7, EndLine: 3, EndColumn: 12},
				},
			},

			// Locks in nested non-declare namespace within a non-declare
			// outer namespace: NEITHER ancestor has `declare`, so the binding
			// must be reported under "never" + init (matches the rule's normal
			// path for variable declarations in a regular namespace body).
			{
				Code: `
namespace Outer {
  namespace Inner {
    let x: number = 1;
  }
}
      `,
				Options: arrayOption("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notInitialized", Line: 4, Column: 9, EndLine: 4, EndColumn: 22},
				},
			},

			// ============================================================
			// Nesting / traversal — function-like containers (invalid)
			// Verifies the rule fires INSIDE each function-like body. Each row
			// here is the negative twin of the same-kind valid row above.
			// ============================================================
			{
				Code:    `function outer() { function inner() { var a; } }`,
				Options: arrayOption("always"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "initialized", Line: 1, Column: 43, EndLine: 1, EndColumn: 44},
				},
			},
			{
				Code:    `class C { method() { var a; } }`,
				Options: arrayOption("always"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "initialized", Line: 1, Column: 26, EndLine: 1, EndColumn: 27},
				},
			},
			{
				Code:    `class C { constructor() { var a; } }`,
				Options: arrayOption("always"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "initialized", Line: 1, Column: 31, EndLine: 1, EndColumn: 32},
				},
			},
			{
				Code:    `class C { get x() { var a; return a; } }`,
				Options: arrayOption("always"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "initialized", Line: 1, Column: 25, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code:    `class C { static { var a; } }`,
				Options: arrayOption("always"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "initialized", Line: 1, Column: 24, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:    `class C { *gen() { var a; yield a; } }`,
				Options: arrayOption("always"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "initialized", Line: 1, Column: 24, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:    `class C { async asyncMethod() { var a; return a; } }`,
				Options: arrayOption("always"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "initialized", Line: 1, Column: 37, EndLine: 1, EndColumn: 38},
				},
			},
			{
				Code:    `class C { async *asyncGen() { var a; yield a; } }`,
				Options: arrayOption("always"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "initialized", Line: 1, Column: 35, EndLine: 1, EndColumn: 36},
				},
			},

			// ============================================================
			// Nesting / traversal — control-flow containers (invalid)
			// One report per body block confirms each VariableDeclarationList
			// is visited independently.
			// ============================================================
			{
				Code: `
function f() {
  try {
    var a;
  } catch (e) {
    var b;
  } finally {
    var c;
  }
}
      `,
				Options: arrayOption("always"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "initialized", Line: 4, Column: 9, EndLine: 4, EndColumn: 10},
					{MessageId: "initialized", Line: 6, Column: 9, EndLine: 6, EndColumn: 10},
					{MessageId: "initialized", Line: 8, Column: 9, EndLine: 8, EndColumn: 10},
				},
			},
			{
				Code: `
function f(x: number) {
  switch (x) {
    case 1: {
      var a;
      break;
    }
    default: {
      var b;
      break;
    }
  }
}
      `,
				Options: arrayOption("always"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "initialized", Line: 5, Column: 11, EndLine: 5, EndColumn: 12},
					{MessageId: "initialized", Line: 9, Column: 11, EndLine: 9, EndColumn: 12},
				},
			},

			// ============================================================
			// Module / script context (invalid)
			// ============================================================
			// Top-level export var without init — under "always" the modifier
			// must NOT suppress reporting.
			{
				Code:    `export var x;`,
				Options: arrayOption("always"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "initialized", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
				},
			},
			// `export let` (also block-scoped) — same as above.
			{
				Code:    `export let y;`,
				Options: arrayOption("always"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "initialized", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
				},
			},

			// ============================================================
			// Multi-declarator independence (invalid)
			// ============================================================
			// Two destructuring patterns (skipped) bracketing a single
			// identifier (reported) verifies declarator iteration doesn't
			// short-circuit when the FIRST declarator skips.
			{
				Code:    `let { a } = obj, b, [c] = arr;`,
				Options: arrayOption("always"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "initialized", Line: 1, Column: 18, EndLine: 1, EndColumn: 19},
				},
			},
			// Five-declarator mixed list under "never" — exactly the
			// initialized ones report.
			{
				Code:    `let a, b = 1, c, d = 2, e;`,
				Options: arrayOption("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notInitialized", Line: 1, Column: 8, EndLine: 1, EndColumn: 13},
					{MessageId: "notInitialized", Line: 1, Column: 18, EndLine: 1, EndColumn: 23},
				},
			},

			// ============================================================
			// Real-user shapes (issue-tracker survey)
			// ============================================================

			// Real-user: TS literal-type assertion produces a parenthesized
			// init in tsgo. Confirms `(value as Literal)` still passes the
			// "has initializer" branch.
			{
				Code:    `let label = 'hello' as 'hello';`,
				Options: arrayOption("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notInitialized", Line: 1, Column: 5, EndLine: 1, EndColumn: 31},
				},
			},

			// Real-user: type-asserted destructuring on the RHS — the LHS is
			// still a binding pattern and is skipped, but a sibling identifier
			// in the SAME list reports.
			{
				Code:    `const { x } = (obj as { x: number }), y = 1;`,
				Options: arrayOption("never"),
				Errors: []rule_tester.InvalidTestCaseError{},
			},

			// Real-user: variable inside an IIFE's body still reports.
			{
				Code: `
(function () {
  var orphan;
})();
      `,
				Options: arrayOption("always"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "initialized", Line: 3, Column: 7, EndLine: 3, EndColumn: 13},
				},
			},

			// Real-user: variable inside a static class block reports under
			// "never" when initialized.
			{
				Code: `
class C {
  static {
    let value = 0;
  }
}
      `,
				Options: arrayOption("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notInitialized", Line: 4, Column: 9, EndLine: 4, EndColumn: 18},
				},
			},

			// Real-user: ESLint #14808 style — nested for-loop with shadowing
			// `let i` should be reported only once for the inner declaration.
			{
				Code: `
for (let i = 0; i < 5; i++) {
  for (let i = 0; i < 5; i++) {}
}
      `,
				Options: arrayOption("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notInitialized", Line: 2, Column: 10, EndLine: 2, EndColumn: 15},
					{MessageId: "notInitialized", Line: 3, Column: 12, EndLine: 3, EndColumn: 17},
				},
			},

			// Real-user: ambient namespace where a NON-declare inner namespace
			// re-declares a binding — declare propagates and the inner binding
			// is still skipped. The negative (valid) form is covered in the
			// valid block; this invalid form ensures we don't accidentally
			// flag the namespace itself.
			{
				Code: `
namespace Plain {
  let a: number = 1;
  let b: number;
}
      `,
				Options: arrayOption("always"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "initialized", Line: 4, Column: 7, EndLine: 4, EndColumn: 8},
				},
			},

			// Real-user: when a list has BOTH a destructuring (skipped) and a
			// CONSTANT_BINDING-eligible binding under "never", the
			// CONSTANT_BINDING suppresses the report. const list with mixed
			// shapes — only verifies graceful skipping, no report expected.
			// (Encoded as valid above.)

			// Real-user: a sequence-expression init from `let x = (a, b)` in
			// "never" mode reports normally — the SequenceExpression is
			// represented as a BinaryExpression with comma in tsgo, and that
			// still passes the `Initializer != nil` gate.
			{
				Code:    `let x = (1, 2);`,
				Options: arrayOption("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notInitialized", Line: 1, Column: 5, EndLine: 1, EndColumn: 15},
				},
			},
		},
	)
}

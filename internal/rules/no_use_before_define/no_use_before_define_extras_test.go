// TestNoUseBeforeDefineExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it covers,
// so future refactors can't silently regress them without breaking a named
// lock-in. The migrated upstream cases live in
// no_use_before_define_upstream_test.go and
// no_use_before_define_upstream_typescript_test.go.
package no_use_before_define

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUseBeforeDefineExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUseBeforeDefineRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: access / key forms — only the object of a member
			// access reads a binding, never the member name ----
			{Code: `const o = { a: 1 }; let a;`},
			{Code: `const o = { "a": 1 }; let a;`},
			{Code: `const o = { 0: 1 }; let zero;`},
			{Code: `const o = { m() {} }; let m;`},
			{Code: `obj.a; let a;`},
			{Code: `obj?.a; let a;`},
			{Code: `obj.a.b.c; let b;`},
			{Code: `obj["a"]; let a;`},
			{Code: `class C { #a = 1; read() { return this.#a; } }`},
			{Code: `interface I { a: string } let a;`},
			{Code: `enum E { A } let A;`},
			{Code: `type T = { a: string }; let a;`},

			// ---- Dimension 4: labels are not bindings ----
			{Code: `outer: while (true) { break outer; } let outer;`},
			{Code: `loop: for (;;) { continue loop; } let loop;`},

			// ---- Dimension 4: declaration / container forms ----
			{Code: `function f() {} f();`},
			{Code: `const f = function () {}; f();`},
			{Code: `const f = () => {}; f();`},
			{Code: `async function f() {} f();`},
			{Code: `function* f() {} f();`},
			{Code: `async function* f() {} f();`},
			{Code: `class C { handler = () => C; }`},
			{Code: `class C { static async m() { return C; } }`},
			{Code: `class C { *gen() { yield C; } }`},
			{Code: `const C = class Named { m() { return Named; } };`},

			// ---- Dimension 4: nesting / traversal boundaries ----
			{Code: `function outer() { var x; function inner() { x; } }`},
			{Code: `class Outer { m() { class Inner { m2() { return Outer; } } return Inner; } }`},
			{Code: `class C { static { class D { m() { return C; } } } }`},

			// ---- Dimension 4: graceful degradation ----
			{Code: `class C {} new C();`},
			{Code: `function f() {} f();`},
			{Code: `const a = 1; const { } = { a };`},
			{Code: `const a = 1; const o = { ...a };`},
			{Code: `const a = 1; const [ , b ] = [a, a];`},

			// ---- Dimension 4: body-absent forms ----
			// Overload signatures must be adjacent to the implementation in
			// TypeScript, so a reference can never land between them; both the
			// signature-first and implementation-first orders resolve to the
			// same merged binding.
			{Code: `function f(a: string): void; function f(a: number): void; function f(a: any) {} f(1);`},
			{Code: `declare function g(): void; g();`},
			{Code: `declare class K { m(): void; } new K();`},
			{Code: `abstract class A { abstract m(): void; } class B extends A { m() {} }`},

			// N/A: Dimension 3 (autofix boundaries) — the rule reports only, it
			// has no autofix and no suggestions.
			// N/A: Dimension 4 receiver-wrapper rows for `X!.y` / `(X as T).y` on
			// the *reported* node — the rule always reports the bare identifier,
			// never a wrapped expression; the wrapper rows are exercised on the
			// invalid side instead, where they must not hide the reference.

			// ---- Locks in upstream isEvaluatedDuringInitialization() arm 3:
			// SENTINEL_TYPE break, reached through a parameter with no default ----
			{Code: `function f(a) { a; }`},
			// ESTree wraps a shorthand method body in a FunctionExpression,
			// which is a SENTINEL_TYPE; tsgo has no wrapper, so the walk has to
			// stop at the member node itself. Otherwise it escapes the method
			// and finds the enclosing initializer, flagging every parameter
			// read inside. Caught by a differential run against ESLint.
			{Code: `const api = { get(node, opts) { return node.range[opts]; } };`},
			{Code: `const api = { set value(node) { use(node); } };`},
			{Code: `const api = { get value() { return 1; } };`},
			{Code: `class C { m(node, opts) { return node.range[opts]; } }`},
			{Code: `class C { constructor(node) { this.node = node; } }`},
			{Code: `class C { get value() { return 1; } set value(node) { use(node); } }`},
			{Code: `const api = { *gen(node) { yield node; } };`},
			{Code: `const api = { async run(node) { return node; } };`},
			// ---- Dimension 4: an `import(...)` type qualifier names an export
			// of the other module, not a binding here. Caught by a differential
			// run against ESLint. ----
			{Code: `type T = typeof import('./m').later; let later;`},
			{Code: `type T = typeof import('./m').a.b; let a;`},
			{Code: `type T = import('./m').Later; let Later;`},
			// ---- Locks in upstream isEvaluatedDuringInitialization() arm 3:
			// SENTINEL_TYPE break at an ImportDeclaration ----
			{Code: `import { a } from './m'; a;`},
			// ---- Locks in upstream isEvaluatedDuringInitialization() arm 1:
			// VariableDeclarator with no initializer falls through to `return false` ----
			{Code: `var a; a;`},
			// ---- Locks in upstream isInRange(): a nil range never contains a
			// location — a static field with no value can't cover a computed key ----
			{Code: `class C { static blank; }`},

			// ---- Locks in upstream isClassStaticInitializerRange() arm 1:
			// a static block covers the location ----
			{Code: `class C { x = 1; static { C; } }`},
			// ---- Locks in upstream isClassStaticInitializerRange() arm 2:
			// a static field's value covers the location ----
			{Code: `class C { static x = C; }`},

			// ---- Locks in upstream shouldCheck() TSQualifiedName arm: a
			// namespace member is not a binding, so only the left-most name is
			// checked ----
			{Code: `namespace A { export const Later = 1; } import Z = A.Later; const Later = 2;`},

			// ---- Locks in upstream shouldCheck() referenceContainsTypeQuery():
			// a nested `typeof A.B.C` behind a qualified name is still a type query ----
			{Code: `interface I { t: typeof Ns.Inner.Deep } namespace Ns { export namespace Inner { export const Deep = 1; } }`},

			// ---- Locks in upstream shouldCheck() isClassRefInClassDecorator():
			// a class self-reference inside its own decorator is safe ----
			{Code: `@register(Widget) class Widget {}`},

			// ---- Options: each boolean asserted in its non-default direction ----
			{Code: `f(); function f() {}`, Options: map[string]any{"functions": false}},
			{Code: `function make() { return new C(); } class C {}`, Options: map[string]any{"classes": false}},
			{Code: `function read() { return v; } let v;`, Options: map[string]any{"variables": false}},
			{Code: `export { later }; const later = 1;`, Options: map[string]any{"allowNamedExports": true}},
			{Code: `const v = E.A; enum E { A }`, Options: map[string]any{"enums": false}},
			{Code: `let v: Later; type Later = string;`, Options: map[string]any{"typedefs": false, "ignoreTypeReferences": false}},
			{Code: `let v: Later; type Later = string;`, Options: map[string]any{"ignoreTypeReferences": true}},
			// All seven options set explicitly to their defaults must behave
			// exactly like passing no options at all.
			{Code: `const a = 1; a;`, Options: map[string]any{
				"functions": true, "classes": true, "variables": true,
				"allowNamedExports": false, "enums": true, "typedefs": true,
				"ignoreTypeReferences": true,
			}},

			// ---- Real-user: mutually recursive function declarations — the shape
			// `functions: false` exists for ----
			{Code: `function isEven(n) { return n === 0 ? true : isOdd(n - 1); } function isOdd(n) { return n === 0 ? false : isEven(n - 1); }`, Options: map[string]any{"functions": false}},
			// ---- Real-user: circular type references (typescript-eslint#2572
			// family) — a type alias may name a type declared later ----
			{Code: `interface Node { children: Leaf[] } interface Leaf { parent: Node }`},
			{Code: `type Tree = { left: Branch }; type Branch = { value: Tree };`},
			// ---- Real-user: typescript-eslint#2824 — Angular's forwardRef
			// pattern names the class from inside its own decorator metadata ----
			{Code: `
@Component({
  providers: [{ provide: NG_VALUE_ACCESSOR, useExisting: forwardRef(() => MyControl) }],
})
export class MyControl {}
`},
			// ---- Real-user: a component referenced from a callback that runs
			// later is a separate execution context, which `variables: false`
			// keys off ----
			{Code: `const List = () => items.map(() => Row()); const Row = () => null; const items = [];`, Tsx: true, Options: map[string]any{"variables": false}},
			// ---- Real-user: JSX intrinsic elements are strings, not bindings,
			// so a later binding with the same name is irrelevant ----
			{Code: `<div />; let div;`, Tsx: true},
			{Code: `<a:b />; let a;`, Tsx: true},

			// ---- Type space and value space are resolved separately: a name
			// only answers a reference that reads the space it declares ----
			{Code: `
const X = 1;
function f() {
  X;
  type X = string;
}
`},
			{Code: `
type X = string;
function f() {
  let y: X;
  const X = 1;
}
`, Options: map[string]any{"ignoreTypeReferences": false}},

			// ---- A parameter default is evaluated before the function body's
			// lexical environment exists, so it reads the outer binding ----
			{Code: `
const X = 0;

function f(a = X) {
  const X = 1;
}
`},
			{Code: `
const X = 0;

const f = (a = X) => {
  const X = 1;
};
`},
			// A body binding that also has a parameter declaration stays visible
			// to the defaults — only body-only names are hidden.
			{Code: `
function f(X, a = X) {
  var X = 1;
}
`},

			// ---- Type parameters are `Type` definitions upstream, so
			// `typedefs` exempts them ----
			{Code: `type Q<T extends U, U = string> = T;`, Options: map[string]any{"typedefs": false, "ignoreTypeReferences": false}},

			// ---- `enums` exempts a reference whatever execution context it
			// sits in, unlike `classes` and `variables` ----
			{Code: `const v = E.A; enum E { A }`, Options: map[string]any{"enums": false}},
			// ---- A name this file never declares is never reported, not even
			// as the local half of a named export ----
			{Code: `export { nothing };`},

			// ---- `ignoreTypeReferences` exempts every type position, however
			// the dotted name is spelled, plus the `export =` operand, which
			// names whichever space its binding lives in ----
			{Code: `class C implements Later {} interface Later {}`},
			{Code: `class C implements NS.Later {} namespace NS { export interface Later {} }`},
			{Code: `interface C extends NS.Later {} namespace NS { export interface Later {} }`},
			{Code: `let x: NS.Later; namespace NS { export type Later = string; }`},
			{Code: `export = X; const X = 1;`},
			// A value-only binding does not intercept the dotted name of a
			// heritage clause, whichever way that name is spelled.
			{Code: `
namespace NS {
  export interface T {}
}

function f() {
  class C implements NS.T {}
  const NS = 1;
}
`},
			{Code: `
namespace NS {
  export interface T {}
}

function f() {
  interface C extends NS.T {}
  const NS = 1;
}
`},

			// ---- Nothing in a function type's signature runs, so a reference
			// from one is never reported ----
			{Code: `let f: (x: Later) => void; type Later = string;`, Options: map[string]any{"ignoreTypeReferences": false}},
			{Code: `let f: (this: Later) => void; type Later = string;`, Options: map[string]any{"ignoreTypeReferences": false}},
			{Code: `let f: new (x: Later) => void; type Later = string;`, Options: map[string]any{"ignoreTypeReferences": false}},
			{Code: `let f: <T extends Later>() => void; type Later = string;`, Options: map[string]any{"ignoreTypeReferences": false}},
			{Code: `interface I { (x: Later): void } type Later = string;`, Options: map[string]any{"ignoreTypeReferences": false}},
			{Code: `interface I { new (x: Later): void } type Later = string;`, Options: map[string]any{"ignoreTypeReferences": false}},
			{Code: `interface I { m(x: Later): void } type Later = string;`, Options: map[string]any{"ignoreTypeReferences": false}},

			// ---- A `this` pseudo-parameter declares no binding, but its type
			// annotation is a real type reference ----
			{Code: `function f(this: Later) {} type Later = unknown;`},
			{Code: `function f(this: number) { return this; }`},

			// ---- TS resolves a string-literal member name to a real name, so
			// the reference reads the member — which has no identifier to
			// report at ----
			{Code: "enum E {\n  A = ((): number => B)(),\n  \"B\" = 1,\n}", Options: map[string]any{"variables": false}},
			{Code: "enum E {\n  A = B,\n  \"B\" = 2,\n}\nconst B = 1;"},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: receiver / expression wrappers must not hide the
			// reference. tsgo keeps parentheses, non-null, and type-assertion
			// wrappers as real nodes where ESTree flattens or omits them. ----
			{
				Code:   `(a).b; var a = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 2, EndLine: 1, EndColumn: 3}},
			},
			{
				Code:   `((a)).b; var a = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 3}},
			},
			{
				Code:   `a!.b; var a = { b: 1 };`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 1}},
			},
			{
				Code:   `(a as any).b; var a = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 2}},
			},
			{
				Code:   `(a satisfies unknown); var a = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 2}},
			},
			{
				Code:   `a?.(); var a = () => {};`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 1}},
			},
			{
				Code:   `(a); var a = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 2}},
			},

			// ---- Dimension 4: access / key forms — the computed key and the
			// element-access argument DO read bindings ----
			{
				Code:   `const o = { [a]: 1 }; let a;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 14}},
			},
			{
				Code:   `const o = { a }; let a;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 13}},
			},
			{
				Code:   `obj[a]; let a;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 5}},
			},

			// ---- Dimension 4: declaration / container forms ----
			{
				Code:   `f(); async function f() {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 1}},
			},
			{
				Code:   `f(); function* f() {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 1}},
			},
			{
				Code:   `f(); async function* f() {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 1}},
			},
			{
				// A class-field arrow is a separate execution context, so this
				// reports only because the binding is declared later.
				Code:   `class C { handler = () => later; } let later;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 27}},
			},

			// ---- Dimension 4: nesting / traversal boundaries — the inner
			// function must not "inherit" the outer scope's later binding ----
			{
				Code:   `function outer() { function inner() { x; } var x; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 39}},
			},
			{
				Code:   `class Outer { m() { class Inner { m2() { return Later; } } } } class Later {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 49}},
			},

			// ---- Dimension 4: graceful degradation ----
			{
				Code:   `const o = { ...a }; let a;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 16}},
			},
			{
				Code:   `const { ...r } = r;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 18}},
			},
			{
				Code:   `const {} = a; let a;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 12}},
			},

			// ---- Dimension 4: body-absent forms ----
			{
				Code:   `g(); declare function g(): void;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 1}},
			},
			{
				Code:   `f(1); function f(a: string): void; function f(a: number): void; function f(a: any) {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 1}},
			},

			// ---- Locks in upstream isInClassStaticInitializerRange() arm 2's
			// `classMember.value &&` guard: a static field with no value does not
			// cover the location, so the computed key is still a TDZ read ----
			{
				Code:    `class C { static blank; [C]; }`,
				Options: map[string]any{"classes": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 26}},
			},

			// ---- Locks in upstream isClassRefInClassDecorator() arm 1: a
			// non-class binding referenced from a decorator is NOT exempt ----
			{
				Code:   `@register(setting) class Widget {} let setting;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 11}},
			},

			// ---- Locks in upstream shouldCheck() TSQualifiedName arm: the
			// left-most name of a qualified reference IS checked ----
			{
				Code:   `import Z = A.X; namespace A { export const X = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 12}},
			},

			// ---- Locks in upstream isFromSeparateExecutionContext(): a static
			// block nested inside a static field initializer folds through two
			// levels back to the class-definition context ----
			{
				Code:    `const C = class { static field = class { static { C; } }; };`,
				Options: map[string]any{"variables": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 51}},
			},

			// ---- Options: each boolean asserted in its default direction ----
			{
				Code:    `f(); function f() {}`,
				Options: map[string]any{"functions": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 1}},
			},
			{
				Code:    `new C(); class C {}`,
				Options: map[string]any{"classes": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 5}},
			},
			{
				Code:    `function read() { return v; } let v;`,
				Options: map[string]any{"variables": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 26}},
			},
			{
				Code:    `const v = E.A; enum E { A }`,
				Options: map[string]any{"enums": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 11}},
			},
			{
				Code:    `let v: Later; type Later = string;`,
				Options: map[string]any{"typedefs": true, "ignoreTypeReferences": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 8}},
			},

			// ---- Real-user: JSX member expressions read their left-most object
			// whatever its casing, unlike a bare lower-case intrinsic tag ----
			{
				Code:   `<ns.Widget />; let ns;`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 2, EndLine: 1, EndColumn: 4}},
			},
			// ---- Real-user: a component referenced from a JSX attribute value ----
			{
				Code: `<Host render={Later} />; function Host() { return null } function Later() { return null }`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "usedBeforeDefined", Line: 1, Column: 2},
					{MessageId: "usedBeforeDefined", Line: 1, Column: 15},
				},
			},
			// ---- Real-user: mutual recursion IS reported under the default
			// `functions: true` — the forward call really does run first ----
			{
				Code:   `function isEven(n) { return isOdd(n); } function isOdd(n) { return isEven(n); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 29}},
			},
			// ---- Dimension 4: a `let`/`const` loop binding is in its temporal
			// dead zone while the iterated expression is evaluated ----
			{
				Code:   `for (let x of x) {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 15}},
			},
			{
				Code:   `for (let x in x) {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 15}},
			},
			// ---- Dimension 4: enum members are bindings in the enum's own
			// scope, so one member may not read a later sibling ----
			{
				Code:   `enum E { A = B, B = 1 }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 14}},
			},

			// ---- Real-user: multi-line report positions ----
			{
				Code: `
function render() {
  return Widget();
}

const Widget = () => null;
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "usedBeforeDefined", Message: "'Widget' was used before it was defined.", Line: 3, Column: 10, EndLine: 3, EndColumn: 16},
				},
			},

			// ---- A type parameter's constraint and default are references in
			// the scope that declares the parameter ----
			{
				Code: `
class C<T extends Later> {}
type Later = string;
`,
				Options: map[string]any{"ignoreTypeReferences": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 2, Column: 19}},
			},
			{
				Code: `
function f<T = Later>() {}
type Later = string;
`,
				Options: map[string]any{"ignoreTypeReferences": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 2, Column: 16}},
			},
			{
				Code:    `type Q<T extends U, U = string> = T;`,
				Options: map[string]any{"typedefs": true, "ignoreTypeReferences": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 18}},
			},

			// ---- A class index signature's key type and result type are
			// references too ----
			{
				Code: `
class C {
  [key: string]: Later;
}

type Later = string;
`,
				Options: map[string]any{"ignoreTypeReferences": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 3, Column: 18}},
			},

			// ---- Enum members are `TSEnumMemberName` definitions upstream, so
			// neither `variables` nor `enums` exempts them ----
			{
				Code: `
enum E {
  A = ((): number => B)(),
  B = 1,
}
`,
				Options: map[string]any{"variables": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 3, Column: 22}},
			},

			// ---- A heritage clause spells its dotted name with property
			// accesses; only `class ... extends` reads a value there ----
			{
				Code:   `class C extends NS.Base {} const NS = { Base: class {} };`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 17}},
			},
			{
				Code:    `class C implements NS.T {} namespace NS { export interface T {} }`,
				Options: map[string]any{"ignoreTypeReferences": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 20}},
			},

			// ---- A `this` pseudo-parameter's type annotation is walked like
			// any other parameter's ----
			{
				Code:    `function f(this: Later) {} type Later = unknown;`,
				Options: map[string]any{"ignoreTypeReferences": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 18}},
			},
			{
				Code:    `class K { m(this: Later) {} } type Later = unknown;`,
				Options: map[string]any{"ignoreTypeReferences": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 19}},
			},
			{
				Code:    `function f(this: NS.Later) {} namespace NS { export type Later = unknown; }`,
				Options: map[string]any{"ignoreTypeReferences": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 18}},
			},
		},
	)
}

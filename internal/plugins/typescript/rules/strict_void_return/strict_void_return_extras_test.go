// TestStrictVoidReturnExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it covers,
// so future refactors can't silently regress them without breaking a named
// lock-in. Upstream parity cases live in strict_void_return_upstream_test.go.
package strict_void_return

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestStrictVoidReturnExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &StrictVoidReturnRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: receiver wrappers — parenthesized callee ----
		// tsgo preserves ParenthesizedExpression nodes that ESTree flattens; the
		// callee here is `(foo)` and must still resolve through to the call sig.
		{Code: `
declare function foo(cb: () => void): void;
(foo)(() => {});
`},
		// ---- Dimension 4: receiver wrappers — multi-level parenthesized callee ----
		{Code: `
declare function foo(cb: () => void): void;
((foo))(() => {});
`},
		// ---- Dimension 4: receiver wrappers — non-null on callee ----
		// `foo!(cb)` exercises NonNullExpression around the callee.
		{Code: `
declare function foo(cb: () => void): void | null;
foo!(() => {});
`},
		// ---- Dimension 4: receiver wrappers — optional chain callee ----
		// `foo?.(cb)` is a CallExpression with the optional-chain flag in tsgo;
		// no ChainExpression wrapper, so the argument check still runs.
		{Code: `
declare const foo: ((cb: () => void) => void) | null;
foo?.(() => {});
`},
		// ---- Dimension 4: element-access callee ----
		// `obj['m'](cb)` — callee is ElementAccessExpression, not PropertyAccess.
		{Code: `
declare const obj: { m(cb: () => void): void };
obj['m'](() => {});
`},
		// ---- N/A: SpreadElement in array literal expected to be skipped ----
		// Spread cannot bind to a single void-fn slot, so the rule must not
		// inspect spread elements.
		{Code: `
declare const cbs: Array<() => void>;
const arr: Array<() => void> = [...cbs];
`},
		// ---- N/A: omitted (sparse) array element ----
		// `[, () => {}]` includes an OmittedExpression for the missing slot.
		{Code: `
const arr: Array<() => void> = [, () => {}];
`},
		// ---- Dimension 4: as-assertion on function-typed value ----
		// The value is a `() => number as () => void` — getApparentType strips
		// the assertion, so the underlying number return triggers nonVoidFunc
		// only when the type actually still reports a value. The assertion sets
		// the type to `() => void`, so we expect this to pass.
		{Code: `
declare function cb(): number;
const foo: () => void = cb as () => void;
`},
		// ---- Dimension 4: satisfies on function-typed value ----
		// `satisfies` preserves the original (`() => void`) inferred type. A
		// truly void function should remain valid through the wrapper.
		{Code: `
const foo: () => void = (() => {}) satisfies () => void;
`},
		// ---- Branch lock-in: empty arg list, allReturnsAllowed early return ----
		// `new Foo` with no args must not crash the argument loop.
		{Code: `
declare class Foo { constructor(cb?: () => void); }
new Foo();
`},
		// ---- Branch lock-in: compound assignment += (non-void-fn RHS context) ----
		// Upstream's note: `+=` / `-=` are inherently safe because the RHS
		// contextual type is numeric — our BinaryExpression listener should not
		// flag these.
		{Code: `
let n = 0;
n += 1;
`},
		// ---- Branch lock-in: object literal with no contextual type ----
		// `let foo = { cb() { return 1; } }` — no contextual type, skipped.
		{Code: `
let foo = { cb() { return 1; } };
`},
		// ---- Branch lock-in: object literal where contextual type lacks the property ----
		{Code: `
interface Foo { fn(): void }
let foo: Foo = { other() { return 1; } } as any;
`},
		// ---- Branch lock-in: shorthand-property pattern with object-assignment-initializer ----
		// `{ cb = () => 1 }` inside an obj literal — upstream comment: "don't
		// check this thing". The shorthand's name is checked against contextual
		// type; the ObjectAssignmentInitializer is ignored.
		{Code: `
declare let foo: { cb: () => void };
foo = {
  cb = () => 1,
};
`},
		// ---- Branch lock-in: class field with no initializer ----
		// `propNode.value == null` early return.
		{Code: `
class Foo { cb: () => void; }
`},
		// ---- Branch lock-in: class member with no body (overload signature) ----
		// `methodNode.value.type === TSEmptyBodyFunctionExpression` early return.
		{Code: `
class Bar { foo() {} }
class Foo extends Bar {
  foo();
  foo() {}
}
`},
		// ---- Branch lock-in: constructor in derived class is never reported ----
		// Upstream's MethodDefinition listener covers constructor too, but
		// the base's constructor isn't exposed via getPropertyOfType, so
		// the heritage walk never matches. Negative lock-in for matrix parity.
		{Code: `
class Foo { constructor() {} }
class Bar extends Foo {
  constructor() {
    super();
  }
}
`},
		// ---- Branch lock-in: object-literal getter returning a value-fn ----
		// `{ get cb() { return () => 1 } }` inside `{ cb: () => void }` —
		// upstream's checkObjectPropertyNode walks the getter's body, so
		// rslint must too. Negative lock-in: getter returning `() => {}`
		// (a true void function) is fine.
		{Code: `
const obj: { cb: () => void } = {
  get cb() {
    return () => {};
  },
};
`},
		// ---- Branch lock-in: return statement with no argument ----
		// `return;` inside an arrow whose context expects void — no contextual
		// type for empty return, skipped.
		{Code: `
declare function foo(cb: () => void): void;
foo(() => {
  return;
});
`},
		// ---- Branch lock-in: VariableDeclaration with no initializer ----
		// Pure declaration — `if (init != nil)` guard skips.
		{Code: `
let foo: () => void;
`},
		// ---- Branch lock-in: JSX attribute with no initializer ----
		// `<Foo cb />` — `init == nil` guard skips.
		{Code: `
declare function Foo(props: { cb?: () => void }): unknown;
const _ = <Foo />;
`, Tsx: true},
		// ---- Branch lock-in: JSX attribute string-literal value ----
		// `<Foo cb="x" />` — Initializer kind is StringLiteral, not JsxExpression.
		{Code: `
declare function Foo(props: { cb?: string }): unknown;
const _ = <Foo cb="x" />;
`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Dimension 4: parenthesized callee with non-void-returning callback ----
		// Confirms the argument check still fires when callee is wrapped in
		// parens that tsgo preserves.
		{Code: `
declare function foo(cb: () => void): void;
(foo)(() => 1);
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 3, Column: 13},
		}},
		// ---- Dimension 4: non-null callee with non-void-returning callback ----
		// Mirrors the upstream `foo!(cb)` case but with arrow + literal body.
		{Code: `
declare function foo(cb: () => void): void | null;
foo!(() => 1);
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 3, Column: 12},
		}},
		// ---- Dimension 4: optional-chain callee with bad callback ----
		// Optional chain on the callee shouldn't bypass the void-function check.
		{Code: `
declare const foo: ((cb: () => void) => void) | null;
foo?.(() => 1);
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 3, Column: 13},
		}},
		// ---- Dimension 4: element-access callee with bad callback ----
		{Code: `
declare const obj: { m(cb: () => void): void };
obj['m'](() => 1);
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 3, Column: 16},
		}},
		// ---- Branch lock-in: spread is skipped, but other elements still checked ----
		// Confirms the SpreadElement skip in the array listener doesn't mask
		// adjacent siblings.
		{Code: `
declare function cb(): number;
declare const cbs: Array<() => void>;
const arr: Array<(() => void) | false> = [false, cb, ...cbs];
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 4, Column: 50},
		}},
		// ---- Branch lock-in: omitted array element is skipped, others still checked ----
		{Code: `
declare function cb(): number;
const arr: Array<(() => void) | undefined> = [, cb];
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 3, Column: 49},
		}},
		// ---- Branch lock-in: object literal spread is skipped, props still checked ----
		{Code: `
declare function cb(): number;
declare const rest: Record<string, () => void>;
const foo: Record<string, () => void> = {
  ...rest,
  bad: cb,
};
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 6, Column: 8},
		}},
		// ---- Branch lock-in: ReturnStatement with non-void contextual type ----
		// Mirrors the inner-arrow `return cb;` reporting that upstream tests
		// indirectly via larger fixtures; lock it down in a minimal shape.
		{Code: `
declare let foo: () => () => void;
foo = () => {
  return () => 1;
};
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 4, Column: 16},
		}},
		// ---- Branch lock-in: nested arrow concise body returning non-void ----
		{Code: `
declare let foo: () => () => void;
foo = () => () => 1;
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 3, Column: 19},
		}},
		// ---- Branch lock-in: generator class member overriding void base ----
		// Generator branch in reportIfNonVoidFunction reports at function head.
		{Code: `
class Foo { cb() {} }
class Bar extends Foo {
  *cb() { yield 1; }
}
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 4, Column: 3},
		}},
		// ---- Real-user: array forEach with sync callback returning a value ----
		// The common dead-code shape (`arr.forEach(x => x * 2)`) — reported.
		{Code: `
[1, 2, 3].forEach(x => x * 2);
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 2, Column: 24},
		}},
		// ---- Real-user: event handler returning a promise (async arrow) ----
		// addEventListener('click', async () => {...}) is the canonical real
		// case the rule is meant to catch.
		{Code: `
declare const el: { addEventListener(t: string, cb: (e: Event) => void): void };
el.addEventListener('click', async e => {
  await Promise.resolve();
});
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "asyncFunc", Line: 3, Column: 38},
		}},
		// ---- Dimension 4: paren-wrapped funcNode passed as callback ----
		// Without SkipParentheses on funcNode entry, the outer paren would
		// shadow the inner arrow's Kind and the rule would degrade to
		// nonVoidFunc at the paren position.
		{Code: `
declare function foo(cb: () => void): void;
foo(((() => 1)));
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 3, Column: 13},
		}},
		// ---- Dimension 4: multi-level paren-wrapped funcNode in assignment ----
		{Code: `
const cb: () => void = (((() => 1)));
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 2, Column: 33},
		}},
		// ---- Dimension 4: AsExpression-wrapped value (TS type-expression wrapper) ----
		// `cb as () => number` — apparent type still has number return; the
		// outer AsExpression isn't a function literal, so we report
		// nonVoidFunc on the whole AsExpression node.
		{Code: `
declare function cb(): number;
const foo: () => void = cb as () => string;
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 3, Column: 25},
		}},
		// ---- Dimension 4: multi-level class inheritance (3 levels) ----
		// Verifies the heritage walk doesn't stop at the immediate parent —
		// `getPropertyOfType` traverses the prototype chain transparently,
		// so a `cb` declared on the grandparent still triggers reporting.
		{Code: `
class A { cb() {} }
class B extends A {}
class C extends B {
  cb() { return 1; }
}
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 5, Column: 10},
		}},
		// ---- Branch lock-in: NewExpression with no arguments ----
		// `node.Arguments()` returns nil/empty — the for-range must not panic
		// and the rule must produce no diagnostic.
		// (Covered by sibling valid case above — listed here as negative.)
		{Code: `
declare const Foo: { new (cb: () => void): void };
declare function cb(): string;
new Foo(cb);
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 4, Column: 9},
		}},
		// ---- Branch lock-in: nested ObjectExpression with method shorthand ----
		// Nested object literal — outer's contextual prop must resolve the
		// inner method's contextual type through two layers of context.
		{Code: `
declare let foo: { inner: { cb(): void } };
foo = { inner: { cb() { return 1; } } };
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 3, Column: 25},
		}},
		// ---- Real-user: Promise constructor with async executor ----
		// `new Promise(async (resolve) => ...)` is a canonical anti-pattern —
		// the executor signature is `(resolve, reject) => void` so async
		// executors are reported. Head loc is the `=>` token.
		{Code: `
new Promise<number>(async resolve => {
  resolve(1);
});
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "asyncFunc", Line: 2, Column: 35},
		}},
		// ---- Real-user: lib.d.ts `setTimeout` accepts loose `Function` ----
		// Negative lock-in: lib.d.ts types `setTimeout`'s handler as
		// `TimerHandler = string | Function`, which has signatures returning
		// `any` (not void-only). The rule must NOT fire on the global
		// setTimeout — same as ESLint behaviour, since the type isn't a
		// void-returning function type per the rule's definition.
		{Code: `
setTimeout(async () => {
  await Promise.resolve();
}, 100);
`},
		// ---- Real-user: strictly-typed timer-style API with async callback ----
		// In contrast to global setTimeout, a project-local typed timer with
		// a `() => void` callback should report — this is the pattern most
		// application code hits.
		{Code: `
declare function scheduleTask(cb: () => void, ms: number): number;
scheduleTask(async () => {
  await Promise.resolve();
}, 100);
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "asyncFunc", Line: 3, Column: 23},
		}},
		// ---- Real-user: Express-style request handler async ----
		// Web frameworks like Express type handlers as `(req, res) => void`;
		// async handlers leak unhandled rejections. Head loc is `=>`.
		{Code: `
interface Req {}
interface Res { send(s: string): void; }
type Handler = (req: Req, res: Res) => void;
declare function get(path: string, handler: Handler): void;
get('/', async (req, res) => {
  res.send('ok');
});
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "asyncFunc", Line: 6, Column: 27},
		}},
		// ---- Dimension 4: AsExpression on callback identifier in call arg ----
		// `foo(cb as () => number)` — outer AsExpression isn't a function
		// literal, must report nonVoidFunc on the whole AsExpression.
		{Code: `
declare function foo(cb: () => void): void;
declare function cb(): number;
foo(cb as () => number);
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 4, Column: 5},
		}},
		// ---- Dimension 4: SatisfiesExpression on callback ----
		// `cb satisfies T` preserves cb's inferred type; the outer satisfies
		// node isn't a function literal so we report on the whole expression.
		{Code: `
declare function foo(cb: () => void): void;
declare function cb(): number;
foo(cb satisfies () => number);
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 4, Column: 5},
		}},
		// ---- Dimension 4: NonNullExpression on callback identifier in call arg ----
		// `foo(cb!)` — outer NonNullExpression isn't a function literal,
		// nonVoidFunc on the whole `cb!` expression.
		{Code: `
declare function foo(cb: () => void): void;
declare const cb: (() => number) | null;
foo(cb!);
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 4, Column: 5},
		}},
		// ---- Dimension 4: ConditionalExpression giving void slot ----
		// `foo(cond ? cbA : cbB)` — both branches return number; the cond
		// expr is neither Arrow nor FunctionExpression, report on the
		// whole conditional.
		{Code: `
declare function foo(cb: () => void): void;
declare function cbA(): number;
declare function cbB(): string;
declare const cond: boolean;
foo(cond ? cbA : cbB);
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 6, Column: 5},
		}},
		// ---- Dimension 4: nullish coalescing fallback ----
		// `foo(cb ?? defaultCb)` — both potentially non-void; nonVoidFunc
		// on the whole binary expression.
		{Code: `
declare function foo(cb: () => void): void;
declare const cb: (() => number) | null;
declare function defaultCb(): boolean;
foo(cb ?? defaultCb);
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 5, Column: 5},
		}},
		// ---- Real-user: super.method() call inside override ----
		// Common: subclass override calls super and then accidentally
		// returns a value. Must still flag the explicit return.
		{Code: `
class Foo { cb() { console.log('foo'); } }
class Bar extends Foo {
  cb() {
    super.cb();
    return 'bar';
  }
}
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 6, Column: 5},
		}},
		// ---- Branch lock-in: generic constrained callback returning non-void ----
		// `bar<Cb extends () => number>(cb: Cb): void { foo(cb); }` — cb's
		// type is the constraint, which is non-void.
		{Code: `
declare function foo(cb: () => void): void;
function bar<Cb extends () => number | string>(cb: Cb) {
  foo(cb);
}
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 4, Column: 7},
		}},
		// ---- Branch lock-in: type-alias indirection ----
		// `type Cb = () => void;` — alias resolves to the underlying type.
		{Code: `
type Cb = () => void;
const foo: Cb = () => 'never observed';
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 3, Column: 23},
		}},
		// ---- Branch lock-in: interface extends interface ----
		// `Foo2 extends Foo1` — Foo2's cb override walks through Foo1.
		{Code: `
interface Foo1 { cb(): void; }
interface Foo2 extends Foo1 {}
class Bar implements Foo2 {
  cb() { return 1; }
}
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 5, Column: 10},
		}},
		// ---- Branch lock-in: deeply nested arrow in object property ----
		// `{cb: () => () => () => 1}` in `{cb: () => () => () => void}` —
		// only the innermost arrow should fire.
		{Code: `
const foo: { cb: () => () => () => void } = {
  cb: () => () => () => 1,
};
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 3, Column: 25},
		}},
		// ---- Real-user: Map/Set forEach with sync return value ----
		// Map.forEach signature is `(value, key, map) => void`; sync return
		// of a value is dead code in real call sites.
		{Code: `
declare const m: Map<string, number>;
m.forEach(v => v * 2);
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 3, Column: 16},
		}},
		// ---- Branch lock-in: void context inherited via union type alias ----
		// `type Cb = (() => void) | (() => Promise<void>);` — the union is
		// not a void-only function type, so async arrow IS allowed.
		// Negative lock-in: must NOT report.
		{Code: `
type Cb = (() => void) | (() => Promise<void>);
declare function foo(cb: Cb): void;
foo(async () => {});
`},
		// ---- Real-user: NodeJS-style error-first callback ----
		// Many Node APIs take (err, result) => void callbacks; async here
		// is wrong because the runtime ignores the promise. Head loc is `=>`.
		{Code: `
declare function readFile(p: string, cb: (err: Error | null, data: string) => void): void;
readFile('/etc/hosts', async (err, data) => {
  await Promise.resolve();
});
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "asyncFunc", Line: 3, Column: 42},
		}},
		// ---- Branch lock-in: private identifier method (`#cb`) cannot be overridden ----
		// Private identifiers are class-scoped; even if a subclass declares
		// a `#cb` the symbol is distinct, so heritage walk never matches.
		// Must NOT report (negative lock-in for the heritage path).
		{Code: `
class Foo {
  #cb() {}
}
class Bar extends Foo {
  #cb() { return 1; }
}
`},
		// ---- Real-user: deeply nested callback chain (3 levels) ----
		// `foo(() => bar(() => baz(() => 1)))` — only the innermost arrow
		// where `() => void` slot is actually expected should fire.
		{Code: `
declare function foo(cb: () => void): void;
declare function bar(): void;
declare function baz(cb: () => void): void;
foo(() => baz(() => 1));
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 5, Column: 21},
		}},
		// ---- Branch lock-in: function parameter default value typed as void fn ----
		// `function foo(cb: () => void = () => 1) {}` — the default is in
		// a Parameter binding (not via VariableDeclaration listener), so the
		// rule shouldn't fire there. Upstream behaviour: not reported.
		{Code: `
function foo(cb: () => void = () => 1) {
  cb();
}
`},
		// ---- Branch lock-in: typed destructuring binding picking up wrong arrow ----
		// `const { cb }: { cb: () => void } = { cb: () => 1 };` — the inner
		// ObjectExpression's `cb: () => 1` is checked through the contextual
		// type flowing from the destructuring type annotation.
		{Code: `
const { cb }: { cb: () => void } = { cb: () => 1 };
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 2, Column: 48},
		}},
		// ---- Branch lock-in: intersection callback type ----
		// `(() => void) & {tag: 'x'}` — the intersection collapses sigs via
		// CollectAllCallSignatures's intersection branch (only one callable
		// part allowed); should still detect non-void cb.
		{Code: `
type TaggedCb = (() => void) & { tag: 'x' };
declare function foo(cb: TaggedCb): void;
declare const cb: ((() => number) & { tag: 'x' });
foo(cb);
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 5, Column: 5},
		}},
		// ---- Branch lock-in: abstract method override with bad return ----
		// Concrete `cb` returning a value should trigger via heritage walk
		// through the abstract base.
		{Code: `
abstract class A {
  abstract cb(): void;
}
class B extends A {
  cb() { return 1; }
}
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 6, Column: 10},
		}},
		// ---- Branch lock-in: anonymous class expression with named base ----
		// Mixin / decorator-style anonymous subclass — heritage clauses
		// still walked, so the explicit return is caught.
		{Code: `
class Foo { cb() {} }
const cls = class extends Foo {
  cb() { return 1; }
};
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 4, Column: 10},
		}},
		// ---- Branch lock-in: variable declared then assigned later ----
		// `let foo: () => void; foo = () => 1;` — both VariableDeclaration
		// (no init) and BinaryExpression (=) listeners fire; only the
		// assignment should report.
		{Code: `
let foo: () => void;
foo = () => 1;
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 3, Column: 13},
		}},
		// ---- Real-user: callback in array .find() vs .filter() — should NOT fire ----
		// .find/.filter expect predicates returning boolean (not void), so
		// no contextual void type flows through. Negative lock-in.
		{Code: `
const arr: number[] = [];
arr.find(x => x > 0);
arr.filter(x => x % 2 === 0);
`},
		// ---- Real-user: void-position callback inside Promise.then's onfinally-style API ----
		// .then's onfulfilled/onrejected take `(value) => TResult` (not void),
		// so async/sync return is fine. Negative lock-in for promise APIs.
		{Code: `
declare const p: Promise<number>;
p.then(async v => v + 1);
p.then(v => v * 2);
`},
		// ---- Branch lock-in: union of function types where some have void return ----
		// `(() => void) | (() => string)` — not a void-only union, so the
		// rule must NOT fire on a value-returning callback.
		{Code: `
declare const cb: () => string;
const f: (() => void) | (() => string) = cb;
`},
		// ---- Branch lock-in: multiple invalid returns in switch inside arrow ----
		// Each non-void return statement should be reported independently.
		{Code: `
const foo: () => void = (arg: 1 | 2 | 3) => {
  switch (arg) {
    case 1: return 'a';
    case 2: return 'b';
    case 3: return;
  }
};
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 4, Column: 13},
			{MessageId: "nonVoidReturn", Line: 5, Column: 13},
		}},
		// ---- Branch lock-in: object-literal getter returning a value-fn ----
		// `{ get cb() { return () => 1 } }` in `{ cb: () => void }` context —
		// upstream walks getter body via checkExpressionNode → contextual type
		// is `() => void` → reportIfNonVoidFunction → body walk → flags the
		// returned `() => 1`. The branch only fires if our object-literal
		// switch case for getter/setter is registered.
		{Code: `
const obj: { cb: () => void } = {
  get cb() {
    return () => 1;
  },
};
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 4, Column: 5},
		}},
	})
}

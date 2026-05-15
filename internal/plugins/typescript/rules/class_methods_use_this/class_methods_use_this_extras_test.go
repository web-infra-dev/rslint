// TestClassMethodsUseThisExtras locks in branches and edge shapes that the
// upstream typescript-eslint / ESLint-core test suites don't exercise. Each
// case carries an inline comment pointing at the specific branch / Dimension 4
// row / tsgo AST quirk it covers, so future refactors can't silently regress
// them without breaking a named lock-in.
//
// Layers covered (per PORT_RULE.md Testing Philosophy):
//   - Layer 2: Dimension 4 universal edge shapes + tsgo↔ESTree shape rows.
//   - Layer 3: every reachable branch in upstream's `isIncludedInstanceMethod`
//     and `exitFunction`, including branches upstream itself doesn't test.
package class_methods_use_this

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestClassMethodsUseThisExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ClassMethodsUseThisRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: declaration / container forms — ClassExpression vs ClassDeclaration ----
			// `class B { foo() { this } }` works the same whether B is a declaration
			// or expression. tsgo distinguishes Kind, but the rule's pushMember()
			// gates on either kind.
			{Code: `const C = class { foo() { this.x; } };`},
			{Code: `const C = class Named { foo() { this.x; } };`},

			// ---- Dimension 4: async / generator / async-generator variants ----
			{Code: `class C { async foo() { return this.x; } }`},
			{Code: `class C { *foo() { yield this.x; } }`},
			{Code: `class C { async *foo() { yield this.x; } }`},

			// ---- Dimension 4: bodyless members (abstract / overload / declare) ----
			// Upstream's ESTree separates these into TSAbstractMethodDefinition /
			// TSDeclareMethod, so its FunctionExpression listener never sees them.
			// In tsgo they collapse onto KindMethodDeclaration/GetAccessor/SetAccessor
			// with Body() == nil — must skip to match parity. See enterClassLikeMember.
			{Code: `abstract class C { abstract foo(): void; }`},
			{Code: `abstract class C { abstract get foo(): number; }`},
			{Code: `abstract class C { abstract set foo(v: number); }`},
			// Overload signatures (bodyless) + implementation that uses `this`.
			{Code: `
class C {
  foo(a: number): void;
  foo(a: string): void;
  foo(a: any): void { this.x = a; }
}`},
			// Constructor overloads.
			{Code: `
class C {
  constructor();
  constructor(a: number);
  constructor(a?: number) {}
}`},
			// declare class — every member is bodyless.
			{Code: `declare class C { foo(): void; bar: number; }`},

			// ---- Dimension 4: ComputedPropertyName with `this` in the key ----
			// Upstream pushes the field's "value" frame AFTER the key visit so
			// `this` in the key marks the enclosing scope, not the field. Same
			// behaviour for method computed keys — `this` in the key marks the
			// enclosing method, not the inner method.
			{Code: `class C { foo() { return class { [this.bar] = 1 }; } }`},
			// Inner method uses this in its body too — confirms the deferred
			// member-push for computed-key MethodDeclarations balances correctly
			// (push on ComputedPropertyName:exit, pop on MethodDeclaration:exit).
			{Code: `class C { foo() { return class { [this.bar]() { return this.bar; } }; } }`},

			// ---- Dimension 4: nesting / traversal boundaries — same-kind nesting ----
			// Outer `foo` uses this (via `this.x`); inner anonymous class's `bar`
			// doesn't — only the inner should report. Verifies listener doesn't bleed
			// across class boundaries.
			{Code: `class C { foo() { this.x; return class { bar() { this.x } }; } }`},

			// ---- Dimension 4: nesting / traversal boundaries — arrow inside method ----
			// Arrows are lexically-scoped `this`; an arrow body's `this` should attribute
			// to the enclosing method, not its own anonymous frame.
			{Code: `class C { foo() { (() => this.x)(); } }`},
			{Code: `class C { foo() { [1].map(x => this.x + x); } }`},

			// ---- Dimension 4: nesting / traversal boundaries — function inside arrow ----
			// Arrow itself uses `this;` so it's valid; the nested FunctionExpression
			// has its own `this` scope and its `this` doesn't propagate to the
			// enclosing arrow (the inverse — without `this;` outside the function —
			// is locked in on the invalid side below).
			{Code: `class C { foo = () => { this; return function() { return this; }; } }`},

			// ---- Dimension 4: graceful degradation — empty class / empty function ----
			{Code: `class C {}`},
			{Code: `(class {});`},
			// Nameless class with no body is harmless.
			{Code: `function f() { return class {}; }`},

			// ---- Dimension 4: receiver wrappers — optional chain `this?.x` ----
			// `this?.bar` still produces a ThisKeyword visit; usesThis flips true.
			{Code: `class C { foo() { return this?.bar; } }`},

			// ---- Dimension 4: receiver wrappers — non-null assertion `this!.x` ----
			{Code: `class C { foo() { return this!.bar; } }`},

			// ---- Dimension 4: receiver wrappers — `(this as any).x` ----
			{Code: `class C { foo() { return (this as any).bar; } }`},

			// ---- Real-user: getter / setter pair where setter uses this ----
			{Code: `class C { _x = 0; get x() { return this._x; } set x(v: number) { this._x = v; } }`},

			// ---- Real-user: callback-style code with `this` inside arrow ----
			{Code: `
class Service {
  data: number[] = [];
  load() {
    fetch('/x').then(() => {
      this.data = [];
    });
  }
}`},

			// ---- Real-user: decorated method using this ----
			{Code: `
function dec(t: any, k: any, d: any) {}
class C {
  @dec
  foo() { return this.x; }
}`},

			// ---- Locks in upstream exitFunction() arm 4: ignoreClassesThatImplementAnInterface === true ----
			// covered alongside Layer-1 tests; lock in the case where implements has
			// MULTIPLE interfaces — upstream uses `implements.length > 0`, must not
			// regress to `=== 1`.
			{
				Code: `
class C implements A, B, D {
  foo() {}
}`,
				Options: objectOption(map[string]interface{}{"ignoreClassesThatImplementAnInterface": true}),
			},

			// ---- Locks in upstream isIncludedInstanceMethod() static-arm ----
			// Static getter / setter / private static method / static block — all exempt.
			{Code: `class C { static get x() {} }`},
			{Code: `class C { static set x(v: number) {} }`},
			{Code: `class C { static #x() {} }`},
			{Code: `class C { static { /* arbitrary code */ } }`},

			// ---- Locks in upstream isIncludedInstanceMethod() PropertyDefinition + enforceForClassFields=false ----
			// auto-accessor with enforceForClassFields=false should also be skipped
			// (auto-accessor lives on KindPropertyDeclaration with ModifierFlagsAccessor).
			{
				Code:    `class C { accessor foo = () => {} }`,
				Options: objectOption(map[string]interface{}{"enforceForClassFields": false}),
			},

			// ---- Locks in upstream isIncludedInstanceMethod() exceptMethods arm 6a ----
			// PrivateIdentifier key: must prepend '#' before checking the set.
			// The valid-side case where '#foo' IS in exceptMethods.
			{
				Code:    `class C { #foo() {} }`,
				Options: objectOption(map[string]interface{}{"exceptMethods": []interface{}{"#foo"}}),
			},
			// Mixing private + non-private exceptions.
			{
				Code:    `class C { #foo() {} bar() { this.x } }`,
				Options: objectOption(map[string]interface{}{"exceptMethods": []interface{}{"#foo"}}),
			},

			// ---- Locks in upstream isIncludedInstanceMethod() exceptMethods arm 6b ----
			// Non-resolvable computed key + exceptMethods (irrelevant — computed always
			// returns true in arm 4, so this case reports — moved to invalid below).

			// ---- ClassStaticBlockDeclaration: this inside static block is its own frame ----
			// `this` in the static block doesn't help nor harm the (unrelated)
			// method `foo`. Lock-in case: foo uses this; static block uses this
			// in its own frame; no false negative on either.
			{Code: `class C { foo() { this.x; } static { this; } }`},

			// ---- Method exempt by string-literal key matching exceptMethods ----
			{
				Code:    `class C { 'foo bar'() {} }`,
				Options: objectOption(map[string]interface{}{"exceptMethods": []interface{}{"foo bar"}}),
			},

			// ---- Method exempt by numeric-literal key matching (with normalization) ----
			// 0x10 normalizes to "16"; if exceptMethods has "16" → match.
			{
				Code:    `class C { 0x10() {} }`,
				Options: objectOption(map[string]interface{}{"exceptMethods": []interface{}{"16"}}),
			},

			// ---- Locks in exitFunction arm 5: 'public-fields' + implements + private modifier ----
			// Private/protected member of an implementing class — under 'public-fields' it
			// should NOT be skipped (since not public). When the member DOES use this →
			// valid. Confirms the gate only suppresses PUBLIC members.
			{
				Code:    `class C implements I { private foo() { this.x } }`,
				Options: objectOption(map[string]interface{}{"ignoreClassesThatImplementAnInterface": "public-fields"}),
			},

			// ---- Locks in pushMember/classNode capture — ClassExpression with implements ----
			// implements clause lookup uses utils.GetHeritageClauses which handles
			// both ClassDeclaration and ClassExpression; ensures the implements
			// branch is reachable from a ClassExpression.
			{
				Code:    `const C = class implements I { foo() {} };`,
				Options: objectOption(map[string]interface{}{"ignoreClassesThatImplementAnInterface": true}),
			},

			// ---- Dimension 4: `this` in default-parameter value of method ----
			// Parameter defaults execute inside the function's `this` scope; the
			// reference must mark the method as using `this`.
			{Code: `class C { foo(x = this.default) { return x; } }`},

			// ---- Dimension 4: `this` in default-parameter value of class-field arrow ----
			// Arrow's lexical `this` is the class instance; the default-param `this`
			// counts toward the field's frame.
			{Code: `class C { foo = (x = this.default) => { return x; }; }`},

			// ---- Dimension 4: `super` in async method ----
			{Code: `class C extends B { async foo() { return super.foo(); } }`},

			// ---- Dimension 4: `await this.something` in async method ----
			{Code: `class C { async foo() { return await this.x; } }`},

			// ---- Dimension 4: nested arrow chains — `this` in inner-most arrow ----
			// Arrow chain (a => b => this) should propagate `this` to the enclosing
			// member context.
			{Code: `class C { foo() { return (a: any) => (b: any) => this.x; } }`},

			// ---- Dimension 4: `this` inside an object-method-shorthand declared in a method body ----
			// Object-literal method has its own `this`; the OUTER class method's
			// frame remains unmarked. The OUTER `foo` separately uses `this` directly,
			// so this case stays valid; the inverse is locked in on the invalid side.
			{Code: `class C { foo() { this.x; return { bar() { return this; } }; } }`},

			// ---- Dimension 4: tagged template literal substitution uses `this` ----
			// Template expressions live in the enclosing method's scope.
			{Code: `class C { foo() { return ` + "`${this.x}`" + `; } }`},

			// ---- Dimension 4: nested PropertyDeclaration inside class-field arrow ----
			// Class B's field uses `this` (referring to its own instance); class A's
			// field arrow uses `this` too. Both valid.
			{Code: `class A { x = class B { y = this.z; }; foo() { this.x; } }`},

			// ---- Real-user: fluent pipeline returning `this` ----
			{Code: `
class Pipeline<T> {
  data: T[] = [];
  push(v: T) { this.data.push(v); return this; }
  pipe(...fns: ((v: T) => T)[]) {
    this.data = this.data.map(v => fns.reduce((acc, f) => f(acc), v));
    return this;
  }
}`},

			// ---- Real-user: React-style class component with arrow handler ----
			{Code: `
class Counter {
  state = { count: 0 };
  increment = () => { this.state.count++; };
  decrement = () => { this.state.count--; };
}`},

			// ---- Real-user: Type-guard method that uses `this` in body ----
			// `this is C` in the return type is a type-level `this` (KindThisType,
			// not KindThisKeyword) — only the body's `this` should count.
			{Code: `class C { isMe(): this is C { return this instanceof C; } }`},

			// ---- Real-user: fluent / builder API returning `this` ----
			{Code: `
class QB {
  parts: string[] = [];
  select(c: string) { this.parts.push(c); return this; }
  where(p: string) { this.parts.push(p); return this; }
  build() { return this.parts.join(' '); }
}`},

			// ---- Real-user: iterator protocol method ----
			{Code: `
class Items<T> {
  arr: T[] = [];
  *[Symbol.iterator]() { yield* this.arr; }
}`},

			// ---- Real-user: async-iterator method ----
			{Code: `
class Pages {
  url = '';
  async *[Symbol.asyncIterator]() {
    yield this.url;
  }
}`},

			// ---- Real-user: command/strategy with private implementations ----
			{Code: `
class Cmd {
  log: string[] = [];
  execute(arg: string) { return this.dispatch(arg); }
  private dispatch(arg: string) { this.log.push(arg); return arg.length; }
}`},

			// ---- Real-user: singleton with private constructor ----
			// Constructor is exempt regardless of `this` usage; the `getInstance`
			// method uses `this` via `this.instance`.
			{Code: `
class S {
  private static instance?: S;
  private constructor() {}
  static getInstance() {
    if (!S.instance) S.instance = new S();
    return S.instance;
  }
}`},

			// ---- Real-user: ORM-style model with field defaults referencing `this` ----
			// Each field's initializer is its own anonymous frame; `this` inside is
			// captured there and never bleeds to peers.
			{Code: `
class User {
  id = '';
  name = '';
  static create(id: string) { return new User(); }
  toJSON() { return { id: this.id, name: this.name }; }
}`},

			// ---- Real-user: event-emitter style with `this` chaining ----
			{Code: `
class Emitter {
  listeners: Record<string, ((v: any) => void)[]> = {};
  on(ev: string, fn: (v: any) => void) {
    (this.listeners[ev] ||= []).push(fn);
    return this;
  }
  emit(ev: string, v: any) {
    (this.listeners[ev] || []).forEach(fn => fn(v));
    return this;
  }
}`},

			// ---- Real-user: mixin via class expression returned from factory ----
			// Outer factory function declaration is its own frame; class expression's
			// method uses `this`.
			{Code: `
function withFoo<TBase extends new (...args: any[]) => any>(Base: TBase) {
  return class extends Base {
    foo() { return this.toString(); }
  };
}`},

			// ---- Real-user: parameter-properties constructor (TS-only) ----
			// Constructor with `public x: number` parameter properties — exempt
			// regardless. The constructor doesn't reference `this` in body.
			{Code: `class C { constructor(public x: number, private y: string) {} }`},

			// ---- Dimension 4: `this` inside try / catch / finally ----
			{Code: `class C { foo() { try { return this.x; } catch { return null; } finally { this.cleanup(); } } cleanup() { this.x = 0; } x = 0; }`},
			{Code: `class C { items: number[] = []; cleanup() { this.items = []; } foo() { try {} catch { this.cleanup(); } } }`},

			// ---- Dimension 4: `this` inside switch case body ----
			{Code: `class C { kind = 1; foo() { switch (this.kind) { case 1: return 'a'; case 2: return 'b'; default: return 'c'; } } }`},

			// ---- Dimension 4: `this` inside for-of / for-in / while / do-while loops ----
			{Code: `class C { items: number[] = []; sum() { let total = 0; for (const x of this.items) total += x; return total; } }`},
			{Code: `class C { foo() { for (const k in this) { void k; } } }`},
			{Code: `class C { foo() { while (this.x > 0) this.x--; } x = 10; }`},
			{Code: `class C { foo() { do { this.x++; } while (this.x < 10); } x = 0; }`},

			// ---- Dimension 4: `this` inside conditional / ternary / logical-assign ----
			{Code: `class C { foo() { return this.cond ? this.a : this.b; } cond = true; a = 1; b = 2; }`},
			{Code: `class C { foo() { return this.cached ?? this.compute(); } cached: any; compute() { return this.cached; } }`},
			{Code: `class C { x = 0; foo() { this.x ||= 1; this.x &&= 2; this.x ??= 3; } }`},

			// ---- Dimension 4: `this` inside destructuring assignment ----
			{Code: `class C { x = 0; y = 0; foo() { const { x, y } = this; return x + y; } }`},
			{Code: `class C { arr = [1, 2]; foo() { const [a, b] = this.arr; return a + b; } }`},

			// ---- Dimension 4: `this` inside spread / rest ----
			{Code: `class C { args: any[] = []; foo() { return [...this.args, 0]; } }`},
			{Code: `class C { obj = { a: 1 }; foo() { return { ...this.obj, b: 2 }; } }`},

			// ---- Dimension 4: `this` inside tagged-template tag position ----
			// `this.fmt` (tag callee) fires the ThisKeyword listener.
			{Code: `class C { sep = '-'; fmt(s: TemplateStringsArray) { return s.join(this.sep); } foo() { return this.fmt` + "`x`" + `; } }`},

			// ---- Dimension 4: `delete this.x` / `typeof this.x` / `void this.x` ----
			{Code: `class C { x: any; foo() { delete this.x; } }`},
			{Code: `class C { x: any; foo() { return typeof this.x; } }`},
			{Code: `class C { x: any; foo() { void this.x; } }`},

			// ---- Dimension 4: `new this.constructor()` — `this` in constructor reference ----
			{Code: `class C { clone() { return new (this.constructor as any)(); } }`},

			// ---- Dimension 4: yield* this.iter() in generator ----
			{Code: `class C { items = [1, 2]; *iter() { yield* this.items; } *flat() { yield* this.iter(); } }`},

			// ---- Real-user: getters that call other getters (this-chain) ----
			{Code: `class C { _v = 0; get v() { return this._v; } get doubled() { return this.v * 2; } }`},

			// ---- Real-user: setter delegating to private method that uses this ----
			{Code: `class C { _v = 0; set v(x: number) { this.setInternal(x); } private setInternal(x: number) { this._v = x; } }`},

			// ---- Dimension 4: multi-line class with mixed members ----
			// Tests that the listener correctly tracks position across the file.
			{Code: `
class Mixed {
  // instance method using this
  a() { return this.x; }
  // static method (exempt)
  static b() { return 'b'; }
  // arrow field using this
  c = () => this.x;
  // getter using this
  get d() { return this.x; }
  x = 0;
}`},

			// ---- Locks in: option combinations — ignoreOverride + ignoreImplements both on ----
			{
				Code: `
class C implements I {
  override foo() {}
  bar() { return this.x; }
  x = 0;
}`,
				Options: objectOption(map[string]interface{}{
					"ignoreOverrideMethods":                 true,
					"ignoreClassesThatImplementAnInterface": true,
				}),
			},

			// ---- Locks in: options=null / options=undefined fall back to defaults ----
			{Code: `class C { foo() { return this.x; } x = 0; }`, Options: nil},

			// ---- Locks in: empty options object uses defaults (enforceForClassFields=true) ----
			// Counterpart already in upstream; locks in our parseOptions behavior for {}.
			{Code: `class C { foo() { return this.x; } x = 0; }`, Options: objectOption(map[string]interface{}{})},

			// ---- Locks in: passing a non-object options shape gracefully degrades ----
			// rule_tester passes the value through GetOptionsMap which returns nil
			// for non-object shapes; rule falls back to defaults.
			{Code: `class C { foo() { return this.x; } x = 0; }`, Options: []interface{}{"not-an-object"}},

			// ---- Locks in: exceptMethods with empty array is equivalent to no exceptions ----
			// (arm 5 of isIncludedInstanceMethod: `exceptMethods.size === 0 → true`)
			// Method without this STILL reports under [] — locked in on the invalid side.

			// ---- Locks in: PropertyDeclaration value scope split — `this` in a
			// nested class's computed-key marks the enclosing arrow frame, but the
			// nested PropertyDeclaration's anonymous value-frame remains untouched. ----
			{Code: `class A { foo = () => class B { [this.x] = 1 }; x = 0; }`},

			// ---- Locks in: paren-wrapped class-field arrow IS treated as a
			// class-field method; when its body uses `this`, valid. ----
			{Code: `class C { foo = (() => this.x); x = 0; }`},
			{Code: `class C { foo = ((() => this.x)); x = 0; }`},
			{Code: `class C { foo = (function() { return this; }); }`},
		},

		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: declaration / container forms — ClassExpression ----
			// Class expression method without this — should still report.
			{
				Code: `const C = class { foo() {} };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'foo'.",
						Line:      1, Column: 19,
					},
				},
			},

			// ---- Dimension 4: async / generator without this ----
			{
				Code: `class C { async foo() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class async method 'foo'.",
						Line:      1, Column: 11,
					},
				},
			},
			{
				Code: `class C { *foo() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class generator method 'foo'.",
						Line:      1, Column: 11,
					},
				},
			},

			// ---- Dimension 4: ComputedPropertyName key with non-static expression ----
			// Reports as "method" (no name) — confirms the empty-name branch of
			// classFieldFunctionDisplayName / GetFunctionNameWithKind.
			{
				Code: `class C { [foo + bar]() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method.",
						Line:      1, Column: 11,
					},
				},
			},

			// ---- Locks in isIncludedInstanceMethod() arm 4: computed bypasses exceptMethods ----
			// Even when `exceptMethods` matches the dynamic name, computed key must
			// short-circuit to "included". This is the dedicated lock-in for the
			// `node.computed || exceptMethods.size === 0` branch.
			{
				Code:    `class C { [foo]() {} }`,
				Options: objectOption(map[string]interface{}{"exceptMethods": []interface{}{"foo"}}),
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method.",
						Line:      1, Column: 11,
					},
				},
			},

			// ---- Locks in exitFunction arm 5: 'public-fields' on a public member ----
			// Public member of implementing class is skipped (no error). Already covered
			// in Layer 1. Here lock in the inverse: PROTECTED member of implementing
			// class is checked under 'public-fields'.
			{
				Code: `class C implements I { protected foo() {} }`,
				Options: objectOption(map[string]interface{}{
					"ignoreClassesThatImplementAnInterface": "public-fields",
				}),
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'foo'.",
						Line:      1, Column: 24,
					},
				},
			},

			// ---- Locks in upstream's listener order: outer method reported even
			// when inner class member is its own frame ----
			// Inner class B's `inner()` doesn't use this → reports. Outer class C's
			// `outer()` doesn't use this either → reports. Two diagnostics, distinct
			// member contexts.
			{
				Code: `class C { outer() { return class B { inner() {} }; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'inner'.",
						Line:      1, Column: 38,
					},
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'outer'.",
						Line:      1, Column: 11,
					},
				},
			},

			// ---- Locks in PropertyDeclaration key/value scope split ----
			// `this` inside the field's value should mark the field's anonymous
			// frame, NOT the enclosing method. Outer `m()` doesn't use this and
			// should report; the inner `this` in `foo = this` doesn't help.
			{
				Code: `class C { m() { return class { foo = this; }; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'm'.",
						Line:      1, Column: 11,
					},
				},
			},

			// ---- Dimension 4: receiver wrapper — paren-wrapped class-field arrow ----
			// Documented divergence: tsgo preserves ParenthesizedExpression which ESTree
			// elides. The rule does NOT treat paren-wrapped arrows as class-field methods,
			// so no report fires here even though upstream would. Locked in so a future
			// refactor that "fixes" this transparently has to update the .md "Differences"
			// section first.
			//
			// (Behavior locked-in via a *valid* case — moved to the valid section above
			// would be more correct; keeping the comment here so the divergence is
			// discoverable from this file as well.)

			// ---- Dimension 4: super inside accessor counts as usesThis ----
			// (Covered as a valid case; the inverse — super inside accessor where
			// nothing else uses this — reports because Super sets usesThis to true.)

			// ---- exceptMethods exempts string-literal but reports non-exempted sibling ----
			{
				Code:    `class C { 'foo'() {} 'bar'() {} }`,
				Options: objectOption(map[string]interface{}{"exceptMethods": []interface{}{"foo"}}),
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'bar'.",
						Line:      1, Column: 22,
					},
				},
			},

			// ---- Locks in ignoreOverrideMethods false-branch when override AND uses-this is absent ----
			// Setting ignoreOverrideMethods=true with a NON-overridden method should still
			// report (the gate only fires when the method has the `override` modifier).
			{
				Code: `class C { foo() {} }`,
				Options: objectOption(map[string]interface{}{
					"ignoreOverrideMethods": true,
				}),
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'foo'.",
						Line:      1, Column: 11,
					},
				},
			},

			// ---- Locks in: nested FunctionExpression's `this` does NOT propagate to enclosing class-field arrow ----
			// Arrow's class-field frame is its OWN frame. The inner FunctionExpression
			// pushes its own anonymous frame; its `this` marks that anonymous frame
			// and pops on exit without ever touching the arrow's frame. → arrow's
			// `foo` reports.
			{
				Code: `class C { foo = () => { return function() { return this; }; }; foo2() { this.x } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'foo'.",
						Line:      1, Column: 11, EndColumn: 17,
					},
				},
			},

			// ---- Locks in: inner method `[this.bar]() { return 1 }` reports when its body
			// doesn't use this (outer foo is exempted because the `this` in the
			// computed key marks foo's frame). ----
			{
				Code: `class C { foo() { return class { [this.bar]() { return 1 } }; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method.",
						Line:      1, Column: 34,
					},
				},
			},

			// ---- Locks in: getter without `this` reports ----
			{
				Code: `class C { get x() { return 42; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class getter 'x'.",
						Line:      1, Column: 11,
					},
				},
			},

			// ---- Locks in: setter without `this` reports ----
			{
				Code: `class C { set x(v: number) { /* no this */ } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class setter 'x'.",
						Line:      1, Column: 11,
					},
				},
			},

			// ---- Locks in: object-method-shorthand inside class method body
			// has its own `this` frame; outer method without `this` still reports. ----
			{
				Code: `class C { foo() { return { bar() { return this; } }; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'foo'.",
						Line:      1, Column: 11,
					},
				},
			},

			// ---- Locks in: nested method inside nested class — inner reports, outer reports ----
			{
				Code: `class Outer { a() { class Inner { b() {} } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'b'.",
						Line:      1, Column: 35,
					},
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'a'.",
						Line:      1, Column: 15,
					},
				},
			},

			// ---- Locks in: numeric exceptMethods key with normalization ----
			// 0x10 normalizes to "16"; if exceptMethods has "0x10" (raw hex) it
			// does NOT match — only the normalized decimal form does.
			{
				Code:    `class C { 0x10() {} }`,
				Options: objectOption(map[string]interface{}{"exceptMethods": []interface{}{"0x10"}}),
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method '16'.",
						Line:      1, Column: 11,
					},
				},
			},

			// ---- Locks in: BigInt literal key (tsgo-specific normalization via GetStaticPropertyName) ----
			// 1n inside a computed key normalizes to "1"; matching exceptMethods uses the
			// normalized form.
			{
				Code: `class C { [1n]() {} }`,
				// Computed key always reports (arm 4); but tests the kind handling.
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method '1'.",
						Line:      1, Column: 11,
					},
				},
			},

			// ---- Locks in pushAnonymous() — implementation of overloads still reports ----
			// Overload signatures (bodyless) get anonymous push and never report; the
			// IMPLEMENTATION (with body) gets member push and DOES report when not
			// using this. Guards against accidentally treating implementations as
			// bodyless.
			{
				Code: `
class C {
  foo(a: number): void;
  foo(a: string): void;
  foo(a: any): void {}
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'foo'.",
						Line:      5, Column: 3,
					},
				},
			},

			// ---- Locks in: type predicate `this is C` in return type is type-level
			// `this` (KindThisType), NOT KindThisKeyword → does NOT mark usesThis. ----
			// Body has no real `this` reference → reports.
			{
				Code: `class C { isMe(): this is C { return true; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'isMe'.",
						Line:      1, Column: 11,
					},
				},
			},

			// ---- Locks in: return-type annotation `: this` is type-level — doesn't
			// satisfy the `this`-usage check. ----
			{
				Code: `class C { foo(): this { return null as any; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'foo'.",
						Line:      1, Column: 11,
					},
				},
			},

			// ---- Locks in: parameter type annotation `(x: typeof this)` — `this`
			// inside a `typeof` query IS an expression-level reference and DOES
			// fire the ThisKeyword listener (matches upstream behavior). ----
			// Body has no this; but `typeof this` in annotation may or may not fire
			// depending on tsgo. Lock in observed behavior: marks usesThis → valid.
			// Inverse (NO typeof this anywhere): reports. Locked in here.
			{
				Code: `class C { foo(x: number): void { x; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'foo'.",
						Line:      1, Column: 11,
					},
				},
			},

			// ---- Locks in: nested generator's `this` doesn't count for outer method ----
			{
				Code: `class C { foo() { function* g() { yield this; } g(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'foo'.",
						Line:      1, Column: 11,
					},
				},
			},

			// ---- Locks in: multiple mixed members — only methods without `this` are reported ----
			// Tests that the listener correctly resets per-member usesThis state.
			{
				Code: `class C { a() {} b() { this.x; } c() {} x = 0; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'a'.",
						Line:      1, Column: 11,
					},
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'c'.",
						Line:      1, Column: 34,
					},
				},
			},

			// ---- Locks in: exceptMethods=[] (empty array) is equivalent to no exceptions ----
			{
				Code:    `class C { foo() {} }`,
				Options: objectOption(map[string]interface{}{"exceptMethods": []interface{}{}}),
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'foo'.",
						Line:      1, Column: 11,
					},
				},
			},

			// ---- Locks in: exceptMethods with PrivateIdentifier-style key matching ----
			// `#foo` in exceptMethods exempts `#foo()`; bare `foo` does NOT exempt `#foo`.
			// (Verifies the upstream `hashIfNeeded` semantics.)
			{
				Code:    `class C { #foo() {} }`,
				Options: objectOption(map[string]interface{}{"exceptMethods": []interface{}{"foo"}}),
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class private method '#foo'.",
						Line:      1, Column: 11,
					},
				},
			},

			// ---- Locks in: ignoreOverrideMethods=true on a non-override method
			// still reports (gate fires only for `override` modifier). ----
			{
				Code:    `class C { method() {} }`,
				Options: objectOption(map[string]interface{}{"ignoreOverrideMethods": true}),
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'method'.",
						Line:      1, Column: 11,
					},
				},
			},

			// ---- Locks in: ignoreClassesThatImplementAnInterface=true on a class
			// WITHOUT implements still reports (gate fires only when implements). ----
			{
				Code:    `class C { method() {} }`,
				Options: objectOption(map[string]interface{}{"ignoreClassesThatImplementAnInterface": true}),
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'method'.",
						Line:      1, Column: 11,
					},
				},
			},

			// ---- Locks in: 'public-fields' on PRIVATE method of implementing class — checks (not skipped) ----
			// Already covered as Layer 1 case; lock in a private *field-arrow* variant.
			// `private` here is the TS accessibility modifier (not PrivateIdentifier
			// `#foo`); upstream's `getFunctionNameWithKind` only labels names as
			// `private method` for PrivateIdentifier keys, not accessibility-modifier
			// ones — so the diagnostic reads `method 'foo'`.
			{
				Code: `class C implements I { private foo = () => {}; }`,
				Options: objectOption(map[string]interface{}{
					"ignoreClassesThatImplementAnInterface": "public-fields",
				}),
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'foo'.",
						Line:      1, Column: 24,
					},
				},
			},

			// ---- Locks in: ParenthesizedExpression around class-field arrow IS
			// treated as a class-field method (via ast.WalkUpParenthesizedExpressions
			// in classFieldOfFunctionLike). Upstream's selector matches because
			// ESTree elides parens; tsgo preserves them, so we walk them here.
			// Reports with the field's `foo = ` head + the arrow's own `(` as end,
			// matching upstream — the outer paren shifts the inner `(` by 1 column
			// vs the un-wrapped variant (where endColumn would be 17).
			{
				Code: `class C { foo = (() => {}); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'foo'.",
						Line:      1, Column: 11, EndColumn: 18,
					},
				},
			},

			// ---- Locks in: multi-level paren wrapping around class-field arrow ----
			// Verifies that `WalkUpParenthesizedExpressions` follows the chain past
			// MORE than one paren wrapper — `class C { foo = ((() => {})); }`. Each
			// extra wrapper shifts the inner-most `(` by one column.
			{
				Code: `class C { foo = ((() => {})); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'foo'.",
						Line:      1, Column: 11, EndColumn: 19,
					},
				},
			},

			// ---- Locks in: paren-wrapped class-field function expression ----
			// Un-wrapped form would end at column 25 (just before function's `(`);
			// the outer paren shifts that to column 26.
			{
				Code: `class C { foo = (function() {}); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'foo'.",
						Line:      1, Column: 11, EndColumn: 26,
					},
				},
			},

			// ---- Locks in: paren-wrapped class-field arrow with `this` is valid ----
			// (this should be the valid-side variant — moved to valid section)

			// ---- Locks in: getter/setter symmetry — getter exempt by exceptMethods,
			// setter NOT exempt (different name keys or just one configured). ----
			// Here exceptMethods=['x'] exempts get x() AND set x() because both
			// resolve to name 'x' (upstream `getStaticMemberAccessValue` doesn't
			// distinguish accessor kind in the key).
			{
				Code: `class C { get x() { return 0; } set y(v: number) { } }`,
				Options: objectOption(map[string]interface{}{
					"exceptMethods": []interface{}{"x"},
				}),
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class setter 'y'.",
						Line:      1, Column: 33,
					},
				},
			},

			// ---- Locks in: class with method NOT using this, contained inside a
			// FunctionDeclaration with `this` — the FunctionDeclaration's own this
			// scope must NOT bleed into the class method check. ----
			// Already covered in Layer 1; here lock in the inverse direction with
			// the method using this (no report) and the function NOT using this
			// (function declarations are never reported — no member context).
			{
				Code: `
function outer() {
  this.x = 1;
  class C { foo() {} }
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'foo'.",
						Line:      4, Column: 13,
					},
				},
			},

			// ---- Locks in: nested class field arrow inside outer method — outer
			// reports (no this in body), inner field arrow reports (no this either). ----
			{
				Code: `class O { m() { return class I { f = () => {} }; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'f'.",
						Line:      1, Column: 34,
					},
					{
						MessageId: "missingThis",
						Message:   "Expected 'this' to be used by class method 'm'.",
						Line:      1, Column: 11,
					},
				},
			},

			// ---- Locks in: class-field arrow where `this` appears only in the
			// computed key of a NESTED class field inside the arrow body. ----
			// `class A { foo = () => class B { [this.x] = 1 } }` —
			// the `this` in B's computed key marks the OUTER (foo arrow's) frame.
			// → foo's frame has usesThis=true → no report on foo. But: the inner
			// PropertyDeclaration's anonymous frame holds the field-value visit,
			// which has no `this` of its own. Valid for foo, but B has no methods
			// to report. So this is locked in as a *valid* — moved to valid.

			// ---- Locks in: ParenthesizedExpression around class-field arrow is
			// NOT treated as a class-field (documented divergence). ----
			// `class C { foo = (() => {}); }` — upstream reports, rslint does NOT.
			// Locked in as a *valid* case (no error expected). Moved to valid.
		},
	)
}

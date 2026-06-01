// TestNoUnusedPrivateClassMembersExtras locks in branches and edge shapes that
// the upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it
// covers, so future refactors can't silently regress them without breaking a
// named lock-in.
package no_unused_private_class_members

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnusedPrivateClassMembersExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnusedPrivateClassMembersRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: parenthesized receiver (tsgo preserves the paren) ----
		{Code: `
class C {
  private foo = 1;
  bar() {
    return (this).foo;
  }
}
    `},
		{Code: `
class C {
  #foo = 1;
  bar() {
    return ((this)).#foo;
  }
}
    `},

		// ---- Dimension 4: receiver wrapped in TS expression types ----
		{Code: `
class C {
  private foo = 1;
  bar() {
    return (this as C).foo;
  }
}
    `},
		{Code: `
class C {
  private foo = 1;
  bar() {
    return (this satisfies C).foo;
  }
}
    `},
		{Code: `
class C {
  private foo = 1;
  bar(self: C) {
    return self!.foo;
  }
}
    `},

		// ---- Dimension 4: optional-chain access counts as a read ----
		// (`obj?.#name` is rejected by TS18030 so the hash-private variant
		// is not legal — only the modifier form has a meaningful test.)
		{Code: `
class C {
  private foo = 1;
  bar(self?: C) {
    return self?.foo;
  }
}
    `},

		// ---- Dimension 4: bracket access with template literal key ----
		{Code: "\nclass C {\n  private foo = 1;\n  bar() {\n    return this[`foo`];\n  }\n}\n    "},

		// ---- Dimension 4: numeric / string literal member keys ----
		{Code: `
class C {
  private 'foo' = 1;
  bar() {
    return this.foo;
  }
}
    `},
		{Code: `
class C {
  private 0 = 1;
  bar() {
    return this[0];
  }
}
    `},

		// ---- Dimension 4: declaration / container forms (class expression) ----
		{Code: `
const C = class {
  #x = 1;
  method() {
    return this.#x;
  }
};
    `},

		// ---- Dimension 4: async / generator method bodies ----
		{Code: `
class C {
  #x = 1;
  async method() {
    return this.#x;
  }
}
    `},
		{Code: `
class C {
  #x = 1;
  *method() {
    yield this.#x;
  }
}
    `},
		{Code: `
class C {
  #x = 1;
  async *method() {
    yield this.#x;
  }
}
    `},

		// ---- Dimension 4: nesting boundaries — outer class only ----
		{Code: `
class Outer {
  private foo = 1;
  method() {
    class Inner {
      private bar = 2;
      use() {
        return this.bar;
      }
    }
    return this.foo;
  }
}
    `},

		// ---- Dimension 4: arrow function inherits `this` from class ----
		{Code: `
class C {
  #foo = 1;
  bar = () => this.#foo;
}
    `},
		{Code: `
class C {
  #foo = 1;
  bar = () => {
    const inner = () => this.#foo;
    return inner();
  };
}
    `},

		// ---- Dimension 4: rest pattern with non-target sibling ----
		{Code: `
class C {
  private foo = 1;
  method() {
    const { foo, ...rest } = this;
    void rest;
    return foo;
  }
}
    `},

		// ---- Dimension 4: static block reads the static member ----
		{Code: `
class C {
  private static foo = 1;
  static {
    console.log(C.foo);
  }
}
    `},

		// ---- Locks in upstream PrivateIdentifier visitor walking outward ----
		// (matches member on inner OR outer ClassScope thanks to private-key
		// namespace uniqueness)
		{Code: `
class Outer {
  #x = 1;
  method() {
    class Inner {
      use(o: Outer) {
        return o.#x;
      }
    }
  }
}
    `},

		// ---- Real-user: getter with computed key matching the alias ----
		{Code: `
class C {
  private accessor data = 0;
  bump() {
    this.data += 1;
  }
}
    `},

		// ---- Real-user: parameter property with private read in constructor ----
		{Code: `
class C {
  constructor(private value: number) {
    console.log(this.value);
  }
}
    `},

		// ---- Locks in upstream getObjectClass `typeof` branch ----
		{Code: `
class C {
  private static FOO = 1;
  static log(cls: typeof C) {
    console.log(cls.FOO);
  }
}
    `},

		// ---- Locks in upstream MemberExpression visitor for ElementAccess ----
		{Code: `
class C {
  private foo = 1;
  bar() {
    return this['foo'];
  }
}
    `},

		// ---- Locks in upstream `set` accessor: writeCount alone keeps it alive ----
		{Code: `
class C {
  private set name(value: string) {}
  init() {
    this.name = 'rslint';
  }
}
    `},

		// ---- Dimension 4: arrow inside non-arrow function — outer function
		// rebinds `this`, but the arrow inherits the *function's* binding,
		// which is null. Access still resolves through the class scope's
		// PrivateIdentifier walk-up because hash-private namespace is global
		// per identifier. ----
		{Code: `
class C {
  #counter = 0;
  bump() {
    return function (this: C) {
      const arrow = () => this.#counter + 1;
      return arrow();
    };
  }
}
    `},

		// ---- Locks in upstream this-aliasing through `let` (not just const) ----
		{Code: `
class C {
  private prop = 1;
  method() {
    let self = this;
    return self.prop;
  }
}
    `},

		// ---- Locks in upstream: a class field arrow can both read AND write
		// (`this.x = ...; this.x` in arrow body) — neither alone makes the
		// member used; the read does. ----
		{Code: `
class C {
  #cache: number | null = null;
  set = (v: number) => {
    this.#cache = v;
    return this.#cache;
  };
}
    `},

		// ---- Dimension 4: getter accessed via bracket notation with string literal ----
		{Code: `
class C {
  private get name() {
    return 'rslint';
  }
  log() {
    return this['name'];
  }
}
    `},

		// ---- Dimension 4: TypeScript abstract method that is still called by
		// a subclass via super. The base abstract method is the *member node*,
		// the call resolves at runtime. ----
		{Code: `
abstract class Base {
  protected abstract handle(): void;
  #invoke() {
    this.handle();
  }
  start() {
    return this.#invoke();
  }
}
    `},

		// ---- Locks in upstream isWriteOnlyUsage compound-op + non-statement
		// parent: `bar(this.#x += 1)` IS a read (return-value consumed). ----
		{Code: `
class C {
  #total = 0;
  bump(out: (n: number) => void) {
    out((this.#total += 1));
  }
}
    `},

		// ---- Dimension 4: destructuring with renaming + default value ----
		{Code: `
class C {
  private foo = 1;
  method() {
    const { foo: aliased = 99 } = this;
    return aliased;
  }
}
    `},

		// ---- Dimension 4: chained TS wrappers on the receiver ----
		{Code: `
class C {
  private foo = 1;
  bar() {
    return ((this as C) satisfies C).foo;
  }
}
    `},
		{Code: `
class C {
  private foo = 1;
  bar() {
    return ((this!) as C)!.foo;
  }
}
    `},

		// ---- Dimension 4: optional-chain link in a longer chain ----
		// (`self: C | null` would not resolve — upstream only matches the
		// TSTypeReference branch, not unions — so we keep the parameter as
		// non-nullable to actually exercise the access path.)
		{Code: `
class C {
  private foo = { bar: 1 };
  use(self: C) {
    return self?.foo?.bar;
  }
}
    `},

		// ---- Dimension 4: read context — typeof / void / delete / await / yield ----
		{Code: `
class C {
  private foo = 1;
  bar() {
    return typeof this.foo;
  }
}
    `},
		{Code: `
class C {
  private foo = 1;
  bar() {
    return void this.foo;
  }
}
    `},
		{Code: `
class C {
  private foo: number | undefined = 1;
  bar() {
    return delete this.foo;
  }
}
    `},
		{Code: `
class C {
  #promise = Promise.resolve(1);
  async bar() {
    return await this.#promise;
  }
}
    `},
		{Code: `
class C {
  #values = [1, 2, 3];
  *gen() {
    yield this.#values;
  }
}
    `},

		// ---- Dimension 4: spread access in value position (read) ----
		{Code: `
class C {
  #arr = [1, 2, 3];
  copy() {
    return [...this.#arr];
  }
}
    `},
		{Code: `
class C {
  private data = { a: 1 };
  copy() {
    return { ...this.data };
  }
}
    `},
		{Code: `
class C {
  #args = [1, 2, 3];
  call(fn: (...args: number[]) => void) {
    fn(...this.#args);
  }
}
    `},

		// ---- Dimension 4: template literal interpolation is a read ----
		{Code: "\nclass C {\n  #title = 'rslint';\n  greet() {\n    return `hello, ${this.#title}`;\n  }\n}\n    "},
		{Code: "\nclass C {\n  private label = 'x';\n  format() {\n    return tagged`prefix-${this.label}-suffix`;\n  }\n}\nfunction tagged(parts: TemplateStringsArray, ...args: string[]) {\n  return parts.join(',') + args.join(',');\n}\n    "},

		// ---- Dimension 4: nested destructuring — only top-level key counted ----
		{Code: `
class C {
  private cfg = { nested: { x: 1 } };
  method() {
    const { cfg: { nested: { x } } } = this;
    return x;
  }
}
    `},

		// ---- Dimension 4: array destructuring of a private member is a read ----
		{Code: `
class C {
  #pair = [1, 2];
  swap() {
    const [a, b] = this.#pair;
    return [b, a];
  }
}
    `},
		{Code: `
class C {
  #arr = [1, 2, 3];
  first() {
    const [, , third] = this.#arr;
    return third;
  }
}
    `},

		// ---- Dimension 4: read followed by write in compound expression ----
		{Code: `
class C {
  private counter = 0;
  inc() {
    return (this.counter = this.counter + 1);
  }
}
    `},
		{Code: `
class C {
  #val = 1;
  combine() {
    return this.#val ?? 0;
  }
}
    `},
		{Code: `
class C {
  #val = 1;
  use() {
    const result = (this.#val, 'sequence');
    return result;
  }
}
    `},

		// ---- Dimension 4: deeply nested classes — outer member used by
		// 3rd-level deep this access ----
		{Code: `
class L1 {
  #shared = 1;
  m1() {
    return class L2 {
      m2() {
        return class L3 {
          m3() {
            // Inner-most class doesn't define #shared, so the walk-up
            // finds L1's. (Hash-private resolution walks past intermediate
            // class scopes whose member maps don't contain the key.)
            const helper = (x: L1) => x.#shared;
            return helper;
          }
        };
      }
    };
  }
}
    `},

		// ---- Dimension 4: class inside arrow inside method ----
		{Code: `
class Outer {
  #seed = 1;
  factory() {
    return () => {
      return class Inner {
        method(o: Outer) {
          return o.#seed;
        }
      };
    };
  }
}
    `},

		// ---- Dimension 4: static block with `this` access (this = class object) ----
		{Code: `
class C {
  private static items: number[] = [];
  static {
    this.items.push(1);
  }
}
    `},

		// ---- Dimension 4: computed name with literal — member key resolves
		// to the literal value via GetStaticPropertyName ----
		{Code: `
class C {
  private ['computedName'] = 1;
  use() {
    return this.computedName;
  }
}
    `},
		{Code: "\nclass C {\n  private [`templateKey`] = 1;\n  use() {\n    return this.templateKey;\n  }\n}\n    "},
		{Code: `
class C {
  private [42] = 'forty-two';
  use() {
    return this[42];
  }
}
    `},

		// ---- Dimension 4: TS abstract get/set pair — only one half used,
		// counts because the abstract member is a single accessor entry. ----
		{Code: `
abstract class C {
  protected abstract get id(): string;
  use() {
    return this.id;
  }
}
    `},

		// ---- Dimension 4: generic class — the type parameter doesn't hide
		// the class scope lookup. ----
		{Code: `
class Box<T> {
  #value: T | null = null;
  get() {
    return this.#value;
  }
  set(v: T) {
    this.#value = v;
  }
}
    `},

		// ---- Dimension 4: class implementing interface ----
		{Code: `
interface Greeter {
  greet(): string;
}
class C implements Greeter {
  private salutation = 'hi';
  greet() {
    return this.salutation;
  }
}
    `},

		// ---- Dimension 4: decorator on a private member doesn't affect
		// tracking; the access on the next line is still counted. ----
		{Code: `
function dec(_target: any, _key: string) {}
class C {
  @dec
  private decorated = 1;
  use() {
    return this.decorated;
  }
}
    `},

		// ---- Dimension 4: arrow function as static field accesses class via
		// the static name (since `this` in field initializer is unreliable). ----
		{Code: `
class C {
  private static count = 0;
  private static incCount = () => C.count + 1;
  static run() {
    return C.incCount();
  }
}
    `},

		// ---- Real-user: singleton pattern (private static instance) ----
		{Code: `
class Singleton {
  private static instance: Singleton | null = null;
  static getInstance() {
    if (!Singleton.instance) {
      Singleton.instance = new Singleton();
    }
    return Singleton.instance;
  }
}
    `},

		// ---- Real-user: builder pattern with chained methods reading own state ----
		{Code: `
class Builder {
  #parts: string[] = [];
  add(part: string) {
    this.#parts.push(part);
    return this;
  }
  build() {
    return this.#parts.join('');
  }
}
    `},

		// ---- Real-user: observer / event emitter pattern ----
		{Code: `
class Emitter {
  #listeners: Array<() => void> = [];
  on(fn: () => void) {
    this.#listeners.push(fn);
  }
  emit() {
    for (const fn of this.#listeners) {
      fn();
    }
  }
}
    `},

		// ---- Real-user: bound method handler (this.handle = this.handle.bind(this)) ----
		{Code: `
class Component {
  #boundHandler: () => void;
  constructor() {
    this.#boundHandler = this.#handle.bind(this);
  }
  #handle() {}
  attach(el: { addEventListener(e: string, fn: () => void): void }) {
    el.addEventListener('click', this.#boundHandler);
  }
}
    `},

		// ---- Real-user: memoization with private cache ----
		{Code: `
class Memo {
  #cache = new Map<string, number>();
  compute(key: string) {
    if (this.#cache.has(key)) {
      return this.#cache.get(key);
    }
    const value = key.length;
    this.#cache.set(key, value);
    return value;
  }
}
    `},

		// ---- Real-user: iterator protocol ----
		{Code: `
class Range {
  #from: number;
  #to: number;
  constructor(from: number, to: number) {
    this.#from = from;
    this.#to = to;
  }
  *[Symbol.iterator]() {
    for (let i = this.#from; i < this.#to; i++) {
      yield i;
    }
  }
}
    `},

		// ---- Locks in member-collection: declare member (ambient) — the
		// declaration is still a class member; access works as usual. ----
		{Code: `
class C {
  private declare foo: number;
  init(v: number) {
    Object.assign(this, { foo: v });
    return this.foo;
  }
}
    `},

		// ---- Locks in member-collection: `static get` / `static set` accessor
		// pair where only the setter is used through `C.x = …`. ----
		{Code: `
class C {
  private static get x() {
    return 1;
  }
  private static set x(_v: number) {}
  static init() {
    C.x = 2;
    return C.x;
  }
}
    `},

		// ---- Locks in handleThisDestructuring with rest sibling: rest gathers
		// non-listed properties but doesn't count any specific member. The
		// explicitly-named `foo` IS counted. ----
		{Code: `
class C {
  private foo = 1;
  use() {
    const { foo, ...rest } = this;
    return [foo, rest];
  }
}
    `},

	}, []rule_tester.InvalidTestCase{
		// ---- Locks in member-collection: overloaded method signatures —
		// only the implementation body matters, all overloads share one
		// member key. Member is reported once (not per overload). The
		// diagnostic anchors on the surviving (last-collected) member,
		// which is the implementation. ----
		{
			Code: `
class C {
  private foo(a: string): void;
  private foo(a: number): void;
  private foo(a: any) {
    return a;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 5, Column: 11},
			},
		},
		// ---- Locks in upstream PrivateIdentifier visitor: nested classes shadow
		// outer hash-private members per spec — outer member stays unused even if
		// inner class uses a member with the same #name. ----
		{
			Code: `
class Outer {
  #shared = 1;
  use() {
    return class Inner {
      #shared = 2;
      consume() {
        return this.#shared;
      }
    };
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 3},
			},
		},

		// ---- Dimension 4: receiver-with-non-null-assertion that doesn't
		// resolve to a tracked class still leaves the member unused. ----
		{
			Code: `
class C {
  private foo = 1;
}

declare const arr: C[];
console.log(arr[0]!.foo);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 11},
			},
		},

		// ---- Dimension 4: shorthand destructuring at the class scope sees the
		// inherited `this`; member is still unused because the destructured
		// binding is consumed but `this.foo` itself is only read into a temp. ----
		// (Verifies handleThisDestructuring counts the read AND that pure
		// assignment without subsequent use is the only thing that flags.)
		{
			Code: `
class C {
  private foo = 1;
  method() {
    this.foo = 2;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 11},
			},
		},

		// ---- Locks in classifyAccess UpdateExpression branch: ++ inside a
		// non-statement context (`++this.#x` used as a value) counts as a
		// READ because the produced value is consumed elsewhere. Upstream:
		// `identifierGreatGrandparent !== ExpressionStatement` falls through
		// to readCount++.
		// Inverse case: when ++ IS a statement → write-only. ----
		{
			Code: `
class C {
  #counter = 0;
  bump() {
    ++this.#counter;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 3},
			},
		},

		// ---- Locks in classifyAccess BinaryExpression compound-op branch:
		// `+=` whose result is discarded (ExpressionStatement parent) ⇒ write.
		// Pair with the upstream valid case `bar((this.#x += 1))`. ----
		{
			Code: `
class C {
  #total = 0;
  bump(amount: number) {
    this.#total += amount;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 3},
			},
		},

		// ---- Locks in handleThisDestructuring computed-key skip: even when
		// the computed key resolves at compile time to a tracked member name,
		// upstream intentionally bails (`prop.computed` skip). ----
		{
			Code: `
class C {
  private foo = 1;
  method() {
    const key = 'foo';
    const { [key]: alias } = this;
    void alias;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 11},
			},
		},

		// ---- Real-user: assignment inside an arrow inherits `this` from the
		// enclosing method scope. `this.#cached = bar()` is a pure write ⇒ the
		// member stays unused. ----
		{
			Code: `
class Memo {
  #cached: number | null = null;
  compute(bar: () => number) {
    const update = () => {
      this.#cached = bar();
    };
    update();
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 3},
			},
		},

		// ---- Locks in upstream getObjectClass: `Foo.helper` accessed from a
		// sibling class is NOT counted (findClassScopeWithName only walks the
		// upper chain, which from inside Bar reaches Bar then root, never Foo).
		// Even though the access is syntactically present, upstream treats it
		// as unresolved — the private member is reported. ----
		{
			Code: `
class Foo {
  private static helper = 1;
}
class Bar extends Foo {
  static use() {
    return Foo.helper;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 18},
			},
		},

		// ---- Real-user: destructuring from `this` only reads matching keys.
		// `private bar` is never destructured, so it stays unused. ----
		{
			Code: `
class C {
  private foo = 1;
  private bar = 2;
  method() {
    const { foo } = this;
    void foo;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 4, Column: 11},
			},
		},

		// ---- Locks the rest-binding bug fix: `const { ...rest } = this`
		// must NOT register `rest` as an access of a same-named class
		// member. The rest target is a *binding name*, not a property key,
		// so it should be filtered out. ----
		{
			Code: `
class C {
  private rest = 1;
  method() {
    const { ...rest } = this;
    return rest;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 11},
			},
		},

		// ---- Locks in upstream: external bracket access with a STATIC literal
		// from outside the class is also treated as unresolved (the receiver
		// `instance` doesn't resolve to a tracked class scope by name). ----
		{
			Code: `
class C {
  private static foo = 1;
}
console.log(C['foo']);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 18},
			},
		},

		// ---- Locks in upstream: union-typed parameter — only TSTypeReference
		// is matched, unions / intersections fall through. ----
		{
			Code: `
class C {
  private prop = 1;
  method(thing: C | null) {
    return thing?.prop;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 11},
			},
		},

		// ---- Locks in upstream: array-typed parameter — `Foo[]` is a
		// TypeReference whose name resolves to `Array`, not `Foo`. So an
		// access through an array element (without indexing) doesn't match. ----
		{
			Code: `
class C {
  private prop = 1;
  process(arr: C[]) {
    return arr.map(x => x.prop);
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 11},
			},
		},

		// ---- Locks in upstream: `this`-aliasing must be direct (`X = this`).
		// `let X = anything-else; X.foo` is not treated as an alias. ----
		{
			Code: `
class C {
  private prop = 1;
  method(other: C) {
    const self = other;
    return self.prop;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 11},
			},
		},

		// ---- Locks in pushFunctionScope: a regular (non-arrow) function
		// nested inside a method rebinds `this`. Without a typed `this`
		// parameter, the inner function's `this.#x` doesn't reach the
		// outer class. ----
		{
			Code: `
class C {
  #data = 1;
  method() {
    function inner(this: void) {
      return this; // void context, no #data access
    }
    inner.call(undefined);
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 3},
			},
		},

		// ---- Locks in classifyAccess: for-in / for-of `this.#x` as the
		// initializer is a write (`for (this.#x in obj)` /
		// `for (this.#x of arr)`). No read. ----
		{
			Code: `
class C {
  #key = '';
  scan(obj: object) {
    for (this.#key in obj) {
    }
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 3},
			},
		},

		// ---- Locks in member-collection: TS abstract property without body
		// is still a tracked member. Without any read it's reported. ----
		{
			Code: `
abstract class C {
  protected abstract count: number;
  private internal = 0;
  abstract describe(): string;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 4, Column: 11},
			},
		},

		// ---- Locks in member-collection: `private static readonly` constants
		// are tracked like any other member. ----
		{
			Code: `
class Config {
  private static readonly TIMEOUT = 5000;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 27},
			},
		},

		// ---- Locks in handleThisDestructuring: rest-only destructuring with
		// a same-named private member must NOT count the member as accessed
		// (regression: the rest binding is a target, not a key). ----
		{
			Code: `
class C {
  private values = 1;
  method() {
    const { ...values } = this;
    return values;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 11},
			},
		},

		// ---- Locks in classifyAccess: ForInStatement initializer through
		// a NESTED destructuring left side — `for (this.#k of arr)` is a
		// write; the wrapping array spread doesn't escape it to a read. ----
		{
			Code: `
class C {
  #buffer: number[] = [];
  fill(src: number[]) {
    for (this.#buffer of [src]) {
      // body intentionally empty
    }
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 3},
			},
		},

		// ---- Locks in classifyAccess: destructuring assignment writes to
		// every nested target. `[this.#a, this.#b] = pair` → both writes. ----
		{
			Code: `
class C {
  #a = 0;
  #b = 0;
  fill(pair: [number, number]) {
    [this.#a, this.#b] = pair;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 3},
				{MessageId: "unusedPrivateClassMember", Line: 4, Column: 3},
			},
		},

		// ---- Locks in resolveBySymbol: a variable with NO type annotation
		// (just an `init`) and not `= this` is left unresolved — even when
		// the runtime value happens to be a class instance. ----
		{
			Code: `
class C {
  private prop = 1;
  static make() {
    const inst = new C();
    return inst.prop;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 11},
			},
		},

		// ---- Locks pushFunctionScope `this:` typed param routing through
		// upper.findClassByName: a function declared OUTSIDE the class
		// can't see the class via lexical upper chain, so its typed `this`
		// parameter doesn't resolve. ----
		{
			Code: `
class C {
  #prop = 1;
}
function freeFn(this: C) {
  return this.#prop;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 3},
			},
		},

		// ---- Locks in handleThisDestructuring: nested destructuring counts
		// only the TOP-LEVEL key. The inner level doesn't count `nested` as
		// an access of a `nested` member of `this`. ----
		{
			Code: `
class C {
  private cfg = { nested: { x: 1 } };
  private nested = 'unrelated';
  method() {
    const { cfg: { nested: { x } } } = this;
    return x;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 4, Column: 11},
			},
		},

		// ---- Locks in classifyAccess: a member appearing both on LHS and RHS
		// of `this.#x = this.#y` — left is a write, right is a read.
		// Only the writer-only member is reported. ----
		{
			Code: `
class C {
  #target = 0;
  #source = 0;
  copy() {
    this.#target = this.#source;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 3},
			},
		},

		// ---- Locks in resolveBySymbol fall-through: a variable typed as the
		// containing class's `T` generic parameter (not the class itself).
		// The type name is `T`, which has no matching classScope. ----
		{
			Code: `
class Box<T> {
  private content: T | null = null;
  static handle<U>(value: U) {
    return value;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 11},
			},
		},

		// ---- Locks in member-collection: getter and setter share the same
		// member key (both isAccessor=true). When neither has any access,
		// only ONE diagnostic is emitted (last-wins anchors on the setter). ----
		{
			Code: `
class C {
  private get count() {
    return 0;
  }
  private set count(_v: number) {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 6, Column: 15},
			},
		},

		// ---- Locks the alias-invalidation fix: once a `let X = this` is
		// reassigned, subsequent `X.foo` no longer counts as a class member
		// access. Matches upstream's `variable.references.some(isWrite)` bail. ----
		{
			Code: `
class C {
  private prop = 1;
  method(other: { prop: number }) {
    let self = this;
    self = other;
    return self.prop;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 11},
			},
		},

		// ---- Locks alias invalidation through compound assignment: any write
		// to the alias name (including +=) disowns the alias. ----
		{
			Code: `
class C {
  private prop = 1;
  method(other: any) {
    let self = this as any;
    self ||= other;
    return self.prop;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 11},
			},
		},

		// ---- Locks ORDER-INDEPENDENT alias invalidation: even when the
		// `X.foo` READ appears textually BEFORE the `X = other` write, the
		// post-walk reconciliation rolls back the count. Matches upstream's
		// scope-manager-based `references.some(isWrite)` bailout — which is
		// also order-independent. ----
		{
			Code: `
class C {
  private prop = 1;
  method(other: { prop: number }) {
    let self = this;
    const captured = self.prop;
    self = other;
    return captured;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 11},
			},
		},

		// ---- Locks invalidation across nested function boundary: a write
		// inside an arrow function still invalidates the outer-scope alias
		// (symbol identity is shared across scopes). ----
		{
			Code: `
class C {
  private prop = 1;
  method(other: { prop: number }) {
    let self = this;
    const useThenWrite = () => {
      const r = self.prop;
      self = other;
      return r;
    };
    return useThenWrite();
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 11},
			},
		},

		// ---- Locks invalidation via destructuring assignment to the alias
		// name: `[self] = [other]` is a write to `self`. ----
		{
			Code: `
class C {
  private prop = 1;
  method(other: { prop: number }) {
    let self = this;
    const r = self.prop;
    [self] = [other];
    return r;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 11},
			},
		},

		// ---- Locks alias-invalidation triggered inside a control-flow loop:
		// `self = other` happens conditionally on each iteration, the
		// reads of `self.prop` appear both before AND inside the loop. ----
		{
			Code: `
class C {
  private prop = 1;
  iterate(others: { prop: number }[]) {
    let self = this;
    const before = self.prop;
    for (const o of others) {
      if (o) {
        self = o;
      }
      void self.prop;
    }
    return before;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 11},
			},
		},

		// ---- Locks invalidation via for-of as a write target ----
		{
			Code: `
class C {
  private prop = 1;
  method(arr: any[]) {
    let self = this as any;
    const r = self.prop;
    for (self of arr) {
      // body intentionally empty
    }
    return r;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedPrivateClassMember", Line: 3, Column: 11},
			},
		},

		// ---- Locks symbol-based aliasing through shadowing: an INNER `let X`
		// that shadows the outer alias `X` does NOT invalidate the outer
		// alias when written, because the symbols are distinct. ----
		// (The outer X is never written → its alias survives. The inner X
		// access doesn't resolve to the outer C class — it's `string`. So
		// `prop` IS used (via the outer self.prop on line 6) and NOT
		// reported.)
		{
			Code: `
class C {
  private prop = 1;
  private inner = 'inner-only';
  method() {
    let self = this;
    const r = self.prop;
    {
      let self = 'shadow';
      self = 'rewritten';
      void self;
    }
    return r;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				// Only `inner` should be reported — the outer self.prop
				// access on `prop` survives the symbol-shadowing test, but
				// `inner` was never accessed.
				{MessageId: "unusedPrivateClassMember", Line: 4, Column: 11},
			},
		},
	})
}

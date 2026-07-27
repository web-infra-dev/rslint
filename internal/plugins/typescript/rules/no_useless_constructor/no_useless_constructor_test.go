package no_useless_constructor

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUselessConstructorRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUselessConstructorRule, []rule_tester.ValidTestCase{
		// ============================================================
		// Basic valid cases (no constructor or constructor with logic)
		// ============================================================
		{Code: `class A {}`},
		{Code: `
class A {
  constructor() {
    doSomething();
  }
}`},
		{Code: `
class A {
  dummyMethod() {
    doSomething();
  }
}`},

		// ============================================================
		// Extends: constructor not just forwarding to super
		// ============================================================

		// Empty body in extended class (not calling super at all)
		{Code: `
class A extends B {
  constructor() {}
}`},
		// Calling super with different/extra args
		{Code: `
class A extends B {
  constructor() {
    super('foo');
  }
}`},
		{Code: `
class A extends B {
  constructor(foo, bar) {
    super(foo, bar, 1);
  }
}`},
		// Super + additional logic
		{Code: `
class A extends B {
  constructor() {
    super();
    doSomething();
  }
}`},
		{Code: `
class A extends B {
  constructor(...args) {
    super(...args);
    doSomething();
  }
}`},
		// Fewer args to super
		{Code: `
class A extends B {
  constructor(a, b, c) {
    super(a, b);
  }
}`},
		{Code: `
class A extends B {
  constructor(foo, bar) {
    super(foo);
  }
}`},
		// Params but super called with no args
		{Code: `
class A extends B {
  constructor(test) {
    super();
  }
}`},
		// Body is not super call
		{Code: `
class A extends B {
  constructor() {
    foo;
  }
}`},
		// Different arg order to super
		{Code: `
class A extends B {
  constructor(foo, bar) {
    super(bar);
  }
}`},

		// ============================================================
		// Extends with property access / complex expressions
		// ============================================================
		{Code: `
class A extends B.C {
  constructor() {
    super(foo);
  }
}`},
		// Deep property access extends
		{Code: `
class A extends a.b.c.D {
  constructor() {
    super(foo);
  }
}`},
		// Extends with call expression (Mixin pattern) — not a useless constructor
		{Code: `
class A extends Mixin(Base) {
  constructor() {
    super();
    this.init();
  }
}`},

		// ============================================================
		// Non-simple parameters (destructuring, defaults)
		// ============================================================
		{Code: `
class A extends B.C {
  constructor([a, b, c]) {
    super(...arguments);
  }
}`},
		{Code: `
class A extends B.C {
  constructor(a = f()) {
    super(...arguments);
  }
}`},
		// Object destructuring
		{Code: `
class A extends B {
  constructor({ x, y }) {
    super(...arguments);
  }
}`},
		// Default value on param
		{Code: `
class A extends B {
  constructor(a = 1) {
    super(a);
  }
}`},

		// ============================================================
		// Declaration-only constructors (no body → skip)
		// ============================================================
		{Code: `
declare class A {
  constructor();
}`},
		{Code: `
class A {
  constructor();
}`},
		{Code: `
abstract class A {
  constructor();
}`},
		// tsgo parses these as constructors, but ESTree exposes methods whose
		// kind is "method", so the upstream rule ignores them.
		{Code: `class A { static constructor() {} }`},
		{Code: `class A extends B { static constructor() {} }`},
		// Overload + implementation with body
		{Code: `
class A {
  constructor(x: number);
  constructor(x: string);
  constructor(x: number | string) {
    doSomething(x);
  }
}`},

		// ============================================================
		// TypeScript parameter properties (useful → skip)
		// ============================================================
		{Code: `
class A {
  constructor(private name: string) {}
}`},
		{Code: `
class A {
  constructor(public name: string) {}
}`},
		{Code: `
class A {
  constructor(protected name: string) {}
}`},
		// readonly parameter property
		{Code: `
class A {
  constructor(readonly name: string) {}
}`},
		// Multiple param properties
		{Code: `
class A {
  constructor(public x: number, public y: number) {}
}`},
		// Mix of param property + normal param
		{Code: `
class A {
  constructor(public name: string, age: number) {}
}`},
		// readonly + visibility
		{Code: `
class A {
  constructor(private readonly id: number) {}
}`},
		// Param property in extended class forwarding to super
		{Code: `
class A extends B {
  constructor(public name: string) {
    super(name);
  }
}`},

		// ============================================================
		// Access modifiers on constructor (useful → skip)
		// ============================================================
		{Code: `
class A {
  private constructor() {}
}`},
		{Code: `
class A {
  protected constructor() {}
}`},
		// Public constructor in extended class (changes parent visibility)
		{Code: `
class A extends B {
  public constructor() {}
}`},
		// Protected constructor with different super args
		{Code: `
class A extends B {
  protected constructor(foo, bar) {
    super(bar);
  }
}`},
		// Private constructor with different super args
		{Code: `
class A extends B {
  private constructor(foo, bar) {
    super(bar);
  }
}`},
		// Public constructor in extended class forwarding args (visibility change)
		{Code: `
class A extends B {
  public constructor(foo) {
    super(foo);
  }
}`},
		{Code: `
class A extends B {
  public constructor(foo) {}
}`},
		// Public constructor with super() in extended class
		{Code: `
class A extends B {
  public constructor() {
    super();
  }
}`},

		// ============================================================
		// Decorators on params (useful → skip)
		// ============================================================
		{Code: `
class A extends Object {
  constructor(@Foo foo: string) {
    super(foo);
  }
}`},
		{Code: `
class A extends Object {
  constructor(foo: string, @Bar() bar) {
    super(foo, bar);
  }
}`},
		// Decorator on param in non-extended class
		{Code: `
class A {
  constructor(@Inject service: Service) {}
}`},

		// ============================================================
		// Constructor with useful body logic
		// ============================================================
		// this.x = x assignment
		{Code: `
class A {
  constructor(x) {
    this.x = x;
  }
}`},
		// Super + property assignment
		{Code: `
class A extends B {
  constructor(x) {
    super(x);
    this.extra = true;
  }
}`},
		// Super call inside try-catch
		{Code: `
class A extends B {
  constructor() {
    try { super(); } catch(e) {}
  }
}`},
		// Multiple statements
		{Code: `
class A extends B {
  constructor() {
    super();
    console.log('constructed');
  }
}`},

		// ============================================================
		// Class expressions (valid cases)
		// ============================================================
		{Code: `const A = class {}`},
		{Code: `
const A = class {
  constructor() {
    doSomething();
  }
}`},
		{Code: `
const A = class extends B {
  constructor() {
    super();
    doSomething();
  }
}`},

		// ============================================================
		// Nested classes (inner class is valid)
		// ============================================================
		{Code: `
class Outer {
  constructor() {
    class Inner {
      constructor() {
        doSomething();
      }
    }
  }
}`},
		// Inner class with useful constructor, outer has none
		{Code: `
class Outer {
  method() {
    return class Inner extends Base {
      constructor() {
        super();
        this.init();
      }
    }
  }
}`},

		// ============================================================
		// Type annotations on params (should not affect logic)
		// ============================================================
		// Type-annotated params forwarding — still useless, but this is valid
		// because the type annotation doesn't change the forwarding semantics
		// Wait — actually this SHOULD be invalid. Hmm, let me reconsider:
		// `constructor(x: number) { super(x); }` — this is just forwarding with
		// a type annotation. Type annotations don't make it useful. BUT we need
		// to verify isSimpleParam handles type-annotated params.
		// Actually this should be in INVALID. Moving it there.

		// Optional param — not forwarding (different behavior)
		{Code: `
class A extends B {
  constructor(x?: number) {}
}`},

		// ============================================================
		// Class with implements (not extends, so no superClass)
		// ============================================================
		{Code: `
class A implements Serializable {
  constructor() {
    this.init();
  }
}`},

		// ============================================================
		// Body content that is NOT super(): various statement types
		// ============================================================
		// Empty statement (semicolon) — has 1 statement, so not empty
		{Code: `
class A {
  constructor() { ; }
}`},
		// Return statement
		{Code: `
class A {
  constructor() { return; }
}`},
		// "use strict" directive
		{Code: `
class A extends B {
  constructor() {
    "use strict";
    super();
  }
}`},

		// ============================================================
		// super.method() vs super() — must be super() call specifically
		// ============================================================
		{Code: `
class A extends B {
  constructor() {
    super.init();
  }
}`},

		// ============================================================
		// Super wrapped in other expressions (not a plain super call)
		// ============================================================
		// void super()
		{Code: `
class A extends B {
  constructor() {
    void super();
  }
}`},
		// Comma expression: super(), doSomething()
		{Code: `
class A extends B {
  constructor() {
    super(), doSomething();
  }
}`},
		// Conditional super
		{Code: `
class A extends B {
  constructor(x) {
    if (x) { super(x); } else { super(); }
  }
}`},

		// ============================================================
		// extends null
		// ============================================================
		{Code: `
class A extends null {
  constructor() {
    doSomething();
  }
}`},

		// ============================================================
		// Extends + implements combined (valid: useful body)
		// ============================================================
		{Code: `
class A extends B implements I {
  constructor() {
    super();
    this.init();
  }
}`},

		// ============================================================
		// new.target usage (useful body)
		// ============================================================
		{Code: `
class A {
  constructor() {
    if (new.target === A) throw new Error();
  }
}`},

		// ============================================================
		// Nested: outer useless, but inner has logic (only outer flagged)
		// Both classes handled independently
		// ============================================================
		// (tested via invalid cases below — outer is useless)

		// Nested: outer has logic, inner is valid
		{Code: `
class Outer {
  constructor() {
    this.inner = class Inner {
      constructor() {
        doSomething();
      }
    }
  }
}`},

		// ============================================================
		// TypeScript `this` parameter (not simple → blocks forwarding check)
		// ============================================================
		// `this` param in extended class prevents forwarding match
		{Code: `
class A extends B {
  constructor(this: Foo) {
    super();
  }
}`},

		// ============================================================
		// Generic class (generics don't affect uselessness logic)
		// ============================================================
		// Generic class with useful constructor body
		{Code: `
class A<T> extends B<T> {
  constructor(x: T) {
    super(x);
    this.data = x;
  }
}`},

		// ============================================================
		// export default class (useful)
		// ============================================================
		{Code: `
export default class extends B {
  constructor() {
    super();
    this.init();
  }
}`},

		// ============================================================
		// Decorator + parameter property on same param
		// ============================================================
		{Code: `
class A {
  constructor(@Inject public name: string) {}
}`},

		// ============================================================
		// readonly param in extended class (param property → skip)
		// ============================================================
		{Code: `
class A extends B {
  constructor(readonly x: number) {
    super(x);
  }
}`},

		// ============================================================
		// Type assertion in super arg (not a plain identifier → not forwarding)
		// ============================================================
		{Code: `
class A extends B {
  constructor(x) {
    super(x as any);
  }
}`},

		// ============================================================
		// Spread of non-identifier expression in super (not forwarding)
		// ============================================================
		{Code: `
class A extends B {
  constructor(...args) {
    super(...args.slice(0));
  }
}`},

		// ============================================================
		// Regular param but super arg is spread (mismatch types)
		// ============================================================
		{Code: `
class A extends B {
  constructor(a, b) {
    super(a, ...b);
  }
}`},

		// ============================================================
		// Rest param but super arg is NOT spread (length matches, type mismatch)
		// Exercises isValidRestSpreadPair: superArg.Kind != KindSpreadElement
		// ============================================================
		{Code: `
class A extends B {
  constructor(...args) {
    super(args);
  }
}`},

		// ============================================================
		// Mixed: params forwarded but extra super args
		// ============================================================
		{Code: `
class A extends B {
  constructor(x) {
    super(x, 'extra');
  }
}`},

		// ============================================================
		// Params reordered in super call
		// ============================================================
		{Code: `
class A extends B {
  constructor(a, b, c) {
    super(c, b, a);
  }
}`},

		// ============================================================
		// Rest param not at last position would be syntax error,
		// but rest param + extra normal param forwarding: not 1:1
		// ============================================================
		{Code: `
class A extends B {
  constructor(a, ...rest) {
    super(a);
  }
}`},
	}, []rule_tester.InvalidTestCase{
		// ============================================================
		// Basic invalid: empty constructor in non-extended class
		// ============================================================
		{
			Code: `
class A {
  constructor() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 3, Column: 3, EndLine: 3, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  \n}"},
				}},
			},
		},

		// ============================================================
		// Extended class: just forwarding to super
		// ============================================================
		{
			Code: `
class A extends B {
  constructor() {
    super();
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 3, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A extends B {\n  \n}"},
				}},
			},
		},
		// ESTree erases grouping parentheses around expressions. tsgo keeps
		// them, so they must not hide the same redundant-super shapes.
		{
			Code: `class A extends B { constructor() { (((super()))); } }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 1, Column: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "class A extends B {  }"},
				}},
			},
		},
		{
			Code: `class A extends B { constructor(value) { super((value)); } }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 1, Column: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "class A extends B {  }"},
				}},
			},
		},
		{
			Code: `class A extends B { constructor(...values) { super(...(values)); } }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 1, Column: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "class A extends B {  }"},
				}},
			},
		},
		{
			Code: `class A extends B { constructor() { super(...(arguments)); } }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 1, Column: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "class A extends B {  }"},
				}},
			},
		},
		// Single param forwarding
		{
			Code: `
class A extends B {
  constructor(foo) {
    super(foo);
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 3, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A extends B {\n  \n}"},
				}},
			},
		},
		// Multiple params forwarding
		{
			Code: `
class A extends B {
  constructor(foo, bar) {
    super(foo, bar);
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 3, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A extends B {\n  \n}"},
				}},
			},
		},
		// Spread args forwarding
		{
			Code: `
class A extends B {
  constructor(...args) {
    super(...args);
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 3, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A extends B {\n  \n}"},
				}},
			},
		},
		// ...arguments forwarding
		{
			Code: `
class A extends B.C {
  constructor() {
    super(...arguments);
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 3, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A extends B.C {\n  \n}"},
				}},
			},
		},
		// Rest params with ...arguments
		{
			Code: `
class A extends B {
  constructor(a, b, ...c) {
    super(...arguments);
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 3, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A extends B {\n  \n}"},
				}},
			},
		},
		// ESTree treats every RestElement as simple. The binding pattern does
		// not block a direct ...arguments pass-through.
		{
			Code: `
class A extends B {
  constructor(...[x, y]) {
    super(...arguments);
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 3, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A extends B {\n  \n}"},
				}},
			},
		},
		// Rest params matching forward
		{
			Code: `
class A extends B {
  constructor(a, b, ...c) {
    super(a, b, ...c);
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 3, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A extends B {\n  \n}"},
				}},
			},
		},

		// Removing a constructor must not let a preceding field initializer
		// consume the next computed member through ASI.
		{
			Code: `
class A {
  field = 'value'
  constructor() {}
  [0]() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 4, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  field = 'value'\n  ;\n  [0]() {}\n}"},
				}},
			},
		},
		{
			Code: `
class A {
  field = 2
  constructor() {}
  *iterator() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 4, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  field = 2\n  ;\n  *iterator() {}\n}"},
				}},
			},
		},
		{
			Code: `
class A {
  method() {}
  constructor() {}
  [0]() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 4, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  method() {}\n  \n  [0]() {}\n}"},
				}},
			},
		},

		// ============================================================
		// Public constructor without superClass (useless)
		// ============================================================
		{
			Code: `
class A {
  public constructor() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 3, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  \n}"},
				}},
			},
		},

		// ============================================================
		// Class expressions (invalid)
		// ============================================================
		// Class expression with empty constructor
		{
			Code: `const A = class { constructor() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 1, Column: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "const A = class {  }"},
				}},
			},
		},
		// Named class expression with empty constructor
		{
			Code: `const A = class MyClass { constructor() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 1, Column: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "const A = class MyClass {  }"},
				}},
			},
		},
		// Class expression extends with just super forwarding
		{
			Code: `const A = class extends B { constructor() { super(); } }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 1, Column: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "const A = class extends B {  }"},
				}},
			},
		},

		// ============================================================
		// Nested classes: inner class with useless constructor
		// ============================================================
		{
			Code: `
class Outer {
  method() {
    class Inner {
      constructor() {}
    }
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 5, Column: 7, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass Outer {\n  method() {\n    class Inner {\n      \n    }\n  }\n}"},
				}},
			},
		},
		// Inner class expression with useless constructor
		{
			Code: `
class Outer {
  method() {
    return class extends Base {
      constructor(x) {
        super(x);
      }
    }
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 5, Column: 7, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass Outer {\n  method() {\n    return class extends Base {\n      \n    }\n  }\n}"},
				}},
			},
		},

		// ============================================================
		// Complex extends expressions
		// ============================================================
		// Deep property access extends
		{
			Code: `
class A extends a.b.c.D {
  constructor() {
    super(...arguments);
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 3, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A extends a.b.c.D {\n  \n}"},
				}},
			},
		},
		// Extends with call expression (mixin pattern)
		{
			Code: `
class A extends Mixin(Base) {
  constructor(...args) {
    super(...args);
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 3, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A extends Mixin(Base) {\n  \n}"},
				}},
			},
		},
		// Extends with generic
		{
			Code: `
class A extends Base<string> {
  constructor() {
    super();
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 3, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A extends Base<string> {\n  \n}"},
				}},
			},
		},

		// ============================================================
		// Type-annotated params: annotations don't make it useful
		// ============================================================
		{
			Code: `
class A extends B {
  constructor(x: number) {
    super(x);
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 3, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A extends B {\n  \n}"},
				}},
			},
		},
		// Multiple type-annotated params
		{
			Code: `
class A extends B {
  constructor(x: string, y: number) {
    super(x, y);
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 3, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A extends B {\n  \n}"},
				}},
			},
		},
		// Rest param with type annotation
		{
			Code: `
class A extends B {
  constructor(...args: any[]) {
    super(...args);
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 3, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A extends B {\n  \n}"},
				}},
			},
		},

		// ============================================================
		// Abstract class with empty constructor body
		// ============================================================
		{
			Code: `
abstract class A {
  constructor() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 3, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nabstract class A {\n  \n}"},
				}},
			},
		},
		// Abstract class extends with super forwarding
		{
			Code: `
abstract class A extends B {
  constructor() {
    super();
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 3, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nabstract class A extends B {\n  \n}"},
				}},
			},
		},

		// ============================================================
		// Class with implements (not extends → no superClass)
		// ============================================================
		{
			Code: `
class A implements Serializable {
  constructor() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 3, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A implements Serializable {\n  \n}"},
				}},
			},
		},

		// ============================================================
		// Optional param forwarding to super (still useless)
		// ============================================================
		{
			Code: `
class A extends B {
  constructor(x?: number) {
    super(x);
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 3, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A extends B {\n  \n}"},
				}},
			},
		},

		// ============================================================
		// extends + implements combo (super forwarding is useless)
		// ============================================================
		{
			Code: `
class A extends B implements I {
  constructor() {
    super();
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 3, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A extends B implements I {\n  \n}"},
				}},
			},
		},

		// ============================================================
		// extends null (extends clause exists → hasSuper = true)
		// ============================================================
		{
			Code: `
class A extends null {
  constructor() {
    super();
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 3, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A extends null {\n  \n}"},
				}},
			},
		},

		// ============================================================
		// Single-line format (position reporting)
		// ============================================================
		{
			Code: `class A { constructor() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 1, Column: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "class A {  }"},
				}},
			},
		},

		// ============================================================
		// Nested: both outer and inner are useless → two errors
		// ============================================================
		{
			Code: `
class Outer {
  constructor() {}
  method() {
    class Inner {
      constructor() {}
    }
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 3, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass Outer {\n  \n  method() {\n    class Inner {\n      constructor() {}\n    }\n  }\n}"},
				}},
				{MessageId: "noUselessConstructor", Line: 6, Column: 7, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass Outer {\n  constructor() {}\n  method() {\n    class Inner {\n      \n    }\n  }\n}"},
				}},
			},
		},

		// ============================================================
		// Many params forwarding to super (stress test)
		// ============================================================
		{
			Code: `
class A extends B {
  constructor(a, b, c, d, e) {
    super(a, b, c, d, e);
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 3, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A extends B {\n  \n}"},
				}},
			},
		},

		// ============================================================
		// Only rest param forwarding (single rest → single spread)
		// ============================================================
		{
			Code: `
class A extends B {
  constructor(...rest) {
    super(...rest);
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 3, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A extends B {\n  \n}"},
				}},
			},
		},

		// ============================================================
		// Generic class (generics don't prevent detection)
		// ============================================================
		{
			Code: `
class A<T> extends B<T> {
  constructor() {
    super();
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 3, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A<T> extends B<T> {\n  \n}"},
				}},
			},
		},
		// Generic param forwarding
		{
			Code: `
class A<T> extends B<T> {
  constructor(x: T) {
    super(x);
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 3, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A<T> extends B<T> {\n  \n}"},
				}},
			},
		},

		// ============================================================
		// export default class (anonymous)
		// ============================================================
		{
			Code: `
export default class extends B {
  constructor() {
    super();
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 3, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nexport default class extends B {\n  \n}"},
				}},
			},
		},
		// export default class with name
		{
			Code: `
export default class Foo {
  constructor() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 3, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nexport default class Foo {\n  \n}"},
				}},
			},
		},

		// ============================================================
		// `this` parameter in non-extended class: empty body → useless
		// (this param is not a param property, so checkParams passes)
		// ============================================================
		{
			Code: `
class A {
  constructor(this: Foo) {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 3, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  \n}"},
				}},
			},
		},

		// ============================================================
		// String literal constructor name: 'constructor'() {}
		// Parsed as the actual constructor
		// ============================================================
		{
			Code: `
class A {
  'constructor'() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 3, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  \n}"},
				}},
			},
		},
		{
			Code: `class A { "\u0063` + `on` + `struct` + `or" /* before paren */ () {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 1, Column: 11, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "class A {  }"},
				}},
			},
		},
	})
}

func TestNoUselessConstructorEditDemand(t *testing.T) {
	t.Parallel()

	const source = `class A {
  /* keep */
  public constructor() {}
}
class B extends A {
  constructor(value: number) {
    super(value);
  }
}
class C {
  field = 'value'
  constructor() {}
  [0]() {}
}`
	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(
		source,
		"no-useless-constructor-edit-demand.ts",
		"tsconfig.json",
	)
	if err != nil {
		t.Fatal(err)
	}

	run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
		t.Helper()

		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program:      program,
			File:         sourceFile.FileName(),
			HasTypeInfo:  true,
			ExcludePaths: []string{},
			GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
				return []linter.ConfiguredRule{{
					Name:     NoUselessConstructorRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return NoUselessConstructorRule.Run(ctx, nil)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{
				Demand: demand,
				Report: func(diagnostic rule.RuleDiagnostic) {
					diagnostics = append(diagnostics, diagnostic)
				},
			},
		})
		if len(diagnostics) != 3 {
			t.Fatalf("demand %d: diagnostics = %d, want 3", demand, len(diagnostics))
		}
		return diagnostics
	}

	diagnostics := map[rule.EditDemand][]rule.RuleDiagnostic{
		rule.EditDemandNone:       run(rule.EditDemandNone),
		rule.EditDemandAutofix:    run(rule.EditDemandAutofix),
		rule.EditDemandSuggestion: run(rule.EditDemandSuggestion),
		rule.EditDemandAll:        run(rule.EditDemandAll),
	}
	withoutEdits := func(diagnostic rule.RuleDiagnostic) rule.RuleDiagnostic {
		diagnostic.FixesPtr = nil
		diagnostic.Suggestions = nil
		return diagnostic
	}

	for index := range diagnostics[rule.EditDemandAll] {
		want := withoutEdits(diagnostics[rule.EditDemandAll][index])
		for demand, demandDiagnostics := range diagnostics {
			if got := withoutEdits(demandDiagnostics[index]); !reflect.DeepEqual(got, want) {
				t.Errorf("diagnostic %d changed for demand %d:\ngot:  %#v\nwant: %#v", index, demand, got, want)
			}
			if demandDiagnostics[index].FixesPtr != nil {
				t.Errorf("diagnostic %d demand %d unexpectedly has autofixes", index, demand)
			}
		}

		if diagnostics[rule.EditDemandNone][index].Suggestions != nil ||
			diagnostics[rule.EditDemandAutofix][index].Suggestions != nil {
			t.Fatalf("diagnostic %d materialized suggestions without suggestion demand", index)
		}

		suggestionOnly := diagnostics[rule.EditDemandSuggestion][index].Suggestions
		allEdits := diagnostics[rule.EditDemandAll][index].Suggestions
		if suggestionOnly == nil || !reflect.DeepEqual(suggestionOnly, allEdits) {
			t.Fatalf("diagnostic %d: suggestion and all-edits demands produced different suggestions", index)
		}
		if len(*suggestionOnly) != 1 || len((*suggestionOnly)[0].Fixes()) != 1 {
			t.Fatalf("diagnostic %d suggestions = %#v, want one suggestion with one fix", index, *suggestionOnly)
		}
		fix := (*suggestionOnly)[0].Fixes()[0]
		wantReportText := []string{"public constructor", "constructor", "constructor"}[index]
		if got := source[want.Range.Pos():want.Range.End()]; got != wantReportText {
			t.Fatalf("diagnostic %d report range covers %q, want %q", index, got, wantReportText)
		}
		wantFixText := []string{
			"public constructor() {}",
			"constructor(value: number) {\n    super(value);\n  }",
			"constructor() {}",
		}[index]
		if got := source[fix.Range.Pos():fix.Range.End()]; got != wantFixText {
			t.Fatalf("diagnostic %d fix range covers %q, want %q", index, got, wantFixText)
		}
		wantReplacement := []string{"", "", ";"}[index]
		if fix.Text != wantReplacement {
			t.Fatalf("diagnostic %d replacement = %q, want %q", index, fix.Text, wantReplacement)
		}
	}
}

func TestNoUselessConstructorRecoveryAST(t *testing.T) {
	t.Parallel()

	for _, source := range []string{
		`class A extends B { constructor() {`,
		`class A extends B { constructor(value) { super(`,
		`class A { public /* between */ constructor /* before paren */ (`,
		`class A { 'constructor`,
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()

			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: "/no-useless-constructor-recovery.ts",
				Path:     "/no-useless-constructor-recovery.ts",
			}, source, core.ScriptKindTS)
			var diagnostics []rule.RuleDiagnostic
			ctx := rule.RuleContext{
				SourceFile:     sourceFile,
				DisableManager: rule.NewDisableManager(sourceFile, rule.NewCommentStore(sourceFile)),
			}.WithDiagnosticConsumer(
				NoUselessConstructorRule.Name,
				rule.SeverityError,
				rule.DiagnosticConsumer{
					Demand: rule.EditDemandSuggestion,
					Report: func(diagnostic rule.RuleDiagnostic) {
						diagnostics = append(diagnostics, diagnostic)
					},
				},
			)
			listener := NoUselessConstructorRule.Run(ctx, nil)[ast.KindConstructor]
			var visit func(*ast.Node) bool
			visit = func(node *ast.Node) bool {
				if node.Kind == ast.KindConstructor {
					listener(node)
				}
				return node.ForEachChild(visit)
			}
			sourceFile.AsNode().ForEachChild(visit)

			for _, diagnostic := range diagnostics {
				if diagnostic.Range.Pos() < 0 ||
					diagnostic.Range.End() < diagnostic.Range.Pos() ||
					diagnostic.Range.End() > len(source) {
					t.Fatalf("out-of-bounds diagnostic range %#v for source length %d", diagnostic.Range, len(source))
				}
				if diagnostic.Suggestions == nil {
					continue
				}
				for _, suggestion := range *diagnostic.Suggestions {
					for _, fix := range suggestion.Fixes() {
						if fix.Range.Pos() < 0 ||
							fix.Range.End() < fix.Range.Pos() ||
							fix.Range.End() > len(source) {
							t.Fatalf("out-of-bounds fix range %#v for source length %d", fix.Range, len(source))
						}
					}
				}
			}
		})
	}
}

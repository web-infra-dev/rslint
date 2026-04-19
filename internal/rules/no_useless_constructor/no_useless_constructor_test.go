package no_useless_constructor

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUselessConstructorRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUselessConstructorRule, []rule_tester.ValidTestCase{
		// ---- No constructor ----
		{Code: `class A { }`},
		// ---- Constructor with body logic ----
		{Code: `class A { constructor(){ doSomething(); } }`},
		{Code: `class A { dummyMethod(){ doSomething(); } }`},

		// ---- extends: non-forwarding bodies ----
		{Code: `class A extends B { constructor(){} }`},
		{Code: `class A extends B { constructor(){ super('foo'); } }`},
		{Code: `class A extends B { constructor(foo, bar){ super(foo, bar, 1); } }`},
		{Code: `class A extends B { constructor(){ super(); doSomething(); } }`},
		{Code: `class A extends B { constructor(...args){ super(...args); doSomething(); } }`},

		// ---- extends with property/call expression ----
		{Code: `class A extends B.C { constructor() { super(foo); } }`},
		{Code: `class A extends B.C { constructor([a, b, c]) { super(...arguments); } }`},
		{Code: `class A extends B.C { constructor(a = f()) { super(...arguments); } }`},

		// ---- Fewer/different args than params ----
		{Code: `class A extends B { constructor(a, b, c) { super(a, b); } }`},
		{Code: `class A extends B { constructor(foo, bar){ super(foo); } }`},
		{Code: `class A extends B { constructor(test) { super(); } }`},
		{Code: `class A extends B { constructor() { foo; } }`},
		{Code: `class A extends B { constructor(foo, bar) { super(bar); } }`},

		// ---- Constructor declaration without body (TS overload / declare / abstract) ----
		{Code: `declare class A { constructor(options: any); }`},
		{Code: `class A { constructor(); }`},
		{Code: `abstract class A { constructor(); }`},
		{Code: `
class A {
  constructor(x: number);
  constructor(x: string);
  constructor(x: number | string) {
    doSomething(x);
  }
}`},

		// ---- Parameter properties (useful → skip) ----
		{Code: `class A { constructor(private name: string) {} }`},
		{Code: `class A { constructor(public name: string) {} }`},
		{Code: `class A { constructor(protected name: string) {} }`},
		{Code: `class A { constructor(readonly name: string) {} }`},
		{Code: `class A { constructor(public x: number, public y: number) {} }`},
		{Code: `class A { constructor(public name: string, age: number) {} }`},
		{Code: `class A { constructor(private readonly id: number) {} }`},
		{Code: `class A extends B { constructor(public name: string) { super(name); } }`},

		// ---- Access modifier on constructor (useful → skip) ----
		{Code: `class A { private constructor() {} }`},
		{Code: `class A { protected constructor() {} }`},
		{Code: `class A extends B { public constructor() {} }`},
		{Code: `class A extends B { public constructor() { super(); } }`},
		{Code: `class A extends B { public constructor(foo) { super(foo); } }`},
		{Code: `class A extends B { public constructor(foo) {} }`},
		{Code: `class A extends B { protected constructor(foo, bar) { super(bar); } }`},
		{Code: `class A extends B { private constructor(foo, bar) { super(bar); } }`},

		// ---- Decorator on params (useful → skip) ----
		{Code: `class A extends Object { constructor(@Foo foo: string) { super(foo); } }`},
		{Code: `class A extends Object { constructor(foo: string, @Bar() bar) { super(foo, bar); } }`},
		{Code: `class A { constructor(@Inject service: Service) {} }`},

		// ---- Body statements that are not a forwarding super() ----
		{Code: `class A extends B { constructor() { super.init(); } }`},
		{Code: `class A extends B { constructor() { void super(); } }`},
		{Code: `class A extends B { constructor() { super(), doSomething(); } }`},
		{Code: `class A extends B { constructor(x) { if (x) { super(x); } else { super(); } } }`},
		{Code: `class A extends B { constructor() { try { super(); } catch(e) {} } }`},

		// ---- Non-simple params (destructuring, defaults) that don't match ----
		{Code: `class A extends B { constructor({ x, y }) { super(...arguments); } }`},
		{Code: `class A extends B { constructor(a = 1) { super(a); } }`},
		// Rest with destructuring + super(...rest) where rest's binding pattern
		// isn't a plain identifier — isValidRestSpreadPair requires identifier
		// name pairing, so this stays valid.
		{Code: `class A extends B { constructor(...[x, y]) { super(...[x, y]); } }`},

		// ---- Type-assertion / call in super args ----
		{Code: `class A extends B { constructor(x) { super(x as any); } }`},
		{Code: `class A extends B { constructor(...args) { super(...args.slice(0)); } }`},
		{Code: `class A extends B { constructor(a, b) { super(a, ...b); } }`},
		{Code: `class A extends B { constructor(...args) { super(args); } }`},
		{Code: `class A extends B { constructor(x) { super(x, 'extra'); } }`},
		{Code: `class A extends B { constructor(a, b, c) { super(c, b, a); } }`},
		{Code: `class A extends B { constructor(a, ...rest) { super(a); } }`},

		// ---- Class expressions ----
		{Code: `const A = class {}`},
		{Code: `const A = class { constructor() { doSomething(); } }`},
		{Code: `const A = class extends B { constructor() { super(); doSomething(); } }`},

		// ---- Nested classes ----
		{Code: `class Outer { constructor() { class Inner { constructor() { doSomething(); } } } }`},
		{Code: `class Outer { method() { return class Inner extends Base { constructor() { super(); this.init(); } } } }`},
		{Code: `class Outer { constructor() { this.inner = class Inner { constructor() { doSomething(); } } } }`},

		// ---- Useful body logic ----
		{Code: `class A { constructor(x) { this.x = x; } }`},
		{Code: `class A extends B { constructor(x) { super(x); this.extra = true; } }`},
		{Code: `class A extends B { constructor() { super(); console.log('constructed'); } }`},
		{Code: `class A { constructor() { if (new.target === A) throw new Error(); } }`},

		// ---- implements only (no extends → no superClass) ----
		{Code: `class A implements Serializable { constructor() { this.init(); } }`},

		// ---- extends null with non-empty body ----
		{Code: `class A extends null { constructor() { doSomething(); } }`},

		// ---- Generic class ----
		{Code: `class A<T> extends B<T> { constructor(x: T) { super(x); this.data = x; } }`},

		// ---- Directive prologue ----
		{Code: `class A extends B { constructor() { "use strict"; super(); } }`},

		// ---- TypeScript "this" parameter in extending class (not simple forwarding) ----
		{Code: `class A extends B { constructor(this: Foo) { super(); } }`},

		// ---- Optional param with different body ----
		{Code: `class A extends B { constructor(x?: number) {} }`},

		// ---- Single-statement empty body (semicolon) ----
		{Code: `class A { constructor() { ; } }`},
		{Code: `class A { constructor() { return; } }`},

		// ---- Decorated param in extending class with forwarding ----
		{Code: `class A extends B { constructor(@D x) { super(x); } }`},

		// ---- 'static constructor' is a static method, not a constructor ----
		{Code: `class A { static constructor() {} }`},
		{Code: `class A extends B { static constructor() {} }`},

		// ---- get/set with name 'constructor' are accessors, not constructors ----
		{Code: `class A { get constructor() { return 1; } }`},
		{Code: `class A { set constructor(v) { this.v = v; } }`},

		// ---- Computed key ['constructor'] is a method, not a constructor ----
		{Code: `class A { ['constructor']() {} }`},

		// ---- Overload signatures + non-empty implementation ----
		{Code: `
class A {
  constructor(x: number);
  constructor(x: string);
  constructor(x: number | string) {
    this.x = x;
  }
}`},
	}, []rule_tester.InvalidTestCase{
		// ---- Basic invalid: empty constructor (full range assertion) ----
		{
			Code: `class A { constructor(){} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 1, Column: 11, EndLine: 1, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "class A {  }"},
				}},
			},
		},
		// Whitespace between `constructor` and `(` — end column still at end of keyword
		{
			Code: `class A { constructor     (){} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 1, Column: 11, EndLine: 1, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "class A {  }"},
				}},
			},
		},
		{
			Code: `class A { constructor     (){} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 1, Column: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "class A {  }"},
				}},
			},
		},
		{
			Code: `class A { 'constructor'(){} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 1, Column: 11, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "class A {  }"},
				}},
			},
		},

		// ---- extends: redundant super() ----
		{
			Code: `class A extends B { constructor() { super(); } }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 1, Column: 21, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "class A extends B {  }"},
				}},
			},
		},

		// ---- Parenthesized super call (paren wrapper must not hide the useless pattern) ----
		{
			Code: `class A extends B { constructor() { (super()); } }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 1, Column: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "class A extends B {  }"},
				}},
			},
		},
		// ---- Parenthesized super arg (single param identifier) ----
		{
			Code: `class A extends B { constructor(foo) { super((foo)); } }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 1, Column: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "class A extends B {  }"},
				}},
			},
		},
		// ---- Spread with parenthesized identifier inside ----
		{
			Code: `class A extends B { constructor(...args) { super(...(args)); } }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 1, Column: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "class A extends B {  }"},
				}},
			},
		},
		// ---- Spread of parenthesized `arguments` ----
		{
			Code: `class A extends B { constructor() { super(...(arguments)); } }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 1, Column: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "class A extends B {  }"},
				}},
			},
		},
		// ---- Nested parentheses ----
		{
			Code: `class A extends B { constructor() { (((super()))); } }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 1, Column: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "class A extends B {  }"},
				}},
			},
		},

		// ---- Body containing only comments (no statements) ----
		{
			Code: `class A { constructor() { /* noop */ } }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 1, Column: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "class A {  }"},
				}},
			},
		},

		// ---- ASI hazard: property (no trailing `;`) → useless ctor → computed-name member.
		//      Fix must emit `;` instead of removing outright.
		{
			Code: `
class A {
  foo = 'bar'
  constructor() { }
  [0]() { }
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 4, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  foo = 'bar'\n  ;\n  [0]() { }\n}"},
				}},
			},
		},
		// ---- Previous property ends with `;` → safe, plain remove.
		{
			Code: `
class A {
  foo = 'bar';
  constructor() { }
  [0]() { }
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 4, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  foo = 'bar';\n  \n  [0]() { }\n}"},
				}},
			},
		},
		// ---- Constructor is first member → previous token is `{` → safe.
		{
			Code: `
class A {
  constructor() { }
  [0]() { }
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 3, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  \n  [0]() { }\n}"},
				}},
			},
		},
		// ---- Previous member is a method (ends with `}`) → safe.
		{
			Code: `
class A {
  m() {}
  constructor() { }
  [0]() { }
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 4, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  m() {}\n  \n  [0]() { }\n}"},
				}},
			},
		},
		// ---- PropertyDeclaration with initializer ending in `}` still needs `;`
		//      (function expression would be member-accessed by `[0]()`).
		{
			Code: `
class A {
  foo = function () {}
  constructor() {}
  [0]() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 4, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  foo = function () {}\n  ;\n  [0]() {}\n}"},
				}},
			},
		},
		// ---- PropertyDeclaration with initializer ending in `]` still needs `;`.
		{
			Code: `
class A {
  foo = [1, 2]
  constructor() {}
  [0]() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 4, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  foo = [1, 2]\n  ;\n  [0]() {}\n}"},
				}},
			},
		},
		// ---- PropertyDeclaration with object-literal initializer still needs `;`.
		{
			Code: `
class A {
  foo = {}
  constructor() {}
  [0]() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 4, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  foo = {}\n  ;\n  [0]() {}\n}"},
				}},
			},
		},
		// ---- Previous accessor (`}` = block end, safe): no `;` needed.
		{
			Code: `
class A {
  get prop() { return 1; }
  constructor() {}
  [0]() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 4, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  get prop() { return 1; }\n  \n  [0]() {}\n}"},
				}},
			},
		},
		// ---- PropertyDeclaration with arrow-block initializer: arrow `}` terminates
		//      the arrow, so the following `[...]` cannot fuse — ESLint does NOT
		//      add `;`, and neither should we.
		{
			Code: `
class A {
  foo = () => {}
  constructor() {}
  [0]() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 4, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  foo = () => {}\n  \n  [0]() {}\n}"},
				}},
			},
		},
		// ---- PropertyDeclaration with async arrow-block initializer: same as arrow.
		{
			Code: `
class A {
  foo = async () => {}
  constructor() {}
  [0]() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 4, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  foo = async () => {}\n  \n  [0]() {}\n}"},
				}},
			},
		},
		// ---- PropertyDeclaration with arrow-expression initializer: ends in identifier
		//      → needs `;`.
		{
			Code: `
class A {
  foo = () => x
  constructor() {}
  [0]() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 4, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  foo = () => x\n  ;\n  [0]() {}\n}"},
				}},
			},
		},
		// ---- PropertyDeclaration with class-expression initializer: still needs `;`
		//      (`class{}[0]` is member access on the class expression).
		{
			Code: `
class A {
  foo = class {}
  constructor() {}
  [0]() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 4, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  foo = class {}\n  ;\n  [0]() {}\n}"},
				}},
			},
		},
		// ---- PropertyDeclaration with call-expression initializer: `)` ending → needs `;`.
		{
			Code: `
class A {
  foo = bar()
  constructor() {}
  [0]() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 4, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  foo = bar()\n  ;\n  [0]() {}\n}"},
				}},
			},
		},
		// ---- PropertyDeclaration with identifier initializer → needs `;`.
		{
			Code: `
class A {
  foo = bar
  constructor() {}
  [0]() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 4, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  foo = bar\n  ;\n  [0]() {}\n}"},
				}},
			},
		},
		// ---- PropertyDeclaration with template-literal initializer → needs `;`.
		{
			Code: "\nclass A {\n  foo = `t`\n  constructor() {}\n  [0]() {}\n}",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 4, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  foo = `t`\n  ;\n  [0]() {}\n}"},
				}},
			},
		},
		// ---- PropertyDeclaration with postfix `++` initializer: ASI fires
		//      unconditionally after `++`, so no explicit `;` needed.
		{
			Code: `
class A {
  foo = x++
  constructor() {}
  [0]() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 4, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  foo = x++\n  \n  [0]() {}\n}"},
				}},
			},
		},
		// ---- PropertyDeclaration with postfix `--` initializer: same as `++`.
		{
			Code: `
class A {
  foo = x--
  constructor() {}
  [0]() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 4, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  foo = x--\n  \n  [0]() {}\n}"},
				}},
			},
		},
		// ---- Postfix `++` on member access: `this.x++` — still AST kind postfix.
		{
			Code: `
class A {
  foo = this.x++
  constructor() {}
  [0]() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 4, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  foo = this.x++\n  \n  [0]() {}\n}"},
				}},
			},
		},
		// ---- Parenthesized `(x++)` — initializer kind is ParenthesizedExpression,
		//      not PostfixUnary — so needs `;` (matches ESLint's closing-paren path).
		{
			Code: `
class A {
  foo = (x++)
  constructor() {}
  [0]() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 4, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  foo = (x++)\n  ;\n  [0]() {}\n}"},
				}},
			},
		},
		// ---- Multiple stray `;` between members: tsgo treats extras as
		//      SemicolonClassElement; prev is a `;` → safe.
		{
			Code: `
class A {
  foo = 'bar'
  ;
  constructor() {}
  [0]() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 5, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  foo = 'bar'\n  ;\n  \n  [0]() {}\n}"},
				}},
			},
		},
		// ---- Nested useless constructor inside class-expression initializer of
		//      an outer class where the outer ctor is also useless with ASI risk.
		//      Both get reported; fixes apply independently.
		{
			Code: `
class Outer extends Base {
  foo = class Inner { constructor() {} }
  constructor() { super(); }
  [0]() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 3, Column: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass Outer extends Base {\n  foo = class Inner {  }\n  constructor() { super(); }\n  [0]() {}\n}"},
				}},
				{MessageId: "noUselessConstructor", Line: 4, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass Outer extends Base {\n  foo = class Inner { constructor() {} }\n  ;\n  [0]() {}\n}"},
				}},
			},
		},
		// ---- `x` + binary `+ 1` ending in digit is NOT postfix — still needs `;`.
		{
			Code: `
class A {
  foo = x + 1
  constructor() {}
  [0]() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 4, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  foo = x + 1\n  ;\n  [0]() {}\n}"},
				}},
			},
		},

		// ---- PropertyDeclaration with type annotation only, no initializer → needs `;`.
		{
			Code: `
class A {
  foo: string
  constructor() {}
  [0]() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 4, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  foo: string\n  ;\n  [0]() {}\n}"},
				}},
			},
		},

		// ---- Previous static block: safe (terminates at its own `}`).
		{
			Code: `
class A {
  static { doSomething(); }
  constructor() {}
  [0]() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 4, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  static { doSomething(); }\n  \n  [0]() {}\n}"},
				}},
			},
		},

		// ---- Previous set accessor (`}` = block end, safe): no `;` needed.
		{
			Code: `
class A {
  set prop(v) { this._v = v; }
  constructor() {}
  [0]() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 4, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  set prop(v) { this._v = v; }\n  \n  [0]() {}\n}"},
				}},
			},
		},
		// ---- Next member has a string-literal name → not computed, safe.
		{
			Code: `
class A {
  foo = 'bar'
  constructor() {}
  'key'() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 4, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  foo = 'bar'\n  \n  'key'() {}\n}"},
				}},
			},
		},
		// ---- Next member has a private-identifier name → not computed, safe.
		{
			Code: `
class A {
  foo = 'bar'
  constructor() {}
  #secret() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 4, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  foo = 'bar'\n  \n  #secret() {}\n}"},
				}},
			},
		},
		// ---- Next member is a TS IndexSignature `[key: string]: any` — also
		//      starts with `[`, so ASI fires; must insert `;`.
		{
			Code: `
class A {
  foo = 'bar'
  constructor() {}
  [key: string]: any;
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 4, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  foo = 'bar'\n  ;\n  [key: string]: any;\n}"},
				}},
			},
		},
		// ---- Next computed member is preceded by a decorator — first token
		//      after the constructor is `@`, not `[`, so ESLint does NOT add `;`.
		{
			Code: `
class A {
  foo = 'bar'
  constructor() {}
  @Dec
  [0]() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 4, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  foo = 'bar'\n  \n  @Dec\n  [0]() {}\n}"},
				}},
			},
		},
		// ---- Next member is `static [x]()` — first token is `static`, not `[`.
		{
			Code: `
class A {
  foo = 'bar'
  constructor() {}
  static [0]() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 4, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  foo = 'bar'\n  \n  static [0]() {}\n}"},
				}},
			},
		},
		// ---- Next member is `readonly [x] = 1` — first token is `readonly`.
		{
			Code: `
class A {
  foo = 'bar'
  constructor() {}
  readonly [x] = 1
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 4, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  foo = 'bar'\n  \n  readonly [x] = 1\n}"},
				}},
			},
		},
		// ---- Next member has a numeric-literal name → not computed, safe.
		{
			Code: `
class A {
  foo = 'bar'
  constructor() {}
  1() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 4, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  foo = 'bar'\n  \n  1() {}\n}"},
				}},
			},
		},

		// ---- Next member has a plain (non-computed) name → no ASI risk.
		{
			Code: `
class A {
  foo = 'bar'
  constructor() { }
  bar() { }
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 4, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  foo = 'bar'\n  \n  bar() { }\n}"},
				}},
			},
		},

		// ---- Public constructor with full end-range assertion (multi-line) ----
		{
			Code: `
class A {
  public constructor() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 3, Column: 3, EndLine: 3, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  \n}"},
				}},
			},
		},

		// ---- Overload signatures with empty implementation body ----
		{
			Code: `
class A {
  constructor(x: number);
  constructor(x: string);
  constructor(x: number | string) {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 5, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  constructor(x: number);\n  constructor(x: string);\n  \n}"},
				}},
			},
		},
		{
			Code: `class A extends B { constructor(foo){ super(foo); } }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 1, Column: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "class A extends B {  }"},
				}},
			},
		},
		{
			Code: `class A extends B { constructor(foo, bar){ super(foo, bar); } }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 1, Column: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "class A extends B {  }"},
				}},
			},
		},
		{
			Code: `class A extends B { constructor(...args){ super(...args); } }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 1, Column: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "class A extends B {  }"},
				}},
			},
		},
		{
			Code: `class A extends B.C { constructor() { super(...arguments); } }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 1, Column: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "class A extends B.C {  }"},
				}},
			},
		},
		{
			Code: `class A extends B { constructor(a, b, ...c) { super(...arguments); } }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 1, Column: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "class A extends B {  }"},
				}},
			},
		},
		{
			Code: `class A extends B { constructor(a, b, ...c) { super(a, b, ...c); } }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 1, Column: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "class A extends B {  }"},
				}},
			},
		},

		// ---- Public constructor without superClass (useless) ----
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

		// ---- Multi-line basic ----
		{
			Code: `
class A {
  constructor() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 3, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "\nclass A {\n  \n}"},
				}},
			},
		},

		// ---- Class expressions ----
		{
			Code: `const A = class { constructor() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 1, Column: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "const A = class {  }"},
				}},
			},
		},
		{
			Code: `const A = class MyClass { constructor() {} }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 1, Column: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "const A = class MyClass {  }"},
				}},
			},
		},
		{
			Code: `const A = class extends B { constructor() { super(); } }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUselessConstructor", Line: 1, Column: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "removeConstructor", Output: "const A = class extends B {  }"},
				}},
			},
		},

		// ---- Nested classes: inner class with useless constructor ----
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

		// ---- Outer + inner both useless → two errors ----
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

		// ---- Complex extends expressions ----
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

		// ---- Type-annotated params (annotations don't make it useful) ----
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

		// ---- Abstract class with empty/redundant body ----
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

		// ---- implements only (no extends) ----
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

		// ---- extends null ----
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

		// ---- Generic class ----
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

		// ---- export default class ----
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

		// ---- Many params forwarding to super ----
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

		// ---- Rest with destructuring + super(...arguments): ESLint treats any
		//      RestElement as "simple" so spread-arguments forwarding still
		//      fires.
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

		// ---- Only rest param forwarding ----
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

		// ---- `this` parameter in non-extended class: empty body → useless ----
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
	})
}

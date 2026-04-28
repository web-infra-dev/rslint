package no_use_before_define

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestEdgeCases covers missing scenarios from the ESLint/typescript-eslint test suites
// and systematically enumerates edge cases across class definitions, scoping, and nesting.
func TestEdgeCases(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUseBeforeDefineRule,

		// =====================================================================
		// VALID cases
		// =====================================================================
		[]rule_tester.ValidTestCase{

			// ----- Class body: method/getter/setter bodies are separate execution context -----
			{Code: `class C { method() { C; } }`},
			{Code: `class C { static method() { C; } }`},
			{Code: `class C { get x() { return C; } }`},
			{Code: `class C { set x(v: any) { C; } }`},

			// ----- Class body: instance field with arrow (body not evaluated during init) -----
			{Code: `class C { field = () => C; }`},
			{Code: `class C { field = class extends C {}; }`},

			// ----- Class body: static field / static block (runs after class binding) -----
			{Code: `class C { static field = C; }`},
			{Code: `class C { static { C; } }`},
			{Code: `class C { static field = class extends C {}; }`},
			{Code: `class C { static field = class { [C as any](){} }; }`},

			// ----- let/const before class static block — same execution context -----
			{Code: `let a = 1; class C { static { a; } }`},
			{Code: `class C { static { let a: any; a; } }`},

			// ----- Class expression: method body -----
			{Code: `const C = class { method() { C; } };`},
			{Code: `(class C { method() { C; } });`},

			// ----- Class expression: static field referencing outer let -----
			// Static field initializer runs during class definition, which is during
			// the initialization of `const C = ...`, but the class binding for named
			// class expressions is available inside the class.
			{Code: `(class C { static field = C; });`},

			// ----- Superclass method bodies referencing the derived class -----
			{Code: `class C extends (class { method() { C; } }) {}`},

			// ----- Superclass instance field referencing derived class (separate context) -----
			{Code: `class C extends (class { field = C; }) {}`},

			// ----- Class instance field arrow function -----
			{Code: `class C { field = () => C; }`},

			// ----- Cross-scope class reference with classes:false -----
			{
				Code:    `function foo() { new A(); } class A {}`,
				Options: map[string]interface{}{"classes": false},
			},

			// ----- Cross-scope variable with variables:false -----
			{
				Code:    `function foo() { bar; } var bar: any;`,
				Options: map[string]interface{}{"variables": false},
			},
			{
				Code:    `var foo = () => bar; var bar: any;`,
				Options: map[string]interface{}{"variables": false},
			},
			{
				Code:    `class C { static { () => foo; } } let foo: any;`,
				Options: map[string]interface{}{"variables": false},
			},

			// ----- typedefs:false + ignoreTypeReferences:false -----
			{
				Code:    `var x: Foo = {} as any; interface Foo {}`,
				Options: map[string]interface{}{"typedefs": false, "ignoreTypeReferences": false},
			},
			{
				Code:    `var x: Foo = {} as any; interface Foo {}`,
				Options: map[string]interface{}{"typedefs": false},
			},
			{
				Code:    `let myVar: MyString; type MyString = string;`,
				Options: map[string]interface{}{"typedefs": false, "ignoreTypeReferences": false},
			},

			// ----- ignoreTypeReferences covers heritage in type-only positions -----
			// Interface `extends` is a pure type position.
			{Code: `interface A extends B {} interface B {}`},
			{Code: `interface A extends B<C> {} interface B<T> { x: T; } interface C {}`},
			{Code: `interface A extends ns.B {} namespace ns { export interface B {} }`},
			{Code: `interface A extends ns.sub.B {} namespace ns { export namespace sub { export interface B {} } }`},
			{Code: `interface A extends B, C {} interface B {} interface C {}`},
			// Class `implements` is a pure type position (single and multiple, generic, qualified).
			{Code: `class A implements B { x: number = 1; } interface B { x: number; }`},
			{Code: `class A implements ns.B { x = 1; } namespace ns { export interface B { x: number; } }`},
			{Code: `class A implements B, C { x = 1; y = 2; } interface B { x: number; } interface C { y: number; }`},
			{Code: `class A implements B<number> { x!: number; } interface B<T> { x: T; }`},
			// Class expression + implements.
			{Code: `const Cls = class implements B { x = 1; }; interface B { x: number; }`},
			// Class extends + implements in same declaration (class extends is NOT ignored,
			// so put declaration first).
			{Code: `class Base {} interface I {} class D extends Base implements I {}`},
			// Heritage with type argument that itself references later type.
			{Code: `interface A extends B<C> {} interface B<T> {} type C = number;`},
			// Class implements where the implemented interface extends a later one.
			{Code: `class A implements B { x = 1 } interface B extends C { x: number } interface C {}`},
			// Explicit ignoreTypeReferences:false but still valid because B is defined before.
			{
				Code:    `interface B {} interface A extends B {}`,
				Options: map[string]interface{}{"ignoreTypeReferences": false},
			},

			// ----- Optional chaining with declared variables -----
			{Code: `const updatedAt = (null as any)?.updatedAt;`},
			{Code: `function f() { return function t() {}; } f()?.();`},
			{Code: `var a = { b: 5 }; alert(a?.b);`},

			// ----- allowNamedExports with TS-specific declarations -----
			{
				Code:    `export { Foo, baz }; enum Foo { BAR } let baz: Enum; enum Enum {}`,
				Options: map[string]interface{}{"allowNamedExports": true},
			},

			// ----- Decorators with classes:false -----
			{
				Code: `
@Directive({
  selector: '[test]',
  providers: [{ useExisting: MyClass }],
})
export class MyClass implements Validator {}
`,
				Options: map[string]interface{}{"classes": false},
			},

			// ----- Constructor with default parameter using this -----
			{Code: `
class A {
  printerName: string = '';
  constructor(printName: string) {
    this.printerName = printName;
  }
  openPort(printerName = this.printerName) {
    return printerName;
  }
}
`},

			// ----- Class body self-references (typescript-eslint considers the class -----
			// ----- name as defined before the body in source order)               -----
			{Code: `class C extends C {}`},
			{Code: `class C { [C as any]() {} }`},
			{Code: `(class C { [C as any]() {} });`},
			{Code: `class C { [C as any]: any; }`},
			{Code: `class C { static [C as any]() {} }`},
			{Code: `const C = class { static field = C; };`},
			{Code: `const C = class { static { C; } };`},
			{Code: `(class C extends C {});`},

			// ----- Type predicate (value is Type) -----
			{Code: `type T = (value: unknown) => value is string;`},

			// ----- JSX: implicit React reference should not trigger -----
			{Code: `import * as React from 'react'; <div />;`, Tsx: true},
			{Code: `import React from 'react'; <div />;`, Tsx: true},
			{Code: `const React = require('react'); <div />;`, Tsx: true},
			{Code: `import { h } from 'preact'; <div />;`, Tsx: true},

			// ----- JSX: component defined before use -----
			{Code: `const App = () => <div/>; <App />;`, Tsx: true},
			{Code: `let Foo: any, Bar: any; <Foo><Bar /></Foo>;`, Tsx: true},
			{Code: `function App() { return <div/> } <App />;`, Tsx: true},
			// Function component used before define — valid with functions:false
			{
				Code:    `<App />; function App() { return <div/> }`,
				Tsx:     true,
				Options: map[string]interface{}{"functions": false},
			},

			// ----- Global augmentation -----
			{Code: `
(globalThis as any).foo = true;
declare global {
  namespace NodeJS {
    interface Global {
      foo?: boolean;
    }
  }
}
`},
		},

		// =====================================================================
		// INVALID cases
		// =====================================================================
		[]rule_tester.InvalidTestCase{

			// ----- Class extends another class declared after -----
			{
				Code: `class C extends D {} class D {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine", Line: 1, Column: 17},
				},
			},

			// ----- Class extends is a VALUE reference — ignoreTypeReferences must not suppress it -----
			{
				Code:    `class C extends D {} class D {}`,
				Options: map[string]interface{}{"ignoreTypeReferences": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine", Line: 1, Column: 17},
				},
			},

			// ----- Heritage in type-only position with ignoreTypeReferences:false must report -----
			{
				Code:    `interface A extends B {} interface B {}`,
				Options: map[string]interface{}{"ignoreTypeReferences": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine", Line: 1, Column: 21},
				},
			},
			{
				Code:    `class A implements B { x: number = 1 } interface B { x: number }`,
				Options: map[string]interface{}{"ignoreTypeReferences": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine", Line: 1, Column: 20},
				},
			},
			// Qualified-name heritage: both `ns` and `B` resolve to later declarations.
			{
				Code:    `interface A extends ns.B {} namespace ns { export interface B {} }`,
				Options: map[string]interface{}{"ignoreTypeReferences": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine", Line: 1, Column: 21},
					{MessageId: "noUseBeforeDefine", Line: 1, Column: 24},
				},
			},

			// ----- Class with computed key referencing variable declared after -----
			{
				Code: `class C { [a as any]() {} } let a: any;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine", Line: 1, Column: 12},
				},
			},

			// ----- Static field referencing variable declared after -----
			{
				Code: `class C { static field = a; } let a: any;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine", Line: 1, Column: 26},
				},
			},

			// ----- Static block referencing variable declared after -----
			{
				Code: `class C { static { a; } } let a: any;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine", Line: 1, Column: 20},
				},
			},

			// ----- Static field referencing class declared after -----
			{
				Code: `class C { static field = D; } class D {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine", Line: 1, Column: 26},
				},
			},

			// ----- variables:false still catches same-scope -----
			{
				Code:    `foo; var foo: any;`,
				Options: map[string]interface{}{"variables": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine", Line: 1, Column: 1},
				},
			},

			// ----- ignoreTypeReferences:false with type annotation -----
			{
				Code:    `let var1: StringOrNumber; type StringOrNumber = string | number;`,
				Options: map[string]interface{}{"ignoreTypeReferences": false, "typedefs": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine"},
				},
			},

			// ----- Optional chaining before define -----
			{
				Code: `f()?.(); function f() { return function t() {}; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine", Line: 1, Column: 1},
				},
			},
			{
				Code: `alert(a?.b); var a = { b: 5 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine", Line: 1, Column: 7},
				},
			},

			// ----- JSX: component used before define -----
			{
				Code: `<App />; const App = () => <div />;`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine"},
				},
			},
			{
				Code: `function render() { return <Widget /> } const Widget = () => <span />;`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine"},
				},
			},
			{
				Code: `<Foo.Bar />; const Foo = { Bar: () => <div/> };`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine"},
				},
			},

			// ----- export default still reports with allowNamedExports -----
			{
				Code:    `export default a; const a = 1;`,
				Options: map[string]interface{}{"allowNamedExports": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine"},
				},
			},

			// ----- same-scope TDZ: classes:false + functions:false still catches class declaration -----
			{
				Code:    `new A(); class A {}`,
				Options: map[string]interface{}{"functions": false, "classes": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine"},
				},
			},

			// ----- export const / export function / export class body still reports with allowNamedExports -----
			{
				Code:    `export const foo = a; const a = 1;`,
				Options: map[string]interface{}{"allowNamedExports": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine"},
				},
			},
			{
				Code:    `export function foo() { return a; } const a = 1;`,
				Options: map[string]interface{}{"allowNamedExports": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine"},
				},
			},
			{
				Code:    `export class C { foo() { return a; } } const a = 1;`,
				Options: map[string]interface{}{"allowNamedExports": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine"},
				},
			},

			// ----- Multiple export specifiers with some before define -----
			{
				Code:    `export { Foo, baz }; enum Foo { BAR } let baz: Enum; enum Enum {}`,
				Options: map[string]interface{}{"allowNamedExports": false, "ignoreTypeReferences": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine"},
					{MessageId: "noUseBeforeDefine"},
				},
			},

			// ----- Deeply nested: function inside arrow inside class method -----
			{
				Code: `
const x = (() => {
  return function inner() {
    return a;
  };
})();
const a = 1;
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine"},
				},
			},

			// ----- let/const in block scope -----
			{
				Code: `{ a; let a = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine", Line: 1, Column: 3},
				},
			},
			{
				Code: `if (true) { function foo() { a; } let a: any; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine"},
				},
			},

			// ----- Array destructuring with earlier ref -----
			{
				Code: `var [b = a, a] = {} as any;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine", Line: 1, Column: 10},
				},
			},

			// ----- nofunc: var function expression still caught -----
			{
				Code:    `a(); var a = function () {};`,
				Options: map[string]interface{}{"functions": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine", Line: 1, Column: 1},
				},
			},

			// ----- classes:false with var = class (still caught, it's a variable) -----
			{
				Code:    `new A(); var A = class {};`,
				Options: map[string]interface{}{"classes": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine"},
				},
			},
			{
				Code:    `function foo() { new A(); } var A = class {};`,
				Options: map[string]interface{}{"classes": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine"},
				},
			},

			// =========== For-loop TDZ ===========

			// for-loop initializer self-reference
			{
				Code: `for (let x = x;;) {} let y: any;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine"},
				},
			},
			// for-in referencing later let
			{
				Code: `for (let x in xs) {} let xs: any;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine"},
				},
			},
			// for-of referencing later let
			{
				Code: `for (let x of xs) {} let xs: any;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine"},
				},
			},

			// =========== Cross-variable: class member referencing later declaration ===========
			// These use DEFAULT options (not variables:false / classes:false) because
			// the typescript-eslint rule treats computed keys as cross-scope references
			// (inside method/property scope), so option=false would suppress them.

			{
				Code: `class C { [a as any]() {} } let a: any;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine"},
				},
			},
			{
				Code: `class C { static [a as any](){} } let a: any;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine"},
				},
			},
			{
				Code: `class C { [a as any]: any; } let a: any;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine"},
				},
			},
			{
				Code: `class C { [a as any] = 1; } let a: any;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine"},
				},
			},
			{
				Code: `class C { static [a as any]: any; } let a: any;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine"},
				},
			},
			{
				Code: `class C { static [a as any] = 1; } let a: any;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine"},
				},
			},

			// =========== Static block/field with later declarations ===========

			{
				Code: `class C { static { D; } } class D {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine"},
				},
			},
			{
				Code: `class C { static { (class extends D {}); } } class D {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine"},
				},
			},
			{
				Code: `class C { static { (class { [a as any](){} }); } } let a: any;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine"},
				},
			},
			{
				Code: `class C { static { (class { static field = a; }); } } let a: any;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine"},
				},
			},

			// =========== Superclass / nested static field with later declarations ===========

			{
				Code: `class C extends (class { [a as any](){} }) {} let a: any;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine"},
				},
			},
			{
				Code: `class C extends (class { static field = a; }) {} let a: any;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine"},
				},
			},
			{
				Code: `class C { static field = class extends D {}; } class D {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine"},
				},
			},
			{
				Code: `class C { static field = class { [a as any](){} }; } let a: any;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine"},
				},
			},
			{
				Code: `class C { static field = class { static field = a; }; } let a: any;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noUseBeforeDefine"},
				},
			},
		},
	)
}

package no_use_before_define

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUseBeforeDefineRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUseBeforeDefineRule, []rule_tester.ValidTestCase{
		// Type declarations before use
		{
			Code: `
type foo = 1;
const x: foo = 1;
`,
		},
		{
			Code: `
type foo = 1;
type bar = foo;
`,
		},
		{
			Code: `
interface Foo {}
const x: Foo = {};
`,
		},
		// Global object method call — no local variable involved
		{Code: `Object.hasOwnProperty.call(a);`},
		// Function using arguments keyword — no forward reference
		{
			Code: `
function a() {
  alert(arguments);
}
`,
		},
		// Variable declared before use
		{
			Code: `
var a = 10;
alert(a);
`,
		},
		{
			Code: `
function b(a: number) {
  alert(a);
}
`,
		},
		// Declare statements
		{
			Code: `declare function a(): void;`,
		},
		{
			Code: `
declare class a {
  foo(): void;
}
`,
		},
		// Optional chaining
		{
			Code: `const updatedAt = (null as any)?.updatedAt;`,
		},
		// "nofunc" option
		{
			Code: `
a();
function a() {
  alert(arguments);
}
`,
			Options: map[string]interface{}{"functions": false},
		},
		// Arrow function IIFE
		{
			Code: `
(() => {
  var a = 42;
  alert(a);
})();
`,
		},
		// Catch clause
		{
			Code: `
a();
try {
  throw new Error();
} catch (a) {}
`,
		},
		// Class declared before use
		{
			Code: `
class A {}
new A();
`,
		},
		// Sequential variable declarations
		{
			Code: `
var a = 0, b = a;
`,
		},
		// Destructuring with sequential defaults (b uses a, which is already bound)
		{Code: `var { a = 0, b = a } = {} as any;`},
		{Code: `var [a = 0, b = a] = {} as any;`},
		// Self-referencing function
		{
			Code: `
function foo() {
  foo();
}
`,
		},
		{
			Code: `
var foo = function () {
  foo();
};
`,
		},
		// For-in/for-of with declared variable
		{
			Code: `
var a: any;
for (a in a) {}
`,
		},
		{
			Code: `
var a: any;
for (a of a) {}
`,
		},
		// Block-level bindings
		{
			Code: `
'use strict';
a();
{
  function a() {}
}
`,
		},
		// a() is in outer scope; block-scoped let function expression doesn't conflict
		{
			Code: `
a();
{
  let a = function () {};
}
`,
		},
		{
			Code: `
'use strict';
{
  a();
  function a() {}
}
`,
			Options: map[string]interface{}{"functions": false},
		},
		{
			Code: `
switch (foo) {
  case 1: {
    a();
  }
  default: {
    let a: any;
  }
}
`,
		},
		// Object style options
		{
			Code: `
a();
function a() {
  alert(arguments);
}
`,
			Options: map[string]interface{}{"functions": false},
		},
		{
			Code: `
function foo() {
  new A();
}
class A {}
`,
			Options: map[string]interface{}{"classes": false},
		},
		// "variables" option
		{
			Code: `
function foo() {
  bar;
}
var bar: any;
`,
			Options: map[string]interface{}{"variables": false},
		},
		{
			Code: `
var foo = () => bar;
var bar: any;
`,
			Options: map[string]interface{}{"variables": false},
		},
		// "typedefs" option
		{
			Code: `
var x: Foo = 2 as any;
type Foo = string | number;
`,
			Options: map[string]interface{}{"typedefs": false},
		},
		// ignoreTypeReferences
		{
			Code: `
interface Bar {
  type: typeof Foo;
}
const Foo = 2;
`,
			Options: map[string]interface{}{"ignoreTypeReferences": true},
		},
		{
			Code: `
interface Bar {
  type: typeof Foo.FOO;
}
class Foo {
  public static readonly FOO = '';
}
`,
			Options: map[string]interface{}{"ignoreTypeReferences": true},
		},
		{
			Code: `
interface Bar {
  type: typeof Foo.Bar.Baz;
}
const Foo = {
  Bar: {
    Baz: 1,
  },
};
`,
			Options: map[string]interface{}{"ignoreTypeReferences": true},
		},
		// Interface with same name as later variable (issue #435)
		{
			Code: `
interface Foo {
  bar: string;
}
const bar = 'blah';
`,
		},
		// Interface members do not conflict with later let/export/namespace (issue #141)
		{
			Code: `
interface ITest {
  first: boolean;
  second: string;
  third: boolean;
}
let first = () => console.log('first');
export let second = () => console.log('second');
export namespace Third {
  export let third = () => console.log('third');
}
`,
		},
		// typeof on parameter member (issue #550)
		{
			Code: `
function test(file: Blob) {
  const slice: typeof file.slice =
    file.slice || (file as any).webkitSlice || (file as any).mozSlice;
  return slice;
}
`,
		},
		// Enums with enums: false
		{
			Code: `
function foo(): Foo {
  return Foo.FOO;
}
enum Foo { FOO }
`,
			Options: map[string]interface{}{"enums": false},
		},
		{
			Code: `
let foo: Foo;
enum Foo { FOO }
`,
			Options: map[string]interface{}{"enums": false},
		},
		{
			Code: `
class Test {
  foo(args: Foo): Foo {
    return Foo.FOO;
  }
}
enum Foo { FOO }
`,
			Options: map[string]interface{}{"enums": false},
		},
		// allowNamedExports
		{
			Code: `
export { a };
const a = 1;
`,
			Options: map[string]interface{}{"allowNamedExports": true},
		},
		{
			Code: `
export { a as b };
const a = 1;
`,
			Options: map[string]interface{}{"allowNamedExports": true},
		},
		{
			Code: `
export { Foo };
enum Foo { BAR }
`,
			Options: map[string]interface{}{"allowNamedExports": true},
		},
		{
			Code: `
export { a, b };
let a: any, b: any;
`,
			Options: map[string]interface{}{"allowNamedExports": true},
		},
		{
			Code: `
export { a };
var a: any;
`,
			Options: map[string]interface{}{"allowNamedExports": true},
		},
		{
			Code: `
export { f };
function f() {}
`,
			Options: map[string]interface{}{"allowNamedExports": true},
		},
		{
			Code: `
export { C };
class C {}
`,
			Options: map[string]interface{}{"allowNamedExports": true},
		},
		{
			Code: `
export { Foo };
namespace Foo {
  export let bar = () => console.log('bar');
}
`,
			Options: map[string]interface{}{"allowNamedExports": true},
		},
		// Decorators
		{
			Code: `
@Directive({
  selector: '[rcCidrIpPattern]',
  providers: [
    {
      provide: NG_VALIDATORS,
      useExisting: CidrIpPatternDirective,
      multi: true,
    },
  ],
})
export class CidrIpPatternDirective implements Validator {}
`,
		},
		// satisfies / as with ignoreTypeReferences: false
		{
			Code: `
const obj = {
  foo: 'foo-value',
  bar: 'bar-value',
} satisfies {
  [key in 'foo' | 'bar']: ` + "`${key}-value`" + `;
};
`,
			Options: map[string]interface{}{"ignoreTypeReferences": false},
		},
		{
			Code: `
const obj = {
  foo: 'foo-value',
  bar: 'bar-value',
} as {
  [key in 'foo' | 'bar']: ` + "`${key}-value`" + `;
};
`,
			Options: map[string]interface{}{"ignoreTypeReferences": false},
		},
		// Nested object with as — ignoreTypeReferences: false (no forward ref, still valid)
		{
			Code: `
const obj = {
  foo: {
    foo: 'foo',
  } as {
    [key in 'foo' | 'bar']: key;
  },
};
`,
			Options: map[string]interface{}{"ignoreTypeReferences": false},
		},
		// Namespace alias
		{
			Code: `
namespace A.X.Y {}
import Z = A.X.Y;
const X = 23;
`,
		},
		// Extended namespace alias (non-dotted form)
		{
			Code: `
namespace A {
  export namespace X {
    export namespace Y {
      export const foo = 40;
    }
  }
}
import Z = A.X.Y;
const X = 23;
`,
		},
		// typeof in satisfies with ignoreTypeReferences: true
		{
			Code: `
const foo = {
  bar: 'bar',
} satisfies {
  bar: typeof baz;
};
const baz = '';
`,
			Options: map[string]interface{}{"ignoreTypeReferences": true},
		},
	}, []rule_tester.InvalidTestCase{
		// Basic use before define
		{
			Code: `
a++;
var a = 19;
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 2, Column: 1},
			},
		},
		{
			Code: `
a();
var a = function () {};
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 2, Column: 1},
			},
		},
		{
			Code: `
alert(a[1]);
var a = [1, 3];
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 2, Column: 7},
			},
		},
		// Function with nested use before define
		{
			Code: `
a();
function a() {
  alert(b);
  var b = 10;
  a();
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 2, Column: 1},
				{MessageId: "noUseBeforeDefine", Line: 4, Column: 9},
			},
		},
		// nofunc still catches var assigned function expression
		{
			Code: `
a();
var a = function () {};
`,
			Options: map[string]interface{}{"functions": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 2, Column: 1},
			},
		},
		// Arrow function
		{
			Code: `
(() => {
  alert(a);
  var a = 42;
})();
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 3, Column: 9},
			},
		},
		// Arrow function calling function before define
		{
			Code: `
(() => a())();
function a() {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 2, Column: 8},
			},
		},
		// Catch clause with var
		{
			Code: `
a();
try {
  throw new Error();
} catch (foo) {
  var a: any;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 2, Column: 1},
			},
		},
		// Arrow expression
		{
			Code: `
var f = () => a;
var a: any;
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 2, Column: 15},
			},
		},
		// Class before define
		{
			Code: `
new A();
class A {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 2, Column: 5},
			},
		},
		{
			Code: `
function foo() {
  new A();
}
class A {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 3, Column: 7},
			},
		},
		{
			Code: `
new A();
var A = class {};
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 2, Column: 5},
			},
		},
		{
			Code: `
function foo() {
  new A();
}
var A = class {};
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 3, Column: 7},
			},
		},
		// Block-level bindings
		{
			Code: `
a++;
{
  var a: any;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 2, Column: 1},
			},
		},
		{
			Code: `
'use strict';
{
  a();
  function a() {}
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 4, Column: 3},
			},
		},
		{
			Code: `
{
  a;
  let a = 1;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 3, Column: 3},
			},
		},
		{
			Code: `
switch (foo) {
  case 1:
    a();
  default:
    let a: any;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 4, Column: 5},
			},
		},
		// Object style options — var assigned function still caught
		{
			Code: `
a();
var a = function () {};
`,
			Options: map[string]interface{}{"classes": false, "functions": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 2, Column: 1},
			},
		},
		// classes: false still catches var assigned class and same-scope class
		{
			Code: `
new A();
var A = class {};
`,
			Options: map[string]interface{}{"classes": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 2, Column: 5},
			},
		},
		{
			Code: `
function foo() {
  new A();
}
var A = class {};
`,
			Options: map[string]interface{}{"classes": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 3, Column: 7},
			},
		},
		// Self-referencing initializer
		{
			Code: `var a = a;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 1, Column: 9},
			},
		},
		{
			Code: `let a = a + b;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 1, Column: 9},
			},
		},
		{
			Code: `const a = foo(a);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 1, Column: 15},
			},
		},
		{
			Code: `function foo(a = a) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 1, Column: 18},
			},
		},
		{
			Code: `var { a = a } = [] as any;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 1, Column: 11},
			},
		},
		{
			Code: `var [a = a] = [] as any;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 1, Column: 10},
			},
		},
		{
			Code: `var { b = a, a } = {} as any;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 1, Column: 11},
			},
		},
		{
			Code: `var [b = a, a] = {} as any;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 1, Column: 10},
			},
		},
		{
			Code: `var { a = 0 } = a;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 1, Column: 17},
			},
		},
		{
			Code: `var [a = 0] = a;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 1, Column: 15},
			},
		},
		// For-in/for-of with inline declaration
		{
			Code: `
for (var a in a) {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 2, Column: 15},
			},
		},
		{
			Code: `
for (var a of a) {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 2, Column: 15},
			},
		},
		// ignoreTypeReferences: false
		{
			Code: `
interface Bar {
  type: typeof Foo;
}
const Foo = 2;
`,
			Options: map[string]interface{}{"ignoreTypeReferences": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 3, Column: 16},
			},
		},
		{
			Code: `
interface Bar {
  type: typeof Foo.FOO;
}
class Foo {
  public static readonly FOO = '';
}
`,
			Options: map[string]interface{}{"ignoreTypeReferences": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 3, Column: 16},
			},
		},
		{
			Code: `
interface Bar {
  type: typeof Foo.Bar.Baz;
}
const Foo = {
  Bar: {
    Baz: 1,
  },
};
`,
			Options: map[string]interface{}{"ignoreTypeReferences": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 3, Column: 16},
			},
		},
		// variables: false still catches same-scope
		{
			Code: `
function foo() {
  bar;
  var bar = 1;
}
var bar: any;
`,
			Options: map[string]interface{}{"variables": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 3, Column: 3},
			},
		},
		// Enum references
		{
			Code: `
class Test {
  foo(args: Foo): Foo {
    return Foo.FOO;
  }
}
enum Foo { FOO }
`,
			Options: map[string]interface{}{"enums": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 4, Column: 12},
			},
		},
		{
			Code: `
function foo(): Foo {
  return Foo.FOO;
}
enum Foo { FOO }
`,
			Options: map[string]interface{}{"enums": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 3, Column: 10},
			},
		},
		{
			Code: `
const foo = Foo.Foo;
enum Foo { FOO }
`,
			Options: map[string]interface{}{"enums": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 2, Column: 13},
			},
		},
		// Named exports (default allowNamedExports: false)
		{
			Code: `
export { a };
const a = 1;
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 2, Column: 10},
			},
		},
		// Empty object options — same as default
		{
			Code: `
export { a };
const a = 1;
`,
			Options: map[string]interface{}{},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 2, Column: 10},
			},
		},
		// "nofunc" string option — allowNamedExports still defaults to false
		{
			Code: `
export { a };
const a = 1;
`,
			Options: "nofunc",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 2, Column: 10},
			},
		},
		// export var before define
		{
			Code: `
export { a };
var a: any;
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 2, Column: 10},
			},
		},
		{
			Code: `
export { a as b };
const a = 1;
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 2, Column: 10},
			},
		},
		{
			Code: `
export { a, b };
let a: any, b: any;
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 2, Column: 10},
				{MessageId: "noUseBeforeDefine", Line: 2, Column: 13},
			},
		},
		{
			Code: `
export { f };
function f() {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 2, Column: 10},
			},
		},
		{
			Code: `
export { C };
class C {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 2, Column: 10},
			},
		},
		// Function call before define
		{
			Code: `
f();
function f() {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 2, Column: 1},
			},
		},
		{
			Code: `
alert(a);
var a = 10;
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 2, Column: 7},
			},
		},
		// Export enum/namespace before define
		{
			Code: `
export { Foo };
enum Foo { BAR }
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 2, Column: 10},
			},
		},
		{
			Code: `
export { Foo };
namespace Foo {
  export let bar = () => console.log('bar');
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 2, Column: 10},
			},
		},
		// satisfies with ignoreTypeReferences: false
		{
			Code: `
const foo = {
  bar: 'bar',
} satisfies {
  bar: typeof baz;
};
const baz = '';
`,
			Options: map[string]interface{}{"ignoreTypeReferences": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noUseBeforeDefine", Line: 5, Column: 15},
			},
		},
	})
}

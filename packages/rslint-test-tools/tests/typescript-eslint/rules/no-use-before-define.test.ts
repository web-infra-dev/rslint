import { RuleTester } from '@typescript-eslint/rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-use-before-define', {
  valid: [
    // Type declarations before use
    `
type foo = 1;
const x: foo = 1;
    `,
    `
type foo = 1;
type bar = foo;
    `,
    `
interface Foo {}
const x: Foo = {};
    `,
    // Variable declared before use
    `
var a = 10;
alert(a);
    `,
    `
function b(a: number) {
  alert(a);
}
    `,
    // Declare statements
    'declare function a(): void;',
    `
declare class a {
  foo(): void;
}
    `,
    // nofunc option
    {
      code: `
a();
function a() {
  alert(arguments);
}
      `,
      options: [{ functions: false }],
    },
    // Arrow function IIFE
    `
(() => {
  var a = 42;
  alert(a);
})();
    `,
    // Catch clause
    `
a();
try {
  throw new Error();
} catch (a) {}
    `,
    // Class declared before use
    `
class A {}
new A();
    `,
    // Self-referencing function
    `
function foo() {
  foo();
}
    `,
    `
var foo = function () {
  foo();
};
    `,
    // For-in/for-of with declared variable
    `
var a: any;
for (a in a) {}
    `,
    `
var a: any;
for (a of a) {}
    `,
    // Block-level bindings
    {
      code: `
'use strict';
a();
{
  function a() {}
}
      `,
    },
    {
      code: `
'use strict';
{
  a();
  function a() {}
}
      `,
      options: [{ functions: false }],
    },
    // Object style options
    {
      code: `
a();
function a() {
  alert(arguments);
}
      `,
      options: [{ functions: false }],
    },
    {
      code: `
function foo() {
  new A();
}
class A {}
      `,
      options: [{ classes: false }],
    },
    // "variables" option
    {
      code: `
function foo() {
  bar;
}
var bar: any;
      `,
      options: [{ variables: false }],
    },
    // "typedefs" option
    {
      code: `
var x: Foo = 2 as any;
type Foo = string | number;
      `,
      options: [{ typedefs: false }],
    },
    // ignoreTypeReferences
    {
      code: `
interface Bar {
  type: typeof Foo;
}
const Foo = 2;
      `,
      options: [{ ignoreTypeReferences: true }],
    },
    {
      code: `
interface Bar {
  type: typeof Foo.FOO;
}
class Foo {
  public static readonly FOO = '';
}
      `,
      options: [{ ignoreTypeReferences: true }],
    },
    // Interface with same name as later variable
    `
interface Foo {
  bar: string;
}
const bar = 'blah';
    `,
    // Enums with enums: false
    {
      code: `
function foo(): Foo {
  return Foo.FOO;
}
enum Foo { FOO }
      `,
      options: [{ enums: false }],
    },
    // allowNamedExports
    {
      code: `
export { a };
const a = 1;
      `,
      options: [{ allowNamedExports: true }],
    },
    {
      code: `
export { a as b };
const a = 1;
      `,
      options: [{ allowNamedExports: true }],
    },
    {
      code: `
export { Foo };
enum Foo { BAR }
      `,
      options: [{ allowNamedExports: true }],
    },
    // Decorators
    `
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
    // Namespace alias
    `
namespace A.X.Y {}
import Z = A.X.Y;
const X = 23;
    `,
  ],
  invalid: [
    // Basic use before define
    {
      code: `
a++;
var a = 19;
      `,
      errors: [{ messageId: 'noUseBeforeDefine' }],
    },
    {
      code: `
a();
var a = function () {};
      `,
      errors: [{ messageId: 'noUseBeforeDefine' }],
    },
    // Function with nested errors
    {
      code: `
a();
function a() {
  alert(b);
  var b = 10;
  a();
}
      `,
      errors: [
        { messageId: 'noUseBeforeDefine' },
        { messageId: 'noUseBeforeDefine' },
      ],
    },
    // Arrow function
    {
      code: `
(() => {
  alert(a);
  var a = 42;
})();
      `,
      errors: [{ messageId: 'noUseBeforeDefine' }],
    },
    // Class before define
    {
      code: `
new A();
class A {}
      `,
      errors: [{ messageId: 'noUseBeforeDefine' }],
    },
    {
      code: `
function foo() {
  new A();
}
class A {}
      `,
      errors: [{ messageId: 'noUseBeforeDefine' }],
    },
    // Self-referencing initializer
    {
      code: 'var a = a;',
      errors: [{ messageId: 'noUseBeforeDefine' }],
    },
    {
      code: 'let a = a + b;',
      errors: [{ messageId: 'noUseBeforeDefine' }],
    },
    {
      code: 'const a = foo(a);',
      errors: [{ messageId: 'noUseBeforeDefine' }],
    },
    {
      code: 'function foo(a = a) {}',
      errors: [{ messageId: 'noUseBeforeDefine' }],
    },
    // Destructuring
    {
      code: 'var { a = a } = [] as any;',
      errors: [{ messageId: 'noUseBeforeDefine' }],
    },
    {
      code: 'var [a = a] = [] as any;',
      errors: [{ messageId: 'noUseBeforeDefine' }],
    },
    {
      code: 'var { b = a, a } = {} as any;',
      errors: [{ messageId: 'noUseBeforeDefine' }],
    },
    {
      code: 'var { a = 0 } = a;',
      errors: [{ messageId: 'noUseBeforeDefine' }],
    },
    {
      code: 'var [a = 0] = a;',
      errors: [{ messageId: 'noUseBeforeDefine' }],
    },
    // For-in/for-of with inline declaration
    {
      code: `
for (var a in a) {}
      `,
      errors: [{ messageId: 'noUseBeforeDefine' }],
    },
    {
      code: `
for (var a of a) {}
      `,
      errors: [{ messageId: 'noUseBeforeDefine' }],
    },
    // ignoreTypeReferences: false
    {
      code: `
interface Bar {
  type: typeof Foo;
}
const Foo = 2;
      `,
      options: [{ ignoreTypeReferences: false }],
      errors: [{ messageId: 'noUseBeforeDefine' }],
    },
    // Enum references
    {
      code: `
function foo(): Foo {
  return Foo.FOO;
}
enum Foo { FOO }
      `,
      options: [{ enums: true }],
      errors: [{ messageId: 'noUseBeforeDefine' }],
    },
    {
      code: `
const foo = Foo.Foo;
enum Foo { FOO }
      `,
      options: [{ enums: true }],
      errors: [{ messageId: 'noUseBeforeDefine' }],
    },
    // Named exports (default allowNamedExports: false)
    {
      code: `
export { a };
const a = 1;
      `,
      errors: [{ messageId: 'noUseBeforeDefine' }],
    },
    {
      code: `
export { a as b };
const a = 1;
      `,
      errors: [{ messageId: 'noUseBeforeDefine' }],
    },
    {
      code: `
export { a, b };
let a: any, b: any;
      `,
      errors: [
        { messageId: 'noUseBeforeDefine' },
        { messageId: 'noUseBeforeDefine' },
      ],
    },
    {
      code: `
export { f };
function f() {}
      `,
      errors: [{ messageId: 'noUseBeforeDefine' }],
    },
    {
      code: `
export { C };
class C {}
      `,
      errors: [{ messageId: 'noUseBeforeDefine' }],
    },
    // Function call before define
    {
      code: `
f();
function f() {}
      `,
      errors: [{ messageId: 'noUseBeforeDefine' }],
    },
    {
      code: `
alert(a);
var a = 10;
      `,
      errors: [{ messageId: 'noUseBeforeDefine' }],
    },
    // Export enum/namespace before define
    {
      code: `
export { Foo };
enum Foo { BAR }
      `,
      errors: [{ messageId: 'noUseBeforeDefine' }],
    },
    {
      code: `
export { Foo };
namespace Foo {
  export let bar = () => console.log('bar');
}
      `,
      errors: [{ messageId: 'noUseBeforeDefine' }],
    },
  ],
});

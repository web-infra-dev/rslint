/**
 * @fileoverview Tests for semi rule.
 * @author Nicholas C. Zakas
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/semi/semi._js_.test.ts
 *   packages/eslint-plugin/rules/semi/semi._ts_.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, lang, ... })` -> `ruleTester.run('semi', null as never, { valid, invalid })`
 *  - The `$` unindent template tag is evaluated to its real multi-line string.
 *  - The `semi._ts_` `.reduce(...)` builders are evaluated by hand into concrete
 *    `{ code, options, output?, errors }` objects (valid: always + never; invalid:
 *    always/missingSemi + never/extraSemi, with `code.replace(/;/g, '')` stripping
 *    every semicolon).
 *  - The rule's only messages are `missingSemi` ('Missing semicolon.') and
 *    `extraSemi` ('Extra semicolon.') — neither interpolates `data`, so errors carry
 *    only `messageId` (+ line/column/endLine/endColumn where upstream gave them).
 *  - `parserOptions` (ecmaVersion / sourceType) dropped — rslint resolves via tsconfig.
 *  - `type` fields (deprecated AST node type) dropped — semi never set them.
 *
 * Cases that surface a real rslint<->upstream gap are NOT deleted or altered: they
 * are moved to the `KNOWN GAPS` block comment at the bottom, each annotated with
 * what upstream expects vs. what rslint produces.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('semi', null as never, {
  valid: [
    // ---- from semi._js_.test.ts ----
    'var x = 5;',
    'var x =5, y;',
    'foo();',
    'x = foo();',
    'setTimeout(function() {foo = "bar"; });',
    'setTimeout(function() {foo = "bar";});',
    'for (var a in b){}',
    'for (var i;;){}',
    'if (true) {}\n;[global, extended].forEach(function(){});',
    'throw new Error(\'foo\');',
    'debugger;',
    {
      code: 'throw new Error(\'foo\')',
      options: ['never'],
    },
    {
      code: 'var x = 5',
      options: ['never'],
    },
    {
      code: 'var x =5, y',
      options: ['never'],
    },
    {
      code: 'foo()',
      options: ['never'],
    },
    {
      code: 'debugger',
      options: ['never'],
    },
    {
      code: 'for (var a in b){}',
      options: ['never'],
    },
    {
      code: 'for (var i;;){}',
      options: ['never'],
    },
    {
      code: 'x = foo()',
      options: ['never'],
    },
    {
      code: 'if (true) {}\n;[global, extended].forEach(function(){})',
      options: ['never'],
    },
    {
      code: '(function bar() {})\n;(function foo(){})',
      options: ['never'],
    },
    {
      code: ';/foo/.test(\'bar\')',
      options: ['never'],
    },
    {
      code: ';+5',
      options: ['never'],
    },
    {
      code: ';-foo()',
      options: ['never'],
    },
    {
      code: 'a++\nb++',
      options: ['never'],
    },
    {
      code: 'a++; b++',
      options: ['never'],
    },
    {
      code: 'for (let thing of {}) {\n  console.log(thing);\n}',
    },
    {
      code: 'do{}while(true)',
      options: ['never'],
    },
    {
      code: 'do{}while(true);',
      options: ['always'],
    },
    {
      code: 'class C { static {} }',
    },
    {
      code: 'class C { static {} }',
      options: ['never'],
    },
    {
      code: 'class C { static { foo(); } }',
    },
    {
      code: 'class C { static { foo(); } }',
      options: ['always'],
    },
    {
      code: 'class C { static { foo(); bar(); } }',
    },
    {
      code: 'class C { static { foo(); bar(); baz();} }',
    },
    {
      code: 'class C { static { foo() } }',
      options: ['never'],
    },
    {
      code: 'class C { static { foo()\nbar() } }',
      options: ['never'],
    },
    {
      code: 'class C { static { foo()\nbar()\nbaz() } }',
      options: ['never'],
    },
    {
      code: 'class C { static { foo(); bar() } }',
      options: ['never'],
    },
    {
      code: 'class C { static { foo();\n (a) } }',
      options: ['never'],
    },
    {
      code: 'class C { static { foo()\n ;(a) } }',
      options: ['never'],
    },
    {
      code: 'class C { static { foo();\n [a] } }',
      options: ['never'],
    },
    {
      code: 'class C { static { foo()\n ;[a] } }',
      options: ['never'],
    },
    {
      code: 'class C { static { foo();\n +a } }',
      options: ['never'],
    },
    {
      code: 'class C { static { foo()\n ;+a } }',
      options: ['never'],
    },
    {
      code: 'class C { static { foo();\n -a } }',
      options: ['never'],
    },
    {
      code: 'class C { static { foo()\n ;-a } }',
      options: ['never'],
    },
    {
      code: 'class C { static { foo();\n /a/ } }',
      options: ['never'],
    },
    {
      code: 'class C { static { foo()\n ;/a/} }',
      options: ['never'],
    },
    {
      code: 'class C { static { foo();\n (a) } }',
      options: ['never', { beforeStatementContinuationChars: 'never' }],
    },
    {
      code: 'class C { static { do ; while (foo)\n (a)} }',
      options: ['never', { beforeStatementContinuationChars: 'never' }],
    },
    {
      code: 'class C { static { do ; while (foo)\n ;(a)} }',
      options: ['never', { beforeStatementContinuationChars: 'always' }],
    },
    // omitLastInOneLineBlock: true
    {
      code: 'if (foo) { bar() }',
      options: ['always', { omitLastInOneLineBlock: true }],
    },
    {
      code: 'if (foo) { bar(); baz() }',
      options: ['always', { omitLastInOneLineBlock: true }],
    },
    {
      code: 'if (foo)\n{ bar(); baz() }',
      options: ['always', { omitLastInOneLineBlock: true }],
    },
    {
      code: 'if (foo) {\n  bar(); baz(); }',
      options: ['always', { omitLastInOneLineBlock: true }],
    },
    {
      code: 'if (foo) { bar(); baz();\n}',
      options: ['always', { omitLastInOneLineBlock: true }],
    },
    {
      code: 'function foo() { bar(); baz() }',
      options: ['always', { omitLastInOneLineBlock: true }],
    },
    {
      code: 'function foo()\n{ bar(); baz() }',
      options: ['always', { omitLastInOneLineBlock: true }],
    },
    {
      code: 'function foo(){\n bar(); baz(); }',
      options: ['always', { omitLastInOneLineBlock: true }],
    },
    {
      code: 'function foo(){ bar(); baz(); \n}',
      options: ['always', { omitLastInOneLineBlock: true }],
    },
    {
      code: '() => { bar(); baz() };',
      options: ['always', { omitLastInOneLineBlock: true }],
    },
    {
      code: '() =>\n { bar(); baz() };',
      options: ['always', { omitLastInOneLineBlock: true }],
    },
    {
      code: '() => {\n bar(); baz(); };',
      options: ['always', { omitLastInOneLineBlock: true }],
    },
    {
      code: '() => { bar(); baz(); \n};',
      options: ['always', { omitLastInOneLineBlock: true }],
    },
    {
      code: 'const obj = { method() { bar(); baz() } };',
      options: ['always', { omitLastInOneLineBlock: true }],
    },
    {
      code: 'const obj = { method()\n { bar(); baz() } };',
      options: ['always', { omitLastInOneLineBlock: true }],
    },
    {
      code: 'const obj = { method() {\n bar(); baz(); } };',
      options: ['always', { omitLastInOneLineBlock: true }],
    },
    {
      code: 'const obj = { method() { bar(); baz(); \n} };',
      options: ['always', { omitLastInOneLineBlock: true }],
    },
    {
      code: 'class C {\n method() { bar(); baz() } \n}',
      options: ['always', { omitLastInOneLineBlock: true }],
    },
    {
      code: 'class C {\n method()\n { bar(); baz() } \n}',
      options: ['always', { omitLastInOneLineBlock: true }],
    },
    {
      code: 'class C {\n method() {\n bar(); baz(); } \n}',
      options: ['always', { omitLastInOneLineBlock: true }],
    },
    {
      code: 'class C {\n method() { bar(); baz(); \n} \n}',
      options: ['always', { omitLastInOneLineBlock: true }],
    },
    {
      code: 'class C {\n static { bar(); baz() } \n}',
      options: ['always', { omitLastInOneLineBlock: true }],
    },
    {
      code: 'class C {\n static\n { bar(); baz() } \n}',
      options: ['always', { omitLastInOneLineBlock: true }],
    },
    {
      code: 'class C {\n static {\n bar(); baz(); } \n}',
      options: ['always', { omitLastInOneLineBlock: true }],
    },
    {
      code: 'class C {\n static { bar(); baz(); \n} \n}',
      options: ['always', { omitLastInOneLineBlock: true }],
    },
    // omitLastInOneLineClassBody: true
    {
      code: 'export class SomeClass{\n    logType(){\n        console.log(this.type);\n    }\n}\n\nexport class Variant1 extends SomeClass{type=1}\nexport class Variant2 extends SomeClass{type=2}\nexport class Variant3 extends SomeClass{type=3}\nexport class Variant4 extends SomeClass{type=4}\nexport class Variant5 extends SomeClass{type=5}',
      options: ['always', { omitLastInOneLineClassBody: true }],
    },
    {
      code: '\n                export class SomeClass{\n                    logType(){\n                        console.log(this.type);\n                        console.log(this.anotherType);\n                    }\n                }\n\n                export class Variant1 extends SomeClass{type=1; anotherType=2}\n            ',
      options: ['always', { omitLastInOneLineClassBody: true }],
    },
    {
      code: '\n                export class SomeClass{\n                    logType(){\n                        console.log(this.type);\n                    }\n                }\n\n                export class Variant1 extends SomeClass{type=1;}\n                export class Variant2 extends SomeClass{type=2;}\n                export class Variant3 extends SomeClass{type=3;}\n                export class Variant4 extends SomeClass{type=4;}\n                export class Variant5 extends SomeClass{type=5;}\n            ',
      options: ['always', { omitLastInOneLineClassBody: false }],
    },
    {
      code: 'class C {\nfoo;}',
      options: ['always', { omitLastInOneLineClassBody: true }],
    },
    {
      code: 'class C {foo;\n}',
      options: ['always', { omitLastInOneLineClassBody: true }],
    },
    {
      code: 'class C {foo;\nbar;}',
      options: ['always', { omitLastInOneLineClassBody: true }],
    },
    {
      code: '{ foo; }',
      options: ['always', { omitLastInOneLineClassBody: true }],
    },
    {
      code: 'class C\n{ foo }',
      options: ['always', { omitLastInOneLineClassBody: true }],
    },
    // method definitions and static blocks don't have a semicolon.
    {
      code: 'class A { a() {} b() {} }',
    },
    {
      code: 'var A = class { a() {} b() {} };',
    },
    {
      code: 'class A { static {} }',
    },
    {
      code: 'import theDefault, { named1, named2 } from \'src/mylib\';',
    },
    {
      code: 'import theDefault, { named1, named2 } from \'src/mylib\'',
      options: ['never'],
    },
    // exports, "always"
    {
      code: 'export * from \'foo\';',
    },
    {
      code: 'export { foo } from \'foo\';',
    },
    {
      code: 'var foo = 0;export { foo };',
    },
    {
      code: 'export var foo;',
    },
    {
      code: 'export function foo () { }',
    },
    {
      code: 'export function* foo () { }',
    },
    {
      code: 'export class Foo { }',
    },
    {
      code: 'export let foo;',
    },
    {
      code: 'export const FOO = 42;',
    },
    {
      code: 'export default function() { }',
    },
    {
      code: 'export default function* () { }',
    },
    {
      code: 'export default class { }',
    },
    {
      code: 'export default foo || bar;',
    },
    {
      code: 'export default (foo) => foo.bar();',
    },
    {
      code: 'export default foo = 42;',
    },
    {
      code: 'export default foo += 42;',
    },
    // exports, "never"
    {
      code: 'export * from \'foo\'',
      options: ['never'],
    },
    {
      code: 'export { foo } from \'foo\'',
      options: ['never'],
    },
    {
      code: 'var foo = 0; export { foo }',
      options: ['never'],
    },
    {
      code: 'export var foo',
      options: ['never'],
    },
    {
      code: 'export function foo () { }',
      options: ['never'],
    },
    {
      code: 'export function* foo () { }',
      options: ['never'],
    },
    {
      code: 'export class Foo { }',
      options: ['never'],
    },
    {
      code: 'export let foo',
      options: ['never'],
    },
    {
      code: 'export const FOO = 42',
      options: ['never'],
    },
    {
      code: 'export default function() { }',
      options: ['never'],
    },
    {
      code: 'export default function* () { }',
      options: ['never'],
    },
    {
      code: 'export default class { }',
      options: ['never'],
    },
    {
      code: 'export default foo || bar',
      options: ['never'],
    },
    {
      code: 'export default (foo) => foo.bar()',
      options: ['never'],
    },
    {
      code: 'export default foo = 42',
      options: ['never'],
    },
    {
      code: 'export default foo += 42',
      options: ['never'],
    },
    {
      code: '++\nfoo;',
      options: ['always'],
    },
    {
      code: 'var a = b;\n+ c',
      options: ['never'],
    },
    // https://github.com/eslint/eslint/issues/7782
    {
      code: 'var a = b;\n/foo/.test(c)',
      options: ['never'],
    },
    {
      code: 'var a = b;\n`foo`',
      options: ['never'],
    },
    // https://github.com/eslint/eslint/issues/9521
    {
      code: 'do; while(a);\n[1,2,3].forEach(doSomething)',
      options: ['never', { beforeStatementContinuationChars: 'any' }],
    },
    {
      code: 'do; while(a)\n[1,2,3].forEach(doSomething)',
      options: ['never', { beforeStatementContinuationChars: 'any' }],
    },
    {
      code: 'import a from "a";\n[1,2,3].forEach(doSomething)',
      options: ['never', { beforeStatementContinuationChars: 'always' }],
    },
    {
      code: 'var a = 0; export {a};\n[a] = b',
      options: ['never', { beforeStatementContinuationChars: 'always' }],
    },
    {
      code: 'function wrap() {\n  return;\n  ({a} = b)\n}',
      options: ['never', { beforeStatementContinuationChars: 'always' }],
    },
    {
      code: 'while (true) {\n  break;\n  +i\n}',
      options: ['never', { beforeStatementContinuationChars: 'always' }],
    },
    {
      code: 'while (true) {\n  continue;\n  [1,2,3].forEach(doSomething)\n}',
      options: ['never', { beforeStatementContinuationChars: 'always' }],
    },
    {
      code: 'do; while(a);\n[1,2,3].forEach(doSomething)',
      options: ['never', { beforeStatementContinuationChars: 'always' }],
    },
    {
      code: 'const f = () => {};\n[1,2,3].forEach(doSomething)',
      options: ['never', { beforeStatementContinuationChars: 'always' }],
    },
    {
      code: 'import a from "a"\n[1,2,3].forEach(doSomething)',
      options: ['never', { beforeStatementContinuationChars: 'never' }],
    },
    {
      code: 'var a = 0; export {a}\n[a] = b',
      options: ['never', { beforeStatementContinuationChars: 'never' }],
    },
    {
      code: 'function wrap() {\n  return\n  ({a} = b)\n}',
      options: ['never', { beforeStatementContinuationChars: 'never' }],
    },
    {
      code: 'while (true) {\n  break\n  +i\n}',
      options: ['never', { beforeStatementContinuationChars: 'never' }],
    },
    {
      code: 'while (true) {\n  continue\n  [1,2,3].forEach(doSomething)\n}',
      options: ['never', { beforeStatementContinuationChars: 'never' }],
    },
    {
      code: 'do; while(a)\n[1,2,3].forEach(doSomething)',
      options: ['never', { beforeStatementContinuationChars: 'never' }],
    },
    {
      code: 'const f = () => {}\n[1,2,3].forEach(doSomething)',
      options: ['never', { beforeStatementContinuationChars: 'never' }],
    },
    // Class fields
    {
      code: 'class C { foo; }',
    },
    {
      code: 'class C { foo; }',
      options: ['always'],
    },
    {
      code: 'class C { foo }',
      options: ['never'],
    },
    {
      code: 'class C { foo = obj\n;[bar] }',
      options: ['never'],
    },
    {
      code: 'class C { foo;\n[bar]; }',
      options: ['always'],
    },
    {
      code: 'class C { foo\n;[bar] }',
      options: ['never'],
    },
    {
      code: 'class C { foo\n[bar] }',
      options: ['never'],
    },
    {
      code: 'class C { foo\n;[bar] }',
      options: ['never', { beforeStatementContinuationChars: 'always' }],
    },
    {
      code: 'class C { foo\n[bar] }',
      options: ['never', { beforeStatementContinuationChars: 'never' }],
    },
    {
      code: 'class C { foo = () => {}\n;[bar] }',
      options: ['never'],
    },
    {
      code: 'class C { foo = () => {}\n[bar] }',
      options: ['never'],
    },
    {
      code: 'class C { foo = () => {}\n;[bar] }',
      options: ['never', { beforeStatementContinuationChars: 'always' }],
    },
    {
      code: 'class C { foo = () => {}\n[bar] }',
      options: ['never', { beforeStatementContinuationChars: 'never' }],
    },
    {
      code: 'class C { foo() {} }',
      options: ['always'],
    },
    {
      code: 'class C { foo() {}; }',
      options: ['never'],
    }, // no-extra-semi reports it
    {
      code: 'class C { static {}; }',
      options: ['never'],
    }, // no-extra-semi reports it
    {
      code: 'class C { a=b;\n*foo() {} }',
      options: ['never'],
    },
    {
      code: 'class C { get;\nfoo() {} }',
      options: ['never'],
    },
    {
      code: 'class C { set;\nfoo() {} }',
      options: ['never'],
    },
    {
      code: 'class C { static;\nfoo() {} }',
      options: ['never'],
    },
    {
      code: 'class C { a=b;\nin }',
      options: ['never'],
    },
    {
      code: 'class C { a=b;\ninstanceof }',
      options: ['never'],
    },
    {
      code: '\n                class C {\n                    x\n                    [foo]\n\n                    x;\n                    [foo]\n\n                    x = "a";\n                    [foo]\n                }\n            ',
      options: ['never', { beforeStatementContinuationChars: 'never' }],
    },
    {
      code: '\n                class C {\n                    x\n                    [foo]\n\n                    x;\n                    [foo]\n\n                    x = 1;\n                    [foo]\n                }\n            ',
      options: ['never', { beforeStatementContinuationChars: 'always' }],
    },
    {
      code: 'class C { foo\n[bar] }',
      options: ['never', { beforeStatementContinuationChars: 'always' }],
    },
    {
      code: 'class C { foo = () => {}\n[bar] }',
      options: ['never', { beforeStatementContinuationChars: 'always' }],
    },
    {
      code: 'class C { foo\n;[bar] }',
      options: ['never', { beforeStatementContinuationChars: 'never' }],
    },
    {
      code: 'class C { foo = () => {}\n;[bar] }',
      options: ['never', { beforeStatementContinuationChars: 'never' }],
    },
    {
      code: 'class C { [foo] = bar;\nin }',
      options: ['never'],
    },
    {
      code: 'class C { #foo = bar;\nin }',
      options: ['never'],
    },
    {
      code: 'class C { static static = bar;\nin }',
      options: ['never'],
    },
    {
      code: 'class C { [foo];\nin }',
      options: ['never'],
    },
    {
      code: 'class C { [get];\nin }',
      options: ['never'],
    },
    {
      code: 'class C { [get] = 5;\nin }',
      options: ['never'],
    },
    {
      code: 'class C { #get;\nin }',
      options: ['never'],
    },
    {
      code: 'class C { #set = 5;\nin }',
      options: ['never'],
    },
    {
      code: 'class C { static static;\nin }',
      options: ['never'],
    },

    // ---- from semi._ts_.test.ts ----
    // https://github.com/typescript-eslint/typescript-eslint/issues/366
    {
      code: 'export = Foo;',
      options: ['always'],
    },
    {
      code: 'export = Foo',
      options: ['never'],
    },
    {
      code: 'import f = require("f");',
      options: ['always'],
    },
    {
      code: 'import f = require("f")',
      options: ['never'],
    },
    {
      code: 'type Foo = {};',
      options: ['always'],
    },
    {
      code: 'type Foo = {}',
      options: ['never'],
    },
    {
      code: 'class C { accessor foo; } ',
      options: ['always'],
    },
    {
      code: 'class C { accessor foo } ',
      options: ['never'],
    },
    {
      code: 'class C { accessor [foo]; } ',
      options: ['always'],
    },
    {
      code: 'class C { accessor [foo] } ',
      options: ['never'],
    },
    // https://github.com/typescript-eslint/typescript-eslint/issues/409
    {
      code: 'class Class {\n    prop: string;\n}',
      options: ['always'],
    },
    {
      code: 'class Class {\n    prop: string\n}',
      options: ['never'],
    },
    {
      code: 'abstract class AbsClass {\n    abstract prop: string;\n    abstract meth(): string;\n}',
      options: ['always'],
    },
    {
      code: 'abstract class AbsClass {\n    abstract prop: string\n    abstract meth(): string\n}',
      options: ['never'],
    },
    {
      code: 'class PanCamera extends FreeCamera {\n  public invertY: boolean = false;\n}',
      options: ['always'],
    },
    {
      code: 'class PanCamera extends FreeCamera {\n  public invertY: boolean = false\n}',
      options: ['never'],
    },
    // https://github.com/typescript-eslint/typescript-eslint/issues/123
    {
      code: 'export default interface test {}',
      options: ['always'],
    },
    {
      code: 'export default interface test {}',
      options: ['never'],
    },
    {
      code: 'declare function declareFn(): string;',
      options: ['always'],
    },
    {
      code: 'declare function declareFn(): string',
      options: ['never'],
    },
  ],

  invalid: [
    // ---- from semi._js_.test.ts ----
    {
      code: 'import * as utils from \'./utils\'',
      output: 'import * as utils from \'./utils\';',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 33,
        },
      ],
    },
    {
      code: 'import { square, diag } from \'lib\'',
      output: 'import { square, diag } from \'lib\';',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 35,
        },
      ],
    },
    {
      code: 'import { default as foo } from \'lib\'',
      output: 'import { default as foo } from \'lib\';',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 37,
        },
      ],
    },
    {
      code: 'import \'src/mylib\'',
      output: 'import \'src/mylib\';',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 19,
        },
      ],
    },
    {
      code: 'import theDefault, { named1, named2 } from \'src/mylib\'',
      output: 'import theDefault, { named1, named2 } from \'src/mylib\';',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 55,
        },
      ],
    },
    {
      code: 'function foo() { return [] }',
      output: 'function foo() { return []; }',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: 'while(true) { break }',
      output: 'while(true) { break; }',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: 'while(true) { continue }',
      output: 'while(true) { continue; }',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'let x = 5',
      output: 'let x = 5;',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 10,
        },
      ],
    },
    {
      code: 'var x = 5',
      output: 'var x = 5;',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 10,
        },
      ],
    },
    {
      code: 'var x = 5, y',
      output: 'var x = 5, y;',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 13,
        },
      ],
    },
    {
      code: 'debugger',
      output: 'debugger;',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 9,
        },
      ],
    },
    {
      code: 'foo()',
      output: 'foo();',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 6,
        },
      ],
    },
    {
      code: 'foo()\n',
      output: 'foo();\n',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 6,
          endLine: 2,
          endColumn: 1,
        },
      ],
    },
    {
      code: 'foo()\r\n',
      output: 'foo();\r\n',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 6,
          endLine: 2,
          endColumn: 1,
        },
      ],
    },
    {
      code: 'foo()\nbar();',
      output: 'foo();\nbar();',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 6,
          endLine: 2,
          endColumn: 1,
        },
      ],
    },
    {
      code: 'foo()\r\nbar();',
      output: 'foo();\r\nbar();',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 6,
          endLine: 2,
          endColumn: 1,
        },
      ],
    },
    {
      code: 'for (var a in b) var i ',
      output: 'for (var a in b) var i; ',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'for (;;){var i}',
      output: 'for (;;){var i;}',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: 'for (;;) var i ',
      output: 'for (;;) var i; ',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: 'for (var j;;) {var i}',
      output: 'for (var j;;) {var i;}',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'var foo = {\n bar: baz\n}',
      output: 'var foo = {\n bar: baz\n};',
      errors: [
        {
          messageId: 'missingSemi',
          line: 3,
          column: 2,
        },
      ],
    },
    {
      code: 'var foo\nvar bar;',
      output: 'var foo;\nvar bar;',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 8,
          endLine: 2,
          endColumn: 1,
        },
      ],
    },
    {
      code: 'throw new Error(\'foo\')',
      output: 'throw new Error(\'foo\');',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 23,
        },
      ],
    },
    {
      code: 'do{}while(true)',
      output: 'do{}while(true);',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 16,
        },
      ],
    },
    {
      code: 'if (foo) {bar()}',
      output: 'if (foo) {bar();}',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: 'if (foo) {bar()} ',
      output: 'if (foo) {bar();} ',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: 'if (foo) {bar()\n}',
      output: 'if (foo) {bar();\n}',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 16,
          endLine: 2,
          endColumn: 1,
        },
      ],
    },
    {
      code: 'throw new Error(\'foo\');',
      options: ['never'],
      output: 'throw new Error(\'foo\')',
      errors: [
        {
          messageId: 'extraSemi',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'function foo() { return []; }',
      options: ['never'],
      output: 'function foo() { return [] }',
      errors: [
        {
          messageId: 'extraSemi',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: 'while(true) { break; }',
      options: ['never'],
      output: 'while(true) { break }',
      errors: [
        {
          messageId: 'extraSemi',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: 'while(true) { continue; }',
      options: ['never'],
      output: 'while(true) { continue }',
      errors: [
        {
          messageId: 'extraSemi',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'let x = 5;',
      options: ['never'],
      output: 'let x = 5',
      errors: [
        {
          messageId: 'extraSemi',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: 'var x = 5;',
      options: ['never'],
      output: 'var x = 5',
      errors: [
        {
          messageId: 'extraSemi',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: 'var x = 5, y;',
      options: ['never'],
      output: 'var x = 5, y',
      errors: [
        {
          messageId: 'extraSemi',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'debugger;',
      options: ['never'],
      output: 'debugger',
      errors: [
        {
          messageId: 'extraSemi',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: 'foo();',
      options: ['never'],
      output: 'foo()',
      errors: [
        {
          messageId: 'extraSemi',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 7,
        },
      ],
    },
    {
      code: 'for (var a in b) var i; ',
      options: ['never'],
      output: 'for (var a in b) var i ',
      errors: [
        {
          messageId: 'extraSemi',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'for (;;){var i;}',
      options: ['never'],
      output: 'for (;;){var i}',
      errors: [
        {
          messageId: 'extraSemi',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: 'for (;;) var i; ',
      options: ['never'],
      output: 'for (;;) var i ',
      errors: [
        {
          messageId: 'extraSemi',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: 'for (var j;;) {var i;}',
      options: ['never'],
      output: 'for (var j;;) {var i}',
      errors: [
        {
          messageId: 'extraSemi',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'var foo = {\n bar: baz\n};',
      options: ['never'],
      output: 'var foo = {\n bar: baz\n}',
      errors: [
        {
          messageId: 'extraSemi',
          line: 3,
          column: 2,
          endLine: 3,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'import theDefault, { named1, named2 } from \'src/mylib\';',
      options: ['never'],
      output: 'import theDefault, { named1, named2 } from \'src/mylib\'',
      errors: [
        {
          messageId: 'extraSemi',
          line: 1,
          column: 55,
          endLine: 1,
          endColumn: 56,
        },
      ],
    },
    {
      code: 'do{}while(true);',
      options: ['never'],
      output: 'do{}while(true)',
      errors: [
        {
          messageId: 'extraSemi',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: 'class C { static { foo() } }',
      output: 'class C { static { foo(); } }',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: 'class C { static { foo() } }',
      options: ['always'],
      output: 'class C { static { foo(); } }',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: 'class C { static { foo(); bar() } }',
      output: 'class C { static { foo(); bar(); } }',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: 'class C { static { foo()\nbar(); } }',
      output: 'class C { static { foo();\nbar(); } }',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 25,
          endLine: 2,
          endColumn: 1,
        },
      ],
    },
    {
      code: 'class C { static { foo(); bar()\nbaz(); } }',
      output: 'class C { static { foo(); bar();\nbaz(); } }',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 32,
          endLine: 2,
          endColumn: 1,
        },
      ],
    },
    {
      code: 'class C { static { foo(); } }',
      options: ['never'],
      output: 'class C { static { foo() } }',
      errors: [
        {
          messageId: 'extraSemi',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: 'class C { static { foo();\nbar() } }',
      options: ['never'],
      output: 'class C { static { foo()\nbar() } }',
      errors: [
        {
          messageId: 'extraSemi',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: 'class C { static { foo()\nbar(); } }',
      options: ['never'],
      output: 'class C { static { foo()\nbar() } }',
      errors: [
        {
          messageId: 'extraSemi',
          line: 2,
          column: 6,
          endLine: 2,
          endColumn: 7,
        },
      ],
    },
    {
      code: 'class C { static { foo()\nbar();\nbaz() } }',
      options: ['never'],
      output: 'class C { static { foo()\nbar()\nbaz() } }',
      errors: [
        {
          messageId: 'extraSemi',
          line: 2,
          column: 6,
          endLine: 2,
          endColumn: 7,
        },
      ],
    },
    {
      code: 'class C { static { do ; while (foo)\n (a)} }',
      options: ['never', { beforeStatementContinuationChars: 'always' }],
      output: 'class C { static { do ; while (foo);\n (a)} }',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 36,
          endLine: 2,
          endColumn: 1,
        },
      ],
    },
    {
      code: 'class C { static { do ; while (foo)\n ;(a)} }',
      options: ['never', { beforeStatementContinuationChars: 'never' }],
      output: 'class C { static { do ; while (foo)\n (a)} }',
      errors: [
        {
          messageId: 'extraSemi',
          line: 2,
          column: 2,
          endLine: 2,
          endColumn: 3,
        },
      ],
    },
    // omitLastInOneLineBlock: true
    {
      code: 'if (foo) { bar()\n }',
      options: ['always', { omitLastInOneLineBlock: true }],
      output: 'if (foo) { bar();\n }',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 17,
          endLine: 2,
          endColumn: 1,
        },
      ],
    },
    {
      code: 'if (foo) {\n bar() }',
      options: ['always', { omitLastInOneLineBlock: true }],
      output: 'if (foo) {\n bar(); }',
      errors: [
        {
          messageId: 'missingSemi',
          line: 2,
          column: 7,
          endLine: 2,
          endColumn: 8,
        },
      ],
    },
    {
      code: 'if (foo) {\n bar(); baz() }',
      options: ['always', { omitLastInOneLineBlock: true }],
      output: 'if (foo) {\n bar(); baz(); }',
      errors: [
        {
          messageId: 'missingSemi',
          line: 2,
          column: 14,
          endLine: 2,
          endColumn: 15,
        },
      ],
    },
    {
      code: 'if (foo) { bar(); }',
      options: ['always', { omitLastInOneLineBlock: true }],
      output: 'if (foo) { bar() }',
      errors: [
        {
          messageId: 'extraSemi',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'function foo() { bar(); baz(); }',
      options: ['always', { omitLastInOneLineBlock: true }],
      output: 'function foo() { bar(); baz() }',
      errors: [
        {
          messageId: 'extraSemi',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: 'function foo()\n{ bar(); baz(); }',
      options: ['always', { omitLastInOneLineBlock: true }],
      output: 'function foo()\n{ bar(); baz() }',
      errors: [
        {
          messageId: 'extraSemi',
          line: 2,
          column: 15,
          endLine: 2,
          endColumn: 16,
        },
      ],
    },
    {
      code: 'function foo() {\n bar(); baz() }',
      options: ['always', { omitLastInOneLineBlock: true }],
      output: 'function foo() {\n bar(); baz(); }',
      errors: [
        {
          messageId: 'missingSemi',
          line: 2,
          column: 14,
          endLine: 2,
          endColumn: 15,
        },
      ],
    },
    {
      code: 'function foo() { bar(); baz() \n}',
      options: ['always', { omitLastInOneLineBlock: true }],
      output: 'function foo() { bar(); baz(); \n}',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: 'class C {\nfoo() { bar(); baz(); }\n}',
      options: ['always', { omitLastInOneLineBlock: true }],
      output: 'class C {\nfoo() { bar(); baz() }\n}',
      errors: [
        {
          messageId: 'extraSemi',
          line: 2,
          column: 21,
          endLine: 2,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'class C {\nfoo() \n{ bar(); baz(); }\n}',
      options: ['always', { omitLastInOneLineBlock: true }],
      output: 'class C {\nfoo() \n{ bar(); baz() }\n}',
      errors: [
        {
          messageId: 'extraSemi',
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 16,
        },
      ],
    },
    {
      code: 'class C {\nfoo() {\n bar(); baz() }\n}',
      options: ['always', { omitLastInOneLineBlock: true }],
      output: 'class C {\nfoo() {\n bar(); baz(); }\n}',
      errors: [
        {
          messageId: 'missingSemi',
          line: 3,
          column: 14,
          endLine: 3,
          endColumn: 15,
        },
      ],
    },
    {
      code: 'class C {\nfoo() { bar(); baz() \n}\n}',
      options: ['always', { omitLastInOneLineBlock: true }],
      output: 'class C {\nfoo() { bar(); baz(); \n}\n}',
      errors: [
        {
          messageId: 'missingSemi',
          line: 2,
          column: 21,
          endLine: 2,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'class C {\nstatic { bar(); baz(); }\n}',
      options: ['always', { omitLastInOneLineBlock: true }],
      output: 'class C {\nstatic { bar(); baz() }\n}',
      errors: [
        {
          messageId: 'extraSemi',
          line: 2,
          column: 22,
          endLine: 2,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'class C {\nstatic \n{ bar(); baz(); }\n}',
      options: ['always', { omitLastInOneLineBlock: true }],
      output: 'class C {\nstatic \n{ bar(); baz() }\n}',
      errors: [
        {
          messageId: 'extraSemi',
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 16,
        },
      ],
    },
    {
      code: 'class C {\nstatic {\n bar(); baz() }\n}',
      options: ['always', { omitLastInOneLineBlock: true }],
      output: 'class C {\nstatic {\n bar(); baz(); }\n}',
      errors: [
        {
          messageId: 'missingSemi',
          line: 3,
          column: 14,
          endLine: 3,
          endColumn: 15,
        },
      ],
    },
    {
      code: 'class C {\nfoo() { bar(); baz() \n}\n}',
      options: ['always', { omitLastInOneLineBlock: true }],
      output: 'class C {\nfoo() { bar(); baz(); \n}\n}',
      errors: [
        {
          messageId: 'missingSemi',
          line: 2,
          column: 21,
          endLine: 2,
          endColumn: 22,
        },
      ],
    },
    // exports, "always"
    {
      code: 'export * from \'foo\'',
      output: 'export * from \'foo\';',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 20,
        },
      ],
    },
    {
      code: 'export { foo } from \'foo\'',
      output: 'export { foo } from \'foo\';',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 26,
        },
      ],
    },
    {
      code: 'var foo = 0;export { foo }',
      output: 'var foo = 0;export { foo };',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 27,
        },
      ],
    },
    {
      code: 'export var foo',
      output: 'export var foo;',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 15,
        },
      ],
    },
    {
      code: 'export let foo',
      output: 'export let foo;',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 15,
        },
      ],
    },
    {
      code: 'export const FOO = 42',
      output: 'export const FOO = 42;',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 22,
        },
      ],
    },
    {
      code: 'export default foo || bar',
      output: 'export default foo || bar;',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 26,
        },
      ],
    },
    {
      code: 'export default (foo) => foo.bar()',
      output: 'export default (foo) => foo.bar();',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 34,
        },
      ],
    },
    {
      code: 'export default foo = 42',
      output: 'export default foo = 42;',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 24,
        },
      ],
    },
    {
      code: 'export default foo += 42',
      output: 'export default foo += 42;',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 25,
        },
      ],
    },
    // exports, "never"
    {
      code: 'export * from \'foo\';',
      options: ['never'],
      output: 'export * from \'foo\'',
      errors: [
        {
          messageId: 'extraSemi',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: 'export { foo } from \'foo\';',
      options: ['never'],
      output: 'export { foo } from \'foo\'',
      errors: [
        {
          messageId: 'extraSemi',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: 'var foo = 0;export { foo };',
      options: ['never'],
      output: 'var foo = 0;export { foo }',
      errors: [
        {
          messageId: 'extraSemi',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: 'export var foo;',
      options: ['never'],
      output: 'export var foo',
      errors: [
        {
          messageId: 'extraSemi',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: 'export let foo;',
      options: ['never'],
      output: 'export let foo',
      errors: [
        {
          messageId: 'extraSemi',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: 'export const FOO = 42;',
      options: ['never'],
      output: 'export const FOO = 42',
      errors: [
        {
          messageId: 'extraSemi',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'export default foo || bar;',
      options: ['never'],
      output: 'export default foo || bar',
      errors: [
        {
          messageId: 'extraSemi',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: 'export default (foo) => foo.bar();',
      options: ['never'],
      output: 'export default (foo) => foo.bar()',
      errors: [
        {
          messageId: 'extraSemi',
          line: 1,
          column: 34,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: 'export default foo = 42;',
      options: ['never'],
      output: 'export default foo = 42',
      errors: [
        {
          messageId: 'extraSemi',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: 'export default foo += 42;',
      options: ['never'],
      output: 'export default foo += 42',
      errors: [
        {
          messageId: 'extraSemi',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: 'a;\n++b',
      options: ['never'],
      output: 'a\n++b',
      errors: [
        {
          messageId: 'extraSemi',
          line: 1,
          column: 2,
          endLine: 1,
          endColumn: 3,
        },
      ],
    },
    // https://github.com/eslint/eslint/issues/9521
    {
      code: 'import a from "a"\n[1,2,3].forEach(doSomething)',
      options: ['never', { beforeStatementContinuationChars: 'always' }],
      output: 'import a from "a";\n[1,2,3].forEach(doSomething)',
      errors: [
        {
          messageId: 'missingSemi',
        },
      ],
    },
    {
      code: 'var a = 0; export {a}\n[a] = b',
      options: ['never', { beforeStatementContinuationChars: 'always' }],
      output: 'var a = 0; export {a};\n[a] = b',
      errors: [
        {
          messageId: 'missingSemi',
        },
      ],
    },
    {
      code: 'function wrap() {\n  return\n  ({a} = b)\n}',
      options: ['never', { beforeStatementContinuationChars: 'always' }],
      output: 'function wrap() {\n  return;\n  ({a} = b)\n}',
      errors: [
        {
          messageId: 'missingSemi',
        },
      ],
    },
    {
      code: 'while (true) {\n  break\n  +i\n}',
      options: ['never', { beforeStatementContinuationChars: 'always' }],
      output: 'while (true) {\n  break;\n  +i\n}',
      errors: [
        {
          messageId: 'missingSemi',
        },
      ],
    },
    {
      code: 'while (true) {\n  continue\n  [1,2,3].forEach(doSomething)\n}',
      options: ['never', { beforeStatementContinuationChars: 'always' }],
      output: 'while (true) {\n  continue;\n  [1,2,3].forEach(doSomething)\n}',
      errors: [
        {
          messageId: 'missingSemi',
        },
      ],
    },
    {
      code: 'do; while(a)\n[1,2,3].forEach(doSomething)',
      options: ['never', { beforeStatementContinuationChars: 'always' }],
      output: 'do; while(a);\n[1,2,3].forEach(doSomething)',
      errors: [
        {
          messageId: 'missingSemi',
        },
      ],
    },
    {
      code: 'const f = () => {}\n[1,2,3].forEach(doSomething)',
      options: ['never', { beforeStatementContinuationChars: 'always' }],
      output: 'const f = () => {};\n[1,2,3].forEach(doSomething)',
      errors: [
        {
          messageId: 'missingSemi',
        },
      ],
    },
    {
      code: 'import a from "a";\n[1,2,3].forEach(doSomething)',
      options: ['never', { beforeStatementContinuationChars: 'never' }],
      output: 'import a from "a"\n[1,2,3].forEach(doSomething)',
      errors: [
        {
          messageId: 'extraSemi',
        },
      ],
    },
    {
      code: 'var a = 0; export {a};\n[a] = b',
      options: ['never', { beforeStatementContinuationChars: 'never' }],
      output: 'var a = 0; export {a}\n[a] = b',
      errors: [
        {
          messageId: 'extraSemi',
        },
      ],
    },
    {
      code: 'function wrap() {\n  return;\n  ({a} = b)\n}',
      options: ['never', { beforeStatementContinuationChars: 'never' }],
      output: 'function wrap() {\n  return\n  ({a} = b)\n}',
      errors: [
        {
          messageId: 'extraSemi',
        },
      ],
    },
    {
      code: 'while (true) {\n  break;\n  +i\n}',
      options: ['never', { beforeStatementContinuationChars: 'never' }],
      output: 'while (true) {\n  break\n  +i\n}',
      errors: [
        {
          messageId: 'extraSemi',
        },
      ],
    },
    {
      code: 'while (true) {\n  continue;\n  [1,2,3].forEach(doSomething)\n}',
      options: ['never', { beforeStatementContinuationChars: 'never' }],
      output: 'while (true) {\n  continue\n  [1,2,3].forEach(doSomething)\n}',
      errors: [
        {
          messageId: 'extraSemi',
        },
      ],
    },
    {
      code: 'do; while(a);\n[1,2,3].forEach(doSomething)',
      options: ['never', { beforeStatementContinuationChars: 'never' }],
      output: 'do; while(a)\n[1,2,3].forEach(doSomething)',
      errors: [
        {
          messageId: 'extraSemi',
        },
      ],
    },
    {
      code: 'const f = () => {};\n[1,2,3].forEach(doSomething)',
      options: ['never', { beforeStatementContinuationChars: 'never' }],
      output: 'const f = () => {}\n[1,2,3].forEach(doSomething)',
      errors: [
        {
          messageId: 'extraSemi',
        },
      ],
    },
    {
      code: 'import a from "a"\n;[1,2,3].forEach(doSomething)',
      options: ['never', { beforeStatementContinuationChars: 'never' }],
      output: 'import a from "a"\n[1,2,3].forEach(doSomething)',
      errors: [
        {
          messageId: 'extraSemi',
        },
      ],
    },
    {
      code: 'var a = 0; export {a}\n;[1,2,3].forEach(doSomething)',
      options: ['never', { beforeStatementContinuationChars: 'never' }],
      output: 'var a = 0; export {a}\n[1,2,3].forEach(doSomething)',
      errors: [
        {
          messageId: 'extraSemi',
        },
      ],
    },
    {
      code: 'function wrap() {\n  return\n  ;[1,2,3].forEach(doSomething)\n}',
      options: ['never', { beforeStatementContinuationChars: 'never' }],
      output: 'function wrap() {\n  return\n  [1,2,3].forEach(doSomething)\n}',
      errors: [
        {
          messageId: 'extraSemi',
        },
      ],
    },
    {
      code: 'while (true) {\n  break\n  ;[1,2,3].forEach(doSomething)\n}',
      options: ['never', { beforeStatementContinuationChars: 'never' }],
      output: 'while (true) {\n  break\n  [1,2,3].forEach(doSomething)\n}',
      errors: [
        {
          messageId: 'extraSemi',
        },
      ],
    },
    {
      code: 'while (true) {\n  continue\n  ;[1,2,3].forEach(doSomething)\n}',
      options: ['never', { beforeStatementContinuationChars: 'never' }],
      output: 'while (true) {\n  continue\n  [1,2,3].forEach(doSomething)\n}',
      errors: [
        {
          messageId: 'extraSemi',
        },
      ],
    },
    {
      code: 'do; while(a)\n;[1,2,3].forEach(doSomething)',
      options: ['never', { beforeStatementContinuationChars: 'never' }],
      output: 'do; while(a)\n[1,2,3].forEach(doSomething)',
      errors: [
        {
          messageId: 'extraSemi',
        },
      ],
    },
    {
      code: 'const f = () => {}\n;[1,2,3].forEach(doSomething)',
      options: ['never', { beforeStatementContinuationChars: 'never' }],
      output: 'const f = () => {}\n[1,2,3].forEach(doSomething)',
      errors: [
        {
          messageId: 'extraSemi',
        },
      ],
    },
    // Class fields
    {
      code: 'class C { foo }',
      output: 'class C { foo; }',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: 'class C { foo }',
      options: ['always'],
      output: 'class C { foo; }',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: 'class C { foo; }',
      options: ['never'],
      output: 'class C { foo }',
      errors: [
        {
          messageId: 'extraSemi',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: 'class C { foo\n[bar]; }',
      options: ['always'],
      output: 'class C { foo;\n[bar]; }',
      errors: [
        {
          messageId: 'missingSemi',
          line: 1,
          column: 14,
          endLine: 2,
          endColumn: 1,
        },
      ],
    },
    // class fields
    {
      code: 'class C { [get];\nfoo\n}',
      options: ['never'],
      output: 'class C { [get]\nfoo\n}',
      errors: [
        {
          messageId: 'extraSemi',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: 'class C { [set];\nfoo\n}',
      options: ['never'],
      output: 'class C { [set]\nfoo\n}',
      errors: [
        {
          messageId: 'extraSemi',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: 'class C { #get;\nfoo\n}',
      options: ['never'],
      output: 'class C { #get\nfoo\n}',
      errors: [
        {
          messageId: 'extraSemi',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: 'class C { #set;\nfoo\n}',
      options: ['never'],
      output: 'class C { #set\nfoo\n}',
      errors: [
        {
          messageId: 'extraSemi',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: 'class C { #static;\nfoo\n}',
      options: ['never'],
      output: 'class C { #static\nfoo\n}',
      errors: [
        {
          messageId: 'extraSemi',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'class C { get=1;\nfoo\n}',
      options: ['never'],
      output: 'class C { get=1\nfoo\n}',
      errors: [
        {
          messageId: 'extraSemi',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: 'class C { static static;\nfoo\n}',
      options: ['never'],
      output: 'class C { static static\nfoo\n}',
      errors: [
        {
          messageId: 'extraSemi',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: 'class C { static;\n}',
      options: ['never'],
      output: 'class C { static\n}',
      errors: [
        {
          messageId: 'extraSemi',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    // omitLastInOneLineClassBody
    {
      code: '\n                export class SomeClass{\n                    logType(){\n                        console.log(this.type);\n                    }\n                }\n\n                export class Variant1 extends SomeClass{type=1}\n            ',
      options: ['always', { omitLastInOneLineClassBody: false }],
      output: '\n                export class SomeClass{\n                    logType(){\n                        console.log(this.type);\n                    }\n                }\n\n                export class Variant1 extends SomeClass{type=1;}\n            ',
      errors: [
        {
          messageId: 'missingSemi',
          line: 8,
          column: 63,
          endLine: 8,
          endColumn: 64,
        },
      ],
    },
    {
      code: '\n                export class SomeClass{\n                    logType(){\n                        console.log(this.type);\n                    }\n                }\n\n                export class Variant1 extends SomeClass{type=1}\n            ',
      options: ['always', { omitLastInOneLineClassBody: false, omitLastInOneLineBlock: true }],
      output: '\n                export class SomeClass{\n                    logType(){\n                        console.log(this.type);\n                    }\n                }\n\n                export class Variant1 extends SomeClass{type=1;}\n            ',
      errors: [
        {
          messageId: 'missingSemi',
          line: 8,
          column: 63,
          endLine: 8,
          endColumn: 64,
        },
      ],
    },
    {
      code: '\n                export class SomeClass{\n                    logType(){\n                        console.log(this.type);\n                    }\n                }\n\n                export class Variant1 extends SomeClass{type=1;}\n            ',
      options: ['always', { omitLastInOneLineClassBody: true, omitLastInOneLineBlock: false }],
      output: '\n                export class SomeClass{\n                    logType(){\n                        console.log(this.type);\n                    }\n                }\n\n                export class Variant1 extends SomeClass{type=1}\n            ',
      errors: [
        {
          messageId: 'extraSemi',
          line: 8,
          column: 63,
          endLine: 8,
          endColumn: 64,
        },
      ],
    },
    {
      code: '\n                export class SomeClass{\n                    logType(){\n                        console.log(this.type);\n                        console.log(this.anotherType);\n                    }\n                }\n\n                export class Variant1 extends SomeClass{type=1; anotherType=2}\n            ',
      options: ['always', { omitLastInOneLineClassBody: false, omitLastInOneLineBlock: true }],
      output: '\n                export class SomeClass{\n                    logType(){\n                        console.log(this.type);\n                        console.log(this.anotherType);\n                    }\n                }\n\n                export class Variant1 extends SomeClass{type=1; anotherType=2;}\n            ',
      errors: [
        {
          messageId: 'missingSemi',
          line: 9,
          column: 78,
          endLine: 9,
          endColumn: 79,
        },
      ],
    },

    // ---- from semi._ts_.test.ts ----
    {
      code: 'class A {\n  method(): void\n  method(arg?: any): void {\n\n  }\n}',
      options: ['always'],
      output: 'class A {\n  method(): void;\n  method(arg?: any): void {\n\n  }\n}',
      errors: [
        {
          messageId: 'missingSemi',
        },
      ],
    },
    {
      code: 'class A {\n  method(): void;\n  method(arg?: any): void {\n\n  }\n}',
      options: ['never'],
      output: 'class A {\n  method(): void\n  method(arg?: any): void {\n\n  }\n}',
      errors: [
        {
          messageId: 'extraSemi',
        },
      ],
    },
    {
      code: 'import a from "a"\n(function() {\n    // ...\n})()',
      options: ['never', { beforeStatementContinuationChars: 'always' }],
      output: 'import a from "a";\n(function() {\n    // ...\n})()',
      errors: [
        {
          messageId: 'missingSemi',
        },
      ],
    },
    {
      code: 'import a from "a"\n;(function() {\n    // ...\n})()',
      options: ['never', { beforeStatementContinuationChars: 'never' }],
      output: 'import a from "a"\n(function() {\n    // ...\n})()',
      errors: [
        {
          messageId: 'extraSemi',
        },
      ],
    },
    {
      code: 'declare function declareFn(): string',
      options: ['always'],
      output: 'declare function declareFn(): string;',
      errors: [
        {
          line: 1,
          messageId: 'missingSemi',
        },
      ],
    },
    {
      code: 'declare function declareFn(): string;',
      options: ['never'],
      output: 'declare function declareFn(): string',
      errors: [
        {
          line: 1,
          messageId: 'extraSemi',
        },
      ],
    },
    // https://github.com/typescript-eslint/typescript-eslint/issues/366
    {
      code: 'export = Foo',
      options: ['always'],
      output: 'export = Foo;',
      errors: [
        {
          line: 1,
          messageId: 'missingSemi',
        },
      ],
    },
    {
      code: 'export = Foo;',
      options: ['never'],
      output: 'export = Foo',
      errors: [
        {
          line: 1,
          messageId: 'extraSemi',
        },
      ],
    },
    {
      code: 'import f = require("f")',
      options: ['always'],
      output: 'import f = require("f");',
      errors: [
        {
          line: 1,
          messageId: 'missingSemi',
        },
      ],
    },
    {
      code: 'import f = require("f");',
      options: ['never'],
      output: 'import f = require("f")',
      errors: [
        {
          line: 1,
          messageId: 'extraSemi',
        },
      ],
    },
    {
      code: 'type Foo = {}',
      options: ['always'],
      output: 'type Foo = {};',
      errors: [
        {
          line: 1,
          messageId: 'missingSemi',
        },
      ],
    },
    {
      code: 'type Foo = {};',
      options: ['never'],
      output: 'type Foo = {}',
      errors: [
        {
          line: 1,
          messageId: 'extraSemi',
        },
      ],
    },
    // https://github.com/typescript-eslint/typescript-eslint/issues/409
    {
      code: 'class Class {\n    prop: string\n}',
      options: ['always'],
      output: 'class Class {\n    prop: string;\n}',
      errors: [
        {
          line: 2,
          messageId: 'missingSemi',
        },
      ],
    },
    {
      code: 'class Class {\n    prop: string;\n}',
      options: ['never'],
      output: 'class Class {\n    prop: string\n}',
      errors: [
        {
          line: 2,
          messageId: 'extraSemi',
        },
      ],
    },
    {
      code: 'abstract class AbsClass {\n    abstract prop: string\n    abstract meth(): string\n}',
      options: ['always'],
      output: 'abstract class AbsClass {\n    abstract prop: string;\n    abstract meth(): string;\n}',
      errors: [
        {
          line: 2,
          messageId: 'missingSemi',
        },
        {
          line: 3,
          messageId: 'missingSemi',
        },
      ],
    },
    {
      code: 'abstract class AbsClass {\n    abstract prop: string;\n    abstract meth(): string;\n}',
      options: ['never'],
      output: 'abstract class AbsClass {\n    abstract prop: string\n    abstract meth(): string\n}',
      errors: [
        {
          line: 2,
          messageId: 'extraSemi',
        },
        {
          line: 3,
          messageId: 'extraSemi',
        },
      ],
    },
    {
      code: 'class PanCamera extends FreeCamera {\n  public invertY: boolean = false\n}',
      options: ['always'],
      output: 'class PanCamera extends FreeCamera {\n  public invertY: boolean = false;\n}',
      errors: [
        {
          line: 2,
          messageId: 'missingSemi',
        },
      ],
    },
    {
      code: 'class PanCamera extends FreeCamera {\n  public invertY: boolean = false;\n}',
      options: ['never'],
      output: 'class PanCamera extends FreeCamera {\n  public invertY: boolean = false\n}',
      errors: [
        {
          line: 2,
          messageId: 'extraSemi',
        },
      ],
    },
    {
      code: 'class C {\n  accessor foo\n  accessor [bar]\n}',
      options: ['always'],
      output: 'class C {\n  accessor foo;\n  accessor [bar];\n}',
      errors: [
        {
          line: 2,
          messageId: 'missingSemi',
        },
        {
          line: 3,
          messageId: 'missingSemi',
        },
      ],
    },
    {
      code: 'class C {\n  accessor foo;\n  accessor [bar];\n}',
      options: ['never'],
      output: 'class C {\n  accessor foo\n  accessor [bar]\n}',
      errors: [
        {
          line: 2,
          messageId: 'extraSemi',
        },
        {
          line: 3,
          messageId: 'extraSemi',
        },
      ],
    },
    {
      code: 'class C {\n  accessor foo\n  accessor [bar]\n}',
      options: ['always'],
      output: 'class C {\n  accessor foo;\n  accessor [bar];\n}',
      errors: [
        {
          line: 2,
          messageId: 'missingSemi',
        },
        {
          line: 3,
          messageId: 'missingSemi',
        },
      ],
    },
    {
      code: 'class C {\n  accessor foo;\n  accessor [bar];\n}',
      options: ['never'],
      output: 'class C {\n  accessor foo\n  accessor [bar]\n}',
      errors: [
        {
          line: 2,
          messageId: 'extraSemi',
        },
        {
          line: 3,
          messageId: 'extraSemi',
        },
      ],
    },
  ],
});

/**
 * ============================ semi — KNOWN GAPS ============================
 *
 * The case below is ported verbatim from upstream but is NOT run through the
 * green `ruleTester.run` above, because it is a *harness/coverage* gap rather
 * than a `@stylistic/semi` mismatch: upstream expects TWO diagnostics for this
 * fixture — one `extraSemi` from `@stylistic/semi` plus one "Unnecessary
 * semicolon." from the core `no-extra-semi` rule (enabled inline via the
 * `/*eslint no-extra-semi: error *\/` directive). The alignment RuleTester
 * mounts ONLY `@stylistic/semi`, and rslint's CLI does not honor inline
 * `/*eslint ... *\/` rule-config directives to turn on an unrelated rule, so
 * rslint emits only the single `extraSemi` diagnostic. Expected count 2 vs
 * actual 1. The semi rule logic itself is correct; the second diagnostic comes
 * from a rule that is not part of this rule's test surface.
 *
 * ---- invalid (upstream expects 2 diagnostics; rslint produces 1) ----
 *
 *   // https://github.com/eslint/eslint/issues/7928
 *   {
 *     code: '/*eslint no-extra-semi: error *\/\nfoo();\n;[0,1,2].forEach(bar)',
 *     output: '/*eslint no-extra-semi: error *\/\nfoo()\n;[0,1,2].forEach(bar)',
 *     options: ['never'],
 *     errors: [
 *       { messageId: 'extraSemi', line: 2, column: 6, endLine: 2, endColumn: 7 },        // from @stylistic/semi
 *       { message: 'Unnecessary semicolon.', line: 3, column: 1, endLine: 3, endColumn: 2 }, // from core no-extra-semi
 *     ],
 *   }
 *
 *   rslint: emits only the `extraSemi` diagnostic at 2:6–2:7 (the semi rule's
 *   own report). The `no-extra-semi` "Unnecessary semicolon." at 3:1–3:2 is not
 *   produced because that rule is not enabled in the generated config.
 *
 * ==========================================================================
 */

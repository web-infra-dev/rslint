import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-promise-executor-return', {
  valid: [
    // --- Not a promise executor ---
    'function foo(resolve, reject) { return 1; }',
    'function Promise(resolve, reject) { return 1; }',
    '(function (resolve, reject) { return 1; })',
    '(function foo(resolve, reject) { return 1; })',
    '(function Promise(resolve, reject) { return 1; })',
    'var foo = function (resolve, reject) { return 1; }',
    'var foo = function Promise(resolve, reject) { return 1; }',
    'var Promise = function (resolve, reject) { return 1; }',
    '(resolve, reject) => { return 1; }',
    '(resolve, reject) => 1',
    'var foo = (resolve, reject) => { return 1; }',
    'var Promise = (resolve, reject) => { return 1; }',
    'var foo = (resolve, reject) => 1',
    'var Promise = (resolve, reject) => 1',
    'var foo = { bar(resolve, reject) { return 1; } }',
    'var foo = { Promise(resolve, reject) { return 1; } }',
    'new foo(function (resolve, reject) { return 1; });',
    'new foo(function bar(resolve, reject) { return 1; });',
    'new foo(function Promise(resolve, reject) { return 1; });',
    'new foo((resolve, reject) => { return 1; });',
    'new foo((resolve, reject) => 1);',
    'new promise(function foo(resolve, reject) { return 1; });',
    'new Promise.foo(function foo(resolve, reject) { return 1; });',
    'new foo.Promise(function foo(resolve, reject) { return 1; });',
    'new Promise.Promise(function foo(resolve, reject) { return 1; });',
    'new Promise()(function foo(resolve, reject) { return 1; });',

    // --- Promise() without new ---
    'Promise(function (resolve, reject) { return 1; });',
    'Promise((resolve, reject) => { return 1; });',
    'Promise((resolve, reject) => 1);',

    // --- Not the first argument ---
    'new Promise(foo, function (resolve, reject) { return 1; });',
    'new Promise(foo, (resolve, reject) => { return 1; });',
    'new Promise(foo, (resolve, reject) => 1);',

    // --- Global Promise doesn't exist ---
    '/* globals Promise:off */ new Promise(function (resolve, reject) { return 1; });',
    {
      code: 'new Promise((resolve, reject) => { return 1; });',
      languageOptions: { globals: { Promise: 'off' } },
    },

    // --- Global Promise is shadowed ---
    'let Promise; new Promise(function (resolve, reject) { return 1; });',
    'function f() { new Promise((resolve, reject) => { return 1; }); var Promise; }',
    'function f(Promise) { new Promise((resolve, reject) => 1); }',
    'if (x) { const Promise = foo(); new Promise(function (resolve, reject) { return 1; }); }',
    'x = function Promise() { new Promise((resolve, reject) => { return 1; }); }',

    // --- return without a value is allowed ---
    'new Promise(function (resolve, reject) { return; });',
    'new Promise(function (resolve, reject) { reject(new Error()); return; });',
    'new Promise(function (resolve, reject) { if (foo) { return; } });',
    'new Promise((resolve, reject) => { return; });',
    'new Promise((resolve, reject) => { if (foo) { resolve(1); return; } reject(new Error()); });',

    // --- throw is allowed ---
    'new Promise(function (resolve, reject) { throw new Error(); });',
    'new Promise((resolve, reject) => { throw new Error(); });',

    // --- Not returning from the promise executor ---
    'new Promise(function (resolve, reject) { function foo() { return 1; } });',
    'new Promise((resolve, reject) => { (function foo() { return 1; })(); });',
    'new Promise(function (resolve, reject) { () => { return 1; } });',
    'new Promise((resolve, reject) => { () => 1 });',
    'function foo() { return new Promise(function (resolve, reject) { resolve(bar); }) };',
    'foo => new Promise((resolve, reject) => { bar(foo, (err, data) => { if (err) { reject(err); return; } resolve(data); })});',

    // --- Promise executors do not affect other functions ---
    'new Promise(function (resolve, reject) {}); function foo() { return 1; }',
    'new Promise((resolve, reject) => {}); (function () { return 1; });',
    'new Promise(function (resolve, reject) {}); () => { return 1; };',
    'new Promise((resolve, reject) => {}); () => 1;',

    // --- Does not report global return ---
    'return 1;',
    'return 1; function foo(){ return 1; } return 1;',
    'function foo(){} return 1; var bar = function*(){ return 1; }; return 1; var baz = () => {}; return 1;',
    'new Promise(function (resolve, reject) {}); return 1;',

    // --- allowVoid: true ---
    { code: 'new Promise((r) => void cbf(r));', options: { allowVoid: true } },
    { code: 'new Promise(r => void 0)', options: { allowVoid: true } },
    {
      code: 'new Promise(r => { return void 0 })',
      options: { allowVoid: true },
    },
    {
      code: 'new Promise(r => { if (foo) { return void 0 } return void 0 })',
      options: { allowVoid: true },
    },
    'new Promise(r => {0})',
  ],
  invalid: [
    {
      code: 'new Promise(function (resolve, reject) { return 1; })',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 42,
          endLine: 1,
          endColumn: 51,
        },
      ],
    },
    {
      code: 'new Promise((resolve, reject) => resolve(1))',
      options: { allowVoid: true },
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 34,
          endLine: 1,
          endColumn: 44,
        },
      ],
    },
    {
      code: 'new Promise((resolve, reject) => { return 1 })',
      options: { allowVoid: true },
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 36,
          endLine: 1,
          endColumn: 44,
        },
      ],
    },
    {
      code: 'new Promise(r => 1)',
      options: { allowVoid: true },
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'new Promise(r => 1 ? 2 : 3)',
      options: { allowVoid: true },
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: 'new Promise(r => (1 ? 2 : 3))',
      options: { allowVoid: true },
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: 'new Promise(r => (1))',
      options: { allowVoid: true },
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: 'new Promise(r => () => {})',
      options: { allowVoid: true },
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: 'new Promise(r => null)',
      options: { allowVoid: true },
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'new Promise(r => null)',
      options: { allowVoid: false },
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'new Promise(r => /*hi*/ ~0)',
      options: { allowVoid: true },
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: 'new Promise(r => /*hi*/ ~0)',
      options: { allowVoid: false },
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: 'new Promise(r => { return 0 })',
      options: { allowVoid: true },
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: 'new Promise(r => { return 0 })',
      options: { allowVoid: false },
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: 'new Promise(r => { if (foo) { return void 0 } return 0 })',
      options: { allowVoid: true },
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 47,
          endLine: 1,
          endColumn: 55,
        },
      ],
    },
    {
      code: 'new Promise(resolve => { return (foo = resolve(1)); })',
      options: { allowVoid: true },
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 52,
        },
      ],
    },
    {
      code: 'new Promise(resolve => r = resolve)',
      options: { allowVoid: true },
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: 'new Promise(r => { return(1) })',
      options: { allowVoid: true },
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: 'new Promise(r =>1)',
      options: { allowVoid: true },
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'new Promise(r => ((1)))',
      options: { allowVoid: true },
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: 'new Promise(function foo(resolve, reject) { return 1; })',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 45,
          endLine: 1,
          endColumn: 54,
        },
      ],
    },
    {
      code: 'new Promise((resolve, reject) => { return 1; })',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 36,
          endLine: 1,
          endColumn: 45,
        },
      ],
    },
    {
      code: 'new Promise(function (resolve, reject) { return undefined; })',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 42,
          endLine: 1,
          endColumn: 59,
        },
      ],
    },
    {
      code: 'new Promise((resolve, reject) => { return null; })',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 36,
          endLine: 1,
          endColumn: 48,
        },
      ],
    },
    {
      code: 'new Promise(function (resolve, reject) { return false; })',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 42,
          endLine: 1,
          endColumn: 55,
        },
      ],
    },
    {
      code: 'new Promise((resolve, reject) => resolve)',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 34,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      code: 'new Promise((resolve, reject) => null)',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 34,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      code: 'new Promise(function (resolve, reject) { return resolve(foo); })',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 42,
          endLine: 1,
          endColumn: 62,
        },
      ],
    },
    {
      code: 'new Promise((resolve, reject) => { return reject(foo); })',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 36,
          endLine: 1,
          endColumn: 55,
        },
      ],
    },
    {
      code: 'new Promise((resolve, reject) => x + y)',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 34,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      code: 'new Promise((resolve, reject) => { return Promise.resolve(42); })',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 36,
          endLine: 1,
          endColumn: 63,
        },
      ],
    },
    {
      code: 'new Promise(function (resolve, reject) { if (foo) { return 1; } })',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 53,
          endLine: 1,
          endColumn: 62,
        },
      ],
    },
    {
      code: 'new Promise((resolve, reject) => { try { return 1; } catch(e) {} })',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 42,
          endLine: 1,
          endColumn: 51,
        },
      ],
    },
    {
      code: 'new Promise(function (resolve, reject) { while (foo){ if (bar) break; else return 1; } })',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 76,
          endLine: 1,
          endColumn: 85,
        },
      ],
    },
    {
      code: 'new Promise(() => { return void 1; })',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: 'new Promise(() => (1))',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '() => new Promise(() => ({}));',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: 'new Promise(function () { return 1; })',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      code: 'new Promise(() => { return 1; })',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: 'new Promise(() => 1)',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: 'function foo() {} new Promise(function () { return 1; });',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 45,
          endLine: 1,
          endColumn: 54,
        },
      ],
    },
    {
      code: 'function foo() { return; } new Promise(() => { return 1; });',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 48,
          endLine: 1,
          endColumn: 57,
        },
      ],
    },
    {
      code: 'function foo() { return 1; } new Promise(() => { return 2; });',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 50,
          endLine: 1,
          endColumn: 59,
        },
      ],
    },
    {
      code: 'function foo () { return new Promise(function () { return 1; }); }',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 52,
          endLine: 1,
          endColumn: 61,
        },
      ],
    },
    {
      code: 'function foo() { return new Promise(() => { bar(() => { return 1; }); return false; }); }',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 71,
          endLine: 1,
          endColumn: 84,
        },
      ],
    },
    {
      code: '() => new Promise(() => { if (foo) { return 0; } else bar(() => { return 1; }); })',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 38,
          endLine: 1,
          endColumn: 47,
        },
      ],
    },
    {
      code: 'function foo () { return 1; return new Promise(function () { return 2; }); return 3;}',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 62,
          endLine: 1,
          endColumn: 71,
        },
      ],
    },
    {
      code: '() => 1; new Promise(() => { return 1; })',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      code: 'new Promise(function () { return 1; }); function foo() { return 1; } ',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      code: '() => new Promise(() => { return 1; });',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      code: '() => new Promise(() => 1);',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '() => new Promise(() => () => 1);',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: '() => new Promise(() => async () => 1);',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      code: '() => new Promise(() => function () {});',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      code: '() => new Promise(() => class {});',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: '() => new Promise(() => function foo() {});',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      code: '() => new Promise(() => class Foo {});',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: '() => new Promise(() => []);',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: 'new Promise((Promise) => { return 1; })',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: 'new Promise(function Promise(resolve, reject) { return 1; })',
      errors: [
        {
          messageId: 'returnsValue',
          line: 1,
          column: 49,
          endLine: 1,
          endColumn: 58,
        },
      ],
    },
  ],
});

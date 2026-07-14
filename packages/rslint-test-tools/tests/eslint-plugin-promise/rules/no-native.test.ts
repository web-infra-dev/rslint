import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const message = '"Promise" is not defined.';

ruleTester.run('no-native', {} as never, {
  valid: [
    {
      code: 'var Promise = null; function x() { return Promise.resolve("hi"); }',
    },
    {
      code: 'var Promise = window.Promise || require("bluebird"); var x = Promise.reject();',
    },
    { code: 'import Promise from "bluebird"; var x = Promise.reject();' },
    { code: 'function f(Promise) { return Promise.resolve(1); }' },
    // a type reference resolved against a local type declaration
    { code: 'type Promise = string; let x: Promise;' },
    { code: 'interface I { m<Promise>(x: Promise): void }' },
    { code: 'type M<T> = { [Promise in keyof T]: Promise };' },
    {
      code: 'type U<T> = T extends Array<infer Promise> ? Promise : never;',
    },
    {
      code: 'const C = class Promise { m(): Promise { return this; } };',
    },
  ],

  invalid: [
    {
      code: 'new Promise(function(reject, resolve) { })',
      errors: [{ message }],
    },
    {
      code: 'Promise.resolve()',
      errors: [{ message }],
    },
    {
      code: '/* global Promise */ Promise.resolve()',
      errors: [{ message }],
    },
    {
      code: 'Promise.all([]); Promise.resolve(1);',
      errors: [{ message }, { message }],
    },
    // a type-only declaration does not satisfy a value reference
    {
      code: 'interface Promise {} Promise.resolve(1);',
      errors: [{ message }],
    },
    // a value declaration does not shadow a type reference
    {
      code: 'var Promise = 1; let y: Promise;',
      errors: [{ message }],
    },
  ],
});

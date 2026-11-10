// Upstream: https://github.com/eslint/eslint/blob/v10.8.1/tests/lib/rules/no-async-promise-executor.js
/**
 * @fileoverview disallow using an async function as a Promise executor
 * @author Teddy Katz
 */

//------------------------------------------------------------------------------
// Requirements
//------------------------------------------------------------------------------

import { RuleTester } from '../rule-tester';

//------------------------------------------------------------------------------
// Tests
//------------------------------------------------------------------------------

const ruleTester = new RuleTester({ parserOptions: { ecmaVersion: 8 } });

ruleTester.run('no-async-promise-executor', {
  valid: [
    'new Promise((resolve, reject) => {})',
    'new Promise((resolve, reject) => {}, async function unrelated() {})',
    'new Foo(async (resolve, reject) => {})',
    '/* global Promise:off */ new Promise(async (resolve, reject) => {})',
    {
      code: 'new Promise(async (resolve, reject) => {})',
      languageOptions: {
        globals: { Promise: 'off' },
      },
    },
    'let Promise; new Promise(async (resolve, reject) => {})',
    'function f() { new Promise(async (resolve, reject) => {}); var Promise; }',
    'function f(Promise) { new Promise(async (resolve, reject) => {}); }',
  ],

  invalid: [
    {
      code: 'new Promise(async function foo(resolve, reject) {})',
      errors: [
        {
          messageId: 'async',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'new Promise(async (resolve, reject) => {})',
      errors: [
        {
          messageId: 'async',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'new Promise(((((async () => {})))))',
      errors: [
        {
          messageId: 'async',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
  ],
});

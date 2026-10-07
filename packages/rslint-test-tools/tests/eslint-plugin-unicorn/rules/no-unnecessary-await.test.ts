// Ported from eslint-plugin-unicorn v77.0.0 tests and documentation.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/test/no-unnecessary-await.js
import path from 'node:path';
import { expect, test } from 'rstack/test';
import type { RslintConfigEntry } from '@rslint/core';
import { lint } from '@rslint/core/internal';
import { RuleTester } from '../rule-tester';

const cases = {
  valid: [
    // JavaScript
    { code: 'await {then}', filename: 'src/virtual.js' },
    { code: 'await a ? b : c', filename: 'src/virtual.js' },
    { code: 'await a || b', filename: 'src/virtual.js' },
    { code: 'await a && b', filename: 'src/virtual.js' },
    { code: 'await a ?? b', filename: 'src/virtual.js' },
    { code: 'await new Foo()', filename: 'src/virtual.js' },
    { code: 'await tagged``', filename: 'src/virtual.js' },
    {
      code: 'class A { async foo() { await this }}',
      filename: 'src/virtual.js',
    },
    {
      code: 'async function * foo() {await (yield bar);}',
      filename: 'src/virtual.js',
    },
    { code: 'await (1, Promise.resolve())', filename: 'src/virtual.js' },
    // TypeScript
    {
      code: 'async function f() { return await (a as Promise<number>); }',
      filename: 'src/virtual.ts',
    },
    {
      code: 'async function f() { return await (a!); }',
      filename: 'src/virtual.ts',
    },
    // Documentation
    { code: 'await promise;', filename: 'src/virtual.js' },
    {
      code: 'await Promise.allSettled([promise1, promise2]);',
      filename: 'src/virtual.js',
    },
  ],
  invalid: [
    // JavaScript
    {
      code: 'await []',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
      output: '[]',
    },
    {
      code: 'await [Promise.resolve()]',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
      output: '[Promise.resolve()]',
    },
    {
      code: 'await (() => {})',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
      output: '(() => {})',
    },
    {
      code: 'await (() => Promise.resolve())',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
      output: '(() => Promise.resolve())',
    },
    {
      code: 'await (a === b)',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
      output: '(a === b)',
    },
    {
      code: 'await (a instanceof Promise)',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
      output: '(a instanceof Promise)',
    },
    {
      code: 'await (a > b)',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
      output: '(a > b)',
    },
    {
      code: 'await class {}',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
      output: null,
    },
    {
      code: 'await class extends Promise {}',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
      output: null,
    },
    {
      code: 'await function() {}',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
      output: null,
    },
    {
      code: 'await function name() {}',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
      output: null,
    },
    {
      code: 'await function() { return Promise.resolve() }',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
      output: null,
    },
    {
      code: 'await (<></>)',
      filename: 'src/virtual.jsx',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
      output: '(<></>)',
    },
    {
      code: 'await (<a></a>)',
      filename: 'src/virtual.jsx',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
      output: '(<a></a>)',
    },
    {
      code: 'await 0',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
      output: '0',
    },
    {
      code: 'await 1',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
      output: '1',
    },
    {
      code: 'await ""',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
      output: '""',
    },
    {
      code: 'await "string"',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
      output: '"string"',
    },
    {
      code: 'await true',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
      output: 'true',
    },
    {
      code: 'await false',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
      output: 'false',
    },
    {
      code: 'await null',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
      output: 'null',
    },
    {
      code: 'await 0n',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
      output: '0n',
    },
    {
      code: 'await 1n',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
      output: '1n',
    },
    {
      code: 'await `${Promise.resolve()}`',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
      output: '`${Promise.resolve()}`',
    },
    {
      code: 'await !Promise.resolve()',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
      output: '!Promise.resolve()',
    },
    {
      code: 'await void Promise.resolve()',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
      output: 'void Promise.resolve()',
    },
    {
      code: 'await +Promise.resolve()',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
      output: '+Promise.resolve()',
    },
    {
      code: 'await ~1',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
      output: '~1',
    },
    {
      code: 'await ++foo',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
      output: '++foo',
    },
    {
      code: 'await foo--',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
      output: 'foo--',
    },
    {
      code: 'await (Promise.resolve(), 1)',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
      output: '(Promise.resolve(), 1)',
    },
    {
      code: 'async function foo() {\n\treturn await\n\t\t// comment\n\t\t1;\n}',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 14,
        },
      ],
      output: 'async function foo() {\n\treturn ( // comment\n\t\t1);\n}',
    },
    {
      code: 'async function foo() {\n\treturn await\n\t\t// comment\n\t\t1\n}',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 14,
        },
      ],
      output: 'async function foo() {\n\treturn ( // comment\n\t\t1)\n}',
    },
    {
      code: 'async function foo() {\n\treturn( await\n\t\t// comment\n\t\t1);\n}',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 2,
          column: 10,
          endLine: 2,
          endColumn: 15,
        },
      ],
      output: 'async function foo() {\n\treturn( // comment\n\t\t1);\n}',
    },
    {
      code: 'foo()\nawait []',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 6,
        },
      ],
      output: 'foo()\n;[]',
    },
    {
      code: 'foo()\nawait +1',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 6,
        },
      ],
      output: 'foo()\n;+1',
    },
    {
      code: 'async function foo() {\n\treturn await\n\t\t// comment\n\t\t[];\n}',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 14,
        },
      ],
      output: 'async function foo() {\n\treturn ( // comment\n\t\t[]);\n}',
    },
    {
      code: 'async function foo() {\n\tthrow await\n\t\t// comment\n\t\t1;\n}',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 2,
          column: 8,
          endLine: 2,
          endColumn: 13,
        },
      ],
      output: 'async function foo() {\n\tthrow ( // comment\n\t\t1);\n}',
    },
    {
      code: 'console.log(\n\tawait\n\t\t// comment\n\t\t[]\n);',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 2,
          column: 2,
          endLine: 2,
          endColumn: 7,
        },
      ],
      output: null,
    },
    {
      code: 'async function foo() {+await +1}',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 29,
        },
      ],
      output: null,
    },
    {
      code: 'async function foo() {-await-1}',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 29,
        },
      ],
      output: null,
    },
    {
      code: 'async function foo() {+await -1}',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 29,
        },
      ],
      output: null,
    },
    {
      code: 'async function foo() {+await ++bar}',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 29,
        },
      ],
      output: null,
    },
    {
      code: 'async function foo() {-await --bar}',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 29,
        },
      ],
      output: null,
    },
    {
      code: 'async function foo() {const a = +await ++b;}',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 34,
          endLine: 1,
          endColumn: 39,
        },
      ],
      output: null,
    },
    {
      code: 'async function foo() {+await --bar}',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 29,
        },
      ],
      output: null,
    },
    {
      code: 'async function foo() {-await ++bar}',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 29,
        },
      ],
      output: null,
    },
    {
      code: 'async function foo() {+await bar++}',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 29,
        },
      ],
      output: null,
    },
    {
      code: 'async function foo() {~await ~1}',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 29,
        },
      ],
      output: null,
    },
    // TypeScript
    {
      code: 'async function f() { return await (1 as number); }',
      filename: 'src/virtual.ts',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 34,
        },
      ],
      output: 'async function f() { return (1 as number); }',
    },
    {
      code: 'async function f() { return await ([1, 2]!); }',
      filename: 'src/virtual.ts',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 34,
        },
      ],
      output: 'async function f() { return ([1, 2]!); }',
    },
    {
      code: 'async function f() { return await ([1, 2] as const); }',
      filename: 'src/virtual.ts',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 34,
        },
      ],
      output: 'async function f() { return ([1, 2] as const); }',
    },
    {
      code: 'async function f() { return await ([1, 2] satisfies number[]); }',
      filename: 'src/virtual.ts',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 34,
        },
      ],
      output: 'async function f() { return ([1, 2] satisfies number[]); }',
    },
    // Microtask ordering
    {
      code: 'async function f() { log("s"); await 1; log("e"); }',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 37,
        },
      ],
      output: null,
    },
    {
      code: 'async function f() { await 1; log("e"); }',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 27,
        },
      ],
      output: null,
    },
    {
      code: 'async function f() { if (q) { run(); } await 1; }',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 40,
          endLine: 1,
          endColumn: 45,
        },
      ],
      output: 'async function f() { if (q) { run(); } 1; }',
    },
    {
      code: 'async function f() { log("s"); await 1; }',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 37,
        },
      ],
      output: 'async function f() { log("s"); 1; }',
    },
    {
      code: 'async function f() { await 1; }',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 27,
        },
      ],
      output: 'async function f() { 1; }',
    },
    {
      code: 'async function f() { return await 1; }',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 34,
        },
      ],
      output: 'async function f() { return 1; }',
    },
    {
      code: 'async function f() { const x = await 1; }',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 37,
        },
      ],
      output: 'async function f() { const x = 1; }',
    },
    {
      code: 'async function f() { if (q) { await 1; } }',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 36,
        },
      ],
      output: 'async function f() { if (q) { 1; } }',
    },
    {
      code: 'const f = async () => { await 1; };',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 30,
        },
      ],
      output: 'const f = async () => { 1; };',
    },
    {
      code: 'async function f() { if (q) { await 1; log("e"); } }',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 36,
        },
      ],
      output: null,
    },
    {
      code: 'async function f() { log(await 1); }',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 31,
        },
      ],
      output: null,
    },
    {
      code: 'async function f() { for (const x of xs) { await 1; } }',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 44,
          endLine: 1,
          endColumn: 49,
        },
      ],
      output: null,
    },
    {
      code: 'async function f() { try { await 1; } finally { log("e"); } }',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 33,
        },
      ],
      output: null,
    },
    {
      code: 'async function outer() { return async () => (await 1, log("e")); }',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 46,
          endLine: 1,
          endColumn: 51,
        },
      ],
      output: null,
    },
    {
      code: 'async function outer() { const f = async () => await 1; run(); }',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 48,
          endLine: 1,
          endColumn: 53,
        },
      ],
      output: 'async function outer() { const f = async () => 1; run(); }',
    },
    {
      code: 'Promise.resolve().then(() => log("a")); await 1; log("b");',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 41,
          endLine: 1,
          endColumn: 46,
        },
      ],
      output: null,
    },
    {
      code: 'run(); await 1;',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 13,
        },
      ],
      output: 'run(); 1;',
    },
    // Documentation
    {
      code: 'await await promise;',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
      output: 'await promise;',
    },
    {
      code: 'await [promise1, promise2];',
      filename: 'src/virtual.js',
      errors: [
        {
          messageId: 'no-unnecessary-await',
          message: 'Do not `await` non-promise value.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
      output: '[promise1, promise2];',
    },
  ],
};

// Match upstream's module parsing without changing the source or JS file type.
const moduleOptions: RslintConfigEntry['languageOptions'] = {
  sourceType: 'module',
  parserOptions: {
    projectService: false,
    project: [
      path.resolve(
        import.meta.dirname,
        '../fixtures/no-unnecessary-await/tsconfig.json',
      ),
    ],
  },
};
const unsupportedJavaScriptCase = cases.invalid.find(
  ({ code }) => code === 'await !Promise.resolve()',
)!;

new RuleTester().run('no-unnecessary-await', {} as never, {
  valid: cases.valid.map((testCase) => ({
    ...testCase,
    languageOptions: moduleOptions,
  })),
  // Preserve the known unsupported JS case below as an explicit skipped test.
  invalid: cases.invalid
    .filter((testCase) => testCase !== unsupportedJavaScriptCase)
    .map((testCase) => ({ ...testCase, languageOptions: moduleOptions })),
});

function javascriptRequest(
  code: string,
  languageOptions: RslintConfigEntry['languageOptions'],
) {
  return {
    config: [
      {
        plugins: ['unicorn'],
        languageOptions,
        rules: { 'unicorn/no-unnecessary-await': 'error' as const },
      },
    ],
    configDirectory: import.meta.dirname,
    workingDirectory: import.meta.dirname,
    fileContents: {
      [path.resolve(import.meta.dirname, '../src/virtual.js')]: code,
    },
  };
}

// Even forced module detection currently rejects bare top-level JS `await !`.
test.skip('upstream JavaScript await !Promise.resolve() (TS8013)', async () => {
  const result = await lint(
    javascriptRequest(unsupportedJavaScriptCase.code, moduleOptions),
  );
  expect(result.diagnostics).toMatchObject([
    { ruleName: 'unicorn/no-unnecessary-await' },
  ]);
});

const sourceOnlyOptions: RslintConfigEntry['languageOptions'] = {
  sourceType: 'module',
  parserOptions: { project: false, projectService: false },
};

test.each([
  { code: 'await []', rules: ['TypeScript(TS1011)'], output: '[]' },
  { code: 'await (a + b)', rules: [], output: '(a + b)' },
])(
  'documents JavaScript module detection for $code',
  async ({ code, rules, output }) => {
    const result = await lint(javascriptRequest(code, sourceOnlyOptions));
    expect(result.fileCount).toBe(1);
    expect(result.diagnostics.map(({ ruleName }) => ruleName)).toEqual(rules);

    for (const [source, options, expectedOutput] of [
      [code, moduleOptions, output],
      [`export {};\n${code}`, sourceOnlyOptions, `export {};\n${output}`],
    ] as const) {
      const request = javascriptRequest(source, options);
      const checked = await lint(request);
      expect(checked.fileCount).toBe(1);
      expect(checked.diagnostics).toMatchObject([
        {
          ruleName: 'unicorn/no-unnecessary-await',
          messageId: 'no-unnecessary-await',
        },
      ]);
      const fixed = await lint({ ...request, fix: true });
      expect(fixed.diagnostics).toEqual([]);
      expect(Object.values(fixed.output ?? {})).toEqual([expectedOutput]);
    }
  },
);

test('documents the JavaScript await ! limitation even in a module', async () => {
  const result = await lint(
    javascriptRequest(unsupportedJavaScriptCase.code, moduleOptions),
  );
  expect(result.fileCount).toBe(1);
  expect(result.diagnostics).toMatchObject([
    { ruleName: 'TypeScript(TS8013)' },
  ]);
});

/**
 * Complete ESLint v10.8.1 no-await-in-loop suite (20 valid, 17 invalid).
 * https://github.com/eslint/eslint/blob/v10.8.1/tests/lib/rules/no-await-in-loop.js
 * Inputs and language options are unchanged. Full messages and ranges are
 * captured from that release; the shared tester also rejects fixes/suggestions.
 */
import assert from 'node:assert/strict';
import { mkdtemp, realpath, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { test } from 'rstack/test';
import { lint } from '@rslint/core/internal';
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';

const ruleTester = new RuleTester({
  plugin: '',
  languageOptions: { ecmaVersion: 2018, sourceType: 'script' },
});

const cases = {
  valid: [
    'async function foo() { await bar; }',
    'async function foo() { for (var bar in await baz) { } }',
    'async function foo() { for (var bar of await baz) { } }',
    'async function foo() { for await (var bar of await baz) { } }',
    'async function foo() { for (var bar = await baz in qux) {} }',

    // While loops
    'async function foo() { while (true) { async function foo() { await bar; } } }', // Blocked by a function declaration
    // For loops
    'async function foo() { for (var i = await bar; i < n; i++) {  } }',

    // Do while loops
    'async function foo() { do { } while (bar); }',

    // Blocked by a function expression
    'async function foo() { while (true) { var y = async function() { await bar; } } }',

    // Blocked by an arrow function
    'async function foo() { while (true) { var y = async () => await foo; } }',
    'async function foo() { while (true) { var y = async () => { await foo; } } }',

    // Blocked by a class method
    'async function foo() { while (true) { class Foo { async foo() { await bar; } } } }',

    // Asynchronous iteration intentionally
    'async function foo() { for await (var x of xs) { await f(x) } }',

    // Explicit Resource Management
    'while (true) { const value = 0; }',
    'while (true) { let value = 0; }',
    'while (true) { var value = 0; }',
    {
      code: 'await using resource = getResource();',
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2026,
      },
    },
    {
      code: 'while (true) { using resource = getResource(); }',
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2026,
      },
    },
    {
      code: 'async function foo() { while (true) { async function foo() { await using resource = getResource(); } } }',
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2026,
      },
    },
    {
      code: 'for (await using resource = getResource(); ;) {}',
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2026,
      },
    },
  ],
  invalid: [
    // While loops
    {
      code: 'async function foo() { while (baz) { await bar; } }',
      errors: [
        {
          messageId: 'unexpectedAwait',
          message: 'Unexpected `await` inside a loop.',
          line: 1,
          column: 38,
          endLine: 1,
          endColumn: 47,
        },
      ],
    },
    {
      code: 'async function foo() { while (await foo()) {  } }',
      errors: [
        {
          messageId: 'unexpectedAwait',
          message: 'Unexpected `await` inside a loop.',
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      code: 'async function foo() { while (baz) { for await (x of xs); } }',
      errors: [
        {
          messageId: 'unexpectedAwait',
          message: 'Unexpected `await` inside a loop.',
          line: 1,
          column: 38,
          endLine: 1,
          endColumn: 58,
        },
      ],
    },

    // For of loops
    {
      code: 'async function foo() { for (var bar of baz) { await bar; } }',
      errors: [
        {
          messageId: 'unexpectedAwait',
          message: 'Unexpected `await` inside a loop.',
          line: 1,
          column: 47,
          endLine: 1,
          endColumn: 56,
        },
      ],
    },
    {
      code: 'async function foo() { for (var bar of baz) await bar; }',
      errors: [
        {
          messageId: 'unexpectedAwait',
          message: 'Unexpected `await` inside a loop.',
          line: 1,
          column: 45,
          endLine: 1,
          endColumn: 54,
        },
      ],
    },

    // For in loops
    {
      code: 'async function foo() { for (var bar in baz) { await bar; } }',
      errors: [
        {
          messageId: 'unexpectedAwait',
          message: 'Unexpected `await` inside a loop.',
          line: 1,
          column: 47,
          endLine: 1,
          endColumn: 56,
        },
      ],
    },

    // For loops
    {
      code: 'async function foo() { for (var i; i < n; i++) { await bar; } }',
      errors: [
        {
          messageId: 'unexpectedAwait',
          message: 'Unexpected `await` inside a loop.',
          line: 1,
          column: 50,
          endLine: 1,
          endColumn: 59,
        },
      ],
    },
    {
      code: 'async function foo() { for (var i; await foo(i); i++) {  } }',
      errors: [
        {
          messageId: 'unexpectedAwait',
          message: 'Unexpected `await` inside a loop.',
          line: 1,
          column: 36,
          endLine: 1,
          endColumn: 48,
        },
      ],
    },
    {
      code: 'async function foo() { for (var i; i < n; i = await bar) {  } }',
      errors: [
        {
          messageId: 'unexpectedAwait',
          message: 'Unexpected `await` inside a loop.',
          line: 1,
          column: 47,
          endLine: 1,
          endColumn: 56,
        },
      ],
    },

    // Do while loops
    {
      code: 'async function foo() { do { await bar; } while (baz); }',
      errors: [
        {
          messageId: 'unexpectedAwait',
          message: 'Unexpected `await` inside a loop.',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      code: 'async function foo() { do { } while (await bar); }',
      errors: [
        {
          messageId: 'unexpectedAwait',
          message: 'Unexpected `await` inside a loop.',
          line: 1,
          column: 38,
          endLine: 1,
          endColumn: 47,
        },
      ],
    },

    // Deep in a loop body
    {
      code: 'async function foo() { while (true) { if (bar) { foo(await bar); } } }',
      errors: [
        {
          messageId: 'unexpectedAwait',
          message: 'Unexpected `await` inside a loop.',
          line: 1,
          column: 54,
          endLine: 1,
          endColumn: 63,
        },
      ],
    },

    // Deep in a loop condition
    {
      code: 'async function foo() { while (xyz || 5 > await x) {  } }',
      errors: [
        {
          messageId: 'unexpectedAwait',
          message: 'Unexpected `await` inside a loop.',
          line: 1,
          column: 42,
          endLine: 1,
          endColumn: 49,
        },
      ],
    },

    // In a nested loop of for-await-of
    {
      code: 'async function foo() { for await (var x of xs) { while (1) await f(x) } }',
      errors: [
        {
          messageId: 'unexpectedAwait',
          message: 'Unexpected `await` inside a loop.',
          line: 1,
          column: 60,
          endLine: 1,
          endColumn: 70,
        },
      ],
    },

    // Explicit Resource Management
    {
      code: 'while (true) { await using resource = getResource(); }',
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2026,
      },
      errors: [
        {
          messageId: 'unexpectedAwait',
          message: 'Unexpected `await` inside a loop.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
    {
      code: 'for (;;) { await using resource = getResource(); }',
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2026,
      },
      errors: [
        {
          messageId: 'unexpectedAwait',
          message: 'Unexpected `await` inside a loop.',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 49,
        },
      ],
    },
    {
      code: 'for (await using resource of resources) {}',
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2026,
      },
      errors: [
        {
          messageId: 'unexpectedAwait',
          message: 'Unexpected `await` inside a loop.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
  ],
} satisfies { valid: (string | ValidTestCase)[]; invalid: TestCase[] };

ruleTester.run('no-await-in-loop', cases);

// This rule has no autofix. Exercise the actual fix pipeline as well as the
// shared tester's strict checks for absent fixes and suggestions.
test('no-await-in-loop leaves every upstream case unchanged with fix enabled', async () => {
  const root = await realpath(
    await mkdtemp(path.join(tmpdir(), 'rslint-no-await-in-loop-')),
  );
  try {
    for (const entry of [...cases.valid, ...cases.invalid]) {
      const item = typeof entry === 'string' ? { code: entry } : entry;
      const result = await lint({
        configDirectory: root,
        workingDirectory: root,
        config: [
          {
            languageOptions: {
              ecmaVersion: 2018,
              sourceType: 'script',
              ...('languageOptions' in item ? item.languageOptions : {}),
              parserOptions: { projectService: false },
            },
            rules: { 'no-await-in-loop': 'error' },
          },
        ],
        fileContents: { [path.join(root, 'input.js')]: item.code },
        fix: true,
      });
      assert.equal(result.fileCount, 1);
      assert.equal(result.ruleCount, 1);
      assert.deepEqual(result.output ?? {}, {}, item.code);
      assert.equal(
        result.diagnostics.length,
        'errors' in item ? item.errors.length : 0,
        item.code,
      );
      for (const diagnostic of result.diagnostics) {
        assert.equal(diagnostic.fixes, undefined);
        assert.equal(diagnostic.suggestions?.length ?? 0, 0);
      }
    }
  } finally {
    await rm(root, {
      recursive: true,
      force: true,
      maxRetries: 20,
      retryDelay: 1000,
    });
  }
});

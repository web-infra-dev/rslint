// Upstream: eslint-plugin-promise v7.3.0 __tests__/no-native.js and rule docs.
import path from 'node:path';

import { lint } from '@rslint/core/internal';

interface TestCase {
  name: string;
  code: string;
  globals?: Record<string, boolean | 'off'>;
  column?: number;
}

// The suite's shared wrapper does not assert ranges or accept globals.
// Use IPC directly so the upstream global-configuration cases are exercised.
const cases: TestCase[] = [
  {
    name: 'local variable',
    code: 'var Promise = null; function x() { return Promise.resolve("hi"); }',
  },
  {
    name: 'polyfill fallback',
    code: 'var Promise = window.Promise || require("bluebird"); var x = Promise.reject();',
  },
  {
    name: 'import',
    code: 'import Promise from "bluebird"; var x = Promise.reject();',
  },
  {
    name: 'constructor',
    code: 'new Promise(function(reject, resolve) { })',
    column: 5,
  },
  {
    name: 'static method',
    code: 'Promise.resolve()',
    column: 1,
  },
  // Relevant globals from the upstream browser/node environments.
  {
    name: 'browser environment',
    code: 'new Promise(function(reject, resolve) { })',
    globals: { Promise: false, window: false },
    column: 5,
  },
  {
    name: 'node environment',
    code: 'new Promise(function(reject, resolve) { })',
    globals: { Promise: false, global: false },
    column: 5,
  },
  {
    name: 'es6 environment',
    code: 'Promise.resolve()',
    column: 1,
  },
  {
    name: 'writable global',
    code: 'Promise.resolve()',
    globals: { Promise: true },
    column: 1,
  },
  {
    name: 'disabled global',
    code: 'Promise.resolve()',
    globals: { Promise: 'off' },
    column: 1,
  },
  {
    name: 'documentation valid',
    code: "const Promise = require('bluebird')\nconst x = Promise.resolve('good')",
  },
  {
    name: 'documentation invalid',
    code: "const x = Promise.resolve('bad')",
    column: 11,
  },
];

describe.each(['js', 'ts'])('promise/no-native (%s)', (extension) => {
  for (const testCase of cases) {
    test(testCase.name, async () => {
      const filename = path.resolve(
        import.meta.dirname,
        `virtual.${extension}`,
      );
      const result = await lint({
        workingDirectory: import.meta.dirname,
        configDirectory: import.meta.dirname,
        config: [
          {
            plugins: ['promise'],
            languageOptions: {
              ecmaVersion: 2015,
              sourceType: 'module',
              globals: testCase.globals,
            },
            rules: { 'promise/no-native': 'error' },
          },
        ],
        fileContents: { [filename]: testCase.code },
      });

      expect(result.fileCount).toBe(1);
      expect(result.ruleCount).toBe(1);
      const expected = testCase.column
        ? [
            {
              ruleName: 'promise/no-native',
              messageId: 'name',
              message: '"Promise" is not defined.',
              severity: 'error',
              range: {
                start: { line: 1, column: testCase.column },
                end: { line: 1, column: testCase.column + 7 },
              },
              fixes: [],
              suggestions: [],
            },
          ]
        : [];
      expect(
        result.diagnostics.map((diagnostic) => ({
          ruleName: diagnostic.ruleName,
          messageId: diagnostic.messageId,
          message: diagnostic.message,
          severity: diagnostic.severity,
          range: diagnostic.range,
          fixes: diagnostic.fixes ?? [],
          suggestions: diagnostic.suggestions ?? [],
        })),
      ).toEqual(expected);
    });
  }
});

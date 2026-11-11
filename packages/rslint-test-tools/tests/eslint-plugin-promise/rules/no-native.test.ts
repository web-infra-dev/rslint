// Upstream: eslint-plugin-promise v7.3.0 __tests__/no-native.js and rule docs.
import { RuleTester } from '@rslint/rule-tester';
import { describe } from 'rstack/test';

for (const extension of ['js', 'ts']) {
  describe(`promise/no-native (${extension})`, () => {
    new RuleTester({
      plugin: 'promise',
      filename: `virtual.${extension}`,
      languageOptions: { ecmaVersion: 2015, sourceType: 'module' },
    }).run('no-native', {
      valid: [
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
          name: 'documentation valid',
          code: "const Promise = require('bluebird')\nconst x = Promise.resolve('good')",
        },
      ],
      invalid: [
        {
          name: 'constructor',
          code: 'new Promise(function(reject, resolve) { })',
          errors: [
            {
              messageId: 'name',
              message: '"Promise" is not defined.',
              line: 1,
              column: 5,
              endLine: 1,
              endColumn: 12,
            },
          ],
        },
        {
          name: 'static method',
          code: 'Promise.resolve()',
          errors: [
            {
              messageId: 'name',
              message: '"Promise" is not defined.',
              line: 1,
              column: 1,
              endLine: 1,
              endColumn: 8,
            },
          ],
        },
        {
          name: 'browser environment',
          code: 'new Promise(function(reject, resolve) { })',
          languageOptions: { globals: { Promise: false, window: false } },
          errors: [
            {
              messageId: 'name',
              message: '"Promise" is not defined.',
              line: 1,
              column: 5,
              endLine: 1,
              endColumn: 12,
            },
          ],
        },
        {
          name: 'node environment',
          code: 'new Promise(function(reject, resolve) { })',
          languageOptions: { globals: { Promise: false, global: false } },
          errors: [
            {
              messageId: 'name',
              message: '"Promise" is not defined.',
              line: 1,
              column: 5,
              endLine: 1,
              endColumn: 12,
            },
          ],
        },
        {
          name: 'es6 environment',
          code: 'Promise.resolve()',
          errors: [
            {
              messageId: 'name',
              message: '"Promise" is not defined.',
              line: 1,
              column: 1,
              endLine: 1,
              endColumn: 8,
            },
          ],
        },
        {
          name: 'writable global',
          code: 'Promise.resolve()',
          languageOptions: { globals: { Promise: true } },
          errors: [
            {
              messageId: 'name',
              message: '"Promise" is not defined.',
              line: 1,
              column: 1,
              endLine: 1,
              endColumn: 8,
            },
          ],
        },
        {
          name: 'disabled global',
          code: 'Promise.resolve()',
          languageOptions: { globals: { Promise: 'off' } },
          errors: [
            {
              messageId: 'name',
              message: '"Promise" is not defined.',
              line: 1,
              column: 1,
              endLine: 1,
              endColumn: 8,
            },
          ],
        },
        {
          name: 'documentation invalid',
          code: "const x = Promise.resolve('bad')",
          errors: [
            {
              messageId: 'name',
              message: '"Promise" is not defined.',
              line: 1,
              column: 11,
              endLine: 1,
              endColumn: 18,
            },
          ],
        },
      ],
    });
  });
}

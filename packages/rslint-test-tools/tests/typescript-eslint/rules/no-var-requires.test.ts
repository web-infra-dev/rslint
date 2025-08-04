import { describe, test, expect } from '@rstest/core';
import { noFormat, RuleTester } from '@typescript-eslint/rule-tester';
import { getFixturesRootDir } from '../RuleTester.ts';

const rootPath = getFixturesRootDir();

const ruleTester = new RuleTester({
  languageOptions: {
    parserOptions: {
      project: './tsconfig.json',
      tsconfigRootDir: rootPath,
    },
  },
});

describe('no-var-requires', () => {
  test('rule tests', () => {
    ruleTester.run('no-var-requires', {
      valid: [
        "import foo = require('foo');",
        "import foo from 'foo';",
        "import * as foo from 'foo';",
        "import { bar } from 'foo';",
        "require('foo');",
        "require?.('foo');",
        {
          code: "const foo = require('foo');",
          options: [{ allow: ['/foo/'] }],
        },
        {
          code: "const foo = require('./foo');",
          options: [{ allow: ['/foo/'] }],
        },
        {
          code: "const foo = require('../foo');",
          options: [{ allow: ['/foo/'] }],
        },
        {
          code: "const foo = require('foo/bar');",
          options: [{ allow: ['/bar/'] }],
        },
      ],
      invalid: [
        {
          code: "var foo = require('foo');",
          errors: [
            {
              messageId: 'noVarReqs',
              line: 1,
              column: 11,
              endLine: 1,
              endColumn: 26,
            },
          ],
        },
        {
          code: "const foo = require('foo');",
          errors: [
            {
              messageId: 'noVarReqs',
              line: 1,
              column: 13,
              endLine: 1,
              endColumn: 28,
            },
          ],
        },
        {
          code: "let foo = require('foo');",
          errors: [
            {
              messageId: 'noVarReqs',
              line: 1,
              column: 11,
              endLine: 1,
              endColumn: 26,
            },
          ],
        },
        {
          code: "var foo = require('foo'), bar = require('bar');",
          errors: [
            {
              messageId: 'noVarReqs',
              line: 1,
              column: 11,
              endLine: 1,
              endColumn: 26,
            },
            {
              messageId: 'noVarReqs',
              line: 1,
              column: 33,
              endLine: 1,
              endColumn: 48,
            },
          ],
        },
        {
          code: "const { foo } = require('foo');",
          errors: [
            {
              messageId: 'noVarReqs',
              line: 1,
              column: 17,
              endLine: 1,
              endColumn: 32,
            },
          ],
        },
        {
          code: "const { foo, bar } = require('foo');",
          errors: [
            {
              messageId: 'noVarReqs',
              line: 1,
              column: 22,
              endLine: 1,
              endColumn: 37,
            },
          ],
        },
        {
          code: "const foo = require?.('foo');",
          errors: [
            {
              messageId: 'noVarReqs',
              line: 1,
              column: 13,
              endLine: 1,
              endColumn: 30,
            },
          ],
        },
        {
          code: "const foo = require('foo'), bar = require('bar');",
          options: [{ allow: ['/bar/'] }],
          errors: [
            {
              messageId: 'noVarReqs',
              line: 1,
              column: 13,
              endLine: 1,
              endColumn: 28,
            },
          ],
        },
        {
          code: "const foo = require('./foo');",
          options: [{ allow: ['/bar/'] }],
          errors: [
            {
              messageId: 'noVarReqs',
              line: 1,
              column: 13,
              endLine: 1,
              endColumn: 30,
            },
          ],
        },
        {
          code: "const foo = require('foo') as Foo;",
          errors: [
            {
              messageId: 'noVarReqs',
              line: 1,
              column: 13,
              endLine: 1,
              endColumn: 28,
            },
          ],
        },
        {
          code: "const foo: Foo = require('foo');",
          errors: [
            {
              messageId: 'noVarReqs',
              line: 1,
              column: 18,
              endLine: 1,
              endColumn: 33,
            },
          ],
        },
      ],
    });
  });
});

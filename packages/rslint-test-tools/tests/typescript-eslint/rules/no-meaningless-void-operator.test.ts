import { describe, test, expect } from '@rstest/core';
import { RuleTester } from '@typescript-eslint/rule-tester';

import { getFixturesRootDir } from '../RuleTester';

const rootDir = getFixturesRootDir();
const ruleTester = new RuleTester({
  languageOptions: {
    parserOptions: {
      project: './tsconfig.json',
      tsconfigRootDir: rootDir,
    },
  },
});

describe('no-meaningless-void-operator', () => {
  test('rule tests', () => {
    ruleTester.run('no-meaningless-void-operator', {
      valid: [
        `
(() => {})();

function foo() {}
foo(); // nothing to discard

function bar(x: number) {
  void x;
  return 2;
}
void bar(); // discarding a number
    `,
        `
function bar(x: never) {
  void x;
}
    `,
      ],
      invalid: [
        {
          code: 'void (() => {})();',
          errors: [
            {
              column: 1,
              line: 1,
              messageId: 'meaninglessVoidOperator',
            },
          ],
          output: '(() => {})();',
        },
        {
          code: `
function foo() {}
void foo();
      `,
          errors: [
            {
              column: 1,
              line: 3,
              messageId: 'meaninglessVoidOperator',
            },
          ],
          output: `
function foo() {}
foo();
      `,
        },
        {
          code: `
function bar(x: never) {
  void x;
}
      `,
          errors: [
            {
              column: 3,
              line: 3,
              messageId: 'meaninglessVoidOperator',
              suggestions: [
                {
                  messageId: 'removeVoid',
                  output: `
function bar(x: never) {
  x;
}
      `,
                },
              ],
            },
          ],
          options: [{ checkNever: true }],
          output: null,
        },
      ],
    });
  });
});

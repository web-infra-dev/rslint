import { describe, test, expect } from '@rstest/core';
import { noFormat, RuleTester } from '@typescript-eslint/rule-tester';
import { AST_NODE_TYPES } from '@typescript-eslint/utils';
import { getFixturesRootDir } from '../RuleTester.ts';

describe('explicit-function-return-type', () => {
  test('rule tests', () => {
    const ruleTester = new RuleTester();

    ruleTester.run('explicit-function-return-type', {
      valid: [
        'function test(): void { return; }',
        'const fn = function(): number { return 1; }',
        'const fn = (): string => "test";',
        'class Test { method(): boolean { return true; } }',
        'const obj = { method(): number { return 42; } };',
        {
          code: 'function test() { return; }',
          options: [{ allowExpressions: true }],
        },
        {
          code: 'const fn = () => "test";',
          options: [{ allowExpressions: true }],
        },
        {
          code: 'const fn = function() { return 1; }',
          options: [{ allowExpressions: true }],
        },
        {
          code: '(() => {})();',
          options: [{ allowExpressions: true }],
        },
        {
          code: 'export default () => {};',
          options: [{ allowExpressions: true }],
        },
        {
          code: 'const foo = { bar() {} };',
          options: [{ allowTypedFunctionExpressions: true }],
        },
        {
          code: 'const foo: Foo = () => {};',
          options: [{ allowTypedFunctionExpressions: true }],
        },
        {
          code: 'const foo = <Foo>(() => {});',
          options: [{ allowTypedFunctionExpressions: true }],
        },
        {
          code: 'const foo = (() => {}) as Foo;',
          options: [{ allowTypedFunctionExpressions: true }],
        },
        {
          code: 'function* test() { yield 1; }',
          options: [{ allowGenerators: true }],
        },
        {
          code: 'const fn = function*() { yield 1; }',
          options: [{ allowGenerators: true }],
        },
        {
          code: 'const higherOrderFn = () => () => 1;',
          options: [{ allowHigherOrderFunctions: true }],
        },
        {
          code: 'const higherOrderFn = () => function() { return 1; };',
          options: [{ allowHigherOrderFunctions: true }],
        },
        {
          code: 'const obj = { set foo(value) {} };',
          options: [{ allowIIFE: false }],
        },
        {
          code: 'class Test { set foo(value) {} }',
          options: [{ allowIIFE: false }],
        },
        {
          code: 'const x = (() => 1)();',
          options: [{ allowIIFE: true }],
        },
        {
          code: '(function() { return 1; })();',
          options: [{ allowIIFE: true }],
        },
      ],
      invalid: [
        {
          code: 'function test() { return; }',
          errors: [
            {
              messageId: 'missingReturnType',
              line: 1,
              column: 9,
              endLine: 1,
              endColumn: 13,
            },
          ],
        },
        {
          code: 'const fn = function() { return 1; }',
          errors: [
            {
              messageId: 'missingReturnType',
              line: 1,
              column: 12,
              endLine: 1,
              endColumn: 20,
            },
          ],
        },
        {
          code: 'const fn = () => "test";',
          errors: [
            {
              messageId: 'missingReturnType',
              line: 1,
              column: 12,
              endLine: 1,
              endColumn: 17,
            },
          ],
        },
        {
          code: 'class Test { method() { return true; } }',
          errors: [
            {
              messageId: 'missingReturnType',
              line: 1,
              column: 14,
              endLine: 1,
              endColumn: 20,
            },
          ],
        },
        {
          code: 'const obj = { method() { return 42; } };',
          errors: [
            {
              messageId: 'missingReturnType',
              line: 1,
              column: 15,
              endLine: 1,
              endColumn: 21,
            },
          ],
        },
        {
          code: 'export default function() {}',
          errors: [
            {
              messageId: 'missingReturnType',
              line: 1,
              column: 16,
              endLine: 1,
              endColumn: 24,
            },
          ],
        },
        {
          code: 'export default () => {};',
          errors: [
            {
              messageId: 'missingReturnType',
              line: 1,
              column: 16,
              endLine: 1,
              endColumn: 21,
            },
          ],
        },
        {
          code: 'const fn = function* () { yield 1; }',
          errors: [
            {
              messageId: 'missingReturnType',
              line: 1,
              column: 12,
              endLine: 1,
              endColumn: 21,
            },
          ],
        },
        {
          code: 'function foo() { return () => 1; }',
          errors: [
            {
              messageId: 'missingReturnType',
              line: 1,
              column: 9,
              endLine: 1,
              endColumn: 12,
            },
          ],
        },
        {
          code: 'const foo = () => () => 1;',
          errors: [
            {
              messageId: 'missingReturnType',
              line: 1,
              column: 13,
              endLine: 1,
              endColumn: 18,
            },
          ],
        },
        {
          code: 'const x = (() => 1)();',
          options: [{ allowIIFE: false }],
          errors: [
            {
              messageId: 'missingReturnType',
              line: 1,
              column: 12,
              endLine: 1,
              endColumn: 17,
            },
          ],
        },
      ],
    });
  });
});

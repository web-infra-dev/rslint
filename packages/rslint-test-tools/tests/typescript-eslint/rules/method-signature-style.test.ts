import { describe, test, expect } from '@rstest/core';
import { noFormat, RuleTester } from '@typescript-eslint/rule-tester';
import { AST_NODE_TYPES } from '@typescript-eslint/utils';
import { getFixturesRootDir } from '../RuleTester.ts';

describe('method-signature-style', () => {
  test('rule tests', () => {
    const ruleTester = new RuleTester();

    ruleTester.run('method-signature-style', {
      valid: [
        'interface Foo { bar(): void; }',
        'interface Foo { bar: () => void; }',
        {
          code: 'interface Foo { bar(): void; }',
          options: ['method'],
        },
        {
          code: 'interface Foo { bar: () => void; }',
          options: ['property'],
        },
        'type Foo = { bar(): void; };',
        'type Foo = { bar: () => void; };',
        {
          code: 'type Foo = { bar(): void; };',
          options: ['method'],
        },
        {
          code: 'type Foo = { bar: () => void; };',
          options: ['property'],
        },
      ],
      invalid: [
        {
          code: 'interface Foo { bar: () => void; }',
          options: ['method'],
          errors: [
            {
              messageId: 'errorMethod',
              line: 1,
              column: 17,
              endLine: 1,
              endColumn: 20,
            },
          ],
          output: 'interface Foo { bar(): void; }',
        },
        {
          code: 'interface Foo { bar(): void; }',
          options: ['property'],
          errors: [
            {
              messageId: 'errorProperty',
              line: 1,
              column: 17,
              endLine: 1,
              endColumn: 20,
            },
          ],
          output: 'interface Foo { bar: () => void; }',
        },
        {
          code: 'interface Foo { bar(x: number): string; }',
          options: ['property'],
          errors: [
            {
              messageId: 'errorProperty',
              line: 1,
              column: 17,
              endLine: 1,
              endColumn: 20,
            },
          ],
          output: 'interface Foo { bar: (x: number) => string; }',
        },
        {
          code: 'interface Foo { bar: (x: number) => string; }',
          options: ['method'],
          errors: [
            {
              messageId: 'errorMethod',
              line: 1,
              column: 17,
              endLine: 1,
              endColumn: 20,
            },
          ],
          output: 'interface Foo { bar(x: number): string; }',
        },
        {
          code: 'type Foo = { bar: () => void; };',
          options: ['method'],
          errors: [
            {
              messageId: 'errorMethod',
              line: 1,
              column: 14,
              endLine: 1,
              endColumn: 17,
            },
          ],
          output: 'type Foo = { bar(): void; };',
        },
        {
          code: 'type Foo = { bar(): void; };',
          options: ['property'],
          errors: [
            {
              messageId: 'errorProperty',
              line: 1,
              column: 14,
              endLine: 1,
              endColumn: 17,
            },
          ],
          output: 'type Foo = { bar: () => void; };',
        },
        {
          code: 'interface Foo { bar<T>(): T; }',
          options: ['property'],
          errors: [
            {
              messageId: 'errorProperty',
              line: 1,
              column: 17,
              endLine: 1,
              endColumn: 20,
            },
          ],
          output: 'interface Foo { bar: <T>() => T; }',
        },
        {
          code: 'interface Foo { bar: <T>() => T; }',
          options: ['method'],
          errors: [
            {
              messageId: 'errorMethod',
              line: 1,
              column: 17,
              endLine: 1,
              endColumn: 20,
            },
          ],
          output: 'interface Foo { bar<T>(): T; }',
        },
      ],
    });
  });
});

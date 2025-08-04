import { describe, test, expect } from '@rstest/core';
import { noFormat, RuleTester } from '@typescript-eslint/rule-tester';
import { getFixturesRootDir } from '../RuleTester.ts';

const rootPath = getFixturesRootDir();
// this rule tests the spacing, which prettier will want to fix and break the tests
/* eslint "@typescript-eslint/internal/plugin-test-formatting": ["error", { formatWithPrettier: false }] */

const ruleTester = new RuleTester({
  languageOptions: {
    parserOptions: {
      project: './tsconfig.json',
      tsconfigRootDir: rootPath,
    },
  },
});

const error = {
  messageId: 'uselessExport',
} as const;

describe('no-useless-empty-export', () => {
  test('should work', () => {
    ruleTester.run('no-useless-empty-export', {
      valid: [
        "declare module '_'",
        "import {} from '_';",
        "import * as _ from '_';",
        'export = {};',
        'export = 3;',
        'export const _ = {};',
        `
      const _ = {};
      export default _;
    `,
        `
      export * from '_';
      export = {};
    `,
        `
      export {};
    `,
        // https://github.com/microsoft/TypeScript/issues/38592
        {
          code: `
        export type A = 1;
        export {};
      `,
          // @ts-ignore
          filename: 'foo.d.ts',
        },
        {
          code: `
        export declare const a = 2;
        export {};
      `,
          // @ts-ignore
          filename: 'foo.d.ts',
        },
        {
          code: `
        import type { A } from '_';
        export {};
      `,
          // @ts-ignore
          filename: 'foo.d.ts',
        },
        {
          code: `
        import { A } from '_';
        export {};
      `,
          // @ts-ignore
          filename: 'foo.d.ts',
        },
      ],
      invalid: [
        {
          code: `
export const _ = {};
export {};
      `,
          errors: [error],
          output: `
export const _ = {};

      `,
        },
        {
          code: `
export * from '_';
export {};
      `,
          errors: [error],
          output: `
export * from '_';

      `,
        },
        {
          code: `
export {};
export * from '_';
      `,
          errors: [error],
          output: `

export * from '_';
      `,
        },
        {
          code: `
const _ = {};
export default _;
export {};
      `,
          errors: [error],
          output: `
const _ = {};
export default _;

      `,
        },
        {
          code: `
export {};
const _ = {};
export default _;
      `,
          errors: [error],
          output: `

const _ = {};
export default _;
      `,
        },
        {
          code: `
const _ = {};
export { _ };
export {};
      `,
          errors: [error],
          output: `
const _ = {};
export { _ };

      `,
        },
        {
          code: `
import _ = require('_');
export {};
      `,
          errors: [error],
          output: `
import _ = require('_');

      `,
        },
        {
          code: `
import _ = require('_');
export {};
export {};
      `,
          errors: [error, error],
          output: `
import _ = require('_');


      `,
        },
      ],
    });
  });
});

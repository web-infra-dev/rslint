// Every upstream case from typescript-eslint v8.72.1, preserving inputs, options
// and outputs. Errors carry the full messages and ranges produced by that release.
// Parser-level `ecmaVersion`/`sourceType` move to language options. Upstream
// runs these cases without type information; here each uses a fixture project
// with the upstream compiler options (without ambient `types`).
// https://github.com/typescript-eslint/typescript-eslint/blob/v8.72.1/packages/eslint-plugin/tests/rules/no-unused-vars/no-unused-vars-enableAutofixRemoval.test.ts
import { RuleTester } from '@rslint/rule-tester';

const noFormat = (strings: TemplateStringsArray, ...values: unknown[]) =>
  String.raw({ raw: strings }, ...values);

const ruleTester = new RuleTester({
  plugin: '@typescript-eslint',
  filename: 'file.ts',
  fixtureFiles: {
    'tsconfig.json': "{\"compilerOptions\":{\"jsx\":\"preserve\",\"target\":\"es2015\",\"module\":\"commonjs\",\"strict\":true,\"types\":[],\"lib\":[\"es2015\",\"es2017\",\"esnext\"],\"experimentalDecorators\":true,\"esModuleInterop\":true},\"include\":[\"file.ts\",\"react.tsx\",\"foo.d.ts\"]}",
  },
  languageOptions: { ecmaVersion: 6, sourceType: 'module', parserOptions: { project: './tsconfig.json' } },
});

ruleTester.run('no-unused-vars', {
  valid: [],
  invalid: [
      {
        code: `
import * as Unused from 'module';
export {};
        `,
        errors: [{ messageId: "unusedVar", message: "'Unused' is defined but never used.", line: 2, column: 13, endLine: 2, endColumn: 19 }],
        options: [
          {
            enableAutofixRemoval: {
              imports: true,
            },
          },
        ],
        output: `
export {};
        `,
      },
      {
        code: `
import Unused from 'module';
export {};
        `,
        errors: [{ messageId: "unusedVar", message: "'Unused' is defined but never used.", line: 2, column: 8, endLine: 2, endColumn: 14 }],
        options: [
          {
            enableAutofixRemoval: {
              imports: true,
            },
          },
        ],
        output: `
export {};
        `,
      },
      {
        code: `
import { Unused } from 'module';
export {};
        `,
        errors: [{ messageId: "unusedVar", message: "'Unused' is defined but never used.", line: 2, column: 10, endLine: 2, endColumn: 16 }],
        options: [
          {
            enableAutofixRemoval: {
              imports: true,
            },
          },
        ],
        output: `
export {};
        `,
      },
      {
        code: `
import { Unused, Unused2 } from 'module';
export {};
        `,
        errors: [{ messageId: "unusedVar", message: "'Unused' is defined but never used.", line: 2, column: 10, endLine: 2, endColumn: 16 }, { messageId: "unusedVar", message: "'Unused2' is defined but never used.", line: 2, column: 18, endLine: 2, endColumn: 25 }],
        options: [
          {
            enableAutofixRemoval: {
              imports: true,
            },
          },
        ],
        output: `
export {};
        `,
      },
      {
        code: `
import { Unused, Used } from 'module';
export { Used };
        `,
        errors: [{ messageId: "unusedVar", message: "'Unused' is defined but never used.", line: 2, column: 10, endLine: 2, endColumn: 16 }],
        options: [
          {
            enableAutofixRemoval: {
              imports: true,
            },
          },
        ],
        output: `
import {  Used } from 'module';
export { Used };
        `,
      },
      {
        code: `
import { Used, Unused } from 'module';
export { Used };
        `,
        errors: [{ messageId: "unusedVar", message: "'Unused' is defined but never used.", line: 2, column: 16, endLine: 2, endColumn: 22 }],
        options: [
          {
            enableAutofixRemoval: {
              imports: true,
            },
          },
        ],
        output: `
import { Used } from 'module';
export { Used };
        `,
      },
      {
        code: noFormat`
import { Used, Unused, } from 'module';
export { Used };
        `,
        errors: [{ messageId: "unusedVar", message: "'Unused' is defined but never used.", line: 2, column: 16, endLine: 2, endColumn: 22 }],
        options: [
          {
            enableAutofixRemoval: {
              imports: true,
            },
          },
        ],
        output: `
import { Used, } from 'module';
export { Used };
        `,
      },
      {
        code: `
import { Used, Unused, Used2 } from 'module';
export { Used, Used2 };
        `,
        errors: [{ messageId: "unusedVar", message: "'Unused' is defined but never used.", line: 2, column: 16, endLine: 2, endColumn: 22 }],
        options: [
          {
            enableAutofixRemoval: {
              imports: true,
            },
          },
        ],
        output: `
import { Used, Used2 } from 'module';
export { Used, Used2 };
        `,
      },
      {
        code: `
import Unused, { Unused2 } from 'module';
export {};
        `,
        errors: [{ messageId: "unusedVar", message: "'Unused' is defined but never used.", line: 2, column: 8, endLine: 2, endColumn: 14 }, { messageId: "unusedVar", message: "'Unused2' is defined but never used.", line: 2, column: 18, endLine: 2, endColumn: 25 }],
        options: [
          {
            enableAutofixRemoval: {
              imports: true,
            },
          },
        ],
        output: `
export {};
        `,
      },
      {
        code: `
import Unused, { Used } from 'module';
export { Used };
        `,
        errors: [{ messageId: "unusedVar", message: "'Unused' is defined but never used.", line: 2, column: 8, endLine: 2, endColumn: 14 }],
        options: [
          {
            enableAutofixRemoval: {
              imports: true,
            },
          },
        ],
        output: `
import  { Used } from 'module';
export { Used };
        `,
      },
      {
        code: `
import Used, { Unused } from 'module';
export { Used };
        `,
        errors: [{ messageId: "unusedVar", message: "'Unused' is defined but never used.", line: 2, column: 16, endLine: 2, endColumn: 22 }],
        options: [
          {
            enableAutofixRemoval: {
              imports: true,
            },
          },
        ],
        output: `
import Used from 'module';
export { Used };
        `,
      },
      {
        code: noFormat`
import Used, { Unused, } from 'module';
export { Used };
        `,
        errors: [{ messageId: "unusedVar", message: "'Unused' is defined but never used.", line: 2, column: 16, endLine: 2, endColumn: 22 }],
        options: [
          {
            enableAutofixRemoval: {
              imports: true,
            },
          },
        ],
        output: `
import Used from 'module';
export { Used };
        `,
      },
      {
        code: `
import Used, { Used2, Unused } from 'module';
export { Used, Used2 };
        `,
        errors: [{ messageId: "unusedVar", message: "'Unused' is defined but never used.", line: 2, column: 23, endLine: 2, endColumn: 29 }],
        options: [
          {
            enableAutofixRemoval: {
              imports: true,
            },
          },
        ],
        output: `
import Used, { Used2 } from 'module';
export { Used, Used2 };
        `,
      },
      {
        code: `
import Used, { Unused, Used2 } from 'module';
export { Used, Used2 };
        `,
        errors: [{ messageId: "unusedVar", message: "'Unused' is defined but never used.", line: 2, column: 16, endLine: 2, endColumn: 22 }],
        options: [
          {
            enableAutofixRemoval: {
              imports: true,
            },
          },
        ],
        output: `
import Used, {  Used2 } from 'module';
export { Used, Used2 };
        `,
      },
      {
        code: `
import Unused, { Unused2, Used } from 'module';
export { Used };
        `,
        errors: [{ messageId: "unusedVar", message: "'Unused' is defined but never used.", line: 2, column: 8, endLine: 2, endColumn: 14 }, { messageId: "unusedVar", message: "'Unused2' is defined but never used.", line: 2, column: 18, endLine: 2, endColumn: 25 }],
        options: [
          {
            enableAutofixRemoval: {
              imports: true,
            },
          },
        ],
        output: `
import  {  Used } from 'module';
export { Used };
        `,
      },
      {
        code: `
import Unused, { Used, Unused2 } from 'module';
export { Used };
        `,
        errors: [{ messageId: "unusedVar", message: "'Unused' is defined but never used.", line: 2, column: 8, endLine: 2, endColumn: 14 }, { messageId: "unusedVar", message: "'Unused2' is defined but never used.", line: 2, column: 24, endLine: 2, endColumn: 31 }],
        options: [
          {
            enableAutofixRemoval: {
              imports: true,
            },
          },
        ],
        output: `
import  { Used } from 'module';
export { Used };
        `,
      },
      {
        code: `
import { Unused as Unused1, Used } from 'module';
export { Used };
        `,
        errors: [{ messageId: "unusedVar", message: "'Unused1' is defined but never used.", line: 2, column: 20, endLine: 2, endColumn: 27 }],
        options: [
          {
            enableAutofixRemoval: {
              imports: true,
            },
          },
        ],
        output: `
import {  Used } from 'module';
export { Used };
        `,
      },
      {
        code: `
import { Used, Unused as Unused1 } from 'module';
export { Used };
        `,
        errors: [{ messageId: "unusedVar", message: "'Unused1' is defined but never used.", line: 2, column: 26, endLine: 2, endColumn: 33 }],
        options: [
          {
            enableAutofixRemoval: {
              imports: true,
            },
          },
        ],
        output: `
import { Used } from 'module';
export { Used };
        `,
      },
      {
        code: noFormat`
import { Used, Unused as Unused1, } from 'module';
export { Used };
        `,
        errors: [{ messageId: "unusedVar", message: "'Unused1' is defined but never used.", line: 2, column: 26, endLine: 2, endColumn: 33 }],
        options: [
          {
            enableAutofixRemoval: {
              imports: true,
            },
          },
        ],
        output: `
import { Used, } from 'module';
export { Used };
        `,
      },
      {
        code: `
/* this is an important comment */ import assert from 'assert';
        `,
        errors: [{ messageId: "unusedVar", message: "'assert' is defined but never used.", line: 2, column: 43, endLine: 2, endColumn: 49 }],
        options: [
          {
            enableAutofixRemoval: {
              imports: true,
            },
          },
        ],
        output: `
/* this is an important comment */\u0020
        `,
      },
      {
        code: `
import assert from 'assert'; /* this is an important comment */
        `,
        errors: [{ messageId: "unusedVar", message: "'assert' is defined but never used.", line: 2, column: 8, endLine: 2, endColumn: 14 }],
        options: [
          {
            enableAutofixRemoval: {
              imports: true,
            },
          },
        ],
        output: `
 /* this is an important comment */
        `,
      },
    ],
});

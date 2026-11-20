// Upstream: typescript-eslint v8.67.0; full diagnostics verified against the upstream rule.
/* eslint-disable @typescript-eslint/internal/no-multiple-lines-of-errors */
/* eslint-disable @typescript-eslint/internal/plugin-test-formatting -- Prettier doesn't yet support TS 5.6 string literal module identifiers */
/* eslint-disable @typescript-eslint/internal/prefer-ast-types-enum */
import { RuleTester } from '@rslint/rule-tester';

const ruleTester = new RuleTester({
  plugin: '@typescript-eslint',
  filename: 'file.ts',
  fixtureFiles: {
    'tsconfig.json': JSON.stringify({
      compilerOptions: {
        strict: true,
        target: 'es2015',
        module: 'commonjs',
        lib: ['es2015', 'es2017', 'esnext', 'dom'],
        types: [],
        experimentalDecorators: true,
      },
      include: ['file.ts'],
    }),
  },
});

// only need parserOptions for the `type` option tests
const parserOptions = {
  project: './tsconfig.json',
  projectService: false,
};

ruleTester.run('naming-convention', {
  invalid: [
    {
      // make sure we handle no options and apply defaults
      code: 'const x_x = 1;',
      errors: [
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Variable name `x_x` must match one of the following formats: camelCase, UPPER_CASE',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      // make sure we handle empty options and apply defaults
      code: 'const x_x = 1;',
      errors: [
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Variable name `x_x` must match one of the following formats: camelCase, UPPER_CASE',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 10,
        },
      ],
      options: [],
    },
    {
      code: `
        const child_process = require('child_process');
      `,
      errors: [
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Variable name `child_process` must match one of the following formats: camelCase',
          line: 2,
          column: 15,
          endLine: 2,
          endColumn: 28,
        },
      ],
      languageOptions: { parserOptions },
      options: [
        {
          filter: {
            match: true,
            regex: 'child_process',
          },
          format: ['camelCase'],
          selector: 'default',
        },
      ],
    },
    {
      code: `
        declare const any_camelCase01: any;
        declare const any_camelCase02: any | null;
        declare const any_camelCase03: any | null | undefined;
        declare const string_camelCase01: string;
        declare const string_camelCase02: string | null;
        declare const string_camelCase03: string | null | undefined;
        declare const string_camelCase04: 'a' | null | undefined;
        declare const string_camelCase05: string | 'a' | null | undefined;
        declare const number_camelCase06: number;
        declare const number_camelCase07: number | null;
        declare const number_camelCase08: number | null | undefined;
        declare const number_camelCase09: 1 | null | undefined;
        declare const number_camelCase10: number | 2 | null | undefined;
        declare const boolean_camelCase11: boolean;
        declare const boolean_camelCase12: boolean | null;
        declare const boolean_camelCase13: boolean | null | undefined;
        declare const boolean_camelCase14: true | null | undefined;
        declare const boolean_camelCase15: false | null | undefined;
        declare const boolean_camelCase16: true | false | null | undefined;
      `,
      errors: [
        {
          messageId: 'doesNotMatchFormatTrimmed',
          message:
            'Variable name `any_camelCase01` trimmed as `camelCase01` must match one of the following formats: UPPER_CASE',
          line: 2,
          column: 23,
          endLine: 2,
          endColumn: 43,
        },
        {
          messageId: 'doesNotMatchFormatTrimmed',
          message:
            'Variable name `any_camelCase02` trimmed as `camelCase02` must match one of the following formats: UPPER_CASE',
          line: 3,
          column: 23,
          endLine: 3,
          endColumn: 50,
        },
        {
          messageId: 'doesNotMatchFormatTrimmed',
          message:
            'Variable name `any_camelCase03` trimmed as `camelCase03` must match one of the following formats: UPPER_CASE',
          line: 4,
          column: 23,
          endLine: 4,
          endColumn: 62,
        },
        {
          messageId: 'doesNotMatchFormatTrimmed',
          message:
            'Variable name `string_camelCase01` trimmed as `camelCase01` must match one of the following formats: snake_case',
          line: 5,
          column: 23,
          endLine: 5,
          endColumn: 49,
        },
        {
          messageId: 'doesNotMatchFormatTrimmed',
          message:
            'Variable name `string_camelCase02` trimmed as `camelCase02` must match one of the following formats: snake_case',
          line: 6,
          column: 23,
          endLine: 6,
          endColumn: 56,
        },
        {
          messageId: 'doesNotMatchFormatTrimmed',
          message:
            'Variable name `string_camelCase03` trimmed as `camelCase03` must match one of the following formats: snake_case',
          line: 7,
          column: 23,
          endLine: 7,
          endColumn: 68,
        },
        {
          messageId: 'doesNotMatchFormatTrimmed',
          message:
            'Variable name `string_camelCase04` trimmed as `camelCase04` must match one of the following formats: snake_case',
          line: 8,
          column: 23,
          endLine: 8,
          endColumn: 65,
        },
        {
          messageId: 'doesNotMatchFormatTrimmed',
          message:
            'Variable name `string_camelCase05` trimmed as `camelCase05` must match one of the following formats: snake_case',
          line: 9,
          column: 23,
          endLine: 9,
          endColumn: 74,
        },
        {
          messageId: 'doesNotMatchFormatTrimmed',
          message:
            'Variable name `number_camelCase06` trimmed as `camelCase06` must match one of the following formats: snake_case',
          line: 10,
          column: 23,
          endLine: 10,
          endColumn: 49,
        },
        {
          messageId: 'doesNotMatchFormatTrimmed',
          message:
            'Variable name `number_camelCase07` trimmed as `camelCase07` must match one of the following formats: snake_case',
          line: 11,
          column: 23,
          endLine: 11,
          endColumn: 56,
        },
        {
          messageId: 'doesNotMatchFormatTrimmed',
          message:
            'Variable name `number_camelCase08` trimmed as `camelCase08` must match one of the following formats: snake_case',
          line: 12,
          column: 23,
          endLine: 12,
          endColumn: 68,
        },
        {
          messageId: 'doesNotMatchFormatTrimmed',
          message:
            'Variable name `number_camelCase09` trimmed as `camelCase09` must match one of the following formats: snake_case',
          line: 13,
          column: 23,
          endLine: 13,
          endColumn: 63,
        },
        {
          messageId: 'doesNotMatchFormatTrimmed',
          message:
            'Variable name `number_camelCase10` trimmed as `camelCase10` must match one of the following formats: snake_case',
          line: 14,
          column: 23,
          endLine: 14,
          endColumn: 72,
        },
        {
          messageId: 'doesNotMatchFormatTrimmed',
          message:
            'Variable name `boolean_camelCase11` trimmed as `camelCase11` must match one of the following formats: snake_case',
          line: 15,
          column: 23,
          endLine: 15,
          endColumn: 51,
        },
        {
          messageId: 'doesNotMatchFormatTrimmed',
          message:
            'Variable name `boolean_camelCase12` trimmed as `camelCase12` must match one of the following formats: snake_case',
          line: 16,
          column: 23,
          endLine: 16,
          endColumn: 58,
        },
        {
          messageId: 'doesNotMatchFormatTrimmed',
          message:
            'Variable name `boolean_camelCase13` trimmed as `camelCase13` must match one of the following formats: snake_case',
          line: 17,
          column: 23,
          endLine: 17,
          endColumn: 70,
        },
        {
          messageId: 'doesNotMatchFormatTrimmed',
          message:
            'Variable name `boolean_camelCase14` trimmed as `camelCase14` must match one of the following formats: snake_case',
          line: 18,
          column: 23,
          endLine: 18,
          endColumn: 67,
        },
        {
          messageId: 'doesNotMatchFormatTrimmed',
          message:
            'Variable name `boolean_camelCase15` trimmed as `camelCase15` must match one of the following formats: snake_case',
          line: 19,
          column: 23,
          endLine: 19,
          endColumn: 68,
        },
        {
          messageId: 'doesNotMatchFormatTrimmed',
          message:
            'Variable name `boolean_camelCase16` trimmed as `camelCase16` must match one of the following formats: snake_case',
          line: 20,
          column: 23,
          endLine: 20,
          endColumn: 75,
        },
      ],
      languageOptions: { parserOptions },
      options: [
        {
          format: ['UPPER_CASE'],
          modifiers: ['const'],
          prefix: ['any_'],
          selector: 'variable',
        },
        {
          format: ['snake_case'],
          prefix: ['string_'],
          selector: 'variable',
          types: ['string'],
        },
        {
          format: ['snake_case'],
          prefix: ['number_'],
          selector: 'variable',
          types: ['number'],
        },
        {
          format: ['snake_case'],
          prefix: ['boolean_'],
          selector: 'variable',
          types: ['boolean'],
        },
      ],
    },
    {
      code: `
        declare const function_camelCase1: () => void;
        declare const function_camelCase2: (() => void) | null;
        declare const function_camelCase3: (() => void) | null | undefined;
        declare const function_camelCase4:
          | (() => void)
          | (() => string)
          | null
          | undefined;
      `,
      errors: [
        {
          messageId: 'doesNotMatchFormatTrimmed',
          message:
            'Variable name `function_camelCase1` trimmed as `camelCase1` must match one of the following formats: snake_case',
          line: 2,
          column: 23,
          endLine: 2,
          endColumn: 54,
        },
        {
          messageId: 'doesNotMatchFormatTrimmed',
          message:
            'Variable name `function_camelCase2` trimmed as `camelCase2` must match one of the following formats: snake_case',
          line: 3,
          column: 23,
          endLine: 3,
          endColumn: 63,
        },
        {
          messageId: 'doesNotMatchFormatTrimmed',
          message:
            'Variable name `function_camelCase3` trimmed as `camelCase3` must match one of the following formats: snake_case',
          line: 4,
          column: 23,
          endLine: 4,
          endColumn: 75,
        },
        {
          messageId: 'doesNotMatchFormatTrimmed',
          message:
            'Variable name `function_camelCase4` trimmed as `camelCase4` must match one of the following formats: snake_case',
          line: 5,
          column: 23,
          endLine: 9,
          endColumn: 22,
        },
      ],
      languageOptions: { parserOptions },
      options: [
        {
          format: ['snake_case'],
          prefix: ['function_'],
          selector: 'variable',
          types: ['function'],
        },
      ],
    },
    {
      code: `
        declare const array_camelCase1: Array<number>;
        declare const array_camelCase2: ReadonlyArray<number> | null;
        declare const array_camelCase3: number[] | null | undefined;
        declare const array_camelCase4: readonly number[] | null | undefined;
        declare const array_camelCase5:
          | number[]
          | (number | string)[]
          | null
          | undefined;
        declare const array_camelCase6: [] | null | undefined;
        declare const array_camelCase7: [number] | null | undefined;
        declare const array_camelCase8:
          | readonly number[]
          | Array<string>
          | [boolean]
          | null
          | undefined;
      `,
      errors: [
        {
          messageId: 'doesNotMatchFormatTrimmed',
          message:
            'Variable name `array_camelCase1` trimmed as `camelCase1` must match one of the following formats: snake_case',
          line: 2,
          column: 23,
          endLine: 2,
          endColumn: 54,
        },
        {
          messageId: 'doesNotMatchFormatTrimmed',
          message:
            'Variable name `array_camelCase2` trimmed as `camelCase2` must match one of the following formats: snake_case',
          line: 3,
          column: 23,
          endLine: 3,
          endColumn: 69,
        },
        {
          messageId: 'doesNotMatchFormatTrimmed',
          message:
            'Variable name `array_camelCase3` trimmed as `camelCase3` must match one of the following formats: snake_case',
          line: 4,
          column: 23,
          endLine: 4,
          endColumn: 68,
        },
        {
          messageId: 'doesNotMatchFormatTrimmed',
          message:
            'Variable name `array_camelCase4` trimmed as `camelCase4` must match one of the following formats: snake_case',
          line: 5,
          column: 23,
          endLine: 5,
          endColumn: 77,
        },
        {
          messageId: 'doesNotMatchFormatTrimmed',
          message:
            'Variable name `array_camelCase5` trimmed as `camelCase5` must match one of the following formats: snake_case',
          line: 6,
          column: 23,
          endLine: 10,
          endColumn: 22,
        },
        {
          messageId: 'doesNotMatchFormatTrimmed',
          message:
            'Variable name `array_camelCase6` trimmed as `camelCase6` must match one of the following formats: snake_case',
          line: 11,
          column: 23,
          endLine: 11,
          endColumn: 62,
        },
        {
          messageId: 'doesNotMatchFormatTrimmed',
          message:
            'Variable name `array_camelCase7` trimmed as `camelCase7` must match one of the following formats: snake_case',
          line: 12,
          column: 23,
          endLine: 12,
          endColumn: 68,
        },
        {
          messageId: 'doesNotMatchFormatTrimmed',
          message:
            'Variable name `array_camelCase8` trimmed as `camelCase8` must match one of the following formats: snake_case',
          line: 13,
          column: 23,
          endLine: 18,
          endColumn: 22,
        },
      ],
      languageOptions: { parserOptions },
      options: [
        {
          format: ['snake_case'],
          prefix: ['array_'],
          selector: 'variable',
          types: ['array'],
        },
      ],
    },
    {
      code: `
        let unused_foo = 'a';
      `,
      errors: [
        {
          messageId: 'satisfyCustom',
          message:
            'Variable name `unused_foo` must not match the RegExp: /^unused_\\w/u',
          line: 2,
          column: 13,
          endLine: 2,
          endColumn: 23,
        },
      ],
      options: [
        {
          custom: {
            match: false,
            regex: /^unused_\w/.source,
          },
          format: ['snake_case'],
          leadingUnderscore: 'allow',
          selector: 'default',
        },
      ],
    },
    {
      code: `
        const _unused_foo = 1;
      `,
      errors: [
        {
          messageId: 'satisfyCustom',
          message:
            'Variable name `_unused_foo` must not match the RegExp: /^unused_\\w/u',
          line: 2,
          column: 15,
          endLine: 2,
          endColumn: 26,
        },
      ],
      options: [
        {
          custom: {
            match: false,
            regex: /^unused_\w/.source,
          },
          format: ['snake_case'],
          leadingUnderscore: 'allow',
          selector: 'default',
        },
      ],
    },
    {
      code: `
        interface IFoo {}
      `,
      errors: [
        {
          messageId: 'satisfyCustom',
          message:
            'Interface name `IFoo` must not match the RegExp: /^I[A-Z]/u',
          line: 2,
          column: 19,
          endLine: 2,
          endColumn: 23,
        },
      ],
      options: [
        {
          custom: {
            match: false,
            regex: /^I[A-Z]/.source,
          },
          format: ['PascalCase'],
          selector: 'typeLike',
        },
      ],
    },
    {
      code: `
        class IBar {}
      `,
      errors: [
        {
          messageId: 'satisfyCustom',
          message: 'Class name `IBar` must not match the RegExp: /^I[A-Z]/u',
          line: 2,
          column: 15,
          endLine: 2,
          endColumn: 19,
        },
      ],
      options: [
        {
          custom: {
            match: false,
            regex: /^I[A-Z]/.source,
          },
          format: ['PascalCase'],
          selector: 'typeLike',
        },
      ],
    },
    {
      code: `
        function fooBar() {}
      `,
      errors: [
        {
          messageId: 'satisfyCustom',
          message: 'Function name `fooBar` must match the RegExp: /function/u',
          line: 2,
          column: 18,
          endLine: 2,
          endColumn: 24,
        },
      ],
      options: [
        {
          custom: {
            match: true,
            regex: /function/.source,
          },
          format: ['camelCase'],
          leadingUnderscore: 'allow',
          selector: 'function',
        },
      ],
    },
    {
      code: `
        let unused_foo = 'a';
      `,
      errors: [
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Variable name `unused_foo` must match one of the following formats: camelCase',
          line: 2,
          column: 13,
          endLine: 2,
          endColumn: 23,
        },
      ],
      options: [
        {
          format: ['camelCase'],
          leadingUnderscore: 'allow',
          selector: ['variable', 'function'],
        },
      ],
    },
    {
      code: `
        const _unused_foo = 1;
      `,
      errors: [
        {
          messageId: 'doesNotMatchFormatTrimmed',
          message:
            'Variable name `_unused_foo` trimmed as `unused_foo` must match one of the following formats: camelCase',
          line: 2,
          column: 15,
          endLine: 2,
          endColumn: 26,
        },
      ],
      options: [
        {
          format: ['camelCase'],
          leadingUnderscore: 'allow',
          selector: ['variable', 'function'],
        },
      ],
    },
    {
      code: `
        function foo_bar() {}
      `,
      errors: [
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Function name `foo_bar` must match one of the following formats: camelCase',
          line: 2,
          column: 18,
          endLine: 2,
          endColumn: 25,
        },
      ],
      options: [
        {
          format: ['camelCase'],
          leadingUnderscore: 'allow',
          selector: ['variable', 'function'],
        },
      ],
    },
    {
      code: `
        interface IFoo {}
      `,
      errors: [
        {
          messageId: 'satisfyCustom',
          message:
            'Interface name `IFoo` must not match the RegExp: /^I[A-Z]/u',
          line: 2,
          column: 19,
          endLine: 2,
          endColumn: 23,
        },
      ],
      options: [
        {
          custom: {
            match: false,
            regex: /^I[A-Z]/.source,
          },
          format: ['PascalCase'],
          selector: ['class', 'interface'],
        },
      ],
    },
    {
      code: `
        class IBar {}
      `,
      errors: [
        {
          messageId: 'satisfyCustom',
          message: 'Class name `IBar` must not match the RegExp: /^I[A-Z]/u',
          line: 2,
          column: 15,
          endLine: 2,
          endColumn: 19,
        },
      ],
      options: [
        {
          format: ['camelCase'],
          leadingUnderscore: 'allow',
          selector: ['variable', 'function'],
        },
        {
          custom: {
            match: false,
            regex: /^I[A-Z]/.source,
          },
          format: ['PascalCase'],
          selector: ['class', 'interface'],
        },
      ],
    },
    {
      code: `
        const foo = {
          'Property Name': 'asdf',
        };
      `,
      errors: [
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Object Literal Property name `Property Name` must match one of the following formats: strictCamelCase',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 26,
        },
      ],
      options: [
        {
          filter: {
            match: false,
            regex: /-/.source,
          },
          format: ['strictCamelCase'],
          selector: 'default',
        },
      ],
    },
    {
      code: `
        const myfoo_bar = 'abcs';
        function fun(myfoo: string) {}
        class foo {
          Myfoo: string;
        }
      `,
      errors: [
        {
          messageId: 'doesNotMatchFormatTrimmed',
          message:
            'Variable name `myfoo_bar` trimmed as `foo_bar` must match one of the following formats: PascalCase',
          line: 2,
          column: 15,
          endLine: 2,
          endColumn: 24,
        },
        {
          messageId: 'doesNotMatchFormatTrimmed',
          message:
            'Parameter name `myfoo` trimmed as `foo` must match one of the following formats: PascalCase',
          line: 3,
          column: 22,
          endLine: 3,
          endColumn: 35,
        },
        {
          messageId: 'doesNotMatchFormatTrimmed',
          message:
            'Class Property name `Myfoo` trimmed as `foo` must match one of the following formats: PascalCase',
          line: 5,
          column: 11,
          endLine: 5,
          endColumn: 16,
        },
      ],
      languageOptions: { parserOptions },
      options: [
        {
          format: ['PascalCase'],
          prefix: ['my', 'My'],
          selector: ['variable', 'property', 'parameter'],
          types: ['string'],
        },
      ],
    },
    {
      code: `
        class foo {
          private readonly fooBar: boolean;
        }
      `,
      errors: [
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Class Property name `fooBar` must match one of the following formats: PascalCase',
          line: 3,
          column: 28,
          endLine: 3,
          endColumn: 34,
        },
      ],
      options: [
        {
          format: ['PascalCase'],
          modifiers: ['private', 'readonly'],
          selector: ['property', 'accessor'],
        },
      ],
    },
    {
      code: `
        function my_foo_bar() {}
      `,
      errors: [
        {
          messageId: 'doesNotMatchFormatTrimmed',
          message:
            'Function name `my_foo_bar` trimmed as `_foo_bar` must match one of the following formats: PascalCase',
          line: 2,
          column: 18,
          endLine: 2,
          endColumn: 28,
        },
      ],
      languageOptions: { parserOptions },
      options: [
        {
          format: ['PascalCase'],
          prefix: ['my', 'My'],
          selector: ['variable', 'function'],
          types: ['string'],
        },
      ],
    },
    {
      code: `
        class SomeClass {
          static otherConstant = 'hello';
        }

        export const { otherConstant } = SomeClass;
      `,
      errors: [
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Class Property name `otherConstant` must match one of the following formats: PascalCase',
          line: 3,
          column: 18,
          endLine: 3,
          endColumn: 31,
        },
      ],
      languageOptions: { parserOptions },
      options: [
        { format: ['PascalCase'], selector: 'property' },
        { format: ['camelCase'], selector: 'variable' },
      ],
    },
    {
      code: `
        declare class Foo {
          Bar(Baz: string): void;
        }
      `,
      errors: [
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Parameter name `Baz` must match one of the following formats: camelCase',
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 26,
        },
      ],
      languageOptions: { parserOptions },
      options: [{ format: ['camelCase'], selector: 'parameter' }],
    },
    {
      code: `
        export const PascalCaseVar = 1;
        export enum PascalCaseEnum {}
        export class PascalCaseClass {}
        export function PascalCaseFunction() {}
        export interface PascalCaseInterface {}
        export type PascalCaseType = {};
      `,
      errors: [
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Variable name `PascalCaseVar` must match one of the following formats: camelCase',
          line: 2,
          column: 22,
          endLine: 2,
          endColumn: 35,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Enum name `PascalCaseEnum` must match one of the following formats: camelCase',
          line: 3,
          column: 21,
          endLine: 3,
          endColumn: 35,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Class name `PascalCaseClass` must match one of the following formats: camelCase',
          line: 4,
          column: 22,
          endLine: 4,
          endColumn: 37,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Function name `PascalCaseFunction` must match one of the following formats: camelCase',
          line: 5,
          column: 25,
          endLine: 5,
          endColumn: 43,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Interface name `PascalCaseInterface` must match one of the following formats: camelCase',
          line: 6,
          column: 26,
          endLine: 6,
          endColumn: 45,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Type Alias name `PascalCaseType` must match one of the following formats: camelCase',
          line: 7,
          column: 21,
          endLine: 7,
          endColumn: 35,
        },
      ],
      options: [
        {
          format: ['snake_case'],
          selector: 'default',
        },
        {
          format: ['camelCase'],
          modifiers: ['exported'],
          selector: 'variable',
        },
        {
          format: ['camelCase'],
          modifiers: ['exported'],
          selector: 'function',
        },
        {
          format: ['camelCase'],
          modifiers: ['exported'],
          selector: 'class',
        },
        {
          format: ['camelCase'],
          modifiers: ['exported'],
          selector: 'interface',
        },
        {
          format: ['camelCase'],
          modifiers: ['exported'],
          selector: 'typeAlias',
        },
        {
          format: ['camelCase'],
          modifiers: ['exported'],
          selector: 'enum',
        },
      ],
    },
    {
      code: `
        const PascalCaseVar = 1;
        enum PascalCaseEnum {}
        class PascalCaseClass {}
        function PascalCaseFunction() {}
        interface PascalCaseInterface {}
        type PascalCaseType = {};
        export {
          PascalCaseVar,
          PascalCaseEnum,
          PascalCaseClass,
          PascalCaseFunction,
          PascalCaseInterface,
          PascalCaseType,
        };
      `,
      errors: [
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Variable name `PascalCaseVar` must match one of the following formats: camelCase',
          line: 2,
          column: 15,
          endLine: 2,
          endColumn: 28,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Enum name `PascalCaseEnum` must match one of the following formats: camelCase',
          line: 3,
          column: 14,
          endLine: 3,
          endColumn: 28,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Class name `PascalCaseClass` must match one of the following formats: camelCase',
          line: 4,
          column: 15,
          endLine: 4,
          endColumn: 30,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Function name `PascalCaseFunction` must match one of the following formats: camelCase',
          line: 5,
          column: 18,
          endLine: 5,
          endColumn: 36,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Interface name `PascalCaseInterface` must match one of the following formats: camelCase',
          line: 6,
          column: 19,
          endLine: 6,
          endColumn: 38,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Type Alias name `PascalCaseType` must match one of the following formats: camelCase',
          line: 7,
          column: 14,
          endLine: 7,
          endColumn: 28,
        },
      ],
      options: [
        { format: ['snake_case'], selector: 'default' },
        {
          format: ['camelCase'],
          modifiers: ['exported'],
          selector: 'variable',
        },
        {
          format: ['camelCase'],
          modifiers: ['exported'],
          selector: 'function',
        },
        {
          format: ['camelCase'],
          modifiers: ['exported'],
          selector: 'class',
        },
        {
          format: ['camelCase'],
          modifiers: ['exported'],
          selector: 'interface',
        },
        {
          format: ['camelCase'],
          modifiers: ['exported'],
          selector: 'typeAlias',
        },
        {
          format: ['camelCase'],
          modifiers: ['exported'],
          selector: 'enum',
        },
      ],
    },
    {
      code: `
        const PascalCaseVar = 1;
        function PascalCaseFunction() {}
        declare function PascalCaseDeclaredFunction();
      `,
      errors: [
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Variable name `PascalCaseVar` must match one of the following formats: camelCase',
          line: 2,
          column: 15,
          endLine: 2,
          endColumn: 28,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Function name `PascalCaseFunction` must match one of the following formats: camelCase',
          line: 3,
          column: 18,
          endLine: 3,
          endColumn: 36,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Function name `PascalCaseDeclaredFunction` must match one of the following formats: camelCase',
          line: 4,
          column: 26,
          endLine: 4,
          endColumn: 52,
        },
      ],
      options: [
        { format: ['snake_case'], selector: 'default' },
        {
          format: ['camelCase'],
          modifiers: ['global'],
          selector: 'variable',
        },
        {
          format: ['camelCase'],
          modifiers: ['global'],
          selector: 'function',
        },
      ],
    },
    {
      code: `
        const { some_name1 } = {};
        const { some_name2 = 2 } = {};
        const { ignored: IgnoredDueToModifiers1 } = {};
        const { ignored: IgnoredDueToModifiers2 = 3 } = {};
        const IgnoredDueToModifiers3 = 1;
      `,
      errors: [
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Variable name `some_name1` must match one of the following formats: UPPER_CASE',
          line: 2,
          column: 17,
          endLine: 2,
          endColumn: 27,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Variable name `some_name2` must match one of the following formats: UPPER_CASE',
          line: 3,
          column: 17,
          endLine: 3,
          endColumn: 27,
        },
      ],
      options: [
        {
          format: ['PascalCase'],
          selector: 'default',
        },
        {
          format: ['UPPER_CASE'],
          modifiers: ['destructured'],
          selector: 'variable',
        },
      ],
    },
    {
      code: `
        export function Foo(
          { aName },
          { anotherName = 1 },
          { ignored: IgnoredDueToModifiers1 },
          { ignored: IgnoredDueToModifiers1 = 2 },
          IgnoredDueToModifiers2,
        ) {}
      `,
      errors: [
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Parameter name `aName` must match one of the following formats: UPPER_CASE',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 18,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Parameter name `anotherName` must match one of the following formats: UPPER_CASE',
          line: 4,
          column: 13,
          endLine: 4,
          endColumn: 24,
        },
      ],
      options: [
        {
          format: ['PascalCase'],
          selector: 'default',
        },
        {
          format: ['UPPER_CASE'],
          modifiers: ['destructured'],
          selector: 'parameter',
        },
      ],
    },
    {
      code: `
        class Ignored {
          private static abstract readonly some_name;
          IgnoredDueToModifiers = 1;
        }
      `,
      errors: [
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Class Property name `some_name` must match one of the following formats: UPPER_CASE',
          line: 3,
          column: 44,
          endLine: 3,
          endColumn: 53,
        },
      ],
      options: [
        {
          format: ['PascalCase'],
          selector: 'default',
        },
        {
          format: ['UPPER_CASE'],
          modifiers: ['static', 'readonly'],
          selector: 'classProperty',
        },
      ],
    },
    {
      code: `
        class Ignored {
          constructor(
            private readonly some_name,
            IgnoredDueToModifiers,
          ) {}
        }
      `,
      errors: [
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Parameter Property name `some_name` must match one of the following formats: UPPER_CASE',
          line: 4,
          column: 30,
          endLine: 4,
          endColumn: 39,
        },
      ],
      options: [
        {
          format: ['PascalCase'],
          selector: 'default',
        },
        {
          format: ['UPPER_CASE'],
          modifiers: ['readonly'],
          selector: 'parameterProperty',
        },
      ],
    },
    {
      code: `
        class Ignored {
          private static some_name() {}
          IgnoredDueToModifiers() {}
        }
      `,
      errors: [
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Class Method name `some_name` must match one of the following formats: UPPER_CASE',
          line: 3,
          column: 26,
          endLine: 3,
          endColumn: 35,
        },
      ],
      options: [
        {
          format: ['PascalCase'],
          selector: 'default',
        },
        {
          format: ['UPPER_CASE'],
          modifiers: ['static'],
          selector: 'classMethod',
        },
      ],
    },
    {
      code: `
        class Ignored {
          private static get some_name() {}
          get IgnoredDueToModifiers() {}
        }
      `,
      errors: [
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Classic Accessor name `some_name` must match one of the following formats: UPPER_CASE',
          line: 3,
          column: 30,
          endLine: 3,
          endColumn: 39,
        },
      ],
      options: [
        {
          format: ['PascalCase'],
          selector: 'default',
        },
        {
          format: ['UPPER_CASE'],
          modifiers: ['private', 'static'],
          selector: 'accessor',
        },
      ],
    },
    {
      code: `
        abstract class some_name {}
        class IgnoredDueToModifier {}
      `,
      errors: [
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Class name `some_name` must match one of the following formats: UPPER_CASE',
          line: 2,
          column: 24,
          endLine: 2,
          endColumn: 33,
        },
      ],
      options: [
        {
          format: ['PascalCase'],
          selector: 'default',
        },
        {
          format: ['UPPER_CASE'],
          modifiers: ['abstract'],
          selector: 'class',
        },
      ],
    },
    {
      code: `
        const UnusedVar = 1;
        function UnusedFunc(
          // this line is intentionally broken out
          UnusedParam: string,
        ) {}
        class UnusedClass {}
        interface UnusedInterface {}
        type UnusedType<
          // this line is intentionally broken out
          UnusedTypeParam,
        > = {};
      `,
      errors: [
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Variable name `UnusedVar` must match one of the following formats: snake_case',
          line: 2,
          column: 15,
          endLine: 2,
          endColumn: 24,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Function name `UnusedFunc` must match one of the following formats: snake_case',
          line: 3,
          column: 18,
          endLine: 3,
          endColumn: 28,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Parameter name `UnusedParam` must match one of the following formats: snake_case',
          line: 5,
          column: 11,
          endLine: 5,
          endColumn: 30,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Class name `UnusedClass` must match one of the following formats: snake_case',
          line: 7,
          column: 15,
          endLine: 7,
          endColumn: 26,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Interface name `UnusedInterface` must match one of the following formats: snake_case',
          line: 8,
          column: 19,
          endLine: 8,
          endColumn: 34,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Type Alias name `UnusedType` must match one of the following formats: snake_case',
          line: 9,
          column: 14,
          endLine: 9,
          endColumn: 24,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Type Parameter name `UnusedTypeParam` must match one of the following formats: snake_case',
          line: 11,
          column: 11,
          endLine: 11,
          endColumn: 26,
        },
      ],
      options: [
        {
          format: ['PascalCase'],
          selector: 'default',
        },
        {
          format: ['snake_case'],
          modifiers: ['unused'],
          selector: 'default',
        },
      ],
    },
    {
      code: `
        const ignored1 = {
          'a a': 1,
          'b b'() {},
          get 'c c'() {
            return 1;
          },
          set 'd d'(value: string) {},
        };
        class ignored2 {
          'a a' = 1;
          'b b'() {}
          get 'c c'() {
            return 1;
          }
          set 'd d'(value: string) {}
        }
        interface ignored3 {
          'a a': 1;
          'b b'(): void;
        }
        type ignored4 = {
          'a a': 1;
          'b b'(): void;
        };
        enum ignored5 {
          'a a',
        }
      `,
      errors: [
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Object Literal Property name `a a` must match one of the following formats: PascalCase',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 16,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Object Literal Method name `b b` must match one of the following formats: PascalCase',
          line: 4,
          column: 11,
          endLine: 4,
          endColumn: 16,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Classic Accessor name `c c` must match one of the following formats: PascalCase',
          line: 5,
          column: 15,
          endLine: 5,
          endColumn: 20,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Classic Accessor name `d d` must match one of the following formats: PascalCase',
          line: 8,
          column: 15,
          endLine: 8,
          endColumn: 20,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Class Property name `a a` must match one of the following formats: PascalCase',
          line: 11,
          column: 11,
          endLine: 11,
          endColumn: 16,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Class Method name `b b` must match one of the following formats: PascalCase',
          line: 12,
          column: 11,
          endLine: 12,
          endColumn: 16,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Classic Accessor name `c c` must match one of the following formats: PascalCase',
          line: 13,
          column: 15,
          endLine: 13,
          endColumn: 20,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Classic Accessor name `d d` must match one of the following formats: PascalCase',
          line: 16,
          column: 15,
          endLine: 16,
          endColumn: 20,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Type Property name `a a` must match one of the following formats: PascalCase',
          line: 19,
          column: 11,
          endLine: 19,
          endColumn: 16,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Type Method name `b b` must match one of the following formats: PascalCase',
          line: 20,
          column: 11,
          endLine: 20,
          endColumn: 16,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Type Property name `a a` must match one of the following formats: PascalCase',
          line: 23,
          column: 11,
          endLine: 23,
          endColumn: 16,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Type Method name `b b` must match one of the following formats: PascalCase',
          line: 24,
          column: 11,
          endLine: 24,
          endColumn: 16,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Enum Member name `a a` must match one of the following formats: PascalCase',
          line: 27,
          column: 11,
          endLine: 27,
          endColumn: 16,
        },
      ],
      options: [
        {
          format: ['snake_case'],
          selector: 'default',
        },
        {
          format: ['PascalCase'],
          modifiers: ['requiresQuotes'],
          selector: 'default',
        },
      ],
    },
    {
      code: `
        type Foo = {
          'foo     Bar': string;
          '': string;
          '0': string;
          'foo': string;
          'foo-bar': string;
          '#foo-bar': string;
        };

        interface Bar {
          'boo-----foo': string;
        }
      `,
      // 6, not 7 because 'foo' is valid
      errors: [
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Type Property name `foo     Bar` must match one of the following formats: camelCase',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 24,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Type Property name `` must match one of the following formats: camelCase',
          line: 4,
          column: 11,
          endLine: 4,
          endColumn: 13,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Type Property name `0` must match one of the following formats: camelCase',
          line: 5,
          column: 11,
          endLine: 5,
          endColumn: 14,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Type Property name `foo-bar` must match one of the following formats: camelCase',
          line: 7,
          column: 11,
          endLine: 7,
          endColumn: 20,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Type Property name `#foo-bar` must match one of the following formats: camelCase',
          line: 8,
          column: 11,
          endLine: 8,
          endColumn: 21,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Type Property name `boo-----foo` must match one of the following formats: camelCase',
          line: 12,
          column: 11,
          endLine: 12,
          endColumn: 24,
        },
      ],
    },
    {
      code: `
        class foo {
          public Bar() {
            return 42;
          }
          public async async_bar() {
            return 42;
          }
          // ❌ error
          public async asyncBar() {
            return 42;
          }
          // ❌ error
          public AsyncBar2 = async () => {
            return 42;
          };
          // ❌ error
          public AsyncBar3 = async function () {
            return 42;
          };
        }
        abstract class foo {
          public abstract Bar(): number;
          public abstract async async_bar(): number;
          // ❌ error
          public abstract async ASYNC_BAR(): number;
        }
      `,
      errors: [
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Class Method name `asyncBar` must match one of the following formats: snake_case',
          line: 10,
          column: 24,
          endLine: 10,
          endColumn: 32,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Class Method name `AsyncBar2` must match one of the following formats: snake_case',
          line: 14,
          column: 18,
          endLine: 14,
          endColumn: 27,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Class Method name `AsyncBar3` must match one of the following formats: snake_case',
          line: 18,
          column: 18,
          endLine: 18,
          endColumn: 27,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Class Method name `ASYNC_BAR` must match one of the following formats: snake_case',
          line: 26,
          column: 33,
          endLine: 26,
          endColumn: 42,
        },
      ],
      languageOptions: { parserOptions },
      options: [
        {
          format: ['camelCase'],
          selector: 'memberLike',
        },
        {
          format: ['PascalCase'],
          selector: 'method',
        },
        {
          format: ['snake_case'],
          modifiers: ['async'],
          selector: ['method', 'objectLiteralMethod'],
        },
      ],
    },
    {
      code: `
        const obj = {
          Bar() {
            return 42;
          },
          async async_bar() {
            return 42;
          },
          // ❌ error
          async AsyncBar() {
            return 42;
          },
          // ❌ error
          AsyncBar2: async () => {
            return 42;
          },
          // ❌ error
          AsyncBar3: async function () {
            return 42;
          },
        };
      `,
      errors: [
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Object Literal Method name `AsyncBar` must match one of the following formats: snake_case',
          line: 10,
          column: 17,
          endLine: 10,
          endColumn: 25,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Object Literal Method name `AsyncBar2` must match one of the following formats: snake_case',
          line: 14,
          column: 11,
          endLine: 14,
          endColumn: 20,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Object Literal Method name `AsyncBar3` must match one of the following formats: snake_case',
          line: 18,
          column: 11,
          endLine: 18,
          endColumn: 20,
        },
      ],
      languageOptions: { parserOptions },
      options: [
        {
          format: ['camelCase'],
          selector: 'memberLike',
        },
        {
          format: ['PascalCase'],
          selector: 'method',
        },
        {
          format: ['snake_case'],
          modifiers: ['async'],
          selector: ['method', 'objectLiteralMethod'],
        },
      ],
    },
    {
      code: `
        const syncbar1 = () => {};
        function syncBar2() {}
        const syncBar3 = function syncBar4() {};

        // ❌ error
        const AsyncBar1 = async () => {};
        const async_bar1 = async () => {};
        const async_bar3 = async function async_bar4() {};
        async function async_bar2() {}
        // ❌ error
        const asyncBar5 = async function async_bar6() {};
      `,
      errors: [
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Variable name `AsyncBar1` must match one of the following formats: snake_case',
          line: 7,
          column: 15,
          endLine: 7,
          endColumn: 24,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Variable name `asyncBar5` must match one of the following formats: snake_case',
          line: 12,
          column: 15,
          endLine: 12,
          endColumn: 24,
        },
      ],
      languageOptions: { parserOptions },
      options: [
        {
          format: ['camelCase'],
          selector: 'variableLike',
        },
        {
          format: ['snake_case'],
          modifiers: ['async'],
          selector: ['variableLike'],
        },
      ],
    },
    {
      code: `
        const syncbar1 = () => {};
        function syncBar2() {}
        const syncBar3 = function syncBar4() {};

        const async_bar1 = async () => {};
        // ❌ error
        async function asyncBar2() {}
        const async_bar3 = async function async_bar4() {};
        async function async_bar2() {}
        // ❌ error
        const async_bar3 = async function ASYNC_BAR4() {};
      `,
      errors: [
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Function name `asyncBar2` must match one of the following formats: snake_case',
          line: 8,
          column: 24,
          endLine: 8,
          endColumn: 33,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Function name `ASYNC_BAR4` must match one of the following formats: snake_case',
          line: 12,
          column: 43,
          endLine: 12,
          endColumn: 53,
        },
      ],
      languageOptions: { parserOptions },
      options: [
        {
          format: ['camelCase'],
          selector: 'variableLike',
        },
        {
          format: ['snake_case'],
          modifiers: ['async'],
          selector: ['variableLike'],
        },
      ],
    },
    {
      code: `
        class foo extends bar {
          public someAttribute = 1;
          public override some_attribute_override = 1;
          // ❌ error
          public override someAttributeOverride = 1;
        }
      `,
      errors: [
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Class Property name `someAttributeOverride` must match one of the following formats: snake_case',
          line: 6,
          column: 27,
          endLine: 6,
          endColumn: 48,
        },
      ],
      languageOptions: { parserOptions },
      options: [
        {
          format: ['camelCase'],
          selector: 'memberLike',
        },
        {
          format: ['snake_case'],
          modifiers: ['override'],
          selector: ['memberLike'],
        },
      ],
    },
    {
      code: `
        class foo extends bar {
          public override some_method_override() {
            return 42;
          }
          // ❌ error
          public override someMethodOverride() {
            return 42;
          }
        }
      `,
      errors: [
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Class Method name `someMethodOverride` must match one of the following formats: snake_case',
          line: 7,
          column: 27,
          endLine: 7,
          endColumn: 45,
        },
      ],
      languageOptions: { parserOptions },
      options: [
        {
          format: ['camelCase'],
          selector: 'memberLike',
        },
        {
          format: ['snake_case'],
          modifiers: ['override'],
          selector: ['memberLike'],
        },
      ],
    },
    {
      code: `
        class foo extends bar {
          public get someGetter(): string;
          public override get some_getter_override(): string;
          // ❌ error
          public override get someGetterOverride(): string;
          public set someSetter(val: string);
          public override set some_setter_override(val: string);
          // ❌ error
          public override set someSetterOverride(val: string);
        }
      `,
      errors: [
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Classic Accessor name `someGetterOverride` must match one of the following formats: snake_case',
          line: 6,
          column: 31,
          endLine: 6,
          endColumn: 49,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Classic Accessor name `someSetterOverride` must match one of the following formats: snake_case',
          line: 10,
          column: 31,
          endLine: 10,
          endColumn: 49,
        },
      ],
      languageOptions: { parserOptions },
      options: [
        {
          format: ['camelCase'],
          selector: 'memberLike',
        },
        {
          format: ['snake_case'],
          modifiers: ['override'],
          selector: ['memberLike'],
        },
      ],
    },
    {
      code: `
        class foo {
          private firstPrivateField = 1;
          // ❌ error
          private first_private_field = 1;
          // ❌ error
          #secondPrivateField = 1;
          #second_private_field = 1;
        }
      `,
      errors: [
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Class Property name `first_private_field` must match one of the following formats: camelCase',
          line: 5,
          column: 19,
          endLine: 5,
          endColumn: 38,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Class Property name `secondPrivateField` must match one of the following formats: snake_case',
          line: 7,
          column: 11,
          endLine: 7,
          endColumn: 30,
        },
      ],
      languageOptions: { parserOptions },
      options: [
        {
          format: ['camelCase'],
          selector: 'memberLike',
        },
        {
          format: ['snake_case'],
          modifiers: ['#private'],
          selector: ['memberLike'],
        },
      ],
    },
    {
      code: `
        class foo {
          private firstPrivateMethod() {}
          // ❌ error
          private first_private_method() {}
          // ❌ error
          #secondPrivateMethod() {}
          #second_private_method() {}
        }
      `,
      errors: [
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Class Method name `first_private_method` must match one of the following formats: camelCase',
          line: 5,
          column: 19,
          endLine: 5,
          endColumn: 39,
        },
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Class Method name `secondPrivateMethod` must match one of the following formats: snake_case',
          line: 7,
          column: 11,
          endLine: 7,
          endColumn: 31,
        },
      ],
      languageOptions: { parserOptions },
      options: [
        {
          format: ['camelCase'],
          selector: 'memberLike',
        },
        {
          format: ['snake_case'],
          modifiers: ['#private'],
          selector: ['memberLike'],
        },
      ],
    },
    {
      code: "import * as fooBar from 'foo_bar';",
      errors: [
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Import name `fooBar` must match one of the following formats: PascalCase',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 19,
        },
      ],
      languageOptions: { parserOptions },
      options: [
        {
          format: ['camelCase'],
          selector: ['import'],
        },
        {
          format: ['PascalCase'],
          modifiers: ['namespace'],
          selector: ['import'],
        },
      ],
    },
    {
      code: "import FooBar from 'foo_bar';",
      errors: [
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Import name `FooBar` must match one of the following formats: camelCase',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 14,
        },
      ],
      languageOptions: { parserOptions },
      options: [
        {
          format: ['camelCase'],
          selector: ['import'],
        },
        {
          format: ['PascalCase'],
          modifiers: ['namespace'],
          selector: ['import'],
        },
      ],
    },
    {
      code: "import { default as foo_bar } from 'foo_bar';",
      errors: [
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Import name `foo_bar` must match one of the following formats: camelCase',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 28,
        },
      ],
      languageOptions: { parserOptions },
      options: [
        {
          format: ['camelCase'],
          selector: ['import'],
        },
        {
          format: ['PascalCase'],
          modifiers: ['namespace'],
          selector: ['import'],
        },
      ],
    },
    {
      code: 'import { "🍎" as foo } from \'foo_bar\';',
      errors: [
        {
          messageId: 'doesNotMatchFormat',
          message:
            'Import name `foo` must match one of the following formats: PascalCase',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 21,
        },
      ],
      languageOptions: { parserOptions },
      options: [
        {
          format: ['PascalCase'],
          selector: ['import'],
        },
      ],
    },
  ],
  valid: [
    {
      code: `
        const child_process = require('child_process');
      `,
      languageOptions: { parserOptions },
      options: [
        {
          filter: {
            match: false,
            regex: 'child_process',
          },
          format: ['camelCase'],
          selector: 'default',
        },
      ],
    },
    {
      code: `
        declare const ANY_UPPER_CASE: any;
        declare const ANY_UPPER_CASE: any | null;
        declare const ANY_UPPER_CASE: any | null | undefined;

        declare const string_camelCase: string;
        declare const string_camelCase: string | null;
        declare const string_camelCase: string | null | undefined;
        declare const string_camelCase: 'a' | null | undefined;
        declare const string_camelCase: string | 'a' | null | undefined;

        declare const number_camelCase: number;
        declare const number_camelCase: number | null;
        declare const number_camelCase: number | null | undefined;
        declare const number_camelCase: 1 | null | undefined;
        declare const number_camelCase: number | 2 | null | undefined;

        declare const boolean_camelCase: boolean;
        declare const boolean_camelCase: boolean | null;
        declare const boolean_camelCase: boolean | null | undefined;
        declare const boolean_camelCase: true | null | undefined;
        declare const boolean_camelCase: false | null | undefined;
        declare const boolean_camelCase: true | false | null | undefined;
      `,
      languageOptions: { parserOptions },
      options: [
        {
          format: ['UPPER_CASE'],
          modifiers: ['const'],
          prefix: ['ANY_'],
          selector: 'variable',
        },
        {
          format: ['camelCase'],
          prefix: ['string_'],
          selector: 'variable',
          types: ['string'],
        },
        {
          format: ['camelCase'],
          prefix: ['number_'],
          selector: 'variable',
          types: ['number'],
        },
        {
          format: ['camelCase'],
          prefix: ['boolean_'],
          selector: 'variable',
          types: ['boolean'],
        },
      ],
    },
    {
      code: `
        let foo = 'a';
        const _foo = 1;
        interface Foo {}
        class Bar {}
        function foo_function_bar() {}
      `,
      options: [
        {
          custom: {
            match: false,
            regex: /^unused_\w/.source,
          },
          format: ['camelCase'],
          leadingUnderscore: 'allow',
          selector: 'default',
        },
        {
          custom: {
            match: false,
            regex: /^I[A-Z]/.source,
          },
          format: ['PascalCase'],
          selector: 'typeLike',
        },
        {
          custom: {
            match: true,
            regex: /_function_/.source,
          },
          format: ['snake_case'],
          leadingUnderscore: 'allow',
          selector: 'function',
        },
      ],
    },
    {
      code: `
        let foo = 'a';
        const _foo = 1;
        interface foo {}
        class bar {}
        function fooFunctionBar() {}
        function _fooFunctionBar() {}
      `,
      options: [
        {
          custom: {
            match: false,
            regex: /^unused_\w/.source,
          },
          format: ['camelCase'],
          leadingUnderscore: 'allow',
          selector: ['default', 'typeLike', 'function'],
        },
      ],
    },
    {
      code: `
        const match = 'test'.match(/test/);
        const [, key, value] = match;
      `,
      options: [
        {
          format: ['camelCase'],
          selector: 'default',
        },
      ],
    },
    // no format selector
    {
      code: 'const snake_case = 1;',
      options: [
        {
          format: ['camelCase'],
          selector: 'default',
        },
        {
          format: null,
          selector: 'variable',
        },
      ],
    },
    {
      code: 'const snake_case = 1;',
      options: [
        {
          format: ['camelCase'],
          selector: 'default',
        },
        {
          format: [],
          selector: 'variable',
        },
      ],
    },
    // https://github.com/typescript-eslint/typescript-eslint/issues/1478
    {
      code: `
        const child_process = require('child_process');
      `,
      options: [
        { format: ['camelCase', 'UPPER_CASE'], selector: 'variable' },
        {
          filter: 'child_process',
          format: ['snake_case'],
          selector: 'variable',
        },
      ],
    },
    {
      code: `
        const foo = {
          'Property-Name': 'asdf',
        };
      `,
      options: [
        {
          filter: {
            match: false,
            regex: /-/.source,
          },
          format: ['strictCamelCase'],
          selector: 'default',
        },
      ],
    },
    {
      code: `
        const foo = {
          'Property-Name': 'asdf',
        };
      `,
      options: [
        {
          filter: {
            match: false,
            regex: /^(Property-Name)$/.source,
          },
          format: ['strictCamelCase'],
          selector: 'default',
        },
      ],
    },
    {
      code: `
        let isFoo = 1;
        class foo {
          shouldBoo: number;
        }
      `,
      languageOptions: { parserOptions },
      options: [
        {
          format: ['PascalCase'],
          prefix: ['is', 'should', 'has', 'can', 'did', 'will'],
          selector: ['variable', 'parameter', 'property', 'accessor'],
          types: ['number'],
        },
      ],
    },
    {
      code: `
        class foo {
          private readonly FooBoo: boolean;
        }
      `,
      languageOptions: { parserOptions },
      options: [
        {
          format: ['PascalCase'],
          modifiers: ['private', 'readonly'],
          selector: ['property', 'accessor'],
          types: ['boolean'],
        },
      ],
    },
    {
      code: `
        class foo {
          private fooBoo: number;
        }
      `,
      options: [
        {
          format: ['camelCase'],
          modifiers: ['private'],
          selector: ['property', 'accessor'],
        },
      ],
    },
    {
      code: `
        const isfooBar = 1;
        function fun(goodfunFoo: number) {}
        class foo {
          private VanFooBar: number;
        }
      `,
      languageOptions: { parserOptions },
      options: [
        {
          format: ['StrictPascalCase'],
          modifiers: ['private'],
          prefix: ['Van'],
          selector: ['property', 'accessor'],
        },
        {
          format: ['camelCase'],
          prefix: ['is', 'good'],
          selector: ['variable', 'parameter'],
          types: ['number'],
        },
      ],
    },
    {
      code: `
        class SomeClass {
          static OtherConstant = 'hello';
        }

        export const { OtherConstant: otherConstant } = SomeClass;
      `,
      options: [
        { format: ['PascalCase'], selector: 'property' },
        { format: ['camelCase'], selector: 'variable' },
      ],
    },
    // treat properties with function expressions as typeMethod
    {
      code: `
        interface SOME_INTERFACE {
          SomeMethod: () => void;

          some_property: string;
        }
      `,
      options: [
        {
          format: ['UPPER_CASE'],
          selector: 'default',
        },
        {
          format: ['PascalCase'],
          selector: 'typeMethod',
        },
        {
          format: ['snake_case'],
          selector: 'typeProperty',
        },
      ],
    },
    {
      code: `
        type Ignored = {
          ignored_due_to_modifiers: string;
          readonly FOO: string;
        };
      `,
      languageOptions: { parserOptions },
      options: [
        {
          format: ['UPPER_CASE'],
          modifiers: ['readonly'],
          selector: 'typeProperty',
        },
      ],
    },
    {
      code: `
        const camelCaseVar = 1;
        enum camelCaseEnum {}
        class camelCaseClass {}
        function camelCaseFunction() {}
        interface camelCaseInterface {}
        type camelCaseType = {};
        export const PascalCaseVar = 1;
        export enum PascalCaseEnum {}
        export class PascalCaseClass {}
        export function PascalCaseFunction() {}
        export interface PascalCaseInterface {}
        export type PascalCaseType = {};
      `,
      options: [
        { format: ['camelCase'], selector: 'default' },
        {
          format: ['PascalCase'],
          modifiers: ['exported'],
          selector: 'variable',
        },
        {
          format: ['PascalCase'],
          modifiers: ['exported'],
          selector: 'function',
        },
        {
          format: ['PascalCase'],
          modifiers: ['exported'],
          selector: 'class',
        },
        {
          format: ['PascalCase'],
          modifiers: ['exported'],
          selector: 'interface',
        },
        {
          format: ['PascalCase'],
          modifiers: ['exported'],
          selector: 'typeAlias',
        },
        {
          format: ['PascalCase'],
          modifiers: ['exported'],
          selector: 'enum',
        },
      ],
    },
    {
      code: `
        const camelCaseVar = 1;
        enum camelCaseEnum {}
        class camelCaseClass {}
        function camelCaseFunction() {}
        interface camelCaseInterface {}
        type camelCaseType = {};
        const PascalCaseVar = 1;
        enum PascalCaseEnum {}
        class PascalCaseClass {}
        function PascalCaseFunction() {}
        interface PascalCaseInterface {}
        type PascalCaseType = {};
        export {
          PascalCaseVar,
          PascalCaseEnum,
          PascalCaseClass,
          PascalCaseFunction,
          PascalCaseInterface,
          PascalCaseType,
        };
      `,
      options: [
        { format: ['camelCase'], selector: 'default' },
        {
          format: ['PascalCase'],
          modifiers: ['exported'],
          selector: 'variable',
        },
        {
          format: ['PascalCase'],
          modifiers: ['exported'],
          selector: 'function',
        },
        {
          format: ['PascalCase'],
          modifiers: ['exported'],
          selector: 'class',
        },
        {
          format: ['PascalCase'],
          modifiers: ['exported'],
          selector: 'interface',
        },
        {
          format: ['PascalCase'],
          modifiers: ['exported'],
          selector: 'typeAlias',
        },
        {
          format: ['PascalCase'],
          modifiers: ['exported'],
          selector: 'enum',
        },
      ],
    },
    {
      code: `
        {
          const camelCaseVar = 1;
          function camelCaseFunction() {}
          declare function camelCaseDeclaredFunction();
        }
        const PascalCaseVar = 1;
        function PascalCaseFunction() {}
        declare function PascalCaseDeclaredFunction();
      `,
      options: [
        { format: ['camelCase'], selector: 'default' },
        {
          format: ['PascalCase'],
          modifiers: ['global'],
          selector: 'variable',
        },
        {
          format: ['PascalCase'],
          modifiers: ['global'],
          selector: 'function',
        },
      ],
    },
    {
      code: `
        const { some_name1 } = {};
        const { ignore: IgnoredDueToModifiers1 } = {};
        const { some_name2 = 2 } = {};
        const IgnoredDueToModifiers2 = 1;
      `,
      options: [
        {
          format: ['PascalCase'],
          selector: 'default',
        },
        {
          format: ['snake_case'],
          modifiers: ['destructured'],
          selector: 'variable',
        },
      ],
    },
    {
      code: `
        const { some_name1 } = {};
        const { ignore: IgnoredDueToModifiers1 } = {};
        const { some_name2 = 2 } = {};
        const IgnoredDueToModifiers2 = 1;
      `,
      options: [
        {
          format: ['PascalCase'],
          selector: 'default',
        },
        {
          format: null,
          modifiers: ['destructured'],
          selector: 'variable',
        },
      ],
    },
    {
      code: `
        export function Foo(
          { aName },
          { anotherName = 1 },
          { ignored: IgnoredDueToModifiers1 },
          { ignored: IgnoredDueToModifiers1 = 2 },
          IgnoredDueToModifiers2,
        ) {}
      `,
      options: [
        {
          format: ['PascalCase'],
          selector: 'default',
        },
        {
          format: ['camelCase'],
          modifiers: ['destructured'],
          selector: 'parameter',
        },
      ],
    },
    {
      code: `
        class Ignored {
          private static abstract readonly some_name;
          IgnoredDueToModifiers = 1;
        }
      `,
      options: [
        {
          format: ['PascalCase'],
          selector: 'default',
        },
        {
          format: ['snake_case'],
          modifiers: ['static', 'readonly'],
          selector: 'classProperty',
        },
      ],
    },
    {
      code: `
        class Ignored {
          constructor(
            private readonly some_name,
            IgnoredDueToModifiers,
          ) {}
        }
      `,
      options: [
        {
          format: ['PascalCase'],
          selector: 'default',
        },
        {
          format: ['snake_case'],
          modifiers: ['readonly'],
          selector: 'parameterProperty',
        },
      ],
    },
    {
      code: `
        class Ignored {
          private static some_name() {}
          IgnoredDueToModifiers() {}
        }
      `,
      options: [
        {
          format: ['PascalCase'],
          selector: 'default',
        },
        {
          format: ['snake_case'],
          modifiers: ['static'],
          selector: 'classMethod',
        },
      ],
    },
    {
      code: `
        class Ignored {
          private static get some_name() {}
          get IgnoredDueToModifiers() {}
        }
      `,
      options: [
        {
          format: ['PascalCase'],
          selector: 'default',
        },
        {
          format: ['snake_case'],
          modifiers: ['private', 'static'],
          selector: 'accessor',
        },
      ],
    },
    {
      code: `
        abstract class some_name {}
        class IgnoredDueToModifier {}
      `,
      options: [
        {
          format: ['PascalCase'],
          selector: 'default',
        },
        {
          format: ['snake_case'],
          modifiers: ['abstract'],
          selector: 'class',
        },
      ],
    },
    {
      code: `
        const UnusedVar = 1;
        function UnusedFunc(
          // this line is intentionally broken out
          UnusedParam: string,
        ) {}
        class UnusedClass {}
        interface UnusedInterface {}
        type UnusedType<
          // this line is intentionally broken out
          UnusedTypeParam,
        > = {};

        export const used_var = 1;
        export function used_func(
          // this line is intentionally broken out
          used_param: string,
        ) {
          return used_param;
        }
        export class used_class {}
        export interface used_interface {}
        export type used_type<
          // this line is intentionally broken out
          used_typeparam,
        > = used_typeparam;
      `,
      options: [
        {
          format: ['snake_case'],
          selector: 'default',
        },
        {
          format: ['PascalCase'],
          modifiers: ['unused'],
          selector: 'default',
        },
      ],
    },
    {
      code: `
        const ignored1 = {
          'a a': 1,
          'b b'() {},
          get 'c c'() {
            return 1;
          },
          set 'd d'(value: string) {},
        };
        class ignored2 {
          'a a' = 1;
          'b b'() {}
          get 'c c'() {
            return 1;
          }
          set 'd d'(value: string) {}
        }
        interface ignored3 {
          'a a': 1;
          'b b'(): void;
        }
        type ignored4 = {
          'a a': 1;
          'b b'(): void;
        };
        enum ignored5 {
          'a a',
        }
      `,
      options: [
        {
          format: ['snake_case'],
          selector: 'default',
        },
        {
          format: null,
          modifiers: ['requiresQuotes'],
          selector: 'default',
        },
      ],
    },
    {
      code: `
        const ignored1 = {
          'a a': 1,
          'b b'() {},
          get 'c c'() {
            return 1;
          },
          set 'd d'(value: string) {},
        };
        class ignored2 {
          'a a' = 1;
          'b b'() {}
          get 'c c'() {
            return 1;
          }
          set 'd d'(value: string) {}
        }
        interface ignored3 {
          'a a': 1;
          'b b'(): void;
        }
        type ignored4 = {
          'a a': 1;
          'b b'(): void;
        };
        enum ignored5 {
          'a a',
        }
      `,
      options: [
        {
          format: ['snake_case'],
          selector: 'default',
        },
        {
          format: null,
          modifiers: ['requiresQuotes'],
          selector: [
            'classProperty',
            'objectLiteralProperty',
            'typeProperty',
            'classMethod',
            'objectLiteralMethod',
            'typeMethod',
            'accessor',
            'enumMember',
          ],
        },
        // making sure the `requiresQuotes` modifier appropriately overrides this
        {
          format: ['PascalCase'],
          selector: [
            'classProperty',
            'objectLiteralProperty',
            'typeProperty',
            'classMethod',
            'objectLiteralMethod',
            'typeMethod',
            'accessor',
            'enumMember',
          ],
        },
      ],
    },
    {
      code: `
        const obj = {
          Foo: 42,
          Bar() {
            return 42;
          },
        };
      `,
      languageOptions: { parserOptions },
      options: [
        {
          format: ['camelCase'],
          selector: 'memberLike',
        },
        {
          format: ['PascalCase'],
          selector: 'property',
        },
        {
          format: ['PascalCase'],
          selector: 'method',
        },
      ],
    },
    {
      code: `
        const obj = {
          Bar() {
            return 42;
          },
          async async_bar() {
            return 42;
          },
        };
        class foo {
          public Bar() {
            return 42;
          }
          public async async_bar() {
            return 42;
          }
        }
        abstract class foo {
          public Bar() {
            return 42;
          }
          public async async_bar() {
            return 42;
          }
        }
      `,
      languageOptions: { parserOptions },
      options: [
        {
          format: ['camelCase'],
          selector: 'memberLike',
        },
        {
          format: ['snake_case'],
          modifiers: ['async'],
          selector: ['method', 'objectLiteralMethod'],
        },
        {
          format: ['PascalCase'],
          selector: 'method',
        },
      ],
    },
    {
      code: `
        const async_bar1 = async () => {};
        async function async_bar2() {}
        const async_bar3 = async function async_bar4() {};
      `,
      languageOptions: { parserOptions },
      options: [
        {
          format: ['camelCase'],
          selector: 'memberLike',
        },
        {
          format: ['PascalCase'],
          selector: 'method',
        },
        {
          format: ['snake_case'],
          modifiers: ['async'],
          selector: ['variable'],
        },
      ],
    },
    {
      code: `
        class foo extends bar {
          public someAttribute = 1;
          public override some_attribute_override = 1;
          public someMethod() {
            return 42;
          }
          public override some_method_override2() {
            return 42;
          }
        }
        abstract class foo extends bar {
          public abstract someAttribute: string;
          public abstract override some_attribute_override: string;
          public abstract someMethod(): string;
          public abstract override some_method_override2(): string;
        }
      `,
      languageOptions: { parserOptions },
      options: [
        {
          format: ['camelCase'],
          selector: 'memberLike',
        },
        {
          format: ['snake_case'],
          modifiers: ['override'],
          selector: ['memberLike'],
        },
      ],
    },
    {
      code: `
        class foo {
          private someAttribute = 1;
          #some_attribute = 1;

          private someMethod() {}
          #some_method() {}
        }
      `,
      languageOptions: { parserOptions },
      options: [
        {
          format: ['camelCase'],
          selector: 'memberLike',
        },
        {
          format: ['snake_case'],
          modifiers: ['#private'],
          selector: ['memberLike'],
        },
      ],
    },
    {
      code: "import * as FooBar from 'foo_bar';",
      languageOptions: { parserOptions },
      options: [
        {
          format: ['PascalCase'],
          selector: ['import'],
        },
        {
          format: ['camelCase'],
          modifiers: ['default'],
          selector: ['import'],
        },
      ],
    },
    {
      code: "import fooBar from 'foo_bar';",
      languageOptions: { parserOptions },
      options: [
        {
          format: ['PascalCase'],
          selector: ['import'],
        },
        {
          format: ['camelCase'],
          modifiers: ['default'],
          selector: ['import'],
        },
      ],
    },
    {
      code: "import { default as fooBar } from 'foo_bar';",
      languageOptions: { parserOptions },
      options: [
        {
          format: ['PascalCase'],
          selector: ['import'],
        },
        {
          format: ['camelCase'],
          modifiers: ['default'],
          selector: ['import'],
        },
      ],
    },
    {
      code: "import { foo_bar } from 'foo_bar';",
      languageOptions: { parserOptions },
      options: [
        {
          format: ['PascalCase'],
          selector: ['import'],
        },
        {
          format: ['camelCase'],
          modifiers: ['default'],
          selector: ['import'],
        },
      ],
    },
    {
      code: 'import { "🍎" as Foo } from \'foo_bar\';',
      languageOptions: { parserOptions },
      options: [
        {
          format: ['PascalCase'],
          selector: ['import'],
        },
      ],
    },
  ],
});

// Preserve the existing Rslint regression cases, with their original expectations
// independently verified and completed using the same upstream version.
// Three historical cases contain abstract method bodies; their upstream
// verification uses allowInvalidAST, while Rslint already parses these inputs.
describe('existing Rslint regressions', () => {
  ruleTester.run('naming-convention', {
    valid: [
      {
        code: '\n        class Ignored {\n          private static abstract some_name() {}\n          IgnoredDueToModifiers() {}\n        }\n      ',
        options: [
          {
            format: ['PascalCase'],
            selector: 'default',
          },
          {
            format: ['snake_case'],
            modifiers: ['abstract', 'static'],
            selector: 'classMethod',
          },
        ],
      },
      {
        code: '\n        const obj = {\n          Bar() {\n            return 42;\n          },\n          async async_bar() {\n            return 42;\n          },\n        };\n        class foo {\n          public Bar() {\n            return 42;\n          }\n          public async async_bar() {\n            return 42;\n          }\n        }\n        abstract class foo {\n          public abstract Bar() {\n            return 42;\n          }\n          public abstract async async_bar() {\n            return 42;\n          }\n        }\n      ',
        languageOptions: {
          parserOptions: {
            project: './tsconfig.json',
          },
        },
        options: [
          {
            format: ['camelCase'],
            selector: 'memberLike',
          },
          {
            format: ['snake_case'],
            modifiers: ['async'],
            selector: ['method', 'objectLiteralMethod'],
          },
          {
            format: ['PascalCase'],
            selector: 'method',
          },
        ],
      },
      {
        code: 'const _ = 1;',
        options: [
          {
            format: ['camelCase'],
            leadingUnderscore: 'allow',
            selector: 'variable',
          },
        ],
      },
      {
        code: 'const [_, setA] = [1, 2];',
        options: [
          {
            format: ['camelCase'],
            leadingUnderscore: 'allow',
            selector: 'variable',
          },
        ],
      },
      {
        code: 'function f(_: number) {}',
        options: [
          {
            format: ['camelCase'],
            leadingUnderscore: 'allow',
            selector: 'parameter',
          },
        ],
      },
      {
        code: 'const $cache = 1; const $svg = 1; const x$y = 1;',
        options: [
          {
            format: ['camelCase'],
            selector: 'variable',
          },
        ],
      },
      {
        code: 'const $Root = 1;',
        options: [
          {
            format: ['PascalCase'],
            selector: 'variable',
          },
        ],
      },
      {
        code: 'class C { private $foo = 1; }',
        options: [
          {
            format: ['camelCase'],
            selector: 'property',
          },
        ],
      },
      {
        code: 'const fooBar$Baz = 1;',
        options: [
          {
            format: ['strictCamelCase'],
            selector: 'variable',
          },
        ],
      },
      {
        code: 'const $cache = 1;',
        options: [
          {
            format: ['camelCase', 'UPPER_CASE', 'PascalCase'],
            leadingUnderscore: 'allow',
            filter: {
              regex:
                '(^_?[a-z][a-z0-9]*([A-Z][a-z]?[a-z0-9]*)*_[a-z][a-z0-9]*([A-Z][a-z]?[a-z0-9]*)*$)',
              match: false,
            },
            selector: 'variable',
          },
        ],
      },
      {
        code: 'const evaluateThunk = (0, eval)("") as (System: any, ...globalArgs: any[]) => void;',
        options: [
          {
            format: ['camelCase'],
            selector: 'parameter',
          },
        ],
      },
      {
        code: 'type Cast = (System: any) => void;',
        options: [
          {
            format: ['camelCase'],
            selector: 'parameter',
          },
        ],
      },
      {
        code: 'interface Foo { method(System: any): void; }',
        options: [
          {
            format: ['camelCase'],
            selector: 'parameter',
          },
        ],
      },
      {
        code: 'interface Foo { (System: any): void; }',
        options: [
          {
            format: ['camelCase'],
            selector: 'parameter',
          },
        ],
      },
      {
        code: 'interface Foo { new (System: any): void; }',
        options: [
          {
            format: ['camelCase'],
            selector: 'parameter',
          },
        ],
      },
      {
        code: 'interface Foo { [System: string]: any; }',
        options: [
          {
            format: ['camelCase'],
            selector: 'parameter',
          },
        ],
      },
      {
        code: 'const _ = 1;',
        options: [
          {
            format: ['camelCase'],
            leadingUnderscore: 'allow',
            selector: 'variable',
          },
        ],
      },
      {
        code: 'const __ = 1;',
        options: [
          {
            format: ['camelCase'],
            leadingUnderscore: 'allowDouble',
            selector: 'variable',
          },
        ],
      },
      {
        code: 'const IS = 1;',
        options: [
          {
            format: ['camelCase'],
            prefix: ['IS'],
            selector: 'variable',
          },
        ],
      },
      {
        code: 'const $a = 1; const a$ = 1; const a$b = 1;',
        options: [
          {
            format: ['camelCase'],
            selector: 'variable',
          },
        ],
      },
      {
        code: 'const $A = 1; const A$ = 1; const A$B = 1;',
        options: [
          {
            format: ['PascalCase'],
            selector: 'variable',
          },
        ],
      },
      {
        code: 'const FOO$BAR = 1;',
        options: [
          {
            format: ['UPPER_CASE'],
            selector: 'variable',
          },
        ],
      },
      {
        code: '\n        const obj = { $fn() { return 1; } };\n        const $flag = true;\n      ',
        options: [
          {
            format: ['camelCase'],
            selector: 'variable',
          },
          {
            format: ['camelCase'],
            selector: 'objectLiteralMethod',
          },
        ],
      },
      {
        code: 'const foo$$Bar = 1;',
        options: [
          {
            format: ['strictCamelCase'],
            selector: 'variable',
          },
        ],
      },
      {
        code: 'const F$ooBar = 1;',
        options: [
          {
            format: ['StrictPascalCase'],
            selector: 'variable',
          },
        ],
      },
    ],
    invalid: [
      {
        code: "\n        let unused_foo = 'a';\n        const _unused_foo = 1;\n        interface IFoo {}\n        class IBar {}\n        function fooBar() {}\n      ",
        errors: [
          {
            messageId: 'satisfyCustom',
            message:
              'Variable name `unused_foo` must not match the RegExp: /^unused_\\w/u',
            line: 2,
            column: 13,
            endLine: 2,
            endColumn: 23,
          },
          {
            messageId: 'satisfyCustom',
            message:
              'Variable name `_unused_foo` must not match the RegExp: /^unused_\\w/u',
            line: 3,
            column: 15,
            endLine: 3,
            endColumn: 26,
          },
          {
            messageId: 'satisfyCustom',
            message:
              'Interface name `IFoo` must not match the RegExp: /^I[A-Z]/u',
            line: 4,
            column: 19,
            endLine: 4,
            endColumn: 23,
          },
          {
            messageId: 'satisfyCustom',
            message: 'Class name `IBar` must not match the RegExp: /^I[A-Z]/u',
            line: 5,
            column: 15,
            endLine: 5,
            endColumn: 19,
          },
          {
            messageId: 'satisfyCustom',
            message:
              'Function name `fooBar` must match the RegExp: /function/u',
            line: 6,
            column: 18,
            endLine: 6,
            endColumn: 24,
          },
        ],
        options: [
          {
            custom: {
              match: false,
              regex: '^unused_\\w',
            },
            format: ['snake_case'],
            leadingUnderscore: 'allow',
            selector: 'default',
          },
          {
            custom: {
              match: false,
              regex: '^I[A-Z]',
            },
            format: ['PascalCase'],
            selector: 'typeLike',
          },
          {
            custom: {
              match: true,
              regex: 'function',
            },
            format: ['camelCase'],
            leadingUnderscore: 'allow',
            selector: 'function',
          },
        ],
      },
      {
        code: "\n        let unused_foo = 'a';\n        const _unused_foo = 1;\n        function foo_bar() {}\n        interface IFoo {}\n        class IBar {}\n      ",
        errors: [
          {
            messageId: 'doesNotMatchFormat',
            message:
              'Variable name `unused_foo` must match one of the following formats: camelCase',
            line: 2,
            column: 13,
            endLine: 2,
            endColumn: 23,
          },
          {
            messageId: 'doesNotMatchFormatTrimmed',
            message:
              'Variable name `_unused_foo` trimmed as `unused_foo` must match one of the following formats: camelCase',
            line: 3,
            column: 15,
            endLine: 3,
            endColumn: 26,
          },
          {
            messageId: 'doesNotMatchFormat',
            message:
              'Function name `foo_bar` must match one of the following formats: camelCase',
            line: 4,
            column: 18,
            endLine: 4,
            endColumn: 25,
          },
          {
            messageId: 'satisfyCustom',
            message:
              'Interface name `IFoo` must not match the RegExp: /^I[A-Z]/u',
            line: 5,
            column: 19,
            endLine: 5,
            endColumn: 23,
          },
          {
            messageId: 'satisfyCustom',
            message: 'Class name `IBar` must not match the RegExp: /^I[A-Z]/u',
            line: 6,
            column: 15,
            endLine: 6,
            endColumn: 19,
          },
        ],
        options: [
          {
            format: ['camelCase'],
            leadingUnderscore: 'allow',
            selector: ['variable', 'function'],
          },
          {
            custom: {
              match: false,
              regex: '^I[A-Z]',
            },
            format: ['PascalCase'],
            selector: ['class', 'interface'],
          },
        ],
      },
      {
        code: '\n        class Ignored {\n          private static abstract some_name() {}\n          IgnoredDueToModifiers() {}\n        }\n      ',
        errors: [
          {
            messageId: 'doesNotMatchFormat',
            message:
              'Class Method name `some_name` must match one of the following formats: UPPER_CASE',
            line: 3,
            column: 35,
            endLine: 3,
            endColumn: 44,
          },
        ],
        options: [
          {
            format: ['PascalCase'],
            selector: 'default',
          },
          {
            format: ['UPPER_CASE'],
            modifiers: ['abstract', 'static'],
            selector: 'classMethod',
          },
        ],
      },
      {
        code: 'const _ = 1;',
        errors: [
          {
            messageId: 'missingUnderscore',
            message: 'Variable name `_` must have two leading underscore(s).',
            line: 1,
            column: 7,
            endLine: 1,
            endColumn: 8,
          },
        ],
        options: [
          {
            format: ['camelCase'],
            leadingUnderscore: 'requireDouble',
            selector: 'variable',
          },
        ],
      },
      {
        code: 'const _x = 1;',
        errors: [
          {
            messageId: 'unexpectedUnderscore',
            message: 'Variable name `_x` must not have a leading underscore.',
            line: 1,
            column: 7,
            endLine: 1,
            endColumn: 9,
          },
        ],
        options: [
          {
            format: ['camelCase'],
            leadingUnderscore: 'forbid',
            selector: 'variable',
          },
        ],
      },
      {
        code: 'const FOOBar = 1;',
        errors: [
          {
            messageId: 'doesNotMatchFormat',
            message:
              'Variable name `FOOBar` must match one of the following formats: StrictPascalCase',
            line: 1,
            column: 7,
            endLine: 1,
            endColumn: 13,
          },
        ],
        options: [
          {
            format: ['StrictPascalCase'],
            selector: 'variable',
          },
        ],
      },
      {
        code: 'const foo__bar = 1;',
        errors: [
          {
            messageId: 'doesNotMatchFormat',
            message:
              'Variable name `foo__bar` must match one of the following formats: snake_case',
            line: 1,
            column: 7,
            endLine: 1,
            endColumn: 15,
          },
        ],
        options: [
          {
            format: ['snake_case'],
            selector: 'variable',
          },
        ],
      },
      {
        code: 'const foo_ = 1;',
        errors: [
          {
            messageId: 'doesNotMatchFormat',
            message:
              'Variable name `foo_` must match one of the following formats: snake_case',
            line: 1,
            column: 7,
            endLine: 1,
            endColumn: 11,
          },
        ],
        options: [
          {
            format: ['snake_case'],
            selector: 'variable',
          },
        ],
      },
      {
        code: 'const FOO__BAR = 1;',
        errors: [
          {
            messageId: 'doesNotMatchFormat',
            message:
              'Variable name `FOO__BAR` must match one of the following formats: UPPER_CASE',
            line: 1,
            column: 7,
            endLine: 1,
            endColumn: 15,
          },
        ],
        options: [
          {
            format: ['UPPER_CASE'],
            selector: 'variable',
          },
        ],
      },
      {
        code: 'const foo_bar = 1;',
        errors: [
          {
            messageId: 'doesNotMatchFormat',
            message:
              'Variable name `foo_bar` must match one of the following formats: camelCase',
            line: 1,
            column: 7,
            endLine: 1,
            endColumn: 14,
          },
        ],
        options: [
          {
            format: ['camelCase'],
            selector: 'variable',
          },
        ],
      },
      {
        code: 'class C { foo() {} }',
        errors: [
          {
            messageId: 'doesNotMatchFormat',
            message:
              'Class Method name `foo` must match one of the following formats: PascalCase',
            line: 1,
            column: 11,
            endLine: 1,
            endColumn: 14,
          },
        ],
        options: [
          {
            format: ['PascalCase'],
            selector: 'method',
          },
        ],
      },
    ],
  });
});

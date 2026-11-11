// Every case from typescript-eslint v8.67.0, with messages and full ranges
// additionally pinned from the same upstream version. No upstream expectations are relaxed.
// https://github.com/typescript-eslint/typescript-eslint/blob/v8.67.0/packages/eslint-plugin/tests/rules/no-misused-spread.test.ts
import { RuleTester } from '@rslint/rule-tester';

const ruleTester = new RuleTester({
  plugin: '@typescript-eslint',
  fixtureFiles: {
    // The upstream fixture has @types/node's `module` declaration to augment.
    'module.d.ts': "declare module 'module' {}",
    'tsconfig.json':
      '{"compilerOptions":{"jsx":"preserve","target":"es2015","module":"commonjs","strict":true,"types":[],"lib":["esnext","DOM"],"experimentalDecorators":true},"include":["*.ts","*.tsx"]}',
  },
  languageOptions: {
    parserOptions: {
      project: './tsconfig.json',
    },
  },
});

ruleTester.run('no-misused-spread', {
  valid: [
    {
      code: 'const a = [...[1, 2, 3]];',
      filename: 'file.ts',
    },
    {
      code: 'const a = [...([1, 2, 3] as const)];',
      filename: 'file.ts',
    },
    {
      code: `
declare const data: any;
const a = [...data];
    `,
      filename: 'file.ts',
    },
    {
      code: `
declare const data: unknown;
const a = [...data];
    `,
      filename: 'file.ts',
    },
    {
      code: `
const a = [1, 2, 3];
const b = [...a];
    `,
      filename: 'file.ts',
    },
    {
      code: `
const a = [1, 2, 3] as const;
const b = [...a];
    `,
      filename: 'file.ts',
    },
    {
      code: `
declare function getArray(): number[];
const a = [...getArray()];
    `,
      filename: 'file.ts',
    },
    {
      code: `
declare function getTuple(): readonly number[];
const a = [...getTuple()];
    `,
      filename: 'file.ts',
    },
    {
      code: `
const iterator = {
  *[Symbol.iterator]() {
    yield 1;
    yield 2;
    yield 3;
  },
};

const a = [...iterator];
    `,
      filename: 'file.ts',
    },
    {
      code: `
declare const data: Iterable<number> | number[];

const a = [...data];
    `,
      filename: 'file.ts',
    },
    {
      code: `
declare const data: Iterable<number> & number[];

const a = [...data];
    `,
      filename: 'file.ts',
    },
    {
      code: `
declare function getIterable(): Iterable<number>;

const a = [...getIterable()];
    `,
      filename: 'file.ts',
    },
    {
      code: `
declare const data: Uint8Array;

const a = [...data];
    `,
      filename: 'file.ts',
    },
    {
      code: `
declare const data: TypedArray;

const a = [...data];
    `,
      filename: 'file.ts',
    },
    {
      code: 'const o = { ...{ a: 1, b: 2 } };',
      filename: 'file.ts',
    },
    {
      code: 'const o = { ...({ a: 1, b: 2 } as const) };',
      filename: 'file.ts',
    },
    {
      code: `
declare const obj: any;

const o = { ...obj };
    `,
      filename: 'file.ts',
    },
    {
      code: `
declare const obj: { a: number; b: number } | any;

const o = { ...obj };
    `,
      filename: 'file.ts',
    },
    {
      code: `
declare const obj: { a: number; b: number } & any;

const o = { ...obj };
    `,
      filename: 'file.ts',
    },
    {
      code: `
const obj = { a: 1, b: 2 };
const o = { ...obj };
    `,
      filename: 'file.ts',
    },
    {
      code: `
declare const obj: { a: number; b: number };
const o = { ...obj };
    `,
      filename: 'file.ts',
    },
    {
      code: `
declare function getObject(): { a: number; b: number };
const o = { ...getObject() };
    `,
      filename: 'file.ts',
    },
    {
      code: `
function f() {}

f.prop = 1;

const o = { ...f };
    `,
      filename: 'file.ts',
    },
    {
      code: `
const f = () => {};

f.prop = 1;

const o = { ...f };
    `,
      filename: 'file.ts',
    },
    {
      code: `
function* generator() {}

generator.prop = 1;

const o = { ...generator };
    `,
      filename: 'file.ts',
    },
    {
      code: `
declare const promiseLike: PromiseLike<number>;

const o = { ...promiseLike };
    `,
      filename: 'file.ts',
    },
    {
      code: `
const obj = { a: 1, b: 2 };
const o = <div {...x} />;
      `,
      filename: 'react.tsx',
      languageOptions: {
        parserOptions: {
          ecmaFeatures: {
            jsx: true,
          },
        },
      },
    },
    {
      code: `
declare const obj: { a: number; b: number } | any;
const o = <div {...x} />;
      `,
      filename: 'react.tsx',
      languageOptions: {
        parserOptions: {
          ecmaFeatures: {
            jsx: true,
          },
        },
      },
    },
    {
      code: `
const promise = new Promise(() => {});
const o = { ...promise };
      `,
      filename: 'file.ts',
      options: [
        {
          allow: ['Promise'],
        },
      ],
    },
    {
      code: `
interface A {}

declare const a: A;

const o = { ...a };
    `,
      filename: 'file.ts',
    },
    {
      code: `
const o = { ...'test' };
    `,
      filename: 'file.ts',
    },
    {
      code: `
const str: string = 'test';
const a = [...str];
      `,
      filename: 'file.ts',
      options: [
        {
          allow: ['string'],
        },
      ],
    },
    {
      code: `
function f() {}

const a = { ...f };
      `,
      filename: 'file.ts',
      options: [
        {
          allow: ['f'],
        },
      ],
    },
    {
      code: `
declare const iterator: Iterable<string>;

const a = { ...iterator };
      `,
      filename: 'file.ts',
      options: [
        {
          allow: [
            {
              from: 'lib',
              name: 'Iterable',
            },
          ],
        },
      ],
    },
    {
      code: `
type BrandedString = string & { __brand: 'safe' };

declare const brandedString: BrandedString;

const spreadBrandedString = [...brandedString];
      `,
      filename: 'file.ts',
      options: [
        {
          allow: [
            {
              from: 'file',
              name: 'BrandedString',
            },
          ],
        },
      ],
    },
    {
      code: `
type CustomIterable = {
  [Symbol.iterator]: () => Generator<string>;
};

declare const iterator: CustomIterable;

const a = { ...iterator };
      `,
      filename: 'file.ts',
      options: [
        {
          allow: ['CustomIterable'],
        },
      ],
    },
    {
      code: `
type CustomIterable = {
  [Symbol.iterator]: () => string;
};

declare const iterator: CustomIterable;

const a = { ...iterator };
      `,
      filename: 'file.ts',
      options: [
        {
          allow: [
            {
              from: 'file',
              name: 'CustomIterable',
            },
          ],
        },
      ],
    },
    {
      code: `
declare module 'module' {
  export type CustomIterable = {
    [Symbol.iterator]: () => string;
  };
}

import { CustomIterable } from 'module';

declare const iterator: CustomIterable;

const a = { ...iterator };
      `,
      filename: 'file.ts',
      options: [
        {
          allow: [
            {
              from: 'package',
              name: 'CustomIterable',
              package: 'module',
            },
          ],
        },
      ],
    },
    {
      code: `
class A {
  a = 1;
}

const a = new A();

const o = { ...a };
      `,
      filename: 'file.ts',
      options: [
        {
          allow: ['A'],
        },
      ],
    },
    {
      code: `
const a = {
  ...class A {
    static value = 1;
  },
};
      `,
      filename: 'file.ts',
      options: [
        {
          allow: ['A'],
        },
      ],
    },
  ],
  invalid: [
    {
      code: "const a = [...'test'];",
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noStringSpread',
          message: `Using the spread operator on a string can mishandle special characters, as can \`.split("")\`.
- \`...\` produces Unicode code points, which will decompose complex emojis into individual emojis
- .split("") produces UTF-16 code units, which breaks rich characters in many languages
Consider using \`Intl.Segmenter\` for locale-aware string decomposition.
Otherwise, if you don't need to preserve emojis or other non-Ascii characters, disable this lint rule on this line or configure the 'allow' rule option.`,
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: `
function withText<Text extends string>(text: Text) {
  return [...text];
}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noStringSpread',
          message: `Using the spread operator on a string can mishandle special characters, as can \`.split("")\`.
- \`...\` produces Unicode code points, which will decompose complex emojis into individual emojis
- .split("") produces UTF-16 code units, which breaks rich characters in many languages
Consider using \`Intl.Segmenter\` for locale-aware string decomposition.
Otherwise, if you don't need to preserve emojis or other non-Ascii characters, disable this lint rule on this line or configure the 'allow' rule option.`,
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 18,
        },
      ],
    },
    {
      code: `
const test = 'hello';
const a = [...test];
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noStringSpread',
          message: `Using the spread operator on a string can mishandle special characters, as can \`.split("")\`.
- \`...\` produces Unicode code points, which will decompose complex emojis into individual emojis
- .split("") produces UTF-16 code units, which breaks rich characters in many languages
Consider using \`Intl.Segmenter\` for locale-aware string decomposition.
Otherwise, if you don't need to preserve emojis or other non-Ascii characters, disable this lint rule on this line or configure the 'allow' rule option.`,
          line: 3,
          column: 12,
          endLine: 3,
          endColumn: 19,
        },
      ],
    },
    {
      code: `
const test = \`he\${'ll'}o\`;
const a = [...test];
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noStringSpread',
          message: `Using the spread operator on a string can mishandle special characters, as can \`.split("")\`.
- \`...\` produces Unicode code points, which will decompose complex emojis into individual emojis
- .split("") produces UTF-16 code units, which breaks rich characters in many languages
Consider using \`Intl.Segmenter\` for locale-aware string decomposition.
Otherwise, if you don't need to preserve emojis or other non-Ascii characters, disable this lint rule on this line or configure the 'allow' rule option.`,
          line: 3,
          column: 12,
          endLine: 3,
          endColumn: 19,
        },
      ],
    },
    {
      code: `
declare const test: string;
const a = [...test];
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noStringSpread',
          message: `Using the spread operator on a string can mishandle special characters, as can \`.split("")\`.
- \`...\` produces Unicode code points, which will decompose complex emojis into individual emojis
- .split("") produces UTF-16 code units, which breaks rich characters in many languages
Consider using \`Intl.Segmenter\` for locale-aware string decomposition.
Otherwise, if you don't need to preserve emojis or other non-Ascii characters, disable this lint rule on this line or configure the 'allow' rule option.`,
          line: 3,
          column: 12,
          endLine: 3,
          endColumn: 19,
        },
      ],
    },
    {
      code: `
declare const test: string | number[];
const a = [...test];
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noStringSpread',
          message: `Using the spread operator on a string can mishandle special characters, as can \`.split("")\`.
- \`...\` produces Unicode code points, which will decompose complex emojis into individual emojis
- .split("") produces UTF-16 code units, which breaks rich characters in many languages
Consider using \`Intl.Segmenter\` for locale-aware string decomposition.
Otherwise, if you don't need to preserve emojis or other non-Ascii characters, disable this lint rule on this line or configure the 'allow' rule option.`,
          line: 3,
          column: 12,
          endLine: 3,
          endColumn: 19,
        },
      ],
    },
    {
      code: `
declare const test: string & { __brand: 'test' };
const a = [...test];
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noStringSpread',
          message: `Using the spread operator on a string can mishandle special characters, as can \`.split("")\`.
- \`...\` produces Unicode code points, which will decompose complex emojis into individual emojis
- .split("") produces UTF-16 code units, which breaks rich characters in many languages
Consider using \`Intl.Segmenter\` for locale-aware string decomposition.
Otherwise, if you don't need to preserve emojis or other non-Ascii characters, disable this lint rule on this line or configure the 'allow' rule option.`,
          line: 3,
          column: 12,
          endLine: 3,
          endColumn: 19,
        },
      ],
    },
    {
      code: `
declare const test: number | (boolean | (string & { __brand: true }));
const a = [...test];
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noStringSpread',
          message: `Using the spread operator on a string can mishandle special characters, as can \`.split("")\`.
- \`...\` produces Unicode code points, which will decompose complex emojis into individual emojis
- .split("") produces UTF-16 code units, which breaks rich characters in many languages
Consider using \`Intl.Segmenter\` for locale-aware string decomposition.
Otherwise, if you don't need to preserve emojis or other non-Ascii characters, disable this lint rule on this line or configure the 'allow' rule option.`,
          line: 3,
          column: 12,
          endLine: 3,
          endColumn: 19,
        },
      ],
    },
    {
      code: `
declare function getString(): string;
const a = [...getString()];
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noStringSpread',
          message: `Using the spread operator on a string can mishandle special characters, as can \`.split("")\`.
- \`...\` produces Unicode code points, which will decompose complex emojis into individual emojis
- .split("") produces UTF-16 code units, which breaks rich characters in many languages
Consider using \`Intl.Segmenter\` for locale-aware string decomposition.
Otherwise, if you don't need to preserve emojis or other non-Ascii characters, disable this lint rule on this line or configure the 'allow' rule option.`,
          line: 3,
          column: 12,
          endLine: 3,
          endColumn: 26,
        },
      ],
    },
    {
      code: `
declare function textIdentity(...args: string[]);

declare const text: string;

textIdentity(...text);
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noStringSpread',
          message: `Using the spread operator on a string can mishandle special characters, as can \`.split("")\`.
- \`...\` produces Unicode code points, which will decompose complex emojis into individual emojis
- .split("") produces UTF-16 code units, which breaks rich characters in many languages
Consider using \`Intl.Segmenter\` for locale-aware string decomposition.
Otherwise, if you don't need to preserve emojis or other non-Ascii characters, disable this lint rule on this line or configure the 'allow' rule option.`,
          line: 6,
          column: 14,
          endLine: 6,
          endColumn: 21,
        },
      ],
    },
    {
      code: `
declare function textIdentity(...args: string[]);

declare const text: string;

textIdentity(...text, 'and', ...text);
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noStringSpread',
          message: `Using the spread operator on a string can mishandle special characters, as can \`.split("")\`.
- \`...\` produces Unicode code points, which will decompose complex emojis into individual emojis
- .split("") produces UTF-16 code units, which breaks rich characters in many languages
Consider using \`Intl.Segmenter\` for locale-aware string decomposition.
Otherwise, if you don't need to preserve emojis or other non-Ascii characters, disable this lint rule on this line or configure the 'allow' rule option.`,
          line: 6,
          column: 14,
          endLine: 6,
          endColumn: 21,
        },
        {
          messageId: 'noStringSpread',
          message: `Using the spread operator on a string can mishandle special characters, as can \`.split("")\`.
- \`...\` produces Unicode code points, which will decompose complex emojis into individual emojis
- .split("") produces UTF-16 code units, which breaks rich characters in many languages
Consider using \`Intl.Segmenter\` for locale-aware string decomposition.
Otherwise, if you don't need to preserve emojis or other non-Ascii characters, disable this lint rule on this line or configure the 'allow' rule option.`,
          line: 6,
          column: 30,
          endLine: 6,
          endColumn: 37,
        },
      ],
    },
    {
      code: `
declare function textIdentity(...args: string[]);

function withText<Text extends string>(text: Text) {
  textIdentity(...text);
}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noStringSpread',
          message: `Using the spread operator on a string can mishandle special characters, as can \`.split("")\`.
- \`...\` produces Unicode code points, which will decompose complex emojis into individual emojis
- .split("") produces UTF-16 code units, which breaks rich characters in many languages
Consider using \`Intl.Segmenter\` for locale-aware string decomposition.
Otherwise, if you don't need to preserve emojis or other non-Ascii characters, disable this lint rule on this line or configure the 'allow' rule option.`,
          line: 5,
          column: 16,
          endLine: 5,
          endColumn: 23,
        },
      ],
    },
    {
      code: `
declare function getString<T extends string>(): T;
const a = [...getString()];
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noStringSpread',
          message: `Using the spread operator on a string can mishandle special characters, as can \`.split("")\`.
- \`...\` produces Unicode code points, which will decompose complex emojis into individual emojis
- .split("") produces UTF-16 code units, which breaks rich characters in many languages
Consider using \`Intl.Segmenter\` for locale-aware string decomposition.
Otherwise, if you don't need to preserve emojis or other non-Ascii characters, disable this lint rule on this line or configure the 'allow' rule option.`,
          line: 3,
          column: 12,
          endLine: 3,
          endColumn: 26,
        },
      ],
    },
    {
      code: `
declare function getString(): string & { __brand: 'test' };
const a = [...getString()];
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noStringSpread',
          message: `Using the spread operator on a string can mishandle special characters, as can \`.split("")\`.
- \`...\` produces Unicode code points, which will decompose complex emojis into individual emojis
- .split("") produces UTF-16 code units, which breaks rich characters in many languages
Consider using \`Intl.Segmenter\` for locale-aware string decomposition.
Otherwise, if you don't need to preserve emojis or other non-Ascii characters, disable this lint rule on this line or configure the 'allow' rule option.`,
          line: 3,
          column: 12,
          endLine: 3,
          endColumn: 26,
        },
      ],
    },
    {
      code: 'const o = { ...[1, 2, 3] };',
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noArraySpreadInObject',
          message:
            'Using the spread operator on an array in an object will result in a list of indices.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: `
const arr = [1, 2, 3];
const o = { ...arr };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noArraySpreadInObject',
          message:
            'Using the spread operator on an array in an object will result in a list of indices.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 19,
        },
      ],
    },
    {
      code: `
const arr = [1, 2, 3] as const;
const o = { ...arr };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noArraySpreadInObject',
          message:
            'Using the spread operator on an array in an object will result in a list of indices.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 19,
        },
      ],
    },
    {
      code: `
declare const arr: number[];
const o = { ...arr };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noArraySpreadInObject',
          message:
            'Using the spread operator on an array in an object will result in a list of indices.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 19,
        },
      ],
    },
    {
      code: `
declare const arr: readonly number[];
const o = { ...arr };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noArraySpreadInObject',
          message:
            'Using the spread operator on an array in an object will result in a list of indices.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 19,
        },
      ],
    },
    {
      code: `
declare const arr: number[] | string[];
const o = { ...arr };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noArraySpreadInObject',
          message:
            'Using the spread operator on an array in an object will result in a list of indices.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 19,
        },
      ],
    },
    {
      code: `
declare const arr: number[] & string[];
const o = { ...arr };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noArraySpreadInObject',
          message:
            'Using the spread operator on an array in an object will result in a list of indices.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 19,
        },
      ],
    },
    {
      code: `
declare function getArray(): number[];
const o = { ...getArray() };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noArraySpreadInObject',
          message:
            'Using the spread operator on an array in an object will result in a list of indices.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 26,
        },
      ],
    },
    {
      code: `
declare function getArray(): readonly number[];
const o = { ...getArray() };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noArraySpreadInObject',
          message:
            'Using the spread operator on an array in an object will result in a list of indices.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 26,
        },
      ],
    },
    {
      code: 'const o = { ...new Set([1, 2, 3]) };',
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noIterableSpreadInObject',
          message:
            'Using the spread operator on an Iterable in an object can cause unexpected behavior.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: `
const set = new Set([1, 2, 3]);
const o = { ...set };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noIterableSpreadInObject',
          message:
            'Using the spread operator on an Iterable in an object can cause unexpected behavior.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 19,
        },
      ],
    },
    {
      code: `
declare const set: Set<number>;
const o = { ...set };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noIterableSpreadInObject',
          message:
            'Using the spread operator on an Iterable in an object can cause unexpected behavior.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 19,
        },
      ],
    },
    {
      code: `
declare const set: WeakSet<object>;
const o = { ...set };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noClassInstanceSpreadInObject',
          message:
            'Using the spread operator on class instances will lose their class prototype.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 19,
        },
      ],
    },
    {
      code: `
declare const set: ReadonlySet<number>;
const o = { ...set };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noIterableSpreadInObject',
          message:
            'Using the spread operator on an Iterable in an object can cause unexpected behavior.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 19,
        },
      ],
    },
    {
      code: `
declare const set: Set<number> | { a: number };
const o = { ...set };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noIterableSpreadInObject',
          message:
            'Using the spread operator on an Iterable in an object can cause unexpected behavior.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 19,
        },
      ],
    },
    {
      code: `
declare function getSet(): Set<number>;
const o = { ...getSet() };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noIterableSpreadInObject',
          message:
            'Using the spread operator on an Iterable in an object can cause unexpected behavior.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 24,
        },
      ],
    },
    {
      code: `
const o = {
  ...new Map([
    ['test-1', 1],
    ['test-2', 2],
  ]),
};
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noMapSpreadInObject',
          message:
            'Using the spread operator on a Map in an object will result in an empty object. Did you mean to use `Object.fromEntries(map)` instead?',
          line: 3,
          column: 3,
          endLine: 6,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'replaceMapSpreadInObject',
              desc: 'Replace map spread in object with `Object.fromEntries()`',
              output: `
const o = Object.fromEntries(new Map([
    ['test-1', 1],
    ['test-2', 2],
  ]));
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
const map = new Map([
  ['test-1', 1],
  ['test-2', 2],
]);

const o = { ...map };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noMapSpreadInObject',
          message:
            'Using the spread operator on a Map in an object will result in an empty object. Did you mean to use `Object.fromEntries(map)` instead?',
          line: 7,
          column: 13,
          endLine: 7,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'replaceMapSpreadInObject',
              desc: 'Replace map spread in object with `Object.fromEntries()`',
              output: `
const map = new Map([
  ['test-1', 1],
  ['test-2', 2],
]);

const o = Object.fromEntries(map);
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare const map: Map<string, number>;
const o = { ...map };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noMapSpreadInObject',
          message:
            'Using the spread operator on a Map in an object will result in an empty object. Did you mean to use `Object.fromEntries(map)` instead?',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'replaceMapSpreadInObject',
              desc: 'Replace map spread in object with `Object.fromEntries()`',
              output: `
declare const map: Map<string, number>;
const o = Object.fromEntries(map);
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
        declare const map: Map<string, number>;
        const o = { ...(map) };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noMapSpreadInObject',
          message:
            'Using the spread operator on a Map in an object will result in an empty object. Did you mean to use `Object.fromEntries(map)` instead?',
          line: 3,
          column: 21,
          endLine: 3,
          endColumn: 29,
          suggestions: [
            {
              messageId: 'replaceMapSpreadInObject',
              desc: 'Replace map spread in object with `Object.fromEntries()`',
              output: `
        declare const map: Map<string, number>;
        const o = Object.fromEntries(map);
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare const map: Map<string, number>;
const o = { ...(map, map) };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noMapSpreadInObject',
          message:
            'Using the spread operator on a Map in an object will result in an empty object. Did you mean to use `Object.fromEntries(map)` instead?',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 26,
          suggestions: [
            {
              messageId: 'replaceMapSpreadInObject',
              desc: 'Replace map spread in object with `Object.fromEntries()`',
              output: `
declare const map: Map<string, number>;
const o = Object.fromEntries((map, map));
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare const map: Map<string, number>;
const others = { a: 1 };
const o = { ...map, ...others };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noMapSpreadInObject',
          message:
            'Using the spread operator on a Map in an object will result in an empty object. Did you mean to use `Object.fromEntries(map)` instead?',
          line: 4,
          column: 13,
          endLine: 4,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'replaceMapSpreadInObject',
              desc: 'Replace map spread in object with `Object.fromEntries()`',
              output: `
declare const map: Map<string, number>;
const others = { a: 1 };
const o = { ...Object.fromEntries(map), ...others };
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare const map: Map<string, number>;
const o = { other: 1, ...map };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noMapSpreadInObject',
          message:
            'Using the spread operator on a Map in an object will result in an empty object. Did you mean to use `Object.fromEntries(map)` instead?',
          line: 3,
          column: 23,
          endLine: 3,
          endColumn: 29,
          suggestions: [
            {
              messageId: 'replaceMapSpreadInObject',
              desc: 'Replace map spread in object with `Object.fromEntries()`',
              output: `
declare const map: Map<string, number>;
const o = { other: 1, ...Object.fromEntries(map) };
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare const map: ReadonlyMap<string, number>;
const o = { ...map };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noMapSpreadInObject',
          message:
            'Using the spread operator on a Map in an object will result in an empty object. Did you mean to use `Object.fromEntries(map)` instead?',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'replaceMapSpreadInObject',
              desc: 'Replace map spread in object with `Object.fromEntries()`',
              output: `
declare const map: ReadonlyMap<string, number>;
const o = Object.fromEntries(map);
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare const map: WeakMap<{ a: number }, string>;
const o = { ...map };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noMapSpreadInObject',
          message:
            'Using the spread operator on a Map in an object will result in an empty object. Did you mean to use `Object.fromEntries(map)` instead?',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'replaceMapSpreadInObject',
              desc: 'Replace map spread in object with `Object.fromEntries()`',
              output: `
declare const map: WeakMap<{ a: number }, string>;
const o = Object.fromEntries(map);
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare const map: Map<string, number> | { a: number };
const o = { ...map };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noMapSpreadInObject',
          message:
            'Using the spread operator on a Map in an object will result in an empty object. Did you mean to use `Object.fromEntries(map)` instead?',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 19,
        },
      ],
    },
    {
      code: `
declare function getMap(): Map<string, number>;
const o = { ...getMap() };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noMapSpreadInObject',
          message:
            'Using the spread operator on a Map in an object will result in an empty object. Did you mean to use `Object.fromEntries(map)` instead?',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 24,
          suggestions: [
            {
              messageId: 'replaceMapSpreadInObject',
              desc: 'Replace map spread in object with `Object.fromEntries()`',
              output: `
declare function getMap(): Map<string, number>;
const o = Object.fromEntries(getMap());
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare const a: Map<boolean, string> & Set<number>;
const o = { ...a };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noMapSpreadInObject',
          message:
            'Using the spread operator on a Map in an object will result in an empty object. Did you mean to use `Object.fromEntries(map)` instead?',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 17,
          suggestions: [
            {
              messageId: 'replaceMapSpreadInObject',
              desc: 'Replace map spread in object with `Object.fromEntries()`',
              output: `
declare const a: Map<boolean, string> & Set<number>;
const o = Object.fromEntries(a);
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
const ref = new WeakRef({ a: 1 });
const o = { ...ref };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noClassInstanceSpreadInObject',
          message:
            'Using the spread operator on class instances will lose their class prototype.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 19,
        },
      ],
    },
    {
      code: `
const promise = new Promise(() => {});
const o = { ...promise };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noPromiseSpreadInObject',
          message:
            'Using the spread operator on Promise in an object can cause unexpected behavior. Did you forget to await the promise?',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 23,
          suggestions: [
            {
              messageId: 'addAwait',
              desc: 'Add await operator.',
              output: `
const promise = new Promise(() => {});
const o = { ...await promise };
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare const promise: Promise<{ a: 1 }>;
async function foo() {
  return { ...(promise || {}) };
}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noPromiseSpreadInObject',
          message:
            'Using the spread operator on Promise in an object can cause unexpected behavior. Did you forget to await the promise?',
          line: 4,
          column: 12,
          endLine: 4,
          endColumn: 30,
          suggestions: [
            {
              messageId: 'addAwait',
              desc: 'Add await operator.',
              output: `
declare const promise: Promise<{ a: 1 }>;
async function foo() {
  return { ...(await (promise || {})) };
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare const promise: Promise<any>;
async function foo() {
  return { ...(Math.random() < 0.5 ? promise : {}) };
}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noPromiseSpreadInObject',
          message:
            'Using the spread operator on Promise in an object can cause unexpected behavior. Did you forget to await the promise?',
          line: 4,
          column: 12,
          endLine: 4,
          endColumn: 51,
          suggestions: [
            {
              messageId: 'addAwait',
              desc: 'Add await operator.',
              output: `
declare const promise: Promise<any>;
async function foo() {
  return { ...(await (Math.random() < 0.5 ? promise : {})) };
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
function withPromise<P extends Promise<void>>(promise: P) {
  return { ...promise };
}
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noPromiseSpreadInObject',
          message:
            'Using the spread operator on Promise in an object can cause unexpected behavior. Did you forget to await the promise?',
          line: 3,
          column: 12,
          endLine: 3,
          endColumn: 22,
          suggestions: [
            {
              messageId: 'addAwait',
              desc: 'Add await operator.',
              output: `
function withPromise<P extends Promise<void>>(promise: P) {
  return { ...await promise };
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare const maybePromise: Promise<number> | { a: number };
const o = { ...maybePromise };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noPromiseSpreadInObject',
          message:
            'Using the spread operator on Promise in an object can cause unexpected behavior. Did you forget to await the promise?',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 28,
          suggestions: [
            {
              messageId: 'addAwait',
              desc: 'Add await operator.',
              output: `
declare const maybePromise: Promise<number> | { a: number };
const o = { ...await maybePromise };
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare const promise: Promise<number> & { a: number };
const o = { ...promise };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noPromiseSpreadInObject',
          message:
            'Using the spread operator on Promise in an object can cause unexpected behavior. Did you forget to await the promise?',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 23,
          suggestions: [
            {
              messageId: 'addAwait',
              desc: 'Add await operator.',
              output: `
declare const promise: Promise<number> & { a: number };
const o = { ...await promise };
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare function getPromise(): Promise<number>;
const o = { ...getPromise() };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noPromiseSpreadInObject',
          message:
            'Using the spread operator on Promise in an object can cause unexpected behavior. Did you forget to await the promise?',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 28,
          suggestions: [
            {
              messageId: 'addAwait',
              desc: 'Add await operator.',
              output: `
declare function getPromise(): Promise<number>;
const o = { ...await getPromise() };
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
declare function getPromise<T extends Promise<number>>(arg: T): T;
const o = { ...getPromise() };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noPromiseSpreadInObject',
          message:
            'Using the spread operator on Promise in an object can cause unexpected behavior. Did you forget to await the promise?',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 28,
          suggestions: [
            {
              messageId: 'addAwait',
              desc: 'Add await operator.',
              output: `
declare function getPromise<T extends Promise<number>>(arg: T): T;
const o = { ...await getPromise() };
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
function f() {}

const o = { ...f };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noFunctionSpreadInObject',
          message:
            'Using the spread operator on a function without additional properties can cause unexpected behavior. Did you forget to call the function?',
          line: 4,
          column: 13,
          endLine: 4,
          endColumn: 17,
        },
      ],
    },
    {
      code: `
interface FunctionWithProps {
  (): string;
  prop: boolean;
}

type FunctionWithoutProps = () => string;

declare const obj: FunctionWithProps | FunctionWithoutProps | object;

const o = { ...obj };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noFunctionSpreadInObject',
          message:
            'Using the spread operator on a function without additional properties can cause unexpected behavior. Did you forget to call the function?',
          line: 11,
          column: 13,
          endLine: 11,
          endColumn: 19,
        },
      ],
    },
    {
      code: `
const f = () => {};

const o = { ...f };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noFunctionSpreadInObject',
          message:
            'Using the spread operator on a function without additional properties can cause unexpected behavior. Did you forget to call the function?',
          line: 4,
          column: 13,
          endLine: 4,
          endColumn: 17,
        },
      ],
    },
    {
      code: `
declare function f(): void;

const o = { ...f };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noFunctionSpreadInObject',
          message:
            'Using the spread operator on a function without additional properties can cause unexpected behavior. Did you forget to call the function?',
          line: 4,
          column: 13,
          endLine: 4,
          endColumn: 17,
        },
      ],
    },
    {
      code: `
declare function getFunction(): () => void;

const o = { ...getFunction() };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noFunctionSpreadInObject',
          message:
            'Using the spread operator on a function without additional properties can cause unexpected behavior. Did you forget to call the function?',
          line: 4,
          column: 13,
          endLine: 4,
          endColumn: 29,
        },
      ],
    },
    {
      code: `
declare const f: () => void;

const o = { ...f };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noFunctionSpreadInObject',
          message:
            'Using the spread operator on a function without additional properties can cause unexpected behavior. Did you forget to call the function?',
          line: 4,
          column: 13,
          endLine: 4,
          endColumn: 17,
        },
      ],
    },
    {
      code: `
declare const f: () => void | { a: number };

const o = { ...f };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noFunctionSpreadInObject',
          message:
            'Using the spread operator on a function without additional properties can cause unexpected behavior. Did you forget to call the function?',
          line: 4,
          column: 13,
          endLine: 4,
          endColumn: 17,
        },
      ],
    },
    {
      code: `
function* generator() {}

const o = { ...generator };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noFunctionSpreadInObject',
          message:
            'Using the spread operator on a function without additional properties can cause unexpected behavior. Did you forget to call the function?',
          line: 4,
          column: 13,
          endLine: 4,
          endColumn: 25,
        },
      ],
    },
    {
      code: `
const iterator = {
  *[Symbol.iterator]() {
    yield 'test';
  },
};

const o = { ...iterator };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noIterableSpreadInObject',
          message:
            'Using the spread operator on an Iterable in an object can cause unexpected behavior.',
          line: 8,
          column: 13,
          endLine: 8,
          endColumn: 24,
        },
      ],
    },
    {
      code: `
type CustomIterable = {
  [Symbol.iterator]: () => Generator<string>;
};

const iterator: CustomIterable = {
  *[Symbol.iterator]() {
    yield 'test';
  },
};

const a = { ...iterator };
      `,
      filename: 'file.ts',
      options: [
        {
          allow: ['AnotherIterable'],
        },
      ],
      errors: [
        {
          messageId: 'noIterableSpreadInObject',
          message:
            'Using the spread operator on an Iterable in an object can cause unexpected behavior.',
          line: 12,
          column: 13,
          endLine: 12,
          endColumn: 24,
        },
      ],
    },
    {
      code: `
declare module 'module' {
  export type CustomIterable = {
    [Symbol.iterator]: () => string;
  };
}

import { CustomIterable } from 'module';

declare const iterator: CustomIterable;

const a = { ...iterator };
      `,
      filename: 'file.ts',
      options: [
        {
          allow: [
            {
              from: 'package',
              name: 'Nothing',
              package: 'module',
            },
          ],
        },
      ],
      errors: [
        {
          messageId: 'noIterableSpreadInObject',
          message:
            'Using the spread operator on an Iterable in an object can cause unexpected behavior.',
          line: 12,
          column: 13,
          endLine: 12,
          endColumn: 24,
        },
      ],
    },
    {
      code: `
declare const iterator: Iterable<string>;

const o = { ...iterator };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noIterableSpreadInObject',
          message:
            'Using the spread operator on an Iterable in an object can cause unexpected behavior.',
          line: 4,
          column: 13,
          endLine: 4,
          endColumn: 24,
        },
      ],
    },
    {
      code: `
declare const iterator: Iterable<string> | { a: number };

const o = { ...iterator };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noIterableSpreadInObject',
          message:
            'Using the spread operator on an Iterable in an object can cause unexpected behavior.',
          line: 4,
          column: 13,
          endLine: 4,
          endColumn: 24,
        },
      ],
    },
    {
      code: `
declare function getIterable(): Iterable<string>;

const o = { ...getIterable() };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noIterableSpreadInObject',
          message:
            'Using the spread operator on an Iterable in an object can cause unexpected behavior.',
          line: 4,
          column: 13,
          endLine: 4,
          endColumn: 29,
        },
      ],
    },
    {
      code: `
class A {
  [Symbol.iterator]() {
    return {
      next() {
        return { done: true, value: undefined };
      },
    };
  }
}

const a = { ...new A() };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noIterableSpreadInObject',
          message:
            'Using the spread operator on an Iterable in an object can cause unexpected behavior.',
          line: 12,
          column: 13,
          endLine: 12,
          endColumn: 23,
        },
      ],
    },
    {
      code: `
const o = { ...new Date() };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noClassInstanceSpreadInObject',
          message:
            'Using the spread operator on class instances will lose their class prototype.',
          line: 2,
          column: 13,
          endLine: 2,
          endColumn: 26,
        },
      ],
    },
    {
      code: `
declare class HTMLElementLike {}
declare const element: HTMLElementLike;
const o = { ...element };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noClassInstanceSpreadInObject',
          message:
            'Using the spread operator on class instances will lose their class prototype.',
          line: 4,
          column: 13,
          endLine: 4,
          endColumn: 23,
        },
      ],
    },
    {
      code: `
declare const regex: RegExp;
const o = { ...regex };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noClassInstanceSpreadInObject',
          message:
            'Using the spread operator on class instances will lose their class prototype.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 21,
        },
      ],
    },
    {
      code: `
class A {
  a = 1;
  public b = 2;
  private c = 3;
  protected d = 4;
  static e = 5;
}

const o = { ...new A() };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noClassInstanceSpreadInObject',
          message:
            'Using the spread operator on class instances will lose their class prototype.',
          line: 10,
          column: 13,
          endLine: 10,
          endColumn: 23,
        },
      ],
    },
    {
      code: `
class A {
  a = 1;
}

const a = new A();

const o = { ...a };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noClassInstanceSpreadInObject',
          message:
            'Using the spread operator on class instances will lose their class prototype.',
          line: 8,
          column: 13,
          endLine: 8,
          endColumn: 17,
        },
      ],
    },
    {
      code: `
class A {
  a = 1;
}

declare const a: A;

const o = { ...a };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noClassInstanceSpreadInObject',
          message:
            'Using the spread operator on class instances will lose their class prototype.',
          line: 8,
          column: 13,
          endLine: 8,
          endColumn: 17,
        },
      ],
    },
    {
      code: `
class A {
  a = 1;
}

declare function getA(): A;

const o = { ...getA() };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noClassInstanceSpreadInObject',
          message:
            'Using the spread operator on class instances will lose their class prototype.',
          line: 8,
          column: 13,
          endLine: 8,
          endColumn: 22,
        },
      ],
    },
    {
      code: `
class A {
  a = 1;
}

declare function getA<T extends A>(arg: T): T;

const o = { ...getA() };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noClassInstanceSpreadInObject',
          message:
            'Using the spread operator on class instances will lose their class prototype.',
          line: 8,
          column: 13,
          endLine: 8,
          endColumn: 22,
        },
      ],
    },
    {
      code: `
class A {
  a = 1;
}

class B extends A {}

const o = { ...new B() };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noClassInstanceSpreadInObject',
          message:
            'Using the spread operator on class instances will lose their class prototype.',
          line: 8,
          column: 13,
          endLine: 8,
          endColumn: 23,
        },
      ],
    },
    {
      code: `
class A {
  a = 1;
}

declare const a: A | { b: string };

const o = { ...a };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noClassInstanceSpreadInObject',
          message:
            'Using the spread operator on class instances will lose their class prototype.',
          line: 8,
          column: 13,
          endLine: 8,
          endColumn: 17,
        },
      ],
    },
    {
      code: `
class A {
  a = 1;
}

declare const a: A & { b: string };

const o = { ...a };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noClassInstanceSpreadInObject',
          message:
            'Using the spread operator on class instances will lose their class prototype.',
          line: 8,
          column: 13,
          endLine: 8,
          endColumn: 17,
        },
      ],
    },
    {
      code: `
class A {}

const o = { ...A };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noClassDeclarationSpreadInObject',
          message:
            'Using the spread operator on class declarations will spread only their static properties, and will lose their class prototype.',
          line: 4,
          column: 13,
          endLine: 4,
          endColumn: 17,
        },
      ],
    },
    {
      code: `
const A = class {};

const o = { ...A };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noClassDeclarationSpreadInObject',
          message:
            'Using the spread operator on class declarations will spread only their static properties, and will lose their class prototype.',
          line: 4,
          column: 13,
          endLine: 4,
          endColumn: 17,
        },
      ],
    },
    {
      code: `
class Declaration {
  declaration?: boolean;
}
const Expression = class {
  expression?: boolean;
};

declare const either: typeof Declaration | typeof Expression;

const o = { ...either };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noClassDeclarationSpreadInObject',
          message:
            'Using the spread operator on class declarations will spread only their static properties, and will lose their class prototype.',
          line: 11,
          column: 13,
          endLine: 11,
          endColumn: 22,
        },
      ],
    },
    {
      code: `
const A = Set<number>;

const o = { ...A };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noClassDeclarationSpreadInObject',
          message:
            'Using the spread operator on class declarations will spread only their static properties, and will lose their class prototype.',
          line: 4,
          column: 13,
          endLine: 4,
          endColumn: 17,
        },
      ],
    },
    {
      code: `
const a = {
  ...class A {
    static value = 1;
    nonStatic = 2;
  },
};
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noClassDeclarationSpreadInObject',
          message:
            'Using the spread operator on class declarations will spread only their static properties, and will lose their class prototype.',
          line: 3,
          column: 3,
          endLine: 6,
          endColumn: 4,
        },
      ],
    },
    {
      code: `
        const a = { ...(class A { static value = 1 }) }
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noClassDeclarationSpreadInObject',
          message:
            'Using the spread operator on class declarations will spread only their static properties, and will lose their class prototype.',
          line: 2,
          column: 21,
          endLine: 2,
          endColumn: 54,
        },
      ],
    },
    {
      code: `
        const a = { ...new (class A { static value = 1; })() };
      `,
      filename: 'file.ts',
      errors: [
        {
          messageId: 'noClassInstanceSpreadInObject',
          message:
            'Using the spread operator on class instances will lose their class prototype.',
          line: 2,
          column: 21,
          endLine: 2,
          endColumn: 61,
        },
      ],
    },
    {
      code: `
const o = <div {...[1, 2, 3]} />;
      `,
      filename: 'react.tsx',
      languageOptions: {
        parserOptions: {
          ecmaFeatures: {
            jsx: true,
          },
        },
      },
      errors: [
        {
          messageId: 'noArraySpreadInObject',
          message:
            'Using the spread operator on an array in an object will result in a list of indices.',
          line: 2,
          column: 16,
          endLine: 2,
          endColumn: 30,
        },
      ],
    },
    {
      code: `
class A {}

const o = <div {...A} />;
      `,
      filename: 'react.tsx',
      languageOptions: {
        parserOptions: {
          ecmaFeatures: {
            jsx: true,
          },
        },
      },
      errors: [
        {
          messageId: 'noClassDeclarationSpreadInObject',
          message:
            'Using the spread operator on class declarations will spread only their static properties, and will lose their class prototype.',
          line: 4,
          column: 16,
          endLine: 4,
          endColumn: 22,
        },
      ],
    },
    {
      code: `
const o = <div {...new Date()} />;
      `,
      filename: 'react.tsx',
      languageOptions: {
        parserOptions: {
          ecmaFeatures: {
            jsx: true,
          },
        },
      },
      errors: [
        {
          messageId: 'noClassInstanceSpreadInObject',
          message:
            'Using the spread operator on class instances will lose their class prototype.',
          line: 2,
          column: 16,
          endLine: 2,
          endColumn: 31,
        },
      ],
    },
    {
      code: `
function f() {}

const o = <div {...f} />;
      `,
      filename: 'react.tsx',
      languageOptions: {
        parserOptions: {
          ecmaFeatures: {
            jsx: true,
          },
        },
      },
      errors: [
        {
          messageId: 'noFunctionSpreadInObject',
          message:
            'Using the spread operator on a function without additional properties can cause unexpected behavior. Did you forget to call the function?',
          line: 4,
          column: 16,
          endLine: 4,
          endColumn: 22,
        },
      ],
    },
    {
      code: `
const o = <div {...new Set([1, 2, 3])} />;
      `,
      filename: 'react.tsx',
      languageOptions: {
        parserOptions: {
          ecmaFeatures: {
            jsx: true,
          },
        },
      },
      errors: [
        {
          messageId: 'noIterableSpreadInObject',
          message:
            'Using the spread operator on an Iterable in an object can cause unexpected behavior.',
          line: 2,
          column: 16,
          endLine: 2,
          endColumn: 39,
        },
      ],
    },
    {
      code: `
declare const map: Map<string, number>;

const o = <div {...map} />;
      `,
      filename: 'react.tsx',
      languageOptions: {
        parserOptions: {
          ecmaFeatures: {
            jsx: true,
          },
        },
      },
      errors: [
        {
          messageId: 'noMapSpreadInObject',
          message:
            'Using the spread operator on a Map in an object will result in an empty object. Did you mean to use `Object.fromEntries(map)` instead?',
          line: 4,
          column: 16,
          endLine: 4,
          endColumn: 24,
          suggestions: [
            {
              messageId: 'replaceMapSpreadInObject',
              desc: 'Replace map spread in object with `Object.fromEntries()`',
              output: `
declare const map: Map<string, number>;

const o = <div {...Object.fromEntries(map)} />;
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
const promise = new Promise(() => {});

const o = <div {...promise} />;
      `,
      filename: 'react.tsx',
      languageOptions: {
        parserOptions: {
          ecmaFeatures: {
            jsx: true,
          },
        },
      },
      errors: [
        {
          messageId: 'noPromiseSpreadInObject',
          message:
            'Using the spread operator on Promise in an object can cause unexpected behavior. Did you forget to await the promise?',
          line: 4,
          column: 16,
          endLine: 4,
          endColumn: 28,
          suggestions: [
            {
              messageId: 'addAwait',
              desc: 'Add await operator.',
              output: `
const promise = new Promise(() => {});

const o = <div {...await promise} />;
      `,
            },
          ],
        },
      ],
    },
  ],
});

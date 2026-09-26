// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/test/prefer-array-last-methods.js
import path from 'node:path';

import { lint } from '@rslint/core/internal';

import { RuleTester } from '../rule-tester';
import { buildConfigForSettings } from '../../src/util/load-test-config';

const ruleTester = new RuleTester();
const filename = 'src/virtual.js';

const valid = [
  'array.findLast(logic);',
  'array.findLastIndex(logic);',
  'array.lastIndexOf(logic);',
  'array.reduceRight(logic);',
  'array.reverse();',
  'array.toReversed();',
  'array.reverse(logic).find(logic);',
  'array.toReversed(logic).find(logic);',
  'array.reverse?.().find(logic);',
  'array.toReversed?.().find(logic);',
  'array?.reverse().find(logic);',
  'array?.toReversed().find(logic);',
  'array[reverse]().find(logic);',
  'array.reverse()[find](logic);',
  'array.reverse().map(logic);',
  'array.toReversed().map(logic);',
  'const reversed = array.reverse(); reversed.find(logic);',
  'array.reverse().find;',
  'array.toReversed().find;',
].map((code) => ({ code, filename }));

const invalid = [
  {
    code: 'array.reverse().find(logic);',
    message: 'Prefer `Array#findLast()` over `Array#reverse().find()`.',
    suggestions: [
      {
        messageId: 'replace',
        desc: 'Replace `.reverse().find()` with `.findLast()`.',
        output: 'array.findLast(logic);',
      },
    ],
  },
  {
    code: 'array.reverse().findIndex(logic);',
    message:
      'Prefer `Array#findLastIndex()` over `Array#reverse().findIndex()`.',
    suggestions: [
      {
        messageId: 'replace',
        desc: 'Replace `.reverse().findIndex()` with `.findLastIndex()`.',
        output: 'array.findLastIndex(logic);',
      },
    ],
  },
  {
    code: 'array.reverse().indexOf(logic);',
    message: 'Prefer `Array#lastIndexOf()` over `Array#reverse().indexOf()`.',
    suggestions: [
      {
        messageId: 'replace',
        desc: 'Replace `.reverse().indexOf()` with `.lastIndexOf()`.',
        output: 'array.lastIndexOf(logic);',
      },
    ],
  },
  {
    code: 'array.reverse().reduce(logic);',
    message: 'Prefer `Array#reduceRight()` over `Array#reverse().reduce()`.',
    suggestions: [
      {
        messageId: 'replace',
        desc: 'Replace `.reverse().reduce()` with `.reduceRight()`.',
        output: 'array.reduceRight(logic);',
      },
    ],
  },
  {
    code: 'array.toReversed().find(logic);',
    message: 'Prefer `Array#findLast()` over `Array#toReversed().find()`.',
    suggestions: [
      {
        messageId: 'replace',
        desc: 'Replace `.toReversed().find()` with `.findLast()`.',
        output: 'array.findLast(logic);',
      },
    ],
  },
  {
    code: 'array.toReversed().findIndex(logic);',
    message:
      'Prefer `Array#findLastIndex()` over `Array#toReversed().findIndex()`.',
    suggestions: [
      {
        messageId: 'replace',
        desc: 'Replace `.toReversed().findIndex()` with `.findLastIndex()`.',
        output: 'array.findLastIndex(logic);',
      },
    ],
  },
  {
    code: 'array.toReversed().indexOf(logic);',
    message:
      'Prefer `Array#lastIndexOf()` over `Array#toReversed().indexOf()`.',
    suggestions: [
      {
        messageId: 'replace',
        desc: 'Replace `.toReversed().indexOf()` with `.lastIndexOf()`.',
        output: 'array.lastIndexOf(logic);',
      },
    ],
  },
  {
    code: 'array.toReversed().reduce(logic);',
    message: 'Prefer `Array#reduceRight()` over `Array#toReversed().reduce()`.',
    suggestions: [
      {
        messageId: 'replace',
        desc: 'Replace `.toReversed().reduce()` with `.reduceRight()`.',
        output: 'array.reduceRight(logic);',
      },
    ],
  },
  {
    code: 'array.reverse().find(logic, thisArgument);',
    message: 'Prefer `Array#findLast()` over `Array#reverse().find()`.',
    suggestions: [
      {
        messageId: 'replace',
        desc: 'Replace `.reverse().find()` with `.findLast()`.',
        output: 'array.findLast(logic, thisArgument);',
      },
    ],
  },
  {
    code: 'array.toReversed().findIndex(logic, thisArgument);',
    message:
      'Prefer `Array#findLastIndex()` over `Array#toReversed().findIndex()`.',
    suggestions: [
      {
        messageId: 'replace',
        desc: 'Replace `.toReversed().findIndex()` with `.findLastIndex()`.',
        output: 'array.findLastIndex(logic, thisArgument);',
      },
    ],
  },
  {
    code: 'array.reverse().reduce(logic, initialValue);',
    message: 'Prefer `Array#reduceRight()` over `Array#reverse().reduce()`.',
    suggestions: [
      {
        messageId: 'replace',
        desc: 'Replace `.reverse().reduce()` with `.reduceRight()`.',
        output: 'array.reduceRight(logic, initialValue);',
      },
    ],
  },
  {
    code: 'array.toReversed().reduce(logic, initialValue);',
    message: 'Prefer `Array#reduceRight()` over `Array#toReversed().reduce()`.',
    suggestions: [
      {
        messageId: 'replace',
        desc: 'Replace `.toReversed().reduce()` with `.reduceRight()`.',
        output: 'array.reduceRight(logic, initialValue);',
      },
    ],
  },
  {
    code: '(array.reverse()).find(logic);',
    message: 'Prefer `Array#findLast()` over `Array#reverse().find()`.',
    suggestions: [
      {
        messageId: 'replace',
        desc: 'Replace `.reverse().find()` with `.findLast()`.',
        output: '(array).findLast(logic);',
      },
    ],
  },
  {
    code: '(array.toReversed()).find(logic);',
    message: 'Prefer `Array#findLast()` over `Array#toReversed().find()`.',
    suggestions: [
      {
        messageId: 'replace',
        desc: 'Replace `.toReversed().find()` with `.findLast()`.',
        output: '(array).findLast(logic);',
      },
    ],
  },
  {
    code: 'array.reverse(/* comment */).find(logic);',
    message: 'Prefer `Array#findLast()` over `Array#reverse().find()`.',
    suggestions: [],
  },
  {
    code: 'array.toReversed() /* comment */ .reduce(logic);',
    message: 'Prefer `Array#reduceRight()` over `Array#toReversed().reduce()`.',
    suggestions: [],
  },
].map(({ code, message, suggestions }) => ({
  code,
  filename,
  output: null,
  errors: [{ messageId: 'prefer-array-last-methods', message, suggestions }],
}));

ruleTester.run('prefer-array-last-methods', null as never, { valid, invalid });

// RuleTester checks diagnostics; exercise the real edit pipeline explicitly.
test('applies the exact upstream suggestions without autofixes', async () => {
  const configFile = path.resolve(import.meta.dirname, '../rslint.config.mjs');
  const absoluteFilename = path.resolve(import.meta.dirname, '..', filename);
  const { config, configDirectory } = await buildConfigForSettings(
    configFile,
    undefined,
  );
  for (const testCase of invalid) {
    const result = await lint({
      workingDirectory: process.cwd(),
      configDirectory,
      config: [
        ...config,
        { rules: { 'unicorn/prefer-array-last-methods': 'error' } },
      ],
      fileContents: { [absoluteFilename]: testCase.code },
      fix: false,
    });
    expect(result.output ?? {}).toEqual({});
    expect(result.diagnostics.flatMap(({ fixes }) => fixes ?? [])).toEqual([]);
    expect(
      result.diagnostics.map(({ messageId, message, suggestions }) => ({
        messageId,
        message,
        suggestions: (suggestions ?? []).map((suggestion) => ({
          messageId: suggestion.messageId,
          desc: suggestion.message,
          output: [...(suggestion.fixes ?? [])]
            .sort((left, right) => right.startPos - left.startPos)
            .reduce(
              (source, fix) =>
                source.slice(0, fix.startPos) +
                fix.text +
                source.slice(fix.endPos),
              testCase.code,
            ),
        })),
      })),
    ).toEqual(testCase.errors);
  }
});

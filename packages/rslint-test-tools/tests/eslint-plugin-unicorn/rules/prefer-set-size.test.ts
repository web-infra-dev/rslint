import path from 'node:path';
import { lint } from '@rslint/core/internal';

import { RuleTester } from '../rule-tester';
import { buildConfigForSettings } from '../../src/util/load-test-config';

const ruleTester = new RuleTester();
const js = (code: string) => ({ code, filename: 'file.js' });
const error = {
  messageId: 'prefer-set-size',
  message: 'Prefer using `Set#size` instead of `Array#length`.',
};
const exactCases: Array<{
  code: string;
  output?: string;
  filename: string;
}> = [];
const invalid = (code: string, output: string, filename = 'file.js') => {
  exactCases.push({ code, output, filename });
  return { code, output, filename, errors: [error] };
};
const invalidWithoutFix = (code: string, filename = 'file.js') => {
  exactCases.push({ code, filename });
  return { ...js(code), filename, errors: [error] };
};

// Rslint-specific helper regressions. These cases are deliberately separate
// from eslint-plugin-unicorn v74.0.0's prefer-set-size test corpus.
const rslintRegressions = {
  valid: [
    {
      code: 'declare const args: [Set<number>]; Array.from(...args).length',
      filename: 'file.ts',
    },
    {
      code: 'function size(value: unknown) { return Array.from(value satisfies Set<string>).length; }',
      filename: 'file.ts',
    },
  ],
  invalid: [
    invalid(
      '[...(flag ? new Set() : new Set())].length',
      '(flag ? new Set() : new Set()).size',
    ),
    invalid(
      '[...(sideEffect(), new Set())].length',
      '(sideEffect(), new Set()).size',
    ),
    invalid(
      'function size(value: unknown) { return [...(new Set() satisfies Set)].length; }',
      'function size(value: unknown) { return (new Set() satisfies Set).size; }',
      'file.ts',
    ),
  ],
};

// Mirrors eslint-plugin-unicorn v74.0.0's prefer-set-size suite. Rslint-only
// helper and fix-safety regressions are explicitly appended below.
ruleTester.run('prefer-set-size', null as never, {
  valid: [
    js('new Set(foo).size'),
    js('for (const foo of bar) console.log([...foo].length)'),
    js('[...new Set(array), foo].length'),
    js('[foo, ...new Set(array), ].length'),
    js('[...new Set(array)].notLength'),
    js('[...new Set(array)]?.length'),
    js('[...new Set(array)][length]'),
    js('[...new Set(array)]["length"]'),
    js('[...new NotSet(array)].length'),
    js('[...Set(array)].length'),
    js('const foo = new NotSet([]);[...foo].length;'),
    js('let foo = new Set([]);[...foo].length;'),
    js('const {foo} = new Set([]);[...foo].length;'),
    js('const [foo] = new Set([]);[...foo].length;'),
    js('[...foo].length'),
    js('var foo = new Set(); var foo = new Set(); [...foo].length'),
    js('[,].length'),
    js('Array.from(foo).length'),
    js('Array.from(new NotSet(array)).length'),
    js('Array.from(Set(array)).length'),
    js('Array.from(new Set(array)).notLength'),
    js('Array.from(new Set(array))?.length'),
    js('Array.from(new Set(array))[length]'),
    js('Array.from(new Set(array))["length"]'),
    js('Array.from(new Set(array), mapFn).length'),
    js('Array?.from(new Set(array)).length'),
    js('Array.from?.(new Set(array)).length'),
    js('const foo = new NotSet([]);Array.from(foo).length;'),
    js('let foo = new Set([]);Array.from(foo).length;'),
    js('const {foo} = new Set([]);Array.from(foo).length;'),
    js('const [foo] = new Set([]);Array.from(foo).length;'),
    js('var foo = new Set(); var foo = new Set(); Array.from(foo).length'),
    js('NotArray.from(new Set(array)).length'),
    ...rslintRegressions.valid,
  ],
  invalid: [
    invalid('[...new Set(array)].length', 'new Set(array).size'),
    invalid(
      'const foo = new Set([]);\nconsole.log([...foo].length);',
      'const foo = new Set([]);\nconsole.log(foo.size);',
    ),
    invalid(
      'function isUnique(array) {\n\treturn[...new Set(array)].length === array.length\n}',
      'function isUnique(array) {\n\treturn new Set(array).size === array.length\n}',
    ),
    invalid('[...new Set(array),].length', 'new Set(array).size'),
    invalid('[...(( new Set(array) ))].length', 'new Set(array).size'),
    invalid('(( [...new Set(array)] )).length', '(( new Set(array) )).size'),
    invalid('foo\n;[...new Set(array)].length', 'foo\n;new Set(array).size'),
    invalidWithoutFix('[/* comment */...new Set(array)].length'),
    invalid(
      '[...new /* comment */ Set(array)].length',
      'new /* comment */ Set(array).size',
    ),
    invalid('Array.from(new Set(array)).length', 'new Set(array).size'),
    invalid(
      'const foo = new Set([]);\nconsole.log(Array.from(foo).length);',
      'const foo = new Set([]);\nconsole.log(foo.size);',
    ),
    invalid('Array.from((( new Set(array) ))).length', 'new Set(array).size'),
    invalid(
      '(( Array.from(new Set(array)) )).length',
      '(( new Set(array) )).size',
    ),
    invalidWithoutFix('Array.from(/* comment */ new Set(array)).length'),
    invalid(
      'Array.from(new /* comment */ Set(array)).length',
      'new /* comment */ Set(array).size',
    ),
    invalid(
      'function isUnique(array) {\n\treturn Array.from(new Set(array)).length === array.length\n}',
      'function isUnique(array) {\n\treturn new Set(array).size === array.length\n}',
    ),
    invalid(
      'function getSize(set: Set<string>) { return Array.from(set).length; }',
      'function getSize(set: Set<string>) { return set.size; }',
      'file.ts',
    ),
    invalid(
      'function getSize(set: ReadonlySet<string>) { return Array.from(set).length; }',
      'function getSize(set: ReadonlySet<string>) { return set.size; }',
      'file.ts',
    ),
    invalid(
      'function getSize(set: unknown) { return [...(set as Set<string>)].length; }',
      'function getSize(set: unknown) { return (set as Set<string>).size; }',
      'file.ts',
    ),
    invalid(
      'function getSize(set: unknown) { return Array.from(set as Set<string>).length; }',
      'function getSize(set: unknown) { return (set as Set<string>).size; }',
      'file.ts',
    ),
    ...rslintRegressions.invalid,
  ],
});

const directory = path.resolve(import.meta.dirname, '..');
const config = path.resolve(directory, 'rslint.config.mjs');
const run = async (code: string, filename: string, fix = false) => {
  const absoluteFilename = path.join(directory, filename);
  const { config: resolvedConfig, configDirectory } =
    await buildConfigForSettings(config, undefined);
  return lint({
    config: [
      ...resolvedConfig,
      {
        rules: { 'unicorn/prefer-set-size': 'error' },
      },
    ],
    configDirectory,
    workingDirectory: process.cwd(),
    fileContents: { [absoluteFilename]: code },
    fix,
  });
};

const lengthRange = (code: string) => {
  const offset = code.indexOf('length');
  const lineStart = code.lastIndexOf('\n', offset) + 1;
  const line = code.slice(0, offset).split('\n').length;
  const column = offset - lineStart + 1;
  return {
    start: { line, column },
    end: { line, column: column + 'length'.length },
  };
};

describe('unicorn/prefer-set-size exact diagnostics and fixes', () => {
  test('matches upstream diagnostics, ranges, and fixed output', async () => {
    for (const item of exactCases) {
      const result = await run(item.code, item.filename);
      expect(result.fileCount).toBe(1);
      expect(result.diagnostics, item.code).toHaveLength(1);
      const [diagnostic] = result.diagnostics;
      expect(diagnostic.ruleName).toBe('unicorn/prefer-set-size');
      expect(diagnostic.messageId).toBe('prefer-set-size');
      expect(diagnostic.message).toBe(error.message);
      expect(diagnostic.range).toEqual(lengthRange(item.code));

      if (item.output === undefined) {
        expect(diagnostic.fixes ?? []).toEqual([]);
        expect(result.fixableErrorCount).toBe(0);
        continue;
      }

      expect(diagnostic.fixes).toHaveLength(2);
      expect(result.fixableErrorCount).toBe(1);
      const fixed = await run(item.code, item.filename, true);
      expect(Object.values(fixed.output ?? {})).toEqual([item.output]);
      expect(fixed.fixableErrorCount).toBe(0);
      expect((await run(item.output, item.filename)).diagnostics).toEqual([]);
    }
  });
});

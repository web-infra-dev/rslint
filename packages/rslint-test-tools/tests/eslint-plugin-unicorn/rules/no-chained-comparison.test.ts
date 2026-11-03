// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/no-chained-comparison.js
import path from 'node:path';
import { runInNewContext } from 'node:vm';

import { lint } from '@rslint/core/internal';

import { buildConfigForSettings } from '../../src/util/load-test-config';
import { RuleTester } from '../rule-tester';

const error = {
  messageId: 'no-chained-comparison/error',
  message:
    'Comparison operators cannot be chained. The inner comparison evaluates to a boolean, which is then compared instead of the operands.',
};

const upstreamCases = {
  valid: [
    // test group 3
    {
      code: '/** @param {boolean} flag */ function run(flag) { return (a < b) === flag; }',
      filename: 'src/file.ts',
    },
    {
      code: '/** @param {() => boolean} read */ function run(read) { return (a < b) !== read(); }',
      filename: 'src/file.ts',
    },
    {
      code: '/** @param {string} text */ function run(text) { return (a < b) == text.startsWith("x"); }',
      filename: 'src/file.ts',
    },
    {
      code: '/** @param {Array<number> | Uint8Array} values */ function run(values) { return values.some(predicate) === (a < b); }',
      filename: 'src/file.ts',
    },
    {
      code: '/** @returns {RegExp} */ function read() { return source; } (a < b) === read().global;',
      filename: 'src/file.ts',
    },
    // test group 6
    {
      code: '/** @param {string} flag */ function run(flag: boolean) { return (a < b) === flag; }',
      filename: 'src/file.ts',
    },
    // snapshot cases
    { code: 'a < b', filename: 'src/case.js' },
    { code: 'a === b', filename: 'src/case.js' },
    { code: 'a < b && b < c', filename: 'src/case.js' },
    { code: 'a < b && b < c && c < d', filename: 'src/case.js' },
    { code: '(a > 0) === (b > 0)', filename: 'src/case.js' },
    { code: '(a > 0) !== (b > 0)', filename: 'src/case.js' },
    { code: '(a > 0) == (b > 0)', filename: 'src/case.js' },
    { code: '(a < b) === true', filename: 'src/case.js' },
    { code: '(a < b) !== false', filename: 'src/case.js' },
    { code: '(a < b) == true', filename: 'src/case.js' },
    { code: 'true === (a < b)', filename: 'src/case.js' },
    { code: 'a == b == true', filename: 'src/case.js' },
    { code: '(a > 0) !== Boolean(b)', filename: 'src/case.js' },
    { code: '(a < b) === !c', filename: 'src/case.js' },
    { code: '(a < b) === (c ? true : false)', filename: 'src/case.js' },
    { code: 'const d = a > b; (a < b) === d;', filename: 'src/case.js' },
    { code: 'a < b || c < d', filename: 'src/case.js' },
    { code: '!(a < b)', filename: 'src/case.js' },
    { code: 'foo(a < b, c)', filename: 'src/case.js' },
    { code: 'a < b ? c : d', filename: 'src/case.js' },
    { code: 'const x = a < b;', filename: 'src/case.js' },
    { code: 'f(a < b) === g(c < d)', filename: 'src/case.js' },
    { code: 'a instanceof b instanceof c', filename: 'src/case.js' },
    { code: 'a in b in c', filename: 'src/case.js' },
    { code: 'a + b + c', filename: 'src/case.js' },
    { code: '(a < b) + c', filename: 'src/case.js' },
    { code: '(a < b) & c', filename: 'src/case.js' },
    { code: 'typeof a < b', filename: 'src/case.js' },
    { code: 'a << b < c', filename: 'src/case.js' },
    { code: '(a as number) < b', filename: 'src/case.ts' },
    // documentation examples
    { code: '// ✅\nif (a < b && b < c) {}', filename: 'src/case.js' },
    { code: '// ✅\nif (a === b && b === c) {}', filename: 'src/case.js' },
    {
      code: '// ✅\n// Intentionally comparing two boolean results.\nif ((a > 0) === (b > 0)) {}',
      filename: 'src/case.js',
    },
  ],
  invalid: [
    // test group 1
    {
      code: 'read(/* keep */) < middle < last;',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: 'first < middle < (/* keep */ last);',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    // test group 2
    {
      code: '(/* keep */ first) < middle < last;',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: 'read(// keep\r\n) < middle < last;',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: 'first < middle < ((last /* keep */));',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: '(first as /* keep */ number) < middle < last;',
      filename: 'src/case.ts',
      output: null,
      errors: [error],
    },
    {
      code: 'first < (middle /* duplicate */) < last;',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: '(/* discard */ first < middle) < last;',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: '(first < middle /* discard */) < last;',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    // test group 3
    {
      code: '/** @param {boolean | undefined} flag */ function run(flag) { return (a < b) === flag; }',
      filename: 'src/file.ts',
      output: null,
      errors: [error],
    },
    {
      code: '/** @param {boolean} [flag] */ function run(flag) { return (a < b) === flag; }',
      filename: 'src/file.ts',
      output: null,
      errors: [error],
    },
    {
      code: '/** @param {() => (() => boolean)} read */ function run(read) { return (a < b) !== read(); }',
      filename: 'src/file.ts',
      output: null,
      errors: [error],
    },
    // test group 6
    {
      code: '/** @param {boolean} flag */ function run(flag: string) { return (a < b) === flag; }',
      filename: 'src/file.ts',
      output: null,
      errors: [error],
    },
    // test group 8
    {
      code: 'a < b < c as boolean',
      filename: 'src/case.ts',
      output: null,
      errors: [error],
    },
    {
      code: '(a < b < c) as boolean',
      filename: 'src/case.ts',
      output: null,
      errors: [error],
    },
    {
      code: 'a < b < c satisfies boolean',
      filename: 'src/case.ts',
      output: null,
      errors: [error],
    },
    {
      code: '(a < b < c) satisfies boolean',
      filename: 'src/case.ts',
      output: null,
      errors: [error],
    },
    // snapshot cases
    {
      code: 'a < b < c',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: 'a > b > c',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: 'a <= b <= c',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: 'a >= b >= c',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: 'a < b <= c',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: 'a < b > c',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: 'a == b == c',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: 'a === b === c',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: 'a != b != c',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: 'a !== b !== c',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: 'a <= b >= c',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: 'a === b !== c',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: '(a < b) < c',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: '(a == b) == c',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: '0 < x < 10',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: 'min <= value <= max',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: '-1 < x < 1',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: 'a < b.c < d',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: 'a < b[c] < d',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: 'a < b?.c < d',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: 'a < b + c < d',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: 'a < b === c',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: 'a === b < c',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: '(a > 0) == 0',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: '(a > 0) === isEnabled',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: 'a == b < c',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: 'a === b > c',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: 'a < f() < c',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: 'a < (b = c) < d',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: 'a < b++ < c',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: 'a < /* comment */ b < c',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: '(a < b) < true',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: 'a < b < false',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: '(a < b) < (c < d)',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: 'a < b < c < d',
      filename: 'src/case.js',
      output: null,
      errors: [error, error],
    },
    {
      code: 'foo === a < b < c',
      filename: 'src/case.js',
      output: null,
      errors: [error, error],
    },
    {
      code: 'a < b < c === foo',
      filename: 'src/case.js',
      output: null,
      errors: [error, error],
    },
    {
      code: 'x & a < b < c',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: 'a ?? b < c < d',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: 'foo === (a < b < c)',
      filename: 'src/case.js',
      output: null,
      errors: [error, error],
    },
    {
      code: 'a && b < c < d',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: 'a || b < c < d',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: 'a < b < c ? x : y',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: 'a < (b ?? 0) < c',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: 'a < b < c',
      filename: 'src/case.ts',
      output: null,
      errors: [error],
    },
    {
      code: 'a < b! < c',
      filename: 'src/case.ts',
      output: null,
      errors: [error],
    },
    {
      code: 'a < (b as number) < c',
      filename: 'src/case.ts',
      output: null,
      errors: [error],
    },
    // test group 10
    {
      code: '({value: 1}.value < b) < c;',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: '(function () {}.length < b) < c;',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: 'const result = ({value: 1}.value < b) < c;',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: 'function check() { return ({value: 1}.value < b) < c; }',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: '(({value: 1}.value < b) < c);',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: '(({value: 1}.value) < b) < c;',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: 'const check = () => ({value: 1}.value < b) < c;',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: 'const check = () => (function () {}.length < b) < c;',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    // documentation examples
    {
      code: '// ❌\nif (a < b < c) {}',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
    {
      code: '// ❌\nif (a === b === c) {}',
      filename: 'src/case.js',
      output: null,
      errors: [error],
    },
  ],
};

new RuleTester().run('no-chained-comparison', {} as never, upstreamCases);

describe('without a TypeScript project', () => {
  const languageOptions = {
    parserOptions: { project: false, projectService: false },
  } as const;
  new RuleTester().run('no-chained-comparison', {} as never, {
    valid: upstreamCases.valid.map((item) => ({ ...item, languageOptions })),
    invalid: upstreamCases.invalid.map((item) => ({
      ...item,
      languageOptions,
    })),
  });
});

// Upstream's runtime assertion: suggestions preserve comments and evaluate the
// first operand once, while deliberately changing the comparison's result.
test('comment-preserving rewrites remain suggestions', async () => {
  const code = 'read(/* keep */) < middle < last;';
  const { config, configDirectory } = await buildConfigForSettings(
    path.resolve(import.meta.dirname, '../rslint.config.mjs'),
    undefined,
  );
  const filename = path.join(configDirectory, 'src/virtual.js');
  const lintCode = (source: string, fix = false) =>
    lint({
      config: [
        ...config,
        { rules: { 'unicorn/no-chained-comparison': 'error' } },
      ],
      configDirectory,
      workingDirectory: process.cwd(),
      fileContents: { [filename]: source },
      fix,
    });
  const result = await lintCode(code, true);
  expect(result.diagnostics).toHaveLength(1);
  expect(result.fixableErrorCount).toBe(0);
  const [diagnostic] = result.diagnostics;
  expect(diagnostic.messageId).toBe('no-chained-comparison/error');
  expect(diagnostic.range).toEqual({
    start: { line: 1, column: 1 },
    end: { line: 1, column: 33 },
  });
  expect(diagnostic.fixes ?? []).toHaveLength(0);
  expect(diagnostic.suggestions).toHaveLength(1);
  const [suggestion] = diagnostic.suggestions!;
  const replacement = 'read(/* keep */) < middle && middle < last';
  expect(suggestion).toEqual({
    messageId: 'no-chained-comparison/suggestion',
    message: `Replace with \`${replacement}\`.`,
    data: { replacement },
    fixes: [{ startPos: 0, endPos: 32, text: replacement }],
  });
  const suggested = `${replacement};`;
  expect((await lintCode(suggested)).diagnostics).toEqual([]);
  for (const first of [0, 5]) {
    for (const source of [code, suggested]) {
      let calls = 0;
      const value: unknown = runInNewContext(source, {
        middle: 2,
        last: 4,
        read() {
          calls++;
          return first;
        },
      });
      expect(calls).toBe(1);
      expect(value).toBe(source === code || first === 0);
    }
  }
});

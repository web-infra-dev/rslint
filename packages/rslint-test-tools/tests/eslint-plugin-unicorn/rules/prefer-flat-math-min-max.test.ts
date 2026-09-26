// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/test/prefer-flat-math-min-max.js
import path from 'node:path';

import { lint } from '@rslint/core/internal';

import { RuleTester } from '../rule-tester';
import { buildConfigForSettings } from '../../src/util/load-test-config';

const ruleTester = new RuleTester();
const filename = 'src/virtual.js';

const valid = [
  'Math.max(a, b, c);',
  'Math.min(a, b, c);',
  'Math.max(Math.min(a, b), c);',
  'Math.min(Math.max(a, b), c);',
  'max(max(a, b), c);',
  'Math.max.apply(Math, [Math.max(a, b), c]);',
  'Math["max"](Math["max"](a, b), c);',
  'Math.max?.(Math.max(a, b), c);',
  'Math?.max(Math.max(a, b), c);',
  'Math.max(Math.max?.(a, b), c);',
  'Math.max(Math?.max(a, b), c);',
  'globalThis.Math.max(Math.max(a, b), c);',
  'Number.max(Number.max(a, b), c);',
  'Math.hypot(Math.hypot(a, b), c);',
].map((code) => ({ code, filename }));

const invalid = [
  {
    code: 'Math.max(Math.max(a, b), c);',
    output: 'Math.max(a, b, c);',
    message: 'Prefer a flat `Math.max()` call instead of nested calls.',
    errorCount: 1,
  },
  {
    code: 'Math.min(a, Math.min(b, c));',
    output: 'Math.min(a, b, c);',
    message: 'Prefer a flat `Math.min()` call instead of nested calls.',
    errorCount: 1,
  },
  {
    code: 'Math.max(Math.max(a, b), Math.max(c, d), e);',
    output: 'Math.max(a, b, c, d, e);',
    message: 'Prefer a flat `Math.max()` call instead of nested calls.',
    errorCount: 1,
  },
  {
    code: 'Math.min(Math.min(Math.min(a, b), c), d);',
    output: 'Math.min(a, b, c, d);',
    message: 'Prefer a flat `Math.min()` call instead of nested calls.',
    errorCount: 1,
  },
  {
    code: 'Math.max(Math.max(a, b));',
    output: 'Math.max(a, b);',
    message: 'Prefer a flat `Math.max()` call instead of nested calls.',
    errorCount: 1,
  },
  {
    code: 'Math.min(Math.min());',
    output: 'Math.min();',
    message: 'Prefer a flat `Math.min()` call instead of nested calls.',
    errorCount: 1,
  },
  {
    code: 'Math.max(Math.max(...values), fallback);',
    output: 'Math.max(...values, fallback);',
    message: 'Prefer a flat `Math.max()` call instead of nested calls.',
    errorCount: 1,
  },
  {
    code: 'const value = Math.max(foo, Math.max(bar, baz)).toString();',
    output: 'const value = Math.max(foo, bar, baz).toString();',
    message: 'Prefer a flat `Math.max()` call instead of nested calls.',
    errorCount: 1,
  },
  {
    code: 'const value = Math.min((Math.min(a, b)), c);',
    output: 'const value = Math.min(a, b, c);',
    message: 'Prefer a flat `Math.min()` call instead of nested calls.',
    errorCount: 1,
  },
  {
    code: 'const value = Math.max(\n\tMath.max(a, b),\n\tc,\n);',
    output: 'const value = Math.max(a, b, c);',
    message: 'Prefer a flat `Math.max()` call instead of nested calls.',
    errorCount: 1,
  },
  {
    code: 'const value = Math.max(\n\tMath.max(\n\t\ta,\n\t\tb,\n\t),\n\tc,\n);',
    output: 'const value = Math.max(a, b, c);',
    message: 'Prefer a flat `Math.max()` call instead of nested calls.',
    errorCount: 1,
  },
  {
    code: 'const value = Math.max(\n\tMath.max(/* keep */ a, b),\n\tc,\n);',
    output: null,
    message: 'Prefer a flat `Math.max()` call instead of nested calls.',
    errorCount: 1,
  },
].map(({ code, output, message, errorCount }) => ({
  code,
  filename,
  output,
  errors: Array.from({ length: errorCount }, () => ({
    messageId: 'prefer-flat-math-min-max',
    message,
  })),
}));

ruleTester.run('prefer-flat-math-min-max', null as never, { valid, invalid });

// RuleTester checks diagnostics; exercise the real edit pipeline explicitly.
test('applies exact upstream autofixes and preserves comment-only reports', async () => {
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
        { rules: { 'unicorn/prefer-flat-math-min-max': 'error' } },
      ],
      fileContents: { [absoluteFilename]: testCase.code },
      fix: true,
    });
    expect(result.output ?? {}).toEqual(
      testCase.output === null ? {} : { [absoluteFilename]: testCase.output },
    );
    expect(result.diagnostics).toHaveLength(
      testCase.output === null ? testCase.errors.length : 0,
    );
  }
});

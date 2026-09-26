// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/test/consistent-template-literal-escape.js
import path from 'node:path';

import { lint } from '@rslint/core/internal';

import { RuleTester } from '../rule-tester';
import { buildConfigForSettings } from '../../src/util/load-test-config';

const ruleTester = new RuleTester();
const filename = 'src/virtual.js';
const valid = [
  'const foo = `\\${a}`',
  'const foo = `hello`',
  'const foo = `$`',
  'const foo = `{`',
  'const foo = ``',
  'const foo = `${a}`',
  'const foo = `${a}${b}`',
  'const foo = String.raw`$\\{a}`',
  'const foo = html`$\\{a}`',
  'const foo = `\\\\\\${a}`',
  "const foo = '$\\{a}'",
].map((code) => ({ code, filename }));

const invalid = [
  {
    code: 'const foo = `$\\{a}`',
    output: 'const foo = `\\${a}`',
    message: 'Use `\\${` instead of `$\\{` to escape in template literals.',
    errorCount: 1,
  },
  {
    code: 'const foo = `\\$\\{a}`',
    output: 'const foo = `\\${a}`',
    message: 'Use `\\${` instead of `$\\{` to escape in template literals.',
    errorCount: 1,
  },
  {
    code: 'const foo = `$\\{a} and $\\{b}`',
    output: 'const foo = `\\${a} and \\${b}`',
    message: 'Use `\\${` instead of `$\\{` to escape in template literals.',
    errorCount: 1,
  },
  {
    code: 'const foo = `\\\\$\\{a}`',
    output: 'const foo = `\\\\\\${a}`',
    message: 'Use `\\${` instead of `$\\{` to escape in template literals.',
    errorCount: 1,
  },
  {
    code: 'const foo = `\\\\\\$\\{a}`',
    output: 'const foo = `\\\\\\${a}`',
    message: 'Use `\\${` instead of `$\\{` to escape in template literals.',
    errorCount: 1,
  },
  {
    code: 'const foo = `$\\{a}${expr}`',
    output: 'const foo = `\\${a}${expr}`',
    message: 'Use `\\${` instead of `$\\{` to escape in template literals.',
    errorCount: 1,
  },
  {
    code: 'const foo = `${expr}$\\{a}`',
    output: 'const foo = `${expr}\\${a}`',
    message: 'Use `\\${` instead of `$\\{` to escape in template literals.',
    errorCount: 1,
  },
  {
    code: 'const foo = `$\\{a}${expr}$\\{b}`',
    output: 'const foo = `\\${a}${expr}\\${b}`',
    message: 'Use `\\${` instead of `$\\{` to escape in template literals.',
    errorCount: 2,
  },
].map(({ code, output, message, errorCount }) => ({
  code,
  filename,
  output,
  errors: Array.from({ length: errorCount }, () => ({
    messageId: 'consistent-template-literal-escape',
    message,
  })),
}));

ruleTester.run('consistent-template-literal-escape', null as never, {
  valid,
  invalid,
});

// RuleTester checks diagnostics; exercise the real edit pipeline explicitly.
test('applies exact upstream autofixes across template elements', async () => {
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
        { rules: { 'unicorn/consistent-template-literal-escape': 'error' } },
      ],
      fileContents: { [absoluteFilename]: testCase.code },
      fix: true,
    });
    expect(result.output ?? {}).toEqual(
      testCase.output === null ? {} : { [filename]: testCase.output },
    );
    expect(result.diagnostics).toHaveLength(
      testCase.output === null ? testCase.errors.length : 0,
    );
  }
});

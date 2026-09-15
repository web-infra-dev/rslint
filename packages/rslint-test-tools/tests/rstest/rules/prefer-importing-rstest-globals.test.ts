import path from 'node:path';

import { lint } from '@rslint/core/internal';

import { buildConfigForSettings } from '../../src/util/load-test-config';
import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-importing-rstest-globals', {} as never, {
  valid: [
    {
      code: `import { describe, expect, it } from '@rstest/core'; describe('suite', () => it('works', () => expect(1).toBe(1)));`,
    },
    {
      code: `import { describe } from 'rstack/test'; describe('suite', () => {});`,
    },
    {
      code: `const { it } = require('@rstest/core'); it('works', () => {});`,
    },
    { code: `function describe(title: string) {} describe('local');` },
    { code: `const expect = makeExpectation; expect(value);` },
    {
      code: `import * as core from '@rstest/core'; core.describe('suite', () => {});`,
    },
  ],
  invalid: [
    {
      code: `describe('user service', () => {
  it('reads a user', () => {
    expect(getUser('1')).toEqual(user);
  });
});`,
      output: `import { describe, expect, it } from '@rstest/core';
describe('user service', () => {
  it('reads a user', () => {
    expect(getUser('1')).toEqual(user);
  });
});`,
      errors: [
        {
          messageId: 'preferImportingRstestGlobals',
          message: 'Import `describe, it, expect` from `@rstest/core`.',
          line: 1,
          column: 1,
        },
      ],
    },
    {
      code: `rs.fn();`,
      output: `import { rs } from '@rstest/core';
rs.fn();`,
      errors: [{ messageId: 'preferImportingRstestGlobals' }],
    },
    {
      code: `import { defineConfig } from 'rstack/test';
test('works', () => {});`,
      output: `import { defineConfig, test } from 'rstack/test';
test('works', () => {});`,
      errors: [{ messageId: 'preferImportingRstestGlobals' }],
    },
  ],
});

describe('interaction with consistent-rstest-namespace', () => {
  const runFix = async (code: string, ruleName: string) => {
    const file = path.resolve(import.meta.dirname, '../src/interaction.ts');
    const configFile = path.resolve(
      import.meta.dirname,
      '../rslint.config.mjs',
    );
    const { config, configDirectory } = await buildConfigForSettings(
      configFile,
      undefined,
    );
    const result = await lint({
      config: [
        ...config,
        {
          rules: {
            'rstest/consistent-rstest-namespace': 'error',
            [ruleName]: 'error',
          },
        },
      ],
      configDirectory,
      workingDirectory: process.cwd(),
      fileContents: { [file]: code },
      fix: true,
    });
    return Object.values(result.output ?? {})[0] ?? code;
  };

  test('no-importing converges after the namespace import is removed', async () => {
    await expect(
      runFix(
        `import { rstest } from '@rstest/core';
rstest.fn();`,
        'rstest/no-importing-rstest-globals',
      ),
    ).resolves.toBe(`
rstest.fn();`);
  });

  test('prefer-importing converges after inserting the used namespace', async () => {
    await expect(
      runFix(`rstest.fn();`, 'rstest/prefer-importing-rstest-globals'),
    ).resolves.toBe(`import { rstest } from '@rstest/core';
rstest.fn();`);
  });
});

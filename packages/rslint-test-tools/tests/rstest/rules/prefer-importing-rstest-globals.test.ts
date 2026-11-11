import path from 'node:path';

import { lint } from '@rslint/core/internal';

import { RuleTester } from '@rslint/rule-tester';
import { config, withFixture } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('prefer-importing-rstest-globals', {
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
          endLine: 1,
          endColumn: 9,
        },
      ],
    },
    {
      code: `rs.fn();`,
      output: `import { rs } from '@rstest/core';
rs.fn();`,
      errors: [
        {
          messageId: 'preferImportingRstestGlobals',
          message: 'Import `rs` from `@rstest/core`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 3,
        },
      ],
    },
    {
      code: `import { defineConfig } from 'rstack/test';
test('works', () => {});`,
      output: `import { defineConfig, test } from 'rstack/test';
test('works', () => {});`,
      errors: [
        {
          messageId: 'preferImportingRstestGlobals',
          message: 'Import `test` from `@rstest/core`.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 5,
        },
      ],
    },
  ],
});

describe('interaction with consistent-rstest-namespace', () => {
  const runFix = async (code: string, ruleName: string) => {
    const result = await withFixture((root) =>
      lint({
        config: [
          {
            plugins: ['rstest'],
            languageOptions: config.languageOptions,
            rules: {
              'rstest/consistent-rstest-namespace': 'error',
              [ruleName]: 'error',
            },
          },
        ],
        configDirectory: root,
        workingDirectory: root,
        fileContents: { [path.join(root, config.filename!)]: code },
        fix: true,
      }),
    );
    // Both rules must execute: consistent-rstest-namespace requires type information.
    expect(result.ruleCount).toBe(2);
    expect(result.diagnostics).toEqual([]);
    return Object.values(result.output ?? {})[0] ?? code;
  };

  test('no-importing converges to the preferred global namespace', async () => {
    await expect(
      runFix(
        `import { rstest } from '@rstest/core';
rstest.fn();`,
        'rstest/no-importing-rstest-globals',
      ),
    ).resolves.toBe(`
rs.fn();`);
  });

  test('prefer-importing converges to the preferred namespace import', async () => {
    await expect(
      runFix(`rstest.fn();`, 'rstest/prefer-importing-rstest-globals'),
    ).resolves.toBe(`import { rs } from '@rstest/core';
rs.fn();`);
  });
});

import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('no-importing-rstest-globals', {
  valid: [
    { code: `import { defineConfig } from '@rstest/core';` },
    { code: `import type { Mock } from '@rstest/core';` },
    { code: `import { type Mock } from '@rstest/core';` },
    { code: `import * as core from '@rstest/core';` },
    { code: `import core from '@rstest/core';` },
    { code: `import { test } from 'vitest';` },
    { code: `import { test } from 'rstack/lib';` },
  ],
  invalid: [
    {
      code: `import { describe, it } from '@rstest/core';`,
      output: '',
      errors: [
        {
          messageId: 'noImportingRstestGlobals',
          message:
            'Do not import `describe` from `@rstest/core`; it is available as a global.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 18,
        },
        {
          messageId: 'noImportingRstestGlobals',
          message:
            'Do not import `it` from `@rstest/core`; it is available as a global.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: `import { defineConfig, expect } from 'rstack/test';`,
      output: `import { defineConfig } from 'rstack/test';`,
      errors: [
        {
          messageId: 'noImportingRstestGlobals',
          message:
            'Do not import `expect` from `rstack/test`; it is available as a global.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: `const { describe, defineConfig } = require('@rstest/core');`,
      output: `const { defineConfig } = require('@rstest/core');`,
      errors: [
        {
          messageId: 'noRequiringRstestGlobals',
          message:
            'Do not require `describe` from `@rstest/core`; it is available as a global.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: `import { it as test } from '@rstest/core'; test('works', () => {});`,
      errors: [
        {
          messageId: 'noImportingRstestGlobals',
          message:
            'Do not import `it` from `@rstest/core`; it is available as a global.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
  ],
});

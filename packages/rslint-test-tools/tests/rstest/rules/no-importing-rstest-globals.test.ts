import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-importing-rstest-globals', {} as never, {
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
        { messageId: 'noImportingRstestGlobals', line: 1, column: 10 },
        { messageId: 'noImportingRstestGlobals', line: 1, column: 20 },
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
        },
      ],
    },
    {
      code: `const { describe, defineConfig } = require('@rstest/core');`,
      output: `const { defineConfig } = require('@rstest/core');`,
      errors: [
        {
          messageId: 'noRequiringRstestGlobals',
          line: 1,
          column: 9,
        },
      ],
    },
    {
      code: `import { it as test } from '@rstest/core'; test('works', () => {});`,
      output: null,
      errors: [{ messageId: 'noImportingRstestGlobals' }],
    },
  ],
});

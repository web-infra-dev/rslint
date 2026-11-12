import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('consistent-rstest-namespace', {
  valid: [
    {
      code: `import { rs } from '@rstest/core';
rs.mock('./service');`,
    },
    { code: `rs.clearAllMocks();` },
    {
      code: `import { rstest } from '@rstest/core';
rstest.mock('./service');`,
      options: [{ fn: 'rstest' }],
    },
    { code: `rstest.clearAllMocks();`, options: [{ fn: 'rstest' }] },
    {
      code: `import { rstest as testUtils } from '@rstest/core';
testUtils.mock('./service');`,
    },
    {
      code: `import { rstest } from './helpers';
rstest.mock('./service');`,
    },
    {
      code: `import * as rstest from '@rstest/core';
rstest.mock('./service');`,
    },
    {
      code: `const rstest = require('@rstest/core');
rstest.mock('./service');`,
    },
    {
      code: `function run(rstest: { mock(path: string): void }) {
  rstest.mock('./service');
}`,
    },
    {
      code: `const rstest = { mock(path: string) {} };
rstest.mock('./service');`,
    },
    {
      code: `const config = { rstest: { mock(path: string) {} } };
config.rstest.mock('./service');`,
    },
    { code: `const spy = rstest.fn;` },
    { code: `import.meta.rstest.mock('./service');` },
    { code: `import type { rstest } from '@rstest/core';` },
    { code: `import { type rstest } from '@rstest/core';` },
  ],
  invalid: [
    {
      code: `import { rstest } from '@rstest/core';
rstest.mock('./service');
rstest.clearAllMocks();`,
      output: `import { rs } from '@rstest/core';
rs.mock('./service');
rs.clearAllMocks();`,
      errors: [
        {
          messageId: 'consistentNamespace',
          message: 'Prefer using `rs` instead of `rstest`',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 16,
        },
        {
          messageId: 'consistentNamespace',
          message: 'Prefer using `rs` instead of `rstest`',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 7,
        },
        {
          messageId: 'consistentNamespace',
          message: 'Prefer using `rs` instead of `rstest`',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 7,
        },
      ],
    },
    {
      code: `import { expect, rs, rstest } from '@rstest/core';`,
      output: `import { expect, rs } from '@rstest/core';`,
      errors: [
        {
          messageId: 'consistentNamespace',
          message: 'Prefer using `rs` instead of `rstest`',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: `import { rstest, rs, expect } from '@rstest/core';`,
      output: `import { rs, expect } from '@rstest/core';`,
      errors: [
        {
          messageId: 'consistentNamespace',
          message: 'Prefer using `rs` instead of `rstest`',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: `import { rstest, /* the namespace */ rs } from '@rstest/core';`,
      output: `import { /* the namespace */ rs } from '@rstest/core';`,
      errors: [
        {
          messageId: 'consistentNamespace',
          message: 'Prefer using `rs` instead of `rstest`',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: `import { rstest /* remove with namespace */, rs } from '@rstest/core';`,
      output: `import { rs } from '@rstest/core';`,
      errors: [
        {
          messageId: 'consistentNamespace',
          message: 'Prefer using `rs` instead of `rstest`',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: `import { rs } from '@rstest/core';
rs.mock('./service');`,
      output: `import { rstest } from '@rstest/core';
rstest.mock('./service');`,
      options: [{ fn: 'rstest' }],
      errors: [
        {
          messageId: 'consistentNamespace',
          message: 'Prefer using `rstest` instead of `rs`',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 12,
        },
        {
          messageId: 'consistentNamespace',
          message: 'Prefer using `rstest` instead of `rs`',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 3,
        },
      ],
    },
    {
      code: `rstest.useFakeTimers();`,
      output: `rs.useFakeTimers();`,
      options: [{ fn: 'rs' }],
      errors: [
        {
          messageId: 'consistentNamespace',
          message: 'Prefer using `rs` instead of `rstest`',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 7,
        },
      ],
    },
    {
      code: `rstest?.mock('./service', () => ({ pay: rstest.fn() }));`,
      output: `rs?.mock('./service', () => ({ pay: rs.fn() }));`,
      errors: [
        {
          messageId: 'consistentNamespace',
          message: 'Prefer using `rs` instead of `rstest`',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 7,
        },
        {
          messageId: 'consistentNamespace',
          message: 'Prefer using `rs` instead of `rstest`',
          line: 1,
          column: 41,
          endLine: 1,
          endColumn: 47,
        },
      ],
    },
    {
      code: `rstest.mocked(pay).mockReturnValue(true);`,
      output: `rs.mocked(pay).mockReturnValue(true);`,
      errors: [
        {
          messageId: 'consistentNamespace',
          message: 'Prefer using `rs` instead of `rstest`',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 7,
        },
      ],
    },
    {
      code: `import { rstest } from '@rstest/core';
rstest!.mock('./service');`,
      output: `import { rs } from '@rstest/core';
rs!.mock('./service');`,
      errors: [
        {
          messageId: 'consistentNamespace',
          message: 'Prefer using `rs` instead of `rstest`',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 16,
        },
        {
          messageId: 'consistentNamespace',
          message: 'Prefer using `rs` instead of `rstest`',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 7,
        },
      ],
    },
    {
      code: `import { rstest } from '@rstest/core';
(rstest as typeof rstest).mock('./service');`,
      errors: [
        {
          messageId: 'consistentNamespace',
          message: 'Prefer using `rs` instead of `rstest`',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 16,
        },
        {
          messageId: 'consistentNamespace',
          message: 'Prefer using `rs` instead of `rstest`',
          line: 2,
          column: 2,
          endLine: 2,
          endColumn: 8,
        },
      ],
    },
    {
      code: `import { rstest } from '@rstest/core';
rstest.fn\`service\`;`,
      output: `import { rs } from '@rstest/core';
rs.fn\`service\`;`,
      errors: [
        {
          messageId: 'consistentNamespace',
          message: 'Prefer using `rs` instead of `rstest`',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 16,
        },
        {
          messageId: 'consistentNamespace',
          message: 'Prefer using `rs` instead of `rstest`',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 7,
        },
      ],
    },
    {
      code: `import { rstest } from '@rstest/core';
const { fn } = rstest;
rstest.mock('./service');`,
      errors: [
        {
          messageId: 'consistentNamespace',
          message: 'Prefer using `rs` instead of `rstest`',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 16,
        },
        {
          messageId: 'consistentNamespace',
          message: 'Prefer using `rs` instead of `rstest`',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 7,
        },
      ],
    },
    {
      code: `import { rstest } from '@rstest/core';
export { rstest };
rstest.mock('./service');`,
      errors: [
        {
          messageId: 'consistentNamespace',
          message: 'Prefer using `rs` instead of `rstest`',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 16,
        },
        {
          messageId: 'consistentNamespace',
          message: 'Prefer using `rs` instead of `rstest`',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 7,
        },
      ],
    },
    {
      code: `const { rstest } = require('@rstest/core');
rstest.mock('./service');`,
      errors: [
        {
          messageId: 'consistentNamespace',
          message: 'Prefer using `rs` instead of `rstest`',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 7,
        },
      ],
    },
    {
      code: `import { rs } from '@rstest/core';
import { rstest } from '@rstest/core';`,
      errors: [
        {
          messageId: 'consistentNamespace',
          message: 'Prefer using `rs` instead of `rstest`',
          line: 2,
          column: 10,
          endLine: 2,
          endColumn: 16,
        },
      ],
    },
    {
      code: `import { expect as check, rs, rstest } from '@rstest/core';`,
      output: `import { expect as check, rs } from '@rstest/core';`,
      errors: [
        {
          messageId: 'consistentNamespace',
          message: 'Prefer using `rs` instead of `rstest`',
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: `import { rs, rstest /* remove with namespace */, expect } from '@rstest/core';`,
      output: `import { rs, expect } from '@rstest/core';`,
      errors: [
        {
          messageId: 'consistentNamespace',
          message: 'Prefer using `rs` instead of `rstest`',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: `import { rs, rstest /* remove with namespace */ } from '@rstest/core';`,
      output: `import { rs } from '@rstest/core';`,
      errors: [
        {
          messageId: 'consistentNamespace',
          message: 'Prefer using `rs` instead of `rstest`',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: `import { rs, rstest, /* the assertion */ expect } from '@rstest/core';`,
      output: `import { rs, /* the assertion */ expect } from '@rstest/core';`,
      errors: [
        {
          messageId: 'consistentNamespace',
          message: 'Prefer using `rs` instead of `rstest`',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: `import { rstest } from '@rstest/core';
const rs = { mock(path: string) {} };
rstest.mock('./service');`,
      errors: [
        {
          messageId: 'consistentNamespace',
          message: 'Prefer using `rs` instead of `rstest`',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 16,
        },
        {
          messageId: 'consistentNamespace',
          message: 'Prefer using `rs` instead of `rstest`',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 7,
        },
      ],
    },
    {
      code: `import { rstest } from '@rstest/core';
function run() {
  const rs = { mock(path: string) {} };
  rstest.mock('./service');
}`,
      errors: [
        {
          messageId: 'consistentNamespace',
          message: 'Prefer using `rs` instead of `rstest`',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 16,
        },
        {
          messageId: 'consistentNamespace',
          message: 'Prefer using `rs` instead of `rstest`',
          line: 4,
          column: 3,
          endLine: 4,
          endColumn: 9,
        },
      ],
    },
    {
      code: `function run() {
  const rs = { mock(path: string) {} };
  rstest.mock('./service');
}`,
      errors: [
        {
          messageId: 'consistentNamespace',
          message: 'Prefer using `rs` instead of `rstest`',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 9,
        },
      ],
    },
    {
      code: `import { rs } from './helpers';
import { rstest } from '@rstest/core';
rstest.mock('./service');`,
      errors: [
        {
          messageId: 'consistentNamespace',
          message: 'Prefer using `rs` instead of `rstest`',
          line: 2,
          column: 10,
          endLine: 2,
          endColumn: 16,
        },
        {
          messageId: 'consistentNamespace',
          message: 'Prefer using `rs` instead of `rstest`',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 7,
        },
      ],
    },
  ],
});

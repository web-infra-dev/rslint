import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const namespaceError = (
  preferred: string,
  disallowed: string,
  line: number,
  column: number,
  endColumn: number,
) => ({
  messageId: 'consistentNamespace',
  message: `Prefer using \`${preferred}\` instead of \`${disallowed}\``,
  line,
  column,
  endLine: line,
  endColumn,
});

ruleTester.run('consistent-rstest-namespace', {} as never, {
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
        namespaceError('rs', 'rstest', 1, 10, 16),
        namespaceError('rs', 'rstest', 2, 1, 7),
        namespaceError('rs', 'rstest', 3, 1, 7),
      ],
    },
    {
      code: `import { expect, rs, rstest } from '@rstest/core';`,
      output: `import { expect, rs } from '@rstest/core';`,
      errors: [namespaceError('rs', 'rstest', 1, 22, 28)],
    },
    {
      code: `import { rstest, rs, expect } from '@rstest/core';`,
      output: `import { rs, expect } from '@rstest/core';`,
      errors: [namespaceError('rs', 'rstest', 1, 10, 16)],
    },
    {
      code: `import { rstest, /* the namespace */ rs } from '@rstest/core';`,
      output: `import { /* the namespace */ rs } from '@rstest/core';`,
      errors: [namespaceError('rs', 'rstest', 1, 10, 16)],
    },
    {
      code: `import { rs } from '@rstest/core';
rs.mock('./service');`,
      output: `import { rstest } from '@rstest/core';
rstest.mock('./service');`,
      options: [{ fn: 'rstest' }],
      errors: [
        namespaceError('rstest', 'rs', 1, 10, 12),
        namespaceError('rstest', 'rs', 2, 1, 3),
      ],
    },
    {
      code: `rstest.useFakeTimers();`,
      output: `rs.useFakeTimers();`,
      options: [{ fn: 'rs' }],
      errors: [namespaceError('rs', 'rstest', 1, 1, 7)],
    },
    {
      code: `rstest?.mock('./service', () => ({ pay: rstest.fn() }));`,
      output: `rs?.mock('./service', () => ({ pay: rs.fn() }));`,
      errors: [
        namespaceError('rs', 'rstest', 1, 1, 7),
        namespaceError('rs', 'rstest', 1, 41, 47),
      ],
    },
    {
      code: `rstest.mocked(pay).mockReturnValue(true);`,
      output: `rs.mocked(pay).mockReturnValue(true);`,
      errors: [namespaceError('rs', 'rstest', 1, 1, 7)],
    },
    {
      code: `import { rstest } from '@rstest/core';
const { fn } = rstest;
rstest.mock('./service');`,
      errors: [
        namespaceError('rs', 'rstest', 1, 10, 16),
        namespaceError('rs', 'rstest', 3, 1, 7),
      ],
    },
    {
      code: `import { rstest } from '@rstest/core';
export { rstest };
rstest.mock('./service');`,
      errors: [
        namespaceError('rs', 'rstest', 1, 10, 16),
        namespaceError('rs', 'rstest', 3, 1, 7),
      ],
    },
    {
      code: `const { rstest } = require('@rstest/core');
rstest.mock('./service');`,
      errors: [namespaceError('rs', 'rstest', 2, 1, 7)],
    },
    {
      code: `import { rs } from '@rstest/core';
import { rstest } from '@rstest/core';`,
      errors: [namespaceError('rs', 'rstest', 2, 10, 16)],
    },
    {
      code: `import { expect as check, rs, rstest } from '@rstest/core';`,
      output: `import { expect as check, rs } from '@rstest/core';`,
      errors: [namespaceError('rs', 'rstest', 1, 31, 37)],
    },
  ],
});

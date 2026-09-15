import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-rs-mocked', {} as never, {
  valid: [
    {
      code: `import type { Mock } from './test-helpers';
(getUser as Mock).mockReturnValue(user);`,
    },
    {
      code: `import type { MockedFunction } from '@rstest/core';
(getUser as MockedFunction).mock;`,
    },
    {
      code: `import type { Mock } from '@rstest/core';
getUser satisfies Mock;`,
    },
    {
      code: `import type { Mock } from '@rstest/core';
const getUser: Mock = createMock();`,
    },
    {
      code: `import * as core from '@rstest/core';
(getUser as core.Mock).mockReturnValue(user);`,
    },
    {
      code: `import type { Mock } from '@rstest/core';
rs.mocked(getUser).mockReturnValue(user);`,
    },
  ],
  invalid: [
    {
      code: `import { rs, type Mock } from '@rstest/core';
(getUser as Mock).mockReturnValue(user);`,
      output: `import { rs, type Mock } from '@rstest/core';
(rs.mocked(getUser)).mockReturnValue(user);`,
      errors: [
        {
          messageId: 'preferRsMocked',
          message: 'Prefer `rs.mocked()` over type assertions',
          line: 2,
          column: 2,
          endLine: 2,
          endColumn: 17,
        },
      ],
    },
    {
      code: `import { rstest, type Mocked } from '@rstest/core';
(mod as Mocked<typeof mod>).init();`,
      output: `import { rstest, type Mocked } from '@rstest/core';
(rstest.mocked(mod)).init();`,
      errors: [
        {
          messageId: 'preferRsMocked',
          line: 2,
          column: 2,
          endLine: 2,
          endColumn: 27,
        },
      ],
    },
    {
      code: `import { type MockInstance } from 'rstack/test';
(spy as MockInstance).mockRestore();`,
      output: `import { type MockInstance } from 'rstack/test';
(rs.mocked(spy)).mockRestore();`,
      errors: [{ messageId: 'preferRsMocked', line: 2, column: 2 }],
    },
    {
      code: `import type { Mock } from '@rstest/core';
(getUser as unknown as Mock).mockReturnValue(user);`,
      output: `import type { Mock } from '@rstest/core';
(rs.mocked(getUser)).mockReturnValue(user);`,
      errors: [{ messageId: 'preferRsMocked', line: 2, column: 2 }],
    },
    {
      code: `import type { Mock } from '@rstest/core';
const rs = createUtilities();
(getUser as Mock).mockReturnValue(user);`,
      output: null,
      errors: [{ messageId: 'preferRsMocked', line: 3, column: 2 }],
    },
    {
      code: `import type { Mock } from '@rstest/core';
const rs = createUtilities();
const rstest = createRunner();
import.meta.rstest!.expect(1).toBe(1);
(getUser as Mock).mockReturnValue(user);`,
      output: null,
      errors: [
        {
          messageId: 'preferRsMocked',
          line: 5,
          column: 2,
          suggestions: [
            {
              messageId: 'preferRsMocked',
              output: `import type { Mock } from '@rstest/core';
const rs = createUtilities();
const rstest = createRunner();
import.meta.rstest!.expect(1).toBe(1);
(import.meta.rstest.rs.mocked(getUser)).mockReturnValue(user);`,
            },
          ],
        },
      ],
    },
  ],
});

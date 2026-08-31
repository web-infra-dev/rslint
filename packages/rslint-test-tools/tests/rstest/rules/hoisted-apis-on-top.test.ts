import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('hoisted-apis-on-top', {} as never, {
  valid: [
    { code: `rs.mock();` },
    {
      code: `
rs.hoisted();
import foo from 'bar';
`,
    },
    {
      code: `
import foo from 'bar';
rs.unmock(baz);
`,
    },
    { code: `const foo = await rs.hoisted(async () => {});` },
    {
      code: `
import { rstest } from '@rstest/core';
rstest.mock('./foo');
`,
    },
    {
      code: `
import { rs as r } from '@rstest/core';
if (foo) {
  r.mock('./foo');
}
`,
    },
  ],
  invalid: [
    {
      code: `
import { rs } from 'some-other-module';
if (foo) {
  rs.mock('./foo');
}
`,
      errors: [
        {
          messageId: 'hoistedApisOnTop',
          message:
            'Hoisted API is used in a runtime location in this file, but it is actually executed before this file is loaded.',
          line: 4,
          column: 3,
          endLine: 4,
          endColumn: 19,
        },
      ],
    },
    {
      code: `
if (foo) {
  rs.mock('foo', () => {});
}
`,
      errors: [
        {
          messageId: 'hoistedApisOnTop',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 27,
        },
      ],
    },
    {
      code: `
import foo from 'bar';

if (foo) {
  rs.hoisted();
}
`,
      errors: [
        {
          messageId: 'hoistedApisOnTop',
          line: 5,
          column: 3,
          endLine: 5,
          endColumn: 15,
        },
      ],
    },
    {
      code: `
import foo from 'bar';

if (foo) {
  rs.unmock();
}
`,
      errors: [
        {
          messageId: 'hoistedApisOnTop',
          line: 5,
          column: 3,
          endLine: 5,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
import foo from 'bar';

if (foo) {
  rs.mock();
}
`,
      errors: [
        {
          messageId: 'hoistedApisOnTop',
          line: 5,
          column: 3,
          endLine: 5,
          endColumn: 12,
        },
      ],
    },
    {
      code: `
if (shouldMock) {
  rs.mock(import('something'), () => bar);
}

import something from 'something';
`,
      errors: [
        {
          messageId: 'hoistedApisOnTop',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 42,
        },
      ],
    },
    {
      code: `
import { rstest } from '@rstest/core';

if (condition) {
  rstest.mock('./foo');
}
`,
      errors: [
        {
          messageId: 'hoistedApisOnTop',
          line: 5,
          column: 3,
          endLine: 5,
          endColumn: 23,
        },
      ],
    },
    {
      code: `
import { rs } from '@rstest/core';

if (condition) {
  rs.hoisted(() => {});
}
`,
      errors: [
        {
          messageId: 'hoistedApisOnTop',
          line: 5,
          column: 3,
          endLine: 5,
          endColumn: 23,
        },
      ],
    },
    {
      code: `
import { rstest } from '@rstest/core';

if (condition) {
  rstest.unmock('./foo');
}
`,
      errors: [
        {
          messageId: 'hoistedApisOnTop',
          line: 5,
          column: 3,
          endLine: 5,
          endColumn: 25,
        },
      ],
    },
    {
      code: `
import { rs } from '@rstest/core';

if (condition) {
  rs.mock('./foo');
}
`,
      errors: [
        {
          messageId: 'hoistedApisOnTop',
          line: 5,
          column: 3,
          endLine: 5,
          endColumn: 19,
        },
      ],
    },
  ],
});

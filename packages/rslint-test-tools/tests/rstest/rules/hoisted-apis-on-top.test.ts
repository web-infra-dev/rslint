import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('hoisted-apis-on-top', {
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
          suggestions: [
            {
              messageId: 'suggestMoveHoistedApiToTop',
              desc: 'Move this hoisted API to the top of the file to better reflect its behavior.',
              output:
                "\nimport { rs } from 'some-other-module';\nrs.mock('./foo');\nif (foo) {\n  ;\n}\n",
            },
            {
              messageId: 'suggestUseNonHoistedApi',
              desc: "Replace 'rs.mock()' with 'rs.doMock()', which is not hoisted.",
              output:
                "\nimport { rs } from 'some-other-module';\nif (foo) {\n  rs.doMock('./foo');\n}\n",
            },
          ],
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
          message:
            'Hoisted API is used in a runtime location in this file, but it is actually executed before this file is loaded.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 27,
          suggestions: [
            {
              messageId: 'suggestMoveHoistedApiToTop',
              desc: 'Move this hoisted API to the top of the file to better reflect its behavior.',
              output: "rs.mock('foo', () => {});\n\nif (foo) {\n  ;\n}\n",
            },
            {
              messageId: 'suggestUseNonHoistedApi',
              desc: "Replace 'rs.mock()' with 'rs.doMock()', which is not hoisted.",
              output: "\nif (foo) {\n  rs.doMock('foo', () => {});\n}\n",
            },
          ],
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
          message:
            'Hoisted API is used in a runtime location in this file, but it is actually executed before this file is loaded.',
          line: 5,
          column: 3,
          endLine: 5,
          endColumn: 15,
          suggestions: [
            {
              messageId: 'suggestMoveHoistedApiToTop',
              desc: 'Move this hoisted API to the top of the file to better reflect its behavior.',
              output:
                "\nimport foo from 'bar';\nrs.hoisted();\n\nif (foo) {\n  ;\n}\n",
            },
          ],
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
          message:
            'Hoisted API is used in a runtime location in this file, but it is actually executed before this file is loaded.',
          line: 5,
          column: 3,
          endLine: 5,
          endColumn: 14,
          suggestions: [
            {
              messageId: 'suggestMoveHoistedApiToTop',
              desc: 'Move this hoisted API to the top of the file to better reflect its behavior.',
              output:
                "\nimport foo from 'bar';\nrs.unmock();\n\nif (foo) {\n  ;\n}\n",
            },
            {
              messageId: 'suggestUseNonHoistedApi',
              desc: "Replace 'rs.unmock()' with 'rs.doUnmock()', which is not hoisted.",
              output:
                "\nimport foo from 'bar';\n\nif (foo) {\n  rs.doUnmock();\n}\n",
            },
          ],
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
          message:
            'Hoisted API is used in a runtime location in this file, but it is actually executed before this file is loaded.',
          line: 5,
          column: 3,
          endLine: 5,
          endColumn: 12,
          suggestions: [
            {
              messageId: 'suggestMoveHoistedApiToTop',
              desc: 'Move this hoisted API to the top of the file to better reflect its behavior.',
              output:
                "\nimport foo from 'bar';\nrs.mock();\n\nif (foo) {\n  ;\n}\n",
            },
            {
              messageId: 'suggestUseNonHoistedApi',
              desc: "Replace 'rs.mock()' with 'rs.doMock()', which is not hoisted.",
              output:
                "\nimport foo from 'bar';\n\nif (foo) {\n  rs.doMock();\n}\n",
            },
          ],
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
          message:
            'Hoisted API is used in a runtime location in this file, but it is actually executed before this file is loaded.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 42,
          suggestions: [
            {
              messageId: 'suggestMoveHoistedApiToTop',
              desc: 'Move this hoisted API to the top of the file to better reflect its behavior.',
              output:
                "\nif (shouldMock) {\n  ;\n}\n\nimport something from 'something';\nrs.mock(import('something'), () => bar);\n",
            },
            {
              messageId: 'suggestUseNonHoistedApi',
              desc: "Replace 'rs.mock()' with 'rs.doMock()', which is not hoisted.",
              output:
                "\nif (shouldMock) {\n  rs.doMock(import('something'), () => bar);\n}\n\nimport something from 'something';\n",
            },
          ],
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
          message:
            'Hoisted API is used in a runtime location in this file, but it is actually executed before this file is loaded.',
          line: 5,
          column: 3,
          endLine: 5,
          endColumn: 23,
          suggestions: [
            {
              messageId: 'suggestMoveHoistedApiToTop',
              desc: 'Move this hoisted API to the top of the file to better reflect its behavior.',
              output:
                "\nimport { rstest } from '@rstest/core';\nrstest.mock('./foo');\n\nif (condition) {\n  ;\n}\n",
            },
            {
              messageId: 'suggestUseNonHoistedApi',
              desc: "Replace 'rstest.mock()' with 'rstest.doMock()', which is not hoisted.",
              output:
                "\nimport { rstest } from '@rstest/core';\n\nif (condition) {\n  rstest.doMock('./foo');\n}\n",
            },
          ],
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
          message:
            'Hoisted API is used in a runtime location in this file, but it is actually executed before this file is loaded.',
          line: 5,
          column: 3,
          endLine: 5,
          endColumn: 23,
          suggestions: [
            {
              messageId: 'suggestMoveHoistedApiToTop',
              desc: 'Move this hoisted API to the top of the file to better reflect its behavior.',
              output:
                "\nimport { rs } from '@rstest/core';\nrs.hoisted(() => {});\n\nif (condition) {\n  ;\n}\n",
            },
          ],
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
          message:
            'Hoisted API is used in a runtime location in this file, but it is actually executed before this file is loaded.',
          line: 5,
          column: 3,
          endLine: 5,
          endColumn: 25,
          suggestions: [
            {
              messageId: 'suggestMoveHoistedApiToTop',
              desc: 'Move this hoisted API to the top of the file to better reflect its behavior.',
              output:
                "\nimport { rstest } from '@rstest/core';\nrstest.unmock('./foo');\n\nif (condition) {\n  ;\n}\n",
            },
            {
              messageId: 'suggestUseNonHoistedApi',
              desc: "Replace 'rstest.unmock()' with 'rstest.doUnmock()', which is not hoisted.",
              output:
                "\nimport { rstest } from '@rstest/core';\n\nif (condition) {\n  rstest.doUnmock('./foo');\n}\n",
            },
          ],
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
          message:
            'Hoisted API is used in a runtime location in this file, but it is actually executed before this file is loaded.',
          line: 5,
          column: 3,
          endLine: 5,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'suggestMoveHoistedApiToTop',
              desc: 'Move this hoisted API to the top of the file to better reflect its behavior.',
              output:
                "\nimport { rs } from '@rstest/core';\nrs.mock('./foo');\n\nif (condition) {\n  ;\n}\n",
            },
            {
              messageId: 'suggestUseNonHoistedApi',
              desc: "Replace 'rs.mock()' with 'rs.doMock()', which is not hoisted.",
              output:
                "\nimport { rs } from '@rstest/core';\n\nif (condition) {\n  rs.doMock('./foo');\n}\n",
            },
          ],
        },
      ],
    },
  ],
});

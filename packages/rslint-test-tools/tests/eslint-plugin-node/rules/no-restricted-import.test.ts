// All upstream tests and documentation examples, eslint-plugin-n v18.3.0.
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-restricted-import.js
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-restricted-import.md
import { RuleTester } from '../rule-tester';

new RuleTester({ languageOptions: { sourceType: 'module' } }).run(
  'no-restricted-import',
  {},
  {
    valid: [
      {
        name: 'upstream valid 1',
        code: 'import "fs"',
        options: [['crypto']],
      },
      {
        name: 'upstream valid 2',
        code: 'import "path"',
        options: [['crypto', 'stream', 'os']],
      },
      {
        name: 'upstream valid 3',
        code: 'import "fs "',
      },
      {
        name: 'upstream valid 4',
        code: 'import "foo/bar";',
        options: [['foo']],
      },
      {
        name: 'upstream valid 5',
        code: 'import "foo/bar";',
        options: [
          [
            {
              name: ['foo', 'bar'],
            },
          ],
        ],
      },
      {
        name: 'upstream valid 6',
        code: 'import "foo/bar";',
        options: [
          [
            {
              name: ['foo/c*'],
            },
          ],
        ],
      },
      {
        name: 'upstream valid 7',
        code: 'import "foo/bar";',
        options: [
          [
            {
              name: ['foo'],
            },
            {
              name: ['foo/c*'],
            },
          ],
        ],
      },
      {
        name: 'upstream valid 8',
        code: 'import "foo/bar";',
        options: [
          [
            {
              name: ['foo'],
            },
            {
              name: ['foo/*', '!foo/bar'],
            },
          ],
        ],
      },
      {
        name: 'upstream valid 9',
        code: 'import "os "',
        options: [['fs', 'crypto ', 'stream', 'os']],
      },
      {
        name: 'upstream valid 10',
        code: 'import "./foo"',
        options: [['foo']],
      },
      {
        name: 'upstream valid 11',
        code: 'import "foo"',
        options: [['./foo']],
      },
      {
        name: 'upstream valid 12',
        code: 'import "foo/bar";',
        options: [
          [
            {
              name: '@foo/bar',
            },
          ],
        ],
      },
      {
        name: 'upstream valid 13',
        code: 'import "../foo";',
        filename: 'lib/sub/test.js',
        options: (root: string) => [
          [
            {
              name: `${root}/foo`,
            },
          ],
        ],
      },
      {
        name: 'upstream valid 14',
        code: 'import(fs)',
        options: [['fs']],
      },
      {
        name: 'documentation valid 1',
        code: "import crypto from 'crypto';\nimport _ from 'lodash';",
        options: [['fs', 'cluster', 'lodash/*']],
      },
      {
        name: 'documentation valid 2',
        code: "import pick from 'lodash/pick';",
        options: [
          [
            'fs',
            'cluster',
            {
              name: ['lodash/*', '!lodash/pick'],
            },
          ],
        ],
      },
      {
        name: 'documentation valid 3',
        code: "import 'foo-module2'; import 'bar-module2';",
        options: [['foo-module', 'bar-module']],
      },
      {
        name: 'documentation valid 4',
        code: "import 'baz-module/good';",
        options: [
          [
            {
              name: [
                'foo-module/private/*',
                'bar-module/*',
                '!baz-module/good',
              ],
              message: 'Please use xyz-module instead.',
            },
          ],
        ],
      },
    ],
    invalid: [
      {
        name: 'upstream invalid 1',
        code: 'import "fs"',
        options: [['fs']],
        errors: [
          {
            messageId: 'restricted',
            message: "'fs' module is restricted from being used.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 12,
          },
        ],
      },
      {
        name: 'upstream invalid 2',
        code: 'import fs from "fs"',
        options: [['fs']],
        errors: [
          {
            messageId: 'restricted',
            message: "'fs' module is restricted from being used.",
            line: 1,
            column: 16,
            endLine: 1,
            endColumn: 20,
          },
        ],
      },
      {
        name: 'upstream invalid 3',
        code: 'import {} from "fs"',
        options: [['fs']],
        errors: [
          {
            messageId: 'restricted',
            message: "'fs' module is restricted from being used.",
            line: 1,
            column: 16,
            endLine: 1,
            endColumn: 20,
          },
        ],
      },
      {
        name: 'upstream invalid 4',
        code: 'export * from "fs"',
        options: [['fs']],
        errors: [
          {
            messageId: 'restricted',
            message: "'fs' module is restricted from being used.",
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 19,
          },
        ],
      },
      {
        name: 'upstream invalid 5',
        code: 'export {} from "fs"',
        options: [['fs']],
        errors: [
          {
            messageId: 'restricted',
            message: "'fs' module is restricted from being used.",
            line: 1,
            column: 16,
            endLine: 1,
            endColumn: 20,
          },
        ],
      },
      {
        name: 'upstream invalid 6',
        code: 'import "foo/bar";',
        options: [['foo/bar']],
        errors: [
          {
            messageId: 'restricted',
            message: "'foo/bar' module is restricted from being used.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 17,
          },
        ],
      },
      {
        name: 'upstream invalid 7',
        code: 'import "foo/bar";',
        options: [
          [
            {
              name: ['foo/bar'],
            },
          ],
        ],
        errors: [
          {
            messageId: 'restricted',
            message: "'foo/bar' module is restricted from being used.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 17,
          },
        ],
      },
      {
        name: 'upstream invalid 8',
        code: 'import "foo/bar";',
        options: [
          [
            {
              name: ['foo/*'],
            },
          ],
        ],
        errors: [
          {
            messageId: 'restricted',
            message: "'foo/bar' module is restricted from being used.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 17,
          },
        ],
      },
      {
        name: 'upstream invalid 9',
        code: 'import "foo/bar";',
        options: [
          [
            {
              name: ['foo/*'],
            },
            {
              name: ['foo'],
            },
          ],
        ],
        errors: [
          {
            messageId: 'restricted',
            message: "'foo/bar' module is restricted from being used.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 17,
          },
        ],
      },
      {
        name: 'upstream invalid 10',
        code: 'import "foo/bar";',
        options: [
          [
            {
              name: ['foo/*', '!foo/baz'],
            },
            {
              name: ['foo'],
            },
          ],
        ],
        errors: [
          {
            messageId: 'restricted',
            message: "'foo/bar' module is restricted from being used.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 17,
          },
        ],
      },
      {
        name: 'upstream invalid 11',
        code: 'import "foo";',
        options: [
          [
            {
              name: 'foo',
              message: "Please use 'bar' module instead.",
            },
          ],
        ],
        errors: [
          {
            messageId: 'restricted',
            message:
              "'foo' module is restricted from being used. Please use 'bar' module instead.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 13,
          },
        ],
      },
      {
        name: 'upstream invalid 12',
        code: 'import "bar";',
        options: [
          [
            'foo',
            {
              name: 'bar',
              message: "Please use 'baz' module instead.",
            },
            'baz',
          ],
        ],
        errors: [
          {
            messageId: 'restricted',
            message:
              "'bar' module is restricted from being used. Please use 'baz' module instead.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 13,
          },
        ],
      },
      {
        name: 'upstream invalid 13',
        code: 'import "@foo/bar";',
        options: [
          [
            {
              name: '@foo/*',
            },
          ],
        ],
        errors: [
          {
            messageId: 'restricted',
            message: "'@foo/bar' module is restricted from being used.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 18,
          },
        ],
      },
      {
        name: 'upstream invalid 14',
        code: 'import "./foo/bar";',
        options: [
          [
            {
              name: './foo/*',
            },
          ],
        ],
        errors: [
          {
            messageId: 'restricted',
            message: "'./foo/bar' module is restricted from being used.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 19,
          },
        ],
      },
      {
        name: 'upstream invalid 15',
        code: 'import "../foo";',
        filename: 'lib/test.js',
        options: (root: string) => [
          [
            {
              name: `${root}/foo`,
            },
          ],
        ],
        errors: [
          {
            messageId: 'restricted',
            message: "'../foo' module is restricted from being used.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 16,
          },
        ],
      },
      {
        name: 'upstream invalid 16',
        code: 'import "../../foo";',
        filename: 'lib/sub/test.js',
        options: (root: string) => [
          [
            {
              name: `${root}/foo`,
            },
          ],
        ],
        errors: [
          {
            messageId: 'restricted',
            message: "'../../foo' module is restricted from being used.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 19,
          },
        ],
      },
      {
        name: 'upstream invalid 17',
        code: 'import("fs")',
        options: [['fs']],
        errors: [
          {
            messageId: 'restricted',
            message: "'fs' module is restricted from being used.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 12,
          },
        ],
      },
      {
        name: 'documentation invalid 1',
        code: "import fs from 'fs';\nimport cluster from 'cluster';\nimport pick from 'lodash/pick';",
        options: [['fs', 'cluster', 'lodash/*']],
        errors: [
          {
            messageId: 'restricted',
            message: "'fs' module is restricted from being used.",
            line: 1,
            column: 16,
            endLine: 1,
            endColumn: 20,
          },
          {
            messageId: 'restricted',
            message: "'cluster' module is restricted from being used.",
            line: 2,
            column: 21,
            endLine: 2,
            endColumn: 30,
          },
          {
            messageId: 'restricted',
            message: "'lodash/pick' module is restricted from being used.",
            line: 3,
            column: 18,
            endLine: 3,
            endColumn: 31,
          },
        ],
      },
      {
        name: 'documentation invalid 2',
        code: "import 'foo-module'; import 'bar-module';",
        options: [['foo-module', 'bar-module']],
        errors: [
          {
            messageId: 'restricted',
            message: "'foo-module' module is restricted from being used.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 20,
          },
          {
            messageId: 'restricted',
            message: "'bar-module' module is restricted from being used.",
            line: 1,
            column: 29,
            endLine: 1,
            endColumn: 41,
          },
        ],
      },
      {
        name: 'documentation invalid 3',
        code: "import 'foo-module'; import 'bar-module';",
        options: [
          [
            {
              name: 'foo-module',
              message: 'Please use foo-module2 instead.',
            },
            {
              name: 'bar-module',
              message: 'Please use bar-module2 instead.',
            },
          ],
        ],
        errors: [
          {
            messageId: 'restricted',
            message:
              "'foo-module' module is restricted from being used. Please use foo-module2 instead.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 20,
          },
          {
            messageId: 'restricted',
            message:
              "'bar-module' module is restricted from being used. Please use bar-module2 instead.",
            line: 1,
            column: 29,
            endLine: 1,
            endColumn: 41,
          },
        ],
      },
      {
        name: 'documentation invalid 4',
        code: "import 'lodash/pick';\nimport 'foo-module/private/a';\nimport 'bar-module/a';",
        options: [
          [
            {
              name: 'lodash/*',
              message: 'Please use xyz-module instead.',
            },
            {
              name: [
                'foo-module/private/*',
                'bar-module/*',
                '!baz-module/good',
              ],
              message: 'Please use xyz-module instead.',
            },
          ],
        ],
        errors: [
          {
            messageId: 'restricted',
            message:
              "'lodash/pick' module is restricted from being used. Please use xyz-module instead.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 21,
          },
          {
            messageId: 'restricted',
            message:
              "'foo-module/private/a' module is restricted from being used. Please use xyz-module instead.",
            line: 2,
            column: 8,
            endLine: 2,
            endColumn: 30,
          },
          {
            messageId: 'restricted',
            message:
              "'bar-module/a' module is restricted from being used. Please use xyz-module instead.",
            line: 3,
            column: 8,
            endLine: 3,
            endColumn: 22,
          },
        ],
      },
      {
        name: 'documentation invalid 5',
        code: "import '../server/api.js';",
        filename: 'client/input.js',
        options: (root: string) => [
          [
            {
              name: `${root}/server/**`,
              message: "Don't use server code from client code.",
            },
          ],
        ],
        errors: [
          {
            messageId: 'restricted',
            message:
              "'../server/api.js' module is restricted from being used. Don't use server code from client code.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 26,
          },
        ],
      },
      {
        name: 'documentation invalid 6',
        code: "import '../client/view.js';",
        filename: 'server/input.js',
        options: (root: string) => [
          [
            {
              name: `${root}/client/**`,
              message: "Don't use client code from server code.",
            },
          ],
        ],
        errors: [
          {
            messageId: 'restricted',
            message:
              "'../client/view.js' module is restricted from being used. Don't use client code from server code.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 27,
          },
        ],
      },
    ],
  },
);

// All upstream tests and documentation examples, eslint-plugin-n v18.3.0.
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-restricted-require.js
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-restricted-require.md
import path from 'node:path';
import { RuleTester } from '../rule-tester';

new RuleTester({
  languageOptions: { sourceType: 'commonjs', globals: { require: 'readonly' } },
}).run(
  'no-restricted-require',
  {},
  {
    valid: [
      {
        name: 'upstream valid 1',
        code: 'require("fs")',
        options: [['crypto']],
      },
      {
        name: 'upstream valid 2',
        code: 'require("path")',
        options: [['crypto', 'stream', 'os']],
      },
      {
        name: 'upstream valid 3',
        code: 'require("fs ")',
      },
      {
        name: 'upstream valid 4',
        code: 'require(2)',
        options: [['crypto']],
      },
      {
        name: 'upstream valid 5',
        code: 'require(foo)',
        options: [['crypto']],
      },
      {
        name: 'upstream valid 6',
        code: "bar('crypto');",
        options: [['crypto']],
      },
      {
        name: 'upstream valid 7',
        code: 'require("foo/bar");',
        options: [['foo']],
      },
      {
        name: 'upstream valid 8',
        code: 'require("foo/bar");',
        options: [
          [
            {
              name: ['foo', 'bar'],
            },
          ],
        ],
      },
      {
        name: 'upstream valid 9',
        code: 'require("foo/bar");',
        options: [
          [
            {
              name: ['foo/c*'],
            },
          ],
        ],
      },
      {
        name: 'upstream valid 10',
        code: 'require("foo/bar");',
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
        name: 'upstream valid 11',
        code: 'require("foo/bar");',
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
        name: 'upstream valid 12',
        code: 'require("os ")',
        options: [['fs', 'crypto ', 'stream', 'os']],
      },
      {
        name: 'upstream valid 13',
        code: 'require("./foo")',
        options: [['foo']],
      },
      {
        name: 'upstream valid 14',
        code: 'require("foo")',
        options: [['./foo']],
      },
      {
        name: 'upstream valid 15',
        code: 'require("foo/bar");',
        options: [
          [
            {
              name: '@foo/bar',
            },
          ],
        ],
      },
      {
        name: 'upstream valid 16',
        code: 'require("../foo");',
        filename: 'lib/sub/test.js',
        options: (root: string) => [
          [
            {
              name: path.join(root, 'foo'),
            },
          ],
        ],
      },
      {
        name: 'upstream valid 17',
        code: 'require("foo/bar/baz")',
        options: [['foo/*']],
      },
      {
        name: 'documentation valid 1',
        code: "const crypto = require('crypto');\nconst _ = require('lodash');",
        options: [['fs', 'cluster', 'lodash/*']],
      },
      {
        name: 'documentation valid 2',
        code: "const pick = require('lodash/pick');",
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
        code: "require('foo-module2'); require('bar-module2');",
        options: [['foo-module', 'bar-module']],
      },
      {
        name: 'documentation valid 4',
        code: "require('baz-module/good');",
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
        code: 'require("fs")',
        options: [['fs']],
        errors: [
          {
            messageId: 'restricted',
            message: "'fs' module is restricted from being used.",
            line: 1,
            column: 9,
            endLine: 1,
            endColumn: 13,
          },
        ],
      },
      {
        name: 'upstream invalid 2',
        code: 'require("foo/bar");',
        options: [['foo/bar']],
        errors: [
          {
            messageId: 'restricted',
            message: "'foo/bar' module is restricted from being used.",
            line: 1,
            column: 9,
            endLine: 1,
            endColumn: 18,
          },
        ],
      },
      {
        name: 'upstream invalid 3',
        code: 'require("foo/bar");',
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
            column: 9,
            endLine: 1,
            endColumn: 18,
          },
        ],
      },
      {
        name: 'upstream invalid 4',
        code: 'require("foo/bar");',
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
            column: 9,
            endLine: 1,
            endColumn: 18,
          },
        ],
      },
      {
        name: 'upstream invalid 5',
        code: 'require("foo/bar");',
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
            column: 9,
            endLine: 1,
            endColumn: 18,
          },
        ],
      },
      {
        name: 'upstream invalid 6',
        code: 'require("foo/bar");',
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
            column: 9,
            endLine: 1,
            endColumn: 18,
          },
        ],
      },
      {
        name: 'upstream invalid 7',
        code: 'require("foo");',
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
            column: 9,
            endLine: 1,
            endColumn: 14,
          },
        ],
      },
      {
        name: 'upstream invalid 8',
        code: 'require("bar");',
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
            column: 9,
            endLine: 1,
            endColumn: 14,
          },
        ],
      },
      {
        name: 'upstream invalid 9',
        code: 'require("@foo/bar");',
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
            column: 9,
            endLine: 1,
            endColumn: 19,
          },
        ],
      },
      {
        name: 'upstream invalid 10',
        code: 'require("./foo/bar");',
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
            column: 9,
            endLine: 1,
            endColumn: 20,
          },
        ],
      },
      {
        name: 'upstream invalid 11',
        code: 'require("../foo");',
        filename: 'lib/test.js',
        options: (root: string) => [
          [
            {
              name: path.join(root, 'foo'),
            },
          ],
        ],
        errors: [
          {
            messageId: 'restricted',
            message: "'../foo' module is restricted from being used.",
            line: 1,
            column: 9,
            endLine: 1,
            endColumn: 17,
          },
        ],
      },
      {
        name: 'upstream invalid 12',
        code: 'require("../../foo");',
        filename: 'lib/sub/test.js',
        options: (root: string) => [
          [
            {
              name: path.join(root, 'foo'),
            },
          ],
        ],
        errors: [
          {
            messageId: 'restricted',
            message: "'../../foo' module is restricted from being used.",
            line: 1,
            column: 9,
            endLine: 1,
            endColumn: 20,
          },
        ],
      },
      {
        name: 'upstream invalid 13',
        code: 'require("../../foo");',
        filename: 'lib/sub/test.js',
        options: [
          [
            {
              name: '**/foo',
            },
          ],
        ],
        errors: [
          {
            messageId: 'restricted',
            message: "'../../foo' module is restricted from being used.",
            line: 1,
            column: 9,
            endLine: 1,
            endColumn: 20,
          },
        ],
      },
      {
        name: 'documentation invalid 1',
        code: "const fs = require('fs');\nconst cluster = require('cluster');\nconst pick = require('lodash/pick');",
        options: [['fs', 'cluster', 'lodash/*']],
        errors: [
          {
            messageId: 'restricted',
            message: "'fs' module is restricted from being used.",
            line: 1,
            column: 20,
            endLine: 1,
            endColumn: 24,
          },
          {
            messageId: 'restricted',
            message: "'cluster' module is restricted from being used.",
            line: 2,
            column: 25,
            endLine: 2,
            endColumn: 34,
          },
          {
            messageId: 'restricted',
            message: "'lodash/pick' module is restricted from being used.",
            line: 3,
            column: 22,
            endLine: 3,
            endColumn: 35,
          },
        ],
      },
      {
        name: 'documentation invalid 2',
        code: "require('foo-module'); require('bar-module');",
        options: [['foo-module', 'bar-module']],
        errors: [
          {
            messageId: 'restricted',
            message: "'foo-module' module is restricted from being used.",
            line: 1,
            column: 9,
            endLine: 1,
            endColumn: 21,
          },
          {
            messageId: 'restricted',
            message: "'bar-module' module is restricted from being used.",
            line: 1,
            column: 32,
            endLine: 1,
            endColumn: 44,
          },
        ],
      },
      {
        name: 'documentation invalid 3',
        code: "require('foo-module'); require('bar-module');",
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
            column: 9,
            endLine: 1,
            endColumn: 21,
          },
          {
            messageId: 'restricted',
            message:
              "'bar-module' module is restricted from being used. Please use bar-module2 instead.",
            line: 1,
            column: 32,
            endLine: 1,
            endColumn: 44,
          },
        ],
      },
      {
        name: 'documentation invalid 4',
        code: "require('lodash/pick');\nrequire('foo-module/private/a');\nrequire('bar-module/a');",
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
            column: 9,
            endLine: 1,
            endColumn: 22,
          },
          {
            messageId: 'restricted',
            message:
              "'foo-module/private/a' module is restricted from being used. Please use xyz-module instead.",
            line: 2,
            column: 9,
            endLine: 2,
            endColumn: 31,
          },
          {
            messageId: 'restricted',
            message:
              "'bar-module/a' module is restricted from being used. Please use xyz-module instead.",
            line: 3,
            column: 9,
            endLine: 3,
            endColumn: 23,
          },
        ],
      },
      {
        name: 'documentation invalid 5',
        code: "require('../server/api.js');",
        filename: 'client/input.js',
        options: (root: string) => [
          [
            {
              name: path.join(root, 'server/**'),
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
            column: 9,
            endLine: 1,
            endColumn: 27,
          },
        ],
      },
      {
        name: 'documentation invalid 6',
        code: "require('../client/view.js');",
        filename: 'server/input.js',
        options: (root: string) => [
          [
            {
              name: path.join(root, 'client/**'),
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
            column: 9,
            endLine: 1,
            endColumn: 28,
          },
        ],
      },
    ],
  },
);

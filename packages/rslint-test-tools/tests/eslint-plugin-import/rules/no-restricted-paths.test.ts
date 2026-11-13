// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, resolveFixturePaths } from '../config';
new RuleTester(fixtureConfig()).run('no-restricted-paths', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: 'import a from "../client/a"',
      filename: 'fixtures/no-restricted-paths-rule/server/b.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              basePath: '<fixture>/fixtures/no-restricted-paths-rule',
              zones: [
                {
                  target: './server',
                  from: './other',
                },
              ],
            },
          ],
          root,
        ),
    },
    {
      name: 'run 0 valid 1',
      code: 'import a from "../client/a"',
      filename: 'fixtures/no-restricted-paths-rule/server/b.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              basePath: '<fixture>/fixtures/no-restricted-paths-rule',
              zones: [
                {
                  target: '**/*',
                  from: './other',
                },
              ],
            },
          ],
          root,
        ),
    },
    {
      name: 'run 0 valid 2',
      code: 'import a from "../client/a"',
      filename: 'fixtures/no-restricted-paths-rule/client/b.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              basePath: '<fixture>/fixtures/no-restricted-paths-rule',
              zones: [
                {
                  target: './!(client)/**/*',
                  from: './client/**/*',
                },
              ],
            },
          ],
          root,
        ),
    },
    {
      name: 'run 0 valid 3',
      code: 'const a = require("../client/a")',
      filename: 'fixtures/no-restricted-paths-rule/server/consumer.js',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              basePath: '<fixture>/fixtures/no-restricted-paths-rule',
              zones: [
                {
                  target: './server',
                  from: './other',
                },
              ],
            },
          ],
          root,
        ),
    },
    {
      name: 'run 0 valid 4',
      code: 'import b from "../server/b"',
      filename: 'fixtures/no-restricted-paths-rule/client/a.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              basePath: '<fixture>/fixtures/no-restricted-paths-rule',
              zones: [
                {
                  target: './client',
                  from: './other',
                },
              ],
            },
          ],
          root,
        ),
    },
    {
      name: 'run 0 valid 5',
      code: 'import a from "./a"',
      filename: 'fixtures/no-restricted-paths-rule/server/one/a.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              basePath: '<fixture>/fixtures/no-restricted-paths-rule',
              zones: [
                {
                  target: './server/one',
                  from: './server',
                  except: ['./one'],
                },
              ],
            },
          ],
          root,
        ),
    },
    {
      name: 'run 0 valid 6',
      code: 'import a from "../two/a"',
      filename: 'fixtures/no-restricted-paths-rule/server/one/a.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              basePath: '<fixture>/fixtures/no-restricted-paths-rule',
              zones: [
                {
                  target: './server/one',
                  from: './server',
                  except: ['./two'],
                },
              ],
            },
          ],
          root,
        ),
    },
    {
      name: 'run 0 valid 7',
      code: 'import a from "../one/a"',
      filename: 'fixtures/no-restricted-paths-rule/server/two-new/a.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              basePath: '<fixture>/fixtures/no-restricted-paths-rule',
              zones: [
                {
                  target: './server/two',
                  from: './server',
                  except: [],
                },
              ],
            },
          ],
          root,
        ),
    },
    {
      name: 'run 0 valid 8',
      code: 'import A from "../two/a"',
      filename: 'fixtures/no-restricted-paths-rule/server/one/a.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              basePath: '<fixture>/fixtures/no-restricted-paths-rule',
              zones: [
                {
                  target: '**/*',
                  from: './server/**/*',
                  except: ['**/a.ts'],
                },
              ],
            },
          ],
          root,
        ),
    },
    {
      name: 'run 0 valid 9',
      code: 'import a from "../client/a"',
      filename: 'fixtures/no-restricted-paths-rule/server/b.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              basePath: '<fixture>/fixtures/no-restricted-paths-rule',
              zones: [
                {
                  target: ['./server'],
                  from: './other',
                },
              ],
            },
          ],
          root,
        ),
    },
    {
      name: 'run 0 valid 10',
      code: 'import a from "../client/a"',
      filename: 'fixtures/no-restricted-paths-rule/server/b.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              basePath: '<fixture>/fixtures/no-restricted-paths-rule',
              zones: [
                {
                  target: './server',
                  from: ['./other'],
                },
              ],
            },
          ],
          root,
        ),
    },
    {
      name: 'run 0 valid 11',
      code: 'import a from "../one/a"',
      filename: 'fixtures/no-restricted-paths-rule/server/two-new/a.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              basePath: '<fixture>/fixtures/no-restricted-paths-rule',
              zones: [
                {
                  target: ['./server/two', './server/three'],
                  from: './server',
                },
              ],
            },
          ],
          root,
        ),
    },
    {
      name: 'run 0 valid 12',
      code: 'import a from "../one/a"',
      filename: 'fixtures/no-restricted-paths-rule/server/two-new/a.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              basePath: '<fixture>/fixtures/no-restricted-paths-rule',
              zones: [
                {
                  target: './server',
                  from: ['./server/two', './server/three'],
                  except: [],
                },
              ],
            },
          ],
          root,
        ),
    },
    {
      name: 'run 0 valid 13',
      code: 'import a from "../client/a"',
      filename: 'fixtures/no-restricted-paths-rule/client/b.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              basePath: '<fixture>/fixtures/no-restricted-paths-rule',
              zones: [
                {
                  target: './!(client)/**/*',
                  from: ['./client/*', './client/one/*'],
                },
              ],
            },
          ],
          root,
        ),
    },
    {
      name: 'run 0 valid 14',
      code: 'import a from "../client/a"',
      filename: 'fixtures/no-restricted-paths-rule/client/b.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              basePath: '<fixture>/fixtures/no-restricted-paths-rule',
              zones: [
                {
                  target: ['./!(client)/**/*', './client/a/'],
                  from: './client/**/*',
                },
              ],
            },
          ],
          root,
        ),
    },
    {
      name: 'run 0 valid 15',
      code: 'notRequire("../server/b")',
    },
    {
      name: 'run 0 valid 16',
      code: 'notRequire("../server/b")',
      filename: 'fixtures/no-restricted-paths-rule/client/a.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              basePath: '<fixture>/fixtures/no-restricted-paths-rule',
              zones: [
                {
                  target: './client',
                  from: './server',
                },
              ],
            },
          ],
          root,
        ),
    },
    {
      name: 'run 0 valid 17',
      code: 'require("../server/b")',
    },
    {
      name: 'run 0 valid 18',
      code: 'import b from "../server/b"',
    },
    {
      name: 'run 0 valid 19',
      code: 'require("os")',
    },
    {
      name: 'run 0 valid 20',
      code: 'import type a from "../client/a"',
      filename: 'fixtures/no-restricted-paths-rule/server/b.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              basePath: '<fixture>/fixtures/no-restricted-paths-rule',
              zones: [
                {
                  target: './server',
                  from: './other',
                },
              ],
            },
          ],
          root,
        ),
    },
    {
      name: 'run 0 valid 21',
      code: 'import type a from "./a"',
      filename: 'fixtures/no-restricted-paths-rule/server/one/a.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              basePath: '<fixture>/fixtures/no-restricted-paths-rule',
              zones: [
                {
                  target: './server/one',
                  from: './server',
                  except: ['./one'],
                },
              ],
            },
          ],
          root,
        ),
    },
    {
      name: 'run 0 valid 22',
      code: 'import type b from "../server/b"',
    },
    {
      name: 'run 0 valid 23',
      code: 'import type * as b from "../server/b"',
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: 'import b from "../server/b"; // 1',
      filename: 'fixtures/no-restricted-paths-rule/client/a.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              basePath: '<fixture>/fixtures/no-restricted-paths-rule',
              zones: [
                {
                  target: './client',
                  from: './server',
                },
              ],
            },
          ],
          root,
        ),
      errors: [
        {
          messageId: 'unexpectedPath',
          message: 'Unexpected path "../server/b" imported in restricted zone.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      name: 'run 0 invalid 1',
      code: 'import b from "../server/b"; // 2',
      filename: 'fixtures/no-restricted-paths-rule/client/a.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              basePath: '<fixture>/fixtures/no-restricted-paths-rule',
              zones: [
                {
                  target: './client/**/*',
                  from: './server',
                },
              ],
            },
          ],
          root,
        ),
      errors: [
        {
          messageId: 'unexpectedPath',
          message: 'Unexpected path "../server/b" imported in restricted zone.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      name: 'run 0 invalid 2',
      code: 'import b from "../server/b";',
      filename: 'fixtures/no-restricted-paths-rule/client/a.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              basePath: '<fixture>/fixtures/no-restricted-paths-rule',
              zones: [
                {
                  target: './client/*.ts',
                  from: './server',
                },
              ],
            },
          ],
          root,
        ),
      errors: [
        {
          messageId: 'unexpectedPath',
          message: 'Unexpected path "../server/b" imported in restricted zone.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      name: 'run 0 invalid 3',
      code: 'import b from "../server/b"; // 2 ter',
      filename: 'fixtures/no-restricted-paths-rule/client/a.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              basePath: '<fixture>/fixtures/no-restricted-paths-rule',
              zones: [
                {
                  target: './client/**',
                  from: './server',
                },
              ],
            },
          ],
          root,
        ),
      errors: [
        {
          messageId: 'unexpectedPath',
          message: 'Unexpected path "../server/b" imported in restricted zone.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      name: 'run 0 invalid 4',
      code: 'import a from "../client/a"\nimport c from "./c"',
      filename: 'fixtures/no-restricted-paths-rule/server/b.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              basePath: '<fixture>/fixtures/no-restricted-paths-rule',
              zones: [
                {
                  target: './server',
                  from: './client',
                },
                {
                  target: './server',
                  from: './server/c.ts',
                },
              ],
            },
          ],
          root,
        ),
      errors: [
        {
          messageId: 'unexpectedPath',
          message: 'Unexpected path "../client/a" imported in restricted zone.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 28,
        },
        {
          messageId: 'unexpectedPath',
          message: 'Unexpected path "./c" imported in restricted zone.',
          line: 2,
          column: 15,
          endLine: 2,
          endColumn: 20,
        },
      ],
    },
    {
      name: 'run 0 invalid 5',
      code: 'const b = require("../server/b")',
      filename: 'fixtures/no-restricted-paths-rule/client/consumer.js',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              basePath: '<fixture>/fixtures/no-restricted-paths-rule',
              zones: [
                {
                  target: './client',
                  from: './server',
                },
              ],
            },
          ],
          root,
        ),
      errors: [
        {
          messageId: 'unexpectedPath',
          message: 'Unexpected path "../server/b" imported in restricted zone.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      name: 'run 0 invalid 6',
      code: 'import b from "../two/a"',
      filename: 'fixtures/no-restricted-paths-rule/server/one/a.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              basePath: '<fixture>/fixtures/no-restricted-paths-rule',
              zones: [
                {
                  target: './server/one',
                  from: './server',
                  except: ['./one'],
                },
              ],
            },
          ],
          root,
        ),
      errors: [
        {
          messageId: 'unexpectedPath',
          message: 'Unexpected path "../two/a" imported in restricted zone.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      name: 'run 0 invalid 7',
      code: 'import b from "../two/a"',
      filename: 'fixtures/no-restricted-paths-rule/server/one/a.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              basePath: '<fixture>/fixtures/no-restricted-paths-rule',
              zones: [
                {
                  target: './server/one',
                  from: './server',
                  except: ['./one'],
                  message: 'Custom message',
                },
              ],
            },
          ],
          root,
        ),
      errors: [
        {
          messageId: 'unexpectedPath',
          message:
            'Unexpected path "../two/a" imported in restricted zone. Custom message',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      name: 'run 0 invalid 8',
      code: 'import b from "../two/a"',
      filename: 'fixtures/no-restricted-paths-rule/server/one/a.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              basePath: '<fixture>/fixtures/no-restricted-paths-rule',
              zones: [
                {
                  target: './server/one',
                  from: './server',
                  except: ['../client/a'],
                },
              ],
            },
          ],
          root,
        ),
      errors: [
        {
          messageId: 'invalidExceptionPath',
          message:
            'Restricted path exceptions must be descendants of the configured `from` path for that zone.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      name: 'run 0 invalid 9',
      code: 'import A from "../two/a"',
      filename: 'fixtures/no-restricted-paths-rule/server/one/a.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              basePath: '<fixture>/fixtures/no-restricted-paths-rule',
              zones: [
                {
                  target: '**/*',
                  from: './server/**/*',
                },
              ],
            },
          ],
          root,
        ),
      errors: [
        {
          messageId: 'unexpectedPath',
          message: 'Unexpected path "../two/a" imported in restricted zone.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      name: 'run 0 invalid 10',
      code: 'import A from "../two/a"',
      filename: 'fixtures/no-restricted-paths-rule/server/one/a.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              basePath: '<fixture>/fixtures/no-restricted-paths-rule',
              zones: [
                {
                  target: '**/*',
                  from: './server/**/*',
                  except: ['a.ts'],
                },
              ],
            },
          ],
          root,
        ),
      errors: [
        {
          messageId: 'invalidExceptionGlob',
          message:
            'Restricted path exceptions must be glob patterns when `from` contains glob patterns',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      name: 'run 0 invalid 11',
      code: 'import b from "../server/b"; // 4',
      filename: 'fixtures/no-restricted-paths-rule/client/a.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              basePath: '<fixture>/fixtures/no-restricted-paths-rule',
              zones: [
                {
                  target: ['./client'],
                  from: './server',
                },
              ],
            },
          ],
          root,
        ),
      errors: [
        {
          messageId: 'unexpectedPath',
          message: 'Unexpected path "../server/b" imported in restricted zone.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      name: 'run 0 invalid 12',
      code: 'import b from "../server/b"; // 5',
      filename: 'fixtures/no-restricted-paths-rule/client/a.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              basePath: '<fixture>/fixtures/no-restricted-paths-rule',
              zones: [
                {
                  target: './client',
                  from: ['./server'],
                },
              ],
            },
          ],
          root,
        ),
      errors: [
        {
          messageId: 'unexpectedPath',
          message: 'Unexpected path "../server/b" imported in restricted zone.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      name: 'run 0 invalid 13',
      code: 'import b from "../server/b"; // 6',
      filename: 'fixtures/no-restricted-paths-rule/client/a.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              basePath: '<fixture>/fixtures/no-restricted-paths-rule',
              zones: [
                {
                  target: ['./client/one', './client'],
                  from: './server',
                },
              ],
            },
          ],
          root,
        ),
      errors: [
        {
          messageId: 'unexpectedPath',
          message: 'Unexpected path "../server/b" imported in restricted zone.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      name: 'run 0 invalid 14',
      code: 'import b from "../server/one/b"\nimport a from "../server/two/a"',
      filename: 'fixtures/no-restricted-paths-rule/client/a.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              basePath: '<fixture>/fixtures/no-restricted-paths-rule',
              zones: [
                {
                  target: './client',
                  from: ['./server/one', './server/two'],
                },
              ],
            },
          ],
          root,
        ),
      errors: [
        {
          messageId: 'unexpectedPath',
          message:
            'Unexpected path "../server/one/b" imported in restricted zone.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 32,
        },
        {
          messageId: 'unexpectedPath',
          message:
            'Unexpected path "../server/two/a" imported in restricted zone.',
          line: 2,
          column: 15,
          endLine: 2,
          endColumn: 32,
        },
      ],
    },
    {
      name: 'run 0 invalid 15',
      code: 'import b from "../server/one/b"\nimport a from "../server/two/a"',
      filename: 'fixtures/no-restricted-paths-rule/client/a.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              basePath: '<fixture>/fixtures/no-restricted-paths-rule',
              zones: [
                {
                  target: './client',
                  from: ['./server/one/*', './server/two/*'],
                },
              ],
            },
          ],
          root,
        ),
      errors: [
        {
          messageId: 'unexpectedPath',
          message:
            'Unexpected path "../server/one/b" imported in restricted zone.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 32,
        },
        {
          messageId: 'unexpectedPath',
          message:
            'Unexpected path "../server/two/a" imported in restricted zone.',
          line: 2,
          column: 15,
          endLine: 2,
          endColumn: 32,
        },
      ],
    },
    {
      name: 'run 0 invalid 16',
      code: 'import b from "../server/b"; // 7',
      filename: 'fixtures/no-restricted-paths-rule/client/a.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              basePath: '<fixture>/fixtures/no-restricted-paths-rule',
              zones: [
                {
                  target: ['./client/one', './client/**/*'],
                  from: './server',
                },
              ],
            },
          ],
          root,
        ),
      errors: [
        {
          messageId: 'unexpectedPath',
          message: 'Unexpected path "../server/b" imported in restricted zone.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      name: 'run 0 invalid 17',
      code: 'import A from "../two/a"',
      filename: 'fixtures/no-restricted-paths-rule/server/one/a.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              basePath: '<fixture>/fixtures/no-restricted-paths-rule',
              zones: [
                {
                  target: '**/*',
                  from: ['./server/**/*'],
                  except: ['a.ts'],
                },
              ],
            },
          ],
          root,
        ),
      errors: [
        {
          messageId: 'invalidExceptionGlob',
          message:
            'Restricted path exceptions must be glob patterns when `from` contains glob patterns',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      name: 'run 0 invalid 18',
      code: 'import b from "../server/one/b"',
      filename: 'fixtures/no-restricted-paths-rule/client/a.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              basePath: '<fixture>/fixtures/no-restricted-paths-rule',
              zones: [
                {
                  target: './client',
                  from: ['./server/one', './server/two/*'],
                },
              ],
            },
          ],
          root,
        ),
      errors: [
        {
          messageId: 'mixedGlobAndNonGlob',
          message:
            'Restricted path `from` must contain either only glob patterns or none',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      name: 'run 0 invalid 19',
      code: 'import type b from "../server/b"',
      filename: 'fixtures/no-restricted-paths-rule/client/a.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              basePath: '<fixture>/fixtures/no-restricted-paths-rule',
              zones: [
                {
                  target: './client',
                  from: './server',
                },
              ],
            },
          ],
          root,
        ),
      errors: [
        {
          messageId: 'unexpectedPath',
          message: 'Unexpected path "../server/b" imported in restricted zone.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      name: 'run 0 invalid 20',
      code: 'import type a from "../client/a"\nimport type c from "./c"',
      filename: 'fixtures/no-restricted-paths-rule/server/b.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              basePath: '<fixture>/fixtures/no-restricted-paths-rule',
              zones: [
                {
                  target: './server',
                  from: ['./client', './server/c.ts'],
                },
              ],
            },
          ],
          root,
        ),
      errors: [
        {
          messageId: 'unexpectedPath',
          message: 'Unexpected path "../client/a" imported in restricted zone.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 33,
        },
        {
          messageId: 'unexpectedPath',
          message: 'Unexpected path "./c" imported in restricted zone.',
          line: 2,
          column: 20,
          endLine: 2,
          endColumn: 25,
        },
      ],
    },
  ],
});

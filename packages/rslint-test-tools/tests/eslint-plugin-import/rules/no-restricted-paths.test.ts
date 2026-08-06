import { test, testFixturePath } from '../utils.js';

import { RuleTester } from '../rule-tester.js';

const ruleTester = new RuleTester();
// Rslint-Modified: we don't require rules
// const rule = require('rules/no-restricted-paths')
const rule = null as never;
// Rslint-Modified end

// Rslint-Modified: upstream resolves the zone paths against `process.cwd()`.
// The harness runs from an unrelated working directory, so every case pins
// `basePath` to the fixture root and uses zone paths relative to it.
const basePath = testFixturePath('./no-restricted-paths-rule');

function filePath(relativePath: string): string {
  return testFixturePath(`./no-restricted-paths-rule/${relativePath}`);
}
// Rslint-Modified end

ruleTester.run('no-restricted-paths', rule, {
  valid: [
    test({
      code: 'import a from "../client/a"',
      filename: filePath('server/b.ts'),
      options: [{ basePath, zones: [{ target: './server', from: './other' }] }],
    }),
    test({
      code: 'import a from "../client/a"',
      filename: filePath('server/b.ts'),
      options: [{ basePath, zones: [{ target: '**/*', from: './other' }] }],
    }),
    test({
      code: 'import a from "../client/a"',
      filename: filePath('client/b.ts'),
      options: [
        {
          basePath,
          zones: [{ target: './!(client)/**/*', from: './client/**/*' }],
        },
      ],
    }),
    test({
      // Rslint-Modified: CommonJS specifiers only take part in module
      // resolution inside JavaScript files, so the require cases run on a
      // `.js` fixture.
      code: 'const a = require("../client/a")',
      filename: filePath('server/consumer.js'),
      options: [{ basePath, zones: [{ target: './server', from: './other' }] }],
    }),
    test({
      code: 'import b from "../server/b"',
      filename: filePath('client/a.ts'),
      options: [{ basePath, zones: [{ target: './client', from: './other' }] }],
    }),
    test({
      code: 'import a from "./a"',
      filename: filePath('server/one/a.ts'),
      options: [
        {
          basePath,
          zones: [
            { target: './server/one', from: './server', except: ['./one'] },
          ],
        },
      ],
    }),
    test({
      code: 'import a from "../two/a"',
      filename: filePath('server/one/a.ts'),
      options: [
        {
          basePath,
          zones: [
            { target: './server/one', from: './server', except: ['./two'] },
          ],
        },
      ],
    }),
    test({
      code: 'import a from "../one/a"',
      filename: filePath('server/two-new/a.ts'),
      options: [
        {
          basePath,
          zones: [{ target: './server/two', from: './server', except: [] }],
        },
      ],
    }),
    test({
      code: 'import A from "../two/a"',
      filename: filePath('server/one/a.ts'),
      options: [
        {
          basePath,
          zones: [
            { target: '**/*', from: './server/**/*', except: ['**/a.ts'] },
          ],
        },
      ],
    }),

    // support of arrays for from and target
    // array with single element
    test({
      code: 'import a from "../client/a"',
      filename: filePath('server/b.ts'),
      options: [
        { basePath, zones: [{ target: ['./server'], from: './other' }] },
      ],
    }),
    test({
      code: 'import a from "../client/a"',
      filename: filePath('server/b.ts'),
      options: [
        { basePath, zones: [{ target: './server', from: ['./other'] }] },
      ],
    }),
    // array with multiple elements
    test({
      code: 'import a from "../one/a"',
      filename: filePath('server/two-new/a.ts'),
      options: [
        {
          basePath,
          zones: [
            { target: ['./server/two', './server/three'], from: './server' },
          ],
        },
      ],
    }),
    test({
      code: 'import a from "../one/a"',
      filename: filePath('server/two-new/a.ts'),
      options: [
        {
          basePath,
          zones: [
            {
              target: './server',
              from: ['./server/two', './server/three'],
              except: [],
            },
          ],
        },
      ],
    }),
    // array with multiple glob patterns in from
    test({
      code: 'import a from "../client/a"',
      filename: filePath('client/b.ts'),
      options: [
        {
          basePath,
          zones: [
            {
              target: './!(client)/**/*',
              from: ['./client/*', './client/one/*'],
            },
          ],
        },
      ],
    }),
    // array with mix of glob and non glob patterns in target
    test({
      code: 'import a from "../client/a"',
      filename: filePath('client/b.ts'),
      options: [
        {
          basePath,
          zones: [
            {
              target: ['./!(client)/**/*', './client/a/'],
              from: './client/**/*',
            },
          ],
        },
      ],
    }),

    // irrelevant function calls
    test({ code: 'notRequire("../server/b")' }),
    test({
      code: 'notRequire("../server/b")',
      filename: filePath('client/a.ts'),
      options: [
        { basePath, zones: [{ target: './client', from: './server' }] },
      ],
    }),

    // no config
    test({ code: 'require("../server/b")' }),
    test({ code: 'import b from "../server/b"' }),

    // builtin (ignore)
    test({ code: 'require("os")' }),

    // typescript: type-only imports
    test({
      code: 'import type a from "../client/a"',
      filename: filePath('server/b.ts'),
      options: [{ basePath, zones: [{ target: './server', from: './other' }] }],
    }),
    test({
      code: 'import type a from "./a"',
      filename: filePath('server/one/a.ts'),
      options: [
        {
          basePath,
          zones: [
            { target: './server/one', from: './server', except: ['./one'] },
          ],
        },
      ],
    }),
    test({ code: 'import type b from "../server/b"' }),
    test({ code: 'import type * as b from "../server/b"' }),
  ],

  invalid: [
    test({
      code: 'import b from "../server/b"; // 1',
      filename: filePath('client/a.ts'),
      options: [
        { basePath, zones: [{ target: './client', from: './server' }] },
      ],
      errors: [
        {
          message: 'Unexpected path "../server/b" imported in restricted zone.',
          line: 1,
          column: 15,
        },
      ],
    }),
    test({
      code: 'import b from "../server/b"; // 2',
      filename: filePath('client/a.ts'),
      options: [
        { basePath, zones: [{ target: './client/**/*', from: './server' }] },
      ],
      errors: [
        {
          message: 'Unexpected path "../server/b" imported in restricted zone.',
          line: 1,
          column: 15,
        },
      ],
    }),
    test({
      code: 'import b from "../server/b";',
      filename: filePath('client/a.ts'),
      options: [
        { basePath, zones: [{ target: './client/*.ts', from: './server' }] },
      ],
      errors: [
        {
          message: 'Unexpected path "../server/b" imported in restricted zone.',
          line: 1,
          column: 15,
        },
      ],
    }),
    test({
      code: 'import b from "../server/b"; // 2 ter',
      filename: filePath('client/a.ts'),
      options: [
        { basePath, zones: [{ target: './client/**', from: './server' }] },
      ],
      errors: [
        {
          message: 'Unexpected path "../server/b" imported in restricted zone.',
          line: 1,
          column: 15,
        },
      ],
    }),
    test({
      code: 'import a from "../client/a"\nimport c from "./c"',
      filename: filePath('server/b.ts'),
      options: [
        {
          basePath,
          zones: [
            { target: './server', from: './client' },
            { target: './server', from: './server/c.ts' },
          ],
        },
      ],
      errors: [
        {
          message: 'Unexpected path "../client/a" imported in restricted zone.',
          line: 1,
          column: 15,
        },
        {
          message: 'Unexpected path "./c" imported in restricted zone.',
          line: 2,
          column: 15,
        },
      ],
    }),
    test({
      code: 'const b = require("../server/b")',
      filename: filePath('client/consumer.js'),
      options: [
        { basePath, zones: [{ target: './client', from: './server' }] },
      ],
      errors: [
        {
          message: 'Unexpected path "../server/b" imported in restricted zone.',
          line: 1,
          column: 19,
        },
      ],
    }),
    test({
      code: 'import b from "../two/a"',
      filename: filePath('server/one/a.ts'),
      options: [
        {
          basePath,
          zones: [
            { target: './server/one', from: './server', except: ['./one'] },
          ],
        },
      ],
      errors: [
        {
          message: 'Unexpected path "../two/a" imported in restricted zone.',
          line: 1,
          column: 15,
        },
      ],
    }),
    test({
      code: 'import b from "../two/a"',
      filename: filePath('server/one/a.ts'),
      options: [
        {
          basePath,
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
      errors: [
        {
          message:
            'Unexpected path "../two/a" imported in restricted zone. Custom message',
          line: 1,
          column: 15,
        },
      ],
    }),
    test({
      code: 'import b from "../two/a"',
      filename: filePath('server/one/a.ts'),
      options: [
        {
          basePath,
          zones: [
            {
              target: './server/one',
              from: './server',
              except: ['../client/a'],
            },
          ],
        },
      ],
      errors: [
        {
          message:
            'Restricted path exceptions must be descendants of the configured `from` path for that zone.',
          line: 1,
          column: 15,
        },
      ],
    }),
    test({
      code: 'import A from "../two/a"',
      filename: filePath('server/one/a.ts'),
      options: [
        { basePath, zones: [{ target: '**/*', from: './server/**/*' }] },
      ],
      errors: [
        {
          message: 'Unexpected path "../two/a" imported in restricted zone.',
          line: 1,
          column: 15,
        },
      ],
    }),
    test({
      code: 'import A from "../two/a"',
      filename: filePath('server/one/a.ts'),
      options: [
        {
          basePath,
          zones: [{ target: '**/*', from: './server/**/*', except: ['a.ts'] }],
        },
      ],
      errors: [
        {
          message:
            'Restricted path exceptions must be glob patterns when `from` contains glob patterns',
          line: 1,
          column: 15,
        },
      ],
    }),

    // support of arrays for from and target
    // array with single element
    test({
      code: 'import b from "../server/b"; // 4',
      filename: filePath('client/a.ts'),
      options: [
        { basePath, zones: [{ target: ['./client'], from: './server' }] },
      ],
      errors: [
        {
          message: 'Unexpected path "../server/b" imported in restricted zone.',
          line: 1,
          column: 15,
        },
      ],
    }),
    test({
      code: 'import b from "../server/b"; // 5',
      filename: filePath('client/a.ts'),
      options: [
        { basePath, zones: [{ target: './client', from: ['./server'] }] },
      ],
      errors: [
        {
          message: 'Unexpected path "../server/b" imported in restricted zone.',
          line: 1,
          column: 15,
        },
      ],
    }),
    // array with multiple elements
    test({
      code: 'import b from "../server/b"; // 6',
      filename: filePath('client/a.ts'),
      options: [
        {
          basePath,
          zones: [{ target: ['./client/one', './client'], from: './server' }],
        },
      ],
      errors: [
        {
          message: 'Unexpected path "../server/b" imported in restricted zone.',
          line: 1,
          column: 15,
        },
      ],
    }),
    test({
      code: 'import b from "../server/one/b"\nimport a from "../server/two/a"',
      filename: filePath('client/a.ts'),
      options: [
        {
          basePath,
          zones: [
            { target: './client', from: ['./server/one', './server/two'] },
          ],
        },
      ],
      errors: [
        {
          message:
            'Unexpected path "../server/one/b" imported in restricted zone.',
          line: 1,
          column: 15,
        },
        {
          message:
            'Unexpected path "../server/two/a" imported in restricted zone.',
          line: 2,
          column: 15,
        },
      ],
    }),
    // array with multiple glob patterns in from
    test({
      code: 'import b from "../server/one/b"\nimport a from "../server/two/a"',
      filename: filePath('client/a.ts'),
      options: [
        {
          basePath,
          zones: [
            { target: './client', from: ['./server/one/*', './server/two/*'] },
          ],
        },
      ],
      errors: [
        {
          message:
            'Unexpected path "../server/one/b" imported in restricted zone.',
          line: 1,
          column: 15,
        },
        {
          message:
            'Unexpected path "../server/two/a" imported in restricted zone.',
          line: 2,
          column: 15,
        },
      ],
    }),
    // array with mix of glob and non glob patterns in target
    test({
      code: 'import b from "../server/b"; // 7',
      filename: filePath('client/a.ts'),
      options: [
        {
          basePath,
          zones: [
            { target: ['./client/one', './client/**/*'], from: './server' },
          ],
        },
      ],
      errors: [
        {
          message: 'Unexpected path "../server/b" imported in restricted zone.',
          line: 1,
          column: 15,
        },
      ],
    }),
    // configuration format
    test({
      code: 'import A from "../two/a"',
      filename: filePath('server/one/a.ts'),
      options: [
        {
          basePath,
          zones: [
            { target: '**/*', from: ['./server/**/*'], except: ['a.ts'] },
          ],
        },
      ],
      errors: [
        {
          message:
            'Restricted path exceptions must be glob patterns when `from` contains glob patterns',
          line: 1,
          column: 15,
        },
      ],
    }),
    test({
      code: 'import b from "../server/one/b"',
      filename: filePath('client/a.ts'),
      options: [
        {
          basePath,
          zones: [
            { target: './client', from: ['./server/one', './server/two/*'] },
          ],
        },
      ],
      errors: [
        {
          message:
            'Restricted path `from` must contain either only glob patterns or none',
          line: 1,
          column: 15,
        },
      ],
    }),

    // typescript: type-only imports
    test({
      code: 'import type b from "../server/b"',
      filename: filePath('client/a.ts'),
      options: [
        { basePath, zones: [{ target: './client', from: './server' }] },
      ],
      errors: [
        {
          message: 'Unexpected path "../server/b" imported in restricted zone.',
          line: 1,
          column: 20,
        },
      ],
    }),
    test({
      code: 'import type a from "../client/a"\nimport type c from "./c"',
      filename: filePath('server/b.ts'),
      options: [
        {
          basePath,
          zones: [{ target: './server', from: ['./client', './server/c.ts'] }],
        },
      ],
      errors: [
        {
          message: 'Unexpected path "../client/a" imported in restricted zone.',
          line: 1,
          column: 20,
        },
        {
          message: 'Unexpected path "./c" imported in restricted zone.',
          line: 2,
          column: 20,
        },
      ],
    }),
  ],
});

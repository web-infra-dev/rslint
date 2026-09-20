// All upstream cases from eslint-plugin-n v18.3.0, including the documentation examples.
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-unpublished-import.js
import { readFileSync } from 'node:fs';
import { RuleTester } from '../rule-tester';
const archive = readFileSync(
  new URL(
    '../../../../../internal/plugins/node/rules/no_unpublished_import/testdata/upstream.txtar',
    import.meta.url,
  ),
  'utf8',
);
const entries = archive.split(/^-- (.+) --\r?\n/m).slice(1);
if (entries.length === 0 || entries.length % 2 !== 0)
  throw new Error('Missing package fixtures');
const fixtureFiles: Record<string, string> = {};
for (let i = 0; i < entries.length; i += 2)
  fixtureFiles[entries[i]] = entries[i + 1];
new RuleTester({ fixtureFiles, languageOptions: { sourceType: 'module' } }).run(
  'no-unpublished-import',
  null,
  {
    valid: [
      {
        name: '1/test.js',
        filename: '1/test.js',
        code: "import fs from 'fs';",
      },
      {
        name: '1/test.js',
        filename: '1/test.js',
        code: "import aaa from 'aaa'; aaa();",
      },
      {
        name: '1/test.js',
        filename: '1/test.js',
        code: "import c from 'aaa/a/b/c';",
      },
      {
        name: '1/test.js',
        filename: '1/test.js',
        code: "import a from './a';",
      },
      {
        name: '1/test.js',
        filename: '1/test.js',
        code: "import a from './a.js';",
      },
      {
        name: '2/ignore1.js',
        filename: '2/ignore1.js',
        code: "import test from './test';",
      },
      {
        name: '2/ignore1.js',
        filename: '2/ignore1.js',
        code: "import bbb from 'bbb';",
      },
      {
        name: '2/ignore1.js',
        filename: '2/ignore1.js',
        code: "import c from 'bbb/a/b/c';",
      },
      {
        name: '2/ignore1.js',
        filename: '2/ignore1.js',
        code: "import ignore2 from './ignore2';",
      },
      {
        name: '3/test.js',
        filename: '3/test.js',
        code: "import a from './pub/a';",
      },
      {
        name: '3/test.js',
        filename: '3/test.js',
        code: "import test2 from './test2';",
      },
      {
        name: '3/test.js',
        filename: '3/test.js',
        code: "import aaa from 'aaa';",
      },
      {
        name: '3/test.js',
        filename: '3/test.js',
        code: "import bbb from 'bbb';",
      },
      {
        name: '3/pub/ignore1.js',
        filename: '3/pub/ignore1.js',
        code: "import bbb from 'bbb';",
      },
      {
        name: '3/pub/test.js',
        filename: '3/pub/test.js',
        code: "import p from '../package.json';",
      },
      {
        name: '3/src/pub/test.js',
        filename: '3/src/pub/test.js',
        code: "import bbb from 'bbb';",
      },
      {
        name: '3/src/pub/test.js',
        filename: '3/src/pub/test.js',
        code: "import bbb from 'bbb!foo?a=b&c=d';",
      },
      {
        code: "import noExistPackage0 from 'no-exist-package-0';",
        skip: 'Upstream <input>: the native parser and lint API require a real filename.',
      },
      {
        code: "import b from './b';",
        skip: 'Upstream <input>: the native parser and lint API require a real filename.',
      },
      {
        name: '2/test.js',
        filename: '2/test.js',
        code: "import aaa from 'aaa';",
      },
      {
        name: '2/test.js',
        filename: '2/test.js',
        code: "import a from './a';",
      },
      {
        name: '1/test.js',
        filename: '1/test.js',
        code: "import electron from 'electron';",
        options: [
          {
            allowModules: ['electron'],
          },
        ],
      },
      {
        name: 'test.js',
        filename: 'test.js',
        code: "import a from 'virtual:package-name';",
        options: [
          {
            allowModules: ['virtual:package-name'],
          },
        ],
      },
      {
        name: 'test.js',
        filename: 'test.js',
        code: "import a from 'virtual:package-scope/name';",
        options: [
          {
            allowModules: ['virtual:package-scope'],
          },
        ],
      },
      {
        name: '4/index.jsx',
        filename: '4/index.jsx',
        code: "import abc from './abc';",
        options: [
          {
            tryExtensions: ['.jsx'],
          },
        ],
      },
      {
        name: '3/src/readme.js',
        filename: '3/src/readme.js',
        code: "import bbb from 'bbb';",
      },
      {
        name: 'negative-in-files/lib/__test__/index.js',
        filename: 'negative-in-files/lib/__test__/index.js',
        code: "import bbb from 'bbb';",
      },
      {
        name: 'private-package/index.js',
        filename: 'private-package/index.js',
        code: "import bbb from 'bbb';",
      },
      {
        name: '1/test.ts',
        filename: '1/test.ts',
        code: "import type foo from 'foo';",
        options: [
          {
            ignoreTypeImport: true,
          },
        ],
      },
      {
        name: 'tsconfig-paths-wildcard/index.ts',
        filename: 'tsconfig-paths-wildcard/index.ts',
        code: "import foo from '@test/dev'",
        options: [
          {
            allowModules: ['@test/dev'],
          },
        ],
      },
    ],
    invalid: [
      {
        name: '2/test.js',
        filename: '2/test.js',
        code: "import ignore1 from './ignore1.js';",
        errors: [
          {
            messageId: 'notPublished',
            message: '"./ignore1.js" is not published.',
            line: 1,
            column: 21,
            endLine: 1,
            endColumn: 35,
          },
        ],
      },
      {
        name: '3/pub/test.js',
        filename: '3/pub/test.js',
        code: "import bbb from 'bbb';",
        errors: [
          {
            messageId: 'notPublished',
            message: '"bbb" is not published.',
            line: 1,
            column: 17,
            endLine: 1,
            endColumn: 22,
          },
        ],
      },
      {
        name: '3/pub/test.js',
        filename: '3/pub/test.js',
        code: "import ignore1 from './ignore1.js';",
        errors: [
          {
            messageId: 'notPublished',
            message: '"./ignore1.js" is not published.',
            line: 1,
            column: 21,
            endLine: 1,
            endColumn: 35,
          },
        ],
      },
      {
        name: '3/pub/test.js',
        filename: '3/pub/test.js',
        code: "import abc from './abc.json';",
        errors: [
          {
            messageId: 'notPublished',
            message: '"./abc.json" is not published.',
            line: 1,
            column: 17,
            endLine: 1,
            endColumn: 29,
          },
        ],
      },
      {
        name: '3/pub/test.js',
        filename: '3/pub/test.js',
        code: "import test from '../test';",
        errors: [
          {
            messageId: 'notPublished',
            message: '"../test" is not published.',
            line: 1,
            column: 18,
            endLine: 1,
            endColumn: 27,
          },
        ],
      },
      {
        name: '3/pub/test.js',
        filename: '3/pub/test.js',
        code: "import a from '../src/pub/a.js';",
        errors: [
          {
            messageId: 'notPublished',
            message: '"../src/pub/a.js" is not published.',
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 32,
          },
        ],
      },
      {
        name: '1/test.js',
        filename: '1/test.js',
        code: "import a from '../a.js';",
        errors: [
          {
            messageId: 'notPublished',
            message: '"../a.js" is not published.',
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 24,
          },
        ],
      },
      {
        name: '2/test.js',
        filename: '2/test.js',
        code: "import ignore1 from './ignore1.js';",
        errors: [
          {
            messageId: 'notPublished',
            message: '"./ignore1.js" is not published.',
            line: 1,
            column: 21,
            endLine: 1,
            endColumn: 35,
          },
        ],
      },
      {
        name: '3/src/test.jsx',
        filename: '3/src/test.jsx',
        code: "import a from '../test.jsx';",
        settings: {
          node: {
            convertPath: {
              'src/**/*.jsx': ['src/(.+?)\\.jsx', 'pub/$1.js'],
            },
          },
        },
        errors: [
          {
            messageId: 'notPublished',
            message: '"../test.jsx" is not published.',
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 28,
          },
        ],
      },
      {
        name: '3/src/test.jsx',
        filename: '3/src/test.jsx',
        code: "import a from '../test.jsx';",
        options: [
          {
            convertPath: {
              'src/**/*.jsx': ['src/(.+?)\\.jsx', 'pub/$1.js'],
            },
          },
        ],
        errors: [
          {
            messageId: 'notPublished',
            message: '"../test.jsx" is not published.',
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 28,
          },
        ],
      },
      {
        name: '3/src/test.jsx',
        filename: '3/src/test.jsx',
        code: "import a from '../test.jsx';",
        settings: {
          node: {
            convertPath: [
              {
                include: ['src/**/*.jsx'],
                replace: ['src/(.+?)\\.jsx', 'pub/$1.js'],
              },
            ],
          },
        },
        errors: [
          {
            messageId: 'notPublished',
            message: '"../test.jsx" is not published.',
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 28,
          },
        ],
      },
      {
        name: '3/src/test.jsx',
        filename: '3/src/test.jsx',
        code: "import a from '../test.jsx';",
        options: [
          {
            convertPath: [
              {
                include: ['src/**/*.jsx'],
                replace: ['src/(.+?)\\.jsx', 'pub/$1.js'],
              },
            ],
          },
        ],
        errors: [
          {
            messageId: 'notPublished',
            message: '"../test.jsx" is not published.',
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 28,
          },
        ],
      },
      {
        name: '4/index.jsx',
        filename: '4/index.jsx',
        code: "import abc from './abc';",
        errors: [
          {
            messageId: 'notPublished',
            message: '"./abc" is not published.',
            line: 1,
            column: 17,
            endLine: 1,
            endColumn: 24,
          },
        ],
      },
      {
        name: '1/test.js',
        filename: '1/test.js',
        code: "import a from '../2/a.js';",
        errors: [
          {
            messageId: 'notPublished',
            message: '"../2/a.js" is not published.',
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 26,
          },
        ],
      },
      {
        name: '2/test.js',
        filename: '2/test.js',
        code: "function f() { import('./ignore1.js') }",
        errors: [
          {
            messageId: 'notPublished',
            message: '"./ignore1.js" is not published.',
            line: 1,
            column: 23,
            endLine: 1,
            endColumn: 37,
          },
        ],
      },
      {
        name: '1/test.ts',
        filename: '1/test.ts',
        code: "import type foo from 'foo';",
        options: [
          {
            ignoreTypeImport: false,
          },
        ],
        errors: [
          {
            messageId: 'notPublished',
            message: '"foo" is not published.',
            line: 1,
            column: 22,
            endLine: 1,
            endColumn: 27,
          },
        ],
      },
      {
        name: '1/test.ts',
        filename: '1/test.ts',
        code: "import type foo from 'foo';",
        errors: [
          {
            messageId: 'notPublished',
            message: '"foo" is not published.',
            line: 1,
            column: 22,
            endLine: 1,
            endColumn: 27,
          },
        ],
      },
      {
        name: 'private-package/index.js',
        filename: 'private-package/index.js',
        code: "import bbb from 'bbb';",
        options: [
          {
            ignorePrivate: false,
          },
        ],
        errors: [
          {
            messageId: 'notPublished',
            message: '"bbb" is not published.',
            line: 1,
            column: 17,
            endLine: 1,
            endColumn: 22,
          },
        ],
      },
    ],
  },
);

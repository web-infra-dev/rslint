// All 63 upstream cases from eslint-plugin-n v18.3.0.
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-unpublished-require.js
import { readFileSync } from 'node:fs';
import { RuleTester } from '../rule-tester';
const archive = readFileSync(
  new URL(
    '../../../../../internal/plugins/node/rules/no_unpublished_require/testdata/upstream.txtar',
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
new RuleTester({
  fixtureFiles,
  languageOptions: {
    sourceType: 'commonjs',
    globals: { require: 'readonly', global: 'readonly' },
  },
}).run('no-unpublished-require', null, {
  valid: [
    {
      filename: '1/test.js',
      code: "require('fs');",
      name: 'upstream valid 1: 1/test.js',
    },
    {
      filename: '1/test.js',
      code: "require('aaa');",
      name: 'upstream valid 2: 1/test.js',
    },
    {
      filename: '1/test.js',
      code: "require('aaa/a/b/c');",
      name: 'upstream valid 3: 1/test.js',
    },
    {
      filename: '1/test.js',
      code: "require('./a');",
      name: 'upstream valid 4: 1/test.js',
    },
    {
      filename: '1/test.js',
      code: "require('./a.js');",
      name: 'upstream valid 5: 1/test.js',
    },
    {
      filename: '2/ignore1.js',
      code: "require('./test');",
      name: 'upstream valid 6: 2/ignore1.js',
    },
    {
      filename: '2/ignore1.js',
      code: "require('bbb');",
      name: 'upstream valid 7: 2/ignore1.js',
    },
    {
      filename: '2/ignore1.js',
      code: "require('bbb/a/b/c');",
      name: 'upstream valid 8: 2/ignore1.js',
    },
    {
      filename: '2/ignore1.js',
      code: "require('./ignore2');",
      name: 'upstream valid 9: 2/ignore1.js',
    },
    {
      filename: '3/test.js',
      code: "require('./pub/a');",
      name: 'upstream valid 10: 3/test.js',
    },
    {
      filename: '3/test.js',
      code: "require('./test2');",
      name: 'upstream valid 11: 3/test.js',
    },
    {
      filename: '3/test.js',
      code: "require('aaa');",
      name: 'upstream valid 12: 3/test.js',
    },
    {
      filename: '3/test.js',
      code: "require('bbb');",
      name: 'upstream valid 13: 3/test.js',
    },
    {
      filename: '3/pub/ignore1.js',
      code: "require('bbb');",
      name: 'upstream valid 14: 3/pub/ignore1.js',
    },
    {
      filename: '3/pub/test.js',
      code: "require('../package.json');",
      name: 'upstream valid 15: 3/pub/test.js',
    },
    {
      filename: '3/src/pub/test.js',
      code: "require('bbb');",
      name: 'upstream valid 16: 3/src/pub/test.js',
    },
    {
      filename: '3/src/pub/test.js',
      code: "require('bbb!foo?a=b&c=d');",
      name: 'upstream valid 17: 3/src/pub/test.js',
    },
    {
      filename: '3/src/test.jsx',
      code: "require('./a');",
      settings: {
        node: {
          convertPath: {
            'src/**/*.jsx': ['src/(.+?)\\.jsx', 'pub/$1.js'],
          },
          tryExtensions: ['.js', '.jsx', '.json'],
        },
      },
      name: 'upstream valid 18: 3/src/test.jsx',
    },
    {
      filename: '3/src/test.jsx',
      code: "require('./a');",
      options: [
        {
          convertPath: {
            'src/**/*.jsx': ['src/(.+?)\\.jsx', 'pub/$1.js'],
          },
          tryExtensions: ['.js', '.jsx', '.json'],
        },
      ],
      name: 'upstream valid 19: 3/src/test.jsx',
    },
    {
      filename: '3/src/test.jsx',
      code: "require('../test');",
      settings: {
        node: {
          convertPath: [
            {
              include: ['src/**/*.jsx'],
              exclude: ['**/test.jsx'],
              replace: ['src/(.+?)\\.jsx', 'pub/$1.js'],
            },
          ],
        },
      },
      name: 'upstream valid 20: 3/src/test.jsx',
    },
    {
      filename: '3/src/test.jsx',
      code: "require('../test');",
      options: [
        {
          convertPath: [
            {
              include: ['src/**/*.jsx'],
              exclude: ['**/test.jsx'],
              replace: ['src/(.+?)\\.jsx', 'pub/$1.js'],
            },
          ],
        },
      ],
      name: 'upstream valid 21: 3/src/test.jsx',
    },
    {
      filename: '1/test.js',
      code: 'require;',
      name: 'upstream valid 22: 1/test.js',
    },
    {
      filename: '1/test.js',
      code: "require('no-exist-package-0');",
      name: 'upstream valid 23: 1/test.js',
    },
    {
      code: "require('no-exist-package-0');",
      skip: 'Upstream <input>: the native parser and lint API require a real filename.',
      name: 'upstream valid 24: <input>',
    },
    {
      code: "require('./b');",
      skip: 'Upstream <input>: the native parser and lint API require a real filename.',
      name: 'upstream valid 25: <input>',
    },
    {
      filename: '1/test.js',
      code: 'require();',
      name: 'upstream valid 26: 1/test.js',
    },
    {
      filename: '1/test.js',
      code: 'require(foo);',
      name: 'upstream valid 27: 1/test.js',
    },
    {
      filename: '1/test.js',
      code: 'require(777);',
      name: 'upstream valid 28: 1/test.js',
    },
    {
      filename: '1/test.js',
      code: 'require(`foo${bar}`);',
      name: 'upstream valid 29: 1/test.js',
    },
    {
      filename: '2/test.js',
      code: "require('aaa');",
      name: 'upstream valid 30: 2/test.js',
    },
    {
      filename: '2/test.js',
      code: "require('./a');",
      name: 'upstream valid 31: 2/test.js',
    },
    {
      filename: 'issue48n/test.js',
      code: "require('.');",
      name: 'upstream valid 32: issue48n/test.js',
    },
    {
      filename: 'issue48n/test.js',
      code: "require('./');",
      name: 'upstream valid 33: issue48n/test.js',
    },
    {
      filename: 'issue48n/test/test.js',
      code: "require('..');",
      name: 'upstream valid 34: issue48n/test/test.js',
    },
    {
      filename: 'issue99/test/bin.js',
      code: "require('./index.js');",
      name: 'upstream valid 35: issue99/test/bin.js',
    },
    {
      filename: 'issue99/test/bin.js',
      code: "require('.');",
      name: 'upstream valid 36: issue99/test/bin.js',
    },
    {
      filename: 'brace-extglob/index.js',
      code: "require('./src/helper.js');",
      name: 'upstream valid 37: brace-extglob/index.js',
    },
    {
      filename: 'case-insensitive/index.js',
      code: "require('./Foo.js');",
      name: 'upstream valid 38: case-insensitive/index.js',
    },
    {
      filename: 'brace-sequence/index.js',
      code: "require('./file1.js');",
      name: 'upstream valid 39: brace-sequence/index.js',
    },
    {
      filename: '1/test.js',
      code: "require('electron');",
      options: [
        {
          allowModules: ['electron'],
        },
      ],
      name: 'upstream valid 40: 1/test.js',
    },
    {
      filename: 'test.js',
      code: "require('virtual:package-name');",
      options: [
        {
          allowModules: ['virtual:package-name'],
        },
      ],
      name: 'upstream valid 41: test.js',
    },
    {
      filename: 'test.js',
      code: "require('virtual:package-scope/name');",
      options: [
        {
          allowModules: ['virtual:package-scope'],
        },
      ],
      name: 'upstream valid 42: test.js',
    },
    {
      filename: '3/src/readme.js',
      code: "require('bbb');",
      name: 'upstream valid 43: 3/src/readme.js',
    },
    {
      filename: 'negative-in-files/lib/__test__/index.js',
      code: "require('bbb');",
      name: 'upstream valid 44: negative-in-files/lib/__test__/index.js',
    },
    {
      filename: 'issue126/lib/test.js',
      code: "require('bbb');",
      name: 'upstream valid 45: issue126/lib/test.js',
    },
    {
      filename: 'private-package/index.js',
      code: "require('bbb');",
      name: 'upstream valid 46: private-package/index.js',
    },
  ],
  invalid: [
    {
      filename: 'brace-extglob/index.js',
      code: "require('./src/helper.test.js');",
      errors: [
        {
          messageId: 'notPublished',
          message: '"./src/helper.test.js" is not published.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 31,
        },
      ],
      name: 'upstream invalid 1: brace-extglob/index.js',
    },
    {
      filename: 'extended-basename-exclusion/index.js',
      code: "require('./src/test.js');",
      errors: [
        {
          messageId: 'notPublished',
          message: '"./src/test.js" is not published.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 24,
        },
      ],
      name: 'upstream invalid 2: extended-basename-exclusion/index.js',
    },
    {
      filename: '2/test.js',
      code: "require('./ignore1.js');",
      errors: [
        {
          messageId: 'notPublished',
          message: '"./ignore1.js" is not published.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 23,
        },
      ],
      name: 'upstream invalid 3: 2/test.js',
    },
    {
      filename: '2/test.js',
      code: "require('./ignore1');",
      errors: [
        {
          messageId: 'notPublished',
          message: '"./ignore1" is not published.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 20,
        },
      ],
      name: 'upstream invalid 4: 2/test.js',
    },
    {
      filename: '3/pub/test.js',
      code: "require('bbb');",
      errors: [
        {
          messageId: 'notPublished',
          message: '"bbb" is not published.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 14,
        },
      ],
      name: 'upstream invalid 5: 3/pub/test.js',
    },
    {
      filename: '3/pub/test.js',
      code: "require('./ignore1');",
      errors: [
        {
          messageId: 'notPublished',
          message: '"./ignore1" is not published.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 20,
        },
      ],
      name: 'upstream invalid 6: 3/pub/test.js',
    },
    {
      filename: '3/pub/test.js',
      code: "require('./abc');",
      errors: [
        {
          messageId: 'notPublished',
          message: '"./abc" is not published.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 16,
        },
      ],
      name: 'upstream invalid 7: 3/pub/test.js',
    },
    {
      filename: '3/pub/test.js',
      code: "require('../test');",
      errors: [
        {
          messageId: 'notPublished',
          message: '"../test" is not published.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 18,
        },
      ],
      name: 'upstream invalid 8: 3/pub/test.js',
    },
    {
      filename: '3/pub/test.js',
      code: "require('../src/pub/a.js');",
      errors: [
        {
          messageId: 'notPublished',
          message: '"../src/pub/a.js" is not published.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 26,
        },
      ],
      name: 'upstream invalid 9: 3/pub/test.js',
    },
    {
      filename: '1/test.js',
      code: "require('../a.js');",
      errors: [
        {
          messageId: 'notPublished',
          message: '"../a.js" is not published.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 18,
        },
      ],
      name: 'upstream invalid 10: 1/test.js',
    },
    {
      filename: '3/src/test.jsx',
      code: "require('../test');",
      errors: [
        {
          messageId: 'notPublished',
          message: '"../test" is not published.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 18,
        },
      ],
      settings: {
        node: {
          convertPath: {
            'src/**/*.jsx': ['src/(.+?)\\.jsx', 'pub/$1.js'],
          },
        },
      },
      name: 'upstream invalid 11: 3/src/test.jsx',
    },
    {
      filename: '3/src/test.jsx',
      code: "require('../test');",
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
          message: '"../test" is not published.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 18,
        },
      ],
      name: 'upstream invalid 12: 3/src/test.jsx',
    },
    {
      filename: '3/src/test.jsx',
      code: "require('../test');",
      errors: [
        {
          messageId: 'notPublished',
          message: '"../test" is not published.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 18,
        },
      ],
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
      name: 'upstream invalid 13: 3/src/test.jsx',
    },
    {
      filename: '3/src/test.jsx',
      code: "require('../test');",
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
          message: '"../test" is not published.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 18,
        },
      ],
      name: 'upstream invalid 14: 3/src/test.jsx',
    },
    {
      filename: '2/test.js',
      code: "require('./ignore1');",
      errors: [
        {
          messageId: 'notPublished',
          message: '"./ignore1" is not published.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 20,
        },
      ],
      name: 'upstream invalid 15: 2/test.js',
    },
    {
      filename: '1/test.js',
      code: "require('../2/a.js');",
      errors: [
        {
          messageId: 'notPublished',
          message: '"../2/a.js" is not published.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 20,
        },
      ],
      name: 'upstream invalid 16: 1/test.js',
    },
    {
      filename: 'private-package/index.js',
      code: "require('bbb');",
      errors: [
        {
          messageId: 'notPublished',
          message: '"bbb" is not published.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 14,
        },
      ],
      options: [
        {
          ignorePrivate: false,
        },
      ],
      name: 'upstream invalid 17: private-package/index.js',
    },
  ],
});

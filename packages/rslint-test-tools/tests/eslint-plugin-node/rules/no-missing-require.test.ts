// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-missing-require.js
import { readFileSync } from 'node:fs';
import path from 'node:path';
import { expect, test } from 'rstack/test';
import { lint } from '@rslint/core/internal';
import { createTempDir, cleanupTempDir } from '../../cli/js-config/helpers';
import { RuleTester } from '../rule-tester';

const archive = readFileSync(
  new URL(
    '../../../../../internal/plugins/node/rules/no_missing_require/testdata/upstream.txtar',
    import.meta.url,
  ),
  'utf8',
);
const entries = archive.split(/^-- (.+) --\r?\n/m).slice(1);
if (entries.length === 0 || entries.length % 2 !== 0)
  throw new Error('Missing require fixtures');
const fixtureFiles: Record<string, string> = {};
for (let i = 0; i < entries.length; i += 2)
  fixtureFiles[entries[i]] = entries[i + 1];
const isCaseSensitiveFileSystem = ['linux', 'freebsd', 'openbsd'].includes(
  process.platform,
);
const languageOptions = {
  sourceType: 'commonjs' as const,
  globals: { require: 'readonly' as const },
};
const ruleTester = new RuleTester({ fixtureFiles, languageOptions });
ruleTester.run('no-missing-require', null, {
  valid: [
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('fs');",
      name: 'upstream valid 1',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('node:fs');",
      name: 'upstream valid 2',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('node:test');",
      name: 'upstream valid 3',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('eslint');",
      name: 'upstream valid 4',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('rimraf/package.json');",
      name: 'upstream valid 5',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('./a');",
      name: 'upstream valid 6',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('./a.js');",
      name: 'upstream valid 7',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('./a.config');",
      name: 'upstream valid 8',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('./a.config.js');",
      name: 'upstream valid 9',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('./b');",
      name: 'upstream valid 10',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('./b.json');",
      name: 'upstream valid 11',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('./c.coffee');",
      name: 'upstream valid 12',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('mocha');",
      name: 'upstream valid 13',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: 'require(`eslint`);',
      name: 'upstream valid 14',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('mocha!foo?a=b&c=d');",
      name: 'upstream valid 15',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('./c');",
      options: [{ tryExtensions: ['.coffee'] }],
      name: 'upstream valid 16',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('./c');",
      settings: { node: { tryExtensions: ['.coffee'] } },
      name: 'upstream valid 17',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('./fixtures/no-missing/a');",
      settings: { node: { resolvePaths: ['tests'] } },
      name: 'upstream valid 18',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('./fixtures/no-missing/a');",
      options: (root: string) => [{ resolvePaths: [root + '/tests'] }],
      name: 'upstream valid 19',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('./fixtures/no-missing/a');",
      options: [{ resolvePaths: ['tests'] }],
      name: 'upstream valid 20',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('./fixtures/no-missing/a');",
      options: [{ resolvePaths: ['scripts', 'tests'] }],
      name: 'upstream valid 21',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('./a');",
      options: [{ resolvePaths: ['tests'] }],
      name: 'upstream valid 22',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('a');",
      options: (root: string) => [
        { resolverConfig: { modules: [root + '/tests/fixtures/no-missing'] } },
      ],
      name: 'upstream valid 23',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('my-module');",
      options: (root: string) => [
        {
          resolverConfig: {
            modules: [root + '/tests/fixtures/no-missing/my_modules'],
          },
        },
      ],
      name: 'upstream valid 24',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: 'require;',
      name: 'upstream valid 25',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('no-exist-package-0');",
      languageOptions: { globals: { require: 'off' } },
      name: 'upstream valid 26',
      errors: [],
    },
    {
      code: "require('no-exist-package-0');",
      name: 'upstream valid 27',
      skip: 'Unknown <input> filenames cannot be represented by the native parser or lint API.',
      filename: 'input.js',
    },
    {
      code: "require('./b');",
      name: 'upstream valid 28',
      skip: 'Unknown <input> filenames cannot be represented by the native parser or lint API.',
      filename: 'input.js',
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: 'require();',
      name: 'upstream valid 29',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: 'require(foo);',
      name: 'upstream valid 30',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: 'require(`foo${bar}`);',
      name: 'upstream valid 31',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('eslint');",
      name: 'upstream valid 32',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('./a');",
      name: 'upstream valid 33',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('.');",
      name: 'upstream valid 34',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('./');",
      name: 'upstream valid 35',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('./foo');",
      name: 'upstream valid 36',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('./foo/');",
      name: 'upstream valid 37',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('electron');",
      options: [{ allowModules: ['electron'] }],
      name: 'upstream valid 38',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('jquery.cookie');",
      options: [{ allowModules: ['jquery.cookie'] }],
      name: 'upstream valid 39',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('virtual:package-name');",
      options: [{ allowModules: ['virtual:package-name'] }],
      name: 'upstream valid 40',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('virtual:package-scope/name');",
      options: [{ allowModules: ['virtual:package-scope'] }],
      name: 'upstream valid 41',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.tsx',
      code: "require('./d.js');",
      settings: {
        node: {
          typescriptExtensionMap: [
            ['', '.js'],
            ['.ts', '.js'],
            ['.cts', '.cjs'],
            ['.mts', '.mjs'],
            ['.tsx', '.js'],
          ],
        },
      },
      name: 'upstream valid 42',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.ts',
      code: "require('./d.js');",
      settings: {
        node: {
          typescriptExtensionMap: [
            ['', '.js'],
            ['.ts', '.js'],
            ['.cts', '.cjs'],
            ['.mts', '.mjs'],
            ['.tsx', '.js'],
          ],
        },
      },
      name: 'upstream valid 43',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.tsx',
      code: "require('./e.js');",
      settings: {
        node: {
          typescriptExtensionMap: [
            ['', '.js'],
            ['.ts', '.js'],
            ['.cts', '.cjs'],
            ['.mts', '.mjs'],
            ['.tsx', '.js'],
          ],
        },
      },
      name: 'upstream valid 44',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.ts',
      code: "require('./e.js');",
      settings: {
        node: {
          typescriptExtensionMap: [
            ['', '.js'],
            ['.ts', '.js'],
            ['.cts', '.cjs'],
            ['.mts', '.mjs'],
            ['.tsx', '.js'],
          ],
        },
      },
      name: 'upstream valid 45',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.tsx',
      code: "require('./d.js');",
      options: [
        {
          typescriptExtensionMap: [
            ['', '.js'],
            ['.ts', '.js'],
            ['.cts', '.cjs'],
            ['.mts', '.mjs'],
            ['.tsx', '.js'],
          ],
        },
      ],
      name: 'upstream valid 46',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.ts',
      code: "require('./d.js');",
      options: [
        {
          typescriptExtensionMap: [
            ['', '.js'],
            ['.ts', '.js'],
            ['.cts', '.cjs'],
            ['.mts', '.mjs'],
            ['.tsx', '.js'],
          ],
        },
      ],
      name: 'upstream valid 47',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.tsx',
      code: "require('./e.js');",
      options: [
        {
          typescriptExtensionMap: [
            ['', '.js'],
            ['.ts', '.js'],
            ['.cts', '.cjs'],
            ['.mts', '.mjs'],
            ['.tsx', '.js'],
          ],
        },
      ],
      name: 'upstream valid 48',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.ts',
      code: "require('./e.js');",
      options: [
        {
          typescriptExtensionMap: [
            ['', '.js'],
            ['.ts', '.js'],
            ['.cts', '.cjs'],
            ['.mts', '.mjs'],
            ['.tsx', '.js'],
          ],
        },
      ],
      name: 'upstream valid 49',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.tsx',
      code: "require('./e.jsx');",
      options: [{ typescriptExtensionMap: 'preserve' }],
      name: 'upstream valid 50',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.tsx',
      code: "require('./e.js');",
      options: [{ typescriptExtensionMap: 'react' }],
      name: 'upstream valid 51',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.tsx',
      code: "require('./e.jsx');",
      settings: { node: { typescriptExtensionMap: 'preserve' } },
      name: 'upstream valid 52',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.tsx',
      code: "require('./e.js');",
      settings: { node: { typescriptExtensionMap: 'react' } },
      name: 'upstream valid 53',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/ts-react/test.tsx',
      code: "require('./e.jsx');",
      options: (root: string) => [
        {
          tsconfigPath:
            root + '/tests/fixtures/no-missing/ts-preserve/tsconfig.json',
        },
      ],
      name: 'upstream valid 54',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/ts-preserve/test.tsx',
      code: "require('./e.js');",
      options: (root: string) => [
        {
          tsconfigPath:
            root + '/tests/fixtures/no-missing/ts-react/tsconfig.json',
        },
      ],
      name: 'upstream valid 55',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/ts-react/test.tsx',
      code: "require('./e.jsx');",
      settings: {
        node: {
          tsconfigPath: 'tests/fixtures/no-missing/ts-preserve/tsconfig.json',
        },
      },
      name: 'upstream valid 56',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/ts-preserve/test.tsx',
      code: "require('./e.js');",
      settings: {
        node: {
          tsconfigPath: 'tests/fixtures/no-missing/ts-react/tsconfig.json',
        },
      },
      name: 'upstream valid 57',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/ts-react/test.tsx',
      code: "require('./e.js');",
      name: 'upstream valid 58',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/ts-react/test.ts',
      code: "require('./d.js');",
      name: 'upstream valid 59',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/ts-preserve/test.tsx',
      code: "require('./e.jsx');",
      name: 'upstream valid 60',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/ts-preserve/test.ts',
      code: "require('./d.js');",
      name: 'upstream valid 61',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/ts-extends/test.tsx',
      code: "require('./e.js');",
      name: 'upstream valid 62',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/ts-extends/test.ts',
      code: "require('./d.js');",
      name: 'upstream valid 63',
      errors: [],
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require.resolve('eslint');",
      name: 'upstream valid 64',
      errors: [],
    },
    {
      name: 'documentation: existing and dynamic targets',
      filename: 'tests/fixtures/no-missing/test.js',
      code: 'var existingFile = require("./existing-file");\nvar existingModule = require("existing-module");\nvar foo = require(FOO_NAME);',
      errors: [],
    },
  ],
  invalid: [
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('no-exist-package-0');",
      errors: [
        {
          messageId: 'notFound',
          message: (root: string) =>
            "Can't resolve 'no-exist-package-0' in '" +
            root +
            "/tests/fixtures/no-missing'",
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 29,
        },
      ],
      name: 'upstream invalid 1',
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('@mysticatea/test');",
      errors: [
        {
          messageId: 'notFound',
          message: (root: string) =>
            "Can't resolve '@mysticatea/test' in '" +
            root +
            "/tests/fixtures/no-missing'",
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 27,
        },
      ],
      name: 'upstream invalid 2',
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('./c');",
      errors: [
        {
          messageId: 'notFound',
          message: (root: string) =>
            "Can't resolve './c' in '" + root + "/tests/fixtures/no-missing'",
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 14,
        },
      ],
      name: 'upstream invalid 3',
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('./d');",
      errors: [
        {
          messageId: 'notFound',
          message: (root: string) =>
            "Can't resolve './d' in '" + root + "/tests/fixtures/no-missing'",
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 14,
        },
      ],
      name: 'upstream invalid 4',
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('./a.json');",
      errors: [
        {
          messageId: 'notFound',
          message: (root: string) =>
            "Can't resolve './a.json' in '" +
            root +
            "/tests/fixtures/no-missing'",
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 19,
        },
      ],
      name: 'upstream invalid 5',
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('no-exist-package-0');",
      errors: [
        {
          messageId: 'notFound',
          message: (root: string) =>
            "Can't resolve 'no-exist-package-0' in '" +
            root +
            "/tests/fixtures/no-missing'",
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 29,
        },
      ],
      name: 'upstream invalid 6',
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('./c');",
      errors: [
        {
          messageId: 'notFound',
          message: (root: string) =>
            "Can't resolve './c' in '" + root + "/tests/fixtures/no-missing'",
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 14,
        },
      ],
      name: 'upstream invalid 7',
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('./bar');",
      errors: [
        {
          messageId: 'notFound',
          message: (root: string) =>
            "Can't resolve './bar' in '" + root + "/tests/fixtures/no-missing'",
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 16,
        },
      ],
      name: 'upstream invalid 8',
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('./bar/');",
      errors: [
        {
          messageId: 'notFound',
          message: (root: string) =>
            "Can't resolve './bar/' in '" +
            root +
            "/tests/fixtures/no-missing'",
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 17,
        },
      ],
      name: 'upstream invalid 9',
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('./A');",
      errors: [
        {
          messageId: 'notFound',
          message: (root: string) =>
            "Can't resolve './A' in '" + root + "/tests/fixtures/no-missing'",
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 14,
        },
      ],
      name: 'upstream invalid 10',
      skip: isCaseSensitiveFileSystem
        ? undefined
        : 'Case-sensitive filesystem only',
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require.resolve('no-exist-package-0');",
      errors: [
        {
          messageId: 'notFound',
          message: (root: string) =>
            "Can't resolve 'no-exist-package-0' in '" +
            root +
            "/tests/fixtures/no-missing'",
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 37,
        },
      ],
      name: 'upstream invalid 11',
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('virtual:package-name');",
      errors: [
        {
          messageId: 'notFound',
          message: (root: string) =>
            "Can't resolve 'virtual:package-name' in '" +
            root +
            "/tests/fixtures/no-missing'",
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 31,
        },
      ],
      name: 'upstream invalid 12',
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('virtual:package-scope/name');",
      errors: [
        {
          messageId: 'notFound',
          message: (root: string) =>
            "Can't resolve 'virtual:package-scope/name' in '" +
            root +
            "/tests/fixtures/no-missing'",
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 37,
        },
      ],
      name: 'upstream invalid 13',
    },
    {
      filename: 'tests/fixtures/no-missing/test.js',
      code: "require('data:text/javascript,const x = 123;');",
      errors: [
        {
          messageId: 'notFound',
          message: (root: string) =>
            "Can't resolve 'data:text/javascript,const x = 123;' in '" +
            root +
            "/tests/fixtures/no-missing'",
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 46,
        },
      ],
      name: 'upstream invalid 14',
    },
    {
      name: 'documentation: runtime error',
      filename: 'documentation/input.js',
      code: 'const foo = require("./foo");',
      errors: [
        {
          messageId: 'notFound',
          message: (root: string) =>
            "Can't resolve './foo' in '" + root + "/documentation'",
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      name: 'documentation: missing file and package',
      filename: 'documentation/input.js',
      code: 'var typoFile = require("./typo-file");\nvar typoModule = require("typo-module");',
      errors: [
        {
          messageId: 'notFound',
          message: (root: string) =>
            "Can't resolve './typo-file' in '" + root + "/documentation'",
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 37,
        },
        {
          messageId: 'notFound',
          message: (root: string) =>
            "Can't resolve 'typo-module' in '" + root + "/documentation'",
          line: 2,
          column: 26,
          endLine: 2,
          endColumn: 39,
        },
      ],
    },
  ],
});

// Upstream's second group changes cwd to the requiring file's directory.
test('node/no-missing-require: specific working directory', async () => {
  const root = await createTempDir(fixtureFiles);
  try {
    const directory = path.join(root, 'tests/fixtures/no-missing');
    const result = await lint({
      workingDirectory: directory,
      configDirectory: root,
      config: [
        {
          plugins: ['node'],
          languageOptions,
          rules: { 'node/no-missing-require': 'error' },
        },
      ],
      fileContents: {
        [path.join(directory, 'test.js')]:
          "require('../../lib/rules/no-missing-require');",
      },
    });
    expect(result.fileCount).toBe(1);
    expect(result.ruleCount).toBe(1);
    expect(result.diagnostics).toEqual([]);
  } finally {
    await cleanupTempDir(root);
  }
});

// Mirrors eslint-plugin-n v18.3.0 tests and documentation examples.
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-missing-import.js
import { existsSync, readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { RuleTester } from '../rule-tester';

const caseInsensitive = existsSync(
  fileURLToPath(import.meta.url).replace(
    'no-missing-import.test.ts',
    'NO-MISSING-IMPORT.test.ts',
  ),
);
const archive = readFileSync(
  new URL(
    '../../../../../internal/plugins/node/rules/no_missing_import/testdata/upstream.txtar',
    import.meta.url,
  ),
  'utf8',
);
const entries = archive.split(/^-- (.+) --\r?\n/m).slice(1);
if (entries.length === 0 || entries.length % 2 !== 0)
  throw new Error('Missing import fixtures');
const fixtureFiles: Record<string, string> = {};
for (let i = 0; i < entries.length; i += 2)
  fixtureFiles[entries[i]] = entries[i + 1];
const ruleTester = new RuleTester({
  fixtureFiles,
  languageOptions: { sourceType: 'module' },
});
ruleTester.run('no-missing-import', null, {
  valid: [
    {
      name: 'upstream valid 1',
      code: "import eslint from 'eslint';",
      filename: 'tests/fixtures/no-missing/test.js',
    },
    {
      name: 'upstream valid 2',
      code: "import fs from 'fs';",
      filename: 'tests/fixtures/no-missing/test.js',
    },
    {
      name: 'upstream valid 3',
      code: "import fs from 'node:fs';",
      filename: 'tests/fixtures/no-missing/test.js',
    },
    {
      name: 'upstream valid 4',
      code: "import eslint from 'eslint'",
      filename: 'tests/fixtures/no-missing/test.js',
    },
    {
      name: 'upstream valid 5',
      code: "import a from './a.js';",
      filename: 'tests/fixtures/no-missing/test.js',
    },
    {
      name: 'upstream valid 6',
      code: "import a from './d.js';",
      filename: 'tests/fixtures/no-missing/test.ts',
    },
    {
      name: 'upstream valid 7',
      code: "import aConfig from './a.config.js';",
      filename: 'tests/fixtures/no-missing/test.js',
    },
    {
      name: 'upstream valid 8',
      code: "import b from './b.json';",
      filename: 'tests/fixtures/no-missing/test.js',
    },
    {
      name: 'upstream valid 9',
      code: "import c from './c.coffee';",
      filename: 'tests/fixtures/no-missing/test.js',
    },
    {
      name: 'upstream valid 10',
      code: "import mocha from 'mocha';",
      filename: 'tests/fixtures/no-missing/test.js',
    },
    {
      name: 'upstream valid 11',
      code: "import something from 'cjs-module-with-no-main';",
      filename: 'tests/fixtures/no-missing/test.js',
    },
    {
      name: 'upstream valid 12',
      code: "import something from 'esm-module';",
      filename: 'tests/fixtures/no-missing/test.js',
    },
    {
      name: 'upstream valid 13',
      code: "import something from 'esm-module/sub';",
      filename: 'tests/fixtures/no-missing/test.js',
    },
    {
      name: 'upstream valid 14',
      code: "import mocha from 'mocha!foo?a=b&c=d';",
      filename: 'tests/fixtures/no-missing/test.js',
    },
    {
      name: 'upstream valid 15',
      code: "import a from './e.jsx';",
      filename: 'tests/fixtures/no-missing/test.tsx',
    },
    {
      name: 'upstream valid 16',
      code: "import 'misconfigured-default';",
      filename: 'tests/fixtures/no-missing/test.js',
    },
    {
      name: 'upstream valid 17',
      code: "import './c';",
      filename: 'tests/fixtures/no-missing/test.js',
      options: [
        {
          tryExtensions: ['.coffee'],
        },
      ],
    },
    {
      name: 'upstream valid 18',
      code: "import './c';",
      filename: 'tests/fixtures/no-missing/test.js',
      settings: {
        node: {
          tryExtensions: ['.coffee'],
        },
      },
    },
    {
      code: "import abc from 'no-exist-package-0';",
      skip: 'Upstream <input>: the native parser and lint API require an absolute filename.',
    },
    {
      code: "import b from './b';",
      skip: 'Upstream <input>: the native parser and lint API require an absolute filename.',
    },
    {
      name: 'upstream valid 21',
      code: 'const foo=0, bar=1; export {foo, bar};',
      filename: 'tests/fixtures/no-missing/test.js',
    },
    {
      name: 'upstream valid 22',
      code: "import eslint from 'eslint'",
      filename: 'tests/fixtures/no-missing/test.js',
    },
    {
      name: 'upstream valid 23',
      code: "import a from './a.js';",
      filename: 'tests/fixtures/no-missing/test.js',
    },
    {
      name: 'upstream valid 24',
      code: "import electron from 'electron';",
      filename: 'tests/fixtures/no-missing/test.js',
      options: [
        {
          allowModules: ['electron'],
        },
      ],
    },
    {
      name: 'upstream valid 25',
      code: "import a from 'virtual:package-name';",
      filename: 'tests/fixtures/no-missing/test.js',
      options: [
        {
          allowModules: ['virtual:package-name'],
        },
      ],
    },
    {
      name: 'upstream valid 26',
      code: "import a from 'virtual:package-scope/name';",
      filename: 'tests/fixtures/no-missing/test.js',
      options: [
        {
          allowModules: ['virtual:package-scope'],
        },
      ],
    },
    {
      name: 'upstream valid 27',
      code: "import a from './fixtures/no-missing/a.js';",
      filename: 'tests/fixtures/no-missing/test.js',
      settings: {
        node: {
          resolvePaths: ['tests'],
        },
      },
    },
    {
      name: 'upstream valid 28',
      code: "import a from './fixtures/no-missing/a.js';",
      filename: 'tests/fixtures/no-missing/test.js',
      options: [
        {
          resolvePaths: ['tests'],
        },
      ],
    },
    {
      name: 'upstream valid 29',
      code: "import a from './fixtures/no-missing/a.js';",
      filename: 'tests/fixtures/no-missing/test.js',
      options: [
        {
          resolvePaths: ['scripts', 'tests'],
        },
      ],
    },
    {
      name: 'upstream valid 30',
      code: "import d from './d.js';",
      filename: 'tests/fixtures/no-missing/test.tsx',
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
    },
    {
      name: 'upstream valid 31',
      code: "import e from './e.js';",
      filename: 'tests/fixtures/no-missing/test.tsx',
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
    },
    {
      name: 'upstream valid 32',
      code: "import d from './d.js';",
      filename: 'tests/fixtures/no-missing/test.tsx',
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
    },
    {
      name: 'upstream valid 33',
      code: "import e from './e.js';",
      filename: 'tests/fixtures/no-missing/test.tsx',
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
    },
    {
      name: 'upstream valid 34',
      code: "import e from './e.jsx';",
      filename: 'tests/fixtures/no-missing/test.tsx',
      options: [
        {
          typescriptExtensionMap: 'preserve',
        },
      ],
    },
    {
      name: 'upstream valid 35',
      code: "import e from './e.js';",
      filename: 'tests/fixtures/no-missing/test.tsx',
      options: [
        {
          typescriptExtensionMap: 'react',
        },
      ],
    },
    {
      name: 'upstream valid 36',
      code: "import e from './e.jsx';",
      filename: 'tests/fixtures/no-missing/test.tsx',
      settings: {
        node: {
          typescriptExtensionMap: 'preserve',
        },
      },
    },
    {
      name: 'upstream valid 37',
      code: "import e from './e.js';",
      filename: 'tests/fixtures/no-missing/test.tsx',
      settings: {
        node: {
          typescriptExtensionMap: 'react',
        },
      },
    },
    {
      name: 'upstream valid 38',
      code: "import e from './e.jsx';",
      filename: 'tests/fixtures/no-missing/ts-react/test.tsx',
      options: [
        {
          tsconfigPath: 'tests/fixtures/no-missing/ts-preserve/tsconfig.json',
        },
      ],
    },
    {
      name: 'upstream valid 39',
      code: "import e from './e.js';",
      filename: 'tests/fixtures/no-missing/ts-preserve/test.tsx',
      options: [
        {
          tsconfigPath: 'tests/fixtures/no-missing/ts-react/tsconfig.json',
        },
      ],
    },
    {
      name: 'upstream valid 40',
      code: "import e from './e.jsx';",
      filename: 'tests/fixtures/no-missing/ts-react/test.tsx',
      settings: {
        node: {
          tsconfigPath: 'tests/fixtures/no-missing/ts-preserve/tsconfig.json',
        },
      },
    },
    {
      name: 'upstream valid 41',
      code: "import e from './e.js';",
      filename: 'tests/fixtures/no-missing/ts-preserve/test.tsx',
      settings: {
        node: {
          tsconfigPath: 'tests/fixtures/no-missing/ts-react/tsconfig.json',
        },
      },
    },
    {
      name: 'upstream valid 42',
      code: "import e from './e.js';",
      filename: 'tests/fixtures/no-missing/ts-react/test.tsx',
    },
    {
      name: 'upstream valid 43',
      code: "import d from './d.js';",
      filename: 'tests/fixtures/no-missing/ts-react/test.ts',
    },
    {
      name: 'upstream valid 44',
      code: "import e from './e.jsx';",
      filename: 'tests/fixtures/no-missing/ts-preserve/test.tsx',
    },
    {
      name: 'upstream valid 45',
      code: "import d from './d.js';",
      filename: 'tests/fixtures/no-missing/ts-preserve/test.ts',
    },
    {
      name: 'upstream valid 46',
      code: "import e from './e.js';",
      filename: 'tests/fixtures/no-missing/ts-extends/test.tsx',
    },
    {
      name: 'upstream valid 47',
      code: "import d from './d.js';",
      filename: 'tests/fixtures/no-missing/ts-extends/test.ts',
    },
    {
      name: 'upstream valid 48',
      code: "import before from '@direct';",
      filename: 'tests/fixtures/no-missing/ts-paths/test.ts',
    },
    {
      name: 'upstream valid 49',
      code: "import before from '@wild/where.js';",
      filename: 'tests/fixtures/no-missing/ts-paths/test.ts',
    },
    {
      name: 'upstream valid 50',
      code: "import('@module');",
      filename: 'tests/fixtures/no-missing/issue-314/src/example.ts',
    },
    {
      name: 'upstream valid 51',
      code: "import type d from 'types-only';",
      filename: 'tests/fixtures/no-missing/test.ts',
    },
    {
      name: 'upstream valid 52',
      code: "import './file.ts';",
      filename: 'tests/fixtures/no-missing/ts-allow-extension/test.ts',
    },
    {
      name: 'upstream valid 53',
      code: "import plugin from 'eslint-plugin-n';",
      filename: 'tests/fixtures/no-missing/test.js',
    },
    {
      name: 'upstream valid 54',
      code: "import isIp from '#is-ip';",
      filename: 'tests/fixtures/no-missing/issue-285/test.js',
    },
    {
      name: 'upstream valid 55',
      code: "import type missing from '@type/this-does-not-exists';",
      filename: 'tests/fixtures/no-missing/test.ts',
      options: [
        {
          ignoreTypeImport: true,
        },
      ],
    },
    {
      name: 'upstream valid 56',
      code: "import 'data:text/javascript,const x = 123;';",
      filename: 'tests/fixtures/no-missing/test.js',
    },
    {
      name: 'upstream valid 57',
      code: "import 'https://example.com/module.js';",
      filename: 'tests/fixtures/no-missing/test.js',
    },
    {
      name: 'upstream valid 58',
      code: "import 'http://localhost/module.js';",
      filename: 'tests/fixtures/no-missing/test.js',
    },
    {
      name: 'upstream valid 59',
      code: 'function f() { import(foo) }',
      filename: 'tests/fixtures/no-missing/test.js',
    },
    {
      name: 'documentation: existing file and package',
      filename: 'tests/fixtures/no-missing/test.js',
      code: 'import existingFile from "./existing-file";\nimport existingModule from "existing-module";',
    },
    {
      name: 'documentation: ignoreTypeImport',
      filename: 'tests/fixtures/no-missing/test.ts',
      code: 'import type { TypeOnly } from "@types/only-types";',
      options: [
        {
          ignoreTypeImport: true,
        },
      ],
    },
  ],
  invalid: [
    {
      name: 'upstream invalid 1',
      code: "import abc from 'no-exist-package-0';",
      filename: 'tests/fixtures/no-missing/test.js',
      errors: [
        {
          messageId: 'notFound',
          message: (root: string) =>
            "Can't resolve 'no-exist-package-0' in '{{root}}/tests/fixtures/no-missing'".replaceAll(
              '{{root}}',
              root,
            ),
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      name: 'upstream invalid 2',
      code: "import abcdef from 'esm-module/sub.mjs';",
      filename: 'tests/fixtures/no-missing/test.js',
      errors: [
        {
          messageId: 'notFound',
          message: (root: string) =>
            '"./sub.mjs" is not exported under the conditions ["node","require","import"] from package {{root}}/tests/fixtures/no-missing/node_modules/esm-module (see exports field in {{root}}/tests/fixtures/no-missing/node_modules/esm-module/package.json)'.replaceAll(
              '{{root}}',
              root,
            ),
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      name: 'upstream invalid 3',
      code: "import test from '@mysticatea/test';",
      filename: 'tests/fixtures/no-missing/test.js',
      errors: [
        {
          messageId: 'notFound',
          message: (root: string) =>
            "Can't resolve '@mysticatea/test' in '{{root}}/tests/fixtures/no-missing'".replaceAll(
              '{{root}}',
              root,
            ),
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      name: 'upstream invalid 4',
      code: "import c from './c';",
      filename: 'tests/fixtures/no-missing/test.js',
      errors: [
        {
          messageId: 'notFound',
          message: (root: string) =>
            "Can't resolve './c' in '{{root}}/tests/fixtures/no-missing'".replaceAll(
              '{{root}}',
              root,
            ),
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      name: 'upstream invalid 5',
      code: "import d from './d';",
      filename: 'tests/fixtures/no-missing/test.ts',
      errors: [
        {
          messageId: 'notFound',
          message: (root: string) =>
            "Can't resolve './d' in '{{root}}/tests/fixtures/no-missing'".replaceAll(
              '{{root}}',
              root,
            ),
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      name: 'upstream invalid 6',
      code: "import d from './d';",
      filename: 'tests/fixtures/no-missing/test.js',
      errors: [
        {
          messageId: 'notFound',
          message: (root: string) =>
            "Can't resolve './d' in '{{root}}/tests/fixtures/no-missing'".replaceAll(
              '{{root}}',
              root,
            ),
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      name: 'upstream invalid 7',
      code: "import a from './a.json';",
      filename: 'tests/fixtures/no-missing/test.js',
      errors: [
        {
          messageId: 'notFound',
          message: (root: string) =>
            "Can't resolve './a.json' in '{{root}}/tests/fixtures/no-missing'".replaceAll(
              '{{root}}',
              root,
            ),
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      name: 'upstream invalid 8',
      code: "import './file.js';",
      filename: 'tests/fixtures/no-missing/ts-allow-extension/test.ts',
      errors: [
        {
          messageId: 'notFound',
          message: (root: string) =>
            "Can't resolve './file.js' in '{{root}}/tests/fixtures/no-missing/ts-allow-extension'".replaceAll(
              '{{root}}',
              root,
            ),
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      name: 'upstream invalid 9',
      code: "import eslint from 'no-exist-package-0';",
      filename: 'tests/fixtures/no-missing/test.js',
      errors: [
        {
          messageId: 'notFound',
          message: (root: string) =>
            "Can't resolve 'no-exist-package-0' in '{{root}}/tests/fixtures/no-missing'".replaceAll(
              '{{root}}',
              root,
            ),
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      name: 'upstream invalid 10',
      code: "import c from './c';",
      filename: 'tests/fixtures/no-missing/test.js',
      errors: [
        {
          messageId: 'notFound',
          message: (root: string) =>
            "Can't resolve './c' in '{{root}}/tests/fixtures/no-missing'".replaceAll(
              '{{root}}',
              root,
            ),
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      name: 'upstream invalid 11',
      code: "import a from './bar';",
      filename: 'tests/fixtures/no-missing/test.js',
      errors: [
        {
          messageId: 'notFound',
          message: (root: string) =>
            "Can't resolve './bar' in '{{root}}/tests/fixtures/no-missing'".replaceAll(
              '{{root}}',
              root,
            ),
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      name: 'upstream invalid 12',
      code: "import a from './bar/';",
      filename: 'tests/fixtures/no-missing/test.js',
      errors: [
        {
          messageId: 'notFound',
          message: (root: string) =>
            "Can't resolve './bar/' in '{{root}}/tests/fixtures/no-missing'".replaceAll(
              '{{root}}',
              root,
            ),
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      name: 'upstream invalid 13',
      code: "import a from '.';",
      filename: 'tests/fixtures/no-missing/test.js',
      errors: [
        {
          messageId: 'notFound',
          message: (root: string) =>
            "Can't resolve '.' in '{{root}}/tests/fixtures/no-missing'".replaceAll(
              '{{root}}',
              root,
            ),
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      name: 'upstream invalid 14',
      code: "import a from './';",
      filename: 'tests/fixtures/no-missing/test.js',
      errors: [
        {
          messageId: 'notFound',
          message: (root: string) =>
            "Can't resolve './' in '{{root}}/tests/fixtures/no-missing'".replaceAll(
              '{{root}}',
              root,
            ),
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      name: 'upstream invalid 15',
      code: "import a from './foo';",
      filename: 'tests/fixtures/no-missing/test.js',
      errors: [
        {
          messageId: 'notFound',
          message: (root: string) =>
            "Can't resolve './foo' in '{{root}}/tests/fixtures/no-missing'".replaceAll(
              '{{root}}',
              root,
            ),
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      name: 'upstream invalid 16',
      code: "import a from './foo/';",
      filename: 'tests/fixtures/no-missing/test.js',
      errors: [
        {
          messageId: 'notFound',
          message: (root: string) =>
            "Can't resolve './foo/' in '{{root}}/tests/fixtures/no-missing'".replaceAll(
              '{{root}}',
              root,
            ),
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      name: 'upstream invalid 17',
      skip: caseInsensitive
        ? 'Upstream skips case sensitivity on this filesystem'
        : undefined,
      code: "import a from './A.js';",
      filename: 'tests/fixtures/no-missing/test.js',
      errors: [
        {
          messageId: 'notFound',
          message: (root: string) =>
            "Can't resolve './A.js' in '{{root}}/tests/fixtures/no-missing'".replaceAll(
              '{{root}}',
              root,
            ),
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      name: 'upstream invalid 18',
      code: "function f() { import('no-exist-package-0') }",
      filename: 'tests/fixtures/no-missing/test.js',
      errors: [
        {
          messageId: 'notFound',
          message: (root: string) =>
            "Can't resolve 'no-exist-package-0' in '{{root}}/tests/fixtures/no-missing'".replaceAll(
              '{{root}}',
              root,
            ),
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 43,
        },
      ],
    },
    {
      name: 'upstream invalid 19',
      code: "import a from 'virtual:package-name';",
      filename: 'tests/fixtures/no-missing/test.js',
      errors: [
        {
          messageId: 'notFound',
          message: (root: string) =>
            "Can't resolve 'virtual:package-name' in '{{root}}/tests/fixtures/no-missing'".replaceAll(
              '{{root}}',
              root,
            ),
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      name: 'upstream invalid 20',
      code: "import a from 'virtual:package-scope/name';",
      filename: 'tests/fixtures/no-missing/test.js',
      errors: [
        {
          messageId: 'notFound',
          message: (root: string) =>
            "Can't resolve 'virtual:package-scope/name' in '{{root}}/tests/fixtures/no-missing'".replaceAll(
              '{{root}}',
              root,
            ),
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 43,
        },
      ],
    },
    {
      name: 'documentation: missing file and package',
      filename: 'tests/fixtures/no-missing/test.js',
      code: 'import typoFile from "./typo-file";\nimport typoModule from "typo-module";',
      errors: [
        {
          messageId: 'notFound',
          message: (root: string) =>
            "Can't resolve './typo-file' in '{{root}}/tests/fixtures/no-missing'".replaceAll(
              '{{root}}',
              root,
            ),
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 35,
        },
        {
          messageId: 'notFound',
          message: (root: string) =>
            "Can't resolve 'typo-module' in '{{root}}/tests/fixtures/no-missing'".replaceAll(
              '{{root}}',
              root,
            ),
          line: 2,
          column: 24,
          endLine: 2,
          endColumn: 37,
        },
      ],
    },
  ],
});

// Mirrors every case in eslint-plugin-n v18.3.0 tests/lib/rules/no-extraneous-import.js.
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-extraneous-import.js
import { readFileSync } from 'node:fs';
import { RuleTester } from '../rule-tester';

// Share the original package trees with Go; no installed workspace dependency
// can accidentally satisfy a fixture import.
const archive = readFileSync(
  new URL(
    '../../../../../internal/plugins/node/rules/no_extraneous_import/testdata/upstream.txtar',
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
const ruleTester = new RuleTester({
  fixtureFiles,
  languageOptions: { sourceType: 'module' },
});
// no-extraneous-import
ruleTester.run('no-extraneous-import', null, {
  valid: [
    {
      filename: 'dependencies/a.js',
      code: "import bbb from './bbb'",
      settings: {
        node: {
          tryExtensions: ['.ts'],
        },
      },
    },
    {
      filename: 'dependencies/a.js',
      code: "import aaa from 'aaa'",
      settings: {
        node: {
          tryExtensions: ['.ts'],
        },
      },
    },
    {
      filename: 'dependencies/a.js',
      code: "import bbb from 'aaa/bbb'",
      settings: {
        node: {
          tryExtensions: ['.ts'],
        },
      },
    },
    {
      filename: 'dependencies/a.js',
      code: "import aaa from '@bbb/aaa'",
      settings: {
        node: {
          tryExtensions: ['.ts'],
        },
      },
    },
    {
      filename: 'dependencies/a.js',
      code: "import bbb from '@bbb/aaa/bbb'",
      settings: {
        node: {
          tryExtensions: ['.ts'],
        },
      },
    },
    {
      filename: 'devDependencies/a.js',
      code: "import aaa from 'aaa'",
      settings: {
        node: {
          tryExtensions: ['.ts'],
        },
      },
    },
    {
      filename: 'peerDependencies/a.js',
      code: "import aaa from 'aaa'",
      settings: {
        node: {
          tryExtensions: ['.ts'],
        },
      },
    },
    {
      filename: 'optionalDependencies/a.js',
      code: "import aaa from 'aaa'",
      settings: {
        node: {
          tryExtensions: ['.ts'],
        },
      },
    },
    {
      filename: 'import-map/a.js',
      code: "import '#b'",
      settings: {
        node: {
          tryExtensions: ['.ts'],
        },
      },
    },
    {
      filename: 'workspace/packages/app/src/index.js',
      code: "import rootDep from 'root-dep'",
      settings: {
        node: {
          tryExtensions: ['.js', '.json', '.node', '.mjs', '.cjs'],
        },
      },
    },
    {
      filename: 'workspace/packages/app/src/index.js',
      code: "import rootDevDep from 'root-dev-dep'",
      settings: {
        node: {
          tryExtensions: ['.js', '.json', '.node', '.mjs', '.cjs'],
        },
      },
    },
    {
      filename: 'workspace-object/packages/app/src/index.js',
      code: "import rootDep from 'root-dep'",
      settings: {
        node: {
          tryExtensions: ['.js', '.json', '.node', '.mjs', '.cjs'],
        },
      },
    },
    {
      filename: 'workspace-nested/inner/packages/app/src/index.js',
      code: "import rootDep from 'root-dep'",
      settings: {
        node: {
          tryExtensions: ['.js', '.json', '.node', '.mjs', '.cjs'],
        },
      },
    },
    {
      filename: 'dependencies/a.js',
      code: "import ccc from 'ccc'",
      settings: {
        node: {
          tryExtensions: ['.ts'],
        },
      },
    },
    {
      filename: 'tsconfig-paths/index.ts',
      code: "import foo from '@configurations/foo'",
      settings: {
        node: {
          tryExtensions: ['.ts'],
        },
      },
    },
    {
      filename: 'tsconfig-paths/index.ts',
      code: "import foo from '~configurations/foo'",
      settings: {
        node: {
          tryExtensions: ['.ts'],
        },
      },
    },
    {
      filename: 'tsconfig-paths/index.ts',
      code: "import foo from '#configurations/foo'",
      settings: {
        node: {
          tryExtensions: ['.ts'],
        },
      },
    },
    {
      filename: 'test.js',
      code: "import a from 'virtual:package-name';",
      settings: {
        node: {
          tryExtensions: ['.ts'],
        },
      },
    },
    {
      filename: 'test.js',
      code: "import a from 'virtual:package-scope/name';",
      settings: {
        node: {
          tryExtensions: ['.ts'],
        },
      },
    },
  ],
  invalid: [
    {
      filename: 'dependencies/a.js',
      code: "import bbb from 'bbb'",
      errors: [
        {
          messageId: 'extraneous',
          message: '"bbb" is extraneous.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 22,
        },
      ],
      settings: {
        node: {
          tryExtensions: ['.ts'],
        },
      },
    },
    {
      filename: 'devDependencies/a.js',
      code: "import bbb from 'bbb'",
      errors: [
        {
          messageId: 'extraneous',
          message: '"bbb" is extraneous.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 22,
        },
      ],
      settings: {
        node: {
          tryExtensions: ['.ts'],
        },
      },
    },
    {
      filename: 'peerDependencies/a.js',
      code: "import bbb from 'bbb'",
      errors: [
        {
          messageId: 'extraneous',
          message: '"bbb" is extraneous.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 22,
        },
      ],
      settings: {
        node: {
          tryExtensions: ['.ts'],
        },
      },
    },
    {
      filename: 'optionalDependencies/a.js',
      code: "import bbb from 'bbb'",
      errors: [
        {
          messageId: 'extraneous',
          message: '"bbb" is extraneous.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 22,
        },
      ],
      settings: {
        node: {
          tryExtensions: ['.ts'],
        },
      },
    },
    {
      filename: 'workspace-negated/packages/excluded/src/index.js',
      code: "import rootDep from 'root-dep'",
      settings: {
        node: {
          tryExtensions: ['.js', '.json', '.node', '.mjs', '.cjs'],
        },
      },
      errors: [
        {
          messageId: 'extraneous',
          message: '"root-dep" is extraneous.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      filename: 'workspace-nested/inner/packages/app/src/index.js',
      code: "import outerDep from 'outer-dep'",
      settings: {
        node: {
          tryExtensions: ['.js', '.json', '.node', '.mjs', '.cjs'],
        },
      },
      errors: [
        {
          messageId: 'extraneous',
          message: '"outer-dep" is extraneous.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      filename: 'dependencies/a.js',
      code: "function f() { import('bbb') }",
      errors: [
        {
          messageId: 'extraneous',
          message: '"bbb" is extraneous.',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 28,
        },
      ],
      settings: {
        node: {
          tryExtensions: ['.ts'],
        },
      },
    },
  ],
});
// no-extraneous-import/typescript
ruleTester.run('no-extraneous-import', null, {
  valid: [
    {
      filename: 'typesOnly/a.ts',
      code: "import type { Glob } from 'picomatch'",
      settings: {},
    },
    {
      filename: 'typesOnly/a.ts',
      code: "export type { Glob } from 'picomatch'",
      settings: {},
    },
  ],
  invalid: [
    {
      filename: 'typesOnly/a.ts',
      code: "import { scan } from 'picomatch'",
      errors: [
        {
          messageId: 'extraneous',
          message: '"picomatch" is extraneous.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 33,
        },
      ],
      settings: {},
    },
    {
      filename: 'typesOnly/a.ts',
      code: "export { scan } from 'picomatch'",
      errors: [
        {
          messageId: 'extraneous',
          message: '"picomatch" is extraneous.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 33,
        },
      ],
      settings: {},
    },
  ],
});

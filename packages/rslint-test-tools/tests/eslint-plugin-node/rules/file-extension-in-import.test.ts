import { RuleTester } from '../rule-tester';

// All upstream cases and documentation examples, pinned to eslint-plugin-n v18.3.0.
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/file-extension-in-import.js
new RuleTester({
  languageOptions: { sourceType: 'module' },
  fixtureFiles: {
    'a.js': '',
    'b.json': '',
    'c.mjs': '',
    'd.ts': '',
    'e.tsx': '',
    'multi.js': '',
    'multi.json': '',
    'util.client.js': '// Fixture file with dot in basename (util.client.js)\n',
    'util.client.ts': '// Fixture file with dot in basename (util.client.ts)\n',
    'utils.client.ts':
      '// Fixture file with dot in basename (utils.client.ts)\nexport const foo = "bar"\n',
    'my-folder/index.js': 'export const util = () => {}\n',
    'my-things.client/index.js':
      '// Fixture: directory with dot in name (my-things.client/index.js)\n',
    'my-things.client/index.ts':
      '// Fixture: directory with dot in name (my-things.client/index.ts)\nexport const bar = "baz"\n',
    'ts-allow-extension/file.ts': '',
    'ts-allow-extension/tsconfig.json':
      '{\n    "compilerOptions": {\n        "noEmit": true,\n        "allowImportingTsExtensions": true\n    }\n}\n',
    'path/to/a/file.js': 'export default {};\n',
    'script.js': 'export default {};\n',
    'styles.css': 'body {}\n',
    'logo.png': 'fixture\n',
  },
}).run(
  'file-extension-in-import',
  {},
  {
    valid: [
      {
        filename: 'test.js',
        code: "import 'eslint'",
        name: 'upstream valid 1',
      },
      {
        filename: 'test.js',
        code: "import '@typescript-eslint/parser'",
        name: 'upstream valid 2',
      },
      {
        filename: 'test.js',
        code: "import '@typescript-eslint\\parser'",
        name: 'upstream valid 3',
      },
      {
        filename: 'test.js',
        code: "import 'punycode/'",
        name: 'upstream valid 4',
      },
      {
        filename: 'test.js',
        code: "import 'xxx'",
        name: 'upstream valid 5',
      },
      {
        filename: 'test.js',
        code: "import './a.js'",
        name: 'upstream valid 6',
      },
      {
        filename: 'test.js',
        code: "import './b.json'",
        name: 'upstream valid 7',
      },
      {
        filename: 'test.js',
        code: "import './c.mjs'",
        name: 'upstream valid 8',
      },
      {
        filename: 'test.js',
        code: "import './d.js'",
        name: 'upstream valid 9',
      },
      {
        filename: 'test.ts',
        code: "import './a.js'",
        name: 'upstream valid 10',
      },
      {
        filename: 'test.ts',
        code: "import './d.js'",
        name: 'upstream valid 11',
      },
      {
        filename: 'test.js',
        code: "import './a.js'",
        options: ['always'],
        name: 'upstream valid 12',
      },
      {
        filename: 'test.js',
        code: "import './b.json'",
        options: ['always'],
        name: 'upstream valid 13',
      },
      {
        filename: 'test.js',
        code: "import './c.mjs'",
        options: ['always'],
        name: 'upstream valid 14',
      },
      {
        filename: 'test.tsx',
        code: "import './d.jsx'",
        options: ['always'],
        name: 'upstream valid 15',
      },
      {
        filename: 'test.js',
        code: "import './a'",
        options: ['never'],
        name: 'upstream valid 16',
      },
      {
        filename: 'test.js',
        code: "import './b'",
        options: ['never'],
        name: 'upstream valid 17',
      },
      {
        filename: 'test.js',
        code: "import './c'",
        options: ['never'],
        name: 'upstream valid 18',
      },
      {
        filename: 'test.js',
        code: "import './a'",
        options: [
          'always',
          {
            '.js': 'never',
          },
        ],
        name: 'upstream valid 19',
      },
      {
        filename: 'test.js',
        code: "import './b.json'",
        options: [
          'always',
          {
            '.js': 'never',
          },
        ],
        name: 'upstream valid 20',
      },
      {
        filename: 'test.js',
        code: "import './c.mjs'",
        options: [
          'always',
          {
            '.js': 'never',
          },
        ],
        name: 'upstream valid 21',
      },
      {
        filename: 'test.js',
        code: "import './a'",
        options: [
          'never',
          {
            '.json': 'always',
          },
        ],
        name: 'upstream valid 22',
      },
      {
        filename: 'test.js',
        code: "import './b.json'",
        options: [
          'never',
          {
            '.json': 'always',
          },
        ],
        name: 'upstream valid 23',
      },
      {
        filename: 'test.js',
        code: "import './c'",
        options: [
          'never',
          {
            '.json': 'always',
          },
        ],
        name: 'upstream valid 24',
      },
      {
        filename: 'test.js',
        code: "import '@apollo/client/core'",
        options: ['always'],
        name: 'upstream valid 25',
      },
      {
        filename: 'test.js',
        code: "import 'yargs/helpers'",
        options: ['always'],
        name: 'upstream valid 26',
      },
      {
        filename: 'test.js',
        code: "import 'firebase-functions/v1/auth'",
        options: ['always'],
        name: 'upstream valid 27',
      },
      {
        filename: 'test.tsx',
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
        name: 'upstream valid 28',
      },
      {
        filename: 'test.tsx',
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
        name: 'upstream valid 29',
      },
      {
        filename: 'test.ts',
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
        name: 'upstream valid 30',
      },
      {
        filename: 'test.ts',
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
        name: 'upstream valid 31',
      },
      {
        filename: 'ts-allow-extension/test.ts',
        code: "require('./file.js');",
        name: 'upstream valid 32',
      },
      {
        filename: 'ts-allow-extension/test.ts',
        code: "require('./file.ts');",
        name: 'upstream valid 33',
      },
      {
        name: 'documentation always correct',
        code: 'import eslint from "eslint"\nimport foo from "./path/to/a/file.js"',
        options: ['always'],
        filename: 'test.js',
      },
      {
        name: 'documentation never correct',
        code: 'import eslint from "eslint"\nimport foo from "./path/to/a/file"',
        options: ['never'],
        filename: 'test.js',
      },
      {
        name: 'documentation overrides',
        code: 'import eslint from "eslint"\nimport script from "./script"\nimport styles from "./styles.css"\nimport logo from "./logo.png"',
        options: [
          'always',
          {
            '.js': 'never',
          },
        ],
        filename: 'test.js',
      },
    ],
    invalid: [
      {
        filename: 'test.js',
        code: "import './a'",
        output: "import './a.js'",
        errors: [
          {
            messageId: 'requireExt',
            message: "require file extension '.js'.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 13,
            fix: {
              range: [11, 11],
              text: '.js',
            },
          },
        ],
        name: 'upstream invalid 1',
      },
      {
        filename: 'test.ts',
        code: "import './a'",
        output: "import './a.js'",
        errors: [
          {
            messageId: 'requireExt',
            message: "require file extension '.js'.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 13,
            fix: {
              range: [11, 11],
              text: '.js',
            },
          },
        ],
        name: 'upstream invalid 2',
      },
      {
        filename: 'test.ts',
        code: "import './d'",
        output: "import './d.js'",
        errors: [
          {
            messageId: 'requireExt',
            message: "require file extension '.js'.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 13,
            fix: {
              range: [11, 11],
              text: '.js',
            },
          },
        ],
        name: 'upstream invalid 3',
      },
      {
        filename: 'test.js',
        code: "import { util } from './my-folder'",
        output: "import { util } from './my-folder/index.js'",
        errors: [
          {
            messageId: 'requireExt',
            message: "require file extension '.js'.",
            line: 1,
            column: 22,
            endLine: 1,
            endColumn: 35,
            fix: {
              range: [33, 33],
              text: '/index.js',
            },
          },
        ],
        name: 'upstream invalid 4',
      },
      {
        filename: 'test.js',
        code: "import { util } from './my-folder/'",
        output: "import { util } from './my-folder/index.js'",
        errors: [
          {
            messageId: 'requireExt',
            message: "require file extension '.js'.",
            line: 1,
            column: 22,
            endLine: 1,
            endColumn: 36,
            fix: {
              range: [34, 34],
              text: 'index.js',
            },
          },
        ],
        name: 'upstream invalid 5',
      },
      {
        filename: 'test.js',
        code: "import './b'",
        output: "import './b.json'",
        errors: [
          {
            messageId: 'requireExt',
            message: "require file extension '.json'.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 13,
            fix: {
              range: [11, 11],
              text: '.json',
            },
          },
        ],
        name: 'upstream invalid 6',
      },
      {
        filename: 'test.js',
        code: "import './c'",
        output: "import './c.mjs'",
        errors: [
          {
            messageId: 'requireExt',
            message: "require file extension '.mjs'.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 13,
            fix: {
              range: [11, 11],
              text: '.mjs',
            },
          },
        ],
        name: 'upstream invalid 7',
      },
      {
        filename: 'test.js',
        code: "import './a'",
        output: "import './a.js'",
        options: ['always'],
        errors: [
          {
            messageId: 'requireExt',
            message: "require file extension '.js'.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 13,
            fix: {
              range: [11, 11],
              text: '.js',
            },
          },
        ],
        name: 'upstream invalid 8',
      },
      {
        filename: 'test.js',
        code: "import './b'",
        output: "import './b.json'",
        options: ['always'],
        errors: [
          {
            messageId: 'requireExt',
            message: "require file extension '.json'.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 13,
            fix: {
              range: [11, 11],
              text: '.json',
            },
          },
        ],
        name: 'upstream invalid 9',
      },
      {
        filename: 'test.js',
        code: "import './c'",
        output: "import './c.mjs'",
        options: ['always'],
        errors: [
          {
            messageId: 'requireExt',
            message: "require file extension '.mjs'.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 13,
            fix: {
              range: [11, 11],
              text: '.mjs',
            },
          },
        ],
        name: 'upstream invalid 10',
      },
      {
        filename: 'test.js',
        code: "import './a.js'",
        output: "import './a'",
        options: ['never'],
        errors: [
          {
            messageId: 'forbidExt',
            message: "forbid file extension '.js'.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 16,
            fix: {
              range: [11, 14],
              text: '',
            },
          },
        ],
        name: 'upstream invalid 11',
      },
      {
        filename: 'test.js',
        code: "import './b.json'",
        output: "import './b'",
        options: ['never'],
        errors: [
          {
            messageId: 'forbidExt',
            message: "forbid file extension '.json'.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 18,
            fix: {
              range: [11, 16],
              text: '',
            },
          },
        ],
        name: 'upstream invalid 12',
      },
      {
        filename: 'test.js',
        code: "import './c.mjs'",
        output: "import './c'",
        options: ['never'],
        errors: [
          {
            messageId: 'forbidExt',
            message: "forbid file extension '.mjs'.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 17,
            fix: {
              range: [11, 15],
              text: '',
            },
          },
        ],
        name: 'upstream invalid 13',
      },
      {
        filename: 'test.js',
        code: "import './a.js'",
        output: "import './a'",
        options: [
          'always',
          {
            '.js': 'never',
          },
        ],
        errors: [
          {
            messageId: 'forbidExt',
            message: "forbid file extension '.js'.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 16,
            fix: {
              range: [11, 14],
              text: '',
            },
          },
        ],
        name: 'upstream invalid 14',
      },
      {
        filename: 'test.js',
        code: "import './b'",
        output: "import './b.json'",
        options: [
          'always',
          {
            '.js': 'never',
          },
        ],
        errors: [
          {
            messageId: 'requireExt',
            message: "require file extension '.json'.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 13,
            fix: {
              range: [11, 11],
              text: '.json',
            },
          },
        ],
        name: 'upstream invalid 15',
      },
      {
        filename: 'test.js',
        code: "import './c'",
        output: "import './c.mjs'",
        options: [
          'always',
          {
            '.js': 'never',
          },
        ],
        errors: [
          {
            messageId: 'requireExt',
            message: "require file extension '.mjs'.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 13,
            fix: {
              range: [11, 11],
              text: '.mjs',
            },
          },
        ],
        name: 'upstream invalid 16',
      },
      {
        filename: 'test.js',
        code: "import './a.js'",
        output: "import './a'",
        options: [
          'never',
          {
            '.json': 'always',
          },
        ],
        errors: [
          {
            messageId: 'forbidExt',
            message: "forbid file extension '.js'.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 16,
            fix: {
              range: [11, 14],
              text: '',
            },
          },
        ],
        name: 'upstream invalid 17',
      },
      {
        filename: 'test.js',
        code: "import './b'",
        output: "import './b.json'",
        options: [
          'never',
          {
            '.json': 'always',
          },
        ],
        errors: [
          {
            messageId: 'requireExt',
            message: "require file extension '.json'.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 13,
            fix: {
              range: [11, 11],
              text: '.json',
            },
          },
        ],
        name: 'upstream invalid 18',
      },
      {
        filename: 'test.js',
        code: "import './c.mjs'",
        output: "import './c'",
        options: [
          'never',
          {
            '.json': 'always',
          },
        ],
        errors: [
          {
            messageId: 'forbidExt',
            message: "forbid file extension '.mjs'.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 17,
            fix: {
              range: [11, 15],
              text: '',
            },
          },
        ],
        name: 'upstream invalid 19',
      },
      {
        filename: 'test.js',
        code: "import './multi'",
        output: "import './multi.js'",
        options: ['always'],
        errors: [
          {
            messageId: 'requireExt',
            message: "require file extension '.js'.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 17,
            fix: {
              range: [15, 15],
              text: '.js',
            },
          },
        ],
        name: 'upstream invalid 20',
      },
      {
        filename: 'test.js',
        code: "import './multi.js'",
        options: ['never'],
        errors: [
          {
            messageId: 'forbidExt',
            message: "forbid file extension '.js'.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 20,
          },
        ],
        name: 'upstream invalid 21',
      },
      {
        filename: 'test.js',
        code: "import './multi.json'",
        options: ['never'],
        errors: [
          {
            messageId: 'forbidExt',
            message: "forbid file extension '.json'.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 22,
          },
        ],
        name: 'upstream invalid 22',
      },
      {
        filename: 'test.ts',
        code: "import './utils.client'",
        output: "import './utils.client.js'",
        errors: [
          {
            messageId: 'requireExt',
            message: "require file extension '.js'.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 24,
            fix: {
              range: [22, 22],
              text: '.js',
            },
          },
        ],
        name: 'upstream invalid 23',
      },
      {
        filename: 'test.ts',
        code: "import './util.client'",
        output: "import './util.client.js'",
        errors: [
          {
            messageId: 'requireExt',
            message: "require file extension '.js'.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 23,
            fix: {
              range: [21, 21],
              text: '.js',
            },
          },
        ],
        name: 'upstream invalid 24',
      },
      {
        filename: 'test.js',
        code: "import './util.client'",
        output: "import './util.client.js'",
        errors: [
          {
            messageId: 'requireExt',
            message: "require file extension '.js'.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 23,
            fix: {
              range: [21, 21],
              text: '.js',
            },
          },
        ],
        name: 'upstream invalid 25',
      },
      {
        filename: 'test.ts',
        code: "import './my-things.client'",
        output: "import './my-things.client/index.js'",
        errors: [
          {
            messageId: 'requireExt',
            message: "require file extension '.js'.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 28,
            fix: {
              range: [26, 26],
              text: '/index.js',
            },
          },
        ],
        name: 'upstream invalid 26',
      },
      {
        filename: 'test.js',
        code: "import './my-things.client'",
        output: "import './my-things.client/index.js'",
        errors: [
          {
            messageId: 'requireExt',
            message: "require file extension '.js'.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 28,
            fix: {
              range: [26, 26],
              text: '/index.js',
            },
          },
        ],
        name: 'upstream invalid 27',
      },
      {
        filename: 'test.js',
        code: "function f() { import('./a') }",
        output: "function f() { import('./a.js') }",
        errors: [
          {
            messageId: 'requireExt',
            message: "require file extension '.js'.",
            line: 1,
            column: 23,
            endLine: 1,
            endColumn: 28,
            fix: {
              range: [26, 26],
              text: '.js',
            },
          },
        ],
        name: 'upstream invalid 28',
      },
      {
        filename: 'test.js',
        code: "function f() { import('./a.js') }",
        output: "function f() { import('./a') }",
        options: ['never'],
        errors: [
          {
            messageId: 'forbidExt',
            message: "forbid file extension '.js'.",
            line: 1,
            column: 23,
            endLine: 1,
            endColumn: 31,
            fix: {
              range: [26, 29],
              text: '',
            },
          },
        ],
        name: 'upstream invalid 29',
      },
      {
        name: 'documentation introduction',
        code: 'import foo from "./path/to/a/file"\nexport * from "./path/to/a/file"',
        output:
          'import foo from "./path/to/a/file.js"\nexport * from "./path/to/a/file.js"',
        filename: 'test.js',
        errors: [
          {
            messageId: 'requireExt',
            message: "require file extension '.js'.",
            line: 1,
            column: 17,
            endLine: 1,
            endColumn: 35,
            fix: {
              range: [33, 33],
              text: '.js',
            },
          },
          {
            messageId: 'requireExt',
            message: "require file extension '.js'.",
            line: 2,
            column: 15,
            endLine: 2,
            endColumn: 33,
            fix: {
              range: [66, 66],
              text: '.js',
            },
          },
        ],
      },
      {
        name: 'documentation always incorrect',
        code: 'import foo from "./path/to/a/file"',
        options: ['always'],
        output: 'import foo from "./path/to/a/file.js"',
        filename: 'test.js',
        errors: [
          {
            messageId: 'requireExt',
            message: "require file extension '.js'.",
            line: 1,
            column: 17,
            endLine: 1,
            endColumn: 35,
            fix: {
              range: [33, 33],
              text: '.js',
            },
          },
        ],
      },
      {
        name: 'documentation never incorrect',
        code: 'import foo from "./path/to/a/file.js"',
        options: ['never'],
        output: 'import foo from "./path/to/a/file"',
        filename: 'test.js',
        errors: [
          {
            messageId: 'forbidExt',
            message: "forbid file extension '.js'.",
            line: 1,
            column: 17,
            endLine: 1,
            endColumn: 38,
            fix: {
              range: [33, 36],
              text: '',
            },
          },
        ],
      },
    ],
  },
);

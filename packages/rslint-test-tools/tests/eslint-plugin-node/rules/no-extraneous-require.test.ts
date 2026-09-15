// All upstream cases from eslint-plugin-n v18.3.0:
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-extraneous-require.js
import { RuleTester } from '../rule-tester';

const tester = new RuleTester({
  languageOptions: {
    sourceType: 'commonjs',
    globals: { require: 'readonly', global: 'readonly' },
  },
  fixtureFiles: {
    'package.json': '{"name":"fixture-root","private":true}\n',
    'dependencies/node_modules/@bbb/aaa.js': '',
    'dependencies/node_modules/aaa.js': '',
    'dependencies/node_modules/bbb/index.js': '',
    'dependencies/node_modules/bbb/package.json':
      '{\n  "name": "bbb",\n  "main": "index.js"\n}',
    'dependencies/package.json':
      '{\n    "name": "test",\n    "version": "0.0.0",\n    "dependencies": {\n        "aaa": "0.0.0",\n        "@bbb/aaa": "0.0.0"\n    }\n}\n',
    'devDependencies/node_modules/@bbb/aaa.js': '',
    'devDependencies/node_modules/aaa.js': '',
    'devDependencies/node_modules/bbb/index.js': '',
    'devDependencies/node_modules/bbb/package.json':
      '{\n  "name": "bbb",\n  "main": "index.js"\n}',
    'devDependencies/package.json':
      '{\n    "name": "test",\n    "version": "0.0.0",\n    "devDependencies": {\n        "aaa": "0.0.0",\n        "@bbb/aaa": "0.0.0"\n    }\n}\n',
    'optionalDependencies/node_modules/@bbb/aaa.js': '',
    'optionalDependencies/node_modules/aaa.js': '',
    'optionalDependencies/node_modules/bbb/index.js': '',
    'optionalDependencies/node_modules/bbb/package.json':
      '{\n  "name": "bbb",\n  "main": "index.js"\n}',
    'optionalDependencies/package.json':
      '{\n    "name": "test",\n    "version": "0.0.0",\n    "optionalDependencies": {\n        "aaa": "0.0.0",\n        "@bbb/aaa": "0.0.0"\n    }\n}\n',
    'peerDependencies/node_modules/@bbb/aaa.js': '',
    'peerDependencies/node_modules/aaa.js': '',
    'peerDependencies/node_modules/bbb/index.js': '',
    'peerDependencies/node_modules/bbb/package.json':
      '{\n  "name": "bbb",\n  "main": "index.js"\n}',
    'peerDependencies/package.json':
      '{\n    "name": "test",\n    "version": "0.0.0",\n    "peerDependencies": {\n        "aaa": "0.0.0",\n        "@bbb/aaa": "0.0.0"\n    }\n}\n',
    'workspace/node_modules/root-dep/index.js': 'export default {}\n',
    'workspace/node_modules/root-dep/package.json':
      '{\n    "name": "root-dep",\n    "version": "0.0.0"\n}\n',
    'workspace/node_modules/root-dev-dep/index.js': 'export default {}\n',
    'workspace/node_modules/root-dev-dep/package.json':
      '{\n    "name": "root-dev-dep",\n    "version": "0.0.0"\n}\n',
    'workspace/package.json':
      '{\n    "name": "workspace-root",\n    "version": "0.0.0",\n    "private": true,\n    "workspaces": [\n        "./{packages,apps}/*/"\n    ],\n    "dependencies": {\n        "root-dep": "0.0.0"\n    },\n    "devDependencies": {\n        "root-dev-dep": "0.0.0"\n    }\n}\n',
    'workspace/packages/app/package.json':
      '{\n    "name": "workspace-app",\n    "version": "0.0.0"\n}\n',
    'workspace/packages/app/src/index.js':
      'import rootDep from "root-dep"\n\nrootDep()\n',
    'workspace-negated/node_modules/root-dep/index.js': 'export default {}\n',
    'workspace-negated/node_modules/root-dep/package.json':
      '{\n    "name": "root-dep",\n    "version": "0.0.0"\n}\n',
    'workspace-negated/package.json':
      '{\n    "name": "workspace-negated-root",\n    "version": "0.0.0",\n    "private": true,\n    "workspaces": [\n        "packages/*",\n        "!packages/excluded"\n    ],\n    "dependencies": {\n        "root-dep": "0.0.0"\n    }\n}\n',
    'workspace-negated/packages/excluded/package.json':
      '{\n    "name": "workspace-excluded-app",\n    "version": "0.0.0"\n}\n',
    'workspace-negated/packages/excluded/src/index.js':
      'import rootDep from "root-dep"\n\nrootDep()\n',
    'workspace-nested/inner/node_modules/root-dep/index.js':
      'export default {}\n',
    'workspace-nested/inner/node_modules/root-dep/package.json':
      '{\n    "name": "root-dep",\n    "version": "0.0.0"\n}\n',
    'workspace-nested/inner/package.json':
      '{\n    "name": "workspace-inner-root",\n    "version": "0.0.0",\n    "private": true,\n    "workspaces": [\n        "packages/*"\n    ],\n    "dependencies": {\n        "root-dep": "0.0.0"\n    }\n}\n',
    'workspace-nested/inner/packages/app/package.json':
      '{\n    "name": "workspace-nested-app",\n    "version": "0.0.0"\n}\n',
    'workspace-nested/inner/packages/app/src/index.js':
      'import rootDep from "root-dep"\n\nrootDep()\n',
    'workspace-nested/node_modules/outer-dep/index.js': 'export default {}\n',
    'workspace-nested/node_modules/outer-dep/package.json':
      '{\n    "name": "outer-dep",\n    "version": "0.0.0"\n}\n',
    'workspace-nested/package.json':
      '{\n    "name": "workspace-outer-root",\n    "version": "0.0.0",\n    "private": true,\n    "workspaces": [\n        "**"\n    ],\n    "dependencies": {\n        "outer-dep": "0.0.0"\n    }\n}\n',
    'workspace-object/node_modules/root-dep/index.js': 'export default {}\n',
    'workspace-object/node_modules/root-dep/package.json':
      '{\n    "name": "root-dep",\n    "version": "0.0.0"\n}\n',
    'workspace-object/package.json':
      '{\n    "name": "workspace-object-root",\n    "version": "0.0.0",\n    "private": true,\n    "workspaces": {\n        "packages": [\n            "@(packages|apps)/*"\n        ]\n    },\n    "dependencies": {\n        "root-dep": "0.0.0"\n    }\n}\n',
    'workspace-object/packages/app/package.json':
      '{\n    "name": "workspace-object-app",\n    "version": "0.0.0"\n}\n',
    'workspace-object/packages/app/src/index.js':
      'import rootDep from "root-dep"\n\nrootDep()\n',
  },
});

tester.run(
  'no-extraneous-require',
  {},
  {
    valid: [
      {
        filename: 'dependencies/a.js',
        code: "$.require('bbb')",
        errors: [],
      },
      {
        filename: 'dependencies/a.js',
        code: "require('./bbb')",
        errors: [],
      },
      {
        filename: 'dependencies/a.js',
        code: 'require(bbb)',
        errors: [],
      },
      {
        filename: 'dependencies/a.js',
        code: "require('aaa')",
        errors: [],
      },
      {
        filename: 'dependencies/a.js',
        code: "require('aaa/bbb')",
        errors: [],
      },
      {
        filename: 'dependencies/a.js',
        code: "require('@bbb/aaa')",
        errors: [],
      },
      {
        filename: 'dependencies/a.js',
        code: "require('@bbb/aaa/bbb')",
        errors: [],
      },
      {
        filename: 'devDependencies/a.js',
        code: "require('aaa')",
        errors: [],
      },
      {
        filename: 'peerDependencies/a.js',
        code: "require('aaa')",
        errors: [],
      },
      {
        filename: 'optionalDependencies/a.js',
        code: "require('aaa')",
        errors: [],
      },
      {
        filename: 'workspace/packages/app/src/index.js',
        code: "require('root-dep')",
        settings: {
          n: {
            tryExtensions: ['.js', '.json', '.node', '.mjs', '.cjs'],
          },
        },
        errors: [],
      },
      {
        filename: 'workspace-object/packages/app/src/index.js',
        code: "require('root-dep')",
        settings: {
          n: {
            tryExtensions: ['.js', '.json', '.node', '.mjs', '.cjs'],
          },
        },
        errors: [],
      },
      {
        filename: 'workspace-nested/inner/packages/app/src/index.js',
        code: "require('root-dep')",
        settings: {
          n: {
            tryExtensions: ['.js', '.json', '.node', '.mjs', '.cjs'],
          },
        },
        errors: [],
      },
      {
        filename: 'dependencies/a.js',
        code: "require('ccc')",
        errors: [],
      },
      {
        filename: 'test.js',
        code: "require('virtual:package-name');",
        errors: [],
      },
      {
        filename: 'test.js',
        code: "require('virtual:package-scope/name');",
        errors: [],
      },
    ],
    invalid: [
      {
        filename: 'dependencies/a.js',
        code: "require('bbb')",
        errors: [
          {
            messageId: 'extraneous',
            message: '"bbb" is extraneous.',
            line: 1,
            column: 9,
            endLine: 1,
            endColumn: 14,
          },
        ],
      },
      {
        filename: 'devDependencies/a.js',
        code: "require('bbb')",
        errors: [
          {
            messageId: 'extraneous',
            message: '"bbb" is extraneous.',
            line: 1,
            column: 9,
            endLine: 1,
            endColumn: 14,
          },
        ],
      },
      {
        filename: 'peerDependencies/a.js',
        code: "require('bbb')",
        errors: [
          {
            messageId: 'extraneous',
            message: '"bbb" is extraneous.',
            line: 1,
            column: 9,
            endLine: 1,
            endColumn: 14,
          },
        ],
      },
      {
        filename: 'optionalDependencies/a.js',
        code: "require('bbb')",
        errors: [
          {
            messageId: 'extraneous',
            message: '"bbb" is extraneous.',
            line: 1,
            column: 9,
            endLine: 1,
            endColumn: 14,
          },
        ],
      },
      {
        filename: 'workspace-negated/packages/excluded/src/index.js',
        code: "require('root-dep')",
        settings: {
          n: {
            tryExtensions: ['.js', '.json', '.node', '.mjs', '.cjs'],
          },
        },
        errors: [
          {
            messageId: 'extraneous',
            message: '"root-dep" is extraneous.',
            line: 1,
            column: 9,
            endLine: 1,
            endColumn: 19,
          },
        ],
      },
      {
        filename: 'workspace-nested/inner/packages/app/src/index.js',
        code: "require('outer-dep')",
        settings: {
          n: {
            tryExtensions: ['.js', '.json', '.node', '.mjs', '.cjs'],
          },
        },
        errors: [
          {
            messageId: 'extraneous',
            message: '"outer-dep" is extraneous.',
            line: 1,
            column: 9,
            endLine: 1,
            endColumn: 20,
          },
        ],
      },
      {
        filename: 'dependencies/a.js',
        code: "require('b'+'bb')",
        errors: [
          {
            messageId: 'extraneous',
            message: '"bbb" is extraneous.',
            line: 1,
            column: 9,
            endLine: 1,
            endColumn: 17,
          },
        ],
      },
    ],
  },
);

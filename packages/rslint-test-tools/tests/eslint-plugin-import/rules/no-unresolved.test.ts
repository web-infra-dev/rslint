import { checkSkip } from '../../../../rslint-rule-tester/src/cases.ts';
import { test } from 'rstack/test';
// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import {
  fixtureConfig,
  resolveFixturePaths,
  fixtureUpperRoot,
  caseSensitiveFileSystem,
} from '../config';
new RuleTester(
  fixtureConfig(
    'internal/plugins/import/rules/no_unresolved/testdata/upstream.txtar',
  ),
).run('no-unresolved', {
  valid: [
    {
      name: 'no-unresolved (node) / run 0 valid 0',
      code: 'import "./malformed.js"',
      filename: 'input.ts',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved (node) / run 0 valid 1',
      code: 'import foo from "./bar";',
      filename: 'input.ts',
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved (node) / run 0 valid 2',
      code: "import bar from './bar.js';",
      filename: 'input.ts',
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved (node) / run 0 valid 3',
      code: "import {someThing} from './test-module';",
      filename: 'input.ts',
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved (node) / run 0 valid 4',
      code: "import fs from 'fs';",
      filename: 'input.ts',
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved (node) / run 0 valid 5 [upstream parser metadata: {"parser":"babel"}]',
      code: "import('fs');",
      filename: 'input.ts',
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved (node) / run 0 valid 6 [upstream parser metadata: {"parserOptions":{"ecmaVersion":2021}}]',
      code: "import('fs');",
      filename: 'input.ts',
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved (node) / run 0 valid 7',
      code: 'import * as foo from "a"',
      filename: 'input.ts',
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved (node) / run 0 valid 8',
      code: 'export { foo } from "./bar"',
      filename: 'input.ts',
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved (node) / run 0 valid 9',
      code: 'export * from "./bar"',
      filename: 'input.ts',
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved (node) / run 0 valid 10',
      code: 'let foo; export { foo }',
      filename: 'input.ts',
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved (node) / run 0 valid 11 [upstream parser metadata: {"parser":"babel"}]',
      code: 'export * as bar from "./bar"',
      filename: 'input.ts',
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved (node) / run 0 valid 12 [upstream parser metadata: {"parser":"babel"}]',
      code: 'export bar from "./bar"',
      filename: 'input.ts',
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: "TypeScript-Go does not support Babel's export-default-from proposal.",
    },
    {
      name: 'no-unresolved (node) / run 0 valid 13',
      code: 'import foo from "./jsx/MyUnCoolComponent.jsx"',
      filename: 'input.ts',
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved (node) / run 0 valid 14',
      code: 'var foo = require("./bar")',
      filename: 'input.ts',
      options: [
        {
          commonjs: true,
        },
      ],
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved (node) / run 0 valid 15',
      code: 'require("./bar")',
      filename: 'input.ts',
      options: [
        {
          commonjs: true,
        },
      ],
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved (node) / run 0 valid 16',
      code: 'require("./does-not-exist")',
      filename: 'input.ts',
      options: [
        {
          commonjs: false,
        },
      ],
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved (node) / run 0 valid 17',
      code: 'require("./does-not-exist")',
      filename: 'input.ts',
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved (node) / run 0 valid 18',
      code: 'require(["./bar"], function (bar) {})',
      filename: 'input.ts',
      options: [
        {
          amd: true,
        },
      ],
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved (node) / run 0 valid 19',
      code: 'define(["./bar"], function (bar) {})',
      filename: 'input.ts',
      options: [
        {
          amd: true,
        },
      ],
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved (node) / run 0 valid 20',
      code: 'require(["./does-not-exist"], function (bar) {})',
      filename: 'input.ts',
      options: [
        {
          amd: false,
        },
      ],
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved (node) / run 0 valid 21',
      code: 'define(["require", "exports", "module"], function (r, e, m) { })',
      filename: 'input.ts',
      options: [
        {
          amd: true,
        },
      ],
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved (node) / run 0 valid 22',
      code: 'require(["./does-not-exist"])',
      filename: 'input.ts',
      options: [
        {
          amd: true,
        },
      ],
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved (node) / run 0 valid 23',
      code: 'define(["./does-not-exist"], function (bar) {})',
      filename: 'input.ts',
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved (node) / run 0 valid 24',
      code: 'require("./does-not-exist", "another arg")',
      filename: 'input.ts',
      options: [
        {
          commonjs: true,
          amd: true,
        },
      ],
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved (node) / run 0 valid 25',
      code: 'proxyquire("./does-not-exist")',
      filename: 'input.ts',
      options: [
        {
          commonjs: true,
          amd: true,
        },
      ],
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved (node) / run 0 valid 26',
      code: '(function() {})("./does-not-exist")',
      filename: 'input.ts',
      options: [
        {
          commonjs: true,
          amd: true,
        },
      ],
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved (node) / run 0 valid 27',
      code: 'define([0, foo], function (bar) {})',
      filename: 'input.ts',
      options: [
        {
          amd: true,
        },
      ],
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved (node) / run 0 valid 28',
      code: 'require(0)',
      filename: 'input.ts',
      options: [
        {
          commonjs: true,
        },
      ],
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved (node) / run 0 valid 29',
      code: 'require(foo)',
      filename: 'input.ts',
      options: [
        {
          commonjs: true,
        },
      ],
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'issue #333 (node) / run 1 valid 0',
      code: 'import foo from "./bar.json"',
      filename: 'input.ts',
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'issue #333 (node) / run 1 valid 1',
      code: 'import foo from "./bar"',
      filename: 'input.ts',
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'issue #333 (node) / run 1 valid 2',
      code: 'import foo from "./bar.json"',
      filename: 'input.ts',
      settings: {
        'import/extensions': ['.js'],
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'issue #333 (node) / run 1 valid 3',
      code: 'import foo from "./bar"',
      filename: 'input.ts',
      settings: {
        'import/extensions': ['.js'],
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'case sensitivity / run 2 valid 0',
      code: 'import foo from "./jsx/MyUncoolComponent.jsx"',
      filename: 'input.ts',
      options: [
        {
          caseSensitive: false,
        },
      ],
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: caseSensitiveFileSystem
        ? 'This case checks incorrect path casing on a case-insensitive filesystem; the fixture filesystem is case-sensitive.'
        : undefined,
    },
    {
      name: 'case sensitivity strict / run 3 valid 0',
      skip: caseSensitiveFileSystem
        ? 'This case checks incorrect path casing on a case-insensitive filesystem; the fixture filesystem is case-sensitive.'
        : undefined,
      create(root: string) {
        const code = resolveFixturePaths(
          'import foo from "<fixture-upper>/jsx/MyUnCoolComponent.jsx"',
          root,
        );
        return {
          code: code,
          filename: 'input.ts',
          settings: {
            'import/resolver': 'node',
            'import/cache': {
              lifetime: 0,
            },
          },
          languageOptions: {
            parserOptions: {
              project: false,
              projectService: false,
            },
          },
        };
      },
    },
    {
      name: 'run 4 valid 0',
      code: 'import "./malformed.js"',
      filename: 'input.ts',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: 'This default-resolver input is already exercised by the Node suite; only its enclosing upstream webpack resolver suite is unsupported.',
    },
    {
      name: 'run 4 valid 1',
      code: 'import foo from "./bar";',
      filename: 'input.ts',
      settings: {
        'import/resolver': 'webpack',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    },
    {
      name: 'run 4 valid 2',
      code: "import bar from './bar.js';",
      filename: 'input.ts',
      settings: {
        'import/resolver': 'webpack',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    },
    {
      name: 'run 4 valid 3',
      code: "import {someThing} from './test-module';",
      filename: 'input.ts',
      settings: {
        'import/resolver': 'webpack',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    },
    {
      name: 'run 4 valid 4',
      code: "import fs from 'fs';",
      filename: 'input.ts',
      settings: {
        'import/resolver': 'webpack',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    },
    {
      name: 'run 4 valid 5 [upstream parser metadata: {"parser":"babel"}]',
      code: "import('fs');",
      filename: 'input.ts',
      settings: {
        'import/resolver': 'webpack',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    },
    {
      name: 'run 4 valid 6 [upstream parser metadata: {"parserOptions":{"ecmaVersion":2021}}]',
      code: "import('fs');",
      filename: 'input.ts',
      settings: {
        'import/resolver': 'webpack',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    },
    {
      name: 'run 4 valid 7',
      code: 'import * as foo from "a"',
      filename: 'input.ts',
      settings: {
        'import/resolver': 'webpack',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    },
    {
      name: 'run 4 valid 8',
      code: 'export { foo } from "./bar"',
      filename: 'input.ts',
      settings: {
        'import/resolver': 'webpack',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    },
    {
      name: 'run 4 valid 9',
      code: 'export * from "./bar"',
      filename: 'input.ts',
      settings: {
        'import/resolver': 'webpack',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    },
    {
      name: 'run 4 valid 10',
      code: 'let foo; export { foo }',
      filename: 'input.ts',
      settings: {
        'import/resolver': 'webpack',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    },
    {
      name: 'run 4 valid 11 [upstream parser metadata: {"parser":"babel"}]',
      code: 'export * as bar from "./bar"',
      filename: 'input.ts',
      settings: {
        'import/resolver': 'webpack',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    },
    {
      name: 'run 4 valid 12 [upstream parser metadata: {"parser":"babel"}]',
      code: 'export bar from "./bar"',
      filename: 'input.ts',
      settings: {
        'import/resolver': 'webpack',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    },
    {
      name: 'run 4 valid 13',
      code: 'import foo from "./jsx/MyUnCoolComponent.jsx"',
      filename: 'input.ts',
      settings: {
        'import/resolver': 'webpack',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    },
    {
      name: 'run 4 valid 14',
      code: 'var foo = require("./bar")',
      filename: 'input.ts',
      options: [
        {
          commonjs: true,
        },
      ],
      settings: {
        'import/resolver': 'webpack',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    },
    {
      name: 'run 4 valid 15',
      code: 'require("./bar")',
      filename: 'input.ts',
      options: [
        {
          commonjs: true,
        },
      ],
      settings: {
        'import/resolver': 'webpack',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    },
    {
      name: 'run 4 valid 16',
      code: 'require("./does-not-exist")',
      filename: 'input.ts',
      options: [
        {
          commonjs: false,
        },
      ],
      settings: {
        'import/resolver': 'webpack',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    },
    {
      name: 'run 4 valid 17',
      code: 'require("./does-not-exist")',
      filename: 'input.ts',
      settings: {
        'import/resolver': 'webpack',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    },
    {
      name: 'run 4 valid 18',
      code: 'require(["./bar"], function (bar) {})',
      filename: 'input.ts',
      options: [
        {
          amd: true,
        },
      ],
      settings: {
        'import/resolver': 'webpack',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    },
    {
      name: 'run 4 valid 19',
      code: 'define(["./bar"], function (bar) {})',
      filename: 'input.ts',
      options: [
        {
          amd: true,
        },
      ],
      settings: {
        'import/resolver': 'webpack',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    },
    {
      name: 'run 4 valid 20',
      code: 'require(["./does-not-exist"], function (bar) {})',
      filename: 'input.ts',
      options: [
        {
          amd: false,
        },
      ],
      settings: {
        'import/resolver': 'webpack',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    },
    {
      name: 'run 4 valid 21',
      code: 'define(["require", "exports", "module"], function (r, e, m) { })',
      filename: 'input.ts',
      options: [
        {
          amd: true,
        },
      ],
      settings: {
        'import/resolver': 'webpack',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    },
    {
      name: 'run 4 valid 22',
      code: 'require(["./does-not-exist"])',
      filename: 'input.ts',
      options: [
        {
          amd: true,
        },
      ],
      settings: {
        'import/resolver': 'webpack',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    },
    {
      name: 'run 4 valid 23',
      code: 'define(["./does-not-exist"], function (bar) {})',
      filename: 'input.ts',
      settings: {
        'import/resolver': 'webpack',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    },
    {
      name: 'run 4 valid 24',
      code: 'require("./does-not-exist", "another arg")',
      filename: 'input.ts',
      options: [
        {
          commonjs: true,
          amd: true,
        },
      ],
      settings: {
        'import/resolver': 'webpack',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    },
    {
      name: 'run 4 valid 25',
      code: 'proxyquire("./does-not-exist")',
      filename: 'input.ts',
      options: [
        {
          commonjs: true,
          amd: true,
        },
      ],
      settings: {
        'import/resolver': 'webpack',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    },
    {
      name: 'run 4 valid 26',
      code: '(function() {})("./does-not-exist")',
      filename: 'input.ts',
      options: [
        {
          commonjs: true,
          amd: true,
        },
      ],
      settings: {
        'import/resolver': 'webpack',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    },
    {
      name: 'run 4 valid 27',
      code: 'define([0, foo], function (bar) {})',
      filename: 'input.ts',
      options: [
        {
          amd: true,
        },
      ],
      settings: {
        'import/resolver': 'webpack',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    },
    {
      name: 'run 4 valid 28',
      code: 'require(0)',
      filename: 'input.ts',
      options: [
        {
          commonjs: true,
        },
      ],
      settings: {
        'import/resolver': 'webpack',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    },
    {
      name: 'run 4 valid 29',
      code: 'require(foo)',
      filename: 'input.ts',
      options: [
        {
          commonjs: true,
        },
      ],
      settings: {
        'import/resolver': 'webpack',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    },
    {
      name: 'run 5 valid 0',
      code: 'import foo from "./bar.json"',
      filename: 'input.ts',
      settings: {
        'import/resolver': 'webpack',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    },
    {
      name: 'run 5 valid 1',
      code: 'import foo from "./bar"',
      filename: 'input.ts',
      settings: {
        'import/resolver': 'webpack',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    },
    {
      name: 'run 5 valid 2',
      code: 'import foo from "./bar.json"',
      filename: 'input.ts',
      settings: {
        'import/extensions': ['.js'],
        'import/resolver': 'webpack',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    },
    {
      name: 'run 5 valid 3',
      code: 'import foo from "./bar"',
      filename: 'input.ts',
      settings: {
        'import/extensions': ['.js'],
        'import/resolver': 'webpack',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    },
    {
      name: 'run 6 valid 0',
      code: 'import foo from "./jsx/MyUncoolComponent.jsx"',
      filename: 'input.ts',
      options: [
        {
          caseSensitive: false,
        },
      ],
      settings: {
        'import/resolver': 'webpack',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    },
    {
      name: 'run 7 valid 0',
      skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
      create(root: string) {
        const code = resolveFixturePaths(
          'import foo from "<fixture-upper>/jsx/MyUnCoolComponent.jsx"',
          root,
        );
        return {
          code: code,
          filename: 'input.ts',
          settings: {
            'import/resolver': 'webpack',
            'import/cache': {
              lifetime: 0,
            },
          },
          languageOptions: {
            parserOptions: {
              project: false,
              projectService: false,
            },
          },
        };
      },
    },
    {
      name: 'no-unresolved (import/resolve legacy) / run 8 valid 0',
      code: "import { DEEP } from 'in-alternate-root';",
      filename: 'input.ts',
      settings: {
        'import/resolve': {
          paths: ['./alternate-root'],
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved (import/resolve legacy) / run 8 valid 1',
      code: "import { DEEP } from 'in-alternate-root'; import { bar } from 'src-bar';",
      filename: 'input.ts',
      settings: {
        'import/resolve': {
          paths: ['./src-root', './alternate-root'],
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved (import/resolve legacy) / run 8 valid 2',
      code: 'import * as foo from "jsx-module/foo"',
      filename: 'input.ts',
      settings: {
        'import/resolve': {
          extensions: ['.jsx'],
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 9 valid 0',
      code: 'import * as foo from "jsx-module/foo"',
      filename: 'input.ts',
      settings: {
        'import/resolver': 'webpack',
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    },
    {
      name: 'run 9 valid 1',
      code: 'import * as foo from "some-loader?with=args!jsx-module/foo"',
      filename: 'input.ts',
      settings: {
        'import/resolver': 'webpack',
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    },
    {
      name: 'no-unresolved ignore list / run 10 valid 0',
      code: 'import "./malformed.js"',
      filename: 'input.ts',
      options: [
        {
          ignore: ['.png$', '.gif$'],
        },
      ],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved ignore list / run 10 valid 1',
      code: 'import "./test.gif.extra"',
      filename: 'input.ts',
      options: [
        {
          ignore: ['.png$', '.gif$'],
        },
      ],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved ignore list / run 10 valid 2',
      code: 'import "./test.gif"',
      filename: 'input.ts',
      options: [
        {
          ignore: ['.png$', '.gif$'],
        },
      ],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved ignore list / run 10 valid 3',
      code: 'import "./test.png"',
      filename: 'input.ts',
      options: [
        {
          ignore: ['.png$', '.gif$'],
        },
      ],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved electron / run 12 valid 0',
      code: 'import "electron"',
      filename: 'input.ts',
      settings: {
        'import/core-modules': ['electron'],
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved syntax verification / run 13 valid 0',
      code: 'for (let { foo, bar } of baz) {}',
      filename: 'input.ts',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved syntax verification / run 13 valid 1',
      code: 'for (let [ foo, bar ] of baz) {}',
      filename: 'input.ts',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved syntax verification / run 13 valid 2',
      code: 'const { x, y } = bar',
      filename: 'input.ts',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved syntax verification / run 13 valid 3 [upstream parser metadata: {"parser":"babel"}]',
      code: 'const { x, y, ...z } = bar',
      filename: 'input.ts',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved syntax verification / run 13 valid 4',
      code: 'let x; export { x }',
      filename: 'input.ts',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved syntax verification / run 13 valid 5',
      code: 'let x; export { x as y }',
      filename: 'input.ts',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved syntax verification / run 13 valid 6',
      code: 'export const x = null',
      filename: 'input.ts',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved syntax verification / run 13 valid 7',
      code: 'export var x = null',
      filename: 'input.ts',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved syntax verification / run 13 valid 8',
      code: 'export let x = null',
      filename: 'input.ts',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved syntax verification / run 13 valid 9',
      code: 'export default x',
      filename: 'input.ts',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved syntax verification / run 13 valid 10',
      code: 'export default class x {}',
      filename: 'input.ts',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved syntax verification / run 13 valid 11',
      code: 'import json from "./data.json"',
      filename: 'input.ts',
      settings: {
        'import/extensions': ['.js'],
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved syntax verification / run 13 valid 12',
      code: 'import foo from "./foobar.json";',
      filename: 'input.ts',
      settings: {
        'import/extensions': ['.js'],
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved syntax verification / run 13 valid 13',
      code: 'import foo from "./foobar";',
      filename: 'input.ts',
      settings: {
        'import/extensions': ['.js'],
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved syntax verification / run 13 valid 14',
      code: 'import { foo } from "./issue-370-commonjs-namespace/bar"',
      filename: 'input.ts',
      settings: {
        'import/ignore': ['foo'],
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved syntax verification / run 13 valid 15',
      code: 'export * from "./issue-370-commonjs-namespace/bar"',
      filename: 'input.ts',
      settings: {
        'import/ignore': ['foo'],
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved syntax verification / run 13 valid 16',
      code: 'import * as a from "./commonjs-namespace/a"; a.b',
      filename: 'input.ts',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'no-unresolved syntax verification / run 13 valid 17',
      code: 'import { foo } from "./ignore.invalid.extension"',
      filename: 'input.ts',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'import() with built-in parser / run 14 valid 0 [upstream parser metadata: {"parserOptions":{"ecmaVersion":2021}}]',
      code: "import('fs');",
      filename: 'input.ts',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'TypeScript / typescript: no-unresolved ignore type-only / run 15 valid 0 [upstream parser metadata: {"parser":"typescript"}]',
      code: 'import type { JSONSchema7Type } from "@types/json-schema";',
      filename: 'input.ts',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'TypeScript / typescript: no-unresolved ignore type-only / run 15 valid 1 [upstream parser metadata: {"parser":"typescript"}]',
      code: 'export type { JSONSchema7Type } from "@types/json-schema";',
      filename: 'input.ts',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
  ],
  invalid: [
    {
      name: 'no-unresolved (node) / run 0 invalid 0',
      code: 'import missing from "./missing/module"',
      filename: 'input.ts',
      settings: {
        'import/ignore': ['^\\./fake/'],
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: "Unable to resolve path to module './missing/module'.",
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      name: 'no-unresolved (node) / run 0 invalid 1',
      code: "import bar from './baz';",
      filename: 'input.ts',
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: "Unable to resolve path to module './baz'.",
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      name: 'no-unresolved (node) / run 0 invalid 2',
      code: "import bar from './empty-folder';",
      filename: 'input.ts',
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: "Unable to resolve path to module './empty-folder'.",
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      name: 'no-unresolved (node) / run 0 invalid 3',
      code: "import { DEEP } from 'in-alternate-root';",
      filename: 'input.ts',
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: "Unable to resolve path to module 'in-alternate-root'.",
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      name: 'no-unresolved (node) / run 0 invalid 4 [upstream parser metadata: {"parser":"babel"}]',
      code: "import('in-alternate-root').then(function({DEEP}) {});",
      filename: 'input.ts',
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: "Unable to resolve path to module 'in-alternate-root'.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      name: 'no-unresolved (node) / run 0 invalid 5',
      code: 'export { foo } from "./does-not-exist"',
      filename: 'input.ts',
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: "Unable to resolve path to module './does-not-exist'.",
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      name: 'no-unresolved (node) / run 0 invalid 6',
      code: 'export * from "./does-not-exist"',
      filename: 'input.ts',
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: "Unable to resolve path to module './does-not-exist'.",
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      name: 'no-unresolved (node) / run 0 invalid 7 [upstream parser metadata: {"parserOptions":{"ecmaVersion":2021}}]',
      code: "import('in-alternate-root').then(function({DEEP}) {});",
      filename: 'input.ts',
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: "Unable to resolve path to module 'in-alternate-root'.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      name: 'no-unresolved (node) / run 0 invalid 8 [upstream parser metadata: {"parser":"babel"}]',
      code: 'export * as bar from "./does-not-exist"',
      filename: 'input.ts',
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: "Unable to resolve path to module './does-not-exist'.",
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      name: 'no-unresolved (node) / run 0 invalid 10',
      code: 'var bar = require("./baz")',
      filename: 'input.ts',
      options: [
        {
          commonjs: true,
        },
      ],
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: "Unable to resolve path to module './baz'.",
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      name: 'no-unresolved (node) / run 0 invalid 11',
      code: 'require("./baz")',
      filename: 'input.ts',
      options: [
        {
          commonjs: true,
        },
      ],
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: "Unable to resolve path to module './baz'.",
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      name: 'no-unresolved (node) / run 0 invalid 12',
      code: 'require(["./baz"], function (bar) {})',
      filename: 'input.ts',
      options: [
        {
          amd: true,
        },
      ],
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: "Unable to resolve path to module './baz'.",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      name: 'no-unresolved (node) / run 0 invalid 13',
      code: 'define(["./baz"], function (bar) {})',
      filename: 'input.ts',
      options: [
        {
          amd: true,
        },
      ],
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: "Unable to resolve path to module './baz'.",
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      name: 'no-unresolved (node) / run 0 invalid 14',
      code: 'define(["./baz", "./bar", "./does-not-exist"], function (bar) {})',
      filename: 'input.ts',
      options: [
        {
          amd: true,
        },
      ],
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: "Unable to resolve path to module './baz'.",
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 16,
        },
        {
          messageId: '',
          message: "Unable to resolve path to module './does-not-exist'.",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 45,
        },
      ],
    },
    {
      name: 'issue #333 (node) / run 1 invalid 0',
      code: 'import bar from "./foo.json"',
      filename: 'input.ts',
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: "Unable to resolve path to module './foo.json'.",
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      name: 'case sensitivity / run 2 invalid 0',
      code: 'import foo from "./jsx/MyUncoolComponent.jsx"',
      filename: 'input.ts',
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: caseSensitiveFileSystem
        ? 'This case checks incorrect path casing on a case-insensitive filesystem; the fixture filesystem is case-sensitive.'
        : undefined,
      errors: [
        {
          messageId: '',
          message:
            'Casing of ./jsx/MyUncoolComponent.jsx does not match the underlying filesystem.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 46,
        },
      ],
    },
    {
      name: 'case sensitivity / run 2 invalid 1',
      code: 'import foo from "./jsx/MyUncoolComponent.jsx"',
      filename: 'input.ts',
      options: [
        {
          caseSensitive: true,
        },
      ],
      settings: {
        'import/resolver': 'node',
        'import/cache': {
          lifetime: 0,
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: caseSensitiveFileSystem
        ? 'This case checks incorrect path casing on a case-insensitive filesystem; the fixture filesystem is case-sensitive.'
        : undefined,
      errors: [
        {
          messageId: '',
          message:
            'Casing of ./jsx/MyUncoolComponent.jsx does not match the underlying filesystem.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 46,
        },
      ],
    },
    {
      name: 'case sensitivity strict / run 3 invalid 0',
      skip: caseSensitiveFileSystem
        ? 'This case checks incorrect path casing on a case-insensitive filesystem; the fixture filesystem is case-sensitive.'
        : undefined,
      create(root: string) {
        const code = resolveFixturePaths(
          'import foo from "<fixture-upper>/jsx/MyUnCoolComponent.jsx"',
          root,
        );
        return {
          code: code,
          filename: 'input.ts',
          options: [
            {
              caseSensitiveStrict: true,
            },
          ],
          settings: {
            'import/resolver': 'node',
            'import/cache': {
              lifetime: 0,
            },
          },
          languageOptions: {
            parserOptions: {
              project: false,
              projectService: false,
            },
          },
          errors: [
            {
              messageId: '',
              message: resolveFixturePaths(
                'Casing of <fixture-upper>/jsx/MyUnCoolComponent.jsx does not match the underlying filesystem.',
                root,
              ),
              line: 1,
              column: 17,
              endLine: 1,
              endColumn: 149 + 1 * (fixtureUpperRoot(root).length - 104),
            },
          ],
        };
      },
    },
    {
      name: 'case sensitivity strict / run 3 invalid 1',
      skip: caseSensitiveFileSystem
        ? 'This case checks incorrect path casing on a case-insensitive filesystem; the fixture filesystem is case-sensitive.'
        : undefined,
      create(root: string) {
        const code = resolveFixturePaths(
          'import foo from "<fixture-upper>/jsx/MyUnCoolComponent.jsx"',
          root,
        );
        return {
          code: code,
          filename: 'input.ts',
          options: [
            {
              caseSensitiveStrict: true,
              caseSensitive: false,
            },
          ],
          settings: {
            'import/resolver': 'node',
            'import/cache': {
              lifetime: 0,
            },
          },
          languageOptions: {
            parserOptions: {
              project: false,
              projectService: false,
            },
          },
          errors: [
            {
              messageId: '',
              message: resolveFixturePaths(
                'Casing of <fixture-upper>/jsx/MyUnCoolComponent.jsx does not match the underlying filesystem.',
                root,
              ),
              line: 1,
              column: 17,
              endLine: 1,
              endColumn: 149 + 1 * (fixtureUpperRoot(root).length - 104),
            },
          ],
        };
      },
    },
    {
      name: 'no-unresolved (import/resolve legacy) / run 8 invalid 0',
      code: 'import * as foo from "jsx-module/foo"',
      filename: 'input.ts',
      settings: {
        'import/resolver': 'node',
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: "Unable to resolve path to module 'jsx-module/foo'.",
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      name: 'no-unresolved ignore list / run 10 invalid 0',
      code: 'import "./test.gif"',
      filename: 'input.ts',
      options: [
        {
          ignore: ['.png$'],
        },
      ],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: "Unable to resolve path to module './test.gif'.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      name: 'no-unresolved ignore list / run 10 invalid 1',
      code: 'import "./test.png"',
      filename: 'input.ts',
      options: [
        {
          ignore: ['.gif$'],
        },
      ],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: "Unable to resolve path to module './test.png'.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      name: 'no-unresolved unknown resolver / run 11 invalid 0',
      code: 'import "./malformed.js"',
      filename: 'input.ts',
      settings: {
        'import/resolver': 'does-not-exist',
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: 'Resolve error: unable to load resolver "does-not-exist".',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 1,
        },
        {
          messageId: '',
          message: "Unable to resolve path to module './malformed.js'.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      name: 'no-unresolved unknown resolver / run 11 invalid 1',
      code: 'import "./malformed.js"; import "./fake.js"',
      filename: 'input.ts',
      settings: {
        'import/resolver': 'does-not-exist',
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: 'Resolve error: unable to load resolver "does-not-exist".',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 1,
        },
        {
          messageId: '',
          message: "Unable to resolve path to module './malformed.js'.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 24,
        },
        {
          messageId: '',
          message: "Unable to resolve path to module './fake.js'.",
          line: 1,
          column: 33,
          endLine: 1,
          endColumn: 44,
        },
      ],
    },
    {
      name: 'no-unresolved electron / run 12 invalid 0',
      code: 'import "electron"',
      filename: 'input.ts',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: "Unable to resolve path to module 'electron'.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      name: 'import() with built-in parser / run 14 invalid 0 [upstream parser metadata: {"parserOptions":{"ecmaVersion":2021}}]',
      code: 'import("./does-not-exist-dynamic").then(() => {})',
      filename: 'input.ts',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "Unable to resolve path to module './does-not-exist-dynamic'.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      name: 'TypeScript / typescript: no-unresolved ignore type-only / run 15 invalid 0 [upstream parser metadata: {"parser":"typescript"}]',
      code: 'import { JSONSchema7Type } from "@types/json-schema";',
      filename: 'input.ts',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: "Unable to resolve path to module '@types/json-schema'.",
          line: 1,
          column: 33,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
    {
      name: 'TypeScript / typescript: no-unresolved ignore type-only / run 15 invalid 1 [upstream parser metadata: {"parser":"typescript"}]',
      code: 'export { JSONSchema7Type } from "@types/json-schema";',
      filename: 'input.ts',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: "Unable to resolve path to module '@types/json-schema'.",
          line: 1,
          column: 33,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
  ],
});

// Preserved upstream inputs/assertions, not executable native expectations.
const skippedInvalidCases = [
  {
    name: 'no-unresolved (node) / run 0 invalid 9 [upstream parser metadata: {"parser":"babel"}]',
    code: 'export bar from "./does-not-exist"',
    filename: 'input.ts',
    settings: {
      'import/resolver': 'node',
      'import/cache': {
        lifetime: 0,
      },
    },
    languageOptions: {
      parserOptions: {
        project: false,
        projectService: false,
      },
    },
    skip: "TypeScript-Go does not support Babel's export-default-from proposal.",
    errors: ["Unable to resolve path to module './does-not-exist'."],
  },
  {
    name: 'run 4 invalid 0',
    code: 'import missing from "./missing/module"',
    filename: 'input.ts',
    settings: {
      'import/ignore': ['^\\./fake/'],
      'import/resolver': 'webpack',
      'import/cache': {
        lifetime: 0,
      },
    },
    languageOptions: {
      parserOptions: {
        project: false,
        projectService: false,
      },
    },
    skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    errors: [
      {
        message: "Unable to resolve path to module './missing/module'.",
      },
    ],
  },
  {
    name: 'run 4 invalid 1',
    code: "import bar from './baz';",
    filename: 'input.ts',
    settings: {
      'import/resolver': 'webpack',
      'import/cache': {
        lifetime: 0,
      },
    },
    languageOptions: {
      parserOptions: {
        project: false,
        projectService: false,
      },
    },
    skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    errors: [
      {
        message: "Unable to resolve path to module './baz'.",
        type: 'Literal',
      },
    ],
  },
  {
    name: 'run 4 invalid 2',
    code: "import bar from './empty-folder';",
    filename: 'input.ts',
    settings: {
      'import/resolver': 'webpack',
      'import/cache': {
        lifetime: 0,
      },
    },
    languageOptions: {
      parserOptions: {
        project: false,
        projectService: false,
      },
    },
    skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    errors: [
      {
        message: "Unable to resolve path to module './empty-folder'.",
        type: 'Literal',
      },
    ],
  },
  {
    name: 'run 4 invalid 3',
    code: "import { DEEP } from 'in-alternate-root';",
    filename: 'input.ts',
    settings: {
      'import/resolver': 'webpack',
      'import/cache': {
        lifetime: 0,
      },
    },
    languageOptions: {
      parserOptions: {
        project: false,
        projectService: false,
      },
    },
    skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    errors: [
      {
        message: "Unable to resolve path to module 'in-alternate-root'.",
        type: 'Literal',
      },
    ],
  },
  {
    name: 'run 4 invalid 4 [upstream parser metadata: {"parser":"babel"}]',
    code: "import('in-alternate-root').then(function({DEEP}) {});",
    filename: 'input.ts',
    settings: {
      'import/resolver': 'webpack',
      'import/cache': {
        lifetime: 0,
      },
    },
    languageOptions: {
      parserOptions: {
        project: false,
        projectService: false,
      },
    },
    skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    errors: [
      {
        message: "Unable to resolve path to module 'in-alternate-root'.",
        type: 'Literal',
      },
    ],
  },
  {
    name: 'run 4 invalid 5',
    code: 'export { foo } from "./does-not-exist"',
    filename: 'input.ts',
    settings: {
      'import/resolver': 'webpack',
      'import/cache': {
        lifetime: 0,
      },
    },
    languageOptions: {
      parserOptions: {
        project: false,
        projectService: false,
      },
    },
    skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    errors: ["Unable to resolve path to module './does-not-exist'."],
  },
  {
    name: 'run 4 invalid 6',
    code: 'export * from "./does-not-exist"',
    filename: 'input.ts',
    settings: {
      'import/resolver': 'webpack',
      'import/cache': {
        lifetime: 0,
      },
    },
    languageOptions: {
      parserOptions: {
        project: false,
        projectService: false,
      },
    },
    skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    errors: ["Unable to resolve path to module './does-not-exist'."],
  },
  {
    name: 'run 4 invalid 7 [upstream parser metadata: {"parserOptions":{"ecmaVersion":2021}}]',
    code: "import('in-alternate-root').then(function({DEEP}) {});",
    filename: 'input.ts',
    settings: {
      'import/resolver': 'webpack',
      'import/cache': {
        lifetime: 0,
      },
    },
    languageOptions: {
      parserOptions: {
        project: false,
        projectService: false,
      },
    },
    skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    errors: [
      {
        message: "Unable to resolve path to module 'in-alternate-root'.",
        type: 'Literal',
      },
    ],
  },
  {
    name: 'run 4 invalid 8 [upstream parser metadata: {"parser":"babel"}]',
    code: 'export * as bar from "./does-not-exist"',
    filename: 'input.ts',
    settings: {
      'import/resolver': 'webpack',
      'import/cache': {
        lifetime: 0,
      },
    },
    languageOptions: {
      parserOptions: {
        project: false,
        projectService: false,
      },
    },
    skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    errors: ["Unable to resolve path to module './does-not-exist'."],
  },
  {
    name: 'run 4 invalid 9 [upstream parser metadata: {"parser":"babel"}]',
    code: 'export bar from "./does-not-exist"',
    filename: 'input.ts',
    settings: {
      'import/resolver': 'webpack',
      'import/cache': {
        lifetime: 0,
      },
    },
    languageOptions: {
      parserOptions: {
        project: false,
        projectService: false,
      },
    },
    skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    errors: ["Unable to resolve path to module './does-not-exist'."],
  },
  {
    name: 'run 4 invalid 10',
    code: 'var bar = require("./baz")',
    filename: 'input.ts',
    options: [
      {
        commonjs: true,
      },
    ],
    settings: {
      'import/resolver': 'webpack',
      'import/cache': {
        lifetime: 0,
      },
    },
    languageOptions: {
      parserOptions: {
        project: false,
        projectService: false,
      },
    },
    skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    errors: [
      {
        message: "Unable to resolve path to module './baz'.",
        type: 'Literal',
      },
    ],
  },
  {
    name: 'run 4 invalid 11',
    code: 'require("./baz")',
    filename: 'input.ts',
    options: [
      {
        commonjs: true,
      },
    ],
    settings: {
      'import/resolver': 'webpack',
      'import/cache': {
        lifetime: 0,
      },
    },
    languageOptions: {
      parserOptions: {
        project: false,
        projectService: false,
      },
    },
    skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    errors: [
      {
        message: "Unable to resolve path to module './baz'.",
        type: 'Literal',
      },
    ],
  },
  {
    name: 'run 4 invalid 12',
    code: 'require(["./baz"], function (bar) {})',
    filename: 'input.ts',
    options: [
      {
        amd: true,
      },
    ],
    settings: {
      'import/resolver': 'webpack',
      'import/cache': {
        lifetime: 0,
      },
    },
    languageOptions: {
      parserOptions: {
        project: false,
        projectService: false,
      },
    },
    skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    errors: [
      {
        message: "Unable to resolve path to module './baz'.",
        type: 'Literal',
      },
    ],
  },
  {
    name: 'run 4 invalid 13',
    code: 'define(["./baz"], function (bar) {})',
    filename: 'input.ts',
    options: [
      {
        amd: true,
      },
    ],
    settings: {
      'import/resolver': 'webpack',
      'import/cache': {
        lifetime: 0,
      },
    },
    languageOptions: {
      parserOptions: {
        project: false,
        projectService: false,
      },
    },
    skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    errors: [
      {
        message: "Unable to resolve path to module './baz'.",
        type: 'Literal',
      },
    ],
  },
  {
    name: 'run 4 invalid 14',
    code: 'define(["./baz", "./bar", "./does-not-exist"], function (bar) {})',
    filename: 'input.ts',
    options: [
      {
        amd: true,
      },
    ],
    settings: {
      'import/resolver': 'webpack',
      'import/cache': {
        lifetime: 0,
      },
    },
    languageOptions: {
      parserOptions: {
        project: false,
        projectService: false,
      },
    },
    skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    errors: [
      {
        message: "Unable to resolve path to module './baz'.",
        type: 'Literal',
      },
      {
        message: "Unable to resolve path to module './does-not-exist'.",
        type: 'Literal',
      },
    ],
  },
  {
    name: 'run 5 invalid 0',
    code: 'import bar from "./foo.json"',
    filename: 'input.ts',
    settings: {
      'import/resolver': 'webpack',
      'import/cache': {
        lifetime: 0,
      },
    },
    languageOptions: {
      parserOptions: {
        project: false,
        projectService: false,
      },
    },
    skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    errors: ["Unable to resolve path to module './foo.json'."],
  },
  {
    name: 'run 6 invalid 0',
    code: 'import foo from "./jsx/MyUncoolComponent.jsx"',
    filename: 'input.ts',
    settings: {
      'import/resolver': 'webpack',
      'import/cache': {
        lifetime: 0,
      },
    },
    languageOptions: {
      parserOptions: {
        project: false,
        projectService: false,
      },
    },
    skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    errors: [
      'Casing of ./jsx/MyUncoolComponent.jsx does not match the underlying filesystem.',
    ],
  },
  {
    name: 'run 6 invalid 1',
    code: 'import foo from "./jsx/MyUncoolComponent.jsx"',
    filename: 'input.ts',
    options: [
      {
        caseSensitive: true,
      },
    ],
    settings: {
      'import/resolver': 'webpack',
      'import/cache': {
        lifetime: 0,
      },
    },
    languageOptions: {
      parserOptions: {
        project: false,
        projectService: false,
      },
    },
    skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    errors: [
      'Casing of ./jsx/MyUncoolComponent.jsx does not match the underlying filesystem.',
    ],
  },
  {
    name: 'run 7 invalid 0',
    skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    create(root: string) {
      const code = resolveFixturePaths(
        'import foo from "<fixture-upper>/jsx/MyUnCoolComponent.jsx"',
        root,
      );
      return {
        code: code,
        filename: 'input.ts',
        options: [
          {
            caseSensitiveStrict: true,
          },
        ],
        settings: {
          'import/resolver': 'webpack',
          'import/cache': {
            lifetime: 0,
          },
        },
        languageOptions: {
          parserOptions: {
            project: false,
            projectService: false,
          },
        },
        errors: [
          'Casing of <fixture-upper>/jsx/MyUnCoolComponent.jsx does not match the underlying filesystem.',
        ],
      };
    },
  },
  {
    name: 'run 7 invalid 1',
    skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    create(root: string) {
      const code = resolveFixturePaths(
        'import foo from "<fixture-upper>/jsx/MyUnCoolComponent.jsx"',
        root,
      );
      return {
        code: code,
        filename: 'input.ts',
        options: [
          {
            caseSensitiveStrict: true,
            caseSensitive: false,
          },
        ],
        settings: {
          'import/resolver': 'webpack',
          'import/cache': {
            lifetime: 0,
          },
        },
        languageOptions: {
          parserOptions: {
            project: false,
            projectService: false,
          },
        },
        errors: [
          'Casing of <fixture-upper>/jsx/MyUnCoolComponent.jsx does not match the underlying filesystem.',
        ],
      };
    },
  },
  {
    name: 'run 9 invalid 0',
    code: 'import * as foo from "jsx-module/foo"',
    filename: 'input.ts',
    settings: {
      'import/resolver': {
        webpack: {
          config: 'webpack.empty.config.js',
        },
      },
    },
    languageOptions: {
      parserOptions: {
        project: false,
        projectService: false,
      },
    },
    skip: 'The case requires a webpack JavaScript resolver; the native Import runtime supports Node and TypeScript resolvers.',
    errors: ["Unable to resolve path to module 'jsx-module/foo'."],
  },
];
for (const item of skippedInvalidCases) {
  checkSkip(item);
  test.skip(`invalid ${item.name}: ${item.skip}`, () => {});
}

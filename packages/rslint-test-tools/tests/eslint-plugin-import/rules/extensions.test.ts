import { checkSkip } from '../../../../rslint-rule-tester/src/cases.ts';
import { test } from 'rstack/test';
// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
new RuleTester(
  fixtureConfig(
    'internal/plugins/import/rules/extensions/testdata/upstream.txtar',
  ),
).run('extensions', {
  valid: [
    {
      name: 'extensions / run 0 valid 0',
      code: 'import a from "@/a"',
      filename: 'files/input.ts',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'extensions / run 0 valid 1',
      code: 'import a from "a"',
      filename: 'files/input.ts',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'extensions / run 0 valid 2',
      code: 'import dot from "./file.with.dot"',
      filename: 'files/input.ts',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'extensions / run 0 valid 3',
      code: 'import a from "a/index.js"',
      filename: 'files/input.ts',
      options: ['always'],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'extensions / run 0 valid 4',
      code: 'import dot from "./file.with.dot.js"',
      filename: 'files/input.ts',
      options: ['always'],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'extensions / run 0 valid 5',
      code: 'import a from "a"\nimport packageConfig from "./package.json"',
      filename: 'files/input.ts',
      options: [
        {
          json: 'always',
          js: 'never',
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
      name: 'extensions / run 0 valid 6',
      code: 'import lib from "./bar"\nimport component from "./bar.jsx"\nimport data from "./bar.json"',
      filename: 'files/input.ts',
      options: ['never'],
      settings: {
        'import/resolve': {
          extensions: ['.js', '.jsx', '.json'],
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
      name: 'extensions / run 0 valid 7',
      code: 'import bar from "./bar"\nimport barjson from "./bar.json"\nimport barhbs from "./bar.hbs"',
      filename: 'files/input.ts',
      options: [
        'always',
        {
          js: 'never',
          jsx: 'never',
        },
      ],
      settings: {
        'import/resolve': {
          extensions: ['.js', '.jsx', '.json', '.hbs'],
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
      name: 'extensions / run 0 valid 8',
      code: 'import bar from "./bar.js"\nimport pack from "./package"',
      filename: 'files/input.ts',
      options: [
        'never',
        {
          js: 'always',
          json: 'never',
        },
      ],
      settings: {
        'import/resolve': {
          extensions: ['.js', '.json'],
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
      name: 'extensions / run 0 valid 9',
      code: 'import path from "path"',
      filename: 'files/input.ts',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'extensions / run 0 valid 10',
      code: 'import path from "path"',
      filename: 'files/input.ts',
      options: ['never'],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'extensions / run 0 valid 11',
      code: 'import path from "path"',
      filename: 'files/input.ts',
      options: ['always'],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'extensions / run 0 valid 12',
      code: 'import thing from "./fake-file.js"',
      filename: 'files/input.ts',
      options: ['always'],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'extensions / run 0 valid 13',
      code: 'import thing from "non-package"',
      filename: 'files/input.ts',
      options: ['never'],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'extensions / run 0 valid 14',
      code: "\n        import foo from './foo.js'\n        import bar from './bar.json'\n        import Component from './Component.jsx'\n        import express from 'express'\n      ",
      filename: 'files/input.ts',
      options: ['ignorePackages'],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'extensions / run 0 valid 15',
      code: "\n        import foo from './foo.js'\n        import bar from './bar.json'\n        import Component from './Component.jsx'\n        import express from 'express'\n      ",
      filename: 'files/input.ts',
      options: [
        'always',
        {
          ignorePackages: true,
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
      name: 'extensions / run 0 valid 16',
      code: "\n        import foo from './foo'\n        import bar from './bar'\n        import Component from './Component'\n        import express from 'express'\n      ",
      filename: 'files/input.ts',
      options: [
        'never',
        {
          ignorePackages: true,
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
      name: 'extensions / run 0 valid 17',
      code: 'import exceljs from "exceljs"',
      filename: 'files/internal-modules/plugins/plugin.js',
      options: [
        'always',
        {
          js: 'never',
          jsx: 'never',
        },
      ],
      settings: {
        'import/resolver': {
          node: {
            extensions: ['.js', '.jsx', '.json'],
          },
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
    },
    {
      name: 'extensions / run 0 valid 18',
      code: 'export { foo } from "./foo.js"\nlet bar; export { bar }',
      filename: 'files/input.ts',
      options: ['always'],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'extensions / run 0 valid 19',
      code: 'export { foo } from "./foo"\nlet bar; export { bar }',
      filename: 'files/input.ts',
      options: ['never'],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'extensions / run 0 valid 20',
      code: 'import lib from "pkg.js"\nimport lib2 from "pgk/package"\nimport lib3 from "@name/pkg.js"',
      filename: 'files/input.ts',
      options: ['never'],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'extensions / run 0 valid 21',
      code: 'import bare from "./foo?a=True.ext"',
      filename: 'files/input.ts',
      options: ['never'],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'extensions / run 0 valid 22',
      code: 'import bare from "./foo.js?a=True"',
      filename: 'files/input.ts',
      options: ['always'],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'extensions / run 0 valid 23',
      code: 'import lib from "pkg"\nimport lib2 from "pgk/package.js"\nimport lib3 from "@name/pkg"',
      filename: 'files/input.ts',
      options: ['always'],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'TypeScript / typescript: extensions ignore type-only / run 1 valid 0 [upstream parser metadata: {"parser":"typescript"}]',
      code: 'import type T from "./typescript-declare";',
      filename: 'files/input.ts',
      options: [
        'always',
        {
          ts: 'never',
          tsx: 'never',
          js: 'never',
          jsx: 'never',
        },
      ],
      settings: {
        'import/resolver': 'node',
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'TypeScript / typescript: extensions ignore type-only / run 1 valid 1 [upstream parser metadata: {"parser":"typescript"}]',
      code: 'export type { MyType } from "./typescript-declare";',
      filename: 'files/input.ts',
      options: [
        'always',
        {
          ts: 'never',
          tsx: 'never',
          js: 'never',
          jsx: 'never',
        },
      ],
      settings: {
        'import/resolver': 'node',
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'TypeScript / typescript: (with TS resolver) extensions are enforced for type imports/export when checkTypeImports is set / run 2 valid 0 [upstream parser metadata: {"parser":"typescript"}]',
      code: 'import type { MyType } from "./typescript-declare.ts";',
      filename: 'files/input.ts',
      options: [
        'always',
        {
          checkTypeImports: true,
        },
      ],
      settings: {
        'import/resolver': {
          typescript: {
            alwaysTryTypes: true,
          },
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
      name: 'TypeScript / typescript: (with TS resolver) extensions are enforced for type imports/export when checkTypeImports is set / run 2 valid 1 [upstream parser metadata: {"parser":"typescript"}]',
      code: 'export type { MyType } from "./typescript-declare.ts";',
      filename: 'files/input.ts',
      options: [
        'always',
        {
          checkTypeImports: true,
        },
      ],
      settings: {
        'import/resolver': {
          typescript: {
            alwaysTryTypes: true,
          },
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
      name: 'TypeScript / typescript: (with TS resolver) extensions are enforced for type imports/export when checkTypeImports is set / run 2 valid 2 [upstream parser metadata: {"parser":"typescript"}]',
      code: "\n              import { ErrorMessage as UpstreamErrorMessage } from '@black-flag/core/util';\n\n              import { $instances } from 'rootverse+debug:src.ts';\n              import { $exists } from 'rootverse+bfe:src/symbols.ts';\n\n              import type { Entries } from 'type-fest';\n            ",
      filename: 'files/input.ts',
      options: [
        'always',
        {
          ignorePackages: true,
          checkTypeImports: true,
          pathGroupOverrides: [
            {
              pattern: 'multiverse{*,*/**}',
              action: 'enforce',
            },
          ],
        },
      ],
      settings: {
        'import/resolver': {
          typescript: {
            alwaysTryTypes: true,
          },
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
      name: 'TypeScript / typescript: (with TS resolver) extensions are enforced for type imports/export when checkTypeImports is set / run 2 valid 3 [upstream parser metadata: {"parser":"typescript"}]',
      code: "\n              import { ErrorMessage as UpstreamErrorMessage } from '@black-flag/core/util';\n\n              import { $instances } from 'rootverse+debug:src.ts';\n              import { $exists } from 'rootverse+bfe:src/symbols.ts';\n\n              import type { Entries } from 'type-fest';\n            ",
      filename: 'files/input.ts',
      options: [
        'always',
        {
          ignorePackages: true,
          checkTypeImports: true,
          pathGroupOverrides: [
            {
              pattern: 'rootverse{*,*/**}',
              action: 'enforce',
            },
          ],
        },
      ],
      settings: {
        'import/resolver': {
          typescript: {
            alwaysTryTypes: true,
          },
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
      name: 'TypeScript / typescript: (with TS resolver) extensions are enforced for type imports/export when checkTypeImports is set / run 2 valid 4 [upstream parser metadata: {"parser":"typescript"}]',
      code: "\n              import { ErrorMessage as UpstreamErrorMessage } from '@black-flag/core/util';\n\n              import { $instances } from 'rootverse+debug:src';\n              import { $exists } from 'rootverse+bfe:src/symbols';\n\n              import type { Entries } from 'type-fest';\n            ",
      filename: 'files/input.ts',
      options: [
        'always',
        {
          ignorePackages: true,
          checkTypeImports: true,
          pathGroupOverrides: [
            {
              pattern: 'multiverse{*,*/**}',
              action: 'enforce',
            },
            {
              pattern: 'rootverse{*,*/**}',
              action: 'ignore',
            },
          ],
        },
      ],
      settings: {
        'import/resolver': {
          typescript: {
            alwaysTryTypes: true,
          },
        },
      },
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
      name: 'extensions / run 0 invalid 0',
      code: 'import a from "a/index.js"',
      filename: 'files/input.ts',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: 'Unexpected use of file extension "js" for "a/index.js"',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      name: 'extensions / run 0 invalid 1',
      code: 'import dot from "./file.with.dot"',
      filename: 'files/input.ts',
      options: ['always'],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: 'Missing file extension "js" for "./file.with.dot"',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      name: 'extensions / run 0 invalid 2',
      code: 'import a from "a/index.js"\nimport packageConfig from "./package"',
      filename: 'files/input.ts',
      options: [
        {
          json: 'always',
          js: 'never',
        },
      ],
      settings: {
        'import/resolve': {
          extensions: ['.js', '.json'],
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
          message: 'Unexpected use of file extension "js" for "a/index.js"',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 27,
        },
        {
          messageId: '',
          message: 'Missing file extension "json" for "./package"',
          line: 2,
          column: 27,
          endLine: 2,
          endColumn: 38,
        },
      ],
    },
    {
      name: 'extensions / run 0 invalid 3',
      code: 'import lib from "./bar.js"\nimport component from "./bar.jsx"\nimport data from "./bar.json"',
      filename: 'files/input.ts',
      options: ['never'],
      settings: {
        'import/resolve': {
          extensions: ['.js', '.jsx', '.json'],
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
          message: 'Unexpected use of file extension "js" for "./bar.js"',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      name: 'extensions / run 0 invalid 4',
      code: 'import lib from "./bar.js"\nimport component from "./bar.jsx"\nimport data from "./bar.json"',
      filename: 'files/input.ts',
      options: [
        {
          json: 'always',
          js: 'never',
          jsx: 'never',
        },
      ],
      settings: {
        'import/resolve': {
          extensions: ['.js', '.jsx', '.json'],
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
          message: 'Unexpected use of file extension "js" for "./bar.js"',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      name: 'extensions / run 0 invalid 5',
      code: 'import component from "./bar.jsx"\nimport data from "./bar.json"',
      filename: 'files/input.ts',
      options: [
        {
          json: 'always',
          js: 'never',
          jsx: 'never',
        },
      ],
      settings: {
        'import/resolve': {
          extensions: ['.jsx', '.json', '.js'],
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
          message: 'Unexpected use of file extension "jsx" for "./bar.jsx"',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      name: 'extensions / run 0 invalid 6',
      code: 'import "./bar.coffee"',
      filename: 'files/input.ts',
      options: [
        'never',
        {
          js: 'always',
          jsx: 'always',
        },
      ],
      settings: {
        'import/resolve': {
          extensions: ['.coffee', '.js'],
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
          message:
            'Unexpected use of file extension "coffee" for "./bar.coffee"',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      name: 'extensions / run 0 invalid 7',
      code: 'import barjs from "./bar.js"\nimport barjson from "./bar.json"\nimport barnone from "./bar"',
      filename: 'files/input.ts',
      options: [
        'always',
        {
          json: 'always',
          js: 'never',
          jsx: 'never',
        },
      ],
      settings: {
        'import/resolve': {
          extensions: ['.js', '.jsx', '.json'],
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
          message: 'Unexpected use of file extension "js" for "./bar.js"',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      name: 'extensions / run 0 invalid 8',
      code: 'import barjs from "."\nimport barjs2 from ".."',
      filename: 'files/input.ts',
      options: ['always'],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: 'Missing file extension "js" for "."',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 22,
        },
        {
          messageId: '',
          message: 'Missing file extension "js" for ".."',
          line: 2,
          column: 20,
          endLine: 2,
          endColumn: 24,
        },
      ],
    },
    {
      name: 'extensions / run 0 invalid 9',
      code: 'import barjs from "./bar.js"\nimport barjson from "./bar.json"\nimport barnone from "./bar"',
      filename: 'files/input.ts',
      options: [
        'never',
        {
          json: 'always',
          js: 'never',
          jsx: 'never',
        },
      ],
      settings: {
        'import/resolve': {
          extensions: ['.js', '.jsx', '.json'],
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
          message: 'Unexpected use of file extension "js" for "./bar.js"',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      name: 'extensions / run 0 invalid 10',
      code: 'import thing from "./fake-file.js"',
      filename: 'files/input.ts',
      options: ['never'],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: 'Unexpected use of file extension "js" for "./fake-file.js"',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      name: 'extensions / run 0 invalid 11',
      code: 'import thing from "non-package/test"',
      filename: 'files/input.ts',
      options: ['always'],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: 'Missing file extension for "non-package/test"',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      name: 'extensions / run 0 invalid 12',
      code: 'import thing from "@name/pkg/test"',
      filename: 'files/input.ts',
      options: ['always'],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: 'Missing file extension for "@name/pkg/test"',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      name: 'extensions / run 0 invalid 13',
      code: 'import thing from "@name/pkg/test.js"',
      filename: 'files/input.ts',
      options: ['never'],
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
            'Unexpected use of file extension "js" for "@name/pkg/test.js"',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      name: 'extensions / run 0 invalid 14',
      code: "\n        import foo from './foo.js'\n        import bar from './bar.json'\n        import Component from './Component'\n        import baz from 'foo/baz'\n        import baw from '@scoped/baw/import'\n        import chart from '@/configs/chart'\n        import express from 'express'\n      ",
      filename: 'files/input.ts',
      options: [
        'always',
        {
          ignorePackages: true,
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
          message: 'Missing file extension for "./Component"',
          line: 4,
          column: 31,
          endLine: 4,
          endColumn: 44,
        },
        {
          messageId: '',
          message: 'Missing file extension for "@/configs/chart"',
          line: 7,
          column: 27,
          endLine: 7,
          endColumn: 44,
        },
      ],
    },
    {
      name: 'extensions / run 0 invalid 15',
      code: "\n        import foo from './foo.js'\n        import bar from './bar.json'\n        import Component from './Component'\n        import baz from 'foo/baz'\n        import baw from '@scoped/baw/import'\n        import chart from '@/configs/chart'\n        import express from 'express'\n      ",
      filename: 'files/input.ts',
      options: ['ignorePackages'],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: 'Missing file extension for "./Component"',
          line: 4,
          column: 31,
          endLine: 4,
          endColumn: 44,
        },
        {
          messageId: '',
          message: 'Missing file extension for "@/configs/chart"',
          line: 7,
          column: 27,
          endLine: 7,
          endColumn: 44,
        },
      ],
    },
    {
      name: 'extensions / run 0 invalid 16',
      code: "\n        import foo from './foo.js'\n        import bar from './bar.json'\n        import Component from './Component.jsx'\n        import express from 'express'\n      ",
      filename: 'files/input.ts',
      options: [
        'never',
        {
          ignorePackages: true,
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
          message: 'Unexpected use of file extension "js" for "./foo.js"',
          line: 2,
          column: 25,
          endLine: 2,
          endColumn: 35,
        },
        {
          messageId: '',
          message:
            'Unexpected use of file extension "jsx" for "./Component.jsx"',
          line: 4,
          column: 31,
          endLine: 4,
          endColumn: 48,
        },
      ],
    },
    {
      name: 'extensions / run 0 invalid 17',
      code: "\n        import foo from './foo.js'\n        import bar from './bar.json'\n        import Component from './Component.jsx'\n      ",
      filename: 'files/input.ts',
      options: [
        'always',
        {
          pattern: {
            jsx: 'never',
          },
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
          message:
            'Unexpected use of file extension "jsx" for "./Component.jsx"',
          line: 4,
          column: 31,
          endLine: 4,
          endColumn: 48,
        },
      ],
    },
    {
      name: 'extensions / run 0 invalid 18',
      code: 'export { foo } from "./foo"\nlet bar; export { bar }',
      filename: 'files/input.ts',
      options: ['always'],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: 'Missing file extension for "./foo"',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      name: 'extensions / run 0 invalid 19',
      code: 'export { foo } from "./foo.js"\nlet bar; export { bar }',
      filename: 'files/input.ts',
      options: ['never'],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: 'Unexpected use of file extension "js" for "./foo.js"',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      name: 'extensions / run 0 invalid 20',
      code: 'import withExtension from "./foo.js?a=True"',
      filename: 'files/input.ts',
      options: ['never'],
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
            'Unexpected use of file extension "js" for "./foo.js?a=True"',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 44,
        },
      ],
    },
    {
      name: 'extensions / run 0 invalid 21',
      code: 'import withoutExtension from "./foo?a=True.ext"',
      filename: 'files/input.ts',
      options: ['always'],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: 'Missing file extension for "./foo?a=True.ext"',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 48,
        },
      ],
    },
    {
      name: 'extensions / run 0 invalid 22',
      code: 'const { foo } = require("./foo")\nexport { foo }',
      filename: 'files/input.ts',
      options: ['always'],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: 'Missing file extension for "./foo"',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      name: 'extensions / run 0 invalid 23',
      code: 'const { foo } = require("./foo.js")\nexport { foo }',
      filename: 'files/input.ts',
      options: ['never'],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: 'Unexpected use of file extension "js" for "./foo.js"',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      name: 'extensions / run 0 invalid 24',
      code: 'export { foo } from "./foo"',
      filename: 'files/input.ts',
      options: ['always'],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: 'Missing file extension for "./foo"',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      name: 'extensions / run 0 invalid 25',
      code: '\n        import foo from "@/ImNotAScopedModule";\n        import chart from \'@/configs/chart\';\n      ',
      filename: 'files/input.ts',
      options: ['always'],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: 'Missing file extension for "@/ImNotAScopedModule"',
          line: 2,
          column: 25,
          endLine: 2,
          endColumn: 47,
        },
        {
          messageId: '',
          message: 'Missing file extension for "@/configs/chart"',
          line: 3,
          column: 27,
          endLine: 3,
          endColumn: 44,
        },
      ],
    },
    {
      name: 'extensions / run 0 invalid 26',
      code: 'export { foo } from "./foo.js"',
      filename: 'files/input.ts',
      options: ['never'],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: 'Unexpected use of file extension "js" for "./foo.js"',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      name: 'extensions / run 0 invalid 27',
      code: 'export * from "./foo"',
      filename: 'files/input.ts',
      options: ['always'],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: 'Missing file extension for "./foo"',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      name: 'extensions / run 0 invalid 28',
      code: 'export * from "./foo.js"',
      filename: 'files/input.ts',
      options: ['never'],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: 'Unexpected use of file extension "js" for "./foo.js"',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      name: 'extensions / run 0 invalid 29',
      code: 'import foo from "@/ImNotAScopedModule.js"',
      filename: 'files/input.ts',
      options: ['never'],
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
            'Unexpected use of file extension "js" for "@/ImNotAScopedModule.js"',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      name: 'extensions / run 0 invalid 31',
      code: 'import * as test from "."',
      filename: 'files/internal-modules/test.js',
      options: ['ignorePackages'],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: 'Missing file extension for "."',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      name: 'extensions / run 0 invalid 32',
      code: 'import * as test from ".."',
      filename: 'files/internal-modules/plugins/plugin.js',
      options: ['ignorePackages'],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: 'Missing file extension for ".."',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      name: 'TypeScript / typescript: extensions ignore type-only / run 1 invalid 0 [upstream parser metadata: {"parser":"typescript"}]',
      code: 'import T from "./typescript-declare";',
      filename: 'files/input.ts',
      options: [
        'always',
        {
          ts: 'never',
          tsx: 'never',
          js: 'never',
          jsx: 'never',
        },
      ],
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
          message: 'Missing file extension for "./typescript-declare"',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      name: 'TypeScript / typescript: extensions ignore type-only / run 1 invalid 1 [upstream parser metadata: {"parser":"typescript"}]',
      code: 'export { MyType } from "./typescript-declare";',
      filename: 'files/input.ts',
      options: [
        'always',
        {
          ts: 'never',
          tsx: 'never',
          js: 'never',
          jsx: 'never',
        },
      ],
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
          message: 'Missing file extension for "./typescript-declare"',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 46,
        },
      ],
    },
    {
      name: 'TypeScript / typescript: extensions ignore type-only / run 1 invalid 2 [upstream parser metadata: {"parser":"typescript"}]',
      code: 'import type T from "./typescript-declare";',
      filename: 'files/input.ts',
      options: [
        'always',
        {
          ts: 'never',
          tsx: 'never',
          js: 'never',
          jsx: 'never',
          checkTypeImports: true,
        },
      ],
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
          message: 'Missing file extension for "./typescript-declare"',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      name: 'TypeScript / typescript: extensions ignore type-only / run 1 invalid 3 [upstream parser metadata: {"parser":"typescript"}]',
      code: 'export type { MyType } from "./typescript-declare";',
      filename: 'files/input.ts',
      options: [
        'always',
        {
          ts: 'never',
          tsx: 'never',
          js: 'never',
          jsx: 'never',
          checkTypeImports: true,
        },
      ],
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
          message: 'Missing file extension for "./typescript-declare"',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 51,
        },
      ],
    },
    {
      name: 'TypeScript / typescript: (with TS resolver) extensions are enforced for type imports/export when checkTypeImports is set / run 2 invalid 0 [upstream parser metadata: {"parser":"typescript"}]',
      code: 'import type { MyType } from "./typescript-declare";',
      filename: 'files/input.ts',
      options: [
        'always',
        {
          checkTypeImports: true,
        },
      ],
      settings: {
        'import/resolver': {
          typescript: {
            alwaysTryTypes: true,
          },
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
          message: 'Missing file extension "ts" for "./typescript-declare"',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 51,
        },
      ],
    },
    {
      name: 'TypeScript / typescript: (with TS resolver) extensions are enforced for type imports/export when checkTypeImports is set / run 2 invalid 1 [upstream parser metadata: {"parser":"typescript"}]',
      code: 'export type { MyType } from "./typescript-declare";',
      filename: 'files/input.ts',
      options: [
        'always',
        {
          checkTypeImports: true,
        },
      ],
      settings: {
        'import/resolver': {
          typescript: {
            alwaysTryTypes: true,
          },
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
          message: 'Missing file extension "ts" for "./typescript-declare"',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 51,
        },
      ],
    },
    {
      name: 'TypeScript / typescript: (with TS resolver) extensions are enforced for type imports/export when checkTypeImports is set / run 2 invalid 2 [upstream parser metadata: {"parser":"typescript"}]',
      code: "\n              import { ErrorMessage as UpstreamErrorMessage } from '@black-flag/core/util';\n\n              import { $instances } from 'rootverse+debug:src';\n              import { $exists } from 'rootverse+bfe:src/symbols';\n\n              import type { Entries } from 'type-fest';\n            ",
      filename: 'files/input.ts',
      options: [
        'always',
        {
          ignorePackages: true,
          checkTypeImports: true,
          pathGroupOverrides: [
            {
              pattern: 'rootverse{*,*/**}',
              action: 'enforce',
            },
            {
              pattern: 'universe{*,*/**}',
              action: 'ignore',
            },
          ],
        },
      ],
      settings: {
        'import/resolver': {
          typescript: {
            alwaysTryTypes: true,
          },
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
          message: 'Missing file extension for "rootverse+debug:src"',
          line: 4,
          column: 42,
          endLine: 4,
          endColumn: 63,
        },
        {
          messageId: '',
          message: 'Missing file extension for "rootverse+bfe:src/symbols"',
          line: 5,
          column: 39,
          endLine: 5,
          endColumn: 66,
        },
      ],
    },
  ],
});

// Preserved upstream inputs/assertions, not executable native expectations.
const skippedInvalidCases = [
  {
    name: 'extensions / run 0 invalid 30',
    code: "\n        import _ from 'lodash';\n        import m from '@test-scope/some-module/index.js';\n\n        import bar from './bar';\n      ",
    filename: 'files/input.ts',
    options: ['never'],
    settings: {
      'import/resolver': 'webpack',
      'import/external-module-folders': ['node_modules', 'symlinked-module'],
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
        message:
          'Unexpected use of file extension "js" for "@test-scope/some-module/index.js"',
        line: 3,
      },
    ],
  },
];
for (const item of skippedInvalidCases) {
  checkSkip(item);
  test.skip(`invalid ${item.name}: ${item.skip}`, () => {});
}

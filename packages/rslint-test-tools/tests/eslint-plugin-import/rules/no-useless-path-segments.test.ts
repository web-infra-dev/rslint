// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
new RuleTester(
  fixtureConfig(
    'internal/plugins/import/rules/no_useless_path_segments/testdata/upstream.txtar',
  ),
).run('no-useless-path-segments', {
  valid: [
    {
      name: 'node / run 0 valid 0',
      code: 'require("./../files/malformed.js")',
      filename: 'files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'node / run 0 valid 1',
      code: 'import "./malformed.js"',
      filename: 'files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'node / run 0 valid 2',
      code: 'import "./test-module"',
      filename: 'files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'node / run 0 valid 3',
      code: 'import "./bar/"',
      filename: 'files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'node / run 0 valid 4',
      code: 'import "."',
      filename: 'files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'node / run 0 valid 5',
      code: 'import ".."',
      filename: 'files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'node / run 0 valid 6',
      code: 'import fs from "fs"',
      filename: 'files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'node / run 0 valid 7',
      code: 'import "../index"',
      filename: 'files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'node / run 0 valid 8',
      code: 'import "../my-custom-index"',
      filename: 'files/foo.js',
      options: [
        {
          noUselessIndex: true,
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
      name: 'node / run 0 valid 9',
      code: 'import "./bar.js"',
      filename: 'files/foo.js',
      options: [
        {
          noUselessIndex: true,
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
      name: 'node / run 0 valid 10',
      code: 'import "./bar"',
      filename: 'files/foo.js',
      options: [
        {
          noUselessIndex: true,
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
      name: 'node / run 0 valid 11',
      code: 'import "./bar/"',
      filename: 'files/foo.js',
      options: [
        {
          noUselessIndex: true,
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
      name: 'node / run 0 valid 12',
      code: 'import "./malformed.js"',
      filename: 'files/foo.js',
      options: [
        {
          noUselessIndex: true,
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
      name: 'node / run 0 valid 13',
      code: 'import "./malformed"',
      filename: 'files/foo.js',
      options: [
        {
          noUselessIndex: true,
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
      name: 'node / run 0 valid 14',
      code: 'import "./importType"',
      filename: 'files/foo.js',
      options: [
        {
          noUselessIndex: true,
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
      name: 'node / run 0 valid 15 [upstream parser metadata: {"parser":"native"}]',
      code: 'import(".")',
      filename: 'files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'node / run 0 valid 16 [upstream parser metadata: {"parser":"native"}]',
      code: 'import("..")',
      filename: 'files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'node / run 0 valid 17 [upstream parser metadata: {"parser":"native"}]',
      code: 'import("fs").then(function(fs) {})',
      filename: 'files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'webpack / run 1 valid 0',
      code: 'require("./../files/malformed.js")',
      filename: 'files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'webpack / run 1 valid 1',
      code: 'import "./malformed.js"',
      filename: 'files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'webpack / run 1 valid 2',
      code: 'import "./test-module"',
      filename: 'files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'webpack / run 1 valid 3',
      code: 'import "./bar/"',
      filename: 'files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'webpack / run 1 valid 4',
      code: 'import "."',
      filename: 'files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'webpack / run 1 valid 5',
      code: 'import ".."',
      filename: 'files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'webpack / run 1 valid 6',
      code: 'import fs from "fs"',
      filename: 'files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'webpack / run 1 valid 7',
      code: 'import "../index"',
      filename: 'files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'webpack / run 1 valid 8',
      code: 'import "../my-custom-index"',
      filename: 'files/foo.js',
      options: [
        {
          noUselessIndex: true,
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
      name: 'webpack / run 1 valid 9',
      code: 'import "./bar.js"',
      filename: 'files/foo.js',
      options: [
        {
          noUselessIndex: true,
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
      name: 'webpack / run 1 valid 10',
      code: 'import "./bar"',
      filename: 'files/foo.js',
      options: [
        {
          noUselessIndex: true,
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
      name: 'webpack / run 1 valid 11',
      code: 'import "./bar/"',
      filename: 'files/foo.js',
      options: [
        {
          noUselessIndex: true,
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
      name: 'webpack / run 1 valid 12',
      code: 'import "./malformed.js"',
      filename: 'files/foo.js',
      options: [
        {
          noUselessIndex: true,
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
      name: 'webpack / run 1 valid 13',
      code: 'import "./malformed"',
      filename: 'files/foo.js',
      options: [
        {
          noUselessIndex: true,
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
      name: 'webpack / run 1 valid 14',
      code: 'import "./importType"',
      filename: 'files/foo.js',
      options: [
        {
          noUselessIndex: true,
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
      name: 'webpack / run 1 valid 15 [upstream parser metadata: {"parser":"native"}]',
      code: 'import(".")',
      filename: 'files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'webpack / run 1 valid 16 [upstream parser metadata: {"parser":"native"}]',
      code: 'import("..")',
      filename: 'files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'webpack / run 1 valid 17 [upstream parser metadata: {"parser":"native"}]',
      code: 'import("fs").then(function(fs) {})',
      filename: 'files/foo.js',
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
      name: 'node / run 0 invalid 0',
      code: 'require("./../files/malformed.js")',
      filename: 'files/foo.js',
      options: [
        {
          commonjs: true,
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
            'Useless path segments for "./../files/malformed.js", should be "../files/malformed.js"',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 34,
        },
      ],
      output: 'require("./malformed.js")',
    },
    {
      name: 'node / run 0 invalid 1',
      code: 'require("./../files/malformed")',
      filename: 'files/foo.js',
      options: [
        {
          commonjs: true,
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
            'Useless path segments for "./../files/malformed", should be "../files/malformed"',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 31,
        },
      ],
      output: 'require("./malformed")',
    },
    {
      name: 'node / run 0 invalid 2',
      code: 'require("../files/malformed.js")',
      filename: 'files/foo.js',
      options: [
        {
          commonjs: true,
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
            'Useless path segments for "../files/malformed.js", should be "./malformed.js"',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 32,
        },
      ],
      output: 'require("./malformed.js")',
    },
    {
      name: 'node / run 0 invalid 3',
      code: 'require("../files/malformed")',
      filename: 'files/foo.js',
      options: [
        {
          commonjs: true,
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
            'Useless path segments for "../files/malformed", should be "./malformed"',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 29,
        },
      ],
      output: 'require("./malformed")',
    },
    {
      name: 'node / run 0 invalid 4',
      code: 'require("./test-module/")',
      filename: 'files/foo.js',
      options: [
        {
          commonjs: true,
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
            'Useless path segments for "./test-module/", should be "./test-module"',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 25,
        },
      ],
      output: 'require("./test-module")',
    },
    {
      name: 'node / run 0 invalid 5',
      code: 'require("./")',
      filename: 'files/foo.js',
      options: [
        {
          commonjs: true,
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
          message: 'Useless path segments for "./", should be "."',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 13,
        },
      ],
      output: 'require(".")',
    },
    {
      name: 'node / run 0 invalid 6',
      code: 'require("../")',
      filename: 'files/foo.js',
      options: [
        {
          commonjs: true,
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
          message: 'Useless path segments for "../", should be ".."',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 14,
        },
      ],
      output: 'require("..")',
    },
    {
      name: 'node / run 0 invalid 7',
      code: 'require("./deep//a")',
      filename: 'files/foo.js',
      options: [
        {
          commonjs: true,
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
            'Useless path segments for "./deep//a", should be "./deep/a"',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 20,
        },
      ],
      output: 'require("./deep/a")',
    },
    {
      name: 'node / run 0 invalid 8',
      code: 'require("./bar/index.js")',
      filename: 'files/foo.js',
      options: [
        {
          commonjs: true,
          noUselessIndex: true,
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
            'Useless path segments for "./bar/index.js", should be "./bar/"',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 25,
        },
      ],
      output: 'require("./bar/")',
    },
    {
      name: 'node / run 0 invalid 9',
      code: 'require("./bar/index")',
      filename: 'files/foo.js',
      options: [
        {
          commonjs: true,
          noUselessIndex: true,
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
            'Useless path segments for "./bar/index", should be "./bar/"',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 22,
        },
      ],
      output: 'require("./bar/")',
    },
    {
      name: 'node / run 0 invalid 10',
      code: 'require("./importPath/")',
      filename: 'files/foo.js',
      options: [
        {
          commonjs: true,
          noUselessIndex: true,
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
            'Useless path segments for "./importPath/", should be "./importPath"',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 24,
        },
      ],
      output: 'require("./importPath")',
    },
    {
      name: 'node / run 0 invalid 11',
      code: 'require("./importPath/index.js")',
      filename: 'files/foo.js',
      options: [
        {
          commonjs: true,
          noUselessIndex: true,
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
            'Useless path segments for "./importPath/index.js", should be "./importPath"',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 32,
        },
      ],
      output: 'require("./importPath")',
    },
    {
      name: 'node / run 0 invalid 12',
      code: 'require("./importType/index")',
      filename: 'files/foo.js',
      options: [
        {
          commonjs: true,
          noUselessIndex: true,
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
            'Useless path segments for "./importType/index", should be "./importType"',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 29,
        },
      ],
      output: 'require("./importType")',
    },
    {
      name: 'node / run 0 invalid 13',
      code: 'require("./index")',
      filename: 'files/foo.js',
      options: [
        {
          commonjs: true,
          noUselessIndex: true,
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
          message: 'Useless path segments for "./index", should be "."',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 18,
        },
      ],
      output: 'require(".")',
    },
    {
      name: 'node / run 0 invalid 14',
      code: 'require("../index")',
      filename: 'files/foo.js',
      options: [
        {
          commonjs: true,
          noUselessIndex: true,
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
          message: 'Useless path segments for "../index", should be ".."',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 19,
        },
      ],
      output: 'require("..")',
    },
    {
      name: 'node / run 0 invalid 15',
      code: 'require("../index.js")',
      filename: 'files/foo.js',
      options: [
        {
          commonjs: true,
          noUselessIndex: true,
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
          message: 'Useless path segments for "../index.js", should be ".."',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 22,
        },
      ],
      output: 'require("..")',
    },
    {
      name: 'node / run 0 invalid 16',
      code: 'import "./../files/malformed.js"',
      filename: 'files/foo.js',
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
            'Useless path segments for "./../files/malformed.js", should be "../files/malformed.js"',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 33,
        },
      ],
      output: 'import "./malformed.js"',
    },
    {
      name: 'node / run 0 invalid 17',
      code: 'import "./../files/malformed"',
      filename: 'files/foo.js',
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
            'Useless path segments for "./../files/malformed", should be "../files/malformed"',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 30,
        },
      ],
      output: 'import "./malformed"',
    },
    {
      name: 'node / run 0 invalid 18',
      code: 'import "../files/malformed.js"',
      filename: 'files/foo.js',
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
            'Useless path segments for "../files/malformed.js", should be "./malformed.js"',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 31,
        },
      ],
      output: 'import "./malformed.js"',
    },
    {
      name: 'node / run 0 invalid 19',
      code: 'import "../files/malformed"',
      filename: 'files/foo.js',
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
            'Useless path segments for "../files/malformed", should be "./malformed"',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 28,
        },
      ],
      output: 'import "./malformed"',
    },
    {
      name: 'node / run 0 invalid 20',
      code: 'import "./test-module/"',
      filename: 'files/foo.js',
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
            'Useless path segments for "./test-module/", should be "./test-module"',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 24,
        },
      ],
      output: 'import "./test-module"',
    },
    {
      name: 'node / run 0 invalid 21',
      code: 'import "./"',
      filename: 'files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: 'Useless path segments for "./", should be "."',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 12,
        },
      ],
      output: 'import "."',
    },
    {
      name: 'node / run 0 invalid 22',
      code: 'import "../"',
      filename: 'files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: 'Useless path segments for "../", should be ".."',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 13,
        },
      ],
      output: 'import ".."',
    },
    {
      name: 'node / run 0 invalid 23',
      code: 'import "./deep//a"',
      filename: 'files/foo.js',
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
            'Useless path segments for "./deep//a", should be "./deep/a"',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 19,
        },
      ],
      output: 'import "./deep/a"',
    },
    {
      name: 'node / run 0 invalid 24',
      code: 'import "./bar/index.js"',
      filename: 'files/foo.js',
      options: [
        {
          noUselessIndex: true,
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
            'Useless path segments for "./bar/index.js", should be "./bar/"',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 24,
        },
      ],
      output: 'import "./bar/"',
    },
    {
      name: 'node / run 0 invalid 25',
      code: 'import "./bar/index"',
      filename: 'files/foo.js',
      options: [
        {
          noUselessIndex: true,
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
            'Useless path segments for "./bar/index", should be "./bar/"',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 21,
        },
      ],
      output: 'import "./bar/"',
    },
    {
      name: 'node / run 0 invalid 26',
      code: 'import "./importPath/"',
      filename: 'files/foo.js',
      options: [
        {
          noUselessIndex: true,
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
            'Useless path segments for "./importPath/", should be "./importPath"',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 23,
        },
      ],
      output: 'import "./importPath"',
    },
    {
      name: 'node / run 0 invalid 27',
      code: 'import "./importPath/index.js"',
      filename: 'files/foo.js',
      options: [
        {
          noUselessIndex: true,
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
            'Useless path segments for "./importPath/index.js", should be "./importPath"',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 31,
        },
      ],
      output: 'import "./importPath"',
    },
    {
      name: 'node / run 0 invalid 28',
      code: 'import "./importPath/index"',
      filename: 'files/foo.js',
      options: [
        {
          noUselessIndex: true,
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
            'Useless path segments for "./importPath/index", should be "./importPath"',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 28,
        },
      ],
      output: 'import "./importPath"',
    },
    {
      name: 'node / run 0 invalid 29',
      code: 'import "./index"',
      filename: 'files/foo.js',
      options: [
        {
          noUselessIndex: true,
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
          message: 'Useless path segments for "./index", should be "."',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 17,
        },
      ],
      output: 'import "."',
    },
    {
      name: 'node / run 0 invalid 30',
      code: 'import "../index"',
      filename: 'files/foo.js',
      options: [
        {
          noUselessIndex: true,
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
          message: 'Useless path segments for "../index", should be ".."',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 18,
        },
      ],
      output: 'import ".."',
    },
    {
      name: 'node / run 0 invalid 31',
      code: 'import "../index.js"',
      filename: 'files/foo.js',
      options: [
        {
          noUselessIndex: true,
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
          message: 'Useless path segments for "../index.js", should be ".."',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 21,
        },
      ],
      output: 'import ".."',
    },
    {
      name: 'node / run 0 invalid 32 [upstream parser metadata: {"parser":"native"}]',
      code: 'import("./")',
      filename: 'files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: 'Useless path segments for "./", should be "."',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 12,
        },
      ],
      output: 'import(".")',
    },
    {
      name: 'node / run 0 invalid 33 [upstream parser metadata: {"parser":"native"}]',
      code: 'import("../")',
      filename: 'files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: 'Useless path segments for "../", should be ".."',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 13,
        },
      ],
      output: 'import("..")',
    },
    {
      name: 'node / run 0 invalid 34 [upstream parser metadata: {"parser":"native"}]',
      code: 'import("./deep//a")',
      filename: 'files/foo.js',
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
            'Useless path segments for "./deep//a", should be "./deep/a"',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 19,
        },
      ],
      output: 'import("./deep/a")',
    },
    {
      name: 'webpack / run 1 invalid 0',
      code: 'require("./../files/malformed.js")',
      filename: 'files/foo.js',
      options: [
        {
          commonjs: true,
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
            'Useless path segments for "./../files/malformed.js", should be "../files/malformed.js"',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 34,
        },
      ],
      output: 'require("./malformed.js")',
    },
    {
      name: 'webpack / run 1 invalid 1',
      code: 'require("./../files/malformed")',
      filename: 'files/foo.js',
      options: [
        {
          commonjs: true,
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
            'Useless path segments for "./../files/malformed", should be "../files/malformed"',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 31,
        },
      ],
      output: 'require("./malformed")',
    },
    {
      name: 'webpack / run 1 invalid 2',
      code: 'require("../files/malformed.js")',
      filename: 'files/foo.js',
      options: [
        {
          commonjs: true,
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
            'Useless path segments for "../files/malformed.js", should be "./malformed.js"',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 32,
        },
      ],
      output: 'require("./malformed.js")',
    },
    {
      name: 'webpack / run 1 invalid 3',
      code: 'require("../files/malformed")',
      filename: 'files/foo.js',
      options: [
        {
          commonjs: true,
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
            'Useless path segments for "../files/malformed", should be "./malformed"',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 29,
        },
      ],
      output: 'require("./malformed")',
    },
    {
      name: 'webpack / run 1 invalid 4',
      code: 'require("./test-module/")',
      filename: 'files/foo.js',
      options: [
        {
          commonjs: true,
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
            'Useless path segments for "./test-module/", should be "./test-module"',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 25,
        },
      ],
      output: 'require("./test-module")',
    },
    {
      name: 'webpack / run 1 invalid 5',
      code: 'require("./")',
      filename: 'files/foo.js',
      options: [
        {
          commonjs: true,
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
          message: 'Useless path segments for "./", should be "."',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 13,
        },
      ],
      output: 'require(".")',
    },
    {
      name: 'webpack / run 1 invalid 6',
      code: 'require("../")',
      filename: 'files/foo.js',
      options: [
        {
          commonjs: true,
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
          message: 'Useless path segments for "../", should be ".."',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 14,
        },
      ],
      output: 'require("..")',
    },
    {
      name: 'webpack / run 1 invalid 7',
      code: 'require("./deep//a")',
      filename: 'files/foo.js',
      options: [
        {
          commonjs: true,
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
            'Useless path segments for "./deep//a", should be "./deep/a"',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 20,
        },
      ],
      output: 'require("./deep/a")',
    },
    {
      name: 'webpack / run 1 invalid 8',
      code: 'require("./bar/index.js")',
      filename: 'files/foo.js',
      options: [
        {
          commonjs: true,
          noUselessIndex: true,
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
            'Useless path segments for "./bar/index.js", should be "./bar/"',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 25,
        },
      ],
      output: 'require("./bar/")',
    },
    {
      name: 'webpack / run 1 invalid 9',
      code: 'require("./bar/index")',
      filename: 'files/foo.js',
      options: [
        {
          commonjs: true,
          noUselessIndex: true,
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
            'Useless path segments for "./bar/index", should be "./bar/"',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 22,
        },
      ],
      output: 'require("./bar/")',
    },
    {
      name: 'webpack / run 1 invalid 10',
      code: 'require("./importPath/")',
      filename: 'files/foo.js',
      options: [
        {
          commonjs: true,
          noUselessIndex: true,
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
            'Useless path segments for "./importPath/", should be "./importPath"',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 24,
        },
      ],
      output: 'require("./importPath")',
    },
    {
      name: 'webpack / run 1 invalid 11',
      code: 'require("./importPath/index.js")',
      filename: 'files/foo.js',
      options: [
        {
          commonjs: true,
          noUselessIndex: true,
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
            'Useless path segments for "./importPath/index.js", should be "./importPath"',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 32,
        },
      ],
      output: 'require("./importPath")',
    },
    {
      name: 'webpack / run 1 invalid 12',
      code: 'require("./importType/index")',
      filename: 'files/foo.js',
      options: [
        {
          commonjs: true,
          noUselessIndex: true,
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
            'Useless path segments for "./importType/index", should be "./importType"',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 29,
        },
      ],
      output: 'require("./importType")',
    },
    {
      name: 'webpack / run 1 invalid 13',
      code: 'require("./index")',
      filename: 'files/foo.js',
      options: [
        {
          commonjs: true,
          noUselessIndex: true,
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
          message: 'Useless path segments for "./index", should be "."',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 18,
        },
      ],
      output: 'require(".")',
    },
    {
      name: 'webpack / run 1 invalid 14',
      code: 'require("../index")',
      filename: 'files/foo.js',
      options: [
        {
          commonjs: true,
          noUselessIndex: true,
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
          message: 'Useless path segments for "../index", should be ".."',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 19,
        },
      ],
      output: 'require("..")',
    },
    {
      name: 'webpack / run 1 invalid 15',
      code: 'require("../index.js")',
      filename: 'files/foo.js',
      options: [
        {
          commonjs: true,
          noUselessIndex: true,
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
          message: 'Useless path segments for "../index.js", should be ".."',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 22,
        },
      ],
      output: 'require("..")',
    },
    {
      name: 'webpack / run 1 invalid 16',
      code: 'import "./../files/malformed.js"',
      filename: 'files/foo.js',
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
            'Useless path segments for "./../files/malformed.js", should be "../files/malformed.js"',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 33,
        },
      ],
      output: 'import "./malformed.js"',
    },
    {
      name: 'webpack / run 1 invalid 17',
      code: 'import "./../files/malformed"',
      filename: 'files/foo.js',
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
            'Useless path segments for "./../files/malformed", should be "../files/malformed"',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 30,
        },
      ],
      output: 'import "./malformed"',
    },
    {
      name: 'webpack / run 1 invalid 18',
      code: 'import "../files/malformed.js"',
      filename: 'files/foo.js',
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
            'Useless path segments for "../files/malformed.js", should be "./malformed.js"',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 31,
        },
      ],
      output: 'import "./malformed.js"',
    },
    {
      name: 'webpack / run 1 invalid 19',
      code: 'import "../files/malformed"',
      filename: 'files/foo.js',
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
            'Useless path segments for "../files/malformed", should be "./malformed"',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 28,
        },
      ],
      output: 'import "./malformed"',
    },
    {
      name: 'webpack / run 1 invalid 20',
      code: 'import "./test-module/"',
      filename: 'files/foo.js',
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
            'Useless path segments for "./test-module/", should be "./test-module"',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 24,
        },
      ],
      output: 'import "./test-module"',
    },
    {
      name: 'webpack / run 1 invalid 21',
      code: 'import "./"',
      filename: 'files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: 'Useless path segments for "./", should be "."',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 12,
        },
      ],
      output: 'import "."',
    },
    {
      name: 'webpack / run 1 invalid 22',
      code: 'import "../"',
      filename: 'files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: 'Useless path segments for "../", should be ".."',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 13,
        },
      ],
      output: 'import ".."',
    },
    {
      name: 'webpack / run 1 invalid 23',
      code: 'import "./deep//a"',
      filename: 'files/foo.js',
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
            'Useless path segments for "./deep//a", should be "./deep/a"',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 19,
        },
      ],
      output: 'import "./deep/a"',
    },
    {
      name: 'webpack / run 1 invalid 24',
      code: 'import "./bar/index.js"',
      filename: 'files/foo.js',
      options: [
        {
          noUselessIndex: true,
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
            'Useless path segments for "./bar/index.js", should be "./bar/"',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 24,
        },
      ],
      output: 'import "./bar/"',
    },
    {
      name: 'webpack / run 1 invalid 25',
      code: 'import "./bar/index"',
      filename: 'files/foo.js',
      options: [
        {
          noUselessIndex: true,
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
            'Useless path segments for "./bar/index", should be "./bar/"',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 21,
        },
      ],
      output: 'import "./bar/"',
    },
    {
      name: 'webpack / run 1 invalid 26',
      code: 'import "./importPath/"',
      filename: 'files/foo.js',
      options: [
        {
          noUselessIndex: true,
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
            'Useless path segments for "./importPath/", should be "./importPath"',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 23,
        },
      ],
      output: 'import "./importPath"',
    },
    {
      name: 'webpack / run 1 invalid 27',
      code: 'import "./importPath/index.js"',
      filename: 'files/foo.js',
      options: [
        {
          noUselessIndex: true,
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
            'Useless path segments for "./importPath/index.js", should be "./importPath"',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 31,
        },
      ],
      output: 'import "./importPath"',
    },
    {
      name: 'webpack / run 1 invalid 28',
      code: 'import "./importPath/index"',
      filename: 'files/foo.js',
      options: [
        {
          noUselessIndex: true,
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
            'Useless path segments for "./importPath/index", should be "./importPath"',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 28,
        },
      ],
      output: 'import "./importPath"',
    },
    {
      name: 'webpack / run 1 invalid 29',
      code: 'import "./index"',
      filename: 'files/foo.js',
      options: [
        {
          noUselessIndex: true,
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
          message: 'Useless path segments for "./index", should be "."',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 17,
        },
      ],
      output: 'import "."',
    },
    {
      name: 'webpack / run 1 invalid 30',
      code: 'import "../index"',
      filename: 'files/foo.js',
      options: [
        {
          noUselessIndex: true,
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
          message: 'Useless path segments for "../index", should be ".."',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 18,
        },
      ],
      output: 'import ".."',
    },
    {
      name: 'webpack / run 1 invalid 31',
      code: 'import "../index.js"',
      filename: 'files/foo.js',
      options: [
        {
          noUselessIndex: true,
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
          message: 'Useless path segments for "../index.js", should be ".."',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 21,
        },
      ],
      output: 'import ".."',
    },
    {
      name: 'webpack / run 1 invalid 32 [upstream parser metadata: {"parser":"native"}]',
      code: 'import("./")',
      filename: 'files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: 'Useless path segments for "./", should be "."',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 12,
        },
      ],
      output: 'import(".")',
    },
    {
      name: 'webpack / run 1 invalid 33 [upstream parser metadata: {"parser":"native"}]',
      code: 'import("../")',
      filename: 'files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: 'Useless path segments for "../", should be ".."',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 13,
        },
      ],
      output: 'import("..")',
    },
    {
      name: 'webpack / run 1 invalid 34 [upstream parser metadata: {"parser":"native"}]',
      code: 'import("./deep//a")',
      filename: 'files/foo.js',
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
            'Useless path segments for "./deep//a", should be "./deep/a"',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 19,
        },
      ],
      output: 'import("./deep/a")',
    },
  ],
});

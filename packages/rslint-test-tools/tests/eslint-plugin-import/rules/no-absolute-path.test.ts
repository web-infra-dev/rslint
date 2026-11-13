// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, resolveFixturePaths } from '../config';
new RuleTester(fixtureConfig()).run('no-absolute-path', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: 'import _ from "lodash"',
    },
    {
      name: 'run 0 valid 1',
      code: 'import find from "lodash.find"',
    },
    {
      name: 'run 0 valid 2',
      code: 'import foo from "./foo"',
    },
    {
      name: 'run 0 valid 3',
      code: 'import foo from "../foo"',
    },
    {
      name: 'run 0 valid 4',
      code: 'import foo from "foo"',
    },
    {
      name: 'run 0 valid 5',
      code: 'import foo from "./"',
    },
    {
      name: 'run 0 valid 6',
      code: 'import foo from "@scope/foo"',
    },
    {
      name: 'run 0 valid 7',
      code: 'var _ = require("lodash")',
    },
    {
      name: 'run 0 valid 8',
      code: 'var find = require("lodash.find")',
    },
    {
      name: 'run 0 valid 9',
      code: 'var foo = require("./foo")',
    },
    {
      name: 'run 0 valid 10',
      code: 'var foo = require("../foo")',
    },
    {
      name: 'run 0 valid 11',
      code: 'var foo = require("foo")',
    },
    {
      name: 'run 0 valid 12',
      code: 'var foo = require("./")',
    },
    {
      name: 'run 0 valid 13',
      code: 'var foo = require("@scope/foo")',
    },
    {
      name: 'run 0 valid 14',
      code: 'import events from "events"',
    },
    {
      name: 'run 0 valid 15',
      code: 'import path from "path"',
    },
    {
      name: 'run 0 valid 16',
      code: 'var events = require("events")',
    },
    {
      name: 'run 0 valid 17',
      code: 'var path = require("path")',
    },
    {
      name: 'run 0 valid 18',
      code: 'import path from "path";import events from "events"',
    },
    {
      name: 'run 0 valid 19',
      code: 'import path from "path"',
      options: [
        {
          amd: true,
        },
      ],
    },
    {
      name: 'run 0 valid 20',
      code: 'require(["/some/path"], function (f) { /* ... */ })',
    },
    {
      name: 'run 0 valid 21',
      code: 'define(["/some/path"], function (f) { /* ... */ })',
    },
    {
      name: 'run 0 valid 22',
      code: 'require(["./some/path"], function (f) { /* ... */ })',
      options: [
        {
          amd: true,
        },
      ],
    },
    {
      name: 'run 0 valid 23',
      code: 'define(["./some/path"], function (f) { /* ... */ })',
      options: [
        {
          amd: true,
        },
      ],
    },
    {
      name: 'run 0 valid 24',
      code: "import _ from 'lodash';",
    },
    {
      name: 'run 0 valid 25',
      code: "import foo from 'foo';",
    },
    {
      name: 'run 0 valid 26',
      code: "import foo from './foo';",
    },
    {
      name: 'run 0 valid 27',
      code: "var _ = require('lodash');",
    },
    {
      name: 'run 0 valid 28',
      code: "var foo = require('foo');",
    },
    {
      name: 'run 0 valid 29',
      code: "var foo = require('./foo');",
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      create(root: string) {
        const code = resolveFixturePaths('import f from "<fixture>/foo"', root);
        return {
          code: code,
          filename: 'foo/bar/index.js',
          errors: [
            {
              messageId: '',
              message: 'Do not import modules using an absolute path',
              line: 1,
              column: 15,
              endLine: 1,
              endColumn: 21 + 1 * root.length,
            },
          ],
          output: resolveFixturePaths('import f from ".."', root),
        };
      },
    },
    {
      name: 'run 0 invalid 1',
      create(root: string) {
        const code = resolveFixturePaths(
          'import f from "<fixture>/foo/bar/baz.js"',
          root,
        );
        return {
          code: code,
          filename: 'foo/bar/index.js',
          errors: [
            {
              messageId: '',
              message: 'Do not import modules using an absolute path',
              line: 1,
              column: 15,
              endLine: 1,
              endColumn: 32 + 1 * root.length,
            },
          ],
          output: resolveFixturePaths('import f from "./baz.js"', root),
        };
      },
    },
    {
      name: 'run 0 invalid 2',
      create(root: string) {
        const code = resolveFixturePaths(
          'import f from "<fixture>/foo/path"',
          root,
        );
        return {
          code: code,
          filename: 'foo/bar/index.js',
          errors: [
            {
              messageId: '',
              message: 'Do not import modules using an absolute path',
              line: 1,
              column: 15,
              endLine: 1,
              endColumn: 26 + 1 * root.length,
            },
          ],
          output: resolveFixturePaths('import f from "../path"', root),
        };
      },
    },
    {
      name: 'run 0 invalid 3',
      create(root: string) {
        const code = resolveFixturePaths(
          'import f from "<fixture>/some/path"',
          root,
        );
        return {
          code: code,
          filename: 'foo/bar/index.js',
          errors: [
            {
              messageId: '',
              message: 'Do not import modules using an absolute path',
              line: 1,
              column: 15,
              endLine: 1,
              endColumn: 27 + 1 * root.length,
            },
          ],
          output: resolveFixturePaths('import f from "../../some/path"', root),
        };
      },
    },
    {
      name: 'run 0 invalid 4',
      create(root: string) {
        const code = resolveFixturePaths(
          'import f from "<fixture>/some/path"',
          root,
        );
        return {
          code: code,
          filename: 'foo/bar/index.js',
          options: [
            {
              amd: true,
            },
          ],
          errors: [
            {
              messageId: '',
              message: 'Do not import modules using an absolute path',
              line: 1,
              column: 15,
              endLine: 1,
              endColumn: 27 + 1 * root.length,
            },
          ],
          output: resolveFixturePaths('import f from "../../some/path"', root),
        };
      },
    },
    {
      name: 'run 0 invalid 5',
      create(root: string) {
        const code = resolveFixturePaths(
          'var f = require("<fixture>/foo")',
          root,
        );
        return {
          code: code,
          filename: 'foo/bar/index.js',
          errors: [
            {
              messageId: '',
              message: 'Do not import modules using an absolute path',
              line: 1,
              column: 17,
              endLine: 1,
              endColumn: 23 + 1 * root.length,
            },
          ],
          output: resolveFixturePaths('var f = require("..")', root),
        };
      },
    },
    {
      name: 'run 0 invalid 6',
      create(root: string) {
        const code = resolveFixturePaths(
          'var f = require("<fixture>/foo/path")',
          root,
        );
        return {
          code: code,
          filename: 'foo/bar/index.js',
          errors: [
            {
              messageId: '',
              message: 'Do not import modules using an absolute path',
              line: 1,
              column: 17,
              endLine: 1,
              endColumn: 28 + 1 * root.length,
            },
          ],
          output: resolveFixturePaths('var f = require("../path")', root),
        };
      },
    },
    {
      name: 'run 0 invalid 7',
      create(root: string) {
        const code = resolveFixturePaths(
          'var f = require("<fixture>/some/path")',
          root,
        );
        return {
          code: code,
          filename: 'foo/bar/index.js',
          errors: [
            {
              messageId: '',
              message: 'Do not import modules using an absolute path',
              line: 1,
              column: 17,
              endLine: 1,
              endColumn: 29 + 1 * root.length,
            },
          ],
          output: resolveFixturePaths(
            'var f = require("../../some/path")',
            root,
          ),
        };
      },
    },
    {
      name: 'run 0 invalid 8',
      create(root: string) {
        const code = resolveFixturePaths(
          'var f = require("<fixture>/some/path")',
          root,
        );
        return {
          code: code,
          filename: 'foo/bar/index.js',
          options: [
            {
              amd: true,
            },
          ],
          errors: [
            {
              messageId: '',
              message: 'Do not import modules using an absolute path',
              line: 1,
              column: 17,
              endLine: 1,
              endColumn: 29 + 1 * root.length,
            },
          ],
          output: resolveFixturePaths(
            'var f = require("../../some/path")',
            root,
          ),
        };
      },
    },
    {
      name: 'run 0 invalid 9',
      create(root: string) {
        const code = resolveFixturePaths(
          'require(["<fixture>/some/path"], function (f) { /* ... */ })',
          root,
        );
        return {
          code: code,
          filename: 'foo/bar/index.js',
          options: [
            {
              amd: true,
            },
          ],
          errors: [
            {
              messageId: '',
              message: 'Do not import modules using an absolute path',
              line: 1,
              column: 10,
              endLine: 1,
              endColumn: 22 + 1 * root.length,
            },
          ],
          output: resolveFixturePaths(
            'require(["../../some/path"], function (f) { /* ... */ })',
            root,
          ),
        };
      },
    },
    {
      name: 'run 0 invalid 10',
      create(root: string) {
        const code = resolveFixturePaths(
          'define(["<fixture>/some/path"], function (f) { /* ... */ })',
          root,
        );
        return {
          code: code,
          filename: 'foo/bar/index.js',
          options: [
            {
              amd: true,
            },
          ],
          errors: [
            {
              messageId: '',
              message: 'Do not import modules using an absolute path',
              line: 1,
              column: 9,
              endLine: 1,
              endColumn: 21 + 1 * root.length,
            },
          ],
          output: resolveFixturePaths(
            'define(["../../some/path"], function (f) { /* ... */ })',
            root,
          ),
        };
      },
    },
    {
      name: 'run 0 invalid 11',
      create(root: string) {
        const code = resolveFixturePaths(
          "import f from '<fixture>/foo';",
          root,
        );
        return {
          code: code,
          filename: 'foo/bar/index.js',
          errors: [
            {
              messageId: '',
              message: 'Do not import modules using an absolute path',
              line: 1,
              column: 15,
              endLine: 1,
              endColumn: 21 + 1 * root.length,
            },
          ],
          output: resolveFixturePaths('import f from "..";', root),
        };
      },
    },
    {
      name: 'run 0 invalid 12',
      create(root: string) {
        const code = resolveFixturePaths(
          "import f from '<fixture>/some/path';",
          root,
        );
        return {
          code: code,
          filename: 'foo/bar/index.js',
          errors: [
            {
              messageId: '',
              message: 'Do not import modules using an absolute path',
              line: 1,
              column: 15,
              endLine: 1,
              endColumn: 27 + 1 * root.length,
            },
          ],
          output: resolveFixturePaths('import f from "../../some/path";', root),
        };
      },
    },
    {
      name: 'run 0 invalid 13',
      create(root: string) {
        const code = resolveFixturePaths(
          "var f = require('<fixture>/foo');",
          root,
        );
        return {
          code: code,
          filename: 'foo/bar/index.js',
          errors: [
            {
              messageId: '',
              message: 'Do not import modules using an absolute path',
              line: 1,
              column: 17,
              endLine: 1,
              endColumn: 23 + 1 * root.length,
            },
          ],
          output: resolveFixturePaths('var f = require("..");', root),
        };
      },
    },
    {
      name: 'run 0 invalid 14',
      create(root: string) {
        const code = resolveFixturePaths(
          "var f = require('<fixture>/some/path');",
          root,
        );
        return {
          code: code,
          filename: 'foo/bar/index.js',
          errors: [
            {
              messageId: '',
              message: 'Do not import modules using an absolute path',
              line: 1,
              column: 17,
              endLine: 1,
              endColumn: 29 + 1 * root.length,
            },
          ],
          output: resolveFixturePaths(
            'var f = require("../../some/path");',
            root,
          ),
        };
      },
    },
    {
      name: 'run 0 invalid 15',
      create(root: string) {
        const code = resolveFixturePaths(
          "define(['<fixture>/foo'], function (foo) { /*...*/ })\nrequire(['<fixture>/foo'], function (foo) { /*...*/ })\nconst foo = require('<fixture>/foo')",
          root,
        );
        return {
          code: code,
          filename: 'foo/bar/index.js',
          options: [
            {
              commonjs: false,
              amd: true,
            },
          ],
          errors: [
            {
              messageId: '',
              message: 'Do not import modules using an absolute path',
              line: 1,
              column: 9,
              endLine: 1,
              endColumn: 15 + 1 * root.length,
            },
            {
              messageId: '',
              message: 'Do not import modules using an absolute path',
              line: 2,
              column: 10,
              endLine: 2,
              endColumn: 16 + 1 * root.length,
            },
          ],
          output: resolveFixturePaths(
            'define([".."], function (foo) { /*...*/ })\nrequire([".."], function (foo) { /*...*/ })\nconst foo = require(\'<fixture>/foo\')',
            root,
          ),
        };
      },
    },
  ],
});

// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
new RuleTester(fixtureConfig()).run('no-self-import', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: 'import _ from "lodash"',
      filename: 'files/no-self-import.js',
    },
    {
      name: 'run 0 valid 1',
      code: 'import find from "lodash.find"',
      filename: 'files/no-self-import.js',
    },
    {
      name: 'run 0 valid 2',
      code: 'import foo from "./foo"',
      filename: 'files/no-self-import.js',
    },
    {
      name: 'run 0 valid 3',
      code: 'import foo from "../foo"',
      filename: 'files/no-self-import.js',
    },
    {
      name: 'run 0 valid 4',
      code: 'import foo from "foo"',
      filename: 'files/no-self-import.js',
    },
    {
      name: 'run 0 valid 5',
      code: 'import foo from "./"',
      filename: 'files/no-self-import.js',
    },
    {
      name: 'run 0 valid 6',
      code: 'import foo from "@scope/foo"',
      filename: 'files/no-self-import.js',
    },
    {
      name: 'run 0 valid 7',
      code: 'var _ = require("lodash")',
      filename: 'files/no-self-import.js',
    },
    {
      name: 'run 0 valid 8',
      code: 'var find = require("lodash.find")',
      filename: 'files/no-self-import.js',
    },
    {
      name: 'run 0 valid 9',
      code: 'var foo = require("./foo")',
      filename: 'files/no-self-import.js',
    },
    {
      name: 'run 0 valid 10',
      code: 'var foo = require("../foo")',
      filename: 'files/no-self-import.js',
    },
    {
      name: 'run 0 valid 11',
      code: 'var foo = require("foo")',
      filename: 'files/no-self-import.js',
    },
    {
      name: 'run 0 valid 12',
      code: 'var foo = require("./")',
      filename: 'files/no-self-import.js',
    },
    {
      name: 'run 0 valid 13',
      code: 'var foo = require("@scope/foo")',
      filename: 'files/no-self-import.js',
    },
    {
      name: 'run 0 valid 14',
      code: 'var bar = require("./bar/index")',
      filename: 'files/no-self-import.js',
    },
    {
      name: 'run 0 valid 15',
      code: 'var bar = require("./bar")',
      filename: 'files/bar/index.js',
    },
    {
      name: 'run 0 valid 16',
      code: 'var bar = require("./bar")',
      filename: '<text>',
      skip: 'The native API ignores the virtual <text> filename; this case has no admitted JavaScript or TypeScript lint target.',
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: 'import bar from "./no-self-import"',
      filename: 'files/no-self-import.js',
      errors: [
        {
          messageId: 'import/no-self-import',
          message: 'Module imports itself.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      name: 'run 0 invalid 1',
      code: 'var bar = require("./no-self-import")',
      filename: 'files/no-self-import.js',
      errors: [
        {
          messageId: 'import/no-self-import',
          message: 'Module imports itself.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      name: 'run 0 invalid 2',
      code: 'var bar = require("./no-self-import.js")',
      filename: 'files/no-self-import.js',
      errors: [
        {
          messageId: 'import/no-self-import',
          message: 'Module imports itself.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      name: 'run 0 invalid 3',
      code: 'var bar = require(".")',
      filename: 'files/index.js',
      errors: [
        {
          messageId: 'import/no-self-import',
          message: 'Module imports itself.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      name: 'run 0 invalid 4',
      code: 'var bar = require("./")',
      filename: 'files/index.js',
      errors: [
        {
          messageId: 'import/no-self-import',
          message: 'Module imports itself.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      name: 'run 0 invalid 5',
      code: 'var bar = require("././././")',
      filename: 'files/index.js',
      errors: [
        {
          messageId: 'import/no-self-import',
          message: 'Module imports itself.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      name: 'run 0 invalid 6',
      code: 'var bar = require("../no-self-import-folder")',
      filename: 'files/no-self-import-folder/index.js',
      errors: [
        {
          messageId: 'import/no-self-import',
          message: 'Module imports itself.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 46,
        },
      ],
    },
  ],
});

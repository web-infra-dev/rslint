// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
new RuleTester(fixtureConfig()).run('no-webpack-loader-syntax', {
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
      code: 'import foo from "./foo.css"',
    },
    {
      name: 'run 0 valid 3',
      code: 'import data from "@scope/my-package/data.json"',
    },
    {
      name: 'run 0 valid 4',
      code: 'var _ = require("lodash")',
    },
    {
      name: 'run 0 valid 5',
      code: 'var find = require("lodash.find")',
    },
    {
      name: 'run 0 valid 6',
      code: 'var foo = require("./foo")',
    },
    {
      name: 'run 0 valid 7',
      code: 'var foo = require("../foo")',
    },
    {
      name: 'run 0 valid 8',
      code: 'var foo = require("foo")',
    },
    {
      name: 'run 0 valid 9',
      code: 'var foo = require("./")',
    },
    {
      name: 'run 0 valid 10',
      code: 'var foo = require("@scope/foo")',
    },
    {
      name: 'run 0 valid 11',
      code: 'var foo = fn("babel!lodash")',
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: 'import _ from "babel!lodash"',
      errors: [
        {
          messageId: 'import/no-webpack-loader-syntax',
          message:
            "Unexpected '!' in 'babel!lodash'. Do not use import syntax to configure webpack loaders.",
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      name: 'run 0 invalid 1',
      code: 'import find from "-babel-loader!lodash.find"',
      errors: [
        {
          messageId: 'import/no-webpack-loader-syntax',
          message:
            "Unexpected '!' in '-babel-loader!lodash.find'. Do not use import syntax to configure webpack loaders.",
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 45,
        },
      ],
    },
    {
      name: 'run 0 invalid 2',
      code: 'import foo from "style!css!./foo.css"',
      errors: [
        {
          messageId: 'import/no-webpack-loader-syntax',
          message:
            "Unexpected '!' in 'style!css!./foo.css'. Do not use import syntax to configure webpack loaders.",
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      name: 'run 0 invalid 3',
      code: 'import data from "json!@scope/my-package/data.json"',
      errors: [
        {
          messageId: 'import/no-webpack-loader-syntax',
          message:
            "Unexpected '!' in 'json!@scope/my-package/data.json'. Do not use import syntax to configure webpack loaders.",
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 52,
        },
      ],
    },
    {
      name: 'run 0 invalid 4',
      code: 'var _ = require("babel!lodash")',
      errors: [
        {
          messageId: 'import/no-webpack-loader-syntax',
          message:
            "Unexpected '!' in 'babel!lodash'. Do not use import syntax to configure webpack loaders.",
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      name: 'run 0 invalid 5',
      code: 'var find = require("-babel-loader!lodash.find")',
      errors: [
        {
          messageId: 'import/no-webpack-loader-syntax',
          message:
            "Unexpected '!' in '-babel-loader!lodash.find'. Do not use import syntax to configure webpack loaders.",
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 47,
        },
      ],
    },
    {
      name: 'run 0 invalid 6',
      code: 'var foo = require("style!css!./foo.css")',
      errors: [
        {
          messageId: 'import/no-webpack-loader-syntax',
          message:
            "Unexpected '!' in 'style!css!./foo.css'. Do not use import syntax to configure webpack loaders.",
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      name: 'run 0 invalid 7',
      code: 'var data = require("json!@scope/my-package/data.json")',
      errors: [
        {
          messageId: 'import/no-webpack-loader-syntax',
          message:
            "Unexpected '!' in 'json!@scope/my-package/data.json'. Do not use import syntax to configure webpack loaders.",
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 54,
        },
      ],
    },
  ],
});

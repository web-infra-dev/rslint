// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
new RuleTester(fixtureConfig()).run('no-unassigned-import', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: 'import _ from "lodash"',
    },
    {
      name: 'run 0 valid 1',
      code: 'import _, {foo} from "lodash"',
    },
    {
      name: 'run 0 valid 2',
      code: 'import _, {foo as bar} from "lodash"',
    },
    {
      name: 'run 0 valid 3',
      code: 'import {foo as bar} from "lodash"',
    },
    {
      name: 'run 0 valid 4',
      code: 'import * as _ from "lodash"',
    },
    {
      name: 'run 0 valid 5',
      code: 'import _ from "./"',
    },
    {
      name: 'run 0 valid 6',
      code: 'const _ = require("lodash")',
    },
    {
      name: 'run 0 valid 7',
      code: 'const {foo} = require("lodash")',
    },
    {
      name: 'run 0 valid 8',
      code: 'const {foo: bar} = require("lodash")',
    },
    {
      name: 'run 0 valid 9',
      code: 'const [a, b] = require("lodash")',
    },
    {
      name: 'run 0 valid 10',
      code: 'const _ = require("./")',
    },
    {
      name: 'run 0 valid 11',
      code: 'foo(require("lodash"))',
    },
    {
      name: 'run 0 valid 12',
      code: 'require("lodash").foo',
    },
    {
      name: 'run 0 valid 13',
      code: 'require("lodash").foo()',
    },
    {
      name: 'run 0 valid 14',
      code: 'require("lodash")()',
    },
    {
      name: 'run 0 valid 15',
      code: 'import "app.css"',
      options: [
        {
          allow: ['**/*.css'],
        },
      ],
    },
    {
      name: 'run 0 valid 16',
      code: 'import "app.css";',
      options: [
        {
          allow: ['*.css'],
        },
      ],
    },
    {
      name: 'run 0 valid 17',
      code: 'import "./app.css"',
      options: [
        {
          allow: ['**/*.css'],
        },
      ],
    },
    {
      name: 'run 0 valid 18',
      code: 'import "foo/bar"',
      options: [
        {
          allow: ['foo/**'],
        },
      ],
    },
    {
      name: 'run 0 valid 19',
      code: 'import "foo/bar"',
      options: [
        {
          allow: ['foo/bar'],
        },
      ],
    },
    {
      name: 'run 0 valid 20',
      code: 'import "../dir/app.css"',
      options: [
        {
          allow: ['**/*.css'],
        },
      ],
    },
    {
      name: 'run 0 valid 21',
      code: 'import "../dir/app.js"',
      options: [
        {
          allow: ['**/dir/**'],
        },
      ],
    },
    {
      name: 'run 0 valid 22',
      code: 'require("./app.css")',
      options: [
        {
          allow: ['**/*.css'],
        },
      ],
    },
    {
      name: 'run 0 valid 23',
      code: 'import "babel-register"',
      options: [
        {
          allow: ['babel-register'],
        },
      ],
    },
    {
      name: 'run 0 valid 24',
      code: 'import "./styles/app.css"',
      filename: 'src/app.js',
      options: [
        {
          allow: ['src/styles/**'],
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
      name: 'run 0 valid 25',
      code: 'import "../scripts/register.js"',
      filename: 'src/app.js',
      options: [
        {
          allow: ['src/styles/**', '**/scripts/*.js'],
        },
      ],
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
      name: 'run 0 invalid 0',
      code: 'import "lodash"',
      errors: [
        {
          messageId: '',
          message: 'Imported module should be assigned',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      name: 'run 0 invalid 1',
      code: 'require("lodash")',
      errors: [
        {
          messageId: '',
          message: 'Imported module should be assigned',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      name: 'run 0 invalid 2',
      code: 'import "./app.css"',
      options: [
        {
          allow: ['**/*.js'],
        },
      ],
      errors: [
        {
          messageId: '',
          message: 'Imported module should be assigned',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      name: 'run 0 invalid 3',
      code: 'import "./app.css"',
      options: [
        {
          allow: ['**/dir/**'],
        },
      ],
      errors: [
        {
          messageId: '',
          message: 'Imported module should be assigned',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      name: 'run 0 invalid 4',
      code: 'require("./app.css")',
      options: [
        {
          allow: ['**/*.js'],
        },
      ],
      errors: [
        {
          messageId: '',
          message: 'Imported module should be assigned',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      name: 'run 0 invalid 5',
      code: 'import "./styles/app.css"',
      filename: 'src/app.js',
      options: [
        {
          allow: ['styles/*.css'],
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
          message: 'Imported module should be assigned',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
  ],
});

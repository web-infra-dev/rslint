// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
new RuleTester(fixtureConfig()).run('no-dynamic-require', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: 'import _ from "lodash"',
    },
    {
      name: 'run 0 valid 1',
      code: 'require("foo")',
    },
    {
      name: 'run 0 valid 2',
      code: 'require(`foo`)',
    },
    {
      name: 'run 0 valid 3',
      code: 'require("./foo")',
    },
    {
      name: 'run 0 valid 4',
      code: 'require("@scope/foo")',
    },
    {
      name: 'run 0 valid 5',
      code: 'require()',
    },
    {
      name: 'run 0 valid 6',
      code: 'require("./foo", "bar" + "okay")',
    },
    {
      name: 'run 0 valid 7',
      code: 'var foo = require("foo")',
    },
    {
      name: 'run 0 valid 8',
      code: 'var foo = require(`foo`)',
    },
    {
      name: 'run 0 valid 9',
      code: 'var foo = require("./foo")',
    },
    {
      name: 'run 0 valid 10',
      code: 'var foo = require("@scope/foo")',
    },
    {
      name: 'run 0 valid 11',
      code: 'import("foo")',
      options: [
        {
          esmodule: true,
        },
      ],
    },
    {
      name: 'run 0 valid 12',
      code: 'import(`foo`)',
      options: [
        {
          esmodule: true,
        },
      ],
    },
    {
      name: 'run 0 valid 13',
      code: 'import("./foo")',
      options: [
        {
          esmodule: true,
        },
      ],
    },
    {
      name: 'run 0 valid 14',
      code: 'import("@scope/foo")',
      options: [
        {
          esmodule: true,
        },
      ],
    },
    {
      name: 'run 0 valid 15',
      code: 'var foo = import("foo")',
      options: [
        {
          esmodule: true,
        },
      ],
    },
    {
      name: 'run 0 valid 16',
      code: 'var foo = import(`foo`)',
      options: [
        {
          esmodule: true,
        },
      ],
    },
    {
      name: 'run 0 valid 17',
      code: 'var foo = import("./foo")',
      options: [
        {
          esmodule: true,
        },
      ],
    },
    {
      name: 'run 0 valid 18',
      code: 'var foo = import("@scope/foo")',
      options: [
        {
          esmodule: true,
        },
      ],
    },
    {
      name: 'run 0 valid 19',
      code: 'import("../" + name)',
    },
    {
      name: 'run 0 valid 20',
      code: 'import(`../${name}`)',
    },
    {
      name: 'run 0 valid 21',
      code: "require('../name');\nrequire(`../name`);",
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: 'require("../" + name)',
      errors: [
        {
          messageId: '',
          message: 'Calls to require() should use string literals',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      name: 'run 0 invalid 1',
      code: 'require(`../${name}`)',
      errors: [
        {
          messageId: '',
          message: 'Calls to require() should use string literals',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      name: 'run 0 invalid 2',
      code: 'require(name)',
      errors: [
        {
          messageId: '',
          message: 'Calls to require() should use string literals',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      name: 'run 0 invalid 3',
      code: 'require(name())',
      errors: [
        {
          messageId: '',
          message: 'Calls to require() should use string literals',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      name: 'run 0 invalid 4',
      code: 'require(name + "foo", "bar")',
      options: [
        {
          esmodule: true,
        },
      ],
      errors: [
        {
          messageId: '',
          message: 'Calls to require() should use string literals',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      name: 'run 0 invalid 5',
      code: 'import("../" + name)',
      options: [
        {
          esmodule: true,
        },
      ],
      errors: [
        {
          messageId: '',
          message: 'Calls to import() should use string literals',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      name: 'run 0 invalid 6',
      code: 'import(`../${name}`)',
      options: [
        {
          esmodule: true,
        },
      ],
      errors: [
        {
          messageId: '',
          message: 'Calls to import() should use string literals',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      name: 'run 0 invalid 7',
      code: 'import(name)',
      options: [
        {
          esmodule: true,
        },
      ],
      errors: [
        {
          messageId: '',
          message: 'Calls to import() should use string literals',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      name: 'run 0 invalid 8',
      code: 'import(name())',
      options: [
        {
          esmodule: true,
        },
      ],
      errors: [
        {
          messageId: '',
          message: 'Calls to import() should use string literals',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      name: 'run 0 invalid 9',
      code: 'require(`foo${x}`)',
      errors: [
        {
          messageId: '',
          message: 'Calls to require() should use string literals',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      name: 'run 0 invalid 10',
      code: 'var foo = require(`foo${x}`)',
      errors: [
        {
          messageId: '',
          message: 'Calls to require() should use string literals',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      name: 'run 0 invalid 11',
      code: "require(name);\nrequire('../' + name);\nrequire(`../${name}`);\nrequire(name());",
      errors: [
        {
          messageId: '',
          message: 'Calls to require() should use string literals',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
        {
          messageId: '',
          message: 'Calls to require() should use string literals',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 22,
        },
        {
          messageId: '',
          message: 'Calls to require() should use string literals',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 22,
        },
        {
          messageId: '',
          message: 'Calls to require() should use string literals',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 16,
        },
      ],
    },
  ],
});

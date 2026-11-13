// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
new RuleTester(fixtureConfig()).run('no-amd', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: 'import "x";',
    },
    {
      name: 'run 0 valid 1',
      code: 'import x from "x"',
    },
    {
      name: 'run 0 valid 2',
      code: 'var x = require("x")',
    },
    {
      name: 'run 0 valid 3',
      code: 'require("x")',
    },
    {
      name: 'run 0 valid 4',
      code: 'require("x", "y")',
    },
    {
      name: 'run 0 valid 5',
      code: 'setTimeout(foo, 100)',
    },
    {
      name: 'run 0 valid 6',
      code: '(a || b)(1, 2, 3)',
    },
    {
      name: 'run 0 valid 7',
      code: 'function x() { define(["a"], function (a) {}) }',
    },
    {
      name: 'run 0 valid 8',
      code: 'function x() { require(["a"], function (a) {}) }',
    },
    {
      name: 'run 0 valid 9',
      code: 'define(0, 1, 2)',
    },
    {
      name: 'run 0 valid 10',
      code: 'define("a")',
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: 'define([], function() {})',
      errors: [
        {
          messageId: '',
          message: 'Expected imports instead of AMD define().',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      name: 'run 0 invalid 1',
      code: 'define(["a"], function(a) { console.log(a); })',
      errors: [
        {
          messageId: '',
          message: 'Expected imports instead of AMD define().',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 47,
        },
      ],
    },
    {
      name: 'run 0 invalid 2',
      code: 'require([], function() {})',
      errors: [
        {
          messageId: '',
          message: 'Expected imports instead of AMD require().',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      name: 'run 0 invalid 3',
      code: 'require(["a"], function(a) { console.log(a); })',
      errors: [
        {
          messageId: '',
          message: 'Expected imports instead of AMD require().',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 48,
        },
      ],
    },
    {
      name: 'run 0 invalid 4',
      code: 'define(["a", "b"], function (a, b) { /* ... */ })',
      errors: [
        {
          messageId: '',
          message: 'Expected imports instead of AMD define().',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
    {
      name: 'run 0 invalid 5',
      code: 'require(["b", "c"], function (b, c) { /* ... */ })',
      errors: [
        {
          messageId: '',
          message: 'Expected imports instead of AMD require().',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 51,
        },
      ],
    },
  ],
});

import { RuleTester } from '../rule-tester.js';

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/no-amd.js
// The suite config uses sourceType: module, matching upstream.
const ruleTester = new RuleTester();

ruleTester.run('no-amd', null as never, {
  valid: [
    { code: 'import "x";' },
    { code: 'import x from "x"' },
    { code: 'var x = require("x")' },
    { code: 'require("x")' },
    // Two arguments, but not an array.
    { code: 'require("x", "y")' },
    // Other functions and non-identifier callees.
    { code: 'setTimeout(foo, 100)' },
    { code: '(a || b)(1, 2, 3)' },
    // Nested scopes.
    { code: 'function x() { define(["a"], function (a) {}) }' },
    { code: 'function x() { require(["a"], function (a) {}) }' },
    // Unmatched argument types/counts.
    { code: 'define(0, 1, 2)' },
    { code: 'define("a")' },
  ],
  // Upstream's ESLint >= 4 branch.
  invalid: [
    {
      code: 'define([], function() {})',
      errors: [{ message: 'Expected imports instead of AMD define().' }],
    },
    {
      code: 'define(["a"], function(a) { console.log(a); })',
      errors: [{ message: 'Expected imports instead of AMD define().' }],
    },
    {
      code: 'require([], function() {})',
      errors: [{ message: 'Expected imports instead of AMD require().' }],
    },
    {
      code: 'require(["a"], function(a) { console.log(a); })',
      errors: [{ message: 'Expected imports instead of AMD require().' }],
    },
    // https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-amd.md
    {
      code: 'define(["a", "b"], function (a, b) { /* ... */ })',
      errors: [{ message: 'Expected imports instead of AMD define().' }],
    },
    {
      code: 'require(["b", "c"], function (b, c) { /* ... */ })',
      errors: [{ message: 'Expected imports instead of AMD require().' }],
    },
  ],
});

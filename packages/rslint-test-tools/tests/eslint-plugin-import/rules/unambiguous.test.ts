import { RuleTester } from '../rule-tester.js';

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/unambiguous.js
// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/unambiguous.md
const ruleTester = new RuleTester();
const message = 'This module could be parsed as a valid script.';

ruleTester.run('unambiguous', null as never, {
  valid: [
    // Upstream's string cases use RuleTester's script default.
    {
      code: 'function x() {}',
      languageOptions: { sourceType: 'script' },
    },
    {
      code: '"use strict"; function y() {}',
      languageOptions: { sourceType: 'script' },
    },
    { code: 'import y from "z"; function x() {}' },
    { code: 'import * as y from "z"; function x() {}' },
    { code: 'import { y } from "z"; function x() {}' },
    { code: 'import z, { y } from "z"; function x() {}' },
    { code: 'function x() {}; export {}' },
    { code: 'function x() {}; export { x }' },
    { code: 'function x() {}; export { y } from "z"' },
    // Upstream uses Babel for this syntax; tsgo supports it natively.
    { code: 'function x() {}; export * as y from "z"' },
    { code: 'export function x() {}' },
    // Upstream documentation.
    { code: "import 'foo'\nfunction x() { return 42 }" },
    { code: 'export function x() { return 42 }' },
    {
      code: "(function x() { return 42 })()\nexport {} // simple way to mark side-effects-only file as 'module' without any imports/exports",
    },
  ],
  invalid: [
    {
      code: 'function x() {}',
      languageOptions: { sourceType: 'module' },
      errors: [message],
      output: null,
    },
    // Upstream documentation.
    {
      code: '(function x() { return 42 })()',
      errors: [message],
      output: null,
    },
  ],
});

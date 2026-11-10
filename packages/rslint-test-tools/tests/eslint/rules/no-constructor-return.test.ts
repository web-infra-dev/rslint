// Upstream: https://github.com/eslint/eslint/blob/v10.8.1/tests/lib/rules/no-constructor-return.js
/**
 * @fileoverview Tests for no-constructor-return rule.
 * @author Pig Fang <https://github.com/g-plane>
 */

//------------------------------------------------------------------------------
// Requirements
//------------------------------------------------------------------------------

import { RuleTester } from '../rule-tester';

//------------------------------------------------------------------------------
// Tests
//------------------------------------------------------------------------------

const ruleTester = new RuleTester({
  parserOptions: { ecmaVersion: 2015, sourceType: 'script' },
});

const errors = [
  {
    messageId: 'unexpected',
  },
];

ruleTester.run('no-constructor-return', {
  valid: [
    'function fn() { return }',
    'function fn(kumiko) { if (kumiko) { return kumiko } }',
    'const fn = function () { return }',
    'const fn = function () { if (kumiko) { return kumiko } }',
    'const fn = () => { return }',
    'const fn = () => { if (kumiko) { return kumiko } }',
    {
      code: "return 'Kumiko Oumae'",
      languageOptions: {
        parserOptions: { ecmaFeatures: { globalReturn: true } },
      },
    },

    'class C {  }',
    'class C { constructor() {} }',
    'class C { constructor() { let v } }',
    "class C { method() { return '' } }",
    "class C { get value() { return '' } }",
    'class C { constructor(a) { if (!a) { return } else { a() } } }',
    'class C { constructor() { function fn() { return true } } }',
    'class C { constructor() { this.fn = function () { return true } } }',
    'class C { constructor() { this.fn = () => { return true } } }',
    'class C { constructor() { return } }',
    'class C { constructor() { { return } } }',
  ],
  invalid: [
    {
      code: "class C { constructor() { return '' } }",
      errors,
    },
    {
      code: "class C { constructor(a) { if (!a) { return '' } else { a() } } }",
      errors,
    },
  ],
});

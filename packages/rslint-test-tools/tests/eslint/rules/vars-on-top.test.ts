import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('vars-on-top', {
  valid: [
    'var first = 0; function foo() { first = 2; }',
    'function foo() { var first; first = 1; }',
    "'use strict'; var x; f();",
    "function f() { 'use strict'; var x; f(); }",
    {
      code: "import React from 'react'; var y;",
      languageOptions: { ecmaVersion: 6, sourceType: 'module' },
    },
    {
      code: 'class C { static { var x; foo(); } }',
      languageOptions: { ecmaVersion: 2022 },
    },
  ],
  invalid: [
    {
      code: 'function foo() { var first; first = 1; var second = 1; }',
      errors: [{ messageId: 'top' }],
    },
    {
      code: 'function foo() { if (true) { var second = true; } }',
      errors: [{ messageId: 'top' }],
    },
    {
      code: 'function foo() { for (var i = 0; i < 10; i++) {} }',
      errors: [{ messageId: 'top' }],
    },
    {
      code: "'use strict'; 0; var x; f();",
      errors: [{ messageId: 'top' }],
    },
  ],
});

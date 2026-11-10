// Every case from ESLint v10.8.1, with messages and full ranges
// additionally pinned from the same upstream version. No upstream expectations are relaxed.
// https://github.com/eslint/eslint/blob/v10.8.1/tests/lib/rules/no-class-assign.js
import { RuleTester } from '@rslint/rule-tester';

const ruleTester = new RuleTester({
  plugin: '',
  languageOptions: { ecmaVersion: 6 },
});

ruleTester.run('no-class-assign', {
  valid: [
    {
      code: 'class A { } foo(A);',
    },
    {
      code: 'let A = class A { }; foo(A);',
    },
    {
      code: 'class A { b(A) { A = 0; } }',
    },
    {
      code: 'class A { b() { let A; A = 0; } }',
    },
    {
      code: 'let A = class { b() { A = 0; } }',
    },
    {
      code: 'var x = 0; x = 1;',
    },
    {
      code: 'let x = 0; x = 1;',
    },
    {
      code: 'const x = 0; x = 1;',
    },
    {
      code: 'function x() {} x = 1;',
    },
    {
      code: 'function foo(x) { x = 1; }',
    },
    {
      code: 'try {} catch (x) { x = 1; }',
    },
  ],
  invalid: [
    {
      code: 'class A { } A = 0;',
      errors: [
        {
          messageId: 'class',
          message: "'A' is a class.",
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'class A { } ({A} = 0);',
      errors: [
        {
          messageId: 'class',
          message: "'A' is a class.",
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: 'class A { } ({b: A = 0} = {});',
      errors: [
        {
          messageId: 'class',
          message: "'A' is a class.",
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'A = 0; class A { }',
      errors: [
        {
          messageId: 'class',
          message: "'A' is a class.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 2,
        },
      ],
    },
    {
      code: 'class A { b() { A = 0; } }',
      errors: [
        {
          messageId: 'class',
          message: "'A' is a class.",
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'let A = class A { b() { A = 0; } }',
      errors: [
        {
          messageId: 'class',
          message: "'A' is a class.",
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: 'class A { } A = 0; A = 1;',
      errors: [
        {
          messageId: 'class',
          message: "'A' is a class.",
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 14,
        },
        {
          messageId: 'class',
          message: "'A' is a class.",
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
  ],
});

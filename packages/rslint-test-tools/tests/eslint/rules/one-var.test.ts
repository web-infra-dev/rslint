import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('one-var', {
  valid: [
    'function foo() { var bar = true; }',
    {
      code: 'function foo() { var bar = true, baz = false; }',
      options: ['always'] as any,
    },
    {
      code: 'function foo() { var bar = true; var baz = false; }',
      options: ['never'] as any,
    },
    {
      code: 'var bar = true; var baz = false;',
      options: [{ initialized: 'never' }] as any,
    },
    { code: 'var bar, baz;', options: [{ initialized: 'never' }] as any },
    { code: 'var bar; var baz;', options: [{ uninitialized: 'never' }] as any },
    {
      code: 'function foo() { let a = 1; let b = 2; }',
      options: ['never'] as any,
    },
    {
      code: 'function foo() { const a = 1; const b = 2; }',
      options: ['never'] as any,
    },
    {
      code: 'function foo() { let a = 1; let b = 2; const c = false; const d = true; var e = true, f = false; }',
      options: [{ var: 'always', let: 'never', const: 'never' }] as any,
    },
    { code: 'var a = 0, b, c;', options: ['consecutive'] as any },
    {
      code: 'var a = 0, b = 1; foo(); var c = 2;',
      options: ['consecutive'] as any,
    },
    {
      code: "var foo = require('foo'), bar = require('bar');",
      options: [{ separateRequires: true, var: 'always' }] as any,
    },
    { code: 'using a = 0, b = 1;' },
    { code: 'await using a = 0, b = 1;' },
    { code: 'using a = 0; using b = 1;', options: ['never'] as any },
  ],
  invalid: [
    {
      code: 'var bar = true, baz = false;',
      options: ['never'] as any,
      errors: [{ messageId: 'split' }],
    },
    {
      code: 'function foo() { var bar = true; var baz = false; }',
      options: ['always'] as any,
      errors: [{ messageId: 'combine' }],
    },
    {
      code: 'function foo() { let a = 1; let b = 2; }',
      options: ['always'] as any,
      errors: [{ messageId: 'combine' }],
    },
    {
      code: 'function foo() { const a = 1; const b = 2; }',
      options: ['always'] as any,
      errors: [{ messageId: 'combine' }],
    },
    {
      code: 'function foo() { var foo = true, bar = false; }',
      options: [{ initialized: 'never' }] as any,
      errors: [{ messageId: 'splitInitialized' }],
    },
    {
      code: 'function foo() { var foo, bar; }',
      options: [{ uninitialized: 'never' }] as any,
      errors: [{ messageId: 'splitUninitialized' }],
    },
    {
      code: 'var a = 1, b; var c;',
      options: ['consecutive'] as any,
      errors: [{ messageId: 'combine' }],
    },
    {
      code: 'using a = 0; using b = 1;',
      options: ['always'] as any,
      errors: [{ messageId: 'combine' }],
    },
    {
      code: 'await using a = 0; await using b = 1;',
      options: ['always'] as any,
      errors: [{ messageId: 'combine' }],
    },
    {
      code: 'export const foo=1, bar=2;',
      options: ['never'] as any,
      errors: [{ messageId: 'split' }],
    },
  ],
});

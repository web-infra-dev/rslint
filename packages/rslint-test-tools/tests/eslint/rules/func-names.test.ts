import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('func-names', {
  valid: [
    'Foo.prototype.bar = function bar(){};',
    'function foo(){}',
    'new function bar(){}',
    { code: 'function foo() {}', options: ['always'] as any },
    { code: 'var a = function foo() {};', options: ['always'] as any },
    { code: 'var foo = function(){};', options: ['as-needed'] as any },
    { code: '({foo: function(){}});', options: ['as-needed'] as any },
    { code: 'function foo() {}', options: ['never'] as any },
    { code: 'var a = function() {};', options: ['never'] as any },
    { code: 'var a = function foo() { foo(); };', options: ['never'] as any },
    { code: 'export default function foo() {}', options: ['always'] as any },
    { code: 'export default function() {}', options: ['never'] as any },
    {
      code: 'var foo = bar(function *baz() {});',
      options: ['always', { generators: 'as-needed' }] as any,
    },
    { code: 'class C { foo = function() {}; }', options: ['as-needed'] as any },
  ],
  invalid: [
    {
      code: 'Foo.prototype.bar = function() {};',
      errors: [
        {
          messageId: 'unnamed',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: 'f(function(){})',
      errors: [
        {
          messageId: 'unnamed',
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: 'var {foo} = function(){};',
      options: ['as-needed'] as any,
      errors: [
        {
          messageId: 'unnamed',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: 'var x = function foo() {};',
      options: ['never'] as any,
      errors: [
        {
          messageId: 'named',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '({foo: function foo() {}})',
      options: ['never'] as any,
      errors: [
        {
          messageId: 'named',
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: 'export default function() {}',
      options: ['always'] as any,
      errors: [
        {
          messageId: 'unnamed',
          column: 16,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'var foo = function*() {};',
      options: ['always'] as any,
      errors: [
        {
          messageId: 'unnamed',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: 'class C { foo = function() {} }',
      options: ['always'] as any,
      errors: [
        {
          messageId: 'unnamed',
          column: 11,
          endColumn: 25,
        },
      ],
    },
  ],
});

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('func-style', {
  valid: [
    {
      code: 'function foo(){}\n function bar(){}',
      options: ['declaration'] as any,
    },
    { code: 'foo.bar = function(){};', options: ['declaration'] as any },
    { code: '(function() { /* code */ }());', options: ['declaration'] as any },
    {
      code: 'var object = { foo: function(){} };',
      options: ['declaration'] as any,
    },
    { code: 'foo.bar = function(){};', options: ['expression'] as any },
    {
      code: 'var foo = function(){};\n var bar = function(){};',
      options: ['expression'] as any,
    },
    {
      code: 'var foo = () => {};\n var bar = () => {}',
      options: ['expression'] as any,
    },
    { code: 'var foo = () => { this; };', options: ['declaration'] as any },
    { code: 'export default function () {};' },
    {
      code: 'var foo = () => {};',
      options: ['declaration', { allowArrowFunctions: true }] as any,
    },
    { code: 'export function foo() {};', options: ['declaration'] as any },
    {
      code: 'export function foo() {};',
      options: [
        'expression',
        { overrides: { namedExports: 'declaration' } },
      ] as any,
    },
    {
      code: 'export var foo = function(){};',
      options: ['expression'] as any,
    },
    {
      code: 'export var foo = function(){};',
      options: ['expression', { overrides: { namedExports: 'ignore' } }] as any,
    },
    { code: '$1: function $2() { }', options: ['declaration'] as any },
    {
      code: 'switch ($0) { case $1: function $2() { } }',
      options: ['declaration'] as any,
    },
  ],
  invalid: [
    {
      code: 'var foo = function(){};',
      options: ['declaration'] as any,
      errors: [
        {
          messageId: 'declaration',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'var foo = () => {};',
      options: ['declaration'] as any,
      errors: [
        {
          messageId: 'declaration',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'function foo(){}',
      options: ['expression'] as any,
      errors: [
        {
          messageId: 'expression',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: 'export function foo(){}',
      options: ['expression'] as any,
      errors: [
        {
          messageId: 'expression',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'export var foo = function(){};',
      options: ['declaration'] as any,
      errors: [
        {
          messageId: 'declaration',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: 'const foo = function() {};',
      options: ['declaration', { allowTypeAnnotation: true }] as any,
      errors: [
        {
          messageId: 'declaration',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '$1: function $2() { }',
      errors: [
        {
          messageId: 'expression',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'if (foo) function bar() {}',
      errors: [
        {
          messageId: 'expression',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
  ],
});

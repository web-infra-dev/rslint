import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('generator-star-spacing', null as never, {
  valid: [
    // default ('before')
    { code: 'function foo(){}' },
    { code: 'function *foo(){}' },
    { code: 'var foo = function *foo(){};' },
    { code: 'var foo = function *(){};' },
    { code: 'var foo = { *foo(){} };' },
    { code: 'class Foo { *foo(){} }' },
    { code: 'class Foo { static *foo(){} }' },

    // 'before'
    { code: 'function *foo(){}', options: ['before'] },
    { code: 'var foo = { *foo(){} };', options: ['before'] },

    // 'after'
    { code: 'function* foo(){}', options: ['after'] },
    { code: 'var foo = { * foo(){} };', options: ['after'] },
    { code: 'class Foo { static* foo(){} }', options: ['after'] },

    // 'both'
    { code: 'function * foo(){}', options: ['both'] },

    // 'neither'
    { code: 'function*foo(){}', options: ['neither'] },

    // object form
    { code: 'function *foo(){}', options: [{ before: true, after: false }] },
    { code: 'function* foo(){}', options: [{ before: false, after: true }] },

    // full configurability
    {
      code: 'class Foo { * foo(){} }',
      options: [{ before: false, after: false, method: 'both' }],
    },
    {
      code: 'var foo = { * foo(){} }',
      options: [{ before: false, after: false, shorthand: 'both' }],
    },
    {
      code: 'function * foo(){}',
      options: [
        { before: false, after: false, named: { before: true, after: true } },
      ],
    },

    // non-generator: rule must not flag
    { code: 'async function foo() { }' },
    { code: 'class A { async foo() { } }' },
  ],

  invalid: [
    // default ('before')
    {
      code: 'function*foo(){}',
      output: 'function *foo(){}',
      errors: [{ messageId: 'missingBefore', line: 1, column: 9 }],
    },
    {
      code: 'function* foo(arg1, arg2){}',
      output: 'function *foo(arg1, arg2){}',
      errors: [
        { messageId: 'missingBefore', line: 1, column: 9 },
        { messageId: 'unexpectedAfter', line: 1, column: 9 },
      ],
    },
    {
      code: 'class Foo { static* foo(){} }',
      output: 'class Foo { static *foo(){} }',
      errors: [
        { messageId: 'missingBefore', line: 1, column: 19 },
        { messageId: 'unexpectedAfter', line: 1, column: 19 },
      ],
    },

    // 'after'
    {
      code: 'function *foo(){}',
      output: 'function* foo(){}',
      options: ['after'],
      errors: [
        { messageId: 'unexpectedBefore', line: 1, column: 10 },
        { messageId: 'missingAfter', line: 1, column: 10 },
      ],
    },

    // 'both'
    {
      code: 'function*foo(){}',
      output: 'function * foo(){}',
      options: ['both'],
      errors: [
        { messageId: 'missingBefore', line: 1, column: 9 },
        { messageId: 'missingAfter', line: 1, column: 9 },
      ],
    },

    // 'neither'
    {
      code: 'function * foo(){}',
      output: 'function*foo(){}',
      options: ['neither'],
      errors: [
        { messageId: 'unexpectedBefore', line: 1, column: 10 },
        { messageId: 'unexpectedAfter', line: 1, column: 10 },
      ],
    },

    // method override
    {
      code: 'class Foo { *foo(){} }',
      output: 'class Foo { * foo(){} }',
      options: [{ before: false, after: false, method: 'both' }],
      errors: [{ messageId: 'missingAfter', line: 1, column: 13 }],
    },

    // anonymous override
    {
      code: 'var foo = function*(){};',
      output: 'var foo = function * (){};',
      options: [{ before: false, after: false, anonymous: 'both' }],
      errors: [
        { messageId: 'missingBefore', line: 1, column: 19 },
        { messageId: 'missingAfter', line: 1, column: 19 },
      ],
    },

    // async generator
    {
      code: '({ async*foo(){} })',
      output: '({ async * foo(){} })',
      options: [{ before: true, after: true }],
      errors: [
        { messageId: 'missingBefore', line: 1, column: 9 },
        { messageId: 'missingAfter', line: 1, column: 9 },
      ],
    },
    {
      code: 'class Foo { static async*foo(){} }',
      output: 'class Foo { static async * foo(){} }',
      options: [{ before: true, after: true }],
      errors: [
        { messageId: 'missingBefore', line: 1, column: 25 },
        { messageId: 'missingAfter', line: 1, column: 25 },
      ],
    },
  ],
});

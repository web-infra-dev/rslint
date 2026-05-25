import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-extend-native', {
  valid: [
    'x.prototype.p = 0',
    "x.prototype['p'] = 0",
    'Object.p = 0',
    'Object.toString.bind = 0',
    "Object['toString'].bind = 0",
    "Object.defineProperty(x, 'p', {value: 0})",
    'Object.defineProperties(x, {p: {value: 0}})',
    'global.Object.prototype.toString = 0',
    'this.Object.prototype.toString = 0',
    'with(Object) { prototype.p = 0; }',
    'o = Object; o.prototype.toString = 0',
    "eval('Object.prototype.toString = 0')",
    'parseFloat.prototype.x = 1',
    {
      code: 'Object.prototype.g = 0',
      options: { exceptions: ['Object'] },
    },
    'obj[Object.prototype] = 0',

    // https://github.com/eslint/eslint/issues/4438
    'Object.defineProperty()',
    'Object.defineProperties()',

    // https://github.com/eslint/eslint/issues/8461
    'function foo() { var Object = function() {}; Object.prototype.p = 0 }',
    '{ let Object = function() {}; Object.prototype.p = 0 }',
  ],
  invalid: [
    {
      code: 'Object.prototype.p = 0',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'BigInt.prototype.p = 0',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'WeakRef.prototype.p = 0',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'FinalizationRegistry.prototype.p = 0',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'AggregateError.prototype.p = 0',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "Function.prototype['p'] = 0",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "String['prototype'].p = 0",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "Number['prototype']['p'] = 0",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "Object.defineProperty(Array.prototype, 'p', {value: 0})",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'Object.defineProperties(Array.prototype, {p: {value: 0}})',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'Object.defineProperties(Array.prototype, {p: {value: 0}, q: {value: 0}})',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "Number['prototype']['p'] = 0",
      options: { exceptions: ['Object'] },
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'Object.prototype.p = 0; Object.prototype.q = 0',
      errors: [{ messageId: 'unexpected' }, { messageId: 'unexpected' }],
    },
    {
      code: 'function foo() { Object.prototype.p = 0 }',
      errors: [{ messageId: 'unexpected' }],
    },

    // Optional chaining
    {
      code: '(Object?.prototype).p = 0',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "Object.defineProperty(Object?.prototype, 'p', { value: 0 })",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "Object?.defineProperty(Object.prototype, 'p', { value: 0 })",
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "(Object?.defineProperty)(Object.prototype, 'p', { value: 0 })",
      errors: [{ messageId: 'unexpected' }],
    },

    // Logical assignments
    {
      code: 'Array.prototype.p &&= 0',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'Array.prototype.p ||= 0',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'Array.prototype.p ??= 0',
      errors: [{ messageId: 'unexpected' }],
    },

    // Parenthesized identifier (tsgo wraps parens as a node; ESTree doesn't).
    {
      code: '(Object).prototype.p = 0',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '((Object)).prototype.p = 0',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: "Object.defineProperty((Array).prototype, 'p', {value: 0})",
      errors: [{ messageId: 'unexpected' }],
    },

    // Chained assignment — both sides report.
    {
      code: 'Object.prototype.p = Array.prototype.q = 0',
      errors: [{ messageId: 'unexpected' }, { messageId: 'unexpected' }],
    },

    // Template literal as static property key.
    {
      code: 'Object[`prototype`].p = 0',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'Object.prototype[`p`] = 0',
      errors: [{ messageId: 'unexpected' }],
    },
  ],
});

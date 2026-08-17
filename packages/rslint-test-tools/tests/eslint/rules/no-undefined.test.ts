import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-undefined', {
  valid: [
    'void 0',
    'void!0',
    'void-0',
    'void+0',
    'null',
    'undefine',
    'ndefined',
    'a.undefined',
    'this.undefined',
    "global['undefined']",

    // https://github.com/eslint/eslint/issues/7964
    '({ undefined: bar })',
    '({ undefined: bar } = foo)',
    '({ undefined() {} })',
    'class Foo { undefined() {} }',
    '(class { undefined() {} })',
    "import { undefined as a } from 'foo'",
    "export { undefined } from 'foo'",
    "export { undefined as a } from 'foo'",
    "export { a as undefined } from 'foo'",
  ],
  invalid: [
    { code: 'undefined', errors: [{ messageId: 'unexpectedUndefined' }] },
    { code: 'undefined.a', errors: [{ messageId: 'unexpectedUndefined' }] },
    { code: 'a[undefined]', errors: [{ messageId: 'unexpectedUndefined' }] },
    { code: 'undefined[0]', errors: [{ messageId: 'unexpectedUndefined' }] },
    { code: 'f(undefined)', errors: [{ messageId: 'unexpectedUndefined' }] },
    {
      code: 'function f(undefined) {}',
      errors: [{ messageId: 'unexpectedUndefined' }],
    },
    {
      code: 'function f() { var undefined; }',
      errors: [{ messageId: 'unexpectedUndefined' }],
    },
    {
      code: 'function f() { undefined = true; }',
      errors: [{ messageId: 'unexpectedUndefined' }],
    },
    { code: 'var undefined;', errors: [{ messageId: 'unexpectedUndefined' }] },
    {
      code: 'try {} catch(undefined) {}',
      errors: [{ messageId: 'unexpectedUndefined' }],
    },
    {
      code: 'function undefined() {}',
      errors: [{ messageId: 'unexpectedUndefined' }],
    },
    {
      code: '(function undefined(){}())',
      errors: [{ messageId: 'unexpectedUndefined' }],
    },
    {
      code: 'var foo = function undefined() {}',
      errors: [{ messageId: 'unexpectedUndefined' }],
    },
    {
      code: 'foo = function undefined() {}',
      errors: [{ messageId: 'unexpectedUndefined' }],
    },
    {
      code: 'undefined = true',
      errors: [{ messageId: 'unexpectedUndefined' }],
    },
    {
      code: 'var undefined = true',
      errors: [{ messageId: 'unexpectedUndefined' }],
    },
    { code: '({ undefined })', errors: [{ messageId: 'unexpectedUndefined' }] },
    {
      code: '({ [undefined]: foo })',
      errors: [{ messageId: 'unexpectedUndefined' }],
    },
    {
      code: '({ bar: undefined })',
      errors: [{ messageId: 'unexpectedUndefined' }],
    },
    {
      code: '({ bar: undefined } = foo)',
      errors: [{ messageId: 'unexpectedUndefined' }],
    },
    {
      code: 'var { undefined } = foo',
      errors: [{ messageId: 'unexpectedUndefined' }],
    },
    {
      code: 'var { bar: undefined } = foo',
      errors: [{ messageId: 'unexpectedUndefined' }],
    },
    {
      code: '({ undefined: function undefined() {} })',
      errors: [{ messageId: 'unexpectedUndefined' }],
    },
    {
      code: '({ foo: function undefined() {} })',
      errors: [{ messageId: 'unexpectedUndefined' }],
    },
    {
      code: 'class Foo { [undefined]() {} }',
      errors: [{ messageId: 'unexpectedUndefined' }],
    },
    {
      code: '(class { [undefined]() {} })',
      errors: [{ messageId: 'unexpectedUndefined' }],
    },
    {
      code: 'var undefined = true; undefined = false;',
      errors: [
        { messageId: 'unexpectedUndefined' },
        { messageId: 'unexpectedUndefined' },
      ],
    },
    {
      code: "import undefined from 'foo'",
      errors: [{ messageId: 'unexpectedUndefined' }],
    },
    {
      code: "import * as undefined from 'foo'",
      errors: [{ messageId: 'unexpectedUndefined' }],
    },
    {
      code: "import { undefined } from 'foo'",
      errors: [{ messageId: 'unexpectedUndefined' }],
    },
    {
      code: "import { a as undefined } from 'foo'",
      errors: [{ messageId: 'unexpectedUndefined' }],
    },
    {
      code: 'let a = [b, ...undefined]',
      errors: [{ messageId: 'unexpectedUndefined' }],
    },
    {
      code: '[a, ...undefined] = b',
      errors: [{ messageId: 'unexpectedUndefined' }],
    },
    {
      code: '[a = undefined] = b',
      errors: [{ messageId: 'unexpectedUndefined' }],
    },
  ],
});

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-undef-init', {
  valid: [
    'var a;',
    'const foo = undefined',
    'class C { field = undefined; }',
    'using foo = undefined',
  ],
  invalid: [
    {
      code: 'var a = undefined;',
      errors: [{ messageId: 'unnecessaryUndefinedInit' }],
    },
    {
      code: 'var a = undefined, b = 1;',
      errors: [{ messageId: 'unnecessaryUndefinedInit' }],
    },
    {
      code: 'var a = 1, b = undefined, c = 5;',
      errors: [{ messageId: 'unnecessaryUndefinedInit' }],
    },
    {
      code: 'var [a] = undefined;',
      errors: [{ messageId: 'unnecessaryUndefinedInit' }],
    },
    {
      code: 'var {a} = undefined;',
      errors: [{ messageId: 'unnecessaryUndefinedInit' }],
    },
    {
      code: 'for(var i in [1,2,3]){var a = undefined; for(var j in [1,2,3]){}}',
      errors: [{ messageId: 'unnecessaryUndefinedInit' }],
    },
    {
      code: 'let a = undefined;',
      errors: [{ messageId: 'unnecessaryUndefinedInit' }],
    },
    {
      code: 'let a = undefined, b = 1;',
      errors: [{ messageId: 'unnecessaryUndefinedInit' }],
    },
    {
      code: 'let a = 1, b = undefined, c = 5;',
      errors: [{ messageId: 'unnecessaryUndefinedInit' }],
    },
    {
      code: 'let [a] = undefined;',
      errors: [{ messageId: 'unnecessaryUndefinedInit' }],
    },
    {
      code: 'let {a} = undefined;',
      errors: [{ messageId: 'unnecessaryUndefinedInit' }],
    },
    {
      code: 'for(var i in [1,2,3]){let a = undefined; for(var j in [1,2,3]){}}',
      errors: [{ messageId: 'unnecessaryUndefinedInit' }],
    },
  ],
});

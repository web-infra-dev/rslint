import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-buffer-constructor', {
  valid: [
    'Buffer.alloc(5)',
    'Buffer.allocUnsafe(5)',
    'new Buffer.Foo()',
    'Buffer.from([1, 2, 3])',
    'foo(Buffer)',
    'Buffer.alloc(res.body.amount)',
    'Buffer.from(res.body.values)',
  ],
  invalid: [
    {
      code: 'Buffer(5)',
      errors: [{ messageId: 'deprecated' }],
    },
    {
      code: 'new Buffer(5)',
      errors: [{ messageId: 'deprecated' }],
    },
    {
      code: 'Buffer([1, 2, 3])',
      errors: [{ messageId: 'deprecated' }],
    },
    {
      code: 'new Buffer([1, 2, 3])',
      errors: [{ messageId: 'deprecated' }],
    },
    {
      code: 'new Buffer(res.body.amount)',
      errors: [{ messageId: 'deprecated' }],
    },
    {
      code: 'new Buffer(res.body.values)',
      errors: [{ messageId: 'deprecated' }],
    },
  ],
});

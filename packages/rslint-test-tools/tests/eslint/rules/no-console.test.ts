import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-console', {
  valid: [
    {
      code: 'console.warn("test");',
      options: { allow: ['warn'] },
    },
    {
      code: 'console.error("test");',
      options: { allow: ['error'] },
    },
    {
      code: 'console.warn("test"); console.error("test");',
      options: { allow: ['warn', 'error'] },
    },
    'var x = { console: 1 };',
    // Shadowed console should not be reported
    'function f(console: any) { console.log("x"); }',
  ],
  invalid: [
    {
      code: 'console.log("test");',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'console.warn("test");',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'console.error("test");',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'console.info("test");',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'console.log("test");',
      options: { allow: ['warn'] },
      errors: [{ messageId: 'unexpected' }],
    },
    // Computed property access
    {
      code: 'console["log"]("test");',
      errors: [{ messageId: 'unexpected' }],
    },
    // Chained member access
    {
      code: 'console.log.bind(null)("test");',
      errors: [{ messageId: 'unexpected' }],
    },
  ],
});

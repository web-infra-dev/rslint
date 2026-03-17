import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-global-assign', {
  valid: ["string = 'hello world';", 'var String: any;'],
  invalid: [
    {
      code: "String = 'hello world';",
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },
    {
      code: 'String++;',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },
    {
      code: 'Array = 1;',
      errors: [{ messageId: 'globalShouldNotBeModified' }],
    },
  ],
});

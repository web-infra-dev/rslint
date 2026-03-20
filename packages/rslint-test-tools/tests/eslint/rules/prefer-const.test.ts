import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-const', {
  valid: [
    'const x = 1;',
    'let x = 1; x = 2;',
    'let x = 1; x++;',
    'let x: number;',
    'var x = 1;',
  ],
  invalid: [
    {
      code: 'let x = 1;',
      errors: [{ messageId: 'useConst' }],
    },
    {
      code: "let x = 'hello';",
      errors: [{ messageId: 'useConst' }],
    },
  ],
});

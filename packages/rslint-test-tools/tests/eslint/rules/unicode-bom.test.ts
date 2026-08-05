import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const BOM = '\ufeff';

ruleTester.run('unicode-bom', {
  valid: [
    { code: 'var a = 123;', options: ['never'] as any },
    { code: `var a = 123; ${BOM}`, options: ['never'] as any },
  ],
  invalid: [
    {
      code: `${BOM} var a = 123;`,
      errors: [{ messageId: 'unexpected', line: 1, column: 1 }],
    },
    {
      code: `${BOM} var a = 123;`,
      options: ['never'] as any,
      errors: [{ messageId: 'unexpected', line: 1, column: 1 }],
    },
  ],
});

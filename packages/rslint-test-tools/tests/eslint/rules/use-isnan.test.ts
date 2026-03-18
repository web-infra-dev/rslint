import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('use-isnan', {
  valid: [
    'var x = NaN;',
    'isNaN(NaN) === true;',
    'Number.isNaN(NaN) === true;',
  ],
  invalid: [
    {
      code: '123 == NaN;',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: '123 === NaN;',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
    {
      code: 'NaN !== 123;',
      errors: [{ messageId: 'comparisonWithNaN' }],
    },
  ],
});

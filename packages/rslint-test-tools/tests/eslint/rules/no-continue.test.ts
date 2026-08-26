import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-continue', {
  valid: [
    'var sum = 0, i; for(i = 0; i < 10; i++){ if(i > 5) { sum += i; } }',
    'var sum = 0, i = 0; while(i < 10) { if(i > 5) { sum += i; } i++; }',
  ],
  invalid: [
    {
      code: 'var sum = 0, i; for(i = 0; i < 10; i++){ if(i <= 5) { continue; } sum += i; }',
      errors: [{ messageId: 'unexpected', line: 1, column: 55 }],
    },
    {
      code: 'var sum = 0, i; myLabel: for(i = 0; i < 10; i++){ if(i <= 5) { continue myLabel; } sum += i; }',
      errors: [{ messageId: 'unexpected', line: 1, column: 64 }],
    },
    {
      code: 'var sum = 0, i = 0; while(i < 10) { if(i <= 5) { i++; continue; } sum += i; i++; }',
      errors: [{ messageId: 'unexpected', line: 1, column: 55 }],
    },
    {
      code: 'var sum = 0, i = 0; myLabel: while(i < 10) { if(i <= 5) { i++; continue myLabel; } sum += i; i++; }',
      errors: [{ messageId: 'unexpected', line: 1, column: 64 }],
    },
  ],
});

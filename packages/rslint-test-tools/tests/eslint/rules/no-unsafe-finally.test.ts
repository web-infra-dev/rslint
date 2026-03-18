import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-unsafe-finally', {
  valid: [
    'try { return 1; } catch(err) { return 2; } finally { console.log("done") }',
    'try {} finally { function a(x) { return x } }',
  ],
  invalid: [
    {
      code: 'try { return 1; } finally { return 3; }',
      errors: [{ messageId: 'unsafeUsage' }],
    },
    {
      code: 'try {} finally { throw new Error() }',
      errors: [{ messageId: 'unsafeUsage' }],
    },
  ],
});

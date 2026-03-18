import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-dupe-else-if', {
  valid: [
    'if (a) {} else if (b) {}',
    'if (a === 1) {} else if (a === 2) {}',
    'if (a || b) {} else if (c) {}',
    'if (a && b) {} else if (a && c) {}',
  ],
  invalid: [
    {
      code: 'if (a) {} else if (a) {}',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'if (a) {} else if (b) {} else if (a) {}',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'if (a || b) {} else if (a) {}',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'if (a) {} else if (a && b) {}',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'if (a && b) {} else if (b && a) {}',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'if (a || b) {} else if (b || a) {}',
      errors: [{ messageId: 'unexpected' }],
    },
  ],
});

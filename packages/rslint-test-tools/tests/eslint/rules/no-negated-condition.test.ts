import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-negated-condition', {
  valid: [
    'if (a) {}',
    'if (a) {} else {}',
    'if (!a) {}',
    'if (!a) {} else if (b) {}',
    'if (!a) {} else if (b) {} else {}',
    'if (a == b) {}',
    'if (a == b) {} else {}',
    'if (a != b) {}',
    'if (a != b) {} else if (b) {}',
    'if (a != b) {} else if (b) {} else {}',
    'if (a !== b) {}',
    'if (a === b) {} else {}',
    'a ? b : c',
  ],
  invalid: [
    {
      code: 'if (!a) {;} else {;}',
      errors: [{ messageId: 'unexpectedNegated' }],
    },
    {
      code: 'if (a != b) {;} else {;}',
      errors: [{ messageId: 'unexpectedNegated' }],
    },
    {
      code: 'if (a !== b) {;} else {;}',
      errors: [{ messageId: 'unexpectedNegated' }],
    },
    {
      code: '!a ? b : c',
      errors: [{ messageId: 'unexpectedNegated' }],
    },
    {
      code: 'a != b ? c : d',
      errors: [{ messageId: 'unexpectedNegated' }],
    },
    {
      code: 'a !== b ? c : d',
      errors: [{ messageId: 'unexpectedNegated' }],
    },
  ],
});

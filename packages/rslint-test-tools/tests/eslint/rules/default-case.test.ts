import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('default-case', {
  valid: [
    'switch (a) { case 1: break; default: break; }',
    'switch (a) { case 1: break; case 2: default: break; }',
    // Line comment exemption
    `switch (a) {
  case 1:
    break;
  // no default
}`,
    'switch (a) { case 1: break; /* no default */ }',
    // Case-insensitive
    `switch (a) {
  case 1:
    break;
  // No Default
}`,
    'switch (a) {}',
  ],
  invalid: [
    {
      code: 'switch (a) { case 1: break; }',
      errors: [{ messageId: 'missingDefaultCase' }],
    },
    {
      code: 'switch (a) { case 1: break; case 2: break; }',
      errors: [{ messageId: 'missingDefaultCase' }],
    },
  ],
});

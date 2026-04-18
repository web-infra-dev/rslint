import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-octal', {
  valid: [
    "var a = 'hello world';",
    '0x1234',
    '0X5;',
    'a = 0;',
    '0.1',
    '0.5e1',
    '0o17',
    '0O17',
    '0b101',
    '0B101',
  ],
  // Invalid cases cannot be tested through the JS test framework because the
  // TypeScript parser reports octal literals (TS1121) and leading-zero decimals
  // (TS1489) as syntax errors, preventing program creation. The detection logic
  // is comprehensively tested via Go unit tests (TestIsOctalLiteralRaw in
  // no_octal_test.go).
  invalid: [],
});

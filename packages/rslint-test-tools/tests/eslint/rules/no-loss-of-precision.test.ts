import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-loss-of-precision', {
  valid: [
    'const x = 12345;',
    'const x = 123.456;',
    'const x = -123.456;',
    'const x = 123e34;',
    'const x = 0x1FFFFFFFFFFFFF;',
    'const x = 0o377777777777777777;',
    'const x = 0b11111111111111111111111111111111111111111111111111111;',
    'const x = 9007199254740991;',
    'const x = 123_456;',
    'const x = 123_00_000_000_000_000_000_000_000;',
    'const x = 123.000_000_000_000_000_000_000_0;',
  ],
  invalid: [
    {
      code: 'const x = 9007199254740993;',
      errors: [{ messageId: 'noLossOfPrecision' }],
    },
    {
      code: 'const x = 9_007_199_254_740_993;',
      errors: [{ messageId: 'noLossOfPrecision' }],
    },
    {
      code: 'const x = 9_007_199_254_740.993e3;',
      errors: [{ messageId: 'noLossOfPrecision' }],
    },
    {
      code: 'const x = 0x20000000000001;',
      errors: [{ messageId: 'noLossOfPrecision' }],
    },
    {
      code: 'const x = 0o400000000000000001;',
      errors: [{ messageId: 'noLossOfPrecision' }],
    },
    {
      code: 'const x = 0b100_000_000_000_000_000_000_000_000_000_000_000_000_000_000_000_000_001;',
      errors: [{ messageId: 'noLossOfPrecision' }],
    },
    {
      code: 'const x = 2e999;',
      errors: [{ messageId: 'noLossOfPrecision' }],
    },
  ],
});

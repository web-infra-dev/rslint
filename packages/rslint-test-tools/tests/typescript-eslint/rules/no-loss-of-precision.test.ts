import { RuleTester } from '../RuleTester.ts';

const rootPath = getFixturesRootDir();
const ruleTester = new RuleTester();

ruleTester.run('no-loss-of-precision', {
  valid: [
    'const x = 12345;',
    'const x = 123.456;',
    'const x = -123.456;',
    'const x = 123_456;',
    'const x = 123_00_000_000_000_000_000_000_000;',
    'const x = 123.000_000_000_000_000_000_000_0;',
    'const x = 0x1234;',
    'const x = 0b1010;',
    'const x = 0o777;',
    'const x = 9007199254740991;', // MAX_SAFE_INTEGER
    'const x = -9007199254740991;', // -MAX_SAFE_INTEGER
    'const x = 900719925474099.1;',
    'const x = 9.007199254740991e15;',
    'const x = 0xFFFF_FFFF_FFFF;', // Still within safe range
    'const x = 0o377777777777777;', // Still within safe range
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
      code: 'const x = 0b100_000_000_000_000_000_000_000_000_000_000_000_000_000_000_000_000_001;',
      errors: [{ messageId: 'noLossOfPrecision' }],
    },
    {
      code: 'const x = -9007199254740993;',
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
      code: 'const x = 18014398509481984;', // 2^54
      errors: [{ messageId: 'noLossOfPrecision' }],
    },
    {
      code: 'const x = 0x40000000000000;', // 2^54 in hex
      errors: [{ messageId: 'noLossOfPrecision' }],
    },
    {
      code: 'const x = 0b1000000000000000000000000000000000000000000000000000000;', // 2^54 in binary
      errors: [{ messageId: 'noLossOfPrecision' }],
    },
  ],
});
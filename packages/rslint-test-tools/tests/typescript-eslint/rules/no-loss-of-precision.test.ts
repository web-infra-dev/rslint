import { noFormat, RuleTester } from '@typescript-eslint/rule-tester';
import { getFixturesRootDir } from '../RuleTester.ts';

const rootPath = getFixturesRootDir();

const ruleTester = new RuleTester({
  // @ts-ignore
  languageOptions: {
    parserOptions: {
      project: './tsconfig.json',
      tsconfigRootDir: rootPath,
    },
  },
});

ruleTester.run('no-loss-of-precision', {
  valid: [
    'const x = 12345;',
    'const x = 123.456;',
    'const x = -123.456;',
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
      code: 'const x = 0b100_000_000_000_000_000_000_000_000_000_000_000_000_000_000_000_000_001;',
      errors: [{ messageId: 'noLossOfPrecision' }],
    },
  ],
});

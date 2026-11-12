import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('consistent-test-filename', {
  valid: [
    { code: 'export {}', filename: '1.test.ts' },
    {
      code: 'export {}',
      filename: '1.spec.ts',
      options: [{ pattern: String.raw`.*\.spec\.ts$` }],
    },
  ],
  invalid: [
    {
      code: 'export {}',
      filename: '1.spec.ts',
      errors: [
        {
          messageId: 'consistentTestFilename',
          message: 'Use test file name pattern .*\\.test\\.(c|m)?[tj]sx?$',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: 'export {}',
      filename: '__tests__/2.ts',
      options: [
        {
          allTestPattern: String.raw`__tests__`,
          pattern: String.raw`.*\.spec\.ts$`,
        },
      ],
      errors: [
        {
          messageId: 'consistentTestFilename',
          message: 'Use test file name pattern .*\\.spec\\.ts$',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
  ],
});

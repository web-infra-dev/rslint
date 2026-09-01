import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('consistent-test-filename', {} as never, {
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
          message: String.raw`Use test file name pattern .*\.test\.(c|m)?[tj]sx?$`,
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 1,
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
          message: String.raw`Use test file name pattern .*\.spec\.ts$`,
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 1,
        },
      ],
    },
  ],
});

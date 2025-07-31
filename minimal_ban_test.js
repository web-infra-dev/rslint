import { RuleTester } from './packages/rslint-test-tools/tests/typescript-eslint/RuleTester.ts';

const ruleTester = new RuleTester({
  languageOptions: {
    parserOptions: {
      project: './tsconfig.json',
    },
  },
});

// Test just one simple case
ruleTester.run('ban-ts-comment', {
  valid: ['// just a comment containing @ts-expect-error somewhere'],
  invalid: [
    {
      code: '// @ts-expect-error',
      errors: [
        {
          line: 1,
          messageId: 'tsDirectiveCommentRequiresDescription',
        },
      ],
    },
  ],
});

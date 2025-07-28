import { RuleTester } from './packages/rslint-test-tools/tests/typescript-eslint/RuleTester.ts';

const ruleTester = new RuleTester({
  languageOptions: {
    parserOptions: {
      project: false,
    },
  },
});

console.log('Testing with debug...');

ruleTester.run('@typescript-eslint/explicit-member-accessibility', {
  valid: [],
  invalid: [
    {
      code: `
export class XXXX {
  public constructor(readonly value: string) {}
}
      `,
      errors: [{ messageId: 'missingAccessibility' }],
      options: [
        {
          accessibility: 'off',
          overrides: {
            parameterProperties: 'explicit',
          },
        },
      ],
    },
  ],
});

console.log('Debug test completed!');
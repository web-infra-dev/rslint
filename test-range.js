import { RuleTester } from './packages/rslint-test-tools/tests/typescript-eslint/RuleTester.ts';

const ruleTester = new RuleTester({
  languageOptions: {
    parserOptions: {
      project: false,
    },
  },
});

console.log('Testing range reporting...');

ruleTester.run('@typescript-eslint/explicit-member-accessibility', {
  valid: [],
  invalid: [
    {
      code: `
export class XXXX {
  public constructor(readonly value: string) {}
}
      `,
      errors: [
        {
          column: 22,
          endColumn: 36,
          endLine: 3,
          line: 3,
          messageId: 'missingAccessibility',
        },
      ],
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

console.log('Range test completed successfully!');
import { RuleTester } from './packages/rslint-test-tools/tests/typescript-eslint/RuleTester.ts';

const ruleTester = new RuleTester({
  languageOptions: {
    parserOptions: {
      project: false,
    },
  },
});

console.log('Testing explicit-member-accessibility rule subset...');

ruleTester.run('@typescript-eslint/explicit-member-accessibility', {
  valid: [
    // The failing test case
    {
      code: `
class Test {
  constructor(public foo: number) {}
}
      `,
      options: [{ accessibility: 'no-public' }],
    },
  ],
  invalid: [
    // One simple invalid case
    {
      code: `
class Test {
  x: number;
}
      `,
      errors: [
        {
          column: 3,
          endColumn: 4,
          endLine: 3,
          line: 3,
          messageId: 'missingAccessibility',
        },
      ],
    },
  ],
});

console.log('Test subset completed successfully!');
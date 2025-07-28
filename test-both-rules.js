import { RuleTester } from './packages/rslint-test-tools/tests/typescript-eslint/RuleTester.ts';

const ruleTester = new RuleTester({
  languageOptions: {
    parserOptions: {
      project: false,
    },
  },
});

console.log('Testing explicit-member-accessibility...');

// Test the originally failing case for explicit-member-accessibility
ruleTester.run('@typescript-eslint/explicit-member-accessibility', {
  valid: [
    {
      code: `
class Test {
  constructor(public foo: number) {}
}
      `,
      options: [{ accessibility: 'no-public' }],
    },
  ],
  invalid: [],
});

console.log('explicit-member-accessibility test passed!');

// Test a simple case for explicit-module-boundary-types
console.log('Testing explicit-module-boundary-types...');

const ruleTester2 = new RuleTester({
  languageOptions: {
    parserOptions: {
      project: './packages/rslint-test-tools/tests/typescript-eslint/fixtures/tsconfig.json',
      projectService: false,
    },
  },
});

ruleTester2.run('explicit-module-boundary-types', {
  valid: [
    {
      code: `
function test(): void {
  return;
}
      `,
    },
  ],
  invalid: [],
});

console.log('Both rules tested successfully!');
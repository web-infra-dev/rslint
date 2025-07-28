import { RuleTester } from './packages/rslint-test-tools/tests/typescript-eslint/RuleTester.ts';

console.log('Starting simple test...');

const ruleTester = new RuleTester({
  languageOptions: {
    parserOptions: {
      project: false,
    },
  },
});

console.log('RuleTester created');

const testResult = ruleTester.run('@typescript-eslint/explicit-member-accessibility', {
  valid: [
    {
      code: `class Test { constructor(public foo: number) {} }`,
      options: [{ accessibility: 'no-public' }],
    },
  ],
  invalid: [],
});

console.log('Test completed:', testResult);
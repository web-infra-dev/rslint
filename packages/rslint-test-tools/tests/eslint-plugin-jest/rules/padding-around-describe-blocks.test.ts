// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

// Shared source templates; reporting ranges remain explicit in each matrix.
const testSetup = `const someText = 'abc';
const someObject = {
  one: 1,
  two: 2,
};

// A comment before describe
describe('someText', () => {
  describe('some condition', () => {
  });

  describe('some other condition', () => {
  });
});

xdescribe('someObject', () => {
  // Another comment
  describe('some condition', () => {
    const anotherThing = 500;

    describe('yet another condition', () => { // A comment over here!
    });
  });
});

fdescribe('weird', () => {});

describe.skip('skip me', () => {});

const BOOP = "boop";

describe
  .skip('skip me too', () => {
    // stuff
  });
`;

new RuleTester(fixtureConfig()).run('padding-around-describe-blocks', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: `setup();

describe('suite', run);`,
    },
    {
      name: 'upstream padding-around-describe-blocks / valid 0',
      code:
        `
foo();
bar();

` + testSetup,
      languageOptions: {
        ecmaVersion: 6,
      },
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: `setup();
describe('suite', run);`,
      errors: [
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 24,
        },
      ],
      output: `setup();

describe('suite', run);`,
    },
    {
      name: 'upstream padding-around-describe-blocks / invalid 0',
      code: `
foo();
bar();

const someText = 'abc';
const someObject = {
  one: 1,
  two: 2,
};
// A comment before describe
describe('someText', () => {
  describe('some condition', () => {
  });
  describe('some other condition', () => {
  });
});
xdescribe('someObject', () => {
  // Another comment
  describe('some condition', () => {
    const anotherThing = 500;
    describe('yet another condition', () => { // A comment over here!
    });
  });
});fdescribe('weird', () => {});
describe.skip('skip me', () => {});
const BOOP = "boop";
describe
  .skip('skip me too', () => {
    // stuff
  });
`,
      filename: 'src/component.test.jsx',
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 11,
          column: 1,
          endLine: 16,
          endColumn: 4,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 14,
          column: 3,
          endLine: 15,
          endColumn: 6,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 17,
          column: 1,
          endLine: 24,
          endColumn: 4,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 21,
          column: 5,
          endLine: 22,
          endColumn: 8,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 24,
          column: 4,
          endLine: 24,
          endColumn: 33,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 25,
          column: 1,
          endLine: 25,
          endColumn: 36,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 26,
          column: 1,
          endLine: 26,
          endColumn: 21,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 27,
          column: 1,
          endLine: 30,
          endColumn: 6,
        },
      ],
      output:
        `
foo();
bar();

` + testSetup,
    },
    {
      name: 'upstream padding-around-describe-blocks / invalid 1',
      code: `
foo();
bar();

const someText = 'abc';
const someObject = {
  one: 1,
  two: 2,
};
// A comment before describe
describe('someText', () => {
  describe('some condition', () => {
  });
  describe('some other condition', () => {
  });
});
xdescribe('someObject', () => {
  // Another comment
  describe('some condition', () => {
    const anotherThing = 500;
    describe('yet another condition', () => { // A comment over here!
    });
  });
});fdescribe('weird', () => {});
describe.skip('skip me', () => {});
const BOOP = "boop";
describe
  .skip('skip me too', () => {
    // stuff
  });
`,
      filename: 'src/component.test.js',
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 11,
          column: 1,
          endLine: 16,
          endColumn: 4,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 14,
          column: 3,
          endLine: 15,
          endColumn: 6,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 17,
          column: 1,
          endLine: 24,
          endColumn: 4,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 21,
          column: 5,
          endLine: 22,
          endColumn: 8,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 24,
          column: 4,
          endLine: 24,
          endColumn: 33,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 25,
          column: 1,
          endLine: 25,
          endColumn: 36,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 26,
          column: 1,
          endLine: 26,
          endColumn: 21,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 27,
          column: 1,
          endLine: 30,
          endColumn: 6,
        },
      ],
      output:
        `
foo();
bar();

` + testSetup,
    },
  ],
});

// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

// Shared source templates; reporting ranges remain explicit in each matrix.
const testSetup = `
it('foo', () => {
  // stuff
});

fit('bar', () => {
  // stuff
});

test('foo foo', () => {});

test('bar bar', () => {});

// Nesting
describe('other bar', () => {
  const thing = 123;

  test('is another bar w/ test', () => {
  });

  // With a comment
  it('is another bar w/ it', () => {
  });

  test.skip('skipping', () => {}); // Another comment

  it.skip('skipping too', () => {});
});

xtest('weird', () => {});

test
  .skip('skippy skip', () => {});

xit('bar foo', () => {});
`;

new RuleTester(fixtureConfig()).run('padding-around-test-blocks', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: `setup();

test('works', run);`,
    },
    {
      name: 'upstream padding-around-test-blocks / valid 0',
      code:
        `
const foo = 'bar';
const bar = 'baz';
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
test('works', run);`,
      errors: [
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 20,
        },
      ],
      output: `setup();

test('works', run);`,
    },
    {
      name: 'upstream padding-around-test-blocks / invalid 0',
      code: `
const foo = 'bar';
const bar = 'baz';
it('foo', () => {
  // stuff
});
fit('bar', () => {
  // stuff
});
test('foo foo', () => {});
test('bar bar', () => {});

// Nesting
describe('other bar', () => {
  const thing = 123;
  test('is another bar w/ test', () => {
  });
  // With a comment
  it('is another bar w/ it', () => {
  });
  test.skip('skipping', () => {}); // Another comment
  it.skip('skipping too', () => {});
});xtest('weird', () => {});
test
  .skip('skippy skip', () => {});
xit('bar foo', () => {});
`,
      filename: 'src/component.test.jsx',
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 4,
          column: 1,
          endLine: 6,
          endColumn: 4,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 7,
          column: 1,
          endLine: 9,
          endColumn: 4,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 10,
          column: 1,
          endLine: 10,
          endColumn: 27,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 11,
          column: 1,
          endLine: 11,
          endColumn: 27,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 16,
          column: 3,
          endLine: 17,
          endColumn: 6,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 19,
          column: 3,
          endLine: 20,
          endColumn: 6,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 21,
          column: 3,
          endLine: 21,
          endColumn: 35,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 22,
          column: 3,
          endLine: 22,
          endColumn: 37,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 23,
          column: 4,
          endLine: 23,
          endColumn: 29,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 24,
          column: 1,
          endLine: 25,
          endColumn: 34,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 26,
          column: 1,
          endLine: 26,
          endColumn: 26,
        },
      ],
      output:
        `
const foo = 'bar';
const bar = 'baz';
` + testSetup,
    },
    {
      name: 'upstream padding-around-test-blocks / invalid 1',
      code: `
const foo = 'bar';
const bar = 'baz';
it('foo', () => {
  // stuff
});
fit('bar', () => {
  // stuff
});
test('foo foo', () => {});
test('bar bar', () => {});

// Nesting
describe('other bar', () => {
  const thing = 123;
  test('is another bar w/ test', () => {
  });
  // With a comment
  it('is another bar w/ it', () => {
  });
  test.skip('skipping', () => {}); // Another comment
  it.skip('skipping too', () => {});
});xtest('weird', () => {});
test
  .skip('skippy skip', () => {});
xit('bar foo', () => {});
`,
      filename: 'src/component.test.js',
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 4,
          column: 1,
          endLine: 6,
          endColumn: 4,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 7,
          column: 1,
          endLine: 9,
          endColumn: 4,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 10,
          column: 1,
          endLine: 10,
          endColumn: 27,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 11,
          column: 1,
          endLine: 11,
          endColumn: 27,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 16,
          column: 3,
          endLine: 17,
          endColumn: 6,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 19,
          column: 3,
          endLine: 20,
          endColumn: 6,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 21,
          column: 3,
          endLine: 21,
          endColumn: 35,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 22,
          column: 3,
          endLine: 22,
          endColumn: 37,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 23,
          column: 4,
          endLine: 23,
          endColumn: 29,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 24,
          column: 1,
          endLine: 25,
          endColumn: 34,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 26,
          column: 1,
          endLine: 26,
          endColumn: 26,
        },
      ],
      output:
        `
const foo = 'bar';
const bar = 'baz';
` + testSetup,
    },
  ],
});

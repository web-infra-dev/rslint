// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('padding-around-after-each-blocks', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: `setup();

afterEach(cleanup);`,
    },
    {
      name: 'upstream padding-around-after-each-blocks / valid 0',
      code: `
const someText = 'abc';

afterEach(() => {
});

describe('someText', () => {
  const something = 'abc';

  // A comment
  afterEach(() => {
    // stuff
  });

  afterEach(() => {
    // other stuff
  });
});
describe('someText', () => {
  const something = 'abc';

  afterEach(() => {
    // stuff
  });
});
`,
      languageOptions: {
        ecmaVersion: 6,
      },
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: `setup();
afterEach(cleanup);`,
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

afterEach(cleanup);`,
    },
    {
      name: 'upstream padding-around-after-each-blocks / invalid 0',
      code: `
const someText = 'abc';
afterEach(() => {
});
describe('someText', () => {
  const something = 'abc';
  // A comment
  afterEach(() => {
    // stuff
  });
  afterEach(() => {
    // other stuff
  });
});
describe('someText', () => {
  const something = 'abc';
  afterEach(() => {
    // stuff
  });
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
          line: 3,
          column: 1,
          endLine: 4,
          endColumn: 4,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 5,
          column: 1,
          endLine: 14,
          endColumn: 4,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 8,
          column: 3,
          endLine: 10,
          endColumn: 6,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 11,
          column: 3,
          endLine: 13,
          endColumn: 6,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 17,
          column: 3,
          endLine: 19,
          endColumn: 6,
        },
      ],
      output: `
const someText = 'abc';

afterEach(() => {
});

describe('someText', () => {
  const something = 'abc';

  // A comment
  afterEach(() => {
    // stuff
  });

  afterEach(() => {
    // other stuff
  });
});
describe('someText', () => {
  const something = 'abc';

  afterEach(() => {
    // stuff
  });
});
`,
    },
    {
      name: 'upstream padding-around-after-each-blocks / invalid 1',
      code: `
const someText = 'abc';
afterEach(() => {
});
describe('someText', () => {
  const something = 'abc';
  // A comment
  afterEach(() => {
    // stuff
  });
  afterEach(() => {
    // other stuff
  });
});
describe('someText', () => {
  const something = 'abc';
  afterEach(() => {
    // stuff
  });
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
          line: 3,
          column: 1,
          endLine: 4,
          endColumn: 4,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 5,
          column: 1,
          endLine: 14,
          endColumn: 4,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 8,
          column: 3,
          endLine: 10,
          endColumn: 6,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 11,
          column: 3,
          endLine: 13,
          endColumn: 6,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 17,
          column: 3,
          endLine: 19,
          endColumn: 6,
        },
      ],
      output: `
const someText = 'abc';

afterEach(() => {
});

describe('someText', () => {
  const something = 'abc';

  // A comment
  afterEach(() => {
    // stuff
  });

  afterEach(() => {
    // other stuff
  });
});
describe('someText', () => {
  const something = 'abc';

  afterEach(() => {
    // stuff
  });
});
`,
    },
  ],
});

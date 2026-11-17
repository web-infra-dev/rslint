// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('padding-around-before-all-blocks', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: `setup();

beforeAll(connect);`,
    },
    {
      name: 'upstream padding-around-before-all-blocks / valid 0',
      code: `
const someText = 'abc';

beforeAll(() => {
});

describe('someText', () => {
  const something = 'abc';

  // A comment
  beforeAll(() => {
    // stuff
  });

  beforeAll(() => {
    // other stuff
  });
});

describe('someText', () => {
  const something = 'abc';

  beforeAll(() => {
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
beforeAll(connect);`,
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

beforeAll(connect);`,
    },
    {
      name: 'upstream padding-around-before-all-blocks / invalid 0',
      code: `
const someText = 'abc';
beforeAll(() => {
});
describe('someText', () => {
  const something = 'abc';
  // A comment
  beforeAll(() => {
    // stuff
  });
  beforeAll(() => {
    // other stuff
  });
});

describe('someText', () => {
  const something = 'abc';
  beforeAll(() => {
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
          line: 18,
          column: 3,
          endLine: 20,
          endColumn: 6,
        },
      ],
      output: `
const someText = 'abc';

beforeAll(() => {
});

describe('someText', () => {
  const something = 'abc';

  // A comment
  beforeAll(() => {
    // stuff
  });

  beforeAll(() => {
    // other stuff
  });
});

describe('someText', () => {
  const something = 'abc';

  beforeAll(() => {
    // stuff
  });
});
`,
    },
    {
      name: 'upstream padding-around-before-all-blocks / invalid 1',
      code: `
const someText = 'abc';
beforeAll(() => {
});
describe('someText', () => {
  const something = 'abc';
  // A comment
  beforeAll(() => {
    // stuff
  });
  beforeAll(() => {
    // other stuff
  });
});

describe('someText', () => {
  const something = 'abc';
  beforeAll(() => {
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
          line: 18,
          column: 3,
          endLine: 20,
          endColumn: 6,
        },
      ],
      output: `
const someText = 'abc';

beforeAll(() => {
});

describe('someText', () => {
  const something = 'abc';

  // A comment
  beforeAll(() => {
    // stuff
  });

  beforeAll(() => {
    // other stuff
  });
});

describe('someText', () => {
  const something = 'abc';

  beforeAll(() => {
    // stuff
  });
});
`,
    },
  ],
});

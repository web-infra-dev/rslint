// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('padding-around-before-each-blocks', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: `setup();

beforeEach(reset);`,
    },
    {
      name: 'upstream padding-around-before-each-blocks / valid 0',
      code: `
const someText = 'abc';

beforeEach(() => {
});

describe('someText', () => {
  const something = 'abc';

  // A comment
  beforeEach(() => {
    // stuff
  });

  beforeEach(() => {
    // other stuff
  });
});

describe('someText', () => {
  const something = 'abc';

  beforeEach(() => {
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
beforeEach(reset);`,
      errors: [
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 19,
        },
      ],
      output: `setup();

beforeEach(reset);`,
    },
    {
      name: 'upstream padding-around-before-each-blocks / invalid 0',
      code: `
const someText = 'abc';
beforeEach(() => {
});
describe('someText', () => {
  const something = 'abc';
  // A comment
  beforeEach(() => {
    // stuff
  });
  beforeEach(() => {
    // other stuff
  });
});

describe('someText', () => {
  const something = 'abc';
  beforeEach(() => {
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

beforeEach(() => {
});

describe('someText', () => {
  const something = 'abc';

  // A comment
  beforeEach(() => {
    // stuff
  });

  beforeEach(() => {
    // other stuff
  });
});

describe('someText', () => {
  const something = 'abc';

  beforeEach(() => {
    // stuff
  });
});
`,
    },
    {
      name: 'upstream padding-around-before-each-blocks / invalid 1',
      code: `
const someText = 'abc';
beforeEach(() => {
});
describe('someText', () => {
  const something = 'abc';
  // A comment
  beforeEach(() => {
    // stuff
  });
  beforeEach(() => {
    // other stuff
  });
});

describe('someText', () => {
  const something = 'abc';
  beforeEach(() => {
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

beforeEach(() => {
});

describe('someText', () => {
  const something = 'abc';

  // A comment
  beforeEach(() => {
    // stuff
  });

  beforeEach(() => {
    // other stuff
  });
});

describe('someText', () => {
  const something = 'abc';

  beforeEach(() => {
    // stuff
  });
});
`,
    },
  ],
});

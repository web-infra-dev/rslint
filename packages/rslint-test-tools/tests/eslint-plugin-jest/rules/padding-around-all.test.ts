// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('padding-around-all', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: `setup();

beforeAll(connect);

test('works', run);`,
    },
    ...(
      [
        `
const someText = 'abc';

afterAll(() => {
});

describe('someText', () => {
  const something = 'abc';

  // A comment
  afterAll(() => {
    // stuff
  });

  afterAll(() => {
    // other stuff
  });
});

describe('someText', () => {
  const something = 'abc';

  afterAll(() => {
    // stuff
  });
});
`,
        `xyz:
afterEach(() => {});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream padding-around-all / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 6,
      },
    })),
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: `setup();
beforeAll(connect);
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
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 20,
        },
      ],
      output: `setup();

beforeAll(connect);

test('works', run);`,
    },
    {
      name: 'upstream padding-around-all / invalid 0',
      code: `
const someText = 'abc';
afterAll(() => {
});
describe('someText', () => {
  const something = 'abc';
  // A comment
  afterAll(() => {
    // stuff
  });
  afterAll(() => {
    // other stuff
  });
});

describe('someText', () => {
  const something = 'abc';
  afterAll(() => {
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

afterAll(() => {
});

describe('someText', () => {
  const something = 'abc';

  // A comment
  afterAll(() => {
    // stuff
  });

  afterAll(() => {
    // other stuff
  });
});

describe('someText', () => {
  const something = 'abc';

  afterAll(() => {
    // stuff
  });
});
`,
    },
    {
      name: 'upstream padding-around-all / invalid 1',
      code: `
const someText = 'abc';
afterAll(() => {
});
describe('someText', () => {
  const something = 'abc';
  // A comment
  afterAll(() => {
    // stuff
  });
  afterAll(() => {
    // other stuff
  });
});

describe('someText', () => {
  const something = 'abc';
  afterAll(() => {
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

afterAll(() => {
});

describe('someText', () => {
  const something = 'abc';

  // A comment
  afterAll(() => {
    // stuff
  });

  afterAll(() => {
    // other stuff
  });
});

describe('someText', () => {
  const something = 'abc';

  afterAll(() => {
    // stuff
  });
});
`,
    },
    ...(
      [
        [
          `const someText = 'abc'
;afterEach(() => {})`,
          2,
          2,
          `const someText = 'abc'

;afterEach(() => {})`,
        ],
        [
          `const someText = 'abc';
xyz:
afterEach(() => {});`,
          1,
          3,
          `const someText = 'abc';

xyz:
afterEach(() => {});`,
        ],
      ] as const
    ).map(([code, column, endLine, output], index): TestCase => ({
      name: `upstream padding-around-all / invalid ${index + 2}`,
      code,
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 2,
          column,
          endLine,
          endColumn: 21,
        },
      ],
      output,
    })),
    {
      name: 'upstream padding-around-all / invalid 4',
      code: `const expr = 'Papayas';
beforeEach(() => {});
it('does something?', () => {
  switch (expr) {
    case 'Oranges':
      expect(expr).toBe('Oranges');
      break;
    case 'Mangoes':
    case 'Papayas':
      const v = 1;
      expect(v).toBe(1);
      console.log('Mangoes and papayas are $2.79 a pound.');
      // Expected output: "Mangoes and papayas are $2.79 a pound."
      break;
    default:
      console.log(\`Sorry, we are out of \${expr}.\`);
  }
});`,
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 22,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 3,
          column: 1,
          endLine: 18,
          endColumn: 4,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 7,
          column: 7,
          endLine: 7,
          endColumn: 13,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 11,
          column: 7,
          endLine: 11,
          endColumn: 25,
        },
        {
          messageId: 'missingPadding',
          message: 'Expected blank line before this statement.',
          line: 12,
          column: 7,
          endLine: 12,
          endColumn: 61,
        },
      ],
      output: `const expr = 'Papayas';

beforeEach(() => {});

it('does something?', () => {
  switch (expr) {
    case 'Oranges':
      expect(expr).toBe('Oranges');

      break;
    case 'Mangoes':
    case 'Papayas':
      const v = 1;

      expect(v).toBe(1);

      console.log('Mangoes and papayas are $2.79 a pound.');
      // Expected output: "Mangoes and papayas are $2.79 a pound."
      break;
    default:
      console.log(\`Sorry, we are out of \${expr}.\`);
  }
});`,
    },
  ],
});

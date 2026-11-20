// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('require-to-throw-message', {
  valid: [
    ...(
      [
        "expect(() => { throw new Error('a'); }).toThrow('a');",
        "expect(() => { throw new Error('a'); }).toThrowError('a');",
        `
      test('string', async () => {
        const throwErrorAsync = async () => { throw new Error('a') };
        await expect(throwErrorAsync()).rejects.toThrow('a');
        await expect(throwErrorAsync()).rejects.toThrowError('a');
      })
   \u0020`,
        "const a = 'a'; expect(() => { throw new Error('a'); }).toThrow(`${a}`);",
        "const a = 'a'; expect(() => { throw new Error('a'); }).toThrowError(`${a}`);",
        `
      test('Template literal', async () => {
        const a = 'a';
        const throwErrorAsync = async () => { throw new Error('a') };
        await expect(throwErrorAsync()).rejects.toThrow(\`\${a}\`);
        await expect(throwErrorAsync()).rejects.toThrowError(\`\${a}\`);
      })
   \u0020`,
        "expect(() => { throw new Error('a'); }).toThrow(/^a$/);",
        "expect(() => { throw new Error('a'); }).toThrowError(/^a$/);",
        `
      test('Regex', async () => {
        const throwErrorAsync = async () => { throw new Error('a') };
        await expect(throwErrorAsync()).rejects.toThrow(/^a$/);
        await expect(throwErrorAsync()).rejects.toThrowError(/^a$/);
      })
   \u0020`,
        "expect(() => { throw new Error('a'); }).toThrow((() => { return 'a'; })());",
        "expect(() => { throw new Error('a'); }).toThrowError((() => { return 'a'; })());",
        `
      test('Function', async () => {
        const throwErrorAsync = async () => { throw new Error('a') };
        const fn = () => { return 'a'; };
        await expect(throwErrorAsync()).rejects.toThrow(fn());
        await expect(throwErrorAsync()).rejects.toThrowError(fn());
      })
   \u0020`,
        "expect(() => { throw new Error('a'); }).not.toThrow();",
        "expect(() => { throw new Error('a'); }).not.toThrowError();",
        `
      test('Allow no message for "not"', async () => {
        const throwErrorAsync = async () => { throw new Error('a') };
        await expect(throwErrorAsync()).resolves.not.toThrow();
        await expect(throwErrorAsync()).resolves.not.toThrowError();
      })
   \u0020`,
        'expect(a);',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    ...(
      [
        "expect(() => { throw new Error('a'); }).toThrow('a');",
        "expect(() => { throw new Error('a'); }).toThrowError('a');",
        `test('string', async () => {
  const throwErrorAsync = async () => { throw new Error('a') };
  await expect(throwErrorAsync()).rejects.toThrow('a');
  await expect(throwErrorAsync()).rejects.toThrowError('a');
})`,
        "const a = 'a'; expect(() => { throw new Error('a'); }).toThrow(`${a}`);",
        "const a = 'a'; expect(() => { throw new Error('a'); }).toThrowError(`${a}`);",
        `test('Template literal', async () => {
  const a = 'a';
  const throwErrorAsync = async () => { throw new Error('a') };
  await expect(throwErrorAsync()).rejects.toThrow(\`\${a}\`);
  await expect(throwErrorAsync()).rejects.toThrowError(\`\${a}\`);
})`,
        "expect(() => { throw new Error('a'); }).toThrow(/^a$/);",
        "expect(() => { throw new Error('a'); }).toThrowError(/^a$/);",
        `test('Regex', async () => {
  const throwErrorAsync = async () => { throw new Error('a') };
  await expect(throwErrorAsync()).rejects.toThrow(/^a$/);
  await expect(throwErrorAsync()).rejects.toThrowError(/^a$/);
})`,
        "expect(() => { throw new Error('a'); }).toThrow((() => { return 'a'; })());",
        "expect(() => { throw new Error('a'); }).toThrowError((() => { return 'a'; })());",
        `test('Function', async () => {
  const throwErrorAsync = async () => { throw new Error('a') };
  const fn = () => { return 'a'; };
  await expect(throwErrorAsync()).rejects.toThrow(fn());
  await expect(throwErrorAsync()).rejects.toThrowError(fn());
})`,
        "expect(() => { throw new Error('a'); }).not.toThrow();",
        "expect(() => { throw new Error('a'); }).not.toThrowError();",
        `test('Allow no message for "not"', async () => {
  const throwErrorAsync = async () => { throw new Error('a') };
  await expect(throwErrorAsync()).resolves.not.toThrow();
  await expect(throwErrorAsync()).resolves.not.toThrowError();
})`,
        'expect(a);',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream require-to-throw-message / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
    })),
  ],
  invalid: [
    ...(
      [
        [
          "expect(() => { throw new Error('a'); }).toThrow();",
          'Add an error message to toThrow()',
          41,
          48,
        ],
        [
          "(expect(() => { throw new Error('a'); }).toThrow)();",
          'Add an error message to toThrow()',
          42,
          49,
        ],
        [
          "expect(() => { throw new Error('a'); }).toThrowError();",
          'Add an error message to toThrowError()',
          41,
          53,
        ],
      ] as const
    ).map(([code, message, column, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      errors: [
        {
          messageId: 'addErrorMessage',
          message,
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    {
      name: 'run 0 invalid 3',
      code: `test('empty rejects.toThrow', async () => {
  const throwErrorAsync = async () => { throw new Error('a') };
  await expect(throwErrorAsync()).rejects.toThrow();
  await expect(throwErrorAsync()).rejects.toThrowError();
})`,
      errors: [
        {
          messageId: 'addErrorMessage',
          message: 'Add an error message to toThrow()',
          line: 3,
          column: 43,
          endLine: 3,
          endColumn: 50,
        },
        {
          messageId: 'addErrorMessage',
          message: 'Add an error message to toThrowError()',
          line: 4,
          column: 43,
          endLine: 4,
          endColumn: 55,
        },
      ],
    },
    ...(
      [
        [
          "expect(() => { throw new Error('a'); }).toThrow();",
          'Add an error message to toThrow()',
          48,
        ],
        [
          "expect(() => { throw new Error('a'); }).toThrowError();",
          'Add an error message to toThrowError()',
          53,
        ],
      ] as const
    ).map(([code, message, endColumn], index): TestCase => ({
      name: `upstream require-to-throw-message / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'addErrorMessage',
          message,
          line: 1,
          column: 41,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    {
      name: 'upstream require-to-throw-message / invalid 2',
      code: `test('empty rejects.toThrow', async () => {
  const throwErrorAsync = async () => { throw new Error('a') };
  await expect(throwErrorAsync()).rejects.toThrow();
  await expect(throwErrorAsync()).rejects.toThrowError();
})`,
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'addErrorMessage',
          message: 'Add an error message to toThrow()',
          line: 3,
          column: 43,
          endLine: 3,
          endColumn: 50,
        },
        {
          messageId: 'addErrorMessage',
          message: 'Add an error message to toThrowError()',
          line: 4,
          column: 43,
          endLine: 4,
          endColumn: 55,
        },
      ],
    },
  ],
});

// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('prefer-mock-promise-shorthand', {
  valid: [
    ...(
      [
        'jest.fn().mockResolvedValue(42)',
        'jest.fn(() => Promise.resolve(42))',
        'aVariable.mockImplementation()',
        'aVariable.mockReturnValue(Promise.all([1, 2, 3]))',
        "jest.spyOn(Thingy, 'method').mockImplementation(param => Promise.resolve(param));",
        `let value = 1;
aVariable.mockImplementation(() => Promise.resolve(value));`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    ...(
      [
        'describe()',
        'it()',
        'describe.skip()',
        'it.skip()',
        'test()',
        'test.skip()',
        'var appliedOnly = describe.only; appliedOnly.apply(describe)',
        'var calledOnly = it.only; calledOnly.call(it)',
        'it.each()()',
        'it.each`table`()',
        'test.each()()',
        'test.each`table`()',
        'test.concurrent()',
        'jest.fn().mockResolvedValue(42)',
        'jest.fn(() => Promise.resolve(42))',
        'jest.fn(() => Promise.reject(42))',
        'aVariable.mockImplementation',
        'aVariable.mockImplementation()',
        'aVariable.mockImplementation([])',
        'aVariable.mockImplementation(() => {})',
        'aVariable.mockImplementation(() => [])',
        'aVariable.mockReturnValue(() => Promise.resolve(1))',
        'aVariable.mockReturnValue(Promise.resolve(1).then(() => 1))',
        'aVariable.mockReturnValue(Promise.reject(1).then(() => 1))',
        'aVariable.mockReturnValue(Promise.reject().then(() => 1))',
        'aVariable.mockReturnValue(new Promise(resolve => resolve(1)))',
        'aVariable.mockReturnValue(new Promise((_, reject) => reject(1)))',
        "jest.spyOn(Thingy, 'method').mockImplementation(param => Promise.resolve(param));",
        `aVariable.mockImplementation(() => {
  const value = new Date();

  return Promise.resolve(value);
});`,
        `aVariable.mockImplementation(() => {
  return Promise.resolve(value)
    .then(value => value + 1);
});`,
        `aVariable.mockImplementation(() => {
  return Promise.all([1, 2, 3]);
});`,
        'aVariable.mockImplementation(() => Promise.all([1, 2, 3]));',
        'aVariable.mockReturnValue(Promise.all([1, 2, 3]));',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream prefer-mock-promise-shorthand / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 6,
      },
    })),
  ],
  invalid: [
    ...(
      [
        [
          'jest.fn().mockImplementation(() => Promise.resolve(42))',
          'Prefer mockResolvedValue',
          29,
          'jest.fn().mockResolvedValue(42)',
        ],
        [
          'jest.fn().mockReturnValue(Promise.reject(42))',
          'Prefer mockRejectedValue',
          26,
          'jest.fn().mockRejectedValue(42)',
        ],
        [
          `aVariable.mockImplementationOnce(() => {
  return Promise.reject(42);
})`,
          'Prefer mockRejectedValueOnce',
          33,
          'aVariable.mockRejectedValueOnce(42)',
        ],
      ] as const
    ).map(([code, message, endColumn, output], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      errors: [
        {
          messageId: 'useMockShorthand',
          message,
          line: 1,
          column: 11,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
    {
      name: 'run 0 invalid 3',
      code: 'aVariable.mockReturnValueOnce(Promise.resolve(42, xyz))',
      errors: [
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockResolvedValueOnce',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      name: 'run 0 invalid 4',
      code: 'aVariable.mockReturnValueOnce(Promise.resolve())',
      errors: [
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockResolvedValueOnce',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 30,
        },
      ],
      output: 'aVariable.mockResolvedValueOnce(undefined)',
    },
    ...(
      [
        [
          'jest.fn().mockImplementation(() => Promise.resolve(42))',
          'Prefer mockResolvedValue',
          'jest.fn().mockResolvedValue(42)',
        ],
        [
          'jest.fn().mockImplementation(() => Promise.reject(42))',
          'Prefer mockRejectedValue',
          'jest.fn().mockRejectedValue(42)',
        ],
        [
          'aVariable.mockImplementation(() => Promise.resolve(42))',
          'Prefer mockResolvedValue',
          'aVariable.mockResolvedValue(42)',
        ],
        [
          `aVariable.mockImplementation(() => {
  return Promise.resolve(42)
})`,
          'Prefer mockResolvedValue',
          'aVariable.mockResolvedValue(42)',
        ],
        [
          'aVariable.mockImplementation(() => Promise.reject(42))',
          'Prefer mockRejectedValue',
          'aVariable.mockRejectedValue(42)',
        ],
      ] as const
    ).map(([code, message, output], index): TestCase => ({
      name: `upstream prefer-mock-promise-shorthand / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'useMockShorthand',
          message,
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 29,
        },
      ],
      output,
    })),
    {
      name: 'upstream prefer-mock-promise-shorthand / invalid 5',
      code: 'aVariable.mockImplementation(() => Promise.reject(42),)',
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockRejectedValue',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 29,
        },
      ],
      output: 'aVariable.mockRejectedValue(42,)',
    },
    ...(
      [
        [
          'aVariable.mockImplementationOnce(() => Promise.resolve(42))',
          'Prefer mockResolvedValueOnce',
          33,
          'aVariable.mockResolvedValueOnce(42)',
        ],
        [
          'aVariable.mockImplementationOnce(() => Promise.reject(42))',
          'Prefer mockRejectedValueOnce',
          33,
          'aVariable.mockRejectedValueOnce(42)',
        ],
        [
          'jest.fn().mockReturnValue(Promise.resolve(42))',
          'Prefer mockResolvedValue',
          26,
          'jest.fn().mockResolvedValue(42)',
        ],
        [
          'jest.fn().mockReturnValue(Promise.reject(42))',
          'Prefer mockRejectedValue',
          26,
          'jest.fn().mockRejectedValue(42)',
        ],
        [
          'aVariable.mockReturnValue(Promise.resolve(42))',
          'Prefer mockResolvedValue',
          26,
          'aVariable.mockResolvedValue(42)',
        ],
        [
          'aVariable.mockReturnValue(Promise.reject(42))',
          'Prefer mockRejectedValue',
          26,
          'aVariable.mockRejectedValue(42)',
        ],
        [
          'aVariable.mockReturnValueOnce(Promise.resolve(42))',
          'Prefer mockResolvedValueOnce',
          30,
          'aVariable.mockResolvedValueOnce(42)',
        ],
        [
          'aVariable.mockReturnValueOnce(Promise.reject(42))',
          'Prefer mockRejectedValueOnce',
          30,
          'aVariable.mockRejectedValueOnce(42)',
        ],
      ] as const
    ).map(([code, message, endColumn, output], index): TestCase => ({
      name: `upstream prefer-mock-promise-shorthand / invalid ${index + 6}`,
      code,
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'useMockShorthand',
          message,
          line: 1,
          column: 11,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
    {
      name: 'upstream prefer-mock-promise-shorthand / invalid 14',
      code: `aVariable.mockReturnValue(Promise.resolve({
  target: 'world',
  message: 'hello'
}))`,
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockResolvedValue',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    ...(
      [
        [
          `aVariable
  .mockImplementation(() => Promise.reject(42))
  .mockImplementation(() => Promise.resolve(42))
  .mockReturnValue(Promise.reject(42))`,
          'Prefer mockRejectedValue',
          22,
          'Prefer mockRejectedValue',
          19,
          `aVariable
  .mockRejectedValue(42)
  .mockResolvedValue(42)
  .mockRejectedValue(42)`,
        ],
        [
          `aVariable
  .mockReturnValueOnce(Promise.reject(42))
  .mockImplementation(() => Promise.resolve(42))
  .mockReturnValueOnce(Promise.reject(42))`,
          'Prefer mockRejectedValueOnce',
          23,
          'Prefer mockRejectedValueOnce',
          23,
          `aVariable
  .mockRejectedValueOnce(42)
  .mockResolvedValue(42)
  .mockRejectedValueOnce(42)`,
        ],
      ] as const
    ).map(
      (
        [code, message1, endColumn1, message3, endColumn3, output],
        index,
      ): TestCase => ({
        name: `upstream prefer-mock-promise-shorthand / invalid ${index + 15}`,
        code,
        languageOptions: {
          ecmaVersion: 6,
        },
        errors: [
          {
            messageId: 'useMockShorthand',
            message: message1,
            line: 2,
            column: 4,
            endLine: 2,
            endColumn: endColumn1,
          },
          {
            messageId: 'useMockShorthand',
            message: 'Prefer mockResolvedValue',
            line: 3,
            column: 4,
            endLine: 3,
            endColumn: 22,
          },
          {
            messageId: 'useMockShorthand',
            message: message3,
            line: 4,
            column: 4,
            endLine: 4,
            endColumn: endColumn3,
          },
        ],
        output,
      }),
    ),
    ...(
      [
        [
          `aVariable.mockReturnValueOnce(
  Promise.reject(
    new Error('oh noes!')
  )
)`,
          'Prefer mockRejectedValueOnce',
          30,
          `aVariable.mockRejectedValueOnce(
  new Error('oh noes!')
)`,
        ],
        [
          'jest.fn().mockReturnValue(Promise.resolve(42), xyz)',
          'Prefer mockResolvedValue',
          26,
          'jest.fn().mockResolvedValue(42, xyz)',
        ],
        [
          'jest.fn().mockImplementation(() => Promise.reject(42), xyz)',
          'Prefer mockRejectedValue',
          29,
          'jest.fn().mockRejectedValue(42, xyz)',
        ],
      ] as const
    ).map(([code, message, endColumn, output], index): TestCase => ({
      name: `upstream prefer-mock-promise-shorthand / invalid ${index + 17}`,
      code,
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'useMockShorthand',
          message,
          line: 1,
          column: 11,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
    {
      name: 'upstream prefer-mock-promise-shorthand / invalid 20',
      code: 'aVariable.mockReturnValueOnce(Promise.resolve(42, xyz))',
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockResolvedValueOnce',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    ...(
      [
        [
          'aVariable.mockReturnValueOnce(Promise.resolve())',
          'Prefer mockResolvedValueOnce',
          11,
          30,
          'aVariable.mockResolvedValueOnce(undefined)',
        ],
        [
          'jest.spyOn(fs, "readFile").mockReturnValue(Promise.reject(new Error("oh noes!")))',
          'Prefer mockRejectedValue',
          28,
          43,
          'jest.spyOn(fs, "readFile").mockRejectedValue(new Error("oh noes!"))',
        ],
      ] as const
    ).map(([code, message, column, endColumn, output], index): TestCase => ({
      name: `upstream prefer-mock-promise-shorthand / invalid ${index + 21}`,
      code,
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'useMockShorthand',
          message,
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
  ],
});

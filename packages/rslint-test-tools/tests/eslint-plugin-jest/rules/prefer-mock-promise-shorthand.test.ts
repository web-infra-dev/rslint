import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-mock-promise-shorthand', {} as never, {
  valid: [
    { code: `jest.fn().mockResolvedValue(42)` },
    { code: `jest.fn(() => Promise.resolve(42))` },
    { code: `aVariable.mockImplementation()` },
    { code: `aVariable.mockReturnValue(Promise.all([1, 2, 3]))` },
    {
      code: `jest.spyOn(Thingy, 'method').mockImplementation(param => Promise.resolve(param));`,
    },
    {
      code: `let value = 1;
aVariable.mockImplementation(() => Promise.resolve(value));`,
    },
  ],

  invalid: [
    {
      code: `jest.fn().mockImplementation(() => Promise.resolve(42))`,
      output: `jest.fn().mockResolvedValue(42)`,
      errors: [{ messageId: 'useMockShorthand', column: 11, line: 1 }],
    },
    {
      code: `jest.fn().mockReturnValue(Promise.reject(42))`,
      output: `jest.fn().mockRejectedValue(42)`,
      errors: [{ messageId: 'useMockShorthand', column: 11, line: 1 }],
    },
    {
      code: `aVariable.mockImplementationOnce(() => {
  return Promise.reject(42);
})`,
      output: `aVariable.mockRejectedValueOnce(42)`,
      errors: [{ messageId: 'useMockShorthand', column: 11, line: 1 }],
    },
    {
      code: `aVariable.mockReturnValueOnce(Promise.resolve(42, xyz))`,
      output: null,
      errors: [{ messageId: 'useMockShorthand', column: 11, line: 1 }],
    },
    {
      code: `aVariable.mockReturnValueOnce(Promise.resolve())`,
      output: `aVariable.mockResolvedValueOnce(undefined)`,
      errors: [{ messageId: 'useMockShorthand', column: 11, line: 1 }],
    },
  ],
});

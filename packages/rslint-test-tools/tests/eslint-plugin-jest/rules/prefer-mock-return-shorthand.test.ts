import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-mock-return-shorthand', {} as never, {
  valid: [
    { code: `jest.fn().mockReturnValue(42)` },
    { code: `jest.fn(() => 42)` },
    { code: `aVariable.mockImplementation()` },
    { code: `jest.fn().mockImplementation(async () => 1);` },
    { code: `aVariable.mockImplementation(() => value++)` },
    {
      code: `jest.spyOn(Thingy, 'method').mockImplementation(param => param * 2);`,
    },
    {
      code: `let value = 1;

aVariable.mockImplementation(() => ({ value }));`,
    },
    { code: `jest.fn().mockImplementation(() => Promise.reject(13))` },
  ],

  invalid: [
    {
      code: `jest.fn().mockImplementation(() => "hello sunshine")`,
      output: `jest.fn().mockReturnValue("hello sunshine")`,
      errors: [{ messageId: 'useMockShorthand', column: 11, line: 1 }],
    },
    {
      code: `jest.fn().mockImplementationOnce(() => "hello world")`,
      output: `jest.fn().mockReturnValueOnce("hello world")`,
      errors: [{ messageId: 'useMockShorthand', column: 11, line: 1 }],
    },
    {
      code: `aVariable.mockImplementation(() => {
  return "hello world";
})`,
      output: `aVariable.mockReturnValue("hello world")`,
      errors: [{ messageId: 'useMockShorthand', column: 11, line: 1 }],
    },
    {
      code: `const currentX = 0;
jest.spyOn(X, getCount).mockImplementation(() => currentX);`,
      output: `const currentX = 0;
jest.spyOn(X, getCount).mockReturnValue(currentX);`,
      errors: [{ messageId: 'useMockShorthand', column: 25, line: 2 }],
    },
  ],
});

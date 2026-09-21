import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-mock-return-shorthand', {} as never, {
  valid: [
    { code: `rs.fn().mockReturnValue(42)` },
    { code: `rs.fn(() => 42)` },
    { code: `aVariable.mockImplementation()` },
    { code: `rs.fn().mockImplementation(async () => 1);` },
    { code: `aVariable.mockImplementation(() => value++)` },
    {
      code: `rs.spyOn(Thingy, 'method').mockImplementation(param => param * 2);`,
    },
    {
      code: `aVariable.mockImplementation(() => {
  throw new Error('oh noes!');
});`,
    },
    {
      code: `let currentX = 0;
rs.spyOn(X, getCount).mockImplementation(() => currentX);`,
    },
    { code: `rs.fn().mockImplementation(() => Promise.reject(13))` },
  ],

  invalid: [
    {
      code: `rs.fn().mockImplementation(() => "hello sunshine")`,
      output: `rs.fn().mockReturnValue("hello sunshine")`,
      errors: [{ messageId: 'useMockShorthand', column: 9, line: 1 }],
    },
    {
      code: `rs.fn().mockImplementationOnce(() => "hello world")`,
      output: `rs.fn().mockReturnValueOnce("hello world")`,
      errors: [{ messageId: 'useMockShorthand', column: 9, line: 1 }],
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
rs.spyOn(X, getCount).mockImplementation(() => currentX);`,
      output: `const currentX = 0;
rs.spyOn(X, getCount).mockReturnValue(currentX);`,
      errors: [{ messageId: 'useMockShorthand', column: 23, line: 2 }],
    },
  ],
});

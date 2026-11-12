import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('prefer-mock-return-shorthand', {
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
      errors: [
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: `rs.fn().mockImplementationOnce(() => "hello world")`,
      output: `rs.fn().mockReturnValueOnce("hello world")`,
      errors: [
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValueOnce',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: `aVariable.mockImplementation(() => {
  return "hello world";
})`,
      output: `aVariable.mockReturnValue("hello world")`,
      errors: [
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: `const currentX = 0;
rs.spyOn(X, getCount).mockImplementation(() => currentX);`,
      output: `const currentX = 0;
rs.spyOn(X, getCount).mockReturnValue(currentX);`,
      errors: [
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockReturnValue',
          line: 2,
          column: 23,
          endLine: 2,
          endColumn: 41,
        },
      ],
    },
  ],
});

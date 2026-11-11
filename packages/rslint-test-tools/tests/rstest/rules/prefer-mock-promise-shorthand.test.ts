import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('prefer-mock-promise-shorthand', {
  valid: [
    { code: `rs.fn().mockResolvedValue(42)` },
    { code: `rs.fn(() => Promise.resolve(42))` },
    { code: `aVariable.mockImplementation()` },
    { code: `aVariable.mockReturnValue(Promise.all([1, 2, 3]))` },
    {
      code: `rs.spyOn(Thingy, 'method').mockImplementation(param => Promise.resolve(param));`,
    },
    {
      code: `let value = 1;
aVariable.mockImplementation(() => Promise.resolve(value));`,
    },
  ],

  invalid: [
    {
      code: `rs.fn().mockImplementation(() => Promise.resolve(42))`,
      output: `rs.fn().mockResolvedValue(42)`,
      errors: [
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockResolvedValue',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: `rs.fn().mockReturnValue(Promise.reject(42))`,
      output: `rs.fn().mockRejectedValue(42)`,
      errors: [
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockRejectedValue',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: `aVariable.mockImplementationOnce(() => {
  return Promise.reject(42);
})`,
      output: `aVariable.mockRejectedValueOnce(42)`,
      errors: [
        {
          messageId: 'useMockShorthand',
          message: 'Prefer mockRejectedValueOnce',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: `aVariable.mockReturnValueOnce(Promise.resolve(42, xyz))`,
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
      code: `aVariable.mockReturnValueOnce(Promise.resolve())`,
      output: `aVariable.mockResolvedValueOnce(undefined)`,
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
  ],
});

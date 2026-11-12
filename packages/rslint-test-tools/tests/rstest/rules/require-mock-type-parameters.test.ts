import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('require-mock-type-parameters', {
  valid: [
    { code: `rs.fn<(...args: any[]) => any>()` },
    { code: `rs.fn<(arg1: string, arg2: boolean) => string>()` },
    { code: `rs.fn<MyProcedure>()` },
    { code: `rs.fn<any>()` },
    { code: `rs.fn<(...args: any[]) => any>(() => {})` },
    {
      code: `rs.fn<() => string | undefined>().mockReturnValue('some error message')`,
    },
    { code: `rs.importActual<{ default: boolean }>('./example.js')` },
    { code: `rs.importMock<MyModule>('./example.js')` },
    { code: `rs.importActual('./example.js')` },
    { code: `rs.importMock('./example.js')` },
  ],
  invalid: [
    {
      code: `rs.fn()`,
      errors: [
        {
          messageId: 'missingTypeParameter',
          message: "'fn' is called without a type parameter.",
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 6,
        },
      ],
    },
    {
      code: `rs.fn(() => {})`,
      errors: [
        {
          messageId: 'missingTypeParameter',
          message: "'fn' is called without a type parameter.",
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 6,
        },
      ],
    },
    {
      code: `rs.importActual('./example.js')`,
      options: [{ checkImportFunctions: true }],
      errors: [
        {
          messageId: 'missingTypeParameter',
          message: "'importActual' is called without a type parameter.",
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: `rs.importMock('./example.js')`,
      options: [{ checkImportFunctions: true }],
      errors: [
        {
          messageId: 'missingTypeParameter',
          message: "'importMock' is called without a type parameter.",
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
  ],
});

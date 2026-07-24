import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-mocks-import', {} as never, {
  valid: [
    { code: 'import something from "something"' },
    { code: 'require("somethingElse")' },
    { code: 'require("./__mocks__.js")' },
    { code: 'require("./__mocks__x")' },
    { code: 'require("./__mocks__x/x")' },
    { code: 'require("./x__mocks__")' },
    { code: 'require("./x__mocks__/x")' },
    { code: 'require()' },
    { code: 'var path = "./__mocks__.js"; require(path)' },
    { code: 'entirelyDifferent(fn)' },
  ],
  invalid: [
    {
      code: 'require("./__mocks__")',
      errors: [
        {
          line: 1,
          column: 9,
          endColumn: 22,
          messageId: 'noManualImport',
          message:
            'Mocks should not be manually imported from a __mocks__ directory. Instead use `rs.mock` and import from the original module path',
        },
      ],
    },
    {
      code: 'require("./__mocks__/")',
      errors: [
        { line: 1, column: 9, endColumn: 23, messageId: 'noManualImport' },
      ],
    },
    {
      code: 'require("./__mocks__/index")',
      errors: [
        { line: 1, column: 9, endColumn: 28, messageId: 'noManualImport' },
      ],
    },
    {
      code: 'require("__mocks__")',
      errors: [
        { line: 1, column: 9, endColumn: 20, messageId: 'noManualImport' },
      ],
    },
    {
      code: 'require("__mocks__/")',
      errors: [
        { line: 1, column: 9, endColumn: 21, messageId: 'noManualImport' },
      ],
    },
    {
      code: 'require("__mocks__/index")',
      errors: [
        { line: 1, column: 9, endColumn: 26, messageId: 'noManualImport' },
      ],
    },
    {
      code: 'import thing from "./__mocks__/index"',
      errors: [
        { line: 1, column: 1, endColumn: 38, messageId: 'noManualImport' },
      ],
    },
  ],
});

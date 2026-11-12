import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('no-mocks-import', {
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
          messageId: 'noManualImport',
          message:
            'Mocks should not be manually imported from a __mocks__ directory. Instead use `rs.mock` and import from the original module path',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'require("./__mocks__/")',
      errors: [
        {
          messageId: 'noManualImport',
          message:
            'Mocks should not be manually imported from a __mocks__ directory. Instead use `rs.mock` and import from the original module path',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'require("./__mocks__/index")',
      errors: [
        {
          messageId: 'noManualImport',
          message:
            'Mocks should not be manually imported from a __mocks__ directory. Instead use `rs.mock` and import from the original module path',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: 'require("__mocks__")',
      errors: [
        {
          messageId: 'noManualImport',
          message:
            'Mocks should not be manually imported from a __mocks__ directory. Instead use `rs.mock` and import from the original module path',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: 'require("__mocks__/")',
      errors: [
        {
          messageId: 'noManualImport',
          message:
            'Mocks should not be manually imported from a __mocks__ directory. Instead use `rs.mock` and import from the original module path',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: 'require("__mocks__/index")',
      errors: [
        {
          messageId: 'noManualImport',
          message:
            'Mocks should not be manually imported from a __mocks__ directory. Instead use `rs.mock` and import from the original module path',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: 'import thing from "./__mocks__/index"',
      errors: [
        {
          messageId: 'noManualImport',
          message:
            'Mocks should not be manually imported from a __mocks__ directory. Instead use `rs.mock` and import from the original module path',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
  ],
});

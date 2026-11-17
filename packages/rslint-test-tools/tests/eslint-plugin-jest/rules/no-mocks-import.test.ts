// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('no-mocks-import', {
  valid: [
    ...(
      [
        'import something from "something"',
        'require("somethingElse")',
        'require("./__mocks__.js")',
        'require("./__mocks__x")',
        'require("./__mocks__x/x")',
        'require("./x__mocks__")',
        'require("./x__mocks__/x")',
        'require()',
        'var path = "./__mocks__.js"; require(path)',
        'entirelyDifferent(fn)',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    {
      name: 'upstream no-mocks-import / valid 0',
      code: 'import something from "something"',
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2015,
      },
    },
    ...(
      [
        'require("somethingElse")',
        'require("./__mocks__.js")',
        'require("./__mocks__x")',
        'require("./__mocks__x/x")',
        'require("./x__mocks__")',
        'require("./x__mocks__/x")',
        'require()',
        'var path = "./__mocks__.js"; require(path)',
        'entirelyDifferent(fn)',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream no-mocks-import / valid ${index + 1}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
    })),
  ],
  invalid: [
    ...(
      [
        ['require("./__mocks__")', 9, 22],
        ['require("./__mocks__/")', 9, 23],
        ['require("./__mocks__/index")', 9, 28],
        ['require("__mocks__")', 9, 20],
        ['require("__mocks__/")', 9, 21],
        ['require("__mocks__/index")', 9, 26],
        ['import thing from "./__mocks__/index"', 1, 38],
      ] as const
    ).map(([code, column, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      errors: [
        {
          messageId: 'noManualImport',
          message:
            'Mocks should not be manually imported from a __mocks__ directory. Instead use `jest.mock` and import from the original module path',
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    ...(
      [
        ['require("./__mocks__")', 22],
        ['require("./__mocks__/")', 23],
        ['require("./__mocks__/index")', 28],
        ['require("__mocks__")', 20],
        ['require("__mocks__/")', 21],
        ['require("__mocks__/index")', 26],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `upstream no-mocks-import / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'noManualImport',
          message:
            'Mocks should not be manually imported from a __mocks__ directory. Instead use `jest.mock` and import from the original module path',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    {
      name: 'upstream no-mocks-import / invalid 6',
      code: 'import thing from "./__mocks__/index"',
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'noManualImport',
          message:
            'Mocks should not be manually imported from a __mocks__ directory. Instead use `jest.mock` and import from the original module path',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
  ],
});

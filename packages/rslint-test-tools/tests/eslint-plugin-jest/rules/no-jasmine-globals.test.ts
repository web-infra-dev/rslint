// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('no-jasmine-globals', {
  valid: [
    ...(
      [
        'jest.spyOn()',
        'jest.fn()',
        'expect.extend()',
        'expect.any()',
        'it("foo", function () {})',
        'test("foo", function () {})',
        'foo()',
        "require('foo')('bar')",
        '(function(){})()',
        'function callback(fail) { fail() }',
        'var spyOn = require("actions"); spyOn("foo")',
        'function callback(pending) { pending() }',
        'function callback(jasmine) { jasmine.any() }',
        'const jasmine = foo;',
        '(this as any).jasmine.foo()',
        '(this as any).jasmine.any()',
        'jasmine[1].any()',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
  ],
  invalid: [
    ...(
      [
        [
          'spyOn(some, "object")',
          'Illegal usage of global `spyOn`, prefer `jest.spyOn`',
          22,
        ],
        [
          'spyOnProperty(some, "object")',
          'Illegal usage of global `spyOnProperty`, prefer `jest.spyOn`',
          30,
        ],
      ] as const
    ).map(([code, message, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      errors: [
        {
          messageId: 'illegalGlobal',
          message,
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    {
      name: 'run 0 invalid 2',
      code: 'fail()',
      errors: [
        {
          messageId: 'illegalFail',
          message:
            'Illegal usage of `fail`, prefer throwing an error, or the `done.fail` callback',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 7,
        },
      ],
    },
    {
      name: 'run 0 invalid 3',
      code: 'pending()',
      errors: [
        {
          messageId: 'illegalPending',
          message:
            'Illegal usage of `pending`, prefer explicitly skipping a test using `test.skip`',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    ...(
      [
        ['jasmine', 1],
        ['const value = jasmine;', 15],
      ] as const
    ).map(([code, column], index): TestCase => ({
      name: `run 0 invalid ${index + 4}`,
      code,
      errors: [
        {
          messageId: 'illegalJasmine',
          message: 'Illegal usage of jasmine global',
          line: 1,
          column,
          endLine: 1,
          endColumn: column + 7,
        },
      ],
    })),
    {
      name: 'run 0 invalid 6',
      code: 'jasmine.DEFAULT_TIMEOUT_INTERVAL = 5000;',
      errors: [
        {
          messageId: 'illegalJasmine',
          message: 'Illegal usage of jasmine global',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
        },
      ],
      output: 'jest.setTimeout(5000);',
    },
    ...(
      [
        'jasmine.DEFAULT_TIMEOUT_INTERVAL = function() {}',
        'jasmine.DEFAULT_TIMEOUT_INTERVAL += 1000;',
      ] as const
    ).map((code, index): TestCase => ({
      name: `run 0 invalid ${index + 7}`,
      code,
      errors: [
        {
          messageId: 'illegalJasmine',
          message: 'Illegal usage of jasmine global',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
        },
      ],
    })),
    {
      name: 'run 0 invalid 9',
      code: 'jasmine["DEFAULT_TIMEOUT_INTERVAL"] = 5000;',
      errors: [
        {
          messageId: 'illegalJasmine',
          message: 'Illegal usage of jasmine global',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 36,
        },
      ],
      output: 'jest.setTimeout(5000);',
    },
    ...(
      [
        'jasmine["DEFAULT_TIMEOUT_INTERVAL"] = function() {}',
        'jasmine["DEFAULT_TIMEOUT_INTERVAL"] += 1000;',
      ] as const
    ).map((code, index): TestCase => ({
      name: `run 0 invalid ${index + 10}`,
      code,
      errors: [
        {
          messageId: 'illegalJasmine',
          message: 'Illegal usage of jasmine global',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 36,
        },
      ],
    })),
    ...(
      [
        [
          'jasmine.addMatchers(matchers)',
          'Illegal usage of `jasmine.addMatchers`, prefer `expect.extend`',
          30,
        ],
        [
          'jasmine.createSpy()',
          'Illegal usage of `jasmine.createSpy`, prefer `jest.fn`',
          20,
        ],
      ] as const
    ).map(([code, message, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index + 12}`,
      code,
      errors: [
        {
          messageId: 'illegalMethod',
          message,
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
    })),
    ...(
      [
        [
          'jasmine.any()',
          'Illegal usage of `jasmine.any`, prefer `expect.any`',
          14,
          'expect.any()',
        ],
        [
          'jasmine.anything()',
          'Illegal usage of `jasmine.anything`, prefer `expect.anything`',
          19,
          'expect.anything()',
        ],
        [
          'jasmine.arrayContaining()',
          'Illegal usage of `jasmine.arrayContaining`, prefer `expect.arrayContaining`',
          26,
          'expect.arrayContaining()',
        ],
        [
          'jasmine.objectContaining()',
          'Illegal usage of `jasmine.objectContaining`, prefer `expect.objectContaining`',
          27,
          'expect.objectContaining()',
        ],
        [
          'jasmine.stringMatching()',
          'Illegal usage of `jasmine.stringMatching`, prefer `expect.stringMatching`',
          25,
          'expect.stringMatching()',
        ],
      ] as const
    ).map(([code, message, endColumn, output], index): TestCase => ({
      name: `run 0 invalid ${index + 14}`,
      code,
      errors: [
        {
          messageId: 'illegalMethod',
          message,
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
    ...(
      [
        ['jasmine.foo.any()', 1, 18],
        ['jasmine.foo.addMatchers()', 1, 26],
        ['jasmine.foo.createSpy()', 1, 24],
        ['console.log(jasmine.version)', 13, 28],
        ['jasmine.getEnv()', 1, 17],
        ['jasmine.empty()', 1, 16],
        ['jasmine.falsy()', 1, 16],
        ['jasmine.truthy()', 1, 17],
        ['jasmine.arrayWithExactContents()', 1, 33],
        ['jasmine.clock()', 1, 16],
        ['jasmine.MAX_PRETTY_PRINT_ARRAY_LENGTH = 42', 1, 38],
        ['jasmine["MAX_PRETTY_PRINT_ARRAY_LENGTH"] = 42', 1, 41],
      ] as const
    ).map(([code, column, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index + 19}`,
      code,
      errors: [
        {
          messageId: 'illegalJasmine',
          message: 'Illegal usage of jasmine global',
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
    })),
  ],
});

// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('prefer-todo', {
  valid: [
    ...(
      [
        'test()',
        'test.concurrent()',
        'test.todo("i need to write this test");',
        'test(obj)',
        'test.concurrent(obj)',
        'fit("foo")',
        'fit.concurrent("foo")',
        'xit("foo")',
        'test("foo", 1)',
        'test("stub", () => expect(1).toBe(1));',
        'test.concurrent("stub", () => expect(1).toBe(1));',
        `
        supportsDone && params.length < test.length
          ? done => test(...params, done)
          : () => test(...params);
     \u0020`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    ...(
      [
        'test()',
        'test.concurrent()',
        'test.todo("i need to write this test");',
        'test(obj)',
        'test.concurrent(obj)',
        'fit("foo")',
        'fit.concurrent("foo")',
        'xit("foo")',
        'test("foo", 1)',
        'test("stub", () => expect(1).toBe(1));',
        'test.concurrent("stub", () => expect(1).toBe(1));',
        `supportsDone && params.length < test.length
  ? done => test(...params, done)
  : () => test(...params);`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream prefer-todo / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
    })),
  ],
  invalid: [
    ...(
      [
        [
          'test("i need to write this test");',
          34,
          'test.todo("i need to write this test");',
        ],
        [
          'test("i need to write this test",);',
          35,
          'test.todo("i need to write this test",);',
        ],
        [
          'test(`i need to write this test`);',
          34,
          'test.todo(`i need to write this test`);',
        ],
      ] as const
    ).map(([code, endColumn, output], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      errors: [
        {
          messageId: 'unimplementedTest',
          message: 'Prefer todo test case over unimplemented test case',
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
        ['it("foo", function () {})', 26, 'it.todo("foo")'],
        ['it("foo", () => {})', 20, 'it.todo("foo")'],
        [
          'test.skip("i need to write this test", () => {});',
          49,
          'test.todo("i need to write this test");',
        ],
        [
          'test.skip("i need to write this test", function() {});',
          54,
          'test.todo("i need to write this test");',
        ],
        [
          'test["skip"]("i need to write this test", function() {});',
          57,
          'test[\'todo\']("i need to write this test");',
        ],
        [
          'test[`skip`]("i need to write this test", function() {});',
          57,
          'test[\'todo\']("i need to write this test");',
        ],
      ] as const
    ).map(([code, endColumn, output], index): TestCase => ({
      name: `run 0 invalid ${index + 3}`,
      code,
      errors: [
        {
          messageId: 'emptyTest',
          message: 'Prefer todo test case over empty test case',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
    {
      name: 'upstream prefer-todo / invalid 0',
      code: 'test("i need to write this test");',
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'unimplementedTest',
          message: 'Prefer todo test case over unimplemented test case',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
        },
      ],
      output: 'test.todo("i need to write this test");',
    },
    {
      name: 'upstream prefer-todo / invalid 1',
      code: 'test("i need to write this test",);',
      languageOptions: {
        ecmaVersion: 2017,
      },
      errors: [
        {
          messageId: 'unimplementedTest',
          message: 'Prefer todo test case over unimplemented test case',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 35,
        },
      ],
      output: 'test.todo("i need to write this test",);',
    },
    {
      name: 'upstream prefer-todo / invalid 2',
      code: 'test(`i need to write this test`);',
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'unimplementedTest',
          message: 'Prefer todo test case over unimplemented test case',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
        },
      ],
      output: 'test.todo(`i need to write this test`);',
    },
    ...(
      [
        ['it("foo", function () {})', 26, 'it.todo("foo")'],
        ['it("foo", () => {})', 20, 'it.todo("foo")'],
        [
          'test.skip("i need to write this test", () => {});',
          49,
          'test.todo("i need to write this test");',
        ],
        [
          'test.skip("i need to write this test", function() {});',
          54,
          'test.todo("i need to write this test");',
        ],
        [
          'test["skip"]("i need to write this test", function() {});',
          57,
          'test[\'todo\']("i need to write this test");',
        ],
        [
          'test[`skip`]("i need to write this test", function() {});',
          57,
          'test[\'todo\']("i need to write this test");',
        ],
      ] as const
    ).map(([code, endColumn, output], index): TestCase => ({
      name: `upstream prefer-todo / invalid ${index + 3}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'emptyTest',
          message: 'Prefer todo test case over empty test case',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
  ],
});

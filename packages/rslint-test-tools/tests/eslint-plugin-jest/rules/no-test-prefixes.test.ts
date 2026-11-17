// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('no-test-prefixes', {
  valid: [
    ...(
      [
        'describe("foo", function () {})',
        'it("foo", function () {})',
        'it.concurrent("foo", function () {})',
        'test("foo", function () {})',
        'test.concurrent("foo", function () {})',
        'describe.only("foo", function () {})',
        'it.only("foo", function () {})',
        'it.each()("foo", function () {})',
        'it.each``("foo", function () {})',
        'test.only("foo", function () {})',
        'test.each()("foo", function () {})',
        'test.each``("foo", function () {})',
        'describe.skip("foo", function () {})',
        'it.skip("foo", function () {})',
        'test.skip("foo", function () {})',
        'foo()',
        '[1,2,3].forEach()',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    ...(
      [
        [
          'it.each``("foo", function () {})',
          'upstream no-test-prefixes / valid 8',
        ],
        [
          'test.each``("foo", function () {})',
          'upstream no-test-prefixes / valid 11',
        ],
      ] as const
    ).map(([code, name]): ValidTestCase => ({
      name,
      code,
      languageOptions: {
        ecmaVersion: 6,
      },
    })),
  ],
  invalid: [
    ...(
      [
        [
          'fdescribe("foo", function () {})',
          'Use "describe.only" instead',
          1,
          1,
          10,
          'describe.only("foo", function () {})',
        ],
        [
          'xdescribe.each([])("foo", function () {})',
          'Use "describe.skip.each" instead',
          1,
          1,
          15,
          'describe.skip.each([])("foo", function () {})',
        ],
        [
          'fit("foo", function () {})',
          'Use "it.only" instead',
          1,
          1,
          4,
          'it.only("foo", function () {})',
        ],
        [
          'xdescribe("foo", function () {})',
          'Use "describe.skip" instead',
          1,
          1,
          10,
          'describe.skip("foo", function () {})',
        ],
        [
          'xit("foo", function () {})',
          'Use "it.skip" instead',
          1,
          1,
          4,
          'it.skip("foo", function () {})',
        ],
        [
          'xtest("foo", function () {})',
          'Use "test.skip" instead',
          1,
          1,
          6,
          'test.skip("foo", function () {})',
        ],
        [
          'xit.each``("foo", function () {})',
          'Use "it.skip.each" instead',
          1,
          1,
          9,
          'it.skip.each``("foo", function () {})',
        ],
        [
          'xtest.each``("foo", function () {})',
          'Use "test.skip.each" instead',
          1,
          1,
          11,
          'test.skip.each``("foo", function () {})',
        ],
        [
          'xit.each([])("foo", function () {})',
          'Use "it.skip.each" instead',
          1,
          1,
          9,
          'it.skip.each([])("foo", function () {})',
        ],
        [
          'xtest.each([])("foo", function () {})',
          'Use "test.skip.each" instead',
          1,
          1,
          11,
          'test.skip.each([])("foo", function () {})',
        ],
        [
          `
        import { xit } from '@jest/globals';

        xit("foo", function () {})
     \u0020`,
          'Use "it.skip" instead',
          4,
          9,
          12,
          `
        import { xit } from '@jest/globals';

        it.skip("foo", function () {})
     \u0020`,
        ],
        [
          `
        import { xit as skipThis } from '@jest/globals';

        skipThis("foo", function () {})
     \u0020`,
          'Use "it.skip" instead',
          4,
          9,
          17,
          `
        import { xit as skipThis } from '@jest/globals';

        it.skip("foo", function () {})
     \u0020`,
        ],
        [
          `
        import { fit as onlyThis } from '@jest/globals';

        onlyThis("foo", function () {})
     \u0020`,
          'Use "it.only" instead',
          4,
          9,
          17,
          `
        import { fit as onlyThis } from '@jest/globals';

        it.only("foo", function () {})
     \u0020`,
        ],
      ] as const
    ).map(
      ([code, message, line, column, endColumn, output], index): TestCase => ({
        name: `run 0 invalid ${index}`,
        code,
        errors: [
          {
            messageId: 'usePreferredName',
            message,
            line,
            column,
            endLine: line,
            endColumn,
          },
        ],
        output,
      }),
    ),
    ...(
      [
        [
          'xit.each``("foo", function () {})',
          'Use "it.skip.each" instead',
          9,
          'it.skip.each``("foo", function () {})',
        ],
        [
          'xtest.each``("foo", function () {})',
          'Use "test.skip.each" instead',
          11,
          'test.skip.each``("foo", function () {})',
        ],
      ] as const
    ).map(([code, message, endColumn, output], index): TestCase => ({
      name: `upstream no-test-prefixes / invalid ${index + 6}`,
      code,
      languageOptions: {
        ecmaVersion: 6,
      },
      errors: [
        {
          messageId: 'usePreferredName',
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
        [
          `import { xit } from '@jest/globals';

xit("foo", function () {})`,
          'Use "it.skip" instead',
          4,
          `import { xit } from '@jest/globals';

it.skip("foo", function () {})`,
        ],
        [
          `import { xit as skipThis } from '@jest/globals';

skipThis("foo", function () {})`,
          'Use "it.skip" instead',
          9,
          `import { xit as skipThis } from '@jest/globals';

it.skip("foo", function () {})`,
        ],
        [
          `import { fit as onlyThis } from '@jest/globals';

onlyThis("foo", function () {})`,
          'Use "it.only" instead',
          9,
          `import { fit as onlyThis } from '@jest/globals';

it.only("foo", function () {})`,
        ],
      ] as const
    ).map(([code, message, endColumn, output], index): TestCase => ({
      name: `upstream no-test-prefixes / invalid ${index + 10}`,
      code,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'usePreferredName',
          message,
          line: 3,
          column: 1,
          endLine: 3,
          endColumn,
        },
      ],
      output,
    })),
  ],
});

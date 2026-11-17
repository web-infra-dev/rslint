// Upstream reference: eslint-plugin-jest@29.16.6; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('consistent-test-it', {
  valid: [
    ...(
      [
        'test("foo")',
        'test.only("foo")',
        'test.skip("foo")',
        'test.concurrent("foo")',
        'xtest("foo")',
        'test.each([])("foo")',
        'test.each``("foo")',
        'describe("suite", () => { test("foo") })',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
      options: [
        {
          fn: 'test',
        },
      ],
    })),
    ...(
      [
        'it("foo")',
        'fit("foo")',
        'xit("foo")',
        'it.only("foo")',
        'it.skip("foo")',
        'it.concurrent("foo")',
        'it.each([])("foo")',
        'it.each``("foo")',
        'describe("suite", () => { it("foo") })',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index + 8}`,
      code,
      options: [
        {
          fn: 'it',
        },
      ],
    })),
    ...(
      [
        'test("foo")',
        'test.only("foo")',
        'test.skip("foo")',
        'test.concurrent("foo")',
        'xtest("foo")',
        '[1,2,3].forEach(() => { test("foo") })',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index + 17}`,
      code,
      options: [
        {
          fn: 'test',
          withinDescribe: 'it',
        },
      ],
    })),
    ...(
      [
        'it("foo")',
        'it.only("foo")',
        'it.skip("foo")',
        'it.concurrent("foo")',
        'xit("foo")',
        '[1,2,3].forEach(() => { it("foo") })',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index + 23}`,
      code,
      options: [
        {
          fn: 'it',
          withinDescribe: 'test',
        },
      ],
    })),
    ...(
      ['describe("suite", () => { test("foo") })', 'test("foo");'] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index + 29}`,
      code,
      options: [
        {
          fn: 'test',
          withinDescribe: 'test',
        },
      ],
    })),
    ...(['describe("suite", () => { it("foo") })', 'it("foo")'] as const).map(
      (code, index): ValidTestCase => ({
        name: `run 0 valid ${index + 31}`,
        code,
        options: [
          {
            fn: 'it',
            withinDescribe: 'it',
          },
        ],
      }),
    ),
    {
      name: 'run 0 valid 33',
      code: 'test("foo")',
    },
    ...(['test("foo")', 'describe("suite", () => { it("foo") })'] as const).map(
      (code, index): ValidTestCase => ({
        name: `run 0 valid ${index + 34}`,
        code,
        options: [
          {
            withinDescribe: 'it',
          },
        ],
      }),
    ),
    ...(
      ['test("foo")', 'describe("suite", () => { test("foo") })'] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index + 36}`,
      code,
      options: [
        {
          withinDescribe: 'test',
        },
      ],
    })),
    ...(
      [
        'test("foo")',
        'test.only("foo")',
        'test.skip("foo")',
        'test.concurrent("foo")',
        'xtest("foo")',
        'test.each([])("foo")',
        'test.each``("foo")',
        'describe("suite", () => { test("foo") })',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream consistent-test-it with fn=test / valid ${index}`,
      code,
      options: [
        {
          fn: 'test',
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
    })),
    ...(
      [
        'it("foo")',
        'fit("foo")',
        'xit("foo")',
        'it.only("foo")',
        'it.skip("foo")',
        'it.concurrent("foo")',
        'it.each([])("foo")',
        'it.each``("foo")',
        'describe("suite", () => { it("foo") })',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream consistent-test-it with fn=it / valid ${index}`,
      code,
      options: [
        {
          fn: 'it',
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
    })),
    ...(
      [
        'test("foo")',
        'test.only("foo")',
        'test.skip("foo")',
        'test.concurrent("foo")',
        'xtest("foo")',
        '[1,2,3].forEach(() => { test("foo") })',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream consistent-test-it with fn=test and withinDescribe=it  / valid ${index}`,
      code,
      options: [
        {
          fn: 'test',
          withinDescribe: 'it',
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
    })),
    ...(
      [
        'it("foo")',
        'it.only("foo")',
        'it.skip("foo")',
        'it.concurrent("foo")',
        'xit("foo")',
        '[1,2,3].forEach(() => { it("foo") })',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream consistent-test-it with fn=it and withinDescribe=test  / valid ${index}`,
      code,
      options: [
        {
          fn: 'it',
          withinDescribe: 'test',
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
    })),
    ...(
      ['describe("suite", () => { test("foo") })', 'test("foo");'] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream consistent-test-it with fn=test and withinDescribe=test  / valid ${index}`,
      code,
      options: [
        {
          fn: 'test',
          withinDescribe: 'test',
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
    })),
    ...(['describe("suite", () => { it("foo") })', 'it("foo")'] as const).map(
      (code, index): ValidTestCase => ({
        name: `upstream consistent-test-it with fn=it and withinDescribe=it  / valid ${index}`,
        code,
        options: [
          {
            fn: 'it',
            withinDescribe: 'it',
          },
        ],
        languageOptions: {
          ecmaVersion: 2015,
        },
      }),
    ),
    {
      name: 'upstream consistent-test-it defaults without config object / valid 0',
      code: 'test("foo")',
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
    ...(['test("foo")', 'describe("suite", () => { it("foo") })'] as const).map(
      (code, index): ValidTestCase => ({
        name: `upstream consistent-test-it with withinDescribe=it / valid ${index}`,
        code,
        options: [
          {
            withinDescribe: 'it',
          },
        ],
        languageOptions: {
          ecmaVersion: 2015,
        },
      }),
    ),
    ...(
      ['test("foo")', 'describe("suite", () => { test("foo") })'] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream consistent-test-it with withinDescribe=test / valid ${index}`,
      code,
      options: [
        {
          withinDescribe: 'test',
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
    })),
  ],
  invalid: [
    ...(
      [
        ['it("foo")', 1, 3, 'test("foo")'],
        [
          `import { it } from '@jest/globals';

it("foo")`,
          3,
          3,
          `import { test, it } from '@jest/globals';

test("foo")`,
        ],
        [
          `import { it as testThisThing } from '@jest/globals';

testThisThing("foo")`,
          3,
          14,
          `import { test, it as testThisThing } from '@jest/globals';

test("foo")`,
        ],
        ['xit("foo")', 1, 4, 'xtest("foo")'],
        ['fit("foo")', 1, 4, 'test.only("foo")'],
        ['it.skip("foo")', 1, 8, 'test.skip("foo")'],
        ['it.concurrent("foo")', 1, 14, 'test.concurrent("foo")'],
        ['it.only("foo")', 1, 8, 'test.only("foo")'],
        ['it.each([])("foo")', 1, 12, 'test.each([])("foo")'],
        ['it.each``("foo")', 1, 10, 'test.each``("foo")'],
      ] as const
    ).map(([code, line, endColumn, output], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      options: [
        {
          fn: 'test',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethod',
          message: "Prefer using 'test' instead of 'it'",
          line,
          column: 1,
          endLine: line,
          endColumn,
        },
      ],
      output,
    })),
    {
      name: 'run 0 invalid 10',
      code: 'describe.each``("foo", () => { it.each``("bar") })',
      options: [
        {
          fn: 'test',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          message: "Prefer using 'test' instead of 'it' within describe",
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 41,
        },
      ],
      output: 'describe.each``("foo", () => { test.each``("bar") })',
    },
    {
      name: 'run 0 invalid 11',
      code: 'describe.each``("foo", () => { test.each``("bar") })',
      options: [
        {
          fn: 'it',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          message: "Prefer using 'it' instead of 'test' within describe",
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 43,
        },
      ],
      output: 'describe.each``("foo", () => { it.each``("bar") })',
    },
    ...(
      [
        [
          `describe.each()("%s", () => {
  test("is valid, but should not be", () => {});

  it("is not valid, but should be", () => {});
});`,
          `describe.each()("%s", () => {
  it("is valid, but should not be", () => {});

  it("is not valid, but should be", () => {});
});`,
        ],
        [
          `describe.only.each()("%s", () => {
  test("is valid, but should not be", () => {});

  it("is not valid, but should be", () => {});
});`,
          `describe.only.each()("%s", () => {
  it("is valid, but should not be", () => {});

  it("is not valid, but should be", () => {});
});`,
        ],
      ] as const
    ).map(([code, output], index): TestCase => ({
      name: `run 0 invalid ${index + 12}`,
      code,
      options: [
        {
          fn: 'test',
          withinDescribe: 'it',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          message: "Prefer using 'it' instead of 'test' within describe",
          line: 2,
          column: 3,
          endLine: 2,
          endColumn: 7,
        },
      ],
      output,
    })),
    {
      name: 'run 0 invalid 14',
      code: 'describe("suite", () => { it("foo") })',
      options: [
        {
          fn: 'test',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          message: "Prefer using 'test' instead of 'it' within describe",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 29,
        },
      ],
      output: 'describe("suite", () => { test("foo") })',
    },
    ...(
      [
        ['test("foo")', 5, 'it("foo")'],
        ['xtest("foo")', 6, 'xit("foo")'],
        ['test.skip("foo")', 10, 'it.skip("foo")'],
        ['test.concurrent("foo")', 16, 'it.concurrent("foo")'],
        ['test.only("foo")', 10, 'it.only("foo")'],
        ['test.each([])("foo")', 14, 'it.each([])("foo")'],
      ] as const
    ).map(([code, endColumn, output], index): TestCase => ({
      name: `run 0 invalid ${index + 15}`,
      code,
      options: [
        {
          fn: 'it',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethod',
          message: "Prefer using 'it' instead of 'test'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
    {
      name: 'run 0 invalid 21',
      code: 'describe.each``("foo", () => { test.each``("bar") })',
      options: [
        {
          fn: 'it',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          message: "Prefer using 'it' instead of 'test' within describe",
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 43,
        },
      ],
      output: 'describe.each``("foo", () => { it.each``("bar") })',
    },
    {
      name: 'run 0 invalid 22',
      code: 'test.each``("foo")',
      options: [
        {
          fn: 'it',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethod',
          message: "Prefer using 'it' instead of 'test'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 12,
        },
      ],
      output: 'it.each``("foo")',
    },
    {
      name: 'run 0 invalid 23',
      code: 'describe("suite", () => { test("foo") })',
      options: [
        {
          fn: 'it',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          message: "Prefer using 'it' instead of 'test' within describe",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 31,
        },
      ],
      output: 'describe("suite", () => { it("foo") })',
    },
    ...(
      [
        [
          'describe("suite", () => { test("foo") })',
          1,
          27,
          31,
          'describe("suite", () => { it("foo") })',
        ],
        [
          'describe("suite", () => { test.only("foo") })',
          1,
          27,
          36,
          'describe("suite", () => { it.only("foo") })',
        ],
        [
          'describe("suite", () => { xtest("foo") })',
          1,
          27,
          32,
          'describe("suite", () => { xit("foo") })',
        ],
        [
          `import { xtest as dontTestThis } from '@jest/globals';

describe("suite", () => { dontTestThis("foo") });`,
          3,
          27,
          39,
          `import { xit, xtest as dontTestThis } from '@jest/globals';

describe("suite", () => { xit("foo") });`,
        ],
        [
          `import { describe as context, xtest as dontTestThis } from '@jest/globals';

context("suite", () => { dontTestThis("foo") });`,
          3,
          26,
          38,
          `import { describe as context, xit, xtest as dontTestThis } from '@jest/globals';

context("suite", () => { xit("foo") });`,
        ],
        [
          'describe("suite", () => { test.skip("foo") })',
          1,
          27,
          36,
          'describe("suite", () => { it.skip("foo") })',
        ],
        [
          'describe("suite", () => { test.concurrent("foo") })',
          1,
          27,
          42,
          'describe("suite", () => { it.concurrent("foo") })',
        ],
      ] as const
    ).map(([code, line, column, endColumn, output], index): TestCase => ({
      name: `run 0 invalid ${index + 24}`,
      code,
      options: [
        {
          fn: 'test',
          withinDescribe: 'it',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          message: "Prefer using 'it' instead of 'test' within describe",
          line,
          column,
          endLine: line,
          endColumn,
        },
      ],
      output,
    })),
    ...(
      [
        [
          'describe("suite", () => { it("foo") })',
          29,
          'describe("suite", () => { test("foo") })',
        ],
        [
          'describe("suite", () => { it.only("foo") })',
          34,
          'describe("suite", () => { test.only("foo") })',
        ],
        [
          'describe("suite", () => { xit("foo") })',
          30,
          'describe("suite", () => { xtest("foo") })',
        ],
        [
          'describe("suite", () => { it.skip("foo") })',
          34,
          'describe("suite", () => { test.skip("foo") })',
        ],
        [
          'describe("suite", () => { it.concurrent("foo") })',
          40,
          'describe("suite", () => { test.concurrent("foo") })',
        ],
      ] as const
    ).map(([code, endColumn, output], index): TestCase => ({
      name: `run 0 invalid ${index + 31}`,
      code,
      options: [
        {
          fn: 'it',
          withinDescribe: 'test',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          message: "Prefer using 'test' instead of 'it' within describe",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
    {
      name: 'run 0 invalid 36',
      code: 'describe("suite", () => { it("foo") })',
      options: [
        {
          fn: 'test',
          withinDescribe: 'test',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          message: "Prefer using 'test' instead of 'it' within describe",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 29,
        },
      ],
      output: 'describe("suite", () => { test("foo") })',
    },
    {
      name: 'run 0 invalid 37',
      code: 'it("foo")',
      options: [
        {
          fn: 'test',
          withinDescribe: 'test',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethod',
          message: "Prefer using 'test' instead of 'it'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 3,
        },
      ],
      output: 'test("foo")',
    },
    {
      name: 'run 0 invalid 38',
      code: 'describe("suite", () => { test("foo") })',
      options: [
        {
          fn: 'it',
          withinDescribe: 'it',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          message: "Prefer using 'it' instead of 'test' within describe",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 31,
        },
      ],
      output: 'describe("suite", () => { it("foo") })',
    },
    {
      name: 'run 0 invalid 39',
      code: 'test("foo")',
      options: [
        {
          fn: 'it',
          withinDescribe: 'it',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethod',
          message: "Prefer using 'it' instead of 'test'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 5,
        },
      ],
      output: 'it("foo")',
    },
    {
      name: 'run 0 invalid 40',
      code: 'describe("suite", () => { test("foo") })',
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          message: "Prefer using 'it' instead of 'test' within describe",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 31,
        },
      ],
      output: 'describe("suite", () => { it("foo") })',
    },
    {
      name: 'run 0 invalid 41',
      code: 'it("foo")',
      options: [
        {
          withinDescribe: 'it',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethod',
          message: "Prefer using 'test' instead of 'it'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 3,
        },
      ],
      output: 'test("foo")',
    },
    {
      name: 'run 0 invalid 42',
      code: 'describe("suite", () => { test("foo") })',
      options: [
        {
          withinDescribe: 'it',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          message: "Prefer using 'it' instead of 'test' within describe",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 31,
        },
      ],
      output: 'describe("suite", () => { it("foo") })',
    },
    {
      name: 'run 0 invalid 43',
      code: 'it("foo")',
      options: [
        {
          withinDescribe: 'test',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethod',
          message: "Prefer using 'test' instead of 'it'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 3,
        },
      ],
      output: 'test("foo")',
    },
    {
      name: 'run 0 invalid 44',
      code: 'describe("suite", () => { it("foo") })',
      options: [
        {
          withinDescribe: 'test',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          message: "Prefer using 'test' instead of 'it' within describe",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 29,
        },
      ],
      output: 'describe("suite", () => { test("foo") })',
    },
    {
      name: 'run 0 invalid 45',
      code: `test('foo'); // valid
test.only('foo'); // valid

it('foo'); // invalid
it.only('foo'); // invalid`,
      options: [
        {
          fn: 'test',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethod',
          message: "Prefer using 'test' instead of 'it'",
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 3,
        },
        {
          messageId: 'consistentMethod',
          message: "Prefer using 'test' instead of 'it'",
          line: 5,
          column: 1,
          endLine: 5,
          endColumn: 8,
        },
      ],
      output: `test('foo'); // valid
test.only('foo'); // valid

test('foo'); // invalid
test.only('foo'); // invalid`,
    },
    {
      name: 'run 0 invalid 46',
      code: `it('foo'); // valid
it.only('foo'); // valid

test('foo'); // invalid
test.only('foo'); // invalid`,
      options: [
        {
          fn: 'it',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethod',
          message: "Prefer using 'it' instead of 'test'",
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 5,
        },
        {
          messageId: 'consistentMethod',
          message: "Prefer using 'it' instead of 'test'",
          line: 5,
          column: 1,
          endLine: 5,
          endColumn: 10,
        },
      ],
      output: `it('foo'); // valid
it.only('foo'); // valid

it('foo'); // invalid
it.only('foo'); // invalid`,
    },
    {
      name: 'run 0 invalid 47',
      code: `it('foo'); // valid
describe('foo', function () {
  test('bar'); // valid
});

test('foo'); // invalid
describe('foo', function () {
  it('bar'); // invalid
});`,
      options: [
        {
          fn: 'it',
          withinDescribe: 'test',
        },
      ],
      errors: [
        {
          messageId: 'consistentMethod',
          message: "Prefer using 'it' instead of 'test'",
          line: 6,
          column: 1,
          endLine: 6,
          endColumn: 5,
        },
        {
          messageId: 'consistentMethodWithinDescribe',
          message: "Prefer using 'test' instead of 'it' within describe",
          line: 8,
          column: 3,
          endLine: 8,
          endColumn: 5,
        },
      ],
      output: `it('foo'); // valid
describe('foo', function () {
  test('bar'); // valid
});

it('foo'); // invalid
describe('foo', function () {
  test('bar'); // invalid
});`,
    },
    {
      name: 'run 0 invalid 48',
      code: `test('foo'); // valid
describe('foo', function () {
  it('bar'); // valid
});

it('foo'); // invalid
describe('foo', function () {
  test('bar'); // invalid
});`,
      errors: [
        {
          messageId: 'consistentMethod',
          message: "Prefer using 'test' instead of 'it'",
          line: 6,
          column: 1,
          endLine: 6,
          endColumn: 3,
        },
        {
          messageId: 'consistentMethodWithinDescribe',
          message: "Prefer using 'it' instead of 'test' within describe",
          line: 8,
          column: 3,
          endLine: 8,
          endColumn: 7,
        },
      ],
      output: `test('foo'); // valid
describe('foo', function () {
  it('bar'); // valid
});

test('foo'); // invalid
describe('foo', function () {
  it('bar'); // invalid
});`,
    },
    {
      name: 'upstream consistent-test-it with fn=test / invalid 0',
      code: 'it("foo")',
      options: [
        {
          fn: 'test',
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'consistentMethod',
          message: "Prefer using 'test' instead of 'it'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 3,
        },
      ],
      output: 'test("foo")',
    },
    ...(
      [
        [
          `import { it } from '@jest/globals';

it("foo")`,
          3,
          `import { test, it } from '@jest/globals';

test("foo")`,
        ],
        [
          `import { it as testThisThing } from '@jest/globals';

testThisThing("foo")`,
          14,
          `import { test, it as testThisThing } from '@jest/globals';

test("foo")`,
        ],
      ] as const
    ).map(([code, endColumn, output], index): TestCase => ({
      name: `upstream consistent-test-it with fn=test / invalid ${index + 1}`,
      code,
      options: [
        {
          fn: 'test',
        },
      ],
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'consistentMethod',
          message: "Prefer using 'test' instead of 'it'",
          line: 3,
          column: 1,
          endLine: 3,
          endColumn,
        },
      ],
      output,
    })),
    ...(
      [
        ['xit("foo")', 4, 'xtest("foo")'],
        ['fit("foo")', 4, 'test.only("foo")'],
        ['it.skip("foo")', 8, 'test.skip("foo")'],
        ['it.concurrent("foo")', 14, 'test.concurrent("foo")'],
        ['it.only("foo")', 8, 'test.only("foo")'],
        ['it.each([])("foo")', 12, 'test.each([])("foo")'],
        ['it.each``("foo")', 10, 'test.each``("foo")'],
      ] as const
    ).map(([code, endColumn, output], index): TestCase => ({
      name: `upstream consistent-test-it with fn=test / invalid ${index + 3}`,
      code,
      options: [
        {
          fn: 'test',
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'consistentMethod',
          message: "Prefer using 'test' instead of 'it'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
    {
      name: 'upstream consistent-test-it with fn=test / invalid 10',
      code: 'describe.each``("foo", () => { it.each``("bar") })',
      options: [
        {
          fn: 'test',
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          message: "Prefer using 'test' instead of 'it' within describe",
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 41,
        },
      ],
      output: 'describe.each``("foo", () => { test.each``("bar") })',
    },
    {
      name: 'upstream consistent-test-it with fn=test / invalid 11',
      code: 'describe.each``("foo", () => { test.each``("bar") })',
      options: [
        {
          fn: 'it',
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          message: "Prefer using 'it' instead of 'test' within describe",
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 43,
        },
      ],
      output: 'describe.each``("foo", () => { it.each``("bar") })',
    },
    ...(
      [
        [
          `describe.each()("%s", () => {
  test("is valid, but should not be", () => {});

  it("is not valid, but should be", () => {});
});`,
          `describe.each()("%s", () => {
  it("is valid, but should not be", () => {});

  it("is not valid, but should be", () => {});
});`,
        ],
        [
          `describe.only.each()("%s", () => {
  test("is valid, but should not be", () => {});

  it("is not valid, but should be", () => {});
});`,
          `describe.only.each()("%s", () => {
  it("is valid, but should not be", () => {});

  it("is not valid, but should be", () => {});
});`,
        ],
      ] as const
    ).map(([code, output], index): TestCase => ({
      name: `upstream consistent-test-it with fn=test / invalid ${index + 12}`,
      code,
      options: [
        {
          fn: 'test',
          withinDescribe: 'it',
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          message: "Prefer using 'it' instead of 'test' within describe",
          line: 2,
          column: 3,
          endLine: 2,
          endColumn: 7,
        },
      ],
      output,
    })),
    {
      name: 'upstream consistent-test-it with fn=test / invalid 14',
      code: 'describe("suite", () => { it("foo") })',
      options: [
        {
          fn: 'test',
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          message: "Prefer using 'test' instead of 'it' within describe",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 29,
        },
      ],
      output: 'describe("suite", () => { test("foo") })',
    },
    ...(
      [
        [
          'test("foo")',
          'upstream consistent-test-it with fn=it / invalid 0',
          5,
          'it("foo")',
        ],
        [
          'xtest("foo")',
          'upstream consistent-test-it with fn=it / invalid 1',
          6,
          'xit("foo")',
        ],
        [
          'test.skip("foo")',
          'upstream consistent-test-it with fn=it / invalid 2',
          10,
          'it.skip("foo")',
        ],
        [
          'test.concurrent("foo")',
          'upstream consistent-test-it with fn=it / invalid 3',
          16,
          'it.concurrent("foo")',
        ],
        [
          'test.only("foo")',
          'upstream consistent-test-it with fn=it / invalid 4',
          10,
          'it.only("foo")',
        ],
        [
          'test.each([])("foo")',
          'upstream consistent-test-it with fn=it / invalid 5',
          14,
          'it.each([])("foo")',
        ],
        [
          'test.each``("foo")',
          'upstream consistent-test-it with fn=it / invalid 7',
          12,
          'it.each``("foo")',
        ],
      ] as const
    ).map(([code, name, endColumn, output]): TestCase => ({
      name,
      code,
      options: [
        {
          fn: 'it',
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'consistentMethod',
          message: "Prefer using 'it' instead of 'test'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
    {
      name: 'upstream consistent-test-it with fn=it / invalid 8',
      code: 'describe("suite", () => { test("foo") })',
      options: [
        {
          fn: 'it',
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          message: "Prefer using 'it' instead of 'test' within describe",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 31,
        },
      ],
      output: 'describe("suite", () => { it("foo") })',
    },
    ...(
      [
        [
          'describe("suite", () => { test("foo") })',
          31,
          'describe("suite", () => { it("foo") })',
        ],
        [
          'describe("suite", () => { test.only("foo") })',
          36,
          'describe("suite", () => { it.only("foo") })',
        ],
        [
          'describe("suite", () => { xtest("foo") })',
          32,
          'describe("suite", () => { xit("foo") })',
        ],
      ] as const
    ).map(([code, endColumn, output], index): TestCase => ({
      name: `upstream consistent-test-it with fn=test and withinDescribe=it  / invalid ${index}`,
      code,
      options: [
        {
          fn: 'test',
          withinDescribe: 'it',
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          message: "Prefer using 'it' instead of 'test' within describe",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
    ...(
      [
        [
          `import { xtest as dontTestThis } from '@jest/globals';

describe("suite", () => { dontTestThis("foo") });`,
          27,
          `import { xit, xtest as dontTestThis } from '@jest/globals';

describe("suite", () => { xit("foo") });`,
        ],
        [
          `import { describe as context, xtest as dontTestThis } from '@jest/globals';

context("suite", () => { dontTestThis("foo") });`,
          26,
          `import { describe as context, xit, xtest as dontTestThis } from '@jest/globals';

context("suite", () => { xit("foo") });`,
        ],
      ] as const
    ).map(([code, column, output], index): TestCase => ({
      name: `upstream consistent-test-it with fn=test and withinDescribe=it  / invalid ${index + 3}`,
      code,
      options: [
        {
          fn: 'test',
          withinDescribe: 'it',
        },
      ],
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          message: "Prefer using 'it' instead of 'test' within describe",
          line: 3,
          column,
          endLine: 3,
          endColumn: column + 12,
        },
      ],
      output,
    })),
    ...(
      [
        [
          'describe("suite", () => { test.skip("foo") })',
          36,
          'describe("suite", () => { it.skip("foo") })',
        ],
        [
          'describe("suite", () => { test.concurrent("foo") })',
          42,
          'describe("suite", () => { it.concurrent("foo") })',
        ],
      ] as const
    ).map(([code, endColumn, output], index): TestCase => ({
      name: `upstream consistent-test-it with fn=test and withinDescribe=it  / invalid ${index + 5}`,
      code,
      options: [
        {
          fn: 'test',
          withinDescribe: 'it',
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          message: "Prefer using 'it' instead of 'test' within describe",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
    ...(
      [
        [
          'describe("suite", () => { it("foo") })',
          29,
          'describe("suite", () => { test("foo") })',
        ],
        [
          'describe("suite", () => { it.only("foo") })',
          34,
          'describe("suite", () => { test.only("foo") })',
        ],
        [
          'describe("suite", () => { xit("foo") })',
          30,
          'describe("suite", () => { xtest("foo") })',
        ],
        [
          'describe("suite", () => { it.skip("foo") })',
          34,
          'describe("suite", () => { test.skip("foo") })',
        ],
        [
          'describe("suite", () => { it.concurrent("foo") })',
          40,
          'describe("suite", () => { test.concurrent("foo") })',
        ],
      ] as const
    ).map(([code, endColumn, output], index): TestCase => ({
      name: `upstream consistent-test-it with fn=it and withinDescribe=test  / invalid ${index}`,
      code,
      options: [
        {
          fn: 'it',
          withinDescribe: 'test',
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          message: "Prefer using 'test' instead of 'it' within describe",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
    {
      name: 'upstream consistent-test-it with fn=test and withinDescribe=test  / invalid 0',
      code: 'describe("suite", () => { it("foo") })',
      options: [
        {
          fn: 'test',
          withinDescribe: 'test',
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          message: "Prefer using 'test' instead of 'it' within describe",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 29,
        },
      ],
      output: 'describe("suite", () => { test("foo") })',
    },
    {
      name: 'upstream consistent-test-it with fn=test and withinDescribe=test  / invalid 1',
      code: 'it("foo")',
      options: [
        {
          fn: 'test',
          withinDescribe: 'test',
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'consistentMethod',
          message: "Prefer using 'test' instead of 'it'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 3,
        },
      ],
      output: 'test("foo")',
    },
    {
      name: 'upstream consistent-test-it with fn=it and withinDescribe=it  / invalid 0',
      code: 'describe("suite", () => { test("foo") })',
      options: [
        {
          fn: 'it',
          withinDescribe: 'it',
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          message: "Prefer using 'it' instead of 'test' within describe",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 31,
        },
      ],
      output: 'describe("suite", () => { it("foo") })',
    },
    {
      name: 'upstream consistent-test-it with fn=it and withinDescribe=it  / invalid 1',
      code: 'test("foo")',
      options: [
        {
          fn: 'it',
          withinDescribe: 'it',
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'consistentMethod',
          message: "Prefer using 'it' instead of 'test'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 5,
        },
      ],
      output: 'it("foo")',
    },
    {
      name: 'upstream consistent-test-it defaults without config object / invalid 0',
      code: 'describe("suite", () => { test("foo") })',
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          message: "Prefer using 'it' instead of 'test' within describe",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 31,
        },
      ],
      output: 'describe("suite", () => { it("foo") })',
    },
    {
      name: 'upstream consistent-test-it with withinDescribe=it / invalid 0',
      code: 'it("foo")',
      options: [
        {
          withinDescribe: 'it',
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'consistentMethod',
          message: "Prefer using 'test' instead of 'it'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 3,
        },
      ],
      output: 'test("foo")',
    },
    {
      name: 'upstream consistent-test-it with withinDescribe=it / invalid 1',
      code: 'describe("suite", () => { test("foo") })',
      options: [
        {
          withinDescribe: 'it',
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          message: "Prefer using 'it' instead of 'test' within describe",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 31,
        },
      ],
      output: 'describe("suite", () => { it("foo") })',
    },
    {
      name: 'upstream consistent-test-it with withinDescribe=test / invalid 0',
      code: 'it("foo")',
      options: [
        {
          withinDescribe: 'test',
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'consistentMethod',
          message: "Prefer using 'test' instead of 'it'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 3,
        },
      ],
      output: 'test("foo")',
    },
    {
      name: 'upstream consistent-test-it with withinDescribe=test / invalid 1',
      code: 'describe("suite", () => { it("foo") })',
      options: [
        {
          withinDescribe: 'test',
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'consistentMethodWithinDescribe',
          message: "Prefer using 'test' instead of 'it' within describe",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 29,
        },
      ],
      output: 'describe("suite", () => { test("foo") })',
    },
  ],
});

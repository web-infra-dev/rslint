// Upstream reference: eslint-plugin-jest@29.16.6; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('prefer-lowercase-title', {
  valid: [
    ...(
      [
        'it.each()',
        'it.each()(1)',
        'randomFunction()',
        'foo.bar()',
        'it()',
        "it(' ', function () {})",
        'it(true, function () {})',
        'it(MY_CONSTANT, function () {})',
        'it(" ", function () {})',
        'it(` `, function () {})',
        "it('foo', function () {})",
        'it("foo", function () {})',
        'it(`foo`, function () {})',
        'it("<Foo/>", function () {})',
        'it("123 foo", function () {})',
        'it(42, function () {})',
        'it(``)',
        'it("")',
        'it(42)',
        'test()',
        "test('foo', function () {})",
        'test("foo", function () {})',
        'test(`foo`, function () {})',
        'test("<Foo/>", function () {})',
        'test("123 foo", function () {})',
        'test("42", function () {})',
        'test(``)',
        'test("")',
        'test(42)',
        'describe()',
        "describe('foo', function () {})",
        'describe("foo", function () {})',
        'describe(`foo`, function () {})',
        'describe("<Foo/>", function () {})',
        'describe("123 foo", function () {})',
        'describe("42", function () {})',
        'describe(function () {})',
        'describe(``)',
        'describe("")',
        'describe.each()(1); describe.each()(2);',
        'jest.doMock("my-module")',
        `
        import { jest } from '@jest/globals';
        jest.doMock('my-module');
     \u0020`,
        'describe(42)',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    ...(
      [
        'it.each()',
        'it.each()(1)',
        'randomFunction()',
        'foo.bar()',
        'it()',
        "it(' ', function () {})",
        'it(true, function () {})',
        'it(MY_CONSTANT, function () {})',
        'it(" ", function () {})',
        'it(` `, function () {})',
        "it('foo', function () {})",
        'it("foo", function () {})',
        'it(`foo`, function () {})',
        'it("<Foo/>", function () {})',
        'it("123 foo", function () {})',
        'it(42, function () {})',
        'it(``)',
        'it("")',
        'it(42)',
        'test()',
        "test('foo', function () {})",
        'test("foo", function () {})',
        'test(`foo`, function () {})',
        'test("<Foo/>", function () {})',
        'test("123 foo", function () {})',
        'test("42", function () {})',
        'test(``)',
        'test("")',
        'test(42)',
        'describe()',
        "describe('foo', function () {})",
        'describe("foo", function () {})',
        'describe(`foo`, function () {})',
        'describe("<Foo/>", function () {})',
        'describe("123 foo", function () {})',
        'describe("42", function () {})',
        'describe(function () {})',
        'describe(``)',
        'describe("")',
        `describe.each()(1);
describe.each()(2);`,
        'jest.doMock("my-module")',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream prefer-lowercase-title / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
    })),
    {
      name: 'upstream prefer-lowercase-title / valid 41',
      code: `import { jest } from '@jest/globals';

jest.doMock('my-module');`,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2015,
      },
    },
    {
      name: 'upstream prefer-lowercase-title / valid 42',
      code: 'describe(42)',
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
    {
      name: 'upstream prefer-lowercase-title / valid 43',
      code: 'describe(42)',
      options: [
        {
          ignore: undefined,
          allowedPrefixes: undefined,
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
    ...(
      [
        "describe('Foo', function () {})",
        'describe("Foo", function () {})',
        'describe(`Foo`, function () {})',
        'fdescribe(`Foo`, function () {})',
        'describe.skip(`Foo`, function () {})',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream prefer-lowercase-title with ignore=describe / valid ${index}`,
      code,
      options: [
        {
          ignore: ['describe'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
    })),
    ...(
      [
        "test('Foo', function () {})",
        'test("Foo", function () {})',
        'test(`Foo`, function () {})',
        'xtest(`Foo`, function () {})',
        'test.only(`Foo`, function () {})',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream prefer-lowercase-title with ignore=test / valid ${index}`,
      code,
      options: [
        {
          ignore: ['test'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
    })),
    ...(
      [
        "it('Foo', function () {})",
        'it("Foo", function () {})',
        'it(`Foo`, function () {})',
        'fit(`Foo`, function () {})',
        'it.skip(`Foo`, function () {})',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream prefer-lowercase-title with ignore=it / valid ${index}`,
      code,
      options: [
        {
          ignore: ['it'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
    })),
    {
      name: 'upstream prefer-lowercase-title with allowedPrefixes / valid 0',
      code: "it('GET /live', function () {})",
      options: [
        {
          allowedPrefixes: ['GET'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
    {
      name: 'upstream prefer-lowercase-title with allowedPrefixes / valid 1',
      code: 'it("POST /live", function () {})',
      options: [
        {
          allowedPrefixes: ['GET', 'POST'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
    {
      name: 'upstream prefer-lowercase-title with allowedPrefixes / valid 2',
      code: 'it(`PATCH /live`, function () {})',
      options: [
        {
          allowedPrefixes: ['GET', 'PATCH'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
    ...(
      [
        'describe("MyClass", () => {});',
        `describe('MyClass', () => {
  describe('#myMethod', () => {
    it('does things', () => {
      //
    });
  });
});`,
        `describe('Strings', () => {
  it('are strings', () => {
    expect('abc').toBe('abc');
  });
});

describe('Booleans', () => {
  it('are booleans', () => {
    expect(true).toBe(true);
  });
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream prefer-lowercase-title with ignoreTopLevelDescribe / valid ${index}`,
      code,
      options: [
        {
          ignoreTopLevelDescribe: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
    })),
    ...(
      [
        'test.todo(`Foo`, function () {})',
        'it.todo(`Foo`, function () {})',
        'it.todo("Foo", () => {})',
        'it.only.todo("Foo", () => {})',
        'it.todo.only("Foo", () => {})',
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream prefer-lowercase-title with ignoreTodos / valid ${index}`,
      code,
      options: [
        {
          ignoreTodos: true,
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
        [
          "it('Foo', function () {})",
          '`it`s should begin with lowercase',
          4,
          9,
          "it('foo', function () {})",
        ],
        [
          "xit('Foo', function () {})",
          '`xit`s should begin with lowercase',
          5,
          10,
          "xit('foo', function () {})",
        ],
        [
          'it("Foo", function () {})',
          '`it`s should begin with lowercase',
          4,
          9,
          'it("foo", function () {})',
        ],
        [
          'it(`Foo`, function () {})',
          '`it`s should begin with lowercase',
          4,
          9,
          'it(`foo`, function () {})',
        ],
        [
          "test('Foo', function () {})",
          '`test`s should begin with lowercase',
          6,
          11,
          "test('foo', function () {})",
        ],
        [
          "xtest('Foo', function () {})",
          '`xtest`s should begin with lowercase',
          7,
          12,
          "xtest('foo', function () {})",
        ],
        [
          "describe('Foo', function () {})",
          '`describe`s should begin with lowercase',
          10,
          15,
          "describe('foo', function () {})",
        ],
        [
          'describe(`Foo`, function () {})',
          '`describe`s should begin with lowercase',
          10,
          15,
          'describe(`foo`, function () {})',
        ],
        [
          "it.each(['green', 'black'])('Should return %', () => {})",
          '`it`s should begin with lowercase',
          29,
          46,
          "it.each(['green', 'black'])('should return %', () => {})",
        ],
        [
          "test('Doesn\\'t mutate', () => {})",
          '`test`s should begin with lowercase',
          6,
          23,
          "test('doesn\\'t mutate', () => {})",
        ],
        [
          'test(`Value: \\${name}`, () => {})',
          '`test`s should begin with lowercase',
          6,
          23,
          'test(`value: \\${name}`, () => {})',
        ],
      ] as const
    ).map(([code, message, column, endColumn, output], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      errors: [
        {
          messageId: 'unexpectedCase',
          message,
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
    ...(
      [
        [
          "it('Foo', function () {})",
          '`it`s should begin with lowercase',
          4,
          "it('foo', function () {})",
        ],
        [
          "xit('Foo', function () {})",
          '`xit`s should begin with lowercase',
          5,
          "xit('foo', function () {})",
        ],
        [
          'it("Foo", function () {})',
          '`it`s should begin with lowercase',
          4,
          'it("foo", function () {})',
        ],
        [
          'it(`Foo`, function () {})',
          '`it`s should begin with lowercase',
          4,
          'it(`foo`, function () {})',
        ],
        [
          "test('Foo', function () {})",
          '`test`s should begin with lowercase',
          6,
          "test('foo', function () {})",
        ],
        [
          "xtest('Foo', function () {})",
          '`xtest`s should begin with lowercase',
          7,
          "xtest('foo', function () {})",
        ],
        [
          'test("Foo", function () {})',
          '`test`s should begin with lowercase',
          6,
          'test("foo", function () {})',
        ],
        [
          'test(`Foo`, function () {})',
          '`test`s should begin with lowercase',
          6,
          'test(`foo`, function () {})',
        ],
        [
          "describe('Foo', function () {})",
          '`describe`s should begin with lowercase',
          10,
          "describe('foo', function () {})",
        ],
        [
          'describe("Foo", function () {})',
          '`describe`s should begin with lowercase',
          10,
          'describe("foo", function () {})',
        ],
        [
          'describe(`Foo`, function () {})',
          '`describe`s should begin with lowercase',
          10,
          'describe(`foo`, function () {})',
        ],
      ] as const
    ).map(([code, message, column, output], index): TestCase => ({
      name: `upstream prefer-lowercase-title / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'unexpectedCase',
          message,
          line: 1,
          column,
          endLine: 1,
          endColumn: column + 5,
        },
      ],
      output,
    })),
    {
      name: 'upstream prefer-lowercase-title / invalid 11',
      code: `import { describe as context } from '@jest/globals';

context(\`Foo\`, () => {});`,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'unexpectedCase',
          message: '`describe`s should begin with lowercase',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 14,
        },
      ],
      output: `import { describe as context } from '@jest/globals';

context(\`foo\`, () => {});`,
    },
    ...(
      [
        [
          'describe(`Some longer description`, function () {})',
          '`describe`s should begin with lowercase',
          10,
          35,
          'describe(`some longer description`, function () {})',
        ],
        [
          'fdescribe(`Some longer description`, function () {})',
          '`fdescribe`s should begin with lowercase',
          11,
          36,
          'fdescribe(`some longer description`, function () {})',
        ],
        [
          "it.each(['green', 'black'])('Should return %', () => {})",
          '`it`s should begin with lowercase',
          29,
          46,
          "it.each(['green', 'black'])('should return %', () => {})",
        ],
        [
          "describe.each(['green', 'black'])('Should return %', () => {})",
          '`describe`s should begin with lowercase',
          35,
          52,
          "describe.each(['green', 'black'])('should return %', () => {})",
        ],
      ] as const
    ).map(([code, message, column, endColumn, output], index): TestCase => ({
      name: `upstream prefer-lowercase-title / invalid ${index + 12}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'unexpectedCase',
          message,
          line: 1,
          column,
          endLine: 1,
          endColumn,
        },
      ],
      output,
    })),
    ...(
      [
        [
          "test('Foo', function () {})",
          '`test`s should begin with lowercase',
          6,
          "test('foo', function () {})",
        ],
        [
          "xit('Foo', function () {})",
          '`xit`s should begin with lowercase',
          5,
          "xit('foo', function () {})",
        ],
      ] as const
    ).map(([code, message, column, output], index): TestCase => ({
      name: `upstream prefer-lowercase-title with ignore=describe / invalid ${index}`,
      code,
      options: [
        {
          ignore: ['describe'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'unexpectedCase',
          message,
          line: 1,
          column,
          endLine: 1,
          endColumn: column + 5,
        },
      ],
      output,
    })),
    ...(
      [
        [
          "describe('Foo', function () {})",
          '`describe`s should begin with lowercase',
          10,
          "describe('foo', function () {})",
        ],
        [
          "it('Foo', function () {})",
          '`it`s should begin with lowercase',
          4,
          "it('foo', function () {})",
        ],
        [
          "xit('Foo', function () {})",
          '`xit`s should begin with lowercase',
          5,
          "xit('foo', function () {})",
        ],
      ] as const
    ).map(([code, message, column, output], index): TestCase => ({
      name: `upstream prefer-lowercase-title with ignore=test / invalid ${index}`,
      code,
      options: [
        {
          ignore: ['test'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'unexpectedCase',
          message,
          line: 1,
          column,
          endLine: 1,
          endColumn: column + 5,
        },
      ],
      output,
    })),
    ...(
      [
        [
          "describe('Foo', function () {})",
          '`describe`s should begin with lowercase',
          10,
          "describe('foo', function () {})",
        ],
        [
          "test('Foo', function () {})",
          '`test`s should begin with lowercase',
          6,
          "test('foo', function () {})",
        ],
        [
          "xtest('Foo', function () {})",
          '`xtest`s should begin with lowercase',
          7,
          "xtest('foo', function () {})",
        ],
      ] as const
    ).map(([code, message, column, output], index): TestCase => ({
      name: `upstream prefer-lowercase-title with ignore=it / invalid ${index}`,
      code,
      options: [
        {
          ignore: ['it'],
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'unexpectedCase',
          message,
          line: 1,
          column,
          endLine: 1,
          endColumn: column + 5,
        },
      ],
      output,
    })),
    {
      name: 'upstream prefer-lowercase-title with ignoreTopLevelDescribe / invalid 0',
      code: 'it("Works!", () => {});',
      options: [
        {
          ignoreTopLevelDescribe: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'unexpectedCase',
          message: '`it`s should begin with lowercase',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 12,
        },
      ],
      output: 'it("works!", () => {});',
    },
    {
      name: 'upstream prefer-lowercase-title with ignoreTopLevelDescribe / invalid 1',
      code: `describe('MyClass', () => {
  describe('MyMethod', () => {
    it('Does things', () => {
      //
    });
  });
});`,
      options: [
        {
          ignoreTopLevelDescribe: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'unexpectedCase',
          message: '`describe`s should begin with lowercase',
          line: 2,
          column: 12,
          endLine: 2,
          endColumn: 22,
        },
        {
          messageId: 'unexpectedCase',
          message: '`it`s should begin with lowercase',
          line: 3,
          column: 8,
          endLine: 3,
          endColumn: 21,
        },
      ],
      output: `describe('MyClass', () => {
  describe('myMethod', () => {
    it('does things', () => {
      //
    });
  });
});`,
    },
    {
      name: 'upstream prefer-lowercase-title with ignoreTopLevelDescribe / invalid 2',
      code: `import { describe, describe as context } from '@jest/globals';

describe('MyClass', () => {
  context('MyMethod', () => {
    it('Does things', () => {
      //
    });
  });
});`,
      options: [
        {
          ignoreTopLevelDescribe: true,
        },
      ],
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'unexpectedCase',
          message: '`describe`s should begin with lowercase',
          line: 4,
          column: 11,
          endLine: 4,
          endColumn: 21,
        },
        {
          messageId: 'unexpectedCase',
          message: '`it`s should begin with lowercase',
          line: 5,
          column: 8,
          endLine: 5,
          endColumn: 21,
        },
      ],
      output: `import { describe, describe as context } from '@jest/globals';

describe('MyClass', () => {
  context('myMethod', () => {
    it('does things', () => {
      //
    });
  });
});`,
    },
    {
      name: 'upstream prefer-lowercase-title with ignoreTopLevelDescribe / invalid 3',
      code: `describe('MyClass', () => {
  describe('MyMethod', () => {
    it('Does things', () => {
      //
    });
  });
});`,
      options: [
        {
          ignoreTopLevelDescribe: false,
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'unexpectedCase',
          message: '`describe`s should begin with lowercase',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 19,
        },
        {
          messageId: 'unexpectedCase',
          message: '`describe`s should begin with lowercase',
          line: 2,
          column: 12,
          endLine: 2,
          endColumn: 22,
        },
        {
          messageId: 'unexpectedCase',
          message: '`it`s should begin with lowercase',
          line: 3,
          column: 8,
          endLine: 3,
          endColumn: 21,
        },
      ],
      output: `describe('myClass', () => {
  describe('myMethod', () => {
    it('does things', () => {
      //
    });
  });
});`,
    },
    ...(
      [
        [
          "describe('Foo', function () {})",
          '`describe`s should begin with lowercase',
          10,
          "describe('foo', function () {})",
        ],
        [
          "it('Foo', function () {})",
          '`it`s should begin with lowercase',
          4,
          "it('foo', function () {})",
        ],
        [
          "test('Foo', function () {})",
          '`test`s should begin with lowercase',
          6,
          "test('foo', function () {})",
        ],
        [
          "test.only('Foo', function () {})",
          '`test`s should begin with lowercase',
          11,
          "test.only('foo', function () {})",
        ],
        [
          "test.skip('Foo', function () {})",
          '`test`s should begin with lowercase',
          11,
          "test.skip('foo', function () {})",
        ],
      ] as const
    ).map(([code, message, column, output], index): TestCase => ({
      name: `upstream prefer-lowercase-title with ignoreTodos / invalid ${index}`,
      code,
      options: [
        {
          ignoreTodos: true,
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'unexpectedCase',
          message,
          line: 1,
          column,
          endLine: 1,
          endColumn: column + 5,
        },
      ],
      output,
    })),
  ],
});

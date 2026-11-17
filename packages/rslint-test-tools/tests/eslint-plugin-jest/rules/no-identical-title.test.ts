// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('no-identical-title', {
  valid: [
    ...(
      [
        'it(); it();',
        'describe(); describe();',
        'describe("foo", () => {}); it("foo", () => {});',
        `
      describe("foo", () => {
        it("works", () => {});
      });
   \u0020`,
        `
      it('one', () => {});
      it('two', () => {});
   \u0020`,
        `
      describe('foo', () => {});
      describe('foe', () => {});
   \u0020`,
        `
      it(\`one\`, () => {});
      it(\`two\`, () => {});
   \u0020`,
        `
      describe(\`foo\`, () => {});
      describe(\`foe\`, () => {});
   \u0020`,
        `
      describe('foo', () => {
        test('this', () => {});
        test('that', () => {});
      });
   \u0020`,
        `
      test.concurrent('this', () => {});
      test.concurrent('that', () => {});
   \u0020`,
        `
      test.concurrent('this', () => {});
      test.only.concurrent('that', () => {});
   \u0020`,
        `
      test.only.concurrent('this', () => {});
      test.concurrent('that', () => {});
   \u0020`,
        `
      test.only.concurrent('this', () => {});
      test.only.concurrent('that', () => {});
   \u0020`,
        `
      test.only('this', () => {});
      test.only('that', () => {});
   \u0020`,
        `
      describe('foo', () => {
        it('works', () => {});

        describe('foe', () => {
          it('works', () => {});
        });
      });
   \u0020`,
        `
      describe('foo', () => {
        describe('foe', () => {
          it('works', () => {});
        });

        it('works', () => {});
      });
   \u0020`,
        "describe('foo', () => describe('foe', () => {}));",
        `
      describe('foo', () => {
        describe('foe', () => {});
      });

      describe('foe', () => {});
   \u0020`,
        'test("number" + n, function() {});',
        'test("number" + n, function() {}); test("number" + n, function() {});',
        'it(`${n}`, function() {});',
        'it(`${n}`, function() {}); it(`${n}`, function() {});',
        `
      describe('a class named ' + myClass.name, () => {
        describe('#myMethod', () => {});
      });

      describe('something else', () => {});
   \u0020`,
        `
      describe('my class', () => {
        describe('#myMethod', () => {});
        describe('a class named ' + myClass.name, () => {});
      });
   \u0020`,
        `
      describe("foo", () => {
        it(\`ignores \${someVar} with the same title\`, () => {});
        it(\`ignores \${someVar} with the same title\`, () => {});
      });
   \u0020`,
        `
      const test = { content: () => "foo" };
      test.content(\`something that is not from jest\`, () => {});
      test.content(\`something that is not from jest\`, () => {});
   \u0020`,
        `
      const describe = { content: () => "foo" };
      describe.content(\`something that is not from jest\`, () => {});
      describe.content(\`something that is not from jest\`, () => {});
   \u0020`,
        `
      describe.each\`
        description
        \${'b'}
      \`('$description', () => {});

      describe.each\`
        description
        \${'a'}
      \`('$description', () => {});
   \u0020`,
        `
      describe('top level', () => {
        describe.each\`\`('nested each', () => {
          describe.each\`\`('nested nested each', () => {});
        });

        describe('nested', () => {});
      });
   \u0020`,
        `
      describe.each\`\`('my title', value => {});
      describe.each\`\`('my title', value => {});
      describe.each([])('my title', value => {});
      describe.each([])('my title', value => {});
   \u0020`,
        `
      describe.each([])('when the value is %s', value => {});
      describe.each([])('when the value is %s', value => {});
   \u0020`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    ...(
      [
        'it(); it();',
        'describe(); describe();',
        'describe("foo", () => {}); it("foo", () => {});',
        `describe("foo", () => {
  it("works", () => {});
});`,
        `it('one', () => {});
it('two', () => {});`,
        `describe('foo', () => {});
describe('foe', () => {});`,
        `it(\`one\`, () => {});
it(\`two\`, () => {});`,
        `describe(\`foo\`, () => {});
describe(\`foe\`, () => {});`,
        `describe('foo', () => {
  test('this', () => {});
  test('that', () => {});
});`,
        `test.concurrent('this', () => {});
test.concurrent('that', () => {});`,
        `test.concurrent('this', () => {});
test.only.concurrent('that', () => {});`,
        `test.only.concurrent('this', () => {});
test.concurrent('that', () => {});`,
        `test.only.concurrent('this', () => {});
test.only.concurrent('that', () => {});`,
        `test.only('this', () => {});
test.only('that', () => {});`,
        `describe('foo', () => {
  it('works', () => {});

  describe('foe', () => {
    it('works', () => {});
  });
});`,
        `describe('foo', () => {
  describe('foe', () => {
    it('works', () => {});
  });

  it('works', () => {});
});`,
        "describe('foo', () => describe('foe', () => {}));",
        `describe('foo', () => {
  describe('foe', () => {});
});

describe('foe', () => {});`,
        'test("number" + n, function() {});',
        'test("number" + n, function() {}); test("number" + n, function() {});',
        'it(`${n}`, function() {});',
        'it(`${n}`, function() {}); it(`${n}`, function() {});',
        `describe('a class named ' + myClass.name, () => {
  describe('#myMethod', () => {});
});

describe('something else', () => {});`,
        `describe('my class', () => {
  describe('#myMethod', () => {});
  describe('a class named ' + myClass.name, () => {});
});`,
        `describe("foo", () => {
  it(\`ignores \${someVar} with the same title\`, () => {});
  it(\`ignores \${someVar} with the same title\`, () => {});
});`,
        `const test = { content: () => "foo" };
test.content(\`something that is not from jest\`, () => {});
test.content(\`something that is not from jest\`, () => {});`,
        `const describe = { content: () => "foo" };
describe.content(\`something that is not from jest\`, () => {});
describe.content(\`something that is not from jest\`, () => {});`,
        `describe.each\`
  description
  \${'b'}
\`('$description', () => {});

describe.each\`
  description
  \${'a'}
\`('$description', () => {});`,
        `describe('top level', () => {
  describe.each\`\`('nested each', () => {
    describe.each\`\`('nested nested each', () => {});
  });

  describe('nested', () => {});
});`,
        `describe.each\`\`('my title', value => {});
describe.each\`\`('my title', value => {});
describe.each([])('my title', value => {});
describe.each([])('my title', value => {});`,
        `describe.each([])('when the value is %s', value => {});
describe.each([])('when the value is %s', value => {});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream no-identical-title / valid ${index}`,
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
          `
        describe('foo', () => {
          it('works', () => {});
          it('works', () => {});
        });
     \u0020`,
          4,
          14,
          21,
        ],
        [
          `
        it('works', () => {});
        it('works', () => {});
     \u0020`,
          3,
          12,
          19,
        ],
        [
          `
        test.only('this', () => {});
        test('this', () => {});
     \u0020`,
          3,
          14,
          20,
        ],
        [
          `
        xtest('this', () => {});
        test('this', () => {});
     \u0020`,
          3,
          14,
          20,
        ],
        [
          `
        test.only('this', () => {});
        test.only('this', () => {});
     \u0020`,
          3,
          19,
          25,
        ],
        [
          `
        test.concurrent('this', () => {});
        test.concurrent('this', () => {});
     \u0020`,
          3,
          25,
          31,
        ],
        [
          `
        test.only('this', () => {});
        test.concurrent('this', () => {});
     \u0020`,
          3,
          25,
          31,
        ],
      ] as const
    ).map(([code, line, column, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index}`,
      code,
      errors: [
        {
          messageId: 'multipleTestTitle',
          message:
            'Test title is used multiple times in the same describe block',
          line,
          column,
          endLine: line,
          endColumn,
        },
      ],
    })),
    ...(
      [
        [
          `
        describe('foo', () => {});
        describe('foo', () => {});
     \u0020`,
          3,
          18,
        ],
        [
          `
        describe('foo', () => {});
        xdescribe('foo', () => {});
     \u0020`,
          3,
          19,
        ],
        [
          `
        fdescribe('foo', () => {});
        describe('foo', () => {});
     \u0020`,
          3,
          18,
        ],
        [
          `
        describe('foo', () => {
          describe('foe', () => {});
        });
        describe('foo', () => {});
     \u0020`,
          5,
          18,
        ],
      ] as const
    ).map(([code, line, column], index): TestCase => ({
      name: `run 0 invalid ${index + 7}`,
      code,
      errors: [
        {
          messageId: 'multipleDescribeTitle',
          message:
            'Describe block title is used multiple times in the same describe block',
          line,
          column,
          endLine: line,
          endColumn: column + 5,
        },
      ],
    })),
    {
      name: 'run 0 invalid 11',
      code: `
        describe("foo", () => {
          it(\`catches backticks with the same title\`, () => {});
          it(\`catches backticks with the same title\`, () => {});
        });
     \u0020`,
      errors: [
        {
          messageId: 'multipleTestTitle',
          message:
            'Test title is used multiple times in the same describe block',
          line: 4,
          column: 14,
          endLine: 4,
          endColumn: 53,
        },
      ],
    },
    ...(
      [
        [
          `describe('foo', () => {
  it('works', () => {});
  it('works', () => {});
});`,
          3,
          6,
          13,
        ],
        [
          `it('works', () => {});
it('works', () => {});`,
          2,
          4,
          11,
        ],
        [
          `test.only('this', () => {});
test('this', () => {});`,
          2,
          6,
          12,
        ],
        [
          `xtest('this', () => {});
test('this', () => {});`,
          2,
          6,
          12,
        ],
        [
          `test.only('this', () => {});
test.only('this', () => {});`,
          2,
          11,
          17,
        ],
        [
          `test.concurrent('this', () => {});
test.concurrent('this', () => {});`,
          2,
          17,
          23,
        ],
        [
          `test.only('this', () => {});
test.concurrent('this', () => {});`,
          2,
          17,
          23,
        ],
      ] as const
    ).map(([code, line, column, endColumn], index): TestCase => ({
      name: `upstream no-identical-title / invalid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'multipleTestTitle',
          message:
            'Test title is used multiple times in the same describe block',
          line,
          column,
          endLine: line,
          endColumn,
        },
      ],
    })),
    ...(
      [
        [
          `describe('foo', () => {});
describe('foo', () => {});`,
          2,
          10,
        ],
        [
          `describe('foo', () => {});
xdescribe('foo', () => {});`,
          2,
          11,
        ],
        [
          `fdescribe('foo', () => {});
describe('foo', () => {});`,
          2,
          10,
        ],
        [
          `describe('foo', () => {
  describe('foe', () => {});
});
describe('foo', () => {});`,
          4,
          10,
        ],
      ] as const
    ).map(([code, line, column], index): TestCase => ({
      name: `upstream no-identical-title / invalid ${index + 7}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'multipleDescribeTitle',
          message:
            'Describe block title is used multiple times in the same describe block',
          line,
          column,
          endLine: line,
          endColumn: column + 5,
        },
      ],
    })),
    {
      name: 'upstream no-identical-title / invalid 11',
      code: `describe("foo", () => {
  it(\`catches backticks with the same title\`, () => {});
  it(\`catches backticks with the same title\`, () => {});
});`,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'multipleTestTitle',
          message:
            'Test title is used multiple times in the same describe block',
          line: 3,
          column: 6,
          endLine: 3,
          endColumn: 45,
        },
      ],
    },
    {
      name: 'upstream no-identical-title / invalid 12',
      code: `context('foo', () => {
  describe('foe', () => {});
});
describe('foo', () => {});`,
      settings: {
        jest: {
          globalAliases: {
            describe: ['context'],
          },
        },
      },
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'multipleDescribeTitle',
          message:
            'Describe block title is used multiple times in the same describe block',
          line: 4,
          column: 10,
          endLine: 4,
          endColumn: 15,
        },
      ],
    },
  ],
});

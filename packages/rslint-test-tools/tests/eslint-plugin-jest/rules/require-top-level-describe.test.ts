// Upstream reference: eslint-plugin-jest@29.16.0; includes existing Rslint regressions.
// Additional cases from 29.16.6 retain their input and parser configuration.
// Complete assertions record current native output; documented differences remain separate.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

new RuleTester(fixtureConfig()).run('require-top-level-describe', {
  valid: [
    ...(
      [
        'it.each()',
        'describe("test suite", () => { test("my test") });',
        'describe("test suite", () => { it("my test") });',
        `
      describe("test suite", () => {
        beforeEach("a", () => {});
        describe("b", () => {});
        test("c", () => {})
      });
   \u0020`,
        'describe("test suite", () => { beforeAll("my beforeAll") });',
        'describe("test suite", () => { afterEach("my afterEach") });',
        'describe("test suite", () => { afterAll("my afterAll") });',
        `
      describe("test suite", () => {
        it("my test", () => {})
        describe("another test suite", () => {
        });
        test("my other test", () => {})
      });
   \u0020`,
        'foo()',
        'describe.each([1, true])("trues", value => { it("an it", () => expect(value).toBe(true) ); });',
        `
      describe('%s', () => {
        it('is fine', () => {
          //
        });
      });

      describe.each('world')('%s', () => {
        it.each([1, 2, 3])('%n', () => {
          //
        });
      });
   \u0020`,
        `
      describe.each('hello')('%s', () => {
        it('is fine', () => {
          //
        });
      });

      describe.each('world')('%s', () => {
        it.each([1, 2, 3])('%n', () => {
          //
        });
      });
   \u0020`,
        `
        import { jest } from '@jest/globals';

        jest.doMock('my-module');
     \u0020`,
        'jest.doMock("my-module")',
        `
        describe('one', () => {});
        describe('two', () => {});
        describe('three', () => {});
     \u0020`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `run 0 valid ${index}`,
      code,
    })),
    {
      name: 'run 0 valid 15',
      code: `
          describe('one', () => {
            describe('two', () => {});
            describe('three', () => {});
          });
       \u0020`,
      options: [
        {
          maxNumberOfTopLevelDescribes: 1,
        },
      ],
    },
    ...(
      [
        'it.each()',
        'describe("test suite", () => { test("my test") });',
        'describe("test suite", () => { it("my test") });',
        `describe("test suite", () => {
  beforeEach("a", () => {});
  describe("b", () => {});
  test("c", () => {})
});`,
        'describe("test suite", () => { beforeAll("my beforeAll") });',
        'describe("test suite", () => { afterEach("my afterEach") });',
        'describe("test suite", () => { afterAll("my afterAll") });',
        `describe("test suite", () => {
  it("my test", () => {})
  describe("another test suite", () => {
  });
  test("my other test", () => {})
});`,
        'foo()',
        'describe.each([1, true])("trues", value => { it("an it", () => expect(value).toBe(true) ); });',
        `describe('%s', () => {
  it('is fine', () => {
    //
  });
});

describe.each('world')('%s', () => {
  it.each([1, 2, 3])('%n', () => {
    //
  });
});`,
        `describe.each('hello')('%s', () => {
  it('is fine', () => {
    //
  });
});

describe.each('world')('%s', () => {
  it.each([1, 2, 3])('%n', () => {
    //
  });
});`,
      ] as const
    ).map((code, index): ValidTestCase => ({
      name: `upstream require-top-level-describe / valid ${index}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
    })),
    {
      name: 'upstream require-top-level-describe / valid 12',
      code: `import { jest } from '@jest/globals';

jest.doMock('my-module');`,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2015,
      },
    },
    {
      name: 'upstream require-top-level-describe / valid 13',
      code: 'jest.doMock("my-module")',
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
    {
      name: 'upstream require-top-level-describe (enforce number of describes) / valid 3',
      code: `describe('one', () => {});
describe('two', () => {});
describe('three', () => {});`,
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
    {
      name: 'upstream require-top-level-describe (enforce number of describes) / valid 4',
      code: `describe('one', () => {
  describe('two', () => {});
  describe('three', () => {});
});`,
      options: [
        {
          maxNumberOfTopLevelDescribes: 1,
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: 'beforeEach("my test", () => {})',
      errors: [
        {
          messageId: 'unexpectedHook',
          message: 'All hooks must be wrapped in a describe block.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    ...(
      [
        `
        test("my test", () => {})
        describe("test suite", () => {});
     \u0020`,
        `
        test("my test", () => {})
        describe("test suite", () => {
          it("test", () => {})
        });
     \u0020`,
      ] as const
    ).map((code, index): TestCase => ({
      name: `run 0 invalid ${index + 1}`,
      code,
      errors: [
        {
          messageId: 'unexpectedTestCase',
          message: 'All test cases must be wrapped in a describe block.',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 34,
        },
      ],
    })),
    ...(
      [
        [
          `
        describe("test suite", () => {});
        afterAll("my test", () => {})
     \u0020`,
          3,
          38,
        ],
        [
          `
        import { describe, afterAll as onceEverythingIsDone } from '@jest/globals';

        describe("test suite", () => {});
        onceEverythingIsDone("my test", () => {})
     \u0020`,
          5,
          50,
        ],
        [
          `
        import { 'describe' as describe, afterAll as onceEverythingIsDone } from '@jest/globals';

        describe("test suite", () => {});
        onceEverythingIsDone("my test", () => {})
     \u0020`,
          5,
          50,
        ],
      ] as const
    ).map(([code, line, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index + 3}`,
      code,
      errors: [
        {
          messageId: 'unexpectedHook',
          message: 'All hooks must be wrapped in a describe block.',
          line,
          column: 9,
          endLine: line,
          endColumn,
        },
      ],
    })),
    ...(
      [
        ["it.skip('test', () => {});", 26],
        ["it.each([1, 2, 3])('%n', () => {});", 35],
        ["it.skip.each([1, 2, 3])('%n', () => {});", 40],
        ["it.skip.each``('%n', () => {});", 31],
        ["it.each``('%n', () => {});", 26],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index + 6}`,
      code,
      errors: [
        {
          messageId: 'unexpectedTestCase',
          message: 'All test cases must be wrapped in a describe block.',
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
          `
          describe('one', () => {});
          describe('two', () => {});
          describe('three', () => {});
       \u0020`,
          4,
          4,
          38,
        ],
        [
          `
          describe('one', () => {
            describe('one (nested)', () => {});
            describe('two (nested)', () => {});
          });
          describe('two', () => {
            describe('one (nested)', () => {});
            describe('two (nested)', () => {});
            describe('three (nested)', () => {});
          });
          describe('three', () => {
            describe('one (nested)', () => {});
            describe('two (nested)', () => {});
            describe('three (nested)', () => {});
          });
       \u0020`,
          11,
          15,
          13,
        ],
        [
          `
          import {
            describe as describe1,
            describe as describe2,
            describe as describe3,
          } from '@jest/globals';

          describe1('one', () => {
            describe('one (nested)', () => {});
            describe('two (nested)', () => {});
          });
          describe2('two', () => {
            describe('one (nested)', () => {});
            describe('two (nested)', () => {});
            describe('three (nested)', () => {});
          });
          describe3('three', () => {
            describe('one (nested)', () => {});
            describe('two (nested)', () => {});
            describe('three (nested)', () => {});
          });
       \u0020`,
          17,
          21,
          13,
        ],
      ] as const
    ).map(([code, line, endLine, endColumn], index): TestCase => ({
      name: `run 0 invalid ${index + 11}`,
      code,
      options: [
        {
          maxNumberOfTopLevelDescribes: 2,
        },
      ],
      errors: [
        {
          messageId: 'tooManyDescribes',
          message: 'There should not be more than 2 describes at the top level',
          line,
          column: 11,
          endLine,
          endColumn,
        },
      ],
    })),
    {
      name: 'run 0 invalid 14',
      code: `
          describe('one', () => {});
          describe('two', () => {});
          describe('three', () => {});
       \u0020`,
      options: [
        {
          maxNumberOfTopLevelDescribes: 1,
        },
      ],
      errors: [
        {
          messageId: 'tooManyDescribes',
          message: 'There should not be more than 1 describe at the top level',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 36,
        },
        {
          messageId: 'tooManyDescribes',
          message: 'There should not be more than 1 describe at the top level',
          line: 4,
          column: 11,
          endLine: 4,
          endColumn: 38,
        },
      ],
    },
    {
      name: 'upstream require-top-level-describe / invalid 0',
      code: 'beforeEach("my test", () => {})',
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'unexpectedHook',
          message: 'All hooks must be wrapped in a describe block.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    ...(
      [
        `test("my test", () => {})
describe("test suite", () => {});`,
        `test("my test", () => {})
describe("test suite", () => {
  it("test", () => {})
});`,
      ] as const
    ).map((code, index): TestCase => ({
      name: `upstream require-top-level-describe / invalid ${index + 1}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'unexpectedTestCase',
          message: 'All test cases must be wrapped in a describe block.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    })),
    {
      name: 'upstream require-top-level-describe / invalid 3',
      code: `describe("test suite", () => {});
afterAll("my test", () => {})`,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'unexpectedHook',
          message: 'All hooks must be wrapped in a describe block.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 30,
        },
      ],
    },
    {
      name: 'upstream require-top-level-describe / invalid 4',
      code: `import { describe, afterAll as onceEverythingIsDone } from '@jest/globals';

describe("test suite", () => {});
onceEverythingIsDone("my test", () => {})`,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'unexpectedHook',
          message: 'All hooks must be wrapped in a describe block.',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 42,
        },
      ],
    },
    {
      name: 'upstream require-top-level-describe / invalid 5',
      code: `import { 'describe' as describe, afterAll as onceEverythingIsDone } from '@jest/globals';

describe("test suite", () => {});
onceEverythingIsDone("my test", () => {})`,
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2022,
      },
      errors: [
        {
          messageId: 'unexpectedHook',
          message: 'All hooks must be wrapped in a describe block.',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 42,
        },
      ],
    },
    ...(
      [
        ["it.skip('test', () => {});", 26],
        ["it.each([1, 2, 3])('%n', () => {});", 35],
        ["it.skip.each([1, 2, 3])('%n', () => {});", 40],
        ["it.skip.each``('%n', () => {});", 31],
        ["it.each``('%n', () => {});", 26],
      ] as const
    ).map(([code, endColumn], index): TestCase => ({
      name: `upstream require-top-level-describe / invalid ${index + 6}`,
      code,
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'unexpectedTestCase',
          message: 'All test cases must be wrapped in a describe block.',
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
          `describe('one', () => {});
describe('two', () => {});
describe('three', () => {});`,
          3,
          3,
          28,
        ],
        [
          `describe('one', () => {
  describe('one (nested)', () => {});
  describe('two (nested)', () => {});
});
describe('two', () => {
  describe('one (nested)', () => {});
  describe('two (nested)', () => {});
  describe('three (nested)', () => {});
});
describe('three', () => {
  describe('one (nested)', () => {});
  describe('two (nested)', () => {});
  describe('three (nested)', () => {});
});`,
          10,
          14,
          3,
        ],
      ] as const
    ).map(([code, line, endLine, endColumn], index): TestCase => ({
      name: `upstream require-top-level-describe (enforce number of describes) / invalid ${index}`,
      code,
      options: [
        {
          maxNumberOfTopLevelDescribes: 2,
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'tooManyDescribes',
          message: 'There should not be more than 2 describes at the top level',
          line,
          column: 1,
          endLine,
          endColumn,
        },
      ],
    })),
    {
      name: 'upstream require-top-level-describe (enforce number of describes) / invalid 2',
      code: `import {
  describe as describe1,
  describe as describe2,
  describe as describe3,
} from '@jest/globals';

describe1('one', () => {
  describe('one (nested)', () => {});
  describe('two (nested)', () => {});
});
describe2('two', () => {
  describe('one (nested)', () => {});
  describe('two (nested)', () => {});
  describe('three (nested)', () => {});
});
describe3('three', () => {
  describe('one (nested)', () => {});
  describe('two (nested)', () => {});
  describe('three (nested)', () => {});
});`,
      options: [
        {
          maxNumberOfTopLevelDescribes: 2,
        },
      ],
      languageOptions: {
        sourceType: 'module',
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'tooManyDescribes',
          message: 'There should not be more than 2 describes at the top level',
          line: 16,
          column: 1,
          endLine: 20,
          endColumn: 3,
        },
      ],
    },
    {
      name: 'upstream require-top-level-describe (enforce number of describes) / invalid 3',
      code: `describe('one', () => {});
describe('two', () => {});
describe('three', () => {});`,
      options: [
        {
          maxNumberOfTopLevelDescribes: 1,
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
      },
      errors: [
        {
          messageId: 'tooManyDescribes',
          message: 'There should not be more than 1 describe at the top level',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 26,
        },
        {
          messageId: 'tooManyDescribes',
          message: 'There should not be more than 1 describe at the top level',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 28,
        },
      ],
    },
  ],
});

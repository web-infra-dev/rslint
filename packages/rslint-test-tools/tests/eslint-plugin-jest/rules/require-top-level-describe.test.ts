import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('require-top-level-describe', {} as never, {
  valid: [
    { code: 'it.each()' },
    { code: 'describe("test suite", () => { test("my test") });' },
    { code: 'describe("test suite", () => { it("my test") });' },
    {
      code: `
      describe("test suite", () => {
        beforeEach("a", () => {});
        describe("b", () => {});
        test("c", () => {})
      });
    `,
    },
    { code: 'describe("test suite", () => { beforeAll("my beforeAll") });' },
    { code: 'describe("test suite", () => { afterEach("my afterEach") });' },
    { code: 'describe("test suite", () => { afterAll("my afterAll") });' },
    {
      code: `
      describe("test suite", () => {
        it("my test", () => {})
        describe("another test suite", () => {
        });
        test("my other test", () => {})
      });
    `,
    },
    { code: 'foo()' },
    {
      code: 'describe.each([1, true])("trues", value => { it("an it", () => expect(value).toBe(true) ); });',
    },
    {
      code: `
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
    `,
    },
    {
      code: `
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
    `,
    },
    {
      code: `
        import { jest } from '@jest/globals';

        jest.doMock('my-module');
      `,
    },
    { code: 'jest.doMock("my-module")' },
    {
      code: `
        describe('one', () => {});
        describe('two', () => {});
        describe('three', () => {});
      `,
    },
    {
      code: `
          describe('one', () => {
            describe('two', () => {});
            describe('three', () => {});
          });
        `,
      options: [{ maxNumberOfTopLevelDescribes: 1 }],
    },
  ],
  invalid: [
    {
      code: 'beforeEach("my test", () => {})',
      errors: [{ messageId: 'unexpectedHook' }],
    },
    {
      code: `
        test("my test", () => {})
        describe("test suite", () => {});
      `,
      errors: [{ messageId: 'unexpectedTestCase' }],
    },
    {
      code: `
        test("my test", () => {})
        describe("test suite", () => {
          it("test", () => {})
        });
      `,
      errors: [{ messageId: 'unexpectedTestCase' }],
    },
    {
      code: `
        describe("test suite", () => {});
        afterAll("my test", () => {})
      `,
      errors: [{ messageId: 'unexpectedHook' }],
    },
    {
      code: `
        import { describe, afterAll as onceEverythingIsDone } from '@jest/globals';

        describe("test suite", () => {});
        onceEverythingIsDone("my test", () => {})
      `,
      errors: [{ messageId: 'unexpectedHook' }],
    },
    {
      code: `
        import { 'describe' as describe, afterAll as onceEverythingIsDone } from '@jest/globals';

        describe("test suite", () => {});
        onceEverythingIsDone("my test", () => {})
      `,
      errors: [{ messageId: 'unexpectedHook' }],
    },
    {
      code: "it.skip('test', () => {});",
      errors: [{ messageId: 'unexpectedTestCase' }],
    },
    {
      code: "it.each([1, 2, 3])('%n', () => {});",
      errors: [{ messageId: 'unexpectedTestCase' }],
    },
    {
      code: "it.skip.each([1, 2, 3])('%n', () => {});",
      errors: [{ messageId: 'unexpectedTestCase' }],
    },
    {
      code: "it.skip.each``('%n', () => {});",
      errors: [{ messageId: 'unexpectedTestCase' }],
    },
    {
      code: "it.each``('%n', () => {});",
      errors: [{ messageId: 'unexpectedTestCase' }],
    },
    {
      code: `
          describe('one', () => {});
          describe('two', () => {});
          describe('three', () => {});
        `,
      options: [{ maxNumberOfTopLevelDescribes: 2 }],
      errors: [{ messageId: 'tooManyDescribes', line: 3 }],
    },
    {
      code: `
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
        `,
      options: [{ maxNumberOfTopLevelDescribes: 2 }],
      errors: [{ messageId: 'tooManyDescribes', line: 10 }],
    },
    {
      code: `
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
        `,
      options: [{ maxNumberOfTopLevelDescribes: 2 }],
      errors: [{ messageId: 'tooManyDescribes', line: 16 }],
    },
    {
      code: `
          describe('one', () => {});
          describe('two', () => {});
          describe('three', () => {});
        `,
      options: [{ maxNumberOfTopLevelDescribes: 1 }],
      errors: [
        { messageId: 'tooManyDescribes', line: 2 },
        { messageId: 'tooManyDescribes', line: 3 },
      ],
    },
  ],
});

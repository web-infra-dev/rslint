import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-standalone-expect', {} as never, {
  valid: [
    { code: 'expect.any(String)' },
    { code: 'expect.extend({})' },
    {
      code: 'describe("a test", () => { it("an it", () => {expect(1).toBe(1); }); });',
    },
    {
      code: 'describe("a test", () => { it("an it", () => { const func = () => { expect(1).toBe(1); }; }); });',
    },
    {
      code: 'describe("a test", () => { const func = () => { expect(1).toBe(1); }; });',
    },
    {
      code: 'describe("a test", () => { function func() { expect(1).toBe(1); }; });',
    },
    {
      code: 'describe("a test", () => { const func = function(){ expect(1).toBe(1); }; });',
    },
    { code: 'it("an it", () => expect(1).toBe(1))' },
    { code: 'const func = function(){ expect(1).toBe(1); };' },
    { code: 'const func = () => expect(1).toBe(1);' },
    { code: '{}' },
    {
      code: 'it.each([1, true])("trues", value => { expect(value).toBe(true); });',
    },
    {
      code: 'it.each([1, true])("trues", value => { expect(value).toBe(true); }); it("an it", () => { expect(1).toBe(1) });',
    },
    {
      code: `
      it.each\`
        num   | value
        \${1} | \${true}
      \`('trues', ({ value }) => {
        expect(value).toBe(true);
      });
    `,
    },
    { code: 'it.only("an only", value => { expect(value).toBe(true); });' },
    {
      code: 'it.concurrent("an concurrent", value => { expect(value).toBe(true); });',
    },
    {
      code: 'describe.each([1, true])("trues", value => { it("an it", () => expect(value).toBe(true) ); });',
    },
    {
      code: 'describe.each([1, true])("trues", value => { it("an it", () => expect(value).toBe(true) ); });',
    },
    {
      code: 'const helpers = { assert() { expect(1).toBe(1); } };',
    },
    {
      code: 'class Helper { assert() { expect(1).toBe(1); } get value() { expect(1).toBe(1); return 1; } }',
    },
    {
      code: `
        describe('scenario', () => {
          const t = Math.random() ? it.only : it;
          t('testing', () => expect(true));
        });
      `,
      options: { additionalTestBlockFunctions: ['t'] },
    },
    {
      code: `
        each([
          [1, 1, 2],
          [1, 2, 3],
          [2, 1, 3],
        ]).test('returns the result of adding %d to %d', (a, b, expected) => {
          expect(a + b).toBe(expected);
        });
      `,
      options: { additionalTestBlockFunctions: ['each.test'] },
    },
  ],
  invalid: [
    {
      code: "(() => {})('testing', () => expect(true).toBe(false))",
      errors: [{ endColumn: 53, column: 29, messageId: 'unexpectedExpect' }],
    },
    {
      code: 'expect.hasAssertions()',
      errors: [{ endColumn: 23, column: 1, messageId: 'unexpectedExpect' }],
    },
    {
      code: 'expect().hasAssertions()',
      errors: [{ endColumn: 25, column: 1, messageId: 'unexpectedExpect' }],
    },
    {
      code: `
        describe('scenario', () => {
          const t = Math.random() ? it.only : it;
          t('testing', () => expect(true).toBe(false));
        });
      `,
      errors: [{ endColumn: 46, column: 22, messageId: 'unexpectedExpect' }],
    },
    {
      code: `
        describe('scenario', () => {
          const t = Math.random() ? it.only : it;
          t('testing', () => expect(true).toBe(false));
        });
      `,
      options: { additionalTestBlockFunctions: undefined },
      errors: [{ endColumn: 46, column: 22, messageId: 'unexpectedExpect' }],
    },
    {
      code: `
        each([
          [1, 1, 2],
          [1, 2, 3],
          [2, 1, 3],
        ]).test('returns the result of adding %d to %d', (a, b, expected) => {
          expect(a + b).toBe(expected);
        });
      `,
      errors: [{ endColumn: 31, column: 3, messageId: 'unexpectedExpect' }],
    },
    {
      code: `
        each([
          [1, 1, 2],
          [1, 2, 3],
          [2, 1, 3],
        ]).test('returns the result of adding %d to %d', (a, b, expected) => {
          expect(a + b).toBe(expected);
        });
      `,
      options: { additionalTestBlockFunctions: ['each'] },
      errors: [{ endColumn: 31, column: 3, messageId: 'unexpectedExpect' }],
    },
    {
      code: `
        each([
          [1, 1, 2],
          [1, 2, 3],
          [2, 1, 3],
        ]).test('returns the result of adding %d to %d', (a, b, expected) => {
          expect(a + b).toBe(expected);
        });
      `,
      options: { additionalTestBlockFunctions: ['test'] },
      errors: [{ endColumn: 31, column: 3, messageId: 'unexpectedExpect' }],
    },
    {
      code: 'describe("a test", () => { expect(1).toBe(1); });',
      errors: [{ endColumn: 45, column: 28, messageId: 'unexpectedExpect' }],
    },
    {
      code: 'describe("a test", () => expect(1).toBe(1));',
      errors: [{ endColumn: 43, column: 26, messageId: 'unexpectedExpect' }],
    },
    {
      code: 'describe("a test", () => { const func = () => { expect(1).toBe(1); }; expect(1).toBe(1); });',
      errors: [{ endColumn: 88, column: 71, messageId: 'unexpectedExpect' }],
    },
    {
      code: 'describe("a test", () => {  it(() => { expect(1).toBe(1); }); expect(1).toBe(1); });',
      errors: [{ endColumn: 80, column: 63, messageId: 'unexpectedExpect' }],
    },
    {
      code: 'expect(1).toBe(1);',
      errors: [{ endColumn: 18, column: 1, messageId: 'unexpectedExpect' }],
    },
    {
      code: '{expect(1).toBe(1)}',
      errors: [{ endColumn: 19, column: 2, messageId: 'unexpectedExpect' }],
    },
    {
      code: 'it.each([1, true])("trues", value => { expect(value).toBe(true); }); expect(1).toBe(1);',
      errors: [{ endColumn: 87, column: 70, messageId: 'unexpectedExpect' }],
    },
    {
      code: 'describe.each([1, true])("trues", value => { expect(value).toBe(true); });',
      errors: [{ endColumn: 70, column: 46, messageId: 'unexpectedExpect' }],
    },
    {
      code: `
        describe.each\`
          num   | value
          \${1} | \${true}
        \`('trues', ({ value }) => {
          expect(value).toBe(true);
        });
      `,
      errors: [{ endColumn: 27, column: 3, messageId: 'unexpectedExpect' }],
    },
    {
      code: `
        it.each\`
          num   | value
          \${1} | \${true}
        \`('trues', ({ value }) => {
          expect(value).toBe(true);
        });

        expect(1).toBe(1);
      `,
      errors: [{ endColumn: 17, column: 1, messageId: 'unexpectedExpect' }],
    },
    {
      code: `
        import { expect as pleaseExpect } from '@jest/globals';

        describe("a test", () => { pleaseExpect(1).toBe(1); });
      `,
      errors: [{ endColumn: 51, column: 28, messageId: 'unexpectedExpect' }],
    },
    {
      code: `
        import { expect as pleaseExpect } from '@jest/globals';

        beforeEach(() => pleaseExpect.hasAssertions());
      `,
      errors: [{ endColumn: 46, column: 18, messageId: 'unexpectedExpect' }],
    },
  ],
});

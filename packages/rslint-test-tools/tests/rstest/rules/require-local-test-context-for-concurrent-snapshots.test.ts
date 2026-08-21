import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run(
  'require-local-test-context-for-concurrent-snapshots',
  {} as never,
  {
    valid: [
      { code: `it('x', () => expect(1).toMatchSnapshot())` },
      { code: `it.concurrent('x', () => expect(true).toBe(true))` },
      {
        code: `it.concurrent('x', () => expect('a').to.be.a('string').and.have.length(1))`,
      },
      {
        code: `it.concurrent('x', ({ expect }) => expect(1).toMatchSnapshot())`,
      },
      {
        code: `it.concurrent('x', ctx => ctx.expect(1).toMatchInlineSnapshot())`,
      },
      {
        code: `describe.concurrent('s', () => it.sequential('x', () => expect(1).toMatchSnapshot()))`,
      },
      {
        code: `test.concurrent.for(rows)('x', (row, ctx) => ctx.expect(row).toMatchSnapshot())`,
      },
      {
        code: `import { test, expect } from 'vitest'; test.concurrent('x', () => expect(1).toMatchSnapshot())`,
      },
      {
        code: `function helper() { expect(1).toMatchSnapshot() } test.concurrent('x', () => helper())`,
      },
      {
        code: `test('x', { concurrent: true }, () => expect(1).toMatchSnapshot())`,
      },
      {
        code: `describe('s', () => { function cb() { expect(1).toMatchSnapshot() } }); test.concurrent('x', cb)`,
      },
    ],
    invalid: [
      {
        code: `it.concurrent('x', () => expect(true).toMatchSnapshot())`,
        errors: [{ messageId: 'requireLocalTestContext' }],
      },
      {
        code: `describe.concurrent('s', () => it('x', () => expect(true).toMatchInlineSnapshot()))`,
        errors: [{ messageId: 'requireLocalTestContext' }],
      },
      {
        code: `const t = test.concurrent; t('x', () => expect(1).toMatchSnapshot())`,
        errors: [{ messageId: 'requireLocalTestContext' }],
      },
      {
        code: `test.concurrent('x', callback); function callback() { expect(1).toMatchSnapshot() }`,
        errors: [{ messageId: 'requireLocalTestContext' }],
      },
      {
        code: `describe.concurrent('s', suite); function suite() { test('x', callback) } function callback() { expect(1).toMatchSnapshot() }`,
        errors: [{ messageId: 'requireLocalTestContext' }],
      },
      {
        code: `test.concurrent('x', { concurrent: false }, () => expect(1).toMatchSnapshot())`,
        errors: [{ messageId: 'requireLocalTestContext' }],
      },
      {
        code: `test.concurrent('x', () => expect(1).matchSnapshot())`,
        errors: [{ messageId: 'requireLocalTestContext' }],
      },
      {
        code: `test.concurrent('x', () => expect('a').to.be.a('string').and.matchSnapshot())`,
        errors: [{ messageId: 'requireLocalTestContext' }],
      },
      {
        code: `it.concurrent('x', async () => { await Promise.all(items.map(async item => { expect(item).toMatchSnapshot() })) })`,
        errors: [{ messageId: 'requireLocalTestContext' }],
      },
    ],
  },
);

import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('require-local-test-context-for-concurrent-snapshots', {
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
      errors: [
        {
          messageId: 'requireLocalTestContext',
          message: 'Use local Test Context instead',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 56,
        },
      ],
    },
    {
      code: `describe.concurrent('s', () => it('x', () => expect(true).toMatchInlineSnapshot()))`,
      errors: [
        {
          messageId: 'requireLocalTestContext',
          message: 'Use local Test Context instead',
          line: 1,
          column: 46,
          endLine: 1,
          endColumn: 82,
        },
      ],
    },
    {
      code: `const t = test.concurrent; t('x', () => expect(1).toMatchSnapshot())`,
      errors: [
        {
          messageId: 'requireLocalTestContext',
          message: 'Use local Test Context instead',
          line: 1,
          column: 41,
          endLine: 1,
          endColumn: 68,
        },
      ],
    },
    {
      code: `test.concurrent('x', callback); function callback() { expect(1).toMatchSnapshot() }`,
      errors: [
        {
          messageId: 'requireLocalTestContext',
          message: 'Use local Test Context instead',
          line: 1,
          column: 55,
          endLine: 1,
          endColumn: 82,
        },
      ],
    },
    {
      code: `describe.concurrent('s', suite); function suite() { test('x', callback) } function callback() { expect(1).toMatchSnapshot() }`,
      errors: [
        {
          messageId: 'requireLocalTestContext',
          message: 'Use local Test Context instead',
          line: 1,
          column: 97,
          endLine: 1,
          endColumn: 124,
        },
      ],
    },
    {
      code: `test.concurrent('x', { concurrent: false }, () => expect(1).toMatchSnapshot())`,
      errors: [
        {
          messageId: 'requireLocalTestContext',
          message: 'Use local Test Context instead',
          line: 1,
          column: 51,
          endLine: 1,
          endColumn: 78,
        },
      ],
    },
    {
      code: `test.concurrent('x', () => expect(1).matchSnapshot())`,
      errors: [
        {
          messageId: 'requireLocalTestContext',
          message: 'Use local Test Context instead',
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
    {
      code: `test.concurrent('x', () => expect('a').to.be.a('string').and.matchSnapshot())`,
      errors: [
        {
          messageId: 'requireLocalTestContext',
          message: 'Use local Test Context instead',
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 77,
        },
      ],
    },
    {
      code: `it.concurrent('x', async () => { await Promise.all(items.map(async item => { expect(item).toMatchSnapshot() })) })`,
      errors: [
        {
          messageId: 'requireLocalTestContext',
          message: 'Use local Test Context instead',
          line: 1,
          column: 78,
          endLine: 1,
          endColumn: 108,
        },
      ],
    },
  ],
});

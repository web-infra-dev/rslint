import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('no-async-mock-factory', {
  valid: [
    { code: `rs.mock('./sum')` },
    { code: `rs.mock('./sum', { spy: true })` },
    { code: `rs.mock('./sum', () => ({ sum: () => 0 }))` },
    { code: `rs.mock('./sum', function* () { yield 1; })` },
    { code: `rs.mock('./sum', async function* () { yield 1; })` },
    { code: `rs.mock('./sum', () => ({ then() {} }))` },
    { code: `rs['mock']('./sum', async () => ({ sum: 0 }))` },
    { code: `rs?.mock('./sum', async () => ({ sum: 0 }))` },
    { code: `rs.mock?.('./sum', async () => ({ sum: 0 }))` },
    { code: `const mocked = rs.mock('./sum', async () => ({ sum: 0 }))` },
    { code: `rs.mock('./sum', async () => ({ sum: 0 }), extra)` },
    { code: `rs.mock('./sum', ...factories)` },
    {
      code: `import { rs as mocker } from '@rstest/core';\nmocker.mock('./sum', async () => ({ sum: 0 }))`,
    },
  ],
  invalid: [
    {
      code: `rs.mock('./sum', async () => ({ sum: () => 0 }))`,
      errors: [
        {
          messageId: 'asyncMockFactory',
          message:
            "Mock factory must return the module object synchronously; 'mock()' throws when the factory returns a Promise. To keep part of the original module, import it with `with { rstest: 'importActual' }` and spread it in.",
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 48,
          suggestions: [
            {
              messageId: 'suggestRemoveAsync',
              desc: "Remove 'async' so the factory returns the module object directly.",
              output: "rs.mock('./sum', () => ({ sum: () => 0 }))",
            },
          ],
        },
      ],
    },
    {
      code: `rstest.doMock('./sum', () => Promise.resolve({ sum: 0 }))`,
      errors: [
        {
          messageId: 'asyncMockFactory',
          message:
            "Mock factory must return the module object synchronously; 'doMock()' throws when the factory returns a Promise. To keep part of the original module, import it with `with { rstest: 'importActual' }` and spread it in.",
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 57,
          suggestions: [
            {
              messageId: 'suggestUnwrapPromiseResolve',
              desc: "Return the module object directly instead of wrapping it in 'Promise.resolve()'.",
              output: "rstest.doMock('./sum', () => ({ sum: 0 }))",
            },
          ],
        },
      ],
    },
    {
      code: `rs.mockRequire('./sum', () => import('./sum'))`,
      errors: [
        {
          messageId: 'asyncMockFactory',
          message:
            "Mock factory must return the module object synchronously; 'mockRequire()' throws when the factory returns a Promise. To keep part of the original module, import it with `with { rstest: 'importActual' }` and spread it in.",
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 46,
        },
      ],
    },
    {
      code: `rs.doMockRequire('./sum', async function () { return { sum: 0 }; })`,
      errors: [
        {
          messageId: 'asyncMockFactory',
          message:
            "Mock factory must return the module object synchronously; 'doMockRequire()' throws when the factory returns a Promise. To keep part of the original module, import it with `with { rstest: 'importActual' }` and spread it in.",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 67,
          suggestions: [
            {
              messageId: 'suggestRemoveAsync',
              desc: "Remove 'async' so the factory returns the module object directly.",
              output:
                "rs.doMockRequire('./sum', function () { return { sum: 0 }; })",
            },
          ],
        },
      ],
    },
    {
      code: `rs.mock<{ sum: number }>('./sum', async () => ({ sum: 0 }))`,
      errors: [
        {
          messageId: 'asyncMockFactory',
          message:
            "Mock factory must return the module object synchronously; 'mock()' throws when the factory returns a Promise. To keep part of the original module, import it with `with { rstest: 'importActual' }` and spread it in.",
          line: 1,
          column: 35,
          endLine: 1,
          endColumn: 59,
          suggestions: [
            {
              messageId: 'suggestRemoveAsync',
              desc: "Remove 'async' so the factory returns the module object directly.",
              output: "rs.mock<{ sum: number }>('./sum', () => ({ sum: 0 }))",
            },
          ],
        },
      ],
    },
  ],
});

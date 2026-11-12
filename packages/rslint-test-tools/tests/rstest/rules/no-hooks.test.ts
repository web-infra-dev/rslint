import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('no-hooks', {
  valid: [
    { code: 'test("foo")' },
    { code: 'describe("foo", () => { it("bar") })' },
    {
      code: 'test("foo", () => { expect(subject.beforeEach()).toBe(true) })',
    },
    {
      code: 'afterEach(() => {}); afterAll(() => {});',
      options: [{ allow: ['afterEach', 'afterAll'] }],
    },
    { code: 'test("foo")', options: [{}] },
  ],
  invalid: [
    {
      code: 'beforeAll(() => {})',
      errors: [
        {
          messageId: 'unexpectedHook',
          message: "Unexpected 'beforeAll' hook",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: 'beforeEach(() => {})',
      errors: [
        {
          messageId: 'unexpectedHook',
          message: "Unexpected 'beforeEach' hook",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: 'afterAll(() => {})',
      errors: [
        {
          messageId: 'unexpectedHook',
          message: "Unexpected 'afterAll' hook",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'afterEach(() => {})',
      errors: [
        {
          messageId: 'unexpectedHook',
          message: "Unexpected 'afterEach' hook",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: [
        "import { 'afterEach' as afterEachTest } from '@rstest/core';",
        '',
        'afterEachTest(() => {})',
      ].join('\n'),
      errors: [
        {
          messageId: 'unexpectedHook',
          message: "Unexpected 'afterEach' hook",
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'beforeEach(() => {}); afterEach(() => { resetModules() });',
      options: [{ allow: ['afterEach'] }],
      errors: [
        {
          messageId: 'unexpectedHook',
          message: "Unexpected 'beforeEach' hook",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: [
        "import { beforeEach as afterEach, afterEach as beforeEach } from '@rstest/core';",
        '',
        'afterEach(() => {});',
        'beforeEach(() => { resetModules() });',
      ].join('\n'),
      options: [{ allow: ['afterEach'] }],
      errors: [
        {
          messageId: 'unexpectedHook',
          message: "Unexpected 'beforeEach' hook",
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 20,
        },
      ],
    },
  ],
});

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-hooks', {} as never, {
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
      errors: [{ message: "Unexpected 'beforeAll' hook" }],
    },
    {
      code: 'beforeEach(() => {})',
      errors: [{ message: "Unexpected 'beforeEach' hook" }],
    },
    {
      code: 'afterAll(() => {})',
      errors: [{ message: "Unexpected 'afterAll' hook" }],
    },
    {
      code: 'afterEach(() => {})',
      errors: [{ message: "Unexpected 'afterEach' hook" }],
    },
    {
      code: [
        "import { 'afterEach' as afterEachTest } from '@rstest/core';",
        '',
        'afterEachTest(() => {})',
      ].join('\n'),
      errors: [{ message: "Unexpected 'afterEach' hook" }],
    },
    {
      code: 'beforeEach(() => {}); afterEach(() => { resetModules() });',
      options: [{ allow: ['afterEach'] }],
      errors: [{ message: "Unexpected 'beforeEach' hook" }],
    },
    {
      code: [
        "import { beforeEach as afterEach, afterEach as beforeEach } from '@rstest/core';",
        '',
        'afterEach(() => {});',
        'beforeEach(() => { resetModules() });',
      ].join('\n'),
      options: [{ allow: ['afterEach'] }],
      errors: [{ message: "Unexpected 'beforeEach' hook" }],
    },
  ],
});

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
      code: `
        import { 'afterEach' as afterEachTest } from '@jest/globals';

        afterEachTest(() => {})
      `,
      errors: [{ message: "Unexpected 'afterEach' hook" }],
    },
    {
      code: 'beforeEach(() => {}); afterEach(() => { jest.resetModules() });',
      options: [{ allow: ['afterEach'] }],
      errors: [{ message: "Unexpected 'beforeEach' hook" }],
    },
    {
      code: `
        import { beforeEach as afterEach, afterEach as beforeEach } from '@jest/globals';

        afterEach(() => {});
        beforeEach(() => { jest.resetModules() });
      `,
      options: [{ allow: ['afterEach'] }],
      errors: [{ message: "Unexpected 'beforeEach' hook" }],
    },
  ],
});

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
      errors: [{ message: 'Unexpected hook: {{beforeAll}}' }],
    },
    {
      code: 'beforeEach(() => {})',
      errors: [{ message: 'Unexpected hook: {{beforeEach}}' }],
    },
    {
      code: 'afterAll(() => {})',
      errors: [{ message: 'Unexpected hook: {{afterAll}}' }],
    },
    {
      code: 'afterEach(() => {})',
      errors: [{ message: 'Unexpected hook: {{afterEach}}' }],
    },
    {
      code: `
        import { 'afterEach' as afterEachTest } from '@jest/globals';

        afterEachTest(() => {})
      `,
      errors: [{ message: 'Unexpected hook: {{afterEach}}' }],
    },
    {
      code: 'beforeEach(() => {}); afterEach(() => { jest.resetModules() });',
      options: [{ allow: ['afterEach'] }],
      errors: [{ message: 'Unexpected hook: {{beforeEach}}' }],
    },
    {
      code: `
        import { beforeEach as afterEach, afterEach as beforeEach } from '@jest/globals';

        afterEach(() => {});
        beforeEach(() => { jest.resetModules() });
      `,
      options: [{ allow: ['afterEach'] }],
      errors: [{ message: 'Unexpected hook: {{beforeEach}}' }],
    },
  ],
});

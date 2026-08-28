import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const message = 'The use of `.todo` is not recommended.';

ruleTester.run('warn-todo', {} as never, {
  valid: [
    { code: 'describe("foo", function () {})' },
    { code: 'it("foo", function () {})' },
    { code: 'it.concurrent("foo", function () {})' },
    { code: 'test("foo", function () {})' },
    { code: 'test("foo", { todo: false }, function () {})' },
    { code: 'test.concurrent("foo", function () {})' },
    { code: 'describe.only("foo", function () {})' },
    { code: 'it.only("foo", function () {})' },
    { code: 'it.each()("foo", function () {})' },
    // Rstest's test options object has no `todo` field.
    { code: 'test("foo", { todo: true }, function () {})' },
  ],
  invalid: [
    {
      code: 'describe.todo("foo", function () {})',
      errors: [{ messageId: 'warnTodo', message, line: 1, column: 10 }],
    },
    {
      code: 'it.todo("foo", function () {})',
      errors: [{ messageId: 'warnTodo', message, line: 1, column: 4 }],
    },
    {
      code: 'test.todo("foo", function () {})',
      errors: [{ messageId: 'warnTodo', message, line: 1, column: 6 }],
    },
    {
      code: 'describe.todo.each([])("foo", function () {})',
      errors: [{ messageId: 'warnTodo', message, line: 1, column: 10 }],
    },
    {
      code: 'it.todo.each([])("foo", function () {})',
      errors: [{ messageId: 'warnTodo', message, line: 1, column: 4 }],
    },
    {
      code: 'test.todo.each([])("foo", function () {})',
      errors: [{ messageId: 'warnTodo', message, line: 1, column: 6 }],
    },
    {
      code: 'describe.only.todo("foo", function () {})',
      errors: [{ messageId: 'warnTodo', message, line: 1, column: 15 }],
    },
    {
      code: 'it.only.todo("foo", function () {})',
      errors: [{ messageId: 'warnTodo', message, line: 1, column: 9 }],
    },
    {
      code: 'test.only.todo("foo", function () {})',
      errors: [{ messageId: 'warnTodo', message, line: 1, column: 11 }],
    },
  ],
});

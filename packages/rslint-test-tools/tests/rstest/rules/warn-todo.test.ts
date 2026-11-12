import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('warn-todo', {
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
      errors: [
        {
          messageId: 'warnTodo',
          message: 'The use of `.todo` is not recommended.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'it.todo("foo", function () {})',
      errors: [
        {
          messageId: 'warnTodo',
          message: 'The use of `.todo` is not recommended.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: 'test.todo("foo", function () {})',
      errors: [
        {
          messageId: 'warnTodo',
          message: 'The use of `.todo` is not recommended.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: 'describe.todo.each([])("foo", function () {})',
      errors: [
        {
          messageId: 'warnTodo',
          message: 'The use of `.todo` is not recommended.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'it.todo.each([])("foo", function () {})',
      errors: [
        {
          messageId: 'warnTodo',
          message: 'The use of `.todo` is not recommended.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: 'test.todo.each([])("foo", function () {})',
      errors: [
        {
          messageId: 'warnTodo',
          message: 'The use of `.todo` is not recommended.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: 'describe.only.todo("foo", function () {})',
      errors: [
        {
          messageId: 'warnTodo',
          message: 'The use of `.todo` is not recommended.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'it.only.todo("foo", function () {})',
      errors: [
        {
          messageId: 'warnTodo',
          message: 'The use of `.todo` is not recommended.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: 'test.only.todo("foo", function () {})',
      errors: [
        {
          messageId: 'warnTodo',
          message: 'The use of `.todo` is not recommended.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
  ],
});

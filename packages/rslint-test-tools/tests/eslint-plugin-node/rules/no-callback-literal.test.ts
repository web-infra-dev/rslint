import { RuleTester } from '../rule-tester';

// All tests and documentation examples from eslint-plugin-n v18.3.0:
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-callback-literal.js
const ruleTester = new RuleTester();
ruleTester.run(
  'no-callback-literal',
  {},
  {
    valid: [
      // ---- random stuff ----
      { code: 'horse()', name: 'horse()' },
      { code: 'sort(null)', name: 'sort(null)' },
      { code: 'require("zyx")', name: 'require("zyx")' },
      { code: 'require("zyx", data)', name: 'require("zyx", data)' },
      // ---- callback() ----
      { code: 'callback()', name: 'callback()' },
      { code: 'callback(undefined)', name: 'callback(undefined)' },
      { code: 'callback(null)', name: 'callback(null)' },
      { code: 'callback(x)', name: 'callback(x)' },
      {
        code: 'callback(new Error("error"))',
        name: 'callback(new Error("error"))',
      },
      { code: 'callback(friendly, data)', name: 'callback(friendly, data)' },
      { code: 'callback(undefined, data)', name: 'callback(undefined, data)' },
      { code: 'callback(null, data)', name: 'callback(null, data)' },
      { code: 'callback(x, data)', name: 'callback(x, data)' },
      {
        code: 'callback(new Error("error"), data)',
        name: 'callback(new Error("error"), data)',
      },
      { code: 'callback(x = obj, data)', name: 'callback(x = obj, data)' },
      { code: 'callback((1, a), data)', name: 'callback((1, a), data)' },
      { code: 'callback(a || b, data)', name: 'callback(a || b, data)' },
      { code: 'callback(a ? b : c, data)', name: 'callback(a ? b : c, data)' },
      { code: 'callback(a ? 1 : c, data)', name: 'callback(a ? 1 : c, data)' },
      { code: 'callback(a ? b : 1, data)', name: 'callback(a ? b : 1, data)' },
      // ---- cb() ----
      { code: 'cb()', name: 'cb()' },
      { code: 'cb(undefined)', name: 'cb(undefined)' },
      { code: 'cb(null)', name: 'cb(null)' },
      { code: 'cb(undefined, "super")', name: 'cb(undefined, "super")' },
      { code: 'cb(null, "super")', name: 'cb(null, "super")' },
      // https://github.com/eslint-community/eslint-plugin-n/issues/162
      { code: 'cb(e as Error)', name: 'cb(e as Error)', filename: 'input.ts' },
      // ---- Documentation: correct example (rule directive omitted) ----
      {
        code: "cb(undefined);\ncb(null, 5);\ncallback(new Error('some error'));\ncallback(someVariable);",
        name: 'cb(undefined);',
      },
    ],
    invalid: [
      // ---- callback ----
      {
        code: 'callback(false, "snork")',
        name: 'callback(false, "snork")',
        errors: [
          {
            messageId: 'unexpectedLiteral',
            message: 'Unexpected literal in error position of callback.',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 25,
          },
        ],
      },
      {
        code: 'callback("help")',
        name: 'callback("help")',
        errors: [
          {
            messageId: 'unexpectedLiteral',
            message: 'Unexpected literal in error position of callback.',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 17,
          },
        ],
      },
      {
        code: 'callback("help", data)',
        name: 'callback("help", data)',
        errors: [
          {
            messageId: 'unexpectedLiteral',
            message: 'Unexpected literal in error position of callback.',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 23,
          },
        ],
      },
      // ---- cb ----
      {
        code: 'cb(false)',
        name: 'cb(false)',
        errors: [
          {
            messageId: 'unexpectedLiteral',
            message: 'Unexpected literal in error position of callback.',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 10,
          },
        ],
      },
      {
        code: 'cb("help")',
        name: 'cb("help")',
        errors: [
          {
            messageId: 'unexpectedLiteral',
            message: 'Unexpected literal in error position of callback.',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 11,
          },
        ],
      },
      {
        code: 'cb("help", data)',
        name: 'cb("help", data)',
        errors: [
          {
            messageId: 'unexpectedLiteral',
            message: 'Unexpected literal in error position of callback.',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 17,
          },
        ],
      },
      {
        code: 'cb({ a: 1 })',
        name: 'cb({ a: 1 })',
        errors: [
          {
            messageId: 'unexpectedLiteral',
            message: 'Unexpected literal in error position of callback.',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 13,
          },
        ],
      },
      {
        code: 'cb([])',
        name: 'cb([])',
        errors: [
          {
            messageId: 'unexpectedLiteral',
            message: 'Unexpected literal in error position of callback.',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 7,
          },
        ],
      },
      {
        code: 'cb(`message ${value}`)',
        name: 'cb(`message ${value}`)',
        errors: [
          {
            messageId: 'unexpectedLiteral',
            message: 'Unexpected literal in error position of callback.',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 23,
          },
        ],
      },
      {
        code: 'callback((a, 1), data)',
        name: 'callback((a, 1), data)',
        errors: [
          {
            messageId: 'unexpectedLiteral',
            message: 'Unexpected literal in error position of callback.',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 23,
          },
        ],
      },
      // ---- Documentation: incorrect example (rule directive omitted) ----
      {
        code: "cb('this is an error string');\ncb({ a: 1 });\ncallback(0);",
        name: "cb('this is an error string');",
        errors: [
          {
            messageId: 'unexpectedLiteral',
            message: 'Unexpected literal in error position of callback.',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 30,
          },
          {
            messageId: 'unexpectedLiteral',
            message: 'Unexpected literal in error position of callback.',
            line: 2,
            column: 1,
            endLine: 2,
            endColumn: 13,
          },
          {
            messageId: 'unexpectedLiteral',
            message: 'Unexpected literal in error position of callback.',
            line: 3,
            column: 1,
            endLine: 3,
            endColumn: 12,
          },
        ],
      },
    ],
  },
);

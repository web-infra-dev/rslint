import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('param-names', {
  valid: [
    // ESLint upstream
    { code: 'new Promise(function(resolve, reject) {})' },
    { code: 'new Promise(function(resolve, _reject) {})' },
    { code: 'new Promise(function(_resolve, reject) {})' },
    { code: 'new Promise(function(_resolve, _reject) {})' },
    { code: 'new Promise(function(resolve) {})' },
    { code: 'new Promise(function(_resolve) {})' },
    { code: 'new Promise(resolve => {})' },
    { code: 'new Promise((resolve, reject) => {})' },
    { code: 'new Promise(() => {})' },
    { code: 'new NonPromise()' },
    {
      code: 'new Promise((yes, no) => {})',
      options: [{ resolvePattern: '^yes$', rejectPattern: '^no$' }],
    },

    // Patterns/defaults/rest skip (mirrors ESLint `.name===undefined`)
    { code: 'new Promise(function({ resolve, reject }) {})' },
    { code: 'new Promise(function([resolve, reject]) {})' },
    { code: 'new Promise(function(resolve = () => {}, reject = () => {}) {})' },
    { code: 'new Promise(function(...args) {})' },
    { code: 'new Promise(function(resolve, { foo }) {})' },
    { code: 'new Promise(function({ foo }, reject) {})' },

    // Not a Promise constructor invocation
    { code: 'new Promise(handler)' },
    { code: 'new Promise(function(reject, resolve) {}, extraArg)' },
    { code: 'Promise(function(reject, resolve) {})' },
    { code: 'new Foo.Promise(function(reject, resolve) {})' },
    { code: 'new globalThis.Promise(function(reject, resolve) {})' },
    { code: 'new Promise()' },

    // Async executor
    { code: 'new Promise(async function(resolve, reject) {})' },
    { code: 'new Promise(async (resolve, reject) => {})' },

    // TS type assertion on callee / executor: rule silently skips,
    // mirroring ESLint's `callee.type !== 'Identifier'` short-circuit.
    { code: 'new (Promise as any)(function(resolve, reject) {})' },
    { code: 'new (Promise as any)(function(ok, fail) {})' },
    { code: 'new Promise((function(ok, fail) {}) as any)' },

    // Partial options
    {
      code: 'new Promise((yes, reject) => {})',
      options: [{ resolvePattern: '^yes$' }],
    },
    {
      code: 'new Promise((resolve, no) => {})',
      options: [{ rejectPattern: '^no$' }],
    },

    // regexp2 / ECMAScript features (lookbehind, \p{...})
    {
      code: 'new Promise((resolve) => {})',
      options: [{ resolvePattern: '(?<!_)resolve' }],
    },
    {
      code: 'new Promise((resolve) => {})',
      options: [{ resolvePattern: '^\\p{Ll}+$' }],
    },

    // Invalid pattern: rule silently no-ops
    {
      code: 'new Promise((reject, resolve) => {})',
      options: [{ resolvePattern: '[unclosed' }],
    },

    // TS generic type argument
    { code: 'new Promise<number>((resolve, reject) => {})' },
    { code: 'new Promise<void>(function(resolve, reject) {})' },

    // 3+ executor params: only first two checked
    { code: 'new Promise((resolve, reject, extra) => {})' },
    { code: 'new Promise((resolve, reject, a, b, c) => {})' },

    // Named function expression
    { code: 'new Promise(function named(resolve, reject) {})' },

    // Empty params on FunctionExpression
    { code: 'new Promise(function() {})' },

    // TS `this` parameter is stripped before resolve/reject indexing
    { code: 'new Promise<void>(function(this: any, resolve, reject) {})' },
    { code: 'new Promise<void>(function(this: unknown, _resolve) {})' },
  ],

  invalid: [
    // ESLint upstream
    {
      code: 'new Promise(function(reject, resolve) {})',
      errors: [
        {
          messageId: 'resolveParamNames',
          message:
            'Promise constructor parameters must be named to match "^_?resolve$"',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 28,
        },
        {
          messageId: 'rejectParamNames',
          message:
            'Promise constructor parameters must be named to match "^_?reject$"',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: 'new Promise(function(resolve, rej) {})',
      errors: [
        {
          messageId: 'rejectParamNames',
          message:
            'Promise constructor parameters must be named to match "^_?reject$"',
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: 'new Promise(yes => {})',
      errors: [
        {
          messageId: 'resolveParamNames',
          message:
            'Promise constructor parameters must be named to match "^_?resolve$"',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: 'new Promise((yes, no) => {})',
      errors: [
        {
          messageId: 'resolveParamNames',
          message:
            'Promise constructor parameters must be named to match "^_?resolve$"',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 17,
        },
        {
          messageId: 'rejectParamNames',
          message:
            'Promise constructor parameters must be named to match "^_?reject$"',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: 'new Promise(function(resolve, reject) {})',
      options: [{ resolvePattern: '^yes$', rejectPattern: '^no$' }],
      errors: [
        {
          messageId: 'resolveParamNames',
          message:
            'Promise constructor parameters must be named to match "^yes$"',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 29,
        },
        {
          messageId: 'rejectParamNames',
          message:
            'Promise constructor parameters must be named to match "^no$"',
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },

    // Single underscore is NOT acceptable under default pattern
    {
      code: 'new Promise(function(_, reject) {})',
      errors: [
        {
          messageId: 'resolveParamNames',
          message:
            'Promise constructor parameters must be named to match "^_?resolve$"',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },

    // Parenthesized Promise identifier
    {
      code: 'new (Promise)(function(ok, fail) {})',
      errors: [
        {
          messageId: 'resolveParamNames',
          message:
            'Promise constructor parameters must be named to match "^_?resolve$"',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 26,
        },
        {
          messageId: 'rejectParamNames',
          message:
            'Promise constructor parameters must be named to match "^_?reject$"',
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },

    // Async executor
    {
      code: 'new Promise(async function(ok, fail) {})',
      errors: [
        {
          messageId: 'resolveParamNames',
          message:
            'Promise constructor parameters must be named to match "^_?resolve$"',
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 30,
        },
        {
          messageId: 'rejectParamNames',
          message:
            'Promise constructor parameters must be named to match "^_?reject$"',
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      code: 'new Promise(async (ok, fail) => {})',
      errors: [
        {
          messageId: 'resolveParamNames',
          message:
            'Promise constructor parameters must be named to match "^_?resolve$"',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 22,
        },
        {
          messageId: 'rejectParamNames',
          message:
            'Promise constructor parameters must be named to match "^_?reject$"',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },

    // Partial options
    {
      code: 'new Promise((no, rejectX) => {})',
      options: [{ resolvePattern: '^yes$' }],
      errors: [
        {
          messageId: 'resolveParamNames',
          message:
            'Promise constructor parameters must be named to match "^yes$"',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 16,
        },
        {
          messageId: 'rejectParamNames',
          message:
            'Promise constructor parameters must be named to match "^_?reject$"',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: 'new Promise((resolveX, yes) => {})',
      options: [{ rejectPattern: '^no$' }],
      errors: [
        {
          messageId: 'resolveParamNames',
          message:
            'Promise constructor parameters must be named to match "^_?resolve$"',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 22,
        },
        {
          messageId: 'rejectParamNames',
          message:
            'Promise constructor parameters must be named to match "^no$"',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },

    // Pattern with backslashes: message must preserve the raw source
    {
      code: 'new Promise((bad) => {})',
      options: [{ resolvePattern: '^\\w+resolve$' }],
      errors: [
        {
          messageId: 'resolveParamNames',
          message:
            'Promise constructor parameters must be named to match "^\\w+resolve$"',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },

    // ECMAScript-only lookbehind compiles under regexp2
    {
      code: 'new Promise((_resolve) => {})',
      options: [{ resolvePattern: '(?<!_)resolve' }],
      errors: [
        {
          messageId: 'resolveParamNames',
          message:
            'Promise constructor parameters must be named to match "(?<!_)resolve"',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },

    // TS generic type argument: still reports bad names
    {
      code: 'new Promise<number>((ok, fail) => {})',
      errors: [
        {
          messageId: 'resolveParamNames',
          message:
            'Promise constructor parameters must be named to match "^_?resolve$"',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 24,
        },
        {
          messageId: 'rejectParamNames',
          message:
            'Promise constructor parameters must be named to match "^_?reject$"',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },

    // 3+ params: only first two checked
    {
      code: 'new Promise((ok, fail, also_wrong) => {})',
      errors: [
        {
          messageId: 'resolveParamNames',
          message:
            'Promise constructor parameters must be named to match "^_?resolve$"',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 16,
        },
        {
          messageId: 'rejectParamNames',
          message:
            'Promise constructor parameters must be named to match "^_?reject$"',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },

    // Named function expression
    {
      code: 'new Promise(function named(ok, fail) {})',
      errors: [
        {
          messageId: 'resolveParamNames',
          message:
            'Promise constructor parameters must be named to match "^_?resolve$"',
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 30,
        },
        {
          messageId: 'rejectParamNames',
          message:
            'Promise constructor parameters must be named to match "^_?reject$"',
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },

    // TS `this` parameter is stripped; bad names in real params still reported
    {
      code: 'new Promise<void>(function(this: any, ok, fail) {})',
      errors: [
        {
          messageId: 'resolveParamNames',
          message:
            'Promise constructor parameters must be named to match "^_?resolve$"',
          line: 1,
          column: 39,
          endLine: 1,
          endColumn: 41,
        },
        {
          messageId: 'rejectParamNames',
          message:
            'Promise constructor parameters must be named to match "^_?reject$"',
          line: 1,
          column: 43,
          endLine: 1,
          endColumn: 47,
        },
      ],
    },
  ],
});

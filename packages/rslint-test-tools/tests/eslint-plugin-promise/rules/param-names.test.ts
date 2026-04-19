import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const messageForResolve =
  'Promise constructor parameters must be named to match "^_?resolve$"';
const messageForReject =
  'Promise constructor parameters must be named to match "^_?reject$"';

ruleTester.run('param-names', {} as never, {
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
      errors: [{ message: messageForResolve }, { message: messageForReject }],
    },
    {
      code: 'new Promise(function(resolve, rej) {})',
      errors: [{ message: messageForReject }],
    },
    {
      code: 'new Promise(yes => {})',
      errors: [{ message: messageForResolve }],
    },
    {
      code: 'new Promise((yes, no) => {})',
      errors: [{ message: messageForResolve }, { message: messageForReject }],
    },
    {
      code: 'new Promise(function(resolve, reject) {})',
      options: [{ resolvePattern: '^yes$', rejectPattern: '^no$' }],
      errors: [
        {
          message:
            'Promise constructor parameters must be named to match "^yes$"',
        },
        {
          message:
            'Promise constructor parameters must be named to match "^no$"',
        },
      ],
    },

    // Single underscore is NOT acceptable under default pattern
    {
      code: 'new Promise(function(_, reject) {})',
      errors: [{ message: messageForResolve }],
    },

    // Parenthesized Promise identifier
    {
      code: 'new (Promise)(function(ok, fail) {})',
      errors: [{ message: messageForResolve }, { message: messageForReject }],
    },

    // Async executor
    {
      code: 'new Promise(async function(ok, fail) {})',
      errors: [{ message: messageForResolve }, { message: messageForReject }],
    },
    {
      code: 'new Promise(async (ok, fail) => {})',
      errors: [{ message: messageForResolve }, { message: messageForReject }],
    },

    // Partial options
    {
      code: 'new Promise((no, rejectX) => {})',
      options: [{ resolvePattern: '^yes$' }],
      errors: [
        {
          message:
            'Promise constructor parameters must be named to match "^yes$"',
        },
        { message: messageForReject },
      ],
    },
    {
      code: 'new Promise((resolveX, yes) => {})',
      options: [{ rejectPattern: '^no$' }],
      errors: [
        { message: messageForResolve },
        {
          message:
            'Promise constructor parameters must be named to match "^no$"',
        },
      ],
    },

    // Pattern with backslashes: message must preserve the raw source
    {
      code: 'new Promise((bad) => {})',
      options: [{ resolvePattern: '^\\w+resolve$' }],
      errors: [
        {
          message:
            'Promise constructor parameters must be named to match "^\\w+resolve$"',
        },
      ],
    },

    // ECMAScript-only lookbehind compiles under regexp2
    {
      code: 'new Promise((_resolve) => {})',
      options: [{ resolvePattern: '(?<!_)resolve' }],
      errors: [
        {
          message:
            'Promise constructor parameters must be named to match "(?<!_)resolve"',
        },
      ],
    },

    // TS generic type argument: still reports bad names
    {
      code: 'new Promise<number>((ok, fail) => {})',
      errors: [{ message: messageForResolve }, { message: messageForReject }],
    },

    // 3+ params: only first two checked
    {
      code: 'new Promise((ok, fail, also_wrong) => {})',
      errors: [{ message: messageForResolve }, { message: messageForReject }],
    },

    // Named function expression
    {
      code: 'new Promise(function named(ok, fail) {})',
      errors: [{ message: messageForResolve }, { message: messageForReject }],
    },

    // TS `this` parameter is stripped; bad names in real params still reported
    {
      code: 'new Promise<void>(function(this: any, ok, fail) {})',
      errors: [{ message: messageForResolve }, { message: messageForReject }],
    },
  ],
});

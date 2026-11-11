// Upstream: https://github.com/eslint-community/eslint-plugin-promise/blob/v7.3.0/__tests__/prefer-await-to-callbacks.js
// Documentation: https://github.com/eslint-community/eslint-plugin-promise/blob/v7.3.0/docs/rules/prefer-await-to-callbacks.md
import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('prefer-await-to-callbacks', {
  valid: [
    // Upstream tests.
    {
      code: 'async function hi() { await thing().catch(err => console.log(err)) }',
    },
    { code: 'async function hi() { await thing().then() }' },
    { code: 'async function hi() { await thing().catch() }' },
    { code: 'dbConn.on("error", err => { console.error(err) })' },
    { code: 'dbConn.once("error", err => { console.error(err) })' },
    { code: 'heart(something => {})' },
    { code: 'getErrors().map(error => responseTo(error))' },
    { code: 'errors.filter(err => err.status === 402)' },
    { code: 'errors.some(err => err.message.includes("Yo"))' },
    { code: 'errors.every(err => err.status === 402)' },
    { code: 'errors.filter(err => console.log(err))' },
    { code: 'const error = errors.find(err => err.stack.includes("file.js"))' },
    { code: 'this.myErrors.forEach(function(error) { log(error); })' },
    {
      code: 'find(errors, function(err) { return  err.type === "CoolError" })',
    },
    {
      code: 'map(errors, function(error) { return  err.type === "CoolError" })',
    },
    {
      code: '_.find(errors, function(error) { return  err.type === "CoolError" })',
    },
    {
      code: '_.map(errors, function(err) { return  err.type === "CoolError" })',
    },
    // Upstream documentation (the yield example is wrapped in a generator).
    { code: 'await doSomething(arg)' },
    { code: 'async function doSomethingElse() {}' },
    { code: 'function* values() { yield yieldValue(err => {}) }' },
    { code: "eventEmitter.on('error', err => {})" },
  ],
  invalid: [
    // Upstream tests.
    {
      code: 'heart(function(err) {})',
      errors: [
        {
          messageId: 'error',
          message: 'Avoid callbacks. Prefer Async/Await.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'heart(err => {})',
      errors: [
        {
          messageId: 'error',
          message: 'Avoid callbacks. Prefer Async/Await.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: 'heart("ball", function(err) {})',
      errors: [
        {
          messageId: 'error',
          message: 'Avoid callbacks. Prefer Async/Await.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: 'function getData(id, callback) {}',
      errors: [
        {
          messageId: 'error',
          message: 'Avoid callbacks. Prefer Async/Await.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: 'const getData = (cb) => {}',
      errors: [
        {
          messageId: 'error',
          message: 'Avoid callbacks. Prefer Async/Await.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: 'var x = function (x, cb) {}',
      errors: [
        {
          messageId: 'error',
          message: 'Avoid callbacks. Prefer Async/Await.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'cb()',
      errors: [
        {
          messageId: 'error',
          message: 'Avoid callbacks. Prefer Async/Await.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 5,
        },
      ],
    },
    {
      code: 'callback()',
      errors: [
        {
          messageId: 'error',
          message: 'Avoid callbacks. Prefer Async/Await.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: 'heart(error => {})',
      errors: [
        {
          messageId: 'error',
          message: 'Avoid callbacks. Prefer Async/Await.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'async.map(files, fs.stat, function(err, results) { if (err) throw err; });',
      errors: [
        {
          messageId: 'error',
          message: 'Avoid callbacks. Prefer Async/Await.',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 73,
        },
      ],
    },
    {
      code: '_.map(files, fs.stat, function(err, results) { if (err) throw err; });',
      errors: [
        {
          messageId: 'error',
          message: 'Avoid callbacks. Prefer Async/Await.',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 69,
        },
      ],
    },
    {
      code: 'map(files, fs.stat, function(err, results) { if (err) throw err; });',
      errors: [
        {
          messageId: 'error',
          message: 'Avoid callbacks. Prefer Async/Await.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 67,
        },
      ],
    },
    {
      code: 'map(function(err, results) { if (err) throw err; });',
      errors: [
        {
          messageId: 'error',
          message: 'Avoid callbacks. Prefer Async/Await.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 51,
        },
      ],
    },
    {
      code: 'customMap(errors, (err) => err.message)',
      errors: [
        {
          messageId: 'error',
          message: 'Avoid callbacks. Prefer Async/Await.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    // Upstream documentation (the yield example is wrapped in a generator).
    {
      code: 'cb()\ncallback()\ndoSomething(arg, (err) => {})\nfunction doSomethingElse(cb) {}',
      errors: [
        {
          messageId: 'error',
          message: 'Avoid callbacks. Prefer Async/Await.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 5,
        },
        {
          messageId: 'error',
          message: 'Avoid callbacks. Prefer Async/Await.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 11,
        },
        {
          messageId: 'error',
          message: 'Avoid callbacks. Prefer Async/Await.',
          line: 3,
          column: 18,
          endLine: 3,
          endColumn: 29,
        },
        {
          messageId: 'error',
          message: 'Avoid callbacks. Prefer Async/Await.',
          line: 4,
          column: 26,
          endLine: 4,
          endColumn: 28,
        },
      ],
    },
  ],
});

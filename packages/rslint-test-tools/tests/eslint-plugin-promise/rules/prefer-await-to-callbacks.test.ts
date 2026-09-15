// Upstream: https://github.com/eslint-community/eslint-plugin-promise/blob/v7.3.0/__tests__/prefer-await-to-callbacks.js
// Documentation: https://github.com/eslint-community/eslint-plugin-promise/blob/v7.3.0/docs/rules/prefer-await-to-callbacks.md
import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();
const message = 'Avoid callbacks. Prefer Async/Await.';

ruleTester.run('prefer-await-to-callbacks', {} as never, {
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
    { code: 'heart(function(err) {})', errors: [{ message }] },
    { code: 'heart(err => {})', errors: [{ message }] },
    { code: 'heart("ball", function(err) {})', errors: [{ message }] },
    { code: 'function getData(id, callback) {}', errors: [{ message }] },
    { code: 'const getData = (cb) => {}', errors: [{ message }] },
    { code: 'var x = function (x, cb) {}', errors: [{ message }] },
    { code: 'cb()', errors: [{ message }] },
    { code: 'callback()', errors: [{ message }] },
    { code: 'heart(error => {})', errors: [{ message }] },
    {
      code: 'async.map(files, fs.stat, function(err, results) { if (err) throw err; });',
      errors: [{ message }],
    },
    {
      code: '_.map(files, fs.stat, function(err, results) { if (err) throw err; });',
      errors: [{ message }],
    },
    {
      code: 'map(files, fs.stat, function(err, results) { if (err) throw err; });',
      errors: [{ message }],
    },
    {
      code: 'map(function(err, results) { if (err) throw err; });',
      errors: [{ message }],
    },
    { code: 'customMap(errors, (err) => err.message)', errors: [{ message }] },
    // Upstream documentation (the yield example is wrapped in a generator).
    {
      code: 'cb()\ncallback()\ndoSomething(arg, (err) => {})\nfunction doSomethingElse(cb) {}',
      errors: [{ message }, { message }, { message }, { message }],
    },
  ],
});

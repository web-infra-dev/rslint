import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const errorMessage = 'Avoid using promises inside of callbacks.';

ruleTester.run('no-promise-in-callback', {} as never, {
  valid: [
    { code: 'go(function() { return Promise.resolve(4) })' },
    { code: 'go(function() { return a.then(b) })' },
    { code: 'go(function() { b.catch(c) })' },
    { code: 'go(function() { b.then(c, d) })' },

    // arrow functions and other things
    { code: 'go(() => Promise.resolve(4))' },
    { code: 'go((errrr) => a.then(b))' },
    { code: 'go((helpers) => { b.catch(c) })' },
    { code: 'go((e) => { b.then(c, d) })' },

    // within promises it won't complain
    { code: 'a.catch((err) => { b.then(c, d) })' },

    // random unrelated things
    { code: 'var x = function() { return Promise.resolve(4) }' },
    { code: 'function y() { return Promise.resolve(4) }' },
    { code: 'function then() { return Promise.reject() }' },
    { code: 'doThing(function(x) { return Promise.reject(x) })' },
    { code: 'doThing().then(function() { return Promise.all([a,b,c]) })' },
    { code: 'doThing().then(function() { return Promise.resolve(4) })' },
    { code: 'doThing().then(() => Promise.resolve(4))' },
    { code: 'doThing().then(() => Promise.all([a]))' },

    // weird case, upstream assumes it's okay if you return
    { code: 'a(function(err) { return doThing().then(a) })' },

    {
      code: `
        function fn(err) {
          return { promise: Promise.resolve(err) };
        }
      `,
      options: [{ exemptDeclarations: true }],
    },
  ],

  invalid: [
    {
      code: 'a(function(err) { doThing().then(a) })',
      errors: [{ message: errorMessage }],
    },
    {
      code: 'a(function(error, zup, supa) { doThing().then(a) })',
      errors: [{ message: errorMessage }],
    },
    {
      code: 'a(function(error) { doThing().then(a) })',
      errors: [{ message: errorMessage }],
    },

    // arrow function
    {
      code: 'a((error) => { doThing().then(a) })',
      errors: [{ message: errorMessage }],
    },
    {
      code: 'a((error) => doThing().then(a))',
      errors: [{ message: errorMessage }],
    },
    {
      code: 'a((err, data) => { doThing().then(a) })',
      errors: [{ message: errorMessage }],
    },
    {
      code: 'a((err, data) => doThing().then(a))',
      errors: [{ message: errorMessage }],
    },

    // function declarations and similar
    {
      code: 'function x(err) { Promise.all() }',
      errors: [{ message: errorMessage }],
    },
    {
      code: 'function x(err) { Promise.allSettled() }',
      errors: [{ message: errorMessage }],
    },
    {
      code: 'function x(err) { Promise.any() }',
      errors: [{ message: errorMessage }],
    },
    {
      code: 'let x = (err) => doThingWith(err).then(a)',
      errors: [{ message: errorMessage }],
    },
  ],
});

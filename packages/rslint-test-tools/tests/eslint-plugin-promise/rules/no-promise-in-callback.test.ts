import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('no-promise-in-callback', {
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

    // object/class methods whose first param isn't err/error, and getters
    // (which take no params) are never err/error callbacks
    { code: 'const o = { onError(e) { Promise.resolve(e) } }' },
    { code: 'class X { get value() { return p.then(a) } }' },
    {
      code: 'class X { constructor(public err: unknown) { Promise.resolve(err) } }',
    },

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
      errors: [
        {
          messageId: 'avoidPromiseInCallback',
          message: 'Avoid using promises inside of callbacks.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: 'a(function(error, zup, supa) { doThing().then(a) })',
      errors: [
        {
          messageId: 'avoidPromiseInCallback',
          message: 'Avoid using promises inside of callbacks.',
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 46,
        },
      ],
    },
    {
      code: 'a(function(error) { doThing().then(a) })',
      errors: [
        {
          messageId: 'avoidPromiseInCallback',
          message: 'Avoid using promises inside of callbacks.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },

    // arrow function
    {
      code: 'a((error) => { doThing().then(a) })',
      errors: [
        {
          messageId: 'avoidPromiseInCallback',
          message: 'Avoid using promises inside of callbacks.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: 'a((error) => doThing().then(a))',
      errors: [
        {
          messageId: 'avoidPromiseInCallback',
          message: 'Avoid using promises inside of callbacks.',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: 'a((err, data) => { doThing().then(a) })',
      errors: [
        {
          messageId: 'avoidPromiseInCallback',
          message: 'Avoid using promises inside of callbacks.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: 'a((err, data) => doThing().then(a))',
      errors: [
        {
          messageId: 'avoidPromiseInCallback',
          message: 'Avoid using promises inside of callbacks.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },

    // function declarations and similar
    {
      code: 'function x(err) { Promise.all() }',
      errors: [
        {
          messageId: 'avoidPromiseInCallback',
          message: 'Avoid using promises inside of callbacks.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: 'function x(err) { Promise.allSettled() }',
      errors: [
        {
          messageId: 'avoidPromiseInCallback',
          message: 'Avoid using promises inside of callbacks.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: 'function x(err) { Promise.any() }',
      errors: [
        {
          messageId: 'avoidPromiseInCallback',
          message: 'Avoid using promises inside of callbacks.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: 'let x = (err) => doThingWith(err).then(a)',
      errors: [
        {
          messageId: 'avoidPromiseInCallback',
          message: 'Avoid using promises inside of callbacks.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      code: 'a(function(err) { return doThing()?.then(a) })',
      errors: [
        {
          messageId: 'avoidPromiseInCallback',
          message: 'Avoid using promises inside of callbacks.',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      code: 'a(function(err) { return Promise?.resolve(err) })',
      errors: [
        {
          messageId: 'avoidPromiseInCallback',
          message: 'Avoid using promises inside of callbacks.',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },

    // object/class methods with an err/error first param are callbacks too
    {
      code: 'const o = { onError(err) { Promise.resolve(err) } }',
      errors: [
        {
          messageId: 'avoidPromiseInCallback',
          message: 'Avoid using promises inside of callbacks.',
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 43,
        },
      ],
    },
    {
      code: 'class X { onError(error) { audit().catch(log) } }',
      errors: [
        {
          messageId: 'avoidPromiseInCallback',
          message: 'Avoid using promises inside of callbacks.',
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
  ],
});

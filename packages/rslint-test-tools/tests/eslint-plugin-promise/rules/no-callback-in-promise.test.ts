import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('no-callback-in-promise', {
  valid: [
    // ESLint upstream
    { code: 'function thing(cb) { cb() }' },
    { code: 'doSomething(function(err) { cb(err) })' },
    { code: 'function thing(callback) { callback() }' },
    { code: 'doSomething(function(err) { callback(err) })' },

    // Support safe callbacks (#220)
    { code: 'whatever.then((err) => { process.nextTick(() => cb()) })' },
    { code: 'whatever.then((err) => { setImmediate(() => cb()) })' },
    { code: 'whatever.then((err) => setImmediate(() => cb()))' },
    { code: 'whatever.then((err) => process.nextTick(() => cb()))' },
    { code: 'whatever.then((err) => process.nextTick(cb))' },
    { code: 'whatever.then((err) => setImmediate(cb))' },

    // Arrow functions and other things
    { code: 'let thing = (cb) => cb()' },
    { code: 'doSomething(err => cb(err))' },

    // Exceptions option
    { code: 'a.then(() => next())', options: [{ exceptions: ['next'] }] },
    {
      code: 'a.then(() => next()).catch((err) => next(err))',
      options: [{ exceptions: ['next'] }],
    },
    { code: 'a.then(next)', options: [{ exceptions: ['next'] }] },
    { code: 'a.then(next).catch(next)', options: [{ exceptions: ['next'] }] },

    // #572
    {
      code: `while (!(step = call(next, iterator)).done) {
       if (result !== undefined) break;
     }`,
    },
    {
      code: `function hasCallbackArg(callback) {
       console.log(callback);
     }`,
    },
  ],

  invalid: [
    // cb directly passed to .then / .catch
    {
      code: 'a.then(cb)',
      errors: [
        {
          messageId: 'callback',
          message: 'Avoid calling back inside of a promise.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: 'a.then(() => cb())',
      errors: [
        {
          messageId: 'callback',
          message: 'Avoid calling back inside of a promise.',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'a.then(function(err) { cb(err) })',
      errors: [
        {
          messageId: 'callback',
          message: 'Avoid calling back inside of a promise.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: 'a.then(function(data) { cb(data) }, function(err) { cb(err) })',
      errors: [
        {
          messageId: 'callback',
          message: 'Avoid calling back inside of a promise.',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 33,
        },
        {
          messageId: 'callback',
          message: 'Avoid calling back inside of a promise.',
          line: 1,
          column: 53,
          endLine: 1,
          endColumn: 60,
        },
      ],
    },
    {
      code: 'a.catch(function(err) { cb(err) })',
      errors: [
        {
          messageId: 'callback',
          message: 'Avoid calling back inside of a promise.',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },

    // "callback" name also flagged
    {
      code: 'a.then(callback)',
      errors: [
        {
          messageId: 'callback',
          message: 'Avoid calling back inside of a promise.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: 'a.then(() => callback())',
      errors: [
        {
          messageId: 'callback',
          message: 'Avoid calling back inside of a promise.',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'a.then(function(err) { callback(err) })',
      errors: [
        {
          messageId: 'callback',
          message: 'Avoid calling back inside of a promise.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: 'a.then(function(data) { callback(data) }, function(err) { callback(err) })',
      errors: [
        {
          messageId: 'callback',
          message: 'Avoid calling back inside of a promise.',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 39,
        },
        {
          messageId: 'callback',
          message: 'Avoid calling back inside of a promise.',
          line: 1,
          column: 59,
          endLine: 1,
          endColumn: 72,
        },
      ],
    },
    {
      code: 'a.catch(function(err) { callback(err) })',
      errors: [
        {
          messageId: 'callback',
          message: 'Avoid calling back inside of a promise.',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },

    // #167 — timeoutsErr: true cases
    {
      code: `
        function wait (callback) {
          return Promise.resolve()
            .then(() => {
              setTimeout(callback);
            });
        }
      `,
      errors: [
        {
          messageId: 'callback',
          message: 'Avoid calling back inside of a promise.',
          line: 5,
          column: 15,
          endLine: 5,
          endColumn: 35,
        },
      ],
      options: [{ timeoutsErr: true }],
    },
    {
      code: `
        function wait (callback) {
          return Promise.resolve()
            .then(() => {
              setTimeout(() => callback());
            });
        }
      `,
      errors: [
        {
          messageId: 'callback',
          message: 'Avoid calling back inside of a promise.',
          line: 5,
          column: 32,
          endLine: 5,
          endColumn: 42,
        },
      ],
      options: [{ timeoutsErr: true }],
    },

    // timeoutsErr: true — timeout functions inside promise handlers
    {
      code: 'whatever.then((err) => { process.nextTick(() => cb()) })',
      errors: [
        {
          messageId: 'callback',
          message: 'Avoid calling back inside of a promise.',
          line: 1,
          column: 49,
          endLine: 1,
          endColumn: 53,
        },
      ],
      options: [{ timeoutsErr: true }],
    },
    {
      code: 'whatever.then((err) => { setImmediate(() => cb()) })',
      errors: [
        {
          messageId: 'callback',
          message: 'Avoid calling back inside of a promise.',
          line: 1,
          column: 45,
          endLine: 1,
          endColumn: 49,
        },
      ],
      options: [{ timeoutsErr: true }],
    },
    {
      code: 'whatever.then((err) => setImmediate(() => cb()))',
      errors: [
        {
          messageId: 'callback',
          message: 'Avoid calling back inside of a promise.',
          line: 1,
          column: 43,
          endLine: 1,
          endColumn: 47,
        },
      ],
      options: [{ timeoutsErr: true }],
    },
    {
      code: 'whatever.then((err) => process.nextTick(() => cb()))',
      errors: [
        {
          messageId: 'callback',
          message: 'Avoid calling back inside of a promise.',
          line: 1,
          column: 47,
          endLine: 1,
          endColumn: 51,
        },
      ],
      options: [{ timeoutsErr: true }],
    },
    {
      code: 'whatever.then((err) => process.nextTick(cb))',
      errors: [
        {
          messageId: 'callback',
          message: 'Avoid calling back inside of a promise.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 44,
        },
      ],
      options: [{ timeoutsErr: true }],
    },
    {
      code: 'whatever.then((err) => setImmediate(cb))',
      errors: [
        {
          messageId: 'callback',
          message: 'Avoid calling back inside of a promise.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 40,
        },
      ],
      options: [{ timeoutsErr: true }],
    },
  ],
});

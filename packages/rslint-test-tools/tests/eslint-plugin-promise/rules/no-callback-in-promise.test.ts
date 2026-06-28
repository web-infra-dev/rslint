import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const callbackMessage = 'Avoid calling back inside of a promise.';

ruleTester.run('no-callback-in-promise', {} as never, {
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
      errors: [{ message: callbackMessage }],
    },
    {
      code: 'a.then(() => cb())',
      errors: [{ message: callbackMessage }],
    },
    {
      code: 'a.then(function(err) { cb(err) })',
      errors: [{ message: callbackMessage }],
    },
    {
      code: 'a.then(function(data) { cb(data) }, function(err) { cb(err) })',
      errors: [{ message: callbackMessage }, { message: callbackMessage }],
    },
    {
      code: 'a.catch(function(err) { cb(err) })',
      errors: [{ message: callbackMessage }],
    },

    // "callback" name also flagged
    {
      code: 'a.then(callback)',
      errors: [{ message: callbackMessage }],
    },
    {
      code: 'a.then(() => callback())',
      errors: [{ message: callbackMessage }],
    },
    {
      code: 'a.then(function(err) { callback(err) })',
      errors: [{ message: callbackMessage }],
    },
    {
      code: 'a.then(function(data) { callback(data) }, function(err) { callback(err) })',
      errors: [{ message: callbackMessage }, { message: callbackMessage }],
    },
    {
      code: 'a.catch(function(err) { callback(err) })',
      errors: [{ message: callbackMessage }],
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
      errors: [{ message: callbackMessage }],
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
      errors: [{ message: callbackMessage }],
      options: [{ timeoutsErr: true }],
    },

    // timeoutsErr: true — timeout functions inside promise handlers
    {
      code: 'whatever.then((err) => { process.nextTick(() => cb()) })',
      errors: [{ message: callbackMessage }],
      options: [{ timeoutsErr: true }],
    },
    {
      code: 'whatever.then((err) => { setImmediate(() => cb()) })',
      errors: [{ message: callbackMessage }],
      options: [{ timeoutsErr: true }],
    },
    {
      code: 'whatever.then((err) => setImmediate(() => cb()))',
      errors: [{ message: callbackMessage }],
      options: [{ timeoutsErr: true }],
    },
    {
      code: 'whatever.then((err) => process.nextTick(() => cb()))',
      errors: [{ message: callbackMessage }],
      options: [{ timeoutsErr: true }],
    },
    {
      code: 'whatever.then((err) => process.nextTick(cb))',
      errors: [{ message: callbackMessage }],
      options: [{ timeoutsErr: true }],
    },
    {
      code: 'whatever.then((err) => setImmediate(cb))',
      errors: [{ message: callbackMessage }],
      options: [{ timeoutsErr: true }],
    },
  ],
});

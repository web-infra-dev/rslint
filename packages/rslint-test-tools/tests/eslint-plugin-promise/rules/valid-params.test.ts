import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('valid-params', {} as never, {
  valid: [
    // valid Promise.resolve()
    { code: 'Promise.resolve()' },
    { code: 'Promise.resolve(1)' },
    { code: 'Promise.resolve({})' },
    { code: 'Promise.resolve(referenceToSomething)' },

    // valid Promise.reject()
    { code: 'Promise.reject()' },
    { code: 'Promise.reject(1)' },
    { code: 'Promise.reject({})' },
    { code: 'Promise.reject(referenceToSomething)' },
    { code: 'Promise.reject(Error())' },

    // valid Promise.race()
    { code: 'Promise.race([])' },
    { code: 'Promise.race(iterable)' },
    { code: 'Promise.race([one, two, three])' },

    // valid Promise.all()
    { code: 'Promise.all([])' },
    { code: 'Promise.all(iterable)' },
    { code: 'Promise.all([one, two, three])' },

    // valid Promise.allSettled()
    { code: 'Promise.allSettled([])' },
    { code: 'Promise.allSettled(iterable)' },
    { code: 'Promise.allSettled([one, two, three])' },

    // valid Promise.any()
    { code: 'Promise.any([])' },
    { code: 'Promise.any(iterable)' },
    { code: 'Promise.any([one, two, three])' },

    // valid Promise.then()
    { code: 'somePromise().then(success)' },
    { code: 'somePromise().then(success, failure)' },
    { code: 'promiseReference.then(() => {})' },
    { code: 'promiseReference.then(() => {}, () => {})' },

    // valid Promise.catch()
    { code: 'somePromise().catch(callback)' },
    { code: 'somePromise().catch(err => {})' },
    { code: 'promiseReference.catch(callback)' },
    { code: 'promiseReference.catch(err => {})' },

    // valid Promise.finally()
    { code: 'somePromise().finally(callback)' },
    { code: 'somePromise().finally(() => {})' },
    { code: 'promiseReference.finally(callback)' },
    { code: 'promiseReference.finally(() => {})' },

    {
      code: `
        somePromise.then(function() {
          return sth();
        }).catch(TypeError, function(e) {
          //
        }).catch(function(e) {
        });
      `,
      options: [{ exclude: ['catch'] }],
    },

    // integration test
    {
      code: `
      Promise.all([
        Promise.resolve(1),
        Promise.resolve(2),
        Promise.reject(Error()),
      ])
        .then(console.log)
        .catch(console.error)
        .finally(console.log)
    `,
    },
  ],
  invalid: [
    // invalid Promise.resolve()
    {
      code: 'Promise.resolve(1, 2)',
      errors: [
        {
          message:
            'Promise.resolve() requires 0 or 1 arguments, but received 2',
        },
      ],
    },
    {
      code: 'Promise.resolve({}, function() {}, 1, 2, 3)',
      errors: [
        {
          message:
            'Promise.resolve() requires 0 or 1 arguments, but received 5',
        },
      ],
    },

    // invalid Promise.reject()
    {
      code: 'Promise.reject(1, 2, 3)',
      errors: [
        {
          message: 'Promise.reject() requires 0 or 1 arguments, but received 3',
        },
      ],
    },
    {
      code: 'Promise.reject({}, function() {}, 1, 2)',
      errors: [
        {
          message: 'Promise.reject() requires 0 or 1 arguments, but received 4',
        },
      ],
    },

    // invalid Promise.race()
    {
      code: 'Promise.race(1, 2)',
      errors: [
        { message: 'Promise.race() requires 1 argument, but received 2' },
      ],
    },
    {
      code: 'Promise.race({}, function() {}, 1, 2, 3)',
      errors: [
        { message: 'Promise.race() requires 1 argument, but received 5' },
      ],
    },

    // invalid Promise.all()
    {
      code: 'Promise.all(1, 2, 3)',
      errors: [
        { message: 'Promise.all() requires 1 argument, but received 3' },
      ],
    },
    {
      code: 'Promise.all({}, function() {}, 1, 2)',
      errors: [
        { message: 'Promise.all() requires 1 argument, but received 4' },
      ],
    },
    // invalid Promise.allSettled()
    {
      code: 'Promise.allSettled(1, 2, 3)',
      errors: [
        { message: 'Promise.allSettled() requires 1 argument, but received 3' },
      ],
    },
    {
      code: 'Promise.allSettled({}, function() {}, 1, 2)',
      errors: [
        { message: 'Promise.allSettled() requires 1 argument, but received 4' },
      ],
    },
    // invalid Promise.any()
    {
      code: 'Promise.any(1, 2, 3)',
      errors: [
        { message: 'Promise.any() requires 1 argument, but received 3' },
      ],
    },
    {
      code: 'Promise.any({}, function() {}, 1, 2)',
      errors: [
        { message: 'Promise.any() requires 1 argument, but received 4' },
      ],
    },

    // invalid Promise.then()
    {
      code: 'somePromise().then()',
      errors: [
        { message: 'Promise.then() requires 1 or 2 arguments, but received 0' },
      ],
    },
    {
      code: 'somePromise().then(() => {}, () => {}, () => {})',
      errors: [
        { message: 'Promise.then() requires 1 or 2 arguments, but received 3' },
      ],
    },
    {
      code: 'promiseReference.then()',
      errors: [
        { message: 'Promise.then() requires 1 or 2 arguments, but received 0' },
      ],
    },
    {
      code: 'promiseReference.then(() => {}, () => {}, () => {})',
      errors: [
        { message: 'Promise.then() requires 1 or 2 arguments, but received 3' },
      ],
    },

    // invalid Promise.catch()
    {
      code: 'somePromise().catch()',
      errors: [
        { message: 'Promise.catch() requires 1 argument, but received 0' },
      ],
    },
    {
      code: 'somePromise().catch(() => {}, () => {})',
      errors: [
        { message: 'Promise.catch() requires 1 argument, but received 2' },
      ],
    },
    {
      code: 'promiseReference.catch()',
      errors: [
        { message: 'Promise.catch() requires 1 argument, but received 0' },
      ],
    },
    {
      code: 'promiseReference.catch(() => {}, () => {})',
      errors: [
        { message: 'Promise.catch() requires 1 argument, but received 2' },
      ],
    },

    // invalid Promise.finally()
    {
      code: 'somePromise().finally()',
      errors: [
        { message: 'Promise.finally() requires 1 argument, but received 0' },
      ],
    },
    {
      code: 'somePromise().finally(() => {}, () => {})',
      errors: [
        { message: 'Promise.finally() requires 1 argument, but received 2' },
      ],
    },
    {
      code: 'promiseReference.finally()',
      errors: [
        { message: 'Promise.finally() requires 1 argument, but received 0' },
      ],
    },
    {
      code: 'promiseReference.finally(() => {}, () => {})',
      errors: [
        { message: 'Promise.finally() requires 1 argument, but received 2' },
      ],
    },
  ],
});

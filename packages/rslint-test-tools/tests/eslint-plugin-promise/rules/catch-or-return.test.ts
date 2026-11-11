import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('catch-or-return', {
  valid: [
    // ESLint upstream — catch
    { code: 'frank().then(go).catch(doIt)' },
    { code: 'frank().then(go).then().then().then().catch(doIt)' },
    { code: 'frank().then(go).then().catch(function() { /* why bother */ })' },
    { code: 'frank.then(go).then(to).catch(jail)' },
    { code: 'Promise.resolve(frank).catch(jail)' },
    { code: 'Promise.resolve(frank)["catch"](jail)' },
    { code: 'frank.then(to).finally(fn).catch(jail)' },

    // Cypress
    { code: 'cy.get(".myClass").then(go)' },
    { code: 'cy.get("button").click().then()' },

    // return
    { code: 'function a() { return frank().then(go) }' },
    { code: 'function a() { return frank().then(go).then().then().then() }' },
    { code: 'function a() { return frank().then(go).then()}' },
    { code: 'function a() { return frank.then(go).then(to) }' },

    // allowThen - .then(null, fn)
    {
      code: 'frank().then(go).then(null, doIt)',
      options: [{ allowThen: true }],
    },
    {
      code: 'frank().then(go).then().then().then().then(null, doIt)',
      options: [{ allowThen: true }],
    },
    {
      code: 'frank().then(go).then().then(null, function() { /* why bother */ })',
      options: [{ allowThen: true }],
    },
    {
      code: 'frank.then(go).then(to).then(null, jail)',
      options: [{ allowThen: true }],
    },
    {
      code: 'frank().then(a).then(b).then(null, c)',
      options: [{ allowThen: true }],
    },
    {
      code: 'frank().then(a).then(b).then().then().then(null, doIt)',
      options: [{ allowThen: true }],
    },
    {
      code: 'frank().then(a).then(b).then(null, function() { /* why bother */ })',
      options: [{ allowThen: true }],
    },

    // allowThen - .then(fn, fn)
    { code: 'frank().then(a, b)', options: [{ allowThen: true }] },
    {
      code: 'frank().then(go).then(zam, doIt)',
      options: [{ allowThen: true }],
    },
    {
      code: 'frank().then(a).then(b).then(c, d)',
      options: [{ allowThen: true }],
    },
    {
      code: 'frank().then(go).then().then().then().then(wham, doIt)',
      options: [{ allowThen: true }],
    },
    {
      code: 'frank().then(go).then().then(function() {}, function() { /* why bother */ })',
      options: [{ allowThen: true }],
    },
    {
      code: 'frank.then(go).then(to).then(pewPew, jail)',
      options: [{ allowThen: true }],
    },

    // allowThenStrict - .then(null, fn)
    {
      code: 'frank().then(go).then(null, doIt)',
      options: [{ allowThenStrict: true }],
    },
    {
      code: 'frank().then(go).then().then().then().then(null, doIt)',
      options: [{ allowThenStrict: true }],
    },
    {
      code: 'frank().then(go).then().then(null, function() { /* why bother */ })',
      options: [{ allowThenStrict: true }],
    },
    {
      code: 'frank.then(go).then(to).then(null, jail)',
      options: [{ allowThenStrict: true }],
    },
    {
      code: 'frank().then(a).then(b).then(null, c)',
      options: [{ allowThenStrict: true }],
    },
    {
      code: 'frank().then(a).then(b).then().then().then(null, doIt)',
      options: [{ allowThenStrict: true }],
    },
    {
      code: 'frank().then(a).then(b).then(null, function() { /* why bother */ })',
      options: [{ allowThenStrict: true }],
    },

    // allowFinally
    {
      code: 'frank().then(go).catch(doIt).finally(fn)',
      options: [{ allowFinally: true }],
    },
    {
      code: 'frank().then(go).then().then().then().catch(doIt).finally(fn)',
      options: [{ allowFinally: true }],
    },
    {
      code: 'frank().then(go).then().catch(function() { /* why bother */ }).finally(fn)',
      options: [{ allowFinally: true }],
    },

    // terminationMethod=done
    {
      code: 'frank().then(go).done()',
      options: [{ terminationMethod: 'done' }],
    },

    // terminationMethod=[catch, done]
    {
      code: 'frank().then(go).catch()',
      options: [{ terminationMethod: ['catch', 'done'] }],
    },
    {
      code: 'frank().then(go).done()',
      options: [{ terminationMethod: ['catch', 'done'] }],
    },
    {
      code: 'frank().then(go).finally()',
      options: [{ terminationMethod: ['catch', 'finally'] }],
    },

    // for coverage
    { code: 'nonPromiseExpressionStatement();' },
  ],

  invalid: [
    // catch failures
    {
      code: 'function callPromise(promise, cb) { promise.then(cb) }',
      errors: [
        {
          messageId: 'terminationMethod',
          message: 'Expected catch() or return',
          line: 1,
          column: 37,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
    {
      code: 'fetch("http://www.yahoo.com").then(console.log.bind(console))',
      errors: [
        {
          messageId: 'terminationMethod',
          message: 'Expected catch() or return',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 62,
        },
      ],
    },
    {
      code: 'a.then(function() { return "x"; }).then(function(y) { throw y; })',
      errors: [
        {
          messageId: 'terminationMethod',
          message: 'Expected catch() or return',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 66,
        },
      ],
    },
    {
      code: 'Promise.resolve(frank)',
      errors: [
        {
          messageId: 'terminationMethod',
          message: 'Expected catch() or return',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'Promise.all([])',
      errors: [
        {
          messageId: 'terminationMethod',
          message: 'Expected catch() or return',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: 'Promise.allSettled([])',
      errors: [
        {
          messageId: 'terminationMethod',
          message: 'Expected catch() or return',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'Promise.any([])',
      errors: [
        {
          messageId: 'terminationMethod',
          message: 'Expected catch() or return',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: 'Promise.race([])',
      errors: [
        {
          messageId: 'terminationMethod',
          message: 'Expected catch() or return',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: 'frank().then(to).catch(fn).then(foo)',
      errors: [
        {
          messageId: 'terminationMethod',
          message: 'Expected catch() or return',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: 'frank().finally(fn)',
      errors: [
        {
          messageId: 'terminationMethod',
          message: 'Expected catch() or return',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: 'frank().then(to).finally(fn)',
      errors: [
        {
          messageId: 'terminationMethod',
          message: 'Expected catch() or return',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: 'frank().then(go).catch(doIt).finally(fn)',
      errors: [
        {
          messageId: 'terminationMethod',
          message: 'Expected catch() or return',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      code: 'frank().then(go).then().then().then().catch(doIt).finally(fn)',
      errors: [
        {
          messageId: 'terminationMethod',
          message: 'Expected catch() or return',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 62,
        },
      ],
    },
    {
      code: 'frank().then(go).then().catch(function() { /* why bother */ }).finally(fn)',
      errors: [
        {
          messageId: 'terminationMethod',
          message: 'Expected catch() or return',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 75,
        },
      ],
    },

    // return failures
    {
      code: 'function a() { frank().then(go) }',
      errors: [
        {
          messageId: 'terminationMethod',
          message: 'Expected catch() or return',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: 'function a() { frank().then(go).then().then().then() }',
      errors: [
        {
          messageId: 'terminationMethod',
          message: 'Expected catch() or return',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
    {
      code: 'function a() { frank().then(go).then()}',
      errors: [
        {
          messageId: 'terminationMethod',
          message: 'Expected catch() or return',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      code: 'function a() { frank.then(go).then(to) }',
      errors: [
        {
          messageId: 'terminationMethod',
          message: 'Expected catch() or return',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },

    // allowFinally=true failures
    {
      code: 'frank().then(go).catch(doIt).finally(fn).then(foo)',
      options: [{ allowFinally: true }],
      errors: [
        {
          messageId: 'terminationMethod',
          message: 'Expected catch() or return',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 51,
        },
      ],
    },
    {
      code: 'frank().then(go).catch(doIt).finally(fn).foobar(foo)',
      options: [{ allowFinally: true }],
      errors: [
        {
          messageId: 'terminationMethod',
          message: 'Expected catch() or return',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },

    // terminationMethod=done failures
    {
      code: 'frank().then(go)',
      options: [{ terminationMethod: 'done' }],
      errors: [
        {
          messageId: 'terminationMethod',
          message: 'Expected done() or return',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: 'frank().catch(go)',
      options: [{ terminationMethod: 'done' }],
      errors: [
        {
          messageId: 'terminationMethod',
          message: 'Expected done() or return',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },

    // assume somePromise.ANYTHING() is a new promise
    {
      code: 'frank().catch(go).someOtherMethod()',
      errors: [
        {
          messageId: 'terminationMethod',
          message: 'Expected catch() or return',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },

    // .then(null, fn) without allowThen
    {
      code: 'frank().then(a).then(b).then(null, c)',
      errors: [
        {
          messageId: 'terminationMethod',
          message: 'Expected catch() or return',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      code: 'frank().then(a).then(b).then().then().then(null, doIt)',
      errors: [
        {
          messageId: 'terminationMethod',
          message: 'Expected catch() or return',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 55,
        },
      ],
    },
    {
      code: 'frank().then(a).then(b).then(null, function() { /* why bother */ })',
      errors: [
        {
          messageId: 'terminationMethod',
          message: 'Expected catch() or return',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 68,
        },
      ],
    },

    // .then(fn, fn) without allowThen
    {
      code: 'frank().then(a, b)',
      errors: [
        {
          messageId: 'terminationMethod',
          message: 'Expected catch() or return',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'frank().then(go).then(zam, doIt)',
      errors: [
        {
          messageId: 'terminationMethod',
          message: 'Expected catch() or return',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: 'frank().then(a).then(b).then(c, d)',
      errors: [
        {
          messageId: 'terminationMethod',
          message: 'Expected catch() or return',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: 'frank().then(go).then().then().then().then(wham, doIt)',
      errors: [
        {
          messageId: 'terminationMethod',
          message: 'Expected catch() or return',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 55,
        },
      ],
    },
    {
      code: 'frank().then(go).then().then(function() {}, function() { /* why bother */ })',
      errors: [
        {
          messageId: 'terminationMethod',
          message: 'Expected catch() or return',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 77,
        },
      ],
    },
    {
      code: 'frank.then(go).then(to).then(pewPew, jail)',
      errors: [
        {
          messageId: 'terminationMethod',
          message: 'Expected catch() or return',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 43,
        },
      ],
    },

    // .then(fn, fn) with allowThenStrict (non-null first arg still invalid)
    {
      code: 'frank().then(a, b)',
      options: [{ allowThenStrict: true }],
      errors: [
        {
          messageId: 'terminationMethod',
          message: 'Expected catch() or return',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'frank().then(go).then(zam, doIt)',
      options: [{ allowThenStrict: true }],
      errors: [
        {
          messageId: 'terminationMethod',
          message: 'Expected catch() or return',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: 'frank().then(a).then(b).then(c, d)',
      options: [{ allowThenStrict: true }],
      errors: [
        {
          messageId: 'terminationMethod',
          message: 'Expected catch() or return',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: 'frank().then(go).then().then().then().then(wham, doIt)',
      options: [{ allowThenStrict: true }],
      errors: [
        {
          messageId: 'terminationMethod',
          message: 'Expected catch() or return',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 55,
        },
      ],
    },
    {
      code: 'frank().then(go).then().then(function() {}, function() { /* why bother */ })',
      options: [{ allowThenStrict: true }],
      errors: [
        {
          messageId: 'terminationMethod',
          message: 'Expected catch() or return',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 77,
        },
      ],
    },
    {
      code: 'frank.then(go).then(to).then(pewPew, jail)',
      options: [{ allowThenStrict: true }],
      errors: [
        {
          messageId: 'terminationMethod',
          message: 'Expected catch() or return',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 43,
        },
      ],
    },
  ],
});

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const catchMessage = 'Expected catch() or return';
const doneMessage = 'Expected done() or return';

ruleTester.run('catch-or-return', {} as never, {
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
      errors: [{ message: catchMessage }],
    },
    {
      code: 'fetch("http://www.yahoo.com").then(console.log.bind(console))',
      errors: [{ message: catchMessage }],
    },
    {
      code: 'a.then(function() { return "x"; }).then(function(y) { throw y; })',
      errors: [{ message: catchMessage }],
    },
    { code: 'Promise.resolve(frank)', errors: [{ message: catchMessage }] },
    { code: 'Promise.all([])', errors: [{ message: catchMessage }] },
    { code: 'Promise.allSettled([])', errors: [{ message: catchMessage }] },
    { code: 'Promise.any([])', errors: [{ message: catchMessage }] },
    { code: 'Promise.race([])', errors: [{ message: catchMessage }] },
    {
      code: 'frank().then(to).catch(fn).then(foo)',
      errors: [{ message: catchMessage }],
    },
    { code: 'frank().finally(fn)', errors: [{ message: catchMessage }] },
    {
      code: 'frank().then(to).finally(fn)',
      errors: [{ message: catchMessage }],
    },
    {
      code: 'frank().then(go).catch(doIt).finally(fn)',
      errors: [{ message: catchMessage }],
    },
    {
      code: 'frank().then(go).then().then().then().catch(doIt).finally(fn)',
      errors: [{ message: catchMessage }],
    },
    {
      code: 'frank().then(go).then().catch(function() { /* why bother */ }).finally(fn)',
      errors: [{ message: catchMessage }],
    },

    // return failures
    {
      code: 'function a() { frank().then(go) }',
      errors: [{ message: catchMessage }],
    },
    {
      code: 'function a() { frank().then(go).then().then().then() }',
      errors: [{ message: catchMessage }],
    },
    {
      code: 'function a() { frank().then(go).then()}',
      errors: [{ message: catchMessage }],
    },
    {
      code: 'function a() { frank.then(go).then(to) }',
      errors: [{ message: catchMessage }],
    },

    // allowFinally=true failures
    {
      code: 'frank().then(go).catch(doIt).finally(fn).then(foo)',
      options: [{ allowFinally: true }],
      errors: [{ message: catchMessage }],
    },
    {
      code: 'frank().then(go).catch(doIt).finally(fn).foobar(foo)',
      options: [{ allowFinally: true }],
      errors: [{ message: catchMessage }],
    },

    // terminationMethod=done failures
    {
      code: 'frank().then(go)',
      options: [{ terminationMethod: 'done' }],
      errors: [{ message: doneMessage }],
    },
    {
      code: 'frank().catch(go)',
      options: [{ terminationMethod: 'done' }],
      errors: [{ message: doneMessage }],
    },

    // assume somePromise.ANYTHING() is a new promise
    {
      code: 'frank().catch(go).someOtherMethod()',
      errors: [{ message: catchMessage }],
    },

    // .then(null, fn) without allowThen
    {
      code: 'frank().then(a).then(b).then(null, c)',
      errors: [{ message: catchMessage }],
    },
    {
      code: 'frank().then(a).then(b).then().then().then(null, doIt)',
      errors: [{ message: catchMessage }],
    },
    {
      code: 'frank().then(a).then(b).then(null, function() { /* why bother */ })',
      errors: [{ message: catchMessage }],
    },

    // .then(fn, fn) without allowThen
    { code: 'frank().then(a, b)', errors: [{ message: catchMessage }] },
    {
      code: 'frank().then(go).then(zam, doIt)',
      errors: [{ message: catchMessage }],
    },
    {
      code: 'frank().then(a).then(b).then(c, d)',
      errors: [{ message: catchMessage }],
    },
    {
      code: 'frank().then(go).then().then().then().then(wham, doIt)',
      errors: [{ message: catchMessage }],
    },
    {
      code: 'frank().then(go).then().then(function() {}, function() { /* why bother */ })',
      errors: [{ message: catchMessage }],
    },
    {
      code: 'frank.then(go).then(to).then(pewPew, jail)',
      errors: [{ message: catchMessage }],
    },

    // .then(fn, fn) with allowThenStrict (non-null first arg still invalid)
    {
      code: 'frank().then(a, b)',
      options: [{ allowThenStrict: true }],
      errors: [{ message: catchMessage }],
    },
    {
      code: 'frank().then(go).then(zam, doIt)',
      options: [{ allowThenStrict: true }],
      errors: [{ message: catchMessage }],
    },
    {
      code: 'frank().then(a).then(b).then(c, d)',
      options: [{ allowThenStrict: true }],
      errors: [{ message: catchMessage }],
    },
    {
      code: 'frank().then(go).then().then().then().then(wham, doIt)',
      options: [{ allowThenStrict: true }],
      errors: [{ message: catchMessage }],
    },
    {
      code: 'frank().then(go).then().then(function() {}, function() { /* why bother */ })',
      options: [{ allowThenStrict: true }],
      errors: [{ message: catchMessage }],
    },
    {
      code: 'frank.then(go).then(to).then(pewPew, jail)',
      options: [{ allowThenStrict: true }],
      errors: [{ message: catchMessage }],
    },
  ],
});

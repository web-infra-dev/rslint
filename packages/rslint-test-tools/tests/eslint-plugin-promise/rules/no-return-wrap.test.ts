import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const rejectMessage = 'Expected throw instead of Promise.reject';
const resolveMessage = 'Avoid wrapping return values in Promise.resolve';

ruleTester.run('no-return-wrap', {} as never, {
  valid: [
    // ESLint upstream
    { code: 'Promise.resolve(4).then(function(x) { return x })' },
    { code: 'Promise.reject(4).then(function(x) { return x })' },
    { code: 'Promise.resolve(4).then(function() {})' },
    { code: 'Promise.reject(4).then(function() {})' },
    { code: 'doThing().then(function() { return 4 })' },
    { code: 'doThing().then(function() { throw 4 })' },
    { code: 'doThing().then(null, function() { return 4 })' },
    { code: 'doThing().then(null, function() { throw 4 })' },
    { code: 'doThing().catch(null, function() { return 4 })' },
    { code: 'doThing().catch(null, function() { throw 4 })' },
    { code: 'doThing().then(function() { return Promise.all([a,b,c]) })' },
    { code: 'doThing().then(() => 4)' },
    { code: 'doThing().then(() => { throw 4 })' },
    { code: 'doThing().then(()=>{}, () => 4)' },
    { code: 'doThing().then(()=>{}, () => { throw 4 })' },
    { code: 'doThing().catch(() => 4)' },
    { code: 'doThing().catch(() => { throw 4 })' },
    { code: 'var x = function() { return Promise.resolve(4) }' },
    { code: 'function y() { return Promise.resolve(4) }' },
    { code: 'function then() { return Promise.reject() }' },
    { code: 'doThing(function(x) { return Promise.reject(x) })' },
    { code: 'doThing().then(function() { return })' },
    {
      code: 'doThing().then(function() { return Promise.reject(4) })',
      options: [{ allowReject: true }],
    },
    {
      code: 'doThing().then((function() { return Promise.resolve(4) }).toString())',
    },
    {
      code: 'doThing().then(() => Promise.reject(4))',
      options: [{ allowReject: true }],
    },
    { code: 'doThing().then(function() { return a() })' },
    { code: 'doThing().then(function() { return Promise.a() })' },
    { code: 'doThing().then(() => { return a() })' },
    { code: 'doThing().then(() => { return Promise.a() })' },
    { code: 'doThing().then(() => a())' },
    { code: 'doThing().then(() => Promise.a())' },

    // Additional port-shape coverage.
    { code: "doThing().then(function() { return Promise['resolve'](4) })" },
    { code: "doThing()['then'](function() { return Promise.resolve(4) })" },
    { code: 'Promise.withResolvers(function() { return Promise.resolve(4) })' },
    {
      code: 'doThing().then(function() { fn(function() { return Promise.resolve(4) }); return 1 })',
    },

    // Upstream alignment: nested methods/accessors/constructors are separate
    // function boundaries and should not be attributed to the outer promise callback.
    {
      code: 'doThing().then(function() { class Foo { method() { return Promise.resolve(4) } } return new Foo() })',
    },
    {
      code: 'doThing().then(function() { class Foo { static method() { return Promise.resolve(4) } } return Foo })',
    },
    {
      code: 'doThing().then(function() { class Foo { get x() { return Promise.resolve(4) } } return new Foo() })',
    },
    {
      code: 'doThing().then(function() { class Foo { set x(value) { return Promise.resolve(value) } } return new Foo() })',
    },
    {
      code: 'doThing().then(function() { class Foo { constructor() { return Promise.resolve(4) } } return new Foo() })',
    },
    {
      code: 'doThing().then(function() { return { m() { return Promise.resolve(4) } } })',
    },
    {
      code: 'doThing().then(function() { return { get x() { return Promise.resolve(4) } } })',
    },
    {
      code: 'doThing().then(function() { return { set x(value) { return Promise.resolve(value) } } })',
    },
  ],

  invalid: [
    // ESLint upstream
    {
      code: 'doThing().then(function() { return Promise.resolve(4) })',
      errors: [{ message: resolveMessage }],
    },
    {
      code: 'doThing().then(null, function() { return Promise.resolve(4) })',
      errors: [{ message: resolveMessage }],
    },
    {
      code: 'doThing().catch(function() { return Promise.resolve(4) })',
      errors: [{ message: resolveMessage }],
    },
    {
      code: 'doThing().then(function() { return Promise.reject(4) })',
      errors: [{ message: rejectMessage }],
    },
    {
      code: 'doThing().then(null, function() { return Promise.reject(4) })',
      errors: [{ message: rejectMessage }],
    },
    {
      code: 'doThing().catch(function() { return Promise.reject(4) })',
      errors: [{ message: rejectMessage }],
    },
    {
      code: 'doThing().then(function(x) { if (x>1) { return Promise.resolve(4) } else { throw "bad" } })',
      errors: [{ message: resolveMessage }],
    },
    {
      code: 'doThing().then(function(x) { if (x>1) { return Promise.reject(4) } })',
      errors: [{ message: rejectMessage }],
    },
    {
      code: 'doThing().then(null, function() { if (true && false) { return Promise.resolve() } })',
      errors: [{ message: resolveMessage }],
    },
    {
      code: 'doThing().catch(function(x) {if (x) { return Promise.resolve(4) } else { return Promise.reject() } })',
      errors: [{ message: resolveMessage }, { message: rejectMessage }],
    },
    {
      code: `
      fn(function() {
        doThing().then(function() {
          return Promise.resolve(4)
        })
        return
      })`,
      errors: [{ message: resolveMessage }],
    },
    {
      code: `
      fn(function() {
        doThing().then(function nm() {
          return Promise.resolve(4)
        })
        return
      })`,
      errors: [{ message: resolveMessage }],
    },
    {
      code: `
      fn(function() {
        fn2(function() {
          doThing().then(function() {
            return Promise.resolve(4)
          })
        })
      })`,
      errors: [{ message: resolveMessage }],
    },
    {
      code: `
      fn(function() {
        fn2(function() {
          doThing().then(function() {
            fn3(function() {
              return Promise.resolve(4)
            })
            return Promise.resolve(4)
          })
        })
      })`,
      errors: [{ message: resolveMessage }],
    },
    {
      code: `
      const o = {
        fn: function() {
          return doThing().then(function() {
            return Promise.resolve(5);
          });
        },
      }
      `,
      errors: [{ message: resolveMessage }],
    },
    {
      code: `
      fn(
        doThing().then(function() {
          return Promise.resolve(5);
        })
      );
      `,
      errors: [{ message: resolveMessage }],
    },
    {
      code: 'doThing().then((function() { return Promise.resolve(4) }).bind(this))',
      errors: [{ message: resolveMessage }],
    },
    {
      code: 'doThing().then((function() { return Promise.resolve(4) }).bind(this).bind(this))',
      errors: [{ message: resolveMessage }],
    },
    {
      code: 'doThing().then(() => { return Promise.resolve(4) })',
      errors: [{ message: resolveMessage }],
    },
    {
      code: `
      function a () {
        return p.then(function(val) {
          return Promise.resolve(val * 4)
        })
      }
      `,
      errors: [{ message: resolveMessage }],
    },
    {
      code: 'doThing().then(() => Promise.resolve(4))',
      errors: [{ message: resolveMessage }],
    },
    {
      code: 'doThing().then(() => Promise.reject(4))',
      errors: [{ message: rejectMessage }],
    },

    // Additional upstream semantic branches.
    {
      code: 'Promise.all(xs).then(function() { return Promise.resolve(4) })',
      errors: [{ message: resolveMessage }],
    },
    {
      code: 'Promise.withResolvers().then(function() { return Promise.resolve(4) })',
      errors: [{ message: resolveMessage }],
    },
  ],
});

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-promise-reject-errors', {
  valid: [
    'Promise.resolve(5)',
    'Foo.reject(5)',
    'Promise.reject(foo)',
    'Promise.reject(foo.bar)',
    'Promise.reject(foo.bar())',
    'Promise.reject(new Error())',
    'Promise.reject(new TypeError)',
    "Promise.reject(new Error('foo'))",
    'Promise.reject(foo || 5)',
    'Promise.reject(5 && foo)',
    'new Foo((resolve, reject) => reject(5))',
    'new Promise(function(resolve, reject) { return function(reject) { reject(5) } })',
    'new Promise(function(resolve, reject) { if (foo) { const reject = somethingElse; reject(5) } })',
    'new Promise(function(resolve, {apply}) { apply(5) })',
    'new Promise(function(resolve, reject) { resolve(5, reject) })',
    'async function foo() { Promise.reject(await foo); }',
    {
      code: 'Promise.reject()',
      options: { allowEmptyReject: true },
    },
    {
      code: 'new Promise(function(resolve, reject) { reject() })',
      options: { allowEmptyReject: true },
    },

    // ---- Optional chaining ----
    'Promise.reject(obj?.foo)',
    'Promise.reject(obj?.foo())',

    // ---- Assignments ----
    'Promise.reject(foo = new Error())',
    'Promise.reject(foo ||= 5)',
    'Promise.reject(foo.bar ??= 5)',
    'Promise.reject(foo[bar] ??= 5)',

    // ---- Private fields ----
    'class C { #reject; foo() { Promise.#reject(5); } }',
    'class C { #error; foo() { Promise.reject(this.#error); } }',

    // ---- ESLint requires params[1].type === "Identifier"; non-plain
    // second-parameter shapes are not analyzed.
    'new Promise((resolve, reject = foo) => reject(5))',
    'new Promise((resolve, ...reject) => reject[0](5))',
    'new Promise(function(resolve, [reject]) { reject(5) })',

    // ---- Identifiers that resolve to globals are still Identifier nodes ----
    'Promise.reject(NaN)',
    'Promise.reject(Infinity)',

    // ---- Calling Promise without `new` is not an executor pattern. ----
    'Promise((resolve, reject) => reject(5))',

    // ---- TaggedTemplateExpression result could be an Error. ----
    'Promise.reject(tag`msg`)',

    // ---- Reject as argument, not callee, does not invoke it. ----
    'new Promise((resolve, reject) => arr.push(reject))',

    // ---- Reject method-like access (.call / .bind / .apply) is treated
    // as a different operation by ESLint and is not flagged.
    'new Promise((resolve, reject) => reject.call(null, new Error()))',
  ],
  invalid: [
    // ---- TS assertion wrappers — NOT transparent in upstream ESLint ----
    {
      code: 'Promise.reject(foo as Error)',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 1 }],
    },
    {
      code: 'Promise.reject(<Error>foo)',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 1 }],
    },
    {
      code: 'Promise.reject(foo!)',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 1 }],
    },
    {
      code: 'Promise.reject(foo satisfies Error)',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 1 }],
    },
    {
      code: 'Promise.reject(5)',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 1 }],
    },
    {
      code: "Promise.reject('foo')",
      errors: [{ messageId: 'rejectAnError', line: 1, column: 1 }],
    },
    {
      code: 'Promise.reject(`foo`)',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 1 }],
    },
    {
      code: 'Promise.reject(!foo)',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 1 }],
    },
    {
      code: 'Promise.reject(void foo)',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 1 }],
    },
    {
      code: 'Promise.reject()',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 1 }],
    },
    {
      code: 'Promise.reject(undefined)',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 1 }],
    },
    {
      code: 'Promise.reject({ foo: 1 })',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 1 }],
    },
    {
      code: 'Promise.reject([1, 2, 3])',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 1 }],
    },
    {
      code: 'Promise.reject()',
      options: { allowEmptyReject: false },
      errors: [{ messageId: 'rejectAnError', line: 1, column: 1 }],
    },
    {
      code: 'new Promise(function(resolve, reject) { reject() })',
      options: { allowEmptyReject: false },
      errors: [{ messageId: 'rejectAnError', line: 1, column: 41 }],
    },
    {
      code: 'Promise.reject(undefined)',
      options: { allowEmptyReject: true },
      errors: [{ messageId: 'rejectAnError', line: 1, column: 1 }],
    },
    {
      code: "Promise.reject('foo', somethingElse)",
      errors: [{ messageId: 'rejectAnError', line: 1, column: 1 }],
    },
    {
      code: 'new Promise(function(resolve, reject) { reject(5) })',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 41 }],
    },
    {
      code: 'new Promise((resolve, reject) => { reject(5) })',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 36 }],
    },
    {
      code: 'new Promise((resolve, reject) => reject(5))',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 34 }],
    },
    {
      code: 'new Promise((resolve, reject) => reject())',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 34 }],
    },
    {
      code: 'new Promise(function(yes, no) { no(5) })',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 33 }],
    },
    {
      code: `
          new Promise((resolve, reject) => {
            fs.readFile('foo.txt', (err, file) => {
              if (err) reject('File not found')
              else resolve(file)
            })
          })
        `,
      errors: [{ messageId: 'rejectAnError', line: 4, column: 24 }],
    },
    {
      code: 'new Promise(({foo, bar, baz}, reject) => reject(5))',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 42 }],
    },
    {
      code: 'new Promise(function(reject, reject) { reject(5) })',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 40 }],
    },
    {
      code: 'new Promise(function(foo, arguments) { arguments(5) })',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 40 }],
    },
    {
      code: 'new Promise((foo, arguments) => arguments(5))',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 33 }],
    },
    {
      code: 'new Promise(function({}, reject) { reject(5) })',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 36 }],
    },
    {
      code: 'new Promise(({}, reject) => reject(5))',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 29 }],
    },
    {
      code: 'new Promise((resolve, reject, somethingElse = reject(5)) => {})',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 47 }],
    },
    {
      code: 'new Promise(function(resolve, reject) { var reject = somethingElse; reject(5) })',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 69 }],
    },

    // ---- reject called from a nested callback inside the executor still
    // resolves to the executor's parameter binding.
    {
      code: 'new Promise((resolve, reject) => setTimeout(() => reject(5), 0))',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 51 }],
    },
    {
      code: "new Promise((resolve, reject) => arr.forEach(function () { reject('bad') }))",
      errors: [{ messageId: 'rejectAnError', line: 1, column: 60 }],
    },

    // ---- Spread cannot be statically classified as an Error candidate ----
    {
      code: 'Promise.reject(...args)',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 1 }],
    },

    // ---- Plain literals that are not Errors ----
    {
      code: 'Promise.reject(null)',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 1 }],
    },
    {
      code: 'Promise.reject(0)',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 1 }],
    },
    {
      code: 'Promise.reject(true)',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 1 }],
    },

    // ---- Computed bracket access with static string ----
    {
      code: "Promise['reject'](5)",
      errors: [{ messageId: 'rejectAnError', line: 1, column: 1 }],
    },
    {
      code: 'Promise[`reject`](5)',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 1 }],
    },

    // ---- Named function expression as executor ----
    {
      code: 'new Promise(function fn(resolve, reject) { reject(5) })',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 44 }],
    },

    // ---- Async / generator executors are still functions, so reject calls
    // inside them should be checked.
    {
      code: 'new Promise(async (resolve, reject) => reject(5))',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 40 }],
    },
    {
      code: 'new Promise(function *(resolve, reject) { reject(5) })',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 43 }],
    },

    // ---- Multiple reject calls in a single executor ----
    {
      code: "new Promise((resolve, reject) => { reject(1); reject('x'); reject(); })",
      errors: [
        { messageId: 'rejectAnError', line: 1, column: 36 },
        { messageId: 'rejectAnError', line: 1, column: 47 },
        { messageId: 'rejectAnError', line: 1, column: 60 },
      ],
    },

    // ---- Parenthesized reject callee inside the executor body ----
    {
      code: 'new Promise((resolve, reject) => (reject)(5))',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 34 }],
    },

    // ---- Optional chaining on the executor reject callback ----
    {
      code: 'new Promise((resolve, reject) => reject?.(5))',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 34 }],
    },

    // ---- Parenthesized / TS-asserted Promise constructor must be
    // transparent (ESTree strips parens, so ESLint already accepts
    // `new (Promise)(executor)`).
    {
      code: 'new (Promise)((resolve, reject) => reject(5))',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 36 }],
    },
    {
      code: 'new Promise(((resolve, reject) => reject(5)))',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 35 }],
    },

    // ---- Optional chaining ----
    {
      code: 'Promise.reject?.(5)',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 1 }],
    },
    {
      code: 'Promise?.reject(5)',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 1 }],
    },
    {
      code: 'Promise?.reject?.(5)',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 1 }],
    },
    {
      code: '(Promise?.reject)(5)',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 1 }],
    },
    {
      code: '(Promise?.reject)?.(5)',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 1 }],
    },

    // ---- Mathematical / bitwise assignments evaluate to primitives or throw ----
    {
      code: 'Promise.reject(foo += new Error())',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 1 }],
    },
    {
      code: 'Promise.reject(foo -= new Error())',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 1 }],
    },
    {
      code: 'Promise.reject(foo **= new Error())',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 1 }],
    },
    {
      code: 'Promise.reject(foo <<= new Error())',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 1 }],
    },
    {
      code: 'Promise.reject(foo |= new Error())',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 1 }],
    },
    {
      code: 'Promise.reject(foo &= new Error())',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 1 }],
    },

    // ---- && short-circuit yields the falsy left or the right operand ----
    {
      code: 'Promise.reject(foo && 5)',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 1 }],
    },
    {
      code: 'Promise.reject(foo &&= 5)',
      errors: [{ messageId: 'rejectAnError', line: 1, column: 1 }],
    },
  ],
});

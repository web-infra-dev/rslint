import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('require-await', {
  valid: [
    `async function foo() { await doSomething() }`,
    `(async function() { await doSomething() })`,
    `async () => { await doSomething() }`,
    `async () => await doSomething()`,
    `({ async foo() { await doSomething() } })`,
    `class A { async foo() { await doSomething() } }`,
    `(class { async foo() { await doSomething() } })`,
    `async function foo() { await (async () => { await doSomething() }) }`,
    `async function foo() {}`,
    `async () => {}`,
    `function foo() { doSomething() }`,
    `async function foo() { for await (x of xs); }`,
    `await foo()`,
    `
for await (let num of asyncIterable) {
    console.log(num);
}
`,
    `async function* run() { yield * anotherAsyncGenerator() }`,
    `
async function* run() {
    await new Promise(resolve => setTimeout(resolve, 100));
    yield 'Hello';
    console.log('World');
}
`,
    `async function* run() { }`,
    `const foo = async function *(){}`,
    `const foo = async function *(){ console.log("bar") }`,
    `async function* run() { console.log("bar") }`,
    `await using resource = getResource();`,
    `async function run() { await using resource = getResource(); }`,
  ],
  invalid: [
    {
      code: `async function foo() { doSomething() }`,
      errors: [{ messageId: 'missingAwait' }],
    },
    {
      code: `(async function() { doSomething() })`,
      errors: [{ messageId: 'missingAwait' }],
    },
    {
      code: `async () => { doSomething() }`,
      errors: [{ messageId: 'missingAwait' }],
    },
    {
      code: `async () => doSomething()`,
      errors: [{ messageId: 'missingAwait' }],
    },
    {
      code: `({ async foo() { doSomething() } })`,
      errors: [{ messageId: 'missingAwait' }],
    },
    {
      code: `class A { async foo() { doSomething() } }`,
      errors: [{ messageId: 'missingAwait' }],
    },
    {
      code: `(class { async foo() { doSomething() } })`,
      errors: [{ messageId: 'missingAwait' }],
    },
    {
      code: `(class { async ''() { doSomething() } })`,
      errors: [{ messageId: 'missingAwait' }],
    },
    {
      code: `async function foo() { async () => { await doSomething() } }`,
      errors: [{ messageId: 'missingAwait' }],
    },
    {
      code: `async function foo() { await (async () => { doSomething() }) }`,
      errors: [{ messageId: 'missingAwait' }],
    },
    {
      code: `const obj = { async: async function foo() { bar(); } }`,
      errors: [{ messageId: 'missingAwait' }],
    },
    {
      code: `async    /* test */ function foo() { doSomething() }`,
      errors: [{ messageId: 'missingAwait' }],
    },
    {
      code: `class A {
    a = 0
    async [b](){ return 0; }
}`,
      errors: [{ messageId: 'missingAwait' }],
    },
    {
      code: `class A {
    a
    async [b](){ return 0; }
}`,
      errors: [{ messageId: 'missingAwait' }],
    },
    {
      code: `class A {
    a = 0
    async in(){ return 0; }
}`,
      errors: [{ messageId: 'missingAwait' }],
    },
    {
      code: `const obj = {
    foo,
    async in(){ return 0; }
}`,
      errors: [{ messageId: 'missingAwait' }],
    },
    {
      code: `foo
    async () => { return 0; }
`,
      errors: [{ messageId: 'missingAwait' }],
    },
    {
      code: `class A {
    foo() {}
    async [bar] () { baz; }
}`,
      errors: [{ messageId: 'missingAwait' }],
    },
    {
      code: `async function run() { using resource = getResource(); }`,
      errors: [{ messageId: 'missingAwait' }],
    },
  ],
});

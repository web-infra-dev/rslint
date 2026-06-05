import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const message = 'Prefer await to then()/catch()/finally().';

ruleTester.run('prefer-await-to-then', {} as never, {
  valid: [
    { code: 'async function hi() { await thing() }' },
    { code: 'async function hi() { await thing().then() }' },
    { code: 'async function hi() { await thing().catch() }' },
    { code: 'async function hi() { await thing().finally() }' },
    // Cypress
    { code: 'function hi() { cy.get(".myClass").then(go) }' },
    { code: 'function hi() { cy.get("button").click().then() }' },
    { code: 'function * hi() { yield thing().then() }' },
    { code: 'something().then(async () => await somethingElse())' },
    { code: 'function foo() { hey.somethingElse(x => {}) }' },
    {
      code: 'class Foo {\n  constructor () {\n    doSomething.then(abc);\n  }\n}',
    },
  ],

  invalid: [
    {
      code: 'function foo() { hey.then(x => {}) }',
      errors: [{ message }],
    },
    {
      code: 'function foo() { hey.then(function() { }).then() }',
      errors: [{ message }, { message }],
    },
    {
      code: 'function foo() { hey.then(function() { }).then(x).catch() }',
      errors: [{ message }, { message }, { message }],
    },
    {
      code: 'async function a() { hey.then(function() { }).then(function() { }) }',
      errors: [{ message }, { message }],
    },
    {
      code: 'function foo() { hey.catch(x => {}) }',
      errors: [{ message }],
    },
    {
      code: 'function foo() { hey.finally(x => {}) }',
      errors: [{ message }],
    },
    {
      code: 'async function hi() { await thing().then() }',
      options: [{ strict: true }],
      errors: [{ message }],
    },
    {
      code: 'class Foo {\n  constructor () {\n    doSomething.then(abc);\n  }\n}',
      options: [{ strict: true }],
      errors: [{ message }],
    },
    {
      code: 'async function hi() { await thing().catch() }',
      options: [{ strict: true }],
      errors: [{ message }],
    },
    {
      code: 'async function hi() { await thing().finally() }',
      options: [{ strict: true }],
      errors: [{ message }],
    },
    {
      code: 'function * hi() { yield thing().then() }',
      options: [{ strict: true }],
      errors: [{ message }],
    },
  ],
});

import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('prefer-await-to-then', {
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
      errors: [
        {
          messageId: 'preferAwaitToCallback',
          message: 'Prefer await to then()/catch()/finally().',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: 'function foo() { hey.then(function() { }).then() }',
      errors: [
        {
          messageId: 'preferAwaitToCallback',
          message: 'Prefer await to then()/catch()/finally().',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 26,
        },
        {
          messageId: 'preferAwaitToCallback',
          message: 'Prefer await to then()/catch()/finally().',
          line: 1,
          column: 43,
          endLine: 1,
          endColumn: 47,
        },
      ],
    },
    {
      code: 'function foo() { hey.then(function() { }).then(x).catch() }',
      errors: [
        {
          messageId: 'preferAwaitToCallback',
          message: 'Prefer await to then()/catch()/finally().',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 26,
        },
        {
          messageId: 'preferAwaitToCallback',
          message: 'Prefer await to then()/catch()/finally().',
          line: 1,
          column: 43,
          endLine: 1,
          endColumn: 47,
        },
        {
          messageId: 'preferAwaitToCallback',
          message: 'Prefer await to then()/catch()/finally().',
          line: 1,
          column: 51,
          endLine: 1,
          endColumn: 56,
        },
      ],
    },
    {
      code: 'async function a() { hey.then(function() { }).then(function() { }) }',
      errors: [
        {
          messageId: 'preferAwaitToCallback',
          message: 'Prefer await to then()/catch()/finally().',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 30,
        },
        {
          messageId: 'preferAwaitToCallback',
          message: 'Prefer await to then()/catch()/finally().',
          line: 1,
          column: 47,
          endLine: 1,
          endColumn: 51,
        },
      ],
    },
    {
      code: 'function foo() { hey.catch(x => {}) }',
      errors: [
        {
          messageId: 'preferAwaitToCallback',
          message: 'Prefer await to then()/catch()/finally().',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: 'function foo() { hey.finally(x => {}) }',
      errors: [
        {
          messageId: 'preferAwaitToCallback',
          message: 'Prefer await to then()/catch()/finally().',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: 'async function hi() { await thing().then() }',
      options: [{ strict: true }],
      errors: [
        {
          messageId: 'preferAwaitToCallback',
          message: 'Prefer await to then()/catch()/finally().',
          line: 1,
          column: 37,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      code: 'class Foo {\n  constructor () {\n    doSomething.then(abc);\n  }\n}',
      options: [{ strict: true }],
      errors: [
        {
          messageId: 'preferAwaitToCallback',
          message: 'Prefer await to then()/catch()/finally().',
          line: 3,
          column: 17,
          endLine: 3,
          endColumn: 21,
        },
      ],
    },
    {
      code: 'async function hi() { await thing().catch() }',
      options: [{ strict: true }],
      errors: [
        {
          messageId: 'preferAwaitToCallback',
          message: 'Prefer await to then()/catch()/finally().',
          line: 1,
          column: 37,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      code: 'async function hi() { await thing().finally() }',
      options: [{ strict: true }],
      errors: [
        {
          messageId: 'preferAwaitToCallback',
          message: 'Prefer await to then()/catch()/finally().',
          line: 1,
          column: 37,
          endLine: 1,
          endColumn: 44,
        },
      ],
    },
    {
      code: 'function * hi() { yield thing().then() }',
      options: [{ strict: true }],
      errors: [
        {
          messageId: 'preferAwaitToCallback',
          message: 'Prefer await to then()/catch()/finally().',
          line: 1,
          column: 33,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
  ],
});

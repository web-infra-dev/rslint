import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('no-nesting', {
  valid: [
    // resolve and reject are sometimes okay
    { code: 'Promise.resolve(4).then(function(x) { return x })' },
    { code: 'Promise.reject(4).then(function(x) { return x })' },
    { code: 'Promise.resolve(4).then(function() {})' },
    { code: 'Promise.reject(4).then(function() {})' },

    // throw and return are fine
    { code: 'doThing().then(function() { return 4 })' },
    { code: 'doThing().then(function() { throw 4 })' },
    { code: 'doThing().then(null, function() { return 4 })' },
    { code: 'doThing().then(null, function() { throw 4 })' },
    { code: 'doThing().catch(null, function() { return 4 })' },
    { code: 'doThing().catch(null, function() { throw 4 })' },

    // arrow functions and other things
    { code: 'doThing().then(() => 4)' },
    { code: 'doThing().then(() => { throw 4 })' },
    { code: 'doThing().then(()=>{}, () => 4)' },
    { code: 'doThing().then(()=>{}, () => { throw 4 })' },
    { code: 'doThing().catch(() => 4)' },
    { code: 'doThing().catch(() => { throw 4 })' },

    // random functions and callback methods
    { code: 'var x = function() { return Promise.resolve(4) }' },
    { code: 'function y() { return Promise.resolve(4) }' },
    { code: 'function then() { return Promise.reject() }' },
    { code: 'doThing(function(x) { return Promise.reject(x) })' },

    // Promise statics and Promise.all are fine inside callbacks
    { code: 'doThing().then(function() { return Promise.all([a,b,c]) })' },
    { code: 'doThing().then(function() { return Promise.resolve(4) })' },
    { code: 'doThing().then(() => Promise.resolve(4))' },
    { code: 'doThing().then(() => Promise.all([a]))' },

    // references vars in closure
    { code: 'doThing().then(a => getB(a).then(b => getC(a, b)))' },
    {
      code: `doThing().then(a => {
        const c = a * 2;
        return getB(c).then(b => getC(c, b))
      })`,
    },
  ],

  invalid: [
    {
      code: 'doThing().then(function() { a.then() })',
      errors: [
        {
          messageId: 'avoidNesting',
          message: 'Avoid nesting promises.',
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: 'doThing().then(function() { b.catch() })',
      errors: [
        {
          messageId: 'avoidNesting',
          message: 'Avoid nesting promises.',
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      code: 'doThing().then(function() { return a.then() })',
      errors: [
        {
          messageId: 'avoidNesting',
          message: 'Avoid nesting promises.',
          line: 1,
          column: 38,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      code: 'doThing().then(function() { return b.catch() })',
      errors: [
        {
          messageId: 'avoidNesting',
          message: 'Avoid nesting promises.',
          line: 1,
          column: 38,
          endLine: 1,
          endColumn: 43,
        },
      ],
    },
    {
      code: 'doThing().then(() => { a.then() })',
      errors: [
        {
          messageId: 'avoidNesting',
          message: 'Avoid nesting promises.',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: 'doThing().then(() => { b.catch() })',
      errors: [
        {
          messageId: 'avoidNesting',
          message: 'Avoid nesting promises.',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: 'doThing().then(() => a.then())',
      errors: [
        {
          messageId: 'avoidNesting',
          message: 'Avoid nesting promises.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: 'doThing().then(() => b.catch())',
      errors: [
        {
          messageId: 'avoidNesting',
          message: 'Avoid nesting promises.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    // references vars in closure
    {
      code: `
      doThing()
        .then(a => getB(a)
          .then(b => getC(b))
        )`,
      errors: [
        {
          messageId: 'avoidNesting',
          message: 'Avoid nesting promises.',
          line: 4,
          column: 12,
          endLine: 4,
          endColumn: 16,
        },
      ],
    },
    {
      code: `
      doThing()
        .then(a => getB(a)
          .then(b => getC(a, b)
            .then(c => getD(a, c))
          )
        )`,
      errors: [
        {
          messageId: 'avoidNesting',
          message: 'Avoid nesting promises.',
          line: 5,
          column: 14,
          endLine: 5,
          endColumn: 18,
        },
      ],
    },
  ],
});

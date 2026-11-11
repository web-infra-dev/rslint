import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('no-new-statics', {
  valid: [
    // ESLint upstream
    { code: 'Promise.resolve()' },
    { code: 'Promise.reject()' },
    { code: 'Promise.all()' },
    { code: 'Promise.race()' },
    { code: 'Promise.withResolvers()' },
    { code: 'new Promise(function (resolve, reject) {})' },
    { code: 'new SomeClass()' },
    { code: 'SomeClass.resolve()' },
    { code: 'new SomeClass.resolve()' },
  ],

  invalid: [
    // ESLint upstream
    {
      code: 'new Promise.resolve()',
      errors: [
        {
          messageId: 'avoidNewStatic',
          message: "Avoid calling 'new' on 'Promise.resolve()'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
      output: 'Promise.resolve()',
    },
    {
      code: 'new Promise.reject()',
      errors: [
        {
          messageId: 'avoidNewStatic',
          message: "Avoid calling 'new' on 'Promise.reject()'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
      output: 'Promise.reject()',
    },
    {
      code: 'new Promise.all()',
      errors: [
        {
          messageId: 'avoidNewStatic',
          message: "Avoid calling 'new' on 'Promise.all()'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
        },
      ],
      output: 'Promise.all()',
    },
    {
      code: 'new Promise.allSettled()',
      errors: [
        {
          messageId: 'avoidNewStatic',
          message: "Avoid calling 'new' on 'Promise.allSettled()'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
      output: 'Promise.allSettled()',
    },
    {
      code: 'new Promise.any()',
      errors: [
        {
          messageId: 'avoidNewStatic',
          message: "Avoid calling 'new' on 'Promise.any()'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
        },
      ],
      output: 'Promise.any()',
    },
    {
      code: 'new Promise.race()',
      errors: [
        {
          messageId: 'avoidNewStatic',
          message: "Avoid calling 'new' on 'Promise.race()'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
      output: 'Promise.race()',
    },
    {
      code: 'new Promise.withResolvers()',
      errors: [
        {
          messageId: 'avoidNewStatic',
          message: "Avoid calling 'new' on 'Promise.withResolvers()'",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
      output: 'Promise.withResolvers()',
    },
    {
      code: 'function a() { return new Promise.resolve(a) }',
      errors: [
        {
          messageId: 'avoidNewStatic',
          message: "Avoid calling 'new' on 'Promise.resolve()'",
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 45,
        },
      ],
      output: 'function a() { return Promise.resolve(a) }',
    },
  ],
});

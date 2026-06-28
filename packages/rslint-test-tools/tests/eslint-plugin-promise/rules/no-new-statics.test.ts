import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const msg = (name: string) => `Avoid calling 'new' on 'Promise.${name}()'`;

ruleTester.run('no-new-statics', {} as never, {
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
      errors: [{ message: msg('resolve') }],
    },
    {
      code: 'new Promise.reject()',
      errors: [{ message: msg('reject') }],
    },
    {
      code: 'new Promise.all()',
      errors: [{ message: msg('all') }],
    },
    {
      code: 'new Promise.allSettled()',
      errors: [{ message: msg('allSettled') }],
    },
    {
      code: 'new Promise.any()',
      errors: [{ message: msg('any') }],
    },
    {
      code: 'new Promise.race()',
      errors: [{ message: msg('race') }],
    },
    {
      code: 'new Promise.withResolvers()',
      errors: [{ message: msg('withResolvers') }],
    },
    {
      code: 'function a() { return new Promise.resolve(a) }',
      errors: [{ message: msg('resolve') }],
    },
  ],
});

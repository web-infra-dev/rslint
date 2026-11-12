import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('avoid-new', {
  valid: [
    { code: 'Promise.resolve()' },
    { code: 'Promise.reject()' },
    { code: 'Promise.all()' },
    { code: 'new Horse()' },
    { code: 'new PromiseLikeThing()' },
    { code: 'new Promise.resolve()' },
  ],

  invalid: [
    {
      code: 'var x = new Promise(function (x, y) {})',
      errors: [
        {
          messageId: 'avoidNew',
          message: 'Avoid creating new promises.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      code: 'new Promise()',
      errors: [
        {
          messageId: 'avoidNew',
          message: 'Avoid creating new promises.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'Thing(new Promise(() => {}))',
      errors: [
        {
          messageId: 'avoidNew',
          message: 'Avoid creating new promises.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
  ],
});

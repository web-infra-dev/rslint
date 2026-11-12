import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('no-multiple-resolved', {
  valid: [
    // if-else mutual exclusion
    {
      code: `new Promise((resolve, reject) => {
        if (error) { reject(error) } else { resolve(value) }
      })`,
    },
    // early return after reject
    {
      code: `new Promise((resolve, reject) => {
        if (error) { reject(error); return }
        resolve(value)
      })`,
    },
    // async try-catch, resolve is last throwable
    {
      code: `new Promise(async (resolve, reject) => {
        try { const r = await foo(); resolve(r); } catch (e) { reject(e); }
      })`,
    },
  ],

  invalid: [
    // sequential reject then resolve
    {
      code: `new Promise((resolve, reject) => {
        reject(error)
        resolve(value)
      })`,
      errors: [
        {
          messageId: '',
          message:
            'Promise should not be resolved multiple times. Promise is already resolved on line 2.',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 23,
        },
      ],
    },
    // if-without-else then resolve
    {
      code: `new Promise((resolve, reject) => {
        if (error) { reject(error) }
        resolve(value)
      })`,
      errors: [
        {
          messageId: '',
          message:
            'Promise should not be resolved multiple times. Promise is potentially resolved on line 2.',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 23,
        },
      ],
    },
    // while loop then resolve
    {
      code: `new Promise((resolve, reject) => {
        while (error) { reject(error) }
        resolve(value)
      })`,
      errors: [
        {
          messageId: '',
          message:
            'Promise should not be resolved multiple times. Promise is potentially resolved on line 2.',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 23,
        },
      ],
    },
  ],
});

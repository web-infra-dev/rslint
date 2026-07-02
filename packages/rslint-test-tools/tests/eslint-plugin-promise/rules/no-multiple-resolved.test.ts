import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-multiple-resolved', {} as never, {
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
          message:
            'Promise should not be resolved multiple times. Promise is already resolved on line 2.',
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
          message:
            'Promise should not be resolved multiple times. Promise is potentially resolved on line 2.',
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
          message:
            'Promise should not be resolved multiple times. Promise is potentially resolved on line 2.',
        },
      ],
    },
  ],
});

import { RuleTester } from '@typescript-eslint/rule-tester';

import { getFixturesRootDir } from '../RuleTester';

const rootDir = getFixturesRootDir();
const ruleTester = new RuleTester({
  languageOptions: {
    parserOptions: {
      project: './tsconfig.json',
      tsconfigRootDir: rootDir,
    },
  },
});

ruleTester.run('ban-tslint-comment', {
  valid: [
    {
      code: 'let a: readonly any[] = [];',
    },
    {
      code: 'let a = new Array();',
    },
    {
      code: '// some other comment',
    },
    {
      code: '// TODO: this is a comment that mentions tslint',
    },
    {
      code: '/* another comment that mentions tslint */',
    },
    {
      code: '// This project used to use tslint',
    },
    {
      code: '/* We migrated from tslint to eslint */',
    },
    {
      code: '// tslint is deprecated',
    },
    {
      code: '/* tslint was a linter */',
    },
    {
      code: '// about tslint:disable',
    },
    {
      code: '/* discussing tslint:enable */',
    },
  ],
  invalid: [
    {
      code: '/* tslint:disable */',
      errors: [
        {
          column: 1,
          data: {
            text: '/* tslint:disable */',
          },
          line: 1,
          messageId: 'commentDetected',
        },
      ],
      output: '',
    },
    {
      code: '/* tslint:enable */',
      errors: [
        {
          column: 1,
          data: {
            text: '/* tslint:enable */',
          },
          line: 1,
          messageId: 'commentDetected',
        },
      ],
      output: '',
    },
    {
      code: '/* tslint:disable:rule1 rule2 rule3... */',
      errors: [
        {
          column: 1,
          data: {
            text: '/* tslint:disable:rule1 rule2 rule3... */',
          },
          line: 1,
          messageId: 'commentDetected',
        },
      ],
      output: '',
    },
    {
      code: '/* tslint:enable:rule1 rule2 rule3... */',
      errors: [
        {
          column: 1,
          data: {
            text: '/* tslint:enable:rule1 rule2 rule3... */',
          },
          line: 1,
          messageId: 'commentDetected',
        },
      ],
      output: '',
    },
    {
      code: '// tslint:disable-next-line',
      errors: [
        {
          column: 1,
          data: {
            text: '// tslint:disable-next-line',
          },
          line: 1,
          messageId: 'commentDetected',
        },
      ],
      output: '',
    },
    {
      code: 'someCode(); // tslint:disable-line',
      errors: [
        {
          column: 13,
          data: {
            text: '// tslint:disable-line',
          },
          line: 1,
          messageId: 'commentDetected',
        },
      ],
      output: 'someCode();',
    },
    {
      code: '// tslint:disable-next-line:rule1 rule2 rule3...',
      errors: [
        {
          column: 1,
          data: { text: '// tslint:disable-next-line:rule1 rule2 rule3...' },
          line: 1,
          messageId: 'commentDetected',
        },
      ],
      output: '',
    },
    {
      code: `
const woah = doSomeStuff();
// tslint:disable-line
console.log(woah);
      `,
      errors: [
        {
          column: 1,
          data: {
            text: '// tslint:disable-line',
          },
          line: 3,
          messageId: 'commentDetected',
        },
      ],
      output: `
const woah = doSomeStuff();
console.log(woah);
      `,
    },
  ],
});

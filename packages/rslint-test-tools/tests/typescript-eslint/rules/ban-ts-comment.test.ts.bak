import { RuleTester, getFixturesRootDir } from '../RuleTester.ts';

const rootPath = getFixturesRootDir();

const ruleTester = new RuleTester({
  languageOptions: {
    parserOptions: {
      project: './tsconfig.json',
      tsconfigRootDir: rootPath,
    },
  },
});

ruleTester.run('ban-ts-comment', {
  valid: [
    '// just a comment containing @ts-expect-error somewhere',
    {
      code: `
if (false) {
  // @ts-expect-error: Unreachable code error
  console.log('hello');
}
      `,
      options: [{ 'ts-expect-error': 'allow-with-description' }],
    },
    {
      code: `
if (false) {
  // @ts-expect-error: Unreachable code error
  console.log('hello');
}
      `,
      options: [{ 'ts-expect-error': true }],
    },
    {
      code: `
if (false) {
  /* @ts-expect-error: Unreachable code error */
  console.log('hello');
}
      `,
      options: [{ 'ts-expect-error': 'allow-with-description' }],
    },
    {
      code: `
if (false) {
  // @ts-expect-error
  console.log('hello');
}
      `,
      options: [{ 'ts-expect-error': true }],
    },
    {
      code: `
if (false) {
  /* @ts-expect-error */
  console.log('hello');
}
      `,
      options: [{ 'ts-expect-error': true }],
    },
    {
      code: `
if (false) {
  // @ts-ignore: Unreachable code error
  console.log('hello');
}
      `,
      options: [{ 'ts-ignore': 'allow-with-description' }],
    },
    {
      code: `
if (false) {
  // @ts-ignore
  console.log('hello');
}
      `,
      options: [{ 'ts-ignore': true }],
    },
    {
      code: `
if (false) {
  /* @ts-ignore */
  console.log('hello');
}
      `,
      options: [{ 'ts-ignore': true }],
    },
    {
      code: `
// @ts-nocheck: Do not check this file
if (false) {
  console.log('hello');
}
      `,
      options: [{ 'ts-nocheck': 'allow-with-description' }],
    },
    {
      code: `
// @ts-nocheck
if (false) {
  console.log('hello');
}
      `,
      options: [{ 'ts-nocheck': true }],
    },
    {
      code: `
/* @ts-nocheck */
if (false) {
  console.log('hello');
}
      `,
      options: [{ 'ts-nocheck': true }],
    },
    {
      code: `
// @ts-check: Check this file
if (false) {
  console.log('hello');
}
      `,
      options: [{ 'ts-check': 'allow-with-description' }],
    },
    {
      code: `
// @ts-check
if (false) {
  console.log('hello');
}
      `,
      options: [{ 'ts-check': true }],
    },
    {
      code: `
/* @ts-check */
if (false) {
  console.log('hello');
}
      `,
      options: [{ 'ts-check': true }],
    },
    {
      code: `
// @ts-expect-error: TODO: fix type issue
console.log('hello');
      `,
      options: [
        {
          'ts-expect-error': {
            descriptionFormat: '^TODO:',
          },
        },
      ],
    },
    {
      code: `
// @ts-expect-error: TODO: fix type issue
console.log('hello');
      `,
      options: [
        {
          'ts-expect-error': {
            descriptionFormat: '^TODO:',
            minimumDescriptionLength: 10,
          },
        },
      ],
    },
    {
      code: `
// @ts-expect-error: very long description that is more than 10 characters
console.log('hello');
      `,
      options: [
        {
          'ts-expect-error': {
            minimumDescriptionLength: 10,
          },
        },
      ],
    },
    {
      code: `
if (false) {
  // @ts-expect-error: Unreachable code error
  console.log('hello');
}
      `,
      options: [{ 'ts-expect-error': 'allow-with-description' }],
    },
    {
      code: `
if (false) {
  // @ts-ignore: Unreachable code error
  console.log('hello');
}
      `,
      options: [{ 'ts-ignore': 'allow-with-description' }],
    },
    {
      code: `
// @ts-nocheck: Do not check this file
if (false) {
  console.log('hello');
}
      `,
      options: [{ 'ts-nocheck': 'allow-with-description' }],
    },
    {
      code: `
// @ts-check: Check this file
if (false) {
  console.log('hello');
}
      `,
      options: [{ 'ts-check': 'allow-with-description' }],
    },
    {
      code: `
// just a comment containing @ts-expect-error somewhere in the middle
if (false) {
  console.log('hello');
}
      `,
    },
    {
      code: `
const a = {
  // @ts-expect-error: FIXME
  b: 1,
};
      `,
      options: [{ 'ts-expect-error': 'allow-with-description' }],
    },
    {
      code: `
const a = {
  /* @ts-expect-error: FIXME */
  b: 1,
};
      `,
      options: [{ 'ts-expect-error': 'allow-with-description' }],
    },
    {
      code: `
// This will trigger a lint error
// eslint-disable-next-line @typescript-eslint/ban-types
export type Example = Object;
      `,
    },
    {
      code: `
/* This will trigger a lint error
eslint-disable-next-line @typescript-eslint/ban-types */
export type Example = Object;
      `,
    },
    {
      code: `
/*
 * This will trigger a lint error
 * eslint-disable-next-line @typescript-eslint/ban-types
 */
export type Example = Object;
      `,
    },
    {
      code: `
/* eslint-disable
@typescript-eslint/ban-types
*/
export type Example = Object;
/* eslint-enable @typescript-eslint/ban-types */
      `,
    },
    {
      code: `
/*
eslint-disable @typescript-eslint/ban-types
*/
export type Example = Object;
/* eslint-enable @typescript-eslint/ban-types */
      `,
    },
    {
      code: `
/*
eslint-disable-next-line
@typescript-eslint/ban-types
*/
export type Example = Object;
      `,
    },
    {
      code: `
// @prettier-ignore
const a: string = 'a';
      `,
    },
    {
      code: `
// @abc-def-ghi
const a: string = 'a';
      `,
    },
    {
      code: `
// @tsx-expect-error
const a: string = 'a';
      `,
    },
  ],
  invalid: [
    {
      code: `
if (false) {
  // @ts-expect-error
  console.log('hello');
}
      `,
      options: [{ 'ts-expect-error': false }],
      errors: [
        {
          data: { directive: 'expect-error' },
          messageId: 'tsDirectiveComment',
          line: 3,
          column: 3,
        },
      ],
    },
    {
      code: `
if (false) {
  /* @ts-expect-error */
  console.log('hello');  
}
      `,
      options: [{ 'ts-expect-error': false }],
      errors: [
        {
          data: { directive: 'expect-error' },
          messageId: 'tsDirectiveComment',
          line: 3,
          column: 3,
        },
      ],
    },
    {
      code: `
if (false) {
  // @ts-expect-error
  console.log('hello');
}
      `,
      options: [{ 'ts-expect-error': 'allow-with-description' }],
      errors: [
        {
          data: { directive: 'expect-error', minimumDescriptionLength: 3 },
          messageId: 'tsDirectiveCommentRequiresDescription',
          line: 3,
          column: 3,
        },
      ],
    },
    {
      code: `
if (false) {
  /* @ts-expect-error */
  console.log('hello');
}
      `,
      options: [{ 'ts-expect-error': 'allow-with-description' }],
      errors: [
        {
          data: { directive: 'expect-error', minimumDescriptionLength: 3 },
          messageId: 'tsDirectiveCommentRequiresDescription',
          line: 3,
          column: 3,
        },
      ],
    },
    {
      code: `
if (false) {
  // @ts-ignore
  console.log('hello');
}
      `,
      options: [{ 'ts-ignore': false }],
      errors: [
        {
          data: { directive: 'ignore' },
          messageId: 'tsDirectiveComment',
          line: 3,
          column: 3,
        },
      ],
    },
    {
      code: `
if (false) {
  /* @ts-ignore */
  console.log('hello');
}
      `,
      options: [{ 'ts-ignore': false }],
      errors: [
        {
          data: { directive: 'ignore' },
          messageId: 'tsDirectiveComment',
          line: 3,
          column: 3,
        },
      ],
    },
    {
      code: `
if (false) {
  // @ts-ignore
  console.log('hello');
}
      `,
      options: [{ 'ts-ignore': 'allow-with-description' }],
      errors: [
        {
          data: { directive: 'ignore', minimumDescriptionLength: 3 },
          messageId: 'tsDirectiveCommentRequiresDescription',
          line: 3,
          column: 3,
        },
      ],
    },
    {
      code: `
if (false) {
  /* @ts-ignore */
  console.log('hello');
}
      `,
      options: [{ 'ts-ignore': 'allow-with-description' }],
      errors: [
        {
          data: { directive: 'ignore', minimumDescriptionLength: 3 },
          messageId: 'tsDirectiveCommentRequiresDescription',
          line: 3,
          column: 3,
        },
      ],
    },
    {
      code: `
// @ts-nocheck
if (false) {
  console.log('hello');
}
      `,
      options: [{ 'ts-nocheck': false }],
      errors: [
        {
          data: { directive: 'nocheck' },
          messageId: 'tsDirectiveComment',
          line: 2,
          column: 1,
        },
      ],
    },
    {
      code: `
/* @ts-nocheck */
if (false) {
  console.log('hello');
}
      `,
      options: [{ 'ts-nocheck': false }],
      errors: [
        {
          data: { directive: 'nocheck' },
          messageId: 'tsDirectiveComment',
          line: 2,
          column: 1,
        },
      ],
    },
    {
      code: `
// @ts-nocheck
if (false) {
  console.log('hello');
}
      `,
      options: [{ 'ts-nocheck': 'allow-with-description' }],
      errors: [
        {
          data: { directive: 'nocheck', minimumDescriptionLength: 3 },
          messageId: 'tsDirectiveCommentRequiresDescription',
          line: 2,
          column: 1,
        },
      ],
    },
    {
      code: `
/* @ts-nocheck */
if (false) {
  console.log('hello');
}
      `,
      options: [{ 'ts-nocheck': 'allow-with-description' }],
      errors: [
        {
          data: { directive: 'nocheck', minimumDescriptionLength: 3 },
          messageId: 'tsDirectiveCommentRequiresDescription',
          line: 2,
          column: 1,
        },
      ],
    },
    {
      code: `
// @ts-check
if (false) {
  console.log('hello');
}
      `,
      options: [{ 'ts-check': false }],
      errors: [
        {
          data: { directive: 'check' },
          messageId: 'tsDirectiveComment',
          line: 2,
          column: 1,
        },
      ],
    },
    {
      code: `
/* @ts-check */
if (false) {
  console.log('hello');
}
      `,
      options: [{ 'ts-check': false }],
      errors: [
        {
          data: { directive: 'check' },
          messageId: 'tsDirectiveComment',
          line: 2,
          column: 1,
        },
      ],
    },
    {
      code: `
// @ts-check
if (false) {
  console.log('hello');
}
      `,
      options: [{ 'ts-check': 'allow-with-description' }],
      errors: [
        {
          data: { directive: 'check', minimumDescriptionLength: 3 },
          messageId: 'tsDirectiveCommentRequiresDescription',
          line: 2,
          column: 1,
        },
      ],
    },
    {
      code: `
/* @ts-check */
if (false) {
  console.log('hello');
}
      `,
      options: [{ 'ts-check': 'allow-with-description' }],
      errors: [
        {
          data: { directive: 'check', minimumDescriptionLength: 3 },
          messageId: 'tsDirectiveCommentRequiresDescription',
          line: 2,
          column: 1,
        },
      ],
    },
    {
      code: `
// @ts-expect-error: TODO fix
console.log('hello');
      `,
      options: [
        {
          'ts-expect-error': {
            descriptionFormat: '^TODO:',
          },
        },
      ],
      errors: [
        {
          data: { directive: 'expect-error', format: '^TODO:' },
          messageId: 'tsDirectiveCommentDescriptionNotMatchFormat',
          line: 2,
          column: 1,
        },
      ],
    },
    {
      code: `
// @ts-expect-error: short
console.log('hello');
      `,
      options: [
        {
          'ts-expect-error': {
            minimumDescriptionLength: 10,
          },
        },
      ],
      errors: [
        {
          data: { directive: 'expect-error', minimumDescriptionLength: 10 },
          messageId: 'tsDirectiveCommentDescriptionTooShort',
          line: 2,
          column: 1,
        },
      ],
    },
    {
      code: `
if (false) {
  // @ts-ignore: Unreachable code error
  console.log('hello');
}
      `,
      errors: [
        {
          data: { directive: 'ignore' },
          messageId: 'tsIgnoreInsteadOfExpectError',
          line: 3,
          column: 3,
          suggestions: [
            {
              messageId: 'replaceTsIgnoreWithTsExpectError',
              output: `
if (false) {
  // @ts-expect-error: Unreachable code error
  console.log('hello');
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
if (false) {
  /* @ts-ignore: Unreachable code error */
  console.log('hello');
}
      `,
      errors: [
        {
          data: { directive: 'ignore' },
          messageId: 'tsIgnoreInsteadOfExpectError',
          line: 3,
          column: 3,
          suggestions: [
            {
              messageId: 'replaceTsIgnoreWithTsExpectError',
              output: `
if (false) {
  /* @ts-expect-error: Unreachable code error */
  console.log('hello');
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
if (false) {
  // @ts-ignore
  console.log('hello');
}
      `,
      errors: [
        {
          data: { directive: 'ignore' },
          messageId: 'tsIgnoreInsteadOfExpectError',
          line: 3,
          column: 3,
          suggestions: [
            {
              messageId: 'replaceTsIgnoreWithTsExpectError',
              output: `
if (false) {
  // @ts-expect-error
  console.log('hello');
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
if (false) {
  /* @ts-ignore */
  console.log('hello');
}
      `,
      errors: [
        {
          data: { directive: 'ignore' },
          messageId: 'tsIgnoreInsteadOfExpectError',
          line: 3,
          column: 3,
          suggestions: [
            {
              messageId: 'replaceTsIgnoreWithTsExpectError',
              output: `
if (false) {
  /* @ts-expect-error */
  console.log('hello');
}
      `,
            },
          ],
        },
      ],
    },
    {
      code: `
// @ts-nocheck: Do not check this file
if (false) {
  console.log('hello');
}
      `,
      errors: [
        {
          data: { directive: 'nocheck' },
          messageId: 'tsDirectiveComment',
          line: 2,
          column: 1,
        },
      ],
    },
    {
      code: `
/* @ts-nocheck: Do not check this file */
if (false) {
  console.log('hello');
}
      `,
      errors: [
        {
          data: { directive: 'nocheck' },
          messageId: 'tsDirectiveComment',
          line: 2,
          column: 1,
        },
      ],
    },
    {
      code: `
// @ts-nocheck
if (false) {
  console.log('hello');
}
      `,
      errors: [
        {
          data: { directive: 'nocheck' },
          messageId: 'tsDirectiveComment',
          line: 2,
          column: 1,
        },
      ],
    },
    {
      code: `
/* @ts-nocheck */
if (false) {
  console.log('hello');
}
      `,
      errors: [
        {
          data: { directive: 'nocheck' },
          messageId: 'tsDirectiveComment',
          line: 2,
          column: 1,
        },
      ],
    },
    {
      code: `
// @ts-check: Check this file
if (false) {
  console.log('hello');
}
      `,
      errors: [
        {
          data: { directive: 'check' },
          messageId: 'tsDirectiveComment',
          line: 2,
          column: 1,
        },
      ],
    },
    {
      code: `
/* @ts-check: Check this file */
if (false) {
  console.log('hello');
}
      `,
      errors: [
        {
          data: { directive: 'check' },
          messageId: 'tsDirectiveComment',
          line: 2,
          column: 1,
        },
      ],
    },
    {
      code: `
// @ts-check
if (false) {
  console.log('hello');
}
      `,
      errors: [
        {
          data: { directive: 'check' },
          messageId: 'tsDirectiveComment',
          line: 2,
          column: 1,
        },
      ],
    },
    {
      code: `
/* @ts-check */
if (false) {
  console.log('hello');
}
      `,
      errors: [
        {
          data: { directive: 'check' },
          messageId: 'tsDirectiveComment',
          line: 2,
          column: 1,
        },
      ],
    },
    {
      code: `
// @ts-expect-error
console.log('hello');
      `,
      errors: [
        {
          data: { directive: 'expect-error' },
          messageId: 'tsDirectiveComment',
          line: 2,
          column: 1,
        },
      ],
    },
    {
      code: `
/* @ts-expect-error */
console.log('hello');
      `,
      errors: [
        {
          data: { directive: 'expect-error' },
          messageId: 'tsDirectiveComment',
          line: 2,
          column: 1,
        },
      ],
    },
    {
      code: `
// @ts-expect-error: Unreachable code error
console.log('hello');
      `,
      errors: [
        {
          data: { directive: 'expect-error' },
          messageId: 'tsDirectiveComment',
          line: 2,
          column: 1,
        },
      ],
    },
    {
      code: `
/* @ts-expect-error: Unreachable code error */
console.log('hello');
      `,
      errors: [
        {
          data: { directive: 'expect-error' },
          messageId: 'tsDirectiveComment',
          line: 2,
          column: 1,
        },
      ],
    },
    {
      code: `
const a = {
  // @ts-expect-error
  b: 1,
};
      `,
      errors: [
        {
          data: { directive: 'expect-error' },
          messageId: 'tsDirectiveComment',
          line: 3,
          column: 3,
        },
      ],
    },
    {
      code: `
const a = {
  /* @ts-expect-error */
  b: 1,
};
      `,
      errors: [
        {
          data: { directive: 'expect-error' },
          messageId: 'tsDirectiveComment',
          line: 3,
          column: 3,
        },
      ],
    },
    {
      code: `
const a = {
  // @ts-expect-error: FIXME
  b: 1,
};
      `,
      errors: [
        {
          data: { directive: 'expect-error' },
          messageId: 'tsDirectiveComment',
          line: 3,
          column: 3,
        },
      ],
    },
    {
      code: `
const a = {
  /* @ts-expect-error: FIXME */
  b: 1,
};
      `,
      errors: [
        {
          data: { directive: 'expect-error' },
          messageId: 'tsDirectiveComment',
          line: 3,
          column: 3,
        },
      ],
    },
  ],
});

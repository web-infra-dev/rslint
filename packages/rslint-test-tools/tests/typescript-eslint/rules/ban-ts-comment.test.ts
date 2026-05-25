import { noFormat, RuleTester } from '@typescript-eslint/rule-tester';

const ruleTester = new RuleTester();

ruleTester.run(
  'ban-ts-comment',
  {
    valid: [
      // ========================
      // Edge cases: directive-like text in non-comment contexts
      // ========================
      'const c = "// @ts-ignore";',
      'const c = "/* @ts-expect-error */";',
      'const c = `// @ts-ignore`;',
      // Trailing comment with sufficient description (default: allow-with-description)
      'const x = 1; // @ts-expect-error: suppress this',

      // ========================
      // ts-expect-error: valid
      // ========================
      '// just a comment containing @ts-expect-error somewhere',
      `
/*
 @ts-expect-error running with long description in a block
*/
    `,
      `
/* @ts-expect-error not on the last line
 */
    `,
      `
/**
 * @ts-expect-error not on the last line
 */
    `,
      `
/* not on the last line
 * @ts-expect-error
 */
    `,
      `
/* @ts-expect-error
 * not on the last line */
    `,
      {
        code: '// @ts-expect-error',
        options: [{ 'ts-expect-error': false }],
      },
      {
        code: '// @ts-expect-error here is why the error is expected',
        options: [
          {
            'ts-expect-error': 'allow-with-description',
          },
        ],
      },
      {
        code: `
/*
 * @ts-expect-error here is why the error is expected */
      `,
        options: [
          {
            'ts-expect-error': 'allow-with-description',
          },
        ],
      },
      {
        code: '// @ts-expect-error exactly 21 characters',
        options: [
          {
            minimumDescriptionLength: 21,
            'ts-expect-error': 'allow-with-description',
          },
        ],
      },
      {
        code: `
/*
 * @ts-expect-error exactly 21 characters*/
      `,
        options: [
          {
            minimumDescriptionLength: 21,
            'ts-expect-error': 'allow-with-description',
          },
        ],
      },
      {
        code: '// @ts-expect-error: TS1234 because xyz',
        options: [
          {
            minimumDescriptionLength: 10,
            'ts-expect-error': {
              descriptionFormat: '^: TS\\d+ because .+$',
            },
          },
        ],
      },
      {
        code: `
/*
 * @ts-expect-error: TS1234 because xyz */
      `,
        options: [
          {
            minimumDescriptionLength: 10,
            'ts-expect-error': {
              descriptionFormat: '^: TS\\d+ because .+$',
            },
          },
        ],
      },
      {
        code: '// @ts-expect-error 👨‍👩‍👧‍👦👨‍👩‍👧‍👦👨‍👩‍👧‍👦',
        options: [
          {
            'ts-expect-error': 'allow-with-description',
          },
        ],
      },

      // ========================
      // ts-ignore: valid
      // ========================
      '// just a comment containing @ts-ignore somewhere',
      {
        code: '// @ts-ignore',
        options: [{ 'ts-ignore': false }],
      },
      {
        code: '// @ts-ignore I think that I am exempted from any need to follow the rules!',
        options: [{ 'ts-ignore': 'allow-with-description' }],
      },
      {
        code: `
/*
 @ts-ignore running with long description in a block
*/
      `,
        options: [
          {
            minimumDescriptionLength: 21,
            'ts-ignore': 'allow-with-description',
          },
        ],
      },
      `
/*
 @ts-ignore
*/
    `,
      `
/* @ts-ignore not on the last line
 */
    `,
      `
/**
 * @ts-ignore not on the last line
 */
    `,
      `
/* not on the last line
 * @ts-expect-error
 */
    `,
      `
/* @ts-ignore
 * not on the last line */
    `,
      {
        code: '// @ts-ignore: TS1234 because xyz',
        options: [
          {
            minimumDescriptionLength: 10,
            'ts-ignore': {
              descriptionFormat: '^: TS\\d+ because .+$',
            },
          },
        ],
      },
      {
        code: '// @ts-ignore 👨‍👩‍👧‍👦👨‍👩‍👧‍👦👨‍👩‍👧‍👦',
        options: [
          {
            'ts-ignore': 'allow-with-description',
          },
        ],
      },
      {
        code: `
/*
 * @ts-ignore here is why the error is expected */
      `,
        options: [
          {
            'ts-ignore': 'allow-with-description',
          },
        ],
      },
      {
        code: '// @ts-ignore exactly 21 characters',
        options: [
          {
            minimumDescriptionLength: 21,
            'ts-ignore': 'allow-with-description',
          },
        ],
      },
      {
        code: `
/*
 * @ts-ignore exactly 21 characters*/
      `,
        options: [
          {
            minimumDescriptionLength: 21,
            'ts-ignore': 'allow-with-description',
          },
        ],
      },
      {
        code: `
/*
 * @ts-ignore: TS1234 because xyz */
      `,
        options: [
          {
            minimumDescriptionLength: 10,
            'ts-ignore': {
              descriptionFormat: '^: TS\\d+ because .+$',
            },
          },
        ],
      },

      // ========================
      // ts-nocheck: valid
      // ========================
      '// just a comment containing @ts-nocheck somewhere',
      {
        code: '// @ts-nocheck',
        options: [{ 'ts-nocheck': false }],
      },
      {
        code: '// @ts-nocheck no doubt, people will put nonsense here from time to time just to get the rule to stop reporting, perhaps even long messages with other nonsense in them like other // @ts-nocheck or // @ts-ignore things',
        options: [{ 'ts-nocheck': 'allow-with-description' }],
      },
      {
        code: `
/*
 @ts-nocheck running with long description in a block
*/
      `,
        options: [
          {
            minimumDescriptionLength: 21,
            'ts-nocheck': 'allow-with-description',
          },
        ],
      },
      {
        code: '// @ts-nocheck: TS1234 because xyz',
        options: [
          {
            minimumDescriptionLength: 10,
            'ts-nocheck': {
              descriptionFormat: '^: TS\\d+ because .+$',
            },
          },
        ],
      },
      {
        code: '// @ts-nocheck 👨‍👩‍👧‍👦👨‍👩‍👧‍👦👨‍👩‍👧‍👦',
        options: [
          {
            'ts-nocheck': 'allow-with-description',
          },
        ],
      },
      '//// @ts-nocheck - pragma comments may contain 2 or 3 leading slashes',
      `
/**
 @ts-nocheck
*/
    `,
      `
/*
 @ts-nocheck
*/
    `,
      '/** @ts-nocheck */',
      '/* @ts-nocheck */',
      `
const a = 1;

// @ts-nocheck - should not be reported

// TS error is not actually suppressed
const b: string = a;
    `,

      // ========================
      // ts-check: valid
      // ========================
      '// just a comment containing @ts-check somewhere',
      `
/*
 @ts-check running with long description in a block
*/
    `,
      {
        code: '// @ts-check',
        options: [{ 'ts-check': false }],
      },
      {
        code: '// @ts-check with a description and also with a no-op // @ts-ignore',
        options: [
          {
            minimumDescriptionLength: 3,
            'ts-check': 'allow-with-description',
          },
        ],
      },
      {
        code: '// @ts-check: TS1234 because xyz',
        options: [
          {
            minimumDescriptionLength: 10,
            'ts-check': {
              descriptionFormat: '^: TS\\d+ because .+$',
            },
          },
        ],
      },
      {
        code: '// @ts-check 👨‍👩‍👧‍👦👨‍👩‍👧‍👦👨‍👩‍👧‍👦',
        options: [
          {
            'ts-check': 'allow-with-description',
          },
        ],
      },
      {
        code: '//// @ts-check - pragma comments may contain 2 or 3 leading slashes',
        options: [{ 'ts-check': true }],
      },
      {
        code: `
/**
 @ts-check
*/
      `,
        options: [{ 'ts-check': true }],
      },
      {
        code: `
/*
 @ts-check
*/
      `,
        options: [{ 'ts-check': true }],
      },
      {
        code: '/** @ts-check */',
        options: [{ 'ts-check': true }],
      },
      {
        code: '/* @ts-check */',
        options: [{ 'ts-check': true }],
      },

      // ========================
      // Default config full flow: step 3 → @ts-expect-error with desc → valid
      // ========================
      '// @ts-expect-error: some valid reason here',
    ],
    invalid: [
      // ========================
      // Default config full flow
      // ========================
      // Step 1: @ts-ignore with defaults → prefer
      // (covered in ts-ignore invalid section below)
      // Step 2: @ts-expect-error (no desc) with defaults → requires description
      {
        code: '// @ts-expect-error',
        errors: [{ messageId: 'tsDirectiveCommentRequiresDescription' }],
      },

      // ========================
      // ts-expect-error: invalid
      // ========================
      {
        code: '// @ts-expect-error',
        errors: [
          {
            data: { directive: 'expect-error' },
            messageId: 'tsDirectiveComment',
          },
        ],
        options: [{ 'ts-expect-error': true }],
      },
      {
        code: '/* @ts-expect-error */',
        errors: [
          {
            data: { directive: 'expect-error' },
            messageId: 'tsDirectiveComment',
          },
        ],
        options: [{ 'ts-expect-error': true }],
      },
      {
        code: `
/*
@ts-expect-error */
      `,
        errors: [
          {
            data: { directive: 'expect-error' },
            messageId: 'tsDirectiveComment',
          },
        ],
        options: [{ 'ts-expect-error': true }],
      },
      {
        code: `
/** on the last line
  @ts-expect-error */
      `,
        errors: [
          {
            data: { directive: 'expect-error' },
            messageId: 'tsDirectiveComment',
          },
        ],
        options: [{ 'ts-expect-error': true }],
      },
      {
        code: `
/** on the last line
 * @ts-expect-error */
      `,
        errors: [
          {
            data: { directive: 'expect-error' },
            messageId: 'tsDirectiveComment',
          },
        ],
        options: [{ 'ts-expect-error': true }],
      },
      {
        code: `
/**
 * @ts-expect-error: TODO */
      `,
        errors: [
          {
            data: { directive: 'expect-error', minimumDescriptionLength: 10 },
            messageId: 'tsDirectiveCommentRequiresDescription',
          },
        ],
        options: [
          {
            minimumDescriptionLength: 10,
            'ts-expect-error': 'allow-with-description',
          },
        ],
      },
      {
        code: `
/**
 * @ts-expect-error: TS1234 because xyz */
      `,
        errors: [
          {
            data: { directive: 'expect-error', minimumDescriptionLength: 25 },
            messageId: 'tsDirectiveCommentRequiresDescription',
          },
        ],
        options: [
          {
            minimumDescriptionLength: 25,
            'ts-expect-error': {
              descriptionFormat: '^: TS\\d+ because .+$',
            },
          },
        ],
      },
      {
        code: `
/**
 * @ts-expect-error: TS1234 */
      `,
        errors: [
          {
            data: { directive: 'expect-error', format: '^: TS\\d+ because .+$' },
            messageId: 'tsDirectiveCommentDescriptionNotMatchPattern',
          },
        ],
        options: [
          {
            'ts-expect-error': {
              descriptionFormat: '^: TS\\d+ because .+$',
            },
          },
        ],
      },
      {
        code: `
/**
 * @ts-expect-error    : TS1234 */
      `,
        errors: [
          {
            data: { directive: 'expect-error', format: '^: TS\\d+ because .+$' },
            messageId: 'tsDirectiveCommentDescriptionNotMatchPattern',
          },
        ],
        options: [
          {
            'ts-expect-error': {
              descriptionFormat: '^: TS\\d+ because .+$',
            },
          },
        ],
      },
      {
        code: `
/**
 * @ts-expect-error 👨‍👩‍👧‍👦 */
      `,
        errors: [
          {
            data: { directive: 'expect-error', minimumDescriptionLength: 3 },
            messageId: 'tsDirectiveCommentRequiresDescription',
          },
        ],
        options: [
          {
            'ts-expect-error': 'allow-with-description',
          },
        ],
      },
      {
        code: '/** @ts-expect-error */',
        errors: [
          {
            data: { directive: 'expect-error' },
            messageId: 'tsDirectiveComment',
          },
        ],
        options: [{ 'ts-expect-error': true }],
      },
      {
        code: '// @ts-expect-error: Suppress next line',
        errors: [
          {
            data: { directive: 'expect-error' },
            messageId: 'tsDirectiveComment',
          },
        ],
        options: [{ 'ts-expect-error': true }],
      },
      {
        code: '/////@ts-expect-error: Suppress next line',
        errors: [
          {
            data: { directive: 'expect-error' },
            messageId: 'tsDirectiveComment',
          },
        ],
        options: [{ 'ts-expect-error': true }],
      },
      {
        code: `
if (false) {
  // @ts-expect-error: Unreachable code error
  console.log('hello');
}
      `,
        errors: [
          {
            data: { directive: 'expect-error' },
            messageId: 'tsDirectiveComment',
          },
        ],
        options: [{ 'ts-expect-error': true }],
      },
      {
        code: '// @ts-expect-error',
        errors: [
          {
            data: { directive: 'expect-error', minimumDescriptionLength: 3 },
            messageId: 'tsDirectiveCommentRequiresDescription',
          },
        ],
        options: [
          {
            'ts-expect-error': 'allow-with-description',
          },
        ],
      },
      {
        code: '// @ts-expect-error: TODO',
        errors: [
          {
            data: { directive: 'expect-error', minimumDescriptionLength: 10 },
            messageId: 'tsDirectiveCommentRequiresDescription',
          },
        ],
        options: [
          {
            minimumDescriptionLength: 10,
            'ts-expect-error': 'allow-with-description',
          },
        ],
      },
      {
        code: '// @ts-expect-error: TS1234 because xyz',
        errors: [
          {
            data: { directive: 'expect-error', minimumDescriptionLength: 25 },
            messageId: 'tsDirectiveCommentRequiresDescription',
          },
        ],
        options: [
          {
            minimumDescriptionLength: 25,
            'ts-expect-error': {
              descriptionFormat: '^: TS\\d+ because .+$',
            },
          },
        ],
      },
      {
        code: '// @ts-expect-error: TS1234',
        errors: [
          {
            data: { directive: 'expect-error', format: '^: TS\\d+ because .+$' },
            messageId: 'tsDirectiveCommentDescriptionNotMatchPattern',
          },
        ],
        options: [
          {
            'ts-expect-error': {
              descriptionFormat: '^: TS\\d+ because .+$',
            },
          },
        ],
      },
      {
        code: '// @ts-expect-error    : TS1234 because xyz',
        errors: [
          {
            data: { directive: 'expect-error', format: '^: TS\\d+ because .+$' },
            messageId: 'tsDirectiveCommentDescriptionNotMatchPattern',
          },
        ],
        options: [
          {
            'ts-expect-error': {
              descriptionFormat: '^: TS\\d+ because .+$',
            },
          },
        ],
      },
      {
        code: '// @ts-expect-error 👨‍👩‍👧‍👦',
        errors: [
          {
            data: { directive: 'expect-error', minimumDescriptionLength: 3 },
            messageId: 'tsDirectiveCommentRequiresDescription',
          },
        ],
        options: [
          {
            'ts-expect-error': 'allow-with-description',
          },
        ],
      },

      // ========================
      // ts-ignore: invalid (with suggestion fix, not autofix)
      // ========================
      // When both ts-ignore and ts-expect-error are true (both banned),
      // do NOT suggest prefer — just report tsDirectiveComment.
      {
        code: '// @ts-ignore',
        errors: [{ messageId: 'tsDirectiveComment' }],
        options: [{ 'ts-expect-error': true, 'ts-ignore': true }],
      },
      // ts-expect-error: allow-with-description → prefer (makes sense)
      {
        code: '// @ts-ignore',
        errors: [{ messageId: 'tsIgnoreInsteadOfExpectError' }],
        options: [
          { 'ts-expect-error': 'allow-with-description', 'ts-ignore': true },
        ],
      },
      // ts-expect-error: descriptionFormat → prefer (expect-error is allowed with format)
      {
        code: '// @ts-ignore',
        errors: [{ messageId: 'tsIgnoreInsteadOfExpectError' }],
        options: [
          {
            'ts-expect-error': { descriptionFormat: '^: TS\\d+' },
            'ts-ignore': true,
          },
        ],
      },
      {
        code: '// @ts-ignore',
        errors: [{ messageId: 'tsIgnoreInsteadOfExpectError' }],
      },
      {
        code: '/* @ts-ignore */',
        errors: [{ messageId: 'tsIgnoreInsteadOfExpectError' }],
        options: [{ 'ts-ignore': true }],
      },
      {
        code: `
/*
 @ts-ignore */
      `,
        errors: [{ messageId: 'tsIgnoreInsteadOfExpectError' }],
        options: [{ 'ts-ignore': true }],
      },
      {
        code: `
/** on the last line
  @ts-ignore */
      `,
        errors: [{ messageId: 'tsIgnoreInsteadOfExpectError' }],
        options: [{ 'ts-ignore': true }],
      },
      {
        code: `
/** on the last line
 * @ts-ignore */
      `,
        errors: [{ messageId: 'tsIgnoreInsteadOfExpectError' }],
        options: [{ 'ts-ignore': true }],
      },
      {
        code: '/** @ts-ignore */',
        errors: [{ messageId: 'tsIgnoreInsteadOfExpectError' }],
        options: [{ 'ts-expect-error': false, 'ts-ignore': true }],
      },
      {
        code: `
/**
 * @ts-ignore: TODO */
      `,
        errors: [{ messageId: 'tsIgnoreInsteadOfExpectError' }],
        options: [
          {
            minimumDescriptionLength: 10,
            'ts-expect-error': 'allow-with-description',
          },
        ],
      },
      {
        code: `
/**
 * @ts-ignore: TS1234 because xyz */
      `,
        errors: [{ messageId: 'tsIgnoreInsteadOfExpectError' }],
        options: [
          {
            minimumDescriptionLength: 25,
            'ts-expect-error': {
              descriptionFormat: '^: TS\\d+ because .+$',
            },
          },
        ],
      },
      {
        code: '// @ts-ignore: Suppress next line',
        errors: [{ messageId: 'tsIgnoreInsteadOfExpectError' }],
      },
      {
        code: '/////@ts-ignore: Suppress next line',
        errors: [{ messageId: 'tsIgnoreInsteadOfExpectError' }],
      },
      {
        code: `
if (false) {
  // @ts-ignore: Unreachable code error
  console.log('hello');
}
      `,
        errors: [{ messageId: 'tsIgnoreInsteadOfExpectError' }],
      },
      {
        code: '// @ts-ignore',
        errors: [
          {
            data: { directive: 'ignore', minimumDescriptionLength: 3 },
            messageId: 'tsDirectiveCommentRequiresDescription',
          },
        ],
        options: [{ 'ts-ignore': 'allow-with-description' }],
      },
      {
        code: noFormat`// @ts-ignore         `,
        errors: [
          {
            data: { directive: 'ignore', minimumDescriptionLength: 3 },
            messageId: 'tsDirectiveCommentRequiresDescription',
          },
        ],
        options: [{ 'ts-ignore': 'allow-with-description' }],
      },
      {
        code: '// @ts-ignore    .',
        errors: [
          {
            data: { directive: 'ignore', minimumDescriptionLength: 3 },
            messageId: 'tsDirectiveCommentRequiresDescription',
          },
        ],
        options: [{ 'ts-ignore': 'allow-with-description' }],
      },
      {
        code: '// @ts-ignore: TS1234 because xyz',
        errors: [
          {
            data: { directive: 'ignore', minimumDescriptionLength: 25 },
            messageId: 'tsDirectiveCommentRequiresDescription',
          },
        ],
        options: [
          {
            minimumDescriptionLength: 25,
            'ts-ignore': {
              descriptionFormat: '^: TS\\d+ because .+$',
            },
          },
        ],
      },
      {
        code: '// @ts-ignore: TS1234',
        errors: [
          {
            data: { directive: 'ignore', format: '^: TS\\d+ because .+$' },
            messageId: 'tsDirectiveCommentDescriptionNotMatchPattern',
          },
        ],
        options: [
          {
            'ts-ignore': {
              descriptionFormat: '^: TS\\d+ because .+$',
            },
          },
        ],
      },
      {
        code: '// @ts-ignore    : TS1234 because xyz',
        errors: [
          {
            data: { directive: 'ignore', format: '^: TS\\d+ because .+$' },
            messageId: 'tsDirectiveCommentDescriptionNotMatchPattern',
          },
        ],
        options: [
          {
            'ts-ignore': {
              descriptionFormat: '^: TS\\d+ because .+$',
            },
          },
        ],
      },
      {
        code: '// @ts-ignore 👨‍👩‍👧‍👦',
        errors: [
          {
            data: { directive: 'ignore', minimumDescriptionLength: 3 },
            messageId: 'tsDirectiveCommentRequiresDescription',
          },
        ],
        options: [
          {
            'ts-ignore': 'allow-with-description',
          },
        ],
      },

      // ========================
      // ts-nocheck: invalid
      // ========================
      {
        code: '// @ts-nocheck',
        errors: [
          {
            data: { directive: 'nocheck' },
            messageId: 'tsDirectiveComment',
          },
        ],
        options: [{ 'ts-nocheck': true }],
      },
      {
        code: '// @ts-nocheck',
        errors: [
          {
            data: { directive: 'nocheck' },
            messageId: 'tsDirectiveComment',
          },
        ],
      },
      {
        code: '// @ts-nocheck: Suppress next line',
        errors: [
          {
            data: { directive: 'nocheck' },
            messageId: 'tsDirectiveComment',
          },
        ],
      },
      {
        code: '// @ts-nocheck',
        errors: [
          {
            data: { directive: 'nocheck', minimumDescriptionLength: 3 },
            messageId: 'tsDirectiveCommentRequiresDescription',
          },
        ],
        options: [{ 'ts-nocheck': 'allow-with-description' }],
      },
      {
        code: '// @ts-nocheck: TS1234 because xyz',
        errors: [
          {
            data: { directive: 'nocheck', minimumDescriptionLength: 25 },
            messageId: 'tsDirectiveCommentRequiresDescription',
          },
        ],
        options: [
          {
            minimumDescriptionLength: 25,
            'ts-nocheck': {
              descriptionFormat: '^: TS\\d+ because .+$',
            },
          },
        ],
      },
      {
        code: '// @ts-nocheck: TS1234',
        errors: [
          {
            data: { directive: 'nocheck', format: '^: TS\\d+ because .+$' },
            messageId: 'tsDirectiveCommentDescriptionNotMatchPattern',
          },
        ],
        options: [
          {
            'ts-nocheck': {
              descriptionFormat: '^: TS\\d+ because .+$',
            },
          },
        ],
      },
      {
        code: '// @ts-nocheck    : TS1234 because xyz',
        errors: [
          {
            data: { directive: 'nocheck', format: '^: TS\\d+ because .+$' },
            messageId: 'tsDirectiveCommentDescriptionNotMatchPattern',
          },
        ],
        options: [
          {
            'ts-nocheck': {
              descriptionFormat: '^: TS\\d+ because .+$',
            },
          },
        ],
      },
      {
        code: '// @ts-nocheck 👨‍👩‍👧‍👦',
        errors: [
          {
            data: { directive: 'nocheck', minimumDescriptionLength: 3 },
            messageId: 'tsDirectiveCommentRequiresDescription',
          },
        ],
        options: [
          {
            'ts-nocheck': 'allow-with-description',
          },
        ],
      },
      {
        // comment's column > first statement's column
        // eslint-disable-next-line @typescript-eslint/internal/plugin-test-formatting
        code: `
 // @ts-nocheck
const a: true = false;
      `,
        errors: [
          {
            data: { directive: 'nocheck', minimumDescriptionLength: 3 },
            messageId: 'tsDirectiveComment',
          },
        ],
      },

      // ========================
      // ts-check: invalid
      // ========================
      {
        code: '// @ts-check',
        errors: [
          {
            data: { directive: 'check' },
            messageId: 'tsDirectiveComment',
          },
        ],
        options: [{ 'ts-check': true }],
      },
      {
        code: '// @ts-check: Suppress next line',
        errors: [
          {
            data: { directive: 'check' },
            messageId: 'tsDirectiveComment',
          },
        ],
        options: [{ 'ts-check': true }],
      },
      {
        code: `
if (false) {
  // @ts-check: Unreachable code error
  console.log('hello');
}
      `,
        errors: [
          {
            data: { directive: 'check' },
            messageId: 'tsDirectiveComment',
          },
        ],
        options: [{ 'ts-check': true }],
      },
      {
        code: '// @ts-check',
        errors: [
          {
            data: { directive: 'check', minimumDescriptionLength: 3 },
            messageId: 'tsDirectiveCommentRequiresDescription',
          },
        ],
        options: [{ 'ts-check': 'allow-with-description' }],
      },
      {
        code: '// @ts-check: TS1234 because xyz',
        errors: [
          {
            data: { directive: 'check', minimumDescriptionLength: 25 },
            messageId: 'tsDirectiveCommentRequiresDescription',
          },
        ],
        options: [
          {
            minimumDescriptionLength: 25,
            'ts-check': {
              descriptionFormat: '^: TS\\d+ because .+$',
            },
          },
        ],
      },
      {
        code: '// @ts-check: TS1234',
        errors: [
          {
            data: { directive: 'check', format: '^: TS\\d+ because .+$' },
            messageId: 'tsDirectiveCommentDescriptionNotMatchPattern',
          },
        ],
        options: [
          {
            'ts-check': {
              descriptionFormat: '^: TS\\d+ because .+$',
            },
          },
        ],
      },
      {
        code: '// @ts-check    : TS1234 because xyz',
        errors: [
          {
            data: { directive: 'check', format: '^: TS\\d+ because .+$' },
            messageId: 'tsDirectiveCommentDescriptionNotMatchPattern',
          },
        ],
        options: [
          {
            'ts-check': {
              descriptionFormat: '^: TS\\d+ because .+$',
            },
          },
        ],
      },
      {
        code: '// @ts-check 👨‍👩‍👧‍👦',
        errors: [
          {
            data: { directive: 'check', minimumDescriptionLength: 3 },
            messageId: 'tsDirectiveCommentRequiresDescription',
          },
        ],
        options: [
          {
            'ts-check': 'allow-with-description',
          },
        ],
      },
    ],
  },
  { description: 'ban-ts-comment' },
);

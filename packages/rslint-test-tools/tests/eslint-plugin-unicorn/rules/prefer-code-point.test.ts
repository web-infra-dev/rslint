// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation. See LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-code-point.js
import { RuleTester } from '../rule-tester';

const charCodeAt = { messageId: 'error/charCodeAt' };
const fromCharCode = { messageId: 'error/fromCharCode' };

const js = (code: string) => ({ code, filename: 'file.mjs' });

const invalid = (
  code: string,
  message: { messageId: string },
  suggestion: boolean,
  filename = 'file.mjs',
) => ({
  code,
  filename,
  errors: [{ ...message, suggestions: suggestion ? [{}] : [] }],
});

new RuleTester().run('prefer-code-point', {} as never, {
  valid: [
    '"🦄".codePointAt(0)',
    'foo.charCodeAt',
    'new foo.charCodeAt',
    'charCodeAt(0)',
    'foo.charCodeAt?.(0)',
    'foo[charCodeAt](0)',
    'foo["charCodeAt"](0)',
    'foo.notCharCodeAt(0)',

    'String.fromCodePoint(0x1f984)',
    'String.fromCodePoint',
    'new String.fromCodePoint',
    'fromCodePoint(foo)',
    'String.fromCodePoint?.(foo)',
    'String?.fromCodePoint(foo)',
    'window.String.fromCodePoint(foo)',
    'String[fromCodePoint](foo)',
    'String["fromCodePoint"](foo)',
    'String.notFromCodePoint(foo)',
    'NotString.fromCodePoint(foo)',
  ].map((code) => js(code)),
  invalid: [
    // `charCodeAt` with a suggestion.
    invalid('string.charCodeAt(index)', charCodeAt, true),
    invalid('string?.charCodeAt(index)', charCodeAt, true),
    invalid('(( (( string )).charCodeAt( ((index)), )))', charCodeAt, true),
    invalid('string.charCodeAt(index).toString(16)', charCodeAt, true),
    invalid('string.charCodeAt(index) > 127', charCodeAt, true),
    invalid('foo(string.charCodeAt(index))', charCodeAt, true),
    invalid('const x = string.charCodeAt(index)', charCodeAt, true),
    invalid('string.charCodeAt(index) === 65', charCodeAt, true),
    invalid('string.charCodeAt(index) || 0', charCodeAt, true),
    invalid('!string.charCodeAt(index)', charCodeAt, true),

    // Numeric use: reported without a suggestion.
    invalid(
      'hash = (((hash * 31) | 0) + string.charCodeAt(offset)) | 0',
      charCodeAt,
      false,
    ),
    invalid('string.charCodeAt(index) + 1', charCodeAt, false),
    invalid('string.charCodeAt(index) * 31', charCodeAt, false),
    invalid('string.charCodeAt(index) - 48', charCodeAt, false),
    invalid('(string.charCodeAt(index) | 0)', charCodeAt, false),
    invalid('string.charCodeAt(index) ^ hash', charCodeAt, false),
    invalid('(hash << 5) - hash + string.charCodeAt(index)', charCodeAt, false),
    invalid('sum += string.charCodeAt(index)', charCodeAt, false),
    invalid(
      'sum += condition ? string.charCodeAt(index) : 0',
      charCodeAt,
      false,
    ),
    invalid('sum += string.charCodeAt(index) || 0', charCodeAt, false),
    invalid(
      'sum += condition ? 0 : string.charCodeAt(index)',
      charCodeAt,
      false,
    ),
    invalid('string?.charCodeAt(index) + 1', charCodeAt, false),
    invalid('~string.charCodeAt(index)', charCodeAt, false),
    invalid('-string.charCodeAt(index)', charCodeAt, false),
    // The call is the condition, not a branch, so a suggestion remains.
    invalid('sum += string.charCodeAt(index) ? 1 : 0', charCodeAt, true),

    // `String.fromCharCode`.
    invalid('String.fromCharCode( code )', fromCharCode, true),
    invalid('(( (( String )).fromCharCode( ((code)), ) ))', fromCharCode, true),
    invalid('String.fromCharCode.bind(String)', fromCharCode, true),
    invalid('const x = String.fromCharCode', fromCharCode, true),
    invalid('String.fromCharCode(65, 66, 67)', fromCharCode, true),

    // TypeScript wrappers around the receiver.
    invalid('foo!.charCodeAt(0)', charCodeAt, true, 'file.ts'),
    invalid('(str as string).charCodeAt(0)', charCodeAt, true, 'file.ts'),

    // TypeScript wrappers around the result keep numeric context.
    invalid('text.charCodeAt(0)! ^ 0;', charCodeAt, false, 'file.ts'),
    invalid(
      '(text.charCodeAt(0) as number) ^ 0;',
      charCodeAt,
      false,
      'file.ts',
    ),
    invalid(
      '(text.charCodeAt(0) satisfies number) ^ 0;',
      charCodeAt,
      false,
      'file.ts',
    ),
    invalid('(<number>text.charCodeAt(0)) ^ 0;', charCodeAt, false, 'file.ts'),
    invalid(
      'sum += (text?.charCodeAt(0)! || 0) as number;',
      charCodeAt,
      false,
      'file.ts',
    ),
    invalid(
      'sum += ready ? (text.charCodeAt(0) as number) : 0;',
      charCodeAt,
      false,
      'file.ts',
    ),
    invalid('-(text.charCodeAt(0) as number);', charCodeAt, false, 'file.ts'),
    invalid('const value = text.charCodeAt(0)!;', charCodeAt, true, 'file.ts'),
  ],
});

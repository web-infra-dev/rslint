// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation. See LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-code-point.js
import assert from 'node:assert/strict';
import path from 'node:path';
import { runInNewContext } from 'node:vm';
import { transpileModule, ScriptTarget } from 'typescript';
import { lint } from '@rslint/core/internal';

import { RuleTester } from '../rule-tester';

const charCodeAt = { messageId: 'error/charCodeAt' } as const;
const fromCharCode = { messageId: 'error/fromCharCode' } as const;

const js = (code: string) => ({ code, filename: 'file.mjs' });

const methodSuggestions = {
  'error/charCodeAt': {
    method: 'charCodeAt',
    suggestionId: 'suggestion/codePointAt',
    replacement: 'codePointAt',
  },
  'error/fromCharCode': {
    method: 'fromCharCode',
    suggestionId: 'suggestion/fromCodePoint',
    replacement: 'fromCodePoint',
  },
} as const;

const invalid = (
  code: string,
  message: { messageId: keyof typeof methodSuggestions },
  suggestion: boolean,
  filename = 'file.mjs',
) => {
  const { method, suggestionId, replacement } =
    methodSuggestions[message.messageId];
  return {
    code,
    filename,
    errors: [
      {
        ...message,
        suggestions: suggestion
          ? [
              {
                messageId: suggestionId,
                output: code.replace(method, replacement),
              },
            ]
          : [],
      },
    ],
  };
};

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

// Upstream's final node:test executes the transpiled output to prove that a
// suggestion in this numeric context would change the string hash. The rename
// suggestion is withheld, so the value must remain the first UTF-16 code unit.
test('type wrappers do not enable suggestions that change string hashes', async () => {
  const code = 'const text = "𝌆"; text.charCodeAt(0)! ^ 0;';
  const filename = path.resolve(import.meta.dirname, 'src/virtual.ts');
  const result = await lint({
    config: [
      {
        plugins: ['unicorn'],
        languageOptions: {
          parserOptions: { project: false, projectService: false },
        },
        rules: { 'unicorn/prefer-code-point': 'error' },
      },
    ],
    configDirectory: import.meta.dirname,
    workingDirectory: process.cwd(),
    fileContents: { [filename]: code },
  });

  assert.equal(result.diagnostics.length, 1);
  assert.equal(result.diagnostics[0]!.messageId, 'error/charCodeAt');
  assert.equal(result.diagnostics[0]!.suggestions, undefined);
  assert.equal(result.output, undefined);

  const { outputText, diagnostics } = transpileModule(code, {
    compilerOptions: { target: ScriptTarget.ESNext },
    reportDiagnostics: true,
  });
  assert.deepEqual(diagnostics, []);
  assert.equal(runInNewContext(outputText), 55_348);
});

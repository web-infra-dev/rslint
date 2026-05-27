/**
 * Tokenizer edge-case regression tests.
 *
 * Each test pins a single behaviour that real-world source flowing
 * through `lintFile` doesn't exercise consistently:
 *
 *   - regex-vs-division for `++` / `--` postfix punctuators.
 *   - `Comment.value` of an unterminated `/* ... ` block must not
 *     drop the real last two characters.
 */
import { describe, test, expect } from '@rstest/core';

import { tokenize } from '../../src/lexer/tokenizer.js';
import { buildLineStartOffsets } from '../../src/ast/normalize-ast.js';

function lex(src: string) {
  return tokenize(src, buildLineStartOffsets(src));
}

describe('tokenizer: division vs regex literal after postfix `++` / `--`', () => {
  test('`i++ / 2` — `/` is division, not regex', () => {
    // Pre-fix: `couldStartRegex` returned true for any Punctuator not
    // in `['}', ')', ']']`, so `++` was treated as expression-prefix
    // and `/` started a regex scan that consumed ` 2;` into a single
    // RegularExpression token, shifting every downstream token.
    const { tokens } = lex(`const x = i++ / 2;`);
    const kinds = tokens.map((t) => `${t.type}:${t.value}`);
    expect(kinds).toEqual([
      'Keyword:const',
      'Identifier:x',
      'Punctuator:=',
      'Identifier:i',
      'Punctuator:++',
      'Punctuator:/',
      'Numeric:2',
      'Punctuator:;',
    ]);
  });

  test('`j-- / k` — `/` is division after `--`', () => {
    const { tokens } = lex(`j-- / k`);
    const kinds = tokens.map((t) => `${t.type}:${t.value}`);
    expect(kinds).toEqual([
      'Identifier:j',
      'Punctuator:--',
      'Punctuator:/',
      'Identifier:k',
    ]);
  });

  test('`return /re/g` — `/` after a keyword IS regex (regression guard)', () => {
    // The fix only special-cases `++` / `--`; expression-prefix
    // keywords like `return` must STILL produce a RegularExpression.
    const { tokens } = lex(`return /re/g`);
    const regex = tokens.find((t) => t.type === 'RegularExpression');
    expect(regex).toBeDefined();
    expect(regex?.value).toBe('/re/g');
  });

  test('`({a:1}) / 2` — object-literal `}` keeps `/` as division', () => {
    // `}` closing an object literal in expression position ends an
    // expression, so the following `/` is division. The tokenizer
    // tracks per-brace context (block vs object literal) so this case
    // is distinguished from a statement-block `}` — see
    // `regression-coverage.test.ts` for the block-`}`-followed-by-
    // regex counterpart.
    const { tokens } = lex(`const k = ({a:1}) / 2;`);
    expect(tokens.find((t) => t.type === 'RegularExpression')).toBeUndefined();
    expect(tokens.some((t) => t.type === 'Punctuator' && t.value === '/')).toBe(
      true,
    );
  });
});

describe('tokenizer: unterminated block comment value', () => {
  test('`/* unterminated TODO` keeps real last 2 characters', () => {
    // Pre-fix: `makeComment` for Block always did `text.slice(start+2, end-2)`,
    // dropping `DO` from `TODO`. Now the call site passes terminated=false
    // and makeComment slices `text.slice(start+2, end)` — full payload.
    const src = `/* unterminated TODO`;
    const { comments } = lex(src);
    expect(comments).toHaveLength(1);
    expect(comments[0].type).toBe('Block');
    expect(comments[0].value).toBe(' unterminated TODO');
    expect(comments[0].range).toEqual([0, src.length]);
  });

  test('degenerate `/*` alone — value is empty string, not negative slice', () => {
    // Pre-fix: `text.slice(0+2, 2-2) = text.slice(2, 0) = ''` — coincidentally
    // produced empty string but only because the slice's reversed bounds
    // round to ''. For `/**` (3 chars, end-2=1), the slice was `text.slice(2, 1) = ''`
    // — still wrong intent. Make sure post-fix the empty case is principled.
    {
      const { comments } = lex(`/*`);
      expect(comments).toHaveLength(1);
      expect(comments[0].value).toBe('');
    }
    {
      const { comments } = lex(`/**`);
      expect(comments).toHaveLength(1);
      expect(comments[0].value).toBe('*');
    }
  });

  test('terminated `/* foo */` still strips both delimiters (no regression)', () => {
    const { comments } = lex(`/* foo */`);
    expect(comments).toHaveLength(1);
    expect(comments[0].value).toBe(' foo ');
    expect(comments[0].range).toEqual([0, 9]);
  });

  test('terminated block followed by code keeps both pieces correct', () => {
    const src = `/* foo */ const x = 1;`;
    const { tokens, comments } = lex(src);
    expect(comments).toHaveLength(1);
    expect(comments[0].value).toBe(' foo ');
    // tokens after the comment should still be `const`, `x`, `=`, `1`, `;`
    expect(tokens.map((t) => t.value)).toEqual(['const', 'x', '=', '1', ';']);
  });
});

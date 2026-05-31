/**
 * Differential token test (DESIGN §10): the native oxc token stream, rebuilt by
 * `token-builder` (the real path used by SourceCode), must match espree's `parse({tokens:true})`
 * tokens — the ESLint token contract rules consume — for `.js`/`.jsx`. This replaces the deleted
 * hand-written-tokenizer unit tests with a live oracle, so the contract is pinned against
 * espree itself rather than a frozen expectation. A few `.ts` cases use a hand-pinned
 * expectation (espree can't parse TS; the values were verified against @typescript-eslint
 * /typescript-estree during the M3 spike).
 */
import { describe, test, expect } from '@rstest/core';
import * as espree from 'espree';
import { parse as nativeParse } from '@rslint/native';

import { buildTokens } from '../../../src/eslint-plugin/source-code/token-builder.js';
import { buildLineStartOffsets } from '../../../src/eslint-plugin/ast/normalize-ast.js';

type Lang = 'js' | 'cjs' | 'jsx';

// type + value + UTF-16 range (multi-byte/astral offsets pinned end-to-end),
// plus the `regex` {pattern,flags} for RegularExpression tokens so a
// pattern/flags split bug is caught — not just the raw value.
function fmt(t: {
  type: string;
  value: string;
  range: readonly [number, number];
  regex?: { pattern: string; flags: string } | null;
}): string {
  const base = `${t.type}:${t.value}@${t.range[0]}-${t.range[1]}`;
  return t.regex ? `${base}|re(${t.regex.pattern}/${t.regex.flags})` : base;
}

function oxcTokens(src: string, lang: Lang): string[] {
  const jsx = lang === 'jsx';
  const st = lang === 'cjs' ? 'script' : 'module';
  const p = nativeParse(jsx ? 't.jsx' : 't.js', src, st, jsx);
  return buildTokens(
    p.tokenTypes,
    p.tokenStarts,
    p.tokenEnds,
    src,
    buildLineStartOffsets(src),
  ).map(fmt);
}

function espreeTokens(src: string, lang: Lang): string[] {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const ast = espree.parse(src, {
    ecmaVersion: 'latest',
    sourceType: lang === 'cjs' ? 'script' : 'module',
    ecmaFeatures: { jsx: lang === 'jsx' },
    tokens: true,
    range: true,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  } as any) as any;
  return ast.tokens.map(fmt);
}

// [label, source, lang] — covers what the deleted lexer unit tests exercised.
const JS_CASES: ReadonlyArray<readonly [string, string, Lang]> = [
  // regex vs division (parser-driven disambiguation)
  ['regex after ctrl-paren', 'if (x) /re/.test(y)', 'js'],
  ['division after call', 'const a = f(x) / 2', 'js'],
  ['regex after block close', '{}\n/re/g', 'js'],
  ['division after object literal', 'const k = ({ a: 1 }) / 2', 'js'],
  ['division after postfix ++', 'const x = i++ / 2', 'js'],
  ['division after postfix --', 'const x = i-- / 2', 'js'],
  ['regex after prefix ++', '++/re/.lastIndex', 'js'],
  ['division after class-expression body', 'const c = (class {} / 2)', 'cjs'],
  ['yield* delegate is Punctuator', 'function* g(){ yield* other() }', 'js'],
  ['regex after fn-decl body', 'function f(){}\n/re/', 'js'],
  ['division after fn-expr body', 'const z = (function(){} / 2)', 'js'],
  ['regex after return', 'function w(){ return /re/g }', 'js'],
  ['division after this', 'const r = this / 2', 'js'],
  ['regex after generator yield', 'function* g(){ yield /re/ }', 'js'],
  [
    'await regex in async',
    'async function f(){ return await /re/g.test(x) }',
    'js',
  ],
  ['division after identifier from/of/as', 'const a = from / of / as', 'cjs'],
  // template
  ['template head/middle/tail', 'const t = `a${b + 1}c${d}e`', 'js'],
  ['nested template', 'const t = `x${ `y${z}` }w`', 'js'],
  ['no-substitution template', 'const t = `plain`', 'js'],
  ['regex inside template', 'const t = `${/re/g.test(x)}`', 'js'],
  ['tagged template', 'const t = tag`x${y}z`', 'js'],
  // numeric
  [
    'numeric variants',
    'const a = [0xFF, 1_000, 1e1_0, 123n, 0.5, 0b1010, 0o17, .5e3]',
    'js',
  ],
  // keyword classification (.js sloppy)
  [
    'reserved as property key',
    'const o = { enum: 1, class: 2, if: 3, static: 4 }',
    'cjs',
  ],
  [
    'reserved as member',
    'o.enum; o.class; o.private; o.static; o.async; o.await; o.of',
    'cjs',
  ],
  ['contextual binding', 'var async = 1, of = 2, as = 3, type = 4', 'cjs'],
  [
    'let/static/yield keyword',
    'let a = 1; class C { static x = 1 } function* g(){ yield 1 }',
    'js',
  ],
  // boolean / null
  ['boolean and null', 'const a = true, b = false, c = null', 'js'],
  // private identifier
  ['private identifier', 'class C { #x = 1; m(){ return this.#x } }', 'js'],
  // string / escapes
  ['string escapes raw', String.raw`const s = 'a\n\t', d = "x\"y"`, 'js'],
  ['identifier unicode escape', String.raw`const abc = 1; obj.\u{62}ar`, 'js'],
  // optional chaining vs ternary-with-decimal
  ['optional chain `?.`', 'const a = obj?.prop', 'js'],
  ['`?.` before digit is `?` + `.5`', 'const a = cond ? .4 : .2', 'js'],
  // meta property
  ['new.target', 'function f(){ return new.target }', 'js'],
  ['import.meta', 'const u = import.meta.url', 'js'],
  // shebang + comments (tokens exclude both)
  [
    'shebang + comments',
    '#!/usr/bin/env node\n// line\nconst x = 1 /* block */',
    'cjs',
  ],
  // JSX
  [
    'jsx element + attr + expr',
    'const e = <div className="x" data-y={z}>text {expr} more</div>',
    'jsx',
  ],
  [
    'jsx fragment + member + self-close',
    'const e = <><Foo.Bar baz /></>',
    'jsx',
  ],
  ['jsx namespaced name', 'const e = <a:b c:d="e" />', 'jsx'],
  ['jsx nested children', 'const e = <ul><li>a</li><li>b</li></ul>', 'jsx'],
  ['jsx entity text', 'const e = <p>a &amp; b &#169;</p>', 'jsx'],
  [
    'jsx attr string with entity (raw, not decoded)',
    'const e = <a b="&amp;c" />',
    'jsx',
  ],
  // regex pattern/flags — `fmt` now diffs the `regex` {pattern,flags} field
  ['regex multi-flags', 'const r = /a+b/gimsuy', 'js'],
  ['regex char class with slash', 'const r = /[a-z/]+/g', 'js'],
  ['regex escaped delimiter', String.raw`const r = /\/\d+/`, 'js'],
  ['regex empty via group', 'const r = /(?:)/', 'js'],
  // empty / trivial input
  ['empty file', '', 'js'],
  ['whitespace + comment only', '  // c\n', 'cjs'],
  // unicode / offset
  ['cjk identifier', 'const 中文变量 = "值"', 'js'],
  ['astral identifier', 'const \u{1D49C} = 1', 'js'],
  ['accented identifier', "const café = 'résumé'", 'js'],
];

describe('token differential: oxc token stream == espree.parse tokens (.js/.jsx)', () => {
  for (const [label, src, lang] of JS_CASES) {
    test(label, () => {
      expect(oxcTokens(src, lang)).toEqual(espreeTokens(src, lang));
    });
  }
});

// .ts: espree can't parse TypeScript. Hand-pinned `type:value` expectations, verified against
// @typescript-eslint/typescript-estree during the M3 spike (keyword split + this + as-const).
describe('token differential: oxc token types for TypeScript (.ts)', () => {
  function tsTokens(src: string): string[] {
    const p = nativeParse('t.ts', src, 'module', false);
    return buildTokens(
      p.tokenTypes,
      p.tokenStarts,
      p.tokenEnds,
      src,
      buildLineStartOffsets(src),
    ).map((t) => `${t.type}:${t.value}`);
  }

  test('enum / interface are Keyword in .ts', () => {
    const t = tsTokens('enum E {} interface I {}');
    expect(t).toContain('Keyword:enum');
    expect(t).toContain('Keyword:interface');
  });

  test('type / namespace / as / keyof are Identifier-type tokens in .ts', () => {
    expect(tsTokens('type T = X')).toContain('Identifier:type');
    expect(tsTokens('const x = y as Z')).toContain('Identifier:as');
    expect(tsTokens('type K = keyof typeof obj')).toContain('Identifier:keyof');
  });

  test('primitive type keywords (number/string) are Identifier in .ts', () => {
    expect(tsTokens('let x: number = 1')).toContain('Identifier:number');
    expect(tsTokens('let s: string = ""')).toContain('Identifier:string');
  });

  test('`as const`: declaration `const` is Keyword, type-position `const` is Identifier', () => {
    expect(tsTokens('const x = [1] as const')).toEqual(
      expect.arrayContaining(['Keyword:const', 'Identifier:const']),
    );
  });

  test('`this` parameter is Identifier; value `this` is Keyword', () => {
    expect(tsTokens('function f(this: Window) {}')).toContain(
      'Identifier:this',
    );
    expect(tsTokens('class C { m(){ return this.x } }')).toContain(
      'Keyword:this',
    );
  });

  test('class accessibility modifiers stay Keyword in .ts', () => {
    const t = tsTokens('class C { public x = 1; private y = 2 }');
    expect(t).toContain('Keyword:public');
    expect(t).toContain('Keyword:private');
  });
});

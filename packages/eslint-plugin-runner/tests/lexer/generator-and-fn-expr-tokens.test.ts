/**
 * Differential tests for two regex/division disambiguation fixes,
 * espree@11.2.0 as the oracle (same source through the runner's
 * `tokenize` and `espree.tokenize`; assert identical type+value+range).
 * Matching espree's TOKEN stream byte-for-byte is the contract — that
 * is what ESLint v10 rules consume — so these tests pin the runner to
 * espree's behavior, including espree's own limitations.
 *
 *   #2 — `function* g(){ yield /re/ }` makes `yield` a regex-operand
 *        position; sloppy non-generator `yield / 2` is identifier
 *        division. A context-free lexer used to pick division
 *        unconditionally, mis-tokenizing every generator `yield <regex>`
 *        (swallowing the operand into a phantom RegularExpression). The
 *        lexer now tracks an enclosing-generator stack, seeded only by
 *        `function*` / `async function*` — exactly the forms espree's
 *        token layer tracks (it does NOT track method generators; see
 *        the "matches espree (division)" group below).
 *
 *   #4 — a function EXPRESSION's body `}` is a value, so a following
 *        `/` is division (`(function(){}/2)`); a function DECLARATION's
 *        body `}` begins a statement, so `/` is a regex
 *        (`function f(){} /re/`). The lexer used to treat both as
 *        declarations (regex).
 *
 * `sourceType: 'script'` so the sloppy `yield`-as-identifier form is
 * legal (it is a SyntaxError under `module`).
 */
import { describe, test, expect } from '@rstest/core';
import * as espree from 'espree';

import { tokenize } from '../../src/lexer/tokenizer.js';
import { buildLineStartOffsets } from '../../src/ast/normalize-ast.js';

interface SimpleToken {
  type: string;
  value: string;
  range: [number, number];
}

function runnerTokens(src: string): SimpleToken[] {
  const { tokens } = tokenize(src, buildLineStartOffsets(src), { jsx: false });
  return tokens.map((t) => ({ type: t.type, value: t.value, range: t.range }));
}

function espreeTokens(src: string): SimpleToken[] {
  const tokens = (
    espree as unknown as {
      tokenize: (
        s: string,
        o: object,
      ) => Array<{ type: string; value: string; range: [number, number] }>;
    }
  ).tokenize(src, {
    ecmaVersion: 'latest',
    sourceType: 'script',
    range: true,
    loc: false,
  });
  return tokens.map((t) => ({ type: t.type, value: t.value, range: t.range }));
}

/** Assert the runner and espree agree on the full token stream. */
function diff(src: string) {
  expect(runnerTokens(src)).toEqual(espreeTokens(src));
}

/** The type of the first token whose value is exactly `value`. */
function typeOfValue(src: string, value: string): string | undefined {
  return runnerTokens(src).find((t) => t.value === value)?.type;
}

/** Count of `/` Punctuator (division) tokens. */
function slashCount(src: string): number {
  return runnerTokens(src).filter(
    (t) => t.type === 'Punctuator' && t.value === '/',
  ).length;
}

describe('#2 function* / async function* — yield is a regex operand', () => {
  test('generator declaration: yield /re/ is a RegularExpression', () => {
    diff('function* g(){ yield /\\d+/; }');
    expect(typeOfValue('function* g(){ yield /\\d+/; }', '/\\d+/')).toBe(
      'RegularExpression',
    );
  });

  test('generator expression: yield /re/ is a RegularExpression', () =>
    diff('const g = function*(){ yield /x/g; };'));

  test('async generator: yield /re/ is a RegularExpression', () =>
    diff('async function* g(){ yield /x/; }'));

  test('yield in a nested block of a generator still sees the generator', () =>
    diff('function* g(){ if (a) { yield /x/; } }'));

  test('decoupled: expression generator — yield is regex, body } is division', () => {
    const src = '(function*(){ yield /x/; }/2/g)';
    diff(src);
    expect(typeOfValue(src, '/x/')).toBe('RegularExpression');
    // The two `/` after the body `}` are division — proves the
    // generator (yield) and decl/expr (body `}`) contexts are tracked
    // independently.
    expect(slashCount(src)).toBe(2);
  });
});

describe('#2 yield is division — matches espree token oracle', () => {
  test('sloppy non-generator function: yield / 2 is division', () => {
    const src = 'function f(){ return yield / 2; }';
    diff(src);
    expect(typeOfValue(src, '/')).toBe('Punctuator');
  });

  test('top-level (no enclosing function): yield / 2 is division', () =>
    diff('yield / 2 / g;'));

  test('nested non-generator function resets the context: division', () =>
    diff('function* g(){ function h(){ return yield / 2; } }'));

  // espree's token layer does NOT track generator context for method
  // shorthand — `*m(){ yield /x/ }` tokenizes `yield`'s `/` as DIVISION
  // even though it is semantically a generator. The runner matches that
  // (only `function*` seeds the generator stack), keeping the token
  // stream identical to what ESLint v10 rules see.
  test('object method generator: yield / is division (espree parity)', () => {
    const src = 'const o = { *m(){ yield /x/; } };';
    diff(src);
    expect(typeOfValue(src, '/')).toBe('Punctuator');
  });

  test('class method generator: yield / is division (espree parity)', () =>
    diff('class C { *m(){ yield /x/; } }'));

  test('static class method generator: yield / is division (espree parity)', () =>
    diff('class C { static *m(){ yield /x/; } }'));

  test('yield* delegate: the * is a Punctuator, not a generator star', () =>
    diff('function* g(){ yield* other(); }'));
});

describe('#4 function expression body } → division', () => {
  test('IIFE-style: (function(){}/2/g) divides', () => {
    const src = '(function(){}/2/g)';
    diff(src);
    expect(typeOfValue(src, '/2/g')).toBeUndefined();
    expect(slashCount(src)).toBe(2);
  });

  test('assigned function expression: = function(){} / 2 divides', () =>
    diff('const x = function(){} / 2;'));

  test('named function expression operand divides', () =>
    diff('const y = (function f(){} / 2);'));
});

describe('#4 control: function/class DECLARATION body } → regex (unchanged)', () => {
  test('function declaration: function f(){} /re/g is a regex', () => {
    const src = 'function f(){} /re/g';
    diff(src);
    expect(typeOfValue(src, '/re/g')).toBe('RegularExpression');
  });

  test('class declaration: class C {} /re/g is a regex', () =>
    diff('class C {} /re/g'));

  test('statement after declaration body: /re/.test(x) is a regex', () =>
    diff('function f(){}\n/re/.test(x);'));
});

describe('multiplication is not mis-read as a generator star (no regression)', () => {
  test('binary multiply: a * b', () => diff('const p = a * b;'));
  test('multiply after ): f() * 2', () => diff('const p = f() * 2;'));
  test('multiply after ]: arr[0] * 2', () => diff('const p = arr[0] * 2;'));
  test('chained multiply: 2 * 3 * 4', () => diff('const p = 2 * 3 * 4;'));
  test('regex then multiply inside a generator', () =>
    diff('function* g(){ const r = /x/ * 2; }'));
  test('exponent: a ** b unaffected', () => diff('const p = a ** b;'));
});

describe('#1b class EXPRESSION body } → division (declaration → regex)', () => {
  test('assigned class expression: x = class {} / 2 divides', () => {
    const src = 'x = class {} / 2';
    diff(src);
    expect(typeOfValue(src, '/')).toBe('Punctuator');
  });
  test('parenthesized class expression divides', () => diff('(class {} / 2)'));
  test('class expression as call arg divides', () => diff('f(class {} / 2);'));
  test('named class expression divides', () =>
    diff('const c = class C {} / 2;'));
  test('control: class DECLARATION body } begins a regex', () => {
    const src = 'class C {}\n/re/.test(0);';
    diff(src);
    expect(typeOfValue(src, '/re/')).toBe('RegularExpression');
  });
});

describe('#1c prefix ++/-- → regex operand; postfix → division', () => {
  test('prefix ++ before regex: ++/re/.lastIndex', () => {
    const src = '++/re/.lastIndex;';
    diff(src);
    expect(typeOfValue(src, '/re/')).toBe('RegularExpression');
  });
  test('prefix -- before regex: --/re/.lastIndex', () =>
    diff('--/re/.lastIndex;'));
  test('control: postfix ++ then division: i++ / 2', () => {
    const src = 'i++ / 2;';
    diff(src);
    expect(typeOfValue(src, '/')).toBe('Punctuator');
  });
  test('control: postfix -- then division: i-- / 2', () => diff('i-- / 2;'));
});

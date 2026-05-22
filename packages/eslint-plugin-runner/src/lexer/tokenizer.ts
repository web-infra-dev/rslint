/**
 * Lightweight JS/TS/JSX tokenizer for the SourceCode `getTokenBefore/
 * After` family.
 *
 * **Why a custom tokenizer**: oxc-parser does NOT expose a token list
 * (verified against oxc-parser@0.132.0 — only `program / module /
 * comments / errors` are emitted; there is no separate `oxc-lexer`
 * package). The ESLint token contract has subtle expectations the
 * runner needs to match (token kinds, value strings, range/loc shape,
 * JSXIdentifier / JSXText for JSX), so we keep a self-contained lexer
 * here. The JSX state machine lives in {@link jsx/tokenizer.ts} and
 * is only activated when `tokenize` is called with `jsx: true`.
 *
 * Scope: covers what the SourceCode token methods need, not full
 * compliance with the ECMAScript lexical grammar. Specifically we handle:
 *
 *   - whitespace & line terminators (skipped)
 *   - line comments + block comments (skipped — comments live on a
 *     separate `comments[]` field; SourceCode token methods filter
 *     comments by default)
 *   - string literals (single / double quotes) with escape sequences
 *   - template literals (with nested `${...}` expressions tokenized
 *     by recursing into the inner expression)
 *   - numeric literals (decimal, hex, binary, octal, BigInt suffix)
 *   - regex literals (heuristic; ambiguous with `/` operator — see
 *     {@link couldStartRegex})
 *   - identifiers & reserved words (incl. Unicode escapes)
 *   - punctuators including multi-char operators (===, !==, >>>, etc.)
 *   - JSX (opt-in via `TokenizeOptions.jsx`) — see `jsx/tokenizer.ts`
 */

import {
  offsetToLineColumn,
  type LocPosition,
  type SourceLocation,
} from '../ast/normalize-ast.js';
import {
  dispatchJsxStateOnEntry,
  tryCloseJsxExprContainer,
  tryEnterJsxFromAngle,
  trackExprBraceDepth,
  type JsxContext,
} from './jsx/tokenizer.js';

export type TokenType =
  | 'Identifier'
  | 'PrivateIdentifier'
  | 'Keyword'
  | 'Punctuator'
  | 'String'
  | 'Numeric'
  | 'RegularExpression'
  | 'Template'
  | 'Boolean'
  | 'Null'
  // JSX-only — emitted when `tokenize` is called with `jsx: true`.
  // `JSXIdentifier` covers tag names, attribute names, and member /
  // namespace name parts (each segment is one JSXIdentifier; the
  // separators `.` / `:` are Punctuators). Hyphens are allowed mid-
  // name (`data-foo`). `JSXText` covers raw text between tags
  // (including whitespace, HTML entities are preserved as-is).
  | 'JSXIdentifier'
  | 'JSXText';

export interface Token {
  type: TokenType;
  value: string;
  range: [number, number];
  loc: SourceLocation;
  /**
   * Only set for `RegularExpression` tokens — matches espree's shape.
   * Plugins like `eslint-plugin-regexp` and ESLint core's
   * `prefer-regex-literals` / `no-invalid-regexp` consult
   * `token.regex.pattern` / `.flags`; without this, every regex
   * literal would crash those rules.
   */
  regex?: { pattern: string; flags: string };
  /**
   * Internal — set only on `)` Punctuator tokens. `'ctrl'` means the
   * matching `(` was opened by a control-flow keyword (if/while/for/
   * switch/catch/with); `'expr'` covers call / grouping / function
   * params / arrow params. `couldStartRegex` reads this to decide
   * whether a `/` immediately after the `)` starts a regex literal
   * (`if (x) /re/.test(y)` — ctrl close) or is division
   * (`f(x) / 2` — expr close). Not part of the espree token contract
   * — leading `_` flags it as runtime-internal metadata and the JSON
   * wire path treats it as a harmless extra field.
   */
  _paren?: 'ctrl' | 'expr' | 'fn';
  /**
   * Internal — set only on `}` Punctuator tokens. `'block'` means the
   * matching `{` opened a statement block (function/control body,
   * arrow block body, top-level block); `'obj'` covers object
   * literals, class / interface bodies, and any other expression-
   * position `{`. `couldStartRegex` reads this to decide whether a
   * `/` after `}` starts a regex (`{} /re/` — block close, statement-
   * level) or is division (`{a:1} / 2` — object literal close).
   */
  _brace?: 'block' | 'obj';
  /**
   * Internal — two uses, both for generator-context tracking:
   *   - On a `yield` Keyword token: `true` when that `yield` sits inside
   *     a generator body, so `couldStartRegex` treats a following `/` as
   *     a regex (`function* g(){ yield /re/ }`) rather than division
   *     (sloppy non-generator `yield / 2`, where `yield` is an
   *     identifier). espree resolves this with full parser state; we
   *     track an enclosing-generator stack and stamp it here.
   *   - On a `)` Punctuator that closes a function/method params list:
   *     `true` when that function/method is a generator, read at the
   *     body `{` to seed the generator stack.
   */
  _gen?: boolean;
  /**
   * Internal — set on a `)` Punctuator that closes a function/method
   * params list. The following `{` opens that function's BODY (a fresh
   * lexical scope), so the generator stack pushes a frame seeded from
   * this `)`'s `_gen` rather than inheriting the enclosing context. A
   * plain block / object / class `{` (no `_fnBody` `)` before it)
   * inherits instead.
   */
  _fnBody?: boolean;
  /**
   * Internal — set on `++` / `--` Punctuator tokens. `true` when the
   * operator is PREFIX (`++/re/.lastIndex` — expression-prefix position, so
   * a following `/` opens a regex), `false` when POSTFIX (`i++ / 2` — the
   * operator completes a value, so `/` is division). espree resolves this
   * with full parser state; we stamp it from the expression-position signal
   * at emit time. `couldStartRegex` reads it.
   */
  _prefix?: boolean;
}

export interface Comment {
  // ESLint v10's three comment kinds. `Shebang` covers a `#!...`
  // hashbang on the first line — recognized only when it's the very
  // first byte of the file, per ECMAScript's HashbangComment grammar.
  type: 'Line' | 'Block' | 'Shebang';
  value: string;
  range: [number, number];
  loc: SourceLocation;
}

// Lexical keyword set — matches espree's tokenizer-emitted `Keyword`
// token type. **Contextual keywords are NOT here**: `await`, `async`,
// `of`, `as`, `from`, etc. show up as `Identifier` at the lexical
// layer and are interpreted contextually by the parser. Previously the
// set lumped them all in as keywords, which differed from ESLint and
// broke rules that branch on `token.type === 'Identifier'` for these
// names.
//
// `let`, `yield`, `static` ARE emitted as `Keyword` by espree's token
// layer (verified against espree@11: `obj.static`, `let / 2`,
// `yield / 2` all tokenize the word as `Keyword`), so they stay here.
//
// The strict-mode FutureReservedWords `enum`, `implements`, `interface`,
// `package`, `private`, `protected`, `public` are NOT here: espree's
// token layer emits them as `Identifier` in every position it produces
// a token at all — member access (`obj.public`), property key
// (`{ private: 1 }`), and sloppy bindings (`var implements = 1`); in
// strict positions where they'd be illegal, espree throws at parse
// time and produces no token. There is no case where espree yields a
// `Keyword` for these seven, so classifying them as `Keyword` diverged
// from ESLint and hid the identifier from `token.type === 'Identifier'`
// rules (`camelcase`, `id-length`, …) on common sloppy code.
const RESERVED_WORDS = new Set<string>([
  'break',
  'case',
  'catch',
  'class',
  'const',
  'continue',
  'debugger',
  'default',
  'delete',
  'do',
  'else',
  'export',
  'extends',
  'finally',
  'for',
  'function',
  'if',
  'import',
  'in',
  'instanceof',
  'let',
  'new',
  'return',
  'super',
  'switch',
  'this',
  'throw',
  'try',
  'typeof',
  'var',
  'void',
  'while',
  'with',
  'yield',
  'static',
]);

// Keywords after which a `/` is DIVISION, not a regex-literal start —
// their value completes an expression (postfix / value position):
//
//   - `this` / `super` — primary expressions.
//   - `static` / `let` — emitted as `Keyword` but appear in value
//     position too (`obj.static / 2`, `let / 2` with `let` as a
//     sloppy identifier) → division.
//
// `yield` is deliberately NOT in this set: it is context-dependent.
// In a generator body `yield /re/` is a regex operand (espree →
// RegularExpression); in sloppy non-generator code `yield / 2` is an
// identifier divided (espree → division). The `couldStartRegex`
// `Keyword` branch resolves it per-token from the `yield` token's
// `_gen` tag (stamped from an enclosing-generator stack), matching
// espree in both contexts.
//
// Contextual identifiers (`await`, `async`, `of`, `as`, `from`) are
// NOT special-cased: espree's token layer treats the `/` after ANY
// identifier as division (verified against espree@11 — `await
// /re/g.test(x)` tokenizes `/` as a Punctuator in both sloppy and
// async-module mode), so the `Identifier` branch of `couldStartRegex`
// returns `false` for all of them.
const REGEX_DIVISION_KEYWORDS = new Set<string>([
  'this',
  'super',
  'static',
  'let',
]);

const BOOLEAN_VALUES = new Set(['true', 'false']);
const NULL_VALUE = 'null';

/**
 * Keywords whose `(` opens a control-flow head. The matching `)`
 * leaves the lexer in expression-prefix position (next `/` is a regex
 * literal), unlike a call / grouping `)` which ends an expression
 * (next `/` is division). espree distinguishes via parser state; we
 * approximate by tracking what came before each `(`.
 */
const CONTROL_PAREN_KEYWORDS = new Set<string>([
  'if',
  'while',
  'for',
  'switch',
  'catch',
  'with',
]);

/**
 * Keywords after which a `{` opens a block, NOT an object literal.
 * `function`/`class` are intentionally excluded here — for those, the
 * `{` follows their `()` params / `extends` clause, so the brace-
 * context decision is made via the `)` close-paren `_paren` annotation
 * (a ctrl-paren `)` followed by `{` → block) rather than direct
 * keyword adjacency.
 */
const BLOCK_HEAD_KEYWORDS = new Set<string>(['else', 'do', 'try', 'finally']);

/**
 * Decide whether the upcoming `{` opens a block (statement context)
 * or an object literal (expression context). Used to tag the matching
 * `}` so `couldStartRegex` knows whether a trailing `/` is a regex
 * literal (block close) or division (object literal close).
 *
 * Heuristic — covers the common cases without a full parser:
 *   - prev is `)` whose matching `(` was a control-paren → block
 *     (`if (x) {…}`, `for (…) {…}`, …)
 *   - prev is a `=>` arrow Punctuator → block (arrow function body)
 *   - prev is `}` of an earlier block → block (block-following-block)
 *   - prev is a statement-terminator Punctuator (`;`, `:`) → block
 *   - prev is one of else/do/try/finally → block
 *   - prev is undefined (start of source) → block
 *   - prev is `)` of a `function`-declaration params list (tagged
 *     `_paren = 'fn'`) → block (function body)
 *   - otherwise → object literal
 *
 * Function declaration / method bodies are recognized via the `'fn'`
 * params-paren tag (set when a `function` keyword or generator `*`
 * armed `pendingFn`); class bodies are forced to a block at the `{`
 * site via `pendingClassBody`. Both make `function f() {} /re/` and
 * `class C {} /re/` tokenize the trailing `/re/` as a RegularExpression
 * (matching espree@11), instead of decomposing it into `/` `re` `/`.
 * Function EXPRESSIONS get `'expr'` params (not `'fn'`), so their body
 * `}` falls through to object-literal kind → a following `/` is
 * division (`(function(){}/2)`).
 */
function decideBraceKind(prev: Token | undefined): 'block' | 'obj' {
  if (!prev) return 'block';
  if (prev.type === 'Keyword') {
    return BLOCK_HEAD_KEYWORDS.has(prev.value) ? 'block' : 'obj';
  }
  if (prev.type === 'Punctuator') {
    if (prev.value === '=>' || prev.value === ';' || prev.value === ':') {
      return 'block';
    }
    if (prev.value === ')' && (prev._paren === 'ctrl' || prev._paren === 'fn'))
      return 'block';
    if (prev.value === '}' && prev._brace === 'block') return 'block';
  }
  return 'obj';
}

// Keywords after which a `function` is a DECLARATION (statement
// position), not an expression operand. `export` / `default` cover
// `export function f(){}` / `export default function(){}`; `else` /
// `do` / `case` cover the Annex-B statement-body forms.
const STATEMENT_HEAD_KEYWORDS = new Set<string>([
  'else',
  'do',
  'export',
  'default',
  'case',
]);

/**
 * Decide whether the just-emitted `function` / `class` keyword starts a
 * DECLARATION (statement position) or an EXPRESSION (operand position).
 * This drives the body `}` regex/division split: a declaration's body
 * `}` begins a regex (`function f(){} /re/`, `class C {} /re/`), an
 * expression's body `}` is a value → division (`(function(){}/2)`,
 * `x = class {} / 2`).
 *
 * `tokens` ends with the keyword itself. Heuristic on the token
 * immediately before it (skipping a leading contextual `async` for
 * `async function`):
 *   - start of source / `;` / `{` / block-close `}` / control `)` →
 *     declaration.
 *   - `else` / `do` / `export` / `default` / `case` → declaration.
 *   - anything else (`(`, `=`, `,`, `return`, operators, `=>`, `:`,
 *     object-close `}`, a value token) → expression.
 */
function keywordIsDeclaration(tokens: readonly Token[]): boolean {
  let idx = tokens.length - 2;
  if (
    idx >= 0 &&
    tokens[idx].type === 'Identifier' &&
    tokens[idx].value === 'async'
  ) {
    idx--;
  }
  const prev = idx >= 0 ? tokens[idx] : undefined;
  if (!prev) return true;
  if (prev.type === 'Keyword') return STATEMENT_HEAD_KEYWORDS.has(prev.value);
  if (prev.type === 'Punctuator') {
    if (prev.value === ';' || prev.value === '{') return true;
    if (prev.value === '}') return prev._brace === 'block';
    if (prev.value === ')') return prev._paren === 'ctrl';
  }
  return false;
}

export interface TokenizeOptions {
  /**
   * Enable JSX lexing. When true, `<` at expression-prefix positions
   * (after `=` / `(` / `,` / `=>` / control keywords / etc.) starts
   * a JSX element scan that emits `JSXIdentifier` / `JSXText` and
   * Punctuators for tag boundaries (`<` / `>` / `/` / `.` / `:` /
   * `=` / `{` / `}`). When false (default) the tokenizer treats every
   * source as plain JS / TS, which is correct for everything except
   * actual JSX content — `<` becomes a comparison Punctuator and
   * tag content fails to tokenize correctly.
   *
   * Set by the caller (ecma-language-plugin) based on
   * `parserOptions.ecmaFeatures.jsx === true` OR a `.jsx` / `.tsx`
   * file extension (oxc auto-infers JSX for those, so we must lex
   * accordingly).
   */
  jsx?: boolean;
}

/**
 * Tokenize the source text. Returns separate token / comment arrays
 * (matching ESLint's SourceCode shape).
 */
export function tokenize(
  text: string,
  lineStartOffsets: number[],
  options?: TokenizeOptions,
): {
  tokens: Token[];
  comments: Comment[];
} {
  const jsxEnabled = options?.jsx === true;
  const tokens: Token[] = [];
  const comments: Comment[] = [];
  let i = 0;
  const n = text.length;
  // Active template-literal expressions. When non-empty, the top
  // entry tracks how many `{` are open inside the current `${...}`.
  // A `}` encountered at `depth === 0` closes the expression and
  // re-enters template-segment scanning until the next `${` or `\``.
  const templateStack: { depth: number }[] = [];
  // Paren / brace context stacks. Populated when `(`/`{` are emitted,
  // consumed by the matching `)`/`}` to tag the close-Punctuator
  // with the kind. `couldStartRegex` reads those tags to decide
  // whether a following `/` is a regex literal or division. Each `(`
  // entry also carries `fnBody`/`gen` so the body `{` after its `)`
  // can seed the generator stack.
  const parenStack: Array<{
    kind: 'ctrl' | 'expr' | 'fn';
    fnBody?: boolean;
    gen?: boolean;
  }> = [];
  const braceStack: Array<'block' | 'obj'> = [];
  // Enclosing-generator stack, pushed/popped in lockstep with
  // `braceStack`. Its top answers "is the current position inside a
  // generator body?" — read when emitting a `yield` Keyword to decide
  // whether a following `/` is a regex operand (generator) or division
  // (sloppy `yield`-as-identifier). A function/method body `{` pushes
  // that function's own generator-ness (resetting the context); every
  // other `{` (plain block, object literal, class body) inherits the
  // enclosing frame.
  const genStack: boolean[] = [];
  // Function / class body tracking (regex-after-`}` + generator fix).
  //   - `pendingFn`: set when a `function` keyword — or a generator
  //     `*` introducing a `*method` — is seen; consumed by the next
  //     `(` (the params). `kind` is `'fn'` for declarations/methods
  //     (body `{` → block, so a trailing `/` after the body `}` starts
  //     a regex: `function f(){} /re/`) and `'expr'` for function
  //     EXPRESSIONS (body `}` is a value → division: `(function(){}/2)`).
  //     `isGen` flows to the body `{` so a `yield` inside resolves to a
  //     regex operand.
  //   - `pendingClassBody`: set when a `class` keyword is emitted;
  //     consumed by the next `{` (the class body), forcing it to a
  //     block. (A `{` object literal inside `extends …` heritage is a
  //     rare contrived case we accept mis-tagging.)
  let pendingFn: { kind: 'fn' | 'expr'; isGen: boolean } | null = null;
  let pendingClassBody = false;
  // JSX state machine. Only consulted when `jsxEnabled` is true; the
  // empty-stack fast path leaves non-JSX files completely unchanged.
  // See `./jsx/tokenizer.ts` for the lifecycle / state details.
  const jsxStack: JsxContext[] = [];

  // ── Hashbang ──────────────────────────────────────────────────────
  // Per ECMAScript's HashbangComment production, a `#!` at the very
  // first byte (offset 0) of the source is consumed as a single
  // shebang comment running through the first line terminator. ESLint
  // exposes this as a Comment of type 'Shebang' on `getAllComments()`
  // (the `#!` prefix is stripped from `.value`).
  //
  // Without this branch the rest of the tokenizer fell into its
  // generic `#` / `!` / `/` paths and emitted 5-10 spurious tokens
  // for what should be a single comment — directly observable on any
  // CLI script, and broke rules that count tokens or walk comments
  // (`getAllComments`, `getInlineConfigNodes` for `eslint-disable` on
  // line 1, etc.).
  if (
    n >= 2 &&
    text.charCodeAt(0) === 35 /* # */ &&
    text.charCodeAt(1) === 33 /* ! */
  ) {
    i = 2;
    while (i < n && !isLineTerminator(text.charCodeAt(i))) i++;
    comments.push(makeComment('Shebang', text, 0, i, lineStartOffsets));
    // Don't consume the trailing LF — the main loop's whitespace skip
    // handles it, keeping line-start offsets aligned with the rest of
    // the file.
  }

  while (i < n) {
    // JSX state dispatch — owns JSXText / tag-content scans. Falls
    // through (`handled === false`) when the stack is empty or its
    // top is an `expr` container, so the JS scanner below takes over
    // the inner expression. See `./jsx/tokenizer.ts`.
    if (jsxEnabled) {
      const r = dispatchJsxStateOnEntry(
        text,
        i,
        n,
        lineStartOffsets,
        tokens,
        jsxStack,
      );
      if (r.handled) {
        i = r.i;
        continue;
      }
    }

    const ch = text.charCodeAt(i);

    // Whitespace
    if (isWhitespace(ch)) {
      i++;
      continue;
    }

    // Line comment
    if (ch === 47 /* / */ && text.charCodeAt(i + 1) === 47 /* / */) {
      const start = i;
      i += 2;
      while (i < n && !isLineTerminator(text.charCodeAt(i))) i++;
      comments.push(makeComment('Line', text, start, i, lineStartOffsets));
      continue;
    }
    // Block comment
    if (ch === 47 /* / */ && text.charCodeAt(i + 1) === 42 /* * */) {
      const start = i;
      i += 2;
      while (
        i < n - 1 &&
        !(text.charCodeAt(i) === 42 && text.charCodeAt(i + 1) === 47)
      )
        i++;
      // Detect actual termination — the while above exits either at
      // `*‍/` (terminated) or at `i >= n - 1` (EOF, unterminated).
      // `makeComment` skips the trailing 2-char strip when unterminated
      // so `/* unterminated TODO` keeps its real last 2 chars (and
      // `/*` alone yields '' instead of the off-by-two negative slice).
      const terminated =
        i < n - 1 && text.charCodeAt(i) === 42 && text.charCodeAt(i + 1) === 47;
      i = Math.min(i + 2, n);
      comments.push(
        makeComment('Block', text, start, i, lineStartOffsets, terminated),
      );
      continue;
    }

    // String literal
    if (ch === 34 /* " */ || ch === 39 /* ' */) {
      const start = i;
      const quote = ch;
      i++;
      while (i < n) {
        const c = text.charCodeAt(i);
        if (c === 92 /* \ */) {
          // `\` + CRLF is a single LineContinuation sequence (3 chars)
          // — a string may legally span the line break. Pre-fix the
          // unconditional `i += 2` skipped `\`+CR and landed on the
          // LF, which the line-terminator check below then treated as
          // an unterminated-string boundary, truncating the token and
          // cascading every downstream token on the line. Other escapes
          // (`\"`, `\n`, `\\`, `\`+LF, `\`+CR-alone, `\`+LS/PS) are
          // 2 chars and handled by the else branch.
          if (
            text.charCodeAt(i + 1) === 13 /* CR */ &&
            text.charCodeAt(i + 2) === 10 /* LF */
          ) {
            i += 3;
          } else {
            i += 2;
          }
          continue;
        }
        if (c === quote) {
          i++;
          break;
        }
        if (isLineTerminator(c)) {
          i++;
          break;
        } // unterminated; recover at line terminator
        i++;
      }
      tokens.push(makeToken('String', text, start, i, lineStartOffsets));
      continue;
    }

    // Template literal — ESLint splits each template into N+1 Template
    // tokens (one per literal segment) with the expression tokens
    // interleaved between them. Previous impl emitted the whole thing
    // as a single Template token, which broke `getTokenBefore/After`
    // and downstream spacing/prefer-template rules. Each `${` opens an
    // expression: tokens get scanned normally until the matching `}`
    // (tracked per-depth on `templateStack`), which closes the
    // expression and resumes template-segment scanning.
    if (ch === 96 /* ` */) {
      const start = i;
      i++;
      let didOpenExpr = false;
      while (i < n) {
        const c = text.charCodeAt(i);
        if (c === 92 /* \ */) {
          i += 2;
          continue;
        }
        if (c === 96) {
          i++;
          break;
        }
        if (c === 36 /* $ */ && text.charCodeAt(i + 1) === 123 /* { */) {
          i += 2; // consume `${`
          didOpenExpr = true;
          break;
        }
        i++;
      }
      tokens.push(makeToken('Template', text, start, i, lineStartOffsets));
      if (didOpenExpr) {
        // Enter expression mode; the matching `}` (at depth 0) is
        // handled below by the punctuator branch via templateStack.
        templateStack.push({ depth: 0 });
      }
      continue;
    }

    // Numeric literal
    if (isDigit(ch) || (ch === 46 && isDigit(text.charCodeAt(i + 1)))) {
      const start = i;
      // hex / oct / bin
      if (
        ch === 48 /* 0 */ &&
        (text[i + 1] === 'x' ||
          text[i + 1] === 'X' ||
          text[i + 1] === 'o' ||
          text[i + 1] === 'O' ||
          text[i + 1] === 'b' ||
          text[i + 1] === 'B')
      ) {
        i += 2;
        while (i < n && /[0-9a-fA-F_]/.test(text[i])) i++;
      } else {
        while (i < n && /[0-9_]/.test(text[i])) i++;
        if (text[i] === '.') {
          i++;
          while (i < n && /[0-9_]/.test(text[i])) i++;
        }
        if (text[i] === 'e' || text[i] === 'E') {
          i++;
          if (text[i] === '+' || text[i] === '-') i++;
          // Exponent digits accept the ES2021 numeric separator `_`
          // (e.g. `1e1_0`), matching the integer / fraction parts
          // above. Pre-fix this used `isDigit` only, so `1e1_0` lexed
          // as `Numeric:1e1` + `Identifier:_0`.
          while (i < n && /[0-9_]/.test(text[i])) i++;
        }
      }
      if (text[i] === 'n') i++; // BigInt
      tokens.push(makeToken('Numeric', text, start, i, lineStartOffsets));
      continue;
    }

    // Regex literal — heuristic: only if previous token suggests
    // expression context (no `)`, `]`, identifier, numeric).
    if (ch === 47 /* / */ && couldStartRegex(tokens)) {
      const start = i;
      i++;
      let inClass = false;
      let closed = false;
      while (i < n) {
        const c = text.charCodeAt(i);
        if (c === 92 /* \ */) {
          i += 2;
          continue;
        }
        if (c === 91 /* [ */) inClass = true;
        else if (c === 93 /* ] */) inClass = false;
        else if (c === 47 /* / */ && !inClass) {
          i++;
          closed = true;
          break;
        } else if (isLineTerminator(c)) break;
        i++;
      }
      // The slash that closed the body sits at `i - 1`; flags follow.
      // When the regex hits EOL without a closing `/`, `i` points at
      // the line terminator and the body extends through the previous
      // char — so the pattern slice MUST include `text[i-1]`. Pre-fix
      // the `slashEnd - 1` formula always trimmed one char, dropping
      // the real last body character on unterminated regexes (e.g.
      // `/abc<EOL>` → pattern `'ab'` instead of `'abc'`).
      const slashEnd = i;
      const bodyEnd = closed ? slashEnd - 1 : slashEnd;
      // flags are only valid when the regex actually closed.
      if (closed) while (i < n && /[a-zA-Z]/.test(text[i])) i++;
      const tok = makeToken(
        'RegularExpression',
        text,
        start,
        i,
        lineStartOffsets,
      );
      // espree-compatible regex metadata. Without this, plugins
      // reading `token.regex.pattern` / `.flags` throw — see the
      // `regex` field comment on the Token interface.
      tok.regex = {
        pattern: text.slice(start + 1, Math.max(start + 1, bodyEnd)),
        flags: closed ? text.slice(slashEnd, i) : '',
      };
      tokens.push(tok);
      continue;
    }

    // Identifier / keyword / boolean / null. ECMAScript also accepts
    // Unicode-escape sequences (`\uXXXX` and `\u{...}`) as identifier
    // chars; espree decodes them so the token's `value` is the
    // resolved string (`a` → `'a'`) while `range` spans the
    // literal source text (the backslash through the last hex digit).
    // Pre-fix the runner had no escape branch — the `\` fell to the
    // "Unknown char — skip" path below, dropping the leading char and
    // emitting an identifier whose `value` and `range` were both off
    // (`var a = 1` got `Identifier{value:'u0061',range:[5,10]}`
    // vs espree's `Identifier{value:'a',range:[4,10]}`).
    if (isIdentStart(ch) || isIdentEscape(text, i)) {
      const start = i;
      let value = '';
      let first = true;
      while (i < n) {
        if (isIdentEscape(text, i)) {
          const dec = decodeIdentEscape(text, i);
          if (dec == null) break;
          const okStart = first ? isIdentStart(dec.codepoint) : true;
          const okCont = !first ? isIdentContinue(dec.codepoint) : true;
          if (!okStart || !okCont) break;
          value += String.fromCodePoint(dec.codepoint);
          i = dec.next;
        } else {
          const c = text.charCodeAt(i);
          if (first ? !isIdentStart(c) : !isIdentContinue(c)) break;
          // Use `slice` so surrogate-pair source positions are
          // preserved verbatim in the accumulated value.
          value += text[i];
          i++;
        }
        first = false;
      }
      if (value.length === 0) {
        // Lone `\` with no valid escape — advance past it so we don't
        // loop forever, but don't emit a degenerate Identifier.
        i = Math.max(i + 1, start + 1);
        continue;
      }
      const type: TokenType = BOOLEAN_VALUES.has(value)
        ? 'Boolean'
        : value === NULL_VALUE
          ? 'Null'
          : RESERVED_WORDS.has(value)
            ? 'Keyword'
            : 'Identifier';
      tokens.push({
        type,
        value,
        range: [start, i],
        loc: locOf(start, i, lineStartOffsets),
      });
      // Arm the declaration-body trackers (see flag decls). `function`
      // → its params `(` becomes `'fn'`; `class` → its body `{` becomes
      // a block. Harmless if the keyword is a member name (`obj.class`):
      // no params `(` / body `{` follows in that position.
      if (type === 'Keyword') {
        if (value === 'function') {
          // Declaration vs expression decides the body `}` regex/division
          // (#4); generator-ness (set later by a `*`) decides `yield`.
          pendingFn = {
            kind: keywordIsDeclaration(tokens) ? 'fn' : 'expr',
            isGen: false,
          };
        } else if (value === 'class') {
          // A class DECLARATION's body `}` begins a regex; a class
          // EXPRESSION's body `}` is a value → division (`x = class {} / 2`,
          // matching espree). Force the body `{` to a block only for
          // declarations (#1b) — same statement-vs-expression split as
          // `function` above.
          pendingClassBody = keywordIsDeclaration(tokens);
        } else if (value === 'yield') {
          // Stamp the enclosing-generator context so `couldStartRegex`
          // can tell `yield /re/` (generator → regex) from `yield / 2`
          // (sloppy identifier → division).
          tokens[tokens.length - 1]._gen =
            genStack[genStack.length - 1] ?? false;
        }
      }
      continue;
    }

    // Template-expression close: a `}` at depth 0 inside an active
    // `${...}` re-enters template-segment scanning instead of being
    // emitted as a Punctuator. We scan until the next `${` (push
    // depth+1 on the stack) or matching `\`` (pop the stack).
    if (
      ch === 125 /* } */ &&
      templateStack.length > 0 &&
      templateStack[templateStack.length - 1].depth === 0
    ) {
      const start = i;
      i++;
      let closedTemplate = false;
      let openedExpr = false;
      while (i < n) {
        const c = text.charCodeAt(i);
        if (c === 92 /* \ */) {
          i += 2;
          continue;
        }
        if (c === 96) {
          i++;
          closedTemplate = true;
          break;
        }
        if (c === 36 /* $ */ && text.charCodeAt(i + 1) === 123 /* { */) {
          i += 2;
          openedExpr = true;
          break;
        }
        i++;
      }
      tokens.push(makeToken('Template', text, start, i, lineStartOffsets));
      if (closedTemplate) {
        templateStack.pop();
      } else if (!openedExpr) {
        // Source ran out before we found `${` or `\`` — malformed
        // template. Bail out of expression mode to keep the rest of
        // the file tokenizing in a sane state.
        templateStack.pop();
      }
      continue;
    }

    // JSX expression-container close: `}` at depth 0 pops the `expr`
    // frame so the surrounding JSX text / tag context resumes. Must
    // run AFTER the template-close branch above (which has its own
    // `}@depth=0` semantics for `${...}` resumption), and BEFORE the
    // generic Punctuator emit which would otherwise stamp this `}`
    // with the JS `braceStack` semantics.
    if (ch === 125 /* } */ && jsxEnabled) {
      if (
        tryCloseJsxExprContainer(text, i, lineStartOffsets, tokens, jsxStack)
      ) {
        i++;
        continue;
      }
    }

    // JSX entry on `<`. The lookahead in `classifyJsxLAngle` rules out
    // comparison (`a < b`) and TSX generics (`<T,>` / `<T extends X>` /
    // `<T>(...)`) — only true JSX shapes return non-null.
    if (ch === 60 /* < */ && jsxEnabled) {
      const r = tryEnterJsxFromAngle(
        text,
        i,
        n,
        lineStartOffsets,
        tokens,
        jsxStack,
      );
      if (r.entered) {
        i = r.i;
        continue;
      }
    }

    // PrivateIdentifier: `#name` — ECMAScript 2022 class private
    // members. ESLint tokenizes these as type `PrivateIdentifier`
    // with `value` stripped of the leading `#`.
    if (ch === 35 /* # */ && isIdentStart(text.charCodeAt(i + 1))) {
      const start = i;
      i++; // consume `#`
      const nameStart = i;
      while (i < n && isIdentContinue(text.charCodeAt(i))) i++;
      const value = text.slice(nameStart, i);
      tokens.push({
        type: 'PrivateIdentifier',
        value,
        range: [start, i],
        loc: locOf(start, i, lineStartOffsets),
      });
      continue;
    }

    // Punctuator (multi-char first, then single-char)
    const punctLen = matchPunctuator(text, i);
    if (punctLen > 0) {
      // Track `{` / `}` depth inside an active template expression so
      // we can distinguish nested blocks from the closing `}` of the
      // `${...}` (handled above when depth hits 0).
      if (punctLen === 1 && templateStack.length > 0) {
        if (ch === 123 /* { */) {
          templateStack[templateStack.length - 1].depth++;
        } else if (ch === 125 /* } */) {
          templateStack[templateStack.length - 1].depth--;
        }
      }
      // Same depth tracking for the active JSX expression container.
      // The container's outer `{`/`}` are handled by the JSX
      // open/close branches; this only tracks nested JS braces so
      // the depth-0 check above correctly identifies the container's
      // closing `}`.
      if (punctLen === 1 && jsxEnabled) {
        trackExprBraceDepth(jsxStack, ch);
      }
      const tok = makeToken(
        'Punctuator',
        text,
        i,
        i + punctLen,
        lineStartOffsets,
      );
      // #1c: tag a `++` / `--` as PREFIX vs POSTFIX. A prefix operator sits
      // at expression-prefix position, so a following `/` opens a regex
      // (`++/re/.lastIndex`); a postfix operator completes a value, so `/`
      // is division (`i++ / 2`). `couldStartRegex(tokens)` here — `tokens`
      // still ends at the token BEFORE this operator — is exactly that test.
      if (tok.value === '++' || tok.value === '--') {
        tok._prefix = couldStartRegex(tokens);
      }
      // Paren / brace context tracking. Pre-fix `couldStartRegex`
      // unconditionally treated `)` and `}` as expression-end (next
      // `/` is division). ESLint v10 uses parser state to know
      // `if (x) /re/.test()` and `if (x) {} /re/` both have `/`
      // starting a regex literal. We approximate by tagging each `)`
      // / `}` with the kind of the matching open paren / brace.
      if (punctLen === 1) {
        const prev = tokens[tokens.length - 1];
        switch (ch) {
          case 40 /* ( */: {
            // A function/method params `(` (consumes `pendingFn`) is
            // tagged `'fn'` for declarations/methods (body `{` → block,
            // so the body `}` starts a regex) or `'expr'` for function
            // EXPRESSIONS (body `}` → division). It also carries
            // `fnBody`/`gen` so the body `{` seeds the generator stack.
            // A control-head `(` → 'ctrl'; everything else → 'expr'.
            if (pendingFn) {
              parenStack.push({
                kind: pendingFn.kind,
                fnBody: true,
                gen: pendingFn.isGen,
              });
              pendingFn = null;
            } else if (
              prev?.type === 'Keyword' &&
              CONTROL_PAREN_KEYWORDS.has(prev.value)
            ) {
              parenStack.push({ kind: 'ctrl' });
            } else {
              parenStack.push({ kind: 'expr' });
            }
            break;
          }
          case 41 /* ) */: {
            // Default to 'expr' for unmatched `)` (defensive — should
            // not happen on valid input).
            const e = parenStack.pop() ?? { kind: 'expr' as const };
            tok._paren = e.kind;
            if (e.fnBody) {
              // The `{` after this `)` opens a function/method body.
              tok._fnBody = true;
              tok._gen = e.gen;
            }
            break;
          }
          case 42 /* * */: {
            // `function* g(){}` / `async function* g(){}`: a `*` right
            // after the `function` keyword makes the function a
            // generator, so a `yield` in its body is a regex-operand
            // position. (`pendingFn` is set by the `function` keyword;
            // guard the null case defensively.)
            //
            // Method generators (`*m(){}`, `static *m(){}`) are
            // deliberately NOT tracked: espree's TOKEN oracle does not
            // track generator context for method shorthand either —
            // `({ *m(){ yield /x/ } })` and `class C { *m(){ yield /x/ } }`
            // both tokenize `yield`'s `/` as DIVISION (verified against
            // espree@11.2.0). Matching that keeps the runner's token
            // stream byte-identical to espree, which is the contract
            // ESLint v10 rules rely on. (A `yield`-delegate `yield *` and
            // ordinary multiplication `a * b` also fall through here as a
            // plain Punctuator, which is correct.)
            if (prev?.type === 'Keyword' && prev.value === 'function') {
              if (pendingFn) pendingFn.isGen = true;
              else pendingFn = { kind: 'fn', isGen: true };
            }
            break;
          }
          case 123 /* { */: {
            // A `class` body `{` (consumes the pending flag) is always
            // a block; otherwise fall back to the prev-token heuristic.
            let kind: 'block' | 'obj';
            if (pendingClassBody) {
              kind = 'block';
              pendingClassBody = false;
            } else {
              kind = decideBraceKind(prev);
            }
            braceStack.push(kind);
            // Generator frame, pushed in lockstep with `braceStack`. A
            // function/method body `{` (its `)` carried `_fnBody`) resets
            // the context to that function's own generator-ness; every
            // other `{` (plain block, object, class body, template
            // interpolation) inherits the enclosing frame so a `yield`
            // in a nested block still sees the surrounding generator.
            if (prev?.value === ')' && prev._fnBody) {
              genStack.push(prev._gen === true);
            } else {
              genStack.push(genStack[genStack.length - 1] ?? false);
            }
            break;
          }
          case 125 /* } */: {
            // Skip if this `}` is a template-expression closer
            // (handled by the `}@depth=0` branch above). Otherwise
            // pop the brace + generator stacks in lockstep.
            tok._brace = braceStack.pop() ?? 'obj';
            genStack.pop();
            break;
          }
        }
      }
      tokens.push(tok);
      i += punctLen;
      continue;
    }

    // Unknown char — skip and keep going (parser would have caught it).
    i++;
  }

  return { tokens, comments };
}

const PUNCTUATORS_BY_LEN = [
  // 4-char
  ['>>>='],
  // 3-char
  ['===', '!==', '>>>', '...', '**=', '<<=', '>>=', '&&=', '||=', '??=', '?.'],
  // 2-char
  [
    '==',
    '!=',
    '<=',
    '>=',
    '&&',
    '||',
    '??',
    '++',
    '--',
    '<<',
    '>>',
    '+=',
    '-=',
    '*=',
    '/=',
    '%=',
    '&=',
    '|=',
    '^=',
    '=>',
    '**',
  ],
  // 1-char
  [
    '{',
    '}',
    '(',
    ')',
    '[',
    ']',
    '.',
    ';',
    ',',
    '<',
    '>',
    '+',
    '-',
    '*',
    '/',
    '%',
    '&',
    '|',
    '^',
    '!',
    '~',
    '?',
    ':',
    '=',
    '@',
  ],
];

function matchPunctuator(text: string, at: number): number {
  // `?.` is special: per ECMAScript spec, optional chain `?.` is NOT
  // permitted when immediately followed by a DecimalDigit. The grammar
  // disambiguates `cond?.4:.2` as `cond` `?` `.4` `:` `.2` (ternary
  // with leading-dot numerics), NOT `cond` `?.` `4` `:` `.2`. The
  // generic 3-char matcher below would greedily eat `?.`, so we
  // pre-check this one case and emit a bare `?` punctuator when the
  // lookahead is a digit.
  if (
    text.charCodeAt(at) === 63 /* ? */ &&
    text.charCodeAt(at + 1) === 46 /* . */
  ) {
    const next = text.charCodeAt(at + 2);
    if (next >= 48 && next <= 57 /* 0-9 */) return 1; // bare `?`
    return 2; // `?.`
  }
  for (const group of PUNCTUATORS_BY_LEN) {
    for (const p of group) {
      if (text.startsWith(p, at)) return p.length;
    }
  }
  return 0;
}

export function makeToken(
  type: TokenType,
  text: string,
  start: number,
  end: number,
  lso: number[],
): Token {
  return {
    type,
    value: text.slice(start, end),
    range: [start, end],
    loc: locOf(start, end, lso),
  };
}

function makeComment(
  type: 'Line' | 'Block' | 'Shebang',
  text: string,
  start: number,
  end: number,
  lso: number[],
  /**
   * Only meaningful for `Block` — whether the source actually contained
   * the closing `*‍/` before `end`. The Block branch otherwise blindly
   * strips two trailing chars, which lops the real last two characters
   * off an unterminated `/* unterminated TODO` comment (and produces
   * an empty `value` for the degenerate `/*` case). For Line / Shebang
   * the parameter is ignored.
   */
  terminated: boolean = true,
): Comment {
  // Strip the comment delimiters from `value` to match ESLint's convention:
  //   - Line:    `// foo`     → ' foo'  (skip leading 2 `//`)
  //   - Block:   `/* foo */`  → ' foo ' (skip leading 2, trailing 2)
  //   - Shebang: `#!/usr/bin` → '/usr/bin' (skip leading 2 `#!`)
  let value: string;
  if (type === 'Line') value = text.slice(start + 2, end);
  else if (type === 'Shebang') value = text.slice(start + 2, end);
  else if (terminated) value = text.slice(start + 2, end - 2);
  else value = text.slice(start + 2, end);
  return {
    type,
    value,
    range: [start, end],
    loc: locOf(start, end, lso),
  };
}

export function locOf(
  start: number,
  end: number,
  lso: number[],
): SourceLocation {
  return {
    start: offsetToLineColumn(start, lso),
    end: offsetToLineColumn(end, lso),
  };
}

/**
 * ECMAScript §11.3 LineTerminator: LF, CR, U+2028 LINE SEPARATOR,
 * U+2029 PARAGRAPH SEPARATOR. Shebang / line comment / unterminated
 * string-recovery scans MUST stop on any of these — checking only LF
 * (10) silently keeps reading on CRLF (Windows) tails and old-Mac CR
 * line endings, producing tokens whose `value` includes the trailing
 * `\r` (or the whole file in the pure-CR case).
 */
export function isLineTerminator(ch: number): boolean {
  return ch === 10 || ch === 13 || ch === 0x2028 || ch === 0x2029;
}

export function isWhitespace(ch: number): boolean {
  // ECMAScript §11.2 WhiteSpace + §11.3 LineTerminator — every char
  // the lexer skips between tokens. ASCII set (space/tab/LF/CR/VT/FF)
  // plus Unicode line separators (U+2028/U+2029) plus the
  // Space_Separator (Zs) code points + Zero-Width-No-Break-Space.
  //
  // Why this is critical: `isIdentStart` falls back to `ch > 127` for
  // non-ASCII identifier starts. Any space-like char in that range
  // that ISN'T listed here gets swallowed by the next-identifier
  // scan, gluing the whitespace onto the following identifier. Real
  // sources hit U+00A0 (NBSP — common when code is pasted from web
  // editors), U+FEFF (ZWNBSP mid-source), and the U+2000-200A space
  // family from Pretty-Printed docs.
  if (
    ch === 32 || // SPACE
    ch === 9 || // TAB
    ch === 10 || // LF
    ch === 13 || // CR
    ch === 11 || // VT
    ch === 12 || // FF
    ch === 0x2028 || // LINE SEPARATOR
    ch === 0x2029 || // PARAGRAPH SEPARATOR
    ch === 0x00a0 || // NO-BREAK SPACE
    ch === 0xfeff || // ZERO-WIDTH NO-BREAK SPACE (BOM in body)
    ch === 0x1680 || // OGHAM SPACE MARK
    ch === 0x202f || // NARROW NO-BREAK SPACE
    ch === 0x205f || // MEDIUM MATHEMATICAL SPACE
    ch === 0x3000 // IDEOGRAPHIC SPACE
  ) {
    return true;
  }
  // U+2000..U+200A — EN QUAD through HAIR SPACE
  if (ch >= 0x2000 && ch <= 0x200a) return true;
  return false;
}
function isDigit(ch: number): boolean {
  return ch >= 48 && ch <= 57;
}
/** Does `text[i..]` look like the start of a Unicode identifier escape? */
function isIdentEscape(text: string, i: number): boolean {
  return (
    text.charCodeAt(i) === 92 /* \ */ && text.charCodeAt(i + 1) === 117 /* u */
  );
}

/**
 * Decode a `\uXXXX` or `\u{...}` escape starting at offset `i`.
 * Returns the resolved codepoint and the offset one past the last
 * escape character, or null if the escape is malformed.
 */
function decodeIdentEscape(
  text: string,
  i: number,
): { codepoint: number; next: number } | null {
  // \u{ABCDE} variable-length form (ES2015+)
  if (text.charCodeAt(i + 2) === 123 /* { */) {
    let j = i + 3;
    while (j < text.length && isHexDigit(text.charCodeAt(j))) j++;
    if (text.charCodeAt(j) !== 125 /* } */) return null;
    const hex = text.slice(i + 3, j);
    if (hex.length === 0 || hex.length > 6) return null;
    const cp = parseInt(hex, 16);
    if (cp > 0x10ffff) return null;
    return { codepoint: cp, next: j + 1 };
  }
  // \uXXXX fixed 4-hex form
  if (i + 6 > text.length) return null;
  for (let k = 0; k < 4; k++) {
    if (!isHexDigit(text.charCodeAt(i + 2 + k))) return null;
  }
  const cp = parseInt(text.slice(i + 2, i + 6), 16);
  return { codepoint: cp, next: i + 6 };
}

function isHexDigit(ch: number): boolean {
  return (
    (ch >= 48 && ch <= 57) || // 0-9
    (ch >= 65 && ch <= 70) || // A-F
    (ch >= 97 && ch <= 102) // a-f
  );
}

export function isIdentStart(ch: number): boolean {
  // ASCII fast-path first.
  if (
    (ch >= 65 && ch <= 90) || // A-Z
    (ch >= 97 && ch <= 122) || // a-z
    ch === 95 ||
    ch === 36 // _ $
  ) {
    return true;
  }
  if (ch <= 127) return false;
  // Non-ASCII fallback — permit anything with `ch > 127` EXCEPT
  // Unicode whitespace / line terminators (NBSP, U+2028/U+2029,
  // the U+2000..200A space family, etc.). Without this exclusion,
  // pasted source that uses NBSP between tokens (common in editors
  // that prettify HTML) would glue the NBSP onto the next
  // identifier, turning `const a` (with NBSP) into a single 7-char
  // identifier and shifting every downstream token. Real ECMAScript
  // also restricts to chars with the Unicode `ID_Start` property,
  // but the simpler "not whitespace" filter catches the bug class
  // we actually see in real source.
  return !isWhitespace(ch);
}
export function isIdentContinue(ch: number): boolean {
  return isIdentStart(ch) || isDigit(ch);
}

/**
 * Heuristic for `/` ambiguity: a slash starts a regex when the previous
 * token leaves us in expression position (no preceding closing paren /
 * bracket / identifier / numeric). This mirrors the standard ESLint
 * tokenizer heuristic; it is correct for almost all real source.
 */
export function couldStartRegex(tokens: readonly Token[]): boolean {
  if (tokens.length === 0) return true;
  const prev = tokens[tokens.length - 1];
  switch (prev.type) {
    case 'Identifier':
      // After ANY identifier token, `/` is division — an identifier is
      // a value-producing expression, never a regex-prefix position.
      // This includes contextual keywords (`await`, `async`, `of`,
      // `as`, `from`): espree's token layer treats `/` after them as
      // division too (verified against espree@11 — `await /re/g.test(x)`
      // → `await` `/` `re` `/` `g` … in both sloppy and async-module
      // mode). An earlier version wrongly returned regex for `await`
      // and swallowed the rest of the line into a phantom
      // RegularExpression. (`yield` is a `Keyword`, handled above.)
      return false;
    case 'Keyword':
      // `yield` is context-dependent: inside a generator body the
      // following `/` opens a regex operand (`yield /re/`); in sloppy
      // non-generator code `yield` is an identifier and `/` is division
      // (`yield / 2`). The `_gen` tag (stamped at emit time from the
      // enclosing-generator stack) records which — matching espree,
      // which resolves it with full parser state.
      if (prev.value === 'yield') return prev._gen === true;
      // Keywords whose value COMPLETES an expression (postfix / value
      // position) → a following `/` is division, not a regex start.
      // Without this, `const r = this / 2;` (and `obj.static / 2`,
      // `let / 2`) would mis-classify `/` as regex-start and consume
      // the rest of the line into a RegularExpression token. See
      // `REGEX_DIVISION_KEYWORDS` for the set.
      if (REGEX_DIVISION_KEYWORDS.has(prev.value)) return false;
      // Other keywords (return, typeof, throw, in, …) are
      // expression-prefix; `/` after them starts a regex literal.
      return true;
    case 'Punctuator':
      // `++` / `--` are context-sensitive (#1c). POSTFIX (`i++ / 2`)
      // completes a value → a following `/` is division; PREFIX
      // (`++/re/.lastIndex`) sits at expression-prefix position → `/`
      // opens a regex. The emit site tags each operator with `_prefix`
      // (espree resolves the same split with full parser state).
      if (prev.value === '++' || prev.value === '--')
        return prev._prefix === true;
      // `)` and `}` are context-sensitive — they can end either an
      // expression (call paren / object literal close → division) or
      // a statement-control construct (`if (x)` close / block close
      // → regex). The tokenizer tags each `)` / `}` with the matching
      // open's kind so we can branch correctly here. Pre-fix the
      // bare `!['}', ')'].includes(...)` returned false for every
      // case → `if (x) /re/.test()` and `if (x) {} /re/` both lost
      // their regex literal to a 3-token decomposition.
      if (prev.value === ')') return prev._paren === 'ctrl';
      if (prev.value === '}') return prev._brace === 'block';
      // `]` always closes an index expression → division.
      return prev.value !== ']';
    case 'Template':
      // Template HEAD (`` `…${ ``) and MIDDLE (`` }…${ ``) end with `${`
      // and leave the lexer in expression position inside the
      // interpolation — `/` immediately after MUST be a regex start
      // (`` `${/re/g.test(x)}` ``). Complete templates (`` `…` ``) and
      // template TAILS (`` }…` ``) are postfix-position values, so `/`
      // after them is division.
      return prev.value.endsWith('${');
    case 'Numeric':
    case 'String':
    case 'RegularExpression':
    case 'Boolean':
    case 'Null':
    case 'PrivateIdentifier':
    case 'JSXIdentifier':
    case 'JSXText':
      // A PrivateIdentifier (`#x`) sits in postfix position
      // (`this.#x / 2`), so a following `/` is division, not regex.
      // JSX tokens only appear inside JSX state — `/` there is a
      // Punctuator handled by the JSX scanner, not `couldStartRegex`,
      // so this branch is defensive and never hit in practice.
      return false;
  }
}

/**
 * Returns the index of the first token at or after `offset`, using a
 * binary search against `tokens[i].range[0]`. -1 if none.
 */
export function tokenIndexAtOrAfter(
  tokens: readonly Token[],
  offset: number,
): number {
  let lo = 0;
  let hi = tokens.length;
  while (lo < hi) {
    const mid = (lo + hi) >> 1;
    if (tokens[mid].range[0] < offset) lo = mid + 1;
    else hi = mid;
  }
  return lo < tokens.length ? lo : -1;
}

// Re-export for SourceCode consumers
export type { LocPosition };

/* rslint-disable @typescript-eslint/no-unsafe-type-assertion */
/**
 * ESLint-compatible `SourceCode` for plugins running inside the runner
 * Worker; tracks ESLint v10. Split from `context.ts` (which owns `report()`
 * wiring) — this file owns token/comment/scope/spacing/AST querying.
 *
 * Supported surface (members removed in v10 are intentionally absent):
 *
 *   text / ast / lines / hasBOM / scopeManager / visitorKeys / parserServices
 *   getText / getLines / getLocFromIndex / getIndexFromLoc
 *   getTokenBefore / After / getFirstToken / getLastToken / getTokens
 *     / getTokensBetween / getFirstTokens / getLastTokens / getTokensBefore
 *     / getTokensAfter / getTokenByRangeStart
 *     / getFirstTokenBetween / getFirstTokensBetween
 *     / getLastTokenBetween / getLastTokensBetween
 *   getCommentsBefore / After / getCommentsInside / getAllComments
 *   getScope / getDeclaredVariables / markVariableAsUsed / isGlobalReference
 *   getRange / getLoc / getAncestors
 *   getNodeByRangeIndex
 *   commentsExistBetween / isSpaceBetween
 *   getInlineConfigNodes
 *
 * Lazy resources (computed on first access, then cached on the
 * SourceCode instance):
 *
 *   - tokens / comments   token rebuild runs on first token-API call
 *   - scopeManager        scope-factory thunk fires on first scope call
 */

// Use the runner's canonical visitor-key table (vendored from
// oxc-parser 0.133, drift-guarded against ESLint's keys; see
// `visitor-keys.ts`). The bare `eslint-visitor-keys` `KEYS` covers
// only ESTree, missing every TS-specific node type (TSEnumDeclaration,
// TSInterfaceDeclaration, TSAsExpression, TSMappedType, …). Using it
// for `sourceCode.visitorKeys` silently dropped TS subtrees from
// `for (const k of sc.visitorKeys[node.type])` walks, and made
// `getNodeByRangeIndex` stop descending into TS bodies.
import {
  VISITOR_KEYS as RUNNER_VISITOR_KEYS,
  getVisitorKeys,
} from '../ast/visitor-keys.js';
import {
  buildLineStartOffsets,
  offsetToLineColumn,
  type LocPosition,
  type SourceLocation,
} from '../ast/normalize-ast.js';
import {
  getDeclaredVariablesFromScopeManager,
  getScopeForNode,
  markVariableAsUsedInScopeChain,
  normalizeCountOpts,
  normalizeFilterOpts,
  normalizeSkipOpts,
  paddedTokenSlice,
} from './source-code-helpers.js';
import {
  buildTokens,
  tokenIndexAtOrAfter,
  type Comment,
  type Token,
} from './token-builder.js';
import { ConfigCommentParser } from '@eslint/plugin-kit';

// Shared parser instance — `parseDirective` doesn't hold per-call state
// (matches ESLint v10's internal usage in source-code.js).
const directiveParser = new ConfigCommentParser();

// ─────────────────────────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────────────────────────

/** Compact node shape — what plugin code reads through `context.sourceCode`. */
export interface ESTreeNode {
  type: string;
  range: [number, number];
  loc: SourceLocation;
  parent?: ESTreeNode;
  [key: string]: unknown;
}

export interface SourceCode {
  text: string;
  ast: ESTreeNode;
  lines: string[];
  hasBOM: boolean;
  scopeManager: unknown;
  /**
   * ESTree visitor keys map (`{ NodeType: [...childKeys] }`). v10
   * exposes this so rules can do their own traversal without pulling
   * `eslint-visitor-keys`. rslint ships the standard ESTree key set;
   * the runner doesn't yet support custom parsers, so this is static.
   */
  visitorKeys: Record<string, readonly string[]>;
  /**
   * Parser-supplied services. Empty `{}` in plain JS (matches v10).
   * Real ESLint would populate this with the TS parser's `program`,
   * `esTreeNodeToTSNodeMap`, etc.; the runner doesn't proxy ts-go
   * type info through, so plugin rules that need TS types should
   * guard via `if (!services.program) return {}`.
   */
  parserServices: Record<string, unknown>;

  getText(node?: ESTreeNode, beforeCount?: number, afterCount?: number): string;
  getLines(): string[];
  getLocFromIndex(index: number): LocPosition;
  getIndexFromLoc(loc: LocPosition): number;
  getRange(node: ESTreeNode): [number, number];
  getLoc(node: ESTreeNode): SourceLocation;
  getAncestors(node: ESTreeNode): ESTreeNode[];
  getNodeByRangeIndex(index: number): ESTreeNode | null;

  // Tokens — singular accessors. ESLint's API accepts:
  //   number  → `skip` (skip N matches, return the (N+1)-th)
  //   Function → `filter` (per-token predicate; no skip)
  //   Object  → `{ skip?, filter?, includeComments? }`
  // See `TokenSkipOpts` below. Empirically verified against ESLint:
  //   getFirstToken(decl, 0) → first; (decl, 1) → 2nd; (decl, 2) → 3rd, etc.
  getTokenBefore(node: ESTreeNode, opts?: TokenSkipOpts): Token | null;
  getTokenAfter(node: ESTreeNode, opts?: TokenSkipOpts): Token | null;
  getFirstToken(node: ESTreeNode, opts?: TokenSkipOpts): Token | null;
  getLastToken(node: ESTreeNode, opts?: TokenSkipOpts): Token | null;
  // `opts` accepts ESLint's `Function|Object` filter form OR a numeric
  // `beforeCount`; when numeric, `afterCount` (the 3rd positional arg)
  // adds tokens after the node's range too — matching ESLint's
  // `getTokens(node, beforeCount, afterCount)` PaddedTokenCursor path.
  getTokens(
    node: ESTreeNode,
    opts?: TokenFilterOpts,
    afterCount?: number,
  ): Token[];
  // `opts` accepts ESLint's `Function|Object` filter form OR a numeric
  // `padding`; numeric padding adds that many tokens on BOTH sides of
  // the between-range — matching ESLint's `getTokensBetween(left, right,
  // padding)` PaddedTokenCursor path (before === after === padding).
  getTokensBetween(
    left: ESTreeNode,
    right: ESTreeNode,
    opts?: TokenFilterOpts,
  ): Token[];
  // Between two nodes — singular returns the {first,last} matching token
  // in the gap; plural returns up to `count` of them. Mirrors ESLint
  // v10's `SourceCode.getFirstTokenBetween` / `getFirstTokensBetween`
  // / `getLastTokenBetween` / `getLastTokensBetween`.
  getFirstTokenBetween(
    left: ESTreeNode,
    right: ESTreeNode,
    opts?: TokenSkipOpts,
  ): Token | null;
  getFirstTokensBetween(
    left: ESTreeNode,
    right: ESTreeNode,
    opts?: TokenCountOpts,
  ): Token[];
  getLastTokenBetween(
    left: ESTreeNode,
    right: ESTreeNode,
    opts?: TokenSkipOpts,
  ): Token | null;
  getLastTokensBetween(
    left: ESTreeNode,
    right: ESTreeNode,
    opts?: TokenCountOpts,
  ): Token[];
  // The plural getters accept ESLint's full options shape:
  //   number → `count` (cap on the number of tokens returned)
  //   Function → `filter` (per-token predicate)
  //   Object → `{ count, filter, includeComments }`
  // Matches ESLint's `lib/languages/js/source-code/token-store` API so
  // plugin rules using `{ filter, includeComments }` work without a shim.
  getFirstTokens(node: ESTreeNode, opts?: TokenCountOpts): Token[];
  getLastTokens(node: ESTreeNode, opts?: TokenCountOpts): Token[];
  getTokensBefore(node: ESTreeNode, opts?: TokenCountOpts): Token[];
  getTokensAfter(node: ESTreeNode, opts?: TokenCountOpts): Token[];
  getTokenByRangeStart(
    start: number,
    opts?: { includeComments?: boolean },
  ): Token | null;

  // Comments
  getCommentsBefore(node: ESTreeNode): Comment[];
  getCommentsAfter(node: ESTreeNode): Comment[];
  getCommentsInside(node: ESTreeNode): Comment[];
  getAllComments(): Comment[];
  /**
   * All code tokens AND comments merged into one stream sorted by `range[0]`
   * — ESLint's `SourceCode#tokensAndComments`. Stylistic whitespace rules
   * (`comma-spacing`, `no-multi-spaces`, `indent`, `indent-binary-ops`,
   * `space-in-parens`, ...) read this array directly.
   */
  readonly tokensAndComments: readonly Token[];
  // NOTE: `getJSDocComment(node)` was removed in ESLint v10 with no
  // replacement; rslint mirrors that removal. Rules that need adjacent
  // JSDoc can walk `getCommentsBefore(node)` themselves.

  // Spacing
  commentsExistBetween(left: ESTreeNode, right: ESTreeNode): boolean;
  isSpaceBetween(left: ESTreeNode, right: ESTreeNode): boolean;

  // Scope
  getScope(node?: ESTreeNode): unknown;
  getDeclaredVariables(node: ESTreeNode): unknown[];
  markVariableAsUsed(name: string, node?: ESTreeNode): boolean;
  /**
   * ESLint v9 `sourceCode.isGlobalReference(node)`. Returns true iff
   * `node` is an Identifier referencing a variable that lives in the
   * global scope AND has no in-source definition. Widely used by
   * community rules (unicorn, ESLint built-ins). See impl for the
   * exact semantics.
   */
  isGlobalReference(node: ESTreeNode): boolean;

  // Inline config
  getInlineConfigNodes(): Comment[];
  /**
   * ESLint v10's `sourceCode.getDisableDirectives()`. Returns the
   * parsed `eslint-disable*` / `eslint-enable` directives in the file
   * plus any parse problems (e.g. multi-line `eslint-disable-line`).
   * Plugin rules like `unicorn/no-abusive-eslint-disable` consume it
   * to report on directive shape.
   */
  getDisableDirectives(): {
    problems: Array<{
      ruleId: null;
      message: string;
      loc: { start: LocPosition; end: LocPosition };
    }>;
    directives: Array<{
      type: 'disable' | 'enable' | 'disable-next-line' | 'disable-line';
      node: Comment;
      value: string;
      justification: string;
    }>;
  };
}

export type TokenFilterOpts =
  | { filter?: (t: Token) => boolean; includeComments?: boolean }
  | ((t: Token) => boolean)
  // Numeric padding — ESLint's `getTokens(node, beforeCount, afterCount)`
  // and `getTokensBetween(left, right, padding)` accept plain numbers
  // that route through `PaddedTokenCursor`, expanding the returned slice
  // by that many CODE tokens on each side. See `normalizeFilterOpts`.
  | number;

/**
 * Singular-token API options. Mirrors ESLint:
 *   - `number` → `skip` count (return the (skip+1)-th matching token)
 *   - `(t: Token) => boolean` → per-token predicate
 *   - object → `{ skip?, filter?, includeComments? }`
 *
 * The `skip` semantics caught us in the audit: `getFirstToken(node, 2)`
 * does NOT mean "first 2 tokens" — it means "skip the first 2 matches
 * and return the 3rd". Plugin code that uses bare numbers here is
 * almost always doing skip; treating the number as anything else
 * silently picks the wrong token. (Verified against ESLint:
 * `getFirstToken('const x = 1;', 2)` returns `=`.)
 */
export type TokenSkipOpts =
  | number
  | ((t: Token) => boolean)
  | {
      skip?: number;
      filter?: (t: Token) => boolean;
      includeComments?: boolean;
    };

/**
 * Plural-token API options. Mirrors ESLint:
 *   - `number` → maximum number of tokens to return
 *   - `(t: Token) => boolean` → per-token predicate
 *   - object → `{ count?, filter?, includeComments? }`
 *
 * `includeComments: true` interleaves comments with code tokens in
 * source order (the same order a single-pass lexer would emit them).
 * Plugins use this to inspect formatting (e.g. whether a comment
 * appears between two tokens).
 */
export type TokenCountOpts =
  | number
  | ((t: Token) => boolean)
  | {
      count?: number;
      filter?: (t: Token) => boolean;
      includeComments?: boolean;
    };

export interface SourceCodeBuildInput {
  text: string;
  ast: ESTreeNode;
  scopeManagerFactory: () => unknown;
  /**
   * Whether the on-disk source file started with a UTF-8 BOM. The
   * caller is expected to STRIP the BOM byte from `text` before
   * passing it in — both the native parser and downstream offset-aware code
   * see the same byte view that way, and `SourceCode.hasBOM` still
   * reflects the original file faithfully via this flag.
   *
   * When omitted, the factory falls back to detecting the BOM on
   * `text[0]` itself (and strips internally) for legacy callers that
   * passed BOM-containing text directly.
   */
  hasBOM?: boolean;
  /**
   * Optional parser-emitted comments. When provided, `getAllComments`
   * and any comment-only access path builds the ESLint-shape Comment[] from
   * these directly (cheap O(N_comments) shape adapter) instead of triggering
   * the full text tokenizer just to recover the same comments. Tokenize is
   * still invoked lazily if a rule asks for the full token stream.
   *
   * Shape from the native parser: `{ type: 'Line' | 'Block', value, start, end }`.
   * oxc does not emit shebang comments — we synthesize a Shebang entry
   * here when `text` starts with `#!`, matching ESLint's getAllComments
   * contract on shebang files.
   */
  parsedComments?: ReadonlyArray<{
    type: 'Line' | 'Block';
    value: string;
    start: number;
    end: number;
  }>;
  /**
   * Native-parser token stream in columnar form (the napi parser's `parse()`:
   * `tokenTypes`/`tokenStarts`/`tokenEnds`, all UTF-16 offsets). The token
   * getters lazily rebuild `Token[]` from these via {@link token-builder.buildTokens} on
   * first use. When omitted, the token stream is empty (the main lint path always supplies
   * it; a SourceCode built without it simply has no tokens, matching "no token APIs used").
   */
  parsedTokens?: {
    types: Uint8Array;
    starts: Uint32Array;
    ends: Uint32Array;
  };
}

// ─────────────────────────────────────────────────────────────────────
// Internal helpers
// ─────────────────────────────────────────────────────────────────────

/**
 * ECMAScript line-terminator regex. Spec (§ 11.3 LineTerminator):
 * U+000A LF, U+000D CR, U+2028 LINE SEPARATOR, U+2029 PARAGRAPH SEPARATOR.
 * A `
` pair counts as ONE terminator. Exported as a module-level
 * constant so the `lines` getter / split path stays in lockstep with
 * `buildLineStartOffsets` (the offset-computing side of the same
 * terminator definition).
 */
export const LINE_TERMINATOR_RE = /\r\n|[\r\n\u2028\u2029]/g;

/**
 * Build ESLint-shape `Comment[]` from the native parser's compact comment list.
 *
 * oxc emits `{ type: 'Line' | 'Block', value, start, end }`. ESLint
 * rules expect `{ type, value, range: [start, end], loc: {start, end} }`
 * where loc positions are 1-indexed line + 0-indexed column. Building
 * those four extra fields per comment is O(N_comments) and skips the
 * O(text_length) full text scan the a full-text scan would have
 * paid for the same comment list.
 *
 * Shebang handling: oxc does not surface `#!` lines as comments
 * (they're file-preamble metadata in its grammar), but ESLint's
 * `getAllComments` contract treats shebang as a leading `Shebang`-type
 * comment so rules that inspect the entire comment stream see it. We
 * detect the shebang by literal `#!` prefix on the text and synthesize
 * one entry, matching ESLint's behaviour bit-for-bit.
 */
function buildCommentsFromParsed(
  parsedComments: ReadonlyArray<{
    type: 'Line' | 'Block';
    value: string;
    start: number;
    end: number;
  }>,
  text: string,
  lso: number[],
): Comment[] {
  const out: Comment[] = [];
  // ESLint's `getAllComments` reports a leading `#!` as `type: 'Shebang'`.
  // The native parser (oxc 0.133) puts the shebang solely on `program.hashbang`
  // for ALL languages and never emits it as a comment, so this relabel is a
  // defensive no-op on the current parser — the synthesis block further down is
  // the path that actually produces the Shebang entry. The relabel is kept so a
  // parser that DID emit a `#!` Line at offset 0 would still be reported correctly.
  const hasShebang =
    text.charCodeAt(0) === 0x23 /* # */ && text.charCodeAt(1) === 0x21; /* ! */
  for (let i = 0; i < parsedComments.length; i++) {
    const c = parsedComments[i];
    const isShebangEntry = hasShebang && i === 0 && c.start === 0;
    out.push({
      type: isShebangEntry ? 'Shebang' : c.type,
      value: c.value,
      range: [c.start, c.end],
      loc: {
        start: offsetToLineColumn(c.start, lso),
        end: offsetToLineColumn(c.end, lso),
      },
    });
  }
  // Synthesize a `Shebang` entry from the source text when the file starts with
  // `#!` but no offset-0 Shebang was produced above. With oxc 0.133 this is the
  // path taken for every language (the shebang lives on `program.hashbang`, not in
  // `comments`). ESLint v10's `getAllComments()` always reports a leading `#!` as a
  // `Shebang` token, so we match its shape (value strips `#!`, range spans to the
  // first line terminator — same shape espree produces).
  if (hasShebang && !(out.length > 0 && out[0].type === 'Shebang')) {
    let end = text.length;
    for (let k = 2; k < text.length; k++) {
      const cc = text.charCodeAt(k);
      // ECMAScript LineTerminator: LF, CR, LS (U+2028), PS (U+2029).
      if (cc === 0x0a || cc === 0x0d || cc === 0x2028 || cc === 0x2029) {
        end = k;
        break;
      }
    }
    out.unshift({
      type: 'Shebang',
      value: text.slice(2, end),
      range: [0, end],
      loc: {
        start: offsetToLineColumn(0, lso),
        end: offsetToLineColumn(end, lso),
      },
    });
  }
  return out;
}

/**
 * Re-export of `normalize-ast.ts`'s `buildLineStartOffsets` for
 * callers that already import from this module. Both AST nodes (via
 * `normalizeAst`) and the SourceCode API surface (via `getLso()`
 * below) compute line/column with this one implementation —
 * guaranteeing `node.loc.line === sourceCode.getLocFromIndex(node.range[0]).line`
 * for every node, regardless of which ECMAScript line terminator
 * family the file uses. Pre-fix the two surfaces ran separate copies
 * that disagreed on bare CR / LS / PS files.
 */
export { buildLineStartOffsets as buildLineStartOffsetsLocal };

// ─────────────────────────────────────────────────────────────────────
// Factory
// ─────────────────────────────────────────────────────────────────────

/**
 * Construct a SourceCode. All non-trivial computations are lazy.
 */
export function createSourceCode(input: SourceCodeBuildInput): SourceCode {
  let text = input.text;
  const ast = input.ast;
  // Detect BOM either via the explicit flag from the caller (ecma-
  // language-plugin strips before parsing and threads the flag) OR
  // via a leading 0xFEFF on the text itself (legacy / direct
  // callers). When the BOM is still present we strip it here so
  // `SourceCode.text` matches ESLint's contract (no leading BOM).
  const hasBOM = input.hasBOM === true || text.charCodeAt(0) === 0xfeff;
  if (text.charCodeAt(0) === 0xfeff) text = text.slice(1);

  // Lazy state — instantiated on first access of the relevant API family.
  let _lso: number[] | null = null;
  let _tokens: Token[] | null = null;
  let _comments: Comment[] | null = null;
  let _scope: unknown = null;
  let _scopeInit = false;
  // Caches a thrown scope-factory error. Pre-fix the `_scopeInit = true`
  // flip happened BEFORE the factory call, so a throwing factory left
  // `_scope = null` + `_scopeInit = true` permanently — every subsequent
  // `getScopeManager()` silently returned `null`, and scope-dependent
  // rules cascaded into `null` scopes with NO error recorded against
  // them. Caching the original error and re-throwing on every access
  // ensures each downstream consumer records its own `ruleErrors` entry.
  let _scopeError: Error | null = null;
  let _disableDirectives: ReturnType<
    SourceCode['getDisableDirectives']
  > | null = null;
  // Cache for the merged tokens+comments stream produced by
  // `streamFor(true)`. SourceCode is single-file/single-threaded; once
  // the stream is built it's immutable for the rest of the file's
  // lifetime, so a single allocation suffices regardless of how many
  // rules call comment-aware token APIs (`getTokenBefore`, etc.).
  let _mergedStream: readonly Token[] | null = null;
  // Cache for `text.split(LINE_TERMINATOR_RE)`. Matches ESLint v10
  // (`SourceCode#lines` is computed once in the constructor and shared).
  let _lines: string[] | null = null;

  const getLso = (): number[] => {
    if (_lso == null) {
      _lso = buildLineStartOffsets(text);
    }
    return _lso;
  };

  const ensureTokens = (): { tokens: Token[]; comments: Comment[] } => {
    if (_tokens == null) {
      // Lazily rebuild Token[] from the native parser's columnar token stream;
      // value is sliced from the original `text`, loc is line/column from `getLso()`. A
      // SourceCode built without parser tokens simply has an empty token stream.
      const pt = input.parsedTokens;
      _tokens = pt
        ? buildTokens(pt.types, pt.starts, pt.ends, text, getLso())
        : [];
    }
    return { tokens: _tokens, comments: ensureComments() };
  };

  /**
   * Comments come from the native parser's compact comment list (`parsedComments`), built
   * into ESLint-shape `Comment[]` on first access (`getAllComments`, `applyDisableDirectives`,
   * rules like `ban-ts-comment`). When absent, the SourceCode simply has no comments.
   */
  const ensureComments = (): Comment[] => {
    if (_comments == null) {
      _comments = input.parsedComments
        ? buildCommentsFromParsed(input.parsedComments, text, getLso())
        : [];
    }
    return _comments ?? [];
  };

  /**
   * Build a tokens-OR-tokens+comments stream sorted by `range[0]`.
   *
   * tokens and comments are each already sorted (parser-emit order),
   * and they don't overlap (a comment never has the same span as a
   * code token), so a linear merge is enough — no full sort needed.
   * Falls back to the tokens array as-is when comments aren't wanted,
   * avoiding the merge cost on the hot path.
   */
  const streamFor = (include: boolean): readonly Token[] => {
    if (!include) return ensureTokens().tokens;
    if (_mergedStream != null) return _mergedStream;
    const { tokens, comments } = ensureTokens();
    if (comments.length === 0) {
      _mergedStream = tokens;
      return tokens;
    }
    const merged: Token[] = [];
    let i = 0,
      j = 0;
    while (i < tokens.length && j < comments.length) {
      if (tokens[i].range[0] <= comments[j].range[0]) merged.push(tokens[i++]);
      // Comments are shape-compatible with Token (type, value, range, loc)
      // — ESLint exposes them via the same SourceCode#getTokens API when
      // `includeComments` is on. Cast keeps the TS signature simple.
      else merged.push(comments[j++] as unknown as Token);
    }
    while (i < tokens.length) merged.push(tokens[i++]);
    while (j < comments.length) merged.push(comments[j++] as unknown as Token);
    _mergedStream = merged;
    return merged;
  };

  // Scope — lazy
  const getScopeManager = (): unknown => {
    if (_scopeError != null) throw _scopeError;
    if (!_scopeInit) {
      _scopeInit = true;
      try {
        _scope = input.scopeManagerFactory();
      } catch (err) {
        _scopeError = err instanceof Error ? err : new Error(String(err));
        throw _scopeError;
      }
    }
    return _scope;
  };

  // Scope helpers (getScopeForNode / getDeclaredVariablesFromScopeManager
  // / markVariableAsUsedInScopeChain) live in `source-code-helpers.ts`
  // as pure functions. The SourceCode methods below pass the lazy
  // `getScopeManager()` result into them. Separating these keeps the
  // factory under ~870 LoC and the scope semantics independently
  // testable.

  // ── public methods (object literal, lazy-bound) ──
  const sc: SourceCode = {
    text,
    ast,
    get lines(): string[] {
      // ESLint splits on ALL ECMAScript line terminators (LF, CR, CRLF,
      // U+2028, U+2029) — see `LINE_TERMINATOR_RE`. Splitting on `\n`
      // only would miss legacy CR-terminated files and the Unicode
      // separators, leaving the array shorter than the file actually has.
      // Cached: ESLint v10 computes this once per SourceCode; rules
      // read `sourceCode.lines` many times across line/column work.
      if (_lines == null) _lines = text.split(LINE_TERMINATOR_RE);
      return _lines;
    },
    hasBOM,
    get scopeManager(): unknown {
      return getScopeManager();
    },
    // ESLint's `SourceCode#tokensAndComments`: the code-token stream merged
    // with comments, sorted by `range[0]`. Reuses the lazy/cached `streamFor`
    // merge (same array `getTokens(..., {includeComments:true})` uses).
    get tokensAndComments(): readonly Token[] {
      return streamFor(true);
    },
    // Full ESTree + TypeScript visitor-keys map (vendored from
    // oxc-parser 0.133, drift-guarded against ESLint's keys; see
    // `visitor-keys.ts`). The runner needs the TS-aware set because
    // plugin rules iterate `sourceCode.visitorKeys[node.type]` on .ts
    // files and expect to recurse into TSEnumDeclaration /
    // TSInterfaceDeclaration / TSAsExpression children. The plain
    // ESTree KEYS would return `undefined` for those types, silently
    // skipping every TS subtree.
    visitorKeys: RUNNER_VISITOR_KEYS as Record<string, readonly string[]>,
    // Empty in plain JS (matches ESLint v10's default). Type-aware
    // plugins that probe `parserServices.program` correctly skip
    // type-aware checks under rslint.
    parserServices: {},

    getText(node?: ESTreeNode, beforeCount = 0, afterCount = 0): string {
      if (!node) return text;
      const start = Math.max(0, node.range[0] - beforeCount);
      const end = Math.min(text.length, node.range[1] + afterCount);
      return text.slice(start, end);
    },
    getLines(): string[] {
      if (_lines == null) _lines = text.split(LINE_TERMINATOR_RE);
      return _lines;
    },
    getLocFromIndex(i: number): LocPosition {
      // Loud validation, matching ESLint v10 `SourceCode#getLocFromIndex`
      // (and the sibling `getIndexFromLoc` below): TypeError on a
      // non-number index, RangeError when out of `[0, text.length]`.
      // Without it a stale / NaN / out-of-bounds index from a buggy
      // rule produced a plausible-looking-but-wrong `{line, column}`
      // and a garbage diagnostic position with no signal.
      if (typeof i !== 'number') {
        throw new TypeError('Expected `index` to be a number.');
      }
      if (i < 0 || i > text.length) {
        throw new RangeError(
          `Index out of range (requested index ${i}, but source text has length ${text.length}).`,
        );
      }
      return offsetToLineColumn(i, getLso());
    },
    getIndexFromLoc(loc: LocPosition): number {
      // ESLint v10 `SourceCode#getIndexFromLoc` validates loc shape AND
      // bounds, throwing TypeError on a malformed argument and
      // RangeError on out-of-range positions. Earlier the runner
      // clamped silently (missing column → 0, line>EOF → text.length,
      // negative column → 0), which masked plugin bugs — a rule that
      // accidentally computed a stale or NaN position would get a
      // plausible-looking offset back and emit a garbage diagnostic
      // with no signal. Loud failure matches v10 and lets rules
      // surface their bugs immediately.
      if (loc == null || typeof loc !== 'object') {
        throw new TypeError('Expected `loc` to be an object.');
      }
      if (typeof loc.line !== 'number') {
        throw new TypeError('Expected `loc.line` to be a number.');
      }
      if (typeof loc.column !== 'number') {
        throw new TypeError('Expected `loc.column` to be a number.');
      }
      const lso = getLso();
      if (loc.line < 1 || loc.line > lso.length) {
        throw new RangeError(
          `Line number out of range (line ${loc.line} requested). Total lines: ${lso.length}`,
        );
      }
      if (loc.column < 0) {
        throw new RangeError(
          `Column number out of range (column ${loc.column} requested).`,
        );
      }
      const lineStart = lso[loc.line - 1];
      // ESLint v10 bounds distinguish the FINAL line from earlier lines.
      // A column may point AT the one-past-last character of the final
      // line (an insertion point at EOF). On any EARLIER line the line
      // terminator is NOT an addressable column — the max valid column
      // lands just before it. `lso[loc.line]` is the start of the next
      // line (one PAST this line's terminator), so the final line uses
      // `>` and every earlier line uses `>=` against it. Pre-fix this
      // used `>` for all lines, so a column pointing at a non-final
      // line's `\n` was wrongly accepted (returned the terminator's
      // offset) where ESLint throws RangeError.
      const isLastLine = loc.line === lso.length;
      const lineEnd = isLastLine ? text.length : lso[loc.line];
      const positionIndex = lineStart + loc.column;
      if (isLastLine ? positionIndex > lineEnd : positionIndex >= lineEnd) {
        throw new RangeError(
          `Column number out of range (column ${loc.column} requested for line ${loc.line}).`,
        );
      }
      return positionIndex;
    },
    getRange(node: ESTreeNode): [number, number] {
      return node.range;
    },
    getLoc(node: ESTreeNode): SourceLocation {
      return node.loc;
    },
    getAncestors(node: ESTreeNode): ESTreeNode[] {
      const out: ESTreeNode[] = [];
      let p = node.parent;
      while (p) {
        out.unshift(p);
        p = p.parent;
      }
      return out;
    },
    getNodeByRangeIndex(index: number): ESTreeNode | null {
      let best: ESTreeNode | null = null;
      const stack: ESTreeNode[] = [ast];
      while (stack.length > 0) {
        const cur = stack.pop()!;
        if (cur.range[0] <= index && index < cur.range[1]) {
          best = cur;
          // Mirror the main visitor's child enumeration (see
          // `listener-merge.ts:520`): consult `visitorKeys[type]` so the
          // recursion set matches ESLint / espree / typescript-eslint
          // exactly. A prior `Object.keys + parent/loc/range blacklist`
          // recursed into non-AST properties and allocated a fresh keys
          // array per node. Uses the TS-aware key table so TS bodies
          // (TSEnumDeclaration members, TSInterfaceDeclaration body,
          // etc.) get visited — without this, `getNodeByRangeIndex`
          // on an offset inside a TS construct stopped at the wrapper
          // node instead of descending to the actual child.
          // Fall back to `getVisitorKeys` for types not in the static
          // table — same convention as `listener-merge.ts:644` and
          // `normalize-ast.ts:300`. The static table covers the common
          // case path; the helper handles parser-specific node types
          // that may not be enumerated upfront (oxc-specific extensions,
          // future TS additions). Pre-fix the bare `?? []` made the
          // walk stop at unknown nodes, mismatching the main visitor.
          const keys =
            RUNNER_VISITOR_KEYS[cur.type] ??
            getVisitorKeys(cur as { type?: string });
          for (let i = 0; i < keys.length; i++) {
            const v = cur[keys[i]];
            if (Array.isArray(v)) {
              for (const c of v as unknown[]) {
                if (c && typeof c === 'object' && (c as ESTreeNode).type)
                  stack.push(c as ESTreeNode);
              }
            } else if (v && typeof v === 'object' && (v as ESTreeNode).type) {
              stack.push(v as ESTreeNode);
            }
          }
        }
      }
      return best;
    },

    // ── tokens ──
    // Singular getters honor ESLint's full options shape:
    //   number → `skip` (skip N matches, return next match)
    //   Function → filter (no skip)
    //   Object → `{ skip?, filter?, includeComments? }`
    // `includeComments: true` injects block / line comments into the
    // stream so they appear as candidates in the same scan.
    getTokenBefore(node, opts) {
      const { skip, filter, includeComments } = normalizeSkipOpts(opts);
      const stream = streamFor(includeComments);
      const targetStart = node.range[0];
      let skipped = 0;
      for (let i = stream.length - 1; i >= 0; i--) {
        const t = stream[i];
        if (t.range[1] > targetStart) continue;
        if (filter && !filter(t)) continue;
        if (skipped < skip) {
          skipped++;
          continue;
        }
        return t;
      }
      return null;
    },
    getTokenAfter(node, opts) {
      const { skip, filter, includeComments } = normalizeSkipOpts(opts);
      const stream = streamFor(includeComments);
      const targetEnd = node.range[1];
      let skipped = 0;
      for (let i = 0; i < stream.length; i++) {
        const t = stream[i];
        if (t.range[0] < targetEnd) continue;
        if (filter && !filter(t)) continue;
        if (skipped < skip) {
          skipped++;
          continue;
        }
        return t;
      }
      return null;
    },
    getFirstToken(node, opts) {
      const { skip, filter, includeComments } = normalizeSkipOpts(opts);
      const stream = streamFor(includeComments);
      let skipped = 0;
      for (let i = 0; i < stream.length; i++) {
        const t = stream[i];
        if (t.range[0] < node.range[0]) continue;
        if (t.range[1] > node.range[1]) break;
        if (filter && !filter(t)) continue;
        if (skipped < skip) {
          skipped++;
          continue;
        }
        return t;
      }
      return null;
    },
    getLastToken(node, opts) {
      const { skip, filter, includeComments } = normalizeSkipOpts(opts);
      const stream = streamFor(includeComments);
      let skipped = 0;
      for (let i = stream.length - 1; i >= 0; i--) {
        const t = stream[i];
        if (t.range[1] > node.range[1]) continue;
        if (t.range[0] < node.range[0]) break;
        if (filter && !filter(t)) continue;
        if (skipped < skip) {
          skipped++;
          continue;
        }
        return t;
      }
      return null;
    },
    getTokens(node, opts, afterCount) {
      // ESLint v10's `sourceCode.getTokens(node, { includeComments })`
      // honors `includeComments` to splice comment tokens into the
      // returned stream. The previous implementation used `filterFn`
      // (which only extracts `.filter`) and walked the code-only
      // token array, silently dropping the option — `@stylistic` /
      // `eslint-plugin-unicorn` rules that pass `includeComments:
      // true` would miss comment tokens. `normalizeFilterOpts` +
      // `streamFor(includeComments)` matches the contract.
      const { filter, includeComments, padding } = normalizeFilterOpts(opts);
      // Numeric form `getTokens(node, beforeCount, afterCount)`: ESLint
      // routes this through a `PaddedTokenCursor` over the CODE tokens —
      // `beforeCount` tokens before the node range + `afterCount` after,
      // clamped to the token array. `padding` here is `beforeCount`; the
      // 3rd positional arg is `afterCount` (default 0, same as ESLint's
      // `afterCount | 0`). No filter / comments on this overload.
      if (padding != null) {
        return paddedTokenSlice(
          ensureTokens().tokens,
          node.range[0],
          node.range[1],
          padding,
          Math.max(0, Math.trunc(afterCount ?? 0)),
        );
      }
      const stream = streamFor(includeComments);
      const out: Token[] = [];
      for (const t of stream) {
        if (t.range[0] >= node.range[0] && t.range[1] <= node.range[1]) {
          if (!filter || filter(t)) out.push(t);
        }
      }
      return out;
    },
    getTokensBetween(left, right, opts) {
      const { filter, includeComments, padding } = normalizeFilterOpts(opts);
      // Numeric form `getTokensBetween(left, right, padding)`: ESLint's
      // `PaddedTokenCursor` uses `padding` for BOTH before and after,
      // expanding the between-range slice by that many CODE tokens on
      // each side. (before === after === padding.)
      if (padding != null) {
        return paddedTokenSlice(
          ensureTokens().tokens,
          left.range[1],
          right.range[0],
          padding,
          padding,
        );
      }
      const stream = streamFor(includeComments);
      const out: Token[] = [];
      for (const t of stream) {
        if (t.range[0] >= left.range[1] && t.range[1] <= right.range[0]) {
          if (!filter || filter(t)) out.push(t);
        }
      }
      return out;
    },
    // ESLint v10 between-pair singular/plural getters. Implementations
    // delegate to the existing stream + normalize helpers used by
    // getFirstToken / getFirstTokens so skip / filter / includeComments
    // semantics match exactly (ESLint's token-store uses the same
    // option shapes — empirically pinned).
    getFirstTokenBetween(left, right, opts) {
      const { skip, filter, includeComments } = normalizeSkipOpts(opts);
      const stream = streamFor(includeComments);
      let matched = 0;
      for (const t of stream) {
        if (t.range[0] < left.range[1]) continue;
        if (t.range[1] > right.range[0]) break;
        if (filter && !filter(t)) continue;
        if (matched++ < skip) continue;
        return t;
      }
      return null;
    },
    getFirstTokensBetween(left, right, opts) {
      const { count, filter, includeComments } = normalizeCountOpts(opts);
      const stream = streamFor(includeComments);
      const out: Token[] = [];
      for (const t of stream) {
        if (t.range[0] < left.range[1]) continue;
        if (t.range[1] > right.range[0]) break;
        if (filter && !filter(t)) continue;
        // Check cap BEFORE push so count:0 returns [] (not [t]). count
        // is Infinity when caller omits the cap, so omitted-opts paths
        // never short-circuit here.
        if (out.length >= count) break;
        out.push(t);
      }
      return out;
    },
    getLastTokenBetween(left, right, opts) {
      const { skip, filter, includeComments } = normalizeSkipOpts(opts);
      const stream = streamFor(includeComments);
      // Collect all candidates first, then walk from the end.
      const matches: Token[] = [];
      for (const t of stream) {
        if (t.range[0] < left.range[1]) continue;
        if (t.range[1] > right.range[0]) break;
        if (filter && !filter(t)) continue;
        matches.push(t);
      }
      const idx = matches.length - 1 - skip;
      return idx >= 0 ? matches[idx] : null;
    },
    getLastTokensBetween(left, right, opts) {
      const { count, filter, includeComments } = normalizeCountOpts(opts);
      const stream = streamFor(includeComments);
      const matches: Token[] = [];
      for (const t of stream) {
        if (t.range[0] < left.range[1]) continue;
        if (t.range[1] > right.range[0]) break;
        if (filter && !filter(t)) continue;
        matches.push(t);
      }
      if (count >= matches.length) return matches;
      return matches.slice(matches.length - count);
    },
    getFirstTokens(node, opts) {
      const { count, filter, includeComments } = normalizeCountOpts(opts);
      const stream = streamFor(includeComments);
      const out: Token[] = [];
      for (const t of stream) {
        if (t.range[0] < node.range[0]) continue;
        if (t.range[1] > node.range[1]) break;
        if (filter && !filter(t)) continue;
        // Check cap BEFORE push so count:0 returns [] (not [t]). count
        // is Infinity when caller omits the cap, so omitted-opts paths
        // never short-circuit here.
        if (out.length >= count) break;
        out.push(t);
      }
      return out;
    },
    getLastTokens(node, opts) {
      const { count, filter, includeComments } = normalizeCountOpts(opts);
      const stream = streamFor(includeComments);
      const out: Token[] = [];
      for (let i = stream.length - 1; i >= 0; i--) {
        const t = stream[i];
        if (t.range[1] > node.range[1]) continue;
        if (t.range[0] < node.range[0]) break;
        if (filter && !filter(t)) continue;
        // Building the result back-to-front so we can short-circuit on
        // the count cap; reverse once at the end. Check BEFORE push so
        // count:0 returns []; Infinity (omitted opts) never trips.
        if (out.length >= count) break;
        out.push(t);
      }
      return out.reverse();
    },
    getTokensBefore(node, opts) {
      const { count, filter, includeComments } = normalizeCountOpts(opts);
      const stream = streamFor(includeComments);
      const out: Token[] = [];
      for (let i = stream.length - 1; i >= 0; i--) {
        const t = stream[i];
        if (t.range[1] > node.range[0]) continue;
        if (filter && !filter(t)) continue;
        // Check cap BEFORE push so count:0 returns [] (not [t]). count
        // is Infinity when caller omits the cap, so omitted-opts paths
        // never short-circuit here.
        if (out.length >= count) break;
        out.push(t);
      }
      return out.reverse();
    },
    getTokensAfter(node, opts) {
      const { count, filter, includeComments } = normalizeCountOpts(opts);
      const stream = streamFor(includeComments);
      const out: Token[] = [];
      for (const t of stream) {
        if (t.range[0] < node.range[1]) continue;
        if (filter && !filter(t)) continue;
        // Check cap BEFORE push so count:0 returns [] (not [t]). count
        // is Infinity when caller omits the cap, so omitted-opts paths
        // never short-circuit here.
        if (out.length >= count) break;
        out.push(t);
      }
      return out;
    },
    getTokenByRangeStart(start, opts) {
      // ESLint v10 honors `{ includeComments: true }` so callers can
      // find a comment that begins at a specific offset (e.g. a
      // JSDoc-bridging rule pinning a comment to a node's range
      // start). Pre-fix the runner ignored the opts entirely and only
      // searched the code-token stream, returning null for any
      // comment-starting offset.
      const stream = streamFor(opts?.includeComments === true);
      // Stream is sorted by range[0]; binary search to the candidate
      // index, then verify exact match (range[0] could equal `start - 1`
      // if `start` lands in a gap).
      const idx = tokenIndexAtOrAfter(stream, start);
      if (idx < 0) return null;
      const t = stream[idx];
      return t.range[0] === start ? t : null;
    },

    // ── comments ──
    //
    // ESLint semantics for getCommentsBefore/After return ONLY the
    // comments that sit between the node and the nearest neighbouring
    // CODE token — not every comment that appears earlier / later in
    // the file. We previously filtered solely by offset (`c.range[1]
    // <= node.range[0]`), which accidentally returned every preceding
    // comment in the file. Verified against ESLint's behavior on a
    // multi-statement fixture: rslint had returned 1+ extra comments.
    //
    // The fix walks the code-token stream once to find the
    // immediately-prior/-next CODE token and then bounds the comment
    // filter by the gap between that token and the node.
    getCommentsBefore(node) {
      const { tokens, comments } = ensureTokens();
      // tokens are sorted by range[0]. The nearest code token *before*
      // `node` sits at index `tokenIndexAtOrAfter(node.range[0]) - 1`
      // (or the last token if no token starts at/after node).
      let prevTokenEnd = 0;
      const idxAfter = tokenIndexAtOrAfter(tokens, node.range[0]);
      const prevIdx = idxAfter < 0 ? tokens.length - 1 : idxAfter - 1;
      if (prevIdx >= 0) prevTokenEnd = tokens[prevIdx].range[1];
      return comments.filter(
        (c) => c.range[0] >= prevTokenEnd && c.range[1] <= node.range[0],
      );
    },
    getCommentsAfter(node) {
      const { tokens, comments } = ensureTokens();
      // First code token whose start >= node.range[1] — binary search.
      const idx = tokenIndexAtOrAfter(tokens, node.range[1]);
      const nextTokenStart = idx < 0 ? text.length : tokens[idx].range[0];
      return comments.filter(
        (c) => c.range[0] >= node.range[1] && c.range[1] <= nextTokenStart,
      );
    },
    getCommentsInside(node) {
      const { comments } = ensureTokens();
      return comments.filter(
        (c) => c.range[0] >= node.range[0] && c.range[1] <= node.range[1],
      );
    },
    getAllComments() {
      // Cheap path when `parsedComments` was provided at construction —
      // skips full text tokenize. See `ensureComments` for the fallback.
      return ensureComments().slice();
    },

    // ── spacing ──
    commentsExistBetween(left, right) {
      const { comments } = ensureTokens();
      // comments are sorted by range[0]. Binary search to the first
      // candidate, then O(1) check that it ends before `right`.
      const idx = tokenIndexAtOrAfter(
        comments as unknown as readonly Token[],
        left.range[1],
      );
      if (idx < 0) return false;
      return comments[idx].range[1] <= right.range[0];
    },
    isSpaceBetween(left, right) {
      // ESLint semantics: returns true iff there is ACTUAL WHITESPACE
      // between the two nodes/tokens. Comments do NOT count as space —
      // a node touching a comment touching another node is "no space".
      //
      // Algorithm (mirrors ESLint's `SourceCode#isSpaceBetween` in
      // `lib/languages/js/source-code/source-code.js`):
      //
      //   1. Overlapping spans → false (nested, can't have space)
      //   2. Walk the tokens-AND-comments stream from the last code
      //      token of `left` to the first code token of `right`,
      //      stepping one element at a time. A gap between any two
      //      adjacent stream entries (`prev.range[1] !== next.range[0]`)
      //      proves there's whitespace in the source. Otherwise the
      //      span is solidly filled with tokens/comments — no space.
      //
      // The naive `right.range[0] > left.range[1]` from the previous
      // implementation flagged `a/*x*/b` as "space" because it only
      // checked range non-adjacency; ESLint returns false there since
      // the comment sits flush against both sides. Real rules
      // (spacing, formatting plugins) rely on the ESLint behavior.
      if (
        !(left.range[1] <= right.range[0] || right.range[1] <= left.range[0])
      ) {
        return false;
      }
      const [startNode, endNode] =
        left.range[1] <= right.range[0] ? [left, right] : [right, left];
      const startAnchor =
        sc.getLastToken(startNode) ?? (startNode as unknown as Token);
      const endAnchor =
        sc.getFirstToken(endNode) ?? (endNode as unknown as Token);
      const stream = streamFor(true);
      // Locate startAnchor by reference. `===` works because streamFor
      // reuses the live token/comment objects without cloning.
      let i = stream.indexOf(startAnchor);
      if (i < 0) {
        // Fallback: if the anchor isn't a tokenizer-emitted object
        // (e.g. the caller passed a raw Token literal we don't know
        // about), binary-search to the last token whose start < the
        // anchor's end — i.e. first index where range[0] >= anchor.end,
        // minus one.
        const after = tokenIndexAtOrAfter(stream, startAnchor.range[1]);
        i = after < 0 ? stream.length - 1 : after - 1;
        if (i < 0) i = 0;
      }
      while (i + 1 < stream.length) {
        const cur = stream[i];
        const next = stream[i + 1];
        if (cur === endAnchor) break;
        if (cur.range[1] !== next.range[0]) return true;
        i++;
        if (next === endAnchor) break;
      }
      return false;
    },

    // ── scope ──
    getScope(node?: ESTreeNode) {
      // ESLint v10 throws on no-arg `sourceCode.getScope()` — see
      // `lib/languages/js/source-code/source-code.js`. Pre-fix the
      // runner silently returned globalScope, which would mask plugin
      // bugs (the rule meant to pass a node and got the wrong scope
      // back with no signal). Loudly surfacing the misuse keeps
      // plugin authors honest and matches the v10 contract.
      if (node == null) {
        throw new TypeError('Missing required argument: node.');
      }
      return getScopeForNode(getScopeManager(), node);
    },
    getDeclaredVariables(node: ESTreeNode) {
      return getDeclaredVariablesFromScopeManager(getScopeManager(), node);
    },
    markVariableAsUsed(name: string, node?: ESTreeNode) {
      return markVariableAsUsedInScopeChain(getScopeManager(), ast, name, node);
    },
    isGlobalReference(node: ESTreeNode): boolean {
      // Mirrors ESLint's `lib/languages/js/source-code/source-code.js`
      // `isGlobalReference`: returns true iff `node` is an Identifier
      // that resolves to a variable in the program's global scope
      // whose name has no in-source definition (i.e. a "leaked-in"
      // global from `languageOptions.globals` or the host).
      //
      // Critical for community rules. Confirmed callers in widely
      // installed plugins (sampled from the test-tools node_modules):
      //   - eslint-plugin-unicorn:
      //       no-typeof-undefined, no-useless-error-capture-stack-trace,
      //       prefer-module, prefer-prototype-methods, error-message
      //   - eslint (core): no-setter-return, no-implied-eval,
      //       prefer-regex-literals, no-promise-executor-return
      // Without this method a rule calling
      // `sourceCode.isGlobalReference(node)` crashes with
      // `TypeError: ... is not a function`. The method previously
      // wasn't on the SourceCode surface at all.
      if (node == null) {
        throw new TypeError('sourceCode.isGlobalReference: node is required');
      }
      if ((node as { type?: string }).type !== 'Identifier') return false;
      const sm = getScopeManager() as {
        scopes?: Array<{
          set?: Map<
            string,
            { defs: unknown[]; references: Array<{ identifier: unknown }> }
          >;
        }>;
      } | null;
      const globalScope = sm?.scopes?.[0];
      if (!globalScope?.set) return false;
      const name = (node as { name?: string }).name;
      if (typeof name !== 'string') return false;
      const variable = globalScope.set.get(name);
      // ESLint: a variable with any in-source definition is NOT a
      // "global reference" — it's a locally-shadowing binding. Only
      // names that resolve in the global scope with zero defs (i.e.
      // injected via `languageOptions.globals` or host environment)
      // count.
      if (!variable || variable.defs.length > 0) return false;
      // ESLint walks `variable.references` to confirm `node` is among
      // them. Skipping this check would yield true for every
      // Identifier matching a global name, even when the Identifier
      // is, e.g., a property key in `{ foo: 1 }` (those aren't refs).
      return variable.references.some((r) => r.identifier === node);
    },

    // ── inline config ──
    getInlineConfigNodes() {
      // ESLint v10's `sourceCode.getInlineConfigNodes()` returns *any*
      // comment whose trimmed value is a recognized inline directive:
      //
      //   - `eslint-disable[-next-line]?` / `eslint-enable`
      //   - `eslint <rule>: <severity>` — inline rule-config
      //   - `global <name>[:rw]` / `globals <list>` — declared globals
      //   - `exported <name>` — module-internal export-tracking helper
      //
      // The pre-fix shim recognized only the disable/enable family and
      // silently dropped the rest; lint-config-walker tooling that uses
      // `getInlineConfigNodes` to surface ALL directives (so users can
      // see "global foo" annotations alongside disables) only saw a
      // subset. Empirically pinned against ESLint v10.3.
      const { comments } = ensureTokens();
      // ESLint v10's `directivesPattern` requires the label to be
      // followed by whitespace or the end of the directive — bare
      // `startsWith` lets `eslint-disable-foo` look like a real
      // `eslint-disable` directive (no word boundary). Match the same
      // boundary contract here so external linter / config-aware tools
      // calling `getInlineConfigNodes()` don't surface bogus entries.
      const LABELS = [
        'eslint-disable-next-line',
        'eslint-disable-line',
        'eslint-disable',
        'eslint-enable',
        'rslint-disable-next-line',
        'rslint-disable-line',
        'rslint-disable',
        'rslint-enable',
        'eslint',
        // ESLint v10 surfaces `eslint-env` via `getInlineConfigNodes()`
        // (its `directivesPattern` matches it). Pre-fix the LABELS
        // list missed it, so plugin rules that introspect environment
        // declarations (e.g. checking for stray `/* eslint-env … */`
        // entries in a flat-config-only project) never saw them.
        'eslint-env',
        'global',
        'globals',
        'exported',
      ];
      const isDirective = (v: string): boolean => {
        for (const label of LABELS) {
          if (!v.startsWith(label)) continue;
          // Exact match (e.g. `/* eslint-disable */`) is fine.
          if (v.length === label.length) return true;
          // Otherwise the next char must be whitespace — `eslint-disable\nfoo`
          // is a real directive (newline counts); `eslint-disable-foo` is not.
          const next = v.charCodeAt(label.length);
          if (next === 0x20 || next === 0x09 || next === 0x0a || next === 0x0d)
            return true;
        }
        return false;
      };
      // Line-comment gate. ESLint v10
      // (`lib/languages/js/source-code/source-code.js`, getInlineConfigNodes)
      // ends its filter with:
      //
      //   comment.type !== "Line" || /^eslint-disable-(next-)?line$/u.test(label)
      //
      // i.e. a Line comment qualifies ONLY when its directive is
      // `eslint-disable-line` / `eslint-disable-next-line`. A
      // `// eslint-disable foo` (Line, but a BLOCK-only directive) is a
      // mistake ESLint never treats as an inline-config node. Pre-fix
      // rslint had no Line gate here, so it returned that bad Line node
      // while `getDisableDirectives` (which re-applies the same gate)
      // dropped it — the two surfaces disagreed about the same comment.
      // We classify the label with the SAME `parseDirective` +
      // `disable-line` regex `getDisableDirectives` uses below, so the
      // two surfaces are now defined by one rule. `rslint-*` is accepted
      // alongside `eslint-*` because the runner treats them as
      // equivalent directive prefixes everywhere else.
      const lineCommentSupported = (c: Comment): boolean => {
        if (c.type !== 'Line') return true;
        const parsed = directiveParser.parseDirective(c.value);
        return (
          parsed != null &&
          /^(?:eslint|rslint)-disable-(?:next-)?line$/u.test(parsed.label)
        );
      };
      return comments.filter(
        // ESLint v10 explicitly skips Shebang comments here regardless
        // of value (`getInlineConfigNodes` returns false for them).
        // Pre-fix runner only filtered by label; in practice the
        // Shebang `!/usr/bin/env …` value never matched any label, so
        // no real-world regression — but matching v10's explicit gate
        // keeps the surface stable against future Shebang values that
        // happen to start with a directive-like word.
        (c) =>
          c.type !== 'Shebang' &&
          isDirective(c.value.trim()) &&
          lineCommentSupported(c),
      );
    },

    /**
     * Mirrors ESLint v10's `SourceCode.getDisableDirectives()`. Parses
     * each `eslint-disable*` / `eslint-enable` comment into a
     * structured Directive (type / node / value / justification) and
     * collects any parse problems (e.g. multi-line
     * `eslint-disable-line` which ESLint rejects). Result is cached
     * on first call — directive parsing is idempotent per file.
     *
     * The parser is `@eslint/plugin-kit`'s `parseDirective` — the same
     * one ESLint v10 uses internally — so the structured shape (label,
     * value, justification) is byte-identical.
     */
    getDisableDirectives() {
      if (_disableDirectives != null) return _disableDirectives;
      const problems: Array<{
        ruleId: null;
        message: string;
        loc: { start: LocPosition; end: LocPosition };
      }> = [];
      const directives: Array<{
        type: 'disable' | 'enable' | 'disable-next-line' | 'disable-line';
        node: Comment;
        value: string;
        justification: string;
      }> = [];

      for (const comment of sc.getInlineConfigNodes()) {
        const parsed = directiveParser.parseDirective(comment.value);
        if (!parsed) continue;
        const { label, value, justification } = parsed;

        // Line comments are valid carriers only for line-flavored
        // directives. Block comments accept any label. Accept both
        // `eslint-*` and `rslint-*` prefixes (same rationale as the
        // suffix-classification block below).
        const lineCommentSupported =
          /^(?:eslint|rslint)-disable-(?:next-)?line$/u.test(label);
        if (comment.type === 'Line' && !lineCommentSupported) continue;

        // `*-disable-line` is a SINGLE-LINE directive by spec — ESLint
        // reports a problem on multi-line block comments carrying that
        // label instead of treating them as a real suppression. Both
        // `eslint-disable-line` and `rslint-disable-line` follow the
        // same rule.
        if (
          (label === 'eslint-disable-line' ||
            label === 'rslint-disable-line') &&
          comment.loc.start.line !== comment.loc.end.line
        ) {
          problems.push({
            ruleId: null,
            message: `${label} comment should not span multiple lines.`,
            loc: comment.loc,
          });
          continue;
        }

        // Accept BOTH `eslint-*` and `rslint-*` prefixes — the
        // suppression engine (`apply-disable-directives.ts`) treats
        // them as equivalent first-class kinds, and
        // `getInlineConfigNodes` (above) already routes both into
        // this loop. Without the parallel `rslint-*` arm here, plugin
        // rules that introspect the directives via
        // `sourceCode.getDisableDirectives()` (e.g.
        // `unicorn/no-abusive-eslint-disable`) saw rslint-prefixed
        // directives as missing while the engine still honored them
        // — inspection API and engine view of the same comment
        // disagreed.
        const eslintSuffix = label.startsWith('eslint-')
          ? label.slice('eslint-'.length)
          : null;
        const rslintSuffix = label.startsWith('rslint-')
          ? label.slice('rslint-'.length)
          : null;
        const suffix = eslintSuffix ?? rslintSuffix;
        if (
          suffix === 'disable' ||
          suffix === 'enable' ||
          suffix === 'disable-next-line' ||
          suffix === 'disable-line'
        ) {
          directives.push({
            type: suffix,
            node: comment,
            value: value ?? '',
            justification: justification ?? '',
          });
        }
      }

      _disableDirectives = { problems, directives };
      return _disableDirectives;
    },
  };

  return sc as unknown as SourceCode;
}

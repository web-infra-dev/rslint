/* rslint-disable @typescript-eslint/no-unsafe-type-assertion */
/**
 * Pure, stateless helpers for `source-code.ts`: scope helpers take the
 * ScopeManager (+ AST root where needed); token-option normalizers take the
 * user's options struct. Split out to keep the SourceCode factory focused on
 * wiring the public object.
 */

import type {
  ESTreeNode,
  TokenCountOpts,
  TokenSkipOpts,
} from './source-code.js';
import type { Token } from './token-builder.js';

// ─────────────────────────────────────────────────────────────────────
// Line-start offset lookup — shared safe accessor for `lso[line - 1]`
// ─────────────────────────────────────────────────────────────────────

/**
 * Return the byte offset where `line1Based` starts. Clamps both ends so
 * out-of-range inputs (a rule reporting `loc.line` beyond EOF, line < 1)
 * resolve to a safe in-range offset instead of `undefined`.
 *
 *   line < 1                        → 0
 *   line > lso.length               → textLength
 *   otherwise                       → lso[line - 1]
 *
 * Centralising this avoids each call site re-deriving the clamp; a prior
 * inline `lso[Math.max(0, line - 1)]` only clamped the low end and let
 * NaN/undefined propagate into diagnostic ranges.
 */
export function lineStartOffset(
  lso: readonly number[],
  line1Based: number,
  textLength: number,
): number {
  if (line1Based <= 0) return 0;
  if (line1Based - 1 >= lso.length) return textLength;
  return lso[line1Based - 1];
}

// ─────────────────────────────────────────────────────────────────────
// Token-option normalizers — shape the polymorphic ESLint inputs
// ─────────────────────────────────────────────────────────────────────

/**
 * Normalize the singular-token API options. See `TokenSkipOpts` for
 * the input shapes. Empirically verified against ESLint:
 *
 *   - `getFirstToken(decl, 0)` → first match
 *   - `getFirstToken(decl, 1)` → second match
 *   - `getFirstToken(decl, 2)` → third match (skip 2)
 *
 * The `skip` semantics caught us in the audit: bare numbers mean
 * "skip N, then return the next", NOT "return the first N". Plugin
 * code that uses a number argument is essentially always doing skip;
 * treating it differently silently picks the wrong token.
 */
export function normalizeSkipOpts(opts: TokenSkipOpts | undefined): {
  skip: number;
  filter: ((t: Token) => boolean) | null;
  includeComments: boolean;
} {
  if (opts == null) return { skip: 0, filter: null, includeComments: false };
  if (typeof opts === 'number') {
    return { skip: Math.max(0, opts), filter: null, includeComments: false };
  }
  if (typeof opts === 'function') {
    return { skip: 0, filter: opts, includeComments: false };
  }
  return {
    skip: typeof opts.skip === 'number' ? Math.max(0, opts.skip) : 0,
    filter: typeof opts.filter === 'function' ? opts.filter : null,
    includeComments: opts.includeComments === true,
  };
}

/**
 * Normalize the `number | Function | Object` shape ESLint accepts on
 * its plural-token API. Defaults match ESLint v10:
 *
 *   - Omitted / function-only opts → no upper bound (returns every
 *     matching token). Represented internally as `Infinity` so the
 *     consumer's `if (out.length >= count) break;` short-circuits
 *     naturally.
 *   - Explicit numeric `count` → cap at that many tokens, INCLUDING
 *     `count = 0` which returns `[]`. This is the v10 behavior
 *     (`SourceCode#getFirstTokens(node, 0) === []`); the previous
 *     implementation treated 0 as "no upper bound" and accidentally
 *     returned the full token range whenever a rule passed a dynamic
 *     count (e.g. `node.params.length` for an empty parameter list).
 *   - Negative counts collapse to 0 — same as `Array.slice` semantics.
 *   - filter = null (every token passes)
 *   - includeComments = false (only code tokens)
 */
export function normalizeCountOpts(opts: TokenCountOpts | undefined): {
  count: number;
  filter: ((t: Token) => boolean) | null;
  includeComments: boolean;
} {
  if (opts == null) {
    return { count: Infinity, filter: null, includeComments: false };
  }
  if (typeof opts === 'number') {
    return { count: Math.max(0, opts), filter: null, includeComments: false };
  }
  if (typeof opts === 'function') {
    return { count: Infinity, filter: opts, includeComments: false };
  }
  return {
    count: typeof opts.count === 'number' ? Math.max(0, opts.count) : Infinity,
    filter: typeof opts.filter === 'function' ? opts.filter : null,
    includeComments: opts.includeComments === true,
  };
}

/**
 * Normalize a `getTokens` / `getTokensBetween` options arg into the
 * three things those methods need: the per-token `filter` predicate, the
 * `includeComments` flag (default `false`, matching ESLint v10's
 * `sourceCode.getTokens(node, { includeComments })` contract), and a
 * numeric `padding`.
 *
 * `includeComments: true` lets plugin code like `@stylistic` and parts
 * of `eslint-plugin-unicorn` collect both code + comment tokens inside a
 * node's range. (The `*Tokens` variants `getFirstTokens` /
 * `getTokensAfter` / `getTokensBefore` use `normalizeCountOpts` /
 * `normalizeSkipOpts`, which already carry `includeComments`.)
 *
 * The `padding` field is non-null ONLY when the caller passed a bare
 * number, mirroring ESLint's `getTokens(node, beforeCount, afterCount)`
 * / `getTokensBetween(left, right, padding)` overloads that route to a
 * `PaddedTokenCursor` (`lib/languages/js/source-code/token-store`):
 * a numeric arg expands the returned slice by that many CODE tokens on
 * each side rather than filtering. `PaddedTokenCursor extends
 * ForwardTokenCursor`, which iterates the TOKENS array only — so the
 * numeric form never interleaves comments (`includeComments` stays
 * false) and never carries a `filter`. `beforeCount | 0` in ESLint
 * coerces the count to a 32-bit int and clamps negatives via the
 * cursor's `Math.max(0, …)`; we reproduce that with `Math.trunc` +
 * `Math.max(0, …)`. Pre-fix this helper had no numeric branch (unlike
 * `normalizeSkipOpts` / `normalizeCountOpts`), so a numeric padding arg
 * was silently dropped and the padded call returned the unpadded slice.
 */
export function normalizeFilterOpts(
  opts: import('./source-code.js').TokenFilterOpts | undefined,
): {
  filter: ((t: Token) => boolean) | null;
  includeComments: boolean;
  padding: number | null;
} {
  if (opts == null)
    return { filter: null, includeComments: false, padding: null };
  if (typeof opts === 'number') {
    return {
      filter: null,
      includeComments: false,
      padding: Math.max(0, Math.trunc(opts)),
    };
  }
  if (typeof opts === 'function')
    return { filter: opts, includeComments: false, padding: null };
  return {
    filter: typeof opts.filter === 'function' ? opts.filter : null,
    includeComments: opts.includeComments === true,
    padding: null,
  };
}

/**
 * `PaddedTokenCursor`-equivalent slice for the numeric-padding overloads
 * of `getTokens` / `getTokensBetween`.
 *
 * Mirrors ESLint's
 * `lib/languages/js/source-code/token-store/padded-token-cursor.js`
 * (a `ForwardTokenCursor` with an inflated index range):
 *
 *   index    = getFirstIndex(tokens, startLoc)          // first token whose
 *                                                       //   range[0] >= startLoc
 *   indexEnd = getLastIndex(tokens, endLoc)             // last token whose
 *                                                       //   range[1] <= endLoc
 *   index    = Math.max(0, index - beforeCount)
 *   indexEnd = Math.min(tokens.length - 1, indexEnd + afterCount)
 *   return tokens.slice(index, indexEnd + 1)
 *
 * Operates on the CODE-token array only (never comments): in ESLint the
 * numeric overloads construct a `PaddedTokenCursor extends
 * ForwardTokenCursor`, which iterates `tokens` and ignores comments. So
 * this is intentionally NOT passed the comment-merged stream.
 *
 * `getFirstIndex` / `getLastIndex` resolve, for a real boundary loc, to
 * the first/last token fully inside `[startLoc, endLoc]`. The linear
 * scans below compute the same first/last index; when the range
 * contains no token (`firstIdx > lastIdx`) the unpadded slice is empty
 * and padding inflates symmetrically around that gap — matching the
 * cursor's index arithmetic.
 */
export function paddedTokenSlice(
  tokens: readonly Token[],
  startLoc: number,
  endLoc: number,
  beforeCount: number,
  afterCount: number,
): Token[] {
  const len = tokens.length;
  if (len === 0) return [];
  // getFirstIndex: first token whose start is at/after startLoc.
  let firstIdx = len;
  for (let i = 0; i < len; i++) {
    if (tokens[i].range[0] >= startLoc) {
      firstIdx = i;
      break;
    }
  }
  // getLastIndex: last token whose end is at/before endLoc.
  let lastIdx = -1;
  for (let i = len - 1; i >= 0; i--) {
    if (tokens[i].range[1] <= endLoc) {
      lastIdx = i;
      break;
    }
  }
  const start = Math.max(0, firstIdx - beforeCount);
  const end = Math.min(len - 1, lastIdx + afterCount);
  if (start > end) return [];
  return tokens.slice(start, end + 1);
}

// ─────────────────────────────────────────────────────────────────────
// Scope helpers — pure functions over the ScopeManager
// ─────────────────────────────────────────────────────────────────────

/**
 * Mirrors ESLint v10's `sourceCode.getScope(node)` algorithm
 * (`lib/languages/js/source-code/source-code.js`):
 *
 *   1. Walk up `node.parent` until `scopeManager.acquire(curr, inner)`
 *      returns a scope. `acquire` only resolves on scope-defining
 *      blocks (Program, Function*, BlockStatement under strict, ...),
 *      so the walk is necessary — calling `acquire(CallExpression)`
 *      returns `undefined`.
 *   2. `inner = true` everywhere except at Program (where `false`
 *      picks the module scope over the global scope for ES modules;
 *      `true` would pick the global scope and miss module-specific
 *      bindings — pre-fix rslint did this and silently broke rules
 *      gated on `variableScope.type === 'module'` like
 *      eslint-plugin-import/no-commonjs and /no-amd).
 *   3. Unwrap `function-expression-name` scopes by returning their
 *      first child — ESLint's exact behavior here.
 *   4. Fall back to `scopeManager.scopes[0]` when the walk hits the
 *      top without a match (defensive — shouldn't happen in practice
 *      because Program always has a scope).
 */
export function getScopeForNode(
  scopeManager: unknown,
  node?: ESTreeNode,
): unknown {
  const sm = scopeManager as {
    acquire?: (n: ESTreeNode, inner?: boolean) => unknown;
    scopes?: Array<{ type?: string; childScopes?: unknown[] }>;
    globalScope?: unknown;
  } | null;
  if (!sm || typeof sm.acquire !== 'function') return null;
  if (!node) return sm.globalScope ?? sm.scopes?.[0] ?? null;
  // ESLint v10: `inner` is computed ONCE from the starting node's
  // type, NOT recomputed at each step of the parent walk. For any
  // descendant of Program, `inner === true` is used the whole way
  // up, so when the walk reaches Program, `acquire(Program, true)`
  // returns the module scope (the inner one for `sourceType: module`)
  // rather than the global scope. Recomputing per step would pick
  // the wrong scope at Program and miscategorise every top-level
  // node — pre-fix rslint did exactly that.
  const inner = node.type !== 'Program';
  for (
    let curr: ESTreeNode | null | undefined = node;
    curr;
    curr = curr.parent
  ) {
    const scope = sm.acquire(curr, inner) as
      | { type?: string; childScopes?: unknown[] }
      | undefined;
    if (scope) {
      if (scope.type === 'function-expression-name') {
        return (scope.childScopes ?? [])[0] ?? scope;
      }
      return scope;
    }
  }
  return sm.scopes?.[0] ?? sm.globalScope ?? null;
}

/**
 * Forwards to the scope manager's `getDeclaredVariables(node)` if
 * present. Both `eslint-scope` and `@typescript-eslint/scope-manager`
 * expose this; returns `[]` if the underlying manager doesn't.
 */
export function getDeclaredVariablesFromScopeManager(
  scopeManager: unknown,
  node: ESTreeNode,
): unknown[] {
  const sm = scopeManager as {
    getDeclaredVariables?: (n: ESTreeNode) => unknown[];
  } | null;
  if (sm && typeof sm.getDeclaredVariables === 'function') {
    return sm.getDeclaredVariables(node) ?? [];
  }
  return [];
}

/**
 * ESLint v10's `sourceCode.markVariableAsUsed(name, refNode)` walks
 * up the scope chain from `refNode`'s enclosing scope through `.upper`
 * pointers, sets `eslintUsed = true` on the first matching variable,
 * and returns whether one was found. `eslint-scope` doesn't expose a
 * `ScopeManager.markVariableAsUsed` method (the helper lives on
 * SourceCode in ESLint itself), so the previous shim — which delegated
 * to `sm.markVariableAsUsed` and bailed early when missing — always
 * returned false. Plugin rules calling this to suppress `no-unused-vars`
 * for their generated bindings (e.g. react/jsx-uses-vars) would
 * silently miss their target.
 */
export function markVariableAsUsedInScopeChain(
  scopeManager: unknown,
  ast: ESTreeNode,
  name: string,
  refNode?: ESTreeNode,
): boolean {
  const sm = scopeManager as {
    acquire?: (n: ESTreeNode, inner?: boolean) => unknown;
    globalScope?: { upper?: unknown };
    scopes?: Array<{ upper?: unknown }>;
  } | null;
  if (!sm || typeof sm.acquire !== 'function') return false;
  // ESLint v10 always uses `inner=true` for the start scope here —
  // even when refNode is `ast` (Program). At Program with sourceType
  // 'module' this picks the module scope (not the global scope), so
  // module-level `const`s are reachable by walking the upper chain.
  // Diverges from `getScope` (which picks `inner = (start ≠ Program)`)
  // — different lookup semantics, matched against ESLint's actual
  // `lib/languages/js/source-code/source-code.js`.
  const startNode = refNode ?? ast;
  type ScopeView = {
    variables?: Array<{ name: string; eslintUsed?: boolean }>;
    upper?: unknown;
  };
  // Resolve the starting scope by walking startNode → parent → ...
  // until `acquire` returns a scope. Critical: do NOT fall through
  // directly to `globalScope` when the starting node isn't itself a
  // scope-defining block (Identifier, Literal, etc.). For a refNode
  // inside an ESM `const` declaration, the enclosing scope is the
  // MODULE scope, not the GLOBAL scope. Pre-fix the initial
  // `acquire(startNode, true) ?? globalScope` short-circuited to
  // globalScope, missed the module-scope binding, and returned false
  // for any `markVariableAsUsed('x', refNode)` call where `x` lived
  // in module / function / block scope. ESLint v9's
  // `lib/languages/js/source-code/source-code.js` walks the parent
  // chain identically.
  let scope: ScopeView | null = null;
  for (
    let curr: ESTreeNode | null | undefined = startNode;
    curr;
    curr = curr.parent
  ) {
    const found = sm.acquire(curr, true) as ScopeView | undefined;
    if (found) {
      scope = found;
      break;
    }
  }
  // Only fall back to globalScope when the parent walk found nothing
  // — i.e. startNode has no enclosing scope-defining ancestor (would
  // happen for orphan AST fragments). In well-formed source, the
  // walk always reaches Program at the top and `acquire(Program, true)`
  // returns the program / module scope.
  if (!scope) {
    scope = (sm.globalScope as ScopeView | undefined) ?? null;
  }
  while (scope) {
    const variable = scope.variables?.find((v) => v.name === name);
    if (variable) {
      variable.eslintUsed = true;
      return true;
    }
    scope = (scope.upper as typeof scope) ?? null;
  }
  return false;
}

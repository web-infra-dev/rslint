/* eslint-disable @typescript-eslint/no-unsafe-type-assertion -- oxc-parser's `visitorKeys` export is typed loosely at the package boundary; the cast projects it into the typed shape this module uses. */
/**
 * Visitor-key access — the "what children does this node have" lookup,
 * used by the AST walker (`normalize-ast`) and the listener traversal
 * (`listener-merge`) to recurse in ESLint-compatible order.
 *
 * Source of truth: **oxc-parser's own `visitorKeys` export**. Since the
 * runner walks the AST oxc PRODUCES, oxc's table is authoritative by
 * construction — it covers exactly the node types oxc can emit (JS +
 * JSX + TS; 165 types). This replaces a former direct dependency on
 * `@typescript-eslint/visitor-keys`.
 *
 * Verified (see `tests/visitor-keys.test.ts`, a drift guard):
 *   - byte-identical to `@typescript-eslint/visitor-keys` for every
 *     shared node type (0 order/content mismatch), so traversal order
 *     stays exactly what espree / @typescript-eslint/parser declare;
 *   - complete for oxc's output (no emitted node type is missing from
 *     the table), so the `getVisitorKeys` fallback is effectively
 *     unreachable in production.
 * The guard re-runs that comparison against the canonical keys so a
 * future oxc-parser version that drifts from ESLint fails CI loudly.
 *
 * Stored as `Object.create(null)` so V8 keeps a stable hidden class for
 * monomorphic property access in the hot traversal loop — measurably
 * faster than `Map.get` on the 5000+-file workload, and 0-alloc per
 * node (vs the old `Object.keys(node)` + blacklist path).
 */
import { visitorKeys as oxcVisitorKeys } from 'oxc-parser';

const _src = oxcVisitorKeys as Record<string, readonly string[]>;
const _vkObj = Object.create(null) as Record<string, readonly string[]>;
for (const k of Object.keys(_src)) {
  _vkObj[k] = _src[k];
}
Object.freeze(_vkObj);
export const VISITOR_KEYS: Readonly<Record<string, readonly string[]>> = _vkObj;

// Keys that are never child nodes — matches `eslint-visitor-keys`'
// `getKeys` blocklist (`parent` back-ref + comment attachments). The
// walker's own `value && value.type` recursion guard skips the
// remaining non-node fields (`type` / `loc` / `range` / `start` /
// `end`), so the fallback only needs to drop these.
const NON_CHILD_KEYS = new Set([
  'parent',
  'leadingComments',
  'trailingComments',
]);

/**
 * ESLint Traverser-equivalent fallback for a node type NOT in the table
 * (`lib/shared/traverser.js` falls back to `vk.getKeys(node)` the same
 * way). oxc.visitorKeys is complete for oxc's output, so this is reached
 * only for synthetic / unexpected nodes: enumerate the node's keys and
 * drop the non-child fields. The caller's `child.type` check filters the
 * rest.
 */
export function getVisitorKeys(node: {
  type?: string;
  [k: string]: unknown;
}): readonly string[] {
  const t = node.type;
  if (t !== undefined) {
    const keys = VISITOR_KEYS[t];
    if (keys !== undefined) return keys;
  }
  return Object.keys(node).filter((k) => !NON_CHILD_KEYS.has(k));
}

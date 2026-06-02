/**
 * Visitor-keys drift guard — pins the runner's `VISITOR_KEYS_TABLE` against
 * the ESLint ecosystem's canonical visitor keys.
 *
 * The table is vendored from oxc-parser 0.133 (the parser whose AST the
 * runner walks), but the runner's JOB is to reproduce ESLint v10 traversal
 * order — so the guard diffs the table against ESLint's own packages, NOT
 * against oxc. That keeps it meaningful after the npm `oxc-parser` dep is
 * removed (M7): nothing here imports oxc.
 *
 *   - TS traversal → `@typescript-eslint/visitor-keys` (what
 *     `@typescript-eslint/parser` declares). The table is byte-identical on
 *     every shared node type.
 *   - plain-JS traversal → `eslint-visitor-keys` (what ESLint core / espree
 *     declare). The table's projection onto each JS node's keys matches.
 *
 * The child ORDER inside each key array is load-bearing: it sets the
 * listener-firing / traversal order, so a reorder/drop/add must fail CI
 * rather than silently change rule behavior.
 *
 * Why two oracles: oxc emits a TS-aware AST even for `.js`, so the table
 * tracks the TS superset (`@typescript-eslint/visitor-keys`). The TS-only
 * keys it adds (`typeAnnotation`, `typeParameters`, `decorators`, ...) are
 * absent from the nodes a plain-JS file produces, so the walker's
 * `child.type` check skips them and JS traversal order is unchanged — which
 * is exactly what the `eslint-visitor-keys` projection test below proves.
 */
import { describe, test, expect } from '@rstest/core';
import { visitorKeys as tsVisitorKeys } from '@typescript-eslint/visitor-keys';
import { KEYS as eslintCoreKeys } from 'eslint-visitor-keys';

import {
  VISITOR_KEYS,
  getVisitorKeys,
} from '../../../src/eslint-plugin/ast/visitor-keys.js';
import { VISITOR_KEYS_TABLE } from '../../../src/eslint-plugin/ast/visitor-keys-table.js';

// Node types oxc emits that have NO `@typescript-eslint/visitor-keys`
// counterpart: oxc-specific shapes (preserved parens, JSDoc nullable types,
// V8 intrinsics). With no ESLint authority to diff against, their keys are
// pinned here by hand. The "membership is locked" test binds this set to the
// table's vendored-only set, so a new oxc-only node type forces an update
// here rather than slipping through unchecked.
const OXC_ONLY_KEYS: Record<string, readonly string[]> = {
  ParenthesizedExpression: ['expression'],
  TSParenthesizedType: ['typeAnnotation'],
  TSJSDocNonNullableType: ['typeAnnotation'],
  TSJSDocNullableType: ['typeAnnotation'],
  TSJSDocUnknownType: [],
  V8IntrinsicExpression: ['name', 'arguments'],
};

// `@typescript-eslint/visitor-keys` node types the runner's table omits, and
// why it's correct to omit them: 9 modifier-as-node `TS*Keyword`s + 2
// deprecated experimental nodes — oxc never emits any of these, so the
// walker never needs their keys.
const TS_ESLINT_ONLY = [
  'ExperimentalRestProperty',
  'ExperimentalSpreadProperty',
  'TSAbstractKeyword',
  'TSAsyncKeyword',
  'TSDeclareKeyword',
  'TSExportKeyword',
  'TSPrivateKeyword',
  'TSProtectedKeyword',
  'TSPublicKeyword',
  'TSReadonlyKeyword',
  'TSStaticKeyword',
];

describe('visitor-keys drift guard (aligned to ESLint v10)', () => {
  const tableKeys = Object.keys(VISITOR_KEYS_TABLE);

  test('byte-identical to @typescript-eslint/visitor-keys on every shared node type', () => {
    const shared = tableKeys.filter((k) => k in tsVisitorKeys);
    // Pins the intersection size — a node type silently leaving it (renamed
    // or dropped on either side) fails here.
    expect(shared).toHaveLength(159);
    // Exact array per type (order + content). The TS superset's order is
    // what @typescript-eslint/parser walks; any drift changes rule behavior.
    // Collect mismatches so a failure names every drifted node type at once.
    const mismatched = shared.filter(
      (k) =>
        JSON.stringify(VISITOR_KEYS_TABLE[k]) !==
        JSON.stringify(tsVisitorKeys[k]),
    );
    expect(mismatched).toEqual([]);
  });

  test('JS-node projection equals eslint-visitor-keys (ESLint core / espree order)', () => {
    const shared = tableKeys.filter((k) => k in eslintCoreKeys);
    // Iterates from the TABLE's keys, so a node type present in
    // eslint-visitor-keys but ENTIRELY ABSENT from the table (the
    // deprecated ExperimentalRest/SpreadProperty) won't enter this loop.
    // That "core-only type" gap is closed by the membership-lock test
    // below — both are pinned in TS_ESLINT_ONLY there.
    expect(shared).toHaveLength(87);
    // Drop the TS-only keys oxc adds, keeping the table's order; what remains
    // must equal eslint-visitor-keys EXACTLY (same keys, same order). Catches
    // both a missing core key and a reorder that would shift plain-JS
    // traversal away from ESLint core.
    const drifted = shared.filter((k) => {
      const core = eslintCoreKeys[k];
      const projection = VISITOR_KEYS_TABLE[k].filter((key) =>
        core.includes(key),
      );
      return JSON.stringify(projection) !== JSON.stringify(core);
    });
    expect(drifted).toEqual([]);
  });

  test('oxc-only node types (no ESLint counterpart) keep pinned keys', () => {
    for (const [k, expected] of Object.entries(OXC_ONLY_KEYS)) {
      expect(VISITOR_KEYS_TABLE[k]).toEqual(expected);
      // Confirm there genuinely is no ESLint authority for these — if one
      // appears upstream, switch the entry to a real cross-check.
      expect(k in tsVisitorKeys).toBe(false);
      expect(k in eslintCoreKeys).toBe(false);
    }
  });

  test('table membership is locked (additions/removals on either side fail)', () => {
    // Exactly the oxc-only set is absent from @typescript-eslint.
    const vendoredOnly = tableKeys.filter((k) => !(k in tsVisitorKeys)).sort();
    expect(vendoredOnly).toEqual(Object.keys(OXC_ONLY_KEYS).sort());

    // Exactly the documented modifier/experimental set is absent from the table.
    const tsOnly = Object.keys(tsVisitorKeys)
      .filter((k) => !(k in VISITOR_KEYS_TABLE))
      .sort();
    expect(tsOnly).toEqual([...TS_ESLINT_ONLY].sort());

    // Total count — a fast, unambiguous signal when a diff is large.
    expect(tableKeys).toHaveLength(165);
  });

  test('VISITOR_KEYS export is the vendored table, on a null prototype', () => {
    // The runtime export must be exactly the vendored table...
    expect({ ...VISITOR_KEYS }).toEqual({ ...VISITOR_KEYS_TABLE });
    // ...stored null-proto for monomorphic hot-loop access (see
    // visitor-keys.ts); Object.create(null) has no prototype.
    expect(Object.getPrototypeOf(VISITOR_KEYS)).toBe(null);
  });

  test('getVisitorKeys uses the table (not enumeration) for known types', () => {
    // A known type carrying noise fields: the table lookup must win and
    // return the fixed entry — NOT `Object.keys(node)` (the fallback),
    // which would leak `extra` / `parent`. The `not.toContain` is what
    // gives this test teeth: it fails if a known type wrongly falls
    // through to the enumerate branch.
    const keys = getVisitorKeys({
      type: 'JSXElement',
      openingElement: {},
      children: [],
      closingElement: {},
      extra: {},
      parent: {},
    });
    expect(keys).toEqual(VISITOR_KEYS['JSXElement']);
    expect(keys).not.toContain('extra');
    expect(keys).not.toContain('parent');
  });

  test('getVisitorKeys fallback enumerates child keys in order, dropping the blocklist', () => {
    const keys = getVisitorKeys({
      type: 'SomeSyntheticNodeType',
      // child-ish fields, in declaration order
      foo: {},
      bar: [],
      // non-child fields the blocklist must drop
      parent: {},
      leadingComments: [],
      trailingComments: [],
    });
    // Exact result + order: the fallback is `Object.keys(node)` minus the
    // NON_CHILD_KEYS blocklist. `type` is intentionally kept (not in the
    // blocklist) — the walker's own `child.type` guard skips its string
    // value, so it's harmless — and the relative order of the real child
    // keys (foo before bar) is preserved, which is load-bearing for
    // traversal order.
    expect(keys).toEqual(['type', 'foo', 'bar']);
  });
});

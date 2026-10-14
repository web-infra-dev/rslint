/** Pin canonical ESTree child order and complete frame-schema coverage. */
import { describe, test, expect } from 'rstack/test';
import { visitorKeys as tsVisitorKeys } from '@typescript-eslint/visitor-keys';
import { KEYS as eslintCoreKeys } from 'eslint-visitor-keys';

import {
  VISITOR_KEYS,
  getVisitorKeys,
} from '../../../src/languages/js/visitor-keys.js';
import { FRAME_SCHEMAS } from '../../../src/languages/js/frame-schema.js';
import { VISITOR_KEYS_TABLE } from '../../../src/languages/js/visitor-keys-table.js';

// Modifier-as-node and deprecated experimental types are not emitted by Program.
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

  test('every frame type has all canonical child fields', () => {
    for (const [type, fields] of FRAME_SCHEMAS) {
      expect(VISITOR_KEYS_TABLE[type], type).toBeDefined();
      for (const key of VISITOR_KEYS_TABLE[type])
        expect(fields, `${type}.${key}`).toContain(key);
    }
  });

  test('JS-node projection equals eslint-visitor-keys (ESLint core / espree order)', () => {
    const shared = tableKeys.filter((k) => k in eslintCoreKeys);
    // Iterates from the TABLE's keys, so a node type present in
    // eslint-visitor-keys but ENTIRELY ABSENT from the table (the
    // deprecated ExperimentalRest/SpreadProperty) won't enter this loop.
    // That "core-only type" gap is closed by the membership-lock test
    // below — both are pinned in TS_ESLINT_ONLY there.
    expect(shared).toHaveLength(87);
    // Drop the TS-only keys in the table, keeping the table's order; what remains
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

  test('table membership is locked (additions/removals on either side fail)', () => {
    // Every supported node type has canonical visitor keys.
    const vendoredOnly = tableKeys.filter((k) => !(k in tsVisitorKeys)).sort();
    expect(vendoredOnly).toEqual([]);

    // Exactly the documented modifier/experimental set is absent from the table.
    const tsOnly = Object.keys(tsVisitorKeys)
      .filter((k) => !(k in VISITOR_KEYS_TABLE))
      .sort();
    expect(tsOnly).toEqual([...TS_ESLINT_ONLY].sort());

    // Total count — a fast, unambiguous signal when a diff is large.
    expect(tableKeys).toHaveLength(159);
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

  test('SourceCode queries retain ESLint traversal of caller-created synthetic nodes', () => {
    expect(
      getVisitorKeys({
        type: 'WrapperSynthetic',
        child: {},
        parent: {},
        leadingComments: [],
        trailingComments: [],
      }),
    ).toEqual(['type', 'child']);
  });
});

import { describe, test, expect } from '@rstest/core';
import { buildParserServicesFromSnapshot } from '../../../src/eslint-plugin/source-code/parser-services-from-snapshot.js';

// Raw tsgo TypeFlags the decoder compares against (mirrors the constants in
// parser-services-from-snapshot.ts).
const TSGO_UNION = 134217728; // 1 << 27

const MAGIC = 0x52534e50; // "RSNP"
const VERSION = 1;

interface TestType {
  id: number;
  flags?: number;
  isArray?: boolean;
  isTuple?: boolean;
  members?: number[];
  typeArgs?: number[];
  callSigs?: number[];
  name?: string;
}

// Test-only mirror of internal/typesnapshot/encode_binary.go so the worker
// decoder can be exercised in isolation (search / lazy-decode / bounds). The
// real producer is Go; the enforce-effect end-to-end test guards Go↔worker
// layout agreement, this guards the decoder's own logic. Returns an ArrayBuffer
// — the exact shape the worker receives post-transfer (no base64 round-trip).
function encode(snap: {
  // [tokenStart, end, typeId]
  node2type: Array<[number, number, number]>;
  types: TestType[];
  primUndefined?: number;
  primNull?: number;
}): ArrayBuffer {
  const n2t = [...snap.node2type].sort((a, b) => a[0] - b[0] || a[1] - b[1]);
  const types = [...snap.types].sort((a, b) => a.id - b.id);

  const blocks = types.map((t) => {
    const name = Buffer.from(t.name ?? '', 'utf8');
    const idLists = [t.members ?? [], t.typeArgs ?? [], t.callSigs ?? []];
    const len =
      4 +
      1 +
      4 +
      name.length +
      idLists.reduce((s, a) => s + 4 + a.length * 4, 0);
    const b = new Uint8Array(len);
    const dv = new DataView(b.buffer);
    let p = 0;
    dv.setInt32(p, t.flags ?? 0, true);
    p += 4;
    dv.setUint8(p, (t.isArray ? 1 : 0) | (t.isTuple ? 2 : 0));
    p += 1;
    dv.setUint32(p, name.length, true);
    p += 4;
    b.set(name, p);
    p += name.length;
    for (const arr of idLists) {
      dv.setUint32(p, arr.length, true);
      p += 4;
      for (const id of arr) {
        dv.setInt32(p, id, true);
        p += 4;
      }
    }
    return b;
  });

  const blobLen = blocks.reduce((s, b) => s + b.length, 0);
  const buf = new Uint8Array(
    24 + n2t.length * 12 + types.length * 12 + blobLen,
  );
  const dv = new DataView(buf.buffer);
  let p = 0;
  dv.setUint32(p, MAGIC, true);
  p += 4;
  dv.setUint32(p, VERSION, true);
  p += 4;
  dv.setUint32(p, n2t.length, true);
  p += 4;
  dv.setUint32(p, types.length, true);
  p += 4;
  dv.setInt32(p, snap.primUndefined ?? 0, true);
  p += 4;
  dv.setInt32(p, snap.primNull ?? 0, true);
  p += 4;
  for (const [s, e, id] of n2t) {
    dv.setInt32(p, s, true);
    dv.setInt32(p + 4, e, true);
    dv.setInt32(p + 8, id, true);
    p += 12;
  }
  let blobOff = 0;
  for (let i = 0; i < types.length; i++) {
    dv.setInt32(p, types[i].id, true);
    dv.setUint32(p + 4, blobOff, true);
    dv.setUint32(p + 8, blocks[i].length, true);
    p += 12;
    blobOff += blocks[i].length;
  }
  for (const b of blocks) {
    buf.set(b, p);
    p += b.length;
  }
  return buf.buffer;
}

describe('buildParserServicesFromSnapshot (binary wire)', () => {
  test('decodes node2type + lazily resolves union / array type blocks', () => {
    const snapshot = encode({
      node2type: [
        [10, 20, 5],
        [5, 8, 3],
        [30, 40, 9],
      ], // deliberately unsorted — encoder sorts, decoder binary-searches
      types: [
        { id: 3, flags: TSGO_UNION, members: [4, 5] },
        { id: 4, flags: 4, name: 'string' },
        { id: 5, flags: 1, isArray: true, typeArgs: [4], callSigs: [3] },
        { id: 9, flags: 4 },
      ],
      primUndefined: 4,
    });
    const ps = buildParserServicesFromSnapshot(snapshot) as any;
    expect(typeof ps.getTypeAtLocation).toBe('function');

    const union = ps.getTypeAtLocation({ range: [5, 8] });
    expect(union.id).toBe(3);
    expect(union.isUnion()).toBe(true);
    expect(union.isIntersection()).toBe(false);
    expect(union.types.map((t: any) => t.id)).toEqual([4, 5]);

    const arr = ps.getTypeAtLocation({ range: [10, 20] });
    const checker = ps.program.getTypeChecker();
    expect(checker.isArrayType(arr)).toBe(true);
    expect(checker.getTypeArguments(arr).map((t: any) => t.id)).toEqual([4]);
    expect(arr.getCallSignatures()[0].getReturnType().id).toBe(3);

    // primTypes resolve; intern identity (same id → same wrapper object) so a
    // rule's `type === checker.getUndefinedType()` comparison works.
    expect(checker.getUndefinedType().id).toBe(4);
    expect(ps.getTypeAtLocation({ range: [5, 8] })).toBe(union);
  });

  test('getTypeAtLocation returns a flagless FALLBACK (never undefined) on a miss', () => {
    const snapshot = encode({
      node2type: [
        [5, 8, 3],
        [10, 10, 7],
        [10, 20, 8],
      ],
      types: [
        { id: 3, flags: 0 },
        { id: 7, flags: 0 },
        { id: 8, flags: 0 },
      ],
    });
    const ps = buildParserServicesFromSnapshot(snapshot) as any;
    expect(ps.getTypeAtLocation({ range: [5, 8] }).id).toBe(3);
    expect(ps.getTypeAtLocation({ range: [10, 10] }).id).toBe(7); // same start, smaller end
    expect(ps.getTypeAtLocation({ range: [10, 20] }).id).toBe(8);
    // A miss returns a never-undefined flagless stand-in (TS contract:
    // getTypeAtLocation never returns undefined; typescript-eslint rules call
    // type.isUnion() with no null check, so undefined would throw mid-rule).
    for (const r of [
      [5, 9],
      [10, 15],
      [1, 2],
      [99, 99],
    ] as const) {
      const t = ps.getTypeAtLocation({ range: r });
      expect(t).not.toBeUndefined();
      expect(t.isUnion()).toBe(false);
      expect(t.isIntersection()).toBe(false);
      expect(t.types).toEqual([]);
    }
  });

  test('getTypeAtLocation strips a type annotation off an Identifier range before lookup', () => {
    // tgo records the Identifier at its NAME span; oxc gives the same annotated
    // identifier a range[1] that runs to the ANNOTATION end. Without stripping,
    // the oxc range misses; with stripping (range[0]+name.length) it hits.
    const snapshot = encode({
      node2type: [[5, 8, 3]], // tgo Identifier "abc" at (5,8) → type 3
      types: [
        { id: 3, flags: TSGO_UNION, members: [4] },
        { id: 4, flags: 4 },
      ],
    });
    const ps = buildParserServicesFromSnapshot(snapshot) as any;
    // oxc annotated identifier `abc: Foo` spans 5..14 (to the annotation end).
    const annotated = {
      type: 'Identifier',
      name: 'abc',
      typeAnnotation: {},
      range: [5, 14] as const,
    };
    expect(ps.getTypeAtLocation(annotated).id).toBe(3);
    expect(ps.getTypeAtLocation(annotated).isUnion()).toBe(true);
    // a plain non-Identifier hit still works on the raw range.
    expect(ps.getTypeAtLocation({ range: [5, 8] }).id).toBe(3);
    // ANY Identifier ends at range[0]+name.length, with OR without a
    // typeAnnotation field — an escaped identifier's oxc range[1] overshoots the
    // decoded name, so stripping is uniform (not gated on typeAnnotation), to
    // match snapshot.go keying identifiers on the decoded name length.
    expect(
      ps.getTypeAtLocation({ type: 'Identifier', name: 'abc', range: [5, 14] })
        .id,
    ).toBe(3);
    // a NON-Identifier whose range[1] matches no entry misses → FALLBACK.
    expect(ps.getTypeAtLocation({ range: [5, 14] }).id).toBe(0);
  });

  test('degrades to {} (silently) on absent / malformed / truncated snapshot', () => {
    // Non-ArrayBuffer inputs: the worker only ever receives a transferred
    // ArrayBuffer, so a base64 string / number / undefined is never valid.
    expect(buildParserServicesFromSnapshot(undefined)).toEqual({});
    expect(buildParserServicesFromSnapshot('')).toEqual({});
    expect(buildParserServicesFromSnapshot(123)).toEqual({});
    // ≥24 bytes (passes the header-length check) but wrong magic
    const wrongMagic = new Uint8Array(24);
    new DataView(wrongMagic.buffer).setUint32(0, 0xdeadbeef, true);
    expect(buildParserServicesFromSnapshot(wrongMagic.buffer)).toEqual({});

    const valid = encode({
      node2type: [[5, 8, 3]],
      types: [{ id: 3, flags: 0 }],
    });
    // shorter than the 24-byte header
    expect(buildParserServicesFromSnapshot(valid.slice(0, 10))).toEqual({});
    // header intact but body truncated (blobStart past end)
    expect(buildParserServicesFromSnapshot(valid.slice(0, 30))).toEqual({});
  });

  test('corrupted/truncated interior block degrades to a flagless type, no throw', () => {
    // Header + node2type + types-index intact (blobStart passes the entry
    // guard), but the type block body is cut off — decodeBlock must bounds-check
    // and degrade silently (the M0.5 design throws RangeError here).
    const valid = encode({
      node2type: [[5, 8, 3]],
      types: [{ id: 3, flags: 134217728, members: [1, 2] }],
    });
    const blobStart = 24 + 1 * 12 + 1 * 12; // header + 1 node2type + 1 index
    const ps = buildParserServicesFromSnapshot(
      valid.slice(0, blobStart + 2),
    ) as any;
    expect(typeof ps.getTypeAtLocation).toBe('function');
    let t: any;
    expect(() => {
      t = ps.getTypeAtLocation({ range: [5, 8] });
    }).not.toThrow();
    // lookup resolved the id, but the unreadable block degrades to flagless
    expect(t.id).toBe(3);
    expect(t.isUnion()).toBe(false);
    expect(t.types).toEqual([]);
  });

  test('esTreeNodeToTSNodeMap.get returns a span-carrying proxy', () => {
    // getParserServices requires esTreeNodeToTSNodeMap; its get() returns a
    // span-only proxy (no kind — the bridge carries no SyntaxKind).
    const snapshot = encode({
      node2type: [[5, 8, 3]],
      types: [{ id: 3, flags: 0 }],
    });
    const ps = buildParserServicesFromSnapshot(snapshot) as any;
    const proxy = ps.esTreeNodeToTSNodeMap.get({ range: [5, 8] });
    expect(proxy.__start).toBe(5);
    expect(proxy.__end).toBe(8);
    expect(proxy.kind).toBeUndefined();
  });
});

/**
 * Reconstructs a minimal `parserServices` from a Go-built type snapshot so
 * third-party type-aware ESLint plugin rules can run in the worker without a
 * second TypeScript implementation.
 *
 * Wire format (random-access binary; see internal/typesnapshot/encode_binary.go):
 *   - The snapshot arrives as a transferred ArrayBuffer: the host decodes the wire
 *     base64 ONCE (or, on the CLI binary frame, gets bytes directly) and transfers
 *     ownership to the worker via postMessage transferList — no per-file base64
 *     decode and no structuredClone copy in the worker. It is read IN PLACE via
 *     DataView — no eager parse, no Map of every node.
 *   - node2type is a fixed-width array sorted by (tokenStart, end); getTypeAtLocation
 *     BINARY-SEARCHES it by an oxc ESTree node's (range[0], range[1]). (verified:
 *     oxc.start == tsgo tokenPos; Go skips type-annotation nodes so the key is
 *     unique across the value/declaration nodes a rule queries.)
 *   - types is a fixed-width index sorted by type-id into a blob section; a type
 *     block is decoded LAZILY (and cached by id) only for the ids a rule actually
 *     touches, so a rule's `type === checker.getUndefinedType()` identity
 *     comparison still works (same id → same wrapper object).
 *   - `isUnion()` / `isIntersection()` use the RAW tsgo TypeFlags constants
 *     (below), NOT a runtime `typescript` import, so this layer needs no
 *     `typescript` of its own. rslint's tsgo type-checker tracks the latest
 *     TypeScript (the ts7 line); a rule's runtime typescript is expected to be
 *     aligned with it, so the raw tsgo `type.flags` a rule reads via
 *     `flags & ts.TypeFlags.X` match directly — ZERO mapping. We deliberately do
 *     NOT support older ts versions (ts5.x renumbered TypeFlags): the bridge is
 *     single-version, aligned to tsgo, with no per-version translation.
 *
 * Degradation stance (handled differently on purpose):
 *   - ABSENT / malformed snapshot (fails the header check) → returns {}. Then
 *     parserServices.program is null and the ecma-language-plugin gate skips the
 *     type-aware rule entirely (it never runs). SILENT.
 *   - MISSING type for a node (getTypeAtLocation finds no (start,end) entry) →
 *     returns FALLBACK_TYPE, NEVER undefined. typescript-eslint rules assume TS's
 *     contract (getTypeAtLocation always yields a type) and call type.isUnion()
 *     with NO null check, so undefined would throw inside the rule. The flagless
 *     stand-in makes the rule see "no banned type" and degrade silently. (The
 *     earlier "return undefined, rule does if(!type)return" stance was WRONG —
 *     real-repo rules don't null-check; this surfaced on rsbuild, #8.) A miss
 *     happens when an oxc node's range doesn't line up with any tsgo node end;
 *     getTypeAtLocation strips a type annotation off an Identifier's range to
 *     reduce these, and FALLBACK_TYPE covers whatever remains.
 *   - CORRUPTED / TRUNCATED interior type block (a count field claiming more
 *     bytes than the block holds) → decodeBlock is bounds-checked and returns
 *     undefined, so that one type degrades to flagless — SILENT, not a throw and
 *     not a whole-snapshot discard. Go's encoder never emits this; the check
 *     keeps the worker symmetric with Go's fail-closed DecodeBinary against any
 *     future wire-format drift.
 *   - UNIMPLEMENTED checker method (a rule calls e.g. getApparentType /
 *     getBaseConstraintOfType, which this minimal checker doesn't expose) →
 *     THROWS, surfacing as a visible rule error. Deliberately NOT stubbed to
 *     undefined: a stub would make the rule silently miss diagnostics (a wrong
 *     answer), whereas a throw honestly signals "not yet supported". Methods are
 *     filled in incrementally as real rules need them (API long-tail), never
 *     stubbed.
 *
 * Known limitations (tracked, not bugs):
 *   - The function-return-type anchor path in some rules goes through
 *     `ts.isFunctionLike(tsNode)`, which needs a runtime-ts SyntaxKind on the
 *     proxy node. That (and SyntaxKind mapping in general) is deferred; the
 *     proxy carries only span, so `ts.isFunctionLike` returns false and that
 *     branch is skipped. The main anchors (variable / parameter / class field /
 *     union / array element) run through `getTypeAtLocation` and are covered.
 *   - A type's `name` (TypeToString) is present in the wire blob but not decoded
 *     here — no covered rule reads `type.name` through this layer yet; add a
 *     lazy `name` getter when one does.
 */

// Raw tsgo TypeFlags (internal/checker/types.go). Stable within tsgo; the
// snapshot carries raw tsgo flags, so we compare against tsgo's own values.
const TSGO_TYPEFLAGS_UNION = 134217728; // 1 << 27
const TSGO_TYPEFLAGS_INTERSECTION = 268435456; // 1 << 28

// Binary header (must match internal/typesnapshot/encode_binary.go).
const SNAPSHOT_MAGIC = 0x52534e50; // "RSNP"
const SNAPSHOT_VERSION = 1;
const HEADER_SIZE = 24;
const N2T_ENTRY_SIZE = 12; // tokenStart i32, end i32, typeId i32
const IDX_ENTRY_SIZE = 12; // typeId i32, blobOffset u32, blobLen u32

interface DecodedBlock {
  flags: number;
  isArray: boolean;
  isTuple: boolean;
  memberTypes: number[];
  typeArgs: number[];
  callSigReturns: number[];
}

// getTypeAtLocation must NEVER return undefined: typescript-eslint rules rely on
// TS's contract (getTypeAtLocation always yields a type) and call type.isUnion()
// with no null check. A node2type miss returns this flagless stand-in instead
// (isUnion()/isIntersection() → false, no members), so the rule sees "no banned
// type" and degrades silently rather than throwing. Mirrors how real TS hands
// back an error/any type for an un-typed node. Shared singleton — its identity is
// irrelevant since it is never a real interned type a rule compares against.
const FALLBACK_TYPE = {
  id: 0,
  flags: 0,
  isUnion: () => false,
  isIntersection: () => false,
  get types(): unknown[] {
    return [];
  },
  getCallSignatures: (): unknown[] => [],
  __block: undefined as DecodedBlock | undefined,
};

/**
 * Build the `parserServices` object injected into SourceCode. `raw` is the
 * snapshot as a transferred ArrayBuffer (the host decoded the wire base64 once
 * and transferList'd ownership in; the worker reads it in place). Returns `{}`
 * when the snapshot is absent or malformed so rules degrade to "no type info"
 * silently.
 */
export function buildParserServicesFromSnapshot(
  raw: unknown,
): Record<string, unknown> {
  if (!(raw instanceof ArrayBuffer) || raw.byteLength < HEADER_SIZE) return {};
  const dv = new DataView(raw);
  if (
    dv.getUint32(0, true) !== SNAPSHOT_MAGIC ||
    dv.getUint32(4, true) !== SNAPSHOT_VERSION
  ) {
    return {};
  }
  const n2tCount = dv.getUint32(8, true);
  const typCount = dv.getUint32(12, true);
  const primUndefined = dv.getInt32(16, true);
  const primNull = dv.getInt32(20, true);

  const n2tStart = HEADER_SIZE;
  const typIdxStart = n2tStart + n2tCount * N2T_ENTRY_SIZE;
  const blobStart = typIdxStart + typCount * IDX_ENTRY_SIZE;
  if (blobStart > dv.byteLength) return {};

  const typeCache = new Map<number, unknown>();

  // Binary-search node2type (sorted by (tokenStart, end)) for a node's typeId
  // (0 = the nil type-id, never a real entry; also the miss sentinel).
  function lookupNode(start: number, end: number): number {
    let lo = 0;
    let hi = n2tCount - 1;
    while (lo <= hi) {
      const mid = (lo + hi) >>> 1;
      const off = n2tStart + mid * N2T_ENTRY_SIZE;
      const ts = dv.getInt32(off, true);
      const e = dv.getInt32(off + 4, true);
      if (ts < start || (ts === start && e < end)) lo = mid + 1;
      else if (ts > start || (ts === start && e > end)) hi = mid - 1;
      else return dv.getInt32(off + 8, true);
    }
    return 0;
  }

  // Binary-search the types index (sorted by id) → [absolute blob start, blob
  // len], or null. The len bounds the lazy block decode below.
  function findBlob(id: number): [number, number] | null {
    let lo = 0;
    let hi = typCount - 1;
    while (lo <= hi) {
      const mid = (lo + hi) >>> 1;
      const off = typIdxStart + mid * IDX_ENTRY_SIZE;
      const tid = dv.getInt32(off, true);
      if (tid < id) lo = mid + 1;
      else if (tid > id) hi = mid - 1;
      else
        return [
          blobStart + dv.getUint32(off + 4, true),
          dv.getUint32(off + 8, true),
        ];
    }
    return null;
  }

  // Decode one type block within [start, start+len), clamped to the buffer end.
  // Every read is bounds-checked: a corrupted/truncated block (a count field
  // claiming more data than present) returns undefined → a flagless type
  // (silent degrade) instead of throwing RangeError mid-rule. Mirrors Go
  // DecodeBinary's per-section checks so both ends fail closed on bad bytes.
  function decodeBlock(start: number, len: number): DecodedBlock | undefined {
    const end = Math.min(start + len, dv.byteLength);
    let q = start;
    if (q + 5 > end) return undefined;
    const flags = dv.getInt32(q, true);
    const bits = dv.getUint8(q + 4);
    q += 5;
    if (q + 4 > end) return undefined;
    const nameLen = dv.getUint32(q, true);
    q += 4 + nameLen; // name is in the blob but not decoded (see header doc)
    if (q > end) return undefined;
    const readIds = (): number[] | null => {
      if (q + 4 > end) return null;
      const n = dv.getUint32(q, true);
      q += 4;
      if (q + n * 4 > end) return null;
      const ids = new Array<number>(n);
      for (let i = 0; i < n; i++) {
        ids[i] = dv.getInt32(q, true);
        q += 4;
      }
      return ids;
    };
    const memberTypes = readIds();
    const typeArgs = readIds();
    const callSigReturns = readIds();
    if (memberTypes === null || typeArgs === null || callSigReturns === null)
      return undefined;
    return {
      flags,
      isArray: (bits & 1) !== 0,
      isTuple: (bits & 2) !== 0,
      memberTypes,
      typeArgs,
      callSigReturns,
    };
  }

  function internType(id: number): unknown {
    if (!id) return undefined;
    const hit = typeCache.get(id);
    if (hit !== undefined) return hit;
    const loc = findBlob(id);
    const block = loc ? decodeBlock(loc[0], loc[1]) : undefined;
    const flags = block?.flags ?? 0;
    const t = {
      id,
      flags,
      isUnion: () => (flags & TSGO_TYPEFLAGS_UNION) !== 0,
      isIntersection: () => (flags & TSGO_TYPEFLAGS_INTERSECTION) !== 0,
      get types() {
        // Lazily resolve union/intersection members; filter(non-null) guards a
        // stray id-0 (Go drops them, but never let a rule iterate an undefined).
        return (block?.memberTypes ?? [])
          .map(internType)
          .filter((x) => x != null);
      },
      getCallSignatures: () =>
        (block?.callSigReturns ?? []).map((rt) => ({
          getReturnType: () => internType(rt),
        })),
      __block: block,
    };
    // Cache BEFORE any member recursion so a cyclic type (its own member)
    // resolves to this same wrapper instead of looping.
    typeCache.set(id, t);
    return t;
  }

  const checker = {
    getUndefinedType: () => internType(primUndefined),
    getNullType: () => internType(primNull),
    isArrayType: (t: { __block?: DecodedBlock } | undefined) =>
      !!t?.__block?.isArray,
    isTupleType: (t: { __block?: DecodedBlock } | undefined) =>
      !!t?.__block?.isTuple,
    getTypeArguments: (t: { __block?: DecodedBlock } | undefined) =>
      (t?.__block?.typeArgs ?? []).map(internType).filter((x) => x != null),
  };

  // Every Identifier's lookup end = range[0] + name.length (UTF-16 of oxc's
  // DECODED name); snapshot.go keys identifiers on the decoded name length too,
  // so both sides agree. This covers uniformly: a type-annotated binding whose
  // oxc range[1] runs to the annotation end (tsgo ends at the name), AND an
  // escaped identifier (e.g. `esc`) whose source token is longer than the
  // decoded name. Non-identifier nodes use range[1] (no name decoding). NOT
  // typeAnnotation.start (a space before the colon would misplace it). See
  // snapshot.go's span note.
  const resolveEnd = (node: {
    range: readonly [number, number];
    type?: string;
    name?: string;
  }): number =>
    node.type === 'Identifier' && typeof node.name === 'string'
      ? node.range[0] + node.name.length
      : node.range[1];

  const getTypeAtLocation = (node: {
    range: readonly [number, number];
    type?: string;
    name?: string;
    typeAnnotation?: unknown;
  }) => {
    const typeId = lookupNode(node.range[0], resolveEnd(node));
    return typeId ? internType(typeId) : FALLBACK_TYPE;
  };

  return {
    program: { getTypeChecker: () => checker },
    // getParserServices requires both maps to be non-null. The forward map's
    // get() returns a span-carrying proxy; the reverse map is unused by the
    // covered rules but must exist.
    esTreeNodeToTSNodeMap: {
      get: (n: { range: readonly [number, number] }) => ({
        __start: n.range[0],
        __end: n.range[1],
      }),
    },
    tsNodeToESTreeNodeMap: new Map(),
    getTypeAtLocation,
  };
}

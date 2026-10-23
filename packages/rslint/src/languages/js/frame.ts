/* rslint-disable @typescript-eslint/no-unsafe-type-assertion */
/** Decode final ESTree records into ordinary, worker-owned JavaScript objects. */
import { FRAME_SCHEMAS } from './frame-schema.js';
import { VISITOR_KEYS } from './visitor-keys.js';
import { buildLineStartOffsets, offsetToLineColumn } from './positions.js';
import type { ESTreeNode } from './types.js';

const HEADER = 40;
const MASK = 0x1fffffff;
const decoder = new TextDecoder('utf-8', { fatal: true, ignoreBOM: true });
const defineOwnProperty = Object.defineProperty;
// Structural fields use the canonical visitor order. Wire node IDs already
// provide identity; materialization needs no second object-to-ID index.
const CHILD_FIELDS = FRAME_SCHEMAS.map(([type, fields]) =>
  VISITOR_KEYS[type].map((key) => fields.indexOf(key)),
);

type FrameNode = ESTreeNode & Record<string, unknown>;
type NodeConstructor = new (
  start: number,
  end: number,
  lineStarts: number[],
  view: DataView,
  base: number,
) => FrameNode;

// A constructor per schema lets the runtime reserve a node's property storage
// together. Only fields present on the wire get slots; optional fields stay
// absent. These constructors capture schema constants, never a frame or AST.
function makeNodeConstructor(
  type: string,
  fields: readonly string[],
  fixed = false,
): NodeConstructor {
  const slot: PropertyDescriptor & { __proto__: null } = {
    __proto__: null,
    value: undefined,
    writable: true,
    enumerable: true,
    configurable: true,
  };
  const own = (node: FrameNode, key: string, value: unknown): void => {
    if (key in Object.prototype) {
      slot.value = value;
      defineOwnProperty(node, key, slot);
      slot.value = undefined;
    } else node[key] = value;
  };
  function Node(
    this: FrameNode,
    start: number,
    end: number,
    lineStarts: number[],
    view: DataView,
    base: number,
  ): void {
    own(this, 'type', type);
    own(this, 'start', start);
    own(this, 'end', end);
    own(this, 'range', [start, end]);
    own(this, 'loc', {
      start: offsetToLineColumn(start, lineStarts),
      end: offsetToLineColumn(end, lineStarts),
    });
    for (let index = 0; index < fields.length; index++) {
      if (
        fixed ||
        (fields[index][0] !== '$' &&
          view.getInt32(base + index * 4, true) >>> 29 !== 7)
      )
        own(this, fields[index], undefined);
    }
  }
  // Plugins receive ordinary objects with the usual Object prototype.
  Node.prototype = Object.prototype;
  return Node as unknown as NodeConstructor;
}

const NODE_CONSTRUCTORS = FRAME_SCHEMAS.map(([type, fields]) =>
  makeNodeConstructor(type, fields),
);

// Keep the instance layout stable when a schema has different present fields
// (for example JS and TS identifiers). Sharing one constructor across those
// shapes can retain unused in-object slots or move later fields out of line.
// Only schema keys are cached, never frame data. Unexpected combinations use
// the generic constructor after eight variants rather than growing a cache.
const NODE_VARIANTS = FRAME_SCHEMAS.map(
  () => new Map<number, NodeConstructor>(),
);

function constructorFor(
  typeID: number,
  view: DataView,
  base: number,
  parentSlot: boolean,
): NodeConstructor {
  const schema = FRAME_SCHEMAS[typeID];
  const type = schema[0],
    fields = schema[1];
  let mask = 0;
  for (let index = 0; index < fields.length; index++) {
    if (
      fields[index][0] !== '$' &&
      view.getInt32(base + index * 4, true) >>> 29 !== 7
    )
      mask |= 1 << index;
  }
  const extra =
    (type === 'Literal' &&
      !('regex' in Object.prototype) &&
      view.getInt32(base + fields.indexOf('$regexPattern') * 4, true) >>> 29 ===
        3 &&
      view.getInt32(base + fields.indexOf('$regexFlags') * 4, true) >>> 29 ===
        3) ||
    (type === 'TemplateElement' && !('value' in Object.prototype));
  if (parentSlot) mask |= 1 << fields.length;
  if (extra) mask |= 1 << (fields.length + 1);
  const variants = NODE_VARIANTS[typeID];
  const cached = variants.get(mask);
  if (cached) return cached;
  if (variants.size === 8) return NODE_CONSTRUCTORS[typeID];
  const keys: string[] = [];
  for (let index = 0; index < fields.length; index++) {
    if (mask & (1 << index)) keys.push(fields[index]);
  }
  if (extra) keys.push(type === 'Literal' ? 'regex' : 'value');
  if (parentSlot) keys.push('parent');
  const constructor = makeNodeConstructor(type, keys, true);
  variants.set(mask, constructor);
  return constructor;
}

export class FrameCancelled extends Error {}

export function checkFrameCancellation(flag?: Int32Array): void {
  if (flag && Atomics.load(flag, 0) !== 0) throw new FrameCancelled();
}

export interface DecodedFrame {
  ast: ESTreeNode;
  text: string;
  /** Original byte positions when source decoding replaces invalid bytes. */
  utf16ToByte?: number[];
  hasBOM: boolean;
  lineStarts: number[];
  tokenTypes: Uint8Array;
  tokenStarts: Uint32Array;
  tokenEnds: Uint32Array;
  comments: Array<{
    type: 'Line' | 'Block';
    value: string;
    start: number;
    end: number;
  }>;
}

function decodeCompilerSource(
  bytes: Uint8Array,
  cancelFlag?: Int32Array,
): Pick<DecodedFrame, 'text' | 'utf16ToByte'> {
  // Match the compiler's code units: each invalid byte becomes U+FFFD,
  // while WTF-8 preserves lone surrogates. Keep original byte positions;
  // re-encoding replacement characters would shift diagnostics and edits.
  const utf16ToByte: number[] = [],
    units: number[] = [],
    chunks: string[] = [];
  const continuation = (value: number): boolean =>
    value >= 0x80 && value <= 0xbf;
  let checkAt = 0;
  for (let offset = 0; offset < bytes.length;) {
    if (offset >= checkAt) {
      checkFrameCancellation(cancelFlag);
      checkAt = offset + 4096;
    }
    const first = bytes[offset],
      second = bytes[offset + 1],
      third = bytes[offset + 2],
      fourth = bytes[offset + 3];
    let code = first,
      width = 1;
    if (first >= 0xc2 && first <= 0xdf && continuation(second)) {
      code = ((first & 0x1f) << 6) | (second & 0x3f);
      width = 2;
    } else if (
      first >= 0xe0 &&
      first <= 0xef &&
      continuation(second) &&
      continuation(third) &&
      (first !== 0xe0 || second >= 0xa0)
    ) {
      code = ((first & 0xf) << 12) | ((second & 0x3f) << 6) | (third & 0x3f);
      width = 3;
    } else if (
      first >= 0xf0 &&
      first <= 0xf4 &&
      continuation(second) &&
      continuation(third) &&
      continuation(fourth) &&
      (first !== 0xf0 || second >= 0x90) &&
      (first !== 0xf4 || second <= 0x8f)
    ) {
      code =
        ((first & 7) << 18) |
        ((second & 0x3f) << 12) |
        ((third & 0x3f) << 6) |
        (fourth & 0x3f);
      width = 4;
    } else if (first >= 0x80) code = 0xfffd;
    utf16ToByte.push(offset);
    if (code > 0xffff) {
      units.push(0xd800 + ((code - 0x10000) >> 10));
      utf16ToByte.push(offset);
      units.push(0xdc00 + ((code - 0x10000) & 0x3ff));
    } else units.push(code);
    offset += width;
    if (units.length >= 4096) {
      chunks.push(String.fromCharCode(...units));
      units.length = 0;
    }
  }
  chunks.push(String.fromCharCode(...units));
  utf16ToByte.push(bytes.length);
  return { text: chunks.join(''), utf16ToByte };
}

export function decodeFrame(
  bytes: Uint8Array,
  cancelFlag?: Int32Array,
): DecodedFrame {
  checkFrameCancellation(cancelFlag);
  const invalid = (): never => {
    throw new Error('invalid ESTree frame');
  };
  if (bytes.byteLength < HEADER) invalid();
  const view = new DataView(bytes.buffer, bytes.byteOffset, bytes.byteLength);
  const u32 = (offset: number): number => view.getUint32(offset, true);
  const flags = u32(0),
    root = u32(4),
    count = u32(8),
    valueCount = u32(12);
  const listWords = u32(16),
    stringCount = u32(20),
    stringBytes = u32(24),
    numberCount = u32(28);
  const tokenCount = u32(32),
    commentCount = u32(36);
  // Derive every section in Number arithmetic before allocating from counts.
  // Each count is u32, so these sums cannot overflow JavaScript's safe integers.
  const align4 = (n: number): number => Math.ceil(n / 4) * 4;
  const types = HEADER,
    starts = align4(types + count * 2),
    ends = starts + count * 4;
  const valueStarts = ends + count * 4,
    values = valueStarts + count * 4;
  const lists = values + valueCount * 4,
    strings = lists + listWords * 4;
  const numbers = strings + stringBytes,
    tokenTypesOffset = numbers + numberCount * 8;
  const tokenStartsOffset = align4(tokenTypesOffset + tokenCount),
    tokenEndsOffset = tokenStartsOffset + tokenCount * 4;
  const commentsOffset = tokenEndsOffset + tokenCount * 4,
    source = commentsOffset + commentCount * 12;
  // Reject an incomplete body before allocating from header counts. Record
  // shapes, references and ranges belong to the matching Go producer, not a
  // second validation pass here.
  if (source > bytes.byteLength || stringCount * 4 > stringBytes) invalid();

  const poll = (index: number): void => {
    if ((index & 1023) === 0) checkFrameCancellation(cancelFlag);
  };
  // String references are sequential IDs. A compact offset column and an
  // owned-string cache replace a per-file hash table of record boundaries.
  const stringOffsets = new Uint32Array(stringCount);
  const stringCache = new Array<string | undefined>(stringCount);
  Object.setPrototypeOf(stringCache, null);
  for (let offset = 0, index = 0; index < stringCount; index++) {
    poll(index);
    stringOffsets[index] = offset;
    const encoded = view.getInt32(strings + offset, true);
    offset = encoded < 0 ? offset + 8 : align4(offset + 4 + encoded * 2);
  }
  let text: string, utf16ToByte: number[] | undefined;
  try {
    text = decoder.decode(bytes.subarray(source));
  } catch {
    ({ text, utf16ToByte } = decodeCompilerSource(
      bytes.subarray(source),
      cancelFlag,
    ));
  }
  const lineStarts = buildLineStartOffsets(text);
  const typeAt = (index: number): number =>
    view.getUint16(types + index * 2, true);
  const reserveParents =
    !('parent' in Object.prototype) &&
    !('regex' in Object.prototype) &&
    !('value' in Object.prototype);
  // Node IDs are private storage, never inherited array properties.
  const nodes: FrameNode[] = Object.setPrototypeOf(new Array(count + 1), null);
  const stringData = Buffer.from(
    bytes.buffer,
    bytes.byteOffset,
    bytes.byteLength,
  );
  const readString = (id: number): string => {
    const cached = stringCache[id];
    if (cached !== undefined) return cached;
    const base = strings + stringOffsets[id],
      encoded = view.getInt32(base, true);
    let value = '';
    if (encoded < 0) {
      const length = ~encoded;
      const start = source + u32(base + 4),
        end = start + length;
      // Decode source spans into owned strings, never slices of the full text
      // or views of transport bytes. Chunk boundaries stay on UTF-8 scalars.
      for (let offset = start; offset < end;) {
        checkFrameCancellation(cancelFlag);
        let next = Math.min(end, offset + 8192);
        while (next < end && (bytes[next] & 0xc0) === 0x80) next--;
        value += stringData.toString('utf8', offset, next);
        offset = next;
      }
      stringCache[id] = value;
      return value;
    }
    const length = encoded;
    // Decode UTF-16 code units into owned JS strings, including unpaired
    // surrogates. Chunking keeps cancellation observable for large literals.
    for (let start = 0; start < length; start += 4096) {
      checkFrameCancellation(cancelFlag);
      value += stringData.toString(
        'utf16le',
        base + 4 + start * 2,
        base + 4 + Math.min(length, start + 4096) * 2,
      );
    }
    stringCache[id] = value;
    return value;
  };
  for (let id = 1; id <= count; id++) {
    poll(id);
    const start = u32(starts + (id - 1) * 4),
      end = u32(ends + (id - 1) * 4);
    const base = values + u32(valueStarts + (id - 1) * 4) * 4;
    const parentSlot = reserveParents && id !== root;
    const Constructor = constructorFor(typeAt(id - 1), view, base, parentSlot);
    nodes[id] = new Constructor(start, end, lineStarts, view, base);
  }
  const readValue = (encoded: number): unknown => {
    const payload = encoded & MASK;
    switch (encoded >>> 29) {
      case 0:
        return null;
      case 1:
        return nodes[payload];
      case 2: {
        const base = lists + payload * 4,
          length = u32(base),
          result: Array<ESTreeNode | null> = Object.setPrototypeOf(
            new Array(length),
            null,
          );
        for (let i = 0; i < length; i++) {
          poll(i);
          const id = u32(base + 4 + i * 4);
          result[i] = id === 0 ? null : nodes[id];
        }
        // ESTree lists become ordinary arrays after every own entry exists.
        Object.setPrototypeOf(result, Array.prototype);
        return result;
      }
      case 3:
        return readString(payload);
      case 4:
        return false;
      case 5:
        return true;
      case 6:
        return view.getFloat64(numbers + payload * 8, true);
      default:
        return undefined;
    }
  };
  for (let id = 1; id <= count; id++) {
    poll(id);
    const node = nodes[id],
      fields = FRAME_SCHEMAS[typeAt(id - 1)][1],
      base = u32(valueStarts + (id - 1) * 4);
    let pattern: unknown, flags: unknown, cooked: unknown, raw: unknown;
    for (let i = 0; i < fields.length; i++) {
      const value = readValue(view.getInt32(values + (base + i) * 4, true));
      switch (fields[i]) {
        case '$regexPattern':
          pattern = value;
          break;
        case '$regexFlags':
          flags = value;
          break;
        case '$cooked':
          cooked = value;
          break;
        case '$raw':
          raw = value;
          break;
        default:
          if (value !== undefined) node[fields[i]] = value;
      }
    }
    if (node.type === 'Literal') {
      if (typeof pattern === 'string' && typeof flags === 'string') {
        node.regex = { pattern, flags };
        try {
          node.value = new RegExp(pattern, flags);
        } catch {
          node.value = null;
        }
      }
      if (typeof node.bigint === 'string') node.value = BigInt(node.bigint);
    } else if (node.type === 'TemplateElement') {
      node.value = { cooked, raw };
    }
  }
  // Parent setters observe complete nodes, in the original traversal order.
  const stack = [root];
  let visited = 0;
  while (stack.length) {
    poll(visited++);
    const id = stack.pop()!;
    const node = nodes[id];
    const base = u32(valueStarts + (id - 1) * 4);
    // Walk the private schema index without a per-node iterator.
    const childFields = CHILD_FIELDS[typeAt(id - 1)];
    for (let childIndex = 0; childIndex < childFields.length; childIndex++) {
      const field = childFields[childIndex];
      const encoded = view.getInt32(values + (base + field) * 4, true),
        tag = encoded >>> 29,
        payload = encoded & MASK;
      if (tag === 0 || tag === 7) continue;
      if (tag === 1) {
        nodes[payload].parent = node;
        stack.push(payload);
      } else if (tag === 2) {
        const list = lists + payload * 4,
          length = u32(list);
        for (let index = 0; index < length; index++) {
          poll(index);
          const childID = u32(list + (index + 1) * 4);
          if (childID === 0) continue;
          nodes[childID].parent = node;
          stack.push(childID);
        }
      }
    }
  }
  const tokenTypes = new Uint8Array(tokenCount),
    tokenStarts = new Uint32Array(tokenCount),
    tokenEnds = new Uint32Array(tokenCount);
  for (let index = 0; index < tokenCount; index++) {
    poll(index);
    const type = bytes[tokenTypesOffset + index],
      start = u32(tokenStartsOffset + index * 4),
      end = u32(tokenEndsOffset + index * 4);
    tokenTypes[index] = type;
    tokenStarts[index] = start;
    tokenEnds[index] = end;
  }
  const comments: DecodedFrame['comments'] = [];
  for (let index = 0; index < commentCount; index++) {
    poll(index);
    const base = commentsOffset + index * 12,
      kind = u32(base),
      start = u32(base + 4),
      end = u32(base + 8);
    comments.push({
      type: kind ? 'Block' : 'Line',
      value: text.slice(start + 2, kind ? end - 2 : end),
      start,
      end,
    });
  }
  checkFrameCancellation(cancelFlag);
  return {
    ast: nodes[root],
    text,
    utf16ToByte,
    hasBOM: flags === 1,
    lineStarts,
    tokenTypes,
    tokenStarts,
    tokenEnds,
    comments,
  };
}

/* rslint-disable @typescript-eslint/no-unsafe-type-assertion */
/** Decode final ESTree records into ordinary, worker-owned JavaScript objects. */
import { FRAME_SCHEMAS, FRAME_SCHEMA_HASH } from './frame-schema.js';
import { VISITOR_KEYS } from './visitor-keys.js';
import { buildLineStartOffsets, offsetToLineColumn } from './positions.js';
import type { ESTreeNode } from './types.js';

const HEADER = 64;
const MASK = 0x1fffffff;
const decoder = new TextDecoder('utf-8', { fatal: true, ignoreBOM: true });
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
const NODE_CONSTRUCTORS = FRAME_SCHEMAS.map(([type, fields]) => {
  const slot = Object.assign(Object.create(null) as PropertyDescriptor, {
    value: undefined,
    writable: true,
    enumerable: true,
    configurable: true,
  });
  const defineProperty = Object.defineProperty;
  const own = (node: FrameNode, key: string, value: unknown): void => {
    if (key in Object.prototype) {
      slot.value = value;
      defineProperty(node, key, slot);
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
        fields[index][0] !== '$' &&
        view.getInt32(base + index * 4, true) >>> 29 !== 7
      )
        own(this, fields[index], undefined);
    }
  }
  // Plugins receive ordinary objects with the usual Object prototype.
  Node.prototype = Object.prototype;
  return Node as unknown as NodeConstructor;
});

export class FrameCancelled extends Error {}

export function checkFrameCancellation(flag?: Int32Array): void {
  if (flag && Atomics.load(flag, 0) !== 0) throw new FrameCancelled();
}

export interface DecodedFrame {
  ast: ESTreeNode;
  text: string;
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

export function decodeFrame(
  bytes: Uint8Array,
  cancelFlag?: Int32Array,
): DecodedFrame {
  checkFrameCancellation(cancelFlag);
  const invalid = (): never => {
    throw new Error('invalid ESTree frame');
  };
  if (
    !(bytes instanceof Uint8Array) ||
    bytes.buffer instanceof SharedArrayBuffer ||
    bytes.length < HEADER ||
    bytes.length > 64 * 1024 * 1024
  )
    invalid();
  const view = new DataView(bytes.buffer, bytes.byteOffset, bytes.byteLength);
  const u32 = (offset: number): number => view.getUint32(offset, true);
  if (u32(0) !== 0x31465445 || u32(4) !== 2 || u32(8) !== FRAME_SCHEMA_HASH) {
    throw new Error('ESTree frame version or schema mismatch');
  }
  const flags = u32(12),
    root = u32(20),
    count = u32(24),
    valueCount = u32(28);
  const listWords = u32(32),
    stringBytes = u32(36),
    numberCount = u32(40);
  const tokenCount = u32(44),
    commentCount = u32(48),
    sourceBytes = u32(52),
    sourceLength = u32(56);
  if (
    u32(16) !== bytes.length ||
    flags > 1 ||
    !root ||
    root > count ||
    stringBytes % 4
  )
    invalid();
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
  if (source !== u32(60) || source + sourceBytes !== bytes.length) invalid();

  const poll = (index: number): void => {
    if ((index & 1023) === 0) checkFrameCancellation(cancelFlag);
  };
  // One byte per wire word marks record boundaries without boxed Set entries.
  const listOffsets = new Uint8Array(listWords);

  for (let offset = 0, index = 0; offset < listWords; index++) {
    poll(index);
    listOffsets[offset] = 1;
    const length = u32(lists + offset * 4);
    if (length > listWords - offset - 1) invalid();
    for (let i = 0; i < length; i++) {
      poll(i);
      if (u32(lists + (offset + 1 + i) * 4) > count) invalid();
    }
    offset += length + 1;
  }
  let stringCount = 0;
  for (let offset = 0; offset < stringBytes; stringCount++) {
    poll(stringCount);
    const length = u32(strings + offset);
    const next = align4(offset + 4 + length * 2);
    if (next > stringBytes) invalid();
    offset = next;
  }
  // String references are sequential IDs. A compact offset column and an
  // owned-string cache replace a per-file hash table of record boundaries.
  const stringOffsets = new Uint32Array(stringCount);
  const stringCache = new Array<string | undefined>(stringCount);
  Object.setPrototypeOf(stringCache, null);
  for (let offset = 0, index = 0; index < stringCount; index++) {
    poll(index);
    stringOffsets[index] = offset;
    offset = align4(offset + 4 + u32(strings + offset) * 2);
  }
  const text = decoder.decode(bytes.subarray(source));
  if (text.length !== sourceLength) invalid();
  const lineStarts = buildLineStartOffsets(text);
  const typeAt = (index: number): number =>
    view.getUint16(types + index * 2, true);
  const validValue = (encoded: number): void => {
    const tag = encoded >>> 29,
      payload = encoded & MASK;
    if (tag === 1) {
      if (!payload || payload > count) invalid();
    } else if (tag === 2) {
      if (listOffsets[payload] !== 1) invalid();
    } else if (tag === 3) {
      if (payload >= stringCount) invalid();
    } else if (tag === 6) {
      if (payload >= numberCount) invalid();
    } else if (payload !== 0) invalid();
  };
  let nextValue = 0;
  for (let index = 0; index < count; index++) {
    poll(index);
    const schema = FRAME_SCHEMAS[typeAt(index)];
    if (
      !schema ||
      !VISITOR_KEYS[schema[0]] ||
      u32(valueStarts + index * 4) !== nextValue
    )
      invalid();
    if (
      u32(starts + index * 4) > u32(ends + index * 4) ||
      u32(ends + index * 4) > sourceLength
    )
      invalid();
    nextValue += schema[1].length;
    if (nextValue > valueCount) invalid();
  }
  if (
    nextValue !== valueCount ||
    FRAME_SCHEMAS[typeAt(root - 1)][0] !== 'Program'
  )
    invalid();
  for (let index = 0; index < valueCount; index++) {
    poll(index);
    // The high bits are tags, not an unsigned magnitude. Signed reads keep
    // boolean/undefined words in the integer representation; >>> restores
    // the exact same tag bits and MASK extracts the unchanged payload.
    validValue(view.getInt32(values + index * 4, true));
  }

  const nodes: FrameNode[] = new Array(count + 1);
  const stringData = Buffer.from(
    bytes.buffer,
    bytes.byteOffset + strings,
    stringBytes,
  );
  const readString = (id: number): string => {
    const cached = stringCache[id];
    if (cached !== undefined) return cached;
    const base = stringOffsets[id],
      length = u32(strings + base);
    let value = '';
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
    nodes[id] = new NODE_CONSTRUCTORS[typeAt(id - 1)](
      start,
      end,
      lineStarts,
      view,
      values + u32(valueStarts + (id - 1) * 4) * 4,
    );
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
          result = new Array<ESTreeNode | null>(length);
        for (let i = 0; i < length; i++) {
          poll(i);
          const id = u32(base + 4 + i * 4);
          result[i] = id === 0 ? null : nodes[id];
        }
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
      case 7:
        return undefined;
      default:
        return invalid();
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
  if (
    nodes[root].sourceType !== 'module' &&
    nodes[root].sourceType !== 'script'
  )
    invalid();
  if (!Array.isArray(nodes[root].body)) invalid();
  // Validate traversal and assign parents iteratively. Plugins may retain any
  // node after lint returns; no mapped bytes or transport capability escapes.
  const state = new Uint8Array(count + 1);
  const stack = [root];
  let visited = 0;
  while (stack.length) {
    poll(visited++);
    const id = stack.pop()!;
    if (id < 0) {
      state[-id] = 2;
      continue;
    }
    if (state[id] !== 0) invalid(); // cycles and shared structural children
    state[id] = 1;
    stack.push(-id);
    const node = nodes[id];
    const base = u32(valueStarts + (id - 1) * 4);
    // Walk the private schema index without a per-node iterator.
    const childFields = CHILD_FIELDS[typeAt(id - 1)];
    for (let childIndex = 0; childIndex < childFields.length; childIndex++) {
      const field = childFields[childIndex];
      if (field < 0) invalid();
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
      } else {
        invalid();
      }
    }
  }
  const tokenTypes = new Uint8Array(tokenCount),
    tokenStarts = new Uint32Array(tokenCount),
    tokenEnds = new Uint32Array(tokenCount);
  let previousEnd = 0;
  for (let index = 0; index < tokenCount; index++) {
    poll(index);
    const type = bytes[tokenTypesOffset + index],
      start = u32(tokenStartsOffset + index * 4),
      end = u32(tokenEndsOffset + index * 4);
    if (type > 12 || start < previousEnd || start > end || end > sourceLength)
      invalid();
    tokenTypes[index] = type;
    tokenStarts[index] = start;
    tokenEnds[index] = end;
    previousEnd = end;
  }
  const comments: DecodedFrame['comments'] = [];
  previousEnd = 0;
  for (let index = 0; index < commentCount; index++) {
    poll(index);
    const base = commentsOffset + index * 12,
      kind = u32(base),
      start = u32(base + 4),
      end = u32(base + 8);
    if (
      kind > 1 ||
      start < previousEnd ||
      end - start < (kind ? 4 : 2) ||
      end > sourceLength
    )
      invalid();
    comments.push({
      type: kind ? 'Block' : 'Line',
      value: text.slice(start + 2, kind ? end - 2 : end),
      start,
      end,
    });
    previousEnd = end;
  }
  checkFrameCancellation(cancelFlag);
  return {
    ast: nodes[root],
    text,
    hasBOM: flags === 1,
    lineStarts,
    tokenTypes,
    tokenStarts,
    tokenEnds,
    comments,
  };
}

/* rslint-disable @typescript-eslint/no-unsafe-type-assertion -- Indexed generic scalar return types follow the checked packet kind. */
import { SCALAR_KINDS } from './type-service-protocol.js';
import type { JSDocTagInfo, SymbolDisplayPart } from 'typescript';
import type {
  NodeRef,
  NodeFact,
  TypeFact,
  SymbolFact,
  SignatureFact,
  IndexInfoFact,
  TypePredicateFact,
  NamedSymbolFact,
  DiagnosticFact,
  ScalarOperation,
  QueryResults,
} from './type-service-protocol.js';

const invalid = (): never => {
  throw new Error('invalid type service packet');
};

export const isRecordBundle = (bytes: Readonly<Uint8Array>): boolean =>
  bytes.byteLength >= 12 &&
  bytes[0] === 0x52 &&
  bytes[1] === 0x54 &&
  bytes[2] === 0x42 &&
  bytes[3] === 0x31;

/** Returns false when Go stamped bit 31 on an RTS5 packet's kind word to mark
 *  a mutable fact (transient merged symbol or context-sensitive signature). */
export const isPacketCacheable = (bytes: Readonly<Uint8Array>): boolean => {
  if (
    bytes.byteLength >= 8 &&
    bytes[0] === 0x52 &&
    bytes[1] === 0x54 &&
    bytes[2] === 0x53 &&
    bytes[3] === 0x35
  ) {
    return (bytes[7] & 0x80) === 0;
  }
  return true;
};

/**
 * Decodes shallow binary records in one pass. Only this disposable reader owns
 * bytes; semantic wrappers retain decoded values, never a reader or byte view.
 */
export class TypePacketReader {
  private bytes: Buffer | undefined;
  private readonly strings = new Map<number, string>();
  readonly kind: number;

  constructor(bytes: Readonly<Uint8Array>) {
    this.bytes = Buffer.from(bytes.buffer, bytes.byteOffset, bytes.byteLength);
    if (bytes.length < 12 || this.word(0) !== 0x35535452) invalid();
    this.kind = this.word(4) & 0x7fff_ffff;
  }

  get retainedBytes(): number {
    return this.bytes?.byteLength ?? 0;
  }

  close(): void {
    this.bytes = undefined;
    this.strings.clear();
  }

  private buf(): Buffer {
    if (!this.bytes) throw new Error('type packet reader is closed');
    return this.bytes;
  }

  private word(offset: number): number {
    const bytes = this.buf();
    if (offset < 0 || offset + 4 > bytes.length) return invalid();
    return bytes.readUInt32LE(offset);
  }

  private text(offset: number): string {
    const old = this.strings.get(offset);
    if (old !== undefined) return old;
    const bytes = this.buf();
    const size = this.word(offset);
    const start = offset + 4;
    const end = start + size * 2;
    if (end > bytes.length) return invalid();
    const value = bytes.toString('utf16le', start, end);
    this.strings.set(offset, value);
    return value;
  }

  private ids(offset: number): number[] {
    const bytes = this.buf();
    const size = this.word(offset);
    if (offset + 4 + size * 4 > bytes.length) return invalid();
    return Array.from({ length: size }, (_, i) =>
      this.word(offset + 4 + i * 4),
    );
  }

  private optionalIDs(at: number): number[] | undefined {
    const offset = this.word(at);
    return offset ? this.ids(offset) : undefined;
  }

  private node(offset: number): NodeRef | undefined {
    if (!offset) return undefined;
    return {
      file: this.text(this.word(offset)),
      index: this.word(offset + 4),
    };
  }

  private requiredNode(offset: number): NodeRef {
    return this.node(offset) ?? invalid();
  }

  nodeFacts(): readonly NodeFact[] {
    if (this.kind !== 8) return invalid();
    const bytes = this.buf();
    const at = this.word(8);
    const count = this.word(at);
    if (at + 4 + count * 28 > bytes.length) return invalid();
    return Array.from({ length: count }, (_, i) => {
      const pos = at + 4 + i * 28;
      const text = this.word(pos + 20);
      return {
        node: this.requiredNode(this.word(pos)),
        kind: this.text(this.word(pos + 4)),
        flags: this.word(pos + 8),
        moduleKeyword: this.word(pos + 24),
        parent: this.node(this.word(pos + 12)),
        name: this.node(this.word(pos + 16)),
        text: text ? this.text(text) : undefined,
      };
    });
  }

  typeFact(): TypeFact {
    if (this.kind !== 5) return invalid();
    const intrinsicAt = this.word(20);
    const escapedAt = this.word(36);
    const objectFlagsAt = this.word(40);
    const tupleAt = this.word(56);
    const tag = this.word(24);
    const valueAt = this.word(28);
    let value: TypeFact['value'];
    if (tag === 0) {
      value = undefined;
    } else if (tag === 1) {
      value = this.text(valueAt);
    } else if (tag === 2) {
      const bytes = this.buf();
      if (valueAt + 8 > bytes.length) return invalid();
      value = bytes.readDoubleLE(valueAt);
    } else if (tag === 3) {
      value = {
        negative: this.word(valueAt) !== 0,
        base10Value: this.text(this.word(valueAt + 4)),
      };
    } else {
      return invalid();
    }
    return {
      flags: this.ids(this.word(8)).map((offset) => this.text(offset)),
      symbol: this.word(12),
      aliasSymbol: this.word(16),
      intrinsicName: intrinsicAt ? this.text(intrinsicAt) : undefined,
      escapedName: escapedAt ? this.text(escapedAt) : undefined,
      objectFlags: objectFlagsAt
        ? this.ids(objectFlagsAt).map((offset) => this.text(offset))
        : undefined,
      value,
      typeOperands: this.optionalIDs(32),
      target: this.word(44),
      aliasTypeArguments: this.optionalIDs(48),
      typeParameters: this.optionalIDs(52),
      tuple: tupleAt
        ? {
            combinedFlags: this.word(tupleAt),
            readonly: this.word(tupleAt + 4) !== 0,
          }
        : undefined,
    };
  }

  symbolFact(): SymbolFact {
    if (this.kind !== 6) return invalid();
    const checkFlagsRaw = this.word(28);
    const declsAt = this.word(20);
    let declarations: NodeRef[] | undefined;
    if (declsAt) {
      const bytes = this.buf();
      const size = this.word(declsAt);
      if (declsAt + 4 + size * 8 > bytes.length) return invalid();
      declarations = Array.from({ length: size }, (_, i) =>
        this.requiredNode(declsAt + 4 + i * 8),
      );
    }
    return {
      name: this.text(this.word(8)),
      escapedName: this.text(this.word(12)),
      flags: this.word(16),
      checkFlags: checkFlagsRaw ? checkFlagsRaw - 1 : undefined,
      declarations,
      valueDeclaration: this.node(this.word(24)),
    };
  }

  signatureFact(): SignatureFact {
    if (this.kind !== 7) return invalid();
    return {
      parameters: this.ids(this.word(8)),
      declaration: this.node(this.word(12)),
      typeParameters: this.optionalIDs(16),
      thisParameter: this.word(20),
      minArgumentCount: this.word(24),
    };
  }

  private indexInfos(at: number): readonly IndexInfoFact[] {
    const bytes = this.buf();
    const count = this.word(at);
    if (at + 4 + count * 20 > bytes.length) return invalid();
    return Array.from({ length: count }, (_, i) => {
      const pos = at + 4 + i * 20;
      return {
        id: this.word(pos),
        keyType: this.word(pos + 4),
        type: this.word(pos + 8),
        isReadonly: this.word(pos + 12) !== 0,
        declaration: this.node(this.word(pos + 16)),
      };
    });
  }

  private typePredicate(): TypePredicateFact | undefined {
    const kind = this.word(8);
    const name = this.word(12);
    return kind
      ? {
          kind: kind - 1,
          parameterName: name ? this.text(name) : undefined,
          parameterIndex: name ? this.word(16) | 0 : undefined,
          type: this.word(20),
        }
      : undefined;
  }

  private namedSymbols(at: number): readonly NamedSymbolFact[] {
    const bytes = this.buf();
    const count = this.word(at);
    if (at + 4 + count * 8 > bytes.length) return invalid();
    return Array.from({ length: count }, (_, i) => {
      const pos = at + 4 + i * 8;
      return { id: this.word(pos), name: this.text(this.word(pos + 4)) };
    });
  }

  private displayParts(at: number): SymbolDisplayPart[] {
    if (!at) return [];
    const bytes = this.buf();
    const count = this.word(at);
    if (at + 4 + count * 12 > bytes.length) return invalid();
    return Array.from({ length: count }, (_, i) => {
      const pos = at + 4 + i * 12;
      const part = {
        text: this.text(this.word(pos)),
        kind: this.text(this.word(pos + 4)),
      };
      const target = this.word(pos + 8);
      return target
        ? {
            ...part,
            target: {
              fileName: this.text(this.word(target)),
              textSpan: {
                start: this.word(target + 4),
                length: this.word(target + 8),
              },
            },
          }
        : part;
    });
  }

  private jsDocTags(): JSDocTagInfo[] {
    const bytes = this.buf();
    const at = this.word(8);
    const count = this.word(at);
    if (at + 4 + count * 8 > bytes.length) return invalid();
    return Array.from({ length: count }, (_, i) => {
      const pos = at + 4 + i * 8;
      const text = this.word(pos + 4);
      return {
        name: this.text(this.word(pos)),
        text: text ? this.displayParts(text) : undefined,
      };
    });
  }

  private diagnostics(rootAt: number): readonly DiagnosticFact[] {
    if (!rootAt) return [];
    const bytes = this.buf();
    const seen = new Set<number>();
    type MutableDiagnosticFact = {
      -readonly [K in keyof DiagnosticFact]: DiagnosticFact[K];
    };
    interface PendingList {
      readonly fact: MutableDiagnosticFact;
      readonly field: 'chain' | 'related';
      readonly at: number;
      readonly minOffset: number;
    }
    const readList = (
      at: number,
      minOffset: number,
    ): {
      readonly facts: MutableDiagnosticFact[];
      readonly listEnd: number;
      readonly pending: readonly PendingList[];
    } => {
      if (at < minOffset || seen.has(at)) return invalid();
      seen.add(at);
      const count = this.word(at);
      const listEnd = at + 4 + count * 44;
      if (listEnd > bytes.length) return invalid();
      const facts: MutableDiagnosticFact[] = [];
      const pending: PendingList[] = [];
      for (let i = 0; i < count; i++) {
        const pos = at + 4 + i * 44;
        const file = this.word(pos);
        const source = this.word(pos + 24);
        const flags = this.word(pos + 28);
        const chainAt = this.word(pos + 32);
        const relatedAt = this.word(pos + 36);
        const filePath = this.word(pos + 40);
        const fact: MutableDiagnosticFact = {
          fileName: file ? this.text(file) : undefined,
          filePath: filePath ? this.text(filePath) : undefined,
          start: this.word(pos + 4),
          length: this.word(pos + 8),
          category: this.word(pos + 12),
          code: this.word(pos + 16),
          text: this.text(this.word(pos + 20)),
          source: source ? this.text(source) : undefined,
          reportsDeprecated: flags & 1 ? true : undefined,
          reportsUnnecessary: flags & 2 ? true : undefined,
          chain: [],
          related: [],
        };
        facts.push(fact);
        if (relatedAt) {
          pending.push({
            fact,
            field: 'related',
            at: relatedAt,
            minOffset: listEnd,
          });
        }
        if (chainAt) {
          pending.push({
            fact,
            field: 'chain',
            at: chainAt,
            minOffset: listEnd,
          });
        }
      }
      return { facts, listEnd, pending };
    };

    const root = readList(rootAt, 12);
    const stack: PendingList[] = [...root.pending];
    while (stack.length > 0) {
      const next = stack.pop();
      if (!next) continue;
      const child = readList(next.at, next.minOffset);
      next.fact[next.field] = child.facts;
      for (const item of child.pending) stack.push(item);
    }
    return root.facts;
  }

  scalar<K extends ScalarOperation>(operation: K): QueryResults[K] {
    if (this.kind !== SCALAR_KINDS[operation]) return invalid();
    let value: QueryResults[ScalarOperation];
    switch (this.kind) {
      case 1:
        value = this.word(8);
        break;
      case 2:
        value = this.word(8) !== 0;
        break;
      case 3:
        value = this.text(this.word(8));
        break;
      case 4:
        value = this.ids(this.word(8));
        break;
      case 9:
        value = this.indexInfos(this.word(8));
        break;
      case 10:
        value = this.typePredicate();
        break;
      case 11:
        value = this.namedSymbols(this.word(8));
        break;
      case 12:
        value = this.displayParts(this.word(8));
        break;
      case 13:
        value = this.jsDocTags();
        break;
      case 14: {
        const bytes = this.buf();
        const len = this.word(8);
        if (20 + len > bytes.length) return invalid();
        value = len
          ? {
              tree: Uint8Array.from(bytes.subarray(20, 20 + len)),
              origins: this.optionalIDs(12)?.map((offset) => this.node(offset)),
              emitFlags: this.optionalIDs(16),
            }
          : undefined;
        break;
      }
      case 15:
        value = this.diagnostics(this.word(8));
        break;
      case 16: {
        const tag = this.word(8);
        if (tag === 0) {
          value = undefined;
        } else if (tag === 1) {
          value = this.text(this.word(12));
        } else if (tag === 2) {
          const bytes = this.buf();
          if (20 > bytes.length) return invalid();
          value = bytes.readDoubleLE(12);
        } else {
          return invalid();
        }
        break;
      }
      case 17:
        value = this.ids(this.word(8)).map((at) => this.text(at));
        break;
      case 18: {
        const list = (offset: number) =>
          this.word(offset) ? this.ids(this.word(offset)) : undefined;
        const minPlusOne = this.word(20);
        const fixedPlusOne = this.word(24);
        value = {
          outer: list(8),
          local: list(12),
          elements: list(16),
          minLength: minPlusOne ? minPlusOne - 1 : undefined,
          fixedLength: fixedPlusOne ? fixedPlusOne - 1 : undefined,
          labels: list(28)?.map((at) => this.node(at)),
          texts: list(32)?.map((at) => this.text(at)),
          root: this.node(this.word(36)),
          rootCheck: this.word(40),
          rootExtends: this.word(44),
          rootAlias: this.word(48),
          distributive: this.word(52) !== 0,
          infer: list(56),
          rootOuter: list(60),
          rootArguments: list(64),
        };
        break;
      }
      case 19: {
        const at = this.word(8);
        value = at ? this.namedSymbols(at) : undefined;
        break;
      }
      case 20: {
        const at = this.word(8);
        value = at ? this.indexInfos(at) : undefined;
        break;
      }
      default:
        return invalid();
    }
    return value as QueryResults[K];
  }
}

export interface TypeReplyRecord {
  readonly op: 'typeRecord' | 'symbolRecord' | 'signatureRecord';
  readonly id: number;
  readonly packet: Uint8Array;
}

export interface TypeReply {
  readonly packet: Uint8Array;
  readonly records: readonly TypeReplyRecord[];
}

export function typeReplyPackets(bytes: Readonly<Uint8Array>): TypeReply {
  if (!isRecordBundle(bytes)) {
    return { packet: bytes, records: [] };
  }
  const view = new DataView(bytes.buffer, bytes.byteOffset, bytes.byteLength);
  const packetLength = view.getUint32(4, true);
  const count = view.getUint32(8, true);
  const packetStart = 12 + count * 16;
  const packetEnd = packetStart + packetLength;
  if (packetEnd > bytes.byteLength) return invalid();
  const records: TypeReplyRecord[] = [];
  for (let i = 0; i < count; i++) {
    const at = 12 + i * 16;
    const tag = view.getUint32(at, true);
    const id = view.getUint32(at + 4, true);
    const offset = view.getUint32(at + 8, true);
    const length = view.getUint32(at + 12, true);
    if (offset < packetEnd || offset + length > bytes.byteLength) {
      return invalid();
    }
    const op =
      tag === 1
        ? 'typeRecord'
        : tag === 2
          ? 'symbolRecord'
          : tag === 3
            ? 'signatureRecord'
            : invalid();
    records.push({
      op,
      id,
      packet: bytes.subarray(offset, offset + length),
    });
  }
  return {
    packet: bytes.subarray(packetStart, packetEnd),
    records,
  };
}

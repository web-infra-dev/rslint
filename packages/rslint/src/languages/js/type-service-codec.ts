/* rslint-disable @typescript-eslint/no-unsafe-type-assertion -- Indexed generic return types follow the finite field switch and checked packet kind. */
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

/** Offset-addressed fields stay lazy. Only this disposable reader owns bytes;
 * semantic wrappers may retain decoded values, never a reader or byte view. */
export class TypePacketReader {
  private bytes: Uint8Array | undefined;
  private view: DataView | undefined;
  private readonly strings = new Map<number, string>();
  readonly kind: number;
  constructor(bytes: Uint8Array) {
    this.bytes = bytes;
    this.view = new DataView(bytes.buffer, bytes.byteOffset, bytes.byteLength);
    if (bytes.length < 12 || this.word(0) !== 0x35535452) invalid();
    this.kind = this.word(4);
    const words = [
      0, 1, 1, 1, 1, 13, 7, 5, 1, 1, 4, 1, 1, 1, 1, 1, 3, 1, 1, 21, 5,
    ][this.kind];
    if (!words || bytes.length < 8 + words * 4) invalid();
  }
  get retainedBytes(): number {
    return this.bytes?.byteLength ?? 0;
  }
  close(): void {
    this.bytes = undefined;
    this.view = undefined;
    this.strings.clear();
  }
  private word(offset: number): number {
    const view = this.view;
    if (!view) throw new Error('type packet reader is closed');
    if (offset < 0 || offset + 4 > view.byteLength) return invalid();
    return view.getUint32(offset, true);
  }
  private text(offset: number): string {
    const old = this.strings.get(offset);
    if (old !== undefined) return old;
    const size = this.word(offset),
      view = this.view!;
    if (offset + 4 + size * 2 > view.byteLength) return invalid();
    const chunks: string[] = [];
    for (let i = 0; i < size; i += 4096) {
      const units: number[] = [];
      for (let j = i; j < Math.min(size, i + 4096); j++)
        units.push(view.getUint16(offset + 4 + j * 2, true));
      chunks.push(String.fromCharCode(...units));
    }
    const value = chunks.join('');
    this.strings.set(offset, value);
    return value;
  }
  private ids(offset: number): number[] {
    const size = this.word(offset);
    if (offset + 4 + size * 4 > this.view!.byteLength) return invalid();
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
    const file = this.text(this.word(offset)),
      index = this.word(offset + 4) | 0;
    return index < 0 ? { file, index: -index, config: true } : { file, index };
  }
  nodeFacts(): readonly NodeFact[] {
    if (this.kind !== 8) return invalid();
    const at = this.word(8),
      count = this.word(at);
    if (at + 4 + count * 28 > this.view!.byteLength) return invalid();
    return Array.from({ length: count }, (_, i) => {
      const pos = at + 4 + i * 28,
        text = this.word(pos + 20);
      return {
        node: this.node(this.word(pos))!,
        kind: this.text(this.word(pos + 4)),
        flags: this.word(pos + 8),
        moduleKeyword: this.word(pos + 24),
        parent: this.node(this.word(pos + 12)),
        name: this.node(this.word(pos + 16)),
        text: text ? this.text(text) : undefined,
      };
    });
  }
  typeField<K extends keyof TypeFact>(field: K): TypeFact[K] {
    if (this.kind !== 5) return invalid();
    let value: TypeFact[keyof TypeFact];
    switch (field) {
      case 'flags':
        value = this.ids(this.word(8)).map((offset) => this.text(offset));
        break;
      case 'symbol':
        value = this.word(12);
        break;
      case 'aliasSymbol':
        value = this.word(16);
        break;
      case 'intrinsicName': {
        const at = this.word(20);
        value = at ? this.text(at) : undefined;
        break;
      }
      case 'escapedName': {
        const at = this.word(36);
        value = at ? this.text(at) : undefined;
        break;
      }
      case 'objectFlags': {
        const at = this.word(40);
        value = at
          ? this.ids(at).map((offset) => this.text(offset))
          : undefined;
        break;
      }
      case 'typeOperands':
        value = this.optionalIDs(32);
        break;
      case 'target':
        value = this.word(44);
        break;
      case 'aliasTypeArguments':
        value = this.optionalIDs(48);
        break;
      case 'typeParameters':
        value = this.optionalIDs(52);
        break;
      case 'tuple': {
        const at = this.word(56);
        value = at
          ? {
              combinedFlags: this.word(at),
              readonly: this.word(at + 4) !== 0,
            }
          : undefined;
        break;
      }
      case 'value': {
        const tag = this.word(24),
          at = this.word(28);
        if (tag === 0) value = undefined;
        else if (tag === 1) value = this.text(at);
        else if (tag === 2) {
          if (at + 8 > this.view!.byteLength) return invalid();
          value = this.view!.getFloat64(at, true);
        } else if (tag === 3)
          value = {
            negative: this.word(at) !== 0,
            base10Value: this.text(this.word(at + 4)),
          };
        else return invalid();
        break;
      }
      default:
        return invalid();
    }
    return value as TypeFact[K];
  }
  symbolField<K extends keyof SymbolFact>(field: K): SymbolFact[K] {
    if (this.kind !== 6) return invalid();
    let value: SymbolFact[keyof SymbolFact];
    switch (field) {
      case 'name':
        value = this.text(this.word(8));
        break;
      case 'escapedName':
        value = this.text(this.word(12));
        break;
      case 'flags':
        value = this.word(16);
        break;
      case 'valueDeclaration':
        value = this.node(this.word(24));
        break;
      case 'checkFlags': {
        const flags = this.word(28);
        value = flags ? flags - 1 : undefined;
        break;
      }
      case 'hasMembers':
        value = (this.word(32) & 1) !== 0;
        break;
      case 'hasExports':
        value = (this.word(32) & 2) !== 0;
        break;
      case 'hasGlobalExports':
        value = (this.word(32) & 4) !== 0;
        break;
      case 'declarations': {
        const at = this.word(20);
        if (!at) value = undefined;
        else {
          const size = this.word(at);
          if (at + 4 + size * 8 > this.view!.byteLength) return invalid();
          value = Array.from({ length: size }, (_, i) =>
            this.node(at + 4 + i * 8)!,
          );
        }
        break;
      }
      default:
        return invalid();
    }
    return value as SymbolFact[K];
  }
  signatureField<K extends keyof SignatureFact>(field: K): SignatureFact[K] {
    if (this.kind !== 7) return invalid();
    let value: SignatureFact[keyof SignatureFact];
    switch (field) {
      case 'parameters':
        value = this.ids(this.word(8));
        break;
      case 'declaration':
        value = this.node(this.word(12));
        break;
      case 'typeParameters':
        value = this.optionalIDs(16);
        break;
      case 'thisParameter':
        value = this.word(20);
        break;
      case 'minArgumentCount':
        value = this.word(24);
        break;
    }
    return value as SignatureFact[K];
  }
  private indexInfos(at: number, idOffset: number): readonly IndexInfoFact[] {
    const count = this.word(at);
    const ids = idOffset ? this.ids(idOffset) : undefined;
    if (ids && ids.length !== count) return invalid();
    if (at + 4 + count * 20 > this.view!.byteLength) return invalid();
    return Array.from({ length: count }, (_, i) => {
      const pos = at + 4 + i * 20;
      return {
        id: ids?.[i],
        keyType: this.word(pos),
        type: this.word(pos + 4),
        isReadonly: this.word(pos + 8) !== 0,
        declaration: this.node(this.word(pos + 12)),
        components: this.word(pos + 16)
          ? this.ids(this.word(pos + 16)).map((at) => this.node(at)!)
          : undefined,
      };
    });
  }
  private typePredicate(): TypePredicateFact | undefined {
    const kind = this.word(8),
      name = this.word(12);
    return kind
      ? {
          kind: kind - 1,
          parameterName: name ? this.text(name) : undefined,
          parameterIndex: name ? this.word(16) | 0 : undefined,
          type: this.word(20),
        }
      : undefined;
  }
  private namedSymbols(): readonly NamedSymbolFact[] {
    const at = this.word(8),
      count = this.word(at);
    if (at + 4 + count * 8 > this.view!.byteLength) return invalid();
    return Array.from({ length: count }, (_, i) => {
      const pos = at + 4 + i * 8;
      return { id: this.word(pos), name: this.text(this.word(pos + 4)) };
    });
  }
  private displayParts(at: number): SymbolDisplayPart[] {
    if (!at) return [];
    const count = this.word(at);
    if (at + 4 + count * 12 > this.view!.byteLength) return invalid();
    return Array.from({ length: count }, (_, i) => {
      const pos = at + 4 + i * 12,
        part = {
          text: this.text(this.word(pos)),
          kind: this.text(this.word(pos + 4)),
        },
        target = this.word(pos + 8);
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
    const at = this.word(8),
      count = this.word(at);
    if (at + 4 + count * 8 > this.view!.byteLength) return invalid();
    return Array.from({ length: count }, (_, i) => {
      const pos = at + 4 + i * 8,
        text = this.word(pos + 4);
      return {
        name: this.text(this.word(pos)),
        text: text ? this.displayParts(text) : undefined,
      };
    });
  }
  private diagnostics(at: number): DiagnosticFact[] {
    if (!at) return [];
    const count = this.word(at);
    if (at + 4 + count * 40 > this.view!.byteLength) return invalid();
    return Array.from({ length: count }, (_, i) => {
      const pos = at + 4 + i * 40;
      const file = this.word(pos),
        source = this.word(pos + 24),
        flags = this.word(pos + 28);
      const config =
        flags & 4
          ? {
              name: this.text(this.word(file)),
              path: this.text(this.word(file + 4)),
              extendedSourceFiles: this.optionalIDs(file + 8)?.map((name) =>
                this.text(name),
              ),
            }
          : undefined;
      return {
        fileName: config?.name ?? (file ? this.text(file) : undefined),
        start: this.word(pos + 4),
        length: this.word(pos + 8),
        category: this.word(pos + 12),
        code: this.word(pos + 16),
        text: this.text(this.word(pos + 20)),
        source: source ? this.text(source) : undefined,
        config,
        reportsDeprecated: flags & 1 ? true : undefined,
        reportsUnnecessary: flags & 2 ? true : undefined,
        chain: this.diagnostics(this.word(pos + 32)),
        related: this.diagnostics(this.word(pos + 36)),
      };
    });
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
        // Earlier packets point straight past the single offset at byte 8.
        value = this.indexInfos(
          this.word(8),
          this.word(8) === 12 ? 0 : this.word(12),
        );
        break;
      case 10:
        value = this.typePredicate();
        break;
      case 11:
        value = this.namedSymbols();
        break;
      case 12:
        value = this.displayParts(this.word(8));
        break;
      case 13:
        value = this.jsDocTags();
        break;
      case 14: {
        const len = this.word(8);
        if (16 + len > this.bytes!.length) return invalid();
        value = len
          ? {
              tree: Uint8Array.from(this.bytes!.subarray(16, 16 + len)),
              origins: this.optionalIDs(12)?.map((offset) => this.node(offset)),
            }
          : undefined;
        break;
      }
      case 16:
        switch (this.word(8)) {
          case 0:
            value = undefined;
            break;
          case 1:
            value = this.text(this.word(12));
            break;
          case 2:
            value = this.view!.getFloat64(12, true);
            break;
          default:
            return invalid();
        }
        break;
      case 20:
        value = this.word(8)
          ? {
              properties: this.ids(this.word(8)),
              calls: this.ids(this.word(12)),
              constructs: this.ids(this.word(16)),
              indices: this.indexInfos(this.word(20), this.word(24)),
            }
          : undefined;
        break;
      case 19: {
        const list = (offset: number) =>
          this.word(offset) ? this.ids(this.word(offset)) : undefined;
        value = {
          pattern: this.node(this.word(8)),
          node: this.node(this.word(12)),
          regular: this.word(20),
          base: this.word(24),
          constraint: this.word(28),
          outer: list(32),
          local: list(36),
          elements: list(40),
          minLength: this.word(44) ? this.word(44) - 1 : undefined,
          fixedLength: this.word(48) ? this.word(48) - 1 : undefined,
          labels: list(52)?.map((at) => this.node(at)),
          texts: list(56)?.map((at) => this.text(at)),
          root: this.node(this.word(60)),
          rootCheck: this.word(64),
          rootExtends: this.word(68),
          rootAlias: this.word(72),
          distributive: this.word(76) !== 0,
          infer: list(80),
          rootOuter: list(84),
          rootArguments: list(88),
        };
        break;
      }
      case 18:
        value = this.ids(this.word(8)).map((at) => this.text(at));
        break;
      case 17:
        value = this.word(8) ? this.ids(this.word(8)) : undefined;
        break;
      case 15:
        value = this.diagnostics(this.word(8));
        break;
      default:
        return invalid();
    }
    return value as QueryResults[K];
  }
}

// Supplemental slices have independent packet offset origins. The owner copies
// each retained packet into exact storage so one small fact never pins a reply.
export function typeReplyPackets(bytes: Uint8Array): {
  packet: Uint8Array;
  records: Array<{
    op: 'typeRecord' | 'symbolRecord' | 'signatureRecord';
    id: number;
    packet: Uint8Array;
  }>;
} {
  const view = new DataView(bytes.buffer, bytes.byteOffset, bytes.byteLength);
  if (bytes.length < 4 || view.getUint32(0, true) !== 0x31425452)
    return { packet: bytes, records: [] };
  if (bytes.length < 12) return invalid();
  const length = view.getUint32(4, true),
    count = view.getUint32(8, true),
    start = 12 + count * 16;
  if (start > bytes.length || start + length > bytes.length) return invalid();
  const records = [];
  for (let i = 0; i < count; i++) {
    const at = 12 + i * 16,
      kind = view.getUint32(at, true),
      id = view.getUint32(at + 4, true),
      offset = view.getUint32(at + 8, true),
      size = view.getUint32(at + 12, true);
    const op = (
      [undefined, 'typeRecord', 'symbolRecord', 'signatureRecord'] as const
    )[kind];
    if (!op || offset < start + length || offset + size > bytes.length)
      return invalid();
    records.push({ op, id, packet: bytes.subarray(offset, offset + size) });
  }
  return { packet: bytes.subarray(start, start + length), records };
}

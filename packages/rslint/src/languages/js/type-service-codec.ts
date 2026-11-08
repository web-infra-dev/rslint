/* rslint-disable @typescript-eslint/no-unsafe-type-assertion -- Indexed generic return types follow the finite field switch and checked packet kind. */
import { SCALAR_KINDS } from './type-service-protocol.js';
import type {
  NodeRef,
  TypeFact,
  SymbolFact,
  SignatureFact,
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
  readonly stats = { strings: 0, lists: 0 };
  constructor(bytes: Uint8Array) {
    this.bytes = bytes;
    this.view = new DataView(bytes.buffer, bytes.byteOffset, bytes.byteLength);
    if (bytes.length < 12 || this.word(0) !== 0x34535452) invalid();
    this.kind = this.word(4);
    const words = [0, 1, 1, 1, 1, 9, 5, 2][this.kind];
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
    this.stats.strings++;
    this.strings.set(offset, value);
    return value;
  }
  private ids(offset: number): number[] {
    this.stats.lists++;
    const size = this.word(offset);
    if (offset + 4 + size * 4 > this.view!.byteLength) return invalid();
    return Array.from({ length: size }, (_, i) =>
      this.word(offset + 4 + i * 4),
    );
  }
  private node(offset: number): NodeRef | undefined {
    return offset
      ? { file: this.text(this.word(offset)), index: this.word(offset + 4) }
      : undefined;
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
      case 'types': {
        const at = this.word(32);
        value = at ? this.ids(at) : undefined;
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
      case 'declarations': {
        const at = this.word(20);
        if (!at) value = undefined;
        else {
          this.stats.lists++;
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
    return (
      field === 'parameters' ? this.ids(this.word(8)) : this.node(this.word(12))
    ) as SignatureFact[K];
  }
  scalar<K extends ScalarOperation>(operation: K): QueryResults[K] {
    if (this.kind !== SCALAR_KINDS[operation]) return invalid();
    let value: number | boolean | string | readonly number[];
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
      default:
        return invalid();
    }
    return value as QueryResults[K];
  }
}

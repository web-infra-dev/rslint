/** Immutable Go facts published by one broker into bounded portable storage. */
export interface SharedTypeCacheDescriptor {
  buffer: SharedArrayBuffer;
  generation: number;
  entries: number;
}
const HEADER = 8,
  SLOT = 4;
export const MAX_SHARED_TYPE_PROBES = 64;
function keyBytes(key: string): Uint8Array {
  const bytes = new Uint8Array(key.length * 2);
  for (let i = 0; i < key.length; i++) {
    const unit = key.charCodeAt(i);
    bytes[i * 2] = unit & 255;
    bytes[i * 2 + 1] = unit >>> 8;
  }
  return bytes;
}
function hash(bytes: Uint8Array): number {
  let value = 2166136261;
  for (const byte of bytes) value = Math.imul(value ^ byte, 16777619);
  return value | 0 || 1;
}
interface Storage {
  descriptor: SharedTypeCacheDescriptor;
  control: Int32Array;
  bytes: Uint8Array;
}
function storage(descriptor: SharedTypeCacheDescriptor): Storage {
  const { buffer, entries, generation } = descriptor;
  if (
    !Number.isSafeInteger(entries) ||
    entries < 1 ||
    buffer.byteLength < (HEADER + entries * SLOT) * 4
  )
    throw new Error('invalid shared type cache');
  const control = new Int32Array(buffer, 0, HEADER + entries * SLOT);
  if (Atomics.load(control, 1) !== generation)
    throw new Error('foreign shared type cache generation');
  return { descriptor, control, bytes: new Uint8Array(buffer) };
}
export class SharedTypeCacheReader {
  private storage: Storage | undefined;
  readonly stats = { probes: 0 };
  constructor(descriptor: SharedTypeCacheDescriptor) {
    this.storage = storage(descriptor);
    this.assertOpen();
    Atomics.add(this.storage.control, 3, 1);
  }
  get descriptor(): SharedTypeCacheDescriptor {
    return this.open().descriptor;
  }
  get retainedBytes(): number {
    return this.storage?.bytes.byteLength ?? 0;
  }
  private open(): Storage {
    const current = this.storage;
    if (!current || Atomics.load(current.control, 0) !== 1)
      throw new Error('shared type cache is closed');
    return current;
  }
  assertOpen(): void {
    this.open();
  }
  get(key: string): Uint8Array | undefined {
    const { descriptor, control, bytes } = this.open();
    const encoded = keyBytes(key),
      value = hash(encoded);
    for (
      let step = 0;
      step < Math.min(descriptor.entries, MAX_SHARED_TYPE_PROBES);
      step++
    ) {
      this.stats.probes++;
      const slot =
        HEADER + (((value >>> 0) + step) % descriptor.entries) * SLOT;
      const found = Atomics.load(control, slot);
      if (found === 0) return undefined;
      if (found !== value) continue;
      const offset = control[slot + 1],
        length = control[slot + 2],
        keyLength = control[slot + 3];
      if (
        keyLength === encoded.length &&
        encoded.every((byte, i) => bytes[offset + i] === byte)
      )
        return bytes.subarray(offset + keyLength, offset + keyLength + length);
    }
    return undefined;
  }
  close(): void {
    if (this.storage) {
      Atomics.sub(this.storage.control, 3, 1);
      this.storage = undefined;
    }
  }
}
/** Existing hits survive saturation. A collision only declines that insertion.
 * JavaScript wrappers never receive these shared byte views. */
export class SharedTypeCacheOwner {
  private storage: Storage | undefined;
  readonly stats = { probes: 0 };
  constructor(
    generation: number,
    budgetBytes: number,
    entries = Math.max(1, 2 ** Math.floor(Math.log2(budgetBytes / 64))),
  ) {
    const start = (HEADER + entries * SLOT) * 4;
    if (
      !Number.isSafeInteger(entries) ||
      entries < 1 ||
      !Number.isSafeInteger(budgetBytes) ||
      budgetBytes < start ||
      budgetBytes > 0x7fffffff ||
      !Number.isInteger(generation) ||
      generation < 1 ||
      generation > 0x7fffffff
    )
      throw new Error('invalid shared type cache budget');
    const descriptor = {
      buffer: new SharedArrayBuffer(budgetBytes),
      entries,
      generation,
    };
    const control = new Int32Array(
      descriptor.buffer,
      0,
      HEADER + entries * SLOT,
    );
    control[1] = generation;
    control[2] = start;
    Atomics.store(control, 0, 1);
    this.storage = storage(descriptor);
  }
  private open(): Storage {
    if (!this.storage) throw new Error('shared type cache is closed');
    return this.storage;
  }
  get descriptor(): SharedTypeCacheDescriptor {
    return this.open().descriptor;
  }
  publish(key: string, packet: Uint8Array): boolean {
    const { descriptor, control, bytes } = this.open();
    const encoded = keyBytes(key),
      value = hash(encoded);
    let available = -1;
    for (
      let step = 0;
      step < Math.min(descriptor.entries, MAX_SHARED_TYPE_PROBES);
      step++
    ) {
      this.stats.probes++;
      const slot =
        HEADER + (((value >>> 0) + step) % descriptor.entries) * SLOT;
      const found = Atomics.load(control, slot);
      if (found === 0) {
        available = slot;
        break;
      }
      if (found !== value || control[slot + 3] !== encoded.length) continue;
      const offset = control[slot + 1];
      if (encoded.every((byte, i) => bytes[offset + i] === byte)) return true;
    }
    if (available < 0) return false;
    const offset = control[2],
      end = offset + encoded.length + packet.length;
    if (end > bytes.length) return false;
    bytes.set(encoded, offset);
    bytes.set(packet, offset + encoded.length);
    control[available + 1] = offset;
    control[available + 2] = packet.length;
    control[available + 3] = encoded.length;
    control[2] = end;
    control[4]++;
    Atomics.store(control, available, value);
    return true;
  }
  close(): void {
    if (this.storage) Atomics.store(this.storage.control, 0, 0);
    this.storage = undefined;
  }
  get readers(): number {
    return Atomics.load(this.open().control, 3);
  }
  get usedBytes(): number {
    return this.open().control[2];
  }
  get entries(): number {
    return this.open().control[4];
  }
  get retainedBytes(): number {
    return this.storage?.bytes.byteLength ?? 0;
  }
}

/** A bounded cache of immutable Go packets, owned by the query broker.
 * Publication copies bytes once into portable JS-owned shared storage. This
 * does not retain native request leases or share decoded JavaScript objects.
 */
export interface SharedTypeCacheDescriptor {
  buffer: SharedArrayBuffer;
  generation: number;
  entries: number;
}
const HEADER = 8,
  SLOT = 4;
const encoder = new TextEncoder();
function hash(bytes: Uint8Array): number {
  let value = 2166136261;
  for (const byte of bytes) value = Math.imul(value ^ byte, 16777619);
  return value | 0 || 1;
}
export class SharedTypeCacheReader {
  protected control: Int32Array;
  protected bytes: Uint8Array;
  private closed = false;
  constructor(readonly descriptor: SharedTypeCacheDescriptor) {
    const { buffer, entries, generation } = descriptor;
    if (
      !Number.isSafeInteger(entries) ||
      entries < 1 ||
      buffer.byteLength < (HEADER + entries * SLOT) * 4
    )
      throw new Error('invalid shared type cache');
    this.control = new Int32Array(buffer, 0, HEADER + entries * SLOT);
    this.bytes = new Uint8Array(buffer);
    if (Atomics.load(this.control, 1) !== generation)
      throw new Error('foreign shared type cache generation');
    this.assertOpen();
    Atomics.add(this.control, 3, 1);
  }
  assertOpen(): void {
    if (this.closed || Atomics.load(this.control, 0) !== 1)
      throw new Error('shared type cache is closed');
  }
  get(key: string): Uint8Array | undefined {
    this.assertOpen();
    const encoded = encoder.encode(key),
      value = hash(encoded);
    for (let step = 0; step < this.descriptor.entries; step++) {
      const slot =
        HEADER + (((value >>> 0) + step) % this.descriptor.entries) * SLOT;
      const found = Atomics.load(this.control, slot);
      if (found === 0) return undefined;
      if (found !== value) continue;
      const offset = this.control[slot + 1],
        length = this.control[slot + 2],
        keyLength = this.control[slot + 3];
      if (keyLength !== encoded.length) continue;
      if (encoded.every((byte, i) => this.bytes[offset + i] === byte))
        return this.bytes.subarray(
          offset + keyLength,
          offset + keyLength + length,
        );
    }
    return undefined;
  }
  close(): void {
    if (!this.closed) {
      this.closed = true;
      Atomics.sub(this.control, 3, 1);
    }
  }
}

/** Exactly one broker may publish. Readers may live in different workers.
 * close invalidates readers immediately; SAB storage is reclaimed by JS after
 * the last reader/descriptor releases it. No lazy plugin value aliases it.
 */
export class SharedTypeCacheOwner {
  readonly descriptor: SharedTypeCacheDescriptor;
  private control: Int32Array;
  private bytes: Uint8Array;
  constructor(generation: number, budgetBytes: number, entries: number) {
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
    const buffer = new SharedArrayBuffer(budgetBytes);
    this.descriptor = { buffer, entries, generation };
    this.control = new Int32Array(buffer, 0, HEADER + entries * SLOT);
    this.bytes = new Uint8Array(buffer);
    this.control[1] = generation;
    this.control[2] = start;
    Atomics.store(this.control, 0, 1);
  }
  publish(key: string, packet: Uint8Array): boolean {
    if (Atomics.load(this.control, 0) !== 1)
      throw new Error('shared type cache is closed');
    const encoded = encoder.encode(key),
      value = hash(encoded);
    for (let step = 0; step < this.descriptor.entries; step++) {
      const slot =
        HEADER + (((value >>> 0) + step) % this.descriptor.entries) * SLOT;
      if (Atomics.load(this.control, slot) !== 0) continue;
      const offset = this.control[2],
        end = offset + encoded.length + packet.length;
      if (end > this.bytes.length) return false;
      this.bytes.set(encoded, offset);
      this.bytes.set(packet, offset + encoded.length);
      this.control[slot + 1] = offset;
      this.control[slot + 2] = packet.length;
      this.control[slot + 3] = encoded.length;
      this.control[2] = end;
      Atomics.store(this.control, slot, value);
      return true;
    }
    return false;
  }
  close(): void {
    Atomics.store(this.control, 0, 0);
  }
  get readers(): number {
    return Atomics.load(this.control, 3);
  }
  get usedBytes(): number {
    return this.control[2];
  }
}

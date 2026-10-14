/** Native value types shared by browser-safe protocols and Node consumers. */

/** Opaque reader capability; never a pointer or a filesystem path. */
export interface SharedBytes {
  lease: number;
  offset: number;
  length: number;
}

/** Complete text, owned bytes, or a request-scoped native byte capability. */
export type ByteInput = string | Uint8Array | SharedBytes;

/** Native arena layout supplied at runtime by Go; no defaults live here. */
export interface MemoryConfiguration {
  version: number;
  slotCount: number;
  slotSize: number;
  headerSize: number;
  publicationStride: number;
}

export interface MemoryBatch {
  slot: number;
  generation: number;
  length: number;
}

/** N-API optional return values may use null; the wire envelope omits them. */
export interface NativeMemoryMapping {
  version: number;
  fd?: number | null;
  handle?: string | null;
  processId?: number | null;
}

export interface MemoryArena {
  fd(): number | null | undefined;
  configure(config: MemoryConfiguration): void;
  descriptor(): NativeMemoryMapping;
  register(batches: MemoryBatch[]): number;
  /** Revoke reads; retry while existing readers prevent reuse. True only once. */
  release(lease: number): boolean;
  close(): void;
}

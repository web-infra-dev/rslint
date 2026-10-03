/** Native value types shared by browser-safe protocols and Node consumers. */

/** ESLint-shape comment (`{ type, value, start, end }`); start/end are UTF-16 offsets. */
export interface CommentObj {
  /** "Line" | "Block" */
  type: string;
  /** Comment body with the `//` or block delimiters stripped. */
  value: string;
  start: number;
  end: number;
}

/** Parser output: ESTree JSON + comments + columnar token arrays (all UTF-16 offsets). */
export interface ParseResult {
  /** ESTree as a JSON string (no `range`; normalize-ast derives it from start/end). */
  program: string;
  comments: Array<CommentObj>;
  /** Parser-driven token stream in columnar form. */
  tokenTypes: Uint8Array;
  tokenStarts: Uint32Array;
  tokenEnds: Uint32Array;
}

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
  /** Transfer the configured Unix descriptor to Go's one-shot listener. */
  sendFd(socketPath: string): void;
  register(batches: MemoryBatch[]): number;
  /** Revoke reads; retry while existing readers prevent reuse. True only once. */
  release(lease: number): boolean;
  close(): void;
}

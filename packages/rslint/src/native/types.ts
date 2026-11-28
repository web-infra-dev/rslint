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

/** Optional bidirectional shared-memory RPC layout advertised by Go. */
export interface RpcConfiguration {
  version: number;
  slotCount: number;
  slotSize: number;
  headerSize: number;
  slotHeaderStride: number;
}

/** N-API descriptor returned by RpcChannel.configure() for transportRpcOpen. */
export interface NativeRpcMapping {
  version: number;
  fd?: number | null;
  doorbellFd?: number | null;
  handle?: string | null;
  nodeToGoEvent?: string | null;
  goToNodeEvent?: string | null;
  processId?: number | null;
}

export interface RpcSlotLease {
  slot: number;
  epoch: number;
}

export interface RpcChannel {
  fd(): number | null | undefined;
  doorbellFd(): number | null | undefined;
  afterSpawn(): void;
  configure(
    config: RpcConfiguration,
    inheritedFd?: number | null,
    inheritedDoorbellFd?: number | null,
    peerPid?: number | null,
  ): NativeRpcMapping;
  channelId(): number | null | undefined;
  close(): void;
}

/** Plain structured-cloneable descriptor passed to a plugin worker via workerData. */
export interface WorkerRpcDescriptor {
  readonly channelId: number;
  readonly slot: number;
  readonly epoch: number;
  readonly slotSize: number;
}

/**
 * Dependency-injected RPC slot controller passed from IPC into WorkerPool.
 * Keeps `dist/eslint-plugin/host.js` completely free of native addon imports.
 */
export interface RpcHostChannel {
  readonly slotSize: number;
  acquireSlot(): WorkerRpcDescriptor | undefined;
  abortSlot(slot: number): void;
  releaseSlot(slot: number): void;
}

/* rslint-disable @typescript-eslint/no-unsafe-type-assertion */
/**
 * Per-task cancel flag infrastructure.
 *
 * `Atomics.load(int32, 0)` is cheap enough to poll on every AST node
 * visit (single-digit nanoseconds over baseline), so we allocate one
 * Int32 per in-flight task in a single growable SharedArrayBuffer, set
 * it from the main thread when the parent requests cancel, and have
 * the Worker observe it via Atomics.load on each visit.
 *
 * Why one shared SAB rather than one SAB per task:
 *
 *   - Allocating SABs has nontrivial cost (zeroed mmap-equivalent).
 *   - Recycling slots in one SAB costs nothing.
 *   - structuredClone's cost on a tiny shared SAB is the same as on
 *     separate SABs.
 *
 * The flag values:
 *
 *   0  = not cancelled (initial)
 *   1  = cancelled (set by main thread; observed by worker)
 *
 * Sizing: the SAB is constructed with `maxByteLength` so it can grow
 * on demand without invalidating worker-held views (Node's growable
 * SAB semantics; the worker's `viewForSlot` view at `slot * 4` stays
 * valid across grow because grow only EXTENDS the buffer — existing
 * bytes / offsets don't move). When a batch dispatches more tasks
 * than the current capacity holds, `acquire` doubles the SAB up to
 * `MAX_CAPACITY`. No artificial low cap; lint runs over arbitrarily
 * many files are first-class.
 */

/** Initial capacity. Doubles on demand. Small so single-file runs don't pay. */
const INITIAL_SLOTS = 1024;

/**
 * Hard upper bound on slot count. Each slot is 4 bytes, so the SAB's
 * maxByteLength = `MAX_CAPACITY * 4`. Set high enough to cover any
 * realistic single-batch lint (millions of files would be a different
 * kind of problem) while still bounding the maxByteLength reservation
 * the runtime tracks per SAB. 4 MiB of slots = ~1 M concurrent in-flight
 * tasks — far above anything we expect.
 */
const MAX_CAPACITY = 1 << 20;

/**
 * Pool of cancel slots backed by one growable SharedArrayBuffer. Each
 * slot is one Int32 (4 bytes). The pool hands out slot indices and
 * recycles them when tasks complete; it grows the SAB when the active
 * task count exceeds the current capacity.
 */
export class CancelFlagPool {
  // sab is never reassigned — only its underlying buffer grows via
  // `SharedArrayBuffer.grow()` (a mutator on the existing instance).
  private readonly sab: SharedArrayBuffer;
  private view: Int32Array;
  /** Free slot stack — pop on alloc, push on free. */
  private readonly freeList: number[];
  /**
   * Per-slot allocation status, byte-addressed so release() can dedup
   * in O(1) instead of O(N) scanning the freeList. Without this guard
   * a double-release would push the same index twice onto the stack,
   * and a later acquire could hand out an index that's still "in
   * use" by another caller — silently breaking cancel routing.
   *
   * 1 = allocated, 0 = free. Resized in lockstep with the SAB.
   */
  private slotInUse: Uint8Array;
  private capacity: number;

  constructor(initialCapacity: number = INITIAL_SLOTS) {
    this.capacity = Math.min(initialCapacity, MAX_CAPACITY);
    // grow() commits incrementally. Workers receive `this.sab` once at
    // spawn; their Int32Array view (built via viewForSlot) stays valid
    // after every grow because:
    //   - grow() can only EXTEND the buffer (never shrink),
    //   - existing byte offsets keep mapping to the same memory,
    //   - per-slot views are length-1 at `slot * 4`, so they don't
    //     span the boundary that grow extends past.
    // `maxByteLength` reserves virtual address space without committing
    // it (lib.dom / es2024 type may not list the options bag yet, so
    // cast the constructor signature for the second arg).
    const SABCtor = SharedArrayBuffer as unknown as new (
      length: number,
      options?: { maxByteLength?: number },
    ) => SharedArrayBuffer;
    this.sab = new SABCtor(this.capacity * 4, {
      maxByteLength: MAX_CAPACITY * 4,
    });
    this.view = new Int32Array(this.sab);
    this.slotInUse = new Uint8Array(this.capacity);
    this.freeList = [];
    for (let i = this.capacity - 1; i >= 0; i--) this.freeList.push(i);
  }

  /** Returns the underlying SAB. Send to Worker via workerData / postMessage. */
  get sharedBuffer(): SharedArrayBuffer {
    return this.sab;
  }

  /** Current allocated capacity. Grows on demand; exposed for diagnostics. */
  get size(): number {
    return this.capacity;
  }

  /**
   * Allocate a slot. Returns the slot index; grows the SAB if the
   * current capacity is fully checked out. Returns -1 only when the
   * hard upper bound `MAX_CAPACITY` is hit (effectively never under
   * realistic workloads — ~1 M concurrent tasks).
   */
  acquire(): number {
    if (this.freeList.length === 0) {
      if (!this.grow()) return -1;
    }
    const idx = this.freeList.pop()!;
    this.slotInUse[idx] = 1;
    Atomics.store(this.view, idx, 0); // ensure clean state
    return idx;
  }

  /**
   * Double the SAB's capacity (capped at MAX_CAPACITY) and refill the
   * freeList with the newly minted slot indices. Workers' SAB
   * references and any existing length-1 views stay valid — grow()
   * only extends the buffer's byte length.
   *
   * Returns false when MAX_CAPACITY is already reached (acquire then
   * surfaces -1; caller falls back to no-cancel for that task — same
   * "best-effort cancel" contract as before).
   */
  private grow(): boolean {
    if (this.capacity >= MAX_CAPACITY) return false;
    const next = Math.min(MAX_CAPACITY, this.capacity * 2);
    // `SharedArrayBuffer.prototype.grow` is the standard growable-SAB
    // API. Node 20+ exposes it; the runner package's `engines.node` is
    // ≥ 20 (enforced at startup by plugin-loader.ensureNodeVersion).
    (this.sab as SharedArrayBuffer & { grow(n: number): void }).grow(next * 4);
    // Rebuild the Int32 view over the now-longer SAB. Note: workers
    // hold THEIR OWN views; those are length-1 at fixed offsets and
    // remain valid (grow only extends, never moves existing bytes).
    this.view = new Int32Array(this.sab);
    // slotInUse byte map needs to extend in lockstep.
    const widened = new Uint8Array(next);
    widened.set(this.slotInUse);
    this.slotInUse = widened;
    // Add freshly minted slot indices to the free list (high indices
    // first so pop() hands out the lowest unused index first — keeps
    // diagnostic output stable across runs).
    for (let i = next - 1; i >= this.capacity; i--) this.freeList.push(i);
    this.capacity = next;
    return true;
  }

  /** Release a slot back to the free list. Idempotent on already-free slots. */
  release(idx: number): void {
    if (idx < 0 || idx >= this.capacity) return;
    // O(1) double-release guard via slotInUse byte map. Previously we
    // scanned freeList.includes — O(N) per release, which becomes
    // visible under sustained LSP per-keystroke loads where release
    // fires for every dispatched task. The byte map gives the same
    // safety with constant cost.
    if (this.slotInUse[idx] === 0) return;
    this.slotInUse[idx] = 0;
    Atomics.store(this.view, idx, 0);
    this.freeList.push(idx);
  }

  /** Set the cancel flag for a slot. Called from the main thread. */
  cancel(idx: number): void {
    if (idx < 0 || idx >= this.capacity) return;
    Atomics.store(this.view, idx, 1);
  }
}

/**
 * Worker side: build an Int32Array view over the supplied SAB and slot
 * index. Length-1 view positioned at the slot's offset. The worker
 * passes this view to the visit() function, which polls Atomics.load(view, 0)
 * on every node.
 */
export function viewForSlot(sab: SharedArrayBuffer, slot: number): Int32Array {
  return new Int32Array(sab, slot * 4, 1);
}

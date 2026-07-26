import { describe, test, expect } from '@rstest/core';
import {
  CancelFlagPool,
  viewForSlot,
} from '../../src/eslint-plugin/cancel-flag.js';

/**
 * CancelFlagPool capacity contract:
 *
 *  - The SAB grows on demand: `acquire` never returns -1 under
 *    realistic load. Any concurrent task count up to MAX_CAPACITY
 *    (~1 M slots) gets a slot. This replaces the previous bounded
 *    design that silently dropped cancellation for tasks beyond the
 *    pool's fixed size — a real problem on large monorepos where a
 *    single batch can exceed 4 K files.
 *
 *  - Worker views over the SAB remain valid across grow() because
 *    grow only extends the buffer; existing per-slot offsets keep
 *    addressing the same memory.
 *
 *  - Slot views observe stores from any holder of the SAB — the
 *    cross-thread invariant the cancel mechanism is built on.
 */

describe('CancelFlagPool', () => {
  test('default capacity handles many concurrent tasks', () => {
    const pool = new CancelFlagPool();
    // Sustained acquires beyond the initial capacity must keep
    // succeeding — the pool grows transparently.
    const slots: number[] = [];
    for (let i = 0; i < 256; i++) {
      const s = pool.acquire();
      expect(s).toBeGreaterThanOrEqual(0);
      slots.push(s);
    }
    expect(new Set(slots).size).toBe(256); // all distinct
    // Release them all — capacity returns.
    for (const s of slots) pool.release(s);
    const reacquired = pool.acquire();
    expect(reacquired).toBeGreaterThanOrEqual(0);
  });

  test('acquire grows the pool past the initial capacity', () => {
    // Initial 4 slots. Acquire 4 — pool full. The 5th must trigger a
    // grow (NOT return -1) and yield a fresh slot. Pre-refactor this
    // returned -1 and surfaced a "pool exhausted" stderr warning.
    const pool = new CancelFlagPool(4);
    const acquired = new Set<number>();
    for (let i = 0; i < 64; i++) {
      const s = pool.acquire();
      expect(s).toBeGreaterThanOrEqual(0);
      // All slots must be distinct — grow() must mint new indices,
      // not recycle in-use ones.
      expect(acquired.has(s)).toBe(false);
      acquired.add(s);
    }
    // Internal capacity now > initial.
    expect(pool.size).toBeGreaterThan(4);
    // Release one, re-acquire — must round-trip.
    const someSlot = acquired.values().next().value as number;
    pool.release(someSlot);
    const fresh = pool.acquire();
    expect(fresh).toBeGreaterThanOrEqual(0);
  });

  test('cancel observed via viewForSlot (cross-thread invariant)', () => {
    const pool = new CancelFlagPool(8);
    const slot = pool.acquire();
    expect(slot).toBeGreaterThanOrEqual(0);

    // Worker side: same SAB, length-1 view at slot offset.
    const view = viewForSlot(pool.sharedBuffer, slot);
    expect(Atomics.load(view, 0)).toBe(0); // initially uncancelled

    // Main thread cancels; worker observes.
    pool.cancel(slot);
    expect(Atomics.load(view, 0)).toBe(1);

    // Release resets the flag for the next user of the slot.
    pool.release(slot);
    expect(Atomics.load(view, 0)).toBe(0);
  });

  test('release tolerates out-of-range or already-released slots', () => {
    const pool = new CancelFlagPool(4);
    // Out-of-range slots are silently ignored (no throw).
    expect(() => pool.release(-1)).not.toThrow();
    expect(() => pool.release(99)).not.toThrow();
    // Idempotent release of a valid slot.
    const s = pool.acquire();
    pool.release(s);
    expect(() => pool.release(s)).not.toThrow();
  });

  // A5 regression. The previous release implementation used
  // `freeList.includes(idx)` to dedup pushes — O(N) per release. The
  // fix uses an O(1) byte map (`slotInUse`). The new path must STILL
  // prevent double-frees from breaking capacity tracking: if release
  // pushed `idx` twice, a later acquire could hand out the same slot
  // to two callers and cancel routing would be silently broken.
  test('double-release does not duplicate the slot in freeList (capacity correctness)', () => {
    // Start narrow so we can observe whether double-release inflated
    // the freeList beyond actual capacity.
    const pool = new CancelFlagPool(3);
    const s1 = pool.acquire();
    const s2 = pool.acquire();
    const s3 = pool.acquire();
    expect(pool.size).toBe(3);

    // Double-release s1. Even though release ran twice, only ONE slot
    // should re-enter the freeList. The invariant: total distinct
    // acquirable slots from the freeList equals 1 (just s1).
    pool.release(s1);
    pool.release(s1); // accidental double-release

    // First acquire after release: must return s1 (the one released).
    const reacquired1 = pool.acquire();
    expect(reacquired1).toBeGreaterThanOrEqual(0);

    // The next acquire would normally grow the pool; we don't care
    // about the value, only that double-release didn't smuggle s1
    // back into the freeList a second time AS A DUPLICATE that could
    // collide with s2 or s3 (both still held). Verify s2/s3 are NOT
    // returned by a fresh acquire.
    const acquired = new Set<number>([reacquired1, s2, s3]);
    for (let i = 0; i < 8; i++) {
      const next = pool.acquire();
      expect(acquired.has(next)).toBe(false); // never collides with held slots
      acquired.add(next);
    }
  });

  // Growable contract pinning: the SAB must extend past the initial
  // capacity without invalidating already-handed-out slot views.
  // Pre-refactor this scenario silently dropped cancellation on tasks
  // beyond the initial capacity; now it just grows.
  test('worker view stays valid across grow()', () => {
    const pool = new CancelFlagPool(2);
    const earlySlot = pool.acquire();
    expect(earlySlot).toBeGreaterThanOrEqual(0);
    // Hold a view over earlySlot, exactly what the worker does.
    const earlyView = viewForSlot(pool.sharedBuffer, earlySlot);
    expect(Atomics.load(earlyView, 0)).toBe(0);

    // Force several grows by acquiring many more slots than the
    // initial capacity. Each acquire that exhausts the current
    // capacity triggers grow().
    for (let i = 0; i < 32; i++) {
      const s = pool.acquire();
      expect(s).toBeGreaterThanOrEqual(0);
    }
    expect(pool.size).toBeGreaterThan(2);

    // The early view MUST still address its original slot. Cancel
    // it from the main side; the long-held view must see the flip.
    pool.cancel(earlySlot);
    expect(Atomics.load(earlyView, 0)).toBe(1);
  });

  test('release never scans the free list and still deduplicates', () => {
    const pool = new CancelFlagPool(8);
    const slot = pool.acquire();
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const internals = pool as any;
    const freeList = internals.freeList as number[];
    let indexedReads = 0;
    let lengthReads = 0;
    const traversalMethods = new Set([
      'entries',
      'every',
      'filter',
      'find',
      'findIndex',
      'forEach',
      'includes',
      'indexOf',
      'keys',
      'lastIndexOf',
      'map',
      'reduce',
      'reduceRight',
      'some',
      'values',
    ]);
    internals.freeList = new Proxy(freeList, {
      get(target, property) {
        if (property === 'push') {
          return (...values: number[]) => target.push(...values);
        }
        if (
          property === Symbol.iterator ||
          (typeof property === 'string' && traversalMethods.has(property))
        ) {
          throw new Error(`release scanned freeList via ${String(property)}`);
        }
        if (property === 'length') lengthReads++;
        if (typeof property === 'string' && /^\d+$/.test(property)) {
          indexedReads++;
        }
        return Reflect.get(target, property);
      },
    });

    expect(() => {
      pool.release(slot);
      pool.release(slot);
    }).not.toThrow();
    // O(1) implementations may inspect length or a tail element. A hand-written
    // scan over the seven free slots would exceed these constant bounds.
    expect(lengthReads).toBeLessThanOrEqual(4);
    expect(indexedReads).toBeLessThanOrEqual(2);
    expect(freeList.filter((value) => value === slot)).toHaveLength(1);
  });
});

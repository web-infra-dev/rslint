import { describe, test, expect } from '@rstest/core';
import { spliceTypeSnapshots } from '../src/engine.js';

// spliceTypeSnapshots is the Node mirror of Go's hoistTypeSnapshots: the CLI
// dispatcher hoists each file's type snapshot into the IPC frame's binary
// trailer (referenced by a 1-based typeSnapshotIndex), and this splices the
// matching ArrayBuffer back onto the file before the worker pool sees it. These
// pin the 1-based index contract end of that handshake.
describe('spliceTypeSnapshots (CLI binary-trailer → typeSnapshot)', () => {
  test('splices each blob onto the file at its 1-based typeSnapshotIndex', () => {
    const blobA = new Uint8Array([1, 2, 3]).buffer;
    const blobB = new Uint8Array([4, 5]).buffer;
    const data = {
      files: [
        { path: 'a.ts', typeSnapshotIndex: 1 },
        { path: 'b.ts' }, // no index → untouched
        { path: 'c.ts', typeSnapshotIndex: 2 },
      ],
    };
    const out = spliceTypeSnapshots(data, [blobA, blobB]) as {
      files: Array<{ typeSnapshot?: unknown; typeSnapshotIndex?: number }>;
    };
    // Mutated in place + returned (same reference).
    expect(out).toBe(data);
    expect(out.files[0].typeSnapshot).toBe(blobA); // index 1 → blobs[0]
    expect(out.files[1].typeSnapshot).toBeUndefined(); // no index
    expect(out.files[2].typeSnapshot).toBe(blobB); // index 2 → blobs[1]
  });

  test('leaves files untouched when there is no binary trailer (LSP path / no snapshots)', () => {
    const data = {
      files: [
        { path: 'a.ts', typeSnapshot: 'base64data', typeSnapshotIndex: 0 },
      ],
    };
    // No binary → return as-is (the LSP path already carries base64 in
    // typeSnapshot; index 0 means "no trailer blob").
    expect(spliceTypeSnapshots(data, undefined)).toBe(data);
    expect(spliceTypeSnapshots(data, [])).toBe(data);
    expect(data.files[0].typeSnapshot).toBe('base64data');
  });

  test('ignores an out-of-range index (defensive — never throws)', () => {
    const blob = new Uint8Array([1]).buffer;
    const data = { files: [{ path: 'a.ts', typeSnapshotIndex: 5 }] };
    const out = spliceTypeSnapshots(data, [blob]) as {
      files: Array<{ typeSnapshot?: unknown }>;
    };
    expect(out.files[0].typeSnapshot).toBeUndefined();
  });

  test('tolerates malformed data without throwing', () => {
    const blob = new Uint8Array([1]).buffer;
    expect(() => spliceTypeSnapshots(null, [blob])).not.toThrow();
    expect(() => spliceTypeSnapshots(42, [blob])).not.toThrow();
    expect(() => spliceTypeSnapshots({}, [blob])).not.toThrow();
    expect(() => spliceTypeSnapshots({ files: 'nope' }, [blob])).not.toThrow();
    // A non-object/null payload is returned verbatim.
    expect(spliceTypeSnapshots(null, [blob])).toBe(null);
  });
});

import { execFileSync } from 'node:child_process';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { describe, expect, test } from 'rstack/test';
import ts from 'typescript';
import { TYPESCRIPT_COMPILER_DATA } from '../src/languages/js/type-service-compiler-data.generated.js';
import { PluginTypeServices } from '../src/eslint-plugin/type-service-host.js';
import { resolveTypeServiceReply } from '../src/internal/plugin-type-service.js';
import type {
  TypeServiceHostRequest,
  TypeServiceHostResponse,
} from '../src/eslint-plugin/plugin/type-service-protocol.js';

function forwardEnum(
  table: Record<string, string | number>,
): Record<string, number> {
  const result: Record<string, number> = {};
  for (const [key, value] of Object.entries(table)) {
    if (typeof value === 'number') {
      result[key] = value;
    }
  }
  return result;
}

describe('TYPESCRIPT_COMPILER_DATA', () => {
  test('matches devDependencies typescript compiler data and stays in sync with generator', () => {
    const genScript = path.resolve(
      path.dirname(fileURLToPath(import.meta.url)),
      '../../../scripts/gen-type-service-enums.mjs',
    );
    execFileSync(process.execPath, [genScript, '--check'], {
      stdio: 'pipe',
    });

    expect(TYPESCRIPT_COMPILER_DATA.SyntaxKind).toEqual(
      forwardEnum(ts.SyntaxKind),
    );
    expect(TYPESCRIPT_COMPILER_DATA.TypeFlags).toEqual(
      forwardEnum(ts.TypeFlags),
    );
    expect(TYPESCRIPT_COMPILER_DATA.ObjectFlags).toEqual(
      forwardEnum(ts.ObjectFlags),
    );
    expect(TYPESCRIPT_COMPILER_DATA.NodeFlags).toEqual(
      forwardEnum(ts.NodeFlags),
    );
    expect(TYPESCRIPT_COMPILER_DATA.JsxEmit).toEqual(forwardEnum(ts.JsxEmit));
    expect(TYPESCRIPT_COMPILER_DATA.ModuleDetectionKind).toEqual(
      forwardEnum(ts.ModuleDetectionKind),
    );
    expect(TYPESCRIPT_COMPILER_DATA.ScriptTarget).toEqual(
      forwardEnum(ts.ScriptTarget),
    );
    expect(Array.isArray(TYPESCRIPT_COMPILER_DATA.builderEmitFlags)).toBe(true);
    expect(TYPESCRIPT_COMPILER_DATA.builderEmitFlags).toHaveLength(8);
    expect(TYPESCRIPT_COMPILER_DATA.jsDocKinds.FunctionDeclaration).toBe(true);
    expect(TYPESCRIPT_COMPILER_DATA.jsDocKinds.ModuleBlock).toBe(true);
  });
});

describe('PluginTypeServices', () => {
  const descriptor = {
    generation: 1,
    currentDirectory: '/repo',
    compilerOptions: {},
    sources: [],
  };

  test('leader cancellation does not poison concurrent follower opening the same session', async () => {
    const host = new PluginTypeServices();
    let openCalls = 0;
    let openSignal: AbortSignal | undefined;
    let resolveOpen!: (res: TypeServiceHostResponse) => void;
    const openDeferred = new Promise<TypeServiceHostResponse>((resolve) => {
      resolveOpen = resolve;
    });

    const transport = async (
      req: TypeServiceHostRequest,
      signal?: AbortSignal,
    ): Promise<TypeServiceHostResponse> => {
      if (req.open) {
        openCalls++;
        openSignal = signal;
        return openDeferred;
      }
      if (req.projection) {
        return {
          projection: {
            origins: [],
            reverse: {},
            aliases: null,
            tree: new Uint8Array([1, 2, 3]),
          },
        };
      }
      return {};
    };

    const leaderAbort = new AbortController();
    const followerAbort = new AbortController();

    const leaderPromise = host.query(
      { session: 's1', generation: 1, open: true, projection: '/repo/a.ts' },
      leaderAbort.signal,
      transport,
    );
    const followerPromise = host.query(
      { session: 's1', generation: 1, open: true, projection: '/repo/b.ts' },
      followerAbort.signal,
      transport,
    );

    leaderAbort.abort();
    await expect(leaderPromise).rejects.toMatchObject({ name: 'AbortError' });
    expect(openSignal?.aborted).toBe(false);

    resolveOpen({ descriptor });
    const followerResult = await followerPromise;
    expect(openCalls).toBe(1);
    expect(followerResult.descriptor).toEqual(descriptor);
    expect(followerResult.projection?.tree).toEqual(new Uint8Array([1, 2, 3]));
    host.close();
  });

  test('leader per-file projection failure does not fail concurrent follower', async () => {
    const host = new PluginTypeServices();
    const transport = async (
      req: TypeServiceHostRequest,
    ): Promise<TypeServiceHostResponse> => {
      if (req.open) return { descriptor };
      if (req.projection === '/repo/a.ts') {
        throw new Error('broken projection for a.ts');
      }
      return {
        projection: {
          origins: [],
          reverse: {},
          aliases: null,
          tree: new Uint8Array([4, 5]),
        },
      };
    };

    const ac = new AbortController();
    const [aResult, bResult] = await Promise.allSettled([
      host.query(
        { session: 's2', generation: 1, open: true, projection: '/repo/a.ts' },
        ac.signal,
        transport,
      ),
      host.query(
        { session: 's2', generation: 1, open: true, projection: '/repo/b.ts' },
        ac.signal,
        transport,
      ),
    ]);

    expect(aResult.status).toBe('rejected');
    expect(bResult.status).toBe('fulfilled');
    if (bResult.status === 'fulfilled') {
      expect(bResult.value.descriptor).toEqual(descriptor);
      expect(bResult.value.projection?.tree).toEqual(new Uint8Array([4, 5]));
    }
    host.close();
  });

  test('aborting all waiters aborts the underlying open transport signal and cleans session state', async () => {
    const host = new PluginTypeServices();
    let openSignal: AbortSignal | undefined;
    let openCalls = 0;

    const transport = async (
      req: TypeServiceHostRequest,
      signal?: AbortSignal,
    ): Promise<TypeServiceHostResponse> => {
      if (req.open) {
        openCalls++;
        if (openCalls === 1) {
          openSignal = signal;
          return new Promise(() => {
            // intentionally hangs until aborted
          });
        }
        return { descriptor };
      }
      return {};
    };

    const firstAbort = new AbortController();
    const first = host.query(
      { session: 's3', generation: 1, open: true },
      firstAbort.signal,
      transport,
    );
    firstAbort.abort();
    await expect(first).rejects.toMatchObject({ name: 'AbortError' });
    expect(openSignal?.aborted).toBe(true);

    const retry = await host.query(
      { session: 's3', generation: 1, open: true },
      new AbortController().signal,
      transport,
    );
    expect(openCalls).toBe(2);
    expect(retry.descriptor).toEqual(descriptor);
    host.close();
  });

  test('caches cacheable queries across workers without extra transport calls and clears on retire', async () => {
    const host = new PluginTypeServices();
    let queryCalls = 0;
    const transport = async (
      req: TypeServiceHostRequest,
    ): Promise<TypeServiceHostResponse> => {
      if (req.open) return { descriptor };
      if (req.queries) {
        queryCalls++;
        return {
          packets: req.queries.map(
            (_, idx) => new Uint8Array([queryCalls, idx]),
          ),
        };
      }
      return {};
    };

    const ac = new AbortController();
    await host.query(
      { session: 'cache-1', generation: 1, open: true },
      ac.signal,
      transport,
    );

    // Non-cacheable queries always hit transport.
    await host.query(
      {
        session: 'cache-1',
        generation: 1,
        queries: [{ generation: 1, op: 'semanticDiagnostics', args: {} }],
      },
      ac.signal,
      transport,
    );
    await host.query(
      {
        session: 'cache-1',
        generation: 1,
        queries: [{ generation: 1, op: 'semanticDiagnostics', args: {} }],
      },
      ac.signal,
      transport,
    );
    expect(queryCalls).toBe(2);

    // First cacheable query populates the bounded host cache; repeat hits memory.
    const first = await host.query(
      {
        session: 'cache-1',
        generation: 1,
        queries: [{ generation: 1, op: 'void', args: {} }],
      },
      ac.signal,
      transport,
    );
    const second = await host.query(
      {
        session: 'cache-1',
        generation: 1,
        queries: [{ generation: 1, op: 'void', args: {} }],
      },
      ac.signal,
      transport,
    );
    expect(queryCalls).toBe(3);
    expect(second.packets).toEqual(first.packets);

    // Retiring removes the session and rejects further non-open queries.
    host.retire(['cache-1']);
    await expect(
      host.query(
        {
          session: 'cache-1',
          generation: 1,
          queries: [{ generation: 1, op: 'void', args: {} }],
        },
        ac.signal,
        transport,
      ),
    ).rejects.toMatchObject({ name: 'TypeServiceClosedError' });
    host.close();
  });

  test('validates syntaxOnly descriptor generation and avoids invoking transport when already aborted', async () => {
    const host = new PluginTypeServices();
    let calls = 0;
    const transport = async (): Promise<TypeServiceHostResponse> => {
      calls++;
      return {
        descriptor: { ...descriptor, generation: 99 },
      };
    };

    const aborted = new AbortController();
    aborted.abort();
    await expect(
      host.query(
        { session: 's-abort', generation: 1, open: true },
        aborted.signal,
        transport,
      ),
    ).rejects.toMatchObject({ name: 'AbortError' });
    expect(calls).toBe(0);

    await expect(
      host.query(
        { session: 's-syntax', generation: 1, open: true, syntaxOnly: true },
        new AbortController().signal,
        transport,
      ),
    ).rejects.toThrow(/missing or foreign type service descriptor/);
    host.close();
  });

  test('concurrent open+projection closed status and host.close() surface TypeServiceClosedError instead of AbortError', async () => {
    // Case 1: projection returns status:'closed' before open settles.
    const host1 = new PluginTypeServices();
    let resolveOpen!: (res: TypeServiceHostResponse) => void;
    const transport1 = async (
      req: TypeServiceHostRequest,
    ): Promise<TypeServiceHostResponse> => {
      if (req.open) {
        return new Promise((resolve) => {
          resolveOpen = resolve;
        });
      }
      return { status: 'closed' };
    };
    const p1 = host1.query(
      {
        session: 'race-1',
        generation: 1,
        open: true,
        projection: '/repo/a.ts',
      },
      new AbortController().signal,
      transport1,
    );
    await expect(p1).rejects.toMatchObject({ name: 'TypeServiceClosedError' });
    resolveOpen({ descriptor });
    host1.close();

    // Case 2: host.close() while openSession is in flight.
    const host2 = new PluginTypeServices();
    const transport2 = async (): Promise<TypeServiceHostResponse> =>
      new Promise(() => {
        // intentionally pending until close()
      });
    const p2 = host2.query(
      { session: 'race-2', generation: 1, open: true },
      new AbortController().signal,
      transport2,
    );
    host2.close();
    await expect(p2).rejects.toMatchObject({ name: 'TypeServiceClosedError' });
  });

  test('caches supplemental bundle records, detaches slices, and preserves warm cache when global headroom is tight', async () => {
    const host = new PluginTypeServices();
    const ac = new AbortController();

    // Build a valid 0x31425452 bundle containing a 4-byte primary packet plus
    // one supplemental typeRecord (tag=1, id=42, payload=[9,8,7,6]).
    const primary = new Uint8Array([1, 2, 3, 4]);
    const supplemental = new Uint8Array([9, 8, 7, 6]);
    const bundleBacking = new ArrayBuffer(64);
    const bundleSlice = new Uint8Array(
      bundleBacking,
      8,
      12 + 16 + primary.byteLength + supplemental.byteLength,
    );
    const view = new DataView(
      bundleSlice.buffer,
      bundleSlice.byteOffset,
      bundleSlice.byteLength,
    );
    view.setUint32(0, 0x31425452, true);
    view.setUint32(4, primary.byteLength, true);
    view.setUint32(8, 1, true);
    const packetStart = 12 + 16;
    const suppOffset = packetStart + primary.byteLength;
    view.setUint32(12, 1, true); // tag 1 = typeRecord
    view.setUint32(16, 42, true); // id = 42
    view.setUint32(20, suppOffset, true);
    view.setUint32(24, supplemental.byteLength, true);
    bundleSlice.set(primary, packetStart);
    bundleSlice.set(supplemental, suppOffset);

    let calls = 0;
    const transport = async (
      req: TypeServiceHostRequest,
    ): Promise<TypeServiceHostResponse> => {
      if (req.open) return { descriptor };
      calls++;
      return { packets: req.queries!.map(() => bundleSlice) };
    };

    await host.query(
      { session: 'supp-1', generation: 1, open: true },
      ac.signal,
      transport,
    );
    await host.query(
      {
        session: 'supp-1',
        generation: 1,
        queries: [{ generation: 1, op: 'void', args: {} }],
      },
      ac.signal,
      transport,
    );
    expect(calls).toBe(1);

    // Querying the supplemental typeRecord hits the host cache with zero extra calls,
    // and its cached Uint8Array is detached from the larger 64-byte backing buffer.
    const suppHit = await host.query(
      {
        session: 'supp-1',
        generation: 1,
        queries: [{ generation: 1, op: 'typeRecord', args: { id: 42 } }],
      },
      ac.signal,
      transport,
    );
    expect(calls).toBe(1);
    expect(suppHit.packets).toEqual([supplemental]);
    expect(suppHit.packets![0].buffer.byteLength).toBe(supplemental.byteLength);

    // Fill 4 sessions near the 32 MiB host budget, then attempt to cache a packet
    // larger than remaining global headroom in a 5th session — existing warm entries
    // in the 5th session must not be wiped.
    const sevenMiB = new Uint8Array(7 << 20);
    const bigTransport = async (
      req: TypeServiceHostRequest,
    ): Promise<TypeServiceHostResponse> => {
      if (req.open) return { descriptor };
      return { packets: req.queries!.map(() => sevenMiB) };
    };
    for (const id of ['fill-1', 'fill-2', 'fill-3', 'fill-4']) {
      await host.query(
        { session: id, generation: 1, open: true },
        ac.signal,
        bigTransport,
      );
      await host.query(
        {
          session: id,
          generation: 1,
          queries: [{ generation: 1, op: 'void', args: {} }],
        },
        ac.signal,
        bigTransport,
      );
    }
    // 28 MiB retained across fill-1..4; 4 MiB headroom left globally.
    // Trying to store a 5 MiB packet in supp-1 must NOT evict supp-1's warm cached entries.
    const fiveMiB = new Uint8Array(5 << 20);
    await host.query(
      {
        session: 'supp-1',
        generation: 1,
        queries: [{ generation: 1, op: 'anyType', args: {} }],
      },
      ac.signal,
      async () => ({ packets: [fiveMiB] }),
    );
    const stillWarm = await host.query(
      {
        session: 'supp-1',
        generation: 1,
        queries: [{ generation: 1, op: 'typeRecord', args: { id: 42 } }],
      },
      ac.signal,
      transport,
    );
    expect(calls).toBe(1);
    expect(stillWarm.packets).toEqual([supplemental]);
    host.close();
  });
});

describe('resolveTypeServiceReply', () => {
  test('resolves zero-copy Uint8Array attachment slices, inline packets, projection, source, and status', () => {
    const backing = new Uint8Array([0, 10, 20, 30, 40, 50]);
    const slice = backing.subarray(2, 5); // [20, 30, 40] with non-zero byteOffset

    const attached = resolveTypeServiceReply({
      kind: 'response',
      id: 1,
      data: {
        projection: {
          origins: [],
          reverse: {},
          aliases: null,
          treeAttachment: 0,
        },
        packetAttachments: [0],
        source: { fileName: '/a.ts', path: '/a.ts', treeAttachment: 0 },
      },
      attachments: [slice],
    });
    expect(Buffer.isBuffer(attached.projection?.tree)).toBe(true);
    expect(attached.projection?.tree).toEqual(Buffer.from([20, 30, 40]));
    expect((attached.projection?.tree as Buffer).buffer).toBe(backing.buffer);
    expect(attached.packets).toEqual([Buffer.from([20, 30, 40])]);
    expect(attached.source?.tree).toEqual(Buffer.from([20, 30, 40]));

    const inlineBase64 = Buffer.from([1, 2, 3]).toString('base64');
    const inlineAndMeta = resolveTypeServiceReply({
      kind: 'response',
      id: 2,
      data: {
        packets: [inlineBase64],
        source: { fileName: '/meta.ts', path: '/meta.ts' },
      },
    });
    expect(inlineAndMeta.packets).toEqual([inlineBase64]);
    expect(inlineAndMeta.source).toEqual({
      fileName: '/meta.ts',
      path: '/meta.ts',
      tree: undefined,
    });

    const closed = resolveTypeServiceReply({
      kind: 'response',
      id: 3,
      data: { status: 'closed' },
    });
    expect(closed.status).toBe('closed');

    expect(() =>
      resolveTypeServiceReply({
        kind: 'response',
        id: 4,
        data: { packetAttachments: [1] },
        attachments: [slice],
      }),
    ).toThrow(/invalid type service attachment index/);

    expect(() =>
      resolveTypeServiceReply({
        kind: 'response',
        id: 5,
        data: { packetAttachments: [0] },
        attachments: ['not-binary'],
      }),
    ).toThrow(/not binary/);
  });
});

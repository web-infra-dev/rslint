import { execFileSync } from 'node:child_process';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { describe, expect, test } from 'rstack/test';
import ts from 'typescript';
import { TYPESCRIPT_COMPILER_DATA } from '../src/languages/js/type-service-compiler-data.generated.js';
import { encodeBinaryTypeQuery } from '../src/languages/js/type-service-request-codec.js';
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

  test('refreshes LRU recency on cache reads and distinguishes retire() vs close() errors on in-flight queries', async () => {
    const host = new PluginTypeServices();
    const ac = new AbortController();
    let calls = 0;
    const threeMiB = (tag: number) => {
      const buf = new Uint8Array(3 << 20);
      buf[0] = tag;
      return buf;
    };
    const transport = async (
      req: TypeServiceHostRequest,
    ): Promise<TypeServiceHostResponse> => {
      if (req.open) return { descriptor };
      calls++;
      return {
        packets: req.queries!.map((q) =>
          threeMiB((q.args as { id: number }).id),
        ),
      };
    };

    await host.query(
      { session: 'lru-1', generation: 1, open: true },
      ac.signal,
      transport,
    );
    // Insert 1 (3 MiB) and 2 (3 MiB) sequentially under the 8 MiB session cap.
    await host.query(
      {
        session: 'lru-1',
        generation: 1,
        queries: [{ generation: 1, op: 'typeRecord', args: { id: 1 } }],
      },
      ac.signal,
      transport,
    );
    await host.query(
      {
        session: 'lru-1',
        generation: 1,
        queries: [{ generation: 1, op: 'typeRecord', args: { id: 2 } }],
      },
      ac.signal,
      transport,
    );
    // Read 1 so it becomes MRU ahead of 2.
    await host.query(
      {
        session: 'lru-1',
        generation: 1,
        queries: [{ generation: 1, op: 'typeRecord', args: { id: 1 } }],
      },
      ac.signal,
      transport,
    );
    expect(calls).toBe(2);
    // Insert 3 (3 MiB) -> exceeds 8 MiB session cap, must evict 2 (LRU) and retain 1 (MRU).
    await host.query(
      {
        session: 'lru-1',
        generation: 1,
        queries: [{ generation: 1, op: 'typeRecord', args: { id: 3 } }],
      },
      ac.signal,
      transport,
    );
    expect(calls).toBe(3);

    // Reading 1 is still a cache hit; reading 2 misses and invokes transport.
    await host.query(
      {
        session: 'lru-1',
        generation: 1,
        queries: [{ generation: 1, op: 'typeRecord', args: { id: 1 } }],
      },
      ac.signal,
      transport,
    );
    expect(calls).toBe(3);
    await host.query(
      {
        session: 'lru-1',
        generation: 1,
        queries: [{ generation: 1, op: 'typeRecord', args: { id: 2 } }],
      },
      ac.signal,
      transport,
    );
    expect(calls).toBe(4);

    // retire() on an in-flight query rejects with AbortError.
    const hangingTransport = async (
      req: TypeServiceHostRequest,
    ): Promise<TypeServiceHostResponse> => {
      if (req.open) return { descriptor };
      return new Promise(() => {
        /* stays pending until retire/close */
      });
    };
    const pendingRetire = host.query(
      {
        session: 'lru-1',
        generation: 1,
        queries: [{ generation: 1, op: 'typeRecord', args: { id: 99 } }],
      },
      ac.signal,
      hangingTransport,
    );
    host.retire(['lru-1']);
    await expect(pendingRetire).rejects.toMatchObject({ name: 'AbortError' });

    // close() on an in-flight query rejects with TypeServiceClosedError.
    await host.query(
      { session: 'lru-2', generation: 1, open: true },
      ac.signal,
      hangingTransport,
    );
    const pendingClose = host.query(
      {
        session: 'lru-2',
        generation: 1,
        queries: [{ generation: 1, op: 'typeRecord', args: { id: 100 } }],
      },
      ac.signal,
      hangingTransport,
    );
    host.close();
    await expect(pendingClose).rejects.toMatchObject({
      name: 'TypeServiceClosedError',
    });
  });

  test('respects Go non-cacheable RTS5 bit-31 flag, detaches base64 slabs, preserves close() precedence, and aborts in-flight queries', async () => {
    const host = new PluginTypeServices();
    const ac = new AbortController();

    // Build an RTS5 packet with bit 31 set (non-cacheable mutable symbolRecord).
    const nonCacheable = new Uint8Array(12);
    const ncView = new DataView(nonCacheable.buffer);
    ncView.setUint32(0, 0x35535452, true); // "RTS5"
    ncView.setUint32(4, 6 | 0x80000000, true); // packetSymbol | non-cacheable
    ncView.setUint32(8, 0, true);

    let symCalls = 0;
    let directSignal: AbortSignal | undefined;
    const transport = async (
      req: TypeServiceHostRequest,
      signal?: AbortSignal,
    ): Promise<TypeServiceHostResponse> => {
      if (req.open) return { descriptor };
      if (req.queries?.[0].op === 'symbolRecord') {
        symCalls++;
        // Return base64 so we also assert slab detachment on small packets.
        return { packets: [Buffer.from(nonCacheable).toString('base64')] };
      }
      directSignal = signal;
      return new Promise((_resolve, reject) => {
        signal?.addEventListener(
          'abort',
          () => {
            // Simulate an AbortSignal-aware transport rejecting with AbortError
            // ahead of PluginTypeServices' own listener.
            reject(new DOMException('transport aborted', 'AbortError'));
          },
          { once: true },
        );
      });
    };

    await host.query(
      { session: 'guard-1', generation: 1, open: true },
      ac.signal,
      transport,
    );

    const r1 = await host.query(
      {
        session: 'guard-1',
        generation: 1,
        queries: [{ generation: 1, op: 'symbolRecord', args: { id: 5 } }],
      },
      ac.signal,
      transport,
    );
    const r2 = await host.query(
      {
        session: 'guard-1',
        generation: 1,
        queries: [{ generation: 1, op: 'symbolRecord', args: { id: 5 } }],
      },
      ac.signal,
      transport,
    );
    // Bit-31 non-cacheable packet must NEVER be served from host cache.
    expect(symCalls).toBe(2);
    // Decoded base64 packet must be detached from Node's 8 KiB Buffer pool slab.
    expect(r1.packets![0].buffer.byteLength).toBe(nonCacheable.byteLength);
    expect(r2.packets![0].buffer.byteLength).toBe(nonCacheable.byteLength);

    // Direct-bypass semanticDiagnostics query must abort on host.close() and
    // preserve TypeServiceClosedError even when transport rejects with AbortError.
    const diagPromise = host.query(
      {
        session: 'guard-1',
        generation: 1,
        queries: [{ generation: 1, op: 'semanticDiagnostics', args: {} }],
      },
      ac.signal,
      transport,
    );
    host.close();
    expect(directSignal?.aborted).toBe(true);
    await expect(diagPromise).rejects.toMatchObject({
      name: 'TypeServiceClosedError',
    });
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

describe('encodeBinaryTypeQuery', () => {
  test('encodes compact binary shapes, UTF-16 lone surrogates, tokens, optional flags, and excludes ProgramRequest/suggestionDiagnostics', () => {
    const dst = new Uint8Array(4096);
    const files = new Map<string, number>([['/repo/a.ts', 2]]);

    // Excluded ops return 0 so callers fall back to HostTypeBridge.
    expect(
      encodeBinaryTypeQuery(
        dst,
        'demo',
        { generation: 1, op: 'rootFileNames', args: {} },
        files,
      ),
    ).toBe(0);
    expect(
      encodeBinaryTypeQuery(
        dst,
        'demo',
        {
          generation: 1,
          op: 'suggestionDiagnostics',
          args: { node: { file: '/repo/a.ts', index: 1 } },
        },
        files,
      ),
    ).toBe(0);

    // Compact NodeRequest with indexed file + token + skipRecords.
    const n1 = encodeBinaryTypeQuery(
      dst,
      'demo-1',
      {
        generation: 7,
        op: 'type',
        skipRecords: true,
        args: {
          node: {
            file: '/repo/a.ts',
            index: 42,
            token: { kind: 80, pos: 10, end: 15 },
          },
        },
      },
      files,
    );
    expect(n1).toBe(2 + 6 + 4 + 1 + 1 + 4 + 4 + 1 + 12);
    const view = new DataView(dst.buffer);
    expect(view.getUint16(0, true)).toBe(6);
    expect(view.getUint32(8, true)).toBe(7);
    expect(dst[12]).toBe(1); // skipRecords
    expect(dst[13]).toBe(20); // op 'type'
    expect(view.getUint32(14, true)).toBe(2); // fileIdx
    expect(view.getUint32(18, true)).toBe(42);
    expect(dst[22]).toBe(1); // hasToken

    // PropertyRequest with lone UTF-16 surrogate uses UTF-16 flag (bit 31).
    const n2 = encodeBinaryTypeQuery(
      dst,
      's',
      {
        generation: 1,
        op: 'property',
        args: { id: 9, name: [0x61, 0xd800, 0x62] },
      },
      files,
    );
    expect(n2).toBe(2 + 1 + 4 + 1 + 1 + 4 + 4 + 6);
    expect(view.getUint32(13, true)).toBe(0x8000_0003);
    expect(view.getUint16(19, true)).toBe(0xd800);

    // PrintRequest distinguishes undefined flags vs explicit flags: 0.
    const nNil = encodeBinaryTypeQuery(
      dst,
      's',
      {
        generation: 1,
        op: 'typeString',
        args: { id: 5, node: undefined, flags: undefined },
      },
      files,
    );
    const nZero = encodeBinaryTypeQuery(
      dst,
      's',
      {
        generation: 1,
        op: 'typeString',
        args: { id: 5, node: undefined, flags: 0 },
      },
      files,
    );
    expect(nZero).toBe(nNil + 4);

    // Tiny buffer overflow cleanly returns 0 without throwing.
    expect(
      encodeBinaryTypeQuery(
        new Uint8Array(8),
        'demo-1',
        { generation: 1, op: 'void', args: {} },
        files,
      ),
    ).toBe(0);
  });
});

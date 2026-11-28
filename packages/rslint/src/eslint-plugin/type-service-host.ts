import {
  DIRECT_RECORD_OPS,
  isCacheableTypeQuery,
  typeQueryKey,
  type TypeServiceDescriptor,
  type TypeServiceQuery,
} from '../languages/js/type-service-protocol.js';
import {
  isPacketCacheable,
  isRecordBundle,
  typeReplyPackets,
} from '../languages/js/type-service-codec.js';
import type {
  TypeServiceHostRequest,
  TypeServiceHostResponse,
  TypeServiceHostTransport,
  TypeServiceWorkerRequest,
  TypeServiceWorkerResponse,
} from './plugin/type-service-protocol.js';

const MAX_SESSION_CACHE_BYTES = 8 << 20;
const MAX_HOST_CACHE_BYTES = 32 << 20;

interface ActiveSession {
  descriptor: TypeServiceDescriptor;
  cache: Map<string, Uint8Array>;
  cacheBytes: number;
}

interface Session {
  generation: number;
  active?: ActiveSession;
  opening?: Promise<ActiveSession>;
  openAbort?: AbortController;
  openWaiters: number;
  activeAborts: Set<AbortController>;
}

function detach(value: Uint8Array): Uint8Array {
  if (value.byteOffset === 0 && value.byteLength === value.buffer.byteLength) {
    return value;
  }
  const copy = new Uint8Array(value.byteLength);
  copy.set(value);
  return copy;
}

// Detach pooled base64 slabs (Node's 8 KiB Buffer pool) and SHM slices once so
// MessagePort structured cloning and host caching never amplify small packets.
const bytes = (value: string | Uint8Array): Uint8Array =>
  detach(typeof value === 'string' ? Buffer.from(value, 'base64') : value);

function closedError(): Error {
  const error = new Error('type service session is stale or closed');
  error.name = 'TypeServiceClosedError';
  return error;
}

function abortError(message = 'type service request cancelled'): DOMException {
  return new DOMException(message, 'AbortError');
}

function signalError(signal: AbortSignal): Error {
  const reason: unknown = signal.reason;
  return reason instanceof Error && reason.name === 'TypeServiceClosedError'
    ? reason
    : abortError();
}

async function awaitWithSignal<T>(
  promise: Promise<T>,
  signal: AbortSignal,
): Promise<T> {
  if (signal.aborted) {
    void promise.catch(() => undefined);
    throw signalError(signal);
  }
  let onAbort: (() => void) | undefined;
  try {
    return await Promise.race([
      promise,
      new Promise<never>((_resolve, reject) => {
        onAbort = () => {
          reject(signalError(signal));
        };
        signal.addEventListener('abort', onAbort, { once: true });
      }),
    ]);
  } finally {
    if (onAbort) signal.removeEventListener('abort', onAbort);
  }
}

function extractProjection(
  requested: string | undefined,
  result: TypeServiceHostResponse | undefined,
): TypeServiceWorkerResponse['projection'] {
  const projection = result?.projection;
  if (requested !== undefined && !projection) {
    throw new Error('missing type service projection');
  }
  return projection
    ? { ...projection, tree: bytes(projection.tree) }
    : undefined;
}

/**
 * The host owns session opening and a bounded LRU cache of immutable facts for
 * JSON-RPC / fallback paths (LSP, syntax-only, program diagnostics, and non-SHM
 * environments) and retires cleanly when Go ends the semantic lease.
 */
export class PluginTypeServices {
  private readonly sessions = new Map<string, Session>();
  private retainedBytes = 0;
  private closed = false;

  private assertActive(
    sessionId: string,
    owner: Session,
    syntaxOnly: boolean | undefined,
    signal: AbortSignal,
  ): void {
    if (this.closed) throw closedError();
    if (signal.aborted) throw signalError(signal);
    if (!syntaxOnly && this.sessions.get(sessionId) !== owner) {
      throw abortError('type service request cancelled or retired');
    }
  }

  private async send(
    sessionId: string,
    owner: Session,
    syntaxOnly: boolean | undefined,
    request: TypeServiceHostRequest,
    transport: TypeServiceHostTransport,
    signal: AbortSignal,
  ): Promise<TypeServiceHostResponse> {
    this.assertActive(sessionId, owner, syntaxOnly, signal);
    let result: TypeServiceHostResponse;
    try {
      result = await awaitWithSignal(transport(request, signal), signal);
    } catch (error) {
      // Preserve lifecycle close/abort precedence even if an AbortSignal-aware
      // transport registered its own listener first and rejected ahead of us.
      if (this.closed) throw closedError();
      if (signal.aborted) throw signalError(signal);
      throw error;
    }
    if (result.status === 'closed') {
      if (!syntaxOnly && this.sessions.get(sessionId) === owner) {
        this.retireSession(sessionId, closedError());
      }
      throw closedError();
    }
    if (result.status === 'cancelled') throw abortError();
    this.assertActive(sessionId, owner, syntaxOnly, signal);
    return result;
  }

  private async sendLinked(
    sessionId: string,
    owner: Session,
    syntaxOnly: boolean | undefined,
    request: TypeServiceHostRequest,
    transport: TypeServiceHostTransport,
    signal: AbortSignal,
  ): Promise<TypeServiceHostResponse> {
    if (syntaxOnly) {
      return this.send(sessionId, owner, true, request, transport, signal);
    }
    const linked = new AbortController();
    owner.activeAborts.add(linked);
    const onCallerAbort = () => {
      linked.abort(signal.reason);
    };
    if (signal.aborted) {
      linked.abort(signal.reason);
    } else {
      signal.addEventListener('abort', onCallerAbort, { once: true });
    }
    try {
      return await this.send(
        sessionId,
        owner,
        false,
        request,
        transport,
        linked.signal,
      );
    } finally {
      signal.removeEventListener('abort', onCallerAbort);
      owner.activeAborts.delete(linked);
    }
  }

  private readCached(
    active: ActiveSession,
    key: string,
  ): Uint8Array | undefined {
    const cached = active.cache.get(key);
    if (!cached) return undefined;
    // Refresh LRU recency so hot library facts are not evicted ahead of cold entries.
    active.cache.delete(key);
    active.cache.set(key, cached);
    return cached;
  }

  private cachePacket(
    active: ActiveSession,
    key: string,
    value: Uint8Array,
  ): void {
    // Facts for a given (generation, queryKey) are immutable: refresh LRU order
    // without allocating or mutating byte accounting when already cached.
    const prev = active.cache.get(key);
    if (prev) {
      active.cache.delete(key);
      active.cache.set(key, prev);
      return;
    }
    const size = key.length * 2 + value.byteLength;
    if (
      size > MAX_SESSION_CACHE_BYTES ||
      this.retainedBytes - active.cacheBytes + size > MAX_HOST_CACHE_BYTES
    ) {
      return;
    }
    const owned = detach(value);
    while (
      active.cache.size > 0 &&
      (active.cacheBytes + size > MAX_SESSION_CACHE_BYTES ||
        this.retainedBytes + size > MAX_HOST_CACHE_BYTES)
    ) {
      const oldest = active.cache.entries().next().value;
      if (!oldest) break;
      const oldestSize = oldest[0].length * 2 + oldest[1].byteLength;
      active.cache.delete(oldest[0]);
      active.cacheBytes -= oldestSize;
      this.retainedBytes -= oldestSize;
    }
    active.cache.set(key, owned);
    active.cacheBytes += size;
    this.retainedBytes += size;
  }

  private storeQueryReply(
    active: ActiveSession,
    query: TypeServiceQuery,
    raw: Uint8Array,
    cacheKey?: string,
  ): void {
    if (
      !query.skipRecords &&
      DIRECT_RECORD_OPS.has(query.op) &&
      isRecordBundle(raw)
    ) {
      const reply = typeReplyPackets(raw);
      let allRecordsCacheable = true;
      for (const record of reply.records) {
        if (!isPacketCacheable(record.packet)) {
          allRecordsCacheable = false;
          continue;
        }
        const recordKey = typeQueryKey({
          generation: query.generation,
          op: record.op,
          args: { id: record.id },
        });
        // Skip warm supplemental records without copying or churning the LRU map.
        if (active.cache.has(recordKey)) continue;
        this.cachePacket(active, recordKey, record.packet);
      }
      if (cacheKey !== undefined && isPacketCacheable(reply.packet)) {
        // Never retain a bundle carrying a mutable supplemental record; store
        // only the immutable scalar handle packet so later lookups stay fresh.
        this.cachePacket(
          active,
          cacheKey,
          allRecordsCacheable ? raw : reply.packet,
        );
      }
      return;
    }
    if (cacheKey !== undefined && isPacketCacheable(raw)) {
      this.cachePacket(active, cacheKey, raw);
    }
  }

  private async openSession(
    sessionId: string,
    generation: number,
    owner: Session,
    transport: TypeServiceHostTransport,
    signal: AbortSignal,
  ): Promise<ActiveSession> {
    if (owner.active) return owner.active;
    let opening = owner.opening;
    if (!opening) {
      const openAbort = new AbortController();
      owner.openAbort = openAbort;
      // Tie the shared open RPC to session lifetime (and abort if all waiters
      // cancel) rather than one leader task's signal or per-file projection.
      const run = (async (): Promise<ActiveSession> => {
        try {
          const result = await this.send(
            sessionId,
            owner,
            false,
            { session: sessionId, generation, open: true },
            transport,
            openAbort.signal,
          );
          if (result.descriptor?.generation !== generation) {
            throw new Error('missing or foreign type service descriptor');
          }
          const active: ActiveSession = {
            descriptor: result.descriptor,
            cache: new Map(),
            cacheBytes: 0,
          };
          owner.active = active;
          return active;
        } catch (error) {
          if (!owner.active && this.sessions.get(sessionId) === owner) {
            this.sessions.delete(sessionId);
          }
          throw error;
        } finally {
          if (owner.openAbort === openAbort) {
            owner.opening = undefined;
            owner.openAbort = undefined;
          }
        }
      })();
      void run.catch(() => undefined);
      owner.opening = run;
      opening = run;
    }
    owner.openWaiters++;
    try {
      const active = await awaitWithSignal(opening, signal);
      this.assertActive(sessionId, owner, false, signal);
      return active;
    } finally {
      owner.openWaiters--;
      if (
        owner.openWaiters === 0 &&
        !owner.active &&
        owner.opening === opening
      ) {
        const toAbort = owner.openAbort;
        owner.opening = undefined;
        owner.openAbort = undefined;
        if (this.sessions.get(sessionId) === owner) {
          this.sessions.delete(sessionId);
        }
        toAbort?.abort();
      }
    }
  }

  private async executeQueryBatch(
    sessionId: string,
    owner: Session,
    active: ActiveSession,
    request: TypeServiceHostRequest,
    queries: readonly TypeServiceQuery[],
    transport: TypeServiceHostTransport,
    signal: AbortSignal,
  ): Promise<TypeServiceWorkerResponse> {
    const resolved = new Array<Uint8Array>(queries.length);
    const missingIndices: number[] = [];
    const missingQueries: TypeServiceQuery[] = [];
    const missingKeys: (string | undefined)[] = [];
    for (let i = 0; i < queries.length; i++) {
      const query = queries[i];
      const key = isCacheableTypeQuery(query) ? typeQueryKey(query) : undefined;
      if (key !== undefined) {
        const cached = this.readCached(active, key);
        if (cached) {
          resolved[i] = cached;
          continue;
        }
      }
      missingIndices.push(i);
      missingQueries.push(query);
      missingKeys.push(key);
    }
    if (missingQueries.length === 0) {
      return { packets: resolved };
    }
    const result = await this.sendLinked(
      sessionId,
      owner,
      false,
      { ...request, queries: missingQueries },
      transport,
      signal,
    );
    if (result.packets?.length !== missingQueries.length) {
      throw new Error('incomplete type service batch reply');
    }
    for (let j = 0; j < missingQueries.length; j++) {
      const raw = bytes(result.packets[j]);
      this.storeQueryReply(active, missingQueries[j], raw, missingKeys[j]);
      resolved[missingIndices[j]] = raw;
    }
    return { packets: resolved };
  }

  async query(
    { syntaxOnly, ...request }: TypeServiceWorkerRequest,
    signal: AbortSignal,
    transport: TypeServiceHostTransport,
  ): Promise<TypeServiceWorkerResponse> {
    if (this.closed) throw closedError();
    if (signal.aborted) throw signalError(signal);

    let owner: Session | undefined = syntaxOnly
      ? {
          generation: request.generation,
          openWaiters: 0,
          activeAborts: new Set(),
        }
      : this.sessions.get(request.session);
    if (!owner) {
      if (!request.open) throw closedError();
      owner = {
        generation: request.generation,
        openWaiters: 0,
        activeAborts: new Set(),
      };
      this.sessions.set(request.session, owner);
    }
    if (owner.generation !== request.generation) {
      throw new Error('foreign type service query generation');
    }

    if (request.open || request.projection !== undefined) {
      if (
        request.open &&
        !syntaxOnly &&
        owner.active &&
        request.projection === undefined
      ) {
        return { descriptor: owner.active.descriptor };
      }
      // Coalesce concurrent open requests across workers sharing one typed
      // Program session while isolating per-file projections and task signals.
      if (request.open && !syntaxOnly && !owner.active) {
        const openPromise = this.openSession(
          request.session,
          request.generation,
          owner,
          transport,
          signal,
        );
        const projectionAbort =
          request.projection !== undefined ? new AbortController() : undefined;
        const onCallerAbort = projectionAbort
          ? () => {
              projectionAbort.abort(signal.reason);
            }
          : undefined;
        if (onCallerAbort) {
          signal.addEventListener('abort', onCallerAbort, { once: true });
        }
        void openPromise.catch(() => {
          projectionAbort?.abort();
        });
        const projectionPromise =
          projectionAbort && request.projection !== undefined
            ? this.sendLinked(
                request.session,
                owner,
                false,
                { ...request, open: undefined },
                transport,
                projectionAbort.signal,
              )
            : undefined;
        void projectionPromise?.catch(() => undefined);
        try {
          const active = await openPromise;
          const projectionResult = projectionPromise
            ? await projectionPromise
            : undefined;
          this.assertActive(request.session, owner, false, signal);
          return {
            descriptor: active.descriptor,
            projection: extractProjection(request.projection, projectionResult),
          };
        } finally {
          if (onCallerAbort) {
            signal.removeEventListener('abort', onCallerAbort);
          }
        }
      }
      const result = await this.sendLinked(
        request.session,
        owner,
        syntaxOnly,
        request.open && owner.active
          ? { ...request, open: undefined }
          : request,
        transport,
        signal,
      );
      if (
        request.open &&
        syntaxOnly &&
        result.descriptor?.generation !== request.generation
      ) {
        throw new Error('missing or foreign type service descriptor');
      }
      return {
        descriptor: request.open
          ? syntaxOnly
            ? result.descriptor
            : owner.active?.descriptor
          : undefined,
        projection: extractProjection(request.projection, result),
      };
    }

    if (request.source !== undefined) {
      const result = await this.sendLinked(
        request.session,
        owner,
        syntaxOnly,
        request,
        transport,
        signal,
      );
      return result.source
        ? {
            source: {
              fileName: result.source.fileName,
              path: result.source.path,
              tree: result.source.tree ? bytes(result.source.tree) : undefined,
            },
          }
        : {};
    }

    const queries = request.queries ?? [];
    if (syntaxOnly) {
      const result = await this.send(
        request.session,
        owner,
        true,
        request,
        transport,
        signal,
      );
      if (result.packets?.length !== queries.length) {
        throw new Error('incomplete type service batch reply');
      }
      return { packets: result.packets.map(bytes) };
    }

    const active = owner.active;
    if (!active) throw new Error('type service descriptor has not been opened');

    for (const q of queries) {
      if (q.generation !== owner.generation) {
        throw new Error('foreign type service query generation');
      }
    }

    return this.executeQueryBatch(
      request.session,
      owner,
      active,
      request,
      queries,
      transport,
      signal,
    );
  }

  private retireSession(id: string, reason?: Error): void {
    const owner = this.sessions.get(id);
    if (!owner) return;
    this.sessions.delete(id);
    const toAbort = owner.openAbort;
    owner.openAbort = undefined;
    owner.opening = undefined;
    if (owner.active) {
      this.retainedBytes -= owner.active.cacheBytes;
      owner.active.cache.clear();
      owner.active.cacheBytes = 0;
    }
    for (const abort of [...owner.activeAborts]) {
      abort.abort(reason);
    }
    owner.activeAborts.clear();
    toAbort?.abort(reason);
  }

  retire(sessions: readonly string[]): void {
    for (const id of sessions) {
      this.retireSession(id);
    }
  }

  close(): void {
    this.closed = true;
    const reason = closedError();
    for (const id of [...this.sessions.keys()]) {
      this.retireSession(id, reason);
    }
  }
}

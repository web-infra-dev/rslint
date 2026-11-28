/**
 * Native bidirectional shared-memory RPC transport owned by one Node IPC session.
 * Mirrors `memory-transport.ts`: zero bytes allocated before `configure()`,
 * feature-detected cleanly when an older or mocked native addon lacks RPC exports,
 * and exposes only a pure `RpcHostChannel` interface to `WorkerPool`.
 */
import {
  getNativeBinding,
  type NativeRpcMapping,
  type RpcConfiguration,
  type RpcHostChannel,
  type WorkerRpcDescriptor,
} from '../native/binding.js';

const MAX_U32 = 0xffff_ffff;

export interface RpcWireMapping {
  version: number;
  fd?: number;
  doorbellFd?: number;
  handle?: string;
  nodeToGoEvent?: string;
  goToNodeEvent?: string;
  processId?: number;
}

function record(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

function positiveU32(value: unknown): value is number {
  return (
    typeof value === 'number' &&
    Number.isInteger(value) &&
    value > 0 &&
    value <= MAX_U32
  );
}

export function parseRpcConfiguration(
  value: unknown,
): RpcConfiguration | undefined {
  if (!record(value)) return undefined;
  if (
    !positiveU32(value.version) ||
    !positiveU32(value.slotCount) ||
    !positiveU32(value.slotSize) ||
    !positiveU32(value.headerSize) ||
    !positiveU32(value.slotHeaderStride)
  ) {
    return undefined;
  }
  return {
    version: value.version,
    slotCount: value.slotCount,
    slotSize: value.slotSize,
    headerSize: value.headerSize,
    slotHeaderStride: value.slotHeaderStride,
  };
}

function toWireMapping(raw: NativeRpcMapping): RpcWireMapping {
  const out: RpcWireMapping = { version: raw.version };
  if (typeof raw.fd === 'number') out.fd = raw.fd;
  if (typeof raw.doorbellFd === 'number') out.doorbellFd = raw.doorbellFd;
  if (typeof raw.handle === 'string') out.handle = raw.handle;
  if (typeof raw.nodeToGoEvent === 'string')
    out.nodeToGoEvent = raw.nodeToGoEvent;
  if (typeof raw.goToNodeEvent === 'string')
    out.goToNodeEvent = raw.goToNodeEvent;
  if (typeof raw.processId === 'number') out.processId = raw.processId;
  return out;
}

export function createRpcTransport() {
  const binding = getNativeBinding();
  if (
    typeof binding.RpcChannel !== 'function' ||
    typeof binding.rpcAcquireSlot !== 'function' ||
    typeof binding.rpcReleaseSlot !== 'function' ||
    typeof binding.rpcAbortSlot !== 'function' ||
    typeof binding.rpcCallSync !== 'function'
  ) {
    return undefined;
  }
  const { RpcChannel, rpcAcquireSlot, rpcReleaseSlot, rpcAbortSlot } = binding;
  const channel = new RpcChannel();
  let fd: number | undefined;
  let doorbellFd: number | undefined;
  try {
    fd = channel.fd() ?? undefined;
    doorbellFd = channel.doorbellFd() ?? undefined;
  } catch (error) {
    channel.close();
    throw error;
  }
  let closed = false;
  let configuredChannelId: number | undefined;

  function close() {
    if (closed) return;
    closed = true;
    configuredChannelId = undefined;
    channel.close();
  }

  return {
    fd() {
      return closed ? undefined : fd;
    },
    doorbellFd() {
      return closed ? undefined : doorbellFd;
    },
    afterSpawn() {
      if (!closed) {
        doorbellFd = undefined;
        channel.afterSpawn();
      }
    },
    configure(
      config: RpcConfiguration,
      inheritedFd?: number,
      inheritedDoorbellFd?: number,
      peerPid?: number,
    ): { mapping: RpcWireMapping; hostChannel: RpcHostChannel } {
      try {
        if (closed) throw new Error('RPC shared memory transport is closed');
        const raw = channel.configure(
          config,
          inheritedFd ?? null,
          inheritedDoorbellFd ?? null,
          peerPid ?? null,
        );
        const id = channel.channelId();
        if (typeof id !== 'number' || id <= 0) {
          throw new Error('RPC channel registration failed');
        }
        configuredChannelId = id;
        const slotSize = config.slotSize;
        const hostChannel: RpcHostChannel = {
          slotSize,
          acquireSlot(): WorkerRpcDescriptor | undefined {
            if (closed || configuredChannelId !== id) return undefined;
            const lease = rpcAcquireSlot(id);
            if (!lease) return undefined;
            return {
              channelId: id,
              slot: lease.slot,
              epoch: lease.epoch,
              slotSize,
            };
          },
          abortSlot(slot: number): void {
            if (closed || configuredChannelId !== id) return;
            rpcAbortSlot(id, slot);
          },
          releaseSlot(slot: number): void {
            if (closed || configuredChannelId !== id) return;
            rpcReleaseSlot(id, slot);
          },
        };
        return { mapping: toWireMapping(raw), hostChannel };
      } catch (error) {
        close();
        throw error;
      }
    },
    close,
  };
}

export type RpcTransport = NonNullable<ReturnType<typeof createRpcTransport>>;

type ServiceRpcOpener = () => Promise<RpcHostChannel | undefined>;

const serviceRpcOpeners = new WeakMap<object, ServiceRpcOpener>();

/**
 * Internal association between a Node service instance and its IPC RPC opener.
 * Kept out of public class shapes so `@rslint/core/internal` `.d.ts` stays unchanged.
 */
export function registerServiceRpcOpener(
  service: object,
  opener: ServiceRpcOpener,
): void {
  serviceRpcOpeners.set(service, opener);
}

export async function openServiceRpc(
  service: object,
): Promise<RpcHostChannel | undefined> {
  const opener = serviceRpcOpeners.get(service);
  return opener ? opener() : undefined;
}

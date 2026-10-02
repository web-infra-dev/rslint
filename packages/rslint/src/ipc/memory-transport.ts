/** Native storage for generic IPC byte attachments. No application payloads. */
import {
  getNativeBinding,
  type MemoryConfiguration,
} from '../native/binding.js';
import type {
  IpcAttachment,
  MemoryBatch,
  MemoryMapping,
  MemoryRange,
} from './protocol.js';

// Publication words and native configuration fields are represented as u32.
// This bounds their representation; the peer selects every layout dimension.
const MAX_U32 = 0xffff_ffff;

function record(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

function uint(value: unknown, max: number): value is number {
  return (
    typeof value === 'number' &&
    Number.isInteger(value) &&
    value >= 0 &&
    value <= max
  );
}

function configurationNumber(value: unknown): number {
  if (!uint(value, MAX_U32) || value === 0) {
    throw new Error('invalid shared memory configuration');
  }
  return value;
}

export function createMemoryTransport() {
  const { MemoryArena } = getNativeBinding();
  const arena = new MemoryArena();
  let fd: number | undefined;
  try {
    // Only prepare the inheritable Unix handle; no layout is known yet.
    fd = arena.fd() ?? undefined;
  } catch (error) {
    arena.close();
    throw error;
  }
  let configuration: MemoryConfiguration | undefined;
  let mapping: MemoryMapping | undefined;
  let closed = false;
  function close() {
    if (closed) return;
    closed = true;
    configuration = undefined;
    mapping = undefined;
    arena.close();
  }
  return {
    fd() {
      return closed ? undefined : fd;
    },
    configure(value: unknown) {
      try {
        if (closed) throw new Error('shared memory storage is closed');
        if (!record(value)) {
          throw new Error('invalid shared memory configuration');
        }
        const config: MemoryConfiguration = {
          version: configurationNumber(value.version),
          slotCount: configurationNumber(value.slotCount),
          slotSize: configurationNumber(value.slotSize),
          headerSize: configurationNumber(value.headerSize),
          publicationStride: configurationNumber(value.publicationStride),
        };
        arena.configure(config);
        const descriptor = arena.descriptor();
        if (descriptor.version !== config.version) {
          throw new Error('inconsistent shared memory version');
        }
        configuration = config;
        mapping = {
          version: descriptor.version,
          fd: descriptor.fd ?? undefined,
          handle: descriptor.handle ?? undefined,
          processId: descriptor.processId ?? undefined,
        };
      } catch (error) {
        close();
        throw error;
      }
    },
    configuration() {
      return configuration;
    },
    descriptor(inheritedFd?: number): MemoryMapping {
      if (!mapping) throw new Error('shared memory storage is not configured');
      // Unix spawn chooses the child's fd index. Windows transfers the handle.
      return {
        ...mapping,
        fd:
          typeof mapping.fd === 'number'
            ? (inheritedFd ?? mapping.fd)
            : undefined,
      };
    },
    register(batches: MemoryBatch[]) {
      return arena.register(batches);
    },
    release(lease: number) {
      return arena.release(lease);
    },
    close,
  };
}

export type MemoryTransport = ReturnType<typeof createMemoryTransport>;

export interface ReceivedAttachments {
  values: IpcAttachment[] | undefined;
  /** Revoke now; false requires a retry after native readers return. */
  release(): MemoryBatch[] | false | undefined;
}

/**
 * Read shared capabilities while their inbound handler is alive. The returned
 * owned Buffer can outlive the handler; it never exposes writable mapped memory.
 */
export function readAttachmentBytes(value: IpcAttachment): Buffer {
  if (typeof value === 'string') return Buffer.from(value, 'utf8');
  if (Buffer.isBuffer(value)) return value;
  if (value instanceof Uint8Array) return Buffer.from(value);
  return getNativeBinding().readBytes(value);
}

function decodeBytes(value: string): Buffer {
  const padding = value.indexOf('=');
  if (
    value.length % 4 !== 0 ||
    /[^A-Za-z0-9+/=]/.test(value) ||
    (padding !== -1 &&
      value.slice(padding) !== '=' &&
      value.slice(padding) !== '==') ||
    (padding !== -1 &&
      Buffer.from(value.slice(-4), 'base64').toString('base64') !==
        value.slice(-4))
  ) {
    throw new Error('invalid IPC attachment bytes');
  }
  return Buffer.from(value, 'base64');
}

/** Validate the complete envelope before admitting a native reader. */
export function receiveAttachments(
  input: unknown,
  rawBatches: unknown,
  memory?: MemoryTransport,
): ReceivedAttachments {
  let batches: MemoryBatch[] | undefined;
  let length = 0;
  if (rawBatches !== undefined) {
    const configuration = memory?.configuration();
    if (!configuration)
      throw new Error('shared IPC attachments are unavailable');
    if (
      !Array.isArray(rawBatches) ||
      rawBatches.length === 0 ||
      rawBatches.length > configuration.slotCount
    ) {
      throw new Error('invalid IPC attachment batches');
    }
    const slots = new Set<number>();
    batches = rawBatches.map((batch: unknown) => {
      if (
        !record(batch) ||
        !uint(batch.slot, configuration.slotCount - 1) ||
        !uint(batch.generation, MAX_U32) ||
        batch.generation === 0 ||
        !uint(batch.length, configuration.slotSize) ||
        slots.has(batch.slot) ||
        batch.length > MAX_U32 - length
      ) {
        throw new Error('invalid IPC attachment batch');
      }
      slots.add(batch.slot);
      length += batch.length;
      return {
        slot: batch.slot,
        generation: batch.generation,
        length: batch.length,
      };
    });
  }
  if (input === undefined && batches === undefined) {
    return { values: undefined, release: () => undefined };
  }
  if (!Array.isArray(input)) throw new Error('invalid IPC attachments');
  let end = 0;
  let ranges = 0;
  const validated: Array<string | Buffer | MemoryRange> = input.map(
    (attachment: unknown) => {
      if (!record(attachment)) throw new Error('invalid IPC attachment');
      const variants = [
        attachment.text,
        attachment.bytes,
        attachment.range,
      ].filter((value) => value !== undefined).length;
      if (variants !== 1) throw new Error('invalid IPC attachment');
      if (typeof attachment.text === 'string') return attachment.text;
      if (typeof attachment.bytes === 'string')
        return decodeBytes(attachment.bytes);
      const range = attachment.range;
      if (
        !batches ||
        !record(range) ||
        !uint(range.offset, length) ||
        !uint(range.length, length - range.offset) ||
        range.offset !== end
      ) {
        throw new Error('invalid IPC attachment range');
      }
      end += range.length;
      ranges++;
      return { offset: range.offset, length: range.length };
    },
  );
  if (!batches) {
    return {
      values: validated.map((value) => {
        if (typeof value !== 'string' && !Buffer.isBuffer(value))
          throw new Error('invalid IPC attachment');
        return value;
      }),
      release: () => undefined,
    };
  }
  if (ranges === 0 || end !== length)
    throw new Error('incomplete IPC attachment batches');
  if (!memory) throw new Error('shared IPC attachments are unavailable');
  const lease = memory.register(batches);
  let active = true;
  return {
    values: validated.map((value) =>
      typeof value === 'string' || Buffer.isBuffer(value)
        ? value
        : { ...value, lease },
    ),
    release() {
      if (!active) return undefined;
      if (!memory.release(lease)) return false;
      active = false;
      return batches;
    },
  };
}

/** Native storage for generic IPC text attachments. No application payloads. */
import {
  getNativeBinding,
  type SourceConfiguration,
} from '../native/binding.js';
import type {
  IpcAttachment,
  SourceBatch,
  SourceMapping,
  SourceRange,
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
    throw new Error('invalid shared source configuration');
  }
  return value;
}

export function createSourceTransport() {
  const { SourceArena } = getNativeBinding();
  const arena = new SourceArena();
  let fd: number | undefined;
  try {
    // Only prepare the inheritable Unix handle; no layout is known yet.
    fd = arena.fd() ?? undefined;
  } catch (error) {
    arena.close();
    throw error;
  }
  let configuration: SourceConfiguration | undefined;
  let mapping: SourceMapping | undefined;
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
        if (closed) throw new Error('shared source storage is closed');
        if (!record(value)) {
          throw new Error('invalid shared source configuration');
        }
        const config: SourceConfiguration = {
          version: configurationNumber(value.version),
          slotCount: configurationNumber(value.slotCount),
          slotSize: configurationNumber(value.slotSize),
          headerSize: configurationNumber(value.headerSize),
          publicationStride: configurationNumber(value.publicationStride),
        };
        arena.configure(config);
        const descriptor = arena.descriptor();
        if (descriptor.version !== config.version) {
          throw new Error('inconsistent shared source version');
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
    descriptor(inheritedFd?: number): SourceMapping {
      if (!mapping) throw new Error('shared source storage is not configured');
      // Unix spawn chooses the child's fd index. Windows transfers the handle.
      return {
        ...mapping,
        fd:
          typeof mapping.fd === 'number'
            ? (inheritedFd ?? mapping.fd)
            : undefined,
      };
    },
    register(batch: SourceBatch) {
      return arena.register(batch.slot, batch.generation, batch.length);
    },
    release(lease: number) {
      return arena.release(lease);
    },
    close,
  };
}

export type SourceTransport = ReturnType<typeof createSourceTransport>;

export interface ReceivedAttachments {
  values: IpcAttachment[] | undefined;
  /** Revoke once; only proven release authorizes a matching acknowledgement. */
  release(): SourceBatch | undefined;
}

/** Validate the entire envelope before admitting any native reader. */
export function receiveAttachments(
  input: unknown,
  rawBatch: unknown,
  sources?: SourceTransport,
): ReceivedAttachments {
  let batch: SourceBatch | undefined;
  if (rawBatch !== undefined) {
    const configuration = sources?.configuration();
    if (!configuration)
      throw new Error('shared IPC attachments are unavailable');
    if (
      !record(rawBatch) ||
      !uint(rawBatch.slot, configuration.slotCount - 1) ||
      !uint(rawBatch.generation, MAX_U32) ||
      rawBatch.generation === 0 ||
      !uint(rawBatch.length, configuration.slotSize)
    ) {
      throw new Error('invalid IPC attachment batch');
    }
    batch = {
      slot: rawBatch.slot,
      generation: rawBatch.generation,
      length: rawBatch.length,
    };
  }
  if (input === undefined && batch === undefined) {
    return { values: undefined, release: () => undefined };
  }
  if (!Array.isArray(input)) throw new Error('invalid IPC attachments');

  let end = 0;
  let ranges = 0;
  const validated: Array<string | SourceRange> = input.map((attachment) => {
    if (!record(attachment)) throw new Error('invalid IPC attachment');
    if (typeof attachment.text === 'string' && attachment.range === undefined) {
      return attachment.text;
    }
    const range = attachment.range;
    if (
      attachment.text !== undefined ||
      !batch ||
      !record(range) ||
      !uint(range.offset, batch.length) ||
      !uint(range.length, batch.length - range.offset) ||
      range.offset !== end
    ) {
      throw new Error('invalid IPC attachment range');
    }
    end += range.length;
    ranges++;
    return { offset: range.offset, length: range.length };
  });
  if (!batch) {
    return {
      values: validated.map((value) => {
        if (typeof value !== 'string')
          throw new Error('invalid IPC attachment');
        return value;
      }),
      release: () => undefined,
    };
  }
  if (ranges === 0 || end !== batch.length) {
    throw new Error('incomplete IPC attachment batch');
  }
  if (!sources) throw new Error('shared IPC attachments are unavailable');

  const lease = sources.register(batch);
  let active = true;
  return {
    values: validated.map((value) =>
      typeof value === 'string' ? value : { ...value, lease },
    ),
    release() {
      if (!active) return undefined;
      active = false;
      return sources.release(lease) ? batch : undefined;
    },
  };
}

/** Native storage for generic IPC text attachments. No application payloads. */
import { getNativeBinding } from '../native/binding.js';
import type { IpcAttachment } from './protocol.js';
import {
  INHERITED_FD,
  MAX_GENERATION,
  PROTOCOL_VERSION,
  SLOT_COUNT,
  SLOT_SIZE,
  type SourceBatch,
  type SourceMapping,
  type SourceRange,
} from './protocol.generated.js';

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

export function createSourceTransport() {
  const { SourceArena } = getNativeBinding();
  const arena = new SourceArena();
  let mapping: SourceMapping;
  try {
    mapping = arena.descriptor();
    if (mapping.version !== PROTOCOL_VERSION) {
      throw new Error('unsupported shared source version');
    }
  } catch (error) {
    arena.close();
    throw error;
  }
  return {
    // Unix spawn duplicates the local fd into the protocol's inherited fd.
    // Windows transfers an anonymous mapping handle inside the descriptor.
    fd: mapping.fd,
    descriptor: {
      ...mapping,
      fd: mapping.fd === undefined ? undefined : INHERITED_FD,
    },
    register(batch: SourceBatch) {
      return arena.register(batch.slot, batch.generation, batch.length);
    },
    release(lease: number) {
      return arena.release(lease);
    },
    close() {
      arena.close();
    },
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
    if (
      !record(rawBatch) ||
      !uint(rawBatch.slot, SLOT_COUNT - 1) ||
      !uint(rawBatch.generation, MAX_GENERATION) ||
      rawBatch.generation === 0 ||
      !uint(rawBatch.length, SLOT_SIZE)
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

/**
 * Private CLI transport adapter. This module owns the native arena and turns
 * wire ranges into revocable reader capabilities before dispatching workers.
 * It never reads source bytes or changes the shared linter/API/LSP contract.
 */
import { getNativeBinding } from '../native/binding.js';

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

const SLOT_SIZE = 16 * 1024 * 1024;

export function createSourceTransport() {
  const { SourceArena } = getNativeBinding();
  const arena = new SourceArena();
  const mapping = arena.descriptor();
  return {
    // Unix spawn duplicates the local fd into child fd 3. Windows duplicates
    // an anonymous mapping HANDLE inside Go, avoiding named-object collisions
    // and Node's unsupported Windows descriptor inheritance.
    fd: mapping.fd,
    descriptor: { ...mapping, fd: mapping.fd === undefined ? undefined : 3 },
    close: () => arena.close(),
    async lint(
      input: unknown,
      dispatch: (request: unknown) => Promise<unknown>,
    ): Promise<unknown> {
      if (!record(input) || input.sourceBatch === undefined) {
        if (
          record(input) &&
          Array.isArray(input.files) &&
          input.files.some(
            (file: unknown) =>
              record(file) &&
              (file.sourceRange !== undefined ||
                file.sharedSource !== undefined),
          )
        ) {
          throw new Error('shared source without a batch');
        }
        return dispatch(input);
      }
      if (!Array.isArray(input.files)) {
        throw new Error('invalid plugin source request');
      }
      const batch = input.sourceBatch;
      if (
        !record(batch) ||
        !uint(batch.slot, 15) ||
        !uint(batch.generation, 0xffffffff) ||
        batch.generation === 0 ||
        !uint(batch.length, SLOT_SIZE)
      ) {
        throw new Error('invalid plugin source batch');
      }
      // Validate all ranges before granting any capability. Requiring contiguous
      // ranges also rejects truncated, overlapping or partially described blobs.
      let end = 0;
      let count = 0;
      for (const file of input.files) {
        if (!record(file) || file.sharedSource !== undefined)
          throw new Error('invalid plugin source file');
        if (file.sourceRange === undefined) continue;
        const range = file.sourceRange;
        if (
          !record(range) ||
          file.text !== undefined ||
          !uint(range.offset, batch.length) ||
          !uint(range.length, batch.length - range.offset) ||
          range.offset !== end
        ) {
          throw new Error('invalid plugin source range');
        }
        end += range.length;
        count++;
      }
      if (count === 0 || end !== batch.length)
        throw new Error('incomplete plugin source batch');
      const lease = arena.register(batch.slot, batch.generation, batch.length);
      let released = false;
      let result: unknown;
      try {
        const files = input.files.map((file: Record<string, unknown>) => {
          if (file.sourceRange === undefined) return file;
          const { sourceRange, ...rest } = file;
          return {
            ...rest,
            sharedSource: { ...(sourceRange as object), lease },
          };
        });
        result = await dispatch({ ...input, files });
      } finally {
        released = arena.release(lease);
      }
      if (!record(result)) throw new Error('invalid plugin source result');
      return { ...result, releasedSource: released ? batch : undefined };
    },
  };
}

export type SourceTransport = ReturnType<typeof createSourceTransport>;

/** Associate plugin files with the complete source attachments received by IPC. */
import type { SharedSource } from '../native/binding.js';

function record(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

export function resolvePluginSources(
  input: unknown,
  attachments: readonly (string | SharedSource)[] = [],
): unknown {
  if (!record(input) || !Array.isArray(input.files)) {
    throw new Error('invalid plugin lint request');
  }
  let nextIndex = 0;
  const files = input.files.map((file: unknown) => {
    if (
      !record(file) ||
      file.sharedSource !== undefined ||
      file.sourceRange !== undefined
    ) {
      throw new Error('invalid plugin source file');
    }
    if (file.sourceIndex === undefined) return file;
    const { sourceIndex, ...rest } = file;
    if (
      file.text !== undefined ||
      sourceIndex !== nextIndex ||
      nextIndex >= attachments.length
    ) {
      throw new Error('invalid plugin source attachment');
    }
    const source = attachments[nextIndex++];
    return typeof source === 'string'
      ? { ...rest, text: source }
      : { ...rest, sharedSource: source };
  });
  if (nextIndex !== attachments.length) {
    throw new Error('unused plugin source attachment');
  }
  return { ...input, files };
}

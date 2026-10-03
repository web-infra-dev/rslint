/** Project plugin file indices into inline text or a native byte capability. */
import { TextDecoder } from 'node:util';
import type { ByteInput } from '../../native/types.js';
import type {
  EslintPluginLintRequest,
  ResolvedEslintPluginLintRequest,
} from './plugin-lint-protocol.js';

// Match the native parser's UTF-8 rejection while leaving BOM handling to parsing.
const sourceDecoder = new TextDecoder('utf-8', {
  fatal: true,
  ignoreBOM: true,
});

function record(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

export function resolvePluginAttachments(
  input: EslintPluginLintRequest,
  attachments: readonly ByteInput[] = [],
): ResolvedEslintPluginLintRequest {
  if (!record(input) || !Array.isArray(input.files)) {
    throw new Error('invalid plugin lint request');
  }
  let nextIndex = 0;
  const files = input.files.map(
    (
      file: EslintPluginLintRequest['files'][number],
    ): ResolvedEslintPluginLintRequest['files'][number] => {
      if (
        !record(file) ||
        ('sharedSource' in file && file.sharedSource !== undefined) ||
        ('sourceRange' in file && file.sourceRange !== undefined) ||
        ('sourceIndex' in file && file.sourceIndex !== undefined)
      ) {
        throw new Error('invalid plugin source file');
      }
      const { textAttachment, ...rest } = file;
      if (textAttachment === undefined) return rest;
      if (
        file.text !== undefined ||
        textAttachment !== nextIndex ||
        nextIndex >= attachments.length
      ) {
        throw new Error('invalid plugin source attachment');
      }
      const source = attachments[nextIndex++];
      if (typeof source === 'string') return { ...rest, text: source };
      if (source instanceof Uint8Array)
        return { ...rest, text: sourceDecoder.decode(source) };
      return { ...rest, sharedSource: source };
    },
  );
  if (nextIndex !== attachments.length) {
    throw new Error('unused plugin source attachment');
  }
  return { ...input, files };
}

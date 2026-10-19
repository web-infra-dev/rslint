/** Resolve binary plugin inputs without reading source or decoding the AST. */
import type { ByteInput } from '../../native/types.js';
import type {
  EslintPluginLintRequest,
  ResolvedEslintPluginLintRequest,
} from './plugin-lint-protocol.js';

function record(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}
export function resolvePluginAttachments(
  input: EslintPluginLintRequest,
  attachments: readonly ByteInput[] = [],
): ResolvedEslintPluginLintRequest {
  if (!record(input) || !Array.isArray(input.files))
    throw new Error('invalid plugin lint request');
  let nextIndex = 0;
  const files = input.files.map(
    (
      file: EslintPluginLintRequest['files'][number],
    ): ResolvedEslintPluginLintRequest['files'][number] => {
      if (
        !record(file) ||
        'text' in file ||
        'textAttachment' in file ||
        'sharedSource' in file ||
        'sourceRange' in file ||
        'sourceIndex' in file
      )
        throw new Error('invalid plugin frame input');
      const { frame, frameAttachment, ...rest } = file;
      if (frameAttachment !== undefined) {
        if (
          frame !== undefined ||
          frameAttachment !== nextIndex ||
          nextIndex >= attachments.length
        )
          throw new Error('invalid plugin frame attachment');
        const bytes = attachments[nextIndex++];
        if (typeof bytes === 'string')
          throw new Error('plugin frame attachment must be binary');
        return { ...rest, frame: bytes };
      }
      if (
        typeof frame !== 'string' ||
        frame.length === 0 ||
        frame.length % 4 !== 0 ||
        !/^[A-Za-z0-9+/]*={0,2}$/.test(frame)
      )
        throw new Error('invalid inline plugin frame');
      const bytes = Buffer.from(frame, 'base64');
      if (bytes.toString('base64') !== frame)
        throw new Error('invalid inline plugin frame');
      return { ...rest, frame: bytes };
    },
  );
  if (nextIndex !== attachments.length)
    throw new Error('unused plugin frame attachment');
  return { ...input, files };
}

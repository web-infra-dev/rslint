/* rslint-disable @typescript-eslint/no-unsafe-type-assertion */
/** Resolve the matching Go producer's frame attachments or inline bytes. */
import type { ByteInput, SharedBytes } from '../../native/types.js';
import type {
  EslintPluginLintRequest,
  ResolvedEslintPluginLintRequest,
} from './plugin-lint-protocol.js';

export function resolvePluginAttachments(
  input: EslintPluginLintRequest,
  attachments: readonly ByteInput[] = [],
): ResolvedEslintPluginLintRequest {
  return {
    ...input,
    files: input.files.map(({ frame, frameAttachment, ...rest }) => ({
      ...rest,
      frame:
        frameAttachment === undefined
          ? Buffer.from(frame, 'base64')
          : (attachments[frameAttachment] as Uint8Array | SharedBytes),
    })),
  };
}

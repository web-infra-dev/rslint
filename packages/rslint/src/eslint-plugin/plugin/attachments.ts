/** Project plugin file indices into inline text or a native byte capability. */
import { TextDecoder } from 'node:util';
import type { ByteInput } from '../../native/binding.js';
import type { EslintPluginLintRequest } from './plugin-lint-protocol.js';

// Match the native parser's UTF-8 rejection while leaving BOM handling to parsing.
const sourceDecoder = new TextDecoder('utf-8', {
  fatal: true,
  ignoreBOM: true,
});

function record(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

export function resolvePluginAttachments(
  input: unknown,
  attachments: readonly ByteInput[] = [],
): EslintPluginLintRequest {
  if (!record(input) || !Array.isArray(input.files)) {
    throw new Error('invalid plugin lint request');
  }
  let nextIndex = 0;
  const files = input.files.map((file: unknown) => {
    if (
      !record(file) ||
      file.sharedSource !== undefined ||
      file.sourceRange !== undefined ||
      file.sourceIndex !== undefined
    ) {
      throw new Error('invalid plugin source file');
    }
    if (file.textAttachment === undefined) return file;
    const { textAttachment, ...rest } = file;
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
  });
  if (nextIndex !== attachments.length) {
    throw new Error('unused plugin source attachment');
  }
  // Other application fields keep the same Go/LSP contract as the task builder.
  // rslint-disable-next-line @typescript-eslint/no-unsafe-type-assertion
  return { ...input, files } as unknown as EslintPluginLintRequest;
}

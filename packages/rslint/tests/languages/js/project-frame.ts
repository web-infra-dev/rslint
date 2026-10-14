/** Test source inputs use the production Go ESTree projection. */
import { execFileSync } from 'node:child_process';
import { decodeFrame } from '../../../src/languages/js/frame.js';

export function projectFrame(
  filePath: string,
  text: string,
  sourceType = 'module',
): Buffer {
  const binary = process.env.RSLINT_ESTREE_PROJECTOR;
  if (!binary)
    throw new Error('test ESTree projector was not built by global setup');
  return execFileSync(binary, {
    input: JSON.stringify({ filePath, text, sourceType }),
    maxBuffer: 64 * 1024 * 1024,
  });
}

export function projectSyntax(
  filePath: string,
  text: string,
  sourceType = 'module',
) {
  return decodeFrame(projectFrame(filePath, text, sourceType));
}

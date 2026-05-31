// Reference: https://github.com/biomejs/biome-vscode/blob/8fa2ca19e612575479c840bd58f6d31e4e503b13/src/utils.ts

import { arch } from 'node:os';
import { FileType, Uri, workspace } from 'vscode';

/**
 * Checks whether a file exists
 *
 * This function checks whether a file exists at the given URI using VS Code's
 * FileSystem API.
 *
 * @param uri URI of the file to check
 * @returns Whether the file exists
 */
export const fileExists = async (uri: Uri): Promise<boolean> => {
  try {
    const stat = await workspace.fs.stat(uri);
    return (stat.type & FileType.File) > 0;
  } catch {
    return false;
  }
};

/**
 * Returns the ordered list of platform-package requests to try-resolve when
 * locating the bundled Go binary, mirroring `packages/rslint/bin/rslint.cjs`.
 *
 * The Go binary lives in the `@rslint/native-{tuple}` subpackage, reached via
 * its `./bin` export. npm installs only the subpackage matching the host
 * os/cpu/libc, so on linux we try gnu then musl and use whichever got
 * installed — no libc sniffing (Go binaries are static, the gnu/musl
 * distinction doesn't matter to them). Callers should resolve each candidate
 * in order and use the first that succeeds.
 */
export const getPlatformBinRequests = (): string[] => {
  const cpu = arch();
  const tuples =
    process.platform === 'linux'
      ? [`linux-${cpu}-gnu`, `linux-${cpu}-musl`]
      : process.platform === 'win32'
        ? [`win32-${cpu}-msvc`]
        : [`${process.platform}-${cpu}`];
  return tuples.map((tuple) => `@rslint/native-${tuple}/bin`);
};

export type RslintBinPath = 'local' | 'built-in' | 'custom';

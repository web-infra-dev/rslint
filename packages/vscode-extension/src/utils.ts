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
  } catch (_err) {
    return false;
  }
};

export const PLATFORM_KEY = `${process.platform}-${arch()}`;
export const PLATFORM_BIN_REQUEST = `@rslint/${PLATFORM_KEY}/rslint`;

export type RslintBinPath = 'local' | 'built-in' | 'custom';

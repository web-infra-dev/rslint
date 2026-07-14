import { randomUUID } from 'node:crypto';
import fs from 'node:fs';
import path from 'node:path';
import * as vscode from 'vscode';

export function temporaryFilePath(
  directory: string,
  prefix: string,
  extension = '.ts',
): string {
  return path.join(directory, `${prefix}${randomUUID()}${extension}`);
}

async function focusTextDocument(document: vscode.TextDocument): Promise<void> {
  await vscode.window.showTextDocument(document, {
    preview: false,
    preserveFocus: false,
  });
  if (
    vscode.window.activeTextEditor?.document.uri.toString() !==
    document.uri.toString()
  ) {
    throw new Error(`Could not focus document: ${document.uri}`);
  }
}

export async function revertTextDocument(
  document: vscode.TextDocument | undefined,
): Promise<void> {
  if (!document || document.isClosed) return;
  if (!document.isDirty) return;
  await focusTextDocument(document);
  await vscode.commands.executeCommand('workbench.action.files.revert');
  if (document.isDirty) {
    throw new Error(`Could not revert dirty document: ${document.uri}`);
  }
}

/** Close the exact editor tab. VS Code may retain its TextDocument model. */
export async function closeTextEditor(
  document: vscode.TextDocument | undefined,
): Promise<void> {
  if (!document || document.isClosed) return;
  await focusTextDocument(document);
  await revertTextDocument(document);
  await vscode.commands.executeCommand('workbench.action.closeActiveEditor');
  if (
    vscode.window.activeTextEditor?.document.uri.toString() ===
    document.uri.toString()
  ) {
    throw new Error(`Could not close editor tab: ${document.uri}`);
  }
}

export function deleteTemporaryFile(filePath: string): void {
  if (fs.existsSync(filePath)) fs.unlinkSync(filePath);
}

/** Close/revert the exact tab, delete its unique file, and verify deletion. */
export async function closeAndDeleteTemporaryDocument(
  document: vscode.TextDocument | undefined,
  filePath: string,
): Promise<void> {
  const errors: unknown[] = [];
  try {
    await closeTextEditor(document);
  } catch (error) {
    errors.push(error);
  }
  try {
    deleteTemporaryFile(filePath);
  } catch (error) {
    errors.push(error);
  }
  if (fs.existsSync(filePath)) {
    errors.push(
      new Error(`Temporary file still exists after delete: ${filePath}`),
    );
  }

  if (errors.length === 1) throw errors[0];
  if (errors.length > 1) {
    throw new AggregateError(errors, 'Temporary document cleanup failed');
  }
}

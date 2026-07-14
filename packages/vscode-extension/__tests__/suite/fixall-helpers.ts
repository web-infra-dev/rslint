import * as assert from 'assert';
import * as vscode from 'vscode';
import path from 'node:path';
import fs from 'node:fs';
import {
  waitForRslintDiagnostics,
  waitForRslintDiagnosticsCount,
  waitForRslintDiagnosticsToChange,
} from '../utils/diagnostics';
import { withCodeActionsOnSave } from '../utils/configuration';
import {
  closeAndDeleteTemporaryDocument,
  temporaryFilePath,
} from '../utils/documents';
import { waitForCodeActionRegistryQuiescence } from '../utils/codeActionRegistry';

export { saveDocumentOnce } from '../utils/codeActionRegistry';

export const waitForDiagnostics = waitForRslintDiagnostics;
export const waitForDiagnosticsCount = waitForRslintDiagnosticsCount;
export const waitForDiagnosticsToChange = waitForRslintDiagnosticsToChange;

export function getFixturesDir(): string {
  const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
  if (!workspaceFolder)
    throw new Error('VS Code test workspace is unavailable');
  return workspaceFolder.uri.fsPath;
}

export async function openFixture(
  filename: string,
): Promise<vscode.TextDocument> {
  return vscode.workspace.openTextDocument(
    path.resolve(getFixturesDir(), 'src/', filename),
  );
}

export function findFixAllAction(
  codeActions: vscode.CodeAction[] | undefined,
): vscode.CodeAction | undefined {
  return codeActions?.find(
    (action) =>
      action.kind?.value === 'source.fixAll.rslint' ||
      action.kind?.value === 'source.fixAll',
  );
}

export async function executeCodeActionProvider(
  uri: vscode.Uri,
  range: vscode.Range,
  kind?: string,
): Promise<vscode.CodeAction[]> {
  await waitForCodeActionRegistryQuiescence();
  const args: unknown[] = ['vscode.executeCodeActionProvider', uri, range];
  if (kind !== undefined) args.push(kind);
  const result = await vscode.commands.executeCommand<vscode.CodeAction[]>(
    ...(args as [string, vscode.Uri, vscode.Range, ...unknown[]]),
  );
  return result ?? [];
}

export async function requestFixAll(
  doc: vscode.TextDocument,
  kind: vscode.CodeActionKind = vscode.CodeActionKind.SourceFixAll.append(
    'rslint',
  ),
): Promise<vscode.CodeAction[]> {
  return executeCodeActionProvider(
    doc.uri,
    new vscode.Range(0, 0, doc.lineCount, 0),
    kind.value,
  );
}

export async function withTmpFile(
  content: string,
  testFn: (
    doc: vscode.TextDocument,
    editor: vscode.TextEditor,
  ) => Promise<void>,
): Promise<void> {
  const tmpFile = temporaryFilePath(
    path.join(getFixturesDir(), 'src'),
    '_fixall_tmp_',
  );
  fs.writeFileSync(tmpFile, content, 'utf-8');
  let doc: vscode.TextDocument | undefined;
  let testError: unknown;
  try {
    doc = await vscode.workspace.openTextDocument(tmpFile);
    const editor = await vscode.window.showTextDocument(doc);
    await testFn(doc, editor);
  } catch (error) {
    testError = error;
  }
  await finishTemporaryDocument(testError, doc, tmpFile);
}

/**
 * Wait until `predicate(content)` becomes true. Subscribes to
 * `vscode.workspace.onDidChangeTextDocument` and returns the moment a
 * matching content arrives, instead of polling on a fixed interval.
 *
 * Use this in preference to a `while (...) await sleep(500)` loop when the
 * test is waiting for a server-driven content change (e.g. on-save fixAll
 * applying an edit): the event-driven path resolves with sub-millisecond
 * latency once the change lands, so the only wall-clock cost is the
 * server's actual response time.
 *
 * Rejects with a descriptive error (including the last seen content) when
 * `timeoutMs` elapses without the predicate being satisfied.
 */
export async function waitForContentChange(
  doc: vscode.TextDocument,
  predicate: (content: string) => boolean,
  timeoutMs: number,
): Promise<string> {
  const initial = doc.getText();
  if (predicate(initial)) return initial;
  return new Promise<string>((resolve, reject) => {
    const docUriString = doc.uri.toString();
    let timer: ReturnType<typeof setTimeout> | undefined;
    const disposable = vscode.workspace.onDidChangeTextDocument((e) => {
      if (e.document.uri.toString() !== docUriString) return;
      const current = doc.getText();
      if (predicate(current)) {
        if (timer) clearTimeout(timer);
        disposable.dispose();
        resolve(current);
      }
    });
    timer = setTimeout(() => {
      disposable.dispose();
      reject(
        new Error(
          `waitForContentChange: predicate not satisfied within ${timeoutMs}ms. Last content:\n${doc.getText()}`,
        ),
      );
    }, timeoutMs);
  });
}

export async function replaceAll(
  editor: vscode.TextEditor,
  newContent: string,
): Promise<void> {
  const doc = editor.document;
  const fullRange = new vscode.Range(
    doc.positionAt(0),
    doc.positionAt(doc.getText().length),
  );
  const ok = await editor.edit((b) => b.replace(fullRange, newContent));
  assert.ok(ok, 'editor.edit should succeed');
}

export async function withOnSaveFixAll(
  testFn: (
    doc: vscode.TextDocument,
    editor: vscode.TextEditor,
  ) => Promise<void>,
  codeActionsOnSave: Record<string, 'always' | 'explicit' | 'never'> = {
    'source.fixAll.rslint': 'explicit',
  },
): Promise<void> {
  const tmpFile = temporaryFilePath(
    path.join(getFixturesDir(), 'src'),
    '_fixall_test_',
  );
  fs.writeFileSync(tmpFile, '// placeholder\n', 'utf-8');

  let doc: vscode.TextDocument | undefined;
  let testError: unknown;
  try {
    const openedDocument = await vscode.workspace.openTextDocument(tmpFile);
    doc = openedDocument;
    const editor = await vscode.window.showTextDocument(openedDocument);
    await withCodeActionsOnSave(openedDocument, codeActionsOnSave, async () => {
      await testFn(openedDocument, editor);
    });
  } catch (error) {
    testError = error;
  }
  await finishTemporaryDocument(testError, doc, tmpFile);
}

async function finishTemporaryDocument(
  testError: unknown,
  document: vscode.TextDocument | undefined,
  filePath: string,
): Promise<void> {
  const errors: unknown[] = [];
  if (testError) errors.push(testError);

  try {
    await closeAndDeleteTemporaryDocument(document, filePath);
  } catch (error) {
    errors.push(error);
  }

  if (errors.length === 1) throw errors[0];
  if (errors.length > 1) {
    throw new AggregateError(errors, 'Test and temporary-file cleanup failed');
  }
}

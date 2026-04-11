import * as assert from 'assert';
import * as vscode from 'vscode';
import path from 'node:path';
import fs from 'node:fs';

export function getFixturesDir(): string {
  return path.resolve(require.resolve('@rslint/core'), '../..', 'fixtures');
}

export async function openFixture(
  filename: string,
): Promise<vscode.TextDocument> {
  return vscode.workspace.openTextDocument(
    path.resolve(getFixturesDir(), 'src/', filename),
  );
}

export async function waitForDiagnostics(
  doc: vscode.TextDocument,
): Promise<vscode.Diagnostic[]> {
  for (let i = 0; i < 10; i++) {
    const diagnostics = vscode.languages.getDiagnostics(doc.uri);
    if (diagnostics.length > 0) {
      return diagnostics;
    }
    await new Promise((resolve) => {
      const disposable = vscode.languages.onDidChangeDiagnostics((e) => {
        for (const uri of e.uris) {
          if (uri.toString() === doc.uri.toString()) {
            disposable.dispose();
            resolve(void 0);
            return;
          }
        }
      });
      setTimeout(() => {
        disposable.dispose();
        resolve(void 0);
      }, 1000);
    });
  }
  return vscode.languages.getDiagnostics(doc.uri);
}

export async function waitForDiagnosticsToChange(
  doc: vscode.TextDocument,
  previousCount: number,
  timeoutMs = 15000,
): Promise<vscode.Diagnostic[]> {
  const startTime = Date.now();
  while (Date.now() - startTime < timeoutMs) {
    const current = vscode.languages.getDiagnostics(doc.uri);
    if (current.length !== previousCount) {
      return current;
    }
    await new Promise((resolve) => {
      const disposable = vscode.languages.onDidChangeDiagnostics((e) => {
        for (const uri of e.uris) {
          if (uri.toString() === doc.uri.toString()) {
            disposable.dispose();
            resolve(void 0);
            return;
          }
        }
      });
      setTimeout(() => {
        disposable.dispose();
        resolve(void 0);
      }, 500);
    });
  }
  return vscode.languages.getDiagnostics(doc.uri);
}

export async function waitForDiagnosticsCount(
  doc: vscode.TextDocument,
  expectedCount: number,
  timeoutMs = 15000,
): Promise<vscode.Diagnostic[]> {
  const startTime = Date.now();
  while (Date.now() - startTime < timeoutMs) {
    const current = vscode.languages.getDiagnostics(doc.uri);
    if (current.length === expectedCount) {
      return current;
    }
    await new Promise((resolve) => {
      const disposable = vscode.languages.onDidChangeDiagnostics((e) => {
        for (const uri of e.uris) {
          if (uri.toString() === doc.uri.toString()) {
            disposable.dispose();
            resolve(void 0);
            return;
          }
        }
      });
      setTimeout(() => {
        disposable.dispose();
        resolve(void 0);
      }, 500);
    });
  }
  return vscode.languages.getDiagnostics(doc.uri);
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

export async function requestFixAll(
  doc: vscode.TextDocument,
  kind: vscode.CodeActionKind = vscode.CodeActionKind.SourceFixAll.append(
    'rslint',
  ),
): Promise<vscode.CodeAction[]> {
  return (
    (await vscode.commands.executeCommand<vscode.CodeAction[]>(
      'vscode.executeCodeActionProvider',
      doc.uri,
      new vscode.Range(0, 0, doc.lineCount, 0),
      kind.value,
    )) ?? []
  );
}

export async function withTmpFile(
  content: string,
  testFn: (
    doc: vscode.TextDocument,
    editor: vscode.TextEditor,
  ) => Promise<void>,
): Promise<void> {
  const tmpFile = path.join(
    getFixturesDir(),
    'src',
    `_fixall_tmp_${Date.now()}_${Math.random().toString(36).slice(2, 8)}.ts`,
  );
  fs.writeFileSync(tmpFile, content, 'utf-8');
  try {
    const doc = await vscode.workspace.openTextDocument(tmpFile);
    const editor = await vscode.window.showTextDocument(doc);
    await testFn(doc, editor);
  } finally {
    // Close the editor tab so VSCode sends a synchronous didClose to the LSP,
    // which cleans up the session overlay. Without this, the file deletion
    // triggers an async didClose via the file watcher, which can race with
    // the next test's LSP requests (all blocking methods are serialized).
    await vscode.commands.executeCommand('workbench.action.closeActiveEditor');
    if (fs.existsSync(tmpFile)) {
      fs.unlinkSync(tmpFile);
    }
  }
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
): Promise<void> {
  const fixturesDir = getFixturesDir();
  const tmpFile = path.join(
    fixturesDir,
    'src',
    `_fixall_test_${Date.now()}.ts`,
  );
  fs.writeFileSync(tmpFile, '// placeholder\n', 'utf-8');

  try {
    const config = vscode.workspace.getConfiguration('editor');
    const previousValue = config.get('codeActionsOnSave');
    await config.update(
      'codeActionsOnSave',
      { 'source.fixAll.rslint': 'explicit' },
      vscode.ConfigurationTarget.Workspace,
    );

    try {
      const doc = await vscode.workspace.openTextDocument(tmpFile);
      const editor = await vscode.window.showTextDocument(doc);
      await testFn(doc, editor);
    } finally {
      await config.update(
        'codeActionsOnSave',
        previousValue,
        vscode.ConfigurationTarget.Workspace,
      );
    }
  } finally {
    // Close the editor tab so VSCode sends a synchronous didClose to the LSP,
    // which cleans up the session overlay. Without this, the file deletion
    // triggers an async didClose via the file watcher, which can race with
    // the next test's LSP requests (all blocking methods are serialized).
    await vscode.commands.executeCommand('workbench.action.closeActiveEditor');
    if (fs.existsSync(tmpFile)) {
      fs.unlinkSync(tmpFile);
    }
  }
}

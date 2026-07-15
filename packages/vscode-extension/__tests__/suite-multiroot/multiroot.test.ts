import * as assert from 'node:assert';
import path from 'node:path';
import * as vscode from 'vscode';

function workspaceFolder(name: string): vscode.WorkspaceFolder {
  const folder = vscode.workspace.workspaceFolders?.find(
    (candidate) => candidate.name === name,
  );
  assert.ok(folder, `workspace folder ${name} is unavailable`);
  return folder;
}

async function openWorkspaceFile(
  folder: vscode.WorkspaceFolder,
  relativePath: string,
): Promise<vscode.TextDocument> {
  return vscode.workspace.openTextDocument(
    path.join(folder.uri.fsPath, relativePath),
  );
}

function rslintDiagnostics(document: vscode.TextDocument): vscode.Diagnostic[] {
  return vscode.languages
    .getDiagnostics(document.uri)
    .filter((diagnostic) => diagnostic.source === 'rslint');
}

async function waitForSingleRslintDiagnostic(
  document: vscode.TextDocument,
): Promise<vscode.Diagnostic[]> {
  const deadline = Date.now() + 30_000;
  while (Date.now() < deadline) {
    const diagnostics = rslintDiagnostics(document);
    if (
      diagnostics.length === 1 &&
      diagnostics[0].message.includes('no-explicit-any')
    ) {
      // Do not accept a transient single result while a duplicate owner is
      // still publishing its first diagnostics.
      await new Promise((resolve) => setTimeout(resolve, 300));
      const stable = rslintDiagnostics(document);
      if (stable.length === 1) return stable;
    }
    await new Promise<void>((resolve) => {
      const listener = vscode.languages.onDidChangeDiagnostics((event) => {
        if (
          event.uris.some((uri) => uri.toString() === document.uri.toString())
        ) {
          listener.dispose();
          resolve();
        }
      });
      setTimeout(() => {
        listener.dispose();
        resolve();
      }, 500);
    });
  }
  return rslintDiagnostics(document);
}

suite('VS Code multi-root ownership', function () {
  this.timeout(60_000);

  test('keeps same-name roots independent', async function () {
    const appFolders = (vscode.workspace.workspaceFolders ?? []).filter(
      (folder) => folder.name === 'app',
    );
    if (appFolders.length === 0) this.skip();
    assert.strictEqual(appFolders.length, 2);
    assert.notStrictEqual(
      appFolders[0].uri.toString(),
      appFolders[1].uri.toString(),
    );

    for (const folder of appFolders) {
      const document = await openWorkspaceFile(folder, 'src/index.ts');
      const diagnostics = await waitForSingleRslintDiagnostic(document);
      assert.strictEqual(
        diagnostics.length,
        1,
        `${folder.uri} should have exactly one Rslint owner`,
      );
    }
  });

  test('routes an initial parent-child overlap to only the child', async function () {
    const folders = vscode.workspace.workspaceFolders ?? [];
    const parent = folders.find((folder) => folder.name === 'parent');
    const nested = folders.find((folder) => folder.name === 'nested');
    if (!parent || !nested) this.skip();

    const document = await openWorkspaceFile(nested, 'src/index.ts');
    const diagnostics = await waitForSingleRslintDiagnostic(document);
    assert.strictEqual(diagnostics.length, 1);
  });

  test('hands a document parent → child → parent without reopening it', async function () {
    const folders = vscode.workspace.workspaceFolders ?? [];
    if (folders.some((folder) => folder.name === 'nested')) this.skip();
    // Keep the first root and a sentinel throughout the test. VS Code may
    // restart the extension host for first-root or single↔multi transitions.
    assert.ok(folders.length >= 2);
    const parent = workspaceFolder('parent');
    workspaceFolder('sentinel');
    const nestedUri = vscode.Uri.joinPath(parent.uri, 'nested');
    const document = await vscode.workspace.openTextDocument(
      vscode.Uri.joinPath(nestedUri, 'src/index.ts'),
    );
    assert.strictEqual(
      (await waitForSingleRslintDiagnostic(document)).length,
      1,
    );

    await expectDiagnosticHandoff(document, async () => {
      const added = vscode.workspace.updateWorkspaceFolders(
        vscode.workspace.workspaceFolders?.length ?? 0,
        0,
        { uri: nestedUri, name: 'nested' },
      );
      assert.strictEqual(added, true);
      await waitForWorkspaceFolder(nestedUri, true);
    });
    assert.strictEqual(
      (await waitForSingleRslintDiagnostic(document)).length,
      1,
    );

    const nestedIndex = vscode.workspace.workspaceFolders?.findIndex(
      (folder) => folder.uri.toString() === nestedUri.toString(),
    );
    assert.notStrictEqual(nestedIndex, undefined);
    assert.ok(nestedIndex !== undefined && nestedIndex >= 0);
    await expectDiagnosticHandoff(document, async () => {
      const removed = vscode.workspace.updateWorkspaceFolders(nestedIndex, 1);
      assert.strictEqual(removed, true);
      await waitForWorkspaceFolder(nestedUri, false);
    });
    assert.strictEqual(
      (await waitForSingleRslintDiagnostic(document)).length,
      1,
    );
  });
});

async function expectDiagnosticHandoff(
  document: vscode.TextDocument,
  changeTopology: () => Promise<void>,
): Promise<void> {
  let events = 0;
  let maximumDiagnosticCount = rslintDiagnostics(document).length;
  const listener = vscode.languages.onDidChangeDiagnostics((event) => {
    if (event.uris.some((uri) => uri.toString() === document.uri.toString())) {
      events++;
      maximumDiagnosticCount = Math.max(
        maximumDiagnosticCount,
        rslintDiagnostics(document).length,
      );
    }
  });
  try {
    await changeTopology();
    const deadline = Date.now() + 30_000;
    // VS Code may coalesce the old collection's delete with the new
    // collection's publication into one aggregate diagnostic event.
    while (events < 1 && Date.now() < deadline) {
      await new Promise((resolve) => setTimeout(resolve, 50));
    }
    assert.ok(
      events >= 1,
      'expected diagnostics to change during ownership handoff',
    );
    assert.strictEqual(
      (await waitForSingleRslintDiagnostic(document)).length,
      1,
    );
    assert.ok(
      maximumDiagnosticCount <= 1,
      `ownership handoff published ${maximumDiagnosticCount} diagnostics`,
    );
  } finally {
    listener.dispose();
  }
}

async function waitForWorkspaceFolder(
  uri: vscode.Uri,
  present: boolean,
): Promise<void> {
  const deadline = Date.now() + 10_000;
  while (Date.now() < deadline) {
    const found = vscode.workspace.workspaceFolders?.some(
      (folder) => folder.uri.toString() === uri.toString(),
    );
    if (found === present) return;
    await new Promise((resolve) => setTimeout(resolve, 50));
  }
  assert.fail(
    `workspace folder ${uri} did not become ${present ? 'present' : 'absent'}`,
  );
}

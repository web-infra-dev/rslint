import * as assert from 'assert';
import * as vscode from 'vscode';
import path from 'node:path';
import { waitForRslintDiagnostics as waitForDiagnostics } from '../utils/diagnostics';
import { closeTextEditor } from '../utils/documents';

// Tests that type-aware rules (e.g. require-await) only run on files covered
// by parserOptions.project, matching CLI behavior.
//
// Fixture: rslint.config.js has project:
// ['./packages/core/tsconfig.lint.json']. The default tsconfig.json deliberately
// excludes index.ts, so tsgo's main Session cannot own the declared lint
// project. This exercises rslint's standalone Program path.
// - packages/core/src/index.ts: IN tsconfig.lint.json, type-aware rules fire
// - packages/cli/src/preview.ts: NOT in tsconfig.lint.json, type-aware rules do not fire
suite('rslint type-aware rule scope', function () {
  this.timeout(120000);

  function getWorkspaceRoot(): string {
    return vscode.workspace.workspaceFolders![0].uri.fsPath;
  }

  test('file IN parserOptions.project tsconfig should get type-aware rules', async () => {
    const filePath = path.join(
      getWorkspaceRoot(),
      'packages/core/src/index.ts',
    );
    const doc = await vscode.workspace.openTextDocument(filePath);
    await vscode.window.showTextDocument(doc);

    const diagnostics = await waitForDiagnostics(doc, (diags) =>
      diags.some((d) => d.message.includes('require-await')),
    );

    // require-await should fire (type-aware, file IS in tsconfig)
    assert.ok(
      diagnostics.some((d) => d.message.includes('require-await')),
      `Expected require-await for file in tsconfig. Got: ${diagnostics.map((d) => d.message).join(', ')}`,
    );

    // no-console should also fire (non-type-aware, always runs)
    assert.ok(
      diagnostics.some((d) => d.message.includes('no-console')),
      'Expected no-console for file in tsconfig',
    );
  });

  test('file NOT in parserOptions.project tsconfig should NOT get type-aware rules', async () => {
    const filePath = path.join(
      getWorkspaceRoot(),
      'packages/cli/src/preview.ts',
    );
    const doc = await vscode.workspace.openTextDocument(filePath);
    await vscode.window.showTextDocument(doc);

    // Wait for no-console (non-type-aware) to appear — proves rslint IS linting the file
    const diagnostics = await waitForDiagnostics(doc, (diags) =>
      diags.some((d) => d.message.includes('no-console')),
    );

    // no-console SHOULD fire (non-type-aware, always runs)
    assert.ok(
      diagnostics.some((d) => d.message.includes('no-console')),
      `Expected no-console for file outside tsconfig. Got: ${diagnostics.map((d) => d.message).join(', ')}`,
    );

    // require-await should NOT fire (type-aware, file is NOT in configured tsconfig)
    assert.ok(
      !diagnostics.some((d) => d.message.includes('require-await')),
      'require-await should NOT fire for file outside parserOptions.project tsconfig',
    );
  });

  test('standalone project stays correct across edit and reopen', async () => {
    const filePath = path.join(
      getWorkspaceRoot(),
      'packages/core/src/index.ts',
    );
    const doc = await vscode.workspace.openTextDocument(filePath);
    const editor = await vscode.window.showTextDocument(doc);
    const initial = await waitForDiagnostics(doc, (diags) =>
      diags.some((d) => d.message.includes('require-await')),
    );
    assert.ok(
      initial.some((d) => d.message.includes('require-await')),
      'Expected require-await before editing the standalone project source',
    );

    const original = doc.getText();
    const changed = original.replace(
      "  console.log('hello');",
      "  await Promise.resolve();\n  console.log('hello');",
    );
    assert.notStrictEqual(
      changed,
      original,
      'Fixture edit did not match source',
    );
    const fullRange = new vscode.Range(
      doc.positionAt(0),
      doc.positionAt(original.length),
    );
    await editor.edit((builder) => builder.replace(fullRange, changed));
    const updated = await waitForDiagnostics(
      doc,
      (diags) =>
        diags.some((d) => d.message.includes('no-console')) &&
        !diags.some((d) => d.message.includes('require-await')),
    );
    assert.ok(
      !updated.some((d) => d.message.includes('require-await')),
      'Incremental standalone Program retained a stale require-await diagnostic',
    );

    await closeTextEditor(doc);
    const reopened = await vscode.workspace.openTextDocument(filePath);
    await vscode.window.showTextDocument(reopened);
    const reopenedDiagnostics = await waitForDiagnostics(reopened, (diags) =>
      diags.some((d) => d.message.includes('require-await')),
    );
    assert.ok(
      reopenedDiagnostics.some((d) => d.message.includes('require-await')),
      'Reopened standalone project source did not restore disk diagnostics',
    );
  });

  test('standalone project refreshes after an external dependency change', async () => {
    const sourcePath = path.join(
      getWorkspaceRoot(),
      'packages/core/src/index.ts',
    );
    const dependencyUri = vscode.Uri.file(
      path.join(getWorkspaceRoot(), 'packages/core/src/dependency.ts'),
    );
    const source = await vscode.workspace.openTextDocument(sourcePath);
    await vscode.window.showTextDocument(source);
    const initial = await waitForDiagnostics(source, (diags) =>
      diags.some((d) => d.message.includes('no-unsafe-member-access')),
    );
    assert.ok(
      initial.some((d) => d.message.includes('no-unsafe-member-access')),
      'Expected the dependency any type to produce no-unsafe-member-access',
    );

    const originalDependency =
      await vscode.workspace.fs.readFile(dependencyUri);
    try {
      await vscode.workspace.fs.writeFile(
        dependencyUri,
        Buffer.from('export const dependency = { value: 1 };\n'),
      );
      const updated = await waitForDiagnostics(
        source,
        (diags) =>
          diags.some((d) => d.message.includes('no-console')) &&
          !diags.some((d) => d.message.includes('no-unsafe-member-access')),
      );
      assert.ok(
        !updated.some((d) => d.message.includes('no-unsafe-member-access')),
        'External dependency change left stale standalone Program diagnostics',
      );
    } finally {
      await vscode.workspace.fs.writeFile(dependencyUri, originalDependency);
    }

    const restored = await waitForDiagnostics(source, (diags) =>
      diags.some((d) => d.message.includes('no-unsafe-member-access')),
    );
    assert.ok(
      restored.some((d) => d.message.includes('no-unsafe-member-access')),
      'Restored dependency did not restore standalone Program diagnostics',
    );
  });
});

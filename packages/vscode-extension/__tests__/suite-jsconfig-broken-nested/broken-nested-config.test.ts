import * as assert from 'assert';
import fs from 'node:fs';
import path from 'node:path';
import * as vscode from 'vscode';
import { waitForRslintDiagnostics as waitForDiagnostics } from '../utils/diagnostics';
import { closeTextEditor } from '../utils/documents';

suite('rslint broken nested JS config', function () {
  this.timeout(120_000);

  function getWorkspaceRoot(): string {
    const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
    assert.ok(workspaceFolder, 'Expected an open workspace folder');
    return workspaceFolder.uri.fsPath;
  }

  async function waitForFile(filePath: string, timeoutMs = 10_000) {
    const deadline = Date.now() + timeoutMs;
    while (Date.now() < deadline) {
      if (fs.existsSync(filePath)) return;
      await new Promise((resolve) => setTimeout(resolve, 25));
    }
    throw new Error(`Timed out waiting for file: ${filePath}`);
  }

  test('a newly discovered broken nested config keeps valid ancestor config active', async () => {
    const root = getWorkspaceRoot();
    const rootConfigPath = path.join(root, 'rslint.config.js');
    const nestedDir = path.join(root, 'broken-nested-config');
    const nestedFilePath = path.join(nestedDir, 'index.ts');
    const nestedConfigPath = path.join(nestedDir, 'rslint.config.js');
    const attemptedLoadPath = path.join(nestedDir, 'config-load-attempted');
    const postFailureFilePath = path.join(nestedDir, 'post-failure.ts');
    const rootDoc = await vscode.workspace.openTextDocument(
      path.join(root, 'src', 'index.ts'),
    );

    const rootConfigWithMarker = `export default [{
  files: ['**/*.ts'],
  languageOptions: {
    parserOptions: { projectService: false, project: ['./tsconfig.json'] },
  },
  rules: {
    '@typescript-eslint/no-unsafe-member-access': 'warn',
    'no-debugger': 'error',
  },
  plugins: ['@typescript-eslint'],
}];
`;

    // Establish a positive initial-config publication before mutating the
    // workspace, so no startup snapshot can satisfy the later assertions.
    await vscode.window.showTextDocument(rootDoc);
    await waitForDiagnostics(rootDoc, (diags) =>
      diags.some(
        (diagnostic) =>
          diagnostic.message.includes('no-unsafe-member-access') &&
          diagnostic.severity === vscode.DiagnosticSeverity.Error,
      ),
    );

    // This suite owns a disposable workspace and Extension Host. Leave watched
    // files in place so runIsolatedSuite removes the workspace only after VS
    // Code exits, instead of racing live Windows filesystem watchers here.
    fs.mkdirSync(nestedDir);
    fs.writeFileSync(nestedFilePath, 'debugger;\n', 'utf8');
    fs.writeFileSync(rootConfigPath, rootConfigWithMarker, 'utf8');

    await waitForDiagnostics(rootDoc, (diags) =>
      diags.some(
        (diagnostic) =>
          diagnostic.message.includes('no-unsafe-member-access') &&
          diagnostic.severity === vscode.DiagnosticSeverity.Warning,
      ),
    );
    const nestedDoc = await vscode.workspace.openTextDocument(nestedFilePath);
    await vscode.window.showTextDocument(nestedDoc);
    await waitForDiagnostics(nestedDoc, (diags) =>
      diags.some(
        (diagnostic) =>
          diagnostic.message.includes('no-debugger') &&
          diagnostic.severity === vscode.DiagnosticSeverity.Error,
      ),
    );
    await closeTextEditor(nestedDoc);

    fs.writeFileSync(
      nestedConfigPath,
      `import fs from 'node:fs';
fs.writeFileSync(${JSON.stringify(attemptedLoadPath)}, 'attempted', 'utf8');
throw new Error('intentional broken nested config');
export default [];
`,
      'utf8',
    );
    await waitForFile(attemptedLoadPath);
    assert.strictEqual(
      fs.readFileSync(attemptedLoadPath, 'utf8'),
      'attempted',
      'The broken nested config must be evaluated before fallback is asserted',
    );

    // This URI does not exist until after the broken module has executed, so
    // its diagnostics cannot be a stale snapshot. Config discovery and didOpen
    // share the serialized server dispatch loop, so this later didOpen cannot
    // run until the transaction that observed the broken child has finished.
    fs.writeFileSync(postFailureFilePath, 'debugger;\n', 'utf8');
    const postFailureDoc =
      await vscode.workspace.openTextDocument(postFailureFilePath);
    await vscode.window.showTextDocument(postFailureDoc);
    const postFailureDiagnostics = await waitForDiagnostics(
      postFailureDoc,
      (diags) =>
        diags.some(
          (diagnostic) =>
            diagnostic.message.includes('no-debugger') &&
            diagnostic.severity === vscode.DiagnosticSeverity.Error,
        ),
    );
    assert.deepStrictEqual(
      postFailureDiagnostics
        .filter((diagnostic) => diagnostic.message.includes('no-debugger'))
        .map((diagnostic) => diagnostic.severity),
      [vscode.DiagnosticSeverity.Error],
      'The valid ancestor must lint a file opened after the broken child was evaluated',
    );
  });
});

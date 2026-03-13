import * as assert from 'assert';
import * as vscode from 'vscode';
import path from 'node:path';
import fs from 'node:fs';

suite('rslint JS config support', function () {
  this.timeout(60000);

  function getWorkspaceRoot(): string {
    return vscode.workspace.workspaceFolders![0].uri.fsPath;
  }

  async function waitForDiagnostics(
    doc: vscode.TextDocument,
    predicate?: (diags: vscode.Diagnostic[]) => boolean,
  ): Promise<vscode.Diagnostic[]> {
    for (let i = 0; i < 20; i++) {
      const diagnostics = vscode.languages.getDiagnostics(doc.uri);
      if (predicate ? predicate(diagnostics) : diagnostics.length > 0) {
        return diagnostics;
      }

      await new Promise(resolve => {
        const disposable = vscode.languages.onDidChangeDiagnostics(e => {
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
        }, 1500);
      });
    }

    return vscode.languages.getDiagnostics(doc.uri);
  }

  async function openFixture(filename: string): Promise<vscode.TextDocument> {
    const filePath = path.join(getWorkspaceRoot(), 'src', filename);
    return vscode.workspace.openTextDocument(filePath);
  }

  test('JS config should produce diagnostics', async () => {
    const doc = await openFixture('index.ts');
    await vscode.window.showTextDocument(doc);

    const diagnostics = await waitForDiagnostics(doc);
    assert.ok(
      diagnostics.length > 0,
      `Expected diagnostics but got ${diagnostics.length}`,
    );
    assert.ok(
      diagnostics.some(d => d.message.includes('no-unsafe-member-access')),
      'Expected no-unsafe-member-access diagnostic from JS config',
    );
  });

  test('JS config should take priority over JSON config', async () => {
    const doc = await openFixture('index.ts');
    await vscode.window.showTextDocument(doc);

    // Wait specifically for JS config diagnostics (no-unsafe-member-access).
    // JSON config may load first with no-explicit-any, but JS config should
    // override via rslint/configUpdate notification.
    const diagnostics = await waitForDiagnostics(doc, diags =>
      diags.some(d => d.message.includes('no-unsafe-member-access')),
    );

    assert.ok(
      diagnostics.some(d => d.message.includes('no-unsafe-member-access')),
      'Expected no-unsafe-member-access from JS config',
    );
    assert.ok(
      !diagnostics.some(d => d.message.includes('no-explicit-any')),
      'Should NOT see no-explicit-any because JS config takes priority over JSON',
    );
  });

  test('config hot reload should update diagnostics', async () => {
    const doc = await openFixture('index.ts');
    const editor = await vscode.window.showTextDocument(doc);

    // 1. Verify initial diagnostics have no-unsafe-member-access
    const initialDiags = await waitForDiagnostics(doc, diags =>
      diags.some(d => d.message.includes('no-unsafe-member-access')),
    );
    assert.ok(
      initialDiags.some(d => d.message.includes('no-unsafe-member-access')),
      'Initial diagnostics should include no-unsafe-member-access',
    );

    // 2. Modify rslint.config.js to use a different rule
    const configPath = path.join(getWorkspaceRoot(), 'rslint.config.js');
    const originalConfig = fs.readFileSync(configPath, 'utf8');

    const newConfig = `export default [
  {
    files: ['**/*.ts'],
    languageOptions: {
      parserOptions: {
        projectService: false,
        project: ['./tsconfig.json'],
      },
    },
    rules: {
      '@typescript-eslint/no-explicit-any': 'error',
      '@typescript-eslint/no-unsafe-member-access': 'off',
    },
    plugins: ['@typescript-eslint'],
  },
];
`;

    try {
      fs.writeFileSync(configPath, newConfig, 'utf8');

      // 3. Wait for file watcher to trigger and config to be sent
      await new Promise(resolve => setTimeout(resolve, 2000));

      // 4. Trigger diagnostic refresh by editing the document
      await editor.edit(eb => {
        eb.insert(new vscode.Position(0, 0), ' ');
      });
      await editor.edit(eb => {
        eb.delete(new vscode.Range(0, 0, 0, 1));
      });

      // 5. Wait for updated diagnostics
      const updatedDiags = await waitForDiagnostics(doc, diags =>
        diags.some(d => d.message.includes('no-explicit-any')),
      );

      assert.ok(
        updatedDiags.some(d => d.message.includes('no-explicit-any')),
        'After hot reload, diagnostics should include no-explicit-any',
      );
      assert.ok(
        !updatedDiags.some(d => d.message.includes('no-unsafe-member-access')),
        'After hot reload, no-unsafe-member-access should be gone',
      );
    } finally {
      // 6. Restore original config
      fs.writeFileSync(configPath, originalConfig, 'utf8');
    }
  });

  test('deleting JS config should clear diagnostics', async () => {
    const doc = await openFixture('index.ts');
    await vscode.window.showTextDocument(doc);

    // 1. Verify initial diagnostics exist
    const initialDiags = await waitForDiagnostics(doc, diags =>
      diags.some(d => d.message.includes('no-unsafe-member-access')),
    );
    assert.ok(
      initialDiags.length > 0,
      'Should have diagnostics before deleting config',
    );

    const configPath = path.join(getWorkspaceRoot(), 'rslint.config.js');
    const originalConfig = fs.readFileSync(configPath, 'utf8');

    try {
      // 2. Delete the JS config file
      fs.unlinkSync(configPath);

      // 3. Wait for file watcher to trigger onDidDelete
      await new Promise(resolve => setTimeout(resolve, 3000));

      // 4. Diagnostics should eventually clear (no JS config, JSON config
      //    has different rules — we check that the old JS config rules are gone)
      const afterDeleteDiags = await waitForDiagnostics(
        doc,
        diags =>
          // Either diagnostics cleared entirely or switched to JSON config rules
          !diags.some(d => d.message.includes('no-unsafe-member-access')),
      );
      assert.ok(
        !afterDeleteDiags.some(d =>
          d.message.includes('no-unsafe-member-access'),
        ),
        'After deleting JS config, no-unsafe-member-access should be gone',
      );
    } finally {
      // 5. Restore config
      fs.writeFileSync(configPath, originalConfig, 'utf8');
      // Wait for watcher to pick up the restore
      await new Promise(resolve => setTimeout(resolve, 2000));
    }
  });

  test('creating a new JS config should load it and produce diagnostics', async () => {
    const configPath = path.join(getWorkspaceRoot(), 'rslint.config.js');
    const originalConfig = fs.readFileSync(configPath, 'utf8');

    // 1. Delete the config first to start from a no-JS-config state
    fs.unlinkSync(configPath);
    await new Promise(resolve => setTimeout(resolve, 3000));

    const doc = await openFixture('index.ts');
    await vscode.window.showTextDocument(doc);

    try {
      // 2. Create a new JS config with a different rule
      const newConfig = `export default [
  {
    files: ['**/*.ts'],
    languageOptions: {
      parserOptions: {
        projectService: false,
        project: ['./tsconfig.json'],
      },
    },
    rules: {
      '@typescript-eslint/no-explicit-any': 'error',
    },
    plugins: ['@typescript-eslint'],
  },
];
`;
      fs.writeFileSync(configPath, newConfig, 'utf8');

      // 3. Wait for file watcher to trigger onDidCreate
      await new Promise(resolve => setTimeout(resolve, 3000));

      // 4. Diagnostics should reflect the new config
      const afterCreateDiags = await waitForDiagnostics(doc, diags =>
        diags.some(d => d.message.includes('no-explicit-any')),
      );
      assert.ok(
        afterCreateDiags.some(d => d.message.includes('no-explicit-any')),
        'After creating new JS config, should see no-explicit-any diagnostic',
      );
    } finally {
      // 5. Restore original config
      fs.writeFileSync(configPath, originalConfig, 'utf8');
      await new Promise(resolve => setTimeout(resolve, 2000));
    }
  });
});

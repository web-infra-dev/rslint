import * as assert from 'assert';
import * as vscode from 'vscode';
import path from 'node:path';
import fs from 'node:fs';

suite('rslint monorepo multi-config support', function () {
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

  async function openFile(relativePath: string): Promise<vscode.TextDocument> {
    const filePath = path.join(getWorkspaceRoot(), relativePath);
    return vscode.workspace.openTextDocument(filePath);
  }

  async function triggerRelint(editor: vscode.TextEditor): Promise<void> {
    await editor.edit(eb => {
      eb.insert(new vscode.Position(0, 0), ' ');
    });
    await editor.edit(eb => {
      eb.delete(new vscode.Range(0, 0, 0, 1));
    });
  }

  // ======== Basic multi-config resolution ========

  test('root file should use root config (no-explicit-any: error)', async () => {
    const doc = await openFile('src/index.ts');
    await vscode.window.showTextDocument(doc);

    const diagnostics = await waitForDiagnostics(doc, diags =>
      diags.some(d => d.message.includes('no-explicit-any')),
    );

    assert.ok(
      diagnostics.some(d => d.message.includes('no-explicit-any')),
      'Root file should see no-explicit-any from root config',
    );
    assert.ok(
      !diagnostics.some(d => d.message.includes('no-unsafe-member-access')),
      'Root file should NOT see no-unsafe-member-access (off in root config)',
    );
  });

  test('foo sub-package file should use foo config (no-unsafe-member-access: error)', async () => {
    const doc = await openFile('packages/foo/src/index.ts');
    await vscode.window.showTextDocument(doc);

    const diagnostics = await waitForDiagnostics(doc, diags =>
      diags.some(d => d.message.includes('no-unsafe-member-access')),
    );

    assert.ok(
      diagnostics.some(d => d.message.includes('no-unsafe-member-access')),
      'Foo file should see no-unsafe-member-access from foo config',
    );
    assert.ok(
      !diagnostics.some(d => d.message.includes('no-explicit-any')),
      'Foo file should NOT see no-explicit-any (off in foo config)',
    );
  });

  test('bar sub-package file should fall back to root config (no sub-config)', async () => {
    const doc = await openFile('packages/bar/src/index.ts');
    await vscode.window.showTextDocument(doc);

    const diagnostics = await waitForDiagnostics(doc, diags =>
      diags.some(d => d.message.includes('no-explicit-any')),
    );

    assert.ok(
      diagnostics.some(d => d.message.includes('no-explicit-any')),
      'Bar file should see no-explicit-any from root config (fallback)',
    );
    assert.ok(
      !diagnostics.some(d => d.message.includes('no-unsafe-member-access')),
      'Bar file should NOT see no-unsafe-member-access (off in root config)',
    );
  });

  // ======== Broken config resilience ========

  test('broken sub-package config should not prevent other configs from loading', async () => {
    // The "broken" package has a syntactically invalid rslint.config.js.
    // Despite this, foo's valid config should still work correctly.
    const doc = await openFile('packages/foo/src/index.ts');
    await vscode.window.showTextDocument(doc);

    const diagnostics = await waitForDiagnostics(doc, diags =>
      diags.some(d => d.message.includes('no-unsafe-member-access')),
    );

    assert.ok(
      diagnostics.some(d => d.message.includes('no-unsafe-member-access')),
      'Foo file should still use foo config despite broken sibling config',
    );
  });

  test('broken sub-package file should fall back to root config', async () => {
    // broken/ has an invalid config, so its files should fall back to root config
    const doc = await openFile('packages/broken/src/index.ts');
    await vscode.window.showTextDocument(doc);

    const diagnostics = await waitForDiagnostics(doc, diags =>
      diags.some(d => d.message.includes('no-explicit-any')),
    );

    assert.ok(
      diagnostics.some(d => d.message.includes('no-explicit-any')),
      'Broken package file should fall back to root config (no-explicit-any: error)',
    );
  });

  // ======== Config hot reload in monorepo ========

  test('changing foo sub-package config should update foo file diagnostics', async () => {
    const doc = await openFile('packages/foo/src/index.ts');
    const editor = await vscode.window.showTextDocument(doc);

    // 1. Verify initial: foo config has no-unsafe-member-access: error
    const initialDiags = await waitForDiagnostics(doc, diags =>
      diags.some(d => d.message.includes('no-unsafe-member-access')),
    );
    assert.ok(
      initialDiags.some(d => d.message.includes('no-unsafe-member-access')),
      'Initial: foo file should have no-unsafe-member-access',
    );

    // 2. Change foo config to enable no-explicit-any instead
    const fooConfigPath = path.join(
      getWorkspaceRoot(),
      'packages/foo/rslint.config.js',
    );
    const originalConfig = fs.readFileSync(fooConfigPath, 'utf8');

    const newConfig = `export default [
  {
    files: ['src/**/*.ts'],
    languageOptions: {
      parserOptions: {
        projectService: false,
        project: ['../../tsconfig.json'],
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
      fs.writeFileSync(fooConfigPath, newConfig, 'utf8');
      await new Promise(resolve => setTimeout(resolve, 2000));
      await triggerRelint(editor);

      const updatedDiags = await waitForDiagnostics(doc, diags =>
        diags.some(d => d.message.includes('no-explicit-any')),
      );

      assert.ok(
        updatedDiags.some(d => d.message.includes('no-explicit-any')),
        'After change: foo file should see no-explicit-any',
      );
      assert.ok(
        !updatedDiags.some(d => d.message.includes('no-unsafe-member-access')),
        'After change: foo file should NOT see no-unsafe-member-access',
      );
    } finally {
      fs.writeFileSync(fooConfigPath, originalConfig, 'utf8');
      await new Promise(resolve => setTimeout(resolve, 2000));
    }
  });

  test('deleting foo sub-package config should make foo file fall back to root config', async () => {
    const doc = await openFile('packages/foo/src/index.ts');
    const editor = await vscode.window.showTextDocument(doc);

    // 1. Verify initial: foo config has no-unsafe-member-access: error
    const initialDiags = await waitForDiagnostics(doc, diags =>
      diags.some(d => d.message.includes('no-unsafe-member-access')),
    );
    assert.ok(
      initialDiags.some(d => d.message.includes('no-unsafe-member-access')),
      'Initial: foo file should have no-unsafe-member-access',
    );

    // 2. Delete foo's config
    const fooConfigPath = path.join(
      getWorkspaceRoot(),
      'packages/foo/rslint.config.js',
    );
    const originalConfig = fs.readFileSync(fooConfigPath, 'utf8');

    try {
      fs.unlinkSync(fooConfigPath);
      await new Promise(resolve => setTimeout(resolve, 3000));
      await triggerRelint(editor);

      // 3. Foo file should now fall back to root config (no-explicit-any: error)
      const afterDeleteDiags = await waitForDiagnostics(doc, diags =>
        diags.some(d => d.message.includes('no-explicit-any')),
      );

      assert.ok(
        afterDeleteDiags.some(d => d.message.includes('no-explicit-any')),
        'After delete: foo file should fall back to root config (no-explicit-any)',
      );
      assert.ok(
        !afterDeleteDiags.some(d =>
          d.message.includes('no-unsafe-member-access'),
        ),
        'After delete: foo file should NOT see no-unsafe-member-access (off in root)',
      );
    } finally {
      fs.writeFileSync(fooConfigPath, originalConfig, 'utf8');
      await new Promise(resolve => setTimeout(resolve, 2000));
    }
  });

  test('corrupting foo config should not break other configs', async () => {
    const fooDoc = await openFile('packages/foo/src/index.ts');
    await vscode.window.showTextDocument(fooDoc);

    // 1. Verify initial: foo config works
    const initialDiags = await waitForDiagnostics(fooDoc, diags =>
      diags.some(d => d.message.includes('no-unsafe-member-access')),
    );
    assert.ok(
      initialDiags.some(d => d.message.includes('no-unsafe-member-access')),
      'Initial: foo file should have no-unsafe-member-access',
    );

    // 2. Corrupt foo's config with syntax error
    const fooConfigPath = path.join(
      getWorkspaceRoot(),
      'packages/foo/rslint.config.js',
    );
    const originalConfig = fs.readFileSync(fooConfigPath, 'utf8');

    try {
      fs.writeFileSync(fooConfigPath, 'export default [BROKEN SYNTAX;', 'utf8');
      await new Promise(resolve => setTimeout(resolve, 2000));

      // 3. Root config should still work for bar
      const barDoc = await openFile('packages/bar/src/index.ts');
      await vscode.window.showTextDocument(barDoc);

      const barDiags = await waitForDiagnostics(barDoc, diags =>
        diags.some(d => d.message.includes('no-explicit-any')),
      );
      assert.ok(
        barDiags.some(d => d.message.includes('no-explicit-any')),
        'Bar file should still use root config after foo config is corrupted',
      );
    } finally {
      fs.writeFileSync(fooConfigPath, originalConfig, 'utf8');
      await new Promise(resolve => setTimeout(resolve, 2000));
    }
  });

  test('changing root config should update bar file diagnostics (no sub-config)', async () => {
    const doc = await openFile('packages/bar/src/index.ts');
    const editor = await vscode.window.showTextDocument(doc);

    // 1. Verify initial: bar uses root config (no-explicit-any: error)
    const initialDiags = await waitForDiagnostics(doc, diags =>
      diags.some(d => d.message.includes('no-explicit-any')),
    );
    assert.ok(
      initialDiags.some(d => d.message.includes('no-explicit-any')),
      'Initial: bar file should have no-explicit-any from root config',
    );

    // 2. Change root config to enable no-unsafe-member-access instead
    const rootConfigPath = path.join(getWorkspaceRoot(), 'rslint.config.js');
    const originalConfig = fs.readFileSync(rootConfigPath, 'utf8');

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
      '@typescript-eslint/no-explicit-any': 'off',
      '@typescript-eslint/no-unsafe-member-access': 'error',
    },
    plugins: ['@typescript-eslint'],
  },
];
`;

    try {
      fs.writeFileSync(rootConfigPath, newConfig, 'utf8');
      await new Promise(resolve => setTimeout(resolve, 2000));
      await triggerRelint(editor);

      const updatedDiags = await waitForDiagnostics(doc, diags =>
        diags.some(d => d.message.includes('no-unsafe-member-access')),
      );

      assert.ok(
        updatedDiags.some(d => d.message.includes('no-unsafe-member-access')),
        'After change: bar file should see no-unsafe-member-access from updated root',
      );
      assert.ok(
        !updatedDiags.some(d => d.message.includes('no-explicit-any')),
        'After change: bar file should NOT see no-explicit-any (off in updated root)',
      );
    } finally {
      fs.writeFileSync(rootConfigPath, originalConfig, 'utf8');
      await new Promise(resolve => setTimeout(resolve, 2000));
    }
  });
});

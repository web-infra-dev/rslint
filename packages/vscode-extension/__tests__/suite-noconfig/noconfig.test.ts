import * as assert from 'assert';
import * as vscode from 'vscode';
import path from 'node:path';
import fs from 'node:fs';

suite('rslint no config fallback', function () {
  this.timeout(60000);

  function getWorkspaceRoot(): string {
    return vscode.workspace.workspaceFolders![0].uri.fsPath;
  }

  async function openFixture(filename: string): Promise<vscode.TextDocument> {
    const filePath = path.join(getWorkspaceRoot(), 'src', filename);
    return vscode.workspace.openTextDocument(filePath);
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

  /**
   * Trigger a no-op edit cycle on the document to force diagnostic refresh.
   */
  async function triggerDiagnosticRefresh(
    doc: vscode.TextDocument,
  ): Promise<void> {
    const editor = await vscode.window.showTextDocument(doc);
    await editor.edit(eb => {
      eb.insert(new vscode.Position(0, 0), ' ');
    });
    await editor.edit(eb => {
      eb.delete(new vscode.Range(0, 0, 0, 1));
    });
  }

  test('extension should start without config file and produce no diagnostics', async () => {
    const doc = await openFixture('index.ts');
    await vscode.window.showTextDocument(doc);

    // Wait a reasonable time for diagnostics to appear (they shouldn't)
    await new Promise(resolve => setTimeout(resolve, 5000));

    const diagnostics = vscode.languages.getDiagnostics(doc.uri);

    // Without any config, no rslint rules are enabled, so no diagnostics
    assert.strictEqual(
      diagnostics.filter(d => d.source === 'rslint').length,
      0,
      'Should have no rslint diagnostics when no config file exists',
    );
  });

  test('config lifecycle: no config → JSON → JS override → delete JS → delete JSON', async () => {
    const root = getWorkspaceRoot();
    const jsonConfigPath = path.join(root, 'rslint.json');
    const jsConfigPath = path.join(root, 'rslint.config.js');

    // Ensure clean state — remove any leftover config files
    try {
      fs.unlinkSync(jsonConfigPath);
    } catch {
      /* ignore */
    }
    try {
      fs.unlinkSync(jsConfigPath);
    } catch {
      /* ignore */
    }

    const doc = await openFixture('index.ts');
    await vscode.window.showTextDocument(doc);

    try {
      // ── Step 1: No config → no diagnostics ──
      await new Promise(resolve => setTimeout(resolve, 3000));
      let diags = vscode.languages.getDiagnostics(doc.uri);
      assert.strictEqual(
        diags.filter(d => d.source === 'rslint').length,
        0,
        'Step 1: Should have no rslint diagnostics without config',
      );

      // ── Step 2: Create rslint.json → no-explicit-any diagnostics ──
      const jsonConfig = JSON.stringify(
        [
          {
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
        ],
        null,
        2,
      );
      fs.writeFileSync(jsonConfigPath, jsonConfig, 'utf8');

      // Wait for file watcher to detect JSON config creation
      await new Promise(resolve => setTimeout(resolve, 3000));
      await triggerDiagnosticRefresh(doc);

      diags = await waitForDiagnostics(doc, ds =>
        ds.some(d => d.message.includes('no-explicit-any')),
      );
      assert.ok(
        diags.some(d => d.message.includes('no-explicit-any')),
        'Step 2: Should see no-explicit-any after JSON config created',
      );
      assert.ok(
        !diags.some(d => d.message.includes('no-unsafe-member-access')),
        'Step 2: Should NOT see no-unsafe-member-access (turned off in JSON)',
      );

      // ── Step 3: Create rslint.config.js → JS overrides JSON ──
      const jsConfig = `export default [
  {
    languageOptions: {
      parserOptions: {
        projectService: false,
        project: ['./tsconfig.json'],
      },
    },
    rules: {
      '@typescript-eslint/no-unsafe-member-access': 'error',
      '@typescript-eslint/no-explicit-any': 'off',
    },
    plugins: ['@typescript-eslint'],
  },
];
`;
      fs.writeFileSync(jsConfigPath, jsConfig, 'utf8');

      // Wait for JS config watcher to detect and send rslint/configUpdate
      await new Promise(resolve => setTimeout(resolve, 3000));
      await triggerDiagnosticRefresh(doc);

      diags = await waitForDiagnostics(doc, ds =>
        ds.some(d => d.message.includes('no-unsafe-member-access')),
      );
      assert.ok(
        diags.some(d => d.message.includes('no-unsafe-member-access')),
        'Step 3: Should see no-unsafe-member-access from JS config',
      );
      assert.ok(
        !diags.some(d => d.message.includes('no-explicit-any')),
        'Step 3: JS config should override JSON — no-explicit-any should be off',
      );

      // ── Step 4: Delete JS config → JSON config restored ──
      fs.unlinkSync(jsConfigPath);

      await new Promise(resolve => setTimeout(resolve, 3000));
      await triggerDiagnosticRefresh(doc);

      diags = await waitForDiagnostics(doc, ds =>
        ds.some(d => d.message.includes('no-explicit-any')),
      );
      assert.ok(
        diags.some(d => d.message.includes('no-explicit-any')),
        'Step 4: After deleting JS config, JSON config should restore no-explicit-any',
      );
      assert.ok(
        !diags.some(d => d.message.includes('no-unsafe-member-access')),
        'Step 4: no-unsafe-member-access should be gone (JS config deleted)',
      );

      // ── Step 5: Delete JSON config → no diagnostics ──
      fs.unlinkSync(jsonConfigPath);

      await new Promise(resolve => setTimeout(resolve, 3000));
      await triggerDiagnosticRefresh(doc);

      // Wait for diagnostics to clear
      diags = await waitForDiagnostics(
        doc,
        ds => ds.filter(d => d.source === 'rslint').length === 0,
      );
      assert.strictEqual(
        diags.filter(d => d.source === 'rslint').length,
        0,
        'Step 5: After deleting all configs, should have no rslint diagnostics',
      );
    } finally {
      // Clean up — ensure no config files remain for other test runs
      try {
        fs.unlinkSync(jsonConfigPath);
      } catch {
        /* ignore */
      }
      try {
        fs.unlinkSync(jsConfigPath);
      } catch {
        /* ignore */
      }
    }
  });
});

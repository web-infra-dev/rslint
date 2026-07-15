import * as assert from 'assert';
import * as vscode from 'vscode';
import path from 'node:path';
import fs from 'node:fs';
import { waitForRslintDiagnostics as waitForDiagnostics } from '../utils/diagnostics';

suite('rslint no config fallback', function () {
  this.timeout(120000);

  function getWorkspaceRoot(): string {
    return vscode.workspace.workspaceFolders![0].uri.fsPath;
  }

  async function openFixture(filename: string): Promise<vscode.TextDocument> {
    const filePath = path.join(getWorkspaceRoot(), 'src', filename);
    return vscode.workspace.openTextDocument(filePath);
  }

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

  async function withConfigFilesAbsent(
    testFn: (paths: { json: string; js: string }) => Promise<void>,
  ): Promise<void> {
    const paths = {
      json: path.join(getWorkspaceRoot(), 'rslint.json'),
      js: path.join(getWorkspaceRoot(), 'rslint.config.js'),
    };
    const removeConfigs = (): void => {
      fs.rmSync(paths.json, { force: true });
      fs.rmSync(paths.js, { force: true });
      if (fs.existsSync(paths.json) || fs.existsSync(paths.js)) {
        throw new Error('No-config fixture cleanup left a config file behind');
      }
    };

    let testError: unknown;
    try {
      removeConfigs();
      await testFn(paths);
    } catch (error) {
      testError = error;
    }

    let cleanupError: unknown;
    try {
      removeConfigs();
    } catch (error) {
      cleanupError = error;
    }
    if (testError && cleanupError) {
      throw new AggregateError(
        [testError, cleanupError],
        'No-config test and config cleanup both failed',
      );
    }
    if (testError) throw testError;
    if (cleanupError) throw cleanupError;
  }

  /**
   * Trigger a no-op edit cycle on the document to force diagnostic refresh.
   */
  async function triggerDiagnosticRefresh(
    doc: vscode.TextDocument,
  ): Promise<void> {
    const editor = await vscode.window.showTextDocument(doc);
    await editor.edit((eb) => {
      eb.insert(new vscode.Position(0, 0), ' ');
    });
    await editor.edit((eb) => {
      eb.delete(new vscode.Range(0, 0, 0, 1));
    });
  }

  test('removing the last config reaches an observed no-config state', async () => {
    await withConfigFilesAbsent(async ({ json }) => {
      const doc = await openFixture('index.ts');
      await vscode.window.showTextDocument(doc);

      // Establish a positive rslint publication first. The subsequent empty
      // snapshot therefore cannot be the document's not-yet-linted state.
      const configured = waitForDiagnostics(doc, (diagnostics) =>
        diagnostics.some((diagnostic) =>
          diagnostic.message.includes('no-explicit-any'),
        ),
      );
      fs.writeFileSync(json, jsonConfig, 'utf8');
      await triggerDiagnosticRefresh(doc);
      await configured;

      const cleared = waitForDiagnostics(
        doc,
        (diagnostics) => diagnostics.length === 0,
      );
      fs.rmSync(json);
      await triggerDiagnosticRefresh(doc);
      const diagnostics = await cleared;
      assert.strictEqual(
        diagnostics.length,
        0,
        'Removing the last config should publish an empty rslint snapshot',
      );
    });
  });

  test('config lifecycle: no config → JSON → JS override → delete JS → delete JSON', async () => {
    await withConfigFilesAbsent(async ({ json, js }) => {
      const doc = await openFixture('index.ts');
      await vscode.window.showTextDocument(doc);
      assert.ok(
        !fs.existsSync(json) && !fs.existsSync(js),
        'The lifecycle must start with no config files',
      );

      // ── Step 1: Create rslint.json → no-explicit-any diagnostics ──
      let transition = waitForDiagnostics(doc, (ds) =>
        ds.some((d) => d.message.includes('no-explicit-any')),
      );
      fs.writeFileSync(json, jsonConfig, 'utf8');
      await triggerDiagnosticRefresh(doc);
      let diags = await transition;
      assert.ok(
        diags.some((d) => d.message.includes('no-explicit-any')),
        'Step 1: Should see no-explicit-any after JSON config created',
      );
      assert.ok(
        !diags.some((d) => d.message.includes('no-unsafe-member-access')),
        'Step 1: Should NOT see no-unsafe-member-access (turned off in JSON)',
      );

      // ── Step 2: Create rslint.config.js → JS overrides JSON ──
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
      transition = waitForDiagnostics(
        doc,
        (ds) =>
          ds.some((d) => d.message.includes('no-unsafe-member-access')) &&
          !ds.some((d) => d.message.includes('no-explicit-any')),
      );
      fs.writeFileSync(js, jsConfig, 'utf8');
      await triggerDiagnosticRefresh(doc);
      diags = await transition;
      assert.ok(
        diags.some((d) => d.message.includes('no-unsafe-member-access')),
        'Step 2: Should see no-unsafe-member-access from JS config',
      );
      assert.ok(
        !diags.some((d) => d.message.includes('no-explicit-any')),
        'Step 2: JS config should override JSON — no-explicit-any should be off',
      );

      // ── Step 3: Delete JS config → JSON config restored ──
      transition = waitForDiagnostics(
        doc,
        (ds) =>
          ds.some((d) => d.message.includes('no-explicit-any')) &&
          !ds.some((d) => d.message.includes('no-unsafe-member-access')),
      );
      fs.rmSync(js);
      await triggerDiagnosticRefresh(doc);
      diags = await transition;
      assert.ok(
        diags.some((d) => d.message.includes('no-explicit-any')),
        'Step 3: After deleting JS config, JSON config should restore no-explicit-any',
      );
      assert.ok(
        !diags.some((d) => d.message.includes('no-unsafe-member-access')),
        'Step 3: no-unsafe-member-access should be gone (JS config deleted)',
      );

      // ── Step 4: Broken JS with no last-good → explicit no-lint state ──
      const disabledByBrokenJS = waitForDiagnostics(
        doc,
        (ds) => ds.length === 0,
      );
      fs.writeFileSync(js, 'export default [BROKEN SYNTAX;', 'utf8');
      await triggerDiagnosticRefresh(doc);
      diags = await disabledByBrokenJS;
      assert.strictEqual(
        diags.length,
        0,
        'Step 4: A broken discovered JS config must not fall back to JSON',
      );

      // ── Step 5: Delete broken JS → legitimate deletion restores JSON ──
      const jsonRestored = waitForDiagnostics(doc, (ds) =>
        ds.some((d) => d.message.includes('no-explicit-any')),
      );
      fs.rmSync(js);
      await triggerDiagnosticRefresh(doc);
      diags = await jsonRestored;
      assert.ok(
        diags.some((d) => d.message.includes('no-explicit-any')),
        'Step 5: Deleting all JS configs should restore the JSON fallback',
      );

      // ── Step 6: Delete JSON config → no diagnostics ──
      const cleared = waitForDiagnostics(doc, (ds) => ds.length === 0);
      fs.rmSync(json);
      await triggerDiagnosticRefresh(doc);
      diags = await cleared;
      assert.strictEqual(
        diags.length,
        0,
        'Step 6: After deleting all configs, should have no rslint diagnostics',
      );
    });
  });
});

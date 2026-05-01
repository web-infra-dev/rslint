import * as assert from 'assert';
import * as vscode from 'vscode';
import path from 'node:path';
import fs from 'node:fs';

suite('rslint JS config support', function () {
  this.timeout(120_000);

  function getWorkspaceRoot(): string {
    return vscode.workspace.workspaceFolders![0].uri.fsPath;
  }

  /**
   * Race-free `onDidChangeDiagnostics` waiter.
   *
   * Compared to a polling `getDiagnostics` + `setTimeout` loop, this
   * subscribes BEFORE checking the initial state, eliminating the
   * subscribe-after-check window where an event fired between the
   * synchronous read and the listener registration would be lost.
   *
   * The waiter returns as soon as `predicate` accepts a snapshot — no
   * file edits, no nudges, no sleeps. It relies on the server's own
   * push channel: `internal/lsp/service.go::handleConfigUpdate` ends
   * with `s.RefreshDiagnostics(ctx)`, which signals `refreshCh`; the
   * dispatch loop in `internal/lsp/server.go` consumes that signal and
   * calls `pushDiagnostics(uri)` for every open document. So the LSP
   * server already publishes new diagnostics after a config reload —
   * tests just need to listen for them.
   *
   * Default 60s timeout (vs the older helper's 30s × 20 polling
   * budget) gives headroom for slow CI runners where the JS-config
   * reload chain (ESM dynamic import + IO) can take 10s+.
   */
  function waitForDiagnostics(
    doc: vscode.TextDocument,
    predicate?: (diags: vscode.Diagnostic[]) => boolean,
    timeoutMs = 60_000,
  ): Promise<vscode.Diagnostic[]> {
    const matches = predicate ?? ((ds) => ds.length > 0);
    return new Promise<vscode.Diagnostic[]>((resolve, reject) => {
      let done = false;
      const finish = (
        value: vscode.Diagnostic[] | undefined,
        err?: Error,
      ): void => {
        if (done) return;
        done = true;
        sub.dispose();
        clearTimeout(timer);
        if (err) reject(err);
        else resolve(value!);
      };
      const sub = vscode.languages.onDidChangeDiagnostics((e) => {
        if (!e.uris.some((u) => u.toString() === doc.uri.toString())) return;
        const ds = vscode.languages.getDiagnostics(doc.uri);
        if (matches(ds)) finish(ds);
      });
      const timer = setTimeout(
        () =>
          finish(
            undefined,
            new Error('waitForDiagnostics: timeout waiting for predicate'),
          ),
        timeoutMs,
      );
      // Check current state in case the matching publish already
      // happened before subscription. Safe because finish() is
      // idempotent.
      const initial = vscode.languages.getDiagnostics(doc.uri);
      if (matches(initial)) finish(initial);
    });
  }

  async function openFixture(filename: string): Promise<vscode.TextDocument> {
    const filePath = path.join(getWorkspaceRoot(), 'src', filename);
    return vscode.workspace.openTextDocument(filePath);
  }

  test('JS config should produce diagnostics', async () => {
    const doc = await openFixture('index.ts');
    await vscode.window.showTextDocument(doc);

    // Wait specifically for JS config diagnostics. JSON config may load first
    // (server-side) before JS config arrives via rslint/configUpdate, so we
    // must use a predicate to avoid returning early with JSON config results.
    const diagnostics = await waitForDiagnostics(doc, (diags) =>
      diags.some((d) => d.message.includes('no-unsafe-member-access')),
    );
    assert.ok(
      diagnostics.length > 0,
      `Expected diagnostics but got ${diagnostics.length}`,
    );
    assert.ok(
      diagnostics.some((d) => d.message.includes('no-unsafe-member-access')),
      'Expected no-unsafe-member-access diagnostic from JS config',
    );
  });

  test('JS config should take priority over JSON config', async () => {
    const doc = await openFixture('index.ts');
    await vscode.window.showTextDocument(doc);

    // Wait specifically for JS config diagnostics (no-unsafe-member-access).
    // JSON config may load first with no-explicit-any, but JS config should
    // override via rslint/configUpdate notification.
    const diagnostics = await waitForDiagnostics(doc, (diags) =>
      diags.some((d) => d.message.includes('no-unsafe-member-access')),
    );

    assert.ok(
      diagnostics.some((d) => d.message.includes('no-unsafe-member-access')),
      'Expected no-unsafe-member-access from JS config',
    );
    assert.ok(
      !diagnostics.some((d) => d.message.includes('no-explicit-any')),
      'Should NOT see no-explicit-any because JS config takes priority over JSON',
    );
  });

  test('config hot reload should update diagnostics', async () => {
    const doc = await openFixture('index.ts');
    await vscode.window.showTextDocument(doc);

    // 1. Verify initial diagnostics have no-unsafe-member-access.
    await waitForDiagnostics(doc, (diags) =>
      diags.some((d) => d.message.includes('no-unsafe-member-access')),
    );

    // 2. Subscribe BEFORE writing the new config — eliminates the
    //    "publish fires between write and subscribe" race window. The
    //    waiter listens; the server pushes after handleConfigUpdate
    //    via RefreshDiagnostics; we resolve.
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
    const reloaded = waitForDiagnostics(
      doc,
      (diags) =>
        diags.some((d) => d.message.includes('no-explicit-any')) &&
        !diags.some((d) => d.message.includes('no-unsafe-member-access')),
    );

    try {
      fs.writeFileSync(configPath, newConfig, 'utf8');
      const updatedDiags = await reloaded;
      assert.ok(
        updatedDiags.some((d) => d.message.includes('no-explicit-any')),
        'After hot reload, diagnostics should include no-explicit-any',
      );
      assert.ok(
        !updatedDiags.some((d) =>
          d.message.includes('no-unsafe-member-access'),
        ),
        'After hot reload, no-unsafe-member-access should be gone',
      );
    } finally {
      fs.writeFileSync(configPath, originalConfig, 'utf8');
    }
  });

  test('deleting JS config should clear diagnostics', async () => {
    const doc = await openFixture('index.ts');
    await vscode.window.showTextDocument(doc);

    await waitForDiagnostics(doc, (diags) =>
      diags.some((d) => d.message.includes('no-unsafe-member-access')),
    );

    const configPath = path.join(getWorkspaceRoot(), 'rslint.config.js');
    const originalConfig = fs.readFileSync(configPath, 'utf8');

    const cleared = waitForDiagnostics(
      doc,
      (diags) =>
        !diags.some((d) => d.message.includes('no-unsafe-member-access')),
    );

    try {
      fs.unlinkSync(configPath);
      const afterDeleteDiags = await cleared;
      assert.ok(
        !afterDeleteDiags.some((d) =>
          d.message.includes('no-unsafe-member-access'),
        ),
        'After deleting JS config, no-unsafe-member-access should be gone',
      );
    } finally {
      // Restore config and wait for the reload to settle so the next
      // test starts from a known state. Subscribe first, then write.
      const restored = waitForDiagnostics(doc, (diags) =>
        diags.some((d) => d.message.includes('no-unsafe-member-access')),
      ).catch(() => undefined);
      fs.writeFileSync(configPath, originalConfig, 'utf8');
      await restored;
    }
  });

  test('creating a new JS config should load it and produce diagnostics', async () => {
    const configPath = path.join(getWorkspaceRoot(), 'rslint.config.js');
    const originalConfig = fs.readFileSync(configPath, 'utf8');
    const doc = await openFixture('index.ts');
    await vscode.window.showTextDocument(doc);

    // Step 1: delete existing config and wait until its diagnostics drop.
    const cleared = waitForDiagnostics(doc, (diags) =>
      diags.every((d) => !d.message.includes('no-unsafe-member-access')),
    );
    fs.unlinkSync(configPath);
    await cleared;

    try {
      // Step 2: subscribe for the new rule, then create the file.
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
      const created = waitForDiagnostics(doc, (diags) =>
        diags.some((d) => d.message.includes('no-explicit-any')),
      );
      fs.writeFileSync(configPath, newConfig, 'utf8');
      const afterCreateDiags = await created;
      assert.ok(
        afterCreateDiags.some((d) => d.message.includes('no-explicit-any')),
        'After creating new JS config, should see no-explicit-any diagnostic',
      );
    } finally {
      const restored = waitForDiagnostics(doc, (diags) =>
        diags.some((d) => d.message.includes('no-unsafe-member-access')),
      ).catch(() => undefined);
      fs.writeFileSync(configPath, originalConfig, 'utf8');
      await restored;
    }
  });
});

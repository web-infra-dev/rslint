/**
 * VS Code e2e — mixed native+plugin & gap-file+plugin coverage.
 *
 * Workspace: __tests__/fixtures-eslint-plugin-mixed/
 * The config enables a native syntax rule (`no-debugger`) AND a plugin
 * rule (`fx/no-forbidden`) together. tsconfig `include` is `src/` only,
 * so `scripts/gap.ts` is a gap file.
 *
 *   (e) an in-tsconfig file (`src/index.ts`) must surface BOTH the
 *       native and the plugin diagnostic — the LSP mixed path.
 *   (f) a gap file (`scripts/gap.ts`, outside tsconfig include) must
 *       still surface the plugin diagnostic.
 *
 * These two LSP-side combinations had no e2e coverage: the existing
 * plugin suites are pure-plugin (zero native rules) and keep every
 * source file inside tsconfig include.
 */
import * as assert from 'assert';
import * as vscode from 'vscode';
import path from 'node:path';

suite('rslint mixed native+plugin and gap-file+plugin support', function () {
  this.timeout(120_000);

  function getWorkspaceRoot(): string {
    return vscode.workspace.workspaceFolders![0].uri.fsPath;
  }

  // Race-free diagnostics waiter — copied from suite-eslint-plugin
  // (subscribe BEFORE the initial check + idempotent finish to avoid the
  // subscribe-after-publish window).
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
      const initial = vscode.languages.getDiagnostics(doc.uri);
      if (matches(initial)) finish(initial);
    });
  }

  const isNative = (d: vscode.Diagnostic): boolean =>
    d.source === 'rslint' && d.message.includes('[no-debugger]');
  const isPlugin = (d: vscode.Diagnostic): boolean =>
    d.source === 'rslint' && d.message.includes('[fx/no-forbidden]');

  test('(e) same in-tsconfig file surfaces BOTH a native and a plugin diagnostic', async () => {
    const filePath = path.join(getWorkspaceRoot(), 'src', 'index.ts');
    const doc = await vscode.workspace.openTextDocument(filePath);
    await vscode.window.showTextDocument(doc);

    // Wait until BOTH a native (no-debugger) and a plugin
    // (fx/no-forbidden) diagnostic are present on the same file.
    const diags = await waitForDiagnostics(
      doc,
      (ds) => ds.some(isNative) && ds.some(isPlugin),
    );

    assert.ok(
      diags.some(isNative),
      'expected the native no-debugger diagnostic on src/index.ts',
    );
    const pluginHits = diags.filter(isPlugin);
    assert.ok(
      pluginHits.length >= 1,
      `expected ≥1 fx/no-forbidden on src/index.ts, got ${pluginHits.length}`,
    );
    for (const hit of pluginHits) {
      assert.strictEqual(doc.getText(hit.range), 'forbidden');
    }
  });

  test('(f) gap file outside tsconfig include still surfaces the plugin diagnostic', async () => {
    const filePath = path.join(getWorkspaceRoot(), 'scripts', 'gap.ts');
    const doc = await vscode.workspace.openTextDocument(filePath);
    await vscode.window.showTextDocument(doc);

    const diags = await waitForDiagnostics(doc, (ds) => ds.some(isPlugin));

    const pluginHits = diags.filter(isPlugin);
    assert.ok(
      pluginHits.length >= 1,
      `expected ≥1 fx/no-forbidden on the gap file, got ${pluginHits.length}`,
    );
    for (const hit of pluginHits) {
      assert.strictEqual(
        doc.getText(hit.range),
        'forbidden',
        'gap-file plugin hit flagged unexpected text',
      );
    }
  });
});

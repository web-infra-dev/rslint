import * as assert from 'assert';
import * as vscode from 'vscode';
import path from 'node:path';

// Type-aware rule scope when parserOptions uses `projectService: true`
// (the shape `ts.configs.recommended` exports) without an explicit
// `project`. The LSP and CLI must agree: only files covered by the
// fallback tsconfig's `include` get type-aware rules.
//
// Fixture: fixtures-project-service-scope
//   - rslint.config.js  — parserOptions.projectService: true (no explicit project)
//   - tsconfig.json     — include: ["src"]
//   - src/covered.ts    — IN tsconfig: no-unused-vars SHOULD fire
//   - test/skills.test.ts — NOT IN tsconfig: no-unused-vars should NOT fire
suite('rslint projectService type-aware scope', function () {
  this.timeout(60000);

  function workspaceRoot(): string {
    return vscode.workspace.workspaceFolders![0].uri.fsPath;
  }

  async function waitForDiagnostics(
    doc: vscode.TextDocument,
    predicate: (diags: vscode.Diagnostic[]) => boolean,
  ): Promise<vscode.Diagnostic[]> {
    for (let i = 0; i < 20; i++) {
      const current = vscode.languages.getDiagnostics(doc.uri);
      if (predicate(current)) return current;
      await new Promise((resolve) => {
        let timer: ReturnType<typeof setTimeout>;
        const disposable = vscode.languages.onDidChangeDiagnostics((e) => {
          if (e.uris.some((uri) => uri.toString() === doc.uri.toString())) {
            clearTimeout(timer);
            disposable.dispose();
            resolve(void 0);
          }
        });
        timer = setTimeout(() => {
          disposable.dispose();
          resolve(void 0);
        }, 1500);
      });
    }
    return vscode.languages.getDiagnostics(doc.uri);
  }

  // Only inspect rslint-originated diagnostics — TS's own 6133 ("declared but
  // never read") is also emitted on the same lines and would otherwise confuse
  // the assertion.
  function rslintDiagnostics(diags: vscode.Diagnostic[]): vscode.Diagnostic[] {
    return diags.filter((d) => d.source === 'rslint');
  }

  test('src/covered.ts (in tsconfig include) — no-unused-vars SHOULD fire', async () => {
    const doc = await vscode.workspace.openTextDocument(
      path.join(workspaceRoot(), 'src/covered.ts'),
    );
    await vscode.window.showTextDocument(doc);

    const diagnostics = await waitForDiagnostics(doc, (diags) =>
      rslintDiagnostics(diags).some((d) =>
        d.message.includes('no-unused-vars'),
      ),
    );

    const rslintDiags = rslintDiagnostics(diagnostics);
    assert.ok(
      rslintDiags.some((d) => d.message.includes('no-unused-vars')),
      `Expected no-unused-vars on src/covered.ts. Got: ${rslintDiags.map((d) => d.message).join(' | ')}`,
    );
  });

  test('test/skills.test.ts (outside tsconfig include) — no-unused-vars should NOT fire', async () => {
    const doc = await vscode.workspace.openTextDocument(
      path.join(workspaceRoot(), 'test/skills.test.ts'),
    );
    await vscode.window.showTextDocument(doc);

    // Give the server time to publish diagnostics. Waiting for a specific
    // predicate is unreliable here because the correct outcome is "no rslint
    // diagnostics" — we instead wait a bounded amount of time.
    await new Promise((r) => setTimeout(r, 5000));
    const rslintDiags = rslintDiagnostics(
      vscode.languages.getDiagnostics(doc.uri),
    );

    assert.ok(
      !rslintDiags.some((d) => d.message.includes('no-unused-vars')),
      `no-unused-vars should NOT fire on a file outside tsconfig.include. Got: ${rslintDiags.map((d) => d.message).join(' | ')}`,
    );
  });
});

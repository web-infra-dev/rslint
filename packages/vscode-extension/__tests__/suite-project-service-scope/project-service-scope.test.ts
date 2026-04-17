import * as assert from 'assert';
import * as vscode from 'vscode';
import path from 'node:path';

// Type-aware rule scope when parserOptions uses `projectService: true`
// (the shape `ts.configs.recommended` exports) without an explicit
// `project`. The LSP and CLI must agree: only files covered by the
// fallback tsconfig's `include` get type-aware rules, AND a nested
// config that has no tsconfig must not leak allow-all onto sibling
// configs' files.
//
// Fixture: fixtures-project-service-scope
//   - rslint.config.js            — parserOptions.projectService: true (no explicit project).
//                                    Also enables `no-console` as a non-type-aware marker
//                                    rule so the negative test case can wait for rslint to
//                                    finish linting a file without a fixed-duration sleep.
//   - tsconfig.json               — include: ["src"]
//   - src/covered.ts              — IN tsconfig: no-unused-vars SHOULD fire
//   - test/skills.test.ts         — NOT IN tsconfig: no-unused-vars should NOT fire.
//                                    Contains a `console.log` so the marker rule triggers.
//   - template-nested/rslint.config.js — nested config, also projectService: true,
//                                    but the dir has NO tsconfig.json.
//   - template-nested/orphan.ts   — under the nested config. Both CLI and LSP
//                                    treat this as "has type info" (CLI via the
//                                    scan-directory fallback Program, LSP via
//                                    allow-all for the nested config's scope),
//                                    so no-unused-vars SHOULD fire.
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

    // Wait for `no-console` (non-type-aware, must fire on the fixture's
    // console.log) — its presence proves rslint has finalized this file's
    // diagnostics, so the negative assertion below can run synchronously
    // instead of waiting on a fixed-duration sleep.
    const diagnostics = await waitForDiagnostics(doc, (diags) =>
      rslintDiagnostics(diags).some((d) => d.message.includes('no-console')),
    );

    const rslintDiags = rslintDiagnostics(diagnostics);
    assert.ok(
      rslintDiags.some((d) => d.message.includes('no-console')),
      `Expected no-console marker to appear on test/skills.test.ts. Got: ${rslintDiags.map((d) => d.message).join(' | ')}`,
    );
    assert.ok(
      !rslintDiags.some((d) => d.message.includes('no-unused-vars')),
      `no-unused-vars should NOT fire on a file outside tsconfig.include. Got: ${rslintDiags.map((d) => d.message).join(' | ')}`,
    );
  });

  test('template-nested/orphan.ts (nested config without tsconfig) — no-unused-vars SHOULD fire, matching CLI', async () => {
    // Alignment check with CLI: when the nearest rslint config has no
    // resolvable tsconfig, the CLI builds a scan-directory AllowJs Program
    // and runs type-aware rules; the LSP falls through to allow-all scoped
    // to that config. Both engines must produce the same no-unused-vars
    // diagnostics on this file.
    const doc = await vscode.workspace.openTextDocument(
      path.join(workspaceRoot(), 'template-nested/orphan.ts'),
    );
    await vscode.window.showTextDocument(doc);

    const diagnostics = await waitForDiagnostics(doc, (diags) =>
      rslintDiagnostics(diags).some((d) =>
        d.message.includes('no-unused-vars'),
      ),
    );

    const rslintDiags = rslintDiagnostics(diagnostics);
    const unusedVarDiags = rslintDiags.filter((d) =>
      d.message.includes('no-unused-vars'),
    );
    assert.strictEqual(
      unusedVarDiags.length,
      3,
      `Expected 3 no-unused-vars diagnostics (command/args/options) to match CLI output. Got ${unusedVarDiags.length}: ${unusedVarDiags.map((d) => d.message).join(' | ')}`,
    );
  });
});

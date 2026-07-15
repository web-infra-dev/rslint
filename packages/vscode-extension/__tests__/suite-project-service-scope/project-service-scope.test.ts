import * as assert from 'assert';
import * as vscode from 'vscode';
import path from 'node:path';
import { findFixAllAction, requestFixAll } from '../suite/fixall-helpers';
import { waitForRslintDiagnostics as waitForDiagnostics } from '../utils/diagnostics';

// Type-aware rule scope when parserOptions uses `projectService: true`
// (the shape `ts.configs.recommended` exports) without an explicit
// `project`. The LSP and CLI must agree: only files covered by the
// fallback tsconfig's `include` get type-aware rules, AND a nested
// config that has no tsconfig must not enable type-aware rules.
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
//   - template-nested/orphan.ts   — under the nested config. no-unused-vars
//                                    should NOT fire, while native no-var should
//                                    diagnose and participate in fixAll.
suite('rslint projectService type-aware scope', function () {
  this.timeout(120000);

  function workspaceRoot(): string {
    return vscode.workspace.workspaceFolders![0].uri.fsPath;
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

  test('template-nested/orphan.ts (nested config without tsconfig) filters type-aware rules', async () => {
    const doc = await vscode.workspace.openTextDocument(
      path.join(workspaceRoot(), 'template-nested/orphan.ts'),
    );
    await vscode.window.showTextDocument(doc);

    const diagnostics = await waitForDiagnostics(doc, (diags) =>
      rslintDiagnostics(diags).some((d) => d.message.includes('no-var')),
    );

    const rslintDiags = rslintDiagnostics(diagnostics);
    assert.ok(
      rslintDiags.some((d) => d.message.includes('no-var')),
      `Expected native no-var marker. Got: ${rslintDiags.map((d) => d.message).join(' | ')}`,
    );
    assert.ok(
      !rslintDiags.some((d) => d.message.includes('no-unused-vars')),
      `no-unused-vars should not run without a resolved tsconfig. Got: ${rslintDiags.map((d) => d.message).join(' | ')}`,
    );
  });

  test('native fixAll remains available without a resolved tsconfig', async () => {
    const doc = await vscode.workspace.openTextDocument(
      path.join(workspaceRoot(), 'template-nested/orphan.ts'),
    );
    await vscode.window.showTextDocument(doc);
    await waitForDiagnostics(doc, (diags) =>
      rslintDiagnostics(diags).some((d) => d.message.includes('no-var')),
    );

    const fixAll = findFixAllAction(await requestFixAll(doc));
    const edits = fixAll?.edit?.get(doc.uri);
    assert.ok(edits && edits.length > 0, 'Expected a native fixAll edit');
    assert.ok(
      edits.some((edit) => edit.newText.includes('let output = command;')),
      `Expected no-var fix in fixAll edit. Got: ${edits.map((edit) => edit.newText).join(' | ')}`,
    );
  });
});

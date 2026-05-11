/**
 * VS Code e2e — single-config ESLint plugin smoke test.
 *
 * Workspace: __tests__/fixtures-eslint-plugin/
 * Contains a single `rslint.config.mjs` that registers a self-
 * contained fake plugin (`./plugin.mjs`) under prefix `fx`. The
 * plugin reports any `Identifier` named `forbidden`.
 *
 * This is the smallest end-to-end path:
 *   user edits → LSP → Rslint.ts.loadAndSendConfig →
 *   CompatPool.reconfigure → WorkerPool spawns → worker imports
 *   the user config → plugin runs → diagnostic published →
 *   vscode.languages.getDiagnostics returns it.
 *
 * If any link in the chain is broken (extension didn't see the
 * config, worker didn't get the right configKey, plugin wasn't
 * loaded, diagnostic format mismatch, etc.) this test fails.
 */
import * as assert from 'assert';
import * as vscode from 'vscode';
import path from 'node:path';
import fs from 'node:fs/promises';

suite('rslint single-config ESLint plugin support', function () {
  this.timeout(120_000);

  function getWorkspaceRoot(): string {
    return vscode.workspace.workspaceFolders![0].uri.fsPath;
  }

  // Race-free diagnostics waiter — copied verbatim from suite-jsconfig.
  // See that file for the rationale (subscribe BEFORE check + idempotent
  // finish to avoid the subscribe-after-publish window).
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

  test('plugin diagnostic from worker reaches vscode.languages.getDiagnostics', async () => {
    const filePath = path.join(getWorkspaceRoot(), 'src', 'index.ts');
    const doc = await vscode.workspace.openTextDocument(filePath);
    await vscode.window.showTextDocument(doc);

    // Wait for the plugin-emitted diagnostic specifically. The
    // rslint LSP server formats every rule diagnostic as
    // `[<ruleName>] <description>` (internal/lsp/service.go) and
    // sets source='rslint'; only a worker that imported the user
    // config, ran rule.create, and round-tripped the result back can
    // produce one with the `fx/no-forbidden` prefix.
    const diagnostics = await waitForDiagnostics(doc, (diags) =>
      diags.some(
        (d) => d.source === 'rslint' && d.message.includes('[fx/no-forbidden]'),
      ),
    );

    const hits = diagnostics.filter(
      (d) => d.source === 'rslint' && d.message.includes('[fx/no-forbidden]'),
    );
    // At least one Identifier(`forbidden`) must fire. The fixture
    // source `const forbidden = 1; ... export { forbidden, safe };`
    // produces multiple Identifier nodes for `forbidden` (declaration
    // + the ExportSpecifier's local + exported), and the rule fires
    // on each — that's expected ESTree visitor behavior, not a bug.
    // What we DO need to guard: every hit must flag `forbidden`, never
    // `safe`. A worker that confuses node identity would mis-flag.
    assert.ok(
      hits.length >= 1,
      `Expected ≥1 fx/no-forbidden diagnostic, got ${hits.length}`,
    );
    for (const hit of hits) {
      assert.strictEqual(
        doc.getText(hit.range),
        'forbidden',
        `Hit flagged unexpected text: ${JSON.stringify(doc.getText(hit.range))} at ${hit.range.start.line}:${hit.range.start.character}`,
      );
    }
  });

  // U9: didSave triggers an immediate plugin re-lint. didSave in
  // rslint LSP (service.go:432) calls pushDiagnostics directly
  // (bypassing the didChange debounce), so the cycle is observable
  // synchronously. The didChange debounced path is exercised by
  // existing native-rule tests in suite/extension.test.ts; this
  // test pins specifically that PLUGIN diagnostics participate in
  // the save → re-lint cycle (a regression here would mean files
  // saved while VS Code is showing plugin diagnostics never refresh
  // them, which is a common workflow when a user edits + Ctrl-S).
  test('U9: didSave refreshes plugin diagnostics', async () => {
    const filePath = path.join(getWorkspaceRoot(), 'src', 'index.ts');
    // Snapshot the original disk contents — we mutate the fixture
    // during the test (need a real save to trigger didSave) and
    // restore it on the way out so subsequent tests + git status
    // see a clean fixture.
    const originalDisk = await fs.readFile(filePath, 'utf8');

    const doc = await vscode.workspace.openTextDocument(filePath);
    const editor = await vscode.window.showTextDocument(doc);

    try {
      await waitForDiagnostics(doc, (diags) =>
        diags.some(
          (d) =>
            d.source === 'rslint' && d.message.includes('[fx/no-forbidden]'),
        ),
      );
      const initialCount = vscode.languages
        .getDiagnostics(doc.uri)
        .filter(
          (d) =>
            d.source === 'rslint' && d.message.includes('[fx/no-forbidden]'),
        ).length;

      // Append 3 new `forbidden` references then save. didSave forces
      // an immediate lint (no debounce).
      const tag = `t_${Date.now()}`;
      await editor.edit((eb) => {
        const endPos = doc.positionAt(doc.getText().length);
        eb.insert(
          endPos,
          `const ${tag}1 = forbidden;\nconst ${tag}2 = forbidden;\nconst ${tag}3 = forbidden;\n`,
        );
      });
      await doc.save();

      await waitForDiagnostics(doc, (diags) => {
        const hits = diags.filter(
          (d) =>
            d.source === 'rslint' && d.message.includes('[fx/no-forbidden]'),
        );
        return hits.length > initialCount;
      });
      const finalCount = vscode.languages
        .getDiagnostics(doc.uri)
        .filter(
          (d) =>
            d.source === 'rslint' && d.message.includes('[fx/no-forbidden]'),
        ).length;
      assert.ok(
        finalCount > initialCount,
        `expected no-forbidden hits to grow after save (initial=${initialCount}, final=${finalCount})`,
      );
    } finally {
      // Restore the fixture on disk. Editor's in-memory state may
      // still hold the mutated text; we write the disk back to its
      // pre-test contents so future tests + git see a clean fixture.
      await fs.writeFile(filePath, originalDisk, 'utf8');
    }
  });

  // U10: source.fixAll code action runs every plugin-provided fix.
  // `fx/rename-banned` (defined in ../fixtures-eslint-plugin/plugin.mjs)
  // ships an autofix that rewrites `BANNED` → `ALLOWED`. The test
  // opens the fixture file, waits for the rule to fire, then triggers
  // the `source.fixAll` code action (the same one Quick Fix's
  // "Fix all" or VS Code's `editor.codeActionsOnSave` invokes) and
  // asserts the source content reflects the rewrite.
  test('U10: source.fixAll applies plugin-supplied autofix to document', async () => {
    const filePath = path.join(getWorkspaceRoot(), 'src', 'index.ts');
    const doc = await vscode.workspace.openTextDocument(filePath);
    await vscode.window.showTextDocument(doc);

    // Wait for the rename-banned diagnostic to land.
    await waitForDiagnostics(doc, (diags) =>
      diags.some(
        (d) =>
          d.source === 'rslint' && d.message.includes('[fx/rename-banned]'),
      ),
    );

    // Trigger the source.fixAll code action over the entire document.
    const range = new vscode.Range(
      new vscode.Position(0, 0),
      doc.lineAt(doc.lineCount - 1).range.end,
    );
    const actions =
      (await vscode.commands.executeCommand<vscode.CodeAction[]>(
        'vscode.executeCodeActionProvider',
        doc.uri,
        range,
        vscode.CodeActionKind.SourceFixAll.value,
      )) ?? [];

    // Expect at least one fixAll-class action that rewrites BANNED.
    // Apply the first matching one and verify the document text.
    let applied = false;
    for (const action of actions) {
      if (
        action.kind &&
        action.kind.intersects(vscode.CodeActionKind.SourceFixAll) &&
        action.edit
      ) {
        const ok = await vscode.workspace.applyEdit(action.edit);
        if (ok) {
          applied = true;
          break;
        }
      }
    }

    // If no SourceFixAll action surfaced, the fix-on-save / fixAll
    // path doesn't reach plugin fixes — that's the regression this
    // test catches.
    assert.ok(
      applied,
      `expected a source.fixAll action to apply a plugin fix; got actions=${actions
        .map((a) => `${a.title}(${a.kind?.value ?? 'no-kind'})`)
        .join(', ')}`,
    );

    // After the fix, the document text must contain ALLOWED and not
    // contain the original BANNED identifier outside of a comment.
    const after = doc.getText();
    assert.ok(after.includes('ALLOWED'), 'expected ALLOWED in document');
    // The "BANNED" string may still appear in our doc comment header
    // referring to the rule by name — strict assertion checks the
    // identifier declaration site (`const BANNED`).
    assert.ok(
      !/const BANNED\b/.test(after),
      'expected `const BANNED` to be gone after fixAll',
    );
  });
});

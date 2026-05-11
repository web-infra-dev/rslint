/**
 * VS Code e2e — multi-config ESLint plugin routing.
 *
 * Workspace: __tests__/fixtures-eslint-plugin-monorepo/
 *   - root rslint.config.mjs    (no plugins, base TS settings)
 *   - packages/x/rslint.config.mjs  imports ./plugin-x.mjs (prefix `px`)
 *   - packages/y/rslint.config.mjs  imports ./plugin-y.mjs (prefix `py`)
 *
 * The two nested configs each have their OWN plugin module — different
 * URL, different prefix, different rule names. This mirrors what a
 * monorepo user does when different packages depend on different
 * ESLint plugin sets.
 *
 * What this proves end-to-end in a real vscode + LSP session:
 *
 *   1. The CompatPool inside the extension reconfigures with BOTH
 *      configs after `loadAndSendConfig` runs.
 *   2. The LSP server's compat dispatcher picks the right configKey
 *      per file (resolved nearest rslint.config), and the worker's
 *      per-config `loadedPluginsByDir` map returns the matching
 *      LoadedPlugins set.
 *   3. A file under packages/x receives only `px/no-foo` diagnostics
 *      — the worker's `resolveRule` against cfgX's LoadedPlugins
 *      returns null for `py/no-bar` (plugin Y was never imported by
 *      cfgX), so it cannot fire there. Symmetrically for packages/y.
 *
 * Worker-pool-e2e covers the same invariants in isolation via direct
 * `pool.lintBatch` calls. This e2e instead exercises the full chain:
 * Rslint.ts → CompatPool → WorkerPool → plugin → LSP publish.
 */
import * as assert from 'assert';
import * as vscode from 'vscode';
import path from 'node:path';

suite('rslint multi-config ESLint plugin routing', function () {
  this.timeout(120_000);

  function getWorkspaceRoot(): string {
    return vscode.workspace.workspaceFolders![0].uri.fsPath;
  }

  // Race-free diagnostics waiter — see suite-jsconfig for rationale.
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

  async function openPackageFile(pkg: 'x' | 'y'): Promise<vscode.TextDocument> {
    const filePath = path.join(
      getWorkspaceRoot(),
      'packages',
      pkg,
      'src',
      'index.ts',
    );
    const doc = await vscode.workspace.openTextDocument(filePath);
    await vscode.window.showTextDocument(doc);
    return doc;
  }

  // rslint LSP formats every rule diagnostic's message as
  // `[<ruleName>] <description>` (internal/lsp/service.go). We match
  // on that prefix — the `code` field is currently unset by the
  // server, but the message is reliable.
  function isHit(d: vscode.Diagnostic, rule: string): boolean {
    return d.source === 'rslint' && d.message.includes(`[${rule}]`);
  }

  test('packages/x file is routed to plugin X (px/no-foo fires on `foo`)', async () => {
    const doc = await openPackageFile('x');

    const diagnostics = await waitForDiagnostics(doc, (diags) =>
      diags.some((d) => isHit(d, 'px/no-foo')),
    );

    // The fixture source declares + exports `foo`, so ESTree produces
    // multiple Identifier(`foo`) visits (declaration + ExportSpecifier
    // local + exported). The rule fires on each — expected, not a
    // bug. The invariants we DO want to enforce:
    //   1. ≥1 hit (the worker actually ran the rule)
    //   2. every hit flags `foo` (no misattribution)
    //   3. zero `py/no-bar` hits (plugin Y did NOT leak from cfgY)
    const fooHits = diagnostics.filter((d) => isHit(d, 'px/no-foo'));
    assert.ok(
      fooHits.length >= 1,
      `Expected ≥1 px/no-foo in packages/x, got ${fooHits.length}`,
    );
    for (const h of fooHits) {
      assert.strictEqual(doc.getText(h.range), 'foo');
    }

    // End-to-end no-leak invariant: packages/x in vscode never shows
    // a py/no-bar diagnostic, even though plugin Y is loaded by the
    // sibling cfgY. NOTE: this assertion is partly trivially satisfied
    // by Go-side `files`-glob dispatch (the Go LSP server doesn't send
    // py/no-bar in the rules dict for a packages/x file, since cfgY's
    // files glob doesn't match). The harder invariant — that the
    // worker's per-config `loadedPluginsByDir` truly keeps plugins
    // disjoint even when both configs' rules ARE requested — is
    // covered separately by `multi-config dispatch with disjoint
    // plugin sets stays isolated` in worker-pool-e2e.test.ts:458,
    // which forces a cross-config rule request and asserts
    // resolveRule returns null → ruleErrors fires. The two tests
    // together cover both the end-to-end user experience and the
    // worker-internal isolation contract.
    const barLeaks = diagnostics.filter((d) => isHit(d, 'py/no-bar'));
    assert.strictEqual(
      barLeaks.length,
      0,
      `plugin Y leaked into packages/x: ${JSON.stringify(barLeaks.map((d) => d.message))}`,
    );
  });

  test('packages/y file is routed to plugin Y (py/no-bar fires on `bar`)', async () => {
    const doc = await openPackageFile('y');

    const diagnostics = await waitForDiagnostics(doc, (diags) =>
      diags.some((d) => isHit(d, 'py/no-bar')),
    );

    const barHits = diagnostics.filter((d) => isHit(d, 'py/no-bar'));
    assert.ok(
      barHits.length >= 1,
      `Expected ≥1 py/no-bar in packages/y, got ${barHits.length}`,
    );
    for (const h of barHits) {
      assert.strictEqual(doc.getText(h.range), 'bar');
    }

    // Symmetric end-to-end no-leak; see the comment on barLeaks in
    // the packages/x test for why worker-pool-e2e covers the stricter
    // worker-internal isolation invariant.
    const fooLeaks = diagnostics.filter((d) => isHit(d, 'px/no-foo'));
    assert.strictEqual(
      fooLeaks.length,
      0,
      `plugin X leaked into packages/y: ${JSON.stringify(fooLeaks.map((d) => d.message))}`,
    );
  });
});

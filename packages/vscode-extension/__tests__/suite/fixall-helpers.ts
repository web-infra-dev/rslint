import * as assert from 'assert';
import * as vscode from 'vscode';
import path from 'node:path';
import fs from 'node:fs';

export function getFixturesDir(): string {
  return path.resolve(require.resolve('@rslint/core'), '../..', 'fixtures');
}

export async function openFixture(
  filename: string,
): Promise<vscode.TextDocument> {
  return vscode.workspace.openTextDocument(
    path.resolve(getFixturesDir(), 'src/', filename),
  );
}

export async function waitForDiagnostics(
  doc: vscode.TextDocument,
): Promise<vscode.Diagnostic[]> {
  // On CI (especially Windows), the LSP server may take longer to start up,
  // load config, type-check, and push initial diagnostics. Use generous
  // iteration count (15) and per-iteration timeout (2s) to avoid flaky failures.
  for (let i = 0; i < 15; i++) {
    const diagnostics = vscode.languages.getDiagnostics(doc.uri);
    if (diagnostics.length > 0) {
      return diagnostics;
    }
    await new Promise((resolve) => {
      const disposable = vscode.languages.onDidChangeDiagnostics((e) => {
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
      }, 2000);
    });
  }
  return vscode.languages.getDiagnostics(doc.uri);
}

export async function waitForDiagnosticsToChange(
  doc: vscode.TextDocument,
  previousCount: number,
  timeoutMs = 30000,
): Promise<vscode.Diagnostic[]> {
  const startTime = Date.now();
  while (Date.now() - startTime < timeoutMs) {
    const current = vscode.languages.getDiagnostics(doc.uri);
    if (current.length !== previousCount) {
      return current;
    }
    await new Promise((resolve) => {
      const disposable = vscode.languages.onDidChangeDiagnostics((e) => {
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
      }, 500);
    });
  }
  return vscode.languages.getDiagnostics(doc.uri);
}

export async function waitForDiagnosticsCount(
  doc: vscode.TextDocument,
  expectedCount: number,
  timeoutMs = 30000,
): Promise<vscode.Diagnostic[]> {
  const startTime = Date.now();
  while (Date.now() - startTime < timeoutMs) {
    const current = vscode.languages.getDiagnostics(doc.uri);
    if (current.length === expectedCount) {
      return current;
    }
    await new Promise((resolve) => {
      const disposable = vscode.languages.onDidChangeDiagnostics((e) => {
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
      }, 500);
    });
  }
  return vscode.languages.getDiagnostics(doc.uri);
}

export function findFixAllAction(
  codeActions: vscode.CodeAction[] | undefined,
): vscode.CodeAction | undefined {
  return codeActions?.find(
    (action) =>
      action.kind?.value === 'source.fixAll.rslint' ||
      action.kind?.value === 'source.fixAll',
  );
}

/**
 * Wraps `vscode.commands.executeCommand('vscode.executeCodeActionProvider', ...)`
 * with retry-on-cancellation. The Code-Action provider's command receives an
 * ambient cancellation token under the hood; on a freshly-started extension
 * host (CI cold start) external events such as Settings Sync state
 * transitions or async ConfigurationService initialization can fire that
 * token mid-call, surfacing as a synthetic `Canceled` error whose stack is
 * entirely inside `extensionHostProcess.js`. The call is otherwise a pure
 * read of the diagnostic state, so a bounded retry with linear backoff is
 * safe and recovers transparently.
 *
 * Only `Canceled` / `Cancellation` errors are retried — any other failure
 * propagates immediately so genuine bugs surface fast.
 */
export async function executeCodeActionProviderWithRetry(
  uri: vscode.Uri,
  range: vscode.Range,
  kind?: string,
  retries = 3,
): Promise<vscode.CodeAction[]> {
  for (let attempt = 0; attempt < retries; attempt++) {
    try {
      const args: unknown[] = ['vscode.executeCodeActionProvider', uri, range];
      if (kind !== undefined) args.push(kind);
      const result = await vscode.commands.executeCommand<vscode.CodeAction[]>(
        ...(args as [string, vscode.Uri, vscode.Range, ...unknown[]]),
      );
      return result ?? [];
    } catch (err) {
      const isLast = attempt === retries - 1;
      const message = err instanceof Error ? err.message : String(err);
      const isCancellation = /cancel/i.test(message);
      if (isLast || !isCancellation) throw err;
      // Linear backoff: 200ms, 400ms, ...
      await new Promise((r) => setTimeout(r, 200 * (attempt + 1)));
    }
  }
  // Unreachable: the loop either returns or throws.
  return [];
}

export async function requestFixAll(
  doc: vscode.TextDocument,
  kind: vscode.CodeActionKind = vscode.CodeActionKind.SourceFixAll.append(
    'rslint',
  ),
): Promise<vscode.CodeAction[]> {
  return executeCodeActionProviderWithRetry(
    doc.uri,
    new vscode.Range(0, 0, doc.lineCount, 0),
    kind.value,
  );
}

export async function withTmpFile(
  content: string,
  testFn: (
    doc: vscode.TextDocument,
    editor: vscode.TextEditor,
  ) => Promise<void>,
): Promise<void> {
  const tmpFile = path.join(
    getFixturesDir(),
    'src',
    `_fixall_tmp_${Date.now()}_${Math.random().toString(36).slice(2, 8)}.ts`,
  );
  fs.writeFileSync(tmpFile, content, 'utf-8');
  try {
    const doc = await vscode.workspace.openTextDocument(tmpFile);
    const editor = await vscode.window.showTextDocument(doc);
    await testFn(doc, editor);
  } finally {
    // Close the editor tab so VSCode sends a synchronous didClose to the LSP,
    // which cleans up the session overlay. Without this, the file deletion
    // triggers an async didClose via the file watcher, which can race with
    // the next test's LSP requests (all blocking methods are serialized).
    await vscode.commands.executeCommand('workbench.action.closeActiveEditor');
    if (fs.existsSync(tmpFile)) {
      fs.unlinkSync(tmpFile);
    }
  }
}

/**
 * Wait until `predicate(content)` becomes true. Subscribes to
 * `vscode.workspace.onDidChangeTextDocument` and returns the moment a
 * matching content arrives, instead of polling on a fixed interval.
 *
 * Use this in preference to a `while (...) await sleep(500)` loop when the
 * test is waiting for a server-driven content change (e.g. on-save fixAll
 * applying an edit): the event-driven path resolves with sub-millisecond
 * latency once the change lands, so the only wall-clock cost is the
 * server's actual response time.
 *
 * Rejects with a descriptive error (including the last seen content) when
 * `timeoutMs` elapses without the predicate being satisfied.
 */
export async function waitForContentChange(
  doc: vscode.TextDocument,
  predicate: (content: string) => boolean,
  timeoutMs: number,
): Promise<string> {
  const initial = doc.getText();
  if (predicate(initial)) return initial;
  return new Promise<string>((resolve, reject) => {
    const docUriString = doc.uri.toString();
    let timer: ReturnType<typeof setTimeout> | undefined;
    const disposable = vscode.workspace.onDidChangeTextDocument((e) => {
      if (e.document.uri.toString() !== docUriString) return;
      const current = doc.getText();
      if (predicate(current)) {
        if (timer) clearTimeout(timer);
        disposable.dispose();
        resolve(current);
      }
    });
    timer = setTimeout(() => {
      disposable.dispose();
      reject(
        new Error(
          `waitForContentChange: predicate not satisfied within ${timeoutMs}ms. Last content:\n${doc.getText()}`,
        ),
      );
    }, timeoutMs);
  });
}

export async function replaceAll(
  editor: vscode.TextEditor,
  newContent: string,
): Promise<void> {
  const doc = editor.document;
  const fullRange = new vscode.Range(
    doc.positionAt(0),
    doc.positionAt(doc.getText().length),
  );
  const ok = await editor.edit((b) => b.replace(fullRange, newContent));
  assert.ok(ok, 'editor.edit should succeed');
}

/**
 * Run the on-save fixAll pipeline once with a tiny ban-types fixture so
 * subsequent tests that rely on `editor.codeActionsOnSave: { 'source.fixAll.rslint': 'explicit' }`
 * skip the first-trigger cold-start cost (VS Code wiring up the on-save
 * code-actions handler, the LSP `textDocument/codeAction` round-trip, the
 * workspace edit application). Best-effort: if ban-types doesn't fire (rule
 * disabled / config not loaded yet) we silently no-op rather than failing
 * the suite — actual coverage lives in the real tests.
 *
 * Call from a suite-level `before()` of any suite whose first test
 * exercises on-save fixAll. Idempotent across the entire test process via
 * a module-level promise cache: the first call kicks off the warm-up; all
 * subsequent calls (from any suite) await the same promise and return
 * immediately once it resolves. Safe to wire up in every suite that uses
 * `withOnSaveFixAll` without paying duplicate cost.
 */
let prewarmPromise: Promise<void> | undefined;
export function prewarmOnSaveFixAll(): Promise<void> {
  if (prewarmPromise) return prewarmPromise;
  prewarmPromise = withOnSaveFixAll(async (doc, editor) => {
    await replaceAll(
      editor,
      "const _prewarmOnSave: String = 'x';\nexport { _prewarmOnSave };\n",
    );
    const diags = await waitForDiagnostics(doc);
    if (!diags.some((d) => d.message.includes('ban-types'))) return;
    await doc.save();
    try {
      await waitForContentChange(
        doc,
        (content) => !content.includes(': String'),
        60000,
      );
    } catch {
      // Pre-warm is best-effort: if the fixAll didn't apply within 60s the
      // actual test will surface a clearer assertion error.
    }
  });
  return prewarmPromise;
}

export async function withOnSaveFixAll(
  testFn: (
    doc: vscode.TextDocument,
    editor: vscode.TextEditor,
  ) => Promise<void>,
): Promise<void> {
  const fixturesDir = getFixturesDir();
  const tmpFile = path.join(
    fixturesDir,
    'src',
    `_fixall_test_${Date.now()}.ts`,
  );
  fs.writeFileSync(tmpFile, '// placeholder\n', 'utf-8');

  try {
    const config = vscode.workspace.getConfiguration('editor');
    const previousValue = config.get('codeActionsOnSave');
    await config.update(
      'codeActionsOnSave',
      { 'source.fixAll.rslint': 'explicit' },
      vscode.ConfigurationTarget.Workspace,
    );

    try {
      const doc = await vscode.workspace.openTextDocument(tmpFile);
      const editor = await vscode.window.showTextDocument(doc);
      await testFn(doc, editor);
    } finally {
      await config.update(
        'codeActionsOnSave',
        previousValue,
        vscode.ConfigurationTarget.Workspace,
      );
    }
  } finally {
    // Close the editor tab so VSCode sends a synchronous didClose to the LSP,
    // which cleans up the session overlay. Without this, the file deletion
    // triggers an async didClose via the file watcher, which can race with
    // the next test's LSP requests (all blocking methods are serialized).
    await vscode.commands.executeCommand('workbench.action.closeActiveEditor');
    if (fs.existsSync(tmpFile)) {
      fs.unlinkSync(tmpFile);
    }
  }
}

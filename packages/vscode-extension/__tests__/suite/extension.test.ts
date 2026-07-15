import * as assert from 'assert';
import * as vscode from 'vscode';
import path from 'node:path';
import { executeCodeActionProvider, getFixturesDir } from './fixall-helpers';
import {
  getRslintDiagnostics,
  waitForRslintDiagnostics as waitForDiagnostics,
  waitForRslintDiagnosticsCount as waitForDiagnosticsCount,
  waitForRslintDiagnosticsToChange as waitForDiagnosticsToChange,
} from '../utils/diagnostics';
import { closeTextEditor, revertTextDocument } from '../utils/documents';

suite('rslint extension', function () {
  this.timeout(90000);

  teardown(async () => {
    const fixturesSource = path.resolve(getFixturesDir(), 'src');
    const dirtyFixtures = vscode.workspace.textDocuments.filter((document) => {
      if (document.uri.scheme !== 'file') return false;
      const relative = path.relative(fixturesSource, document.uri.fsPath);
      return (
        document.isDirty &&
        relative !== '' &&
        !relative.startsWith(`..${path.sep}`) &&
        !path.isAbsolute(relative)
      );
    });
    for (const document of dirtyFixtures) {
      await revertTextDocument(document);
    }
  });

  function waitForDiagnosticsWithMessage(
    doc: vscode.TextDocument,
    messageSubstring: string,
    timeoutMs = 30000,
  ): Promise<vscode.Diagnostic[]> {
    return waitForDiagnostics(
      doc,
      (diagnostics) =>
        diagnostics.some((diagnostic) =>
          diagnostic.message.includes(messageSubstring),
        ),
      timeoutMs,
    );
  }

  // Helper function to open a test fixture
  async function openFixture(filename: string): Promise<vscode.TextDocument> {
    return vscode.workspace.openTextDocument(
      path.resolve(getFixturesDir(), 'src', filename),
    );
  }

  test('diagnostics', async () => {
    const doc = await openFixture('index.ts');
    await vscode.window.showTextDocument(doc);

    const diagnostics = await waitForDiagnostics(doc);
    assert.ok(
      diagnostics.length > 0,
      `Expected diagnostics but got ${diagnostics.length}`,
    );
  });

  test('.gitignore excludes diagnostics for an opened file', async () => {
    const doc = await openFixture('gitignored.ts');
    await vscode.window.showTextDocument(doc);

    const control = await openFixture('disable.ts');
    await vscode.window.showTextDocument(control);
    const controlDiagnostics = await waitForDiagnosticsWithMessage(
      control,
      'no-unsafe-member-access',
    );
    assert.ok(
      controlDiagnostics.some(
        (diagnostic) =>
          diagnostic.source === 'rslint' &&
          diagnostic.message.includes('no-unsafe-member-access'),
      ),
      'Expected the unignored control file to produce an rslint diagnostic',
    );

    const diagnostics = vscode.languages
      .getDiagnostics(doc.uri)
      .filter((diagnostic) => diagnostic.source === 'rslint');
    assert.strictEqual(
      diagnostics.length,
      0,
      `Expected no diagnostics for a gitignored file, got: ${diagnostics
        .map((diagnostic) => diagnostic.message)
        .join(', ')}`,
    );
  });

  test('code actions - auto fix', async () => {
    const doc = await openFixture('autofix.ts');
    await vscode.window.showTextDocument(doc);

    const diagnostics = await waitForDiagnostics(doc);
    assert.ok(diagnostics.length > 0, 'Should have diagnostics');

    // Find the no-unnecessary-type-assertion diagnostic
    const typeAssertionDiag = diagnostics.find(
      (d) =>
        d.message.includes('no-unnecessary-type-assertion') ||
        (d.source === 'rslint' && d.message.includes('assertion')),
    );
    assert.ok(
      typeAssertionDiag,
      `Expected a no-unnecessary-type-assertion diagnostic. Got: ${diagnostics
        .map((diagnostic) => diagnostic.message)
        .join(' | ')}`,
    );

    // Request code actions for the diagnostic range
    const codeActions = await executeCodeActionProvider(
      doc.uri,
      typeAssertionDiag.range,
    );

    assert.ok(codeActions.length > 0, 'Should have code actions');

    // Look for auto fix action
    const autoFixAction = codeActions.find(
      (action) =>
        action.title.toLowerCase().includes('fix') &&
        action.kind?.value === vscode.CodeActionKind.QuickFix.value,
    );

    assert.ok(autoFixAction, 'Should have auto fix action');
    assert.ok(
      autoFixAction.isPreferred,
      'Auto fix should be marked as preferred',
    );

    // Verify the action has an edit
    assert.ok(autoFixAction.edit, 'Auto fix action should have an edit');
    const autoFixEdits = autoFixAction.edit.get(doc.uri);
    assert.ok(
      autoFixEdits && autoFixEdits.length > 0,
      'Auto fix edit should not be empty',
    );
  });

  test('code actions - disable rule for line', async () => {
    const doc = await openFixture('disable.ts');
    await vscode.window.showTextDocument(doc);

    const diagnostics = await waitForDiagnostics(doc);
    assert.ok(diagnostics.length > 0, 'Should have diagnostics');

    // Find an unsafe diagnostic (these typically don't have auto fixes)
    const unsafeDiag = diagnostics.find(
      (d) => d.message.includes('unsafe') || d.message.includes('Unsafe'),
    );
    assert.ok(
      unsafeDiag,
      `Expected an unsafe diagnostic. Got: ${diagnostics
        .map((diagnostic) => diagnostic.message)
        .join(' | ')}`,
    );

    // Request code actions for the diagnostic range
    const codeActions = await executeCodeActionProvider(
      doc.uri,
      unsafeDiag.range,
    );

    assert.ok(codeActions.length > 0, 'Should have code actions');

    // Look for disable rule for line action
    const disableLineAction = codeActions.find(
      (action) =>
        action.title.toLowerCase().includes('disable') &&
        action.title.toLowerCase().includes('line'),
    );

    assert.ok(disableLineAction, 'Should have disable rule for line action');
    assert.ok(
      !disableLineAction.isPreferred,
      'Disable action should not be marked as preferred',
    );

    // Verify the action has an edit
    assert.ok(disableLineAction.edit, 'Disable action should have an edit');

    // Verify the edit contains rslint-disable-next-line
    const workspaceEdit = disableLineAction.edit;
    const edits = workspaceEdit.get(doc.uri);
    assert.ok(
      edits && edits.length > 0,
      'Disable-line edit should not be empty',
    );
    const editText = edits[0].newText;
    assert.ok(
      editText.includes('rslint-disable-next-line'),
      'Edit should contain rslint-disable-next-line comment',
    );
  });

  test('code actions - disable rule for file', async () => {
    const doc = await openFixture('disable-file.ts');
    await vscode.window.showTextDocument(doc);

    const diagnostics = await waitForDiagnostics(doc);
    assert.ok(diagnostics.length > 0, 'Should have diagnostics');

    // Find an unsafe diagnostic
    const unsafeDiag = diagnostics.find(
      (d) => d.message.includes('unsafe') || d.message.includes('Unsafe'),
    );
    assert.ok(
      unsafeDiag,
      `Expected an unsafe diagnostic. Got: ${diagnostics
        .map((diagnostic) => diagnostic.message)
        .join(' | ')}`,
    );

    // Request code actions for the diagnostic range
    const codeActions = await executeCodeActionProvider(
      doc.uri,
      unsafeDiag.range,
    );

    assert.ok(codeActions.length > 0, 'Should have code actions');

    // Look for disable rule for file action
    const disableFileAction = codeActions.find(
      (action) =>
        action.title.toLowerCase().includes('disable') &&
        action.title.toLowerCase().includes('file'),
    );

    assert.ok(disableFileAction, 'Should have disable rule for file action');
    assert.ok(
      !disableFileAction.isPreferred,
      'Disable action should not be marked as preferred',
    );

    // Verify the action has an edit
    assert.ok(disableFileAction.edit, 'Disable action should have an edit');

    // Verify the edit contains rslint-disable comment
    const workspaceEdit = disableFileAction.edit;
    const edits = workspaceEdit.get(doc.uri);
    assert.ok(
      edits && edits.length > 0,
      'Disable-file edit should not be empty',
    );
    const editText = edits[0].newText;
    assert.ok(
      editText.includes('rslint-disable') && !editText.includes('-next-line'),
      'Edit should contain rslint-disable comment for entire file',
    );
  });

  test('code actions - range overlap', async () => {
    const doc = await openFixture('index.ts');
    await vscode.window.showTextDocument(doc);

    await waitForDiagnostics(doc);

    // Test that code actions are only provided for ranges that overlap with diagnostics
    const codeActionsEmptyRange = await executeCodeActionProvider(
      doc.uri,
      new vscode.Range(100, 0, 100, 0), // Range with no diagnostics
    );

    // Should either be empty or only contain general actions (not diagnostic-specific)
    if (codeActionsEmptyRange) {
      const diagnosticSpecificActions = codeActionsEmptyRange.filter(
        (action) =>
          action.title.toLowerCase().includes('fix') ||
          action.title.toLowerCase().includes('disable'),
      );
      assert.strictEqual(
        diagnosticSpecificActions.length,
        0,
        'Should not have diagnostic-specific actions for empty range',
      );
    }
  });

  test('diagnostics refresh after edit - removing errors', async () => {
    const doc = await openFixture('autofix.ts');
    const editor = await vscode.window.showTextDocument(doc);

    // 1. Wait for initial diagnostics
    const initialDiags = await waitForDiagnostics(doc);
    assert.ok(
      initialDiags.length > 0,
      `Expected initial diagnostics but got ${initialDiags.length}`,
    );
    const initialCount = initialDiags.length;

    // 2. Replace file content with error-free code
    const fullRange = new vscode.Range(
      doc.positionAt(0),
      doc.positionAt(doc.getText().length),
    );
    await editor.edit((editBuilder) => {
      editBuilder.replace(fullRange, '// no lint errors\nexport {};\n');
    });

    // 3. Wait for the exact final state; accepting the first smaller
    // intermediate publication could hide diagnostics that never clear.
    const updatedDiags = await waitForDiagnosticsCount(doc, 0);

    assert.strictEqual(
      updatedDiags.length,
      0,
      `Expected zero diagnostics after removing errors. Before: ${initialCount}, After: ${updatedDiags.length}`,
    );
  });

  test('diagnostics refresh after edit - introducing errors', async () => {
    const doc = await openFixture('autofix.ts');
    const editor = await vscode.window.showTextDocument(doc);

    // 1. Wait for initial diagnostics
    const initialDiags = await waitForDiagnostics(doc);
    const initialCount = initialDiags.length;

    // 2. Append code that introduces additional lint errors
    const endPos = doc.positionAt(doc.getText().length);
    await editor.edit((editBuilder) => {
      editBuilder.insert(
        endPos,
        '\nconst anyVal: any = 123;\nanyVal.foo = 1;\n',
      );
    });

    // 3. Wait for diagnostics to update (should increase)
    const updatedDiags = await waitForDiagnosticsToChange(doc, initialCount);

    assert.ok(
      updatedDiags.length > initialCount,
      `Expected more diagnostics after introducing errors. Before: ${initialCount}, After: ${updatedDiags.length}`,
    );
  });

  test('diagnostics refresh after rapid successive edits', async () => {
    const doc = await openFixture('autofix.ts');
    const editor = await vscode.window.showTextDocument(doc);

    // 1. Wait for initial diagnostics
    const initialDiags = await waitForDiagnostics(doc);
    assert.ok(initialDiags.length > 0, 'Should have initial diagnostics');
    const initialCount = initialDiags.length;

    // 2. Perform rapid successive edits — simulates fast typing
    //    Each edit replaces the full content. The server should debounce
    //    and only produce diagnostics for the final state.
    const fullRange = () =>
      new vscode.Range(doc.positionAt(0), doc.positionAt(doc.getText().length));

    // Edit 1: still has errors
    await editor.edit((b) =>
      b.replace(fullRange(), 'const x: any = 1;\nexport {};\n'),
    );
    // Edit 2: still has errors
    await editor.edit((b) =>
      b.replace(
        fullRange(),
        'const y: any = 2;\nconst z: any = 3;\nexport {};\n',
      ),
    );
    // Edit 3: error-free — the final state that matters
    await editor.edit((b) =>
      b.replace(fullRange(), '// all clean\nexport {};\n'),
    );

    // 3. Wait for diagnostics to settle — should reflect the error-free final state
    const finalDiags = await waitForDiagnosticsCount(doc, 0);

    assert.strictEqual(
      finalDiags.length,
      0,
      `After rapid edits ending with clean code, expected zero diagnostics. Before: ${initialCount}, After: ${finalDiags.length}`,
    );
  });

  test('diagnostics clear completely when all errors removed', async () => {
    const doc = await openFixture('index.ts');
    const editor = await vscode.window.showTextDocument(doc);

    // 1. Wait for initial diagnostics
    const initialDiags = await waitForDiagnostics(doc);
    assert.ok(initialDiags.length > 0, 'Should have initial diagnostics');

    // 2. Replace with completely clean content
    const fullRange = new vscode.Range(
      doc.positionAt(0),
      doc.positionAt(doc.getText().length),
    );
    await editor.edit((b) => {
      b.replace(fullRange, '// empty file\nexport {};\n');
    });

    // 3. Wait for diagnostics to reach zero — use waitForDiagnosticsCount
    // to avoid catching intermediate states from debounce
    const finalDiags = await waitForDiagnosticsCount(doc, 0);

    assert.strictEqual(
      finalDiags.length,
      0,
      `Expected zero diagnostics after clearing all errors, got ${finalDiags.length}`,
    );
  });

  test('diagnostics update across multiple edit cycles', async () => {
    const doc = await openFixture('disable-file.ts');
    const editor = await vscode.window.showTextDocument(doc);

    const fullRange = () =>
      new vscode.Range(doc.positionAt(0), doc.positionAt(doc.getText().length));

    // Start from a published non-empty snapshot, so the following zero cannot
    // be the document's not-yet-linted initial state.
    await waitForDiagnostics(doc);

    // Step 1: start from clean state to establish baseline
    await editor.edit((b) =>
      b.replace(fullRange(), '// no errors\nexport {};\n'),
    );
    await waitForDiagnosticsCount(doc, 0, 10000);
    const cleanCount = getRslintDiagnostics(doc).length;

    // Step 2: introduce errors
    await editor.edit((b) =>
      b.replace(
        fullRange(),
        'const x: any = 1;\nconst y: any = 2;\nx.foo;\ny.bar;\nexport {};\n',
      ),
    );
    const diags2 = await waitForDiagnosticsToChange(doc, cleanCount);
    assert.ok(
      diags2.length > cleanCount,
      `After introducing errors: expected more diagnostics. Before: ${cleanCount}, After: ${diags2.length}`,
    );
    const errorCount = diags2.length;

    // Step 3: clear errors again — use waitForDiagnosticsCount to avoid
    // catching intermediate states from debounce on CI
    await editor.edit((b) =>
      b.replace(fullRange(), '// clean again\nexport {};\n'),
    );
    const diags3 = await waitForDiagnosticsCount(doc, 0);
    assert.ok(
      diags3.length < errorCount,
      `After clearing errors: expected fewer diagnostics. Before: ${errorCount}, After: ${diags3.length}`,
    );
  });

  test('diagnostics transition: clean → error A → error B → clean', async () => {
    const doc = await openFixture('error-transitions.ts');
    const editor = await vscode.window.showTextDocument(doc);

    const fullRange = () =>
      new vscode.Range(doc.positionAt(0), doc.positionAt(doc.getText().length));

    // Establish a published non-empty baseline first, so waiting for zero
    // cannot pass on the clean document's not-yet-linted initial state.
    await editor.edit((b) =>
      b.replace(
        fullRange(),
        'const baseline: any = {};\nbaseline.member;\nexport {};\n',
      ),
    );
    await waitForDiagnosticsWithMessage(doc, 'no-unsafe-member-access');

    // Step 1: Start with clean code — should have zero diagnostics
    await editor.edit((b) =>
      b.replace(fullRange(), '// no errors\nexport {};\n'),
    );
    const cleanDiags = await waitForDiagnosticsCount(doc, 0, 30_000);
    assert.strictEqual(
      cleanDiags.length,
      0,
      `Step 1 (clean): expected 0 diagnostics, got ${cleanDiags.length}`,
    );

    // Step 2: Introduce error A — no-unsafe-member-access
    await editor.edit((b) =>
      b.replace(
        fullRange(),
        'const obj: any = {};\nobj.foo.bar;\nexport {};\n',
      ),
    );
    const errorADiags = await waitForDiagnosticsWithMessage(
      doc,
      'no-unsafe-member-access',
    );
    assert.ok(
      errorADiags.length > 0,
      `Step 2 (error A): expected diagnostics, got ${errorADiags.length}`,
    );
    assert.ok(
      errorADiags.some((d) => d.message.includes('no-unsafe-member-access')),
      `Step 2 (error A): expected no-unsafe-member-access diagnostic, got: ${errorADiags.map((d) => d.message).join(', ')}`,
    );

    // Step 3: Change to error B — no-unnecessary-type-assertion (different rule)
    await editor.edit((b) =>
      b.replace(
        fullRange(),
        "const someValue: string = 'hello';\nconst result = someValue as string;\nexport {};\n",
      ),
    );
    const errorBDiags = await waitForDiagnosticsWithMessage(
      doc,
      'no-unnecessary-type-assertion',
    );
    assert.ok(
      errorBDiags.length > 0,
      `Step 3 (error B): expected diagnostics, got ${errorBDiags.length}`,
    );
    assert.ok(
      errorBDiags.some((d) =>
        d.message.includes('no-unnecessary-type-assertion'),
      ),
      `Step 3 (error B): expected no-unnecessary-type-assertion diagnostic, got: ${errorBDiags.map((d) => d.message).join(', ')}`,
    );
    // Verify error A is gone
    assert.ok(
      !errorBDiags.some((d) => d.message.includes('no-unsafe-member-access')),
      `Step 3 (error B): no-unsafe-member-access should be gone`,
    );

    // Step 4: Back to clean code — should have zero diagnostics.
    // Use waitForDiagnosticsCount instead of waitForDiagnosticsToChange
    // because debounce can cause intermediate diagnostic states on CI.
    await editor.edit((b) =>
      b.replace(fullRange(), '// all clean again\nexport {};\n'),
    );
    const finalDiags = await waitForDiagnosticsCount(doc, 0);
    assert.strictEqual(
      finalDiags.length,
      0,
      `Step 4 (clean again): expected 0 diagnostics, got ${finalDiags.length}`,
    );
  });

  test('code actions - preference order', async () => {
    const doc = await openFixture('autofix.ts');
    await vscode.window.showTextDocument(doc);

    const diagnostics = await waitForDiagnostics(doc);

    let comparedAutoFixAndDisable = false;
    for (const diagnostic of diagnostics) {
      // Filter quick fixes
      const codeActions = (
        await executeCodeActionProvider(doc.uri, diagnostic.range)
      ).filter(
        (action) => action.kind?.value === vscode.CodeActionKind.QuickFix.value,
      );

      // Check that if there are auto fixes, they are marked as preferred
      const autoFixActions = codeActions.filter(
        (action) =>
          action.title.toLowerCase().includes('fix') &&
          !action.title.toLowerCase().includes('disable'),
      );

      const disableActions = codeActions.filter((action) =>
        action.title.toLowerCase().includes('disable'),
      );

      // If both auto fix and disable actions exist, auto fix should be preferred
      if (autoFixActions.length > 0 && disableActions.length > 0) {
        comparedAutoFixAndDisable = true;
        assert.ok(
          autoFixActions.some((action) => action.isPreferred),
          'Auto fix actions should be marked as preferred',
        );
        assert.ok(
          !disableActions.some((action) => action.isPreferred),
          'Disable actions should not be marked as preferred when auto fixes exist',
        );
      }
    }
    assert.ok(
      comparedAutoFixAndDisable,
      `Expected at least one diagnostic with both auto-fix and disable actions. Diagnostics: ${diagnostics
        .map((diagnostic) => diagnostic.message)
        .join(' | ')}`,
    );
  });

  test('diagnostics correct after reverting and reopening the editor tab', async () => {
    // VS Code may retain a TextDocument model after its last editor closes, so
    // this test deliberately covers editor-tab lifecycle rather than claiming
    // an LSP didClose cycle: edit → revert → close tab → reopen.
    const doc = await openFixture('close-test.ts');
    const editor = await vscode.window.showTextDocument(doc);

    // 1. Wait for initial diagnostics
    const initialDiags = await waitForDiagnostics(doc);
    assert.ok(
      initialDiags.length > 0,
      `Expected diagnostics but got ${initialDiags.length}`,
    );
    const initialCount = initialDiags.length;

    // 2. Edit to clean code — diagnostics should drop to 0
    const fullRange = new vscode.Range(
      doc.positionAt(0),
      doc.positionAt(doc.getText().length),
    );
    await editor.edit((b) => b.replace(fullRange, '// clean\nexport {};\n'));
    const cleanDiags = await waitForDiagnosticsCount(doc, 0);
    assert.strictEqual(
      cleanDiags.length,
      0,
      `Expected 0 diagnostics after cleaning, got ${cleanDiags.length}`,
    );

    // 3. Revert the dirty overlay and close the exact editor tab before
    // reopening the original error content from disk.
    await closeTextEditor(doc);
    const doc2 = await openFixture('close-test.ts');
    await vscode.window.showTextDocument(doc2);

    // 4. Diagnostics should reappear — server correctly handles the cycle
    const reopenDiags = await waitForDiagnostics(doc2);
    assert.ok(
      reopenDiags.length > 0,
      `Expected diagnostics after restoring errors, got ${reopenDiags.length}`,
    );
  });

  test('no diagnostics for non-TypeScript files', async () => {
    const doc = await openFixture('styles.css');
    await vscode.window.showTextDocument(doc);

    // Wait a reasonable amount of time — diagnostics should NOT appear
    await new Promise((r) => setTimeout(r, 3000));

    const diagnostics = getRslintDiagnostics(doc);
    assert.strictEqual(
      diagnostics.length,
      0,
      `Expected 0 diagnostics for CSS file, got ${diagnostics.length}`,
    );
  });
});

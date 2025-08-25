import * as assert from 'assert';
import * as vscode from 'vscode';
import path from 'node:path';

suite('rslint extension', function () {
  this.timeout(50000);

  test('restart command is registered', async () => {
    // Check that the restart command is available
    const commands = await vscode.commands.getCommands(true);
    assert.ok(
      commands.includes('rslint.restart'),
      'rslint.restart command should be registered',
    );
  });

  // Helper function to wait for diagnostics
  async function waitForDiagnostics(
    doc: vscode.TextDocument,
  ): Promise<vscode.Diagnostic[]> {
    // Try multiple times to get diagnostics
    for (let i = 0; i < 10; i++) {
      const diagnostics = vscode.languages.getDiagnostics(doc.uri);
      if (diagnostics.length > 0) {
        return diagnostics;
      }

      // Wait for diagnostics change event or timeout
      await new Promise(resolve => {
        const disposable = vscode.languages.onDidChangeDiagnostics(e => {
          // Check if this event is for our document
          for (const uri of e.uris) {
            if (uri.toString() === doc.uri.toString()) {
              disposable.dispose();
              resolve(void 0);
              return;
            }
          }
        });
        // Wait 1 second then check again
        setTimeout(() => {
          disposable.dispose();
          resolve(void 0);
        }, 1000);
      });
    }

    return vscode.languages.getDiagnostics(doc.uri);
  }

  // Helper function to open a test fixture
  async function openFixture(filename: string): Promise<vscode.TextDocument> {
    return vscode.workspace.openTextDocument(
      path.resolve(
        require.resolve('@rslint/core'),
        '../..',
        `fixtures/src/`,
        filename,
      ),
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

  test('code actions - auto fix', async () => {
    const doc = await openFixture('autofix.ts');
    await vscode.window.showTextDocument(doc);

    const diagnostics = await waitForDiagnostics(doc);
    assert.ok(diagnostics.length > 0, 'Should have diagnostics');

    // Find the no-unnecessary-type-assertion diagnostic
    const typeAssertionDiag = diagnostics.find(
      d =>
        d.message.includes('no-unnecessary-type-assertion') ||
        (d.source === 'rslint' && d.message.includes('assertion')),
    );

    if (typeAssertionDiag) {
      // Request code actions for the diagnostic range
      const codeActions = await vscode.commands.executeCommand<
        vscode.CodeAction[]
      >('vscode.executeCodeActionProvider', doc.uri, typeAssertionDiag.range);

      assert.ok(
        codeActions && codeActions.length > 0,
        'Should have code actions',
      );

      // Look for auto fix action
      const autoFixAction = codeActions.find(
        action =>
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
    }
  });

  test('code actions - disable rule for line', async () => {
    const doc = await openFixture('disable.ts');
    await vscode.window.showTextDocument(doc);

    const diagnostics = await waitForDiagnostics(doc);
    assert.ok(diagnostics.length > 0, 'Should have diagnostics');

    // Find an unsafe diagnostic (these typically don't have auto fixes)
    const unsafeDiag = diagnostics.find(
      d => d.message.includes('unsafe') || d.message.includes('Unsafe'),
    );

    if (unsafeDiag) {
      // Request code actions for the diagnostic range
      const codeActions = await vscode.commands.executeCommand<
        vscode.CodeAction[]
      >('vscode.executeCodeActionProvider', doc.uri, unsafeDiag.range);

      assert.ok(
        codeActions && codeActions.length > 0,
        'Should have code actions',
      );

      // Look for disable rule for line action
      const disableLineAction = codeActions.find(
        action =>
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

      // Verify the edit contains eslint-disable-next-line
      const workspaceEdit = disableLineAction.edit;
      const edits = workspaceEdit.get(doc.uri);
      if (edits && edits.length > 0) {
        const editText = edits[0].newText;
        assert.ok(
          editText.includes('eslint-disable-next-line'),
          'Edit should contain eslint-disable-next-line comment',
        );
      }
    }
  });

  test('code actions - disable rule for file', async () => {
    const doc = await openFixture('disable-file.ts');
    await vscode.window.showTextDocument(doc);

    const diagnostics = await waitForDiagnostics(doc);
    assert.ok(diagnostics.length > 0, 'Should have diagnostics');

    // Find an unsafe diagnostic
    const unsafeDiag = diagnostics.find(
      d => d.message.includes('unsafe') || d.message.includes('Unsafe'),
    );

    if (unsafeDiag) {
      // Request code actions for the diagnostic range
      const codeActions = await vscode.commands.executeCommand<
        vscode.CodeAction[]
      >('vscode.executeCodeActionProvider', doc.uri, unsafeDiag.range);

      assert.ok(
        codeActions && codeActions.length > 0,
        'Should have code actions',
      );

      // Look for disable rule for file action
      const disableFileAction = codeActions.find(
        action =>
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

      // Verify the edit contains eslint-disable comment
      const workspaceEdit = disableFileAction.edit;
      const edits = workspaceEdit.get(doc.uri);
      if (edits && edits.length > 0) {
        const editText = edits[0].newText;
        assert.ok(
          editText.includes('eslint-disable') &&
            !editText.includes('-next-line'),
          'Edit should contain eslint-disable comment for entire file',
        );
      }
    }
  });

  test('code actions - range overlap', async () => {
    const doc = await openFixture('index.ts');
    await vscode.window.showTextDocument(doc);

    await waitForDiagnostics(doc);

    // Test that code actions are only provided for ranges that overlap with diagnostics
    const codeActionsEmptyRange = await vscode.commands.executeCommand<
      vscode.CodeAction[]
    >(
      'vscode.executeCodeActionProvider',
      doc.uri,
      new vscode.Range(100, 0, 100, 0), // Range with no diagnostics
    );

    // Should either be empty or only contain general actions (not diagnostic-specific)
    if (codeActionsEmptyRange) {
      const diagnosticSpecificActions = codeActionsEmptyRange.filter(
        action =>
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

  test('code actions - preference order', async () => {
    const doc = await openFixture('autofix.ts');
    await vscode.window.showTextDocument(doc);

    const diagnostics = await waitForDiagnostics(doc);

    for (const diagnostic of diagnostics) {
      const codeActions = await vscode.commands.executeCommand<
        vscode.CodeAction[]
      >('vscode.executeCodeActionProvider', doc.uri, diagnostic.range);

      if (codeActions && codeActions.length > 0) {
        // Check that if there are auto fixes, they are marked as preferred
        const autoFixActions = codeActions.filter(
          action =>
            action.title.toLowerCase().includes('fix') &&
            !action.title.toLowerCase().includes('disable'),
        );

        const disableActions = codeActions.filter(action =>
          action.title.toLowerCase().includes('disable'),
        );

        // If both auto fix and disable actions exist, auto fix should be preferred
        if (autoFixActions.length > 0 && disableActions.length > 0) {
          assert.ok(
            autoFixActions.some(action => action.isPreferred),
            'Auto fix actions should be marked as preferred',
          );
          assert.ok(
            !disableActions.some(action => action.isPreferred),
            'Disable actions should not be marked as preferred when auto fixes exist',
          );
        }
      }
    }
  });
});

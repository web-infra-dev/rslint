import * as assert from 'assert';
import * as vscode from 'vscode';
import path from 'node:path';

suite('rslint extension', function () {
  this.timeout(50000);

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

  // Helper function to wait for diagnostics to change from a previous state
  async function waitForDiagnosticsToChange(
    doc: vscode.TextDocument,
    previousCount: number,
    timeoutMs = 15000,
  ): Promise<vscode.Diagnostic[]> {
    const startTime = Date.now();

    while (Date.now() - startTime < timeoutMs) {
      const current = vscode.languages.getDiagnostics(doc.uri);
      if (current.length !== previousCount) {
        return current;
      }

      // Wait for diagnostics change event or short timeout
      await new Promise(resolve => {
        const disposable = vscode.languages.onDidChangeDiagnostics(e => {
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

  // Helper function to wait for diagnostics to reach a specific count
  async function waitForDiagnosticsCount(
    doc: vscode.TextDocument,
    expectedCount: number,
    timeoutMs = 15000,
  ): Promise<vscode.Diagnostic[]> {
    const startTime = Date.now();

    while (Date.now() - startTime < timeoutMs) {
      const current = vscode.languages.getDiagnostics(doc.uri);
      if (current.length === expectedCount) {
        return current;
      }

      await new Promise(resolve => {
        const disposable = vscode.languages.onDidChangeDiagnostics(e => {
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

  // Helper function to wait for diagnostics containing a specific message
  async function waitForDiagnosticsWithMessage(
    doc: vscode.TextDocument,
    messageSubstring: string,
    timeoutMs = 15000,
  ): Promise<vscode.Diagnostic[]> {
    const startTime = Date.now();

    while (Date.now() - startTime < timeoutMs) {
      const current = vscode.languages.getDiagnostics(doc.uri);
      if (current.some(d => d.message.includes(messageSubstring))) {
        return current;
      }

      // Wait for diagnostics change event or short timeout
      await new Promise(resolve => {
        const disposable = vscode.languages.onDidChangeDiagnostics(e => {
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
    await editor.edit(editBuilder => {
      editBuilder.replace(fullRange, '// no lint errors\nexport {};\n');
    });

    // 3. Wait for diagnostics to update (should decrease)
    const updatedDiags = await waitForDiagnosticsToChange(doc, initialCount);

    assert.ok(
      updatedDiags.length < initialCount,
      `Expected fewer diagnostics after removing errors. Before: ${initialCount}, After: ${updatedDiags.length}`,
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
    await editor.edit(editBuilder => {
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
    await editor.edit(b =>
      b.replace(fullRange(), 'const x: any = 1;\nexport {};\n'),
    );
    // Edit 2: still has errors
    await editor.edit(b =>
      b.replace(
        fullRange(),
        'const y: any = 2;\nconst z: any = 3;\nexport {};\n',
      ),
    );
    // Edit 3: error-free — the final state that matters
    await editor.edit(b =>
      b.replace(fullRange(), '// all clean\nexport {};\n'),
    );

    // 3. Wait for diagnostics to settle — should reflect the error-free final state
    const finalDiags = await waitForDiagnosticsToChange(doc, initialCount);

    assert.ok(
      finalDiags.length < initialCount,
      `After rapid edits ending with clean code, expected fewer diagnostics. Before: ${initialCount}, After: ${finalDiags.length}`,
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
    await editor.edit(b => {
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

    // Step 1: start from clean state to establish baseline
    await editor.edit(b =>
      b.replace(fullRange(), '// no errors\nexport {};\n'),
    );
    await waitForDiagnosticsCount(doc, 0, 10000);
    const cleanCount = vscode.languages.getDiagnostics(doc.uri).length;

    // Step 2: introduce errors
    await editor.edit(b =>
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
    await editor.edit(b =>
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

    // Step 1: Start with clean code — should have zero diagnostics
    await editor.edit(b =>
      b.replace(fullRange(), '// no errors\nexport {};\n'),
    );
    // -1 as previousCount ensures the condition (current.length !== -1) is always
    // true on the first check, effectively meaning "wait for any diagnostics event".
    await waitForDiagnosticsToChange(doc, -1, 5000);
    await new Promise(r => setTimeout(r, 1000));
    const cleanDiags = vscode.languages.getDiagnostics(doc.uri);
    assert.strictEqual(
      cleanDiags.length,
      0,
      `Step 1 (clean): expected 0 diagnostics, got ${cleanDiags.length}`,
    );

    // Step 2: Introduce error A — no-unsafe-member-access
    await editor.edit(b =>
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
      errorADiags.some(d => d.message.includes('no-unsafe-member-access')),
      `Step 2 (error A): expected no-unsafe-member-access diagnostic, got: ${errorADiags.map(d => d.message).join(', ')}`,
    );

    // Step 3: Change to error B — no-unnecessary-type-assertion (different rule)
    await editor.edit(b =>
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
      errorBDiags.some(d =>
        d.message.includes('no-unnecessary-type-assertion'),
      ),
      `Step 3 (error B): expected no-unnecessary-type-assertion diagnostic, got: ${errorBDiags.map(d => d.message).join(', ')}`,
    );
    // Verify error A is gone
    assert.ok(
      !errorBDiags.some(d => d.message.includes('no-unsafe-member-access')),
      `Step 3 (error B): no-unsafe-member-access should be gone`,
    );

    // Step 4: Back to clean code — should have zero diagnostics.
    // Use waitForDiagnosticsCount instead of waitForDiagnosticsToChange
    // because debounce can cause intermediate diagnostic states on CI.
    await editor.edit(b =>
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

    for (const diagnostic of diagnostics) {
      // Filter quick fixes
      const codeActions = (
        await vscode.commands.executeCommand<vscode.CodeAction[]>(
          'vscode.executeCodeActionProvider',
          doc.uri,
          diagnostic.range,
        )
      ).filter(
        action => action.kind?.value === vscode.CodeActionKind.QuickFix.value,
      );

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
  });

  test('diagnostics correct after close and reopen cycle', async () => {
    // Note: VSCode's test framework caches documents from openTextDocument,
    // so didClose is not reliably sent when editors close. Instead, we test
    // the full lifecycle: open → edit to clean → close → reopen original →
    // verify diagnostics reappear (server state is consistent).
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
    await editor.edit(b => b.replace(fullRange, '// clean\nexport {};\n'));
    const cleanDiags = await waitForDiagnosticsToChange(doc, initialCount);
    assert.strictEqual(
      cleanDiags.length,
      0,
      `Expected 0 diagnostics after cleaning, got ${cleanDiags.length}`,
    );

    // 3. Close editor and reopen with original error content
    await vscode.commands.executeCommand('workbench.action.closeActiveEditor');
    const doc2 = await openFixture('close-test.ts');
    const editor2 = await vscode.window.showTextDocument(doc2);

    // 4. Restore original error content
    const fullRange2 = new vscode.Range(
      doc2.positionAt(0),
      doc2.positionAt(doc2.getText().length),
    );
    await editor2.edit(b =>
      b.replace(fullRange2, 'const unsafeVal: any = 42;\nunsafeVal.prop;\n'),
    );

    // 5. Diagnostics should reappear — server correctly handles the cycle
    const reopenDiags = await waitForDiagnosticsToChange(doc2, 0);
    assert.ok(
      reopenDiags.length > 0,
      `Expected diagnostics after restoring errors, got ${reopenDiags.length}`,
    );
  });

  test('no diagnostics for non-TypeScript files', async () => {
    const doc = await vscode.workspace.openTextDocument(
      path.resolve(
        require.resolve('@rslint/core'),
        '../..',
        'fixtures/src/',
        'styles.css',
      ),
    );
    await vscode.window.showTextDocument(doc);

    // Wait a reasonable amount of time — diagnostics should NOT appear
    await new Promise(r => setTimeout(r, 3000));

    const diagnostics = vscode.languages.getDiagnostics(doc.uri);
    assert.strictEqual(
      diagnostics.length,
      0,
      `Expected 0 diagnostics for CSS file, got ${diagnostics.length}`,
    );
  });
});

import * as assert from 'assert';
import * as vscode from 'vscode';
import {
  waitForDiagnostics,
  waitForDiagnosticsCount,
  waitForContentChange,
  withOnSaveFixAll,
  replaceAll,
} from './fixall-helpers';

function assertHasFixableDiagnostic(
  diagnostics: vscode.Diagnostic[],
  context: string,
): void {
  assert.ok(
    diagnostics.some((d) =>
      d.message.includes('no-unnecessary-type-assertion'),
    ),
    `${context}: expected no-unnecessary-type-assertion. Got: ${diagnostics
      .map((d) => d.message)
      .join(' | ')}`,
  );
}

suite('rslint fixAll - on-save', function () {
  this.timeout(120000);

  test('generic source.fixAll triggers rslint via on-save', async () => {
    await withOnSaveFixAll(
      async (doc, editor) => {
        await replaceAll(
          editor,
          "const gfVal: string = 'x';\nconst gfRes = (gfVal as string).trim();\n",
        );

        const diags = await waitForDiagnostics(doc);
        assertHasFixableDiagnostic(diags, 'generic source.fixAll setup');

        assert.ok(
          await doc.save(),
          'Document should complete the generic source.fixAll save pipeline',
        );
        await waitForContentChange(
          doc,
          (content) => !content.includes('gfVal as string'),
          60000,
        );

        assert.ok(
          !doc.getText().includes('gfVal as string'),
          `Generic source.fixAll should trigger rslint fixAll.\nContent: ${doc.getText()}`,
        );
      },
      { 'source.fixAll': 'explicit' },
    );
  });

  test('fixable issues get auto-fixed', async () => {
    await withOnSaveFixAll(async (doc, editor) => {
      const fixableContent = [
        "const saveVal: string = 'hello';",
        'const saveResult = (saveVal as string).toUpperCase();',
        '',
      ].join('\n');
      await replaceAll(editor, fixableContent);

      const diags = await waitForDiagnostics(doc);
      assertHasFixableDiagnostic(diags, 'fixable on-save setup');

      assert.ok(await doc.save(), 'Fixable document should save');
      await waitForContentChange(
        doc,
        (content) => !content.includes('saveVal as string'),
        60000,
      );

      assert.ok(
        !doc.getText().includes('saveVal as string'),
        `Type assertion should be removed after on-save fixAll.\nContent: ${doc.getText()}`,
      );
    });
  });

  test('clean file saves without content change', async () => {
    await withOnSaveFixAll(async (doc, editor) => {
      // Probe: prove on-save is active
      await replaceAll(
        editor,
        "const probeVal: string = 'x';\nconst probeRes = (probeVal as string).trim();\n",
      );
      const probeDiags = await waitForDiagnostics(doc);
      assertHasFixableDiagnostic(probeDiags, 'clean-file probe setup');
      assert.ok(await doc.save(), 'Clean-file probe should save');
      await waitForContentChange(
        doc,
        (content) => !content.includes('probeVal as string'),
        60000,
      );
      assert.ok(
        !doc.getText().includes('probeVal as string'),
        'Probe: on-save fixAll should be active',
      );

      // Clean content
      const cleanContent =
        '// no issues\nconst cleanOnSave = 42;\nexport {};\n';
      await replaceAll(editor, cleanContent);
      await new Promise((r) => setTimeout(r, 3000));

      assert.ok(await doc.save(), 'Clean document should save');
      await new Promise((r) => setTimeout(r, 2000));

      assert.strictEqual(
        doc.getText(),
        cleanContent,
        'Clean file content should not change after on-save with fixAll',
      );
    });
  });

  test('non-fixable diagnostics remain, content unchanged', async () => {
    await withOnSaveFixAll(async (doc, editor) => {
      // Probe
      await replaceAll(
        editor,
        "const probeVal2: string = 'x';\nconst probeRes2 = (probeVal2 as string).trim();\n",
      );
      const probeDiags = await waitForDiagnostics(doc);
      assertHasFixableDiagnostic(probeDiags, 'non-fixable probe setup');
      assert.ok(await doc.save(), 'Non-fixable probe should save');
      await waitForContentChange(
        doc,
        (content) => !content.includes('probeVal2 as string'),
        60000,
      );
      assert.ok(
        !doc.getText().includes('probeVal2 as string'),
        'Probe: on-save fixAll should be active',
      );

      // Non-fixable content
      const content = 'const nfOnSave: any = {};\nnfOnSave.foo;\n';
      await replaceAll(editor, content);

      const diags = await waitForDiagnostics(doc);
      assert.ok(diags.length > 0, 'Should have non-fixable diagnostics');

      assert.ok(await doc.save(), 'Non-fixable document should save');
      await new Promise((r) => setTimeout(r, 2000));

      assert.strictEqual(
        doc.getText(),
        content,
        'Non-fixable file content should not change after on-save',
      );

      const diagsAfter = vscode.languages.getDiagnostics(doc.uri);
      assert.ok(
        diagsAfter.length > 0,
        'Non-fixable diagnostics should remain after save',
      );
    });
  });

  test('edit then immediately save (debounce not fired)', async () => {
    await withOnSaveFixAll(async (doc, editor) => {
      await replaceAll(editor, '// clean start\nexport {};\n');
      assert.ok(await doc.save(), 'Initial clean document should save');
      const cleanDiagnostics = await waitForDiagnosticsCount(doc, 0);
      assert.strictEqual(
        cleanDiagnostics.length,
        0,
        `Initial clean document should have no diagnostics. Got: ${cleanDiagnostics
          .map((d) => d.message)
          .join(' | ')}`,
      );

      await replaceAll(
        editor,
        "const quickVal: string = 'x';\nconst quickRes = (quickVal as string).trim();\n",
      );

      assert.ok(
        await doc.save(),
        'Edited document should complete the code-action-on-save pipeline',
      );

      // Event-driven wait — Windows CI needs more headroom than the previous
      // 20 s polling loop, and `onDidChangeTextDocument` resolves the moment
      // the on-save fixAll edit lands (sub-ms vs. 500 ms poll cadence).
      await waitForContentChange(
        doc,
        (content) => !content.includes('quickVal as string'),
        60000,
      );

      assert.ok(
        !doc.getText().includes('quickVal as string'),
        `On-save fixAll should work even when debounce has not fired.\nContent: ${doc.getText()}`,
      );
    });
  });
});

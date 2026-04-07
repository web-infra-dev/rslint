import * as assert from 'assert';
import * as vscode from 'vscode';
import fs from 'node:fs';
import path from 'node:path';
import {
  waitForDiagnostics,
  waitForDiagnosticsCount,
  withOnSaveFixAll,
  replaceAll,
  getFixturesDir,
} from './fixall-helpers';

suite('rslint fixAll - on-save', function () {
  this.timeout(50000);

  test('generic source.fixAll triggers rslint via on-save', async () => {
    const fixturesDir = getFixturesDir();
    const tmpFile = path.join(
      fixturesDir,
      'src',
      `_fixall_generic_${Date.now()}.ts`,
    );
    fs.writeFileSync(tmpFile, '// placeholder\n', 'utf-8');

    try {
      const config = vscode.workspace.getConfiguration('editor');
      const previousValue = config.get('codeActionsOnSave');
      await config.update(
        'codeActionsOnSave',
        { 'source.fixAll': 'explicit' },
        vscode.ConfigurationTarget.Workspace,
      );

      try {
        const doc = await vscode.workspace.openTextDocument(tmpFile);
        const editor = await vscode.window.showTextDocument(doc);

        await replaceAll(
          editor,
          "const gfVal: string = 'x';\nconst gfRes = (gfVal as string).trim();\n",
        );

        const diags = await waitForDiagnostics(doc);
        if (
          !diags.some((d) =>
            d.message.includes('no-unnecessary-type-assertion'),
          )
        ) {
          return;
        }

        await doc.save();

        const startTime = Date.now();
        while (
          doc.getText().includes('gfVal as string') &&
          Date.now() - startTime < 10000
        ) {
          await new Promise((r) => setTimeout(r, 500));
        }

        assert.ok(
          !doc.getText().includes('gfVal as string'),
          `Generic source.fixAll should trigger rslint fixAll.\nContent: ${doc.getText()}`,
        );
      } finally {
        await config.update(
          'codeActionsOnSave',
          previousValue,
          vscode.ConfigurationTarget.Workspace,
        );
      }
    } finally {
      if (fs.existsSync(tmpFile)) {
        fs.unlinkSync(tmpFile);
      }
    }
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
      const hasFixable = diags.some((d) =>
        d.message.includes('no-unnecessary-type-assertion'),
      );
      if (!hasFixable) return;

      await doc.save();

      const startTime = Date.now();
      while (
        doc.getText().includes('saveVal as string') &&
        Date.now() - startTime < 10000
      ) {
        await new Promise((r) => setTimeout(r, 500));
      }

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
      if (
        !probeDiags.some((d) =>
          d.message.includes('no-unnecessary-type-assertion'),
        )
      ) {
        return;
      }
      await doc.save();
      const probeStart = Date.now();
      while (
        doc.getText().includes('probeVal as string') &&
        Date.now() - probeStart < 10000
      ) {
        await new Promise((r) => setTimeout(r, 500));
      }
      assert.ok(
        !doc.getText().includes('probeVal as string'),
        'Probe: on-save fixAll should be active',
      );

      // Clean content
      const cleanContent =
        '// no issues\nconst cleanOnSave = 42;\nexport {};\n';
      await replaceAll(editor, cleanContent);
      await new Promise((r) => setTimeout(r, 3000));

      await doc.save();
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
      if (
        !probeDiags.some((d) =>
          d.message.includes('no-unnecessary-type-assertion'),
        )
      ) {
        return;
      }
      await doc.save();
      const probeStart = Date.now();
      while (
        doc.getText().includes('probeVal2 as string') &&
        Date.now() - probeStart < 10000
      ) {
        await new Promise((r) => setTimeout(r, 500));
      }
      assert.ok(
        !doc.getText().includes('probeVal2 as string'),
        'Probe: on-save fixAll should be active',
      );

      // Non-fixable content
      const content = 'const nfOnSave: any = {};\nnfOnSave.foo;\n';
      await replaceAll(editor, content);

      const diags = await waitForDiagnostics(doc);
      assert.ok(diags.length > 0, 'Should have non-fixable diagnostics');

      await doc.save();
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
      await doc.save();
      await waitForDiagnosticsCount(doc, 0);

      await replaceAll(
        editor,
        "const quickVal: string = 'x';\nconst quickRes = (quickVal as string).trim();\n",
      );

      await doc.save();

      const startTime = Date.now();
      while (
        doc.getText().includes('quickVal as string') &&
        Date.now() - startTime < 10000
      ) {
        await new Promise((r) => setTimeout(r, 500));
      }

      assert.ok(
        !doc.getText().includes('quickVal as string'),
        `On-save fixAll should work even when debounce has not fired.\nContent: ${doc.getText()}`,
      );
    });
  });
});

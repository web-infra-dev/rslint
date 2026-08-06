import * as assert from 'assert';
import * as vscode from 'vscode';
import path from 'node:path';
import fs from 'node:fs';
import { waitForRslintDiagnostics } from '../utils/diagnostics';
import { waitForCodeActionRegistryQuiescence } from '../utils/codeActionRegistry';

/**
 * A byte order mark is the one thing a linter can see about a file that an
 * editor's document cannot hold. VS Code decodes the file before handing it to
 * the model — the mark becomes the document's *encoding*, shown in the status
 * bar as "UTF-8 with BOM", never a character the text contains.
 *
 * So the language server does not run `unicode-bom` at all. The mark is not in
 * the document, no text edit reaches it, and for an unsaved buffer the only
 * remaining witness is the file on disk, which the buffer may already disagree
 * with. `rslint --fix`, which rewrites the file itself, is where the rule
 * applies.
 *
 * `src/marked.ts` is written here rather than committed: a file whose first
 * bytes are EF BB BF is invisible in review and formatters keep wanting to
 * strip it. The mark is only ever written as the escape `\uFEFF`, never as the
 * character itself.
 *
 * The fixture also enables `no-var` and the source declares one, so every
 * assertion of silence below runs against a file the server is demonstrably
 * linting.
 */
suite('rslint unicode-bom over LSP', function () {
  this.timeout(120000);

  const BOM = '\uFEFF';
  const markedSource =
    'export function marked() {\n  var value = 1;\n  return value;\n}\n';

  function workspaceRoot(): string {
    const folder = vscode.workspace.workspaceFolders?.[0];
    if (!folder) throw new Error('VS Code test workspace is unavailable');
    return folder.uri.fsPath;
  }

  function fixturePath(filename: string): string {
    return path.join(workspaceRoot(), 'src', filename);
  }

  async function openMarked(): Promise<vscode.TextDocument> {
    const doc = await vscode.workspace.openTextDocument(
      fixturePath('marked.ts'),
    );
    await vscode.window.showTextDocument(doc);
    return doc;
  }

  /** Waits for the pass that reports `no-var`, and returns everything it said. */
  async function lintedDiagnostics(
    doc: vscode.TextDocument,
  ): Promise<vscode.Diagnostic[]> {
    return waitForRslintDiagnostics(doc, (all) =>
      all.some((d) => d.message.includes('no-var')),
    );
  }

  suiteSetup(() => {
    fs.writeFileSync(fixturePath('marked.ts'), BOM + markedSource, 'utf8');
  });

  suiteTeardown(() => {
    fs.rmSync(fixturePath('marked.ts'), { force: true });
  });

  // ======== The constraint the behavior rests on ========

  test('the file carries a mark that its editor document does not', async () => {
    const bytes = fs.readFileSync(fixturePath('marked.ts'));
    assert.deepStrictEqual(
      [...bytes.subarray(0, 3)],
      [0xef, 0xbb, 0xbf],
      'the fixture must start with the UTF-8 encoding of U+FEFF',
    );

    const doc = await openMarked();
    const text = doc.getText();

    assert.ok(
      !text.startsWith(BOM),
      'VS Code decodes the mark into the document encoding, so the model text ' +
        'must not contain U+FEFF. If this ever fails, an editor could see the ' +
        'mark after all and skipping the rule would be worth revisiting.',
    );
    assert.strictEqual(
      text,
      markedSource,
      'the document should hold the source without the mark',
    );
  });

  // ======== The rule does not run in the editor ========

  test('reports nothing about the mark on a marked file', async () => {
    const doc = await openMarked();
    const diagnostics = await lintedDiagnostics(doc);

    assert.deepStrictEqual(
      diagnostics
        .filter((d) => d.message.includes('unicode-bom'))
        .map((d) => d.message),
      [],
      `expected no unicode-bom diagnostic, got: ${diagnostics
        .map((d) => d.message)
        .join(' | ')}`,
    );
  });

  test('offers no code action for the mark', async () => {
    const doc = await openMarked();
    await lintedDiagnostics(doc);
    await waitForCodeActionRegistryQuiescence();

    const before = doc.getText();
    const quickFixes =
      (await vscode.commands.executeCommand<vscode.CodeAction[]>(
        'vscode.executeCodeActionProvider',
        doc.uri,
        new vscode.Range(0, 0, doc.lineCount, 0),
        vscode.CodeActionKind.QuickFix.value,
      )) ?? [];
    const titles = quickFixes.map((action) => action.title);

    assert.ok(
      !titles.some(
        (title) =>
          title.includes('unicode-bom') || title.includes('Unicode BOM'),
      ),
      `no unicode-bom action should be offered, got: ${titles.join(' | ')}`,
    );
    // no-var's own actions are what make the absence above meaningful: quick
    // fixes were computed for this file, and only unicode-bom's are missing.
    assert.ok(
      titles.includes('Disable no-var for this line'),
      `expected no-var's actions to be present, got: ${titles.join(' | ')}`,
    );

    assert.strictEqual(doc.getText(), before, 'the document must be untouched');
    assert.ok(!doc.isDirty, 'nothing should have modified the document');
  });
});

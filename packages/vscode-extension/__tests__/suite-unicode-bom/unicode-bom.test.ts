import * as assert from 'assert';
import * as vscode from 'vscode';
import path from 'node:path';
import fs from 'node:fs';
import {
  waitForRslintDiagnostics,
  waitForRslintDiagnosticsCount,
} from '../utils/diagnostics';
import { waitForCodeActionRegistryQuiescence } from '../utils/codeActionRegistry';

/**
 * A byte order mark is the one thing a linter can see about a file that an
 * editor's document cannot hold. VS Code decodes the file before handing it to
 * the model — the mark becomes the document's *encoding*, shown in the status
 * bar as "UTF-8 with BOM", never a character the text contains.
 *
 * So `unicode-bom` is reported over LSP but cannot be fixed there: its fix is a
 * range over [-1, 0], one position ahead of the text, and no text edit says
 * that. The server withholds the fix rather than offering an action that runs
 * and changes nothing. `rslint --fix`, which rewrites the file itself, removes
 * the mark.
 *
 * `src/marked.ts` is written here rather than committed: a file whose first
 * bytes are EF BB BF is invisible in review and formatters keep wanting to
 * strip it. The mark is only ever written as the escape `\uFEFF`, never as the
 * character itself.
 */
suite('rslint unicode-bom over LSP', function () {
  this.timeout(120000);

  const BOM = '\uFEFF';
  const markedSource = 'export const marked = 1;\n';

  function workspaceRoot(): string {
    const folder = vscode.workspace.workspaceFolders?.[0];
    if (!folder) throw new Error('VS Code test workspace is unavailable');
    return folder.uri.fsPath;
  }

  function fixturePath(filename: string): string {
    return path.join(workspaceRoot(), 'src', filename);
  }

  async function openFixture(filename: string): Promise<vscode.TextDocument> {
    const doc = await vscode.workspace.openTextDocument(fixturePath(filename));
    await vscode.window.showTextDocument(doc);
    return doc;
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

    const doc = await openFixture('marked.ts');
    const text = doc.getText();

    assert.ok(
      !text.startsWith(BOM),
      'VS Code decodes the mark into the document encoding, so the model text ' +
        'must not contain U+FEFF. If this ever fails, a text edit could remove ' +
        'the mark after all and withholding the fix would be wrong.',
    );
    assert.strictEqual(
      text,
      markedSource,
      'the document should hold the source without the mark',
    );
  });

  // ======== Reported, but not fixable ========

  test('reports the mark on a marked file', async () => {
    const doc = await openFixture('marked.ts');
    const diagnostics = await waitForRslintDiagnostics(doc, (all) =>
      all.some((d) => d.message.includes('unicode-bom')),
    );
    const marks = diagnostics.filter((d) => d.message.includes('unicode-bom'));

    assert.strictEqual(
      marks.length,
      1,
      `expected exactly one unicode-bom diagnostic, got: ${diagnostics
        .map((d) => d.message)
        .join(' | ')}`,
    );
    assert.strictEqual(marks[0].range.start.line, 0, 'sits on line 1');
    assert.strictEqual(marks[0].range.start.character, 0, 'sits on column 1');
  });

  test('offers no code action that claims to remove the mark', async () => {
    const doc = await openFixture('marked.ts');
    await waitForRslintDiagnostics(doc, (all) =>
      all.some((d) => d.message.includes('unicode-bom')),
    );
    await waitForCodeActionRegistryQuiescence();

    const before = doc.getText();
    const actionsOfKind = async (
      kind: vscode.CodeActionKind,
    ): Promise<vscode.CodeAction[]> =>
      (await vscode.commands.executeCommand<vscode.CodeAction[]>(
        'vscode.executeCodeActionProvider',
        doc.uri,
        new vscode.Range(0, 0, doc.lineCount, 0),
        kind.value,
      )) ?? [];

    const quickFixes = await actionsOfKind(vscode.CodeActionKind.QuickFix);
    const titles = quickFixes.map((action) => action.title);

    // rslint titles a rule's own fix "Fix: <message>". That is the one the
    // server withholds, because no text edit removes a mark the document never
    // had. Matched on the message too, so a TypeScript quick fix on this file
    // could neither satisfy nor trip the assertion.
    assert.ok(
      !titles.some(
        (title) => title.startsWith('Fix:') && title.includes('Unicode BOM'),
      ),
      `no fix action should be offered, got: ${titles.join(' | ')}`,
    );
    // The disable-directive actions still appear. They are what make the
    // absence above meaningful: quick fixes were computed for this diagnostic,
    // and only the fix itself is missing.
    for (const expected of [
      'Disable unicode-bom for this line',
      'Disable unicode-bom for entire file',
    ]) {
      assert.ok(
        titles.includes(expected),
        `expected ${JSON.stringify(expected)}, got: ${titles.join(' | ')}`,
      );
    }

    // fixAll applies fixes only, never disable directives, so it has nothing
    // to contribute at all.
    const fixAll = await actionsOfKind(
      vscode.CodeActionKind.SourceFixAll.append('rslint'),
    );
    assert.deepStrictEqual(
      fixAll.flatMap((action) => action.edit?.get(doc.uri) ?? []),
      [],
      'fixAll should produce no edits for a marked file',
    );

    assert.strictEqual(doc.getText(), before, 'the document must be untouched');
    assert.ok(!doc.isDirty, 'nothing should have modified the document');
  });

  test('reports nothing on a file whose bytes carry no mark', async () => {
    // Lint the marked file first: zero diagnostics is also what a server that
    // never answered looks like, so establish the rule is live in this
    // workspace before reading silence as an answer.
    const marked = await openFixture('marked.ts');
    await waitForRslintDiagnostics(marked, (all) =>
      all.some((d) => d.message.includes('unicode-bom')),
    );

    const doc = await openFixture('unmarked.ts');
    const diagnostics = await waitForRslintDiagnosticsCount(doc, 0);
    assert.deepStrictEqual(
      diagnostics.map((d) => d.message),
      [],
      'an unmarked file should report nothing',
    );
  });
});

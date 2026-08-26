import * as assert from 'assert';
import * as vscode from 'vscode';
import path from 'node:path';
import { waitForRslintDiagnostics } from '../utils/diagnostics';
import { closeTextEditor } from '../utils/documents';

/**
 * `import/no-cycle` is the one rule whose answer for a file depends on every
 * other file of the program, and its cross-file structures are cached per
 * Program. In the editor that cache is exercised the hard way: every buffer
 * edit produces a new Program, the entry keyed by the old one becomes
 * unreachable, and the next lint must rebuild from the new Program's overlay
 * text — never answer from the graph of a Program that no longer describes
 * the workspace.
 *
 * The fixture is a three-file cycle, `a.ts => b.ts => c.ts => a.ts`, plus an
 * acyclic `leaf.ts`. Every file declares a `var`, so `no-var` marks each
 * publish that reaches the client and an assertion of "no cycle reported"
 * always runs against a pass that demonstrably linted the file.
 *
 * The edits below stay in the editor buffer and are never saved: the files on
 * disk hold the cycle throughout, and only the overlay the language server
 * mirrors changes. A report that clears — or returns — therefore proves the
 * lint answered from the current overlay Program, not from any cached graph
 * of a previous one.
 */
suite('rslint import/no-cycle over LSP', function () {
  this.timeout(120000);

  const cycleMarker = '[import/no-cycle]';
  const sentinelMarker = '[no-var]';

  const brokenC = [
    'export var witnessC = 1;',
    '',
    'export function fromC(): number {',
    '  return witnessC;',
    '}',
    '',
  ].join('\n');

  let touchCount = 0;
  const openedDocuments = new Set<vscode.TextDocument>();
  let originalC: string | undefined;

  function workspaceRoot(): string {
    const folder = vscode.workspace.workspaceFolders?.[0];
    if (!folder) throw new Error('VS Code test workspace is unavailable');
    return folder.uri.fsPath;
  }

  async function openFixture(filename: string): Promise<vscode.TextDocument> {
    const doc = await vscode.workspace.openTextDocument(
      path.join(workspaceRoot(), 'src', filename),
    );
    await vscode.window.showTextDocument(doc, { preview: false });
    openedDocuments.add(doc);
    return doc;
  }

  function cycleDiagnostics(
    diagnostics: vscode.Diagnostic[],
  ): vscode.Diagnostic[] {
    return diagnostics.filter((d) => d.message.includes(cycleMarker));
  }

  function isLintedPass(diagnostics: vscode.Diagnostic[]): boolean {
    return diagnostics.some((d) => d.message.includes(sentinelMarker));
  }

  /** Replaces the whole buffer, leaving the document dirty and unsaved. */
  async function replaceDocumentText(
    doc: vscode.TextDocument,
    text: string,
  ): Promise<void> {
    const editor = await vscode.window.showTextDocument(doc, {
      preview: false,
    });
    const fullRange = new vscode.Range(
      new vscode.Position(0, 0),
      doc.lineAt(doc.lineCount - 1).range.end,
    );
    const applied = await editor.edit((edit) => edit.replace(fullRange, text));
    assert.ok(applied, `could not replace the content of ${doc.uri}`);
  }

  /**
   * Diagnostics are published per document and only when that document is
   * linted again, so an edit elsewhere does not repaint this file by itself.
   * Appending a comment line is the editor gesture that forces the next pass —
   * and, being an edit, it also forces that pass onto yet another new Program.
   */
  async function touchAndWait(
    doc: vscode.TextDocument,
    predicate: (diagnostics: vscode.Diagnostic[]) => boolean,
  ): Promise<vscode.Diagnostic[]> {
    const editor = await vscode.window.showTextDocument(doc, {
      preview: false,
    });
    touchCount += 1;
    const appended = await editor.edit((edit) =>
      edit.insert(
        new vscode.Position(doc.lineCount, 0),
        `// relint ${touchCount}\n`,
      ),
    );
    assert.ok(appended, `could not touch ${doc.uri}`);
    return waitForRslintDiagnostics(doc, predicate);
  }

  suiteTeardown(async () => {
    for (const doc of openedDocuments) {
      await closeTextEditor(doc);
    }
  });

  test('every member of the cycle reports it, with the route as written', async () => {
    const docA = await openFixture('a.ts');
    const diagnosticsA = await waitForRslintDiagnostics(
      docA,
      (all) => cycleDiagnostics(all).length > 0,
    );
    const [cycleA] = cycleDiagnostics(diagnosticsA);
    assert.ok(
      cycleA.message.includes('Dependency cycle via ./c:1'),
      `a.ts should report the route through b.ts's import, got: ${cycleA.message}`,
    );
    assert.strictEqual(
      cycleA.range.start.line,
      0,
      'the report should sit on the import declaration',
    );

    const docB = await openFixture('b.ts');
    const diagnosticsB = await waitForRslintDiagnostics(
      docB,
      (all) => cycleDiagnostics(all).length > 0,
    );
    assert.ok(
      cycleDiagnostics(diagnosticsB)[0].message.includes(
        'Dependency cycle via ./a:1',
      ),
      `b.ts should report its own route, got: ${diagnosticsB
        .map((d) => d.message)
        .join(' | ')}`,
    );
  });

  test('an acyclic file of the same program stays clean while demonstrably linted', async () => {
    const doc = await openFixture('leaf.ts');
    const diagnostics = await waitForRslintDiagnostics(doc, isLintedPass);
    assert.deepStrictEqual(
      cycleDiagnostics(diagnostics).map((d) => d.message),
      [],
      'leaf.ts imports nothing and must not be caught in the cycle',
    );
  });

  test('an unsaved edit two hops away clears the report, and its revert restores it', async () => {
    const docA = await openFixture('a.ts');
    await waitForRslintDiagnostics(
      docA,
      (all) => cycleDiagnostics(all).length > 0,
    );

    // Break the cycle at its far end: a.ts keeps its import of b.ts, but the
    // route back from c.ts disappears from the overlay only.
    const docC = await openFixture('c.ts');
    originalC = docC.getText();
    await replaceDocumentText(docC, brokenC);
    const diagnosticsC = await waitForRslintDiagnostics(
      docC,
      (all) => isLintedPass(all) && cycleDiagnostics(all).length === 0,
    );
    assert.ok(
      isLintedPass(diagnosticsC),
      'c.ts should have been re-linted from its edited buffer',
    );

    // a.ts is unchanged in meaning, so only a fresh cross-file answer — built
    // from the Program that holds c.ts's edited buffer — can clear its report.
    const clearedA = await touchAndWait(
      docA,
      (all) => isLintedPass(all) && cycleDiagnostics(all).length === 0,
    );
    assert.deepStrictEqual(
      cycleDiagnostics(clearedA).map((d) => d.message),
      [],
      'a.ts must stop reporting once the overlay no longer closes the cycle',
    );

    // Put the import back, still without saving: the report must return just
    // as promptly, proving the cleared answer was not cached against a.ts.
    await replaceDocumentText(docC, originalC);
    await waitForRslintDiagnostics(
      docC,
      (all) => cycleDiagnostics(all).length > 0,
    );
    const restoredA = await touchAndWait(
      docA,
      (all) => cycleDiagnostics(all).length > 0,
    );
    assert.ok(
      cycleDiagnostics(restoredA)[0].message.includes(
        'Dependency cycle via ./c:1',
      ),
      `the restored report should carry the same route, got: ${restoredA
        .map((d) => d.message)
        .join(' | ')}`,
    );
  });
});

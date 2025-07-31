import * as assert from 'assert';
import * as vscode from 'vscode';
import path from 'node:path';

suite('rslint extension', function () {
  this.timeout(50000);
  test('diagnostics', async () => {
    const doc = await vscode.workspace.openTextDocument(
      path.resolve(
        require.resolve('@rslint/core'),
        '../..',
        'fixtures/src/index.ts',
      ),
    );

    await vscode.window.showTextDocument(doc);
    await new Promise(resolve => {
      const disposable = vscode.languages.onDidChangeDiagnostics(() => {
        disposable.dispose();
        resolve(void 0);
      });
    });

    const diagnostics = vscode.languages.getDiagnostics(doc.uri);
    assert.ok(diagnostics.length > 0);
  });

  test('restart server command exists', async () => {
    // Check that the restart server command is registered
    const commands = await vscode.commands.getCommands(true);
    assert.ok(
      commands.includes('rslint.restartServer'),
      'rslint.restartServer command should be registered',
    );
  });

  test('restart server command executes', async () => {
    // This test verifies that the restart server command can be executed
    // without throwing errors
    try {
      await vscode.commands.executeCommand('rslint.restartServer');
      // If we reach here, the command executed successfully
      assert.ok(true, 'restart server command executed successfully');
    } catch (error) {
      assert.fail(`restart server command failed: ${error}`);
    }
  });
});

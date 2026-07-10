import fastGlob from 'fast-glob';
import path from 'node:path';
import Mocha from 'mocha';
import * as vscode from 'vscode';

const extensionId = 'rstack.rslint';

export function run(
  testPath: string,
  callback: (error: unknown, failures?: number) => void,
): void {
  void activateAndRun(testPath, callback);
}

async function activateAndRun(
  testPath: string,
  callback: (error: unknown, failures?: number) => void,
): Promise<void> {
  try {
    const extension = vscode.extensions.getExtension(extensionId);
    if (!extension) {
      throw new Error(`Extension ${extensionId} is unavailable`);
    }

    // Extension.activate() is the public readiness contract. Await it before
    // loading any tests so cold-start saves and code-action requests cannot run
    // while the language client and workspace configuration are still starting.
    await extension.activate();

    const files = fastGlob.sync('**/*.test.js', { cwd: testPath });
    const mocha = new Mocha({ ui: 'tdd' });
    files.forEach((file) => mocha.addFile(path.join(testPath, file)));
    mocha.run((failures) => callback(null, failures));
  } catch (error) {
    callback(error);
  }
}

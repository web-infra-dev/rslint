import fastGlob from 'fast-glob';
import path from 'node:path';
import Mocha from 'mocha';
import * as vscode from 'vscode';

const extensionId = 'rstack.rslint';
const typescriptExtensionId = 'vscode.typescript-language-features';
const typescriptCodeActionCommands = [
  '_typescript.applyCodeActionCommand',
  '_typescript.didApplyRefactoring',
  '_typescript.didOrganizeImports',
] as const;

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
    await waitForTypeScriptCodeActionProviders();

    const files = fastGlob.sync('**/*.test.js', { cwd: testPath }).sort();
    if (files.length === 0) {
      throw new Error(`No compiled test files found in ${testPath}`);
    }
    const mocha = new Mocha({ ui: 'tdd' });
    files.forEach((file) => mocha.addFile(path.join(testPath, file)));
    mocha.run((failures) => callback(null, failures));
  } catch (error) {
    callback(error);
  }
}

/**
 * VS Code's built-in TypeScript extension activates lazily. Its activation
 * function returns before tsserver is ready, then four modules asynchronously
 * register the TypeScript code-action providers. A provider registration while
 * code actions on save are running cancels that save in VS Code 1.128.
 *
 * Wait for the observable registrations from the quick-fix, refactor,
 * organize-imports, and fix-all modules. This is a readiness barrier only:
 * the save assertions themselves remain single-shot.
 */
async function waitForTypeScriptCodeActionProviders(): Promise<void> {
  const typescriptExtension = vscode.extensions.getExtension(
    typescriptExtensionId,
  );
  if (!typescriptExtension) {
    throw new Error(`Extension ${typescriptExtensionId} is unavailable`);
  }

  const [typescriptFile] = await vscode.workspace.findFiles(
    '**/*.ts',
    '**/{node_modules,.git}/**',
    1,
  );
  if (!typescriptFile) {
    return;
  }

  const document = await vscode.workspace.openTextDocument(typescriptFile);
  await typescriptExtension.activate();

  const timeoutMs = 60_000;
  const deadline = Date.now() + timeoutMs;
  let lastProbeError: unknown;

  while (Date.now() < deadline) {
    // `filterInternal: false` is intentional: the three sentinels are
    // internal commands whose registration happens in the constructors of
    // the quick-fix, refactor, and organize-imports providers.
    const commands = await vscode.commands.getCommands(false);
    const commandsReady = typescriptCodeActionCommands.every((command) =>
      commands.includes(command),
    );

    let fixAllReady = false;
    if (commandsReady) {
      try {
        const actions = await vscode.commands.executeCommand<
          vscode.CodeAction[]
        >(
          'vscode.executeCodeActionProvider',
          document.uri,
          new vscode.Range(0, 0, 0, 0),
          'source.fixAll.ts',
        );
        fixAllReady =
          actions?.some(
            (action) => action.kind?.value === 'source.fixAll.ts',
          ) ?? false;
        lastProbeError = undefined;
      } catch (error) {
        // A provider registering during this non-mutating probe cancels it.
        // Keep waiting for the complete provider set, but fail closed on
        // timeout.
        lastProbeError = error;
      }
    }

    if (commandsReady && fixAllReady) {
      return;
    }
    await new Promise((resolve) => setTimeout(resolve, 50));
  }

  const commands = await vscode.commands.getCommands(false);
  const missingCommands = typescriptCodeActionCommands.filter(
    (command) => !commands.includes(command),
  );
  throw new Error(
    `Timed out after ${timeoutMs}ms waiting for TypeScript code-action providers` +
      (missingCommands.length > 0
        ? `; missing commands: ${missingCommands.join(', ')}`
        : '') +
      (lastProbeError ? `; last probe error: ${String(lastProbeError)}` : ''),
  );
}

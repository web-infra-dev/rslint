import fastGlob from 'fast-glob';
import path from 'node:path';
import Mocha from 'mocha';
import * as vscode from 'vscode';

const extensionId = 'rstack.rslint';
const typescriptCodeActionProbeSource = `interface RslintCodeActionProbe {
  value: number;
}
class RslintCodeActionProbeImpl implements RslintCodeActionProbe {}
function rslintCodeActionProbeFunction() {
  return 1 + 2;
}
`;

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
 * Probe the public behavior of the quick-fix, refactor, organize-imports, and
 * fix-all providers on a controlled untitled document. This is a readiness
 * barrier only: the save assertions themselves remain single-shot.
 */
async function waitForTypeScriptCodeActionProviders(): Promise<void> {
  const document = await vscode.workspace.openTextDocument({
    content: typescriptCodeActionProbeSource,
    language: 'typescript',
  });

  const quickFixRange = document.lineAt(3).range;
  const refactorLine = document.lineAt(5);
  const refactorExpression = '1 + 2';
  const refactorStart = refactorLine.text.indexOf(refactorExpression);
  if (refactorStart < 0) {
    throw new Error('TypeScript code-action probe expression is unavailable');
  }

  const emptyRange = new vscode.Range(0, 0, 0, 0);
  const probes = new Map<
    string,
    {
      kind: vscode.CodeActionKind;
      range: vscode.Range | vscode.Selection;
    }
  >([
    [
      'quick fix',
      { kind: vscode.CodeActionKind.QuickFix, range: quickFixRange },
    ],
    [
      'refactor',
      {
        kind: vscode.CodeActionKind.Refactor,
        range: new vscode.Selection(
          refactorLine.lineNumber,
          refactorStart,
          refactorLine.lineNumber,
          refactorStart + refactorExpression.length,
        ),
      },
    ],
    [
      'organize imports',
      { kind: vscode.CodeActionKind.SourceOrganizeImports, range: emptyRange },
    ],
    [
      'fix all',
      { kind: vscode.CodeActionKind.SourceFixAll, range: emptyRange },
    ],
  ]);

  const timeoutMs = 60_000;
  const deadline = Date.now() + timeoutMs;
  let lastProbeError: unknown;

  while (Date.now() < deadline) {
    for (const [name, probe] of probes) {
      try {
        const actions = await vscode.commands.executeCommand<
          vscode.CodeAction[]
        >(
          'vscode.executeCodeActionProvider',
          document.uri,
          probe.range,
          probe.kind.value,
        );
        if (
          actions?.some(
            (action) => action.kind && probe.kind.contains(action.kind),
          )
        ) {
          probes.delete(name);
        }
      } catch (error) {
        // A provider registering during this non-mutating probe cancels it.
        // Keep waiting for the complete provider set, but fail closed on
        // timeout.
        lastProbeError = error;
      }
    }

    if (probes.size === 0) {
      return;
    }
    await new Promise((resolve) => setTimeout(resolve, 100));
  }

  throw new Error(
    `Timed out after ${timeoutMs}ms waiting for TypeScript code-action providers` +
      (probes.size > 0
        ? `; missing actions: ${[...probes.keys()].join(', ')}`
        : '') +
      (lastProbeError ? `; last probe error: ${String(lastProbeError)}` : ''),
  );
}

import fastGlob from 'fast-glob';
import fs from 'node:fs';
import path from 'node:path';
import Mocha from 'mocha';
import * as vscode from 'vscode';
import { isCodeActionCancellation } from './utils/codeActionRegistry';
import { runBeforeDeadline } from './utils/deadline';

const extensionId = 'rstack.rslint';
const workspaceMarkerFile = '.rslint-vscode-test-sandbox.json';
const startupTimeoutMs = 60_000;
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
    // This deadline starts before any extension activation. Mocha has not been
    // created yet, so every asynchronous startup step must be independently
    // bounded instead of relying on a per-test timeout that does not exist yet.
    const startupDeadline = Date.now() + startupTimeoutMs;
    verifyIsolatedWorkspace();
    const extension = vscode.extensions.getExtension(extensionId);
    if (!extension) {
      throw new Error(`Extension ${extensionId} is unavailable`);
    }

    // Extension.activate() is the public readiness contract. Await it before
    // loading any tests so cold-start saves and code-action requests cannot run
    // while the language client and workspace configuration are still starting.
    await runBeforeDeadline(
      () => extension.activate(),
      startupDeadline,
      `activation of ${extensionId}`,
    );
    await activateBuiltInCodeActionExtensions(startupDeadline);
    await waitForTypeScriptCodeActionProviders(startupDeadline);

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

function canonicalPath(filePath: string): string {
  const realPath = fs.realpathSync.native(filePath);
  return process.platform === 'win32' ? realPath.toLowerCase() : realPath;
}

function verifyIsolatedWorkspace(): void {
  const folders = vscode.workspace.workspaceFolders;
  if (folders?.length !== 1) {
    throw new Error(
      `Expected exactly one isolated workspace folder, got ${folders?.length ?? 0}`,
    );
  }

  const actualPath = canonicalPath(folders[0].uri.fsPath);
  const markerPath = path.join(actualPath, workspaceMarkerFile);
  let marker: unknown;
  try {
    marker = JSON.parse(fs.readFileSync(markerPath, 'utf8'));
  } catch (error) {
    throw new Error(
      `Could not read isolated workspace marker ${markerPath}: ${String(error)}`,
    );
  }
  if (
    typeof marker !== 'object' ||
    marker === null ||
    !('version' in marker) ||
    marker.version !== 1 ||
    !('nonce' in marker) ||
    typeof marker.nonce !== 'string' ||
    !/^[0-9a-f-]{36}$/i.test(marker.nonce) ||
    !('expectedWorkspace' in marker) ||
    typeof marker.expectedWorkspace !== 'string' ||
    !('sourceWorkspace' in marker) ||
    typeof marker.sourceWorkspace !== 'string'
  ) {
    throw new Error(`Invalid isolated workspace marker ${markerPath}`);
  }

  const expectedPath = canonicalPath(marker.expectedWorkspace);
  const sourcePath = canonicalPath(marker.sourceWorkspace);
  if (actualPath !== expectedPath) {
    throw new Error(
      `VS Code opened ${actualPath} instead of isolated workspace ${expectedPath}`,
    );
  }
  if (actualPath === sourcePath) {
    throw new Error(
      'VS Code tests must not run against tracked source fixtures',
    );
  }
}

async function activateBuiltInCodeActionExtensions(
  startupDeadline: number,
): Promise<void> {
  for (const id of ['vscode.typescript-language-features', 'vscode.git']) {
    const extension = vscode.extensions.getExtension(id);
    if (!extension) {
      throw new Error(`Built-in extension ${id} is unavailable`);
    }
    await runBeforeDeadline(
      () => extension.activate(),
      startupDeadline,
      `activation of built-in extension ${id}`,
    );
  }
}

/**
 * VS Code's built-in TypeScript extension activates lazily. Its activation
 * function returns before tsserver is ready, then four modules asynchronously
 * register the TypeScript code-action providers. A provider registration while
 * code actions on save are running cancels that save in VS Code 1.128.
 *
 * Probe the public behavior of the quick-fix, refactor, organize-imports, and
 * fix-all providers on a controlled untitled document. This eagerly resolves
 * known lazy registrations; the generic registry-quiescence sentinel still
 * runs immediately before every save. Save assertions remain single-shot.
 */
async function waitForTypeScriptCodeActionProviders(
  startupDeadline: number,
): Promise<void> {
  const document = await runBeforeDeadline(
    () =>
      vscode.workspace.openTextDocument({
        content: typescriptCodeActionProbeSource,
        language: 'typescript',
      }),
    startupDeadline,
    'the TypeScript code-action probe document to open',
  );

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

  let lastProbeError: unknown;

  while (Date.now() < startupDeadline) {
    for (const [name, probe] of probes) {
      try {
        const actions = await executeCodeActionProbeBeforeDeadline(
          document,
          probe,
          name,
          startupDeadline,
        );
        if (
          actions?.some(
            (action) => action.kind && probe.kind.contains(action.kind),
          )
        ) {
          probes.delete(name);
        }
      } catch (error) {
        // Only the known cancellation caused by an in-flight provider
        // registration is retryable. Unexpected command failures fail fast.
        if (!isCodeActionCancellation(error)) throw error;
        lastProbeError = error;
      }
    }

    if (probes.size === 0) {
      return;
    }
    await new Promise((resolve) =>
      setTimeout(
        resolve,
        Math.min(100, Math.max(0, startupDeadline - Date.now())),
      ),
    );
  }

  throw new Error(
    `Timed out after ${startupTimeoutMs}ms during VS Code test startup while waiting for TypeScript code-action providers` +
      (probes.size > 0
        ? `; missing actions: ${[...probes.keys()].join(', ')}`
        : '') +
      (lastProbeError ? `; last probe error: ${String(lastProbeError)}` : ''),
  );
}

async function executeCodeActionProbeBeforeDeadline(
  document: vscode.TextDocument,
  probe: {
    kind: vscode.CodeActionKind;
    range: vscode.Range | vscode.Selection;
  },
  name: string,
  deadline: number,
): Promise<vscode.CodeAction[] | undefined> {
  return runBeforeDeadline(
    () =>
      vscode.commands.executeCommand<vscode.CodeAction[]>(
        'vscode.executeCodeActionProvider',
        document.uri,
        probe.range,
        probe.kind.value,
      ),
    deadline,
    `the TypeScript ${name} code-action probe to settle`,
  );
}

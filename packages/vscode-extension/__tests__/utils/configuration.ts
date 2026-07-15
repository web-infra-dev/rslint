import fs from 'node:fs';
import path from 'node:path';
import { isDeepStrictEqual } from 'node:util';
import * as vscode from 'vscode';

type CodeActionsOnSave = Record<string, 'always' | 'explicit' | 'never'>;

interface EventWaiter {
  promise: Promise<boolean>;
  dispose(): void;
}

interface WorkspaceSettingsSnapshot {
  directoryPath: string;
  directoryExisted: boolean;
  filePath: string;
  content: Buffer | undefined;
}

function captureWorkspaceSettings(
  document: vscode.TextDocument,
): WorkspaceSettingsSnapshot {
  const workspaceFolder = vscode.workspace.getWorkspaceFolder(document.uri);
  if (!workspaceFolder) {
    throw new Error(`No workspace folder contains ${document.uri.toString()}`);
  }
  const directoryPath = path.join(workspaceFolder.uri.fsPath, '.vscode');
  const filePath = path.join(directoryPath, 'settings.json');
  return {
    directoryPath,
    directoryExisted: fs.existsSync(directoryPath),
    filePath,
    content: fs.existsSync(filePath) ? fs.readFileSync(filePath) : undefined,
  };
}

function restoreWorkspaceSettings(snapshot: WorkspaceSettingsSnapshot): void {
  if (snapshot.content) {
    fs.mkdirSync(snapshot.directoryPath, { recursive: true });
    fs.writeFileSync(snapshot.filePath, snapshot.content);
    if (!fs.readFileSync(snapshot.filePath).equals(snapshot.content)) {
      throw new Error(
        `Could not restore workspace settings: ${snapshot.filePath}`,
      );
    }
    return;
  }

  if (fs.existsSync(snapshot.filePath)) fs.unlinkSync(snapshot.filePath);
  if (!snapshot.directoryExisted && fs.existsSync(snapshot.directoryPath)) {
    const entries = fs.readdirSync(snapshot.directoryPath);
    if (entries.length === 0) fs.rmdirSync(snapshot.directoryPath);
  }
  if (fs.existsSync(snapshot.filePath)) {
    throw new Error(
      `Could not remove generated settings: ${snapshot.filePath}`,
    );
  }
}

function configurationEventWaiter(
  section: string,
  scope: vscode.ConfigurationScope,
  timeoutMs = 10_000,
): EventWaiter {
  let finish: (observed: boolean) => void;
  let settled = false;
  const promise = new Promise<boolean>((resolve) => {
    finish = (observed) => {
      if (settled) return;
      settled = true;
      subscription.dispose();
      clearTimeout(timer);
      resolve(observed);
    };
  });
  const subscription = vscode.workspace.onDidChangeConfiguration((event) => {
    if (event.affectsConfiguration(section, scope)) finish(true);
  });
  const timer = setTimeout(() => finish(false), timeoutMs);
  return { promise, dispose: () => finish(false) };
}

async function updateWorkspaceLanguageValue(
  document: vscode.TextDocument,
  value: CodeActionsOnSave | undefined,
): Promise<void> {
  const section = 'editor.codeActionsOnSave';
  const scope = { uri: document.uri, languageId: document.languageId };
  const configuration = vscode.workspace.getConfiguration('editor', scope);
  if (
    isDeepStrictEqual(
      configuration.inspect<CodeActionsOnSave>('codeActionsOnSave')
        ?.workspaceLanguageValue,
      value,
    )
  ) {
    return;
  }

  const waiter = configurationEventWaiter(section, scope);
  try {
    await configuration.update(
      'codeActionsOnSave',
      value,
      vscode.ConfigurationTarget.Workspace,
      true,
    );
    if (!(await waiter.promise)) {
      throw new Error(`No configuration change event received for ${section}`);
    }
  } finally {
    waiter.dispose();
  }
}

/**
 * Set the resource/language-specific on-save value and restore the exact prior
 * workspace-language value. Never writes an inherited effective value back to
 * workspace settings.
 */
export async function withCodeActionsOnSave<T>(
  document: vscode.TextDocument,
  value: CodeActionsOnSave,
  callback: () => Promise<T>,
): Promise<T> {
  const scope = { uri: document.uri, languageId: document.languageId };
  const configuration = vscode.workspace.getConfiguration('editor', scope);
  const inspection =
    configuration.inspect<CodeActionsOnSave>('codeActionsOnSave');
  if (!inspection) {
    throw new Error('editor.codeActionsOnSave is unavailable');
  }

  const previousEffective =
    configuration.get<CodeActionsOnSave>('codeActionsOnSave');
  const previousWorkspaceLanguageValue = inspection.workspaceLanguageValue;
  const changed = !isDeepStrictEqual(previousEffective, value);
  const settingsSnapshot = changed
    ? captureWorkspaceSettings(document)
    : undefined;

  let result: T | undefined;
  let callbackCompleted = false;
  let callbackError: unknown;
  try {
    if (changed) {
      await updateWorkspaceLanguageValue(document, value);
      const effective = vscode.workspace
        .getConfiguration('editor', scope)
        .get<CodeActionsOnSave>('codeActionsOnSave');
      if (!isDeepStrictEqual(effective, value)) {
        throw new Error(
          'editor.codeActionsOnSave did not resolve to the requested test value',
        );
      }
    }
    result = await callback();
    callbackCompleted = true;
  } catch (error) {
    callbackError = error;
  }

  const restoreErrors: unknown[] = [];
  if (changed) {
    try {
      await updateWorkspaceLanguageValue(
        document,
        previousWorkspaceLanguageValue,
      );
      const restored = vscode.workspace
        .getConfiguration('editor', scope)
        .get<CodeActionsOnSave>('codeActionsOnSave');
      if (!isDeepStrictEqual(restored, previousEffective)) {
        throw new Error(
          'editor.codeActionsOnSave was not restored to its prior effective value',
        );
      }
    } catch (error) {
      restoreErrors.push(error);
    }
    if (!settingsSnapshot) {
      restoreErrors.push(new Error('Workspace settings snapshot is missing'));
    } else {
      try {
        restoreWorkspaceSettings(settingsSnapshot);
      } catch (error) {
        restoreErrors.push(error);
      }
    }
  }

  const restoreError =
    restoreErrors.length > 1
      ? new AggregateError(
          restoreErrors,
          'Multiple editor.codeActionsOnSave restoration steps failed',
        )
      : restoreErrors[0];

  if (callbackError && restoreError) {
    throw new AggregateError(
      [callbackError, restoreError],
      'Test callback and editor.codeActionsOnSave restoration both failed',
    );
  }
  if (callbackError) throw callbackError;
  if (restoreError) throw restoreError;
  if (!callbackCompleted) {
    throw new Error('editor.codeActionsOnSave callback did not complete');
  }
  return result as T;
}

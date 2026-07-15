import { randomUUID } from 'node:crypto';
import * as vscode from 'vscode';

const defaultQuietWindowMs = 750;
const defaultSuccessfulWindows = 2;
const defaultTimeoutMs = 60_000;
const defaultRetryDelayMs = 25;

export interface CodeActionRegistryProbeOptions {
  quietWindowMs?: number;
  consecutiveSuccessfulWindows?: number;
  timeoutMs?: number;
  retryDelayMs?: number;
  /** Test-only hook. It runs after the probe provider receives the request. */
  onAttemptStarted?: (attempt: number, probeUri: vscode.Uri) => void;
}

export interface CodeActionRegistryProbeResult {
  attempts: number;
  interruptedWindows: number;
}

interface ProbeLoopOptions {
  consecutiveSuccessfulWindows: number;
  timeoutMs: number;
  retryDelayMs: number;
  description: string;
}

interface SaveableDocument {
  save(): Thenable<boolean>;
}

function delay(timeoutMs: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, timeoutMs));
}

export function isCodeActionCancellation(error: unknown): boolean {
  return (
    error instanceof vscode.CancellationError ||
    (error instanceof Error &&
      error.name === 'Canceled' &&
      error.message === 'Canceled')
  );
}

/**
 * Require N consecutive successful readiness windows. A failed or cancelled
 * window resets the count; timeout rejects instead of accepting a partial run.
 * The executor is injectable so the fail-closed state machine can be tested
 * without relying on a particular VS Code release's cancellation behavior.
 */
export async function waitForConsecutiveSuccessfulProbeWindows(
  executeWindow: (attempt: number) => Promise<boolean>,
  options: ProbeLoopOptions,
): Promise<CodeActionRegistryProbeResult> {
  if (options.consecutiveSuccessfulWindows < 1) {
    throw new Error('consecutiveSuccessfulWindows must be at least 1');
  }
  if (options.timeoutMs < 1) {
    throw new Error('timeoutMs must be at least 1');
  }

  const deadline = Date.now() + options.timeoutMs;
  let attempts = 0;
  let interruptedWindows = 0;
  let successfulWindows = 0;

  const timeoutError = (): Error =>
    new Error(
      `Timed out after ${options.timeoutMs}ms waiting for ${options.description}; ` +
        `attempts=${attempts}, interruptedWindows=${interruptedWindows}`,
    );

  while (Date.now() < deadline) {
    attempts += 1;
    let timer: ReturnType<typeof setTimeout> | undefined;
    const remainingMs = Math.max(1, deadline - Date.now());
    const hardTimeout = new Promise<never>((_resolve, reject) => {
      timer = setTimeout(() => reject(timeoutError()), remainingMs);
    });
    let completed: boolean;
    try {
      completed = await Promise.race([executeWindow(attempts), hardTimeout]);
    } finally {
      if (timer) clearTimeout(timer);
    }
    if (Date.now() >= deadline) throw timeoutError();

    if (completed) {
      successfulWindows += 1;
      if (successfulWindows >= options.consecutiveSuccessfulWindows) {
        return { attempts, interruptedWindows };
      }
      continue;
    }

    interruptedWindows += 1;
    successfulWindows = 0;
    if (options.retryDelayMs > 0) {
      await delay(
        Math.min(options.retryDelayMs, Math.max(0, deadline - Date.now())),
      );
    }
  }

  throw timeoutError();
}

/**
 * Public-API sentinel for VS Code's code-action provider registry.
 *
 * VS Code versions affected by microsoft/vscode's filtered-provider comparison
 * bug cancel an in-flight source action whenever any provider is registered or
 * disposed. Two providers match this private URI scheme, but only one matches
 * the requested kind. That makes an unrelated registry mutation cancel this
 * harmless request in exactly the same way it cancels code actions on save.
 *
 * Providers remain registered for the lifetime of the probe so their own
 * disposal cannot race the real save. They match only a randomized private URI
 * scheme and therefore never participate in normal test documents.
 */
export class CodeActionRegistryProbe implements vscode.Disposable {
  private readonly probeKind: vscode.CodeActionKind;
  private readonly documentPromise: Thenable<vscode.TextDocument>;
  private readonly disposables: vscode.Disposable[];
  private queue: Promise<void> = Promise.resolve();
  private disposed = false;
  private activeQuietWindowMs = defaultQuietWindowMs;
  private activeAttempt = 0;
  private activeAttemptHook:
    | ((attempt: number, probeUri: vscode.Uri) => void)
    | undefined;

  constructor() {
    const id = randomUUID().replaceAll('-', '');
    const scheme = `rslint-registry-probe-${id}`;
    const selector: vscode.DocumentSelector = { scheme };
    this.probeKind = vscode.CodeActionKind.Source.append(
      `rslintTest.registryProbe.${id}`,
    );
    const excludedKind = vscode.CodeActionKind.QuickFix.append(
      `rslintTest.registryProbe.${id}`,
    );

    const contentProvider =
      vscode.workspace.registerTextDocumentContentProvider(scheme, {
        provideTextDocumentContent: () => '// registry probe\n',
      });
    const includedProvider = vscode.languages.registerCodeActionsProvider(
      selector,
      {
        provideCodeActions: (document, _range, _context, token) =>
          this.provideProbeAction(document.uri, token),
      },
      { providedCodeActionKinds: [this.probeKind] },
    );
    // This provider deliberately matches the document but not the requested
    // kind. It exposes VS Code's filtered-vs-unfiltered registry comparison.
    const excludedProvider = vscode.languages.registerCodeActionsProvider(
      selector,
      { provideCodeActions: () => [] },
      { providedCodeActionKinds: [excludedKind] },
    );

    this.disposables = [excludedProvider, includedProvider, contentProvider];
    this.documentPromise = vscode.workspace.openTextDocument(
      vscode.Uri.parse(`${scheme}:/probe.ts`),
    );
  }

  wait(
    options: CodeActionRegistryProbeOptions = {},
  ): Promise<CodeActionRegistryProbeResult> {
    if (this.disposed) {
      return Promise.reject(new Error('CodeActionRegistryProbe is disposed'));
    }

    const execution = this.queue.then(() => this.waitUnqueued(options));
    this.queue = execution.then(
      () => undefined,
      () => undefined,
    );
    return execution;
  }

  dispose(): void {
    if (this.disposed) return;
    this.disposed = true;
    for (const disposable of this.disposables) disposable.dispose();
  }

  private async waitUnqueued(
    options: CodeActionRegistryProbeOptions,
  ): Promise<CodeActionRegistryProbeResult> {
    if (this.disposed) {
      throw new Error('CodeActionRegistryProbe is disposed');
    }

    const document = await this.documentPromise;
    const quietWindowMs = options.quietWindowMs ?? defaultQuietWindowMs;
    if (quietWindowMs < 1) {
      throw new Error('quietWindowMs must be at least 1');
    }

    this.activeQuietWindowMs = quietWindowMs;
    this.activeAttemptHook = options.onAttemptStarted;
    try {
      return await waitForConsecutiveSuccessfulProbeWindows(
        async (attempt) => {
          this.activeAttempt = attempt;
          let actions: vscode.CodeAction[] | undefined;
          try {
            actions = await vscode.commands.executeCommand<vscode.CodeAction[]>(
              'vscode.executeCodeActionProvider',
              document.uri,
              new vscode.Range(0, 0, 0, 0),
              this.probeKind.value,
            );
          } catch (error) {
            if (isCodeActionCancellation(error)) return false;
            throw error;
          }
          return Boolean(
            actions?.some(
              (action) => action.kind?.value === this.probeKind.value,
            ),
          );
        },
        {
          consecutiveSuccessfulWindows:
            options.consecutiveSuccessfulWindows ?? defaultSuccessfulWindows,
          timeoutMs: options.timeoutMs ?? defaultTimeoutMs,
          retryDelayMs: options.retryDelayMs ?? defaultRetryDelayMs,
          description: 'the VS Code code-action registry to become quiescent',
        },
      );
    } finally {
      this.activeAttempt = 0;
      this.activeAttemptHook = undefined;
    }
  }

  private provideProbeAction(
    probeUri: vscode.Uri,
    token: vscode.CancellationToken,
  ): Promise<vscode.CodeAction[]> {
    const attempt = this.activeAttempt;
    this.activeAttemptHook?.(attempt, probeUri);

    return new Promise<vscode.CodeAction[]>((resolve, reject) => {
      let settled = false;
      let cancellation: vscode.Disposable | undefined;
      const finish = (
        actions: vscode.CodeAction[] | undefined,
        error?: unknown,
      ): void => {
        if (settled) return;
        settled = true;
        clearTimeout(timer);
        cancellation?.dispose();
        if (error) reject(error);
        else resolve(actions ?? []);
      };
      const timer = setTimeout(() => {
        finish([
          new vscode.CodeAction('rslint test registry probe', this.probeKind),
        ]);
      }, this.activeQuietWindowMs);
      cancellation = token.onCancellationRequested(() =>
        finish(undefined, new vscode.CancellationError()),
      );
      if (token.isCancellationRequested) {
        finish(undefined, new vscode.CancellationError());
      }
    });
  }
}

let sharedProbe: CodeActionRegistryProbe | undefined;

export function waitForCodeActionRegistryQuiescence(): Promise<CodeActionRegistryProbeResult> {
  sharedProbe ??= new CodeActionRegistryProbe();
  return sharedProbe.wait();
}

/**
 * Wait for readiness, then invoke the real save exactly once. A false result is
 * surfaced as a failure and is never retried, preventing a cancelled first save
 * from being hidden by a successful second save.
 */
export async function saveDocumentOnce(
  document: SaveableDocument,
  failureMessage: string,
  waitForReadiness: () => Promise<unknown> = waitForCodeActionRegistryQuiescence,
): Promise<void> {
  await waitForReadiness();
  const saved = await document.save();
  if (!saved) {
    throw new Error(
      `${failureMessage}: TextDocument.save() returned false; the real save was not retried`,
    );
  }
}

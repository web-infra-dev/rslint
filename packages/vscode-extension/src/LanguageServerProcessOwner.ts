import { spawn, type ChildProcessWithoutNullStreams } from 'node:child_process';

const GRACEFUL_EXIT_TIMEOUT_MS = 500;
const FORCED_EXIT_TIMEOUT_MS = 1_500;

function hasExited(process: ChildProcessWithoutNullStreams): boolean {
  return process.exitCode !== null || process.signalCode !== null;
}

async function waitForClose(
  closed: Promise<void>,
  timeoutMs: number,
): Promise<boolean> {
  return new Promise<boolean>((resolve) => {
    let settled = false;
    const finish = (closedInTime: boolean): void => {
      if (settled) return;
      settled = true;
      clearTimeout(timer);
      resolve(closedInTime);
    };
    const timer = setTimeout(() => {
      finish(false);
    }, timeoutMs);
    void closed.then(() => {
      finish(true);
    });
  });
}

async function terminateProcess(
  process: ChildProcessWithoutNullStreams,
  closed: Promise<void>,
): Promise<void> {
  if (!hasExited(process)) {
    try {
      process.kill('SIGTERM');
    } catch {
      // The close check below distinguishes a process that raced to completion
      // from one that still needs a forced termination attempt.
    }
  }
  if (await waitForClose(closed, GRACEFUL_EXIT_TIMEOUT_MS)) return;

  if (!hasExited(process)) {
    try {
      process.kill('SIGKILL');
    } catch {
      // Report only if the transport remains open after the bounded wait.
    }
  }
  if (await waitForClose(closed, FORCED_EXIT_TIMEOUT_MS)) return;

  throw new Error(
    `language server process ${String(process.pid)} did not close its transports after SIGKILL`,
  );
}

/**
 * Owns every native server child created for one workspace runtime, including
 * children created by vscode-languageclient's automatic restart path.
 */
export class LanguageServerProcessOwner {
  private readonly processes = new Map<
    ChildProcessWithoutNullStreams,
    Promise<void>
  >();
  private startTail: Promise<void> = Promise.resolve();
  private closePromise: Promise<void> | undefined;
  private closing = false;

  public constructor(
    private readonly command: string,
    private readonly args: readonly string[],
    private readonly cwd: string,
    private readonly env?: NodeJS.ProcessEnv,
  ) {}

  public beginClose(): void {
    this.closing = true;
  }

  public async start(): Promise<ChildProcessWithoutNullStreams> {
    const operation = this.startTail.then(
      async () => this.startImpl(),
      async () => this.startImpl(),
    );
    this.startTail = operation.then(
      () => undefined,
      () => undefined,
    );
    const child = await operation;
    return child;
  }

  private async startImpl(): Promise<ChildProcessWithoutNullStreams> {
    if (this.closing) {
      throw new Error('language server process owner is closing');
    }

    // A transport can close before its process exits. Automatic restart must
    // never create a second native generation while that old child is alive.
    await this.terminateTrackedProcesses();
    if (this.closing) {
      throw new Error('language server process owner is closing');
    }

    const child = spawn(this.command, [...this.args], {
      cwd: this.cwd,
      env: this.env,
      stdio: ['pipe', 'pipe', 'pipe'],
    });
    const closed = new Promise<void>((resolve) => {
      child.once('close', () => {
        resolve();
      });
    });
    this.processes.set(child, closed);
    void closed.then(() => {
      this.processes.delete(child);
    });

    let spawned = false;
    let resolveStarted!: () => void;
    let rejectStarted!: (error: unknown) => void;
    const started = new Promise<void>((resolve, reject) => {
      resolveStarted = resolve;
      rejectStarted = reject;
    });

    child.once('spawn', () => {
      spawned = true;
      if (this.closing) {
        rejectStarted(new Error('language server process owner is closing'));
      } else {
        resolveStarted();
      }
    });
    child.on('error', (error) => {
      if (spawned) return;
      rejectStarted(error);
    });

    await started;
    return child;
  }

  public async close(): Promise<void> {
    await (this.closePromise ??= this.closeImpl());
  }

  private async closeImpl(): Promise<void> {
    this.beginClose();
    await this.startTail;
    await this.terminateTrackedProcesses();
  }

  private async terminateTrackedProcesses(): Promise<void> {
    const results = await Promise.allSettled(
      [...this.processes].map(async ([process, closed]) => {
        await terminateProcess(process, closed);
      }),
    );
    const errors: unknown[] = [];
    for (const result of results) {
      if (result.status === 'rejected') {
        const reason: unknown = result.reason;
        errors.push(reason);
      }
    }
    if (errors.length > 0) {
      throw new AggregateError(
        errors,
        'failed to terminate language server processes',
      );
    }
  }
}

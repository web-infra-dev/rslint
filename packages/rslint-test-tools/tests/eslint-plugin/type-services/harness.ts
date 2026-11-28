import { afterAll, beforeAll } from 'rstack/test';
import { execFileSync } from 'node:child_process';
import { mkdtempSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { resolve } from 'node:path';
import { Worker } from 'node:worker_threads';
import { TYPESCRIPT_COMPILER_DATA } from '../../../../rslint/src/languages/js/type-service-compiler-data.generated.js';
import { cleanupTempDir } from '../../helpers/cleanup-temp-dir.js';

export const root = resolve(import.meta.dirname, '../../../../..');
export const enums = TYPESCRIPT_COMPILER_DATA;

export class Host {
  mailbox: SharedArrayBuffer;
  control: Int32Array;
  worker: Worker;
  private workerExited = false;
  private workerExitCode: number | null = null;
  private workerError: unknown = undefined;
  private reportedWorkerFailure = false;
  private reportedChildFailure = false;
  private closePromise: Promise<void> | null = null;

  constructor(binary: string, mailboxBytes = 16 << 20) {
    if (!Number.isInteger(mailboxBytes) || mailboxBytes < 256) {
      throw new RangeError('mailboxBytes must be an integer >= 256');
    }
    this.mailbox = new SharedArrayBuffer(mailboxBytes);
    this.control = new Int32Array(this.mailbox, 0, 2);
    this.worker = new Worker(
      resolve(
        root,
        'packages/rslint-test-tools/tests/eslint-plugin/type-service-broker.mjs',
      ),
      { workerData: { binary, cwd: root, mailbox: this.mailbox } },
    );
    this.worker.once('exit', (code) => {
      this.workerExited = true;
      this.workerExitCode = code;
    });
    this.worker.on('error', (error) => {
      this.workerError ??=
        error ?? new Error('type service broker worker error');
    });
  }

  request(request: object): any {
    if (this.workerError !== undefined) {
      this.reportedWorkerFailure = true;
      throw this.workerError;
    }
    if (this.workerExited) {
      this.reportedWorkerFailure = true;
      throw new Error(
        `type service broker worker already exited with code ${String(this.workerExitCode)}`,
      );
    }
    Atomics.store(this.control, 0, 1);
    try {
      this.worker.postMessage(request);
    } catch (error) {
      Atomics.store(this.control, 0, 0);
      throw error;
    }
    if (Atomics.wait(this.control, 0, 1, 60_000) === 'timed-out') {
      this.reportedChildFailure = true;
      throw new Error('type service host timed out');
    }
    const response = JSON.parse(
      Buffer.from(this.mailbox, 8, Atomics.load(this.control, 1)).toString(),
    ) as {
      result?: unknown;
      error?: string;
      childFatal?: boolean;
      workerFatal?: boolean;
    };
    if (response.error) {
      if (response.workerFatal) {
        this.workerExited = true;
        this.workerExitCode ??= 1;
        this.reportedWorkerFailure = true;
        if (this.workerError !== undefined) {
          throw this.workerError;
        }
      } else if (response.childFatal) {
        this.reportedChildFailure = true;
      }
      throw new Error(response.error);
    }
    this.reportedChildFailure = false;
    return response.result;
  }

  source = (
    scope: { generation: number },
    file: string,
    metadataOnly = false,
  ) => {
    const reply = this.request({
      method: 'source',
      generation: scope.generation,
      file,
      metadataOnly,
    });
    return reply
      ? {
          fileName: reply.fileName,
          path: reply.path,
          tree: reply.tree ? Buffer.from(reply.tree, 'base64') : undefined,
        }
      : undefined;
  };

  async close(): Promise<void> {
    if (this.closePromise) {
      return this.closePromise;
    }
    this.closePromise = (async () => {
      // Track worker failures and Go child failures independently so a caught
      // recoverable error (or even an earlier child failure) never suppresses
      // an independent broker worker crash during teardown.
      const hadReportedWorkerFailure = this.reportedWorkerFailure;
      const hadReportedChildFailure = this.reportedChildFailure;
      // Flush any queued libuv worker `error`/`exit` events that occurred while
      // the main thread was blocked in synchronous work before `close()`.
      await new Promise<void>((resolve) => setImmediate(resolve));
      const unexpectedExitError = (code: number | null, phase = '') =>
        new Error(
          `type service broker worker exited unexpectedly with code ${String(code)}${phase}`,
        );
      let workerCloseError: unknown =
        this.workerError ??
        (this.workerExited
          ? unexpectedExitError(this.workerExitCode)
          : undefined);
      let childCloseError: unknown;
      if (
        !hadReportedWorkerFailure &&
        !hadReportedChildFailure &&
        !this.workerExited &&
        this.workerError === undefined
      ) {
        try {
          this.request({ method: 'close' });
        } catch (error) {
          if (this.workerError !== undefined || this.workerExited) {
            workerCloseError ??=
              this.workerError ?? unexpectedExitError(this.workerExitCode);
          } else {
            childCloseError ??= error;
          }
        }
        workerCloseError ??=
          this.workerError ??
          (this.workerExited
            ? unexpectedExitError(this.workerExitCode)
            : undefined);
      }
      try {
        if (!this.workerExited && this.workerError === undefined) {
          await new Promise<void>((done) => {
            let stopTimer: ReturnType<typeof setTimeout> | undefined;
            const cleanup = () => {
              clearTimeout(stopTimer);
              this.worker.off('message', onMessage);
              this.worker.off('error', onError);
              this.worker.off('exit', onFinish);
            };
            const onMessage = (message: {
              stopped?: boolean;
              error?: string;
            }) => {
              if (message?.stopped) {
                if (message.error) {
                  childCloseError ??= new Error(message.error);
                }
                workerCloseError ??= this.workerError;
                cleanup();
                done();
              }
            };
            const onError = (error: unknown) => {
              workerCloseError ??=
                error ?? new Error('type service broker worker error');
              cleanup();
              done();
            };
            const onFinish = (code: number) => {
              workerCloseError ??=
                this.workerError ??
                unexpectedExitError(code, ' before stopping child');
              cleanup();
              done();
            };
            this.worker.on('message', onMessage);
            this.worker.once('error', onError);
            this.worker.once('exit', onFinish);
            if (this.workerExited || this.workerError !== undefined) {
              workerCloseError ??=
                this.workerError ??
                unexpectedExitError(
                  this.workerExitCode,
                  ' before stopping child',
                );
              cleanup();
              done();
              return;
            }
            stopTimer = setTimeout(() => {
              workerCloseError ??= new Error(
                'timed out waiting for type service broker worker to stop',
              );
              cleanup();
              done();
            }, 10_000);
            stopTimer.unref();
            this.worker.postMessage({ method: 'stop' });
          });
        }
      } finally {
        await this.worker.terminate();
      }
      workerCloseError ??= this.workerError;
      if (workerCloseError !== undefined && !hadReportedWorkerFailure) {
        throw workerCloseError;
      }
      if (
        childCloseError !== undefined &&
        !hadReportedChildFailure &&
        !hadReportedWorkerFailure
      ) {
        throw childCloseError;
      }
    })();
    return this.closePromise;
  }
}

export function useTypeServiceHarness() {
  let temp = '';
  let binary = '';
  beforeAll(() => {
    temp = mkdtempSync(resolve(tmpdir(), 'rslint-types-'));
    const shared = process.env.RSLINT_TYPE_SERVICE_SERVER;
    if (shared) {
      binary = shared;
      return;
    }
    binary = resolve(
      temp,
      process.platform === 'win32' ? 'types.exe' : 'types',
    );
    execFileSync(
      'go',
      ['build', '-o', binary, './internal/typeservice/testdata/server'],
      { cwd: root, stdio: 'pipe' },
    );
  });
  afterAll(async () => {
    await cleanupTempDir(temp);
  });
  return {
    get temp() {
      return temp;
    },
    createHost: (mailboxBytes?: number) => new Host(binary, mailboxBytes),
  };
}

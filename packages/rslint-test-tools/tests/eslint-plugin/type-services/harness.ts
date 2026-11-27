import { afterAll, beforeAll } from 'rstack/test';
import { execFileSync } from 'node:child_process';
import { mkdtempSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { resolve } from 'node:path';
import { Worker } from 'node:worker_threads';
import ts from 'typescript';
import { typeServiceCompilerData } from '../../../../rslint/src/languages/js/type-service-compiler-data.js';
import { cleanupTempDir } from '../../helpers/cleanup-temp-dir.js';

export const root = resolve(import.meta.dirname, '../../../../..');
export const enums = typeServiceCompilerData(ts);

export class Host {
  mailbox = new SharedArrayBuffer(16 << 20);
  control = new Int32Array(this.mailbox, 0, 2);
  worker: Worker;
  private workerExited = false;
  private workerError: unknown = null;
  private requestFailed = false;
  private closePromise: Promise<void> | null = null;

  constructor(binary: string) {
    this.worker = new Worker(
      resolve(
        root,
        'packages/rslint-test-tools/tests/eslint-plugin/type-service-broker.mjs',
      ),
      { workerData: { binary, cwd: root, mailbox: this.mailbox } },
    );
    this.worker.once('exit', () => {
      this.workerExited = true;
    });
    this.worker.on('error', (error) => {
      this.workerError ??= error;
    });
  }

  request(request: object): any {
    try {
      if (this.workerError !== null) {
        throw this.workerError;
      }
      if (this.workerExited) {
        throw new Error('type service broker worker already exited');
      }
      Atomics.store(this.control, 0, 0);
      this.worker.postMessage(request);
      if (Atomics.wait(this.control, 0, 0, 60_000) === 'timed-out')
        throw new Error('type service host timed out');
      const response = JSON.parse(
        Buffer.from(this.mailbox, 8, Atomics.load(this.control, 1)).toString(),
      );
      if (response.error) throw new Error(response.error);
      return response.result;
    } catch (error) {
      this.requestFailed = true;
      throw error;
    }
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
      const hadPriorFailure =
        this.requestFailed || this.workerExited || this.workerError !== null;
      let closeError: unknown;
      if (!this.workerExited && this.workerError === null) {
        try {
          this.request({ method: 'close' });
        } catch (error) {
          closeError = error;
        }
      }
      try {
        if (!this.workerExited) {
          await new Promise<void>((done) => {
            const cleanup = () => {
              this.worker.off('message', onMessage);
              this.worker.off('error', onError);
              this.worker.off('exit', onFinish);
            };
            const onMessage = (message: { stopped?: boolean }) => {
              if (message?.stopped) {
                cleanup();
                done();
              }
            };
            const onError = (error: unknown) => {
              closeError ??= error;
              cleanup();
              done();
            };
            const onFinish = () => {
              cleanup();
              done();
            };
            this.worker.on('message', onMessage);
            this.worker.once('error', onError);
            this.worker.once('exit', onFinish);
            if (this.workerExited) {
              cleanup();
              done();
              return;
            }
            this.worker.postMessage({ method: 'stop' });
          });
        }
      } finally {
        await this.worker.terminate();
      }
      if (!hadPriorFailure && closeError !== undefined) {
        throw closeError;
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
    createHost: () => new Host(binary),
  };
}

import { afterAll, beforeAll } from 'rstack/test';
import { execFileSync } from 'node:child_process';
import { mkdtempSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { resolve } from 'node:path';
import { Worker } from 'node:worker_threads';
import ts from 'typescript';
import { typeServiceCompilerData } from '../../../../rslint/src/languages/js/type-service-compiler-data.js';

export const root = resolve(import.meta.dirname, '../../../../..');
export const enums = typeServiceCompilerData(ts);

export class Host {
  mailbox = new SharedArrayBuffer(16 << 20);
  control = new Int32Array(this.mailbox, 0, 2);
  worker: Worker;
  constructor(binary: string) {
    this.worker = new Worker(
      resolve(
        root,
        'packages/rslint-test-tools/tests/eslint-plugin/type-service-broker.mjs',
      ),
      { workerData: { binary, cwd: root, mailbox: this.mailbox } },
    );
  }
  request(request: object): any {
    Atomics.store(this.control, 0, 0);
    this.worker.postMessage(request);
    if (Atomics.wait(this.control, 0, 0, 60_000) === 'timed-out')
      throw new Error('type service host timed out');
    const response = JSON.parse(
      Buffer.from(this.mailbox, 8, Atomics.load(this.control, 1)).toString(),
    );
    if (response.error) throw new Error(response.error);
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
  async close() {
    this.request({ method: 'close' });
    const stopped = new Promise<void>((done, reject) => {
      this.worker.once('message', () => done());
      this.worker.once('error', reject);
    });
    this.worker.postMessage({ method: 'stop' });
    await stopped;
    await this.worker.terminate();
  }
}

export function useTypeServiceHarness() {
  let temp = '';
  let binary = '';
  beforeAll(() => {
    temp = mkdtempSync(resolve(tmpdir(), 'rslint-types-'));
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
  afterAll(() => rmSync(temp, { recursive: true, force: true }));
  return {
    get temp() {
      return temp;
    },
    createHost: () => new Host(binary),
  };
}

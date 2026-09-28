import type { ChildProcess } from 'node:child_process';
import { Socket } from 'node:net';
import { RSLintService } from '../service/service.js';
import { spawnIpcProcess, type IpcClient } from '../ipc/index.js';
import { resolveRslintBinary } from './resolve-binary.js';
import type {
  RslintServiceInterface,
  RSlintOptions,
  InboundRequestHandler,
  LintOptions,
  LintResponse,
} from '../types.js';

/**
 * Node.js implementation of RslintService using child processes
 */
export class NodeRslintService implements RslintServiceInterface {
  private readonly process: ChildProcess;
  private readonly client: IpcClient;
  private dead = false;
  private activeRequests = 0;
  private activeInboundRequests = 0;
  private inboundHandler: InboundRequestHandler | null = null;

  constructor(options: RSlintOptions = {}) {
    const { child, client } = spawnIpcProcess({
      binPath: options.rslintPath || resolveRslintBinary(),
      goArgs: ['--api'],
      cwd: options.workingDirectory || undefined,
      // API plugin requests retain their existing files[].text wire contract.
      // They share framing and lifecycle with CLI without allocating unused
      // source storage or requiring a new capability from existing API hosts.
      sharedSources: false,
    });
    this.process = child;
    this.client = client;
    child.once('error', this.onProcessError);
    child.once('exit', this.onProcessExit);
    child.once('close', () => {
      child.off('error', this.onProcessError);
      child.off('exit', this.onProcessExit);
    });
    client.setInboundHandler(async (message) => {
      this.activeInboundRequests++;
      this.updateLoopActivity();
      try {
        if (!this.inboundHandler) {
          throw new Error(
            `no inbound handler registered (kind=${message.kind})`,
          );
        }
        return await this.inboundHandler({
          id: message.id,
          kind: message.kind,
          data: message.data,
        });
      } finally {
        this.activeInboundRequests--;
        this.updateLoopActivity();
      }
    });
    client.start();
    // EOF can precede process exit, including while a reverse handler is
    // still awaiting user code. A closed channel cannot serve more requests;
    // release this backend's child without waiting for those handlers.
    void client.done.then(() => {
      this.terminate();
    });
    // A resident service must not keep a one-shot Node script alive merely
    // because the caller omitted close(). Active requests re-reference it.
    this.updateLoopActivity();
  }

  private setLoopActive(active: boolean): void {
    if (active) this.process.ref();
    else this.process.unref();
    for (const stream of [this.process.stdin, this.process.stdout]) {
      if (stream instanceof Socket) {
        if (active) stream.ref();
        else stream.unref();
      }
    }
  }

  private updateLoopActivity(): void {
    this.setLoopActive(
      !this.dead &&
        !this.client.isClosed &&
        (this.activeRequests > 0 || this.activeInboundRequests > 0),
    );
  }

  /** Install the API's request-scoped reverse handler. IPC owns dispatch. */
  setInboundHandler(handler: InboundRequestHandler | null): void {
    this.inboundHandler = handler;
  }

  async sendMessage(kind: string, data: any): Promise<any> {
    if (this.dead) {
      throw new Error('rslint service is no longer running');
    }
    this.activeRequests++;
    this.updateLoopActivity();
    try {
      const response = await this.client.sendRequest(kind, data);
      return response.data;
    } catch (error) {
      // A peer can close stdout before its exit acknowledgement is delivered.
      // This best-effort shutdown request alone tolerates transport closure;
      // all ordinary in-flight requests still reject through IpcClient.
      if (
        kind === 'exit' &&
        this.client.isClosed &&
        error === (await this.client.done)
      )
        return null;
      throw error;
    } finally {
      this.activeRequests--;
      this.updateLoopActivity();
    }
  }

  private finish(error: Error): void {
    if (this.dead) return;
    this.dead = true;
    this.inboundHandler = null;
    this.client.close(error);
    this.updateLoopActivity();
  }

  private readonly onProcessError = (error: Error): void => {
    this.finish(new Error(`rslint process error: ${error.message}`));
  };

  private readonly onProcessExit = (
    code: number | null,
    signal: NodeJS.Signals | null,
  ): void => {
    this.finish(
      new Error(
        `rslint process exited unexpectedly (code=${code}, signal=${signal})`,
      ),
    );
  };

  terminate(): void {
    this.finish(new Error('rslint service terminated'));
    if (
      !this.process.killed &&
      this.process.exitCode === null &&
      this.process.signalCode === null
    ) {
      this.process.stdin?.end();
      this.process.kill();
    }
  }
}

/**
 * One-shot convenience: spin up a Node-backed service, run a single lint
 * request, then tear it down. This is an internal/tooling surface (the
 * rule-tester and the ESLint-plugin conformance harnesses) reached via the
 * `@rslint/core/internal` subpath — the package root deliberately exposes only
 * the high-level `Rslint` class as its linting surface, not this low-level engine.
 */
export async function lint(options: LintOptions): Promise<LintResponse> {
  const service = new RSLintService(
    new NodeRslintService({
      workingDirectory: options.workingDirectory,
    }),
  );
  try {
    return await service.lint(options);
  } finally {
    await service.close();
  }
}

export type { LintOptions, LintResponse, Diagnostic } from '../types.js';

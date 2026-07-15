import { describe, expect, test } from '@rstest/core';

import { RSLintService } from '../src/service/service.js';
import type {
  InboundRequestHandler,
  IpcMessage,
  RslintServiceInterface,
} from '../src/types.js';

class ReverseLintBackend implements RslintServiceInterface {
  inbound: InboundRequestHandler | null = null;
  lintPayloads: unknown[] = [];

  setInboundHandler(handler: InboundRequestHandler | null): void {
    this.inbound = handler;
  }

  async sendMessage(kind: string, data: any): Promise<any> {
    if (kind === 'handshake')
      return {
        version: '3.0.0',
        ok: true,
        capabilities: ['reversePluginLint'],
      };
    if (kind === 'exit') return {};
    if (kind !== 'lint') throw new Error(`unexpected kind ${kind}`);

    this.lintPayloads.push(data);
    // Yield before dispatching the reverse request. Without service-level
    // serialization, a concurrent lint could replace the active handler here.
    await new Promise((resolve) => setTimeout(resolve, 5));
    if (!this.inbound) throw new Error('missing inbound handler');
    return this.inbound({
      id: 100,
      kind: 'pluginLint',
      data: { token: data.files[0] },
    } satisfies IpcMessage);
  }

  terminate(): void {
    // No process is owned by this in-memory backend.
  }
}

class HangingExitBackend extends ReverseLintBackend {
  terminated = false;

  override async sendMessage(kind: string, data: any): Promise<any> {
    if (kind === 'exit')
      return new Promise(() => {
        // Deliberately unresolved to exercise forced shutdown.
      });
    return super.sendMessage(kind, data);
  }

  override terminate(): void {
    this.terminated = true;
  }
}

class HangingLintBackend extends ReverseLintBackend {
  terminated = false;
  exitRequests = 0;
  readonly lintStarted: Promise<void>;
  private markLintStarted!: () => void;

  constructor() {
    super();
    this.lintStarted = new Promise((resolve) => {
      this.markLintStarted = resolve;
    });
  }

  override async sendMessage(kind: string, data: any): Promise<any> {
    if (kind === 'lint') {
      this.markLintStarted();
      return new Promise(() => {
        // Deliberately unresolved to keep the lint request active.
      });
    }
    if (kind === 'exit') {
      this.exitRequests++;
      return {};
    }
    return super.sendMessage(kind, data);
  }

  override terminate(): void {
    this.terminated = true;
  }
}

class ReverseConfigBackend extends ReverseLintBackend {
  handshakeCapabilities: string[] = [];

  override async sendMessage(kind: string, data: any): Promise<any> {
    if (kind === 'handshake') {
      this.handshakeCapabilities = [...(data.capabilities ?? [])];
      return {
        version: '3.0.0',
        ok: true,
        capabilities: [],
      };
    }
    if (kind === 'lint') {
      if (!this.inbound) throw new Error('missing inbound handler');
      const loaded = await this.inbound({
        id: 101,
        kind: 'loadConfigs',
        data: { transactionId: 'tx-service', candidates: ['candidate'] },
      });
      const evaluated = await this.inbound({
        id: 102,
        kind: 'evaluateConfigPredicates',
        data: { transactionId: 'tx-service', calls: ['predicate'] },
      });
      const activated = await this.inbound({
        id: 103,
        kind: 'activateConfigs',
        data: { transactionId: 'tx-service', effectiveConfigIds: ['config'] },
      });
      return { loaded, evaluated, activated };
    }
    return super.sendMessage(kind, data);
  }
}

describe('RSLintService reverse lint request scoping', () => {
  test('forwards eslintPlugins and dispatches pluginLint to the call handler', async () => {
    const backend = new ReverseLintBackend();
    const service = new RSLintService(backend);
    const eslintPlugins = [{ prefix: 'local', ruleNames: ['program'] }];

    await expect(
      service.lint(
        { files: ['a.ts'], eslintPlugins },
        { pluginLint: (request) => ({ request, ok: true }) },
      ),
    ).resolves.toEqual({ request: { token: 'a.ts' }, ok: true });
    expect(backend.lintPayloads[0]).toMatchObject({ eslintPlugins });
    await service.close();
  });

  test('serializes concurrent lint frames so handlers cannot overwrite each other', async () => {
    const backend = new ReverseLintBackend();
    const service = new RSLintService(backend);

    const [a, b] = await Promise.all([
      service.lint({ files: ['a.ts'] }, { pluginLint: () => ({ host: 'a' }) }),
      service.lint({ files: ['b.ts'] }, { pluginLint: () => ({ host: 'b' }) }),
    ]);

    expect(a).toEqual({ host: 'a' });
    expect(b).toEqual({ host: 'b' });
    await service.close();
  });

  test('rejects an incompatible backend protocol before linting', async () => {
    const backend = new ReverseLintBackend();
    backend.sendMessage = async (kind: string, data: any): Promise<any> => {
      if (kind === 'handshake') return { version: '1.0.0', ok: true };
      return ReverseLintBackend.prototype.sendMessage.call(backend, kind, data);
    };
    const service = new RSLintService(backend);
    await expect(service.lint({ files: ['a.ts'] })).rejects.toThrow(
      /protocol mismatch.*3\.0\.0.*1\.0\.0/,
    );
    await service.close();
  });

  test('requires negotiated reverse lint support for community plugins', async () => {
    const backend = new ReverseLintBackend();
    backend.sendMessage = async (kind: string, data: any): Promise<any> => {
      if (kind === 'handshake') {
        return { version: '3.0.0', ok: true, capabilities: [] };
      }
      return ReverseLintBackend.prototype.sendMessage.call(backend, kind, data);
    };
    const service = new RSLintService(backend);
    await expect(
      service.lint(
        { files: ['a.ts'] },
        { pluginLint: () => ({ diagnostics: [] }) },
      ),
    ).rejects.toThrow(/does not support reverse pluginLint/);
    await service.close();
  });

  test('negotiates and dispatches reverse config discovery requests', async () => {
    const backend = new ReverseConfigBackend();
    const service = new RSLintService(backend);

    await expect(
      service.lint(
        { configDiscovery: { mode: 'auto', inputs: ['.'] } },
        {
          loadConfigs: (request) => ({ loaded: request }),
          evaluateConfigPredicates: (request) => ({ evaluated: request }),
          activateConfigs: (request) => ({ activated: request }),
        },
      ),
    ).resolves.toEqual({
      loaded: {
        loaded: {
          transactionId: 'tx-service',
          candidates: ['candidate'],
        },
      },
      evaluated: {
        evaluated: {
          transactionId: 'tx-service',
          calls: ['predicate'],
        },
      },
      activated: {
        activated: {
          transactionId: 'tx-service',
          effectiveConfigIds: ['config'],
        },
      },
    });
    expect(backend.handshakeCapabilities).toEqual([]);
    expect(backend.lintPayloads).toEqual([]);
    await service.close();
  });

  test('requires reverse config handlers as an atomic pair', async () => {
    const service = new RSLintService(new ReverseConfigBackend());
    await expect(
      service.lint(
        { configDiscovery: { mode: 'auto', inputs: ['.'] } },
        { loadConfigs: () => ({ results: [] }) },
      ),
    ).rejects.toThrow(
      /loadConfigs, evaluateConfigPredicates, and activateConfigs handlers together/,
    );
    await service.close();
  });

  test('bounds graceful shutdown when the peer never acknowledges exit', async () => {
    const backend = new HangingExitBackend();
    const service = new RSLintService(backend);
    const started = Date.now();
    await service.close();
    expect(backend.terminated).toBe(true);
    expect(Date.now() - started).toBeLessThan(2_000);
  });

  test('starts the shutdown bound while an active request is hung', async () => {
    const backend = new HangingLintBackend();
    const service = new RSLintService(backend);
    void service.lint({ files: ['hung.ts'] });
    await backend.lintStarted;

    const started = Date.now();
    await service.close();

    expect(Date.now() - started).toBeLessThan(2_000);
    expect(backend.terminated).toBe(true);
    expect(backend.exitRequests).toBe(0);
  });
});

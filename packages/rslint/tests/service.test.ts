import { describe, expect, test } from '@rstest/core';

import { RSLintService } from '../src/service/service.js';
import { API_REVERSE_CONFIG_LOAD_CAPABILITY } from '../src/service/protocol.js';
import type {
  InboundRequestHandler,
  IpcMessage,
  RslintServiceInterface,
} from '../src/types.js';

class ReverseLintBackend implements RslintServiceInterface {
  inbound: InboundRequestHandler | null = null;
  lintPayloads: unknown[] = [];

  protected beforeReverseLint(): Promise<void> {
    return Promise.resolve();
  }

  setInboundHandler(handler: InboundRequestHandler | null): void {
    this.inbound = handler;
  }

  async sendMessage(kind: string, data: any): Promise<any> {
    if (kind === 'handshake')
      return {
        version: '2.0.0',
        ok: true,
        capabilities: ['reversePluginLint'],
      };
    if (kind === 'exit') return {};
    if (kind !== 'lint') throw new Error(`unexpected kind ${kind}`);

    this.lintPayloads.push(data);
    await this.beforeReverseLint();
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

class GatedReverseLintBackend extends ReverseLintBackend {
  readonly firstLintStarted: Promise<void>;
  private markFirstLintStarted!: () => void;
  private readonly firstLintGate: Promise<void>;
  private releaseFirstLintGate!: () => void;
  private lintCount = 0;

  constructor() {
    super();
    this.firstLintStarted = new Promise((resolve) => {
      this.markFirstLintStarted = resolve;
    });
    this.firstLintGate = new Promise((resolve) => {
      this.releaseFirstLintGate = resolve;
    });
  }

  releaseFirstLint(): void {
    this.releaseFirstLintGate();
  }

  protected override async beforeReverseLint(): Promise<void> {
    this.lintCount++;
    if (this.lintCount !== 1) return;
    this.markFirstLintStarted();
    await this.firstLintGate;
  }
}

class HangingExitBackend extends ReverseLintBackend {
  terminated = false;
  exitRequests = 0;
  readonly exitRequested: Promise<void>;
  private markExitRequested!: () => void;

  constructor() {
    super();
    this.exitRequested = new Promise((resolve) => {
      this.markExitRequested = resolve;
    });
  }

  override async sendMessage(kind: string, data: any): Promise<any> {
    if (kind === 'exit') {
      this.exitRequests++;
      this.markExitRequested();
      return new Promise(() => {
        // Deliberately unresolved to exercise forced shutdown.
      });
    }
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

interface CapturedTimer {
  delay: number;
  fired: boolean;
  readonly cleared: boolean;
  fire: () => void;
}

const NON_FIRING_TIMER_DELAY_MS = 24 * 60 * 60 * 1_000;

function captureServiceCloseTimer(service: RSLintService): {
  close: Promise<void>;
  timers: CapturedTimer[];
} {
  const originalSetTimeout = globalThis.setTimeout;
  const timers: CapturedTimer[] = [];

  globalThis.setTimeout = ((
    callback: (...args: any[]) => void,
    delay?: number,
    ...args: any[]
  ) => {
    if (delay !== 1_000) {
      return originalSetTimeout(callback, delay, ...args);
    }
    // Return a real (unref'ed) handle so production can clear it through the
    // original global after this synchronous capture window is restored.
    const handle = originalSetTimeout(
      () => undefined,
      NON_FIRING_TIMER_DELAY_MS,
    );
    handle.unref();
    const timer: CapturedTimer = {
      delay,
      fired: false,
      get cleared() {
        return (
          (handle as unknown as { _destroyed?: boolean })._destroyed === true
        );
      },
      fire() {
        if (timer.fired || timer.cleared) return;
        timer.fired = true;
        callback(...args);
      },
    };
    timers.push(timer);
    return handle;
  }) as typeof setTimeout;
  try {
    return { close: service.close(), timers };
  } finally {
    // close() installs its timeout before returning its Promise. Never retain
    // a global patch across an await: Rstest timeouts do not cancel a stuck
    // test body, so async restoration would leak into later tests.
    globalThis.setTimeout = originalSetTimeout;
  }
}

class ReverseConfigBackend extends ReverseLintBackend {
  handshakeCapabilities: string[] = [];

  override async sendMessage(kind: string, data: any): Promise<any> {
    if (kind === 'handshake') {
      this.handshakeCapabilities = [...(data.capabilities ?? [])];
      return {
        version: '2.0.0',
        ok: true,
        capabilities: [API_REVERSE_CONFIG_LOAD_CAPABILITY],
      };
    }
    if (kind === 'lint') {
      if (!this.inbound) throw new Error('missing inbound handler');
      const loaded = await this.inbound({
        id: 101,
        kind: 'loadConfigs',
        data: { transactionId: 'tx-service', candidates: ['candidate'] },
      });
      const activated = await this.inbound({
        id: 102,
        kind: 'activateConfigs',
        data: { transactionId: 'tx-service', effectiveConfigIds: ['config'] },
      });
      return { loaded, activated };
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

  test('forwards canonical paths parallel to lint files', async () => {
    const backend = new ReverseLintBackend();
    const service = new RSLintService(backend);

    await service.lint(
      {
        files: ['/lexical/a.ts'],
        canonicalFiles: ['/physical/a.ts'],
      },
      { pluginLint: () => ({ diagnostics: [] }) },
    );
    expect(backend.lintPayloads[0]).toMatchObject({
      files: ['/lexical/a.ts'],
      canonicalFiles: ['/physical/a.ts'],
    });
    await service.close();
  });

  test('serializes concurrent lint frames so handlers cannot overwrite each other', async () => {
    const backend = new GatedReverseLintBackend();
    const service = new RSLintService(backend);

    const first = service.lint(
      { files: ['a.ts'] },
      { pluginLint: () => ({ host: 'a' }) },
    );
    await backend.firstLintStarted;
    const second = service.lint(
      { files: ['b.ts'] },
      { pluginLint: () => ({ host: 'b' }) },
    );

    try {
      expect(backend.lintPayloads).toHaveLength(1);
      backend.releaseFirstLint();
      await expect(first).resolves.toEqual({ host: 'a' });
      await expect(second).resolves.toEqual({ host: 'b' });
      expect(backend.lintPayloads).toHaveLength(2);
    } finally {
      backend.releaseFirstLint();
      await service.close();
    }
  });

  test('rejects an incompatible backend protocol before linting', async () => {
    const backend = new ReverseLintBackend();
    backend.sendMessage = async (kind: string, data: any): Promise<any> => {
      if (kind === 'handshake') return { version: '1.0.0', ok: true };
      return ReverseLintBackend.prototype.sendMessage.call(backend, kind, data);
    };
    const service = new RSLintService(backend);
    await expect(service.lint({ files: ['a.ts'] })).rejects.toThrow(
      /protocol mismatch.*2\.0\.0.*1\.0\.0/,
    );
    await service.close();
  });

  test('requires negotiated reverse lint support for community plugins', async () => {
    const backend = new ReverseLintBackend();
    backend.sendMessage = async (kind: string, data: any): Promise<any> => {
      if (kind === 'handshake') {
        return { version: '2.0.0', ok: true, capabilities: [] };
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
        { configDiscovery: {} },
        {
          loadConfigs: (request) => ({ loaded: request }),
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
      activated: {
        activated: {
          transactionId: 'tx-service',
          effectiveConfigIds: ['config'],
        },
      },
    });
    expect(backend.handshakeCapabilities).toEqual([
      API_REVERSE_CONFIG_LOAD_CAPABILITY,
    ]);
    expect(backend.lintPayloads).toEqual([]);
    await service.close();
  });

  test('requires negotiated reverse config loading support', async () => {
    const backend = new ReverseLintBackend();
    const service = new RSLintService(backend);
    await expect(
      service.lint(
        { configDiscovery: {} },
        {
          loadConfigs: () => ({ results: [] }),
          activateConfigs: () => ({ eslintPluginEntries: [] }),
        },
      ),
    ).rejects.toThrow(/does not support reverse config loading/);
    await service.close();
  });

  test('requires reverse config handlers as an atomic pair', async () => {
    const service = new RSLintService(new ReverseConfigBackend());
    await expect(
      service.lint(
        { configDiscovery: {} },
        { loadConfigs: () => ({ results: [] }) },
      ),
    ).rejects.toThrow(/loadConfigs and activateConfigs handlers together/);
    await service.close();
  });

  test('bounds graceful shutdown when the peer never acknowledges exit', async () => {
    const backend = new HangingExitBackend();
    const service = new RSLintService(backend);
    const captured = captureServiceCloseTimer(service);
    expect(captured.timers).toHaveLength(1);
    expect(captured.timers[0]?.delay).toBe(1_000);
    // Observe the actual exit request rather than assuming how many promise
    // layers the idle queue needs before dispatching it.
    await backend.exitRequested;
    expect(backend.exitRequests).toBe(1);
    captured.timers[0]?.fire();
    await captured.close;
    expect(captured.timers).toHaveLength(1);
    expect(backend.terminated).toBe(true);
    expect(backend.exitRequests).toBe(1);
    expect(captured.timers[0]?.fired).toBe(true);
    expect(captured.timers[0]?.cleared).toBe(true);
  });

  test('starts the shutdown bound while an active request is hung', async () => {
    const backend = new HangingLintBackend();
    const service = new RSLintService(backend);
    void service.lint({ files: ['hung.ts'] });
    await backend.lintStarted;

    const captured = captureServiceCloseTimer(service);
    expect(captured.timers).toHaveLength(1);
    expect(captured.timers[0]?.delay).toBe(1_000);
    expect(backend.exitRequests).toBe(0);
    captured.timers[0]?.fire();
    await captured.close;
    expect(captured.timers).toHaveLength(1);
    expect(backend.terminated).toBe(true);
    expect(backend.exitRequests).toBe(0);
    expect(captured.timers[0]?.fired).toBe(true);
    expect(captured.timers[0]?.cleared).toBe(true);
  });
});

import { describe, expect, test } from '@rstest/core';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';

import {
  CONFIG_DISCOVERY_PROTOCOL_VERSION,
  ConfigModuleHost,
  type ConfigModuleCandidate,
  type LoadConfigsRequest,
} from '../src/config/config-loader.js';

interface Deferred<T> {
  promise: Promise<T>;
  resolve(value: T): void;
}

function deferred<T>(): Deferred<T> {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((res) => {
    resolve = res;
  });
  return { promise, resolve };
}

function createTempDir(): string {
  return fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-config-host-test-'));
}

function candidate(
  root: string,
  id: string,
  name = `${id}.config.js`,
  configDirectory = root,
): ConfigModuleCandidate {
  const configPath = path.join(root, name);
  if (!fs.existsSync(configPath)) fs.writeFileSync(configPath, `// ${id}\n`);
  return { id, configPath, configDirectory };
}

function request(
  transactionId: string,
  candidates: ConfigModuleCandidate[],
  loadMode: 'cached' | 'fresh' = 'cached',
  singleThreaded = false,
): LoadConfigsRequest {
  return {
    protocolVersion: CONFIG_DISCOVERY_PROTOCOL_VERSION,
    transactionId,
    loadMode,
    ...(singleThreaded ? { singleThreaded: true } : {}),
    candidates,
  };
}

describe('ConfigModuleHost', () => {
  test('loads a batch concurrently and returns results in candidate order', async () => {
    const root = createTempDir();
    const slow = candidate(root, 'slow');
    const fast = candidate(root, 'fast');
    const slowStarted = deferred<void>();
    const fastStarted = deferred<void>();
    const releaseSlow = deferred<void>();
    const releaseFast = deferred<void>();
    const completionOrder: string[] = [];

    const host = new ConfigModuleHost({
      loadCached: async (configPath) => {
        const id = path.basename(configPath).split('.')[0];
        if (id === 'slow') {
          slowStarted.resolve();
          await releaseSlow.promise;
        } else {
          fastStarted.resolve();
          await releaseFast.promise;
        }
        completionOrder.push(id);
        return [{ name: id }];
      },
    });

    try {
      const loading = host.loadConfigs(request('tx-order', [slow, fast]));
      await Promise.all([slowStarted.promise, fastStarted.promise]);
      releaseFast.resolve();
      await Promise.resolve();
      releaseSlow.resolve();
      const response = await loading;

      expect(completionOrder).toEqual(['fast', 'slow']);
      expect(response.results.map(({ id }) => id)).toEqual(['slow', 'fast']);
      expect(response.results.map(({ status }) => status)).toEqual([
        'loaded',
        'loaded',
      ]);
    } finally {
      fs.rmSync(root, { recursive: true, force: true });
    }
  });

  test('loads a batch serially when singleThreaded is requested', async () => {
    const root = createTempDir();
    const first = candidate(root, 'first');
    const second = candidate(root, 'second');
    const firstStarted = deferred<void>();
    const secondStarted = deferred<void>();
    const releaseFirst = deferred<void>();
    const releaseSecond = deferred<void>();
    const completionOrder: string[] = [];
    let secondHasStarted = false;
    const host = new ConfigModuleHost({
      loadCached: async (configPath) => {
        const id = path.basename(configPath).split('.')[0];
        if (id === 'first') {
          firstStarted.resolve();
          await releaseFirst.promise;
        } else {
          secondHasStarted = true;
          secondStarted.resolve();
          await releaseSecond.promise;
        }
        completionOrder.push(id);
        return [{ name: id }];
      },
    });

    try {
      const loading = host.loadConfigs(
        request('tx-serial', [first, second], 'cached', true),
      );
      await firstStarted.promise;
      await Promise.resolve();
      expect(secondHasStarted).toBe(false);
      releaseFirst.resolve();
      await secondStarted.promise;
      releaseSecond.resolve();
      const response = await loading;

      expect(completionOrder).toEqual(['first', 'second']);
      expect(response.results.map(({ id }) => id)).toEqual(['first', 'second']);
      await expect(
        host.loadConfigs(request('tx-serial', [], 'cached', false)),
      ).rejects.toThrow('cannot change singleThreaded mode');
    } finally {
      fs.rmSync(root, { recursive: true, force: true });
    }
  });

  test('contains config evaluation and normalization failures per candidate', async () => {
    const root = createTempDir();
    const good = candidate(root, 'good');
    const runtimeFailure = candidate(root, 'runtime');
    const invalid = candidate(root, 'invalid');
    const host = new ConfigModuleHost({
      loadCached: async (configPath) => {
        const id = path.basename(configPath).split('.')[0];
        if (id === 'runtime') throw new Error('boom from config');
        if (id === 'invalid') return { rules: {} };
        return [{ rules: { 'no-console': 'error' } }];
      },
    });

    try {
      const response = await host.loadConfigs(
        request('tx-failures', [good, runtimeFailure, invalid]),
      );
      expect(response.results[0]).toMatchObject({
        id: 'good',
        status: 'loaded',
      });
      expect(response.results[1]).toMatchObject({
        id: 'runtime',
        status: 'failed',
        error: { code: 'load', message: 'boom from config' },
      });
      expect(response.results[2]).toMatchObject({
        id: 'invalid',
        status: 'failed',
        error: {
          code: 'invalid',
          message: expect.stringContaining('must export an array'),
        },
      });
      for (const result of response.results) {
        expect(result.sourceFingerprint).toMatch(/^\d+:[a-f0-9]{64}$/);
      }
    } finally {
      fs.rmSync(root, { recursive: true, force: true });
    }
  });

  test('uses the fresh loader only for fresh sessions', async () => {
    const root = createTempDir();
    const cachedConfig = candidate(root, 'cached');
    const freshConfig = candidate(root, 'fresh');
    const calls: string[] = [];
    const host = new ConfigModuleHost({
      loadCached: async () => {
        calls.push('cached');
        return [];
      },
      loadFresh: async () => {
        calls.push('fresh');
        return [];
      },
    });

    try {
      await host.loadConfigs(request('tx-cached', [cachedConfig], 'cached'));
      await host.loadConfigs(request('tx-fresh', [freshConfig], 'fresh'));
      expect(calls).toEqual(['cached', 'fresh']);
      await expect(
        host.loadConfigs(request('tx-fresh', [], 'cached')),
      ).rejects.toThrow('cannot change loadMode');
    } finally {
      fs.rmSync(root, { recursive: true, force: true });
    }
  });

  test('uses the production config loader and normalizer by default', async () => {
    const root = createTempDir();
    const config = candidate(root, 'default', 'rslint.config.mjs');
    fs.writeFileSync(
      config.configPath,
      'export default [{ rules: { "no-console": "error" } }];\n',
    );
    const host = new ConfigModuleHost();

    try {
      const response = await host.loadConfigs(request('tx-default', [config]));
      expect(response.results[0]).toMatchObject({
        id: 'default',
        status: 'loaded',
        entries: [{ rules: { 'no-console': 'error' } }],
      });
    } finally {
      fs.rmSync(root, { recursive: true, force: true });
    }
  });

  test('marks a config changed during evaluation as failed', async () => {
    const root = createTempDir();
    const changed = candidate(root, 'changed');
    const host = new ConfigModuleHost({
      loadCached: async (configPath) => {
        fs.writeFileSync(configPath, '// changed while evaluating\n');
        return [];
      },
    });

    try {
      const response = await host.loadConfigs(request('tx-changed', [changed]));
      expect(response.results[0]).toMatchObject({
        id: 'changed',
        status: 'failed',
        error: {
          code: 'CONFIG_CHANGED_DURING_LOAD',
          message: expect.stringContaining('changed while it was being loaded'),
        },
      });
    } finally {
      fs.rmSync(root, { recursive: true, force: true });
    }
  });

  test('summarizes only effective configs and derives plugin descriptors', async () => {
    const root = createTempDir();
    const first = candidate(root, 'first');
    const secondDirectory = path.join(root, 'nested');
    fs.mkdirSync(secondDirectory);
    const second = candidate(
      root,
      'second',
      'second.config.js',
      secondDirectory,
    );
    const host = new ConfigModuleHost({
      loadCached: async (configPath) => {
        const id = path.basename(configPath).split('.')[0];
        if (id === 'first') {
          return [
            {
              plugins: {
                shared: { rules: { alpha: {}, common: {} } },
              },
            },
          ];
        }
        return [
          {
            plugins: {
              shared: { rules: { beta: {}, common: {} } },
              nested: { rules: { only: {} } },
            },
          },
        ];
      },
    });

    try {
      const loaded = await host.loadConfigs(
        request('tx-summary', [first, second]),
      );
      // Mutating a response must not corrupt the immutable session copy.
      const firstResult = loaded.results[0];
      if (firstResult.status === 'loaded') firstResult.entries.length = 0;

      const summary = await host.summarizeEffectiveConfigs('tx-summary', [
        'second',
        'first',
      ]);
      expect(summary.configs.map(({ id }) => id)).toEqual(['second', 'first']);
      expect(summary.configs[1].entries).toHaveLength(1);
      expect(summary.pluginConfigs).toEqual([
        {
          configPath: second.configPath,
          configDirectory: second.configDirectory,
        },
        {
          configPath: first.configPath,
          configDirectory: first.configDirectory,
        },
      ]);
      expect(summary.eslintPluginEntries).toEqual([
        { prefix: 'shared', ruleNames: ['beta', 'common', 'alpha'] },
        { prefix: 'nested', ruleNames: ['only'] },
      ]);

      const firstOnly = await host.summarizeEffectiveConfigs('tx-summary', [
        'first',
      ]);
      expect(firstOnly.pluginConfigs).toEqual([
        {
          configPath: first.configPath,
          configDirectory: first.configDirectory,
        },
      ]);
      expect(firstOnly.eslintPluginEntries).toEqual([
        { prefix: 'shared', ruleNames: ['alpha', 'common'] },
      ]);

      const activation = await host.activateConfigs({
        protocolVersion: CONFIG_DISCOVERY_PROTOCOL_VERSION,
        transactionId: 'tx-summary',
        effectiveConfigIds: ['second'],
      });
      expect(activation).toMatchObject({
        transactionId: 'tx-summary',
        configs: [{ id: 'second' }],
        eslintPluginEntries: [
          { prefix: 'shared', ruleNames: ['beta', 'common'] },
          { prefix: 'nested', ruleNames: ['only'] },
        ],
      });
    } finally {
      fs.rmSync(root, { recursive: true, force: true });
    }
  });

  test('rechecks source fingerprints before producing a session summary', async () => {
    const root = createTempDir();
    const config = candidate(root, 'stale');
    const host = new ConfigModuleHost({ loadCached: async () => [] });

    try {
      await host.loadConfigs(request('tx-stale', [config]));
      fs.writeFileSync(config.configPath, '// edited after load\n');
      await expect(
        host.summarizeEffectiveConfigs('tx-stale', ['stale']),
      ).rejects.toThrow('changed while it was being loaded');
    } finally {
      fs.rmSync(root, { recursive: true, force: true });
    }
  });

  test('rejects activation when plugin preparation changes an effective source', async () => {
    const root = createTempDir();
    const config = candidate(root, 'prepare-race');
    const host = new ConfigModuleHost({ loadCached: async () => [] });
    const activationRequest = {
      protocolVersion: CONFIG_DISCOVERY_PROTOCOL_VERSION,
      transactionId: 'tx-prepare-race',
      effectiveConfigIds: ['prepare-race'],
    } as const;
    let prepared = false;

    try {
      await host.loadConfigs(request('tx-prepare-race', [config]));
      await expect(
        host.activateConfigs(activationRequest, undefined, async () => {
          prepared = true;
          fs.writeFileSync(config.configPath, '// worker imported new bytes\n');
        }),
      ).rejects.toThrow('plugin host was being prepared');
      expect(prepared).toBe(true);
      // The caller still owns explicit commit/abort cleanup after a rejected
      // activation; no stale activation was returned from the session.
      expect(host.hasSession('tx-prepare-race')).toBe(true);
      expect(host.deleteSession('tx-prepare-race')).toBe(true);
    } finally {
      fs.rmSync(root, { recursive: true, force: true });
    }
  });

  test('reuses a candidate ID within one session', async () => {
    const root = createTempDir();
    const config = candidate(root, 'same');
    let calls = 0;
    const host = new ConfigModuleHost({
      loadCached: async () => {
        calls++;
        return [{ name: 'original' }];
      },
    });

    try {
      const first = await host.loadConfigs(request('tx-repeat', [config]));
      if (first.results[0].status === 'loaded') {
        first.results[0].entries[0].name = 'mutated';
      }
      const second = await host.loadConfigs(request('tx-repeat', [config]));
      expect(calls).toBe(1);
      expect(second.results[0]).toMatchObject({
        status: 'loaded',
        entries: [{ name: 'original' }],
      });
    } finally {
      fs.rmSync(root, { recursive: true, force: true });
    }
  });

  test('rejects ambiguous IDs and unknown effective candidates', async () => {
    const root = createTempDir();
    const one = candidate(root, 'one');
    const host = new ConfigModuleHost({ loadCached: async () => [] });

    try {
      await expect(
        host.loadConfigs(request('tx-duplicates', [one, one])),
      ).rejects.toThrow('duplicate id');
      await host.loadConfigs(request('tx-known', [one]));
      await expect(
        host.summarizeEffectiveConfigs('tx-known', ['missing']),
      ).rejects.toThrow('unknown effective config id');
      expect(host.hasSession('tx-known')).toBe(true);
      expect(host.deleteSession('tx-known')).toBe(true);
      expect(host.hasSession('tx-known')).toBe(false);
    } finally {
      fs.rmSync(root, { recursive: true, force: true });
    }
  });
});

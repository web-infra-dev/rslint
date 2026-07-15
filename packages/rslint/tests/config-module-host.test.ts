import { describe, expect, test } from '@rstest/core';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';

import {
  ConfigModuleHost,
  type ActivateConfigsResponse,
  type ConfigModuleActivationPlan,
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
    transactionId,
    loadMode,
    ...(singleThreaded ? { singleThreaded: true } : {}),
    candidates,
  };
}

async function activateWithPlan(
  host: ConfigModuleHost,
  transactionId: string,
  effectiveConfigIds: string[],
): Promise<{
  response: ActivateConfigsResponse;
  plan: ConfigModuleActivationPlan;
}> {
  let plan: ConfigModuleActivationPlan | undefined;
  const response = await host.activateConfigs(
    {
      transactionId,
      effectiveConfigIds,
    },
    undefined,
    async (candidate) => {
      plan = candidate;
    },
  );
  if (!plan) throw new Error('activation prepare callback was not called');
  return { response, plan };
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

  test('does not serialize independent parallel loads in one transaction', async () => {
    const root = createTempDir();
    const blocked = candidate(root, 'blocked');
    const ready = candidate(root, 'ready');
    const blockedStarted = deferred<void>();
    const releaseBlocked = deferred<void>();
    let blockedSettled = false;
    const host = new ConfigModuleHost({
      loadCached: async (configPath) => {
        if (configPath === blocked.configPath) {
          blockedStarted.resolve();
          await releaseBlocked.promise;
        }
        return [{ name: path.basename(configPath) }];
      },
    });

    try {
      const blockedLoad = host
        .loadConfigs(request('tx-independent', [blocked]))
        .then((result) => {
          blockedSettled = true;
          return result;
        });
      await blockedStarted.promise;
      const readyResult = await host.loadConfigs(
        request('tx-independent', [ready]),
      );
      expect(readyResult.results[0]).toMatchObject({
        id: 'ready',
        status: 'loaded',
      });
      expect(blockedSettled).toBe(false);
      releaseBlocked.resolve();
      await blockedLoad;
    } finally {
      fs.rmSync(root, { recursive: true, force: true });
    }
  });

  test('coalesces concurrent requests for the same candidate id', async () => {
    const root = createTempDir();
    const shared = candidate(root, 'shared');
    const started = deferred<void>();
    const release = deferred<void>();
    let moduleLoads = 0;
    let iterations = 0;
    const exported = new Proxy([{}], {
      get(target, property, receiver) {
        if (property === Symbol.iterator) iterations++;
        return Reflect.get(target, property, receiver);
      },
    });
    const host = new ConfigModuleHost({
      loadCached: async () => {
        moduleLoads++;
        started.resolve();
        await release.promise;
        return exported;
      },
    });

    try {
      const first = host.loadConfigs(request('tx-same-id', [shared]));
      await started.promise;
      const second = host.loadConfigs(request('tx-same-id', [shared]));
      release.resolve();
      const [firstResult, secondResult] = await Promise.all([first, second]);
      expect(moduleLoads).toBe(1);
      expect(iterations).toBe(1);
      expect(secondResult).toEqual(firstResult);
    } finally {
      fs.rmSync(root, { recursive: true, force: true });
    }
  });

  test('shares one physical module load while normalizing every lexical candidate', async () => {
    const root = createTempDir();
    const physicalDirectory = path.join(root, 'physical');
    const aliasDirectory = path.join(root, 'alias');
    fs.mkdirSync(physicalDirectory);
    fs.symlinkSync(
      physicalDirectory,
      aliasDirectory,
      process.platform === 'win32' ? 'junction' : 'dir',
    );
    const first = candidate(
      physicalDirectory,
      'first',
      'shared.config.js',
      physicalDirectory,
    );
    const second: ConfigModuleCandidate = {
      ...first,
      id: 'second',
      configPath: path.join(aliasDirectory, 'shared.config.js'),
      configDirectory: aliasDirectory,
    };
    let moduleLoads = 0;
    let iterations = 0;
    const matcher = () => true;
    const exported = new Proxy(
      [
        {
          files: [matcher],
          plugins: { shared: { rules: { rule: {} } } },
        },
      ],
      {
        get(target, property, receiver) {
          if (property === Symbol.iterator) iterations++;
          return Reflect.get(target, property, receiver);
        },
      },
    );
    const host = new ConfigModuleHost({
      loadFresh: async () => {
        moduleLoads++;
        return exported;
      },
    });

    try {
      const loaded = await host.loadConfigs(
        request('tx-physical-alias', [first, second], 'fresh'),
      );
      expect(moduleLoads).toBe(1);
      expect(iterations).toBe(2);
      expect(loaded.results.map(({ id, status }) => ({ id, status }))).toEqual([
        { id: 'first', status: 'loaded' },
        { id: 'second', status: 'loaded' },
      ]);
      const predicateIds = loaded.results.map((result) => {
        if (result.status !== 'loaded') throw new Error(result.error.message);
        const entry = result.entries[0] as {
          files: Array<{ $rslintPredicate: string }>;
        };
        return entry.files[0].$rslintPredicate;
      });
      expect(predicateIds[0]).not.toBe(predicateIds[1]);
      expect(predicateIds[0]).toContain('first:predicate-');
      expect(predicateIds[1]).toContain('second:predicate-');

      const { plan } = await activateWithPlan(host, 'tx-physical-alias', [
        'first',
        'second',
      ]);
      expect(
        plan.configs.map(({ configDirectory }) => configDirectory),
      ).toEqual([physicalDirectory, aliasDirectory]);
      expect(plan.pluginConfigs).toEqual([
        {
          configPath: first.configPath,
          configDirectory: physicalDirectory,
        },
        {
          configPath: second.configPath,
          configDirectory: aliasDirectory,
        },
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

  test('serializes independent requests in a singleThreaded transaction', async () => {
    const root = createTempDir();
    const first = candidate(root, 'first-request');
    const second = candidate(root, 'second-request');
    const firstStarted = deferred<void>();
    const releaseFirst = deferred<void>();
    let secondStarted = false;
    const host = new ConfigModuleHost({
      loadCached: async (configPath) => {
        if (configPath === first.configPath) {
          firstStarted.resolve();
          await releaseFirst.promise;
        } else {
          secondStarted = true;
        }
        return [{}];
      },
    });

    try {
      const firstLoad = host.loadConfigs(
        request('tx-serial-requests', [first], 'cached', true),
      );
      await firstStarted.promise;
      const secondLoad = host.loadConfigs(
        request('tx-serial-requests', [second], 'cached', true),
      );
      await Promise.resolve();
      expect(secondStarted).toBe(false);
      releaseFirst.resolve();
      await Promise.all([firstLoad, secondLoad]);
      expect(secondStarted).toBe(true);
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
        if (id === 'invalid') return 42;
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
          message: expect.stringContaining('must be an object'),
        },
      });
      expect(
        response.results.every(
          (result) =>
            !('sourceFingerprint' in result) && !('eslintPlugins' in result),
        ),
      ).toBe(true);
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
    // Go's canonical path representation uses forward slashes on every OS.
    // The routing identity must survive the Node module host byte-for-byte;
    // path.normalize would otherwise rewrite it on Windows only.
    const secondRoutingDirectory = secondDirectory.replaceAll(path.sep, '/');
    const second = candidate(
      root,
      'second',
      'second.config.js',
      secondRoutingDirectory,
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

      const { plan: summary } = await activateWithPlan(host, 'tx-summary', [
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

      const { plan: firstOnly } = await activateWithPlan(host, 'tx-summary', [
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

      const { response: activation } = await activateWithPlan(
        host,
        'tx-summary',
        ['second'],
      );
      expect(activation).toEqual({
        transactionId: 'tx-summary',
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
        host.activateConfigs({
          transactionId: 'tx-stale',
          effectiveConfigIds: ['stale'],
        }),
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
      expect(host.deleteSession('tx-prepare-race')).toBe(true);
      expect(host.deleteSession('tx-prepare-race')).toBe(false);
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
        host.activateConfigs({
          transactionId: 'tx-known',
          effectiveConfigIds: ['missing'],
        }),
      ).rejects.toThrow('unknown effective config id');
      expect(host.deleteSession('tx-known')).toBe(true);
      expect(host.deleteSession('tx-known')).toBe(false);
    } finally {
      fs.rmSync(root, { recursive: true, force: true });
    }
  });

  test('evaluates live matchers synchronously with native paths and JS truthiness', async () => {
    const root = createTempDir();
    const config = candidate(root, 'predicates');
    const observed: Array<{ path: string; receiver: unknown }> = [];
    const record = function (this: unknown, absolutePath: string): void {
      observed.push({ path: absolutePath, receiver: this });
    };
    const host = new ConfigModuleHost({
      loadCached: async () => [
        {
          files: [
            function (this: unknown, absolutePath: string) {
              record.call(this, absolutePath);
              return 0;
            },
            function (this: unknown, absolutePath: string) {
              record.call(this, absolutePath);
              return { matched: false };
            },
            function (this: unknown, absolutePath: string) {
              record.call(this, absolutePath);
              return Promise.resolve(false);
            },
            function (this: unknown, absolutePath: string) {
              record.call(this, absolutePath);
              throw new Error('matcher boom');
            },
          ],
        },
      ],
    });

    try {
      const loaded = await host.loadConfigs(request('tx-predicates', [config]));
      const result = loaded.results[0];
      if (result.status !== 'loaded') throw new Error(result.error.message);
      const entry = result.entries[0] as {
        files: Array<{ $rslintPredicate: string }>;
      };
      const predicateIds = entry.files.map(
        (descriptor) => descriptor.$rslintPredicate,
      );
      expect(new Set(predicateIds).size).toBe(4);

      const directory = path.join(root, 'nested');
      const response = await host.evaluateConfigPredicates({
        transactionId: 'tx-predicates',
        calls: predicateIds.map((predicateId, index) => ({
          callId: `call-${index}`,
          predicateId,
          absolutePath: directory,
          directory: true,
        })),
      });
      expect(response.results).toEqual([
        { callId: 'call-0', status: 'evaluated', value: false },
        { callId: 'call-1', status: 'evaluated', value: true },
        { callId: 'call-2', status: 'evaluated', value: true },
        {
          callId: 'call-3',
          status: 'failed',
          error: { code: 'predicate', message: 'matcher boom' },
        },
      ]);
      expect(observed).toEqual(
        Array.from({ length: 4 }, () => ({
          path: `${path.normalize(directory)}${path.sep}`,
          receiver: undefined,
        })),
      );

      await activateWithPlan(host, 'tx-predicates', ['predicates']);
      await expect(
        host.evaluateConfigPredicates({
          transactionId: 'tx-predicates',
          calls: [
            {
              callId: 'unknown',
              predicateId: 'predicates:predicate-999999',
              absolutePath: directory,
            },
          ],
        }),
      ).rejects.toThrow('unknown predicate id');
      expect(host.deleteSession('tx-predicates')).toBe(true);
      await expect(
        host.evaluateConfigPredicates({
          transactionId: 'tx-predicates',
          calls: [
            {
              callId: 'stale',
              predicateId: predicateIds[0],
              absolutePath: directory,
            },
          ],
        }),
      ).rejects.toThrow('unknown transaction');
    } finally {
      fs.rmSync(root, { recursive: true, force: true });
    }
  });
});

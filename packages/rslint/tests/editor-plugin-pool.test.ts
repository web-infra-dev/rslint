import { describe, expect, test } from '@rstest/core';

import type { ConfigModuleActivationPlan } from '../src/config/config-loader.js';
import {
  EditorPluginPool,
  type EditorPluginLintHost,
  type EditorPluginModule,
} from '../src/editor-runtime/editor-plugin-pool.js';

interface FakeHost extends EditorPluginLintHost {
  shutdowns: number;
}

function activationPlan(
  transactionId: string,
  sourceFingerprint: string,
): ConfigModuleActivationPlan {
  const configPath = '/workspace/rslint.config.mjs';
  return {
    transactionId,
    configs: [
      {
        id: 'config',
        configPath,
        configDirectory: '/workspace',
        entries: [],
        sourceFingerprint,
      },
    ],
    eslintPluginEntries: [{ prefix: 'test', ruleNames: ['rule'] }],
    pluginConfigs: [{ configPath, configDirectory: '/workspace' }],
  };
}

function fakeModule(
  create: () => Promise<EditorPluginLintHost> | EditorPluginLintHost,
): EditorPluginModule {
  return {
    async createPluginLintHost() {
      return create();
    },
  };
}

function simpleHost(): FakeHost {
  return {
    shutdowns: 0,
    async lint(request) {
      return request;
    },
    async shutdown() {
      this.shutdowns++;
    },
  };
}

describe('editor plugin generation pool', () => {
  test('reuses the active host only for the exact source and dependency revision', async () => {
    const hosts: FakeHost[] = [];
    const pool = new EditorPluginPool({
      graceGenerationMs: 60_000,
      loadPluginModule: async () =>
        fakeModule(() => {
          const host = simpleHost();
          hosts.push(host);
          return host;
        }),
    });
    try {
      expect(await pool.prepare(activationPlan('g1', 'source-a'), 4)).toBe(
        true,
      );
      expect(await pool.commit('g1')).toBe(true);
      expect(await pool.prepare(activationPlan('g2', 'source-a'), 4)).toBe(
        true,
      );
      expect(await pool.commit('g2')).toBe(true);
      expect(hosts).toHaveLength(1);

      expect(await pool.prepare(activationPlan('g3', 'source-a'), 5)).toBe(
        true,
      );
      expect(await pool.commit('g3')).toBe(true);
      expect(hosts).toHaveLength(2);
    } finally {
      await pool.dispose();
    }
    expect(hosts.map((host) => host.shutdowns)).toEqual([1, 1]);
  });

  test('bounds live history to active plus one grace generation', async () => {
    const hosts: FakeHost[] = [];
    const pool = new EditorPluginPool({
      graceGenerationMs: 60_000,
      maxGraceGenerations: 1,
      loadPluginModule: async () =>
        fakeModule(() => {
          const host = simpleHost();
          hosts.push(host);
          return host;
        }),
    });
    try {
      for (const [generation, source] of [
        ['g1', 'source-1'],
        ['g2', 'source-2'],
        ['g3', 'source-3'],
      ] as const) {
        expect(await pool.prepare(activationPlan(generation, source), 0)).toBe(
          true,
        );
        expect(await pool.commit(generation)).toBe(true);
      }
      await Promise.resolve();
      expect(hosts).toHaveLength(3);
      expect(hosts.map((host) => host.shutdowns)).toEqual([1, 0, 0]);
    } finally {
      await pool.dispose();
    }
    expect(hosts.map((host) => host.shutdowns)).toEqual([1, 1, 1]);
  });

  test('does not shut down a retired host until its in-flight lint lease ends', async () => {
    let finishLint!: () => void;
    const lintFinished = new Promise<void>((resolve) => {
      finishLint = resolve;
    });
    const first = simpleHost();
    first.lint = async () => {
      await lintFinished;
      return { results: [] };
    };
    const second = simpleHost();
    const hosts = [first, second];
    let nextHost = 0;
    const pool = new EditorPluginPool({
      maxGraceGenerations: 0,
      loadPluginModule: async () => fakeModule(() => hosts[nextHost++]),
    });
    try {
      expect(await pool.prepare(activationPlan('g1', 'source-1'), 0)).toBe(
        true,
      );
      expect(await pool.commit('g1')).toBe(true);
      const lint = pool.lint({ generation: 'g1' });
      await Promise.resolve();

      expect(await pool.prepare(activationPlan('g2', 'source-2'), 0)).toBe(
        true,
      );
      expect(await pool.commit('g2')).toBe(true);
      expect(first.shutdowns).toBe(0);

      finishLint();
      await lint;
      await Promise.resolve();
      expect(first.shutdowns).toBe(1);
    } finally {
      finishLint();
      await pool.dispose();
    }
  });

  test('sidecar disposal force-shuts a host with an active lint lease', async () => {
    let finishLint!: (value: unknown) => void;
    const host = simpleHost();
    host.lint = async () =>
      new Promise((resolve) => {
        finishLint = resolve;
      });
    host.shutdown = async () => {
      host.shutdowns++;
      finishLint({ results: [] });
    };
    const pool = new EditorPluginPool({
      loadPluginModule: async () => fakeModule(() => host),
    });
    expect(await pool.prepare(activationPlan('g1', 'source'), 0)).toBe(true);
    expect(await pool.commit('g1')).toBe(true);
    const lint = pool.lint({ generation: 'g1' });
    await Promise.resolve();

    await pool.dispose();

    expect(host.shutdowns).toBe(1);
    await expect(lint).resolves.toEqual({ results: [] });
  });

  test('retries an unavailable active fingerprint on the next transaction', async () => {
    let attempts = 0;
    const recovered = simpleHost();
    const pool = new EditorPluginPool({
      maxGraceGenerations: 0,
      log: () => undefined,
      loadPluginModule: async () =>
        fakeModule(() => {
          attempts++;
          if (attempts === 1) throw new Error('initialization failed');
          return recovered;
        }),
    });
    try {
      expect(await pool.prepare(activationPlan('g1', 'same-source'), 0)).toBe(
        false,
      );
      expect(await pool.commit('g1')).toBe(true);
      expect(await pool.prepare(activationPlan('g2', 'same-source'), 0)).toBe(
        true,
      );
      expect(await pool.commit('g2')).toBe(true);
      expect(attempts).toBe(2);
    } finally {
      await pool.dispose();
    }
    expect(recovered.shutdowns).toBe(1);
  });

  test('aborting a staged generation releases its host', async () => {
    const host = simpleHost();
    const pool = new EditorPluginPool({
      loadPluginModule: async () => fakeModule(() => host),
    });
    expect(await pool.prepare(activationPlan('staged', 'source'), 0)).toBe(
      true,
    );
    await pool.abort('staged');
    await Promise.resolve();
    expect(host.shutdowns).toBe(1);
    await pool.dispose();
  });
});

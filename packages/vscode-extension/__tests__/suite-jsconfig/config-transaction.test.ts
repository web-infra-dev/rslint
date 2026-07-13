import * as assert from 'node:assert';

import {
  CONFIG_DISCOVERY_PROTOCOL_VERSION,
  type ActivateConfigsRequest,
  type ActivateConfigsResponse,
  type LoadConfigsRequest,
  type LoadConfigsResponse,
} from '@rslint/core/config-loader';
import {
  CONFIG_REFRESH_WATCH_GLOB,
  CONFIG_WATCH_GLOB,
  LspConfigTransactionAdapter,
  configRefreshReasonForPath,
  isConfigSourceChangeDuringTransaction,
  normalizeConfigTransactionVersion,
  retryConfigRefreshOnSourceChange,
  type ConfigModuleHostAdapter,
  type PluginLintPoolAdapter,
} from '../../src/Rslint';

suite('initial config refresh retry classification', () => {
  test('recognizes source-change failures without hiding unrelated startup failures', () => {
    assert.strictEqual(
      isConfigSourceChangeDuringTransaction({
        code: 'CONFIG_CHANGED_DURING_LOAD',
        message: 'wrapped transport error',
      }),
      true,
    );
    assert.strictEqual(
      isConfigSourceChangeDuringTransaction(
        new Error(
          'activate configs: config changed while its plugin host was being prepared',
        ),
      ),
      true,
    );
    assert.strictEqual(
      isConfigSourceChangeDuringTransaction(new Error('plugin host failed')),
      false,
    );
  });

  test('retries one transient initial source-change failure', async () => {
    let initialCalls = 0;
    let retryCalls = 0;
    const retried = await retryConfigRefreshOnSourceChange(
      async () => {
        initialCalls++;
        throw new Error(
          'config changed while its plugin host was being prepared',
        );
      },
      async () => {
        retryCalls++;
      },
    );
    assert.strictEqual(retried, true);
    assert.strictEqual(initialCalls, 1);
    assert.strictEqual(retryCalls, 1);
  });

  test('does not retry unrelated initial failures', async () => {
    let retryCalls = 0;
    await assert.rejects(
      retryConfigRefreshOnSourceChange(
        async () => {
          throw new Error('protocol failure');
        },
        async () => {
          retryCalls++;
        },
      ),
      /protocol failure/,
    );
    assert.strictEqual(retryCalls, 0);
  });
});

class TestConfigHost implements ConfigModuleHostAdapter {
  readonly loadRequests: LoadConfigsRequest[] = [];
  readonly activationRequests: ActivateConfigsRequest[] = [];
  readonly deletedTransactions: string[] = [];
  loadError: Error | undefined;
  changedDuringPrepare = false;
  activation: ActivateConfigsResponse = {
    transactionId: 'tx-1',
    configs: [
      {
        id: 'root',
        configPath: '/workspace/rslint.config.mjs',
        configDirectory: '/workspace',
        entries: [],
        sourceFingerprint: '10:abc',
      },
    ],
    eslintPluginEntries: [{ prefix: 'local', ruleNames: ['no-foo'] }],
    pluginConfigs: [
      {
        configPath: '/workspace/rslint.config.mjs',
        configDirectory: '/workspace',
      },
    ],
  };

  async loadConfigs(request: LoadConfigsRequest): Promise<LoadConfigsResponse> {
    this.loadRequests.push(request);
    if (this.loadError) throw this.loadError;
    return { transactionId: request.transactionId, results: [] };
  }

  async activateConfigs(
    request: ActivateConfigsRequest,
    _signal?: AbortSignal,
    prepare?: (activation: ActivateConfigsResponse) => Promise<void>,
  ): Promise<ActivateConfigsResponse> {
    this.activationRequests.push(request);
    const activation = {
      ...this.activation,
      transactionId: request.transactionId,
    };
    await prepare?.(activation);
    if (this.changedDuringPrepare) {
      throw new Error(
        'config changed while its plugin host was being prepared',
      );
    }
    return activation;
  }

  deleteSession(transactionId: string): boolean {
    this.deletedTransactions.push(transactionId);
    return true;
  }
}

class TestPluginPool implements PluginLintPoolAdapter {
  readonly prepareCalls: Array<{
    descriptors: unknown[];
    fingerprint: string;
    generation: string;
  }> = [];
  readonly commitCalls: string[] = [];
  readonly abortCalls: string[] = [];
  ready = true;
  commitResult = true;
  onPrepare: (() => void | Promise<void>) | undefined;

  async prepare(
    descriptors: Array<{
      configPath: string;
      configDirectory: string;
    }>,
    fingerprint: string,
    generation: string,
  ): Promise<boolean> {
    this.prepareCalls.push({ descriptors, fingerprint, generation });
    await this.onPrepare?.();
    return this.ready;
  }

  async commit(generation: string): Promise<boolean> {
    this.commitCalls.push(generation);
    return this.commitResult;
  }

  async abort(generation: string): Promise<void> {
    this.abortCalls.push(generation);
  }
}

function loadRequest(transactionId = 'tx-1'): LoadConfigsRequest {
  return {
    protocolVersion: CONFIG_DISCOVERY_PROTOCOL_VERSION,
    transactionId,
    loadMode: 'cached',
    candidates: [],
  };
}

suite('LSP config discovery transactions', () => {
  test('transaction capability versions preserve both legacy protocols', () => {
    assert.strictEqual(normalizeConfigTransactionVersion(undefined), 0);
    assert.strictEqual(normalizeConfigTransactionVersion(0), 0);
    assert.strictEqual(normalizeConfigTransactionVersion(1), 1);
    assert.strictEqual(normalizeConfigTransactionVersion(2), 2);
    assert.strictEqual(normalizeConfigTransactionVersion(3), 2);
    assert.strictEqual(normalizeConfigTransactionVersion(2.5), 0);
  });

  test('legacy and v2 direct watchers keep one owner for gitignore changes', () => {
    assert.match(CONFIG_WATCH_GLOB, /rslint\.config\.mjs/);
    assert.match(CONFIG_WATCH_GLOB, /rslint\.jsonc/);
    assert.match(CONFIG_WATCH_GLOB, /\.gitignore/);
    assert.match(CONFIG_WATCH_GLOB, /pnpm-lock\.yaml/);
    assert.match(CONFIG_REFRESH_WATCH_GLOB, /rslint\.config\.mjs/);
    assert.match(CONFIG_REFRESH_WATCH_GLOB, /rslint\.jsonc/);
    assert.match(CONFIG_REFRESH_WATCH_GLOB, /pnpm-lock\.yaml/);
    assert.doesNotMatch(CONFIG_REFRESH_WATCH_GLOB, /\.gitignore/);
    assert.strictEqual(
      configRefreshReasonForPath('/workspace/packages/app/.gitignore'),
      'gitignore-change',
    );
    assert.strictEqual(
      configRefreshReasonForPath('/workspace/packages/app/pnpm-lock.yaml'),
      'dependency-change',
    );
    assert.strictEqual(
      configRefreshReasonForPath('/workspace/rslint.config.mjs'),
      'config-change',
    );
  });

  test('loads fresh, stages only effective plugins, then commits atomically', async () => {
    const host = new TestConfigHost();
    const pool = new TestPluginPool();
    const adapter = new LspConfigTransactionAdapter(
      host,
      pool,
      () => 'fingerprint-1',
    );

    const loaded = await adapter.loadConfigs(loadRequest());
    assert.strictEqual(loaded.transactionId, 'tx-1');
    assert.strictEqual(host.loadRequests[0].loadMode, 'fresh');

    const activated = await adapter.activateConfigs({
      protocolVersion: CONFIG_DISCOVERY_PROTOCOL_VERSION,
      transactionId: 'tx-1',
      effectiveConfigIds: ['root'],
    });
    assert.deepStrictEqual(activated, {
      transactionId: 'tx-1',
      generation: 'tx-1',
      eslintPluginEntries: [{ prefix: 'local', ruleNames: ['no-foo'] }],
      pluginHostReady: true,
    });
    assert.deepStrictEqual(pool.prepareCalls, [
      {
        descriptors: host.activation.pluginConfigs,
        fingerprint: 'fingerprint-1',
        generation: 'tx-1',
      },
    ]);

    const committed = await adapter.commitConfigs({
      protocolVersion: CONFIG_DISCOVERY_PROTOCOL_VERSION,
      transactionId: 'tx-1',
    });
    assert.deepStrictEqual(committed, {
      transactionId: 'tx-1',
      generation: 'tx-1',
      committed: true,
    });
    assert.deepStrictEqual(pool.commitCalls, ['tx-1']);
    assert.deepStrictEqual(host.deletedTransactions, ['tx-1']);
  });

  test('a rejected commit retains the session until Go aborts it', async () => {
    const host = new TestConfigHost();
    const pool = new TestPluginPool();
    pool.commitResult = false;
    const adapter = new LspConfigTransactionAdapter(
      host,
      pool,
      () => 'fingerprint-1',
    );

    await adapter.loadConfigs(loadRequest('tx-abort'));
    await adapter.activateConfigs({
      protocolVersion: CONFIG_DISCOVERY_PROTOCOL_VERSION,
      transactionId: 'tx-abort',
      effectiveConfigIds: ['root'],
    });
    await assert.rejects(
      adapter.commitConfigs({
        protocolVersion: CONFIG_DISCOVERY_PROTOCOL_VERSION,
        transactionId: 'tx-abort',
      }),
      /failed to commit plugin-host generation/,
    );
    assert.deepStrictEqual(host.deletedTransactions, []);

    const aborted = await adapter.abortConfigs({
      protocolVersion: CONFIG_DISCOVERY_PROTOCOL_VERSION,
      transactionId: 'tx-abort',
    });
    assert.deepStrictEqual(aborted, {
      transactionId: 'tx-abort',
      generation: 'tx-abort',
      aborted: true,
    });
    assert.deepStrictEqual(pool.abortCalls, ['tx-abort']);
    assert.deepStrictEqual(host.deletedTransactions, ['tx-abort']);
  });

  test('a failed first plugin prepare can commit a degraded no-host generation', async () => {
    const host = new TestConfigHost();
    const pool = new TestPluginPool();
    pool.ready = false;
    const adapter = new LspConfigTransactionAdapter(
      host,
      pool,
      () => 'fingerprint-degraded',
    );

    await adapter.loadConfigs(loadRequest('tx-degraded'));
    const activated = await adapter.activateConfigs({
      protocolVersion: CONFIG_DISCOVERY_PROTOCOL_VERSION,
      transactionId: 'tx-degraded',
      effectiveConfigIds: ['root'],
    });
    assert.deepStrictEqual(activated, {
      transactionId: 'tx-degraded',
      generation: 'tx-degraded',
      eslintPluginEntries: [],
      pluginHostReady: false,
    });

    const committed = await adapter.commitConfigs({
      protocolVersion: CONFIG_DISCOVERY_PROTOCOL_VERSION,
      transactionId: 'tx-degraded',
    });
    assert.strictEqual(committed.committed, true);
    assert.deepStrictEqual(pool.commitCalls, ['tx-degraded']);
    assert.deepStrictEqual(pool.abortCalls, []);
    assert.deepStrictEqual(host.deletedTransactions, ['tx-degraded']);
  });

  test('a config rewrite during plugin prepare aborts without commit or leaked session', async () => {
    const host = new TestConfigHost();
    const pool = new TestPluginPool();
    pool.onPrepare = () => {
      host.changedDuringPrepare = true;
    };
    const adapter = new LspConfigTransactionAdapter(
      host,
      pool,
      () => 'fingerprint-before-prepare',
    );

    await adapter.loadConfigs(loadRequest('tx-prepare-race'));
    await assert.rejects(
      adapter.activateConfigs({
        protocolVersion: CONFIG_DISCOVERY_PROTOCOL_VERSION,
        transactionId: 'tx-prepare-race',
        effectiveConfigIds: ['root'],
      }),
      /plugin host was being prepared/,
    );
    assert.deepStrictEqual(pool.abortCalls, ['tx-prepare-race']);
    assert.deepStrictEqual(pool.commitCalls, []);
    assert.deepStrictEqual(host.deletedTransactions, ['tx-prepare-race']);
  });

  test('an abort after a successful local commit is still compensating', async () => {
    const host = new TestConfigHost();
    const pool = new TestPluginPool();
    const adapter = new LspConfigTransactionAdapter(
      host,
      pool,
      () => 'fingerprint-1',
    );

    await adapter.loadConfigs(loadRequest('tx-response-lost'));
    await adapter.activateConfigs({
      protocolVersion: CONFIG_DISCOVERY_PROTOCOL_VERSION,
      transactionId: 'tx-response-lost',
      effectiveConfigIds: ['root'],
    });
    await adapter.commitConfigs({
      protocolVersion: CONFIG_DISCOVERY_PROTOCOL_VERSION,
      transactionId: 'tx-response-lost',
    });

    const aborted = await adapter.abortConfigs({
      protocolVersion: CONFIG_DISCOVERY_PROTOCOL_VERSION,
      transactionId: 'tx-response-lost',
    });
    assert.deepStrictEqual(aborted, {
      transactionId: 'tx-response-lost',
      generation: 'tx-response-lost',
      aborted: true,
    });
    assert.deepStrictEqual(pool.commitCalls, ['tx-response-lost']);
    assert.deepStrictEqual(pool.abortCalls, ['tx-response-lost']);
    assert.deepStrictEqual(host.deletedTransactions, [
      'tx-response-lost',
      'tx-response-lost',
    ]);
  });

  test('a failed module-host request cannot leak transaction state', async () => {
    const host = new TestConfigHost();
    host.loadError = new Error('load failed');
    const adapter = new LspConfigTransactionAdapter(
      host,
      new TestPluginPool(),
      () => 'fingerprint-1',
    );

    await assert.rejects(adapter.loadConfigs(loadRequest()), /load failed/);
    assert.deepStrictEqual(host.deletedTransactions, ['tx-1']);
  });
});

import * as assert from 'node:assert';

import {
  CONFIG_DISCOVERY_PROTOCOL_VERSION,
  type ActivateConfigsRequest,
  type ActivateConfigsResponse,
  type ConfigModuleActivationPlan,
  type LoadConfigsRequest,
  type LoadConfigsResponse,
} from '@rslint/core/config-loader';
import {
  CONFIG_REFRESH_WATCH_GLOB,
  configRefreshReasonForPath,
  createLanguageClientOptions,
  isConfigSourceChangeDuringTransaction,
  recoverConfigDiscoveryOnServerState,
  retryConfigRefreshOnSourceChange,
} from '../../src/Rslint';
import { LspConfigTransactionAdapter } from '../../src/ConfigTransactionAdapter';
import { State } from 'vscode-languageclient/node';
import {
  RelativePattern,
  Uri,
  type DocumentFilter,
  type WorkspaceFolder,
} from 'vscode';

suite('initial config refresh retry classification', () => {
  test('isolates each language client and its documents to one workspace', () => {
    const firstFolder: WorkspaceFolder = {
      index: 0,
      name: 'first-root',
      uri: Uri.file('/workspace/first-root'),
    };
    const secondFolder: WorkspaceFolder = {
      index: 1,
      name: 'second-root',
      uri: Uri.file('/workspace/second-root'),
    };
    const firstOptions = createLanguageClientOptions(firstFolder, undefined);
    const secondOptions = createLanguageClientOptions(secondFolder, undefined);

    assert.strictEqual(firstOptions.workspaceFolder, firstFolder);
    assert.strictEqual(secondOptions.workspaceFolder, secondFolder);
    for (const [options, folder] of [
      [firstOptions, firstFolder],
      [secondOptions, secondFolder],
    ] as const) {
      assert.ok(Array.isArray(options.documentSelector));
      assert.strictEqual(options.documentSelector.length, 4);
      for (const selector of options.documentSelector as DocumentFilter[]) {
        assert.ok(selector.pattern instanceof RelativePattern);
        assert.strictEqual(
          selector.pattern.baseUri.toString(),
          folder.uri.toString(),
        );
        assert.strictEqual(selector.pattern.pattern, '**/*');
      }
    }
  });

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

class TestConfigHost {
  readonly loadRequests: LoadConfigsRequest[] = [];
  readonly activationRequests: ActivateConfigsRequest[] = [];
  readonly deletedTransactions: string[] = [];
  loadError: Error | undefined;
  changedDuringPrepare = false;
  activation: ConfigModuleActivationPlan = {
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
    prepare?: (activation: ConfigModuleActivationPlan) => Promise<void>,
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
    return {
      transactionId: activation.transactionId,
      eslintPluginEntries: activation.eslintPluginEntries,
    };
  }

  deleteSession(transactionId: string): boolean {
    this.deletedTransactions.push(transactionId);
    return true;
  }
}

class TestPluginPool {
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
  test('the extension watcher leaves gitignore ownership to Go', () => {
    assert.match(CONFIG_REFRESH_WATCH_GLOB, /rslint\.config\.js/);
    assert.match(CONFIG_REFRESH_WATCH_GLOB, /rslint\.config\.mjs/);
    assert.match(CONFIG_REFRESH_WATCH_GLOB, /rslint\.config\.ts/);
    assert.match(CONFIG_REFRESH_WATCH_GLOB, /rslint\.config\.mts/);
    assert.doesNotMatch(CONFIG_REFRESH_WATCH_GLOB, /rslint\.config\.\*/);
    assert.match(CONFIG_REFRESH_WATCH_GLOB, /rslint\.jsonc/);
    assert.match(CONFIG_REFRESH_WATCH_GLOB, /pnpm-lock\.yaml/);
    assert.doesNotMatch(CONFIG_REFRESH_WATCH_GLOB, /\.gitignore/);
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
      aborted: true,
    });
    assert.deepStrictEqual(pool.abortCalls, ['tx-abort']);
    assert.deepStrictEqual(host.deletedTransactions, ['tx-abort']);
  });

  test('native-server restart aborts orphaned state and accepts a new transaction', async () => {
    const host = new TestConfigHost();
    const pool = new TestPluginPool();
    const adapter = new LspConfigTransactionAdapter(
      host,
      pool,
      () => 'fingerprint-restart',
    );

    await adapter.loadConfigs(loadRequest('old-process-tx'));
    await adapter.activateConfigs({
      protocolVersion: CONFIG_DISCOVERY_PROTOCOL_VERSION,
      transactionId: 'old-process-tx',
      effectiveConfigIds: ['root'],
    });
    await adapter.resetForServerRestart();
    assert.deepStrictEqual(pool.abortCalls, ['old-process-tx']);
    assert.deepStrictEqual(host.deletedTransactions, ['old-process-tx']);

    await adapter.loadConfigs(loadRequest('new-process-tx'));
    const activation = await adapter.activateConfigs({
      protocolVersion: CONFIG_DISCOVERY_PROTOCOL_VERSION,
      transactionId: 'new-process-tx',
      effectiveConfigIds: ['root'],
    });
    assert.strictEqual(activation.transactionId, 'new-process-tx');
  });

  test('a later Running state resets orphaned state before requesting an initial catalog', async () => {
    const host = new TestConfigHost();
    const pool = new TestPluginPool();
    const adapter = new LspConfigTransactionAdapter(
      host,
      pool,
      () => 'fingerprint-restart',
    );
    await adapter.loadConfigs(loadRequest('orphaned-tx'));

    const events: string[] = [];
    const recovery = recoverConfigDiscoveryOnServerState(
      State.Running,
      async (reason, beforeRequest) => {
        events.push(`request:${reason}`);
        await beforeRequest?.(adapter);
        events.push('send');
      },
    );
    await recovery;

    assert.deepStrictEqual(events, ['request:initial', 'send']);
    assert.deepStrictEqual(pool.abortCalls, ['orphaned-tx']);
    assert.deepStrictEqual(host.deletedTransactions, ['orphaned-tx']);

    let stoppedRefresh = false;
    const ignored = recoverConfigDiscoveryOnServerState(
      State.Stopped,
      async () => {
        stoppedRefresh = true;
      },
    );
    assert.strictEqual(ignored, undefined);
    assert.strictEqual(stoppedRefresh, false);
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

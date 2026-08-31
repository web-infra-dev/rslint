import * as assert from 'node:assert';
import { PassThrough } from 'node:stream';
import {
  ConfigurationTarget,
  extensions,
  Uri,
  ViewColumn,
  workspace,
  type OutputChannel,
  type WorkspaceFolder,
} from 'vscode';
import {
  LanguageClient,
  State,
  type StreamInfo,
} from 'vscode-languageclient/node';
import {
  createProtocolConnection,
  ExitNotification,
  InitializeRequest,
  SetTraceNotification,
  ShutdownRequest,
  type ProtocolConnection,
  type TraceValues,
} from 'vscode-languageserver/node';
import { createLanguageClientOptions } from '../../src/Rslint';
import {
  createRuntimeTraceLabel,
  SharedTraceOutputChannel,
} from '../../src/SharedTraceOutputChannel';

class MemoryOutputChannel implements OutputChannel {
  public readonly name = 'Rslint trace test';
  public value = '';
  public clearCalls = 0;
  public disposeCalls = 0;
  public hideCalls = 0;
  public readonly showCalls: Array<
    readonly [ViewColumn | boolean | undefined, boolean | undefined]
  > = [];

  public append(value: string): void {
    this.value += value;
  }

  public appendLine(value: string): void {
    this.value += `${value}\n`;
  }

  public replace(value: string): void {
    this.value = value;
  }

  public clear(): void {
    this.clearCalls++;
    this.value = '';
  }

  public show(preserveFocus?: boolean): void;
  public show(column?: ViewColumn, preserveFocus?: boolean): void;
  public show(
    columnOrPreserveFocus?: ViewColumn | boolean,
    preserveFocus?: boolean,
  ): void {
    this.showCalls.push([columnOrPreserveFocus, preserveFocus]);
  }

  public hide(): void {
    this.hideCalls++;
  }

  public dispose(): void {
    this.disposeCalls++;
  }
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

async function eventually(
  predicate: () => boolean,
  description: string,
): Promise<void> {
  const deadline = Date.now() + 3_000;
  while (Date.now() < deadline) {
    if (predicate()) return;
    await new Promise((resolve) => setTimeout(resolve, 10));
  }
  assert.fail(`Timed out waiting for ${description}`);
}

interface TraceClientHarness {
  readonly client: LanguageClient;
  readonly initialTrace: TraceValues | undefined;
  readonly serverStarts: number;
  readonly traceUpdates: TraceValues[];
  dispose(): Promise<void>;
}

function createTraceClientHarness(
  name: string,
  workspaceFolder: WorkspaceFolder,
  serverLog: OutputChannel,
  traceView: OutputChannel,
): TraceClientHarness {
  let serverStarts = 0;
  let initialTrace: TraceValues | undefined;
  const traceUpdates: TraceValues[] = [];
  let connection: ProtocolConnection | undefined;
  let clientInput: PassThrough | undefined;
  let clientOutput: PassThrough | undefined;
  const client = new LanguageClient(
    'rslint',
    name,
    async (): Promise<StreamInfo> => {
      serverStarts++;
      clientInput = new PassThrough();
      clientOutput = new PassThrough();
      connection = createProtocolConnection(clientOutput, clientInput);
      connection.onRequest(InitializeRequest.type, (params) => {
        initialTrace = params.trace;
        return { capabilities: {} };
      });
      connection.onNotification(SetTraceNotification.type, (params) => {
        traceUpdates.push(params.value);
      });
      connection.onRequest(ShutdownRequest.type, () => undefined);
      connection.onNotification(ExitNotification.type, () => undefined);
      connection.listen();
      return { reader: clientInput, writer: clientOutput };
    },
    createLanguageClientOptions(workspaceFolder, serverLog, traceView),
  );

  return {
    client,
    traceUpdates,
    get initialTrace() {
      return initialTrace;
    },
    get serverStarts() {
      return serverStarts;
    },
    async dispose(): Promise<void> {
      try {
        await client.dispose();
      } finally {
        connection?.dispose();
        clientInput?.destroy();
        clientOutput?.destroy();
      }
    },
  };
}

suite('shared LSP trace output', () => {
  test('declares one window-wide trace setting', () => {
    const extension = extensions.getExtension('rstack.rslint');
    assert.ok(extension, 'Rslint extension must be installed in the test host');
    const manifest: unknown = extension.packageJSON;
    assert.ok(isRecord(manifest));
    assert.ok(isRecord(manifest.contributes));
    assert.ok(isRecord(manifest.contributes.configuration));
    assert.ok(isRecord(manifest.contributes.configuration.properties));
    const traceSetting =
      manifest.contributes.configuration.properties['rslint.trace.server'];
    assert.ok(isRecord(traceSetting));
    assert.strictEqual(traceSetting.scope, 'window');
    assert.strictEqual(traceSetting.default, 'off');
  });

  test('distinguishes same-name workspaces and same-version core copies', () => {
    const leftWorkspace = createRuntimeTraceLabel(
      'file:///repo/twins/left/app',
      '/repo/node_modules/@rslint/core',
    );
    const rightWorkspace = createRuntimeTraceLabel(
      'file:///repo/twins/right/app',
      '/repo/node_modules/@rslint/core',
    );
    const nestedCore = createRuntimeTraceLabel(
      'file:///repo/twins/left/app',
      '/repo/packages/nested/node_modules/@rslint/core',
    );

    assert.notStrictEqual(leftWorkspace, rightWorkspace);
    assert.notStrictEqual(leftWorkspace, nestedCore);
    assert.strictEqual(
      createRuntimeTraceLabel('file:///repo\nforged', '/core\rforged'),
      'workspace="file:///repo\\nforged" core="/core\\rforged"',
    );
    assert.strictEqual(
      createRuntimeTraceLabel('file:///repo\u2028forged', '/core\u2029forged'),
      'workspace="file:///repo\\u2028forged" core="/core\\u2029forged"',
    );
  });

  test('updates every workspace and core runtime without replacing clients', async function () {
    this.timeout(15_000);
    const workspaceFolder = workspace.workspaceFolders?.[0];
    assert.ok(workspaceFolder, 'test requires a workspace folder');
    const secondWorkspaceFolder: WorkspaceFolder = {
      index: workspaceFolder.index + 1,
      name: workspaceFolder.name,
      uri: Uri.joinPath(workspaceFolder.uri, 'second-runtime'),
    };
    const configuration = workspace.getConfiguration('rslint');
    const originalWorkspaceValue =
      configuration.inspect<string>('trace.server')?.workspaceValue;
    const serverLogs = [new MemoryOutputChannel(), new MemoryOutputChannel()];
    const traceLog = new MemoryOutputChannel();
    const shared = new SharedTraceOutputChannel(traceLog);
    const firstLabel = createRuntimeTraceLabel(
      workspaceFolder.uri.toString(),
      '/virtual/core-a',
    );
    const secondLabel = createRuntimeTraceLabel(
      secondWorkspaceFolder.uri.toString(),
      '/virtual/core-b',
    );
    const traceViews = [
      shared.forRuntime(firstLabel),
      shared.forRuntime(secondLabel),
    ];
    const harnesses = [
      createTraceClientHarness(
        'Rslint live trace test A',
        workspaceFolder,
        serverLogs[0],
        traceViews[0],
      ),
      createTraceClientHarness(
        'Rslint live trace test B',
        secondWorkspaceFolder,
        serverLogs[1],
        traceViews[1],
      ),
    ];

    try {
      await configuration.update(
        'trace.server',
        'off',
        ConfigurationTarget.Workspace,
      );
      await harnesses[0].client.start();
      assert.strictEqual(harnesses[0].initialTrace, 'off');
      assert.strictEqual(harnesses[0].serverStarts, 1);
      assert.strictEqual(harnesses[1].serverStarts, 0);

      await harnesses[0].client.sendNotification('rslint/traceProbe', {
        phase: 'off',
      });
      assert.strictEqual(traceLog.value, '');

      await configuration.update(
        'trace.server',
        'messages',
        ConfigurationTarget.Workspace,
      );
      await eventually(
        () => harnesses[0].traceUpdates.includes('messages'),
        'the messages trace notification for the running runtime',
      );
      // A runtime created after the setting changed must inherit the current
      // level in initialize rather than waiting for another settings event.
      await harnesses[1].client.start();
      assert.strictEqual(harnesses[1].initialTrace, 'messages');
      assert.strictEqual(harnesses[1].serverStarts, 1);
      await Promise.all(
        harnesses.map(({ client }, index) =>
          client.sendNotification('rslint/traceProbe', {
            phase: `messages-${String(index)}`,
          }),
        ),
      );
      assert.ok(traceLog.value.includes(`[${firstLabel}] `));
      assert.ok(traceLog.value.includes(`[${secondLabel}] `));
      assert.ok(
        serverLogs.every(
          (serverLog) =>
            !serverLog.value.includes(`[${firstLabel}]`) &&
            !serverLog.value.includes(`[${secondLabel}]`),
        ),
      );

      await configuration.update(
        'trace.server',
        'verbose',
        ConfigurationTarget.Workspace,
      );
      await eventually(
        () =>
          harnesses.every(({ traceUpdates }) =>
            traceUpdates.includes('verbose'),
          ),
        'the verbose trace notification for every runtime',
      );
      await Promise.all(
        harnesses.map(({ client }, index) =>
          client.sendNotification('rslint/traceProbe', {
            phase: `verbose-payload-${String(index)}`,
          }),
        ),
      );
      assert.ok(traceLog.value.includes('verbose-payload-0'));
      assert.ok(traceLog.value.includes('verbose-payload-1'));

      await configuration.update(
        'trace.server',
        'off',
        ConfigurationTarget.Workspace,
      );
      await eventually(
        () =>
          harnesses.every(({ traceUpdates }) => traceUpdates.at(-1) === 'off'),
        'the disabled trace notification for every runtime',
      );
      const traceLengthAfterDisable = traceLog.value.length;
      await Promise.all(
        harnesses.map(({ client }) =>
          client.sendNotification('rslint/traceProbe', {
            phase: 'disabled-again',
          }),
        ),
      );
      assert.strictEqual(traceLog.value.length, traceLengthAfterDisable);
      for (const harness of harnesses) {
        assert.strictEqual(harness.client.state, State.Running);
        assert.strictEqual(harness.serverStarts, 1);
      }
    } finally {
      try {
        await Promise.all(harnesses.map((harness) => harness.dispose()));
      } finally {
        traceViews.forEach((traceView) => traceView.dispose());
        serverLogs.forEach((serverLog) => serverLog.dispose());
        traceLog.dispose();
        await configuration.update(
          'trace.server',
          originalWorkspaceValue,
          ConfigurationTarget.Workspace,
        );
      }
    }
  });

  test('prefixes append sequences and every physical line', () => {
    const channel = new MemoryOutputChannel();
    const shared = new SharedTraceOutputChannel(channel);
    const label = createRuntimeTraceLabel('file:///workspace', '/core/a');
    const runtime = shared.forRuntime(label);

    runtime.append('request: ');
    runtime.appendLine('initialize');
    runtime.appendLine('first payload line\nsecond payload line');

    assert.strictEqual(
      channel.value,
      `[${label}] request: initialize\n` +
        `[${label}] first payload line\n` +
        `[${label}] second payload line\n`,
    );
  });

  test('never merges partial lines from interleaved runtimes', () => {
    const channel = new MemoryOutputChannel();
    const shared = new SharedTraceOutputChannel(channel);
    const firstLabel = createRuntimeTraceLabel('file:///first', '/core/a');
    const secondLabel = createRuntimeTraceLabel('file:///second', '/core/b');
    const first = shared.forRuntime(firstLabel);
    const second = shared.forRuntime(secondLabel);

    first.append('partial request');
    second.appendLine('complete response');
    first.appendLine('continued request');

    assert.strictEqual(
      channel.value,
      `[${firstLabel}] partial request\n` +
        `[${secondLabel}] complete response\n` +
        `[${firstLabel}] continued request\n`,
    );
  });

  test('keeps shared channel ownership in the extension', () => {
    const channel = new MemoryOutputChannel();
    const shared = new SharedTraceOutputChannel(channel);
    const firstLabel = createRuntimeTraceLabel('file:///first', '/core/a');
    const secondLabel = createRuntimeTraceLabel('file:///second', '/core/b');
    const first = shared.forRuntime(firstLabel);
    const second = shared.forRuntime(secondLabel);

    assert.strictEqual(first.name, channel.name);
    first.appendLine('discarded by replacement');
    second.replace('replacement\ncontinued');
    assert.strictEqual(
      channel.value,
      `[${secondLabel}] replacement\n[${secondLabel}] continued`,
    );

    first.clear();
    assert.strictEqual(channel.value, '');
    assert.strictEqual(channel.clearCalls, 1);

    first.dispose();
    first.appendLine('must not be written');
    second.appendLine('still active');
    assert.strictEqual(channel.value, `[${secondLabel}] still active\n`);
    assert.strictEqual(channel.disposeCalls, 0);

    second.show(true);
    second.show(ViewColumn.Two, false);
    second.hide();
    assert.deepStrictEqual(channel.showCalls, [
      [true, undefined],
      [ViewColumn.Two, false],
    ]);
    assert.strictEqual(channel.hideCalls, 1);

    second.dispose();
    assert.strictEqual(channel.disposeCalls, 0);
  });

  test('rejects an unattributed runtime view', () => {
    const shared = new SharedTraceOutputChannel(new MemoryOutputChannel());
    assert.throws(() => shared.forRuntime(''), /must not be empty/);
  });
});

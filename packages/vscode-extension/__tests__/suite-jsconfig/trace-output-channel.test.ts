import * as assert from 'node:assert';
import { PassThrough } from 'node:stream';
import {
  ConfigurationTarget,
  extensions,
  Uri,
  workspace,
  type OutputChannel,
  type ViewColumn,
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

class MemoryOutputChannel implements OutputChannel {
  public readonly name = 'Rslint trace test';
  public value = '';
  public disposeCalls = 0;

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
    this.value = '';
  }

  public show(preserveFocus?: boolean): void;
  public show(column?: ViewColumn, preserveFocus?: boolean): void;
  public show(
    _columnOrPreserveFocus?: ViewColumn | boolean,
    _preserveFocus?: boolean,
  ): void {}

  public hide(): void {}

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
  traceOutputChannel: OutputChannel,
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
    createLanguageClientOptions(workspaceFolder, serverLog, traceOutputChannel),
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

suite('window-wide LSP trace output', () => {
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
    const serverLogs = [
      new MemoryOutputChannel(),
      new MemoryOutputChannel(),
      new MemoryOutputChannel(),
    ];
    const traceLog = new MemoryOutputChannel();
    const harnesses = [
      createTraceClientHarness(
        'Rslint live trace test A',
        workspaceFolder,
        serverLogs[0],
        traceLog,
      ),
      createTraceClientHarness(
        'Rslint live trace test B',
        secondWorkspaceFolder,
        serverLogs[1],
        traceLog,
      ),
      // A second client for the same workspace models documents that resolve
      // to another physical core installation within that workspace.
      createTraceClientHarness(
        'Rslint live trace test C',
        workspaceFolder,
        serverLogs[2],
        traceLog,
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
      assert.strictEqual(harnesses[2].serverStarts, 0);

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
      await Promise.all([
        harnesses[1].client.start(),
        harnesses[2].client.start(),
      ]);
      for (const harness of harnesses.slice(1)) {
        assert.strictEqual(harness.initialTrace, 'messages');
        assert.strictEqual(harness.serverStarts, 1);
      }
      await Promise.all(
        harnesses.map(({ client }, index) =>
          client.sendNotification('rslint/traceProbe', {
            phase: `messages-${String(index)}`,
          }),
        ),
      );
      assert.ok(
        traceLog.value.split('rslint/traceProbe').length - 1 >=
          harnesses.length,
        'every runtime should write protocol messages to the shared channel',
      );
      assert.ok(
        serverLogs.every(
          (serverLog) => !serverLog.value.includes('rslint/traceProbe'),
        ),
        'protocol traces must not fall back to a server log',
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
      for (const index of harnesses.keys()) {
        assert.ok(traceLog.value.includes(`verbose-payload-${String(index)}`));
      }

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
      let traceDisposeCallsAfterClientClose: number | undefined;
      try {
        await Promise.all(harnesses.map((harness) => harness.dispose()));
      } finally {
        traceDisposeCallsAfterClientClose = traceLog.disposeCalls;
        serverLogs.forEach((serverLog) => serverLog.dispose());
        traceLog.dispose();
        await configuration.update(
          'trace.server',
          originalWorkspaceValue,
          ConfigurationTarget.Workspace,
        );
      }
      assert.strictEqual(
        traceDisposeCallsAfterClientClose,
        0,
        'language clients must not dispose the extension-owned channel',
      );
    }
  });
});

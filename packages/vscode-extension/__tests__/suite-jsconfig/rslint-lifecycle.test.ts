import * as assert from 'node:assert';
import { window } from 'vscode';
import { CloseAction, ErrorAction, State } from 'vscode-languageclient/node';
import { LanguageServerProcessOwner } from '../../src/LanguageServerProcessOwner';
import {
  disposeLanguageClient,
  ManagedLanguageClient,
  shouldResetDocumentSessionOnServerState,
  waitForPromiseSettlement,
} from '../../src/Rslint';

const FORCE_KILL_CHILD =
  "process.on('SIGTERM', () => undefined); setTimeout(() => process.stdout.write('ready\\n'), 20); setInterval(() => undefined, 1_000)";

function processOwner(source = FORCE_KILL_CHILD): LanguageServerProcessOwner {
  return new LanguageServerProcessOwner(
    process.execPath,
    ['-e', source],
    process.cwd(),
    { ...process.env, ELECTRON_RUN_AS_NODE: '1' },
  );
}

async function waitForOutput(
  stream: NodeJS.ReadableStream,
  expected: string,
): Promise<void> {
  await new Promise<void>((resolve) => {
    const onData = (chunk: Buffer): void => {
      if (!chunk.toString().includes(expected)) return;
      stream.removeListener('data', onData);
      resolve();
    };
    stream.on('data', onData);
  });
}

suite('Rslint lifecycle', () => {
  test('disposes diagnostics without waiting for a Starting client handshake', async () => {
    let clientDisposeCalls = 0;
    let diagnosticDisposeCalls = 0;

    await disposeLanguageClient({
      state: State.Starting,
      diagnostics: {
        dispose() {
          diagnosticDisposeCalls++;
        },
      },
      async dispose() {
        clientDisposeCalls++;
        throw new Error('client is still starting');
      },
    });

    assert.strictEqual(clientDisposeCalls, 1);
    assert.strictEqual(diagnosticDisposeCalls, 1);
  });

  test('reports a Running client disposal failure after disposing diagnostics', async () => {
    let diagnosticDisposeCalls = 0;

    await assert.rejects(
      disposeLanguageClient({
        state: State.Running,
        diagnostics: {
          dispose() {
            diagnosticDisposeCalls++;
          },
        },
        async dispose() {
          throw new Error('shutdown failed');
        },
      }),
      /shutdown failed/,
    );

    assert.strictEqual(diagnosticDisposeCalls, 1);
  });

  test('awaits native child termination and rejects later restarts', async () => {
    const owner = processOwner();
    const child = await owner.start();
    await waitForOutput(child.stdout, 'ready');

    await owner.close();

    assert.ok(
      child.exitCode !== null || child.signalCode !== null,
      'close should resolve only after the child exits',
    );
    if (process.platform !== 'win32') {
      assert.strictEqual(child.signalCode, 'SIGKILL');
    }
    await assert.rejects(owner.start(), /process owner is closing/);
  });

  test('terminates the prior child before an automatic restart spawn', async () => {
    const owner = processOwner();
    const first = await owner.start();
    await waitForOutput(first.stdout, 'ready');

    const second = await owner.start();

    assert.notStrictEqual(first.pid, second.pid);
    assert.ok(
      first.exitCode !== null || first.signalCode !== null,
      'the old child must close before start returns its replacement',
    );
    await waitForOutput(second.stdout, 'ready');
    await owner.close();
  });

  test('settles a hung initialize tail after forced transport close', async () => {
    const source =
      "process.on('SIGTERM', () => undefined); process.stdin.on('data', () => process.stderr.write('initialize-request\\n')); setTimeout(() => process.stderr.write('ready\\n'), 20); setInterval(() => undefined, 1_000)";
    const owner = processOwner(source);
    const outputChannel = window.createOutputChannel(
      `Rslint lifecycle hang ${Date.now()}`,
    );
    let closing = false;
    let spawnCount = 0;
    let initializeReceived!: () => void;
    const receivedInitialize = new Promise<void>((resolve) => {
      initializeReceived = resolve;
    });
    const client = new ManagedLanguageClient(
      `rslint-lifecycle-hang-${Date.now()}`,
      'Rslint lifecycle hang probe',
      async () => {
        const child = await owner.start();
        spawnCount++;
        child.stderr.on('data', (chunk: Buffer) => {
          if (chunk.toString().includes('initialize-request')) {
            initializeReceived();
          }
        });
        await waitForOutput(child.stderr, 'ready');
        return child;
      },
      {
        documentSelector: [],
        outputChannel,
        errorHandler: {
          error: () => ({ action: ErrorAction.Shutdown }),
          closed: () => ({
            action: closing ? CloseAction.DoNotRestart : CloseAction.Restart,
            handled: closing,
          }),
        },
      },
    );

    const startPromise = client.start();
    void startPromise.catch(() => undefined);
    await waitForPromiseSettlement(
      receivedInitialize,
      2_000,
      'initialize request probe',
    );
    assert.strictEqual(client.state, State.Starting);

    closing = true;
    owner.beginClose();
    await disposeLanguageClient(client);
    await owner.close();
    await waitForPromiseSettlement(
      startPromise,
      2_000,
      'hung language client start',
    );

    assert.strictEqual(client.state, State.Stopped);
    assert.strictEqual(spawnCount, 1, 'closing must suppress restart');
    outputChannel.dispose();
  });

  test('resets document sessions as soon as a Running server exits', () => {
    assert.strictEqual(
      shouldResetDocumentSessionOnServerState(State.Running, State.Stopped),
      true,
    );
    assert.strictEqual(
      shouldResetDocumentSessionOnServerState(State.Stopped, State.Starting),
      false,
    );
    assert.strictEqual(
      shouldResetDocumentSessionOnServerState(State.Starting, State.Running),
      false,
    );
  });
});

import { describe, test, expect } from 'rstack/test';
import path from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';
import { once } from 'node:events';
import { spawn, type ChildProcess } from 'node:child_process';
import { NodeRslintService } from '../src/internal/node.js';
import type { IpcClient } from '../src/ipc/client.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const FAKE = path.resolve(__dirname, './fixtures/fake-api-binary.cjs');

// The fake binary is exec'd directly via its shebang (NodeRslintService spawns
// `rslintPath --api` with no node wrapper). Shebang dispatch doesn't apply on
// win32, so skip there — the reject-all-pending logic under test is pure,
// platform-independent JS (the real win32 path uses the .exe binary).
const suite = process.platform === 'win32' ? describe.skip : describe;

// Reach the private child handle to simulate an external SIGKILL (TS-private,
// present at runtime).
function childOf(svc: NodeRslintService): ChildProcess {
  return (svc as unknown as { process: ChildProcess }).process;
}

suite('NodeRslintService reject-all-pending on crash/terminate', () => {
  test('answers an inbound request without confusing a colliding outbound id', async () => {
    const svc = new NodeRslintService({ rslintPath: FAKE });
    svc.setInboundHandler(async (message) => ({
      kind: message.kind,
      file: message.data.files[0].path,
    }));
    await expect(svc.sendMessage('reverse', {})).resolves.toEqual({
      reverseKind: 'response',
      reverseData: { kind: 'pluginLint', file: 'probe.ts' },
    });
    svc.terminate();
  });

  test('returns an error frame when an inbound handler throws', async () => {
    const svc = new NodeRslintService({ rslintPath: FAKE });
    svc.setInboundHandler(() => {
      throw new Error('plugin host failed');
    });
    await expect(svc.sendMessage('reverse', {})).resolves.toEqual({
      reverseKind: 'error',
      reverseData: { message: 'plugin host failed' },
    });
    svc.terminate();
  });

  test('returns a clear error frame when no inbound handler is installed', async () => {
    const svc = new NodeRslintService({ rslintPath: FAKE });
    const result = await svc.sendMessage('reverse', {});
    expect(result.reverseKind).toBe('error');
    expect(result.reverseData.message).toMatch(
      /no inbound handler.*pluginLint/,
    );
    svc.terminate();
  });

  test('rejects in-flight requests when the process crashes', async () => {
    const svc = new NodeRslintService({ rslintPath: FAKE });
    const inflight = svc.sendMessage('lint', {}); // never answered → in-flight
    svc.sendMessage('crash', {}).catch(() => {
      /* the fake exits before acking — the rejection is expected, ignore it */
    }); // make the fake exit(42)
    // Either EOF or child exit may arrive first; both must reject immediately.
    await expect(inflight).rejects.toThrow(
      /exited unexpectedly|peer closed input stream/,
    );
  });

  test('rejects in-flight requests on an external SIGKILL', async () => {
    const svc = new NodeRslintService({ rslintPath: FAKE });
    const inflight = svc.sendMessage('lint', {});
    childOf(svc).kill('SIGKILL');
    await expect(inflight).rejects.toThrow(
      /exited unexpectedly|peer closed input stream/,
    );
  });

  test('rejects in-flight requests on terminate()', async () => {
    const svc = new NodeRslintService({ rslintPath: FAKE });
    const inflight = svc.sendMessage('lint', {});
    svc.terminate();
    await expect(inflight).rejects.toThrow(/terminated/);
  });

  test('rejects on stdout EOF even while the Go process stays alive', async () => {
    const svc = new NodeRslintService({ rslintPath: FAKE });
    try {
      await expect(svc.sendMessage('close-output', {})).rejects.toThrow(
        /peer closed input stream/,
      );
      expect(childOf(svc).exitCode).toBeNull();
      await expect(svc.sendMessage('lint', {})).rejects.toThrow(
        /closed|no longer running/,
      );
    } finally {
      svc.terminate();
    }
  });

  test('owns late pipe errors until the terminated child has closed', async () => {
    const svc = new NodeRslintService({ rslintPath: FAKE });
    await svc.sendMessage('handshake', { version: '3.1.0' });
    const child = childOf(svc);
    const closed = once(child, 'close');
    svc.terminate();
    // IpcClient has detached its listeners, but the child still owns its
    // pipes. A queued asynchronous write may emit EPIPE in this interval.
    expect(() =>
      child.stdin!.emit('error', new Error('late EPIPE')),
    ).not.toThrow();
    expect(() =>
      child.stdout!.emit('error', new Error('late read error')),
    ).not.toThrow();
    await closed;
    expect(child.stdin!.listenerCount('error')).toBe(0);
    expect(child.stdout!.listenerCount('error')).toBe(0);
  });

  test('does not harm a normal request/response round-trip', async () => {
    const svc = new NodeRslintService({ rslintPath: FAKE });
    await expect(
      svc.sendMessage('handshake', { version: '3.1.0' }),
    ).resolves.toEqual({
      version: '3.1.0',
      ok: true,
      capabilities: ['reversePluginLint'],
    });
    svc.terminate();
  });

  test('rejects (does not hang) a request sent after the service is dead', async () => {
    const svc = new NodeRslintService({ rslintPath: FAKE });
    svc.terminate();
    await expect(svc.sendMessage('lint', {})).rejects.toThrow(
      /no longer running/,
    );
  });

  test('silent exit resolves only the exit request and rejects other pending work', async () => {
    // The peer exits without acknowledging shutdown. The API adapter retains
    // the best-effort exit result, while IPC still rejects unfinished linting.
    process.env.RSLINT_FAKE_EXIT_SILENT = '1';
    try {
      const svc = new NodeRslintService({ rslintPath: FAKE });
      await svc.sendMessage('handshake', { version: '3.1.0' });
      const [lint, exit] = await Promise.allSettled([
        svc.sendMessage('lint', {}),
        svc.sendMessage('exit', {}),
      ]);
      expect(lint.status).toBe('rejected');
      expect(exit).toEqual({ status: 'fulfilled', value: null });
    } finally {
      delete process.env.RSLINT_FAKE_EXIT_SILENT;
    }
  });

  test('preserves an exit rejection when EOF arrives before its catch continuation', async () => {
    const svc = new NodeRslintService({ rslintPath: FAKE });
    const client = (svc as unknown as { client: IpcClient }).client;
    const originalSend = client.sendRequest;
    // Delay delivery of the real peer rejection until EOF. A concurrent
    // transport close must not turn an application error into success.
    client.sendRequest = function (kind, data, attachments) {
      return originalSend
        .call(this, kind, data, attachments)
        .catch(async (error: unknown) => {
          await this.done;
          throw error;
        });
    } as typeof client.sendRequest;
    try {
      await expect(svc.sendMessage('exit', { reject: true })).rejects.toThrow(
        'exit rejected',
      );
    } finally {
      svc.terminate();
    }
  });

  test('rejects in-flight requests on a spawn failure (bad binary path)', async () => {
    // A nonexistent binary makes spawn emit 'error' (ENOENT) asynchronously,
    // after sendMessage's stdin.write returns — the ONLY reject-all-pending path
    // driven by process.on('error'). The other tests exercise the 'exit' and
    // terminate() paths; without this one a spawn failure would hang in-flight
    // promises forever with no test signal.
    const svc = new NodeRslintService({
      rslintPath: '/nonexistent/rslint-binary-xyz',
    });
    const inflight = svc.sendMessage('lint', {});
    await expect(inflight).rejects.toThrow(/rslint process error/);
  });

  test('EOF releases a resident process with an unanswered inbound request', async () => {
    // Use a separate Node process: the test runner's own handles would hide a
    // leaked child/pipe reference. Finish the handshake before asking the fake
    // for an unsolicited request, so no outbound finally can drive cleanup.
    const entry = pathToFileURL(
      path.resolve(__dirname, '../dist/internal.js'),
    ).href;
    const script = `
      import { NodeRslintService } from ${JSON.stringify(entry)};
      const keepAlive = setInterval(() => {}, 1000);
      const service = new NodeRslintService({ rslintPath: ${JSON.stringify(FAKE)} });
      await service.sendMessage('handshake', { version: '3.1.0' });
      service.setInboundHandler(() => {
        clearInterval(keepAlive);
        process.stdout.write('INBOUND_PENDING\\n');
        return new Promise(() => {});
      });
      // Reach the transport only to trigger a request without creating an
      // outbound API pending. The service still owns all lifecycle decisions.
      service.client.done.then(() => process.stdout.write('IPC_CLOSED\\n'));
      service.client.sendNotification('reverse-close-output', {});
    `;
    const child = spawn(
      process.execPath,
      ['--input-type=module', '-e', script],
      {
        stdio: ['ignore', 'pipe', 'pipe'],
      },
    );
    let stdout = '';
    let stderr = '';
    child.stdout.on('data', (chunk) => {
      stdout += String(chunk);
    });
    child.stderr.on('data', (chunk) => {
      stderr += String(chunk);
    });
    let timedOut = false;
    const watchdog = setTimeout(() => {
      timedOut = true;
      child.kill('SIGKILL');
    }, 5_000);
    try {
      const [code, signal] = await once(child, 'close');
      expect({ code, signal, timedOut, stderr }).toEqual({
        code: 0,
        signal: null,
        timedOut: false,
        stderr: '',
      });
      expect(stdout).toBe('INBOUND_PENDING\nIPC_CLOSED\n');
    } finally {
      clearTimeout(watchdog);
      child.kill('SIGKILL');
    }
  }, 10_000);
});

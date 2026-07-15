import { describe, test, expect } from '@rstest/core';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { NodeRslintService } from '../src/internal/node.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const FAKE = path.resolve(__dirname, './fixtures/fake-api-binary.cjs');

// The fake binary is exec'd directly via its shebang (NodeRslintService spawns
// `rslintPath --api` with no node wrapper). Shebang dispatch doesn't apply on
// win32, so skip there — the reject-all-pending logic under test is pure,
// platform-independent JS (the real win32 path uses the .exe binary).
const suite = process.platform === 'win32' ? describe.skip : describe;

// Reach the private child handle to simulate an external SIGKILL (TS-private,
// present at runtime).
function childOf(svc: NodeRslintService): { kill: (sig?: string) => void } {
  return (svc as unknown as { process: { kill: (sig?: string) => void } })
    .process;
}

suite('NodeRslintService reject-all-pending on crash/terminate', () => {
  test('an orphaned reverse handler does not pin the process after its outer request settles', async () => {
    const { spawn } = await import('node:child_process');
    const script = path.resolve(
      __dirname,
      './fixtures/orphan-reverse-exit.mjs',
    );
    const child = spawn(process.execPath, [script], {
      cwd: path.resolve(__dirname, '..'),
      stdio: 'inherit',
    });
    const code = await new Promise<number | 'TIMEOUT-HANG'>((resolve) => {
      const timer = setTimeout(() => {
        child.kill('SIGKILL');
        resolve('TIMEOUT-HANG');
      }, 10_000);
      child.on('exit', (exitCode) => {
        clearTimeout(timer);
        resolve(exitCode ?? 1);
      });
    });
    expect(code).toBe(0);
  });

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
    // Asserts the exit handler (not some watchdog) rejected it.
    await expect(inflight).rejects.toThrow(/exited unexpectedly/);
  });

  test('rejects in-flight requests on an external SIGKILL', async () => {
    const svc = new NodeRslintService({ rslintPath: FAKE });
    const inflight = svc.sendMessage('lint', {});
    childOf(svc).kill('SIGKILL');
    await expect(inflight).rejects.toThrow(/exited unexpectedly/);
  });

  test('rejects in-flight requests on terminate()', async () => {
    const svc = new NodeRslintService({ rslintPath: FAKE });
    const inflight = svc.sendMessage('lint', {});
    svc.terminate();
    await expect(inflight).rejects.toThrow(/terminated/);
  });

  test('does not harm a normal request/response round-trip', async () => {
    const svc = new NodeRslintService({ rslintPath: FAKE });
    await expect(
      svc.sendMessage('handshake', { version: '3.0.0' }),
    ).resolves.toEqual({
      version: '3.0.0',
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

  test('graceful exit resolves the request even when the peer exits before acking', async () => {
    // Silent-exit fake: the peer exits(0) WITHOUT acking 'exit', so the process
    // 'exit' event fires before any response is read. The 'exit' kind flags
    // closing, so the exit handler RESOLVES the pending instead of rejecting it
    // — this is what keeps RSLintService.close()'s awaited 'exit' request from
    // rejecting into an unhandledRejection. (Pre-fix, the exit handler rejected
    // unconditionally and this would reject with /exited unexpectedly/.)
    process.env.RSLINT_FAKE_EXIT_SILENT = '1';
    try {
      const svc = new NodeRslintService({ rslintPath: FAKE });
      await svc.sendMessage('handshake', { version: '3.0.0' });
      await expect(svc.sendMessage('exit', {})).resolves.toBeNull();
    } finally {
      delete process.env.RSLINT_FAKE_EXIT_SILENT;
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
});

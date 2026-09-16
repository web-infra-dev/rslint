import { afterEach, describe, test, expect, rs } from 'rstack/test';
import * as childProcess from 'child_process';
import { EventEmitter } from 'node:events';
import { mkdtemp, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { PassThrough } from 'node:stream';
import path from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';
import { lint, NodeRslintService } from '../src/internal/node.js';
import { RSLintService } from '../src/service/service.js';

rs.mock('child_process', { spy: true });

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
    await svc.terminate();
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
    await svc.terminate();
  });

  test('returns a clear error frame when no inbound handler is installed', async () => {
    const svc = new NodeRslintService({ rslintPath: FAKE });
    const result = await svc.sendMessage('reverse', {});
    expect(result.reverseKind).toBe('error');
    expect(result.reverseData.message).toMatch(
      /no inbound handler.*pluginLint/,
    );
    await svc.terminate();
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
    const rejected = expect(inflight).rejects.toThrow(/terminated/);
    await svc.terminate();
    await rejected;
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
    await svc.terminate();
  });

  test('rejects (does not hang) a request sent after the service is dead', async () => {
    const svc = new NodeRslintService({ rslintPath: FAKE });
    await svc.terminate();
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
      await svc.sendMessage('handshake', { version: '3.1.0' });
      await expect(svc.sendMessage('exit', {})).resolves.toBeNull();
      await svc.terminate();
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
    await svc.terminate();
  });
});

// These tests run on Windows too. Control only the process boundary so that
// exit and stdio close can be separated without sleeps or a shebang fixture.
class ControlledChild extends EventEmitter {
  readonly pid = 123;
  readonly stdin = new PassThrough();
  readonly stdout = new PassThrough();
  exitCode: number | null = null;
  signalCode: NodeJS.Signals | null = null;
  killed = false;
  readonly ref = rs.fn(() => this);
  readonly unref = rs.fn(() => this);
  readonly kill = rs.fn((_signal?: NodeJS.Signals) => {
    this.killed = true;
    return true;
  });

  exit(): void {
    this.exitCode = 0;
    this.emit('exit', 0, null);
  }

  close(): void {
    this.stdin.destroy();
    this.stdout.destroy();
    this.emit('close', this.exitCode, this.signalCode);
  }
}

describe('NodeRslintService close completion', () => {
  let child: ControlledChild | undefined;

  function mockChild(): void {
    child = new ControlledChild();
    rs.spyOn(childProcess, 'spawn').mockReturnValue(
      child as unknown as childProcess.ChildProcess,
    );
  }

  function controlledService(): NodeRslintService {
    mockChild();
    return new NodeRslintService();
  }

  async function expectPending(promise: Promise<void>): Promise<void> {
    let settled = false;
    void promise.then(
      () => {
        settled = true;
      },
      () => {
        settled = true;
      },
    );
    // Let settled continuations run while the explicit close gate stays shut.
    await new Promise<void>((resolve) => setImmediate(resolve));
    expect(settled).toBe(false);
  }

  afterEach(() => {
    child?.close();
    child = undefined;
    rs.useRealTimers();
    rs.restoreAllMocks();
  });

  test('waits for close after exit and shares repeated termination', async () => {
    const svc = controlledService();
    const terminating = svc.terminate();
    expect(terminating).toBeInstanceOf(Promise);
    expect(svc.terminate()).toBe(terminating);
    child!.exit();

    await expectPending(terminating);
    child!.close();
    await terminating;
    await svc.terminate();
    expect(child!.kill).toHaveBeenCalledTimes(1);
  });

  test('remembers a child that closed before termination was requested', async () => {
    const svc = controlledService();
    child!.exit();
    child!.close();
    await svc.terminate();
    expect(child!.kill).not.toHaveBeenCalled();
  });

  test('a spawn error rejects requests but does not stand in for close', async () => {
    const svc = controlledService();
    const rejected = expect(svc.sendMessage('lint', {})).rejects.toThrow(
      'rslint process error: spawn failed',
    );
    child!.emit('error', new Error('spawn failed'));
    await rejected;
    const terminating = svc.terminate();
    await expectPending(terminating);
    child!.close();
    await terminating;
  });

  test('escalates a sent SIGTERM and still requires close after SIGKILL', async () => {
    rs.useFakeTimers({ toFake: ['setTimeout', 'clearTimeout'] });
    const svc = controlledService();
    const terminating = svc.terminate();
    expect(child!.killed).toBe(true);
    await rs.advanceTimersByTimeAsync(1_000);
    expect(child!.kill.mock.calls).toEqual([['SIGTERM'], ['SIGKILL']]);
    child!.exit();
    await expectPending(terminating);
    child!.close();
    await terminating;
    expect(rs.getTimerCount()).toBe(0);
  });

  test.each(['throw', 'false'])(
    'rejects when kill returns %s and the child never closes',
    async (mode) => {
      rs.useFakeTimers({ toFake: ['setTimeout', 'clearTimeout'] });
      const svc = controlledService();
      child!.kill.mockImplementation(() => {
        if (mode === 'false') return false;
        throw new Error('permission denied');
      });
      const terminating = svc.terminate();
      const rejected = expect(terminating).rejects.toThrow(
        'did not close after termination',
      );
      await rs.advanceTimersByTimeAsync(31_000);
      await rejected;
      expect(child!.kill.mock.calls).toEqual([['SIGTERM'], ['SIGKILL']]);
      expect(svc.terminate()).toBe(terminating);
      expect(rs.getTimerCount()).toBe(0);
    },
  );

  test.each(['lint failure', 'both failures', 'close failure'])(
    'one-shot lint preserves %s',
    async (mode) => {
      mockChild();
      const lintError = new Error('original lint error');
      const closeError = new Error('shutdown error');
      const lintSpy = rs.spyOn(RSLintService.prototype, 'lint');
      const closeSpy = rs.spyOn(RSLintService.prototype, 'close');
      if (mode === 'close failure')
        lintSpy.mockResolvedValue({
          diagnostics: [],
          errorCount: 0,
          warningCount: 0,
          fixableErrorCount: 0,
          fixableWarningCount: 0,
          fileCount: 0,
          ruleCount: 0,
        });
      else lintSpy.mockRejectedValue(lintError);
      if (mode === 'lint failure') closeSpy.mockResolvedValue(undefined);
      else closeSpy.mockRejectedValue(closeError);

      const result = lint({});
      if (mode === 'both failures') {
        await expect(result).rejects.toMatchObject({
          name: 'AggregateError',
          errors: [lintError, closeError],
        });
      } else {
        await expect(result).rejects.toBe(
          mode === 'lint failure' ? lintError : closeError,
        );
      }
      expect(closeSpy).toHaveBeenCalledTimes(1);
    },
  );
});

describe('NodeRslintService real process cleanup', () => {
  test('keeps a standalone host alive when a late inbound handler releases its references', () => {
    const entry = pathToFileURL(path.resolve(__dirname, '../dist/internal.js'));
    const script = `
      import childProcess from 'node:child_process';
      import { syncBuiltinESMExports } from 'node:module';
      import { EventEmitter } from 'node:events';
      import { PassThrough } from 'node:stream';
      const child = Object.assign(new EventEmitter(), {
        stdin: new PassThrough(), stdout: new PassThrough(),
        exitCode: null, signalCode: null, pid: 123,
        ref() {}, unref() {}, kill() { return true; },
      });
      childProcess.spawn = () => child;
      syncBuiltinESMExports();
      const { NodeRslintService } = await import(${JSON.stringify(entry.href)});
      const service = new NodeRslintService();
      let releaseInbound;
      service.setInboundHandler(() => new Promise(resolve => {
        releaseInbound = resolve;
      }));
      const message = Buffer.from(JSON.stringify({ id: 1, kind: 'pluginLint', data: {} }));
      const header = Buffer.alloc(4);
      header.writeUInt32LE(message.length);
      child.stdout.write(Buffer.concat([header, message]));
      await new Promise(resolve => setImmediate(resolve));
      const terminating = service.terminate();
      releaseInbound({});
      await new Promise(resolve => setImmediate(resolve));
      // Only the production shutdown wait can keep the loop alive now.
      setImmediate(() => {
        child.exitCode = 0;
        child.emit('exit', 0, null);
        child.emit('close', 0, null);
      }).unref();
      await terminating;
      console.log('CLOSED');
    `;
    const result = childProcess.spawnSync(
      process.execPath,
      ['--input-type=module', '--eval', script],
      { encoding: 'utf8', timeout: 30 * 60_000 },
    );
    expect(result.error).toBeUndefined();
    expect(result.signal).toBeNull();
    expect(result.status, result.stderr).toBe(0);
    expect(result.stdout.trim()).toBe('CLOSED');
    expect(result.stderr).toBe('');
  });

  test('closes the real child before its working directory is removed', async () => {
    const cwd = await mkdtemp(path.join(tmpdir(), 'rslint-close-'));
    const backend = new NodeRslintService({ workingDirectory: cwd });
    const service = new RSLintService(backend);
    const child = (backend as unknown as { process: childProcess.ChildProcess })
      .process;
    let closed = false;
    child.once('close', () => {
      closed = true;
    });
    try {
      const file = path.join(cwd, 'probe.ts');
      const result = await service.lint({
        config: [{ rules: { 'no-debugger': 'error' } }],
        configDirectory: cwd,
        workingDirectory: cwd,
        files: [file],
        fileContents: { [file]: 'debugger;\n' },
      });
      expect(result.diagnostics.map((d) => d.ruleName)).toEqual([
        'no-debugger',
      ]);
      await service.close();
      expect(closed).toBe(true);
      // No sleep or retry: a successful close must release the child's cwd.
      await rm(cwd, { recursive: true });
    } finally {
      await service.close();
      await rm(cwd, { recursive: true, force: true });
    }
  });
});

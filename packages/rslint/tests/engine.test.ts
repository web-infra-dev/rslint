import { afterAll, beforeAll, describe, test, expect } from 'rstack/test';
import { ChildProcess, execFile } from 'node:child_process';
import { once } from 'node:events';
import fs from 'node:fs';
import os from 'node:os';
import { PassThrough, Writable } from 'node:stream';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { promisify } from 'node:util';
import { runEngine } from '../src/cli/engine.js';
import { ConfigModuleHost } from '../src/config/config-loader.js';
import { resolveRslintBinary } from '../src/internal/resolve-binary.js';
import { createMemoryTransport } from '../src/ipc/memory-transport.js';
import { IpcClient, decodeFrame, encodeFrame } from '../src/ipc/client.js';
import { createPluginLintHost } from '../src/eslint-plugin/host.js';
import { resolvePluginAttachments } from '../src/eslint-plugin/plugin/attachments.js';
import {
  parse,
  parseSharedBytes,
} from '../src/eslint-plugin/native/load-binding.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const FAKE_BIN = path.resolve(__dirname, './fixtures/fake-ipc-binary.cjs');
const CONFIG_RACE_BIN = path.resolve(
  __dirname,
  './fixtures/fake-config-activation-race.cjs',
);
const EXIT_DURING_CONFIG_ACTIVATION_BIN = path.resolve(
  __dirname,
  './fixtures/fake-exit-during-config-activation.cjs',
);
const CONFIG_ACTIVATION_OUTER_DEADLOCK_SENTINEL_MS = 35 * 60_000;

describe('CLI shared source integration', () => {
  test.each(['snapshot', 'fix', 'inline', 'configure-failure'] as const)(
    'preserves complete source and parser output: %s',
    async (mode) => {
      const root = fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-source-'));
      const file = path.join(root, 'input.ts');
      const config = path.join(root, 'rslint.config.mjs');
      const original = '\ufeffconst oldName = "😀 café";\r\n// \u0000\n';
      fs.writeFileSync(file, original);
      fs.writeFileSync(
        config,
        `export default [{ files: ['**/*.ts'], plugins: { local: { rules: { rename: {
      meta: { fixable: 'code' },
      create(context) { return { Identifier(node) {
        if (node.name === 'oldName') context.report({ node, message: 'rename identifier', fix: fixer => fixer.replaceText(node, 'newName') });
      } }; }
    } } } }, rules: { 'local/rename': 'error' } }];`,
      );
      const transport = mode === 'inline' ? undefined : createMemoryTransport();
      const shared = mode !== 'inline' && mode !== 'configure-failure';
      let configurationCalls = 0;
      if (transport) {
        const configure = transport.configure;
        transport.configure = (config) => {
          configurationCalls++;
          if (mode === 'configure-failure')
            throw new Error('allocation failed');
          configure(config);
        };
      }
      const snapshots: string[] = [];
      const stdout = new PassThrough();
      let output = '';
      stdout.on('data', (data: Buffer) => {
        output += data.toString();
      });
      try {
        const exit = await runEngine({
          binPath: resolveRslintBinary(),
          goArgs: [...(mode === 'fix' ? ['--fix'] : []), '--no-color', file],
          cwd: root,
          stdout,
          stderr: new PassThrough(),
          runtime: { singleThreaded: true },
          extraInit: { configDiscovery: { explicitConfigPath: config } },
          createMemoryTransport: () => transport,
          createPluginLintHost: async (configs, log, singleThreaded) => {
            const host = await createPluginLintHost(
              configs,
              log,
              singleThreaded,
            );
            return {
              shutdown: () => host.shutdown(),
              lint: async (request: any, signal, attachments) => {
                const resolved = resolvePluginAttachments(request, attachments);
                for (const input of resolved.files) {
                  if (shared) {
                    expect(input.text).toBeUndefined();
                    expect(input.sharedSource).toBeDefined();
                    if (!input.sharedSource)
                      throw new Error('missing shared snapshot');
                    const native = parseSharedBytes(
                      input.path,
                      input.sharedSource,
                      'module',
                      false,
                    );
                    expect(native.parsed).toEqual(
                      parse(input.path, native.sourceText, 'module', false),
                    );
                    snapshots.push(native.sourceText);
                  } else {
                    expect(input.sharedSource).toBeUndefined();
                    if (typeof input.text !== 'string')
                      throw new Error('missing inline snapshot');
                    snapshots.push(input.text);
                  }
                }
                // A concurrent disk edit after Go took its snapshot must never
                // change what either the native parser or the real worker lints.
                if (mode === 'snapshot')
                  fs.writeFileSync(file, 'const changedOnDisk = 1;');
                return host.lint(request, signal, attachments);
              },
            };
          },
        });
        expect(exit).toBe(mode === 'fix' ? 0 : 1);
        expect(configurationCalls).toBe(mode === 'inline' ? 0 : 1);
        expect(snapshots[0]).toBe(shared ? original.slice(1) : original);
        if (mode === 'fix') {
          expect(snapshots).toEqual([
            original.slice(1),
            original.slice(1).replace('oldName', 'newName'),
          ]);
          expect(fs.readFileSync(file, 'utf8')).toBe(
            original.replace('oldName', 'newName'),
          );
        } else {
          expect(output).toContain('rename identifier');
        }
      } finally {
        transport?.close();
        fs.rmSync(root, { recursive: true, force: true });
      }
    },
  );

  test('configures storage on demand from a non-default peer layout', async () => {
    const transport = createMemoryTransport();
    const fd = transport.fd();
    if (fd !== undefined) expect(fs.fstatSync(fd).size).toBe(0);
    expect(transport.configuration()).toBeUndefined();
    transport.transfer = () => transport.descriptor();
    const configurations: unknown[] = [];
    const configure = transport.configure;
    transport.configure = (config) => {
      configurations.push(config);
      configure(config);
    };
    try {
      const exitCode = await runEngine({
        binPath: process.execPath,
        goArgs: [FAKE_BIN, 'require-mapping'],
        stdout: new PassThrough(),
        stderr: new PassThrough(),
        createMemoryTransport: () => transport,
      });
      expect(exitCode).toBe(0);
      expect(configurations).toEqual([
        {
          version: 1,
          slotCount: 3,
          slotSize: 4096,
          headerSize: 512,
          publicationStride: 32,
        },
      ]);
    } finally {
      transport.close();
    }
  });
});

/**
 * Runs the engine against the fake IPC binary, which echoes the `init`
 * payload it received back through an `output` frame — letting the tests
 * assert on what actually crossed the wire, not on engine internals.
 */
async function runWithSink(sink: PassThrough): Promise<{
  exitCode: number;
  payload: { runtime?: { stdoutIsTTY?: boolean } };
}> {
  let captured = '';
  sink.on('data', (d: Buffer) => {
    captured += d.toString();
  });
  const exitCode = await runEngine({
    binPath: process.execPath,
    goArgs: [FAKE_BIN],
    stdout: sink,
    stderr: new PassThrough(),
  });
  return { exitCode, payload: JSON.parse(captured) };
}

describe('runEngine init payload TTY fact', () => {
  test('sends runtime.stdoutIsTTY=true when the output sink is a TTY', async () => {
    const sink = Object.assign(new PassThrough(), { isTTY: true });
    const { exitCode, payload } = await runWithSink(sink);
    expect(exitCode).toBe(0);
    expect(payload.runtime?.stdoutIsTTY).toBe(true);
  });

  test('sends runtime.stdoutIsTTY=false for a non-TTY sink', async () => {
    const { exitCode, payload } = await runWithSink(new PassThrough());
    expect(exitCode).toBe(0);
    expect(payload.runtime?.stdoutIsTTY).toBe(false);
  });
});

describe('runEngine IPC disconnect cleanup', () => {
  let fixtureRoot: string;
  let fixtureBin: string;

  beforeAll(async () => {
    fixtureRoot = fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-disconnect-'));
    fixtureBin = path.join(
      fixtureRoot,
      process.platform === 'win32' ? 'peer.exe' : 'peer',
    );
    // Windows libuv fs.close deliberately leaves descriptors 0-2 open.
    // A Go peer can close the real stdout handle while remaining alive on
    // stdin, matching the process whose lifetime runEngine actually manages.
    await promisify(execFile)(
      'go',
      [
        'build',
        '-o',
        fixtureBin,
        path.join(__dirname, 'fixtures/ipc-disconnect/main.go'),
      ],
      { timeout: 60_000 },
    );
  }, 70_000);

  afterAll(() => {
    if (fixtureRoot) fs.rmSync(fixtureRoot, { recursive: true, force: true });
  });

  function start(mode: string) {
    const stderr = new PassThrough();
    const state = { stderr: '', timedOut: false };
    stderr.on('data', (chunk: Buffer) => {
      state.stderr += chunk.toString();
    });
    let child: ChildProcess | undefined;
    const originalOnce = ChildProcess.prototype.once;
    ChildProcess.prototype.once = function (event, listener) {
      // Capture the actual spawned process to control EOF/exit ordering.
      // rslint-disable-next-line @typescript-eslint/no-this-alias
      if (event === 'exit' && !child) child = this;
      return originalOnce.call(this, event, listener);
    } as typeof ChildProcess.prototype.once;
    let run: Promise<number>;
    try {
      run = runEngine({
        binPath: fixtureBin,
        goArgs: [mode],
        stdout: new PassThrough(),
        stderr,
        createMemoryTransport: () => undefined,
      });
    } finally {
      ChildProcess.prototype.once = originalOnce;
    }
    if (!child)
      throw new Error('engine did not install its child exit handler');
    const spawned = child;
    // Only a failed lifecycle reaches this watchdog. Successful tests wait
    // for actual EOF/process events and never advance via a short sleep.
    const watchdog = setTimeout(() => {
      state.timedOut = true;
      spawned.kill('SIGKILL');
    }, 15_000);
    return {
      child: spawned,
      run,
      state,
      cleanup() {
        clearTimeout(watchdog);
        spawned.kill('SIGKILL');
      },
    };
  }

  test.each(['eof-before-init', 'eof-after-init'])(
    'terminates a child that remains alive after %s and reports host failure',
    async (mode) => {
      const fixture = start(mode);
      try {
        await once(fixture.child.stdout!, 'end');
        expect(fixture.child.exitCode).toBeNull();
        expect(await fixture.run).toBe(2);
        expect(fixture.state.timedOut).toBe(false);
        expect(fixture.state.stderr).toContain(
          'Go IPC closed without process exit; terminating Go process',
        );
        expect(fixture.child.killed).toBe(true);
        expect(
          fixture.child.exitCode !== null || fixture.child.signalCode !== null,
        ).toBe(true);
      } finally {
        fixture.cleanup();
      }
    },
    20_000,
  );

  test.each(
    ['eof-before-init', 'eof-after-init'].flatMap((mode) => [
      { mode, exitCode: 0 },
      { mode, exitCode: 23 },
    ]),
  )(
    'preserves the natural child exit code after EOF: %j',
    async ({ mode, exitCode }) => {
      const fixture = start(mode);
      try {
        await once(fixture.child.stdout!, 'end');
        expect(fixture.child.exitCode).toBeNull();
        expect(fixture.child.killed).toBe(false);
        // This fixture control travels directly over the still-open pipe only
        // after Node has observed EOF. The closed IpcClient sends no new work.
        fixture.child.stdin!.write(
          encodeFrame({
            kind: 'exit-after-eof',
            id: 0,
            data: { code: exitCode },
          }),
        );
        expect(await fixture.run).toBe(exitCode);
        expect(fixture.state.timedOut).toBe(false);
        expect(fixture.child.killed).toBe(false);
        expect(fixture.state.stderr).not.toContain('terminating Go process');
      } finally {
        fixture.cleanup();
      }
    },
    20_000,
  );

  test('flushes the complete init response before closing stdout', async () => {
    const fixture = start('eof-after-init');
    const chunks: Buffer[] = [];
    fixture.child.stdout!.on('data', (chunk: Buffer) => chunks.push(chunk));
    try {
      await once(fixture.child.stdout!, 'end');
      const output = Buffer.concat(chunks);
      const frame = decodeFrame(output);
      expect(frame?.msg).toMatchObject({
        kind: 'response',
        data: { ok: true },
      });
      expect(frame?.consumed).toBe(output.length);
      expect(fixture.child.exitCode).toBeNull();
      expect(fixture.child.killed).toBe(false);
      fixture.child.stdin!.write(
        encodeFrame({ kind: 'exit-after-eof', id: 0, data: { code: 23 } }),
      );
      expect(await fixture.run).toBe(23);
      expect(fixture.state.timedOut).toBe(false);
      expect(fixture.state.stderr).not.toContain('terminating Go process');
    } finally {
      fixture.cleanup();
    }
  }, 20_000);

  test('terminates an init-rejecting child without treating it as EOF', async () => {
    const fixture = start('reject-init');
    try {
      expect(await fixture.run).toBe(2);
      expect(fixture.state.timedOut).toBe(false);
      expect(fixture.state.stderr).toContain(
        'init failed: injected init failure',
      );
      expect(fixture.state.stderr).not.toContain(
        'IPC closed without process exit',
      );
      expect(fixture.child.killed).toBe(true);
    } finally {
      fixture.cleanup();
    }
  }, 20_000);

  test('keeps an init error as failure when EOF precedes its catch continuation', async () => {
    const originalSend = IpcClient.prototype.sendRequest;
    let businessError: unknown;
    let closeReason: unknown;
    IpcClient.prototype.sendRequest = function (kind, data, attachments) {
      const response = originalSend.call(this, kind, data, attachments);
      if (kind !== 'init') return response;
      // Hold the already-rejected business request until real EOF, making
      // this microtask ordering deterministic across process schedulers.
      return response.catch(async (error: unknown) => {
        businessError = error;
        closeReason = await this.done;
        throw error;
      });
    } as typeof IpcClient.prototype.sendRequest;
    let fixture: ReturnType<typeof start>;
    try {
      fixture = start('reject-init-eof');
    } finally {
      IpcClient.prototype.sendRequest = originalSend;
    }
    try {
      await once(fixture.child.stdout!, 'end');
      fixture.child.stdin!.write(
        encodeFrame({ kind: 'exit-after-eof', id: 0, data: { code: 0 } }),
      );
      expect(await fixture.run).toBe(2);
      expect(businessError).toMatchObject({ message: 'injected init failure' });
      expect(businessError).not.toBe(closeReason);
      expect(fixture.state.timedOut).toBe(false);
      expect(fixture.state.stderr).toContain(
        'init failed: injected init failure',
      );
    } finally {
      fixture.cleanup();
    }
  }, 20_000);
});

describe('runEngine output write barriers', () => {
  test('waits for the destination callback before acknowledging output', async () => {
    const events: string[] = [];
    let releaseFirstWrite!: () => void;
    let markFirstWriteStarted!: () => void;
    const firstWriteStarted = new Promise<void>((resolve) => {
      markFirstWriteStarted = resolve;
    });
    const firstWriteGate = new Promise<void>((resolve) => {
      releaseFirstWrite = resolve;
    });
    let stdoutWrites = 0;
    const stdout = new Writable({
      highWaterMark: 1,
      write(_chunk, _encoding, callback) {
        stdoutWrites++;
        const write = stdoutWrites;
        events.push(`stdout:${write}:start`);
        if (write === 1) {
          markFirstWriteStarted();
          void firstWriteGate.then(() => {
            events.push(`stdout:${write}:done`);
            callback();
          });
          return;
        }
        events.push(`stdout:${write}:done`);
        callback();
      },
    });
    let settled = false;
    const run = runEngine({
      binPath: process.execPath,
      goArgs: [FAKE_BIN],
      stdout,
      stderr: new PassThrough(),
    }).then((code) => {
      settled = true;
      return code;
    });

    try {
      await firstWriteStarted;
      await new Promise<void>((resolve) => setImmediate(resolve));
      expect(settled).toBe(false);
      expect(events).toEqual(['stdout:1:start']);
    } finally {
      releaseFirstWrite();
    }

    expect(await run).toBe(0);
    expect(events).toEqual([
      'stdout:1:start',
      'stdout:1:done',
      'stdout:2:start',
      'stdout:2:done',
    ]);
  });

  test('waits for later stdout writes before acknowledging shutdown', async () => {
    let releaseSecondWrite!: () => void;
    let markSecondWriteStarted!: () => void;
    const secondWriteStarted = new Promise<void>((resolve) => {
      markSecondWriteStarted = resolve;
    });
    const secondWriteGate = new Promise<void>((resolve) => {
      releaseSecondWrite = resolve;
    });
    let stdoutWrites = 0;
    const stdout = new Writable({
      write(_chunk, _encoding, callback) {
        stdoutWrites++;
        if (stdoutWrites === 2) {
          markSecondWriteStarted();
          void secondWriteGate.then(() => callback());
          return;
        }
        callback();
      },
    });
    let settled = false;
    const run = runEngine({
      binPath: process.execPath,
      goArgs: [FAKE_BIN],
      stdout,
      stderr: new PassThrough(),
    }).then((code) => {
      settled = true;
      return code;
    });

    try {
      await secondWriteStarted;
      await new Promise<void>((resolve) => setImmediate(resolve));
      expect(settled).toBe(false);
    } finally {
      releaseSecondWrite();
    }

    expect(await run).toBe(0);
    expect(stdoutWrites).toBe(2);
  });

  test('returns failure when acknowledged stdout forwarding fails', async () => {
    const stdout = new Writable({
      write(_chunk, _encoding, callback) {
        callback(new Error('injected stdout failure'));
      },
    });
    const exitCode = await runEngine({
      binPath: process.execPath,
      goArgs: [FAKE_BIN],
      stdout,
      stderr: new PassThrough(),
    });

    expect(exitCode).toBe(2);
  });

  test('returns failure without hanging when stdout closes during acknowledged write', async () => {
    const stdout = new Writable({
      write() {
        stdout.destroy();
      },
    });
    const exitCode = await runEngine({
      binPath: process.execPath,
      goArgs: [FAKE_BIN],
      stdout,
      stderr: new PassThrough(),
    });

    expect(exitCode).toBe(2);
  });

  test('rejects shutdown when a later stdout notification fails', async () => {
    let stdoutWrites = 0;
    const stdout = new Writable({
      write(_chunk, _encoding, callback) {
        stdoutWrites++;
        callback(
          stdoutWrites === 2
            ? new Error('injected notification failure')
            : undefined,
        );
      },
    });
    const exitCode = await runEngine({
      binPath: process.execPath,
      goArgs: [FAKE_BIN],
      stdout,
      stderr: new PassThrough(),
    });

    expect(exitCode).toBe(2);
    expect(stdoutWrites).toBe(2);
  });

  test('rejects shutdown without hanging when stdout closes during a later notification', async () => {
    let stdoutWrites = 0;
    const stdout = new Writable({
      write(_chunk, _encoding, callback) {
        stdoutWrites++;
        if (stdoutWrites === 2) {
          stdout.destroy();
          return;
        }
        callback();
      },
    });
    const exitCode = await runEngine({
      binPath: process.execPath,
      goArgs: [FAKE_BIN],
      stdout,
      stderr: new PassThrough(),
    });

    expect(exitCode).toBe(2);
    expect(stdoutWrites).toBe(2);
  });
});

describe('runEngine config activation', () => {
  test('disposes and never publishes a host whose prepare changes its config', async () => {
    const root = fs.mkdtempSync(
      path.join(os.tmpdir(), 'rslint-cli-config-activation-'),
    );
    const configPath = path.join(root, 'rslint.config.mjs');
    fs.writeFileSync(
      configPath,
      'export default [{ plugins: { local: { rules: { example: {} } } } }];\n',
    );
    const stdout = new PassThrough();
    let captured = '';
    let lintCalls = 0;
    let shutdownCalls = 0;
    stdout.on('data', (chunk: Buffer) => {
      captured += chunk.toString();
    });

    try {
      const exitCode = await runEngine({
        binPath: process.execPath,
        goArgs: [CONFIG_RACE_BIN, configPath],
        stdout,
        stderr: new PassThrough(),
        createPluginLintHost: async () => {
          fs.writeFileSync(configPath, '// changed by mocked worker prepare\n');
          return {
            async lint() {
              lintCalls++;
              return { results: ['stale-host-was-visible'] };
            },
            async shutdown() {
              shutdownCalls++;
            },
          };
        },
      });

      expect(exitCode).toBe(0);
      expect(captured).toContain('plugin host was being prepared');
      expect(captured.match(/plugin host was being prepared/g)).toHaveLength(2);
      expect(lintCalls).toBe(0);
      expect(shutdownCalls).toBe(1);
    } finally {
      fs.rmSync(root, { recursive: true, force: true });
    }
  });

  test(
    'disposes a plugin host that finishes after the Go child exits',
    async () => {
      const root = fs.mkdtempSync(
        path.join(os.tmpdir(), 'rslint-cli-config-exit-'),
      );
      const configPath = path.join(root, 'rslint.config.mjs');
      fs.writeFileSync(
        configPath,
        'export default [{ plugins: { local: { rules: { example: {} } } } }];\n',
      );
      const buildMarker = path.join(root, 'plugin-host-build-started');
      let buildStarted = false;
      let shutdownCalls = 0;
      let markShutdownStarted!: () => void;
      const shutdownStarted = new Promise<void>((resolve) => {
        markShutdownStarted = resolve;
      });
      let releaseShutdown!: () => void;
      const shutdownGate = new Promise<void>((resolve) => {
        releaseShutdown = resolve;
      });
      let markChildExited!: () => void;
      const childExited = new Promise<void>((resolve) => {
        markChildExited = resolve;
      });
      const originalOnce = ChildProcess.prototype.once;
      let exitListenerWrapped = false;
      ChildProcess.prototype.once = function (event, listener) {
        if (event !== 'exit' || exitListenerWrapped) {
          return originalOnce.call(this, event, listener);
        }
        exitListenerWrapped = true;
        return originalOnce.call(this, event, function (...args: unknown[]) {
          Reflect.apply(listener, this, args);
          markChildExited();
        });
      } as typeof ChildProcess.prototype.once;

      try {
        let run: Promise<number>;
        try {
          run = runEngine({
            binPath: process.execPath,
            goArgs: [
              EXIT_DURING_CONFIG_ACTIVATION_BIN,
              configPath,
              buildMarker,
            ],
            stdout: new PassThrough(),
            stderr: new PassThrough(),
            createPluginLintHost: async () => {
              buildStarted = true;
              fs.writeFileSync(buildMarker, 'started');
              await childExited;
              return {
                async lint() {
                  return { results: [] };
                },
                async shutdown() {
                  shutdownCalls++;
                  markShutdownStarted();
                  await shutdownGate;
                },
              };
            },
          });
        } finally {
          ChildProcess.prototype.once = originalOnce;
        }
        let runSettled = false;
        const runBoundary = run.then(
          () => {
            runSettled = true;
            return 'run-settled' as const;
          },
          () => {
            runSettled = true;
            return 'run-settled' as const;
          },
        );
        const firstBoundary = await Promise.race([
          shutdownStarted.then(() => 'shutdown-started' as const),
          runBoundary,
        ]);
        expect(firstBoundary).toBe('shutdown-started');
        // Holding the shutdown gate across a full event-loop turn distinguishes
        // an awaited teardown from a fire-and-forget call that merely starts it.
        await new Promise<void>((resolve) => setImmediate(resolve));
        expect(runSettled).toBe(false);
        releaseShutdown();
        const exitCode = await run;

        expect(exitCode).toBe(0);
        expect(exitListenerWrapped).toBe(true);
        expect(buildStarted).toBe(true);
        expect(fs.readFileSync(buildMarker, 'utf8')).toBe('started');
        expect(shutdownCalls).toBe(1);
      } finally {
        releaseShutdown();
        ChildProcess.prototype.once = originalOnce;
        fs.rmSync(root, { recursive: true, force: true });
      }
    },
    CONFIG_ACTIVATION_OUTER_DEADLOCK_SENTINEL_MS,
  );

  test(
    'disposes a staged host before returning when Go exits during post-prepare verification',
    async () => {
      const root = fs.mkdtempSync(
        path.join(os.tmpdir(), 'rslint-cli-config-staged-exit-'),
      );
      const configPath = path.join(root, 'rslint.config.mjs');
      const postPrepareMarker = path.join(root, 'post-prepare-started');
      fs.writeFileSync(configPath, '// stable config bytes\n');
      let fingerprintReads = 0;
      let releasePostPrepare!: () => void;
      let markPostPrepareStarted!: () => void;
      let markPostPrepareFinished!: () => void;
      let postPrepareEntered = false;
      const postPrepareStarted = new Promise<void>((resolve) => {
        markPostPrepareStarted = resolve;
      });
      const postPrepareFinished = new Promise<void>((resolve) => {
        markPostPrepareFinished = resolve;
      });
      const postPrepareGate = new Promise<void>((resolve) => {
        releasePostPrepare = resolve;
      });
      const configModuleHost = new ConfigModuleHost({
        loadCached: async () => [
          { plugins: { local: { rules: { example: {} } } } },
        ],
        readSource: async (sourcePath) => {
          fingerprintReads++;
          if (fingerprintReads === 4) {
            postPrepareEntered = true;
            fs.writeFileSync(postPrepareMarker, 'started');
            markPostPrepareStarted();
            try {
              await postPrepareGate;
              return await fs.promises.readFile(sourcePath);
            } finally {
              markPostPrepareFinished();
            }
          }
          return fs.promises.readFile(sourcePath);
        },
      });
      let shutdownCalls = 0;
      let markShutdownStarted!: () => void;
      const shutdownStarted = new Promise<void>((resolve) => {
        markShutdownStarted = resolve;
      });
      let releaseShutdown!: () => void;
      const shutdownGate = new Promise<void>((resolve) => {
        releaseShutdown = resolve;
      });

      try {
        const run = runEngine({
          binPath: process.execPath,
          goArgs: [
            EXIT_DURING_CONFIG_ACTIVATION_BIN,
            configPath,
            postPrepareMarker,
          ],
          stdout: new PassThrough(),
          stderr: new PassThrough(),
          configModuleHost,
          createPluginLintHost: async () => ({
            async lint() {
              return { results: [] };
            },
            async shutdown() {
              shutdownCalls++;
              markShutdownStarted();
              await shutdownGate;
            },
          }),
        });

        let runSettled = false;
        const runBoundary = run.then(
          () => {
            runSettled = true;
            return 'run-settled' as const;
          },
          () => {
            runSettled = true;
            return 'run-settled' as const;
          },
        );
        const activationBoundary = await Promise.race([
          postPrepareStarted.then(() => 'post-prepare-started' as const),
          runBoundary,
        ]);
        expect(activationBoundary).toBe('post-prepare-started');
        const firstBoundary = await Promise.race([
          shutdownStarted.then(() => 'shutdown-started' as const),
          runBoundary,
        ]);
        expect(firstBoundary).toBe('shutdown-started');
        await new Promise<void>((resolve) => setImmediate(resolve));
        expect(runSettled).toBe(false);
        releaseShutdown();
        const exitCode = await run;
        expect(exitCode).toBe(0);
        expect(shutdownCalls).toBe(1);
      } finally {
        releaseShutdown();
        releasePostPrepare();
        if (postPrepareEntered) await postPrepareFinished;
        fs.rmSync(root, { recursive: true, force: true });
      }
    },
    CONFIG_ACTIVATION_OUTER_DEADLOCK_SENTINEL_MS,
  );
});

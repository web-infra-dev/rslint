import { describe, expect, test } from 'rstack/test';
import { ChildProcess } from 'node:child_process';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { PassThrough, Writable } from 'node:stream';
import { fileURLToPath } from 'node:url';
import { runEngine, type EngineRunOptions } from '../src/cli/engine.js';
import { ConfigModuleHost } from '../src/config/config-loader.js';

const FAKE_BIN = fileURLToPath(
  new URL('./fixtures/fake-worker-warmup.cjs', import.meta.url),
);
const PLUGIN_CONFIG = [{ plugins: { local: { rules: { example: {} } } } }];

function gate<T = void>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((done) => {
    resolve = done;
  });
  return { promise, resolve };
}

interface WireResult {
  kind: string;
  data: Record<string, unknown>;
}

interface FixtureEvent {
  event: string;
  preparation?: WireResult;
  duplicates?: WireResult[];
  conflicts?: WireResult[];
  loadAfterActivation?: WireResult;
  settled?: boolean | boolean[];
  results?: WireResult[];
  activation?: WireResult;
  plugin?: WireResult;
}

function fixture(singleThreaded: boolean) {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-worker-warmup-'));
  const configPath = path.join(root, 'rslint.config.mjs');
  fs.writeFileSync(configPath, '// original config bytes\n');
  const events = new Map<string, ReturnType<typeof gate<FixtureEvent>>>();
  const eventGate = (name: string) => {
    let deferred = events.get(name);
    if (!deferred) {
      deferred = gate<FixtureEvent>();
      events.set(name, deferred);
    }
    return deferred;
  };
  let beforeAcknowledge = (_event: FixtureEvent): Promise<void> =>
    Promise.resolve();
  const stdout = new Writable({
    write(chunk: Buffer, _encoding, callback) {
      const event: FixtureEvent = JSON.parse(chunk.toString());
      eventGate(event.event).resolve(event);
      void beforeAcknowledge(event).then(
        () => callback(),
        (error: Error) => callback(error),
      );
    },
  });
  return {
    configPath,
    setAcknowledge(callback: typeof beforeAcknowledge) {
      beforeAcknowledge = callback;
    },
    async event(name: string): Promise<FixtureEvent> {
      // This is only a deadlock cleanup sentinel. Success and ordering are
      // established by the peer's acknowledged IPC gates, never elapsed time.
      let timer: ReturnType<typeof setTimeout> | undefined;
      try {
        return await Promise.race([
          eventGate(name).promise,
          new Promise<never>((_, reject) => {
            timer = setTimeout(
              () => reject(new Error(`warmup peer did not report ${name}`)),
              30_000,
            );
          }),
        ]);
      } finally {
        clearTimeout(timer);
      }
    },
    run(mode: string, options: Partial<EngineRunOptions> = {}) {
      return runEngine({
        binPath: process.execPath,
        goArgs: [FAKE_BIN, configPath, mode],
        stdout,
        stderr: new PassThrough(),
        runtime: { singleThreaded },
        configModuleHost: new ConfigModuleHost({
          loadCached: async () => PLUGIN_CONFIG,
        }),
        ...options,
      });
    },
    dispose() {
      fs.rmSync(root, { recursive: true, force: true });
    },
  };
}

describe.each([false, true])(
  'CLI worker warmup (singleThreaded=%s)',
  (singleThreaded) => {
    test('returns planning metadata once and joins concurrent activation and plugin requests', async () => {
      const peer = fixture(singleThreaded);
      const build = gate();
      const buildStarted = gate();
      let createCalls = 0;
      let shutdownCalls = 0;
      const lintRequests: unknown[] = [];
      peer.setAcknowledge(async ({ event }) => {
        if (event === 'waiters-started') await buildStarted.promise;
      });
      const run = peer.run('concurrent', {
        createPluginLintHost: async (
          _configs,
          _onLog,
          workerSingleThreaded,
        ) => {
          expect(workerSingleThreaded).toBe(singleThreaded);
          createCalls++;
          buildStarted.resolve();
          await build.promise;
          return {
            async lint(request) {
              lintRequests.push(request);
              return { results: [] };
            },
            async shutdown() {
              shutdownCalls++;
            },
          };
        },
      });
      try {
        const prepared = await peer.event('prepared');
        expect(prepared.preparation).toMatchObject({
          kind: 'response',
          data: {
            transactionId: 'cli-worker-warmup',
            eslintPluginEntries: [{ prefix: 'local', ruleNames: ['example'] }],
          },
        });
        const repeated = await peer.event('repeated');
        expect(repeated.duplicates).toHaveLength(2);
        for (const duplicate of repeated.duplicates ?? []) {
          expect(duplicate.data).toEqual(prepared.preparation?.data);
          expect(duplicate.kind).toBe('response');
        }
        expect(repeated.conflicts).toHaveLength(2);
        for (const conflict of repeated.conflicts ?? []) {
          expect(conflict).toMatchObject({
            kind: 'error',
            data: { message: 'engine: conflicting config activation' },
          });
        }
        expect(repeated.loadAfterActivation).toMatchObject({ kind: 'error' });
        expect((await peer.event('waiters-pending')).settled).toEqual([
          false,
          false,
          false,
        ]);
        expect(createCalls).toBe(1);
        expect(lintRequests).toEqual([]);
        build.resolve();
        const completed = await peer.event('completed');
        expect(completed.results?.map(({ kind }) => kind)).toEqual([
          'response',
          'response',
          'response',
        ]);
        expect(completed.results?.[0].data).toEqual(prepared.preparation?.data);
        expect(lintRequests).toEqual([
          { request: 'first' },
          { request: 'second' },
        ]);
        expect(await run).toBe(0);
        expect(shutdownCalls).toBe(1);
      } finally {
        build.resolve();
        await run;
        peer.dispose();
      }
    });

    test.each([
      ['before preparation', 'changed while it was being loaded', 0],
      ['during preparation', 'plugin host was being prepared', 1],
      ['factory failure', 'mocked worker import failed', 0],
      ['without tasks', 'mocked worker import failed', 0],
    ] as const)(
      'propagates failure %s to the activation barrier',
      async (phase, message, expectedShutdowns) => {
        const peer = fixture(singleThreaded);
        const build = gate();
        let reads = 0;
        let createCalls = 0;
        let lintCalls = 0;
        let shutdownCalls = 0;
        const run = peer.run(
          phase === 'without tasks' ? 'failure-without-tasks' : 'failure',
          {
            configModuleHost: new ConfigModuleHost({
              loadCached: async () => PLUGIN_CONFIG,
              readSource: async (sourcePath) => {
                reads++;
                if (phase === 'before preparation' && reads === 3) {
                  fs.writeFileSync(
                    sourcePath,
                    '// changed before preparation\n',
                  );
                }
                return fs.promises.readFile(sourcePath);
              },
            }),
            createPluginLintHost: async () => {
              createCalls++;
              await build.promise;
              if (phase !== 'during preparation') {
                throw new Error('mocked worker import failed');
              }
              fs.writeFileSync(
                peer.configPath,
                '// changed during preparation\n',
              );
              return {
                async lint() {
                  lintCalls++;
                  return { results: [] };
                },
                async shutdown() {
                  shutdownCalls++;
                },
              };
            },
          },
        );
        try {
          const prepared = await peer.event('prepared');
          expect(prepared.preparation?.kind).toBe(
            phase === 'before preparation' ? 'error' : 'response',
          );
          build.resolve();
          const completed = await peer.event('completed');
          expect(completed.activation?.kind).toBe('error');
          expect(completed.activation?.data.message).toContain(message);
          if (phase !== 'without tasks') {
            expect(completed.plugin?.kind).toBe('error');
            expect(completed.plugin?.data.message).toContain(message);
          } else {
            expect(completed.plugin).toBeUndefined();
          }
          expect(await run).toBe(0);
          expect(createCalls).toBe(phase === 'before preparation' ? 0 : 1);
          expect(lintCalls).toBe(0);
          expect(shutdownCalls).toBe(expectedShutdowns);
        } finally {
          build.resolve();
          await run;
          peer.dispose();
        }
      },
    );

    test('keeps native-only preparation synchronous', async () => {
      const peer = fixture(singleThreaded);
      const preparation = gate();
      const preparationStarted = gate();
      let reads = 0;
      let createCalls = 0;
      let shutdownCalls = 0;
      peer.setAcknowledge(async ({ event }) => {
        if (event === 'waiters-started') await preparationStarted.promise;
      });
      const run = peer.run('blocked-prepare', {
        configModuleHost: new ConfigModuleHost({
          loadCached: async () => [],
          readSource: async (sourcePath) => {
            reads++;
            if (reads === 4) {
              preparationStarted.resolve();
              await preparation.promise;
            }
            return fs.promises.readFile(sourcePath);
          },
        }),
        createPluginLintHost: async () => {
          createCalls++;
          return {
            async lint() {
              return { results: [] };
            },
            async shutdown() {
              shutdownCalls++;
            },
          };
        },
      });
      try {
        expect((await peer.event('waiters-pending')).settled).toBe(false);
        preparation.resolve();
        expect((await peer.event('completed')).preparation?.kind).toBe(
          'response',
        );
        expect(await run).toBe(0);
        expect(createCalls).toBe(0);
        expect(shutdownCalls).toBe(createCalls);
      } finally {
        preparation.resolve();
        preparationStarted.resolve();
        await run;
        peer.dispose();
      }
    });

    test('joins shutdown for a warmed host that finishes after child exit', async () => {
      const peer = fixture(singleThreaded);
      const childExited = gate();
      const shutdownStarted = gate();
      const shutdown = gate();
      let shutdownCalls = 0;
      let lintCalls = 0;
      const originalOnce = ChildProcess.prototype.once;
      let wrappedExit = false;
      ChildProcess.prototype.once = function (event, listener) {
        if (event !== 'exit' || wrappedExit) {
          return originalOnce.call(this, event, listener);
        }
        wrappedExit = true;
        return originalOnce.call(this, event, function (...args: unknown[]) {
          Reflect.apply(listener, this, args);
          childExited.resolve();
        });
      } as typeof ChildProcess.prototype.once;
      let run: Promise<number>;
      try {
        run = peer.run('exit', {
          createPluginLintHost: async () => {
            await childExited.promise;
            return {
              async lint() {
                lintCalls++;
                return { results: [] };
              },
              async shutdown() {
                shutdownCalls++;
                shutdownStarted.resolve();
                await shutdown.promise;
              },
            };
          },
        });
      } finally {
        ChildProcess.prototype.once = originalOnce;
      }
      let settled = false;
      void run.then(() => {
        settled = true;
      });
      try {
        expect((await peer.event('prepared')).preparation?.kind).toBe(
          'response',
        );
        await shutdownStarted.promise;
        await new Promise<void>((resolve) => setImmediate(resolve));
        expect(settled).toBe(false);
        expect(wrappedExit).toBe(true);
        expect(shutdownCalls).toBe(1);
        expect(lintCalls).toBe(0);
        shutdown.resolve();
        expect(await run).toBe(0);
        expect(shutdownCalls).toBe(1);
      } finally {
        childExited.resolve();
        shutdown.resolve();
        await run;
        peer.dispose();
      }
    });
  },
);

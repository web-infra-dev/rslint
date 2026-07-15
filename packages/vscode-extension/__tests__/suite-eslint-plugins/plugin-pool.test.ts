import * as assert from 'node:assert';

import type {
  ConfigDescriptor,
  EslintPluginLintRequest,
  EslintPluginLintResult,
  PluginLintHost,
} from '@rslint/core/eslint-plugin';
import { CancellationTokenSource } from 'vscode';
import { PluginLintPool } from '../../src/PluginLintPool';
import type { Logger } from '../../src/logger';

function deferred<T>(): {
  promise: Promise<T>;
  resolve: (value: T) => void;
} {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((done) => {
    resolve = done;
  });
  return { promise, resolve };
}

class TestHost implements PluginLintHost {
  readonly started = deferred<void>();
  lintCalls = 0;
  shutdownCalls = 0;
  lintResult: Promise<EslintPluginLintResult> = Promise.resolve({
    results: [],
  });

  async lint(): Promise<EslintPluginLintResult> {
    this.lintCalls++;
    this.started.resolve();
    return this.lintResult;
  }

  async shutdown(): Promise<void> {
    this.shutdownCalls++;
  }
}

function descriptor(name: string): ConfigDescriptor[] {
  return [{ configPath: `/${name}.mjs`, configDirectory: `file:///${name}` }];
}

function request(generation: string): EslintPluginLintRequest {
  return {
    generation,
    files: [],
    rules: {},
    fix: false,
    suggestionsMode: 'off',
  };
}

function testLogger(): Logger {
  return {
    error() {},
    debug() {},
  } as unknown as Logger;
}

suite('PluginLintPool generations', () => {
  test('an installed generation does not wait for a slow prepare', async () => {
    const hostA = new TestHost();
    const hostB = new TestHost();
    const hostBReady = deferred<PluginLintHost>();
    const hostBCreateStarted = deferred<void>();
    const pool = new PluginLintPool(testLogger(), async (configs) => {
      if (configs[0].configPath === '/a.mjs') return hostA;
      hostBCreateStarted.resolve();
      return hostBReady.promise;
    });

    assert.strictEqual(
      await pool.prepare(descriptor('a'), 'fingerprint-a', 'a'),
      true,
    );
    assert.strictEqual(await pool.commit('a'), true);

    const prepareB = pool.prepare(descriptor('b'), 'fingerprint-b', 'b');
    await hostBCreateStarted.promise;
    const lintA = pool.lint(request('a'));
    await Promise.resolve();
    const callsBeforeBWasReady = hostA.lintCalls;

    hostBReady.resolve(hostB);
    await prepareB;
    await lintA;
    await pool.abort('b');
    await pool.dispose();

    assert.strictEqual(
      callsBeforeBWasReady,
      1,
      'an unrelated host build must not block an installed generation',
    );
  });

  test('a generation still being installed waits and is rechecked', async () => {
    const host = new TestHost();
    const hostReady = deferred<PluginLintHost>();
    const hostCreateStarted = deferred<void>();
    const pool = new PluginLintPool(testLogger(), async () => {
      hostCreateStarted.resolve();
      return hostReady.promise;
    });

    const prepare = pool.prepare(descriptor('a'), 'fingerprint-a', 'a');
    await hostCreateStarted.promise;
    const lint = pool.lint(request('a'));
    await Promise.resolve();
    assert.strictEqual(host.lintCalls, 0);

    hostReady.resolve(host);
    assert.strictEqual(await prepare, true);
    await lint;
    assert.strictEqual(host.lintCalls, 1);

    await pool.abort('a');
    await pool.dispose();
  });

  test('cancellation interrupts the wait for an uninstalled generation', async () => {
    const host = new TestHost();
    const hostReady = deferred<PluginLintHost>();
    const hostCreateStarted = deferred<void>();
    const pool = new PluginLintPool(testLogger(), async () => {
      hostCreateStarted.resolve();
      return hostReady.promise;
    });
    const cancellation = new CancellationTokenSource();

    const prepare = pool.prepare(descriptor('a'), 'fingerprint-a', 'a');
    await hostCreateStarted.promise;
    let lintSettled = false;
    const lint = pool.lint(request('a'), cancellation.token).then((result) => {
      lintSettled = true;
      return result;
    });
    cancellation.cancel();
    await new Promise<void>((resolve) => setImmediate(resolve));
    const settledBeforeHostWasReady = lintSettled;

    hostReady.resolve(host);
    await prepare;
    const result = await lint;
    await pool.abort('a');
    await pool.dispose();
    cancellation.dispose();

    assert.strictEqual(settledBeforeHostWasReady, true);
    assert.deepStrictEqual(result, { results: [] });
    assert.strictEqual(host.lintCalls, 0);
  });

  test('staged generations route safely and old hosts drain before shutdown', async () => {
    const hosts = new Map<string, TestHost>();
    const pool = new PluginLintPool(
      testLogger(),
      async (configs) => {
        const host = new TestHost();
        hosts.set(configs[0].configPath, host);
        return host;
      },
      0,
    );

    assert.strictEqual(
      await pool.prepare(descriptor('a'), 'fingerprint-a', 'a'),
      true,
    );
    assert.strictEqual(await pool.commit('a'), true);
    const hostA = hosts.get('/a.mjs')!;
    const lintAResult = deferred<EslintPluginLintResult>();
    hostA.lintResult = lintAResult.promise;
    const lintA = pool.lint(request('a'));
    await hostA.started.promise;

    assert.strictEqual(
      await pool.prepare(descriptor('b'), 'fingerprint-b', 'b'),
      true,
    );
    const hostB = hosts.get('/b.mjs')!;
    await pool.lint(request('b'));
    assert.strictEqual(
      hostB.lintCalls,
      1,
      'Go can route an accepted generation before Node observes the ack',
    );

    assert.strictEqual(await pool.commit('b'), true);
    assert.strictEqual(
      hostA.shutdownCalls,
      0,
      'the old host still has an active lint lease',
    );
    await pool.lint(request('b'));
    assert.strictEqual(hostB.lintCalls, 2);

    lintAResult.resolve({ results: [] });
    await lintA;
    await new Promise((resolve) => setTimeout(resolve, 0));
    assert.strictEqual(
      hostA.shutdownCalls,
      0,
      'the predecessor remains rollback-capable until a later commit proves Go accepted b',
    );

    assert.strictEqual(
      await pool.prepare(descriptor('c'), 'fingerprint-c', 'c'),
      true,
    );
    const hostC = hosts.get('/c.mjs')!;
    assert.strictEqual(await pool.commit('c'), true);
    await new Promise((resolve) => setTimeout(resolve, 0));
    assert.strictEqual(hostA.shutdownCalls, 1);

    await pool.dispose();
    assert.strictEqual(hostB.shutdownCalls, 1);
    assert.strictEqual(hostC.shutdownCalls, 1);
  });

  test('abort rolls back an active commit whose response was lost', async () => {
    const hosts = new Map<string, TestHost>();
    const pool = new PluginLintPool(testLogger(), async (configs) => {
      const host = new TestHost();
      hosts.set(configs[0].configPath, host);
      return host;
    });

    await pool.prepare(descriptor('a'), 'fingerprint-a', 'a');
    await pool.commit('a');
    await pool.prepare(descriptor('b'), 'fingerprint-b', 'b');
    await pool.commit('b');

    // Node already switched to b, but Go did not receive/accept the commit
    // response and compensates with abort. The pool must return to a rather
    // than leaving both processes on different generations.
    await pool.abort('b');
    await pool.lint(request('a'));
    assert.strictEqual(hosts.get('/a.mjs')!.lintCalls, 1);
    await assert.rejects(pool.lint(request('b')), /unknown.*generation/);
    assert.strictEqual(hosts.get('/b.mjs')!.shutdownCalls, 1);

    await pool.dispose();
    assert.strictEqual(hosts.get('/a.mjs')!.shutdownCalls, 1);
  });

  test('default grace retains at most two old WorkerPools', async () => {
    const hosts = new Map<string, TestHost>();
    const pool = new PluginLintPool(testLogger(), async (configs) => {
      const host = new TestHost();
      hosts.set(configs[0].configPath, host);
      return host;
    });

    for (const name of ['a', 'b', 'c', 'd']) {
      assert.strictEqual(
        await pool.prepare(descriptor(name), `fingerprint-${name}`, name),
        true,
      );
      assert.strictEqual(await pool.commit(name), true);
    }

    const shutdownsDuringGrace = ['a', 'b', 'c', 'd'].map(
      (name) => hosts.get(`/${name}.mjs`)!.shutdownCalls,
    );
    await pool.dispose();

    assert.deepStrictEqual(
      shutdownsDuringGrace,
      [1, 0, 0, 0],
      'the oldest pool is retired without waiting for the default 30s delay',
    );
  });

  test('the grace cap drains an acquired lease before shutdown', async () => {
    const hosts = new Map<string, TestHost>();
    const pool = new PluginLintPool(testLogger(), async (configs) => {
      const host = new TestHost();
      hosts.set(configs[0].configPath, host);
      return host;
    });

    assert.strictEqual(
      await pool.prepare(descriptor('a'), 'fingerprint-a', 'a'),
      true,
    );
    assert.strictEqual(await pool.commit('a'), true);
    const hostA = hosts.get('/a.mjs')!;
    const lintAResult = deferred<EslintPluginLintResult>();
    hostA.lintResult = lintAResult.promise;
    const lintA = pool.lint(request('a'));
    await hostA.started.promise;

    for (const name of ['b', 'c', 'd']) {
      assert.strictEqual(
        await pool.prepare(descriptor(name), `fingerprint-${name}`, name),
        true,
      );
      assert.strictEqual(await pool.commit(name), true);
    }
    assert.strictEqual(
      hostA.shutdownCalls,
      0,
      'capacity eviction must not shut down an acquired lease',
    );

    lintAResult.resolve({ results: [] });
    await lintA;
    assert.strictEqual(hostA.shutdownCalls, 1);

    await pool.dispose();
  });

  test('dispose shuts down a retired host that still has an active lease', async () => {
    const hosts = new Map<string, TestHost>();
    const pool = new PluginLintPool(
      testLogger(),
      async (configs) => {
        const host = new TestHost();
        hosts.set(configs[0].configPath, host);
        return host;
      },
      0,
    );

    await pool.prepare(descriptor('a'), 'fingerprint-a', 'a');
    await pool.commit('a');
    const hostA = hosts.get('/a.mjs')!;
    const lintResult = deferred<EslintPluginLintResult>();
    hostA.lintResult = lintResult.promise;
    const lint = pool.lint(request('a'));
    await hostA.started.promise;

    await pool.prepare(descriptor('b'), 'fingerprint-b', 'b');
    await pool.commit('b');
    await new Promise((resolve) => setTimeout(resolve, 0));
    assert.strictEqual(hostA.shutdownCalls, 0);

    await pool.dispose();
    assert.strictEqual(
      hostA.shutdownCalls,
      1,
      'dispose must include retired states that are no longer routable',
    );
    assert.strictEqual(hosts.get('/b.mjs')!.shutdownCalls, 1);

    lintResult.resolve({ results: [] });
    await lint;
  });

  test('a failed replacement can be aborted without changing the active host', async () => {
    const hostA = new TestHost();
    const pool = new PluginLintPool(
      testLogger(),
      async (configs) => {
        if (configs[0].configPath === '/b.mjs')
          throw new Error('broken plugin');
        return hostA;
      },
      0,
    );

    assert.strictEqual(
      await pool.prepare(descriptor('a'), 'fingerprint-a', 'a'),
      true,
    );
    assert.strictEqual(await pool.commit('a'), true);
    assert.strictEqual(
      await pool.prepare(descriptor('b'), 'fingerprint-b', 'b'),
      false,
    );
    await pool.abort('b');

    await pool.lint(request('a'));
    assert.strictEqual(hostA.lintCalls, 1);
    assert.strictEqual(hostA.shutdownCalls, 0);

    await pool.dispose();
    assert.strictEqual(hostA.shutdownCalls, 1);
  });

  test('a degraded generation without a host rejects plugin lint instead of returning a false green', async () => {
    const pool = new PluginLintPool(testLogger(), async () => {
      throw new Error('broken first plugin');
    });

    assert.strictEqual(
      await pool.prepare(descriptor('a'), 'fingerprint-a', 'a'),
      false,
    );
    assert.strictEqual(await pool.commit('a'), true);
    await assert.rejects(
      pool.lint(request('a')),
      /pluginLint requested.*generation "a".*without an activated plugin host/,
    );

    await pool.dispose();
  });

  test('an empty native-only generation has no host and disposed lint stays benign', async () => {
    let createCalls = 0;
    const pool = new PluginLintPool(testLogger(), async () => {
      createCalls++;
      return new TestHost();
    });

    assert.strictEqual(await pool.prepare([], 'fingerprint-a', 'a'), true);
    assert.strictEqual(await pool.commit('a'), true);
    assert.strictEqual(createCalls, 0);
    await assert.rejects(
      pool.lint(request('a')),
      /pluginLint requested.*generation "a".*without an activated plugin host/,
    );

    await pool.dispose();
    assert.deepStrictEqual(await pool.lint(request('a')), { results: [] });
  });
});

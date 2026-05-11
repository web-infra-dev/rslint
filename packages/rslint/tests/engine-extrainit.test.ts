import { describe, test, expect } from '@rstest/core';
import path from 'node:path';
import { pathToFileURL } from 'node:url';

import { runEngine } from '../src/engine.js';

/**
 * Regression test for engine.ts's `init` payload construction.
 *
 * The payload spreads `opts.extraInit` BEFORE the four authoritative
 * fields (configs / eslintPluginEntries / runtime) so a caller's
 * extraInit cannot accidentally override them. We assert this by
 * spawning a fake binary (tests/fixtures/fake-rslint-binary.cjs) that
 * echoes the received init payload on stderr; the test compares the
 * echoed `configs` / `runtime` / `eslintPluginEntries` against what
 * engine.ts was given and against the colliding extraInit values.
 */

const FAKE_BIN = path.resolve(__dirname, 'fixtures', 'fake-rslint-binary.cjs');

interface CapturedInit {
  configs?: unknown;
  eslintPluginEntries?: unknown;
  runtime?: unknown;
  files?: unknown;
  bogusExtra?: unknown;
}

/**
 * Run engine.ts against the fake binary, capture the init payload it
 * received via stderr, parse it. The fake binary writes `__FAKE_INIT__`-
 * prefixed JSON per frame.
 */
async function captureInit(opts: {
  configs: unknown[];
  eslintPluginEntries: { prefix: string; ruleNames: string[] }[];
  runtime: { forceColor?: boolean; singleThreaded?: boolean };
  extraInit: Record<string, unknown>;
}): Promise<CapturedInit> {
  // engine.ts forwards `output` notifications from the child to
  // opts.stdout. The fake binary echoes its received init payload
  // through an `output` notification so we can capture it here.
  let stdoutBuf = '';
  const stdoutCapture = {
    write(chunk: string | Buffer): boolean {
      stdoutBuf += typeof chunk === 'string' ? chunk : chunk.toString();
      return true;
    },
  } as NodeJS.WritableStream;

  await runEngine({
    binPath: FAKE_BIN,
    goArgs: [],
    eslintPluginEntries: opts.eslintPluginEntries,
    workerConfigs: [],
    configs: opts.configs,
    runtime: opts.runtime,
    extraInit: opts.extraInit,
    stdout: stdoutCapture,
  });

  const marker = '__FAKE_INIT__';
  const i = stdoutBuf.indexOf(marker);
  if (i === -1) throw new Error(`no init line in stdout: ${stdoutBuf}`);
  const newlineAt = stdoutBuf.indexOf('\n', i);
  const jsonStr = stdoutBuf.slice(i + marker.length, newlineAt);
  return JSON.parse(jsonStr) as CapturedInit;
}

describe('engine.ts init payload construction', () => {
  test('extraInit cannot override configs / eslintPluginEntries / runtime', async () => {
    const realConfigs = [{ tag: 'REAL_CONFIG' }];
    // Use empty eslintPluginEntries so engine.ts's WorkerPool takes the
    // no-op `workerCount=0` fast path — pool.init() returns instantly
    // without actually loading any plugin module. That keeps this test
    // purely about wire-payload construction and free of fixture-plugin
    // resolution coupling.
    const realEntries: { prefix: string; ruleNames: string[] }[] = [];
    const realRuntime = { forceColor: true, singleThreaded: false };

    // A buggy or malicious caller stuffs colliding keys into extraInit.
    // The four authoritative fields below MUST win because engine.ts
    // spreads `extraInit` first.
    const evilExtraInit = {
      configs: [{ tag: 'HIJACKED' }],
      eslintPluginEntries: [{ prefix: 'evil', ruleNames: [] }],
      runtime: { forceColor: false, singleThreaded: true, bogus: 99 },
      // ...but additive (non-clashing) keys must still pass through.
      files: ['legitimate.ts'],
      bogusExtra: 'whatever',
    };

    const got = await captureInit({
      configs: realConfigs,
      eslintPluginEntries: realEntries,
      runtime: realRuntime,
      extraInit: evilExtraInit,
    });

    // Core fields — must match runEngine's inputs verbatim, NOT the
    // values that extraInit tried to inject.
    expect(got.configs).toEqual(realConfigs);
    expect(got.eslintPluginEntries).toEqual(realEntries);
    expect(got.runtime).toEqual({
      forceColor: true,
      singleThreaded: false,
    });
    // Non-clashing extraInit keys flow through.
    expect(got.files).toEqual(['legitimate.ts']);
    expect(got.bogusExtra).toBe('whatever');
  }, 30_000);
});

// silence unused-import lint — pathToFileURL is kept available for tests
// that grow into multi-plugin scenarios.
void pathToFileURL;

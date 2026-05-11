import { describe, test, expect } from '@rstest/core';
import { spawn } from 'node:child_process';
import { existsSync } from 'node:fs';
import path from 'node:path';

import { IpcClient } from '@rslint/eslint-plugin-runner';

/**
 * Wire-level smoke for the rslint Go binary.
 *
 * Spawns the real Go binary with no mode flag (the default `runCLI`
 * branch — every user-facing CLI invocation lands here) and drives it
 * through a minimal IpcClient — the same one engine.ts uses in
 * production. This proves the IPC handshake works end-to-end against
 * real OS stdio pipes without standing up a full WorkerPool. The
 * cooperating WorkerPool path is covered by the JS-side cli tests in
 * packages/rslint-test-tools.
 *
 * Verifies:
 *   - real OS stdio pipes round-trip length-prefixed IPC frames
 *   - Go reads the init payload, ack's `{ok:true}`, and waits for the
 *     parent to drive the lint phase (empty here)
 *   - sending no lint requests + closing our stdin shuts the binary
 *     down via its disconnect-watcher
 *
 * Skipped if the Go binary isn't built; CI runs `pnpm build:bin` first.
 */

const RSLINT_BIN = path.resolve(__dirname, '..', 'bin', 'rslint');

describe('rslint binary IPC spawn smoke', () => {
  test('init handshake over stdio, clean shutdown on stdin EOF', async () => {
    if (!existsSync(RSLINT_BIN)) {
      console.warn('[skip] rslint binary not built; run `pnpm build:bin`');
      return;
    }

    // No mode flag — runCLI is the default branch. Pass --start-time
    // so the binary's flag.Parse picks something up; positional args
    // are empty so the lint scope is "nothing to lint".
    const child = spawn(RSLINT_BIN, [`--start-time=${Date.now()}`], {
      stdio: ['pipe', 'pipe', 'pipe'],
    });

    const stderrChunks: Buffer[] = [];
    child.stderr.on('data', (c: Buffer) => stderrChunks.push(c));

    const ipc = new IpcClient(child.stdout!, child.stdin!);
    ipc.start();

    try {
      // Minimal init: no configs, no plugins, no files. Go ack's
      // {ok:true}, then enters the lint pipeline against an empty
      // scope — which exits cleanly with no diagnostics.
      const initResp = await ipc.sendRequest('init', {
        configs: [],
        eslintPluginEntries: [],
        runtime: { singleThreaded: true },
      });
      expect((initResp.data as { ok: boolean }).ok).toBe(true);
    } finally {
      // Closing our IPC end (which closes child.stdin) triggers the
      // binary's stdin EOF → disconnect path. We DON'T register an
      // inbound handler for `shutdown` here, so Go's shutdown
      // request would otherwise time out on us — closing first lets
      // it bail cleanly.
      ipc.close();
    }

    const exitCode: number = await new Promise((resolveExit) => {
      child.on('exit', (code) => resolveExit(code ?? -1));
    });

    const stderrStr = Buffer.concat(stderrChunks).toString('utf8');

    // The previous `[0,1,2].toContain(exitCode)` accepted any of
    // three plausible outcomes — so the "clean shutdown on stdin
    // EOF" half of the test name was never actually verified.
    //
    // Pin the EXPECTED exit code (1 = "no rslint config file
    // found") and ALSO assert the binary printed the canonical
    // config-not-found marker on stderr. Together they prove:
    //   (a) the binary reached its lint phase (so the init
    //       handshake genuinely completed),
    //   (b) it exited cooperatively for the documented reason,
    //       not via crash / timeout / signal.
    // A regression where the init handshake actually fails would
    // produce a non-1 code OR omit the marker. A regression where
    // the binary hangs would hit the 30s test timeout.
    if (exitCode !== 1) {
      console.error('---stderr---\n' + stderrStr);
    }
    expect(exitCode).toBe(1);
    // The Go binary writes "No rslint config file found" when it
    // can't locate a config. Anchor on a stable substring; an
    // exact-match regex would flake on platform-specific path
    // formatting in the surrounding diagnostic.
    expect(stderrStr.toLowerCase()).toContain('no rslint config');
  }, 30000); // 30s — first cold Go binary launch can take a few seconds
});

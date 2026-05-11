import { describe, test, expect } from '@rstest/core';
import { spawn, ChildProcess } from 'node:child_process';
import { existsSync } from 'node:fs';
import path from 'node:path';

/**
 * End-to-end signal-forwarding integration for engine.ts.
 *
 * The hand-written unit test in `engine-signal-forward.test.ts` only
 * verifies that listeners are registered (its header explicitly notes
 * that emitting a synthetic SIGINT inside rstest tears down the runner
 * itself). The contract we actually want to guarantee — that a real
 * SIGINT delivered to a real `runEngine`-hosted Node parent forwards
 * to its Go child and exits the whole tree within the documented
 * grace window — needs a child-process boundary so signals don't
 * touch the test runner.
 *
 * Setup:
 *   parent: `tests/fixtures/engine-runner-stub.cjs` — calls
 *           `runEngine` with `binPath` pointing at the fake binary.
 *   child:  `tests/fixtures/long-running-fake-binary.cjs` — ack's
 *           init, parks, logs `__FAKE_PID__` so we can later poll
 *           `process.kill(pid, 0)` to confirm the OS reaped it.
 *
 * Path under test:
 *   1. stub spawns fake-binary, sends `init`, awaits the fake's ack
 *   2. stub is now sitting in `await childExit`
 *   3. test sends SIGINT to stub
 *   4. engine.ts's onSignal fires → safeKillGo → SIGTERM to fake-binary
 *   5. fake-binary's SIGTERM handler logs marker, exits 143
 *   6. stub's await childExit resolves; cleanup runs; stub exits
 *   7. test asserts: exit happened fast, marker present, fake's PID is gone
 */

const STUB_PATH = path.resolve(__dirname, 'fixtures', 'engine-runner-stub.mjs');
const FAKE_BIN_PATH = path.resolve(
  __dirname,
  'fixtures',
  'long-running-fake-binary.cjs',
);
// A hybrid fake binary that ack's init (so engine reaches steady
// state) but ignores SIGTERM. Used to verify engine.ts's real
// safeKillGo escalates to SIGKILL — the previous safe-kill.test.ts
// only tested a local copy of the pattern, not engine's actual code.
const SIGTERM_IGNORING_FAKE_BIN_PATH = path.resolve(
  __dirname,
  'fixtures',
  'sigterm-ignoring-fake-binary.cjs',
);

// The stub imports from packages/rslint/dist/engine.js (a deliberate
// test-only direct path, so we don't have to alter package.json
// exports or src/index.ts re-exports for a fixture). CI builds dist
// before tests; locally a stale checkout may not have. Skip if
// missing — fail-soft beats a confusing import-not-found stack trace.
const DIST_ENGINE = path.resolve(__dirname, '..', 'dist', 'engine.js');

interface SpawnedStub {
  child: ChildProcess;
  stderrChunks: Buffer[];
  /** Resolves with the fake binary's reported PID once init completes. */
  fakePid: Promise<number>;
  /** Resolves once the child process emits `exit`. */
  exited: Promise<{ code: number | null; signal: string | null }>;
}

function spawnStub(fakeBinOverride?: string): SpawnedStub {
  const child = spawn(process.execPath, [STUB_PATH], {
    stdio: ['pipe', 'pipe', 'pipe'],
    env: { ...process.env, FAKE_BIN_PATH: fakeBinOverride ?? FAKE_BIN_PATH },
  });
  const stderrChunks: Buffer[] = [];
  let fakePidResolve!: (pid: number) => void;
  const fakePid = new Promise<number>((r) => {
    fakePidResolve = r;
  });
  child.stderr.on('data', (chunk: Buffer) => {
    stderrChunks.push(chunk);
    const text = chunk.toString();
    const m = text.match(/__FAKE_PID__:(\d+)/);
    if (m) fakePidResolve(parseInt(m[1], 10));
  });

  const exited = new Promise<{ code: number | null; signal: string | null }>(
    (resolve) => {
      child.once('exit', (code, signal) =>
        resolve({ code, signal: signal ?? null }),
      );
    },
  );
  return { child, stderrChunks, fakePid, exited };
}

function joinedStderr(chunks: Buffer[]): string {
  return Buffer.concat(chunks).toString();
}

async function waitForMarker(
  chunks: Buffer[],
  marker: string,
  timeoutMs: number,
): Promise<void> {
  const start = Date.now();
  while (Date.now() - start < timeoutMs) {
    if (joinedStderr(chunks).includes(marker)) return;
    await new Promise((r) => setTimeout(r, 25));
  }
  throw new Error(
    `timeout waiting for marker ${JSON.stringify(marker)}; stderr so far:\n${joinedStderr(chunks)}`,
  );
}

/**
 * Polls `process.kill(pid, 0)`, which throws ESRCH once the OS has
 * reaped the process. Use AFTER the parent stub has exited, so the
 * child has had a chance to detach.
 */
async function waitForPidGone(pid: number, timeoutMs: number): Promise<void> {
  const start = Date.now();
  while (Date.now() - start < timeoutMs) {
    try {
      process.kill(pid, 0);
    } catch (err) {
      if ((err as NodeJS.ErrnoException).code === 'ESRCH') return;
      throw err;
    }
    await new Promise((r) => setTimeout(r, 25));
  }
  throw new Error(`PID ${pid} still alive after ${timeoutMs}ms`);
}

describe('CLI signal-forwarding integration', () => {
  test('SIGINT on Node parent terminates the Go child and exits clean', async () => {
    if (!existsSync(DIST_ENGINE)) {
      console.warn(
        '[skip] packages/rslint/dist/engine.js missing; run `pnpm build`',
      );
      return;
    }
    const stub = spawnStub();
    try {
      // Wait for fake binary's init-ack marker — guarantees the
      // engine handler is wired AND we're in `await childExit`.
      await waitForMarker(stub.stderrChunks, '__FAKE_INIT_OK__', 5_000);
      const fakePid = await stub.fakePid;

      const start = Date.now();
      stub.child.kill('SIGINT');
      const exitInfo = await stub.exited;
      const elapsed = Date.now() - start;

      // Engine.ts forwarded SIGINT → safeKillGo → SIGTERM. Fake
      // binary's SIGTERM handler logged the marker and exited.
      // If the engine's forwarding ever regresses, the fake child
      // would NOT see SIGTERM and this marker would be absent.
      const stderr = joinedStderr(stub.stderrChunks);
      expect(stderr).toContain('__FAKE_SIGTERM__');

      // Stub itself exits — code != 0 (130 from SIGINT-via-signal-
      // fallback, see engine.ts:144). The runtime may surface the
      // exit either as `code: 130` or `signal: 'SIGINT'` depending
      // on platform/timing — accept either.
      const killedByCode = exitInfo.code != null && exitInfo.code !== 0;
      const killedBySignal = exitInfo.signal === 'SIGINT';
      expect(killedByCode || killedBySignal).toBe(true);

      // Lower bound: ≥ 1 ms after the signal was sent. Without a
      // lower bound, a regression that returned an exitCode IMMEDIATELY
      // on signal receipt (without waiting for the child to actually
      // exit via the `await childExit` path) would slip through —
      // the orphan child would then be caught only by the PID-gone
      // check below. The lower bound here pins the awaited-exit
      // ordering explicitly.
      expect(elapsed).toBeGreaterThanOrEqual(1);
      // Upper bound: safeKillGo grace is 5s before SIGKILL; cooperating
      // fake responds well within that window. 7s = grace + Node teardown.
      expect(elapsed).toBeLessThan(7_000);

      // Whole process tree is gone — no orphan fake binary left.
      await waitForPidGone(fakePid, 3_000);
    } finally {
      if (stub.child.exitCode == null && stub.child.signalCode == null) {
        stub.child.kill('SIGKILL');
      }
    }
  }, 15_000);

  test('SIGTERM on Node parent terminates the Go child and exits clean', async () => {
    if (!existsSync(DIST_ENGINE)) {
      console.warn(
        '[skip] packages/rslint/dist/engine.js missing; run `pnpm build`',
      );
      return;
    }
    const stub = spawnStub();
    try {
      await waitForMarker(stub.stderrChunks, '__FAKE_INIT_OK__', 5_000);
      const fakePid = await stub.fakePid;

      const start = Date.now();
      stub.child.kill('SIGTERM');
      const exitInfo = await stub.exited;
      const elapsed = Date.now() - start;

      const stderr = joinedStderr(stub.stderrChunks);
      expect(stderr).toContain('__FAKE_SIGTERM__');

      const killedByCode = exitInfo.code != null && exitInfo.code !== 0;
      const killedBySignal = exitInfo.signal === 'SIGTERM';
      expect(killedByCode || killedBySignal).toBe(true);

      expect(elapsed).toBeGreaterThanOrEqual(1);
      expect(elapsed).toBeLessThan(7_000);

      await waitForPidGone(fakePid, 3_000);
    } finally {
      if (stub.child.exitCode == null && stub.child.signalCode == null) {
        stub.child.kill('SIGKILL');
      }
    }
  }, 15_000);

  test('engine.ts safeKillGo escalates to SIGKILL when child ignores SIGTERM', async () => {
    if (!existsSync(DIST_ENGINE)) {
      console.warn(
        '[skip] packages/rslint/dist/engine.js missing; run `pnpm build`',
      );
      return;
    }
    // Drive a child that ack's init (so engine reaches steady
    // state) but DELIBERATELY IGNORES SIGTERM. engine.ts's real
    // safeKillGo must:
    //   1. send SIGTERM (fixture confirms via marker)
    //   2. wait the documented grace window (~5s)
    //   3. send SIGKILL (un-catchable; ends the process)
    //
    // Previously this contract was tested by safe-kill.test.ts but
    // that test replicated the safeKillGo pattern INLINE rather
    // than going through engine.ts. A regression in engine's
    // real safeKillGo (e.g. removing the SIGKILL backstop, leaving
    // a stuck child forever) would not have been caught. This test
    // closes that gap end-to-end.
    const stub = spawnStub(SIGTERM_IGNORING_FAKE_BIN_PATH);
    try {
      await waitForMarker(stub.stderrChunks, '__FAKE_INIT_OK__', 5_000);
      const fakePid = await stub.fakePid;

      const start = Date.now();
      stub.child.kill('SIGINT');
      const exitInfo = await stub.exited;
      const elapsed = Date.now() - start;

      const stderr = joinedStderr(stub.stderrChunks);
      // Marker proves engine.ts actually sent SIGTERM (first step
      // of safeKillGo). Without this, the child would have exited
      // for some other reason and the test below would still pass
      // a naive elapsed-bound — anchor explicitly.
      expect(stderr).toContain('__FAKE_RECEIVED_SIGTERM__');

      // The fixture explicitly does NOT exit on SIGTERM. So the
      // only way it died is engine.ts's SIGKILL backstop. We
      // expect elapsed ≈ 5s (grace) + a small overhead.
      // Lower bound 4.5s: a regression that immediately sent
      // SIGKILL (skipping the grace) would finish faster and
      // fail this. Upper bound 8s: a regression where the
      // SIGKILL backstop never fires would hit the 15s test
      // timeout instead.
      expect(elapsed).toBeGreaterThanOrEqual(4_500);
      expect(elapsed).toBeLessThan(8_000);

      // The child was killed by SIGKILL (un-catchable). Both
      // shapes accepted because Node may surface a SIGKILL'd
      // child as either signal-based or code-based depending
      // on platform/race.
      const wasKilled =
        exitInfo.signal === 'SIGKILL' ||
        (exitInfo.code != null && exitInfo.code !== 0);
      expect(wasKilled).toBe(true);

      // No orphan.
      await waitForPidGone(fakePid, 3_000);
    } finally {
      if (stub.child.exitCode == null && stub.child.signalCode == null) {
        stub.child.kill('SIGKILL');
      }
    }
  }, 20_000);
});

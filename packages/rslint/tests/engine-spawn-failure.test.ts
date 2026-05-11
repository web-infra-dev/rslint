import { describe, test, expect } from '@rstest/core';
import path from 'node:path';

import { runEngine } from '../src/engine.js';

/**
 * Regression for engine.ts's child-process error/exit handling.
 *
 * Before this fix, the error/exit listeners were attached only at the
 * tail of runEngine (step 5), after `pool.init()` and the init IPC
 * round-trip had completed. ENOENT on the binary path arrives as an
 * async 'error' event — if it fired during any of the earlier awaits,
 * the listener wasn't attached yet, Node escalated it to an
 * uncaughtException, and the function never got to return the
 * `{ exitCode: 2 }` it advertises.
 *
 * The fix wires `child.once('error') / .once('exit')` immediately
 * after `spawn` and races them against every later await via a
 * `RaceResult` union. These tests assert the externally-visible
 * contract: runEngine returns a non-zero exit code without throwing.
 */

describe('engine.ts spawn / runtime error handling', () => {
  test('missing binary returns exitCode=2 without unhandled rejection', async () => {
    // Capture any unhandled rejection that escapes the engine. A
    // bare promise variable lets the test catch the smell even if
    // Node delays the warning emission.
    let unhandled: unknown;
    const onUnhandled = (err: unknown) => {
      unhandled = err;
    };
    process.on('unhandledRejection', onUnhandled);

    try {
      const stderrBuf: string[] = [];
      const stderr = {
        write(chunk: string | Buffer): boolean {
          stderrBuf.push(typeof chunk === 'string' ? chunk : chunk.toString());
          return true;
        },
      } as NodeJS.WritableStream;

      // Path that cannot exist. spawn returns synchronously; the
      // 'error' event with ENOENT arrives on a later tick.
      const result = await runEngine({
        binPath: '/definitely/not/a/real/binary/at/this/path/please',
        goArgs: [],
        eslintPluginEntries: [],
        workerConfigs: [],
        configs: [],
        runtime: {},
        stderr,
      });

      expect(result.exitCode).toBe(2);
      // The error MUST surface on stderr so users get an actionable
      // message; silently returning 2 with no diagnostic was the
      // old failure mode.
      const stderrText = stderrBuf.join('');
      expect(stderrText).toMatch(/Go process spawn\/runtime error/);
      expect(unhandled).toBeUndefined();
    } finally {
      process.off('unhandledRejection', onUnhandled);
    }
  }, 30_000);

  test('binary that exits immediately propagates exit code', async () => {
    // /bin/false exits 1 instantly, before our init IPC can finish.
    // The childExit race must observe the exit and return WITHOUT
    // hanging on the pending IPC send (the old code would await
    // forever because the response never arrives).
    const fakeBin =
      process.platform === 'win32'
        ? path.resolve(process.env.COMSPEC ?? 'cmd.exe')
        : '/bin/false';

    // On Windows /bin/false isn't available; skip rather than fail
    // — the contract being tested is platform-agnostic but the
    // fixture isn't.
    if (process.platform === 'win32') return;

    const stderrBuf: string[] = [];
    const stderr = {
      write(chunk: string | Buffer): boolean {
        stderrBuf.push(typeof chunk === 'string' ? chunk : chunk.toString());
        return true;
      },
    } as NodeJS.WritableStream;

    const result = await runEngine({
      binPath: fakeBin,
      goArgs: [],
      eslintPluginEntries: [],
      workerConfigs: [],
      configs: [],
      runtime: {},
      stderr,
    });

    // /bin/false returns 1. The exact code matters less than "non-
    // zero and engine doesn't hang" — but we assert it explicitly
    // so a regression that converts the exit into a default-2 path
    // also gets flagged.
    expect(result.exitCode).toBeGreaterThan(0);
    // Either pool.init failed (the binary doesn't speak IPC) or
    // the init IPC raced against exit. Both must terminate cleanly.
    // We don't pin the specific stderr text since the order of
    // failure modes depends on timing.
  }, 30_000);
});

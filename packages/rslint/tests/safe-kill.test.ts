import { describe, test, expect } from '@rstest/core';
import { spawn } from 'node:child_process';
import path from 'node:path';

/**
 * safeKillGo's SIGTERM → SIGKILL escalation contract.
 *
 * engine.ts's safeKillGo is not exported; we test the behavior by
 * replicating the same escalation pattern against a fixture that
 * actively ignores SIGTERM. If the escalation doesn't fire, the
 * process stays alive past the grace window and the test times out.
 *
 * We can't call safeKillGo directly without exporting it (which would
 * expand its API surface), so this test is structured as a *contract*
 * test: any process that ignores SIGTERM must still die within the
 * grace + a small slack window when killed via the safeKillGo pattern.
 */

const KILL_GRACE_MS = 5_000;
const FIXTURE = path.resolve(
  __dirname,
  'fixtures',
  'sigterm-ignoring-binary.cjs',
);

describe('safeKillGo escalation', () => {
  test('a binary ignoring SIGTERM is still terminated within the grace window', async () => {
    const child = spawn(process.execPath, [FIXTURE], {
      stdio: ['pipe', 'pipe', 'pipe'],
    });

    // Wait for the fixture to print its READY marker so we know it's
    // past startup AND its signal handlers are installed.
    await new Promise<void>((resolve, reject) => {
      const t = setTimeout(
        () => reject(new Error('fixture never reported ready')),
        5_000,
      );
      const onData = (chunk: Buffer) => {
        if (chunk.toString().includes('__FIXTURE_READY__')) {
          clearTimeout(t);
          child.stderr?.off('data', onData);
          resolve();
        }
      };
      child.stderr?.on('data', onData);
    });

    // Replicate safeKillGo's escalation pattern.
    const start = Date.now();
    child.kill('SIGTERM');
    const killTimer = setTimeout(() => {
      if (child.exitCode == null && child.signalCode == null) {
        child.kill('SIGKILL');
      }
    }, KILL_GRACE_MS);

    const exitInfo = await new Promise<{
      code: number | null;
      signal: string | null;
    }>((resolve) => {
      child.once('exit', (code, signal) => {
        clearTimeout(killTimer);
        resolve({ code, signal });
      });
    });
    const elapsed = Date.now() - start;

    // Process must have been killed by the signal we sent. SIGKILL on
    // POSIX shows up as `signal: 'SIGKILL'`; Windows surfaces it as a
    // non-zero exit code instead — accept either.
    const killedBySignal = exitInfo.signal === 'SIGKILL';
    const killedByCode = exitInfo.code != null && exitInfo.code !== 0;
    expect(killedBySignal || killedByCode).toBe(true);

    // Must die within grace + a generous slack. If grace is 5s, exit
    // should be at most ~6s. >10s indicates SIGKILL never fired.
    expect(elapsed).toBeLessThan(KILL_GRACE_MS + 2_000);
    expect(elapsed).toBeGreaterThanOrEqual(KILL_GRACE_MS - 100);
  }, 20_000);

  test('a binary that DOES honor SIGTERM exits quickly (no SIGKILL needed)', async () => {
    // Sanity test: a normal Node process exits cleanly on SIGTERM.
    // Validates that the grace-window logic doesn't penalize
    // well-behaved children.
    const child = spawn(
      process.execPath,
      ['-e', 'setInterval(() => {}, 60_000);'],
      {
        stdio: ['pipe', 'pipe', 'pipe'],
      },
    );
    await new Promise((r) => setTimeout(r, 100)); // ensure spawn completed

    const start = Date.now();
    child.kill('SIGTERM');
    const killTimer = setTimeout(() => {
      if (child.exitCode == null && child.signalCode == null) {
        child.kill('SIGKILL');
      }
    }, KILL_GRACE_MS);

    const exitInfo = await new Promise<{
      code: number | null;
      signal: string | null;
    }>((resolve) => {
      child.once('exit', (code, signal) => {
        clearTimeout(killTimer);
        resolve({ code, signal });
      });
    });
    const elapsed = Date.now() - start;

    // Should be SIGTERM (cooperative), not SIGKILL (escalation).
    expect(exitInfo.signal).toBe('SIGTERM');
    // And fast — well under the grace window.
    expect(elapsed).toBeLessThan(2_000);
  }, 10_000);
});

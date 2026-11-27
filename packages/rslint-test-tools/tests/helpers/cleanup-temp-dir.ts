// Keep `packages/rslint/tests/cleanup-temp-dir.ts` and
// `packages/rslint-test-tools/tests/helpers/cleanup-temp-dir.ts` in sync.
import fs from 'node:fs/promises';

const RETRYABLE_CLEANUP_CODES = new Set([
  'EBUSY',
  'EPERM',
  'ENOTEMPTY',
  'EACCES',
  'EMFILE',
  'ENFILE',
]);

// Short initial probes clear millisecond-scale Windows child-exit and scanner
// races without stalling for a full second; sustained 1s steps preserve >20s
// of slack when a busy CI runner delays kernel handle teardown.
const DEFAULT_RETRY_DELAYS_MS: readonly number[] = [
  50,
  100,
  200,
  500,
  ...Array.from({ length: 20 }, () => 1_000),
];

export interface CleanupTempDirOptions {
  retryDelaysMs?: readonly number[];
  rmImpl?: (
    targetPath: string,
    options: { recursive: true; force: true },
  ) => Promise<void>;
  sleepImpl?: (ms: number) => Promise<void>;
}

function defaultSleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function isRetryableCleanupError(error: unknown): boolean {
  if (typeof error !== 'object' || error === null || !('code' in error)) {
    return false;
  }
  const { code } = error as { code?: unknown };
  return typeof code === 'string' && RETRYABLE_CLEANUP_CODES.has(code);
}

/**
 * Removes a temporary test directory or file with progressive retries for
 * transient Windows filesystem locks (such as an exiting child process still
 * releasing its working directory or executable image handle).
 *
 * Test execution and assertions are never retried, and persistent lock errors
 * are rethrown once retries are exhausted. When called from Rstest hooks in
 * packages without a global `hookTimeout` override (such as `packages/rslint`),
 * pass an explicit hook timeout `>= 30_000` (e.g. `35 * 60_000`).
 */
export async function cleanupTempDir(
  targetPath: string,
  options: CleanupTempDirOptions = {},
): Promise<void> {
  if (!targetPath) return;
  const delays = options.retryDelaysMs ?? DEFAULT_RETRY_DELAYS_MS;
  const rm = options.rmImpl ?? fs.rm;
  const sleep = options.sleepImpl ?? defaultSleep;

  for (let attempt = 0; ; attempt++) {
    try {
      await rm(targetPath, { recursive: true, force: true });
      return;
    } catch (error) {
      if (attempt >= delays.length || !isRetryableCleanupError(error)) {
        throw error;
      }
      await sleep(delays[attempt]);
    }
  }
}

import fs from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { describe, expect, test } from 'rstack/test';
import { cleanupTempDir } from './cleanup-temp-dir.js';

function makeFsError(code: string, message = `${code}: mock lock`): Error {
  return Object.assign(new Error(message), { code });
}

describe('cleanupTempDir', () => {
  test('removes a real directory including read-only files and tolerates real-timer retries', async () => {
    const dir = await fs.mkdtemp(path.join(os.tmpdir(), 'rslint-cleanup-ok-'));
    const readOnlyFile = path.join(dir, 'readonly.pack');
    try {
      await fs.writeFile(readOnlyFile, 'ok', 'utf8');
      await fs.chmod(readOnlyFile, 0o444);
      let injectedLockOnce = true;

      await cleanupTempDir(dir, {
        rmImpl: async (target, options) => {
          if (injectedLockOnce) {
            injectedLockOnce = false;
            throw makeFsError('EBUSY');
          }
          await fs.rm(target, options);
        },
      });

      await expect(fs.access(dir)).rejects.toMatchObject({ code: 'ENOENT' });
    } finally {
      await fs.chmod(readOnlyFile, 0o666).catch(() => undefined);
      await fs.rm(dir, { recursive: true, force: true }).catch(() => undefined);
    }
  });

  test('removes an unlocked real directory with default options', async () => {
    const dir = await fs.mkdtemp(path.join(os.tmpdir(), 'rslint-cleanup-dir-'));
    try {
      await fs.writeFile(path.join(dir, 'probe.txt'), 'ok', 'utf8');

      await cleanupTempDir(dir);

      await expect(fs.access(dir)).rejects.toMatchObject({ code: 'ENOENT' });
    } finally {
      await fs.rm(dir, { recursive: true, force: true }).catch(() => undefined);
    }
  });

  test('skips the retry timer when removal succeeds on the first attempt', async () => {
    const sleeps: number[] = [];
    let attempts = 0;

    await cleanupTempDir('/virtual/unlocked', {
      rmImpl: async (_target, options) => {
        expect(options).toEqual({ recursive: true, force: true });
        attempts++;
      },
      sleepImpl: async (ms) => {
        sleeps.push(ms);
      },
    });

    expect(attempts).toBe(1);
    expect(sleeps).toEqual([]);
  });

  test('preserves fast initial probes and >=210s cumulative wait in the default schedule', async () => {
    let virtualElapsedMs = 0;
    const sleepsOn25sLock: number[] = [];

    await cleanupTempDir('/virtual/locked-25s', {
      sleepImpl: async (ms) => {
        sleepsOn25sLock.push(ms);
        virtualElapsedMs += ms;
      },
      rmImpl: async () => {
        if (virtualElapsedMs < 25_000) {
          throw makeFsError('EBUSY');
        }
      },
    });

    expect(sleepsOn25sLock.slice(0, 4)).toEqual([50, 100, 200, 500]);
    expect(virtualElapsedMs).toBeGreaterThanOrEqual(25_000);
    expect(virtualElapsedMs).toBeLessThanOrEqual(30_000);

    const exhaustedSleeps: number[] = [];
    await expect(
      cleanupTempDir('/virtual/locked-forever', {
        sleepImpl: async (ms) => {
          exhaustedSleeps.push(ms);
        },
        rmImpl: async () => {
          throw makeFsError('EBUSY');
        },
      }),
    ).rejects.toMatchObject({ code: 'EBUSY' });

    const totalBudgetMs = exhaustedSleeps.reduce((sum, ms) => sum + ms, 0);
    expect(totalBudgetMs).toBeGreaterThanOrEqual(210_000);
  });

  test('treats an empty target path as a no-op', async () => {
    let calls = 0;
    await cleanupTempDir('', {
      rmImpl: async () => {
        calls++;
      },
    });
    expect(calls).toBe(0);
  });

  test.each(['EBUSY', 'EPERM', 'ENOTEMPTY', 'EACCES', 'EMFILE', 'ENFILE'])(
    'retries transient %s errors with progressive backoff until unlocked',
    async (code) => {
      const sleeps: number[] = [];
      let attempts = 0;

      await cleanupTempDir('/virtual/temp', {
        retryDelaysMs: [50, 100, 200],
        sleepImpl: async (ms) => {
          sleeps.push(ms);
        },
        rmImpl: async (_target, options) => {
          expect(options).toEqual({ recursive: true, force: true });
          attempts++;
          if (attempts < 3) {
            throw makeFsError(code);
          }
        },
      });

      expect(attempts).toBe(3);
      expect(sleeps).toEqual([50, 100]);
    },
  );

  test('rethrows non-retryable errors immediately without sleeping', async () => {
    const fatal = makeFsError('EINVAL', 'invalid path');
    const sleeps: number[] = [];
    let attempts = 0;

    await expect(
      cleanupTempDir('/virtual/temp', {
        retryDelaysMs: [50, 100],
        sleepImpl: async (ms) => {
          sleeps.push(ms);
        },
        rmImpl: async () => {
          attempts++;
          throw fatal;
        },
      }),
    ).rejects.toBe(fatal);

    expect(attempts).toBe(1);
    expect(sleeps).toEqual([]);
  });

  test('rethrows the persistent lock error once retries are exhausted', async () => {
    const persistent = makeFsError('EBUSY', 'EBUSY: still locked');
    const sleeps: number[] = [];
    let attempts = 0;

    await expect(
      cleanupTempDir('/virtual/temp', {
        retryDelaysMs: [50, 100],
        sleepImpl: async (ms) => {
          sleeps.push(ms);
        },
        rmImpl: async () => {
          attempts++;
          throw persistent;
        },
      }),
    ).rejects.toBe(persistent);

    expect(attempts).toBe(3);
    expect(sleeps).toEqual([50, 100]);
  });

  test('stays byte-for-byte identical across rslint and rslint-test-tools copies', async () => {
    const coreCopy = fileURLToPath(
      new URL('./cleanup-temp-dir.ts', import.meta.url),
    );
    const toolsCopy = fileURLToPath(
      new URL(
        '../../rslint-test-tools/tests/helpers/cleanup-temp-dir.ts',
        import.meta.url,
      ),
    );
    const [coreSource, toolsSource] = await Promise.all([
      fs.readFile(coreCopy, 'utf8'),
      fs.readFile(toolsCopy, 'utf8'),
    ]);
    expect(toolsSource).toBe(coreSource);
  });
});

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
  test('removes a real directory including read-only files without sleeping when unlocked', async () => {
    const dir = await fs.mkdtemp(path.join(os.tmpdir(), 'rslint-cleanup-ok-'));
    const readOnlyFile = path.join(dir, 'readonly.pack');
    await fs.writeFile(readOnlyFile, 'ok', 'utf8');
    await fs.chmod(readOnlyFile, 0o444);
    const sleeps: number[] = [];

    await cleanupTempDir(dir, {
      sleepImpl: async (ms) => {
        sleeps.push(ms);
      },
    });

    await expect(fs.access(dir)).rejects.toMatchObject({ code: 'ENOENT' });
    expect(sleeps).toEqual([]);
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

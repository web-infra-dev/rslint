import { execFileSync } from 'node:child_process';
import { mkdtempSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { resolve } from 'node:path';
import { cleanupTempDir } from '../../helpers/cleanup-temp-dir.js';

/** Build the test-only Go producer once per test run, never reuse a stale frame schema. */
export default async function setup() {
  const directory = mkdtempSync(resolve(tmpdir(), 'rslint-estree-test-'));
  const binary = resolve(
    directory,
    process.platform === 'win32' ? 'projector.exe' : 'projector',
  );
  try {
    execFileSync(
      'go',
      [
        'build',
        '-o',
        binary,
        './internal/languages/js/estree/testdata/projector',
      ],
      {
        cwd: resolve(import.meta.dirname, '../../../../..'),
        stdio: 'pipe',
      },
    );
  } catch (error) {
    await cleanupTempDir(directory);
    throw error;
  }
  process.env.RSLINT_ESTREE_PROJECTOR = binary;
  return async () => {
    delete process.env.RSLINT_ESTREE_PROJECTOR;
    await cleanupTempDir(directory);
  };
}

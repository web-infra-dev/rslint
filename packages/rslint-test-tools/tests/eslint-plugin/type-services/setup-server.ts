import { execFileSync } from 'node:child_process';
import { mkdtempSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { resolve } from 'node:path';
import { cleanupTempDir } from '../../helpers/cleanup-temp-dir.js';

/** Build the test-only Go type-service server once per test run across split suites. */
export default async function setup() {
  const directory = mkdtempSync(resolve(tmpdir(), 'rslint-type-server-'));
  const binary = resolve(
    directory,
    process.platform === 'win32' ? 'types.exe' : 'types',
  );
  try {
    execFileSync(
      'go',
      ['build', '-o', binary, './internal/typeservice/testdata/server'],
      {
        cwd: resolve(import.meta.dirname, '../../../../..'),
        stdio: 'pipe',
      },
    );
  } catch (error) {
    await cleanupTempDir(directory);
    throw error;
  }
  process.env.RSLINT_TYPE_SERVICE_SERVER = binary;
  return async () => {
    delete process.env.RSLINT_TYPE_SERVICE_SERVER;
    await cleanupTempDir(directory);
  };
}

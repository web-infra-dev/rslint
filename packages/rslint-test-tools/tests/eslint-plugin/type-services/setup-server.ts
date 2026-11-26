import { execFileSync } from 'node:child_process';
import { mkdtempSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { resolve } from 'node:path';

/** Build the test-only Go type-service server once per test run across split suites. */
export default function setup() {
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
    rmSync(directory, { recursive: true, force: true });
    throw error;
  }
  process.env.RSLINT_TYPE_SERVICE_SERVER = binary;
  return () => {
    delete process.env.RSLINT_TYPE_SERVICE_SERVER;
    rmSync(directory, { recursive: true, force: true });
  };
}

import { execFileSync } from 'node:child_process';
import { mkdtempSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { resolve } from 'node:path';

/** Build the test-only Go producer once per test run, never reuse a stale frame schema. */
export default function setup() {
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
    rmSync(directory, { recursive: true, force: true });
    throw error;
  }
  process.env.RSLINT_ESTREE_PROJECTOR = binary;
  return () => {
    delete process.env.RSLINT_ESTREE_PROJECTOR;
    rmSync(directory, { recursive: true, force: true });
  };
}

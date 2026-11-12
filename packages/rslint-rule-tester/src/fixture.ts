import path from 'node:path';
import fs from 'node:fs/promises';
import { tmpdir } from 'node:os';

/** Creates a temporary fixture root containing the given files. */
export async function createFixtureRoot(
  files: Record<string, string>,
): Promise<string> {
  // Resolved imports use real paths. Keep the fixture root in the same path
  // space on macOS, where the temporary directory is a symlink.
  const root = await fs.realpath(
    await fs.mkdtemp(path.join(tmpdir(), 'rslint-rule-tester-')),
  );
  for (const [file, content] of Object.entries(files)) {
    const fullPath = path.join(root, file);
    await fs.mkdir(path.dirname(fullPath), { recursive: true });
    await fs.writeFile(fullPath, content, 'utf8');
  }
  return root;
}

export async function removeFixtureRoot(root: string): Promise<void> {
  // Windows can retain a child process's working directory during teardown,
  // so retry the removal.
  await fs.rm(root, {
    recursive: true,
    force: true,
    maxRetries: 20,
    retryDelay: 1_000,
  });
}

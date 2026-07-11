import { access, mkdtemp, rm } from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import { Rslint } from '@rslint/core';

const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-api-close-plugin-'));
const marker = path.join(tmp, 'started');
process.env.RSLINT_API_CLOSE_MARKER = marker;
const fixtureDirectory = path.resolve(
  import.meta.dirname,
  '../eslint-plugin/fixtures',
);
const rslint = new Rslint({
  cwd: fixtureDirectory,
  overrideConfigFile: 'close-hang.config.mjs',
});
const lint = rslint
  .lintText('const value = 1;\n', { filePath: 'probe.ts' })
  .catch(() => undefined);

try {
  const deadline = Date.now() + 15000;
  while (true) {
    try {
      await access(marker);
      break;
    } catch {
      if (Date.now() >= deadline) {
        console.error('plugin listener did not start');
        process.exit(2);
      }
      await new Promise((resolve) => setTimeout(resolve, 25));
    }
  }

  await rslint.close();
  await lint;
} finally {
  delete process.env.RSLINT_API_CLOSE_MARKER;
  await rm(tmp, { recursive: true, force: true });
}

import { access, mkdtemp, rm, writeFile } from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import { Rslint } from '@rslint/core';
import { RSLintService } from '@rslint/core/service';

const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-api-close-plugin-'));
const started = path.join(tmp, 'started');
const release = path.join(tmp, 'release');
process.env.RSLINT_API_CLOSE_STARTED = started;
process.env.RSLINT_API_CLOSE_RELEASE = release;
const fixtureDirectory = path.resolve(
  import.meta.dirname,
  '../eslint-plugin/fixtures',
);
const rslint = new Rslint({
  cwd: fixtureDirectory,
  overrideConfigFile: 'close-hang.config.mjs',
});
const originalServiceClose = RSLintService.prototype.close;
let serviceCloseStarted = false;
RSLintService.prototype.close = function serviceCloseWithProbe() {
  serviceCloseStarted = true;
  return Reflect.apply(originalServiceClose, this, []);
};
const lint = rslint.lintText('const value = 1;\n', {
  filePath: 'probe.ts',
});
let close;

try {
  while (true) {
    try {
      await access(started);
      break;
    } catch {
      await new Promise((resolve) => setTimeout(resolve, 10));
    }
  }

  let closeSettled = false;
  close = rslint.close().finally(() => {
    closeSettled = true;
  });
  await Promise.resolve();
  await Promise.resolve();
  if (closeSettled) {
    throw new Error('close settled before the active plugin gate was released');
  }
  if (serviceCloseStarted) {
    throw new Error(
      'close advanced to the service before plugin host shutdown completed',
    );
  }

  await writeFile(release, 'release');
  const [results] = await Promise.all([lint, close]);
  if (!serviceCloseStarted) {
    throw new Error('close never advanced to service shutdown');
  }
  if (
    results.length !== 1 ||
    results[0].filePath !== path.join(fixtureDirectory, 'probe.ts') ||
    results[0].messages.length !== 0
  ) {
    throw new Error(
      'active close returned an unexpected lint result: ' +
        JSON.stringify(results),
    );
  }
  const serialized = JSON.stringify(results);
  if (
    serialized.includes('CLOSE_GATE_LEAKED') ||
    serialized.includes('task_timeout') ||
    serialized.includes('worker_crashed')
  ) {
    throw new Error(`active close leaked a worker result: ${serialized}`);
  }
  console.log('FIXTURE_OK:close-active-plugin');
} finally {
  await writeFile(release, 'release').catch(() => undefined);
  await Promise.allSettled([lint, close].filter(Boolean));
  RSLintService.prototype.close = originalServiceClose;
  delete process.env.RSLINT_API_CLOSE_STARTED;
  delete process.env.RSLINT_API_CLOSE_RELEASE;
  await rm(tmp, { recursive: true, force: true });
}

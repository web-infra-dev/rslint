import path from 'node:path';

import type { LintTask } from '../../src/eslint-plugin/worker-pool.js';

/**
 * Shared fixtures + helpers for the WorkerPool end-to-end suite.
 *
 * The suite was split (by concern) into several
 * `worker-pool-e2e-<concern>.test.ts` files; this module holds the
 * constants and the `task(...)` factory that more than one of them use.
 *
 * It lives in `tests/` (NOT a subdirectory) so its `__dirname`-derived
 * `LOCAL_CONFIG_*` paths resolve against the tests directory, the same
 * as the test files' own `path.resolve(__dirname, 'fixtures', ...)`.
 *
 * It is deliberately NOT a `*.test.ts` file so the runner does not
 * collect it as a (zero-test) suite.
 */

export const LOCAL_CONFIG_PATH = path.resolve(
  __dirname,
  'fixtures',
  'local.config.mjs',
);
export const LOCAL_CONFIG_DIR = path.dirname(LOCAL_CONFIG_PATH);

export const localConfigs = [
  { configPath: LOCAL_CONFIG_PATH, configDirectory: LOCAL_CONFIG_DIR },
];

export function task(
  filePath: string,
  text: string,
  rule = 'local/no-null',
): LintTask {
  return {
    filePath,
    text,
    rules: { [rule]: { options: [] } },
    collectFixes: false,
    suggestionsMode: 'off',
    configKey: LOCAL_CONFIG_DIR,
  };
}

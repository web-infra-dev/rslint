import { fileURLToPath } from 'node:url';
import { dirname, resolve } from 'node:path';
import { setWorkerEntryForTests } from '../../src/eslint-plugin/worker-pool.js';

/**
 * rstest setup: point the lint worker at the built artifact.
 *
 * The worker tests run the source `.ts` through rstest, but `worker_threads`
 * can't execute TypeScript — it needs the rslib-built `dist/eslint-plugin/
 * lint-worker.js`. `worker-pool.ts` resolves the bundled sibling in production;
 * this test-only hook redirects it to the build output. Run `pnpm build` once
 * before testing so the artifact exists.
 */
const here = dirname(fileURLToPath(import.meta.url)); // tests/eslint-plugin
setWorkerEntryForTests(
  resolve(here, '../../dist/eslint-plugin/lint-worker.js'),
);

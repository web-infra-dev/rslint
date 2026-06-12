/**
 * Whether to skip the real-worker teardown e2e suites on win32.
 *
 * Tearing down a worker that holds the napi parser addon (`@rslint/native`)
 * used to abort below the JS layer on win32 (nodejs/node#34567), crashing the
 * rstest child — so these suites were win32-skipped. `terminateWorker`
 * (worker-pool.ts) now destroys the worker's stdio pipes before terminating,
 * which should remove that race.
 *
 * This flag is `false` so the suites RUN on win32 in CI, validating that
 * mitigation before the packaged VS Code extension ships the same addon in its
 * eslint-plugin worker on the win32 vsix (a forced terminate is reachable in
 * production via the per-task timeout / crash-respawn paths). If win32 CI shows
 * the abort returns, flip this to `true` to restore the skip and gate
 * eslintPlugins off on win32 in the extension's PluginLintPool.
 */
export const SKIP_WIN32_NAPI_TEARDOWN = false;

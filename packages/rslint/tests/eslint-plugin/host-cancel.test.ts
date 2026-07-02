import { describe, test, expect, rs } from '@rstest/core';

import { createPluginLintHost } from '../../src/eslint-plugin/host.js';
import { WorkerPool } from '../../src/eslint-plugin/worker-pool.js';
import type { EslintPluginLintRequest } from '../../src/eslint-plugin/plugin/plugin-lint-protocol.js';

import { LOCAL_CONFIG_DIR, localConfigs } from './worker-pool-e2e-helpers.js';
import { SKIP_WIN32_NAPI_TEARDOWN } from './win32-napi-teardown.js';

/**
 * Plugin-lint host — cancellation via AbortSignal. The LSP path bridges Go's
 * $/cancelRequest (a superseding keystroke / document close) to an AbortSignal;
 * host.lint must cancel the dispatched worker tasks so the worker stops instead
 * of running to completion.
 *
 * win32 teardown is gated by SKIP_WIN32_NAPI_TEARDOWN (same napi-teardown
 * reason as the other worker e2e suites); the flag is false so this runs too.
 */
describe.skipIf(SKIP_WIN32_NAPI_TEARDOWN && process.platform === 'win32')(
  'createPluginLintHost — cancellation via AbortSignal',
  () => {
    const req: EslintPluginLintRequest = {
      files: [
        { path: 'a.ts', text: 'const x = null;', configKey: LOCAL_CONFIG_DIR },
      ],
      rules: { 'local/no-null': { options: [] } },
      fix: false,
      suggestionsMode: 'off',
    };

    test('a pre-aborted signal cancels the dispatched task (no diagnostics)', async () => {
      const host = await createPluginLintHost(localConfigs);
      try {
        // Aborted before lint starts → onTaskDispatched sees signal.aborted and
        // cancels the task, so it produces no diagnostics.
        const ac = new AbortController();
        ac.abort();
        const cancelled = await host.lint(req, ac.signal);
        expect(cancelled.results).toHaveLength(1);
        expect(cancelled.results[0].diagnostics).toHaveLength(0);
        // Pin the positive cancellation signal, not just an empty diag array.
        expect(cancelled.results[0].cancelled).toBe(true);

        // Control: the SAME request without a signal runs to completion and the
        // local/no-null rule reports on the `null` literal — proving the empty
        // result above is the cancellation, not a plugin that failed to load.
        const normal = await host.lint(req);
        expect(normal.results).toHaveLength(1);
        expect(normal.results[0].diagnostics).toHaveLength(1);
        const diag = normal.results[0].diagnostics[0] as { ruleName?: string };
        expect(diag.ruleName).toBe('local/no-null');
      } finally {
        await host.shutdown();
      }
    }, 20_000);

    test('aborting after dispatch cancels via the onAbort listener (mid-flight)', async () => {
      const cancelSpy = rs.spyOn(WorkerPool.prototype, 'cancelTask');
      const host = await createPluginLintHost(localConfigs);
      try {
        const ac = new AbortController();
        // host.lint synchronously builds tasks + registers the abort listener,
        // then dispatches (onTaskDispatched pushes the id) before awaiting the
        // worker. Aborting now drives the onAbort-listener path (mid-flight),
        // distinct from the pre-abort branch above.
        const p = host.lint(req, ac.signal);
        ac.abort();
        const res = await p;
        expect(res.results).toHaveLength(1);
        // The cancel was routed to the worker pool (onAbort → cancelTask).
        expect(cancelSpy).toHaveBeenCalled();
      } finally {
        cancelSpy.mockRestore();
        await host.shutdown();
      }
    }, 20_000);
  },
);

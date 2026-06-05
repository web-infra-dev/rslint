import { describe, test, expect } from '@rstest/core';

import { WorkerPool } from '../../src/eslint-plugin/worker-pool.js';
import type { LintTask } from '../../src/eslint-plugin/worker-pool.js';

import { localConfigs, task } from './worker-pool-e2e-helpers.js';

/**
 * WorkerPool end-to-end — concurrent lintBatch (the #1050 runtime shape).
 *
 * #1050 made the Go core dispatch each homogeneous eslint-plugin batch as an
 * independent concurrent reverse request, so multiple `lintBatch` calls can be
 * in flight on the same WorkerPool at once. Each call's results must stay
 * partitioned to its own caller and rule — a cross-batch bleed (a result routed
 * to the wrong promise, or a doubled dispatch) would surface here.
 */
describe.skipIf(process.platform === 'win32')(
  'WorkerPool concurrent lintBatch (#1050 runtime shape)',
  () => {
    test('two batches in flight partition results to their own caller + rule', async () => {
      const pool = new WorkerPool({ configs: localConfigs, workerCount: 2 });
      await pool.init();

      const batchA: LintTask[] = Array.from({ length: 8 }, (_, i) =>
        task(`a${i}.ts`, `const x${i} = null;`, 'local/no-null'),
      );
      const batchB: LintTask[] = Array.from({ length: 8 }, (_, i) =>
        task(
          `b${i}.ts`,
          `const r${i} = [1].filter((x) => x > 0).length > 0;`,
          'local/prefer-array-some',
        ),
      );

      // Fire both WITHOUT awaiting the first.
      const [resA, resB] = await Promise.all([
        pool.lintBatch(batchA),
        pool.lintBatch(batchB),
      ]);

      expect(resA).toHaveLength(8);
      expect(resB).toHaveLength(8);
      // No cross-batch bleed: each result carries exactly its own batch's rule.
      for (const r of resA) {
        expect(r.diagnostics).toHaveLength(1);
        expect(r.diagnostics[0].ruleName).toBe('local/no-null');
      }
      for (const r of resB) {
        expect(r.diagnostics).toHaveLength(1);
        expect(r.diagnostics[0].ruleName).toBe('local/prefer-array-some');
      }

      await pool.shutdown();
    });
  },
);

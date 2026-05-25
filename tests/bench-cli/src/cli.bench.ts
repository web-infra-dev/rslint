import { spawnSync } from 'node:child_process';
import os from 'node:os';
import path from 'node:path';
import {
  addCodspeedCompatibleTask,
  type BenchTaskDurationOverride,
  createCodspeedCompatibleBench,
} from './utils/bench-runtime.js';
import { assertExists } from './utils/fs.js';
import {
  STAMP_FD,
  readPreBinaryLatencyNs,
  summarizeSamples,
} from './utils/pre-binary-stamp.js';

// This benchmark reports two tasks to tinybench/CodSpeed:
// 1) `cli@vscode`: parent start -> CLI process completion
// 2) `cli@vscode_before_go_exec`: parent start -> first Go binary invocation
const repoRoot = path.resolve(import.meta.dirname, '../../../');
const cliEntrypoint = path.join(repoRoot, 'packages/rslint/bin/rslint.cjs');
const vscodeRepoDir = path.join(os.tmpdir(), 'rslint-bench', 'vscode');
const benchmarkTaskName = 'cli@vscode';
const preBinaryBenchmarkName = 'cli@vscode_before_go_exec';
const execFileSyncPreloadPath = path.resolve(
  import.meta.dirname,
  '../scripts/exec-file-sync-preload.cjs',
);

const cliArgs = ['--format', 'jsonline', '--quiet', '.'] as const;
const cliSamplesNs: number[] = [];
const preBinarySamplesNs: number[] = [];
const pendingPreBinaryDurationsNs: number[] = [];
const bench = createCodspeedCompatibleBench();

/**
 * Runs one CLI invocation and returns both durations from the same start point.
 *
 * Timing model used here:
 * - `startedAtMs` is captured in the parent process right before `spawnSync`.
 * - `completedAtMs` is captured right after the CLI process finishes in parent.
 * - preload writes `interceptedAtMs` when child process first calls `execFileSync`.
 * - total latency ns: `(completedAtMs - startedAtMs) * 1e6`.
 * - before-go latency ns: `readPreBinaryLatencyNs()` => `(interceptedAtMs - startedAtMs) * 1e6`.
 *
 * Returned values are used to feed both benchmark tasks without running CLI twice.
 */
async function runCLI(): Promise<{ totalNs: number; beforeGoNs: number }> {
  const startedAtMs = performance.timeOrigin + performance.now();

  // Keep stdout/stderr detached from benchmark transport so wall time reflects
  // lint execution instead of output buffering costs.
  const result = spawnSync(
    process.execPath,
    ['--require', execFileSyncPreloadPath, cliEntrypoint, ...cliArgs],
    {
      cwd: vscodeRepoDir,
      env: {
        ...process.env,
        RSLINT_BENCH_STAMP_FD: String(STAMP_FD),
      },
      stdio: ['ignore', 'ignore', 'ignore', 'pipe'],
    },
  );

  if (result.error) {
    throw result.error;
  }
  if (result.signal != null) {
    throw new Error(`Benchmark CLI exited with signal ${result.signal}`);
  }
  if (typeof result.status !== 'number') {
    throw new Error('Benchmark CLI did not report an exit status');
  }

  const completedAtMs = performance.timeOrigin + performance.now();
  const totalNs = Math.round((completedAtMs - startedAtMs) * 1_000_000);
  const beforeGoNs = readPreBinaryLatencyNs(
    startedAtMs,
    result.output?.[STAMP_FD] ?? null,
  );

  return {
    totalNs,
    beforeGoNs,
  };
}

function toOverriddenDuration(ns: number): BenchTaskDurationOverride {
  return {
    overriddenDuration: ns / 1_000_000,
  };
}

await Promise.all([
  assertExists(cliEntrypoint),
  assertExists(vscodeRepoDir),
  assertExists(execFileSyncPreloadPath),
]);

addCodspeedCompatibleTask(bench, benchmarkTaskName, async () => {
  const sample = await runCLI();
  cliSamplesNs.push(sample.totalNs);
  preBinarySamplesNs.push(sample.beforeGoNs);
  pendingPreBinaryDurationsNs.push(sample.beforeGoNs);

  return toOverriddenDuration(sample.totalNs);
});
addCodspeedCompatibleTask(bench, preBinaryBenchmarkName, () => {
  const sample = pendingPreBinaryDurationsNs.shift();
  if (sample == null) {
    throw new Error(
      `${preBinaryBenchmarkName} has no paired sample from ${benchmarkTaskName}. ` +
        'This benchmark requires iteration-based runs (set RSLINT_BENCH_TIME_MS=0).',
    );
  }

  return toOverriddenDuration(sample);
});

console.error(
  `[bench] ${benchmarkTaskName}: running ${bench.iterations} iteration(s) against ${vscodeRepoDir}. Results print after all iterations finish.`,
);
await bench.run();
console.error(`[bench] ${benchmarkTaskName}: completed`);
/* rslint-disable */
console.table([summarizeSamples(benchmarkTaskName, cliSamplesNs)]);
console.table([summarizeSamples(preBinaryBenchmarkName, preBinarySamplesNs)]);

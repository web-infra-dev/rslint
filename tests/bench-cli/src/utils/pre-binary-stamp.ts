import path from 'node:path';

export const STAMP_FD = 3;

type ExecFileSyncStamp = {
  interceptedAtMs?: number;
  file?: string | null;
};

function parseExecFileSyncStamp(output: string | Buffer | null): {
  payload: ExecFileSyncStamp;
  rawOutput: string;
} {
  if (output == null) {
    throw new Error('Missing execFileSync stamp output');
  }

  const rawOutput =
    typeof output === 'string' ? output : output.toString('utf8');
  const lines = rawOutput
    .split('\n')
    .map((line) => line.trim())
    .filter((line) => line.length > 0);

  if (lines.length === 0) {
    throw new Error('Empty execFileSync stamp output');
  }

  return {
    payload: JSON.parse(lines[lines.length - 1]) as ExecFileSyncStamp,
    rawOutput,
  };
}

export function readPreBinaryLatencyNs(
  startedAtMs: number,
  output: string | Buffer | null,
): number {
  // "before_go_exec" latency is computed against the parent timestamp:
  // startedAtMs (parent pre-spawn) -> interceptedAtMs (child first execFileSync).
  const { payload, rawOutput } = parseExecFileSyncStamp(output);

  if (typeof payload.interceptedAtMs !== 'number') {
    throw new Error(`Invalid execFileSync stamp payload: ${rawOutput}`);
  }

  const expectedBinaryName =
    process.platform === 'win32' ? 'rslint.exe' : 'rslint';
  if (
    typeof payload.file !== 'string' ||
    path.basename(payload.file) !== expectedBinaryName
  ) {
    throw new Error(
      `Unexpected execFileSync target in stamp payload: ${rawOutput}`,
    );
  }

  return Math.round((payload.interceptedAtMs - startedAtMs) * 1_000_000);
}

export function summarizeSamples(
  taskName: string,
  samples: number[],
): Record<string, number | string> {
  if (samples.length === 0) {
    throw new Error(`No samples collected for ${taskName}`);
  }

  const sorted = [...samples].sort((left, right) => left - right);
  const total = samples.reduce((sum, sample) => sum + sample, 0);
  const middle = Math.floor(sorted.length / 2);
  const median =
    sorted.length % 2 === 0
      ? Math.round((sorted[middle - 1] + sorted[middle]) / 2)
      : sorted[middle];
  const avgNs = Math.round(total / samples.length);
  const throughputAvg = Number((1_000_000_000 / avgNs).toFixed(6));
  const throughputMed = Number((1_000_000_000 / median).toFixed(6));

  // Keep schema aligned with tinybench.table() output fields for side-by-side
  // comparison in benchmark logs.
  //
  // Metric definitions:
  // - Latency avg (ns): arithmetic mean of sample latency in nanoseconds.
  // - Latency med (ns): median latency in nanoseconds (robust to outliers).
  // - Throughput avg (ops/s): 1e9 / Latency avg, estimated operations per second.
  // - Throughput med (ops/s): 1e9 / Latency med, median-based ops/s estimate.
  // - Samples: number of benchmark samples included in this summary.
  return {
    'Task name': taskName,
    'Latency avg (ns)': avgNs,
    'Latency med (ns)': median,
    'Throughput avg (ops/s)': throughputAvg,
    'Throughput med (ops/s)': throughputMed,
    Samples: samples.length,
  };
}

import { withCodSpeed } from '@codspeed/tinybench-plugin';
import { Bench } from 'tinybench';

export type BenchTaskDurationOverride = {
  overriddenDuration: number;
};

type BenchTaskFnResult = void | BenchTaskDurationOverride;

export function readIntEnv(name: string, fallback: number): number {
  const value = process.env[name];
  if (!value) {
    return fallback;
  }

  const parsed = Number.parseInt(value, 10);
  return Number.isNaN(parsed) ? fallback : parsed;
}

function createBenchmarkOptions() {
  return {
    iterations: readIntEnv('RSLINT_BENCH_ITERATIONS', 10),
    time: readIntEnv('RSLINT_BENCH_TIME_MS', 0),
    // CodSpeed plugin v5.2.0 expects latency samples on each task result.
    retainSamples: true,
    warmup: false,
  };
}

export function createCodspeedCompatibleBench(): Bench {
  const bench = new Bench({
    ...createBenchmarkOptions(),
    throws: true,
  });

  // tinybench@6 exposes runtime options on the instance. CodSpeed plugin
  // v5.2.0 still reads `bench.opts`, so provide a compatibility getter.
  if (Reflect.get(bench, 'opts') == null) {
    Object.defineProperty(bench, 'opts', {
      configurable: true,
      get() {
        return {
          iterations: bench.iterations,
          time: bench.time,
          warmup: bench.warmup,
          warmupIterations: bench.warmupIterations,
          warmupTime: bench.warmupTime,
          throws: bench.throws,
        };
      },
    });
  }

  return withCodSpeed(bench);
}

export function addCodspeedCompatibleTask(
  bench: Bench,
  name: string,
  fn: () => BenchTaskFnResult | Promise<BenchTaskFnResult>,
): void {
  bench.add(name, fn);
  const task = bench.getTask(name);

  // tinybench@6 keeps task fn in private state, while CodSpeed simulation mode
  // still reads `task.fn` / `task.fnOpts`.
  if (task) {
    Reflect.set(task, 'fn', fn);
    if (Reflect.get(task, 'fnOpts') == null) {
      Reflect.set(task, 'fnOpts', {});
    }
  }
}

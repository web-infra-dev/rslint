// Subprocess entry for process-isolated worker-pool scenarios.
//
// Why this exists: the worker-pool e2e scenarios that force a worker
// `terminate()` can native-abort BELOW the JS layer on Windows — libuv tears
// down the oxc-napi worker's stdio pipes during terminate and faults. When the
// pool runs inside the rstest test process, that abort crashes the test process
// itself ("Rstest exited unexpectedly"). Running the pool in THIS throwaway
// subprocess confines the abort here; the parent rstest test (harness.ts)
// observes our outcome via milestones and never crashes.
//
// Progress is reported to a milestone FILE (env RSLINT_MILESTONE_FILE) via
// synchronous appendFileSync: a milestone written immediately before a native
// abort is already on disk, so the parent can always tell how far the pool got.
// process.stdout is NOT used for milestones — its userland buffer can be lost
// when the process aborts.
import fs from 'node:fs';
import path from 'node:path';
import os from 'node:os';
import { pathToFileURL, fileURLToPath } from 'node:url';

const HERE = path.dirname(fileURLToPath(import.meta.url));
// Shared fixtures live one level up (tests/eslint-plugin/fixtures); the worker
// imports the config via plugin-loader's pathToFileURL, so absolute paths work
// cross-platform.
const FIXTURES = path.resolve(HERE, '../fixtures');
const LOCAL_CONFIG = {
  configPath: path.join(FIXTURES, 'local.config.mjs'),
  configDirectory: FIXTURES,
};
const HANG_CONFIG = {
  configPath: path.join(FIXTURES, 'hang.config.mjs'),
  configDirectory: FIXTURES,
};

const MILESTONE_FILE = process.env.RSLINT_MILESTONE_FILE;
const report = (obj) => {
  if (!MILESTONE_FILE) return;
  try {
    fs.appendFileSync(MILESTONE_FILE, JSON.stringify(obj) + '\n');
  } catch {
    /* parent or disk gone — nothing we can do, and we must not throw */
  }
};
const milestone = (name) => report({ kind: 'milestone', name });
const check = (name, pass, detail) =>
  report({ kind: 'assert', name, pass: !!pass, detail });

async function loadWorkerPool() {
  const dist = process.env.RSLINT_DIST_ESLINT_PLUGIN;
  if (!dist) throw new Error('RSLINT_DIST_ESLINT_PLUGIN not set');
  // Production resolution path: WorkerPool resolves its sibling
  // dist/eslint-plugin/lint-worker.js automatically — no setWorkerEntryForTests.
  // pathToFileURL is required on Windows: ESM import() rejects a bare absolute
  // path like `C:\...` (ERR_UNSUPPORTED_ESM_URL_SCHEME) and needs a file:// URL.
  const mod = await import(pathToFileURL(dist).href);
  return mod.WorkerPool;
}

const makeFixtureDir = (files) => {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-pool-iso-fx-'));
  for (const [name, content] of Object.entries(files)) {
    fs.writeFileSync(path.join(dir, name), content);
  }
  return dir;
};

const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
const localTask = (filePath, text, rule = 'local/no-null') => ({
  filePath,
  text,
  rules: { [rule]: { options: [] } },
  collectFixes: false,
  suggestionsMode: 'off',
  configKey: FIXTURES,
});

// Each scenario drives the real dist WorkerPool through one forced-terminate
// flow and reports `check(...)` assertions + an emitted `terminate-point`
// milestone just before the terminate that can native-abort on Windows.
const scenarios = {
  // U11: a refed top-level setInterval keeps the worker event loop alive, so
  // shutdown must escalate to a forced terminate().
  u11: async () => {
    const WorkerPool = await loadWorkerPool();
    const dir = makeFixtureDir({
      'plugin.mjs':
        'const _i = setInterval(() => {}, 60_000);\n' +
        "export default { meta: { name: 'u11' }, rules: { noop: { meta: {}, create() { return {}; } } } };\n",
      'config.mjs':
        "import plugin from './plugin.mjs';\nexport default [{ plugins: { u11: plugin } }];\n",
    });
    const pool = new WorkerPool({
      configs: [
        { configPath: path.join(dir, 'config.mjs'), configDirectory: dir },
      ],
      workerCount: 1,
    });
    await pool.init();
    milestone('init-done');
    milestone('terminate-point');
    const start = Date.now();
    await pool.shutdown();
    const elapsed = Date.now() - start;
    check('shutdown-bounded', elapsed < 8_000, `elapsed=${elapsed}ms`);
    check(
      'pool-drained',
      Array.isArray(pool.workers) && pool.workers.length === 0,
      `workers=${pool.workers?.length}`,
    );
  },

  // sync-wedged: the hang plugin spins forever on Program, so the worker can't
  // even process the inbound shutdown message — shutdown must escalate to a
  // forced terminate() after the 5s grace.
  'hang-shutdown': async () => {
    const WorkerPool = await loadWorkerPool();
    const pool = new WorkerPool({
      configs: [HANG_CONFIG],
      workerCount: 1,
      // 60s so a task_timeout respawn can't preempt the shutdown path.
      taskTimeoutMs: 60_000,
    });
    await pool.init();
    milestone('init-done');
    const wedgeP = pool.lintBatch([
      localTask('wedge.ts', 'const x = 1;\n', 'hang/hang'),
    ]);
    await sleep(200); // let the worker enter the sync hang before shutdown
    milestone('terminate-point');
    const start = Date.now();
    await pool.shutdown();
    const elapsed = Date.now() - start;
    check(
      'shutdown-bounded',
      elapsed >= 4_500 && elapsed < 8_000,
      `elapsed=${elapsed}ms`,
    );
    const wedge = await wedgeP;
    check(
      'wedge-shutdown',
      wedge.length === 1 && wedge[0]?.parseError === 'shutdown',
      `len=${wedge.length} parseError=${wedge[0]?.parseError}`,
    );
    check(
      'pool-drained',
      pool.workers.length === 0,
      `workers=${pool.workers.length}`,
    );
    const s2 = Date.now();
    await pool.shutdown(); // idempotent + instant
    check(
      'second-shutdown-fast',
      Date.now() - s2 < 50,
      `2nd=${Date.now() - s2}ms`,
    );
  },

  // worker-exit-race: directly terminate one worker of a 2-worker pool, then
  // shutdown. The leaked respawn must not survive; lintBatch after shutdown
  // rejects /closed/.
  'worker-exit-race': async () => {
    const WorkerPool = await loadWorkerPool();
    const pool = new WorkerPool({
      configs: [LOCAL_CONFIG],
      workerCount: 2,
      taskTimeoutMs: 5_000,
    });
    await pool.init();
    milestone('init-done');
    check(
      'workers-spawned',
      pool.workers.length > 0,
      `workers=${pool.workers.length}`,
    );
    milestone('terminate-point');
    void pool.workers[0].worker.terminate();
    const start = Date.now();
    await pool.shutdown();
    check(
      'shutdown-bounded',
      Date.now() - start < 10_000,
      `elapsed=${Date.now() - start}ms`,
    );
    let rejectedClosed = false;
    try {
      await pool.lintBatch([localTask('x.ts', 'const x = 1;')]);
    } catch (e) {
      rejectedClosed = /closed/.test(String(e?.message ?? e));
    }
    check('lint-batch-rejects-closed', rejectedClosed, '');
    await sleep(200);
    check(
      'pool-drained',
      pool.workers.length === 0,
      `workers=${pool.workers.length}`,
    );
  },

  // task-timeout: the hang plugin trips the 600ms per-task timeout, which
  // terminates the worker and respawns it; the next batch must succeed on the
  // replacement.
  'task-timeout': async () => {
    const WorkerPool = await loadWorkerPool();
    const logs = [];
    const pool = new WorkerPool({
      configs: [HANG_CONFIG],
      workerCount: 1,
      taskTimeoutMs: 600,
      onLog: (rec) => logs.push(rec),
    });
    await pool.init();
    milestone('init-done');
    milestone('terminate-point'); // the timeout-driven terminate fires below
    const hangResult = await pool.lintBatch([
      localTask('wedge.ts', 'const x = 1;\n', 'hang/hang'),
    ]);
    check(
      'hang-task-timeout',
      hangResult.length === 1 && hangResult[0]?.parseError === 'task_timeout',
      `len=${hangResult.length} parseError=${hangResult[0]?.parseError}`,
    );
    await sleep(500);
    const okResult = await pool.lintBatch([
      localTask('ok.ts', 'const TRIGGER = 1;\n', 'hang/noop'),
    ]);
    check(
      'recovery-ok',
      okResult.length === 1 &&
        okResult[0]?.parseError === undefined &&
        okResult[0]?.diagnostics?.length === 1 &&
        okResult[0]?.diagnostics[0]?.message === 'noop fired',
      JSON.stringify(okResult[0]),
    );
    check(
      'respawn-logged',
      logs.some((l) => l.text.includes('respawning')),
      '',
    );
    await pool.shutdown();
  },

  // all-degraded: drive every slot past its respawn cap (crashCount=cap +
  // terminate), so the in-flight batch drains as parseError:pool_degraded.
  'all-degraded': async () => {
    const WorkerPool = await loadWorkerPool();
    const pool = new WorkerPool({
      configs: [LOCAL_CONFIG],
      workerCount: 1,
      retryCap: 1,
    });
    await pool.init();
    milestone('init-done');
    pool.workers[0].ready = false; // hold non-ready so kickQueue can't dispatch
    const batchP = pool.lintBatch(
      [1, 2].map((i) => localTask(`q${i}.ts`, 'const x = null;\n')),
    );
    await sleep(30);
    milestone('terminate-point');
    pool.workers[0].crashCount = pool.opts.retryCap;
    await pool.workers[0].worker.terminate();
    const result = await batchP;
    check(
      'degraded-drain',
      result.length === 2 &&
        result.every(
          (r) =>
            r.parseError === 'pool_degraded' &&
            r.cancelled === false &&
            Array.isArray(r.diagnostics) &&
            r.diagnostics.length === 0,
        ),
      JSON.stringify(result.map((r) => r.parseError)),
    );
    check(
      'queue-drained',
      pool.pendingQueue.length === 0,
      `pending=${pool.pendingQueue.length}`,
    );
    await pool.shutdown();
  },

  // lint-batch-after-degraded: like all-degraded, but a SECOND batch issued
  // AFTER the pool settled terminal must also resolve pool_degraded (not hang).
  'lint-batch-after-degraded': async () => {
    const WorkerPool = await loadWorkerPool();
    const pool = new WorkerPool({
      configs: [LOCAL_CONFIG],
      workerCount: 1,
      retryCap: 1,
    });
    await pool.init();
    milestone('init-done');
    pool.workers[0].ready = false;
    const firstBatch = pool.lintBatch([
      localTask('first.ts', 'const x = null;\n'),
    ]);
    await sleep(30);
    milestone('terminate-point');
    pool.workers[0].crashCount = pool.opts.retryCap;
    await pool.workers[0].worker.terminate();
    const firstResult = await firstBatch;
    check(
      'first-degraded',
      firstResult.length === 1 && firstResult[0].parseError === 'pool_degraded',
      JSON.stringify(firstResult.map((r) => r.parseError)),
    );
    check(
      'terminal-state',
      pool.closed === false &&
        pool.workers.length === 1 &&
        pool.workers[0].ready === false &&
        pool.workers[0].respawning === false &&
        pool.workers[0].exited === true,
      `closed=${pool.closed} ready=${pool.workers[0]?.ready} exited=${pool.workers[0]?.exited}`,
    );
    const secondBatch = pool.lintBatch([
      localTask('second-a.ts', 'const y = null;\n'),
      localTask('second-b.ts', 'const z = null;\n'),
    ]);
    const guard = new Promise((_, rej) =>
      setTimeout(() => rej(new Error('second lintBatch hung')), 4_000),
    );
    const secondResult = await Promise.race([secondBatch, guard]);
    check(
      'second-degraded',
      secondResult.length === 2 &&
        secondResult.every(
          (r) =>
            r.parseError === 'pool_degraded' &&
            r.cancelled === false &&
            Array.isArray(r.diagnostics) &&
            r.diagnostics.length === 0,
        ),
      JSON.stringify(secondResult.map((r) => r.parseError)),
    );
    check(
      'queue-drained',
      pool.pendingQueue.length === 0,
      `pending=${pool.pendingQueue.length}`,
    );
    await pool.shutdown();
  },
};

const name = process.argv[2];
const fn = scenarios[name];
if (typeof fn !== 'function') {
  report({ kind: 'error', detail: `unknown scenario: ${name}` });
  process.exit(2);
}
try {
  await fn();
  milestone('done'); // explicit success — the last line on the happy path
  process.exit(0);
} catch (err) {
  // An ORDERLY error (import/init threw) — report it so the parent fails the
  // test, rather than mistaking a silent exit for a tolerable native abort.
  report({ kind: 'error', detail: String(err?.stack ?? err) });
  process.exit(1);
}

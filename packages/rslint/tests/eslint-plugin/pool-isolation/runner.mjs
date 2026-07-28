// Subprocess entry for process-isolated worker-pool scenarios.
//
// Forced termination of a worker containing the native parser can abort below
// JavaScript on Windows. This process confines that abort. The parent accepts
// it only when the last durable journal record is emitted by a wrapper around
// the real Worker.prototype.terminate call.
import fs from 'node:fs';
import path from 'node:path';
import { Worker } from 'node:worker_threads';
import { pathToFileURL, fileURLToPath } from 'node:url';

const JOURNAL_VERSION = 1;
const HERE = path.dirname(fileURLToPath(import.meta.url));
const FIXTURES = path.resolve(HERE, '../fixtures');
const LOCAL_CONFIG = {
  configPath: path.join(FIXTURES, 'local.config.mjs'),
  configDirectory: FIXTURES,
};
const HANG_CONFIG = {
  configPath: path.join(FIXTURES, 'hang.config.mjs'),
  configDirectory: FIXTURES,
};

const SCENARIO = process.argv[2] ?? '';
const MILESTONE_FILE = process.env.RSLINT_MILESTONE_FILE;
const RUN_ID = process.env.RSLINT_RUN_ID;
const SCENARIO_TMP_DIR = process.env.RSLINT_SCENARIO_TMP_DIR;
const JOURNAL_RECORD_LIMIT = 1_024;
const NON_FIRING_TASK_TIMEOUT_MS = 24 * 60 * 60 * 1_000;
const FIXTURE_KEEPALIVE_INTERVAL_MS = 60_000;
let sequence = 0;
let terminalRecordWritten = false;

const report = (record) => {
  if (!MILESTONE_FILE || !RUN_ID) {
    throw new Error('journal environment is incomplete');
  }
  if (sequence >= JOURNAL_RECORD_LIMIT) {
    throw new Error(
      `journal exceeded ${JOURNAL_RECORD_LIMIT} record safety limit`,
    );
  }
  sequence++;
  fs.appendFileSync(
    MILESTONE_FILE,
    `${JSON.stringify({
      ...record,
      version: JOURNAL_VERSION,
      runId: RUN_ID,
      scenario: SCENARIO,
      seq: sequence,
    })}\n`,
  );
};
const milestone = (name) => report({ kind: 'milestone', name });
const check = (name, pass, detail) =>
  report({
    kind: 'assert',
    name,
    pass: Boolean(pass),
    ...(detail === undefined ? {} : { detail }),
  });
const terminal = (record) => {
  report(record);
  terminalRecordWritten = true;
};

async function loadWorkerPool() {
  const dist = process.env.RSLINT_DIST_ESLINT_PLUGIN;
  if (!dist) throw new Error('RSLINT_DIST_ESLINT_PLUGIN not set');
  // Production resolution path: WorkerPool resolves its sibling
  // dist/eslint-plugin/lint-worker.js automatically. pathToFileURL is required
  // on Windows because ESM import rejects bare `C:\...` paths.
  const mod = await import(pathToFileURL(dist).href);
  return mod.WorkerPool;
}

let fixtureIndex = 0;
const makeFixtureDir = (files) => {
  if (!SCENARIO_TMP_DIR) {
    throw new Error('RSLINT_SCENARIO_TMP_DIR not set');
  }
  const dir = path.join(SCENARIO_TMP_DIR, `fixture-${++fixtureIndex}`);
  fs.mkdirSync(dir);
  for (const [name, content] of Object.entries(files)) {
    if (path.basename(name) !== name) {
      throw new Error(`fixture name must not escape its directory: ${name}`);
    }
    fs.writeFileSync(path.join(dir, name), content);
  }
  return dir;
};

const markerPath = (name) => {
  if (!SCENARIO_TMP_DIR) {
    throw new Error('RSLINT_SCENARIO_TMP_DIR not set');
  }
  return path.join(SCENARIO_TMP_DIR, name);
};

const delay = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

// Polling is used only to observe an explicit cross-thread file barrier.
// The parent harness owns the one deadlock watchdog for the whole scenario;
// a shorter child deadline would turn Windows scheduler contention into a
// second, non-product failure boundary.
async function waitForFile(file, description) {
  while (!fs.existsSync(file)) {
    await delay(10);
  }
  void description;
}

// Observe the exact JS entry to Worker.terminate without changing its returned
// promise, arguments, or production implementation. If native teardown aborts,
// this is the last record the parent is allowed to accept.
function observeWorkerTerminate() {
  const original = Worker.prototype.terminate;
  const wrapped = function () {
    milestone('terminate-invoked');
    return original.call(this);
  };
  Worker.prototype.terminate = wrapped;
  return () => {
    if (Worker.prototype.terminate === wrapped) {
      Worker.prototype.terminate = original;
    }
  };
}

// Capture one production timeout callback without waiting for wall-clock time.
// The WorkerPool schedules it synchronously while lintBatch dispatches. The
// returned fake handle is only ever passed to clearTimeout.
function captureTimeout(delayMs, action) {
  const original = globalThis.setTimeout;
  const callbacks = [];
  globalThis.setTimeout = function (callback, delayValue, ...args) {
    if (delayValue === delayMs) {
      callbacks.push(() => callback(...args));
      return Object.create(null);
    }
    return original(callback, delayValue, ...args);
  };
  let value;
  try {
    value = action();
  } finally {
    globalThis.setTimeout = original;
  }
  return {
    value,
    count: callbacks.length,
    fire() {
      if (callbacks.length !== 1) {
        throw new Error(
          `expected exactly one ${delayMs}ms timeout, got ${callbacks.length}`,
        );
      }
      callbacks[0]();
    },
  };
}

const localTask = (filePath, text, rule = 'local/no-null') => ({
  filePath,
  text,
  rules: { [rule]: { options: [] } },
  collectFixes: false,
  suggestionsMode: 'off',
  configKey: FIXTURES,
});

const scenarios = {
  // A refed top-level interval keeps the worker event loop alive. Shutdown must
  // use its existing grace fallback and leave no live worker/handle behind.
  u11: async () => {
    const WorkerPool = await loadWorkerPool();
    const loadedFile = markerPath('u11-plugin-loaded');
    const dir = makeFixtureDir({
      'plugin.mjs':
        "import fs from 'node:fs';\n" +
        `fs.writeFileSync(${JSON.stringify(loadedFile)}, 'loaded');\n` +
        `const _interval = setInterval(() => {}, ${FIXTURE_KEEPALIVE_INTERVAL_MS});\n` +
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
    await waitForFile(loadedFile, 'refed plugin load barrier');
    milestone('refed-plugin-loaded');
    await pool.shutdown();
    check(
      'pool-drained',
      Array.isArray(pool.workers) && pool.workers.length === 0,
      `workers=${pool.workers?.length}`,
    );
  },

  // The listener proves it entered its synchronous wedge via a file barrier.
  // Shutdown must then force terminate and settle the in-flight task.
  'hang-shutdown': async () => {
    const WorkerPool = await loadWorkerPool();
    const enteredFile = markerPath('hang-shutdown-entered');
    process.env.RSLINT_HANG_ENTERED_FILE = enteredFile;
    const pool = new WorkerPool({
      configs: [HANG_CONFIG],
      workerCount: 1,
      taskTimeoutMs: NON_FIRING_TASK_TIMEOUT_MS,
    });
    await pool.init();
    milestone('init-done');
    const wedgeP = pool.lintBatch([
      localTask('wedge.ts', 'const x = 1;\n', 'hang/hang'),
    ]);
    await waitForFile(enteredFile, 'sync listener entry');
    milestone('hang-entered');
    await pool.shutdown();
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
    await pool.shutdown();
    check('second-shutdown-returned', pool.closed === true, '');
  },

  // Hold a replacement worker inside config import, start shutdown while that
  // respawn promise is provably in flight, then release it. Shutdown must await
  // and terminate the replacement rather than leaking it.
  'worker-exit-race': async () => {
    const WorkerPool = await loadWorkerPool();
    const countFile = markerPath('respawn-import-count');
    const enteredFile = markerPath('respawn-import-entered');
    const releaseFile = markerPath('respawn-import-release');
    const dir = makeFixtureDir({
      'config.mjs':
        "import fs from 'node:fs';\n" +
        `const countFile = ${JSON.stringify(countFile)};\n` +
        `const enteredFile = ${JSON.stringify(enteredFile)};\n` +
        `const releaseFile = ${JSON.stringify(releaseFile)};\n` +
        "const count = fs.existsSync(countFile) ? Number(fs.readFileSync(countFile, 'utf8')) + 1 : 1;\n" +
        'fs.writeFileSync(countFile, String(count));\n' +
        'if (count > 1) {\n' +
        "  fs.writeFileSync(enteredFile, 'entered');\n" +
        '  while (!fs.existsSync(releaseFile)) {\n' +
        '    await new Promise((resolve) => setTimeout(resolve, 10));\n' +
        '  }\n' +
        '}\n' +
        "const plugin = { meta: { name: 'respawn-race' }, rules: {\n" +
        '  noop: { meta: {}, create() { return {}; } },\n' +
        '  crash: { meta: {}, create() { return { Program() { process.exit(42); } }; } },\n' +
        '} };\n' +
        'export default [{ plugins: { race: plugin } }];\n',
    });
    const pool = new WorkerPool({
      configs: [
        { configPath: path.join(dir, 'config.mjs'), configDirectory: dir },
      ],
      workerCount: 1,
    });
    await pool.init();
    milestone('init-done');
    milestone('initial-worker-ready');
    check(
      'one-worker-spawned',
      pool.workers.length === 1 && pool.workers[0].ready === true,
      `workers=${pool.workers.length} ready=${pool.workers[0]?.ready}`,
    );

    const crashResult = await pool.lintBatch([
      {
        ...localTask('crash.ts', 'const x = 1;\n', 'race/crash'),
        configKey: dir,
      },
    ]);
    check(
      'worker-hard-exit',
      crashResult.length === 1 &&
        crashResult[0]?.parseError?.includes('worker_crashed'),
      JSON.stringify(crashResult[0]),
    );
    await waitForFile(enteredFile, 'replacement worker import barrier');
    milestone('respawn-import-entered');
    check(
      'respawn-in-flight',
      pool.workers[0].respawning === true && pool.respawns.size === 1,
      `respawning=${pool.workers[0].respawning} respawns=${pool.respawns.size}`,
    );
    milestone('respawn-in-flight');

    const shutdownP = pool.shutdown();
    check('closed-before-respawn-release', pool.closed === true, '');
    milestone('shutdown-started');
    fs.writeFileSync(releaseFile, 'release');
    await shutdownP;
    check(
      'pool-drained',
      pool.workers.length === 0 && pool.respawns.size === 0,
      `workers=${pool.workers.length} respawns=${pool.respawns.size}`,
    );

    let rejectedClosed = false;
    try {
      await pool.lintBatch([localTask('x.ts', 'const x = 1;')]);
    } catch (err) {
      rejectedClosed = /closed/.test(String(err?.message ?? err));
    }
    check('lint-batch-rejects-closed', rejectedClosed, '');
  },

  // Capture and fire the exact timeout callback created by dispatch. This pins
  // timeout → terminate → respawn → recovery without a scheduler-dependent
  // sleep or a small wall-clock budget.
  'task-timeout': async () => {
    const WorkerPool = await loadWorkerPool();
    const enteredFile = markerPath('task-timeout-hang-entered');
    process.env.RSLINT_HANG_ENTERED_FILE = enteredFile;
    const logs = [];
    const capturedDelay = NON_FIRING_TASK_TIMEOUT_MS;
    const pool = new WorkerPool({
      configs: [HANG_CONFIG],
      workerCount: 1,
      taskTimeoutMs: capturedDelay,
      onLog: (record) => logs.push(record),
    });
    await pool.init();
    milestone('init-done');

    const captured = captureTimeout(capturedDelay, () =>
      pool.lintBatch([localTask('wedge.ts', 'const x = 1;\n', 'hang/hang')]),
    );
    check(
      'one-task-timeout-captured',
      captured.count === 1,
      `captured=${captured.count}`,
    );
    await waitForFile(enteredFile, 'task timeout listener entry');
    milestone('hang-entered');
    captured.fire();

    const hangResult = await captured.value;
    check(
      'hang-task-timeout',
      hangResult.length === 1 && hangResult[0]?.parseError === 'task_timeout',
      `len=${hangResult.length} parseError=${hangResult[0]?.parseError}`,
    );
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
      logs.some((record) => record.text.includes('respawning')),
      '',
    );
    await pool.shutdown();
    check('pool-drained', pool.workers.length === 0, '');
  },

  // Enqueue is synchronous. Force the replacement spawn to reject and prove
  // that terminal path drains the queue without a scheduler delay. The
  // lint-batch-after-degraded scenario separately covers retry-cap exhaustion.
  'all-degraded': async () => {
    const WorkerPool = await loadWorkerPool();
    const pool = new WorkerPool({
      configs: [LOCAL_CONFIG],
      workerCount: 1,
      retryCap: 1,
    });
    await pool.init();
    milestone('init-done');
    pool.workers[0].ready = false;
    let respawnAttempts = 0;
    pool.spawnWorker = async () => {
      respawnAttempts++;
      throw new Error('injected respawn rejection');
    };
    const batchP = pool.lintBatch(
      [1, 2].map((index) => localTask(`q${index}.ts`, 'const x = null;\n')),
    );
    check(
      'queue-enqueued',
      pool.pendingQueue.length === 2,
      `pending=${pool.pendingQueue.length}`,
    );
    milestone('degraded-queue-ready');
    await pool.workers[0].worker.terminate();
    const result = await batchP;
    check(
      'respawn-rejected',
      respawnAttempts === 1 &&
        pool.workers[0].ready === false &&
        pool.workers[0].respawning === false,
      `attempts=${respawnAttempts} ready=${pool.workers[0]?.ready} respawning=${pool.workers[0]?.respawning}`,
    );
    check(
      'degraded-drain',
      result.length === 2 &&
        result.every(
          (item) =>
            item.parseError === 'pool_degraded' &&
            item.cancelled === false &&
            Array.isArray(item.diagnostics) &&
            item.diagnostics.length === 0,
        ),
      JSON.stringify(result.map((item) => item.parseError)),
    );
    check(
      'queue-drained',
      pool.pendingQueue.length === 0,
      `pending=${pool.pendingQueue.length}`,
    );
    await pool.shutdown();
  },

  // After the first queue proves terminal degradation, a future batch must be
  // drained synchronously by lintBatch's terminal-state guard.
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
    check(
      'first-queue-enqueued',
      pool.pendingQueue.length === 1,
      `pending=${pool.pendingQueue.length}`,
    );
    milestone('degraded-queue-ready');
    pool.workers[0].crashCount = pool.opts.retryCap;
    await pool.workers[0].worker.terminate();
    const firstResult = await firstBatch;
    check(
      'first-degraded',
      firstResult.length === 1 && firstResult[0].parseError === 'pool_degraded',
      JSON.stringify(firstResult.map((item) => item.parseError)),
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
    const secondResult = await pool.lintBatch([
      localTask('second-a.ts', 'const y = null;\n'),
      localTask('second-b.ts', 'const z = null;\n'),
    ]);
    check(
      'second-degraded',
      secondResult.length === 2 &&
        secondResult.every(
          (item) =>
            item.parseError === 'pool_degraded' &&
            item.cancelled === false &&
            Array.isArray(item.diagnostics) &&
            item.diagnostics.length === 0,
        ),
      JSON.stringify(secondResult.map((item) => item.parseError)),
    );
    check(
      'queue-drained',
      pool.pendingQueue.length === 0,
      `pending=${pool.pendingQueue.length}`,
    );
    await pool.shutdown();
  },
};

if (
  !MILESTONE_FILE ||
  !RUN_ID ||
  !SCENARIO_TMP_DIR ||
  !process.env.RSLINT_DIST_ESLINT_PLUGIN
) {
  console.error('isolated runner environment is incomplete');
  process.exitCode = 2;
} else {
  // `beforeExit` is skipped by process.exit(), while the synchronous `exit`
  // event still runs. Record that orderly early-exit path so an exit code 0
  // can never masquerade as the known Windows native abort (which bypasses JS
  // teardown entirely). Natural success writes `done` first; reported errors
  // also mark the journal terminal and therefore do not get duplicated here.
  process.once('exit', (code) => {
    if (terminalRecordWritten) return;
    try {
      terminal({
        kind: 'error',
        detail: `runner exited through Node before a terminal record (code=${code})`,
      });
    } catch (err) {
      console.error(`could not journal early process exit: ${String(err)}`);
    }
  });
  milestone('started');
  const scenario = scenarios[SCENARIO];
  if (typeof scenario !== 'function') {
    terminal({ kind: 'error', detail: `unknown scenario: ${SCENARIO}` });
    process.exitCode = 2;
  } else {
    const restoreTerminate = observeWorkerTerminate();
    try {
      await scenario();
      process.exitCode = 0;
      // Write `done` only when Node itself proves the event loop is quiescent.
      // A leaked Worker/refed handle therefore trips the parent watchdog
      // instead of being hidden by an eager journal record or process.exit(0).
      process.once('beforeExit', () => {
        restoreTerminate();
        terminal({ kind: 'milestone', name: 'done' });
      });
    } catch (err) {
      restoreTerminate();
      try {
        terminal({ kind: 'error', detail: String(err?.stack ?? err) });
      } catch (reportError) {
        console.error(
          `scenario failed and its journal could not be updated: ${String(
            reportError,
          )}\noriginal error: ${String(err?.stack ?? err)}`,
        );
      }
      process.exitCode = 1;
    }
  }
}

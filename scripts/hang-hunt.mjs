// hang-hunt.mjs — drive the intermittent windows-latest CLI hang to the
// surface in ONE CI run, and capture WHY when it fires.
//
// It loops many concurrent `@rslint/core` CLI invocations over a tiny
// type-aware fixture (the exact shape of the test that hangs in CI:
// disable-comments.test.ts), with the Go-side hang watchdog + phase trace
// enabled (see cmd/rslint/hangdiag.go). When an invocation wedges, the Go
// watchdog dumps every goroutine stack to stderr and exits with code 99; this
// script detects that, prints the dump, copies any dump file to the artifact
// dir, and exits non-zero so the job goes red WITH the deadlock stack attached.
//
// Env knobs (with CI-oriented defaults):
//   HANG_CONCURRENCY   parallel invocations per round            (default 8)
//   HANG_ROUNDS        number of rounds                          (default 400)
//   HANG_NFILES        .ts files per invocation (1 == exact      (default 24)
//                      failing-test shape; >1 exercises the
//                      per-file shard parallelism from #1088)
//   HANG_WATCHDOG_MS   per-invocation Go watchdog deadline       (default 45000)
//   HANG_DUMP_DIR      artifact dir for goroutine dump files     (default ./hang-dumps)
//   HANG_STOP_ON_FIRST stop+fail on the first hang (1) or run    (default 1)
//                      all rounds collecting every hang (0)
import { spawn } from 'node:child_process';
import fs from 'node:fs/promises';
import path from 'node:path';
import os from 'node:os';
import { fileURLToPath } from 'node:url';

const HERE = path.dirname(fileURLToPath(import.meta.url));
const CJS = path.resolve(HERE, '../packages/rslint/bin/rslint.cjs');

const CONCURRENCY = Number(process.env.HANG_CONCURRENCY ?? 8);
const ROUNDS = Number(process.env.HANG_ROUNDS ?? 400);
const NFILES = Number(process.env.HANG_NFILES ?? 24);
const WATCHDOG_MS = Number(process.env.HANG_WATCHDOG_MS ?? 45_000);
const DUMP_DIR = path.resolve(process.env.HANG_DUMP_DIR ?? './hang-dumps');
const STOP_ON_FIRST = (process.env.HANG_STOP_ON_FIRST ?? '1') !== '0';
// Wall-clock budget so the job ends cleanly (green "no repro" / red "with dump")
// before the workflow's hard timeout-minutes cancels it mid-round.
const MAX_MINUTES = Number(process.env.HANG_MAX_MINUTES ?? 70);
const DEADLINE = Date.now() + MAX_MINUTES * 60_000;
// Hard ceiling per invocation so a wedge that even the Go watchdog can't escape
// (e.g. a Node-side wedge) still can't stall the runner indefinitely.
const HARD_TIMEOUT_MS = WATCHDOG_MS + 30_000;
const WATCHDOG_EXIT = 99;

const cfg = `export default [${JSON.stringify({
  files: ['**/*.ts'],
  languageOptions: { parserOptions: { projectService: false, project: ['./tsconfig.json'] } },
  rules: { '@typescript-eslint/no-explicit-any': 'error' },
  plugins: ['@typescript-eslint'],
})}];`;
const TSCONFIG = JSON.stringify({
  compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
  include: ['**/*.ts'],
});

async function createTempDir() {
  const dir = await fs.mkdtemp(path.join(os.tmpdir(), 'rslint-hang-'));
  await fs.writeFile(path.join(dir, 'rslint.config.mjs'), cfg);
  await fs.writeFile(path.join(dir, 'tsconfig.json'), TSCONFIG);
  for (let i = 0; i < NFILES; i++) {
    const imp = i > 0 ? `import { v${i - 1} } from './f${i - 1}';\n` : '';
    await fs.writeFile(
      path.join(dir, `f${i}.ts`),
      `${imp}export const v${i}: any = ${i};\nconst local${i}: any = '${i}';\nexport { local${i} };\n`,
    );
  }
  return dir;
}

function runOnce(dir, id) {
  return new Promise((resolve) => {
    const child = spawn(process.execPath, [CJS], {
      cwd: dir,
      stdio: ['pipe', 'pipe', 'pipe'],
      env: {
        ...process.env,
        RSLINT_HANG_WATCHDOG_MS: String(WATCHDOG_MS),
        RSLINT_HANG_TRACE: '1',
        RSLINT_HANG_DUMP_DIR: DUMP_DIR,
        GOTRACEBACK: 'all',
      },
    });
    let stderr = '';
    let stdout = '';
    child.stderr.on('data', (d) => (stderr += d));
    child.stdout.on('data', (d) => (stdout += d));
    let settled = false;
    const finish = (status, exitCode) => {
      if (settled) return;
      settled = true;
      clearTimeout(timer);
      resolve({ status, exitCode, stderr, stdout, id });
    };
    const timer = setTimeout(() => {
      try { child.kill('SIGKILL'); } catch {}
      finish('hard-timeout', null);
    }, HARD_TIMEOUT_MS);
    child.on('close', (code) => {
      if (code === WATCHDOG_EXIT) finish('watchdog-hang', code);
      else finish('ok', code);
    });
  });
}

const hangs = [];
let total = 0;

await fs.mkdir(DUMP_DIR, { recursive: true });
console.log(
  `hang-hunt: concurrency=${CONCURRENCY} rounds=${ROUNDS} nfiles=${NFILES} ` +
    `watchdog=${WATCHDOG_MS}ms bin=${CJS}`,
);

outer: for (let r = 1; r <= ROUNDS; r++) {
  if (Date.now() > DEADLINE) {
    console.log(`time budget (${MAX_MINUTES}min) reached at round ${r}; stopping.`);
    break;
  }
  const dirs = await Promise.all(
    Array.from({ length: CONCURRENCY }, () => createTempDir()),
  );
  const results = await Promise.all(dirs.map((d, i) => runOnce(d, `r${r}c${i}`)));
  for (const res of results) {
    total++;
    if (res.status !== 'ok') {
      hangs.push(res);
      console.log(`\n################ HANG #${hangs.length} (id=${res.id}, ${res.status}) ################`);
      console.log('--- captured child stderr (Go phase trace + goroutine dump) ---');
      console.log(res.stderr || '(empty)');
      console.log('--- end captured stderr ---');
      await fs.writeFile(
        path.join(DUMP_DIR, `hang-${res.id}-stderr.txt`),
        res.stderr,
      );
      if (STOP_ON_FIRST) break outer;
    }
  }
  for (const d of dirs) fs.rm(d, { recursive: true, force: true }).catch(() => {});
  if (r % 10 === 0 || hangs.length) {
    console.log(`round ${r}/${ROUNDS}: total=${total} hangs=${hangs.length}`);
  }
}

console.log(`\nhang-hunt FINAL: total=${total} hangs=${hangs.length}`);
if (hangs.length) {
  console.log(`reproduced the hang ${hangs.length}×. Dumps in ${DUMP_DIR}.`);
  process.exit(1);
}
console.log('no hang reproduced in this run.');

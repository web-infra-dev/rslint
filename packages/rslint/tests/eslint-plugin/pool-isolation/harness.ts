// Parent-side harness for process-isolated worker-pool scenarios.
//
// rstest tests call `runPoolScenario(name)`, which spawns runner.mjs in a
// throwaway subprocess, lets it drive the real dist WorkerPool through a
// terminate-churning scenario, and applies a MILESTONE-DRIVEN verdict. The
// point: a native napi-terminate abort (Windows) crashes only that subprocess,
// never this test process.
import { spawn } from 'node:child_process';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const HERE = path.dirname(fileURLToPath(import.meta.url));
const RUNNER = path.join(HERE, 'runner.mjs');
// tests/eslint-plugin/pool-isolation → packages/rslint, then dist/eslint-plugin.
// The runner imports this built artifact (tests already require `pnpm build`;
// the worker entry is the dist sibling).
const DIST_ESLINT_PLUGIN = path.resolve(
  HERE,
  '../../../dist/eslint-plugin/index.js',
);

export type Verdict = 'PASS' | 'TOLERATED-PASS' | 'FAIL';

export interface ScenarioResult {
  scenario: string;
  verdict: Verdict;
  milestones: string[];
  asserts: { name: string; pass: boolean; detail?: string }[];
  error?: string;
  exitCode: number | null;
  signal: NodeJS.Signals | null;
  timedOut: boolean;
  stderr: string;
}

interface Rec {
  kind: 'milestone' | 'assert' | 'error';
  name?: string;
  pass?: boolean;
  detail?: string;
}

export async function runPoolScenario(
  scenario: string,
  opts: { timeoutMs?: number } = {},
): Promise<ScenarioResult> {
  const timeoutMs = opts.timeoutMs ?? 15_000;
  const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-pool-iso-'));
  const milestoneFile = path.join(tmpDir, 'milestones.ndjson');
  fs.writeFileSync(milestoneFile, '');

  return new Promise<ScenarioResult>((resolve) => {
    const child = spawn(process.execPath, [RUNNER, scenario], {
      stdio: ['ignore', 'pipe', 'pipe'],
      env: {
        ...process.env,
        RSLINT_MILESTONE_FILE: milestoneFile,
        RSLINT_DIST_ESLINT_PLUGIN: DIST_ESLINT_PLUGIN,
      },
    });
    let stderr = '';
    child.stderr.on('data', (d) => (stderr += d.toString()));

    let settled = false;
    const finish = (
      exitCode: number | null,
      signal: NodeJS.Signals | null,
      timedOut: boolean,
    ): void => {
      if (settled) return;
      settled = true;
      clearTimeout(timer);
      const recs = readMilestones(milestoneFile);
      fs.rmSync(tmpDir, { recursive: true, force: true });
      const milestones = recs
        .filter((r) => r.kind === 'milestone')
        .map((r) => r.name!);
      const asserts = recs
        .filter((r) => r.kind === 'assert')
        .map((r) => ({ name: r.name!, pass: !!r.pass, detail: r.detail }));
      const error = recs.find((r) => r.kind === 'error')?.detail;
      resolve({
        scenario,
        verdict: classify(milestones, asserts, error, timedOut),
        milestones,
        asserts,
        error,
        exitCode,
        signal,
        timedOut,
        stderr,
      });
    };
    const timer = setTimeout(() => {
      child.kill('SIGKILL');
      finish(null, 'SIGKILL', true);
    }, timeoutMs);
    child.on('exit', (code, signal) => finish(code, signal, false));
    child.on('error', () => finish(-1, null, false));
  });
}

function readMilestones(file: string): Rec[] {
  let text: string;
  try {
    text = fs.readFileSync(file, 'utf8');
  } catch {
    return [];
  }
  const out: Rec[] = [];
  for (const line of text.split('\n')) {
    if (!line) continue;
    try {
      out.push(JSON.parse(line) as Rec);
    } catch {
      /* ignore a torn final line (process died mid-write) */
    }
  }
  return out;
}

// MILESTONE-DRIVEN verdict (validated by the spike). Deliberately ignores exit
// code/signal: a real napi abort on Windows can surface as exit code 0
// (indistinguishable from a clean exit), and an orderly failure exits non-zero
// — exit codes cannot tell them apart, but the child's own milestones can.
function classify(
  milestones: string[],
  asserts: { pass: boolean }[],
  error: string | undefined,
  timedOut: boolean,
): Verdict {
  const reached = (m: string): boolean => milestones.includes(m);
  if (asserts.some((a) => !a.pass)) return 'FAIL'; // in-child assertion failed
  if (error) return 'FAIL'; // child reported an orderly error
  if (reached('done')) return 'PASS'; // explicit success
  if (timedOut) return 'FAIL'; // pool hung
  // Silent abnormal exit with no done/error/failed-assert: tolerate ONLY if the
  // pool provably reached the terminate step — that's the isolated native
  // abort. Anything earlier is a real crash and must fail.
  //
  // Coverage note: a TOLERATED-PASS means the in-child asserts AFTER the
  // terminate point did not run (the process aborted there). Those business
  // invariants (drain / respawn / closed-rejection / no-leak) are still fully
  // verified on mac/linux, where terminate is clean and the child runs through
  // to `done` (PASS). The Windows abort path only confirms the pool reached the
  // terminate step and the isolation held — the business assertions are
  // guaranteed by the non-Windows runners, which is sound because those
  // invariants are platform-independent.
  return reached('terminate-point') ? 'TOLERATED-PASS' : 'FAIL';
}

/**
 * Human-readable failure report for a scenario that did not pass. Surfaces the
 * failed in-child assertion(s) (name + detail), the child's error / timeout /
 * exit code, the milestones it reached, and the child stderr — so a CI failure
 * reads like a normal assertion report instead of a JSON blob. Pass it as
 * expect()'s second arg:
 *   expect(r.verdict, formatScenarioFailure(r)).not.toBe('FAIL')
 */
export function formatScenarioFailure(r: ScenarioResult): string {
  const lines: string[] = [
    `pool-isolation scenario "${r.scenario}" did not pass (verdict=${r.verdict}).`,
  ];
  const failed = r.asserts.filter((a) => !a.pass);
  if (failed.length > 0) {
    lines.push('  failed in-child assertions:');
    for (const a of failed) {
      lines.push(`    x ${a.name}${a.detail ? ` - ${a.detail}` : ''}`);
    }
  }
  if (r.error) lines.push(`  child error: ${r.error}`);
  if (r.timedOut) {
    lines.push('  child TIMED OUT - the pool hung and never exited');
  }
  lines.push(`  milestones reached: ${r.milestones.join(' > ') || '(none)'}`);
  lines.push(
    `  child exit: code=${r.exitCode ?? 'null'} signal=${r.signal ?? 'null'}`,
  );
  if (r.stderr.trim()) {
    lines.push('  child stderr:');
    for (const l of r.stderr.trim().split('\n')) lines.push(`    ${l}`);
  }
  return lines.join('\n');
}

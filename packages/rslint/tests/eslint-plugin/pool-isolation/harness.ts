// Parent-side harness for process-isolated worker-pool scenarios.
//
// rstest tests call `runPoolScenario(name)`, which spawns runner.mjs in a
// throwaway subprocess, lets it drive the real dist WorkerPool through a
// terminate-churning scenario, and applies a milestone-driven verdict. A
// native napi-terminate abort (Windows) crashes only that subprocess, never
// the rstest process.
import { spawn } from 'node:child_process';
import { randomUUID } from 'node:crypto';
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

export const JOURNAL_VERSION = 1;
const STDERR_LIMIT_BYTES = 256 * 1024;
const JOURNAL_LIMIT_BYTES = 1024 * 1024;

interface ScenarioContract {
  /**
   * Ordered state prefix required both for a clean pass and for the narrow
   * Windows native-abort exception. The final item is always the exact
   * Worker.prototype.terminate entry observed by runner.mjs.
   */
  requiredProgress: readonly string[];
  requiredSuccessAsserts: readonly string[];
}

// One fail-closed contract table prevents allowlist, prerequisite, and
// assertion requirements from drifting apart when a scenario is added.
const SCENARIO_CONTRACTS = {
  u11: {
    requiredProgress: [
      'started',
      'init-done',
      'refed-plugin-loaded',
      'terminate-invoked',
    ],
    requiredSuccessAsserts: ['pool-drained'],
  },
  'hang-shutdown': {
    requiredProgress: [
      'started',
      'init-done',
      'hang-entered',
      'terminate-invoked',
    ],
    requiredSuccessAsserts: [
      'wedge-shutdown',
      'pool-drained',
      'second-shutdown-returned',
    ],
  },
  'worker-exit-race': {
    requiredProgress: [
      'started',
      'init-done',
      'initial-worker-ready',
      'respawn-in-flight',
      'shutdown-started',
      'terminate-invoked',
    ],
    requiredSuccessAsserts: [
      'one-worker-spawned',
      'worker-hard-exit',
      'respawn-in-flight',
      'closed-before-respawn-release',
      'pool-drained',
      'lint-batch-rejects-closed',
    ],
  },
  'task-timeout': {
    requiredProgress: [
      'started',
      'init-done',
      'hang-entered',
      'terminate-invoked',
    ],
    requiredSuccessAsserts: [
      'one-task-timeout-captured',
      'hang-task-timeout',
      'recovery-ok',
      'respawn-logged',
      'pool-drained',
    ],
  },
  'all-degraded': {
    requiredProgress: [
      'started',
      'init-done',
      'degraded-queue-ready',
      'terminate-invoked',
    ],
    requiredSuccessAsserts: [
      'queue-enqueued',
      'degraded-drain',
      'queue-drained',
    ],
  },
  'lint-batch-after-degraded': {
    requiredProgress: [
      'started',
      'init-done',
      'degraded-queue-ready',
      'terminate-invoked',
    ],
    requiredSuccessAsserts: [
      'first-queue-enqueued',
      'first-degraded',
      'terminal-state',
      'second-degraded',
      'queue-drained',
    ],
  },
} as const satisfies Record<string, ScenarioContract>;

export type Verdict = 'PASS' | 'EXPECTED-NATIVE-ABORT' | 'FAIL';

export interface JournalRecord {
  version: number;
  runId: string;
  scenario: string;
  seq: number;
  kind: 'milestone' | 'assert' | 'error';
  name?: string;
  pass?: boolean;
  detail?: string;
}

export interface ScenarioResult {
  scenario: string;
  verdict: Verdict;
  milestones: string[];
  asserts: { name: string; pass: boolean; detail?: string }[];
  error?: string;
  journalError?: string;
  exitCode: number | null;
  signal: NodeJS.Signals | null;
  timedOut: boolean;
  stderr: string;
  stderrTruncated: boolean;
}

export interface JournalReadResult {
  records: JournalRecord[];
  error?: string;
}

export interface ClassificationInput {
  scenario: string;
  records: JournalRecord[];
  journalError?: string;
  processError?: string;
  exitCode: number | null;
  signal: NodeJS.Signals | null;
  timedOut: boolean;
  platform?: NodeJS.Platform;
}

export async function runPoolScenario(
  scenario: string,
  opts: { timeoutMs?: number } = {},
): Promise<ScenarioResult> {
  // This is only a dead-process/deadlock sentinel. Scenario correctness is
  // established by recorded state transitions, never elapsed wall time.
  const timeoutMs = opts.timeoutMs ?? 60_000;
  const runId = randomUUID();
  const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-pool-iso-'));
  const milestoneFile = path.join(tmpDir, 'milestones.ndjson');
  const scenarioTmpDir = path.join(tmpDir, 'scenario');
  fs.mkdirSync(scenarioTmpDir);
  fs.writeFileSync(milestoneFile, '');

  return new Promise<ScenarioResult>((resolve) => {
    const child = spawn(process.execPath, [RUNNER, scenario], {
      stdio: ['ignore', 'pipe', 'pipe'],
      env: {
        ...process.env,
        RSLINT_MILESTONE_FILE: milestoneFile,
        RSLINT_DIST_ESLINT_PLUGIN: DIST_ESLINT_PLUGIN,
        RSLINT_RUN_ID: runId,
        RSLINT_SCENARIO_TMP_DIR: scenarioTmpDir,
      },
    });

    // Never let an unconsumed pipe backpressure the child. Scenario output is
    // intentionally irrelevant; durable state lives in the journal.
    child.stdout.resume();

    const stderrChunks: Buffer[] = [];
    let stderrBytes = 0;
    let stderrTruncated = false;
    child.stderr.on('data', (value: Buffer | string) => {
      const chunk = Buffer.isBuffer(value) ? value : Buffer.from(value);
      const remaining = STDERR_LIMIT_BYTES - stderrBytes;
      if (remaining > 0) {
        const kept = chunk.subarray(0, remaining);
        stderrChunks.push(kept);
        stderrBytes += kept.length;
      }
      if (chunk.length > remaining) stderrTruncated = true;
    });

    let timedOut = false;
    let spawnError: string | undefined;
    const timer = setTimeout(() => {
      // Do not resolve here: the journal and stdio are not final until
      // ChildProcess emits `close`. A late `done` cannot override timedOut.
      timedOut = true;
      child.kill('SIGKILL');
    }, timeoutMs);

    child.on('error', (err) => {
      // Node guarantees `close` after `error`; let that single terminal path
      // collect the complete journal and stream state.
      spawnError = `failed to run isolated child: ${err.message}`;
    });

    child.on('close', (exitCode, signal) => {
      clearTimeout(timer);
      const journal = readJournal(milestoneFile, { runId, scenario });
      let cleanupError: string | undefined;
      try {
        fs.rmSync(tmpDir, {
          recursive: true,
          force: true,
          maxRetries: 10,
          retryDelay: 50,
        });
      } catch (err) {
        cleanupError = `failed to remove scenario temp directory: ${String(
          err,
        )}`;
      }

      const milestones = journal.records
        .filter((r) => r.kind === 'milestone')
        .map((r) => r.name!);
      const asserts = journal.records
        .filter((r) => r.kind === 'assert')
        .map((r) => ({ name: r.name!, pass: r.pass!, detail: r.detail }));
      const reportedError = journal.records.find(
        (r) => r.kind === 'error',
      )?.detail;
      const processError = reportedError ?? spawnError ?? cleanupError;
      const stderr = Buffer.concat(stderrChunks).toString('utf8');

      resolve({
        scenario,
        verdict: classifyScenario({
          scenario,
          records: journal.records,
          journalError: journal.error,
          processError,
          exitCode,
          signal,
          timedOut,
        }),
        milestones,
        asserts,
        error: processError,
        journalError: journal.error,
        exitCode,
        signal,
        timedOut,
        stderr,
        stderrTruncated,
      });
    });
  });
}

function readJournal(
  file: string,
  expected: { runId: string; scenario: string },
): JournalReadResult {
  let text: string;
  try {
    const size = fs.statSync(file).size;
    if (size > JOURNAL_LIMIT_BYTES) {
      return {
        records: [],
        error: `journal exceeds ${JOURNAL_LIMIT_BYTES} byte limit (size=${size})`,
      };
    }
    text = fs.readFileSync(file, 'utf8');
  } catch (err) {
    return {
      records: [],
      error: `journal could not be read: ${String(err)}`,
    };
  }
  return parseJournalForTests(text, expected);
}

/** Pure parser exported for adversarial tests; production code uses readJournal. */
export function parseJournalForTests(
  text: string,
  expected: { runId: string; scenario: string },
): JournalReadResult {
  if (text.length === 0) return { records: [] };
  if (!text.endsWith('\n')) {
    return {
      records: [],
      error: 'journal ended with a torn record (missing newline)',
    };
  }

  const records: JournalRecord[] = [];
  const lines = text.slice(0, -1).split('\n');
  for (let index = 0; index < lines.length; index++) {
    const lineNumber = index + 1;
    if (lines[index].length === 0) {
      return {
        records,
        error: `journal contains an empty record at line ${lineNumber}`,
      };
    }

    let value: unknown;
    try {
      value = JSON.parse(lines[index]);
    } catch (err) {
      return {
        records,
        error: `journal contains invalid JSON at line ${lineNumber}: ${String(
          err,
        )}`,
      };
    }

    const recordError = validateRecord(value, {
      ...expected,
      seq: lineNumber,
    });
    if (recordError) {
      return {
        records,
        error: `journal record ${lineNumber} is invalid: ${recordError}`,
      };
    }
    records.push(value as JournalRecord);
  }

  const terminalIndexes = records.flatMap((record, index) =>
    record.kind === 'error' ||
    (record.kind === 'milestone' && record.name === 'done')
      ? [index]
      : [],
  );
  if (terminalIndexes.length > 1) {
    return {
      records,
      error: 'journal contains more than one terminal record',
    };
  }
  if (
    terminalIndexes.length === 1 &&
    terminalIndexes[0] !== records.length - 1
  ) {
    return {
      records,
      error: 'journal contains records after its terminal record',
    };
  }
  const assertionNames = records
    .filter((record) => record.kind === 'assert')
    .map((record) => record.name!);
  if (new Set(assertionNames).size !== assertionNames.length) {
    return {
      records,
      error: 'journal contains duplicate assertion names',
    };
  }

  return { records };
}

function validateRecord(
  value: unknown,
  expected: { runId: string; scenario: string; seq: number },
): string | undefined {
  if (value === null || typeof value !== 'object' || Array.isArray(value)) {
    return 'record must be an object';
  }
  const record = value as Record<string, unknown>;
  const allowedKeys = new Set([
    'version',
    'runId',
    'scenario',
    'seq',
    'kind',
    'name',
    'pass',
    'detail',
  ]);
  const extraKey = Object.keys(record).find((key) => !allowedKeys.has(key));
  if (extraKey) return `unknown field "${extraKey}"`;
  if (record.version !== JOURNAL_VERSION) {
    return `version must be ${JOURNAL_VERSION}`;
  }
  if (record.runId !== expected.runId) return 'runId does not match this run';
  if (record.scenario !== expected.scenario) {
    return 'scenario does not match this run';
  }
  if (record.seq !== expected.seq) {
    return `seq must be ${expected.seq}`;
  }
  if (
    record.kind !== 'milestone' &&
    record.kind !== 'assert' &&
    record.kind !== 'error'
  ) {
    return 'kind is not recognized';
  }

  if (record.kind === 'milestone') {
    if (typeof record.name !== 'string' || record.name.length === 0) {
      return 'milestone.name must be a non-empty string';
    }
    if (record.pass !== undefined || record.detail !== undefined) {
      return 'milestone may not contain pass or detail';
    }
  } else if (record.kind === 'assert') {
    if (typeof record.name !== 'string' || record.name.length === 0) {
      return 'assert.name must be a non-empty string';
    }
    if (typeof record.pass !== 'boolean') {
      return 'assert.pass must be a boolean';
    }
    if (record.detail !== undefined && typeof record.detail !== 'string') {
      return 'assert.detail must be a string when present';
    }
  } else {
    if (record.name !== undefined || record.pass !== undefined) {
      return 'error may not contain name or pass';
    }
    if (typeof record.detail !== 'string' || record.detail.length === 0) {
      return 'error.detail must be a non-empty string';
    }
  }
  return undefined;
}

/**
 * Strict verdict oracle. Exported only so test code can attack every branch
 * with synthetic journals without spawning more subprocesses.
 */
export function classifyScenario(input: ClassificationInput): Verdict {
  const platform = input.platform ?? process.platform;
  const assertions = input.records.filter((r) => r.kind === 'assert');
  if (input.journalError) return 'FAIL';
  if (input.records.some((record) => record.kind === 'error')) return 'FAIL';
  if (assertions.some((record) => record.pass !== true)) return 'FAIL';
  if (input.processError) return 'FAIL';
  // A watchdog firing is terminal even if the child managed to write `done`
  // just before it was killed.
  if (input.timedOut) return 'FAIL';

  const milestones = input.records
    .filter((r) => r.kind === 'milestone')
    .map((r) => r.name!);
  const last = input.records.at(-1);
  const contract = Object.prototype.hasOwnProperty.call(
    SCENARIO_CONTRACTS,
    input.scenario,
  )
    ? SCENARIO_CONTRACTS[input.scenario as keyof typeof SCENARIO_CONTRACTS]
    : undefined;
  if (!contract) return 'FAIL';
  const hasRequiredProgress = containsOrdered(
    milestones,
    contract.requiredProgress,
  );
  if (last?.kind === 'milestone' && last.name === 'done') {
    const assertNames = new Set(assertions.map((record) => record.name!));
    const hasAllSuccessAssertions = contract.requiredSuccessAsserts.every(
      (name) => assertNames.has(name),
    );
    return hasRequiredProgress &&
      hasAllSuccessAssertions &&
      input.exitCode === 0 &&
      input.signal === null
      ? 'PASS'
      : 'FAIL';
  }

  // A native napi teardown abort is tolerated only on Windows, only in an
  // explicitly enumerated terminate scenario, and only when the last durable
  // record proves the real Worker.terminate() call was entered. A generic
  // "about to terminate" marker is not sufficient.
  const exactNativeAbortBoundary =
    platform === 'win32' &&
    hasRequiredProgress &&
    last?.kind === 'milestone' &&
    last.name === 'terminate-invoked';
  return exactNativeAbortBoundary ? 'EXPECTED-NATIVE-ABORT' : 'FAIL';
}

function containsOrdered(
  actual: readonly string[],
  required: readonly string[],
): boolean {
  let requiredIndex = 0;
  for (const value of actual) {
    if (value === required[requiredIndex]) requiredIndex++;
  }
  return requiredIndex === required.length;
}

/**
 * Human-readable failure report for a scenario that did not pass.
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
  if (r.journalError) lines.push(`  invalid child journal: ${r.journalError}`);
  if (r.error) lines.push(`  child/harness error: ${r.error}`);
  if (r.timedOut) {
    lines.push('  child TIMED OUT - the scenario deadlocked or never exited');
  }
  lines.push(`  milestones reached: ${r.milestones.join(' > ') || '(none)'}`);
  lines.push(
    `  child exit: code=${r.exitCode ?? 'null'} signal=${r.signal ?? 'null'}`,
  );
  if (r.stderr.trim()) {
    lines.push(`  child stderr${r.stderrTruncated ? ' (truncated)' : ''}:`);
    for (const line of r.stderr.trim().split('\n')) {
      lines.push(`    ${line}`);
    }
  }
  return lines.join('\n');
}

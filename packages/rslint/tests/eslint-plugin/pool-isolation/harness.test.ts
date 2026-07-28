import { describe, expect, test } from '@rstest/core';

import {
  JOURNAL_VERSION,
  POOL_ISOLATION_AUDIT_PREFIX,
  WINDOWS_ACCESS_VIOLATION_EXIT_CODE,
  WINDOWS_ACCESS_VIOLATION_SIGNED_EXIT_CODE,
  classifyScenario,
  formatScenarioAudit,
  parseJournalForTests,
  type ClassificationInput,
  type JournalRecord,
} from './harness.js';

const RUN_ID = 'adversarial-run';
const SCENARIO = 'u11';

function records(
  ...items: Array<
    | { kind: 'milestone'; name: string }
    | { kind: 'assert'; name: string; pass: boolean; detail?: string }
    | { kind: 'error'; detail: string }
  >
): JournalRecord[] {
  return items.map((item, index) => ({
    version: JOURNAL_VERSION,
    runId: RUN_ID,
    scenario: SCENARIO,
    seq: index + 1,
    ...item,
  }));
}

const cleanSuccess = (): JournalRecord[] =>
  records(
    { kind: 'milestone', name: 'started' },
    { kind: 'milestone', name: 'init-done' },
    { kind: 'milestone', name: 'refed-plugin-loaded' },
    { kind: 'milestone', name: 'terminate-invoked' },
    { kind: 'assert', name: 'pool-drained', pass: true },
    { kind: 'milestone', name: 'done' },
  );

function classify(
  overrides: Partial<ClassificationInput> = {},
): ReturnType<typeof classifyScenario> {
  return classifyScenario({
    scenario: SCENARIO,
    records: cleanSuccess(),
    exitCode: 0,
    signal: null,
    timedOut: false,
    platform: 'darwin',
    ...overrides,
  });
}

describe('pool-isolation verdict is fail-closed', () => {
  test('accepts only a complete clean-success journal', () => {
    expect(classify()).toBe('PASS');
  });

  test('watchdog timeout wins even if done was already written', () => {
    expect(classify({ timedOut: true })).toBe('FAIL');
  });

  test('done cannot hide an abnormal process close', () => {
    expect(classify({ exitCode: null, signal: 'SIGTERM' })).toBe('FAIL');
  });

  test('clean child stderr cannot hide behind a complete journal', () => {
    expect(classify({ stderr: 'unexpected native warning\n' })).toBe('FAIL');
    expect(classify({ stderr: '  \n' })).toBe('PASS');
  });

  test('clean exit still requires the exact terminate boundary', () => {
    expect(
      classify({
        records: cleanSuccess().filter(
          (record) =>
            !(
              record.kind === 'milestone' && record.name === 'terminate-invoked'
            ),
        ),
      }),
    ).toBe('FAIL');
  });

  test('clean exit still requires every scenario assertion', () => {
    expect(
      classify({
        records: cleanSuccess().filter((record) => record.kind !== 'assert'),
      }),
    ).toBe('FAIL');
  });

  test('milestones must satisfy the scenario contract in order', () => {
    expect(
      classify({
        records: records(
          { kind: 'milestone', name: 'started' },
          { kind: 'milestone', name: 'refed-plugin-loaded' },
          { kind: 'milestone', name: 'init-done' },
          { kind: 'milestone', name: 'terminate-invoked' },
          { kind: 'assert', name: 'pool-drained', pass: true },
          { kind: 'milestone', name: 'done' },
        ),
      }),
    ).toBe('FAIL');
    expect(classify({ scenario: 'not-an-allowed-scenario' })).toBe('FAIL');
  });

  test('native abort exception is Windows-only and exact-boundary-only', () => {
    const nativeAbortRecords = cleanSuccess().slice(0, -2);
    expect(
      classify({
        records: nativeAbortRecords,
        exitCode: null,
        signal: 'SIGABRT',
        platform: 'win32',
      }),
    ).toBe('EXPECTED-NATIVE-ABORT');
    expect(
      classify({
        records: nativeAbortRecords,
        exitCode: null,
        signal: 'SIGTERM',
        platform: 'win32',
      }),
    ).toBe('FAIL');
    expect(
      classify({
        records: nativeAbortRecords,
        exitCode: 3,
        platform: 'win32',
      }),
    ).toBe('FAIL');
    expect(
      classify({
        records: nativeAbortRecords,
        exitCode: WINDOWS_ACCESS_VIOLATION_EXIT_CODE,
        platform: 'win32',
      }),
    ).toBe('EXPECTED-NATIVE-ABORT');
    expect(
      classify({
        records: nativeAbortRecords,
        exitCode: WINDOWS_ACCESS_VIOLATION_SIGNED_EXIT_CODE,
        platform: 'win32',
      }),
    ).toBe('EXPECTED-NATIVE-ABORT');
    expect(
      classify({
        records: nativeAbortRecords,
        exitCode: 0,
        platform: 'win32',
      }),
    ).toBe('FAIL');
    expect(
      classify({
        records: nativeAbortRecords,
        exitCode: null,
        signal: null,
        platform: 'win32',
      }),
    ).toBe('FAIL');
    expect(
      classify({
        records: nativeAbortRecords,
        exitCode: 0,
        platform: 'linux',
      }),
    ).toBe('FAIL');
    expect(
      classify({
        records: records(
          { kind: 'milestone', name: 'started' },
          { kind: 'milestone', name: 'init-done' },
          { kind: 'milestone', name: 'refed-plugin-loaded' },
          { kind: 'milestone', name: 'terminate-point' },
        ),
        platform: 'win32',
      }),
    ).toBe('FAIL');
  });

  test('failed assertion, orderly exit/error, and invalid journal always fail', () => {
    const failedAssert = cleanSuccess();
    failedAssert.splice(-1, 0, {
      version: JOURNAL_VERSION,
      runId: RUN_ID,
      scenario: SCENARIO,
      seq: failedAssert.length,
      kind: 'assert',
      name: 'injected-failure',
      pass: false,
    });
    expect(classify({ records: failedAssert })).toBe('FAIL');
    expect(
      classify({
        records: records(
          { kind: 'milestone', name: 'started' },
          { kind: 'milestone', name: 'init-done' },
          { kind: 'milestone', name: 'refed-plugin-loaded' },
          { kind: 'milestone', name: 'terminate-invoked' },
          {
            kind: 'error',
            detail: 'runner exited through Node before done',
          },
        ),
        exitCode: 0,
        platform: 'win32',
      }),
    ).toBe('FAIL');
    expect(classify({ processError: 'orderly failure' })).toBe('FAIL');
    expect(classify({ journalError: 'torn record' })).toBe('FAIL');
  });
});

describe('pool-isolation CI audit record', () => {
  test('is one machine-readable line that exposes clean versus tolerated exit', () => {
    const line = formatScenarioAudit({
      scenario: SCENARIO,
      verdict: 'EXPECTED-NATIVE-ABORT',
      milestones: [
        'started',
        'init-done',
        'refed-plugin-loaded',
        'terminate-invoked',
      ],
      asserts: [],
      exitCode: null,
      signal: 'SIGABRT',
      timedOut: false,
      stderr: '',
      stderrTruncated: false,
    });

    expect(line.startsWith(POOL_ISOLATION_AUDIT_PREFIX)).toBe(true);
    expect(line).not.toContain('\n');
    expect(JSON.parse(line.slice(POOL_ISOLATION_AUDIT_PREFIX.length))).toEqual({
      version: 1,
      scenario: SCENARIO,
      verdict: 'EXPECTED-NATIVE-ABORT',
      exitCode: null,
      signal: 'SIGABRT',
      timedOut: false,
      lastMilestone: 'terminate-invoked',
      milestoneCount: 4,
      assertionCount: 0,
      failedAssertionCount: 0,
      journalValid: true,
      harnessError: false,
      stderrBytes: 0,
      stderrTruncated: false,
    });
  });
});

describe('pool-isolation journal parser rejects ambiguous evidence', () => {
  const expected = { runId: RUN_ID, scenario: SCENARIO };
  const line = (record: JournalRecord): string => `${JSON.stringify(record)}\n`;

  test('accepts a fully framed, sequential journal', () => {
    const journal = cleanSuccess().map(line).join('');
    const result = parseJournalForTests(journal, expected);
    expect(result.error).toBeUndefined();
    expect(result.records).toHaveLength(6);
  });

  test('rejects torn and invalid JSON records instead of ignoring them', () => {
    const first = cleanSuccess()[0];
    expect(parseJournalForTests(JSON.stringify(first), expected).error).toMatch(
      /torn/,
    );
    expect(parseJournalForTests('{broken}\n', expected).error).toMatch(
      /invalid JSON/,
    );
  });

  test('rejects stale/forged run identity and sequence gaps', () => {
    const first = cleanSuccess()[0];
    expect(
      parseJournalForTests(
        line({ ...first, runId: 'some-other-run' }),
        expected,
      ).error,
    ).toMatch(/runId/);
    expect(
      parseJournalForTests(line({ ...first, seq: 2 }), expected).error,
    ).toMatch(/seq/);
  });

  test('rejects records after done and duplicate assertion names', () => {
    const afterDone = records(
      { kind: 'milestone', name: 'done' },
      { kind: 'milestone', name: 'late' },
    )
      .map(line)
      .join('');
    expect(parseJournalForTests(afterDone, expected).error).toMatch(
      /after its terminal/,
    );

    const duplicate = records(
      { kind: 'assert', name: 'same', pass: true },
      { kind: 'assert', name: 'same', pass: true },
    )
      .map(line)
      .join('');
    expect(parseJournalForTests(duplicate, expected).error).toMatch(
      /duplicate assertion/,
    );
  });
});

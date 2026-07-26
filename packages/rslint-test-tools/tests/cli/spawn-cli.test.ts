import { describe, expect, test } from '@rstest/core';

import { validateConformanceStderrContract } from './js-config/eslint-plugin-conformance/stderr-contract.js';
import { validateCliClose } from './spawn-cli.js';

/** Windows NTSTATUS 0xC0000005, in unsigned and signed representations. */
const WINDOWS_ACCESS_VIOLATION_EXIT_CODE = 0xc0000005;
const WINDOWS_ACCESS_VIOLATION_SIGNED_EXIT_CODE =
  WINDOWS_ACCESS_VIOLATION_EXIT_CODE - 0x1_0000_0000;

describe('CLI child close validation', () => {
  const clean = {
    timedOut: false,
    code: 0,
    signal: null,
  } as const;

  test('accepts only the CLI contract exit codes', () => {
    expect(validateCliClose(clean)).toBeUndefined();
    expect(validateCliClose({ ...clean, code: 1 })).toBeUndefined();
    expect(validateCliClose({ ...clean, code: 2 })).toBeUndefined();
  });

  test('rejects numeric crash and unsupported exit codes', () => {
    for (const code of [
      3,
      134,
      137,
      WINDOWS_ACCESS_VIOLATION_EXIT_CODE,
      WINDOWS_ACCESS_VIOLATION_SIGNED_EXIT_CODE,
    ]) {
      expect(validateCliClose({ ...clean, code })).toMatch(
        /unsupported exit code/,
      );
    }
  });

  test('rejects null code even when no signal is reported', () => {
    expect(validateCliClose({ ...clean, code: null })).toMatch(
      /closed abnormally/,
    );
  });

  test('rejects signal, timeout, and spawn-error exits', () => {
    expect(validateCliClose({ ...clean, signal: 'SIGABRT' })).toMatch(
      /SIGABRT/,
    );
    expect(validateCliClose({ ...clean, timedOut: true })).toMatch(/timed out/);
    expect(
      validateCliClose({
        ...clean,
        spawnError: new Error('synthetic ENOENT'),
      }),
    ).toMatch(/synthetic ENOENT/);
    expect(validateCliClose({ ...clean, outputOverflow: true })).toMatch(
      /output exceeded/,
    );
  });

  test('rejects runtime-fatal output even with a supported CLI exit code', () => {
    for (const output of [
      'panic: synthetic Go panic\n\ngoroutine 1 [running]:',
      'fatal error: concurrent map writes',
      'FATAL ERROR: Reached heap limit',
      "thread 'main' panicked at src/main.rs:1:1:",
      'thread "worker" (42) panicked at src/lib.rs:1:1:',
      'fatal runtime error: Rust cannot catch foreign exceptions',
      'Error [ERR_WORKER_OUT_OF_MEMORY]: Worker terminated',
    ]) {
      expect(validateCliClose({ ...clean, code: 2, stderr: output })).toMatch(
        /runtime-fatal marker/,
      );
    }
  });

  test('does not confuse ordinary diagnostic text with a runtime fatal', () => {
    expect(
      validateCliClose({
        ...clean,
        code: 1,
        stdout: 'panic: is literal fixture output\nFATAL ERROR: also a fixture',
        stderr: 'fatal error handling config: expected fixture diagnostic',
      }),
    ).toBeUndefined();
  });
});

describe('CLI conformance stderr validation', () => {
  test('accepts an empty contract and an exact match', () => {
    expect(validateConformanceStderrContract([], [])).toBeUndefined();
    expect(
      validateConformanceStderrContract(['warning-a'], ['warning-a']),
    ).toBeUndefined();
    expect(
      validateConformanceStderrContract(
        ['warning-b', 'warning-a'],
        ['warning-a', 'warning-b'],
      ),
    ).toBeUndefined();
  });

  test('accepts scheduling-dependent warning de-duplication', () => {
    expect(
      validateConformanceStderrContract(
        ['warning-a', 'warning-a'],
        ['warning-a', 'warning-a', 'warning-a'],
      ),
    ).toBeUndefined();
  });

  test('requires every distinct expected warning', () => {
    expect(
      validateConformanceStderrContract(
        ['warning-a'],
        ['warning-a', 'warning-b'],
      ),
    ).toContain('missing required stderr line "warning-b"');
  });

  test('rejects every unrelated stderr line', () => {
    expect(
      validateConformanceStderrContract(
        ['warning-a', 'unexpected'],
        ['warning-a', 'warning-a'],
      ),
    ).toContain('unexpected stderr line "unexpected" (1 occurrence)');
  });

  test('rejects warning amplification beyond annotated cases', () => {
    expect(
      validateConformanceStderrContract(
        ['warning-a', 'warning-a', 'warning-a'],
        ['warning-a', 'warning-a'],
      ),
    ).toContain('stderr line "warning-a" occurred 3 times; expected at most 2');
  });
});

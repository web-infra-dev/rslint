import { describe, test, expect, beforeAll, afterAll } from '@rstest/core';
import { spawn } from 'child_process';
import { RSLINT_BIN, createTempDir, cleanupTempDir } from './js-config/helpers';

/**
 * End-to-end pins for the color-decision behavior matrix (issue #1080).
 *
 * Each row spawns the real CLI (Node host + Go binary over IPC) with a fully
 * explicit color environment and asserts the presence or absence of ANSI
 * escapes in stdout. stdout here is always a pipe (never a TTY), so the
 * colored rows exercise the force-on tiers and the colorless rows pin both
 * the disable tiers and their precedence over weaker force-on signals.
 * The TTY tier itself is pinned by packages/rslint/tests/engine.test.ts
 * (fact wiring) plus the Go table test (decision).
 */

const ANSI = /\x1b\[/;

interface CliResult {
  exitCode: number;
  stdout: string;
  stderr: string;
}

// Baseline env: strip every color-deciding variable inherited from the
// runner (CI sets GITHUB_ACTIONS; developers may set the others), then pin
// TERM so a TERM=dumb runner can't flip tier 5. Each row layers its own
// overrides on top of this deterministic base.
const { GITHUB_ACTIONS, FORCE_COLOR, NO_COLOR, ...restEnv } = process.env;
const BASE_ENV = { ...restEnv, TERM: 'xterm-256color' };

function runCli(
  args: string[],
  cwd: string,
  envOverrides: Record<string, string> = {},
): Promise<CliResult> {
  return new Promise((resolve) => {
    const child = spawn(process.execPath, [RSLINT_BIN, ...args], {
      cwd,
      stdio: ['pipe', 'pipe', 'pipe'],
      env: { ...BASE_ENV, ...envOverrides },
    });
    let stdout = '';
    let stderr = '';
    child.stdout.on('data', (d: Buffer) => (stdout += d.toString()));
    child.stderr.on('data', (d: Buffer) => (stderr += d.toString()));
    child.on('close', (code) => {
      resolve({ exitCode: code ?? 0, stdout, stderr });
    });
  });
}

describe('CLI color matrix', () => {
  let tempDir: string;

  beforeAll(async () => {
    tempDir = await createTempDir({
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['**/*.ts'],
      }),
      'rslint.config.mjs': `export default [
        {
          files: ['**/*.ts'],
          rules: { '@typescript-eslint/ban-ts-comment': 'error' },
          plugins: ['@typescript-eslint'],
        },
      ];`,
      'src/index.ts': '// @ts-ignore\nconst a = 1;\n',
    });
  });

  afterAll(async () => {
    await cleanupTempDir(tempDir);
  });

  // Every row must observe a real diagnostic run: exit 1 with the rule name
  // in stdout. Otherwise a broken/empty run would vacuously pass the
  // "no ANSI" rows.
  async function expectLintRun(
    args: string[],
    env: Record<string, string>,
    colored: boolean,
  ): Promise<void> {
    const result = await runCli(args, tempDir, env);
    expect(result.exitCode).toBe(1);
    expect(result.stdout).toContain('ban-ts-comment');
    expect(ANSI.test(result.stdout)).toBe(colored);
  }

  test('piped stdout with clean env stays colorless', async () => {
    await expectLintRun([], {}, false);
  });

  test('--force-color colors a piped stdout', async () => {
    await expectLintRun(['--force-color'], {}, true);
  });

  test('FORCE_COLOR=1 colors a piped stdout', async () => {
    await expectLintRun([], { FORCE_COLOR: '1' }, true);
  });

  test('FORCE_COLOR=0 force-disables, beating GITHUB_ACTIONS', async () => {
    await expectLintRun(
      [],
      { FORCE_COLOR: '0', GITHUB_ACTIONS: 'true' },
      false,
    );
  });

  test('NO_COLOR disables', async () => {
    await expectLintRun([], { NO_COLOR: '1' }, false);
  });

  test('FORCE_COLOR=1 beats NO_COLOR (Node semantics)', async () => {
    await expectLintRun([], { NO_COLOR: '1', FORCE_COLOR: '1' }, true);
  });

  test('FORCE_COLOR set-but-empty enables (LookupEnv, not non-empty)', async () => {
    await expectLintRun([], { FORCE_COLOR: '' }, true);
  });

  test('NO_COLOR set-but-empty disables, beating GITHUB_ACTIONS', async () => {
    await expectLintRun([], { NO_COLOR: '', GITHUB_ACTIONS: 'true' }, false);
  });

  test('--no-color beats FORCE_COLOR=1', async () => {
    await expectLintRun(['--no-color'], { FORCE_COLOR: '1' }, false);
  });

  test('--force-color beats NO_COLOR', async () => {
    await expectLintRun(['--force-color'], { NO_COLOR: '1' }, true);
  });

  test('GITHUB_ACTIONS forces color on a pipe (documented deviation)', async () => {
    await expectLintRun([], { GITHUB_ACTIONS: 'true' }, true);
  });

  test('TERM=dumb beats GITHUB_ACTIONS', async () => {
    await expectLintRun([], { TERM: 'dumb', GITHUB_ACTIONS: 'true' }, false);
  });
});

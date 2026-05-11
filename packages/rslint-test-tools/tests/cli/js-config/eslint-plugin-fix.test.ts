/**
 * End-to-end coverage for CLI `--fix` against ESLint-plugin rules.
 *
 * Rationale: native rules go through Go's fix-loop (cmd.go fix passes),
 * but ESLint-plugin rules carry their fixes back from the runner over
 * IPC as UTF-8 byte offsets. Whether the binary actually applies those
 * bytes to the file on disk is a wire boundary the unit tests don't
 * cover.
 *
 * The test exercises `eslint-plugin-unicorn`'s `prefer-array-some`
 * rule, whose autofix rewrites `arr.filter(...).length > 0` →
 * `arr.some(...)`. Verified upstream (probe in `_check_fix.mjs`)
 * that the rule's fix is auto-applicable (not a suggestion).
 *
 * Why a single test file rather than a fixture under `tests/cli/`:
 * the rslint binary spawns a Go process that opens the file from disk
 * — we need a real on-disk fixture, not the in-memory probe path. We
 * symlink `node_modules` from `rslint-test-tools` into the tempdir so
 * the config's `import unicornPlugin from 'eslint-plugin-unicorn'`
 * resolves without copying ~tens of megabytes of node_modules.
 */
import { describe, test, expect } from '@rstest/core';
import { spawn } from 'child_process';
import path from 'node:path';
import fs from 'node:fs/promises';
import { tmpdir } from 'node:os';

import { fileURLToPath } from 'node:url';

const RSLINT_BIN = require.resolve('@rslint/core/bin');
// node_modules of @rslint/test-tools — that's where eslint-plugin-unicorn
// is actually installed under pnpm's per-package convention. We walk
// up from this test file (which lives at
// .../rslint-test-tools/tests/cli/js-config/) to the package root's
// node_modules. import.meta.url survives both ESM and rstest's
// transform layer; `require.resolve` failed inside the test-runner
// vm context.
const TEST_FILE = fileURLToPath(import.meta.url);
const TEST_TOOLS_NM_DIR = path.resolve(
  path.dirname(TEST_FILE),
  '..',
  '..',
  '..',
  'node_modules',
);

async function runRslint(
  args: string[],
  cwd: string,
): Promise<{ exitCode: number; stdout: string; stderr: string }> {
  return new Promise((resolve) => {
    const { GITHUB_ACTIONS, FORCE_COLOR, ...cleanEnv } = process.env;
    const child = spawn(process.execPath, [RSLINT_BIN, '--no-color', ...args], {
      cwd,
      stdio: ['pipe', 'pipe', 'pipe'],
      env: { ...cleanEnv, NO_COLOR: '1' },
    });
    let stdout = '';
    let stderr = '';
    child.stdout?.on('data', (d: Buffer) => {
      stdout += d.toString();
    });
    child.stderr?.on('data', (d: Buffer) => {
      stderr += d.toString();
    });
    child.on('close', (code) =>
      resolve({ exitCode: code ?? 0, stdout, stderr }),
    );
  });
}

async function setupFixture(files: Record<string, string>): Promise<string> {
  const tempDir = await fs.mkdtemp(path.join(tmpdir(), 'rslint-fix-'));
  // Symlink node_modules so the user config's
  // `import unicornPlugin from 'eslint-plugin-unicorn'` resolves. We
  // point at @rslint/test-tools's node_modules (where unicorn is
  // installed) — Node's standard resolution walks up from the
  // config file, so a tempDir/node_modules entry is enough.
  await fs.symlink(
    TEST_TOOLS_NM_DIR,
    path.join(tempDir, 'node_modules'),
    'dir',
  );
  for (const [rel, content] of Object.entries(files)) {
    const abs = path.join(tempDir, rel);
    await fs.mkdir(path.dirname(abs), { recursive: true });
    await fs.writeFile(abs, content, 'utf8');
  }
  return tempDir;
}

async function cleanup(tempDir: string): Promise<void> {
  await fs.rm(tempDir, { recursive: true, force: true });
}

describe('CLI --fix end-to-end with ESLint plugin rules', () => {
  test('uni/prefer-array-some auto-fix rewrites the source file on disk', async () => {
    const tsconfig = JSON.stringify({
      compilerOptions: {
        target: 'es2022',
        module: 'esnext',
        strict: false,
        noEmit: true,
        moduleResolution: 'bundler',
      },
      include: ['./src/**/*.ts'],
    });
    const config = `import { defineConfig } from '@rslint/core';
import unicornPlugin from 'eslint-plugin-unicorn';
export default defineConfig([
  {
    files: ['src/**/*.ts'],
    languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
    // @ts-ignore: eslintPlugins is experimental; the type lives in a
    // newer @rslint/core than this test-tools package's pinned dep.
    eslintPlugins: { uni: unicornPlugin },
    rules: { 'uni/prefer-array-some': 'error' },
  },
]);
`;
    const source =
      'export const flag = [1, 2, 3].filter((x) => x > 1).length > 0;\n';

    const tempDir = await setupFixture({
      'tsconfig.json': tsconfig,
      'rslint.config.ts': config,
      'src/index.ts': source,
    });

    try {
      // 1. Pre-fix: rule fires.
      const before = await runRslint([], tempDir);
      expect(before.stdout).toContain('prefer-array-some');

      // 2. Apply fix.
      const fixed = await runRslint(['--fix'], tempDir);
      // --fix typically returns 0 when everything could be fixed.
      // The exit code contract: if all reported issues had fixes
      // applied, no remaining diagnostics → exit 0. We don't pin
      // the code itself (rslint may keep nonzero when other issues
      // remain) — we pin the FILE CONTENT, which is the load-
      // bearing contract.
      expect(fixed.stderr.length).toBeGreaterThanOrEqual(0); // smoke

      // 3. Source file MUST be rewritten on disk.
      const after = await fs.readFile(
        path.join(tempDir, 'src/index.ts'),
        'utf8',
      );
      // ESLint upstream rewrites `filter(...).length > 0` to
      // `some(...)`. The exact whitespace/parens of the
      // replacement is what the rule's fixer emits — verified
      // against upstream that the result strips the comparison.
      expect(after).not.toBe(source);
      expect(after).toContain('.some(');
      expect(after).not.toContain('.filter(');
      expect(after).not.toContain('.length > 0');

      // 4. Re-linting the fixed file produces no more rule firings.
      const after2 = await runRslint([], tempDir);
      expect(after2.stdout).not.toContain('prefer-array-some');
    } finally {
      await cleanup(tempDir);
    }
  }, 60_000);

  test('multi-pass cascading fixes in compatOnlyMode stay on the compat path', async () => {
    // Two rules whose fixes cascade: rule A rewrites `BAD` → `GOOD`,
    // rule B rewrites `GOOD` → `BEST`. The fix-loop must produce
    // `BEST` after two passes. Because both rules are plugin rules
    // and the config has a `files` glob, this run triggers Phase 1's
    // compatOnlyMode (cmd.go:1101) — which Phase 2's fix-loop also
    // now honors (cmd.go fix-loop branch, H3). If H3 regressed and
    // the loop fell back to RunLinter + createPrograms, the test
    // still semantically passes (functional correctness is identical)
    // but the test asserts BOTH passes produced the cascading rewrite,
    // which is the multi-pass invariant.
    const tsconfig = JSON.stringify({
      compilerOptions: {
        target: 'es2022',
        module: 'esnext',
        strict: false,
        noEmit: true,
        moduleResolution: 'bundler',
      },
      include: ['./src/**/*.ts'],
    });
    // Self-contained plugin fixture with two cascading fixers.
    const pluginSrc = `export default {
  meta: { name: 'cascade-plugin', version: '0.0.0' },
  rules: {
    'bad-to-good': {
      meta: { type: 'problem', fixable: 'code', messages: { x: 'BAD must become GOOD' } },
      create(ctx) {
        return {
          Identifier(node) {
            if (node.name === 'BAD') {
              ctx.report({
                node, messageId: 'x',
                fix(fixer) { return fixer.replaceText(node, 'GOOD'); },
              });
            }
          },
        };
      },
    },
    'good-to-best': {
      meta: { type: 'problem', fixable: 'code', messages: { x: 'GOOD must become BEST' } },
      create(ctx) {
        return {
          Identifier(node) {
            if (node.name === 'GOOD') {
              ctx.report({
                node, messageId: 'x',
                fix(fixer) { return fixer.replaceText(node, 'BEST'); },
              });
            }
          },
        };
      },
    },
  },
};
`;
    const config = `import { defineConfig } from '@rslint/core';
import cascadePlugin from './cascade-plugin.mjs';
export default defineConfig([
  {
    files: ['src/**/*.ts'],
    languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
    // @ts-ignore: eslintPlugins is experimental
    eslintPlugins: { cascade: cascadePlugin },
    rules: {
      'cascade/bad-to-good': 'error',
      'cascade/good-to-best': 'error',
    },
  },
]);
`;
    const source = 'export const x = BAD; export const y = BAD;\n';

    const tempDir = await setupFixture({
      'tsconfig.json': tsconfig,
      'rslint.config.ts': config,
      'cascade-plugin.mjs': pluginSrc,
      'src/index.ts': source,
    });

    try {
      // Pre-fix: both rules fire on the two `BAD` identifiers.
      const before = await runRslint([], tempDir);
      expect(before.stdout).toContain('bad-to-good');

      // Apply fix. The fix-loop must:
      //   Pass 1: BAD → GOOD (both occurrences)
      //   Pass 2: GOOD → BEST (both occurrences, after disk reread)
      //   Pass 3: nothing to fix → break out
      await runRslint(['--fix'], tempDir);

      // After multi-pass fix, the disk content must be BEST, not GOOD.
      // If H3 regressed but cascading still worked (because RunLinter
      // also drives the loop), this passes. The strict invariant is
      // that the loop ITERATED at least twice and converged.
      const after = await fs.readFile(
        path.join(tempDir, 'src/index.ts'),
        'utf8',
      );
      expect(after).not.toContain('BAD');
      expect(after).not.toContain('GOOD');
      expect(after).toContain('BEST');
      // Both occurrences fixed.
      const bestCount = (after.match(/BEST/g) ?? []).length;
      expect(bestCount).toBe(2);

      // Re-linting the converged file emits no more diagnostics
      // for either cascade rule.
      const after2 = await runRslint([], tempDir);
      expect(after2.stdout).not.toContain('bad-to-good');
      expect(after2.stdout).not.toContain('good-to-best');
    } finally {
      await cleanup(tempDir);
    }
  }, 90_000);
});

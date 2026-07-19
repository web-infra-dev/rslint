import { describe, test, expect } from '@rstest/core';
import { runRslint, createTempDir, cleanupTempDir, TS_CONFIG } from './helpers';

// End-to-end coverage for the `--type-check-only` mode (PR #905) and the
// short-circuit / lint-only-warning fixes that landed alongside it (PR #905
// review-111). The contracts under test:
//
//   1. Type-check-only skips the lint phase entirely (no rule diagnostics,
//      no lint-file count) while still running Phase 2 type-check.
//   2. `--fix` and `--rule` are rejected because the lint phase is gone.
//   3. Lint-phase per-file warnings ("X was not found" and
//      "X is ignored because of a matching ignore pattern") are suppressed
//      in --type-check-only — they describe a phase that didn't run, and
//      Phase 2 itself ignores rslint's `ignores`/scope.
//   4. `--type-check` (non-only) must NOT take the empty-scope short-circuit
//      either: Phase 2 runs program-wide and may produce diagnostics even
//      when Phase 1 visited zero files.
//   5. In plain lint mode the existing "is ignored" warning still fires —
//      lock-in so the suppression in (3) doesn't accidentally regress it.
//
// All assertions go through the actual CLI binary so behavior is verified
// end-to-end (config loading, ignores resolution, program scoping, output).

interface Diagnostic {
  ruleName: string;
  filePath: string;
  severity?: string;
}

function makeConfigPlain(): string {
  return `export default [
  {
    files: ['**/*.ts'],
    languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
    plugins: ['@typescript-eslint'],
    rules: { '@typescript-eslint/no-unsafe-member-access': 'error' },
  }
];
`;
}

function makeConfigWithIgnores(ignores: string[]): string {
  return `export default [
  { ignores: ${JSON.stringify(ignores)} },
  {
    files: ['**/*.ts'],
    languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
    plugins: ['@typescript-eslint'],
    rules: { '@typescript-eslint/no-unsafe-member-access': 'error' },
  }
];
`;
}

function parseJsonline(stdout: string): Diagnostic[] {
  return stdout
    .trim()
    .split('\n')
    .filter((l) => l.trim().startsWith('{'))
    .map((l) => JSON.parse(l) as Diagnostic);
}

describe('--type-check-only basic', () => {
  test('reports type errors and skips lint diagnostics on the same file', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': makeConfigPlain(),
      // Triggers BOTH:
      //   - @typescript-eslint/no-unsafe-member-access (lint, on `a.b`)
      //   - TS2322 (type-check, on `const x: number = 'oops'`)
      'a.ts': "const x: number = 'oops';\nlet a: any = 1;\na.b = 2;\n",
    });
    try {
      const r = await runRslint(
        ['--type-check-only', '--format', 'jsonline'],
        tempDir,
      );
      const diags = parseJsonline(r.stdout);
      const tsDiags = diags.filter((d) => d.ruleName.startsWith('TypeScript('));
      const lintDiags = diags.filter(
        (d) => !d.ruleName.startsWith('TypeScript('),
      );
      expect(tsDiags.length).toBeGreaterThan(0);
      expect(tsDiags.some((d) => d.ruleName.includes('TS2322'))).toBe(true);
      expect(lintDiags.length).toBe(0);
      expect(r.exitCode).not.toBe(0);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('implies --type-check (no need to pass both)', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': makeConfigPlain(),
      'a.ts': "const x: number = 'oops';\n",
    });
    try {
      // Without --type-check passed explicitly: should still type-check.
      const r = await runRslint(
        ['--type-check-only', '--format', 'jsonline'],
        tempDir,
      );
      const diags = parseJsonline(r.stdout);
      expect(diags.some((d) => d.ruleName.includes('TS2322'))).toBe(true);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('summary reports "type-checked N files" instead of "linted N files"', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': makeConfigPlain(),
      'a.ts': 'export const ok = 1;\n',
    });
    try {
      const r = await runRslint(
        ['--type-check-only', '--singleThreaded'],
        tempDir,
      );
      expect(r.stdout).toContain('type-checked');
      expect(r.stdout).toContain('type error');
      expect(r.stdout).toMatch(
        /Found 0 type errors \(type-checked 1 file in .+ using 1 thread\)\n/,
      );
      expect(r.stdout).not.toContain('using 1 threads');
      // The summary line owns "linted N files"; --type-check-only uses a
      // different summary so this phrasing should NOT appear.
      expect(r.stdout).not.toContain('linted');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('gitignored nested config does not contribute a type-check Program', async () => {
    const tempDir = await createTempDir({
      '.gitignore': 'packages/ignored/\n',
      'rslint.config.mjs': makeConfigPlain(),
      'tsconfig.json': JSON.stringify({
        compilerOptions: {
          target: 'ES2020',
          module: 'ESNext',
          strict: true,
        },
        files: ['root.ts'],
      }),
      'root.ts': "export const root: number = 'wrong';\n",
      'packages/ignored/rslint.config.mjs': makeConfigPlain(),
      'packages/ignored/tsconfig.json': JSON.stringify({
        compilerOptions: {
          target: 'ES2020',
          module: 'ESNext',
          strict: true,
        },
        files: ['bad.ts'],
      }),
      'packages/ignored/bad.ts': "export const bad: number = 'wrong';\n",
    });
    try {
      const r = await runRslint(
        ['--type-check-only', '--format', 'jsonline'],
        tempDir,
      );
      const diags = parseJsonline(r.stdout);
      expect(
        diags.some(
          (diagnostic) =>
            diagnostic.filePath?.includes('root.ts') &&
            diagnostic.ruleName.includes('TS2322'),
        ),
      ).toBe(true);
      expect(
        diags.some(
          (diagnostic) =>
            diagnostic.filePath?.includes('packages/ignored/bad.ts') &&
            diagnostic.ruleName.includes('TS2322'),
        ),
      ).toBe(false);
      expect(r.exitCode).not.toBe(0);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

describe('--type-check-only flag compatibility', () => {
  test('rejects --fix with exit code 2', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': makeConfigPlain(),
      'a.ts': 'export const ok = 1;\n',
    });
    try {
      const r = await runRslint(['--type-check-only', '--fix'], tempDir);
      expect(r.exitCode).toBe(2);
      expect(r.stderr).toContain('--fix');
      expect(r.stderr).toContain('--type-check-only');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('rejects --rule with exit code 2', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': makeConfigPlain(),
      'a.ts': 'export const ok = 1;\n',
    });
    try {
      const r = await runRslint(
        ['--type-check-only', '--rule', 'no-console: error'],
        tempDir,
      );
      expect(r.exitCode).toBe(2);
      expect(r.stderr).toContain('--rule');
      expect(r.stderr).toContain('--type-check-only');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

describe('--type-check-only suppresses lint-phase warnings', () => {
  test('"was not found" warning is suppressed', async () => {
    // A missing CLI-specified file is a lint-phase warning. Phase 2 is
    // program-wide and unaffected, so the warning is noise in
    // --type-check-only.
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': makeConfigPlain(),
      'existing.ts': 'export const ok = 1;\n',
    });
    try {
      const r = await runRslint(
        ['--type-check-only', 'nonexistent.ts'],
        tempDir,
      );
      expect(r.stderr).not.toContain('was not found');
      expect(r.stderr).not.toContain('nonexistent.ts');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('"is ignored because of a matching ignore pattern" warning is suppressed', async () => {
    // The user asked for type-check; the lint-phase "ignored" notice would
    // be misleading next to a type error on the same file (see (3) below).
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': makeConfigWithIgnores(['ignored.ts']),
      'ignored.ts': "const x: number = 'bad';\n", // TS2322 still surfaces
    });
    try {
      const r = await runRslint(['--type-check-only', 'ignored.ts'], tempDir);
      expect(r.stderr).not.toContain('matching ignore pattern');
      expect(r.stderr).not.toContain('is ignored');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('type errors are still reported for files matched by rslint ignores', async () => {
    // Documents the contract from type-checking.md L145-L152: rslint
    // `ignores` is a lint-phase concept. Type-check is program-wide.
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': makeConfigWithIgnores(['ignored.ts']),
      'ignored.ts': "const x: number = 'bad';\n",
    });
    try {
      const r = await runRslint(
        ['--type-check-only', '--format', 'jsonline'],
        tempDir,
      );
      const diags = parseJsonline(r.stdout);
      const onIgnored = diags.filter((d) => d.filePath?.includes('ignored.ts'));
      expect(onIgnored.some((d) => d.ruleName.includes('TS2322'))).toBe(true);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

describe('--type-check (non-only) does not short-circuit on Phase 2 diagnostics', () => {
  test('--type-check with a non-program file argument still reports program-wide type errors', async () => {
    // Locks review-111 Issue 1: previously, `--type-check nonexistent.ts`
    // would set scopeRestricted=true, lintedFileCount=0, and trip the
    // short-circuit — silently dropping the TS2322 in src/bad.ts that
    // Phase 2 had already collected.
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': makeConfigPlain(),
      'src/bad.ts': "const x: number = 'oops';\n", // TS2322
    });
    try {
      const r = await runRslint(
        ['--type-check', '--format', 'jsonline', 'nonexistent.ts'],
        tempDir,
      );
      const diags = parseJsonline(r.stdout);
      // The src/bad.ts type error must surface — short-circuiting on
      // lintedFileCount==0 used to swallow it.
      expect(
        diags.some(
          (d) =>
            d.filePath?.includes('src/bad.ts') && d.ruleName.includes('TS2322'),
        ),
      ).toBe(true);
      // And the lint-side "not found" warning still fires in --type-check
      // (non-only) mode — only --type-check-only suppresses it.
      expect(r.stderr).toContain('nonexistent.ts');
      expect(r.stderr).toContain('was not found');

      const summary = await runRslint(
        ['--type-check', '--singleThreaded', 'nonexistent.ts'],
        tempDir,
      );
      expect(summary.stdout).toMatch(
        /Found 0 lint errors, 1 type error and 0 warnings \(linted 0 files with 0 rules, type-checked 1 file in .+ using 1 thread\)\n/,
      );
      expect(summary.stdout).not.toContain('using 1 threads');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

describe('lint mode baseline (lock-in)', () => {
  test('"is ignored because of a matching ignore pattern" warning surfaces in plain lint mode', async () => {
    // Counterpart to the --type-check-only suppression test above: in
    // plain lint mode the warning MUST still appear, otherwise the
    // suppression patch would have over-corrected.
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': makeConfigWithIgnores(['ignored.ts']),
      'ignored.ts': 'export const ok = 1;\n',
      // A second non-ignored file so the run produces a clean summary.
      'kept.ts': 'export const kept = 2;\n',
    });
    try {
      const r = await runRslint(['ignored.ts'], tempDir);
      expect(r.stderr).toContain('ignored.ts');
      expect(r.stderr).toContain(
        'is ignored because of a matching ignore pattern',
      );
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

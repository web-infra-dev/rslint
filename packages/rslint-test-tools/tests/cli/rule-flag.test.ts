import { describe, test, expect } from '@rstest/core';
import {
  runRslint,
  createTempDir,
  cleanupTempDir,
  TS_CONFIG,
} from './js-config/helpers.js';

// Generate a JS config (.mjs) with only the specified rules enabled.
// No rules are auto-enabled — only what's explicitly listed will run.
function mjsConfig(rules: Record<string, unknown>): string {
  const entry = {
    files: ['**/*.ts'],
    languageOptions: {
      parserOptions: {
        projectService: false,
        project: ['./tsconfig.json'],
      },
    },
    rules,
    plugins: ['@typescript-eslint'],
  };
  return `export default [${JSON.stringify(entry)}];`;
}

// Multi-entry JS config for per-file override tests
function mjsMultiConfig(entries: Record<string, unknown>[]): string {
  return `export default ${JSON.stringify(entries)};`;
}

describe('CLI --rule flag', () => {
  // -----------------------------------------------------------------------
  // Basic usage
  // -----------------------------------------------------------------------

  test('--rule enables a rule that was off in config', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': mjsConfig({ 'no-debugger': 'off' }),
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'debugger;\n',
    });

    try {
      // Without --rule: no-debugger is off, should pass
      const clean = await runRslint([], tempDir);
      expect(clean.stdout).not.toContain('no-debugger');

      // With --rule: override to error
      const result = await runRslint(['--rule', 'no-debugger: error'], tempDir);
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-debugger');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('--rule turns off a rule that was error in config', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': mjsConfig({ 'no-debugger': 'error' }),
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'debugger;\n',
    });

    try {
      // Without --rule: should report error
      const before = await runRslint([], tempDir);
      expect(before.stdout).toContain('no-debugger');

      // With --rule off: should pass
      const result = await runRslint(['--rule', 'no-debugger: off'], tempDir);
      expect(result.stdout).not.toContain('no-debugger');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('--rule changes severity from error to warn', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': mjsConfig({ 'no-debugger': 'error' }),
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'debugger;\n',
    });

    try {
      const result = await runRslint(
        ['--rule', 'no-debugger: warn', '--format', 'github'],
        tempDir,
      );
      // warn-only should exit 0, not error
      expect(result.exitCode).toBe(0);
      expect(result.stdout).toContain('::warning');
      expect(result.stdout).not.toContain('::error');
      expect(result.stdout).toContain('no-debugger');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -----------------------------------------------------------------------
  // Multiple --rule flags
  // -----------------------------------------------------------------------

  test('multiple --rule flags override multiple rules', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': mjsConfig({
        'no-debugger': 'off',
        'no-var': 'off',
      }),
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'debugger;\nexport var x = 1;\n',
    });

    try {
      const result = await runRslint(
        ['--rule', 'no-debugger: error', '--rule', 'no-var: error'],
        tempDir,
      );
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-debugger');
      expect(result.stdout).toContain('no-var');
      expect(result.stdout).toContain('2 error');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('later --rule overrides earlier --rule for same rule', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': mjsConfig({}),
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'debugger;\n',
    });

    try {
      // First --rule enables it, second --rule turns it off
      const result = await runRslint(
        ['--rule', 'no-debugger: error', '--rule', 'no-debugger: off'],
        tempDir,
      );
      expect(result.exitCode).toBe(0);
      expect(result.stdout).not.toContain('no-debugger');
      expect(result.stdout).toContain('0 error');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -----------------------------------------------------------------------
  // Argument ordering matrix: --rule, files, and other flags in every position
  // -----------------------------------------------------------------------

  // Helper: all ordering tests share the same fixture
  const orderingFixture = {
    'rslint.config.mjs': mjsConfig({ 'no-debugger': 'off' }),
    'tsconfig.json': TS_CONFIG,
    'test.ts': 'debugger;\n',
  };

  // Assert for every ordering variant: no-debugger fires as error with exit 1.
  // Does not check summary format since some callers use --format github.
  async function expectNoDebuggerError(args: string[]): Promise<void> {
    const tempDir = await createTempDir(orderingFixture);
    try {
      const result = await runRslint(args, tempDir);
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-debugger');
    } finally {
      await cleanupTempDir(tempDir);
    }
  }

  test('ordering: --rule before file', async () => {
    await expectNoDebuggerError(['--rule', 'no-debugger: error', 'test.ts']);
  });

  test('ordering: file before --rule', async () => {
    await expectNoDebuggerError(['test.ts', '--rule', 'no-debugger: error']);
  });

  test('ordering: --rule between two files', async () => {
    const tempDir = await createTempDir({
      ...orderingFixture,
      'other.ts': 'debugger;\n',
    });
    try {
      const result = await runRslint(
        ['test.ts', '--rule', 'no-debugger: error', 'other.ts'],
        tempDir,
      );
      expect(result.stdout).toContain('no-debugger');
      // Both files should be linted
      expect(result.stdout).toContain('linted 2 file');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('ordering: --format before --rule before file', async () => {
    await expectNoDebuggerError([
      '--format',
      'github',
      '--rule',
      'no-debugger: error',
      'test.ts',
    ]);
  });

  test('ordering: file before --format before --rule', async () => {
    await expectNoDebuggerError([
      'test.ts',
      '--format',
      'default',
      '--rule',
      'no-debugger: error',
    ]);
  });

  test('ordering: --rule before file before --format', async () => {
    const tempDir = await createTempDir(orderingFixture);
    try {
      const result = await runRslint(
        ['--rule', 'no-debugger: error', 'test.ts', '--format', 'github'],
        tempDir,
      );
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('::error');
      expect(result.stdout).toContain('title=no-debugger');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('ordering: --quiet before file before --rule', async () => {
    await expectNoDebuggerError([
      '--quiet',
      'test.ts',
      '--rule',
      'no-debugger: error',
    ]);
  });

  test('ordering: file before --rule before --quiet', async () => {
    await expectNoDebuggerError([
      'test.ts',
      '--rule',
      'no-debugger: error',
      '--quiet',
    ]);
  });

  test('ordering: all flags after file', async () => {
    const tempDir = await createTempDir(orderingFixture);
    try {
      const result = await runRslint(
        [
          'test.ts',
          '--rule',
          'no-debugger: error',
          '--format',
          'github',
          '--quiet',
        ],
        tempDir,
      );
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('::error');
      expect(result.stdout).toContain('title=no-debugger');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('ordering: all flags before file', async () => {
    const tempDir = await createTempDir(orderingFixture);
    try {
      const result = await runRslint(
        [
          '--rule',
          'no-debugger: error',
          '--format',
          'github',
          '--quiet',
          'test.ts',
        ],
        tempDir,
      );
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('::error');
      expect(result.stdout).toContain('title=no-debugger');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('ordering: multiple --rule flags interleaved with files', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': mjsConfig({
        'no-debugger': 'off',
        'no-var': 'off',
      }),
      'tsconfig.json': TS_CONFIG,
      'a.ts': 'debugger;\n',
      'b.ts': 'export var x = 1;\n',
    });
    try {
      const result = await runRslint(
        [
          '--rule',
          'no-debugger: error',
          'a.ts',
          '--rule',
          'no-var: error',
          'b.ts',
        ],
        tempDir,
      );
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-debugger');
      expect(result.stdout).toContain('no-var');
      expect(result.stdout).toContain('2 error');
      expect(result.stdout).toContain('linted 2 file');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('ordering: --rule=value (equals) mixed with files and flags', async () => {
    const tempDir = await createTempDir(orderingFixture);
    try {
      const result = await runRslint(
        ['test.ts', '--rule=no-debugger: error', '--format', 'github'],
        tempDir,
      );
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('::error');
      expect(result.stdout).toContain('title=no-debugger');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('ordering: --rule=value and --rule space mixed', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': mjsConfig({
        'no-debugger': 'off',
        'no-var': 'off',
      }),
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'debugger;\nexport var x = 1;\n',
    });
    try {
      const result = await runRslint(
        ['--rule=no-debugger: error', 'test.ts', '--rule', 'no-var: error'],
        tempDir,
      );
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-debugger');
      expect(result.stdout).toContain('no-var');
      expect(result.stdout).toContain('2 error');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('ordering: --config before --rule before file', async () => {
    const tempDir = await createTempDir({
      'custom.config.mjs': mjsConfig({ 'no-debugger': 'off' }),
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'debugger;\n',
    });
    try {
      const result = await runRslint(
        [
          '--config',
          'custom.config.mjs',
          '--rule',
          'no-debugger: error',
          'test.ts',
        ],
        tempDir,
      );
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-debugger');
      expect(result.stdout).toContain('1 error');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('ordering: file before --config before --rule', async () => {
    const tempDir = await createTempDir({
      'custom.config.mjs': mjsConfig({ 'no-debugger': 'off' }),
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'debugger;\n',
    });
    try {
      const result = await runRslint(
        [
          'test.ts',
          '--config',
          'custom.config.mjs',
          '--rule',
          'no-debugger: error',
        ],
        tempDir,
      );
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-debugger');
      expect(result.stdout).toContain('1 error');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -----------------------------------------------------------------------
  // Multiple --rule flags interleaved with other flags
  // -----------------------------------------------------------------------

  test('ordering: multiple --rule with --format between them', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': mjsConfig({
        'no-debugger': 'off',
        'no-var': 'off',
      }),
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'debugger;\nexport var x = 1;\n',
    });
    try {
      const result = await runRslint(
        [
          '--rule',
          'no-debugger: error',
          '--format',
          'github',
          '--rule',
          'no-var: error',
          'test.ts',
        ],
        tempDir,
      );
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('title=no-debugger');
      expect(result.stdout).toContain('title=no-var');
      // Both are errors in github format
      const lines = result.stdout.trim().split('\n');
      expect(lines.filter(l => l.startsWith('::error')).length).toBe(2);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('ordering: file, --rule, --quiet, --rule', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': mjsConfig({
        'no-debugger': 'off',
        'no-var': 'off',
      }),
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'debugger;\nexport var x = 1;\n',
    });
    try {
      const result = await runRslint(
        [
          'test.ts',
          '--rule',
          'no-debugger: error',
          '--quiet',
          '--rule',
          'no-var: error',
        ],
        tempDir,
      );
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-debugger');
      expect(result.stdout).toContain('no-var');
      expect(result.stdout).toContain('2 error');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('ordering: --rule, file, --format, file, --rule', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': mjsConfig({
        'no-debugger': 'off',
        'no-var': 'off',
      }),
      'tsconfig.json': TS_CONFIG,
      'a.ts': 'debugger;\n',
      'b.ts': 'export var x = 1;\n',
    });
    try {
      const result = await runRslint(
        [
          '--rule',
          'no-debugger: error',
          'a.ts',
          '--format',
          'github',
          'b.ts',
          '--rule',
          'no-var: error',
        ],
        tempDir,
      );
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('title=no-debugger');
      expect(result.stdout).toContain('title=no-var');
      const lines = result.stdout.trim().split('\n');
      expect(lines.filter(l => l.startsWith('::error')).length).toBe(2);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('ordering: multiple --rule with --max-warnings between them', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': mjsConfig({
        'no-debugger': 'off',
        'no-var': 'off',
      }),
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'debugger;\nexport var x = 1;\n',
    });
    try {
      // Both rules as warn, max-warnings=1 should cause exit code 1 (2 warnings > 1)
      const result = await runRslint(
        [
          '--rule',
          'no-debugger: warn',
          '--max-warnings',
          '1',
          '--rule',
          'no-var: warn',
          'test.ts',
        ],
        tempDir,
      );
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-debugger');
      expect(result.stdout).toContain('no-var');
      expect(result.stdout).toContain('0 error');
      expect(result.stdout).toContain('2 warning');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('ordering: --config, --rule, file, --rule, --format', async () => {
    const tempDir = await createTempDir({
      'custom.config.mjs': mjsConfig({
        'no-debugger': 'off',
        'no-var': 'off',
      }),
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'debugger;\nexport var x = 1;\n',
    });
    try {
      const result = await runRslint(
        [
          '--config',
          'custom.config.mjs',
          '--rule',
          'no-debugger: error',
          'test.ts',
          '--rule',
          'no-var: warn',
          '--format',
          'github',
        ],
        tempDir,
      );
      expect(result.exitCode).not.toBe(0);
      const lines = result.stdout.trim().split('\n');
      // Exactly 1 error (no-debugger) and 1 warning (no-var)
      const errors = lines.filter(l => l.startsWith('::error'));
      const warnings = lines.filter(l => l.startsWith('::warning'));
      expect(errors.length).toBe(1);
      expect(warnings.length).toBe(1);
      expect(errors[0]).toContain('title=no-debugger');
      expect(warnings[0]).toContain('title=no-var');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -----------------------------------------------------------------------
  // --rule with array options
  // -----------------------------------------------------------------------

  test('--rule with JSON array options', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': mjsConfig({ 'no-console': 'off' }),
      'tsconfig.json': TS_CONFIG,
      'test.ts': `
console.log("hello");
console.warn("warning");
console.error("error");
`,
    });

    try {
      // Allow warn and error, only console.log should trigger
      const result = await runRslint(
        ['--rule', 'no-console: ["error", {"allow": ["warn", "error"]}]'],
        tempDir,
      );
      expect(result.stdout).toContain('no-console');
      // Should report exactly 1 error (console.log), not 3
      expect(result.stdout).toContain('1 error');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -----------------------------------------------------------------------
  // --rule with plugin rules
  // -----------------------------------------------------------------------

  test('--rule with @typescript-eslint plugin rule', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': mjsConfig({
        '@typescript-eslint/no-explicit-any': 'off',
      }),
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'export const x: any = 1;\n',
    });

    try {
      const result = await runRslint(
        ['--rule', '@typescript-eslint/no-explicit-any: error'],
        tempDir,
      );
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-explicit-any');
      expect(result.stdout).toContain('1 error');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -----------------------------------------------------------------------
  // --rule whitespace variations
  // -----------------------------------------------------------------------

  test('--rule with no space after colon', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': mjsConfig({ 'no-debugger': 'off' }),
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'debugger;\n',
    });

    try {
      const result = await runRslint(['--rule', 'no-debugger:error'], tempDir);
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-debugger');
      expect(result.stdout).toContain('1 error');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('--rule with extra spaces around name and value', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': mjsConfig({ 'no-debugger': 'off' }),
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'debugger;\n',
    });

    try {
      const result = await runRslint(
        ['--rule', '  no-debugger  :  error  '],
        tempDir,
      );
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-debugger');
      expect(result.stdout).toContain('1 error');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('--rule with extra spaces inside JSON array options', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': mjsConfig({ 'no-console': 'off' }),
      'tsconfig.json': TS_CONFIG,
      'test.ts': `
console.log("hello");
console.warn("warning");
console.error("error");
`,
    });

    try {
      const result = await runRslint(
        [
          '--rule',
          'no-console:  ["error",  { "allow" :  ["warn",  "error"] }]',
        ],
        tempDir,
      );
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-console');
      // Only console.log triggers, console.warn and console.error are allowed
      expect(result.stdout).toContain('1 error');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -----------------------------------------------------------------------
  // --rule overrides per-file config for all files
  // -----------------------------------------------------------------------

  test('--rule overrides different per-file configs uniformly', async () => {
    // Config: no-debugger is error for .ts, off for .test.ts
    // CLI --rule should override both
    const config = mjsMultiConfig([
      {
        files: ['**/*.ts'],
        languageOptions: {
          parserOptions: {
            projectService: false,
            project: ['./tsconfig.json'],
          },
        },
        rules: { 'no-debugger': 'error' },
        plugins: ['@typescript-eslint'],
      },
      {
        files: ['**/*.test.ts'],
        rules: { 'no-debugger': 'off' },
      },
    ]);

    const tempDir = await createTempDir({
      'rslint.config.mjs': config,
      'tsconfig.json': TS_CONFIG,
      'src/app.ts': 'debugger;\n',
      'src/app.test.ts': 'debugger;\n',
    });

    try {
      // Without --rule: app.ts has error, app.test.ts has off
      const before = await runRslint([], tempDir);
      expect(before.stdout).toContain('app.ts');
      expect(before.stdout).not.toContain('app.test.ts');

      // With --rule off: both should be clean
      const result = await runRslint(['--rule', 'no-debugger: off'], tempDir);
      expect(result.stdout).not.toContain('no-debugger');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -----------------------------------------------------------------------
  // Error handling
  // -----------------------------------------------------------------------

  test('--rule with invalid format exits with error', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': mjsConfig({}),
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'const x = 1;\n',
    });

    try {
      const result = await runRslint(['--rule', 'no-colon-here'], tempDir);
      expect(result.exitCode).not.toBe(0);
      expect(result.stderr).toContain('invalid --rule format');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('--rule with invalid JSON array exits with error', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': mjsConfig({}),
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'const x = 1;\n',
    });

    try {
      const result = await runRslint(
        ['--rule', 'no-console: [broken'],
        tempDir,
      );
      expect(result.exitCode).not.toBe(0);
      expect(result.stderr).toContain('invalid --rule format');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -----------------------------------------------------------------------
  // --rule for non-existent rule (should silently skip)
  // -----------------------------------------------------------------------

  test('--rule with non-existent rule name is silently ignored', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': mjsConfig({}),
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'export const x = 1;\n',
    });

    try {
      const result = await runRslint(
        ['--rule', 'this-rule-does-not-exist: error'],
        tempDir,
      );
      // Should not crash — the non-existent rule is silently skipped
      expect(result.stdout).not.toContain('this-rule-does-not-exist');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -----------------------------------------------------------------------
  // -- option-terminator is preserved
  // -----------------------------------------------------------------------

  test('--rule before -- still works, positionals after -- treated as files', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': mjsConfig({ 'no-debugger': 'off' }),
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'debugger;\n',
    });

    try {
      // --rule is before --, so it's a flag; test.ts is after --, treated as file
      const result = await runRslint(
        ['--rule', 'no-debugger: error', '--', 'test.ts'],
        tempDir,
      );
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-debugger');
      expect(result.stdout).toContain('1 error');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -----------------------------------------------------------------------
  // --help shows --rule
  // -----------------------------------------------------------------------

  test('--help output includes --rule flag', async () => {
    const result = await runRslint(['--help']);
    expect(result.exitCode).toBe(0);
    expect(result.stderr).toContain('--rule');
  });
});

import { describe, test, expect } from '@rstest/core';
import {
  runRslint,
  createTempDir,
  cleanupTempDir,
  TS_CONFIG,
  makeConfig,
} from './helpers';

describe('--type-check basic behavior', () => {
  test('should not report type errors without --type-check', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': makeConfig(),
      'test.ts': "const x: number = 'hello';\n",
    });
    try {
      const result = await runRslint([], tempDir);
      expect(result.stdout).not.toContain('TS2322');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should exit with non-zero code when type errors exist', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': makeConfig(),
      'test.ts': "const x: number = 'hello';\n",
    });
    try {
      const result = await runRslint(['--type-check'], tempDir);
      expect(result.exitCode).not.toBe(0);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should exit with zero when code is type-safe', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': makeConfig(),
      'test.ts': 'const x: number = 42;\n',
    });
    try {
      const result = await runRslint(['--type-check'], tempDir);
      expect(result.exitCode).toBe(0);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('@ts-expect-error should suppress type errors', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': makeConfig(),
      'test.ts': "// @ts-expect-error\nconst x: number = 'hello';\n",
    });
    try {
      const result = await runRslint(['--type-check'], tempDir);
      expect(result.stdout).not.toContain('TS2322');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('unused @ts-expect-error should report TS2578', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': makeConfig(),
      'test.ts': '// @ts-expect-error\nconst x: number = 42;\n',
    });
    try {
      const result = await runRslint(['--type-check'], tempDir);
      expect(result.stdout).toContain('TS2578');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should show both lint errors and type errors', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': makeConfig({
        '@typescript-eslint/no-explicit-any': 'error',
      }),
      'test.ts':
        "let a: any = 'hello';\nconst x: number = a;\nconst y: number = 'world';\n",
    });
    try {
      const result = await runRslint(['--type-check'], tempDir);
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-explicit-any');
      expect(result.stdout).toContain('TS2322');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('--quiet should still show type errors but suppress warnings', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': makeConfig({
        '@typescript-eslint/no-explicit-any': 'warn',
      }),
      'test.ts': "let a: any = 1;\nconst x: number = 'hello';\n",
    });
    try {
      const result = await runRslint(['--type-check', '--quiet'], tempDir);
      expect(result.stdout).toContain('TS2322');
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should report multiple type errors across files', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': makeConfig(),
      'a.ts': "const x: number = 'hello';\n",
      'b.ts': 'const y: string = 42;\n',
    });
    try {
      const result = await runRslint(
        ['--type-check', '--format', 'jsonline'],
        tempDir,
      );
      const lines = result.stdout
        .trim()
        .split('\n')
        .filter(l => l.trim());
      const tsLines = lines.filter(l => {
        const parsed = JSON.parse(l);
        return parsed.ruleName?.startsWith('TypeScript(');
      });
      expect(tsLines.length).toBeGreaterThanOrEqual(2);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

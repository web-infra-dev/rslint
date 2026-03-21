import { describe, test, expect } from '@rstest/core';
import {
  runRslint,
  createTempDir,
  cleanupTempDir,
  TS_CONFIG,
  jsConfig,
} from './helpers.js';

describe('CLI config discovery (upward traversal)', () => {
  test('should find config in parent directory', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.js': jsConfig(),
      'child/test.ts': 'let a: any = 10;\na.b = 20;\n',
    });
    try {
      // Run from child/ — config is in parent
      const result = await runRslint(['test.ts'], `${tempDir}/child`);
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-unsafe-member-access');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should find config in grandparent directory', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.js': jsConfig(),
      'a/b/test.ts': 'let a: any = 10;\na.b = 20;\n',
    });
    try {
      const result = await runRslint(['test.ts'], `${tempDir}/a/b`);
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-unsafe-member-access');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should use nearest config when multiple exist', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      // Parent config enables no-explicit-any
      'rslint.config.js': jsConfig({
        rules: {
          '@typescript-eslint/no-explicit-any': 'error',
          '@typescript-eslint/no-unsafe-member-access': 'off',
        },
      }),
      // Child has its own tsconfig and config
      'child/tsconfig.json': TS_CONFIG,
      'child/rslint.config.js': jsConfig({
        rules: {
          '@typescript-eslint/no-unsafe-member-access': 'error',
          '@typescript-eslint/no-explicit-any': 'off',
        },
      }),
      'child/test.ts': 'let a: any = 10;\na.b = 20;\n',
    });
    try {
      // From child/ — should use child's config (no-unsafe-member-access on, no-explicit-any off)
      const result = await runRslint(
        ['--format', 'jsonline', 'test.ts'],
        `${tempDir}/child`,
      );
      expect(result.stdout).toContain('no-unsafe-member-access');
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('no args should scope linting to CWD', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.js': jsConfig(),
      'child/test.ts': 'let a: any = 10;\na.b = 20;\n',
      'sibling.ts': 'let b: any = 10;\nb.c = 20;\n',
    });
    try {
      // Run from child/ with no args — should only lint child/test.ts, not sibling.ts
      const result = await runRslint(
        ['--format', 'jsonline'],
        `${tempDir}/child`,
      );
      expect(result.stdout).toContain('test.ts');
      expect(result.stdout).not.toContain('sibling.ts');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('no args from monorepo root should discover sub-package configs', async () => {
    const tempDir = await createTempDir({
      // Root config: no-explicit-any
      'tsconfig.json': TS_CONFIG,
      'rslint.config.js': jsConfig({
        rules: {
          '@typescript-eslint/no-explicit-any': 'error',
          '@typescript-eslint/no-unsafe-member-access': 'off',
        },
      }),
      // Sub-package with its own config: no-unsafe-member-access
      'packages/foo/tsconfig.json': TS_CONFIG,
      'packages/foo/rslint.config.js': jsConfig({
        rules: {
          '@typescript-eslint/no-unsafe-member-access': 'error',
          '@typescript-eslint/no-explicit-any': 'off',
        },
      }),
      'packages/foo/src/a.ts': 'let a: any = 10;\na.b = 20;\n',
      // File at root level (uses root config)
      'root.ts': 'let b: any = 10;\nb.c = 20;\n',
    });
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      const lines = result.stdout
        .trim()
        .split('\n')
        .filter((l: string) => l.trim());
      const diags = lines.map((l: string) => JSON.parse(l));

      // foo/src/a.ts should use foo's config (no-unsafe-member-access)
      const fooDiags = diags.filter((d: { filePath: string }) =>
        d.filePath.includes('foo'),
      );
      expect(
        fooDiags.some(
          (d: { ruleName: string }) =>
            d.ruleName === '@typescript-eslint/no-unsafe-member-access',
        ),
      ).toBe(true);

      // root.ts should use root config (no-explicit-any)
      const rootDiags = diags.filter(
        (d: { filePath: string }) => d.filePath === 'root.ts',
      );
      expect(
        rootDiags.some(
          (d: { ruleName: string }) =>
            d.ruleName === '@typescript-eslint/no-explicit-any',
        ),
      ).toBe(true);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('broken sub-package config should be skipped in multi-config', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.js': jsConfig(),
      // Sub-package with intentionally broken config
      'packages/broken/rslint.config.js':
        'export default [INVALID SYNTAX HERE;',
      'packages/broken/tsconfig.json': TS_CONFIG,
      'packages/broken/src/a.ts': 'let a: any = 10;\na.b = 20;\n',
      // File at root level
      'root.ts': 'let b: any = 10;\nb.c = 20;\n',
    });
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      // Should still lint root.ts with root config, skipping broken package
      expect(result.stdout).toContain('root.ts');
      expect(result.stderr).toContain('Warning: skipping config');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('--config should override automatic discovery', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      // Auto-discovered config has no-explicit-any
      'rslint.config.js': jsConfig({
        rules: {
          '@typescript-eslint/no-explicit-any': 'error',
          '@typescript-eslint/no-unsafe-member-access': 'off',
        },
      }),
      // Explicit config has no-unsafe-member-access
      'custom.config.js': jsConfig({
        rules: {
          '@typescript-eslint/no-unsafe-member-access': 'error',
          '@typescript-eslint/no-explicit-any': 'off',
        },
      }),
      'test.ts': 'let a: any = 10;\na.b = 20;\n',
    });
    try {
      const result = await runRslint(
        ['--config', 'custom.config.js', '--format', 'jsonline', 'test.ts'],
        tempDir,
      );
      // Should use custom config, not auto-discovered
      expect(result.stdout).toContain('no-unsafe-member-access');
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

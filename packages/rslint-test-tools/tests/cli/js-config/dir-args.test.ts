import { describe, test, expect } from '@rstest/core';
import path from 'node:path';
import {
  runRslint,
  createTempDir,
  cleanupTempDir,
  TS_CONFIG,
  jsConfig,
} from './helpers.js';

describe('CLI directory arguments', () => {
  test('should lint only files under specified directory', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.js': jsConfig(),
      'src/error.ts': 'let a: any = 10;\na.b = 20;\n',
      'lib/error.ts': 'let b: any = 10;\nb.c = 20;\n',
    });
    try {
      const result = await runRslint(['--format', 'jsonline', 'src/'], tempDir);
      expect(result.stdout).toContain('src/error.ts');
      expect(result.stdout).not.toContain('lib/error.ts');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should scope to directory even with broad config files pattern', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      // Config matches all .ts files
      'rslint.config.js': jsConfig(),
      'src/a.ts': 'let a: any = 10;\na.b = 20;\n',
      'lib/b.ts': 'let b: any = 10;\nb.c = 20;\n',
      'root.ts': 'let c: any = 10;\nc.d = 20;\n',
    });
    try {
      // Only src/ should be linted
      const result = await runRslint(['--format', 'jsonline', 'src/'], tempDir);
      expect(result.stdout).toContain('src/a.ts');
      expect(result.stdout).not.toContain('lib/b.ts');
      expect(result.stdout).not.toContain('root.ts');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should support mixed file and directory arguments', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.js': jsConfig(),
      'src/a.ts': 'let a: any = 10;\na.b = 20;\n',
      'lib/b.ts': 'let b: any = 10;\nb.c = 20;\n',
      'other.ts': 'let c: any = 10;\nc.d = 20;\n',
    });
    try {
      // Lint src/ directory + lib/b.ts file, but not other.ts
      const result = await runRslint(
        ['--format', 'jsonline', 'src/', 'lib/b.ts'],
        tempDir,
      );
      expect(result.stdout).toContain('src/a.ts');
      expect(result.stdout).toContain('lib/b.ts');
      expect(result.stdout).not.toContain('other.ts');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should not match similarly-named directories', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.js': jsConfig(),
      'src/a.ts': 'let a: any = 10;\na.b = 20;\n',
      'src-other/b.ts': 'let b: any = 10;\nb.c = 20;\n',
    });
    try {
      // src/ should NOT match src-other/
      const result = await runRslint(['--format', 'jsonline', 'src/'], tempDir);
      expect(result.stdout).toContain('src/a.ts');
      expect(result.stdout).not.toContain('src-other');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should find config by walking up from directory arg', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.js': jsConfig(),
      'src/a.ts': 'let a: any = 10;\na.b = 20;\n',
    });
    try {
      // Config is in parent, dir arg is src/ — should find config by walking up
      const result = await runRslint(['src/'], tempDir);
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-unsafe-member-access');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should work with --fix and directory argument', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.js': jsConfig({
        rules: {
          '@typescript-eslint/no-inferrable-types': 'error',
          '@typescript-eslint/no-unsafe-member-access': 'off',
        },
      }),
      'src/fixable.ts': 'const x: number = 42;\n',
      'lib/fixable.ts': 'const y: string = "hello";\n',
    });
    try {
      // Fix only src/ — lib/ should not be touched
      await runRslint(['--fix', 'src/'], tempDir);

      const { readFile } = await import('node:fs/promises');
      const srcContent = await readFile(
        path.join(tempDir, 'src/fixable.ts'),
        'utf8',
      );
      expect(srcContent).not.toContain(': number');

      const libContent = await readFile(
        path.join(tempDir, 'lib/fixable.ts'),
        'utf8',
      );
      expect(libContent).toContain(': string');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should warn when directory has no matching files', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.js': jsConfig(),
      'src/a.ts': 'export const x = 1;\n',
    });
    try {
      // empty-dir/ doesn't exist in tsconfig — no files match, warn and exit 0
      const result = await runRslint(['empty-dir/'], tempDir);
      expect(result.exitCode).toBe(0);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

describe('CLI multi-config (monorepo)', () => {
  test('should use different configs for files in different packages', async () => {
    const tempDir = await createTempDir({
      // foo package: config with no-explicit-any
      'packages/foo/tsconfig.json': TS_CONFIG,
      'packages/foo/rslint.config.js': jsConfig({
        rules: {
          '@typescript-eslint/no-explicit-any': 'error',
          '@typescript-eslint/no-unsafe-member-access': 'off',
        },
      }),
      'packages/foo/src/a.ts': 'let a: any = 10;\na.b = 20;\n',
      // bar package: config with no-unsafe-member-access
      'packages/bar/tsconfig.json': TS_CONFIG,
      'packages/bar/rslint.config.js': jsConfig({
        rules: {
          '@typescript-eslint/no-unsafe-member-access': 'error',
          '@typescript-eslint/no-explicit-any': 'off',
        },
      }),
      'packages/bar/src/b.ts': 'let b: any = 10;\nb.c = 20;\n',
    });
    try {
      const result = await runRslint(
        [
          '--format',
          'jsonline',
          'packages/foo/src/a.ts',
          'packages/bar/src/b.ts',
        ],
        tempDir,
      );

      // Parse jsonline output
      const lines = result.stdout
        .trim()
        .split('\n')
        .filter((l: string) => l.trim());
      const diags = lines.map((l: string) => JSON.parse(l));

      // foo/a.ts should have no-explicit-any (from foo's config)
      const fooDiags = diags.filter((d: { filePath: string }) =>
        d.filePath.includes('foo'),
      );
      expect(
        fooDiags.some(
          (d: { ruleName: string }) =>
            d.ruleName === '@typescript-eslint/no-explicit-any',
        ),
      ).toBe(true);
      expect(
        fooDiags.some(
          (d: { ruleName: string }) =>
            d.ruleName === '@typescript-eslint/no-unsafe-member-access',
        ),
      ).toBe(false);

      // bar/b.ts should have no-unsafe-member-access (from bar's config)
      const barDiags = diags.filter((d: { filePath: string }) =>
        d.filePath.includes('bar'),
      );
      expect(
        barDiags.some(
          (d: { ruleName: string }) =>
            d.ruleName === '@typescript-eslint/no-unsafe-member-access',
        ),
      ).toBe(true);
      expect(
        barDiags.some(
          (d: { ruleName: string }) =>
            d.ruleName === '@typescript-eslint/no-explicit-any',
        ),
      ).toBe(false);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should deduplicate when multiple files find same config', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.js': jsConfig(),
      'src/a.ts': 'let a: any = 10;\na.b = 20;\n',
      'src/b.ts': 'let b: any = 10;\nb.c = 20;\n',
    });
    try {
      // Both files find the same config — should work without issues
      const result = await runRslint(['src/a.ts', 'src/b.ts'], tempDir);
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('linted 2 files');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

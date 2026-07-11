import { describe, test, expect } from '@rstest/core';
import fs from 'node:fs';
import path from 'node:path';
import {
  runRslint,
  createTempDir,
  cleanupTempDir,
  TS_CONFIG,
  jsConfig,
} from './helpers.js';

describe('CLI config discovery (upward traversal)', () => {
  test('invalid empty files array should fail before linting', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': `export default [
        {
          files: [],
          rules: { 'no-debugger': 'error' },
        },
      ];`,
      'src/index.js': 'debugger;\n',
    });
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);

      expect(result.exitCode).toBe(1);
      expect(result.stderr).toContain('invalid config');
      expect(result.stderr).toContain('"files" must be a non-empty array');
      expect(result.stdout).not.toContain('no-debugger');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('invalid files value should fail before linting', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': `export default [
        {
          files: '**/*.js',
          rules: { 'no-debugger': 'error' },
        },
      ];`,
      'src/index.js': 'debugger;\n',
    });
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);

      expect(result.exitCode).toBe(1);
      expect(result.stderr).toContain('invalid config');
      expect(result.stderr).toContain('"files" must be an array');
      expect(result.stdout).not.toContain('no-debugger');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

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

  test('directory discovery uses js config priority within the same directory', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.js': `export default [{
        files: ['**/*.ts'],
        rules: { 'no-console': 'error' },
      }];`,
      'rslint.config.mjs': `export default [{
        files: ['**/*.ts'],
        rules: { 'no-debugger': 'error' },
      }];`,
      'test.ts': `console.log('x');\ndebugger;\n`,
    });
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      expect(result.stdout).toContain('no-console');
      expect(result.stdout).not.toContain('no-debugger');
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

  test('explicit file falls back from a broken nearest config to its ancestor', async () => {
    const tempDir = await createTempDir({
      'rslint.config.js': `export default [{
        files: ['**/*.ts'],
        rules: { 'no-debugger': 'error' },
      }];`,
      'packages/broken/rslint.config.js':
        'export default [INVALID CHILD CONFIG;',
      'packages/broken/index.ts': 'debugger;\n',
    });
    try {
      const result = await runRslint(
        ['--format', 'jsonline', 'packages/broken/index.ts'],
        tempDir,
      );
      expect(result.stdout).toContain('no-debugger');
      expect(result.stderr).toContain('Warning: skipping config');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('explicit file symlink keeps lexical config ownership', async () => {
    const tempDir = await createTempDir({
      'physical/rslint.config.js': `export default [{
        files: ['**/*.ts'],
        rules: { 'no-debugger': 'error' },
      }];`,
      'physical/index.ts': 'console.log("value");\ndebugger;\n',
      'lexical/rslint.config.js': `export default [{
        files: ['**/*.ts'],
        rules: { 'no-console': 'error' },
      }];`,
    });
    try {
      const lexicalFile = path.join(tempDir, 'lexical/index.ts');
      try {
        fs.symlinkSync(path.join(tempDir, 'physical/index.ts'), lexicalFile);
      } catch {
        return;
      }

      const result = await runRslint(
        ['--format', 'jsonline', 'lexical/index.ts'],
        tempDir,
      );
      expect(result.stdout).toContain('no-console');
      expect(result.stdout).not.toContain('no-debugger');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('file selector matching is independent from symlink target Program membership', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': `export default [{
        files: ['link.ts'],
        languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
        rules: { 'no-console': 'error' },
      }];`,
      'tsconfig.json': JSON.stringify({ files: ['physical/index.ts'] }),
      'physical/index.ts': 'console.log("value");\n',
    });
    try {
      try {
        fs.symlinkSync(
          path.join(tempDir, 'physical/index.ts'),
          path.join(tempDir, 'link.ts'),
        );
      } catch {
        return;
      }

      const result = await runRslint(
        ['--format', 'jsonline', 'link.ts'],
        tempDir,
      );
      expect(result.stdout).toContain('no-console');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('explicit file ownership remains lexical when its physical config is also loaded', async () => {
    const tempDir = await createTempDir({
      'rslint.config.js': `export default [{
        files: ['**/*.ts'],
        rules: { 'no-console': 'error' },
      }];`,
      'packages/app/rslint.config.js': `export default [{
        files: ['**/*.ts'],
        rules: { 'no-debugger': 'error' },
      }];`,
      'packages/app/index.ts': 'console.log("value");\ndebugger;\n',
      'packages/app/other.ts': 'debugger;\n',
    });
    try {
      const lexicalFile = path.join(tempDir, 'link.ts');
      try {
        fs.symlinkSync(
          path.join(tempDir, 'packages/app/index.ts'),
          lexicalFile,
        );
      } catch {
        return;
      }

      const result = await runRslint(
        ['--format', 'jsonline', 'link.ts', 'packages/app/other.ts'],
        tempDir,
      );
      const diagnostics = result.stdout
        .trim()
        .split('\n')
        .filter(Boolean)
        .map(
          (line) => JSON.parse(line) as { filePath: string; ruleName: string },
        );
      expect(
        diagnostics.some(
          (diagnostic) =>
            diagnostic.filePath === 'link.ts' &&
            diagnostic.ruleName === 'no-console',
        ),
      ).toBe(true);
      expect(
        diagnostics.some(
          (diagnostic) =>
            diagnostic.filePath === 'link.ts' &&
            diagnostic.ruleName === 'no-debugger',
        ),
      ).toBe(false);
      expect(
        diagnostics.some(
          (diagnostic) =>
            diagnostic.filePath.includes('other.ts') &&
            diagnostic.ruleName === 'no-debugger',
        ),
      ).toBe(true);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('case aliases load one physical config once', async () => {
    const tempDir = await createTempDir({
      'Project/rslint.config.mjs': `
        import fs from 'node:fs';
        fs.appendFileSync(new URL('./loads.txt', import.meta.url), 'x');
        export default [{ rules: { 'no-debugger': 'error' } }];
      `,
      'Project/a.ts': 'debugger;\n',
      'Project/b.ts': 'debugger;\n',
    });
    try {
      const upperDir = path.join(tempDir, 'Project');
      const lowerDir = path.join(tempDir, 'project');
      if (!fs.existsSync(lowerDir)) {
        return;
      }

      const result = await runRslint(
        ['--format', 'jsonline', 'Project/a.ts', 'project/b.ts'],
        tempDir,
      );
      expect(result.stderr).not.toContain('duplicate config directories');
      expect(result.stdout).toContain('a.ts');
      expect(result.stdout).toContain('b.ts');
      expect(fs.readFileSync(path.join(upperDir, 'loads.txt'), 'utf8')).toBe(
        'x',
      );
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('broken config fallback reuses an already loaded native case alias', async () => {
    const tempDir = await createTempDir({
      'Project/rslint.config.mjs': `
        import fs from 'node:fs';
        fs.appendFileSync(new URL('./loads.txt', import.meta.url), 'x');
        export default [{ rules: { 'no-debugger': 'error' } }];
      `,
      'Project/a.ts': 'debugger;\n',
      'Project/broken/rslint.config.mjs':
        'export default [INVALID CHILD CONFIG;',
      'Project/broken/b.ts': 'debugger;\n',
    });
    try {
      const upperDir = path.join(tempDir, 'Project');
      const lowerDir = path.join(tempDir, 'project');
      if (!fs.existsSync(lowerDir)) {
        return;
      }

      const result = await runRslint(
        ['--format', 'jsonline', 'Project/a.ts', 'project/broken/b.ts'],
        tempDir,
      );
      expect(result.exitCode).toBe(1);
      expect(result.stderr).toContain('Warning: skipping config');
      expect(result.stderr).not.toContain('same filesystem location');
      expect(result.stdout).toContain('a.ts');
      expect(result.stdout).toContain('b.ts');
      expect(fs.readFileSync(path.join(upperDir, 'loads.txt'), 'utf8')).toBe(
        'x',
      );
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('directory aliases of one physical config root are rejected', async () => {
    const tempDir = await createTempDir({
      'shared/rslint.config.mjs': `export default [{
        rules: { 'no-debugger': 'error' },
      }];`,
      'shared/a.ts': 'debugger;\n',
      'shared/b.ts': 'debugger;\n',
    });
    try {
      try {
        fs.symlinkSync(
          path.join(tempDir, 'shared'),
          path.join(tempDir, 'owner-a'),
          'dir',
        );
        fs.symlinkSync(
          path.join(tempDir, 'shared'),
          path.join(tempDir, 'owner-b'),
          'dir',
        );
      } catch {
        return;
      }

      const result = await runRslint(
        ['--format', 'jsonline', 'owner-a/a.ts', 'owner-b/b.ts'],
        tempDir,
      );
      expect(result.exitCode).toBe(1);
      expect(result.stderr).toContain(
        'resolve to the same filesystem location',
      );
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('alternate casing of one directory symlink remains one config owner', async () => {
    const tempDir = await createTempDir({
      'real/rslint.config.mjs': `export default [{
        rules: { 'no-debugger': 'error' },
      }];`,
      'real/a.ts': 'debugger;\n',
      'real/b.ts': 'debugger;\n',
    });
    try {
      const upperDir = path.join(tempDir, 'Project');
      const lowerDir = path.join(tempDir, 'project');
      try {
        fs.symlinkSync(path.join(tempDir, 'real'), upperDir, 'dir');
      } catch {
        return;
      }
      if (!fs.existsSync(lowerDir)) {
        return;
      }

      const result = await runRslint(
        ['--format', 'jsonline', 'Project/a.ts', 'project/b.ts'],
        tempDir,
      );
      expect(result.stderr).not.toContain('same filesystem location');
      expect(result.stdout).toContain('a.ts');
      expect(result.stdout).toContain('b.ts');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('directory symlink uses its physical config scope', async () => {
    const tempDir = await createTempDir({
      'real/rslint.config.mjs': `export default [{
        rules: { 'no-debugger': 'error' },
      }];`,
      'real/sub/index.ts': 'debugger;\n',
    });
    try {
      const linkDir = path.join(tempDir, 'link');
      try {
        fs.symlinkSync(path.join(tempDir, 'real/sub'), linkDir, 'dir');
      } catch {
        return;
      }

      const result = await runRslint(['--format', 'jsonline', 'link'], tempDir);
      expect(result.stdout).toContain('no-debugger');
      const diagnostic = JSON.parse(result.stdout.trim().split('\n')[0]) as {
        filePath: string;
      };
      expect(diagnostic.filePath.split(path.sep).join('/')).toMatch(
        /(^|\/)link\/index\.ts$/,
      );
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('explicit directory ownership remains lexical when its physical config is also loaded', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': `export default [{
        rules: { 'no-console': 'error' },
      }];`,
      'real/rslint.config.mjs': `export default [{
        rules: { 'no-debugger': 'error' },
      }];`,
      'real/sub/index.ts': 'console.log("value");\ndebugger;\n',
      'real/sub/nested/rslint.config.mjs': `export default [{
        rules: { 'no-alert': 'error' },
      }];`,
      'real/sub/nested/child.ts': 'alert("value");\ndebugger;\n',
      'real/other.ts': 'debugger;\n',
    });
    try {
      try {
        fs.symlinkSync(
          path.join(tempDir, 'real/sub'),
          path.join(tempDir, 'link'),
          'dir',
        );
      } catch {
        return;
      }

      const result = await runRslint(
        ['--format', 'jsonline', 'link', 'real/other.ts'],
        tempDir,
      );
      const diagnostics = result.stdout
        .trim()
        .split('\n')
        .filter(Boolean)
        .map(
          (line) => JSON.parse(line) as { filePath: string; ruleName: string },
        );
      expect(
        diagnostics.some(
          ({ filePath, ruleName }) =>
            filePath.split(path.sep).join('/').endsWith('link/index.ts') &&
            ruleName === 'no-console',
        ),
      ).toBe(true);
      expect(
        diagnostics.some(
          ({ filePath, ruleName }) =>
            filePath.split(path.sep).join('/').endsWith('link/index.ts') &&
            ruleName === 'no-debugger',
        ),
      ).toBe(false);
      expect(
        diagnostics.some(
          ({ filePath, ruleName }) =>
            filePath.split(path.sep).join('/').endsWith('real/other.ts') &&
            ruleName === 'no-debugger',
        ),
      ).toBe(true);
      expect(
        diagnostics.some(
          ({ filePath, ruleName }) =>
            filePath
              .split(path.sep)
              .join('/')
              .endsWith('link/nested/child.ts') && ruleName === 'no-alert',
        ),
      ).toBe(true);
      expect(
        diagnostics.some(
          ({ filePath, ruleName }) =>
            filePath
              .split(path.sep)
              .join('/')
              .endsWith('link/nested/child.ts') && ruleName === 'no-debugger',
        ),
      ).toBe(false);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('two lexical owners for one physical file are rejected', async () => {
    const tempDir = await createTempDir({
      'shared.ts': 'debugger;\n',
      'owner-a/rslint.config.js': `export default [{
        files: ['**/*.ts'],
        rules: { 'no-debugger': 'error' },
      }];`,
      'owner-b/rslint.config.js': `export default [{
        files: ['**/*.ts'],
        rules: { 'no-debugger': 'error' },
      }];`,
    });
    try {
      const sharedFile = path.join(tempDir, 'shared.ts');
      try {
        fs.symlinkSync(sharedFile, path.join(tempDir, 'owner-a/index.ts'));
        fs.symlinkSync(sharedFile, path.join(tempDir, 'owner-b/index.ts'));
      } catch {
        return;
      }

      const result = await runRslint(
        ['--format', 'jsonline', 'owner-a/index.ts', 'owner-b/index.ts'],
        tempDir,
      );
      expect(result.exitCode).toBe(1);
      expect(result.stderr).toContain('governed by different configs');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('all broken JS configs should not fall back to JSON config', async () => {
    const tempDir = await createTempDir({
      'rslint.json': JSON.stringify([
        {
          files: ['**/*.ts'],
          rules: { 'no-debugger': 'error' },
        },
      ]),
      'rslint.config.js': 'export default [INVALID ROOT CONFIG;',
      'packages/broken/rslint.config.js':
        'export default [INVALID CHILD CONFIG;',
      'root.ts': 'debugger;\n',
      'packages/broken/src/a.ts': 'debugger;\n',
    });
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      expect(result.exitCode).toBe(1);
      expect(result.stdout).not.toContain('no-debugger');
      expect(result.stderr).toContain('Warning: skipping config');
      expect(result.stderr).not.toContain('JSON configuration is deprecated');
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

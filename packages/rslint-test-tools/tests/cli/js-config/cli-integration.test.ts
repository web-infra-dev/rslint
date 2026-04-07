import { describe, test, expect } from '@rstest/core';
import fs from 'node:fs/promises';
import {
  runRslint,
  createTempDir,
  cleanupTempDir,
  TS_CONFIG,
  jsConfig,
} from './helpers.js';

describe('CLI JS config integration', () => {
  test('should auto-detect rslint.config.js', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'let a: any = 10;\na.b = 20;\n',
      'rslint.config.js': jsConfig(),
    });
    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-unsafe-member-access');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should support --config=path (equals format) for JS config', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'let a: any = 10;\na.b = 20;\n',
      'custom.config.js': jsConfig(),
    });
    try {
      const result = await runRslint(['--config=custom.config.js'], tempDir);
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-unsafe-member-access');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should support --config path (space format) for JS config', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'let a: any = 10;\na.b = 20;\n',
      'custom.config.js': jsConfig(),
    });
    try {
      const result = await runRslint(['--config', 'custom.config.js'], tempDir);
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-unsafe-member-access');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should pass through unknown flags to Go binary', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'let a: any = 10;\na.b = 20;\n',
      'rslint.config.js': jsConfig(),
    });
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      const lines = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim());
      for (const line of lines) {
        expect(() => JSON.parse(line)).not.toThrow();
      }
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should report error for non-existent JS config via --config', async () => {
    const tempDir = await createTempDir({ 'test.ts': 'const a = 1;\n' });
    try {
      const result = await runRslint(
        ['--config', 'nonexistent.config.js'],
        tempDir,
      );
      expect(result.exitCode).not.toBe(0);
      expect(result.stderr).toContain('not found');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should prefer JS config over JSON config', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'let a: any = 10;\na.b = 20;\n',
      'rslint.json': JSON.stringify([
        {
          files: ['**/*.ts'],
          languageOptions: {
            parserOptions: {
              projectService: false,
              project: ['./tsconfig.json'],
            },
          },
          rules: { '@typescript-eslint/no-explicit-any': 'error' },
          plugins: ['@typescript-eslint'],
        },
      ]),
      'rslint.config.js': jsConfig({
        rules: {
          '@typescript-eslint/no-unsafe-member-access': 'error',
          '@typescript-eslint/no-explicit-any': 'off',
        },
      }),
    });
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      expect(result.stdout).toContain('no-unsafe-member-access');
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should handle JS config with global ignores', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': JSON.stringify({
        compilerOptions: {
          target: 'ES2020',
          module: 'ESNext',
          strict: true,
        },
        include: ['src/**/*.ts', 'dist/**/*.ts'],
      }),
      'src/index.ts': 'let a: any = 10;\na.b = 20;\n',
      'dist/index.ts': 'let a: any = 10;\na.b = 20;\n',
      'rslint.config.js': `export default [
        { ignores: ["dist/**"] },
        ${JSON.stringify({
          files: ['**/*.ts'],
          languageOptions: {
            parserOptions: {
              projectService: false,
              project: ['./tsconfig.json'],
            },
          },
          rules: { '@typescript-eslint/no-unsafe-member-access': 'error' },
          plugins: ['@typescript-eslint'],
        })},
      ];`,
    });
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      expect(result.stdout).not.toContain('dist/index.ts');
      expect(result.stdout).toContain('no-unsafe-member-access');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should preserve -- separator when passing args to Go binary', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'const a = 1;\n',
      'rslint.config.js': jsConfig({
        rules: { '@typescript-eslint/no-explicit-any': 'off' },
      }),
    });
    try {
      const result = await runRslint(['--', 'test.ts'], tempDir);
      expect(result.stderr).not.toContain('unknown flag');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should handle JS config with rule turned off', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'let a: any = 10;\na.b = 20;\n',
      'rslint.config.js': jsConfig({
        rules: {
          '@typescript-eslint/no-unsafe-member-access': 'off',
          '@typescript-eslint/no-explicit-any': 'off',
        },
      }),
    });
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      expect(result.stdout).not.toContain('no-unsafe-member-access');
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should auto-detect rslint.config.ts', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'let a: any = 10;\na.b = 20;\n',
      'rslint.config.ts': `export default [${JSON.stringify({
        files: ['**/*.ts'],
        languageOptions: {
          parserOptions: {
            projectService: false,
            project: ['./tsconfig.json'],
          },
        },
        rules: { '@typescript-eslint/no-unsafe-member-access': 'error' },
        plugins: ['@typescript-eslint'],
      })}];`,
    });
    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-unsafe-member-access');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should pass --fix flag through with JS config', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'const a = 1;\n',
      'rslint.config.js': jsConfig({
        rules: { '@typescript-eslint/no-explicit-any': 'off' },
      }),
    });
    try {
      const result = await runRslint(['--fix'], tempDir);
      expect(result.exitCode).toBe(0);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should pass --quiet flag through with JS config', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'let a: any = 10;\na.b = 20;\n',
      'rslint.config.js': jsConfig(),
    });
    try {
      const result = await runRslint(['--quiet'], tempDir);
      expect(result.exitCode).not.toBe(0);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('--init should generate rslint.config.ts when tsconfig.json exists', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
    });
    try {
      const result = await runRslint(['--init'], tempDir);
      expect(result.exitCode).toBe(0);
      const files = await fs.readdir(tempDir);
      expect(files).toContain('rslint.config.ts');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('--init should generate rslint.config.js for ESM project (type: module)', async () => {
    const tempDir = await createTempDir({
      'package.json': '{"type": "module"}',
    });
    try {
      const result = await runRslint(['--init'], tempDir);
      expect(result.exitCode).toBe(0);
      const files = await fs.readdir(tempDir);
      expect(files).toContain('rslint.config.js');
      expect(files).not.toContain('rslint.config.mjs');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('--init should generate rslint.config.mjs for non-ESM project', async () => {
    const tempDir = await createTempDir({
      'package.json': '{"name": "my-project"}',
    });
    try {
      const result = await runRslint(['--init'], tempDir);
      expect(result.exitCode).toBe(0);
      const files = await fs.readdir(tempDir);
      expect(files).toContain('rslint.config.mjs');
      expect(files).not.toContain('rslint.config.js');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('--init should generate rslint.config.mjs when no package.json', async () => {
    const tempDir = await createTempDir({});
    try {
      const result = await runRslint(['--init'], tempDir);
      expect(result.exitCode).toBe(0);
      const files = await fs.readdir(tempDir);
      expect(files).toContain('rslint.config.mjs');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

describe('CLI JS config error handling', () => {
  test('should show friendly error for syntactically invalid JS config', async () => {
    const tempDir = await createTempDir({
      'rslint.config.js': 'export default [INVALID SYNTAX',
    });
    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).not.toBe(0);
      expect(result.stderr).toContain('failed to load config');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should show friendly error for config that is not an array', async () => {
    const tempDir = await createTempDir({
      'rslint.config.js': 'export default { rules: {} };',
    });
    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).not.toBe(0);
      expect(result.stderr).toContain('invalid config');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

describe('JSON config regression', () => {
  test('JSON config should allow plugin rules without plugins declaration', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'const x: any = 1;\n',
      // JSON config: explicit rule WITHOUT plugins declaration
      // Should still work (JSON config has no plugin enforcement)
      'rslint.json': JSON.stringify([
        {
          files: ['**/*.ts'],
          languageOptions: {
            parserOptions: {
              projectService: false,
              project: ['./tsconfig.json'],
            },
          },
          rules: { '@typescript-eslint/no-explicit-any': 'error' },
        },
      ]),
    });
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      // JSON config has no plugin gate — rule should fire even without plugins
      expect(result.stdout).toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

describe('CLI config directory resolution', () => {
  test('src/**/*.ts pattern should match when cwd equals config directory', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': `export default [{
        files: ['src/**/*.ts'],
        languageOptions: {
          parserOptions: { projectService: false, project: ['./tsconfig.json'] },
        },
        rules: { '@typescript-eslint/no-explicit-any': 'error' },
        plugins: ['@typescript-eslint'],
      }];`,
      'src/index.ts': 'const x: any = 1;\n',
    });
    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

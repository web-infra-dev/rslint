import { describe, test, expect } from '@rstest/core';
import { spawn } from 'child_process';
import path from 'node:path';
import fs from 'node:fs/promises';
import { tmpdir } from 'node:os';
const RSLINT_BIN = require.resolve('@rslint/core/bin');

interface CliTestResult {
  exitCode: number;
  stdout: string;
  stderr: string;
}

async function runRslint(args: string[], cwd?: string): Promise<CliTestResult> {
  return new Promise((resolve) => {
    // Strip GITHUB_ACTIONS env to prevent setupColors() from force-enabling colors,
    // which would override --no-color and embed ANSI codes in the output.
    const { GITHUB_ACTIONS, FORCE_COLOR, ...cleanEnv } = process.env;
    const child = spawn(process.execPath, [RSLINT_BIN, '--no-color', ...args], {
      cwd: cwd || process.cwd(),
      stdio: ['pipe', 'pipe', 'pipe'],
      env: { ...cleanEnv, NO_COLOR: '1' },
    });

    let stdout = '';
    let stderr = '';

    child.stdout?.on('data', (data: Buffer) => {
      stdout += data.toString();
    });

    child.stderr?.on('data', (data: Buffer) => {
      stderr += data.toString();
    });

    child.on('close', (code) => {
      resolve({
        exitCode: code || 0,
        stdout,
        stderr,
      });
    });
  });
}

async function createTempDir(files: Record<string, string>): Promise<string> {
  const tempDir = await fs.mkdtemp(path.join(tmpdir(), 'rslint-test-'));

  for (const [filePath, content] of Object.entries(files)) {
    const fullPath = path.join(tempDir, filePath);
    await fs.mkdir(path.dirname(fullPath), { recursive: true });
    await fs.writeFile(fullPath, content, 'utf8');
  }

  return tempDir;
}

async function cleanupTempDir(tempDir: string): Promise<void> {
  await fs.rm(tempDir, { recursive: true, force: true });
}

const baseConfig = `export default [${JSON.stringify({
  files: ['**/*.ts'],
  languageOptions: {
    parserOptions: {
      projectService: false,
      project: ['./tsconfig.json'],
    },
  },
  rules: {
    '@typescript-eslint/no-unsafe-member-access': 'error',
  },
  plugins: ['@typescript-eslint'],
})}];`;

const baseJsonConfig = JSON.stringify([
  {
    language: 'javascript',
    files: ['**/*.ts'],
    languageOptions: {
      parserOptions: {
        projectService: false,
        project: ['./tsconfig.json'],
      },
    },
    rules: {
      '@typescript-eslint/no-unsafe-member-access': 'error',
    },
    plugins: ['@typescript-eslint'],
  },
]);

const baseTsConfig = JSON.stringify({
  compilerOptions: {
    target: 'ES2020',
    module: 'ESNext',
    strict: true,
  },
  include: ['**/*.ts'],
});

describe('CLI File Arguments', () => {
  test('should only lint specified file', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': baseConfig,
      'tsconfig.json': baseTsConfig,
      'error.ts': 'let a: any = 10;\na.b = 20;\n',
      'clean.ts': 'export const x = 1;\n',
    });

    try {
      const result = await runRslint(['error.ts'], tempDir);
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-unsafe-member-access');
      expect(result.stdout).toContain('linted 1 file');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should only lint the clean file and find no errors', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': baseConfig,
      'tsconfig.json': baseTsConfig,
      'error.ts': 'let a: any = 10;\na.b = 20;\n',
      'clean.ts': 'export const x = 1;\n',
    });

    try {
      // Lint only the clean file — should find no errors
      const result = await runRslint(['clean.ts'], tempDir);
      expect(result.exitCode).toBe(0);
      expect(result.stdout).not.toContain('no-unsafe-member-access');
      expect(result.stdout).toContain('linted 1 file');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should lint multiple specified files', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': baseConfig,
      'tsconfig.json': baseTsConfig,
      'a.ts': 'let a: any = 10;\na.b = 20;\n',
      'b.ts': 'let b: any = 20;\nb.c = 30;\n',
      'c.ts': 'export const c = 1;\n',
    });

    try {
      // Lint a.ts and b.ts, skip c.ts
      const result = await runRslint(['a.ts', 'b.ts'], tempDir);
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-unsafe-member-access');
      expect(result.stdout).toContain('linted 2 files');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should work with --fix and file arguments', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': `export default [${JSON.stringify({
        files: ['**/*.ts'],
        languageOptions: {
          parserOptions: {
            projectService: false,
            project: ['./tsconfig.json'],
          },
        },
        rules: {
          '@typescript-eslint/no-inferrable-types': 'error',
        },
        plugins: ['@typescript-eslint'],
      })}];`,
      'tsconfig.json': baseTsConfig,
      'fixable.ts': 'const x: number = 42;\n',
      'other.ts': 'const y: string = "hello";\n',
    });

    try {
      // Fix only fixable.ts
      await runRslint(['--fix', 'fixable.ts'], tempDir);

      // Check that fixable.ts was modified
      const fixedContent = await fs.readFile(
        path.join(tempDir, 'fixable.ts'),
        'utf8',
      );
      expect(fixedContent).not.toContain(': number');

      // Check that other.ts was NOT modified
      const otherContent = await fs.readFile(
        path.join(tempDir, 'other.ts'),
        'utf8',
      );
      expect(otherContent).toContain(': string');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should work with --format jsonline and file arguments', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': baseConfig,
      'tsconfig.json': baseTsConfig,
      'error.ts': 'let a: any = 10;\na.b = 20;\n',
    });

    try {
      const result = await runRslint(
        ['--format', 'jsonline', 'error.ts'],
        tempDir,
      );

      const lines = result.stdout
        .trim()
        .split('\n')
        .filter((line) => line.trim());
      for (const line of lines) {
        expect(() => JSON.parse(line)).not.toThrow();
      }
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should warn per-file when specified file is not in the project', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': baseConfig,
      'tsconfig.json': baseTsConfig,
      'existing.ts': 'export const x = 1;\n',
    });

    try {
      const result = await runRslint(['nonexistent.ts'], tempDir);
      // Should warn about the file not found and exit with 0
      expect(result.exitCode).toBe(0);
      expect(result.stderr).toContain('nonexistent.ts');
      expect(result.stderr).toContain('not found in the project');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should warn per-file when multiple specified files are all not in the project', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': baseConfig,
      'tsconfig.json': baseTsConfig,
      'existing.ts': 'export const x = 1;\n',
    });

    try {
      const result = await runRslint(
        ['nonexistent1.ts', 'nonexistent2.ts'],
        tempDir,
      );
      expect(result.exitCode).toBe(0);
      // Each file should have its own warning
      expect(result.stderr).toContain('nonexistent1.ts');
      expect(result.stderr).toContain('nonexistent2.ts');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should warn for non-project file while linting project files', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': baseConfig,
      'tsconfig.json': baseTsConfig,
      'clean.ts': 'export const x = 1;\n',
    });

    try {
      const result = await runRslint(['clean.ts', 'nonexistent.ts'], tempDir);
      expect(result.exitCode).toBe(0);
      expect(result.stdout).toContain('linted 1 file');
      // Should still warn about the non-project file
      expect(result.stderr).toContain('nonexistent.ts');
      expect(result.stderr).toContain('not found in the project');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should exit 0 with --max-warnings 0 when files are not in project', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': baseConfig,
      'tsconfig.json': baseTsConfig,
      'existing.ts': 'export const x = 1;\n',
    });

    try {
      const result = await runRslint(
        ['--max-warnings', '0', 'nonexistent.ts'],
        tempDir,
      );
      // Warning goes to stderr, not counted as lint warning, so exit 0
      expect(result.exitCode).toBe(0);
      expect(result.stderr).toContain('nonexistent.ts');
      expect(result.stderr).toContain('not found in the project');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should lint all files when no file arguments provided', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': baseConfig,
      'tsconfig.json': baseTsConfig,
      'a.ts': 'let a: any = 10;\na.x = 1;\n',
      'b.ts': 'let b: any = 20;\nb.y = 2;\n',
    });

    try {
      // No file args — should lint all files
      const result = await runRslint([], tempDir);
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('linted 2 files');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should work with file arguments and JS config', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': `
        export default [
          {
            language: 'javascript',
            files: ['**/*.ts'],
            languageOptions: {
              parserOptions: {
                projectService: false,
                project: ['./tsconfig.json'],
              },
            },
            rules: {
              '@typescript-eslint/no-unsafe-member-access': 'error',
            },
            plugins: ['@typescript-eslint'],
          },
        ];
      `,
      'tsconfig.json': baseTsConfig,
      'error.ts': 'let a: any = 10;\na.b = 20;\n',
      'clean.ts': 'export const x = 1;\n',
    });

    try {
      // Lint only the error file with JS config
      const result = await runRslint(['error.ts'], tempDir);
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-unsafe-member-access');
      expect(result.stdout).toContain('linted 1 file');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should accept absolute file paths', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': baseConfig,
      'tsconfig.json': baseTsConfig,
      'error.ts': 'let a: any = 10;\na.b = 20;\n',
    });

    try {
      // Use realpath to normalize symlinks (e.g. /var → /private/var on macOS)
      // so the file arg matches the config-discovered path for dedup.
      const realTempDir = await fs.realpath(tempDir);
      const absPath = path.join(realTempDir, 'error.ts');
      const result = await runRslint([absPath], tempDir);
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-unsafe-member-access');
      expect(result.stdout).toContain('linted 1 file');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should accept subdirectory file paths', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': baseConfig,
      'tsconfig.json': baseTsConfig,
      'src/error.ts': 'let a: any = 10;\na.b = 20;\n',
      'src/clean.ts': 'export const x = 1;\n',
    });

    try {
      const result = await runRslint(['src/error.ts'], tempDir);
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-unsafe-member-access');
      expect(result.stdout).toContain('linted 1 file');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should handle symlink-resolved absolute paths', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': baseConfig,
      'tsconfig.json': baseTsConfig,
      'error.ts': 'let a: any = 10;\na.b = 20;\n',
    });

    try {
      // Use realpath to get the symlink-resolved path (e.g. /private/var on macOS)
      // Both resolved and unresolved paths should work
      const realTempDir = await fs.realpath(tempDir);
      const absPath = path.join(realTempDir, 'error.ts');
      const result = await runRslint([absPath], tempDir);
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-unsafe-member-access');
      expect(result.stdout).toContain('linted 1 file');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should handle mix of existing and nonexistent file args', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': baseConfig,
      'tsconfig.json': baseTsConfig,
      'error.ts': 'let a: any = 10;\na.b = 20;\n',
    });

    try {
      // One file exists, one does not — should still lint the existing one
      const result = await runRslint(['error.ts', 'nonexistent.ts'], tempDir);
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-unsafe-member-access');
      expect(result.stdout).toContain('linted 1 file');
      // Should warn about the nonexistent file
      expect(result.stderr).toContain('nonexistent.ts');
      expect(result.stderr).toContain('not found in the project');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should deduplicate when same file is passed twice', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': baseConfig,
      'tsconfig.json': baseTsConfig,
      'error.ts': 'let a: any = 10;\na.b = 20;\n',
    });

    try {
      const result = await runRslint(['error.ts', 'error.ts'], tempDir);
      expect(result.exitCode).not.toBe(0);
      // Should only lint the file once even if specified twice
      expect(result.stdout).toContain('linted 1 file');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should resolve ../ in relative paths', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': baseConfig,
      'tsconfig.json': baseTsConfig,
      'src/error.ts': 'let a: any = 10;\na.b = 20;\n',
    });

    try {
      // src/../src/error.ts should resolve to src/error.ts
      const result = await runRslint(['src/../src/error.ts'], tempDir);
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-unsafe-member-access');
      expect(result.stdout).toContain('linted 1 file');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should work with --config and file arguments', async () => {
    const tempDir = await createTempDir({
      'custom.json': baseJsonConfig,
      'tsconfig.json': baseTsConfig,
      'error.ts': 'let a: any = 10;\na.b = 20;\n',
      'clean.ts': 'export const x = 1;\n',
    });

    try {
      const result = await runRslint(
        ['--config', 'custom.json', 'error.ts'],
        tempDir,
      );
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-unsafe-member-access');
      expect(result.stdout).toContain('linted 1 file');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should work with --quiet and file arguments', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': `export default [${JSON.stringify({
        files: ['**/*.ts'],
        languageOptions: {
          parserOptions: {
            projectService: false,
            project: ['./tsconfig.json'],
          },
        },
        rules: {
          '@typescript-eslint/no-unsafe-member-access': 'error',
          '@typescript-eslint/no-explicit-any': 'warn',
        },
        plugins: ['@typescript-eslint'],
      })}];`,
      'tsconfig.json': baseTsConfig,
      'error.ts': 'let a: any = 10;\na.b = 20;\n',
    });

    try {
      const result = await runRslint(['--quiet', 'error.ts'], tempDir);
      // --quiet should suppress warnings, only show errors
      expect(result.stdout).toContain('no-unsafe-member-access');
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should respect --max-warnings with file arguments', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': `export default [${JSON.stringify({
        files: ['**/*.ts'],
        languageOptions: {
          parserOptions: {
            projectService: false,
            project: ['./tsconfig.json'],
          },
        },
        rules: {
          '@typescript-eslint/no-explicit-any': 'warn',
        },
        plugins: ['@typescript-eslint'],
      })}];`,
      'tsconfig.json': baseTsConfig,
      'warn.ts': 'let a: any = 10;\nlet b: any = 20;\n',
    });

    try {
      // 2 warnings, threshold is 0 -> should fail
      const result = await runRslint(
        ['--max-warnings', '0', 'warn.ts'],
        tempDir,
      );
      expect(result.exitCode).not.toBe(0);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should work without tsconfig (pure JS project)', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': `export default [${JSON.stringify({
        files: ['**/*.js'],
        rules: {
          'no-empty': 'error',
        },
      })}];`,
      'error.js': 'if (true) {}\n',
      'clean.js': 'const x = 1;\n',
    });

    try {
      const result = await runRslint(['error.js'], tempDir);
      expect(result.stdout).toContain('no-empty');
      expect(result.stdout).toContain('linted 1 file');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

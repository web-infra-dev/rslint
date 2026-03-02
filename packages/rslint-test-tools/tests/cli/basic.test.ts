import { describe, test, expect } from '@rstest/core';
import { spawn } from 'child_process';
import path from 'node:path';
import fs from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import { tmpdir } from 'node:os';
const RSLINT_BIN = require.resolve('@rslint/core/bin');

interface CliTestResult {
  exitCode: number;
  stdout: string;
  stderr: string;
}

/**
 * Helper function to run rslint CLI command
 */
async function runRslint(args: string[], cwd?: string): Promise<CliTestResult> {
  return new Promise(resolve => {
    const child = spawn(RSLINT_BIN, args, {
      cwd: cwd || process.cwd(),
      stdio: ['pipe', 'pipe', 'pipe'],
    });

    let stdout = '';
    let stderr = '';

    child.stdout?.on('data', (data: Buffer) => {
      stdout += data.toString();
    });

    child.stderr?.on('data', (data: Buffer) => {
      stderr += data.toString();
    });

    child.on('close', code => {
      resolve({
        exitCode: code || 0,
        stdout,
        stderr,
      });
    });
  });
}

/**
 * Create a temporary directory with test files
 */
async function createTempDir(files: Record<string, string>): Promise<string> {
  const tempDir = await fs.mkdtemp(path.join(tmpdir(), 'rslint-test-'));

  for (const [filePath, content] of Object.entries(files)) {
    const fullPath = path.join(tempDir, filePath);
    await fs.mkdir(path.dirname(fullPath), { recursive: true });
    await fs.writeFile(fullPath, content, 'utf8');
  }

  return tempDir;
}

/**
 * Cleanup temporary directory
 */
async function cleanupTempDir(tempDir: string): Promise<void> {
  await fs.rm(tempDir, { recursive: true, force: true });
}

describe('CLI Configuration Tests', () => {
  test('should show help when --help flag is used', async () => {
    const result = await runRslint(['--help']);
    expect(result.exitCode).toBe(0);
  });

  test('should show help when -h flag is used', async () => {
    const result = await runRslint(['-h']);
    expect(result.exitCode).toBe(0);
  });

  test('should use default config when no config specified', async () => {
    const tempDir = await createTempDir({
      'rslint.json': JSON.stringify([
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
      ]),
      'tsconfig.json': JSON.stringify({
        compilerOptions: {
          target: 'ES2020',
          module: 'ESNext',
          strict: true,
        },
        include: ['**/*.ts'],
      }),
      'test.ts': `
        let a: any = 10;
        a.b = 20; // This should trigger no-unsafe-member-access
      `,
    });

    try {
      const result = await runRslint([], tempDir);

      // Should find and use the default rslint.json config
      expect(result.stdout).toContain('no-unsafe-member-access');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should use custom config when --config flag is specified', async () => {
    const tempDir = await createTempDir({
      'custom-config.json': JSON.stringify([
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
            '@typescript-eslint/no-unsafe-assignment': 'error',
          },
          plugins: ['@typescript-eslint'],
        },
      ]),
      'tsconfig.json': JSON.stringify({
        compilerOptions: {
          target: 'ES2020',
          module: 'ESNext',
          strict: true,
        },
        include: ['**/*.ts'],
      }),
      'test.ts': `
        let a: any = getValue();
        let b: string = a; // This should trigger no-unsafe-assignment
      `,
    });

    try {
      const result = await runRslint(
        ['--config', 'custom-config.json'],
        tempDir,
      );

      expect(result.stdout).toContain('no-unsafe-assignment');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should output in jsonline format when --format jsonline is used', async () => {
    const tempDir = await createTempDir({
      'rslint.json': JSON.stringify([
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
      ]),
      'tsconfig.json': JSON.stringify({
        compilerOptions: {
          target: 'ES2020',
          module: 'ESNext',
          strict: true,
        },
        include: ['**/*.ts'],
      }),
      'test.ts': `
        let a: any = 10;
        a.b = 20;
      `,
    });

    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);

      // Should output valid JSON lines
      const lines = result.stdout
        .trim()
        .split('\n')
        .filter(line => line.trim());
      for (const line of lines) {
        // eslint-disable-next-line
        expect(() => JSON.parse(line)).not.toThrow();
      }
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should output in github workflow format when --format github is used', async () => {
    const tempDir = await createTempDir({
      'rslint.json': JSON.stringify([
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
            '@typescript-eslint/no-explicit-any': 'warn',
            '@typescript-eslint/no-unsafe-member-access': 'error',
          },
          plugins: ['@typescript-eslint'],
        },
      ]),
      'tsconfig.json': JSON.stringify({
        compilerOptions: {
          target: 'ES2020',
          module: 'ESNext',
          strict: true,
        },
        include: ['**/*.ts'],
      }),
      'test:%,\r\nfile.ts': `
        let a: any = 10;
        a.b = 20;
      `,
    });

    try {
      const result = await runRslint(['--format', 'github'], tempDir);

      const lines = result.stdout
        .trim()
        .split('\n')
        .filter(line => line.trim());

      expect(lines.length).toBe(2);
      expect(lines[0]).toBe(
        '::warning file=test%3A%25%2C%0D%0Afile.ts,line=2,endLine=2,col=16,endColumn=19,title=@typescript-eslint/no-explicit-any::Unexpected any. Specify a different type.',
      );
      expect(lines[1]).toBe(
        '::error file=test%3A%25%2C%0D%0Afile.ts,line=3,endLine=3,col=11,endColumn=12,title=@typescript-eslint/no-unsafe-member-access::Unsafe member access .b on an `any` value.',
      );
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should only report errors when --quiet flag is used', async () => {
    const tempDir = await createTempDir({
      'rslint.json': JSON.stringify([
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
            '@typescript-eslint/return-await': 'warn',
          },
          plugins: ['@typescript-eslint'],
        },
      ]),
      'tsconfig.json': JSON.stringify({
        compilerOptions: {
          target: 'ES2020',
          module: 'ESNext',
          strict: true,
        },
        include: ['**/*.ts'],
      }),
      'test.ts': `
        let a: any = 10;
        a.b = 20;
        async function invalidInTryCatch1() {
        try {
          return Promise.reject('try');
        } catch (e) {
          // Doesn't execute due to missing await.
          }
        }
      `,
    });

    try {
      const result = await runRslint(['--quiet'], tempDir);
      // Should contain error reports but not verbose output
      expect(result.stdout).toContain('no-unsafe-member-access');
      // Should not contain summary information typically shown in verbose mode
      expect(result.stdout).not.toContain('return-await');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should handle invalid config file gracefully', async () => {
    const tempDir = await createTempDir({
      'invalid-config.json': 'invalid json content',
      'test.ts': 'export const a = 1;',
    });

    try {
      const result = await runRslint(
        ['--config', 'invalid-config.json'],
        tempDir,
      );

      expect(result.exitCode).not.toBe(0);
      expect(result.stderr || result.stdout).toContain('invalid');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should handle non-existent config file gracefully', async () => {
    const tempDir = await createTempDir({
      'test.ts': 'export const a = 1;',
    });

    try {
      const result = await runRslint(
        ['--config', 'non-existent.json'],
        tempDir,
      );

      expect(result.exitCode).not.toBe(0);
      expect(result.stderr || result.stdout).toMatch(/config|file|not found/i);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should handle directory with no matching files', async () => {
    const tempDir = await createTempDir({
      'rslint.json': JSON.stringify([
        {
          language: 'javascript',
          files: ['**/*.ts'],
          languageOptions: {
            parserOptions: {
              projectService: false,
              project: ['./tsconfig.json'],
            },
          },
          rules: {},
          plugins: ['@typescript-eslint'],
        },
      ]),
      'tsconfig.json': JSON.stringify({
        compilerOptions: {
          target: 'ES2020',
          module: 'ESNext',
        },
        include: ['**/*.ts'],
      }),
      'readme.md': '# Test project',
      'package.json': '{"name": "test"}',
    });

    try {
      const result = await runRslint([], tempDir);

      expect(result.exitCode).toBe(0);
      // Should not report any errors since no files match the pattern
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should exit with non-zero code when linting errors are found', async () => {
    const tempDir = await createTempDir({
      'rslint.json': JSON.stringify([
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
            'no-unsafe-member-access': 'error',
          },
          plugins: ['@typescript-eslint'],
        },
      ]),
      'tsconfig.json': JSON.stringify({
        compilerOptions: {
          target: 'ES2020',
          module: 'ESNext',
          strict: true,
        },
        include: ['**/*.ts'],
      }),
      'test.ts': `
        let a: any = 10;
        a.b = 20; // This should trigger an error
      `,
    });

    try {
      const result = await runRslint([], tempDir);

      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-unsafe-member-access');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should exit with zero code when no linting errors are found', async () => {
    const tempDir = await createTempDir({
      'rslint.json': JSON.stringify([
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
            'no-unsafe-member-access': 'error',
            'no-console': 'off',
          },
          plugins: ['@typescript-eslint'],
        },
      ]),
      'tsconfig.json': JSON.stringify({
        compilerOptions: {
          target: 'ES2020',
          module: 'ESNext',
          strict: true,
        },
        include: ['**/*.ts'],
      }),
      'test.ts': `
        const a = "hello";
        console.log(a.length); // This is safe
      `,
    });

    try {
      const result = await runRslint([], tempDir);

      expect(result.exitCode).toBe(0);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

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

/**
 * Run the rslint entry point explicitly via `node`, matching how pnpm/.cmd
 * wrappers invoke it on Windows. This ensures the dynamic `import()` inside
 * bin/rslint.cjs is exercised — on Windows, bare `spawn('file.cjs')` may
 * bypass Node.js entirely because shebangs are not supported.
 */
function runRslintViaNode(
  args: string[],
  cwd?: string,
): Promise<CliTestResult> {
  return new Promise((resolve) => {
    const { GITHUB_ACTIONS, FORCE_COLOR, ...cleanEnv } = process.env;
    const child = spawn(process.execPath, [RSLINT_BIN, ...args], {
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
      resolve({ exitCode: code || 0, stdout, stderr });
    });
  });
}

async function createTempDir(files: Record<string, string>): Promise<string> {
  const tempDir = await fs.mkdtemp(path.join(tmpdir(), 'rslint-entry-'));
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

describe('Entry point (bin/rslint.cjs)', () => {
  test('should load ESM cli module without ERR_UNSUPPORTED_ESM_URL_SCHEME', async () => {
    const result = await runRslintViaNode(['--help']);
    expect(result.stderr).not.toContain('ERR_UNSUPPORTED_ESM_URL_SCHEME');
    expect(result.exitCode).toBe(0);
  });

  test('should run lint with JS config via node entry point', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': `export default [${JSON.stringify({
        files: ['**/*.ts'],
        languageOptions: {
          parserOptions: {
            projectService: false,
            project: ['./tsconfig.json'],
          },
        },
        rules: { 'no-debugger': 'error' },
        plugins: ['@typescript-eslint'],
      })}];`,
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['**/*.ts'],
      }),
      'test.ts': 'debugger;\n',
    });

    try {
      const result = await runRslintViaNode([], tempDir);
      expect(result.stderr).not.toContain('ERR_UNSUPPORTED_ESM_URL_SCHEME');
      expect(result.stdout).toContain('no-debugger');
      expect(result.exitCode).not.toBe(0);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

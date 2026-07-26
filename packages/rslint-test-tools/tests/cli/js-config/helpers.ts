import path from 'node:path';
import fs from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { runCliProcess, type CliProcessResult } from '../spawn-cli.js';

export const RSLINT_BIN = require.resolve('@rslint/core/bin');

export async function runRslint(
  args: string[],
  cwd?: string,
): Promise<CliProcessResult> {
  // Strip GITHUB_ACTIONS/FORCE_COLOR to prevent Go binary from force-enabling
  // ANSI colors, which would embed escape codes in stdout and break assertions.
  const { GITHUB_ACTIONS, FORCE_COLOR, ...cleanEnv } = process.env;
  return runCliProcess(process.execPath, [RSLINT_BIN, ...args], {
    cwd: cwd || process.cwd(),
    env: { ...cleanEnv, NO_COLOR: '1' },
  });
}

export async function createTempDir(
  files: Record<string, string>,
): Promise<string> {
  const tempDir = await fs.mkdtemp(path.join(tmpdir(), 'rslint-jsconfig-'));
  for (const [filePath, content] of Object.entries(files)) {
    const fullPath = path.join(tempDir, filePath);
    await fs.mkdir(path.dirname(fullPath), { recursive: true });
    await fs.writeFile(fullPath, content, 'utf8');
  }
  return tempDir;
}

export async function cleanupTempDir(tempDir: string): Promise<void> {
  await fs.rm(tempDir, { recursive: true, force: true });
}

export const TS_CONFIG = JSON.stringify({
  compilerOptions: {
    target: 'ES2020',
    module: 'ESNext',
    strict: true,
  },
  include: ['**/*.ts'],
});

export function jsConfig(overrides: Record<string, unknown> = {}): string {
  const entry = {
    files: ['**/*.ts'],
    languageOptions: {
      parserOptions: {
        projectService: false,
        project: ['./tsconfig.json'],
      },
    },
    rules: { '@typescript-eslint/no-unsafe-member-access': 'error' },
    plugins: ['@typescript-eslint'],
    ...overrides,
  };
  return `export default [${JSON.stringify(entry)}];`;
}

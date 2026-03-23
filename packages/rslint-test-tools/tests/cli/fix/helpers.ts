import { spawn } from 'child_process';
import path from 'node:path';
import fs from 'node:fs/promises';
import { tmpdir } from 'node:os';

const RSLINT_BIN = require.resolve('@rslint/core/bin');

export interface CliTestResult {
  exitCode: number;
  stdout: string;
  stderr: string;
}

export async function runRslint(
  args: string[],
  cwd?: string,
): Promise<CliTestResult> {
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
      resolve({ exitCode: code || 0, stdout, stderr });
    });
  });
}

export async function createTempDir(
  files: Record<string, string>,
): Promise<string> {
  const tempDir = await fs.mkdtemp(path.join(tmpdir(), 'rslint-fix-test-'));
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

export const TSCONFIG = JSON.stringify({
  compilerOptions: { strict: true, target: 'es2020', module: 'es2020' },
  include: ['*.ts'],
});

/**
 * Create a rslint.config.mjs content string with the given rules.
 * Uses JS config format (not JSON) so only explicitly listed rules are active,
 * without plugin defaults leaking in.
 */
export function makeConfig(rules: Record<string, string>): string {
  const rulesStr = Object.entries(rules)
    .map(([k, v]) => `    '${k}': '${v}'`)
    .join(',\n');

  return `export default [{
  files: ['**/*.ts'],
  languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
  rules: {
${rulesStr}
  }
}];
`;
}

/**
 * Create standard temp dir files with rslint.config.mjs and tsconfig.json.
 */
export async function createFixTestDir(
  rules: Record<string, string>,
  sourceFiles: Record<string, string>,
): Promise<string> {
  return createTempDir({
    'tsconfig.json': TSCONFIG,
    'rslint.config.mjs': makeConfig(rules),
    ...sourceFiles,
  });
}

export { path, fs };

import { spawn } from 'child_process';
import path from 'node:path';
import fs from 'node:fs/promises';
import { tmpdir } from 'node:os';

export const RSLINT_BIN = require.resolve('@rslint/core/bin');

export interface CliTestResult {
  exitCode: number;
  stdout: string;
  stderr: string;
}

export async function runRslint(
  args: string[],
  cwd: string,
): Promise<CliTestResult> {
  return new Promise((resolve) => {
    const { GITHUB_ACTIONS, FORCE_COLOR, ...cleanEnv } = process.env;
    const child = spawn(process.execPath, [RSLINT_BIN, ...args], {
      cwd,
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

export async function createTempDir(
  files: Record<string, string>,
): Promise<string> {
  const tempDir = await fs.mkdtemp(path.join(tmpdir(), 'rslint-typecheck-'));
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

export async function addUntypedPackage(
  tempDir: string,
  pkgName: string,
): Promise<void> {
  const pkgDir = path.join(tempDir, 'node_modules', pkgName);
  await fs.mkdir(pkgDir, { recursive: true });
  await fs.writeFile(
    path.join(pkgDir, 'package.json'),
    JSON.stringify({ name: pkgName, version: '1.0.0', main: 'index.js' }),
  );
  await fs.writeFile(path.join(pkgDir, 'index.js'), 'module.exports = {};');
}

/**
 * Normalize CLI output for snapshot stability:
 * - Replace temp dir paths with <TEMPDIR>
 * - Replace timing/thread info in summary line
 */
export function normalizeOutput(output: string, tempDir: string): string {
  // Diagnostic output always uses forward slashes, even on Windows where
  // tempDir (from fs.mkdtemp) uses backslashes, so compare using a
  // posix-style form of tempDir rather than the raw path.
  const posixTempDir = tempDir.replace(/\\/g, '/');
  const privateTempDir = `/private${posixTempDir}`;
  let result = output;
  if (privateTempDir.length > posixTempDir.length) {
    result = result.replaceAll(privateTempDir, '<TEMPDIR>');
  }
  result = result.replaceAll(posixTempDir, '<TEMPDIR>');
  // macOS: relative paths through /private/tmp symlink
  result = result.replace(
    /(?:\.\.\/)+private\/tmp\/rslint-typecheck-[^\s/)]+/g,
    '<TEMPDIR>',
  );
  result = result.replace(
    /(?:\.\.\/)+tmp\/rslint-typecheck-[^\s/)]+/g,
    '<TEMPDIR>',
  );
  // Windows: absolute paths with possible 8.3 short names (e.g. RUNNER~1),
  // regardless of which directory %TEMP%/%TMP% points to on the runner
  result = result.replace(
    /[A-Z]:\/(?:[^\s/)]+\/)*rslint-typecheck-[^\s/)]+/g,
    '<TEMPDIR>',
  );
  // Windows: relative paths through 8.3 short name directories
  result = result.replace(
    /(?:\.\.\/)+(?:[^\s/)]+\/)*rslint-typecheck-[^\s/)]+/g,
    '<TEMPDIR>',
  );
  result = result.replace(
    /in [\d.]+m?s using \d+ threads?/g,
    'in <TIME> using <N> thread(s)',
  );
  return result;
}

export const TS_CONFIG = JSON.stringify({
  compilerOptions: {
    target: 'ES2020',
    module: 'ESNext',
    strict: true,
    moduleResolution: 'bundler',
  },
  include: ['**/*.ts'],
});

export function makeConfig(rules: Record<string, string> = {}): string {
  const rulesStr = Object.entries(rules)
    .map(([k, v]) => `    '${k}': '${v}'`)
    .join(',\n');

  return `export default [{
  files: ['**/*.ts'],
  languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
  plugins: ['@typescript-eslint'],
  rules: {
${rulesStr}
  }
}];
`;
}

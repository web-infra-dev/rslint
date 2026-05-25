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
    const child = spawn(process.execPath, [RSLINT_BIN, ...args], {
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

    child.on('close', (code) => {
      resolve({ exitCode: code || 0, stdout, stderr });
    });
  });
}

async function createTempDir(files: Record<string, string>): Promise<string> {
  const tempDir = await fs.mkdtemp(path.join(tmpdir(), 'rslint-disable-test-'));

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

function makeConfig(rules: Record<string, unknown>) {
  return `export default [${JSON.stringify({
    files: ['**/*.ts'],
    languageOptions: {
      parserOptions: {
        projectService: false,
        project: ['./tsconfig.json'],
      },
    },
    rules,
    plugins: ['@typescript-eslint'],
  })}];`;
}

const TSCONFIG = JSON.stringify({
  compilerOptions: {
    target: 'ES2020',
    module: 'ESNext',
    strict: true,
  },
  include: ['**/*.ts'],
});

const RULE = '@typescript-eslint/no-explicit-any';

describe('rslint-disable comment directives', () => {
  // -------------------------------------------------------------------
  // rslint-disable-next-line
  // -------------------------------------------------------------------

  test('rslint-disable-next-line suppresses diagnostic on next line', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        '// rslint-disable-next-line @typescript-eslint/no-explicit-any',
        'const x: any = 1;',
        'export { x };',
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).toBe(0);
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -------------------------------------------------------------------
  // rslint-disable-line
  // -------------------------------------------------------------------

  test('rslint-disable-line suppresses diagnostic on same line', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        'const x: any = 1; // rslint-disable-line @typescript-eslint/no-explicit-any',
        'export { x };',
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).toBe(0);
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -------------------------------------------------------------------
  // rslint-disable + rslint-enable (block range)
  // -------------------------------------------------------------------

  test('rslint-disable + rslint-enable suppresses diagnostics inside range', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        '/* rslint-disable @typescript-eslint/no-explicit-any */',
        'const a: any = 1;',
        'const b: any = 2;',
        '/* rslint-enable @typescript-eslint/no-explicit-any */',
        'export { a, b };',
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).toBe(0);
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -------------------------------------------------------------------
  // rslint-disable without enable (rest of file)
  // -------------------------------------------------------------------

  test('rslint-disable without enable suppresses rest of file', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        '/* rslint-disable @typescript-eslint/no-explicit-any */',
        'const a: any = 1;',
        'const b: any = 2;',
        'export { a, b };',
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).toBe(0);
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -------------------------------------------------------------------
  // wildcard (no rule name)
  // -------------------------------------------------------------------

  test('rslint-disable without rule name suppresses all rules', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        '/* rslint-disable */',
        'const a: any = 1;',
        '/* rslint-enable */',
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).toBe(0);
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -------------------------------------------------------------------
  // mixed prefixes: rslint-disable + eslint-enable
  // -------------------------------------------------------------------

  test('rslint-disable + eslint-enable mixed prefixes work', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        '/* rslint-disable @typescript-eslint/no-explicit-any */',
        'const a: any = 1;',
        '/* eslint-enable @typescript-eslint/no-explicit-any */',
        'export { a };',
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).toBe(0);
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('eslint-disable + rslint-enable mixed prefixes work', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        '/* eslint-disable @typescript-eslint/no-explicit-any */',
        'const a: any = 1;',
        '/* rslint-enable @typescript-eslint/no-explicit-any */',
        'export { a };',
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).toBe(0);
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -------------------------------------------------------------------
  // negative tests
  // -------------------------------------------------------------------

  test('rslint-disable-next-line does not suppress other lines', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        '// rslint-disable-next-line @typescript-eslint/no-explicit-any',
        'const x: any = 1;',
        'const y: any = 2;', // not suppressed
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('rslint-disable + rslint-enable does NOT suppress diagnostics outside range', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        '/* rslint-disable @typescript-eslint/no-explicit-any */',
        'const a: any = 1;',
        '/* rslint-enable @typescript-eslint/no-explicit-any */',
        'const b: any = 2;', // outside range — should trigger
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -------------------------------------------------------------------
  // block comment style for line directives
  // -------------------------------------------------------------------

  test('block comment style works for rslint-disable-next-line', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        '/* rslint-disable-next-line @typescript-eslint/no-explicit-any */',
        'const x: any = 1;',
        'export { x };',
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).toBe(0);
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -------------------------------------------------------------------
  // rslint-disable-next-line with -- description
  // -------------------------------------------------------------------

  test('rslint-disable-next-line with -- description suppresses diagnostic', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        '// rslint-disable-next-line @typescript-eslint/no-explicit-any -- needed for legacy API',
        'const x: any = 1;',
        'export { x };',
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).toBe(0);
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

describe('eslint-disable comment directives', () => {
  // -------------------------------------------------------------------
  // eslint-disable-next-line
  // -------------------------------------------------------------------

  test('eslint-disable-next-line suppresses diagnostic on next line', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        '// eslint-disable-next-line @typescript-eslint/no-explicit-any',
        'const x: any = 1;',
        'export { x };',
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).toBe(0);
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('eslint-disable-next-line does not suppress other lines', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        '// eslint-disable-next-line @typescript-eslint/no-explicit-any',
        'const x: any = 1;',
        'const y: any = 2;', // not suppressed
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -------------------------------------------------------------------
  // eslint-disable-line
  // -------------------------------------------------------------------

  test('eslint-disable-line suppresses diagnostic on same line', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        'const x: any = 1; // eslint-disable-line @typescript-eslint/no-explicit-any',
        'export { x };',
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).toBe(0);
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -------------------------------------------------------------------
  // eslint-disable + eslint-enable (block range) — core bug fix
  // -------------------------------------------------------------------

  test('eslint-disable + eslint-enable suppresses diagnostics inside range', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        '/* eslint-disable @typescript-eslint/no-explicit-any */',
        'const a: any = 1;',
        'const b: any = 2;',
        '/* eslint-enable @typescript-eslint/no-explicit-any */',
        'export { a, b };',
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).toBe(0);
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('eslint-disable + eslint-enable does NOT suppress diagnostics outside range', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        '/* eslint-disable @typescript-eslint/no-explicit-any */',
        'const a: any = 1;',
        '/* eslint-enable @typescript-eslint/no-explicit-any */',
        'const b: any = 2;', // outside range — should trigger
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -------------------------------------------------------------------
  // eslint-disable without enable (rest of file)
  // -------------------------------------------------------------------

  test('eslint-disable without enable suppresses rest of file', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        '/* eslint-disable @typescript-eslint/no-explicit-any */',
        'const a: any = 1;',
        'const b: any = 2;',
        'const c: any = 3;',
        'export { a, b, c };',
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).toBe(0);
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -------------------------------------------------------------------
  // wildcard (no rule name)
  // -------------------------------------------------------------------

  test('eslint-disable without rule name suppresses all rules', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        '/* eslint-disable */',
        'const a: any = 1;',
        '/* eslint-enable */',
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).toBe(0);
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('wildcard eslint-disable-next-line suppresses all rules on next line', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': ['// eslint-disable-next-line', 'const x: any = 1;'].join(
        '\n',
      ),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).toBe(0);
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -------------------------------------------------------------------
  // block comment style for line directives
  // -------------------------------------------------------------------

  test('block comment style works for disable-next-line', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        '/* eslint-disable-next-line @typescript-eslint/no-explicit-any */',
        'const x: any = 1;',
        'export { x };',
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).toBe(0);
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -------------------------------------------------------------------
  // multiple disable/enable ranges
  // -------------------------------------------------------------------

  // -------------------------------------------------------------------
  // disable-next-line with -- description
  // -------------------------------------------------------------------

  test('eslint-disable-next-line with -- description suppresses diagnostic', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        '// eslint-disable-next-line @typescript-eslint/no-explicit-any -- needed for legacy API',
        'const x: any = 1;',
        'export { x };',
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).toBe(0);
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -------------------------------------------------------------------
  // disable-next-line inside function arguments (leading trivia fix)
  // -------------------------------------------------------------------

  test('eslint-disable-next-line works inside function call arguments', async () => {
    const ASSERTION_RULE = '@typescript-eslint/consistent-type-assertions';
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({
        [ASSERTION_RULE]: ['error', { assertionStyle: 'never' }],
      }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        'type Foo = Record<string, unknown>;',
        'const result = [1, 2, 3].reduce(',
        '  (acc, _item) => acc,',
        '  // eslint-disable-next-line @typescript-eslint/consistent-type-assertions',
        '  {} as Foo,',
        ');',
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      // Check the specific error format, not just the rule name string,
      // since the rule name also appears in the disable comment context lines
      expect(result.stdout).not.toContain(
        'consistent-type-assertions  \u2014 [error]',
      );
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('eslint-disable-next-line works inside array literals', async () => {
    const ASSERTION_RULE = '@typescript-eslint/consistent-type-assertions';
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({
        [ASSERTION_RULE]: ['error', { assertionStyle: 'never' }],
      }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        'type Foo = Record<string, unknown>;',
        'const arr = [',
        '  // eslint-disable-next-line @typescript-eslint/consistent-type-assertions',
        '  {} as Foo,',
        '];',
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.stdout).not.toContain(
        'consistent-type-assertions  \u2014 [error]',
      );
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // -------------------------------------------------------------------
  // multiple disable/enable ranges
  // -------------------------------------------------------------------

  test('multiple disable/enable ranges work independently', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': makeConfig({ [RULE]: 'error' }),
      'tsconfig.json': TSCONFIG,
      'test.ts': [
        '/* eslint-disable @typescript-eslint/no-explicit-any */',
        'export const a: any = 1;', // suppressed
        '/* eslint-enable @typescript-eslint/no-explicit-any */',
        'export const b = 2;', // no violation
        '/* eslint-disable @typescript-eslint/no-explicit-any */',
        'export const c: any = 3;', // suppressed
        '/* eslint-enable @typescript-eslint/no-explicit-any */',
      ].join('\n'),
    });

    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).toBe(0);
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

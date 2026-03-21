import { describe, test, expect } from '@rstest/core';
import {
  runRslint,
  createFixTestDir,
  cleanupTempDir,
  path,
  fs,
} from './helpers';

describe('CLI --fix exit code', () => {
  test('all fixable errors → exit 0', async () => {
    const tempDir = await createFixTestDir(
      { '@typescript-eslint/no-unnecessary-type-assertion': 'error' },
      {
        'index.ts':
          "const x: string = 'hello';\nconst y = (x as string).toUpperCase();\nexport { y };\n",
      },
    );

    try {
      const result = await runRslint(['--fix', 'index.ts'], tempDir);
      expect(result.exitCode).toBe(0);
      expect(result.stdout).toContain('Found 0 errors');
      expect(result.stdout).toContain('fixed 1 issue');

      const content = await fs.readFile(path.join(tempDir, 'index.ts'), 'utf8');
      expect(content).not.toContain('as string');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('non-fixable errors remain → exit 1', async () => {
    const tempDir = await createFixTestDir(
      { '@typescript-eslint/no-unsafe-member-access': 'error' },
      { 'index.ts': 'const z: any = {};\nz.foo;\nexport { z };\n' },
    );

    try {
      const result = await runRslint(['--fix', 'index.ts'], tempDir);
      expect(result.exitCode).toBe(1);
      expect(result.stdout).toMatch(/Found \d+ error/);
      expect(result.stdout).not.toContain('Found 0 errors');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('mixed fixable + non-fixable → exit 1, fixable applied', async () => {
    const tempDir = await createFixTestDir(
      {
        '@typescript-eslint/no-unnecessary-type-assertion': 'error',
        '@typescript-eslint/no-unsafe-member-access': 'error',
      },
      {
        'index.ts':
          "const x: string = 'hello';\nconst y = (x as string).toUpperCase();\nconst z: any = {};\nz.foo;\nexport { y, z };\n",
      },
    );

    try {
      const result = await runRslint(['--fix', 'index.ts'], tempDir);
      expect(result.exitCode).toBe(1);
      expect(result.stdout).toContain('fixed 1 issue');
      expect(result.stdout).not.toContain('Found 0 errors');

      const content = await fs.readFile(path.join(tempDir, 'index.ts'), 'utf8');
      expect(content).not.toContain('as string');
      expect(content).toContain('z.foo');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('no errors → exit 0', async () => {
    const tempDir = await createFixTestDir(
      { '@typescript-eslint/no-unnecessary-type-assertion': 'error' },
      { 'index.ts': "const x = 'hello';\nexport { x };\n" },
    );

    try {
      const result = await runRslint(['--fix', 'index.ts'], tempDir);
      expect(result.exitCode).toBe(0);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('nothing fixable → exit 1, skips multi-pass loop', async () => {
    const tempDir = await createFixTestDir(
      { '@typescript-eslint/no-unsafe-member-access': 'error' },
      { 'index.ts': 'const z: any = {};\nz.foo;\nexport { z };\n' },
    );

    try {
      const result = await runRslint(['--fix', 'index.ts'], tempDir);
      expect(result.exitCode).toBe(1);
      expect(result.stdout).not.toContain('fixed');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('warn-only fixable → exit 0 after fix', async () => {
    const tempDir = await createFixTestDir(
      { '@typescript-eslint/no-unnecessary-type-assertion': 'warn' },
      {
        'index.ts':
          "const x: string = 'hello';\nconst y = (x as string).toUpperCase();\nexport { y };\n",
      },
    );

    try {
      const result = await runRslint(['--fix', 'index.ts'], tempDir);
      expect(result.exitCode).toBe(0);

      const content = await fs.readFile(path.join(tempDir, 'index.ts'), 'utf8');
      expect(content).not.toContain('as string');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('warn fixable + error non-fixable → exit 1', async () => {
    const tempDir = await createFixTestDir(
      {
        '@typescript-eslint/no-unnecessary-type-assertion': 'warn',
        '@typescript-eslint/no-unsafe-member-access': 'error',
      },
      {
        'index.ts':
          "const x: string = 'hello';\nconst y = (x as string).toUpperCase();\nconst z: any = {};\nz.foo;\nexport { y, z };\n",
      },
    );

    try {
      const result = await runRslint(['--fix', 'index.ts'], tempDir);
      expect(result.exitCode).toBe(1);

      const content = await fs.readFile(path.join(tempDir, 'index.ts'), 'utf8');
      expect(content).not.toContain('as string');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

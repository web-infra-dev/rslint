import { describe, test, expect } from '@rstest/core';
import { runRslint, createFixTestDir, cleanupTempDir } from './helpers';

describe('CLI --fix output', () => {
  test('all fixable → fixed diagnostics NOT shown in output', async () => {
    const tempDir = await createFixTestDir(
      { '@typescript-eslint/no-unnecessary-type-assertion': 'error' },
      {
        'index.ts':
          "const x: string = 'hello';\nconst y = (x as string).trim();\nexport { y };\n",
      },
    );

    try {
      const result = await runRslint(['--fix', 'index.ts'], tempDir);
      // Fixed diagnostic should NOT appear in output
      expect(result.stdout).not.toContain('no-unnecessary-type-assertion');
      // Summary should show fix count
      expect(result.stdout).toContain('fixed 1 issue');
      expect(result.stdout).toContain('Found 0 errors');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('mixed → only unfixed diagnostics shown', async () => {
    const tempDir = await createFixTestDir(
      {
        '@typescript-eslint/no-unnecessary-type-assertion': 'error',
        '@typescript-eslint/no-unsafe-member-access': 'error',
      },
      {
        'index.ts':
          "const x: string = 'hello';\nconst y = (x as string).trim();\nconst z: any = {};\nz.foo;\nexport { y, z };\n",
      },
    );

    try {
      const result = await runRslint(['--fix', 'index.ts'], tempDir);
      // Fixed rule should NOT appear
      expect(result.stdout).not.toContain('no-unnecessary-type-assertion');
      // Unfixed rule SHOULD appear
      expect(result.stdout).toContain('no-unsafe-member-access');
      expect(result.stdout).toContain('fixed 1 issue');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('nothing fixable → all diagnostics shown', async () => {
    const tempDir = await createFixTestDir(
      { '@typescript-eslint/no-unsafe-member-access': 'error' },
      { 'index.ts': 'const z: any = {};\nz.foo;\nexport { z };\n' },
    );

    try {
      const result = await runRslint(['--fix', 'index.ts'], tempDir);
      // All diagnostics should be shown (nothing was fixed)
      expect(result.stdout).toContain('no-unsafe-member-access');
      expect(result.stdout).not.toContain('fixed');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('without --fix → all diagnostics shown', async () => {
    const tempDir = await createFixTestDir(
      { '@typescript-eslint/no-unnecessary-type-assertion': 'error' },
      {
        'index.ts':
          "const x: string = 'hello';\nconst y = (x as string).trim();\nexport { y };\n",
      },
    );

    try {
      const result = await runRslint(['index.ts'], tempDir);
      // Without --fix, all diagnostics should be shown
      expect(result.stdout).toContain('no-unnecessary-type-assertion');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('--fix --quiet → only remaining errors, not warnings', async () => {
    const tempDir = await createFixTestDir(
      {
        '@typescript-eslint/no-unnecessary-type-assertion': 'warn',
        '@typescript-eslint/no-unsafe-member-access': 'error',
      },
      {
        'index.ts':
          "const x: string = 'hello';\nconst y = (x as string).trim();\nconst z: any = {};\nz.foo;\nexport { y, z };\n",
      },
    );

    try {
      const result = await runRslint(['--fix', '--quiet', 'index.ts'], tempDir);
      // Error diagnostic should be shown
      expect(result.stdout).toContain('no-unsafe-member-access');
      // Warning was fixed AND quiet mode suppresses warnings
      expect(result.stdout).not.toContain('no-unnecessary-type-assertion');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('cascade --fix → neither rule shown after fix', async () => {
    const tempDir = await createFixTestDir(
      {
        '@typescript-eslint/ban-types': 'error',
        '@typescript-eslint/no-inferrable-types': 'error',
      },
      {
        'index.ts':
          "const a: String = 'hello';\nconst b: Number = 42;\nexport { a, b };\n",
      },
    );

    try {
      const result = await runRslint(['--fix', 'index.ts'], tempDir);
      // Both rules' diagnostics should NOT appear (all fixed via cascade)
      expect(result.stdout).not.toContain('ban-types');
      expect(result.stdout).not.toContain('no-inferrable-types');
      expect(result.stdout).toContain('Found 0 errors');
      expect(result.stdout).toContain('fixed 4 issues');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

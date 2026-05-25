import { describe, test, expect } from '@rstest/core';
import {
  runRslint,
  createFixTestDir,
  cleanupTempDir,
  path,
  fs,
} from './helpers';

describe('CLI --fix edge cases', () => {
  test('empty file → exit 0, no crash', async () => {
    const tempDir = await createFixTestDir(
      { '@typescript-eslint/no-unnecessary-type-assertion': 'error' },
      { 'index.ts': '' },
    );

    try {
      const result = await runRslint(['--fix', 'index.ts'], tempDir);
      expect(result.exitCode).toBe(0);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('syntax errors → no crash', async () => {
    const tempDir = await createFixTestDir(
      { '@typescript-eslint/no-unnecessary-type-assertion': 'error' },
      { 'index.ts': 'const x = \nfunction (\n' },
    );

    try {
      const result = await runRslint(['--fix', 'index.ts'], tempDir);
      expect(result.exitCode).toBeDefined();
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('multiple files both get fixed', async () => {
    const tempDir = await createFixTestDir(
      { '@typescript-eslint/no-unnecessary-type-assertion': 'error' },
      {
        'a.ts':
          "const a: string = 'hello';\nconst ra = (a as string).trim();\nexport { ra };\n",
        'b.ts':
          "const b: string = 'world';\nconst rb = (b as string).trim();\nexport { rb };\n",
      },
    );

    try {
      const result = await runRslint(['--fix'], tempDir);

      const contentA = await fs.readFile(path.join(tempDir, 'a.ts'), 'utf8');
      const contentB = await fs.readFile(path.join(tempDir, 'b.ts'), 'utf8');
      expect(contentA).not.toContain('as string');
      expect(contentB).not.toContain('as string');
      expect(result.exitCode).toBe(0);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('fix output shows correct count', async () => {
    const tempDir = await createFixTestDir(
      { '@typescript-eslint/no-unnecessary-type-assertion': 'error' },
      {
        'index.ts':
          "const x: string = 'a';\nconst r1 = (x as string).trim();\nconst r2 = (x as string).toUpperCase();\nexport { r1, r2 };\n",
      },
    );

    try {
      const result = await runRslint(['--fix', 'index.ts'], tempDir);
      expect(result.exitCode).toBe(0);
      // 2 fixes: two type assertions (no-inferrable-types not in config)
      expect(result.stdout).toContain('fixed 2 issues');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

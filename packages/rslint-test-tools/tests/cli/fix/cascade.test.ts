import { describe, test, expect } from '@rstest/core';
import {
  runRslint,
  createFixTestDir,
  cleanupTempDir,
  path,
  fs,
} from './helpers';

describe('CLI --fix multi-pass cascade', () => {
  test('ban-types → no-inferrable-types fixed in single run', async () => {
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
      expect(result.exitCode).toBe(0);

      const content = await fs.readFile(path.join(tempDir, 'index.ts'), 'utf8');
      // Pass 1: String → string, Number → number (ban-types)
      // Pass 2: `: string` removed, `: number` removed (no-inferrable-types)
      expect(content).not.toContain(': String');
      expect(content).not.toContain(': string');
      expect(content).not.toContain(': Number');
      expect(content).not.toContain(': number');
      expect(result.stdout).toContain('fixed 4 issues');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

import { describe, test, expect } from '@rstest/core';
import {
  runRslint,
  createTempDir,
  cleanupTempDir,
  TS_CONFIG,
} from './helpers.js';

interface Diagnostic {
  ruleName: string;
  filePath: string;
  severity: string;
}

async function lintAndParse(
  files: Record<string, string>,
  args: string[] = [],
): Promise<{ diagnostics: Diagnostic[]; cleanup: () => Promise<void> }> {
  const tempDir = await createTempDir(files);
  const result = await runRslint(['--format', 'jsonline', ...args], tempDir);
  const lines = result.stdout
    .trim()
    .split('\n')
    .filter(l => l.trim());
  const diagnostics = lines.map(l => JSON.parse(l) as Diagnostic);
  return { diagnostics, cleanup: () => cleanupTempDir(tempDir) };
}

function diagsFor(diagnostics: Diagnostic[], pathPart: string): Diagnostic[] {
  return diagnostics.filter(
    d => d.filePath === pathPart || d.filePath.startsWith(pathPart + '/'),
  );
}

function ruleNames(diagnostics: Diagnostic[]): string[] {
  return diagnostics.map(d => d.ruleName);
}

describe('Ignore negation: ! re-include patterns', () => {
  test('global ignore with negation re-includes specific file', async () => {
    const { diagnostics, cleanup } = await lintAndParse({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': `export default [
        { ignores: ['build/**/*', '!build/keep.ts'] },
        {
          files: ['**/*.ts'],
          plugins: ['@typescript-eslint'],
          languageOptions: {
            parserOptions: { projectService: false, project: ['./tsconfig.json'] },
          },
          rules: { '@typescript-eslint/no-explicit-any': 'error' },
        },
      ];`,
      'build/keep.ts': `const x: any = 1;\n`,
      'build/other.ts': `const y: any = 2;\n`,
      'src/index.ts': `const z: any = 3;\n`,
    });
    try {
      // build/keep.ts re-included → linted
      expect(diagsFor(diagnostics, 'build/keep.ts').length).toBeGreaterThan(0);
      // build/other.ts still ignored → not linted
      expect(diagsFor(diagnostics, 'build/other.ts').length).toBe(0);
      // src/index.ts always linted
      expect(diagsFor(diagnostics, 'src/index.ts').length).toBeGreaterThan(0);
    } finally {
      await cleanup();
    }
  });

  test('entry-level ignore with negation re-includes files', async () => {
    const { diagnostics, cleanup } = await lintAndParse({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': `export default [
        {
          files: ['**/*.ts'],
          ignores: ['vendor/**/*', '!vendor/keep/**/*'],
          rules: { 'no-debugger': 'error' },
        },
      ];`,
      'vendor/keep/src/b.ts': `debugger;\n`,
      'vendor/lib/src/a.ts': `debugger;\n`,
      'src/index.ts': `debugger;\n`,
    });
    try {
      // vendor/keep re-included → no-debugger fires
      expect(ruleNames(diagsFor(diagnostics, 'vendor/keep'))).toContain(
        'no-debugger',
      );
      // vendor/lib still ignored by entry
      expect(diagsFor(diagnostics, 'vendor/lib').length).toBe(0);
      // src always linted
      expect(diagsFor(diagnostics, 'src/index.ts').length).toBeGreaterThan(0);
    } finally {
      await cleanup();
    }
  });

  test('cross-entry global ignore negation works', async () => {
    const { diagnostics, cleanup } = await lintAndParse({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': `export default [
        { ignores: ['build/**/*'] },
        { ignores: ['!build/keep.ts'] },
        {
          files: ['**/*.ts'],
          rules: { 'no-debugger': 'error' },
        },
      ];`,
      'build/keep.ts': `debugger;\n`,
      'build/other.ts': `debugger;\n`,
      'src/index.ts': `debugger;\n`,
    });
    try {
      // build/keep.ts re-included by second ignore entry
      expect(ruleNames(diagsFor(diagnostics, 'build/keep.ts'))).toContain(
        'no-debugger',
      );
      // build/other.ts still ignored
      expect(diagsFor(diagnostics, 'build/other.ts').length).toBe(0);
      // src always linted
      expect(diagsFor(diagnostics, 'src/index.ts').length).toBeGreaterThan(0);
    } finally {
      await cleanup();
    }
  });

  test('dir/** with negation: file-level negation still works even with dir pattern', async () => {
    // Note: dir/** blocks directory traversal in JS-side config discovery,
    // but at the Go-side file matching level, negation works normally
    // because files already in tsconfig Programs are matched directly.
    const { diagnostics, cleanup } = await lintAndParse({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': `export default [
        { ignores: ['build/**', '!build/keep.ts'] },
        {
          files: ['**/*.ts'],
          rules: { 'no-debugger': 'error' },
        },
      ];`,
      'build/keep.ts': `debugger;\n`,
      'build/other.ts': `debugger;\n`,
      'src/index.ts': `debugger;\n`,
    });
    try {
      // build/keep.ts re-included by negation at file level
      expect(ruleNames(diagsFor(diagnostics, 'build/keep.ts'))).toContain(
        'no-debugger',
      );
      // build/other.ts still ignored
      expect(diagsFor(diagnostics, 'build/other.ts').length).toBe(0);
      // src always linted
      expect(diagsFor(diagnostics, 'src/index.ts').length).toBeGreaterThan(0);
    } finally {
      await cleanup();
    }
  });

  test('multiple negation patterns re-include multiple files', async () => {
    const { diagnostics, cleanup } = await lintAndParse({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': `export default [
        { ignores: ['generated/**/*', '!generated/types.ts', '!generated/constants.ts'] },
        {
          files: ['**/*.ts'],
          rules: { 'no-debugger': 'error' },
        },
      ];`,
      'generated/types.ts': `debugger;\n`,
      'generated/constants.ts': `debugger;\n`,
      'generated/other.ts': `debugger;\n`,
      'src/index.ts': `debugger;\n`,
    });
    try {
      expect(ruleNames(diagsFor(diagnostics, 'generated/types.ts'))).toContain(
        'no-debugger',
      );
      expect(
        ruleNames(diagsFor(diagnostics, 'generated/constants.ts')),
      ).toContain('no-debugger');
      expect(diagsFor(diagnostics, 'generated/other.ts').length).toBe(0);
      expect(diagsFor(diagnostics, 'src/index.ts').length).toBeGreaterThan(0);
    } finally {
      await cleanup();
    }
  });
});

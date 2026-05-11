import { describe, test, expect } from '@rstest/core';
import { lintJsonline, diagsAt, rules } from './helpers.js';

// Files-driven lint: ignore dimension — directory-level ignore blocking in
// gap discovery, and tsconfig-include vs global-ignore overlap.

describe('Files-driven lint: directory-level ignore blocking in gap discovery', () => {
  test('dir/** blocks gap file discovery, negation cannot re-include', async () => {
    // tsconfig only includes src/ → build/ files are gap files
    // build/** blocks directory → !build/keep.ts has no effect
    const { diagnostics, cleanup } = await lintJsonline({
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['src/**/*.ts'],
      }),
      'rslint.config.mjs': `export default [
        { ignores: ['build/**', '!build/keep.ts'] },
        {
          files: ['**/*.ts'],
          rules: { 'no-debugger': 'error' },
        },
      ];`,
      'src/index.ts': `debugger;\n`,
      'build/keep.ts': `debugger;\n`,
      'build/other.ts': `debugger;\n`,
    });
    try {
      // src/index.ts linted (in tsconfig)
      expect(diagsAt(diagnostics, 'src/index.ts').length).toBeGreaterThan(0);
      // build/ entirely blocked — gap files NOT discovered
      expect(diagsAt(diagnostics, 'build').length).toBe(0);
    } finally {
      await cleanup();
    }
  });

  test('dir/**/* allows gap file discovery, negation re-includes', async () => {
    // build/**/* is file-level → gap files discoverable → ! re-includes keep.ts
    const { diagnostics, cleanup } = await lintJsonline({
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['src/**/*.ts'],
      }),
      'rslint.config.mjs': `export default [
        { ignores: ['build/**/*', '!build/keep.ts'] },
        {
          files: ['**/*.ts'],
          rules: { 'no-debugger': 'error' },
        },
      ];`,
      'src/index.ts': `debugger;\n`,
      'build/keep.ts': `debugger;\n`,
      'build/other.ts': `debugger;\n`,
    });
    try {
      // src/index.ts linted (in tsconfig)
      expect(diagsAt(diagnostics, 'src/index.ts').length).toBeGreaterThan(0);
      // build/keep.ts re-included by negation → discovered as gap file → linted
      expect(rules(diagsAt(diagnostics, 'build/keep.ts'))).toContain(
        'no-debugger',
      );
      // build/other.ts still ignored → not discovered
      expect(diagsAt(diagnostics, 'build/other.ts').length).toBe(0);
    } finally {
      await cleanup();
    }
  });
});

describe('Files-driven lint: tsconfig include + global ignore overlap', () => {
  test('file in tsconfig AND in global ignore is NOT linted', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['**/*.ts'],
      }),
      'rslint.config.mjs': `export default [
        { ignores: ['build/**'] },
        { files: ['**/*.ts'], rules: { 'no-debugger': 'error' } },
      ];`,
      'build/test.ts': `debugger;\n`,
      'src/index.ts': `debugger;\n`,
    });
    try {
      expect(diagsAt(diagnostics, 'build').length).toBe(0);
      expect(diagsAt(diagnostics, 'src/index.ts').length).toBeGreaterThan(0);
    } finally {
      await cleanup();
    }
  });

  test('file in tsconfig + dir/** ignore: ! negation blocked by directory-level ignore', async () => {
    // build/** is directory-level → blocks entirely, even for tsconfig files
    // ! negation cannot undo dir/** (aligned with ESLint v10)
    const { diagnostics, cleanup } = await lintJsonline({
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['**/*.ts'],
      }),
      'rslint.config.mjs': `export default [
        { ignores: ['build/**', '!build/keep.ts'] },
        { files: ['**/*.ts'], rules: { 'no-debugger': 'error' } },
      ];`,
      'build/keep.ts': `debugger;\n`,
      'build/other.ts': `debugger;\n`,
      'src/index.ts': `debugger;\n`,
    });
    try {
      // build/** blocks directory → both files ignored, ! has no effect
      expect(diagsAt(diagnostics, 'build').length).toBe(0);
      expect(diagsAt(diagnostics, 'src/index.ts').length).toBeGreaterThan(0);
    } finally {
      await cleanup();
    }
  });

  test('file in tsconfig + dir/**/* ignore: ! negation works (file-level)', async () => {
    // build/**/* is file-level → ! CAN re-include
    const { diagnostics, cleanup } = await lintJsonline({
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['**/*.ts'],
      }),
      'rslint.config.mjs': `export default [
        { ignores: ['build/**/*', '!build/keep.ts'] },
        { files: ['**/*.ts'], rules: { 'no-debugger': 'error' } },
      ];`,
      'build/keep.ts': `debugger;\n`,
      'build/other.ts': `debugger;\n`,
      'src/index.ts': `debugger;\n`,
    });
    try {
      // keep.ts re-included by ! at file level → linted
      expect(rules(diagsAt(diagnostics, 'build/keep.ts'))).toContain(
        'no-debugger',
      );
      // other.ts still ignored
      expect(diagsAt(diagnostics, 'build/other.ts').length).toBe(0);
      expect(diagsAt(diagnostics, 'src/index.ts').length).toBeGreaterThan(0);
    } finally {
      await cleanup();
    }
  });
});

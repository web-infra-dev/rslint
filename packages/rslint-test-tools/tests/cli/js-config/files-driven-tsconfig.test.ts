import { describe, test, expect } from '@rstest/core';
import { runRslint, createTempDir, cleanupTempDir } from './helpers.js';

// Files-driven lint: tsconfig/program dimension — multiple tsconfigs,
// recommended + user-override cascade, and parser/program edge cases.

describe('Files-driven lint: multiple tsconfigs', () => {
  test('file in one tsconfig but not another should get type-aware rules', async () => {
    const tempDir = await createTempDir({
      'tsconfig.app.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['src/**/*.ts'],
      }),
      'tsconfig.test.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['test/**/*.ts'],
      }),
      'rslint.config.mjs': `export default [
        {
          files: ['**/*.ts'],
          languageOptions: {
            parserOptions: {
              projectService: false,
              project: ['./tsconfig.app.json', './tsconfig.test.json'],
            },
          },
          rules: {
            '@typescript-eslint/ban-ts-comment': 'error',
            '@typescript-eslint/no-unsafe-member-access': 'error',
          },
          plugins: ['@typescript-eslint'],
        },
      ];`,
      // In tsconfig.app → type-aware rules
      'src/index.ts': `// @ts-ignore\nlet a: any = 1;\na.b = 2;\n`,
      // In tsconfig.test → type-aware rules
      'test/test.ts': `// @ts-ignore\nlet b: any = 1;\nb.c = 2;\n`,
      // NOT in any tsconfig → gap file → syntax only
      'scripts/build.ts': `// @ts-ignore\nlet x: any = 1;\nx.y = 2;\n`,
    });
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);

      const lines = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim());
      const diagnostics = lines.map((l) => JSON.parse(l));

      // src/index.ts → all rules
      const srcRules = diagnostics
        .filter((d: any) => d.filePath.includes('src/index.ts'))
        .map((d: any) => d.ruleName);
      expect(srcRules).toContain('@typescript-eslint/no-unsafe-member-access');

      // test/test.ts → all rules
      const testRules = diagnostics
        .filter((d: any) => d.filePath.includes('test/test.ts'))
        .map((d: any) => d.ruleName);
      expect(testRules).toContain('@typescript-eslint/no-unsafe-member-access');

      // scripts/build.ts → syntax only
      const scriptRules = diagnostics
        .filter((d: any) => d.filePath.includes('scripts/build.ts'))
        .map((d: any) => d.ruleName);
      expect(scriptRules).toContain('@typescript-eslint/ban-ts-comment');
      expect(scriptRules).not.toContain(
        '@typescript-eslint/no-unsafe-member-access',
      );
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

describe('Files-driven lint: recommended + user override cascade', () => {
  // Simulates the real-world config pattern:
  //   Entry 0: { ignores: [...] }
  //   Entry 1: ts.configs.recommended → files: ['**/*.ts', '**/*.tsx', '**/*.mts', '**/*.cts']
  //   Entry 2: user override → files: ['**/*.ts', '**/*.tsx'] (narrower)

  test('gap .ts file matched by both recommended and user entry should get cascaded syntax rules', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['src/**/*.ts'],
      }),
      // Simulate: recommended entry + user override entry
      'rslint.config.mjs': `export default [
        {
          files: ['**/*.ts', '**/*.tsx', '**/*.mts', '**/*.cts'],
          plugins: ['@typescript-eslint'],
          languageOptions: {
            parserOptions: { projectService: false, project: ['./tsconfig.json'] },
          },
          rules: {
            '@typescript-eslint/ban-ts-comment': 'error',
            '@typescript-eslint/no-explicit-any': 'error',
            '@typescript-eslint/no-unsafe-member-access': 'error',
          },
        },
        {
          files: ['**/*.ts', '**/*.tsx'],
          rules: {
            '@typescript-eslint/no-explicit-any': 'warn',
          },
        },
      ];`,
      'src/index.ts': `// @ts-ignore\nlet a: any = 1;\na.b = 2;\n`,
      // Gap file matched by BOTH entries → cascade: no-explicit-any should be 'warn' (overridden)
      'scripts/build.ts': `// @ts-ignore\nlet x: any = 1;\nx.y = 2;\n`,
    });
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);

      const lines = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim());
      const diagnostics = lines.map((l) => JSON.parse(l));

      const gapDiags = diagnostics.filter((d: any) =>
        d.filePath.includes('scripts/build.ts'),
      );
      const gapRules = gapDiags.map((d: any) => d.ruleName);

      // Syntax rules should fire
      expect(gapRules).toContain('@typescript-eslint/ban-ts-comment');
      // no-explicit-any is syntax-level (RequiresTypeInfo=false) but overridden to 'warn' by entry 2
      const anyDiag = gapDiags.find(
        (d: any) => d.ruleName === '@typescript-eslint/no-explicit-any',
      );
      if (anyDiag) {
        expect(anyDiag.severity).toBe('warn');
      }
      // Type-aware rules should NOT fire
      expect(gapRules).not.toContain(
        '@typescript-eslint/no-unsafe-member-access',
      );
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('.mts gap file matched only by recommended (not user entry) should get only recommended syntax rules', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['src/**/*.ts'],
      }),
      'rslint.config.mjs': `export default [
        {
          files: ['**/*.ts', '**/*.tsx', '**/*.mts', '**/*.cts'],
          plugins: ['@typescript-eslint'],
          languageOptions: {
            parserOptions: { projectService: false, project: ['./tsconfig.json'] },
          },
          rules: {
            '@typescript-eslint/ban-ts-comment': 'error',
            '@typescript-eslint/no-unsafe-member-access': 'error',
          },
        },
        {
          files: ['**/*.ts', '**/*.tsx'],
          rules: {
            '@typescript-eslint/no-explicit-any': 'error',
          },
        },
      ];`,
      'src/index.ts': `const a = 1;\n`,
      // .mts gap file — matched by entry 1 (recommended) but NOT entry 2 (user)
      'scripts/utils.mts': `// @ts-ignore\nlet x: any = 1;\nx.y = 2;\n`,
    });
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);

      const lines = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim());
      const diagnostics = lines.map((l) => JSON.parse(l));

      const mtsDiags = diagnostics.filter((d: any) =>
        d.filePath.includes('scripts/utils.mts'),
      );
      const mtsRules = mtsDiags.map((d: any) => d.ruleName);

      // Should get ban-ts-comment from entry 1 (syntax, recommended)
      expect(mtsRules).toContain('@typescript-eslint/ban-ts-comment');
      // Should NOT get no-explicit-any — entry 2 doesn't match .mts files
      expect(mtsRules).not.toContain('@typescript-eslint/no-explicit-any');
      // Should NOT get type-aware rules
      expect(mtsRules).not.toContain(
        '@typescript-eslint/no-unsafe-member-access',
      );
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('.mts file in tsconfig and matched only by recommended should get all recommended rules including type-aware', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['src/**/*.ts', 'src/**/*.mts'],
      }),
      'rslint.config.mjs': `export default [
        {
          files: ['**/*.ts', '**/*.tsx', '**/*.mts', '**/*.cts'],
          plugins: ['@typescript-eslint'],
          languageOptions: {
            parserOptions: { projectService: false, project: ['./tsconfig.json'] },
          },
          rules: {
            '@typescript-eslint/ban-ts-comment': 'error',
            '@typescript-eslint/no-unsafe-member-access': 'error',
          },
        },
        {
          files: ['**/*.ts', '**/*.tsx'],
          rules: {
            '@typescript-eslint/no-explicit-any': 'error',
          },
        },
      ];`,
      // .mts file IN tsconfig → has type info → should get type-aware rules from entry 1
      'src/utils.mts': `// @ts-ignore\nlet a: any = 1;\na.b = 2;\n`,
    });
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);

      const lines = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim());
      const diagnostics = lines.map((l) => JSON.parse(l));

      const mtsDiags = diagnostics.filter((d: any) =>
        d.filePath.includes('src/utils.mts'),
      );
      const mtsRules = mtsDiags.map((d: any) => d.ruleName);

      // In tsconfig + matched by entry 1 → ALL entry 1 rules including type-aware
      expect(mtsRules).toContain('@typescript-eslint/ban-ts-comment');
      expect(mtsRules).toContain('@typescript-eslint/no-unsafe-member-access');
      // NOT matched by entry 2 (no .mts in files) → no-explicit-any should NOT fire
      expect(mtsRules).not.toContain('@typescript-eslint/no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('real-world monorepo: root scripts are gap files, package src files have type info', async () => {
    // Simulates the rslint project's own structure:
    // - packages/app/tsconfig.json includes packages/app/src/
    // - packages/lib/tsconfig.json includes packages/lib/src/
    // - root-level scripts/ has no tsconfig
    // - config uses recommended + user overrides with project: ['packages/*/tsconfig.json']
    const tempDir = await createTempDir({
      'packages/app/tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['src/**/*.ts'],
      }),
      'packages/lib/tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['src/**/*.ts'],
      }),
      'rslint.config.mjs': `export default [
        { ignores: ['node_modules/**', '**/dist/**'] },
        {
          files: ['**/*.ts', '**/*.tsx', '**/*.mts', '**/*.cts'],
          plugins: ['@typescript-eslint'],
          languageOptions: {
            parserOptions: {
              projectService: false,
              project: ['./packages/*/tsconfig.json'],
            },
          },
          rules: {
            '@typescript-eslint/ban-ts-comment': 'error',
            '@typescript-eslint/no-unsafe-member-access': 'error',
          },
        },
      ];`,
      // Package files → in tsconfig → all rules
      'packages/app/src/index.ts': `// @ts-ignore\nlet a: any = 1;\na.b = 2;\n`,
      'packages/lib/src/utils.ts': `// @ts-ignore\nlet b: any = 1;\nb.c = 2;\n`,
      // Root script → gap file → syntax rules only
      'scripts/deploy.ts': `// @ts-ignore\nlet x: any = 1;\nx.y = 2;\n`,
    });
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);

      const lines = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim());
      const diagnostics = lines.map((l) => JSON.parse(l));

      // packages/app/src/index.ts → all rules
      const appRules = diagnostics
        .filter((d: any) => d.filePath.includes('packages/app/src/index.ts'))
        .map((d: any) => d.ruleName);
      expect(appRules).toContain('@typescript-eslint/ban-ts-comment');
      expect(appRules).toContain('@typescript-eslint/no-unsafe-member-access');

      // packages/lib/src/utils.ts → all rules
      const libRules = diagnostics
        .filter((d: any) => d.filePath.includes('packages/lib/src/utils.ts'))
        .map((d: any) => d.ruleName);
      expect(libRules).toContain('@typescript-eslint/ban-ts-comment');
      expect(libRules).toContain('@typescript-eslint/no-unsafe-member-access');

      // scripts/deploy.ts → gap file → syntax only
      const scriptRules = diagnostics
        .filter((d: any) => d.filePath.includes('scripts/deploy.ts'))
        .map((d: any) => d.ruleName);
      expect(scriptRules).toContain('@typescript-eslint/ban-ts-comment');
      expect(scriptRules).not.toContain(
        '@typescript-eslint/no-unsafe-member-access',
      );
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('lint-staged scenario: explicit file args with mix of tsconfig and gap files', async () => {
    // Simulates: git commit triggers lint-staged, which passes changed files to rslint
    // Some files are in tsconfig (src/), some are not (scripts/, config files)
    const tempDir = await createTempDir({
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['src/**/*.ts'],
      }),
      'rslint.config.mjs': `export default [
        {
          files: ['**/*.ts'],
          plugins: ['@typescript-eslint'],
          languageOptions: {
            parserOptions: {
              projectService: false,
              project: ['./tsconfig.json'],
            },
          },
          rules: {
            '@typescript-eslint/ban-ts-comment': 'error',
            '@typescript-eslint/no-unsafe-member-access': 'error',
          },
        },
      ];`,
      'src/feature.ts': `// @ts-ignore\nlet a: any = 1;\na.b = 2;\n`,
      'scripts/migrate.ts': `// @ts-ignore\nlet x: any = 1;\nx.y = 2;\n`,
    });
    try {
      // lint-staged passes both files
      const result = await runRslint(
        ['--format', 'jsonline', 'src/feature.ts', 'scripts/migrate.ts'],
        tempDir,
      );

      const lines = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim());
      const diagnostics = lines.map((l) => JSON.parse(l));

      // src/feature.ts → in tsconfig → all rules
      const srcRules = diagnostics
        .filter((d: any) => d.filePath.includes('src/feature.ts'))
        .map((d: any) => d.ruleName);
      expect(srcRules).toContain('@typescript-eslint/ban-ts-comment');
      expect(srcRules).toContain('@typescript-eslint/no-unsafe-member-access');

      // scripts/migrate.ts → gap → syntax only
      const scriptRules = diagnostics
        .filter((d: any) => d.filePath.includes('scripts/migrate.ts'))
        .map((d: any) => d.ruleName);
      expect(scriptRules).toContain('@typescript-eslint/ban-ts-comment');
      expect(scriptRules).not.toContain(
        '@typescript-eslint/no-unsafe-member-access',
      );
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('type-aware rule turned off by user override should not fire even on tsconfig files', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['src/**/*.ts'],
      }),
      'rslint.config.mjs': `export default [
        {
          files: ['**/*.ts', '**/*.tsx', '**/*.mts', '**/*.cts'],
          plugins: ['@typescript-eslint'],
          languageOptions: {
            parserOptions: { projectService: false, project: ['./tsconfig.json'] },
          },
          rules: {
            '@typescript-eslint/no-unsafe-member-access': 'error',
            '@typescript-eslint/ban-ts-comment': 'error',
          },
        },
        {
          files: ['**/*.ts'],
          rules: {
            '@typescript-eslint/no-unsafe-member-access': 'off',
          },
        },
      ];`,
      // In tsconfig, but user turned off no-unsafe-member-access for **/*.ts
      'src/index.ts': `// @ts-ignore\nlet a: any = 1;\na.b = 2;\n`,
    });
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);

      const lines = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim());
      const diagnostics = lines.map((l) => JSON.parse(l));

      const srcDiags = diagnostics.filter((d: any) =>
        d.filePath.includes('src/index.ts'),
      );
      const srcRules = srcDiags.map((d: any) => d.ruleName);

      // ban-ts-comment should fire (not overridden)
      expect(srcRules).toContain('@typescript-eslint/ban-ts-comment');
      // no-unsafe-member-access should NOT fire (turned off by user override)
      expect(srcRules).not.toContain(
        '@typescript-eslint/no-unsafe-member-access',
      );
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

describe('Files-driven lint: edge cases', () => {
  test('gap file with syntax errors should still be linted (not crash)', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['src/**/*.ts'],
      }),
      'rslint.config.mjs': `export default [
        {
          files: ['**/*.ts'],
          plugins: ['@typescript-eslint'],
          languageOptions: {
            parserOptions: {
              projectService: false,
              project: ['./tsconfig.json'],
            },
          },
          rules: {
            '@typescript-eslint/ban-ts-comment': 'error',
          },
        },
      ];`,
      'src/index.ts': `const a = 1;\n`,
      // Gap file with syntax error — should not crash the entire run
      'scripts/broken.ts': `// @ts-ignore\nconst x = {;\n`,
    });
    try {
      const result = await runRslint([], tempDir);
      // Should not crash (exit code can be 0 or 1, but not a process error)
      expect(result.exitCode).toBeLessThanOrEqual(1);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('pure JS project with no tsconfig should lint with syntax rules', async () => {
    const tempDir = await createTempDir({
      // No tsconfig.json at all
      'rslint.config.mjs': `export default [
        {
          files: ['**/*.js'],
          rules: {
            'no-empty': 'error',
          },
        },
      ];`,
      'index.js': `if (true) {}\n`,
      'utils.js': `const x = 1;\n`,
    });
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);

      const lines = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim());
      const diagnostics = lines.map((l) => JSON.parse(l));

      // index.js should get no-empty
      const indexDiags = diagnostics.filter((d: any) =>
        d.filePath.includes('index.js'),
      );
      expect(indexDiags.length).toBeGreaterThan(0);
      expect(indexDiags[0].ruleName).toBe('no-empty');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('config files pattern matches no files should not crash', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['src/**/*.ts'],
      }),
      'rslint.config.mjs': `export default [
        {
          files: ['**/*.vue'],
          rules: { 'no-empty': 'error' },
        },
      ];`,
      // Only .ts files, no .vue files
      'src/index.ts': `const a = 1;\n`,
    });
    try {
      const result = await runRslint([], tempDir);
      // Should not crash
      expect(result.exitCode).toBeLessThanOrEqual(1);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

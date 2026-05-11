import { describe, test, expect } from '@rstest/core';
import fs from 'node:fs/promises';
import { runRslint, createTempDir, cleanupTempDir } from './helpers.js';

/**
 * Tests for files-driven lint: config `files` controls what gets linted,
 * tsconfig `include` controls which files get type-aware rules.
 */
describe('Files-driven lint: gap file auto-degrade', () => {
  test('file matching config files but NOT in tsconfig include should only get syntax rules', async () => {
    const tempDir = await createTempDir({
      // tsconfig only includes src/
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['src/**/*.ts'],
      }),
      // config files covers **/*.ts (broader than tsconfig include)
      'rslint.config.mjs': `export default [
        {
          files: ['**/*.ts'],
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
          plugins: ['@typescript-eslint'],
        },
      ];`,
      // This file is in tsconfig include → all rules should run
      'src/index.ts': `// @ts-ignore\nlet a: any = 1;\na.b = 2;\n`,
      // This file is NOT in tsconfig include (gap file) → only syntax rules
      'scripts/build.ts': `// @ts-ignore\nlet x: any = 1;\nx.y = 2;\n`,
    });
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);

      const lines = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim());
      const diagnostics = lines.map((l) => JSON.parse(l));

      // src/index.ts should have both ban-ts-comment (syntax) and no-unsafe-member-access (type-aware)
      const srcDiags = diagnostics.filter((d: any) =>
        d.filePath.includes('src/index.ts'),
      );
      const srcRules = srcDiags.map((d: any) => d.ruleName);
      expect(srcRules).toContain('@typescript-eslint/ban-ts-comment');
      expect(srcRules).toContain('@typescript-eslint/no-unsafe-member-access');

      // scripts/build.ts should ONLY have ban-ts-comment (syntax), NOT no-unsafe-member-access (type-aware)
      const scriptDiags = diagnostics.filter((d: any) =>
        d.filePath.includes('scripts/build.ts'),
      );
      const scriptRules = scriptDiags.map((d: any) => d.ruleName);
      expect(scriptRules).toContain('@typescript-eslint/ban-ts-comment');
      expect(scriptRules).not.toContain(
        '@typescript-eslint/no-unsafe-member-access',
      );
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('file in tsconfig include but NOT matching config files should NOT be linted', async () => {
    const tempDir = await createTempDir({
      // tsconfig includes both src/ and test/
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['src/**/*.ts', 'test/**/*.ts'],
      }),
      // config files only covers src/
      'rslint.config.mjs': `export default [
        {
          files: ['src/**/*.ts'],
          languageOptions: {
            parserOptions: {
              projectService: false,
              project: ['./tsconfig.json'],
            },
          },
          rules: {
            '@typescript-eslint/ban-ts-comment': 'error',
          },
          plugins: ['@typescript-eslint'],
        },
      ];`,
      'src/index.ts': `// @ts-ignore\nconst a = 1;\n`,
      'test/test.ts': `// @ts-ignore\nconst b = 2;\n`,
    });
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);

      const lines = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim());
      const diagnostics = lines.map((l) => JSON.parse(l));

      // src/index.ts should be linted
      const srcDiags = diagnostics.filter((d: any) =>
        d.filePath.includes('src/index.ts'),
      );
      expect(srcDiags.length).toBeGreaterThan(0);

      // test/test.ts should NOT be linted (not matching config files)
      const testDiags = diagnostics.filter((d: any) =>
        d.filePath.includes('test/test.ts'),
      );
      expect(testDiags.length).toBe(0);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('CLI explicit file arg not in tsconfig should get syntax rules only', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['src/**/*.ts'],
      }),
      'rslint.config.mjs': `export default [
        {
          files: ['**/*.ts'],
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
          plugins: ['@typescript-eslint'],
        },
      ];`,
      'src/index.ts': `const a = 1;\n`,
      'scripts/build.ts': `// @ts-ignore\nlet x: any = 1;\nx.y = 2;\n`,
    });
    try {
      // Explicitly lint scripts/build.ts
      const result = await runRslint(
        ['--format', 'jsonline', 'scripts/build.ts'],
        tempDir,
      );

      const lines = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim());
      const diagnostics = lines.map((l) => JSON.parse(l));
      const rules = diagnostics.map((d: any) => d.ruleName);

      // Should get syntax rules
      expect(rules).toContain('@typescript-eslint/ban-ts-comment');
      // Should NOT get type-aware rules
      expect(rules).not.toContain('@typescript-eslint/no-unsafe-member-access');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('JSON config without files field should use legacy behavior (tsconfig-driven)', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['src/**/*.ts'],
      }),
      // JSON config: no files field
      'rslint.json': JSON.stringify([
        {
          languageOptions: {
            parserOptions: {
              projectService: false,
              project: ['./tsconfig.json'],
            },
          },
          rules: {
            '@typescript-eslint/no-unsafe-member-access': 'error',
          },
          plugins: ['@typescript-eslint'],
        },
      ]),
      'src/index.ts': `let a: any = 1;\na.b = 2;\n`,
      // This file is NOT in tsconfig include — should NOT be linted in legacy mode
      'scripts/build.ts': `let x: any = 1;\nx.y = 2;\n`,
    });
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);

      const lines = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim());
      const diagnostics = lines.map((l) => JSON.parse(l));

      // src/index.ts should be linted (in tsconfig)
      const srcDiags = diagnostics.filter((d: any) =>
        d.filePath.includes('src/index.ts'),
      );
      expect(srcDiags.length).toBeGreaterThan(0);

      // scripts/build.ts should NOT be linted (legacy mode: tsconfig-driven)
      const scriptDiags = diagnostics.filter((d: any) =>
        d.filePath.includes('scripts/build.ts'),
      );
      expect(scriptDiags.length).toBe(0);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('gap file with --type-check should not get semantic diagnostics', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['src/**/*.ts'],
      }),
      'rslint.config.mjs': `export default [
        {
          files: ['**/*.ts'],
          languageOptions: {
            parserOptions: {
              projectService: false,
              project: ['./tsconfig.json'],
            },
          },
          rules: {},
          plugins: ['@typescript-eslint'],
        },
      ];`,
      // Gap file with type error
      'scripts/build.ts': `const x: number = 'hello';\n`,
    });
    try {
      const result = await runRslint(
        ['--format', 'jsonline', '--type-check'],
        tempDir,
      );

      const lines = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim());

      // Gap files should NOT get TypeScript semantic diagnostics
      for (const line of lines) {
        const d = JSON.parse(line);
        if (d.filePath.includes('scripts/build.ts')) {
          expect(d.ruleName).not.toMatch(/^TypeScript\(/);
        }
      }
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

describe('Files-driven lint: file type support', () => {
  test('.tsx gap file should have JSX parsed correctly and get syntax rules', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': JSON.stringify({
        compilerOptions: {
          target: 'ES2020',
          module: 'ESNext',
          strict: true,
          jsx: 'react-jsx',
        },
        include: ['src/**/*.tsx'],
      }),
      'rslint.config.mjs': `export default [
        {
          files: ['**/*.tsx'],
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
          plugins: ['@typescript-eslint'],
        },
      ];`,
      'src/App.tsx': `// @ts-ignore\nconst App = () => <div />;\nexport default App;\n`,
      // Gap .tsx file — should parse JSX and run syntax rules
      'components/Button.tsx': `// @ts-ignore\nconst Button = () => <button />;\nexport default Button;\n`,
    });
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);

      const lines = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim());
      const diagnostics = lines.map((l) => JSON.parse(l));

      // Gap .tsx file should get syntax rules (JSX parsed correctly)
      const gapDiags = diagnostics.filter((d: any) =>
        d.filePath.includes('components/Button.tsx'),
      );
      const gapRules = gapDiags.map((d: any) => d.ruleName);
      expect(gapRules).toContain('@typescript-eslint/ban-ts-comment');
      // Should NOT get type-aware rules
      expect(gapRules).not.toContain(
        '@typescript-eslint/no-unsafe-member-access',
      );
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('.js gap file should be linted with syntax rules', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['src/**/*.ts'],
      }),
      'rslint.config.mjs': `export default [
        {
          files: ['**/*.ts', '**/*.js'],
          languageOptions: {
            parserOptions: {
              projectService: false,
              project: ['./tsconfig.json'],
            },
          },
          rules: {
            'no-empty': 'error',
          },
          plugins: ['@typescript-eslint'],
        },
      ];`,
      'src/index.ts': `const a = 1;\n`,
      // Gap .js file
      'scripts/build.js': `if (true) {}\n`,
    });
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);

      const lines = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim());
      const diagnostics = lines.map((l) => JSON.parse(l));

      // Gap .js file should get no-empty (syntax rule)
      const jsDiags = diagnostics.filter((d: any) =>
        d.filePath.includes('scripts/build.js'),
      );
      expect(jsDiags.length).toBeGreaterThan(0);
      expect(jsDiags[0].ruleName).toBe('no-empty');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

describe('Files-driven lint: flat config semantics', () => {
  test('entry without files should apply to gap files (universal entry)', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['src/**/*.ts'],
      }),
      'rslint.config.mjs': `export default [
        ${JSON.stringify({
          rules: { 'no-empty': 'error' },
        })},
        {
          files: ['**/*.ts'],
          languageOptions: {
            parserOptions: {
              projectService: false,
              project: ['./tsconfig.json'],
            },
          },
          rules: {
            '@typescript-eslint/ban-ts-comment': 'error',
          },
          plugins: ['@typescript-eslint'],
        },
      ];`,
      'src/index.ts': `// @ts-ignore\nif (true) {}\n`,
      // Gap file should get rules from BOTH entries:
      // - no-empty from universal entry (no files)
      // - ban-ts-comment from **/*.ts entry
      'scripts/build.ts': `// @ts-ignore\nif (true) {}\n`,
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

      expect(gapRules).toContain('no-empty');
      expect(gapRules).toContain('@typescript-eslint/ban-ts-comment');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('gap file matching global ignores should NOT be linted', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['src/**/*.ts'],
      }),
      'rslint.config.mjs': `export default [
        { ignores: ['scripts/**'] },
        {
          files: ['**/*.ts'],
          languageOptions: {
            parserOptions: {
              projectService: false,
              project: ['./tsconfig.json'],
            },
          },
          rules: {
            '@typescript-eslint/ban-ts-comment': 'error',
          },
          plugins: ['@typescript-eslint'],
        },
      ];`,
      'src/index.ts': `// @ts-ignore\nconst a = 1;\n`,
      // This gap file matches **/*.ts but is in global ignores
      'scripts/build.ts': `// @ts-ignore\nconst x = 1;\n`,
    });
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);

      const lines = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim());
      const diagnostics = lines.map((l) => JSON.parse(l));

      // scripts/build.ts should be globally ignored
      const scriptDiags = diagnostics.filter((d: any) =>
        d.filePath.includes('scripts/build.ts'),
      );
      expect(scriptDiags.length).toBe(0);

      // src/index.ts should still be linted
      const srcDiags = diagnostics.filter((d: any) =>
        d.filePath.includes('src/index.ts'),
      );
      expect(srcDiags.length).toBeGreaterThan(0);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

describe('Files-driven lint: CLI interaction', () => {
  test('--fix should work on gap files', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['src/**/*.ts'],
      }),
      'rslint.config.mjs': `export default [
        {
          files: ['**/*.ts'],
          languageOptions: {
            parserOptions: {
              projectService: false,
              project: ['./tsconfig.json'],
            },
          },
          rules: {
            '@typescript-eslint/no-inferrable-types': 'error',
          },
          plugins: ['@typescript-eslint'],
        },
      ];`,
      'src/index.ts': `const a = 1;\n`,
      // Gap file with fixable issue
      'scripts/build.ts': `const x: number = 42;\n`,
    });
    try {
      await runRslint(['--fix'], tempDir);

      // Read the fixed file
      const fs = await import('node:fs/promises');
      const fixed = await fs.readFile(`${tempDir}/scripts/build.ts`, 'utf8');
      // no-inferrable-types should have removed the type annotation
      expect(fixed).not.toContain(': number');
      expect(fixed).toContain('const x = 42');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('directory arg should scope gap file discovery', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['src/**/*.ts'],
      }),
      'rslint.config.mjs': `export default [
        {
          files: ['**/*.ts'],
          languageOptions: {
            parserOptions: {
              projectService: false,
              project: ['./tsconfig.json'],
            },
          },
          rules: {
            '@typescript-eslint/ban-ts-comment': 'error',
          },
          plugins: ['@typescript-eslint'],
        },
      ];`,
      'src/index.ts': `const a = 1;\n`,
      'scripts/build.ts': `// @ts-ignore\nconst x = 1;\n`,
      'tools/gen.ts': `// @ts-ignore\nconst y = 1;\n`,
    });
    try {
      // Only lint scripts/ directory
      const result = await runRslint(
        ['--format', 'jsonline', 'scripts/'],
        tempDir,
      );

      const lines = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim());
      const diagnostics = lines.map((l) => JSON.parse(l));

      // scripts/build.ts should be linted
      const scriptDiags = diagnostics.filter((d: any) =>
        d.filePath.includes('scripts/build.ts'),
      );
      expect(scriptDiags.length).toBeGreaterThan(0);

      // tools/gen.ts should NOT be linted (outside dir arg scope)
      const toolsDiags = diagnostics.filter((d: any) =>
        d.filePath.includes('tools/gen.ts'),
      );
      expect(toolsDiags.length).toBe(0);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

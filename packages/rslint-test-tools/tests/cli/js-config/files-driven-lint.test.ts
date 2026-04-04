import { describe, test, expect } from '@rstest/core';
import { runRslint, createTempDir, cleanupTempDir } from './helpers.js';

interface Diagnostic {
  ruleName: string;
  filePath: string;
  severity: string;
}

async function lintJsonline(
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

function diagsAt(diagnostics: Diagnostic[], pathPart: string): Diagnostic[] {
  return diagnostics.filter(
    d => d.filePath === pathPart || d.filePath.startsWith(pathPart + '/'),
  );
}

function rules(diagnostics: Diagnostic[]): string[] {
  return diagnostics.map(d => d.ruleName);
}

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
        .filter(l => l.trim());
      const diagnostics = lines.map(l => JSON.parse(l));

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
        .filter(l => l.trim());
      const diagnostics = lines.map(l => JSON.parse(l));

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
        .filter(l => l.trim());
      const diagnostics = lines.map(l => JSON.parse(l));
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
        .filter(l => l.trim());
      const diagnostics = lines.map(l => JSON.parse(l));

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
        .filter(l => l.trim());

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
        .filter(l => l.trim());
      const diagnostics = lines.map(l => JSON.parse(l));

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
        .filter(l => l.trim());
      const diagnostics = lines.map(l => JSON.parse(l));

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
        .filter(l => l.trim());
      const diagnostics = lines.map(l => JSON.parse(l));

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
        .filter(l => l.trim());
      const diagnostics = lines.map(l => JSON.parse(l));

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
        .filter(l => l.trim());
      const diagnostics = lines.map(l => JSON.parse(l));

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
        .filter(l => l.trim());
      const diagnostics = lines.map(l => JSON.parse(l));

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
        .filter(l => l.trim());
      const diagnostics = lines.map(l => JSON.parse(l));

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
        .filter(l => l.trim());
      const diagnostics = lines.map(l => JSON.parse(l));

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
        .filter(l => l.trim());
      const diagnostics = lines.map(l => JSON.parse(l));

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
        .filter(l => l.trim());
      const diagnostics = lines.map(l => JSON.parse(l));

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
        .filter(l => l.trim());
      const diagnostics = lines.map(l => JSON.parse(l));

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
        .filter(l => l.trim());
      const diagnostics = lines.map(l => JSON.parse(l));

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
        .filter(l => l.trim());
      const diagnostics = lines.map(l => JSON.parse(l));

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

// Monorepo multi-config + tsconfig matrix tests are in files-driven-monorepo.test.ts

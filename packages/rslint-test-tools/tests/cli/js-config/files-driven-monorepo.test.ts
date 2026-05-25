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
    .filter((l) => l.trim());
  const diagnostics = lines.map((l) => JSON.parse(l) as Diagnostic);
  return { diagnostics, cleanup: () => cleanupTempDir(tempDir) };
}

function diagsAt(diagnostics: Diagnostic[], pathPart: string): Diagnostic[] {
  return diagnostics.filter(
    (d) => d.filePath === pathPart || d.filePath.startsWith(pathPart + '/'),
  );
}

function rules(diagnostics: Diagnostic[]): string[] {
  return diagnostics.map((d) => d.ruleName);
}

describe('Monorepo multi-config: ownership dedup', () => {
  test('B1: root+child configs, each has own tsconfig — no duplicate violations', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      'rslint.config.mjs': `export default [{ files: ["*.ts"], rules: { "no-console": "error" } }];`,
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['*.ts'],
      }),
      'root.ts': `console.log("test");\n`,
      'packages/app/rslint.config.mjs': `export default [{ files: ["**/*.ts"], rules: { "no-debugger": "error" } }];`,
      'packages/app/tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['src/**/*.ts'],
      }),
      'packages/app/src/index.ts': `debugger;\nconsole.log("test");\n`,
    });
    try {
      const rootDiags = diagsAt(diagnostics, 'root.ts');
      expect(rules(rootDiags)).toContain('no-console');

      const appDiags = diagsAt(diagnostics, 'packages/app/src/index.ts');
      expect(appDiags.filter((d) => d.ruleName === 'no-debugger').length).toBe(
        1,
      );
      expect(rules(appDiags)).not.toContain('no-console');
    } finally {
      await cleanup();
    }
  });

  test('B2: root+child configs, only root has tsconfig — no duplicate violations', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      'rslint.config.mjs': `export default [{ files: ["**/*.ts"], rules: { "no-console": "error" } }];`,
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['**/*.ts'],
      }),
      'root.ts': `console.log("test");\n`,
      'packages/app/rslint.config.mjs': `export default [{ files: ["**/*.ts"], rules: { "no-debugger": "error" } }];`,
      'packages/app/src/index.ts': `debugger;\nconsole.log("test");\n`,
    });
    try {
      expect(rules(diagsAt(diagnostics, 'root.ts'))).toContain('no-console');

      const appDiags = diagsAt(diagnostics, 'packages/app/src/index.ts');
      expect(appDiags.filter((d) => d.ruleName === 'no-debugger').length).toBe(
        1,
      );
      expect(rules(appDiags)).not.toContain('no-console');
    } finally {
      await cleanup();
    }
  });

  test('B3: root+child configs, no tsconfig anywhere — no duplicate violations', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      'rslint.config.mjs': `export default [{ files: ["*.ts"], rules: { "no-console": "error" } }];`,
      'root.ts': `console.log("test");\n`,
      'packages/app/rslint.config.mjs': `export default [{ files: ["**/*.ts"], rules: { "no-debugger": "error" } }];`,
      'packages/app/src/index.ts': `debugger;\nconsole.log("test");\n`,
    });
    try {
      expect(rules(diagsAt(diagnostics, 'root.ts'))).toContain('no-console');

      const appDiags = diagsAt(diagnostics, 'packages/app/src/index.ts');
      expect(appDiags.filter((d) => d.ruleName === 'no-debugger').length).toBe(
        1,
      );
      expect(rules(appDiags)).not.toContain('no-console');
    } finally {
      await cleanup();
    }
  });

  test('B5: root+child configs, child files broader than tsconfig — gap files not duplicated', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      'rslint.config.mjs': `export default [{ files: ["*.ts"], rules: { "no-console": "error" } }];`,
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['*.ts'],
      }),
      'root.ts': `console.log("test");\n`,
      'packages/app/rslint.config.mjs': `export default [{ files: ["**/*.ts"], rules: { "no-debugger": "error" } }];`,
      'packages/app/tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['src/**/*.ts'],
      }),
      'packages/app/src/index.ts': `debugger;\n`,
      'packages/app/scripts/build.ts': `debugger;\n`,
    });
    try {
      expect(rules(diagsAt(diagnostics, 'root.ts'))).toContain('no-console');
      expect(
        diagsAt(diagnostics, 'packages/app/src/index.ts').filter(
          (d) => d.ruleName === 'no-debugger',
        ).length,
      ).toBe(1);
      expect(
        diagsAt(diagnostics, 'packages/app/scripts/build.ts').filter(
          (d) => d.ruleName === 'no-debugger',
        ).length,
      ).toBe(1);
    } finally {
      await cleanup();
    }
  });

  test('C1: sibling configs with tsconfig — isolated, no cross-contamination', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      'packages/a/rslint.config.mjs': `export default [{ files: ["**/*.ts"], rules: { "no-console": "error" } }];`,
      'packages/a/tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['src/**/*.ts'],
      }),
      'packages/a/src/index.ts': `console.log("test");\n`,
      'packages/b/rslint.config.mjs': `export default [{ files: ["**/*.ts"], rules: { "no-debugger": "error" } }];`,
      'packages/b/tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['src/**/*.ts'],
      }),
      'packages/b/src/index.ts': `debugger;\n`,
    });
    try {
      const aDiags = diagsAt(diagnostics, 'packages/a/src/index.ts');
      expect(rules(aDiags)).toContain('no-console');
      expect(rules(aDiags)).not.toContain('no-debugger');

      const bDiags = diagsAt(diagnostics, 'packages/b/src/index.ts');
      expect(rules(bDiags)).toContain('no-debugger');
      expect(rules(bDiags)).not.toContain('no-console');
    } finally {
      await cleanup();
    }
  });

  test('C2: sibling configs without tsconfig — isolated, no duplicates', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      'packages/a/rslint.config.mjs': `export default [{ files: ["**/*.ts"], rules: { "no-console": "error" } }];`,
      'packages/a/src/index.ts': `console.log("test");\n`,
      'packages/b/rslint.config.mjs': `export default [{ files: ["**/*.ts"], rules: { "no-debugger": "error" } }];`,
      'packages/b/src/index.ts': `debugger;\n`,
    });
    try {
      const aDiags = diagsAt(diagnostics, 'packages/a/src/index.ts');
      expect(rules(aDiags)).toContain('no-console');
      expect(aDiags.filter((d) => d.ruleName === 'no-console').length).toBe(1);

      const bDiags = diagsAt(diagnostics, 'packages/b/src/index.ts');
      expect(rules(bDiags)).toContain('no-debugger');
      expect(bDiags.filter((d) => d.ruleName === 'no-debugger').length).toBe(1);
    } finally {
      await cleanup();
    }
  });
});

describe('Monorepo multi-config: CLI invocation variants', () => {
  test('B2-file: specify child file — uses child config, no duplicate, no spurious warnings', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': `export default [{ files: ["**/*.ts"], rules: { "no-console": "error" } }];`,
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['**/*.ts'],
      }),
      'root.ts': `console.log("test");\n`,
      'packages/app/rslint.config.mjs': `export default [{ files: ["**/*.ts"], rules: { "no-debugger": "error" } }];`,
      'packages/app/src/index.ts': `debugger;\nconsole.log("test");\n`,
    });
    try {
      const result = await runRslint(
        ['--format', 'jsonline', 'packages/app/src/index.ts'],
        tempDir,
      );
      const diagnostics = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim())
        .map((l) => JSON.parse(l) as Diagnostic);

      expect(
        diagnostics.filter((d) => d.ruleName === 'no-debugger').length,
      ).toBe(1);
      expect(
        diagnostics.filter((d) => d.ruleName === 'no-console').length,
      ).toBe(0);
      // --start-time should not leak as a spurious "not found" warning
      expect(result.stderr).not.toContain('start-time');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('B3-dir: specify child directory — uses child config, no duplicate', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': `export default [{ files: ["*.ts"], rules: { "no-console": "error" } }];`,
      'root.ts': `console.log("test");\n`,
      'packages/app/rslint.config.mjs': `export default [{ files: ["**/*.ts"], rules: { "no-debugger": "error" } }];`,
      'packages/app/src/index.ts': `debugger;\nconsole.log("test");\n`,
    });
    try {
      const result = await runRslint(
        ['--format', 'jsonline', 'packages/app/'],
        tempDir,
      );
      const diagnostics = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim())
        .map((l) => JSON.parse(l) as Diagnostic);

      expect(
        diagnostics.filter((d) => d.ruleName === 'no-debugger').length,
      ).toBe(1);
      expect(
        diagnostics.filter((d) => d.ruleName === 'no-console').length,
      ).toBe(0);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('lint-staged: file args spanning multiple configs', async () => {
    // Simulates: lint-staged passes files from different packages
    const tempDir = await createTempDir({
      'packages/a/rslint.config.mjs': `export default [{ files: ["**/*.ts"], rules: { "no-console": "error" } }];`,
      'packages/a/src/index.ts': `console.log("test");\n`,
      'packages/b/rslint.config.mjs': `export default [{ files: ["**/*.ts"], rules: { "no-debugger": "error" } }];`,
      'packages/b/src/index.ts': `debugger;\n`,
    });
    try {
      const result = await runRslint(
        [
          '--format',
          'jsonline',
          'packages/a/src/index.ts',
          'packages/b/src/index.ts',
        ],
        tempDir,
      );
      const diagnostics = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim())
        .map((l) => JSON.parse(l) as Diagnostic);

      // Each file uses its nearest config
      const aDiags = diagsAt(diagnostics, 'packages/a/src/index.ts');
      expect(rules(aDiags)).toContain('no-console');
      expect(rules(aDiags)).not.toContain('no-debugger');

      const bDiags = diagsAt(diagnostics, 'packages/b/src/index.ts');
      expect(rules(bDiags)).toContain('no-debugger');
      expect(rules(bDiags)).not.toContain('no-console');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('--fix in multi-config monorepo — no duplicate fixes', async () => {
    // no-inferrable-types has auto-fix (removes redundant type annotations)
    const tempDir = await createTempDir({
      'rslint.config.mjs': `export default [{
        files: ["*.ts"], plugins: ["@typescript-eslint"],
        rules: { "@typescript-eslint/no-inferrable-types": "error" }
      }];`,
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['*.ts'],
      }),
      'root.ts': `const x: number = 1;\n`,
      'packages/app/rslint.config.mjs': `export default [{
        files: ["**/*.ts"], plugins: ["@typescript-eslint"],
        rules: { "@typescript-eslint/no-inferrable-types": "error" }
      }];`,
      'packages/app/tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['src/**/*.ts'],
      }),
      'packages/app/src/index.ts': `const y: number = 2;\n`,
    });
    try {
      await runRslint(['--fix'], tempDir);

      const fs = await import('node:fs/promises');
      const path = await import('node:path');

      // Both files should have type annotation removed
      const rootContent = await fs.readFile(
        path.join(tempDir, 'root.ts'),
        'utf8',
      );
      expect(rootContent).not.toContain(': number');

      const appContent = await fs.readFile(
        path.join(tempDir, 'packages/app/src/index.ts'),
        'utf8',
      );
      expect(appContent).not.toContain(': number');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

describe('Monorepo multi-config: global ignores + config discovery', () => {
  test('B4: root global-ignores child dir — child not linted in no-args', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      'rslint.config.mjs': `export default [
        { ignores: ["packages/ignored/**"] },
        { files: ["**/*.ts"], rules: { "no-console": "error" } }
      ];`,
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['src/**/*.ts'],
      }),
      'src/a.ts': `console.log("test");\n`,
      'packages/ignored/rslint.config.mjs': `export default [{ files: ["**/*.ts"], rules: { "no-debugger": "error" } }];`,
      'packages/ignored/tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['src/**/*.ts'],
      }),
      'packages/ignored/src/index.ts': `debugger;\nconsole.log("test");\n`,
    });
    try {
      expect(rules(diagsAt(diagnostics, 'src/a.ts'))).toContain('no-console');
      expect(diagsAt(diagnostics, 'packages/ignored').length).toBe(0);
    } finally {
      await cleanup();
    }
  });

  test('B4-file: specify file in globally-ignored child dir — uses child config', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': `export default [
        { ignores: ["packages/ignored/**"] },
        { files: ["**/*.ts"], rules: { "no-console": "error" } }
      ];`,
      'packages/ignored/rslint.config.mjs': `export default [{ files: ["**/*.ts"], rules: { "no-debugger": "error" } }];`,
      'packages/ignored/src/index.ts': `debugger;\nconsole.log("test");\n`,
    });
    try {
      const result = await runRslint(
        ['--format', 'jsonline', 'packages/ignored/src/index.ts'],
        tempDir,
      );
      const diagnostics = result.stdout
        .trim()
        .split('\n')
        .filter((l) => l.trim())
        .map((l) => JSON.parse(l) as Diagnostic);

      expect(
        diagnostics.filter((d) => d.ruleName === 'no-debugger').length,
      ).toBe(1);
      expect(
        diagnostics.filter((d) => d.ruleName === 'no-console').length,
      ).toBe(0);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

describe('Monorepo multi-config: type-aware rules + tsconfig', () => {
  test('B6: root type-aware, child syntax-only — isolation', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      'rslint.config.mjs': `export default [{
        files: ["src/**/*.ts"],
        plugins: ["@typescript-eslint"],
        languageOptions: { parserOptions: { projectService: false, project: ["./tsconfig.json"] } },
        rules: { "@typescript-eslint/ban-ts-comment": "error", "@typescript-eslint/no-unsafe-member-access": "error" }
      }];`,
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['src/**/*.ts'],
      }),
      'src/index.ts': `// @ts-ignore\nlet a: any = 1;\na.b = 2;\n`,
      'packages/app/rslint.config.mjs': `export default [{ files: ["**/*.ts"], rules: { "no-debugger": "error" } }];`,
      'packages/app/src/index.ts': `// @ts-ignore\ndebugger;\nlet x: any = 1;\nx.y = 2;\n`,
    });
    try {
      const rootRules = rules(diagsAt(diagnostics, 'src/index.ts'));
      expect(rootRules).toContain('@typescript-eslint/ban-ts-comment');
      expect(rootRules).toContain('@typescript-eslint/no-unsafe-member-access');

      const appDiags = diagsAt(diagnostics, 'packages/app/src/index.ts');
      expect(rules(appDiags)).toContain('no-debugger');
      expect(rules(appDiags)).not.toContain(
        '@typescript-eslint/no-unsafe-member-access',
      );
      expect(rules(appDiags)).not.toContain(
        '@typescript-eslint/ban-ts-comment',
      );
    } finally {
      await cleanup();
    }
  });

  test('B2-typeaware: root tsconfig covers child, child no tsconfig — child loses type info', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      'rslint.config.mjs': `export default [{
        files: ["**/*.ts"],
        plugins: ["@typescript-eslint"],
        languageOptions: { parserOptions: { projectService: false, project: ["./tsconfig.json"] } },
        rules: {
          "@typescript-eslint/ban-ts-comment": "error",
          "@typescript-eslint/no-unsafe-member-access": "error",
        }
      }];`,
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['**/*.ts'],
      }),
      'src/index.ts': `// @ts-ignore\nlet a: any = 1;\na.b = 2;\n`,
      'packages/app/rslint.config.mjs': `export default [{ files: ["**/*.ts"], rules: { "no-debugger": "error" } }];`,
      'packages/app/src/index.ts': `// @ts-ignore\ndebugger;\nlet x: any = 1;\nx.y = 2;\n`,
    });
    try {
      const rootRules = rules(diagsAt(diagnostics, 'src/index.ts'));
      expect(rootRules).toContain('@typescript-eslint/ban-ts-comment');
      expect(rootRules).toContain('@typescript-eslint/no-unsafe-member-access');

      const appDiags = diagsAt(diagnostics, 'packages/app/src/index.ts');
      expect(rules(appDiags)).toContain('no-debugger');
      expect(rules(appDiags)).not.toContain(
        '@typescript-eslint/no-unsafe-member-access',
      );
      expect(rules(appDiags)).not.toContain(
        '@typescript-eslint/ban-ts-comment',
      );
      expect(appDiags.filter((d) => d.ruleName === 'no-debugger').length).toBe(
        1,
      );
    } finally {
      await cleanup();
    }
  });

  test('B2-typeaware-child: each has own tsconfig + type-aware — isolated', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      'rslint.config.mjs': `export default [{
        files: ["**/*.ts"],
        plugins: ["@typescript-eslint"],
        languageOptions: { parserOptions: { projectService: false, project: ["./tsconfig.json"] } },
        rules: { "@typescript-eslint/no-unsafe-member-access": "error" }
      }];`,
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['src/**/*.ts'],
      }),
      'src/index.ts': `let a: any = 1;\na.b = 2;\n`,
      'packages/app/rslint.config.mjs': `export default [{
        files: ["**/*.ts"],
        plugins: ["@typescript-eslint"],
        languageOptions: { parserOptions: { projectService: false, project: ["./tsconfig.json"] } },
        rules: { "@typescript-eslint/no-unsafe-member-access": "error" }
      }];`,
      'packages/app/tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['src/**/*.ts'],
      }),
      'packages/app/src/index.ts': `let x: any = 1;\nx.y = 2;\n`,
    });
    try {
      expect(rules(diagsAt(diagnostics, 'src/index.ts'))).toContain(
        '@typescript-eslint/no-unsafe-member-access',
      );

      const appDiags = diagsAt(diagnostics, 'packages/app/src/index.ts');
      expect(rules(appDiags)).toContain(
        '@typescript-eslint/no-unsafe-member-access',
      );
      expect(
        appDiags.filter(
          (d) => d.ruleName === '@typescript-eslint/no-unsafe-member-access',
        ).length,
      ).toBe(1);
    } finally {
      await cleanup();
    }
  });
});

describe('Monorepo multi-config: real-world scenarios', () => {
  test('recommended preset + monorepo with multiple packages', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      'rslint.config.mjs': `export default [
        { ignores: ["**/dist/**"] },
        {
          files: ["**/*.ts"],
          plugins: ["@typescript-eslint"],
          languageOptions: { parserOptions: { projectService: false, project: ["./packages/*/tsconfig.json"] } },
          rules: {
            "@typescript-eslint/ban-ts-comment": "error",
            "@typescript-eslint/no-unsafe-member-access": "error",
            "no-debugger": "error",
          }
        }
      ];`,
      'packages/app/tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['src/**/*.ts'],
      }),
      'packages/lib/tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['src/**/*.ts'],
      }),
      'packages/app/src/index.ts': `// @ts-ignore\nlet a: any = 1;\na.b = 2;\n`,
      'packages/lib/src/utils.ts': `// @ts-ignore\nlet b: any = 1;\nb.c = 2;\n`,
      'scripts/deploy.ts': `// @ts-ignore\nlet x: any = 1;\nx.y = 2;\n`,
      'packages/app/dist/bundle.ts': `debugger;\n`,
    });
    try {
      // Package files → type-aware rules
      expect(
        rules(diagsAt(diagnostics, 'packages/app/src/index.ts')),
      ).toContain('@typescript-eslint/no-unsafe-member-access');
      expect(
        rules(diagsAt(diagnostics, 'packages/lib/src/utils.ts')),
      ).toContain('@typescript-eslint/no-unsafe-member-access');

      // Gap file → syntax only
      const scriptRules = rules(diagsAt(diagnostics, 'scripts/deploy.ts'));
      expect(scriptRules).toContain('@typescript-eslint/ban-ts-comment');
      expect(scriptRules).not.toContain(
        '@typescript-eslint/no-unsafe-member-access',
      );

      // dist ignored
      expect(diagsAt(diagnostics, 'dist').length).toBe(0);

      // No duplicates anywhere
      const allFiles = [...new Set(diagnostics.map((d) => d.filePath))];
      for (const file of allFiles) {
        const fileDiags = diagnostics.filter((d) => d.filePath === file);
        const ruleSet = new Set(
          fileDiags.map((d) => `${d.ruleName}:${d.filePath}`),
        );
        expect(fileDiags.length).toBe(ruleSet.size);
      }
    } finally {
      await cleanup();
    }
  });

  test('nested node_modules inside package — pruned from gap file discovery', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      'rslint.config.mjs': `export default [{ files: ["**/*.ts"], rules: { "no-console": "error" } }];`,
      'src/index.ts': `console.log("test");\n`,
      'packages/app/node_modules/dep/index.ts': `console.log("dep");\n`,
      'node_modules/root-dep/index.ts': `console.log("root-dep");\n`,
    });
    try {
      expect(diagsAt(diagnostics, 'src/index.ts').length).toBeGreaterThan(0);
      // node_modules at any level should be excluded
      expect(diagsAt(diagnostics, 'node_modules').length).toBe(0);
      expect(diagsAt(diagnostics, 'packages/app/node_modules').length).toBe(0);
    } finally {
      await cleanup();
    }
  });

  test('file args should not produce --start-time spurious warning', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': `export default [{ files: ["**/*.ts"], rules: { "no-console": "error" } }];`,
      'src/a.ts': `console.log("test");\n`,
      'src/b.ts': `console.log("test");\n`,
    });
    try {
      const result = await runRslint(['src/a.ts', 'src/b.ts'], tempDir);
      // No --start-time in warnings
      expect(result.stderr).not.toContain('start-time');
      expect(result.stdout).not.toContain('start-time');
      // Files should still be linted
      expect(result.exitCode).toBe(1);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  // Reproduces the rsbuild issue: single config with parserOptions.project
  // pointing to multiple tsconfigs that have project references between them.
  // Files from referenced projects should not be linted twice.
  test('multi-tsconfig with project references — no duplicate diagnostics', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      'rslint.config.mjs': `export default [
        {
          files: ["**/*.ts"],
          languageOptions: {
            parserOptions: {
              projectService: false,
              project: ["./packages/*/tsconfig.json"],
            },
          },
          rules: { "prefer-const": "error" },
        }
      ];`,
      'packages/core/tsconfig.json': JSON.stringify({
        compilerOptions: {
          target: 'ES2020',
          module: 'ESNext',
          strict: true,
          composite: true,
          outDir: './dist',
          rootDir: 'src',
        },
        include: ['src'],
      }),
      'packages/core/src/lib.ts': `let x = 1;\nexport { x };\n`,
      'packages/plugin/tsconfig.json': JSON.stringify({
        compilerOptions: {
          target: 'ES2020',
          module: 'ESNext',
          strict: true,
          outDir: './dist',
          rootDir: 'src',
        },
        include: ['src'],
        references: [{ path: '../core' }],
      }),
      'packages/plugin/src/index.ts': `import { x } from '../../core/src/lib';\nlet y = x;\nexport { y };\n`,
    });
    try {
      // core/src/lib.ts should have exactly 1 prefer-const error (not duplicated)
      const coreDiags = diagsAt(diagnostics, 'packages/core/src/lib.ts');
      expect(
        coreDiags.filter((d) => d.ruleName === 'prefer-const').length,
      ).toBe(1);

      // plugin/src/index.ts should have exactly 1 prefer-const error
      const pluginDiags = diagsAt(diagnostics, 'packages/plugin/src/index.ts');
      expect(
        pluginDiags.filter((d) => d.ruleName === 'prefer-const').length,
      ).toBe(1);
    } finally {
      await cleanup();
    }
  });

  // Single config with multiple tsconfigs + gap files.
  // The gap file should be linted, but files already covered by tsconfig programs
  // should not be re-linted by the gap program.
  test('multi-tsconfig with gap files — no duplicates', async () => {
    const { diagnostics, cleanup } = await lintJsonline({
      'rslint.config.mjs': `export default [
        {
          files: ["**/*.ts"],
          languageOptions: {
            parserOptions: {
              projectService: false,
              project: ["./packages/core/tsconfig.json"],
            },
          },
          rules: { "no-console": "error" },
        }
      ];`,
      'packages/core/tsconfig.json': JSON.stringify({
        compilerOptions: {
          target: 'ES2020',
          module: 'ESNext',
          strict: true,
        },
        include: ['src'],
      }),
      'packages/core/src/lib.ts': `console.log("core");\n`,
      // gap file: not in any tsconfig
      'scripts/build.ts': `console.log("gap");\n`,
    });
    try {
      // core/src/lib.ts: exactly 1 no-console error
      const coreDiags = diagsAt(diagnostics, 'packages/core/src/lib.ts');
      expect(coreDiags.filter((d) => d.ruleName === 'no-console').length).toBe(
        1,
      );

      // scripts/build.ts (gap file): exactly 1 no-console error
      const scriptDiags = diagsAt(diagnostics, 'scripts/build.ts');
      expect(
        scriptDiags.filter((d) => d.ruleName === 'no-console').length,
      ).toBe(1);
    } finally {
      await cleanup();
    }
  });

  test('specify nonexistent file — warns without crashing', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': `export default [{ files: ["**/*.ts"], rules: { "no-console": "error" } }];`,
      'src/index.ts': `console.log("test");\n`,
    });
    try {
      const result = await runRslint(['nonexistent.ts'], tempDir);
      // Should not crash
      expect(result.exitCode).toBe(0);
      // Should have a warning about the file
      expect(result.stderr + result.stdout).toContain('warning');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

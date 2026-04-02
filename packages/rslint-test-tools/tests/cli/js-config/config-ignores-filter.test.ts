import { describe, test, expect } from '@rstest/core';
import {
  runRslint,
  createTempDir,
  cleanupTempDir,
  TS_CONFIG,
} from './helpers.js';

// --- Helpers to reduce e2e boilerplate ---

interface Diagnostic {
  ruleName: string;
  filePath: string;
  severity: string;
}

/** Run rslint with jsonline format and parse diagnostics. */
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
  return {
    diagnostics,
    cleanup: () => cleanupTempDir(tempDir),
  };
}

/**
 * Filter diagnostics by path prefix. The pathPart must match from the start
 * of the relative file path on a segment boundary.
 *
 * Examples:
 *   diagsFor(d, 'src/index.ts') → matches 'src/index.ts', NOT 'packages/app/src/index.ts'
 *   diagsFor(d, 'packages/app') → matches 'packages/app/src/index.ts'
 *   diagsFor(d, '__tests__')    → matches '__tests__/fixtures/src/test.ts'
 */
function diagsFor(diagnostics: Diagnostic[], pathPart: string): Diagnostic[] {
  return diagnostics.filter(
    d => d.filePath === pathPart || d.filePath.startsWith(pathPart + '/'),
  );
}

/** Extract rule names from diagnostics. */
function ruleNames(diagnostics: Diagnostic[]): string[] {
  return diagnostics.map(d => d.ruleName);
}

/** Common root config with global ignores + TS rules. */
function rootConfig(globalIgnores: string[]): string {
  return `export default [
    { ignores: ${JSON.stringify(globalIgnores)} },
    {
      files: ['**/*.ts'],
      plugins: ['@typescript-eslint'],
      languageOptions: {
        parserOptions: {
          projectService: false,
          project: ['./tsconfig.json'],
        },
      },
      rules: { '@typescript-eslint/no-explicit-any': 'error' },
    },
  ];`;
}

/** Simple nested config with a specific rule. */
function nestedConfig(rule: string): string {
  return `export default [
    { files: ['**/*.ts'], rules: { '${rule}': 'error' } },
  ];`;
}

// --- Tests ---

describe('Config discovery: parent global ignores filter nested configs', () => {
  test('nested config in globally ignored directory should not be used', async () => {
    const { diagnostics, cleanup } = await lintAndParse({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': rootConfig(['__tests__/**']),
      '__tests__/fixtures/rslint.config.mjs': nestedConfig('no-console'),
      'src/index.ts': `const x: any = 1;\n`,
      '__tests__/fixtures/src/test.ts': `console.log('test');\nconst a: any = 1;\n`,
    });
    try {
      expect(diagsFor(diagnostics, 'src/index.ts').length).toBeGreaterThan(0);
      // __tests__ completely ignored — no diagnostics at all
      expect(diagsFor(diagnostics, '__tests__').length).toBe(0);
    } finally {
      await cleanup();
    }
  });

  test('nested config NOT in ignored directory should still work', async () => {
    const { diagnostics, cleanup } = await lintAndParse({
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['packages/*/src/**/*.ts'],
      }),
      'rslint.config.mjs': rootConfig(['**/dist/**']),
      'packages/app/rslint.config.mjs': nestedConfig('no-console'),
      'packages/app/src/index.ts': `console.log('app');\nconst x: any = 1;\n`,
    });
    try {
      const appRules = ruleNames(diagsFor(diagnostics, 'packages/app'));
      // App uses its own nearest config → no-console fires
      expect(appRules).toContain('no-console');
      // Root config's no-explicit-any should NOT apply (app has its own config)
      expect(appRules).not.toContain('@typescript-eslint/no-explicit-any');
    } finally {
      await cleanup();
    }
  });

  test('entry-level ignores should NOT filter nested configs', async () => {
    const { diagnostics, cleanup } = await lintAndParse({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': `export default [
        {
          files: ['**/*.ts'],
          ignores: ['__tests__/**'],
          rules: { 'no-debugger': 'error' },
        },
      ];`,
      '__tests__/fixtures/rslint.config.mjs': nestedConfig('no-console'),
      '__tests__/fixtures/src/test.ts': `console.log('test');\n`,
    });
    try {
      const testRules = ruleNames(diagsFor(diagnostics, '__tests__/fixtures'));
      // Nested config fires (entry-level ignore didn't block discovery)
      expect(testRules).toContain('no-console');
      // Root's no-debugger should NOT apply (file uses nearest config)
      expect(testRules).not.toContain('no-debugger');
    } finally {
      await cleanup();
    }
  });

  test('deeply nested config with **/pattern/** should be filtered', async () => {
    const { diagnostics, cleanup } = await lintAndParse({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': `export default [
        { ignores: ['**/fixtures/**'] },
        { files: ['**/*.ts'], rules: { 'no-debugger': 'error' } },
      ];`,
      'packages/ext/__tests__/fixtures/rslint.config.mjs':
        nestedConfig('no-console'),
      'packages/ext/__tests__/fixtures/src/test.ts': `console.log('deep');\n`,
      'src/index.ts': `debugger;\n`,
    });
    try {
      // Root source linted with no-debugger
      expect(ruleNames(diagsFor(diagnostics, 'src/index.ts'))).toContain(
        'no-debugger',
      );
      // fixtures completely filtered — no diagnostics
      expect(
        diagsFor(diagnostics, 'packages/ext/__tests__/fixtures').length,
      ).toBe(0);
    } finally {
      await cleanup();
    }
  });

  test('single config (no nesting) should not be affected', async () => {
    const { diagnostics, cleanup } = await lintAndParse({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': rootConfig(['dist/**']),
      'src/index.ts': `const x: any = 1;\n`,
    });
    try {
      expect(diagnostics.length).toBeGreaterThan(0);
      expect(diagnostics[0].ruleName).toBe(
        '@typescript-eslint/no-explicit-any',
      );
    } finally {
      await cleanup();
    }
  });

  test('real-world monorepo: root ignores multiple dirs, packages have own configs', async () => {
    const { diagnostics, cleanup } = await lintAndParse({
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['packages/*/src/**/*.ts'],
      }),
      'rslint.config.mjs': rootConfig([
        '**/fixtures/**',
        '**/dist/**',
        'e2e/**',
      ]),
      'packages/app/rslint.config.mjs': nestedConfig('no-console'),
      'packages/lib/rslint.config.mjs': nestedConfig('no-debugger'),
      'packages/ext/__tests__/fixtures/rslint.config.mjs':
        nestedConfig('no-console'),
      'e2e/rslint.config.mjs': nestedConfig('no-console'),
      'packages/app/src/index.ts': `console.log('app');\n`,
      'packages/lib/src/utils.ts': `debugger;\n`,
      'packages/ext/__tests__/fixtures/src/test.ts': `console.log('fix');\n`,
      'e2e/src/test.ts': `console.log('e2e');\n`,
    });
    try {
      const appRules = ruleNames(diagsFor(diagnostics, 'packages/app'));
      const libRules = ruleNames(diagsFor(diagnostics, 'packages/lib'));
      // Each package uses own config
      expect(appRules).toContain('no-console');
      expect(libRules).toContain('no-debugger');
      // Cross-package isolation: app doesn't get lib's rules and vice versa
      expect(appRules).not.toContain('no-debugger');
      expect(libRules).not.toContain('no-console');
      // Ignored dirs: no diagnostics
      expect(
        diagsFor(diagnostics, 'packages/ext/__tests__/fixtures').length,
      ).toBe(0);
      expect(diagsFor(diagnostics, 'e2e').length).toBe(0);
    } finally {
      await cleanup();
    }
  });

  test('wildcard in middle of ignore pattern: packages/*/dist/**', async () => {
    const { diagnostics, cleanup } = await lintAndParse({
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['src/**/*.ts', 'packages/*/src/**/*.ts'],
      }),
      'rslint.config.mjs': `export default [
        { ignores: ['packages/*/dist/**'] },
        {
          files: ['**/*.ts'],
          plugins: ['@typescript-eslint'],
          languageOptions: {
            parserOptions: {
              projectService: false,
              project: ['./tsconfig.json'],
            },
          },
          rules: { '@typescript-eslint/no-explicit-any': 'error' },
        },
      ];`,
      'packages/app/dist/generated/rslint.config.mjs':
        nestedConfig('no-console'),
      'packages/app/dist/generated/src/gen.ts': `console.log('gen');\nconst x: any = 1;\n`,
      'src/index.ts': `const y: any = 1;\n`,
    });
    try {
      expect(diagsFor(diagnostics, 'packages/app/dist').length).toBe(0);
      expect(diagsFor(diagnostics, 'src/index.ts').length).toBeGreaterThan(0);
    } finally {
      await cleanup();
    }
  });

  test('brace expansion in ignore: {__tests__,e2e}/**', async () => {
    const { diagnostics, cleanup } = await lintAndParse({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': `export default [
        { ignores: ['{__tests__,e2e}/**'] },
        {
          files: ['**/*.ts'],
          rules: { 'no-debugger': 'error' },
        },
      ];`,
      '__tests__/fixtures/rslint.config.mjs': nestedConfig('no-console'),
      'e2e/rslint.config.mjs': nestedConfig('no-console'),
      '__tests__/fixtures/src/test.ts': `console.log('test');\n`,
      'e2e/src/e2e.ts': `console.log('e2e');\n`,
      'src/index.ts': `debugger;\n`,
    });
    try {
      expect(diagsFor(diagnostics, '__tests__').length).toBe(0);
      expect(diagsFor(diagnostics, 'e2e').length).toBe(0);
      const srcRules = ruleNames(diagsFor(diagnostics, 'src/index.ts'));
      expect(srcRules).toContain('no-debugger');
      // no-console from ignored configs should NOT leak to root
      expect(srcRules).not.toContain('no-console');
    } finally {
      await cleanup();
    }
  });

  test('specifying file in ignored dir uses nearest config (ESLint v10 aligned)', async () => {
    // ESLint v10: explicit file arg → find nearest config per-file.
    // Even though root ignores __tests__/**, the file uses its own nearest
    // config (__tests__/fixtures/rslint.config.mjs) because file args
    // go through findJSConfigUp, not findJSConfigsInDir.
    const { diagnostics, cleanup } = await lintAndParse(
      {
        'tsconfig.json': TS_CONFIG,
        'rslint.config.mjs': `export default [
        { ignores: ['__tests__/**'] },
        { files: ['**/*.ts'], rules: { 'no-debugger': 'error' } },
      ];`,
        '__tests__/fixtures/rslint.config.mjs': nestedConfig('no-console'),
        '__tests__/fixtures/src/test.ts': `console.log('test');\n`,
        'src/index.ts': `const a = 1;\n`,
      },
      ['__tests__/fixtures/src/test.ts'],
    );
    try {
      const testRules = ruleNames(
        diagsFor(diagnostics, '__tests__/fixtures/src/test.ts'),
      );
      // Uses nearest config (fixtures) → no-console fires
      expect(testRules).toContain('no-console');
      // Root's no-debugger should NOT apply (different config)
      expect(testRules).not.toContain('no-debugger');
    } finally {
      await cleanup();
    }
  });

  test('three levels: root → package → sub-package with ignores at each level', async () => {
    const { diagnostics, cleanup } = await lintAndParse({
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['packages/**/src/**/*.ts'],
      }),
      'rslint.config.mjs': `export default [
        { ignores: ['vendor/**'] },
        { files: ['**/*.ts'], rules: { 'no-debugger': 'error' } },
      ];`,
      'packages/app/rslint.config.mjs': `export default [
        { ignores: ['generated/**'] },
        { files: ['**/*.ts'], rules: { 'no-console': 'error' } },
      ];`,
      'packages/app/generated/rslint.config.mjs': nestedConfig('no-debugger'),
      'vendor/lib/rslint.config.mjs': nestedConfig('no-debugger'),
      'packages/app/src/index.ts': `console.log('app');\n`,
      'packages/app/generated/src/gen.ts': `debugger;\n`,
      'vendor/lib/src/v.ts': `debugger;\n`,
    });
    try {
      const appRules = ruleNames(diagsFor(diagnostics, 'packages/app/src'));
      // app uses nearest config → no-console
      expect(appRules).toContain('no-console');
      // root's no-debugger should NOT apply to app (app has own config)
      expect(appRules).not.toContain('no-debugger');
      // generated and vendor filtered
      expect(diagsFor(diagnostics, 'packages/app/generated').length).toBe(0);
      expect(diagsFor(diagnostics, 'vendor').length).toBe(0);
    } finally {
      await cleanup();
    }
  });

  test('intermediate config ignores apply to its children only', async () => {
    const { diagnostics, cleanup } = await lintAndParse({
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['packages/*/src/**/*.ts'],
      }),
      'rslint.config.mjs': `export default [
        { files: ['**/*.ts'], rules: { 'no-debugger': 'error' } },
      ];`,
      'packages/app/rslint.config.mjs': `export default [
        { ignores: ['generated/**'] },
        { files: ['**/*.ts'], rules: { 'no-console': 'error' } },
      ];`,
      'packages/app/generated/rslint.config.mjs': nestedConfig('no-debugger'),
      'packages/lib/rslint.config.mjs': nestedConfig('no-console'),
      'packages/app/src/index.ts': `console.log('app');\n`,
      'packages/app/generated/src/gen.ts': `debugger;\n`,
      'packages/lib/src/utils.ts': `console.log('lib');\n`,
    });
    try {
      const appRules = ruleNames(diagsFor(diagnostics, 'packages/app/src'));
      const libRules = ruleNames(diagsFor(diagnostics, 'packages/lib'));
      // app: no-console from own config
      expect(appRules).toContain('no-console');
      // app should NOT get root's no-debugger (has own nearest config)
      expect(appRules).not.toContain('no-debugger');
      // lib: no-console from own config
      expect(libRules).toContain('no-console');
      // lib should NOT get root's no-debugger (has own nearest config)
      expect(libRules).not.toContain('no-debugger');
      // generated filtered by app's ignore
      expect(diagsFor(diagnostics, 'packages/app/generated').length).toBe(0);
    } finally {
      await cleanup();
    }
  });

  test('no root config, sibling package configs do not affect each other', async () => {
    const { diagnostics, cleanup } = await lintAndParse({
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['packages/*/src/**/*.ts'],
      }),
      'packages/app/rslint.config.mjs': `export default [
        { ignores: ['generated/**'] },
        { files: ['**/*.ts'], rules: { 'no-console': 'error' } },
      ];`,
      'packages/lib/rslint.config.mjs': `export default [
        { ignores: ['vendor/**'] },
        { files: ['**/*.ts'], rules: { 'no-debugger': 'error' } },
      ];`,
      'packages/app/src/index.ts': `console.log('app');\n`,
      'packages/lib/src/utils.ts': `debugger;\n`,
    });
    try {
      const appRules = ruleNames(diagsFor(diagnostics, 'packages/app/src'));
      const libRules = ruleNames(diagsFor(diagnostics, 'packages/lib/src'));
      // Each gets own rules
      expect(appRules).toContain('no-console');
      expect(libRules).toContain('no-debugger');
      // Cross-isolation: neither gets the other's rules
      expect(appRules).not.toContain('no-debugger');
      expect(libRules).not.toContain('no-console');
    } finally {
      await cleanup();
    }
  });

  test('sibling A global ignore does NOT filter sibling B nested config with same dir name', async () => {
    const { diagnostics, cleanup } = await lintAndParse({
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['packages/**/src/**/*.ts'],
      }),
      'rslint.config.mjs': `export default [
        { files: ['**/*.ts'], rules: { 'no-empty': 'error' } },
      ];`,
      'packages/app/rslint.config.mjs': `export default [
        { ignores: ['generated/**'] },
        { files: ['**/*.ts'], rules: { 'no-console': 'error' } },
      ];`,
      'packages/app/generated/rslint.config.mjs': nestedConfig('no-debugger'),
      'packages/lib/rslint.config.mjs': `export default [
        { files: ['**/*.ts'], rules: { 'no-console': 'error' } },
      ];`,
      'packages/lib/generated/rslint.config.mjs': nestedConfig('no-debugger'),
      'packages/app/src/index.ts': `console.log('app');\n`,
      'packages/app/generated/src/gen.ts': `debugger;\n`,
      'packages/lib/src/utils.ts': `console.log('lib');\n`,
      'packages/lib/generated/src/gen.ts': `debugger;\n`,
    });
    try {
      // app/generated filtered → no diagnostics
      expect(diagsFor(diagnostics, 'packages/app/generated').length).toBe(0);
      // lib/generated NOT filtered → no-debugger from its own config
      const libGenRules = ruleNames(
        diagsFor(diagnostics, 'packages/lib/generated'),
      );
      expect(libGenRules).toContain('no-debugger');
      // lib/generated should NOT get lib parent's no-console (has own nearest config)
      expect(libGenRules).not.toContain('no-console');
      // Both src files use their package config
      expect(ruleNames(diagsFor(diagnostics, 'packages/app/src'))).toContain(
        'no-console',
      );
      expect(ruleNames(diagsFor(diagnostics, 'packages/lib/src'))).toContain(
        'no-console',
      );
    } finally {
      await cleanup();
    }
  });

  test('config with both global and entry-level ignores: only global filters nested configs', async () => {
    const { diagnostics, cleanup } = await lintAndParse({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': `export default [
        { ignores: ['dist/**'] },
        {
          files: ['**/*.ts'],
          ignores: ['test/**'],
          rules: { 'no-debugger': 'error' },
        },
      ];`,
      'dist/generated/rslint.config.mjs': nestedConfig('no-console'),
      'test/fixtures/rslint.config.mjs': nestedConfig('no-console'),
      'dist/generated/src/gen.ts': `console.log('gen');\n`,
      'test/fixtures/src/test.ts': `console.log('test');\n`,
      'src/index.ts': `debugger;\n`,
    });
    try {
      // dist/ filtered by global ignore
      expect(diagsFor(diagnostics, 'dist').length).toBe(0);
      // test/ NOT filtered → nested config runs
      const testRules = ruleNames(diagsFor(diagnostics, 'test/fixtures'));
      expect(testRules).toContain('no-console');
      // test/ should NOT get root's no-debugger (has own nearest config)
      expect(testRules).not.toContain('no-debugger');
      // Root src linted
      const srcRules = ruleNames(diagsFor(diagnostics, 'src/index.ts'));
      expect(srcRules).toContain('no-debugger');
      // Root should NOT get nested config's no-console
      expect(srcRules).not.toContain('no-console');
    } finally {
      await cleanup();
    }
  });

  test('child global ignore does NOT affect parent config', async () => {
    const { diagnostics, cleanup } = await lintAndParse({
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['src/**/*.ts', 'packages/*/src/**/*.ts'],
      }),
      'rslint.config.mjs': `export default [
        { files: ['**/*.ts'], rules: { 'no-debugger': 'error' } },
      ];`,
      'packages/app/rslint.config.mjs': `export default [
        { ignores: ['vendor/**'] },
        { files: ['**/*.ts'], rules: { 'no-console': 'error' } },
      ];`,
      'src/index.ts': `debugger;\n`,
      'packages/app/src/index.ts': `console.log('app');\n`,
    });
    try {
      const rootRules = ruleNames(diagsFor(diagnostics, 'src/index.ts'));
      const appRules = ruleNames(diagsFor(diagnostics, 'packages/app/src'));
      // Root uses root config
      expect(rootRules).toContain('no-debugger');
      // Root should NOT get app's no-console
      expect(rootRules).not.toContain('no-console');
      // App uses own config
      expect(appRules).toContain('no-console');
      // App should NOT get root's no-debugger
      expect(appRules).not.toContain('no-debugger');
    } finally {
      await cleanup();
    }
  });

  test('both siblings ignore same dir name, each only filters own children', async () => {
    const { diagnostics, cleanup } = await lintAndParse({
      'tsconfig.json': JSON.stringify({
        compilerOptions: { target: 'ES2020', module: 'ESNext', strict: true },
        include: ['packages/**/src/**/*.ts'],
      }),
      'rslint.config.mjs': `export default [
        { files: ['**/*.ts'], rules: { 'no-empty': 'error' } },
      ];`,
      'packages/app/rslint.config.mjs': `export default [
        { ignores: ['dist/**'] },
        { files: ['**/*.ts'], rules: { 'no-console': 'error' } },
      ];`,
      'packages/lib/rslint.config.mjs': `export default [
        { ignores: ['dist/**'] },
        { files: ['**/*.ts'], rules: { 'no-debugger': 'error' } },
      ];`,
      'packages/app/dist/gen/rslint.config.mjs': nestedConfig('no-empty'),
      'packages/lib/dist/gen/rslint.config.mjs': nestedConfig('no-empty'),
      'packages/app/src/index.ts': `console.log('app');\n`,
      'packages/lib/src/utils.ts': `debugger;\n`,
      'packages/app/dist/gen/src/a.ts': `if (true) {}\n`,
      'packages/lib/dist/gen/src/b.ts': `if (true) {}\n`,
    });
    try {
      const appRules = ruleNames(diagsFor(diagnostics, 'packages/app/src'));
      const libRules = ruleNames(diagsFor(diagnostics, 'packages/lib/src'));
      // Each package uses own config
      expect(appRules).toContain('no-console');
      expect(libRules).toContain('no-debugger');
      // Cross-isolation
      expect(appRules).not.toContain('no-debugger');
      expect(libRules).not.toContain('no-console');
      // Both dist/ dirs filtered
      expect(diagsFor(diagnostics, 'packages/app/dist').length).toBe(0);
      expect(diagsFor(diagnostics, 'packages/lib/dist').length).toBe(0);
    } finally {
      await cleanup();
    }
  });
});

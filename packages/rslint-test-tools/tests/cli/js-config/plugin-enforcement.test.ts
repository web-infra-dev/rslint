import { describe, test, expect } from '@rstest/core';
import {
  runRslint,
  createTempDir,
  cleanupTempDir,
  TS_CONFIG,
} from './helpers.js';

describe('CLI JS config plugin enforcement', () => {
  test('plugin rule should be blocked when plugin is not declared', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'const x: any = 1;\n',
      // JS config with TS rule but WITHOUT plugins declaration
      'rslint.config.js': `export default [${JSON.stringify({
        files: ['**/*.ts'],
        languageOptions: {
          parserOptions: {
            projectService: false,
            project: ['./tsconfig.json'],
          },
        },
        rules: { '@typescript-eslint/no-explicit-any': 'error' },
        // NOTE: no plugins field
      })}];`,
    });
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      // Without plugins declaration, the TS rule should NOT fire
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('plugin rule should work when plugin is declared', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'const x: any = 1;\n',
      'rslint.config.js': `export default [${JSON.stringify({
        files: ['**/*.ts'],
        plugins: ['@typescript-eslint'],
        languageOptions: {
          parserOptions: {
            projectService: false,
            project: ['./tsconfig.json'],
          },
        },
        rules: { '@typescript-eslint/no-explicit-any': 'error' },
      })}];`,
    });
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      expect(result.stdout).toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('core rules should work without any plugins declaration', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'debugger;\n',
      // JS config with core rule, no plugins field at all
      'rslint.config.js': `export default [${JSON.stringify({
        files: ['**/*.ts'],
        languageOptions: {
          parserOptions: {
            projectService: false,
            project: ['./tsconfig.json'],
          },
        },
        rules: { 'no-debugger': 'error' },
      })}];`,
    });
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      // Core rules should always work regardless of plugins
      expect(result.stdout).toContain('no-debugger');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('multiple plugins in same entry: declared ones work, undeclared ones blocked', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      // Triggers both no-explicit-any and no-debugger
      'test.ts': 'const x: any = 1;\ndebugger;\n',
      'rslint.config.js': `export default [${JSON.stringify({
        files: ['**/*.ts'],
        plugins: ['@typescript-eslint'],
        languageOptions: {
          parserOptions: {
            projectService: false,
            project: ['./tsconfig.json'],
          },
        },
        rules: {
          '@typescript-eslint/no-explicit-any': 'error',
          'no-debugger': 'error',
          // import plugin NOT declared → this rule should be blocked
          'import/no-self-import': 'error',
        },
      })}];`,
    });
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      // Declared plugin rule should fire
      expect(result.stdout).toContain('no-explicit-any');
      // Core rule should fire
      expect(result.stdout).toContain('no-debugger');
      // Undeclared plugin rule should NOT fire
      expect(result.stdout).not.toContain('no-self-import');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('plugin declared in one entry, rule in another entry', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'const x: any = 1;\n',
      'rslint.config.js': `export default [
        ${JSON.stringify({
          plugins: ['@typescript-eslint'],
        })},
        ${JSON.stringify({
          files: ['**/*.ts'],
          languageOptions: {
            parserOptions: {
              projectService: false,
              project: ['./tsconfig.json'],
            },
          },
          rules: { '@typescript-eslint/no-explicit-any': 'error' },
        })}
      ];`,
    });
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      // Plugin from entry1 merged with rule from entry2 → should fire
      expect(result.stdout).toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('plugin entry does not match file → rule blocked', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'const x: any = 1;\n',
      'rslint.config.js': `export default [
        ${JSON.stringify({
          files: ['**/*.jsx'],
          plugins: ['@typescript-eslint'],
        })},
        ${JSON.stringify({
          files: ['**/*.ts'],
          languageOptions: {
            parserOptions: {
              projectService: false,
              project: ['./tsconfig.json'],
            },
          },
          rules: { '@typescript-eslint/no-explicit-any': 'error' },
        })}
      ];`,
    });
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      // Plugin entry only matches .jsx, not .ts → plugin not in merged set → blocked
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('eslint-plugin- prefix should be normalized to match rule prefix', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      // import/no-self-import triggers on `import foo from "./test"`
      'test.ts': 'import foo from "./test";\n',
      'rslint.config.js': `export default [${JSON.stringify({
        files: ['**/*.ts'],
        // "eslint-plugin-import" should normalize to "import" to match "import/" rule prefix
        plugins: ['eslint-plugin-import'],
        languageOptions: {
          parserOptions: {
            projectService: false,
            project: ['./tsconfig.json'],
          },
        },
        rules: { 'import/no-self-import': 'error' },
      })}];`,
    });
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      expect(result.stdout).toContain('no-self-import');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('preset spread + local override should work', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      // Triggers both no-explicit-any and ban-ts-comment
      'test.ts': 'const x: any = 1;\n// @ts-ignore\nconst y = 2;\n',
      'rslint.config.js': `export default [
        ${JSON.stringify({
          files: ['**/*.ts'],
          plugins: ['@typescript-eslint'],
          languageOptions: {
            parserOptions: {
              projectService: false,
              project: ['./tsconfig.json'],
            },
          },
          rules: {
            '@typescript-eslint/no-explicit-any': 'error',
            '@typescript-eslint/ban-ts-comment': 'error',
          },
        })},
        ${JSON.stringify({
          rules: {
            '@typescript-eslint/no-explicit-any': 'off',
          },
        })}
      ];`,
    });
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      // no-explicit-any overridden to off → should NOT fire
      expect(result.stdout).not.toContain('no-explicit-any');
      // ban-ts-comment from preset → should still fire
      expect(result.stdout).toContain('ban-ts-comment');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

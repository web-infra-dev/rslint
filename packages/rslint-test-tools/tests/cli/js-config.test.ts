import { describe, test, expect } from '@rstest/core';
import { spawn } from 'child_process';
import path from 'node:path';
import fs from 'node:fs/promises';
import { tmpdir } from 'node:os';
import {
  findJSConfig,
  normalizeConfig,
  loadConfigFile,
} from '@rslint/core/config-loader';
import { defineConfig } from '@rslint/core';
import configs from '@rslint/core/configs';

const RSLINT_BIN = require.resolve('@rslint/core/bin');

interface CliTestResult {
  exitCode: number;
  stdout: string;
  stderr: string;
}

async function runRslint(args: string[], cwd?: string): Promise<CliTestResult> {
  return new Promise(resolve => {
    const child = spawn(RSLINT_BIN, args, {
      cwd: cwd || process.cwd(),
      stdio: ['pipe', 'pipe', 'pipe'],
    });

    let stdout = '';
    let stderr = '';

    child.stdout?.on('data', (data: Buffer) => {
      stdout += data.toString();
    });

    child.stderr?.on('data', (data: Buffer) => {
      stderr += data.toString();
    });

    child.on('close', code => {
      resolve({ exitCode: code || 0, stdout, stderr });
    });
  });
}

async function createTempDir(files: Record<string, string>): Promise<string> {
  const tempDir = await fs.mkdtemp(path.join(tmpdir(), 'rslint-jsconfig-'));
  for (const [filePath, content] of Object.entries(files)) {
    const fullPath = path.join(tempDir, filePath);
    await fs.mkdir(path.dirname(fullPath), { recursive: true });
    await fs.writeFile(fullPath, content, 'utf8');
  }
  return tempDir;
}

async function cleanupTempDir(tempDir: string): Promise<void> {
  await fs.rm(tempDir, { recursive: true, force: true });
}

const TS_CONFIG = JSON.stringify({
  compilerOptions: {
    target: 'ES2020',
    module: 'ESNext',
    strict: true,
  },
  include: ['**/*.ts'],
});

function jsConfig(overrides: Record<string, unknown> = {}): string {
  const entry = {
    files: ['**/*.ts'],
    languageOptions: {
      parserOptions: {
        projectService: false,
        project: ['./tsconfig.json'],
      },
    },
    rules: { '@typescript-eslint/no-unsafe-member-access': 'error' },
    plugins: ['@typescript-eslint'],
    ...overrides,
  };
  return `export default [${JSON.stringify(entry)}];`;
}

// --- normalizeConfig unit tests ---

describe('normalizeConfig', () => {
  test('should accept a valid flat config array', () => {
    const result = normalizeConfig([
      { files: ['**/*.ts'], rules: { 'no-console': 'error' } },
    ]);
    expect(result).toHaveLength(1);
    expect(result[0].files).toEqual(['**/*.ts']);
    expect(result[0].rules).toEqual({ 'no-console': 'error' });
  });

  test('should throw when config is not an array', () => {
    expect(() => normalizeConfig({ rules: {} })).toThrow(
      'rslint config must export an array',
    );
  });

  test('should strip unknown fields', () => {
    const result = normalizeConfig([
      {
        name: 'my-config',
        files: ['**/*.ts'],
        rules: {},
        unknownField: 123,
      },
    ]);
    expect(result[0]).not.toHaveProperty('name');
    expect(result[0]).not.toHaveProperty('unknownField');
  });

  test('should preserve all known fields', () => {
    const result = normalizeConfig([
      {
        files: ['**/*.ts'],
        ignores: ['dist/**'],
        languageOptions: {
          parserOptions: { project: ['./tsconfig.json'] },
        },
        rules: { 'no-console': 'error' },
        plugins: ['@typescript-eslint'],
        settings: { key: 'value' },
      },
    ]);
    const entry = result[0];
    expect(entry.files).toEqual(['**/*.ts']);
    expect(entry.ignores).toEqual(['dist/**']);
    expect(entry.rules).toEqual({ 'no-console': 'error' });
    expect(entry.plugins).toEqual(['@typescript-eslint']);
    expect(entry.settings).toEqual({ key: 'value' });
  });

  test('should handle empty array', () => {
    expect(normalizeConfig([])).toEqual([]);
  });

  test('should skip null and non-object entries', () => {
    const result = normalizeConfig([
      { files: ['**/*.ts'], rules: { 'no-console': 'error' } },
      null,
      undefined,
      42,
      'string',
    ]);
    expect(result).toHaveLength(1);
    expect(result[0].rules).toEqual({ 'no-console': 'error' });
  });

  test('should throw when files is a string instead of array', () => {
    expect(() => normalizeConfig([{ files: '**/*.ts', rules: {} }])).toThrow(
      '"files" must be an array',
    );
  });

  test('should throw when ignores is a string instead of array', () => {
    expect(() => normalizeConfig([{ ignores: 'dist/**', rules: {} }])).toThrow(
      '"ignores" must be an array',
    );
  });

  test('should allow omitted files and ignores', () => {
    const result = normalizeConfig([{ rules: { 'no-console': 'error' } }]);
    expect(result).toHaveLength(1);
    expect(result[0].files).toBeUndefined();
    expect(result[0].ignores).toBeUndefined();
  });
});

// --- loadConfigFile unit tests ---

describe('loadConfigFile', () => {
  test('should load a .js config file with default export', async () => {
    const tempDir = await createTempDir({
      'rslint.config.js':
        'export default [{ files: ["**/*.ts"], rules: { "no-console": "error" } }];',
    });
    try {
      const result = await loadConfigFile(
        path.join(tempDir, 'rslint.config.js'),
      );
      expect(Array.isArray(result)).toBe(true);
      expect((result as any[])[0].rules).toEqual({ 'no-console': 'error' });
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should load a .mjs config file', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs':
        'export default [{ files: ["**/*.js"], rules: {} }];',
    });
    try {
      const result = await loadConfigFile(
        path.join(tempDir, 'rslint.config.mjs'),
      );
      expect(Array.isArray(result)).toBe(true);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should resolve thenable (Promise) default export', async () => {
    const tempDir = await createTempDir({
      'rslint.config.js':
        'export default Promise.resolve([{ files: ["**/*.ts"], rules: { "no-console": "error" } }]);',
    });
    try {
      const result = await loadConfigFile(
        path.join(tempDir, 'rslint.config.js'),
      );
      expect(Array.isArray(result)).toBe(true);
      expect((result as any[])[0].rules).toEqual({ 'no-console': 'error' });
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should throw for unsupported extension', async () => {
    const tempDir = await createTempDir({
      'rslint.config.yaml': 'rules: {}',
    });
    try {
      await expect(
        loadConfigFile(path.join(tempDir, 'rslint.config.yaml')),
      ).rejects.toThrow('Unsupported config file extension');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

// --- findJSConfig unit tests ---

describe('findJSConfig', () => {
  test('should return null when no config file exists', async () => {
    const tempDir = await createTempDir({ 'test.ts': '' });
    try {
      expect(findJSConfig(tempDir)).toBeNull();
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should find rslint.config.js', async () => {
    const tempDir = await createTempDir({
      'rslint.config.js': 'export default [];',
    });
    try {
      expect(findJSConfig(tempDir)).toContain('rslint.config.js');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should prefer .js over .ts when both exist', async () => {
    const tempDir = await createTempDir({
      'rslint.config.js': 'export default [];',
      'rslint.config.ts': 'export default [];',
    });
    try {
      expect(findJSConfig(tempDir)).toContain('rslint.config.js');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should prefer .mjs over .ts when both exist', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': 'export default [];',
      'rslint.config.ts': 'export default [];',
    });
    try {
      expect(findJSConfig(tempDir)).toContain('rslint.config.mjs');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should prefer .js over .mjs when both exist', async () => {
    const tempDir = await createTempDir({
      'rslint.config.js': 'export default [];',
      'rslint.config.mjs': 'export default [];',
    });
    try {
      expect(findJSConfig(tempDir)).toContain('rslint.config.js');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should find .ts config when no .js/.mjs exists', async () => {
    const tempDir = await createTempDir({
      'rslint.config.ts': 'export default [];',
    });
    try {
      expect(findJSConfig(tempDir)).toContain('rslint.config.ts');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should follow priority order: .js > .mjs > .ts > .mts', async () => {
    const tempDir = await createTempDir({
      'rslint.config.js': 'export default [];',
      'rslint.config.mjs': 'export default [];',
      'rslint.config.ts': 'export default [];',
      'rslint.config.mts': 'export default [];',
    });
    try {
      expect(findJSConfig(tempDir)).toContain('rslint.config.js');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

// --- CLI integration tests for JS config ---

describe('CLI JS config integration', () => {
  test('should auto-detect rslint.config.js', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'let a: any = 10;\na.b = 20;\n',
      'rslint.config.js': jsConfig(),
    });
    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-unsafe-member-access');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should support --config=path (equals format) for JS config', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'let a: any = 10;\na.b = 20;\n',
      'custom.config.js': jsConfig(),
    });
    try {
      const result = await runRslint(['--config=custom.config.js'], tempDir);
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-unsafe-member-access');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should support --config path (space format) for JS config', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'let a: any = 10;\na.b = 20;\n',
      'custom.config.js': jsConfig(),
    });
    try {
      const result = await runRslint(['--config', 'custom.config.js'], tempDir);
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-unsafe-member-access');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should pass through unknown flags to Go binary', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'let a: any = 10;\na.b = 20;\n',
      'rslint.config.js': jsConfig(),
    });
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      const lines = result.stdout
        .trim()
        .split('\n')
        .filter(l => l.trim());
      for (const line of lines) {
        expect(() => JSON.parse(line)).not.toThrow();
      }
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should report error for non-existent JS config via --config', async () => {
    const tempDir = await createTempDir({ 'test.ts': 'const a = 1;\n' });
    try {
      const result = await runRslint(
        ['--config', 'nonexistent.config.js'],
        tempDir,
      );
      expect(result.exitCode).not.toBe(0);
      expect(result.stderr).toContain('not found');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should prefer JS config over JSON config', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'let a: any = 10;\na.b = 20;\n',
      'rslint.json': JSON.stringify([
        {
          files: ['**/*.ts'],
          languageOptions: {
            parserOptions: {
              projectService: false,
              project: ['./tsconfig.json'],
            },
          },
          rules: { '@typescript-eslint/no-explicit-any': 'error' },
          plugins: ['@typescript-eslint'],
        },
      ]),
      'rslint.config.js': jsConfig({
        rules: {
          '@typescript-eslint/no-unsafe-member-access': 'error',
          '@typescript-eslint/no-explicit-any': 'off',
        },
      }),
    });
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      expect(result.stdout).toContain('no-unsafe-member-access');
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should handle JS config with global ignores', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': JSON.stringify({
        compilerOptions: {
          target: 'ES2020',
          module: 'ESNext',
          strict: true,
        },
        include: ['src/**/*.ts', 'dist/**/*.ts'],
      }),
      'src/index.ts': 'let a: any = 10;\na.b = 20;\n',
      'dist/index.ts': 'let a: any = 10;\na.b = 20;\n',
      'rslint.config.js': `export default [
        { ignores: ["dist/**"] },
        ${JSON.stringify({
          files: ['**/*.ts'],
          languageOptions: {
            parserOptions: {
              projectService: false,
              project: ['./tsconfig.json'],
            },
          },
          rules: { '@typescript-eslint/no-unsafe-member-access': 'error' },
          plugins: ['@typescript-eslint'],
        })},
      ];`,
    });
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      expect(result.stdout).not.toContain('dist/index.ts');
      expect(result.stdout).toContain('no-unsafe-member-access');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should preserve -- separator when passing args to Go binary', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'const a = 1;\n',
      'rslint.config.js': jsConfig({
        rules: { '@typescript-eslint/no-explicit-any': 'off' },
      }),
    });
    try {
      // The `--` separator followed by a file path should not be lost.
      // If `--` is dropped, `--looks-like-flag.ts` could be parsed as a flag.
      const result = await runRslint(['--', 'test.ts'], tempDir);
      // We just verify it doesn't crash with an unknown flag error
      expect(result.stderr).not.toContain('unknown flag');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should handle JS config with rule turned off', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'let a: any = 10;\na.b = 20;\n',
      'rslint.config.js': jsConfig({
        rules: {
          '@typescript-eslint/no-unsafe-member-access': 'off',
          '@typescript-eslint/no-explicit-any': 'off',
        },
      }),
    });
    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);
      expect(result.stdout).not.toContain('no-unsafe-member-access');
      expect(result.stdout).not.toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should auto-detect rslint.config.ts', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'let a: any = 10;\na.b = 20;\n',
      'rslint.config.ts': `export default [${JSON.stringify({
        files: ['**/*.ts'],
        languageOptions: {
          parserOptions: {
            projectService: false,
            project: ['./tsconfig.json'],
          },
        },
        rules: { '@typescript-eslint/no-unsafe-member-access': 'error' },
        plugins: ['@typescript-eslint'],
      })}];`,
    });
    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-unsafe-member-access');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should pass --fix flag through with JS config', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'const a = 1;\n',
      'rslint.config.js': jsConfig({
        rules: { '@typescript-eslint/no-explicit-any': 'off' },
      }),
    });
    try {
      // --fix should not crash and should produce exit code 0 on clean file
      const result = await runRslint(['--fix'], tempDir);
      expect(result.exitCode).toBe(0);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should pass --quiet flag through with JS config', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'test.ts': 'let a: any = 10;\na.b = 20;\n',
      'rslint.config.js': jsConfig(),
    });
    try {
      const result = await runRslint(['--quiet'], tempDir);
      // --quiet suppresses warnings, but errors still show
      expect(result.exitCode).not.toBe(0);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('--init should generate rslint.config.ts when tsconfig.json exists', async () => {
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
    });
    try {
      const result = await runRslint(['--init'], tempDir);
      expect(result.exitCode).toBe(0);
      const files = await fs.readdir(tempDir);
      expect(files).toContain('rslint.config.ts');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('--init should generate rslint.config.js for ESM project (type: module)', async () => {
    const tempDir = await createTempDir({
      'package.json': '{"type": "module"}',
    });
    try {
      const result = await runRslint(['--init'], tempDir);
      expect(result.exitCode).toBe(0);
      const files = await fs.readdir(tempDir);
      expect(files).toContain('rslint.config.js');
      expect(files).not.toContain('rslint.config.mjs');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('--init should generate rslint.config.mjs for non-ESM project', async () => {
    const tempDir = await createTempDir({
      'package.json': '{"name": "my-project"}',
    });
    try {
      const result = await runRslint(['--init'], tempDir);
      expect(result.exitCode).toBe(0);
      const files = await fs.readdir(tempDir);
      expect(files).toContain('rslint.config.mjs');
      expect(files).not.toContain('rslint.config.js');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('--init should generate rslint.config.mjs when no package.json', async () => {
    const tempDir = await createTempDir({});
    try {
      const result = await runRslint(['--init'], tempDir);
      expect(result.exitCode).toBe(0);
      const files = await fs.readdir(tempDir);
      expect(files).toContain('rslint.config.mjs');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

// --- Config error handling ---

describe('CLI JS config error handling', () => {
  test('should show friendly error for syntactically invalid JS config', async () => {
    const tempDir = await createTempDir({
      'rslint.config.js': 'export default [INVALID SYNTAX',
    });
    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).not.toBe(0);
      expect(result.stderr).toContain('failed to load config');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should show friendly error for config that is not an array', async () => {
    const tempDir = await createTempDir({
      'rslint.config.js': 'export default { rules: {} };',
    });
    try {
      const result = await runRslint([], tempDir);
      expect(result.exitCode).not.toBe(0);
      expect(result.stderr).toContain('invalid config');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

// --- Config directory resolution (cwd correctness) ---

describe('CLI config directory resolution', () => {
  test('src/**/*.ts pattern should match when cwd equals config directory', async () => {
    // Config uses "src/**/*.ts" (not "**/*.ts"), requiring correct cwd for matching
    const tempDir = await createTempDir({
      'tsconfig.json': TS_CONFIG,
      'rslint.config.mjs': `export default [{
        files: ['src/**/*.ts'],
        languageOptions: {
          parserOptions: { projectService: false, project: ['./tsconfig.json'] },
        },
        rules: { '@typescript-eslint/no-explicit-any': 'error' },
        plugins: ['@typescript-eslint'],
      }];`,
      'src/index.ts': 'const x: any = 1;\n',
    });
    try {
      const result = await runRslint([], tempDir);
      // Should find and report the no-explicit-any error
      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-explicit-any');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

// --- defineConfig and config presets ---

describe('defineConfig and config presets', () => {
  test('defineConfig should be importable and return input as-is', () => {
    const input = [
      { files: ['**/*.ts'], rules: { 'no-console': 'error' as const } },
    ];
    const result = defineConfig(input);
    expect(result).toBe(input);
  });

  test('config presets should be importable', () => {
    expect(configs).toBeDefined();
    expect(configs.ts).toBeDefined();
    expect(configs.ts.recommended).toBeDefined();
    expect(configs.js).toBeDefined();
    expect(configs.js.recommended).toBeDefined();
    expect(configs.react).toBeDefined();
    expect(configs.react.recommended).toBeDefined();
    expect(configs.import).toBeDefined();
    expect(configs.import.recommended).toBeDefined();
  });

  test('config presets should be valid config entries', () => {
    for (const preset of Object.values(configs)) {
      const rec = (preset as { recommended: unknown }).recommended;
      expect(typeof rec).toBe('object');
      expect(rec).not.toBeNull();
    }
  });

  test('defineConfig with preset should work with normalizeConfig', () => {
    const config = defineConfig([
      configs.ts.recommended,
      { rules: { '@typescript-eslint/no-explicit-any': 'off' } },
    ]);
    const normalized = normalizeConfig(config);
    expect(normalized.length).toBe(2);
    const lastEntry = normalized[normalized.length - 1];
    expect(lastEntry.rules).toEqual({
      '@typescript-eslint/no-explicit-any': 'off',
    });
  });
});

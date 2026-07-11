import { describe, test, expect } from '@rstest/core';
import {
  coalesceCaseAliasedConfigs,
  discoverConfigs,
  filterConfigsByParentIgnores,
  findJSConfig,
  findJSConfigForTarget,
  findJSConfigUp,
  findJSConfigsInDir,
  JS_CONFIG_FILES,
} from '../src/utils/config-discovery.js';
import type { RslintConfigEntry } from '../src/config/define-config.js';
import fs from 'node:fs';
import path from 'node:path';
import os from 'node:os';

function createTempDir(): string {
  return fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-discovery-test-'));
}

function cleanup(dir: string): void {
  fs.rmSync(dir, { recursive: true, force: true });
}

describe('discoverConfigs', () => {
  test('no files/dirs uses cwd', async () => {
    const tmp = createTempDir();
    try {
      fs.writeFileSync(path.join(tmp, 'rslint.config.js'), 'export default []');
      const result = await discoverConfigs([], [], tmp, null);
      expect(result.size).toBe(1);
      const configPath = [...result.keys()][0];
      expect(configPath).toBe(path.join(tmp, 'rslint.config.js'));
    } finally {
      cleanup(tmp);
    }
  });

  test('explicit config overrides discovery and uses cwd as configDirectory', async () => {
    const tmp = createTempDir();
    const cwd = path.join(tmp, 'project');
    const configDir = path.join(tmp, 'wrapper');
    const configFile = path.join(configDir, 'custom.config.js');
    try {
      fs.mkdirSync(cwd);
      fs.mkdirSync(configDir);
      fs.writeFileSync(path.join(cwd, 'rslint.config.js'), 'export default []');
      fs.writeFileSync(configFile, 'export default []');
      const result = await discoverConfigs([], [], cwd, configFile);
      expect([...result]).toEqual([[configFile, cwd]]);
    } finally {
      cleanup(tmp);
    }
  });

  test('deduplicates files in same directory', async () => {
    const tmp = createTempDir();
    const src = path.join(tmp, 'src');
    try {
      fs.mkdirSync(src);
      fs.writeFileSync(path.join(tmp, 'rslint.config.js'), 'export default []');
      const result = await discoverConfigs(
        [path.join(src, 'a.ts'), path.join(src, 'b.ts')],
        [],
        tmp,
        null,
      );
      expect(result.size).toBe(1);
    } finally {
      cleanup(tmp);
    }
  });

  test('different directories find different configs', async () => {
    const tmp = createTempDir();
    const foo = path.join(tmp, 'packages', 'foo');
    const bar = path.join(tmp, 'packages', 'bar');
    try {
      fs.mkdirSync(foo, { recursive: true });
      fs.mkdirSync(bar, { recursive: true });
      fs.writeFileSync(path.join(foo, 'rslint.config.js'), 'export default []');
      fs.writeFileSync(path.join(bar, 'rslint.config.js'), 'export default []');
      const result = await discoverConfigs(
        [path.join(foo, 'a.ts'), path.join(bar, 'b.ts')],
        [],
        tmp,
        null,
      );
      expect(result.size).toBe(2);
    } finally {
      cleanup(tmp);
    }
  });

  test('no config found returns empty map', async () => {
    const tmp = createTempDir();
    const deep = path.join(tmp, 'a', 'b');
    try {
      fs.mkdirSync(deep, { recursive: true });
      const result = await discoverConfigs(
        [path.join(deep, 'file.ts')],
        [],
        tmp,
        null,
      );
      expect(result.size).toBe(0);
    } finally {
      cleanup(tmp);
    }
  });

  test('directory arg uses dir as start point', async () => {
    const tmp = createTempDir();
    const src = path.join(tmp, 'src');
    try {
      fs.mkdirSync(src);
      fs.writeFileSync(path.join(tmp, 'rslint.config.js'), 'export default []');
      const result = await discoverConfigs([], [src], tmp, null);
      expect(result.size).toBe(1);
    } finally {
      cleanup(tmp);
    }
  });

  test('files and dirs finding same config deduplicated', async () => {
    const tmp = createTempDir();
    const src = path.join(tmp, 'src');
    const lib = path.join(tmp, 'lib');
    try {
      fs.mkdirSync(src);
      fs.mkdirSync(lib);
      fs.writeFileSync(path.join(tmp, 'rslint.config.js'), 'export default []');
      const result = await discoverConfigs(
        [path.join(src, 'a.ts')],
        [lib],
        tmp,
        null,
      );
      expect(result.size).toBe(1);
    } finally {
      cleanup(tmp);
    }
  });

  test('no args discovers nested configs in monorepo', async () => {
    const tmp = createTempDir();
    try {
      fs.mkdirSync(path.join(tmp, 'packages', 'foo'), { recursive: true });
      fs.mkdirSync(path.join(tmp, 'packages', 'bar'), { recursive: true });
      fs.writeFileSync(path.join(tmp, 'rslint.config.js'), 'export default []');
      fs.writeFileSync(
        path.join(tmp, 'packages', 'foo', 'rslint.config.js'),
        'export default []',
      );
      fs.writeFileSync(
        path.join(tmp, 'packages', 'bar', 'rslint.config.ts'),
        'export default []',
      );
      const result = await discoverConfigs([], [], tmp, null);
      expect(result.size).toBe(3);
    } finally {
      cleanup(tmp);
    }
  });

  test('dir arg discovers nested configs within scope', async () => {
    const tmp = createTempDir();
    try {
      fs.mkdirSync(path.join(tmp, 'packages', 'foo'), { recursive: true });
      fs.mkdirSync(path.join(tmp, 'packages', 'bar'), { recursive: true });
      fs.writeFileSync(path.join(tmp, 'rslint.config.js'), 'export default []');
      fs.writeFileSync(
        path.join(tmp, 'packages', 'foo', 'rslint.config.js'),
        'export default []',
      );
      fs.writeFileSync(
        path.join(tmp, 'packages', 'bar', 'rslint.config.js'),
        'export default []',
      );
      const result = await discoverConfigs(
        [],
        [path.join(tmp, 'packages')],
        tmp,
        null,
      );
      expect(result.size).toBe(3);
    } finally {
      cleanup(tmp);
    }
  });

  test('file args do not trigger nested config scan', async () => {
    const tmp = createTempDir();
    try {
      fs.mkdirSync(path.join(tmp, 'packages', 'foo'), { recursive: true });
      fs.mkdirSync(path.join(tmp, 'packages', 'bar'), { recursive: true });
      fs.writeFileSync(path.join(tmp, 'rslint.config.js'), 'export default []');
      fs.writeFileSync(
        path.join(tmp, 'packages', 'bar', 'rslint.config.js'),
        'export default []',
      );
      const result = await discoverConfigs(
        [path.join(tmp, 'packages', 'foo', 'a.ts')],
        [],
        tmp,
        null,
      );
      expect(result.size).toBe(1);
    } finally {
      cleanup(tmp);
    }
  });

  test('file symlink uses lexical config before physical config', async () => {
    const tmp = createTempDir();
    const physicalDir = path.join(tmp, 'physical');
    const lexicalDir = path.join(tmp, 'lexical');
    const physicalConfig = path.join(physicalDir, 'rslint.config.js');
    const lexicalConfig = path.join(lexicalDir, 'rslint.config.js');
    const physicalFile = path.join(physicalDir, 'index.ts');
    const lexicalFile = path.join(lexicalDir, 'index.ts');
    try {
      fs.mkdirSync(physicalDir);
      fs.mkdirSync(lexicalDir);
      fs.writeFileSync(physicalConfig, 'export default []');
      fs.writeFileSync(lexicalConfig, 'export default []');
      fs.writeFileSync(physicalFile, 'export {};');
      fs.symlinkSync(physicalFile, lexicalFile, 'file');

      expect(findJSConfigForTarget(lexicalFile, false)).toBe(lexicalConfig);
      const result = await discoverConfigs([lexicalFile], [], tmp, null);
      expect([...result.keys()]).toEqual([lexicalConfig]);
    } finally {
      cleanup(tmp);
    }
  });

  test('file symlink falls back to physical config without a lexical config', async () => {
    const tmp = createTempDir();
    const physicalDir = path.join(tmp, 'physical');
    const lexicalDir = path.join(tmp, 'lexical');
    const physicalConfig = path.join(physicalDir, 'rslint.config.js');
    const physicalFile = path.join(physicalDir, 'index.ts');
    const lexicalFile = path.join(lexicalDir, 'index.ts');
    try {
      fs.mkdirSync(physicalDir);
      fs.mkdirSync(lexicalDir);
      fs.writeFileSync(physicalConfig, 'export default []');
      fs.writeFileSync(physicalFile, 'export {};');
      fs.symlinkSync(physicalFile, lexicalFile, 'file');

      const canonicalConfig = fs.realpathSync(physicalConfig);
      expect(findJSConfigForTarget(lexicalFile, false)).toBe(canonicalConfig);
      const result = await discoverConfigs([lexicalFile], [], tmp, null);
      expect([...result.keys()]).toEqual([canonicalConfig]);
    } finally {
      cleanup(tmp);
    }
  });
});

describe('coalesceCaseAliasedConfigs', () => {
  test('keeps case-distinct config roots when only their config files share a target', () => {
    const tmp = createTempDir();
    const upperDir = path.join(tmp, 'Project');
    const lowerDir = path.join(tmp, 'project');
    try {
      fs.mkdirSync(upperDir);
      try {
        fs.mkdirSync(lowerDir);
      } catch {
        return;
      }
      const sharedConfig = path.join(tmp, 'shared.config.mjs');
      fs.writeFileSync(sharedConfig, 'export default []');
      const upperConfig = path.join(upperDir, 'rslint.config.mjs');
      const lowerConfig = path.join(lowerDir, 'rslint.config.mjs');
      try {
        fs.symlinkSync(sharedConfig, upperConfig, 'file');
        fs.symlinkSync(sharedConfig, lowerConfig, 'file');
      } catch {
        return;
      }

      const result = coalesceCaseAliasedConfigs(
        new Map([
          [upperConfig, upperDir],
          [lowerConfig, lowerDir],
        ]),
      );
      expect(result.configs.size).toBe(2);
    } finally {
      cleanup(tmp);
    }
  });
});

describe('findJSConfig', () => {
  test('skips a higher-priority filename that is a directory', () => {
    const tmp = createTempDir();
    try {
      fs.mkdirSync(path.join(tmp, 'rslint.config.js'));
      const configPath = path.join(tmp, 'rslint.config.mjs');
      fs.writeFileSync(configPath, 'export default []');

      expect(findJSConfig(tmp)).toBe(configPath);
    } finally {
      cleanup(tmp);
    }
  });
});

// --- filterConfigsByParentIgnores ---

/**
 * Build a platform-appropriate absolute path for filter tests.
 * Uses a non-existent root so that realpathSync consistently fails for all
 * paths (avoiding symlink resolution mismatches, e.g. /tmp → /private/tmp
 * on macOS). path.resolve ensures the correct format per platform
 * (forward slashes on Unix, drive-letter + backslashes on Windows).
 */
const ROOT = path.resolve('/rslint-test-nonexistent');
const P = (...segments: string[]) => path.join(ROOT, ...segments);

/** Shorthand config entry builder. */
function cfg(dir: string, ...entries: Record<string, unknown>[]) {
  return { configDirectory: dir, entries: entries as RslintConfigEntry[] };
}

function globalIgnore(...patterns: string[]): Record<string, unknown> {
  return { ignores: patterns };
}

function ruleEntry(
  files: string[],
  rules: Record<string, string>,
): Record<string, unknown> {
  return { files, rules };
}

function dirs(result: { configDirectory: string }[]): string[] {
  return result.map((r) => r.configDirectory);
}

describe('filterConfigsByParentIgnores', () => {
  test('single config is not affected', () => {
    const result = filterConfigsByParentIgnores([
      cfg(P(), globalIgnore('dist/**'), ruleEntry(['**/*.ts'], {})),
    ]);
    expect(result).toHaveLength(1);
  });

  test('empty input returns empty', () => {
    expect(filterConfigsByParentIgnores([])).toHaveLength(0);
  });

  test('filters nested config in globally ignored directory', () => {
    const result = filterConfigsByParentIgnores([
      cfg(P(), globalIgnore('__tests__/**'), ruleEntry(['**/*.ts'], {})),
      cfg(
        P('__tests__', 'fixtures'),
        ruleEntry(['**/*.ts'], { 'no-console': 'error' }),
      ),
    ]);
    expect(result).toHaveLength(1);
    expect(dirs(result)).toEqual([P()]);
  });

  test.each([
    'packages//ignored/**',
    'packages/./ignored/**',
    './packages/ignored/**',
  ])(
    'normalizes parent ignore pattern %s before hierarchy matching',
    (pattern) => {
      const result = filterConfigsByParentIgnores([
        cfg(P(), globalIgnore(pattern), ruleEntry(['**/*.ts'], {})),
        cfg(P('packages', 'ignored'), ruleEntry(['**/*.ts'], {})),
      ]);

      expect(dirs(result)).toEqual([P()]);
    },
  );

  test('builds the effective transaction catalog from loaded candidates', () => {
    const parent = {
      ...cfg(P(), globalIgnore('generated/**')),
      configPath: P('rslint.config.js'),
    };
    const ignoredNested = {
      ...cfg(P('generated', 'package'), ruleEntry(['**/*.ts'], {})),
      configPath: P('generated', 'package', 'rslint.config.js'),
    };
    const sibling = {
      ...cfg(P('packages', 'app'), ruleEntry(['**/*.ts'], {})),
      configPath: P('packages', 'app', 'rslint.config.js'),
    };

    const result = filterConfigsByParentIgnores([
      ignoredNested,
      sibling,
      parent,
    ]);

    expect(result).toEqual([parent, sibling]);
    expect(result).not.toContain(ignoredNested);
  });

  test('explicit empty config fields make ignores entry-level', () => {
    for (const emptyField of [
      { rules: {} },
      { plugins: [] },
      { plugins: {} },
      { settings: {} },
      { languageOptions: {} },
    ]) {
      const result = filterConfigsByParentIgnores([
        cfg(
          P(),
          { ignores: ['__tests__/**'], ...emptyField },
          ruleEntry(['**/*.ts'], {}),
        ),
        cfg(
          P('__tests__', 'fixtures'),
          ruleEntry(['**/*.ts'], { 'no-console': 'error' }),
        ),
      ]);
      expect(result).toHaveLength(2);
      expect(dirs(result)).toContain(P('__tests__', 'fixtures'));
    }
  });

  test('filters with **/prefix/** pattern', () => {
    const result = filterConfigsByParentIgnores([
      cfg(P(), globalIgnore('**/fixtures/**'), ruleEntry(['**/*.ts'], {})),
      cfg(
        P('packages', 'ext', '__tests__', 'fixtures'),
        ruleEntry(['**/*.ts'], {}),
      ),
    ]);
    expect(result).toHaveLength(1);
  });

  test('filters with **/dist/** pattern', () => {
    const result = filterConfigsByParentIgnores([
      cfg(P(), globalIgnore('**/dist/**'), ruleEntry(['**/*.ts'], {})),
      cfg(
        P('packages', 'app', 'dist', 'generated'),
        ruleEntry(['**/*.ts'], {}),
      ),
    ]);
    expect(result).toHaveLength(1);
  });

  test('filters with trailing slash', () => {
    const result = filterConfigsByParentIgnores([
      cfg(P() + '/', globalIgnore('__tests__/')),
      cfg(P('__tests__', 'fixtures'), ruleEntry(['**/*.ts'], {})),
    ]);
    expect(result).toHaveLength(1);
  });

  test('filters with multiple patterns in one entry', () => {
    const result = filterConfigsByParentIgnores([
      cfg(
        P(),
        globalIgnore('__tests__/**', 'e2e/**', 'examples/**'),
        ruleEntry(['**/*.ts'], {}),
      ),
      cfg(P('__tests__', 'fixtures'), ruleEntry(['**/*.ts'], {})),
      cfg(P('e2e', 'setup'), ruleEntry(['**/*.ts'], {})),
      cfg(P('packages', 'app'), ruleEntry(['**/*.ts'], {})),
    ]);
    expect(result).toHaveLength(2);
    expect(dirs(result)).toContain(P());
    expect(dirs(result)).toContain(P('packages', 'app'));
  });

  test('filters with multiple global ignore entries', () => {
    const result = filterConfigsByParentIgnores([
      cfg(
        P(),
        globalIgnore('__tests__/**'),
        globalIgnore('e2e/**'),
        ruleEntry(['**/*.ts'], {}),
      ),
      cfg(P('__tests__', 'fixtures'), ruleEntry(['**/*.ts'], {})),
      cfg(P('e2e', 'setup'), ruleEntry(['**/*.ts'], {})),
    ]);
    expect(result).toHaveLength(1);
  });

  test('entry-level ignores do not filter nested configs', () => {
    const result = filterConfigsByParentIgnores([
      cfg(P(), {
        files: ['**/*.ts'],
        ignores: ['__tests__/**'],
        rules: { r: 'error' },
      }),
      cfg(P('__tests__', 'fixtures'), ruleEntry(['**/*.ts'], {})),
    ]);
    expect(result).toHaveLength(2);
  });

  test('sibling configs do not affect each other', () => {
    const result = filterConfigsByParentIgnores([
      cfg(P('packages', 'app'), ruleEntry(['**/*.ts'], { a: 'error' })),
      cfg(P('packages', 'lib'), ruleEntry(['**/*.ts'], { b: 'error' })),
    ]);
    expect(result).toHaveLength(2);
  });

  test('parent without global ignores does not filter children', () => {
    const result = filterConfigsByParentIgnores([
      cfg(P(), ruleEntry(['**/*.ts'], { a: 'error' })),
      cfg(P('packages', 'app'), ruleEntry(['**/*.ts'], { b: 'error' })),
    ]);
    expect(result).toHaveLength(2);
  });

  test('grandparent global ignores filter grandchild', () => {
    const result = filterConfigsByParentIgnores([
      cfg(P(), globalIgnore('vendor/**'), ruleEntry(['**/*.ts'], {})),
      cfg(P('packages', 'app'), ruleEntry(['**/*.ts'], {})),
      cfg(P('vendor', 'lib'), ruleEntry(['**/*.ts'], {})),
    ]);
    expect(result).toHaveLength(2);
    expect(dirs(result)).not.toContain(P('vendor', 'lib'));
  });

  test('intermediate config ignores only affect its children', () => {
    const result = filterConfigsByParentIgnores([
      cfg(P(), ruleEntry(['**/*.ts'], {})),
      cfg(
        P('packages', 'app'),
        globalIgnore('generated/**'),
        ruleEntry(['**/*.ts'], {}),
      ),
      cfg(
        P('packages', 'app', 'generated', 'config'),
        ruleEntry(['**/*.ts'], {}),
      ),
      cfg(P('packages', 'lib'), ruleEntry(['**/*.ts'], {})),
    ]);
    expect(result).toHaveLength(3);
    expect(dirs(result)).not.toContain(
      P('packages', 'app', 'generated', 'config'),
    );
    expect(dirs(result)).toContain(P('packages', 'lib'));
  });

  test('nearest config boundary replaces ancestor global ignores', () => {
    const result = filterConfigsByParentIgnores([
      cfg(
        P(),
        globalIgnore('packages/app/generated/**'),
        ruleEntry(['**/*.ts'], {}),
      ),
      cfg(P('packages', 'app'), ruleEntry(['**/*.ts'], {})),
      cfg(P('packages', 'app', 'generated'), ruleEntry(['**/*.ts'], {})),
    ]);
    expect(dirs(result)).toContain(P('packages', 'app', 'generated'));
  });

  test('global ignore entry with name field is still recognized', () => {
    const result = filterConfigsByParentIgnores([
      cfg(
        P(),
        { name: 'global-ignores', ignores: ['vendor/**'] },
        ruleEntry(['**/*.ts'], {}),
      ),
      cfg(P('vendor', 'lib'), ruleEntry(['**/*.ts'], {})),
    ]);
    expect(result).toHaveLength(1);
  });

  test('real-world monorepo pattern', () => {
    const result = filterConfigsByParentIgnores([
      cfg(
        P(),
        globalIgnore('**/fixtures/**', '**/dist/**', 'e2e/**'),
        ruleEntry(['**/*.ts'], {}),
      ),
      cfg(P('packages', 'app'), ruleEntry(['**/*.ts'], { a: 'error' })),
      cfg(P('packages', 'lib'), ruleEntry(['**/*.ts'], { b: 'error' })),
      cfg(
        P('packages', 'vscode-ext', '__tests__', 'fixtures'),
        ruleEntry(['**/*.ts'], {}),
      ),
      cfg(P('e2e', 'helpers'), ruleEntry(['**/*.ts'], {})),
      cfg(P('packages', 'app', 'dist', 'gen'), ruleEntry(['**/*.ts'], {})),
    ]);
    expect(result).toHaveLength(4);
    expect(dirs(result)).toContain(P());
    expect(dirs(result)).toContain(P('packages', 'app'));
    expect(dirs(result)).toContain(P('packages', 'lib'));
    expect(dirs(result)).toContain(P('packages', 'app', 'dist', 'gen'));
    expect(dirs(result)).not.toContain(
      P('packages', 'vscode-ext', '__tests__', 'fixtures'),
    );
    expect(dirs(result)).not.toContain(P('e2e', 'helpers'));
  });

  // --- Glob pattern coverage (picomatch) ---

  test('wildcard in middle: packages/*/dist/**', () => {
    const result = filterConfigsByParentIgnores([
      cfg(P(), globalIgnore('packages/*/dist/**'), ruleEntry(['**/*.ts'], {})),
      cfg(P('packages', 'app', 'dist', 'gen'), ruleEntry(['**/*.ts'], {})),
      cfg(P('packages', 'app', 'src'), ruleEntry(['**/*.ts'], {})),
    ]);
    expect(result).toHaveLength(2);
    expect(dirs(result)).not.toContain(P('packages', 'app', 'dist', 'gen'));
    expect(dirs(result)).toContain(P('packages', 'app', 'src'));
  });

  test('brace expansion: {__tests__,e2e}/**', () => {
    const result = filterConfigsByParentIgnores([
      cfg(P(), globalIgnore('{__tests__,e2e}/**'), ruleEntry(['**/*.ts'], {})),
      cfg(P('__tests__', 'fixtures'), ruleEntry(['**/*.ts'], {})),
      cfg(P('e2e', 'setup'), ruleEntry(['**/*.ts'], {})),
      cfg(P('packages', 'app'), ruleEntry(['**/*.ts'], {})),
    ]);
    expect(result).toHaveLength(2);
    expect(dirs(result)).toContain(P());
    expect(dirs(result)).toContain(P('packages', 'app'));
  });

  test('bare directory name without glob: dist', () => {
    const result = filterConfigsByParentIgnores([
      cfg(P(), globalIgnore('dist'), ruleEntry(['**/*.ts'], {})),
      cfg(P('dist'), ruleEntry(['**/*.ts'], {})),
      cfg(P('dist', 'sub'), ruleEntry(['**/*.ts'], {})),
    ]);
    // 'dist' matches the dir exactly, and dist/sub is nested under it
    expect(dirs(result)).not.toContain(P('dist'));
    expect(dirs(result)).not.toContain(P('dist', 'sub'));
  });

  test('dot-prefixed directories: .cache/**', () => {
    const result = filterConfigsByParentIgnores([
      cfg(P(), globalIgnore('.cache/**'), ruleEntry(['**/*.ts'], {})),
      cfg(P('.cache', 'generated'), ruleEntry(['**/*.ts'], {})),
    ]);
    expect(result).toHaveLength(1);
    expect(dirs(result)).toEqual([P()]);
  });

  test('directory names starting with .. are still child directories', () => {
    const result = filterConfigsByParentIgnores([
      cfg(P(), globalIgnore('..foo/**'), ruleEntry(['**/*.ts'], {})),
      cfg(P('..foo'), ruleEntry(['**/*.ts'], {})),
    ]);
    expect(result).toHaveLength(1);
    expect(result[0].configDirectory).toBe(P());
  });

  test('negation cannot re-include a descendant of an ignored directory', () => {
    // `vendor/**` ignores the parent directory before the descendant negation
    // can be reached, matching ESLint v10 directory traversal.
    const result = filterConfigsByParentIgnores([
      cfg(
        P(),
        globalIgnore('vendor/**', '!vendor/keep/**'),
        ruleEntry(['**/*.ts'], {}),
      ),
      cfg(P('vendor', 'lib'), ruleEntry(['**/*.ts'], {})),
      cfg(P('vendor', 'keep'), ruleEntry(['**/*.ts'], {})),
      // Unrelated directories are not affected by the negation pattern.
      cfg(P('packages', 'app'), ruleEntry(['**/*.ts'], {})),
    ]);
    expect(dirs(result)).not.toContain(P('vendor', 'lib'));
    expect(dirs(result)).not.toContain(P('vendor', 'keep'));
    expect(dirs(result)).toContain(P('packages', 'app'));
    expect(dirs(result)).toContain(P());
  });

  test('only-negation global ignore does not filter anything', () => {
    // If global ignore only has negation patterns, nothing should be filtered
    const result = filterConfigsByParentIgnores([
      cfg(P(), globalIgnore('!vendor/**'), ruleEntry(['**/*.ts'], {})),
      cfg(P('vendor', 'lib'), ruleEntry(['**/*.ts'], {})),
      cfg(P('packages', 'app'), ruleEntry(['**/*.ts'], {})),
    ]);
    expect(result).toHaveLength(3);
  });

  test('dir/** and dir/**/* both block ignored descendant directories', () => {
    const withDirPattern = filterConfigsByParentIgnores([
      cfg(P(), globalIgnore('vendor/**'), ruleEntry(['**/*.ts'], {})),
      cfg(P('vendor', 'keep'), ruleEntry(['**/*.ts'], {})),
    ]);
    expect(withDirPattern).toHaveLength(1);

    const withFilePattern = filterConfigsByParentIgnores([
      cfg(P(), globalIgnore('vendor/**/*'), ruleEntry(['**/*.ts'], {})),
      cfg(P('vendor', 'keep'), ruleEntry(['**/*.ts'], {})),
    ]);
    expect(withFilePattern).toHaveLength(1);
  });

  test('dir/* blocks a directly matched child directory', () => {
    const result = filterConfigsByParentIgnores([
      cfg(P(), globalIgnore('vendor/*'), ruleEntry(['**/*.ts'], {})),
      cfg(P('vendor', 'keep'), ruleEntry(['**/*.ts'], {})),
    ]);
    expect(result).toHaveLength(1);
  });

  test('directory negations keep nested config traversal reachable', () => {
    const result = filterConfigsByParentIgnores([
      cfg(
        P(),
        globalIgnore('vendor/**/*', '!vendor/**/*/'),
        ruleEntry(['**/*.ts'], {}),
      ),
      cfg(P('vendor', 'keep'), ruleEntry(['**/*.ts'], {})),
      cfg(P('vendor', 'drop'), ruleEntry(['**/*.ts'], {})),
    ]);
    expect(dirs(result)).toContain(P('vendor', 'keep'));
    expect(dirs(result)).toContain(P('vendor', 'drop'));
  });

  test('top-level negation re-includes a directory', () => {
    const result = filterConfigsByParentIgnores([
      cfg(P(), globalIgnore('*', '!vendor'), ruleEntry(['**/*.ts'], {})),
      cfg(P('vendor', 'keep'), ruleEntry(['**/*.ts'], {})),
      cfg(P('packages', 'app'), ruleEntry(['**/*.ts'], {})),
    ]);
    expect(dirs(result)).toContain(P('vendor', 'keep'));
    expect(dirs(result)).not.toContain(P('packages', 'app'));
  });

  // --- Mixed global + entry-level ignores in same config ---

  test('config with both global and entry-level ignores: only global filters nested configs', () => {
    const result = filterConfigsByParentIgnores([
      cfg(
        P(),
        // Global ignore → should filter nested configs in dist/
        globalIgnore('dist/**'),
        // Entry-level ignore → should NOT filter nested configs in test/
        { files: ['**/*.ts'], ignores: ['test/**'], rules: { r: 'error' } },
      ),
      // In dist/ → should be filtered (global ignore)
      cfg(P('dist', 'generated'), ruleEntry(['**/*.ts'], {})),
      // In test/ → should NOT be filtered (entry-level ignore)
      cfg(P('test', 'fixtures'), ruleEntry(['**/*.ts'], {})),
      // Normal child → should NOT be filtered
      cfg(P('packages', 'app'), ruleEntry(['**/*.ts'], {})),
    ]);
    expect(result).toHaveLength(3);
    expect(dirs(result)).not.toContain(P('dist', 'generated'));
    expect(dirs(result)).toContain(P('test', 'fixtures'));
    expect(dirs(result)).toContain(P('packages', 'app'));
  });

  // --- Path edge cases ---

  test('trailing slash in configDirectory is handled', () => {
    const result = filterConfigsByParentIgnores([
      cfg(P() + '/', globalIgnore('__tests__/**'), ruleEntry(['**/*.ts'], {})),
      cfg(P('__tests__', 'fixtures'), ruleEntry(['**/*.ts'], {})),
    ]);
    expect(result).toHaveLength(1);
  });

  test('trailing slash on both parent and child', () => {
    const result = filterConfigsByParentIgnores([
      cfg(P() + '/', globalIgnore('vendor/**'), ruleEntry(['**/*.ts'], {})),
      cfg(P('vendor', 'lib') + '/', ruleEntry(['**/*.ts'], {})),
    ]);
    expect(result).toHaveLength(1);
  });

  test('trailing slash normalization preserves filesystem root', () => {
    const root = path.parse(path.resolve('/')).root;
    const child = path.join(root, 'rslint-root-child');
    const result = filterConfigsByParentIgnores([
      cfg(
        root,
        globalIgnore('rslint-root-child/**'),
        ruleEntry(['**/*.ts'], {}),
      ),
      cfg(child, ruleEntry(['**/*.ts'], {})),
    ]);

    expect(result).toHaveLength(1);
    expect(dirs(result)).toContain(root);
    expect(dirs(result)).not.toContain(child);
  });

  // --- Cross-package ignore isolation ---

  test('child global ignore does NOT bubble up to filter parent', () => {
    const result = filterConfigsByParentIgnores([
      cfg(P(), ruleEntry(['**/*.ts'], {})),
      cfg(
        P('packages', 'app'),
        globalIgnore('generated/**'),
        ruleEntry(['**/*.ts'], {}),
      ),
    ]);
    // Parent should never be filtered by child's ignores
    expect(result).toHaveLength(2);
    expect(dirs(result)).toContain(P());
    expect(dirs(result)).toContain(P('packages', 'app'));
  });

  test('sibling A global ignore does NOT filter sibling B nested config', () => {
    // app ignores generated/, lib also has a generated/ subdir with a config.
    // lib's generated/ should NOT be filtered by app's ignore.
    const result = filterConfigsByParentIgnores([
      cfg(P(), ruleEntry(['**/*.ts'], {})),
      cfg(
        P('packages', 'app'),
        globalIgnore('generated/**'),
        ruleEntry(['**/*.ts'], {}),
      ),
      cfg(
        P('packages', 'app', 'generated', 'output'),
        ruleEntry(['**/*.ts'], {}),
      ),
      cfg(P('packages', 'lib'), ruleEntry(['**/*.ts'], {})),
      cfg(
        P('packages', 'lib', 'generated', 'output'),
        ruleEntry(['**/*.ts'], {}),
      ),
    ]);
    // app/generated should be filtered (by app's ignore)
    expect(dirs(result)).not.toContain(
      P('packages', 'app', 'generated', 'output'),
    );
    // lib/generated should NOT be filtered (app's ignore doesn't affect lib)
    expect(dirs(result)).toContain(P('packages', 'lib', 'generated', 'output'));
    expect(result).toHaveLength(4);
  });

  test('both siblings have same-name ignored dir, only own children filtered', () => {
    const result = filterConfigsByParentIgnores([
      cfg(P(), ruleEntry(['**/*.ts'], {})),
      cfg(
        P('packages', 'app'),
        globalIgnore('dist/**'),
        ruleEntry(['**/*.ts'], {}),
      ),
      cfg(P('packages', 'app', 'dist', 'gen'), ruleEntry(['**/*.ts'], {})),
      cfg(
        P('packages', 'lib'),
        globalIgnore('dist/**'),
        ruleEntry(['**/*.ts'], {}),
      ),
      cfg(P('packages', 'lib', 'dist', 'gen'), ruleEntry(['**/*.ts'], {})),
    ]);
    // Both dist/gen configs filtered by their respective parent
    expect(dirs(result)).not.toContain(P('packages', 'app', 'dist', 'gen'));
    expect(dirs(result)).not.toContain(P('packages', 'lib', 'dist', 'gen'));
    // Root + both packages survive
    expect(result).toHaveLength(3);
    expect(dirs(result)).toContain(P());
    expect(dirs(result)).toContain(P('packages', 'app'));
    expect(dirs(result)).toContain(P('packages', 'lib'));
  });

  test('filters lexical child configs in ignored directories', () => {
    const tmp = createTempDir();
    const nestedDir = path.join(tmp, 'sub', 'nested');
    fs.mkdirSync(nestedDir, { recursive: true });
    try {
      const result = filterConfigsByParentIgnores([
        cfg(tmp, globalIgnore('sub/**'), ruleEntry(['**/*.ts'], {})),
        cfg(nestedDir, ruleEntry(['**/*.ts'], {})),
      ]);
      expect(result).toHaveLength(1);
      expect(result[0].configDirectory).toBe(tmp);
    } finally {
      cleanup(tmp);
    }
  });

  test('does not infer config ancestry through a directory symlink', () => {
    const tmp = createTempDir();
    const physicalRoot = path.join(tmp, 'physical');
    const physicalChild = path.join(physicalRoot, 'child');
    const aliasRoot = path.join(tmp, 'alias');
    fs.mkdirSync(physicalChild, { recursive: true });
    try {
      fs.symlinkSync(
        physicalRoot,
        aliasRoot,
        process.platform === 'win32' ? 'junction' : 'dir',
      );
      const result = filterConfigsByParentIgnores([
        cfg(aliasRoot, globalIgnore('child/**')),
        cfg(physicalChild, ruleEntry(['**/*.ts'], {})),
      ]);

      expect(dirs(result)).toContain(aliasRoot);
      expect(dirs(result)).toContain(physicalChild);
    } finally {
      cleanup(tmp);
    }
  });
});

describe('findJSConfig', () => {
  test('keeps the full config filename priority order', () => {
    expect(JS_CONFIG_FILES).toEqual([
      'rslint.config.js',
      'rslint.config.mjs',
      'rslint.config.ts',
      'rslint.config.mts',
    ]);
  });

  test('finds rslint.config.js in cwd', () => {
    const tmp = createTempDir();
    try {
      fs.writeFileSync(path.join(tmp, 'rslint.config.js'), 'export default []');
      const result = findJSConfig(tmp);
      expect(result).toBe(path.join(tmp, 'rslint.config.js'));
    } finally {
      cleanup(tmp);
    }
  });

  test('returns null when no config exists', () => {
    const tmp = createTempDir();
    try {
      const result = findJSConfig(tmp);
      expect(result).toBe(null);
    } finally {
      cleanup(tmp);
    }
  });

  test('uses the automatic config filename priority', () => {
    const tmp = createTempDir();
    try {
      // Create every automatically discoverable config variant.
      for (const name of JS_CONFIG_FILES) {
        fs.writeFileSync(path.join(tmp, name), 'export default []');
      }
      const result = findJSConfig(tmp);
      // Should find rslint.config.js (first in priority order)
      expect(result).toBe(path.join(tmp, 'rslint.config.js'));
    } finally {
      cleanup(tmp);
    }
  });

  test('finds mjs when js does not exist', () => {
    const tmp = createTempDir();
    try {
      fs.writeFileSync(
        path.join(tmp, 'rslint.config.mjs'),
        'export default []',
      );
      const result = findJSConfig(tmp);
      expect(result).toBe(path.join(tmp, 'rslint.config.mjs'));
    } finally {
      cleanup(tmp);
    }
  });

  test('finds .ts config when js and mjs do not exist', () => {
    const tmp = createTempDir();
    try {
      fs.writeFileSync(path.join(tmp, 'rslint.config.ts'), 'export default []');
      const result = findJSConfig(tmp);
      expect(result).toBe(path.join(tmp, 'rslint.config.ts'));
    } finally {
      cleanup(tmp);
    }
  });

  test('finds .mts config when no other config exists', () => {
    const tmp = createTempDir();
    try {
      fs.writeFileSync(
        path.join(tmp, 'rslint.config.mts'),
        'export default []',
      );
      const result = findJSConfig(tmp);
      expect(result).toBe(path.join(tmp, 'rslint.config.mts'));
    } finally {
      cleanup(tmp);
    }
  });

  test.each(['cjs', 'cts'])(
    'does not auto-discover .%s configs',
    (extension) => {
      const tmp = createTempDir();
      try {
        fs.writeFileSync(path.join(tmp, `rslint.config.${extension}`), '');
        expect(findJSConfig(tmp)).toBe(null);
      } finally {
        cleanup(tmp);
      }
    },
  );

  test('returns null for non-existent directory', () => {
    const result = findJSConfig('/tmp/definitely-does-not-exist-99999');
    expect(result).toBe(null);
  });
});

describe('findJSConfigUp', () => {
  test('finds config in current directory', () => {
    const tmp = createTempDir();
    try {
      fs.writeFileSync(path.join(tmp, 'rslint.config.js'), 'export default []');
      const result = findJSConfigUp(tmp);
      expect(result).toBe(path.join(tmp, 'rslint.config.js'));
    } finally {
      cleanup(tmp);
    }
  });

  test('finds config in parent directory', () => {
    const tmp = createTempDir();
    const child = path.join(tmp, 'child');
    try {
      fs.mkdirSync(child);
      fs.writeFileSync(path.join(tmp, 'rslint.config.ts'), 'export default []');
      const result = findJSConfigUp(child);
      expect(result).toBe(path.join(tmp, 'rslint.config.ts'));
    } finally {
      cleanup(tmp);
    }
  });

  test('finds config in grandparent directory', () => {
    const tmp = createTempDir();
    const deep = path.join(tmp, 'a', 'b', 'c');
    try {
      fs.mkdirSync(deep, { recursive: true });
      fs.writeFileSync(
        path.join(tmp, 'rslint.config.mjs'),
        'export default []',
      );
      const result = findJSConfigUp(deep);
      expect(result).toBe(path.join(tmp, 'rslint.config.mjs'));
    } finally {
      cleanup(tmp);
    }
  });

  test('stops at nearest config (does not walk further)', () => {
    const tmp = createTempDir();
    const child = path.join(tmp, 'child');
    try {
      fs.mkdirSync(child);
      // Config in both parent and child
      fs.writeFileSync(
        path.join(tmp, 'rslint.config.js'),
        'export default ["parent"]',
      );
      fs.writeFileSync(
        path.join(child, 'rslint.config.js'),
        'export default ["child"]',
      );
      const result = findJSConfigUp(child);
      // Should find child's config (nearest)
      expect(result).toBe(path.join(child, 'rslint.config.js'));
    } finally {
      cleanup(tmp);
    }
  });

  test('returns null when no config found up to root', () => {
    const tmp = createTempDir();
    const deep = path.join(tmp, 'a', 'b');
    try {
      fs.mkdirSync(deep, { recursive: true });
      // No config files anywhere in tmp
      const result = findJSConfigUp(deep);
      // Should return null (eventually hits filesystem root with no config)
      expect(result).toBe(null);
    } finally {
      cleanup(tmp);
    }
  });

  test('works with resolved absolute paths', () => {
    const tmp = createTempDir();
    const child = path.join(tmp, 'src');
    try {
      fs.mkdirSync(child);
      fs.writeFileSync(path.join(tmp, 'rslint.config.js'), 'export default []');
      // Pass a relative-looking path (resolve should handle it)
      const result = findJSConfigUp(child);
      expect(result).toBe(path.join(tmp, 'rslint.config.js'));
    } finally {
      cleanup(tmp);
    }
  });

  test('handles non-existent startDir gracefully', () => {
    // Non-existent directory — path.resolve works, findJSConfig returns null
    // for each level, eventually terminates at root
    const result = findJSConfigUp(
      '/tmp/definitely-does-not-exist-12345/deep/path',
    );
    expect(result).toBe(null);
  });

  test('terminates at root without infinite loop', () => {
    // Starting from root directory should terminate immediately
    const result = findJSConfigUp('/');
    expect(result).toBe(null);
  });

  test('child config priority over parent with different extensions', () => {
    const tmp = createTempDir();
    const child = path.join(tmp, 'child');
    try {
      fs.mkdirSync(child);
      // Parent has .js, child has .mts — child should win (nearest)
      fs.writeFileSync(
        path.join(tmp, 'rslint.config.js'),
        'export default ["parent"]',
      );
      fs.writeFileSync(
        path.join(child, 'rslint.config.mts'),
        'export default ["child"]',
      );
      const result = findJSConfigUp(child);
      expect(result).toBe(path.join(child, 'rslint.config.mts'));
    } finally {
      cleanup(tmp);
    }
  });

  test('finds .ts config via upward traversal', () => {
    const tmp = createTempDir();
    const child = path.join(tmp, 'packages', 'foo');
    try {
      fs.mkdirSync(child, { recursive: true });
      fs.writeFileSync(path.join(tmp, 'rslint.config.ts'), 'export default []');
      const result = findJSConfigUp(child);
      expect(result).toBe(path.join(tmp, 'rslint.config.ts'));
    } finally {
      cleanup(tmp);
    }
  });
});

describe('findJSConfigsInDir', () => {
  test('finds config in root directory', async () => {
    const tmp = createTempDir();
    try {
      fs.writeFileSync(path.join(tmp, 'rslint.config.js'), 'export default []');
      const result = await findJSConfigsInDir(tmp);
      expect(result).toEqual([path.join(tmp, 'rslint.config.js')]);
    } finally {
      cleanup(tmp);
    }
  });

  test('finds configs in nested directories', async () => {
    const tmp = createTempDir();
    try {
      fs.mkdirSync(path.join(tmp, 'packages', 'foo'), { recursive: true });
      fs.mkdirSync(path.join(tmp, 'packages', 'bar'), { recursive: true });
      fs.writeFileSync(path.join(tmp, 'rslint.config.js'), 'export default []');
      fs.writeFileSync(
        path.join(tmp, 'packages', 'foo', 'rslint.config.ts'),
        'export default []',
      );
      fs.writeFileSync(
        path.join(tmp, 'packages', 'bar', 'rslint.config.mjs'),
        'export default []',
      );
      const result = (await findJSConfigsInDir(tmp)).sort();
      expect(result).toEqual(
        [
          path.join(tmp, 'rslint.config.js'),
          path.join(tmp, 'packages', 'bar', 'rslint.config.mjs'),
          path.join(tmp, 'packages', 'foo', 'rslint.config.ts'),
        ].sort(),
      );
    } finally {
      cleanup(tmp);
    }
  });

  test('skips node_modules', async () => {
    const tmp = createTempDir();
    try {
      fs.mkdirSync(path.join(tmp, 'node_modules', 'pkg'), { recursive: true });
      fs.writeFileSync(
        path.join(tmp, 'node_modules', 'pkg', 'rslint.config.js'),
        'export default []',
      );
      fs.writeFileSync(path.join(tmp, 'rslint.config.js'), 'export default []');
      const result = await findJSConfigsInDir(tmp);
      expect(result).toEqual([path.join(tmp, 'rslint.config.js')]);
    } finally {
      cleanup(tmp);
    }
  });

  test('skips .git directory', async () => {
    const tmp = createTempDir();
    try {
      fs.mkdirSync(path.join(tmp, '.git', 'hooks'), { recursive: true });
      fs.writeFileSync(
        path.join(tmp, '.git', 'rslint.config.js'),
        'export default []',
      );
      fs.writeFileSync(path.join(tmp, 'rslint.config.js'), 'export default []');
      const result = await findJSConfigsInDir(tmp);
      expect(result).toEqual([path.join(tmp, 'rslint.config.js')]);
    } finally {
      cleanup(tmp);
    }
  });

  test('returns empty array when no configs found', async () => {
    const tmp = createTempDir();
    try {
      fs.mkdirSync(path.join(tmp, 'src'), { recursive: true });
      fs.writeFileSync(path.join(tmp, 'src', 'index.ts'), 'const x = 1;');
      const result = await findJSConfigsInDir(tmp);
      expect(result).toEqual([]);
    } finally {
      cleanup(tmp);
    }
  });

  test('handles non-existent directory gracefully', async () => {
    const result = await findJSConfigsInDir('/tmp/does-not-exist-99999');
    expect(result).toEqual([]);
  });

  test('does not traverse into nested node_modules', async () => {
    const tmp = createTempDir();
    try {
      fs.mkdirSync(path.join(tmp, 'packages', 'foo', 'node_modules', 'dep'), {
        recursive: true,
      });
      fs.writeFileSync(
        path.join(
          tmp,
          'packages',
          'foo',
          'node_modules',
          'dep',
          'rslint.config.js',
        ),
        'export default []',
      );
      fs.writeFileSync(
        path.join(tmp, 'packages', 'foo', 'rslint.config.js'),
        'export default []',
      );
      const result = await findJSConfigsInDir(tmp);
      expect(result).toEqual([
        path.join(tmp, 'packages', 'foo', 'rslint.config.js'),
      ]);
    } finally {
      cleanup(tmp);
    }
  });

  test('uses config file priority within the same directory', async () => {
    const tmp = createTempDir();
    try {
      for (const name of JS_CONFIG_FILES) {
        fs.writeFileSync(path.join(tmp, name), 'export default []');
      }
      const result = await findJSConfigsInDir(tmp);
      expect(result).toEqual([path.join(tmp, 'rslint.config.js')]);
    } finally {
      cleanup(tmp);
    }
  });

  test('does not follow symlinked directories', async () => {
    const tmp = createTempDir();
    const realDir = path.join(tmp, 'real');
    const linkDir = path.join(tmp, 'link');
    try {
      fs.mkdirSync(realDir);
      fs.writeFileSync(
        path.join(realDir, 'rslint.config.js'),
        'export default []',
      );
      fs.symlinkSync(realDir, linkDir, 'dir');

      const result = await findJSConfigsInDir(tmp);
      expect(result).toEqual([path.join(realDir, 'rslint.config.js')]);
    } finally {
      cleanup(tmp);
    }
  });
});

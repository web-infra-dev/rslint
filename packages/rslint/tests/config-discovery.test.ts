import { describe, test, expect } from '@rstest/core';
import {
  discoverConfigs,
  filterConfigsByParentIgnores,
} from '../src/utils/config-discovery.js';
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
  test('no files/dirs uses cwd', () => {
    const tmp = createTempDir();
    try {
      fs.writeFileSync(path.join(tmp, 'rslint.config.js'), 'export default []');
      const result = discoverConfigs([], [], tmp, null);
      expect(result.size).toBe(1);
      const configPath = [...result.keys()][0];
      expect(configPath).toBe(path.join(tmp, 'rslint.config.js'));
    } finally {
      cleanup(tmp);
    }
  });

  test('explicit config overrides discovery', () => {
    const tmp = createTempDir();
    const configFile = path.join(tmp, 'custom.config.js');
    try {
      fs.writeFileSync(configFile, 'export default []');
      const result = discoverConfigs(
        ['/some/file.ts'],
        ['/some/dir'],
        tmp,
        configFile,
      );
      expect(result.size).toBe(1);
      expect([...result.keys()][0]).toBe(configFile);
    } finally {
      cleanup(tmp);
    }
  });

  test('deduplicates files in same directory', () => {
    const tmp = createTempDir();
    const src = path.join(tmp, 'src');
    try {
      fs.mkdirSync(src);
      fs.writeFileSync(path.join(tmp, 'rslint.config.js'), 'export default []');
      const result = discoverConfigs(
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

  test('different directories find different configs', () => {
    const tmp = createTempDir();
    const foo = path.join(tmp, 'packages', 'foo');
    const bar = path.join(tmp, 'packages', 'bar');
    try {
      fs.mkdirSync(foo, { recursive: true });
      fs.mkdirSync(bar, { recursive: true });
      fs.writeFileSync(path.join(foo, 'rslint.config.js'), 'export default []');
      fs.writeFileSync(path.join(bar, 'rslint.config.js'), 'export default []');
      const result = discoverConfigs(
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

  test('no config found returns empty map', () => {
    const tmp = createTempDir();
    const deep = path.join(tmp, 'a', 'b');
    try {
      fs.mkdirSync(deep, { recursive: true });
      const result = discoverConfigs(
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

  test('directory arg uses dir as start point', () => {
    const tmp = createTempDir();
    const src = path.join(tmp, 'src');
    try {
      fs.mkdirSync(src);
      fs.writeFileSync(path.join(tmp, 'rslint.config.js'), 'export default []');
      const result = discoverConfigs([], [src], tmp, null);
      expect(result.size).toBe(1);
    } finally {
      cleanup(tmp);
    }
  });

  test('files and dirs finding same config deduplicated', () => {
    const tmp = createTempDir();
    const src = path.join(tmp, 'src');
    const lib = path.join(tmp, 'lib');
    try {
      fs.mkdirSync(src);
      fs.mkdirSync(lib);
      fs.writeFileSync(path.join(tmp, 'rslint.config.js'), 'export default []');
      const result = discoverConfigs(
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

  test('no args discovers nested configs in monorepo', () => {
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
      const result = discoverConfigs([], [], tmp, null);
      expect(result.size).toBe(3);
    } finally {
      cleanup(tmp);
    }
  });

  test('dir arg discovers nested configs within scope', () => {
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
      const result = discoverConfigs(
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

  test('file args do not trigger nested config scan', () => {
    const tmp = createTempDir();
    try {
      fs.mkdirSync(path.join(tmp, 'packages', 'foo'), { recursive: true });
      fs.mkdirSync(path.join(tmp, 'packages', 'bar'), { recursive: true });
      fs.writeFileSync(path.join(tmp, 'rslint.config.js'), 'export default []');
      fs.writeFileSync(
        path.join(tmp, 'packages', 'bar', 'rslint.config.js'),
        'export default []',
      );
      const result = discoverConfigs(
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
});

// --- filterConfigsByParentIgnores ---

/** Shorthand config entry builder. */
function cfg(dir: string, ...entries: Record<string, unknown>[]) {
  return { configDirectory: dir, entries: entries as unknown[] };
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
  return result.map(r => r.configDirectory);
}

describe('filterConfigsByParentIgnores', () => {
  test('single config is not affected', () => {
    const result = filterConfigsByParentIgnores([
      cfg('/project', globalIgnore('dist/**'), ruleEntry(['**/*.ts'], {})),
    ]);
    expect(result).toHaveLength(1);
  });

  test('empty input returns empty', () => {
    expect(filterConfigsByParentIgnores([])).toHaveLength(0);
  });

  test('filters nested config in globally ignored directory', () => {
    const result = filterConfigsByParentIgnores([
      cfg('/project', globalIgnore('__tests__/**'), ruleEntry(['**/*.ts'], {})),
      cfg(
        '/project/__tests__/fixtures',
        ruleEntry(['**/*.ts'], { 'no-console': 'error' }),
      ),
    ]);
    expect(result).toHaveLength(1);
    expect(dirs(result)).toEqual(['/project']);
  });

  test('filters with **/prefix/** pattern', () => {
    const result = filterConfigsByParentIgnores([
      cfg(
        '/project',
        globalIgnore('**/fixtures/**'),
        ruleEntry(['**/*.ts'], {}),
      ),
      cfg(
        '/project/packages/ext/__tests__/fixtures',
        ruleEntry(['**/*.ts'], {}),
      ),
    ]);
    expect(result).toHaveLength(1);
  });

  test('filters with **/dist/** pattern', () => {
    const result = filterConfigsByParentIgnores([
      cfg('/project', globalIgnore('**/dist/**'), ruleEntry(['**/*.ts'], {})),
      cfg('/project/packages/app/dist/generated', ruleEntry(['**/*.ts'], {})),
    ]);
    expect(result).toHaveLength(1);
  });

  test('filters with trailing slash', () => {
    const result = filterConfigsByParentIgnores([
      cfg('/project', globalIgnore('__tests__/')),
      cfg('/project/__tests__/fixtures', ruleEntry(['**/*.ts'], {})),
    ]);
    expect(result).toHaveLength(1);
  });

  test('filters with multiple patterns in one entry', () => {
    const result = filterConfigsByParentIgnores([
      cfg(
        '/project',
        globalIgnore('__tests__/**', 'e2e/**', 'examples/**'),
        ruleEntry(['**/*.ts'], {}),
      ),
      cfg('/project/__tests__/fixtures', ruleEntry(['**/*.ts'], {})),
      cfg('/project/e2e/setup', ruleEntry(['**/*.ts'], {})),
      cfg('/project/packages/app', ruleEntry(['**/*.ts'], {})),
    ]);
    expect(result).toHaveLength(2);
    expect(dirs(result)).toContain('/project');
    expect(dirs(result)).toContain('/project/packages/app');
  });

  test('filters with multiple global ignore entries', () => {
    const result = filterConfigsByParentIgnores([
      cfg(
        '/project',
        globalIgnore('__tests__/**'),
        globalIgnore('e2e/**'),
        ruleEntry(['**/*.ts'], {}),
      ),
      cfg('/project/__tests__/fixtures', ruleEntry(['**/*.ts'], {})),
      cfg('/project/e2e/setup', ruleEntry(['**/*.ts'], {})),
    ]);
    expect(result).toHaveLength(1);
  });

  test('entry-level ignores do not filter nested configs', () => {
    const result = filterConfigsByParentIgnores([
      cfg('/project', {
        files: ['**/*.ts'],
        ignores: ['__tests__/**'],
        rules: { r: 'error' },
      }),
      cfg('/project/__tests__/fixtures', ruleEntry(['**/*.ts'], {})),
    ]);
    expect(result).toHaveLength(2);
  });

  test('sibling configs do not affect each other', () => {
    const result = filterConfigsByParentIgnores([
      cfg('/project/packages/app', ruleEntry(['**/*.ts'], { a: 'error' })),
      cfg('/project/packages/lib', ruleEntry(['**/*.ts'], { b: 'error' })),
    ]);
    expect(result).toHaveLength(2);
  });

  test('parent without global ignores does not filter children', () => {
    const result = filterConfigsByParentIgnores([
      cfg('/project', ruleEntry(['**/*.ts'], { a: 'error' })),
      cfg('/project/packages/app', ruleEntry(['**/*.ts'], { b: 'error' })),
    ]);
    expect(result).toHaveLength(2);
  });

  test('grandparent global ignores filter grandchild', () => {
    const result = filterConfigsByParentIgnores([
      cfg('/project', globalIgnore('vendor/**'), ruleEntry(['**/*.ts'], {})),
      cfg('/project/packages/app', ruleEntry(['**/*.ts'], {})),
      cfg('/project/vendor/lib', ruleEntry(['**/*.ts'], {})),
    ]);
    expect(result).toHaveLength(2);
    expect(dirs(result)).not.toContain('/project/vendor/lib');
  });

  test('intermediate config ignores only affect its children', () => {
    const result = filterConfigsByParentIgnores([
      cfg('/project', ruleEntry(['**/*.ts'], {})),
      cfg(
        '/project/packages/app',
        globalIgnore('generated/**'),
        ruleEntry(['**/*.ts'], {}),
      ),
      cfg('/project/packages/app/generated/config', ruleEntry(['**/*.ts'], {})),
      cfg('/project/packages/lib', ruleEntry(['**/*.ts'], {})),
    ]);
    expect(result).toHaveLength(3);
    expect(dirs(result)).not.toContain(
      '/project/packages/app/generated/config',
    );
    expect(dirs(result)).toContain('/project/packages/lib');
  });

  test('global ignore entry with name field is still recognized', () => {
    const result = filterConfigsByParentIgnores([
      cfg(
        '/project',
        { name: 'global-ignores', ignores: ['vendor/**'] },
        ruleEntry(['**/*.ts'], {}),
      ),
      cfg('/project/vendor/lib', ruleEntry(['**/*.ts'], {})),
    ]);
    expect(result).toHaveLength(1);
  });

  test('real-world monorepo pattern', () => {
    const result = filterConfigsByParentIgnores([
      cfg(
        '/project',
        globalIgnore('**/fixtures/**', '**/dist/**', 'e2e/**'),
        ruleEntry(['**/*.ts'], {}),
      ),
      cfg('/project/packages/app', ruleEntry(['**/*.ts'], { a: 'error' })),
      cfg('/project/packages/lib', ruleEntry(['**/*.ts'], { b: 'error' })),
      cfg(
        '/project/packages/vscode-ext/__tests__/fixtures',
        ruleEntry(['**/*.ts'], {}),
      ),
      cfg('/project/e2e/helpers', ruleEntry(['**/*.ts'], {})),
      cfg('/project/packages/app/dist/gen', ruleEntry(['**/*.ts'], {})),
    ]);
    expect(result).toHaveLength(3);
    expect(dirs(result)).toContain('/project');
    expect(dirs(result)).toContain('/project/packages/app');
    expect(dirs(result)).toContain('/project/packages/lib');
  });

  // --- Glob pattern coverage (picomatch) ---

  test('wildcard in middle: packages/*/dist/**', () => {
    const result = filterConfigsByParentIgnores([
      cfg(
        '/project',
        globalIgnore('packages/*/dist/**'),
        ruleEntry(['**/*.ts'], {}),
      ),
      cfg('/project/packages/app/dist/gen', ruleEntry(['**/*.ts'], {})),
      cfg('/project/packages/app/src', ruleEntry(['**/*.ts'], {})),
    ]);
    expect(result).toHaveLength(2);
    expect(dirs(result)).not.toContain('/project/packages/app/dist/gen');
    expect(dirs(result)).toContain('/project/packages/app/src');
  });

  test('brace expansion: {__tests__,e2e}/**', () => {
    const result = filterConfigsByParentIgnores([
      cfg(
        '/project',
        globalIgnore('{__tests__,e2e}/**'),
        ruleEntry(['**/*.ts'], {}),
      ),
      cfg('/project/__tests__/fixtures', ruleEntry(['**/*.ts'], {})),
      cfg('/project/e2e/setup', ruleEntry(['**/*.ts'], {})),
      cfg('/project/packages/app', ruleEntry(['**/*.ts'], {})),
    ]);
    expect(result).toHaveLength(2);
    expect(dirs(result)).toContain('/project');
    expect(dirs(result)).toContain('/project/packages/app');
  });

  test('bare directory name without glob: dist', () => {
    const result = filterConfigsByParentIgnores([
      cfg('/project', globalIgnore('dist'), ruleEntry(['**/*.ts'], {})),
      cfg('/project/dist', ruleEntry(['**/*.ts'], {})),
      cfg('/project/dist/sub', ruleEntry(['**/*.ts'], {})),
    ]);
    // 'dist' matches the dir exactly, and dist/sub is nested under it
    expect(dirs(result)).not.toContain('/project/dist');
    expect(dirs(result)).not.toContain('/project/dist/sub');
  });

  test('dot-prefixed directories: .cache/**', () => {
    const result = filterConfigsByParentIgnores([
      cfg('/project', globalIgnore('.cache/**'), ruleEntry(['**/*.ts'], {})),
      cfg('/project/.cache/generated', ruleEntry(['**/*.ts'], {})),
    ]);
    expect(result).toHaveLength(1);
    expect(dirs(result)).toEqual(['/project']);
  });

  test('negation pattern is skipped — does not re-include or accidentally filter', () => {
    // Negation (!) is skipped at directory level (aligned with ESLint v10).
    // `vendor/**` filters vendor dirs, `!vendor/keep/**` is ignored.
    const result = filterConfigsByParentIgnores([
      cfg(
        '/project',
        globalIgnore('vendor/**', '!vendor/keep/**'),
        ruleEntry(['**/*.ts'], {}),
      ),
      cfg('/project/vendor/lib', ruleEntry(['**/*.ts'], {})),
      cfg('/project/vendor/keep', ruleEntry(['**/*.ts'], {})),
      // Unrelated dirs should NOT be affected by the negation pattern
      cfg('/project/packages/app', ruleEntry(['**/*.ts'], {})),
    ]);
    // vendor/* filtered by positive pattern
    expect(dirs(result)).not.toContain('/project/vendor/lib');
    expect(dirs(result)).not.toContain('/project/vendor/keep');
    // Unrelated dirs must NOT be filtered (negation pattern skipped safely)
    expect(dirs(result)).toContain('/project/packages/app');
    expect(dirs(result)).toContain('/project');
  });

  test('only-negation global ignore does not filter anything', () => {
    // If global ignore only has negation patterns, nothing should be filtered
    const result = filterConfigsByParentIgnores([
      cfg('/project', globalIgnore('!vendor/**'), ruleEntry(['**/*.ts'], {})),
      cfg('/project/vendor/lib', ruleEntry(['**/*.ts'], {})),
      cfg('/project/packages/app', ruleEntry(['**/*.ts'], {})),
    ]);
    expect(result).toHaveLength(3);
  });

  // --- File-level vs directory-level patterns (ESLint v10 aligned) ---

  test('dir/** blocks traversal, dir/**/* does not (ESLint v10 behavior)', () => {
    // vendor/** = directory-level → filters nested configs
    // vendor/**/* = file-level → does NOT filter (allows traversal)
    const withDirPattern = filterConfigsByParentIgnores([
      cfg('/project', globalIgnore('vendor/**'), ruleEntry(['**/*.ts'], {})),
      cfg('/project/vendor/keep', ruleEntry(['**/*.ts'], {})),
    ]);
    expect(withDirPattern).toHaveLength(1); // vendor/keep filtered

    const withFilePattern = filterConfigsByParentIgnores([
      cfg('/project', globalIgnore('vendor/**/*'), ruleEntry(['**/*.ts'], {})),
      cfg('/project/vendor/keep', ruleEntry(['**/*.ts'], {})),
    ]);
    expect(withFilePattern).toHaveLength(2); // vendor/keep NOT filtered
    expect(dirs(withFilePattern)).toContain('/project/vendor/keep');
  });

  test('dir/* (single-level file glob) does not block traversal', () => {
    const result = filterConfigsByParentIgnores([
      cfg('/project', globalIgnore('vendor/*'), ruleEntry(['**/*.ts'], {})),
      cfg('/project/vendor/keep', ruleEntry(['**/*.ts'], {})),
    ]);
    expect(result).toHaveLength(2);
    expect(dirs(result)).toContain('/project/vendor/keep');
  });

  // --- Mixed global + entry-level ignores in same config ---

  test('config with both global and entry-level ignores: only global filters nested configs', () => {
    const result = filterConfigsByParentIgnores([
      cfg(
        '/project',
        // Global ignore → should filter nested configs in dist/
        globalIgnore('dist/**'),
        // Entry-level ignore → should NOT filter nested configs in test/
        { files: ['**/*.ts'], ignores: ['test/**'], rules: { r: 'error' } },
      ),
      // In dist/ → should be filtered (global ignore)
      cfg('/project/dist/generated', ruleEntry(['**/*.ts'], {})),
      // In test/ → should NOT be filtered (entry-level ignore)
      cfg('/project/test/fixtures', ruleEntry(['**/*.ts'], {})),
      // Normal child → should NOT be filtered
      cfg('/project/packages/app', ruleEntry(['**/*.ts'], {})),
    ]);
    expect(result).toHaveLength(3);
    expect(dirs(result)).not.toContain('/project/dist/generated');
    expect(dirs(result)).toContain('/project/test/fixtures');
    expect(dirs(result)).toContain('/project/packages/app');
  });

  // --- Path edge cases ---

  test('trailing slash in configDirectory is handled', () => {
    const result = filterConfigsByParentIgnores([
      cfg(
        '/project/',
        globalIgnore('__tests__/**'),
        ruleEntry(['**/*.ts'], {}),
      ),
      cfg('/project/__tests__/fixtures', ruleEntry(['**/*.ts'], {})),
    ]);
    expect(result).toHaveLength(1);
  });

  test('trailing slash on both parent and child', () => {
    const result = filterConfigsByParentIgnores([
      cfg('/project/', globalIgnore('vendor/**'), ruleEntry(['**/*.ts'], {})),
      cfg('/project/vendor/lib/', ruleEntry(['**/*.ts'], {})),
    ]);
    expect(result).toHaveLength(1);
  });

  // --- Cross-package ignore isolation ---

  test('child global ignore does NOT bubble up to filter parent', () => {
    const result = filterConfigsByParentIgnores([
      cfg('/project', ruleEntry(['**/*.ts'], {})),
      cfg(
        '/project/packages/app',
        globalIgnore('generated/**'),
        ruleEntry(['**/*.ts'], {}),
      ),
    ]);
    // Parent should never be filtered by child's ignores
    expect(result).toHaveLength(2);
    expect(dirs(result)).toContain('/project');
    expect(dirs(result)).toContain('/project/packages/app');
  });

  test('sibling A global ignore does NOT filter sibling B nested config', () => {
    // app ignores generated/, lib also has a generated/ subdir with a config.
    // lib's generated/ should NOT be filtered by app's ignore.
    const result = filterConfigsByParentIgnores([
      cfg('/project', ruleEntry(['**/*.ts'], {})),
      cfg(
        '/project/packages/app',
        globalIgnore('generated/**'),
        ruleEntry(['**/*.ts'], {}),
      ),
      cfg('/project/packages/app/generated/output', ruleEntry(['**/*.ts'], {})),
      cfg('/project/packages/lib', ruleEntry(['**/*.ts'], {})),
      cfg('/project/packages/lib/generated/output', ruleEntry(['**/*.ts'], {})),
    ]);
    // app/generated should be filtered (by app's ignore)
    expect(dirs(result)).not.toContain(
      '/project/packages/app/generated/output',
    );
    // lib/generated should NOT be filtered (app's ignore doesn't affect lib)
    expect(dirs(result)).toContain('/project/packages/lib/generated/output');
    expect(result).toHaveLength(4);
  });

  test('both siblings have same-name ignored dir, only own children filtered', () => {
    const result = filterConfigsByParentIgnores([
      cfg('/project', ruleEntry(['**/*.ts'], {})),
      cfg(
        '/project/packages/app',
        globalIgnore('dist/**'),
        ruleEntry(['**/*.ts'], {}),
      ),
      cfg('/project/packages/app/dist/gen', ruleEntry(['**/*.ts'], {})),
      cfg(
        '/project/packages/lib',
        globalIgnore('dist/**'),
        ruleEntry(['**/*.ts'], {}),
      ),
      cfg('/project/packages/lib/dist/gen', ruleEntry(['**/*.ts'], {})),
    ]);
    // Both dist/gen configs filtered by their respective parent
    expect(dirs(result)).not.toContain('/project/packages/app/dist/gen');
    expect(dirs(result)).not.toContain('/project/packages/lib/dist/gen');
    // Root + both packages survive
    expect(result).toHaveLength(3);
    expect(dirs(result)).toContain('/project');
    expect(dirs(result)).toContain('/project/packages/app');
    expect(dirs(result)).toContain('/project/packages/lib');
  });

  test('uses real filesystem paths for symlink resolution', () => {
    // Create actual temp dirs to test realpathSync
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
});

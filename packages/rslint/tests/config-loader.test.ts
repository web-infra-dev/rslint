import { describe, test, expect } from '@rstest/core';
import {
  findJSConfig,
  findJSConfigUp,
  findJSConfigsInDir,
  JS_CONFIG_FILES,
} from '../src/config-loader.js';
import fs from 'node:fs';
import path from 'node:path';
import os from 'node:os';

function createTempDir(): string {
  return fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-test-'));
}

function cleanup(dir: string): void {
  fs.rmSync(dir, { recursive: true, force: true });
}

describe('findJSConfig', () => {
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

  test('prefers js over mjs over ts over mts', () => {
    const tmp = createTempDir();
    try {
      // Create all four config files
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
  test('finds config in root directory', () => {
    const tmp = createTempDir();
    try {
      fs.writeFileSync(path.join(tmp, 'rslint.config.js'), 'export default []');
      const result = findJSConfigsInDir(tmp);
      expect(result).toEqual([path.join(tmp, 'rslint.config.js')]);
    } finally {
      cleanup(tmp);
    }
  });

  test('finds configs in nested directories', () => {
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
      const result = findJSConfigsInDir(tmp).sort();
      expect(result).toEqual(
        [
          path.join(tmp, 'rslint.config.js'),
          path.join(tmp, 'packages', 'foo', 'rslint.config.ts'),
          path.join(tmp, 'packages', 'bar', 'rslint.config.mjs'),
        ].sort(),
      );
    } finally {
      cleanup(tmp);
    }
  });

  test('skips node_modules', () => {
    const tmp = createTempDir();
    try {
      fs.mkdirSync(path.join(tmp, 'node_modules', 'pkg'), { recursive: true });
      fs.writeFileSync(
        path.join(tmp, 'node_modules', 'pkg', 'rslint.config.js'),
        'export default []',
      );
      fs.writeFileSync(path.join(tmp, 'rslint.config.js'), 'export default []');
      const result = findJSConfigsInDir(tmp);
      expect(result).toEqual([path.join(tmp, 'rslint.config.js')]);
    } finally {
      cleanup(tmp);
    }
  });

  test('skips .git directory', () => {
    const tmp = createTempDir();
    try {
      fs.mkdirSync(path.join(tmp, '.git', 'hooks'), { recursive: true });
      fs.writeFileSync(
        path.join(tmp, '.git', 'rslint.config.js'),
        'export default []',
      );
      fs.writeFileSync(path.join(tmp, 'rslint.config.js'), 'export default []');
      const result = findJSConfigsInDir(tmp);
      expect(result).toEqual([path.join(tmp, 'rslint.config.js')]);
    } finally {
      cleanup(tmp);
    }
  });

  test('returns empty array when no configs found', () => {
    const tmp = createTempDir();
    try {
      fs.mkdirSync(path.join(tmp, 'src'), { recursive: true });
      fs.writeFileSync(path.join(tmp, 'src', 'index.ts'), 'const x = 1;');
      const result = findJSConfigsInDir(tmp);
      expect(result).toEqual([]);
    } finally {
      cleanup(tmp);
    }
  });

  test('handles non-existent directory gracefully', () => {
    const result = findJSConfigsInDir('/tmp/does-not-exist-99999');
    expect(result).toEqual([]);
  });

  test('does not traverse into nested node_modules', () => {
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
      const result = findJSConfigsInDir(tmp);
      expect(result).toEqual([
        path.join(tmp, 'packages', 'foo', 'rslint.config.js'),
      ]);
    } finally {
      cleanup(tmp);
    }
  });

  test('finds all config file types', () => {
    const tmp = createTempDir();
    try {
      for (const name of JS_CONFIG_FILES) {
        fs.writeFileSync(path.join(tmp, name), 'export default []');
      }
      const result = findJSConfigsInDir(tmp).sort();
      expect(result).toEqual(
        JS_CONFIG_FILES.map(name => path.join(tmp, name)).sort(),
      );
    } finally {
      cleanup(tmp);
    }
  });
});

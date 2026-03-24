import fs from 'node:fs';
import path from 'node:path';

export const JS_CONFIG_FILES = [
  'rslint.config.js',
  'rslint.config.mjs',
  'rslint.config.ts',
  'rslint.config.mts',
];

const SCAN_EXCLUDE_DIRS = new Set(['node_modules', '.git']);

export function findJSConfig(cwd: string): string | null {
  for (const name of JS_CONFIG_FILES) {
    const p = path.join(cwd, name);
    if (fs.existsSync(p)) return p;
  }
  return null;
}

/**
 * Walk upward from startDir to the filesystem root, returning the first
 * rslint JS/TS config file found. Returns null if none is found.
 */
export function findJSConfigUp(startDir: string): string | null {
  let dir = path.resolve(startDir);
  while (true) {
    const found = findJSConfig(dir);
    if (found) return found;
    const parent = path.dirname(dir);
    if (parent === dir) return null;
    dir = parent;
  }
}

/**
 * Recursively scan a directory for all rslint JS/TS config files.
 * Skips node_modules and .git directories (aligned with ESLint defaults).
 * Uses native fs.globSync when available (Node 22+, C++ impl), falls back
 * to hand-written recursive walk.
 */
export function findJSConfigsInDir(startDir: string): string[] {
  const resolved = path.resolve(startDir);

  // Node 22+ native globSync (C++ implementation, faster)
  if (typeof (fs as any).globSync === 'function') {
    const pattern = '**/rslint.config.{js,mjs,ts,mts}';
    return (fs as any)
      .globSync(pattern, {
        cwd: resolved,
        exclude: (f: string) => SCAN_EXCLUDE_DIRS.has(path.basename(f)),
      })
      .map((p: string) => path.join(resolved, p));
  }

  // Fallback: recursive walk
  const configs: string[] = [];
  const walk = (dir: string): void => {
    let entries: fs.Dirent[];
    try {
      entries = fs.readdirSync(dir, { withFileTypes: true });
    } catch {
      return;
    }
    for (const entry of entries) {
      if (entry.isDirectory()) {
        if (SCAN_EXCLUDE_DIRS.has(entry.name)) continue;
        walk(path.join(dir, entry.name));
      } else if (JS_CONFIG_FILES.includes(entry.name)) {
        configs.push(path.join(dir, entry.name));
      }
    }
  };
  walk(resolved);
  return configs;
}

/**
 * Discover JS/TS config files for the given targets.
 *
 * For file arguments, config is searched upward from each file's directory,
 * so different files can find different configs (monorepo multi-config).
 *
 * For no-args and directory arguments, config is searched upward from the
 * starting point AND nested configs within the scope are scanned. This
 * ensures sub-package configs in a monorepo are discovered when linting
 * from the root.
 */
export function discoverConfigs(
  files: string[],
  dirs: string[],
  cwd: string,
  explicitConfig: string | null,
): Map<string, string> {
  // Map: configPath -> configDirectory
  const configs = new Map<string, string>();

  const addConfig = (configPath: string): void => {
    if (!configs.has(configPath)) {
      configs.set(configPath, path.dirname(configPath));
    }
  };

  if (explicitConfig) {
    const resolved = path.resolve(cwd, explicitConfig);
    addConfig(resolved);
    return configs;
  }

  // Collect unique start directories for upward config search
  const startDirs = new Set<string>();
  // Collect directories to scan for nested configs
  const scanDirs: string[] = [];

  if (files.length === 0 && dirs.length === 0) {
    startDirs.add(cwd);
    scanDirs.push(cwd);
  }

  // Deduplicate file directories before searching
  for (const file of files) {
    startDirs.add(path.dirname(file));
  }

  for (const dir of dirs) {
    startDirs.add(dir);
    scanDirs.push(dir);
  }

  // Upward traversal: find nearest config for each start directory
  for (const startDir of startDirs) {
    const configPath = findJSConfigUp(startDir);
    if (configPath) {
      addConfig(configPath);
    }
  }

  // Scan for nested configs within the target scope (no-args and dir-args).
  // Broken configs are tolerated by runWithJSConfigs (skipped with warning).
  for (const dir of scanDirs) {
    for (const configPath of findJSConfigsInDir(dir)) {
      addConfig(configPath);
    }
  }

  return configs;
}

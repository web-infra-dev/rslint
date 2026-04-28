import fs from 'node:fs';
import path from 'node:path';
import picomatch from 'picomatch';
import { globSync as tinyglobbySync } from 'tinyglobby';
import { type RslintConfigEntry } from '../define-config.ts';

export const JS_CONFIG_FILES = [
  'rslint.config.js',
  'rslint.config.mjs',
  'rslint.config.ts',
  'rslint.config.mts',
];

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
 * Uses tinyglobby (fdir-backed) for fast directory traversal.
 *
 * tinyglobby returns POSIX-style paths even on Windows, so the result is
 * normalized through path.normalize to match the native separator that
 * findJSConfigUp / path.join produce. Without this, Map<configPath, ...>
 * dedupe against findJSConfigUp results fails on Windows.
 */
export function findJSConfigsInDir(startDir: string): string[] {
  const resolved = path.resolve(startDir);
  return tinyglobbySync(['**/rslint.config.{js,mjs,ts,mts}'], {
    cwd: resolved,
    absolute: true,
    dot: true,
    ignore: ['**/node_modules/**', '**/.git/**'],
  }).map((p) => path.normalize(p));
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

/**
 * Check if a config entry is a "global ignore" entry — an entry with only
 * `ignores` and no other meaningful fields. Aligns with ESLint flat config
 * semantics where such entries prevent directory traversal.
 */
function isGlobalIgnoreEntry(
  entry: RslintConfigEntry,
): entry is Required<Pick<RslintConfigEntry, 'ignores'>> {
  const ignores = entry.ignores;
  if (!Array.isArray(ignores) || ignores.length === 0) return false;

  return (
    entry.files == null &&
    entry.rules == null &&
    entry.plugins == null &&
    entry.languageOptions == null &&
    entry.settings == null
  );
}

/**
 * Extract global ignore patterns from a config's entries.
 */
function getGlobalIgnores(entries: RslintConfigEntry[]): string[] {
  const patterns: string[] = [];
  for (const entry of entries) {
    if (isGlobalIgnoreEntry(entry)) {
      for (const pattern of entry.ignores) {
        patterns.push(pattern);
      }
    }
  }
  return patterns;
}

/**
 * Check if a directory path is matched by any of the given ignore patterns.
 * Patterns are resolved relative to the parent config's directory.
 * Uses picomatch for full glob support (**, *, {}, etc.).
 *
 * A directory is considered ignored if the pattern matches the directory
 * itself or any of its ancestor paths. For example, pattern `__tests__/**`
 * matches both `__tests__/` and `__tests__/fixtures/`.
 */
function isDirIgnoredByPatterns(
  dirPath: string,
  patterns: string[],
  parentConfigDir: string,
): boolean {
  const relDir = path.relative(parentConfigDir, dirPath);
  if (!relDir || relDir.startsWith('..')) return false;

  const normalizedRelDir = relDir.split(path.sep).join('/');

  for (const pattern of patterns) {
    // Skip empty or negation patterns.
    if (!pattern || pattern.startsWith('!')) continue;

    // Skip file-level patterns (ending with /**/* or /*). These only ignore
    // files inside a directory, NOT the directory itself. ESLint v10's
    // isDirectoryIgnored does not block traversal for file-level patterns,
    // allowing `!` re-include to work for files inside.
    // Only directory-level patterns (ending with /** or /) block traversal.
    if (
      pattern.endsWith('/**/*') ||
      (pattern.endsWith('/*') && !pattern.endsWith('/**'))
    )
      continue;

    const isMatch = picomatch(pattern, { dot: true });

    // Check if the pattern matches the directory itself or a file inside it.
    // We test: the dir path, dir path + trailing slash, and a synthetic
    // child path to handle patterns like `dir/**`.
    if (
      isMatch(normalizedRelDir) ||
      isMatch(normalizedRelDir + '/') ||
      isMatch(normalizedRelDir + '/x')
    ) {
      return true;
    }

    // For nested dirs, also check if any parent segment matches.
    // e.g., pattern `__tests__/**` should match `__tests__/fixtures/deep`.
    const segments = normalizedRelDir.split('/');
    for (let i = 1; i < segments.length; i++) {
      const partial = segments.slice(0, i).join('/');
      if (
        isMatch(partial) ||
        isMatch(partial + '/') ||
        isMatch(partial + '/x')
      ) {
        return true;
      }
    }
  }

  return false;
}

export interface ConfigEntry {
  configDirectory: string;
  entries: RslintConfigEntry[];
}

/**
 * Filter out nested configs whose directory is covered by an ancestor config's
 * global ignores. Aligns with ESLint v10 behavior: when traversing directories,
 * global ignores in a parent config prevent entering ignored directories, so
 * nested configs in those directories are never discovered.
 *
 * Example: root config has `{ ignores: ['__tests__/**'] }`.
 * A nested config at `__tests__/fixtures/rslint.config.js` is filtered out
 * because `__tests__/fixtures/` is within the root config's global ignores.
 */
export function filterConfigsByParentIgnores(
  configEntries: ConfigEntry[],
): ConfigEntry[] {
  if (configEntries.length <= 1) return configEntries;

  // Resolve symlinks and normalize trailing slashes for reliable ancestor checks.
  const resolvedDirs = new Map<ConfigEntry, string>();
  for (const entry of configEntries) {
    let dir = entry.configDirectory.replace(/[/\\]+$/, '');
    try {
      dir = fs.realpathSync(dir);
    } catch {
      // Keep the original (stripped) path if realpath fails
    }
    resolvedDirs.set(entry, dir);
  }

  // Sort by directory depth (shallowest first) so parents are processed first
  const sorted = [...configEntries].sort(
    (a, b) =>
      (resolvedDirs.get(a)?.length ?? 0) - (resolvedDirs.get(b)?.length ?? 0),
  );

  const result: ConfigEntry[] = [];

  for (const config of sorted) {
    let ignored = false;
    const configDir = resolvedDirs.get(config)!;

    for (const parent of result) {
      const parentDir = resolvedDirs.get(parent)!;
      if (
        !configDir.startsWith(parentDir + path.sep) &&
        configDir !== parentDir
      ) {
        continue;
      }

      const globalIgnores = getGlobalIgnores(parent.entries);
      if (globalIgnores.length === 0) continue;

      if (isDirIgnoredByPatterns(configDir, globalIgnores, parentDir)) {
        ignored = true;
        break;
      }
    }

    if (!ignored) {
      result.push(config);
    }
  }

  return result;
}

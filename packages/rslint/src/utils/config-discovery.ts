import fs from 'node:fs';
import path from 'node:path';
import { glob } from 'tinyglobby';
import { JS_CONFIG_FILES } from '../config/config-loader.ts';

export {
  filterConfigsByParentIgnores,
  type ConfigEntry,
} from '../config/config-hierarchy.ts';

export { JS_CONFIG_FILES };

export function findJSConfig(cwd: string): string | null {
  for (const name of JS_CONFIG_FILES) {
    const p = path.join(cwd, name);
    try {
      if (fs.statSync(p).isFile()) return p;
    } catch {
      // Continue through the configured filename priority.
    }
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
 * Find the governing config for one target without discarding lexical path
 * ownership. Physical ancestry is only a fallback when the lexical ancestry
 * contains no config.
 */
export function findJSConfigForTarget(
  targetPath: string,
  isDirectory: boolean,
): string | null {
  const lexicalPath = path.resolve(targetPath);
  const lexicalDirectory = isDirectory
    ? lexicalPath
    : path.dirname(lexicalPath);
  const lexicalConfig = findJSConfigUp(lexicalDirectory);
  if (lexicalConfig) return lexicalConfig;

  return findJSConfigFromCanonicalTarget(
    lexicalPath,
    lexicalDirectory,
    isDirectory,
  );
}

function findJSConfigFromCanonicalTarget(
  lexicalPath: string,
  lexicalDirectory: string,
  isDirectory: boolean,
  configByCanonicalDirectory?: Map<string, string | null>,
): string | null {
  try {
    const canonicalPath = fs.realpathSync(lexicalPath);
    const canonicalDirectory = isDirectory
      ? canonicalPath
      : path.dirname(canonicalPath);
    if (
      path.normalize(canonicalDirectory) !== path.normalize(lexicalDirectory)
    ) {
      const normalizedDirectory = path.normalize(canonicalDirectory);
      const cached = configByCanonicalDirectory?.get(normalizedDirectory);
      if (cached !== undefined) return cached;
      const configPath = findJSConfigUp(normalizedDirectory);
      configByCanonicalDirectory?.set(normalizedDirectory, configPath);
      return configPath;
    }
  } catch {
    // Missing or inaccessible targets have no physical fallback.
  }
  return null;
}

/**
 * Recursively scan a directory for effective rslint JS/TS config files.
 * Skips node_modules and .git directories (aligned with ESLint defaults).
 * Uses tinyglobby's async crawl for filesystem traversal performance and edge
 * cases, then applies rslint config semantics to the small candidate set.
 */
export async function findJSConfigsInDir(startDir: string): Promise<string[]> {
  const resolved = path.resolve(startDir);
  const matches = await glob(
    JS_CONFIG_FILES.map((name) => `**/${name}`),
    {
      cwd: resolved,
      absolute: true,
      dot: true,
      followSymbolicLinks: false,
      ignore: ['**/node_modules/**', '**/.git/**'],
    },
  );

  const candidateDirectories = new Set<string>();
  for (const match of matches) {
    candidateDirectories.add(path.dirname(path.normalize(match)));
  }

  const effectiveConfigs: string[] = [];
  for (const configDirectory of candidateDirectories) {
    const configPath = findJSConfig(configDirectory);
    if (configPath) effectiveConfigs.push(path.normalize(configPath));
  }
  return effectiveConfigs;
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
export async function discoverConfigs(
  files: string[],
  dirs: string[],
  cwd: string,
  explicitConfig: string | null,
  fileConfigs?: ReadonlyMap<string, string | null>,
  directoryConfigs?: ReadonlyMap<string, string | null>,
): Promise<Map<string, string>> {
  // Map: configPath -> configDirectory
  const configs = new Map<string, string>();

  const addConfig = (
    configPath: string,
    configDirectory = path.dirname(configPath),
  ): void => {
    if (!configs.has(configPath)) {
      configs.set(configPath, configDirectory);
    }
  };

  if (explicitConfig) {
    const resolved = path.resolve(cwd, explicitConfig);
    // Explicit --config follows ESLint flat config semantics: files, ignores,
    // and project patterns resolve from the invocation cwd.
    addConfig(resolved, cwd);
    return configs;
  }

  // Collect directories to scan for nested configs
  const scanDirs: string[] = [];

  if (files.length === 0 && dirs.length === 0) {
    const configPath = findJSConfigForTarget(cwd, true);
    if (configPath) addConfig(configPath);
    scanDirs.push(cwd);
  }

  const resolvedFileConfigs = fileConfigs ?? findJSConfigsForFiles(files);
  for (const file of files) {
    const configPath = resolvedFileConfigs.get(path.normalize(file));
    if (configPath) addConfig(configPath);
  }

  for (const dir of dirs) {
    const configPath =
      directoryConfigs?.get(path.normalize(dir)) ??
      findJSConfigForTarget(dir, true);
    if (configPath) addConfig(configPath);
    scanDirs.push(dir);
  }

  // Scan for nested configs within the target scope (no-args and dir-args).
  // Broken configs are tolerated by runWithJSConfigs (skipped with warning).
  // Serial await (not Promise.all over scanDirs): a single async glob already
  // saturates the libuv thread pool, so parallelizing across scanDirs adds no
  // speed — it only creates I/O contention when scanDirs overlap (e.g. `.` plus
  // a nested subdir each crawling the shared subtree at the same time).
  for (const dir of scanDirs) {
    for (const configPath of await findJSConfigsInDir(dir)) {
      addConfig(configPath);
    }
  }

  return configs;
}

/** Resolve each explicit file's config once while preserving lexical ownership. */
export function findJSConfigsForFiles(
  files: readonly string[],
): Map<string, string | null> {
  const result = new Map<string, string | null>();
  const lexicalConfigByDirectory = new Map<string, string | null>();
  const canonicalConfigByDirectory = new Map<string, string | null>();
  for (const rawFile of files) {
    const file = path.normalize(rawFile);
    const fileDirectory = path.dirname(file);
    let configPath = lexicalConfigByDirectory.get(fileDirectory);
    if (configPath === undefined) {
      configPath = findJSConfigUp(fileDirectory);
      lexicalConfigByDirectory.set(fileDirectory, configPath);
    }
    if (!configPath) {
      configPath = findJSConfigFromCanonicalTarget(
        path.resolve(file),
        fileDirectory,
        false,
        canonicalConfigByDirectory,
      );
    }
    result.set(file, configPath);
  }
  return result;
}

/** Resolve each explicit directory's config once while preserving its alias. */
export function findJSConfigsForDirectories(
  directories: readonly string[],
): Map<string, string | null> {
  const result = new Map<string, string | null>();
  const lexicalConfigByDirectory = new Map<string, string | null>();
  const canonicalConfigByDirectory = new Map<string, string | null>();
  for (const rawDirectory of directories) {
    const directory = path.normalize(rawDirectory);
    let configPath = lexicalConfigByDirectory.get(directory);
    if (configPath === undefined) {
      configPath = findJSConfigUp(directory);
      lexicalConfigByDirectory.set(directory, configPath);
    }
    if (!configPath) {
      configPath = findJSConfigFromCanonicalTarget(
        path.resolve(directory),
        directory,
        true,
        canonicalConfigByDirectory,
      );
    }
    result.set(directory, configPath);
  }
  return result;
}

function caseDifferenceTraversesSymlink(
  leftDirectory: string,
  rightDirectory: string,
): boolean {
  const leftRoot = path.parse(leftDirectory).root;
  const rightRoot = path.parse(rightDirectory).root;
  const leftSegments = leftDirectory.slice(leftRoot.length).split(path.sep);
  const rightSegments = rightDirectory.slice(rightRoot.length).split(path.sep);
  let leftCurrent = leftRoot;
  let rightCurrent = rightRoot;
  for (let index = 0; index < leftSegments.length; index++) {
    const leftSegment = leftSegments[index];
    const rightSegment = rightSegments[index];
    if (!leftSegment || !rightSegment) continue;
    leftCurrent = path.join(leftCurrent, leftSegment);
    rightCurrent = path.join(rightCurrent, rightSegment);
    if (leftSegment === rightSegment) continue;
    try {
      const leftInfo = fs.lstatSync(leftCurrent);
      const rightInfo = fs.lstatSync(rightCurrent);
      const leftIsSymlink = leftInfo.isSymbolicLink();
      const rightIsSymlink = rightInfo.isSymbolicLink();
      if (leftIsSymlink || rightIsSymlink) {
        // Alternate casing of one symlink/junction is one native path. Distinct
        // symlink entries that happen to share a target remain separate owners.
        if (
          leftIsSymlink &&
          rightIsSymlink &&
          (leftInfo.dev !== 0 || leftInfo.ino !== 0) &&
          leftInfo.dev === rightInfo.dev &&
          leftInfo.ino === rightInfo.ino
        ) {
          continue;
        }
        return true;
      }
    } catch {
      return true;
    }
  }
  return false;
}

function isNativeCaseAlias(
  leftPath: string,
  leftDirectory: string,
  rightPath: string,
  rightDirectory: string,
): boolean {
  if (
    leftPath === rightPath ||
    leftPath.toLowerCase() !== rightPath.toLowerCase()
  ) {
    return false;
  }
  try {
    if (
      path.normalize(fs.realpathSync.native(leftDirectory)) !==
        path.normalize(fs.realpathSync.native(rightDirectory)) ||
      path.normalize(fs.realpathSync.native(leftPath)) !==
        path.normalize(fs.realpathSync.native(rightPath))
    ) {
      return false;
    }
  } catch {
    return false;
  }
  return !caseDifferenceTraversesSymlink(leftDirectory, rightDirectory);
}

/** Find an already-known spelling of the same native case-aliased config. */
export function findNativeCaseAliasConfigPath(
  configPath: string,
  configDirectory: string,
  configs: ReadonlyMap<string, string>,
): string | null {
  const normalizedPath = path.normalize(configPath);
  const normalizedDirectory = path.normalize(configDirectory);
  const foldedPath = normalizedPath.toLowerCase();
  for (const [candidatePath, candidateDirectory] of configs) {
    if (path.normalize(candidatePath).toLowerCase() !== foldedPath) continue;
    if (
      isNativeCaseAlias(
        normalizedPath,
        normalizedDirectory,
        path.normalize(candidatePath),
        path.normalize(candidateDirectory),
      )
    ) {
      return candidatePath;
    }
  }
  return null;
}

/** Coalesce native case aliases without collapsing explicit symlink roots. */
export function coalesceCaseAliasedConfigs(
  configs: Map<string, string>,
  explicitTargets: Map<string, string[]> = new Map(),
  explicitDirectories: Map<string, string[]> = new Map(),
): {
  configs: Map<string, string>;
  explicitFileTargetsByConfigPath: Map<string, string[]>;
  explicitDirectoryTargetsByConfigPath: Map<string, string[]>;
} {
  const coalescedConfigs = new Map<string, string>();
  const coalescedTargets = new Map<string, string[]>();
  const coalescedDirectories = new Map<string, string[]>();

  for (const [rawConfigPath, rawConfigDirectory] of configs) {
    const configPath = path.normalize(rawConfigPath);
    const configDirectory = path.normalize(rawConfigDirectory);
    const representative = findNativeCaseAliasConfigPath(
      configPath,
      configDirectory,
      coalescedConfigs,
    );
    const selectedPath = representative ?? configPath;
    if (!representative) {
      coalescedConfigs.set(configPath, configDirectory);
    }

    const mergeTargets = (
      source: Map<string, string[]>,
      destination: Map<string, string[]>,
    ): void => {
      const targets = source.get(rawConfigPath) ?? source.get(configPath);
      if (!targets || targets.length === 0) return;
      const existing = destination.get(selectedPath) ?? [];
      destination.set(selectedPath, [...new Set([...existing, ...targets])]);
    };
    mergeTargets(explicitTargets, coalescedTargets);
    mergeTargets(explicitDirectories, coalescedDirectories);
  }

  return {
    configs: coalescedConfigs,
    explicitFileTargetsByConfigPath: coalescedTargets,
    explicitDirectoryTargetsByConfigPath: coalescedDirectories,
  };
}

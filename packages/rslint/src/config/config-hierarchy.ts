import path from 'node:path';
import picomatch from 'picomatch';

export interface ConfigEntry {
  configDirectory: string;
  entries: unknown[];
}

/**
 * Check if a config entry is a "global ignore" entry: an entry with only
 * `ignores` and an optional `name`.
 */
function isGlobalIgnoreEntry(
  entry: unknown,
): entry is { ignores: string[]; name?: string } {
  if (entry === null || typeof entry !== 'object' || Array.isArray(entry)) {
    return false;
  }

  const ignores = 'ignores' in entry ? entry.ignores : undefined;
  if (
    !Array.isArray(ignores) ||
    ignores.length === 0 ||
    ignores.some((pattern) => typeof pattern !== 'string')
  ) {
    return false;
  }

  return Object.keys(entry).every((key) => key === 'ignores' || key === 'name');
}

function getGlobalIgnores(entries: unknown[]): string[] {
  const patterns: string[] = [];
  for (const entry of entries) {
    if (isGlobalIgnoreEntry(entry)) patterns.push(...entry.ignores);
  }
  return patterns;
}

function normalizeGlobPattern(pattern: string): string {
  const normalized = path.posix.normalize(pattern.replaceAll('\\', '/'));
  return normalized === '.' ? '' : normalized.replace(/^\.\//, '');
}

/** Check whether a parent config's global ignores block directory traversal. */
function isDirIgnoredByPatterns(
  dirPath: string,
  patterns: string[],
  parentConfigDir: string,
): boolean {
  const relDir = path.relative(parentConfigDir, dirPath);
  if (!isRelativeChildPath(relDir)) return false;

  const segments = relDir.split(path.sep).filter(Boolean);
  let current = '';

  // ESLint checks each ancestor directory in order. Once an ancestor remains
  // ignored after applying the patterns sequentially, descendants cannot be
  // reached and therefore cannot be re-included by a later pattern.
  for (const segment of segments) {
    current = current ? `${current}/${segment}` : segment;
    let ignored = false;

    for (const rawPattern of patterns) {
      if (!rawPattern) continue;
      const negated = rawPattern.startsWith('!');
      if (negated !== ignored) continue;

      const pattern = normalizeGlobPattern(
        negated ? rawPattern.slice(1) : rawPattern,
      );
      if (!pattern) continue;
      const isMatch = picomatch(pattern, { dot: true });
      if (isMatch(current) || isMatch(`${current}/`)) {
        ignored = !negated;
      }
    }

    if (ignored) return true;
  }

  return false;
}

function isRelativeChildPath(relPath: string): boolean {
  return (
    relPath !== '' &&
    relPath !== '..' &&
    !relPath.startsWith(`..${path.sep}`) &&
    !path.isAbsolute(relPath)
  );
}

function isSameOrChildDirectory(parentDir: string, childDir: string): boolean {
  const parent = path.normalize(parentDir);
  const child = path.normalize(childDir);
  if (parent === child) return true;
  const prefix = parent.endsWith(path.sep) ? parent : `${parent}${path.sep}`;
  return child.startsWith(prefix);
}

function normalizeConfigDirectory(configDirectory: string): string {
  let dir = configDirectory;
  const root = path.parse(dir).root;
  while (dir.length > root.length && /[/\\]$/.test(dir)) {
    dir = dir.slice(0, -1);
  }
  return path.normalize(dir);
}

/**
 * Filter loaded nested config candidates whose directory is covered by an
 * ancestor config's global ignores. Candidate discovery and module loading are
 * intentionally independent from this effective-catalog step.
 *
 * `forceIncludeConfigDirectories` is for explicit CLI file targets: ESLint
 * resolves the nearest config for an explicit file even when parent directory
 * traversal would have been blocked.
 */
export function filterConfigsByParentIgnores<T extends ConfigEntry>(
  configEntries: T[],
  forceIncludeConfigDirectories = new Set<string>(),
): T[] {
  if (configEntries.length <= 1) return configEntries;

  const forceIncludedDirs = new Set(
    [...forceIncludeConfigDirectories].map(normalizeConfigDirectory),
  );
  const normalizedEntries = configEntries
    .map((config) => ({
      config,
      directory: normalizeConfigDirectory(config.configDirectory),
    }))
    // Ordering only ensures an ancestor is evaluated before its descendants;
    // containment, not sort position, determines the effective parent.
    .sort((a, b) => a.directory.length - b.directory.length);
  const result: typeof normalizedEntries = [];

  for (const entry of normalizedEntries) {
    let ignored = false;
    let nearestParent: (typeof normalizedEntries)[number] | undefined;

    for (const parent of result) {
      if (!isSameOrChildDirectory(parent.directory, entry.directory)) continue;
      if (
        nearestParent === undefined ||
        parent.directory.length > nearestParent.directory.length
      ) {
        nearestParent = parent;
      }
    }

    if (nearestParent !== undefined) {
      const globalIgnores = getGlobalIgnores(nearestParent.config.entries);
      ignored =
        globalIgnores.length > 0 &&
        isDirIgnoredByPatterns(
          entry.directory,
          globalIgnores,
          nearestParent.directory,
        );
    }

    if (!ignored || forceIncludedDirs.has(entry.directory)) result.push(entry);
  }

  return result.map(({ config }) => config);
}

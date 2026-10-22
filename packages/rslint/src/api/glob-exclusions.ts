import path from 'node:path';
import { convertPathToPattern } from 'tinyglobby';
import picomatch from 'picomatch';

/** Compile exclusions once for both literal files and glob results. */
export function compileGlobExclusions(
  patterns: readonly string[],
  cwd: string,
  paths: Pick<typeof path, 'sep' | 'isAbsolute'> = path,
): { ignore: string[]; matches(filePath: string): boolean } {
  const windowsVolume = /^(?:[a-z]:|\/\/[^/]+\/[^/]+)(?=\/|$)/i;
  const normalize = (pattern: string): string => {
    if (paths.sep === '\\') {
      // Accept Windows separators without consuming escaped glob syntax.
      pattern = pattern.replace(/\\(?![()[\]{}!+@])/g, '/');
    }
    const volume =
      paths.sep === '\\' ? windowsVolume.exec(pattern)?.[0] || '' : '';
    // Normalize only the suffix: `..` cannot remove a drive or UNC share.
    let suffix = path.posix.normalize(pattern.slice(volume.length) || '/');
    if (suffix !== '/' && suffix.endsWith('/')) suffix = suffix.slice(0, -1);
    return volume.toLowerCase() + suffix;
  };

  const normalizedCwd = normalize(cwd.split(paths.sep).join('/'));
  const escapedCwd = convertPathToPattern(normalizedCwd);
  const cwdPrefix = escapedCwd.endsWith('/') ? escapedCwd : `${escapedCwd}/`;
  const relativePatterns: string[] = [];
  const absolutePatterns: string[] = [];
  const ignore: string[] = [];

  for (const input of patterns) {
    if (!input) continue;
    let pattern = normalize(input);
    const absolute = paths.isAbsolute(pattern);
    if (absolute && paths.sep === '\\' && pattern.startsWith('/')) {
      // A root-relative Windows pattern inherits cwd's drive or UNC share.
      if (!pattern.startsWith('//')) {
        pattern = (windowsVolume.exec(normalizedCwd)?.[0] || '') + pattern;
      }
    }
    if (!absolute) {
      // Cancel parent segments against the escaped cwd, never a host path API
      // that interprets glob escapes as Windows separators.
      pattern =
        path.posix.relative(escapedCwd, normalize(`${cwdPrefix}${pattern}`)) ||
        '.';
    }
    // Excluding a matching directory also excludes its descendants, including
    // when the directory itself matches a wildcard such as `dist*`.
    const expanded = pattern.endsWith('/**')
      ? [pattern]
      : [pattern, normalize(`${pattern}/**`)];
    (absolute ? absolutePatterns : relativePatterns).push(...expanded);

    const relative = absolute
      ? pattern === escapedCwd
        ? '.'
        : pattern.startsWith(cwdPrefix)
          ? pattern.slice(cwdPrefix.length)
          : undefined
      : pattern;
    // tinyglobby relativizes absolute patterns and parses leading negation.
    // Only give it equivalent cwd-relative pruning hints; final matching below
    // handles ancestor wildcards and leading literal `!` without that rewrite.
    if (
      relative !== undefined &&
      relative !== '..' &&
      !relative.startsWith('../') &&
      !relative.startsWith('!')
    ) {
      ignore.push(relative);
    }
  }

  // Exclusion roles are already classified by the caller. cspell:ignore nonegate
  const options = { dot: true, nocase: false, nonegate: true, posix: true };
  const relativeMatcher = picomatch(relativePatterns, options);
  const absoluteMatcher = picomatch(absolutePatterns, options);
  return {
    ignore,
    matches(filePath) {
      const absolute = normalize(filePath.split(paths.sep).join('/'));
      return (
        relativeMatcher(path.posix.relative(normalizedCwd, absolute) || '.') ||
        absoluteMatcher(absolute)
      );
    },
  };
}

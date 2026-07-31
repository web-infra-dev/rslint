import path from 'node:path';

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

/** True for absolute POSIX or Windows paths (including drive-letter forms). */
function isAbsolutePattern(pattern: string): boolean {
  return path.posix.isAbsolute(pattern) || path.win32.isAbsolute(pattern);
}

/**
 * Escape glob metacharacters in a resolved basePath so the directory is treated
 * literally when spliced into glob patterns. ESLint treats basePath as a literal
 * path; without escaping, basePath "packages/[locale]" would be interpreted as a
 * character class and never match. Character-class escapes ([*], [?], [[] ...)
 * are understood by both the Go (doublestar) and TS (picomatch) matchers on
 * every platform.
 */
function escapeGlobBasePath(s: string): string {
  return s
    .replaceAll('[', '[[]')
    .replaceAll('*', '[*]')
    .replaceAll('?', '[?]')
    .replaceAll('{', '[{]');
}

/**
 * Resolve an entry `basePath` to a glob-escaped POSIX path relative to
 * `configDirectory`. Absolute base paths are rebased via `path.relative`.
 * Empty / "." means the config root itself (no prefix).
 *
 * A basePath resolving outside `configDirectory` is rejected: the resulting
 * `../`-prefixed (or cross-root) patterns could never match under the Go
 * engine's single-match-root model, so it fails fast instead of silently
 * contributing nothing (mirrors the Go `ResolveBasePaths` rejection).
 */
export function resolveRelativeBasePath(
  basePath: string,
  configDirectory: string,
): string {
  const root = configDirectory || process.cwd();
  const absoluteBase = path.resolve(root, basePath);
  let relative = path.relative(root, absoluteBase).replace(/\\/g, '/');
  if (
    relative === '..' ||
    relative.startsWith('../') ||
    isAbsolutePattern(relative)
  ) {
    throw new Error(
      `basePath "${basePath}" resolves outside the config match root "${root}"; ` +
        'basePath must be a subdirectory of the config root',
    );
  }
  relative = path.posix.normalize(relative);
  if (relative === '.' || relative === '') {
    return '';
  }
  return escapeGlobBasePath(relative).replace(/^\.\//, '');
}

/**
 * Rebase one glob or path pattern under the glob-escaped `relativeBase`
 * produced by {@link resolveRelativeBasePath}.
 *
 * - Leading `!` negation is preserved (odd count → negated).
 * - Leading `./` is stripped before joining.
 * - Absolute patterns are left unchanged (not prefixed).
 * - Empty relativeBase is a no-op besides `./` cleanup.
 */
export function rebasePattern(pattern: string, relativeBase: string): string {
  let negated = false;
  let body = pattern;
  while (body.startsWith('!')) {
    negated = !negated;
    body = body.slice(1);
  }

  body = body.replace(/\\/g, '/');
  while (body.startsWith('./')) {
    body = body.slice(2);
  }

  if (body === '' || isAbsolutePattern(body)) {
    return (negated ? '!' : '') + body;
  }

  if (!relativeBase) {
    return (negated ? '!' : '') + body;
  }

  // path.posix.join collapses `a` + `../b` and keeps `**` segments intact.
  // Strip trailing slashes so the result matches Go's path.Join, which drops
  // them — trailing-slash patterns must not behave differently per side.
  const joined = path.posix.join(relativeBase, body).replace(/\/+$/, '');
  return (negated ? '!' : '') + joined;
}

function rebaseFilesSelector(selector: unknown, relativeBase: string): unknown {
  if (typeof selector === 'string') {
    return rebasePattern(selector, relativeBase);
  }
  if (Array.isArray(selector)) {
    return selector.map((pattern) =>
      typeof pattern === 'string'
        ? rebasePattern(pattern, relativeBase)
        : pattern,
    );
  }
  return selector;
}

function rebaseProjectValue(project: unknown, relativeBase: string): unknown {
  if (typeof project === 'string') {
    return rebasePattern(project, relativeBase);
  }
  if (Array.isArray(project)) {
    return project.map((item) =>
      typeof item === 'string' ? rebasePattern(item, relativeBase) : item,
    );
  }
  return undefined;
}

/**
 * Desugar one config entry's `basePath` into ordinary relative patterns and
 * drop the `basePath` field. Entries without `basePath` are returned as-is
 * (still without a residual `basePath` key).
 *
 * When `basePath` is set and `files` is omitted but the entry still carries
 * rules / plugins / languageOptions / settings, a catch-all
 * `files: ["<base>/**"]` is injected so the entry is scoped to that subtree —
 * matching ESLint's "bare basePath applies under that directory" intent.
 */
export function applyBasePathToEntry(
  entry: Record<string, unknown>,
  configDirectory: string,
): Record<string, unknown> {
  const hasBasePath = Object.prototype.hasOwnProperty.call(entry, 'basePath');
  if (!hasBasePath) {
    return entry;
  }

  const rawBasePath = entry.basePath;
  if (typeof rawBasePath !== 'string') {
    // Caller validates; defensive no-op strip.
    const { basePath: _dropped, ...rest } = entry;
    return rest;
  }

  const relativeBase = resolveRelativeBasePath(rawBasePath, configDirectory);
  const { basePath: _dropped, ...rest } = entry;
  const next: Record<string, unknown> = { ...rest };

  if (Array.isArray(next.files)) {
    next.files = next.files.map((selector) =>
      rebaseFilesSelector(selector, relativeBase),
    );
  }

  if (Array.isArray(next.ignores)) {
    next.ignores = next.ignores.map((pattern) =>
      typeof pattern === 'string'
        ? rebasePattern(pattern, relativeBase)
        : pattern,
    );
  }

  if (isRecord(next.languageOptions)) {
    const languageOptions = { ...next.languageOptions };
    if (isRecord(languageOptions.parserOptions)) {
      const parserOptions = { ...languageOptions.parserOptions };
      if (Object.prototype.hasOwnProperty.call(parserOptions, 'project')) {
        const rebased = rebaseProjectValue(parserOptions.project, relativeBase);
        if (rebased !== undefined) {
          parserOptions.project = rebased;
        }
      }
      languageOptions.parserOptions = parserOptions;
    }
    next.languageOptions = languageOptions;
  }

  const hasFiles = Object.prototype.hasOwnProperty.call(next, 'files');
  const scopesToSubtree =
    next.rules !== undefined ||
    next.plugins !== undefined ||
    next.languageOptions !== undefined ||
    next.settings !== undefined;

  // Scoped entry with basePath but no files → limit to the base subtree.
  if (!hasFiles && scopesToSubtree) {
    const catchAll = relativeBase ? path.posix.join(relativeBase, '**') : '**';
    next.files = [catchAll];
  }

  return next;
}

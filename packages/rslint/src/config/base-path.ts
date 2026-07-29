import path from 'node:path';

/**
 * Options for desugaring per-entry `basePath` into ordinary relative
 * `files` / `ignores` / `parserOptions.project` patterns.
 *
 * `configDirectory` is the match root the Go engine already uses:
 * config-file directory for auto-discovered configs, cwd for `--config`
 * / explicit override configs. Relative `basePath` values resolve against it.
 */
export interface ApplyBasePathOptions {
  configDirectory?: string;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

/** True for absolute POSIX or Windows paths (including drive-letter forms). */
export function isAbsolutePattern(pattern: string): boolean {
  if (path.posix.isAbsolute(pattern) || path.win32.isAbsolute(pattern)) {
    return true;
  }
  // path.win32.isAbsolute requires a platform-aware check; also catch
  // forward-slash drive forms that appear in serialized configs.
  return /^[a-zA-Z]:[\\/]/.test(pattern);
}

/**
 * Resolve an entry `basePath` to a POSIX path relative to `configDirectory`.
 * Absolute base paths are rebased via `path.relative`. Empty / "." means the
 * config root itself (no prefix).
 */
export function resolveRelativeBasePath(
  basePath: string,
  configDirectory: string,
): string {
  const root = configDirectory || process.cwd();
  const absoluteBase = path.resolve(root, basePath);
  let relative = path.relative(root, absoluteBase);
  relative = relative.split(path.sep).join('/');
  relative = path.posix.normalize(relative);
  if (relative === '.' || relative === '') {
    return '';
  }
  return relative.replace(/^\.\//, '');
}

/**
 * Rebase one glob or path pattern under `relativeBase`.
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
  const joined = path.posix.join(relativeBase, body);
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

/**
 * Apply {@link applyBasePathToEntry} across a whole flat-config array.
 */
export function applyBasePathToConfig(
  entries: Record<string, unknown>[],
  configDirectory: string = process.cwd(),
): Record<string, unknown>[] {
  return entries.map((entry) => applyBasePathToEntry(entry, configDirectory));
}

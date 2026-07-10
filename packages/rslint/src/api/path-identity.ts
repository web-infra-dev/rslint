import path from 'node:path';

type PathOperations = Pick<
  typeof path,
  'dirname' | 'isAbsolute' | 'normalize' | 'parse' | 'relative' | 'sep'
>;

export interface PathIdentity {
  readonly paths: PathOperations;
  key(filePath: string): string;
  normalize(filePath: string): string;
  equals(left: string, right: string): boolean;
  isSameOrChild(parent: string, child: string): boolean;
  compare(left: string, right: string): number;
}

/**
 * Create lexical filesystem-path identity for a platform path implementation.
 * The caller supplies the host filesystem's case behavior independently from
 * its separator/path flavor (notably, macOS uses POSIX paths but folds case).
 */
export function createPathIdentity(
  paths: PathOperations,
  caseSensitive: boolean,
): PathIdentity {
  const normalize = (filePath: string): string => {
    let normalized = paths.normalize(filePath);
    const root = paths.parse(normalized).root;
    while (normalized.length > root.length && normalized.endsWith(paths.sep)) {
      normalized = normalized.slice(0, -1);
    }
    return normalized;
  };

  const key = (filePath: string): string => {
    const normalized = normalize(filePath);
    return caseSensitive ? normalized : normalized.toLowerCase();
  };

  return {
    paths,
    key,
    normalize,
    equals: (left, right) => key(left) === key(right),
    isSameOrChild: (parent, child) => {
      if (key(parent) === key(child)) return true;
      const relative = paths.relative(key(parent), key(child));
      return (
        relative !== '' &&
        relative !== '..' &&
        !relative.startsWith(`..${paths.sep}`) &&
        !paths.isAbsolute(relative)
      );
    },
    compare: (left, right) =>
      key(left).localeCompare(key(right)) ||
      normalize(left).localeCompare(normalize(right)),
  };
}

export const nativePathIdentity = createPathIdentity(
  path,
  process.platform !== 'win32' && process.platform !== 'darwin',
);

/** Cache an upward lookup, including misses, for every traversed directory. */
export function createCachedAncestorFinder<T>(
  findInDirectory: (directory: string) => T | undefined,
  identity: PathIdentity = nativePathIdentity,
): (startDirectory: string) => T | undefined {
  const cache = new Map<string, T | undefined>();

  return (startDirectory: string): T | undefined => {
    const visited: string[] = [];
    let directory = identity.normalize(startDirectory);
    let found: T | undefined;

    while (true) {
      const key = identity.key(directory);
      if (cache.has(key)) {
        found = cache.get(key);
        break;
      }
      visited.push(key);

      found = findInDirectory(directory);
      if (found !== undefined) break;

      const parent = identity.paths.dirname(directory);
      if (identity.equals(parent, directory)) break;
      directory = parent;
    }

    for (const key of visited) cache.set(key, found);
    return found;
  };
}

/**
 * Nearest-ancestor lookup over an immutable directory set. Misses and hits are
 * cached for every traversed directory, so repeated files in one tree do only
 * lexical path work and never trigger filesystem probes.
 */
export class AncestorPathIndex<T> {
  readonly #identity: PathIdentity;
  readonly #entries = new Map<string, T>();
  readonly #cache = new Map<string, T | undefined>();

  constructor(
    entries: Iterable<readonly [directory: string, value: T]>,
    identity: PathIdentity = nativePathIdentity,
  ) {
    this.#identity = identity;
    for (const [directory, value] of entries) {
      const key = identity.key(directory);
      if (!this.#entries.has(key)) this.#entries.set(key, value);
    }
  }

  find(startDirectory: string): T | undefined {
    const startKey = this.#identity.key(startDirectory);
    if (this.#cache.has(startKey)) return this.#cache.get(startKey);

    const visited: string[] = [];
    let directory = this.#identity.normalize(startDirectory);
    let found: T | undefined;

    while (true) {
      const key = this.#identity.key(directory);
      visited.push(key);

      if (this.#entries.has(key)) {
        found = this.#entries.get(key);
        break;
      }
      if (this.#cache.has(key)) {
        found = this.#cache.get(key);
        break;
      }

      const parent = this.#identity.paths.dirname(directory);
      if (this.#identity.equals(parent, directory)) break;
      directory = parent;
    }

    for (const key of visited) this.#cache.set(key, found);
    return found;
  }

  findParent(directory: string): T | undefined {
    const normalized = this.#identity.normalize(directory);
    const parent = this.#identity.paths.dirname(normalized);
    if (this.#identity.equals(parent, normalized)) return undefined;
    return this.find(parent);
  }
}

import fs from 'node:fs';
import path from 'node:path';

type PathOperations = Pick<
  typeof path,
  'dirname' | 'normalize' | 'parse' | 'sep'
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
 * Physical equivalence is deliberately outside this helper and must be based on
 * a filesystem-resolved path.
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
      const parentKey = key(parent);
      const childKey = key(child);
      if (parentKey === childKey) return true;
      const prefix = parentKey.endsWith(paths.sep)
        ? parentKey
        : `${parentKey}${paths.sep}`;
      return childKey.startsWith(prefix);
    },
    compare: (left, right) =>
      key(left).localeCompare(key(right)) ||
      normalize(left).localeCompare(normalize(right)),
  };
}

export const nativePathIdentity = createPathIdentity(path, true);

export interface ResolvedFilesystemPath {
  lexicalPath: string;
  lexicalKey: string;
  canonicalPath: string;
  canonicalKey: string;
}

type ResolvedPathState = ResolvedFilesystemPath & {
  physicallyResolved: boolean;
};

async function realpathNative(filePath: string): Promise<string> {
  const resolvedPath = await new Promise<string>((resolve, reject) => {
    fs.realpath.native(filePath, (error, resolvedPath) => {
      if (error) reject(error);
      else resolve(resolvedPath);
    });
  });
  return resolvedPath;
}

/**
 * Resolves physical path identity for one lint operation. The memo is scoped to
 * this instance and callers should discard it when the operation completes.
 */
export class RunPathResolver {
  readonly #resolved = new Map<string, Promise<ResolvedPathState>>();

  async resolve(filePath: string): Promise<ResolvedFilesystemPath> {
    const resolved = await this.#resolveState(filePath);
    return resolved;
  }

  async #resolveState(filePath: string): Promise<ResolvedPathState> {
    const lexicalPath = nativePathIdentity.normalize(filePath);
    const lexicalKey = nativePathIdentity.key(lexicalPath);
    let pending = this.#resolved.get(lexicalKey);
    if (!pending) {
      pending = (async () => {
        let canonicalPath = lexicalPath;
        let physicallyResolved = false;
        try {
          canonicalPath = nativePathIdentity.normalize(
            await realpathNative(lexicalPath),
          );
          physicallyResolved = true;
        } catch {
          // Virtual, missing, or inaccessible paths retain exact lexical identity.
        }
        return {
          lexicalPath,
          lexicalKey,
          canonicalPath,
          canonicalKey: nativePathIdentity.key(canonicalPath),
          physicallyResolved,
        };
      })();
      this.#resolved.set(lexicalKey, pending);
    }
    const resolved = await pending;
    return resolved;
  }

  /**
   * Resolve a virtual or missing path through its nearest existing ancestor.
   * The unresolved suffix is appended to that ancestor's realpath, preserving
   * directory-symlink identity without requiring the target itself on disk.
   */
  async resolveWithAncestorFallback(
    filePath: string,
  ): Promise<ResolvedFilesystemPath> {
    const exact = await this.#resolveState(filePath);
    if (exact.physicallyResolved) return exact;

    const suffix: string[] = [];
    let current = exact.lexicalPath;
    while (true) {
      const parent = path.dirname(current);
      if (nativePathIdentity.equals(parent, current)) return exact;
      suffix.unshift(path.basename(current));

      const resolvedParent = await this.#resolveState(parent);
      if (resolvedParent.physicallyResolved) {
        const canonicalPath = nativePathIdentity.normalize(
          path.resolve(resolvedParent.canonicalPath, ...suffix),
        );
        return {
          lexicalPath: exact.lexicalPath,
          lexicalKey: exact.lexicalKey,
          canonicalPath,
          canonicalKey: nativePathIdentity.key(canonicalPath),
        };
      }
      current = parent;
    }
  }

  async resolveAll(
    filePaths: readonly string[],
    concurrency = 32,
  ): Promise<ResolvedFilesystemPath[]> {
    const results = new Array<ResolvedFilesystemPath>(filePaths.length);
    let next = 0;
    const worker = async (): Promise<void> => {
      while (true) {
        const index = next++;
        if (index >= filePaths.length) return;
        results[index] = await this.resolve(filePaths[index]);
      }
    };
    const workerCount = Math.min(
      filePaths.length,
      Math.max(1, Math.floor(concurrency)),
    );
    await Promise.all(Array.from({ length: workerCount }, worker));
    return results;
  }
}

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

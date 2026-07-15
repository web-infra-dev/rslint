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

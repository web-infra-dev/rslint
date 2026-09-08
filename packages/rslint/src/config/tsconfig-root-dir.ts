import path from 'node:path';
import { fileURLToPath } from 'node:url';
import type { RslintConfigEntry } from './define-config.js';

const originDescription = 'rslint.tsconfigRootDir';

// Like typescript-eslint's preset getters, infer from the nearest config frame.
// Keep the origin on the returned data: unrelated config loads and API instances
// must not share a mutable candidate set. Distinct symbols survive rules spread
// from multiple presets without overwriting one another.
export function withTSConfigRootDir<
  T extends RslintConfigEntry | RslintConfigEntry[],
>(preset: T): T {
  const directory = configDirectoryFromStack();
  if (directory === undefined) return preset;
  const origin = Symbol(originDescription);
  const annotate = (entry: RslintConfigEntry): RslintConfigEntry => {
    const annotated = {
      ...entry,
      [origin]: directory,
      ...(entry.rules
        ? { rules: { ...entry.rules, [origin]: directory } }
        : {}),
    };
    return annotated;
  };
  return (Array.isArray(preset) ? preset.map(annotate) : annotate(preset)) as T;
}

/** Read only our provenance, without inspecting arbitrary user object graphs. */
export function collectTSConfigRootDirs(
  entry: Record<string, unknown>,
): string[] {
  const directories = new Set<string>();
  for (const source of [entry, entry.rules]) {
    if (source === null || typeof source !== 'object') continue;
    for (const key of Object.getOwnPropertySymbols(source)) {
      if (key.description !== originDescription) continue;
      const value: unknown = Reflect.get(source, key);
      if (typeof value === 'string') directories.add(value);
    }
  }
  return [...directories];
}

function configDirectoryFromStack(): string | undefined {
  const previousLimit = Error.stackTraceLimit;
  const previousPrepare = Error.prepareStackTrace;
  let stack: NodeJS.CallSite[];
  try {
    Error.stackTraceLimit = Infinity;
    Error.prepareStackTrace = (_error, sites) => sites;
    const captured: { stack?: unknown } = {};
    Error.captureStackTrace(captured, configDirectoryFromStack);
    stack = captured.stack as NodeJS.CallSite[];
  } finally {
    Error.stackTraceLimit = previousLimit;
    Error.prepareStackTrace = previousPrepare;
  }
  for (const frame of stack) {
    const name = frame.getFileName();
    if (!name) continue;
    const filename = name.startsWith('file://') ? fileURLToPath(name) : name;
    const parsed = path.parse(filename);
    if (/^rslint\.config\.(c|m)?(j|t)s$/.test(parsed.base)) {
      return process.platform === 'win32'
        ? parsed.dir.replaceAll('/', path.sep)
        : parsed.dir;
    }
  }
  return undefined;
}

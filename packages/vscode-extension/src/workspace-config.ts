/**
 * Pure helper for the LSP-side parent-ignores filter, lifted out of
 * `Rslint.ts::loadAndSendConfig` so it can be unit-tested without a
 * live VS Code runtime.
 *
 * Background: `cli.ts:253` runs `filterConfigsByParentIgnores(loaded)`
 * after discovering JS configs so a root config's global `ignores`
 * (e.g. `{ ignores: ['fixtures/**'] }`) ALSO drops nested configs
 * under those ignored directories — matches ESLint v10 behavior of
 * not traversing into globally-ignored dirs. Without this, the LSP
 * would happily register a nested config under an ignored tree and
 * `getConfigForURI(ignored_file_uri)` would route the file through
 * the nested config, producing diagnostics the CLI never emits.
 *
 * The shared helper from `@rslint/core/config-loader` takes
 * `{ configDirectory: <fs path>, configPath, entries }`. The LSP
 * carries `configDirectoryUri` (URI form, what Go's `jsConfigs` map
 * keys off) alongside the filesystem path, so this helper projects
 * the LSP record into the shared shape, runs the filter, and projects
 * surviving records back.
 */
import {
  filterConfigsByParentIgnores,
  type DiscoveredConfigEntry,
} from '@rslint/core/config-loader';

export interface LSPLoadedConfig {
  /** Absolute filesystem path of the config directory.
   *  Used by `filterConfigsByParentIgnores` (it does `fs.realpathSync`
   *  + ancestor `startsWith` checks against this). */
  configDirectoryFsPath: string;
  /** URI form of the same directory. Go's `jsConfigs` map keys off
   *  this, so the wire payload retains the URI form. */
  configDirectoryUri: string;
  /** Absolute path of the config file itself. Used as a stable
   *  identity key when intersecting filter input/output. */
  configPath: string;
  /** Normalized config entries (whatever `normalizeConfig` returned). */
  entries: DiscoveredConfigEntry['entries'];
}

/**
 * Drop nested LSP-loaded configs whose directory is covered by an
 * ancestor LSP-loaded config's global `ignores`. Pure: input array is
 * not mutated; output preserves the original input order of surviving
 * entries.
 *
 * The intersection uses `configPath` (unique per config file) as the
 * identity key. `configPath` is always set on LSP records because
 * VS Code's `workspace.findFiles` returns concrete file URIs.
 */
export function applyParentIgnoresFilter<T extends LSPLoadedConfig>(
  loaded: readonly T[],
): T[] {
  if (loaded.length <= 1) return [...loaded];
  const filterInputs: DiscoveredConfigEntry[] = loaded.map((e) => ({
    configDirectory: e.configDirectoryFsPath,
    configPath: e.configPath,
    entries: e.entries,
  }));
  const kept = new Set(
    filterConfigsByParentIgnores(filterInputs)
      .map((e) => e.configPath)
      .filter((p): p is string => p != null),
  );
  return loaded.filter((e) => kept.has(e.configPath));
}

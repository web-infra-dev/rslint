/* rslint-disable @typescript-eslint/no-unsafe-type-assertion */
/**
 * Plugin loader for the runner Worker.
 *
 * Each Worker calls {@link loadPluginsFromConfigs} once at startup to
 * import every rslint config file assigned to it. The config files
 * themselves carry live plugin instances on their object-form `plugins`
 * maps; the worker pulls them out, sorts them
 * by prefix, and caches the resulting `LoadedPlugins` per config
 * directory for the worker's entire lifetime. Subsequent lint tasks
 * select the right `LoadedPlugins` via `configKey` (= config directory).
 *
 * CJS / ESM unwrap. ESLint plugins ship in both module systems and the
 * `plugins.<prefix>` value the user supplied can be either:
 *
 *   - the plugin object directly (the conventional shape), or
 *   - a `{ default: pluginObj }` wrapper (when the user wrote
 *     `import unicorn from 'eslint-plugin-unicorn'` in an ESM config
 *     against a CJS plugin and the bundler injected the default).
 *
 * We pass through {@link unwrapPluginModule} so both shapes work.
 *
 * Node version requirement: ≥ 20, matching the package's
 * `engines.node` field. Enforced at install time by the package
 * manager and re-checked at startup by {@link ensureNodeVersion} so a
 * direct-binary invocation against a too-old Node fails fast with a
 * clear message instead of a cryptic syntax/feature error later.
 */

import path from 'node:path';
import { pathToFileURL } from 'node:url';

import type { ConfigDescriptor } from '../types.js';
import {
  selectPluginSource,
  unwrapPluginModule,
} from '../../config/plugin-source.js';

/**
 * Extension-aware config loader for worker init. Mirrors the strategy
 * `packages/rslint/src/config/config-file-loader.ts::loadConfigFile` uses on the
 * main thread so a `.ts`/`.mts`/`.cts` config file loads identically in both
 * places.
 *
 * The previous implementation directly called `await import(url)` for
 * every extension; on Node 20 (declared support floor) `.ts` configs
 * have no native loader and a workspace that uses
 * `rslint.config.ts` + object-form `plugins` would fail worker init even
 * though the main thread had already loaded the same file via jiti.
 *
 * Resolution order for `.ts`/`.mts`/`.cts`:
 *
 *   1. `process.features.typescript` (Node ≥ 22.6 with native TS) →
 *      native `import()`.
 *   2. Otherwise `jiti` (declared as optional peer in package.json).
 *   3. If jiti is unavailable, throw an actionable error pointing the
 *      user at either upgrading Node or `npm install -D jiti`.
 *
 * Cache-busting `?t=Date.now()`: NOT used here. Worker init imports
 * each config once per worker lifetime; long-running LSP sessions
 * rebuild the worker pool via PluginLintPool when configs change
 * (host-side mtime + size fingerprint), so the host already controls
 * staleness. Re-importing with `?t=` would mean every reconfigure
 * pays a fresh module-graph build (≈300 ms for unicorn /
 * typescript-eslint per worker) instead of relying on Node's ESM
 * module cache for unchanged dependencies.
 */
async function importConfigFile(configFilePath: string): Promise<unknown> {
  const ext = path.extname(configFilePath);
  if (ext === '.ts' || ext === '.mts' || ext === '.cts') {
    const useNative = Boolean(
      (process.features as { typescript?: boolean }).typescript,
    );
    if (useNative) {
      return import(pathToFileURL(configFilePath).href);
    }
    let jiti;
    try {
      const { createJiti } = (await import('jiti')) as {
        createJiti: (
          base: string,
          opts?: { interopDefault?: boolean },
        ) => { import: (p: string) => Promise<unknown> };
      };
      jiti = createJiti(path.dirname(configFilePath), {
        interopDefault: true,
      });
    } catch {
      throw new Error(
        `Failed to load TypeScript config file in worker: ${configFilePath}\n` +
          `To load .ts/.mts/.cts config files, either:\n` +
          `  1. Use Node.js >= 22.6 (with native TypeScript support), or\n` +
          `  2. Install jiti as a dev dependency: npm install -D jiti\n` +
          `(The main thread successfully loaded this file, so the worker ` +
          `must use the same fallback to stay consistent.)`,
      );
    }
    const resolved = await jiti.import(configFilePath);
    // jiti returns the module namespace — wrap so the rest of
    // `loadPluginsFromConfigFile` can keep its
    // `(configMod as { default?: unknown }).default ?? configMod`
    // unwrap.
    return resolved;
  }
  return import(pathToFileURL(configFilePath).href);
}

/**
 * Minimum Node major. Mirrors `engines.node` in package.json — keep the
 * two in sync. Bump only when the codebase actually needs a newer API.
 */
const MIN_NODE_MAJOR = 20;

/**
 * Loaded plugin shape: only the fields the runner consumes. Plugins may
 * carry far more (configs, processors, etc.) — we keep references intact
 * for downstream needs but type only what we use.
 */
export interface LoadedPlugin {
  prefix: string;
  /** The unwrapped plugin module, ready for `plugin.rules['ruleName']`. */
  plugin: {
    meta?: { name?: string; version?: string };
    name?: string;
    rules?: Record<string, unknown>;
    configs?: Record<string, unknown>;
    [key: string]: unknown;
  };
}

/**
 * Loaded plugin set for a single rslint config. `rules` is keyed by
 * `<prefix>/<ruleName>` and is the only lookup `lintFile` consults
 * — the worker has already picked the right `LoadedPlugins` for this
 * file via its `configKey` map before calling `lintFile`, so there's
 * no cross-config prefix collision to worry about.
 */
export interface LoadedPlugins {
  plugins: LoadedPlugin[];
  rules: Map<string, unknown>;
}

/**
 * Errors thrown by the loader carry enough context (the config that
 * declared the failing plugin) for the runtime to surface actionable
 * messages to users.
 */
export class PluginLoaderError extends Error {
  constructor(
    public readonly configPath: string,
    message: string,
  ) {
    super(message);
    this.name = 'PluginLoaderError';
  }
}

/**
 * Load every plugin in `entries` and return them keyed by prefix, with a
 * Import the user's rslint config file directly and extract plugin
 * instances from its object-form `plugins` map(s). Each worker calls
 * this through {@link loadPluginsFromConfigs}
 * once per assigned config at init.
 *
 * Each worker independently imports the config (and transitively its
 * plugins), naturally anchoring `node_modules` walks at the config's
 * own location — which is what makes monorepo setups Just Work: a
 * sub-package config gets its sub-package's node_modules, not the
 * root's.
 *
 * @throws PluginLoaderError on config-import failure or a Node version
 *   too old to satisfy {@link MIN_NODE_MAJOR}.
 */
export async function loadPluginsFromConfigFile(
  configFilePath: string,
): Promise<LoadedPlugins> {
  ensureNodeVersion();

  const plugins: LoadedPlugin[] = [];
  const rules = new Map<string, unknown>();

  let configMod: unknown;
  try {
    configMod = await importConfigFile(configFilePath);
  } catch (err) {
    throw new PluginLoaderError(
      configFilePath,
      `failed to import config file ${configFilePath}: ${
        (err as Error)?.message ?? String(err)
      }`,
    );
  }

  // Config file's default export is `RslintConfigEntry[]`; each entry may
  // carry `plugins: { <prefix>: pluginObj }`.
  const exportedDefault =
    (configMod as { default?: unknown }).default ?? configMod;
  const configArray = Array.isArray(exportedDefault) ? exportedDefault : [];

  // Track which prefixes we've already loaded: a re-declared prefix with the
  // SAME plugin instance is deduped (collapse to one LoadedPlugin), while a
  // DIFFERENT instance throws below. This worker-init check — NOT the host-side
  // collectPluginMeta dedupe (which silently first-wins) — is where a genuine
  // cross-entry prefix conflict is actually caught.
  const seenPrefix = new Set<string>();

  // Per entry, the live community plugins come from the object-form
  // `plugins` map (`plugins: { uc: unicornPlugin }`), selected via the
  // shared `selectPluginSource`. This mirrors `normalizeConfig` in
  // packages/rslint/src/config/config-file-loader.ts — both paths must extract plugins
  // from the same source, or the worker holds a plugin set that doesn't
  // match what the main thread believed was active. Array-form `plugins`
  // (the native-name whitelist) carries no live objects and yields no
  // source here.
  for (const entry of configArray) {
    const source = selectPluginSource(entry);
    if (source == null) continue;
    for (const prefix of Object.keys(source)) {
      const pluginObj = source[prefix];
      // `unwrapPluginModule` ALREADY prefers `.default` when present
      // and falls back to the value itself, so we hand it the raw
      // `pluginObj`. A previous (incorrect) call site wrapped this
      // as `{ default: pluginObj }` first — that artificial wrap let
      // the function only ever peel the synthetic outer layer, so a
      // genuine `{ default: realPlugin }` shape (the one produced by
      // `import * as p from 'pkg'` for a pure-ESM plugin package, or
      // by some bundler interop output) was returned unchanged with
      // its `.rules` still nested under `.default`, and every rule
      // for that prefix silently dropped from the dispatch map.
      const plugin = unwrapPluginModule<LoadedPlugin['plugin']>(pluginObj);
      if (plugin == null || typeof plugin !== 'object') continue;

      // Cross-entry redefinition check, aligned with ESLint v10
      // (`lib/config/flat-config-schema.js::pluginsSchema.merge`).
      // The same prefix re-declared with the SAME plugin instance is
      // a harmless dedupe (the second declaration adds no info).
      // The same prefix re-declared with a DIFFERENT instance is a
      // configuration error: rslint can only load ONE plugin object
      // per prefix per worker, and silently picking either side
      // would surprise the author who edited the config. Throw with
      // ESLint's exact error message so users familiar with that
      // diagnostic find rslint behaves the same way.
      if (seenPrefix.has(prefix)) {
        const existing = plugins.find((p) => p.prefix === prefix);
        if (existing && existing.plugin !== plugin) {
          throw new PluginLoaderError(
            configFilePath,
            `Cannot redefine plugin "${prefix}".`,
          );
        }
        // Same instance — skip without throw (dedupe).
        continue;
      }
      seenPrefix.add(prefix);

      // unwrapPluginModule + the `plugin == null || typeof !== 'object'`
      // continue above already narrow `plugin` to `LoadedPlugin['plugin']`.
      const loaded: LoadedPlugin = {
        prefix,
        plugin,
      };
      plugins.push(loaded);

      if (plugin.rules) {
        for (const [ruleName, ruleDef] of Object.entries(plugin.rules)) {
          rules.set(`${prefix}/${ruleName}`, ruleDef);
        }
      }
    }
  }

  return { plugins, rules };
}

/**
 * Import every config in `configs` and return a map keyed by each
 * config's directory. This is the worker-side entry point for the new
 * config-loads-in-worker flow: each worker calls this once at init,
 * caches the result for its entire lifetime, and per-file lint tasks
 * pick the right `LoadedPlugins` via `request.configKey === configDirectory`.
 *
 * Fail-fast on the first config import failure — same contract as the
 * old `loadPlugins(entries, baseUrl)` path. Surfacing a partial success
 * silently degrades lint quality across the workspace (files under the
 * failing config get no plugin rules), so we prefer a clean, loud
 * failure that the user can fix and retry.
 */
export async function loadPluginsFromConfigs(
  configs: readonly ConfigDescriptor[],
): Promise<Map<string, LoadedPlugins>> {
  ensureNodeVersion();
  // Parallelize the per-config dynamic imports. Each
  // `loadPluginsFromConfigFile` is an independent `await import(...)`
  // chain (resolving the config's plugin npm modules through Node's
  // ESM loader); they share no mutable state. Serializing them — as
  // the previous `for (...) await` loop did — meant a workspace with
  // 5 nested configs each pulling in unicorn/typescript-eslint
  // synchronously waited ~5×300 ms ≈ 1.5 s on worker init. The LSP
  // first-lint after VS Code reload blocks on this. Promise.all keeps
  // the same overall failure semantics (any rejection still surfaces)
  // — `loadPluginsFromConfigFile` itself throws on `import()` failure
  // and the caller catches once. Ordering of the returned Map is by
  // insertion of `out.set(...)` calls; since we resolve all promises
  // first then iterate `configs` in original order to populate the
  // Map, the map's iteration order matches the input order exactly.
  const loadedList = await Promise.all(
    configs.map(async (cfg) => loadPluginsFromConfigFile(cfg.configPath)),
  );
  const out = new Map<string, LoadedPlugins>();
  for (let i = 0; i < configs.length; i++) {
    const dir = configs[i].configDirectory;
    const existing = out.get(dir);
    // Duplicate descriptors are harmless when they resolve to the same plugin
    // instances. A different plugin instance under the same prefix and routing
    // key is invalid even when the two instances expose disjoint rule names.
    out.set(
      dir,
      existing
        ? mergeLoadedPlugins(
            existing,
            loadedList[i],
            dir,
            configs[i].configPath,
          )
        : loadedList[i],
    );
  }
  return out;
}

/**
 * Merge two `LoadedPlugins` for the SAME `configDirectory`. Plugin prefixes
 * are routing identities: the same plugin instance is harmless deduplication,
 * while a different instance under the same prefix is always rejected.
 */
function mergeLoadedPlugins(
  a: LoadedPlugins,
  b: LoadedPlugins,
  dir: string,
  configPath: string,
): LoadedPlugins {
  const plugins = [...a.plugins];
  const pluginsByPrefix = new Map(
    plugins.map((loaded) => [loaded.prefix, loaded] as const),
  );
  for (const loaded of b.plugins) {
    const existing = pluginsByPrefix.get(loaded.prefix);
    if (existing) {
      if (existing.plugin !== loaded.plugin) {
        throw new PluginLoaderError(
          configPath,
          `Cannot redefine plugin "${loaded.prefix}" for routing key "${dir}".`,
        );
      }
      continue;
    }
    pluginsByPrefix.set(loaded.prefix, loaded);
    plugins.push(loaded);
  }

  const rules = new Map(a.rules);
  for (const [key, rule] of b.rules) {
    const existing = rules.get(key);
    if (existing !== undefined && existing !== rule) {
      throw new PluginLoaderError(
        configPath,
        `[rslint] plugin rule "${key}" is defined by different plugin ` +
          `instances for the same config key (${dir})`,
      );
    }
    rules.set(key, rule);
  }
  return { plugins, rules };
}

function ensureNodeVersion(): void {
  const v = process.versions.node;
  const major = parseInt(v.split('.')[0], 10);
  if (Number.isNaN(major) || major < MIN_NODE_MAJOR) {
    throw new Error(
      `[plugin-loader] Node ≥ ${MIN_NODE_MAJOR} required, current is v${v}. ` +
        `See engines.node in this package's package.json.`,
    );
  }
}

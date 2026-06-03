import path from 'node:path';
import { pathToFileURL } from 'node:url';
import { NATIVE_PLUGIN_PREFIXES } from './define-config.js';
import { selectPluginSource, unwrapPluginModule } from './plugin-source.js';

/**
 * Load a JS/TS config file.
 * - .js/.mjs: native import()
 * - .ts/.mts: native import() when Node.js has TypeScript support (>= 22.6),
 *             otherwise fall back to jiti
 */
export async function loadConfigFile(configPath: string): Promise<unknown> {
  const ext = path.extname(configPath);

  if (ext === '.js' || ext === '.mjs') {
    const mod: Record<string, unknown> = await import(
      pathToFileURL(configPath).href
    );
    return mod.default ?? mod;
  }

  if (ext === '.ts' || ext === '.mts') {
    // Use feature detection to decide the loading strategy (same as rsbuild).
    // process.features.typescript is available in Node.js >= 22.6.
    const useNative = Boolean(process.features.typescript);

    if (useNative) {
      const mod: Record<string, unknown> = await import(
        pathToFileURL(configPath).href
      );
      return mod.default ?? mod;
    }

    const jiti = await loadJiti(configPath);
    if (jiti) {
      const resolved = await jiti.import(configPath);
      return extractDefault(resolved);
    }

    throw new Error(
      `Failed to load TypeScript config file: ${configPath}\n` +
        `To load .ts config files, either:\n` +
        `  1. Use Node.js >= 22.6 (with native TypeScript support)\n` +
        `  2. Install jiti as a dependency: npm install -D jiti`,
    );
  }

  throw new Error(`Unsupported config file extension: ${ext}`);
}

/**
 * Try to load jiti (optional peer dependency).
 */
async function loadJiti(configPath: string): Promise<{
  import: (path: string) => Promise<unknown>;
} | null> {
  try {
    const { createJiti } = await import('jiti');
    return createJiti(path.dirname(configPath), { interopDefault: true });
  } catch {
    return null;
  }
}

function extractDefault(mod: unknown): unknown {
  if (typeof mod === 'object' && mod !== null && 'default' in mod) {
    return mod.default;
  }
  return mod;
}

/**
 * Validate and strip non-serializable fields from the config.
 */
export function normalizeConfig(config: unknown): Record<string, unknown>[] {
  if (!Array.isArray(config)) {
    throw new Error(
      `rslint config must export an array (flat config format), got ${typeof config}`,
    );
  }

  return config
    .filter((entry: unknown, index: number) => {
      if (entry == null || typeof entry !== 'object') {
        console.warn(
          `[rslint] Config entry at index ${index} is not an object (got ${entry === null ? 'null' : typeof entry}), skipping.`,
        );
        return false;
      }
      return true;
    })
    .map((entry: Record<string, unknown>, index: number) => {
      if (entry.files != null && !Array.isArray(entry.files)) {
        throw new Error(
          `[rslint] Config entry at index ${index}: "files" must be an array, got ${typeof entry.files}`,
        );
      }
      if (entry.ignores != null && !Array.isArray(entry.ignores)) {
        throw new Error(
          `[rslint] Config entry at index ${index}: "ignores" must be an array, got ${typeof entry.ignores}`,
        );
      }

      // Extract ESLint-plugin metadata. Live plugin objects carry
      // functions and must NEVER reach the serializable payload sent to
      // Go — only {prefix, ruleNames} crosses the wire. The worker
      // re-imports this config file to obtain the live instances. Source
      // selection (eslintPlugins over object-form plugins) is shared with
      // the worker loader via `selectPluginSource` so the two never drift.
      const pluginSource = selectPluginSource(entry);
      const eslintPluginMeta: Record<string, { ruleNames: string[] }> = {};
      const pluginPrefixes: string[] = [];
      if (pluginSource != null) {
        for (const prefix of Object.keys(pluginSource)) {
          if (NATIVE_PLUGIN_PREFIXES.has(prefix)) {
            throw new Error(
              `[rslint] Config entry at index ${index}: eslintPlugins prefix "${prefix}" collides with the built-in plugin of the same name; choose a different prefix.`,
            );
          }
          const plugin = unwrapPluginModule(pluginSource[prefix]);
          const pluginRules = plugin?.rules;
          if (pluginRules == null || typeof pluginRules !== 'object') {
            throw new Error(
              `[rslint] Config entry at index ${index}: eslintPlugins["${prefix}"] must expose a "rules" object.`,
            );
          }
          eslintPluginMeta[prefix] = {
            ruleNames: Object.keys(pluginRules).sort(),
          };
          pluginPrefixes.push(prefix);
        }
      }

      // The serializable `plugins` handed to Go is a string[] of declared
      // prefixes (its native-plugin gate keys off this set). Merge any
      // string entries the user wrote with the mounted eslintPlugins
      // prefixes; drop the live object-form `plugins` (functions can't be
      // JSON-serialized to Go).
      const stringPlugins = Array.isArray(entry.plugins)
        ? entry.plugins.filter((p): p is string => typeof p === 'string')
        : [];
      const plugins = [...new Set([...stringPlugins, ...pluginPrefixes])];

      return {
        files: entry.files,
        ignores: entry.ignores,
        languageOptions: entry.languageOptions,
        rules: entry.rules,
        plugins,
        settings: entry.settings,
        ...(pluginPrefixes.length > 0
          ? { eslintPlugins: eslintPluginMeta }
          : {}),
      };
    });
}

/** A worker-pool config descriptor: which config file to import and the
 *  directory key per-file plugin-lint tasks route on. */
export interface PluginConfigDescriptor {
  configPath: string;
  configDirectory: string;
}

/**
 * Derive, from normalized configs, the ESLint-plugin metadata the Go core
 * needs (`{prefix, ruleNames}` for placeholder rules) and the worker-pool
 * descriptors (only for configs that mount plugins — others stay
 * zero-overhead). Shared by the CLI (cli.ts) and the VS Code extension
 * (Rslint.ts) so both produce the identical `eslintPlugins` shape Go parses.
 * ruleNames for a shared prefix are merged (set union) across configs: Go's
 * placeholder registry is a per-prefix superset, while the worker routes the
 * actual rules per-config. This is NOT a validation — a genuinely conflicting
 * redefinition (same prefix, different plugin object) is caught at worker init
 * (plugin-loader throws ESLint's `Cannot redefine plugin` error), not here.
 */
export function collectPluginMeta(
  configs: ReadonlyArray<{
    configPath: string;
    configDirectory: string;
    entries: ReadonlyArray<unknown>;
  }>,
): {
  eslintPluginEntries: Array<{ prefix: string; ruleNames: string[] }>;
  pluginConfigs: PluginConfigDescriptor[];
} {
  // Entries are normalizeConfig output, so an `eslintPlugins` field — when
  // present — is already a `{ prefix: { ruleNames } }` map. A type predicate
  // narrows the `unknown` entry without an unsafe assertion.
  const isPluginMetaMap = (
    v: unknown,
  ): v is Record<string, { ruleNames: string[] }> =>
    v !== null && typeof v === 'object';
  const byPrefix = new Map<string, string[]>();
  const pluginConfigs: PluginConfigDescriptor[] = [];
  for (const c of configs) {
    let hasPlugins = false;
    for (const entry of c.entries) {
      if (
        entry === null ||
        typeof entry !== 'object' ||
        !('eslintPlugins' in entry)
      ) {
        continue;
      }
      const ep = entry.eslintPlugins;
      if (!isPluginMetaMap(ep)) continue;
      for (const [prefix, meta] of Object.entries(ep)) {
        hasPlugins = true;
        // Union ruleNames across every config that mounts this prefix. Go
        // registers ONE global placeholder set per prefix, but the worker
        // routes per-config (each config's own LoadedPlugins), so a rule unique
        // to a second config's same-prefix plugin must still be registered —
        // otherwise Go never resolves/dispatches it and it silently never runs
        // (false green). First-wins would drop it.
        const existing = byPrefix.get(prefix);
        if (existing) {
          for (const name of meta.ruleNames) {
            if (!existing.includes(name)) existing.push(name);
          }
        } else {
          byPrefix.set(prefix, [...meta.ruleNames]);
        }
      }
    }
    if (hasPlugins) {
      pluginConfigs.push({
        configPath: c.configPath,
        configDirectory: c.configDirectory,
      });
    }
  }
  const eslintPluginEntries = [...byPrefix.entries()].map(
    ([prefix, ruleNames]) => ({ prefix, ruleNames }),
  );
  return { eslintPluginEntries, pluginConfigs };
}

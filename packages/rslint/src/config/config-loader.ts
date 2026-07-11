import path from 'node:path';
import { createRequire } from 'node:module';
import { pathToFileURL } from 'node:url';
import { NATIVE_PLUGIN_RESERVED_NAMES } from './define-config.js';
import { selectPluginSource, unwrapPluginModule } from './plugin-source.js';

export {
  filterConfigsByParentIgnores,
  type ConfigEntry,
} from './config-hierarchy.js';

export const JS_CONFIG_FILES = [
  'rslint.config.js',
  'rslint.config.mjs',
  'rslint.config.cjs',
  'rslint.config.ts',
  'rslint.config.mts',
  'rslint.config.cts',
] as const;

let freshConfigLoadNonce = 0;

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

/**
 * Load a JS/TS config file.
 * - .js/.mjs/.cjs: native import()
 * - .ts/.mts/.cts: native import() when Node.js has TypeScript support (>= 22.6),
 *             otherwise fall back to jiti
 */
export async function loadConfigFile(configPath: string): Promise<unknown> {
  const ext = path.extname(configPath);

  if (ext === '.js' || ext === '.mjs' || ext === '.cjs') {
    const mod: Record<string, unknown> = await import(
      pathToFileURL(configPath).href
    );
    return mod.default ?? mod;
  }

  if (ext === '.ts' || ext === '.mts' || ext === '.cts') {
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
 * Load a config without reusing the config module from a previous reload.
 * TypeScript loading selects native stripping or jiti before evaluation so a
 * runtime exception from the config is propagated without executing it again.
 */
export async function loadConfigFileFresh(
  configPath: string,
): Promise<unknown> {
  const ext = path.extname(configPath);
  const requireFromConfig = createRequire(pathToFileURL(configPath));
  const evictCommonJSCache = (): void => {
    try {
      Reflect.deleteProperty(
        requireFromConfig.cache,
        requireFromConfig.resolve(configPath),
      );
    } catch {
      // ESM-only configs are not present in the CommonJS module cache.
    }
  };

  if (ext === '.cjs') {
    evictCommonJSCache();
    return requireFromConfig(configPath) as unknown;
  }

  if (ext === '.js' || ext === '.mjs') {
    evictCommonJSCache();
    return importConfigFresh(configPath);
  }

  if (ext === '.ts' || ext === '.mts' || ext === '.cts') {
    evictCommonJSCache();
    if (process.features.typescript) {
      return importConfigFresh(configPath);
    }
    return loadConfigFile(configPath);
  }

  throw new Error(`Unsupported config file extension: ${ext}`);
}

async function importConfigFresh(configPath: string): Promise<unknown> {
  const url = pathToFileURL(configPath);
  url.searchParams.set('rslint', String(freshConfigLoadNonce++));
  const mod: Record<string, unknown> = await import(url.href);
  return mod.default ?? mod;
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

  for (let index = 0; index < config.length; index++) {
    if (!Object.prototype.hasOwnProperty.call(config, index)) {
      throw new Error(
        `[rslint] Config entry at index ${index}: unexpected undefined config`,
      );
    }
  }

  return config.map((rawEntry: unknown, index: number) => {
    if (rawEntry === null) {
      throw new Error(
        `[rslint] Config entry at index ${index}: unexpected null config`,
      );
    }
    if (Array.isArray(rawEntry)) {
      throw new Error(
        `[rslint] Config entry at index ${index}: unexpected array`,
      );
    }
    if (!isRecord(rawEntry)) {
      throw new Error(
        `[rslint] Config entry at index ${index} must be an object, got ${typeof rawEntry}`,
      );
    }

    const entry = rawEntry;

    const hasFiles = Object.prototype.hasOwnProperty.call(entry, 'files');
    if (hasFiles && !Array.isArray(entry.files)) {
      throw new Error(
        `[rslint] Config entry at index ${index}: "files" must be an array, got ${typeof entry.files}`,
      );
    }
    if (hasFiles && Array.isArray(entry.files) && entry.files.length === 0) {
      throw new Error(
        `[rslint] Config entry at index ${index}: "files" must be a non-empty array`,
      );
    }
    if (
      hasFiles &&
      Array.isArray(entry.files) &&
      entry.files.some(
        (pattern) =>
          typeof pattern !== 'string' &&
          (!Array.isArray(pattern) ||
            pattern.some((nestedPattern) => typeof nestedPattern !== 'string')),
      )
    ) {
      throw new Error(
        `[rslint] Config entry at index ${index}: "files" must contain only strings or arrays of strings`,
      );
    }
    const hasIgnores = Object.prototype.hasOwnProperty.call(entry, 'ignores');
    if (hasIgnores && !Array.isArray(entry.ignores)) {
      throw new Error(
        `[rslint] Config entry at index ${index}: "ignores" must be an array, got ${typeof entry.ignores}`,
      );
    }
    if (
      Array.isArray(entry.ignores) &&
      entry.ignores.some((pattern) => typeof pattern !== 'string')
    ) {
      throw new Error(
        `[rslint] Config entry at index ${index}: "ignores" must contain only strings`,
      );
    }

    for (const key of ['rules', 'languageOptions', 'settings'] as const) {
      if (
        Object.prototype.hasOwnProperty.call(entry, key) &&
        (entry[key] === null ||
          typeof entry[key] !== 'object' ||
          Array.isArray(entry[key]))
      ) {
        throw new Error(
          `[rslint] Config entry at index ${index}: "${key}" must be an object`,
        );
      }
    }

    if (isRecord(entry.rules)) {
      validateRules(entry.rules, index);
    }

    const languageOptions = isRecord(entry.languageOptions)
      ? entry.languageOptions
      : undefined;
    if (
      languageOptions &&
      Object.prototype.hasOwnProperty.call(languageOptions, 'globals')
    ) {
      validateGlobals(languageOptions.globals, index);
    }

    const hasPlugins = Object.prototype.hasOwnProperty.call(entry, 'plugins');
    if (
      hasPlugins &&
      (entry.plugins === null || typeof entry.plugins !== 'object')
    ) {
      throw new Error(
        `[rslint] Config entry at index ${index}: "plugins" must be an object or an array of strings`,
      );
    }
    if (
      Array.isArray(entry.plugins) &&
      entry.plugins.some((plugin) => typeof plugin !== 'string')
    ) {
      throw new Error(
        `[rslint] Config entry at index ${index}: "plugins" must contain only strings`,
      );
    }
    if (
      Object.prototype.hasOwnProperty.call(entry, 'name') &&
      typeof entry.name !== 'string'
    ) {
      throw new Error(
        `[rslint] Config entry at index ${index}: "name" must be a string`,
      );
    }

    // Extract ESLint-plugin metadata from the object-form `plugins`. Live
    // plugin objects carry functions and must NEVER reach the serializable
    // payload sent to Go — only {prefix, ruleNames} crosses the wire. The
    // worker re-imports this config file to obtain the live instances.
    // Source selection (object-form `plugins`) is shared with the worker
    // loader via `selectPluginSource` so the two never drift.
    const pluginSource = selectPluginSource(entry);
    const eslintPluginMeta: Record<string, { ruleNames: string[] }> = {};
    const pluginPrefixes: string[] = [];
    if (pluginSource != null) {
      for (const prefix of Object.keys(pluginSource)) {
        if (NATIVE_PLUGIN_RESERVED_NAMES.has(prefix)) {
          throw new Error(
            `[rslint] Config entry at index ${index}: plugins prefix "${prefix}" collides with the built-in plugin of the same name; choose a different prefix.`,
          );
        }
        const plugin = unwrapPluginModule(pluginSource[prefix]);
        const pluginRules = plugin?.rules;
        if (pluginRules == null || typeof pluginRules !== 'object') {
          throw new Error(
            `[rslint] Config entry at index ${index}: plugins["${prefix}"] must expose a "rules" object.`,
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
    // array-form (native) names the user wrote with the object-form
    // (community) prefixes; drop the live object-form `plugins` values
    // (functions can't be JSON-serialized to Go). Preserve an explicitly
    // empty field, but do not add `plugins` when the field was omitted.
    const stringPlugins = Array.isArray(entry.plugins)
      ? entry.plugins.filter((p): p is string => typeof p === 'string')
      : [];
    const plugins = [...new Set([...stringPlugins, ...pluginPrefixes])];
    // ESLint decides whether an ignores-bearing object is global from its
    // authored keys, including keys whose value is undefined. Preserve that
    // distinction when normalization omits an undefined or unsupported field.
    const authoredNonGlobalKey = Object.keys(entry).some(
      (key) => key !== 'name' && key !== 'ignores',
    );
    const serializesNonGlobalKey =
      hasFiles ||
      entry.languageOptions !== undefined ||
      entry.rules !== undefined ||
      hasPlugins ||
      entry.settings !== undefined;
    const needsNonGlobalShapeMarker =
      authoredNonGlobalKey && !serializesNonGlobalKey;

    return {
      ...(entry.name !== undefined ? { name: entry.name } : {}),
      ...(hasFiles ? { files: entry.files } : {}),
      ...(entry.ignores !== undefined ? { ignores: entry.ignores } : {}),
      ...(entry.languageOptions !== undefined
        ? { languageOptions: entry.languageOptions }
        : {}),
      ...(entry.rules !== undefined ? { rules: entry.rules } : {}),
      ...(hasPlugins ? { plugins } : {}),
      ...(entry.settings !== undefined
        ? { settings: entry.settings }
        : needsNonGlobalShapeMarker
          ? { settings: {} }
          : {}),
      ...(pluginPrefixes.length > 0 ? { eslintPlugins: eslintPluginMeta } : {}),
    };
  });
}

const GLOBAL_ACCESS_VALUES = new Set<unknown>([
  true,
  'true',
  'writable',
  'writeable',
  false,
  'false',
  'readonly',
  'readable',
  null,
  'off',
]);

const RULE_SEVERITIES = new Set<unknown>(['off', 'warn', 'error', 0, 1, 2]);

function validateRules(
  rules: Record<string, unknown>,
  entryIndex: number,
): void {
  for (const [name, value] of Object.entries(rules)) {
    const severity = Array.isArray(value) ? value[0] : value;
    if (
      (Array.isArray(value) && value.length === 0) ||
      !RULE_SEVERITIES.has(severity)
    ) {
      throw new Error(
        `[rslint] Config entry at index ${entryIndex}: rule "${name}" must use severity ` +
          `'off', 'warn', 'error', 0, 1, or 2`,
      );
    }
  }
}

function validateGlobals(value: unknown, entryIndex: number): void {
  if (value === null || typeof value !== 'object' || Array.isArray(value)) {
    throw new Error(
      `[rslint] Config entry at index ${entryIndex}: "languageOptions.globals" must be an object`,
    );
  }
  for (const [name, access] of Object.entries(value)) {
    if (name !== name.trim()) {
      throw new Error(
        `[rslint] Config entry at index ${entryIndex}: global "${name}" has leading or trailing whitespace`,
      );
    }
    if (!GLOBAL_ACCESS_VALUES.has(access)) {
      throw new Error(
        `[rslint] Config entry at index ${entryIndex}: global "${name}" must be "readonly", "writable", or "off"`,
      );
    }
  }
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
 * actual rules per config. This is not the validation boundary: worker config
 * loading rejects conflicting plugin definitions, including duplicate routing
 * descriptors that mount different instances under the same prefix.
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

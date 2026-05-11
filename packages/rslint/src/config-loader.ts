import path from 'node:path';
import { pathToFileURL } from 'node:url';
import type { ESLintPluginShape } from './define-config.js';
import { NATIVE_PLUGIN_PREFIXES } from './define-config.js';
import { selectPluginSource } from '@rslint/eslint-plugin-runner/plugin-source';

// Re-export utilities that external consumers may depend on
export {
  JS_CONFIG_FILES,
  findJSConfig,
  findJSConfigUp,
  findJSConfigsInDir,
  filterConfigsByParentIgnores,
} from './utils/config-discovery.js';
export type { ConfigEntry as DiscoveredConfigEntry } from './utils/config-discovery.js';

/**
 * Lean wire-format plugin entry. Carries only what Go's
 * `RegisterEslintPluginRules` and the `enforcePlugins` gate read off
 * the IPC `init` payload — `prefix` for the rule namespace, `ruleNames`
 * for the placeholder-rule registration.
 *
 * Production: workers no longer consume this — they import the user's
 * config file directly and pull live plugin instances out of it (see
 * `loadPluginsFromConfigs` in the runner). This shape exists purely to
 * keep Go's existing IPC contract working without re-shipping plugin
 * metadata that Node already has on the wire path.
 *
 * The Go-side `internal/config.EslintPluginEntry` struct has
 * `omitempty` on `specifier`/`version`/`options`/`resolvedPath`, so
 * decoding this leaner shape is forward-compatible.
 */
export interface EslintPluginEntry {
  /** User-chosen rule namespace, e.g. `'uc'`. Becomes `<prefix>/<ruleName>`. */
  prefix: string;
  /** Names of rules contributed by this plugin (sorted, stable across runs). */
  ruleNames: string[];
}

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
 *
 * In the configs-flow architecture the worker imports the user's
 * rslint config file directly and pulls plugin instances from there.
 * The main thread no longer needs to resolve plugin paths or fingerprint
 * versions — it only walks the normalized config to (a) collect the
 * `(prefix, ruleNames)` tuples Go's IPC `init` payload requires and
 * (b) build the `ConfigDescriptor[]` the worker pool consumes.
 */
export function normalizeConfig(config: unknown): Record<string, unknown>[] {
  if (!Array.isArray(config)) {
    throw new Error(
      `rslint config must export an array (flat config format), got ${typeof config}`,
    );
  }

  return config
    .map((entry: unknown, index: number) => ({ entry, index }))
    .filter(
      (item): item is { entry: Record<string, unknown>; index: number } => {
        const { entry, index } = item;
        if (entry == null || typeof entry !== 'object') {
          // process.stderr.write instead of console.warn so the message
          // honors the host's stderr redirection. In LSP mode the
          // extension routes stderr through the language-client output
          // channel; console.warn bypassed that and landed in the
          // developer console instead — invisible to users debugging
          // their config.
          process.stderr.write(
            `[rslint] Config entry at index ${index} is not an object (got ${entry === null ? 'null' : typeof entry}), skipping.\n`,
          );
          return false;
        }
        return true;
      },
    )
    .map(({ entry, index }) => {
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

      // ── plugins field shape detection ───────────────────────────
      //
      // Two valid shapes for `entry.plugins`:
      //
      //   - string[]      → JSON-config style, plugin NAMES only. Used by
      //                     Go's MergedConfig.Plugins as a name-allowlist
      //                     for native rule enabling. No live plugin
      //                     instance is supplied.
      //   - object        → Standard ESLint flat-config style:
      //                     `{ prefix: pluginInstance }`. We MUST fold the
      //                     values into `eslintPlugins` so plugin rules
      //                     actually load — without folding, the plugin
      //                     instance hits the wire as `{}` (functions
      //                     don't survive JSON) and Go interprets the
      //                     missing string-array as zero allowed plugins.
      //
      // Anything else (number, boolean, ...) is rejected — the previous
      // silent passthrough produced "rule not found" diagnostics with no
      // hint, which is the bug review #7 surfaced.
      let pluginsForOutput: unknown = entry.plugins;
      if (entry.plugins != null) {
        if (Array.isArray(entry.plugins)) {
          // JSON-config style — string array of names. Validate each
          // element is a string so a mistyped `plugins: [{...}]` fails
          // loudly instead of producing junk in Go's Plugins set.
          for (let i = 0; i < entry.plugins.length; i++) {
            if (typeof entry.plugins[i] !== 'string') {
              throw new Error(
                `[rslint] Config entry at index ${index}: "plugins[${i}]" must be a string ` +
                  `when "plugins" is an array, got ${typeof entry.plugins[i]}`,
              );
            }
          }
        } else if (typeof entry.plugins === 'object') {
          // Flat-config style — object of `{ prefix: pluginInstance }`.
          // Keys become the name allowlist; the live instances are picked
          // up later via selectPluginSource(entry).
          pluginsForOutput = Object.keys(entry.plugins);
        } else {
          throw new Error(
            `[rslint] Config entry at index ${index}: "plugins" must be a string array or ` +
              `an object mapping prefix → plugin instance, got ${typeof entry.plugins}`,
          );
        }
      }

      // ── eslintPlugins resolution ────────────────────────────────
      //
      // The output `eslintPlugins` is a lean wire-format projection
      // (`{ prefix, ruleNames }[]`). Two things consume it downstream:
      //
      //   - Go's IPC `init` payload reads `prefix + ruleNames` to
      //     register placeholder rules + the `enforcePlugins` gate
      //     (see `internal/config.RegisterEslintPluginRules`).
      //   - The host (`cli.ts` / vscode `Rslint.ts`) reads `eslintPlugins`
      //     to know whether a given config has any plugins (workerCount=0
      //     fast path) — the array's presence/length is the signal.
      //
      // Source precedence (matches `loadPluginsFromConfigFile` in the
      // runner so worker and host see the same prefix set):
      //   - explicit `eslintPlugins`: takes precedence.
      //   - object-form `plugins`: implicit fallback.
      //   - neither: no eslintPlugins on output.
      //
      // The live plugin instances are NOT shipped — workers import the
      // config file themselves and pull instances from there. This is
      // what unlocks local-path plugins (they don't have an npm
      // specifier the main thread could resolve) and keeps plugin
      // resolution naturally anchored at each config's own directory.
      let eslintPlugins: EslintPluginEntry[] | undefined;
      // Fail-fast on a malformed explicit `eslintPlugins` (the host is the
      // config entry point; the shared predicate stays lenient for the
      // worker). Source SELECTION is delegated to `selectPluginSource` so
      // host and worker share one precedence rule (explicit `eslintPlugins`
      // over object-form `plugins`) and can't drift.
      if (
        entry.eslintPlugins != null &&
        (typeof entry.eslintPlugins !== 'object' ||
          Array.isArray(entry.eslintPlugins))
      ) {
        throw new Error(
          `[rslint] Config entry at index ${index}: "eslintPlugins" must be an object ` +
            `mapping prefix → plugin instance, got ${Array.isArray(entry.eslintPlugins) ? 'array' : typeof entry.eslintPlugins}`,
        );
      }
      const extractionSource: Record<string, ESLintPluginShape> | undefined =
        (selectPluginSource(entry) as Record<
          string,
          ESLintPluginShape
        > | null) ?? undefined;

      if (extractionSource != null) {
        // Fail fast on prefixes that shadow rslint's built-in (Go-side
        // ported) plugin namespaces. The rule-name winner logic
        // (native rule wins on collision, per
        // `internal/config/rule_registry.go`) is intentionally
        // unchanged — but using one of these prefixes as the key in
        // `eslintPlugins` means most of the user's plugin object is
        // silently ignored at the rule-lookup stage. That surprise is
        // best caught here, before the Go child is even spawned, with
        // a message that tells the user EXACTLY what to do (rename
        // the prefix).
        const conflicts: string[] = [];
        for (const prefix of Object.keys(extractionSource)) {
          if ((NATIVE_PLUGIN_PREFIXES as readonly string[]).includes(prefix)) {
            conflicts.push(prefix);
          }
        }
        if (conflicts.length > 0) {
          const list = conflicts.map((p) => `  "${p}"`).join('\n');
          const nativeList = NATIVE_PLUGIN_PREFIXES.join(', ');
          throw new Error(
            `[rslint] Config entry at index ${index}: "eslintPlugins" uses prefix(es) ` +
              `that conflict with rslint's built-in native plugins:\n${list}\n\n` +
              `rslint already ships ported rules under these namespaces, and on name ` +
              `collision the native rule wins — so passing your own plugin instance under ` +
              `the same prefix means your same-named rules are silently shadowed and only ` +
              `your different-named rules fire. That mix is almost always a mistake.\n\n` +
              `Fix: rename the prefix in eslintPlugins to something unique, e.g.\n` +
              `  eslintPlugins: { myUnicorn: pluginObj },\n` +
              `  rules: { "myUnicorn/some-rule": "error" }\n\n` +
              `Reserved native namespaces: ${nativeList}`,
          );
        }
        eslintPlugins = [];
        for (const prefix of Object.keys(extractionSource)) {
          const plugin = extractionSource[prefix];
          if (plugin == null || typeof plugin !== 'object') {
            throw new Error(
              `[rslint] Config entry at index ${index}: "eslintPlugins.${prefix}" ` +
                `must be an ESLint plugin object, got ${plugin === null ? 'null' : typeof plugin}`,
            );
          }
          // Object.keys returns a fresh array, so .sort() in place
          // doesn't mutate plugin.rules's iteration order. Sorted to
          // keep the IPC payload byte-stable across runs (golden tests
          // and conformance suites rely on this).
          //
          // Validate `rules` is a plain object. Previous logic was
          // `plugin.rules ? Object.keys(plugin.rules).sort() : []` —
          // which accepts truthy non-objects like strings or functions.
          // `Object.keys('abc')` returns `['0','1','2']`, polluting the
          // wire payload with junk rule names. Strict-check to keep
          // garbage data out of Go's placeholder registry.
          let ruleNames: string[] = [];
          if (
            plugin.rules != null &&
            typeof plugin.rules === 'object' &&
            !Array.isArray(plugin.rules)
          ) {
            ruleNames = Object.keys(plugin.rules).sort();
          } else if (plugin.rules != null) {
            throw new Error(
              `[rslint] Config entry at index ${index}: ` +
                `"eslintPlugins.${prefix}.rules" must be a plain object, ` +
                `got ${Array.isArray(plugin.rules) ? 'array' : typeof plugin.rules}`,
            );
          }
          eslintPlugins.push({ prefix, ruleNames });
        }
      }

      return {
        files: entry.files,
        ignores: entry.ignores,
        languageOptions: entry.languageOptions,
        rules: entry.rules,
        plugins: pluginsForOutput,
        settings: entry.settings,
        ...(eslintPlugins != null && eslintPlugins.length > 0
          ? { eslintPlugins }
          : {}),
      };
    });
}

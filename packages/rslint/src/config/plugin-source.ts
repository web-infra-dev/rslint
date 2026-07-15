/* rslint-disable @typescript-eslint/no-unsafe-type-assertion */
/**
 * Shared source-selection + unwrap logic for ESLint-plugin mounting.
 *
 * Used by BOTH the worker loader (`plugin-loader.ts`, which instantiates
 * live plugins) and the host config normalizer (`config-file-loader.ts`, which
 * extracts `{prefix, ruleNames}` metadata for Go). Centralizing the
 * precedence + unwrap rules guarantees host and worker agree on which
 * plugin object belongs to which prefix — otherwise Go could register a
 * prefix the worker never loaded, and every file under it would report
 * "rule not found".
 *
 * Pure functions, no runtime deps — safe to bundle into the main library
 * (where `config-file-loader.ts` lives) and the worker alike.
 */

/**
 * Select the live community-plugin map for a config entry: the object-form
 * `plugins` (standard ESLint flat-config users write
 * `plugins: { uc: unicornPlugin }`). Array-form `plugins` (the native-name
 * whitelist) is ignored — it carries no live objects. Returns null when
 * `plugins` is absent or not an object map.
 */
export function selectPluginSource(
  entry: unknown,
): Record<string, unknown> | null {
  if (entry == null || typeof entry !== 'object') return null;
  const e = entry as { plugins?: unknown };
  if (
    e.plugins != null &&
    typeof e.plugins === 'object' &&
    !Array.isArray(e.plugins)
  ) {
    return e.plugins as Record<string, unknown>;
  }
  return null;
}

/**
 * Unwrap CJS/ESM dual-shape: prefer `mod.default` if present (ESM/interop),
 * fall back to `mod` itself (legacy CJS without `default`). Returns null
 * for non-object inputs. The generic lets the worker loader recover its
 * concrete plugin shape without an external cast.
 */
export function unwrapPluginModule<T = Record<string, unknown>>(
  mod: unknown,
): T | null {
  if (mod == null || typeof mod !== 'object') return null;
  const m = mod as { default?: unknown };
  if (m.default != null && typeof m.default === 'object') {
    return m.default as T;
  }
  return mod as T;
}

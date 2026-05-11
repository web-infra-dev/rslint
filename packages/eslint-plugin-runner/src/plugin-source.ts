/**
 * Per-entry plugin source selection — the SINGLE definition of the
 * host↔worker precedence contract.
 *
 * Both the host (`@rslint/core`'s `normalizeConfig`, which extracts
 * `{ prefix, ruleNames }` for Go's placeholder registry) and the worker
 * (`plugin-loader`, which extracts live plugin instances) must pick the
 * SAME source per config entry — otherwise Go registers rule names the
 * worker never loads (→ "rule not found"), or vice versa. This precedence
 * previously lived in two hand-synced copies kept aligned only by comment.
 *
 * Precedence: explicit `eslintPlugins` wins over object-form `plugins`.
 * Both must be plain (non-array) objects to qualify; returns null when
 * neither is usable.
 *
 * Intentionally lenient — a malformed `eslintPlugins` (non-object) falls
 * through to `plugins` rather than throwing, because the worker re-imports
 * the raw config and must not crash at runtime. The host layers its own
 * fail-fast validation on top (its `normalizeConfig` throws on a malformed
 * `eslintPlugins`); that is a host concern, not part of this shared
 * precedence contract.
 */
function isRecord(value: unknown): value is Record<string, unknown> {
  return value != null && typeof value === 'object' && !Array.isArray(value);
}

export function selectPluginSource(entry: {
  eslintPlugins?: unknown;
  plugins?: unknown;
}): Record<string, unknown> | null {
  if (isRecord(entry.eslintPlugins)) {
    return entry.eslintPlugins;
  }
  if (isRecord(entry.plugins)) {
    return entry.plugins;
  }
  return null;
}

/**
 * Whether a NORMALIZED config entry declares at least one eslint-plugin —
 * i.e. its `eslintPlugins` is a non-empty wire array of
 * `{ prefix, ruleNames }`. Both worker-descriptor builders gate on this to
 * decide whether a config needs a worker at all: the CLI host (`cli.ts`)
 * and the LSP host (vscode `compat-pool-helpers`). Sharing the predicate
 * keeps the two from disagreeing on which configs are plugin-bearing
 * (a mismatch routes a file to an empty plugin set).
 *
 * Note: this inspects the NORMALIZED shape (post-`normalizeConfig`, where
 * `eslintPlugins` is an array) — distinct from `selectPluginSource`, which
 * reads the RAW config (object-form maps).
 */
export function configEntryHasEslintPlugins(entry: {
  eslintPlugins?: unknown;
}): boolean {
  return Array.isArray(entry.eslintPlugins) && entry.eslintPlugins.length > 0;
}

/**
 * Per-worker cache for a rule's resolved `context.options`, keyed by
 * `(configKey, ruleName, options content)`.
 *
 * Real ESLint runs a whole lint pass in one process and never re-clones
 * `context.options` per file (`deepMergeArrays` returns the user's
 * option objects by reference when no default merge is needed), so
 * `context.options[i]` stays the SAME object across every file for a
 * static rule config. Plugin rules rely on that — e.g. caching derived
 * state in a `WeakMap` keyed by `context.options` itself.
 *
 * rslint dispatches each file to a worker thread via
 * `worker.postMessage`, which structured-clones the message payload —
 * so `req.rules[rule].options` is already a fresh object on arrival at
 * `lintFile`, independent of anything `applyOptionDefaults` does. This
 * cache restores the identity stability plugins expect by memoizing
 * the resolved options per rule for the lifetime of the worker: the
 * first file that hits a given (configKey, ruleName, content) triple
 * computes it, every later file with the same triple reuses the same
 * object.
 *
 * Cached results are deep-frozen. ESLint tolerates a rule mutating its
 * own options only because nothing else reads the corrupted state
 * before the process exits for that file; here the same object is
 * reused by later files, so a mutating rule now throws immediately
 * (caught by the per-rule try/catch in `ecma-language-plugin.ts`)
 * instead of silently leaking state into unrelated files.
 */

const cache = new Map<string, unknown[]>();

/**
 * Return the cached resolved options for `(configKey, ruleName,
 * userOptions)`, computing and freezing them via `compute()` on a
 * cache miss. Falls back to always calling `compute()` uncached when
 * `configKey` is empty (ad-hoc / test callers that don't scope a
 * config) or when `userOptions` isn't safely key-able (contains a
 * function, a Proxy, or another value `structuredClone` rejects).
 */
export function resolveOptionsCached(
  configKey: string,
  ruleName: string,
  userOptions: readonly unknown[],
  compute: () => unknown[],
): unknown[] {
  if (configKey === '') return compute();

  const contentKey = stableKey(userOptions);
  if (contentKey === undefined) return compute();

  const cacheKey = `${configKey} ${ruleName} ${contentKey}`;
  const hit = cache.get(cacheKey);
  if (hit !== undefined) return hit;

  const computed = compute();
  deepFreeze(computed);
  cache.set(cacheKey, computed);
  return computed;
}

/**
 * A stable, sorted-key JSON string for `value`, or `undefined` if
 * `value` contains something that can't be told apart through JSON
 * (functions collapse to `null` in arrays / are dropped from objects,
 * which would make two DIFFERENT function-valued options key-collide).
 * Gated on `structuredClone` succeeding first — the same non-cloneable
 * values (functions, Proxies, ...) that `options-defaults.ts`'s
 * `cloneDefault` guards against are exactly the ones JSON can't tell
 * apart, so reuse that check rather than walking the value twice.
 */
function stableKey(value: readonly unknown[]): string | undefined {
  try {
    structuredClone(value);
  } catch {
    return undefined;
  }
  return JSON.stringify(value, sortKeysReplacer);
}

function sortKeysReplacer(_key: string, value: unknown): unknown {
  if (value === null || typeof value !== 'object' || Array.isArray(value)) {
    return value;
  }
  const sorted: Record<string, unknown> = {};
  for (const k of Object.keys(value as Record<string, unknown>).sort()) {
    sorted[k] = (value as Record<string, unknown>)[k];
  }
  return sorted;
}

function deepFreeze(value: unknown): void {
  if (value === null || typeof value !== 'object' || Object.isFrozen(value)) {
    return;
  }
  Object.freeze(value);
  for (const key of Object.keys(value as Record<string, unknown>)) {
    deepFreeze((value as Record<string, unknown>)[key]);
  }
}

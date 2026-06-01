/* eslint-disable @typescript-eslint/no-unsafe-type-assertion -- AST / parser / scope-manager / plugin-API boundary casts. Each site projects from an `any` / `unknown` peer surface (oxc-parser output, user plugin objects, ESLint v10 wire shapes) into the typed shape this module uses; the contract is runtime-validated at the call boundaries above, not at the cast. Bulk-disabling here instead of per-line keeps the cast sites readable. */
/**
 * Compute a rule's runtime `context.options` from the user's options and
 * `rule.meta.defaultOptions`, matching ESLint v10.
 *
 * **Why this exists** (the canonical motivating case is unicorn/no-null):
 * Many ESLint plugins destructure `context.options[0]` eagerly inside
 * `create()`:
 *
 *   create(context) {
 *     const { checkStrictEquality } = context.options[0];  // ← throws if undefined
 *     ...
 *   }
 *
 * When the user writes `'unicorn/no-null': 'error'` (no options array),
 * the rule still gets a populated `options[0]` because the plugin ships
 * `meta.defaultOptions: [{ checkStrictEquality: false }]`. ESLint v10
 * materializes that via `deepMergeArrays(meta.defaultOptions ?? [],
 * userOptions)` (`linter.js` getRuleOptions); the runner does the same.
 *
 * **Schema property `default`s ARE materialized** (verified against
 * eslint@10.4.0 / the identical eslint@9.32.0 oracle): after
 * `getRuleOptions` builds the options array via `deepMergeArrays`, the
 * rule validator runs ajv with `useDefaults: true` (`shared/ajv.js`),
 * which fills each schema `properties[k].default` IN PLACE into the
 * object already present at a slot — and the filled array IS what the
 * rule receives as `context.options`. The two passes compose; this
 * function reproduces both in order:
 *
 *   1. `meta.defaultOptions` deep-merge   (the `deepMergeArrays` part)
 *   2. schema property-default fill        (the ajv `useDefaults` part)
 *
 * ajv `useDefaults` fills missing PROPERTIES of an object that ALREADY
 * occupies a slot; it NEVER creates a positional slot. So a slot-level
 * `default` (one not inside `properties`) is never materialized, and a
 * severity-only config never fabricates an option object:
 *
 *   schema:[{default:'always'}]                 , no user → []   (no slot)
 *   schema:[{type:'object'}]                    , no user → []   (no slot)
 *   schema:[{properties:{mode:{default:'auto'}}}], user [{}] → [{mode:'auto'}]
 *   nested properties.a.properties.x.default=1  , user [{a:{}}] → [{a:{x:1}}]
 *   user value present                          , user [{mode:'x'}] → [{mode:'x'}]
 *
 * When a slot carries both a slot-level `default` and a conflicting
 * `meta.defaultOptions`, v10 takes defaultOptions (the author's current
 * intent) — preserved here because the property fill only looks at
 * `properties`, never the slot-level `default`.
 */

/**
 * The shape of `rule.meta.schema` we tolerate. ESLint's API has many
 * variations; we read defensively.
 */
export type RuleSchema =
  | readonly RuleSchemaEntry[]
  | RuleSchemaEntry
  | undefined
  | null;

export interface RuleSchemaEntry {
  type?: string;
  default?: unknown;
  /** ajv `useDefaults` fills these into an object already at a slot. */
  properties?: Record<string, RuleSchemaEntry>;
  /**
   * Array-schema form: a single object schema can carry `items` to
   * describe its positional elements — either a tuple (`items: [e0, e1]`,
   * one entry per slot) or a single schema applied to every element
   * (`items: { … }`).
   */
  items?: RuleSchemaEntry | readonly RuleSchemaEntry[];
  // Anything else is ignored.
  [key: string]: unknown;
}

/**
 * Return a new options array where the user's values are preserved
 * verbatim, `meta.defaultOptions` is deep-merged under each slot, and
 * schema `properties[k].default`s are filled into object slots that
 * already exist. No positional slot is ever fabricated.
 *
 * @example
 * applyOptionDefaults(
 *   [{}],
 *   [{ type: 'object', properties: { mode: { default: 'auto' } } }]
 * )
 * // → [{ mode: 'auto' }]
 */
export function applyOptionDefaults(
  userOptions: readonly unknown[] | undefined,
  schema: RuleSchema,
  defaultOptions?: readonly unknown[],
): unknown[] {
  const ua = userOptions ? userOptions.map(cloneDefault) : [];

  // ── Step 1: `rule.meta.defaultOptions` overlay ──
  // Deep-merge each entry under the user's value for the same positional
  // slot (user wins on collision). This IS the v10
  // `deepMergeArrays(defaultOptions, userOptions)` behavior. Many real
  // plugins rely on it exclusively (e.g. `unicorn/no-null` ships
  // `defaultOptions:[{checkStrictEquality:false}]`,
  // `unicorn/prefer-number-properties` ships
  // `defaultOptions:[{checkInfinity:false,checkNaN:true}]`) — it's what
  // materializes `context.options[0]` for rules that destructure it
  // eagerly.
  if (Array.isArray(defaultOptions)) {
    for (let i = 0; i < defaultOptions.length; i++) {
      const d = defaultOptions[i];
      if (d === undefined) continue;
      ua[i] = mergeDefault(ua[i], d);
    }
  }

  // ── Step 2: schema property-default fill (ajv `useDefaults: true`) ──
  // Runs AFTER step 1 so a slot that step 1 (or the user) materialized
  // gets its missing schema-default properties filled — matching v10,
  // where the rule validator validates the already-merged options array
  // in place. Only fills INTO an object value at a slot that already
  // exists; never creates a slot (so a severity-only config stays []).
  for (let i = 0; i < ua.length; i++) {
    const slotSchema = schemaForSlot(schema, i);
    if (slotSchema) fillSchemaPropertyDefaults(ua[i], slotSchema);
  }

  return ua;
}

/**
 * Resolve the schema entry that governs positional option slot `i`.
 *
 * ESLint accepts the per-slot schema as either an array (`[e0, e1, …]`,
 * one entry per slot) or a single object schema carrying `items` — a
 * tuple (`items: [e0, e1]`) or a single schema applied to all elements
 * (`items: { … }`). A bare non-array object schema (no `items`) is NOT a
 * per-slot schema in ESLint (it validates the whole options array and
 * ESLint rejects such a rule config), so it yields no per-slot entry.
 */
function schemaForSlot(
  schema: RuleSchema,
  i: number,
): RuleSchemaEntry | undefined {
  if (schema == null) return undefined;
  if (Array.isArray(schema)) {
    return schema[i] as RuleSchemaEntry | undefined;
  }
  const items = (schema as RuleSchemaEntry).items;
  if (items === undefined) return undefined;
  if (Array.isArray(items)) {
    return items[i] as RuleSchemaEntry | undefined;
  }
  // Single schema applies to every element.
  return items as RuleSchemaEntry;
}

/**
 * Fill `schema.properties[k].default` into `target` for each key `k`
 * that `target` is missing, then recurse into nested object schemas.
 * No-op unless `target` is a plain object (mirrors ajv `useDefaults`,
 * which only fills missing properties of an object that is present).
 */
function fillSchemaPropertyDefaults(
  target: unknown,
  schemaEntry: RuleSchemaEntry,
): void {
  if (!isPlainObject(target)) return;
  const props = schemaEntry.properties;
  if (props == null || typeof props !== 'object') return;
  for (const key of Object.keys(props)) {
    const propSchema = props[key];
    if (propSchema == null || typeof propSchema !== 'object') continue;
    if (!(key in target) && 'default' in propSchema) {
      target[key] = cloneDefault(propSchema.default);
    }
    // Recurse into a nested object schema regardless of whether the key
    // was just defaulted or supplied by the user — v10's ajv fills
    // nested property defaults at every depth.
    if (isPlainObject(target[key])) {
      fillSchemaPropertyDefaults(target[key], propSchema);
    }
  }
}

function isPlainObject(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

/**
 * Deep merge `defaultValue` under `userValue`. User wins on conflicts.
 * Objects merge per-key; arrays / primitives use user's value verbatim
 * when present, else default.
 */
function mergeDefault(userValue: unknown, defaultValue: unknown): unknown {
  if (userValue === undefined) {
    return cloneDefault(defaultValue);
  }
  if (
    userValue &&
    defaultValue &&
    typeof userValue === 'object' &&
    typeof defaultValue === 'object' &&
    !Array.isArray(userValue) &&
    !Array.isArray(defaultValue)
  ) {
    const result: Record<string, unknown> = {
      ...(defaultValue as Record<string, unknown>),
    };
    for (const k of Object.keys(userValue as Record<string, unknown>)) {
      result[k] = mergeDefault(
        (userValue as Record<string, unknown>)[k],
        (defaultValue as Record<string, unknown>)[k],
      );
    }
    return result;
  }
  return userValue;
}

function cloneDefault(value: unknown): unknown {
  // Deep-clone the schema's / rule's default value before handing it to
  // a plugin rule's `create()`. A rule that mutates `context.options[0]`
  // (push to an array, assign to a property — common cache / hit-counter
  // / lazy-init pattern) would otherwise corrupt the schema's default
  // in place, and the NEXT file's lint would see the mutated value as
  // its starting state.
  //
  // Matches ESLint v10's effective behavior: schema.default is inserted
  // via ajv with `useDefaults: true`, which internally JSON-clones the
  // default before mutation. `defaultOptions` overlay goes through
  // `deepMergeArrays` → fresh nested objects. Our previous shallow
  // clone (`{ ...value }` / `value.slice()`) left inner arrays /
  // objects shared, so a rule pushing to `opts.ignored` still
  // polluted the schema's default across calls.
  //
  // `structuredClone` (Node 17+; runner engines >= 20) handles arrays,
  // plain objects, Date, Map, Set, etc. Primitives pass through.
  if (value == null || typeof value !== 'object') return value;
  try {
    return structuredClone(value);
  } catch {
    // `structuredClone` throws `DataCloneError` on a non-cloneable value
    // — most commonly a function on an option object (`{ cb() {} }`) or
    // a Proxy. ESLint never deep-clones options (it passes them by
    // reference), so the safe and v10-matching fallback is to return the
    // value by reference. The clone exists only to stop a rule mutating
    // a SHARED default in place and leaking into the next file's lint; a
    // function / Proxy can't be mutated that way (their identity, not
    // their contents, is what a rule reads), so by-ref is harmless here.
    // Without this guard the throw escaped the per-rule try/catch on the
    // in-process `lintFile` path and failed the WHOLE file.
    return value;
  }
}

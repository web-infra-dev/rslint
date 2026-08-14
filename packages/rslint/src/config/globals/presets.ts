/**
 * Rslint-specific global sets that do not exist in the upstream `globals`
 * package. Add small presets here; they are bundled as ordinary runtime
 * objects and do not pass through the JSON asset emitter.
 */
export const RSLINT_GLOBAL_SETS = {} as const satisfies Record<
  string,
  Readonly<Record<string, boolean>>
>;

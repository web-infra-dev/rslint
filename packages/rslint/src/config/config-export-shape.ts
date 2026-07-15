/**
 * Normalize the modern flat-config module export shape without touching entry
 * contents. ESLint accepts either one config object or an array; a missing
 * export is an empty config. Keeping this boundary separate lets the main
 * serializer and plugin worker consume the exact same entry list while the
 * worker retains live plugin objects.
 */
export function configExportEntries(config: unknown): unknown[] {
  if (config === undefined) return [];
  if (Array.isArray(config)) {
    for (let index = 0; index < config.length; index++) {
      if (!Object.prototype.hasOwnProperty.call(config, index)) {
        throw new Error(
          `[rslint] Config entry at index ${index}: unexpected undefined config`,
        );
      }
    }
    // ESLint consumes the exported array through its iterator for every
    // lexical config-array normalization. Copying here preserves observable
    // iterator side effects while callers cannot mutate the raw module export.
    return [...config];
  }
  if (
    config !== null &&
    typeof config === 'object' &&
    Object.keys(config).length === 0
  ) {
    return [];
  }
  return [config];
}

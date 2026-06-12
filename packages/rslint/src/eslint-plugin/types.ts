/**
 * Types shared between the eslint-plugin host and its lint workers.
 *
 * Wire-format / IPC frame types (the Go↔Node frame contract) live in
 * `src/ipc/protocol.ts` — the single source the CLI host consumes.
 */

/**
 * Per-config descriptor handed to the worker pool. Each worker imports
 * every descriptor's `configPath` once at init, then routes per-file
 * lint tasks via `configKey === configDirectory` to the right plugin
 * instances. The `configDirectory` here MUST match the value Go writes
 * into `EslintPluginLintFile.ConfigKey` byte-for-byte; the worker uses
 * it as a Map key for per-file dispatch.
 */
export interface ConfigDescriptor {
  /** Absolute filesystem path of the rslint config file (`rslint.config.{js,mjs,ts,mts}`). */
  configPath: string;
  /** Absolute filesystem path of the directory holding the config file.
   *  Matches the `ConfigKey` Go emits per file during plugin-lint dispatch. */
  configDirectory: string;
}

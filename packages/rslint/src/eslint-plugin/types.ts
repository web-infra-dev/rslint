/**
 * Types shared between the eslint-plugin host and its lint workers.
 *
 * Wire-format / IPC frame types (the Go↔Node frame contract) live in
 * `src/ipc/protocol.ts` — the single source the CLI host consumes.
 */

/**
 * ESLint-compatible access values carried to the plugin worker. This mirrors
 * the public config authoring type without coupling the worker build project to
 * the config-loader build project.
 */
export type GlobalAccess =
  | boolean
  | null
  | 'true'
  | 'false'
  | 'readonly'
  | 'readable'
  | 'writable'
  | 'writeable'
  | 'off';

export type GlobalsConfig = Record<string, GlobalAccess>;

/**
 * Per-config descriptor handed to the worker pool. Each worker imports
 * every descriptor's `configPath` once at init, then routes per-file
 * lint tasks via `configKey === configDirectory` to the right plugin
 * instances. The `configDirectory` here MUST match the value Go writes
 * into `EslintPluginLintFile.ConfigKey` byte-for-byte; the worker uses
 * it as a Map key for per-file dispatch.
 */
export interface ConfigDescriptor {
  /** Absolute filesystem path of the selected JS/TS config file. */
  configPath: string;
  /** Go-authoritative absolute matching/routing directory. In explicit mode
   *  this can differ from the config file's parent and MUST byte-match the
   *  `ConfigKey` Go emits during plugin-lint dispatch. */
  configDirectory: string;
}

/**
 * Logical protocol shared by the native CLI, programmatic API, and LSP config
 * discovery adapters. The transports differ (framed IPC versus JSON-RPC), but
 * every adapter sends these payloads to the same Node-side module host.
 *
 * Paths in requests identify values selected by Go. Node may native-normalize
 * configPath for local I/O, but configDirectory is an opaque routing identity
 * and must retain its exact spelling. Load responses correlate only by opaque
 * candidate ID. Activation responses contain only the transaction identity and
 * plugin metadata consumed by Go; Node-only preparation data never crosses the
 * transport boundary.
 */

export const CONFIG_DISCOVERY_PROTOCOL_VERSION = 1 as const;

export type ConfigModuleLoadMode = 'cached' | 'fresh';

export interface ConfigModuleCandidate {
  /** Opaque, transaction-local identity allocated by the Go coordinator. */
  id: string;
  /** Absolute path to the JS/TS config module Node must execute. */
  configPath: string;
  /** Go-authoritative directory used for config matching and plugin routing. */
  configDirectory: string;
}

export interface LoadConfigsRequest {
  protocolVersion: typeof CONFIG_DISCOVERY_PROTOCOL_VERSION;
  /** Isolates batches belonging to one discovery transaction. */
  transactionId: string;
  /**
   * `cached` preserves one-shot CLI module-import semantics. `fresh` is used
   * by long-lived API and editor refreshes; it cache-busts the entry module,
   * while static transitive imports retain Node's normal module cache.
   */
  loadMode: ConfigModuleLoadMode;
  /** Serialize module evaluation when the CLI requested --singleThreaded. */
  singleThreaded?: boolean;
  candidates: ConfigModuleCandidate[];
}

export interface ConfigModuleEslintPluginEntry {
  prefix: string;
  ruleNames: string[];
}

export interface LoadedConfigModuleResult {
  id: string;
  status: 'loaded';
  /** Serializable normalizeConfig output. Live plugin objects never cross. */
  entries: Record<string, unknown>[];
}

export interface FailedConfigModuleResult {
  id: string;
  status: 'failed';
  error: {
    message: string;
    code?: string;
  };
}

export type ConfigModuleLoadResult =
  LoadedConfigModuleResult | FailedConfigModuleResult;

export interface LoadConfigsResponse {
  transactionId: string;
  /** Results are in exactly the same order as request.candidates. */
  results: ConfigModuleLoadResult[];
}

/** Go's final effective-ID selection after ignore/ownership resolution. */
export interface ActivateConfigsRequest {
  protocolVersion: typeof CONFIG_DISCOVERY_PROTOCOL_VERSION;
  transactionId: string;
  effectiveConfigIds: string[];
}

/** Minimal activation payload returned across IPC/JSON-RPC to Go. */
export interface ActivateConfigsResponse {
  transactionId: string;
  eslintPluginEntries: ConfigModuleEslintPluginEntry[];
}

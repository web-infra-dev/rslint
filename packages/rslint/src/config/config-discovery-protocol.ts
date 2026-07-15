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
  /** Isolates batches belonging to one discovery transaction. */
  transactionId: string;
  /**
   * `cached` uses Node's ordinary module-import semantics for CLI and native
   * API calls. `fresh` is reserved for editor refresh transactions and
   * cache-busts the entry module; static imports retain Node's normal cache.
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
  | LoadedConfigModuleResult
  | FailedConfigModuleResult;

export interface LoadConfigsResponse {
  transactionId: string;
  /** Results are in exactly the same order as request.candidates. */
  results: ConfigModuleLoadResult[];
}

/** Go's final effective-ID selection after ignore/ownership resolution. */
export interface ActivateConfigsRequest {
  transactionId: string;
  effectiveConfigIds: string[];
}

/** Minimal activation payload returned across IPC/JSON-RPC to Go. */
export interface ActivateConfigsResponse {
  transactionId: string;
  eslintPluginEntries: ConfigModuleEslintPluginEntry[];
}

export interface ConfigPredicateCall {
  /** Opaque, transaction-local correlation ID allocated by Go. */
  callId: string;
  /** Opaque closure occurrence ID allocated while Node normalizes a module. */
  predicateId: string;
  /** Absolute lexical path. Node converts separators to the host-native form. */
  absolutePath: string;
  /** Directory predicates receive a native trailing path separator. */
  directory?: boolean;
}

export interface EvaluateConfigPredicatesRequest {
  transactionId: string;
  calls: ConfigPredicateCall[];
}

export type ConfigPredicateResult =
  | {
      callId: string;
      status: 'evaluated';
      value: boolean;
    }
  | {
      callId: string;
      status: 'failed';
      error: { message: string; code?: string };
    };

export interface EvaluateConfigPredicatesResponse {
  transactionId: string;
  /** Results preserve request order exactly. */
  results: ConfigPredicateResult[];
}

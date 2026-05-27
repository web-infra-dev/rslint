/**
 * Wire-format types shared between Go (`internal/api`) and the runner.
 *
 * The frame layout is `[4 bytes u32 LE length][JSON payload]`. The payload
 * shape is the {@link IpcMessage} below — a {kind, id, data} triple that
 * mirrors `internal/api.Message` byte-for-byte. The two definitions must
 * stay in lockstep; the cross-language IPC tests in
 * `tests/ipc-client.test.ts` lock the contract.
 */

/**
 * Numeric kinds used in {@link IpcMessage.kind}. Defined as a string union
 * so consumers can switch exhaustively. The kinds map directly to Go's
 * `MessageKind` constants (`internal/api/api.go::Kind*`).
 */
export type MessageKind =
  // ── shared infrastructure (also consumed by `--api` mode for wasm/rslint-api) ──
  | 'lint'
  | 'applyFixes'
  | 'getAstInfo'
  | 'response'
  | 'error'
  | 'handshake'
  | 'exit'
  // ── CLI host-process IPC kinds ──
  // (Go child ↔ Node parent over stdio. The LSP path is NOT wired
  //  through these frames — it uses LSP custom requests instead,
  //  routed by the VS Code extension's CompatPool.)
  | 'init'
  | 'lintEslintPlugin'
  | 'cancel'
  | 'output'
  | 'log'
  | 'shutdown';

/**
 * Single IPC frame (JSON-decoded). `id` is 0 for notifications and a
 * positive monotonic integer for requests/responses. `data` is the
 * untyped payload — handlers re-decode into a typed shape as needed.
 */
export interface IpcMessage<T = unknown> {
  kind: MessageKind;
  id: number;
  data?: T;
}

/**
 * Canonical error payload sent in `error` frames. Mirrors Go's
 * `internal/api.ErrorResponse`.
 */
export interface ErrorResponseData {
  message: string;
}

/**
 * Inbound request handler signature. Returning a value resolves the
 * matching `response` frame; throwing surfaces an `error` frame with
 * the thrown error's message.
 */
export type InboundRequestHandler<TIn = unknown, TOut = unknown> = (
  msg: IpcMessage<TIn>,
) => Promise<TOut> | TOut;

/**
 * Inbound notification handler signature. Notifications are id=0 frames
 * that expect no reply; thrown errors are logged and discarded.
 */
export type NotificationHandler<TIn = unknown> = (
  msg: IpcMessage<TIn>,
) => Promise<void> | void;

/**
 * Per-config descriptor handed to the worker pool. Each worker imports
 * every descriptor's `configPath` once at init, then routes per-file
 * lint tasks via `configKey === configDirectory` to the right plugin
 * instances. The `configDirectory` here MUST match the value Go writes
 * into `CompatLintFile.ConfigKey` byte-for-byte; the worker uses it as
 * a Map key for per-file dispatch.
 */
export interface ConfigDescriptor {
  /** Absolute filesystem path of the rslint config file (`rslint.config.{js,mjs,ts,mts}`). */
  configPath: string;
  /** Absolute filesystem path of the directory holding the config file.
   *  Matches the `ConfigKey` Go emits per file during compat dispatch. */
  configDirectory: string;
}

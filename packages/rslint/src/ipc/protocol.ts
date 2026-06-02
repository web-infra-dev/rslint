/**
 * Wire-format protocol types for the Go↔Node IPC transport, shared between
 * the Go side (`internal/ipc.Channel`, and `internal/api` for `--api` mode)
 * and the Node {@link IpcClient}.
 *
 * The frame layout is `[4 bytes u32 LE length][JSON payload]`. The payload is
 * the {@link IpcMessage} `{kind, id, data}` triple, mirroring Go's
 * `ipc.Message` byte-for-byte. The two definitions must stay in lockstep; the
 * cross-language contract tests pin it.
 *
 * This is pure transport protocol — it carries no knowledge of any specific
 * task (lint, …); those live in their own layers.
 */

/**
 * Frame kinds used in {@link IpcMessage.kind}. A string union so consumers can
 * switch exhaustively. These map directly to Go's `MessageKind` constants.
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
  //  through these frames — it uses LSP custom requests instead.)
  | 'init'
  | 'cancel'
  | 'output'
  | 'log'
  | 'shutdown';

/**
 * Single IPC frame (JSON-decoded). `id` is 0 for notifications and a positive
 * monotonic integer for requests/responses. `data` is the untyped payload —
 * handlers re-decode into a typed shape as needed.
 */
export interface IpcMessage<T = unknown> {
  kind: MessageKind;
  id: number;
  data?: T;
}

/**
 * Canonical error payload sent in `error` frames. Mirrors Go's
 * `ipc.ErrorResponseData`.
 */
export interface ErrorResponseData {
  message: string;
}

/**
 * Inbound request handler signature. Returning a value resolves the matching
 * `response` frame; throwing surfaces an `error` frame with the thrown error's
 * message.
 */
export type InboundRequestHandler<TIn = unknown, TOut = unknown> = (
  msg: IpcMessage<TIn>,
) => Promise<TOut> | TOut;

/**
 * Inbound notification handler signature. Notifications are id=0 frames that
 * expect no reply; thrown errors are logged and discarded.
 */
export type NotificationHandler<TIn = unknown> = (
  msg: IpcMessage<TIn>,
) => Promise<void> | void;

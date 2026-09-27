import type { SharedSource } from '../native/binding.js';
import type { WireMessage } from './protocol.generated.js';

/**
 * Handler-facing types for Go↔Node IPC. Go owns the generated wire envelope
 * and framing constants. IpcClient resolves its text attachments and hides
 * storage metadata before delivering a message to an application handler.
 *
 * This is pure transport protocol — it carries no knowledge of any specific
 * task (lint, …); those live in their own layers.
 */

/**
 * Application kinds remain opaque to IPC; infrastructure kinds come from Go.
 */
export type MessageKind = WireMessage['kind'];

/** Complete inline text or a revocable, pointer-free native reader capability. */
export type IpcAttachment = string | SharedSource;

/**
 * Single IPC frame (JSON-decoded). `id` is 0 for notifications and a positive
 * monotonic integer for requests/responses. `data` is the untyped payload —
 * handlers re-decode into a typed shape as needed.
 */
export interface IpcMessage<T = unknown> extends Omit<
  WireMessage,
  'data' | 'attachments' | 'transport'
> {
  data?: T;
  attachments?: readonly IpcAttachment[];
}

/**
 * Canonical error payload sent in `error` frames. Mirrors Go's
 * `ipc.ErrorResponseData`.
 */
export type { ErrorResponseData } from './protocol.generated.js';

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

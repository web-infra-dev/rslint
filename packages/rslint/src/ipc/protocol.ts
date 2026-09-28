import type { ByteInput } from '../native/binding.js';

/**
 * Wire and handler-facing types for Go↔Node IPC. Go provides shared storage
 * configuration at runtime. IpcClient resolves byte attachments and hides
 * storage metadata before delivering a message to an application handler.
 *
 * This is pure transport protocol — it carries no knowledge of any specific
 * task (lint, …); those live in their own layers.
 */

/**
 * Application kinds remain opaque to IPC.
 */
export type MessageKind = WireMessage['kind'];

export interface MemoryMapping {
  version: number;
  fd?: number;
  handle?: string;
  processId?: number;
}

export interface MemoryBatch {
  slot: number;
  generation: number;
  length: number;
}

export interface MemoryRange {
  offset: number;
  length: number;
}

export interface WireAttachment {
  text?: string;
  bytes?: string;
  range?: MemoryRange;
}

export interface TransportMetadata {
  mapping?: MemoryMapping;
  batches?: MemoryBatch[];
  released?: MemoryBatch[];
}

export interface WireMessage {
  kind: string;
  id: number;
  data?: unknown;
  attachments?: WireAttachment[];
  transport?: TransportMetadata;
}

/** Owned inline text/bytes or a request-scoped, pointer-free native capability. */
export type IpcAttachment = ByteInput;

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

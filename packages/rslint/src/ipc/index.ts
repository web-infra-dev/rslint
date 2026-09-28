/**
 * Go↔Node IPC transport, Node side. This is a task-agnostic communication
 * module between the Go binary and the Node host — independent of any specific
 * task. It lives in core because core owns the CLI host that drives the Go
 * child; callers layer their task (currently lint) on top of this transport.
 */
export { IpcClient, encodeFrame, decodeFrame } from './client.js';
export { spawnIpcProcess, type IpcProcessOptions } from './process.js';
export type {
  MessageKind,
  IpcMessage,
  IpcAttachment,
  ErrorResponseData,
  InboundRequestHandler,
  NotificationHandler,
} from './protocol.js';

export { readAttachmentBytes } from './memory-transport.js';

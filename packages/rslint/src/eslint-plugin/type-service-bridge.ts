import type { MessagePort } from 'node:worker_threads';
import {
  BRIDGE_CANCELLED_TASK_IDX,
  BRIDGE_CONTROL_BYTES,
  BRIDGE_STATE_CLOSED,
  BRIDGE_STATE_ERROR,
  BRIDGE_STATE_IDX,
  BRIDGE_WAKE_IDX,
  type OpenTypeServiceBridge,
  type TypeServiceBridgeReplyEnvelope,
  type TypeServiceBridgeRequestEnvelope,
  type TypeServiceWorkerRequest,
  type TypeServiceWorkerResponse,
} from './plugin/type-service-protocol.js';

export type WorkerTypeTransport = (
  request: TypeServiceWorkerRequest,
  signal: AbortSignal,
) => Promise<TypeServiceWorkerResponse>;

export interface BridgeTaskContext {
  readonly taskId: number;
  readonly cancelled: boolean;
  readonly typeSession?: string;
  readonly typeServices?: WorkerTypeTransport;
}

/**
 * Main-thread half of one worker's persistent synchronous type-service channel
 * (paired with `WorkerTypeServices` in `type-service-worker.ts`).
 *
 * Owns the transferred `MessagePort`, the 12-byte control `SharedArrayBuffer`,
 * and per-in-flight-request `AbortController`s so every cancellation or teardown
 * aborts active requests and wakes the blocked worker atomically.
 */
export class HostTypeBridge {
  private closed = false;
  private readonly activeCancels = new Set<AbortController>();

  private constructor(
    private readonly port: MessagePort,
    private readonly control: Int32Array,
  ) {}

  static reject(msg: OpenTypeServiceBridge): void {
    if (
      msg.control instanceof SharedArrayBuffer &&
      msg.control.byteLength >= BRIDGE_CONTROL_BYTES
    ) {
      const control = new Int32Array(msg.control);
      Atomics.store(control, BRIDGE_STATE_IDX, BRIDGE_STATE_CLOSED);
      HostTypeBridge.wake(control);
    }
    msg.port.close();
  }

  static attach(
    msg: OpenTypeServiceBridge,
    resolveTask: (taskId: number) => BridgeTaskContext | undefined,
    onFatalError: (bridge: HostTypeBridge) => void,
  ): HostTypeBridge | undefined {
    if (
      !(msg.control instanceof SharedArrayBuffer) ||
      msg.control.byteLength < BRIDGE_CONTROL_BYTES
    ) {
      msg.port.close();
      return undefined;
    }
    const control = new Int32Array(msg.control);
    const bridge = new HostTypeBridge(msg.port, control);

    msg.port.on('message', (envelope: TypeServiceBridgeRequestEnvelope) => {
      void (async () => {
        if (bridge.closed) return;
        const task = resolveTask(envelope.taskId);
        if (!task || task.cancelled) return;
        const cancel = new AbortController();
        bridge.activeCancels.add(cancel);
        let reply: TypeServiceBridgeReplyEnvelope;
        try {
          if (envelope.request.session !== task.typeSession) {
            throw new Error('type service task has a foreign session');
          }
          if (!task.typeServices) {
            throw new Error('type service transport unavailable');
          }
          reply = {
            reqId: envelope.reqId,
            result: await task.typeServices(envelope.request, cancel.signal),
          };
        } catch (error) {
          reply = {
            reqId: envelope.reqId,
            error: {
              name: error instanceof Error ? error.name : 'Error',
              message: error instanceof Error ? error.message : String(error),
            },
          };
        } finally {
          bridge.activeCancels.delete(cancel);
        }
        if (
          bridge.closed ||
          cancel.signal.aborted ||
          resolveTask(envelope.taskId) !== task ||
          task.cancelled
        ) {
          return;
        }
        try {
          msg.port.postMessage(reply);
          HostTypeBridge.wake(control);
        } catch (postError) {
          try {
            const errorReply: TypeServiceBridgeReplyEnvelope = {
              reqId: envelope.reqId,
              error: {
                name: postError instanceof Error ? postError.name : 'Error',
                message:
                  postError instanceof Error
                    ? postError.message
                    : String(postError),
              },
            };
            msg.port.postMessage(errorReply);
            HostTypeBridge.wake(control);
          } catch {
            onFatalError(bridge);
            bridge.close(BRIDGE_STATE_ERROR);
          }
        }
      })();
    });

    return bridge;
  }

  private static wake(control: Int32Array): void {
    Atomics.add(control, BRIDGE_WAKE_IDX, 1);
    Atomics.notify(control, BRIDGE_WAKE_IDX);
  }

  private abortActive(): void {
    for (const cancel of [...this.activeCancels]) {
      cancel.abort();
    }
    this.activeCancels.clear();
  }

  cancelTask(taskId: number): void {
    if (this.closed) return;
    this.abortActive();
    Atomics.store(this.control, BRIDGE_CANCELLED_TASK_IDX, taskId);
    HostTypeBridge.wake(this.control);
  }

  resetCancelledTask(): void {
    if (this.closed) return;
    Atomics.store(this.control, BRIDGE_CANCELLED_TASK_IDX, 0);
  }

  close(
    state:
      | typeof BRIDGE_STATE_CLOSED
      | typeof BRIDGE_STATE_ERROR = BRIDGE_STATE_CLOSED,
  ): void {
    if (this.closed) return;
    this.closed = true;
    this.abortActive();
    Atomics.store(this.control, BRIDGE_STATE_IDX, state);
    HostTypeBridge.wake(this.control);
    this.port.close();
  }
}

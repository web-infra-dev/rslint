import type { MessagePort } from 'node:worker_threads';
import type {
  FileBinding,
  TypeServiceDescriptor,
} from '../../languages/js/type-service-protocol.js';
import type {
  TypeServiceHostRequest,
  TypeServiceProjectionData,
} from '../../service/plugin-type-service-protocol.js';

export interface PluginFileTypeServices {
  session: string;
  binding: FileBinding;
  syntaxOnly?: boolean;
  emitDecoratorMetadata?: boolean;
  experimentalDecorators?: boolean;
  isolatedDeclarations?: boolean;
}
export type {
  TypeServiceHostRequest,
  TypeServiceHostResponse,
  TypeServiceHostTransport,
} from '../../service/plugin-type-service-protocol.js';

/** 12-byte (3 x Int32) control SAB shared by one worker's persistent bridge:
 *  - [0] wakeEpoch: monotonic counter bumped on every reply, cancel, or close
 *  - [1] cancelledTaskId: taskId of the cancelled task (exact equality match)
 *  - [2] bridgeState: 0 = open, 1 = closed/shutdown, 2 = fatal port error */
export const BRIDGE_CONTROL_BYTES = 12;
export const BRIDGE_WAKE_IDX = 0;
export const BRIDGE_CANCELLED_TASK_IDX = 1;
export const BRIDGE_STATE_IDX = 2;
export const BRIDGE_STATE_OPEN = 0;
export const BRIDGE_STATE_CLOSED = 1;
export const BRIDGE_STATE_ERROR = 2;

export interface OpenTypeServiceBridge {
  kind: 'type-bridge';
  port: MessagePort;
  control: SharedArrayBuffer;
}
export interface TypeServiceBridgeRequestEnvelope {
  taskId: number;
  reqId: number;
  request: TypeServiceWorkerRequest;
}
export interface TypeServiceBridgeReplyEnvelope {
  reqId: number;
  result?: TypeServiceWorkerResponse;
  error?: { name: string; message: string };
}
export type TypeServiceWorkerRequest = TypeServiceHostRequest & {
  syntaxOnly?: boolean;
};
export interface TypeServiceWorkerResponse {
  descriptor?: TypeServiceDescriptor;
  projection?: TypeServiceProjectionData & { tree: Uint8Array };
  packets?: readonly Uint8Array[];
  source?: { fileName: string; path?: string; tree?: Uint8Array };
}

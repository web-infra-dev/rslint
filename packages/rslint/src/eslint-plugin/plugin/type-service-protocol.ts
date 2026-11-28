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

export interface OpenTypeServiceBridge {
  kind: 'type-bridge';
  taskId: number;
  session: string;
  port: MessagePort;
  notification: SharedArrayBuffer;
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

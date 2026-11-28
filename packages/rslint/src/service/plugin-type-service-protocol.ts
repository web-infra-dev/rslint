import type {
  TypeServiceDescriptor,
  TypeServiceProjection,
  TypeServiceQuery,
} from '../languages/js/type-service-protocol.js';

export interface TypeServiceHostRequest {
  session: string;
  generation: number;
  open?: boolean;
  projection?: string;
  queries?: readonly TypeServiceQuery[];
  source?: string;
  metadataOnly?: boolean;
}
export type TypeServiceProjectionData = Omit<
  TypeServiceProjection,
  'sourceFileName' | 'binding' | 'tree'
>;
export interface TypeServiceHostResponse {
  status?: 'closed' | 'cancelled';
  descriptor?: TypeServiceDescriptor;
  projection?: TypeServiceProjectionData & { tree: string | Uint8Array };
  packets?: readonly (string | Uint8Array)[];
  source?: { fileName: string; path?: string; tree?: string | Uint8Array };
}
export type TypeServiceHostTransport = (
  request: TypeServiceHostRequest,
  signal?: AbortSignal,
) => Promise<TypeServiceHostResponse>;

import type { IpcAttachment, IpcMessage } from '../ipc/protocol.js';
import type {
  TypeServiceHostResponse,
  TypeServiceProjectionData,
} from '../service/plugin-type-service-protocol.js';

export interface NativeTypeServiceResponse {
  status?: 'closed' | 'cancelled';
  descriptor?: TypeServiceHostResponse['descriptor'];
  projection?: TypeServiceProjectionData & { treeAttachment: number };
  packets?: readonly string[];
  packetAttachments?: number[];
  source?: { fileName: string; path?: string; treeAttachment?: number };
}

function readBinaryAttachment(
  attachments: readonly IpcAttachment[],
  index: unknown,
): Buffer {
  if (
    typeof index !== 'number' ||
    !Number.isInteger(index) ||
    index < 0 ||
    index >= attachments.length
  ) {
    throw new Error(`invalid type service attachment index: ${String(index)}`);
  }
  const value = attachments[index];
  if (Buffer.isBuffer(value)) return value;
  if (value instanceof Uint8Array) {
    return Buffer.from(value.buffer, value.byteOffset, value.byteLength);
  }
  throw new Error(`type service attachment ${index} is not binary`);
}

export function resolveTypeServiceReply(
  message: IpcMessage<NativeTypeServiceResponse>,
): TypeServiceHostResponse {
  const data = message.data;
  if (data === null || typeof data !== 'object' || Array.isArray(data)) {
    throw new Error('missing type service reply');
  }
  const attachments = message.attachments ?? [];
  const { projection } = data;
  let resolvedProjection: TypeServiceHostResponse['projection'];
  if (projection) {
    const { treeAttachment, ...originTables } = projection;
    resolvedProjection = {
      ...originTables,
      tree: readBinaryAttachment(attachments, treeAttachment),
    };
  }
  let packets: TypeServiceHostResponse['packets'];
  if (data.packetAttachments !== undefined) {
    if (!Array.isArray(data.packetAttachments)) {
      throw new Error('invalid type service packet attachments');
    }
    packets = data.packetAttachments.map((i) =>
      readBinaryAttachment(attachments, i),
    );
  } else if (data.packets !== undefined) {
    if (
      !Array.isArray(data.packets) ||
      !data.packets.every((packet) => typeof packet === 'string')
    ) {
      throw new Error('invalid inline type service packets');
    }
    packets = data.packets;
  }
  return {
    status: data.status,
    descriptor: data.descriptor,
    projection: resolvedProjection,
    packets,
    source: data.source
      ? {
          fileName: data.source.fileName,
          path: data.source.path,
          tree:
            data.source.treeAttachment === undefined
              ? undefined
              : readBinaryAttachment(attachments, data.source.treeAttachment),
        }
      : undefined,
  };
}

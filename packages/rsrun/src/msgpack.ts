export const MESSAGE_TYPE_REQUEST = 1;
export const MESSAGE_TYPE_RESPONSE = 4;
export const MESSAGE_TYPE_ERROR = 5;

const MSGPACK_ARRAY3 = 0x93;
const MSGPACK_U8 = 0xcc;
const MSGPACK_BIN8 = 0xc4;
const MSGPACK_BIN16 = 0xc5;
const MSGPACK_BIN32 = 0xc6;

function encodeMsgpackBin(buffer: Buffer): Buffer {
  const length = buffer.length;
  if (length < 256) {
    return Buffer.from([MSGPACK_BIN8, length]);
  }
  if (length < 1 << 16) {
    const header = Buffer.alloc(3);
    header[0] = MSGPACK_BIN16;
    header.writeUInt16BE(length, 1);
    return header;
  }
  const header = Buffer.alloc(5);
  header[0] = MSGPACK_BIN32;
  header.writeUInt32BE(length, 1);
  return header;
}

export function encodeMessage(
  messageType: number,
  method: string,
  payloadBuffer?: Buffer,
): Buffer {
  const methodBuffer = Buffer.from(method, 'utf8');
  const payload = payloadBuffer ?? Buffer.alloc(0);
  return Buffer.concat([
    Buffer.from([MSGPACK_ARRAY3, MSGPACK_U8, messageType]),
    encodeMsgpackBin(methodBuffer),
    methodBuffer,
    encodeMsgpackBin(payload),
    payload,
  ]);
}

interface DecodeBinResult {
  value: Buffer;
  nextOffset: number;
}

function decodeMsgpackBin(buffer: Buffer, offset: number): DecodeBinResult {
  const tag = buffer[offset];
  let length: number;
  let headerSize: number;
  if (tag === MSGPACK_BIN8) {
    length = buffer[offset + 1];
    headerSize = 2;
  } else if (tag === MSGPACK_BIN16) {
    length = buffer.readUInt16BE(offset + 1);
    headerSize = 3;
  } else if (tag === MSGPACK_BIN32) {
    length = buffer.readUInt32BE(offset + 1);
    headerSize = 5;
  } else {
    throw new Error(`rsrun: invalid msgpack bin tag 0x${tag.toString(16)}`);
  }
  const start = offset + headerSize;
  const end = start + length;
  if (end > buffer.length) {
    throw new Error('rsrun: unexpected end of msgpack payload');
  }
  return { value: buffer.slice(start, end), nextOffset: end };
}

export interface DecodedMessage {
  messageType: number;
  method: string;
  payload: Buffer;
  nextOffset: number;
}

export function decodeMessage(buffer: Buffer, offset: number): DecodedMessage {
  if (buffer[offset] !== MSGPACK_ARRAY3) {
    throw new Error('rsrun: invalid msgpack header');
  }
  if (buffer[offset + 1] !== MSGPACK_U8) {
    throw new Error('rsrun: invalid msgpack message type');
  }
  const messageType = buffer[offset + 2];
  let cursor = offset + 3;
  const methodResult = decodeMsgpackBin(buffer, cursor);
  cursor = methodResult.nextOffset;
  const payloadResult = decodeMsgpackBin(buffer, cursor);
  cursor = payloadResult.nextOffset;
  return {
    messageType,
    method: methodResult.value.toString('utf8'),
    payload: payloadResult.value,
    nextOffset: cursor,
  };
}

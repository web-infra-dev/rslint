/** Decode independently owned Go facts. Shared readers pass immutable views. */
export function decodeTypePacket(bytes: Uint8Array): unknown {
  const invalid = (): never => {
    throw new Error('invalid type service packet');
  };
  if (
    bytes.length < 5 ||
    bytes[0] !== 82 ||
    bytes[1] !== 84 ||
    bytes[2] !== 83 ||
    bytes[3] !== 1
  )
    invalid();
  const view = new DataView(bytes.buffer, bytes.byteOffset, bytes.byteLength);
  let offset = 4;
  const word = (): number => {
    if (offset + 4 > bytes.length) invalid();
    const n = view.getUint32(offset, true);
    offset += 4;
    return n;
  };
  const read = (depth: number): unknown => {
    if (depth > 128 || offset >= bytes.length) invalid();
    switch (bytes[offset++]) {
      case 0:
        return undefined;
      case 1:
        return false;
      case 2:
        return true;
      case 3: {
        const size = word();
        if (offset + size * 2 > bytes.length) invalid();
        const chunks: string[] = [];
        for (let i = 0; i < size; i += 4096) {
          const units: number[] = [];
          for (let j = i; j < Math.min(size, i + 4096); j++)
            units.push(view.getUint16(offset + j * 2, true));
          chunks.push(String.fromCharCode(...units));
        }
        offset += size * 2;
        return chunks.join('');
      }
      case 4: {
        if (offset + 8 > bytes.length) invalid();
        const n = view.getFloat64(offset, true);
        offset += 8;
        return n;
      }
      case 5: {
        const size = word();
        if (size > bytes.length - offset) invalid();
        return Array.from({ length: size }, () => read(depth + 1));
      }
      case 6: {
        const size = word();
        if (size > (bytes.length - offset) / 2) invalid();
        const result: Record<string, unknown> = Object.create(null);
        for (let i = 0; i < size; i++) {
          const key = read(depth + 1);
          if (typeof key !== 'string') return invalid();
          if (Object.hasOwn(result, key)) return invalid();
          result[key] = read(depth + 1);
        }
        return result;
      }
      case 7: {
        const size = word();
        if (offset + size > bytes.length) invalid();
        const data = Uint8Array.from(bytes.subarray(offset, offset + size));
        offset += size;
        return data;
      }
      default:
        return invalid();
    }
  };
  const value = read(0);
  if (offset !== bytes.length) invalid();
  return value;
}

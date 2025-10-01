// Re-export the public TypeScript API
export * from '@typescript/api';

// Minimal RemoteSourceFile used by consumers (tests, playground)
// to decode the source text from the encoded SourceFile buffer
// returned by the Go IPC. This mirrors the upstream format
// without importing deep submodule internals.
export class RemoteSourceFile {
  text: string;
  constructor(data: Uint8Array, decoder: TextDecoder) {
    const view = new DataView(data.buffer, data.byteOffset, data.byteLength);
    // Header offsets
    const HEADER_OFFSET_STRING_TABLE_OFFSETS = 4;
    const HEADER_OFFSET_STRING_TABLE = 8;
    const HEADER_OFFSET_EXTENDED_DATA = 12;
    const HEADER_OFFSET_NODES = 16;
    const NODE_OFFSET_DATA = 20;
    const NODE_LEN = 24;
    const NODE_EXTENDED_DATA_MASK = 0x00_ff_ff_ff;

    const offsetStringTableOffsets = view.getUint32(
      HEADER_OFFSET_STRING_TABLE_OFFSETS,
      true,
    );
    const offsetStringTable = view.getUint32(HEADER_OFFSET_STRING_TABLE, true);
    const offsetExtendedData = view.getUint32(
      HEADER_OFFSET_EXTENDED_DATA,
      true,
    );
    const offsetNodes = view.getUint32(HEADER_OFFSET_NODES, true);

    // SourceFile node is at index 1
    const index = 1;
    const byteIndex = offsetNodes + index * NODE_LEN;
    const dataField = view.getUint32(byteIndex + NODE_OFFSET_DATA, true);

    // For SourceFile, first dword at extended data is the text string index
    const extendedDataOffset =
      offsetExtendedData + (dataField & NODE_EXTENDED_DATA_MASK);
    const stringIndex = view.getUint32(extendedDataOffset, true);

    const start = view.getUint32(
      offsetStringTableOffsets + stringIndex * 4,
      true,
    );
    const end = view.getUint32(
      offsetStringTableOffsets + (stringIndex + 1) * 4,
      true,
    );

    const textBytes = new Uint8Array(
      view.buffer,
      offsetStringTable + start,
      end - start,
    );
    this.text = decoder.decode(textBytes);
  }
}

// Provide a value export for `Node` to satisfy consumers that import it as a value
// in JS. At runtime they don't use it; type information remains available via
// the export * above.
export const Node: unknown = undefined;

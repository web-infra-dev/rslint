/**
 * Compact binary request encoder for the direct shared-memory RPC fast path.
 * Mirrors `internal/typeservice/binary_request.go` byte-for-byte.
 *
 * Writes directly into a reusable worker-owned `Uint8Array` scratch buffer so
 * hot-path checker queries perform zero intermediate buffer allocations.
 * Returns `0` for operations excluded from SHM (`ProgramRequest`,
 * `suggestionDiagnostics`) or requests exceeding `dst.byteLength`, signaling
 * the caller to route transparently through `HostTypeBridge`.
 */
import type {
  NodeRef,
  Operation,
  TypeServiceQuery,
} from './type-service-protocol.js';

const FLAG_SKIP_RECORDS = 0x01;
const INLINE_FILE_SENTINEL = 0xffff_ffff;
const NAME_UTF16_FLAG = 0x8000_0000;
const OP_RAW_JSON = 255;

const EXCLUDED_OPS = new Set<Operation>([
  'rootFileNames',
  'configFileParsingDiagnostics',
  'optionsDiagnostics',
  'globalDiagnostics',
  'syntacticDiagnostics',
  'semanticDiagnostics',
  'declarationDiagnostics',
  'suggestionDiagnostics',
]);

const CHECKER_OP_CODES: Partial<Record<Operation, number>> = {
  errorType: 1,
  anyType: 2,
  unknownType: 3,
  stringType: 4,
  numberType: 5,
  booleanType: 6,
  bigIntType: 7,
  eSSymbolType: 8,
  nonPrimitiveType: 9,
  neverType: 10,
  nullType: 11,
  undefinedType: 12,
  trueType: 13,
  falseType: 14,
  void: 15,
  ambientModules: 16,
};

const NODE_OP_CODES: Partial<Record<Operation, number>> = {
  type: 20,
  fromTypeNode: 21,
  context: 22,
  symbol: 23,
  declarationSymbol: 24,
  signature: 25,
  nodeRecord: 26,
  shorthandValueSymbol: 27,
  exportSpecifierLocalTargetSymbol: 28,
  propertySymbolOfDestructuringAssignment: 29,
  jsxIntrinsicTagNamesAt: 30,
  constantValue: 31,
  isOptionalParameter: 32,
  isImplementationOfOverload: 33,
};

const SINGLE_ID_OP_CODES: Partial<Record<Operation, number>> = {
  // TypeRequest (40..61)
  constraint: 40,
  apparent: 41,
  baseLiteral: 42,
  nonNullable: 43,
  properties: 44,
  augmentedPropertiesOfType: 45,
  typeArguments: 46,
  isArray: 47,
  isArrayLike: 48,
  isTuple: 49,
  numberIndexType: 50,
  stringIndexType: 51,
  typeRecord: 52,
  typeDetails: 53,
  declaredIndexInfos: 54,
  awaited: 55,
  widened: 56,
  baseTypes: 57,
  defaultType: 58,
  indexInfos: 59,
  mappedFields: 60,
  classThisType: 61,
  // SymbolRequest (65..79)
  rootSymbols: 65,
  mergedSymbol: 66,
  isArgumentsSymbol: 67,
  isUndefinedSymbol: 68,
  aliasedSymbol: 69,
  immediateAliasedSymbol: 70,
  isUnknownSymbol: 71,
  exportsOfModule: 72,
  exportSymbol: 73,
  fullyQualifiedName: 74,
  symbolMembers: 75,
  symbolExports: 76,
  declaredType: 77,
  typeOfSymbol: 78,
  symbolRecord: 79,
  // SignatureRequest (80..84)
  returnType: 80,
  signatureRecord: 81,
  typePredicate: 82,
  signatureDocumentation: 83,
  signatureJSDocTags: 84,
};

const PRINT_OP_CODES: Partial<Record<Operation, number>> = {
  typeString: 120,
  signatureString: 121,
  symbolString: 122,
  typeNode: 123,
};

const encoder = new TextEncoder();

function isU32(value: number): boolean {
  return Number.isInteger(value) && value >= 0 && value <= 0xffff_ffff;
}

function isI32(value: number): boolean {
  return (
    Number.isInteger(value) && value >= -0x8000_0000 && value <= 0x7fff_ffff
  );
}

class BinaryWriter {
  pos = 0;
  readonly view: DataView;

  constructor(private readonly buf: Uint8Array) {
    this.view = new DataView(buf.buffer, buf.byteOffset, buf.byteLength);
  }

  private ensure(n: number): boolean {
    return this.pos + n <= this.buf.byteLength;
  }

  u8(v: number): boolean {
    if (!this.ensure(1)) return false;
    this.buf[this.pos++] = v & 0xff;
    return true;
  }

  u16(v: number): boolean {
    if (!this.ensure(2)) return false;
    this.view.setUint16(this.pos, v, true);
    this.pos += 2;
    return true;
  }

  u32(v: number): boolean {
    if (!isU32(v) || !this.ensure(4)) return false;
    this.view.setUint32(this.pos, v, true);
    this.pos += 4;
    return true;
  }

  i32(v: number): boolean {
    if (!isI32(v) || !this.ensure(4)) return false;
    this.view.setInt32(this.pos, v, true);
    this.pos += 4;
    return true;
  }

  asciiOrUtf8Bytes(str: string): number {
    // Fast ASCII copy without TextEncoder allocation for session IDs & paths.
    const start = this.pos;
    let ascii = true;
    for (let i = 0; i < str.length; i++) {
      if (str.charCodeAt(i) >= 0x80) {
        ascii = false;
        break;
      }
    }
    if (ascii) {
      if (!this.ensure(str.length)) return -1;
      for (let i = 0; i < str.length; i++) {
        this.buf[this.pos++] = str.charCodeAt(i);
      }
      return this.pos - start;
    }
    const encoded = encoder.encode(str);
    if (!this.ensure(encoded.byteLength)) return -1;
    this.buf.set(encoded, this.pos);
    this.pos += encoded.byteLength;
    return encoded.byteLength;
  }

  nodeRef(node: NodeRef, fileIndex: ReadonlyMap<string, number>): boolean {
    const idx = fileIndex.get(node.file);
    if (idx !== undefined) {
      if (!this.u32(idx)) return false;
    } else {
      if (!this.u32(INLINE_FILE_SENTINEL)) return false;
      const lenPos = this.pos;
      if (!this.ensure(4)) return false;
      this.pos += 4;
      const written = this.asciiOrUtf8Bytes(node.file);
      if (written < 0) return false;
      this.view.setUint32(lenPos, written, true);
    }
    if (!this.u32(node.index)) return false;
    if (node.token === undefined) {
      return this.u8(0);
    }
    return (
      this.u8(1) &&
      this.i32(node.token.kind) &&
      this.i32(node.token.pos) &&
      this.i32(node.token.end)
    );
  }

  name(units: readonly number[]): boolean {
    const len = units.length;
    if (len > 0x7fff_ffff) return false;
    let ascii = true;
    for (let i = 0; i < len; i++) {
      const u = units[i];
      if (!Number.isInteger(u) || u < 0 || u > 0xffff) return false;
      if (u >= 0x80) ascii = false;
    }
    if (ascii) {
      if (!this.ensure(4 + len)) return false;
      this.view.setUint32(this.pos, len, true);
      this.pos += 4;
      for (let i = 0; i < len; i++) {
        this.buf[this.pos++] = units[i];
      }
      return true;
    }
    if (!this.ensure(4 + len * 2)) return false;
    this.view.setUint32(this.pos, (NAME_UTF16_FLAG | len) >>> 0, true);
    this.pos += 4;
    for (let i = 0; i < len; i++) {
      this.view.setUint16(this.pos, units[i], true);
      this.pos += 2;
    }
    return true;
  }
}

/**
 * Encode one semantic `TypeServiceQuery` into `dst`.
 * Returns the written byte length (> 0) on success, or `0` when the query
 * must fall back to `HostTypeBridge` (excluded op, overflow, or out-of-range int).
 */
export function encodeBinaryTypeQuery(
  dst: Uint8Array,
  session: string,
  query: TypeServiceQuery,
  fileIndex: ReadonlyMap<string, number>,
): number {
  if (EXCLUDED_OPS.has(query.op)) return 0;

  const w = new BinaryWriter(dst);
  // Header: [sessionLen: u16][sessionBytes][generation: u32][flags: u8][opCode: u8]
  if (session.length > 0xffff || !w.u16(0)) return 0;
  const sessionBytes = w.asciiOrUtf8Bytes(session);
  if (sessionBytes <= 0 || sessionBytes > 0xffff) return 0;
  w.view.setUint16(0, sessionBytes, true);
  if (!w.u32(query.generation)) return 0;
  if (!w.u8(query.skipRecords ? FLAG_SKIP_RECORDS : 0)) return 0;

  const checkerCode = CHECKER_OP_CODES[query.op];
  if (checkerCode !== undefined) {
    return w.u8(checkerCode) ? w.pos : 0;
  }

  switch (query.op) {
    case 'type':
    case 'fromTypeNode':
    case 'context':
    case 'symbol':
    case 'declarationSymbol':
    case 'signature':
    case 'nodeRecord':
    case 'shorthandValueSymbol':
    case 'exportSpecifierLocalTargetSymbol':
    case 'propertySymbolOfDestructuringAssignment':
    case 'jsxIntrinsicTagNamesAt':
    case 'constantValue':
    case 'isOptionalParameter':
    case 'isImplementationOfOverload': {
      const code = NODE_OP_CODES[query.op];
      return code !== undefined &&
        w.u8(code) &&
        w.nodeRef(query.args.node, fileIndex)
        ? w.pos
        : 0;
    }
    case 'constraint':
    case 'apparent':
    case 'baseLiteral':
    case 'nonNullable':
    case 'properties':
    case 'augmentedPropertiesOfType':
    case 'typeArguments':
    case 'isArray':
    case 'isArrayLike':
    case 'isTuple':
    case 'numberIndexType':
    case 'stringIndexType':
    case 'typeRecord':
    case 'typeDetails':
    case 'declaredIndexInfos':
    case 'awaited':
    case 'widened':
    case 'baseTypes':
    case 'defaultType':
    case 'indexInfos':
    case 'mappedFields':
    case 'classThisType':
    case 'rootSymbols':
    case 'mergedSymbol':
    case 'isArgumentsSymbol':
    case 'isUndefinedSymbol':
    case 'aliasedSymbol':
    case 'immediateAliasedSymbol':
    case 'isUnknownSymbol':
    case 'exportsOfModule':
    case 'exportSymbol':
    case 'fullyQualifiedName':
    case 'symbolMembers':
    case 'symbolExports':
    case 'declaredType':
    case 'typeOfSymbol':
    case 'symbolRecord':
    case 'returnType':
    case 'signatureRecord':
    case 'typePredicate':
    case 'signatureDocumentation':
    case 'signatureJSDocTags': {
      const code = SINGLE_ID_OP_CODES[query.op];
      return code !== undefined && w.u8(code) && w.u32(query.args.id)
        ? w.pos
        : 0;
    }
    case 'typeString':
    case 'typeNode':
    case 'signatureString':
    case 'symbolString': {
      const code = PRINT_OP_CODES[query.op];
      if (code === undefined) return 0;
      const { id, node, flags } = query.args;
      const kind =
        query.op === 'signatureString' || query.op === 'symbolString'
          ? query.args.kind
          : 0;
      const presence =
        (node !== undefined ? 1 : 0) | (flags !== undefined ? 2 : 0);
      if (!w.u8(code) || !w.u32(id) || !w.u8(presence)) return 0;
      if (!w.u32(kind)) return 0;
      if (flags !== undefined && !w.u32(flags)) return 0;
      if (node !== undefined && !w.nodeRef(node, fileIndex)) return 0;
      return w.pos;
    }
    case 'isTypeAssignableTo':
      return w.u8(90) && w.u32(query.args.source) && w.u32(query.args.target)
        ? w.pos
        : 0;
    case 'signatures':
      return w.u8(91) && w.u32(query.args.id) && w.u32(query.args.kind)
        ? w.pos
        : 0;
    case 'indexInfo':
      return w.u8(92) && w.u32(query.args.id) && w.u32(query.args.kind)
        ? w.pos
        : 0;
    case 'nullableType':
      return w.u8(93) && w.u32(query.args.id) && w.u32(query.args.flags)
        ? w.pos
        : 0;
    case 'parameterType':
      return w.u8(94) && w.u32(query.args.id) && w.i32(query.args.position)
        ? w.pos
        : 0;
    case 'symbolDocumentation':
      return w.u8(95) &&
        w.u32(query.args.id) &&
        w.u8(query.args.withChecker ? 1 : 0)
        ? w.pos
        : 0;
    case 'symbolJSDocTags':
      return w.u8(96) &&
        w.u32(query.args.id) &&
        w.u8(query.args.withChecker ? 1 : 0)
        ? w.pos
        : 0;
    case 'property':
      return w.u8(100) && w.u32(query.args.id) && w.name(query.args.name)
        ? w.pos
        : 0;
    case 'propertyType':
      return w.u8(101) && w.u32(query.args.id) && w.name(query.args.name)
        ? w.pos
        : 0;
    case 'moduleMember':
      return w.u8(102) && w.u32(query.args.id) && w.name(query.args.name)
        ? w.pos
        : 0;
    case 'symbolType':
      return w.u8(110) &&
        w.u32(query.args.id) &&
        w.nodeRef(query.args.node, fileIndex)
        ? w.pos
        : 0;
    case 'contextualArgument':
      return w.u8(111) &&
        w.nodeRef(query.args.node, fileIndex) &&
        w.i32(query.args.index)
        ? w.pos
        : 0;
    case 'resolvedSignature': {
      const optFlags =
        (query.args.withCandidates ? 1 : 0) |
        (query.args.argumentCount !== undefined ? 2 : 0);
      if (
        !w.u8(112) ||
        !w.nodeRef(query.args.node, fileIndex) ||
        !w.u8(optFlags)
      ) {
        return 0;
      }
      if (
        query.args.argumentCount !== undefined &&
        !w.i32(query.args.argumentCount)
      ) {
        return 0;
      }
      return w.pos;
    }
    case 'symbolsInScope':
      return w.u8(113) &&
        w.nodeRef(query.args.node, fileIndex) &&
        w.i32(query.args.meaning)
        ? w.pos
        : 0;
    case 'validPropertyAccess':
      return w.u8(114) &&
        w.nodeRef(query.args.node, fileIndex) &&
        w.name(query.args.name)
        ? w.pos
        : 0;
    case 'parameterPropertySymbols':
      return w.u8(115) &&
        w.nodeRef(query.args.node, fileIndex) &&
        w.name(query.args.name)
        ? w.pos
        : 0;
    default: {
      // OP_RAW_JSON escape hatch for rare ops (stringLiteralType, numberLiteralType,
      // bigIntLiteralType, typePredicateString, resolveName).
      if (!w.u8(OP_RAW_JSON)) return 0;
      const raw = JSON.stringify(query);
      const written = w.asciiOrUtf8Bytes(raw);
      return written > 0 ? w.pos : 0;
    }
  }
}

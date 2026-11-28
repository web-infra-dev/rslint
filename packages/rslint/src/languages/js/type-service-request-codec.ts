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
import type { NodeRef, TypeServiceQuery } from './type-service-protocol.js';

const FLAG_SKIP_RECORDS = 0x01;
const INLINE_FILE_SENTINEL = 0xffff_ffff;
const NAME_UTF16_FLAG = 0x8000_0000;
const OP_RAW_JSON = 255;

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
  const w = new BinaryWriter(dst);
  // Header: [sessionLen: u16][sessionBytes][generation: u32][flags: u8][opCode: u8]
  if (session.length > 0xffff || !w.u16(0)) return 0;
  const sessionBytes = w.asciiOrUtf8Bytes(session);
  if (sessionBytes <= 0 || sessionBytes > 0xffff) return 0;
  w.view.setUint16(0, sessionBytes, true);
  if (!w.u32(query.generation)) return 0;
  if (!w.u8(query.skipRecords ? FLAG_SKIP_RECORDS : 0)) return 0;

  const op = (code: number): number => (w.u8(code) ? w.pos : 0);
  const nodeOp = (code: number, node: NodeRef): number =>
    w.u8(code) && w.nodeRef(node, fileIndex) ? w.pos : 0;
  const idOp = (code: number, id: number): number =>
    w.u8(code) && w.u32(id) ? w.pos : 0;
  const printOp = (
    code: number,
    args: {
      readonly id: number;
      readonly node: NodeRef | undefined;
      readonly flags: number | undefined;
    },
    kind: number,
  ): number => {
    const presence =
      (args.node !== undefined ? 1 : 0) | (args.flags !== undefined ? 2 : 0);
    if (!w.u8(code) || !w.u32(args.id) || !w.u8(presence) || !w.u32(kind)) {
      return 0;
    }
    if (args.flags !== undefined && !w.u32(args.flags)) return 0;
    if (args.node !== undefined && !w.nodeRef(args.node, fileIndex)) return 0;
    return w.pos;
  };

  switch (query.op) {
    // Excluded from SHM (always routed through HostTypeBridge).
    case 'rootFileNames':
    case 'configFileParsingDiagnostics':
    case 'optionsDiagnostics':
    case 'globalDiagnostics':
    case 'syntacticDiagnostics':
    case 'semanticDiagnostics':
    case 'declarationDiagnostics':
    case 'suggestionDiagnostics':
      return 0;

    // CheckerRequest (1..16)
    case 'errorType':
      return op(1);
    case 'anyType':
      return op(2);
    case 'unknownType':
      return op(3);
    case 'stringType':
      return op(4);
    case 'numberType':
      return op(5);
    case 'booleanType':
      return op(6);
    case 'bigIntType':
      return op(7);
    case 'eSSymbolType':
      return op(8);
    case 'nonPrimitiveType':
      return op(9);
    case 'neverType':
      return op(10);
    case 'nullType':
      return op(11);
    case 'undefinedType':
      return op(12);
    case 'trueType':
      return op(13);
    case 'falseType':
      return op(14);
    case 'void':
      return op(15);
    case 'ambientModules':
      return op(16);

    // NodeRequest (20..33)
    case 'type':
      return nodeOp(20, query.args.node);
    case 'fromTypeNode':
      return nodeOp(21, query.args.node);
    case 'context':
      return nodeOp(22, query.args.node);
    case 'symbol':
      return nodeOp(23, query.args.node);
    case 'declarationSymbol':
      return nodeOp(24, query.args.node);
    case 'signature':
      return nodeOp(25, query.args.node);
    case 'nodeRecord':
      return nodeOp(26, query.args.node);
    case 'shorthandValueSymbol':
      return nodeOp(27, query.args.node);
    case 'exportSpecifierLocalTargetSymbol':
      return nodeOp(28, query.args.node);
    case 'propertySymbolOfDestructuringAssignment':
      return nodeOp(29, query.args.node);
    case 'jsxIntrinsicTagNamesAt':
      return nodeOp(30, query.args.node);
    case 'constantValue':
      return nodeOp(31, query.args.node);
    case 'isOptionalParameter':
      return nodeOp(32, query.args.node);
    case 'isImplementationOfOverload':
      return nodeOp(33, query.args.node);

    // TypeRequest (40..61)
    case 'constraint':
      return idOp(40, query.args.id);
    case 'apparent':
      return idOp(41, query.args.id);
    case 'baseLiteral':
      return idOp(42, query.args.id);
    case 'nonNullable':
      return idOp(43, query.args.id);
    case 'properties':
      return idOp(44, query.args.id);
    case 'augmentedPropertiesOfType':
      return idOp(45, query.args.id);
    case 'typeArguments':
      return idOp(46, query.args.id);
    case 'isArray':
      return idOp(47, query.args.id);
    case 'isArrayLike':
      return idOp(48, query.args.id);
    case 'isTuple':
      return idOp(49, query.args.id);
    case 'numberIndexType':
      return idOp(50, query.args.id);
    case 'stringIndexType':
      return idOp(51, query.args.id);
    case 'typeRecord':
      return idOp(52, query.args.id);
    case 'typeDetails':
      return idOp(53, query.args.id);
    case 'declaredIndexInfos':
      return idOp(54, query.args.id);
    case 'awaited':
      return idOp(55, query.args.id);
    case 'widened':
      return idOp(56, query.args.id);
    case 'baseTypes':
      return idOp(57, query.args.id);
    case 'defaultType':
      return idOp(58, query.args.id);
    case 'indexInfos':
      return idOp(59, query.args.id);
    case 'mappedFields':
      return idOp(60, query.args.id);
    case 'classThisType':
      return idOp(61, query.args.id);

    // SymbolRequest (65..79)
    case 'rootSymbols':
      return idOp(65, query.args.id);
    case 'mergedSymbol':
      return idOp(66, query.args.id);
    case 'isArgumentsSymbol':
      return idOp(67, query.args.id);
    case 'isUndefinedSymbol':
      return idOp(68, query.args.id);
    case 'aliasedSymbol':
      return idOp(69, query.args.id);
    case 'immediateAliasedSymbol':
      return idOp(70, query.args.id);
    case 'isUnknownSymbol':
      return idOp(71, query.args.id);
    case 'exportsOfModule':
      return idOp(72, query.args.id);
    case 'exportSymbol':
      return idOp(73, query.args.id);
    case 'fullyQualifiedName':
      return idOp(74, query.args.id);
    case 'symbolMembers':
      return idOp(75, query.args.id);
    case 'symbolExports':
      return idOp(76, query.args.id);
    case 'declaredType':
      return idOp(77, query.args.id);
    case 'typeOfSymbol':
      return idOp(78, query.args.id);
    case 'symbolRecord':
      return idOp(79, query.args.id);

    // SignatureRequest (80..84)
    case 'returnType':
      return idOp(80, query.args.id);
    case 'signatureRecord':
      return idOp(81, query.args.id);
    case 'typePredicate':
      return idOp(82, query.args.id);
    case 'signatureDocumentation':
      return idOp(83, query.args.id);
    case 'signatureJSDocTags':
      return idOp(84, query.args.id);

    // Multi-argument ops (90..115)
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

    // PrintRequest (120..123)
    case 'typeString':
      return printOp(120, query.args, 0);
    case 'signatureString':
      return printOp(121, query.args, query.args.kind);
    case 'symbolString':
      return printOp(122, query.args, query.args.kind);
    case 'typeNode':
      return printOp(123, query.args, 0);

    // OP_RAW_JSON escape hatch for rare ops.
    case 'stringLiteralType':
    case 'numberLiteralType':
    case 'bigIntLiteralType':
    case 'typePredicateString':
    case 'resolveName': {
      if (!w.u8(OP_RAW_JSON)) return 0;
      const raw = JSON.stringify(query);
      const written = w.asciiOrUtf8Bytes(raw);
      return written > 0 ? w.pos : 0;
    }

    default: {
      const _exhaustive: never = query;
      return _exhaustive;
    }
  }
}

package typeservice

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

const (
	// BinaryInlineFileSentinel marks a NodeRef carrying an inline UTF-8 file
	// path instead of an index into Session.sourceFiles.
	BinaryInlineFileSentinel uint32 = 0xffff_ffff

	binaryFlagSkipRecords byte   = 0x01
	binaryNameUTF16Flag   uint32 = 0x8000_0000
	binaryOpRawJSON       byte   = 255
)

var (
	binaryCheckerOps = map[byte]Operation{
		1:  "errorType",
		2:  "anyType",
		3:  "unknownType",
		4:  "stringType",
		5:  "numberType",
		6:  "booleanType",
		7:  "bigIntType",
		8:  "eSSymbolType",
		9:  "nonPrimitiveType",
		10: "neverType",
		11: "nullType",
		12: "undefinedType",
		13: "trueType",
		14: "falseType",
		15: "void",
		16: "ambientModules",
	}

	binaryNodeOps = map[byte]Operation{
		20: "type",
		21: "fromTypeNode",
		22: "context",
		23: "symbol",
		24: "declarationSymbol",
		25: "signature",
		26: "nodeRecord",
		27: "shorthandValueSymbol",
		28: "exportSpecifierLocalTargetSymbol",
		29: "propertySymbolOfDestructuringAssignment",
		30: "jsxIntrinsicTagNamesAt",
		31: "constantValue",
		32: "isOptionalParameter",
		33: "isImplementationOfOverload",
	}

	binaryTypeIDOps = map[byte]Operation{
		40: "constraint",
		41: "apparent",
		42: "baseLiteral",
		43: "nonNullable",
		44: "properties",
		45: "augmentedPropertiesOfType",
		46: "typeArguments",
		47: "isArray",
		48: "isArrayLike",
		49: "isTuple",
		50: "numberIndexType",
		51: "stringIndexType",
		52: "typeRecord",
		53: "typeDetails",
		54: "declaredIndexInfos",
		55: "awaited",
		56: "widened",
		57: "baseTypes",
		58: "defaultType",
		59: "indexInfos",
		60: "mappedFields",
		61: "classThisType",
	}

	binarySymbolIDOps = map[byte]Operation{
		65: "rootSymbols",
		66: "mergedSymbol",
		67: "isArgumentsSymbol",
		68: "isUndefinedSymbol",
		69: "aliasedSymbol",
		70: "immediateAliasedSymbol",
		71: "isUnknownSymbol",
		72: "exportsOfModule",
		73: "exportSymbol",
		74: "fullyQualifiedName",
		75: "symbolMembers",
		76: "symbolExports",
		77: "declaredType",
		78: "typeOfSymbol",
		79: "symbolRecord",
	}

	binarySignatureIDOps = map[byte]Operation{
		80: "returnType",
		81: "signatureRecord",
		82: "typePredicate",
		83: "signatureDocumentation",
		84: "signatureJSDocTags",
	}

	binaryPrintOps = map[byte]Operation{
		120: "typeString",
		121: "signatureString",
		122: "symbolString",
		123: "typeNode",
	}
)

type binaryReader struct {
	data []byte
	pos  int
}

func (r *binaryReader) remaining() int { return len(r.data) - r.pos }

func (r *binaryReader) u8() (byte, error) {
	if r.remaining() < 1 {
		return 0, errors.New("truncated binary type service request")
	}
	v := r.data[r.pos]
	r.pos++
	return v, nil
}

func (r *binaryReader) u16() (uint16, error) {
	if r.remaining() < 2 {
		return 0, errors.New("truncated binary type service request")
	}
	v := binary.LittleEndian.Uint16(r.data[r.pos : r.pos+2])
	r.pos += 2
	return v, nil
}

func (r *binaryReader) u32() (uint32, error) {
	if r.remaining() < 4 {
		return 0, errors.New("truncated binary type service request")
	}
	v := binary.LittleEndian.Uint32(r.data[r.pos : r.pos+4])
	r.pos += 4
	return v, nil
}

func (r *binaryReader) i32() (int32, error) {
	v, err := r.u32()
	return int32(v), err
}

func (r *binaryReader) slice(n int) ([]byte, error) {
	if n < 0 || r.remaining() < n {
		return nil, errors.New("truncated binary type service request")
	}
	b := r.data[r.pos : r.pos+n]
	r.pos += n
	return b, nil
}

func (r *binaryReader) nodeRef() (NodeRef, uint32, error) {
	fileIdx, err := r.u32()
	if err != nil {
		return NodeRef{}, BinaryInlineFileSentinel, err
	}
	var file string
	if fileIdx == BinaryInlineFileSentinel {
		pathLen, err := r.u32()
		if err != nil {
			return NodeRef{}, BinaryInlineFileSentinel, err
		}
		raw, err := r.slice(int(pathLen))
		if err != nil {
			return NodeRef{}, BinaryInlineFileSentinel, err
		}
		file = string(raw)
	}
	index, err := r.u32()
	if err != nil {
		return NodeRef{}, BinaryInlineFileSentinel, err
	}
	hasToken, err := r.u8()
	if err != nil {
		return NodeRef{}, BinaryInlineFileSentinel, err
	}
	var token *SyntaxToken
	switch hasToken {
	case 0:
	case 1:
		kind, err := r.i32()
		if err != nil {
			return NodeRef{}, BinaryInlineFileSentinel, err
		}
		if kind < math.MinInt16 || kind > math.MaxInt16 {
			return NodeRef{}, BinaryInlineFileSentinel, errors.New("invalid binary syntax token kind")
		}
		pos, err := r.i32()
		if err != nil {
			return NodeRef{}, BinaryInlineFileSentinel, err
		}
		end, err := r.i32()
		if err != nil {
			return NodeRef{}, BinaryInlineFileSentinel, err
		}
		token = &SyntaxToken{Kind: ast.Kind(kind), Pos: int(pos), End: int(end)}
	default:
		return NodeRef{}, BinaryInlineFileSentinel, errors.New("invalid binary node token flag")
	}
	return NodeRef{File: file, Index: index, Token: token}, fileIdx, nil
}

func (r *binaryReader) name() (Name, error) {
	header, err := r.u32()
	if err != nil {
		return "", err
	}
	if header&binaryNameUTF16Flag == 0 {
		raw, err := r.slice(int(header))
		if err != nil {
			return "", err
		}
		for _, b := range raw {
			if b >= 0x80 {
				return "", errors.New("invalid non-ASCII byte in ASCII binary name")
			}
		}
		return Name(string(raw)), nil
	}
	count := int(header &^ binaryNameUTF16Flag)
	if count > r.remaining()/2 {
		return "", errors.New("truncated UTF-16 binary name")
	}
	raw, err := r.slice(count * 2)
	if err != nil {
		return "", err
	}
	units := make([]uint16, count)
	for i := range count {
		units[i] = binary.LittleEndian.Uint16(raw[i*2 : i*2+2])
	}
	return Name(ecmascript.StringFromCodeUnits(units)), nil
}

// DecodedBinaryQuery holds the parsed header and semantic Query for one
// shared-memory RPC request.
type DecodedBinaryQuery struct {
	Session    string
	Generation uint32
	Query      Query
	FileIdx    uint32
}

// DecodeBinaryQuery parses one binary request produced by
// `packages/rslint/src/languages/js/type-service-request-codec.ts`.
func DecodeBinaryQuery(payload []byte) (DecodedBinaryQuery, error) {
	r := binaryReader{data: payload}
	sessionLen, err := r.u16()
	if err != nil || sessionLen == 0 {
		return DecodedBinaryQuery{}, errors.New("invalid binary type service session header")
	}
	sessionBytes, err := r.slice(int(sessionLen))
	if err != nil {
		return DecodedBinaryQuery{}, err
	}
	session := string(sessionBytes)
	generation, err := r.u32()
	if err != nil {
		return DecodedBinaryQuery{}, err
	}
	flags, err := r.u8()
	if err != nil {
		return DecodedBinaryQuery{}, err
	}
	if flags&^binaryFlagSkipRecords != 0 {
		return DecodedBinaryQuery{}, errors.New("unsupported binary type service flags")
	}
	skipRecords := flags&binaryFlagSkipRecords != 0
	opCode, err := r.u8()
	if err != nil {
		return DecodedBinaryQuery{}, err
	}

	out := DecodedBinaryQuery{
		Session:    session,
		Generation: generation,
		Query: Query{
			Scope:       Scope{Generation: generation},
			SkipRecords: skipRecords,
		},
		FileIdx: BinaryInlineFileSentinel,
	}

	if opCode == binaryOpRawJSON {
		raw := r.data[r.pos:]
		var q Query
		if err := json.Unmarshal(raw, &q); err != nil {
			return DecodedBinaryQuery{}, err
		}
		if q.Generation != generation {
			return DecodedBinaryQuery{}, errors.New("binary type service generation mismatch")
		}
		if q.SkipRecords != skipRecords {
			return DecodedBinaryQuery{}, errors.New("binary type service skipRecords mismatch")
		}
		if _, isProgram := q.Request.(*ProgramRequest); isProgram {
			return DecodedBinaryQuery{}, errors.New("program queries are not supported on binary RPC")
		}
		if nr, ok := q.Request.(*NodeRequest); ok && nr.Op == "suggestionDiagnostics" {
			return DecodedBinaryQuery{}, errors.New("suggestionDiagnostics is not supported on binary RPC")
		}
		out.Query = q
		return out, nil
	}

	var req Request
	switch {
	case binaryCheckerOps[opCode] != "":
		req = &CheckerRequest{Op: binaryCheckerOps[opCode]}
	case binaryNodeOps[opCode] != "":
		node, fileIdx, err := r.nodeRef()
		if err != nil {
			return DecodedBinaryQuery{}, err
		}
		out.FileIdx = fileIdx
		req = &NodeRequest{Op: binaryNodeOps[opCode], Node: node}
	case binaryTypeIDOps[opCode] != "":
		id, err := r.u32()
		if err != nil {
			return DecodedBinaryQuery{}, err
		}
		req = &TypeRequest{Op: binaryTypeIDOps[opCode], ID: id}
	case binarySymbolIDOps[opCode] != "":
		id, err := r.u32()
		if err != nil {
			return DecodedBinaryQuery{}, err
		}
		req = &SymbolRequest{Op: binarySymbolIDOps[opCode], ID: id}
	case binarySignatureIDOps[opCode] != "":
		id, err := r.u32()
		if err != nil {
			return DecodedBinaryQuery{}, err
		}
		req = &SignatureRequest{Op: binarySignatureIDOps[opCode], ID: id}
	case binaryPrintOps[opCode] != "":
		id, err := r.u32()
		if err != nil {
			return DecodedBinaryQuery{}, err
		}
		presence, err := r.u8()
		if err != nil {
			return DecodedBinaryQuery{}, err
		}
		if presence&^0x03 != 0 {
			return DecodedBinaryQuery{}, errors.New("invalid binary print presence flags")
		}
		kind, err := r.u32()
		if err != nil {
			return DecodedBinaryQuery{}, err
		}
		var flagsPtr *uint32
		if presence&2 != 0 {
			v, err := r.u32()
			if err != nil {
				return DecodedBinaryQuery{}, err
			}
			flagsPtr = &v
		}
		var nodePtr *NodeRef
		if presence&1 != 0 {
			node, fileIdx, err := r.nodeRef()
			if err != nil {
				return DecodedBinaryQuery{}, err
			}
			out.FileIdx = fileIdx
			nodePtr = &node
		}
		req = &PrintRequest{
			Op:    binaryPrintOps[opCode],
			ID:    id,
			Node:  nodePtr,
			Flags: flagsPtr,
			Kind:  kind,
		}
	default:
		switch opCode {
		case 90: // isTypeAssignableTo
			source, err := r.u32()
			if err != nil {
				return DecodedBinaryQuery{}, err
			}
			target, err := r.u32()
			if err != nil {
				return DecodedBinaryQuery{}, err
			}
			req = &AssignableRequest{Source: source, Target: target}
		case 91: // signatures
			id, err := r.u32()
			if err != nil {
				return DecodedBinaryQuery{}, err
			}
			kind, err := r.u32()
			if err != nil {
				return DecodedBinaryQuery{}, err
			}
			req = &SignaturesRequest{ID: id, Kind: kind}
		case 92: // indexInfo
			id, err := r.u32()
			if err != nil {
				return DecodedBinaryQuery{}, err
			}
			kind, err := r.u32()
			if err != nil {
				return DecodedBinaryQuery{}, err
			}
			req = &IndexInfoRequest{ID: id, Kind: kind}
		case 93: // nullableType
			id, err := r.u32()
			if err != nil {
				return DecodedBinaryQuery{}, err
			}
			tFlags, err := r.u32()
			if err != nil {
				return DecodedBinaryQuery{}, err
			}
			req = &NullableTypeRequest{ID: id, Flags: tFlags}
		case 94: // parameterType
			id, err := r.u32()
			if err != nil {
				return DecodedBinaryQuery{}, err
			}
			pos, err := r.i32()
			if err != nil {
				return DecodedBinaryQuery{}, err
			}
			req = &ParameterTypeRequest{ID: id, Position: int(pos)}
		case 95, 96: // symbolDocumentation, symbolJSDocTags
			id, err := r.u32()
			if err != nil {
				return DecodedBinaryQuery{}, err
			}
			wc, err := r.u8()
			if err != nil {
				return DecodedBinaryQuery{}, err
			}
			op := Operation("symbolDocumentation")
			if opCode == 96 {
				op = "symbolJSDocTags"
			}
			req = &SymbolDocsRequest{Op: op, ID: id, WithChecker: wc != 0}
		case 100, 101, 102: // property, propertyType, moduleMember
			id, err := r.u32()
			if err != nil {
				return DecodedBinaryQuery{}, err
			}
			name, err := r.name()
			if err != nil {
				return DecodedBinaryQuery{}, err
			}
			var op Operation
			switch opCode {
			case 100:
				op = "property"
			case 101:
				op = "propertyType"
			case 102:
				op = "moduleMember"
			}
			req = &PropertyRequest{Op: op, ID: id, Name: name}
		case 110: // symbolType
			id, err := r.u32()
			if err != nil {
				return DecodedBinaryQuery{}, err
			}
			node, fileIdx, err := r.nodeRef()
			if err != nil {
				return DecodedBinaryQuery{}, err
			}
			out.FileIdx = fileIdx
			req = &SymbolTypeRequest{ID: id, Node: node}
		case 111: // contextualArgument
			node, fileIdx, err := r.nodeRef()
			if err != nil {
				return DecodedBinaryQuery{}, err
			}
			idx, err := r.i32()
			if err != nil {
				return DecodedBinaryQuery{}, err
			}
			out.FileIdx = fileIdx
			req = &ContextualArgumentRequest{Node: node, Index: int(idx)}
		case 112: // resolvedSignature
			node, fileIdx, err := r.nodeRef()
			if err != nil {
				return DecodedBinaryQuery{}, err
			}
			optFlags, err := r.u8()
			if err != nil {
				return DecodedBinaryQuery{}, err
			}
			if optFlags&^0x03 != 0 {
				return DecodedBinaryQuery{}, errors.New("invalid binary resolvedSignature flags")
			}
			var argCount *int
			if optFlags&2 != 0 {
				v, err := r.i32()
				if err != nil {
					return DecodedBinaryQuery{}, err
				}
				iv := int(v)
				argCount = &iv
			}
			out.FileIdx = fileIdx
			req = &ResolvedSignatureRequest{
				Node:           node,
				WithCandidates: optFlags&1 != 0,
				ArgumentCount:  argCount,
			}
		case 113: // symbolsInScope
			node, fileIdx, err := r.nodeRef()
			if err != nil {
				return DecodedBinaryQuery{}, err
			}
			meaning, err := r.i32()
			if err != nil {
				return DecodedBinaryQuery{}, err
			}
			out.FileIdx = fileIdx
			req = &ScopeSymbolsRequest{Node: node, Meaning: meaning}
		case 114, 115: // validPropertyAccess, parameterPropertySymbols
			node, fileIdx, err := r.nodeRef()
			if err != nil {
				return DecodedBinaryQuery{}, err
			}
			name, err := r.name()
			if err != nil {
				return DecodedBinaryQuery{}, err
			}
			op := Operation("validPropertyAccess")
			if opCode == 115 {
				op = "parameterPropertySymbols"
			}
			out.FileIdx = fileIdx
			req = &NamedNodeRequest{Op: op, Node: node, Name: name}
		default:
			return DecodedBinaryQuery{}, fmt.Errorf("unsupported binary type service opcode %d", opCode)
		}
	}

	if r.remaining() != 0 {
		return DecodedBinaryQuery{}, errors.New("trailing bytes in binary type service request")
	}
	out.Query.Request = req
	return out, nil
}

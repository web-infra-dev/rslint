package typeservice

import (
	"encoding/json"
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// Requests form a finite protocol. Each operand category has required fields;
// no untyped map, optional-argument bag, or arbitrary operation is accepted.
type Operation string
type Scope struct {
	Generation uint32 `json:"generation"`
	Checker    uint32 `json:"checker"`
}
type NodeRef struct {
	File  string `json:"file"`
	Index uint32 `json:"index"`
}

// Name is a lossless JS property/identifier name on the finite query wire.
// JSON strings normalize lone surrogates, so this field carries UTF-16 units.
type Name string

func (name *Name) UnmarshalJSON(data []byte) error {
	var units []uint16
	if err := json.Unmarshal(data, &units); err != nil {
		return err
	}
	*name = Name(ecmascript.StringFromCodeUnits(units))
	return nil
}

type Request interface{ operation() Operation }
type NodeRequest struct {
	Op   Operation `json:"-"`
	Node NodeRef   `json:"node"`
}
type ResolvedSignatureRequest struct {
	Node           NodeRef `json:"node"`
	WithCandidates bool    `json:"withCandidates"`
	ArgumentCount  *int    `json:"argumentCount"`
}
type ContextualArgumentRequest struct {
	Node  NodeRef `json:"node"`
	Index int     `json:"index"`
}
type TypeRequest struct {
	Op Operation `json:"-"`
	ID uint32    `json:"id"`
}
type SymbolRequest struct {
	Op Operation `json:"-"`
	ID uint32    `json:"id"`
}
type SymbolDocsRequest struct {
	Op          Operation `json:"-"`
	ID          uint32    `json:"id"`
	WithChecker bool      `json:"withChecker"`
}
type SignatureRequest struct {
	Op Operation `json:"-"`
	ID uint32    `json:"id"`
}
type ProgramRequest struct {
	Op    Operation `json:"-"`
	Node  *NodeRef  `json:"node"`
	File  string    `json:"file"`
	Index int       `json:"index"`
}

func (r *ProgramRequest) operation() Operation { return r.Op }

type Strings []string

func (Strings) result() {}

type IntrinsicRequest struct {
	Op Operation `json:"-"`
}

func (r *IntrinsicRequest) operation() Operation { return r.Op }

type VoidRequest struct{}
type PropertyRequest struct {
	Op   Operation `json:"-"`
	ID   uint32    `json:"id"`
	Name Name      `json:"name"`
}
type SymbolTypeRequest struct {
	ID   uint32  `json:"id"`
	Node NodeRef `json:"node"`
}
type SignaturesRequest struct {
	ID   uint32 `json:"id"`
	Kind uint32 `json:"kind"`
}
type AssignableRequest struct {
	Source uint32 `json:"source"`
	Target uint32 `json:"target"`
}
type IndexInfoRequest struct {
	ID   uint32 `json:"id"`
	Kind uint32 `json:"kind"`
}
type ScopeSymbolsRequest struct {
	Node    NodeRef `json:"node"`
	Meaning int32   `json:"meaning"`
}
type ParameterTypeRequest struct {
	ID       uint32 `json:"id"`
	Position int    `json:"position"`
}
type PrintRequest struct {
	Op    Operation `json:"-"`
	ID    uint32    `json:"id"`
	Node  *NodeRef  `json:"node"`
	Flags *uint32   `json:"flags"`
	Kind  uint32    `json:"kind"` // SignatureKind, defaults to Call.
}
type ResolveNameRequest struct {
	Name           Name     `json:"name"`
	Node           *NodeRef `json:"node"`
	Kind           int32    `json:"kind"` // JavaScript SymbolFlags.All is -1.
	ExcludeGlobals bool     `json:"excludeGlobals"`
}

func (r *NodeRequest) operation() Operation             { return r.Op }
func (*ContextualArgumentRequest) operation() Operation { return "contextualArgument" }
func (*ResolvedSignatureRequest) operation() Operation {
	return "resolvedSignatureWithOptions"
}
func (r *TypeRequest) operation() Operation       { return r.Op }
func (r *SymbolRequest) operation() Operation     { return r.Op }
func (r *SymbolDocsRequest) operation() Operation { return r.Op }
func (r *SignatureRequest) operation() Operation  { return r.Op }
func (*VoidRequest) operation() Operation         { return "void" }
func (r *PropertyRequest) operation() Operation   { return r.Op }
func (*SymbolTypeRequest) operation() Operation   { return "symbolType" }
func (*SignaturesRequest) operation() Operation   { return "signatures" }
func (*AssignableRequest) operation() Operation   { return "isTypeAssignableTo" }
func (*IndexInfoRequest) operation() Operation    { return "indexInfo" }
func (*ScopeSymbolsRequest) operation() Operation { return "symbolsInScope" }
func (*ParameterTypeRequest) operation() Operation {
	return "parameterType"
}
func (r *PrintRequest) operation() Operation     { return r.Op }
func (*ResolveNameRequest) operation() Operation { return "resolveName" }

type StringLiteralRequest struct {
	Op   Operation `json:"-"`
	Name Name      `json:"name"`
}

func (r *StringLiteralRequest) operation() Operation { return r.Op }

type NumberLiteralRequest struct {
	Op    Operation `json:"-"`
	Value string    `json:"value"`
}

func (r *NumberLiteralRequest) operation() Operation { return r.Op }

type BigIntLiteralRequest struct {
	Op       Operation `json:"-"`
	Negative bool      `json:"negative"`
	Digits   string    `json:"digits"`
}

func (r *BigIntLiteralRequest) operation() Operation { return r.Op }

type NullableRequest struct {
	Op          Operation `json:"-"`
	ID          uint32    `json:"id"`
	Nullable    bool      `json:"nullable"`
	Undefinable bool      `json:"undefinable"`
}

func (r *NullableRequest) operation() Operation { return r.Op }

type PrivatePropertyRequest struct {
	Op   Operation `json:"-"`
	ID   uint32    `json:"id"`
	Name Name      `json:"name"`
	Node NodeRef   `json:"node"`
}

func (r *PrivatePropertyRequest) operation() Operation { return r.Op }

type NamedNodeRequest struct {
	Op   Operation `json:"-"`
	Name Name      `json:"name"`
	Node NodeRef   `json:"node"`
}

func (r *NamedNodeRequest) operation() Operation { return r.Op }

type ExpandoRequest struct {
	Op               Operation `json:"-"`
	AllowDeclaration bool      `json:"allowDeclaration"`
	Node             NodeRef   `json:"node"`
}

func (r *ExpandoRequest) operation() Operation { return r.Op }

type IndexSymbolRequest struct {
	Op       Operation `json:"-"`
	ID       uint32    `json:"id"`
	Siblings []uint32  `json:"siblings"`
}

func (r *IndexSymbolRequest) operation() Operation { return r.Op }

type PredicateStringRequest struct {
	Op       Operation `json:"-"`
	Kind     int32     `json:"kind"`
	Name     Name      `json:"name"`
	Position int32     `json:"position"`
	Type     uint32    `json:"type"`
	Node     *NodeRef  `json:"node"`
	Flags    uint32    `json:"flags"`
}

func (r *PredicateStringRequest) operation() Operation { return r.Op }

type IndexDeclarationRequest struct {
	Op          Operation `json:"-"`
	Key         uint32    `json:"key"`
	Value       uint32    `json:"value"`
	Readonly    bool      `json:"readonly"`
	Declaration *NodeRef  `json:"declaration"`
	Node        *NodeRef  `json:"node"`
	Flags       uint32    `json:"flags"`
}

func (r *IndexDeclarationRequest) operation() Operation { return r.Op }

type Query struct {
	Scope
	Request Request
}

func (q *Query) UnmarshalJSON(data []byte) error {
	var wire struct {
		Scope
		Op   Operation       `json:"op"`
		Args json.RawMessage `json:"args"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	var request Request
	switch wire.Op {
	case "conditionalInstantiations", "freshLiteral", "regularType", "evolvingElement", "finalArray", "simplifiedForReading", "simplifiedForWriting", "typeParameterConstraint":
		request = &TypeRequest{Op: wire.Op}
	case "typeDetails", "resolvedTrueType", "resolvedFalseType", "declaredMembers":
		request = &TypeRequest{Op: wire.Op}
	case "symbolExports", "symbolGlobalExports":
		request = &SymbolRequest{Op: wire.Op}
	case "configFileParsingDiagnostics", "optionsDiagnostics", "globalDiagnostics", "syntacticDiagnostics", "semanticDiagnostics", "declarationDiagnostics", "nodeCount", "identifierCount", "symbolCount", "typeCount", "instantiationCount", "relationCacheSizes", "rootFileNames", "modeForUsageLocation", "modeForResolutionAtIndex", "projectReferences", "resolvedProjectReferences":
		request = &ProgramRequest{Op: wire.Op}
	case "stringLiteralType":
		request = &StringLiteralRequest{Op: wire.Op}
	case "numberLiteralType":
		request = &NumberLiteralRequest{Op: wire.Op}
	case "bigintLiteralType":
		request = &BigIntLiteralRequest{Op: wire.Op}
	case "nullableType":
		request = &NullableRequest{Op: wire.Op}
	case "privateIdentifierProperty":
		request = &PrivatePropertyRequest{Op: wire.Op}
	case "moduleMember":
		request = &PropertyRequest{Op: wire.Op}
	case "validPropertyAccess", "parameterPropertySymbols":
		request = &NamedNodeRequest{Op: wire.Op}
	case "expandoSymbol":
		request = &ExpandoRequest{Op: wire.Op}
	case "indexSymbolInfos":
		request = &IndexSymbolRequest{Op: wire.Op}
	case "signatureDeclaration", "parameterDeclaration", "typeParameterDeclarations", "typeParameterDeclaration", "entityName", "symbolExpression", "symbolString":
		request = &PrintRequest{Op: wire.Op}
	case "predicateString":
		request = &PredicateStringRequest{Op: wire.Op}
	case "indexDeclaration":
		request = &IndexDeclarationRequest{Op: wire.Op}
	case "anyType", "unknownType", "stringType", "numberType", "booleanType", "bigIntType", "eSSymbolType", "nonPrimitiveType", "neverType", "nullType", "undefinedType", "trueType", "falseType", "ambientModules":
		request = &IntrinsicRequest{Op: wire.Op}
	case "augmentedPropertiesOfType":
		request = &TypeRequest{Op: wire.Op}
	case "rootSymbols", "mergedSymbol", "isArgumentsSymbol", "isUndefinedSymbol":
		request = &SymbolRequest{Op: wire.Op}
	case "exportSpecifierLocalTargetSymbol", "propertySymbolOfDestructuringAssignment", "jsxIntrinsicTagNamesAt", "typeOfAssignmentPattern", "isOptionalParameter", "constantValue", "isImplementationOfOverload":
		request = &NodeRequest{Op: wire.Op}
	case "typeArgumentsForResolvedSignature":
		request = &SignatureRequest{Op: wire.Op}
	case "type", "fromTypeNode", "context", "symbol", "signature", "resolvedSignature", "nodeRecord", "shorthandValueSymbol", "suggestionDiagnostics":
		request = &NodeRequest{Op: wire.Op}
	case "contextualArgument":
		request = &ContextualArgumentRequest{}
	case "resolvedSignatureWithOptions":
		request = &ResolvedSignatureRequest{}
	case "constraint", "apparent", "baseLiteral", "nonNullable", "properties", "typeArguments", "isArray", "isArrayLike", "isTuple", "numberIndexType", "stringIndexType", "typeRecord",
		"awaited", "widened", "baseTypes", "defaultType", "apparentProperties", "indexInfos", "mappedFields", "classThisType":
		request = &TypeRequest{Op: wire.Op}
	case "aliasedSymbol", "declaredType", "symbolRecord", "typeOfSymbol", "exportsOfModule", "exportSymbol", "immediateAliasedSymbol", "isUnknownSymbol", "fullyQualifiedName", "symbolMembers", "symbolLinkType":
		request = &SymbolRequest{Op: wire.Op}
	case "symbolDocumentation", "symbolJSDocTags":
		request = &SymbolDocsRequest{Op: wire.Op}
	case "returnType", "signatureRecord", "typePredicate", "signatureDocumentation", "signatureJSDocTags":
		request = &SignatureRequest{Op: wire.Op}
	case "void":
		request = &VoidRequest{}
	case "property", "propertyType":
		request = &PropertyRequest{Op: wire.Op}
	case "symbolType":
		request = &SymbolTypeRequest{}
	case "signatures":
		request = &SignaturesRequest{}
	case "isTypeAssignableTo":
		request = &AssignableRequest{}
	case "indexInfo":
		request = &IndexInfoRequest{}
	case "symbolsInScope":
		request = &ScopeSymbolsRequest{}
	case "parameterType":
		request = &ParameterTypeRequest{}
	case "typeString", "signatureString", "typeNode":
		request = &PrintRequest{Op: wire.Op}
	case "resolveName":
		request = &ResolveNameRequest{}
	default:
		return fmt.Errorf("unsupported type service operation %q", wire.Op)
	}
	if err := json.Unmarshal(wire.Args, request); err != nil {
		return err
	}
	q.Scope, q.Request = wire.Scope, request
	return nil
}

func requestNode(request Request) *NodeRef {
	switch r := request.(type) {
	case *ProgramRequest:
		return r.Node
	case *PrivatePropertyRequest:
		return &r.Node
	case *NamedNodeRequest:
		return &r.Node
	case *ExpandoRequest:
		return &r.Node
	case *PredicateStringRequest:
		return r.Node
	case *IndexDeclarationRequest:
		return r.Node
	case *NodeRequest:
		return &r.Node
	case *ResolvedSignatureRequest:
		return &r.Node
	case *ContextualArgumentRequest:
		return &r.Node
	case *SymbolTypeRequest:
		return &r.Node
	case *ScopeSymbolsRequest:
		return &r.Node
	case *PrintRequest:
		return r.Node
	case *ResolveNameRequest:
		return r.Node
	default:
		return nil
	}
}

// Result is closed to this package. Variable fields are copied under the
// checker lease; syntax references are resolved afterward, before encoding.
type Result interface{ result() }
type Handle uint32
type Handles []uint32
type OptionalHandles []uint32

func (OptionalHandles) result() {}
func (Literal) result()         {}

type Boolean bool
type Text string
type EncodedTypeNode []byte
type SuggestionDiagnostics []DiagnosticFact
type DiagnosticFact struct {
	File, Message, Source         string
	Start, Length, Category, Code uint32
	Deprecated, Unnecessary       bool
	Chain, Related                SuggestionDiagnostics
}

func (EncodedTypeNode) result()       {}
func (SuggestionDiagnostics) result() {}

type TypeFact struct {
	Flags               []string
	Symbol, AliasSymbol uint32
	IntrinsicName       string
	EscapedName         string
	ObjectFlags         []string
	Literal             Literal
	TypeOperands        []uint32 // Contents follow the type flags.
	Target              uint32
	AliasTypeArguments  []uint32
	TypeParameters      []uint32
	Tuple               *TupleFact
}
type TupleFact struct {
	CombinedFlags uint32
	Readonly      bool
}
type Literal struct {
	Kind     uint32
	Text     string
	Number   float64
	Negative bool
}
type SymbolFact struct {
	Name, EscapedName                        string
	Flags                                    uint32
	CheckFlags                               uint32
	HasMembers, HasExports, HasGlobalExports bool
	Declarations                             []NodeRef
	ValueDeclaration                         *NodeRef
	nodes                                    []*ast.Node
	valueNode                                *ast.Node
}

// NodeFact is a finite syntax summary. Structural traversal still uses exact
// source syntax; declaration ancestry never requires serializing that source.
type NodeFact struct {
	Node          NodeRef
	Kind          string
	Flags         uint32
	ModuleKeyword uint32
	Parent, Name  *NodeRef
	Text          *string
}
type NodeFacts []NodeFact

// Record supplements a queried handle with its directly requested object fact.
// It never recursively expands referenced types or their properties.
type Record struct {
	Op   Operation
	ID   uint32
	Fact Result
}

type SignatureFact struct {
	Parameters       []uint32
	Declaration      *NodeRef
	TypeParameters   []uint32
	ThisParameter    uint32
	MinArgumentCount int
	node             *ast.Node
}
type NamedSymbols []NamedSymbol
type NamedSymbol struct {
	ID   uint32
	Name string
}
type IndexInfos []IndexInfo
type IndexInfo struct {
	ID             uint32
	KeyType, Type  uint32
	IsReadonly     bool
	Declaration    *NodeRef
	node           *ast.Node
	componentNodes []*ast.Node
	Components     []*NodeRef
}
type TypePredicateFact struct {
	Kind           uint32
	ParameterName  *string
	ParameterIndex int32
	Type           uint32
}

type DisplayParts []DisplayPart
type DisplayPart struct {
	Text, Kind string
	Target     *DisplayPartTarget
}
type DisplayPartTarget struct {
	File          string
	Start, Length uint32
}
type JSDocTags []JSDocTag
type JSDocTag struct {
	Name string
	Text DisplayParts
}

func (NodeFacts) result()          {}
func (Handle) result()             {}
func (Handles) result()            {}
func (Boolean) result()            {}
func (Text) result()               {}
func (*TypeFact) result()          {}
func (*SymbolFact) result()        {}
func (*SignatureFact) result()     {}
func (NamedSymbols) result()       {}
func (IndexInfos) result()         {}
func (*TypePredicateFact) result() {}
func (DisplayParts) result()       {}
func (JSDocTags) result()          {}

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
}

type NodeRef struct {
	File  string       `json:"file"`
	Index uint32       `json:"index"`
	Token *SyntaxToken `json:"token,omitempty"`
}

// The binary compiler tree stores punctuation and some keywords on their parent.
// A syntax reader identifies those children by their UTF-16 range.
type SyntaxToken struct {
	Kind ast.Kind `json:"kind"`
	Pos  int      `json:"pos"`
	End  int      `json:"end"`
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

type ContextualArgumentRequest struct {
	Node  NodeRef `json:"node"`
	Index int     `json:"index"`
}

type NamedNodeRequest struct {
	Op   Operation `json:"-"`
	Name Name      `json:"name"`
	Node NodeRef   `json:"node"`
}

type ProgramRequest struct {
	Op   Operation `json:"-"`
	Node *NodeRef  `json:"node"`
}

type CheckerRequest struct {
	Op Operation `json:"-"`
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
	Kind  uint32    `json:"kind"` // Meaning for symbolString; SignatureKind for signatureString.
}

type ResolvedSignatureRequest struct {
	Node           NodeRef `json:"node"`
	WithCandidates bool    `json:"withCandidates,omitempty"`
	ArgumentCount  *int    `json:"argumentCount,omitempty"`
}

type StringLiteralRequest struct {
	Value Name `json:"value"`
}

type NumberLiteralRequest struct {
	Value     float64 `json:"value"`
	NonFinite string  `json:"nonFinite,omitempty"`
}

type BigIntLiteralRequest struct {
	Negative    bool   `json:"negative"`
	Base10Value string `json:"base10Value"`
}

type NullableTypeRequest struct {
	ID    uint32 `json:"id"`
	Flags uint32 `json:"flags"`
}

type TypePredicateStringRequest struct {
	Kind           uint32   `json:"kind"`
	ParameterName  *Name    `json:"parameterName,omitempty"`
	ParameterIndex int32    `json:"parameterIndex,omitempty"`
	Type           uint32   `json:"type"`
	Node           *NodeRef `json:"node"`
	Flags          *uint32  `json:"flags"`
}

type ResolveNameRequest struct {
	Name           Name     `json:"name"`
	Node           *NodeRef `json:"node"`
	Kind           int32    `json:"kind"` // JavaScript SymbolFlags.All is -1.
	ExcludeGlobals bool     `json:"excludeGlobals"`
}

func (r *NodeRequest) operation() Operation              { return r.Op }
func (*ContextualArgumentRequest) operation() Operation  { return "contextualArgument" }
func (r *NamedNodeRequest) operation() Operation         { return r.Op }
func (r *ProgramRequest) operation() Operation           { return r.Op }
func (r *CheckerRequest) operation() Operation           { return r.Op }
func (r *TypeRequest) operation() Operation              { return r.Op }
func (r *SymbolRequest) operation() Operation            { return r.Op }
func (r *SymbolDocsRequest) operation() Operation        { return r.Op }
func (r *SignatureRequest) operation() Operation         { return r.Op }
func (r *PropertyRequest) operation() Operation          { return r.Op }
func (*SymbolTypeRequest) operation() Operation          { return "symbolType" }
func (*SignaturesRequest) operation() Operation          { return "signatures" }
func (*AssignableRequest) operation() Operation          { return "isTypeAssignableTo" }
func (*IndexInfoRequest) operation() Operation           { return "indexInfo" }
func (*ScopeSymbolsRequest) operation() Operation        { return "symbolsInScope" }
func (*ParameterTypeRequest) operation() Operation       { return "parameterType" }
func (r *PrintRequest) operation() Operation             { return r.Op }
func (*ResolvedSignatureRequest) operation() Operation   { return "resolvedSignature" }
func (*StringLiteralRequest) operation() Operation       { return "stringLiteralType" }
func (*NumberLiteralRequest) operation() Operation       { return "numberLiteralType" }
func (*BigIntLiteralRequest) operation() Operation       { return "bigIntLiteralType" }
func (*NullableTypeRequest) operation() Operation        { return "nullableType" }
func (*TypePredicateStringRequest) operation() Operation { return "typePredicateString" }
func (*ResolveNameRequest) operation() Operation         { return "resolveName" }

type Query struct {
	Scope
	Request     Request
	SkipRecords bool
}

func (q *Query) UnmarshalJSON(data []byte) error {
	var wire struct {
		Scope
		Op          Operation       `json:"op"`
		Args        json.RawMessage `json:"args"`
		SkipRecords bool            `json:"skipRecords,omitempty"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	var request Request
	switch wire.Op {
	case "rootFileNames", "configFileParsingDiagnostics", "optionsDiagnostics",
		"globalDiagnostics", "syntacticDiagnostics", "semanticDiagnostics", "declarationDiagnostics":
		request = &ProgramRequest{Op: wire.Op}
	case "errorType", "anyType", "unknownType", "stringType", "numberType", "booleanType",
		"bigIntType", "eSSymbolType", "nonPrimitiveType", "neverType", "nullType",
		"undefinedType", "trueType", "falseType", "void", "ambientModules":
		request = &CheckerRequest{Op: wire.Op}
	case "type", "fromTypeNode", "context", "symbol", "declarationSymbol", "signature",
		"nodeRecord", "shorthandValueSymbol", "suggestionDiagnostics",
		"exportSpecifierLocalTargetSymbol", "propertySymbolOfDestructuringAssignment",
		"jsxIntrinsicTagNamesAt", "constantValue", "isOptionalParameter",
		"isImplementationOfOverload":
		request = &NodeRequest{Op: wire.Op}
	case "resolvedSignature":
		request = &ResolvedSignatureRequest{}
	case "contextualArgument":
		request = &ContextualArgumentRequest{}
	case "validPropertyAccess", "parameterPropertySymbols":
		request = &NamedNodeRequest{Op: wire.Op}
	case "stringLiteralType":
		request = &StringLiteralRequest{}
	case "numberLiteralType":
		request = &NumberLiteralRequest{}
	case "bigIntLiteralType":
		request = &BigIntLiteralRequest{}
	case "nullableType":
		request = &NullableTypeRequest{}
	case "typePredicateString":
		request = &TypePredicateStringRequest{}
	case "constraint", "apparent", "baseLiteral", "nonNullable", "properties",
		"augmentedPropertiesOfType", "typeArguments", "isArray", "isArrayLike", "isTuple",
		"numberIndexType", "stringIndexType", "typeRecord", "typeDetails", "declaredIndexInfos",
		"awaited", "widened", "baseTypes", "defaultType", "indexInfos", "mappedFields",
		"classThisType":
		request = &TypeRequest{Op: wire.Op}
	case "rootSymbols", "mergedSymbol", "isArgumentsSymbol", "isUndefinedSymbol",
		"aliasedSymbol", "immediateAliasedSymbol", "isUnknownSymbol", "exportsOfModule",
		"exportSymbol", "fullyQualifiedName", "symbolMembers", "symbolExports",
		"declaredType", "typeOfSymbol", "symbolRecord":
		request = &SymbolRequest{Op: wire.Op}
	case "symbolDocumentation", "symbolJSDocTags":
		request = &SymbolDocsRequest{Op: wire.Op}
	case "returnType", "signatureRecord", "typePredicate", "signatureDocumentation", "signatureJSDocTags":
		request = &SignatureRequest{Op: wire.Op}
	case "property", "propertyType", "moduleMember":
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
	case "typeString", "signatureString", "symbolString", "typeNode":
		request = &PrintRequest{Op: wire.Op}
	case "resolveName":
		request = &ResolveNameRequest{}
	default:
		return fmt.Errorf("unsupported type service operation %q", wire.Op)
	}
	if err := json.Unmarshal(wire.Args, request); err != nil {
		return err
	}
	q.Scope, q.Request, q.SkipRecords = wire.Scope, request, wire.SkipRecords
	return nil
}

func requestNode(request Request) *NodeRef {
	switch r := request.(type) {
	case *ProgramRequest:
		return r.Node
	case *NamedNodeRequest:
		return &r.Node
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
	case *TypePredicateStringRequest:
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
type Boolean bool
type Text string
type Strings []string

type Literal struct {
	Kind     uint32
	Text     string
	Number   float64
	Negative bool
}

type EncodedTypeNode struct {
	tree       []byte
	parseNodes []*ast.Node
	origins    []*NodeRef
	emitFlags  []uint32
}

type SuggestionDiagnostics []DiagnosticFact

type DiagnosticFact struct {
	File, FilePath, Message, Source string
	Start, Length, Category, Code   uint32
	Deprecated, Unnecessary         bool
	Chain, Related                  SuggestionDiagnostics
}

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

type SymbolFact struct {
	Name, EscapedName string
	Flags             uint32
	CheckFlags        uint32
	Declarations      []NodeRef
	ValueDeclaration  *NodeRef
	nodes             []*ast.Node
	valueNode         *ast.Node
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

type OptionalNamedSymbols struct {
	Present bool
	Symbols NamedSymbols
}

type IndexInfos []IndexInfo
type IndexInfo struct {
	ID            uint32
	KeyType, Type uint32
	IsReadonly    bool
	Declaration   *NodeRef
	node          *ast.Node
}

type OptionalIndexInfos struct {
	Present bool
	Infos   IndexInfos
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

func (Handle) result()                {}
func (Handles) result()               {}
func (Boolean) result()               {}
func (Text) result()                  {}
func (Strings) result()               {}
func (Literal) result()               {}
func (*EncodedTypeNode) result()      {}
func (SuggestionDiagnostics) result() {}
func (*TypeFact) result()             {}
func (*SymbolFact) result()           {}
func (*SignatureFact) result()        {}
func (NodeFacts) result()             {}
func (NamedSymbols) result()          {}
func (OptionalNamedSymbols) result()  {}
func (IndexInfos) result()            {}
func (OptionalIndexInfos) result()    {}
func (*TypePredicateFact) result()    {}
func (DisplayParts) result()          {}
func (JSDocTags) result()             {}

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
type TypeRequest struct {
	Op Operation `json:"-"`
	ID uint32    `json:"id"`
}
type SymbolRequest struct {
	Op Operation `json:"-"`
	ID uint32    `json:"id"`
}
type SignatureRequest struct {
	Op Operation `json:"-"`
	ID uint32    `json:"id"`
}
type VoidRequest struct{}
type PropertyRequest struct {
	ID   uint32 `json:"id"`
	Name Name   `json:"name"`
}
type SymbolTypeRequest struct {
	ID   uint32  `json:"id"`
	Node NodeRef `json:"node"`
}
type SignaturesRequest struct {
	ID   uint32 `json:"id"`
	Kind uint32 `json:"kind"`
}
type ResolveNameRequest struct {
	Name           Name     `json:"name"`
	Node           *NodeRef `json:"node"`
	Kind           uint32   `json:"kind"`
	ExcludeGlobals bool     `json:"excludeGlobals"`
}

func (r *NodeRequest) operation() Operation      { return r.Op }
func (r *TypeRequest) operation() Operation      { return r.Op }
func (r *SymbolRequest) operation() Operation    { return r.Op }
func (r *SignatureRequest) operation() Operation { return r.Op }
func (*VoidRequest) operation() Operation        { return "void" }
func (*PropertyRequest) operation() Operation    { return "property" }
func (*SymbolTypeRequest) operation() Operation  { return "symbolType" }
func (*SignaturesRequest) operation() Operation  { return "signatures" }
func (*ResolveNameRequest) operation() Operation { return "resolveName" }

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
	case "type", "fromTypeNode", "context", "symbol", "signature", "resolvedSignature":
		request = &NodeRequest{Op: wire.Op}
	case "constraint", "apparent", "baseLiteral", "nonNullable", "typeString", "properties", "typeArguments", "isArray", "isTuple", "typeRecord":
		request = &TypeRequest{Op: wire.Op}
	case "aliasedSymbol", "declaredType", "symbolRecord":
		request = &SymbolRequest{Op: wire.Op}
	case "returnType", "signatureRecord":
		request = &SignatureRequest{Op: wire.Op}
	case "void":
		request = &VoidRequest{}
	case "property":
		request = &PropertyRequest{}
	case "symbolType":
		request = &SymbolTypeRequest{}
	case "signatures":
		request = &SignaturesRequest{}
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
	case *NodeRequest:
		return &r.Node
	case *SymbolTypeRequest:
		return &r.Node
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
type TypeFact struct {
	Flags               []string
	Symbol, AliasSymbol uint32
	IntrinsicName       string
	Literal             Literal
	Types               []uint32
}
type Literal struct {
	Kind     uint32
	Text     string
	Number   float64
	Negative bool
}
type SymbolFact struct {
	Name, EscapedName string
	Flags             uint32
	Declarations      []NodeRef
	ValueDeclaration  *NodeRef
	nodes             []*ast.Node
	valueNode         *ast.Node
}
type SignatureFact struct {
	Parameters  []uint32
	Declaration *NodeRef
	node        *ast.Node
}

func (Handle) result()         {}
func (Handles) result()        {}
func (Boolean) result()        {}
func (Text) result()           {}
func (*TypeFact) result()      {}
func (*SymbolFact) result()    {}
func (*SignatureFact) result() {}

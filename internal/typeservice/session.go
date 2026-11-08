package typeservice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/microsoft/TypeScript/tsc/shim/api/encoder"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/jsnum"
	"github.com/web-infra-dev/rslint/internal/program"
)

var nextGeneration atomic.Uint32
var ErrClosed = errors.New("type service session is closed")

// Public TypeScript ObjectFlags have different bit positions from Go's flags.
// Do not expose checker-private cache state through the compatibility surface.
var publicObjectFlags = [...]struct {
	flag checker.ObjectFlags
	name string
}{
	{checker.ObjectFlagsClass, "Class"},
	{checker.ObjectFlagsInterface, "Interface"},
	{checker.ObjectFlagsReference, "Reference"},
	{checker.ObjectFlagsTuple, "Tuple"},
	{checker.ObjectFlagsAnonymous, "Anonymous"},
	{checker.ObjectFlagsMapped, "Mapped"},
	{checker.ObjectFlagsInstantiated, "Instantiated"},
	{checker.ObjectFlagsObjectLiteral, "ObjectLiteral"},
	{checker.ObjectFlagsEvolvingArray, "EvolvingArray"},
	{checker.ObjectFlagsObjectLiteralPatternWithComputedProperties, "ObjectLiteralPatternWithComputedProperties"},
	{checker.ObjectFlagsReverseMapped, "ReverseMapped"},
	{checker.ObjectFlagsJsxAttributes, "JsxAttributes"},
	{checker.ObjectFlagsJSLiteral, "JSLiteral"},
	{checker.ObjectFlagsFreshLiteral, "FreshLiteral"},
	{checker.ObjectFlagsArrayLiteral, "ArrayLiteral"},
	{checker.ObjectFlagsSingleSignatureType, "SingleSignatureType"},
	{checker.ObjectFlagsContainsSpread, "ContainsSpread"},
	{checker.ObjectFlagsObjectRestType, "ObjectRestType"},
	{checker.ObjectFlagsInstantiationExpressionType, "InstantiationExpressionType"},
}

// Session owns the Program and checker objects until Close. It must not be
// placed in Program.Cached, whose weak cache must not retain the Program.
// One anchored checker supplies all objects, including cross-file declarations.
// Exclusive access is acquired and released for each query, never for the
// lifetime of the session. Callers must close every client before Close.
type Session struct {
	mu         sync.Mutex
	program    *program.Program
	anchor     *ast.SourceFile
	eligible   map[*ast.SourceFile]bool
	checker    *checker.Checker
	generation uint32
	types      map[uint32]*checker.Type
	symbols    map[uint32]*ast.Symbol
	signatures map[uint32]*checker.Signature
}

type Descriptor struct {
	Generation       uint32         `json:"generation"`
	Checker          uint32         `json:"checker"`
	CurrentDirectory string         `json:"currentDirectory"`
	CompilerOptions  map[string]any `json:"compilerOptions"`
	Sources          []Source       `json:"sources"`
}
type Source struct {
	FileName       string   `json:"fileName"`
	Path           string   `json:"path"`
	Packages       []string `json:"packages"`
	External       bool     `json:"external"`
	DefaultLibrary bool     `json:"defaultLibrary"`
}

// Binding authorizes one eligible lint target. Dependency nodes remain valid
// semantic operands; binding identity is not part of immutable fact cache keys.
type Binding struct {
	Scope
	SourceFileName string `json:"sourceFileName"`
}

func (s *Session) BindTarget(file *ast.SourceFile) (Binding, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.program == nil {
		return Binding{}, ErrClosed
	}
	if !s.eligible[file] {
		return Binding{}, errors.New("source is not eligible for typed rules")
	}
	return Binding{Scope{s.generation, 1}, file.FileName()}, nil
}

// Source captures an exact immutable source under a brief generation check.
// Callers encode/project it after this method returns, outside every lease.
func (s *Session) Source(generation uint32, name string) (*ast.SourceFile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.program == nil {
		return nil, ErrClosed
	}
	if generation != s.generation {
		return nil, errors.New("stale type service generation")
	}
	return s.program.GetSourceFile(name), nil //nolint:nilnil // A missing Program source is a legitimate result.
}

// New accepts exact Program sources that the caller has explicitly admitted
// for typed rules. Checker capability alone is never eligibility.
func New(ctx context.Context, p *program.Program, eligible []*ast.SourceFile) (*Session, error) {
	if !p.IsValid() || len(eligible) == 0 {
		return nil, errors.New("type services require a bound Program and explicitly eligible sources")
	}
	s := &Session{program: p, anchor: eligible[0], eligible: make(map[*ast.SourceFile]bool), generation: nextGeneration.Add(1), types: make(map[uint32]*checker.Type), symbols: make(map[uint32]*ast.Symbol), signatures: make(map[uint32]*checker.Signature)}
	for _, f := range eligible {
		if !p.CanProvideTypeChecker(f) {
			return nil, errors.New("ineligible or foreign type service source")
		}
		s.eligible[f] = true
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	c, done := p.TypeCheckerForFileExclusive(core.WithCheckerLifetime(ctx, core.CheckerLifetimeAPI), s.anchor)
	defer done()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if c == nil {
		return nil, errors.New("program cannot provide type services")
	}
	s.checker = c
	return s, nil
}

func (s *Session) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.program = nil
	s.anchor = nil
	s.checker = nil
	s.eligible = nil
	s.types = nil
	s.symbols = nil
	s.signatures = nil
}
func (s *Session) Descriptor() (Descriptor, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.program == nil {
		return Descriptor{}, ErrClosed
	}
	// The compiler's JSON contract preserves explicit false and omitted values.
	data, err := json.Marshal(s.program.Options())
	if err != nil {
		return Descriptor{}, err
	}
	options := map[string]any{}
	if err = json.Unmarshal(data, &options); err != nil {
		return Descriptor{}, err
	}
	// These compiler enums differ from TypeScript's JavaScript enum layout.
	opts := s.program.Options()
	if opts.NewLine != core.NewLineKindNone {
		options["newLine"] = int(opts.NewLine) - 1
	}
	if opts.Jsx != core.JsxEmitNone {
		options["jsx"] = map[core.JsxEmit]string{core.JsxEmitPreserve: "Preserve", core.JsxEmitReact: "React", core.JsxEmitReactNative: "ReactNative", core.JsxEmitReactJSX: "ReactJSX", core.JsxEmitReactJSXDev: "ReactJSXDev"}[opts.Jsx]
	}
	if opts.ModuleDetection != core.ModuleDetectionKindNone {
		options["moduleDetection"] = map[core.ModuleDetectionKind]string{core.ModuleDetectionKindAuto: "Auto", core.ModuleDetectionKindLegacy: "Legacy", core.ModuleDetectionKindForce: "Force"}[opts.ModuleDetection]
	}
	d := Descriptor{Generation: s.generation, Checker: 1, CurrentDirectory: s.program.CurrentDirectory(), CompilerOptions: options}
	for _, f := range s.program.SourceFiles() {
		d.Sources = append(d.Sources, s.source(f))
	}
	return d, nil
}
func (s *Session) source(f *ast.SourceFile) Source {
	return Source{FileName: f.FileName(), Path: string(f.Path()), Packages: s.program.PackageNamesForSourceFile(f), External: s.program.IsSourceFileFromExternalLibrary(f), DefaultLibrary: s.program.IsSourceFileDefaultLibrary(f)}
}

// Query copies checker facts while exclusively leased. Source indexing and
// serialization happen afterward; they cannot delay another checker borrower.
func (s *Session) Query(ctx context.Context, q Query) (packet []byte, err error) {
	var op Operation
	defer func() {
		if value := recover(); value != nil {
			packet = nil
			err = fmt.Errorf("type service %s: %v", op, value)
		}
	}()
	s.mu.Lock()
	if s.program == nil {
		s.mu.Unlock()
		return nil, ErrClosed
	}
	if q.Generation != s.generation || q.Checker != 1 {
		s.mu.Unlock()
		return nil, errors.New("stale or foreign type service handle")
	}
	p := s.program
	s.mu.Unlock()
	if q.Request == nil {
		return nil, errors.New("missing type service request")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	op = q.Request.operation()
	node := requestNode(q.Request)
	var native *ast.Node
	if node != nil {
		f := p.GetSourceFile(node.File)
		if f == nil {
			return nil, errors.New("foreign source handle")
		}
		table := encoder.GetNodeIndexTable(f)
		if node.Index == 0 || int(node.Index) >= len(table.Nodes) {
			return nil, errors.New("invalid compiler node handle")
		}
		native = table.Nodes[node.Index]
		if native == nil {
			return nil, errors.New("invalid compiler node handle")
		}
	}
	result, err := s.check(ctx, q.Request, native)
	if err != nil {
		return nil, err
	}
	switch fact := result.(type) {
	case *SymbolFact:
		if fact.nodes != nil {
			fact.Declarations = make([]NodeRef, 0, len(fact.nodes))
			for _, node := range fact.nodes {
				fact.Declarations = append(fact.Declarations, *nodeReference(p, node))
			}
		}
		fact.ValueDeclaration = nodeReference(p, fact.valueNode)
		fact.nodes = nil
		fact.valueNode = nil
	case *SignatureFact:
		fact.Declaration = nodeReference(p, fact.node)
		fact.node = nil
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return EncodePacket(result)
}
func (s *Session) check(ctx context.Context, request Request, native *ast.Node) (Result, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.program == nil {
		return nil, ErrClosed
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	c, done := s.program.TypeCheckerForFileExclusive(core.WithCheckerLifetime(ctx, core.CheckerLifetimeAPI), s.anchor)
	defer done()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if c != s.checker {
		return nil, errors.New("type service checker identity changed")
	}
	return s.execute(c, request, native)
}
func nodeReference(p *program.Program, n *ast.Node) *NodeRef {
	if n == nil {
		return nil
	}
	f := ast.GetSourceFileOfNode(n)
	if !p.OwnsSourceFile(f) {
		panic("declaration is outside Program")
	}
	index := encoder.GetNodeIndexTable(f).GetIndex(n)
	if index == 0 {
		panic("declaration has no compiler node handle")
	}
	return &NodeRef{f.FileName(), index}
}
func (s *Session) typeByID(id uint32) *checker.Type {
	t := s.types[id]
	if t == nil {
		panic("invalid type handle")
	}
	return t
}
func (s *Session) symbolByID(id uint32) *ast.Symbol {
	v := s.symbols[id]
	if v == nil {
		panic("invalid symbol handle")
	}
	return v
}
func (s *Session) signatureByID(id uint32) *checker.Signature {
	v := s.signatures[id]
	if v == nil {
		panic("invalid signature handle")
	}
	return v
}
func (s *Session) typeID(t *checker.Type) uint32 {
	if t == nil {
		return 0
	}
	id := uint32(t.Id())
	s.types[id] = t
	return id
}
func (s *Session) symbolID(v *ast.Symbol) uint32 {
	if v == nil {
		return 0
	}
	id := uint32(ast.GetSymbolId(v))
	s.symbols[id] = v
	return id
}
func (s *Session) signatureID(v *checker.Signature) uint32 {
	if v == nil {
		return 0
	}
	id := uint32(v.Id())
	s.signatures[id] = v
	return id
}
func (s *Session) execute(c *checker.Checker, request Request, native *ast.Node) (Result, error) {
	switch r := request.(type) {
	case *NodeRequest:
		switch r.Op {
		case "type":
			return Handle(s.typeID(c.GetTypeAtLocation(native))), nil
		case "fromTypeNode":
			return Handle(s.typeID(c.GetTypeFromTypeNode(native))), nil
		case "context":
			return Handle(s.typeID(c.GetContextualType(native, checker.ContextFlagsNone))), nil
		case "symbol":
			return Handle(s.symbolID(c.GetSymbolAtLocation(native))), nil
		case "signature":
			return Handle(s.signatureID(c.GetSignatureFromDeclaration(native))), nil
		case "resolvedSignature":
			return Handle(s.signatureID(c.GetResolvedSignature(native))), nil
		}
	case *TypeRequest:
		switch r.Op {
		case "constraint":
			return Handle(s.typeID(c.GetBaseConstraintOfType(s.typeByID(r.ID)))), nil
		case "apparent":
			return Handle(s.typeID(c.GetApparentType(s.typeByID(r.ID)))), nil
		case "baseLiteral":
			return Handle(s.typeID(c.GetBaseTypeOfLiteralType(s.typeByID(r.ID)))), nil
		case "nonNullable":
			return Handle(s.typeID(c.GetNonNullableType(s.typeByID(r.ID)))), nil
		case "typeString":
			return Text(c.TypeToString(s.typeByID(r.ID))), nil
		case "properties":
			ids := Handles{}
			for _, v := range c.GetPropertiesOfType(s.typeByID(r.ID)) {
				ids = append(ids, s.symbolID(v))
			}
			return ids, nil
		case "typeArguments":
			ids := Handles{}
			for _, v := range c.GetTypeArguments(s.typeByID(r.ID)) {
				ids = append(ids, s.typeID(v))
			}
			return ids, nil
		case "isArray":
			return Boolean(c.IsArrayType(s.typeByID(r.ID))), nil
		case "isArrayLike":
			return Boolean(c.IsArrayLikeType(s.typeByID(r.ID))), nil
		case "isTuple":
			return Boolean(checker.IsTupleType(s.typeByID(r.ID))), nil
		case "numberIndexType":
			return Handle(s.typeID(checker.Checker_getIndexTypeOfType(c, s.typeByID(r.ID), c.GetNumberType()))), nil
		case "typeRecord":
			t := s.typeByID(r.ID)
			fact := &TypeFact{Flags: checker.FormatTypeFlags(t.Flags()), Symbol: s.symbolID(t.Symbol()), AliasSymbol: s.symbolID(t.Alias().Symbol())}
			if t.Flags()&checker.TypeFlagsIntrinsic != 0 {
				fact.IntrinsicName = t.AsIntrinsicType().IntrinsicName()
			}
			if t.Flags()&checker.TypeFlagsUniqueESSymbol != 0 {
				fact.EscapedName = ast.EscapeSymbolName(checker.GetPropertyNameFromType(t))
			}
			if t.Flags()&checker.TypeFlagsObject != 0 {
				fact.ObjectFlags = []string{}
				for _, flag := range publicObjectFlags {
					if t.ObjectFlags()&flag.flag != 0 {
						fact.ObjectFlags = append(fact.ObjectFlags, flag.name)
					}
				}
			}
			if t.Flags()&checker.TypeFlagsUnionOrIntersection != 0 {
				fact.Types = []uint32{}
				for _, v := range t.Types() {
					fact.Types = append(fact.Types, s.typeID(v))
				}
			}
			if t.Flags()&checker.TypeFlagsBooleanLiteral != 0 {
				fact.IntrinsicName = fmt.Sprint(t.AsLiteralType().Value())
			}
			if t.Flags()&(checker.TypeFlagsStringLiteral|checker.TypeFlagsNumberLiteral|checker.TypeFlagsBigIntLiteral) != 0 {
				switch value := t.AsLiteralType().Value().(type) {
				case string:
					fact.Literal = Literal{Kind: 1, Text: value}
				case jsnum.Number:
					fact.Literal = Literal{Kind: 2, Number: float64(value)}
				case jsnum.PseudoBigInt:
					text := value.Base10Value
					if text == "" {
						text = "0"
					}
					fact.Literal = Literal{Kind: 3, Text: text, Negative: value.Negative}
				case nil: // Computed enum members have no literal value.
				default:
					return nil, fmt.Errorf("unsupported compiler literal %T", value)
				}
			}

			return fact, nil
		}
	case *SymbolRequest:
		switch r.Op {
		case "aliasedSymbol":
			return Handle(s.symbolID(c.GetAliasedSymbol(s.symbolByID(r.ID)))), nil
		case "declaredType":
			return Handle(s.typeID(c.GetDeclaredTypeOfSymbol(s.symbolByID(r.ID)))), nil
		case "symbolRecord":
			sym := s.symbolByID(r.ID)
			return &SymbolFact{Name: ast.EscapeInternalSymbolName(ast.SymbolName(sym)), EscapedName: ast.EscapeSymbolName(sym.Name), Flags: uint32(sym.Flags), nodes: appendNodes(sym.Declarations), valueNode: sym.ValueDeclaration}, nil
		}
	case *SignatureRequest:
		switch r.Op {
		case "returnType":
			return Handle(s.typeID(c.GetReturnTypeOfSignature(s.signatureByID(r.ID)))), nil
		case "signatureRecord":
			sig := s.signatureByID(r.ID)
			ids := []uint32{}
			for _, parameter := range sig.Parameters() {
				ids = append(ids, s.symbolID(parameter))
			}
			return &SignatureFact{Parameters: ids, node: sig.Declaration()}, nil
		}
	case *VoidRequest:
		return Handle(s.typeID(c.GetVoidType())), nil
	case *PropertyRequest:
		return Handle(s.symbolID(c.GetPropertyOfType(s.typeByID(r.ID), string(r.Name)))), nil
	case *SymbolTypeRequest:
		return Handle(s.typeID(c.GetTypeOfSymbolAtLocation(s.symbolByID(r.ID), native))), nil
	case *ResolveNameRequest:
		return Handle(s.symbolID(c.ResolveName(string(r.Name), native, ast.SymbolFlags(r.Kind), r.ExcludeGlobals))), nil
	case *SignaturesRequest:
		if r.Kind > 1 {
			return nil, errors.New("invalid signature kind")
		}
		ids := Handles{}
		for _, v := range c.GetSignaturesOfType(s.typeByID(r.ID), checker.SignatureKind(r.Kind)) {
			ids = append(ids, s.signatureID(v))
		}
		return ids, nil
	}
	return nil, fmt.Errorf("unsupported type service operation %q", request.operation())
}

func appendNodes(nodes []*ast.Node) []*ast.Node {
	if nodes == nil {
		return nil
	}
	return append(make([]*ast.Node, 0, len(nodes)), nodes...)
}

package typeservice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/microsoft/TypeScript/tsc/shim/api/encoder"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/jsnum"
	"github.com/web-infra-dev/rslint/internal/languages/js/estree"
	"github.com/web-infra-dev/rslint/internal/program"
)

var nextGeneration atomic.Uint32
var ErrClosed = errors.New("type service session is closed")

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
type Projection struct {
	SourceFileName string `json:"sourceFileName"`
	Frame          []byte `json:"frame"`
	Tree           []byte `json:"tree"`
	// Origin IDs are exact encoder indices. A zero denotes a synthetic record
	// whose compiler node is deliberately unavailable; consumers must reject it.
	Origins    []uint32          `json:"origins"`
	Reverse    map[uint32]uint32 `json:"reverse"`
	Generation uint32            `json:"generation"`
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

func (s *Session) Project(ctx context.Context, file *ast.SourceFile, options estree.Options) (Projection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.program == nil {
		return Projection{}, ErrClosed
	}
	if !s.eligible[file] {
		return Projection{}, errors.New("source is not eligible for typed rules")
	}
	tree, table, err := encoder.EncodeSourceFile(file)
	if err != nil {
		return Projection{}, err
	}
	result := Projection{SourceFileName: file.FileName(), Tree: tree, Generation: s.generation, Origins: []uint32{0}, Reverse: make(map[uint32]uint32)}
	options.RecordOrigin = func(id uint32, node *ast.Node) {
		if int(id) != len(result.Origins) {
			panic("unordered projection origins")
		}
		index := uint32(0)
		if node != nil {
			index = table.GetIndex(node)
		}
		result.Origins = append(result.Origins, index)

	}
	options.RecordReverse = func(node *ast.Node, id uint32) {
		if index := table.GetIndex(node); index != 0 {
			result.Reverse[index] = id
		}
	}
	var b estree.Builder
	result.Frame, err = b.Build(ctx, file, options)
	return result, err
}

// Query carries every semantic argument. ID is scoped to Generation, Checker
// and the operation's object kind. File/Index identify original Program nodes.
type Query struct {
	Generation     uint32 `json:"generation"`
	Checker        uint32 `json:"checker"`
	Op             string `json:"op"`
	ID             uint32 `json:"id,omitempty"`
	File           string `json:"file,omitempty"`
	Index          uint32 `json:"index,omitempty"`
	Name           string `json:"name,omitempty"`
	Kind           uint32 `json:"kind,omitempty"`
	ExcludeGlobals bool   `json:"excludeGlobals,omitempty"`
}

// Query returns a binary value or an explicit error. Cached JS/shared facts
// have the same ownership and must be invalidated when their session closes.
func (s *Session) Query(ctx context.Context, q Query) (packet []byte, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.program == nil {
		return nil, ErrClosed
	}
	if q.Generation != s.generation || q.Checker != 1 {
		return nil, errors.New("stale or foreign type service handle")
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
	// Upstream compiler assertions remain failures, never empty query results.
	defer func() {
		if value := recover(); value != nil {
			packet = nil
			err = fmt.Errorf("type service %s: %v", q.Op, value)
		}
	}()
	value, err := s.execute(c, q)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return EncodePacket(value)
}
func (s *Session) node(q Query) *ast.Node {
	f := s.program.GetSourceFile(q.File)
	if f == nil {
		panic("foreign source handle")
	}
	table := encoder.GetNodeIndexTable(f)
	if q.Index == 0 || int(q.Index) >= len(table.Nodes) || table.Nodes[q.Index] == nil {
		panic("invalid compiler node handle")
	}
	return table.Nodes[q.Index]
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
func (s *Session) nodeRef(n *ast.Node) any {
	if n == nil {
		return nil
	}
	f := ast.GetSourceFileOfNode(n)
	if !s.program.OwnsSourceFile(f) {
		panic("declaration is outside Program")
	}
	return map[string]any{"file": f.FileName(), "index": encoder.GetNodeIndexTable(f).GetIndex(n)}
}
func (s *Session) execute(c *checker.Checker, q Query) (any, error) {
	switch q.Op {
	case "source":
		f := s.program.GetSourceFile(q.File)
		if f == nil {
			return nil, nil //nolint:nilnil // TypeScript getSourceFile returns undefined for a missing file.
		}
		tree, _, err := encoder.EncodeSourceFile(f)
		if err != nil {
			return nil, err
		}
		return map[string]any{"tree": tree, "fileName": f.FileName()}, nil
	case "type":
		return s.typeID(c.GetTypeAtLocation(s.node(q))), nil
	case "fromTypeNode":
		return s.typeID(c.GetTypeFromTypeNode(s.node(q))), nil
	case "context":
		return s.typeID(c.GetContextualType(s.node(q), checker.ContextFlagsNone)), nil
	case "void":
		return s.typeID(c.GetVoidType()), nil
	case "constraint":
		return s.typeID(c.GetBaseConstraintOfType(s.typeByID(q.ID))), nil
	case "apparent":
		return s.typeID(c.GetApparentType(s.typeByID(q.ID))), nil
	case "baseLiteral":
		return s.typeID(c.GetBaseTypeOfLiteralType(s.typeByID(q.ID))), nil
	case "nonNullable":
		return s.typeID(c.GetNonNullableType(s.typeByID(q.ID))), nil
	case "typeString":
		return c.TypeToString(s.typeByID(q.ID)), nil
	case "symbol":
		return s.symbolID(c.GetSymbolAtLocation(s.node(q))), nil
	case "resolveName":
		var location *ast.Node
		if q.File != "" {
			location = s.node(q)
		}
		return s.symbolID(c.ResolveName(q.Name, location, ast.SymbolFlags(q.Kind), q.ExcludeGlobals)), nil
	case "aliasedSymbol":
		return s.symbolID(c.GetAliasedSymbol(s.symbolByID(q.ID))), nil
	case "symbolType":
		return s.typeID(c.GetTypeOfSymbolAtLocation(s.symbolByID(q.ID), s.node(q))), nil
	case "declaredType":
		return s.typeID(c.GetDeclaredTypeOfSymbol(s.symbolByID(q.ID))), nil
	case "property":
		return s.symbolID(c.GetPropertyOfType(s.typeByID(q.ID), q.Name)), nil
	case "properties":
		ids := []uint32{}
		for _, v := range c.GetPropertiesOfType(s.typeByID(q.ID)) {
			ids = append(ids, s.symbolID(v))
		}
		return ids, nil
	case "typeArguments":
		ids := []uint32{}
		for _, v := range c.GetTypeArguments(s.typeByID(q.ID)) {
			ids = append(ids, s.typeID(v))
		}
		return ids, nil
	case "isArray":
		return c.IsArrayType(s.typeByID(q.ID)), nil
	case "isTuple":
		return checker.IsTupleType(s.typeByID(q.ID)), nil
	case "signature":
		return s.signatureID(c.GetSignatureFromDeclaration(s.node(q))), nil
	case "resolvedSignature":
		return s.signatureID(c.GetResolvedSignature(s.node(q))), nil
	case "returnType":
		return s.typeID(c.GetReturnTypeOfSignature(s.signatureByID(q.ID))), nil
	case "signatures":
		if q.Kind > 1 {
			return nil, errors.New("invalid signature kind")
		}
		ids := []uint32{}
		for _, v := range c.GetSignaturesOfType(s.typeByID(q.ID), checker.SignatureKind(q.Kind)) {
			ids = append(ids, s.signatureID(v))
		}
		return ids, nil
	case "typeRecord":
		t := s.typeByID(q.ID)
		r := map[string]any{"flags": checker.FormatTypeFlags(t.Flags()), "symbol": s.symbolID(t.Symbol()), "aliasSymbol": s.symbolID(t.Alias().Symbol())}
		if t.Flags()&checker.TypeFlagsIntrinsic != 0 {
			r["intrinsicName"] = t.AsIntrinsicType().IntrinsicName()
		}
		if t.Flags()&checker.TypeFlagsUnionOrIntersection != 0 {
			ids := []uint32{}
			for _, v := range t.Types() {
				ids = append(ids, s.typeID(v))
			}
			r["types"] = ids
		}
		if t.Flags()&(checker.TypeFlagsStringLiteral|checker.TypeFlagsNumberLiteral|checker.TypeFlagsBigIntLiteral) != 0 {
			value := t.AsLiteralType().Value()
			if bigint, ok := value.(jsnum.PseudoBigInt); ok {
				base10 := bigint.Base10Value
				if base10 == "" {
					base10 = "0"
				}
				value = map[string]any{"negative": bigint.Negative, "base10Value": base10}
			}
			r["value"] = value

		}
		if t.Flags()&checker.TypeFlagsBooleanLiteral != 0 {
			r["intrinsicName"] = fmt.Sprint(t.AsLiteralType().Value())
		}
		return r, nil
	case "symbolRecord":
		sym := s.symbolByID(q.ID)
		var declarations any
		if sym.Declarations != nil {
			decls := make([]any, 0, len(sym.Declarations))
			for _, n := range sym.Declarations {
				decls = append(decls, s.nodeRef(n))
			}
			declarations = decls
		}
		return map[string]any{"name": ast.EscapeInternalSymbolName(ast.SymbolName(sym)), "escapedName": ast.EscapeSymbolName(sym.Name), "flags": uint32(sym.Flags), "declarations": declarations, "valueDeclaration": s.nodeRef(sym.ValueDeclaration)}, nil
	case "signatureRecord":
		sig := s.signatureByID(q.ID)
		ids := []uint32{}
		for _, p := range sig.Parameters() {
			ids = append(ids, s.symbolID(p))
		}
		return map[string]any{"parameters": ids, "declaration": s.nodeRef(sig.Declaration())}, nil
	default:
		return nil, fmt.Errorf("unsupported type service operation %q", strings.TrimSpace(q.Op))
	}
}

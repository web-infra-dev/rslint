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
	FileName          string   `json:"fileName"`
	Path              string   `json:"path"`
	Packages          []string `json:"packages"`
	External          bool     `json:"external"`
	DefaultLibrary    bool     `json:"defaultLibrary"`
	IsDeclarationFile bool     `json:"isDeclarationFile"`
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
	return Source{FileName: f.FileName(), Path: string(f.Path()), Packages: s.program.PackageNamesForSourceFile(f), External: s.program.IsSourceFileFromExternalLibrary(f), DefaultLibrary: s.program.IsSourceFileDefaultLibrary(f), IsDeclarationFile: f.IsDeclarationFile}
}

// Query returns one authoritative answer, with shallow facts for its direct
// handles. Callers batch known misses to amortize the transport and checker lease.
func (s *Session) Query(ctx context.Context, q Query) ([]byte, error) {
	packets, err := s.QueryBatch(ctx, []Query{q})
	if err != nil {
		return nil, err
	}
	return packets[0], nil
}

// QueryBatch resolves syntax before one short exclusive checker lease. Every
// query is authoritative; no listener replay or semantic prediction is involved.
func (s *Session) QueryBatch(ctx context.Context, queries []Query) (packets [][]byte, err error) {
	defer func() {
		if value := recover(); value != nil {
			packets = nil
			err = fmt.Errorf("type service: %v", value)
		}
	}()
	if len(queries) > 256 {
		return nil, errors.New("type service batch exceeds 256 queries")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.Lock()
	if s.program == nil {
		s.mu.Unlock()
		return nil, ErrClosed
	}
	p := s.program
	for _, q := range queries {
		if q.Generation != s.generation || q.Checker != 1 {
			s.mu.Unlock()
			return nil, errors.New("stale or foreign type service handle")
		}
		if q.Request == nil {
			s.mu.Unlock()
			return nil, errors.New("missing type service request")
		}
	}
	s.mu.Unlock()
	natives := make([]*ast.Node, len(queries))
	for i, q := range queries {
		if node := requestNode(q.Request); node != nil {
			f := p.GetSourceFile(node.File)
			if f == nil {
				return nil, errors.New("foreign source handle")
			}
			table := encoder.GetNodeIndexTable(f)
			if node.Index == 0 || int(node.Index) >= len(table.Nodes) || table.Nodes[node.Index] == nil {
				return nil, errors.New("invalid compiler node handle")
			}
			natives[i] = table.Nodes[node.Index]
		}
	}
	results := make([]Result, len(queries))
	records := make([][]Record, len(queries))
	semantic := false
	for _, query := range queries {
		if query.Request.operation() != "nodeRecord" {
			semantic = true
			break
		}
	}
	if semantic {
		err = s.checkBatch(ctx, queries, natives, results, records)
		if err != nil {
			return nil, err
		}
	}
	packets = make([][]byte, len(queries))
	for i, result := range results {
		if queries[i].Request.operation() == "nodeRecord" {
			result = nodeFacts(p, natives[i])
		}
		materializeReferences(p, result)
		for _, record := range records[i] {
			materializeReferences(p, record.Fact)
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		packets[i], err = EncodeReply(result, records[i])
		if err != nil {
			return nil, err
		}
	}
	return packets, nil
}
func materializeReferences(p *program.Program, result Result) {
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
	case IndexInfos:
		for i := range fact {
			fact[i].Declaration = nodeReference(p, fact[i].node)
			fact[i].node = nil
		}
	}
}
func (s *Session) checkBatch(ctx context.Context, queries []Query, natives []*ast.Node, results []Result, records [][]Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.program == nil {
		return ErrClosed
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	c, done := s.program.TypeCheckerForFileExclusive(core.WithCheckerLifetime(ctx, core.CheckerLifetimeAPI), s.anchor)
	defer done()
	if c != s.checker {
		return errors.New("type service checker identity changed")
	}
	for i, q := range queries {
		if err := ctx.Err(); err != nil {
			return err
		}
		if q.Request.operation() == "nodeRecord" {
			continue
		}
		var err error
		results[i], err = s.execute(c, q.Request, natives[i])
		if err != nil {
			return err
		}
		records[i], err = s.directRecords(c, q.Request.operation(), results[i])
		if err != nil {
			return err
		}
	}
	return nil
}
func (s *Session) directRecords(c *checker.Checker, op Operation, result Result) ([]Record, error) {
	var kind Operation
	switch op {
	case "type", "fromTypeNode", "context", "constraint", "apparent", "baseLiteral", "nonNullable", "void", "declaredType", "symbolType", "returnType", "numberIndexType", "typeArguments",
		"stringIndexType", "awaited", "widened", "baseTypes", "defaultType", "typeOfSymbol", "indexInfo", "indexInfos", "typePredicate", "parameterType", "propertyType":
		kind = "typeRecord"
	case "symbol", "aliasedSymbol", "property", "properties", "resolveName",
		"apparentProperties", "exportSymbol", "immediateAliasedSymbol", "shorthandValueSymbol":
		kind = "symbolRecord"
	case "signature", "resolvedSignature", "signatures":
		kind = "signatureRecord"
	default:
		return nil, nil
	}
	var ids []uint32
	switch value := result.(type) {
	case Handle:
		ids = []uint32{uint32(value)}
	case Handles:
		ids = value
	case IndexInfos:
		ids = make([]uint32, 0, 2*len(value))
		for _, info := range value {
			ids = append(ids, info.KeyType, info.Type)
		}
	case *TypePredicateFact:
		if value != nil {
			ids = []uint32{value.Type}
		}
	}
	records := []Record{}
	seen := make(map[[2]uint32]bool)
	var add func(Operation, uint32) error
	add = func(kind Operation, id uint32) error {
		if id == 0 {
			return nil
		}
		tag := uint32(1)
		if kind == "symbolRecord" {
			tag = 2
		}
		if kind == "signatureRecord" {
			tag = 3
		}
		key := [2]uint32{tag, id}
		if seen[key] {
			return nil
		}
		seen[key] = true
		var request Request
		switch kind {
		case "typeRecord":
			request = &TypeRequest{Op: kind, ID: id}
		case "symbolRecord":
			request = &SymbolRequest{Op: kind, ID: id}
		case "signatureRecord":
			request = &SignatureRequest{Op: kind, ID: id}
		}
		fact, err := s.execute(c, request, nil)
		if err != nil {
			return err
		}
		records = append(records, Record{kind, id, fact})
		// Only directly referenced symbols are included, never member/type graphs.
		switch value := fact.(type) {
		case *TypeFact:
			if err := add("symbolRecord", value.Symbol); err != nil {
				return err
			}
			return add("symbolRecord", value.AliasSymbol)
		case *SignatureFact:
			for _, parameter := range value.Parameters {
				if err := add("symbolRecord", parameter); err != nil {
					return err
				}
			}
			return add("symbolRecord", value.ThisParameter)
		}
		return nil
	}
	for _, id := range ids {
		if err := add(kind, id); err != nil {
			return nil, err
		}
	}
	return records, nil
}

// Only the requested declaration's ancestors and declaration names are copied.
// Descendant syntax stays demand-loaded through the immutable source transport.
func nodeFacts(p *program.Program, node *ast.Node) NodeFacts {
	facts := NodeFacts{}
	seen := map[*ast.Node]bool{}
	add := func(n *ast.Node) {
		if n == nil || seen[n] {
			return
		}
		seen[n] = true
		fact := NodeFact{Node: *nodeReference(p, n), Kind: strings.TrimPrefix(n.Kind.String(), "Kind"), Flags: uint32(n.Flags), Parent: nodeReference(p, n.Parent)}
		if ast.IsIdentifier(n) || ast.IsPrivateIdentifier(n) || ast.IsStringLiteral(n) || ast.IsNumericLiteral(n) {
			text := n.Text()
			fact.Text = &text
		}
		if n.Name() != nil {
			fact.Name = nodeReference(p, n.Name())
		}
		facts = append(facts, fact)
	}
	for current := node; current != nil; current = current.Parent {
		add(current)
		if current.Name() != nil {
			add(current.Name())
		}
	}
	return facts
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
func (s *Session) typeIDs(types []*checker.Type) Handles {
	if types == nil {
		return nil
	}
	ids := make(Handles, len(types))
	for i, t := range types {
		ids[i] = s.typeID(t)
	}
	return ids
}
func (s *Session) symbolIDs(symbols []*ast.Symbol) Handles {
	ids := make(Handles, len(symbols))
	for i, symbol := range symbols {
		ids[i] = s.symbolID(symbol)
	}
	return ids
}
func symbolName(symbol *ast.Symbol) string {
	return ast.EscapeInternalSymbolName(ast.SymbolName(symbol))
}
func (s *Session) namedSymbols(symbols []*ast.Symbol) NamedSymbols {
	result := make(NamedSymbols, len(symbols))
	for i, symbol := range symbols {
		result[i] = NamedSymbol{ID: s.symbolID(symbol), Name: symbolName(symbol)}
	}
	return result
}
func (s *Session) indexInfo(info *checker.IndexInfo) IndexInfo {
	return IndexInfo{
		KeyType:    s.typeID(info.KeyType()),
		Type:       s.typeID(info.ValueType()),
		IsReadonly: info.IsReadonly(),
		node:       info.Declaration(),
	}
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
		case "shorthandValueSymbol":
			return Handle(s.symbolID(c.GetShorthandAssignmentValueSymbol(native))), nil
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
		case "awaited":
			return Handle(s.typeID(checker.Checker_getAwaitedType(c, s.typeByID(r.ID)))), nil
		case "widened":
			return Handle(s.typeID(c.GetWidenedType(s.typeByID(r.ID)))), nil
		case "baseTypes":
			t := s.typeByID(r.ID)
			if t.ObjectFlags()&checker.ObjectFlagsReference != 0 {
				t = t.Target()
			}
			return s.typeIDs(c.GetBaseTypes(t)), nil
		case "defaultType":
			return Handle(s.typeID(c.GetDefaultFromTypeParameter(s.typeByID(r.ID)))), nil
		case "properties":
			return s.symbolIDs(c.GetPropertiesOfType(s.typeByID(r.ID))), nil
		case "apparentProperties":
			return s.symbolIDs(c.GetApparentProperties(s.typeByID(r.ID))), nil
		case "typeArguments":
			return s.typeIDs(c.GetTypeArguments(s.typeByID(r.ID))), nil
		case "isArray":
			return Boolean(c.IsArrayType(s.typeByID(r.ID))), nil
		case "isArrayLike":
			return Boolean(c.IsArrayLikeType(s.typeByID(r.ID))), nil
		case "isTuple":
			return Boolean(checker.IsTupleType(s.typeByID(r.ID))), nil
		case "numberIndexType":
			return Handle(s.typeID(c.GetNumberIndexType(s.typeByID(r.ID)))), nil
		case "stringIndexType":
			return Handle(s.typeID(c.GetStringIndexType(s.typeByID(r.ID)))), nil
		case "indexInfos":
			infos := c.GetIndexInfosOfType(s.typeByID(r.ID))
			result := make(IndexInfos, len(infos))
			for i, info := range infos {
				result[i] = s.indexInfo(info)
			}
			return result, nil
		case "typeRecord":
			t := s.typeByID(r.ID)
			flags := t.Flags()
			fact := &TypeFact{
				Flags:              checker.FormatTypeFlags(flags),
				Symbol:             s.symbolID(t.Symbol()),
				AliasSymbol:        s.symbolID(t.Alias().Symbol()),
				AliasTypeArguments: s.typeIDs(t.Alias().TypeArguments()),
			}
			if flags&checker.TypeFlagsIntrinsic != 0 {
				fact.IntrinsicName = t.AsIntrinsicType().IntrinsicName()
			}
			if flags&checker.TypeFlagsUniqueESSymbol != 0 {
				fact.EscapedName = ast.EscapeSymbolName(checker.GetPropertyNameFromType(t))
			}
			if flags&checker.TypeFlagsObject != 0 {
				objectFlags := t.ObjectFlags()
				fact.ObjectFlags = []string{}
				for _, flag := range publicObjectFlags {
					if objectFlags&flag.flag != 0 {
						fact.ObjectFlags = append(fact.ObjectFlags, flag.name)
					}
				}
				if objectFlags&checker.ObjectFlagsReference != 0 {
					fact.Target = s.typeID(t.Target())
				}
				if objectFlags&(checker.ObjectFlagsClassOrInterface|checker.ObjectFlagsTuple) != 0 {
					if parameters := t.AsInterfaceType().TypeParameters(); len(parameters) > 0 {
						fact.TypeParameters = s.typeIDs(parameters)
					}
				}
				if objectFlags&checker.ObjectFlagsTuple != 0 {
					tuple := t.AsTupleType()
					fact.Tuple = &TupleFact{
						Readonly:      tuple.IsReadonly(),
						CombinedFlags: uint32(checker.TupleType_combinedFlags(tuple)),
					}
				}
			}
			switch {
			case flags&(checker.TypeFlagsUnionOrIntersection|checker.TypeFlagsTemplateLiteral) != 0:
				fact.TypeOperands = s.typeIDs(t.Types())
			case flags&(checker.TypeFlagsIndex|checker.TypeFlagsStringMapping) != 0:
				fact.TypeOperands = []uint32{s.typeID(t.Target())}
			case flags&checker.TypeFlagsIndexedAccess != 0:
				indexed := t.AsIndexedAccessType()
				fact.TypeOperands = []uint32{s.typeID(indexed.ObjectType()), s.typeID(indexed.IndexType())}
			case flags&checker.TypeFlagsConditional != 0:
				conditional := t.AsConditionalType()
				fact.TypeOperands = []uint32{s.typeID(conditional.CheckType()), s.typeID(conditional.ExtendsType())}
			}
			if flags&checker.TypeFlagsBooleanLiteral != 0 {
				fact.IntrinsicName = fmt.Sprint(t.AsLiteralType().Value())
			}
			if flags&(checker.TypeFlagsStringLiteral|checker.TypeFlagsNumberLiteral|checker.TypeFlagsBigIntLiteral) != 0 {
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
		case "immediateAliasedSymbol":
			return Handle(s.symbolID(c.GetImmediateAliasedSymbol(s.symbolByID(r.ID)))), nil
		case "isUnknownSymbol":
			return Boolean(c.IsUnknownSymbol(s.symbolByID(r.ID))), nil
		case "exportsOfModule":
			return s.namedSymbols(c.GetExportsOfModule(s.symbolByID(r.ID))), nil
		case "exportSymbol":
			return Handle(s.symbolID(c.GetExportSymbolOfSymbol(s.symbolByID(r.ID)))), nil
		case "declaredType":
			return Handle(s.typeID(c.GetDeclaredTypeOfSymbol(s.symbolByID(r.ID)))), nil
		case "typeOfSymbol":
			return Handle(s.typeID(c.GetTypeOfSymbol(s.symbolByID(r.ID)))), nil
		case "symbolRecord":
			sym := s.symbolByID(r.ID)
			return &SymbolFact{Name: symbolName(sym), EscapedName: ast.EscapeSymbolName(sym.Name), Flags: uint32(sym.Flags), nodes: appendNodes(sym.Declarations), valueNode: sym.ValueDeclaration}, nil
		}
	case *SignatureRequest:
		switch r.Op {
		case "returnType":
			return Handle(s.typeID(c.GetReturnTypeOfSignature(s.signatureByID(r.ID)))), nil
		case "typePredicate":
			predicate := c.GetTypePredicateOfSignature(s.signatureByID(r.ID))
			if predicate == nil {
				return (*TypePredicateFact)(nil), nil
			}
			fact := &TypePredicateFact{Kind: uint32(predicate.Kind()), Type: s.typeID(predicate.Type())}
			if predicate.Kind() == checker.TypePredicateKindIdentifier || predicate.Kind() == checker.TypePredicateKindAssertsIdentifier {
				name := predicate.ParameterName()
				fact.ParameterName = &name
				fact.ParameterIndex = predicate.ParameterIndex()
			}
			return fact, nil
		case "signatureRecord":
			sig := s.signatureByID(r.ID)
			return &SignatureFact{
				Parameters:       s.symbolIDs(sig.Parameters()),
				TypeParameters:   s.typeIDs(sig.TypeParameters()),
				ThisParameter:    s.symbolID(sig.ThisParameter()),
				MinArgumentCount: sig.MinArgumentCount(),
				node:             sig.Declaration(),
			}, nil
		}
	case *VoidRequest:
		return Handle(s.typeID(c.GetVoidType())), nil
	case *PropertyRequest:
		switch r.Op {
		case "property":
			return Handle(s.symbolID(c.GetPropertyOfType(s.typeByID(r.ID), string(r.Name)))), nil
		case "propertyType":
			return Handle(s.typeID(c.GetTypeOfPropertyOfType(s.typeByID(r.ID), string(r.Name)))), nil
		}
	case *SymbolTypeRequest:
		return Handle(s.typeID(c.GetTypeOfSymbolAtLocation(s.symbolByID(r.ID), native))), nil
	case *ResolveNameRequest:
		return Handle(s.symbolID(c.ResolveName(string(r.Name), native, ast.SymbolFlags(r.Kind), r.ExcludeGlobals))), nil
	case *AssignableRequest:
		return Boolean(c.IsTypeAssignableTo(s.typeByID(r.Source), s.typeByID(r.Target))), nil
	case *IndexInfoRequest:
		key := c.GetNumberType()
		if r.Kind == 0 {
			key = c.GetStringType()
		}
		info := c.GetIndexInfoOfType(s.typeByID(r.ID), key)
		if info == nil {
			return IndexInfos{}, nil
		}
		return IndexInfos{s.indexInfo(info)}, nil
	case *ScopeSymbolsRequest:
		return s.namedSymbols(c.GetSymbolsInScope(native, ast.SymbolFlags(r.Meaning))), nil
	case *ParameterTypeRequest:
		return Handle(s.typeID(c.GetTypeParameterAtPosition(s.signatureByID(r.ID), r.Position))), nil
	case *PrintRequest:
		flags := checker.TypeFormatFlagsNone
		if r.Flags != nil {
			flags = checker.TypeFormatFlags(*r.Flags)
		} else if r.Op == "typeString" {
			flags = checker.TypeFormatFlagsAllowUniqueESSymbolType | checker.TypeFormatFlagsUseAliasDefinedOutsideCurrentScope
		}
		if flags&checker.TypeFormatFlagsMultilineObjectLiterals != 0 {
			return nil, errors.New("unsupported TypeScript operation: TypeFormatFlags.MultilineObjectLiterals")
		}
		switch r.Op {
		case "typeString":
			return Text(c.TypeToStringEx(s.typeByID(r.ID), native, flags, nil)), nil
		case "signatureString":
			sig := s.signatureByID(r.ID)
			if r.Kind != uint32(checker.SignatureKindConstruct) {
				flags |= checker.TypeFormatFlagsWriteCallStyleSignature
			}
			text := c.SignatureToStringEx(sig, native, flags, nil)
			if r.Kind == uint32(checker.SignatureKindConstruct) && sig.Flags()&checker.SignatureFlagsConstruct == 0 {
				// TypeScript also accepts Construct when rendering a call signature.
				text = "new " + text
			}
			return Text(text), nil
		}
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

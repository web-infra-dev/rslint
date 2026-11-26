package typeservice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"math"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/api/encoder"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/jsnum"
	"github.com/microsoft/TypeScript/tsc/shim/locale"
	"github.com/microsoft/TypeScript/tsc/shim/nodebuilder"
	"github.com/microsoft/TypeScript/tsc/shim/printer"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var nextGeneration atomic.Uint32
var (
	ErrClosed   = errors.New("type service session is closed")
	ErrPoisoned = errors.New("type service session is unusable after an internal panic")
)

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

// Session owns a Program view and checker handle tables until Close.
// One anchored checker supplies all objects, including cross-file declarations.
// Exclusive checker access is acquired and released per query batch.
//
// Checker lease contract:
//   - On project/LSP pools (core.CheckerLifetimeAPI), tsgo routes leases to a
//     dedicated persistent checker that does not contend with diagnostic or
//     query checkers. If that checker is later replaced (e.g. after a canceled
//     request), QueryBatch returns "type service checker identity changed" and
//     the caller must close this Session and create a new generation/session.
//   - On CLI compiler pools (compiler.checkerPool), TypeCheckerForFileExclusive
//     locks the anchor file's shard mutex directly. Callers MUST NOT invoke
//     QueryBatch concurrently with (or from within) native lint shards
//     holding exclusive checker leases on the same Program; JS typed linting
//     must run in its own phase outside native shards.
//
// Locking order: lifeMu (RLock for active operations, Lock for Close) is always
// acquired before checkMu (serializing exclusive checker access and handle maps).
type Session struct {
	lifeMu       sync.RWMutex
	checkMu      sync.Mutex
	extraMu      sync.RWMutex
	poisoned     atomic.Bool
	program      *program.Program
	syntaxOnly   bool
	anchor       *ast.SourceFile
	eligible     map[*ast.SourceFile]bool
	extraSources map[string]*ast.SourceFile
	checker      *checker.Checker
	generation   uint32
	types        map[uint32]*checker.Type
	symbols      map[uint32]*ast.Symbol
	signatures   map[uint32]*checker.Signature
	indexInfoIDs map[*checker.IndexInfo]uint32
}

type Descriptor struct {
	SyntaxOnly        bool            `json:"syntaxOnly,omitempty"`
	Generation        uint32          `json:"generation"`
	CurrentDirectory  string          `json:"currentDirectory"`
	CompilerOptions   json.RawMessage `json:"compilerOptions"`
	Sources           []Source        `json:"sources"`
	DiagnosticSources []Source        `json:"diagnosticSources,omitempty"`
}

type Source struct {
	ImpliedNodeFormat core.ResolutionMode `json:"impliedNodeFormat,omitempty"`
	CheckJSDirective  *checkJSDirective   `json:"checkJsDirective,omitempty"`
	FileName          string              `json:"fileName"`
	Path              string              `json:"path"`
	Packages          []string            `json:"packages"`
	External          bool                `json:"external"`
	DefaultLibrary    bool                `json:"defaultLibrary"`
	IsDeclarationFile bool                `json:"isDeclarationFile"`
}

type checkJSDirective struct {
	Enabled bool `json:"enabled"`
	End     int  `json:"end"`
	Pos     int  `json:"pos"`
}

// Binding authorizes one eligible lint target. Dependency nodes remain valid
// semantic operands; binding identity is not part of immutable fact cache keys.
type Binding struct {
	Scope
	SourceFileName string `json:"sourceFileName"`
}

// New accepts exact Program sources explicitly admitted for typed rules.
// Checker capability alone is never eligibility.
func New(ctx context.Context, p *program.Program, eligible []*ast.SourceFile) (*Session, error) {
	return newSession(ctx, p, eligible, false)
}

// NewSyntaxOnly exposes the admitted files' compiler syntax without authorizing
// a type checker, including when the underlying Program also serves typed files.
func NewSyntaxOnly(ctx context.Context, p *program.Program, eligible []*ast.SourceFile) (*Session, error) {
	return newSession(ctx, p, eligible, true)
}

func newSession(ctx context.Context, p *program.Program, eligible []*ast.SourceFile, syntaxOnly bool) (*Session, error) {
	if !p.IsValid() || len(eligible) == 0 {
		return nil, errors.New("type services require a bound Program and explicitly eligible sources")
	}
	s := &Session{
		program:    p,
		syntaxOnly: syntaxOnly,
		anchor:     eligible[0],
		eligible:   make(map[*ast.SourceFile]bool, len(eligible)),
		generation: nextGeneration.Add(1),
	}
	for _, f := range eligible {
		if !p.OwnsSourceFile(f) || (!syntaxOnly && !p.CanProvideTypeChecker(f)) {
			return nil, errors.New("ineligible or foreign type service source")
		}
		s.eligible[f] = true
	}
	if !syntaxOnly {
		if q := p.CompilerQueries(); q != nil {
			if cmd := q.CommandLine(); cmd != nil && cmd.ConfigFile != nil && cmd.ConfigFile.SourceFile != nil {
				s.recordExtraSource(cmd.ConfigFile.SourceFile)
			}
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Session) recordExtraSource(f *ast.SourceFile) {
	if f == nil {
		return
	}
	s.extraMu.Lock()
	defer s.extraMu.Unlock()
	if s.extraSources == nil {
		s.extraSources = make(map[string]*ast.SourceFile)
	}
	s.extraSources[f.FileName()] = f
	if path := string(f.Path()); path != "" {
		s.extraSources[path] = f
	}
	if norm := tspath.NormalizePath(f.FileName()); norm != "" {
		s.extraSources[norm] = f
	}
}

func (s *Session) extraSource(name string) *ast.SourceFile {
	s.extraMu.RLock()
	defer s.extraMu.RUnlock()
	if s.extraSources == nil {
		return nil
	}
	if f := s.extraSources[name]; f != nil {
		return f
	}
	return s.extraSources[tspath.NormalizePath(name)]
}

func (s *Session) sourceFile(name string) *ast.SourceFile {
	if f := s.program.GetSourceFile(name); f != nil {
		return f
	}
	if !s.syntaxOnly {
		return s.extraSource(name)
	}
	return nil
}

func (s *Session) diagnosticSources() []Source {
	s.extraMu.RLock()
	defer s.extraMu.RUnlock()
	if len(s.extraSources) == 0 {
		return nil
	}
	seen := make(map[*ast.SourceFile]bool, len(s.extraSources))
	files := make([]*ast.SourceFile, 0, len(s.extraSources))
	for _, f := range s.extraSources {
		if f != nil && !seen[f] {
			seen[f] = true
			files = append(files, f)
		}
	}
	slices.SortFunc(files, func(a, b *ast.SourceFile) int {
		return strings.Compare(a.FileName(), b.FileName())
	})
	sources := make([]Source, len(files))
	for i, f := range files {
		sources[i] = syntaxSource(f)
	}
	return sources
}

// Close waits for any in-flight operations to finish before releasing references.
func (s *Session) Close() {
	s.lifeMu.Lock()
	defer s.lifeMu.Unlock()
	s.program = nil
	s.anchor = nil
	s.checker = nil
	s.eligible = nil
	s.extraMu.Lock()
	s.extraSources = nil
	s.extraMu.Unlock()
	s.types = nil
	s.symbols = nil
	s.signatures = nil
	s.indexInfoIDs = nil
}

func (s *Session) BindTarget(file *ast.SourceFile) (Binding, error) {
	s.lifeMu.RLock()
	defer s.lifeMu.RUnlock()
	if s.program == nil {
		return Binding{}, ErrClosed
	}
	if !s.eligible[file] {
		return Binding{}, errors.New("source is not eligible for typed rules")
	}
	return Binding{Scope{Generation: s.generation}, file.FileName()}, nil
}

// Source returns the exact immutable Program source for a valid generation.
func (s *Session) Source(generation uint32, name string) (*ast.SourceFile, error) {
	s.lifeMu.RLock()
	defer s.lifeMu.RUnlock()
	if s.program == nil {
		return nil, ErrClosed
	}
	if generation != s.generation {
		return nil, errors.New("stale type service generation")
	}
	file := s.sourceFile(name)
	if s.syntaxOnly && !s.eligible[file] {
		return nil, errors.New("foreign syntax service source")
	}
	return file, nil //nolint:nilnil // A missing Program source is a legitimate result.
}

// SyntaxDescriptor opens only the requested target for a file-scoped reader.
func (s *Session) SyntaxDescriptor(file *ast.SourceFile) (Descriptor, error) {
	s.lifeMu.RLock()
	defer s.lifeMu.RUnlock()
	if s.program == nil {
		return Descriptor{}, ErrClosed
	}
	if !s.syntaxOnly || !s.eligible[file] {
		return Descriptor{}, errors.New("foreign syntax projection target")
	}
	d, err := s.syntaxDescriptor()
	if err != nil {
		return Descriptor{}, err
	}
	d.Sources = []Source{syntaxSource(file)}
	return d, nil
}

func (s *Session) syntaxDescriptor() (Descriptor, error) {
	options := json.RawMessage("{}")
	if opts := s.program.Options(); opts != nil {
		var err error
		options, err = json.Marshal(compilerOptions(opts))
		if err != nil {
			return Descriptor{}, err
		}
	}
	return Descriptor{
		SyntaxOnly:       true,
		Generation:       s.generation,
		CurrentDirectory: s.program.CurrentDirectory(),
		CompilerOptions:  options,
	}, nil
}

func syntaxSource(f *ast.SourceFile) Source {
	source := Source{
		FileName:          f.FileName(),
		Path:              string(f.Path()),
		IsDeclarationFile: f.IsDeclarationFile,
	}
	if directive := f.CheckJsDirective; directive != nil {
		positions := f.GetPositionMap()
		source.CheckJSDirective = &checkJSDirective{
			Enabled: directive.Enabled,
			End:     positions.UTF8ToUTF16(directive.Range.End()),
			Pos:     positions.UTF8ToUTF16(directive.Range.Pos()),
		}
	}
	return source
}

func (s *Session) Descriptor() (Descriptor, error) {
	s.lifeMu.RLock()
	defer s.lifeMu.RUnlock()
	if s.program == nil {
		return Descriptor{}, ErrClosed
	}
	if s.syntaxOnly {
		d, err := s.syntaxDescriptor()
		if err != nil {
			return Descriptor{}, err
		}
		for _, file := range s.program.SourceFiles() {
			if s.eligible[file] {
				d.Sources = append(d.Sources, syntaxSource(file))
			}
		}
		return d, nil
	}
	options, err := json.Marshal(compilerOptions(s.program.Options()))
	if err != nil {
		return Descriptor{}, err
	}
	d := Descriptor{
		Generation:        s.generation,
		CurrentDirectory:  s.program.CurrentDirectory(),
		CompilerOptions:   options,
		Sources:           make([]Source, 0, len(s.program.SourceFiles())),
		DiagnosticSources: s.diagnosticSources(),
	}
	for _, f := range s.program.SourceFiles() {
		d.Sources = append(d.Sources, s.source(f))
	}
	return d, nil
}

// Embed the native options so nested maps such as paths retain their key order.
type compilerOptionsJSON struct {
	*core.CompilerOptions
	JSX             string `json:"jsx,omitempty"`
	ModuleDetection string `json:"moduleDetection,omitempty"`
	NewLine         *int   `json:"newLine,omitempty"`
}

func compilerOptions(opts *core.CompilerOptions) compilerOptionsJSON {
	options := compilerOptionsJSON{
		CompilerOptions: opts,
		JSX: map[core.JsxEmit]string{
			core.JsxEmitPreserve:    "Preserve",
			core.JsxEmitReact:       "React",
			core.JsxEmitReactNative: "ReactNative",
			core.JsxEmitReactJSX:    "ReactJSX",
			core.JsxEmitReactJSXDev: "ReactJSXDev",
		}[opts.Jsx],
		ModuleDetection: map[core.ModuleDetectionKind]string{
			core.ModuleDetectionKindAuto:   "Auto",
			core.ModuleDetectionKindLegacy: "Legacy",
			core.ModuleDetectionKindForce:  "Force",
		}[opts.ModuleDetection],
	}
	if opts.NewLine != core.NewLineKindNone {
		value := int(opts.NewLine) - 1
		options.NewLine = &value
	}
	return options
}

var formatExtensions = []string{tspath.ExtensionMts, tspath.ExtensionMjs, tspath.ExtensionCts, tspath.ExtensionCjs}

func (s *Session) source(f *ast.SourceFile) Source {
	source := syntaxSource(f)
	source.Packages = s.program.PackageNamesForSourceFile(f)
	source.External = s.program.IsSourceFileFromExternalLibrary(f)
	source.DefaultLibrary = s.program.IsSourceFileDefaultLibrary(f)
	resolution := s.program.Options().GetModuleResolutionKind()
	if resolution == core.ModuleResolutionKindNode16 || resolution == core.ModuleResolutionKindNodeNext ||
		strings.Contains(source.FileName, "/node_modules/") || tspath.FileExtensionIsOneOf(source.FileName, formatExtensions) {
		source.ImpliedNodeFormat = s.program.SourceFileMetadata(f).ImpliedNodeFormat
	}
	return source
}

// Query returns one authoritative answer, with shallow facts for its direct handles.
func (s *Session) Query(ctx context.Context, q Query) ([]byte, error) {
	packets, err := s.QueryBatch(ctx, []Query{q})
	if err != nil {
		return nil, err
	}
	return packets[0], nil
}

// QueryBatch resolves syntax before acquiring a short exclusive checker lease.
// Concurrent calls to QueryBatch and Close are synchronized via lifeMu.
func (s *Session) QueryBatch(ctx context.Context, queries []Query) (packets [][]byte, err error) {
	defer func() {
		if value := recover(); value != nil {
			s.poisoned.Store(true)
			packets = nil
			err = fmt.Errorf("type service: %v", value)
		}
	}()
	if s.poisoned.Load() {
		return nil, ErrPoisoned
	}
	if len(queries) > 256 {
		return nil, errors.New("type service batch exceeds 256 queries")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.lifeMu.RLock()
	defer s.lifeMu.RUnlock()
	if s.poisoned.Load() {
		return nil, ErrPoisoned
	}
	if s.program == nil {
		return nil, ErrClosed
	}
	p := s.program
	eligible := s.eligible
	for _, q := range queries {
		if q.Generation != s.generation {
			return nil, errors.New("stale or foreign type service handle")
		}
		if q.Request == nil {
			return nil, errors.New("missing type service request")
		}
		if s.syntaxOnly && q.Request.operation() != "nodeRecord" {
			return nil, errors.New("semantic queries require a typed project")
		}
	}
	natives := make([]*ast.Node, len(queries))
	for i, q := range queries {
		if ref := requestNode(q.Request); ref != nil {
			f := s.sourceFile(ref.File)
			if s.syntaxOnly && !eligible[f] {
				return nil, errors.New("foreign syntax service source")
			}
			natives[i], err = resolveNode(f, *ref)
			if err != nil {
				return nil, err
			}
		}
	}
	results := make([]Result, len(queries))
	records := make([][]Record, len(queries))
	// Program diagnostics may acquire multiple checker shards; run them outside
	// the single-file exclusive checker lease while preserving request order.
	for start := 0; start < len(queries); {
		if r, ok := queries[start].Request.(*ProgramRequest); ok {
			results[start], err = programQuery(ctx, s, p, r, natives[start])
			if err != nil {
				return nil, err
			}
			start++
			continue
		}
		if queries[start].Request.operation() == "nodeRecord" {
			ref := requestNode(queries[start].Request)
			results[start] = nodeFacts(s.sourceFile(ref.File), natives[start])
			start++
			continue
		}
		end := start + 1
		for end < len(queries) {
			if _, ok := queries[end].Request.(*ProgramRequest); ok || queries[end].Request.operation() == "nodeRecord" {
				break
			}
			end++
		}
		if err = s.checkBatch(ctx, queries[start:end], natives[start:end], results[start:end], records[start:end]); err != nil {
			return nil, err
		}
		start = end
	}
	packets = make([][]byte, len(queries))
	for i, result := range results {
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
	case *EncodedTypeNode:
		fact.origins = nodeRefs(p, fact.parseNodes)
		fact.parseNodes = nil
	case *TypeDetails:
		fact.rootRef = nodeReference(p, fact.rootNode)
		fact.labelRefs = nodeRefs(p, fact.labels)
		fact.rootNode = nil
		fact.labels = nil
	case *SymbolFact:
		if fact.nodes != nil {
			fact.Declarations = make([]NodeRef, 0, len(fact.nodes))
			for _, node := range fact.nodes {
				ref := nodeReference(p, node)
				if ref == nil {
					continue
				}
				fact.Declarations = append(fact.Declarations, *ref)
				if node == fact.valueNode {
					fact.ValueDeclaration = ref
				}
			}
		}
		if fact.ValueDeclaration == nil {
			fact.ValueDeclaration = nodeReference(p, fact.valueNode)
		}
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
	case OptionalIndexInfos:
		if fact.Present {
			materializeReferences(p, fact.Infos)
		}
	}
}

func (s *Session) checkBatch(ctx context.Context, queries []Query, natives []*ast.Node, results []Result, records [][]Record) error {
	s.checkMu.Lock()
	defer s.checkMu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	c, done := s.program.TypeCheckerForFileExclusive(core.WithCheckerLifetime(ctx, core.CheckerLifetimeAPI), s.anchor)
	defer done()
	if err := ctx.Err(); err != nil {
		return err
	}
	if c == nil {
		return errors.New("program cannot provide type services")
	}
	if s.checker == nil {
		s.checker = c
		s.types = make(map[uint32]*checker.Type)
		s.symbols = make(map[uint32]*ast.Symbol)
		s.signatures = make(map[uint32]*checker.Signature)
	} else if c != s.checker {
		return errors.New("type service checker identity changed")
	}
	for i, q := range queries {
		if err := ctx.Err(); err != nil {
			return err
		}
		var err error
		results[i], err = s.execute(ctx, c, q.Request, natives[i])
		if err != nil {
			return err
		}
		if !q.SkipRecords {
			records[i], err = s.directRecords(ctx, c, q.Request.operation(), results[i])
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func directRecordKind(op Operation) Operation {
	switch op {
	case "type", "fromTypeNode", "context", "constraint", "apparent", "baseLiteral", "nonNullable",
		"errorType", "anyType", "unknownType", "stringType", "numberType", "booleanType",
		"bigIntType", "eSSymbolType", "nonPrimitiveType", "neverType", "nullType", "undefinedType",
		"trueType", "falseType", "stringLiteralType", "numberLiteralType", "bigIntLiteralType",
		"nullableType", "void",
		"declaredType", "symbolType", "returnType", "numberIndexType", "typeArguments",
		"stringIndexType", "awaited", "widened", "baseTypes", "defaultType", "typeOfSymbol",
		"indexInfo", "indexInfos", "declaredIndexInfos", "typePredicate", "parameterType", "propertyType",
		"mappedFields", "classThisType", "contextualArgument":
		return "typeRecord"
	case "symbol", "declarationSymbol", "aliasedSymbol", "property", "properties", "resolveName",
		"exportSymbol", "immediateAliasedSymbol", "shorthandValueSymbol",
		"exportSpecifierLocalTargetSymbol", "propertySymbolOfDestructuringAssignment",
		"mergedSymbol", "moduleMember", "parameterPropertySymbols":
		return "symbolRecord"
	case "signature", "resolvedSignature", "signatures":
		return "signatureRecord"
	default:
		return ""
	}
}

func (s *Session) directRecords(ctx context.Context, c *checker.Checker, op Operation, result Result) ([]Record, error) {
	kind := directRecordKind(op)
	if kind == "" {
		return nil, nil
	}
	var ids []uint32
	switch value := result.(type) {
	case Handle:
		ids = []uint32{uint32(value)}
	case Handles:
		ids = value
	case NamedSymbols:
		ids = make([]uint32, 0, len(value))
		for _, item := range value {
			ids = append(ids, item.ID)
		}
	case IndexInfos:
		ids = make([]uint32, 0, 2*len(value))
		for _, info := range value {
			ids = append(ids, info.KeyType, info.Type)
		}
	case OptionalIndexInfos:
		if value.Present {
			ids = make([]uint32, 0, 2*len(value.Infos))
			for _, info := range value.Infos {
				ids = append(ids, info.KeyType, info.Type)
			}
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
		fact, err := s.execute(ctx, c, request, nil)
		if err != nil {
			return err
		}
		records = append(records, Record{kind, id, fact})
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
func nodeFacts(file *ast.SourceFile, node *ast.Node) NodeFacts {
	facts := NodeFacts{}
	seen := map[*ast.Node]bool{}
	table := encoder.GetNodeIndexTable(file)
	reference := func(n *ast.Node) *NodeRef { return indexedNodeReference(file, table, n) }
	add := func(n *ast.Node) *ast.Node {
		if n == nil {
			return nil
		}
		ref := reference(n)
		if int(ref.Index) < len(table.Nodes) {
			n = table.Nodes[ref.Index]
		}
		if seen[n] {
			return n
		}
		seen[n] = true
		fact := NodeFact{
			Node:   *ref,
			Kind:   strings.TrimPrefix(n.Kind.String(), "Kind"),
			Flags:  uint32(n.Flags),
			Parent: reference(syntaxParent(n)),
		}
		if n.Flags&ast.NodeFlagsPossiblyContainsDeprecatedTag != 0 && ast.GetJSDocDeprecatedTag(n) == nil {
			fact.Flags &^= uint32(ast.NodeFlagsPossiblyContainsDeprecatedTag)
		}
		if ast.IsModuleDeclaration(n) {
			fact.ModuleKeyword = uint32(n.AsModuleDeclaration().Keyword)
			if n.Parent != nil && ast.IsModuleDeclaration(n.Parent) {
				fact.Flags |= uint32(ast.NodeFlagsNestedNamespace)
			}
		}
		if ast.IsIdentifier(n) || ast.IsPrivateIdentifier(n) || ast.IsStringLiteral(n) || ast.IsNumericLiteral(n) {
			text := n.Text()
			fact.Text = &text
		}
		fact.Name = reference(syntaxName(n))
		facts = append(facts, fact)
		return n
	}
	for current := node; current != nil; current = syntaxParent(current) {
		current = add(current)
		add(syntaxName(current))
	}
	return facts
}

func syntaxParent(node *ast.Node) *ast.Node {
	parent := node.Parent
	for parent != nil && (parent.Kind == ast.KindAsExpression || parent.Kind == ast.KindSatisfiesExpression) && utils.JSDocTypeCastExpression(parent) != nil {
		parent = parent.Parent
	}
	return parent
}

func syntaxName(node *ast.Node) *ast.Node {
	if node.Kind == ast.KindJSDocSeeTag {
		return node.AsJSDocSeeTag().NameExpression
	}
	name := node.Name()
	if node.Kind == ast.KindJSDocTypedefTag || node.Kind == ast.KindJSDocCallbackTag {
		for name != nil && ast.IsModuleDeclaration(name) {
			if name.AsModuleDeclaration().Body == nil {
				return name.Name()
			}
			name = name.AsModuleDeclaration().Body
		}
	}
	return name
}

func nodeReference(p *program.Program, n *ast.Node) *NodeRef {
	if n == nil {
		return nil
	}
	f := ast.GetSourceFileOfNode(n)
	if f == nil || !p.OwnsSourceFile(f) {
		panic("declaration is outside Program")
	}
	return indexedNodeReference(f, encoder.GetNodeIndexTable(f), n)
}

func indexedNodeReference(f *ast.SourceFile, table *encoder.NodeIndexTable, n *ast.Node) *NodeRef {
	if n == nil {
		return nil
	}
	index := uint32(0)
	if n.Flags&ast.NodeFlagsJavaScriptFile != 0 && n.Flags&(ast.NodeFlagsJSDoc|ast.NodeFlagsReparsed) != 0 {
		index = jsDocNodeIndex(n, f, table)
	}
	if index == 0 {
		index = table.GetIndex(n)
	}
	if index == 0 {
		panic("declaration has no compiler node handle")
	}
	return &NodeRef{File: f.FileName(), Index: index}
}

func (s *Session) typeByID(id uint32) (*checker.Type, error) {
	t := s.types[id]
	if t == nil {
		return nil, errors.New("invalid type handle")
	}
	return t, nil
}

func (s *Session) symbolByID(id uint32) (*ast.Symbol, error) {
	v := s.symbols[id]
	if v == nil {
		return nil, errors.New("invalid symbol handle")
	}
	return v, nil
}

func (s *Session) signatureByID(id uint32) (*checker.Signature, error) {
	v := s.signatures[id]
	if v == nil {
		return nil, errors.New("invalid signature handle")
	}
	return v, nil
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
	if s.indexInfoIDs == nil {
		s.indexInfoIDs = make(map[*checker.IndexInfo]uint32)
	}
	id := s.indexInfoIDs[info]
	if id == 0 {
		id = uint32(len(s.indexInfoIDs)) + 1
		s.indexInfoIDs[info] = id
	}
	return IndexInfo{
		ID:         id,
		KeyType:    s.typeID(info.KeyType()),
		Type:       s.typeID(info.ValueType()),
		IsReadonly: info.IsReadonly(),
		node:       info.Declaration(),
	}
}

func (s *Session) execute(ctx context.Context, c *checker.Checker, request Request, native *ast.Node) (Result, error) {
	switch r := request.(type) {
	case *NodeRequest:
		switch r.Op {
		case "exportSpecifierLocalTargetSymbol":
			if (native.Kind != ast.KindExportSpecifier || native.Parent == nil || native.Parent.Parent == nil) &&
				native.Kind != ast.KindIdentifier {
				return nil, errors.New("expected export specifier or identifier")
			}
			return Handle(s.symbolID(c.GetExportSpecifierLocalTargetSymbol(native))), nil
		case "propertySymbolOfDestructuringAssignment":
			if native.Kind != ast.KindIdentifier || native.Parent == nil || native.Parent.Parent == nil {
				return Handle(0), nil
			}
			return Handle(s.symbolID(c.GetPropertySymbolOfDestructuringAssignment(native))), nil
		case "jsxIntrinsicTagNamesAt":
			return s.namedSymbols(c.GetJsxIntrinsicTagNamesAt(native)), nil
		case "constantValue":
			return constantValue(c.GetConstantValue(native)), nil
		case "isOptionalParameter":
			return Boolean(checker.Checker_isOptionalParameter(c, native)), nil
		case "isImplementationOfOverload":
			return overloadImplementation(c, native), nil
		case "type":
			switch native.Kind {
			case ast.KindJSDocPropertyTag, ast.KindJSDocTypedefTag, ast.KindJSDocCallbackTag:
				if bound := boundJSDocDeclaration(native); bound != nil {
					native = bound
				}
			}
			return Handle(s.typeID(c.GetTypeAtLocation(native))), nil
		case "fromTypeNode":
			return Handle(s.typeID(c.GetTypeFromTypeNode(native))), nil
		case "context":
			if !ast.IsExpression(native) {
				return Handle(0), nil
			}
			return Handle(s.typeID(c.GetContextualType(native, checker.ContextFlagsNone))), nil
		case "declarationSymbol":
			return Handle(s.symbolID(native.Symbol())), nil
		case "symbol":
			if native.Kind == ast.KindSourceFile && !ast.IsExternalModule(native.AsSourceFile()) {
				return Handle(0), nil
			}
			return Handle(s.symbolID(c.GetSymbolAtLocation(native))), nil
		case "signature":
			if native.Kind == ast.KindJSDocSignature {
				if bound := boundJSDocDeclaration(native); bound != nil {
					native = bound
				}
			}
			if !ast.IsFunctionLike(native) {
				return Handle(0), nil
			}
			return Handle(s.signatureID(c.GetSignatureFromDeclaration(native))), nil
		case "shorthandValueSymbol":
			return Handle(s.symbolID(c.GetShorthandAssignmentValueSymbol(native))), nil
		case "suggestionDiagnostics":
			if native.Kind != ast.KindSourceFile {
				return nil, errors.New("expected a source file for suggestion diagnostics")
			}
			return diagnosticFacts(c.GetSuggestionDiagnostics(ctx, native.AsSourceFile())), ctx.Err()
		}
	case *ResolvedSignatureRequest:
		if !ast.IsCallLikeExpression(native) {
			return Handles{0}, nil
		}
		var candidates []*checker.Signature
		var candidatesOut *[]*checker.Signature
		if r.WithCandidates {
			candidatesOut = &candidates
		}
		sig := checker.Checker_getResolvedSignatureWithArgumentCount(
			c, native, candidatesOut, r.ArgumentCount, checker.CheckModeNormal,
		)
		ids := make(Handles, 1, 1+len(candidates))
		ids[0] = s.signatureID(sig)
		for _, candidate := range candidates {
			ids = append(ids, s.signatureID(candidate))
		}
		return ids, nil
	case *NamedNodeRequest:
		switch r.Op {
		case "validPropertyAccess":
			switch native.Kind {
			case ast.KindPropertyAccessExpression, ast.KindQualifiedName, ast.KindImportType:
				return Boolean(c.IsValidPropertyAccess(native, string(r.Name))), nil
			default:
				return nil, errors.New("expected property access, qualified name, or import type")
			}
		case "parameterPropertySymbols":
			if !ast.IsParameterPropertyDeclaration(native, native.Parent) ||
				native.Parent == nil || native.Parent.Parent == nil ||
				native.Parent.Locals() == nil || native.Parent.Parent.Symbol() == nil {
				return nil, errors.New("parameter is not a parameter property declaration")
			}
			param, prop := c.GetSymbolsOfParameterPropertyDeclaration(native, string(r.Name))
			ids := Handles{}
			if param != nil {
				ids = append(ids, s.symbolID(param))
			}
			if prop != nil {
				ids = append(ids, s.symbolID(prop))
			}
			return ids, nil
		}
	case *ContextualArgumentRequest:
		if !ast.IsCallLikeExpression(native) {
			return Handle(0), nil
		}
		return Handle(s.typeID(c.GetContextualTypeForArgumentAtIndex(native, r.Index))), nil
	case *TypeRequest:
		t, err := s.typeByID(r.ID)
		if err != nil {
			return nil, err
		}
		switch r.Op {
		case "augmentedPropertiesOfType":
			return s.namedSymbols(checker.Checker_getAugmentedPropertiesOfType(c, t)), nil
		case "constraint":
			return Handle(s.typeID(c.GetBaseConstraintOfType(t))), nil
		case "apparent":
			return Handle(s.typeID(c.GetApparentType(t))), nil
		case "baseLiteral":
			return Handle(s.typeID(c.GetBaseTypeOfLiteralType(t))), nil
		case "nonNullable":
			return Handle(s.typeID(c.GetNonNullableType(t))), nil
		case "awaited":
			return Handle(s.typeID(checker.Checker_getAwaitedType(c, t))), nil
		case "widened":
			return Handle(s.typeID(c.GetWidenedType(t))), nil
		case "baseTypes":
			if t.ObjectFlags()&checker.ObjectFlagsReference != 0 {
				t = t.Target()
			}
			return s.typeIDs(c.GetBaseTypes(t)), nil
		case "defaultType":
			return Handle(s.typeID(c.GetDefaultFromTypeParameter(t))), nil
		case "declaredIndexInfos":
			return s.declaredIndexInfos(t), nil
		case "typeDetails":
			return s.typeDetails(t), nil
		case "classThisType":
			iface := t.AsInterfaceType()
			if iface == nil {
				return Handle(0), nil
			}
			return Handle(s.typeID(checker.InterfaceType_thisType(iface))), nil
		case "mappedFields":
			if t.ObjectFlags()&checker.ObjectFlagsMapped == 0 {
				return Handles(nil), nil
			}
			mapped := t.AsMappedType()
			return Handles{
				s.typeID(checker.MappedType_typeParameter(mapped)),
				s.typeID(checker.MappedType_constraintType(mapped)),
				s.typeID(checker.MappedType_templateType(mapped)),
				s.typeID(checker.MappedType_modifiersType(mapped)),
			}, nil
		case "properties":
			return s.namedSymbols(c.GetPropertiesOfType(t)), nil
		case "typeArguments":
			return s.typeIDs(c.GetTypeArguments(t)), nil
		case "isArray":
			return Boolean(c.IsArrayType(t)), nil
		case "isArrayLike":
			return Boolean(c.IsArrayLikeType(t)), nil
		case "isTuple":
			return Boolean(checker.IsTupleType(t)), nil
		case "numberIndexType":
			return Handle(s.typeID(c.GetNumberIndexType(t))), nil
		case "stringIndexType":
			return Handle(s.typeID(c.GetStringIndexType(t))), nil
		case "indexInfos":
			infos := c.GetIndexInfosOfType(t)
			result := make(IndexInfos, len(infos))
			for i, info := range infos {
				result[i] = s.indexInfo(info)
			}
			return result, nil
		case "typeRecord":
			return s.typeFact(t)
		}
	case *SymbolRequest:
		sym, err := s.symbolByID(r.ID)
		if err != nil {
			return nil, err
		}
		switch r.Op {
		case "rootSymbols":
			return s.namedSymbols(c.GetRootSymbols(sym)), nil
		case "mergedSymbol":
			return Handle(s.symbolID(c.GetMergedSymbol(sym))), nil
		case "isArgumentsSymbol":
			return Boolean(c.GetArgumentsSymbol() == sym), nil
		case "isUndefinedSymbol":
			return Boolean(c.GetUndefinedSymbol() == sym), nil
		case "aliasedSymbol":
			return Handle(s.symbolID(c.GetAliasedSymbol(sym))), nil
		case "immediateAliasedSymbol":
			return Handle(s.symbolID(c.GetImmediateAliasedSymbol(sym))), nil
		case "isUnknownSymbol":
			return Boolean(c.IsUnknownSymbol(sym)), nil
		case "exportsOfModule":
			exports := c.GetExportsOfModule(sym)
			slices.SortFunc(exports, c.CompareSymbols)
			return s.namedSymbols(exports), nil
		case "fullyQualifiedName":
			return Text(c.GetFullyQualifiedName(sym)), nil
		case "symbolMembers", "symbolExports":
			return s.symbolTable(c, sym, r.Op == "symbolExports"), nil
		case "exportSymbol":
			return Handle(s.symbolID(c.GetExportSymbolOfSymbol(sym))), nil
		case "declaredType":
			return Handle(s.typeID(c.GetDeclaredTypeOfSymbol(sym))), nil
		case "typeOfSymbol":
			return Handle(s.typeID(c.GetTypeOfSymbol(sym))), nil
		case "symbolRecord":
			declarations, value := symbolDeclarations(sym, true)
			return &SymbolFact{
				Name:        symbolName(sym),
				EscapedName: ast.EscapeSymbolName(sym.Name),
				Flags:       uint32(sym.Flags),
				CheckFlags:  jsCheckFlags(sym.CheckFlags),
				nodes:       declarations,
				valueNode:   value,
			}, nil
		}
	case *SymbolDocsRequest:
		if !r.WithChecker {
			c = nil
		}
		symbol, err := s.symbolByID(r.ID)
		if err != nil {
			return nil, err
		}
		active := map[*ast.Symbol]bool{symbol: true}
		declarations, _ := symbolDeclarations(symbol, false)
		if r.Op == "symbolDocumentation" {
			return documentation(c, declarations, active), nil
		}
		return jsDocTags(c, declarations, active), nil
	case *SignatureRequest:
		sig, err := s.signatureByID(r.ID)
		if err != nil {
			return nil, err
		}
		switch r.Op {
		case "returnType":
			return Handle(s.typeID(c.GetReturnTypeOfSignature(sig))), nil
		case "signatureDocumentation", "signatureJSDocTags":
			declarations := []*ast.Node{sig.Declaration()}
			active := make(map[*ast.Symbol]bool)
			if r.Op == "signatureDocumentation" {
				return documentation(c, declarations, active), nil
			}
			return jsDocTags(c, declarations, active), nil
		case "typePredicate":
			predicate := c.GetTypePredicateOfSignature(sig)
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
			return &SignatureFact{
				Parameters:       s.symbolIDs(sig.Parameters()),
				TypeParameters:   s.typeIDs(sig.TypeParameters()),
				ThisParameter:    s.symbolID(sig.ThisParameter()),
				MinArgumentCount: sig.MinArgumentCount(),
				node:             sig.Declaration(),
			}, nil
		}
	case *CheckerRequest:
		switch r.Op {
		case "errorType":
			return Handle(s.typeID(c.GetErrorType())), nil
		case "anyType":
			return Handle(s.typeID(c.GetAnyType())), nil
		case "unknownType":
			return Handle(s.typeID(c.GetUnknownType())), nil
		case "stringType":
			return Handle(s.typeID(c.GetStringType())), nil
		case "numberType":
			return Handle(s.typeID(c.GetNumberType())), nil
		case "booleanType":
			return Handle(s.typeID(c.GetBooleanType())), nil
		case "bigIntType":
			return Handle(s.typeID(c.GetBigIntType())), nil
		case "eSSymbolType":
			return Handle(s.typeID(c.GetESSymbolType())), nil
		case "nonPrimitiveType":
			return Handle(s.typeID(c.GetNonPrimitiveType())), nil
		case "neverType":
			return Handle(s.typeID(c.GetNeverType())), nil
		case "nullType":
			return Handle(s.typeID(c.GetNullType())), nil
		case "undefinedType":
			return Handle(s.typeID(c.GetUndefinedType())), nil
		case "trueType":
			return Handle(s.typeID(c.GetBooleanType().Types()[1])), nil
		case "falseType":
			return Handle(s.typeID(c.GetBooleanType().Types()[0])), nil
		case "void":
			return Handle(s.typeID(c.GetVoidType())), nil
		case "ambientModules":
			modules := slices.Clone(c.GetAmbientModules())
			slices.SortFunc(modules, c.CompareSymbols)
			return s.namedSymbols(modules), nil
		}
	case *PropertyRequest:
		switch r.Op {
		case "moduleMember":
			sym, err := s.symbolByID(r.ID)
			if err != nil {
				return nil, err
			}
			return Handle(s.symbolID(c.TryGetMemberInModuleExports(string(r.Name), sym))), nil
		case "property":
			t, err := s.typeByID(r.ID)
			if err != nil {
				return nil, err
			}
			return Handle(s.symbolID(c.GetPropertyOfType(t, string(r.Name)))), nil
		case "propertyType":
			t, err := s.typeByID(r.ID)
			if err != nil {
				return nil, err
			}
			return Handle(s.typeID(c.GetTypeOfPropertyOfType(t, string(r.Name)))), nil
		}
	case *SymbolTypeRequest:
		sym, err := s.symbolByID(r.ID)
		if err != nil {
			return nil, err
		}
		return Handle(s.typeID(c.GetTypeOfSymbolAtLocation(sym, native))), nil
	case *ResolveNameRequest:
		return Handle(s.symbolID(resolveName(c, string(r.Name), native, ast.SymbolFlags(r.Kind), r.ExcludeGlobals))), nil
	case *AssignableRequest:
		source, err := s.typeByID(r.Source)
		if err != nil {
			return nil, err
		}
		target, err := s.typeByID(r.Target)
		if err != nil {
			return nil, err
		}
		return Boolean(c.IsTypeAssignableTo(source, target)), nil
	case *IndexInfoRequest:
		if r.Kind > 1 {
			return nil, errors.New("invalid index kind")
		}
		t, err := s.typeByID(r.ID)
		if err != nil {
			return nil, err
		}
		key := c.GetNumberType()
		if r.Kind == 0 {
			key = c.GetStringType()
		}
		info := c.GetIndexInfoOfType(t, key)
		if info == nil {
			return IndexInfos{}, nil
		}
		return IndexInfos{s.indexInfo(info)}, nil
	case *ScopeSymbolsRequest:
		return s.namedSymbols(symbolsInScope(c, native, ast.SymbolFlags(r.Meaning))), nil
	case *ParameterTypeRequest:
		sig, err := s.signatureByID(r.ID)
		if err != nil {
			return nil, err
		}
		return Handle(s.typeID(c.GetTypeParameterAtPosition(sig, r.Position))), nil
	case *StringLiteralRequest:
		return Handle(s.typeID(checker.Checker_getStringLiteralType(c, string(r.Value)))), nil
	case *NumberLiteralRequest:
		value := r.Value
		switch r.NonFinite {
		case "":
		case "NaN":
			value = math.NaN()
		case "Infinity":
			value = math.Inf(1)
		case "-Infinity":
			value = math.Inf(-1)
		default:
			return nil, errors.New("invalid non-finite number literal")
		}
		return Handle(s.typeID(checker.Checker_getNumberLiteralType(c, jsnum.Number(value)))), nil
	case *BigIntLiteralRequest:
		text := r.Base10Value
		if text == "" {
			text = "0"
		}
		negative := r.Negative && text != "0"
		return Handle(s.typeID(checker.Checker_getBigIntLiteralType(c, jsnum.PseudoBigInt{Negative: negative, Base10Value: text}))), nil
	case *NullableTypeRequest:
		t, err := s.typeByID(r.ID)
		if err != nil {
			return nil, err
		}
		return Handle(s.typeID(checker.Checker_getNullableType(c, t, checker.TypeFlags(r.Flags)))), nil
	case *TypePredicateStringRequest:
		if r.Kind > uint32(checker.TypePredicateKindAssertsIdentifier) {
			return nil, errors.New("invalid type predicate kind")
		}
		var name string
		if r.ParameterName != nil {
			name = string(*r.ParameterName)
		}
		var predType *checker.Type
		if r.Type != 0 {
			var err error
			predType, err = s.typeByID(r.Type)
			if err != nil {
				return nil, err
			}
		}
		pred := checker.Checker_newTypePredicate(c, checker.TypePredicateKind(r.Kind), name, r.ParameterIndex, predType)
		flags := checker.TypeFormatFlagsUseAliasDefinedOutsideCurrentScope
		if r.Flags != nil {
			flags = checker.TypeFormatFlags(*r.Flags)
		}
		return Text(checker.Checker_typePredicateToStringEx(c, pred, native, flags)), nil
	case *PrintRequest:
		return s.executePrint(c, r, native)
	case *SignaturesRequest:
		if r.Kind > 1 {
			return nil, errors.New("invalid signature kind")
		}
		t, err := s.typeByID(r.ID)
		if err != nil {
			return nil, err
		}
		ids := Handles{}
		for _, v := range c.GetSignaturesOfType(t, checker.SignatureKind(r.Kind)) {
			ids = append(ids, s.signatureID(v))
		}
		return ids, nil
	}
	return nil, fmt.Errorf("unsupported type service operation %q", request.operation())
}

func (s *Session) symbolTable(c *checker.Checker, sym *ast.Symbol, exports bool) OptionalNamedSymbols {
	table := sym.Members
	expectedFlags := ast.SymbolFlagsClass | ast.SymbolFlagsInterface | ast.SymbolFlagsTypeLiteral | ast.SymbolFlagsObjectLiteral
	if exports {
		table = sym.Exports
		expectedFlags = ast.SymbolFlagsClass | ast.SymbolFlagsEnum | ast.SymbolFlagsModule | ast.SymbolFlagsVariable
	}
	present := table != nil
	if !present && len(sym.Declarations) != 0 {
		// TypeScript's binder creates empty member/export maps for container symbols.
		bound := sym.Declarations[0].Symbol()
		if (bound == sym || c.GetMergedSymbol(bound) == sym) && sym.Flags&expectedFlags != 0 {
			present = true
		}
	}
	if !present {
		return OptionalNamedSymbols{}
	}
	members := slices.Collect(maps.Values(table))
	slices.SortFunc(members, c.CompareSymbols)
	result := make(NamedSymbols, len(members))
	for i, member := range members {
		result[i] = NamedSymbol{ID: s.symbolID(member), Name: ast.EscapeSymbolName(member.Name)}
	}
	return OptionalNamedSymbols{Present: true, Symbols: result}
}

func (s *Session) typeFact(t *checker.Type) (*TypeFact, error) {
	flags := t.Flags()
	fact := &TypeFact{
		Flags:              checker.FormatTypeFlags(flags),
		Symbol:             s.symbolID(t.Symbol()),
		AliasSymbol:        s.symbolID(t.Alias().Symbol()),
		AliasTypeArguments: s.typeIDs(t.Alias().TypeArguments()),
	}
	if flags&checker.TypeFlagsIntrinsic != 0 {
		fact.IntrinsicName = t.AsIntrinsicType().IntrinsicName()
	} else if flags&checker.TypeFlagsBoolean != 0 {
		fact.IntrinsicName = "boolean"
	}
	if flags&checker.TypeFlagsUniqueESSymbol != 0 {
		fact.EscapedName = ast.EscapeSymbolName(checker.GetPropertyNameFromType(t))
	}
	if flags&(checker.TypeFlagsObject|checker.TypeFlagsTypeParameter) != 0 {
		fact.Target = s.typeID(t.Target())
	}
	if flags&(checker.TypeFlagsObject|checker.TypeFlagsUnionOrIntersection|checker.TypeFlagsIntrinsic|checker.TypeFlagsBooleanLiteral) != 0 {
		objectFlags := t.ObjectFlags()
		fact.ObjectFlags = []string{}
		for _, flag := range publicObjectFlags {
			if objectFlags&flag.flag != 0 {
				fact.ObjectFlags = append(fact.ObjectFlags, flag.name)
			}
		}
	}
	if flags&checker.TypeFlagsObject != 0 {
		if definition := t.AsInterfaceType(); definition != nil {
			if parameters := definition.TypeParameters(); len(parameters) > 0 {
				fact.TypeParameters = s.typeIDs(parameters)
			}
			if t.ObjectFlags()&checker.ObjectFlagsTuple != 0 {
				tuple := t.AsTupleType()
				fact.Tuple = &TupleFact{
					Readonly:      tuple.IsReadonly(),
					CombinedFlags: uint32(checker.TupleType_combinedFlags(tuple)),
				}
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

func (s *Session) executePrint(c *checker.Checker, r *PrintRequest, native *ast.Node) (Result, error) {
	switch r.Op {
	case "symbolString":
		sym, err := s.symbolByID(r.ID)
		if err != nil {
			return nil, err
		}
		flags := checker.SymbolFormatFlagsAllowAnyNodeKind
		if r.Flags != nil {
			flags = checker.SymbolFormatFlags(*r.Flags)
		}
		return Text(c.SymbolToStringEx(sym, native, ast.SymbolFlags(r.Kind), flags)), nil
	case "typeNode":
		t, err := s.typeByID(r.ID)
		if err != nil {
			return nil, err
		}
		var builderFlags nodebuilder.Flags
		if r.Flags != nil {
			builderFlags = nodebuilder.Flags(*r.Flags)
		}
		builder := checker.NewNodeBuilder(c, printer.NewEmitContext())
		node := builder.TypeToTypeNode(t, native, builderFlags, nodebuilder.InternalFlagsNone, nil)
		return encodeGeneratedNode(node, builder, native)
	case "typeString":
		t, err := s.typeByID(r.ID)
		if err != nil {
			return nil, err
		}
		flags := checker.TypeFormatFlagsAllowUniqueESSymbolType | checker.TypeFormatFlagsUseAliasDefinedOutsideCurrentScope
		if r.Flags != nil {
			flags = checker.TypeFormatFlags(*r.Flags)
		}
		return Text(sanitizeTruncatedUTF8(c.TypeToStringEx(t, native, flags, nil))), nil
	case "signatureString":
		if r.Kind > uint32(checker.SignatureKindConstruct) {
			return nil, errors.New("invalid signature kind")
		}
		sig, err := s.signatureByID(r.ID)
		if err != nil {
			return nil, err
		}
		flags := checker.TypeFormatFlagsNone
		if r.Flags != nil {
			flags = checker.TypeFormatFlags(*r.Flags)
		}
		if r.Kind == uint32(checker.SignatureKindCall) {
			flags |= checker.TypeFormatFlagsWriteCallStyleSignature
			return Text(c.SignatureToStringEx(sig, native, flags, nil)), nil
		}
		// TypeScript's signatureToString lets callers format any signature with
		// SignatureKind.Construct; tsgo's SignatureToStringEx only inspects sig.flags.
		sigOutput := ast.KindConstructSignature
		if flags&checker.TypeFormatFlagsWriteArrowStyleSignature != 0 {
			sigOutput = ast.KindConstructorType
		}
		emitCtx := printer.NewEmitContext()
		builder := checker.NewNodeBuilder(c, emitCtx)
		combinedFlags := nodebuilder.Flags(flags&checker.TypeFormatFlagsNodeBuilderFlagsMask) |
			nodebuilder.FlagsIgnoreErrors |
			nodebuilder.FlagsWriteTypeParametersInQualifiedName
		node := builder.SignatureToSignatureDeclaration(sig, sigOutput, native, combinedFlags, nodebuilder.InternalFlagsNone, nil)
		p := printer.NewPrinter(printer.PrinterOptions{
			RemoveComments:        true,
			OmitTrailingSemicolon: true,
			NeverAsciiEscape:      true,
		}, printer.PrintHandlers{}, emitCtx)
		var sf *ast.SourceFile
		if native != nil {
			sf = ast.GetSourceFileOfNode(native)
		}
		if flags&checker.TypeFormatFlagsMultilineObjectLiterals != 0 {
			w := printer.NewTextWriter("\n", 0)
			p.Write(node, sf, w, nil)
			return Text(w.String()), nil
		}
		w, put := printer.GetSingleLineStringWriter()
		defer put()
		p.Write(node, sf, w, nil)
		return Text(w.String()), nil
	default:
		return nil, fmt.Errorf("unsupported print operation %q", r.Op)
	}
}

func sanitizeTruncatedUTF8(s string) string {
	const ellipsis = "..."
	if !strings.HasSuffix(s, ellipsis) || utf8.ValidString(s) {
		return s
	}
	prefix := s[:len(s)-len(ellipsis)]
	for len(prefix) > 0 && !utf8.ValidString(prefix) {
		prefix = prefix[:len(prefix)-1]
	}
	return prefix + ellipsis
}

func resolveNode(f *ast.SourceFile, node NodeRef) (*ast.Node, error) {
	if f == nil {
		return nil, errors.New("foreign source handle")
	}
	table := encoder.GetNodeIndexTable(f)
	if node.Index >= uint32(len(table.Nodes)) || table.Nodes[node.Index] == nil {
		return nil, errors.New("invalid compiler node handle")
	}
	parent := table.Nodes[node.Index]
	if node.Token == nil {
		return parent, nil
	}
	token := node.Token
	positions := f.GetPositionMap()
	pos, end := positions.UTF16ToUTF8(token.Pos), positions.UTF16ToUTF8(token.End)
	if pos < parent.Pos() || end < pos || end > parent.End() {
		return nil, errors.New("invalid compiler token range")
	}
	if token.Kind == ast.KindSyntaxList {
		list := ast.NewNodeFactory(ast.NodeFactoryHooks{}).NewSyntaxList(nil)
		list.Loc, list.Parent = core.NewTextRange(pos, end), parent
		return list, nil
	}
	if !ast.IsTokenKind(token.Kind) {
		return nil, errors.New("invalid compiler token kind")
	}
	tokenStart, flags, ok := scanValidatedToken(f, token.Kind, pos, end)
	if !ok {
		return nil, errors.New("compiler token does not match source")
	}
	return createEphemeralToken(f, token.Kind, pos, tokenStart, end, parent, flags), nil
}

// scanValidatedToken verifies that a client-supplied token kind and range match
// the actual source text without writing into the shared SourceFile tokenCache.
func scanValidatedToken(f *ast.SourceFile, kind ast.Kind, pos, end int) (int, ast.TokenFlags, bool) {
	s := scanner.GetScannerForSourceFile(f, pos)
	scanned := s.Token()
	if (scanned == kind || (kind == ast.KindIdentifier && ast.IsKeywordKind(scanned))) && s.TokenEnd() == end {
		return s.TokenStart(), s.TokenFlags(), true
	}
	switch kind {
	case ast.KindGreaterThanToken, ast.KindGreaterThanGreaterThanToken, ast.KindGreaterThanEqualsToken,
		ast.KindGreaterThanGreaterThanEqualsToken, ast.KindGreaterThanGreaterThanGreaterThanToken,
		ast.KindGreaterThanGreaterThanGreaterThanEqualsToken:
		if s.ReScanGreaterThanToken() == kind && s.TokenEnd() == end {
			return s.TokenStart(), s.TokenFlags(), true
		}
	case ast.KindLessThanToken:
		if s.ReScanLessThanToken() == kind && s.TokenEnd() == end {
			return s.TokenStart(), s.TokenFlags(), true
		}
	case ast.KindQuestionToken:
		if s.ReScanQuestionToken() == kind && s.TokenEnd() == end {
			return s.TokenStart(), s.TokenFlags(), true
		}
	case ast.KindRegularExpressionLiteral:
		if s.ReScanSlashToken() == kind && s.TokenEnd() == end {
			return s.TokenStart(), s.TokenFlags(), true
		}
	case ast.KindTemplateMiddle, ast.KindTemplateTail:
		if s.ReScanTemplateToken(false) == kind && s.TokenEnd() == end {
			return s.TokenStart(), s.TokenFlags(), true
		}
	case ast.KindJsxText, ast.KindJsxTextAllWhiteSpaces, ast.KindLessThanSlashToken:
		if s.ReScanJsxToken(true) == kind && s.TokenEnd() == end {
			return s.TokenStart(), s.TokenFlags(), true
		}
	case ast.KindHashToken:
		if s.ReScanHashToken() == kind && s.TokenEnd() == end {
			return s.TokenStart(), s.TokenFlags(), true
		}
	case ast.KindAsteriskEqualsToken:
		if s.ReScanAsteriskEqualsToken() == kind && s.TokenEnd() == end {
			return s.TokenStart(), s.TokenFlags(), true
		}
	case ast.KindStringLiteral:
		if s.ReScanJsxAttributeValue() == kind && s.TokenEnd() == end {
			return s.TokenStart(), s.TokenFlags(), true
		}
	}
	return 0, ast.TokenFlagsNone, false
}

// createEphemeralToken builds a request-local token node without mutating
// SourceFile.tokenCache.
func createEphemeralToken(f *ast.SourceFile, kind ast.Kind, pos, tokenStart, end int, parent *ast.Node, flags ast.TokenFlags) *ast.Node {
	factory := ast.NewNodeFactory(ast.NodeFactoryHooks{})
	text := f.Text()[tokenStart:end]
	var tok *ast.Node
	switch kind {
	case ast.KindNumericLiteral:
		tok = factory.NewNumericLiteral(text, flags)
	case ast.KindBigIntLiteral:
		tok = factory.NewBigIntLiteral(text, flags)
	case ast.KindStringLiteral:
		tok = factory.NewStringLiteral(text, flags)
	case ast.KindJsxText, ast.KindJsxTextAllWhiteSpaces:
		tok = factory.NewJsxText(text, kind == ast.KindJsxTextAllWhiteSpaces)
	case ast.KindRegularExpressionLiteral:
		tok = factory.NewRegularExpressionLiteral(text, flags)
	case ast.KindNoSubstitutionTemplateLiteral:
		tok = factory.NewNoSubstitutionTemplateLiteral(text, flags)
	case ast.KindTemplateHead:
		tok = factory.NewTemplateHead(text, "", flags)
	case ast.KindTemplateMiddle:
		tok = factory.NewTemplateMiddle(text, "", flags)
	case ast.KindTemplateTail:
		tok = factory.NewTemplateTail(text, "", flags)
	case ast.KindIdentifier:
		tok = factory.NewIdentifier(text)
	case ast.KindPrivateIdentifier:
		tok = factory.NewPrivateIdentifier(text)
	default:
		tok = factory.NewToken(kind)
	}
	tok.Loc = core.NewTextRange(pos, end)
	tok.Parent = parent
	return tok
}

func encodeGeneratedNode(node *ast.Node, builder *checker.NodeBuilder, enclosing *ast.Node) (Result, error) {
	if node == nil {
		return &EncodedTypeNode{}, nil
	}
	tree, table, err := encoder.EncodeNode(node, nil)
	if err != nil {
		return nil, err
	}
	result := &EncodedTypeNode{tree: tree}
	emit := builder.EmitContext()
	for index, generated := range table.Nodes {
		if flags := emit.EmitFlags(generated); flags != 0 {
			if result.emitFlags == nil {
				result.emitFlags = make([]uint32, len(table.Nodes))
			}
			if flags&printer.EFSingleLine != 0 {
				result.emitFlags[index] |= 1
			}
			if flags&printer.EFNoIndentation != 0 {
				result.emitFlags[index] |= 2
			}
			if flags&printer.EFNoAsciiEscaping != 0 {
				result.emitFlags[index] |= 4
			}
		}
		origin := emit.ParseNode(generated)
		if origin != nil && ast.GetSourceFileOfNode(origin) != nil {
			if result.parseNodes == nil {
				result.parseNodes = make([]*ast.Node, len(table.Nodes))
			}
			result.parseNodes[index] = origin
		}
	}
	prepareGeneratedSyntax(tree, table.Nodes, result.parseNodes, ast.GetSourceFileOfNode(enclosing))
	return result, nil
}

func overloadImplementation(c *checker.Checker, node *ast.Node) Boolean {
	if !ast.IsFunctionLike(node) || !ast.NodeIsPresent(node.Body()) ||
		ast.IsGetAccessorDeclaration(node) || ast.IsSetAccessorDeclaration(node) {
		return false
	}
	signatures := checker.Checker_getSignaturesOfSymbol(c, checker.Checker_getSymbolOfDeclaration(c, node))
	if len(signatures) > 1 {
		return true
	}
	if len(signatures) == 1 {
		decl := signatures[0].Declaration()
		return Boolean(decl != nil && decl != node && decl.Flags&ast.NodeFlagsJSDoc == 0)
	}
	return false
}

func constantValue(value any) Literal {
	switch v := value.(type) {
	case string:
		return Literal{Kind: 1, Text: v}
	case jsnum.Number:
		return Literal{Kind: 2, Number: float64(v)}
	default:
		return Literal{}
	}
}

// Go inserts three write-access bits before ContainsStatic and has private
// cache flags above Unresolved. Neither group exists in JS CheckFlags.
func jsCheckFlags(flags ast.CheckFlags) uint32 {
	const low = ast.CheckFlagsContainsWritePublic - 1
	const high = (ast.CheckFlagsUnresolved*2 - 1) &^ (ast.CheckFlagsContainsStatic - 1)
	return uint32(flags&low | (flags&high)>>3)
}

func diagnosticFacts(diagnostics []*ast.Diagnostic) SuggestionDiagnostics {
	if len(diagnostics) == 0 {
		return nil
	}
	result := make(SuggestionDiagnostics, len(diagnostics))
	for i, diagnostic := range diagnostics {
		d := DiagnosticFact{
			Code:        uint32(diagnostic.Code()),
			Category:    uint32(diagnostic.Category()),
			Message:     diagnostic.Localize(locale.Default),
			Source:      diagnostic.Source(),
			Deprecated:  diagnostic.ReportsDeprecated(),
			Unnecessary: diagnostic.ReportsUnnecessary(),
			Chain:       diagnosticFacts(diagnostic.MessageChain()),
			Related:     diagnosticFacts(diagnostic.RelatedInformation()),
		}
		if f := diagnostic.File(); f != nil {
			d.File = f.FileName()
			d.FilePath = string(f.Path())
			positions := f.GetPositionMap()
			start := positions.UTF8ToUTF16(diagnostic.Pos())
			d.Start = uint32(start)
			d.Length = uint32(positions.UTF8ToUTF16(diagnostic.End()) - start)
		}
		result[i] = d
	}
	return result
}

func isJSDocOverloadDeclaration(decl *ast.Node) bool {
	return decl.Flags&ast.NodeFlagsReparsed != 0 && (ast.IsFunctionDeclaration(decl) || ast.IsMethodDeclaration(decl) || ast.IsConstructorDeclaration(decl))
}

// Choose the authored value, or the first declaration for a type-only link.
func symbolTargetDeclaration(symbol *ast.Symbol) *ast.Node {
	value := symbol.ValueDeclaration
	if value != nil && !isJSDocOverloadDeclaration(value) {
		return value
	}
	afterValue := value == nil
	for _, decl := range symbol.Declarations {
		if afterValue && !isJSDocOverloadDeclaration(decl) {
			return decl
		}
		if decl == value {
			afterValue = true
		}
	}
	return value
}

func symbolDeclarations(symbol *ast.Symbol, snapshot bool) ([]*ast.Node, *ast.Node) {
	nodes := symbol.Declarations
	value := symbol.ValueDeclaration
	var kept []*ast.Node
	start := 0
	useImplementation := false
	// The JS parser inserts generated overloads immediately before their
	// implementation. Usually all authored declarations form a single suffix.
	for i, decl := range nodes {
		if isJSDocOverloadDeclaration(decl) {
			if start < i {
				kept = append(kept, nodes[start:i]...)
			}
			start = i + 1
			useImplementation = useImplementation || decl == value
		} else if useImplementation {
			value = decl
			useImplementation = false
		}
	}
	if kept != nil {
		return append(kept, nodes[start:]...), value
	}
	declarations := nodes[start:]
	if snapshot {
		declarations = slices.Clone(declarations)
	}
	return declarations, value
}

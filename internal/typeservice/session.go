package typeservice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/microsoft/TypeScript/tsc/shim/api/encoder"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/collections"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/jsnum"
	"github.com/microsoft/TypeScript/tsc/shim/locale"
	"github.com/microsoft/TypeScript/tsc/shim/nodebuilder"
	"github.com/microsoft/TypeScript/tsc/shim/printer"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/utils"
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
	mu           sync.Mutex
	program      *program.Program
	syntaxOnly   bool
	anchor       *ast.SourceFile
	eligible     map[*ast.SourceFile]bool
	checker      *checker.Checker
	generation   uint32
	types        map[uint32]*checker.Type
	symbols      map[uint32]*ast.Symbol
	signatures   map[uint32]*checker.Signature
	indexInfoIDs map[*checker.IndexInfo]uint32
}

type Descriptor struct {
	SyntaxOnly          bool              `json:"syntaxOnly,omitempty"`
	Generation          uint32            `json:"generation"`
	CurrentDirectory    string            `json:"currentDirectory"`
	CompilerOptions     json.RawMessage   `json:"compilerOptions"`
	ConfigFile          *configSourceInfo `json:"configFile,omitempty"`
	ImplicitImportTrees map[string][]byte `json:"implicitImportTrees,omitempty"`
	Sources             []Source          `json:"sources"`
}
type Source struct {
	ImpliedNodeFormat core.ResolutionMode `json:"impliedNodeFormat,omitempty"`
	CheckJSDirective  *checkJSDirective   `json:"checkJsDirective,omitempty"`
	FileName          string              `json:"fileName"`
	Path              string              `json:"path"`
	Packages          []string            `json:"packages"`
	ImplicitImports   []string            `json:"implicitImports,omitempty"`
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

func (s *Session) BindTarget(file *ast.SourceFile) (Binding, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.program == nil {
		return Binding{}, ErrClosed
	}
	if !s.eligible[file] {
		return Binding{}, errors.New("source is not eligible for typed rules")
	}
	return Binding{Scope{Generation: s.generation}, file.FileName()}, nil
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
	file := s.program.GetSourceFile(name)
	if file == nil && !s.syntaxOnly {
		path := tspath.ToPath(name, s.program.CurrentDirectory(), s.program.FS().UseCaseSensitiveFileNames())
		if c := parsedConfig(s.program.CompilerQueries(), path); c != nil && c.ConfigFile != nil {
			file = c.ConfigFile.SourceFile
		}
	}
	if s.syntaxOnly && !s.eligible[file] {
		return nil, errors.New("foreign syntax service source")
	}
	return file, nil //nolint:nilnil // A missing Program source is a legitimate result.
}

// New accepts exact Program sources that the caller has explicitly admitted
// for typed rules. Checker capability alone is never eligibility.
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
	s := &Session{program: p, syntaxOnly: syntaxOnly, anchor: eligible[0], eligible: make(map[*ast.SourceFile]bool, len(eligible)), generation: nextGeneration.Add(1)}
	for _, f := range eligible {
		if !p.OwnsSourceFile(f) || (!syntaxOnly && !p.CanProvideTypeChecker(f)) {
			return nil, errors.New("ineligible or foreign type service source")
		}
		s.eligible[f] = true
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
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
	s.indexInfoIDs = nil
}

// SyntaxDescriptor opens only the requested target for a file-scoped reader.
func (s *Session) SyntaxDescriptor(file *ast.SourceFile) (Descriptor, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.program == nil {
		return Descriptor{}, ErrClosed
	}
	if !s.syntaxOnly || !s.eligible[file] {
		return Descriptor{}, errors.New("foreign syntax projection target")
	}
	d := s.syntaxDescriptor()
	d.Sources = []Source{syntaxSource(file)}
	return d, nil
}
func (s *Session) syntaxDescriptor() Descriptor {
	return Descriptor{SyntaxOnly: true, Generation: s.generation, CurrentDirectory: s.program.CurrentDirectory(), CompilerOptions: json.RawMessage("{}")}
}
func syntaxSource(f *ast.SourceFile) Source {
	source := Source{FileName: f.FileName(), Path: string(f.Path()), IsDeclarationFile: f.IsDeclarationFile}
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
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.program == nil {
		return Descriptor{}, ErrClosed
	}
	if s.syntaxOnly {
		d := s.syntaxDescriptor()
		for _, file := range s.program.SourceFiles() {
			if s.eligible[file] {
				d.Sources = append(d.Sources, syntaxSource(file))
			}
		}
		return d, nil
	}
	compiler := s.program.CompilerQueries()
	commandLine := compiler.CommandLine()
	options, err := json.Marshal(compilerOptions(commandLine))
	if err != nil {
		return Descriptor{}, err
	}
	d := Descriptor{Generation: s.generation, CurrentDirectory: s.program.CurrentDirectory(), CompilerOptions: options, ConfigFile: configSource(s.program, commandLine)}
	for _, f := range s.program.SourceFiles() {
		source := s.source(f)
		for _, specifier := range compiler.ImplicitImports(f.Path()) {
			name := specifier.Text()
			source.ImplicitImports = append(source.ImplicitImports, name)
			if d.ImplicitImportTrees[name] == nil {
				tree, _, err := encoder.EncodeNode(specifier.Parent, nil)
				if err != nil {
					return Descriptor{}, err
				}
				if d.ImplicitImportTrees == nil {
					d.ImplicitImportTrees = make(map[string][]byte)
				}
				d.ImplicitImportTrees[name] = tree
			}
		}
		d.Sources = append(d.Sources, source)
	}
	return d, nil
}

// Embed the native options so nested maps such as paths retain their key order.
type compilerOptionsJSON struct {
	*core.CompilerOptions
	JSX             string `json:"jsx,omitempty"`
	ModuleDetection string `json:"moduleDetection,omitempty"`
	NewLine         *int   `json:"newLine,omitempty"`
	Plugins         []any  `json:"plugins,omitzero"`
}

func compilerOptions(c *tsoptions.ParsedCommandLine) compilerOptionsJSON {
	opts := c.CompilerOptions()
	options := compilerOptionsJSON{
		CompilerOptions: opts,
		JSX:             map[core.JsxEmit]string{core.JsxEmitPreserve: "Preserve", core.JsxEmitReact: "React", core.JsxEmitReactNative: "ReactNative", core.JsxEmitReactJSX: "ReactJSX", core.JsxEmitReactJSXDev: "ReactJSXDev"}[opts.Jsx],
		ModuleDetection: map[core.ModuleDetectionKind]string{core.ModuleDetectionKindAuto: "Auto", core.ModuleDetectionKindLegacy: "Legacy", core.ModuleDetectionKindForce: "Force"}[opts.ModuleDetection],
		Plugins:         configPlugins(c.Raw),
	}
	if opts.NewLine != core.NewLineKindNone {
		value := int(opts.NewLine) - 1
		options.NewLine = &value
	}
	return options
}

type jsonObject = collections.OrderedMap[string, any]

// tsgo keeps language-service plugins only in the config file's raw options.
func configPlugins(raw any) []any {
	config, _ := raw.(*jsonObject)
	if config == nil {
		return nil
	}
	opts, _ := config.GetOrZero("compilerOptions").(*jsonObject)
	if opts == nil {
		return nil
	}
	values, _ := opts.GetOrZero("plugins").([]any)
	if values == nil {
		return nil
	}
	plugins := make([]any, 0, len(values))
	for _, value := range values {
		switch value.(type) {
		case *jsonObject, []any:
			plugins = append(plugins, value)
		}
	}
	return plugins
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
	eligible := s.eligible // immutable after construction
	for _, q := range queries {
		if q.Generation != s.generation {
			s.mu.Unlock()
			return nil, errors.New("stale or foreign type service handle")
		}
		if q.Request == nil {
			s.mu.Unlock()
			return nil, errors.New("missing type service request")
		}
		if s.syntaxOnly && q.Request.operation() != "nodeRecord" {
			s.mu.Unlock()
			return nil, errors.New("semantic queries require a typed project")
		}
	}
	s.mu.Unlock()
	natives := make([]queryNodes, len(queries))
	for i, q := range queries {
		if ref := requestNode(q.Request); ref != nil {
			f := sourceForNode(p, ref)
			if s.syntaxOnly && !eligible[f] {
				return nil, errors.New("foreign syntax service source")
			}
			natives[i].primary, err = resolveNode(p, f, *ref)
			if err != nil {
				return nil, err
			}
		}
		if r, ok := q.Request.(*IndexDeclarationRequest); ok && r.Declaration != nil {
			natives[i].declaration, err = resolveNode(p, sourceForNode(p, r.Declaration), *r.Declaration)
			if err != nil {
				return nil, err
			}
		}
	}
	results := make([]Result, len(queries))
	records := make([][]Record, len(queries))
	// Program diagnostics may acquire all checker shards.
	// Preserve request order, releasing the file checker before these reads.
	for start := 0; start < len(queries); {
		if r, ok := queries[start].Request.(*ProgramRequest); ok {
			results[start], err = programQuery(ctx, p, r, natives[start].primary)
			if err != nil {
				return nil, err
			}
			start++
			continue
		}
		if queries[start].Request.operation() == "nodeRecord" {
			ref := requestNode(queries[start].Request)
			results[start] = nodeFacts(p, sourceForNode(p, ref), ref.Config, natives[start].primary)
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
	case *DeclaredMembers:
		if fact != nil {
			materializeReferences(p, fact.indices)
		}
	case *TypeDetails:
		fact.patternRef, fact.nodeRef, fact.rootRef = nodeReference(p, fact.pattern), nodeReference(p, fact.node), nodeReference(p, fact.rootNode)
		fact.labelRefs = nodeRefs(p, fact.labels)
	case *SymbolFact:
		if fact.nodes != nil {
			fact.Declarations = make([]NodeRef, 0, len(fact.nodes))
			for _, node := range fact.nodes {
				ref := nodeReference(p, node)
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
			fact[i].Components = nodeRefs(p, fact[i].componentNodes)
			fact[i].componentNodes = nil
			fact[i].node = nil
		}
	}
}
func (s *Session) checkBatch(ctx context.Context, queries []Query, natives []queryNodes, results []Result, records [][]Record) error {
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
		records[i], err = s.directRecords(ctx, c, q.Request.operation(), results[i])
		if err != nil {
			return err
		}
	}
	return nil
}
func (s *Session) directRecords(ctx context.Context, c *checker.Checker, op Operation, result Result) ([]Record, error) {
	var kind Operation
	switch op {
	case "type", "fromTypeNode", "context", "constraint", "apparent", "baseLiteral", "nonNullable", "void", "declaredType", "symbolType", "returnType", "numberIndexType", "typeArguments",
		"stringIndexType", "awaited", "widened", "baseTypes", "defaultType", "typeOfSymbol", "indexInfo", "indexInfos", "typePredicate", "parameterType", "propertyType",
		"mappedFields", "classThisType", "contextualArgument":
		kind = "typeRecord"
	case "symbol", "aliasedSymbol", "property", "properties", "resolveName",
		"apparentProperties", "exportSymbol", "immediateAliasedSymbol", "shorthandValueSymbol":
		kind = "symbolRecord"
	case "signature", "resolvedSignature", "resolvedSignatureWithOptions", "signatures":
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
		fact, err := s.execute(ctx, c, request, queryNodes{})
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
func nodeFacts(p *program.Program, file *ast.SourceFile, config string, node *ast.Node) NodeFacts {
	facts := NodeFacts{}
	seen := map[*ast.Node]bool{}
	table := encoder.GetNodeIndexTable(file)
	reference := func(n *ast.Node) *NodeRef { return indexedNodeReference(p, file, table, config, n) }
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
		fact := NodeFact{Node: *ref, Kind: strings.TrimPrefix(n.Kind.String(), "Kind"), Flags: uint32(n.Flags), Parent: reference(syntaxParent(n))}
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
	var config string
	if !p.OwnsSourceFile(f) {
		if f == nil {
			panic("declaration is outside Program")
		}
		config = configID(f)
		if configFile(p, config) != f {
			panic("declaration is outside Program")
		}
	}
	return indexedNodeReference(p, f, encoder.GetNodeIndexTable(f), config, n)
}

func indexedNodeReference(p *program.Program, f *ast.SourceFile, table *encoder.NodeIndexTable, config string, n *ast.Node) *NodeRef {
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
	if index == 0 && config == "" && p.CompilerQueries() != nil {
		for i, specifier := range p.CompilerQueries().ImplicitImports(f.Path()) {
			switch n {
			case specifier.Parent:
				index = uint32(len(table.Nodes) + 2*i)
			case specifier:
				index = uint32(len(table.Nodes) + 2*i + 1)
			}
		}
	}
	if index == 0 {
		panic("declaration has no compiler node handle")
	}
	return &NodeRef{File: f.FileName(), Index: index, Config: config}
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
	if s.indexInfoIDs == nil {
		s.indexInfoIDs = make(map[*checker.IndexInfo]uint32)
	}
	id := s.indexInfoIDs[info]
	if id == 0 {
		id = uint32(len(s.indexInfoIDs)) + 1
		s.indexInfoIDs[info] = id
	}
	return IndexInfo{
		ID:             id,
		KeyType:        s.typeID(info.KeyType()),
		Type:           s.typeID(info.ValueType()),
		IsReadonly:     info.IsReadonly(),
		node:           info.Declaration(),
		componentNodes: checker.IndexInfo_components(info),
	}
}

type queryNodes struct{ primary, declaration *ast.Node }

func (s *Session) execute(ctx context.Context, c *checker.Checker, request Request, nodes queryNodes) (Result, error) {
	native := nodes.primary
	switch r := request.(type) {
	case *NodeRequest:
		switch r.Op {
		case "exportSpecifierLocalTargetSymbol":
			return Handle(s.symbolID(c.GetExportSpecifierLocalTargetSymbol(native))), nil
		case "propertySymbolOfDestructuringAssignment":
			return Handle(s.symbolID(c.GetPropertySymbolOfDestructuringAssignment(native))), nil
		case "jsxIntrinsicTagNamesAt":
			return s.namedSymbols(c.GetJsxIntrinsicTagNamesAt(native)), nil
		case "typeOfAssignmentPattern":
			return Handle(s.typeID(checker.Checker_getTypeOfAssignmentPattern(c, native))), nil
		case "isOptionalParameter":
			return Boolean(checker.Checker_isOptionalParameter(c, native)), nil
		case "constantValue":
			return constantValue(c.GetConstantValue(native)), nil
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
			return Handle(s.signatureID(c.GetSignatureFromDeclaration(native))), nil
		case "resolvedSignature":
			if !ast.IsCallLikeExpression(native) {
				return Handle(0), nil
			}
			return Handle(s.signatureID(c.GetResolvedSignature(native))), nil
		case "shorthandValueSymbol":
			return Handle(s.symbolID(c.GetShorthandAssignmentValueSymbol(native))), nil
		case "suggestionDiagnostics":
			if native.Kind != ast.KindSourceFile {
				return nil, errors.New("expected a source file for suggestion diagnostics")
			}
			return diagnosticFacts(s.program, c.GetSuggestionDiagnostics(ctx, native.AsSourceFile())), ctx.Err()
		}
	case *StringLiteralRequest:
		return Handle(s.typeID(checker.Checker_getStringLiteralType(c, string(r.Name)))), nil
	case *NumberLiteralRequest:
		value, err := strconv.ParseFloat(r.Value, 64)
		if err != nil {
			return nil, err
		}
		return Handle(s.typeID(checker.Checker_getNumberLiteralType(c, jsnum.Number(value)))), nil
	case *BigIntLiteralRequest:
		return Handle(s.typeID(checker.Checker_getBigIntLiteralType(c, jsnum.NewPseudoBigInt(r.Digits, r.Negative)))), nil
	case *NullableRequest:
		flags := checker.TypeFlagsNone
		if r.Nullable {
			flags |= checker.TypeFlagsNull
		}
		if r.Undefinable {
			flags |= checker.TypeFlagsUndefined
		}
		return Handle(s.typeID(checker.Checker_getNullableType(c, s.typeByID(r.ID), flags))), nil
	case *PrivatePropertyRequest:
		symbol := checker.Checker_lookupSymbolForPrivateIdentifierDeclaration(c, string(r.Name), native)
		if symbol == nil {
			return Handle(0), nil
		}
		return Handle(s.symbolID(checker.Checker_getPrivateIdentifierPropertyOfType(c, s.typeByID(r.ID), symbol))), nil
	case *NamedNodeRequest:
		if r.Op == "validPropertyAccess" {
			return Boolean(c.IsValidPropertyAccess(native, string(r.Name))), nil
		}
		parameter, property := c.GetSymbolsOfParameterPropertyDeclaration(native, string(r.Name))
		return Handles{s.symbolID(parameter), s.symbolID(property)}, nil
	case *IndexSymbolRequest:
		symbol := s.symbolByID(r.ID)
		var siblings []*ast.Symbol
		if r.Siblings == nil && symbol.Parent != nil {
			// The TS checker appends late-bound members to the binder's table.
			var late []*ast.Symbol
			for name, member := range checker.Checker_getMembersOfSymbol(c, symbol.Parent) {
				if symbol.Parent.Members[name] != nil {
					siblings = append(siblings, member)
				} else {
					late = append(late, member)
				}
			}
			slices.SortFunc(siblings, c.CompareSymbols)
			slices.SortFunc(late, c.CompareSymbols)
			siblings = append(siblings, late...)
		} else {
			for _, id := range r.Siblings {
				siblings = append(siblings, s.symbolByID(id))
			}
		}
		infos := checker.Checker_getIndexInfosOfIndexSymbol(c, symbol, siblings)
		result := make(IndexInfos, len(infos))
		for i, info := range infos {
			result[i] = s.indexInfo(info)
		}
		return result, nil
	case *ExpandoRequest:
		return Handle(s.symbolID(expandoSymbol(c, native, r.AllowDeclaration))), nil
	case *PredicateStringRequest:
		var typ *checker.Type
		if r.Type != 0 {
			typ = s.typeByID(r.Type)
		}
		predicate := checker.Checker_newTypePredicate(c, checker.TypePredicateKind(r.Kind), string(r.Name), r.Position, typ)
		return Text(checker.Checker_typePredicateToStringEx(c, predicate, native, checker.TypeFormatFlags(r.Flags))), nil
	case *IndexDeclarationRequest:
		// The JS caller can supply a structural IndexInfo. This temporary value
		// belongs to the print request, outside the checker's long-lived arena.
		var info checker.IndexInfo
		*checker.IndexInfo_keyTypePtr(&info) = s.typeByID(r.Key)
		*checker.IndexInfo_valueTypePtr(&info) = s.typeByID(r.Value)
		*checker.IndexInfo_isReadonlyPtr(&info) = r.Readonly
		*checker.IndexInfo_declarationPtr(&info) = nodes.declaration
		builder := checker.NewNodeBuilder(c, printer.NewEmitContext())
		return encodeGeneratedNode(builder.IndexInfoToIndexSignatureDeclaration(&info, native, nodebuilder.Flags(r.Flags), nodebuilder.InternalFlagsNone, nil), builder, native)
	case *ContextualArgumentRequest:
		if !ast.IsCallLikeExpression(native) {
			return Handle(0), nil
		}
		return Handle(s.typeID(c.GetContextualTypeForArgumentAtIndex(native, r.Index))), nil
	case *ResolvedSignatureRequest:
		ids := Handles{0}
		if !ast.IsCallLikeExpression(native) {
			return ids, nil
		}
		var candidates []*checker.Signature
		var output *[]*checker.Signature
		if r.WithCandidates {
			output = &candidates
		}
		count := checker.Checker_apparentArgumentCountPtr(c)
		saved := *count
		*count = r.ArgumentCount
		defer func() { *count = saved }()
		sig := checker.Checker_getResolvedSignature(c, native, output, checker.CheckModeNormal)
		ids[0] = s.signatureID(sig)
		for _, candidate := range candidates {
			ids = append(ids, s.signatureID(candidate))
		}
		return ids, nil
	case *TypeRequest:
		switch r.Op {
		case "augmentedPropertiesOfType":
			return s.namedSymbols(checker.Checker_getAugmentedPropertiesOfType(c, s.typeByID(r.ID))), nil
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
		case "conditionalInstantiations":
			root := checker.ConditionalType_root(s.typeByID(r.ID).AsConditionalType())
			table := checker.ConditionalRoot_instantiations(root)
			entries := make(NamedSymbols, 0, len(table))
			for key, typ := range table {
				entries = append(entries, NamedSymbol{Name: fmt.Sprintf("%016x%016x", key.Hi, key.Lo), ID: s.typeID(typ)})
			}
			slices.SortFunc(entries, func(a, b NamedSymbol) int { return strings.Compare(a.Name, b.Name) })
			return entries, nil
		case "freshLiteral", "evolvingElement", "finalArray", "typeParameterConstraint":
			t := s.typeByID(r.ID)
			var value *checker.Type
			switch r.Op {
			case "freshLiteral":
				if t.Flags()&(checker.TypeFlagsLiteral|checker.TypeFlagsEnum) != 0 {
					value = t.AsLiteralType().FreshType()
				}
			case "evolvingElement", "finalArray":
				if t.ObjectFlags()&checker.ObjectFlagsEvolvingArray != 0 {
					arr := t.AsEvolvingArrayType()
					if r.Op == "evolvingElement" {
						value = checker.EvolvingArrayType_elementType(arr)
					} else {
						value = checker.EvolvingArrayType_finalArrayType(arr)
					}
				}
			case "typeParameterConstraint":
				if t.Flags()&checker.TypeFlagsTypeParameter != 0 {
					value = checker.TypeParameter_constraint(t.AsTypeParameter())
				}
			}
			return Handle(s.typeID(value)), nil
		case "declaredMembers":
			return s.declaredMembers(c, s.typeByID(r.ID)), nil
		case "typeDetails":
			return s.typeDetails(c, s.typeByID(r.ID)), nil
		case "resolvedTrueType", "resolvedFalseType":
			t := s.typeByID(r.ID)
			if t.Flags()&checker.TypeFlagsConditional == 0 {
				return Handle(0), nil
			}
			cond := t.AsConditionalType()
			if r.Op == "resolvedTrueType" {
				return Handle(s.typeID(checker.ConditionalType_resolvedTrueType(cond))), nil
			}
			return Handle(s.typeID(checker.ConditionalType_resolvedFalseType(cond))), nil
		case "classThisType":
			t := s.typeByID(r.ID).AsInterfaceType()
			if t == nil {
				return Handle(0), nil
			}
			return Handle(s.typeID(checker.InterfaceType_thisType(t))), nil
		case "mappedFields":
			t := s.typeByID(r.ID)
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
				// Fresh array literals can copy the target's Tuple flag onto a
				// reference. Only the definition owns type parameters and tuple data.
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
	case *SymbolRequest:
		switch r.Op {
		case "rootSymbols":
			return s.namedSymbols(c.GetRootSymbols(s.symbolByID(r.ID))), nil
		case "mergedSymbol":
			return Handle(s.symbolID(c.GetMergedSymbol(s.symbolByID(r.ID)))), nil
		case "isArgumentsSymbol":
			return Boolean(c.GetArgumentsSymbol() == s.symbolByID(r.ID)), nil
		case "isUndefinedSymbol":
			return Boolean(c.GetUndefinedSymbol() == s.symbolByID(r.ID)), nil
		case "aliasedSymbol":
			return Handle(s.symbolID(c.GetAliasedSymbol(s.symbolByID(r.ID)))), nil
		case "immediateAliasedSymbol":
			return Handle(s.symbolID(c.GetImmediateAliasedSymbol(s.symbolByID(r.ID)))), nil
		case "isUnknownSymbol":
			return Boolean(c.IsUnknownSymbol(s.symbolByID(r.ID))), nil
		case "exportsOfModule":
			exports := c.GetExportsOfModule(s.symbolByID(r.ID))
			slices.SortFunc(exports, c.CompareSymbols)
			return s.namedSymbols(exports), nil
		case "fullyQualifiedName":
			return Text(c.GetFullyQualifiedName(s.symbolByID(r.ID))), nil
		case "symbolMembers", "symbolExports", "symbolGlobalExports":
			symbol := s.symbolByID(r.ID)
			table := symbol.Members
			if r.Op == "symbolExports" {
				table = symbol.Exports
			}
			if r.Op == "symbolGlobalExports" {
				table = globalExports(symbol)
			}
			members := slices.Collect(maps.Values(table))
			slices.SortFunc(members, c.CompareSymbols)
			result := make(NamedSymbols, len(members))
			for i, member := range members {
				result[i] = NamedSymbol{ID: s.symbolID(member), Name: ast.EscapeSymbolName(member.Name)}
			}
			return result, nil
		case "exportSymbol":
			return Handle(s.symbolID(c.GetExportSymbolOfSymbol(s.symbolByID(r.ID)))), nil
		case "declaredType":
			return Handle(s.typeID(c.GetDeclaredTypeOfSymbol(s.symbolByID(r.ID)))), nil
		case "typeOfSymbol":
			return Handle(s.typeID(c.GetTypeOfSymbol(s.symbolByID(r.ID)))), nil
		case "symbolRecord":
			sym := s.symbolByID(r.ID)
			// The record outlives the checker lease.
			declarations, value := symbolDeclarations(sym, true)
			// The JS binder creates these tables even for declarations with no entries.
			var declared ast.SymbolFlags
			if len(sym.Declarations) != 0 {
				bound := sym.Declarations[0].Symbol()
				if bound == sym || c.GetMergedSymbol(bound) == sym {
					declared = sym.Flags
				}
			}
			return &SymbolFact{
				Name: symbolName(sym), EscapedName: ast.EscapeSymbolName(sym.Name), Flags: uint32(sym.Flags), CheckFlags: jsCheckFlags(sym.CheckFlags),
				HasMembers:       sym.Members != nil || declared&(ast.SymbolFlagsClass|ast.SymbolFlagsInterface|ast.SymbolFlagsTypeLiteral|ast.SymbolFlagsObjectLiteral) != 0,
				HasExports:       sym.Exports != nil || declared&(ast.SymbolFlagsClass|ast.SymbolFlagsEnum|ast.SymbolFlagsModule|ast.SymbolFlagsVariable) != 0,
				HasGlobalExports: globalExports(sym) != nil, nodes: declarations, valueNode: value,
			}, nil
		}
	case *SymbolDocsRequest:
		if !r.WithChecker {
			c = nil
		}
		symbol := s.symbolByID(r.ID)
		active := map[*ast.Symbol]bool{symbol: true}
		declarations, _ := symbolDeclarations(symbol, false)
		if r.Op == "symbolDocumentation" {
			return documentation(c, declarations, active), nil
		}
		return jsDocTags(c, declarations, active), nil
	case *SignatureRequest:
		switch r.Op {
		case "typeArgumentsForResolvedSignature":
			return OptionalHandles(s.typeIDs(resolvedSignatureTypeArguments(c, s.signatureByID(r.ID)))), nil
		case "returnType":
			return Handle(s.typeID(c.GetReturnTypeOfSignature(s.signatureByID(r.ID)))), nil
		case "signatureDocumentation", "signatureJSDocTags":
			declarations := []*ast.Node{s.signatureByID(r.ID).Declaration()}
			active := make(map[*ast.Symbol]bool)
			if r.Op == "signatureDocumentation" {
				return documentation(c, declarations, active), nil
			}
			return jsDocTags(c, declarations, active), nil
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
	case *CheckerRequest:
		switch r.Op {
		case "typeCount":
			return Handle(c.TypeCount), nil
		case "symbolCount":
			count := c.SymbolCount
			for _, file := range s.program.SourceFiles() {
				count += uint32(file.SymbolCount)
			}
			return Handle(count), nil
		case "instantiationCount":
			return Handle(c.TotalInstantiationCount), nil
		case "relationCacheSizes":
			return Handles{
				uint32(len(checker.Relation_results(checker.Checker_assignableRelation(c)))),
				uint32(len(checker.Relation_results(checker.Checker_identityRelation(c)))),
				uint32(len(checker.Relation_results(checker.Checker_subtypeRelation(c)))),
				uint32(len(checker.Relation_results(checker.Checker_strictSubtypeRelation(c)))),
			}, nil
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
			return Handle(s.typeID(checker.Checker_regularTrueType(c))), nil
		case "falseType":
			return Handle(s.typeID(checker.Checker_regularFalseType(c))), nil
		case "ambientModules":
			modules := slices.Clone(c.GetAmbientModules())
			slices.SortFunc(modules, c.CompareSymbols)
			return s.namedSymbols(modules), nil
		}
	case *VoidRequest:
		return Handle(s.typeID(c.GetVoidType())), nil
	case *PropertyRequest:
		switch r.Op {
		case "moduleMember":
			return Handle(s.symbolID(c.TryGetMemberInModuleExports(string(r.Name), s.symbolByID(r.ID)))), nil
		case "property":
			return Handle(s.symbolID(c.GetPropertyOfType(s.typeByID(r.ID), string(r.Name)))), nil
		case "propertyType":
			return Handle(s.typeID(c.GetTypeOfPropertyOfType(s.typeByID(r.ID), string(r.Name)))), nil
		}
	case *SymbolTypeRequest:
		return Handle(s.typeID(c.GetTypeOfSymbolAtLocation(s.symbolByID(r.ID), native))), nil
	case *ResolveNameRequest:
		return Handle(s.symbolID(resolveName(c, string(r.Name), native, ast.SymbolFlags(r.Kind), r.ExcludeGlobals))), nil
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
		return s.namedSymbols(symbolsInScope(c, native, ast.SymbolFlags(r.Meaning))), nil
	case *ParameterTypeRequest:
		return Handle(s.typeID(c.GetTypeParameterAtPosition(s.signatureByID(r.ID), r.Position))), nil
	case *PrintRequest:
		var builderFlags nodebuilder.Flags
		if r.Flags != nil {
			builderFlags = nodebuilder.Flags(*r.Flags)
		}
		switch r.Op {
		case "symbolString":
			flags := checker.SymbolFormatFlagsAllowAnyNodeKind
			if r.Flags != nil {
				flags = checker.SymbolFormatFlags(*r.Flags)
			}
			return Text(c.SymbolToStringEx(s.symbolByID(r.ID), native, ast.SymbolFlags(r.Kind), flags)), nil
		case "typeNode", "signatureDeclaration", "parameterDeclaration", "typeParameterDeclarations", "typeParameterDeclaration", "entityName", "symbolExpression":
			// The shared string printer's builder can return a cached tree with
			// detached ranges. A returned syntax tree needs its own emit context.
			builder := checker.NewNodeBuilder(c, printer.NewEmitContext())
			var node *ast.Node
			switch r.Op {
			case "typeNode":
				node = builder.TypeToTypeNode(s.typeByID(r.ID), native, builderFlags, nodebuilder.InternalFlagsNone, nil)
			case "signatureDeclaration":
				node = builder.SignatureToSignatureDeclaration(s.signatureByID(r.ID), ast.Kind(r.Kind), native, builderFlags, nodebuilder.InternalFlagsNone, nil)
			case "parameterDeclaration":
				node = builder.SymbolToParameterDeclaration(s.symbolByID(r.ID), native, builderFlags, nodebuilder.InternalFlagsNone, nil)
			case "typeParameterDeclaration":
				node = builder.TypeParameterToDeclaration(s.typeByID(r.ID), native, builderFlags, nodebuilder.InternalFlagsNone, nil)
			case "entityName":
				node = builder.SymbolToEntityName(s.symbolByID(r.ID), ast.SymbolFlags(r.Kind), native, builderFlags, nodebuilder.InternalFlagsNone, nil)
			case "symbolExpression":
				node = builder.SymbolToExpression(s.symbolByID(r.ID), ast.SymbolFlags(r.Kind), native, builderFlags, nodebuilder.InternalFlagsNone, nil)
			case "typeParameterDeclarations":
				symbol := s.symbolByID(r.ID)
				target := c.GetTargetSymbol(symbol)
				if target.Flags&(ast.SymbolFlagsClass|ast.SymbolFlagsInterface|ast.SymbolFlagsTypeAlias) != 0 {
					params := c.GetLocalTypeParametersOfClassOrInterfaceOrTypeAlias(symbol)
					decls := make([]*ast.Node, len(params))
					for i, param := range params {
						decls[i] = builder.TypeParameterToDeclaration(param, native, builderFlags, nodebuilder.InternalFlagsNone, nil)
					}
					factory := ast.NewNodeFactory(ast.NodeFactoryHooks{})
					node = factory.NewFunctionTypeNode(factory.NewNodeList(decls), factory.NewNodeList(nil), nil)
				}
			}
			return encodeGeneratedNode(node, builder, native)
		}
		flags := checker.TypeFormatFlagsNone
		if r.Flags != nil {
			flags = checker.TypeFormatFlags(*r.Flags)
		} else if r.Op == "typeString" {
			flags = checker.TypeFormatFlagsAllowUniqueESSymbolType | checker.TypeFormatFlagsUseAliasDefinedOutsideCurrentScope
		}
		switch r.Op {
		case "typeString":
			return Text(typeToString(c, s.typeByID(r.ID), native, flags, s.program.Options().NoErrorTruncation == core.TSTrue)), nil
		case "signatureString":
			return Text(signatureToString(c, s.signatureByID(r.ID), native, flags, checker.SignatureKind(r.Kind))), nil
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

func sourceForNode(p *program.Program, ref *NodeRef) *ast.SourceFile {
	if ref.Config != "" {
		return configFile(p, ref.Config)
	}
	return p.GetSourceFile(ref.File)
}

func resolveNode(p *program.Program, f *ast.SourceFile, node NodeRef) (*ast.Node, error) {
	if f == nil {
		return nil, errors.New("foreign source handle")
	}
	table := encoder.GetNodeIndexTable(f)
	var parent *ast.Node
	if node.Index < uint32(len(table.Nodes)) {
		parent = table.Nodes[node.Index]
	} else if node.Config == "" && p.CompilerQueries() != nil {
		// The parser's node table excludes Program-owned implicit imports.
		// The source reader appends each native declaration followed by its literal.
		offset := node.Index - uint32(len(table.Nodes))
		imports := p.CompilerQueries().ImplicitImports(f.Path())
		if offset/2 < uint32(len(imports)) {
			parent = imports[offset/2]
			if offset%2 == 0 {
				parent = parent.Parent
			}
		}
	}
	if parent == nil {
		return nil, errors.New("invalid compiler node handle")
	}
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
	return f.GetOrCreateToken(token.Kind, pos, end, parent, ast.TokenFlagsNone), nil
}

func encodeGeneratedNode(node *ast.Node, builder *checker.NodeBuilder, enclosing *ast.Node) (Result, error) {
	if node == nil {
		return &EncodedTypeNode{}, nil
	}
	// Passing a SourceFile makes the encoder reserve space for its entire AST.
	// A generated subtree only needs that file's UTF-16 coordinate map.
	tree, table, err := encoder.EncodeNode(node, nil)
	if err != nil {
		return nil, err
	}
	result := &EncodedTypeNode{tree: tree}
	emit := builder.EmitContext()
	for index, generated := range table.Nodes {
		// Only these three flags are set by the node builder. Their positions
		// in the printer's EmitFlags differ between tsgo and TypeScript.
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

func overloadImplementation(c *checker.Checker, node *ast.Node) Boolean {
	if !ast.NodeIsPresent(node.Body()) || ast.IsGetAccessorDeclaration(node) || ast.IsSetAccessorDeclaration(node) {
		return false
	}
	signatures := checker.Checker_getSignaturesOfSymbol(c, checker.Checker_getSymbolOfDeclaration(c, node))
	return Boolean(len(signatures) > 1 || len(signatures) == 1 && signatures[0].Declaration() != node)
}

// Go inserts three write-access bits before ContainsStatic and has private
// cache flags above Unresolved. Neither group exists in JS CheckFlags.
func jsCheckFlags(flags ast.CheckFlags) uint32 {
	const low = ast.CheckFlagsContainsWritePublic - 1
	const high = (ast.CheckFlagsUnresolved*2 - 1) &^ (ast.CheckFlagsContainsStatic - 1)
	return uint32(flags&low | (flags&high)>>3)
}

func diagnosticFacts(p *program.Program, diagnostics []*ast.Diagnostic) SuggestionDiagnostics {
	return collectDiagnostics(p, diagnostics, make(map[*ast.SourceFile]*configSourceInfo))
}

func collectDiagnostics(p *program.Program, diagnostics []*ast.Diagnostic, configs map[*ast.SourceFile]*configSourceInfo) SuggestionDiagnostics {
	if len(diagnostics) == 0 {
		return nil
	}
	result := make(SuggestionDiagnostics, len(diagnostics))
	for i, diagnostic := range diagnostics {
		d := DiagnosticFact{
			Code: uint32(diagnostic.Code()), Category: uint32(diagnostic.Category()),
			Message: diagnostic.Localize(locale.Default), Source: diagnostic.Source(),
			Deprecated: diagnostic.ReportsDeprecated(), Unnecessary: diagnostic.ReportsUnnecessary(),
			Chain: collectDiagnostics(p, diagnostic.MessageChain(), configs), Related: collectDiagnostics(p, diagnostic.RelatedInformation(), configs),
		}
		if f := diagnostic.File(); f != nil {
			d.File = f.FileName()
			d.Config = configs[f]
			if d.Config == nil && !p.OwnsSourceFile(f) {
				if c := parsedConfig(p.CompilerQueries(), f.Path()); c != nil && c.ConfigFile != nil && c.ConfigFile.SourceFile == f {
					d.Config = configSource(p, c)
				} else {
					d.Config = registerConfig(p, f)
					d.Config.ExtendedSourceFiles = utils.ConfigAncestors(f)
				}
				configs[f] = d.Config
			}
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

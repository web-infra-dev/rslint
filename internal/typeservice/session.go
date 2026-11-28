package typeservice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/program"
)

var nextGeneration atomic.Uint32
var (
	ErrClosed   = errors.New("type service session is closed")
	ErrPoisoned = errors.New("type service session is unusable after an internal panic")
)

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
	closed       bool
	program      *program.Program
	syntaxOnly   bool
	syntax       *syntaxSessionData
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

// syntaxSessionData stores only detached value metadata and source text so
// syntax-only sessions never retain a Program or compiler SourceFile, while
// lazily caching each parsed AST across queries until Session.Close.
type syntaxSessionData struct {
	currentDirectory string
	compilerOptions  json.RawMessage
	caseSensitive    bool
	order            []string
	files            map[string]*syntaxSourceEntry
	byKey            map[string]string
}

type syntaxSourceEntry struct {
	fileName     string
	path         tspath.Path
	text         string
	parseOptions ast.SourceFileParseOptions
	scriptKind   core.ScriptKind
	desc         Source

	parseOnce sync.Once
	parsed    *ast.SourceFile
}

func (e *syntaxSourceEntry) sourceFile() *ast.SourceFile {
	e.parseOnce.Do(func() {
		e.parsed = parser.ParseSourceFile(e.parseOptions, e.text, e.scriptKind)
	})
	return e.parsed
}

func (d *syntaxSessionData) lookup(name string) *syntaxSourceEntry {
	if d == nil {
		return nil
	}
	if entry := d.files[name]; entry != nil {
		return entry
	}
	if canonical, ok := d.byKey[tspath.NormalizePath(name)]; ok {
		return d.files[canonical]
	}
	if canonical, ok := d.byKey[string(tspath.ToPath(name, d.currentDirectory, d.caseSensitive))]; ok {
		return d.files[canonical]
	}
	return nil
}

type Descriptor struct {
	SyntaxOnly        bool            `json:"syntaxOnly,omitempty,omitzero"`
	Generation        uint32          `json:"generation"`
	CurrentDirectory  string          `json:"currentDirectory"`
	CompilerOptions   json.RawMessage `json:"compilerOptions"`
	Sources           []Source        `json:"sources"`
	DiagnosticSources []Source        `json:"diagnosticSources,omitempty"`
}

type Source struct {
	ImpliedNodeFormat core.ResolutionMode `json:"impliedNodeFormat,omitempty,omitzero"`
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
	if syntaxOnly {
		options := json.RawMessage("{}")
		if opts := p.Options(); opts != nil {
			var err error
			options, err = json.Marshal(compilerOptions(opts))
			if err != nil {
				return nil, err
			}
		}
		currentDir := p.CurrentDirectory()
		caseSensitive := p.FS() != nil && p.FS().UseCaseSensitiveFileNames()
		data := &syntaxSessionData{
			currentDirectory: currentDir,
			compilerOptions:  options,
			caseSensitive:    caseSensitive,
			order:            make([]string, 0, len(eligible)),
			files:            make(map[string]*syntaxSourceEntry, len(eligible)),
			byKey:            make(map[string]string, len(eligible)*3),
		}
		for _, f := range eligible {
			if !p.OwnsSourceFile(f) {
				return nil, errors.New("ineligible or foreign type service source")
			}
			fileName := f.FileName()
			if _, exists := data.files[fileName]; exists {
				continue
			}
			entry := &syntaxSourceEntry{
				fileName:     fileName,
				path:         f.Path(),
				text:         f.Text(),
				parseOptions: f.ParseOptions(),
				scriptKind:   f.ScriptKind,
				desc:         syntaxSource(f),
			}
			data.order = append(data.order, fileName)
			data.files[fileName] = entry
			data.byKey[tspath.NormalizePath(fileName)] = fileName
			if pathKey := string(f.Path()); pathKey != "" {
				data.byKey[pathKey] = fileName
			}
			data.byKey[string(tspath.ToPath(fileName, currentDir, caseSensitive))] = fileName
		}
		slices.Sort(data.order)
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		return &Session{
			syntaxOnly: true,
			syntax:     data,
			generation: nextGeneration.Add(1),
		}, nil
	}
	s := &Session{
		program:    p,
		anchor:     eligible[0],
		eligible:   make(map[*ast.SourceFile]bool, len(eligible)),
		generation: nextGeneration.Add(1),
	}
	for _, f := range eligible {
		if !p.OwnsSourceFile(f) || !p.CanProvideTypeChecker(f) {
			return nil, errors.New("ineligible or foreign type service source")
		}
		s.eligible[f] = true
	}
	if q := p.CompilerQueries(); q != nil {
		s.recordExtraSource(q.ConfigSourceFile())
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
	if s.syntaxOnly {
		if entry := s.syntax.lookup(name); entry != nil {
			return entry.sourceFile()
		}
		return nil
	}
	if f := s.program.GetSourceFile(name); f != nil {
		return f
	}
	return s.extraSource(name)
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
	s.closed = true
	s.program = nil
	s.syntax = nil
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

func (s *Session) isClosed() bool {
	if s.closed {
		return true
	}
	if s.syntaxOnly {
		return s.syntax == nil
	}
	return s.program == nil
}

func (s *Session) BindTarget(file *ast.SourceFile) (Binding, error) {
	s.lifeMu.RLock()
	defer s.lifeMu.RUnlock()
	if s.isClosed() {
		return Binding{}, ErrClosed
	}
	if s.syntaxOnly {
		if file == nil {
			return Binding{}, errors.New("source is not eligible for typed rules")
		}
		entry := s.syntax.files[file.FileName()]
		if entry == nil || entry.path != file.Path() {
			return Binding{}, errors.New("source is not eligible for typed rules")
		}
		return Binding{Scope{Generation: s.generation}, entry.fileName}, nil
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
	if s.isClosed() {
		return nil, ErrClosed
	}
	if generation != s.generation {
		return nil, errors.New("stale type service generation")
	}
	file := s.sourceFile(name)
	if s.syntaxOnly && file == nil {
		return nil, errors.New("foreign syntax service source")
	}
	return file, nil //nolint:nilnil // A missing Program source is a legitimate result.
}

// SyntaxDescriptor opens only the requested target for a file-scoped reader.
func (s *Session) SyntaxDescriptor(file *ast.SourceFile) (Descriptor, error) {
	s.lifeMu.RLock()
	defer s.lifeMu.RUnlock()
	if s.isClosed() {
		return Descriptor{}, ErrClosed
	}
	if !s.syntaxOnly || file == nil {
		return Descriptor{}, errors.New("foreign syntax projection target")
	}
	entry := s.syntax.files[file.FileName()]
	if entry == nil || entry.path != file.Path() {
		return Descriptor{}, errors.New("foreign syntax projection target")
	}
	d := s.syntaxDescriptor()
	d.Sources = []Source{entry.desc}
	return d, nil
}

// SyntaxDescriptorForFile opens only the requested target file name for a
// syntax-only session without requiring the caller to retain an *ast.SourceFile.
func (s *Session) SyntaxDescriptorForFile(fileName string) (Descriptor, error) {
	s.lifeMu.RLock()
	defer s.lifeMu.RUnlock()
	if s.isClosed() {
		return Descriptor{}, ErrClosed
	}
	if !s.syntaxOnly {
		return Descriptor{}, errors.New("foreign syntax projection target")
	}
	entry := s.syntax.files[fileName]
	if entry == nil {
		return Descriptor{}, errors.New("foreign syntax projection target")
	}
	d := s.syntaxDescriptor()
	d.Sources = []Source{entry.desc}
	return d, nil
}

func (s *Session) syntaxDescriptor() Descriptor {
	return Descriptor{
		SyntaxOnly:       true,
		Generation:       s.generation,
		CurrentDirectory: s.syntax.currentDirectory,
		CompilerOptions:  s.syntax.compilerOptions,
	}
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
	if s.isClosed() {
		return Descriptor{}, ErrClosed
	}
	if s.syntaxOnly {
		d := s.syntaxDescriptor()
		d.Sources = make([]Source, 0, len(s.syntax.order))
		for _, fileName := range s.syntax.order {
			d.Sources = append(d.Sources, s.syntax.files[fileName].desc)
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

func isNodeRecordRequest(req Request) (*NodeRequest, bool) {
	if req.operation() != "nodeRecord" {
		return nil, false
	}
	r, ok := req.(*NodeRequest)
	return r, ok
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
	// Mark the session poisoned before releasing lifeMu so concurrent callers
	// synchronized with lifeMu cannot miss a panic during unwind.
	defer func() {
		if value := recover(); value != nil {
			s.poisoned.Store(true)
			panic(value)
		}
	}()
	if s.poisoned.Load() {
		return nil, ErrPoisoned
	}
	if s.isClosed() {
		return nil, ErrClosed
	}
	p := s.program
	for _, q := range queries {
		if q.Generation != s.generation {
			return nil, errors.New("stale or foreign type service handle")
		}
		if q.Request == nil {
			return nil, errors.New("missing type service request")
		}
		if s.syntaxOnly {
			if _, ok := isNodeRecordRequest(q.Request); !ok {
				return nil, errors.New("semantic queries require a typed project")
			}
		}
	}
	natives := make([]*ast.Node, len(queries))
	for i, q := range queries {
		if ref := requestNode(q.Request); ref != nil {
			f := s.sourceFile(ref.File)
			if s.syntaxOnly && f == nil {
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
		if s.poisoned.Load() {
			return nil, ErrPoisoned
		}
		if r, ok := queries[start].Request.(*ProgramRequest); ok {
			results[start], err = programQuery(ctx, s, p, r, natives[start])
			if err != nil {
				return nil, err
			}
			start++
			continue
		}
		if r, ok := isNodeRecordRequest(queries[start].Request); ok {
			results[start] = nodeFacts(s.sourceFile(r.Node.File), natives[start])
			start++
			continue
		}
		end := start + 1
		for end < len(queries) {
			if _, ok := queries[end].Request.(*ProgramRequest); ok {
				break
			}
			if _, ok := isNodeRecordRequest(queries[end].Request); ok {
				break
			}
			end++
		}
		if err = s.checkBatch(ctx, queries[start:end], natives[start:end], results[start:end], records[start:end]); err != nil {
			return nil, err
		}
		start = end
	}
	if s.poisoned.Load() {
		return nil, ErrPoisoned
	}
	packets = make([][]byte, len(queries))
	for i, result := range results {
		if s.poisoned.Load() {
			return nil, ErrPoisoned
		}
		if p != nil {
			materializeReferences(p, result)
			for _, record := range records[i] {
				materializeReferences(p, record.Fact)
			}
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
	if s.poisoned.Load() {
		return ErrPoisoned
	}
	// Mark the session poisoned while still holding checkMu so any goroutine
	// waiting on checkMu is happens-before guaranteed to observe ErrPoisoned
	// instead of touching a potentially inconsistent checker.
	defer func() {
		if value := recover(); value != nil {
			s.poisoned.Store(true)
			panic(value)
		}
	}()
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
		if s.poisoned.Load() {
			return ErrPoisoned
		}
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

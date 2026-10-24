package program

import (
	"context"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/module"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/utils/pathutil"
)

// Program is rslint's immutable facade over one complete source generation.
// Parser, filesystem, package, module-resolution, syntax, and optional type
// services are exposed through one contract. Which private adapter supplies
// them is deliberately not observable outside this package.
//
// Program pointer identity is generation identity. Reparse/rebuild operations
// create a new Program rather than replacing files or services in place, so a
// prepared lint plan can bind to one exact source world.
type Program struct {
	source sourceBackend
	types  typeBackend
	cache  *derivedCache
}

// sourceBackend is the private Adapter side of the facade. It intentionally
// contains only source-generation capabilities; lint selection, rule config,
// checker eligibility, type-check policy, diagnostics delivery, and fixes
// belong to the lint plan or its caller.
type sourceBackend interface {
	sourceFiles() []*ast.SourceFile
	rootFileNames() []string
	ownsSourceFile(file *ast.SourceFile) bool
	compilerOptions() *core.CompilerOptions
	fileSystem() vfs.FS
	workingDirectory() string
	defaultLibraryPath() string
	nearestPackageJSONDirectory(directory string) string
	fileExists(path string) bool
	modeForUsageLocation(sourceFile ast.HasFileName, location *ast.StringLiteralLike) core.ResolutionMode
	resolvedModule(sourceFile ast.HasFileName, moduleName string, mode core.ResolutionMode) *module.ResolvedModule
	resolveModuleName(moduleName string, containingFile string, mode core.ResolutionMode) *module.ResolvedModule
	sourceFileForResolvedModule(resolved *module.ResolvedModule) *ast.SourceFile
	sourceFile(fileName string) *ast.SourceFile
	sourceFileMetadata(file *ast.SourceFile) ast.SourceFileMetaData
	syntacticDiagnostics(ctx context.Context, file *ast.SourceFile) []*ast.Diagnostic
	isSourceFileDefaultLibrary(file *ast.SourceFile) bool
	isSourceFileFromExternalLibrary(file *ast.SourceFile) bool
	packageNamesBySourceFile() map[tspath.PathKey][]string
}

// typeBackend is an optional private capability adapter. Its absence is not a
// backend kind exposed to callers: the public facade simply cannot provide a
// checker or program-wide type diagnostics for that source generation.
type typeBackend interface {
	typeCheckerForFile(ctx context.Context, file *ast.SourceFile) (*checker.Checker, func())
	typeCheckerForFileExclusive(ctx context.Context, file *ast.SourceFile) (*checker.Checker, func())
	noEmitDiagnostics(ctx context.Context) []*ast.Diagnostic
}

type compilerBackend struct {
	raw *compiler.Program
}

// IsValid reports whether Program was created by this package. The zero value
// is invalid and is rejected at linter boundaries before callbacks run.
func (p *Program) IsValid() bool {
	return p != nil && p.source != nil && p.cache != nil
}

// NewFromCompiler adapts a ts-go Program into rslint's Program facade. The
// constructor name describes assembly only; consumers cannot recover or test
// which adapter backs the returned Program. Derived caches remain shared for
// the raw generation's lifetime, including gaps between temporary facades.
func NewFromCompiler(raw *compiler.Program) *Program {
	if raw == nil {
		return nil
	}
	backend := &compilerBackend{raw: raw}
	return &Program{source: backend, types: backend, cache: cacheForCompiler(raw)}
}

// NewFromCompilers adapts programs in stable input order.
func NewFromCompilers(rawPrograms []*compiler.Program) []*Program {
	programs := make([]*Program, len(rawPrograms))
	for index, raw := range rawPrograms {
		programs[index] = NewFromCompiler(raw)
	}
	return programs
}

// NewOwnedFromCompilers adapts freshly constructed compiler generations for a
// load pass that retains their complete facade set. The facades own derived
// caches directly; duplicate raw inputs share a cache within this call, while
// separate calls create independent caches. The caller must reuse these
// facades for this load pass rather than adapting their raw inputs again.
// Callers that reuse raw generations across facade lifetimes use NewFromCompiler.
func NewOwnedFromCompilers(rawPrograms []*compiler.Program) []*Program {
	programs := make([]*Program, len(rawPrograms))
	caches := make(map[*compiler.Program]*derivedCache, len(rawPrograms))
	for index, raw := range rawPrograms {
		if raw == nil {
			continue
		}
		cache := caches[raw]
		if cache == nil {
			cache = &derivedCache{values: make(map[any]*derivedValue)}
			caches[raw] = cache
		}
		backend := &compilerBackend{raw: raw}
		programs[index] = &Program{source: backend, types: backend, cache: cache}
	}
	return programs
}

// SourceFiles returns the complete immutable source universe in stable order.
// The returned slice and ASTs are read-only for the Program's lifetime.
func (p *Program) SourceFiles() []*ast.SourceFile {
	if !p.IsValid() {
		return nil
	}
	return p.source.sourceFiles()
}

// RootFileNames returns the source roots that define ownership. Dependencies
// present only because resolution loaded them are not roots.
func (p *Program) RootFileNames() []string {
	if !p.IsValid() {
		return nil
	}
	return p.source.rootFileNames()
}

// OwnsSourceFile reports whether file is the exact AST generation exposed by
// this Program.
func (p *Program) OwnsSourceFile(file *ast.SourceFile) bool {
	return p.IsValid() && file != nil && p.source.ownsSourceFile(file)
}

// ownsSourceReference accepts the small HasFileName contract used by ts-go's
// resolution helpers while preserving Program generation identity whenever
// the concrete value is a SourceFile. A same-path AST from another parse must
// never borrow this Program's cached module answers.
func (p *Program) ownsSourceReference(source ast.HasFileName) bool {
	if !p.IsValid() || source == nil {
		return false
	}
	if file, ok := source.(*ast.SourceFile); ok {
		return p.OwnsSourceFile(file)
	}
	file := p.GetSourceFile(source.FileName().AsString())
	return file != nil && file.PathKey() == source.PathKey()
}

func (p *Program) Options() *core.CompilerOptions {
	if !p.IsValid() {
		return nil
	}
	return p.source.compilerOptions()
}

func (p *Program) FS() vfs.FS {
	if !p.IsValid() {
		return nil
	}
	return p.source.fileSystem()
}

func (p *Program) CurrentDirectory() string {
	if !p.IsValid() {
		return ""
	}
	return p.source.workingDirectory()
}

func (p *Program) DefaultLibraryPath() string {
	if !p.IsValid() {
		return ""
	}
	return p.source.defaultLibraryPath()
}

func (p *Program) NearestPackageJSONDirectory(directory string) string {
	if !p.IsValid() {
		return ""
	}
	return p.source.nearestPackageJSONDirectory(directory)
}

func (p *Program) FileExists(path tspath.RootedFilePath) bool {
	return p.IsValid() && p.source.fileExists(path.AsString())
}

func (p *Program) GetModeForUsageLocation(sourceFile ast.HasFileName, location *ast.StringLiteralLike) core.ResolutionMode {
	if !p.ownsSourceReference(sourceFile) {
		return core.ResolutionModeNone
	}
	return p.source.modeForUsageLocation(sourceFile, location)
}

func (p *Program) GetResolvedModule(sourceFile ast.HasFileName, moduleName string, mode core.ResolutionMode) *module.ResolvedModule {
	if !p.ownsSourceReference(sourceFile) {
		return nil
	}
	return p.source.resolvedModule(sourceFile, moduleName, mode)
}

func (p *Program) GetResolvedModuleFromModuleSpecifier(sourceFile ast.HasFileName, specifier *ast.StringLiteralLike) *module.ResolvedModule {
	if !p.IsValid() || sourceFile == nil || specifier == nil {
		return nil
	}
	mode := p.GetModeForUsageLocation(sourceFile, specifier)
	return p.GetResolvedModule(sourceFile, specifier.Text(), mode)
}

func (p *Program) ResolveModuleName(moduleName string, containingFile string, mode core.ResolutionMode) *module.ResolvedModule {
	if !p.IsValid() {
		return nil
	}
	return p.source.resolveModuleName(moduleName, containingFile, mode)
}

func (p *Program) GetSourceFileForResolvedModule(resolved *module.ResolvedModule) *ast.SourceFile {
	if resolved == nil || !p.IsValid() {
		return nil
	}
	return p.source.sourceFileForResolvedModule(resolved)
}

func (p *Program) GetSourceFile(fileName string) *ast.SourceFile {
	if !p.IsValid() {
		return nil
	}
	return p.source.sourceFile(fileName)
}

func (p *Program) SourceFileMetadata(file *ast.SourceFile) ast.SourceFileMetaData {
	if !p.OwnsSourceFile(file) {
		return ast.SourceFileMetaData{}
	}
	return p.source.sourceFileMetadata(file)
}

// SyntacticDiagnostics returns the complete syntax diagnostics for file under
// the same options and source host that produced its AST.
func (p *Program) SyntacticDiagnostics(ctx context.Context, file *ast.SourceFile) []*ast.Diagnostic {
	if !p.OwnsSourceFile(file) {
		return nil
	}
	return p.source.syntacticDiagnostics(ctx, file)
}

// CanProvideTypeChecker reports a source capability, not lint eligibility.
// The lint plan independently decides whether one file may run a type-aware
// rule; callers must never infer policy from this method.
func (p *Program) CanProvideTypeChecker(file *ast.SourceFile) bool {
	return p.IsValid() && p.types != nil && p.OwnsSourceFile(file)
}

// CanProvideProgramDiagnostics reports whether this source generation can run
// a complete tsc-style no-emit diagnostic pass. Callers use this capability to
// coordinate diagnostic ownership; it does not reveal how the Program was
// constructed and does not grant per-file rule eligibility.
func (p *Program) CanProvideProgramDiagnostics() bool {
	return p.IsValid() && p.types != nil
}

// TypeCheckerForFile acquires a checker when the source generation can provide
// one. The returned release function is always safe to call.
func (p *Program) TypeCheckerForFile(ctx context.Context, file *ast.SourceFile) (*checker.Checker, func()) {
	if !p.CanProvideTypeChecker(file) {
		return nil, func() {}
	}
	return p.types.typeCheckerForFile(ctx, file)
}

func (p *Program) TypeCheckerForFileExclusive(ctx context.Context, file *ast.SourceFile) (*checker.Checker, func()) {
	if !p.CanProvideTypeChecker(file) {
		return nil, func() {}
	}
	return p.types.typeCheckerForFileExclusive(ctx, file)
}

// NoEmitDiagnostics returns program-wide type diagnostics when this source
// generation supports them. An unavailable capability produces no diagnostics.
func (p *Program) NoEmitDiagnostics(ctx context.Context) []*ast.Diagnostic {
	if !p.IsValid() || p.types == nil {
		return nil
	}
	return p.types.noEmitDiagnostics(ctx)
}

// IsSourceFileDefaultLibrary answers against the source generation's compiler
// options and host without exposing its private adapter.
func (p *Program) IsSourceFileDefaultLibrary(file *ast.SourceFile) bool {
	if !p.OwnsSourceFile(file) || !file.IsDeclarationFile {
		return false
	}
	if p.source.isSourceFileDefaultLibrary(file) {
		return true
	}
	options := p.Options()
	if options == nil || options.NoLib.IsTrue() {
		return false
	}
	var libraries []string
	if options.Lib == nil {
		libraries = append(libraries, tspath.CombinePaths(p.DefaultLibraryPath(), tsoptions.GetDefaultLibFileName(options)))
	} else {
		for _, library := range options.Lib {
			if name, ok := tsoptions.GetLibFileName(library); ok {
				libraries = append(libraries, tspath.CombinePaths(p.DefaultLibraryPath(), name))
			}
		}
	}
	fs := p.FS()
	if fs == nil {
		return false
	}
	for _, library := range libraries {
		if pathutil.ComparePaths(file.FileName().AsString(), library, pathutil.ComparePathsOptions{
			CurrentDirectory:          p.CurrentDirectory(),
			UseCaseSensitiveFileNames: fs.CaseSensitivity().IsCaseSensitive(),
		}) == 0 {
			return true
		}
	}
	return false
}

func (p *Program) IsSourceFileFromExternalLibrary(file *ast.SourceFile) bool {
	return p.OwnsSourceFile(file) && p.source.isSourceFileFromExternalLibrary(file)
}

type packageNamesCacheKey struct{}

// PackageNamesForSourceFile returns package identities reconstructed from the
// source generation's module and type-reference resolution caches.
func (p *Program) PackageNamesForSourceFile(file *ast.SourceFile) []string {
	if !p.OwnsSourceFile(file) {
		return nil
	}
	names := Cached(p, packageNamesCacheKey{}, p.source.packageNamesBySourceFile)
	return names[file.PathKey()]
}

func packageNamesFromResolutions(
	getSourceFile func(tspath.RootedFilePath, tspath.PathKey) *ast.SourceFile,
	moduleResolutions map[tspath.PathKey]module.ModeAwareCache[*module.ResolvedModule],
	typeResolutions map[tspath.PathKey]module.ModeAwareCache[*module.ResolvedTypeReferenceDirective],
) map[tspath.PathKey][]string {
	names := make(map[tspath.PathKey][]string)
	add := func(fileName tspath.RootedFilePath, path tspath.PathKey, packageID module.PackageId) {
		if packageID.Name == "" {
			return
		}
		file := getSourceFile(fileName, path)
		if file == nil {
			return
		}
		name := packageID.PackageName()
		if !slices.Contains(names[file.PathKey()], name) {
			names[file.PathKey()] = append(names[file.PathKey()], name)
		}
	}
	for _, resolutions := range moduleResolutions {
		for _, resolution := range resolutions {
			if resolution != nil {
				add(resolution.ResolvedFileName, resolution.ResolvedPath, resolution.PackageId)
			}
		}
	}
	for _, resolutions := range typeResolutions {
		for _, resolution := range resolutions {
			if resolution != nil {
				add(resolution.ResolvedFileName, resolution.ResolvedPath, resolution.PackageId)
			}
		}
	}
	return names
}

func (b *compilerBackend) sourceFiles() []*ast.SourceFile { return b.raw.SourceFiles() }
func (b *compilerBackend) rootFileNames() []string {
	return pathutil.Strings(b.raw.CommandLine().FileNames())
}
func (b *compilerBackend) ownsSourceFile(file *ast.SourceFile) bool {
	return b.raw.GetSourceFile(file.FileName()) == file
}
func (b *compilerBackend) compilerOptions() *core.CompilerOptions { return b.raw.Options() }
func (b *compilerBackend) fileSystem() vfs.FS                     { return b.raw.Host().FS() }
func (b *compilerBackend) workingDirectory() string               { return b.raw.GetCurrentDirectory().AsString() }
func (b *compilerBackend) defaultLibraryPath() string {
	return b.raw.Host().DefaultLibraryPath().AsString()
}
func (b *compilerBackend) nearestPackageJSONDirectory(directory string) string {
	return b.raw.GetNearestAncestorDirectoryWithPackageJson(pathutil.Directory(directory)).AsString()
}
func (b *compilerBackend) fileExists(path string) bool { return b.raw.FileExists(pathutil.File(path)) }
func (b *compilerBackend) modeForUsageLocation(file ast.HasFileName, location *ast.StringLiteralLike) core.ResolutionMode {
	return b.raw.GetModeForUsageLocation(file, location)
}
func (b *compilerBackend) resolvedModule(file ast.HasFileName, name string, mode core.ResolutionMode) *module.ResolvedModule {
	return b.raw.GetResolvedModule(file, name, mode)
}
func (b *compilerBackend) resolveModuleName(name string, containingFile string, mode core.ResolutionMode) *module.ResolvedModule {
	resolved, _, _ := compiler.Program_newResolver(b.raw).ResolveModuleName(name, pathutil.File(containingFile), mode, nil)
	return resolved
}
func (b *compilerBackend) sourceFileForResolvedModule(resolved *module.ResolvedModule) *ast.SourceFile {
	return b.raw.GetSourceFileForResolvedModule(resolved)
}
func (b *compilerBackend) sourceFile(fileName string) *ast.SourceFile {
	return b.raw.GetSourceFile(tspath.ToRootedFilePath(fileName, b.raw.GetCurrentDirectory()))
}
func (b *compilerBackend) sourceFileMetadata(file *ast.SourceFile) ast.SourceFileMetaData {
	return b.raw.GetSourceFileMetaData(file.PathKey())
}
func (b *compilerBackend) syntacticDiagnostics(ctx context.Context, file *ast.SourceFile) []*ast.Diagnostic {
	return b.raw.GetSyntacticDiagnostics(ctx, file)
}
func (b *compilerBackend) isSourceFileDefaultLibrary(file *ast.SourceFile) bool {
	return b.raw.IsSourceFileDefaultLibrary(file.PathKey())
}
func (b *compilerBackend) isSourceFileFromExternalLibrary(file *ast.SourceFile) bool {
	return b.raw.IsSourceFileFromExternalLibrary(file)
}
func (b *compilerBackend) packageNamesBySourceFile() map[tspath.PathKey][]string {
	return packageNamesFromResolutions(
		func(fileName tspath.RootedFilePath, path tspath.PathKey) *ast.SourceFile {
			return b.raw.GetSourceFileForResolvedModule(&module.ResolvedModule{ResolvedFileName: fileName, ResolvedPath: path})
		},
		b.raw.GetResolvedModules(),
		b.raw.GetResolvedTypeReferenceDirectives(),
	)
}
func (b *compilerBackend) typeCheckerForFile(ctx context.Context, file *ast.SourceFile) (*checker.Checker, func()) {
	return b.raw.GetTypeCheckerForFile(ctx, file)
}
func (b *compilerBackend) typeCheckerForFileExclusive(ctx context.Context, file *ast.SourceFile) (*checker.Checker, func()) {
	return b.raw.GetTypeCheckerForFileExclusive(ctx, file)
}
func (b *compilerBackend) noEmitDiagnostics(ctx context.Context) []*ast.Diagnostic {
	return collectNoEmitDiagnostics(ctx, b.raw)
}

func (s *parsedBackend) sourceFiles() []*ast.SourceFile { return s.files }
func (s *parsedBackend) rootFileNames() []string {
	roots := make([]string, len(s.files))
	for index, file := range s.files {
		roots[index] = file.FileName().AsString()
	}
	return roots
}
func (s *parsedBackend) ownsSourceFile(file *ast.SourceFile) bool {
	return s.sourcesByPath[file.PathKey()] == file
}
func (s *parsedBackend) compilerOptions() *core.CompilerOptions { return s.options }
func (s *parsedBackend) fileSystem() vfs.FS                     { return s.fs }
func (s *parsedBackend) workingDirectory() string               { return s.currentDirectory }
func (s *parsedBackend) defaultLibraryPath() string             { return s.host.DefaultLibraryPath().AsString() }
func (s *parsedBackend) nearestPackageJSONDirectory(directory string) string {
	scope := s.resolver.GetPackageScopeForPath(pathutil.Directory(directory))
	if scope.Exists() {
		return scope.PackageDirectory.AsDirectoryPath().AsString()
	}
	return ""
}
func (s *parsedBackend) fileExists(path string) bool {
	return s.fs != nil && s.fs.FileExists(pathutil.File(path))
}
func (s *parsedBackend) modeForUsageLocation(file ast.HasFileName, location *ast.StringLiteralLike) core.ResolutionMode {
	if file == nil {
		return core.ResolutionModeNone
	}
	return compiler.GetModeForUsageLocation(file.FileName(), s.metadataByPath[file.PathKey()], location, s.options)
}
func (s *parsedBackend) resolvedModule(file ast.HasFileName, name string, mode core.ResolutionMode) *module.ResolvedModule {
	if file == nil {
		return nil
	}
	return s.resolvedModules[file.PathKey()][module.ModeAwareCacheKey{Name: name, Mode: mode}]
}
func (s *parsedBackend) resolveModuleName(name string, containingFile string, mode core.ResolutionMode) *module.ResolvedModule {
	resolved, _, _ := s.resolver.ResolveModuleName(name, pathutil.File(containingFile), mode, nil)
	return resolved
}
func (s *parsedBackend) sourceFileForResolvedModule(resolved *module.ResolvedModule) *ast.SourceFile {
	return s.sourceFile(resolved.ResolvedFileName.AsString())
}
func (s *parsedBackend) sourceFileMetadata(file *ast.SourceFile) ast.SourceFileMetaData {
	return s.metadataByPath[file.PathKey()]
}
func (s *parsedBackend) syntacticDiagnostics(_ context.Context, file *ast.SourceFile) []*ast.Diagnostic {
	return s.syntacticDiagnosticsByPath[file.PathKey()]
}
func (s *parsedBackend) isSourceFileDefaultLibrary(*ast.SourceFile) bool      { return false }
func (s *parsedBackend) isSourceFileFromExternalLibrary(*ast.SourceFile) bool { return false }
func (s *parsedBackend) packageNamesBySourceFile() map[tspath.PathKey][]string {
	return packageNamesFromResolutions(func(fileName tspath.RootedFilePath, _ tspath.PathKey) *ast.SourceFile {
		return s.sourceFile(fileName.AsString())
	}, s.resolvedModules, nil)
}

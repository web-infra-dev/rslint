package program

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/binder"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/module"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
)

// RootOptions describes one parser/binder-built source universe.
// RootFileNames are parsed exactly; imports may resolve to paths, but the
// source Program does not recursively materialize an import closure.
type RootOptions struct {
	RootFileNames   []string
	Host            compiler.CompilerHost
	CompilerOptions *core.CompilerOptions
	SingleThreaded  bool
}

type parsedBackend struct {
	host                       compiler.CompilerHost
	fs                         vfs.FS
	currentDirectory           string
	options                    *core.CompilerOptions
	resolver                   *module.Resolver
	files                      []*ast.SourceFile
	metadataByPath             map[tspath.Path]ast.SourceFileMetaData
	sourcesByPath              map[tspath.Path]*ast.SourceFile
	resolvedModules            map[tspath.Path]module.ModeAwareCache[*module.ResolvedModule]
	syntacticDiagnosticsByPath map[tspath.Path][]*ast.Diagnostic
}

type parsedResult struct {
	file                 *ast.SourceFile
	metadata             ast.SourceFileMetaData
	resolvedModules      module.ModeAwareCache[*module.ResolvedModule]
	syntacticDiagnostics []*ast.Diagnostic
	err                  error
}

// NewFromRoots parses, resolves direct imports for, and binds one immutable
// source universe. The returned Program owns its source-file slice
// and AST generation; callers must build a new Program after source changes.
func NewFromRoots(opts RootOptions) (*Program, error) {
	if isNilInterface(opts.Host) {
		return nil, errors.New("program: root construction requires a compiler host")
	}
	if opts.CompilerOptions == nil {
		return nil, errors.New("program: root construction requires compiler options")
	}
	fs := opts.Host.FS()
	if isNilInterface(fs) {
		return nil, errors.New("program: root construction requires a filesystem")
	}
	currentDirectory := opts.Host.GetCurrentDirectory()
	rootFileNames, err := normalizeRootFileNames(opts.RootFileNames, currentDirectory, fs)
	if err != nil {
		return nil, err
	}

	backend := &parsedBackend{
		host:             opts.Host,
		fs:               fs,
		currentDirectory: currentDirectory,
		options:          opts.CompilerOptions,
		resolver:         module.NewResolver(opts.Host, opts.CompilerOptions, "", ""),
		metadataByPath:   make(map[tspath.Path]ast.SourceFileMetaData, len(rootFileNames)),
		sourcesByPath:    make(map[tspath.Path]*ast.SourceFile, len(rootFileNames)),
		resolvedModules:  make(map[tspath.Path]module.ModeAwareCache[*module.ResolvedModule], len(rootFileNames)),
	}
	results := make([]parsedResult, len(rootFileNames))

	parse := func(index int) {
		rootFileName := rootFileNames[index]
		file, metadata := backend.parse(rootFileName)
		if file == nil {
			results[index].err = fmt.Errorf("program: could not read root %q", rootFileName)
			return
		}
		syntacticDiagnostics := backend.computeSyntacticDiagnostics(file)
		resolvedModules := backend.resolveImports(file, metadata)
		binder.BindSourceFile(file)
		results[index] = parsedResult{
			file:                 file,
			metadata:             metadata,
			resolvedModules:      resolvedModules,
			syntacticDiagnostics: syntacticDiagnostics,
		}
	}

	workerCount := min(runtime.GOMAXPROCS(0), len(results))
	if opts.SingleThreaded || workerCount < 2 {
		for index := range results {
			parse(index)
		}
	} else {
		chunkSize := (len(results) + workerCount - 1) / workerCount
		work := core.NewWorkGroup(false)
		for worker := range workerCount {
			start := worker * chunkSize
			end := min(start+chunkSize, len(results))
			if start >= end {
				continue
			}
			work.Queue(func() {
				for index := start; index < end; index++ {
					parse(index)
				}
			})
		}
		work.RunAndWait()
	}

	backend.files = make([]*ast.SourceFile, 0, len(results))
	for _, result := range results {
		if result.err != nil {
			return nil, result.err
		}
		if previous := backend.sourcesByPath[result.file.Path()]; previous != nil {
			if previous != result.file {
				return nil, fmt.Errorf(
					"program: source universe contains different ASTs for path %q",
					result.file.Path(),
				)
			}
			continue
		}
		backend.install(result.file, result.metadata, result.resolvedModules, result.syntacticDiagnostics)
		backend.files = append(backend.files, result.file)
	}
	return &Program{source: backend, cache: &derivedCache{}}, nil
}

func normalizeRootFileNames(rootFileNames []string, currentDirectory string, fs vfs.FS) ([]string, error) {
	normalized := make([]string, 0, len(rootFileNames))
	seen := make(map[tspath.Path]string, len(rootFileNames))
	useCaseSensitive := fs.UseCaseSensitiveFileNames()
	for _, rootFileName := range rootFileNames {
		absolute := tspath.GetNormalizedAbsolutePath(rootFileName, currentDirectory)
		path := tspath.ToPath(absolute, currentDirectory, useCaseSensitive)
		if previous, duplicate := seen[path]; duplicate {
			if previous != absolute {
				return nil, fmt.Errorf(
					"program: roots %q and %q have the same path identity %q",
					previous,
					absolute,
					path,
				)
			}
			continue
		}
		seen[path] = absolute
		normalized = append(normalized, absolute)
	}
	return normalized, nil
}

func isNilInterface(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}

// NewFromBoundSources creates a source-only rslint Program from an already-
// bound subset of one ts-go Program. The ts-go Program supplies
// non-type source services and ownership validation, but is deliberately not
// exposed through the returned facade: files never gain type-aware rule
// eligibility merely because construction reused source services from ts-go.
func NewFromBoundSources(
	typeScript *compiler.Program,
	files []*ast.SourceFile,
) (*Program, error) {
	if typeScript == nil {
		return nil, errors.New("program: bound-source construction requires source services")
	}
	options := typeScript.Options()
	if options == nil {
		return nil, errors.New("program: bound-source construction requires compiler options")
	}
	host := typeScript.Host()
	if isNilInterface(host) {
		return nil, errors.New("program: bound-source construction requires a compiler host")
	}
	fs := host.FS()
	if isNilInterface(fs) {
		return nil, errors.New("program: bound-source construction requires a filesystem")
	}
	currentDirectory := host.GetCurrentDirectory()

	backend := &parsedBackend{
		host:             host,
		fs:               fs,
		currentDirectory: currentDirectory,
		options:          options,
		resolver:         module.NewResolver(host, options, "", ""),
		metadataByPath:   make(map[tspath.Path]ast.SourceFileMetaData, len(files)),
		sourcesByPath:    make(map[tspath.Path]*ast.SourceFile, len(files)),
		resolvedModules:  make(map[tspath.Path]module.ModeAwareCache[*module.ResolvedModule], len(files)),
	}
	normalized, err := normalizeSourceFiles(files)
	if err != nil {
		return nil, err
	}
	for _, file := range normalized {
		if !file.IsBound() {
			return nil, fmt.Errorf("program: source %q is not bound", file.FileName())
		}
		if typeScript.GetSourceFile(file.FileName()) != file {
			return nil, fmt.Errorf("program: source services do not own source %q", file.FileName())
		}
		metadata := backend.sourceFileMetaData(file.FileName())
		backend.install(
			file,
			metadata,
			backend.resolveImports(file, metadata),
			backend.computeSyntacticDiagnostics(file),
		)
	}
	backend.files = normalized
	return &Program{source: backend, cache: &derivedCache{}}, nil
}

func normalizeSourceFiles(files []*ast.SourceFile) ([]*ast.SourceFile, error) {
	normalized := make([]*ast.SourceFile, 0, len(files))
	seenPaths := make(map[tspath.Path]*ast.SourceFile, len(files))
	for _, file := range files {
		if file == nil {
			continue
		}
		path := file.Path()
		if path == "" {
			return nil, fmt.Errorf("program: source %q has no Path", file.FileName())
		}
		if previous, duplicate := seenPaths[path]; duplicate {
			if previous != file {
				return nil, fmt.Errorf("program: source universe contains different ASTs for path %q", path)
			}
			continue
		}
		seenPaths[path] = file
		normalized = append(normalized, file)
	}
	return normalized, nil
}

func (p *parsedBackend) sourceFileMetaData(fileName string) ast.SourceFileMetaData {
	packageJSONScope := p.resolver.GetPackageScopeForPath(tspath.GetDirectoryPath(fileName))
	moduleResolutionKind := p.options.GetModuleResolutionKind()

	var packageJSONType, packageJSONDirectory string
	if packageJSONScope.Exists() {
		packageJSONDirectory = packageJSONScope.PackageDirectory
		if value, ok := packageJSONScope.Contents.Type.GetValue(); ok {
			hasExplicitFormat := tspath.FileExtensionIsOneOf(fileName, []string{
				tspath.ExtensionMts,
				tspath.ExtensionCts,
				tspath.ExtensionMjs,
				tspath.ExtensionCjs,
			})
			usesNodeFormat := core.ModuleResolutionKindNode16 <= moduleResolutionKind &&
				moduleResolutionKind <= core.ModuleResolutionKindNodeNext
			if (!hasExplicitFormat && usesNodeFormat) || strings.Contains(fileName, "/node_modules/") {
				packageJSONType = value
			}
		}
	}

	return ast.SourceFileMetaData{
		PackageJsonType:      packageJSONType,
		PackageJsonDirectory: packageJSONDirectory,
		ImpliedNodeFormat:    ast.GetImpliedNodeFormatForFile(fileName, packageJSONType),
	}
}

func (p *parsedBackend) parse(rootFileName string) (*ast.SourceFile, ast.SourceFileMetaData) {
	fileName := tspath.GetNormalizedAbsolutePath(rootFileName, p.currentDirectory)
	metadata := p.sourceFileMetaData(fileName)
	file := p.host.GetSourceFile(ast.SourceFileParseOptions{
		FileName: fileName,
		Path: tspath.ToPath(
			fileName,
			p.currentDirectory,
			p.fs.UseCaseSensitiveFileNames(),
		),
		ExternalModuleIndicatorOptions: ast.GetExternalModuleIndicatorOptions(fileName, p.options, metadata),
	})
	return file, metadata
}

func (p *parsedBackend) resolveImports(
	file *ast.SourceFile,
	metadata ast.SourceFileMetaData,
) module.ModeAwareCache[*module.ResolvedModule] {
	moduleNames := append([]*ast.Node(nil), file.Imports()...)
	for _, augmentation := range file.ModuleAugmentations {
		if augmentation.Kind == ast.KindStringLiteral {
			moduleNames = append(moduleNames, augmentation)
		}
	}
	if len(moduleNames) == 0 {
		return nil
	}

	resolutions := make(module.ModeAwareCache[*module.ResolvedModule], len(moduleNames))
	for _, moduleName := range moduleNames {
		name := moduleName.Text()
		if name == "" {
			continue
		}
		mode := compiler.GetModeForUsageLocation(file.FileName(), metadata, moduleName, p.options)
		resolved, _ := p.resolver.ResolveModuleName(name, file.FileName(), mode, nil)
		resolutions[module.ModeAwareCacheKey{Name: name, Mode: mode}] = resolved
	}
	return resolutions
}

func (p *parsedBackend) install(
	file *ast.SourceFile,
	metadata ast.SourceFileMetaData,
	resolvedModules module.ModeAwareCache[*module.ResolvedModule],
	syntacticDiagnostics []*ast.Diagnostic,
) {
	p.metadataByPath[file.Path()] = metadata
	p.sourcesByPath[file.Path()] = file
	if resolvedModules != nil {
		p.resolvedModules[file.Path()] = resolvedModules
	}
	if len(syntacticDiagnostics) > 0 {
		if p.syntacticDiagnosticsByPath == nil {
			p.syntacticDiagnosticsByPath = make(map[tspath.Path][]*ast.Diagnostic)
		}
		p.syntacticDiagnosticsByPath[file.Path()] = syntacticDiagnostics
	}
}

func (p *parsedBackend) computeSyntacticDiagnostics(file *ast.SourceFile) []*ast.Diagnostic {
	diagnostics := append([]*ast.Diagnostic(nil), file.Diagnostics()...)
	diagnostics = append(diagnostics, file.JSDiagnostics()...)
	if ast.IsSourceFileJS(file) && !ast.IsCheckJSEnabledForFile(file, p.options) {
		diagnostics = append(diagnostics, compiler.GetAdditionalJSSyntacticDiagnostics(file, p.options)...)
	}
	return compiler.SortAndDeduplicateDiagnostics(diagnostics)
}

func (p *parsedBackend) sourceFile(fileName string) *ast.SourceFile {
	path := tspath.ToPath(fileName, p.currentDirectory, p.fs.UseCaseSensitiveFileNames())
	return p.sourcesByPath[path]
}

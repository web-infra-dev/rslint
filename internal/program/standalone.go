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

// StandaloneOptions describes one parser/binder-backed source universe.
// RootFileNames are parsed exactly; imports may resolve to paths, but the
// standalone Program does not recursively materialize an import closure.
type StandaloneOptions struct {
	RootFileNames   []string
	Host            compiler.CompilerHost
	CompilerOptions *core.CompilerOptions
	SingleThreaded  bool
}

type standaloneProgram struct {
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

type standaloneParseResult struct {
	file                 *ast.SourceFile
	metadata             ast.SourceFileMetaData
	resolvedModules      module.ModeAwareCache[*module.ResolvedModule]
	syntacticDiagnostics []*ast.Diagnostic
	err                  error
}

// NewStandalone parses, resolves direct imports for, and binds one immutable
// standalone source universe. The returned Program owns its source-file slice
// and AST generation; callers must build a new Program after source changes.
func NewStandalone(opts StandaloneOptions) (*Program, error) {
	if isNilInterface(opts.Host) {
		return nil, errors.New("program: standalone Program requires a compiler host")
	}
	if opts.CompilerOptions == nil {
		return nil, errors.New("program: standalone Program requires compiler options")
	}
	fs := opts.Host.FS()
	if isNilInterface(fs) {
		return nil, errors.New("program: standalone Program requires a filesystem")
	}
	currentDirectory := opts.Host.GetCurrentDirectory()
	rootFileNames, err := normalizeStandaloneRootFileNames(opts.RootFileNames, currentDirectory, fs)
	if err != nil {
		return nil, err
	}

	standalone := &standaloneProgram{
		host:             opts.Host,
		fs:               fs,
		currentDirectory: currentDirectory,
		options:          opts.CompilerOptions,
		resolver:         module.NewResolver(opts.Host, opts.CompilerOptions, "", ""),
		metadataByPath:   make(map[tspath.Path]ast.SourceFileMetaData, len(rootFileNames)),
		sourcesByPath:    make(map[tspath.Path]*ast.SourceFile, len(rootFileNames)),
		resolvedModules:  make(map[tspath.Path]module.ModeAwareCache[*module.ResolvedModule], len(rootFileNames)),
	}
	results := make([]standaloneParseResult, len(rootFileNames))

	parse := func(index int) {
		rootFileName := rootFileNames[index]
		file, metadata := standalone.parse(rootFileName)
		if file == nil {
			results[index].err = fmt.Errorf("program: standalone Program could not read root %q", rootFileName)
			return
		}
		syntacticDiagnostics := standalone.computeSyntacticDiagnostics(file)
		resolvedModules := standalone.resolveImports(file, metadata)
		binder.BindSourceFile(file)
		results[index] = standaloneParseResult{
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

	standalone.files = make([]*ast.SourceFile, 0, len(results))
	for _, result := range results {
		if result.err != nil {
			return nil, result.err
		}
		if previous := standalone.sourcesByPath[result.file.Path()]; previous != nil {
			if previous != result.file {
				return nil, fmt.Errorf(
					"program: standalone Program contains different ASTs for path %q",
					result.file.Path(),
				)
			}
			continue
		}
		standalone.install(result.file, result.metadata, result.resolvedModules, result.syntacticDiagnostics)
		standalone.files = append(standalone.files, result.file)
	}
	return &Program{standalone: standalone}, nil
}

func normalizeStandaloneRootFileNames(rootFileNames []string, currentDirectory string, fs vfs.FS) ([]string, error) {
	normalized := make([]string, 0, len(rootFileNames))
	seen := make(map[tspath.Path]string, len(rootFileNames))
	useCaseSensitive := fs.UseCaseSensitiveFileNames()
	for _, rootFileName := range rootFileNames {
		absolute := tspath.GetNormalizedAbsolutePath(rootFileName, currentDirectory)
		path := tspath.ToPath(absolute, currentDirectory, useCaseSensitive)
		if previous, duplicate := seen[path]; duplicate {
			if previous != absolute {
				return nil, fmt.Errorf(
					"program: standalone roots %q and %q have the same path identity %q",
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

// NewStandaloneFromTypeScriptSources creates a standalone rslint Program from
// an already-bound subset of one ts-go Program. The ts-go Program supplies
// non-type source services and ownership validation, but is deliberately not
// exposed through TypeScriptProgram: files in the returned Program never gain
// type-aware rule eligibility merely because those services came from ts-go.
func NewStandaloneFromTypeScriptSources(
	typeScript *compiler.Program,
	files []*ast.SourceFile,
) (*Program, error) {
	if typeScript == nil {
		return nil, errors.New("program: standalone Program requires source services")
	}
	options := typeScript.Options()
	if options == nil {
		return nil, errors.New("program: standalone Program requires compiler options")
	}
	host := typeScript.Host()
	if isNilInterface(host) {
		return nil, errors.New("program: standalone Program requires a compiler host")
	}
	fs := host.FS()
	if isNilInterface(fs) {
		return nil, errors.New("program: standalone Program requires a filesystem")
	}
	currentDirectory := host.GetCurrentDirectory()

	standalone := &standaloneProgram{
		host:             host,
		fs:               fs,
		currentDirectory: currentDirectory,
		options:          options,
		resolver:         module.NewResolver(host, options, "", ""),
		metadataByPath:   make(map[tspath.Path]ast.SourceFileMetaData, len(files)),
		sourcesByPath:    make(map[tspath.Path]*ast.SourceFile, len(files)),
		resolvedModules:  make(map[tspath.Path]module.ModeAwareCache[*module.ResolvedModule], len(files)),
	}
	normalized, err := normalizeStandaloneSourceFiles(files)
	if err != nil {
		return nil, err
	}
	for _, file := range normalized {
		if !file.IsBound() {
			return nil, fmt.Errorf("program: standalone source %q is not bound", file.FileName())
		}
		if typeScript.GetSourceFile(file.FileName()) != file {
			return nil, fmt.Errorf("program: source services do not own standalone source %q", file.FileName())
		}
		metadata := standalone.sourceFileMetaData(file.FileName())
		standalone.install(
			file,
			metadata,
			standalone.resolveImports(file, metadata),
			standalone.computeSyntacticDiagnostics(file),
		)
	}
	standalone.files = normalized
	return &Program{standalone: standalone}, nil
}

func normalizeStandaloneSourceFiles(files []*ast.SourceFile) ([]*ast.SourceFile, error) {
	normalized := make([]*ast.SourceFile, 0, len(files))
	seenPaths := make(map[tspath.Path]*ast.SourceFile, len(files))
	for _, file := range files {
		if file == nil {
			continue
		}
		path := file.Path()
		if path == "" {
			return nil, fmt.Errorf("program: standalone source %q has no Path", file.FileName())
		}
		if previous, duplicate := seenPaths[path]; duplicate {
			if previous != file {
				return nil, fmt.Errorf("program: standalone Program contains different ASTs for path %q", path)
			}
			continue
		}
		seenPaths[path] = file
		normalized = append(normalized, file)
	}
	return normalized, nil
}

func (p *standaloneProgram) sourceFileMetaData(fileName string) ast.SourceFileMetaData {
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

func (p *standaloneProgram) parse(rootFileName string) (*ast.SourceFile, ast.SourceFileMetaData) {
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

func (p *standaloneProgram) resolveImports(
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

func (p *standaloneProgram) install(
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

func (p *standaloneProgram) computeSyntacticDiagnostics(file *ast.SourceFile) []*ast.Diagnostic {
	diagnostics := append([]*ast.Diagnostic(nil), file.Diagnostics()...)
	diagnostics = append(diagnostics, file.JSDiagnostics()...)
	if ast.IsSourceFileJS(file) && !ast.IsCheckJSEnabledForFile(file, p.options) {
		diagnostics = append(diagnostics, compiler.GetAdditionalJSSyntacticDiagnostics(file, p.options)...)
	}
	return compiler.SortAndDeduplicateDiagnostics(diagnostics)
}

func (p *standaloneProgram) sourceFile(fileName string) *ast.SourceFile {
	path := tspath.ToPath(fileName, p.currentDirectory, p.fs.UseCaseSensitiveFileNames())
	return p.sourcesByPath[path]
}

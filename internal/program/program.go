package program

import (
	"context"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/module"
	"github.com/microsoft/typescript-go/shim/vfs"
)

// Program is rslint's immutable view of one source universe. A Program is
// backed either by a TypeScript Program or by a parser/binder-backed
// standalone source universe. Callers use the same source and module-resolution
// services in both cases; TypeScriptProgram remains optional and is exposed
// only for operations that genuinely require ts-go's Program machinery.
//
// Program identity is generation identity. Reparse/rebuild operations create
// a new *Program instead of replacing the backend or source files in place, so
// prepared lint plans can bind to one exact pointer.
type Program struct {
	typeScript *compiler.Program
	standalone *standaloneProgram
}

// IsValid reports whether Program was created by one of this package's
// constructors and has exactly one source backend. The zero value is invalid.
func (p *Program) IsValid() bool {
	return p != nil && (p.typeScript != nil) != (p.standalone != nil)
}

// NewTypeScript wraps one ts-go Program source universe. The wrapper does not
// decide whether a lint target has type information or whether the project
// participates in program-wide type checking; those are run-plan policies.
func NewTypeScript(typeScript *compiler.Program) *Program {
	if typeScript == nil {
		return nil
	}
	return &Program{typeScript: typeScript}
}

// WrapTypeScriptPrograms wraps programs in stable input order.
func WrapTypeScriptPrograms(typeScriptPrograms []*compiler.Program) []*Program {
	programs := make([]*Program, len(typeScriptPrograms))
	for i, typeScript := range typeScriptPrograms {
		programs[i] = NewTypeScript(typeScript)
	}
	return programs
}

// TypeScriptProgram returns the underlying ts-go Program when this source
// universe is compiler-backed. Compatibility Programs return their ts-go
// Program here too; callers must not infer per-file type information or
// Phase-2 eligibility from this method.
func (p *Program) TypeScriptProgram() *compiler.Program {
	if p == nil {
		return nil
	}
	return p.typeScript
}

// IsStandalone reports whether this Program owns parser/binder-backed sources
// without a ts-go Program or TypeChecker.
func (p *Program) IsStandalone() bool {
	return p != nil && p.standalone != nil
}

// SourceFiles returns the complete immutable source universe in stable order.
// The returned slice and its entries are read-only for the Program's lifetime.
func (p *Program) SourceFiles() []*ast.SourceFile {
	if p == nil {
		return nil
	}
	if p.typeScript != nil {
		return p.typeScript.SourceFiles()
	}
	if p.standalone != nil {
		return p.standalone.files
	}
	return nil
}

// OwnsSourceFile reports whether file is the exact AST generation exposed by
// this Program.
func (p *Program) OwnsSourceFile(file *ast.SourceFile) bool {
	if p == nil || file == nil {
		return false
	}
	if p.typeScript != nil {
		return p.typeScript.GetSourceFile(file.FileName()) == file
	}
	if p.standalone == nil {
		return false
	}
	return p.standalone.sourcesByPath[file.Path()] == file
}

func (p *Program) Options() *core.CompilerOptions {
	if p == nil {
		return nil
	}
	if p.typeScript != nil {
		return p.typeScript.Options()
	}
	if p.standalone != nil {
		return p.standalone.options
	}
	return nil
}

func (p *Program) FS() vfs.FS {
	if p == nil {
		return nil
	}
	if p.typeScript != nil {
		return p.typeScript.Host().FS()
	}
	if p.standalone != nil {
		return p.standalone.fs
	}
	return nil
}

func (p *Program) CurrentDirectory() string {
	if p == nil {
		return ""
	}
	if p.typeScript != nil {
		return p.typeScript.Host().GetCurrentDirectory()
	}
	if p.standalone != nil {
		return p.standalone.currentDirectory
	}
	return ""
}

func (p *Program) NearestPackageJSONDirectory(directory string) string {
	if p == nil {
		return ""
	}
	if p.typeScript != nil {
		return p.typeScript.GetNearestAncestorDirectoryWithPackageJson(directory)
	}
	if p.standalone == nil {
		return ""
	}
	scope := p.standalone.resolver.GetPackageScopeForPath(directory)
	if scope.Exists() {
		return scope.PackageDirectory
	}
	return ""
}

func (p *Program) FileExists(path string) bool {
	if p == nil {
		return false
	}
	if p.typeScript != nil {
		return p.typeScript.FileExists(path)
	}
	fs := p.FS()
	return fs != nil && fs.FileExists(path)
}

func (p *Program) GetModeForUsageLocation(sourceFile ast.HasFileName, location *ast.StringLiteralLike) core.ResolutionMode {
	if p == nil {
		return core.ResolutionModeNone
	}
	if p.typeScript != nil {
		return p.typeScript.GetModeForUsageLocation(sourceFile, location)
	}
	if p.standalone == nil || sourceFile == nil {
		return core.ResolutionModeNone
	}
	return compiler.GetModeForUsageLocation(
		sourceFile.FileName(),
		p.standalone.metadataByPath[sourceFile.Path()],
		location,
		p.standalone.options,
	)
}

func (p *Program) GetResolvedModule(sourceFile ast.HasFileName, moduleName string, mode core.ResolutionMode) *module.ResolvedModule {
	if p == nil {
		return nil
	}
	if p.typeScript != nil {
		return p.typeScript.GetResolvedModule(sourceFile, moduleName, mode)
	}
	if p.standalone == nil || sourceFile == nil {
		return nil
	}
	return p.standalone.resolvedModules[sourceFile.Path()][module.ModeAwareCacheKey{Name: moduleName, Mode: mode}]
}

func (p *Program) ResolveModuleName(moduleName string, containingFile string, mode core.ResolutionMode) *module.ResolvedModule {
	if p == nil {
		return nil
	}
	if p.typeScript != nil {
		return p.typeScript.ResolveModuleName(moduleName, containingFile, mode)
	}
	if p.standalone == nil {
		return nil
	}
	resolved, _ := p.standalone.resolver.ResolveModuleName(moduleName, containingFile, mode, nil)
	return resolved
}

func (p *Program) GetSourceFileForResolvedModule(fileName string) *ast.SourceFile {
	if p == nil {
		return nil
	}
	if p.typeScript != nil {
		return p.typeScript.GetSourceFileForResolvedModule(fileName)
	}
	return p.GetSourceFile(fileName)
}

func (p *Program) GetSourceFile(fileName string) *ast.SourceFile {
	if p == nil {
		return nil
	}
	if p.typeScript != nil {
		return p.typeScript.GetSourceFile(fileName)
	}
	if p.standalone == nil {
		return nil
	}
	return p.standalone.sourceFile(fileName)
}

// SourceFileMetadata returns the module-format metadata used when the file was
// parsed and resolved.
func (p *Program) SourceFileMetadata(file *ast.SourceFile) ast.SourceFileMetaData {
	if p == nil || file == nil {
		return ast.SourceFileMetaData{}
	}
	if p.typeScript != nil {
		return p.typeScript.GetSourceFileMetaData(file.Path())
	}
	if p.standalone != nil {
		return p.standalone.metadataByPath[file.Path()]
	}
	return ast.SourceFileMetaData{}
}

// SyntacticDiagnostics returns the backend's complete syntactic diagnostics
// for file. Standalone JavaScript includes ts-go's additional JS syntactic
// checks under the same compiler options used to parse the source.
func (p *Program) SyntacticDiagnostics(ctx context.Context, file *ast.SourceFile) []*ast.Diagnostic {
	if p == nil || file == nil {
		return nil
	}
	if p.typeScript != nil {
		return p.typeScript.GetSyntacticDiagnostics(ctx, file)
	}
	if p.standalone == nil {
		return nil
	}
	diagnostics := append([]*ast.Diagnostic(nil), file.Diagnostics()...)
	diagnostics = append(diagnostics, file.JSDiagnostics()...)
	if ast.IsSourceFileJS(file) && !ast.IsCheckJSEnabledForFile(file, p.standalone.options) {
		diagnostics = append(diagnostics, compiler.GetAdditionalJSSyntacticDiagnostics(file, p.standalone.options)...)
	}
	return compiler.SortAndDeduplicateDiagnostics(diagnostics)
}

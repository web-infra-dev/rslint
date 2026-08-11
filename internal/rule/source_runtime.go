package rule

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/module"
	"github.com/microsoft/typescript-go/shim/vfs"
)

type programSourceRuntime struct {
	program *compiler.Program
}

// SourceRuntimeForProgram adapts the non-type-aware services of a real
// TypeScript Program to SourceRuntime. A nil Program has no source runtime.
func SourceRuntimeForProgram(program *compiler.Program) SourceRuntime {
	if program == nil {
		return nil
	}
	return programSourceRuntime{program: program}
}

func (r programSourceRuntime) SameSourceRuntime(other SourceRuntime) bool {
	otherRuntime, ok := other.(programSourceRuntime)
	return ok && r.program == otherRuntime.program
}

func (r programSourceRuntime) OwnsSourceFile(file *ast.SourceFile) bool {
	return file != nil && r.program.GetSourceFile(file.FileName()) == file
}

func (r programSourceRuntime) Options() *core.CompilerOptions {
	return r.program.Options()
}

func (r programSourceRuntime) FS() vfs.FS {
	return r.program.Host().FS()
}

func (r programSourceRuntime) CurrentDirectory() string {
	return r.program.Host().GetCurrentDirectory()
}

func (r programSourceRuntime) NearestPackageJSONDirectory(directory string) string {
	return r.program.GetNearestAncestorDirectoryWithPackageJson(directory)
}

func (r programSourceRuntime) FileExists(path string) bool {
	return r.program.FileExists(path)
}

func (r programSourceRuntime) GetModeForUsageLocation(sourceFile ast.HasFileName, location *ast.StringLiteralLike) core.ResolutionMode {
	return r.program.GetModeForUsageLocation(sourceFile, location)
}

func (r programSourceRuntime) GetResolvedModule(sourceFile ast.HasFileName, moduleReference string, mode core.ResolutionMode) *module.ResolvedModule {
	return r.program.GetResolvedModule(sourceFile, moduleReference, mode)
}

func (r programSourceRuntime) ResolveModuleName(moduleName string, containingFile string, mode core.ResolutionMode) *module.ResolvedModule {
	return r.program.ResolveModuleName(moduleName, containingFile, mode)
}

func (r programSourceRuntime) GetSourceFileForResolvedModule(fileName string) *ast.SourceFile {
	return r.program.GetSourceFileForResolvedModule(fileName)
}

func (r programSourceRuntime) GetSourceFile(fileName string) *ast.SourceFile {
	return r.program.GetSourceFile(fileName)
}

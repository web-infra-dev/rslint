package lsp

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
)

// UpdateProgram already checks imports, references, parse options, and other
// graph inputs. These source-local values also affect module resolution but
// are not part of its reusable-graph check.
func lintProgramGraphReusable(
	oldProgram *compiler.Program,
	newProgram *compiler.Program,
	oldFile *ast.SourceFile,
	newFile *ast.SourceFile,
) bool {
	if oldFile.Flags&ast.NodeFlagsPermanentlySetIncrementalFlags !=
		newFile.Flags&ast.NodeFlagsPermanentlySetIncrementalFlags {
		return false
	}

	options := lintCompilerOptionsForSourceFile(oldProgram, oldFile)
	if ast.GetJSXRuntimeImport(ast.GetJSXImplicitImportBase(options, oldFile), options) !=
		ast.GetJSXRuntimeImport(ast.GetJSXImplicitImportBase(options, newFile), options) {
		return false
	}

	oldImports, newImports := oldFile.Imports(), newFile.Imports()
	if len(oldImports) != len(newImports) {
		return false
	}
	for index, oldImport := range oldImports {
		if oldProgram.GetModeForUsageLocation(oldFile, oldImport) !=
			newProgram.GetModeForUsageLocation(newFile, newImports[index]) {
			return false
		}
	}

	if len(oldFile.ModuleAugmentations) != len(newFile.ModuleAugmentations) {
		return false
	}
	for index, oldAugmentation := range oldFile.ModuleAugmentations {
		if oldAugmentation.Kind == ast.KindStringLiteral &&
			oldProgram.GetModeForUsageLocation(oldFile, oldAugmentation) !=
				newProgram.GetModeForUsageLocation(newFile, newFile.ModuleAugmentations[index]) {
			return false
		}
	}
	return true
}

func lintCompilerOptionsForSourceFile(
	program *compiler.Program,
	sourceFile *ast.SourceFile,
) *core.CompilerOptions {
	if redirect := program.GetRedirectForResolution(sourceFile); redirect != nil {
		if options := redirect.CompilerOptions(); options != nil {
			return options
		}
	}
	return program.Options()
}

package utils

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/program"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

// Match eslint-module-utils/unambiguous before consulting the parsed module.
// In particular, side-effect imports and spaced dynamic imports alone do not
// pass this text check. JavaScript whitespace and multiline anchors matter.
var defaultImportModulePattern = esregexp.MustCompile(`(^|[;})])\s*(export|import)((\s+\w)|(\s*[{*=]))|import\(`, "m")

type defaultImportInfo struct {
	available        bool
	syntheticDefault bool
}

// Default-import compatibility is separate from the authored export index:
// an interop default must not become a named export for other import rules.
func (index *ModuleIndex) defaultImportInfoOf(sourceProgram *program.Program, file *ast.SourceFile) defaultImportInfo {
	return index.defaultImports.Get(file, func() defaultImportInfo {
		if !defaultImportModulePattern.Test(file.Text()) {
			return defaultImportInfo{}
		}
		isModuleDeclaration := func(stmt *ast.Node) bool {
			return stmt != nil && (stmt.Kind == ast.KindImportDeclaration || stmt.Kind == ast.KindExportDeclaration ||
				stmt.Kind == ast.KindExportAssignment || ast.HasSyntacticModifier(stmt, ast.ModifierFlagsExport))
		}
		if !isModuleDeclaration(file.ExternalModuleIndicator) &&
			(file.Statements == nil || !slices.ContainsFunc(file.Statements.Nodes, isModuleDeclaration)) &&
			(file.AsNode().Flags&ast.NodeFlagsPossiblyContainsDynamicImport == 0 || file.SubtreeFacts()&ast.SubtreeContainsDynamicImport == 0) {
			return defaultImportInfo{}
		}
		info := defaultImportInfo{available: true}
		interop := explicitESModuleInterop(sourceProgram)
		syntheticDefault := compilerOptionsAllowSyntheticDefaultImports(sourceProgram, interop)
		// Only declarations use the implicit-default fallback below.
		// Export assignments are checked during the subsequent name lookup.
		if !interop && (!syntheticDefault || !file.IsDeclarationFile) {
			return info
		}
		if !syntheticDefault {
			// Do not let upstream's local-namespace fallback override an
			// explicit synthetic-default opt-out for CommonJS declarations
			// or export assignments. Authored defaults are still checked.
			if isCommonJSDeclaration(sourceProgram, file) || file.Statements != nil && slices.ContainsFunc(file.Statements.Nodes, func(stmt *ast.Node) bool {
				return stmt != nil && stmt.Kind == ast.KindExportAssignment && stmt.AsExportAssignment().IsExportEquals
			}) {
				return info
			}
		}
		local := index.localExportsOf(sourceProgram, file)
		if syntheticDefault && file.IsDeclarationFile && local.ImplicitDefault {
			info.syntheticDefault = true
			return info
		}
		if !interop {
			return info
		}
		// Upstream synthesizes a default only for its local namespace. Named
		// re-exports and export-star dependencies are stored separately.
		for _, step := range local.Steps {
			if step.Kind == exportStepLocalDefault ||
				step.Kind == exportStepNames && len(step.Names) != 0 ||
				step.Kind == exportStepNamed && !step.FromModule && len(step.Specs) != 0 {
				info.syntheticDefault = true
				break
			}
		}
		return info
	})
}

// Upstream reads the parsed tsconfig field rather than TypeScript's inferred
// interop default for a particular module kind.
//
//nolint:staticcheck // Match the explicit option consumed by eslint-plugin-import.
func explicitESModuleInterop(sourceProgram *program.Program) bool {
	return sourceProgram.Options() != nil && sourceProgram.Options().ESModuleInterop == core.TSTrue
}

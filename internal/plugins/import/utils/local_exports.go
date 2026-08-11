package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
	rslint_utils "github.com/web-infra-dev/rslint/internal/utils"
)

// localExports is everything one file says about its own exports, with every
// module specifier resolved and none of them followed. It is a function of
// that file's syntax alone, which is what makes it safe to compute once per
// effective source set: what a dependency exports can change with the order
// modules are visited in, but what a file itself declares cannot.
type localExports struct {
	// Steps replay the file's export statements in source order. Applying
	// them in order reproduces what re-walking the statements would produce,
	// including a later statement overwriting an earlier one.
	Steps []exportStep
	// Imports lists the file's import declarations that have an import
	// clause, in source order, each with its resolved target. Resolving a
	// locally re-exported namespace walks all of them, and the order matters
	// because that walk builds their export maps along the way.
	Imports []importedModule
}

type exportStepKind uint8

const (
	// exportStepNames declares names outright: an exported declaration, a
	// namespace export, an `export as namespace` clause.
	exportStepNames exportStepKind = iota
	// exportStepLocalDefault is `export default <expression>`, whose meta
	// comes from a namespace import when the expression names one.
	exportStepLocalDefault
	// exportStepStar is `export * from`, which merges the dependency's map.
	exportStepStar
	// exportStepNamed is `export { … }`, with or without a module specifier.
	exportStepNamed
)

type exportStep struct {
	Kind  exportStepKind
	Names []string
	// Local is the identifier `export default` names, empty when the
	// expression is not a plain identifier.
	Local string
	Link  exportLink
	Specs []exportSpec
	// FromModule distinguishes `export { a } from "m"` from `export { a }`.
	FromModule bool
}

// exportLink is a resolved `… from "specifier"`. Resolved is false when the
// specifier does not resolve, is covered by `import/ignore`, or names a file
// that is not an ES module — the three cases the export map treats alike.
type exportLink struct {
	Target   *ast.SourceFile
	Resolved bool
}

type exportSpec struct {
	Exported string
	// Local is the name looked up in the dependency. LocalOK is false when
	// the specifier's source name is not a static name.
	Local   string
	LocalOK bool
	// LocalIdent is the source name's identifier text, empty when it is not a
	// plain identifier. Only a plain identifier can name a namespace import.
	LocalIdent string
}

// importedModule is one `import … from "specifier"` with an import clause.
type importedModule struct {
	Link exportLink
	// NamespaceName is the local name of an `import * as ns` clause, empty
	// for every other clause shape.
	NamespaceName string
}

func collectLocalExports(runtime rslint_utils.ModuleResolutionRuntime, sourceFile *ast.SourceFile, settings *ModuleSettings) *localExports {
	local := &localExports{}
	if sourceFile == nil || sourceFile.Statements == nil {
		return local
	}

	for _, stmt := range sourceFile.Statements.Nodes {
		if stmt == nil {
			continue
		}
		if stmt.Kind == ast.KindImportDeclaration {
			local.Imports = append(local.Imports, importBinding(runtime, sourceFile, settings, stmt.AsImportDeclaration()))
		}
		local.appendStatement(runtime, sourceFile, settings, stmt)
	}

	return local
}

func (local *localExports) appendStatement(runtime rslint_utils.ModuleResolutionRuntime, sourceFile *ast.SourceFile, settings *ModuleSettings, stmt *ast.Node) {
	if isExportedDeclaration(stmt) {
		if names := exportedDeclarationNames(stmt); len(names) > 0 {
			local.Steps = append(local.Steps, exportStep{Kind: exportStepNames, Names: names})
		}
		return
	}

	switch stmt.Kind {
	case ast.KindExportAssignment:
		exportAssignment := stmt.AsExportAssignment()
		if exportAssignment.IsExportEquals {
			return
		}
		local.Steps = append(local.Steps, exportStep{Kind: exportStepLocalDefault, Local: referencedIdentifierText(exportAssignment.Expression)})
	case ast.KindNamespaceExportDeclaration:
		if name := stmt.Name(); name != nil {
			local.Steps = append(local.Steps, exportStep{Kind: exportStepNames, Names: []string{name.Text()}})
		}
	case ast.KindExportDeclaration:
		local.appendExportDeclaration(runtime, sourceFile, settings, stmt.AsExportDeclaration())
	}
}

func (local *localExports) appendExportDeclaration(runtime rslint_utils.ModuleResolutionRuntime, sourceFile *ast.SourceFile, settings *ModuleSettings, exportDecl *ast.ExportDeclaration) {
	if exportDecl == nil {
		return
	}

	if exportDecl.ExportClause == nil {
		if exportDecl.ModuleSpecifier == nil {
			return
		}
		local.Steps = append(local.Steps, exportStep{
			Kind: exportStepStar,
			Link: resolveExportLink(runtime, sourceFile, settings, exportDecl.ModuleSpecifier),
		})
		return
	}

	switch exportDecl.ExportClause.Kind {
	case ast.KindNamedExports:
		namedExports := exportDecl.ExportClause.AsNamedExports()
		if namedExports == nil || namedExports.Elements == nil {
			return
		}
		step := exportStep{Kind: exportStepNamed, FromModule: exportDecl.ModuleSpecifier != nil}
		if step.FromModule {
			step.Link = resolveExportLink(runtime, sourceFile, settings, exportDecl.ModuleSpecifier)
		}
		for _, spec := range namedExports.Elements.Nodes {
			if spec == nil || spec.Kind != ast.KindExportSpecifier {
				continue
			}
			specifier := spec.AsExportSpecifier()
			exportedName, ok := moduleExportName(specifier.Name())
			if !ok {
				continue
			}
			sourceName := specifier.PropertyName
			if sourceName == nil {
				sourceName = specifier.Name()
			}
			localName, localOK := moduleExportName(sourceName)
			step.Specs = append(step.Specs, exportSpec{
				Exported:   exportedName,
				Local:      localName,
				LocalOK:    localOK,
				LocalIdent: referencedIdentifierText(sourceName),
			})
		}
		local.Steps = append(local.Steps, step)
	case ast.KindNamespaceExport:
		namespaceExport := exportDecl.ExportClause.AsNamespaceExport()
		if namespaceExport == nil || namespaceExport.Name() == nil {
			return
		}
		name, ok := moduleExportName(namespaceExport.Name())
		if !ok {
			return
		}
		local.Steps = append(local.Steps, exportStep{Kind: exportStepNames, Names: []string{name}})
	}
}

func importBinding(runtime rslint_utils.ModuleResolutionRuntime, sourceFile *ast.SourceFile, settings *ModuleSettings, importDecl *ast.ImportDeclaration) importedModule {
	binding := importedModule{}
	if importDecl == nil || importDecl.ImportClause == nil {
		return binding
	}
	binding.Link = resolveExportLink(runtime, sourceFile, settings, importDecl.ModuleSpecifier)

	importClause := importDecl.ImportClause.AsImportClause()
	if importClause == nil || importClause.NamedBindings == nil || importClause.NamedBindings.Kind != ast.KindNamespaceImport {
		return binding
	}
	namespaceImport := importClause.NamedBindings.AsNamespaceImport()
	if namespaceImport != nil && namespaceImport.Name() != nil {
		binding.NamespaceName = namespaceImport.Name().Text()
	}
	return binding
}

// resolveExportLink answers what getExportMap would answer for one specifier,
// without building the target's map.
func resolveExportLink(runtime rslint_utils.ModuleResolutionRuntime, sourceFile *ast.SourceFile, settings *ModuleSettings, moduleSpecifier *ast.Node) exportLink {
	if runtime == nil || moduleSpecifier == nil || !ast.IsStringLiteralLike(moduleSpecifier) {
		return exportLink{}
	}
	_, target, ok := rslint_utils.ResolveModuleFile(runtime, sourceFile, moduleSpecifier)
	if !ok || settings.IsIgnoredPath(target.FileName()) || !ast.IsExternalModule(target) {
		return exportLink{}
	}
	return exportLink{Target: target, Resolved: true}
}

func exportedDeclarationNames(stmt *ast.Node) []string {
	if ast.HasSyntacticModifier(stmt, ast.ModifierFlagsDefault) {
		return []string{defaultExportName}
	}

	switch stmt.Kind {
	case ast.KindVariableStatement:
		var names []string
		collectVariableStatementNames(stmt, func(name string) {
			names = append(names, name)
		})
		return names
	case ast.KindFunctionDeclaration,
		ast.KindClassDeclaration,
		ast.KindInterfaceDeclaration,
		ast.KindTypeAliasDeclaration,
		ast.KindEnumDeclaration,
		ast.KindModuleDeclaration:
		if name := stmt.Name(); name != nil {
			return []string{name.Text()}
		}
	}
	return nil
}

// referencedIdentifierText returns the identifier an expression is, after
// parentheses, and "" for every other expression. A named function or class
// expression does not count: its name binds inside its own body, not to a
// binding of the file, so it can never reference a namespace import.
func referencedIdentifierText(expr *ast.Node) string {
	expr = ast.SkipParentheses(expr)
	if expr == nil || expr.Kind != ast.KindIdentifier {
		return ""
	}
	return expr.AsIdentifier().Text
}

func collectVariableStatementNames(stmt *ast.Node, visit func(name string)) {
	if stmt == nil || visit == nil {
		return
	}
	declList := stmt.AsVariableStatement().DeclarationList
	if declList == nil || !ast.IsVariableDeclarationList(declList) {
		return
	}
	for _, decl := range declList.AsVariableDeclarationList().Declarations.Nodes {
		if decl == nil || !ast.IsVariableDeclaration(decl) {
			continue
		}
		rslint_utils.CollectBindingNames(decl.AsVariableDeclaration().Name(), func(_ *ast.Node, bindingName string) {
			visit(bindingName)
		})
	}
}

// isExportedDeclaration reports whether stmt is a declaration exported by its
// own modifier, the statements exportedDeclarationNames reads names from.
func isExportedDeclaration(stmt *ast.Node) bool {
	if stmt == nil {
		return false
	}
	switch stmt.Kind {
	case ast.KindVariableStatement,
		ast.KindFunctionDeclaration,
		ast.KindClassDeclaration,
		ast.KindInterfaceDeclaration,
		ast.KindTypeAliasDeclaration,
		ast.KindEnumDeclaration,
		ast.KindModuleDeclaration:
		return ast.HasSyntacticModifier(stmt, ast.ModifierFlagsExport)
	}
	return false
}

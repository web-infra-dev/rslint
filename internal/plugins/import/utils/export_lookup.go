package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	rslint_utils "github.com/web-infra-dev/rslint/internal/utils"
)

// This file answers whether a module statically exports one name, by walking
// the target's statements directly. It is the looser counterpart of the export
// map in export_map.go: it synthesizes default exports from compiler interop
// settings, which the map deliberately does not.

// HasDefaultExport resolves moduleSpecifier from ctx.SourceFile and reports
// whether the resolved module has a statically visible default export. The
// second result is false when no export map is available, matching
// eslint-plugin-import's "imports == null" branch.
func HasDefaultExport(ctx rule.RuleContext, moduleSpecifier *ast.Node) (bool, bool) {
	return HasExport(ctx, moduleSpecifier, defaultExportName)
}

// HasExport resolves moduleSpecifier from ctx.SourceFile and reports whether
// the resolved module statically exports exportName. The second result is false
// when the target is unresolved or is not an ES module.
func HasExport(ctx rule.RuleContext, moduleSpecifier *ast.Node, exportName string) (bool, bool) {
	if !ctx.Program().IsValid() || ctx.SourceFile == nil || moduleSpecifier == nil || !ast.IsStringLiteralLike(moduleSpecifier) {
		return false, false
	}
	return hasExport(ctx.SourceFile, moduleSpecifier, exportName, newExportBuilder(IndexFor(ctx), ctx.Program()))
}

// exportKey is one (file, name) lookup in flight, so a re-export chain that
// reaches the same question again is answered "not found" instead of recursing
// forever.
type exportKey struct {
	file *ast.SourceFile
	name string
}

func hasExport(origin *ast.SourceFile, moduleSpecifier *ast.Node, exportName string, builder *exportBuilder) (bool, bool) {
	link := resolveExportLinkForLookup(builder.program(), origin, builder.index.settings, moduleSpecifier)
	if link.Target == nil {
		return false, false
	}
	return sourceFileHasExport(link.Target, exportName, builder)
}

// resolveExportLinkForLookup is the name-lookup counterpart of
// resolveExportLink: it stops before the is-an-ES-module test, which
// sourceFileHasExport applies itself.
func resolveExportLinkForLookup(sourceProgram *program.Program, origin *ast.SourceFile, settings *ModuleSettings, moduleSpecifier *ast.Node) exportLink {
	_, sourceFile, ok := sourceProgram.ResolveModule(origin, moduleSpecifier)
	if !ok || sourceFile == nil || settings.IsIgnoredPath(sourceFile.FileName()) {
		return exportLink{}
	}
	return exportLink{Target: sourceFile, Resolved: true}
}

func sourceFileHasExport(sourceFile *ast.SourceFile, exportName string, builder *exportBuilder) (bool, bool) {
	if sourceFile == nil || !ast.IsExternalModule(sourceFile) {
		return false, false
	}

	key := exportKey{file: sourceFile, name: exportName}
	if builder.seen[key] {
		return false, true
	}
	builder.seen[key] = true
	defer delete(builder.seen, key)

	statements := sourceFile.Statements
	if statements == nil {
		return false, true
	}

	for _, stmt := range statements.Nodes {
		if stmt == nil {
			continue
		}

		if exportedDeclarationHasName(stmt, exportName) {
			return true, true
		}

		switch stmt.Kind {
		case ast.KindExportAssignment:
			if exportName == defaultExportName && exportAssignmentHasDefault(builder.program(), sourceFile, stmt.AsExportAssignment()) {
				return true, true
			}
		case ast.KindNamespaceExportDeclaration:
			if exportName == defaultExportName && compilerOptionsESModuleInterop(builder.program()) {
				return true, true
			}
		case ast.KindExportDeclaration:
			found, done := exportDeclarationHasName(sourceFile, stmt.AsExportDeclaration(), exportName, builder)
			if done {
				return found, true
			}
		}
	}

	if exportName == defaultExportName && compilerOptionsESModuleInterop(builder.program()) && sourceFileHasDirectNamespaceExport(sourceFile) {
		return true, true
	}

	return false, true
}

func exportedDeclarationHasName(stmt *ast.Node, exportName string) bool {
	if !ast.HasSyntacticModifier(stmt, ast.ModifierFlagsExport) {
		return false
	}

	if ast.HasSyntacticModifier(stmt, ast.ModifierFlagsDefault) {
		return exportName == defaultExportName
	}

	switch stmt.Kind {
	case ast.KindVariableStatement:
		return variableStatementDeclaresName(stmt, exportName)
	case ast.KindFunctionDeclaration,
		ast.KindClassDeclaration,
		ast.KindInterfaceDeclaration,
		ast.KindTypeAliasDeclaration,
		ast.KindEnumDeclaration,
		ast.KindModuleDeclaration:
		name := stmt.Name()
		return name != nil && moduleExportNameMatches(name, exportName)
	}

	return false
}

func exportAssignmentHasDefault(sourceProgram *program.Program, sourceFile *ast.SourceFile, exportAssignment *ast.ExportAssignment) bool {
	if exportAssignment == nil {
		return false
	}
	if !exportAssignment.IsExportEquals {
		return true
	}

	// Match eslint-plugin-import's TypeScript export-assignment visitor:
	// `export = namespace` gets a synthetic default only under esModuleInterop,
	// while non-namespace local declarations and re-export-like expressions do.
	name, ok := exportAssignmentReferencedIdentifier(exportAssignment.Expression)
	if !ok {
		return true
	}
	kind, ok := sourceFileExportAssignmentLocalDeclarationKind(sourceFile, name)
	if !ok {
		return true
	}
	if kind != exportAssignmentLocalDeclarationModule {
		return true
	}
	return compilerOptionsESModuleInterop(sourceProgram)
}

// exportAssignmentReferencedIdentifier returns the identifier an expression
// names, after parentheses. A named function or class expression names
// itself.
func exportAssignmentReferencedIdentifier(expr *ast.Node) (string, bool) {
	expr = ast.SkipParentheses(expr)
	if expr == nil {
		return "", false
	}
	switch expr.Kind {
	case ast.KindIdentifier:
		return expr.AsIdentifier().Text, true
	case ast.KindFunctionExpression, ast.KindClassExpression:
		if name := expr.Name(); name != nil {
			return name.Text(), true
		}
	}
	return "", false
}

// The tsgo shim exposes CompilerOptions fields but not GetESModuleInterop.
//
//nolint:staticcheck // esModuleInterop still needs to be inspected for import/export compatibility.
func compilerOptionsESModuleInterop(sourceProgram *program.Program) bool {
	if !sourceProgram.IsValid() || sourceProgram.Options() == nil {
		return false
	}
	options := sourceProgram.Options()
	if options.ESModuleInterop != core.TSUnknown {
		return options.ESModuleInterop == core.TSTrue
	}
	switch options.Module {
	case core.ModuleKindNode16, core.ModuleKindNodeNext, core.ModuleKindPreserve:
		return true
	default:
		return false
	}
}

type exportAssignmentLocalDeclarationKind int

const (
	exportAssignmentLocalDeclarationOther exportAssignmentLocalDeclarationKind = iota
	exportAssignmentLocalDeclarationModule
)

func sourceFileExportAssignmentLocalDeclarationKind(sourceFile *ast.SourceFile, name string) (exportAssignmentLocalDeclarationKind, bool) {
	if sourceFile == nil || sourceFile.Statements == nil {
		return exportAssignmentLocalDeclarationOther, false
	}
	for _, stmt := range sourceFile.Statements.Nodes {
		if stmt == nil {
			continue
		}

		switch stmt.Kind {
		case ast.KindVariableStatement:
			if variableStatementDeclaresName(stmt, name) {
				return exportAssignmentLocalDeclarationOther, true
			}
		case ast.KindFunctionDeclaration:
			if ast.HasSyntacticModifier(stmt, ast.ModifierFlagsAmbient) && declarationHasName(stmt, name) {
				return exportAssignmentLocalDeclarationOther, true
			}
		case ast.KindClassDeclaration,
			ast.KindInterfaceDeclaration,
			ast.KindTypeAliasDeclaration,
			ast.KindEnumDeclaration:
			if declarationHasName(stmt, name) {
				return exportAssignmentLocalDeclarationOther, true
			}
		case ast.KindModuleDeclaration:
			if declarationHasName(stmt, name) {
				return exportAssignmentLocalDeclarationModule, true
			}
		}
	}
	return exportAssignmentLocalDeclarationOther, false
}

func declarationHasName(stmt *ast.Node, name string) bool {
	declName := stmt.Name()
	return declName != nil && moduleExportNameMatches(declName, name)
}

func variableStatementDeclaresName(stmt *ast.Node, name string) bool {
	declList := stmt.AsVariableStatement().DeclarationList
	if declList == nil || !ast.IsVariableDeclarationList(declList) {
		return false
	}
	for _, decl := range declList.AsVariableDeclarationList().Declarations.Nodes {
		if decl == nil || !ast.IsVariableDeclaration(decl) {
			continue
		}
		matched := false
		rslint_utils.CollectBindingNames(decl.AsVariableDeclaration().Name(), func(_ *ast.Node, bindingName string) {
			if bindingName == name {
				matched = true
			}
		})
		if matched {
			return true
		}
	}
	return false
}

func sourceFileHasDirectNamespaceExport(sourceFile *ast.SourceFile) bool {
	if sourceFile == nil || sourceFile.Statements == nil {
		return false
	}
	for _, stmt := range sourceFile.Statements.Nodes {
		if stmt == nil {
			continue
		}
		if exportedDeclarationAddsNamespaceExport(stmt) {
			return true
		}
		if stmt.Kind == ast.KindExportDeclaration && exportDeclarationAddsNamespaceExport(stmt.AsExportDeclaration()) {
			return true
		}
	}
	return false
}

func exportedDeclarationAddsNamespaceExport(stmt *ast.Node) bool {
	if !ast.HasSyntacticModifier(stmt, ast.ModifierFlagsExport) {
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
		return true
	}
	return false
}

func exportDeclarationAddsNamespaceExport(exportDecl *ast.ExportDeclaration) bool {
	if exportDecl == nil || exportDecl.ModuleSpecifier != nil || exportDecl.ExportClause == nil {
		return false
	}
	switch exportDecl.ExportClause.Kind {
	case ast.KindNamedExports:
		namedExports := exportDecl.ExportClause.AsNamedExports()
		return namedExports.Elements != nil && len(namedExports.Elements.Nodes) > 0
	case ast.KindNamespaceExport:
		return true
	}
	return false
}

func exportDeclarationHasName(sourceFile *ast.SourceFile, exportDecl *ast.ExportDeclaration, exportName string, builder *exportBuilder) (bool, bool) {
	if exportDecl == nil {
		return false, false
	}

	if exportDecl.ExportClause == nil {
		if exportDecl.ModuleSpecifier == nil || exportName == defaultExportName {
			return false, false
		}
		found, ok := hasExport(sourceFile, exportDecl.ModuleSpecifier, exportName, builder)
		if !ok {
			return true, true
		}
		return found, found
	}

	switch exportDecl.ExportClause.Kind {
	case ast.KindNamedExports:
		namedExports := exportDecl.ExportClause.AsNamedExports()
		if namedExports.Elements == nil {
			return false, false
		}
		for _, spec := range namedExports.Elements.Nodes {
			if spec == nil || spec.Kind != ast.KindExportSpecifier {
				continue
			}

			exportSpec := spec.AsExportSpecifier()
			if !moduleExportNameMatches(exportSpec.Name(), exportName) {
				continue
			}

			if exportDecl.ModuleSpecifier == nil {
				return true, true
			}

			sourceName := exportSpec.PropertyName
			if sourceName == nil {
				sourceName = exportSpec.Name()
			}

			localName, ok := moduleExportName(sourceName)
			if !ok {
				return false, true
			}

			hasName, ok := hasExport(sourceFile, exportDecl.ModuleSpecifier, localName, builder)
			if !ok {
				return true, true
			}
			return hasName, hasName
		}
	case ast.KindNamespaceExport:
		namespaceExport := exportDecl.ExportClause.AsNamespaceExport()
		matched := moduleExportNameMatches(namespaceExport.Name(), exportName)
		return matched, matched
	}

	return false, false
}

func moduleExportNameMatches(node *ast.Node, exportName string) bool {
	if node == nil {
		return false
	}
	if exportName == defaultExportName {
		return ast.ModuleExportNameIsDefault(node)
	}
	name, ok := moduleExportName(node)
	return ok && name == exportName
}

package utils

import (
	"regexp"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	rslint_utils "github.com/web-infra-dev/rslint/internal/utils"
)

const defaultExportName = "default"

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
	if ctx.Program == nil || ctx.SourceFile == nil || moduleSpecifier == nil || !ast.IsStringLiteralLike(moduleSpecifier) {
		return false, false
	}
	return hasExport(ctx, ctx.SourceFile, moduleSpecifier, exportName, make(map[exportKey]bool))
}

type exportKey struct {
	fileName string
	name     string
}

func hasExport(ctx rule.RuleContext, origin *ast.SourceFile, moduleSpecifier *ast.Node, exportName string, seen map[exportKey]bool) (bool, bool) {
	resolved := ctx.Program.GetResolvedModuleFromModuleSpecifier(origin, moduleSpecifier)
	if resolved == nil || resolved.ResolvedFileName == "" {
		return false, false
	}

	sourceFile := ctx.Program.GetSourceFileForResolvedModule(resolved.ResolvedFileName)
	if sourceFile == nil {
		return false, false
	}
	if IsImportPathIgnored(ctx.Settings, sourceFile.FileName()) {
		return false, false
	}

	return sourceFileHasExport(ctx, sourceFile, exportName, seen)
}

// IsImportPathIgnored matches eslint-plugin-import's shared `import/ignore`
// setting for resolved import target paths.
func IsImportPathIgnored(settings map[string]interface{}, fileName string) bool {
	if settings == nil {
		return false
	}

	rawPatterns, ok := settings["import/ignore"]
	if !ok {
		return false
	}

	var patterns []string
	switch typed := rawPatterns.(type) {
	case []string:
		patterns = typed
	case []interface{}:
		for _, item := range typed {
			if pattern, ok := item.(string); ok {
				patterns = append(patterns, pattern)
			}
		}
	}

	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err == nil && re.MatchString(fileName) {
			return true
		}
	}
	return false
}

func sourceFileHasExport(ctx rule.RuleContext, sourceFile *ast.SourceFile, exportName string, seen map[exportKey]bool) (bool, bool) {
	if sourceFile == nil || !ast.IsExternalModule(sourceFile) {
		return false, false
	}

	key := exportKey{fileName: sourceFile.FileName(), name: exportName}
	if seen[key] {
		return false, true
	}
	seen[key] = true
	defer delete(seen, key)

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
			if exportName == defaultExportName && exportAssignmentHasDefault(ctx, sourceFile, stmt.AsExportAssignment()) {
				return true, true
			}
		case ast.KindNamespaceExportDeclaration:
			if exportName == defaultExportName && compilerOptionsESModuleInterop(ctx) {
				return true, true
			}
		case ast.KindExportDeclaration:
			found, done := exportDeclarationHasName(ctx, sourceFile, stmt.AsExportDeclaration(), exportName, seen)
			if done {
				return found, true
			}
		}
	}

	if exportName == defaultExportName && compilerOptionsESModuleInterop(ctx) && sourceFileHasDirectNamespaceExport(sourceFile) {
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

func exportAssignmentHasDefault(ctx rule.RuleContext, sourceFile *ast.SourceFile, exportAssignment *ast.ExportAssignment) bool {
	if exportAssignment == nil {
		return false
	}
	if !exportAssignment.IsExportEquals {
		return true
	}

	// Match eslint-plugin-import's TypeScript export-assignment visitor:
	// `export = namespace` gets a synthetic default only under esModuleInterop,
	// while non-namespace local declarations and re-export-like expressions do.
	name, ok := exportAssignmentReferencedName(exportAssignment.Expression)
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
	return compilerOptionsESModuleInterop(ctx)
}

// The tsgo shim exposes CompilerOptions fields but not GetESModuleInterop.
func compilerOptionsESModuleInterop(ctx rule.RuleContext) bool {
	if ctx.Program == nil || ctx.Program.Options() == nil {
		return false
	}
	options := ctx.Program.Options()
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

func exportAssignmentReferencedName(expr *ast.Node) (string, bool) {
	expr = ast.SkipParentheses(expr)
	if expr == nil {
		return "", false
	}
	switch expr.Kind {
	case ast.KindIdentifier:
		return expr.AsIdentifier().Text, true
	case ast.KindFunctionExpression, ast.KindClassExpression:
		name := expr.Name()
		if name != nil {
			return name.Text(), true
		}
	}
	return "", false
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

func exportDeclarationHasName(ctx rule.RuleContext, sourceFile *ast.SourceFile, exportDecl *ast.ExportDeclaration, exportName string, seen map[exportKey]bool) (bool, bool) {
	if exportDecl == nil {
		return false, false
	}

	if exportDecl.ExportClause == nil {
		if exportDecl.ModuleSpecifier == nil || exportName == defaultExportName {
			return false, false
		}
		found, ok := hasExport(ctx, sourceFile, exportDecl.ModuleSpecifier, exportName, seen)
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

			hasName, ok := hasExport(ctx, sourceFile, exportDecl.ModuleSpecifier, localName, seen)
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

func moduleExportName(node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	return rslint_utils.GetStaticPropertyName(node)
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

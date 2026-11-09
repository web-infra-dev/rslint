package prefer_export_from

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/modules"
)

func reexportFixes(ctx rule.RuleContext, match reexport, directExports []*ast.ExportDeclaration, exportedText string) []rule.RuleFix {
	file := ctx.SourceFile
	importDecl := match.imported.DefNode.AsImportDeclaration()
	exportStmt := match.statement()
	// Moving a declaration to the end of the file must not move it out of a
	// TypeScript module block or turn a deferred import into an eager export.
	if importDecl.Parent != file.AsNode() || exportStmt.Parent != file.AsNode() ||
		importDecl.ImportClause.AsImportClause().PhaseModifier == ast.KindDeferKeyword {
		return nil
	}
	comments := ctx.Comments.All()
	for _, node := range []*ast.Node{importDecl.AsNode(), exportStmt} {
		span := utils.TrimNodeTextRange(file, node)
		if utils.HasCommentInSpan(comments, span.Pos(), span.End()) {
			return nil
		}
	}

	sourceStart := utils.TrimNodeTextRange(file, importDecl.ModuleSpecifier).Pos()
	from, ok := utils.TokenBeforePosition(file, sourceStart)
	if !ok || from.Kind != ast.KindFromKeyword {
		return nil
	}
	typeOnly := match.typeOnly || match.imported.IsTypeOnlyImport
	// Under verbatimModuleSyntax, a lone inline type import still loads the
	// module. A type-only re-export can omit that request entirely.
	if match.removeImport && match.importCount == 1 && match.imported.IsTypeOnlyImport && !importDecl.ImportClause.IsTypeOnly() {
		if options := ctx.Program().Options(); options != nil && options.VerbatimModuleSyntax.IsTrue() {
			match.removeImport = false
		}
	}
	var target *ast.ExportDeclaration
	var prefix, specifierText string
	specifier := match.imported.ID.Parent
	if specifier.Kind == ast.KindNamespaceImport {
		prefix = "export * as "
		if typeOnly {
			prefix = "export type * as "
		}
		prefix += exportedText
	} else {
		var importedName *ast.Node
		if specifier.Kind == ast.KindImportSpecifier {
			importedName = specifier.PropertyNameOrName()
		}
		specifierText = exportedText
		if !sameModuleName(importedName, match.name) {
			specifierText = nameText(file, importedName) + " as " + exportedText
		}
		target = reusableExport(importDecl, directExports, typeOnly)
		if typeOnly && (target == nil || !target.IsTypeOnly) {
			specifierText = "type " + specifierText
		}
		prefix = "export {" + specifierText + "}"
	}

	// Keep the first request in place when moving the entire runtime import
	// would cross another dependency and could change evaluation order.
	inPlace := match.removeImport && match.importCount == 1 && !typeOnly && crossesModuleRequest(file, importDecl, target)
	var insertion rule.RuleFix
	switch {
	case inPlace:
		insertion = rule.RuleFixReplace(file, importDecl.AsNode(), prefix+" "+file.Text()[from.Start:importDecl.End()])
	case target == nil:
		insertion = appendExport(file, importDecl, from.Start, prefix)
	default:
		clause := target.ExportClause
		elements := clause.Elements()
		if len(elements) != 0 {
			insertion = rule.RuleFixInsertAfter(elements[len(elements)-1], ", "+specifierText)
		} else {
			brace := utils.TrimNodeTextRange(file, clause).Pos()
			insertion = rule.RuleFixReplaceRange(core.NewTextRange(brace+1, brace+1), specifierText)
		}
	}
	fixes := []rule.RuleFix{insertion}
	if match.removeImport && !inPlace {
		if match.importCount == 1 {
			fixes = append(fixes, removeDeclaration(file, importDecl.AsNode()))
		} else if specifier.Kind == ast.KindNamespaceImport || specifier.Kind == ast.KindImportSpecifier &&
			len(specifier.Parent.Elements()) == 1 {
			// Removing the last named or namespace specifier leaves the default
			// binding, so remove its separating comma along with the group.
			defaultName := importDecl.ImportClause.Name()
			if defaultName == nil {
				return nil
			}
			comma, ok := utils.TokenAtOrAfter(file, defaultName.End())
			if !ok || comma.Kind != ast.KindCommaToken {
				return nil
			}
			space := ""
			if defaultName.End() == comma.Start {
				space = " "
			}
			fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(comma.Start, from.Start), space))
		} else {
			if specifier.Kind == ast.KindImportClause {
				specifier = match.imported.ID
			}
			fixes = removeSpecifier(file, fixes, specifier)
		}
	}
	if match.node.Kind == ast.KindExportSpecifier && len(match.node.Parent.Elements()) > 1 {
		return removeSpecifier(file, fixes, match.node)
	}
	return append(fixes, removeDeclaration(file, exportStmt))
}

func crossesModuleRequest(file *ast.SourceFile, imported *ast.ImportDeclaration, target *ast.ExportDeclaration) bool {
	destination := len(file.Text())
	if target != nil {
		destination = target.Pos()
	}
	start, end := min(imported.Pos(), destination), max(imported.Pos(), destination)
	attributes := modules.SourceFromSpecifier(imported.ModuleSpecifier).Attributes()
	for _, source := range modules.Collect(file, modules.ESModuleReferences) {
		decl := source.Declaration()
		if decl.Parent != imported.Parent || decl.Pos() <= start || decl.Pos() >= end ||
			source.Kind() == modules.ModuleReferenceDynamicImport || ast.IsExclusivelyTypeOnlyImportOrExport(decl) {
			continue
		}
		if source.Specifier().Text() != imported.ModuleSpecifier.Text() || source.Attributes().Key() != attributes.Key() {
			return true
		}
	}
	return false
}

func sameModuleName(imported, exported *ast.Node) bool {
	if isDefaultName(imported) {
		return isDefaultName(exported)
	}
	return exported != nil && imported.Kind == exported.Kind && imported.Text() == exported.Text()
}

func reusableExport(importDecl *ast.ImportDeclaration, directExports []*ast.ExportDeclaration, typeOnly bool) *ast.ExportDeclaration {
	attributes := modules.SourceFromSpecifier(importDecl.ModuleSpecifier).Attributes()
	if attributes.State == modules.AttributesInvalid {
		return nil
	}
	var valueExport *ast.ExportDeclaration
	for _, target := range directExports {
		if target.Parent != importDecl.Parent || target.ModuleSpecifier.Text() != importDecl.ModuleSpecifier.Text() ||
			target.IsTypeOnly && !typeOnly {
			continue
		}
		// The source string alone does not identify a module request when the
		// import or export has attributes. Retain each request's attributes.
		other := modules.SourceFromSpecifier(target.ModuleSpecifier).Attributes()
		if other.State == modules.AttributesInvalid || other.Key() != attributes.Key() {
			continue
		}
		if target.IsTypeOnly || !typeOnly {
			return target
		}
		if valueExport == nil {
			valueExport = target
		}
	}
	return valueExport
}

func appendExport(file *ast.SourceFile, importDecl *ast.ImportDeclaration, from int, prefix string) rule.RuleFix {
	text := file.Text()
	linebreak := "\n"
	if lines := file.ECMALineMap(); len(lines) > 1 {
		next := int(lines[1])
		linebreak = text[utils.LineContentEnd(text, next):next]
	}
	return rule.RuleFixReplaceRange(core.NewTextRange(len(text), len(text)), linebreak+prefix+" "+text[from:importDecl.End()])
}

func removeSpecifier(file *ast.SourceFile, fixes []rule.RuleFix, specifier *ast.Node) []rule.RuleFix {
	fixes = append(fixes, rule.RuleFixRemove(file, specifier))
	if comma, ok := utils.TokenAtOrAfter(file, specifier.End()); ok && comma.Kind == ast.KindCommaToken {
		fixes = append(fixes, rule.RuleFixRemoveRange(comma.Range()))
	}
	return fixes
}

func removeDeclaration(file *ast.SourceFile, node *ast.Node) rule.RuleFix {
	span := utils.TrimNodeTextRange(file, node)
	replacement := ""
	// The adjacent declaration may also be removed. Protect the preceding
	// statement against a continuation token even if it is not yet adjacent.
	if previous, ok := utils.TokenBeforePosition(file, span.Pos()); ok && unicornutil.NeedsSemicolonAfter(file, previous, "(", false) {
		replacement = ";"
	}
	return rule.RuleFixReplaceRange(span, replacement)
}

package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type ModuleReferenceOptions struct {
	ESModule bool
	CommonJS bool
	AMD      bool
}

type ModuleReference struct {
	Source       *ast.Node
	Importer     *ast.Node
	SourceFile   *ast.SourceFile
	Target       *ast.SourceFile
	ResolvedPath string
	Specifier    string
	Dynamic      bool
	OnlyTypes    bool
}

// CollectModuleReferences returns sourceFile's module references in source
// order, each with its target resolved. Rules that need this for more than
// the file they are linting should go through [ModuleIndex] instead, which
// answers the same question once per Program.
func CollectModuleReferences(ctx rule.RuleContext, sourceFile *ast.SourceFile, options ModuleReferenceOptions) []ModuleReference {
	return collectModuleReferences(ctx, sourceFile, options, compileModuleSettings(ctx.Settings))
}

func collectModuleReferences(ctx rule.RuleContext, sourceFile *ast.SourceFile, options ModuleReferenceOptions, settings *ModuleSettings) []ModuleReference {
	if ctx.Program == nil || sourceFile == nil || (!options.ESModule && !options.CommonJS && !options.AMD) {
		return nil
	}

	// SourceFile.Imports is populated by the parser and avoids walking every
	// AST node in the usual static-ESM case. The generic collector remains
	// necessary for call-based edges, parser recovery, and imports inside
	// module bodies.
	if options.CommonJS || options.AMD || needsFullModuleScan(sourceFile) {
		return collectModuleReferencesByWalk(ctx, sourceFile, options, settings)
	}
	return collectStaticImports(ctx, sourceFile, settings)
}

// collectStaticImports reads the module specifiers the parser already
// recorded. It handles the shapes those specifiers can take in a file with no
// dynamic import, no module declaration and no parse error, which is what
// needsFullModuleScan checks for.
func collectStaticImports(ctx rule.RuleContext, sourceFile *ast.SourceFile, settings *ModuleSettings) []ModuleReference {
	imports := sourceFile.Imports()
	refs := make([]ModuleReference, 0, len(imports))
	for _, source := range imports {
		importer := ast.TryGetImportFromModuleSpecifier(source)
		if importer == nil {
			continue
		}

		onlyTypes := false
		switch importer.Kind {
		case ast.KindImportDeclaration, ast.KindJSImportDeclaration:
			onlyTypes = importDeclarationOnlyImportsTypes(importer.AsImportDeclaration())
		case ast.KindExportDeclaration:
			onlyTypes = ast.IsTypeOnlyImportOrExportDeclaration(importer)
		default:
			continue
		}

		ref := ModuleReference{
			Source:     source,
			Importer:   importer,
			SourceFile: sourceFile,
			Specifier:  source.Text(),
			OnlyTypes:  onlyTypes,
		}
		ref.ResolvedPath, ref.Target, _ = ResolveSourceFileFromSourceFile(ctx, sourceFile, source)
		if ref.Target != nil && settings.IsIgnoredPath(ref.Target.FileName()) {
			continue
		}
		refs = append(refs, ref)
	}
	return refs
}

// needsFullModuleScan reports whether sourceFile can hold module specifiers
// that are not reachable from the parser's own list in the shapes
// collectStaticImports understands.
func needsFullModuleScan(sourceFile *ast.SourceFile) bool {
	if sourceFile.Flags&ast.NodeFlagsPossiblyContainsDynamicImport != 0 || len(sourceFile.Diagnostics()) != 0 {
		return true
	}
	for _, statement := range sourceFile.Statements.Nodes {
		if statement != nil && statement.Kind == ast.KindModuleDeclaration {
			return true
		}
	}
	return false
}

func collectModuleReferencesByWalk(ctx rule.RuleContext, sourceFile *ast.SourceFile, options ModuleReferenceOptions, settings *ModuleSettings) []ModuleReference {
	var refs []ModuleReference
	visitDescendants(sourceFile.AsNode(), func(node *ast.Node) {
		if node == nil {
			return
		}

		switch node.Kind {
		case ast.KindImportDeclaration, ast.KindJSImportDeclaration:
			if !options.ESModule {
				return
			}
			importDecl := node.AsImportDeclaration()
			addModuleReference(ctx, sourceFile, &refs, settings, importDecl.ModuleSpecifier, node, false, importDeclarationOnlyImportsTypes(importDecl))
		case ast.KindExportDeclaration:
			if !options.ESModule {
				return
			}
			exportDecl := node.AsExportDeclaration()
			// tsgo matches eslint-plugin-import here: only `export type * from`
			// is exclusively type-only; named type re-exports still stay graph edges.
			addModuleReference(ctx, sourceFile, &refs, settings, exportDecl.ModuleSpecifier, node, false, ast.IsTypeOnlyImportOrExportDeclaration(node))
		case ast.KindCallExpression:
			collectCallModuleReferences(ctx, sourceFile, &refs, settings, node.AsCallExpression(), options)
		}
	})

	return refs
}

func collectCallModuleReferences(ctx rule.RuleContext, sourceFile *ast.SourceFile, refs *[]ModuleReference, settings *ModuleSettings, call *ast.CallExpression, options ModuleReferenceOptions) {
	if call == nil {
		return
	}

	callee := ast.SkipParentheses(call.Expression)
	if callee == nil {
		return
	}

	if options.ESModule && callee.Kind == ast.KindImportKeyword {
		if len(call.Arguments.Nodes) == 0 {
			return
		}
		addModuleReference(ctx, sourceFile, refs, settings, ast.SkipParentheses(call.Arguments.Nodes[0]), call.AsNode(), true, false)
		return
	}

	if callee.Kind != ast.KindIdentifier {
		return
	}

	calleeName := callee.AsIdentifier().Text
	if options.CommonJS && ast.IsRequireCall(call.AsNode(), false) {
		arg := ast.SkipParentheses(call.Arguments.Nodes[0])
		if arg != nil && ast.IsStringLiteralLike(arg) {
			addModuleReference(ctx, sourceFile, refs, settings, arg, call.AsNode(), false, false)
		}
		return
	}

	if options.AMD && (calleeName == "require" || calleeName == "define") {
		if len(call.Arguments.Nodes) == 0 {
			return
		}
		arg := ast.SkipParentheses(call.Arguments.Nodes[0])
		if arg == nil || arg.Kind != ast.KindArrayLiteralExpression {
			return
		}
		for _, element := range arg.AsArrayLiteralExpression().Elements.Nodes {
			element = ast.SkipParentheses(element)
			if element == nil || !ast.IsStringLiteralLike(element) {
				continue
			}
			addModuleReference(ctx, sourceFile, refs, settings, element, call.AsNode(), false, false)
		}
	}
}

func addModuleReference(ctx rule.RuleContext, sourceFile *ast.SourceFile, refs *[]ModuleReference, settings *ModuleSettings, source *ast.Node, importer *ast.Node, dynamic bool, onlyTypes bool) {
	if source == nil {
		return
	}
	source = ast.SkipParentheses(source)
	if source == nil || !ast.IsStringLiteralLike(source) {
		return
	}

	ref := ModuleReference{
		Source:     source,
		Importer:   importer,
		SourceFile: sourceFile,
		Specifier:  source.Text(),
		Dynamic:    dynamic,
		OnlyTypes:  onlyTypes,
	}

	ref.ResolvedPath, ref.Target, _ = ResolveSourceFileFromSourceFile(ctx, sourceFile, source)
	if ref.Target != nil && settings.IsIgnoredPath(ref.Target.FileName()) {
		return
	}

	*refs = append(*refs, ref)
}

func importDeclarationOnlyImportsTypes(importDecl *ast.ImportDeclaration) bool {
	if importDecl == nil || importDecl.ImportClause == nil {
		return false
	}

	importClause := importDecl.ImportClause
	if importClause.IsTypeOnly() {
		return true
	}

	clause := importClause.AsImportClause()
	if clause == nil || clause.Name() != nil || clause.NamedBindings == nil {
		return false
	}

	if clause.NamedBindings.Kind != ast.KindNamedImports {
		return false
	}
	namedImports := clause.NamedBindings.AsNamedImports()
	if namedImports == nil || namedImports.Elements == nil || len(namedImports.Elements.Nodes) == 0 {
		return false
	}

	for _, specifier := range namedImports.Elements.Nodes {
		if specifier == nil || specifier.Kind != ast.KindImportSpecifier || !ast.IsTypeOnlyImportDeclaration(specifier) {
			return false
		}
	}
	return true
}

func visitDescendants(node *ast.Node, visit func(*ast.Node)) {
	if node == nil {
		return
	}
	visit(node)
	node.ForEachChild(func(child *ast.Node) bool {
		visitDescendants(child, visit)
		return false
	})
}

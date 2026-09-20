package modules

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// ReferenceKind names the syntax a module reference was written in.
type ReferenceKind uint8

const (
	// ModuleReferenceImport is an import declaration, in TypeScript or JavaScript.
	ModuleReferenceImport ReferenceKind = iota
	// ModuleReferenceExport is an export declaration carrying a module specifier.
	ModuleReferenceExport
	// ModuleReferenceDynamicImport is an `import()` call.
	ModuleReferenceDynamicImport
	// ModuleReferenceRequire is a `require()` call.
	ModuleReferenceRequire
	// ModuleReferenceAMD is one entry of a `define([…])` or `require([…])` list.
	ModuleReferenceAMD
)

// ReferenceKinds is a set of module-reference syntaxes.
type ReferenceKinds uint8

const (
	ESModuleReferences ReferenceKinds = 1<<ModuleReferenceImport |
		1<<ModuleReferenceExport |
		1<<ModuleReferenceDynamicImport
	CommonJSReferences  ReferenceKinds = 1 << ModuleReferenceRequire
	AMDReferences       ReferenceKinds = 1 << ModuleReferenceAMD
	AllModuleReferences                = ESModuleReferences | CommonJSReferences | AMDReferences
)

func (kinds ReferenceKinds) includes(kind ReferenceKind) bool {
	return kinds&(1<<kind) != 0
}

// Source is a module source expression and its syntax context. Specifier preserves its
// original shape and can be a dynamic expression, template or TypeScript wrapper.
// TypeOnly follows the module graph's emitted-dependency semantics; callers
// that distinguish explicit import/export type keywords inspect Declaration.
// Collection performs no name resolution, filesystem access or evaluation.
type Source struct {
	Specifier   *ast.Node
	Declaration *ast.Node
	Kind        ReferenceKind
	TypeOnly    bool
}

func collectSpecifiers(file *ast.SourceFile, kinds ReferenceKinds) []Source {
	// SourceFile.Imports is populated by the parser and avoids walking every
	// AST node in the usual static-ESM case. The generic collector remains
	// necessary for call-based references, parser recovery, and imports
	// inside module bodies.
	if kinds&(CommonJSReferences|AMDReferences) != 0 || needsFullModuleScan(file) {
		return collectByWalk(file, kinds)
	}
	return collectStaticImports(file, kinds)
}

// collectStaticImports reads the module specifiers the parser already
// recorded. It handles the shapes those specifiers can take in a file with no
// dynamic import, no module declaration and no parse error, which is what
// needsFullModuleScan checks for.
func collectStaticImports(file *ast.SourceFile, kinds ReferenceKinds) []Source {
	imports := file.Imports()
	specifiers := make([]Source, 0, len(imports))
	for _, specifier := range imports {
		declaration := ast.TryGetImportFromModuleSpecifier(specifier)
		if declaration == nil {
			continue
		}

		var kind ReferenceKind
		typeOnly := false
		switch declaration.Kind {
		case ast.KindImportDeclaration, ast.KindJSImportDeclaration:
			kind = ModuleReferenceImport
			typeOnly = importDeclarationOnlyImportsTypes(declaration.AsImportDeclaration())
		case ast.KindExportDeclaration:
			kind = ModuleReferenceExport
			typeOnly = ast.IsTypeOnlyImportOrExportDeclaration(declaration)
		default:
			continue
		}

		if kinds.includes(kind) {
			specifiers = append(specifiers, Source{
				Specifier:   specifier,
				Declaration: declaration,
				Kind:        kind,
				TypeOnly:    typeOnly,
			})
		}
	}
	return specifiers
}

// needsFullModuleScan reports whether file can hold module specifiers that are
// not reachable from the parser's own list in the shapes collectStaticImports
// understands.
func needsFullModuleScan(file *ast.SourceFile) bool {
	if file.Flags&ast.NodeFlagsPossiblyContainsDynamicImport != 0 || len(file.Diagnostics()) != 0 {
		return true
	}
	for _, statement := range file.Statements.Nodes {
		if statement != nil && statement.Kind == ast.KindModuleDeclaration {
			return true
		}
		// The parser's resolution list omits empty module names. Rules
		// still need their source nodes to diagnose imports and re-exports.
		if statement != nil && ast.IsAnyImportOrReExport(statement) {
			if specifier := ast.GetExternalModuleName(statement); specifier != nil && specifier.Text() == "" {
				return true
			}
		}
	}
	return false
}

func collectByWalk(file *ast.SourceFile, kinds ReferenceKinds) []Source {
	var specifiers []Source
	// A module specifier can appear anywhere a call can, so every subtree is
	// walked; no node accounts for its own children here.
	visitModuleReferenceNodes(file.AsNode(), func(node *ast.Node) bool {
		switch node.Kind {
		case ast.KindImportDeclaration, ast.KindJSImportDeclaration:
			if !kinds.includes(ModuleReferenceImport) {
				return true
			}
			importDecl := node.AsImportDeclaration()
			appendSpecifier(&specifiers, importDecl.ModuleSpecifier, node, ModuleReferenceImport, importDeclarationOnlyImportsTypes(importDecl))
		case ast.KindExportDeclaration:
			if !kinds.includes(ModuleReferenceExport) {
				return true
			}
			exportDecl := node.AsExportDeclaration()
			// tsgo matches eslint-plugin-import here: only `export type * from`
			// is exclusively type-only; named type re-exports stay references.
			appendSpecifier(&specifiers, exportDecl.ModuleSpecifier, node, ModuleReferenceExport, ast.IsTypeOnlyImportOrExportDeclaration(node))
		case ast.KindCallExpression:
			appendCallSpecifiers(&specifiers, node.AsCallExpression(), kinds)
		}
		return true
	})
	return specifiers
}

func visitModuleReferenceNodes(node *ast.Node, visit func(*ast.Node) bool) {
	if node == nil || !visit(node) {
		return
	}
	node.ForEachChild(func(child *ast.Node) bool {
		visitModuleReferenceNodes(child, visit)
		return false
	})
}

func appendCallSpecifiers(specifiers *[]Source, call *ast.CallExpression, kinds ReferenceKinds) {
	if call == nil || call.Arguments == nil {
		return
	}

	callee := ast.SkipParentheses(call.Expression)
	if callee == nil {
		return
	}

	if kinds.includes(ModuleReferenceDynamicImport) && callee.Kind == ast.KindImportKeyword {
		if len(call.Arguments.Nodes) == 0 {
			return
		}
		appendSpecifier(specifiers, call.Arguments.Nodes[0], call.AsNode(), ModuleReferenceDynamicImport, false)
		return
	}

	if callee.Kind != ast.KindIdentifier {
		return
	}

	calleeName := callee.AsIdentifier().Text
	if kinds.includes(ModuleReferenceRequire) && ast.IsRequireCall(call.AsNode(), false) {
		appendSpecifier(specifiers, call.Arguments.Nodes[0], call.AsNode(), ModuleReferenceRequire, false)
		return
	}

	if kinds.includes(ModuleReferenceAMD) && (calleeName == "require" || calleeName == "define") {
		if len(call.Arguments.Nodes) == 0 {
			return
		}
		arg := ast.SkipParentheses(call.Arguments.Nodes[0])
		if arg == nil || arg.Kind != ast.KindArrayLiteralExpression {
			return
		}
		for _, element := range arg.AsArrayLiteralExpression().Elements.Nodes {
			appendSpecifier(specifiers, element, call.AsNode(), ModuleReferenceAMD, false)
		}
	}
}

func appendSpecifier(specifiers *[]Source, specifier *ast.Node, declaration *ast.Node, kind ReferenceKind, typeOnly bool) {
	if specifier == nil {
		return
	}
	*specifiers = append(*specifiers, Source{
		Specifier:   specifier,
		Declaration: declaration,
		Kind:        kind,
		TypeOnly:    typeOnly,
	})
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

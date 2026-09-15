package program

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// ModuleReferenceKind names the syntax a module reference was written in.
type ModuleReferenceKind uint8

const (
	// ModuleReferenceImport is an import declaration, in TypeScript or JavaScript.
	ModuleReferenceImport ModuleReferenceKind = iota
	// ModuleReferenceExport is an export declaration carrying a module specifier.
	ModuleReferenceExport
	// ModuleReferenceDynamicImport is an `import()` call.
	ModuleReferenceDynamicImport
	// ModuleReferenceRequire is a `require()` call.
	ModuleReferenceRequire
	// ModuleReferenceAMD is one entry of a `define([…])` or `require([…])` list.
	ModuleReferenceAMD
)

// ModuleReferenceKinds is a set of module-reference syntaxes.
type ModuleReferenceKinds uint8

const (
	ESModuleReferences ModuleReferenceKinds = 1<<ModuleReferenceImport |
		1<<ModuleReferenceExport |
		1<<ModuleReferenceDynamicImport
	CommonJSReferences  ModuleReferenceKinds = 1 << ModuleReferenceRequire
	AMDReferences       ModuleReferenceKinds = 1 << ModuleReferenceAMD
	AllModuleReferences                      = ESModuleReferences | CommonJSReferences | AMDReferences
)

func (kinds ModuleReferenceKinds) includes(kind ModuleReferenceKind) bool {
	return kinds&(1<<kind) != 0
}

// ModuleReference is one module specifier written in a file, with the file it
// names already resolved.
type ModuleReference struct {
	// Specifier is the string literal holding the module name.
	Specifier *ast.Node
	// Declaration is the syntax the specifier belongs to: the import or
	// export declaration, or the call expression.
	Declaration *ast.Node
	// From is the file the specifier was written in.
	From *ast.SourceFile
	// Target is the file it names, nil when nothing in the Program answers
	// for it. A specifier can resolve to a path the Program never materialized, in
	// which case ResolvedPath is set and Target is not.
	Target *ast.SourceFile
	// ResolvedPath is the path the specifier resolves to, empty when it
	// resolves nowhere.
	ResolvedPath string
	Kind         ModuleReferenceKind
	// TypeOnly reports that the syntax cannot survive into emitted
	// JavaScript: `import type`, `export type *`, and a named import clause
	// whose every specifier is type-only.
	TypeOnly bool
}

// Text is the module name as written.
func (reference ModuleReference) Text() string {
	if reference.Specifier == nil {
		return ""
	}
	return reference.Specifier.Text()
}

// Path is the file the reference names: the runtime's file for it when there
// is one, and otherwise the path it resolved to. Empty when it resolves
// nowhere.
func (reference ModuleReference) Path() string {
	if reference.Target != nil {
		return reference.Target.FileName()
	}
	return reference.ResolvedPath
}

// Dynamic reports whether the reference is an `import()` call, which defers
// loading rather than requiring the module up front.
func (reference ModuleReference) Dynamic() bool {
	return reference.Kind == ModuleReferenceDynamicImport
}

// ModuleGraph answers which modules each file of one Program references and
// what they resolve to. Resolved edges are derived once per immutable Program
// generation however many rules, files, or lint passes ask for them. Their
// syntax-only half follows an exact SourceFile across editor Programs; resolved
// targets always remain local to the Program generation.
//
// It reports what the syntax says and nothing more. Which of these references
// a rule treats as a dependency — whether type-only imports count, whether
// anything under node_modules counts — is the rule's own question.
type ModuleGraph struct {
	program *Program
}

// moduleReferencesCacheKey pairs a file with the syntaxes the caller asked about, which
// is what decides the answer.
type moduleReferencesCacheKey struct {
	file  *ast.SourceFile
	kinds ModuleReferenceKinds
}

// ModuleGraph returns a lightweight view over module references in p. The
// view owns no source state: identity, resolution and cache lifetime all remain
// properties of the Program generation.
func (p *Program) ModuleGraph() ModuleGraph {
	if !p.IsValid() {
		return ModuleGraph{}
	}
	return ModuleGraph{program: p}
}

// Files returns every file of the source set in its stable input order. A
// file's position in this slice is stable for the lifetime of the graph, so
// callers that need a dense numbering can adopt it. The result is read-only.
func (graph ModuleGraph) Files() []*ast.SourceFile {
	if graph.program == nil {
		return nil
	}
	return graph.program.SourceFiles()
}

// References returns the module references file writes in the selected
// syntaxes, in source order. The result is shared with every other caller and
// must not be modified.
func (graph ModuleGraph) References(file *ast.SourceFile, kinds ModuleReferenceKinds) []ModuleReference {
	if graph.program == nil || !graph.program.OwnsSourceFile(file) || kinds == 0 {
		return nil
	}

	key := moduleReferencesCacheKey{file: file, kinds: kinds}
	return Cached(graph.program, key, func() []ModuleReference {
		return graph.resolveAll(file, cachedModuleSpecifiers(file, kinds))
	})
}

// resolveAll turns what a file writes into what it references, which is the
// half of the answer only this Program can give.
func (graph ModuleGraph) resolveAll(file *ast.SourceFile, specifiers []moduleSpecifier) []ModuleReference {
	if len(specifiers) == 0 {
		return nil
	}
	references := make([]ModuleReference, len(specifiers))
	for i := range specifiers {
		references[i] = ModuleReference{
			Specifier:   specifiers[i].specifier,
			Declaration: specifiers[i].declaration,
			From:        file,
			Kind:        specifiers[i].kind,
			TypeOnly:    specifiers[i].typeOnly,
		}
		references[i].ResolvedPath, references[i].Target, _ =
			graph.program.ResolveModule(file, specifiers[i].specifier)
	}
	return references
}

// moduleSpecifier is the half of a ModuleReference that a file's own syntax
// decides: which string literal names a module, what syntax it was written
// in, and whether that syntax survives into emitted JavaScript. What the
// specifier resolves to is deliberately absent — that is the Program's answer,
// and two Programs holding the same unchanged file can disagree about it.
type moduleSpecifier struct {
	specifier   *ast.Node
	declaration *ast.Node
	kind        ModuleReferenceKind
	typeOnly    bool
}

func collectSpecifiers(file *ast.SourceFile, kinds ModuleReferenceKinds) []moduleSpecifier {
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
func collectStaticImports(file *ast.SourceFile, kinds ModuleReferenceKinds) []moduleSpecifier {
	imports := file.Imports()
	specifiers := make([]moduleSpecifier, 0, len(imports))
	for _, specifier := range imports {
		declaration := ast.TryGetImportFromModuleSpecifier(specifier)
		if declaration == nil {
			continue
		}

		var kind ModuleReferenceKind
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
			specifiers = append(specifiers, moduleSpecifier{
				specifier:   specifier,
				declaration: declaration,
				kind:        kind,
				typeOnly:    typeOnly,
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
	}
	return false
}

func collectByWalk(file *ast.SourceFile, kinds ModuleReferenceKinds) []moduleSpecifier {
	var specifiers []moduleSpecifier
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

func appendCallSpecifiers(specifiers *[]moduleSpecifier, call *ast.CallExpression, kinds ModuleReferenceKinds) {
	if call == nil {
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
		appendSpecifier(specifiers, ast.SkipParentheses(call.Arguments.Nodes[0]), call.AsNode(), ModuleReferenceDynamicImport, false)
		return
	}

	if callee.Kind != ast.KindIdentifier {
		return
	}

	calleeName := callee.AsIdentifier().Text
	if kinds.includes(ModuleReferenceRequire) && ast.IsRequireCall(call.AsNode(), false) {
		arg := ast.SkipParentheses(call.Arguments.Nodes[0])
		if arg != nil && ast.IsStringLiteralLike(arg) {
			appendSpecifier(specifiers, arg, call.AsNode(), ModuleReferenceRequire, false)
		}
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
			element = ast.SkipParentheses(element)
			if element == nil || !ast.IsStringLiteralLike(element) {
				continue
			}
			appendSpecifier(specifiers, element, call.AsNode(), ModuleReferenceAMD, false)
		}
	}
}

func appendSpecifier(specifiers *[]moduleSpecifier, specifier *ast.Node, declaration *ast.Node, kind ModuleReferenceKind, typeOnly bool) {
	if specifier == nil {
		return
	}
	specifier = ast.SkipParentheses(specifier)
	if specifier == nil || !ast.IsStringLiteralLike(specifier) {
		return
	}
	*specifiers = append(*specifiers, moduleSpecifier{
		specifier:   specifier,
		declaration: declaration,
		kind:        kind,
		typeOnly:    typeOnly,
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

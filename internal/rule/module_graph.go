package rule

import (
	"sync"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// ModuleSyntax selects which spellings of a module reference count.
// JavaScript has three module systems and a rule cares about a different set
// depending on what it checks, so the caller says which it means.
type ModuleSyntax struct {
	ESModule bool
	CommonJS bool
	AMD      bool
}

func (syntax ModuleSyntax) none() bool {
	return !syntax.ESModule && !syntax.CommonJS && !syntax.AMD
}

// ModuleEdgeKind names the syntax a module reference was written in.
type ModuleEdgeKind uint8

const (
	// ModuleEdgeImport is an import declaration, in TypeScript or JavaScript.
	ModuleEdgeImport ModuleEdgeKind = iota
	// ModuleEdgeExport is an export declaration carrying a module specifier.
	ModuleEdgeExport
	// ModuleEdgeDynamicImport is an `import()` call.
	ModuleEdgeDynamicImport
	// ModuleEdgeRequire is a `require()` call.
	ModuleEdgeRequire
	// ModuleEdgeAMD is one entry of a `define([…])` or `require([…])` list.
	ModuleEdgeAMD
)

// ModuleEdge is one module specifier written in a file, with the file it
// names already resolved.
type ModuleEdge struct {
	// Specifier is the string literal holding the module name.
	Specifier *ast.Node
	// Declaration is the syntax the specifier belongs to: the import or
	// export declaration, or the call expression.
	Declaration *ast.Node
	// From is the file the specifier was written in.
	From *ast.SourceFile
	// Target is the file it names, nil when nothing in the Program answers
	// for it. A specifier can resolve to a path the Program never loaded, in
	// which case ResolvedPath is set and Target is not.
	Target *ast.SourceFile
	// ResolvedPath is the path the specifier resolves to, empty when it
	// resolves nowhere.
	ResolvedPath string
	Kind         ModuleEdgeKind
	// TypeOnly reports that the syntax cannot survive into emitted
	// JavaScript: `import type`, `export type *`, and a named import clause
	// whose every specifier is type-only.
	TypeOnly bool
}

// Text is the module name as written.
func (edge ModuleEdge) Text() string {
	if edge.Specifier == nil {
		return ""
	}
	return edge.Specifier.Text()
}

// Dynamic reports whether the reference is an `import()` call, which defers
// loading rather than requiring the module up front.
func (edge ModuleEdge) Dynamic() bool {
	return edge.Kind == ModuleEdgeDynamicImport
}

// ModuleGraph answers which modules each file of one Program references and
// what they resolve to. A file's answer depends only on its own syntax, so it
// is derived once per lint run however many rules and files ask for it — the
// alternative being that each rule re-reads and re-resolves the same imports
// for every file it looks past.
//
// It reports what the syntax says and nothing more. Which of these references
// a rule treats as a dependency — whether type-only imports count, whether
// anything under node_modules counts — is the rule's own question.
type ModuleGraph struct {
	program *compiler.Program

	mu    sync.Mutex
	edges map[moduleEdgeKey][]ModuleEdge
}

// moduleEdgeKey pairs a file with the syntaxes the caller asked about, which
// is what decides the answer.
type moduleEdgeKey struct {
	file   *ast.SourceFile
	syntax ModuleSyntax
}

func NewModuleGraph(program *compiler.Program) *ModuleGraph {
	return &ModuleGraph{
		program: program,
		edges:   make(map[moduleEdgeKey][]ModuleEdge),
	}
}

// Files returns every file of the Program, in the Program's own order. A
// file's position in this slice is stable for the lifetime of the graph, so
// callers that need a dense numbering of the Program can adopt it.
func (graph *ModuleGraph) Files() []*ast.SourceFile {
	if graph == nil || graph.program == nil {
		return nil
	}
	return graph.program.SourceFiles()
}

// Edges returns the module references file writes in the given syntaxes, in
// source order. The result is shared with every other caller and must not be
// modified.
func (graph *ModuleGraph) Edges(file *ast.SourceFile, syntax ModuleSyntax) []ModuleEdge {
	if graph == nil || graph.program == nil || file == nil || syntax.none() {
		return nil
	}

	key := moduleEdgeKey{file: file, syntax: syntax}
	graph.mu.Lock()
	edges, ok := graph.edges[key]
	graph.mu.Unlock()
	if ok {
		return edges
	}

	// Collected outside the lock: collection is pure, so two files racing on
	// their first request cost one redundant collection at worst, which is
	// cheaper than serializing every file in the run behind one mutex.
	edges = graph.collect(file, syntax)

	graph.mu.Lock()
	if existing, ok := graph.edges[key]; ok {
		edges = existing
	} else {
		graph.edges[key] = edges
	}
	graph.mu.Unlock()
	return edges
}

func (graph *ModuleGraph) collect(file *ast.SourceFile, syntax ModuleSyntax) []ModuleEdge {
	// SourceFile.Imports is populated by the parser and avoids walking every
	// AST node in the usual static-ESM case. The generic collector remains
	// necessary for call-based references, parser recovery, and imports
	// inside module bodies.
	if syntax.CommonJS || syntax.AMD || needsFullModuleScan(file) {
		return graph.collectByWalk(file, syntax)
	}
	return graph.collectStaticImports(file)
}

// collectStaticImports reads the module specifiers the parser already
// recorded. It handles the shapes those specifiers can take in a file with no
// dynamic import, no module declaration and no parse error, which is what
// needsFullModuleScan checks for.
func (graph *ModuleGraph) collectStaticImports(file *ast.SourceFile) []ModuleEdge {
	imports := file.Imports()
	edges := make([]ModuleEdge, 0, len(imports))
	for _, specifier := range imports {
		declaration := ast.TryGetImportFromModuleSpecifier(specifier)
		if declaration == nil {
			continue
		}

		var kind ModuleEdgeKind
		typeOnly := false
		switch declaration.Kind {
		case ast.KindImportDeclaration, ast.KindJSImportDeclaration:
			kind = ModuleEdgeImport
			typeOnly = importDeclarationOnlyImportsTypes(declaration.AsImportDeclaration())
		case ast.KindExportDeclaration:
			kind = ModuleEdgeExport
			typeOnly = ast.IsTypeOnlyImportOrExportDeclaration(declaration)
		default:
			continue
		}

		edges = append(edges, graph.newEdge(file, specifier, declaration, kind, typeOnly))
	}
	return edges
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

func (graph *ModuleGraph) collectByWalk(file *ast.SourceFile, syntax ModuleSyntax) []ModuleEdge {
	var edges []ModuleEdge
	// A module specifier can appear anywhere a call can, so every subtree is
	// walked; no node accounts for its own children here.
	utils.VisitDescendants(file.AsNode(), func(node *ast.Node) bool {
		switch node.Kind {
		case ast.KindImportDeclaration, ast.KindJSImportDeclaration:
			if !syntax.ESModule {
				return true
			}
			importDecl := node.AsImportDeclaration()
			graph.appendEdge(&edges, file, importDecl.ModuleSpecifier, node, ModuleEdgeImport, importDeclarationOnlyImportsTypes(importDecl))
		case ast.KindExportDeclaration:
			if !syntax.ESModule {
				return true
			}
			exportDecl := node.AsExportDeclaration()
			// tsgo matches eslint-plugin-import here: only `export type * from`
			// is exclusively type-only; named type re-exports still stay edges.
			graph.appendEdge(&edges, file, exportDecl.ModuleSpecifier, node, ModuleEdgeExport, ast.IsTypeOnlyImportOrExportDeclaration(node))
		case ast.KindCallExpression:
			graph.appendCallEdges(&edges, file, node.AsCallExpression(), syntax)
		}
		return true
	})
	return edges
}

func (graph *ModuleGraph) appendCallEdges(edges *[]ModuleEdge, file *ast.SourceFile, call *ast.CallExpression, syntax ModuleSyntax) {
	if call == nil {
		return
	}

	callee := ast.SkipParentheses(call.Expression)
	if callee == nil {
		return
	}

	if syntax.ESModule && callee.Kind == ast.KindImportKeyword {
		if len(call.Arguments.Nodes) == 0 {
			return
		}
		graph.appendEdge(edges, file, ast.SkipParentheses(call.Arguments.Nodes[0]), call.AsNode(), ModuleEdgeDynamicImport, false)
		return
	}

	if callee.Kind != ast.KindIdentifier {
		return
	}

	calleeName := callee.AsIdentifier().Text
	if syntax.CommonJS && ast.IsRequireCall(call.AsNode(), false) {
		arg := ast.SkipParentheses(call.Arguments.Nodes[0])
		if arg != nil && ast.IsStringLiteralLike(arg) {
			graph.appendEdge(edges, file, arg, call.AsNode(), ModuleEdgeRequire, false)
		}
		return
	}

	if syntax.AMD && (calleeName == "require" || calleeName == "define") {
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
			graph.appendEdge(edges, file, element, call.AsNode(), ModuleEdgeAMD, false)
		}
	}
}

func (graph *ModuleGraph) appendEdge(edges *[]ModuleEdge, file *ast.SourceFile, specifier *ast.Node, declaration *ast.Node, kind ModuleEdgeKind, typeOnly bool) {
	if specifier == nil {
		return
	}
	specifier = ast.SkipParentheses(specifier)
	if specifier == nil || !ast.IsStringLiteralLike(specifier) {
		return
	}
	*edges = append(*edges, graph.newEdge(file, specifier, declaration, kind, typeOnly))
}

func (graph *ModuleGraph) newEdge(file *ast.SourceFile, specifier *ast.Node, declaration *ast.Node, kind ModuleEdgeKind, typeOnly bool) ModuleEdge {
	edge := ModuleEdge{
		Specifier:   specifier,
		Declaration: declaration,
		From:        file,
		Kind:        kind,
		TypeOnly:    typeOnly,
	}
	edge.ResolvedPath, edge.Target, _ = utils.ResolveModuleFile(graph.program, file, specifier)
	return edge
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

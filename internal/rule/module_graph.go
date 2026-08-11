package rule

import (
	"reflect"

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
	// Target is the file it names, nil when nothing in the source runtime answers
	// for it. A specifier can resolve to a path the runtime never loaded, in
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

// Path is the file the reference names: the runtime's file for it when there
// is one, and otherwise the path it resolved to. Empty when it resolves
// nowhere.
func (edge ModuleEdge) Path() string {
	if edge.Target != nil {
		return edge.Target.FileName()
	}
	return edge.ResolvedPath
}

// Dynamic reports whether the reference is an `import()` call, which defers
// loading rather than requiring the module up front.
func (edge ModuleEdge) Dynamic() bool {
	return edge.Kind == ModuleEdgeDynamicImport
}

// ModuleGraph answers which modules each file of one source set references
// and what they resolve to. A file's answer is derived once per lint run
// however many rules and files ask for it. Its syntax-only half can outlive
// the run when the caller opts into SourceFile-owned reuse; resolution remains
// local to this source set's runtime.
//
// It reports what the syntax says and nothing more. Which of these references
// a rule treats as a dependency — whether type-only imports count, whether
// anything under node_modules counts — is the rule's own question.
type ModuleGraph struct {
	program *compiler.Program
	runtime utils.ModuleResolutionRuntime
	files   []*ast.SourceFile
	// Standalone source sets have no Program identity for CachedByProgram.
	// Their run-scoped derived structures share this cache instead.
	cache programCache
	// Collection is pure, so LazyMap's build-outside-the-lock contract holds:
	// two files racing on their first request for one key cost one redundant
	// collection at worst, which is cheaper than serializing every file in the
	// run behind one mutex.
	edges utils.LazyMap[moduleEdgeKey, []ModuleEdge]
	// cacheModuleSpecifiers attaches the syntactic half to each immutable
	// SourceFile. Resolution remains local to this graph's Program.
	cacheModuleSpecifiers bool
}

// sourceRuntime returns the graph's standalone source environment.
// Standalone rule contexts use the graph as their single shared runtime
// carrier, avoiding one interface field in every per-rule context.
func (graph *ModuleGraph) sourceRuntime() SourceRuntime {
	if graph == nil || graph.runtime == nil {
		return nil
	}
	if runtime, ok := graph.runtime.(SourceRuntime); ok {
		return runtime
	}
	return nil
}

// moduleEdgeKey pairs a file with the syntaxes the caller asked about, which
// is what decides the answer.
type moduleEdgeKey struct {
	file   *ast.SourceFile
	syntax ModuleSyntax
}

func NewModuleGraph(program *compiler.Program) *ModuleGraph {
	if program == nil {
		return &ModuleGraph{}
	}
	// The Program owns an immutable source-file slice for the graph's lifetime,
	// so retain that owner instead of copying every file pointer per lint pass.
	return &ModuleGraph{program: program, runtime: program}
}

// NewCachedModuleGraph returns a graph that shares its syntax-only collection
// with other graphs holding the exact same SourceFile objects. Each graph
// still resolves those specifiers against its own source runtime.
func NewCachedModuleGraph(program *compiler.Program) *ModuleGraph {
	if program == nil {
		return &ModuleGraph{}
	}
	return &ModuleGraph{program: program, runtime: program, cacheModuleSpecifiers: true}
}

// NewStandaloneModuleGraph returns a run-scoped graph over one standalone
// source set. The caller owns files and runtime for at least as long as the
// graph; the copied slice fixes the graph's stable file order.
func NewStandaloneModuleGraph(files []*ast.SourceFile, runtime SourceRuntime) *ModuleGraph {
	return newStandaloneModuleGraph(files, runtime, true)
}

// NewOwnedStandaloneModuleGraph avoids a second pointer-slice copy when the
// caller owns an immutable files slice for the graph's full lifetime. Prefer
// NewStandaloneModuleGraph for caller-owned or mutable input.
func NewOwnedStandaloneModuleGraph(files []*ast.SourceFile, runtime SourceRuntime) *ModuleGraph {
	return newStandaloneModuleGraph(files, runtime, false)
}

func newStandaloneModuleGraph(files []*ast.SourceFile, runtime SourceRuntime, copyFiles bool) *ModuleGraph {
	if isNilSourceRuntime(runtime) {
		return &ModuleGraph{}
	}
	if copyFiles {
		files = append([]*ast.SourceFile(nil), files...)
	}
	return &ModuleGraph{
		runtime: runtime,
		files:   files,
	}
}

func isNilSourceRuntime(runtime SourceRuntime) bool {
	if runtime == nil {
		return true
	}
	value := reflect.ValueOf(runtime)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

// Files returns every file of the source set in its stable input order. A
// file's position in this slice is stable for the lifetime of the graph, so
// callers that need a dense numbering can adopt it. The result is read-only.
func (graph *ModuleGraph) Files() []*ast.SourceFile {
	if graph == nil || graph.runtime == nil {
		return nil
	}
	if graph.program != nil {
		return graph.program.SourceFiles()
	}
	return graph.files
}

// Edges returns the module references file writes in the given syntaxes, in
// source order. The result is shared with every other caller and must not be
// modified.
func (graph *ModuleGraph) Edges(file *ast.SourceFile, syntax ModuleSyntax) []ModuleEdge {
	if graph == nil || graph.runtime == nil || file == nil || syntax.none() {
		return nil
	}

	key := moduleEdgeKey{file: file, syntax: syntax}
	return graph.edges.Get(key, func() []ModuleEdge {
		return graph.resolveAll(file, graph.specifiersOf(file, syntax))
	})
}

// specifiersOf returns what file writes. Collection is a pure function of the
// file's own syntax, so an attached answer is valid for that SourceFile's
// entire lifetime.
func (graph *ModuleGraph) specifiersOf(file *ast.SourceFile, syntax ModuleSyntax) []moduleSpecifier {
	if !graph.cacheModuleSpecifiers {
		return collectSpecifiers(file, syntax)
	}
	return cachedModuleSpecifiers(file, syntax)
}

// resolveAll turns what a file writes into what it references, which is the
// half of the answer only this Program can give.
func (graph *ModuleGraph) resolveAll(file *ast.SourceFile, specifiers []moduleSpecifier) []ModuleEdge {
	if len(specifiers) == 0 {
		return nil
	}
	edges := make([]ModuleEdge, len(specifiers))
	for i := range specifiers {
		edges[i] = ModuleEdge{
			Specifier:   specifiers[i].specifier,
			Declaration: specifiers[i].declaration,
			From:        file,
			Kind:        specifiers[i].kind,
			TypeOnly:    specifiers[i].typeOnly,
		}
		edges[i].ResolvedPath, edges[i].Target, _ =
			utils.ResolveModuleFile(graph.runtime, file, specifiers[i].specifier)
	}
	return edges
}

// moduleSpecifier is the half of a ModuleEdge that a file's own syntax
// decides: which string literal names a module, what syntax it was written
// in, and whether that syntax survives into emitted JavaScript. What the
// specifier resolves to is deliberately absent — that is the Program's answer,
// and two Programs holding the same unchanged file can disagree about it.
type moduleSpecifier struct {
	specifier   *ast.Node
	declaration *ast.Node
	kind        ModuleEdgeKind
	typeOnly    bool
}

func collectSpecifiers(file *ast.SourceFile, syntax ModuleSyntax) []moduleSpecifier {
	// SourceFile.Imports is populated by the parser and avoids walking every
	// AST node in the usual static-ESM case. The generic collector remains
	// necessary for call-based references, parser recovery, and imports
	// inside module bodies.
	if syntax.CommonJS || syntax.AMD || needsFullModuleScan(file) {
		return collectByWalk(file, syntax)
	}
	return collectStaticImports(file)
}

// collectStaticImports reads the module specifiers the parser already
// recorded. It handles the shapes those specifiers can take in a file with no
// dynamic import, no module declaration and no parse error, which is what
// needsFullModuleScan checks for.
func collectStaticImports(file *ast.SourceFile) []moduleSpecifier {
	imports := file.Imports()
	specifiers := make([]moduleSpecifier, 0, len(imports))
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

		specifiers = append(specifiers, moduleSpecifier{
			specifier:   specifier,
			declaration: declaration,
			kind:        kind,
			typeOnly:    typeOnly,
		})
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

func collectByWalk(file *ast.SourceFile, syntax ModuleSyntax) []moduleSpecifier {
	var specifiers []moduleSpecifier
	// A module specifier can appear anywhere a call can, so every subtree is
	// walked; no node accounts for its own children here.
	utils.VisitDescendants(file.AsNode(), func(node *ast.Node) bool {
		switch node.Kind {
		case ast.KindImportDeclaration, ast.KindJSImportDeclaration:
			if !syntax.ESModule {
				return true
			}
			importDecl := node.AsImportDeclaration()
			appendSpecifier(&specifiers, importDecl.ModuleSpecifier, node, ModuleEdgeImport, importDeclarationOnlyImportsTypes(importDecl))
		case ast.KindExportDeclaration:
			if !syntax.ESModule {
				return true
			}
			exportDecl := node.AsExportDeclaration()
			// tsgo matches eslint-plugin-import here: only `export type * from`
			// is exclusively type-only; named type re-exports still stay edges.
			appendSpecifier(&specifiers, exportDecl.ModuleSpecifier, node, ModuleEdgeExport, ast.IsTypeOnlyImportOrExportDeclaration(node))
		case ast.KindCallExpression:
			appendCallSpecifiers(&specifiers, node.AsCallExpression(), syntax)
		}
		return true
	})
	return specifiers
}

func appendCallSpecifiers(specifiers *[]moduleSpecifier, call *ast.CallExpression, syntax ModuleSyntax) {
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
		appendSpecifier(specifiers, ast.SkipParentheses(call.Arguments.Nodes[0]), call.AsNode(), ModuleEdgeDynamicImport, false)
		return
	}

	if callee.Kind != ast.KindIdentifier {
		return
	}

	calleeName := callee.AsIdentifier().Text
	if syntax.CommonJS && ast.IsRequireCall(call.AsNode(), false) {
		arg := ast.SkipParentheses(call.Arguments.Nodes[0])
		if arg != nil && ast.IsStringLiteralLike(arg) {
			appendSpecifier(specifiers, arg, call.AsNode(), ModuleEdgeRequire, false)
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
			appendSpecifier(specifiers, element, call.AsNode(), ModuleEdgeAMD, false)
		}
	}
}

func appendSpecifier(specifiers *[]moduleSpecifier, specifier *ast.Node, declaration *ast.Node, kind ModuleEdgeKind, typeOnly bool) {
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

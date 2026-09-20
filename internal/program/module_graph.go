package program

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils/modules"
)

// ModuleReferenceKind and ModuleReferenceKinds use the shared syntax collector.
type ModuleReferenceKind = modules.ReferenceKind
type ModuleReferenceKinds = modules.ReferenceKinds

const (
	ModuleReferenceImport        = modules.ModuleReferenceImport
	ModuleReferenceExport        = modules.ModuleReferenceExport
	ModuleReferenceDynamicImport = modules.ModuleReferenceDynamicImport
	ModuleReferenceRequire       = modules.ModuleReferenceRequire
	ModuleReferenceAMD           = modules.ModuleReferenceAMD
	ESModuleReferences           = modules.ESModuleReferences
	CommonJSReferences           = modules.CommonJSReferences
	AMDReferences                = modules.AMDReferences
	AllModuleReferences          = modules.AllModuleReferences
)

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
		return graph.resolveAll(file, modules.Collect(file, kinds))
	})
}

// resolveAll turns what a file writes into what it references, which is the
// half of the answer only this Program can give.
func (graph ModuleGraph) resolveAll(file *ast.SourceFile, specifiers []modules.Source) []ModuleReference {
	if len(specifiers) == 0 {
		return nil
	}
	var references []ModuleReference
	for _, source := range specifiers {
		specifier := ast.SkipParentheses(source.Specifier)
		if specifier == nil || !ast.IsStringLiteralLike(specifier) {
			continue
		}
		reference := ModuleReference{
			Specifier: specifier, Declaration: source.Declaration,
			From: file, Kind: source.Kind, TypeOnly: source.TypeOnly,
		}
		reference.ResolvedPath, reference.Target, _ = graph.program.ResolveModule(file, specifier)
		references = append(references, reference)
	}
	return references
}

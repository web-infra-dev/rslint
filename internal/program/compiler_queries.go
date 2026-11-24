package program

import (
	"context"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
)

// CompilerQueries provides program-wide reads. Diagnostics can
// acquire several checkers and must be requested outside a file checker lease.
type CompilerQueries interface {
	Diagnostics(ctx context.Context, kind DiagnosticKind, source *ast.SourceFile) []*ast.Diagnostic
	CommandLine() *tsoptions.ParsedCommandLine
	ImplicitImports(path tspath.Path) []*ast.Node
	ResolvedProjectReferences() []*tsoptions.ParsedCommandLine
	ResolvedProjectReference(path tspath.Path) *tsoptions.ParsedCommandLine
	IdentifierCount() int
}

type DiagnosticKind uint8

const (
	ConfigFileParsingDiagnostics DiagnosticKind = iota
	OptionsDiagnostics
	GlobalDiagnostics
	SyntacticDiagnostics
	SemanticDiagnostics
	DeclarationDiagnostics
)

func (p *Program) CompilerQueries() CompilerQueries                  { return p.types }
func (b *compilerBackend) CommandLine() *tsoptions.ParsedCommandLine { return b.raw.CommandLine() }
func (b *compilerBackend) ImplicitImports(path tspath.Path) []*ast.Node {
	var imports []*ast.Node
	if helpers := b.raw.GetImportHelpersImportSpecifier(path); helpers != nil {
		imports = append(imports, helpers)
	}
	if _, jsx := b.raw.GetJSXRuntimeImportSpecifier(path); jsx != nil {
		imports = append(imports, jsx)
	}
	return imports
}
func (b *compilerBackend) ResolvedProjectReferences() []*tsoptions.ParsedCommandLine {
	return b.raw.GetResolvedProjectReferences()
}
func (b *compilerBackend) ResolvedProjectReference(path tspath.Path) *tsoptions.ParsedCommandLine {
	c, _ := b.raw.GetResolvedProjectReferenceFor(path)
	return c
}
func (b *compilerBackend) IdentifierCount() int { return b.raw.IdentifierCount() }
func (b *compilerBackend) Diagnostics(ctx context.Context, kind DiagnosticKind, file *ast.SourceFile) []*ast.Diagnostic {
	switch kind {
	case ConfigFileParsingDiagnostics:
		return b.raw.GetConfigFileParsingDiagnostics()
	case OptionsDiagnostics:
		ds := b.raw.GetProgramDiagnostics()
		if config := b.raw.CommandLine().ConfigFile; config != nil {
			ds = append(slices.Clone(ds), b.raw.GetIncludeProcessorDiagnostics(config.SourceFile)...)
		}
		return compiler.SortAndDeduplicateDiagnostics(ds)
	case GlobalDiagnostics:
		if len(b.raw.SourceFiles()) == 0 {
			return nil
		}
		pool := b.raw.GetCheckerPool()
		var globals []*ast.Diagnostic
		if collector, ok := pool.(interface{ GetGlobalDiagnostics() []*ast.Diagnostic }); ok {
			globals = collector.GetGlobalDiagnostics()
		}
		// The editor accumulates globals when its temporary checkers are released.
		// Its persistent API checker has a separate lifetime and owns its diagnostics.
		c, release := pool.GetChecker(core.WithCheckerLifetime(ctx, core.CheckerLifetimeAPI), nil)
		defer release()
		if fromAPI := c.GetGlobalDiagnostics(); len(fromAPI) > 0 {
			return compiler.SortAndDeduplicateDiagnostics(slices.Concat(globals, fromAPI))
		}
		return globals
	case SyntacticDiagnostics:
		return b.raw.GetSyntacticDiagnostics(ctx, file)
	case SemanticDiagnostics:
		return b.raw.GetSemanticDiagnostics(ctx, file)
	case DeclarationDiagnostics:
		return b.raw.GetDeclarationDiagnostics(ctx, file)
	default:
		panic("unknown compiler diagnostic kind")
	}
}

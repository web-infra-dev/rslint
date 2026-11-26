package program

import (
	"context"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
)

// CompilerQueries provides read-only program-wide queries. Diagnostics may
// acquire multiple checker shards and must run outside a file checker lease.
type CompilerQueries interface {
	Diagnostics(ctx context.Context, kind DiagnosticKind, source *ast.SourceFile) []*ast.Diagnostic
	CommandLine() *tsoptions.ParsedCommandLine
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
		c, release := pool.GetChecker(core.WithCheckerLifetime(ctx, core.CheckerLifetimeAPI), nil)
		defer release()
		if fromAPI := c.GetGlobalDiagnostics(); len(fromAPI) > 0 {
			globals = slices.Concat(globals, fromAPI)
		}
		return compiler.SortAndDeduplicateDiagnostics(globals)
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

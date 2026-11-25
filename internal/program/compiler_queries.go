package program

import (
	"context"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
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
	SyntacticDiagnostics DiagnosticKind = iota
	SemanticDiagnostics
	DeclarationDiagnostics
)

func (p *Program) CompilerQueries() CompilerQueries                  { return p.types }
func (b *compilerBackend) CommandLine() *tsoptions.ParsedCommandLine { return b.raw.CommandLine() }
func (b *compilerBackend) Diagnostics(ctx context.Context, kind DiagnosticKind, file *ast.SourceFile) []*ast.Diagnostic {
	switch kind {
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

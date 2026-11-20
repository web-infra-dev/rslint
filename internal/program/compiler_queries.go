package program

import (
	"context"
	"slices"
	"sync/atomic"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
)

// CompilerQueries provides program-wide reads. Diagnostics and statistics can
// acquire several checkers and must be requested outside a file checker lease.
type CompilerQueries interface {
	Diagnostics(ctx context.Context, kind DiagnosticKind, source *ast.SourceFile) []*ast.Diagnostic
	CommandLine() *tsoptions.ParsedCommandLine
	ResolvedProjectReferences() []*tsoptions.ParsedCommandLine
	ResolvedProjectReference(path tspath.Path) *tsoptions.ParsedCommandLine
	IdentifierCount() int
	TypeCount() int
	InstantiationCount() int
	SymbolCount() int
	RelationCacheSizes() [4]uint32
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
func (b *compilerBackend) ResolvedProjectReferences() []*tsoptions.ParsedCommandLine {
	return b.raw.GetResolvedProjectReferences()
}
func (b *compilerBackend) ResolvedProjectReference(path tspath.Path) *tsoptions.ParsedCommandLine {
	c, _ := b.raw.GetResolvedProjectReferenceFor(path)
	return c
}
func (b *compilerBackend) IdentifierCount() int    { return b.raw.IdentifierCount() }
func (b *compilerBackend) TypeCount() int          { return b.raw.TypeCount() }
func (b *compilerBackend) InstantiationCount() int { return b.raw.InstantiationCount() }
func (b *compilerBackend) SymbolCount() int        { return b.raw.SymbolCount() }
func (b *compilerBackend) RelationCacheSizes() [4]uint32 {
	var sizes [4]atomic.Uint32
	b.raw.ForEachCheckerParallel(func(_ int, c *checker.Checker) {
		for i, relation := range [...]*checker.Relation{
			checker.Checker_assignableRelation(c), checker.Checker_identityRelation(c),
			checker.Checker_subtypeRelation(c), checker.Checker_strictSubtypeRelation(c),
		} {
			sizes[i].Add(uint32(len(checker.Relation_results(relation))))
		}
	})
	return [4]uint32{sizes[0].Load(), sizes[1].Load(), sizes[2].Load(), sizes[3].Load()}
}
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
		return b.raw.GetGlobalDiagnostics(ctx)
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

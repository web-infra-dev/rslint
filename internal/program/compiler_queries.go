package program

import (
	"context"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
)

// CompilerQueries provides read-only program-wide queries. Diagnostics may
// acquire multiple checker shards and must run outside a file checker lease.
type CompilerQueries interface {
	Diagnostics(ctx context.Context, kind DiagnosticKind, source *ast.SourceFile) []*ast.Diagnostic
	ConfigSourceFile() *ast.SourceFile
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

func (p *Program) CompilerQueries() CompilerQueries { return p.types }
func (b *compilerBackend) ConfigSourceFile() *ast.SourceFile {
	if cmd := b.raw.CommandLine(); cmd != nil && cmd.ConfigFile != nil {
		return cmd.ConfigFile.SourceFile
	}
	return nil
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
		if len(b.raw.SourceFiles()) == 0 {
			return nil
		}
		if unlock := b.isolatedDeclReadLock(); unlock != nil {
			defer unlock()
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
		if file != nil && !b.ownsSourceFile(file) {
			return nil
		}
		if unlock := b.isolatedDeclReadLock(); unlock != nil {
			defer unlock()
		}
		if file == nil {
			// Clear request affinity before tsgo fans out per-file checker borrows
			// across goroutines so project.CheckerPool cannot alias one checker
			// onto concurrent workers sharing the same RequestID.
			ctx = core.WithRequestID(ctx, "")
		}
		return b.raw.GetSemanticDiagnostics(ctx, file)
	case DeclarationDiagnostics:
		// Guard non-program files (e.g. tsconfig.json) before delegating: tsgo's
		// compiler.checkerPool only associates program source files. Do NOT
		// pre-acquire GetTypeCheckerForFileExclusive here: compiler.checkerPool's
		// EmitResolver already locks the checker's own mutex (the same lock as
		// p.locks[idx], so pre-locking self-deadlocks), while external pools such
		// as project.CheckerPool acquire an exclusive checker inside
		// GetDeclarationDiagnostics.
		if file != nil && (file.IsDeclarationFile || !b.ownsSourceFile(file)) {
			return nil
		}
		if unlock := b.isolatedDeclWriteLock(); unlock != nil {
			defer unlock()
			if file == nil {
				var diagnostics []*ast.Diagnostic
				for _, sourceFile := range b.raw.SourceFiles() {
					if ctx.Err() != nil {
						return nil
					}
					if sourceFile.IsDeclarationFile {
						continue
					}
					diagnostics = append(diagnostics, b.raw.GetDeclarationDiagnostics(ctx, sourceFile)...)
				}
				return compiler.SortAndDeduplicateDiagnostics(diagnostics)
			}
			return b.raw.GetDeclarationDiagnostics(ctx, file)
		}
		if file == nil {
			ctx = core.WithRequestID(ctx, "")
		}
		return b.raw.GetDeclarationDiagnostics(ctx, file)
	default:
		panic("unknown compiler diagnostic kind")
	}
}

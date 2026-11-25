package typeservice

import (
	"context"
	"errors"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/program"
)

func programQuery(ctx context.Context, p *program.Program, r *ProgramRequest, node *ast.Node) (Result, error) {
	q := p.CompilerQueries()
	if q == nil {
		return nil, errors.New("program query requires a typed project")
	}
	var file *ast.SourceFile
	if node != nil {
		if node.Kind != ast.KindSourceFile {
			return nil, errors.New("expected a Program source file")
		}
		file = node.AsSourceFile()
	}
	if r.Op == "rootFileNames" {
		return Strings(p.RootFileNames()), nil
	}
	kind, ok := diagnosticKinds[r.Op]
	if !ok {
		return nil, errors.New("unknown Program query")
	}
	return diagnosticFacts(q.Diagnostics(ctx, kind, file)), ctx.Err()
}

var diagnosticKinds = map[Operation]program.DiagnosticKind{
	"syntacticDiagnostics":   program.SyntacticDiagnostics,
	"semanticDiagnostics":    program.SemanticDiagnostics,
	"declarationDiagnostics": program.DeclarationDiagnostics,
}

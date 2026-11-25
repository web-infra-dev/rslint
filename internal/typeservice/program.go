package typeservice

import (
	"context"
	"errors"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/program"
)

func programQuery(ctx context.Context, s *Session, p *program.Program, r *ProgramRequest, node *ast.Node) (Result, error) {
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
	ds := q.Diagnostics(ctx, kind, file)
	recordDiagnosticSources(s, p, ds)
	return diagnosticFacts(ds), ctx.Err()
}

func recordDiagnosticSources(s *Session, p *program.Program, ds []*ast.Diagnostic) {
	for _, d := range ds {
		if f := d.File(); f != nil && p.GetSourceFile(f.FileName()) == nil {
			s.recordExtraSource(f)
		}
		if related := d.RelatedInformation(); len(related) > 0 {
			recordDiagnosticSources(s, p, related)
		}
	}
}

var diagnosticKinds = map[Operation]program.DiagnosticKind{
	"configFileParsingDiagnostics": program.ConfigFileParsingDiagnostics,
	"optionsDiagnostics":           program.OptionsDiagnostics,
	"globalDiagnostics":            program.GlobalDiagnostics,
	"syntacticDiagnostics":         program.SyntacticDiagnostics,
	"semanticDiagnostics":          program.SemanticDiagnostics,
	"declarationDiagnostics":       program.DeclarationDiagnostics,
}

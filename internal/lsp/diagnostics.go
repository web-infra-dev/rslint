package lsp

import (
	"context"
	"fmt"
	"log"
	"runtime"

	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
	"github.com/microsoft/typescript-go/shim/scanner"

	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// documentProgressiveDiagnostics implements the presentation and asynchronous
// admission ports of a progressive lint request. It does not decide producer
// order or eligibility; RunPipeline calls these capabilities in that order.
type documentProgressiveDiagnostics struct {
	server           *Server
	uri              lsproto.DocumentUri
	generation       uint64
	pluginGeneration string
}

func (p *documentProgressiveDiagnostics) PublishBaseline(
	ctx context.Context,
	diagnostics []rule.RuleDiagnostic,
) {
	if diagnostics == nil {
		diagnostics = []rule.RuleDiagnostic{}
	}
	p.server.diagnostics[p.uri] = diagnostics
	lspDiagnostics := make([]*lsproto.Diagnostic, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		lspDiagnostics = append(lspDiagnostics, convertRuleDiagnosticToLSP(diagnostic))
	}
	if err := p.server.PublishDiagnostics(ctx, &lsproto.PublishDiagnosticsParams{
		Uri:         p.uri,
		Diagnostics: lspDiagnostics,
	}); err != nil {
		log.Printf("Error publishing diagnostics: %v", err)
	}
}

func (s *Server) invalidateOpenDocumentDiagnostics() {
	for uri := range s.documents {
		s.docGeneration[uri]++
		s.cancelInflightPluginDispatch(uri)
		delete(s.diagnostics, uri)
	}
}

// isTsConfigURI returns true if the URI points to a tsconfig/jsconfig file,
// including variants like tsconfig.build.json, tsconfig.app.json, etc.
func convertRuleDiagnosticToLSP(ruleDiag rule.RuleDiagnostic) *lsproto.Diagnostic {
	diagnosticStart := ruleDiag.Range.Pos()
	diagnosticEnd := ruleDiag.Range.End()
	startLine, startColumn := scanner.GetECMALineAndUTF16CharacterOfPosition(ruleDiag.SourceFile, diagnosticStart)
	endLine, endColumn := scanner.GetECMALineAndUTF16CharacterOfPosition(ruleDiag.SourceFile, diagnosticEnd)

	return &lsproto.Diagnostic{
		Range: lsproto.Range{
			Start: lsproto.Position{
				Line:      uint32(startLine),
				Character: uint32(startColumn),
			},
			End: lsproto.Position{
				Line:      uint32(endLine),
				Character: uint32(endColumn),
			},
		},
		Severity: ptrTo(lsproto.DiagnosticSeverity(ruleDiag.Severity.Int())),
		Source:   ptrTo("rslint"),
		Message:  lsproto.StringOrMarkupContent{String: ptrTo(fmt.Sprintf("[%s] %s", ruleDiag.RuleName, ruleDiag.Message.Description))},
	}
}

// pushDiagnostics runs the linter for the given URI and pushes results to the client.
// Must be called synchronously from the LSP message loop (not from a goroutine)
// because session is not goroutine-safe.
func (s *Server) pushDiagnostics(uri lsproto.DocumentUri) {
	if s.session == nil {
		return
	}

	ctx := s.backgroundCtx

	if !isLintableScriptFile(uri) {
		return
	}

	// Supersede the previous enrichment before linting. The request-scoped
	// presentation adapter is stamped so the main loop can reject an older
	// asynchronous result.
	s.docGeneration[uri]++
	generation := s.docGeneration[uri]
	s.cancelInflightPluginDispatch(uri)
	snapshot := s.documentLintSnapshot(uri)
	if snapshot.unavailable {
		delete(s.diagnostics, uri)
		if err := s.PublishDiagnostics(ctx, &lsproto.PublishDiagnosticsParams{
			Uri:         uri,
			Diagnostics: []*lsproto.Diagnostic{},
		}); err != nil {
			log.Printf("Error clearing diagnostics for unavailable config: %v", err)
		}
		return
	}

	presentation := &documentProgressiveDiagnostics{
		server:           s,
		uri:              uri,
		generation:       generation,
		pluginGeneration: snapshot.pluginGeneration,
	}
	_, err := linter.RunPipeline(ctx, linter.NewProgressiveLintRequest(
		&documentGenerationProvider{server: s, uri: uri, snapshot: snapshot},
		linter.ArtifactDemand{
			Native: rule.EditDemandAll,
			Plugin: rule.EditDemandAll,
		},
		presentation,
	))
	if err != nil {
		log.Printf("Error running lint for push diagnostics: %v", err)
		delete(s.diagnostics, uri)
		if publishErr := s.PublishDiagnostics(ctx, &lsproto.PublishDiagnosticsParams{
			Uri:         uri,
			Diagnostics: []*lsproto.Diagnostic{},
		}); publishErr != nil {
			log.Printf("Error clearing diagnostics after lint failure: %v", publishErr)
		}
		return
	}
	// The pacer cannot see that the lint just dropped what it derived.
	go runtime.GC()
}

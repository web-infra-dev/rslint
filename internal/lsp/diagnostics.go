package lsp

import (
	"context"
	"errors"
	"fmt"
	"log"
	"runtime"

	"github.com/microsoft/TypeScript/tsc/shim/lsp/lsproto"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"

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
	diagnostics      []rule.RuleDiagnostic
	lspDiagnostics   []*lsproto.Diagnostic
	enrichment       linter.DeferredPluginRun
}

func (p *documentProgressiveDiagnostics) PublishBaseline(
	_ context.Context,
	diagnostics []rule.RuleDiagnostic,
) {
	if diagnostics == nil {
		diagnostics = []rule.RuleDiagnostic{}
	}
	p.diagnostics = diagnostics
}

// publish commits a completed observation on the dispatch loop. In between
// calculation and publication, queued notifications may have closed the document
// or invalidated its inputs. Enrichment starts only after a valid baseline.
func (p *documentProgressiveDiagnostics) publish() {
	s := p.server
	if _, open := s.documents[p.uri]; !open || s.docGeneration[p.uri] != p.generation {
		return
	}
	if p.diagnostics == nil {
		delete(s.diagnostics, p.uri)
	} else {
		s.diagnostics[p.uri] = p.diagnostics
	}
	lspDiagnostics := p.lspDiagnostics
	if lspDiagnostics == nil {
		lspDiagnostics = make([]*lsproto.Diagnostic, 0, len(p.diagnostics))
		for _, diagnostic := range p.diagnostics {
			lspDiagnostics = append(lspDiagnostics, convertRuleDiagnosticToLSP(diagnostic))
		}
	}
	if err := s.PublishDiagnostics(s.backgroundCtx, &lsproto.PublishDiagnosticsParams{
		Uri:         p.uri,
		Diagnostics: lspDiagnostics,
	}); err != nil {
		log.Printf("Error publishing diagnostics: %v", err)
	}
	s.startDiagnosticEnrichment(s.backgroundCtx, p.uri, p.generation, p.pluginGeneration, p.enrichment)
}

func (s *Server) invalidateOpenDocumentDiagnostics() {
	for uri := range s.documents {
		s.invalidateDocumentDiagnostics(uri)
	}
}

func (s *Server) invalidateDocumentDiagnostics(uri lsproto.DocumentUri) {
	s.docGeneration[uri]++
	s.cancelInflightPluginDispatch(uri)
	delete(s.diagnostics, uri)
	if result := s.pendingDiagnosticResult; result != nil && result.uri == uri {
		s.pendingDiagnosticResult = nil
	}
}

// An editor update can change another document's dependency view even when its
// own text is unchanged. Use the Program store's affected owners; documents
// without a resident graph must conservatively observe the new overlay too.
// Invalidation is immediate; dependent observations share the typing debounce
// so opening and closing transient overlays does not repeatedly lint consumers.
func (s *Server) invalidateDocumentInputs(uri lsproto.DocumentUri, owners []lsproto.DocumentUri) {
	s.invalidateDocumentDiagnostics(uri)
	if !s.initialized {
		return
	}
	invalidateDependent := func(dependent lsproto.DocumentUri) {
		if dependent == uri {
			return
		}
		s.invalidateDocumentDiagnostics(dependent)
		s.scheduleLint(dependent)
	}
	for _, owner := range owners {
		invalidateDependent(owner)
	}
	for document := range s.documents {
		if s.lintPrograms != nil {
			if _, owned := s.lintPrograms.documentPrograms[document]; owned {
				continue
			}
		}
		invalidateDependent(document)
	}
}

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
// because document and Program state belongs to that loop.
func (s *Server) pushDiagnostics(uri lsproto.DocumentUri) {
	s.consumeDiagnosticRefresh()
	delete(s.lintSchedule.pending, uri)
	if result := s.prepareDiagnostics(uri); result != nil {
		result.publish()
	}
}

func (s *Server) prepareDiagnostics(uri lsproto.DocumentUri) *documentProgressiveDiagnostics {
	if !s.initialized {
		return nil
	}

	ctx := s.backgroundCtx

	if !isLintableScriptFile(uri) {
		return nil
	}

	// Supersede the previous enrichment before linting. The request-scoped
	// presentation adapter is stamped so the main loop can reject an older
	// asynchronous result.
	s.docGeneration[uri]++
	generation := s.docGeneration[uri]
	s.cancelInflightPluginDispatch(uri)
	snapshot := s.documentLintSnapshot(uri)
	presentation := &documentProgressiveDiagnostics{
		server:           s,
		uri:              uri,
		generation:       generation,
		pluginGeneration: snapshot.pluginGeneration,
	}
	if snapshot.unavailable {
		s.releaseDocumentProgram(uri)
		return presentation
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
		presentation.diagnostics = nil
		presentation.enrichment = nil
		var projectErr *documentProjectConfigError
		if ctx.Err() == nil && errors.As(err, &projectErr) {
			presentation.lspDiagnostics = []*lsproto.Diagnostic{{
				Severity: ptrTo(lsproto.DiagnosticSeverityError),
				Source:   ptrTo("rslint"),
				Message:  lsproto.StringOrMarkupContent{String: ptrTo(fmt.Sprintf("Project configuration error: %v", projectErr))},
			}}
		}
		return presentation
	}
	// The pacer cannot see that the lint just dropped what it derived.
	go runtime.GC()
	return presentation
}

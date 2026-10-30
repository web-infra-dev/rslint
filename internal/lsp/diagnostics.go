package lsp

import (
	"context"
	"errors"
	"fmt"
	"log"
	"runtime"

	"github.com/microsoft/TypeScript/tsc/shim/lsp/lsproto"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs"

	"github.com/web-infra-dev/rslint/internal/config"
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
	delete(s.diagnosticInputs, uri)
	if result := s.pendingDiagnosticResult; result != nil && result.uri == uri {
		s.pendingDiagnosticResult = nil
	}
}

// An editor update can change another document's dependency view even when its
// own text is unchanged. Use the Program store's affected owners and source-only
// observations' filesystem lookups. Unknown inputs must conservatively observe
// the new overlay, but cannot postpone an already pending calculation.
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
		if _, pending := s.lintSchedule.pending[dependent]; !pending {
			s.scheduleLint(dependent)
		}
	}
	for _, owner := range owners {
		invalidateDependent(owner)
	}
	for document := range s.documents {
		if document == uri || !isLintableScriptFile(document) {
			continue
		}
		if s.lintPrograms != nil {
			if _, owned := s.lintPrograms.documentPrograms[document]; owned {
				continue
			}
		}
		if inputs, known := s.diagnosticInputs[document]; known {
			identity := config.FreezeFileIdentity(uriToPath(uri), s.fs)
			if !inputs.containsPath(identity.Path, s.fs) && !inputs.containsPath(identity.CanonicalPath, s.fs) {
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
func (s *Server) pushDiagnostics(ctx context.Context, uri lsproto.DocumentUri) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.consumeDiagnosticRefresh()
	result := s.prepareDiagnostics(ctx, uri)
	if err := ctx.Err(); err != nil {
		// A canceled quick fix must leave the background demand in place.
		return err
	}
	delete(s.lintSchedule.pending, uri)
	if result != nil {
		result.publish()
	}
	return nil
}

func (s *Server) prepareDiagnostics(ctx context.Context, uri lsproto.DocumentUri) *documentProgressiveDiagnostics {
	if !s.initialized || ctx.Err() != nil {
		return nil
	}

	if !isLintableScriptFile(uri) {
		return nil
	}

	// Supersede the previous enrichment before linting. The request-scoped
	// presentation adapter is stamped so the main loop can reject an older
	// asynchronous result.
	s.docGeneration[uri]++
	generation := s.docGeneration[uri]
	s.cancelInflightPluginDispatch(uri)
	delete(s.diagnosticInputs, uri)
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
	provider := &documentGenerationProvider{server: s, uri: uri, snapshot: snapshot}
	_, err := linter.RunPipeline(ctx, linter.NewProgressiveLintRequest(
		provider,
		linter.ArtifactDemand{
			Native: rule.EditDemandAll,
			Plugin: rule.EditDemandAll,
		},
		presentation,
	))
	if ctx.Err() != nil {
		return nil
	}
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
	if provider.sourceOnlyInputs != nil {
		lookups := provider.sourceOnlyInputs.drain()
		// Freeze dependency identities before returning to the message loop.
		// Missing files and parents can later be opened from either side of an
		// alias, and a moved alias must not reinterpret an earlier observation.
		identities := cachedvfs.From(provider.sourceOnlyInputs.Inner)
		for _, path := range lookups.seenFiles {
			identity := config.FreezeFileIdentity(path, identities)
			physical := identity.CanonicalPath
			if tspath.GetBaseFileName(path) == lintProgramDirectoryWatchName {
				physical = lintProgramDirectoryWatchInput(identity.CanonicalParentPath)
			}
			if physical != identity.Path {
				lookups.seenFiles = append(lookups.seenFiles, physical)
			}
		}
		if s.diagnosticInputs == nil {
			s.diagnosticInputs = make(map[lsproto.DocumentUri]lintProgramLookups)
		}
		s.diagnosticInputs[uri] = lookups
	}
	// The pacer cannot see that the lint just dropped what it derived.
	go runtime.GC()
	return presentation
}

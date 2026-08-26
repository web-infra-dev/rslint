package lsp

import (
	"fmt"
	"log"
	"runtime"

	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
	"github.com/microsoft/typescript-go/shim/scanner"

	"github.com/web-infra-dev/rslint/internal/rule"
)

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

	// Supersede the previous plugin pass before linting. Native diagnostics are
	// published synchronously below; the next plugin pass is stamped with this
	// generation so the main loop can reject an older result.
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

	lintResult, err := s.runConfiguredLint(uri, ctx, snapshot)
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
	ruleDiags := lintResult.Diagnostics

	s.diagnostics[uri] = ruleDiags

	// Must use empty slice (not nil) so JSON serializes as [] instead of null
	lspDiags := make([]*lsproto.Diagnostic, 0, len(ruleDiags))
	for _, d := range ruleDiags {
		lspDiags = append(lspDiags, convertRuleDiagnosticToLSP(d))
	}

	if err := s.PublishDiagnostics(ctx, &lsproto.PublishDiagnosticsParams{
		Uri:         uri,
		Diagnostics: lspDiags,
	}); err != nil {
		log.Printf("Error publishing diagnostics: %v", err)
	}

	// Dispatch eslint-plugin rules off the main loop. The reverse request
	// MUST NOT run synchronously here — it would block the dispatch loop (and
	// thus all editor interaction) until the Node worker replies. Results merge
	// back via pluginResultCh on the main loop (s.diagnostics is lock-free).
	if !lintResult.HasSyntaxErrors {
		s.dispatchPluginLintWithSnapshot(uri, generation, snapshot)
	}

	// The pacer cannot see that the lint just dropped what it derived.
	go runtime.GC()
}

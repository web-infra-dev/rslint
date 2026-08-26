package lsp

import (
	"context"
	"log"
	"time"

	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
)

// lintDebounceDelay is how long to wait after the last keystroke before
// running the linter. This avoids linting on every keystroke against
// incomplete/broken syntax that can cause panics or waste CPU.
const lintDebounceDelay = 200 * time.Millisecond

// scheduleLint marks a URI for deferred linting and resets the debounce timer.
// When the timer fires it signals debounceCh, which is consumed by the main
// dispatch loop. Must be called from the main dispatch loop goroutine.
func (s *Server) scheduleLint(uri lsproto.DocumentUri) {
	s.pendingLintURIs[uri] = struct{}{}
	if s.lintTimer != nil {
		s.lintTimer.Stop()
	}
	s.lintTimer = time.AfterFunc(lintDebounceDelay, func() {
		select {
		case s.debounceCh <- struct{}{}:
		default:
			// Already pending — no need to queue another signal
		}
	})
}

func (s *Server) handleDidOpen(ctx context.Context, params *lsproto.DidOpenTextDocumentParams) error {
	log.Printf("Handling didOpen: %s", params.TextDocument.Uri)

	uri := params.TextDocument.Uri
	content := params.TextDocument.Text

	s.documents[uri] = content
	if s.lintPrograms != nil {
		existedOnDisk := s.fs != nil && s.fs.FileExists(uriToPath(uri))
		s.lintPrograms.DidOpen(uri, content, existedOnDisk)
	}

	// Notify session about the opened file so it creates the overlay.
	if s.session != nil {
		s.session.DidOpenFile(ctx, uri, params.TextDocument.Version, content, params.TextDocument.LanguageId)
		s.pushDiagnostics(uri)
	}
	return nil
}

func (s *Server) handleDidChange(ctx context.Context, params *lsproto.DidChangeTextDocumentParams) error {
	log.Printf("Handling didChange: %s (version %d)", params.TextDocument.Uri, params.TextDocument.Version)

	uri := params.TextDocument.Uri

	content, err := applyDocumentChanges(s.documents[uri], params.ContentChanges)
	if err != nil {
		// didChange is a notification, so the server cannot return an error to
		// request a retry. Keep both document mirrors on their previous content
		// instead of partially applying malformed input or sending an invalid
		// JSON-RPC response with a null request ID.
		log.Printf("Ignoring invalid didChange for %s: %v", uri, err)
		return nil
	}
	s.documents[uri] = content
	if s.lintPrograms != nil {
		s.lintPrograms.DidChange(uri, s.documents[uri])
	}

	// A content version change supersedes plugin work immediately, not when the
	// debounce timer eventually starts the next lint. Otherwise an older worker
	// result can be published against the new buffer while the user is typing.
	s.docGeneration[uri]++
	s.cancelInflightPluginDispatch(uri)
	delete(s.diagnostics, uri)

	// Notify session immediately so tsgo's overlay stays up-to-date for
	// other LSP features (completions, hover, etc.).  Lint is deferred
	// via scheduleLint to avoid running the linter on every keystroke.
	if s.session != nil {
		s.session.DidChangeFile(ctx, uri, params.TextDocument.Version, params.ContentChanges)
		s.scheduleLint(uri)
	}
	return nil
}

func (s *Server) handleDidSave(ctx context.Context, params *lsproto.DidSaveTextDocumentParams) error {
	log.Printf("Handling didSave: %s", params.TextDocument.Uri)
	uri := params.TextDocument.Uri

	// didChange is authoritative for the current content of an open document.
	// didSave may include the text that reached disk, but carries no document
	// version, so a save for an older buffer can arrive after a newer didChange.
	// Never replace the versioned document mirror with this unversioned snapshot.
	currentContent, open := s.documents[uri]
	forwardSave := shouldForwardDidSave(currentContent, open, params.Text)
	if !forwardSave {
		log.Printf("Ignoring stale didSave for open document %s", uri)
	}

	// Clear pending debounce lint for this URI — pushDiagnostics below
	// will lint it immediately, so the debounce would be redundant.
	delete(s.pendingLintURIs, uri)

	if s.lintPrograms != nil && forwardSave {
		s.lintPrograms.DidSave(uri, open)
	}

	// Notify session about the save event.
	if s.session != nil {
		if forwardSave {
			s.session.DidSaveFile(ctx, uri)
		}
		s.pushDiagnostics(uri)
	}
	return nil
}

// shouldForwardDidSave suppresses only saves that are known to describe an
// older version of an open document. Saves without text and saves for documents
// not tracked as open are forwarded for LSP client compatibility and so tsgo
// can observe out-of-band disk changes.
func shouldForwardDidSave(currentContent string, open bool, savedText *string) bool {
	return savedText == nil || !open || currentContent == *savedText
}

func (s *Server) handleDidClose(ctx context.Context, params *lsproto.DidCloseTextDocumentParams) error {
	log.Printf("Handling didClose: %s", params.TextDocument.Uri)
	uri := params.TextDocument.Uri
	delete(s.documents, uri)
	delete(s.diagnostics, uri)
	delete(s.pendingLintURIs, uri)
	// Bump (do NOT delete) the generation on close so any in-flight plugin
	// result for this URI is stale. Keeping the counter monotonic — rather
	// than resetting it to 0 on a later reopen — prevents a generation collision
	// where a pre-close worker result could match a freshly reopened document.
	s.docGeneration[uri]++
	// Cancel an in-flight plugin dispatch for the closed doc so its Node worker
	// stops instead of running to completion — no superseding keystroke will.
	s.cancelInflightPluginDispatch(uri)
	if s.lintPrograms != nil {
		s.lintPrograms.DidClose(uri)
	}

	if s.session != nil {
		// Push empty diagnostics to clear the client's display before closing
		if err := s.PublishDiagnostics(ctx, &lsproto.PublishDiagnosticsParams{
			Uri:         uri,
			Diagnostics: []*lsproto.Diagnostic{},
		}); err != nil {
			log.Printf("Error clearing diagnostics on close: %v", err)
		}
		s.session.DidCloseFile(ctx, uri)
	}
	return nil
}

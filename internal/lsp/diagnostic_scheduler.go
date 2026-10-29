package lsp

import (
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/lsp/lsproto"
)

const lintDebounceDelay = 200 * time.Millisecond

type scheduledDiagnostic struct {
	due      time.Time
	sequence uint64
}

// diagnosticScheduler owns pending work, not document or Program snapshots.
// The dispatch loop is its sole caller. Deadlines are per document so editing
// one buffer cannot postpone diagnostics for another buffer.
type diagnosticScheduler struct {
	pending  map[lsproto.DocumentUri]scheduledDiagnostic
	sequence uint64
}

func (q *diagnosticScheduler) schedule(uri lsproto.DocumentUri, due time.Time) {
	if q.pending == nil {
		q.pending = make(map[lsproto.DocumentUri]scheduledDiagnostic)
	}
	entry, exists := q.pending[uri]
	if !exists {
		q.sequence++
		entry.sequence = q.sequence
	}
	entry.due = due
	q.pending[uri] = entry
}

// next returns the oldest ready document, or the delay until the next deadline.
// A rescheduled document keeps its place among ready work without preventing
// other documents from running while it is still being edited.
func (q *diagnosticScheduler) next(now time.Time) (lsproto.DocumentUri, time.Duration, bool) {
	var readyURI lsproto.DocumentUri
	var sequence uint64
	var earliest time.Time
	for uri, entry := range q.pending {
		if !entry.due.After(now) {
			if sequence == 0 || entry.sequence < sequence {
				readyURI, sequence = uri, entry.sequence
			}
		} else if earliest.IsZero() || entry.due.Before(earliest) {
			earliest = entry.due
		}
	}
	if sequence != 0 {
		return readyURI, 0, true
	}
	if !earliest.IsZero() {
		return "", earliest.Sub(now), true
	}
	return "", 0, false
}

func (s *Server) scheduleLint(uri lsproto.DocumentUri) {
	s.lintSchedule.schedule(uri, time.Now().Add(lintDebounceDelay))
}

func (s *Server) queueDiagnostics(uri lsproto.DocumentUri) {
	s.lintSchedule.schedule(uri, time.Time{})
}

func (s *Server) consumeDiagnosticRefresh() {
	select {
	case <-s.refreshCh:
		s.queueOpenDocumentDiagnostics()
	default:
	}
}

func (s *Server) queueOpenDocumentDiagnostics() {
	s.invalidateOpenDocumentDiagnostics()
	for uri := range s.documents {
		s.queueDiagnostics(uri)
	}
}

// A quick-fix request observes the input preceding it in the request queue.
// Flush its pending diagnostics here, before processing any later notification.
func (s *Server) flushDocumentDiagnostics(uri lsproto.DocumentUri) {
	s.consumeDiagnosticRefresh()
	if result := s.pendingDiagnosticResult; result != nil && result.uri == uri {
		s.pendingDiagnosticResult = nil
		result.publish()
	}
	if _, pending := s.lintSchedule.pending[uri]; pending {
		s.pushDiagnostics(uri)
	}
}

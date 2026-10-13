package program

import (
	"context"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
)

type diagnosticBackend interface {
	noEmitDiagnostics(ctx context.Context) []*ast.Diagnostic
}

// DiagnosticSource is an immutable program-wide diagnostic capability over one
// source generation. It retains its backend and path authority independently of
// the lint facade and its derived caches.
type DiagnosticSource struct {
	backend          diagnosticBackend
	fs               vfs.FS
	currentDirectory string
}

// DiagnosticSource borrows only the state required for complete diagnostics.
// An unavailable capability returns nil; acquiring it performs no compiler work.
func (p *Program) DiagnosticSource() *DiagnosticSource {
	if !p.CanProvideProgramDiagnostics() {
		return nil
	}
	return &DiagnosticSource{backend: p.types, fs: p.FS(), currentDirectory: p.CurrentDirectory()}
}

func (d *DiagnosticSource) NoEmitDiagnostics(ctx context.Context) []*ast.Diagnostic {
	if d == nil || d.backend == nil {
		return nil
	}
	return d.backend.noEmitDiagnostics(ctx)
}

func (d *DiagnosticSource) FS() vfs.FS {
	if d == nil {
		return nil
	}
	return d.fs
}

func (d *DiagnosticSource) CurrentDirectory() string {
	if d == nil {
		return ""
	}
	return d.currentDirectory
}

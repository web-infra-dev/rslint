// Package loader assembles run-scoped rslint Programs from configured
// TypeScript projects and target-plan files. Backend selection and
// target-to-source binding stay private to this package; callers receive only
// the unified Program sequence and its lint projection.
package loader

import (
	"errors"
	"reflect"

	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
)

// Session owns Program-construction services for one immutable CLI/API source
// generation. A caller may explicitly invalidate its source snapshots before
// reusing it, but the lint pipeline's overlay providers use a fresh Session per
// fix generation. Sessions are never shared between requests or with the LSP
// lifecycle.
type Session struct {
	context            *buildContext
	initialPrograms    []*compiler.Program
	initialProgramsSet bool
}

func NewSession(fsys vfs.FS) *Session {
	if isNilFS(fsys) {
		return &Session{}
	}
	return &Session{context: newBuildContext(fsys)}
}

func isNilFS(fsys vfs.FS) bool {
	if fsys == nil {
		return true
	}
	value := reflect.ValueOf(fsys)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func (s *Session) validate() error {
	if s == nil || s.context == nil || isNilFS(s.context.FS()) {
		return errors.New("program loader: invalid session")
	}
	return nil
}

func (s *Session) FS() vfs.FS {
	if s == nil || s.context == nil {
		return nil
	}
	return s.context.FS()
}

// InvalidateSourceSnapshots begins a fresh source generation before a fix
// rebuild. Program objects themselves remain immutable and must be rebuilt.
func (s *Session) InvalidateSourceSnapshots() {
	if s != nil && s.context != nil {
		s.context.invalidateSourceSnapshots()
	}
}

// ReleaseBuildCache ends an AST reuse window after all construction and target
// binding have joined. It drops only construction references, never published
// Programs, their sources, or the Session's source/metadata snapshots.
//
// Reusing the Session remains supported: later builds repopulate the AST index
// from the same frozen text, or from a new source generation after explicit
// invalidation. Callers that want AST reuse across builds simply keep the cache.
// Like construction itself, this method must not overlap another Session build.
func (s *Session) ReleaseBuildCache() {
	if s == nil || s.context == nil {
		return
	}
	s.initialPrograms = nil
	s.initialProgramsSet = false
	s.context.retainOnlySourceFiles(nil)
}

func (s *Session) retainCompilerPrograms(current []*compiler.Program) {
	if s == nil || s.context == nil {
		return
	}
	if !s.initialProgramsSet {
		s.initialProgramsSet = true
		s.initialPrograms = append([]*compiler.Program(nil), current...)
	}
	live := make([]*compiler.Program, 0, len(current)+len(s.initialPrograms))
	live = append(live, current...)
	live = append(live, s.initialPrograms...)
	s.context.retainOnlySourceFiles(live)
}

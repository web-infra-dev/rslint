// Package loader assembles run-scoped rslint Programs from configured
// TypeScript projects and config-owned lint targets. Backend selection and
// target-to-source binding stay private to this package; callers receive only
// the unified Program sequence and its lint projection.
package loader

import (
	"errors"
	"reflect"

	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/vfs"
)

// Session owns Program-construction services for one CLI invocation or API
// request. It is reused across fix generations but never shared between
// requests or with the LSP session lifecycle.
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

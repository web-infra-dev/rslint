package program

import (
	"errors"
	"fmt"
)

// SourceSet is an ordered set of roots whose ASTs have not been materialized.
// It owns construction inputs, not execution policy or a partially populated
// Program. Build preserves the complete source universe; BuildFile creates an
// independent one-file Program for consumers that require no other root ASTs.
// The caller keeps the host and options immutable, and the host must not retain
// parsed ASTs. Concurrent builds may share immutable source-text snapshots.
type SourceSet struct {
	options RootOptions
}

func NewSourceSet(options RootOptions) (*SourceSet, error) {
	if isNilInterface(options.Host) {
		return nil, errors.New("program: source set requires a compiler host")
	}
	if options.CompilerOptions == nil {
		return nil, errors.New("program: source set requires compiler options")
	}
	fs := options.Host.FS()
	if isNilInterface(fs) {
		return nil, errors.New("program: source set requires a filesystem")
	}
	roots, err := normalizeRootFileNames(options.RootFileNames, options.Host.GetCurrentDirectory(), fs)
	if err != nil {
		return nil, err
	}
	options.RootFileNames = roots
	return &SourceSet{options: options}, nil
}

func (s *SourceSet) IsValid() bool {
	return s != nil && s.options.Host != nil && s.options.CompilerOptions != nil
}

// FileNames returns the normalized roots in stable order. The slice is read-only.
func (s *SourceSet) FileNames() []string {
	if !s.IsValid() {
		return nil
	}
	return s.options.RootFileNames
}

func (s *SourceSet) Build() (*Program, error) {
	if !s.IsValid() {
		return nil, errors.New("program: invalid source set")
	}
	return NewFromRoots(s.options)
}

// BuildFile does not cache the resulting Program. Its consumer owns the entire
// lifetime, so completed work cannot be retained by the remaining input queue.
func (s *SourceSet) BuildFile(index int) (*Program, error) {
	if !s.IsValid() || index < 0 || index >= len(s.options.RootFileNames) {
		return nil, fmt.Errorf("program: invalid source set file index %d", index)
	}
	options := s.options
	options.RootFileNames = options.RootFileNames[index : index+1]
	options.SingleThreaded = true
	return NewFromRoots(options)
}

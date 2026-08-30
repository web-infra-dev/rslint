package linter

import (
	"errors"
	"fmt"
	"sort"
)

// autofixState is the pipeline's sole mutable autofix memory. Generation
// providers receive detached snapshots and cannot advance or roll back it.
type autofixState struct {
	initial            map[string]string
	current            map[string]string
	appliedDiagnostics map[string]int
}

func newAutofixState() *autofixState {
	return &autofixState{
		initial:            make(map[string]string),
		current:            make(map[string]string),
		appliedDiagnostics: make(map[string]int),
	}
}

func (s *autofixState) snapshot() SourceSnapshot {
	paths := make([]string, 0, len(s.current))
	for path := range s.current {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	files := make([]SourceFileSnapshot, 0, len(paths))
	texts := make(map[string]string, len(paths))
	for _, path := range paths {
		text := s.current[path]
		files = append(files, SourceFileSnapshot{Path: path, Text: text})
		texts[path] = text
	}
	return SourceSnapshot{files: files, texts: texts}
}

func (s *autofixState) text(path string) (string, bool) {
	if text, changed := s.current[path]; changed {
		return text, true
	}
	text, known := s.initial[path]
	return text, known
}

func (s *autofixState) apply(changes []FileChange) (FixRoundResult, error) {
	if len(changes) == 0 {
		return FixRoundResult{}, errors.New("linter pipeline: cannot apply an empty in-memory change set")
	}
	seen := make(map[string]struct{}, len(changes))
	paths := make([]string, 0, len(changes))
	for _, change := range changes {
		if change.Path == "" {
			return FixRoundResult{}, errors.New("linter pipeline: in-memory change path must not be empty")
		}
		if _, duplicate := seen[change.Path]; duplicate {
			return FixRoundResult{}, fmt.Errorf("linter pipeline: duplicate in-memory change %q", change.Path)
		}
		seen[change.Path] = struct{}{}
		current, known := s.text(change.Path)
		if !known {
			current = change.Before
		}
		if current != change.Before {
			return FixRoundResult{}, fmt.Errorf("linter pipeline: stale in-memory change for %q", change.Path)
		}
		paths = append(paths, change.Path)
	}

	round := FixRoundResult{
		ChangedPaths: append([]string(nil), paths...),
	}
	for _, change := range changes {
		if _, known := s.initial[change.Path]; !known {
			s.initial[change.Path] = change.Before
		}
		if change.After == s.initial[change.Path] {
			delete(s.current, change.Path)
		} else {
			s.current[change.Path] = change.After
		}
		s.appliedDiagnostics[change.Path] += change.AppliedDiagnostics
		round.AppliedDiagnostics += change.AppliedDiagnostics
	}
	round.RestoredInitial = len(s.current) == 0
	return round, nil
}

func (s *autofixState) finalChanges() []FileChange {
	paths := make([]string, 0, len(s.current))
	for path := range s.current {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	changes := make([]FileChange, 0, len(paths))
	for _, path := range paths {
		changes = append(changes, FileChange{
			Path:               path,
			Before:             s.initial[path],
			After:              s.current[path],
			AppliedDiagnostics: s.appliedDiagnostics[path],
		})
	}
	return changes
}

func (s *autofixState) finalSources() []SourceFileSnapshot {
	paths := make([]string, 0, len(s.appliedDiagnostics))
	for path := range s.appliedDiagnostics {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	files := make([]SourceFileSnapshot, 0, len(paths))
	for _, path := range paths {
		text, known := s.text(path)
		if !known {
			continue
		}
		files = append(files, SourceFileSnapshot{Path: path, Text: text})
	}
	return files
}

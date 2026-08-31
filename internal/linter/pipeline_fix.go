package linter

import (
	"errors"
	"fmt"
	"sort"

	"github.com/web-infra-dev/rslint/internal/rule"
)

// FileChange is a pure whole-file change in stable target path space.
// Persistence, aliases, and protocol projection belong to terminal adapters.
type FileChange struct {
	Path               string
	Before             string
	After              string
	AppliedDiagnostics int
}

// CommitResult identifies paths whose complete final After text is known to
// have reached the external medium.
type CommitResult struct {
	ConfirmedPaths []string
}

type fixTextSnapshot map[string]string

func planFixes(diagnostics []rule.RuleDiagnostic, texts fixTextSnapshot) ([]FileChange, error) {
	byPath := make(map[string][]rule.RuleDiagnostic)
	for _, diagnostic := range diagnostics {
		if len(diagnostic.Fixes()) == 0 {
			continue
		}
		if diagnostic.FilePath == "" {
			return nil, errors.New("linter pipeline: fix diagnostic path must not be empty")
		}
		byPath[diagnostic.FilePath] = append(byPath[diagnostic.FilePath], cloneDiagnosticFixes(diagnostic))
	}
	paths := make([]string, 0, len(byPath))
	for path := range byPath {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	changes := make([]FileChange, 0, len(paths))
	for _, path := range paths {
		text, ok := texts[path]
		if !ok {
			return nil, fmt.Errorf("linter pipeline: fix target %q has no frozen text", path)
		}
		fixed, unapplied, changed := ApplyRuleFixes(text, byPath[path])
		if !changed {
			continue
		}
		changes = append(changes, FileChange{
			Path:               path,
			Before:             text,
			After:              fixed,
			AppliedDiagnostics: len(byPath[path]) - len(unapplied),
		})
	}
	return changes, nil
}

func cloneDiagnosticFixes(diagnostic rule.RuleDiagnostic) rule.RuleDiagnostic {
	if diagnostic.FixesPtr == nil {
		return diagnostic
	}
	fixes := append([]rule.RuleFix(nil), (*diagnostic.FixesPtr)...)
	diagnostic.FixesPtr = &fixes
	return diagnostic
}

func validateCommitResult(
	planned []FileChange,
	result CommitResult,
	commitErr error,
) ([]string, error) {
	plannedByPath := make(map[string]struct{}, len(planned))
	for _, change := range planned {
		if change.Path == "" {
			return nil, errors.Join(commitErr, errors.New("linter pipeline: final change path must not be empty"))
		}
		if _, duplicate := plannedByPath[change.Path]; duplicate {
			return nil, errors.Join(commitErr, fmt.Errorf("linter pipeline: duplicate final change %q", change.Path))
		}
		plannedByPath[change.Path] = struct{}{}
	}

	confirmed := make(map[string]struct{}, len(result.ConfirmedPaths))
	var contractErrors []error
	for _, path := range result.ConfirmedPaths {
		if path == "" {
			contractErrors = append(contractErrors, errors.New("linter pipeline: confirmed path must not be empty"))
			continue
		}
		if _, duplicate := confirmed[path]; duplicate {
			contractErrors = append(contractErrors, fmt.Errorf("linter pipeline: duplicate confirmed path %q", path))
			continue
		}
		if _, plannedPath := plannedByPath[path]; !plannedPath {
			contractErrors = append(contractErrors, fmt.Errorf("linter pipeline: confirmed path %q was not a final change", path))
			continue
		}
		confirmed[path] = struct{}{}
	}
	if commitErr == nil && len(confirmed) != len(plannedByPath) {
		contractErrors = append(contractErrors, fmt.Errorf(
			"linter pipeline: final committer confirmed %d of %d paths without an error",
			len(confirmed), len(plannedByPath),
		))
	}
	if commitErr != nil && len(confirmed) == len(plannedByPath) {
		contractErrors = append(contractErrors, errors.New(
			"linter pipeline: final committer returned an error after confirming the complete change set",
		))
	}
	confirmedPaths := make([]string, 0, len(confirmed))
	for _, change := range planned {
		if _, ok := confirmed[change.Path]; ok {
			confirmedPaths = append(confirmedPaths, change.Path)
		}
	}
	return confirmedPaths, errors.Join(append([]error{commitErr}, contractErrors...)...)
}

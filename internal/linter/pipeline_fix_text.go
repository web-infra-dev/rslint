package linter

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// A fix text snapshot freezes the exact generation text addressed by
// diagnostics. Fix planning can therefore run after generation resources are
// released without falling back to disk or another mutable source.
func freezeFixTextsForDiagnostics(
	ctx context.Context,
	generation Generation,
	snapshot SourceSnapshot,
	diagnostics []rule.RuleDiagnostic,
) (fixTextSnapshot, error) {
	sources, err := fixSourcesFromDiagnostics(diagnostics)
	if err != nil {
		return nil, err
	}
	return freezeFixTexts(ctx, generation, snapshot, sources)
}

func freezeFixTexts(
	ctx context.Context,
	generation Generation,
	snapshot SourceSnapshot,
	sources map[string]ast.SourceFileLike,
) (fixTextSnapshot, error) {
	if len(sources) == 0 {
		return fixTextSnapshot{}, nil
	}
	texts := make(fixTextSnapshot, len(sources))
	paths := make([]string, 0, len(sources))
	for path := range sources {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		source := sources[path]
		if source == nil {
			return nil, fmt.Errorf("linter pipeline: fix target %q has no source frame", path)
		}
		text, err := readGenerationText(generation, snapshot, path, source)
		if err != nil {
			return nil, fmt.Errorf("linter pipeline: read fix target %q: %w", path, err)
		}
		texts[path] = text
	}
	return texts, ctx.Err()
}

func retainFixTextsForDiagnostics(
	texts fixTextSnapshot,
	diagnostics []rule.RuleDiagnostic,
) (fixTextSnapshot, error) {
	paths, err := fixTargetPaths(diagnostics)
	if err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		return fixTextSnapshot{}, nil
	}
	result := make(fixTextSnapshot, len(paths))
	for path := range paths {
		text, ok := texts[path]
		if !ok {
			return nil, fmt.Errorf("linter pipeline: fix target %q has no frozen text", path)
		}
		result[path] = text
	}
	return result, nil
}

func fixTargetPaths(diagnostics []rule.RuleDiagnostic) (map[string]struct{}, error) {
	paths := make(map[string]struct{})
	for _, diagnostic := range diagnostics {
		if len(diagnostic.Fixes()) == 0 {
			continue
		}
		if diagnostic.FilePath == "" {
			return nil, errors.New("linter pipeline: fix diagnostic path must not be empty")
		}
		paths[diagnostic.FilePath] = struct{}{}
	}
	return paths, nil
}

func fixSourcesFromDiagnostics(
	diagnostics []rule.RuleDiagnostic,
) (map[string]ast.SourceFileLike, error) {
	sources := make(map[string]ast.SourceFileLike)
	for _, diagnostic := range diagnostics {
		if len(diagnostic.Fixes()) == 0 {
			continue
		}
		if diagnostic.FilePath == "" {
			return nil, errors.New("linter pipeline: fix diagnostic path must not be empty")
		}
		if diagnostic.SourceFile == nil {
			return nil, fmt.Errorf("linter pipeline: fix target %q has no source frame", diagnostic.FilePath)
		}
		if previous, duplicate := sources[diagnostic.FilePath]; duplicate && previous != diagnostic.SourceFile {
			return nil, fmt.Errorf("linter pipeline: duplicate fix target %q", diagnostic.FilePath)
		}
		sources[diagnostic.FilePath] = diagnostic.SourceFile
	}
	return sources, nil
}

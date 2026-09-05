package linter

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// releaseLease gives every acquired generation exact-once release semantics,
// including early returns, panics, and explicit release before detached work.
type releaseLease struct {
	release ReleaseFunc
	once    sync.Once
}

func (l *releaseLease) close() {
	if l != nil {
		l.once.Do(func() {
			if l.release != nil {
				l.release()
			}
		})
	}
}

// projectGenerationTargets validates the stable target identity of one
// generation, proves that an adapter materialized the exact pipeline-owned
// in-memory state, and optionally returns the target projection requested by a
// result consumer. Keeping those responsibilities in one traversal avoids
// making CLI/LSP materialize the API's source-file result.
func projectGenerationTargets(
	ctx context.Context,
	generation Generation,
	plan *LintPlan,
	snapshot SourceSnapshot,
	collectLintedFiles bool,
) ([]LintedFile, error) {
	if plan == nil {
		if snapshot.Empty() {
			return nil, nil
		}
		return nil, errors.New("linter pipeline: an in-memory snapshot requires a lint plan")
	}
	fileCount := plan.fileCount()
	sources := make(map[string]ast.SourceFileLike, fileCount)
	var lintedFiles []LintedFile
	if collectLintedFiles {
		lintedFiles = make([]LintedFile, 0, fileCount)
	}
	for _, programPlan := range plan.programs {
		for _, filePlan := range programPlan.files {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			source := filePlan.file
			if source == nil {
				continue
			}
			path := projectTargetPath(generation.Target.Path, source.FileName())
			if path == "" {
				return nil, errors.New("linter pipeline: projected target path must not be empty")
			}
			if previous, duplicate := sources[path]; duplicate && previous != source {
				return nil, fmt.Errorf("linter pipeline: duplicate projected target %q", path)
			}
			sources[path] = source
			if collectLintedFiles {
				lintedFiles = append(lintedFiles, LintedFile{Path: path, SourceFile: source})
			}
		}
	}
	if snapshot.Empty() {
		return lintedFiles, nil
	}
	if generation.Target.ReadText == nil {
		return nil, errors.New("linter pipeline: an in-memory snapshot requires a target text reader")
	}
	for _, file := range snapshot.Files() {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		source := sources[file.Path]
		if source == nil {
			return nil, fmt.Errorf("linter pipeline: in-memory target %q is absent from its generation", file.Path)
		}
		text, err := generation.Target.ReadText(file.Path, source)
		if err != nil {
			return nil, fmt.Errorf("linter pipeline: validate in-memory target %q: %w", file.Path, err)
		}
		if text != file.Text {
			return nil, fmt.Errorf("linter pipeline: generation did not materialize in-memory target %q", file.Path)
		}
	}
	return lintedFiles, ctx.Err()
}

func readGenerationText(
	generation Generation,
	snapshot SourceSnapshot,
	targetPath string,
	source ast.SourceFileLike,
) (string, error) {
	if text, changed := snapshot.Text(targetPath); changed {
		return text, nil
	}
	if generation.Target.ReadText == nil {
		return "", errors.New("target text reader is not configured")
	}
	return generation.Target.ReadText(targetPath, source)
}

func projectTargetPath(project func(string) string, sourcePath string) string {
	if project == nil {
		return sourcePath
	}
	return project(sourcePath)
}

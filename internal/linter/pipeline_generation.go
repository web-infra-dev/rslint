package linter

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// releaseLease gives every acquired generation exact-once release semantics,
// including early returns, panics, and explicit release before detached work.
type releaseLease struct {
	release        ReleaseFunc
	once           sync.Once
	mu             sync.Mutex
	readers        int
	closed         bool
	releaseOnOwner func(ReleaseFunc)
}

func (l *releaseLease) close() {
	if l != nil {
		l.once.Do(func() {
			l.mu.Lock()
			l.closed = true
			l.finishLocked(false)
		})
	}
}

// retain gives typed detached work a narrow borrow of the original producer.
// Source-only plugin work still releases it before the reverse request.
func (l *releaseLease) retain() ReleaseFunc {
	l.mu.Lock()
	if l.closed {
		l.mu.Unlock()
		panic("linter pipeline: retain released generation")
	}
	l.readers++
	l.mu.Unlock()
	var once sync.Once
	return func() {
		once.Do(func() {
			l.mu.Lock()
			l.readers--
			l.finishLocked(true)
		})
	}
}

func (l *releaseLease) finishLocked(detached bool) {
	var release ReleaseFunc
	if l.closed && l.readers == 0 {
		release, l.release = l.release, nil
	}
	l.mu.Unlock()
	if release != nil {
		if detached && l.releaseOnOwner != nil {
			l.releaseOnOwner(release)
		} else {
			release()
		}
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
	demand ArtifactDemand,
) ([]LintedFile, error) {
	if plan == nil {
		if snapshot.Empty() {
			return nil, nil
		}
		return nil, errors.New("linter pipeline: an in-memory snapshot requires a lint plan")
	}
	fileCount := plan.fileCount()
	sources := make(map[string]ast.SourceFileLike, fileCount)
	collectLintedFiles := demand.LintedFiles || demand.IncompleteRuleResults || demand.SourceFiles
	var lintedFiles []LintedFile
	if collectLintedFiles {
		lintedFiles = make([]LintedFile, 0, fileCount)
	}
	var skippedByConfig map[ruleSetKey][]string
	if demand.IncompleteRuleResults {
		skippedByConfig = make(map[ruleSetKey][]string)
	}
	for _, programPlan := range plan.programs {
		for index, filePlan := range programPlan.files {
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
				file := LintedFile{Path: path, SyntaxError: filePlan.syntaxError}
				if configs := programPlan.rulesBeforeFilter; skippedByConfig != nil && configs != nil && len(configs[index]) > 0 {
					file.SkippedErrorRules = skippedErrorRules(skippedByConfig, configs[index])
				}
				if demand.SourceFiles {
					file.SourceFile = source
				}
				lintedFiles = append(lintedFiles, file)
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

// Files sharing an immutable config also share its excluded error-rule names.
func skippedErrorRules(cache map[ruleSetKey][]string, rules []rule.ConfiguredRule) []string {
	key := ruleSetKey{first: &rules[0], length: len(rules)}
	if names, ok := cache[key]; ok {
		return names
	}
	var names []string
	for _, configured := range rules {
		if configured.RequiresTypeInfo && configured.Severity == rule.SeverityError {
			names = append(names, configured.Name)
		}
	}
	cache[key] = names
	return names
}

type sourceTextReader func(targetPath string, source ast.SourceFileLike) (string, error)

func readTargetText(
	readText sourceTextReader,
	snapshot SourceSnapshot,
	targetPath string,
	source ast.SourceFileLike,
) (string, error) {
	if text, changed := snapshot.Text(targetPath); changed {
		return text, nil
	}
	return readGenerationText(readText, targetPath, source)
}

func readGenerationText(readText sourceTextReader, targetPath string, source ast.SourceFileLike) (string, error) {
	if readText == nil {
		return "", errors.New("target text reader is not configured")
	}
	return readText(targetPath, source)
}

func projectTargetPath(project func(string) string, sourcePath string) string {
	if project == nil {
		return sourcePath
	}
	return project(sourcePath)
}

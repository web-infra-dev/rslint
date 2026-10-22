package target

import (
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
)

// ExplicitFileOutcome records the target-selection facts needed by a caller
// that reports skipped literal files. Ignored and Exists are deliberately
// independent: the CLI historically reports an ignored missing path as
// ignored, so presentation can preserve that priority without re-reading the
// filesystem or re-running config ownership after the lint plan is frozen.
type ExplicitFileOutcome struct {
	Path    string
	Ignored bool
	Exists  bool
}

type explicitLintTarget struct {
	target    File
	supported bool
	exists    bool
	ignored   bool
	evaluated bool
}

type explicitLintTargetSet struct {
	byPath         map[tspath.Path]*explicitLintTarget
	requestedPaths []string
}

func (target *explicitLintTarget) selectedBy(
	configDirectory string,
	scanRoot string,
	useCaseSensitive bool,
	matcher *rslintconfig.TargetMatcher,
) bool {
	if target == nil {
		return false
	}
	if !target.evaluated {
		target.target.ConfigDirectory = tspath.NormalizePath(configDirectory)
		target.ignored = rslintconfig.IsDefaultExcludedPath(
			target.target.Path,
			scanRoot,
			useCaseSensitive,
		)
		if matcher != nil {
			match := matcher.MatchFile(target.target.Identity())
			if match.GloballyIgnored() {
				target.ignored = true
			}
			if target.supported && target.exists && !target.ignored {
				target.target.match = new(rslintconfig.TargetMatch)
				*target.target.match = match
			}
		}
		target.evaluated = true
	}
	return target.supported && target.exists && !target.ignored
}

func newExplicitLintTargetSet(
	files []rslintconfig.PathIdentity,
	fsys vfs.FS,
) *explicitLintTargetSet {
	if files == nil {
		return nil
	}
	set := &explicitLintTargetSet{
		byPath:         make(map[tspath.Path]*explicitLintTarget, len(files)),
		requestedPaths: make([]string, 0, len(files)),
	}
	for _, identity := range files {
		filePath := tspath.NormalizePath(identity.Path)
		set.requestedPaths = append(set.requestedPaths, filePath)
		pathID := tspath.ToPath(filePath, "", true)
		if _, exists := set.byPath[pathID]; exists {
			continue
		}
		exists := true
		if fsys != nil {
			physicalPath := identity.CanonicalPath
			if physicalPath == "" {
				physicalPath = filePath
			}
			exists = fsys.FileExists(tspath.NormalizePath(physicalPath))
		}
		set.byPath[pathID] = &explicitLintTarget{
			target:    File{PathIdentity: identity},
			supported: rslintconfig.IsSupportedLintFile(filePath),
			exists:    exists,
		}
	}
	return set
}

func (set *explicitLintTargetSet) targetsForIdentities(files []rslintconfig.PathIdentity) []*explicitLintTarget {
	if files == nil {
		return nil
	}
	paths := make([]string, len(files))
	for index, file := range files {
		paths[index] = file.Path
	}
	return set.targetsForPaths(paths)
}

func (set *explicitLintTargetSet) targetsForPaths(filePaths []string) []*explicitLintTarget {
	if filePaths == nil {
		return nil
	}
	targets := make([]*explicitLintTarget, 0, len(filePaths))
	seen := make(map[tspath.Path]struct{}, len(filePaths))
	for _, filePath := range filePaths {
		pathID := tspath.ToPath(tspath.NormalizePath(filePath), "", true)
		if _, exists := seen[pathID]; exists {
			continue
		}
		seen[pathID] = struct{}{}
		if target := set.byPath[pathID]; target != nil {
			targets = append(targets, target)
		}
	}
	return targets
}

func (set *explicitLintTargetSet) outcomes() []ExplicitFileOutcome {
	if set == nil || len(set.requestedPaths) == 0 {
		return nil
	}
	outcomes := make([]ExplicitFileOutcome, 0, len(set.requestedPaths))
	for _, filePath := range set.requestedPaths {
		target := set.byPath[tspath.ToPath(filePath, "", true)]
		if target == nil {
			continue
		}
		outcomes = append(outcomes, ExplicitFileOutcome{
			Path:    filePath,
			Ignored: target.ignored,
			Exists:  target.exists,
		})
	}
	return outcomes
}

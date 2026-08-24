package target

import (
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
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
		if matcher != nil && matcher.MatchFile(target.target.Identity()).GloballyIgnored {
			target.ignored = true
		}
		target.evaluated = true
	}
	return target.supported && target.exists && !target.ignored
}

func newExplicitLintTargetSet(
	filePaths []string,
	fsys vfs.FS,
) *explicitLintTargetSet {
	if filePaths == nil {
		return nil
	}
	set := &explicitLintTargetSet{
		byPath:         make(map[tspath.Path]*explicitLintTarget, len(filePaths)),
		requestedPaths: make([]string, 0, len(filePaths)),
	}
	set.add(filePaths, true, fsys)
	return set
}

func (set *explicitLintTargetSet) add(
	filePaths []string,
	requested bool,
	fsys vfs.FS,
) {
	if set == nil {
		return
	}
	for _, filePath := range filePaths {
		filePath = tspath.NormalizePath(filePath)
		if requested {
			set.requestedPaths = append(set.requestedPaths, filePath)
		}
		pathID := tspath.ToPath(filePath, "", true)
		if _, exists := set.byPath[pathID]; exists {
			continue
		}
		exists := true
		if fsys != nil {
			exists = fsys.FileExists(filePath)
		}
		set.byPath[pathID] = &explicitLintTarget{
			target:    File{PathIdentity: FreezeFileIdentity(filePath, fsys)},
			supported: rslintconfig.IsSupportedLintFile(filePath),
			exists:    exists,
		}
	}
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

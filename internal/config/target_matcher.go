package config

import (
	"errors"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
)

// TargetMatch is the config policy result needed by lint-target planning.
// It intentionally excludes merged config and rule state.
type TargetMatch struct {
	Selected        bool
	GloballyIgnored bool
}

// TargetMatcher evaluates target selection and directory pruning for one
// immutable config and path-space generation.
type TargetMatcher struct {
	resolver *configTargetResolver
}

// NewTargetMatcherWithPathSpaces binds config matching to an existing frozen
// generation. Missing authored bases are rejected instead of being read from
// the filesystem as an implicit fallback.
func NewTargetMatcherWithPathSpaces(
	config RslintConfig,
	configDirectory string,
	fsys vfs.FS,
	pathSpaces *PathSpaceSnapshot,
) (TargetMatcher, error) {
	if pathSpaces == nil {
		return TargetMatcher{}, errors.New("path-space snapshot is required")
	}
	if _, err := pathSpaces.requireBase(configDirectory); err != nil {
		return TargetMatcher{}, err
	}
	for _, entry := range config {
		origin := configEntryPathOrigin(entry, configDirectory)
		if _, err := pathSpaces.requireBase(origin.directory); err != nil {
			return TargetMatcher{}, err
		}
		if origin.basePathScoped {
			if _, err := pathSpaces.requireBase(origin.configArrayBase); err != nil {
				return TargetMatcher{}, err
			}
		}
	}
	return TargetMatcher{resolver: newConfigTargetResolverWithBases(config, configDirectory, fsys, pathSpaces.bases)}, nil
}

// MatchFile evaluates files, ignores, and the implicit extension baseline.
func (matcher TargetMatcher) MatchFile(target PathIdentity) TargetMatch {
	if matcher.resolver == nil {
		return TargetMatch{}
	}
	decision := matcher.resolver.resolveTarget(target)
	return TargetMatch{
		Selected:        decision.selected,
		GloballyIgnored: decision.globallyIgnored,
	}
}

// CanPruneDirectory reports whether no descendant can be selected after
// applying ordered global ignores.
func (matcher TargetMatcher) CanPruneDirectory(directory DirectoryIdentity) bool {
	return matcher.resolver != nil && matcher.resolver.canPruneDirectory(
		tspath.NormalizePath(directory.LexicalPath),
		tspath.NormalizePath(directory.CanonicalPath),
	)
}

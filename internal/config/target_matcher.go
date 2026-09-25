package config

import (
	"errors"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
)

// TargetMatch is an immutable selection result for one target and config
// generation. Target planning may carry it to a FileConfigResolver, but cannot
// inspect or change the matched entries. It carries no effective config or
// configured rule state.
type TargetMatch struct {
	target   PathIdentity
	source   *targetMatchSource
	decision configTargetDecision
}

// Selected reports whether selectors or the implicit extension baseline select
// the target. Global ignores take precedence over selection.
func (match TargetMatch) Selected() bool { return match.decision.selected }

// GloballyIgnored reports whether the target must be skipped entirely.
func (match TargetMatch) GloballyIgnored() bool { return match.decision.globallyIgnored }

// TargetMatcher evaluates target selection and directory pruning for one
// immutable config and path-space generation.
type TargetMatcher struct {
	resolver *configTargetResolver
	source   *targetMatchSource
}

// targetMatchSource binds a config array to its owner and path-space generation.
// It excludes the matcher itself so carried matches cannot retain the walk's
// directory caches or filesystem wrappers.
type targetMatchSource struct {
	config           RslintConfig
	configDirectory  string
	pathSpaces       *PathSpaceSnapshot
	useCaseSensitive bool
	hasFS            bool
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
	resolver := newConfigTargetResolverWithBases(config, configDirectory, fsys, pathSpaces.bases)
	return TargetMatcher{
		resolver: resolver,
		source: &targetMatchSource{
			config:           config,
			configDirectory:  normalizeAuthoredBase(configDirectory),
			pathSpaces:       pathSpaces,
			useCaseSensitive: resolver.useCaseSensitive,
			hasFS:            fsys != nil,
		},
	}, nil
}

// MatchFile evaluates files, ignores, and the implicit extension baseline.
func (matcher TargetMatcher) MatchFile(target PathIdentity) TargetMatch {
	if matcher.resolver == nil {
		return TargetMatch{}
	}
	return TargetMatch{
		target:   target,
		source:   matcher.source,
		decision: matcher.resolver.resolveTarget(target),
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

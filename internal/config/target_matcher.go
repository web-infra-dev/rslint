package config

import (
	"errors"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/config/gitignore"
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
	gitPolicy        *GitPolicy
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
	gitPolicy *GitPolicy,
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
	resolver := newConfigTargetResolverWithBases(config, configDirectory, fsys, pathSpaces.bases, gitPolicy)
	return TargetMatcher{
		resolver: resolver,
		source: &targetMatchSource{
			config:           config,
			configDirectory:  normalizeAuthoredBase(configDirectory),
			pathSpaces:       pathSpaces,
			gitPolicy:        gitPolicy,
			useCaseSensitive: resolver.useCaseSensitive,
			hasFS:            fsys != nil,
		},
	}, nil
}

// MatchFile evaluates files, ignores, and the implicit extension baseline.
func (matcher TargetMatcher) MatchFile(target PathIdentity) TargetMatch {
	return matcher.MatchFileWithGitEntryBlock(target, false)
}

// MatchFileWithGitEntryBlock preserves the current walk's outer Git admission
// restriction even when the file's own repository has different raw rules.
func (matcher TargetMatcher) MatchFileWithGitEntryBlock(target PathIdentity, gitEntryBlocked bool) TargetMatch {
	if matcher.resolver == nil {
		return TargetMatch{}
	}
	return TargetMatch{
		target:   target,
		source:   matcher.source,
		decision: matcher.resolver.resolveTargetWithGitEntryBlock(target, gitEntryBlocked),
	}
}

// GitEntryBlocked retains the current scan route's outer admission gate. A
// negation admitting a nested Git root can reset that gate; a possible leaf
// exception only keeps traversal open. Compute this once per directory.
func (matcher TargetMatcher) GitEntryBlocked(directory, scanRoot DirectoryIdentity) bool {
	if matcher.resolver == nil {
		return false
	}
	policy := matcher.resolver.gitPolicy
	blocked := policy.blockedDirectoryOnRoute(directory.CanonicalPath, scanRoot.CanonicalPath)
	if blocked == "" {
		return false
	}
	for physical, lexical := directory.CanonicalPath, directory.LexicalPath; physical != ""; physical, lexical = gitignore.ParentDirectory(physical), tspath.GetDirectoryPath(lexical) {
		_, repositoryRoot := policy.roots[policy.pathID(physical)]
		if repositoryRoot && matcher.resolver.reopensDirectory(lexical, physical) {
			return false
		}
		if policy.pathID(physical) == policy.pathID(blocked) {
			return true
		}
	}
	return false
}

// CanPruneDirectory reports whether no descendant can be selected after
// applying ordered global ignores.
func (matcher TargetMatcher) CanPruneDirectory(directory DirectoryIdentity) bool {
	return matcher.resolver != nil && matcher.resolver.canPruneDirectory(
		tspath.NormalizePath(directory.LexicalPath),
		tspath.NormalizePath(directory.CanonicalPath),
	)
}

// CanPruneDirectoryEntry additionally respects the containing repository's
// directory admission decision. Explicit walk roots use CanPruneDirectory;
// walking into a child must not borrow another input's independent entry point.
func (matcher TargetMatcher) CanPruneDirectoryEntry(directory DirectoryIdentity) bool {
	return matcher.resolver != nil && matcher.resolver.canPruneDirectoryWithEntry(
		tspath.NormalizePath(directory.LexicalPath),
		tspath.NormalizePath(directory.CanonicalPath),
		true,
	)
}

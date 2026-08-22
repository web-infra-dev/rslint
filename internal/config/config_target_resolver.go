package config

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
)

// configTargetResolver matches one immutable flat-config array while retaining
// the authored base of every composed entry. Most entries share the owning
// config directory; API overrideConfig entries may instead use invocation cwd.
// Each distinct base resolves the target identity once per decision, so target
// selection and effective config merging cannot drift onto different spaces.
type configTargetResolver struct {
	config             RslintConfig
	fs                 vfs.FS
	bases              []configTargetBase
	entries            []configTargetEntry
	globalEntryIndexes []int
	directoryBlocks    publishOnceCache[configTargetIdentity, bool]
}

type configTargetBase struct {
	directory              string
	physicalDirectory      string
	physicalAliasAncestors []string
}

type configTargetEntry struct {
	baseIndex           int
	ignorePatterns      []IgnorePattern
	gitScopes           []collectedGitignoreScope
	ignorePatternGroups []ignorePatternCoordinateGroup
}

// ignorePatternCoordinateGroup is one contiguous ignore-pattern segment whose
// decisions use the same target coordinate and matching semantics. The
// grouping is frozen with the resolver so both file matching and directory
// pruning project their path once per authored root instead of once per
// pattern.
type ignorePatternCoordinateGroup struct {
	patterns      []IgnorePattern
	negationReach negReach
}

type configTargetMatch struct {
	path      string
	directory string
	matcher   fileMatchPath
}

// configTargetIdentity is the immutable caller-visible/canonical pair for one
// target decision. Entry matching may project this pair into several authored
// config bases, but root-scoped matchers such as collected .gitignore rules
// must always derive their coordinates from this original identity.
type configTargetIdentity struct {
	path               string
	canonicalMatchPath string
}

type configTargetDecision struct {
	key             configMatchKey
	matched         bool
	selected        bool
	globallyIgnored bool
}

func newConfigTargetResolver(config RslintConfig, defaultDirectory string, fsys vfs.FS) *configTargetResolver {
	return newConfigTargetResolverWithBases(config, defaultDirectory, fsys, nil)
}

func newConfigTargetResolverWithBases(
	config RslintConfig,
	defaultDirectory string,
	fsys vfs.FS,
	frozenBases map[string]configTargetBase,
) *configTargetResolver {
	resolver := &configTargetResolver{
		config:  config,
		fs:      fsys,
		entries: make([]configTargetEntry, len(config)),
	}
	baseIndexByDirectory := make(map[string]int)
	for index, entry := range config {
		directory := tspath.NormalizePath(configEntryBaseDirectory(entry, defaultDirectory))
		baseID := exactPathID(directory)
		baseIndex, exists := baseIndexByDirectory[baseID]
		if !exists {
			base, frozen := frozenBases[baseID]
			if !frozen {
				base = freezeConfigTargetBase(directory, fsys)
			}
			baseIndex = len(resolver.bases)
			baseIndexByDirectory[baseID] = baseIndex
			resolver.bases = append(resolver.bases, base)
		}
		patterns := configEntryIgnorePatterns(entry)
		resolver.entries[index] = configTargetEntry{
			baseIndex:      baseIndex,
			ignorePatterns: patterns,
		}
		if entry.collectedGitignore != nil {
			resolver.entries[index].gitScopes = entry.collectedGitignore.scopes
		}
		if isGlobalIgnoreEntry(entry) {
			resolver.entries[index].ignorePatternGroups = buildIgnorePatternCoordinateGroups(patterns)
			resolver.globalEntryIndexes = append(resolver.globalEntryIndexes, index)
		}
	}
	return resolver
}

func freezeConfigTargetBase(directory string, fsys vfs.FS) configTargetBase {
	directory = tspath.NormalizePath(directory)
	physicalDirectory := directory
	if fsys != nil {
		if realPath := fsys.Realpath(directory); realPath != "" {
			physicalDirectory = tspath.NormalizePath(realPath)
		}
	}
	return configTargetBase{
		directory:              directory,
		physicalDirectory:      physicalDirectory,
		physicalAliasAncestors: verifiedConfigAliasAncestors(directory, physicalDirectory, fsys),
	}
}

func configEntryIgnorePatterns(entry ConfigEntry) []IgnorePattern {
	if entry.collectedGitignore != nil {
		return entry.collectedGitignore.ignores
	}
	return ParseIgnorePatterns(entry.Ignores)
}

func buildIgnorePatternCoordinateGroups(patterns []IgnorePattern) []ignorePatternCoordinateGroup {
	var groups []ignorePatternCoordinateGroup
	for start := 0; start < len(patterns); {
		end := start + 1
		for end < len(patterns) && patternsShareTargetCoordinate(patterns[start], patterns[end]) {
			end++
		}
		groupPatterns := patterns[start:end]
		groups = append(groups, ignorePatternCoordinateGroup{
			patterns:      groupPatterns,
			negationReach: buildNegReach(groupPatterns),
		})
		start = end
	}
	return groups
}

func patternsShareTargetCoordinate(left IgnorePattern, right IgnorePattern) bool {
	if left.GitPattern != right.GitPattern || left.CaseInsensitive != right.CaseInsensitive {
		return false
	}
	leftScoped := patternHasMatchDirectory(left)
	rightScoped := patternHasMatchDirectory(right)
	if leftScoped != rightScoped {
		return false
	}
	if !leftScoped {
		return true
	}
	return left.gitScope == right.gitScope &&
		left.MatchDirectory == right.MatchDirectory &&
		left.PhysicalMatchDirectory == right.PhysicalMatchDirectory &&
		left.LexicalMatchDirectory == right.LexicalMatchDirectory
}

func (resolver *configTargetResolver) resolve(path string, canonicalPath string) configTargetDecision {
	return resolver.resolveTarget(DiscoveredLintTarget{
		Path:          path,
		CanonicalPath: canonicalPath,
	})
}

func (resolver *configTargetResolver) resolveTarget(
	discovered DiscoveredLintTarget,
) configTargetDecision {
	target, matches := resolver.resolvePathSpacesWithCanonicalParent(
		discovered.Path,
		discovered.CanonicalPath,
		discovered.CanonicalParentPath,
		false,
	)

	decision := configTargetDecision{selected: isDefaultLintFile(target.path)}
	decision.globallyIgnored = resolver.globallyIgnores(target, matches)
	if decision.globallyIgnored {
		return decision
	}

	var tail []byte
	if tailEntryCount := len(resolver.config) - 64; tailEntryCount > 0 {
		tail = make([]byte, (tailEntryCount+7)/8)
	}
	for index, entry := range resolver.config {
		if isGlobalIgnoreEntry(entry) {
			continue
		}
		prepared := resolver.entries[index]
		match := &matches[prepared.baseIndex]
		if hasFileSelectors(entry) && !match.matcher.matchesConfigEntry(entry) {
			continue
		}
		if match.matcher.isIgnored(prepared.ignorePatterns) {
			continue
		}

		decision.matched = true
		if hasFileSelectors(entry) {
			decision.selected = true
		}
		decision.key.add(index, tail)
	}
	if tail != nil {
		decision.key.tail = string(tail)
	}
	return decision
}

func (resolver *configTargetResolver) resolvePathSpaces(
	path string,
	canonicalPath string,
	directory bool,
) (configTargetIdentity, []configTargetMatch) {
	return resolver.resolvePathSpacesWithCanonicalParent(
		path,
		canonicalPath,
		"",
		directory,
	)
}

func (resolver *configTargetResolver) resolvePathSpacesWithCanonicalParent(
	path string,
	canonicalPath string,
	canonicalParentPath string,
	directory bool,
) (configTargetIdentity, []configTargetMatch) {
	path = tspath.NormalizePath(path)
	if canonicalPath != "" {
		canonicalPath = tspath.NormalizePath(canonicalPath)
	} else if resolver.fs != nil && resolver.needsCanonicalPath(path) {
		if realPath := resolver.fs.Realpath(path); realPath != "" {
			canonicalPath = tspath.NormalizePath(realPath)
		}
	}
	if canonicalParentPath != "" {
		canonicalParentPath = tspath.NormalizePath(canonicalParentPath)
	}
	matches := make([]configTargetMatch, len(resolver.bases))
	for index, base := range resolver.bases {
		matchPath, matchDirectory := resolveConfigPathSpaceWithCanonicalParent(
			path,
			canonicalPath,
			canonicalParentPath,
			base.directory,
			base.physicalDirectory,
			base.physicalAliasAncestors,
			resolver.fs,
			!directory,
		)
		matches[index] = configTargetMatch{
			path:      matchPath,
			directory: matchDirectory,
			matcher:   newFileMatchPath(matchPath, matchDirectory),
		}
	}
	canonicalMatchPath := canonicalPath
	if canonicalMatchPath == "" && resolver.fs == nil {
		// Filesystem-free compatibility callers cannot distinguish a canonical
		// spelling from an ordinary lexical one. After lexical scope selection
		// has failed, allow the supplied path itself to represent a physical
		// spelling. Filesystem-aware paths still protect leaf symlinks below.
		canonicalMatchPath = path
	}
	if !directory && canonicalMatchPath != "" {
		identityAlreadyExact := canonicalMatchPath == path &&
			(canonicalParentPath == "" || canonicalParentPath == tspath.GetDirectoryPath(path))
		if !identityAlreadyExact && hasDistinctFileIdentityWithCanonicalParent(
			path,
			canonicalMatchPath,
			canonicalParentPath,
			resolver.fs,
		) {
			canonicalMatchPath = ""
		}
	}
	target := configTargetIdentity{
		path:               path,
		canonicalMatchPath: canonicalMatchPath,
	}
	return target, matches
}

// needsCanonicalPath reports whether lexical containment alone cannot resolve
// every authored config base and collected Git scope. Directory walks normally
// stay inside both, so they avoid a redundant filesystem Realpath per node;
// aliases, native-casing fallbacks, external configs, and independent Git
// scopes still request the canonical identity before the shared decision.
func (resolver *configTargetResolver) needsCanonicalPath(path string) bool {
	for _, base := range resolver.bases {
		if isPathWithinNormalizedRoot(path, base.directory) {
			continue
		}
		if !isPathWithinNormalizedRoot(path, base.physicalDirectory) {
			return true
		}
	}
	for _, entry := range resolver.entries {
		if len(entry.gitScopes) == 0 {
			continue
		}
		withinLexicalScope := false
		for _, scope := range entry.gitScopes {
			if (!scope.caseInsensitive && isPathWithinNormalizedRoot(path, scope.lexicalDirectory)) ||
				(scope.caseInsensitive && isPathWithinRootCaseInsensitive(path, scope.lexicalDirectory)) {
				withinLexicalScope = true
				break
			}
		}
		if !withinLexicalScope {
			return true
		}
	}
	return false
}

func isPathWithinRootCaseInsensitive(path string, root string) bool {
	_, within := RelativePathWithinConfigRoot(path, root, false)
	return within
}

// selectGitScope chooses exactly one collection scope for a target. Lexical
// containment is authoritative. Canonical-to-physical containment is only a
// fallback for targets that have no lexical scope, and uses the same stable
// scope order produced by collection planning. A projected config-owner path
// never participates in authority selection.
func (entry configTargetEntry) selectGitScope(target configTargetIdentity) uint32 {
	for index, scope := range entry.gitScopes {
		if (!scope.caseInsensitive && isPathWithinNormalizedRoot(target.path, scope.lexicalDirectory)) ||
			(scope.caseInsensitive && isPathWithinRootCaseInsensitive(target.path, scope.lexicalDirectory)) {
			return uint32(index + 1)
		}
	}
	if target.canonicalMatchPath == "" {
		return 0
	}
	for index, scope := range entry.gitScopes {
		if (!scope.caseInsensitive && isPathWithinNormalizedRoot(target.canonicalMatchPath, scope.physicalDirectory)) ||
			(scope.caseInsensitive && isPathWithinRootCaseInsensitive(target.canonicalMatchPath, scope.physicalDirectory)) {
			return uint32(index + 1)
		}
	}
	return 0
}

func (resolver *configTargetResolver) globallyIgnores(
	target configTargetIdentity,
	matches []configTargetMatch,
) bool {
	ignored := false
	if resolver.hasAbsoluteDirectoryBlock(target, matches, true) {
		return true
	}
	for _, entryIndex := range resolver.globalEntryIndexes {
		entry := resolver.entries[entryIndex]
		match := &matches[entry.baseIndex]
		ignored = target.applyIgnorePatternGroups(
			match,
			entry.ignorePatternGroups,
			entry.selectGitScope(target),
			ignored,
		)
	}
	return ignored
}

func (target configTargetIdentity) applyIgnorePatternGroups(
	match *configTargetMatch,
	groups []ignorePatternCoordinateGroup,
	selectedGitScope uint32,
	ignored bool,
) bool {
	for _, group := range groups {
		patterns := group.patterns
		var normalizedPath, unixPath string
		var ok bool
		if patternHasMatchDirectory(patterns[0]) {
			if patterns[0].gitScope != 0 && patterns[0].gitScope != selectedGitScope {
				continue
			}
			normalizedPath, unixPath, ok = target.pathsForPatternGroup(match, patterns)
		} else {
			normalizedPath, unixPath, ok = match.matcher.pathsForPatternGroup(patterns)
		}
		if ok {
			ignored = applyIgnorePatternGroup(patterns, normalizedPath, unixPath, ignored)
		}
	}
	return ignored
}

func (target configTargetIdentity) pathsForPatternGroup(
	match *configTargetMatch,
	patterns []IgnorePattern,
) (string, string, bool) {
	pattern := patterns[0]
	lexicalRoot := pattern.LexicalMatchDirectory
	if lexicalRoot == "" {
		lexicalRoot = pattern.MatchDirectory
	}
	if target.path == match.path && lexicalRoot == match.directory {
		normalizedPath, unixPath := match.matcher.normalizedPaths()
		if pathEscapesCwd(unixPath) && hasCaseInsensitivePattern(patterns) {
			normalizedPath = normalizePathWithCaseSensitivity(match.path, match.directory, false)
			unixPath = strings.ReplaceAll(normalizedPath, "\\", "/")
		}
		if !pathEscapesCwd(unixPath) {
			return normalizedPath, unixPath, true
		}
	}
	return pathsForGitPatternGroup(
		target.path,
		target.canonicalMatchPath,
		match.path,
		patterns,
	)
}

func (resolver *configTargetResolver) blocksDirectory(path string, canonicalPath string) bool {
	target, matches := resolver.resolvePathSpaces(path, canonicalPath, true)
	return resolver.hasAbsoluteDirectoryBlock(target, matches, false)
}

func (resolver *configTargetResolver) hasAbsoluteDirectoryBlock(
	target configTargetIdentity,
	matches []configTargetMatch,
	fileTarget bool,
) bool {
	if fileTarget {
		target.path = tspath.GetDirectoryPath(target.path)
		if target.canonicalMatchPath != "" {
			target.canonicalMatchPath = tspath.GetDirectoryPath(target.canonicalMatchPath)
		}
		return resolver.directoryBlocks.getOrInit(target, func() bool {
			return resolver.hasAbsoluteDirectoryBlockForDirectory(target, matches, true)
		})
	}
	return resolver.hasAbsoluteDirectoryBlockForDirectory(target, matches, false)
}

func (resolver *configTargetResolver) hasAbsoluteDirectoryBlockForDirectory(
	target configTargetIdentity,
	matches []configTargetMatch,
	fileTarget bool,
) bool {
	for _, entryIndex := range resolver.globalEntryIndexes {
		entry := resolver.entries[entryIndex]
		match := &matches[entry.baseIndex]
		selectedGitScope := entry.selectGitScope(target)
		directoryMatch := match
		if fileTarget {
			path := tspath.GetDirectoryPath(match.path)
			directoryMatch = &configTargetMatch{
				path:      path,
				directory: match.directory,
				matcher:   newFileMatchPath(path, match.directory),
			}
		}
		for _, group := range entry.ignorePatternGroups {
			relative, ok := targetPathForPatternGroup(
				target,
				directoryMatch,
				group.patterns,
				selectedGitScope,
			)
			if ok && isDirAbsolutelyBlocked(relative, group.patterns) {
				return true
			}
		}
	}
	return false
}

// canPruneDirectory proves that no global-ignore negation can select a file
// below path. Unlike the final file decision, pruning must stay conservative:
// one reachable negation in any authored path space keeps the subtree open.
func (resolver *configTargetResolver) canPruneDirectory(path string, canonicalPath string) bool {
	target, matches := resolver.resolvePathSpaces(path, canonicalPath, true)
	negationCanReach := false
	fileLevelCovered := false
	for _, entryIndex := range resolver.globalEntryIndexes {
		entry := resolver.entries[entryIndex]
		match := &matches[entry.baseIndex]
		selectedGitScope := entry.selectGitScope(target)
		for _, group := range entry.ignorePatternGroups {
			relative, ok := targetPathForPatternGroup(
				target,
				match,
				group.patterns,
				selectedGitScope,
			)
			if !ok {
				continue
			}
			if isDirAbsolutelyBlocked(relative, group.patterns) {
				return true
			}
			if !negationCanReach && group.negationReach.overlaps(relative) {
				negationCanReach = true
			}
			if !fileLevelCovered {
				for _, pattern := range group.patterns {
					if !pattern.Negated && pattern.Kind == dirFileLevelCover &&
						ignorePatternMatches(pattern, relative+"/x") {
						fileLevelCovered = true
						break
					}
				}
			}
		}
	}
	return fileLevelCovered && !negationCanReach
}

func targetPathForPatternGroup(
	target configTargetIdentity,
	match *configTargetMatch,
	patterns []IgnorePattern,
	selectedGitScope uint32,
) (string, bool) {
	pattern := patterns[0]
	if !patternHasMatchDirectory(pattern) {
		_, relative := match.matcher.normalizedPaths()
		relative = strings.TrimSuffix(relative, "/")
		if pathEscapesCwd(relative) && hasCaseInsensitivePattern(patterns) {
			relative = normalizePathWithCaseSensitivity(match.path, match.directory, false)
			relative = strings.TrimSuffix(strings.ReplaceAll(relative, "\\", "/"), "/")
		}
		return relative, relative != "" && relative != "."
	}
	if pattern.gitScope != 0 && pattern.gitScope != selectedGitScope {
		return "", false
	}

	_, relative, ok := target.pathsForPatternGroup(match, patterns)
	if !ok {
		return "", false
	}
	relative = strings.TrimSuffix(relative, "/")
	return relative, relative != "" && relative != "."
}

func (resolver *configTargetResolver) reopensDirectory(path string, canonicalPath string) bool {
	target, matches := resolver.resolvePathSpaces(path, canonicalPath, true)
	reopened := false
	for _, entryIndex := range resolver.globalEntryIndexes {
		entry := resolver.entries[entryIndex]
		match := &matches[entry.baseIndex]
		selectedGitScope := entry.selectGitScope(target)
		for _, group := range entry.ignorePatternGroups {
			relativePath, ok := targetPathForPatternGroup(
				target,
				match,
				group.patterns,
				selectedGitScope,
			)
			if !ok {
				continue
			}
			for patternIndex := range group.patterns {
				pattern := &group.patterns[patternIndex]
				matched := false
				if pattern.GitPattern {
					gitPath := relativePath
					if pattern.CaseInsensitive {
						gitPath = strings.ToLower(gitPath)
					}
					matched = gitIgnorePatternMatchesNode(pattern, gitPath)
				} else {
					matched = ignorePatternMatches(*pattern, relativePath) ||
						ignorePatternMatches(*pattern, relativePath+"/")
				}
				if matched {
					reopened = pattern.Negated
				}
			}
		}
	}
	return reopened
}

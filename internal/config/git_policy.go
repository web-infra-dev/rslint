package config

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/config/gitignore"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/pathutil"
)

type gitPolicySource struct {
	patterns      []IgnorePattern
	negationReach negReach
}

// GitPolicy is the immutable Git observation and matcher generation shared by
// discovery, target selection and effective configuration. It owns no VFS or
// config owner: all paths are physical and all observations precede matching.
type GitPolicy struct {
	useCaseSensitive bool
	roots            map[tspath.PathKey]string
	sources          map[tspath.PathKey]gitPolicySource
	covered          map[tspath.PathKey]struct{}
	blockedEntries   map[tspath.PathKey]struct{}
}

func NewGitPolicy(snapshot gitignore.Snapshot) *GitPolicy {
	policy := &GitPolicy{
		useCaseSensitive: snapshot.UseCaseSensitive,
		roots:            make(map[tspath.PathKey]string),
		sources:          make(map[tspath.PathKey]gitPolicySource),
		covered:          make(map[tspath.PathKey]struct{}),
		blockedEntries:   make(map[tspath.PathKey]struct{}),
	}
	for _, directory := range snapshot.Roots {
		policy.roots[policy.pathID(directory)] = tspath.NormalizePath(directory)
	}
	for _, source := range snapshot.Sources {
		if len(source.Patterns) == 0 {
			continue
		}
		patterns := parseCollectedGitignorePatterns(source.Patterns, !snapshot.UseCaseSensitive)
		policy.sources[policy.pathID(source.Directory)] = gitPolicySource{
			patterns: patterns, negationReach: buildNegReach(patterns),
		}
	}
	for _, directory := range snapshot.CoveredDirectories {
		policy.covered[policy.pathID(directory)] = struct{}{}
	}
	for _, directory := range snapshot.BlockedDirectoryEntries {
		policy.blockedEntries[policy.pathID(directory)] = struct{}{}
	}
	return policy
}

// blocksDirectoryEntry retains the outer walk's admission decision even when
// another input independently establishes this directory as a Git root.
func (policy *GitPolicy) blocksDirectoryEntry(physicalDir string) bool {
	if policy == nil {
		return false
	}
	_, blocked := policy.blockedEntries[policy.pathID(physicalDir)]
	return blocked
}

// blockedDirectoryOnRoute finds the nearest raw admission restriction within
// the scan entry's initial repository, including across later nested roots.
func (policy *GitPolicy) blockedDirectoryOnRoute(physicalDir, scanRoot string) string {
	if policy == nil || len(policy.blockedEntries) == 0 {
		return ""
	}
	root, _ := policy.repository(scanRoot, true)
	if root == "" {
		return ""
	}
	rootID := policy.pathID(root)
	for current := physicalDir; current != "" && policy.pathID(current) != rootID; current = gitignore.ParentDirectory(current) {
		if policy.blocksDirectoryEntry(current) {
			return current
		}
	}
	return ""
}

func (policy *GitPolicy) pathID(path string) tspath.PathKey {
	return gitignore.DirectoryPathID(path, policy.useCaseSensitive)
}

// CoversDirectory reports a completed chain observation for this exact
// physical directory. A known root does not imply coverage of its descendants.
func (policy *GitPolicy) CoversDirectory(physicalDir string) bool {
	if policy == nil {
		return false
	}
	_, covered := policy.covered[policy.pathID(physicalDir)]
	return covered
}

// repository locates the nearest observed root without allocating per-target
// source lists. Sources are evaluated directly along the same ancestor chain.
func (policy *GitPolicy) repository(physicalPath string, directory bool) (string, string) {
	if policy == nil {
		return "", ""
	}
	start := physicalPath
	if !directory {
		start = tspath.GetDirectoryPath(start)
	}
	for current := start; current != ""; current = gitignore.ParentDirectory(current) {
		if root, exists := policy.roots[policy.pathID(current)]; exists {
			return root, start
		}
	}
	return "", ""
}

func (policy *GitPolicy) ignored(physicalPath string) bool {
	root, directory := policy.repository(physicalPath, false)
	if root == "" {
		return false
	}
	relative, within := RelativePathWithinConfigRoot(physicalPath, root, policy.useCaseSensitive)
	if !within {
		return false
	}
	state := newGitIgnorePathState(relative)
	// The containing source starts at the file node. Each parent source can
	// decide one more ancestor node, while never deciding its own directory.
	state.minimumDepth = state.fileDepth
	rootID := policy.pathID(root)
	for current := directory; current != ""; current = gitignore.ParentDirectory(current) {
		id := policy.pathID(current)
		if source, exists := policy.sources[id]; exists && state.evaluate(source.patterns) {
			return true
		}
		if id == rootID {
			break
		}
		state.minimumDepth--
	}
	return false
}

func (policy *GitPolicy) directoryCoverage(physicalDir string) (covered, negationCanReach bool) {
	root, directory := policy.repository(physicalDir, true)
	if root == "" {
		return false, false
	}
	relative, within := RelativePathWithinConfigRoot(physicalDir, root, policy.useCaseSensitive)
	if !within || relative == "" {
		return false, false
	}
	minimumDepth := 1 + strings.Count(relative, "/")
	rootID := policy.pathID(root)
	for current := directory; current != ""; current = gitignore.ParentDirectory(current) {
		id := policy.pathID(current)
		source := policy.sources[id]
		negationCanReach = negationCanReach || source.negationReach.overlaps(relative)
		if !covered {
			for _, pattern := range source.patterns {
				if !pattern.Negated && pattern.Kind == dirFileLevelCover && ignorePatternMatches(pattern, relative+"/x") {
					// The projection also matches a directory-only wildcard's
					// synthetic leaf. Verify actual directory nodes before treating
					// that as coverage of every file in the directory.
					state := newGitIgnorePathState(relative + "/x")
					state.minimumDepth = minimumDepth
					if state.apply(&pattern) {
						covered = true
						break
					}
				}
			}
		}
		if id == rootID {
			break
		}
		minimumDepth--
	}
	return covered, negationCanReach
}

// GitPath preserves the original directory entry name even for a leaf symlink
// whose target belongs to another repository.
func GitPath(target PathIdentity) string {
	parent := target.CanonicalParentPath
	if parent == "" {
		parent = tspath.GetDirectoryPath(target.Path)
	}
	return tspath.CombinePaths(parent, tspath.GetBaseFileName(target.Path))
}

// CollectGitPolicy is the direct-configuration entrypoint. It observes only
// supplied file chains and recursive roots; callers supply any implicit cwd
// root explicitly. Authored configuration is never modified.
func CollectGitPolicy(
	authored RslintConfig,
	configDirectory string,
	fsys vfs.FS,
	files []PathIdentity,
	directories []DirectoryIdentity,
) *GitPolicy {
	reader := gitignore.NewReader(fsys)
	for _, file := range files {
		directory := tspath.GetDirectoryPath(GitPath(file))
		reader.Observe(reader.Ancestors(directory), directory)
	}
	matcher := NewGlobalIgnoreMatcher(authored, configDirectory, fsys)
	var visit func(string, string, gitignore.Cursor)
	visit = func(lexical, physical string, cursor gitignore.Cursor) {
		if IsDefaultExcludedPath(lexical, "", fsys == nil || fsys.CaseSensitivity().IsCaseSensitive()) || matcher.CanPruneDirectory(lexical, physical) {
			return
		}
		cursor = reader.Observe(cursor, physical)
		if fsys == nil {
			return
		}
		entries := fsys.GetAccessibleEntries(pathutil.Directory(physical))
		for _, name := range entries.Directories {
			if excludedGitDirectory(name, fsys.CaseSensitivity().IsCaseSensitive()) {
				continue
			}
			childLexical := tspath.CombinePaths(lexical, name)
			if entries.Symlinks != nil {
				if _, symlink := entries.Symlinks[name]; symlink {
					continue
				}
			} else {
				if resolved := fsys.Realpath(pathutil.Path(tspath.CombinePaths(physical, name))); resolved != "" &&
					!PathsEqual(resolved.AsString(), tspath.CombinePaths(physical, name), fsys.CaseSensitivity().IsCaseSensitive()) {
					continue
				}
			}
			childPhysical := tspath.CombinePaths(physical, name)
			childCursor, entryBlocked := reader.Enter(cursor, childPhysical)
			if (entryBlocked || !childCursor.SourceReachable()) && !matcher.MayReopenGitDescendants(childLexical, childPhysical) {
				reader.Observe(childCursor.BlockSourceTraversal(), childPhysical)
				continue
			}
			visit(childLexical, childPhysical, childCursor)
		}
	}
	for _, directory := range directories {
		physical := directory.CanonicalPath
		if physical == "" {
			physical = directory.LexicalPath
		}
		cursor := reader.Ancestors(physical)
		if !cursor.SourceReachable() && !matcher.MayReopenGitDescendants(directory.LexicalPath, physical) {
			reader.Observe(cursor, physical)
			continue
		}
		visit(directory.LexicalPath, physical, cursor)
	}
	return NewGitPolicy(reader.Snapshot())
}

func excludedGitDirectory(name string, useCaseSensitive bool) bool {
	for _, excluded := range utils.DefaultExcludeDirNames {
		if name == excluded || (!useCaseSensitive && strings.EqualFold(name, excluded)) {
			return true
		}
	}
	return false
}

func parseCollectedGitignorePatterns(collected []gitignore.Pattern, caseInsensitive bool) []IgnorePattern {
	patterns := make([]IgnorePattern, 0, len(collected))
	parse := func(raw string) IgnorePattern {
		pattern := ParseIgnorePattern(raw)
		pattern.CaseInsensitive = caseInsensitive
		return pattern
	}
	for _, source := range collected {
		body := source.Glob
		if source.Negated {
			body = strings.TrimPrefix(body, "!")
		}
		nodeGlob := normalizePattern(source.NodeGlob)
		if body == "" || nodeGlob == "" {
			continue
		}

		// Every Git rule can match a directory node, including a rule without a
		// trailing slash (for example "build"). Keep a subtree projection for
		// sound walk pruning. The original node matcher remains a prefix of that
		// projection and GitNodeGlobEnd records its boundary.
		projection := body
		if strings.HasSuffix(projection, "/**") && !strings.HasSuffix(projection, "/**/*") {
			projection += "/*"
		} else if !strings.HasSuffix(projection, "/**/*") {
			projection += "/**/*"
		}
		if source.Negated {
			projection = "!" + projection
		}
		pattern := parse(projection)
		pattern.GitPattern = true
		pattern.GitDirectoryOnly = source.DirectoryOnly
		pattern.GitContentsOnly = source.ContentsOnly
		if caseInsensitive {
			// Git patterns are immutable after parsing. Fold them once here;
			// the per-file matcher then only needs to fold the target path once,
			// rather than allocating lower-case copies for every pattern/node.
			pattern.Glob = strings.ToLower(pattern.Glob)
			nodeGlob = strings.ToLower(nodeGlob)
		}
		if !strings.HasPrefix(pattern.Glob, nodeGlob) {
			// Collected patterns always satisfy this compact-representation
			// invariant. Reject an inconsistent manually constructed value
			// instead of slicing an unrelated projection at match time.
			continue
		}
		pattern.GitNodeGlobEnd = len(nodeGlob)
		patterns = append(patterns, pattern)
	}
	return patterns
}

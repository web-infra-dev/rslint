package config

import (
	"path"
	"strings"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/config/gitignore"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// mayIgnoreUnlistedFile asks whether any allowed spelling of an admitted
// file could instead be ignored. Only its leaf is unknown. Complete literal
// names can be checked by the same matcher used for ordinary inputs.
func (policy *GitPolicy) mayIgnoreUnlistedFile(path string, status entryNameStatus) bool {
	root, directory := policy.repository(path, false)
	if root == "" {
		return false
	}
	parent, _ := RelativePathWithinConfigRoot(directory, root, true)
	base := tspath.GetBaseFileName(path)
	for current := directory; ; current = gitignore.ParentDirectory(current) {
		id := gitignore.DirectoryPathID(current)
		for i := range policy.sources[id].patterns {
			p := &policy.sources[id].patterns[i]
			if p.Negated || p.GitDirectoryOnly {
				continue
			}
			leaf, possible := gitLeafInDirectory(p, parent)
			if !possible {
				continue
			}
			prefix := leaf
			wildcard := strings.IndexAny(leaf, "*?")
			if wildcard >= 0 {
				prefix = leaf[:wildcard]
			}
			suffix := gitLiteralSuffix(gitIgnoreNodeGlob(p))
			if status.distinctASCIIAnchors(base, prefix, p, false) || status.distinctASCIIAnchors(base, suffix, p, true) {
				continue
			}
			candidate := path
			match := unlistedGitMatch{positive: p, suffix: suffix}
			if leaf != "" && wildcard < 0 {
				candidate = tspath.CombinePaths(directory, leaf)
				match.literal = true
			}
			if policy.matchFile(candidate, &match) {
				return true
			}
		}
		if id == gitignore.DirectoryPathID(root) {
			return false
		}
	}
}

// gitLeafInDirectory returns the filename pattern under the known parent.
// An empty result with possible set means the rule cannot be split that far.
func gitLeafInDirectory(p *IgnorePattern, parent string) (leaf string, possible bool) {
	glob := gitIgnoreNodeGlob(p)
	directory, leaf := path.Split(glob)
	if p.CaseInsensitive {
		parent = gitignore.FoldCase(parent)
	}
	if strings.ContainsAny(leaf, "[]{}") {
		// A class can consume a slash, including the one Split found. If the
		// known parent lies deeper, this leaf's prefix could name a directory.
		directory, leaf = path.Split(glob[:strings.IndexAny(glob, "[]{}")])
		if utils.MatchGlob(directory+"**/*", parent) {
			return "", true
		}
		leaf += "*"
	}
	if leaf == "**" {
		return "", utils.MatchGlob(directory+"**", parent)
	}
	return leaf, utils.MatchGlob(strings.TrimSuffix(directory, "/"), parent)
}

// A wildcard admits more names than we can enumerate. Its original positive
// rule remains possible unless a later negation covers the entire family.
type unlistedGitMatch struct {
	positive *IgnorePattern
	literal  bool
	suffix   string
}

func (m *unlistedGitMatch) matches(p *IgnorePattern, path string) bool {
	if p == m.positive {
		return true
	}
	if !p.Negated {
		return false
	}
	// A case-sensitive negation cannot settle all variants of a literal
	// from a case-insensitive source, e.g. on a nested mount without .git.
	if m.literal && (!m.positive.CaseInsensitive || p.CaseInsensitive) {
		return gitIgnorePatternMatchesNode(p, path)
	}
	if p.CaseInsensitive != m.positive.CaseInsensitive {
		return false
	}
	glob := gitIgnoreNodeGlob(p)
	if glob == gitIgnoreNodeGlob(m.positive) && p.GitContentsOnly == m.positive.GitContentsOnly {
		return true
	}
	// A contents-only rule depends on the known parent. A final * or
	// *.suffix covers all names in this family when its parent matches.
	// Evaluate these complete paths with the normal Git node matcher.
	if p.GitContentsOnly {
		return gitIgnorePatternMatchesNode(p, path)
	}
	suffix := gitLiteralSuffix(glob)
	if !strings.HasSuffix(glob, "/*"+suffix) || !strings.HasSuffix(m.suffix, suffix) {
		return false
	}
	parent := tspath.GetDirectoryPath(path)
	return gitIgnorePatternMatchesNode(p, tspath.CombinePaths(parent, "x"+suffix))
}

func gitLiteralSuffix(glob string) string {
	// Both ends of a class are metacharacters. A slash inside [a/] cannot
	// turn part of that class into a literal part of a filename.
	return glob[strings.LastIndexAny(glob, "*?[]{}/")+1:]
}

// Compare only contiguous ASCII bytes at the same end. A Unicode spelling
// can expand or compose, so the first non-ASCII byte ends this proof.
func (status entryNameStatus) distinctASCIIAnchors(left, right string, p *IgnorePattern, suffix bool) bool {
	if status != entryNameUnlistedExactASCII && status != entryNameUnlistedFoldedASCII {
		return false
	}
	if status == entryNameUnlistedFoldedASCII || p.CaseInsensitive {
		left, right = gitignore.FoldCase(left), gitignore.FoldCase(right)
	}
	for i := range min(len(left), len(right)) {
		a, b := left[i], right[i]
		if suffix {
			a, b = left[len(left)-1-i], right[len(right)-1-i]
		}
		if a >= utf8.RuneSelf || b >= utf8.RuneSelf {
			return false
		}
		if a != b {
			return true
		}
	}
	return false
}

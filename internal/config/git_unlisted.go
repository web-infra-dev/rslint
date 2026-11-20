package config

import (
	"strings"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/config/gitignore"
)

// mayIgnoreUnlistedFile asks whether any allowed spelling of an admitted
// file could instead be ignored. Only its leaf is unknown. Complete literal
// names can be checked by the same matcher used for ordinary inputs.
func (policy *GitPolicy) mayIgnoreUnlistedFile(path string, asciiAnchors bool) bool {
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
			glob := gitIgnoreNodeGlob(p)
			if name, ok := gitIgnoreRootBasenameGlob(p); ok {
				glob = name
			}
			if !strings.ContainsAny(glob, "*?[{") {
				name := tspath.GetBaseFileName(glob)
				if asciiAnchors && (distinctASCIIAnchors(base, name, false) || distinctASCIIAnchors(base, name, true)) {
					continue
				}
				candidate := tspath.CombinePaths(directory, name)
				relative := tspath.CombinePaths(parent, name)
				if p.CaseInsensitive {
					relative = strings.ToLower(relative)
				}
				if gitIgnorePatternMatchesNode(p, relative) && policy.matchFile(candidate, &unlistedGitMatch{positive: p, literal: true}) {
					return true
				}
				continue
			}

			prefix, dir := literalSegmentPrefix(glob), parent
			if p.CaseInsensitive {
				dir = strings.ToLower(dir)
			}
			if prefix != "" && !segPrefix(prefix, dir) {
				continue
			}
			suffix := gitLiteralSuffix(glob)
			if asciiAnchors && distinctASCIIAnchors(base, suffix, true) {
				continue
			}
			if policy.matchFile(path, &unlistedGitMatch{positive: p, suffix: suffix}) {
				return true
			}
		}
		if id == gitignore.DirectoryPathID(root) {
			return false
		}
	}
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
func distinctASCIIAnchors(left, right string, suffix bool) bool {
	for i := range min(len(left), len(right)) {
		a, b := left[i], right[i]
		if suffix {
			a, b = left[len(left)-1-i], right[len(right)-1-i]
		}
		if a >= utf8.RuneSelf || b >= utf8.RuneSelf {
			return false
		}
		if a >= 'A' && a <= 'Z' {
			a += 'a' - 'A'
		}
		if b >= 'A' && b <= 'Z' {
			b += 'a' - 'A'
		}
		if a != b {
			return true
		}
	}
	return false
}

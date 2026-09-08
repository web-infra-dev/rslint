package gitignore

import "strings"

// Matcher applies an explicit list of Git ignore patterns without discovering
// files or applying lint configuration policy. It shares the collector's parser
// and path-node matching, including ordered negation and excluded parents.
type Matcher struct {
	patterns         []Pattern
	useCaseSensitive bool
}

func NewMatcher(patterns []string, useCaseSensitive bool) *Matcher {
	return &Matcher{
		patterns:         convertGitignoreToPatterns(strings.Join(patterns, "\n"), ""),
		useCaseSensitive: useCaseSensitive,
	}
}

// Match accepts a slash-separated relative path. A trailing slash marks a
// directory; invalid paths outside the matching root do not match.
func (matcher *Matcher) Match(path string) bool {
	if matcher == nil || path == "" || strings.HasPrefix(path, "/") {
		return false
	}
	directory := strings.HasSuffix(path, "/")
	parts := strings.Split(strings.TrimSuffix(path, "/"), "/")
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return false
		}
	}
	end := 0
	for index, part := range parts {
		if index > 0 {
			end++
		}
		end += len(part)
		isDirectory := index < len(parts)-1 || directory
		ignored := false
		for _, pattern := range matcher.patterns {
			if pattern.DirectoryOnly && !isDirectory {
				continue
			}
			if gitignorePrunePatternMatchesDirectoryNode(pattern, path[:end], matcher.useCaseSensitive) {
				ignored = !pattern.Negated
			}
		}
		if ignored {
			return true
		}
	}
	return false
}

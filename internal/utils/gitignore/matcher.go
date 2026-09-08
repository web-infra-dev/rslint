package gitignore

import (
	"strings"

	"github.com/web-infra-dev/rslint/internal/utils/minimatch3"
)

// Matcher applies ordered Git ignore patterns and parent exclusions, using
// minimatch for wildcard syntax, JavaScript case folding, and UTF-16 matching.
// Backslash quoting and class negation retain their glob meaning; ignore 5's
// accidental regexp escapes and class rewriting are not reproduced.
// It is immutable after construction and safe to share across files.
type Matcher struct{ patterns []compiledPattern }

type compiledPattern struct {
	match         *minimatch3.Matcher
	negated       bool
	directoryOnly bool
}

// NewMatcher accepts a list of individual patterns. An embedded newline stays
// inside its pattern; it does not introduce another rule.
func NewMatcher(patterns []string, useCaseSensitive bool) *Matcher {
	matcher := &Matcher{}
	for _, line := range patterns {
		line, ok := normalizeIgnorePattern(line)
		if !ok {
			continue
		}
		pattern, ok := parsePattern(line, "", func(pattern string) string { return pattern })
		if !ok {
			continue
		}
		glob := pattern.NodeGlob
		if pattern.ContentsOnly {
			glob += "/*"
		}
		matcher.patterns = append(matcher.patterns, compiledPattern{
			match: minimatch3.New(glob, minimatch3.Options{
				Dot: true, NoBrace: true, NoExt: true, NoNegate: true,
				NoComment: true, NoCase: !useCaseSensitive, PreserveWhitespace: true,
			}),
			negated: pattern.Negated, directoryOnly: pattern.DirectoryOnly,
		})
	}
	return matcher
}

// Only unescaped trailing spaces are insignificant in a Git ignore pattern.
// Array entries are atomic; CRLF handling belongs exclusively to the text API.
func normalizeIgnorePattern(line string) (string, bool) {
	line = strings.TrimPrefix(line, "\uFEFF")
	line = trimUnescapedTrailingWhitespace(line, " ")
	if line == "" || strings.HasPrefix(line, "#") ||
		strings.HasSuffix(line, `\`) && !strings.HasSuffix(line, `\\`) || strings.Contains(line, "//") {
		return "", false
	}
	return line, true
}

// NewMatcherFromText splits LF/CRLF-delimited ignore file contents into rules.
func NewMatcherFromText(content string, useCaseSensitive bool) *Matcher {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSuffix(line, "\r")
	}
	return NewMatcher(lines, useCaseSensitive)
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
			// A positive rule cannot change an already ignored node, nor can
			// a negation change one that is still included.
			if pattern.negated != ignored {
				continue
			}
			if pattern.directoryOnly && !isDirectory {
				continue
			}
			if pattern.match.Match(path[:end]) {
				ignored = !pattern.negated
			}
		}
		if ignored {
			return true
		}
	}
	return false
}

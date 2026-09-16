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
	matcher.appendPatterns(patterns, "", useCaseSensitive)
	return matcher
}

func (matcher *Matcher) appendPatterns(patterns []string, baseDir string, useCaseSensitive bool) {
	for _, line := range patterns {
		line, ok := normalizeIgnorePattern(line)
		if !ok {
			continue
		}
		pattern, ok := parsePattern(line, baseDir, func(pattern string) string { return pattern })
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
	return NewMatcherFromTextSources([]TextSource{{Text: content}}, useCaseSensitive)
}

// TextSource is ignore-file text scoped to a directory relative to the matcher root.
type TextSource struct{ BaseDir, Text string }

// NewMatcherFromTextSources combines ignore files in precedence order while
// preserving parent directory exclusions across their individual scopes.
func NewMatcherFromTextSources(sources []TextSource, useCaseSensitive bool) *Matcher {
	matcher := &Matcher{}
	for _, source := range sources {
		lines := strings.Split(source.Text, "\n")
		for i, line := range lines {
			lines[i] = strings.TrimSuffix(line, "\r")
		}
		matcher.appendPatterns(lines, source.BaseDir, useCaseSensitive)
	}
	return matcher
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

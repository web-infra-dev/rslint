// Package gitignore parses Git ignore patterns without filesystem discovery or
// lint configuration policy. Matcher applies explicit patterns with JavaScript
// glob string semantics; ParseText supplies the existing config glob projection.
package gitignore

import (
	"iter"
	"strings"
)

// Pattern retains the named path node separately from its descendants, so an
// excluded directory cannot be reopened by a negation targeting only its child.
// Glob is the file-level projection, including its leading negation marker.
// NodeGlob names the node before directory-only expansion. ContentsOnly records
// a genuine trailing /**, which must not match the directory before it.
type Pattern struct {
	Glob          string
	NodeGlob      string
	Negated       bool
	DirectoryOnly bool
	ContentsOnly  bool
}

// ParseText projects Git ignore file text to globs relative to baseDir.
// The projection preserves Git character classes and is suitable for rslint's
// configuration glob pipeline; Matcher separately evaluates explicit patterns.
func ParseText(content string, baseDir string) iter.Seq[Pattern] {
	return func(yield func(Pattern) bool) {
		for line := range strings.SplitSeq(content, "\n") {
			line, ok := normalizeGitignoreLine(line)
			if !ok {
				continue
			}
			pattern, ok := parsePattern(line, baseDir, gitignorePatternForConfigGlob)
			if ok && !yield(pattern) {
				return
			}
		}
	}
}

func normalizeGitignoreLine(line string) (string, bool) {
	// A CR from CRLF is a line terminator, not pattern text. Git otherwise
	// preserves leading whitespace and removes only unescaped trailing
	// whitespace. In particular, ` leading` is a real filename pattern and
	// `trailing\ ` retains its final space.
	line = strings.TrimSuffix(line, "\r")
	line = trimUnescapedTrailingWhitespace(line, " \t")
	if line == "" || strings.HasPrefix(line, "#") {
		return "", false
	}
	return line, true
}

func trimUnescapedTrailingWhitespace(line, whitespace string) string {
	for len(line) > 0 {
		last := line[len(line)-1]
		if !strings.ContainsRune(whitespace, rune(last)) {
			break
		}
		backslashes := 0
		for index := len(line) - 2; index >= 0 && line[index] == '\\'; index-- {
			backslashes++
		}
		if backslashes%2 == 1 {
			break
		}
		line = line[:len(line)-1]
	}
	return line
}

func parsePattern(line string, baseDir string, quote func(string) string) (Pattern, bool) {
	negated := false
	if strings.HasPrefix(line, "!") {
		negated = true
		line = line[1:]
	}
	if containsParentPathComponent(line) {
		return Pattern{}, false
	}

	// Trailing / means directory-only; for glob purposes, we match dir/**
	dirOnly := false
	if strings.HasSuffix(line, "/") {
		dirOnly = true
		line = strings.TrimSuffix(line, "/")
	}

	// Determine if rooted
	rooted := false
	if strings.HasPrefix(line, "/") {
		rooted = true
		line = strings.TrimPrefix(line, "/")
	} else if strings.Contains(line, "/") {
		// Contains / in middle → implicitly rooted relative to .gitignore dir
		rooted = true
	}

	if line == "" {
		return Pattern{}, false
	}
	contentsOnly := strings.HasSuffix(line, "/**")
	line = quote(line)
	baseGlob := gitignoreBaseDirForConfigGlob(baseDir)

	// Build the glob pattern
	var glob string
	if rooted {
		if baseGlob != "" {
			glob = baseGlob + "/" + line
		} else {
			glob = line
		}
	} else {
		// Unrooted: matches at any depth
		if baseGlob != "" {
			glob = baseGlob + "/**/" + line
		} else {
			glob = "**/" + line
		}
	}

	nodeGlob := glob

	// Append /**/* for directory patterns to match all contents.
	// Use /**/* (file-level) instead of /** (directory-level) because collected
	// rules participate in one ordered global-ignore sequence and later
	// negations must be able to re-include descendants.
	if dirOnly && !strings.HasSuffix(glob, "/**/*") {
		if contentsOnly {
			// A genuine trailing /** already supplies the recursive part; add
			// one required component so the directory before it is not matched.
			glob += "/*"
		} else {
			glob += "/**/*"
		}
	}

	if negated {
		glob = "!" + glob
	} else if strings.HasPrefix(glob, "!") {
		// A leading `\!` is a literal exclamation mark. Keep the raw config glob
		// from looking like an rslint negation; NormalizePath later removes `./`
		// without changing the matcher.
		glob = "./" + glob
	}
	return Pattern{
		Glob:          glob,
		NodeGlob:      nodeGlob,
		Negated:       negated,
		DirectoryOnly: dirOnly,
		ContentsOnly:  contentsOnly,
	}, true
}

// gitignoreBaseDirForConfigGlob quotes the real directory names contributed by
// the filesystem. Unlike the .gitignore rule itself, this prefix is never glob
// syntax: a repository directory literally named pkg[1] or pkg{a} must remain
// literal on Unix, Windows, and case-insensitive macOS volumes.
func gitignoreBaseDirForConfigGlob(baseDir string) string {
	if !strings.ContainsAny(baseDir, "*?[{}") {
		return baseDir
	}
	var result strings.Builder
	result.Grow(len(baseDir))
	for index := range len(baseDir) {
		character := baseDir[index]
		if character == '/' {
			result.WriteByte(character)
		} else {
			appendConfigGlobLiteral(&result, character)
		}
	}
	return result.String()
}

// gitignorePatternForConfigGlob removes Git's backslash quoting before a
// pattern enters rslint's generic glob pipeline. That pipeline normalizes
// backslashes as path separators, so escaped glob metacharacters must instead
// be represented as one-character classes. Doublestar supports brace
// alternation while Git does not, therefore unescaped braces are quoted too.
func gitignorePatternForConfigGlob(pattern string) string {
	var result strings.Builder
	result.Grow(len(pattern))
	inClass := false
	for index := 0; index < len(pattern); index++ {
		character := pattern[index]
		if character == '\\' && index+1 < len(pattern) {
			index++
			appendConfigGlobLiteral(&result, pattern[index])
			continue
		}
		if character == '[' {
			inClass = true
			result.WriteByte(character)
			continue
		}
		if character == ']' && inClass {
			inClass = false
			result.WriteByte(character)
			continue
		}
		if !inClass && (character == '{' || character == '}') {
			appendConfigGlobLiteral(&result, character)
			continue
		}
		result.WriteByte(character)
	}
	return result.String()
}

func appendConfigGlobLiteral(result *strings.Builder, character byte) {
	switch character {
	case '*':
		result.WriteString("[*]")
	case '?':
		result.WriteString("[?]")
	case '[':
		result.WriteString("[[]")
	case '{':
		result.WriteString("[{]")
	case '}':
		result.WriteString("[}]")
	default:
		result.WriteByte(character)
	}
}

func containsParentPathComponent(pattern string) bool {
	// Git patterns always use `/` as the separator; backslash quotes the next
	// byte even on Windows and must not manufacture a parent component.
	for _, component := range strings.Split(pattern, "/") {
		if component == ".." {
			return true
		}
	}
	return false
}

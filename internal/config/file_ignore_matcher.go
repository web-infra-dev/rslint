package config

import (
	"strings"
	"unicode/utf8"
)

// fileIgnoreMatcher is an immutable entry-local ignore list, prepared once by
// the config-target resolver. It consumes the resolver's existing path space;
// global ignores, directory blocking and Git parent semantics stay with their
// existing matchers. Patterns and indexes share the resolver's lifetime.
type fileIgnoreMatcher struct {
	patterns []IgnorePattern
	steps    []fileIgnoreStep
}

// A step retains either an ordered pattern segment or a set of consecutive
// positive literals. Negations and general globs keep their original positions.
type fileIgnoreStep struct {
	patterns []IgnorePattern
	literals map[string]struct{}
}

func newFileIgnoreMatcher(patterns []IgnorePattern) fileIgnoreMatcher {
	matcher := fileIgnoreMatcher{patterns: patterns}
	// Keep short lists on the allocation-free preparation path.
	const minimumLiteralRun = 8
	if len(patterns) < minimumLiteralRun {
		return matcher
	}
	for _, pattern := range patterns {
		if pattern.GitPattern || pattern.CaseInsensitive || patternHasMatchDirectory(pattern) {
			return matcher
		}
	}
	pending := 0
	for index := 0; index < len(patterns); {
		if !isPositiveFileIgnoreLiteral(patterns[index]) {
			index++
			continue
		}
		end := index + 1
		for end < len(patterns) && isPositiveFileIgnoreLiteral(patterns[end]) {
			end++
		}
		if end-index >= minimumLiteralRun {
			if pending < index {
				matcher.steps = append(matcher.steps, fileIgnoreStep{patterns: patterns[pending:index]})
			}
			literals := make(map[string]struct{}, end-index)
			for _, pattern := range patterns[index:end] {
				literals[pattern.Glob] = struct{}{}
			}
			matcher.steps = append(matcher.steps, fileIgnoreStep{literals: literals})
			pending = end
		}
		index = end
	}
	if len(matcher.steps) != 0 && pending < len(patterns) {
		matcher.steps = append(matcher.steps, fileIgnoreStep{patterns: patterns[pending:]})
	}
	return matcher
}

func isPositiveFileIgnoreLiteral(pattern IgnorePattern) bool {
	if pattern.Negated || strings.ContainsAny(pattern.Glob, "*?[]{}\\") {
		return false
	}
	// The glob engine compares decoded runes. Invalid UTF-8 and the replacement
	// rune must retain that behavior instead of switching to byte equality.
	return utf8.ValidString(pattern.Glob) && !strings.ContainsRune(pattern.Glob, utf8.RuneError)
}

func (matcher fileIgnoreMatcher) isIgnored(path *fileMatchPath) bool {
	if len(matcher.steps) == 0 {
		return path.isIgnored(matcher.patterns)
	}
	normalizedPath, unixPath := path.normalizedPaths()
	ignored := false
	for _, step := range matcher.steps {
		if step.literals == nil {
			ignored = applyFileIgnorePatternsNormalized(normalizedPath, unixPath, step.patterns, ignored)
		} else if !ignored {
			_, ignored = step.literals[normalizedPath]
			if !ignored && unixPath != normalizedPath {
				_, ignored = step.literals[unixPath]
			}
		}
	}
	return ignored
}

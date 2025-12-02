package config

import (
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/microsoft/typescript-go/shim/tspath"
)

// negPatternPrefix is the prefix character used to indicate a negation pattern.
const negPatternPrefix = "!"

// normalizeAbsolutePath converts file path to be a absolute path to cwd
func normalizeAbsolutePath(filePath, cwd string) string {
	return tspath.GetNormalizedAbsolutePath(filePath, cwd)
}

// parseNegationPattern parses a glob pattern and distinguishes whether it's a negation pattern.
// Returns (isNeg, actualPattern) where isNeg indicates if the pattern starts with "!".
func parseNegationPattern(pattern string) (isNeg bool, actualPattern string) {
	if strings.HasPrefix(pattern, negPatternPrefix) {
		return true, pattern[1:]
	}
	return false, pattern
}

// classifyPatterns splits a list of glob patterns into its positive and negative patterns.
func classifyPatterns(patterns []string) (positivePatterns []string, negativePatterns []string) {
	for _, pattern := range patterns {
		isNeg, actualPattern := parseNegationPattern(pattern)
		if isNeg {
			negativePatterns = append(negativePatterns, actualPattern)
		} else {
			positivePatterns = append(positivePatterns, actualPattern)
		}
	}
	return positivePatterns, negativePatterns
}

// normalizePatterns normalizes a list of glob patterns to be absolute paths relative to cwd.
func normalizePatterns(patterns []string, cwd string) []string {
	normalizedPatterns := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		normalizedPatterns = append(normalizedPatterns, normalizeAbsolutePath(pattern, cwd))
	}
	return normalizedPatterns
}

// normalizedFilePatterns is a struct that contains the positive glob patterns and negative glob patterns normalized to be absolute paths relative to cwd.
type normalizedFilePatterns struct {
	positivePatterns []string
	negativePatterns []string
}

func newNormalizedFilePatterns(patterns []string, cwd string) *normalizedFilePatterns {
	positivePatterns, negativePatterns := classifyPatterns(patterns)
	return &normalizedFilePatterns{
		positivePatterns: normalizePatterns(positivePatterns, cwd),
		negativePatterns: normalizePatterns(negativePatterns, cwd),
	}
}

// isFileMatched checks if a file is matched by the normalized file patterns.
// A file is matched if:
//   - It matches at least one positive pattern (or there are no positive patterns), and
//   - It is not excluded by any negative pattern.
func (n *normalizedFilePatterns) isFileMatched(absoluteFilePath string) bool {
	isMatched := len(n.positivePatterns) == 0
	for _, pattern := range n.positivePatterns {
		if matched, err := doublestar.Match(pattern, absoluteFilePath); err == nil && matched {
			isMatched = true
			break
		}
	}
	if !isMatched {
		return false
	}
	for _, pattern := range n.negativePatterns {
		if matched, err := doublestar.Match(pattern, absoluteFilePath); err == nil && matched {
			isMatched = false
			break
		}
	}
	return isMatched
}

type normalizedIgnorePattern struct {
	isNeg   bool
	pattern string
}

// normalizedIgnorePatterns is a collection of normalized ignore patterns.
type normalizedIgnorePatterns []*normalizedIgnorePattern

func newNormalizedIgnorePatterns(patterns []string, cwd string) normalizedIgnorePatterns {
	parsedPatterns := make([]*normalizedIgnorePattern, 0, len(patterns))
	for _, pattern := range patterns {
		isNeg, actualPattern := parseNegationPattern(pattern)
		parsedPatterns = append(parsedPatterns, &normalizedIgnorePattern{
			isNeg:   isNeg,
			pattern: normalizeAbsolutePath(actualPattern, cwd),
		})
	}
	return parsedPatterns
}

// isFileIgnored checks if a file should be ignored based on the normalized ignore patterns.
// It follows ESLint's "last match wins" rule: patterns are processed in order,
// and the last matching pattern determines whether the file is ignored.
func (n normalizedIgnorePatterns) isFileIgnored(absoluteFilePath string) bool {
	isIgnored := false
	for _, pattern := range n {
		if matched, err := doublestar.Match(pattern.pattern, absoluteFilePath); err == nil {
			if matched {
				if pattern.isNeg {
					isIgnored = false
				} else {
					isIgnored = true
				}
			}
		}
	}
	return isIgnored
}

type fileMatcher struct {
	normalizedFilePatterns   *normalizedFilePatterns
	normalizedIgnorePatterns normalizedIgnorePatterns
}

func newFileMatcher(config *ConfigEntry, cwd string) *fileMatcher {
	return &fileMatcher{
		normalizedFilePatterns:   newNormalizedFilePatterns(config.Files, cwd),
		normalizedIgnorePatterns: newNormalizedIgnorePatterns(config.Ignores, cwd),
	}
}

func (f *fileMatcher) isFileMatched(absoluteFilePath string) bool {
	if f.normalizedIgnorePatterns.isFileIgnored(absoluteFilePath) {
		return false
	}
	return f.normalizedFilePatterns.isFileMatched(absoluteFilePath)
}

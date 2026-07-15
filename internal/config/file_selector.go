package config

import (
	"strings"

	"github.com/web-infra-dev/rslint/internal/config/minimatch"
)

type compiledFileSelector struct {
	matcher *minimatch.SearchPattern
}

func compileFileSelector(pattern string) compiledFileSelector {
	compiled, err := minimatch.CompileRelativePattern(normalizeConfigPattern(pattern))
	if err != nil {
		return compiledFileSelector{}
	}
	return compiledFileSelector{matcher: &compiled}
}

// ConfigArray removes exactly one leading "./" from ordinary selectors, or
// the "./" after exactly one leading negation. Additional leading negations
// are left for Minimatch, so !!./ and !!!./ deliberately retain their dot
// segment and have different match results.
func normalizeConfigPattern(pattern string) string {
	if strings.HasPrefix(pattern, "./") {
		return strings.TrimPrefix(pattern, "./")
	}
	if strings.HasPrefix(pattern, "!./") {
		return "!" + strings.TrimPrefix(pattern, "!./")
	}
	return pattern
}

func (selector compiledFileSelector) matches(paths []string) bool {
	if selector.matcher == nil {
		return false
	}
	for _, path := range paths {
		if selector.matcher.Match(path) {
			return true
		}
	}
	return false
}

type compiledFileSelectorEntry struct {
	basePath          string
	caseSensitive     bool
	hasSelectors      bool
	selectors         []compiledFileSelector
	groups            [][]compiledFileSelector
	specificSelectors []compiledFileSelector
	specificGroups    [][]compiledFileSelector
	ignores           []IgnorePattern
}

func (entry compiledFileSelectorEntry) matches(filePath string) bool {
	paths, within := selectorMatchPaths(filePath, entry.basePath, true, entry.caseSensitive)
	if !within {
		return false
	}
	for _, selector := range entry.selectors {
		if selector.matches(paths) {
			return true
		}
	}
	for _, group := range entry.groups {
		matched := true
		for _, selector := range group {
			if !selector.matches(paths) {
				matched = false
				break
			}
		}
		if matched {
			return true
		}
	}
	return false
}

func (entry compiledFileSelectorEntry) contains(filePath string) bool {
	_, within := selectorMatchPaths(filePath, entry.basePath, true, entry.caseSensitive)
	return within
}

func (entry compiledFileSelectorEntry) ignoresFile(filePath string) bool {
	paths, within := selectorMatchPaths(filePath, entry.basePath, true, entry.caseSensitive)
	if !within || len(paths) == 0 {
		return false
	}
	return isFileIgnored(paths[0], entry.ignores, "")
}

func (entry compiledFileSelectorEntry) specificallyMatches(filePath string) bool {
	paths, within := selectorMatchPaths(filePath, entry.basePath, true, entry.caseSensitive)
	if !within {
		return false
	}
	for _, selector := range entry.specificSelectors {
		if selector.matches(paths) {
			return true
		}
	}
	for _, group := range entry.specificGroups {
		matched := true
		for _, selector := range group {
			if !selector.matches(paths) {
				matched = false
				break
			}
		}
		if matched {
			return true
		}
	}
	return false
}

func isUniversalConfigPattern(pattern string) bool {
	return pattern == "*" || strings.HasPrefix(pattern, "!") ||
		strings.HasSuffix(pattern, "/*") || strings.HasSuffix(pattern, "/**")
}

// FileSelectorMatcher is the immutable, concurrency-safe selector union for
// one effective flat config. Discovery owners build it once and share it across
// every concurrent directory walk, so authored files globs and entry-local
// ignores are never reparsed per candidate file.
type FileSelectorMatcher struct {
	cwd           string
	caseSensitive bool
	entries       []compiledFileSelectorEntry
}

func NewFileSelectorMatcher(config RslintConfig, cwd string) *FileSelectorMatcher {
	cwd = normalizePathForRoot(cwd, cwd)
	caseSensitive := selectorScopeCaseSensitive(cwd)
	matcher := &FileSelectorMatcher{
		cwd:           cwd,
		caseSensitive: caseSensitive,
		entries:       make([]compiledFileSelectorEntry, len(config)),
	}
	for entryIndex, entry := range config {
		compiled := compiledFileSelectorEntry{
			basePath:          resolveConfigEntryBasePath(cwd, entry.BasePath),
			caseSensitive:     selectorScopeCaseSensitive(resolveConfigEntryBasePath(cwd, entry.BasePath)),
			hasSelectors:      hasFileSelectors(entry),
			selectors:         make([]compiledFileSelector, 0, len(entry.Files)),
			groups:            make([][]compiledFileSelector, 0, len(entry.FilePatternGroups)),
			specificSelectors: make([]compiledFileSelector, 0, len(entry.Files)),
			specificGroups:    make([][]compiledFileSelector, 0, len(entry.FilePatternGroups)),
			ignores:           ParseIgnorePatterns(entry.Ignores),
		}
		for _, pattern := range entry.Files {
			selector := compileFileSelector(pattern)
			compiled.selectors = append(compiled.selectors, selector)
			if !isUniversalConfigPattern(pattern) {
				compiled.specificSelectors = append(compiled.specificSelectors, selector)
			}
		}
		for _, group := range entry.FilePatternGroups {
			compiledGroup := make([]compiledFileSelector, 0, len(group))
			universal := true
			for _, pattern := range group {
				compiledGroup = append(compiledGroup, compileFileSelector(pattern))
				universal = universal && isUniversalConfigPattern(pattern)
			}
			compiled.groups = append(compiled.groups, compiledGroup)
			if !universal {
				compiled.specificGroups = append(compiled.specificGroups, compiledGroup)
			}
		}
		matcher.entries[entryIndex] = compiled
	}
	return matcher
}

// ConfigArray uses the host path implementation for basePath containment:
// POSIX paths compare lexically, while Windows drive and UNC paths compare
// casing-insensitively. Minimatch selectors themselves remain case-sensitive.
func selectorScopeCaseSensitive(basePath string) bool {
	basePath = NormalizeHostPath(basePath)
	if len(basePath) >= 2 && basePath[1] == ':' {
		return false
	}
	return !strings.HasPrefix(basePath, "//")
}

// Selects reports whether the implicit extension baseline or at least one
// explicit files entry selects filePath. Global ignores are deliberately not
// evaluated here; discovery's ordered ignore matcher owns that independent
// admission decision.
func (matcher *FileSelectorMatcher) Selects(filePath string) bool {
	if matcher == nil {
		return false
	}
	filePath = matcher.resolveFilePath(filePath)
	if _, within := selectorMatchPaths(filePath, matcher.cwd, false, matcher.caseSensitive); !within {
		return false
	}
	if isDefaultLintFile(filePath) {
		return true
	}
	for _, entry := range matcher.entries {
		if entry.hasSelectors && entry.specificallyMatches(filePath) && !entry.ignoresFile(filePath) {
			return true
		}
	}
	return false
}

func (matcher *FileSelectorMatcher) entryMatches(entryIndex int, filePath string) bool {
	if matcher == nil || entryIndex < 0 || entryIndex >= len(matcher.entries) {
		return false
	}
	return matcher.entries[entryIndex].matches(matcher.resolveFilePath(filePath))
}

func (matcher *FileSelectorMatcher) entryContains(entryIndex int, filePath string) bool {
	if matcher == nil || entryIndex < 0 || entryIndex >= len(matcher.entries) {
		return false
	}
	return matcher.entries[entryIndex].contains(matcher.resolveFilePath(filePath))
}

func (matcher *FileSelectorMatcher) entryIgnoresFile(entryIndex int, filePath string) bool {
	if matcher == nil || entryIndex < 0 || entryIndex >= len(matcher.entries) {
		return false
	}
	return matcher.entries[entryIndex].ignoresFile(matcher.resolveFilePath(filePath))
}

func (matcher *FileSelectorMatcher) resolveFilePath(filePath string) string {
	filePath = normalizePathForRoot(matcher.cwd, filePath)
	if !pathIsAbsoluteForRoot(matcher.cwd, filePath) && matcher.cwd != "" && matcher.cwd != "." {
		return resolvePathForRoot(matcher.cwd, matcher.cwd, filePath)
	}
	return filePath
}

func selectorMatchPaths(filePath string, basePath string, allowBaseRoot bool, useCaseSensitive bool) ([]string, bool) {
	governingRoot := basePath
	filePath = normalizePathForRoot(governingRoot, filePath)
	basePath = normalizePathForRoot(governingRoot, basePath)
	var normalized string
	if basePath != "" && basePath != "." {
		var within bool
		normalized, within = RelativePathWithinConfigRoot(filePath, basePath, useCaseSensitive)
		if !within {
			return nil, false
		}
	} else if !pathIsAbsoluteForRoot(governingRoot, filePath) {
		normalized = filePath
		if normalized == ".." || strings.HasPrefix(normalized, "../") {
			return nil, false
		}
	} else {
		normalized = filePath
	}
	if normalized == "" || normalized == "." {
		if !allowBaseRoot {
			return nil, false
		}
		return []string{""}, true
	}
	normalized = strings.TrimPrefix(normalizePathForRoot(governingRoot, normalized), "./")
	paths := []string{normalized}
	if !pathUsesNativePOSIXSemantics(basePath) {
		unixPath := strings.ReplaceAll(normalized, "\\", "/")
		if unixPath != normalized {
			paths = append(paths, unixPath)
		}
	}
	return paths, true
}

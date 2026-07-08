package config

import (
	"sort"
	"strings"

	"github.com/microsoft/typescript-go/shim/vfs"
)

// ReadGitignoreAsGlobs reads .gitignore files relevant to configDir and
// converts their patterns to rslint glob format suitable for use as global
// ignore patterns.
//
// It collects patterns from two sources:
//  1. Ancestor .gitignore files: walks UP from configDir to filesystem root,
//     collecting .gitignore patterns at each level. This implements gitignore
//     inheritance (root .gitignore affects all subdirectories).
//  2. Descendant .gitignore files: walks DOWN from configDir, collecting
//     nested .gitignore with directory-scoped prefixes.
//
// configIgnores are the user-configured global ignore patterns (from config
// entries with only ignores), already parsed. Directories that are
// directory-level blocked by these patterns are skipped during the descendant
// scan — their .gitignore files are not collected because files in those
// directories will never be linted (isDirAbsolutelyBlocked guarantees this).
//
// Returns nil if no .gitignore files are found.
func ReadGitignoreAsGlobs(configDir string, fsys vfs.FS, configIgnores []IgnorePattern) []string {
	if fsys == nil {
		return nil
	}
	normalizedRoot := normalizeGlobPath(configDir)
	var allGlobs []string

	// Phase 1: collect ancestor .gitignore files (walk UP). Their patterns
	// are interpreted relative to the ancestor .gitignore's directory, then
	// projected into configDir-relative globs because rslint ignore matching
	// evaluates paths relative to each configDir.
	ancestorGlobs, configRootIgnored := collectAncestorGitignoreGlobs(normalizedRoot, fsys)
	allGlobs = append(allGlobs, ancestorGlobs...)
	if configRootIgnored {
		return allGlobs
	}

	// Phase 2: collect descendant .gitignore files (walk DOWN from configDir).
	collectGitignoreGlobs(normalizedRoot, "", fsys, &allGlobs, configIgnores)

	if len(allGlobs) == 0 {
		return nil
	}
	return allGlobs
}

// ReadGitignoreAsGlobsForFiles reads only .gitignore files on the ancestor
// chains of explicit target files. This is used by API-style calls where the
// target set is already known; unlike ReadGitignoreAsGlobs, it does not scan
// every descendant of configDir.
func ReadGitignoreAsGlobsForFiles(configDir string, fsys vfs.FS, files []string) []string {
	if fsys == nil || len(files) == 0 {
		return nil
	}

	normalizedConfigDir := normalizeGlobPath(configDir)
	dirSet := make(map[string]struct{})
	for _, file := range files {
		current := dirOfPath(normalizeGlobPath(file))
		for current != "" {
			dirSet[current] = struct{}{}
			next := parentDir(current)
			if next == current {
				break
			}
			current = next
		}
	}

	dirs := make([]string, 0, len(dirSet))
	for dir := range dirSet {
		dirs = append(dirs, dir)
	}
	sortByPathDepth(dirs)

	var allGlobs []string
	var pruneRules []gitignorePruneRule
	for _, dir := range dirs {
		if isDirIgnoredByPruneRules(dir, pruneRules) {
			continue
		}

		content, ok := fsys.ReadFile(dir + "/.gitignore")
		if !ok {
			continue
		}
		if rel, ok := relativeDir(normalizedConfigDir, dir); ok {
			allGlobs = append(allGlobs, convertGitignoreToGlobs(content, rel)...)
		} else {
			allGlobs = append(allGlobs, convertAncestorGitignoreToGlobs(content, dir, normalizedConfigDir)...)
		}
		if rule, ok := newGitignorePruneRule(dir, content); ok {
			pruneRules = append(pruneRules, rule)
		}
	}
	if len(allGlobs) == 0 {
		return nil
	}
	return allGlobs
}

// ExtractConfigIgnores collects global ignore patterns from config entries and
// parses them once into structured form. These are patterns from entries that
// have only ignores (no files/rules/etc.), representing user-configured
// directories to exclude from linting. Parsing here (rather than per file/dir)
// keeps the lint hot path and the directory walks off the string-classification
// cost.
func ExtractConfigIgnores(config RslintConfig) []IgnorePattern {
	var ignores []string
	for _, entry := range config {
		if isGlobalIgnoreEntry(entry) {
			ignores = append(ignores, entry.Ignores...)
		}
	}
	return ParseIgnorePatterns(ignores)
}

// collectAncestorGitignoreGlobs walks UP from configDir to filesystem root,
// collecting .gitignore patterns from each ancestor directory.
// Returns patterns in root-first order (outermost ancestor first).
func collectAncestorGitignoreGlobs(configDir string, fsys vfs.FS) ([]string, bool) {
	// Collect ancestor dirs (from dir's parent up to root)
	var ancestors []string
	current := parentDir(configDir)
	for current != "" && current != configDir {
		ancestors = append(ancestors, current)
		next := parentDir(current)
		if next == current {
			break
		}
		current = next
	}

	// Reverse to get root-first order
	for i, j := 0, len(ancestors)-1; i < j; i, j = i+1, j-1 {
		ancestors[i], ancestors[j] = ancestors[j], ancestors[i]
	}

	// Read .gitignore from each ancestor
	var globs []string
	var pruneRules []gitignorePruneRule
	configRootIgnored := false
	for _, ancestor := range ancestors {
		if isDirIgnoredByPruneRules(ancestor, pruneRules) {
			configRootIgnored = true
			break
		}
		gitignorePath := ancestor + "/.gitignore"
		if content, ok := fsys.ReadFile(gitignorePath); ok {
			converted := convertAncestorGitignoreToGlobs(content, ancestor, configDir)
			globs = append(globs, converted...)
			if rule, ok := newGitignorePruneRule(ancestor, content); ok {
				pruneRules = append(pruneRules, rule)
				configRootIgnored = isDirIgnoredByPruneRules(configDir, pruneRules)
			}
		}
	}
	return globs, configRootIgnored
}

type gitignorePruneRule struct {
	baseDir  string
	patterns []gitignorePrunePattern
}

type gitignorePrunePattern struct {
	negated bool
	rooted  bool
	pattern string
}

func newGitignorePruneRule(baseDir string, content string) (gitignorePruneRule, bool) {
	patterns := parseGitignorePrunePatterns(content)
	if len(patterns) == 0 {
		return gitignorePruneRule{}, false
	}
	return gitignorePruneRule{baseDir: baseDir, patterns: patterns}, true
}

func parseGitignorePrunePatterns(content string) []gitignorePrunePattern {
	var patterns []gitignorePrunePattern
	for _, line := range strings.Split(content, "\n") {
		line, ok := normalizeGitignoreLine(line)
		if !ok {
			continue
		}
		negated := false
		if strings.HasPrefix(line, "!") {
			negated = true
			line = line[1:]
		}
		line = strings.TrimSuffix(line, "/")
		rooted := false
		if strings.HasPrefix(line, "/") {
			rooted = true
			line = strings.TrimPrefix(line, "/")
		} else if strings.Contains(line, "/") {
			rooted = true
		}
		if line == "" {
			continue
		}
		patterns = append(patterns, gitignorePrunePattern{
			negated: negated,
			rooted:  rooted,
			pattern: line,
		})
	}
	return patterns
}

func isDirIgnoredByPruneRules(dir string, rules []gitignorePruneRule) bool {
	ignored := false
	for _, rule := range rules {
		rel, ok := relativeDir(rule.baseDir, dir)
		if !ok || rel == "" {
			continue
		}
		for _, pattern := range rule.patterns {
			if gitignorePrunePatternCoversDir(pattern, rel) {
				ignored = !pattern.negated
			}
		}
	}
	return ignored
}

func gitignorePrunePatternCoversDir(pattern gitignorePrunePattern, relDir string) bool {
	for current := relDir; current != ""; current = parentDir(current) {
		if gitignorePrunePatternMatchesPath(pattern, current) {
			return true
		}
	}
	return false
}

func gitignorePrunePatternMatchesPath(pattern gitignorePrunePattern, relPath string) bool {
	if pattern.rooted {
		return matchGlob(pattern.pattern, relPath)
	}
	for _, part := range splitPathComponents(relPath) {
		if matchGlob(pattern.pattern, part) {
			return true
		}
	}
	return false
}

// parentDir returns the parent directory of dir, or "" if at root.
func parentDir(dir string) string {
	idx := strings.LastIndex(dir, "/")
	if idx <= 0 {
		return ""
	}
	return dir[:idx]
}

func dirOfPath(filePath string) string {
	idx := strings.LastIndex(filePath, "/")
	if idx < 0 {
		return ""
	}
	if idx == 0 {
		return "/"
	}
	return filePath[:idx]
}

func relativeDir(root string, dir string) (string, bool) {
	if root == dir {
		return "", true
	}
	prefix := root
	if prefix != "/" {
		prefix += "/"
	}
	if strings.HasPrefix(dir, prefix) {
		return strings.TrimPrefix(dir, prefix), true
	}
	return "", false
}

func sortByPathDepth(paths []string) {
	sort.Slice(paths, func(i, j int) bool {
		depthI := strings.Count(paths[i], "/")
		depthJ := strings.Count(paths[j], "/")
		if depthI != depthJ {
			return depthI < depthJ
		}
		return paths[i] < paths[j]
	})
}

// collectGitignoreGlobs recursively scans for .gitignore files and converts
// their patterns to globs. Raw patterns from parent .gitignore files are used
// to prune directories during scanning (avoids entering e.g. target/ with
// thousands of subdirectories) without confusing "dir contents ignored" with
// "dir itself ignored".
//
// configIgnores provides additional directory-level pruning from the user's
// lint config. If isDirAbsolutelyBlocked returns true for a directory against
// these patterns, the directory is skipped. This is safe because
// isDirAbsolutelyBlocked is the same predicate used by the linter's
// GetConfigForFile (via isDirBlockedByIgnores) — if a directory is blocked
// here, its files will never be linted, so collecting its .gitignore is
// unnecessary.
func collectGitignoreGlobs(absDir string, relDir string, fsys vfs.FS, result *[]string, configIgnores []IgnorePattern) {
	collectGitignoreGlobsRecursive(absDir, relDir, fsys, result, configIgnores, nil)
}

func collectGitignoreGlobsRecursive(absDir string, relDir string, fsys vfs.FS, result *[]string, configIgnores []IgnorePattern, pruneRules []gitignorePruneRule) {
	gitignorePath := absDir + "/.gitignore"
	if content, ok := fsys.ReadFile(gitignorePath); ok {
		localGlobs := convertGitignoreToGlobs(content, relDir)
		*result = append(*result, localGlobs...)
		if rule, ok := newGitignorePruneRule(absDir, content); ok {
			pruneRules = append(pruneRules, rule)
		}
	}

	entries := fsys.GetAccessibleEntries(absDir)
	for _, dir := range entries.Directories {
		if _, excluded := defaultExcludeDirs[dir]; excluded {
			continue
		}

		childRel := dir
		if relDir != "" {
			childRel = relDir + "/" + dir
		}

		// Prune directories that are directory-level blocked by config ignores.
		// isDirAbsolutelyBlocked is the same predicate the linter uses in
		// GetConfigForFile → isDirBlockedByIgnores. If it returns true here,
		// the linter will also return nil for any file in this directory,
		// meaning files here are never linted. Therefore their .gitignore
		// patterns are irrelevant and we can safely skip collecting them.
		// Checked first because configIgnores is typically a short list (a few
		// user-defined patterns), whereas *result grows as we collect more
		// .gitignore patterns — checking configIgnores first avoids a linear
		// scan of the longer list for directories blocked by config.
		if len(configIgnores) > 0 && isDirAbsolutelyBlocked(childRel, configIgnores) {
			continue
		}

		childAbs := absDir + "/" + dir
		// Prune directories already ignored as directories by collected
		// gitignore patterns. Critical for performance: without it, scanning
		// rspack enters target/ (6,277 Rust build dirs, 0 .ts files).
		if isDirIgnoredByPruneRules(childAbs, pruneRules) {
			continue
		}

		collectGitignoreGlobsRecursive(childAbs, childRel, fsys, result, configIgnores, pruneRules)
	}
}

// convertGitignoreToGlobs converts .gitignore file content into rslint glob
// patterns. baseDir is the directory containing the .gitignore relative to
// the config root (empty for root .gitignore).
//
// Conversion rules (aligned with git spec):
//
//	"dist/"     → "**/dist/**"        (unrooted dir, matches at any depth)
//	"/dist"     → "dist/**"           (root-anchored)
//	"*.log"     → "**/*.log"          (unrooted file pattern)
//	"src/dist"  → "src/dist/**"       (contains /, implicitly rooted)
//	"!dist/"    → "!**/dist/**"       (negation preserved)
//
// For nested .gitignore (baseDir != ""), patterns are prefixed:
//
//	baseDir="pkg/app", "tmp/" → "pkg/app/**/tmp/**"
func convertGitignoreToGlobs(content string, baseDir string) []string {
	var globs []string
	for _, line := range strings.Split(content, "\n") {
		line, ok := normalizeGitignoreLine(line)
		if !ok {
			continue
		}
		glob := convertSinglePattern(line, baseDir)
		if glob != "" {
			globs = append(globs, glob)
		}
	}
	return globs
}

func normalizeGitignoreLine(line string) (string, bool) {
	// Strip trailing whitespace (git spec)
	line = strings.TrimRight(line, " \t\r")
	if line == "" || strings.HasPrefix(line, "#") {
		return "", false
	}
	// Strip leading whitespace
	line = strings.TrimLeft(line, " \t")
	if line == "" || strings.HasPrefix(line, "#") {
		return "", false
	}
	return line, true
}

func convertAncestorGitignoreToGlobs(content string, ancestorDir string, configDir string) []string {
	configRel := strings.TrimPrefix(configDir, ancestorDir+"/")
	if configRel == configDir {
		configRel = ""
	}

	var globs []string
	for _, line := range strings.Split(content, "\n") {
		line, ok := normalizeGitignoreLine(line)
		if !ok {
			continue
		}
		glob := convertAncestorSinglePattern(line, configRel)
		if glob != "" {
			globs = append(globs, glob)
		}
	}
	return globs
}

func convertAncestorSinglePattern(line string, configRel string) string {
	negated := false
	body := line
	if strings.HasPrefix(body, "!") {
		negated = true
		body = body[1:]
	}
	body = strings.TrimSuffix(body, "/")

	rooted := strings.HasPrefix(body, "/") || strings.Contains(body, "/")
	if !rooted {
		glob := convertSinglePattern(line, "")
		if configRel != "" && unrootedPatternCoversConfigRel(body, configRel) {
			if negated {
				return "!**/*"
			}
			return "**/*"
		}
		return glob
	}

	glob := convertSinglePattern(line, "")
	if glob == "" || configRel == "" {
		return glob
	}
	if negated {
		glob = strings.TrimPrefix(glob, "!")
	}

	relGlob, ok := stripAncestorConfigPrefix(glob, configRel)
	if !ok {
		return ""
	}
	if negated {
		return "!" + relGlob
	}
	return relGlob
}

func unrootedPatternCoversConfigRel(pattern string, configRel string) bool {
	for _, part := range splitPathComponents(configRel) {
		if matchGlob(pattern, part) {
			return true
		}
	}
	return false
}

func stripAncestorConfigPrefix(glob string, configRel string) (string, bool) {
	globParts := splitPathComponents(glob)
	configParts := splitPathComponents(configRel)
	if len(globParts) == 0 || len(configParts) == 0 {
		return glob, true
	}

	pi := 0
	for ci := 0; ci < len(configParts); ci++ {
		if pi >= len(globParts) {
			return "**/*", true
		}
		part := globParts[pi]
		if part == "**" {
			return strings.Join(globParts[pi:], "/"), true
		}
		if !matchGlob(part, configParts[ci]) {
			return "", false
		}
		pi++
	}

	if pi >= len(globParts) {
		return "**/*", true
	}
	return strings.Join(globParts[pi:], "/"), true
}

func splitPathComponents(p string) []string {
	var parts []string
	for _, part := range strings.Split(p, "/") {
		if part == "" || part == "." {
			continue
		}
		parts = append(parts, part)
	}
	return parts
}

// convertSinglePattern converts one gitignore pattern line to a glob.
func convertSinglePattern(line string, baseDir string) string {
	negated := false
	if strings.HasPrefix(line, "!") {
		negated = true
		line = line[1:]
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
		return ""
	}

	// Build the glob pattern
	var glob string
	if rooted {
		if baseDir != "" {
			glob = baseDir + "/" + line
		} else {
			glob = line
		}
	} else {
		// Unrooted: matches at any depth
		if baseDir != "" {
			glob = baseDir + "/**/" + line
		} else {
			glob = "**/" + line
		}
	}

	// Append /**/* for directory patterns to match all contents.
	// Use /**/* (file-level) instead of /** (directory-level) because gitignore
	// directories must stay `!`-reversible: ParseIgnorePattern classifies
	// `dir/**/*` as dirFileLevelCover (prunable but negation-aware) and `dir/**`
	// as dirAbsoluteBlock (which `!` can never re-include). gitignore's `!`
	// override patterns require the former so isFileIgnored's sequential
	// negation can re-include individual files.
	if dirOnly {
		if !strings.HasSuffix(glob, "/**/*") {
			glob = glob + "/**/*"
		}
	}

	if negated {
		glob = "!" + glob
	}
	return glob
}

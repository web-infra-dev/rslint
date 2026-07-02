package config

import (
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

	// Phase 1: collect ancestor .gitignore files (walk UP).
	// Ancestor patterns apply globally (no path prefix needed since they
	// are above configDir and affect everything below).
	ancestorGlobs := collectAncestorGitignoreGlobs(normalizedRoot, fsys)
	allGlobs = append(allGlobs, ancestorGlobs...)

	// Phase 2: collect descendant .gitignore files (walk DOWN from configDir).
	collectGitignoreGlobs(normalizedRoot, "", fsys, &allGlobs, configIgnores)

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

// collectAncestorGitignoreGlobs walks UP from dir to filesystem root,
// collecting .gitignore patterns from each ancestor directory.
// Returns patterns in root-first order (outermost ancestor first).
func collectAncestorGitignoreGlobs(dir string, fsys vfs.FS) []string {
	// Collect ancestor dirs (from dir's parent up to root)
	var ancestors []string
	current := parentDir(dir)
	for current != "" && current != dir {
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
	for _, ancestor := range ancestors {
		gitignorePath := ancestor + "/.gitignore"
		if content, ok := fsys.ReadFile(gitignorePath); ok {
			// Ancestor patterns have no prefix — they apply globally
			// relative to the config, matching at any depth.
			converted := convertGitignoreToGlobs(content, "")
			globs = append(globs, converted...)
		}
	}
	return globs
}

// parentDir returns the parent directory of dir, or "" if at root.
func parentDir(dir string) string {
	idx := strings.LastIndex(dir, "/")
	if idx <= 0 {
		return ""
	}
	return dir[:idx]
}

// collectGitignoreGlobs recursively scans for .gitignore files and converts
// their patterns to globs. Already-converted patterns from parent .gitignore
// are used to prune directories during scanning (avoids entering e.g. target/
// with thousands of subdirectories).
//
// configIgnores provides additional directory-level pruning from the user's
// lint config. If isDirAbsolutelyBlocked returns true for a directory against
// these patterns, the directory is skipped. This is safe because
// isDirAbsolutelyBlocked is the same predicate used by the linter's
// GetConfigForFile (via isDirBlockedByIgnores) — if a directory is blocked
// here, its files will never be linted, so collecting its .gitignore is
// unnecessary.
func collectGitignoreGlobs(absDir string, relDir string, fsys vfs.FS, result *[]string, configIgnores []IgnorePattern) {
	gitignorePath := absDir + "/.gitignore"
	if content, ok := fsys.ReadFile(gitignorePath); ok {
		localGlobs := convertGitignoreToGlobs(content, relDir)
		*result = append(*result, localGlobs...)
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

		// Prune directories already matched by collected gitignore patterns.
		// This is a scan-local heuristic over the growing glob set (sequential
		// `!` handling); the gap-file walk independently re-validates soundness
		// via canPruneDir, so this stays a fast string-set match. Critical for
		// performance: without it, scanning rspack enters target/ (6,277 Rust
		// build dirs, 0 .ts files).
		if isDirIgnoredByGlobs(*result, childRel) {
			continue
		}

		childAbs := absDir + "/" + dir
		collectGitignoreGlobs(childAbs, childRel, fsys, result, configIgnores)
	}
}

// isDirIgnoredByGlobs reports whether relDir is ignored by the collected
// gitignore globs (sequential `!` evaluation, later overrides earlier). It is
// scan-local: a fast string-set match used only to prune which nested
// .gitignore files get collected, NOT a soundness predicate. The gap-file walk
// re-derives directory pruning from the structured patterns via canPruneDir, so
// for the patterns that ARE collected the walk stays sound regardless. The one
// behavior this skips is a `!` re-include inside an already-ignored directory's
// nested .gitignore — which matches git's own rule that a file cannot be
// re-included once a parent directory is excluded. This logic is unchanged by
// the IgnorePattern refactor.
func isDirIgnoredByGlobs(globs []string, relDir string) bool {
	ignored := false
	for _, pattern := range globs {
		if strings.HasPrefix(pattern, "!") {
			if matchGlob(pattern[1:], relDir) || matchGlob(pattern[1:], relDir+"/x") {
				ignored = false
			}
		} else {
			if matchGlob(pattern, relDir) || matchGlob(pattern, relDir+"/x") {
				ignored = true
			}
		}
	}
	return ignored
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
		// Strip trailing whitespace (git spec)
		line = strings.TrimRight(line, " \t\r")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Strip leading whitespace
		line = strings.TrimLeft(line, " \t")

		glob := convertSinglePattern(line, baseDir)
		if glob != "" {
			globs = append(globs, glob)
		}
	}
	return globs
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

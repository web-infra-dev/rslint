package config

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/vfs"
)

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

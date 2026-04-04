package config

import (
	"io/fs"
	"path"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// defaultExcludeDirs is a set of directory names always excluded from walking.
var defaultExcludeDirs = func() map[string]struct{} {
	m := make(map[string]struct{}, len(utils.DefaultExcludeDirNames))
	for _, name := range utils.DefaultExcludeDirNames {
		m[name] = struct{}{}
	}
	return m
}()

// DiscoverGapFiles scans the filesystem for "gap files" — files that match a config
// entry's `files` pattern but are not in any tsconfig Program. These files get a
// fallback Program (AST-only, no type info) and only run non-type-aware rules.
//
// Files pass through these filters:
//  1. Match at least one config entry's `files` pattern
//  2. Not in global ignores (directory-level and file-level)
//  3. Not already in programFiles (existing tsconfig Programs)
//  4. Pass GetConfigForFile (would actually receive lint rules)
//
// Uses single-pass fs.WalkDir with directory-level pruning (node_modules, .git,
// and isDirBlocked patterns are skipped at walk time, not after traversal).
//
// When allowFiles/allowDirs are provided (CLI args), only files within scope.
//
// Returns:
//   - nil: no config entry has a `files` field → caller uses legacy tsconfig-only behavior
//   - []: `files` present but no gaps found
//   - [...]: gap files to create a fallback Program for
func DiscoverGapFiles(
	config RslintConfig,
	configDir string,
	fsys vfs.FS,
	programFiles map[string]struct{},
	allowFiles []string,
	allowDirs []string,
) []string {
	// 1. Collect global ignore patterns and files patterns from config entries.
	var globalIgnores []string
	var allFilesPatterns []string
	hasFilesField := false

	for _, entry := range config {
		if isGlobalIgnoreEntry(entry) {
			globalIgnores = append(globalIgnores, entry.Ignores...)
			continue
		}
		if len(entry.Files) > 0 {
			hasFilesField = true
			allFilesPatterns = append(allFilesPatterns, entry.Files...)
		}
	}

	// No entry has files → backward compat, signal caller to skip new logic.
	if !hasFilesField {
		return nil
	}

	// Deduplicate patterns.
	allFilesPatterns = deduplicate(allFilesPatterns)

	// Build allowFiles set for fast lookup.
	var allowFileSet map[string]struct{}
	if allowFiles != nil {
		allowFileSet = make(map[string]struct{}, len(allowFiles))
		for _, f := range allowFiles {
			allowFileSet[tspath.NormalizePath(f)] = struct{}{}
		}
	}

	// 2. Prepend default directory ignores so they are always active
	// regardless of user config.
	globalIgnores = append(utils.DefaultIgnoreDirGlobs(), globalIgnores...)

	// Use non-nil empty slice to distinguish "files field present, no gaps"
	// from "no files field" (nil).
	gapFiles := []string{}

	// Fast path: when only specific files are provided (e.g., lint-staged),
	// check each file directly instead of walking the entire filesystem.
	if allowFileSet != nil && allowDirs == nil {
		for f := range allowFileSet {
			if _, exists := programFiles[f]; exists {
				continue
			}
			if isFileIgnored(f, globalIgnores, configDir) {
				continue
			}
			if config.GetConfigForFile(f, configDir) == nil {
				continue
			}
			gapFiles = append(gapFiles, f)
		}
		return deduplicate(gapFiles)
	}

	// 3. Walk directory tree once with directory-level pruning, matching all
	// patterns per file. This replaces the previous per-pattern GlobWalk
	// approach which could not prune directories (node_modules, .git, etc.)
	// and resulted in O(patterns × all_files) traversal.
	normalizedConfigDir := normalizeGlobPath(configDir)
	fsAdapter := &vfsAdapter{vfs: fsys, root: normalizedConfigDir}

	// Normalize patterns relative to configDir for matching.
	relativePatterns := make([]string, len(allFilesPatterns))
	for i, pattern := range allFilesPatterns {
		normalizedPattern := normalizeGlobPath(tspath.ResolvePath(configDir, pattern))
		relativePatterns[i] = strings.TrimPrefix(normalizedPattern, normalizedConfigDir+"/")
	}

	// Cache for directory ignore checks.
	dirIgnoreCache := make(map[string]bool)

	seen := make(map[string]struct{})

	_ = fs.WalkDir(fsAdapter, ".", func(walkPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}

		if d.IsDir() {
			// Prune default-excluded directories (node_modules, .git).
			if _, excluded := defaultExcludeDirs[d.Name()]; excluded && walkPath != "." {
				return fs.SkipDir
			}

			// Prune directories blocked by global ignores (directory-level patterns).
			if walkPath != "." {
				if blocked, ok := dirIgnoreCache[walkPath]; ok {
					if blocked {
						return fs.SkipDir
					}
				} else {
					blocked = isDirPathBlocked(walkPath, globalIgnores)
					dirIgnoreCache[walkPath] = blocked
					if blocked {
						return fs.SkipDir
					}
				}
			}

			return nil
		}

		// File: check if it matches any files pattern.
		matched := false
		for _, pattern := range relativePatterns {
			if ok, _ := doublestar.Match(pattern, walkPath); ok {
				matched = true
				break
			}
		}
		if !matched {
			return nil
		}

		fullPath := tspath.NormalizePath(path.Join(normalizedConfigDir, walkPath))

		// Skip files already in tsconfig Programs.
		if _, exists := programFiles[fullPath]; exists {
			return nil
		}

		// Skip files in global ignores (file-level evaluation with ! support).
		if isFileIgnored(fullPath, globalIgnores, configDir) {
			return nil
		}

		// Scope to CLI args if provided.
		if allowFileSet != nil || allowDirs != nil {
			inScope := false
			if allowFileSet != nil {
				if _, ok := allowFileSet[fullPath]; ok {
					inScope = true
				}
			}
			if !inScope && allowDirs != nil {
				for _, dir := range allowDirs {
					if tspath.StartsWithDirectory(fullPath, dir, true) {
						inScope = true
						break
					}
				}
			}
			if !inScope {
				return nil
			}
		}

		// Final check: would this file actually receive lint rules?
		if config.GetConfigForFile(fullPath, configDir) == nil {
			return nil
		}

		if _, exists := seen[fullPath]; !exists {
			seen[fullPath] = struct{}{}
			gapFiles = append(gapFiles, fullPath)
		}
		return nil
	})

	return gapFiles
}

// DiscoverGapFilesMultiConfig runs DiscoverGapFiles for each config in a
// monorepo config map and returns the union of all gap files.
func DiscoverGapFilesMultiConfig(
	configMap map[string]RslintConfig,
	fsys vfs.FS,
	programFiles map[string]struct{},
	allowFiles []string,
	allowDirs []string,
) []string {
	seen := make(map[string]struct{})
	var allGapFiles []string

	for configDir, cfg := range configMap {
		gapFiles := DiscoverGapFiles(cfg, configDir, fsys, programFiles, allowFiles, allowDirs)
		for _, f := range gapFiles {
			if _, exists := seen[f]; !exists {
				seen[f] = struct{}{}
				allGapFiles = append(allGapFiles, f)
			}
		}
	}

	if len(allGapFiles) == 0 {
		return nil
	}
	return allGapFiles
}

// deduplicate returns a copy of the input slice with duplicates removed, preserving order.
func deduplicate(items []string) []string {
	if len(items) == 0 {
		return items
	}
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

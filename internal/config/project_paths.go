package config

import (
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
)

func resolveDeclaredProjectPaths(fsys vfs.FS, rslintConfig RslintConfig, configDirectory string) ([]string, error) {
	tsConfigs := []string{}
	seenPaths := make(map[string]struct{})

	for _, entry := range rslintConfig {
		if entry.LanguageOptions == nil || entry.LanguageOptions.ParserOptions == nil {
			continue
		}
		entryBaseDirectory := configEntryBaseDirectory(entry, configDirectory)

		for _, config := range entry.LanguageOptions.ParserOptions.Project {
			if containsGlobPattern(config) {
				matches, err := expandProjectGlob(fsys, entryBaseDirectory, config)
				if err != nil {
					return nil, err
				}
				if len(matches) == 0 {
					return nil, fmt.Errorf("glob pattern %q matched no files", config)
				}
				for _, match := range matches {
					tsConfigs = appendUniqueConfigPath(tsConfigs, seenPaths, match)
				}
				continue
			}

			tsconfigPath := tspath.ResolvePath(entryBaseDirectory, config)

			if !fsys.FileExists(tsconfigPath) {
				return nil, fmt.Errorf("tsconfig file %q doesn't exist", tsconfigPath)
			}

			tsConfigs = appendUniqueConfigPath(tsConfigs, seenPaths, tsconfigPath)
		}
	}

	return tsConfigs, nil
}

// ResolveTsConfigPaths extracts tsconfig paths from a rslint config's parserOptions.project,
// with an auto-detection fallback to tsconfig.json in the config directory.
// Returns (nil, nil) when no tsconfigs are found. Returns (nil, err) when
// config validation fails (e.g. glob matched no files, tsconfig doesn't exist).
func ResolveTsConfigPaths(rslintConfig RslintConfig, cwd string, fs vfs.FS) ([]string, error) {
	if fs == nil {
		return nil, nil
	}
	tsConfigs, err := resolveDeclaredProjectPaths(fs, rslintConfig, cwd)
	if err != nil {
		return nil, err
	}
	if len(tsConfigs) == 0 {
		// An explicit empty list means "use no TypeScript project". It must not
		// silently turn into the default tsconfig.json; callers that want the
		// default discovery behavior omit parserOptions.project.
		if hasExplicitProjectSetting(rslintConfig) {
			return nil, nil
		}
		defaultTsConfig := tspath.ResolvePath(cwd, "tsconfig.json")
		if fs.FileExists(defaultTsConfig) {
			return []string{defaultTsConfig}, nil
		}
		return nil, nil
	}
	return tsConfigs, nil
}

func hasExplicitProjectSetting(config RslintConfig) bool {
	for _, entry := range config {
		if entry.LanguageOptions != nil &&
			entry.LanguageOptions.ParserOptions != nil &&
			entry.LanguageOptions.ParserOptions.Project != nil {
			return true
		}
	}
	return false
}

func appendUniqueConfigPath(paths []string, seenPaths map[string]struct{}, configPath string) []string {
	normalizedPath := tspath.NormalizePath(configPath)
	if _, exists := seenPaths[normalizedPath]; exists {
		return paths
	}
	seenPaths[normalizedPath] = struct{}{}
	return append(paths, normalizedPath)
}

func expandProjectGlob(fsys vfs.FS, configDirectory string, pattern string) ([]string, error) {
	// The effective config base is a literal directory, not part of the authored glob.
	normalizedPattern := normalizeGlobPath(pattern)
	authoredSearchRoot := globSearchRoot(normalizedPattern, ".")
	searchRoot := normalizeGlobPath(tspath.ResolvePath(configDirectory, authoredSearchRoot))
	resolvedPattern := normalizeGlobPath(tspath.ResolvePath(configDirectory, normalizedPattern))

	if !fsys.DirectoryExists(searchRoot) {
		return nil, nil
	}

	relativePattern := relativeGlobPattern(searchRoot, resolvedPattern)
	// Project globs follow symlinks (e.g. tsconfig
	// referenced via packages/*/tsconfig.json where packages may be
	// symlinks in pnpm workspaces). It runs single-threaded under
	// doublestar.GlobWalk, so the cycle dedupe is deterministic.
	globFS := &vfsAdapter{vfs: fsys, root: searchRoot}

	matches := []string{}
	err := doublestar.GlobWalk(globFS, relativePattern, func(path string, d fs.DirEntry) error {
		fullPath := tspath.ResolvePath(searchRoot, path)
		matches = append(matches, tspath.NormalizePath(fullPath))
		return nil
	}, doublestar.WithFilesOnly())
	if err != nil {
		return nil, fmt.Errorf("error expanding glob pattern %q: %w", pattern, err)
	}

	sort.Strings(matches)
	return matches, nil
}

func relativeGlobPattern(searchRoot string, resolvedPattern string) string {
	relativePattern := strings.TrimPrefix(resolvedPattern, searchRoot)
	return strings.TrimPrefix(relativePattern, "/")
}

func containsGlobPattern(path string) bool {
	return strings.ContainsAny(path, "*?[")
}

func globSearchRoot(pattern string, fallback string) string {
	firstGlob := strings.IndexAny(pattern, "*?[")
	if firstGlob == -1 {
		return pattern
	}

	prefix := pattern[:firstGlob]
	if prefix == "" {
		return fallback
	}

	if strings.HasSuffix(prefix, "/") {
		root := strings.TrimSuffix(prefix, "/")
		if root == "" {
			return "/"
		}
		if strings.HasSuffix(root, ":") {
			return root + "/"
		}
		return root
	}

	lastSlash := strings.LastIndex(prefix, "/")
	if lastSlash == -1 {
		return fallback
	}

	root := strings.TrimSuffix(prefix[:lastSlash], "/")
	if root == "" {
		return "/"
	}
	if strings.HasSuffix(root, ":") {
		return root + "/"
	}
	return root
}

func normalizeGlobPath(path string) string {
	return strings.ReplaceAll(tspath.NormalizePath(path), "\\", "/")
}

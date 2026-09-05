package config

import (
	"path"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// DefaultLintFileExtensions are the file extensions rslint discovers when a
// config entry omits `files`. This intentionally extends ESLint's default
// .js/.mjs/.cjs set with JSX and TypeScript-family files.
var DefaultLintFileExtensions = []string{".js", ".mjs", ".cjs", ".jsx", ".ts", ".tsx", ".mts", ".cts"}

var defaultLintFileExtensionSet = func() map[string]struct{} {
	m := make(map[string]struct{}, len(DefaultLintFileExtensions))
	for _, ext := range DefaultLintFileExtensions {
		m[ext] = struct{}{}
	}
	return m
}()

// IsSupportedLintFile reports whether rslint can parse and lint this path.
func IsSupportedLintFile(filePath string) bool {
	_, ok := defaultLintFileExtensionSet[strings.ToLower(path.Ext(filePath))]
	return ok
}

func isDefaultLintFile(filePath string) bool {
	_, ok := defaultLintFileExtensionSet[path.Ext(filePath)]
	return ok
}

// isFileSelectedByConfig reports whether the config itself selects filePath.
// The implicit default baseline is always present. An explicit `files` entry
// extends it only for paths not excluded by that same entry's `ignores`.
func isFileSelectedByConfig(config RslintConfig, filePath string, configDir string) bool {
	if configNeedsTargetResolver(config) {
		decision := newConfigTargetResolver(config, configDir, nil).resolve(filePath, "")
		return decision.selected && !decision.globallyIgnored
	}
	if isDefaultLintFile(filePath) {
		return true
	}
	for _, entry := range config {
		if !isGlobalIgnoreEntry(entry) &&
			hasFileSelectors(entry) &&
			isFileMatchedByConfigEntry(filePath, entry, configDir) &&
			!isFileIgnored(filePath, ParseIgnorePatterns(entry.Ignores), configDir) {
			return true
		}
	}
	return false
}

// IsDefaultExcludedPath reports whether filePath crosses a directory that is
// never traversed by default below scanRoot.
func IsDefaultExcludedPath(filePath string, scanRoot string, useCaseSensitive bool) bool {
	filePath = tspath.NormalizePath(filePath)
	scanRoot = tspath.NormalizePath(scanRoot)
	if hasDefaultExcludedSegment(scanRoot, useCaseSensitive) {
		return true
	}
	if pathsEqual(filePath, scanRoot, useCaseSensitive) ||
		tspath.StartsWithDirectory(filePath, scanRoot, useCaseSensitive) {
		relativePath := tspath.GetRelativePathFromDirectory(
			scanRoot,
			filePath,
			tspath.ComparePathsOptions{
				CurrentDirectory:          scanRoot,
				UseCaseSensitiveFileNames: useCaseSensitive,
			},
		)
		return hasDefaultExcludedSegment(relativePath, useCaseSensitive)
	}
	return hasDefaultExcludedSegment(filePath, useCaseSensitive)
}

func IsDefaultExcludedDirectoryName(name string, useCaseSensitive bool) bool {
	for excluded := range defaultExcludeDirs {
		if pathsEqual(name, excluded, useCaseSensitive) {
			return true
		}
	}
	return false
}

func hasDefaultExcludedSegment(walkPath string, useCaseSensitive bool) bool {
	for _, segment := range strings.Split(walkPath, "/") {
		if IsDefaultExcludedDirectoryName(segment, useCaseSensitive) {
			return true
		}
	}
	return false
}

var defaultExcludeDirs = func() map[string]struct{} {
	m := make(map[string]struct{}, len(utils.DefaultExcludeDirNames))
	for _, name := range utils.DefaultExcludeDirNames {
		m[name] = struct{}{}
	}
	return m
}()

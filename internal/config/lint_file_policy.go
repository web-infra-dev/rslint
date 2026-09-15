package config

import (
	"path"
	"strings"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/vue/vuesfc"
)

// DefaultLintFileExtensions are the file extensions rslint discovers when a
// config entry omits `files`. This intentionally extends ESLint's default
// .js/.mjs/.cjs set with JSX and TypeScript-family files.
var DefaultLintFileExtensions = []string{".js", ".mjs", ".cjs", ".jsx", ".ts", ".tsx", ".mts", ".cts"}

// OptInLintFileExtensions are extensions rslint can parse but never discovers
// on its own: a config reaches them only by naming them in `files`.
//
// A Vue Single File Component is here rather than in the discovery baseline
// because adding it there would change what every existing project lints. A
// repository holding .vue files would suddenly report every core and
// typescript-eslint diagnostic in their <script> blocks on the run after an
// upgrade, without anyone asking for it. Enabling them is a decision, so it
// takes writing `files: ['**/*.vue']` or extending a config that does.
var OptInLintFileExtensions = []string{vuesfc.Extension}

var defaultLintFileExtensionSet = newExtensionSet(DefaultLintFileExtensions)

var supportedLintFileExtensionSet = newExtensionSet(
	append(append([]string(nil), DefaultLintFileExtensions...), OptInLintFileExtensions...),
)

func newExtensionSet(extensions []string) map[string]struct{} {
	set := make(map[string]struct{}, len(extensions))
	for _, extension := range extensions {
		set[extension] = struct{}{}
	}
	return set
}

// IsSupportedLintFile reports whether rslint can parse and lint this path.
// Being supported does not make a file discovered: see isDefaultLintFile.
func IsSupportedLintFile(filePath string) bool {
	_, ok := supportedLintFileExtensionSet[strings.ToLower(path.Ext(filePath))]
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

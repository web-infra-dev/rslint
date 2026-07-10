package config

import (
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
)

// FindNearestConfig finds the config whose configDirectory is the nearest
// ancestor of (or exact match for) filePath. It picks the deepest (longest)
// matching configDirectory, mirroring ESLint v10's per-file config lookup.
// Returns the configDirectory and the config, or ("", nil) if no config matches.
func FindNearestConfig(filePath string, configMap map[string]RslintConfig) (string, RslintConfig) {
	return FindNearestConfigWithCaseSensitivity(filePath, configMap, true)
}

// FindNearestConfigWithCaseSensitivity applies the same nearest-ancestor
// lookup using the filesystem's path comparison semantics.
func FindNearestConfigWithCaseSensitivity(filePath string, configMap map[string]RslintConfig, useCaseSensitive bool) (string, RslintConfig) {
	bestDir := ""
	var bestConfig RslintConfig

	for configDir, cfg := range configMap {
		if tspath.StartsWithDirectory(filePath, configDir, useCaseSensitive) {
			if len(configDir) > len(bestDir) {
				bestDir = configDir
				bestConfig = cfg
			}
		}
	}

	return bestDir, bestConfig
}

// FindNearestConfigWithFS applies filesystem case and realpath semantics. It
// first preserves a caller's lexical config-root alias when one matches, then
// falls back to physical config roots for targets expressed through another
// alias of the same directory.
func FindNearestConfigWithFS(filePath string, configMap map[string]RslintConfig, fsys vfs.FS) (string, RslintConfig) {
	index := newConfigDirectoryIndex(configMap, fsys)
	configDir, ok := index.nearestConfig(filePath)
	if !ok {
		return "", nil
	}
	return configDir, configMap[configDir]
}

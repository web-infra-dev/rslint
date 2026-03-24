package config

import (
	"github.com/microsoft/typescript-go/shim/tspath"
)

// FindNearestConfig finds the config whose configDirectory is the nearest
// ancestor of (or exact match for) filePath. It picks the deepest (longest)
// matching configDirectory, mirroring ESLint v10's per-file config lookup.
// Returns the configDirectory and the config, or ("", nil) if no config matches.
func FindNearestConfig(filePath string, configMap map[string]RslintConfig) (string, RslintConfig) {
	bestDir := ""
	var bestConfig RslintConfig

	for configDir, cfg := range configMap {
		if tspath.StartsWithDirectory(filePath, configDir, true) {
			if len(configDir) > len(bestDir) {
				bestDir = configDir
				bestConfig = cfg
			}
		}
	}

	return bestDir, bestConfig
}

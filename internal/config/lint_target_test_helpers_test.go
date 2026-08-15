package config

import (
	"sort"

	"github.com/microsoft/typescript-go/shim/vfs"
)

// discoverFilesOutsideProgramsForTest preserves the historical discovery
// test matrix without publishing Program membership as a config-layer API.
// Production code discovers owned targets first; program/loader decides their
// membership afterward.
func discoverFilesOutsideProgramsForTest(
	config RslintConfig,
	configDir string,
	fsys vfs.FS,
	programFiles map[string]struct{},
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
) []string {
	targets := DiscoverLintFiles(config, configDir, fsys, allowFiles, allowDirs, singleThreaded)
	result := make([]string, 0, len(targets))
	for _, target := range targets {
		if _, exists := programFiles[target]; !exists {
			result = append(result, target)
		}
	}
	return result
}

func discoverFilesOutsideProgramsMultiConfigForTest(
	configMap map[string]RslintConfig,
	fsys vfs.FS,
	programFiles map[string]struct{},
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
) []string {
	if len(configMap) == 0 {
		return nil
	}

	index := newConfigDirectoryIndex(configMap, fsys)
	configDirs := make([]string, 0, len(configMap))
	for configDir := range configMap {
		configDirs = append(configDirs, configDir)
	}
	sort.Strings(configDirs)

	seen := make(map[string]struct{})
	var result []string
	for _, configDir := range configDirs {
		targets := discoverLintTargetsForConfigInMap(
			configMap,
			index,
			nil,
			configDir,
			fsys,
			allowFiles,
			allowDirs,
			singleThreaded,
		)
		for _, target := range targets {
			if _, exists := programFiles[target.Path]; exists {
				continue
			}
			if _, exists := seen[target.Path]; exists {
				continue
			}
			seen[target.Path] = struct{}{}
			result = append(result, target.Path)
		}
	}

	sort.Strings(result)
	return result
}

package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// ConfigLoader handles loading and parsing of rslint and tsconfig files
type ConfigLoader struct {
	fs               vfs.FS
	currentDirectory string
}

// NewConfigLoader creates a new configuration loader
func NewConfigLoader(fs vfs.FS, currentDirectory string) *ConfigLoader {
	return &ConfigLoader{
		fs:               fs,
		currentDirectory: currentDirectory,
	}
}

// LoadRslintConfig loads and parses a rslint configuration file.
// For JSON/JSONC files, a deprecation warning is printed to stderr.
func (loader *ConfigLoader) LoadRslintConfig(configPath string) (RslintConfig, string, error) {
	configFileName := tspath.ResolvePath(loader.currentDirectory, configPath)
	if !loader.fs.FileExists(configFileName) {
		return nil, "", fmt.Errorf("rslint config file %q doesn't exist", configFileName)
	}

	// Deprecation warning for JSON/JSONC config
	if strings.HasSuffix(configFileName, ".json") || strings.HasSuffix(configFileName, ".jsonc") {
		fmt.Fprintf(os.Stderr,
			"\n[rslint] Warning: JSON configuration is deprecated and will be removed in a future version.\n"+
				"[rslint] Please migrate to a JS/TS config. Run `rslint --init` to generate a new config file.\n\n",
		)
	}

	data, ok := loader.fs.ReadFile(configFileName)
	if !ok {
		return nil, "", fmt.Errorf("error reading rslint config file %q", configFileName)
	}

	var config RslintConfig
	// Use JSONC parser to support comments and trailing commas
	if err := utils.ParseJSONC([]byte(data), &config); err != nil {
		return nil, "", fmt.Errorf("error parsing rslint config file %q: %w", configFileName, err)
	}
	if err := ValidateConfig(config); err != nil {
		return nil, "", fmt.Errorf("invalid rslint config file %q: %w", configFileName, err)
	}

	// Normalize JSON config: inject core rules and plugin rules into each entry's Rules map.
	// User-specified rules take precedence (they are applied after the defaults).
	config = normalizeJSONConfig(config)

	// Update current directory to the config file's directory
	configDirectory := tspath.GetDirectoryPath(configFileName)
	return config, configDirectory, nil
}

// normalizeJSONConfig injects core rules and plugin rules into each entry's Rules map.
// This ensures JSON config and JS config are processed identically in GetConfigForFile.
// User-specified rules always take precedence over auto-enabled defaults.
// NOTE: This function mutates the input slice in-place (modifies entry Rules maps directly).
func normalizeJSONConfig(config RslintConfig) RslintConfig {
	for i := range config {
		entry := &config[i]

		// Skip global-ignore-only entries (no rules, plugins, or other fields)
		if isGlobalIgnoreEntry(*entry) {
			continue
		}

		if entry.Rules == nil {
			entry.Rules = make(Rules)
		}

		// Auto-enable core rules as defaults
		for _, r := range GetCoreRules() {
			if _, exists := entry.Rules[r.Name]; !exists {
				entry.Rules[r.Name] = "error"
			}
		}

		// Auto-enable plugin rules as defaults
		for _, plugin := range entry.Plugins {
			info, ok := pluginByDeclName[plugin]
			if !ok {
				continue
			}
			for _, r := range info.getAllRules() {
				if _, exists := entry.Rules[r.Name]; !exists {
					entry.Rules[r.Name] = "error"
				}
			}
		}
	}

	return config
}

// LoadDefaultRslintConfig attempts to load default configuration files
func (loader *ConfigLoader) LoadDefaultRslintConfig() (RslintConfig, string, error) {
	defaultConfigs := []string{"rslint.json", "rslint.jsonc"}

	for _, defaultConfig := range defaultConfigs {
		defaultConfigPath := tspath.ResolvePath(loader.currentDirectory, defaultConfig)
		if loader.fs.FileExists(defaultConfigPath) {
			return loader.LoadRslintConfig(defaultConfig)
		}
	}

	return nil, "", errors.New("no rslint config file found. Expected rslint.json or rslint.jsonc")
}

// resolveProjectPaths resolves one effective parserOptions.project value. It
// performs path/glob validation but does not parse a tsconfig or build a
// Program. The authored lexical path is preserved because TypeScript resolves
// includes and extends relative to that declaration location.
func (loader *ConfigLoader) resolveProjectPaths(projects ProjectPaths, configDirectory string) ([]string, error) {
	paths := make([]string, 0, len(projects))
	seenPaths := make(map[string]struct{})
	for _, project := range projects {
		if containsGlobPattern(project) {
			matches, err := loader.expandProjectGlob(configDirectory, project)
			if err != nil {
				return nil, err
			}
			if len(matches) == 0 {
				return nil, fmt.Errorf("glob pattern %q matched no files", project)
			}
			for _, match := range matches {
				paths = appendUniqueConfigPath(paths, seenPaths, match)
			}
			continue
		}

		tsconfigPath := tspath.ResolvePath(configDirectory, project)
		if !loader.fs.FileExists(tsconfigPath) {
			return nil, fmt.Errorf("tsconfig file %q doesn't exist", tsconfigPath)
		}
		paths = appendUniqueConfigPath(paths, seenPaths, tsconfigPath)
	}
	return paths, nil
}

func appendUniqueConfigPath(paths []string, seenPaths map[string]struct{}, configPath string) []string {
	normalizedPath := tspath.NormalizePath(configPath)
	if _, exists := seenPaths[normalizedPath]; exists {
		return paths
	}
	seenPaths[normalizedPath] = struct{}{}
	return append(paths, normalizedPath)
}

func (loader *ConfigLoader) expandProjectGlob(configDirectory string, pattern string) ([]string, error) {
	resolvedPattern := normalizeGlobPath(tspath.ResolvePath(configDirectory, pattern))
	searchRoot := globSearchRoot(resolvedPattern, normalizeGlobPath(configDirectory))

	if !loader.fs.DirectoryExists(searchRoot) {
		return nil, nil
	}

	relativePattern := relativeGlobPattern(searchRoot, resolvedPattern)
	// expandProjectGlob historically follows symlinks (e.g. tsconfig
	// referenced via packages/*/tsconfig.json where packages may be
	// symlinks in pnpm workspaces). It runs single-threaded under
	// doublestar.GlobWalk, so the cycle dedupe is deterministic.
	fsys := &vfsAdapter{vfs: loader.fs, root: searchRoot, followSymlinks: true}

	matches := []string{}
	err := doublestar.GlobWalk(fsys, relativePattern, func(path string, d fs.DirEntry) error {
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

// LoadRslintConfiguration loads and validates only the rslint configuration.
// Project resolution is a separate orchestration step because plain linting
// first needs the effective target set.
func (loader *ConfigLoader) LoadRslintConfiguration(configPath string) (RslintConfig, string, error) {
	if configPath != "" {
		return loader.LoadRslintConfig(configPath)
	}
	return loader.LoadDefaultRslintConfig()
}

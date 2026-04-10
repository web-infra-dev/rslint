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
			for _, r := range GetPluginRules(NormalizePluginName(plugin)) {
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

// LoadTsConfigsFromRslintConfig extracts and validates TypeScript configuration paths from rslint config.
// Returns an empty slice (no error) when no parserOptions.project is specified — this is valid
// for pure JS projects that don't need explicit TypeScript configuration.
func (loader *ConfigLoader) LoadTsConfigsFromRslintConfig(rslintConfig RslintConfig, configDirectory string) ([]string, error) {
	tsConfigs := []string{}
	seenPaths := make(map[string]struct{})

	for _, entry := range rslintConfig {
		if entry.LanguageOptions == nil || entry.LanguageOptions.ParserOptions == nil {
			continue
		}

		for _, config := range entry.LanguageOptions.ParserOptions.Project {
			if containsGlobPattern(config) {
				matches, err := loader.expandProjectGlob(configDirectory, config)
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

			tsconfigPath := tspath.ResolvePath(configDirectory, config)

			if !loader.fs.FileExists(tsconfigPath) {
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
	loader := NewConfigLoader(fs, cwd)
	tsConfigs, err := loader.LoadTsConfigsFromRslintConfig(rslintConfig, cwd)
	if err != nil {
		return nil, err
	}
	if len(tsConfigs) == 0 {
		defaultTsConfig := tspath.ResolvePath(cwd, "tsconfig.json")
		if fs.FileExists(defaultTsConfig) {
			return []string{defaultTsConfig}, nil
		}
		return nil, nil
	}
	return tsConfigs, nil
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

	relativePattern := strings.TrimPrefix(resolvedPattern, searchRoot+"/")
	fsys := &vfsAdapter{vfs: loader.fs, root: searchRoot}

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

// LoadConfiguration is a convenience method that loads both rslint and tsconfig configurations
func (loader *ConfigLoader) LoadConfiguration(configPath string) (RslintConfig, []string, string, error) {
	var rslintConfig RslintConfig
	var configDirectory string
	var err error

	if configPath != "" {
		rslintConfig, configDirectory, err = loader.LoadRslintConfig(configPath)
	} else {
		rslintConfig, configDirectory, err = loader.LoadDefaultRslintConfig()
	}

	if err != nil {
		return nil, nil, "", err
	}

	tsConfigs, err := loader.LoadTsConfigsFromRslintConfig(rslintConfig, configDirectory)
	if err != nil {
		return nil, nil, "", err
	}

	return rslintConfig, tsConfigs, configDirectory, nil
}

// LoadConfigurationWithFallback loads configuration and handles errors by printing to stderr and exiting
// This is for backward compatibility with the existing cmd behavior
func LoadConfigurationWithFallback(configPath string, currentDirectory string, fs vfs.FS) (RslintConfig, []string, string) {
	loader := NewConfigLoader(fs, currentDirectory)

	rslintConfig, tsConfigs, configDirectory, err := loader.LoadConfiguration(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	return rslintConfig, tsConfigs, configDirectory
}

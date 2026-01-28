package config

import (
	"errors"
	"fmt"
	"os"
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

// LoadRslintConfig loads and parses a rslint configuration file
func (loader *ConfigLoader) LoadRslintConfig(configPath string) (RslintConfig, string, error) {
	configFileName := tspath.ResolvePath(loader.currentDirectory, configPath)
	if !loader.fs.FileExists(configFileName) {
		return nil, "", fmt.Errorf("rslint config file %q doesn't exist", configFileName)
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

	// Update current directory to the config file's directory
	configDirectory := tspath.GetDirectoryPath(configFileName)
	return config, configDirectory, nil
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

// LoadTsConfigsFromRslintConfig extracts and validates TypeScript configuration paths from rslint config
// Now supports glob patterns like "./packages/*/tsconfig.json"
func (loader *ConfigLoader) LoadTsConfigsFromRslintConfig(rslintConfig RslintConfig, configDirectory string) ([]string, error) {
	tsConfigs := []string{}
	seenPaths := make(map[string]bool) // Track unique paths to avoid duplicates

	for _, entry := range rslintConfig {
		if entry.LanguageOptions == nil || entry.LanguageOptions.ParserOptions == nil {
			continue
		}

		for _, config := range entry.LanguageOptions.ParserOptions.Project {
			// Check if the config path contains glob characters
			if containsGlobPattern(config) {
				// Resolve the glob pattern relative to config directory
				pattern := tspath.ResolvePath(configDirectory, config)

				// Expand the glob pattern using doublestar which supports ** patterns
				matches, err := doublestar.FilepathGlob(pattern)
				if err != nil {
					return nil, fmt.Errorf("error expanding glob pattern %q: %w", config, err)
				}

				if len(matches) == 0 {
					return nil, fmt.Errorf("glob pattern %q matched no files", config)
				}

				// Add all matched files
				for _, match := range matches {
					// Verify each matched file exists and is a file (not a directory)
					if !loader.fs.FileExists(match) {
						continue // Skip if file doesn't exist in VFS
					}

					// Deduplicate paths
					if !seenPaths[match] {
						tsConfigs = append(tsConfigs, match)
						seenPaths[match] = true
					}
				}
			} else {
				// Non-glob path - handle as before
				tsconfigPath := tspath.ResolvePath(configDirectory, config)

				if !loader.fs.FileExists(tsconfigPath) {
					return nil, fmt.Errorf("tsconfig file %q doesn't exist", tsconfigPath)
				}

				// Deduplicate paths
				if !seenPaths[tsconfigPath] {
					tsConfigs = append(tsConfigs, tsconfigPath)
					seenPaths[tsconfigPath] = true
				}
			}
		}
	}

	if len(tsConfigs) == 0 {
		return nil, errors.New("no TypeScript configuration found in rslint config")
	}

	return tsConfigs, nil
}

// containsGlobPattern checks if a path contains glob pattern characters
func containsGlobPattern(path string) bool {
	return strings.ContainsAny(path, "*?[")
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

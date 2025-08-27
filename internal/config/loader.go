package config

import (
	"errors"
	"fmt"
	"os"

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
func (loader *ConfigLoader) LoadTsConfigsFromRslintConfig(rslintConfig RslintConfig, configDirectory string) ([]string, error) {
	tsConfigs := []string{}

	for _, entry := range rslintConfig {
		if entry.LanguageOptions == nil || entry.LanguageOptions.ParserOptions == nil {
			continue
		}

		for _, config := range entry.LanguageOptions.ParserOptions.Project {
			tsconfigPath := tspath.ResolvePath(configDirectory, config)

			if !loader.fs.FileExists(tsconfigPath) {
				return nil, fmt.Errorf("tsconfig file %q doesn't exist", tsconfigPath)
			}

			tsConfigs = append(tsConfigs, tsconfigPath)
		}
	}

	if len(tsConfigs) == 0 {
		return nil, errors.New("no TypeScript configuration found in rslint config")
	}

	return tsConfigs, nil
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

package main

import (
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/discovery"
)

func explicitConfigCatalogForTest(
	configDirectory string,
	entries rslintconfig.RslintConfig,
) *discovery.ConfigCatalog {
	configDirectory = tspath.NormalizePath(configDirectory)
	return &discovery.ConfigCatalog{
		Configs: map[string]rslintconfig.RslintConfig{
			configDirectory: entries,
		},
		Explicit: true,
	}
}

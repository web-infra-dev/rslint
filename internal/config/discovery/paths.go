package discovery

import (
	"strings"

	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
)

func isDefaultDiscoveryExcluded(path string, cwd string, useCaseSensitive bool) bool {
	return rslintconfig.IsDefaultExcludedPath(path, cwd, useCaseSensitive)
}

func discoveryPathsEqual(a string, b string, useCaseSensitive bool) bool {
	if useCaseSensitive {
		return a == b
	}
	return strings.EqualFold(a, b)
}

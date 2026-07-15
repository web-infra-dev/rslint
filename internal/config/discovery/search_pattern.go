package discovery

import "github.com/web-infra-dev/rslint/internal/config/minimatch"

var ErrInvalidSearchPattern = minimatch.ErrInvalidSearchPattern

type SearchPattern = minimatch.SearchPattern

func CompileSearchPattern(pattern string, basePath string) (SearchPattern, error) {
	return minimatch.CompileSearchPattern(pattern, basePath)
}

func GlobParent(pattern string) string {
	return minimatch.GlobParent(pattern)
}

func isGlobPattern(pattern string) bool {
	return minimatch.IsGlobPattern(pattern)
}

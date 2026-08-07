package utils

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/tspath"
)

// ExternalModuleFolders returns eslint-plugin-import's configured external
// module folders, defaulting to node_modules.
func ExternalModuleFolders(settings map[string]interface{}) []string {
	folders := []string{"node_modules"}
	if settings == nil {
		return folders
	}

	raw, ok := settings["import/external-module-folders"]
	if !ok {
		return folders
	}

	switch typed := raw.(type) {
	case []string:
		return append([]string{}, typed...)
	case []interface{}:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			if folder, ok := item.(string); ok && folder != "" {
				result = append(result, folder)
			}
		}
		if len(result) > 0 {
			return result
		}
	}

	return folders
}

// IsExternalModulePath reports whether a resolved path or unresolved bare
// specifier should be treated as external by eslint-plugin-import rules.
// Callers that ask more than once for the same settings should compile them
// with [compileModuleSettings] and ask that instead.
func IsExternalModulePath(settings map[string]interface{}, specifier string, resolvedPath string) bool {
	return compileModuleSettings(settings).IsExternalPath(specifier, resolvedPath)
}

func pathContainsSegment(fileName string, segment string) bool {
	if fileName == "" || segment == "" {
		return false
	}
	normalizedFileName := "/" + strings.Trim(tspath.NormalizePath(fileName), "/") + "/"
	normalizedSegment := strings.Trim(tspath.NormalizeSlashes(segment), "/")
	return strings.Contains(normalizedFileName, "/"+normalizedSegment+"/")
}

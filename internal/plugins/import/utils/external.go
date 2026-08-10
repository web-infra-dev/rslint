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
func IsExternalModulePath(settings map[string]interface{}, specifier string, resolvedPath string) bool {
	for _, folder := range ExternalModuleFolders(settings) {
		if pathContainsSegment(resolvedPath, folder) {
			return true
		}
	}
	return specifier != "" && !tspath.IsExternalModuleNameRelative(specifier) && resolvedPath == ""
}

// IsExternalModulePathFromPackage mirrors importType.js's resolved-path
// classification. Relative external-module-folders are resolved from the
// nearest package root; their names are only searched as arbitrary path
// segments when the resolved target lies outside that package (for hoisted
// dependencies). Absolute folders are checked directly.
func IsExternalModulePathFromPackage(settings map[string]interface{}, packagePath string, resolvedPath string, caseSensitive bool) bool {
	if resolvedPath == "" {
		return false
	}
	compareOptions := tspath.ComparePathsOptions{
		UseCaseSensitiveFileNames: caseSensitive,
	}
	isOutsidePackage := packagePath != "" && !tspath.ContainsPath(packagePath, resolvedPath, compareOptions)
	normalizedPath := tspath.NormalizeSlashes(resolvedPath)

	for _, folder := range ExternalModuleFolders(settings) {
		if folder == "" {
			continue
		}
		if tspath.IsRootedDiskPath(folder) {
			if tspath.ContainsPath(tspath.NormalizePath(folder), resolvedPath, compareOptions) {
				return true
			}
			continue
		}
		if packagePath != "" {
			folderPath := tspath.ResolvePath(packagePath, folder)
			if tspath.ContainsPath(folderPath, resolvedPath, compareOptions) {
				return true
			}
		}
		if isOutsidePackage {
			cleanFolder := strings.TrimRight(tspath.NormalizeSlashes(folder), "/")
			if cleanFolder != "" && strings.Contains(normalizedPath, "/"+cleanFolder+"/") {
				return true
			}
		}
	}
	return false
}

func pathContainsSegment(fileName string, segment string) bool {
	if fileName == "" || segment == "" {
		return false
	}
	normalizedFileName := "/" + strings.Trim(tspath.NormalizePath(fileName), "/") + "/"
	normalizedSegment := strings.Trim(tspath.NormalizeSlashes(segment), "/")
	return strings.Contains(normalizedFileName, "/"+normalizedSegment+"/")
}

package utils

import (
	"github.com/microsoft/typescript-go/shim/tspath"
)

// ExternalModuleFolders returns eslint-plugin-import's configured external
// module folders, defaulting to node_modules when the setting is absent.
func ExternalModuleFolders(settings map[string]interface{}) []string {
	return externalModuleFolders(settings)
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
// classification. Any resolved target outside the nearest package is
// external. For targets inside it, relative external-module-folders are
// resolved from that package root and absolute folders are checked directly.
func IsExternalModulePathFromPackage(settings map[string]interface{}, packagePath string, resolvedPath string, caseSensitive bool) bool {
	compiled := &ModuleSettings{externalFolders: externalModuleFolders(settings)}
	return compiled.IsExternalPathFromPackage(packagePath, resolvedPath, caseSensitive)
}

// IsExternalPathFromPackage classifies a resolved path using the external
// module folders already compiled for this rule configuration.
func (compiled *ModuleSettings) IsExternalPathFromPackage(packagePath string, resolvedPath string, caseSensitive bool) bool {
	if compiled == nil || resolvedPath == "" {
		return false
	}
	compareOptions := tspath.ComparePathsOptions{
		UseCaseSensitiveFileNames: caseSensitive,
	}
	if packagePath != "" && !tspath.ContainsPath(packagePath, resolvedPath, compareOptions) {
		return true
	}

	for _, folder := range compiled.externalFolders {
		folderPath := tspath.ResolvePath(packagePath, folder)
		if tspath.ContainsPath(folderPath, resolvedPath, compareOptions) {
			return true
		}
	}
	return false
}

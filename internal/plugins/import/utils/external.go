package utils

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/tspath"
)

// ExternalModuleFolders returns eslint-plugin-import's configured external
// module folders, defaulting to node_modules when the setting names none.
func ExternalModuleFolders(settings map[string]interface{}) []string {
	var folders []string
	for _, folder := range settingsStringList(settings, "import/external-module-folders") {
		if folder != "" {
			folders = append(folders, folder)
		}
	}
	if len(folders) == 0 {
		return []string{"node_modules"}
	}
	return folders
}

func pathContainsSegment(fileName string, segment string) bool {
	if fileName == "" || segment == "" {
		return false
	}
	normalizedFileName := "/" + strings.Trim(tspath.NormalizePath(fileName), "/") + "/"
	normalizedSegment := strings.Trim(tspath.NormalizeSlashes(segment), "/")
	return strings.Contains(normalizedFileName, "/"+normalizedSegment+"/")
}

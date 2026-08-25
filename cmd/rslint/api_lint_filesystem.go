package main

import (
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type canonicalPathVFS struct {
	vfs.FS
	canonicalPaths map[string]string
}

func (fs *canonicalPathVFS) Realpath(filePath string) string {
	if canonicalPath := fs.canonicalPaths[exactFilesystemPathID(filePath)]; canonicalPath != "" {
		return canonicalPath
	}
	return fs.FS.Realpath(filePath)
}

// SourceHasBOM forwards to the wrapped VFS. Embedding vfs.FS promotes only
// that interface's methods, so this layer has to pass [utils.BOMSource]
// through by hand or every overlay identity below it becomes invisible.
func (fs *canonicalPathVFS) SourceHasBOM(filePath string) bool {
	return utils.SourceHasBOM(fs.FS, filePath)
}

func addEquivalentFileContentPaths(fileContents map[string]string, configDirectory string, currentDirectory string, fs vfs.FS) {
	if len(fileContents) == 0 || fs == nil {
		return
	}

	type fileContentEntry struct {
		path    string
		content string
	}
	entries := make([]fileContentEntry, 0, len(fileContents))
	for filePath, content := range fileContents {
		entries = append(entries, fileContentEntry{path: filePath, content: content})
	}

	comparePathOptions := tspath.ComparePathsOptions{
		CurrentDirectory:          currentDirectory,
		UseCaseSensitiveFileNames: true,
	}
	addAlias := func(alias string, content string) {
		if alias == "" {
			return
		}
		if _, exists := fileContents[alias]; exists {
			return
		}
		fileContents[alias] = content
	}
	addDirectoryAlias := func(fromDir string, toDir string, filePath string, content string) {
		if fromDir == "" || toDir == "" || !tspath.ContainsPath(fromDir, filePath, comparePathOptions) {
			return
		}
		relativePath := tspath.GetRelativePathFromDirectory(fromDir, filePath, comparePathOptions)
		if relativePath == "" {
			return
		}
		addAlias(tspath.ResolvePath(toDir, relativePath), content)
	}

	realConfigDirectory := fs.Realpath(configDirectory)
	for _, entry := range entries {
		if realPath := fs.Realpath(entry.path); realPath != "" && realPath != entry.path {
			addAlias(realPath, entry.content)
		}
		if realConfigDirectory != "" && tspath.ComparePaths(configDirectory, realConfigDirectory, comparePathOptions) != 0 {
			addDirectoryAlias(configDirectory, realConfigDirectory, entry.path, entry.content)
			addDirectoryAlias(realConfigDirectory, configDirectory, entry.path, entry.content)
		}
	}
}
